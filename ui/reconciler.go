// Package ui provides the Fiber Reconciler for incremental rendering.
package ui

import (
	"time"

	"github.com/wwsheng009/mint/framework"
	"github.com/wwsheng009/mint/framework/component"
	"github.com/wwsheng009/mint/runtime/paint"
)

// =============================================================================
// Fiber Reconciler
// =============================================================================
// Reconciler manages the Fiber tree and the reconciliation process.
// It's responsible for:
// - Scheduling updates with priority (lanes)
// - Executing the work loop (render phase)
// - Committing changes to the buffer (commit phase)
// - Time slicing for interruptible rendering
// =============================================================================

// Reconciler manages Fiber reconciliation
type Reconciler struct {
	// === Fiber Trees ===
	root           *Fiber // Current committed tree
	workInProgress *Fiber // Work-in-progress tree

	// === State ===
	lanes     Lane // Pending lanes (work to do)
	isWorking bool  // Currently working

	// === Scheduling ===
	deadline   time.Time     // Current frame deadline
	timeBudget time.Duration // Time slice budget per frame

	// === Integration ===
	app          *framework.App   // Framework app
	instanceMgr  *InstanceManager // Component instance manager
	rootComponent ComponentFunc    // Root component function
	ctx          *ComponentContext // Root component context

	// === Render State ===
	buffer        *paint.Buffer           // Render target
	paintCtx      component.PaintContext
	renderCallback RenderFunc              // Callback for rendering VNodes

	// === Configuration ===
	enableFiber bool // Use Fiber reconciliation (env controlled)
}

// ReconcilerConfig configures the reconciler
type ReconcilerConfig struct {
	TimeBudget      time.Duration // Time slice budget
	EnableProfiling bool           // Enable performance profiling
	EnableFiber     bool           // Enable Fiber reconciliation
}

// NewReconciler creates a new reconciler
func NewReconciler(app *framework.App, rootComponent ComponentFunc, config ReconcilerConfig) *Reconciler {
	timeBudget := config.TimeBudget
	if timeBudget == 0 {
		timeBudget = 5 * time.Millisecond // Default 5ms budget
	}

	return &Reconciler{
		app:           app,
		rootComponent: rootComponent,
		instanceMgr:   NewInstanceManager(),
		timeBudget:    timeBudget,
		ctx:           NewComponentContextForRoot(),
		enableFiber:   config.EnableFiber,
	}
}

// =============================================================================
// Public API
// =============================================================================

// Render executes the rendering process
// This is the main entry point called from declarativeRoot.Paint
func (r *Reconciler) Render(ctx component.PaintContext, buffer *paint.Buffer, renderFunc func() VNode) {
	if !r.enableFiber {
		return // Fiber not enabled, use legacy rendering
	}

	r.buffer = buffer
	r.paintCtx = ctx

	// Phase 1: Create or update Fiber tree from VNode
	r.prepareFreshStack(renderFunc)

	// Phase 2: Process work (render phase)
	r.workLoopSync()

	// Phase 3: Commit changes
	r.CommitRoot()
}

// ScheduleUpdate schedules a state update
func (r *Reconciler) ScheduleUpdate(lane Lane) {
	r.lanes = MergeLanes(r.lanes, lane)
	r.requestWork()
}

// MarkDirty marks the reconciler as needing work
func (r *Reconciler) MarkDirty() {
	r.requestWork()
}

// =============================================================================
// Work Loop
// =============================================================================

// prepareFreshStack prepares a fresh Fiber stack for rendering
func (r *Reconciler) prepareFreshStack(renderFunc func() VNode) {
	// Reset hook context for root
	r.ctx.ResetContext()
	SetCurrentContext(r.ctx)

	// Call root component to get VNode
	vnode := renderFunc()

	// Clear current context
	SetCurrentContext(nil)

	// Create or update Fiber tree
	if r.root == nil {
		// First render - create new tree
		r.root = CreateFiberFromVNode(vnode)
		r.workInProgress = r.root
	} else {
		// Subsequent render - create work-in-progress tree
		r.workInProgress = r.createWorkInProgress(r.root, vnode)
	}
}

// workLoopSync processes the work loop synchronously
// In Phase 1, this is a simple synchronous implementation
// Phase 3 will add time slicing
func (r *Reconciler) workLoopSync() {
	if r.workInProgress == nil {
		return
	}

	// Process all work units
	workInProgress := r.workInProgress

	for workInProgress != nil {
		// BeginWork: reconcile and create children
		workInProgress = BeginWork(nil, workInProgress)

		// CompleteWork: finalize and collect effects
		if workInProgress != nil {
			workInProgress = CompleteWork(nil, workInProgress)
		}

		// Move to next work unit
		workInProgress = r.getNextWorkUnit(workInProgress)
	}

	// Work complete, prepare for commit
	r.workInProgress = nil
}

// createWorkInProgress creates a work-in-progress fiber
func (r *Reconciler) createWorkInProgress(current *Fiber, vnode VNode) *Fiber {
	if current == nil {
		return CreateFiberFromVNode(vnode)
	}

	// Clone the current fiber for work-in-progress
	work := CloneFiber(current)
	work.VNode = vnode
	work.Props = vnode.Props()
	work.Lanes = LaneNoLane
	work.Flags = EffectNoEffect

	// Link to alternate (double buffering)
	work.Alternate = current
	if current.Alternate != nil {
		current.Alternate.Alternate = nil // Break the old link
	}

	return work
}

// getNextWorkUnit returns the next work unit using depth-first traversal
func (r *Reconciler) getNextWorkUnit(current *Fiber) *Fiber {
	if current == nil {
		return nil
	}

	// Depth-first: child -> sibling -> return
	if current.Child != nil {
		return current.Child
	}

	// No child, look for sibling
	next := current.Sibling
	for next == nil && current.Return != nil {
		// No sibling, go up and look for uncle
		current = current.Return
		next = current.Sibling
	}

	return next
}

// =============================================================================
// Commit Phase
// =============================================================================

// CommitRoot commits all changes to the buffer
func (r *Reconciler) CommitRoot() {
	if r.root == nil {
		return
	}

	// Render the Fiber tree to buffer
	r.renderFiberToBuffer(r.root, 0, 0, r.buffer)

	// Validate hooks finished correctly
	if err := r.ctx.FinishRender(); err != nil {
		return
	}

	// Run effects after render
	r.ctx.RunEffects()

	// Cleanup unused component instances
	// activeKeys are collected during the render phase
}

// renderFiberToBuffer renders a Fiber tree to the buffer
func (r *Reconciler) renderFiberToBuffer(fiber *Fiber, x, y int, buffer *paint.Buffer) {
	if fiber == nil {
		return
	}

	// Render this fiber based on its type
	r.renderFiber(fiber, x, y, buffer)

	// Render children
	childX := x
	childY := y
	for child := fiber.Child; child != nil; child = child.Sibling {
		r.renderFiberToBuffer(child, childX, childY, buffer)
		// Move position for next sibling (vertical layout)
		offsetY := r.measureFiberHeight(child)
		childY += offsetY
	}
}

// renderFiber renders a single Fiber to the buffer
func (r *Reconciler) renderFiber(fiber *Fiber, x, y int, buffer *paint.Buffer) {
	if fiber == nil || fiber.VNode == nil {
		return
	}

	// Use the render callback if set
	if r.renderCallback != nil {
		r.renderCallback(fiber.VNode, x, y, buffer)
	}
}

// measureFiberHeight measures the height of a fiber
func (r *Reconciler) measureFiberHeight(fiber *Fiber) int {
	if fiber == nil || fiber.VNode == nil {
		return 0
	}

	// Simple height measurement
	// TODO: Implement proper height calculation
	return 1
}

// RenderFunc is a function to render a VNode to the buffer
type RenderFunc func(vnode VNode, x, y int, buffer *paint.Buffer)

// SetRenderCallback sets the render callback
func (r *Reconciler) SetRenderCallback(cb RenderFunc) {
	r.renderCallback = cb
}

// =============================================================================
// Scheduling
// =============================================================================

// requestWork requests the framework to schedule a frame
func (r *Reconciler) requestWork() {
	if r.app != nil {
		r.app.MarkDirty()
	}
}

// hasMoreWork checks if there's more work to do
func (r *Reconciler) hasMoreWork() bool {
	return r.workInProgress != nil || r.lanes != LaneNoLane
}

// =============================================================================
// Instance Management
// =============================================================================

// GetInstanceManager returns the instance manager
func (r *Reconciler) GetInstanceManager() *InstanceManager {
	return r.instanceMgr
}

// GetContext returns the root component context
func (r *Reconciler) GetContext() *ComponentContext {
	return r.ctx
}

// =============================================================================
// Debug / Profiling
// =============================================================================

// Stats returns reconciler statistics
func (r *Reconciler) Stats() map[string]interface{} {
	return map[string]interface{}{
		"hasWork":    r.hasMoreWork(),
		"lanes":      r.lanes,
		"isWorking":  r.isWorking,
		"fiberCount": CountFibers(r.root),
		"instances":  r.instanceMgr.Count(),
	}
}
