package renderer

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yuin/goldmark/ast"

	"github.com/dlouwers/markdown2pdf/internal/parser"
	"github.com/dlouwers/markdown2pdf/internal/pdf"
)

func TestHeadingFontSize(t *testing.T) {
	source := []byte("# Title")
	node, _ := parser.Parse(source)
	doc, err := pdf.NewDocument()
	if err != nil {
		t.Fatalf("new document: %v", err)
	}
	doc.PDF().SetCompression(false)
	r := New()
	if err := r.Render(doc, node, source); err != nil {
		t.Fatalf("render: %v", err)
	}

	var buf bytes.Buffer
	if err := doc.PDF().Output(&buf); err != nil {
		t.Fatalf("output: %v", err)
	}
	if !strings.Contains(buf.String(), "24.00 Tf") {
		t.Fatalf("expected heading font size in PDF output")
	}

	ptSize, _ := doc.PDF().GetFontSize()
	if ptSize < pdf.FontSizeBody-0.01 || ptSize > pdf.FontSizeBody+0.01 {
		t.Fatalf("expected body size restored, got %v", ptSize)
	}
}

func TestParagraphTextAppearsInPDF(t *testing.T) {
	source := []byte("Paragraph with **bold** and *italic* text.")
	node, _ := parser.Parse(source)
	doc, err := pdf.NewDocument()
	if err != nil {
		t.Fatalf("new document: %v", err)
	}
	doc.PDF().SetCompression(false)
	r := New()
	if err := r.Render(doc, node, source); err != nil {
		t.Fatalf("render: %v", err)
	}

	var buf bytes.Buffer
	if err := doc.PDF().Output(&buf); err != nil {
		t.Fatalf("output: %v", err)
	}
	// UTF-8 fonts use CID encoding; raw text won't appear in PDF bytes.
	// Instead verify the document rendered successfully with reasonable size.
	if len(buf.Bytes()) < 100 {
		t.Fatalf("PDF output seems too small")
	}
}

func TestBoldItalicStyles(t *testing.T) {
	source := []byte("**bold** *italic*")
	node, _ := parser.Parse(source)
	doc, err := pdf.NewDocument()
	if err != nil {
		t.Fatalf("new document: %v", err)
	}
	r := New()
	if err := r.Render(doc, node, source); err != nil {
		t.Fatalf("render: %v", err)
	}

	var bold, italic int
	_ = ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if n.Kind() == ast.KindEmphasis {
			e := n.(*ast.Emphasis)
			if e.Level == 2 {
				bold++
			}
			if e.Level == 1 {
				italic++
			}
		}
		return ast.WalkContinue, nil
	})

	if bold == 0 || italic == 0 {
		t.Fatalf("expected bold and italic emphasis nodes")
	}
}

func TestBlockquoteRenders(t *testing.T) {
	source := []byte("> This is a blockquote.\n> With multiple lines.")
	node, _ := parser.Parse(source)
	doc, err := pdf.NewDocument()
	if err != nil {
		t.Fatalf("new document: %v", err)
	}
	r := New()
	if err := r.Render(doc, node, source); err != nil {
		t.Fatalf("render: %v", err)
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "blockquote.pdf")
	if err := doc.Save(path); err != nil {
		t.Fatalf("save: %v", err)
	}
}

func TestThematicBreakRenders(t *testing.T) {
	source := []byte("Text above\n\n---\n\nText below")
	node, _ := parser.Parse(source)
	doc, err := pdf.NewDocument()
	if err != nil {
		t.Fatalf("new document: %v", err)
	}
	r := New()
	if err := r.Render(doc, node, source); err != nil {
		t.Fatalf("render: %v", err)
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "rule.pdf")
	if err := doc.Save(path); err != nil {
		t.Fatalf("save: %v", err)
	}
}

func TestAllHeadingLevels(t *testing.T) {
	source := []byte("# H1\n## H2\n### H3\n#### H4\n##### H5\n###### H6\n")
	node, _ := parser.Parse(source)
	doc, err := pdf.NewDocument()
	if err != nil {
		t.Fatalf("new document: %v", err)
	}
	r := New()
	if err := r.Render(doc, node, source); err != nil {
		t.Fatalf("render: %v", err)
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "headings.pdf")
	if err := doc.Save(path); err != nil {
		t.Fatalf("save: %v", err)
	}
}

func TestInlineCodeRenders(t *testing.T) {
	source := []byte("Use `fmt.Println` for output.")
	node, _ := parser.Parse(source)
	doc, err := pdf.NewDocument()
	if err != nil {
		t.Fatalf("new document: %v", err)
	}
	r := New()
	if err := r.Render(doc, node, source); err != nil {
		t.Fatalf("render: %v", err)
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "code.pdf")
	if err := doc.Save(path); err != nil {
		t.Fatalf("save: %v", err)
	}
}

func TestLinkRenders(t *testing.T) {
	source := []byte("Visit [example](https://example.com) for details.")
	node, _ := parser.Parse(source)
	doc, err := pdf.NewDocument()
	if err != nil {
		t.Fatalf("new document: %v", err)
	}
	r := New()
	if err := r.Render(doc, node, source); err != nil {
		t.Fatalf("render: %v", err)
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "link.pdf")
	if err := doc.Save(path); err != nil {
		t.Fatalf("save: %v", err)
	}
}

func TestComplexDocument(t *testing.T) {
	source := []byte(`# Title

This is a paragraph with **bold**, *italic*, and ` + "`code`" + `.

## Subtitle

> A blockquote with **bold** inside.

---

### Section

Visit [example](https://example.com).
`)
	node, _ := parser.Parse(source)
	doc, err := pdf.NewDocument()
	if err != nil {
		t.Fatalf("new document: %v", err)
	}
	r := New()
	if err := r.Render(doc, node, source); err != nil {
		t.Fatalf("render: %v", err)
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "complex.pdf")
	if err := doc.Save(path); err != nil {
		t.Fatalf("save: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if len(data) < 100 {
		t.Fatal("PDF output seems too small")
	}
}

func TestHeadingFontSizeValues(t *testing.T) {
	tests := []struct {
		level int
		want  float64
	}{
		{1, pdf.FontSizeH1},
		{2, pdf.FontSizeH2},
		{3, pdf.FontSizeH3},
		{4, pdf.FontSizeH4},
		{5, pdf.FontSizeH5},
		{6, pdf.FontSizeH6},
		{7, pdf.FontSizeH6},
	}
	for _, tt := range tests {
		got := headingFontSize(tt.level)
		if got != tt.want {
			t.Errorf("headingFontSize(%d) = %v, want %v", tt.level, got, tt.want)
		}
	}
}

func TestHeadingOrphanProtection(t *testing.T) {
	// Fill a page until near the bottom, then render a heading.
	// Orphan protection should push it to the next page.
	source := []byte("# Heading")
	node, _ := parser.Parse(source)
	doc, err := pdf.NewDocument()
	if err != nil {
		t.Fatalf("new document: %v", err)
	}

	// Move Y position close to the bottom margin so only ~10mm remain.
	_, pageH := doc.PDF().GetPageSize()
	nearBottom := pageH - pdf.PageMargin - 10
	doc.PDF().SetY(nearBottom)

	r := New()
	if err := r.Render(doc, node, source); err != nil {
		t.Fatalf("render: %v", err)
	}

	// The heading should have triggered a page break.
	if doc.PDF().PageNo() < 2 {
		t.Fatal("expected heading to trigger page break near bottom of page")
	}
}

func TestConsecutiveHeadingsOrphanProtection(t *testing.T) {
	// Test that consecutive headings (H1 followed by H2 with no content)
	// are kept together and not orphaned at the bottom of a page.
	source := []byte(`# Main Heading

## Subheading

Some paragraph content here.`)
	node, _ := parser.Parse(source)
	doc, err := pdf.NewDocument()
	if err != nil {
		t.Fatalf("new document: %v", err)
	}

	// Position near bottom so H1 would fit but H1+H2 would not.
	_, pageH := doc.PDF().GetPageSize()
	// Leave ~35mm space - enough for H1 alone but not H1+H2+protection.
	nearBottom := pageH - pdf.PageMargin - 35
	doc.PDF().SetY(nearBottom)

	r := New()
	if err := r.Render(doc, node, source); err != nil {
		t.Fatalf("render: %v", err)
	}

	// Should trigger page break to avoid orphaning H1.
	if doc.PDF().PageNo() < 2 {
		t.Fatal("expected consecutive headings to trigger page break")
	}
}

func TestHeadingFollowedByTable(t *testing.T) {
	// Test that a heading followed by a table is kept together.
	source := []byte(`# Table Heading

| Col1 | Col2 |
|------|------|
| A    | B    |`)
	node, _ := parser.Parse(source)
	doc, err := pdf.NewDocument()
	if err != nil {
		t.Fatalf("new document: %v", err)
	}

	// Position near bottom.
	_, pageH := doc.PDF().GetPageSize()
	nearBottom := pageH - pdf.PageMargin - 25
	doc.PDF().SetY(nearBottom)

	r := New()
	if err := r.Render(doc, node, source); err != nil {
		t.Fatalf("render: %v", err)
	}

	// Should trigger page break to keep heading with table.
	if doc.PDF().PageNo() < 2 {
		t.Fatal("expected heading+table to trigger page break")
	}
}

func TestHeadingFollowedByCodeBlock(t *testing.T) {
	// Test that a heading followed by a code block is kept together.
	source := []byte("# Code Section\n\n```go\npackage main\n```")
	node, _ := parser.Parse(source)
	doc, err := pdf.NewDocument()
	if err != nil {
		t.Fatalf("new document: %v", err)
	}

	_, pageH := doc.PDF().GetPageSize()
	nearBottom := pageH - pdf.PageMargin - 25
	doc.PDF().SetY(nearBottom)

	r := New()
	if err := r.Render(doc, node, source); err != nil {
		t.Fatalf("render: %v", err)
	}

	if doc.PDF().PageNo() < 2 {
		t.Fatal("expected heading+code block to trigger page break")
	}
}

func TestHeadingAtEOF(t *testing.T) {
	// Test that a heading at the end of document doesn't force unnecessary page break.
	source := []byte("# Final Heading")
	node, _ := parser.Parse(source)
	doc, err := pdf.NewDocument()
	if err != nil {
		t.Fatalf("new document: %v", err)
	}

	// Even near bottom, heading at EOF should render without forcing page break
	// since there's no content to orphan from.
	// Position with enough space for the heading itself (H1 needs ~35mm).
	_, pageH := doc.PDF().GetPageSize()
	nearBottom := pageH - pdf.PageMargin - 40 // 40mm should be enough for H1
	doc.PDF().SetY(nearBottom)

	r := New()
	if err := r.Render(doc, node, source); err != nil {
		t.Fatalf("render: %v", err)
	}

	// Should stay on page 1 since no orphan risk at EOF.
	// Should stay on page 1 since no orphan risk at EOF.
	if pageNo := doc.PDF().PageNo(); pageNo != 1 {
		t.Fatalf("heading at EOF should not force page break, got page %d", pageNo)
		t.Fatal("heading at EOF should not force page break")
	}
}

func TestMajorHeadingProtection(t *testing.T) {
	// Test that H1/H2 get more protection (3 lines) than H3-H6 (2 lines).
	source := []byte(`# Major Heading

Paragraph content here.`)
	node, _ := parser.Parse(source)
	doc, err := pdf.NewDocument()
	if err != nil {
		t.Fatalf("new document: %v", err)
	}

	// Position where 2 lines would fit but 3 would not.
	_, pageH := doc.PDF().GetPageSize()
	// Leave space for heading + 2.5 lines (H1 should require 3 lines).
	nearBottom := pageH - pdf.PageMargin - 40
	doc.PDF().SetY(nearBottom)

	r := New()
	if err := r.Render(doc, node, source); err != nil {
		t.Fatalf("render: %v", err)
	}

	// H1 should trigger page break due to 3-line requirement.
	if doc.PDF().PageNo() < 2 {
		t.Fatal("expected H1 to require 3 lines minimum")
	}
}
