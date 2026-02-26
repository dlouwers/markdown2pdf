// Package pdf manages PDF document creation, page layout, and styling.
package pdf

import (
	"fmt"

	"github.com/go-pdf/fpdf"
)

// DocumentOption configures Document creation.
type DocumentOption func(*documentConfig)

type documentConfig struct {
	customFontArchive    string
	customSymbolsArchive string
	customEmojiArchive   string
}

// WithCustomFont returns an option that loads fonts from a zip or tar.gz
// archive instead of the default embedded Noto Sans.
func WithCustomFont(archivePath string) DocumentOption {
	return func(c *documentConfig) {
		c.customFontArchive = archivePath
	}
}

// WithCustomSymbolsFont returns an option that loads a symbols fallback font
// from a zip or tar.gz archive instead of the default embedded Noto Sans Symbols 2.
func WithCustomSymbolsFont(archivePath string) DocumentOption {
	return func(c *documentConfig) {
		c.customSymbolsArchive = archivePath
	}
}

// WithCustomEmojiFont returns an option that loads an emoji fallback font
// from a zip or tar.gz archive instead of the default embedded Noto Emoji.
func WithCustomEmojiFont(archivePath string) DocumentOption {
	return func(c *documentConfig) {
		c.customEmojiArchive = archivePath
	}
}

type Document struct {
	pdf              *fpdf.Fpdf
	footerCalls      int
	baseDir          string
	bodyFontBytes    []byte // regular body font TTF bytes for glyph detection
	symbolsFontBytes []byte // symbols fallback font TTF bytes for glyph detection
	emojiFontBytes   []byte // emoji fallback font TTF bytes for glyph detection
}

func NewDocument(opts ...DocumentOption) (*Document, error) {
	cfg := &documentConfig{}
	for _, o := range opts {
		o(cfg)
	}

	pdfDoc := fpdf.New("P", "mm", "A4", "")
	doc := &Document{pdf: pdfDoc}

	if cfg.customFontArchive != "" {
		bodyBytes, err := LoadCustomFonts(pdfDoc, cfg.customFontArchive)
		if err != nil {
			return nil, fmt.Errorf("load custom fonts: %w", err)
		}
		doc.bodyFontBytes = bodyBytes
		// Still register the mono font for code blocks.
		pdfDoc.AddUTF8FontFromBytes(FontCode, "", notoSansMonoRegular)
		pdfDoc.AddUTF8FontFromBytes(FontCode, "B", notoSansMonoBold)
	} else {
		doc.bodyFontBytes = RegisterFonts(pdfDoc)
	}

	// Register symbols fallback font.
	if cfg.customSymbolsArchive != "" {
		symBytes, err := LoadCustomSymbolsFont(pdfDoc, cfg.customSymbolsArchive)
		if err != nil {
			return nil, fmt.Errorf("load custom symbols font: %w", err)
		}
		doc.symbolsFontBytes = symBytes
	} else {
		doc.symbolsFontBytes = RegisterSymbolsFont(pdfDoc)
	}

	// Register emoji fallback font.
	if cfg.customEmojiArchive != "" {
		emojiBytes, err := LoadCustomEmojiFont(pdfDoc, cfg.customEmojiArchive)
		if err != nil {
			return nil, fmt.Errorf("load custom emoji font: %w", err)
		}
		doc.emojiFontBytes = emojiBytes
	} else {
		doc.emojiFontBytes = RegisterEmojiFont(pdfDoc)
	}

	pdfDoc.SetMargins(PageMargin, PageMargin, PageMargin)
	pdfDoc.SetAutoPageBreak(true, PageMargin)
	pdfDoc.SetFont(FontBody, "", FontSizeBody)
	pdfDoc.SetTextColor(ColorText.R, ColorText.G, ColorText.B)
	pdfDoc.SetFooterFunc(func() {
		doc.footerCalls++
		pdfDoc.SetY(-PageMargin + 5)
		pdfDoc.SetFont(FontBody, "", 9)
		pdfDoc.SetTextColor(120, 120, 120)
		pdfDoc.CellFormat(0, 10, fmt.Sprintf("%d", pdfDoc.PageNo()), "", 0, "C", false, 0, "")
		pdfDoc.SetFont(FontBody, "", FontSizeBody)
		pdfDoc.SetTextColor(ColorText.R, ColorText.G, ColorText.B)
	})
	pdfDoc.AddPage()

	return doc, nil
}

func (d *Document) Save(path string) error {
	return d.pdf.OutputFileAndClose(path)
}

func (d *Document) PDF() *fpdf.Fpdf {
	return d.pdf
}

func (d *Document) FooterCalls() int {
	return d.footerCalls
}

func (d *Document) BaseDir() string {
	return d.baseDir
}

func (d *Document) SetBaseDir(dir string) {
	d.baseDir = dir
}

// BodyFontBytes returns the raw TTF bytes of the regular body font,
// used for glyph detection (e.g. checking if bullet characters exist).
func (d *Document) BodyFontBytes() []byte {
	return d.bodyFontBytes
}

// SymbolsFontBytes returns the raw TTF bytes of the symbols fallback font,
// used for glyph detection when the body font lacks a glyph.
func (d *Document) SymbolsFontBytes() []byte {
	return d.symbolsFontBytes
}

// EmojiFontBytes returns the raw TTF bytes of the emoji fallback font,
// used for glyph detection when neither the body nor symbols font has a glyph.
func (d *Document) EmojiFontBytes() []byte {
	return d.emojiFontBytes
}
