package pdf

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	_ "embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-pdf/fpdf"
	"golang.org/x/image/font/sfnt"
)

//go:embed fonts/NotoSans-Regular.ttf
var notoSansRegular []byte

//go:embed fonts/NotoSans-Bold.ttf
var notoSansBold []byte

//go:embed fonts/NotoSans-Italic.ttf
var notoSansItalic []byte

//go:embed fonts/NotoSans-BoldItalic.ttf
var notoSansBoldItalic []byte

//go:embed fonts/NotoSansMono-Regular.ttf
var notoSansMonoRegular []byte

//go:embed fonts/NotoSansMono-Bold.ttf
var notoSansMonoBold []byte

// RegisterFonts registers the embedded Noto Sans font families with the PDF
// document. After registration, FontBody and FontCode constants can be used
// with SetFont as usual. It returns the regular body font bytes for glyph
// detection.
func RegisterFonts(pdf *fpdf.Fpdf) []byte {
	pdf.AddUTF8FontFromBytes(FontBody, "", notoSansRegular)
	pdf.AddUTF8FontFromBytes(FontBody, "B", notoSansBold)
	pdf.AddUTF8FontFromBytes(FontBody, "I", notoSansItalic)
	pdf.AddUTF8FontFromBytes(FontBody, "BI", notoSansBoldItalic)

	pdf.AddUTF8FontFromBytes(FontCode, "", notoSansMonoRegular)
	pdf.AddUTF8FontFromBytes(FontCode, "B", notoSansMonoBold)

	return notoSansRegular
}

// FontSupportsGlyph checks whether the given TTF font data contains a glyph
// for the specified rune. It uses golang.org/x/image/font/sfnt to parse the
// font's cmap table. A GlyphIndex of 0 means the .notdef (missing) glyph.
func FontSupportsGlyph(fontData []byte, r rune) bool {
	f, err := sfnt.Parse(fontData)
	if err != nil {
		return false
	}
	var buf sfnt.Buffer
	idx, err := f.GlyphIndex(&buf, r)
	if err != nil {
		return false
	}
	return idx != 0
}

// LoadCustomFonts reads a zip or tar.gz archive from path, extracts all .ttf
// files, and registers them as the body font family on the PDF document. It
// maps filename patterns (Regular, Bold, Italic, BoldItalic) to fpdf styles.
// Returns the regular font bytes (for glyph detection) or an error.
func LoadCustomFonts(pdf *fpdf.Fpdf, archivePath string) ([]byte, error) {
	data, err := os.ReadFile(archivePath)
	if err != nil {
		return nil, fmt.Errorf("read font archive: %w", err)
	}

	var fonts map[string][]byte
	lower := strings.ToLower(archivePath)
	switch {
	case strings.HasSuffix(lower, ".zip"):
		fonts, err = extractTTFFromZip(data)
	case strings.HasSuffix(lower, ".tar.gz") || strings.HasSuffix(lower, ".tgz"):
		fonts, err = extractTTFFromTarGz(data)
	default:
		return nil, fmt.Errorf("unsupported archive format: %s (expected .zip or .tar.gz)", archivePath)
	}
	if err != nil {
		return nil, fmt.Errorf("extract fonts: %w", err)
	}
	if len(fonts) == 0 {
		return nil, fmt.Errorf("no .ttf files found in %s", archivePath)
	}

	return registerExtractedFonts(pdf, fonts)
}

// registerExtractedFonts maps extracted TTF files to font styles and registers
// them under the FontBody family name. At minimum a regular font must be found.
func registerExtractedFonts(pdf *fpdf.Fpdf, fonts map[string][]byte) ([]byte, error) {
	// Style detection patterns, ordered most-specific first.
	type styleMapping struct {
		patterns []string
		style    string
	}
	mappings := []styleMapping{
		{patterns: []string{"bolditalic", "bold_italic", "bold-italic", "bi"}, style: "BI"},
		{patterns: []string{"bold", "-b.", "_b."}, style: "B"},
		{patterns: []string{"italic", "oblique", "-i.", "_i."}, style: "I"},
		{patterns: []string{"regular", "normal", "-r.", "_r."}, style: ""},
	}

	registered := map[string]bool{}
	var regularBytes []byte

	for filename, data := range fonts {
		lower := strings.ToLower(filename)
		style := "" // default to regular
		for _, m := range mappings {
			found := false
			for _, p := range m.patterns {
				if strings.Contains(lower, p) {
					style = m.style
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !registered[style] {
			pdf.AddUTF8FontFromBytes(FontBody, style, data)
			registered[style] = true
			if style == "" {
				regularBytes = data
			}
		}
	}

	if regularBytes == nil {
		// No explicit regular found; use any font as the regular variant.
		for filename, data := range fonts {
			pdf.AddUTF8FontFromBytes(FontBody, "", data)
			regularBytes = data
			_ = filename
			break
		}
	}
	if regularBytes == nil {
		return nil, fmt.Errorf("could not determine a regular font variant")
	}
	// Register the regular font under any missing styles so that fpdf
	// does not error when SetFont is called with bold/italic.
	for _, style := range []string{"B", "I", "BI"} {
		if !registered[style] {
			pdf.AddUTF8FontFromBytes(FontBody, style, regularBytes)
		}
	}
	return regularBytes, nil
}

// extractTTFFromZip extracts .ttf files from a zip archive in memory.
func extractTTFFromZip(data []byte) (map[string][]byte, error) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}
	fonts := make(map[string][]byte)
	for _, f := range r.File {
		if !strings.HasSuffix(strings.ToLower(f.Name), ".ttf") {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		b, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			return nil, err
		}
		fonts[filepath.Base(f.Name)] = b
	}
	return fonts, nil
}

// extractTTFFromTarGz extracts .ttf files from a gzipped tar archive.
func extractTTFFromTarGz(data []byte) (map[string][]byte, error) {
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer func() { _ = gz.Close() }()

	tr := tar.NewReader(gz)
	fonts := make(map[string][]byte)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if hdr.Typeflag == tar.TypeDir {
			continue
		}
		if !strings.HasSuffix(strings.ToLower(hdr.Name), ".ttf") {
			continue
		}
		b, err := io.ReadAll(tr)
		if err != nil {
			return nil, err
		}
		fonts[filepath.Base(hdr.Name)] = b
	}
	return fonts, nil
}
