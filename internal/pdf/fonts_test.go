package pdf

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
)

func TestFontSupportsGlyphBullet(t *testing.T) {
	// Noto Sans should support the bullet character U+2022.
	if !FontSupportsGlyph(notoSansRegular, '•') {
		t.Fatal("expected Noto Sans to support U+2022 BULLET")
	}
}

func TestFontSupportsGlyphTriangularBullet(t *testing.T) {
	if !FontSupportsGlyph(notoSansRegular, '\u2023') {
		t.Fatal("expected Noto Sans to support U+2023 TRIANGULAR BULLET")
	}
}

func TestFontSupportsGlyphHyphenBullet(t *testing.T) {
	if !FontSupportsGlyph(notoSansRegular, '\u2043') {
		t.Fatal("expected Noto Sans to support U+2043 HYPHEN BULLET")
	}
}

func TestFontSupportsGlyphMissing(t *testing.T) {
	// Private Use Area codepoint — very unlikely to be in Noto Sans.
	if FontSupportsGlyph(notoSansRegular, '\uF8FF') {
		t.Fatal("did not expect Noto Sans to contain U+F8FF (Apple logo)")
	}
}

func TestFontDoesNotSupportGeometricShapes(t *testing.T) {
	// Noto Sans does not include geometric shape characters like ◦ or ▪.
	// This confirms the fallback path would be used for these.
	if FontSupportsGlyph(notoSansRegular, '◦') {
		t.Fatal("did not expect Noto Sans to support U+25E6 WHITE BULLET")
	}
	if FontSupportsGlyph(notoSansRegular, '▪') {
		t.Fatal("did not expect Noto Sans to support U+25AA BLACK SMALL SQUARE")
	}
}

func TestFontSupportsGlyphInvalidData(t *testing.T) {
	if FontSupportsGlyph([]byte("not a font"), '•') {
		t.Fatal("expected false for invalid font data")
	}
}

func TestFontSupportsGlyphEmptyData(t *testing.T) {
	if FontSupportsGlyph(nil, '•') {
		t.Fatal("expected false for nil font data")
	}
}

func TestExtractTTFFromZip(t *testing.T) {
	// Create a zip with a fake TTF and a non-TTF file.
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)

	f, err := w.Create("TestFont-Regular.ttf")
	if err != nil {
		t.Fatalf("zip create: %v", err)
	}
	if _, err := f.Write([]byte("fake-ttf-data")); err != nil {
		t.Fatalf("zip write: %v", err)
	}

	f2, err := w.Create("readme.txt")
	if err != nil {
		t.Fatalf("zip create: %v", err)
	}
	if _, err := f2.Write([]byte("not a font")); err != nil {
		t.Fatalf("zip write: %v", err)
	}

	if err := w.Close(); err != nil {
		t.Fatalf("zip close: %v", err)
	}

	fonts, err := extractTTFFromZip(buf.Bytes())
	if err != nil {
		t.Fatalf("extract: %v", err)
	}
	if len(fonts) != 1 {
		t.Fatalf("expected 1 TTF, got %d", len(fonts))
	}
	if _, ok := fonts["TestFont-Regular.ttf"]; !ok {
		t.Fatalf("expected TestFont-Regular.ttf in extracted fonts")
	}
}

func TestExtractTTFFromTarGz(t *testing.T) {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	data := []byte("fake-ttf-data")
	hdr := &tar.Header{
		Name: "fonts/TestFont-Bold.ttf",
		Size: int64(len(data)),
		Mode: 0644,
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatalf("tar header: %v", err)
	}
	if _, err := tw.Write(data); err != nil {
		t.Fatalf("tar write: %v", err)
	}

	// Also write a non-TTF file.
	txtData := []byte("readme")
	hdr2 := &tar.Header{
		Name: "README.md",
		Size: int64(len(txtData)),
		Mode: 0644,
	}
	if err := tw.WriteHeader(hdr2); err != nil {
		t.Fatalf("tar header: %v", err)
	}
	if _, err := tw.Write(txtData); err != nil {
		t.Fatalf("tar write: %v", err)
	}

	if err := tw.Close(); err != nil {
		t.Fatalf("tar close: %v", err)
	}
	if err := gw.Close(); err != nil {
		t.Fatalf("gzip close: %v", err)
	}

	fonts, err := extractTTFFromTarGz(buf.Bytes())
	if err != nil {
		t.Fatalf("extract: %v", err)
	}
	if len(fonts) != 1 {
		t.Fatalf("expected 1 TTF, got %d", len(fonts))
	}
	if _, ok := fonts["TestFont-Bold.ttf"]; !ok {
		t.Fatalf("expected TestFont-Bold.ttf in extracted fonts")
	}
}

func TestLoadCustomFontsZip(t *testing.T) {
	// Create a zip containing the embedded Noto Sans Regular font.
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	f, err := w.Create("NotoSans-Regular.ttf")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := f.Write(notoSansRegular); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("zip close: %v", err)
	}

	// Write the zip to a temp file.
	dir := t.TempDir()
	zipPath := filepath.Join(dir, "fonts.zip")
	if err := os.WriteFile(zipPath, buf.Bytes(), 0644); err != nil {
		t.Fatalf("write zip: %v", err)
	}

	doc, err := NewDocument(WithCustomFont(zipPath))
	if err != nil {
		t.Fatalf("new document: %v", err)
	}
	if doc.BodyFontBytes() == nil {
		t.Fatal("expected body font bytes after custom font load")
	}
	if !FontSupportsGlyph(doc.BodyFontBytes(), '•') {
		t.Fatal("expected custom-loaded Noto Sans to support bullet")
	}
}

func TestLoadCustomFontsUnsupportedFormat(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "font.rar")
	if err := os.WriteFile(path, []byte("data"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	_, err := NewDocument(WithCustomFont(path))
	if err == nil {
		t.Fatal("expected error for unsupported format")
	}
}

func TestLoadCustomFontsMissingFile(t *testing.T) {
	_, err := NewDocument(WithCustomFont("/nonexistent/fonts.zip"))
	if err == nil {
		t.Fatal("expected error for missing archive")
	}
}
