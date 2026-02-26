package renderer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dlouwers/markdown2pdf/internal/parser"
	"github.com/dlouwers/markdown2pdf/internal/pdf"
)

func renderPDF(t *testing.T, source []byte, disableCompression bool) []byte {
	t.Helper()

	node, _ := parser.Parse(source)
	doc := pdf.NewDocument()
	if disableCompression {
		doc.PDF().SetCompression(false)
	}

	r := New()
	if err := r.Render(doc, node, source); err != nil {
		t.Fatalf("render: %v", err)
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "code-blocks.pdf")
	if err := doc.Save(path); err != nil {
		t.Fatalf("save: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}

	return data
}

func TestFencedCodeBlockRenders(t *testing.T) {
	source := []byte("```go\npackage main\n\nfunc main() {}\n```")
	data := renderPDF(t, source, true)
	if len(data) < 100 {
		t.Fatalf("PDF output seems too small")
	}
}

func TestFencedCodeBlockWithLanguage(t *testing.T) {
	source := []byte("```go\npackage main\n\nfunc main() {}\n```")
	data := renderPDF(t, source, true)
	if !strings.Contains(string(data), "Courier") {
		t.Fatalf("expected Courier font in PDF output")
	}
}

func TestFencedCodeBlockNoLanguage(t *testing.T) {
	source := []byte("```\nno language here\n```")
	data := renderPDF(t, source, false)
	if len(data) == 0 {
		t.Fatalf("expected PDF output")
	}
}

func TestIndentedCodeBlockRenders(t *testing.T) {
	source := []byte("    indented code\n    more code")
	data := renderPDF(t, source, false)
	if len(data) == 0 {
		t.Fatalf("expected PDF output")
	}
}

func TestCodeBlockHasBackground(t *testing.T) {
	source := []byte("```go\npackage main\n```")
	data := renderPDF(t, source, true)
	if !strings.Contains(string(data), " re ") {
		t.Fatalf("expected rectangle operation in PDF output")
	}
}

func TestMultipleCodeBlocks(t *testing.T) {
	source := []byte("```go\npackage main\n```\n\n```python\nprint('hi')\n```")
	data := renderPDF(t, source, false)
	if len(data) == 0 {
		t.Fatalf("expected PDF output")
	}
}

func TestCodeBlocksFromFixture(t *testing.T) {
	path := filepath.Join("..", "..", "testdata", "code_blocks.md")
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	data := renderPDF(t, source, false)
	if len(data) < 100 {
		t.Fatalf("PDF output seems too small")
	}
}
