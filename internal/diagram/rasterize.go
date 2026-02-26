// Package diagram provides SVG rasterization via headless Chromium.
package diagram

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

// rasterizeTimeout is the maximum time allowed for a single SVG rasterization.
const rasterizeTimeout = 30 * time.Second

// browserPool manages a shared headless Chromium instance for SVG rasterization.
// The browser is lazily initialized on first use and reused across calls.
var browserPool struct {
	once   sync.Once
	ctx    context.Context
	cancel context.CancelFunc
	err    error
}

// initBrowser lazily starts a headless Chromium instance.
func initBrowser() (context.Context, error) {
	browserPool.once.Do(func() {
		opts := append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("no-sandbox", true),
			chromedp.Flag("disable-dev-shm-usage", true),
			chromedp.Flag("disable-gpu", true),
			chromedp.Flag("disable-software-rasterizer", true),
			chromedp.Flag("headless", true),
		)

		// Use a user-specified or auto-detected Chromium-based browser if
		// chromedp's built-in search (Chrome/Chromium only) would miss it.
		if p := findBrowser(); p != "" {
			opts = append(opts, chromedp.ExecPath(p))
		}

		allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)
		ctx, cancel := chromedp.NewContext(allocCtx)

		browserPool.ctx = ctx
		browserPool.cancel = func() {
			cancel()
			allocCancel()
		}

		// Warm up the browser by running a no-op.
		browserPool.err = chromedp.Run(ctx)
	})

	return browserPool.ctx, browserPool.err
}

// ChromiumAvailable reports whether a headless Chromium browser can be started.
func ChromiumAvailable() bool {
	_, err := initBrowser()
	return err == nil
}

// RasterizeSVGChrome converts SVG bytes to PNG bytes using headless Chromium.
// The SVG is mounted in an HTML page and the <svg> element is screenshotted at
// 2× device pixel ratio for crisp rendering. This approach handles all SVG
// features including <text>, <marker>, <mask>, CSS styles, and embedded fonts.
func RasterizeSVGChrome(svgBytes []byte) ([]byte, error) {
	parentCtx, err := initBrowser()
	if err != nil {
		return nil, fmt.Errorf("chromium not available: %w", err)
	}

	ctx, cancel := chromedp.NewContext(parentCtx)
	defer cancel()

	ctx, timeoutCancel := context.WithTimeout(ctx, rasterizeTimeout)
	defer timeoutCancel()

	// Ensure the outer SVG has explicit width/height attributes so headless
	// Chromium renders it at the correct size. D2 SVGs typically have a viewBox
	// but omit explicit width/height, causing the element to collapse to 0×0.
	svgStr := ensureSVGDimensions(string(svgBytes))

	html := `<!doctype html><meta charset="utf-8">
<style>
  html, body { margin: 0; padding: 0; background: #fff; }
  #stage { display: inline-block; }
</style>
<div id="stage">` + svgStr + `</div>`

	dataURI := "data:text/html;base64," + base64.StdEncoding.EncodeToString([]byte(html))

	var pngBuf []byte

	err = chromedp.Run(ctx,
		chromedp.EmulateViewport(1920, 1080, chromedp.EmulateScale(2)),
		chromedp.Navigate(dataURI),
		chromedp.WaitReady("#stage", chromedp.ByQuery),
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Find the SVG element and screenshot it directly for tight
			// bounding-box output (no extra whitespace).
			node, err := dom.GetDocument().Do(ctx)
			if err != nil {
				return err
			}
			stageNodes, err := dom.QuerySelectorAll(node.NodeID, "#stage").Do(ctx)
			if err != nil || len(stageNodes) == 0 {
				return fmt.Errorf("no #stage element found in rendered page")
			}

			// Get the box model of the SVG element for precise screenshot.
			box, err := dom.GetBoxModel().WithNodeID(stageNodes[0]).Do(ctx)
			if err != nil {
				return fmt.Errorf("get stage box model: %w", err)
			}

			// The content quad gives us the exact pixel coordinates.
			// Quad is [x1,y1, x2,y2, x3,y3, x4,y4] for the four corners.
			quad := box.Content
			x := quad[0]
			y := quad[1]
			w := quad[2] - quad[0]
			h := quad[5] - quad[1]

			if w <= 0 || h <= 0 {
				return fmt.Errorf("SVG has zero dimensions: %.0fx%.0f", w, h)
			}

			// Capture at device pixel ratio for high-DPI output.
			clip := &page.Viewport{
				X:      x,
				Y:      y,
				Width:  w,
				Height: h,
				Scale:  1,
			}
			data, err := page.CaptureScreenshot().
				WithClip(clip).
				WithCaptureBeyondViewport(true).
				Do(ctx)
			if err != nil {
				return fmt.Errorf("screenshot SVG: %w", err)
			}
			pngBuf = data
			return nil
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("rasterize SVG: %w", err)
	}

	if len(pngBuf) == 0 {
		return nil, fmt.Errorf("screenshot produced empty PNG")
	}

	return pngBuf, nil
}

// firstSVGTagRe matches the first <svg ...> opening tag.
var firstSVGTagRe = regexp.MustCompile(`(?s)(<svg\b[^>]*>)`)

// ensureSVGDimensions inspects the root (first) <svg> element and adds explicit
// width and height attributes derived from the viewBox if they are missing.
// This prevents headless Chromium from collapsing SVGs with only a viewBox to 0×0.
func ensureSVGDimensions(svg string) string {
	m := firstSVGTagRe.FindStringSubmatch(svg)
	if m == nil {
		return svg
	}
	firstTag := m[1]

	// Already has explicit width — nothing to do.
	if regexp.MustCompile(`\swidth\s*=`).MatchString(firstTag) {
		return svg
	}

	// Extract viewBox from the first <svg> tag.
	vb := regexp.MustCompile(`viewBox="([^"]+)"`).FindStringSubmatch(firstTag)
	if vb == nil {
		return svg
	}

	// Parse viewBox="minX minY width height".
	parts := regexp.MustCompile(`[\s,]+`).Split(vb[1], -1)
	if len(parts) != 4 {
		return svg
	}
	vbW, err1 := strconv.ParseFloat(parts[2], 64)
	vbH, err2 := strconv.ParseFloat(parts[3], 64)
	if err1 != nil || err2 != nil || vbW <= 0 || vbH <= 0 {
		return svg
	}

	// Inject width and height into the first <svg> tag only.
	attrs := fmt.Sprintf(` width="%.0f" height="%.0f"`, vbW, vbH)
	newTag := `<svg` + attrs + firstTag[4:] // Insert after "<svg"
	return strings.Replace(svg, firstTag, newTag, 1)
}

// findBrowser returns the path to a Chromium-based browser, checking:
//  1. CHROME_PATH environment variable (explicit user override)
//  2. Well-known locations for Brave, Chrome, Chromium, and Edge
//
// Returns "" if nothing is found, letting chromedp fall back to its own search.
func findBrowser() string {
	if p := os.Getenv("CHROME_PATH"); p != "" {
		if _, err := exec.LookPath(p); err == nil {
			return p
		}
		// Might be an absolute path that LookPath can't find but exists.
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	for _, p := range browserPaths() {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

// browserPaths returns platform-specific paths to Chromium-based browsers,
// ordered by preference: Brave, Chrome, Chromium, Edge.
func browserPaths() []string {
	switch runtime.GOOS {
	case "darwin":
		return []string{
			"/Applications/Brave Browser.app/Contents/MacOS/Brave Browser",
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
			"/Applications/Chromium.app/Contents/MacOS/Chromium",
			"/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge",
		}
	case "windows":
		userProfile := os.Getenv("USERPROFILE")
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			localAppData = userProfile + `\AppData\Local`
		}
		return []string{
			localAppData + `\BraveSoftware\Brave-Browser\Application\brave.exe`,
			`C:\Program Files\BraveSoftware\Brave-Browser\Application\brave.exe`,
			`C:\Program Files (x86)\BraveSoftware\Brave-Browser\Application\brave.exe`,
			localAppData + `\Google\Chrome\Application\chrome.exe`,
			`C:\Program Files\Google\Chrome\Application\chrome.exe`,
			`C:\Program Files (x86)\Google\Chrome\Application\chrome.exe`,
			localAppData + `\Chromium\Application\chrome.exe`,
			`C:\Program Files\Microsoft\Edge\Application\msedge.exe`,
			`C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe`,
		}
	default: // Linux, FreeBSD, etc.
		return []string{
			"/usr/bin/brave-browser",
			"/usr/bin/brave",
			"/snap/bin/brave",
			"/usr/bin/google-chrome",
			"/usr/bin/google-chrome-stable",
			"/usr/bin/chromium",
			"/usr/bin/chromium-browser",
			"/snap/bin/chromium",
			"/usr/bin/microsoft-edge",
			"/usr/bin/microsoft-edge-stable",
		}
	}
}
