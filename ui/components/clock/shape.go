package clock

// DialShape controls the clock face geometry.
type DialShape int

const (
	DialShapeCircle DialShape = iota
	DialShapeEllipse
)

// DefaultCellAspectX compensates for the terminal cell aspect ratio so a
// logically circular dial appears visually round in typical monospace grids.
const DefaultCellAspectX = 2.0

// DialShapeName returns a stable human-readable name for the dial shape.
func DialShapeName(shape DialShape) string {
	switch shape {
	case DialShapeEllipse:
		return "Ellipse"
	default:
		return "Circle"
	}
}
