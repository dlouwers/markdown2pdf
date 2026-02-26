package renderer

import (
	"math"

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
	state.fpdf.SetFont(pdf.FontBody, "", pdf.FontSizeTable)

	rowHeight := pdf.LineHeight + 2*pdf.TableCellPadding
	alignments := table.Alignments

	if header, ok := table.FirstChild().(*extast.TableHeader); ok {
		renderTableHeader(state, header, source, columnWidths, alignments, rowHeight)
	}

	for row := table.FirstChild(); row != nil; row = row.NextSibling() {
		bodyRow, ok := row.(*extast.TableRow)
		if !ok {
			continue
		}
		renderTableRow(state, bodyRow, source, columnWidths, alignments, rowHeight, false)
	}

	state.fpdf.SetLineWidth(0.2)
	state.fpdf.SetDrawColor(0, 0, 0)
	state.fpdf.Ln(2)
	resetFont(state)
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
			text := collectInlineText(cellNode, source)
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

func renderTableHeader(state *renderState, header *extast.TableHeader, source []byte, widths []float64, alignments []extast.Alignment, rowHeight float64) {
	state.fpdf.SetFillColor(pdf.ColorTableHeader.R, pdf.ColorTableHeader.G, pdf.ColorTableHeader.B)
	state.fpdf.SetFont(pdf.FontBody, "B", pdf.FontSizeTable)
	col := 0
	for cell := header.FirstChild(); cell != nil && col < len(widths); cell = cell.NextSibling() {
		cellNode, ok := cell.(*extast.TableCell)
		if !ok {
			continue
		}
		text := collectInlineText(cellNode, source)
		align := alignmentForColumn(cellNode, alignments, col)
		drawTableCell(state, widths[col], rowHeight, text, align, true)
		col++
	}
	for col < len(widths) {
		drawTableCell(state, widths[col], rowHeight, "", alignStr(extast.AlignNone), true)
		col++
	}
	state.fpdf.Ln(rowHeight)
}

func renderTableRow(state *renderState, row *extast.TableRow, source []byte, widths []float64, alignments []extast.Alignment, rowHeight float64, header bool) {
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
		text := collectInlineText(cellNode, source)
		align := alignmentForColumn(cellNode, alignments, col)
		drawTableCell(state, widths[col], rowHeight, text, align, false)
		col++
	}
	for col < len(widths) {
		drawTableCell(state, widths[col], rowHeight, "", alignStr(extast.AlignNone), false)
		col++
	}
	state.fpdf.Ln(rowHeight)
}

func drawTableCell(state *renderState, width, height float64, text, align string, fill bool) {
	x, y := state.fpdf.GetXY()
	state.fpdf.CellFormat(width, height, "", "1", 0, "", fill, 0, "")

	innerWidth := width - 2*pdf.TableCellPadding
	if innerWidth < 0 {
		innerWidth = 0
	}
	state.fpdf.SetXY(x+pdf.TableCellPadding, y)
	state.fpdf.CellFormat(innerWidth, height, text, "", 0, align, false, 0, "")
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
