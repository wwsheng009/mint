package reconciler

import (
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// TreeWalker provides a unified abstraction for traversing Fiber trees.
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
func (w *TreeWalker) CollectFocusable() []rtui.FocusableVNode {
	count := 0
	w.walk(w.root, func(fiber *Fiber) bool {
		if fiber != nil {
			count++
		}
		return true
	})

	result := make([]rtui.FocusableVNode, 0, count/4+4)

	w.walk(w.root, func(fiber *Fiber) bool {
		if fiber == nil {
			return true
		}
		// Skip ComponentVNode wrappers
		if fiber.Type == rtui.VNodeComponent {
			return true
		}
		// Check if ComponentInstance is focusable via interface check
		if fiber.ComponentInstance != nil {
			if focusable, ok := fiber.ComponentInstance.(interface{ IsFocusable() bool }); ok && focusable.IsFocusable() {
				if vnode, ok := focusable.(rtui.FocusableVNode); ok {
					result = append(result, vnode)
				}
			}
		}
		return true
	})
	return result
}

// FindByTag finds all Fibers with the given tag name.
func (w *TreeWalker) FindByTag(tag string) []*Fiber {
	result := make([]*Fiber, 0)
	w.walk(w.root, func(fiber *Fiber) bool {
		if fiber == nil {
			return true
		}
		// Skip ComponentVNode wrappers
		if fiber.Type == rtui.VNodeComponent {
			return true
		}
		// Check tag
		if fiber.Tag == tag {
			result = append(result, fiber)
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
func (w *TreeWalker) walk(fiber *Fiber, visitor func(*Fiber) bool) {
	if fiber == nil {
		return
	}

	if !visitor(fiber) {
		return
	}

	if fiber.Child != nil {
		w.walk(fiber.Child, visitor)
	}

	if fiber.Sibling != nil {
		w.walk(fiber.Sibling, visitor)
	}
}

// CollectFocusableFromFiber collects all focusable VNodes from a Fiber tree.
func CollectFocusableFromFiber(root *Fiber) []rtui.FocusableVNode {
	return NewTreeWalker(root).CollectFocusable()
}

// FindFibersByTag finds all Fibers with the given tag name.
func FindFibersByTag(root *Fiber, tag string) []*Fiber {
	return NewTreeWalker(root).FindByTag(tag)
}
