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
	if log.HitMapLogger.Enabled() {
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
			log.UILogger.Debug("[BeginWork] Called: Type=%d(%s), Key=%q, Tag=%q",
				workInProgress.Type, typeName, workInProgress.Key, workInProgress.Tag)
		}
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
	if log.UILogger.Enabled() {
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
	// Get children using the VNode interface (works for any VNode type)
	// This handles ButtonVNode, TextVNode, and other custom VNode types
	children := workInProgress.VNode.Children()

	// ✨ NEW: Create/reuse VNodeComponentInstance for VNode struct components
	// This enables persistent event handlers and state for Button, Text, etc.

	if currentReconciler != nil && currentReconciler.instanceMgr != nil && workInProgress.Key != "" {
		// ✨ IMPORTANT: Use Fiber.Path instead of Fiber.Key for instance key generation
		// This ensures instance keys match HitMap NodeIDs which use full paths.
		// For user-keyed elements (e.g., button with key="btn-event"):
		//   - Fiber.Key = "btn-event" (user's original key)
		//   - Fiber.Path = "/root/base[0]/.../button[0]/key[btn-event]" (full path)
		// HitMap stores NodeID = Fiber.Path, so instance key must use Path too!
		lookupKey := workInProgress.Path
		if lookupKey == "" {
			// Fallback to Key if Path is not set (shouldn't happen with proper path generation)
			lookupKey = workInProgress.Key
		}
		instanceKey := "vnode:" + lookupKey

		if os.Getenv("TUI_DEBUG_INSTANCE") == "true" || os.Getenv("TUI_DEBUG_HITMAP") == "true" {
			log.UILogger.Debug("[beginWorkElement] Creating instance for key=%s (fiber.Key=%q, fiber.Path=%q)",
				instanceKey, workInProgress.Key, workInProgress.Path)
		}

		// Get or create VNode component instance
		instance := currentReconciler.instanceMgr.GetOrCreate(instanceKey, func() rtui.ComponentInstance {
			return createVNodeComponentInstance(instanceKey, workInProgress.VNode)
		})

		// Update the instance with the new VNode
		if vnodeInst, ok := instance.(*state.VNodeComponentInstance); ok {
			vnodeInst.UpdateVNode(workInProgress.VNode)
		}

		// Store the instance in the fiber
		workInProgress.ComponentInstance = instance

		if os.Getenv("TUI_DEBUG_INSTANCE") == "true" || os.Getenv("TUI_DEBUG_HITMAP") == "true" {
			log.UILogger.Debug("[beginWorkElement] ✅ Created/Updated instance: key=%s, fiber.Key=%q, type=%d",
				instanceKey, workInProgress.Key, workInProgress.Type)
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

// =============================================================================
// VNode Component Instance Support
// =============================================================================

// createVNodeComponentInstance creates a new VNode component instance
// This is a factory function called by InstanceManager.GetOrCreate
func createVNodeComponentInstance(key string, vnode rtui.VNode) rtui.ComponentInstance {
	return state.NewVNodeComponentInstance(key, vnode)
}
