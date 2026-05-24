package platform

import (
	"os"
	"strings"
)

// ProbeGraphicsCapabilities returns the current terminal graphics capabilities
// using a conservative strategy suitable for Phase 1 roll-out.
func ProbeGraphicsCapabilities() GraphicsCapabilities {
	cfg := readGraphicsEnvConfig()

	caps, ok := probeGraphicsFromEnv(cfg)
	if !ok {
		caps = probeGraphicsHeuristics()
		caps = applyGraphicsEnvCellPixels(caps, cfg)
		caps = caps.WithNotes(cfg.Notes...)
	}

	if cfg.StrictSet && cfg.Strict && caps.HasGraphics() && !caps.Reliable {
		return GraphicsCapabilities{
			Mode:        GraphicsModeNone,
			Reliable:    true,
			ProbeSource: "strict-fallback",
			Notes: append(
				append([]string(nil), caps.Notes...),
				"graphics probe not reliable, fallback to none",
			),
		}
	}

	if caps.UsesTerminalFramePresentation() && !cfg.AllowTerminalFrame && !isExplicitGraphicsMode(cfg) && !isDefaultWindowsTerminalSixel(caps) {
		notes := append([]string(nil), caps.Notes...)
		notes = append(
			notes,
			caps.Mode.String()+" terminal-frame presentation disabled by default because redraws visibly flicker in current terminals",
			"set "+allowTerminalFrameEnvVar+"=1 to force experimental enablement",
		)
		return GraphicsCapabilities{
			Mode:        GraphicsModeNone,
			Reliable:    true,
			ProbeSource: "terminal-frame-disabled",
			Notes:       notes,
		}
	}

	return caps
}

func probeGraphicsFromEnv(cfg graphicsEnvConfig) (GraphicsCapabilities, bool) {
	if !cfg.ModeSet {
		return GraphicsCapabilities{}, false
	}

	if !cfg.ModeValid {
		return GraphicsCapabilities{
			Mode:        GraphicsModeNone,
			Reliable:    true,
			ProbeSource: "env-invalid-fallback",
			Notes:       append([]string(nil), cfg.Notes...),
		}, true
	}

	if cfg.AutoMode {
		return GraphicsCapabilities{}, false
	}

	switch cfg.Mode {
	case GraphicsModeNone:
		caps := GraphicsCapabilities{
			Mode:        GraphicsModeNone,
			Reliable:    true,
			ProbeSource: "env-override",
			Notes:       append([]string(nil), cfg.Notes...),
		}
		return applyGraphicsEnvCellPixels(caps, cfg), true
	case GraphicsModeKitty:
		caps := GraphicsCapabilities{
			Mode:              GraphicsModeKitty,
			PresentationModel: GraphicsPresentationModelOverlay,
			Reliable:          cfg.CellPixelsValid,
			SupportsPlacement: true,
			SupportsReplace:   true,
			SupportsDelete:    true,
			ProbeSource:       "env-override",
			Notes:             append([]string(nil), cfg.Notes...),
		}
		if !cfg.CellPixelsValid {
			caps = caps.WithNotes("kitty forced via env without cell pixel metrics")
		}
		return applyGraphicsEnvCellPixels(caps, cfg), true
	case GraphicsModeInlineImage:
		verifiedTerminal, terminalLabel := probeVerifiedInlineImageTerminal()
		if !verifiedTerminal && !cfg.AllowUnverifiedInlineImage {
			notes := append([]string(nil), cfg.Notes...)
			notes = append(
				notes,
				"inline-image forced via env but no WezTerm/iTerm2 terminal markers were detected",
				"run inside a real WezTerm/iTerm2 session or set "+allowUnverifiedInlineImageEnvVar+"=1 to force experimental enablement",
			)
			return GraphicsCapabilities{
				Mode:        GraphicsModeNone,
				Reliable:    true,
				ProbeSource: "inline-image-unverified-terminal",
				Notes:       notes,
			}, true
		}
		caps := GraphicsCapabilities{
			Mode:              GraphicsModeInlineImage,
			PresentationModel: GraphicsPresentationModelOverlay,
			Reliable:          true,
			SupportsPlacement: true,
			SupportsReplace:   true,
			ProbeSource:       "env-override",
			Notes:             append([]string(nil), cfg.Notes...),
		}
		if !verifiedTerminal {
			caps = caps.WithNotes(
				"inline-image forced on an unverified terminal session",
				"terminal markers were not detected; behavior is experimental",
			)
		} else {
			caps = caps.WithNotes("verified " + terminalLabel + " inline-image session detected")
		}
		return applyGraphicsEnvCellPixels(caps, cfg), true
	case GraphicsModeSixel:
		caps := GraphicsCapabilities{
			Mode:              GraphicsModeSixel,
			PresentationModel: GraphicsPresentationModelTerminalFrame,
			Reliable:          true,
			SupportsPlacement: true,
			SupportsReplace:   true,
			ProbeSource:       "env-override",
			Notes:             append([]string(nil), cfg.Notes...),
		}
		if !cfg.CellPixelsValid {
			caps = caps.WithNotes("sixel forced via env without cell pixel metrics; image size will use source pixels")
		}
		return applyGraphicsEnvCellPixels(caps, cfg), true
	default:
		return GraphicsCapabilities{
			Mode:        GraphicsModeNone,
			Reliable:    true,
			ProbeSource: "env-invalid-fallback",
			Notes: append(
				append([]string(nil), cfg.Notes...),
				"unsupported graphics mode for phase 1",
			),
		}, true
	}
}

func isExplicitGraphicsMode(cfg graphicsEnvConfig) bool {
	return cfg.ModeSet && cfg.ModeValid && !cfg.AutoMode && cfg.Mode != GraphicsModeNone
}

func isDefaultWindowsTerminalSixel(caps GraphicsCapabilities) bool {
	return caps.Mode == GraphicsModeSixel && caps.ProbeSource == "heuristic-windows-terminal"
}

func applyGraphicsEnvCellPixels(caps GraphicsCapabilities, cfg graphicsEnvConfig) GraphicsCapabilities {
	if !cfg.CellPixelsValid {
		return caps
	}

	caps.CellPixelWidth = cfg.CellPixelWidth
	caps.CellPixelHeight = cfg.CellPixelHeight
	if (caps.Mode == GraphicsModeKitty || caps.Mode == GraphicsModeSixel) && caps.CellPixelsKnown() {
		caps.Reliable = true
	}
	return caps
}

func probeGraphicsHeuristics() GraphicsCapabilities {
	term := strings.TrimSpace(strings.ToLower(os.Getenv("TERM")))
	termProgram := strings.TrimSpace(strings.ToLower(os.Getenv("TERM_PROGRAM")))
	lcTerminal := strings.TrimSpace(strings.ToLower(os.Getenv("LC_TERMINAL")))
	kittyWindowID := strings.TrimSpace(os.Getenv("KITTY_WINDOW_ID"))
	wtSession := strings.TrimSpace(os.Getenv("WT_SESSION"))
	weztermPane := strings.TrimSpace(os.Getenv("WEZTERM_PANE"))
	weztermExecutable := strings.TrimSpace(os.Getenv("WEZTERM_EXECUTABLE"))

	if looksLikeKitty(term, termProgram, kittyWindowID) {
		return GraphicsCapabilities{
			Mode:              GraphicsModeKitty,
			PresentationModel: GraphicsPresentationModelOverlay,
			Reliable:          false,
			SupportsPlacement: true,
			SupportsReplace:   true,
			SupportsDelete:    true,
			ProbeSource:       "heuristic-kitty",
			Notes:             []string{"kitty detected heuristically; cell pixel size unknown"},
		}
	}
	if looksLikeWezTerm(termProgram, weztermPane, weztermExecutable) {
		return GraphicsCapabilities{
			Mode:              GraphicsModeInlineImage,
			PresentationModel: GraphicsPresentationModelOverlay,
			Reliable:          false,
			SupportsPlacement: true,
			SupportsReplace:   true,
			ProbeSource:       "heuristic-wezterm-inline-image",
			Notes: []string{
				"wezterm detected heuristically; iTerm2 inline image protocol not auto-enabled by default",
				"set MINT_GRAPHICS=inline-image to force enable",
			},
		}
	}
	if looksLikeITerm2(termProgram, lcTerminal) {
		return GraphicsCapabilities{
			Mode:              GraphicsModeInlineImage,
			PresentationModel: GraphicsPresentationModelOverlay,
			Reliable:          false,
			SupportsPlacement: true,
			SupportsReplace:   true,
			ProbeSource:       "heuristic-iterm2-inline-image",
			Notes: []string{
				"iterm2 detected heuristically; inline image protocol not auto-enabled by default",
				"set MINT_GRAPHICS=inline-image to force enable",
			},
		}
	}
	if looksLikeWindowsTerminal(wtSession) {
		return GraphicsCapabilities{
			Mode:              GraphicsModeSixel,
			PresentationModel: GraphicsPresentationModelTerminalFrame,
			Reliable:          true,
			SupportsPlacement: true,
			SupportsReplace:   true,
			ProbeSource:       "heuristic-windows-terminal",
			Notes: []string{
				"windows terminal detected heuristically; sixel enabled by default",
				"set MINT_GRAPHICS=off to disable terminal graphics",
				"set MINT_CELL_PIXELS=<width>x<height> to improve image scaling",
			},
		}
	}

	return GraphicsCapabilities{
		Mode:        GraphicsModeNone,
		Reliable:    true,
		ProbeSource: "heuristic-none",
	}
}

func looksLikeKitty(term, termProgram, kittyWindowID string) bool {
	if kittyWindowID != "" {
		return true
	}
	if termProgram == "kitty" {
		return true
	}
	return strings.Contains(term, "kitty")
}

func looksLikeWindowsTerminal(wtSession string) bool {
	return strings.TrimSpace(wtSession) != ""
}

func looksLikeWezTerm(termProgram, weztermPane, weztermExecutable string) bool {
	if strings.TrimSpace(weztermPane) != "" || strings.TrimSpace(weztermExecutable) != "" {
		return true
	}
	return strings.EqualFold(strings.TrimSpace(termProgram), "wezterm")
}

func looksLikeITerm2(termProgram, lcTerminal string) bool {
	termProgram = strings.TrimSpace(strings.ToLower(termProgram))
	lcTerminal = strings.TrimSpace(strings.ToLower(lcTerminal))
	return termProgram == "iterm.app" || lcTerminal == "iterm2"
}

func probeVerifiedInlineImageTerminal() (bool, string) {
	termProgram := strings.TrimSpace(os.Getenv("TERM_PROGRAM"))
	lcTerminal := strings.TrimSpace(os.Getenv("LC_TERMINAL"))
	weztermPane := strings.TrimSpace(os.Getenv("WEZTERM_PANE"))
	weztermExecutable := strings.TrimSpace(os.Getenv("WEZTERM_EXECUTABLE"))

	if looksLikeWezTerm(termProgram, weztermPane, weztermExecutable) {
		return true, "wezterm"
	}
	if looksLikeITerm2(termProgram, lcTerminal) {
		return true, "iterm2"
	}
	return false, ""
}
