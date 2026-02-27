package renderer

import (
	"strings"

	"github.com/go-pdf/fpdf"
	"github.com/yuin/goldmark/ast"
	extast "github.com/yuin/goldmark/extension/ast"

	"github.com/dlouwers/markdown2pdf/internal/emoji"
	"github.com/dlouwers/markdown2pdf/internal/pdf"
)

type renderState struct {
	doc        *pdf.Document
	fpdf       *fpdf.Fpdf
	style      textStyle
	stack      []textStyle
	tocLinks   []int // link IDs for TOC; consumed in heading order
	tocLinkIdx int   // next index into tocLinks
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

// determineMinLinesAfterHeading inspects what follows a heading to determine
// the minimum number of body lines to keep with the heading (orphan protection).
// Based on LaTeX/TeX standards: minimum 2 lines for regular headings, 3 for major.
func determineMinLinesAfterHeading(heading *ast.Heading) int {
	// Find what follows this heading by traversing siblings.
	nextSibling := findNextSibling(heading)
	if nextSibling == nil {
		// Heading at end of document - no orphan risk.
		return 0
	}

	switch nextSibling.(type) {
	case *ast.Heading:
		// Consecutive headings: keep them together by requiring space for both.
		// Use penalty = height of next heading (converted to line equivalents).
		// For simplicity, use 3 lines minimum to ensure visual grouping.
		return 3
	case *extast.Table, *ast.FencedCodeBlock, *ast.CodeBlock:
		// Heading followed by block element: keep heading with block start.
		// Use 2 lines minimum (LaTeX standard).
		return 2
	case *ast.Paragraph, *ast.List, *ast.Blockquote:
		// Heading followed by text content: use LaTeX standard.
		// H1/H2 get 3 lines minimum (major headings), H3-H6 get 2 lines.
		if heading.Level <= 2 {
			return 3 // Major headings
		}
		return 2 // Minor headings
	default:
		// Unknown content type: use conservative 2 lines.
		return 2
	}
}

// findNextSibling returns the next sibling node in the AST, or nil if none exists.
func findNextSibling(node ast.Node) ast.Node {
	parent := node.Parent()
	if parent == nil {
		return nil
	}

	// Walk through parent's children to find this node, then return the next one.
	found := false
	for child := parent.FirstChild(); child != nil; child = child.NextSibling() {
		if found {
			return child
		}
		if child == node {
			found = true
		}
	}
	return nil
}

func renderHeading(state *renderState, heading *ast.Heading, source []byte) {
	size := headingFontSize(heading.Level)
	lineH := pdf.HeadingLineHeight(size)
	spaceBefore := pdf.HeadingSpaceBefore(heading.Level)
	spaceAfter := pdf.HeadingSpaceAfter(heading.Level)

	// Pre-calculate heading text height for orphan protection.
	state.fpdf.SetFont(pdf.FontBody, "B", size)
	text := pdf.SubstituteUnsupportedGlyphs(state.doc.BodyFontBytes(), state.doc.SymbolsFontBytes(), state.doc.EmojiFontBytes(), collectInlineText(heading, source))
	left, _, right, _ := state.fpdf.GetMargins()
	pageW, _ := state.fpdf.GetPageSize()
	textWidth := pageW - left - right
	lines := splitTextLines(state.fpdf, text, textWidth)
	headingHeight := spaceBefore + float64(len(lines))*lineH + spaceAfter

	// Enhanced orphan protection following LaTeX/TeX standards:
	// - Ensure heading + minimum 2-3 body lines fit (not just 1)
	// - Handle consecutive headings (keep them together)
	// - Handle heading followed by block elements
	minLinesAfter := determineMinLinesAfterHeading(heading)

	// Only apply orphan protection if there's content after the heading.
	// A heading at EOF has minLinesAfter=0 and needs no protection.
	if minLinesAfter > 0 {
		needed := headingHeight + float64(minLinesAfter)*pdf.LineHeight

		_, topMargin, _, bottomMargin := state.fpdf.GetMargins()
		_, pageH := state.fpdf.GetPageSize()
		maxPageH := pageH - topMargin - bottomMargin
		remaining := pageH - bottomMargin - state.fpdf.GetY()

		// Only break if we're not already near the top and there isn't enough room.
		if needed > remaining && remaining < maxPageH-1 {
			state.fpdf.AddPage()
		}
	}

	// Set TOC link destination at the heading position.
	if state.tocLinkIdx < len(state.tocLinks) {
		state.fpdf.SetLink(state.tocLinks[state.tocLinkIdx], state.fpdf.GetY(), -1)
		state.tocLinkIdx++
	}

	state.fpdf.Ln(spaceBefore)
	state.fpdf.SetFont(pdf.FontBody, "B", size)
	for i, line := range lines {
		state.fpdf.Write(lineH, string(line))
		if i < len(lines)-1 {
			state.fpdf.Ln(lineH)
		}
	}
	state.fpdf.Ln(spaceAfter)
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
			if n.HardLineBreak() || n.SoftLineBreak() {
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
		if n.HardLineBreak() || n.SoftLineBreak() {
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
	segments := pdf.SegmentText(state.doc.BodyFontBytes(), state.doc.SymbolsFontBytes(), state.doc.EmojiFontBytes(), text)
	for _, seg := range segments {
		switch seg.Kind {
		case pdf.FontKindSymbols:
			writeSymbolsSegment(state, seg.Text)
		case pdf.FontKindEmoji:
			writeEmojiSegment(state, seg.Text)
		default:
			applyFont(state)
			if state.style.linkDest != "" {
				state.fpdf.WriteLinkString(pdf.LineHeight, seg.Text, state.style.linkDest)
			} else {
				state.fpdf.Write(pdf.LineHeight, seg.Text)
			}
		}
	}
}

// writeSymbolsSegment renders text using the symbols fallback font, preserving
// the current bold/italic style indicators.
func writeSymbolsSegment(state *renderState, text string) {
	style := ""
	if state.style.bold {
		style += "B"
	}
	if state.style.italic {
		style += "I"
	}
	size := pdf.FontSizeBody
	if state.style.linkDest != "" {
		style += "U"
		state.fpdf.SetTextColor(pdf.ColorLink.R, pdf.ColorLink.G, pdf.ColorLink.B)
	}
	state.fpdf.SetFont(pdf.FontSymbols, style, size)
	if state.style.linkDest != "" {
		state.fpdf.WriteLinkString(pdf.LineHeight, text, state.style.linkDest)
	} else {
		state.fpdf.Write(pdf.LineHeight, text)
	}
	applyFont(state) // restore body font
}

// writeEmojiSegment renders text using emoji PNGs for common emoji (when available),
// falling back to the emoji font or text substitution for others.
func writeEmojiSegment(state *renderState, text string) {
	style := ""
	if state.style.bold {
		style += "B"
	}
	if state.style.italic {
		style += "I"
	}
	size := pdf.FontSizeBody
	if state.style.linkDest != "" {
		style += "U"
		state.fpdf.SetTextColor(pdf.ColorLink.R, pdf.ColorLink.G, pdf.ColorLink.B)
	}

	// Process each rune individually
	for _, r := range text {
		// Try PNG rendering for common emoji first
		if emoji.IsCommonEmoji(r) {
			codepoint := emoji.ToTwemojiCodepoint(r)
			if pngData, err := emoji.GetPNG(codepoint); err == nil {
				if embedEmojiInline(state, pngData, r) {
					continue // Success - skip to next rune
				}
			}
		}

		// PNG failed or not a common emoji - use font/substitution
		// Substitute SMP characters (>U+FFFF) to prevent fpdf panic
		var fallbackText string
		if r > 0xFFFF {
			// Use text substitution for SMP characters
			fallbackText = pdf.SubstituteUnsupportedGlyphs(
				state.doc.BodyFontBytes(),
				state.doc.SymbolsFontBytes(),
				state.doc.EmojiFontBytes(),
				string(r),
			)
		} else {
			// BMP character - safe to render with font
			fallbackText = string(r)
		}

		state.fpdf.SetFont(pdf.FontEmoji, style, size)
		if state.style.linkDest != "" {
			state.fpdf.WriteLinkString(pdf.LineHeight, fallbackText, state.style.linkDest)
		} else {
			state.fpdf.Write(pdf.LineHeight, fallbackText)
		}
	}

	applyFont(state) // restore body font
}

func writeCode(state *renderState, text string) {
	if text == "" {
		return
	}
	// Substitute unsupported glyphs (including SMP emoji) to prevent fpdf panics
	text = pdf.SubstituteUnsupportedGlyphs(state.doc.BodyFontBytes(), state.doc.SymbolsFontBytes(), state.doc.EmojiFontBytes(), text)

	// Calculate available width for code
	left, _, right, _ := state.fpdf.GetMargins()
	pageW, _ := state.fpdf.GetPageSize()
	maxWidth := pageW - left - right

	// Split code at safe break points if it doesn't fit
	segments := splitCodeAtBreakPoints(state.fpdf, text, maxWidth, pdf.CodeContinuationIndicator)

	// Render each segment with background color
	for i, segment := range segments {
		state.fpdf.SetFillColor(pdf.ColorCodeFill.R, pdf.ColorCodeFill.G, pdf.ColorCodeFill.B)
		
		// Calculate segment width with padding
		segmentWidth := state.fpdf.GetStringWidth(segment) + 2
		
		// Check if segment fits on current line
		currentX := state.fpdf.GetX()
		left, _, right, _ := state.fpdf.GetMargins()
		pageW, _ := state.fpdf.GetPageSize()
		availableWidth := pageW - right - currentX
		
		// If segment doesn't fit and we're not at the start of the line, move to next line
		if segmentWidth > availableWidth && currentX > left {
			state.fpdf.Ln(pdf.LineHeight)
		}
		
		// Render segment
		state.fpdf.CellFormat(segmentWidth, pdf.LineHeight, segment, "", 0, "", true, 0, "")
		state.fpdf.SetFillColor(255, 255, 255)
		
		// Add line break after each segment except the last
		if i < len(segments)-1 {
			state.fpdf.Ln(pdf.LineHeight)
		}
	}
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
