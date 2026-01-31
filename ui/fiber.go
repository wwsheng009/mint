package ui

// =============================================================================
// Fiber Architecture
// =============================================================================
// Fiber 是 React 16+ 的协调算法，将渲染工作分解为小单元
// 每个 Fiber 节点代表 UI 中的一个组件，并形成一棵树
// =============================================================================

import (
	"fmt"
)

// EffectFlag represents the side effects of a fiber
type EffectFlag int

const (
	// EffectNoEffect indicates no side effects
	EffectNoEffect EffectFlag = 0
	// EffectPlacement indicates the fiber was newly inserted
	EffectPlacement EffectFlag = 1 << iota
	// EffectUpdate indicates the fiber was updated
	EffectUpdate
	// EffectDeletion indicates the fiber was deleted
	EffectDeletion
	// EffectRef indicates a ref changed
	EffectRef
)

// Lane represents the priority of work (lanes model)
type Lane uint64

const (
	// LaneNoLane indicates no lane
	LaneNoLane Lane = 0
	// LaneSyncLane is the default synchronous lane
	LaneSyncLane Lane = 1
	// LaneInputContinuousLane is for continuous input (dragging, typing)
	LaneInputContinuousLane Lane = 1 << 1
	// LaneDefaultLane is for default updates
	LaneDefaultLane Lane = 1 << 2
	// LaneIdleLane is for low-priority work
	LaneIdleLane Lane = 1 << 3
)

// LaneRoot represents all lanes combined
const LaneRoot Lane = LaneSyncLane | LaneInputContinuousLane | LaneDefaultLane | LaneIdleLane

// Fiber represents a unit of work in the reconciler
// Each Fiber corresponds to a VNode and contains work-in-progress state
type Fiber struct {
	// === VNode Reference ===
	// The virtual node this fiber represents
	VNode VNode

	// === Tree Structure ===
	// Pointer to parent fiber
	Return *Fiber
	// First child fiber
	Child *Fiber
	// Next sibling fiber
	Sibling *Fiber
	// For alternate tree (double buffering)
	Alternate *Fiber

	// === Props & State ===
	// Props for this fiber
	Props Props
	// Memoized props (previous props)
	MemoizedProps Props
	// Memoized state
	MemoizedState interface{}
	// Update queue
	UpdateQueue *UpdateQueue

	// === Effect Flags ===
	// Flags indicating what side effects this fiber has
	Flags EffectFlag
	// SubtreeFlags indicating effects in descendants
	SubtreeFlags EffectFlag

	// === Priority ===
	// Lanes representing work priority
	Lanes Lane
	// ChildLanes for pending work in children
	ChildLanes Lane

	// === Key for Diff ===
	// Key for reconciling lists
	Key string

	// === Type ===
	// The type of this fiber (cached from VNode)
	Type VNodeType

	// === Tag for debugging ===
	// Component or element tag
	Tag string
}

// Update represents a state update
type Update struct {
	// The new state or updater function
	Payload interface{}
	// The next update in the queue
	Next *Update
	// Lane for this update
	Lane Lane
	// Callback after update is committed
	Callback func()
}

// UpdateQueue holds pending updates
type UpdateQueue struct {
	// First pending update
	First *Update
	// Last pending update
	Last *Update
}

// =============================================================================
// Fiber Creation
// =============================================================================

// CreateFiber creates a new fiber from a VNode
func CreateFiber(vnode VNode) *Fiber {
	if vnode == nil {
		return nil
	}

	fiber := &Fiber{
		VNode:        vnode,
		Type:         vnode.Type(),
		Props:        vnode.Props(),
		MemoizedProps: vnode.Props(),
		Key:          vnode.Key(),
		Lanes:        LaneNoLane,
		ChildLanes:   LaneNoLane,
		Flags:        EffectNoEffect,
		SubtreeFlags: EffectNoEffect,
	}

	// Set tag based on type
	switch n := vnode.(type) {
	case *ElementVNode:
		fiber.Tag = n.Tag()
	case *TextVNode:
		fiber.Tag = "text"
	case *ComponentVNode:
		fiber.Tag = n.Name()
	case *LayoutNode:
		fiber.Tag = "layout"
	default:
		fiber.Tag = "unknown"
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
// Fiber Helpers
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

// HasNoPendingWork returns true if the fiber has no pending work
func (f *Fiber) HasNoPendingWork() bool {
	return f.Lanes == LaneNoLane && f.ChildLanes == LaneNoLane
}

// HasEffect returns true if the fiber has effects
func (f *Fiber) HasEffect() bool {
	return f.Flags != EffectNoEffect
}

// HasSubtreeEffect returns true if any descendant has effects
func (f *Fiber) HasSubtreeEffect() bool {
	return f.SubtreeFlags != EffectNoEffect
}

// MarkUpdate marks the fiber as needing an update
func (f *Fiber) MarkUpdate(lane Lane) {
	f.Lanes = mergeLanes(f.Lanes, lane)
	f.Flags |= EffectUpdate

	// Propagate to parents
	for parent := f.Return; parent != nil; parent = parent.Return {
		parent.ChildLanes = mergeLanes(parent.ChildLanes, lane)
	}
}

// =============================================================================
// Update Queue Operations
// =============================================================================

// EnqueueUpdate adds an update to the fiber's queue
func (f *Fiber) EnqueueUpdate(update *Update) {
	if f.UpdateQueue == nil {
		f.UpdateQueue = &UpdateQueue{}
	}

	update.Next = nil
	if f.UpdateQueue.Last == nil {
		f.UpdateQueue.First = update
		f.UpdateQueue.Last = update
	} else {
		f.UpdateQueue.Last.Next = update
		f.UpdateQueue.Last = update
	}

	// Mark fiber as having work
	f.MarkUpdate(LaneSyncLane)
}

// =============================================================================
// Lane Operations
// =============================================================================

// mergeLanes combines two lane values
func mergeLanes(a, b Lane) Lane {
	return a | b
}

// removeLanes removes lanes from a
func removeLanes(a, b Lane) Lane {
	return a &^ b
}

// hasLanes checks if 'a' contains any lanes from 'b'
func hasLanes(a, b Lane) bool {
	return (a & b) != 0
}

// isSubsetLanes checks if 'a' is a subset of 'b'
func isSubsetLanes(a, b Lane) bool {
	return (a & b) == a
}

// getHighestPriorityLane returns the highest priority lane set
func getHighestPriorityLane(lanes Lane) Lane {
	// Find the least significant bit (highest priority)
	return lanes & (-lanes)
}

// =============================================================================
// Fiber String Representation
// =============================================================================

// String returns a string representation of the fiber
func (f *Fiber) String() string {
	if f == nil {
		return "nil"
	}

	return fmt.Sprintf("Fiber{Tag: %s, Key: %s, Flags: %d, Lanes: %d}",
		f.Tag, f.Key, f.Flags, f.Lanes)
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
