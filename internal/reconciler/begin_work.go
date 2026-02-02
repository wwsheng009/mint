package reconciler

// =============================================================================
// BeginWork Phase
// =============================================================================
// BeginWork is where reconciliation happens.
// For each Fiber, we:
// 1. Process updates (state changes)
// 2. Reconcile children (diff old vs new)
// 3. Return the next Fiber to process
//
// This is the "beginning" of processing a work unit.
// After BeginWork comes CompleteWork.
// =============================================================================

import (
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// BeginWork processes a Fiber node during the render phase
// Returns the next Fiber to process (usually nil, since we traverse in workLoop)
func BeginWork(current, workInProgress *Fiber) *Fiber {
	if workInProgress == nil {
		return nil
	}

	// Process updates in the queue
	processUpdateQueue(workInProgress)

	// Dispatch based on Fiber type
	switch workInProgress.Type {
	case rtui.VNodeComponent:
		return beginWorkComponent(current, workInProgress)

	case rtui.VNodeText:
		return beginWorkText(current, workInProgress)

	case rtui.VNodeElement:
		return beginWorkElement(current, workInProgress)

	case rtui.VNodeFragment:
		return beginWorkFragment(current, workInProgress)

	default:
		// Unknown type, skip
		return workInProgress
	}
}

// =============================================================================
// Component BeginWork
// =============================================================================

// beginWorkComponent processes a component Fiber
func beginWorkComponent(current, workInProgress *Fiber) *Fiber {
	componentVNode, ok := workInProgress.VNode.(*rtui.ComponentVNode)
	if !ok {
		return workInProgress
	}

	// Generate or get the component key for instance management
	componentKey := workInProgress.Key
	if componentKey == "" {
		// Use component name as key for single-instance components
		componentKey = "component:" + componentVNode.Name()
	}

	// Get or create component instance from InstanceManager
	// This ensures hooks state is preserved across renders for the same component
	var instance rtui.ComponentInstance
	var ctx *rtui.ComponentContext

	if currentReconciler != nil && currentReconciler.instanceMgr != nil {
		instance = currentReconciler.instanceMgr.GetOrCreate(componentKey, func() rtui.ComponentInstance {
			if componentVNode.FnWithProps() != nil {
				return rtui.NewBaseComponentInstanceWithProps(componentKey, componentVNode.FnWithProps(), workInProgress.Props)
			}
			return rtui.NewBaseComponentInstance(componentKey, componentVNode.Fn())
		})

		// Update props if they changed
		if workInProgress.Props != nil {
			instance.SetProps(workInProgress.Props)
		}

		// Store the instance in the fiber for later use
		workInProgress.ComponentInstance = instance

		// Get context from instance
		ctx = instance.GetContext()
	} else {
		// Fallback: create a temporary context if no reconciler
		// This should not happen in normal Fiber mode, but provides safety
		ctx = rtui.NewComponentContextForRoot() // Use root context as fallback
	}

	// CRITICAL: Reset hook index before re-rendering
	// This ensures hooks are called in the same order each render
	ctx.ResetContext()

	// Use the context for hooks
	oldContext := rtui.GetCurrentContext()
	rtui.SetCurrentContext(ctx)

	// Get children by calling the component function
	var children []rtui.VNode

	if componentVNode.Fn() != nil {
		// Simple component function
		vnode := componentVNode.Fn()()
		children = []rtui.VNode{vnode}
	} else if componentVNode.FnWithProps() != nil {
		// Component function with props
		vnode := componentVNode.FnWithProps()(workInProgress.Props)
		children = []rtui.VNode{vnode}
	} else {
		// No function - use rendered value from VNode
		rendered := componentVNode.Render()
		if rendered != nil {
			children = []rtui.VNode{rendered}
		}
	}

	// Restore old context
	rtui.SetCurrentContext(oldContext)

	// Get current child for reconciliation
	var currentChild *Fiber
	if current != nil {
		currentChild = current.Child
	}

	// Reconcile children
	workInProgress.Child = reconcileChildren(
		workInProgress,
		currentChild,
		children,
		workInProgress.Lanes,
	)

	return workInProgress
}

// =============================================================================
// Text BeginWork
// =============================================================================

// beginWorkText processes a text Fiber
// Text nodes have no children, so we just return
func beginWorkText(current, workInProgress *Fiber) *Fiber {
	// Text nodes are leaf nodes - no children to reconcile
	return workInProgress
}

// =============================================================================
// Element BeginWork
// =============================================================================

// beginWorkElement processes an element Fiber
func beginWorkElement(current, workInProgress *Fiber) *Fiber {
	var children []rtui.VNode

	// Handle both ElementVNode and LayoutNode (which embeds ElementVNode)
	switch v := workInProgress.VNode.(type) {
	case *rtui.ElementVNode:
		children = v.Children()
	case *rtui.LayoutNode:
		// LayoutNode embeds ElementVNode, so Children() works
		children = v.Children()
	default:
		return workInProgress
	}

	// Get current child for reconciliation
	var currentChild *Fiber
	if current != nil {
		currentChild = current.Child
	}

	// Reconcile children
	workInProgress.Child = reconcileChildren(
		workInProgress,
		currentChild,
		children,
		workInProgress.Lanes,
	)

	return workInProgress
}

// =============================================================================
// Fragment BeginWork
// =============================================================================

// beginWorkFragment processes a fragment Fiber
func beginWorkFragment(current, workInProgress *Fiber) *Fiber {
	fragmentVNode, ok := workInProgress.VNode.(*rtui.FragmentVNode)
	if !ok {
		return workInProgress
	}

	// Get children from fragment
	children := fragmentVNode.Children()

	// Get current child for reconciliation
	var currentChild *Fiber
	if current != nil {
		currentChild = current.Child
	}

	// Reconcile children
	workInProgress.Child = reconcileChildren(
		workInProgress,
		currentChild,
		children,
		workInProgress.Lanes,
	)

	return workInProgress
}

// =============================================================================
// Update Queue Processing
// =============================================================================

// processUpdateQueue processes all pending updates in the fiber's queue
func processUpdateQueue(workInProgress *Fiber) {
	if workInProgress.UpdateQueue == nil {
		return
	}

	// Process all updates in order
	for update := workInProgress.UpdateQueue.First; update != nil; update = update.Next {
		// Apply the update
		if update.Payload != nil {
			if fn, ok := update.Payload.(func(interface{}) interface{}); ok {
				// Functional update
				result := fn(workInProgress.MemoizedState)
				workInProgress.MemoizedState = result
			} else {
				// Direct value update
				workInProgress.MemoizedState = update.Payload
			}
		}

		// Merge lanes
		workInProgress.Lanes = MergeLanes(workInProgress.Lanes, update.Lane)

		// Mark as updated
		workInProgress.Flags |= EffectUpdate
	}

	// Clear the queue after processing
	workInProgress.UpdateQueue = nil
}
