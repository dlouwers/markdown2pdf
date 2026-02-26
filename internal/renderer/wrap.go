package renderer

import (
	"strings"
	"unicode/utf8"

	gopdf "github.com/go-pdf/fpdf"
)

// splitTextLines splits text into lines that fit within maxWidth, operating on
// runes (not bytes) for correct UTF-8 handling. fpdf.SplitLines works at the
// byte level and can split multi-byte UTF-8 characters mid-rune, producing
// garbled output. This function uses GetStringWidth for accurate measurement.
//
// Break strategy (matching best typesetting practices):
//  1. Break at spaces (consume the space)
//  2. Break after hyphens (keep the hyphen on the current line)
//  3. If a single word exceeds maxWidth, break at rune boundaries (last resort)
func splitTextLines(fpdf *gopdf.Fpdf, text string, maxWidth float64) []string {
	if maxWidth <= 0 {
		return []string{text}
	}

	// Handle explicit newlines first.
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")

	var result []string
	for _, paragraph := range strings.Split(text, "\n") {
		lines := wrapParagraph(fpdf, paragraph, maxWidth)
		result = append(result, lines...)
	}
	return result
}

// wrapParagraph wraps a single paragraph (no newlines) into lines.
func wrapParagraph(fpdf *gopdf.Fpdf, text string, maxWidth float64) []string {
	if text == "" {
		return []string{""}
	}

	// Fast path: entire text fits on one line.
	if fpdf.GetStringWidth(text) <= maxWidth {
		return []string{text}
	}

	var lines []string
	var line strings.Builder
	lineWidth := 0.0

	i := 0
	runes := []rune(text)
	n := len(runes)

	for i < n {
		// Find the next word (sequence of non-space runes, including trailing hyphen).
		wordStart := i

		// Skip leading spaces — they become the separator.
		spaceCount := 0
		for i < n && runes[i] == ' ' {
			spaceCount++
			i++
		}

		// Collect the word (up to a space, or break after a hyphen).
		wordRunes := make([]rune, 0, 32)
		for i < n && runes[i] != ' ' {
			wordRunes = append(wordRunes, runes[i])
			i++
			if runes[i-1] == '-' {
				break // break after hyphen, keeping it in this word
			}
		}

		// Build the chunk to append: spaces + word.
		var chunk strings.Builder
		if line.Len() > 0 && spaceCount > 0 {
			// Only add one space as separator when joining words.
			chunk.WriteByte(' ')
		} else if line.Len() == 0 && wordStart == 0 {
			// Preserve leading spaces only at the very start of the paragraph.
			for range spaceCount {
				chunk.WriteByte(' ')
			}
		}
		for _, r := range wordRunes {
			chunk.WriteRune(r)
		}

		chunkStr := chunk.String()
		chunkWidth := fpdf.GetStringWidth(chunkStr)

		if lineWidth+chunkWidth <= maxWidth {
			// Chunk fits on the current line.
			line.WriteString(chunkStr)
			lineWidth += chunkWidth
		} else if line.Len() == 0 {
			// Word is too long for any line — break it at rune boundaries.
			lines = append(lines, breakLongWord(fpdf, chunkStr, maxWidth, &line, &lineWidth)...)
		} else {
			// Start a new line, then try the chunk again.
			lines = append(lines, line.String())
			line.Reset()
			lineWidth = 0.0

			// Re-measure without the leading space separator.
			wordStr := string(wordRunes)
			wordWidth := fpdf.GetStringWidth(wordStr)
			if wordWidth <= maxWidth {
				line.WriteString(wordStr)
				lineWidth = wordWidth
			} else {
				// Word alone doesn't fit — break at rune boundaries.
				lines = append(lines, breakLongWord(fpdf, wordStr, maxWidth, &line, &lineWidth)...)
			}
		}
	}

	if line.Len() > 0 {
		lines = append(lines, line.String())
	}

	// Ensure at least one line (empty input case).
	if len(lines) == 0 {
		lines = []string{""}
	}

	return lines
}

// breakLongWord splits a string that is wider than maxWidth at rune boundaries.
// It fills lines and leaves any remainder in the provided line builder.
func breakLongWord(fpdf *gopdf.Fpdf, word string, maxWidth float64, line *strings.Builder, lineWidth *float64) []string {
	var lines []string
	for len(word) > 0 {
		r, size := utf8.DecodeRuneInString(word)
		rStr := string(r)
		rWidth := fpdf.GetStringWidth(rStr)

		if *lineWidth+rWidth > maxWidth && line.Len() > 0 {
			lines = append(lines, line.String())
			line.Reset()
			*lineWidth = 0.0
		}

		line.WriteRune(r)
		*lineWidth += rWidth
		word = word[size:]
	}
	return lines
}
