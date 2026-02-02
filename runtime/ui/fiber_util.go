package ui

// =============================================================================
// Fiber Creation
// =============================================================================

// CreateFiber creates a new fiber from a VNode
func CreateFiber(vnode VNode) *Fiber {
	if vnode == nil {
		return nil
	}

	fiber := &Fiber{
		VNode:         vnode,
		Type:          vnode.Type(),
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
func WalkFiberDepthFirst(root *Fiber, callback func(*Fiber) bool) bool {
	if root == nil {
		return true
	}

	// Visit current node
	if !callback(root) {
		return false
	}

	// Visit children
	if root.Child != nil {
		if !WalkFiberDepthFirst(root.Child, callback) {
			return false
		}
	}

	// Visit siblings
	if root.Sibling != nil {
		if !WalkFiberDepthFirst(root.Sibling, callback) {
			return false
		}
	}

	return true
}

// WalkFiberBreadthFirst walks the fiber tree in breadth-first order
func WalkFiberBreadthFirst(root *Fiber, callback func(*Fiber) bool) bool {
	if root == nil {
		return true
	}

	queue := []*Fiber{root}

	for len(queue) > 0 {
		// Dequeue
		current := queue[0]
		queue = queue[1:]

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
		UpdateQueue:   fiber.UpdateQueue,
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
