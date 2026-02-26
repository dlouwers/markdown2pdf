package renderer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUnorderedListRenders(t *testing.T) {
	source := []byte("- Item one\n- Item two\n- Item three")
	data := renderPDF(t, source, false)
	if len(data) == 0 {
		t.Fatalf("expected PDF output")
	}
}

func TestOrderedListRenders(t *testing.T) {
	source := []byte("1. First\n2. Second\n3. Third")
	data := renderPDF(t, source, false)
	if len(data) == 0 {
		t.Fatalf("expected PDF output")
	}
}

func TestNestedListRenders(t *testing.T) {
	source := []byte("- Parent\n  - Child\n    - Grandchild\n- Another")
	data := renderPDF(t, source, false)
	if len(data) == 0 {
		t.Fatalf("expected PDF output")
	}
}

func TestTaskListRenders(t *testing.T) {
	source := []byte("- [x] Done\n- [ ] Todo\n- [x] Again")
	data := renderPDF(t, source, true)
	if len(data) < 100 {
		t.Fatalf("PDF output seems too small")
	}
}

func TestListFromFixture(t *testing.T) {
	path := filepath.Join("..", "..", "testdata", "lists.md")
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	data := renderPDF(t, source, false)
	if len(data) < 100 {
		t.Fatalf("PDF output seems too small")
	}
}
