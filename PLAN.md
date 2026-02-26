# markdown2pdf — Implementation Plan

> Referenced from [AGENTS.md](./AGENTS.md) stage 1.
Status: **Phase 2 complete. Phase 3 next.**

---

## 1. Architecture Overview

```
┌─────────────┐     ┌──────────┐     ┌────────────┐     ┌───────────┐
│  CLI (cobra) │────▶│  Parser  │────▶│  Renderer  │────▶│  PDF Gen  │
│  main.go     │     │ goldmark │     │  pipeline  │     │  gofpdf   │
└─────────────┘     └──────────┘     └────────────┘     └───────────┘
                         │                 │
                         ▼                 ▼
                    ┌──────────┐     ┌────────────┐
                    │  AST     │     │  Diagram   │
                    │  walker  │     │  engines   │
                    └──────────┘     │ ┌────────┐ │
                                     │ │  D2    │ │
                                     │ │(native)│ │
                                     │ └────────┘ │
                                     │ ┌────────┐ │
                                     │ │Mermaid │ │
                                     │ │ (CLI)  │ │
                                     │ └────────┘ │
                                     └────────────┘
```

**Pipeline**: Markdown → AST → walk nodes → render each element → compose PDF

---

## 2. Technology Stack

### 2.1 Markdown Parsing — `goldmark`

- **Library**: [yuin/goldmark](https://github.com/yuin/goldmark) (CommonMark compliant, extensible, pure Go)
- **Why not blackfriday**: blackfriday v2 is in maintenance mode; goldmark is the actively maintained standard
- **Extensions**: GFM tables, strikethrough, task lists, autolinks via `goldmark-extensions`

### 2.2 PDF Generation — `gofpdf`

- **Library**: [go-pdf/fpdf](https://github.com/go-pdf/fpdf) (community fork of jung-kurt/gofpdf, actively maintained)
- **Why**: Pure Go, no CGo, static compilation works out of the box, rich feature set (images, fonts, Unicode via UTF-8 helper)
- **Also considered**:
  - `stephenafamo/goldmark-pdf` — direct goldmark→PDF renderer built on gofpdf. Promising shortcut, but we'll likely need custom rendering for diagram embedding, chroma-highlighted code blocks, and fine-grained layout control. Worth revisiting if custom rendering proves too costly.
  - `maroto` — higher-level wrapper around gofpdf, adds unnecessary abstraction for our use case
  - `chromedp` — requires a Chrome binary, violates "no runtime dependencies"
  - `go-wkhtmltopdf` — requires wkhtmltopdf binary, same issue

### 2.3 Syntax Highlighting — `chroma`

- **Library**: [alecthomas/chroma](https://github.com/alecthomas/chroma) (v2)
- **Why**: Pure Go, 200+ language lexers, multiple output formats, used by Hugo and Gitea
- **Approach**: Tokenize code blocks → render each token with appropriate color/style into PDF using gofpdf styled text

### 2.4 D2 Diagrams — native Go library

- **Library**: [oss.terrastruct.com/d2](https://github.com/terrastruct/d2)
- **Why**: D2 is written in Go and exposes its compiler as a library — we can render D2 → SVG directly in-process
- **No external dependency** — pure Go, compiles statically
- **Layout engine**: Use `dagre` (default, pure Go) — avoid `elk` which requires Java

### 2.5 Mermaid Diagrams — external CLI (mermaid-cli)

- **Reality check**: There is no pure-Go Mermaid renderer. Mermaid is a JavaScript library that requires a browser/Puppeteer to render.
- **Approach**: Shell out to `mmdc` (mermaid-cli) if available on PATH
- **Graceful degradation**: If `mmdc` is not installed, emit a placeholder box with the raw Mermaid source and a warning
- **DevContainer**: Include `mermaid-cli` in the devcontainer so CI builds always have it
- **Output**: `mmdc` renders to SVG/PNG → embed resulting image in PDF

### 2.6 SVG → PNG Rasterization

- **Problem**: gofpdf does not natively support SVG embedding. Both D2 and Mermaid produce SVG.
- **Solution options** (in order of preference):
  1. Render diagrams directly to PNG (D2 supports PNG output via its `d2renderers`; `mmdc` supports `-e png`)
  2. If SVG is needed first, use a Go SVG rasterizer like [srwiley/oksvg](https://github.com/srwiley/oksvg) + [srwiley/rasterx](https://github.com/srwiley/rasterx) (pure Go)
- **Recommendation**: **Render to PNG directly** — simpler, avoids SVG parsing edge cases

### 2.7 CLI Framework

- **Library**: Standard library `flag` package (keep it simple — single command, few flags)
- **If scope grows**: Migrate to [spf13/cobra](https://github.com/spf13/cobra)

---

## 3. Project Structure

```
markdown2pdf/
├── .devcontainer/
│   ├── devcontainer.json
│   └── Dockerfile
├── .github/
│   └── workflows/
│       ├── ci.yml              # lint + test + build on every push/PR
│       └── release.yml         # GoReleaser on tag push
├── cmd/
│   └── markdown2pdf/
│       └── main.go             # CLI entry point
├── internal/
│   ├── parser/
│   │   └── parser.go           # goldmark setup + AST production
│   ├── renderer/
│   │   ├── renderer.go         # AST walker → PDF element dispatch
│   │   ├── text.go             # headings, paragraphs, inline styles
│   │   ├── code.go             # code blocks with chroma highlighting
│   │   ├── image.go            # embedded PNG/SVG images
│   │   ├── table.go            # GFM table rendering
│   │   └── list.go             # ordered/unordered lists, task lists
│   ├── diagram/
│   │   ├── d2.go               # D2 → PNG rendering (native Go)
│   │   └── mermaid.go          # Mermaid → PNG via mmdc CLI
│   └── pdf/
│       ├── document.go         # PDF document setup, page layout, margins
│       └── style.go            # fonts, colors, spacing constants
├── testdata/                   # sample markdown files for testing
│   ├── basic.md
│   ├── code_blocks.md
│   ├── diagrams.md
│   └── images.md
├── .goreleaser.yml
├── go.mod
├── go.sum
├── AGENTS.md
├── PLAN.md
├── LICENSE
└── README.md
```

---

## 4. Implementation Phases

### Phase 1: Project Scaffolding ✅
- [x] `go mod init github.com/dlouwers/markdown2pdf` (adjust module path as needed)
- [x] Set up project directory structure
- [x] DevContainer configuration (Go 1.22+, mermaid-cli, golangci-lint)
- [x] Basic CLI that reads a markdown file path and output PDF path
- [x] CI workflow (lint + test + build)
### Phase 2: Core Markdown → PDF ✅
- [x] Goldmark parser with GFM extensions
- [x] AST walker skeleton
- [x] PDF document setup (A4, margins, page numbers)
- [x] Render: headings (H1–H6 with distinct sizes/weights)
- [x] Render: paragraphs with word-wrapping
- [x] Render: inline styles (bold, italic, code, links)
- [x] Render: horizontal rules
- [x] Render: blockquotes (indented, left-border styled)
### Phase 3: Code Blocks
- [ ] Chroma integration — tokenize by language
- [ ] Render tokens with colored monospace text in PDF
- [ ] Code block background shading
- [ ] Line numbers (optional, off by default)

### Phase 4: Lists & Tables
- [ ] Unordered lists (nested, with bullet styles)
- [ ] Ordered lists (nested, with numbering)
- [ ] Task lists (checkbox rendering)
- [ ] GFM tables with borders, header styling, column alignment

### Phase 5: Images
- [ ] Embedded PNG support (inline `![alt](path)`)
- [ ] Embedded SVG support (rasterize via oksvg → PNG → embed)
- [ ] Image scaling to fit page width with aspect ratio preserved
- [ ] Base64-encoded data URI images

### Phase 6: Diagrams
- [ ] D2 code block detection (` ```d2 `)
- [ ] D2 → PNG rendering via native Go library
- [ ] Mermaid code block detection (` ```mermaid `)
- [ ] Mermaid → PNG rendering via `mmdc` CLI
- [ ] Diagram light-theme styling (white background, dark lines — good for PDF)
- [ ] Diagram scaling to fit page width
- [ ] Graceful fallback when mmdc is unavailable

### Phase 7: Polish & Release
- [ ] Page breaks — avoid orphaned headings, split long tables
- [ ] Table of contents generation (optional flag)
- [ ] GoReleaser configuration
- [ ] Release workflow (tag → build → publish binaries)
- [ ] README with usage examples
- [ ] Integration tests with golden PDF comparison

---

## 5. DevContainer Configuration

```jsonc
// .devcontainer/devcontainer.json
{
  "name": "markdown2pdf",
  "image": "mcr.microsoft.com/devcontainers/go:1.22",
  "features": {
    "ghcr.io/devcontainers/features/node:1": {}  // needed for mermaid-cli
  },
  "postCreateCommand": "npm install -g @mermaid-js/mermaid-cli && go mod download",
  "customizations": {
    "vscode": {
      "extensions": [
        "golang.go",
        "bierner.markdown-mermaid"
      ]
    }
  }
}
```

---

## 6. CI/CD & Release Strategy

### 6.1 CI Workflow (every push/PR)
- `golangci-lint run`
- `go test ./...`
- `go build -o /dev/null ./cmd/markdown2pdf` (verify it compiles)

### 6.2 Release Workflow (on `v*` tags)
- Uses GoReleaser
- Triggered by pushing a semver tag: `git tag v0.0.1 && git push --tags`
- Cross-compiles for:
  - `darwin/amd64`, `darwin/arm64`
  - `linux/amd64`, `linux/arm64`
  - `windows/amd64` (ARM64 excluded — poor cross-compilation support)
- Static compilation: `CGO_ENABLED=0`
- Produces `.tar.gz` (macOS/Linux) and `.zip` (Windows) archives
- Attaches binaries to GitHub Release

### 6.3 GoReleaser Config

```yaml
# .goreleaser.yml
version: 2
before:
  hooks:
    - go mod tidy
builds:
  - id: markdown2pdf
    main: ./cmd/markdown2pdf
    binary: markdown2pdf
    env:
      - CGO_ENABLED=0
    goos:
      - darwin
      - linux
      - windows
    goarch:
      - amd64
      - arm64
    ignore:
      - goos: windows
        goarch: arm64
    flags:
      - -trimpath
    ldflags:
      - -s -w -X main.version={{.Version}} -X main.commit={{.Commit}} -X main.date={{.CommitDate}}
archives:
  - name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"
    format_overrides:
      - goos: windows
        format: zip
checksum:
  name_template: 'checksums.txt'
release:
  github:
    owner: dlouwers      # adjust as needed
    name: markdown2pdf
changelog:
  sort: asc
  filters:
    exclude:
      - '^docs:'
      - '^test:'
      - '^ci:'

### 6.4 Versioning
- Start at `v0.0.1`
- Increment patch for each pre-release improvement
- `v0.1.0` when core markdown → PDF is complete
- `v1.0.0` when all features (diagrams, images, tables) are stable

---

## 7. Key Design Decisions

### 7.1 Markdown → AST → PDF (not Markdown → HTML → PDF)
We render directly from the goldmark AST to PDF primitives. This avoids needing a headless browser or HTML/CSS engine, keeps the binary self-contained, and gives us full control over PDF layout.

### 7.2 Diagram rendering as a pre-processing step
Diagram code blocks are detected during AST walking. Before rendering the node, we invoke the appropriate diagram engine (D2 native / Mermaid CLI), produce a PNG, and embed it as an image. This keeps the renderer clean — it just sees images.

### 7.3 Light theme for diagrams
Both D2 and Mermaid will be configured with light/white themes:
- D2: use `d2themescatalog.NeutralDefault` (or similar light theme)
- Mermaid: pass `--theme neutral` and `--backgroundColor white` to mmdc

### 7.4 Mermaid as optional dependency
The binary itself compiles statically. Mermaid support requires `mmdc` on PATH at runtime. This is clearly documented and available in the DevContainer. If missing, we warn and render the raw source.

---

## 8. Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| gofpdf Unicode support is limited | Broken rendering for non-Latin text | Use UTF-8 helper with embedded TTF fonts (e.g., Noto Sans) |
| D2 library pulls in heavy dependencies | Bloated binary, slow compile | Accept the tradeoff — D2's Go lib is well-maintained; binary size is secondary to functionality |
| Mermaid CLI not available on user's system | Mermaid diagrams not rendered | Graceful fallback + clear error message + document installation |
| Complex tables overflow page width | Layout breaks | Auto-shrink font size or switch to landscape for wide tables |
| SVG rasterization quality | Blurry diagrams | Render at 2x resolution, scale down in PDF for retina-quality |

---

## 9. Dependencies Summary

| Dependency | Purpose | Pure Go | Static-Safe |
|------------|---------|---------|-------------|
| `yuin/goldmark` | Markdown parsing | ✅ | ✅ |
| `go-pdf/fpdf` | PDF generation | ✅ | ✅ |
| `alecthomas/chroma/v2` | Syntax highlighting | ✅ | ✅ |
| `oss.terrastruct.com/d2` | D2 diagram rendering | ✅ | ✅ |
| `srwiley/oksvg` + `rasterx` | SVG → PNG rasterization | ✅ | ✅ |
| `@mermaid-js/mermaid-cli` | Mermaid rendering | N/A (runtime CLI) | N/A (external) |

All Go dependencies are pure Go with `CGO_ENABLED=0` compatibility. The binary itself is fully static. Mermaid is the only optional external runtime dependency.

---

## Next Steps

**Awaiting your review.** Once you confirm the plan (or request changes), implementation begins at Phase 1.
