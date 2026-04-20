package platform

import (
	"strings"
	"testing"
)

func TestGraphicsModeStringAndParse(t *testing.T) {
	tests := []struct {
		name     string
		mode     GraphicsMode
		want     string
		parseRaw string
	}{
		{name: "none", mode: GraphicsModeNone, want: "none", parseRaw: "none"},
		{name: "kitty", mode: GraphicsModeKitty, want: "kitty", parseRaw: "kitty"},
		{name: "sixel", mode: GraphicsModeSixel, want: "sixel", parseRaw: "sixel"},
		{name: "inline", mode: GraphicsModeInlineImage, want: "inline-image", parseRaw: "inline-image"},
		{name: "iterm2 alias", mode: GraphicsModeInlineImage, want: "inline-image", parseRaw: "iterm2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.mode.String(); got != tt.want {
				t.Fatalf("mode.String() = %q, want %q", got, tt.want)
			}
			parsed, ok := ParseGraphicsMode(tt.parseRaw)
			if !ok || parsed != tt.mode {
				t.Fatalf("ParseGraphicsMode(%q) = (%v, %v), want (%v, true)", tt.parseRaw, parsed, ok, tt.mode)
			}
		})
	}

	if _, ok := ParseGraphicsMode("broken"); ok {
		t.Fatal("expected invalid graphics mode parse to fail")
	}
}

func TestGraphicsCapabilitiesHelpers(t *testing.T) {
	caps := GraphicsCapabilities{
		Mode:            GraphicsModeKitty,
		Reliable:        true,
		CellPixelWidth:  8,
		CellPixelHeight: 16,
	}

	if !caps.HasGraphics() {
		t.Fatal("expected HasGraphics to be true")
	}
	if !caps.HasReliableGraphics() {
		t.Fatal("expected HasReliableGraphics to be true")
	}
	if !caps.CellPixelsKnown() {
		t.Fatal("expected CellPixelsKnown to be true")
	}
	if caps.EffectivePresentationModel() != GraphicsPresentationModelOverlay {
		t.Fatalf("EffectivePresentationModel() = %v, want overlay", caps.EffectivePresentationModel())
	}
	if caps.UsesTerminalFramePresentation() {
		t.Fatal("expected kitty caps to avoid terminal-frame presentation")
	}
}

func TestGraphicsCapabilitiesPresentationModelDefaultsByMode(t *testing.T) {
	tests := []struct {
		name string
		caps GraphicsCapabilities
		want GraphicsPresentationModel
	}{
		{
			name: "kitty defaults to overlay",
			caps: GraphicsCapabilities{Mode: GraphicsModeKitty},
			want: GraphicsPresentationModelOverlay,
		},
		{
			name: "sixel defaults to terminal frame",
			caps: GraphicsCapabilities{Mode: GraphicsModeSixel},
			want: GraphicsPresentationModelTerminalFrame,
		},
		{
			name: "explicit field wins",
			caps: GraphicsCapabilities{
				Mode:              GraphicsModeKitty,
				PresentationModel: GraphicsPresentationModelTerminalFrame,
			},
			want: GraphicsPresentationModelTerminalFrame,
		},
		{
			name: "none stays unknown",
			caps: GraphicsCapabilities{Mode: GraphicsModeNone},
			want: GraphicsPresentationModelUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.caps.EffectivePresentationModel(); got != tt.want {
				t.Fatalf("EffectivePresentationModel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGraphicsCapabilitiesSummary(t *testing.T) {
	caps := GraphicsCapabilities{
		Mode:              GraphicsModeKitty,
		Reliable:          true,
		CellPixelWidth:    8,
		CellPixelHeight:   16,
		SupportsPlacement: true,
		SupportsReplace:   true,
		SupportsDelete:    true,
		ProbeSource:       "env-override",
		Notes:             []string{"first", "second"},
	}

	summary := caps.Summary()
	for _, want := range []string{
		"mode=kitty",
		"reliable=true",
		"source=env-override",
		"cell=8x16",
		"features=",
		"placement",
		"replace",
		"delete",
		"notes=first; second",
	} {
		if !strings.Contains(summary, want) {
			t.Fatalf("summary %q does not contain %q", summary, want)
		}
	}
}

func TestGraphicsCapabilitiesWithNotesCopiesSlice(t *testing.T) {
	original := GraphicsCapabilities{Notes: []string{"base"}}
	derived := original.WithNotes("extra")

	if len(original.Notes) != 1 {
		t.Fatalf("original notes mutated: %+v", original.Notes)
	}
	if len(derived.Notes) != 2 || derived.Notes[1] != "extra" {
		t.Fatalf("unexpected derived notes: %+v", derived.Notes)
	}
}
