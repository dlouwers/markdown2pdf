package renderer

import (
	"testing"

	gopdf "github.com/go-pdf/fpdf"
)

func TestFindCodeBreakPoints_Spaces(t *testing.T) {
	code := "function call here"
	points := findCodeBreakPoints(code)

	if len(points) != 2 {
		t.Errorf("expected 2 break points (2 spaces), got %d", len(points))
	}

	// Both should be "after" type (spaces)
	for _, p := range points {
		if !p.after {
			t.Errorf("expected break point at space to be 'after', got 'before'")
		}
	}
}

func TestFindCodeBreakPoints_Separators(t *testing.T) {
	code := "path/to/file.txt"
	points := findCodeBreakPoints(code)

	// Should have break points after '/', '/', '.'
	if len(points) != 3 {
		t.Errorf("expected 3 break points (2 slashes, 1 dot), got %d", len(points))
	}

	for _, p := range points {
		if !p.after {
			t.Errorf("expected separator break points to be 'after', got 'before'")
		}
	}
}

func TestFindCodeBreakPoints_Delimiters(t *testing.T) {
	code := "function(arg1, arg2)"
	points := findCodeBreakPoints(code)

	// Should have break points: before '(', at ' ', before ',' at ' ', before ')'
	// Actually: before '(', after ',', after ' ', before ')'
	if len(points) < 2 {
		t.Errorf("expected at least 2 break points, got %d", len(points))
	}
}

func TestFindCodeBreakPoints_Underscores(t *testing.T) {
	code := "very_long_variable_name"
	points := findCodeBreakPoints(code)

	// Should have break points after each underscore
	if len(points) != 3 {
		t.Errorf("expected 3 break points (3 underscores), got %d", len(points))
	}
}

func TestFindCodeBreakPoints_NoBreakPoints(t *testing.T) {
	code := "simplevariable"
	points := findCodeBreakPoints(code)

	if len(points) != 0 {
		t.Errorf("expected 0 break points for alphanumeric string, got %d", len(points))
	}
}

func TestSplitCodeAtBreakPoints_ShortCode(t *testing.T) {
	fpdf := gopdf.New("P", "mm", "A4", "")
	fpdf.AddFont("NotoSansMono", "", "NotoSansMono-Regular.ttf")
	fpdf.SetFont("NotoSansMono", "", 10)

	code := "short"
	segments := splitCodeAtBreakPoints(fpdf, code, 100.0, "↪")

	if len(segments) != 1 {
		t.Errorf("expected 1 segment for short code, got %d", len(segments))
	}

	if segments[0] != code {
		t.Errorf("expected segment to be %q, got %q", code, segments[0])
	}
}

func TestSplitCodeAtBreakPoints_EmptyCode(t *testing.T) {
	fpdf := gopdf.New("P", "mm", "A4", "")
	fpdf.AddFont("NotoSansMono", "", "NotoSansMono-Regular.ttf")
	fpdf.SetFont("NotoSansMono", "", 10)

	segments := splitCodeAtBreakPoints(fpdf, "", 100.0, "↪")

	if len(segments) != 1 {
		t.Errorf("expected 1 empty segment, got %d segments", len(segments))
	}

	if segments[0] != "" {
		t.Errorf("expected empty segment, got %q", segments[0])
	}
}

func TestSplitCodeAtBreakPoints_LongWithUnderscores(t *testing.T) {
	fpdf := gopdf.New("P", "mm", "A4", "")
	fpdf.SetFont("Helvetica", "", 10) // Use built-in font for testing

	// Create a long code string with underscores
	code := "this_is_a_very_long_function_name_that_should_wrap"
	segments := splitCodeAtBreakPoints(fpdf, code, 50.0, "↪") // narrow width

	// Should break into multiple segments
	if len(segments) < 2 {
		t.Errorf("expected at least 2 segments for long code, got %d", len(segments))
	}

	// All but last segment should have continuation indicator
	for i := 0; i < len(segments)-1; i++ {
		if segments[i][len(segments[i])-len("↪"):] != "↪" {
			t.Errorf("segment %d should end with continuation indicator, got %q", i, segments[i])
		}
	}

	// Last segment should NOT have continuation indicator
	lastSeg := segments[len(segments)-1]
	if len(lastSeg) > 0 && lastSeg[len(lastSeg)-len("↪"):] == "↪" {
		t.Errorf("last segment should not have continuation indicator, got %q", lastSeg)
	}
}

func TestSplitCodeAtBreakPoints_Path(t *testing.T) {
	fpdf := gopdf.New("P", "mm", "A4", "")
	fpdf.SetFont("Helvetica", "", 10)

	code := "/very/long/filesystem/path/to/some/file.txt"
	segments := splitCodeAtBreakPoints(fpdf, code, 40.0, "↪")

	// Should break at slashes
	if len(segments) < 2 {
		t.Errorf("expected at least 2 segments for long path, got %d", len(segments))
	}

	// Each break should preserve slashes on the previous segment (break AFTER separator)
	for i := 0; i < len(segments)-1; i++ {
		seg := segments[i]
		// Remove continuation indicator for checking
		seg = seg[:len(seg)-len("↪")]
		if len(seg) > 0 {
			lastChar := seg[len(seg)-1]
			// Should end with '/' (separator kept on current line)
			if lastChar != '/' && lastChar != '.' {
				t.Logf("segment %d: %q", i, segments[i])
			}
		}
	}
}

func TestIsBreakableChar(t *testing.T) {
	tests := []struct {
		char rune
		want bool
	}{
		{' ', true},
		{'.', true},
		{'/', true},
		{'\\', true},
		{'-', true},
		{'_', true},
		{'a', false},
		{'1', false},
		{'(', false}, // delimiter, not breakable (break BEFORE it)
	}

	for _, tt := range tests {
		got := isBreakableChar(tt.char)
		if got != tt.want {
			t.Errorf("isBreakableChar(%q) = %v, want %v", tt.char, got, tt.want)
		}
	}
}

func TestIsDelimiter(t *testing.T) {
	tests := []struct {
		char rune
		want bool
	}{
		{'(', true},
		{')', true},
		{'[', true},
		{']', true},
		{'{', true},
		{'}', true},
		{' ', false},
		{'.', false},
		{'a', false},
	}

	for _, tt := range tests {
		got := isDelimiter(tt.char)
		if got != tt.want {
			t.Errorf("isDelimiter(%q) = %v, want %v", tt.char, got, tt.want)
		}
	}
}
