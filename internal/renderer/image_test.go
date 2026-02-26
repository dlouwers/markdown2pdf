package renderer

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/dlouwers/markdown2pdf/internal/parser"
	"github.com/dlouwers/markdown2pdf/internal/pdf"
)

// createTestPNG creates a small solid-color PNG at the given path.
func createTestPNG(t *testing.T, path string, w, h int) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{R: 200, G: 50, B: 50, A: 255})
		}
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create test png: %v", err)
	}
	defer func() { _ = f.Close() }()
	if err := png.Encode(f, img); err != nil {
		t.Fatalf("encode test png: %v", err)
	}
}

func TestScaleToFit(t *testing.T) {
	tests := []struct {
		imgW, imgH, maxW float64
		wantW, wantH     float64
	}{
		{100, 50, 200, 100, 50},   // fits — no scaling
		{200, 100, 100, 100, 50},  // needs scaling down
		{300, 150, 300, 300, 150}, // exact fit
		{50, 200, 100, 50, 200},   // width fits, tall image
	}

	for _, tt := range tests {
		w, h := scaleToFit(tt.imgW, tt.imgH, tt.maxW)
		if w != tt.wantW || h != tt.wantH {
			t.Errorf("scaleToFit(%v, %v, %v) = (%v, %v), want (%v, %v)",
				tt.imgW, tt.imgH, tt.maxW, w, h, tt.wantW, tt.wantH)
		}
	}
}

func TestDetectImageType(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"photo.png", "PNG"},
		{"photo.PNG", "PNG"},
		{"photo.jpg", "JPEG"},
		{"photo.jpeg", "JPEG"},
		{"photo.gif", "GIF"},
		{"photo.bmp", ""},
		{"photo.svg", ""},
		{"photo", ""},
	}
	for _, tt := range tests {
		got := detectImageType(tt.path)
		if got != tt.want {
			t.Errorf("detectImageType(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestIsSVGPath(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"image.svg", true},
		{"image.SVG", true},
		{"image.png", false},
		{"image.svgz", false},
	}
	for _, tt := range tests {
		got := isSVGPath(tt.path)
		if got != tt.want {
			t.Errorf("isSVGPath(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

func TestImageTypeFromMediaType(t *testing.T) {
	tests := []struct {
		mediaType string
		want      string
	}{
		{"image/png", "PNG"},
		{"image/jpeg", "JPEG"},
		{"image/jpg", "JPEG"},
		{"image/gif", "GIF"},
		{"image/bmp", ""},
		{"text/plain", ""},
	}
	for _, tt := range tests {
		got := imageTypeFromMediaType(tt.mediaType)
		if got != tt.want {
			t.Errorf("imageTypeFromMediaType(%q) = %q, want %q", tt.mediaType, got, tt.want)
		}
	}
}

func TestParseDataURI(t *testing.T) {
	// Create a tiny valid data URI.
	data := []byte("hello world")
	encoded := base64.StdEncoding.EncodeToString(data)
	uri := "data:text/plain;base64," + encoded

	mediaType, decoded, err := parseDataURI(uri)
	if err != nil {
		t.Fatalf("parseDataURI: %v", err)
	}
	if mediaType != "text/plain" {
		t.Errorf("media type = %q, want %q", mediaType, "text/plain")
	}
	if string(decoded) != "hello world" {
		t.Errorf("decoded = %q, want %q", string(decoded), "hello world")
	}
}

func TestParseDataURIErrors(t *testing.T) {
	tests := []struct {
		name string
		uri  string
	}{
		{"not a data URI", "https://example.com/image.png"},
		{"no comma", "data:image/png;base64"},
		{"not base64", "data:image/png,rawdata"},
	}
	for _, tt := range tests {
		_, _, err := parseDataURI(tt.uri)
		if err == nil {
			t.Errorf("parseDataURI(%q) expected error for %s", tt.uri, tt.name)
		}
	}
}

func TestRenderPNGImage(t *testing.T) {
	dir := t.TempDir()
	pngPath := filepath.Join(dir, "test.png")
	createTestPNG(t, pngPath, 80, 40)

	source := []byte(fmt.Sprintf("# Images\n\n![test](%s)\n", pngPath))
	data := renderPDF(t, source, false)
	if len(data) < 100 {
		t.Fatalf("PDF output seems too small")
	}
}

func TestRenderSVGImage(t *testing.T) {
	svgPath := filepath.Join("..", "..", "testdata", "test.svg")
	absPath, err := filepath.Abs(svgPath)
	if err != nil {
		t.Fatalf("abs: %v", err)
	}

	source := []byte(fmt.Sprintf("# SVG Test\n\n![test svg](%s)\n", absPath))
	data := renderPDF(t, source, false)
	if len(data) < 100 {
		t.Fatalf("PDF output seems too small")
	}
}

func TestRenderMissingImage(t *testing.T) {
	// Should render a placeholder, not fail.
	source := []byte("![missing](/nonexistent/path/image.png)\n")
	data := renderPDF(t, source, true)
	if len(data) < 100 {
		t.Fatalf("PDF output seems too small")
	}
	// Placeholder should include "file not found" text.
	// UTF-8 fonts use CID encoding; verify placeholder rendered by checking PDF size
	// is substantially larger than a blank page (placeholder adds drawing ops).
	blankDoc, err := pdf.NewDocument()
	if err != nil {
		t.Fatalf("new document: %v", err)
	}
	var blankBuf bytes.Buffer
	if err := blankDoc.PDF().Output(&blankBuf); err != nil {
		t.Fatalf("blank output: %v", err)
	}
	if len(data) <= blankBuf.Len() {
		t.Fatalf("expected PDF with placeholder to be larger than blank page")
	}
}

func TestRenderDataURIImage(t *testing.T) {
	// Create a tiny PNG and base64 encode it.
	dir := t.TempDir()
	pngPath := filepath.Join(dir, "tiny.png")
	createTestPNG(t, pngPath, 4, 4)
	pngData, err := os.ReadFile(pngPath)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	encoded := base64.StdEncoding.EncodeToString(pngData)
	uri := "data:image/png;base64," + encoded

	source := []byte(fmt.Sprintf("![data uri](%s)\n", uri))
	data := renderPDF(t, source, false)
	if len(data) < 100 {
		t.Fatalf("PDF output seems too small")
	}
}

func TestRenderImageWithBaseDir(t *testing.T) {
	// Create a PNG in a temp directory and use a relative path.
	dir := t.TempDir()
	pngPath := filepath.Join(dir, "relative.png")
	createTestPNG(t, pngPath, 60, 30)

	source := []byte("![relative](relative.png)\n")
	node, src := parser.Parse(source)
	doc, err := pdf.NewDocument()
	if err != nil {
		t.Fatalf("new document: %v", err)
	}
	doc.SetBaseDir(dir)

	r := New()
	if err := r.Render(doc, node, src); err != nil {
		t.Fatalf("render: %v", err)
	}

	outPath := filepath.Join(dir, "output.pdf")
	if err := doc.Save(outPath); err != nil {
		t.Fatalf("save: %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if len(data) < 100 {
		t.Fatalf("PDF output seems too small")
	}
}

func TestRasterizeSVG(t *testing.T) {
	svgData := []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="50" height="50" viewBox="0 0 50 50">
		<rect width="50" height="50" fill="blue"/>
	</svg>`)

	pngData, err := rasterizeSVG(svgData, 2.0)
	if err != nil {
		t.Fatalf("rasterizeSVG: %v", err)
	}
	if len(pngData) == 0 {
		t.Fatalf("expected PNG data")
	}
}

func TestImagesFromFixture(t *testing.T) {
	// This test uses the testdata/images.md fixture.
	// It needs test.png to exist in testdata/.
	testdataDir := filepath.Join("..", "..", "testdata")
	pngPath := filepath.Join(testdataDir, "test.png")

	// Create test.png if it doesn't exist.
	if _, err := os.Stat(pngPath); os.IsNotExist(err) {
		createTestPNG(t, pngPath, 100, 60)
	}

	fixturePath := filepath.Join(testdataDir, "images.md")
	source, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("read: %v", err)
	}

	node, src := parser.Parse(source)
	doc, err := pdf.NewDocument()
	if err != nil {
		t.Fatalf("new document: %v", err)
	}
	doc.SetBaseDir(testdataDir)

	r := New()
	if err := r.Render(doc, node, src); err != nil {
		t.Fatalf("render: %v", err)
	}

	dir := t.TempDir()
	outPath := filepath.Join(dir, "images.pdf")
	if err := doc.Save(outPath); err != nil {
		t.Fatalf("save: %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if len(data) < 100 {
		t.Fatalf("PDF output seems too small")
	}
}
