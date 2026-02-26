package pdf

import "github.com/go-pdf/fpdf"

const (
	FontBody = "Helvetica"
	FontCode = "Courier"

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
)

var (
	ColorText       = fpdf.RGBType{R: 20, G: 20, B: 20}
	ColorLink       = fpdf.RGBType{R: 20, G: 90, B: 200}
	ColorBlockquote = fpdf.RGBType{R: 230, G: 230, B: 230}
	ColorCodeFill   = fpdf.RGBType{R: 240, G: 240, B: 240}
)
