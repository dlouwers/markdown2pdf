package main

import (
	"fmt"
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

func TestRunWithTOC(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "toc.md")
	content := `# Main Title

Some intro text.

## Section One

First section content.

### Subsection 1.1

Details.

## Section Two

Second section content.
`
	if err := os.WriteFile(input, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	output := filepath.Join(dir, "toc.pdf")
	code := run([]string{"-o", output, "--toc", input})
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
	// With TOC, expect at least 2 pages (TOC + content).
	if len(data) < 500 {
		t.Fatalf("PDF seems too small: %d bytes", len(data))
	}
}

func TestEndToEndWithCodeBlocks(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "code.md")
	content := `# Code Examples

## Go Function

` + "```go" + `
package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}
` + "```" + `

## JSON Config

` + "```json" + `
{
  "name": "markdown2pdf",
  "version": "0.0.1"
}
` + "```" + `

## Python Script

` + "```python" + `
def greet(name: str) -> str:
    return f"Hello, {name}!"
` + "```" + `
`
	if err := os.WriteFile(input, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	output := filepath.Join(dir, "code.pdf")
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
	if len(data) < 1000 {
		t.Fatalf("PDF with code blocks too small: %d bytes", len(data))
	}
}

func TestEndToEndWithListsAndTables(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "mixed.md")
	content := `# Lists and Tables

## Unordered List

- Alpha
  - Nested one
  - Nested two
- Bravo
- Charlie

## Ordered List

1. First
2. Second
3. Third

## Task List

- [x] Done task
- [ ] Pending task
- [x] Another done

## Data Table

| Name    | Role      | Status  |
|---------|-----------|---------|
| Alice   | Engineer  | Active  |
| Bob     | Designer  | Active  |
| Charlie | Manager   | Away    |
`
	if err := os.WriteFile(input, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	output := filepath.Join(dir, "mixed.pdf")
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
}

func TestEndToEndLongTablePageBreak(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "longtable.md")

	var sb strings.Builder
	sb.WriteString("# Long Table Test\n\n")
	sb.WriteString("| Row | Name | Value |\n|-----|------|-------|\n")
	for i := range 60 {
		fmt.Fprintf(&sb, "| %d | Item-%d | Value-%d |\n", i+1, i+1, i+1)
	}
	sb.WriteString("\n## After the Table\n\nContent after the long table.\n")

	if err := os.WriteFile(input, []byte(sb.String()), 0644); err != nil {
		t.Fatal(err)
	}

	output := filepath.Join(dir, "longtable.pdf")
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
	// A 60-row table should produce a multi-page PDF.
	if len(data) < 2000 {
		t.Fatalf("multi-page table PDF too small: %d bytes", len(data))
	}
}

func TestEndToEndTOCWithAllFeatures(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "full_toc.md")
	content := `# Document Title

Introductory paragraph.

## Code Section

` + "```go" + `
fmt.Println("hello")
` + "```" + `

## List Section

- Item A
- Item B

## Table Section

| Col1 | Col2 |
|------|------|
| A    | B    |
| C    | D    |

## Conclusion

Final paragraph.
`
	if err := os.WriteFile(input, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Without TOC
	outputNoTOC := filepath.Join(dir, "no_toc.pdf")
	code := run([]string{"-o", outputNoTOC, input})
	if code != 0 {
		t.Fatalf("no-toc: expected exit code 0, got %d", code)
	}

	// With TOC
	outputTOC := filepath.Join(dir, "with_toc.pdf")
	code = run([]string{"-o", outputTOC, "--toc", input})
	if code != 0 {
		t.Fatalf("toc: expected exit code 0, got %d", code)
	}

	noTOCData, err := os.ReadFile(outputNoTOC)
	if err != nil {
		t.Fatalf("read no-toc: %v", err)
	}
	tocData, err := os.ReadFile(outputTOC)
	if err != nil {
		t.Fatalf("read toc: %v", err)
	}

	// TOC version should be larger (extra TOC page).
	if len(tocData) <= len(noTOCData) {
		t.Errorf("TOC PDF (%d bytes) should be larger than no-TOC PDF (%d bytes)",
			len(tocData), len(noTOCData))
	}
}
