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
	"fmt"
	"os"
	"runtime/debug"

	"github.com/wwsheng009/mint/internal/log"
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

	// Check for ErrorBoundary - handle before regular component processing
	if boundary, ok := workInProgress.VNode.(*rtui.ErrorBoundaryVNode); ok {
		return beginWorkErrorBoundary(current, workInProgress, boundary)
	}

	// Check for Memo - handle memoization to skip unnecessary renders
	if memo, ok := workInProgress.VNode.(*rtui.MemoVNode); ok {
		return beginWorkMemo(current, workInProgress, memo)
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

	// Debug: verify context is set
	if os.Getenv("TUI_DEBUG_UI") == "true" {
		log.UILogger.Debug("beginWorkComponent: SetCurrentContext(ctx=%p, ComponentID=%s), GetCurrentContext()=%p\n",
			ctx, ctx.ComponentID, rtui.GetCurrentContext())
	}

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

// =============================================================================
// ErrorBoundary BeginWork
// =============================================================================

// beginWorkErrorBoundary processes an error boundary Fiber
// Error boundaries catch panics from their child components and render a fallback UI
func beginWorkErrorBoundary(current, workInProgress *Fiber, boundary *rtui.ErrorBoundaryVNode) *Fiber {
	// The error boundary wraps a component that might panic
	// We need to call the component's function with panic recovery

	// Get the component function from the boundary
	componentFn := boundary.Component()

	// Try to render the component with panic recovery
	var children []rtui.VNode
	var hadPanic bool

	func() {
		defer func() {
			if r := recover(); r != nil {
				hadPanic = true

				// Update the boundary's error state
				boundary.SetError(r.(error), fmt.Sprintf("panic in %s: %v", boundary.Name(), r), string(debug.Stack()))
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
		fallback := boundary.Fallback()
		if fallback != nil {
			children = []rtui.VNode{fallback}
		} else {
			// No fallback, render empty fragment
			children = []rtui.VNode{rtui.Fragment()}
		}
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
func beginWorkMemo(current, workInProgress *Fiber, memo *rtui.MemoVNode) *Fiber {
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
		compare := memo.GetCompare()
		if compare != nil {
			// Use custom comparison from memo
			// Note: compare returns true if props are equal (no update needed)
			propsEqual := compare(oldProps, newProps)
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
	wrappedComponent := memo.GetComponent()

	// Get current child for reconciliation
	var currentChild *Fiber
	if current != nil {
		currentChild = current.Child
	}

	// Reconcile the wrapped component as a single child
	children := []rtui.VNode{wrappedComponent}
	workInProgress.Child = reconcileChildren(
		workInProgress,
		currentChild,
		children,
		workInProgress.Lanes,
	)

	return workInProgress
}
