package renderer

import (
	"strings"

	"github.com/go-pdf/fpdf"
	"github.com/yuin/goldmark/ast"
	extast "github.com/yuin/goldmark/extension/ast"

	"github.com/dlouwers/markdown2pdf/internal/pdf"
)

type renderState struct {
	doc   *pdf.Document
	fpdf  *fpdf.Fpdf
	style textStyle
	stack []textStyle
}

type textStyle struct {
	bold     bool
	italic   bool
	code     bool
	linkDest string
}

func newRenderState(doc *pdf.Document) *renderState {
	return &renderState{doc: doc, fpdf: doc.PDF()}
}

func renderNode(state *renderState, node ast.Node, source []byte) error {
	return ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		switch n := n.(type) {
		case *ast.Document:
			return ast.WalkContinue, nil
		case *ast.Heading:
			if entering {
				renderHeading(state, n, source)
			}
			return ast.WalkSkipChildren, nil
		case *ast.Paragraph:
			if entering {
				renderParagraph(state, n, source)
			}
			return ast.WalkSkipChildren, nil
		case *ast.ThematicBreak:
			if entering {
				renderThematicBreak(state)
			}
			return ast.WalkSkipChildren, nil
		case *ast.FencedCodeBlock:
			if entering {
				renderCodeBlock(state, n, source)
			}
			return ast.WalkSkipChildren, nil
		case *ast.CodeBlock:
			if entering {
				renderCodeBlock(state, n, source)
			}
			return ast.WalkSkipChildren, nil
		case *ast.List:
			if entering {
				renderList(state, n, source, 0)
			}
			return ast.WalkSkipChildren, nil
		case *extast.Table:
			if entering {
				renderTable(state, n, source)
			}
			return ast.WalkSkipChildren, nil
		case *ast.Blockquote:
			if entering {
				renderBlockquote(state, n, source)
			}
			return ast.WalkSkipChildren, nil
		}
		return ast.WalkContinue, nil
	})
}

func renderHeading(state *renderState, heading *ast.Heading, source []byte) {
	state.fpdf.Ln(2)
	size := headingFontSize(heading.Level)
	state.fpdf.SetFont(pdf.FontBody, "B", size)
	text := collectInlineText(heading, source)
	state.fpdf.MultiCell(0, pdf.LineHeight+1, text, "", "", false)
	state.fpdf.Ln(2)
	resetFont(state)
}

func renderParagraph(state *renderState, paragraph *ast.Paragraph, source []byte) {
	state.fpdf.SetFont(pdf.FontBody, "", pdf.FontSizeBody)
	state.fpdf.Ln(1)
	renderInline(state, paragraph, source)
	state.fpdf.Ln(pdf.LineHeight)
	resetFont(state)
}

func renderThematicBreak(state *renderState) {
	left, _, right, _ := state.fpdf.GetMargins()
	width, _ := state.fpdf.GetPageSize()
	y := state.fpdf.GetY() + 3
	state.fpdf.SetDrawColor(180, 180, 180)
	state.fpdf.Line(left, y, width-right, y)
	state.fpdf.Ln(6)
	state.fpdf.SetDrawColor(0, 0, 0)
}

func renderBlockquote(state *renderState, blockquote *ast.Blockquote, source []byte) {
	left, top, right, bottom := state.fpdf.GetMargins()
	state.fpdf.SetMargins(left+pdf.BlockquoteIndent, top, right)

	startY := state.fpdf.GetY()
	barX := left
	barY := startY

	_ = ast.Walk(blockquote, func(n ast.Node, entering bool) (ast.WalkStatus, error) { //nolint:errcheck // callback never returns error
		if !entering {
			return ast.WalkContinue, nil
		}
		switch n := n.(type) {
		case *ast.Paragraph:
			renderParagraph(state, n, source)
			return ast.WalkSkipChildren, nil
		case *ast.Heading:
			renderHeading(state, n, source)
			return ast.WalkSkipChildren, nil
		case *ast.ThematicBreak:
			renderThematicBreak(state)
			return ast.WalkSkipChildren, nil
		}
		return ast.WalkContinue, nil
	})

	endY := state.fpdf.GetY()
	if endY < barY+1 {
		endY = barY + 1
	}
	state.fpdf.SetFillColor(pdf.ColorBlockquote.R, pdf.ColorBlockquote.G, pdf.ColorBlockquote.B)
	state.fpdf.Rect(barX, barY, pdf.BlockquoteBar, endY-barY, "F")

	state.fpdf.SetMargins(left, top, right)
	state.fpdf.SetAutoPageBreak(true, bottom)
}

func renderInline(state *renderState, container ast.Node, source []byte) {
	for child := container.FirstChild(); child != nil; child = child.NextSibling() {
		renderInlineNode(state, child, source)
	}
}

func renderInlineNode(state *renderState, node ast.Node, source []byte) {
	switch n := node.(type) {
	case *ast.Text:
		text := string(n.Value(source))
		if !strings.Contains(text, "\n") {
			writeText(state, text)
			if n.SoftLineBreak() {
				state.fpdf.Ln(pdf.LineHeight)
			}
			return
		}
		parts := strings.Split(text, "\n")
		for i, part := range parts {
			if part != "" {
				writeText(state, part)
			}
			if i < len(parts)-1 {
				state.fpdf.Ln(pdf.LineHeight)
			}
		}
		if n.SoftLineBreak() {
			state.fpdf.Ln(pdf.LineHeight)
		}
	case *ast.Emphasis:
		state.pushStyle()
		switch n.Level {
		case 1:
			state.style.italic = true
		case 2:
			state.style.bold = true
		default:
			state.style.bold = true
			state.style.italic = true
		}
		applyFont(state)
		renderInline(state, n, source)
		state.popStyle()
		applyFont(state)
	case *ast.CodeSpan:
		state.pushStyle()
		state.style.code = true
		applyFont(state)
		codeText := collectCodeSpanText(n, source)
		writeCode(state, codeText)
		state.popStyle()
		applyFont(state)
	case *ast.Link:
		dest := string(n.Destination)
		state.pushStyle()
		state.style.linkDest = dest
		applyFont(state)
		renderInline(state, n, source)
		state.popStyle()
		applyFont(state)
	case *ast.AutoLink:
		dest := string(n.URL(source))
		state.pushStyle()
		state.style.linkDest = dest
		applyFont(state)
		text := string(n.Label(source))
		writeText(state, text)
		state.popStyle()
		applyFont(state)
	case *ast.Image:
		renderImage(state, n, source)
	default:
		renderInline(state, node, source)
	}
}

func applyFont(state *renderState) {
	style := ""
	if state.style.bold {
		style += "B"
	}
	if state.style.italic {
		style += "I"
	}

	family := pdf.FontBody
	size := pdf.FontSizeBody
	if state.style.code {
		family = pdf.FontCode
		size = pdf.FontSizeBody - 1
		style = ""
	}

	if state.style.linkDest != "" {
		style += "U"
		state.fpdf.SetTextColor(pdf.ColorLink.R, pdf.ColorLink.G, pdf.ColorLink.B)
	} else {
		state.fpdf.SetTextColor(pdf.ColorText.R, pdf.ColorText.G, pdf.ColorText.B)
	}

	state.fpdf.SetFont(family, style, size)
}

func writeText(state *renderState, text string) {
	if text == "" {
		return
	}
	applyFont(state)
	if state.style.linkDest != "" {
		state.fpdf.WriteLinkString(pdf.LineHeight, text, state.style.linkDest)
		return
	}
	state.fpdf.Write(pdf.LineHeight, text)
}

func writeCode(state *renderState, text string) {
	if text == "" {
		return
	}
	state.fpdf.SetFillColor(pdf.ColorCodeFill.R, pdf.ColorCodeFill.G, pdf.ColorCodeFill.B)
	width := state.fpdf.GetStringWidth(text) + 2
	state.fpdf.CellFormat(width, pdf.LineHeight, text, "", 0, "", true, 0, "")
	state.fpdf.SetFillColor(255, 255, 255)
}

func headingFontSize(level int) float64 {
	switch level {
	case 1:
		return pdf.FontSizeH1
	case 2:
		return pdf.FontSizeH2
	case 3:
		return pdf.FontSizeH3
	case 4:
		return pdf.FontSizeH4
	case 5:
		return pdf.FontSizeH5
	default:
		return pdf.FontSizeH6
	}
}

func collectInlineText(container ast.Node, source []byte) string {
	var builder strings.Builder
	var walk func(n ast.Node)
	walk = func(n ast.Node) {
		switch n.Kind() {
		case ast.KindText:
			if t, ok := n.(*ast.Text); ok {
				builder.WriteString(string(t.Value(source)))
			}
		case ast.KindCodeSpan:
			builder.WriteString(collectCodeSpanText(n, source))
		default:
			for child := n.FirstChild(); child != nil; child = child.NextSibling() {
				walk(child)
			}
		}
	}
	for child := container.FirstChild(); child != nil; child = child.NextSibling() {
		walk(child)
	}
	return builder.String()
}

// collectCodeSpanText gathers text content from a CodeSpan node's children.
// This avoids using the deprecated n.Text(source) method on CodeSpan nodes.
func collectCodeSpanText(n ast.Node, source []byte) string {
	var builder strings.Builder
	for child := n.FirstChild(); child != nil; child = child.NextSibling() {
		if t, ok := child.(*ast.Text); ok {
			builder.WriteString(string(t.Value(source)))
		}
	}
	return builder.String()
}

func resetFont(state *renderState) {
	state.style = textStyle{}
	applyFont(state)
}

func (state *renderState) pushStyle() {
	state.stack = append(state.stack, state.style)
}

func (state *renderState) popStyle() {
	if len(state.stack) == 0 {
		state.style = textStyle{}
		return
	}
	state.style = state.stack[len(state.stack)-1]
	state.stack = state.stack[:len(state.stack)-1]
}
