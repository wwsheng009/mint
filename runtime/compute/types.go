// Package compute provides constraint-driven layout engine for TUI components
package compute

import (
	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/event"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// FlexDistributionInfo caches flex distribution calculation for a parent container
// This avoids O(N²) re-measurement when calculating constraints for each flex child
type FlexDistributionInfo struct {
	TotalFlexFactor int // Sum of all flex factors
	FixedSize       int // Sum of all non-flex children sizes
	ChildCount      int // Number of children (for gap calculation)
	Valid           bool // Whether the cache is valid
}

// VNode is an alias for the VNode type from runtime/ui
type VNode = rtui.VNode

// ComputedLayout contains the computed layout information for a VNode tree
// This is the output of the new Layout Engine and input to the Paint Engine
type ComputedLayout struct {
	Root   *ComputedBox
	HitMap *event.HitMap // HitMap built from final ComputedBox positions
}

// ComputedBox represents the computed position and size of a single node
// This is NOT a VNode - it's the result of layout calculation
type ComputedBox struct {
	// VNode reference
	VNode VNode

	// Computed position and size (embedded from runtime.Box)
	runtime.Box

	// Children layout boxes (in order)
	Children []*ComputedBox

	// Parent reference (for dirty tracking)
	Parent *ComputedBox

	// Layout state
	LayoutDirty bool
	LayoutHash  uint64

	// RenderedText contains the final text to render (with padding if needed)
	// This is calculated during layout phase to avoid modifying content during paint
	RenderedText string

	// NaturalWidth stores the natural (unconstrained) width of the element
	// This is used for alignment calculations (center, end) when the element
	// is stretched to fill available space (flex layout)
	NaturalWidth int
}

// =============================================================================
// Computed Layout Construction
// =============================================================================

// NewComputedLayout creates a new ComputedLayout with a root box
func NewComputedLayout(root *ComputedBox) *ComputedLayout {
	return &ComputedLayout{Root: root}
}

// FindByPosition finds the innermost layout box containing the given position
func (cl *ComputedLayout) FindByPosition(x, y int) *ComputedBox {
	if cl.Root == nil {
		return nil
	}
	return cl.Root.FindByPosition(x, y)
}

// FindByID finds a layout box by VNode key
func (cl *ComputedLayout) FindByID(id string) *ComputedBox {
	if cl.Root == nil {
		return nil
	}
	return cl.Root.FindByID(id)
}

// FindByPosition finds the innermost layout box containing the position
func (cb *ComputedBox) FindByPosition(x, y int) *ComputedBox {
	// Check if point is in this box
	if !cb.Box.Contains(x, y) {
		return nil
	}

	// Check children (reverse order - topmost first)
	for i := len(cb.Children) - 1; i >= 0; i-- {
		if child := cb.Children[i].FindByPosition(x, y); child != nil {
			return child
		}
	}

	return cb
}

// FindByID finds a layout box by VNode key
func (cb *ComputedBox) FindByID(id string) *ComputedBox {
	if cb.VNode != nil {
		if key := cb.VNode.Key(); key == id {
			return cb
		}
	}

	for _, child := range cb.Children {
		if found := child.FindByID(id); found != nil {
			return found
		}
	}

	return nil
}

// MarkDirty marks this box and all ancestors as needing layout
func (cb *ComputedBox) MarkDirty() {
	cb.LayoutDirty = true
	for parent := cb.Parent; parent != nil; parent = parent.Parent {
		if !parent.LayoutDirty {
			parent.LayoutDirty = true
		} else {
			// Already marked, no need to continue
			break
		}
	}
}

// ClearDirty clears the dirty flag for this box and all descendants
func (cb *ComputedBox) ClearDirty() {
	cb.LayoutDirty = false
	for _, child := range cb.Children {
		child.ClearDirty()
	}
}

// Depth returns the depth of this box in the tree (root = 0)
func (cb *ComputedBox) Depth() int {
	depth := 0
	for parent := cb.Parent; parent != nil; parent = parent.Parent {
		depth++
	}
	return depth
}

// Count returns the total number of boxes in this subtree
func (cb *ComputedBox) Count() int {
	count := 1
	for _, child := range cb.Children {
		count += child.Count()
	}
	return count
}
