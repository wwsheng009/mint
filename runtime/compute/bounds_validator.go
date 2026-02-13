package compute

import (
	"fmt"

	"github.com/wwsheng009/mint/internal/log"
)

// BoundsValidator validates consistency between ComputedBox.Box and component bounds
type BoundsValidator struct{}

// NewBoundsValidator creates a new bounds validator
func NewBoundsValidator() *BoundsValidator {
	return &BoundsValidator{}
}

// ValidateComputedBox checks if a ComputedBox's bounds are consistent with its VNode's bounds
// Returns an error if inconsistency is detected, nil otherwise
func (v *BoundsValidator) ValidateComputedBox(box *ComputedBox) error {
	if !log.ValidationLogger.Enabled() || box == nil {
		return nil
	}

	if box.VNode == nil {
		return nil
	}

	// Check if VNode implements bounds inspection interface
	if boundsAware, ok := box.VNode.(interface{ GetBounds() [4]int }); ok {
		componentBounds := boundsAware.GetBounds()
		expectedBounds := [4]int{
			box.Box.X,
			box.Box.Y,
			box.Box.Width,
			box.Box.Height,
		}

		// Compare bounds
		if componentBounds != expectedBounds {
			return fmt.Errorf("bounds inconsistency detected for %T: ComputedBox.Box=%v but Component.bounds=%v",
				box.VNode,
				expectedBounds,
				componentBounds)
		}
	}

	// Recursively validate children
	for _, child := range box.Children {
		if err := v.ValidateComputedBox(child); err != nil {
			return err
		}
	}

	return nil
}

// ValidateLayout validates all bounds in a ComputedLayout
func (v *BoundsValidator) ValidateLayout(layout *ComputedLayout) error {
	if layout == nil || layout.Root == nil {
		return nil
	}
	return v.ValidateComputedBox(layout.Root)
}
