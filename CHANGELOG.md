# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Comprehensive README refactoring with improved structure and examples
- CONTRIBUTING.md with detailed development workflow and release process
- CHANGELOG.md to track project history

## [1.2.2] - 2024-02-27

### Fixed
- Cover page layout to prevent subtitle overlapping metadata when titles wrap to multiple lines ([#16](https://github.com/dlouwers/markdown2pdf/issues/16))
- Cover page title and subtitle wrapping now uses dynamic positioning with proper spacing

### Changed
- Cover page test fixture updated with long title/subtitle for integration testing

### Improved
- CI workflow now includes comprehensive caching for faster builds
- CI skips for documentation-only changes
- Installation instructions updated with Homebrew and Linux support
- Code formatting with `gofmt -s` for simplified syntax
- DevContainer includes Go Report Card tools

## [1.2.1] - 2024-02-20

### Fixed
- Orphan protection for consecutive headings now works correctly

### Added
- Test fixture for orphan heading behavior verification

### Changed
- Updated tests to reflect correct orphan protection behavior

## [1.2.0] - 2024-02-18

### Added
- **Smart inline code wrapping** with safe break points (spaces, underscores, dots, slashes)
- Continuation indicators for wrapped inline code spans
- Comprehensive test suite for code wrapping edge cases
- Test fixture demonstrating inline code wrapping behavior

### Changed
- `writeCode` now uses smart wrapping with margin checks
- Parser metadata signature updated for better type safety

### Improved
- Inline code spans now break intelligently instead of overflowing page margins

## [1.1.0] - 2024-02-15

### Added
- **Cover page generation** from YAML frontmatter metadata
  - `--cover-page` command-line flag for optional cover page
  - Support for `title`, `subtitle`, `author`, `date`, `version` fields
  - Professional cover page typography with centered layout
- YAML frontmatter parsing via goldmark-meta
- Cover page rendering integrated into document flow
- Cover page style constants and typography configuration
- Test fixtures for cover page features

### Fixed
- README duplicate text removed

### Improved
- Documentation for cover page feature and frontmatter format

## [1.0.2] - 2024-02-10

### Fixed
- **Critical**: SMP (Supplementary Multilingual Plane) character panic in PDF generation
- SMP emoji handling in table cell rendering

### Changed
- Dependency updates: golang.org/x/image (go-dependencies group)
- GitHub Actions updates (github-actions group)

## [1.0.1] - 2024-02-08

### Added
- Multiline lists test fixture for visual testing

## [1.0.0] - 2024-02-05

### Added
- **Emoji support** with color rendering via Twemoji graphics
  - Top 100 common emoji rendered as color PNGs
  - Inline emoji PNG embedding with spacing for readability
  - SMP emoji substitution fallback for unsupported characters
  - Margin checking and line wrapping for inline emoji
  - Comprehensive emoji test fixture
- README badges (Go version, license, release, Go Report Card)
- Twemoji attribution and licensing documentation
- Dependabot configuration for Go modules and GitHub Actions

### Changed
- Glyph processing now skips SMP emoji without substitutions gracefully

## [0.0.4] - 2024-01-28

### Added
- **UTF-8-safe text wrapping** for international character support
- **CSS 2.1 table column distribution** algorithm for better table layouts

### Improved
- Table column width calculation follows web standards
- Text wrapping handles multi-byte characters correctly

## [0.0.3] - 2024-01-25

### Added
- **Emoji font fallback** support
- Noto Emoji embedded as third-tier font fallback
- Text substitution for glyphs unsupported by all fonts

### Changed
- Font cascade now includes: body font → symbols font → emoji font → text substitution

## [0.0.2] - 2024-01-20

### Added
- **Homebrew distribution** support via tap
- Installation instructions for macOS and Linux via Homebrew

### Changed
- GoReleaser configuration updated for Homebrew formula generation

## [0.0.1] - 2024-01-15

### Added
- **Table of contents (TOC)** functionality with `--toc` flag
- Clickable navigation links in generated TOC
- Heading collection and rendering
- Initial project structure with Markdown to PDF conversion
- Syntax highlighting via Chroma (200+ languages)
- GFM table support with alignment and borders
- Image embedding (PNG, JPEG, SVG, data URIs)
- Mermaid diagram rendering via mermaid-cli
- D2 diagram rendering via D2 Go library
- Noto Sans font embedding for UTF-8 support
- Noto Sans Symbols 2 fallback font
- Orphan protection for headings
- Static binary compilation with GoReleaser
- CI/CD with GitHub Actions
- Cross-platform releases (macOS, Linux, Windows; AMD64, ARM64)

[Unreleased]: https://github.com/dlouwers/markdown2pdf/compare/v1.2.2...HEAD
[1.2.2]: https://github.com/dlouwers/markdown2pdf/compare/v1.2.1...v1.2.2
[1.2.1]: https://github.com/dlouwers/markdown2pdf/compare/v1.2.0...v1.2.1
[1.2.0]: https://github.com/dlouwers/markdown2pdf/compare/v1.1.0...v1.2.0
[1.1.0]: https://github.com/dlouwers/markdown2pdf/compare/v1.0.2...v1.1.0
[1.0.2]: https://github.com/dlouwers/markdown2pdf/compare/v1.0.1...v1.0.2
[1.0.1]: https://github.com/dlouwers/markdown2pdf/compare/v1.0.0...v1.0.1
[1.0.0]: https://github.com/dlouwers/markdown2pdf/compare/v0.0.4...v1.0.0
[0.0.4]: https://github.com/dlouwers/markdown2pdf/compare/v0.0.3...v0.0.4
[0.0.3]: https://github.com/dlouwers/markdown2pdf/compare/v0.0.2...v0.0.3
[0.0.2]: https://github.com/dlouwers/markdown2pdf/compare/v0.0.1...v0.0.2
[0.0.1]: https://github.com/dlouwers/markdown2pdf/releases/tag/v0.0.1
