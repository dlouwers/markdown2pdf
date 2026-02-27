// Package parser provides Markdown parsing using goldmark.
package parser

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	meta "github.com/yuin/goldmark-meta"
)

// Parse parses markdown source and returns the AST, source, and extracted metadata.
// Metadata is extracted from YAML frontmatter (--- delimited block at file start).
func Parse(source []byte) (ast.Node, []byte, *Metadata) {
	markdown := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			meta.Meta,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
	)

	reader := text.NewReader(source)
	context := parser.NewContext()
	doc := markdown.Parser().Parse(reader, parser.WithContext(context))

	// Extract metadata from context (populated by goldmark-meta extension)
	metaData := extractMetadata(context)

	return doc, source, metaData
}

// extractMetadata pulls frontmatter fields from the parser context.
// Returns nil if no frontmatter exists.
func extractMetadata(ctx parser.Context) *Metadata {
	metaData := meta.Get(ctx)
	if metaData == nil {
		return nil
	}

	m := &Metadata{}

	if title, ok := metaData["title"].(string); ok {
		m.Title = title
	}
	if subtitle, ok := metaData["subtitle"].(string); ok {
		m.Subtitle = subtitle
	}
	if author, ok := metaData["author"].(string); ok {
		m.Author = author
	}
	if date, ok := metaData["date"].(string); ok {
		m.Date = date
	}
	if version, ok := metaData["version"].(string); ok {
		m.Version = version
	}

	return m
}
