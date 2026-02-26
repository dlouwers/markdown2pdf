// Package renderer walks a goldmark AST and emits PDF elements.
package renderer

import (
	"github.com/yuin/goldmark/ast"

	"github.com/dlouwers/markdown2pdf/internal/pdf"
)

type Renderer struct{}

func New() *Renderer {
	return &Renderer{}
}

func (r *Renderer) Render(doc *pdf.Document, node ast.Node, source []byte) error {
	state := newRenderState(doc)
	return renderNode(state, node, source)
}
