package renderer

import (
	"math"

	"github.com/dlouwers/markdown2pdf/internal/parser"
	"github.com/dlouwers/markdown2pdf/internal/pdf"

	gopdf "github.com/go-pdf/fpdf"
)
// renderCoverPage renders a professional cover page using frontmatter metadata.
// The cover page is rendered on its own page before any document content.
// Layout: centered title (large), subtitle (if present), author, date, version.

// calculateCoverFontSizes determines optimal font sizes for title and subtitle
// based on line wrap counts. Reduces font size when titles wrap to 3+ lines
// following LaTeX and typography best practices.
func calculateCoverFontSizes(fpdf *gopdf.Fpdf, metadata *parser.Metadata, contentWidth float64) (titleSize, subtitleSize float64) {
	// Measure title at full size to count lines
	fpdf.SetFont(pdf.FontBody, "B", pdf.FontSizeCoverTitle)
	titleLines := splitTextLines(fpdf, metadata.Title, contentWidth)

	// Calculate scale factor based on line count
	// 1-2 lines: full size (1.0)
	// 3+ lines: reduce by 12% per extra line beyond 2, capped at 70%
	scaleFactor := 1.0
	if len(titleLines) > 2 {
		scaleFactor = 1.0 - (0.12 * float64(len(titleLines)-2))
		scaleFactor = math.Max(scaleFactor, 0.7)
	}

	titleSize = pdf.FontSizeCoverTitle * scaleFactor
	subtitleSize = pdf.FontSizeCoverSubtitle * scaleFactor
	return
}
func renderCoverPage(state *renderState, metadata *parser.Metadata) {
	if metadata == nil || metadata.Title == "" {
		return // Cover page requires at least a title
	}

	fpdf := state.fpdf
	pageWidth, pageHeight := fpdf.GetPageSize()
	leftMargin, _, rightMargin, _ := fpdf.GetMargins()
	contentWidth := pageWidth - leftMargin - rightMargin

	// Start Y position for title (upper third of page)
	y := pdf.CoverTitleY

	// Calculate optimal font sizes based on title length
	titleFontSize, subtitleFontSize := calculateCoverFontSizes(fpdf, metadata, contentWidth)

	// Title (large, bold, centered) - wrap at word boundaries following LaTeX conventions
	fpdf.SetFont(pdf.FontBody, "B", titleFontSize)
	titleLines := splitTextLines(fpdf, metadata.Title, contentWidth)
	lineHeight := titleFontSize * 0.353 // ~12.7mm for 36pt font
	for _, line := range titleLines {
		fpdf.SetY(y)
		fpdf.CellFormat(contentWidth, lineHeight, line, "", 0, "C", false, 0, "")
		y += lineHeight
	}
	y += pdf.CoverSpacing

	// Subtitle (if present)
	if metadata.Subtitle != "" {
		fpdf.SetFont(pdf.FontBody, "I", subtitleFontSize)
		subtitleLines := splitTextLines(fpdf, metadata.Subtitle, contentWidth)
		subtitleLineHeight := subtitleFontSize * 0.353
		for _, line := range subtitleLines {
			fpdf.SetY(y)
			fpdf.CellFormat(contentWidth, subtitleLineHeight, line, "", 0, "C", false, 0, "")
			y += subtitleLineHeight
		}
	}

	// Position metadata below title/subtitle block with minimum spacing
	// Use dynamic positioning to prevent overlap when title/subtitle wrap to many lines
	titleBottom := y
	minSpacing := subtitleFontSize * 0.353 * 1.5 // 1.5em spacing (LaTeX standard)
	metadataY := titleBottom + minSpacing

	// Use the larger of: calculated position or traditional center position
	y = metadataY
	if pageHeight/2 > metadataY {
		y = pageHeight / 2
	}

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
