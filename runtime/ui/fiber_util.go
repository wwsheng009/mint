package ui

// =============================================================================
// Fiber Creation
// =============================================================================

import (
	"os"

	"github.com/wwsheng009/mint/internal/log"
)

// CreateFiber creates a new fiber from a VNode
func CreateFiber(vnode VNode) *Fiber {
	if vnode == nil {
		return nil
	}

	vnodeType := vnode.Type()

	// Debug logging to understand VNode types
	if os.Getenv("TUI_DEBUG_HITMAP") == "true" {
		// Get actual type name for debugging
		actualType := "unknown"
		switch n := vnode.(type) {
		case *ElementVNode:
			actualType = "ElementVNode"
		case *ComponentVNode:
			actualType = "ComponentVNode"
		case *TextVNode:
			actualType = "TextVNode"
		case *LayoutNode:
			actualType = "LayoutNode"
		default:
			// Check if it's a component type (like ButtonVNode)
			if _, ok := n.(interface{ Tag() string }); ok {
				actualType = "ComponentWithElement"
			}
		}
		log.RenderLogger.Debug("[CREATEFIBER] Type=%s Key=%s actualType=%s", vnodeType.String(), vnode.Key(), actualType)
	}

	fiber := &Fiber{
		VNode:         vnode,
		Type:          vnodeType,
		Props:         vnode.Props(),
		MemoizedProps: vnode.Props(),
		Key:           vnode.Key(),
		Lanes:         LaneNoLane,
		ChildLanes:    LaneNoLane,
		Flags:         EffectNoEffect,
		SubtreeFlags:  EffectNoEffect,
	}

	// Set tag based on type
	switch n := vnode.(type) {
	case *ElementVNode:
		fiber.Tag = n.Tag()
	case *ComponentVNode:
		fiber.Tag = n.Name()
	case *LayoutNode:
		fiber.Tag = "layout"
	default:
		// For component types like TextVNode (from components package)
		// They implement VNode but aren't core types
		if t := vnode.Type(); t == VNodeText {
			fiber.Tag = "text"
		} else {
			fiber.Tag = "unknown"
		}
	}

	return fiber
}

// CreateFiberFromVNode creates a fiber tree from a VNode tree
func CreateFiberFromVNode(vnode VNode) *Fiber {
	root := CreateFiber(vnode)
	if root == nil {
		return nil
	}

	// Build fiber tree for children
	buildFiberTree(root, vnode)
	return root
}

// buildFiberTree recursively builds fiber tree for children
func buildFiberTree(parentFiber *Fiber, parentVNode VNode) {
	children := parentVNode.Children()
	if len(children) == 0 {
		return
	}

	var previousChild *Fiber
	for i, childVNode := range children {
		childFiber := CreateFiber(childVNode)

		// Link to parent
		childFiber.Return = parentFiber

		// Link siblings
		if i == 0 {
			parentFiber.Child = childFiber
		} else {
			previousChild.Sibling = childFiber
		}

		previousChild = childFiber

		// Recursively build for this child's children
		buildFiberTree(childFiber, childVNode)
	}
}

// =============================================================================
// Fiber Tree Traversal
// =============================================================================

// WalkFiberDepthFirst walks the fiber tree in depth-first order
// Uses iterative approach to avoid stack overflow on very deep trees
func WalkFiberDepthFirst(root *Fiber, callback func(*Fiber) bool) bool {
	if root == nil {
		return true
	}

	// Use explicit stack for iterative traversal
	// This avoids stack overflow for very deep trees (e.g., deeply nested lists)
	type frame struct {
		fiber    *Fiber
		state    int // 0 = visit self, 1 = visit children, 2 = visit siblings, 3 = done
		children bool // whether children were visited
		siblings bool // whether siblings were visited
	}

	stack := make([]frame, 0, 32)
	stack = append(stack, frame{fiber: root, state: 0})

	for len(stack) > 0 {
		// Peek at top of stack
		top := &stack[len(stack)-1]

		switch top.state {
		case 0: // Visit current node
			if !callback(top.fiber) {
				return false // Stop traversal
			}
			top.state = 1

		case 1: // Visit children
			if !top.children && top.fiber.Child != nil {
				top.children = true
				// Push child onto stack
				stack = append(stack, frame{fiber: top.fiber.Child, state: 0})
			} else {
				top.state = 2
			}

		case 2: // Visit siblings
			if !top.siblings && top.fiber.Sibling != nil {
				top.siblings = true
				// Push sibling onto stack
				stack = append(stack, frame{fiber: top.fiber.Sibling, state: 0})
			}
			top.state = 3

		case 3: // Done with this frame
			stack = stack[:len(stack)-1]
		}
	}

	return true
}

// WalkFiberBreadthFirst walks the fiber tree in breadth-first order
// Optimized to avoid slice allocation on each dequeue operation
func WalkFiberBreadthFirst(root *Fiber, callback func(*Fiber) bool) bool {
	if root == nil {
		return true
	}

	// Pre-allocate queue with reasonable capacity to reduce allocations
	queue := make([]*Fiber, 0, 32)
	queue = append(queue, root)

	for i := 0; i < len(queue); i++ {
		// Dequeue by index - avoids slice allocation from queue[1:]
		current := queue[i]

		// Visit current node
		if !callback(current) {
			return false
		}

		// Enqueue children
		for child := current.Child; child != nil; child = child.Sibling {
			queue = append(queue, child)
		}
	}

	return true
}

// =============================================================================
// Fiber Utilities
// =============================================================================

// CloneFiber creates a shallow copy of a fiber
func CloneFiber(fiber *Fiber) *Fiber {
	if fiber == nil {
		return nil
	}

	return &Fiber{
		VNode:         fiber.VNode,
		Type:          fiber.Type,
		Tag:           fiber.Tag,
		Key:           fiber.Key,
		Props:         fiber.Props,
		MemoizedProps: fiber.MemoizedProps,
		MemoizedState: fiber.MemoizedState,
		Return:        fiber.Return,
		Child:         fiber.Child,
		Sibling:       fiber.Sibling,
		Alternate:     fiber.Alternate,
		// Don't share UpdateQueue - cloned fiber gets its own empty queue
		// This prevents updates to the clone from affecting the original
		UpdateQueue:   nil,
		Flags:         fiber.Flags,
		SubtreeFlags:  fiber.SubtreeFlags,
		Lanes:         fiber.Lanes,
		ChildLanes:    fiber.ChildLanes,
	}
}

// =============================================================================
// Fiber Tree Utilities
// =============================================================================

// FindFiberByKey searches for a fiber with the given key in the subtree
func FindFiberByKey(root *Fiber, key string) *Fiber {
	var result *Fiber

	WalkFiberDepthFirst(root, func(fiber *Fiber) bool {
		if fiber.Key == key {
			result = fiber
			return false // Stop traversal
		}
		return true
	})

	return result
}

// CountFibers counts all fibers in the tree
func CountFibers(root *Fiber) int {
	count := 0

	WalkFiberDepthFirst(root, func(_ *Fiber) bool {
		count++
		return true
	})

	return count
}

// GetFiberDepth returns the depth of a fiber in the tree
func GetFiberDepth(fiber *Fiber) int {
	depth := 0
	for p := fiber.Return; p != nil; p = p.Return {
		depth++
	}
	return depth
}

// CollectFibersWithFlags collects all fibers with specific flags
func CollectFibersWithFlags(root *Fiber, flags EffectFlag) []*Fiber {
	var result []*Fiber

	WalkFiberDepthFirst(root, func(fiber *Fiber) bool {
		if (fiber.Flags & flags) != 0 {
			result = append(result, fiber)
		}
		return true
	})

	return result
}
