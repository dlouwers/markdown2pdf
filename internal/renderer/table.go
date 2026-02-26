package renderer

import (
	"math"

	"github.com/yuin/goldmark/ast"
	extast "github.com/yuin/goldmark/extension/ast"

	"github.com/dlouwers/markdown2pdf/internal/pdf"
)

const minTableColumnWidth = 20.0

func renderTable(state *renderState, table *extast.Table, source []byte) {
	columnWidths := measureTableColumns(state, table, source)
	if len(columnWidths) == 0 {
		return
	}

	left, _, right, _ := state.fpdf.GetMargins()
	pageWidth, _ := state.fpdf.GetPageSize()
	contentWidth := pageWidth - left - right

	totalWidth := 0.0
	for _, width := range columnWidths {
		totalWidth += width
	}
	if totalWidth > contentWidth {
		scale := contentWidth / totalWidth
		for i, width := range columnWidths {
			columnWidths[i] = width * scale
		}
	}

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
		lines := state.fpdf.SplitLines([]byte(text), innerW)
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

func measureTableColumns(state *renderState, table *extast.Table, source []byte) []float64 {
	columnCount := tableColumnCount(table)
	if columnCount == 0 {
		return nil
	}

	widths := make([]float64, columnCount)
	for i := range widths {
		widths[i] = minTableColumnWidth
	}

	state.fpdf.SetFont(pdf.FontBody, "", pdf.FontSizeTable)
	for row := table.FirstChild(); row != nil; row = row.NextSibling() {
		cells := row.FirstChild()
		column := 0
		for cell := cells; cell != nil && column < columnCount; cell = cell.NextSibling() {
			cellNode, ok := cell.(*extast.TableCell)
			if !ok {
				continue
			}
			text := pdf.SubstituteUnsupportedGlyphs(state.doc.BodyFontBytes(), state.doc.SymbolsFontBytes(), state.doc.EmojiFontBytes(), collectInlineText(cellNode, source))
			width := state.fpdf.GetStringWidth(text) + 2*pdf.TableCellPadding
			widths[column] = math.Max(widths[column], width)
			column++
		}
	}

	return widths
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
	lines := state.fpdf.SplitLines([]byte(text), innerWidth)
	textY := y + pdf.TableCellPadding
	for _, line := range lines {
		lineStr := string(line)
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
