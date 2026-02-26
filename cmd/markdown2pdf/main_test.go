package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunNoArgs(t *testing.T) {
	code := run([]string{})
	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
}

func TestRunTooManyArgs(t *testing.T) {
	code := run([]string{"a.md", "b.md"})
	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
}

func TestRunInvalidFlag(t *testing.T) {
	code := run([]string{"--nonexistent"})
	if code != 2 {
		t.Errorf("expected exit code 2, got %d", code)
	}
}

func TestRunVersion(t *testing.T) {
	code := run([]string{"--version"})
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
}

func TestRunFileNotFound(t *testing.T) {
	code := run([]string{"nonexistent.md"})
	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
}

func TestRunInputIsDirectory(t *testing.T) {
	dir := t.TempDir()
	code := run([]string{dir})
	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
}

func TestRunValidInput(t *testing.T) {
	input := filepath.Join("..", "..", "testdata", "basic.md")
	if _, err := os.Stat(input); err != nil {
		t.Skipf("testdata not available: %v", err)
	}

	output := filepath.Join(t.TempDir(), "valid.pdf")
	code := run([]string{"-o", output, input})
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}

	// Verify PDF was created.
	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatalf("expected output PDF: %v", err)
	}
	if !strings.HasPrefix(string(data), "%PDF") {
		t.Fatal("output file does not start with %PDF header")
	}
}

func TestRunExplicitOutput(t *testing.T) {
	input := filepath.Join("..", "..", "testdata", "basic.md")
	if _, err := os.Stat(input); err != nil {
		t.Skipf("testdata not available: %v", err)
	}

	output := filepath.Join(t.TempDir(), "out.pdf")
	code := run([]string{"-o", output, input})
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}

	if _, err := os.Stat(output); err != nil {
		t.Fatalf("expected output file: %v", err)
	}
}

func TestRunDefaultOutputDerivation(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "test.md")
	if err := os.WriteFile(input, []byte("# Test"), 0644); err != nil {
		t.Fatal(err)
	}

	code := run([]string{input})
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}

	// Verify the derived output path was used.
	expected := filepath.Join(dir, "test.pdf")
	data, err := os.ReadFile(expected)
	if err != nil {
		t.Fatalf("expected derived output at %s: %v", expected, err)
	}
	if !strings.HasPrefix(string(data), "%PDF") {
		t.Fatal("derived output is not a valid PDF")
	}
}

func TestEndToEndFullDocument(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "full.md")
	content := `# Main Title

This is a paragraph with **bold**, *italic*, and ` + "`inline code`" + `.

## Second Section

> A blockquote with some content.
> It spans multiple lines.

---

### Third Section

Visit [example](https://example.com) for more info.

#### Fourth Level

##### Fifth Level

###### Sixth Level
`
	if err := os.WriteFile(input, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	output := filepath.Join(dir, "full.pdf")
	code := run([]string{"-o", output, input})
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}

	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if !strings.HasPrefix(string(data), "%PDF") {
		t.Fatal("output is not a valid PDF")
	}
	if len(data) < 500 {
		t.Fatalf("PDF seems too small: %d bytes", len(data))
	}
}
