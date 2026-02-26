package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
		output      string
		showVersion bool
	)

	fs.StringVar(&output, "o", "", "output PDF file path (default: input with .pdf extension)")
	fs.BoolVar(&showVersion, "version", false, "print version information and exit")

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

	// TODO: Phase 2 — parse markdown and generate PDF.
	fmt.Printf("converting %s → %s\n", input, output)

	return 0
}
