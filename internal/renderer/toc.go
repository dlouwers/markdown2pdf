package renderer

import (
	"github.com/yuin/goldmark/ast"

	"github.com/dlouwers/markdown2pdf/internal/pdf"
)

// tocEntry represents a heading collected for the table of contents.
type tocEntry struct {
	text   string
	level  int
	linkID int // fpdf internal link ID pointing to the heading destination
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
		entries = append(entries, tocEntry{
		text:   pdf.SubstituteUnsupportedGlyphs(state.doc.BodyFontBytes(), state.doc.SymbolsFontBytes(), state.doc.EmojiFontBytes(), collectInlineText(heading, source)),
			level:  heading.Level,
			linkID: linkID,
		})
		return ast.WalkContinue, nil
	})
	return entries
}

// renderTOC renders a table of contents page with clickable entries that
// link to the corresponding headings in the document.
func renderTOC(state *renderState, entries []tocEntry) {
	// Store link IDs so renderHeading can set destinations.
	linkIDs := make([]int, len(entries))
	for i, e := range entries {
		linkIDs[i] = e.linkID
	}
	state.tocLinks = linkIDs

	// TOC title.
	state.fpdf.SetFont(pdf.FontBody, "B", pdf.FontSizeH2)
	state.fpdf.MultiCell(0, pdf.LineHeight+1, "Table of Contents", "", "", false)
	state.fpdf.Ln(4)

	for _, entry := range entries {
		indent := float64(entry.level-1) * 6.0 // indent sub-headings
		left, _, _, _ := state.fpdf.GetMargins()
		state.fpdf.SetX(left + indent)

		// Size: H1/H2 entries use body size; deeper headings slightly smaller.
		fontSize := pdf.FontSizeBody
		fontStyle := ""
		if entry.level <= 2 {
			fontStyle = "B"
		}
		if entry.level >= 4 {
			fontSize = pdf.FontSizeBody - 1
		}

		state.fpdf.SetTextColor(pdf.ColorLink.R, pdf.ColorLink.G, pdf.ColorLink.B)

		// Render TOC entry with font-segment awareness for symbols.
		segments := pdf.SegmentText(state.doc.BodyFontBytes(), state.doc.SymbolsFontBytes(), state.doc.EmojiFontBytes(), entry.text)
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

		state.fpdf.Ln(pdf.LineHeight + 1)
		state.fpdf.SetTextColor(pdf.ColorText.R, pdf.ColorText.G, pdf.ColorText.B)
	}

	// Start content on a new page after the TOC.
	state.fpdf.AddPage()
	resetFont(state)
}
