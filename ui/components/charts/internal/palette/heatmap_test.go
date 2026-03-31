package palette

import (
	"strconv"
	"strings"
	"testing"

	fwtheme "github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime/style"
)

func TestHeatmapColorModes(t *testing.T) {
	t.Run("truecolor", func(t *testing.T) {
		color := HeatmapColor(0.65, fwtheme.ColorModeTrueColor)
		if !strings.HasPrefix(string(color), "#") {
			t.Fatalf("HeatmapColor(truecolor) = %q, want hex color", color)
		}
	})

	t.Run("256", func(t *testing.T) {
		color := HeatmapColor(0.65, fwtheme.ColorMode256)
		if _, err := strconv.Atoi(string(color)); err != nil {
			t.Fatalf("HeatmapColor(256) = %q, want numeric code", color)
		}
	})

	t.Run("16", func(t *testing.T) {
		color := HeatmapColor(0.85, fwtheme.ColorMode16)
		switch color {
		case style.BrightBlack, style.BrightCyan, style.BrightYellow, style.BrightRed:
		default:
			t.Fatalf("HeatmapColor(16) = %q, want named fallback color", color)
		}
	})

	t.Run("none", func(t *testing.T) {
		color := HeatmapColor(0.85, fwtheme.ColorModeNone)
		if color != style.NoColor {
			t.Fatalf("HeatmapColor(none) = %q, want no color", color)
		}
	})
}

func TestRGBToANSI256(t *testing.T) {
	if code := rgbToANSI256(255, 0, 0); code < 16 || code > 255 {
		t.Fatalf("rgbToANSI256(red) = %d, want 16..255", code)
	}
	if code := rgbToANSI256(128, 128, 128); code < 232 || code > 255 {
		t.Fatalf("rgbToANSI256(gray) = %d, want grayscale palette code", code)
	}
}
