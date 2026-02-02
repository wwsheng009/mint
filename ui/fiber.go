package ui

import rtui "github.com/wwsheng009/mint/runtime/ui"

// =============================================================================
// Fiber Types (re-exported from runtime/types)
// =============================================================================

// Lane represents the priority of work (lanes model)
type Lane = rtui.Lane

const (
	// LaneNoLane indicates no lane
	LaneNoLane = rtui.LaneNoLane
	// LaneSyncLane is the default synchronous lane
	LaneSyncLane = rtui.LaneSyncLane
	// LaneInputContinuousLane is for continuous input (dragging, typing)
	LaneInputContinuousLane = rtui.LaneInputContinuousLane
	// LaneDefaultLane is for default updates
	LaneDefaultLane = rtui.LaneDefaultLane
	// LaneIdleLane is for low-priority work
	LaneIdleLane = rtui.LaneIdleLane
	// LaneRoot represents all lanes combined
	LaneRoot = rtui.LaneRoot
)

// EffectFlag represents the side effects of a fiber
type EffectFlag = rtui.EffectFlag

const (
	// EffectNoEffect indicates no side effects
	EffectNoEffect = rtui.EffectNoEffect
	// EffectPlacement indicates the fiber was newly inserted
	EffectPlacement = rtui.EffectPlacement
	// EffectUpdate indicates the fiber was updated
	EffectUpdate = rtui.EffectUpdate
	// EffectDeletion indicates the fiber was deleted
	EffectDeletion = rtui.EffectDeletion
	// EffectRef indicates a ref changed
	EffectRef = rtui.EffectRef
)

// Fiber represents a unit of work in the reconciler
type Fiber = rtui.Fiber

// Update represents a state update
type Update = rtui.Update

// UpdateQueue holds pending updates
type UpdateQueue = rtui.UpdateQueue

// =============================================================================
// Fiber Functions (forwarded to runtime/types)
// =============================================================================

// CreateFiber creates a new fiber from a VNode
func CreateFiber(vnode VNode) *Fiber {
	return rtui.CreateFiber(vnode)
}

// CreateFiberFromVNode creates a fiber tree from a VNode tree
func CreateFiberFromVNode(vnode VNode) *Fiber {
	return rtui.CreateFiberFromVNode(vnode)
}

// WalkFiberDepthFirst walks the fiber tree in depth-first order
func WalkFiberDepthFirst(root *Fiber, callback func(*Fiber) bool) bool {
	return rtui.WalkFiberDepthFirst(root, callback)
}

// WalkFiberBreadthFirst walks the fiber tree in breadth-first order
func WalkFiberBreadthFirst(root *Fiber, callback func(*Fiber) bool) bool {
	return rtui.WalkFiberBreadthFirst(root, callback)
}

// CloneFiber creates a shallow copy of a fiber
func CloneFiber(fiber *Fiber) *Fiber {
	return rtui.CloneFiber(fiber)
}

// MergeLanes combines two lane values
func MergeLanes(a, b Lane) Lane {
	return rtui.MergeLanes(a, b)
}

// FindFiberByKey searches for a fiber with the given key in the subtree
func FindFiberByKey(root *Fiber, key string) *Fiber {
	return rtui.FindFiberByKey(root, key)
}

// CountFibers counts all fibers in the tree
func CountFibers(root *Fiber) int {
	return rtui.CountFibers(root)
}

// GetFiberDepth returns the depth of a fiber in the tree
func GetFiberDepth(fiber *Fiber) int {
	return rtui.GetFiberDepth(fiber)
}

// CollectFibersWithFlags collects all fibers with specific flags
func CollectFibersWithFlags(root *Fiber, flags EffectFlag) []*Fiber {
	return rtui.CollectFibersWithFlags(root, flags)
}
