package emoji

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// TwemojiCDN is the base URL for Twemoji PNG assets (72x72 size).
// Using jsDelivr CDN for reliable, fast access to Twitter's emoji graphics.
const TwemojiCDN = "https://cdn.jsdelivr.net/gh/twitter/twemoji@14/assets/72x72/%s.png"

// GetPNG retrieves the Twemoji PNG for a given codepoint string.
// It first checks the local cache, then downloads from CDN if needed.
//
// Codepoint format: lowercase hex without prefix (e.g., "1f680" for 🚀)
// Multi-codepoint emoji use hyphens (e.g., "1f1fa-1f1f8" for 🇺🇸)
//
// Returns PNG bytes or error if emoji not found or download fails.
func GetPNG(codepoint string) ([]byte, error) {
	cachePath := getCachePath(codepoint)

	// Check cache first
	if data, err := os.ReadFile(cachePath); err == nil {
		return data, nil
	}

	// Download from CDN
	url := fmt.Sprintf(TwemojiCDN, codepoint)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download emoji %s: %w", codepoint, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("emoji not found: %s (HTTP %d)", codepoint, resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read emoji data: %w", err)
	}

	// Cache for next time
	if err := cacheEmoji(cachePath, data); err != nil {
		// Non-fatal: log but return the data anyway
		// In production, might want structured logging here
		_ = err
	}

	return data, nil
}

// ToTwemojiCodepoint converts a rune to its Twemoji codepoint filename
// (lowercase hex without the .png extension).
//
// Examples:
//
//	🚀 (U+1F680) → "1f680"
//	😀 (U+1F600) → "1f600"
//
// For multi-codepoint sequences (flags, skin tones, etc.), use
// ToTwemojiCodepointMulti instead.
func ToTwemojiCodepoint(r rune) string {
	return fmt.Sprintf("%x", r)
}

// ToTwemojiCodepointMulti converts a sequence of runes to Twemoji codepoint
// format with hyphen separators. Variation selectors (U+FE0E, U+FE0F) are
// automatically stripped.
//
// Examples:
//
//	[]rune{0x1F1FA, 0x1F1F8} → "1f1fa-1f1f8" (🇺🇸 US flag)
//	[]rune{0x1F44D, 0x1F3FB} → "1f44d-1f3fb" (👍🏻 thumbs up light skin)
func ToTwemojiCodepointMulti(runes []rune) string {
	parts := make([]string, 0, len(runes))
	for _, r := range runes {
		// Skip variation selectors
		if r == 0xFE0F || r == 0xFE0E {
			continue
		}
		parts = append(parts, fmt.Sprintf("%x", r))
	}
	return strings.Join(parts, "-")
}

// getCachePath returns the local filesystem path for caching a Twemoji PNG.
func getCachePath(codepoint string) string {
	homeDir, err := os.UserCacheDir()
	if err != nil {
		// Fallback to current directory if cache dir unavailable
		homeDir = "."
	}
	return filepath.Join(homeDir, "markdown2pdf", "emoji", "twemoji", codepoint+".png")
}

// cacheEmoji writes PNG data to the local cache.
func cacheEmoji(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
