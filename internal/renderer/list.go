package renderer

import (
	"fmt"

	gopdf "github.com/go-pdf/fpdf"
	"github.com/yuin/goldmark/ast"
	extast "github.com/yuin/goldmark/extension/ast"

	"github.com/dlouwers/markdown2pdf/internal/pdf"
)

func renderList(state *renderState, list *ast.List, source []byte, depth int) {
	state.fpdf.Ln(1)
	for item := list.FirstChild(); item != nil; item = item.NextSibling() {
		listItem, ok := item.(*ast.ListItem)
		if !ok {
			continue
		}
		renderListItem(state, list, listItem, source, depth)
	}
	state.fpdf.Ln(1)
}

func renderListItem(state *renderState, list *ast.List, item *ast.ListItem, source []byte, depth int) {
	left, top, right, _ := state.fpdf.GetMargins()
	indent := left + pdf.ListIndent + float64(depth)*pdf.ListIndent

	// LaTeX-style fixed text-start position:
	//   |<-indent->|<-labelWidth->|<-labelSep->|text...
	textStart := indent + pdf.ListLabelWidth + pdf.ListLabelSep
	labelRight := indent + pdf.ListLabelWidth // right edge of label box

	// Orphan protection: ensure the bullet + at least the first line of
	// text fit on the current page. Without this check, the bullet is drawn
	// with Text() at an absolute Y position while renderInline's Write()
	// auto-paginates, leaving the bullet stranded on the previous page.
	minItemH := pdf.LineHeight + pdf.ListItemSpacing
	_, pageH := state.fpdf.GetPageSize()
	_, _, _, bottomMargin := state.fpdf.GetMargins()
	remaining := pageH - bottomMargin - state.fpdf.GetY()
	if minItemH > remaining {
		state.fpdf.AddPage()
	}

	state.fpdf.SetX(indent)
	startY := state.fpdf.GetY() + pdf.LineHeight/2

	box, checked := taskListState(item)
	if box {
		drawTaskBox(state, indent, startY, checked)
		state.fpdf.SetX(textStart)
	} else if list.IsOrdered() {
		// Right-align the number label within the label box.
		label := fmt.Sprintf("%d.", listStartNumber(list, item))
		state.fpdf.SetFont(pdf.FontBody, "", pdf.FontSizeBody)
		labelW := state.fpdf.GetStringWidth(label)
		state.fpdf.Text(labelRight-labelW, startY+1, label)
		state.fpdf.SetX(textStart)
	} else {
		drawBullet(state, indent, labelRight, startY, depth)
		state.fpdf.SetX(textStart)
	}

	for child := item.FirstChild(); child != nil; child = child.NextSibling() {
		switch n := child.(type) {
		case *ast.Paragraph:
			state.fpdf.SetMargins(textStart, top, right)
			state.fpdf.SetX(textStart)
			renderInline(state, n, source)
			state.fpdf.Ln(pdf.LineHeight)
			state.fpdf.SetMargins(left, top, right)
		case *ast.List:
			renderList(state, n, source, depth+1)
		default:
			renderInline(state, child, source)
			state.fpdf.Ln(pdf.LineHeight)
		}
	}
	state.fpdf.Ln(pdf.ListItemSpacing)
}

func listStartNumber(list *ast.List, item *ast.ListItem) int {
	start := list.Start
	if start < 1 {
		start = 1
	}
	index := 0
	for child := list.FirstChild(); child != nil; child = child.NextSibling() {
		if child == item {
			break
		}
		index++
	}
	return start + index
}

func drawBullet(state *renderState, labelLeft, labelRight, y float64, depth int) {
	// Pick the bullet character for the nesting depth.
	idx := depth
	if idx >= len(pdf.ListBulletChars) {
		idx = len(pdf.ListBulletChars) - 1
	}
	bulletRune := pdf.ListBulletChars[idx]

	// Center of the label box.
	cx := (labelLeft + labelRight) / 2

	// Try UTF-8 text bullet if the body font supports it.
	if fontBytes := state.doc.BodyFontBytes(); len(fontBytes) > 0 && pdf.FontSupportsGlyph(fontBytes, bulletRune) {
		state.fpdf.SetFont(pdf.FontBody, "", pdf.FontSizeBody)
		bulletW := state.fpdf.GetStringWidth(string(bulletRune))
		state.fpdf.Text(cx-bulletW/2, y+1, string(bulletRune))
		return
	}

	// Try the symbols fallback font.
	if symBytes := state.doc.SymbolsFontBytes(); len(symBytes) > 0 && pdf.FontSupportsGlyph(symBytes, bulletRune) {
		state.fpdf.SetFont(pdf.FontSymbols, "", pdf.FontSizeBody)
		bulletW := state.fpdf.GetStringWidth(string(bulletRune))
		state.fpdf.Text(cx-bulletW/2, y+1, string(bulletRune))
		state.fpdf.SetFont(pdf.FontBody, "", pdf.FontSizeBody) // restore
		return
	}

	// Try the emoji fallback font.
	if emojiBytes := state.doc.EmojiFontBytes(); len(emojiBytes) > 0 && pdf.FontSupportsGlyph(emojiBytes, bulletRune) {
		state.fpdf.SetFont(pdf.FontEmoji, "", pdf.FontSizeBody)
		bulletW := state.fpdf.GetStringWidth(string(bulletRune))
		state.fpdf.Text(cx-bulletW/2, y+1, string(bulletRune))
		state.fpdf.SetFont(pdf.FontBody, "", pdf.FontSizeBody) // restore
		return
	}

	// Geometric fallback when the font lacks the glyph.
	// Sizes derived from actual Noto Sans glyph bounds at 11pt body:
	//   • (U+2022) visible ~0.86×0.96mm  → circle r≈0.45mm
	//   ‣ (U+2023) visible ~1.01×1.14mm  → triangle ≈0.5mm
	//   ⁃ (U+2043) visible ~0.94×0.30mm  → dash 0.9×0.3mm
	switch depth {
	case 0:
		state.fpdf.Circle(cx, y, 0.45, "F")
	case 1:
		// Small right-pointing triangle.
		state.fpdf.Polygon([]gopdf.PointType{
			{X: cx - 0.4, Y: y - 0.5},
			{X: cx + 0.5, Y: y},
			{X: cx - 0.4, Y: y + 0.5},
		}, "F")
	default:
		// Short horizontal dash.
		state.fpdf.Rect(cx-0.45, y-0.15, 0.9, 0.3, "F")
	}
}

func taskListState(item *ast.ListItem) (bool, bool) {
	first := item.FirstChild()
	if first == nil {
		return false, false
	}
	// The container may be a Paragraph (loose list) or a TextBlock (tight list).
	var inline ast.Node
	switch n := first.(type) {
	case *ast.Paragraph:
		inline = n.FirstChild()
	case *ast.TextBlock:
		inline = n.FirstChild()
	default:
		return false, false
	}
	if inline == nil {
		return false, false
	}
	checkbox, ok := inline.(*extast.TaskCheckBox)
	if !ok {
		return false, false
	}
	return true, checkbox.IsChecked
}

func drawTaskBox(state *renderState, x, y float64, checked bool) {
	size := pdf.ListBulletSize * 2
	state.fpdf.Rect(x, y-size/2, size, size, "D")
	if checked {
		state.fpdf.Rect(x+0.5, y-size/2+0.5, size-1, size-1, "F")
	}
}
