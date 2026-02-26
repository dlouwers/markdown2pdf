// Package pdf manages PDF document creation, page layout, and styling.
package pdf

import (
	"fmt"

	"github.com/go-pdf/fpdf"
)

type Document struct {
	pdf         *fpdf.Fpdf
	footerCalls int
	baseDir     string
}

func NewDocument() *Document {
	pdfDoc := fpdf.New("P", "mm", "A4", "")
	doc := &Document{pdf: pdfDoc}
	RegisterFonts(pdfDoc)

	pdfDoc.SetMargins(PageMargin, PageMargin, PageMargin)
	pdfDoc.SetAutoPageBreak(true, PageMargin)
	pdfDoc.SetFont(FontBody, "", FontSizeBody)
	pdfDoc.SetTextColor(ColorText.R, ColorText.G, ColorText.B)
	pdfDoc.SetFooterFunc(func() {
		doc.footerCalls++
		pdfDoc.SetY(-PageMargin + 5)
		pdfDoc.SetFont(FontBody, "", 9)
		pdfDoc.SetTextColor(120, 120, 120)
		pdfDoc.CellFormat(0, 10, fmt.Sprintf("%d", pdfDoc.PageNo()), "", 0, "C", false, 0, "")
		pdfDoc.SetFont(FontBody, "", FontSizeBody)
		pdfDoc.SetTextColor(ColorText.R, ColorText.G, ColorText.B)
	})
	pdfDoc.AddPage()

	return doc
}

func (d *Document) Save(path string) error {
	return d.pdf.OutputFileAndClose(path)
}

func (d *Document) PDF() *fpdf.Fpdf {
	return d.pdf
}

func (d *Document) FooterCalls() int {
	return d.footerCalls
}

func (d *Document) BaseDir() string {
	return d.baseDir
}

func (d *Document) SetBaseDir(dir string) {
	d.baseDir = dir
}
