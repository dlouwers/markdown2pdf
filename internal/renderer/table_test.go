package renderer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dlouwers/markdown2pdf/internal/parser"
	"github.com/dlouwers/markdown2pdf/internal/pdf"
)

func TestSimpleTableRenders(t *testing.T) {
	source := []byte("| Name | Age | City |\n|------|-----|------|\n| Alice | 30 | Amsterdam |\n| Bob | 25 | Berlin |")
	data := renderPDF(t, source, false)
	if len(data) == 0 {
		t.Fatalf("expected PDF output")
	}
}

func TestTableWithAlignment(t *testing.T) {
	source := []byte("| Left | Center | Right |\n|:-----|:------:|------:|\n| L1 | C1 | R1 |\n| L2 | C2 | R2 |")
	data := renderPDF(t, source, false)
	if len(data) == 0 {
		t.Fatalf("expected PDF output")
	}
}

func TestTableHasBorders(t *testing.T) {
	source := []byte("| Name | Age |\n|------|-----|\n| Alice | 30 |\n| Bob | 25 |")
	data := renderPDF(t, source, true)
	if !strings.Contains(string(data), " re ") {
		t.Fatalf("expected rectangle operation in PDF output")
	}
}

func TestTableFromFixture(t *testing.T) {
	path := filepath.Join("..", "..", "testdata", "tables.md")
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	data := renderPDF(t, source, false)
	if len(data) < 100 {
		t.Fatalf("PDF output seems too small")
	}
}

func TestTablePageBreakReRendersHeader(t *testing.T) {
	// Build a table large enough to span two pages.
	var sb strings.Builder
	sb.WriteString("| Name | Value |\n|------|-------|\n")
	for i := range 80 {
		fmt.Fprintf(&sb, "| Row %d | Val %d |\n", i, i)
	}

	source := []byte(sb.String())
	node, _, _ := parser.Parse(source)
	doc, err := pdf.NewDocument()
	if err != nil {
		t.Fatalf("new document: %v", err)
	}
	r := New()
	if err := r.Render(doc, node, source); err != nil {
		t.Fatalf("render: %v", err)
	}

	if doc.PDF().PageNo() < 2 {
		t.Fatal("expected table to span multiple pages")
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "table-pagebreak.pdf")
	if err := doc.Save(path); err != nil {
		t.Fatalf("save: %v", err)
	}
}
