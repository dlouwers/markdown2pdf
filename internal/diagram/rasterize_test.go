package diagram

import (
	"testing"
)

func TestRasterizeSVGChromeSimple(t *testing.T) {
	if !ChromiumAvailable() {
		t.Skip("headless Chromium not available")
	}

	svg := []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="200" height="100">
  <rect x="10" y="10" width="180" height="80" fill="#E3F2FD" stroke="#1565C0" stroke-width="2"/>
  <text x="100" y="55" text-anchor="middle" font-size="16" fill="#0A0F25">Hello</text>
</svg>`)

	pngBytes, err := RasterizeSVGChrome(svg)
	if err != nil {
		t.Fatalf("RasterizeSVGChrome: %v", err)
	}
	assertPNG(t, pngBytes)

	// A 200×100 SVG at 2× should produce a non-trivial PNG.
	if len(pngBytes) < 500 {
		t.Fatalf("PNG too small: %d bytes", len(pngBytes))
	}
}

func TestRasterizeSVGChromeWithStyles(t *testing.T) {
	if !ChromiumAvailable() {
		t.Skip("headless Chromium not available")
	}

	// SVG with embedded CSS styles and text — the features oksvg can't handle.
	svg := []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="300" height="150">
  <style>
    .box { fill: #F7F8FE; stroke: #0D32B2; stroke-width: 2; }
    .label { font-family: sans-serif; font-size: 14px; fill: #0A0F25; text-anchor: middle; }
  </style>
  <rect class="box" x="10" y="10" width="130" height="50"/>
  <text class="label" x="75" y="40">Service A</text>
  <rect class="box" x="160" y="10" width="130" height="50"/>
  <text class="label" x="225" y="40">Service B</text>
  <line x1="140" y1="35" x2="160" y2="35" stroke="#0D32B2" stroke-width="2" marker-end="url(#arrow)"/>
  <defs>
    <marker id="arrow" markerWidth="10" markerHeight="7" refX="10" refY="3.5" orient="auto">
      <polygon points="0 0, 10 3.5, 0 7" fill="#0D32B2"/>
    </marker>
  </defs>
</svg>`)

	pngBytes, err := RasterizeSVGChrome(svg)
	if err != nil {
		t.Fatalf("RasterizeSVGChrome: %v", err)
	}
	assertPNG(t, pngBytes)
}

func TestChromiumAvailable(t *testing.T) {
	// This just verifies the function doesn't panic. The actual result depends
	// on whether Chrome/Chromium is installed on the system.
	available := ChromiumAvailable()
	t.Logf("ChromiumAvailable: %v", available)
}
