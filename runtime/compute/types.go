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

// VNode is an alias for VNode type from runtime/ui
type VNode = rtui.VNode

// ComputedLayout contains computed layout information for a VNode tree
// This is the output of the new Layout Engine and input to Paint Engine
type ComputedLayout struct {
	Root   *ComputedBox
	HitMap *event.HitMap // HitMap built from final ComputedBox positions
	// RenderPlanes stores layer-separated layout boxes (Phase 3: RenderPlane Introduction)
	// Type: *layer.RenderPlanes (stored as interface{} to avoid import cycle)
	// Cast: planes := layout.RenderPlanes.(*layer.RenderPlanes)
	RenderPlanes interface{}
}

// ComputedBox represents the computed position and size of a single node
// This is NOT a VNode - it's the result of layout calculation
type ComputedBox struct {
	// VNode reference (DEPRECATED - will be removed in Fiber-first migration)
	// Use ChildFiber, NodeID, DiffKey instead
	VNode VNode

	// Computed position and size (embedded from runtime.Box)
	runtime.Box

	// Children layout boxes (in order)
	Children []*ComputedBox

	// Parent reference (for dirty tracking)
	Parent *ComputedBox

	// Layout state
	LayoutDirty bool
	LayoutHash uint64

	// RenderedText contains the final text to render (with padding if needed)
	// This is calculated during layout phase to avoid modifying content during paint
	RenderedText string

	// NaturalWidth stores the natural (unconstrained) width of an element
	// This is used for alignment calculations (center, end) when element
	// is stretched to fill available space (flex layout)
	NaturalWidth int

	// NodeID associates this layout box with a Fiber node
	// This provides stable runtime identity independent of VNode keys and paths
	// See: docs/render/fiber/IDENTITY_REFACTORING_PLAN.md
	NodeID uint64

	// DiffKey stores the key for dirty tracking (Fiber-first)
	// Copied from Fiber.DiffKey during layout
	// This replaces VNode.Key() for dirty tracking
	DiffKey string

	// Layer specifies rendering layer (Z-order) for this box
	// Layers: Base(0) < Overlay(1) < Modal(2) < Tooltip(3) < Inspector(4)
	// This is copied from Fiber.Layer during layout
	// See: docs/render/fiber/diff_layer.md
	Layer rtui.Layer

	// ChildFiber stores the Fiber node for this box (used for NodeID propagation to children)
	// See: docs/render/fiber/FIBER_ID.md - Option 2 implementation
	// This is set during buildComputedBox when a Fiber is provided
	ChildFiber *rtui.Fiber
}

// =============================================================================
// Computed Layout Construction
// =============================================================================

// NewComputedLayout creates a new ComputedLayout with a root box
func NewComputedLayout(root *ComputedBox) *ComputedLayout {
	return &ComputedLayout{Root: root}
}

// FindByPosition finds innermost layout box containing given position
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

// FindByPosition finds the innermost layout box containing position
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

// =============================================================================
// Fiber Access Methods (Phase 4: Fiber-First Transition)
// =============================================================================
// These methods provide Fiber-first access to ComputedBox data.
// During transition, they prioritize Fiber fields and fall back to VNode.
// This allows gradual migration without breaking existing code.

// GetFiber returns the associated Fiber node if available
func (cb *ComputedBox) GetFiber() *rtui.Fiber {
	// Try to get Fiber from ChildFiber first
	if cb.ChildFiber != nil {
		return cb.ChildFiber
	}
	// Fallback: no Fiber associated
	return nil
}

// GetVNode returns the associated VNode (deprecated during Phase 4)
// Note: This method will be removed in Phase 6 when VNode dependency is fully eliminated
func (cb *ComputedBox) GetVNode() rtui.VNode {
	return cb.VNode
}

// GetNodeType returns the VNode type if VNode is available
func (cb *ComputedBox) GetNodeType() rtui.VNodeType {
	if vnode := cb.GetVNode(); vnode != nil {
		return vnode.Type()
	}
	return 0
}

// GetNodeTag returns the tag of the associated VNode
func (cb *ComputedBox) GetNodeTag() string {
	if tagger, ok := cb.GetVNode().(interface{ Tag() string }); ok {
		return tagger.Tag()
	}
	return ""
}

// GetNodeKey returns the key of the associated VNode
func (cb *ComputedBox) GetNodeKey() string {
	if keyGetter, ok := cb.GetVNode().(interface{ Key() string }); ok {
		return keyGetter.Key()
	}
	return ""
}

// GetLayoutInfoFromFiber retrieves layout info from associated Fiber
// This is the preferred method for Fiber-first layout
func (cb *ComputedBox) GetLayoutInfoFromFiber() rtui.LayoutInfo {
	if fiber := cb.GetFiber(); fiber != nil {
		// Use Fiber's layout info (set during completeWork)
		return rtui.LayoutInfo{
			IsHorizontal: fiber.GetDirection() == rtui.DirectionRow,
			Gap:         fiber.GetGap(),
			Flex:        fiber.GetFlex(),
			Align:       fiber.GetAlign(),
			CrossAlign:   fiber.GetCrossAlign(),
			Padding:     fiber.GetPadding(),
		}
	}
	// Fallback to VNode (during transition)
	if vnode := cb.GetVNode(); vnode != nil {
		return rtui.GetLayoutInfo(vnode)
	}
	return rtui.LayoutInfo{}
}

