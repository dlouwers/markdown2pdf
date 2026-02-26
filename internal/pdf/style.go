package pdf

import "github.com/go-pdf/fpdf"

const (
	FontBody = "NotoSans"
	FontCode = "NotoSansMono"
	FontSymbols = "NotoSansSymbols2"

	FontSizeBody = 11.0
	FontSizeH1   = 24.0
	FontSizeH2   = 20.0
	FontSizeH3   = 16.0
	FontSizeH4   = 14.0
	FontSizeH5   = 12.0
	FontSizeH6   = 11.0

	LineHeight = 6.0

	PageMargin = 20.0

	BlockquoteIndent = 10.0
	BlockquoteBar    = 2.0

	FontSizeCode     = 9.0
	CodeBlockPadding = 6.0
	CodeBlockMarginV = 4.0

	ListIndent      = 7.0  // left margin increase per nesting level (LaTeX ~2.5em)
	ListLabelWidth  = 5.0  // width of the label box (right-aligned labels)
	ListLabelSep    = 2.0  // gap between label box and text (LaTeX \labelsep ≈ 5pt)
	ListBulletSize  = 2.0
	ListItemSpacing = 2.0

	TableCellPadding = 3.0
	TableBorderWidth = 0.3
	FontSizeTable    = 10.0

	ImageMarginV    = 4.0
	SVGRenderScale  = 2.0

	// Heading spacing (LaTeX article.cls inspired, ~1.5:1 before:after ratio).
	// LaTeX \section uses 3.5ex/2.3ex (~1.52:1), CSS uses 1:1 symmetric.
	// We use ~1.4–1.5:1 as a balanced professional compromise.
	// Before: space above heading (proximity break from previous content).
	// After:  space below heading (closer to its own content).
	HeadingBeforeH1 = 14.0 // ~40pt
	HeadingBeforeH2 = 12.0 // ~34pt
	HeadingBeforeH3 = 10.0 // ~28pt
	HeadingBeforeH4 = 8.0  // ~23pt
	HeadingBeforeH5 = 6.0  // ~17pt
	HeadingBeforeH6 = 5.0  // ~14pt

	HeadingAfterH1 = 10.0 // ~28pt (ratio 1.40:1)
	HeadingAfterH2 = 8.0  // ~23pt (ratio 1.50:1)
	HeadingAfterH3 = 6.5  // ~18pt (ratio 1.54:1)
	HeadingAfterH4 = 5.5  // ~16pt (ratio 1.45:1)
	HeadingAfterH5 = 4.0  // ~11pt (ratio 1.50:1)
	HeadingAfterH6 = 3.5  // ~10pt (ratio 1.43:1)

	// Heading line height multiplier (tighter than body text for visual impact).
	HeadingLeading = 1.25
)

var (
	// ListBulletChars are the UTF-8 bullet characters by nesting depth.
	ListBulletChars = []rune{'•', '‣', '⁃'}

	ColorText            = fpdf.RGBType{R: 20, G: 20, B: 20}
	ColorLink            = fpdf.RGBType{R: 20, G: 90, B: 200}
	ColorBlockquote      = fpdf.RGBType{R: 230, G: 230, B: 230}
	ColorCodeFill        = fpdf.RGBType{R: 240, G: 240, B: 240}
	ColorCodeBlockBG     = fpdf.RGBType{R: 245, G: 245, B: 245}
	ColorCodeBlockBorder = fpdf.RGBType{R: 220, G: 220, B: 220}
	ColorTableHeader     = fpdf.RGBType{R: 240, G: 240, B: 240}
	ColorTableBorder     = fpdf.RGBType{R: 180, G: 180, B: 180}
)

// HeadingSpaceBefore returns the vertical space (mm) to insert above a heading.
func HeadingSpaceBefore(level int) float64 {
	switch level {
	case 1:
		return HeadingBeforeH1
	case 2:
		return HeadingBeforeH2
	case 3:
		return HeadingBeforeH3
	case 4:
		return HeadingBeforeH4
	case 5:
		return HeadingBeforeH5
	default:
		return HeadingBeforeH6
	}
}

// HeadingSpaceAfter returns the vertical space (mm) to insert below a heading.
func HeadingSpaceAfter(level int) float64 {
	switch level {
	case 1:
		return HeadingAfterH1
	case 2:
		return HeadingAfterH2
	case 3:
		return HeadingAfterH3
	case 4:
		return HeadingAfterH4
	case 5:
		return HeadingAfterH5
	default:
		return HeadingAfterH6
	}
}

// HeadingLineHeight returns the line height (mm) for multi-line headings.
// It uses a tighter leading (1.25×) than body text for visual impact.
func HeadingLineHeight(fontSize float64) float64 {
	// Convert font size from pt to mm (1pt ≈ 0.353mm), then apply leading.
	return fontSize * 0.353 * HeadingLeading
}
