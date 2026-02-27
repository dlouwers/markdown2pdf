package renderer

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-pdf/fpdf"
	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
	"github.com/yuin/goldmark/ast"

	"github.com/dlouwers/markdown2pdf/internal/pdf"
)

// imageCounter is used to generate unique names for in-memory images.
var imageCounter int

func renderImage(state *renderState, img *ast.Image, source []byte) {
	dest := string(img.Destination)
	if dest == "" {
		return
	}

	state.fpdf.Ln(pdf.ImageMarginV)

	if strings.HasPrefix(dest, "data:") {
		renderDataURIImage(state, dest)
	} else if isSVGPath(dest) {
		renderSVGImage(state, dest)
	} else {
		renderFileImage(state, dest)
	}

	state.fpdf.Ln(pdf.ImageMarginV)
}

// renderFileImage embeds a PNG/JPEG/GIF image from a file path.
func renderFileImage(state *renderState, path string) {
	// Resolve relative path from base directory if set.
	resolvedPath := resolvePath(state, path)

	info, err := os.Stat(resolvedPath)
	if err != nil || info.IsDir() {
		renderImagePlaceholder(state, path, "file not found")
		return
	}

	imgType := detectImageType(resolvedPath)
	if imgType == "" {
		renderImagePlaceholder(state, path, "unsupported format")
		return
	}

	opt := fpdf.ImageOptions{ImageType: imgType, ReadDpi: true}
	infoPtr := state.fpdf.RegisterImageOptions(resolvedPath, opt)
	if state.fpdf.Err() {
		renderImagePlaceholder(state, path, "failed to load")
		return
	}

	embedImage(state, resolvedPath, infoPtr)
}

// renderSVGImage rasterizes an SVG file and embeds it as PNG.
func renderSVGImage(state *renderState, path string) {
	resolvedPath := resolvePath(state, path)

	svgData, err := os.ReadFile(resolvedPath)
	if err != nil {
		renderImagePlaceholder(state, path, "SVG file not found")
		return
	}

	pngData, err := rasterizeSVG(svgData, pdf.SVGRenderScale)
	if err != nil {
		renderImagePlaceholder(state, path, fmt.Sprintf("SVG render error: %v", err))
		return
	}

	embedPNGBytes(state, pngData, path)
}

// renderDataURIImage decodes a data URI and embeds the image.
func renderDataURIImage(state *renderState, uri string) {
	mediaType, data, err := parseDataURI(uri)
	if err != nil {
		renderImagePlaceholder(state, "data URI", err.Error())
		return
	}

	if strings.Contains(mediaType, "svg") {
		pngData, err := rasterizeSVG(data, pdf.SVGRenderScale)
		if err != nil {
			renderImagePlaceholder(state, "data URI (SVG)", fmt.Sprintf("SVG render error: %v", err))
			return
		}
		embedPNGBytes(state, pngData, "data:svg")
		return
	}

	imgType := imageTypeFromMediaType(mediaType)
	if imgType == "" {
		renderImagePlaceholder(state, "data URI", "unsupported media type: "+mediaType)
		return
	}

	embedImageBytes(state, data, imgType, "data:"+mediaType)
}

// rasterizeSVG converts SVG bytes to PNG bytes at the given scale factor.
func rasterizeSVG(svgData []byte, scale float64) ([]byte, error) {
	icon, err := oksvg.ReadIconStream(bytes.NewReader(svgData), oksvg.WarnErrorMode)
	if err != nil {
		return nil, fmt.Errorf("parse SVG: %w", err)
	}

	w := int(icon.ViewBox.W * scale)
	h := int(icon.ViewBox.H * scale)
	if w <= 0 || h <= 0 {
		return nil, fmt.Errorf("invalid SVG dimensions: %dx%d", w, h)
	}

	icon.SetTarget(0, 0, float64(w), float64(h))

	img := image.NewRGBA(image.Rect(0, 0, w, h))
	scanner := rasterx.NewScannerGV(w, h, img, img.Bounds())
	raster := rasterx.NewDasher(w, h, scanner)
	icon.Draw(raster, 1.0)

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("encode PNG: %w", err)
	}

	return buf.Bytes(), nil
}

// Image placement constants inspired by LaTeX float placement defaults.
// LaTeX uses \textfraction=0.2 (min 20% text per page) and \topfraction=0.7
// (max 70% for floats). We use a shrink-to-fit threshold: if remaining space
// is >= this fraction of a full page, shrink the image to fit rather than
// wasting space with a page break. Below this threshold, a page break
// produces a better-looking result.
const (
	// minShrinkThreshold is the minimum remaining page fraction (of full content
	// height) at which we shrink an image to fit rather than breaking to a new
	// page. At 0.4 (40%), an image is shrunk to at most 60% of its natural size
	// before we prefer a page break. This balances space usage against diagram
	// readability.
	minShrinkThreshold = 0.4
)

// embedImage places a registered image at the current position, scaled to fit
// available space using LaTeX-inspired placement heuristics:
//
//  1. Scale to fit page width (preserving aspect ratio).
//  2. If the image fits the remaining space, place it as-is.
//  3. If it doesn't fit but remaining space >= 40% of a full page,
//     shrink it to fill the remaining space (avoids large whitespace gaps).
//  4. If remaining space < 40%, add a page break and place at natural size.
//  5. If the image is taller than a full page, scale to fit one page.
//
// We use flow=false because we handle page breaks manually to avoid
// double-break conflicts with fpdf's built-in auto-pagination.
func embedImage(state *renderState, name string, info *fpdf.ImageInfoType) {
	imgW, imgH := info.Extent()
	if imgW <= 0 || imgH <= 0 {
		return
	}

	maxW := contentWidth(state)
	w, h := scaleToFit(imgW, imgH, maxW)

	// Compute page geometry.
	_, topMargin, _, bottomMargin := state.fpdf.GetMargins()
	_, pageH := state.fpdf.GetPageSize()
	maxPageH := pageH - topMargin - bottomMargin
	remaining := pageH - bottomMargin - state.fpdf.GetY()

	switch {
	case h <= remaining:
		// Fits on current page — place as-is.

	case remaining >= minShrinkThreshold*maxPageH:
		// Doesn't fit, but enough space to shrink into without looking bad.
		scale := remaining / h
		w *= scale
		h = remaining

	default:
		// Not enough room — start a fresh page.
		state.fpdf.AddPage()

	}

	// If image is still taller than a full page (very large diagram), scale to fit.
	if h > maxPageH {
		scale := maxPageH / h
		w *= scale
		h *= scale
	}

	state.fpdf.Image(name, state.fpdf.GetX(), state.fpdf.GetY(), w, h, false, "", 0, "")
	// Advance Y position manually since flow=false.
	state.fpdf.SetY(state.fpdf.GetY() + h)
}

// embedPNGBytes registers PNG bytes and embeds the image.
func embedPNGBytes(state *renderState, pngData []byte, label string) {
	embedImageBytes(state, pngData, "PNG", label)
}

// embedImageBytes registers image bytes of a given type and embeds the image.
func embedImageBytes(state *renderState, data []byte, imgType, label string) {
	imageCounter++
	name := fmt.Sprintf("__img_%d_%s", imageCounter, label)

	opt := fpdf.ImageOptions{ImageType: imgType}
	info := state.fpdf.RegisterImageOptionsReader(name, opt, bytes.NewReader(data))
	if state.fpdf.Err() {
		renderImagePlaceholder(state, label, "failed to register image")
		return
	}

	embedImage(state, name, info)
}

// embedEmojiInline embeds a Twemoji PNG inline at the current cursor position.
// Unlike embedImage (block-level), this advances X not Y.
// Adds spacing before/after emoji for better visual separation from surrounding text.
// Returns true on success, false if embedding failed (caller should fallback to font).
func embedEmojiInline(state *renderState, pngData []byte, r rune) bool {
	imageCounter++
	name := fmt.Sprintf("emoji_%x_%d", r, imageCounter)

	opt := fpdf.ImageOptions{ImageType: "PNG"}
	_ = state.fpdf.RegisterImageOptionsReader(name, opt, bytes.NewReader(pngData))
	if state.fpdf.Err() {
		return false // Silent fail, caller handles fallback
	}

	// Size emoji to match line height (slightly smaller for better alignment)
	size := pdf.LineHeight * 0.9

	// Add spacing before/after emoji (0.15em is standard for inline emoji)
	// This prevents emoji from touching adjacent characters
	spacing := pdf.FontSizeBody * 0.15

	// Get current position
	x, y := state.fpdf.GetX(), state.fpdf.GetY()

	// Add leading space
	x += spacing

	// Calculate baseline offset to align emoji bottom with text baseline
	// Text sits on baseline; emoji should too
	yOffset := (pdf.LineHeight - size) / 2

	// Embed image at current position with baseline offset
	state.fpdf.Image(name, x, y+yOffset, size, size, false, "", 0, "")

	// Advance cursor by image width + trailing space
	state.fpdf.SetX(x + size + spacing)

	return true
}

// scaleToFit scales width and height to fit within maxWidth, preserving aspect ratio.
func scaleToFit(imgW, imgH, maxWidth float64) (float64, float64) {
	if imgW <= maxWidth {
		return imgW, imgH
	}
	scale := maxWidth / imgW
	return imgW * scale, imgH * scale
}

// contentWidth returns the available content width on the page.
func contentWidth(state *renderState) float64 {
	left, _, right, _ := state.fpdf.GetMargins()
	pageW, _ := state.fpdf.GetPageSize()
	return pageW - left - right
}

// renderImagePlaceholder draws a labeled box when an image can't be rendered.
func renderImagePlaceholder(state *renderState, label, reason string) {
	left, _, _, _ := state.fpdf.GetMargins()
	y := state.fpdf.GetY()
	w := contentWidth(state)
	h := pdf.LineHeight * 3

	state.fpdf.SetDrawColor(200, 200, 200)
	state.fpdf.SetFillColor(250, 250, 250)
	state.fpdf.Rect(left, y, w, h, "FD")

	state.fpdf.SetFont(pdf.FontBody, "I", pdf.FontSizeBody-1)
	state.fpdf.SetTextColor(150, 150, 150)
	text := fmt.Sprintf("[Image: %s \u2014 %s]", label, reason)
	state.fpdf.SetXY(left+pdf.CodeBlockPadding, y+pdf.CodeBlockPadding)
	state.fpdf.Write(pdf.LineHeight, text)

	state.fpdf.SetXY(left, y+h)
	resetFont(state)
}

// resolvePath resolves a relative image path against the base directory.
func resolvePath(state *renderState, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	if state.doc.BaseDir() != "" {
		return filepath.Join(state.doc.BaseDir(), path)
	}
	return path
}

// detectImageType returns the fpdf image type string for a file path.
func detectImageType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".png":
		return "PNG"
	case ".jpg", ".jpeg":
		return "JPEG"
	case ".gif":
		return "GIF"
	default:
		return ""
	}
}

// isSVGPath returns true if the path has an .svg extension.
func isSVGPath(path string) bool {
	return strings.ToLower(filepath.Ext(path)) == ".svg"
}

// parseDataURI parses a data URI and returns the media type and decoded bytes.
func parseDataURI(uri string) (string, []byte, error) {
	// Format: data:[<mediatype>][;base64],<data>
	if !strings.HasPrefix(uri, "data:") {
		return "", nil, fmt.Errorf("not a data URI")
	}

	rest := uri[5:]
	commaIdx := strings.Index(rest, ",")
	if commaIdx < 0 {
		return "", nil, fmt.Errorf("malformed data URI: no comma")
	}

	meta := rest[:commaIdx]
	encoded := rest[commaIdx+1:]

	isBase64 := strings.HasSuffix(meta, ";base64")
	mediaType := strings.TrimSuffix(meta, ";base64")
	if mediaType == "" {
		mediaType = "application/octet-stream"
	}

	if !isBase64 {
		return "", nil, fmt.Errorf("only base64 data URIs are supported")
	}

	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		// Try URL-safe encoding or with padding.
		data, err = base64.RawStdEncoding.DecodeString(encoded)
		if err != nil {
			return "", nil, fmt.Errorf("decode base64: %w", err)
		}
	}

	return mediaType, data, nil
}

// imageTypeFromMediaType maps a MIME type to fpdf image type.
func imageTypeFromMediaType(mediaType string) string {
	switch {
	case strings.Contains(mediaType, "png"):
		return "PNG"
	case strings.Contains(mediaType, "jpeg") || strings.Contains(mediaType, "jpg"):
		return "JPEG"
	case strings.Contains(mediaType, "gif"):
		return "GIF"
	default:
		return ""
	}
}
