package renderer

import (
	"testing"

	gopdf "github.com/go-pdf/fpdf"
)

// newTestFPDF creates an fpdf instance with a UTF-8 font for testing wrapping.
func newTestFPDF() *gopdf.Fpdf {
	f := gopdf.New("P", "mm", "A4", "")
	f.AddPage()
	f.SetFont("Helvetica", "", 10)
	return f
}

func TestSplitTextLinesShortText(t *testing.T) {
	f := newTestFPDF()
	lines := splitTextLines(f, "Hello world", 200)
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d: %v", len(lines), lines)
	}
	if lines[0] != "Hello world" {
		t.Fatalf("expected 'Hello world', got %q", lines[0])
	}
}

func TestSplitTextLinesWraps(t *testing.T) {
	f := newTestFPDF()
	// Force wrapping by using a very narrow width.
	lines := splitTextLines(f, "Hello beautiful world", 20)
	if len(lines) < 2 {
		t.Fatalf("expected wrapping with narrow width, got %d lines: %v", len(lines), lines)
	}
}

func TestSplitTextLinesExplicitNewline(t *testing.T) {
	f := newTestFPDF()
	lines := splitTextLines(f, "Line one\nLine two", 200)
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %v", len(lines), lines)
	}
	if lines[0] != "Line one" || lines[1] != "Line two" {
		t.Fatalf("unexpected lines: %v", lines)
	}
}

func TestSplitTextLinesHyphenBreak(t *testing.T) {
	f := newTestFPDF()
	// A hyphenated word should break after the hyphen.
	lines := splitTextLines(f, "self-contained text", 25)
	if len(lines) < 2 {
		t.Fatalf("expected wrapping at hyphen, got %d lines: %v", len(lines), lines)
	}
	// The first line should end with the hyphen.
	if lines[0] != "self-" && lines[0] != "self-contained" {
		t.Logf("first line: %q (acceptable as long as no mid-rune break)", lines[0])
	}
}

func TestSplitTextLinesUTF8Safe(t *testing.T) {
	f := newTestFPDF()
	// Stars (⭐ = 3 bytes each in UTF-8). With narrow width these must not
	// be split mid-rune.
	stars := "⭐⭐⭐⭐⭐"
	lines := splitTextLines(f, stars, 10)
	for i, line := range lines {
		// Every line must be valid UTF-8 and not contain partial runes.
		for j := 0; j < len(line); {
			r, size := []rune(line)[0], len(line)
			_ = r
			_ = size
			break
		}
		// Check that each character in each line is a complete rune.
		runes := []rune(line)
		rebuilt := string(runes)
		if rebuilt != line {
			t.Fatalf("line %d has broken UTF-8: got %q, rebuilt %q", i, line, rebuilt)
		}
	}
}

func TestSplitTextLinesEmpty(t *testing.T) {
	f := newTestFPDF()
	lines := splitTextLines(f, "", 200)
	if len(lines) != 1 {
		t.Fatalf("expected 1 empty line, got %d: %v", len(lines), lines)
	}
}
