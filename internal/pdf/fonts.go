package pdf

import (
	_ "embed"

	"github.com/go-pdf/fpdf"
)

//go:embed fonts/NotoSans-Regular.ttf
var notoSansRegular []byte

//go:embed fonts/NotoSans-Bold.ttf
var notoSansBold []byte

//go:embed fonts/NotoSans-Italic.ttf
var notoSansItalic []byte

//go:embed fonts/NotoSans-BoldItalic.ttf
var notoSansBoldItalic []byte

//go:embed fonts/NotoSansMono-Regular.ttf
var notoSansMonoRegular []byte

//go:embed fonts/NotoSansMono-Bold.ttf
var notoSansMonoBold []byte

// RegisterFonts registers the embedded Noto Sans font families with the PDF
// document. After registration, FontBody and FontCode constants can be used
// with SetFont as usual.
func RegisterFonts(pdf *fpdf.Fpdf) {
	pdf.AddUTF8FontFromBytes(FontBody, "", notoSansRegular)
	pdf.AddUTF8FontFromBytes(FontBody, "B", notoSansBold)
	pdf.AddUTF8FontFromBytes(FontBody, "I", notoSansItalic)
	pdf.AddUTF8FontFromBytes(FontBody, "BI", notoSansBoldItalic)

	pdf.AddUTF8FontFromBytes(FontCode, "", notoSansMonoRegular)
	pdf.AddUTF8FontFromBytes(FontCode, "B", notoSansMonoBold)
}
