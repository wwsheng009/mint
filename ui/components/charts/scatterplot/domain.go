package scatterplot

// Domain describes an optional explicit x/y range override for the scatter plot.
type Domain struct {
	MinX float64
	MaxX float64
	MinY float64
	MaxY float64
	HasX bool
	HasY bool
}

// NewDomain creates a domain override for both axes.
func NewDomain(minX, maxX, minY, maxY float64) Domain {
	return Domain{
		MinX: minX,
		MaxX: maxX,
		MinY: minY,
		MaxY: maxY,
		HasX: true,
		HasY: true,
	}
}

// XDomain creates an x-axis-only domain override.
func XDomain(minX, maxX float64) Domain {
	return Domain{
		MinX: minX,
		MaxX: maxX,
		HasX: true,
	}
}

// YDomain creates a y-axis-only domain override.
func YDomain(minY, maxY float64) Domain {
	return Domain{
		MinY: minY,
		MaxY: maxY,
		HasY: true,
	}
}

func normalizeDomainSpec(spec Domain) Domain {
	if spec.HasX && spec.MinX > spec.MaxX {
		spec.MinX, spec.MaxX = spec.MaxX, spec.MinX
	}
	if spec.HasY && spec.MinY > spec.MaxY {
		spec.MinY, spec.MaxY = spec.MaxY, spec.MinY
	}
	return spec
}
