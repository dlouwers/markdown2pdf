// Package renderer walks a goldmark AST and emits PDF elements.
package renderer

import (
	"github.com/yuin/goldmark/ast"

	"github.com/dlouwers/markdown2pdf/internal/pdf"
)

// Renderer converts a goldmark AST into PDF elements on a Document.
type Renderer struct {
	// TOC enables generation of a clickable table of contents before content.
	TOC bool
}

// New creates a Renderer with default settings.
func New() *Renderer {
	return &Renderer{}
}

// Render walks the AST and writes PDF elements to doc. If TOC is enabled,
// a table of contents is rendered on the first page(s) before the content.
func (r *Renderer) Render(doc *pdf.Document, node ast.Node, source []byte) error {
	state := newRenderState(doc)

	if r.TOC {
		entries := collectTOCEntries(node, source, state)
		if len(entries) > 0 {
			renderTOC(state, entries)
		}
	}

	return renderNode(state, node, source)
}
