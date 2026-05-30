package render

import (
	"github.com/wwsheng009/mint/framework/component"
	"github.com/wwsheng009/mint/internal/reconciler"
	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/runtime/paint"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Fiber Reconciler Integration
// =============================================================================
// These functions create and configure the Fiber reconciler.

// fiberReconcilerAdapter adapts internal/reconciler.Reconciler to rtui.Reconciler interface
type fiberReconcilerAdapter struct {
	r *reconciler.Reconciler
}

// newFiberReconciler creates a new Fiber reconciler for the given scheduler, render function and root context
func newFiberReconciler(scheduler reconciler.Scheduler, fn rtui.ComponentFunc, rootCtx *rtui.ComponentContext) rtui.Reconciler {
	// Create the actual reconciler from internal/reconciler
	r := reconciler.NewReconciler(scheduler, fn, reconciler.ReconcilerConfig{})

	// Set the root component context for global state management
	// This ensures Intent Handlers update the same state that App() reads
	r.SetRootContext(rootCtx)

	// Set up the render callback to render Fibers to the buffer
	r.SetRenderCallback(func(fiber *reconciler.Fiber, x, y int, buffer *paint.Buffer) {
		renderFiberToBuffer(fiber, x, y, buffer)
	})

	// Wrap in adapter to satisfy rtui.Reconciler interface
	return &fiberReconcilerAdapter{r: r}
}

// Render executes the rendering process (adapter method with interface{} parameters)
func (a *fiberReconcilerAdapter) Render(ctx interface{}, buffer interface{}, renderFunc func() rtui.VNode) {
	// Type assert to concrete types
	paintCtx, ok := ctx.(component.PaintContext)
	paintBuffer, ok := buffer.(*paint.Buffer)
	if !ok || paintBuffer == nil {
		return
	}

	// Call the actual reconciler's Render method
	a.r.Render(paintCtx, paintBuffer, renderFunc)
}

// SetApp sets the framework app (adapter method)
// Note: This is kept for backward compatibility, the actual storage uses Scheduler interface
func (a *fiberReconcilerAdapter) SetApp(app interface{}) {
	if scheduler, ok := app.(reconciler.Scheduler); ok {
		a.r.SetScheduler(scheduler)
	}
}

// SetFocusManager sets the focus manager (adapter method)
// Supports both FiberFocusManager (Fiber-first) and VNodeFocusManager (legacy)
func (a *fiberReconcilerAdapter) SetFocusManager(mgr interface{}) {
	switch m := mgr.(type) {
	case *rtui.FiberFocusManager:
		a.r.SetFocusManager(m)
	case *rtui.VNodeFocusManager:
		// Legacy: convert to FiberFocusManager is not possible, so we skip
		// This should not happen in normal Fiber mode
	}
}

// SetFiberTarget sets the receiver for the committed Fiber root.
// Phase 8: This allows reconciler to call SetFiber for NodeID propagation.
func (a *fiberReconcilerAdapter) SetFiberTarget(target interface{ SetFiber(*reconciler.Fiber) }) {
	a.r.SetFiberTarget(target)
}

// GetRenderedRoot returns the rendered VNode tree (adapter method)
func (a *fiberReconcilerAdapter) GetRenderedRoot() rtui.VNode {
	return a.r.GetRenderedRoot()
}

// GetInstanceMgr returns the InstanceManager from the Fiber reconciler
func (a *fiberReconcilerAdapter) GetInstanceMgr() interface{} {
	return a.r.GetInstanceManager()
}

// GetAllInteractionInstances returns all instances that implement interaction interfaces
// This is used by App to register instances with InteractionContext
func (a *fiberReconcilerAdapter) GetAllInteractionInstances() map[int]interface{} {
	instanceMgr := a.r.GetInstanceManager()
	if instanceMgr == nil {
		return nil
	}

	// Get all instances by ID
	allInstances := instanceMgr.GetAllInstancesByID()

	// Filter for instances that implement interaction interfaces
	result := make(map[int]interface{})
	for nodeID, inst := range allInstances {
		// Check if instance implements ResetPressed() - this is the primary marker
		// for controls that have pressed state (Button with PressableBehavior, etc.)
		if _, ok := interface{}(inst).(interface{ ResetPressed() }); ok {
			result[int(nodeID)] = inst
			continue
		}

		// Check if instance implements HandleAction() - ActionHandlerInstance
		// This includes all interactive components (Button, Checkbox, Input, Select, etc.)
		if _, ok := interface{}(inst).(interface{ HandleAction(*action.Action) bool }); ok {
			result[int(nodeID)] = inst
		}
	}

	return result
}

// GetFiberRoot returns Fiber root from the underlying reconciler

// Phase 8: Fiber to Layout Engine NodeID propagation
func (a *fiberReconcilerAdapter) GetFiberRoot() *reconciler.Fiber {
	return a.r.GetFiberRoot()
}

func (a *fiberReconcilerAdapter) Dispose() {
	if a == nil || a.r == nil {
		return
	}
	a.r.Dispose()
}
