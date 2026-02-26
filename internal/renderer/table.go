package renderer

import (
	"math"
	"strings"

	"github.com/yuin/goldmark/ast"
	extast "github.com/yuin/goldmark/extension/ast"

	"github.com/dlouwers/markdown2pdf/internal/pdf"
)

const minTableColumnWidth = 15.0

func renderTable(state *renderState, table *extast.Table, source []byte) {
	metrics := measureTableMetrics(state, table, source)
	if len(metrics) == 0 {
		return
	}

	left, _, right, _ := state.fpdf.GetMargins()
	pageWidth, _ := state.fpdf.GetPageSize()
	contentWidth := pageWidth - left - right

	// CSS 2.1-style auto table layout: distribute available width using
	// min-content (longest word) and max-content (no-wrap) widths.
	columnWidths := distributeColumnWidths(metrics, contentWidth)

	state.fpdf.Ln(2)
	state.fpdf.SetLineWidth(pdf.TableBorderWidth)
	state.fpdf.SetDrawColor(pdf.ColorTableBorder.R, pdf.ColorTableBorder.G, pdf.ColorTableBorder.B)

	alignments := table.Alignments

	// Extract header node for potential re-rendering on page breaks.
	var header *extast.TableHeader
	var firstDataRow *extast.TableRow
	if h, ok := table.FirstChild().(*extast.TableHeader); ok {
		header = h
		// Find the first data row to enforce orphan prevention:
		// the header must never appear alone at the bottom of a page.
		for sib := header.NextSibling(); sib != nil; sib = sib.NextSibling() {
			if r, ok := sib.(*extast.TableRow); ok {
				firstDataRow = r
				break
			}
		}
	}

	_, pageH := state.fpdf.GetPageSize()
	_, _, _, bottomMargin := state.fpdf.GetMargins()

	// Orphan-header prevention: ensure header + at least one data row
	// fit on the current page before rendering anything.
	if header != nil {
		state.fpdf.SetFont(pdf.FontBody, "B", pdf.FontSizeTable)
		headerH := calcRowHeight(state, header, source, columnWidths)
		minFirstRow := pdf.LineHeight + 2*pdf.TableCellPadding // fallback
		if firstDataRow != nil {
			state.fpdf.SetFont(pdf.FontBody, "", pdf.FontSizeTable)
			minFirstRow = calcRowHeight(state, firstDataRow, source, columnWidths)
		}
		remaining := pageH - bottomMargin - state.fpdf.GetY()
		if headerH+minFirstRow > remaining {
			state.fpdf.AddPage()
			state.fpdf.Ln(2)
			state.fpdf.SetLineWidth(pdf.TableBorderWidth)
			state.fpdf.SetDrawColor(pdf.ColorTableBorder.R, pdf.ColorTableBorder.G, pdf.ColorTableBorder.B)
		}
		renderTableSection(state, header, source, columnWidths, alignments, true)
	}


	for row := table.FirstChild(); row != nil; row = row.NextSibling() {
		bodyRow, ok := row.(*extast.TableRow)
		if !ok {
			continue
		}

		state.fpdf.SetFont(pdf.FontBody, "", pdf.FontSizeTable)
		rowH := calcRowHeight(state, bodyRow, source, columnWidths)

		// Check if the row fits on the current page; if not, break and
		// re-render the header row on the new page for readability.
		remaining := pageH - bottomMargin - state.fpdf.GetY()
		if rowH > remaining {
			state.fpdf.AddPage()
			state.fpdf.SetLineWidth(pdf.TableBorderWidth)
			state.fpdf.SetDrawColor(pdf.ColorTableBorder.R, pdf.ColorTableBorder.G, pdf.ColorTableBorder.B)
			if header != nil {
				renderTableSection(state, header, source, columnWidths, alignments, true)
			}
		}

		state.fpdf.SetFont(pdf.FontBody, "", pdf.FontSizeTable)
		renderRowCells(state, bodyRow, source, columnWidths, alignments, rowH, false)
	}

	state.fpdf.SetLineWidth(0.2)
	state.fpdf.SetDrawColor(0, 0, 0)
	state.fpdf.Ln(2)
	resetFont(state)
}

// renderTableSection renders a header or body row with auto-calculated height.
func renderTableSection(state *renderState, row ast.Node, source []byte, widths []float64, alignments []extast.Alignment, header bool) {
	if header {
		state.fpdf.SetFont(pdf.FontBody, "B", pdf.FontSizeTable)
		state.fpdf.SetFillColor(pdf.ColorTableHeader.R, pdf.ColorTableHeader.G, pdf.ColorTableHeader.B)
	} else {
		state.fpdf.SetFont(pdf.FontBody, "", pdf.FontSizeTable)
	}
	rowH := calcRowHeight(state, row, source, widths)
	renderRowCells(state, row, source, widths, alignments, rowH, header)
}

// calcRowHeight computes the height of a table row by splitting each cell's
// text into wrapped lines and returning the tallest cell height (plus padding).
func calcRowHeight(state *renderState, row ast.Node, source []byte, widths []float64) float64 {
	maxLines := 1
	col := 0
	for cell := row.FirstChild(); cell != nil && col < len(widths); cell = cell.NextSibling() {
		cellNode, ok := cell.(*extast.TableCell)
		if !ok {
			continue
		}
		text := pdf.SubstituteUnsupportedGlyphs(state.doc.BodyFontBytes(), state.doc.SymbolsFontBytes(), state.doc.EmojiFontBytes(), collectInlineText(cellNode, source))
		innerW := widths[col] - 2*pdf.TableCellPadding
		if innerW < 1 {
			innerW = 1
		}
		lines := splitTextLines(state.fpdf, text, innerW)
		if len(lines) > maxLines {
			maxLines = len(lines)
		}
		col++
	}
	return float64(maxLines)*pdf.LineHeight + 2*pdf.TableCellPadding
}

// renderRowCells renders all cells in a row with the given pre-computed height.
func renderRowCells(state *renderState, row ast.Node, source []byte, widths []float64, alignments []extast.Alignment, rowHeight float64, header bool) {
	if header {
		state.fpdf.SetFont(pdf.FontBody, "B", pdf.FontSizeTable)
	} else {
		state.fpdf.SetFont(pdf.FontBody, "", pdf.FontSizeTable)
	}

	col := 0
	for cell := row.FirstChild(); cell != nil && col < len(widths); cell = cell.NextSibling() {
		cellNode, ok := cell.(*extast.TableCell)
		if !ok {
			continue
		}
		text := pdf.SubstituteUnsupportedGlyphs(state.doc.BodyFontBytes(), state.doc.SymbolsFontBytes(), state.doc.EmojiFontBytes(), collectInlineText(cellNode, source))
		align := alignmentForColumn(cellNode, alignments, col)
		drawTableCell(state, widths[col], rowHeight, text, align, header)
		col++
	}

	// Fill remaining empty columns.
	for col < len(widths) {
		drawTableCell(state, widths[col], rowHeight, "", alignStr(extast.AlignNone), header)
		col++
	}
	state.fpdf.Ln(rowHeight)
}

// columnMetrics holds min-content and max-content widths for a column.
type columnMetrics struct {
	minW float64 // longest unbreakable word + padding
	maxW float64 // widest cell without any wrapping + padding
}

// measureTableMetrics computes per-column min-content (longest word) and
// max-content (no-wrap) widths for CSS 2.1-style auto table layout.
func measureTableMetrics(state *renderState, table *extast.Table, source []byte) []columnMetrics {
	columnCount := tableColumnCount(table)
	if columnCount == 0 {
		return nil
	}

	metrics := make([]columnMetrics, columnCount)
	for i := range metrics {
		metrics[i].minW = minTableColumnWidth
		metrics[i].maxW = minTableColumnWidth
	}

	state.fpdf.SetFont(pdf.FontBody, "", pdf.FontSizeTable)
	for row := table.FirstChild(); row != nil; row = row.NextSibling() {
		col := 0
		for cell := row.FirstChild(); cell != nil && col < columnCount; cell = cell.NextSibling() {
			cellNode, ok := cell.(*extast.TableCell)
			if !ok {
				continue
			}
			text := pdf.SubstituteUnsupportedGlyphs(state.doc.BodyFontBytes(), state.doc.SymbolsFontBytes(), state.doc.EmojiFontBytes(), collectInlineText(cellNode, source))
			padding := 2 * pdf.TableCellPadding

			// Max-content: full text width without wrapping.
			maxW := state.fpdf.GetStringWidth(text) + padding
			metrics[col].maxW = math.Max(metrics[col].maxW, maxW)

			// Min-content: longest single word (break at spaces and hyphens).
			for _, word := range splitWords(text) {
				wW := state.fpdf.GetStringWidth(word) + padding
				metrics[col].minW = math.Max(metrics[col].minW, wW)
			}
			col++
		}
	}

	// Ensure minW never exceeds maxW.
	for i := range metrics {
		if metrics[i].minW > metrics[i].maxW {
			metrics[i].minW = metrics[i].maxW
		}
	}

	return metrics
}

// distributeColumnWidths implements CSS 2.1 auto table layout.
// 1. If total max-content fits, use max-content widths (no wrapping needed).
// 2. If total min-content exceeds table width, use proportional min widths.
// 3. Otherwise, give each column its min-content width, then distribute the
//    remaining space proportionally to each column's flex (max - min).
//    This ensures narrow columns keep their natural width while wider
//    columns absorb compression, preventing mid-word breaks.
func distributeColumnWidths(metrics []columnMetrics, tableWidth float64) []float64 {
	n := len(metrics)
	widths := make([]float64, n)

	totalMin := 0.0
	totalMax := 0.0
	for _, m := range metrics {
		totalMin += m.minW
		totalMax += m.maxW
	}

	// Case 1: everything fits without wrapping.
	if totalMax <= tableWidth {
		// Distribute extra space proportionally to max-content.
		for i, m := range metrics {
			widths[i] = m.maxW
		}
		extra := tableWidth - totalMax
		if extra > 0 && totalMax > 0 {
			for i, m := range metrics {
				widths[i] += extra * (m.maxW / totalMax)
			}
		}
		return widths
	}

	// Case 2: even minimum widths don't fit — scale proportionally.
	if totalMin >= tableWidth {
		scale := tableWidth / totalMin
		for i, m := range metrics {
			widths[i] = m.minW * scale
		}
		return widths
	}

	// Case 3: table needs wrapping. Give every column its min-content,
	// then distribute remaining space proportionally to flex (max - min).
	// Narrow columns that need less space get proportionally more of their
	// preferred width, while wide columns absorb the compression.
	totalFlex := 0.0
	for i, m := range metrics {
		widths[i] = m.minW
		totalFlex += m.maxW - m.minW
	}
	remaining := tableWidth - totalMin
	if totalFlex > 0 && remaining > 0 {
		for i, m := range metrics {
			flex := m.maxW - m.minW
			widths[i] += remaining * (flex / totalFlex)
		}
	}

	return widths
}

// splitWords breaks text at spaces and hyphens (with the hyphen kept at the
// end of the preceding segment), matching fpdf.SplitLines break behavior.
func splitWords(text string) []string {
	var words []string
	var current strings.Builder
	for _, r := range text {
		switch r {
		case ' ':
			if current.Len() > 0 {
				words = append(words, current.String())
				current.Reset()
			}
		case '-':
			current.WriteRune(r)
			words = append(words, current.String())
			current.Reset()
		default:
			current.WriteRune(r)
		}
	}
	if current.Len() > 0 {
		words = append(words, current.String())
	}
	return words
}

func tableColumnCount(table *extast.Table) int {
	count := len(table.Alignments)
	if count > 0 {
		return count
	}

	for row := table.FirstChild(); row != nil; row = row.NextSibling() {
		columns := 0
		for cell := row.FirstChild(); cell != nil; cell = cell.NextSibling() {
			if _, ok := cell.(*extast.TableCell); ok {
				columns++
			}
		}
		if columns > count {
			count = columns
		}
	}

	return count
}

func drawTableCell(state *renderState, width, height float64, text, align string, fill bool) {
	x, y := state.fpdf.GetXY()

	// Draw cell border and optional fill.
	state.fpdf.CellFormat(width, height, "", "1", 0, "", fill, 0, "")

	innerWidth := width - 2*pdf.TableCellPadding
	if innerWidth < 1 {
		innerWidth = 1
	}

	// Font style: bold for header rows, regular for body rows.
	bodyStyle := ""
	if fill {
		bodyStyle = "B"
	}

	// Split text into wrapped lines and render each with font-segment awareness.
	lines := splitTextLines(state.fpdf, text, innerWidth)
	textY := y + pdf.TableCellPadding
	for _, lineStr := range lines {
		segments := pdf.SegmentText(state.doc.BodyFontBytes(), state.doc.SymbolsFontBytes(), state.doc.EmojiFontBytes(), lineStr)

		// Calculate total rendered width for alignment.
		totalW := 0.0
		for _, seg := range segments {
			switch seg.Kind {
			case pdf.FontKindSymbols:
				state.fpdf.SetFont(pdf.FontSymbols, bodyStyle, pdf.FontSizeTable)
			case pdf.FontKindEmoji:
				state.fpdf.SetFont(pdf.FontEmoji, bodyStyle, pdf.FontSizeTable)
			default:
				state.fpdf.SetFont(pdf.FontBody, bodyStyle, pdf.FontSizeTable)
			}
			totalW += state.fpdf.GetStringWidth(seg.Text)
		}

		// Determine starting X based on alignment.
		var startX float64
		switch align {
		case "CM":
			startX = x + pdf.TableCellPadding + (innerWidth-totalW)/2
		case "RM":
			startX = x + pdf.TableCellPadding + innerWidth - totalW
		default: // LM
			startX = x + pdf.TableCellPadding
		}

		// Render each segment with the appropriate font.
		curX := startX
		for _, seg := range segments {
			switch seg.Kind {
			case pdf.FontKindSymbols:
				state.fpdf.SetFont(pdf.FontSymbols, bodyStyle, pdf.FontSizeTable)
			case pdf.FontKindEmoji:
				state.fpdf.SetFont(pdf.FontEmoji, bodyStyle, pdf.FontSizeTable)
			default:
				state.fpdf.SetFont(pdf.FontBody, bodyStyle, pdf.FontSizeTable)
			}
			state.fpdf.Text(curX, textY+pdf.LineHeight*0.75, seg.Text)
			curX += state.fpdf.GetStringWidth(seg.Text)
		}
		textY += pdf.LineHeight
	}

	// Restore X to after this cell for the next cell.
	state.fpdf.SetXY(x+width, y)
}

func alignmentForColumn(cell *extast.TableCell, alignments []extast.Alignment, index int) string {
	if cell.Alignment != extast.AlignNone {
		return alignStr(cell.Alignment)
	}
	if index < len(alignments) {
		return alignStr(alignments[index])
	}
	return alignStr(extast.AlignNone)
}

func alignStr(a extast.Alignment) string {
	switch a {
	case extast.AlignCenter:
		return "CM"
	case extast.AlignRight:
		return "RM"
	case extast.AlignLeft:
		return "LM"
	default:
		return "LM"
	}
}
