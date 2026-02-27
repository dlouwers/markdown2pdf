package renderer

import (
	"github.com/dlouwers/markdown2pdf/internal/parser"
	"github.com/dlouwers/markdown2pdf/internal/pdf"
)

// renderCoverPage renders a professional cover page using frontmatter metadata.
// The cover page is rendered on its own page before any document content.
// Layout: centered title (large), subtitle (if present), author, date, version.
func renderCoverPage(state *renderState, metadata *parser.Metadata) {
	if metadata == nil || metadata.Title == "" {
		return // Cover page requires at least a title
	}

	fpdf := state.fpdf
	pageWidth, pageHeight := fpdf.GetPageSize()
	_, _, rightMargin, _ := fpdf.GetMargins()
	contentWidth := pageWidth - pdf.PageMargin - rightMargin

	// Start Y position for title (upper third of page)
	y := pdf.CoverTitleY

	// Title (large, bold, centered)
	fpdf.SetFont(pdf.FontBody, "B", pdf.FontSizeCoverTitle)
	fpdf.SetY(y)
	fpdf.CellFormat(contentWidth, pdf.FontSizeCoverTitle*0.353, metadata.Title, "", 0, "C", false, 0, "")
	y += pdf.FontSizeCoverTitle*0.353 + pdf.CoverSpacing

	// Subtitle (if present)
	if metadata.Subtitle != "" {
		fpdf.SetFont(pdf.FontBody, "I", pdf.FontSizeCoverSubtitle)
		fpdf.SetY(y)
		fpdf.CellFormat(contentWidth, pdf.FontSizeCoverSubtitle*0.353, metadata.Subtitle, "", 0, "C", false, 0, "")
		// y updated here but immediately overwritten below, so assignment removed
	}

	// Move to middle section for metadata
	y = pageHeight / 2

	// Author
	if metadata.Author != "" {
		fpdf.SetFont(pdf.FontBody, "", pdf.FontSizeCoverMeta)
		fpdf.SetY(y)
		fpdf.CellFormat(contentWidth, pdf.FontSizeCoverMeta*0.353, metadata.Author, "", 0, "C", false, 0, "")
		y += pdf.FontSizeCoverMeta*0.353 + pdf.CoverMetaSpace
	}

	// Date
	if metadata.Date != "" {
		fpdf.SetFont(pdf.FontBody, "", pdf.FontSizeCoverMeta)
		fpdf.SetY(y)
		fpdf.CellFormat(contentWidth, pdf.FontSizeCoverMeta*0.353, metadata.Date, "", 0, "C", false, 0, "")
		y += pdf.FontSizeCoverMeta*0.353 + pdf.CoverMetaSpace
	}

	// Version
	if metadata.Version != "" {
		fpdf.SetFont(pdf.FontBody, "", pdf.FontSizeCoverMeta)
		fpdf.SetY(y)
		fpdf.CellFormat(contentWidth, pdf.FontSizeCoverMeta*0.353, "Version "+metadata.Version, "", 0, "C", false, 0, "")
	}

	// Add new page for actual content
	fpdf.AddPage()

	// Reset font to body defaults
	fpdf.SetFont(pdf.FontBody, "", pdf.FontSizeBody)
}
