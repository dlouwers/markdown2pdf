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

func TestMultilineUnorderedListItemsRender(t *testing.T) {
	source := []byte(`- Short item
- This is a much longer item that should wrap to multiple lines when rendered in the PDF. The continuation lines should align with the start of the text on the first line.
- Another short`)
	data := renderPDF(t, source, false)
	if len(data) < 100 {
		t.Fatalf("PDF output seems too small")
	}
}

func TestMultilineOrderedListItemsRender(t *testing.T) {
	source := []byte(`1. Short
2. Second item is intentionally very long so that it wraps across multiple lines in the PDF output. The continuation lines should be indented to align with the text start position.
3. Third`)
	data := renderPDF(t, source, false)
	if len(data) < 100 {
		t.Fatalf("PDF output seems too small")
	}
}

func TestMultilineNestedListItemsRender(t *testing.T) {
	source := []byte(`- Parent item with some longer text that might wrap
  - Child item that is also quite long and should wrap to multiple lines while maintaining proper indentation relative to its own bullet point
  - Short child`)
	data := renderPDF(t, source, false)
	if len(data) < 100 {
		t.Fatalf("PDF output seems too small")
	}
}

func TestMultilineTaskListItemsRender(t *testing.T) {
	source := []byte(`- [x] Completed task with a very long description that spans multiple lines to test whether continuation lines properly align with the first line of text
- [ ] Short incomplete
- [x] Done`)
	data := renderPDF(t, source, true)
	if len(data) < 100 {
		t.Fatalf("PDF output seems too small")
	}
}

func TestMultilineListsFromFixture(t *testing.T) {
	path := filepath.Join("..", "..", "testdata", "multiline-lists.md")
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	data := renderPDF(t, source, true)
	if len(data) < 100 {
		t.Fatalf("PDF output seems too small")
	}
}
