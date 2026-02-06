// Package runtime provides the single-pass layout measurement interface
package runtime

// =============================================================================
// ChildMeasurer Interface
// =============================================================================

// ChildMeasurer is an interface for measuring child nodes.
// This is defined in runtime to avoid import cycles between runtime/ui and runtime/compute.
// The Engine in runtime/compute implements this interface.
type ChildMeasurer interface {
	// MeasureChild measures a single child node with the given constraints.
	MeasureChild(child interface{}, constraints BoxConstraints) Size
}

// =============================================================================
// LayoutMeasurer Interface
// =============================================================================

// LayoutMeasurer is implemented by VNodes that want custom layout logic
// with single-pass measurement.
//
// This interface is defined in the runtime package (not compute) to avoid
// import cycles between runtime/ui and runtime/compute.
//
// The key difference from the old Measurable interface is that this
// receives a ChildMeasurer callback instead of the full Engine, breaking
// the import cycle.
type LayoutMeasurer interface {
	// IsLayoutMeasurer is a marker method for type assertion
	IsLayoutMeasurer()

	// MeasureLayout measures this node and returns layout information.
	//
	// The measurer parameter is a callback to measure children without
	// needing to import the compute package.
	MeasureLayout(
		measurer ChildMeasurer,
		constraints BoxConstraints,
	) LayoutMeasurement
}

// =============================================================================
// LayoutMeasurement
// =============================================================================

// LayoutMeasurement contains the result of measuring a layout node.
type LayoutMeasurement struct {
	// Size is the measured size of this node
	Size Size

	// ChildConstraints contains the constraints passed to each child.
	ChildConstraints []BoxConstraints

	// ChildSizes contains the measured size of each child (optional).
	ChildSizes []Size
}

// NewLayoutMeasurement creates a new LayoutMeasurement.
func NewLayoutMeasurement(size Size, childConstraints []BoxConstraints) LayoutMeasurement {
	return LayoutMeasurement{
		Size:            size,
		ChildConstraints: childConstraints,
		ChildSizes:      make([]Size, len(childConstraints)),
	}
}
