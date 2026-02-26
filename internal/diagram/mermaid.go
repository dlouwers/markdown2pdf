package diagram

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// ErrMermaidNotFound is returned when mmdc is not installed or not on PATH.
var ErrMermaidNotFound = fmt.Errorf("mmdc (mermaid-cli) not found on PATH")

// MermaidAvailable reports whether the mmdc binary is on PATH.
func MermaidAvailable() bool {
	_, err := exec.LookPath("mmdc")
	return err == nil
}

// RenderMermaid renders Mermaid diagram source to PNG bytes by shelling out to mmdc.
// Returns ErrMermaidNotFound if mmdc is not installed.
func RenderMermaid(source string) ([]byte, error) {
	mmdcPath, err := exec.LookPath("mmdc")
	if err != nil {
		return nil, ErrMermaidNotFound
	}

	tmpDir, err := os.MkdirTemp("", "mermaid-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	inputFile := filepath.Join(tmpDir, "diagram.mmd")
	outputFile := filepath.Join(tmpDir, "diagram.png")
	puppeteerConfig := filepath.Join(tmpDir, "puppeteer-config.json")

	if err := os.WriteFile(inputFile, []byte(source), 0o644); err != nil {
		return nil, fmt.Errorf("write input: %w", err)
	}

	// Puppeteer config for container compatibility (headless Chrome sandbox).
	puppeteerCfg := `{
  "headless": "shell",
  "args": [
    "--no-sandbox",
    "--disable-setuid-sandbox",
    "--disable-dev-shm-usage",
    "--disable-gpu"
  ]
}`
	if err := os.WriteFile(puppeteerConfig, []byte(puppeteerCfg), 0o644); err != nil {
		return nil, fmt.Errorf("write puppeteer config: %w", err)
	}

	cmd := exec.Command(mmdcPath,
		"-i", inputFile,
		"-o", outputFile,
		"-t", "neutral",
		"-b", "white",
		"-w", "1600",
		"-s", "3",
		"-p", puppeteerConfig,
		"-q",
	)
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("mmdc: %w", err)
	}

	data, err := os.ReadFile(outputFile)
	if err != nil {
		return nil, fmt.Errorf("read output: %w", err)
	}

	return data, nil
}
