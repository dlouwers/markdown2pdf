package pdf

import (
	"strings"
	"unicode/utf8"

	"github.com/dlouwers/markdown2pdf/internal/emoji"
)

// FontKind indicates which font family a text segment should be rendered with.
type FontKind int

const (
	// FontKindBody means the segment should be rendered with the body font.
	FontKindBody FontKind = iota
	// FontKindSymbols means the segment should be rendered with the symbols fallback font.
	FontKindSymbols
	// FontKindEmoji means the segment should be rendered with the emoji fallback font.
	FontKindEmoji
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

	// Arrows (Unicode Arrows block U+2190-U+21FF — not in our Noto fonts)
	'\u2190': "<-",  // ← LEFTWARDS ARROW
	'\u2191': "^",   // ↑ UPWARDS ARROW
	'\u2192': "->",  // → RIGHTWARDS ARROW
	'\u2193': "v",   // ↓ DOWNWARDS ARROW
	'\u21D0': "<=",  // ⇐ LEFTWARDS DOUBLE ARROW
	'\u21D1': "^^",  // ⇑ UPWARDS DOUBLE ARROW
	'\u21D2': "=>",  // ⇒ RIGHTWARDS DOUBLE ARROW
	'\u21D3': "vv",  // ⇓ DOWNWARDS DOUBLE ARROW
	'\u21D4': "<=>", // ⇔ LEFT RIGHT DOUBLE ARROW
	'\u21A9': "<-'", // ↩ LEFTWARDS ARROW WITH HOOK
	'\u21AA': "'->", // ↪ RIGHTWARDS ARROW WITH HOOK
	// Heavy/filled arrows (Dingbats/Misc blocks)
	'\u27A1': "->",  // ➡ BLACK RIGHTWARDS ARROW
	'\u2B05': "<-",  // ⬅ LEFT BLACK ARROW
	'\u2B06': "^",   // ⬆ UPWARDS BLACK ARROW
	'\u2B07': "v",   // ⬇ DOWNWARDS BLACK ARROW
	'\u2194': "<->", // ↔ LEFT RIGHT ARROW

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
func SegmentText(bodyFont, symbolsFont, emojiFont []byte, text string) []TextSegment {
	// Fast path: pure ASCII is always body font.
	if isASCII(text) {
		return []TextSegment{{Text: text, Kind: FontKindBody}}
	}

	hasBody := len(bodyFont) > 0
	hasSymbols := len(symbolsFont) > 0
	hasEmoji := len(emojiFont) > 0
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

		// Check if emoji font can render this rune
		emojiSupports := hasEmoji && FontSupportsGlyph(emojiFont, r)
		// Also route common SMP emoji (>U+FFFF) to emoji segment for PNG rendering attempt,
		// but let BMP emoji fall through to normal font checks (they can use fonts safely)
		if !emojiSupports && emoji.IsCommonEmoji(r) && r > 0xFFFF {
			emojiSupports = true
		}
		if emojiSupports {
			if currentKind != FontKindEmoji {
				flush(FontKindEmoji)
			}
			buf.WriteRune(r)
			i += size
			continue
		}

		// Neither font supports the rune. Use text substitution if available,
		// otherwise skip it (fpdf cannot render codepoints > 0xFFFF).
		sub, hasSub := emojiSubstitutions[r]
		if hasSub {
			if currentKind != FontKindBody {
				flush(FontKindBody)
			}
			buf.WriteString(sub)
		} else if r <= 0xFFFF {
			// Only pass through BMP characters to avoid fpdf panic.
			// Characters > 0xFFFF will be silently dropped.
			if currentKind != FontKindBody {
				flush(FontKindBody)
			}
			buf.WriteRune(r)
		}
		// else: skip SMP character entirely
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
func SubstituteUnsupportedGlyphs(bodyFont, symbolsFont, emojiFont []byte, text string) string {
	// Fast path: if text is pure ASCII, nothing to substitute.
	if isASCII(text) {
		return text
	}

	hasBody := len(bodyFont) > 0
	hasSymbols := len(symbolsFont) > 0
	hasEmoji := len(emojiFont) > 0
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
		if (hasBody && FontSupportsGlyph(bodyFont, r)) || (hasSymbols && FontSupportsGlyph(symbolsFont, r)) || (hasEmoji && FontSupportsGlyph(emojiFont, r)) {
			b.WriteRune(r)
			i += size
			continue
		}

		// Text substitution fallback.
		sub, hasSub := emojiSubstitutions[r]
		if hasSub {
			b.WriteString(sub)
		} else if r <= 0xFFFF {
			// Keep BMP characters even without substitution (fonts might support them)
			b.WriteRune(r)
		}
		// Skip SMP characters (>U+FFFF) without substitution - fpdf cannot handle them
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
