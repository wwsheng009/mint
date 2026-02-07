package theme

import (
	"fmt"
	"strings"
)

// ColorMode represents the terminal color capability
type ColorMode int

const (
	// ColorModeTrueColor supports 24-bit RGB colors
	ColorModeTrueColor ColorMode = iota
	// ColorMode256 supports 256 colors (ANSI 256-color mode)
	ColorMode256
	// ColorMode16 supports 16 colors (basic ANSI)
	ColorMode16
	// ColorModeNone disables all colors
	ColorModeNone
)

// TerminalColorCapabilities represents terminal color support detection
type TerminalColorCapabilities struct {
	mode ColorMode
}

// NewTerminalColorCapabilities creates a new color capability detector
func NewTerminalColorCapabilities() *TerminalColorCapabilities {
	return &TerminalColorCapabilities{
		mode: detectColorMode(),
	}
}

// SetMode sets the color mode explicitly
func (t *TerminalColorCapabilities) SetMode(mode ColorMode) {
	t.mode = mode
}

// GetMode returns the current color mode
func (t *TerminalColorCapabilities) GetMode() ColorMode {
	return t.mode
}

// detectColorMode detects the terminal's color capability
// This is a simplified implementation - in production, you would check
// environment variables like COLORTERM, TERM, etc.
func detectColorMode() ColorMode {
	// Default to TrueColor for modern terminals
	// Can be enhanced with proper detection logic
	return ColorModeTrueColor
}

// ToANSI converts a ColorValue to an ANSI escape sequence based on the color mode
func (c *ColorValue) ToANSI(isBackground bool) string {
	return c.ToANSIWithMode(isBackground, ColorModeTrueColor)
}

// ToANSIWithMode converts a ColorValue to an ANSI escape sequence with specified color mode
func (c *ColorValue) ToANSIWithMode(isBackground bool, mode ColorMode) string {
	if mode == ColorModeNone {
		return ""
	}

	// Try TrueColor first (best quality)
	if mode == ColorModeTrueColor {
		return c.toTrueColorANSI(isBackground)
	}

	// Fallback to 256-color mode
	if mode == ColorMode256 {
		return c.to256ColorANSI(isBackground)
	}

	// Final fallback to 16-color mode
	return c.to16ColorANSI(isBackground)
}

// toTrueColorANSI converts to TrueColor (RGB) ANSI sequence
func (c *ColorValue) toTrueColorANSI(isBackground bool) string {
	if isBackground {
		return fmt.Sprintf("48;2;%d;%d;%d", c.RGB[0], c.RGB[1], c.RGB[2])
	}
	return fmt.Sprintf("38;2;%d;%d;%d", c.RGB[0], c.RGB[1], c.RGB[2])
}

// to256ColorANSI converts to 256-color ANSI sequence
func (c *ColorValue) to256ColorANSI(isBackground bool) string {
	code := c.ANSI256
	if isBackground {
		return fmt.Sprintf("48;5;%d", code)
	}
	return fmt.Sprintf("38;5;%d", code)
}

// to16ColorANSI converts to 16-color ANSI sequence
func (c *ColorValue) to16ColorANSI(isBackground bool) string {
	code := c.ANSI16

	// Handle bright colors (codes 8-15)
	// In 16-color mode, bright colors use 90-97 (fg) and 100-107 (bg)
	if code >= 8 && code <= 15 {
		baseCode := code - 8
		if isBackground {
			return fmt.Sprintf("%d", 100+baseCode)
		}
		return fmt.Sprintf("%d", 90+baseCode)
	}

	// Normal colors (codes 0-7)
	// Use 30-37 (fg) and 40-47 (bg)
	if isBackground {
		return fmt.Sprintf("%d", 40+code)
	}
	return fmt.Sprintf("%d", 30+code)
}

// WrapANSI wraps text with ANSI color codes
func (c *ColorValue) WrapANSI(text string, isBackground bool) string {
	ansiCode := c.ToANSI(isBackground)
	if ansiCode == "" {
		return text
	}
	return fmt.Sprintf("\x1b[%sm%s\x1b[0m", ansiCode, text)
}

// WrapANSIFg wraps text with foreground color
func (c *ColorValue) WrapANSIFg(text string) string {
	return c.WrapANSI(text, false)
}

// WrapANSIBg wraps text with background color
func (c *ColorValue) WrapANSIBg(text string) string {
	return c.WrapANSI(text, true)
}

// WrapANSICombined wraps text with both foreground and background colors
func WrapANSICombined(text string, fg, bg *ColorValue) string {
	parts := []string{}

	if fg != nil {
		parts = append(parts, fmt.Sprintf("\x1b[%sm", fg.ToANSI(false)))
	}
	if bg != nil {
		parts = append(parts, fmt.Sprintf("\x1b[%sm", bg.ToANSI(true)))
	}

	if len(parts) == 0 {
		return text
	}

	openCode := strings.Join(parts, "")
	return fmt.Sprintf("%s%s\x1b[0m", openCode, text)
}

// RGBToHex converts RGB values to hex color string
func (c *ColorValue) RGBToHex() string {
	return fmt.Sprintf("#%02x%02x%02x", c.RGB[0], c.RGB[1], c.RGB[2])
}

// HexToRGB converts hex color string to RGB values
func HexToRGB(hex string) ([3]int, error) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return [3]int{}, fmt.Errorf("invalid hex color: %s", hex)
	}

	var rgb [3]int
	for i := 0; i < 3; i++ {
		val, err := parseHexByte(hex[i*2 : i*2+2])
		if err != nil {
			return [3]int{}, err
		}
		rgb[i] = val
	}
	return rgb, nil
}

// parseHexByte parses a 2-character hex string to byte
func parseHexByte(s string) (int, error) {
	var val int
	_, err := fmt.Sscanf(s, "%02x", &val)
	return val, err
}

// Blend blends two colors with a given ratio (0.0 = color1, 1.0 = color2)
func Blend(color1, color2 *ColorValue, ratio float64) *ColorValue {
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}

	r := int(float64(color1.RGB[0])*(1-ratio) + float64(color2.RGB[0])*ratio)
	g := int(float64(color1.RGB[1])*(1-ratio) + float64(color2.RGB[1])*ratio)
	b := int(float64(color1.RGB[2])*(1-ratio) + float64(color2.RGB[2])*ratio)

	// Clamp values
	if r < 0 {
		r = 0
	}
	if r > 255 {
		r = 255
	}
	if g < 0 {
		g = 0
	}
	if g > 255 {
		g = 255
	}
	if b < 0 {
		b = 0
	}
	if b > 255 {
		b = 255
	}

	return &ColorValue{
		RGB:     [3]int{r, g, b},
		ANSI16:  color1.ANSI16,
		ANSI256: color1.ANSI256,
	}
}
