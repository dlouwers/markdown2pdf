# markdown2pdf

Convert Markdown documents to clean, professional PDFs — with syntax-highlighted code blocks, tables, images, and Mermaid/D2 diagrams.

## Features

- **Headings, paragraphs, inline styles** (bold, italic, code, links)
- **Syntax-highlighted code blocks** via [Chroma](https://github.com/alecthomas/chroma) (200+ languages)
- **GFM tables** with alignment, borders, and header styling
- **Ordered, unordered, and task lists** with nested support
- **Images** — PNG, JPEG, SVG, and base64 data URIs
- **Mermaid diagrams** — rendered via `mmdc` (mermaid-cli)
- **D2 diagrams** — rendered natively via the D2 Go library
- **Table of contents** — optional `--toc` flag with clickable links
- **Orphan protection** — headings never appear stranded at the bottom of a page
- **Noto Sans font** — embedded for full UTF-8 support (override with `--font`)
- **Single static binary** — no runtime dependencies (except `mmdc` for Mermaid)

## Installation

### From binary releases

Download a prebuilt binary from [GitHub Releases](https://github.com/dlouwers/markdown2pdf/releases):

```bash
# macOS (Apple Silicon)
curl -Lo markdown2pdf.tar.gz https://github.com/dlouwers/markdown2pdf/releases/latest/download/markdown2pdf_*_darwin_arm64.tar.gz
tar xzf markdown2pdf.tar.gz
sudo mv markdown2pdf /usr/local/bin/
```

### From source

```bash
go install github.com/dlouwers/markdown2pdf/cmd/markdown2pdf@latest
```

Requires Go 1.25 or later.

## Usage

```bash
# Basic conversion (output: document.pdf)
markdown2pdf document.md

# Specify output path
markdown2pdf -o output.pdf document.md

# Generate with table of contents
markdown2pdf --toc document.md

# Print version
markdown2pdf --version

# Use a custom font (zip or tar.gz with TTF files)
markdown2pdf --font /path/to/MyFont.zip document.md
```

## Diagram support

### Mermaid

Fenced code blocks tagged with `mermaid` are rendered to PNG via [mermaid-cli](https://github.com/mermaid-js/mermaid-cli). Install it with:

```bash
npm install -g @mermaid-js/mermaid-cli
```

If `mmdc` is not available, Mermaid blocks render as a placeholder with the raw source.

### D2

Fenced code blocks tagged with `d2` are rendered natively using the [D2 Go library](https://github.com/terrastruct/d2) — no external tools required.

### Browser detection

Diagram rendering uses a headless Chromium-based browser. The following browsers are auto-detected (in order):

1. `CHROME_PATH` environment variable (if set)
2. Brave Browser
3. Google Chrome
4. Chromium
5. Microsoft Edge

## Building

The project uses a DevContainer with all required tooling pre-installed:

```bash
# Build
go build -o markdown2pdf ./cmd/markdown2pdf

# Test
go test ./... -count=1

# Lint
golangci-lint run --timeout=5m ./...

# Static binary (no CGO)
CGO_ENABLED=0 go build -trimpath -o markdown2pdf ./cmd/markdown2pdf
```

## Releasing

Releases are automated with [GoReleaser](https://goreleaser.com/). Push a semver tag to create a release:

```bash
git tag v0.0.1
git push --tags
```

This builds binaries for:
- macOS (amd64, arm64)
- Linux (amd64, arm64)
- Windows (amd64)

## License

See [LICENSE](LICENSE) for details.
