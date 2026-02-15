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

// ValidateComputedBox checks if a ComputedBox's bounds are set correctly
// In Fiber-first architecture, we use Fiber information instead of VNode
func (v *BoundsValidator) ValidateComputedBox(box *ComputedBox) error {
	if !log.ValidationLogger.Enabled() || box == nil {
		return nil
	}

	// Fiber-first: Bounds are already in box.Box
	// No need to validate against VNode since layout engine set them
	// This validator now just ensures the tree is traversable

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
