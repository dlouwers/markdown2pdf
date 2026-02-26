package renderer

import (
	"fmt"

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
	state.fpdf.SetX(indent)

	startX := indent
	startY := state.fpdf.GetY() + pdf.LineHeight/2

	box, checked := taskListState(item)
	if box {
		drawTaskBox(state, startX, startY, checked)
		state.fpdf.SetX(indent + pdf.ListIndent)
	} else if list.IsOrdered() {
		label := fmt.Sprintf("%d.", listStartNumber(list, item))
		state.fpdf.SetFont(pdf.FontBody, "", pdf.FontSizeBody)
		state.fpdf.Write(pdf.LineHeight, label+" ")
	} else {
		drawBullet(state, startX, startY, depth)
		state.fpdf.SetX(indent + pdf.ListIndent)
	}

	for child := item.FirstChild(); child != nil; child = child.NextSibling() {
		switch n := child.(type) {
		case *ast.Paragraph:
			state.fpdf.SetMargins(indent+pdf.ListIndent, top, right)
			state.fpdf.SetX(indent + pdf.ListIndent)
			renderInline(state, n, source)
			state.fpdf.Ln(pdf.LineHeight)
			state.fpdf.SetMargins(left, top, right)
		case *ast.List:
			renderList(state, n, source, depth+1)
		default:
			if container, ok := child.(ast.Node); ok {
				renderInline(state, container, source)
				state.fpdf.Ln(pdf.LineHeight)
			}
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

func drawBullet(state *renderState, x, y float64, depth int) {
	size := pdf.ListBulletSize
	switch {
	case depth == 0:
		state.fpdf.Circle(x+size, y, size, "F")
	case depth == 1:
		state.fpdf.Circle(x+size, y, size-0.5, "D")
	default:
		state.fpdf.Rect(x+size-0.75, y-0.75, 1.5, 1.5, "F")
	}
}

func taskListState(item *ast.ListItem) (bool, bool) {
	paragraph, ok := item.FirstChild().(*ast.Paragraph)
	if !ok || paragraph == nil {
		return false, false
	}
	first := paragraph.FirstChild()
	if first == nil {
		return false, false
	}
	checkbox, ok := first.(*extast.TaskCheckBox)
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
