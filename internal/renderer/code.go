package renderer

import (
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/go-pdf/fpdf"
	"github.com/yuin/goldmark/ast"

	"github.com/dlouwers/markdown2pdf/internal/diagram"
	"github.com/dlouwers/markdown2pdf/internal/pdf"
)

func renderCodeBlock(state *renderState, node ast.Node, source []byte) {
	codeText := collectCodeText(node, source)
	if codeText == "" {
		return
	}

	// Check for diagram languages before rendering as a code block.
	if fenced, ok := node.(*ast.FencedCodeBlock); ok {
		lang := strings.TrimSpace(string(fenced.Language(source)))
		if renderDiagram(state, codeText, lang) {
			return
		}
	}

	renderStyledCodeBlock(state, node, codeText, source)
}

// renderDiagram attempts to render a diagram for the given language.
// Returns true if the language was a diagram language and rendering was handled.
func renderDiagram(state *renderState, codeText, lang string) bool {
	switch strings.ToLower(lang) {
	case "d2":
		renderD2Diagram(state, codeText)
		return true
	case "mermaid":
		renderMermaidDiagram(state, codeText)
		return true
	default:
		return false
	}
}

func renderD2Diagram(state *renderState, source string) {
	pngData, err := diagram.RenderD2(source)
	if err != nil {
		renderDiagramPlaceholder(state, "D2", err)
		return
	}

	state.fpdf.Ln(pdf.ImageMarginV)
	embedPNGBytes(state, pngData, "d2-diagram")
	state.fpdf.Ln(pdf.ImageMarginV)
}

func renderMermaidDiagram(state *renderState, source string) {
	pngData, err := diagram.RenderMermaid(source)
	if err != nil {
		renderDiagramPlaceholder(state, "Mermaid", err)
		return
	}

	state.fpdf.Ln(pdf.ImageMarginV)
	embedPNGBytes(state, pngData, "mermaid-diagram")
	state.fpdf.Ln(pdf.ImageMarginV)
}

// renderDiagramPlaceholder renders an error placeholder when a diagram fails to render.
func renderDiagramPlaceholder(state *renderState, diagramType string, err error) {
	state.fpdf.Ln(pdf.ImageMarginV)
	renderImagePlaceholder(state, diagramType+" diagram", err.Error())
	state.fpdf.Ln(pdf.ImageMarginV)
}

// codeBlockState tracks rendering state for a code block that may span pages.
type codeBlockState struct {
	state      *renderState
	blockX     float64
	blockWidth float64
	lineHeight float64
	padding    float64
	isFirst    bool // true for the first line (draw top border + top padding)
}

// beginCodeBlock initializes the code block rendering context and draws the
// top padding row of the background.
func beginCodeBlock(state *renderState) *codeBlockState {
	lineHeight := pdf.LineHeight
	padding := pdf.CodeBlockPadding

	leftMargin, _, rightMargin, _ := state.fpdf.GetMargins()
	pageWidth, _ := state.fpdf.GetPageSize()
	blockWidth := pageWidth - leftMargin - rightMargin

	state.fpdf.Ln(pdf.CodeBlockMarginV)

	cb := &codeBlockState{
		state:      state,
		blockX:     leftMargin,
		blockWidth: blockWidth,
		lineHeight: lineHeight,
		padding:    padding,
		isFirst:    true,
	}

	cb.drawTopPadding()
	return cb
}

// drawTopPadding draws the top padding strip of the code block background.
func (cb *codeBlockState) drawTopPadding() {
	cb.state.fpdf.SetDrawColor(pdf.ColorCodeBlockBorder.R, pdf.ColorCodeBlockBorder.G, pdf.ColorCodeBlockBorder.B)
	cb.state.fpdf.SetFillColor(pdf.ColorCodeBlockBG.R, pdf.ColorCodeBlockBG.G, pdf.ColorCodeBlockBG.B)

	x, y := cb.blockX, cb.state.fpdf.GetY()

	// Top border line.
	cb.state.fpdf.Line(x, y, x+cb.blockWidth, y)

	// Top padding row (filled, with left+right borders).
	cb.state.fpdf.SetXY(x, y)
	cb.state.fpdf.CellFormat(cb.blockWidth, cb.padding, "", "LR", 0, "", true, 0, "")
	cb.state.fpdf.SetXY(x, y+cb.padding)

	cb.isFirst = true
}

// drawBottomPadding draws the bottom padding strip and border.
func (cb *codeBlockState) drawBottomPadding() {
	cb.state.fpdf.SetDrawColor(pdf.ColorCodeBlockBorder.R, pdf.ColorCodeBlockBorder.G, pdf.ColorCodeBlockBorder.B)
	cb.state.fpdf.SetFillColor(pdf.ColorCodeBlockBG.R, pdf.ColorCodeBlockBG.G, pdf.ColorCodeBlockBG.B)

	x, y := cb.blockX, cb.state.fpdf.GetY()

	// Bottom padding row.
	cb.state.fpdf.SetXY(x, y)
	cb.state.fpdf.CellFormat(cb.blockWidth, cb.padding, "", "LR", 0, "", true, 0, "")

	// Bottom border line.
	bottomY := y + cb.padding
	cb.state.fpdf.Line(x, bottomY, x+cb.blockWidth, bottomY)
	cb.state.fpdf.SetXY(x, bottomY)
}

// checkPageBreak detects if the next line will exceed the page and, if so,
// closes the current block, adds a page, and reopens the block.
func (cb *codeBlockState) checkPageBreak() {
	_, pageH := cb.state.fpdf.GetPageSize()
	_, _, _, bottomMargin := cb.state.fpdf.GetMargins()
	remaining := pageH - bottomMargin - cb.state.fpdf.GetY()

	if cb.lineHeight+cb.padding > remaining {
		// Close current page's code block with bottom border.
		cb.drawBottomPadding()
		cb.state.fpdf.AddPage()
		// Reopen code block on new page.
		cb.drawTopPadding()
	}
}

// drawLineBackground draws the filled background strip for a code line with
// left and right borders, then positions the cursor for text.
func (cb *codeBlockState) drawLineBackground() {
	cb.state.fpdf.SetFillColor(pdf.ColorCodeBlockBG.R, pdf.ColorCodeBlockBG.G, pdf.ColorCodeBlockBG.B)
	cb.state.fpdf.SetDrawColor(pdf.ColorCodeBlockBorder.R, pdf.ColorCodeBlockBorder.G, pdf.ColorCodeBlockBorder.B)

	x := cb.blockX
	y := cb.state.fpdf.GetY()

	// Draw the line-height strip as background with left+right borders.
	cb.state.fpdf.SetXY(x, y)
	cb.state.fpdf.CellFormat(cb.blockWidth, cb.lineHeight, "", "LR", 0, "", true, 0, "")

	// Position cursor inside the padding for text.
	cb.state.fpdf.SetXY(x+cb.padding, y)
}

func renderStyledCodeBlock(state *renderState, node ast.Node, codeText string, source []byte) {
	lineCount := countCodeLines(codeText)
	if lineCount == 0 {
		return
	}

	cb := beginCodeBlock(state)

	defaultStyle := styles.Get("github")

	if fenced, ok := node.(*ast.FencedCodeBlock); ok {
		lang := strings.TrimSpace(string(fenced.Language(source)))
		renderHighlightedCode(state, cb, codeText, lang, defaultStyle)
	} else {
		renderPlainCode(state, cb, codeText)
	}

	cb.drawBottomPadding()
	state.fpdf.Ln(pdf.CodeBlockMarginV)
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

func renderPlainCode(state *renderState, cb *codeBlockState, codeText string) {
	setCodeFont(state, "", pdf.ColorText)
	writeCodeLines(state, cb, codeText)
}

func renderHighlightedCode(state *renderState, cb *codeBlockState, codeText, lang string, style *chroma.Style) {
	lexer := lexers.Get(lang)
	if lexer == nil {
		lexer = lexers.Fallback
	}

	tokens, err := chroma.Tokenise(lexer, nil, codeText)
	if err != nil {
		renderPlainCode(state, cb, codeText)
		return
	}

	// We need to render the first line's background before any tokens.
	cb.checkPageBreak()
	cb.drawLineBackground()
	lineStarted := true

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

		// A token may contain newlines; split and handle each segment.
		parts := strings.Split(token.Value, "\n")
		for i, part := range parts {
			if i > 0 {
				// End of a code line → advance to next line.
				state.fpdf.Ln(cb.lineHeight)
				cb.checkPageBreak()
				cb.drawLineBackground()
				lineStarted = true
				setCodeFont(state, fontStyle, color) // Restore after background draw.
			}
			if part != "" {
				if !lineStarted {
					lineStarted = true
				}
				state.fpdf.Write(cb.lineHeight, part)
			}
		}
	}

	// Advance past the last line.
	state.fpdf.Ln(cb.lineHeight)
}

// writeCodeLines renders plain (non-highlighted) code line-by-line with
// page-break-aware backgrounds.
func writeCodeLines(state *renderState, cb *codeBlockState, text string) {
	if text == "" {
		return
	}

	parts := strings.Split(text, "\n")
	for i, part := range parts {
		cb.checkPageBreak()
		cb.drawLineBackground()
		if part != "" {
			state.fpdf.Write(cb.lineHeight, part)
		}
		if i < len(parts)-1 {
			state.fpdf.Ln(cb.lineHeight)
		}
	}
	// Advance past the last line.
	state.fpdf.Ln(cb.lineHeight)
}

func setCodeFont(state *renderState, fontStyle string, color fpdf.RGBType) {
	state.fpdf.SetFont(pdf.FontCode, fontStyle, pdf.FontSizeCode)
	state.fpdf.SetTextColor(color.R, color.G, color.B)
}

func fpdfColor(color chroma.Colour) fpdf.RGBType {
	return fpdf.RGBType{R: int(color.Red()), G: int(color.Green()), B: int(color.Blue())}
}
