package renderer

import (
	"strings"

	gopdf "github.com/go-pdf/fpdf"
)

// codeBreakPoint represents a safe location to break inline code.
type codeBreakPoint struct {
	position int  // rune index in the string
	after    bool // true = break after this char, false = break before
}

// findCodeBreakPoints identifies safe locations to break inline code,
// following LaTeX/typography best practices:
// - Break at spaces (always safe)
// - Break after: . / \ - _ (path/namespace separators)
// - Break before: ( ) [ ] { } (delimiter pairs)
// - NEVER break mid-alphanumeric sequence
func findCodeBreakPoints(code string) []codeBreakPoint {
	var points []codeBreakPoint
	runes := []rune(code)

	for i, r := range runes {
		switch r {
		case ' ':
			// Always safe to break at spaces (consume the space)
			points = append(points, codeBreakPoint{position: i, after: true})
		case '.', '/', '\\', '-', '_':
			// Safe to break after separators (keep them on current line)
			points = append(points, codeBreakPoint{position: i, after: true})
		case '(', ')', '[', ']', '{', '}':
			// Safe to break before delimiters (visual grouping)
			points = append(points, codeBreakPoint{position: i, after: false})
		}
	}

	return points
}

// splitCodeAtBreakPoints splits code text into segments that fit within maxWidth,
// using safe break points and adding continuation indicators.
// Returns segments ready to render (WITHOUT background color - that's handled by caller).
func splitCodeAtBreakPoints(fpdf *gopdf.Fpdf, code string, maxWidth float64, continuationIndicator string) []string {
	if code == "" {
		return []string{""}
	}

	// Fast path: entire code fits on one line
	if fpdf.GetStringWidth(code) <= maxWidth {
		return []string{code}
	}

	breakPoints := findCodeBreakPoints(code)
	runes := []rune(code)
	var segments []string
	var currentSegment strings.Builder
	lastBreakIdx := 0 // last used break point index
	segmentStartPos := 0

	// Build segments character by character, breaking at safe points
	for i := 0; i < len(runes); i++ {
		currentSegment.WriteRune(runes[i])
		segmentText := currentSegment.String()

		// Check if adding continuation indicator would overflow
		testWidth := fpdf.GetStringWidth(segmentText + continuationIndicator)

		if testWidth > maxWidth && currentSegment.Len() > 1 {
			// Need to break. Find the best break point before current position.
			breakPos := findBestBreakPoint(breakPoints, lastBreakIdx, i, segmentStartPos)

			if breakPos >= 0 {
				// Break at the safe point
				var segmentToAdd string
				if breakPos < segmentStartPos {
					// No safe break point found in this segment - force break at current position
					segmentToAdd = string(runes[segmentStartPos:i]) + continuationIndicator
					segmentStartPos = i
				} else {
					bp := breakPoints[breakPos]
					actualBreakPos := bp.position

					if bp.after {
						// Break after this character (include it in current segment)
						segmentToAdd = string(runes[segmentStartPos:actualBreakPos+1]) + continuationIndicator
						segmentStartPos = actualBreakPos + 1
						// Skip spaces at the start of next segment
						for segmentStartPos < len(runes) && runes[segmentStartPos] == ' ' {
							segmentStartPos++
						}
					} else {
						// Break before this character (exclude it from current segment)
						segmentToAdd = string(runes[segmentStartPos:actualBreakPos]) + continuationIndicator
						segmentStartPos = actualBreakPos
					}
				}

				segments = append(segments, segmentToAdd)
				currentSegment.Reset()
				lastBreakIdx = breakPos + 1

				// Rebuild current segment from new start position
				for j := segmentStartPos; j <= i; j++ {
					currentSegment.WriteRune(runes[j])
				}
			} else {
				// No break point found - force break at rune boundary (last resort)
				segmentToAdd := string(runes[segmentStartPos:i]) + continuationIndicator
				segments = append(segments, segmentToAdd)
				currentSegment.Reset()
				currentSegment.WriteRune(runes[i])
				segmentStartPos = i
			}
		}
	}

	// Add remaining text
	if currentSegment.Len() > 0 {
		segments = append(segments, currentSegment.String())
	}

	// Ensure at least one segment
	if len(segments) == 0 {
		segments = []string{code}
	}

	return segments
}

// findBestBreakPoint finds the best break point in the range [startIdx, beforePos).
// Returns the index in breakPoints array, or -1 if no suitable break point exists.
func findBestBreakPoint(breakPoints []codeBreakPoint, startIdx int, beforePos int, segmentStart int) int {
	bestIdx := -1

	for i := startIdx; i < len(breakPoints); i++ {
		bp := breakPoints[i]

		// Only consider break points within current segment
		if bp.position < segmentStart {
			continue
		}

		if bp.position >= beforePos {
			break
		}

		// Prefer later break points (maximize text per line)
		bestIdx = i
	}

	return bestIdx
}

// isBreakableChar returns true if the character is safe to break after in code.
func isBreakableChar(r rune) bool {
	switch r {
	case ' ', '.', '/', '\\', '-', '_':
		return true
	default:
		return false
	}
}

// isDelimiter returns true if the character is a delimiter we can break before.
func isDelimiter(r rune) bool {
	switch r {
	case '(', ')', '[', ']', '{', '}':
		return true
	default:
		return false
	}
}

