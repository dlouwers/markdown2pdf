# Contributing to markdown2pdf

Thank you for your interest in contributing! This document covers the development workflow, building, testing, and release process.

## Table of Contents

- [Development Setup](#development-setup)
- [Building](#building)
- [Testing](#testing)
- [Code Quality](#code-quality)
- [Submitting Changes](#submitting-changes)
- [Releasing](#releasing)

## Development Setup

### Prerequisites

- **Go 1.25 or later** ([download](https://go.dev/dl/))
- **Docker** (for DevContainer development)
- **Node.js** (for Mermaid diagram support via `mmdc`)

### DevContainer (Recommended)

The project includes a DevContainer with all required tooling pre-installed:

- Go 1.25 with `GOTOOLCHAIN=auto`
- golangci-lint v2.10.1
- Chromium (headless, for D2 diagram rasterization)
- mermaid-cli (`mmdc`)
- Node.js

**Using VS Code:**

1. Install the [Dev Containers extension](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers)
2. Open the project in VS Code
3. Click "Reopen in Container" when prompted (or use Command Palette: `Dev Containers: Reopen in Container`)

**Using `devcontainer` CLI:**

```bash
# Start the container
devcontainer up --workspace-folder .

# Run commands inside the container
devcontainer exec --workspace-folder . -- go build ./...
```

### Local Development (Without DevContainer)

If you prefer local development:

1. Install Go 1.25+
2. Install golangci-lint: `go install github.com/golangci/golangci-lint/cmd/golangci-lint@v2.10.1`
3. Install mermaid-cli (optional, for Mermaid diagram tests): `npm install -g @mermaid-js/mermaid-cli`
4. Install Chromium or Chrome (for D2 diagram rasterization)

## Building

### Standard Build

```bash
go build -o markdown2pdf ./cmd/markdown2pdf
```

### Static Binary (No CGO Dependencies)

For distribution or deployment:

```bash
CGO_ENABLED=0 go build -trimpath -o markdown2pdf ./cmd/markdown2pdf
```

This produces a fully static binary with no external dependencies (except `mmdc` for Mermaid, which is optional).

### Cross-Platform Builds

GoReleaser handles cross-platform builds automatically during releases. To build manually:

```bash
# macOS ARM64
GOOS=darwin GOARCH=arm64 go build -o markdown2pdf-darwin-arm64 ./cmd/markdown2pdf

# Linux AMD64
GOOS=linux GOARCH=amd64 go build -o markdown2pdf-linux-amd64 ./cmd/markdown2pdf

# Windows AMD64
GOOS=windows GOARCH=amd64 go build -o markdown2pdf-windows-amd64.exe ./cmd/markdown2pdf
```

## Testing

### Run All Tests

```bash
go test ./... -count=1
```

The `-count=1` flag disables test caching to ensure fresh results.

### Run Tests with Coverage

```bash
go test ./... -cover -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Generate Visual Test PDFs

The `testdata/` directory contains example markdown files. Generate PDFs to verify rendering:

```bash
# Generate all test PDFs
for f in testdata/*.md; do
  go run ./cmd/markdown2pdf "$f"
done

# Generate a specific example
go run ./cmd/markdown2pdf testdata/diagrams.md
```

Output PDFs are created next to the source markdown (e.g., `testdata/diagrams.md` → `testdata/diagrams.pdf`). Open them to visually verify rendering quality.

### DevContainer Testing

If using the DevContainer, run all commands inside the container:

```bash
devcontainer exec --workspace-folder . -- go test ./... -count=1

# Generate test PDFs
devcontainer exec --workspace-folder . -- sh -c 'for f in testdata/*.md; do go run ./cmd/markdown2pdf "$f"; done'
```

## Code Quality

### Linting

Run golangci-lint before submitting changes:

```bash
golangci-lint run --timeout=5m ./...
```

**In DevContainer:**

```bash
devcontainer exec --workspace-folder . -- golangci-lint run --timeout=5m ./...
```

### Formatting

```bash
# Format all Go files
gofmt -w -s .

# Check formatting without modifying files
gofmt -l .
```

### Pre-Commit Checklist

Before committing:

- [ ] All tests pass: `go test ./... -count=1`
- [ ] Linter passes: `golangci-lint run --timeout=5m ./...`
- [ ] Code is formatted: `gofmt -w -s .`
- [ ] Visual verification: Generate test PDFs and check rendering
- [ ] Documentation updated (README, code comments) if adding features

## Submitting Changes

### Pull Request Process

1. **Fork the repository** and create a feature branch:
   ```bash
   git checkout -b feature/my-new-feature
   ```

2. **Make your changes** following the code quality guidelines above

3. **Write or update tests** for your changes

4. **Update documentation**:
   - Add or update examples in `testdata/` if relevant
   - Update README.md if adding user-facing features
   - Add inline code comments for complex logic

5. **Commit with clear messages**:
   ```bash
   git commit -m "Add feature: smart table column width calculation"
   ```

6. **Push to your fork** and open a Pull Request:
   ```bash
   git push origin feature/my-new-feature
   ```

### Commit Message Guidelines

- Use imperative mood: "Add feature" not "Added feature"
- Keep first line under 72 characters
- Reference issues: "Fix #123: handle empty frontmatter gracefully"
- For multiple changes, use atomic commits (one logical change per commit)

### Code Review

All submissions require review. We use GitHub pull requests for this purpose. Reviewers will check:

- Code quality and style
- Test coverage
- Documentation completeness
- Backward compatibility

## Releasing

**Note**: Only project maintainers can create releases.

Releases are automated with [GoReleaser](https://goreleaser.com/). The process:

1. **Ensure main branch is clean** and all CI checks pass

2. **Create a semver tag**:
   ```bash
   git tag v1.3.0
   ```

3. **Push the tag** to trigger the release workflow:
   ```bash
   git push origin v1.3.0
   ```

4. **GitHub Actions** automatically:
   - Builds binaries for macOS (amd64, arm64), Linux (amd64, arm64), Windows (amd64)
   - Creates a GitHub Release with changelog
   - Uploads binaries as release assets
   - Updates Homebrew tap (if configured)

### Release Workflow

The `.github/workflows/release.yml` workflow handles everything. GoReleaser configuration is in `.goreleaser.yml`.

### Versioning

We follow [Semantic Versioning](https://semver.org/):

- **Major** (v2.0.0): Breaking changes
- **Minor** (v1.3.0): New features, backward compatible
- **Patch** (v1.2.1): Bug fixes, backward compatible

### Changelog

Update `CHANGELOG.md` before tagging a release:

```markdown
## [1.3.0] - 2026-03-01

### Added
- Smart table column width calculation
- Support for custom CSS stylesheets

### Fixed
- Cover page layout with very long titles
```

Follow [Keep a Changelog](https://keepachangelog.com/) format.

---

## Questions?

Open an issue for questions or feature discussions. We're happy to help!
