package pdf_test

import (
	"os"
	"sort"
	"testing"

	"github.com/dlouwers/markdown2pdf/internal/pdf"
	"github.com/go-pdf/fpdf"
)

func TestGlyphSupport(t *testing.T) {
	type glyph struct {
		name string
		r    rune
	}
	glyphs := []glyph{
		{"★ (black star)", '★'},
		{"☆ (white star)", '☆'},
		{"✓ (check mark)", '✓'},
		{"✗ (ballot X)", '✗'},
		{"☐ (ballot box)", '☐'},
		{"☑ (ballot checked)", '☑'},
		{"☒ (ballot X-ed)", '☒'},
		{"• (bullet)", '•'},
		{"→ (right arrow)", '→'},
		{"← (left arrow)", '←'},
		{"♠ (spade)", '♠'},
		{"♥ (heart)", '♥'},
		{"⚠ (warning)", '⚠'},
		{"⌘ (command)", '⌘'},
		{"❤ (heavy heart)", '❤'},
	}
	sort.Slice(glyphs, func(i, j int) bool { return glyphs[i].name < glyphs[j].name })

	notoData, err := os.ReadFile("fonts/NotoSans-Regular.ttf")
	if err != nil {
		t.Fatal(err)
	}

	t.Log("=== Noto Sans (default) ===")
	for _, g := range glyphs {
		ok := pdf.FontSupportsGlyph(notoData, g.r)
		s := "NO "
		if ok {
			s = "YES"
		}
		t.Logf("  %s  %s", s, g.name)
	}

	p := fpdf.New("P", "mm", "A4", "")
	nfData, err := pdf.LoadCustomFonts(p, "../../testdata/JetBrainsMonoNerdFont.tar.gz")
	if err != nil {
		t.Fatal(err)
	}
	if nfData == nil {
		t.Fatal("No Regular font returned from NerdFont archive")
	}

	t.Log("\n=== JetBrains Mono NerdFont ===")
	for _, g := range glyphs {
		ok := pdf.FontSupportsGlyph(nfData, g.r)
		s := "NO "
		if ok {
			s = "YES"
		}
		t.Logf("  %s  %s", s, g.name)
	}
}
