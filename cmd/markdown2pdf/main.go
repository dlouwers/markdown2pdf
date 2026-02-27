package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dlouwers/markdown2pdf/internal/parser"
	"github.com/dlouwers/markdown2pdf/internal/pdf"
	"github.com/dlouwers/markdown2pdf/internal/renderer"
)

// Set by ldflags at build time.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

// run parses flags and validates arguments. It returns an exit code.
func run(args []string) int {
	fs := flag.NewFlagSet("markdown2pdf", flag.ContinueOnError)

	var (
		output          string
		showVersion     bool
		toc             bool
		coverPage       bool
		fontPath        string
		symbolsFontPath string
		emojiFontPath   string
	)

	fs.StringVar(&output, "o", "", "output PDF file path (default: input with .pdf extension)")
	fs.BoolVar(&showVersion, "version", false, "print version information and exit")
	fs.BoolVar(&toc, "toc", false, "generate a table of contents")
	fs.BoolVar(&coverPage, "cover-page", false, "generate a cover page from frontmatter metadata")

	fs.StringVar(&fontPath, "font", "", "path to a .zip or .tar.gz archive containing TTF font files")
	fs.StringVar(&symbolsFontPath, "symbols-font", "", "path to a .zip or .tar.gz archive containing a TTF symbols fallback font")
	fs.StringVar(&emojiFontPath, "emoji-font", "", "path to a .zip or .tar.gz archive containing a TTF emoji fallback font")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	if showVersion {
		fmt.Printf("markdown2pdf %s (commit: %s, built: %s)\n", version, commit, date)
		return 0
	}

	if fs.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "error: no input file specified")
		fmt.Fprintln(os.Stderr, "usage: markdown2pdf [-o output.pdf] input.md")
		return 1
	}

	if fs.NArg() > 1 {
		fmt.Fprintln(os.Stderr, "error: too many arguments; expected exactly one input file")
		return 1
	}

	input := fs.Arg(0)

	// Validate input file exists and is a markdown file.
	info, err := os.Stat(input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	if info.IsDir() {
		fmt.Fprintf(os.Stderr, "error: %s is a directory, not a file\n", input)
		return 1
	}

	// Derive output path if not specified.
	if output == "" {
		ext := filepath.Ext(input)
		output = strings.TrimSuffix(input, ext) + ".pdf"
	}

	// Read the markdown source.
	source, err := os.ReadFile(input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: reading input: %v\n", err)
		return 1
	}

	// Parse markdown to AST and extract metadata.
	node, src, metadata := parser.Parse(source)

	// Create PDF document.
	var opts []pdf.DocumentOption
	if fontPath != "" {
		opts = append(opts, pdf.WithCustomFont(fontPath))
	}
	if symbolsFontPath != "" {
		opts = append(opts, pdf.WithCustomSymbolsFont(symbolsFontPath))
	}
	if emojiFontPath != "" {
		opts = append(opts, pdf.WithCustomEmojiFont(emojiFontPath))
	}
	doc, err := pdf.NewDocument(opts...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: creating document: %v\n", err)
		return 1
	}
	doc.SetBaseDir(filepath.Dir(input))

	// Render AST to PDF.
	r := renderer.New()
	r.TOC = toc
	r.CoverPage = coverPage
	r.Metadata = metadata
	if err := r.Render(doc, node, src); err != nil {
		fmt.Fprintf(os.Stderr, "error: rendering PDF: %v\n", err)
		return 1
	}

	// Save the PDF.
	if err := doc.Save(output); err != nil {
		fmt.Fprintf(os.Stderr, "error: saving PDF: %v\n", err)
		return 1
	}

	fmt.Printf("converted %s → %s\n", input, output)
	return 0
}
