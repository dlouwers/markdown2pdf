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

	// Point-to-mm conversion: 1pt = 25.4mm / 72 ≈ 0.3528mm.
	// 1em at body font size = FontSizeBody * ptToMM.
	const ptToMM = 25.4 / 72.0
	const em = pdf.FontSizeBody * ptToMM // 1em in mm at body font size

	// LaTeX standard measurements (converted from em to mm).
	const (
		pnumWidth   = 1.55 * em  // LaTeX \@pnumwidth = 1.55em
		indentH1    = 0.0        // No indent for top-level
		indentH2    = 1.5 * em   // LaTeX standard section indent
		indentH3    = 3.8 * em   // LaTeX standard subsection indent
		indentH4    = 7.0 * em   // LaTeX standard subsubsection indent
		indentOther = 10.0 * em  // Deep nesting
	)

	left, _, right, _ := state.fpdf.GetMargins()
	pageWidth, _ := state.fpdf.GetPageSize()
	contentWidth := pageWidth - left - right

	for i, entry := range entries {
		// LaTeX vertical spacing: 1em before top-level sections (except very first).
		if i > 0 && entry.level == 1 {
			state.fpdf.Ln(1.0 * em) // 1em vertical spacing
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

		// Leader dots and page number layout.
		// Layout:  |<indent>|<title>|<gap>|<dots...>|<gap>|<page number>|
		const gap = 0.5 * em // 0.5em gap around dots
		// LaTeX \@dotsep = 4.5mu; leader box = kern(4.5mu) + "." + kern(4.5mu)
		// Center-to-center spacing ≈ 0.77em gives visually distinct dots.
		const dotSpacing = 0.77 * em

		// Page number positioning: use Text() for exact placement.
		// The alias (e.g. "{toc:0}") is wider than the final page number,
		// so we compute positions based on the alias width to ensure dots
		// never overlap. After RegisterAlias replacement, the shorter page
		// number will appear right-padded within the same space.
		state.fpdf.SetFont(pdf.FontBody, "", fontSize)
		aliasWidth := state.fpdf.GetStringWidth(entry.pageAlias)
		// Use the wider of pnumWidth and aliasWidth to reserve enough space.
		reservedWidth := pnumWidth
		if aliasWidth > reservedWidth {
			reservedWidth = aliasWidth
		}
		pageNumX := left + contentWidth - reservedWidth // left edge of reserved space
		dotsStartX := left + indent + titleWidth + gap
		dotsEndX := pageNumX - gap

		// Baseline Y for Text(): CellFormat (used by WriteLinkID) places text at
		// y + 0.5*h + 0.3*fontSize_mm. Text() uses y as the baseline directly.
		// Match CellFormat's baseline so dots align with the title text.
		baselineY := y + 0.5*pdf.LineHeight + 0.3*(fontSize*ptToMM)

		// Render leader dots at evenly-spaced absolute positions.
		if dotsEndX-dotsStartX > dotSpacing*2 {
			dotCharWidth := state.fpdf.GetStringWidth(".")
			x := dotsStartX
			for x+dotCharWidth <= dotsEndX {
				state.fpdf.Text(x, baselineY, ".")
				x += dotSpacing
			}
		}

		// Render page number alias at fixed position using Text().
		// RegisterAlias replaces the alias with the real page number at
		// PDF output time. Text() places at exact coordinates without
		// creating a cell, avoiding width/overflow issues.
		state.fpdf.Text(pageNumX, baselineY, entry.pageAlias)

		// Advance to next line. Text() doesn't move the cursor, so reset Y
		// to the title line's Y before advancing.
		state.fpdf.SetY(y)
		state.fpdf.Ln(pdf.LineHeight + 1)
		state.fpdf.SetTextColor(pdf.ColorText.R, pdf.ColorText.G, pdf.ColorText.B)
	}

	// Start content on a new page after the TOC.
	state.fpdf.AddPage()
	resetFont(state)
}
