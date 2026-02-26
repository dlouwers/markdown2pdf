package diagram

import (
	"errors"
	"testing"
)

func TestMermaidAvailable(t *testing.T) {
	// This just exercises the function — result depends on environment.
	_ = MermaidAvailable()
}

func TestRenderMermaidWhenNotAvailable(t *testing.T) {
	if MermaidAvailable() {
		t.Skip("mmdc is available; skipping unavailable test")
	}
	_, err := RenderMermaid("graph TD\n    A --> B")
	if !errors.Is(err, ErrMermaidNotFound) {
		t.Fatalf("expected ErrMermaidNotFound, got: %v", err)
	}
}

func TestRenderMermaidWhenAvailable(t *testing.T) {
	if !MermaidAvailable() {
		t.Skip("mmdc not available; skipping render test")
	}

	pngData, err := RenderMermaid("graph TD\n    A --> B")
	if err != nil {
		t.Fatalf("RenderMermaid: %v", err)
	}
	if len(pngData) == 0 {
		t.Fatal("expected PNG output")
	}
	// PNG files start with the PNG magic bytes.
	if len(pngData) < 8 || string(pngData[:4]) != "\x89PNG" {
		t.Fatal("output does not appear to be a valid PNG")
	}
}
