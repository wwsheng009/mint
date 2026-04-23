package scatterplot

// Viewport describes an optional visible window override for the scatter plot.
// Unlike Domain, viewport clips rendering to a smaller visible range.
type Viewport struct {
	MinX float64
	MaxX float64
	MinY float64
	MaxY float64
	HasX bool
	HasY bool
}

// NewViewport creates a visible window override for both axes.
func NewViewport(minX, maxX, minY, maxY float64) Viewport {
	return Viewport{
		MinX: minX,
		MaxX: maxX,
		MinY: minY,
		MaxY: maxY,
		HasX: true,
		HasY: true,
	}
}

// XViewport creates an x-axis-only viewport override.
func XViewport(minX, maxX float64) Viewport {
	return Viewport{
		MinX: minX,
		MaxX: maxX,
		HasX: true,
	}
}

// YViewport creates a y-axis-only viewport override.
func YViewport(minY, maxY float64) Viewport {
	return Viewport{
		MinY: minY,
		MaxY: maxY,
		HasY: true,
	}
}

func normalizeViewportSpec(spec Viewport) Viewport {
	if spec.HasX && spec.MinX > spec.MaxX {
		spec.MinX, spec.MaxX = spec.MaxX, spec.MinX
	}
	if spec.HasY && spec.MinY > spec.MaxY {
		spec.MinY, spec.MaxY = spec.MaxY, spec.MinY
	}
	return spec
}
