// Package ui provides the BeginWork phase for Fiber reconciliation.
package ui

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

// currentReconciler holds the active reconciler during render
// This is set by the reconciler before calling BeginWork
var currentReconciler *Reconciler

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
	case VNodeComponent:
		return beginWorkComponent(current, workInProgress)

	case VNodeText:
		return beginWorkText(current, workInProgress)

	case VNodeElement:
		return beginWorkElement(current, workInProgress)

	case VNodeFragment:
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
	componentVNode, ok := workInProgress.VNode.(*ComponentVNode)
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
	var instance ComponentInstance
	if currentReconciler != nil && currentReconciler.instanceMgr != nil {
		instance = currentReconciler.instanceMgr.GetOrCreate(componentKey, func() ComponentInstance {
			if componentVNode.FnWithProps() != nil {
				return NewBaseComponentInstanceWithProps(componentKey, componentVNode.FnWithProps(), workInProgress.Props)
			}
			return NewBaseComponentInstance(componentKey, componentVNode.Fn())
		})

		// Update props if they changed
		if workInProgress.Props != nil {
			instance.SetProps(workInProgress.Props)
		}

		// Store the instance in the fiber for later use
		workInProgress.ComponentInstance = instance

		// Use the instance's context for hooks
		oldContext := GetCurrentContext()
		SetCurrentContext(instance.GetContext())
		defer SetCurrentContext(oldContext)
	}

	// Get children by calling the component function
	var children []VNode

	if componentVNode.Fn() != nil {
		// Simple component function
		vnode := componentVNode.Fn()()
		children = []VNode{vnode}
	} else if componentVNode.FnWithProps() != nil {
		// Component function with props
		vnode := componentVNode.FnWithProps()(workInProgress.Props)
		children = []VNode{vnode}
	} else {
		// No function - use rendered value from VNode
		rendered := componentVNode.Render()
		if rendered != nil {
			children = []VNode{rendered}
		}
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
	var children []VNode

	// Handle both ElementVNode and LayoutNode (which embeds ElementVNode)
	switch v := workInProgress.VNode.(type) {
	case *ElementVNode:
		children = v.Children()
	case *LayoutNode:
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
	fragmentVNode, ok := workInProgress.VNode.(*FragmentVNode)
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
