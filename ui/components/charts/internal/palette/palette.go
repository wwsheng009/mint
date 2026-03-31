package palette

import (
	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime/style"
)

// SeriesColor returns the fixed semantic series color for a dataset index.
func SeriesColor(index int) style.Color {
	switch normalizeIndex(index) {
	case 0:
		return theme.Primary()
	case 1:
		return theme.Accent()
	case 2:
		return theme.Secondary()
	case 3:
		return theme.Success()
	case 4:
		return theme.Warning()
	default:
		return theme.Error()
	}
}

// AxisColor returns the semantic axis color.
func AxisColor() style.Color {
	return theme.Border()
}

// GridColor returns the semantic grid color.
func GridColor() style.Color {
	return theme.Border()
}

// ReferenceLineColor returns the semantic color for chart reference lines.
func ReferenceLineColor() style.Color {
	return theme.Warning()
}

// ReferenceBandColor returns the semantic color for chart reference bands.
func ReferenceBandColor() style.Color {
	return theme.Secondary()
}

// CollisionColor returns the semantic color for dense point collisions.
func CollisionColor() style.Color {
	return theme.Warning()
}

// LabelColor returns the semantic label color.
func LabelColor() style.Color {
	return theme.Muted()
}

// TitleColor returns the semantic title color.
func TitleColor() style.Color {
	return theme.Text()
}

// UpColor returns the semantic up-trend color.
func UpColor() style.Color {
	return theme.Success()
}

// DownColor returns the semantic down-trend color.
func DownColor() style.Color {
	return theme.Error()
}

// FlatColor returns the semantic unchanged/neutral color.
func FlatColor() style.Color {
	return theme.Secondary()
}

func normalizeIndex(index int) int {
	if index < 0 {
		return 0
	}
	return index % 6
}
