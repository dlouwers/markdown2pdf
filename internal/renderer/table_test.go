package renderer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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
