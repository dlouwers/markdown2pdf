package pdf

import (
	"strings"
	"unicode/utf8"
)

// FontKind indicates which font family a text segment should be rendered with.
type FontKind int

const (
	// FontKindBody means the segment should be rendered with the body font.
	FontKindBody FontKind = iota
	// FontKindSymbols means the segment should be rendered with the symbols fallback font.
	FontKindSymbols
)

// TextSegment is a contiguous run of text that should be rendered with a
// single font family. The renderer switches fonts between segments.
type TextSegment struct {
	Text string
	Kind FontKind
}

// emojiSubstitutions maps common emoji/symbol runes to ASCII-safe text
// equivalents for use when neither the body nor symbols font has a glyph.
var emojiSubstitutions = map[rune]string{
	// Status indicators
	'✅': "[v]",
	'❌': "[x]",
	'⭐': "[*]",
	'⚠': "[!]",
	'⏭': "[>>]",

	// Common document emoji
	'✓': "[v]",
	'✗': "[x]",
	'☑': "[v]",
	'☒': "[x]",
	'☐': "[ ]",
	'★': "[*]",
	'☆': "[*]",
	'❤': "[<3]",
	'♥': "[<3]",

	// Arrows
	'➡': "->",
	'⬅': "<-",
	'⬆': "^",
	'⬇': "v",
	'↔': "<->",

	// Misc symbols
	'🔴': "(R)",
	'🟢': "(G)",
	'🟡': "(Y)",
	'🔵': "(B)",
	'⚡': "[!]",
	'🔥': "[!]",
	'💯': "[100]",
	'🛑': "[STOP]",
	'🔒': "[locked]",
	'🔓': "[unlocked]",
	'🚀': "[>]",
	'💡': "[i]",
	'🎯': "[o]",
	'📌': "[pin]",
	'📋': "[doc]",
	'🔍': "[?]",
	'🧪': "[test]",
	'📦': "[pkg]",
	'🔗': "[link]",
}

// SegmentText splits text into segments tagged with the font that can render
// them. It checks each rune against the body font first, then the symbols
// font, and falls back to text substitution if neither font has the glyph.
//
// Adjacent runes using the same font are coalesced into a single segment.
func SegmentText(bodyFont, symbolsFont []byte, text string) []TextSegment {
	// Fast path: pure ASCII is always body font.
	if isASCII(text) {
		return []TextSegment{{Text: text, Kind: FontKindBody}}
	}

	hasBody := len(bodyFont) > 0
	hasSymbols := len(symbolsFont) > 0

	var segments []TextSegment
	var buf strings.Builder
	currentKind := FontKindBody

	flush := func(kind FontKind) {
		if buf.Len() > 0 {
			segments = append(segments, TextSegment{Text: buf.String(), Kind: currentKind})
			buf.Reset()
		}
		currentKind = kind
	}

	for i := 0; i < len(text); {
		r, size := utf8.DecodeRuneInString(text[i:])

		// Skip variation selectors (U+FE0E, U+FE0F).
		if r == 0xFE0E || r == 0xFE0F {
			i += size
			continue
		}

		// Determine which font can render this rune.
		bodySupports := hasBody && FontSupportsGlyph(bodyFont, r)
		if bodySupports {
			if currentKind != FontKindBody {
				flush(FontKindBody)
			}
			buf.WriteRune(r)
			i += size
			continue
		}

		symbolsSupports := hasSymbols && FontSupportsGlyph(symbolsFont, r)
		if symbolsSupports {
			if currentKind != FontKindSymbols {
				flush(FontKindSymbols)
			}
			buf.WriteRune(r)
			i += size
			continue
		}

		// Neither font supports the rune. Use text substitution if available,
		// otherwise pass through as body font (will render as .notdef).
		sub, hasSub := emojiSubstitutions[r]
		if hasSub {
			if currentKind != FontKindBody {
				flush(FontKindBody)
			}
			buf.WriteString(sub)
		} else {
			if currentKind != FontKindBody {
				flush(FontKindBody)
			}
			buf.WriteRune(r)
		}
		i += size
	}

	// Flush remaining.
	if buf.Len() > 0 {
		segments = append(segments, TextSegment{Text: buf.String(), Kind: currentKind})
	}

	return segments
}

// SubstituteUnsupportedGlyphs replaces runes that neither the body nor symbols
// font can render with ASCII text equivalents from the substitution table.
// Runes renderable by the symbols font are kept as-is (they will be rendered
// via font fallback). This function is used for text measurement where we
// need a flat string rather than segmented output.
func SubstituteUnsupportedGlyphs(bodyFont, symbolsFont []byte, text string) string {
	// Fast path: if text is pure ASCII, nothing to substitute.
	if isASCII(text) {
		return text
	}

	hasBody := len(bodyFont) > 0
	hasSymbols := len(symbolsFont) > 0
	var b strings.Builder
	b.Grow(len(text))

	for i := 0; i < len(text); {
		r, size := utf8.DecodeRuneInString(text[i:])

		// Skip variation selectors (U+FE0E, U+FE0F).
		if r == 0xFE0E || r == 0xFE0F {
			i += size
			continue
		}

		// Keep if body font or symbols font supports it.
		if (hasBody && FontSupportsGlyph(bodyFont, r)) || (hasSymbols && FontSupportsGlyph(symbolsFont, r)) {
			b.WriteRune(r)
			i += size
			continue
		}

		// Text substitution fallback.
		sub, hasSub := emojiSubstitutions[r]
		if hasSub {
			b.WriteString(sub)
		} else {
			b.WriteRune(r)
		}
		i += size
	}

	return b.String()
}

// isASCII returns true if every byte in s is in the ASCII range [0, 127].
func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > 127 {
			return false
		}
	}
	return true
}
