package renderer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dlouwers/markdown2pdf/internal/diagram"
	"github.com/dlouwers/markdown2pdf/internal/parser"
	"github.com/dlouwers/markdown2pdf/internal/pdf"
)

func renderPDF(t *testing.T, source []byte, disableCompression bool) []byte {
	t.Helper()

	node, _, _ := parser.Parse(source)
	doc, err := pdf.NewDocument()
	if err != nil {
		t.Fatalf("new document: %v", err)
	}
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

func TestFencedCodeBlockWithMonoFont(t *testing.T) {
	source := []byte("```go\npackage main\n\nfunc main() {}\n```")
	data := renderPDF(t, source, true)
	if !strings.Contains(string(data), "utf8notosansmono") {
		t.Fatalf("expected NotoSansMono font in PDF output")
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

func TestD2DiagramRenders(t *testing.T) {
	if !diagram.ChromiumAvailable() {
		t.Skip("headless Chromium not available")
	}

	source := []byte("```d2\nx -> y\n```")
	data := renderPDF(t, source, false)
	if len(data) < 100 {
		t.Fatalf("PDF output seems too small")
	}
}

func TestD2DiagramWithInvalidSource(t *testing.T) {
	// Invalid D2 source should render a placeholder, not crash.
	source := []byte("```d2\n{{{{invalid}}}}\n```")
	data := renderPDF(t, source, false)
	if len(data) < 100 {
		t.Fatalf("PDF output seems too small")
	}
}

func TestMermaidDiagramRendersOrPlaceholder(t *testing.T) {
	// If mmdc is available, renders diagram; otherwise renders placeholder.
	// Either way, it should produce valid PDF output.
	source := []byte("```mermaid\ngraph TD\n    A --> B\n```")
	data := renderPDF(t, source, false)
	if len(data) < 100 {
		t.Fatalf("PDF output seems too small")
	}
}

func TestRegularCodeBlockStillWorks(t *testing.T) {
	if !diagram.ChromiumAvailable() {
		t.Skip("headless Chromium not available")
	}

	// Ensure that non-diagram fenced code blocks still render normally.
	source := []byte("```go\npackage main\n```\n\n```d2\na -> b\n```")
	data := renderPDF(t, source, true)
	if !strings.Contains(string(data), "utf8notosansmono") {
		t.Fatalf("expected NotoSansMono font for Go code block")
	}
}
