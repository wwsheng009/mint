package ui

import (
	"sort"

	"github.com/wwsheng009/mint/internal/log"
	runtimelayout "github.com/wwsheng009/mint/runtime/layout"
	rtuievent "github.com/wwsheng009/mint/runtime/event"
)

// =============================================================================
// NodeID Generation
// =============================================================================
// Global ID allocator for NodeID generation
var nodeIDGenerator uint64 = 0

// generateNodeID generates a new unique NodeID
// This provides stable runtime identity for Fiber nodes
// See: docs/render/fiber/IDENTITY_REFACTORING_PLAN.md
func generateNodeID() uint64 {
	nodeIDGenerator++
	return nodeIDGenerator
}

// =============================================================================
// Fiber Creation
// =============================================================================

// CreateFiber creates a new fiber from a VNode
func CreateFiber(vnode VNode) *Fiber {
	if vnode == nil {
		return nil
	}

	vnodeType := vnode.Type()

	// Debug logging to understand VNode types
	if log.HitMapLogger.Enabled() {
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
		log.HitMapLogger.Debug("[CREATEFIBER] Type=%s Key=%s actualType=%s", vnodeType.String(), vnode.Key(), actualType)
	}

	// ✨ DiffKey: Copy from VNode.Key() without any modification
	// This is the PRIMARY key used for diffing (reconciliation)
	// It is NOT generated from Path - Path is only for debugging
	diffKey := vnode.Key()

	fiber := &Fiber{
		VNode:         vnode,
		Type:          vnodeType,
		Props:         vnode.Props(),
		MemoizedProps: vnode.Props(),
		DiffKey:       diffKey,  // ✨ Copy DiffKey directly
		Key:           diffKey,  // Backward compatibility
		NodeID:        generateNodeID(), // ✨ Allocate unique NodeID
		Layer:         vnode.GetLayer(), // ✨ Copy Layer from VNode
		Lanes:         LaneNoLane,
		ChildLanes:    LaneNoLane,
		Flags:         EffectNoEffect,
		SubtreeFlags:  EffectNoEffect,
		ComputedBox:   nil, // ✨ ComputedBox is nil initially
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

// WalkFiberDepthFirst walks fiber tree in depth-first order
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

// WalkFiberBreadthFirst walks fiber tree in breadth-first order
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
// ✨ Preserves NodeID for stable runtime identity
func CloneFiber(fiber *Fiber) *Fiber {
	if fiber == nil {
		return nil
	}

	return &Fiber{
		VNode:         fiber.VNode,
		Type:          fiber.Type,
		Tag:           fiber.Tag,
		DiffKey:       fiber.DiffKey,  // ✨ Preserve DiffKey for diffing
		Key:           fiber.Key,       // Backward compatibility
		NodeID:        fiber.NodeID,   // ✨ Preserve NodeID for stable identity
		Layer:         fiber.Layer,    // ✨ Preserve Layer
		Props:         fiber.Props,
		MemoizedProps: fiber.MemoizedProps,
		MemoizedState: fiber.MemoizedState,
		Return:        fiber.Return,
		Child:         fiber.Child,
		Sibling:       fiber.Sibling,
		Alternate:     fiber.Alternate,
		// Don't share UpdateQueue - cloned fiber gets its own empty queue
		// This prevents updates to clone from affecting original
		UpdateQueue:   nil,
		Flags:         fiber.Flags,
		SubtreeFlags:  fiber.SubtreeFlags,
		Lanes:         fiber.Lanes,
		ChildLanes:    fiber.ChildLanes,
		ComputedBox:   nil, // ✨ Reset ComputedBox (will be re-calculated)
	}
}

// =============================================================================
// BuildHitMapFromFiber (Non-Reflection Implementation)
// =============================================================================
// BuildHitMapFromFiber builds a HitMap from a Fiber tree without using reflection
//
// This implementation:
// 1. Uses WalkFiberDepthFirst to traverse the Fiber tree
// 2. Reads ComputedBox from each Fiber (set during layout phase)
// 3. Builds HitMap entries with NodeID, Layer, Bounds, etc.
// 4. Sorts by Layer (Z-order): Base(0) < Overlay(1) < Modal(2) < Tooltip(3) < Inspector(4)
// 5. Returns *event.HitMap
//
// NOTE: This function requires Fiber.ComputedBox to be populated by
// Engine.layoutFiber() during the layout phase.
//
// NO REFLECTION: Unlike the version in runtime/event/hitmap.go, this implementation
// uses direct field access and WalkFiberDepthFirst for tree traversal.
//
// Usage: After layout phase, when Fiber.ComputedBox is populated
//
// See: docs/plan/fiber/TODO_LIST.md Phase 1.5, Phase 2.4
func BuildHitMapFromFiber(root *Fiber) *rtuievent.HitMap {
	if root == nil {
		return rtuievent.NewHitMap()
	}

	// Collect entries from Fiber tree
	var entries []rtuievent.HitMapEntryInternal

	WalkFiberDepthFirst(root, func(fiber *Fiber) bool {
		// Skip fibers without ComputedBox
		if fiber.ComputedBox == nil {
			return true
		}

		// We need to access Box fields without importing compute package
		// Use type assertion to access the underlying struct
		// ComputedBox embeds runtime.Box which has X, Y, Width, Height fields
		if box, ok := fiber.ComputedBox.(interface {
			GetX() int
			GetY() int
			GetWidth() int
			GetHeight() int
		}); ok {
			x := box.GetX()
			y := box.GetY()
			width := box.GetWidth()
			height := box.GetHeight()

			// Use Fiber.NodeID as the primary ID
			nodeID := fiber.NodeID

			// Get Layer from Fiber
			layer := fiber.Layer

			// Calculate tree depth for Z-order calculation
			depth := 0
			for p := fiber.Return; p != nil; p = p.Return {
				depth++
			}

			// Calculate Z-order: Layer * 10000 + treeDepth
			// This ensures higher layers are prioritized in HitTest
			zOrder := int(layer)*10000 + depth

			// Create HitMap entry
			entry := rtuievent.HitMapEntryInternal{
				NodeID: nodeID,
				Node:   nil, // Node is nil because we can't create layout.Node without importing more types
				Bounds: runtimelayout.Rect{
					X:      x,
					Y:      y,
					Width:  width,
					Height: height,
				},
				LocalXY: func(screenX, screenY int) (int, int) {
					return screenX - x, screenY - y
				},
				ZOrder:   zOrder,
				Instance: nil,
			}

			entries = append(entries, entry)
		}

		return true
	})

	// Sort by ZOrder descending (higher layers first)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].ZOrder > entries[j].ZOrder
	})

	return rtuievent.BuildHitMapFromEntries(entries)
}

// =============================================================================
// Fiber Tree Utilities
// =============================================================================

// FindFiberByKey searches for a fiber with a given key in subtree
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

// CountFibers counts all fibers in tree
func CountFibers(root *Fiber) int {
	count := 0
	WalkFiberDepthFirst(root, func(_ *Fiber) bool {
		count++
		return true
	})
	return count
}

// GetFiberDepth returns depth of a fiber in tree
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

// =============================================================================
// Fiber Layout Helper Methods (Phase 1)
// =============================================================================
// These methods enable Fiber-first layout by providing access to child fibers
// and layout properties directly from the Fiber struct.

// GetChildFibers returns all child fibers as a slice
// Converts the Child → Sibling linked list to an array for easier iteration
func (f *Fiber) GetChildFibers() []*Fiber {
	var children []*Fiber
	for child := f.Child; child != nil; child = child.Sibling {
		children = append(children, child)
	}
	return children
}

// GetChildCount returns the number of children
func (f *Fiber) GetChildCount() int {
	count := 0
	for child := f.Child; child != nil; child = child.Sibling {
		count++
	}
	return count
}

// GetDirection returns the layout direction
// Prioritizes Fiber.LayoutDirection field, falls back to VNode
func (f *Fiber) GetDirection() Direction {
	if f.LayoutDirection != 0 {
		return f.LayoutDirection
	}
	// Fallback to VNode during transition
	if f.VNode != nil {
		if ln, ok := f.VNode.(*LayoutNode); ok {
			return ln.direction
		}
	}
	return DirectionRow // default
}

// GetAlign returns the main axis alignment
// Prioritizes Fiber.LayoutAlign field, falls back to VNode
func (f *Fiber) GetAlign() Align {
	if f.LayoutAlign != 0 {
		return f.LayoutAlign
	}
	if f.VNode != nil {
		if ln, ok := f.VNode.(*LayoutNode); ok {
			return ln.align
		}
	}
	return AlignStart // default
}

// GetCrossAlign returns the cross axis alignment
// Prioritizes Fiber.LayoutCrossAlign field, falls back to VNode
func (f *Fiber) GetCrossAlign() Align {
	if f.LayoutCrossAlign != 0 {
		return f.LayoutCrossAlign
	}
	if f.VNode != nil {
		if ln, ok := f.VNode.(*LayoutNode); ok {
			return ln.crossAlign
		}
	}
	return AlignStart // default
}

// GetGap returns the gap spacing between children
// Prioritizes Fiber.LayoutGap field, falls back to VNode
func (f *Fiber) GetGap() int {
	if f.LayoutGap != 0 || f.VNode == nil {
		return f.LayoutGap
	}
	if ln, ok := f.VNode.(*LayoutNode); ok {
		return ln.gap
	}
	return 0 // default
}

// GetPadding returns the padding [top, right, bottom, left]
// Prioritizes Fiber.LayoutPadding field, falls back to VNode
func (f *Fiber) GetPadding() [4]int {
	// Check if any padding value is non-zero
	if f.LayoutPadding[0] != 0 || f.LayoutPadding[1] != 0 ||
		f.LayoutPadding[2] != 0 || f.LayoutPadding[3] != 0 || f.VNode == nil {
		return f.LayoutPadding
	}
	if ln, ok := f.VNode.(*LayoutNode); ok {
		return ln.padding
	}
	return [4]int{0, 0, 0, 0} // default
}

// GetFlex returns the flex factor
// Prioritizes Fiber.LayoutFlex field, falls back to VNode
func (f *Fiber) GetFlex() int {
	if f.LayoutFlex != 0 || f.VNode == nil {
		return f.LayoutFlex
	}
	if ln, ok := f.VNode.(*LayoutNode); ok {
		return ln.flex
	}
	return 0 // default
}
