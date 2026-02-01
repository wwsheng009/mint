// Package ui provides the Fiber Reconciler for incremental rendering.
package ui

import (
	"fmt"
	"os"
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
	isWorking bool // Currently working

	// === Scheduling ===
	deadline   time.Time     // Current frame deadline
	timeBudget time.Duration // Time slice budget per frame

	// === Integration ===
	app                 *framework.App           // Framework app
	instanceMgr         *InstanceManager         // Component instance manager
	interactionStateMgr *InteractionStateManager // Interaction state (hover/focus/etc)
	keyValidator        *KeyValidator            // Key validation
	rootComponent       ComponentFunc            // Root component function
	ctx                 *ComponentContext        // Root component context

	// === Render State ===
	buffer         *paint.Buffer // Render target
	paintCtx       component.PaintContext
	renderCallback RenderFunc // Callback for rendering VNodes

	// === Configuration ===
	enableFiber bool // Use Fiber reconciliation (env controlled)
}

// ReconcilerConfig configures the reconciler
type ReconcilerConfig struct {
	TimeBudget      time.Duration // Time slice budget
	EnableProfiling bool          // Enable performance profiling
	EnableFiber     bool          // Enable Fiber reconciliation
}

// NewReconciler creates a new reconciler
func NewReconciler(app *framework.App, rootComponent ComponentFunc, config ReconcilerConfig) *Reconciler {
	timeBudget := config.TimeBudget
	if timeBudget == 0 {
		timeBudget = 5 * time.Millisecond // Default 5ms budget
	}

	return &Reconciler{
		app:                 app,
		rootComponent:       rootComponent,
		instanceMgr:         NewInstanceManager(),
		interactionStateMgr: NewInteractionStateManager(),
		keyValidator:        NewKeyValidator(),
		timeBudget:          timeBudget,
		ctx:                 NewComponentContextForRoot(),
		enableFiber:         config.EnableFiber,
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
// IMPORTANT: We do NOT call renderFunc() here directly.
// Instead, we wrap it as a ComponentVNode so that beginWorkComponent
// will handle the actual component invocation with proper Context management.
// This ensures all hooks use the same ComponentInstance's context.
func (r *Reconciler) prepareFreshStack(renderFunc func() VNode) {
	// Wrap the root component as a ComponentVNode
	// This ensures it goes through beginWorkComponent which manages Context properly
	rootComponentVNode := NewComponent("RootComponent", renderFunc)
	rootComponentVNode.SetKey("root")

	// Create or update Fiber tree
	if r.root == nil {
		// First render - create new tree from the wrapped component
		r.root = CreateFiberFromVNode(rootComponentVNode)
		r.workInProgress = r.root
	} else {
		// Subsequent render - create work-in-progress tree
		r.workInProgress = r.createWorkInProgress(r.root, rootComponentVNode)
	}
}

// workLoopSync processes the work loop synchronously
// In Phase 1, this is a simple synchronous implementation
// Phase 3 will add time slicing
func (r *Reconciler) workLoopSync() {
	if r.workInProgress == nil {
		return
	}

	// Set current reconciler for BeginWork to access InstanceManager
	currentReconciler = r
	defer func() { currentReconciler = nil }()

	// Process all work units using correct Fiber traversal
	// The traversal follows: BeginWork down the tree, then CompleteWork back up
	r.performUnitOfWork(r.workInProgress)

	// CRITICAL: Swap workInProgress tree with root tree (double buffering)
	// After the work loop completes, workInProgress becomes the new current tree
	// This is the core of React Fiber's double buffering architecture
	r.root = r.workInProgress
	r.workInProgress = nil
}

// performUnitOfWork processes a single fiber and its subtree
func (r *Reconciler) performUnitOfWork(unitOfWork *Fiber) {
	if unitOfWork == nil {
		return
	}

	// BeginWork: process this fiber and create children
	next := BeginWork(unitOfWork.Alternate, unitOfWork)

	// If BeginWork returned a child, process it first (depth-first)
	if next != nil && next.Child != nil {
		r.performUnitOfWork(next.Child)
	}

	// CompleteWork: finalize this fiber
	CompleteWork(unitOfWork.Alternate, unitOfWork)

	// Process siblings
	if unitOfWork.Sibling != nil {
		r.performUnitOfWork(unitOfWork.Sibling)
	}
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
	// r.root now contains the updated tree from workInProgress (swapped in workLoopSync)
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

	// Debug: log fiber traversal
	if os.Getenv("TUI_DEBUG_UI") == "true" {
		fmt.Fprintf(os.Stderr, "renderFiberToBuffer: type=%T, x=%d, y=%d\n", fiber.VNode, x, y)
	}

	// Render this fiber based on its type
	r.renderFiber(fiber, x, y, buffer)

	// Render children with proper layout
	childX := x
	childY := y

	// Check if this is a LayoutNode to handle horizontal layout
	isHStack := false
	var gap int
	if layoutNode, ok := fiber.VNode.(*LayoutNode); ok {
		isHStack = layoutNode.direction == 0 // DirectionRow = 0
		gap = layoutNode.Gap()
	}

	for child := fiber.Child; child != nil; child = child.Sibling {
		r.renderFiberToBuffer(child, childX, childY, buffer)

		// Move position for next sibling
		if isHStack {
			// Horizontal layout: move X, keep Y
			width := r.measureFiberWidth(child)
			childX += width + gap
		} else {
			// Vertical layout: move Y, keep X
			offsetY := r.measureFiberHeight(child)
			childY += offsetY
		}
	}
}

// measureFiberWidth measures the width of a fiber node
func (r *Reconciler) measureFiberWidth(fiber *Fiber) int {
	if fiber == nil || fiber.VNode == nil {
		return 0
	}

	switch v := fiber.VNode.(type) {
	case *TextVNode:
		return len(v.Content())
	case *ButtonVNode:
		return len(v.Label()) + 2 // [label]
	case *InputVNode:
		return 22 // [ + 20 chars + ]
	case *SelectVNode:
		maxLen := 10
		for _, opt := range v.Options() {
			if len(opt.Label) > maxLen {
				maxLen = len(opt.Label)
			}
		}
		return maxLen + 4 // [label + ▼]
	case *CheckboxVNode:
		width := 4 // [X] or [ ]
		if v.Label() != "" {
			width += 1 + len(v.Label())
		}
		return width
	case *ElementVNode, *LayoutNode, *FragmentVNode:
		// Containers - sum up children widths
		width := 0
		for child := fiber.Child; child != nil; child = child.Sibling {
			width += r.measureFiberWidth(child)
		}
		return width
	default:
		return 10
	}
}

// renderFiber renders a single Fiber to the buffer
func (r *Reconciler) renderFiber(fiber *Fiber, x, y int, buffer *paint.Buffer) {
	if fiber == nil || fiber.VNode == nil {
		return
	}

	// Skip ComponentVNode - its children are already expanded in the Fiber tree
	// The renderCallback should only be called for rendered nodes, not component definitions
	if fiber.VNode.Type() == VNodeComponent {
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

// GetInteractionStateManager returns the interaction state manager
func (r *Reconciler) GetInteractionStateManager() *InteractionStateManager {
	return r.interactionStateMgr
}

// GetKeyValidator returns the key validator
func (r *Reconciler) GetKeyValidator() *KeyValidator {
	return r.keyValidator
}

// GetContext returns the root component context
func (r *Reconciler) GetContext() *ComponentContext {
	return r.ctx
}

// SetInstanceManager sets the instance manager (shared with declarativeRoot)
func (r *Reconciler) SetInstanceManager(mgr *InstanceManager) {
	r.instanceMgr = mgr
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
