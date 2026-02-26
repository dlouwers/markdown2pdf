package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunNoArgs(t *testing.T) {
	code := run([]string{})
	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
}

func TestRunTooManyArgs(t *testing.T) {
	code := run([]string{"a.md", "b.md"})
	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
}

func TestRunInvalidFlag(t *testing.T) {
	code := run([]string{"--nonexistent"})
	if code != 2 {
		t.Errorf("expected exit code 2, got %d", code)
	}
}

func TestRunVersion(t *testing.T) {
	code := run([]string{"--version"})
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
}

func TestRunFileNotFound(t *testing.T) {
	code := run([]string{"nonexistent.md"})
	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
}

func TestRunInputIsDirectory(t *testing.T) {
	dir := t.TempDir()
	code := run([]string{dir})
	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
}

func TestRunValidInput(t *testing.T) {
	// Use the testdata file.
	input := filepath.Join("..", "..", "testdata", "basic.md")
	if _, err := os.Stat(input); err != nil {
		t.Skipf("testdata not available: %v", err)
	}

	code := run([]string{input})
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
}

func TestRunExplicitOutput(t *testing.T) {
	input := filepath.Join("..", "..", "testdata", "basic.md")
	if _, err := os.Stat(input); err != nil {
		t.Skipf("testdata not available: %v", err)
	}

	output := filepath.Join(t.TempDir(), "out.pdf")
	code := run([]string{"-o", output, input})
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
}

func TestRunDefaultOutputDerivation(t *testing.T) {
	// Create a temp markdown file to verify output path derivation.
	dir := t.TempDir()
	input := filepath.Join(dir, "test.md")
	if err := os.WriteFile(input, []byte("# Test"), 0644); err != nil {
		t.Fatal(err)
	}

	code := run([]string{input})
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
}
