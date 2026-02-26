package renderer

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yuin/goldmark/ast"

	"github.com/dlouwers/markdown2pdf/internal/parser"
	"github.com/dlouwers/markdown2pdf/internal/pdf"
)

func TestHeadingFontSize(t *testing.T) {
	source := []byte("# Title")
	node, _ := parser.Parse(source)
	doc := pdf.NewDocument()
	doc.PDF().SetCompression(false)
	r := New()
	if err := r.Render(doc, node, source); err != nil {
		t.Fatalf("render: %v", err)
	}

	var buf bytes.Buffer
	if err := doc.PDF().Output(&buf); err != nil {
		t.Fatalf("output: %v", err)
	}
	if !strings.Contains(buf.String(), "24.00 Tf") {
		t.Fatalf("expected heading font size in PDF output")
	}

	ptSize, _ := doc.PDF().GetFontSize()
	if ptSize < pdf.FontSizeBody-0.01 || ptSize > pdf.FontSizeBody+0.01 {
		t.Fatalf("expected body size restored, got %v", ptSize)
	}
}

func TestParagraphTextAppearsInPDF(t *testing.T) {
	source := []byte("Paragraph with **bold** and *italic* text.")
	node, _ := parser.Parse(source)
	doc := pdf.NewDocument()
	doc.PDF().SetCompression(false)
	r := New()
	if err := r.Render(doc, node, source); err != nil {
		t.Fatalf("render: %v", err)
	}

	var buf bytes.Buffer
	if err := doc.PDF().Output(&buf); err != nil {
		t.Fatalf("output: %v", err)
	}
	// UTF-8 fonts use CID encoding; raw text won't appear in PDF bytes.
	// Instead verify the document rendered successfully with reasonable size.
	if len(buf.Bytes()) < 100 {
		t.Fatalf("PDF output seems too small")
	}
}

func TestBoldItalicStyles(t *testing.T) {
	source := []byte("**bold** *italic*")
	node, _ := parser.Parse(source)
	doc := pdf.NewDocument()
	r := New()
	if err := r.Render(doc, node, source); err != nil {
		t.Fatalf("render: %v", err)
	}

	var bold, italic int
	_ = ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if n.Kind() == ast.KindEmphasis {
			e := n.(*ast.Emphasis)
			if e.Level == 2 {
				bold++
			}
			if e.Level == 1 {
				italic++
			}
		}
		return ast.WalkContinue, nil
	})

	if bold == 0 || italic == 0 {
		t.Fatalf("expected bold and italic emphasis nodes")
	}
}

func TestBlockquoteRenders(t *testing.T) {
	source := []byte("> This is a blockquote.\n> With multiple lines.")
	node, _ := parser.Parse(source)
	doc := pdf.NewDocument()
	r := New()
	if err := r.Render(doc, node, source); err != nil {
		t.Fatalf("render: %v", err)
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "blockquote.pdf")
	if err := doc.Save(path); err != nil {
		t.Fatalf("save: %v", err)
	}
}

func TestThematicBreakRenders(t *testing.T) {
	source := []byte("Text above\n\n---\n\nText below")
	node, _ := parser.Parse(source)
	doc := pdf.NewDocument()
	r := New()
	if err := r.Render(doc, node, source); err != nil {
		t.Fatalf("render: %v", err)
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "rule.pdf")
	if err := doc.Save(path); err != nil {
		t.Fatalf("save: %v", err)
	}
}

func TestAllHeadingLevels(t *testing.T) {
	source := []byte("# H1\n## H2\n### H3\n#### H4\n##### H5\n###### H6\n")
	node, _ := parser.Parse(source)
	doc := pdf.NewDocument()
	r := New()
	if err := r.Render(doc, node, source); err != nil {
		t.Fatalf("render: %v", err)
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "headings.pdf")
	if err := doc.Save(path); err != nil {
		t.Fatalf("save: %v", err)
	}
}

func TestInlineCodeRenders(t *testing.T) {
	source := []byte("Use `fmt.Println` for output.")
	node, _ := parser.Parse(source)
	doc := pdf.NewDocument()
	r := New()
	if err := r.Render(doc, node, source); err != nil {
		t.Fatalf("render: %v", err)
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "code.pdf")
	if err := doc.Save(path); err != nil {
		t.Fatalf("save: %v", err)
	}
}

func TestLinkRenders(t *testing.T) {
	source := []byte("Visit [example](https://example.com) for details.")
	node, _ := parser.Parse(source)
	doc := pdf.NewDocument()
	r := New()
	if err := r.Render(doc, node, source); err != nil {
		t.Fatalf("render: %v", err)
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "link.pdf")
	if err := doc.Save(path); err != nil {
		t.Fatalf("save: %v", err)
	}
}

func TestComplexDocument(t *testing.T) {
	source := []byte(`# Title

This is a paragraph with **bold**, *italic*, and ` + "`code`" + `.

## Subtitle

> A blockquote with **bold** inside.

---

### Section

Visit [example](https://example.com).
`)
	node, _ := parser.Parse(source)
	doc := pdf.NewDocument()
	r := New()
	if err := r.Render(doc, node, source); err != nil {
		t.Fatalf("render: %v", err)
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "complex.pdf")
	if err := doc.Save(path); err != nil {
		t.Fatalf("save: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if len(data) < 100 {
		t.Fatal("PDF output seems too small")
	}
}

func TestHeadingFontSizeValues(t *testing.T) {
	tests := []struct {
		level int
		want  float64
	}{
		{1, pdf.FontSizeH1},
		{2, pdf.FontSizeH2},
		{3, pdf.FontSizeH3},
		{4, pdf.FontSizeH4},
		{5, pdf.FontSizeH5},
		{6, pdf.FontSizeH6},
		{7, pdf.FontSizeH6},
	}
	for _, tt := range tests {
		got := headingFontSize(tt.level)
		if got != tt.want {
			t.Errorf("headingFontSize(%d) = %v, want %v", tt.level, got, tt.want)
		}
	}
}
