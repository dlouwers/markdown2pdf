package pdf

import "testing"

func TestSubstituteUnsupportedGlyphs_ASCII(t *testing.T) {
	// Pure ASCII should pass through unchanged.
	in := "Hello, World!"
	got := SubstituteUnsupportedGlyphs(nil, nil, in)
	if got != in {
		t.Errorf("expected %q, got %q", in, got)
	}
}

func TestSubstituteUnsupportedGlyphs_NoFontData(t *testing.T) {
	// Without font data, all known emoji should be substituted.
	tests := []struct {
		in   string
		want string
	}{
		{"✅ Yes", "[v] Yes"},
		{"❌ No", "[x] No"},
		{"⭐⭐⭐", "[*][*][*]"},
		{"⚠️ Partial", "[!] Partial"}, // ⚠ + U+FE0F variation selector
		{"⏭ Next", "[>>] Next"},
		{"No emoji here", "No emoji here"},
		{"Mixed ✅ and ❌ text", "Mixed [v] and [x] text"},
	}
	for _, tt := range tests {
		got := SubstituteUnsupportedGlyphs(nil, nil, tt.in)
		if got != tt.want {
			t.Errorf("SubstituteUnsupportedGlyphs(nil, nil, %q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestSubstituteUnsupportedGlyphs_WithFont(t *testing.T) {
	// With the embedded Noto Sans font, glyphs the font supports should
	// NOT be substituted; unsupported ones should be.
	fontData := notoSansRegular

	// Noto Sans supports '•' (bullet) but not '✅' or '❌'.
	bullet := SubstituteUnsupportedGlyphs(fontData, nil, "• item")
	if bullet != "• item" {
		t.Errorf("bullet should not be substituted with Noto Sans: got %q", bullet)
	}

	check := SubstituteUnsupportedGlyphs(fontData, nil, "✅ Yes")
	if check != "[v] Yes" {
		t.Errorf("check mark should be substituted with Noto Sans: got %q", check)
	}
}

func TestSubstituteUnsupportedGlyphs_WithSymbolsFont(t *testing.T) {
	// With the symbols font, glyphs it supports should be kept (not substituted).
	bodyFont := notoSansRegular
	symFont := notoSansSymbols2Regular

	// ⚠ (U+26A0) is not in Noto Sans body but IS in Noto Sans Symbols 2.
	got := SubstituteUnsupportedGlyphs(bodyFont, symFont, "⚠ Warning")
	if got != "⚠ Warning" {
		t.Errorf("warning sign should be kept when symbols font supports it: got %q", got)
	}

	// ★ (U+2605) is not in body but IS in symbols.
	got = SubstituteUnsupportedGlyphs(bodyFont, symFont, "★ star")
	if got != "★ star" {
		t.Errorf("black star should be kept when symbols font supports it: got %q", got)
	}

	// • (bullet) is in Noto Sans body font; should still be kept.
	got = SubstituteUnsupportedGlyphs(bodyFont, symFont, "• item")
	if got != "• item" {
		t.Errorf("bullet should not be substituted: got %q", got)
	}
}

func TestSubstituteUnsupportedGlyphs_VariationSelector(t *testing.T) {
	// U+26A0 (⚠) followed by U+FE0F should produce "[!]" with no
	// leftover variation selector character.
	got := SubstituteUnsupportedGlyphs(nil, nil, "\u26A0\uFE0F")
	if got != "[!]" {
		t.Errorf("expected %q, got %q", "[!]", got)
	}
}

func TestSegmentText_ASCII(t *testing.T) {
	segs := SegmentText(nil, nil, "hello")
	if len(segs) != 1 || segs[0].Text != "hello" || segs[0].Kind != FontKindBody {
		t.Errorf("expected single body segment 'hello', got %v", segs)
	}
}

func TestSegmentText_WithSymbolsFont(t *testing.T) {
	body := notoSansRegular
	sym := notoSansSymbols2Regular

	// "Status: ⚠ OK" — body font has "Status: ", symbols has "⚠", body has " OK"
	segs := SegmentText(body, sym, "Status: ⚠ OK")
	if len(segs) < 2 {
		t.Fatalf("expected multiple segments, got %d: %v", len(segs), segs)
	}

	// First segment should be body font with "Status: "
	if segs[0].Kind != FontKindBody {
		t.Errorf("first segment should be body, got kind %d", segs[0].Kind)
	}

	// There should be at least one symbols segment containing ⚠
	foundSymbols := false
	for _, s := range segs {
		if s.Kind == FontKindSymbols {
			foundSymbols = true
		}
	}
	if !foundSymbols {
		t.Errorf("expected at least one symbols segment for ⚠, got %v", segs)
	}
}

func TestSegmentText_FallbackSubstitution(t *testing.T) {
	// With no fonts, emoji should get text-substituted as body segments.
	segs := SegmentText(nil, nil, "✅ Yes")
	if len(segs) != 1 {
		t.Fatalf("expected 1 segment (all body fallback), got %d: %v", len(segs), segs)
	}
	if segs[0].Kind != FontKindBody || segs[0].Text != "[v] Yes" {
		t.Errorf("expected body segment '[v] Yes', got %v", segs[0])
	}
}

func TestIsASCII(t *testing.T) {
	if !isASCII("hello") {
		t.Error("expected 'hello' to be ASCII")
	}
	if isASCII("héllo") {
		t.Error("expected 'héllo' to not be ASCII")
	}
	if isASCII("✅") {
		t.Error("expected '✅' to not be ASCII")
	}
}
