package parser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yuin/goldmark/ast"
)

func TestParseReturnsDocument(t *testing.T) {
	source := []byte("# Title\n\nParagraph text")
	node, _, _ := Parse(source)
	if node == nil {
		t.Fatal("expected non-nil node")
	}
	if node.Kind() != ast.KindDocument {
		t.Fatalf("expected document node, got %s", node.Kind())
	}
}

func TestParseBasicMarkdown(t *testing.T) {
	path := filepath.Join("..", "..", "testdata", "basic.md")
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read testdata: %v", err)
	}

	node, _, _ := Parse(source)
	if node.Kind() != ast.KindDocument {
		t.Fatalf("expected document node, got %s", node.Kind())
	}

	var (
		headings   int
		paragraphs int
		blockquote int
		thematic   int
	)

	_ = ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch n.Kind() {
		case ast.KindHeading:
			headings++
		case ast.KindParagraph:
			paragraphs++
		case ast.KindBlockquote:
			blockquote++
		case ast.KindThematicBreak:
			thematic++
		}
		return ast.WalkContinue, nil
	})

	if headings < 3 {
		t.Fatalf("expected headings, got %d", headings)
	}
	if paragraphs < 2 {
		t.Fatalf("expected paragraphs, got %d", paragraphs)
	}
	if blockquote < 1 {
		t.Fatalf("expected blockquote, got %d", blockquote)
	}
	if thematic < 1 {
		t.Fatalf("expected thematic break, got %d", thematic)
	}
}
