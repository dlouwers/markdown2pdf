package renderer

import (
	"math"
	"strings"

	"github.com/yuin/goldmark/ast"
	extast "github.com/yuin/goldmark/extension/ast"

	"github.com/dlouwers/markdown2pdf/internal/pdf"
)

const minTableColumnWidth = 15.0

// tableRenderMetrics holds calculated font size, padding, and line height for a table.
type tableRenderMetrics struct {
	fontSize   float64
	padding    float64
	lineHeight float64
}

// calculateTableMetrics returns proportional font size, padding, and line height
// based on column count. Follows LaTeX and CSS typography best practices:
//  - 1-6 columns: full size (10pt font, 0.5em padding, 1.3× leading)
//  - 7-8 columns: 90% reduction (~9pt font)
//  - 9+ columns: 80% reduction (~8pt font)
func calculateTableMetrics(columnCount int) tableRenderMetrics {
	var scaleFactor float64
	switch {
	case columnCount >= 9:
		scaleFactor = 0.8 // 80% for 9+ columns
	case columnCount >= 7:
		scaleFactor = 0.9 // 90% for 7-8 columns
	default:
		scaleFactor = 1.0 // Full size for 1-6 columns
	}

	fontSize := pdf.FontSizeTable * scaleFactor
	// 0.5em horizontal padding (CSS standard for table cells)
	padding := fontSize * 0.353 * 0.5
	// 1.3× leading for tables (tighter than body text, standard for tabular data)
	lineHeight := fontSize * 0.353 * 1.3

	return tableRenderMetrics{
		fontSize:   fontSize,
		padding:    padding,
		lineHeight: lineHeight,
	}
}

func renderTable(state *renderState, table *extast.Table, source []byte) {
	metrics := measureTableMetrics(state, table, source)
	if len(metrics) == 0 {
		return
	}

	left, _, right, _ := state.fpdf.GetMargins()
	pageWidth, _ := state.fpdf.GetPageSize()
	contentWidth := pageWidth - left - right

	columnCount := len(metrics)
	
	// Calculate proportional font size, padding, and line height based on column count
	tm := calculateTableMetrics(columnCount)
	
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
		state.fpdf.SetFont(pdf.FontBody, "B", tm.fontSize)
		headerH := calcRowHeight(state, header, source, columnWidths, tm)
		minFirstRow := tm.lineHeight + 2*tm.padding // fallback
		if firstDataRow != nil {
			state.fpdf.SetFont(pdf.FontBody, "", tm.fontSize)
			minFirstRow = calcRowHeight(state, firstDataRow, source, columnWidths, tm)
		}
		remaining := pageH - bottomMargin - state.fpdf.GetY()
		if headerH+minFirstRow > remaining {
			state.fpdf.AddPage()
			state.fpdf.Ln(2)
			state.fpdf.SetLineWidth(pdf.TableBorderWidth)
			state.fpdf.SetDrawColor(pdf.ColorTableBorder.R, pdf.ColorTableBorder.G, pdf.ColorTableBorder.B)
		}
		renderTableSection(state, header, source, columnWidths, alignments, true, tm)
	}

	for row := table.FirstChild(); row != nil; row = row.NextSibling() {
		bodyRow, ok := row.(*extast.TableRow)
		if !ok {
			continue
		}

		state.fpdf.SetFont(pdf.FontBody, "", tm.fontSize)
		rowH := calcRowHeight(state, bodyRow, source, columnWidths, tm)

		// Check if the row fits on the current page; if not, break and
		// re-render the header row on the new page for readability.
		remaining := pageH - bottomMargin - state.fpdf.GetY()
		if rowH > remaining {
			state.fpdf.AddPage()
			state.fpdf.SetLineWidth(pdf.TableBorderWidth)
			state.fpdf.SetDrawColor(pdf.ColorTableBorder.R, pdf.ColorTableBorder.G, pdf.ColorTableBorder.B)
			if header != nil {
			renderTableSection(state, header, source, columnWidths, alignments, true, tm)
			}
		}

		state.fpdf.SetFont(pdf.FontBody, "", tm.fontSize)
		renderRowCells(state, bodyRow, source, columnWidths, alignments, rowH, false, tm)
	}

	state.fpdf.SetLineWidth(0.2)
	state.fpdf.SetDrawColor(0, 0, 0)
	state.fpdf.Ln(2)
	resetFont(state)
}

// renderTableSection renders a header or body row with auto-calculated height.
func renderTableSection(state *renderState, row ast.Node, source []byte, widths []float64, alignments []extast.Alignment, header bool, tm tableRenderMetrics) {
	if header {
		state.fpdf.SetFont(pdf.FontBody, "B", tm.fontSize)
		state.fpdf.SetFillColor(pdf.ColorTableHeader.R, pdf.ColorTableHeader.G, pdf.ColorTableHeader.B)
	} else {
		state.fpdf.SetFont(pdf.FontBody, "", tm.fontSize)
	}
	rowH := calcRowHeight(state, row, source, widths, tm)
	renderRowCells(state, row, source, widths, alignments, rowH, header, tm)
}

// calcRowHeight computes the height of a table row by splitting each cell's
// text into wrapped lines and returning the tallest cell height (plus padding).
func calcRowHeight(state *renderState, row ast.Node, source []byte, widths []float64, tm tableRenderMetrics) float64 {
	maxLines := 1
	col := 0
	for cell := row.FirstChild(); cell != nil && col < len(widths); cell = cell.NextSibling() {
		cellNode, ok := cell.(*extast.TableCell)
		if !ok {
			continue
		}
		text := pdf.SubstituteUnsupportedGlyphs(state.doc.BodyFontBytes(), state.doc.SymbolsFontBytes(), state.doc.EmojiFontBytes(), collectInlineText(cellNode, source))
		innerW := widths[col] - 2*tm.padding
		if innerW < 1 {
			innerW = 1
		}
		lines := splitTextLines(state.fpdf, text, innerW)
		if len(lines) > maxLines {
			maxLines = len(lines)
		}
		col++
	}
	return float64(maxLines)*tm.lineHeight + 2*tm.padding
}

// renderRowCells renders all cells in a row with the given pre-computed height.
func renderRowCells(state *renderState, row ast.Node, source []byte, widths []float64, alignments []extast.Alignment, rowHeight float64, header bool, tm tableRenderMetrics) {
	if header {
		state.fpdf.SetFont(pdf.FontBody, "B", tm.fontSize)
	} else {
		state.fpdf.SetFont(pdf.FontBody, "", tm.fontSize)
	}

	col := 0
	for cell := row.FirstChild(); cell != nil && col < len(widths); cell = cell.NextSibling() {
		cellNode, ok := cell.(*extast.TableCell)
		if !ok {
			continue
		}
		text := pdf.SubstituteUnsupportedGlyphs(state.doc.BodyFontBytes(), state.doc.SymbolsFontBytes(), state.doc.EmojiFontBytes(), collectInlineText(cellNode, source))
		align := alignmentForColumn(cellNode, alignments, col)
		drawTableCell(state, widths[col], rowHeight, text, align, header, tm)
		col++
	}

	// Fill remaining empty columns.
	for col < len(widths) {
		drawTableCell(state, widths[col], rowHeight, "", alignStr(extast.AlignNone), header, tm)
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

	// Calculate metrics for current column count to get proper font size
	tm := calculateTableMetrics(columnCount)
	state.fpdf.SetFont(pdf.FontBody, "", tm.fontSize)
	for row := table.FirstChild(); row != nil; row = row.NextSibling() {
		col := 0
		for cell := row.FirstChild(); cell != nil && col < columnCount; cell = cell.NextSibling() {
			cellNode, ok := cell.(*extast.TableCell)
			if !ok {
				continue
			}
			text := pdf.SubstituteUnsupportedGlyphs(state.doc.BodyFontBytes(), state.doc.SymbolsFontBytes(), state.doc.EmojiFontBytes(), collectInlineText(cellNode, source))
			padding := 2 * tm.padding

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
//  1. If total max-content fits, use max-content widths (no wrapping needed).
//  2. If total min-content exceeds table width, use proportional min widths.
//  3. Otherwise, give each column its min-content width, then distribute the
//     remaining space proportionally to each column's flex (max - min).
//     This ensures narrow columns keep their natural width while wider
//     columns absorb compression, preventing mid-word breaks.
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

func drawTableCell(state *renderState, width, height float64, text, align string, fill bool, tm tableRenderMetrics) {
	x, y := state.fpdf.GetXY()

	// Draw cell border and optional fill.
	state.fpdf.CellFormat(width, height, "", "1", 0, "", fill, 0, "")

	innerWidth := width - 2*tm.padding
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
	textY := y + tm.padding
	for _, lineStr := range lines {
		segments := pdf.SegmentText(state.doc.BodyFontBytes(), state.doc.SymbolsFontBytes(), state.doc.EmojiFontBytes(), lineStr)

		// Calculate total rendered width for alignment.
		totalW := 0.0
		for _, seg := range segments {
			switch seg.Kind {
			case pdf.FontKindSymbols:
				state.fpdf.SetFont(pdf.FontSymbols, bodyStyle, tm.fontSize)
				totalW += state.fpdf.GetStringWidth(seg.Text)
			case pdf.FontKindEmoji:
				state.fpdf.SetFont(pdf.FontEmoji, bodyStyle, tm.fontSize)
				// Substitute SMP emoji for width calculation
				segText := pdf.SubstituteUnsupportedGlyphs(state.doc.BodyFontBytes(), state.doc.SymbolsFontBytes(), state.doc.EmojiFontBytes(), seg.Text)
				totalW += state.fpdf.GetStringWidth(segText)
			default:
				state.fpdf.SetFont(pdf.FontBody, bodyStyle, tm.fontSize)
				totalW += state.fpdf.GetStringWidth(seg.Text)
			}
		}

		// Determine starting X based on alignment.
		var startX float64
		switch align {
		case "CM":
			startX = x + tm.padding + (innerWidth-totalW)/2
		case "RM":
			startX = x + tm.padding + innerWidth - totalW
		default: // LM
			startX = x + tm.padding
		}

		// Render each segment with the appropriate font.
		curX := startX
		for _, seg := range segments {
			switch seg.Kind {
			case pdf.FontKindSymbols:
				state.fpdf.SetFont(pdf.FontSymbols, bodyStyle, tm.fontSize)
				state.fpdf.Text(curX, textY+tm.lineHeight*0.75, seg.Text)
				curX += state.fpdf.GetStringWidth(seg.Text)
			case pdf.FontKindEmoji:
				state.fpdf.SetFont(pdf.FontEmoji, bodyStyle, tm.fontSize)
				// Substitute SMP emoji to prevent fpdf panic (tables can't embed PNGs inline yet)
				segText := pdf.SubstituteUnsupportedGlyphs(state.doc.BodyFontBytes(), state.doc.SymbolsFontBytes(), state.doc.EmojiFontBytes(), seg.Text)
				state.fpdf.Text(curX, textY+tm.lineHeight*0.75, segText)
				curX += state.fpdf.GetStringWidth(segText)
			default:
				state.fpdf.SetFont(pdf.FontBody, bodyStyle, tm.fontSize)
				state.fpdf.Text(curX, textY+tm.lineHeight*0.75, seg.Text)
				curX += state.fpdf.GetStringWidth(seg.Text)
			}
		}
		textY += tm.lineHeight
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
