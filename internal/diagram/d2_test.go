package diagram

import (
	"testing"
)

func TestRenderD2Simple(t *testing.T) {
	if !ChromiumAvailable() {
		t.Skip("headless Chromium not available")
	}

	pngBytes, err := RenderD2("x -> y")
	if err != nil {
		t.Fatalf("RenderD2: %v", err)
	}
	if len(pngBytes) == 0 {
		t.Fatal("expected PNG output")
	}
	assertPNG(t, pngBytes)
}

func TestRenderD2MultiNode(t *testing.T) {
	if !ChromiumAvailable() {
		t.Skip("headless Chromium not available")
	}

	source := `
User -> API: request
API -> Database: query
Database -> API: result
API -> User: response
`
	pngBytes, err := RenderD2(source)
	if err != nil {
		t.Fatalf("RenderD2: %v", err)
	}
	assertPNG(t, pngBytes)

	// PNG should be non-trivial (multi-node diagram is larger than a simple one).
	if len(pngBytes) < 1000 {
		t.Fatalf("PNG too small for multi-node diagram: %d bytes", len(pngBytes))
	}
}

func TestRenderD2InvalidSource(t *testing.T) {
	if !ChromiumAvailable() {
		t.Skip("headless Chromium not available")
	}

	// Completely malformed D2 — should return an error.
	_, err := RenderD2("{{{{invalid}}}}")
	if err == nil {
		t.Fatal("expected error for invalid D2 source")
	}
}

func TestRenderD2EmptySource(t *testing.T) {
	if !ChromiumAvailable() {
		t.Skip("headless Chromium not available")
	}

	// Empty source should still compile (produces empty diagram).
	pngBytes, err := RenderD2("")
	if err != nil {
		t.Fatalf("RenderD2 with empty source: %v", err)
	}
	if len(pngBytes) == 0 {
		t.Fatal("expected PNG output even for empty source")
	}
	assertPNG(t, pngBytes)
}

func TestRenderD2ProducesPNG(t *testing.T) {
	if !ChromiumAvailable() {
		t.Skip("headless Chromium not available")
	}

	pngBytes, err := RenderD2("a -> b -> c")
	if err != nil {
		t.Fatalf("RenderD2: %v", err)
	}
	assertPNG(t, pngBytes)
}

// assertPNG checks that the given bytes start with the PNG magic header.
func assertPNG(t *testing.T, data []byte) {
	t.Helper()
	if len(data) < 8 {
		t.Fatal("data too short to be PNG")
	}
	// PNG magic bytes: 137 80 78 71 13 10 26 10
	pngMagic := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	for i, b := range pngMagic {
		if data[i] != b {
			t.Fatalf("not a valid PNG: byte %d is 0x%02x, expected 0x%02x", i, data[i], b)
		}
	}
}
