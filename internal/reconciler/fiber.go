package reconciler

// =============================================================================
// Fiber Architecture
// =============================================================================
// Fiber 是 React 16+ 的协调算法，将渲染工作分解为小单元
// 每个 Fiber 节点代表 UI 中的一个组件，并形成一棵树
// =============================================================================
//
// This file now uses the Fiber types from runtime/ui/fiber.go
// The EffectFlag, Lane, and Fiber types are imported from rtui
// Reconciler-specific operations work with these types

import (
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// Re-export types from runtime/ui for convenience within the reconciler package
// This allows the reconciler code to use these types without the rtui prefix
type EffectFlag = rtui.EffectFlag
type Lane = rtui.Lane
type Fiber = rtui.Fiber
type Update = rtui.Update
type UpdateQueue = rtui.UpdateQueue

// Re-export constants
const (
	EffectNoEffect    = rtui.EffectNoEffect
	EffectPlacement   = rtui.EffectPlacement
	EffectUpdate      = rtui.EffectUpdate
	EffectDeletion    = rtui.EffectDeletion
	EffectRef         = rtui.EffectRef
	LaneNoLane        = rtui.LaneNoLane
	LaneSyncLane      = rtui.LaneSyncLane
	LaneInputContinuousLane = rtui.LaneInputContinuousLane
	LaneDefaultLane   = rtui.LaneDefaultLane
	LaneIdleLane      = rtui.LaneIdleLane
	LaneRoot          = rtui.LaneRoot
)

// =============================================================================
// Re-exported Functions from runtime/ui
// =============================================================================
// The following functions are re-exported from runtime/ui for convenience
// within the reconciler package. This allows reconciler code to use these
// functions without the rtui prefix.

// CreateFiber creates a new fiber from a VNode
var CreateFiber = rtui.CreateFiber

// CreateFiberFromVNode creates a fiber tree from a VNode tree
var CreateFiberFromVNode = rtui.CreateFiberFromVNode

// CloneFiber creates a shallow copy of a fiber
var CloneFiber = rtui.CloneFiber

// WalkFiberDepthFirst walks the fiber tree in depth-first order
var WalkFiberDepthFirst = rtui.WalkFiberDepthFirst

// WalkFiberBreadthFirst walks the fiber tree in breadth-first order
var WalkFiberBreadthFirst = rtui.WalkFiberBreadthFirst

// FindFiberByKey searches for a fiber with the given key in the subtree
var FindFiberByKey = rtui.FindFiberByKey

// CountFibers counts all fibers in the tree
var CountFibers = rtui.CountFibers

// GetFiberDepth returns the depth of a fiber in the tree
var GetFiberDepth = rtui.GetFiberDepth

// CollectFibersWithFlags collects all fibers with specific flags
var CollectFibersWithFlags = rtui.CollectFibersWithFlags

// MergeLanes combines two lane values
var MergeLanes = rtui.MergeLanes

// RemoveLanes removes lanes from a
var RemoveLanes = rtui.RemoveLanes

// HasLanes checks if 'a' contains any lanes from 'b'
var HasLanes = rtui.HasLanes

// IsSubsetLanes checks if 'a' is a subset of 'b'
var IsSubsetLanes = rtui.IsSubsetLanes

// GetHighestPriorityLane returns the highest priority lane set
var GetHighestPriorityLane = rtui.GetHighestPriorityLane

