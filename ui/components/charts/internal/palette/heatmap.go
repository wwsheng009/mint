package palette

import (
	"strconv"

	fwtheme "github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime/style"
)

type gradientStop struct {
	ratio float64
	color fwtheme.Color
}

// HeatmapColor resolves a heatmap cell color for the given ratio and terminal color mode.
// The mapping uses a semantic gradient in truecolor mode and deterministic fallbacks in
// 256-color, 16-color, and no-color modes.
func HeatmapColor(ratio float64, mode fwtheme.ColorMode) style.Color {
	ratio = clamp01(ratio)

	switch mode {
	case fwtheme.ColorModeNone:
		return style.NoColor
	case fwtheme.ColorMode16:
		return heatmapNamedColor(ratio)
	case fwtheme.ColorMode256:
		color := heatmapGradientColor(ratio)
		r, g, b := color.RGBValue()
		return style.Color(strconv.Itoa(rgbToANSI256(r, g, b)))
	default:
		return style.Color(heatmapGradientColor(ratio).ToStyleString())
	}
}

func heatmapGradientColor(ratio float64) fwtheme.Color {
	stops := []gradientStop{
		{ratio: 0.00, color: fwtheme.GetColor("muted").Lighten(10)},
		{ratio: 0.35, color: fwtheme.GetColor("primary")},
		{ratio: 0.70, color: fwtheme.GetColor("warning")},
		{ratio: 1.00, color: fwtheme.GetColor("error")},
	}

	if ratio <= stops[0].ratio {
		return stops[0].color
	}
	for i := 1; i < len(stops); i++ {
		if ratio <= stops[i].ratio {
			start := stops[i-1]
			end := stops[i]
			segmentRatio := (ratio - start.ratio) / (end.ratio - start.ratio)
			return blendThemeColors(start.color, end.color, segmentRatio)
		}
	}
	return stops[len(stops)-1].color
}

func heatmapNamedColor(ratio float64) style.Color {
	switch {
	case ratio < 0.20:
		return style.BrightBlack
	case ratio < 0.50:
		return style.BrightCyan
	case ratio < 0.80:
		return style.BrightYellow
	default:
		return style.BrightRed
	}
}

func blendThemeColors(from, to fwtheme.Color, ratio float64) fwtheme.Color {
	ratio = clamp01(ratio)
	fromR, fromG, fromB := from.RGBValue()
	toR, toG, toB := to.RGBValue()

	blend := func(a, b int) int {
		return int(float64(a)*(1-ratio) + float64(b)*ratio)
	}

	return fwtheme.Color{
		Type: fwtheme.ColorRGB,
		Value: [3]int{
			blend(fromR, toR),
			blend(fromG, toG),
			blend(fromB, toB),
		},
	}
}

func rgbToANSI256(r, g, b int) int {
	if r == g && g == b {
		if r < 8 {
			return 16
		}
		if r > 248 {
			return 231
		}
		return 232 + ((r-8)*24)/247
	}

	quantize := func(v int) int {
		return int((float64(v)/255.0)*5.0 + 0.5)
	}

	rr := quantize(r)
	gg := quantize(g)
	bb := quantize(b)
	return 16 + 36*rr + 6*gg + bb
}

func clamp01(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}
