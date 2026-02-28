# markdown2pdf

[![CI Status](https://github.com/dlouwers/markdown2pdf/actions/workflows/ci.yml/badge.svg)](https://github.com/dlouwers/markdown2pdf/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/dlouwers/markdown2pdf.svg)](https://pkg.go.dev/github.com/dlouwers/markdown2pdf)
[![Go Report Card](https://goreportcard.com/badge/github.com/dlouwers/markdown2pdf)](https://goreportcard.com/report/github.com/dlouwers/markdown2pdf)
[![Release](https://img.shields.io/github/v/release/dlouwers/markdown2pdf)](https://github.com/dlouwers/markdown2pdf/releases)

A command-line tool that converts Markdown to beautifully formatted PDFs with syntax-highlighted code blocks, tables, diagrams, and professional typography.

## What is it?

markdown2pdf takes your Markdown documents and generates clean, professional PDFs ready for distribution. Unlike browser-based converters, it's a single static binary with embedded fonts, smart typography, and native support for Mermaid and D2 diagrams. Perfect for documentation, reports, technical specs, and any content that needs to look polished in print.

## Quick Start

```bash
# Install via Homebrew
brew install dlouwers/tap/markdown2pdf

# Convert markdown to PDF
markdown2pdf document.md
# → Generates document.pdf

# With cover page and table of contents
markdown2pdf --cover-page --toc document.md
```

## Installation

### Package Manager (Recommended)

#### macOS / Linux

```bash
brew install dlouwers/tap/markdown2pdf
```

### Pre-built Binaries

Download from [GitHub Releases](https://github.com/dlouwers/markdown2pdf/releases) — available for macOS (Intel/ARM), Linux (AMD64/ARM64), and Windows (AMD64/ARM64).

<details>
<summary><b>macOS Installation</b></summary>

```bash
# Apple Silicon (M1/M2/M3)
curl -Lo markdown2pdf.tar.gz https://github.com/dlouwers/markdown2pdf/releases/download/v1.2.2/markdown2pdf_1.2.2_darwin_arm64.tar.gz
tar xzf markdown2pdf.tar.gz
sudo mv markdown2pdf /usr/local/bin/

# Intel
curl -Lo markdown2pdf.tar.gz https://github.com/dlouwers/markdown2pdf/releases/download/v1.2.2/markdown2pdf_1.2.2_darwin_amd64.tar.gz
tar xzf markdown2pdf.tar.gz
sudo mv markdown2pdf /usr/local/bin/
```

</details>

<details>
<summary><b>Linux Installation</b></summary>

```bash
# ARM64
curl -Lo markdown2pdf.tar.gz https://github.com/dlouwers/markdown2pdf/releases/download/v1.2.2/markdown2pdf_1.2.2_linux_arm64.tar.gz
tar xzf markdown2pdf.tar.gz
sudo mv markdown2pdf /usr/local/bin/

# AMD64/x86_64
curl -Lo markdown2pdf.tar.gz https://github.com/dlouwers/markdown2pdf/releases/download/v1.2.2/markdown2pdf_1.2.2_linux_amd64.tar.gz
tar xzf markdown2pdf.tar.gz
sudo mv markdown2pdf /usr/local/bin/
```

</details>

<details>
<summary><b>Windows Installation</b></summary>

Download the `.zip` file for your architecture from [releases](https://github.com/dlouwers/markdown2pdf/releases/latest), extract it, and add `markdown2pdf.exe` to your PATH.

</details>

### From Source

Requires Go 1.25 or later.

```bash
go install github.com/dlouwers/markdown2pdf/cmd/markdown2pdf@latest
```

## Features

### Document Conversion

- **Headings, paragraphs, inline styles** — bold, italic, code, links
- **Syntax-highlighted code blocks** — 200+ languages via [Chroma](https://github.com/alecthomas/chroma)
- **Smart inline code wrapping** — long code spans break at safe points (spaces, underscores, dots, slashes) with continuation indicators
- **GFM tables** — alignment, borders, header styling, and automatic font scaling for wide tables (8+ columns)
- **Lists** — ordered, unordered, and task lists with nested support
- **Images** — PNG, JPEG, SVG (local files, HTTP/HTTPS URLs, and base64 data URIs)

### Advanced Features

- **Professional cover pages** — generated from YAML frontmatter with automatic font scaling for long titles
- **Table of contents** — optional `--toc` flag creates clickable navigation links
- **Mermaid diagrams** — flowcharts, sequence diagrams, etc. (requires `mmdc`)
- **D2 diagrams** — rendered natively, no external dependencies
- **Orphan protection** — headings never appear stranded at the bottom of a page

### Typography & Fonts

- **Embedded Noto Sans font** — full UTF-8 support out of the box
- **Smart font fallback cascade** — body font → symbols font → emoji font → text substitution
- **Custom font support** — override with `--font`, `--symbols-font`, `--emoji-font`
- **Proportional table spacing** — em-based padding scales with font size for professional appearance
- **Single static binary** — no runtime dependencies except `mmdc` for Mermaid diagrams


## Usage

### Basic Examples

```bash
# Simple conversion (output: document.pdf)
markdown2pdf document.md

# Specify output path
markdown2pdf -o output.pdf document.md
```

### With Advanced Features

```bash
# Generate with table of contents
markdown2pdf --toc document.md

# Generate with cover page from frontmatter metadata
markdown2pdf --cover-page document.md

# Combine cover page with table of contents
markdown2pdf --cover-page --toc document.md
```

### Custom Fonts

```bash
# Use a custom body font (zip or tar.gz with TTF files)
markdown2pdf --font /path/to/MyFont.zip document.md

# Use a custom symbols fallback font
markdown2pdf --symbols-font /path/to/Symbols.tar.gz document.md

# Use a custom emoji fallback font
markdown2pdf --emoji-font /path/to/Emoji.tar.gz document.md
```

### Other Options

```bash
# Print version
markdown2pdf --version

# View all options
markdown2pdf --help
```


## Cover Pages

Generate professional cover pages from YAML frontmatter metadata. When the `--cover-page` flag is used, markdown2pdf extracts metadata from the frontmatter block and renders it on a dedicated first page.

**Smart font scaling**: Titles that wrap to 3+ lines automatically reduce font size (90% at 3 lines, 80% at 4+ lines) to maintain professional proportions.

### Frontmatter Format

Add a YAML block at the start of your markdown file:

```yaml
---
title: My Document Title
subtitle: An Optional Subtitle
author: Your Name
date: February 27, 2026
version: 1.0.0
---
```

### Supported Fields

| Field | Required | Description |
|-------|----------|-------------|
| `title` | Yes* | Large, bold, centered at top |
| `subtitle` | No | Italic, below title |
| `author` | No | Centered in middle section |
| `date` | No | Centered below author |
| `version` | No | Displayed as "Version X.Y.Z" |

*Required only if `--cover-page` flag is used. If title is missing, no cover page is generated.
## Fonts

markdown2pdf uses a **four-tier font cascade** to handle all characters, from basic text to complex emoji:

1. **Body font**: [Noto Sans](https://fonts.google.com/noto/specimen/Noto+Sans) (embedded, full UTF-8 coverage)
   - Override: `--font /path/to/font.zip`
   - Archive should contain TTF files: Regular, Bold, Italic, BoldItalic (auto-detected by filename)

2. **Symbols fallback**: [Noto Sans Symbols 2](https://fonts.google.com/noto/specimen/Noto+Sans+Symbols+2) (embedded)
   - Renders glyphs the body font lacks (mathematical symbols, arrows, geometric shapes)
   - Override: `--symbols-font /path/to/symbols.tar.gz`

3. **Emoji fallback**: [Noto Emoji](https://fonts.google.com/noto/specimen/Noto+Emoji) (embedded)
   - Renders pictographic symbols as black & white
   - Top 100 common emoji use **Twemoji** color graphics (PNG format)
   - Override: `--emoji-font /path/to/emoji.tar.gz`

4. **Text substitution**: Glyphs unsupported by all fonts → ASCII equivalents (e.g. 🚀 → `[>]`)

**Full cascade**: body font → symbols font → emoji font → text substitution
## Diagrams

### Mermaid

Fenced code blocks tagged with `mermaid` are rendered to PNG via [mermaid-cli](https://github.com/mermaid-js/mermaid-cli).

**Installation:**

```bash
npm install -g @mermaid-js/mermaid-cli
```

If `mmdc` is not available, Mermaid blocks render as a placeholder with the raw source.

### D2

Fenced code blocks tagged with `d2` are rendered natively using the [D2 Go library](https://github.com/terrastruct/d2) — **no external tools required**.

### Browser Detection

Diagram rendering uses a headless Chromium-based browser for rasterization. Auto-detected in order:

1. `CHROME_PATH` environment variable (if set)
2. Brave Browser
3. Google Chrome
4. Chromium
5. Microsoft Edge
## Examples

The [`testdata/`](testdata/) directory contains comprehensive examples demonstrating all features:

- **Cover pages**: [`cover-page.md`](testdata/cover-page.md), [`cover-three-line-title.md`](testdata/cover-three-line-title.md) (font scaling demo)
- **Tables**: [`tables.md`](testdata/tables.md), [`table-8col.md`](testdata/table-8col.md) (proportional spacing demo)
- **Diagrams**: [`diagrams.md`](testdata/diagrams.md) (Mermaid + D2)
- **Code blocks**: [`code_blocks.md`](testdata/code_blocks.md), [`inline-code-wrapping.md`](testdata/inline-code-wrapping.md)
- **Images**: [`images.md`](testdata/images.md)
- **Emoji**: [`emoji-color.md`](testdata/emoji-color.md)

Each example includes a generated PDF showing the output. Generate them yourself:

```bash
# Generate all examples
for f in testdata/*.md; do markdown2pdf "$f"; done
```
## Contributing

Contributions are welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for build instructions, development workflow, and release process.
## Credits

**Twemoji graphics** (top 100 emoji): Copyright 2020 Twitter, Inc and other contributors  
Licensed under [CC-BY 4.0](https://creativecommons.org/licenses/by/4.0/)

## License

MIT License — see [LICENSE](LICENSE) for details.
