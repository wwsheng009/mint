package reconciler

import (
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// TreeWalker provides a unified abstraction for traversing Fiber trees.
// It handles ComponentVNode skipping internally and provides a clean API
// for common operations like collecting focusable elements.
type TreeWalker struct {
	root *Fiber
}

// NewTreeWalker creates a new TreeWalker for the given Fiber tree.
func NewTreeWalker(root *Fiber) *TreeWalker {
	return &TreeWalker{root: root}
}

// SetRoot sets a new root for this TreeWalker.
func (w *TreeWalker) SetRoot(root *Fiber) {
	w.root = root
}

// CollectFocusable collects all focusable VNodes from the Fiber tree in order.
// This replaces the recursive collectFocusableFromFiber function.
// Optimized to pre-allocate slice capacity based on tree count.
func (w *TreeWalker) CollectFocusable() []rtui.FocusableVNode {
	// First pass: count total fibers to pre-allocate
	count := 0
	w.walk(w.root, func(fiber *Fiber) bool {
		if fiber != nil {
			count++
		}
		return true
	})

	// Pre-allocate with estimated capacity (focusable nodes are usually a subset)
	result := make([]rtui.FocusableVNode, 0, count/4+4)

	// Second pass: collect focusable nodes
	w.walk(w.root, func(fiber *Fiber) bool {
		if fiber == nil || fiber.VNode == nil {
			return true
		}
		// Skip ComponentVNode wrappers
		if fiber.VNode.Type() == rtui.VNodeComponent {
			return true
		}
		// Check if current VNode is focusable
		if focusable, ok := fiber.VNode.(rtui.FocusableVNode); ok && focusable.IsFocusable() {
			result = append(result, focusable)
		}
		return true
	})
	return result
}

// FindByTag finds all VNodes with the given tag name.
// This handles both ElementVNode and LayoutNode types.
func (w *TreeWalker) FindByTag(tag string) []*Fiber {
	result := make([]*Fiber, 0)
	w.walk(w.root, func(fiber *Fiber) bool {
		if fiber == nil || fiber.VNode == nil {
			return true
		}
		// Skip ComponentVNode wrappers
		if fiber.VNode.Type() == rtui.VNodeComponent {
			return true
		}
		// Check for Tag() method
		if tagger, ok := fiber.VNode.(interface{ Tag() string }); ok {
			if tagger.Tag() == tag {
				result = append(result, fiber)
			}
		}
		return true
	})
	return result
}

// Count returns the total number of Fibers in the tree.
func (w *TreeWalker) Count() int {
	count := 0
	w.walk(w.root, func(fiber *Fiber) bool {
		if fiber != nil {
			count++
		}
		return true
	})
	return count
}

// walk is the internal recursive walk function.
// It traverses the Fiber tree depth-first, calling the visitor function for each node.
// If the visitor returns false, traversal stops for that subtree.
func (w *TreeWalker) walk(fiber *Fiber, visitor func(*Fiber) bool) {
	if fiber == nil {
		return
	}

	// Visit current node
	if !visitor(fiber) {
		return
	}

	// Traverse children
	if fiber.Child != nil {
		w.walk(fiber.Child, visitor)
	}

	// Traverse siblings
	if fiber.Sibling != nil {
		w.walk(fiber.Sibling, visitor)
	}
}

// =============================================================================
// Helper Functions
// =============================================================================

// CollectFocusableFromFiber collects all focusable VNodes from a Fiber tree.
// This is a convenience function that creates a TreeWalker and collects focusable nodes.
func CollectFocusableFromFiber(root *Fiber) []rtui.FocusableVNode {
	return NewTreeWalker(root).CollectFocusable()
}

// FindFibersByTag finds all Fibers with the given tag name.
// This is a convenience function that creates a TreeWalker and finds nodes by tag.
func FindFibersByTag(root *Fiber, tag string) []*Fiber {
	return NewTreeWalker(root).FindByTag(tag)
}
