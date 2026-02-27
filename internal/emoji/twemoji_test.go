package emoji

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

func TestToTwemojiCodepoint(t *testing.T) {
	tests := []struct {
		name string
		rune rune
		want string
	}{
		{"rocket", '🚀', "1f680"},
		{"grinning face", '😀', "1f600"},
		{"thumbs up", '👍', "1f44d"},
		{"fire", '🔥', "1f525"},
		{"check mark", '✅', "2705"},
		{"red heart", '❤', "2764"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToTwemojiCodepoint(tt.rune)
			if got != tt.want {
				t.Errorf("ToTwemojiCodepoint(%U) = %s, want %s", tt.rune, got, tt.want)
			}
		})
	}
}

func TestToTwemojiCodepointMulti(t *testing.T) {
	tests := []struct {
		name  string
		runes []rune
		want  string
	}{
		{
			"single rune",
			[]rune{'🚀'},
			"1f680",
		},
		{
			"US flag",
			[]rune{0x1F1FA, 0x1F1F8},
			"1f1fa-1f1f8",
		},
		{
			"thumbs up with skin tone",
			[]rune{0x1F44D, 0x1F3FB},
			"1f44d-1f3fb",
		},
		{
			"with variation selector (stripped)",
			[]rune{0x2764, 0xFE0F}, // ❤️ (red heart with VS16)
			"2764",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToTwemojiCodepointMulti(tt.runes)
			if got != tt.want {
				t.Errorf("ToTwemojiCodepointMulti(%v) = %s, want %s", tt.runes, got, tt.want)
			}
		})
	}
}

func TestIsCommonEmoji(t *testing.T) {
	tests := []struct {
		name string
		rune rune
		want bool
	}{
		{"rocket (in list)", '🚀', true},
		{"thumbs up (in list)", '👍', true},
		{"fire (in list)", '🔥', true},
		{"check mark (in list)", '✅', true},
		{"latin letter a (not emoji)", 'a', false},
		{"digit 5 (not emoji)", '5', false},
		{"uncommon emoji (not in list)", '🦘', false}, // kangaroo
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsCommonEmoji(tt.rune)
			if got != tt.want {
				t.Errorf("IsCommonEmoji(%U) = %v, want %v", tt.rune, got, tt.want)
			}
		})
	}
}

func TestGetPNG_CachePath(t *testing.T) {
	// Test that cache path is constructed correctly
	codepoint := "1f680"
	path := getCachePath(codepoint)

	homeDir, _ := os.UserCacheDir()
	expected := filepath.Join(homeDir, "markdown2pdf", "emoji", "twemoji", "1f680.png")

	if path != expected {
		t.Errorf("getCachePath(%s) = %s, want %s", codepoint, path, expected)
	}
}

// TestGetPNG_Download tests actual download from Twemoji CDN.
// This test requires internet access and may be slow.
func TestGetPNG_Download(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping network test in short mode")
	}

	// Clear cache for this emoji to force download
	codepoint := "1f680" // rocket
	cachePath := getCachePath(codepoint)
	os.Remove(cachePath)

	data, err := GetPNG(codepoint)
	if err != nil {
		t.Fatalf("GetPNG(%s) failed: %v", codepoint, err)
	}

	if len(data) == 0 {
		t.Errorf("GetPNG(%s) returned empty data", codepoint)
	}

	// Verify PNG signature
	if !bytes.HasPrefix(data, []byte{0x89, 0x50, 0x4E, 0x47}) {
		t.Errorf("GetPNG(%s) data does not start with PNG signature", codepoint)
	}

	// Verify cache was created
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		t.Errorf("Cache file not created at %s", cachePath)
	}

	// Test cache hit
	data2, err := GetPNG(codepoint)
	if err != nil {
		t.Fatalf("GetPNG(%s) cache read failed: %v", codepoint, err)
	}

	if !bytes.Equal(data, data2) {
		t.Errorf("Cached data differs from original download")
	}
}

// TestGetPNG_NotFound tests behavior when emoji doesn't exist.
func TestGetPNG_NotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping network test in short mode")
	}

	// Use an invalid codepoint
	codepoint := "zzzzz"
	_, err := GetPNG(codepoint)

	if err == nil {
		t.Errorf("GetPNG(%s) should have failed for invalid codepoint", codepoint)
	}
}

// Mock HTTP transport for testing without network
type mockTransport struct {
	roundTripFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.roundTripFunc(req)
}

func TestGetPNG_Mock(t *testing.T) {
	// Create mock PNG data
	mockPNG := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A} // PNG signature

	// Save original http.DefaultTransport
	oldTransport := http.DefaultTransport
	defer func() { http.DefaultTransport = oldTransport }()

	// Install mock transport
	http.DefaultTransport = &mockTransport{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader(mockPNG)),
				Header:     make(http.Header),
			}, nil
		},
	}

	// Clear cache
	codepoint := "1f600"
	cachePath := getCachePath(codepoint)
	os.Remove(cachePath)

	// Test download
	data, err := GetPNG(codepoint)
	if err != nil {
		t.Fatalf("GetPNG(%s) failed: %v", codepoint, err)
	}

	if !bytes.Equal(data, mockPNG) {
		t.Errorf("GetPNG(%s) returned wrong data", codepoint)
	}
}
