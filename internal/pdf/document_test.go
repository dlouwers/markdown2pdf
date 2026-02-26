package pdf

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewDocumentCreatesPDF(t *testing.T) {
	doc := NewDocument()
	if doc == nil {
		t.Fatal("expected document")
	}
	if doc.PDF() == nil {
		t.Fatal("expected underlying pdf")
	}
}

func TestSaveWritesPDFHeader(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.pdf")
	doc := NewDocument()
	if err := doc.Save(path); err != nil {
		t.Fatalf("save: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !strings.HasPrefix(string(data), "%PDF") {
		t.Fatalf("expected PDF header")
	}
}

func TestFooterPageNumbers(t *testing.T) {
	doc := NewDocument()
	pdf := doc.PDF()
	pdf.AddPage()
	pdf.AddPage()

	dir := t.TempDir()
	path := filepath.Join(dir, "paged.pdf")
	if err := doc.Save(path); err != nil {
		t.Fatalf("save: %v", err)
	}

	if doc.FooterCalls() < 1 {
		t.Fatalf("expected footer to be called")
	}
}
