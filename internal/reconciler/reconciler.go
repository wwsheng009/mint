package reconciler

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

import (
	"fmt"
	"os"
	"time"

	"github.com/wwsheng009/mint/framework"
	"github.com/wwsheng009/mint/framework/component"
	"github.com/wwsheng009/mint/internal/state"
	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/paint"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui"
) // Note: ui is still needed for component-specific VNode types (TextVNode, ButtonVNode, etc.)

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
	instanceMgr         *state.InstanceManager         // Component instance manager
	interactionStateMgr *state.InteractionStateManager // Interaction state (hover/focus/etc)
	keyValidator        *state.KeyValidator            // Key validation
	rootComponent       rtui.ComponentFunc             // Root component function
	ctx                 *rtui.ComponentContext         // Root component context

	// === Render State ===
	buffer         *paint.Buffer // Render target
	paintCtx       component.PaintContext
	renderCallback RenderFunc // Callback for rendering VNodes

	// === Layout Integration ===
	vnodeConverter *VNodeConverter         // VNode → runtime.LayoutNode converter
	layoutRoot     *runtime.LayoutNode    // Root of the layout tree
	layoutBoxes    []runtime.LayoutBox     // Layout boxes for hit testing

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
func NewReconciler(app *framework.App, rootComponent rtui.ComponentFunc, config ReconcilerConfig) *Reconciler {
	timeBudget := config.TimeBudget
	if timeBudget == 0 {
		timeBudget = 5 * time.Millisecond // Default 5ms budget
	}

	return &Reconciler{
		app:                 app,
		rootComponent:       rootComponent,
		instanceMgr:         state.NewInstanceManager(),
		interactionStateMgr: state.NewInteractionStateManager(),
		keyValidator:        state.NewKeyValidator(),
		timeBudget:          timeBudget,
		ctx:                 rtui.NewComponentContextForRoot(),
		enableFiber:         config.EnableFiber,
		vnodeConverter:      NewVNodeConverter(),
	}
}

// =============================================================================
// Public API
// =============================================================================

// Render executes the rendering process
// This is the main entry point called from declarativeRoot.Paint
func (r *Reconciler) Render(ctx component.PaintContext, buffer *paint.Buffer, renderFunc func() rtui.VNode) {
	// Note: renderFunc returns ui.VNode (VNode interface is from ui package)
	// This is correct as VNode implementations are in ui package
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
func (r *Reconciler) prepareFreshStack(renderFunc func() rtui.VNode) {
	// Wrap the root component as a ComponentVNode
	// This ensures it goes through beginWorkComponent which manages Context properly
	rootComponentVNode := rtui.NewComponent("RootComponent", renderFunc)
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
func (r *Reconciler) createWorkInProgress(current *Fiber, vnode rtui.VNode) *Fiber {
	// Note: vnode is ui.VNode - VNode interface and implementations are from ui package
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

	// Phase 1: Build layout tree from Fiber tree
	// Convert VNode tree to runtime.LayoutNode tree
	r.layoutRoot = r.buildLayoutTree(r.root)

	// Phase 2: Calculate layout
	// Use the runtime flex layout to calculate positions
	r.calculateLayout(r.buffer.Width, r.buffer.Height)

	// Phase 3: Generate LayoutBoxes for hit testing
	if r.layoutRoot != nil {
		r.layoutBoxes = r.vnodeConverter.GenerateLayoutBoxes(r.layoutRoot)
	}

	// Phase 4: Render the Fiber tree to buffer
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
	if layoutNode, ok := fiber.VNode.(*ui.LayoutNode); ok {
		isHStack = layoutNode.Direction() == 0 // DirectionRow = 0
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
	case *ui.TextVNode:
		return len(v.Content())
	case *ui.ButtonVNode:
		return len(v.Label()) + 2 // [label]
	case *ui.InputVNode:
		return 22 // [ + 20 chars + ]
	case *ui.SelectVNode:
		maxLen := 10
		for _, opt := range v.Options() {
			if len(opt.Label) > maxLen {
				maxLen = len(opt.Label)
			}
		}
		return maxLen + 4 // [label + ▼]
	case *ui.CheckboxVNode:
		width := 4 // [X] or [ ]
		if v.Label() != "" {
			width += 1 + len(v.Label())
		}
		return width
	case *ui.ElementVNode, *ui.LayoutNode, *ui.FragmentVNode:
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
	if fiber.VNode.Type() == 	rtui.VNodeComponent {
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
type RenderFunc func(vnode rtui.VNode, x, y int, buffer *paint.Buffer)
// Note: vnode is ui.VNode - VNode interface and implementations are from ui package

// SetRenderCallback sets the render callback
func (r *Reconciler) SetRenderCallback(cb RenderFunc) {
	r.renderCallback = cb
}

// =============================================================================
// Layout Tree Integration
// =============================================================================

// buildLayoutTree builds a runtime.LayoutNode tree from a Fiber tree
func (r *Reconciler) buildLayoutTree(fiber *Fiber) *runtime.LayoutNode {
	if fiber == nil {
		return nil
	}

	// Convert the root VNode to LayoutNode
	// The VNodeConverter will recursively convert children
	return r.vnodeConverter.Convert(fiber.VNode)
}

// calculateLayout calculates layout for the layout tree using runtime flex layout
func (r *Reconciler) calculateLayout(width, height int) {
	if r.layoutRoot == nil {
		return
	}

	// Create constraints for the root node
	constraints := runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth:  width,
		MinHeight: 0,
		MaxHeight: height,
	}

	// Measure and layout each node
	r.measureAndLayoutNode(r.layoutRoot, constraints)
}

// measureAndLayoutNode measures and layouts a single node and its children
func (r *Reconciler) measureAndLayoutNode(node *runtime.LayoutNode, constraints runtime.BoxConstraints) {
	if node == nil {
		return
	}

	// Measure this node's size
	size := r.measureNode(node, constraints)
	node.MeasuredWidth = size.Width
	node.MeasuredHeight = size.Height

	// Layout children based on direction
	if len(node.Children) > 0 {
		r.layoutChildren(node, constraints)
	}
}

// measureNode measures a single node's size
func (r *Reconciler) measureNode(node *runtime.LayoutNode, constraints runtime.BoxConstraints) runtime.Size {
	// If node has a component, try to measure it
	if node.Component != nil {
		if componentSize := node.Component.Measure(constraints); componentSize.Width > 0 || componentSize.Height > 0 {
			return componentSize
		}
	}

	// Default measurement based on style
	width := node.Style.Width
	height := node.Style.Height

	// Handle auto sizing (-1)
	if width < 0 {
		width = constraints.MinWidth
		if width <= 0 {
			width = 10 // Minimum default width
		}
	}
	if height < 0 {
		height = constraints.MinHeight
		if height <= 0 {
			height = 1 // Minimum default height
		}
	}

	// Apply constraints
	if width < constraints.MinWidth {
		width = constraints.MinWidth
	}
	if width > constraints.MaxWidth && constraints.MaxWidth > 0 {
		width = constraints.MaxWidth
	}
	if height < constraints.MinHeight {
		height = constraints.MinHeight
	}
	if height > constraints.MaxHeight && constraints.MaxHeight > 0 {
		height = constraints.MaxHeight
	}

	return runtime.Size{Width: width, Height: height}
}

// layoutChildren layouts children based on the node's direction
func (r *Reconciler) layoutChildren(parent *runtime.LayoutNode, constraints runtime.BoxConstraints) {
	if parent == nil || len(parent.Children) == 0 {
		return
	}

	// Create child constraints
	childConstraints := constraints
	if parent.Style.Direction == runtime.DirectionRow {
		// For row: split width among children
		childConstraints.MaxWidth = constraints.MaxWidth
		childConstraints.MaxHeight = parent.MeasuredHeight
	} else {
		// For column: split height among children
		childConstraints.MaxWidth = parent.MeasuredWidth
		childConstraints.MaxHeight = constraints.MaxHeight
	}

	// Calculate child positions
	x := parent.Style.Padding.Left
	y := parent.Style.Padding.Top

	for _, child := range parent.Children {
		// Measure child
		r.measureAndLayoutNode(child, childConstraints)

		// Set position
		child.X = x
		child.Y = y

		// Move to next position
		if parent.Style.Direction == runtime.DirectionRow {
			x += child.MeasuredWidth + parent.Style.Gap
		} else {
			y += child.MeasuredHeight + parent.Style.Gap
		}
	}
}

// GetLayoutBoxes returns the calculated layout boxes for hit testing
func (r *Reconciler) GetLayoutBoxes() []runtime.LayoutBox {
	return r.layoutBoxes
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
func (r *Reconciler) GetInstanceManager() *state.InstanceManager {
	return r.instanceMgr
}

// GetInteractionStateManager returns the interaction state manager
func (r *Reconciler) GetInteractionStateManager() *state.InteractionStateManager {
	return r.interactionStateMgr
}

// GetKeyValidator returns the key validator
func (r *Reconciler) GetKeyValidator() *state.KeyValidator {
	return r.keyValidator
}

// GetContext returns the root component context
func (r *Reconciler) GetContext() *rtui.ComponentContext {
	return r.ctx
}

// SetInstanceManager sets the instance manager (shared with declarativeRoot)
func (r *Reconciler) SetInstanceManager(mgr *state.InstanceManager) {
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

// =============================================================================
// Global Current Reconciler (for BeginWork/CompleteWork access)
// =============================================================================

// currentReconciler holds the currently executing reconciler
var currentReconciler *Reconciler

// GetCurrentReconciler returns the current reconciler
func GetCurrentReconciler() *Reconciler {
	return currentReconciler
}
