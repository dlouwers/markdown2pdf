# Goal
This is to be an application that converts markdown with codeblocks to good looking pdf files. We will use the Go programming language to create them.

# Stages
1. Write a plan to PLAN.md and refer to it from this document — **✅ Done: see [PLAN.md](./PLAN.md)**
2. Implement after user verification

# Features
- builds through a devcontainer on github actions
- statically compiles to a binary without dependencies
- uses semver and releases binaries for OSX, Linux, Windows both AMD as well as ARM.
- start with a 0.0.x version until ready for release
- supports embedded images in PNG or SVG format
- supports embedded Mermaid and D2 diagrams
- assures that the generated diagrams look good and have the proper "light" look for a pdf document
- assures that the generated diagrams fit well within the pdf document

# Working Directory

**At the start of every session**, read this file to determine the project root. The default workspace provided by the environment is often wrong.

Project root: `/Users/dirk/Documents/projects/markdown2pdf`
All shell commands must use this as the working directory (pass `workdir` to the Bash tool or use `docker exec -w`).

# Go Version
This project uses **Go 1.25**. The devcontainer image is `mcr.microsoft.com/devcontainers/go:1.25` and `GOTOOLCHAIN=auto` is set so toolchain downloads happen automatically. Do not upgrade to a newer Go version without explicit approval.

# DevContainer Tooling
All build, test, lint, and PDF generation commands **must** run inside the devcontainer — never install project tooling (Chromium, mermaid-cli, golangci-lint, etc.) on the host.

## Running commands inside the devcontainer
Use `docker exec` against the running devcontainer, or use `devcontainer exec`:

```bash
# Build
devcontainer exec --workspace-folder /Users/dirk/Documents/projects/markdown2pdf -- go build ./...

# Lint
devcontainer exec --workspace-folder /Users/dirk/Documents/projects/markdown2pdf -- golangci-lint run --timeout=5m ./...

# Test
devcontainer exec --workspace-folder /Users/dirk/Documents/projects/markdown2pdf -- go test ./... -count=1

# Static build check
devcontainer exec --workspace-folder /Users/dirk/Documents/projects/markdown2pdf -- env CGO_ENABLED=0 go build -trimpath -o /dev/null ./cmd/markdown2pdf
```

## Generating PDFs for visual verification
PDF generation that involves Mermaid or D2 diagrams **requires** the devcontainer (Chromium + mermaid-cli are installed there). Always generate into the `testdata/` folder.

```bash
# Generate all test PDFs at once
devcontainer exec --workspace-folder /Users/dirk/Documents/projects/markdown2pdf -- sh -c '\
  for f in testdata/*.md; do \
    go run ./cmd/markdown2pdf "$f"; \
  done'

# Generate a single PDF
devcontainer exec --workspace-folder /Users/dirk/Documents/projects/markdown2pdf -- go run ./cmd/markdown2pdf testdata/diagrams.md
```

Output PDFs land next to their source markdown (e.g. `testdata/diagrams.md` → `testdata/diagrams.pdf`). Open them on the host for visual inspection.

## What the devcontainer provides
- **Go 1.25** with `GOTOOLCHAIN=auto`
- **golangci-lint** v2.10.1
- **Chromium** (headless, for D2 diagram rasterization via chromedp)
- **mermaid-cli** (`mmdc`, for Mermaid diagram rendering)
- **Node.js** (required by mermaid-cli)
