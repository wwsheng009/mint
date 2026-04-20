package platform

import "testing"

func TestParseGraphicsCellPixels(t *testing.T) {
	tests := []struct {
		name       string
		raw        string
		wantWidth  int
		wantHeight int
		wantErr    bool
	}{
		{name: "valid", raw: "8x16", wantWidth: 8, wantHeight: 16},
		{name: "valid uppercase separator", raw: "10X20", wantWidth: 10, wantHeight: 20},
		{name: "invalid missing separator", raw: "8", wantErr: true},
		{name: "invalid width", raw: "0x16", wantErr: true},
		{name: "invalid height", raw: "8x0", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			width, height, err := parseGraphicsCellPixels(tt.raw)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if width != tt.wantWidth || height != tt.wantHeight {
				t.Fatalf("parseGraphicsCellPixels(%q) = %dx%d, want %dx%d", tt.raw, width, height, tt.wantWidth, tt.wantHeight)
			}
		})
	}
}

func TestParseGraphicsBool(t *testing.T) {
	tests := []struct {
		raw    string
		want   bool
		wantOK bool
	}{
		{raw: "1", want: true, wantOK: true},
		{raw: "true", want: true, wantOK: true},
		{raw: "on", want: true, wantOK: true},
		{raw: "0", want: false, wantOK: true},
		{raw: "false", want: false, wantOK: true},
		{raw: "off", want: false, wantOK: true},
		{raw: "maybe", want: false, wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.raw, func(t *testing.T) {
			got, ok := parseGraphicsBool(tt.raw)
			if got != tt.want || ok != tt.wantOK {
				t.Fatalf("parseGraphicsBool(%q) = (%v, %v), want (%v, %v)", tt.raw, got, ok, tt.want, tt.wantOK)
			}
		})
	}
}

func TestReadGraphicsEnvConfig(t *testing.T) {
	t.Run("empty env", func(t *testing.T) {
		t.Setenv(graphicsModeEnvVar, "")
		t.Setenv(cellPixelsEnvVar, "")
		t.Setenv(graphicsStrictEnvVar, "")
		t.Setenv(allowTerminalFrameEnvVar, "")
		t.Setenv(allowUnverifiedInlineImageEnvVar, "")

		cfg := readGraphicsEnvConfig()
		if cfg.ModeSet || cfg.CellPixelsSet || cfg.StrictSet || cfg.AllowTerminalFrameSet || cfg.AllowUnverifiedInlineImageSet || len(cfg.Notes) != 0 {
			t.Fatalf("unexpected non-empty config: %+v", cfg)
		}
	})

	t.Run("kitty with cell pixels and strict", func(t *testing.T) {
		t.Setenv(graphicsModeEnvVar, "kitty")
		t.Setenv(cellPixelsEnvVar, "8x16")
		t.Setenv(graphicsStrictEnvVar, "1")
		t.Setenv(allowTerminalFrameEnvVar, "")
		t.Setenv(allowUnverifiedInlineImageEnvVar, "")

		cfg := readGraphicsEnvConfig()
		if !cfg.ModeSet || !cfg.ModeValid || cfg.Mode != GraphicsModeKitty {
			t.Fatalf("unexpected mode config: %+v", cfg)
		}
		if !cfg.CellPixelsSet || !cfg.CellPixelsValid || cfg.CellPixelWidth != 8 || cfg.CellPixelHeight != 16 {
			t.Fatalf("unexpected cell pixels config: %+v", cfg)
		}
		if !cfg.StrictSet || !cfg.Strict {
			t.Fatalf("unexpected strict config: %+v", cfg)
		}
	})

	t.Run("inline image mode without cell pixels", func(t *testing.T) {
		t.Setenv(graphicsModeEnvVar, "iterm2")
		t.Setenv(cellPixelsEnvVar, "")
		t.Setenv(cellPixelsLegacyEnvVar, "")
		t.Setenv(graphicsStrictEnvVar, "")
		t.Setenv(allowTerminalFrameEnvVar, "")
		t.Setenv(allowUnverifiedInlineImageEnvVar, "1")

		cfg := readGraphicsEnvConfig()
		if !cfg.ModeSet || !cfg.ModeValid || cfg.Mode != GraphicsModeInlineImage {
			t.Fatalf("unexpected inline-image mode config: %+v", cfg)
		}
		if cfg.CellPixelsSet || cfg.CellPixelsValid {
			t.Fatalf("unexpected inline-image cell pixel config: %+v", cfg)
		}
		if !cfg.AllowUnverifiedInlineImageSet || !cfg.AllowUnverifiedInlineImage {
			t.Fatalf("unexpected inline-image verification override config: %+v", cfg)
		}
	})

	t.Run("sixel with legacy cell pixels alias", func(t *testing.T) {
		t.Setenv(graphicsModeEnvVar, "sixel")
		t.Setenv(cellPixelsEnvVar, "")
		t.Setenv(cellPixelsLegacyEnvVar, "9x18")
		t.Setenv(graphicsStrictEnvVar, "")
		t.Setenv(allowTerminalFrameEnvVar, "true")
		t.Setenv(allowUnverifiedInlineImageEnvVar, "")

		cfg := readGraphicsEnvConfig()
		if !cfg.ModeSet || !cfg.ModeValid || cfg.Mode != GraphicsModeSixel {
			t.Fatalf("unexpected sixel mode config: %+v", cfg)
		}
		if !cfg.CellPixelsSet || !cfg.CellPixelsValid || cfg.CellPixelWidth != 9 || cfg.CellPixelHeight != 18 {
			t.Fatalf("unexpected legacy cell pixels config: %+v", cfg)
		}
		if !cfg.AllowTerminalFrameSet || !cfg.AllowTerminalFrame {
			t.Fatalf("unexpected terminal-frame allow config: %+v", cfg)
		}
		if len(cfg.Notes) == 0 {
			t.Fatalf("expected deprecated alias note, got %+v", cfg)
		}
	})

	t.Run("auto mode", func(t *testing.T) {
		t.Setenv(graphicsModeEnvVar, "auto")
		t.Setenv(cellPixelsEnvVar, "")
		t.Setenv(graphicsStrictEnvVar, "")
		t.Setenv(allowTerminalFrameEnvVar, "")
		t.Setenv(allowUnverifiedInlineImageEnvVar, "")

		cfg := readGraphicsEnvConfig()
		if !cfg.ModeSet || !cfg.ModeValid || !cfg.AutoMode {
			t.Fatalf("expected auto mode config, got %+v", cfg)
		}
	})

	t.Run("invalid values accumulate notes", func(t *testing.T) {
		t.Setenv(graphicsModeEnvVar, "broken")
		t.Setenv(cellPixelsEnvVar, "8")
		t.Setenv(graphicsStrictEnvVar, "maybe")
		t.Setenv(allowTerminalFrameEnvVar, "perhaps")
		t.Setenv(allowUnverifiedInlineImageEnvVar, "sometimes")

		cfg := readGraphicsEnvConfig()
		if !cfg.ModeSet || cfg.ModeValid {
			t.Fatalf("expected invalid mode config, got %+v", cfg)
		}
		if len(cfg.Notes) != 5 {
			t.Fatalf("expected 5 notes, got %d: %+v", len(cfg.Notes), cfg.Notes)
		}
	})
}
