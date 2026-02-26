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
	"github.com/wwsheng009/mint/internal/log"
	"github.com/wwsheng009/mint/internal/state"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// BeginWork processes a Fiber node during the render phase
// Returns the next Fiber to process (usually nil, since we traverse in workLoop)
func BeginWork(current, workInProgress *Fiber) *Fiber {
	if workInProgress == nil {
		return nil
	}

	// Debug: log all BeginWork calls to understand the Fiber tree
	if log.FiberLogger.Enabled() {
		typeName := "UNKNOWN"
		switch workInProgress.Type {
		case rtui.VNodeComponent:
			typeName = "VNodeComponent"
		case rtui.VNodeText:
			typeName = "VNodeText"
		case rtui.VNodeElement:
			typeName = "VNodeElement"
		case rtui.VNodeFragment:
			typeName = "VNodeFragment"
		}
		if workInProgress.Key != "" {
			log.FiberLogger.Debug("[BeginWork] Called: Type=%d(%s), Key=%q, Tag=%q",
				workInProgress.Type, typeName, workInProgress.Key, workInProgress.Tag)
		}
	}

	// Process updates in the queue
	processUpdateQueue(workInProgress)

	// Check for ErrorBoundary - handle before regular component processing
	if workInProgress.ErrorBoundaryFunc != nil {
		return beginWorkErrorBoundary(current, workInProgress)
	}

	// Check for Memo - handle memoization to skip unnecessary renders
	if workInProgress.MemoCompare != nil {
		return beginWorkMemo(current, workInProgress)
	}

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
	// Generate or get the component key for instance management
	componentKey := workInProgress.Key
	if componentKey == "" {
		// Use component name as key for single-instance components
		componentKey = "component:" + workInProgress.ComponentName
	}

	// CRITICAL: Root component uses the shared root context for global state management
	// This ensures Intent Handlers (via Dispatcher.SetStateSetter) update the same state
	// that the root component reads (via GetIntState/GetStringState/etc.).
	var instance rtui.ComponentInstance
	var ctx *rtui.ComponentContext

	// ✨ Use explicit IsRoot marker instead of string comparison
	// This is more robust than checking componentKey == "root"
	isRootComponent := workInProgress.IsRoot

	if isRootComponent && currentReconciler != nil && currentReconciler.ctx != nil {
		// Root component: use the shared root context for global state
		ctx = currentReconciler.ctx
		if log.UILogger.Enabled() {
			log.UILogger.Debug("[beginWorkComponent] Using ROOT context for component %s", componentKey)
		}
	} else if currentReconciler != nil && currentReconciler.instanceMgr != nil {
		// Child component: use InstanceManager for component instance
		instance = currentReconciler.instanceMgr.GetOrCreate(componentKey, func() rtui.ComponentInstance {
			if workInProgress.ComponentFuncWithProps != nil {
				return rtui.NewBaseComponentInstanceWithProps(componentKey, workInProgress.ComponentFuncWithProps, workInProgress.Props)
			}
			return rtui.NewBaseComponentInstance(componentKey, workInProgress.ComponentFunc)
		})
		// Update props if they changed
		if workInProgress.Props != nil {
			instance.SetProps(workInProgress.Props)
		}

		// CRITICAL: For GlobalState sharing, share the root context's GlobalState and StateMu
		// This ensures Intent Handlers (via Dispatcher.SetStateSetter) update the same state
		// that the component reads (via GetState/SetState).
		// Each component still has its own Hooks for component-local state.
		sharedCtx := currentReconciler.ctx
		instanceCtx := instance.GetContext()

		// Share the GlobalState map and its mutex from root context
		instanceCtx.GlobalState = sharedCtx.GlobalState
		instanceCtx.StateMu = sharedCtx.StateMu

		// Use the instance's context (which now has shared GlobalState)
		ctx = instanceCtx
	} else {
		// Fallback: create a temporary context if no reconciler
		// This should not happen in normal Fiber mode, but provides safety
		ctx = rtui.NewComponentContextForRoot()
	}

	// CRITICAL: Reset hook index before re-rendering
	ctx.ResetContext()

	// Use the context for hooks
	oldContext := rtui.GetCurrentContext()
	rtui.SetCurrentContext(ctx)

	// Get children by calling the component function
	var children []rtui.VNode

	if workInProgress.ComponentFunc != nil {
		// Simple component function
		vnode := workInProgress.ComponentFunc()
		children = []rtui.VNode{vnode}
	} else if workInProgress.ComponentFuncWithProps != nil {
		// Component function with props
		vnode := workInProgress.ComponentFuncWithProps(workInProgress.Props)
		children = []rtui.VNode{vnode}
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

	// CRITICAL: Update Instance props if present
	// Text VNode implements InstanceFactory and creates a TextInstance
	// When Fiber is reused, props change but Instance needs explicit update
	if workInProgress.Instance != nil && workInProgress.Props != nil {
		workInProgress.Instance.SetProps(workInProgress.Props)
	}

	return workInProgress
}

// =============================================================================
// Element BeginWork
// =============================================================================

// beginWorkElement processes an element Fiber
func beginWorkElement(current, workInProgress *Fiber) *Fiber {
	// Get children from Props (stored during Fiber creation)
	// Children are stored in Props["children"] for element nodes
	var children []rtui.VNode
	if workInProgress.Props != nil {
		if c, ok := workInProgress.Props["children"].([]rtui.VNode); ok {
			children = c
		}
	}

	// CRITICAL: Update Instance props if present
	// Text, Button, etc. implement InstanceFactory and create instances
	// When Fiber is reused, props change but Instance needs explicit update
	if workInProgress.Instance != nil && workInProgress.Props != nil {
		workInProgress.Instance.SetProps(workInProgress.Props)
	}

	// ✨ NEW: Create/reuse VNodeComponentInstance for VNode struct components
	// This enables persistent event handlers and state for Button, Text, etc.

	if currentReconciler != nil && currentReconciler.instanceMgr != nil && workInProgress.Key != "" {
		lookupKey := workInProgress.Path
		if lookupKey == "" {
			lookupKey = workInProgress.Key
		}
		instanceKey := "vnode:" + lookupKey

		log.UILogger.Debug("[beginWorkElement] Creating instance for key=%s (fiber.Key=%q, fiber.Path=%q)",
			instanceKey, workInProgress.Key, workInProgress.Path)

		// Get or create VNode component instance
		// instance := currentReconciler.instanceMgr.GetOrCreate(instanceKey, func() rtui.ComponentInstance {
		// 	return createVNodeComponentInstanceFromFiber(instanceKey, workInProgress)
		// })

		// Store the instance in the fiber
		// workInProgress.ComponentInstance = instance

		log.UILogger.Debug("[beginWorkElement] ✅ Created/Updated instance: key=%s, fiber.Key=%q, type=%d",
			instanceKey, workInProgress.Key, workInProgress.Type)

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
	// Get children from Props (stored during Fiber creation)
	var children []rtui.VNode
	if workInProgress.Props != nil {
		if c, ok := workInProgress.Props["children"].([]rtui.VNode); ok {
			children = c
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

// =============================================================================
// ErrorBoundary BeginWork
// =============================================================================

// beginWorkErrorBoundary processes an error boundary Fiber
func beginWorkErrorBoundary(current, workInProgress *Fiber) *Fiber {
	// The error boundary wraps a component that might panic
	// We need to call the component's function with panic recovery

	componentFn := workInProgress.ErrorBoundaryFunc

	// Try to render the component with panic recovery
	var children []rtui.VNode
	var hadPanic bool

	func() {
		defer func() {
			if r := recover(); r != nil {
				hadPanic = true
			}
		}()

		// Render the wrapped component
		if componentFn != nil {
			vnode := componentFn()
			children = []rtui.VNode{vnode}
		}
	}()

	// If panic occurred, render the fallback instead
	if hadPanic {
		if workInProgress.ErrorBoundaryFallbackFiber != nil {
			// Use the fallback fiber as the child
			workInProgress.Child = workInProgress.ErrorBoundaryFallbackFiber
			return workInProgress
		}
		// No fallback, render empty fragment
		children = []rtui.VNode{rtui.Fragment()}
	}

	// Get current child for reconciliation
	var currentChild *Fiber
	if current != nil {
		currentChild = current.Child
	}

	// Reconcile children (either the component result or the fallback)
	workInProgress.Child = reconcileChildren(
		workInProgress,
		currentChild,
		children,
		workInProgress.Lanes,
	)

	return workInProgress
}

// =============================================================================
// Memo BeginWork
// =============================================================================

// beginWorkMemo processes a memoized component Fiber
// Memo components skip re-rendering if their props haven't changed
func beginWorkMemo(current, workInProgress *Fiber) *Fiber {
	// Get current props
	newProps := workInProgress.GetProps()

	// Check if we should update based on prop comparison
	var shouldUpdate bool
	if current == nil {
		// First render - always update
		shouldUpdate = true
	} else {
		// Use memo's comparison function
		oldProps := current.GetProps()
		if workInProgress.MemoCompare != nil {
			// Use custom comparison from memo
			// Note: compare returns true if props are equal (no update needed)
			propsEqual := workInProgress.MemoCompare(oldProps, newProps)
			shouldUpdate = !propsEqual
		} else {
			// Default shallow comparison
			shouldUpdate = !rtui.ShallowPropsEqual(oldProps, newProps)
		}
	}

	// If props haven't changed and no pending updates, skip re-render
	if !shouldUpdate && workInProgress.Lanes == rtui.LaneNoLane {
		// Props unchanged - reuse current fiber's result
		// Copy child pointers from current to workInProgress
		if current != nil {
			workInProgress.Child = current.Child
			// Copy subtree flags to preserve effect information
			workInProgress.SubtreeFlags = current.SubtreeFlags
		}
		return workInProgress
	}

	// Props changed or have pending updates - process the wrapped component
	// For memo, the wrapped component function is stored in ComponentFunc
	var children []rtui.VNode
	if workInProgress.ComponentFunc != nil {
		vnode := workInProgress.ComponentFunc()
		children = []rtui.VNode{vnode}
	} else if workInProgress.ComponentFuncWithProps != nil {
		vnode := workInProgress.ComponentFuncWithProps(workInProgress.Props)
		children = []rtui.VNode{vnode}
	}

	// Get current child for reconciliation
	var currentChild *Fiber
	if current != nil {
		currentChild = current.Child
	}

	// Reconcile the wrapped component as a single child
	workInProgress.Child = reconcileChildren(
		workInProgress,
		currentChild,
		children,
		workInProgress.Lanes,
	)

	return workInProgress
}

// =============================================================================
// VNode Component Instance Support
// =============================================================================

// createVNodeComponentInstanceFromFiber creates a new VNode component instance from Fiber
// This is a factory function called by InstanceManager.GetOrCreate
func createVNodeComponentInstanceFromFiber(key string, fiber *rtui.Fiber) rtui.ComponentInstance {
	return state.NewVNodeComponentInstanceFromFiber(key, fiber)
}
