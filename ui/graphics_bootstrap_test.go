package ui

import (
	"io"
	"testing"

	runtimeplatform "github.com/wwsheng009/mint/runtime/platform"
)

func TestProbeGraphicsBootstrap_KittyEnvInstallsPresenter(t *testing.T) {
	t.Setenv("MINT_GRAPHICS", "kitty")
	t.Setenv("MINT_CELL_PIXELS", "8x16")
	t.Setenv("MINT_GRAPHICS_ALLOW_TERMINAL_FRAME", "")
	t.Setenv("MINT_GRAPHICS_ALLOW_UNVERIFIED_INLINE_IMAGE", "")
	t.Setenv("KITTY_WINDOW_ID", "")
	t.Setenv("TERM_PROGRAM", "")
	t.Setenv("TERM", "")

	caps, presenter := probeGraphicsBootstrap(io.Discard)
	if caps.Mode != runtimeplatform.GraphicsModeKitty {
		t.Fatalf("caps.Mode = %v, want kitty", caps.Mode)
	}
	if !caps.HasReliableGraphics() {
		t.Fatalf("caps.Reliable = %v, want true", caps.Reliable)
	}
	if presenter == nil {
		t.Fatal("presenter = nil, want kitty presenter")
	}
	if _, ok := presenter.(*runtimeplatform.KittyGraphicsPresenter); !ok {
		t.Fatalf("presenter type = %T, want *KittyGraphicsPresenter", presenter)
	}
}

func TestProbeGraphicsBootstrap_OffDisablesPresenter(t *testing.T) {
	t.Setenv("MINT_GRAPHICS", "off")
	t.Setenv("MINT_CELL_PIXELS", "")
	t.Setenv("MINT_GRAPHICS_ALLOW_TERMINAL_FRAME", "")
	t.Setenv("MINT_GRAPHICS_ALLOW_UNVERIFIED_INLINE_IMAGE", "")
	t.Setenv("KITTY_WINDOW_ID", "1")
	t.Setenv("TERM_PROGRAM", "kitty")
	t.Setenv("TERM", "xterm-kitty")

	caps, presenter := probeGraphicsBootstrap(io.Discard)
	if caps.Mode != runtimeplatform.GraphicsModeNone {
		t.Fatalf("caps.Mode = %v, want none", caps.Mode)
	}
	if presenter != nil {
		t.Fatalf("presenter = %T, want nil", presenter)
	}
}

func TestProbeGraphicsBootstrap_InlineImageEnvInstallsPresenter(t *testing.T) {
	t.Setenv("MINT_GRAPHICS", "inline-image")
	t.Setenv("MINT_CELL_PIXELS", "")
	t.Setenv("MINT_GRAPHICS_ALLOW_TERMINAL_FRAME", "")
	t.Setenv("MINT_GRAPHICS_ALLOW_UNVERIFIED_INLINE_IMAGE", "")
	t.Setenv("TERM_PROGRAM", "WezTerm")
	t.Setenv("WEZTERM_PANE", "pane-1")
	t.Setenv("WT_SESSION", "")

	caps, presenter := probeGraphicsBootstrap(io.Discard)
	if caps.Mode != runtimeplatform.GraphicsModeInlineImage {
		t.Fatalf("caps.Mode = %v, want inline-image", caps.Mode)
	}
	if !caps.HasReliableGraphics() {
		t.Fatalf("caps.Reliable = %v, want true", caps.Reliable)
	}
	if presenter == nil {
		t.Fatal("presenter = nil, want inline-image presenter")
	}
	if _, ok := presenter.(*runtimeplatform.InlineImageGraphicsPresenter); !ok {
		t.Fatalf("presenter type = %T, want *InlineImageGraphicsPresenter", presenter)
	}
}

func TestProbeGraphicsBootstrap_InlineImageEnvWithoutVerifiedTerminalFallsBackToNone(t *testing.T) {
	t.Setenv("MINT_GRAPHICS", "inline-image")
	t.Setenv("MINT_CELL_PIXELS", "")
	t.Setenv("MINT_GRAPHICS_ALLOW_TERMINAL_FRAME", "")
	t.Setenv("MINT_GRAPHICS_ALLOW_UNVERIFIED_INLINE_IMAGE", "")
	t.Setenv("TERM_PROGRAM", "")
	t.Setenv("WEZTERM_PANE", "")
	t.Setenv("WEZTERM_EXECUTABLE", "")
	t.Setenv("LC_TERMINAL", "")
	t.Setenv("WT_SESSION", "")

	caps, presenter := probeGraphicsBootstrap(io.Discard)
	if caps.Mode != runtimeplatform.GraphicsModeNone {
		t.Fatalf("caps.Mode = %v, want none for unverified inline-image terminal", caps.Mode)
	}
	if caps.ProbeSource != "inline-image-unverified-terminal" {
		t.Fatalf("caps.ProbeSource = %q, want inline-image-unverified-terminal", caps.ProbeSource)
	}
	if presenter != nil {
		t.Fatalf("presenter = %T, want nil for unverified inline-image terminal", presenter)
	}
}

func TestProbeGraphicsBootstrap_SixelEnvInstallsPresenter(t *testing.T) {
	t.Setenv("MINT_GRAPHICS", "sixel")
	t.Setenv("MINT_CELL_PIXELS", "8x16")
	t.Setenv("MINT_GRAPHICS_ALLOW_TERMINAL_FRAME", "1")
	t.Setenv("MINT_GRAPHICS_ALLOW_UNVERIFIED_INLINE_IMAGE", "")
	t.Setenv("WT_SESSION", "")

	caps, presenter := probeGraphicsBootstrap(io.Discard)
	if caps.Mode != runtimeplatform.GraphicsModeSixel {
		t.Fatalf("caps.Mode = %v, want sixel", caps.Mode)
	}
	if !caps.HasReliableGraphics() {
		t.Fatalf("caps.Reliable = %v, want true", caps.Reliable)
	}
	if presenter == nil {
		t.Fatal("presenter = nil, want sixel presenter")
	}
	if _, ok := presenter.(*runtimeplatform.SixelGraphicsPresenter); !ok {
		t.Fatalf("presenter type = %T, want *SixelGraphicsPresenter", presenter)
	}
}

func TestProbeGraphicsBootstrap_SixelEnvDisabledWithoutExplicitAllow(t *testing.T) {
	t.Setenv("MINT_GRAPHICS", "sixel")
	t.Setenv("MINT_CELL_PIXELS", "8x16")
	t.Setenv("MINT_GRAPHICS_ALLOW_TERMINAL_FRAME", "")
	t.Setenv("MINT_GRAPHICS_ALLOW_UNVERIFIED_INLINE_IMAGE", "")
	t.Setenv("WT_SESSION", "")

	caps, presenter := probeGraphicsBootstrap(io.Discard)
	if caps.Mode != runtimeplatform.GraphicsModeNone {
		t.Fatalf("caps.Mode = %v, want none when terminal-frame is disabled", caps.Mode)
	}
	if caps.ProbeSource != "terminal-frame-disabled" {
		t.Fatalf("caps.ProbeSource = %q, want terminal-frame-disabled", caps.ProbeSource)
	}
	if presenter != nil {
		t.Fatalf("presenter = %T, want nil when terminal-frame is disabled", presenter)
	}
}

func TestGraphicsWriters(t *testing.T) {
	if runtimeGraphicsWriter() == nil {
		t.Fatal("runtimeGraphicsWriter() returned nil")
	}
	if testGraphicsWriter() != nil {
		t.Fatal("testGraphicsWriter() should return nil")
	}
}
