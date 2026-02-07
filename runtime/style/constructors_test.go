package style

import "testing"

// TestNewConstructors tests the new convenience constructor functions
func TestNewConstructors(t *testing.T) {
	tests := []struct {
		name     string
		style    Style
		wantFG   Color
		wantBG   Color
		wantBold bool
	}{
		{
			name:     "Foreground",
			style:    Foreground(Red),
			wantFG:   Red,
			wantBold: false,
		},
		{
			name:     "Background",
			style:    Background(Blue),
			wantBG:   Blue,
		},
		{
			name:     "Bold",
			style:    Bold(),
			wantBold: true,
		},
		{
			name:     "Italic",
			style:    Italic(),
			wantBold: false,
		},
		{
			name:     "Underline",
			style:    Underline(),
			wantBold: false,
		},
		{
			name:     "Reverse",
			style:    Reverse(),
			wantBold: false,
		},
		{
			name:     "FgBold",
			style:    FgBold(Green),
			wantFG:   Green,
			wantBold: true,
		},
		{
			name:     "FgBg",
			style:    FgBg(Yellow, Cyan),
			wantFG:   Yellow,
			wantBG:   Cyan,
		},
		{
			name:     "FgBgBold",
			style:    FgBgBold(Magenta, White),
			wantFG:   Magenta,
			wantBG:   White,
			wantBold: true,
		},
		{
			name:     "FgUnderline",
			style:    FgUnderline(Blue),
			wantFG:   Blue,
		},
		{
			name:     "FgBgUnderline",
			style:    FgBgUnderline(Red, Green),
			wantFG:   Red,
			wantBG:   Green,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.style.FG != tt.wantFG {
				t.Errorf("FG = %v, want %v", tt.style.FG, tt.wantFG)
			}
			if tt.style.BG != tt.wantBG {
				t.Errorf("BG = %v, want %v", tt.style.BG, tt.wantBG)
			}
			if tt.style.IsBold() != tt.wantBold {
				t.Errorf("IsBold() = %v, want %v", tt.style.IsBold(), tt.wantBold)
			}
		})
	}
}

// TestStylePresets tests the predefined style presets
func TestStylePresets(t *testing.T) {
	t.Run("None", func(t *testing.T) {
		if !None.IsEmpty() {
			t.Error("None should be empty")
		}
	})

	t.Run("ReverseStyle", func(t *testing.T) {
		if !ReverseStyle.IsReverse() {
			t.Error("ReverseStyle should be reversed")
		}
	})

	t.Run("BoldStyle", func(t *testing.T) {
		if !BoldStyle.IsBold() {
			t.Error("BoldStyle should be bold")
		}
	})

	t.Run("UnderlineStyle", func(t *testing.T) {
		if !UnderlineStyle.IsUnderline() {
			t.Error("UnderlineStyle should be underlined")
		}
	})
}

// TestIsEmpty tests the IsEmpty method
func TestIsEmpty(t *testing.T) {
	tests := []struct {
		name string
		s    Style
		want bool
	}{
		{"Zero value", Style{}, true},
		{"None preset", None, true},
		{"With FG", Foreground(Red), false},
		{"With BG", Background(Blue), false},
		{"Bold", Bold(), false},
		{"Reverse", Reverse(), false},
		{"FG + BG", FgBg(Red, Blue), false},
		{"FG + Bold", FgBold(Green), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.IsEmpty(); got != tt.want {
				t.Errorf("IsEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestClone tests the Clone method
func TestClone(t *testing.T) {
	original := FgBgBold(Red, Blue)
	cloned := original.Clone()

	// Verify cloned has same properties
	if cloned.FG != original.FG {
		t.Errorf("Clone() FG = %v, want %v", cloned.FG, original.FG)
	}
	if cloned.BG != original.BG {
		t.Errorf("Clone() BG = %v, want %v", cloned.BG, original.BG)
	}
	if cloned.IsBold() != original.IsBold() {
		t.Errorf("Clone() IsBold() = %v, want %v", cloned.IsBold(), original.IsBold())
	}

	// Verify modifying clone doesn't affect original
	cloned.FG = Green
	if original.FG == Green {
		t.Error("Modifying clone affected original")
	}
}

// TestNewAPIComparison compares old and new API usage
func TestNewAPIComparison(t *testing.T) {
	// Old API
	oldStyle := Style{}.Foreground(Red).Bold(true)
	oldStyle = oldStyle.Background(Blue)

	// New API
	newStyle := FgBgBold(Red, Blue)

	// Should produce same result
	if oldStyle.FG != newStyle.FG || oldStyle.BG != newStyle.BG || oldStyle.IsBold() != newStyle.IsBold() {
		t.Errorf("New API doesn't match old API: old=%v, new=%v", oldStyle, newStyle)
	}
}

// Benchmark comparison
func BenchmarkOldAPI(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Style{}.Foreground(Red).Background(Blue).Bold(true)
	}
}

func BenchmarkNewAPI(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = FgBgBold(Red, Blue)
	}
}

func BenchmarkOldAPI_Simple(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Style{}.Bold(true)
	}
}

func BenchmarkNewAPI_Simple(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Bold()
	}
}
