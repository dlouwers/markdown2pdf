// Package diagram handles rendering of D2 and Mermaid diagrams to images.
package diagram

import (
	"context"
	"fmt"

	"oss.terrastruct.com/d2/d2graph"
	"oss.terrastruct.com/d2/d2layouts/d2dagrelayout"
	"oss.terrastruct.com/d2/d2lib"
	"oss.terrastruct.com/d2/d2renderers/d2svg"
	"oss.terrastruct.com/d2/d2themes/d2themescatalog"
	"oss.terrastruct.com/d2/lib/log"
	"oss.terrastruct.com/d2/lib/textmeasure"
	"oss.terrastruct.com/util-go/go2"
)

// RenderD2 compiles D2 source code and returns PNG bytes.
// It uses the dagre layout engine and NeutralDefault theme (white background,
// dark lines — suitable for PDF documents). The D2 source is compiled to SVG
// and then rasterized to PNG via headless Chromium for full-fidelity rendering
// of text labels, arrowheads, markers, and embedded fonts.
//
// If headless Chromium is not available, returns an error.
func RenderD2(source string) ([]byte, error) {
	ruler, err := textmeasure.NewRuler()
	if err != nil {
		return nil, fmt.Errorf("textmeasure.NewRuler: %w", err)
	}

	layoutResolver := func(engine string) (d2graph.LayoutGraph, error) {
		return d2dagrelayout.DefaultLayout, nil
	}

	renderOpts := &d2svg.RenderOpts{
		Pad:     go2.Pointer(int64(5)),
		ThemeID: &d2themescatalog.NeutralDefault.ID,
	}

	compileOpts := &d2lib.CompileOptions{
		LayoutResolver: layoutResolver,
		Ruler:          ruler,
	}

	ctx := log.WithDefault(context.Background())
	diagram, _, err := d2lib.Compile(ctx, source, compileOpts, renderOpts)
	if err != nil {
		return nil, fmt.Errorf("d2lib.Compile: %w", err)
	}

	svgBytes, err := d2svg.Render(diagram, renderOpts)
	if err != nil {
		return nil, fmt.Errorf("d2svg.Render: %w", err)
	}

	// Rasterize the SVG to PNG using headless Chromium. This handles all SVG
	// features: text labels, arrowheads, markers, masks, CSS styles, and
	// embedded WOFF fonts — unlike oksvg which only supports basic shapes.
	pngBytes, err := RasterizeSVGChrome(svgBytes)
	if err != nil {
		return nil, fmt.Errorf("rasterize D2 SVG: %w", err)
	}

	return pngBytes, nil
}
