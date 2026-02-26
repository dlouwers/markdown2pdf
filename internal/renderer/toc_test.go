package renderer

import (
	"path/filepath"
	"testing"

	"github.com/dlouwers/markdown2pdf/internal/parser"
	"github.com/dlouwers/markdown2pdf/internal/pdf"
)

func TestTOCGeneratesLinks(t *testing.T) {
	source := []byte("# Title\n\nIntro.\n\n## Section\n\nBody.\n\n### Sub\n\nDetails.")
	node, _ := parser.Parse(source)
	doc := pdf.NewDocument()
	r := New()
	r.TOC = true
	if err := r.Render(doc, node, source); err != nil {
		t.Fatalf("render: %v", err)
	}

	// TOC page + content page(s) — expect at least 2 pages.
	if doc.PDF().PageNo() < 2 {
		t.Fatalf("expected at least 2 pages with TOC, got %d", doc.PDF().PageNo())
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "toc.pdf")
	if err := doc.Save(path); err != nil {
		t.Fatalf("save: %v", err)
	}
}

func TestTOCSkippedWhenDisabled(t *testing.T) {
	source := []byte("# Title\n\nBody text.")
	node, _ := parser.Parse(source)
	doc := pdf.NewDocument()
	r := New()
	r.TOC = false
	if err := r.Render(doc, node, source); err != nil {
		t.Fatalf("render: %v", err)
	}

	// Without TOC, a short document should be 1 page.
	if doc.PDF().PageNo() != 1 {
		t.Fatalf("expected 1 page without TOC, got %d", doc.PDF().PageNo())
	}
}

func TestTOCEmptyDocument(t *testing.T) {
	source := []byte("No headings here, just a paragraph.")
	node, _ := parser.Parse(source)
	doc := pdf.NewDocument()
	r := New()
	r.TOC = true
	if err := r.Render(doc, node, source); err != nil {
		t.Fatalf("render: %v", err)
	}

	// No headings means no TOC page added.
	if doc.PDF().PageNo() != 1 {
		t.Fatalf("expected 1 page with no headings, got %d", doc.PDF().PageNo())
	}
}
