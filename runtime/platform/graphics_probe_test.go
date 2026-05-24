package platform

import (
	"strings"
	"testing"
)

func TestProbeGraphicsCapabilities_EnvOffOverridesHeuristics(t *testing.T) {
	t.Setenv(graphicsModeEnvVar, "off")
	t.Setenv(allowUnverifiedInlineImageEnvVar, "")
	t.Setenv("TERM", "xterm-kitty")
	t.Setenv("TERM_PROGRAM", "kitty")
	t.Setenv("KITTY_WINDOW_ID", "12")
	t.Setenv("WT_SESSION", "")

	caps := ProbeGraphicsCapabilities()
	if caps.Mode != GraphicsModeNone || !caps.Reliable || caps.ProbeSource != "env-override" {
		t.Fatalf("unexpected capabilities: %+v", caps)
	}
}

func TestProbeGraphicsCapabilities_EnvKittyWithCellPixelsIsReliable(t *testing.T) {
	t.Setenv(graphicsModeEnvVar, "kitty")
	t.Setenv(cellPixelsEnvVar, "8x16")
	t.Setenv(graphicsStrictEnvVar, "")
	t.Setenv(allowTerminalFrameEnvVar, "")
	t.Setenv(allowUnverifiedInlineImageEnvVar, "")
	t.Setenv("WT_SESSION", "")

	caps := ProbeGraphicsCapabilities()
	if caps.Mode != GraphicsModeKitty {
		t.Fatalf("mode = %v, want kitty", caps.Mode)
	}
	if !caps.Reliable || caps.CellPixelWidth != 8 || caps.CellPixelHeight != 16 {
		t.Fatalf("unexpected kitty capabilities: %+v", caps)
	}
	if !caps.SupportsPlacement || !caps.SupportsReplace || !caps.SupportsDelete {
		t.Fatalf("expected kitty placement lifecycle support, got %+v", caps)
	}
	if caps.EffectivePresentationModel() != GraphicsPresentationModelOverlay {
		t.Fatalf("kitty presentation model = %v, want overlay", caps.EffectivePresentationModel())
	}
}

func TestProbeGraphicsCapabilities_EnvInlineImageVerifiedWezTermIsReliableWithoutCellPixels(t *testing.T) {
	t.Setenv(graphicsModeEnvVar, "inline-image")
	t.Setenv(cellPixelsEnvVar, "")
	t.Setenv(graphicsStrictEnvVar, "")
	t.Setenv(allowTerminalFrameEnvVar, "")
	t.Setenv(allowUnverifiedInlineImageEnvVar, "")
	t.Setenv("WT_SESSION", "")
	t.Setenv("TERM_PROGRAM", "WezTerm")
	t.Setenv("WEZTERM_PANE", "pane-1")
	t.Setenv("WEZTERM_EXECUTABLE", "")
	t.Setenv("LC_TERMINAL", "")

	caps := ProbeGraphicsCapabilities()
	if caps.Mode != GraphicsModeInlineImage {
		t.Fatalf("mode = %v, want inline-image", caps.Mode)
	}
	if !caps.Reliable {
		t.Fatalf("expected env inline-image to be reliable: %+v", caps)
	}
	if !caps.SupportsPlacement || !caps.SupportsReplace || caps.SupportsDelete {
		t.Fatalf("unexpected inline-image capabilities: %+v", caps)
	}
	if caps.EffectivePresentationModel() != GraphicsPresentationModelOverlay {
		t.Fatalf("inline-image presentation model = %v, want overlay", caps.EffectivePresentationModel())
	}
}

func TestProbeGraphicsCapabilities_EnvInlineImageWithoutVerifiedTerminalFallsBackToNone(t *testing.T) {
	t.Setenv(graphicsModeEnvVar, "inline-image")
	t.Setenv(cellPixelsEnvVar, "")
	t.Setenv(graphicsStrictEnvVar, "")
	t.Setenv(allowTerminalFrameEnvVar, "")
	t.Setenv(allowUnverifiedInlineImageEnvVar, "")
	t.Setenv("WT_SESSION", "")
	t.Setenv("TERM_PROGRAM", "")
	t.Setenv("WEZTERM_PANE", "")
	t.Setenv("WEZTERM_EXECUTABLE", "")
	t.Setenv("LC_TERMINAL", "")

	caps := ProbeGraphicsCapabilities()
	if caps.Mode != GraphicsModeNone || !caps.Reliable {
		t.Fatalf("expected verified fallback to none, got %+v", caps)
	}
	if caps.ProbeSource != "inline-image-unverified-terminal" {
		t.Fatalf("ProbeSource = %q, want inline-image-unverified-terminal", caps.ProbeSource)
	}
	if len(caps.Notes) == 0 {
		t.Fatalf("expected inline-image verification notes, got %+v", caps)
	}
}

func TestProbeGraphicsCapabilities_EnvInlineImageAllowsExplicitUnverifiedOverride(t *testing.T) {
	t.Setenv(graphicsModeEnvVar, "inline-image")
	t.Setenv(cellPixelsEnvVar, "")
	t.Setenv(graphicsStrictEnvVar, "")
	t.Setenv(allowTerminalFrameEnvVar, "")
	t.Setenv(allowUnverifiedInlineImageEnvVar, "1")
	t.Setenv("WT_SESSION", "")
	t.Setenv("TERM_PROGRAM", "")
	t.Setenv("WEZTERM_PANE", "")
	t.Setenv("WEZTERM_EXECUTABLE", "")
	t.Setenv("LC_TERMINAL", "")

	caps := ProbeGraphicsCapabilities()
	if caps.Mode != GraphicsModeInlineImage || !caps.Reliable {
		t.Fatalf("expected explicit unverified inline-image override to enable graphics, got %+v", caps)
	}
	if caps.ProbeSource != "env-override" {
		t.Fatalf("ProbeSource = %q, want env-override", caps.ProbeSource)
	}
	if !notesContainProbeNote(caps.Notes, "unverified terminal session") {
		t.Fatalf("expected unverified inline-image note, got %+v", caps.Notes)
	}
}

func TestProbeGraphicsCapabilities_EnvSixelWithCellPixelsIsReliable(t *testing.T) {
	t.Setenv(graphicsModeEnvVar, "sixel")
	t.Setenv(cellPixelsEnvVar, "8x16")
	t.Setenv(graphicsStrictEnvVar, "")
	t.Setenv(allowTerminalFrameEnvVar, "")
	t.Setenv(allowUnverifiedInlineImageEnvVar, "")
	t.Setenv("WT_SESSION", "")

	caps := ProbeGraphicsCapabilities()
	if caps.Mode != GraphicsModeSixel {
		t.Fatalf("mode = %v, want sixel", caps.Mode)
	}
	if !caps.Reliable || caps.CellPixelWidth != 8 || caps.CellPixelHeight != 16 {
		t.Fatalf("unexpected sixel capabilities: %+v", caps)
	}
	if !caps.SupportsPlacement || !caps.SupportsReplace || caps.SupportsDelete {
		t.Fatalf("unexpected sixel capabilities: %+v", caps)
	}
	if caps.EffectivePresentationModel() != GraphicsPresentationModelTerminalFrame {
		t.Fatalf("sixel presentation model = %v, want terminal-frame", caps.EffectivePresentationModel())
	}
}

func TestProbeGraphicsCapabilities_EnvSixelWithoutCellPixelsStillForcesReliable(t *testing.T) {
	t.Setenv(graphicsModeEnvVar, "sixel")
	t.Setenv(cellPixelsEnvVar, "")
	t.Setenv(graphicsStrictEnvVar, "")
	t.Setenv(allowTerminalFrameEnvVar, "")
	t.Setenv(allowUnverifiedInlineImageEnvVar, "")
	t.Setenv("WT_SESSION", "")

	caps := ProbeGraphicsCapabilities()
	if caps.Mode != GraphicsModeSixel || !caps.Reliable {
		t.Fatalf("expected explicit sixel to force reliable graphics, got %+v", caps)
	}
	if caps.CellPixelsKnown() {
		t.Fatalf("cell pixels should remain unknown without env metrics: %+v", caps)
	}
	if !notesContainProbeNote(caps.Notes, "without cell pixel metrics") {
		t.Fatalf("expected missing cell pixel note, got %+v", caps.Notes)
	}
}

func TestProbeGraphicsCapabilities_HeuristicNoneByDefault(t *testing.T) {
	t.Setenv(graphicsModeEnvVar, "")
	t.Setenv(cellPixelsEnvVar, "")
	t.Setenv(graphicsStrictEnvVar, "")
	t.Setenv(allowTerminalFrameEnvVar, "")
	t.Setenv(allowUnverifiedInlineImageEnvVar, "")
	t.Setenv("TERM", "xterm-256color")
	t.Setenv("TERM_PROGRAM", "")
	t.Setenv("KITTY_WINDOW_ID", "")
	t.Setenv("WT_SESSION", "")

	caps := ProbeGraphicsCapabilities()
	if caps.Mode != GraphicsModeNone || !caps.Reliable || caps.ProbeSource != "heuristic-none" {
		t.Fatalf("unexpected capabilities: %+v", caps)
	}
}

func TestProbeGraphicsCapabilities_HeuristicKittyNeedsCellPixelsForReliability(t *testing.T) {
	t.Setenv(graphicsModeEnvVar, "")
	t.Setenv(cellPixelsEnvVar, "")
	t.Setenv(graphicsStrictEnvVar, "")
	t.Setenv(allowTerminalFrameEnvVar, "")
	t.Setenv(allowUnverifiedInlineImageEnvVar, "")
	t.Setenv("TERM", "xterm-kitty")
	t.Setenv("TERM_PROGRAM", "")
	t.Setenv("KITTY_WINDOW_ID", "")
	t.Setenv("WT_SESSION", "")

	caps := ProbeGraphicsCapabilities()
	if caps.Mode != GraphicsModeKitty {
		t.Fatalf("mode = %v, want kitty", caps.Mode)
	}
	if caps.Reliable {
		t.Fatalf("expected heuristic kitty without cell pixels to be unreliable: %+v", caps)
	}
}

func TestProbeGraphicsCapabilities_HeuristicWindowsTerminalSixelReliableByDefault(t *testing.T) {
	t.Setenv(graphicsModeEnvVar, "")
	t.Setenv(cellPixelsEnvVar, "")
	t.Setenv(graphicsStrictEnvVar, "")
	t.Setenv(allowTerminalFrameEnvVar, "")
	t.Setenv(allowUnverifiedInlineImageEnvVar, "")
	t.Setenv("TERM", "")
	t.Setenv("TERM_PROGRAM", "")
	t.Setenv("KITTY_WINDOW_ID", "")
	t.Setenv("WT_SESSION", "demo-session")

	caps := ProbeGraphicsCapabilities()
	if caps.Mode != GraphicsModeSixel {
		t.Fatalf("mode = %v, want sixel for windows terminal", caps.Mode)
	}
	if !caps.Reliable {
		t.Fatalf("expected windows terminal sixel to be reliable by default: %+v", caps)
	}
	if caps.ProbeSource != "heuristic-windows-terminal" {
		t.Fatalf("ProbeSource = %q, want heuristic-windows-terminal", caps.ProbeSource)
	}
	if len(caps.Notes) == 0 {
		t.Fatalf("expected windows terminal notes, got %+v", caps)
	}
}

func TestProbeGraphicsCapabilities_HeuristicWezTermInlineImageNeedsExplicitEnable(t *testing.T) {
	t.Setenv(graphicsModeEnvVar, "")
	t.Setenv(cellPixelsEnvVar, "")
	t.Setenv(graphicsStrictEnvVar, "")
	t.Setenv(allowTerminalFrameEnvVar, "")
	t.Setenv(allowUnverifiedInlineImageEnvVar, "")
	t.Setenv("TERM", "")
	t.Setenv("TERM_PROGRAM", "WezTerm")
	t.Setenv("KITTY_WINDOW_ID", "")
	t.Setenv("WT_SESSION", "")
	t.Setenv("WEZTERM_PANE", "pane-1")
	t.Setenv("WEZTERM_EXECUTABLE", "")
	t.Setenv("LC_TERMINAL", "")

	caps := ProbeGraphicsCapabilities()
	if caps.Mode != GraphicsModeInlineImage {
		t.Fatalf("mode = %v, want inline-image", caps.Mode)
	}
	if caps.Reliable {
		t.Fatalf("expected heuristic inline-image to remain opt-in: %+v", caps)
	}
	if caps.ProbeSource != "heuristic-wezterm-inline-image" {
		t.Fatalf("ProbeSource = %q, want heuristic-wezterm-inline-image", caps.ProbeSource)
	}
}

func TestProbeGraphicsCapabilities_HeuristicITerm2InlineImageNeedsExplicitEnable(t *testing.T) {
	t.Setenv(graphicsModeEnvVar, "")
	t.Setenv(cellPixelsEnvVar, "")
	t.Setenv(graphicsStrictEnvVar, "")
	t.Setenv(allowTerminalFrameEnvVar, "")
	t.Setenv(allowUnverifiedInlineImageEnvVar, "")
	t.Setenv("TERM", "")
	t.Setenv("TERM_PROGRAM", "iTerm.app")
	t.Setenv("KITTY_WINDOW_ID", "")
	t.Setenv("WT_SESSION", "")
	t.Setenv("WEZTERM_PANE", "")
	t.Setenv("WEZTERM_EXECUTABLE", "")
	t.Setenv("LC_TERMINAL", "iTerm2")

	caps := ProbeGraphicsCapabilities()
	if caps.Mode != GraphicsModeInlineImage {
		t.Fatalf("mode = %v, want inline-image", caps.Mode)
	}
	if caps.Reliable {
		t.Fatalf("expected heuristic inline-image to remain opt-in: %+v", caps)
	}
	if caps.ProbeSource != "heuristic-iterm2-inline-image" {
		t.Fatalf("ProbeSource = %q, want heuristic-iterm2-inline-image", caps.ProbeSource)
	}
}

func TestProbeGraphicsCapabilities_StrictFallback(t *testing.T) {
	t.Setenv(graphicsModeEnvVar, "")
	t.Setenv(cellPixelsEnvVar, "")
	t.Setenv(graphicsStrictEnvVar, "1")
	t.Setenv(allowTerminalFrameEnvVar, "")
	t.Setenv(allowUnverifiedInlineImageEnvVar, "")
	t.Setenv("TERM", "xterm-kitty")
	t.Setenv("TERM_PROGRAM", "")
	t.Setenv("KITTY_WINDOW_ID", "")
	t.Setenv("WT_SESSION", "")

	caps := ProbeGraphicsCapabilities()
	if caps.Mode != GraphicsModeNone || caps.ProbeSource != "strict-fallback" {
		t.Fatalf("unexpected strict fallback capabilities: %+v", caps)
	}
}

func TestProbeGraphicsCapabilities_InvalidEnvFallsBackToNone(t *testing.T) {
	t.Setenv(graphicsModeEnvVar, "broken")
	t.Setenv(cellPixelsEnvVar, "")
	t.Setenv(graphicsStrictEnvVar, "")
	t.Setenv(allowTerminalFrameEnvVar, "")
	t.Setenv(allowUnverifiedInlineImageEnvVar, "")
	t.Setenv("WT_SESSION", "")

	caps := ProbeGraphicsCapabilities()
	if caps.Mode != GraphicsModeNone || caps.ProbeSource != "env-invalid-fallback" || !caps.Reliable {
		t.Fatalf("unexpected capabilities: %+v", caps)
	}
	if len(caps.Notes) == 0 {
		t.Fatalf("expected invalid env note, got %+v", caps)
	}
}

func TestProbeGraphicsCapabilities_EnvCellPixelsUpgradeHeuristicKitty(t *testing.T) {
	t.Setenv(graphicsModeEnvVar, "")
	t.Setenv(cellPixelsEnvVar, "9x18")
	t.Setenv(graphicsStrictEnvVar, "")
	t.Setenv(allowTerminalFrameEnvVar, "")
	t.Setenv(allowUnverifiedInlineImageEnvVar, "")
	t.Setenv("TERM", "xterm-kitty")
	t.Setenv("TERM_PROGRAM", "")
	t.Setenv("KITTY_WINDOW_ID", "")
	t.Setenv("WT_SESSION", "")

	caps := ProbeGraphicsCapabilities()
	if caps.Mode != GraphicsModeKitty || !caps.Reliable || caps.CellPixelWidth != 9 || caps.CellPixelHeight != 18 {
		t.Fatalf("unexpected capabilities: %+v", caps)
	}
}

func TestProbeGraphicsCapabilities_AutoWindowsTerminalDefaultsToSixel(t *testing.T) {
	t.Setenv(graphicsModeEnvVar, "auto")
	t.Setenv(cellPixelsEnvVar, "8x16")
	t.Setenv(graphicsStrictEnvVar, "")
	t.Setenv(allowTerminalFrameEnvVar, "")
	t.Setenv(allowUnverifiedInlineImageEnvVar, "")
	t.Setenv("WT_SESSION", "demo-session")

	caps := ProbeGraphicsCapabilities()
	if caps.Mode != GraphicsModeSixel || caps.ProbeSource != "heuristic-windows-terminal" || !caps.Reliable {
		t.Fatalf("unexpected capabilities: %+v", caps)
	}
	if caps.CellPixelWidth != 8 || caps.CellPixelHeight != 16 {
		t.Fatalf("expected cell pixel env to apply, got %+v", caps)
	}
}

func notesContainProbeNote(notes []string, needle string) bool {
	for _, note := range notes {
		if strings.Contains(note, needle) {
			return true
		}
	}
	return false
}
