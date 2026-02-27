// Package parser provides Markdown parsing using goldmark.
package parser

// Metadata holds document-level information extracted from YAML frontmatter.
// All fields are optional except when a cover page is requested, which requires Title.
type Metadata struct {
	Title    string // Document title (required for cover page)
	Subtitle string // Optional subtitle
	Author   string // Document author(s)
	Date     string // Document date
	Version  string // Document version
}
