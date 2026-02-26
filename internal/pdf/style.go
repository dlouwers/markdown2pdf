package pdf

import "github.com/go-pdf/fpdf"

const (
	FontBody = "NotoSans"
	FontCode = "NotoSansMono"

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
