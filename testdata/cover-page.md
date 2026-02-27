---
title: markdown2pdf User Guide
subtitle: A Comprehensive Guide to Converting Markdown to PDF
author: Documentation Team
date: February 27, 2026
version: 1.0.0
---

# Introduction

This document demonstrates the cover page feature of markdown2pdf. When the `--cover-page` flag is used, metadata from the YAML frontmatter (the block at the top of this file between `---` markers) is rendered on a professional cover page.

## Cover Page Fields

The following fields are supported in frontmatter:

- **title**: Document title (required for cover page, large and centered)
- **subtitle**: Optional subtitle (italic, below title)
- **author**: Document author(s)
- **date**: Publication or modification date
- **version**: Document version number

## Example Usage

```bash
# Generate PDF with cover page
markdown2pdf --cover-page document.md

# Combine with table of contents
markdown2pdf --cover-page --toc document.md
```

## Document Content

All markdown features work normally after the frontmatter:

- **Bold text**, *italic text*, and `inline code`
- [Links](https://example.com)
- Lists, tables, code blocks, diagrams

> Blockquotes work as expected

### Code Example

```go
func main() {
    fmt.Println("Hello, markdown2pdf!")
}
```

### Table Example

| Feature | Supported |
|---------|-----------|
| Cover pages | ✅ |
| Frontmatter | ✅ |
| YAML metadata | ✅ |

## Conclusion

The cover page feature provides a professional first impression for your markdown documents while maintaining full compatibility with all existing markdown2pdf features.
