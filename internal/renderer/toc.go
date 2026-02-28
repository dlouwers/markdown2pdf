package renderer

import (
	"fmt"

	"github.com/yuin/goldmark/ast"

	"github.com/dlouwers/markdown2pdf/internal/pdf"
)

// tocEntry represents a heading collected for the table of contents.
type tocEntry struct {
	text      string
	level     int
	linkID    int    // fpdf internal link ID pointing to the heading destination
	pageAlias string // Placeholder alias for page number (e.g., "{toc:0}")
}

// collectTOCEntries walks the AST and extracts headings, creating an
// internal link ID for each so the TOC can link to them.
func collectTOCEntries(node ast.Node, source []byte, state *renderState) []tocEntry {
	var entries []tocEntry
	_ = ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) { //nolint:errcheck // callback never returns error
		if !entering {
			return ast.WalkContinue, nil
		}
		heading, ok := n.(*ast.Heading)
		if !ok {
			return ast.WalkContinue, nil
		}
		linkID := state.fpdf.AddLink()
		pageAlias := fmt.Sprintf("{toc:%d}", len(entries))
		entries = append(entries, tocEntry{
			text:      pdf.SubstituteUnsupportedGlyphs(state.doc.BodyFontBytes(), state.doc.SymbolsFontBytes(), state.doc.EmojiFontBytes(), collectInlineText(heading, source)),
			level:     heading.Level,
			linkID:    linkID,
			pageAlias: pageAlias,
		})
		return ast.WalkContinue, nil
	})
	return entries
}

// renderTOC renders a LaTeX-style table of contents page with page numbers,
// leader dots, and proper hierarchical formatting.
func renderTOC(state *renderState, entries []tocEntry) {
	// Store link IDs and aliases for page number registration during content rendering.
	linkIDs := make([]int, len(entries))
	pageAliases := make([]string, len(entries))
	for i := range entries {
		linkIDs[i] = entries[i].linkID
		pageAliases[i] = entries[i].pageAlias
	}
	state.tocLinks = linkIDs
	state.tocPageAliases = pageAliases

	// TOC title with LaTeX spacing.
	state.fpdf.SetFont(pdf.FontBody, "B", pdf.FontSizeH2)
	state.fpdf.MultiCell(0, pdf.LineHeight+1, "Table of Contents", "", "", false)
	state.fpdf.Ln(6) // LaTeX: ~2-3em before first entry

	// LaTeX standard measurements (converted from em to points).
	const (
		pnumWidth   = 1.55 * pdf.FontSizeBody / 11.0 // 1.55em for page numbers
		dotSpacing  = 0.5 * pdf.FontSizeBody / 11.0  // 0.5em between dot centers
		indentH1    = 0.0                             // No indent for top-level
		indentH2    = 1.5 * pdf.FontSizeBody / 11.0  // 1.5em
		indentH3    = 3.8 * pdf.FontSizeBody / 11.0  // 3.8em
		indentH4    = 7.0 * pdf.FontSizeBody / 11.0  // 7.0em
		indentOther = 10.0 * pdf.FontSizeBody / 11.0 // 10em for H5+
	)

	left, _, right, _ := state.fpdf.GetMargins()
	pageWidth, _ := state.fpdf.GetPageSize()
	contentWidth := pageWidth - left - right

	for i, entry := range entries {
		// LaTeX vertical spacing: 1em before top-level sections (except very first).
		if i > 0 && entry.level == 1 {
			state.fpdf.Ln(pdf.FontSizeBody / 11.0 * 1.0) // 1em
		}

		// Calculate indent based on heading level (LaTeX standard).
		var indent float64
		switch entry.level {
		case 1:
			indent = indentH1
		case 2:
			indent = indentH2
		case 3:
			indent = indentH3
		case 4:
			indent = indentH4
		default:
			indent = indentOther
		}

		// Font styling: bold for H1-H2, regular for deeper levels.
		fontSize := pdf.FontSizeBody
		fontStyle := ""
		if entry.level <= 2 {
			fontStyle = "B"
		}
		if entry.level >= 4 {
			fontSize = pdf.FontSizeBody - 1
		}

		// Set position at left margin + indent.
		state.fpdf.SetX(left + indent)
		y := state.fpdf.GetY()

		// Render entry text with font-segment awareness.
		segments := pdf.SegmentText(state.doc.BodyFontBytes(), state.doc.SymbolsFontBytes(), state.doc.EmojiFontBytes(), entry.text)
		state.fpdf.SetTextColor(pdf.ColorText.R, pdf.ColorText.G, pdf.ColorText.B)

		// Calculate title width.
		var titleWidth float64
		for _, seg := range segments {
			switch seg.Kind {
			case pdf.FontKindSymbols:
				state.fpdf.SetFont(pdf.FontSymbols, fontStyle, fontSize)
			case pdf.FontKindEmoji:
				state.fpdf.SetFont(pdf.FontEmoji, fontStyle, fontSize)
			default:
				state.fpdf.SetFont(pdf.FontBody, fontStyle, fontSize)
			}
			titleWidth += state.fpdf.GetStringWidth(seg.Text)
		}

		// Render title as clickable link.
		state.fpdf.SetXY(left+indent, y)
		for _, seg := range segments {
			switch seg.Kind {
			case pdf.FontKindSymbols:
				state.fpdf.SetFont(pdf.FontSymbols, fontStyle, fontSize)
			case pdf.FontKindEmoji:
				state.fpdf.SetFont(pdf.FontEmoji, fontStyle, fontSize)
			default:
				state.fpdf.SetFont(pdf.FontBody, fontStyle, fontSize)
			}
			state.fpdf.WriteLinkID(pdf.LineHeight, seg.Text, entry.linkID)
		}

		// Calculate space for leader dots.
		// Current X position is: left + indent + titleWidth
		// Page number box starts at: left + contentWidth - pnumWidth
		// Dots should fill the gap between title end and page number box start
		const gap = 0.5 * pdf.FontSizeBody / 11.0 // 0.5em gap after title
		dotsStartX := state.fpdf.GetX() + gap
		pageNumStartX := left + contentWidth - pnumWidth
		dotsWidth := pageNumStartX - dotsStartX
		// Render leader dots if space allows.
		if dotsWidth > dotSpacing*2 {
			state.fpdf.SetFont(pdf.FontBody, "", fontSize)
			// Position cursor at start of dots area (after gap)
			state.fpdf.SetX(dotsStartX)
			numDots := int(dotsWidth / dotSpacing)
			dots := ""
			for j := 0; j < numDots; j++ {
				dots += "."
			}
			// Render dots filling the available width, left-aligned within the cell
			state.fpdf.CellFormat(dotsWidth, pdf.LineHeight, dots, "", 0, "L", false, 0, "")
		} else {
			// Not enough space for dots, just position cursor before page number box
			state.fpdf.SetX(pageNumStartX)
		}

		// Render page number alias (will be replaced with actual number during PDF close).
		state.fpdf.SetFont(pdf.FontBody, "", fontSize)
		// Position cursor at exact page number box location
		state.fpdf.SetX(pageNumStartX)
		state.fpdf.CellFormat(pnumWidth, pdf.LineHeight, entry.pageAlias, "", 0, "R", false, 0, "")

		state.fpdf.Ln(pdf.LineHeight + 1)
		state.fpdf.SetTextColor(pdf.ColorText.R, pdf.ColorText.G, pdf.ColorText.B)
	}

	// Start content on a new page after the TOC.
	state.fpdf.AddPage()
	resetFont(state)
}
