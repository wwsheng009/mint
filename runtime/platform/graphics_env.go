package platform

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	graphicsModeEnvVar               = "MINT_GRAPHICS"
	cellPixelsEnvVar                 = "MINT_CELL_PIXELS"
	cellPixelsLegacyEnvVar           = "MINT_GRAPHICS_CELL_PIXELS"
	graphicsStrictEnvVar             = "MINT_GRAPHICS_STRICT"
	allowTerminalFrameEnvVar         = "MINT_GRAPHICS_ALLOW_TERMINAL_FRAME"
	allowUnverifiedInlineImageEnvVar = "MINT_GRAPHICS_ALLOW_UNVERIFIED_INLINE_IMAGE"
)

type graphicsEnvConfig struct {
	Mode                          GraphicsMode
	ModeSet                       bool
	ModeValid                     bool
	AutoMode                      bool
	Strict                        bool
	StrictSet                     bool
	AllowTerminalFrame            bool
	AllowTerminalFrameSet         bool
	AllowUnverifiedInlineImage    bool
	AllowUnverifiedInlineImageSet bool
	CellPixelWidth                int
	CellPixelHeight               int
	CellPixelsSet                 bool
	CellPixelsValid               bool
	Notes                         []string
}

func readGraphicsEnvConfig() graphicsEnvConfig {
	var cfg graphicsEnvConfig

	if raw := strings.TrimSpace(strings.ToLower(os.Getenv(graphicsModeEnvVar))); raw != "" {
		cfg.ModeSet = true
		switch raw {
		case "auto":
			cfg.AutoMode = true
			cfg.ModeValid = true
		default:
			mode, ok := ParseGraphicsMode(raw)
			if ok && isPhase1GraphicsMode(mode) {
				cfg.Mode = mode
				cfg.ModeValid = true
			} else {
				cfg.Notes = append(cfg.Notes, fmt.Sprintf("invalid %s value %q, fallback to none", graphicsModeEnvVar, raw))
			}
		}
	}

	cellPixelsRaw := strings.TrimSpace(os.Getenv(cellPixelsEnvVar))
	if cellPixelsRaw == "" {
		cellPixelsRaw = strings.TrimSpace(os.Getenv(cellPixelsLegacyEnvVar))
		if cellPixelsRaw != "" {
			cfg.Notes = append(cfg.Notes, fmt.Sprintf("using deprecated %s; prefer %s", cellPixelsLegacyEnvVar, cellPixelsEnvVar))
		}
	}
	if cellPixelsRaw != "" {
		cfg.CellPixelsSet = true
		width, height, err := parseGraphicsCellPixels(cellPixelsRaw)
		if err != nil {
			cfg.Notes = append(cfg.Notes, fmt.Sprintf("invalid %s value %q: %v", cellPixelsEnvVar, cellPixelsRaw, err))
		} else {
			cfg.CellPixelWidth = width
			cfg.CellPixelHeight = height
			cfg.CellPixelsValid = true
		}
	}

	if raw := strings.TrimSpace(strings.ToLower(os.Getenv(graphicsStrictEnvVar))); raw != "" {
		enabled, ok := parseGraphicsBool(raw)
		if !ok {
			cfg.Notes = append(cfg.Notes, fmt.Sprintf("invalid %s value %q, ignoring", graphicsStrictEnvVar, raw))
		} else {
			cfg.Strict = enabled
			cfg.StrictSet = true
		}
	}

	if raw := strings.TrimSpace(strings.ToLower(os.Getenv(allowTerminalFrameEnvVar))); raw != "" {
		enabled, ok := parseGraphicsBool(raw)
		if !ok {
			cfg.Notes = append(cfg.Notes, fmt.Sprintf("invalid %s value %q, ignoring", allowTerminalFrameEnvVar, raw))
		} else {
			cfg.AllowTerminalFrame = enabled
			cfg.AllowTerminalFrameSet = true
		}
	}

	if raw := strings.TrimSpace(strings.ToLower(os.Getenv(allowUnverifiedInlineImageEnvVar))); raw != "" {
		enabled, ok := parseGraphicsBool(raw)
		if !ok {
			cfg.Notes = append(cfg.Notes, fmt.Sprintf("invalid %s value %q, ignoring", allowUnverifiedInlineImageEnvVar, raw))
		} else {
			cfg.AllowUnverifiedInlineImage = enabled
			cfg.AllowUnverifiedInlineImageSet = true
		}
	}

	return cfg
}

// GraphicsCellPixelsFromEnv returns the terminal cell pixel metrics configured
// via MINT_CELL_PIXELS (or the legacy alias) when present and valid.
func GraphicsCellPixelsFromEnv() (width, height int, ok bool) {
	cfg := readGraphicsEnvConfig()
	if !cfg.CellPixelsValid {
		return 0, 0, false
	}
	return cfg.CellPixelWidth, cfg.CellPixelHeight, true
}

func parseGraphicsCellPixels(raw string) (int, int, error) {
	normalized := strings.TrimSpace(strings.ToLower(raw))
	sep := strings.IndexRune(normalized, 'x')
	if sep <= 0 || sep >= len(normalized)-1 {
		return 0, 0, fmt.Errorf("expected <width>x<height>")
	}

	width, err := strconv.Atoi(strings.TrimSpace(normalized[:sep]))
	if err != nil || width <= 0 {
		return 0, 0, fmt.Errorf("invalid width")
	}

	height, err := strconv.Atoi(strings.TrimSpace(normalized[sep+1:]))
	if err != nil || height <= 0 {
		return 0, 0, fmt.Errorf("invalid height")
	}

	return width, height, nil
}

func parseGraphicsBool(raw string) (bool, bool) {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case "1", "true", "yes", "on":
		return true, true
	case "0", "false", "no", "off":
		return false, true
	default:
		return false, false
	}
}

func isPhase1GraphicsMode(mode GraphicsMode) bool {
	switch mode {
	case GraphicsModeNone, GraphicsModeKitty, GraphicsModeSixel, GraphicsModeInlineImage:
		return true
	default:
		return false
	}
}
