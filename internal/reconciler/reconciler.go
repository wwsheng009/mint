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
	"os"
	"time"

	"github.com/wwsheng009/mint/framework"
	"github.com/wwsheng009/mint/framework/component"
	"github.com/wwsheng009/mint/internal/log"
	"github.com/wwsheng009/mint/internal/state"
	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/paint"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// Reconciler manages Fiber reconciliation
type Reconciler struct {
	// === Fiber Trees ===
	root           *Fiber // Current committed tree
	workInProgress *Fiber // Work-in-progress tree

	// === State ===
	lanes     Lane // Pending lanes (work to do)
	isWorking bool // Currently working

	// === Scheduling ===
	timeBudget time.Duration // Time slice budget per frame

	// === Integration ===
	app                 *framework.App                 // Framework app
	instanceMgr         *state.InstanceManager         // Component instance manager
	interactionStateMgr *state.InteractionStateManager // Interaction state (hover/focus/etc)
	keyValidator        *state.KeyValidator            // Key validation
	rootComponent       rtui.ComponentFunc             // Root component function
	ctx                 *rtui.ComponentContext         // Root component context
	focusMgr            *rtui.FiberFocusManager        // Focus manager for keyboard navigation (Fiber-first)

	// === Render State ===
	buffer         *paint.Buffer // Render target
	paintCtx       component.PaintContext
	renderCallback RenderFunc // Callback for rendering VNodes
	renderedRoot   rtui.VNode // The rendered VNode tree (for focus management, etc.)

	// === Layout Integration ===
	vnodeConverter *VNodeConverter     // VNode → runtime.LayoutNode converter
	layoutRoot     *runtime.LayoutNode // Root of the layout tree
	layoutBoxes    []runtime.LayoutBox // Layout boxes for hit testing

	// === Path Generation ===
	pathGenerator *PathGenerator // Automatic path key generator for static UI

	// === Renderer ===
	renderer rtui.VNodeRenderer // Renderer for SetFiber call

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
		pathGenerator:       NewPathGenerator(), // ✨ Initialize path generator
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
		log.FiberLogger.Debug("[Reconciler.Render] ⚠️  Fiber NOT enabled! enableFiber=%v", r.enableFiber)
		return // Fiber not enabled, use legacy rendering
	}

	log.FiberLogger.Debug("[Reconciler.Render] ✅ Fiber enabled, starting render...")
	r.buffer = buffer
	r.paintCtx = ctx

	// Phase 1: Create or update Fiber tree from VNode
	r.prepareFreshStack(renderFunc)

	// Phase 2: Process work (render phase)
	r.workLoopSync()

	// Phase 3: Commit changes
	r.CommitRoot()

	// Store the rendered VNode tree for focus management and other purposes
	// The root is a ComponentVNode, its children contain the actual rendered content
	r.updateRenderedRoot()
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
		// ✨ Set root Fiber's Path for proper path generation in children
		// Children will inherit this path: /root + /segment = /root/segment
		r.root.Path = "/root"
		r.root.Key = "root"
		r.workInProgress = r.root
	} else {
		// Subsequent render - create work-in-progress tree
		r.workInProgress = r.createWorkInProgress(r.root, rootComponentVNode)
		// Ensure workInProgress also has the correct Path
		r.workInProgress.Path = "/root"
	}
}

// workLoopSync processes the work loop synchronously
// In Phase 1, this is a simple synchronous implementation
// Phase 3 will add time slicing
func (r *Reconciler) workLoopSync() {
	if r.workInProgress == nil {
		if log.HitMapLogger.Enabled() {
			log.FiberLogger.Debug("[workLoopSync] ⚠️  workInProgress is nil!")
		}
		return
	}

	// Set current reconciler BEFORE processing work units
	// This ensures BeginWork can access InstanceManager for all fibers
	currentReconciler = r
	defer func() { currentReconciler = nil }()

	// ✨ Set path generator for automatic key generation
	pathGenerator = r.pathGenerator

	if os.Getenv("TUI_DEBUG_HITMAP") == "true" {
		log.FiberLogger.Debug("[workLoopSync] Starting work loop...")
	}

	// Process all work units using correct Fiber traversal
	// The traversal follows: BeginWork down the tree, then CompleteWork back up
	// performUnitOfWork recursively processes the entire tree
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

	if os.Getenv("TUI_DEBUG_HITMAP") == "true" {
		typeName := "UNKNOWN"
		switch unitOfWork.Type {
		case rtui.VNodeComponent:
			typeName = "VNodeComponent"
		case rtui.VNodeText:
			typeName = "VNodeText"
		case rtui.VNodeElement:
			typeName = "VNodeElement"
		case rtui.VNodeFragment:
			typeName = "VNodeFragment"
		}
		log.FiberLogger.Debug("[performUnitOfWork] Processing: Type=%d(%s), Key=%q, Tag=%q, hasChild=%v",
			unitOfWork.Type, typeName, unitOfWork.Key, unitOfWork.Tag, unitOfWork.Child != nil)
	}

	// BeginWork: process this fiber and create children
	next := BeginWork(unitOfWork.Alternate, unitOfWork)

	if os.Getenv("TUI_DEBUG_HITMAP") == "true" && next != nil {
		log.FiberLogger.Debug("[performUnitOfWork] After BeginWork: next.Child=%v, next.Sibling=%v",
			next.Child != nil, next.Sibling != nil)
	}

	// If BeginWork returned a child, process it first (depth-first)
	if next != nil && next.Child != nil {
		r.performUnitOfWork(next.Child)
	}

	// CompleteWork: finalize this fiber
	CompleteWork(unitOfWork.Alternate, unitOfWork)

	// Collect child effects - bubble up SubtreeFlags from children
	// This ensures parent fibers know about descendant effects for proper commit
	collectChildEffects(unitOfWork)

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

	// Extract props and ensure children are stored for elements/fragments
	props := vnode.Props()
	if props == nil {
		props = make(rtui.Props)
	}
	children := vnode.Children()
	vnodeType := vnode.Type()
	if vnodeType == rtui.VNodeElement || vnodeType == rtui.VNodeFragment {
		if len(children) > 0 {
			newProps := make(rtui.Props)
			for k, v := range props {
				newProps[k] = v
			}
			newProps["children"] = children
			props = newProps
		}
	}

	work.Props = props
	work.Style = vnode.Style()
	work.Layer = vnode.GetLayer()
	if tagger, ok := vnode.(interface{ Tag() string }); ok {
		work.Tag = tagger.Tag()
	}

	// Update FocusableVNode from new VNode (Fiber-first)
	if f, ok := vnode.(rtui.FocusableVNode); ok && f.IsFocusable() {
		work.FocusableVNode = f
	} else {
		work.FocusableVNode = nil
	}

	// CRITICAL: Update ComponentFunc from new VNode for component types
	// This ensures the latest render function is used
	if vnodeType == rtui.VNodeComponent {
		switch n := vnode.(type) {
		case *rtui.ComponentVNode:
			if n.Fn() != nil {
				work.ComponentFunc = n.Fn()
			}
			if n.FnWithProps() != nil {
				work.ComponentFuncWithProps = n.FnWithProps()
			}
			if n.Name() != "" {
				work.ComponentName = n.Name()
			}
		}
	}

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

	// Phase -1: Commit deletions before any other work
	// This ensures deleted nodes are cleaned up (hooks, refs, etc.) before rendering
	r.commitDeletions(r.root)

	// Phase 0: Apply focus state to Fiber tree before rendering
	// IMPORTANT: We must collect the focusable elements from the NEW Fiber tree first
	// then apply the focus manager's current index to set focus on the right element
	r.applyFocusStateToFiber(r.root)

	// Phase 8: Set Fiber on renderer for NodeID propagation before layout
	// This ensures layout engine has access to Fiber tree for NodeID propagation
	if r.renderer != nil {
		if adapter, ok := r.renderer.(interface{ SetFiber(*Fiber) }); ok {
			adapter.SetFiber(r.root)
		}
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

	// After rendering, update the focus manager with the new Fiber tree's focusable elements
	// This ensures the next render will have the correct focus state
	r.updateFocusManagerFromFiber(r.root)

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

	// Skip ComponentVNode nodes - render their children directly
	if fiber.Type == rtui.VNodeComponent {
		if fiber.Child != nil {
			r.renderFiberToBuffer(fiber.Child, x, y, buffer)
		}
		return
	}

	// Debug: log fiber traversal
	tag := fiber.Tag
	log.RenderLogger.Debug("renderFiberToBuffer tag=%q, x=%d, y=%d", tag, x, y)

	// Render this fiber based on its type
	r.renderFiber(fiber, x, y, buffer)

	// Render children with proper layout
	childX := x
	childY := y

	isHStack := fiber.GetDirection() == rtui.DirectionRow
	gap := fiber.GetGap()

	for child := fiber.Child; child != nil; child = child.Sibling {
		r.renderFiberToBuffer(child, childX, childY, buffer)

		// Move position for next sibling
		if isHStack {
			width := r.measureFiberWidth(child)
			childX += width + gap
		} else {
			offsetY := r.measureFiberHeight(child)
			childY += offsetY
		}
	}
}

// measureFiberWidth measures the width of a fiber node
func (r *Reconciler) measureFiberWidth(fiber *Fiber) int {
	if fiber == nil {
		return 0
	}

	// ComponentVNode - measure the child instead
	if fiber.Type == rtui.VNodeComponent {
		if fiber.Child != nil {
			return r.measureFiberWidth(fiber.Child)
		}
		return 0
	}

	// For text nodes, get content from MemoizedState or Props
	if fiber.Type == rtui.VNodeText {
		// Try to get content from MemoizedState first
		if content, ok := fiber.MemoizedState.(string); ok {
			return paint.StringWidth(content)
		}
		// Fall back to Props
		if fiber.Props != nil {
			if content, ok := fiber.Props["content"].(string); ok {
				return paint.StringWidth(content)
			}
		}
		return 10
	}

	// Get width from ComputedBox if available
	if fiber.ComputedBox != nil {
		if box, ok := fiber.ComputedBox.(interface{ GetWidth() int }); ok {
			return box.GetWidth()
		}
	}

	// Get width from Props
	if fiber.Props != nil {
		if w, ok := fiber.Props["width"].(int); ok && w > 0 {
			return w
		}
	}

	// Default width based on tag
	switch fiber.Tag {
	case "button":
		label := ""
		if fiber.Props != nil {
			if l, ok := fiber.Props["label"].(string); ok {
				label = l
			}
		}
		return paint.StringWidth(label) + 2
	case "input":
		return 22
	case "checkbox":
		return 4
	default:
		// Sum children widths
		width := 0
		for child := fiber.Child; child != nil; child = child.Sibling {
			width += r.measureFiberWidth(child)
		}
		return width
	}
}

// renderFiber renders a single Fiber to the buffer
func (r *Reconciler) renderFiber(fiber *Fiber, x, y int, buffer *paint.Buffer) {
	if fiber == nil {
		return
	}

	// Skip ComponentVNode - its children are already expanded in the Fiber tree
	if fiber.Type == rtui.VNodeComponent {
		return
	}

	// Use the render callback if set
	if r.renderCallback != nil {
		r.renderCallback(fiber, x, y, buffer)
	}
}

// measureFiberHeight measures the height of a fiber
func (r *Reconciler) measureFiberHeight(fiber *Fiber) int {
	if fiber == nil {
		return 0
	}

	// ComponentVNode - measure the child instead
	if fiber.Type == rtui.VNodeComponent {
		if fiber.Child != nil {
			return r.measureFiberHeight(fiber.Child)
		}
		return 0
	}

	// Check if Fiber has explicit height prop
	if fiber.Props != nil {
		if h, ok := fiber.Props["height"].(int); ok && h > 0 {
			return h
		}
	}

	// Get layout info from Fiber
	isHStack := fiber.GetDirection() == rtui.DirectionRow

	switch fiber.Type {
	case rtui.VNodeText, rtui.VNodeElement:
		// Text elements are single line
		if fiber.Type == rtui.VNodeText {
			return 1
		}
		// Fall through to container handling
		fallthrough
	case rtui.VNodeFragment:
		// For containers, calculate height based on layout direction
		if isHStack {
			// HStack: height is the maximum of children's heights
			maxHeight := 0
			for child := fiber.Child; child != nil; child = child.Sibling {
				childHeight := r.measureFiberHeight(child)
				if childHeight > maxHeight {
					maxHeight = childHeight
				}
			}
			return maxHeight
		} else {
			// VStack: height is the sum of children's heights
			totalHeight := 0
			for child := fiber.Child; child != nil; child = child.Sibling {
				totalHeight += r.measureFiberHeight(child)
			}
			return totalHeight
		}
	default:
		return 1
	}
}

// RenderFunc is a function to render a Fiber to the buffer
type RenderFunc func(fiber *Fiber, x, y int, buffer *paint.Buffer)

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

	// Skip the root ComponentVNode wrapper (key="root")
	if fiber.Key == "root" && fiber.Type == rtui.VNodeComponent {
		if fiber.Child != nil {
			return r.buildLayoutTreeFromFiber(fiber.Child)
		}
		return nil
	}

	return r.buildLayoutTreeFromFiber(fiber)
}

// buildLayoutTreeFromFiber builds a layout tree by traversing the Fiber tree
func (r *Reconciler) buildLayoutTreeFromFiber(fiber *Fiber) *runtime.LayoutNode {
	if fiber == nil {
		return nil
	}

	// Skip ComponentVNode nodes and process their children directly
	if fiber.Type == rtui.VNodeComponent {
		if child := fiber.Child; child != nil {
			return r.buildLayoutTreeFromFiber(child)
		}
		return nil
	}

	// Determine NodeType from Fiber
	var nodeType runtime.NodeType
	switch fiber.Tag {
	case "text":
		nodeType = runtime.NodeTypeText
	case "hstack", "row":
		nodeType = runtime.NodeTypeRow
	case "vstack", "column":
		nodeType = runtime.NodeTypeColumn
	default:
		nodeType = runtime.NodeTypeCustom
	}

	// Convert style.Style to runtime.Style
	style := runtime.Style{
		Width:  fiber.Style.Width,
		Height: fiber.Style.Height,
	}

	// Create a LayoutNode from Fiber data
	node := runtime.NewLayoutNode(fiber.Tag, nodeType, style)

	// Process children from the Fiber tree
	child := fiber.Child
	for child != nil {
		childNode := r.buildLayoutTreeFromFiber(child)
		if childNode != nil {
			node.AddChild(childNode)
		}
		child = child.Sibling
	}

	return node
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

// GetRenderedRoot returns the rendered VNode tree for focus management, etc.
// This is the VNode tree that was actually rendered to the buffer.
func (r *Reconciler) GetRenderedRoot() rtui.VNode {
	return r.renderedRoot
}

// GetFiberRoot returns the root Fiber node for traversing the Fiber tree
// This is used by GetButtons() and other collection methods to find interactive elements
func (r *Reconciler) GetFiberRoot() *Fiber {
	return r.root
}

// updateRenderedRoot extracts and stores the rendered content from the Fiber tree
// In Fiber-first architecture, we don't need to build a VNode tree
func (r *Reconciler) updateRenderedRoot() {
	log.FocusLogger.Debug("updateRenderedRoot called, r.root=%v", r.root != nil)
	// In Fiber-first, we don't maintain a VNode tree
	// The Fiber tree itself is the source of truth
	r.renderedRoot = nil
}

// SetApp sets the framework app for the reconciler
func (r *Reconciler) SetApp(app *framework.App) {
	r.app = app
}

// SetFocusManager sets the focus manager for keyboard navigation (Fiber-first)
func (r *Reconciler) SetFocusManager(mgr *rtui.FiberFocusManager) {
	r.focusMgr = mgr
}

// =============================================================================
// Deletion Handling
// =============================================================================

// commitDeletions processes all fibers marked for deletion
// This traverses the fiber tree and cleans up any nodes with the EffectDeletion flag
// Cleanup includes:
// - Running useEffect cleanup functions
// - Clearing refs
// - Removing component instances
func (r *Reconciler) commitDeletions(fiber *Fiber) {
	if fiber == nil {
		return
	}

	// Collect all fibers marked for deletion
	deletedFibers := r.collectDeletedFibers(fiber)

	log.FiberLogger.Debug("commitDeletions found %d fibers to delete", len(deletedFibers))

	// Process each deleted fiber
	for _, deleted := range deletedFibers {
		r.cleanupDeletedFiber(deleted)
	}
}

// collectDeletedFibers collects all fibers marked with EffectDeletion flag
func (r *Reconciler) collectDeletedFibers(fiber *Fiber) []*Fiber {
	var result []*Fiber

	if fiber == nil {
		return result
	}

	// Check if this fiber is marked for deletion
	if fiber.Flags&EffectDeletion != 0 {
		result = append(result, fiber)
	}

	// Recursively check children (but not siblings - they will be checked by the caller)
	if fiber.Child != nil {
		childDeletions := r.collectDeletedFibers(fiber.Child)
		result = append(result, childDeletions...)
	}

	log.FiberLogger.Debug("collectDeletedFibers found %d fibers to delete", len(result))

	return result
}

// cleanupDeletedFiber performs cleanup for a single deleted fiber
func (r *Reconciler) cleanupDeletedFiber(fiber *Fiber) {
	if fiber == nil {
		return
	}

	// For component fibers, run cleanup effects
	if fiber.Type == rtui.VNodeComponent {
		// Run useEffect cleanup functions if any
		// This is done by the instance manager
		if r.instanceMgr != nil {
			// The instance will be cleaned up during instance manager's cleanup
			_ = r.instanceMgr
		}
	}

	// Recursively cleanup children
	if fiber.Child != nil {
		r.cleanupDeletedFiber(fiber.Child)
	}
}

// =============================================================================
// Focus Management
// =============================================================================

// applyFocusStateToFiber applies focus state from the focus manager to Fiber tree
// This must be called before rendering to ensure focused elements are rendered correctly
// Uses Fiber-first approach: operates on Fiber nodes directly
func (r *Reconciler) applyFocusStateToFiber(fiber *Fiber) {
	if r.focusMgr == nil || fiber == nil {
		return
	}

	// Get the currently focused index
	focusedIndex := r.focusMgr.CurrentIndex()
	if focusedIndex < 0 {
		// No focused element, clear all focus
		r.clearFocusOnFiber(fiber)
		return
	}

	// Collect all focusable Fibers (Fiber-first)
	r.focusMgr.CollectFromFiber(fiber)
	focusable := r.focusMgr.GetFocusable()

	log.FocusLogger.Debug("applyFocus focusedIndex=%d, totalFocusable=%d", focusedIndex, len(focusable))

	// Set focus by index on Fiber nodes (Fiber-first)
	for i, f := range focusable {
		if i == focusedIndex {
			log.FocusLogger.Debug("applyFocus setting focus=true on index %d (%s)", i, f.FocusableMeta.FocusID)
			// Update FocusableVNode focus state
			if f.FocusableVNode != nil {
				f.FocusableVNode.SetFocus(true)
			}
		} else {
			// Update FocusableVNode focus state
			if f.FocusableVNode != nil {
				f.FocusableVNode.SetFocus(false)
			}
		}
	}
}

// clearFocusOnFiber recursively clears focus from all Fibers
func (r *Reconciler) clearFocusOnFiber(fiber *Fiber) {
	if fiber == nil {
		return
	}

	// Clear focus via ComponentInstance if available
	if fiber.ComponentInstance != nil {
		if focusable, ok := fiber.ComponentInstance.(interface{ SetFocus(bool) }); ok {
			focusable.SetFocus(false)
		}
	}

	if fiber.Child != nil {
		r.clearFocusOnFiber(fiber.Child)
	}
	if fiber.Sibling != nil {
		r.clearFocusOnFiber(fiber.Sibling)
	}
}

// updateFocusManagerFromFiber updates the focus manager with the new Fiber tree's focusable elements
// This should be called AFTER rendering to ensure the next render has the correct focus state
// Uses Fiber-first approach: collects Fiber nodes, not VNodes
// Also handles Layer-aware focus trapping: when Modal is open, focus is trapped in Modal layer
func (r *Reconciler) updateFocusManagerFromFiber(fiber *Fiber) {
	if r.focusMgr == nil || fiber == nil {
		return
	}

	// Check for Modal layer presence before updating
	hasModal := r.hasLayerFibers(fiber, rtui.LayerModal)
	hasOverlay := r.hasLayerFibers(fiber, rtui.LayerOverlay)

	// Determine the highest active layer
	activeLayer := rtui.LayerBase
	if hasModal {
		activeLayer = rtui.LayerModal
	} else if hasOverlay {
		activeLayer = rtui.LayerOverlay
	}

	// Update active layer (this will auto-focus first item in layer if needed)
	r.focusMgr.SetActiveLayer(activeLayer)

	// Collect all focusable Fibers from the Fiber tree (Fiber-first)
	r.focusMgr.CollectFromFiber(fiber)

	// If we have an active layer (Modal/Overlay), focus is already set by SetActiveLayer
	// Otherwise, preserve focus index
	if activeLayer == rtui.LayerBase {
		currentIndex := r.focusMgr.CurrentIndex()
		focusableCount := r.focusMgr.Count()
		if currentIndex >= 0 {
			if currentIndex >= focusableCount {
				currentIndex = focusableCount - 1
			}
			if currentIndex >= 0 {
				r.focusMgr.SetFocusByIndex(currentIndex)
			}
		} else if focusableCount > 0 {
			r.focusMgr.SetFocusByIndex(0)
		}
	}
}

// hasLayerFibers checks if there are any fibers in the specified layer
func (r *Reconciler) hasLayerFibers(fiber *rtui.Fiber, layer rtui.Layer) bool {
	if fiber == nil {
		return false
	}

	// Check current fiber
	if fiber.Layer == layer && fiber.FocusableMeta != nil && fiber.FocusableMeta.IsFocusable() {
		return true
	}

	// Check children
	if r.hasLayerFibers(fiber.Child, layer) {
		return true
	}

	// Check siblings
	if r.hasLayerFibers(fiber.Sibling, layer) {
		return true
	}

	return false
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

// SetRenderer sets the VNode renderer for SetFiber call
func (r *Reconciler) SetRenderer(renderer rtui.VNodeRenderer) {
	r.renderer = renderer
}

// GetCurrentReconciler returns the current reconciler
func GetCurrentReconciler() *Reconciler {
	return currentReconciler
}
