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
	"github.com/wwsheng009/mint/internal/reconciler"
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
	focusMgr            *rtui.VNodeFocusManager        // Focus manager for keyboard navigation

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
		if adapter, ok := r.renderer.(interface{ SetFiber(*reconciler.Fiber) }); ok {
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
	// ComponentVNodes are just wrappers for their rendered content
	if fiber.VNode != nil && fiber.VNode.Type() == rtui.VNodeComponent {
		// Render the single child at the same position
		if fiber.Child != nil {
			r.renderFiberToBuffer(fiber.Child, x, y, buffer)
		}
		return
	}

	// Debug: log fiber traversal
	var tag string
	if t, ok := fiber.VNode.(interface{ Tag() string }); ok {
		tag = t.Tag()
	}
	layoutInfo := rtui.GetLayoutInfo(fiber.VNode)
	log.RenderLogger.Debug("renderFiberToBuffer tag=%q, x=%d, y=%d, isHStack=%v", tag, x, y, layoutInfo.IsHorizontal)

	// Render this fiber based on its type
	r.renderFiber(fiber, x, y, buffer)

	// Render children with proper layout
	childX := x
	childY := y

	isHStack := layoutInfo.IsHorizontal
	gap := layoutInfo.Gap

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

	// ComponentVNode - measure the child instead
	if fiber.VNode.Type() == rtui.VNodeComponent {
		if fiber.Child != nil {
			return r.measureFiberWidth(fiber.Child)
		}
		return 0
	}

	// For text nodes, get content from Props or Content() method
	if fiber.VNode.Type() == rtui.VNodeText {
		// Try to get content from Props first (works for both ui.TextVNode and basic.TextVNode)
		if props := fiber.VNode.Props(); props != nil {
			if content, ok := props["content"].(string); ok {
				return paint.StringWidth(content)
			}
		}
		// Try Content() method for types that implement it
		if contenter, ok := fiber.VNode.(interface{ Content() string }); ok {
			return paint.StringWidth(contenter.Content())
		}
		return 10
	}

	switch v := fiber.VNode.(type) {
	case *rtui.ButtonVNode:
		return paint.StringWidth(v.Label()) + 2 // [label]
	case *rtui.InputVNode:
		return 22 // [ + 20 chars + ]
	case *rtui.SelectVNode:
		maxLen := 10
		for _, opt := range v.Options() {
			w := paint.StringWidth(opt.Label)
			if w > maxLen {
				maxLen = w
			}
		}
		return maxLen + 4 // [label + ▼]
	case *rtui.CheckboxVNode:
		width := 4 // [X] or [ ]
		if v.Label() != "" {
			width += 1 + paint.StringWidth(v.Label())
		}
		return width
	case *rtui.ElementVNode, *rtui.LayoutNode, *rtui.FragmentVNode:
		// Containers - sum up children widths
		width := 0
		for child := fiber.Child; child != nil; child = child.Sibling {
			width += r.measureFiberWidth(child)
		}
		return width
	default:
		// For any other type, try to get content from Props
		if props := fiber.VNode.Props(); props != nil {
			if content, ok := props["content"].(string); ok {
				return paint.StringWidth(content)
			}
		}
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
	if fiber.VNode.Type() == rtui.VNodeComponent {
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

	// ComponentVNode - measure the child instead
	if fiber.VNode.Type() == rtui.VNodeComponent {
		if fiber.Child != nil {
			return r.measureFiberHeight(fiber.Child)
		}
		return 0
	}

	// Check if VNode has explicit height prop
	if props := fiber.VNode.Props(); props != nil {
		if h, ok := props["height"].(int); ok && h > 0 {
			return h
		}
	}

	// Get layout info to determine if this is a vertical or horizontal container
	layoutInfo := rtui.GetLayoutInfo(fiber.VNode)

	switch fiber.VNode.(type) {
	case *rtui.ButtonVNode, *rtui.InputVNode, *rtui.CheckboxVNode:
		return 1 // Single-line controls
	case *rtui.SelectVNode:
		return 1 // Single-line dropdown
	case *rtui.TextVNode, *rtui.ElementVNode:
		// Text elements are single line
		if rtui.GetTextContent(fiber.VNode) != "" {
			return 1
		}
		// Fall through to container handling
	case *rtui.LayoutNode, *rtui.FragmentVNode:
		// For containers, calculate height based on layout direction
		if layoutInfo.IsHorizontal {
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
		// For any other type, default to 1
		return 1
	}
	return 1 // Should never reach here, but needed for Go's type switch
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

	// Skip the root ComponentVNode wrapper (key="root")
	// The actual content is in the Fiber's children, not in VNode.Children()
	// This is because ComponentVNode.Children() returns nil, but the Fiber tree
	// has the expanded children via beginWorkComponent
	if fiber.Key == "root" && fiber.VNode.Type() == rtui.VNodeComponent {
		if fiber.Child != nil {
			return r.buildLayoutTreeFromFiber(fiber.Child)
		}
		return nil
	}

	// For non-root nodes, try to convert via VNodeConverter
	// But use Fiber children instead of VNode children for ComponentVNode
	return r.buildLayoutTreeFromFiber(fiber)
}

// buildLayoutTreeFromFiber builds a layout tree by traversing the Fiber tree
// This ensures we use the expanded children from the Fiber tree, not VNode.Children()
func (r *Reconciler) buildLayoutTreeFromFiber(fiber *Fiber) *runtime.LayoutNode {
	if fiber == nil || fiber.VNode == nil {
		return nil
	}

	// Skip ComponentVNode nodes and process their children directly
	if fiber.VNode.Type() == rtui.VNodeComponent {
		// Process children instead
		if child := fiber.Child; child != nil {
			return r.buildLayoutTreeFromFiber(child)
		}
		return nil
	}

	// Convert this VNode to LayoutNode
	node := r.vnodeConverter.Convert(fiber.VNode)
	if node == nil {
		// Try to build from children if conversion failed
		if child := fiber.Child; child != nil {
			return r.buildLayoutTreeFromFiber(child)
		}
		return nil
	}

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

// updateRenderedRoot extracts and stores the rendered VNode tree from the Fiber tree
// The root Fiber is a ComponentVNode wrapper, its children contain the actual content
func (r *Reconciler) updateRenderedRoot() {
	log.FocusLogger.Debug("updateRenderedRoot called, r.root=%v", r.root != nil)
	if r.root != nil && r.root.VNode != nil {
		log.FocusLogger.Debug("updateRenderedRoot r.root.VNode type=%d", r.root.VNode.Type())
	}

	if r.root == nil || r.root.VNode == nil {
		r.renderedRoot = nil
		log.FocusLogger.Debug("updateRenderedRoot root or VNode is nil, setting renderedRoot=nil")
		return
	}

	// The root is a ComponentVNode wrapper
	// Its children contain the actual rendered VNode tree
	// We need to reconstruct the VNode tree from the Fiber tree
	r.renderedRoot = r.buildVNodeTree(r.root)

	if r.renderedRoot == nil {
		log.FocusLogger.Debug("updateRenderedRoot buildVNodeTree returned nil!")
	} else {
		log.FocusLogger.Debug("updateRenderedRoot buildVNodeTree returned type=%d, children=%d",
			r.renderedRoot.Type(), len(r.renderedRoot.Children()))
	}
}

// buildVNodeTree reconstructs a VNode tree from a Fiber tree
func (r *Reconciler) buildVNodeTree(fiber *Fiber) rtui.VNode {
	if fiber == nil || fiber.VNode == nil {
		return nil
	}

	// Skip the root ComponentVNode wrapper and return its children
	// This is the actual rendered content
	if fiber.VNode.Type() == rtui.VNodeComponent && fiber.Key == "root" {
		// Return the children as a Fragment
		children := r.buildVNodeList(fiber.Child)
		if len(children) == 0 {
			return nil
		}
		if len(children) == 1 {
			// For single child, still need to expand any ComponentVNodes in its subtree
			return r.expandVNodeTree(children[0], fiber.Child)
		}
		return rtui.Fragment(children...)
	}

	// For ComponentVNode (other than root), we need to recursively rebuild
	// because the VNode's children might have changed during reconciliation
	if fiber.VNode.Type() == rtui.VNodeComponent {
		// Get the component instance's rendered children from the Fiber tree
		children := r.buildVNodeList(fiber.Child)
		if len(children) == 0 {
			return fiber.VNode
		}
		if len(children) == 1 {
			return children[0]
		}
		return rtui.Fragment(children...)
	}

	// For other fibers, return the VNode
	return fiber.VNode
}

// expandVNodeTree recursively expands ComponentVNodes in a VNode tree
// This is needed because ComponentVNode.Children() returns nil, so we need
// to use the Fiber tree to get the actual rendered children
func (r *Reconciler) expandVNodeTree(vnode rtui.VNode, fiber *Fiber) rtui.VNode {
	if vnode == nil {
		return nil
	}

	// If this is a ComponentVNode and we have the corresponding Fiber, expand it
	if vnode.Type() == rtui.VNodeComponent && fiber != nil {
		children := r.buildVNodeList(fiber.Child)
		if len(children) == 0 {
			return vnode
		}
		if len(children) == 1 {
			return r.expandVNodeTree(children[0], fiber.Child)
		}
		return rtui.Fragment(children...)
	}

	// For other VNodes, recursively expand children
	if vnode.Type() == rtui.VNodeElement || vnode.Type() == rtui.VNodeFragment {
		originalChildren := vnode.Children()
		if len(originalChildren) == 0 {
			return vnode
		}

		// If fiber is nil, we can't expand children - return original vnode
		if fiber == nil {
			return vnode
		}

		// ✨ IMPORTANT: Build children from Fiber tree, not from VNode.Children()
		// Fiber.VNode has the correct Key set by reconciliation, but VNode.Children()
		// returns the original children which may have outdated keys.
		// We use Fiber to get the correct child VNodes with proper Keys.
		expandedChildren := r.buildVNodeList(fiber.Child)
		if len(expandedChildren) == 0 {
			return vnode
		}

		// Clone the VNode with expanded children (which now have correct Keys from Fiber)
		cloned := r.cloneVNodeWithChildren(vnode, expandedChildren)
		return cloned
	}

	return vnode
}

// cloneVNodeWithChildren creates a new VNode with the same properties but different children
func (r *Reconciler) cloneVNodeWithChildren(vnode rtui.VNode, children []rtui.VNode) rtui.VNode {
	// For ElementVNode (including LayoutNode which embeds ElementVNode)
	if vnode.Type() == rtui.VNodeElement {
		// Get the tag - check for Tag() method
		var tag string
		if tagger, ok := vnode.(interface{ Tag() string }); ok {
			tag = tagger.Tag()
		}

		// Create new element and copy properties
		cloned := rtui.NewElement(tag)

		// Copy props
		if props := vnode.Props(); props != nil && len(props) > 0 {
			cloned.SetProps(props)
		}

		// Copy key
		if key := vnode.Key(); key != "" {
			cloned.SetKey(key)
		}

		// Set the new children
		cloned.SetChildren(children)

		return cloned
	}

	// Handle FragmentVNode
	if vnode.Type() == rtui.VNodeFragment {
		cloned := rtui.NewFragment(children...)
		if key := vnode.Key(); key != "" {
			cloned.SetKey(key)
		}
		return cloned
	}

	// For other types, return as-is
	return vnode
}

// buildVNodeList builds a list of VNodes from sibling Fibers
func (r *Reconciler) buildVNodeList(fiber *Fiber) []rtui.VNode {
	var result []rtui.VNode
	for fiber != nil {
		if vnode := r.buildVNodeTree(fiber); vnode != nil {
			result = append(result, vnode)
		}
		fiber = fiber.Sibling
	}
	return result
}

// SetInstanceManager sets the instance manager (shared with declarativeRoot)
func (r *Reconciler) SetInstanceManager(mgr *state.InstanceManager) {
	r.instanceMgr = mgr
}

// SetApp sets the framework app for the reconciler
func (r *Reconciler) SetApp(app *framework.App) {
	r.app = app
}

// SetFocusManager sets the focus manager for keyboard navigation
func (r *Reconciler) SetFocusManager(mgr *rtui.VNodeFocusManager) {
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

// applyFocusStateToFiber applies focus state from the focus manager to Fiber tree VNodes
// This must be called before rendering to ensure focused elements are rendered correctly
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

	// Collect all focusable VNodes in order
	focusable := CollectFocusableFromFiber(fiber)

	log.FocusLogger.Debug("applyFocus focusedIndex=%d, totalFocusable=%d", focusedIndex, len(focusable))

	// Set focus by index (not by ID, because multiple elements may have the same ID)
	for i, elem := range focusable {
		if i == focusedIndex {
			log.FocusLogger.Debug("applyFocus setting focus=true on index %d (%s)", i, elem.GetFocusID())
			elem.SetFocus(true)
		} else {
			elem.SetFocus(false)
		}
	}
}

// clearFocusOnFiber recursively clears focus from all VNodes
func (r *Reconciler) clearFocusOnFiber(fiber *Fiber) {
	if fiber == nil || fiber.VNode == nil {
		return
	}

	if focusable, ok := fiber.VNode.(rtui.FocusableVNode); ok {
		focusable.SetFocus(false)
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
func (r *Reconciler) updateFocusManagerFromFiber(fiber *Fiber) {
	if r.focusMgr == nil || fiber == nil {
		return
	}

	// Collect all focusable VNodes from the Fiber tree
	focusable := CollectFocusableFromFiber(fiber)
	if len(focusable) == 0 {
		return
	}

	// Get the current focus index BEFORE updating the list
	currentIndex := r.focusMgr.CurrentIndex()

	// Directly update the focusable list in the focus manager
	// We do this directly instead of calling SetFocusable to avoid the ID-based focus reset
	// that happens when multiple elements have the same ID
	r.focusMgr.UpdateFocusableList(focusable)

	// Preserve the current focus index by clamping it to the new list size
	if currentIndex >= 0 {
		if currentIndex >= len(focusable) {
			currentIndex = len(focusable) - 1
		}
		if currentIndex >= 0 {
			r.focusMgr.SetFocusByIndex(currentIndex)
		}
	} else if len(focusable) > 0 {
		// If no current focus and there are focusable nodes, focus the first one
		// This ensures that in a new render, the first element gets focus
		r.focusMgr.SetFocusByIndex(0)
	}
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

// SetRenderer sets VNode renderer for SetFiber call
func (r *Reconciler) SetRenderer(renderer rtui.VNodeRenderer) {
	r.renderer = renderer
}

// GetCurrentReconciler returns the current reconciler
func GetCurrentReconciler() *Reconciler {
	return currentReconciler
}
