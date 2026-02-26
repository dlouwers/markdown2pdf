package renderer

import (
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/go-pdf/fpdf"
	"github.com/yuin/goldmark/ast"

	"github.com/dlouwers/markdown2pdf/internal/pdf"
)

func renderCodeBlock(state *renderState, node ast.Node, source []byte) {
	codeText := collectCodeText(node, source)
	lineCount := countCodeLines(codeText)
	if lineCount == 0 {
		return
	}

	lineHeight := pdf.LineHeight
	padding := pdf.CodeBlockPadding
	margin := pdf.CodeBlockMarginV

	leftMargin, _, rightMargin, _ := state.fpdf.GetMargins()
	pageWidth, _ := state.fpdf.GetPageSize()
	blockWidth := pageWidth - leftMargin - rightMargin
	blockHeight := float64(lineCount)*lineHeight + 2*padding

	state.fpdf.Ln(margin)
	blockX := leftMargin
	blockY := state.fpdf.GetY()

	state.fpdf.SetDrawColor(pdf.ColorCodeBlockBorder.R, pdf.ColorCodeBlockBorder.G, pdf.ColorCodeBlockBorder.B)
	state.fpdf.SetFillColor(pdf.ColorCodeBlockBG.R, pdf.ColorCodeBlockBG.G, pdf.ColorCodeBlockBG.B)
	state.fpdf.Rect(blockX, blockY, blockWidth, blockHeight, "FD")

	startX := blockX + padding
	startY := blockY + padding
	state.fpdf.SetXY(startX, startY)

	defaultStyle := styles.Get("github")

	if fenced, ok := node.(*ast.FencedCodeBlock); ok {
		lang := strings.TrimSpace(string(fenced.Language(source)))
		renderHighlightedCode(state, codeText, lang, defaultStyle)
	} else {
		renderPlainCode(state, codeText)
	}

	state.fpdf.SetXY(blockX, blockY+blockHeight)
	state.fpdf.Ln(margin)
	resetFont(state)
}

func collectCodeText(node ast.Node, source []byte) string {
	lines := node.Lines()
	if lines == nil || lines.Len() == 0 {
		return ""
	}

	var builder strings.Builder
	for i := 0; i < lines.Len(); i++ {
		segment := lines.At(i)
		builder.Write(segment.Value(source))
	}
	return builder.String()
}

func countCodeLines(codeText string) int {
	if codeText == "" {
		return 0
	}
	return strings.Count(codeText, "\n") + 1
}

func renderPlainCode(state *renderState, codeText string) {
	setCodeFont(state, "", pdf.ColorText)
	writeCodeText(state, codeText)
}

func renderHighlightedCode(state *renderState, codeText, lang string, style *chroma.Style) {
	lexer := lexers.Get(lang)
	if lexer == nil {
		lexer = lexers.Fallback
	}

	tokens, err := chroma.Tokenise(lexer, nil, codeText)
	if err != nil {
		renderPlainCode(state, codeText)
		return
	}

	for _, token := range tokens {
		entry := style.Get(token.Type)
		color := pdf.ColorText
		if entry.Colour.IsSet() {
			color = fpdfColor(entry.Colour)
		}

		fontStyle := ""
		if entry.Bold == chroma.Yes {
			fontStyle += "B"
		}
		if entry.Italic == chroma.Yes {
			fontStyle += "I"
		}

		setCodeFont(state, fontStyle, color)
		writeCodeText(state, token.Value)
	}
}

func setCodeFont(state *renderState, fontStyle string, color fpdf.RGBType) {
	state.fpdf.SetFont(pdf.FontCode, fontStyle, pdf.FontSizeCode)
	state.fpdf.SetTextColor(color.R, color.G, color.B)
}

func writeCodeText(state *renderState, text string) {
	if text == "" {
		return
	}

	lineHeight := pdf.LineHeight
	leftMargin, _, _, _ := state.fpdf.GetMargins()
	leftPadding := leftMargin + pdf.CodeBlockPadding

	parts := strings.Split(text, "\n")
	for i, part := range parts {
		if part != "" {
			state.fpdf.Write(lineHeight, part)
		}
		if i < len(parts)-1 {
			state.fpdf.Ln(lineHeight)
			state.fpdf.SetX(leftPadding)
		}
	}
}

func fpdfColor(color chroma.Colour) fpdf.RGBType {
	return fpdf.RGBType{R: color.Red(), G: color.Green(), B: color.Blue()}
}
