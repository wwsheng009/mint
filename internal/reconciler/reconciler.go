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
	deadline   time.Time     // Current frame deadline
	timeBudget time.Duration // Time slice budget per frame

	// === Integration ===
	app                 *framework.App           // Framework app
	instanceMgr         *state.InstanceManager         // Component instance manager
	interactionStateMgr *state.InteractionStateManager // Interaction state (hover/focus/etc)
	keyValidator        *state.KeyValidator            // Key validation
	rootComponent       rtui.ComponentFunc             // Root component function
	ctx                 *rtui.ComponentContext         // Root component context
	focusMgr            *rtui.VNodeFocusManager         // Focus manager for keyboard navigation

	// === Render State ===
	buffer         *paint.Buffer // Render target
	paintCtx       component.PaintContext
	renderCallback RenderFunc // Callback for rendering VNodes
	renderedRoot   rtui.VNode  // The rendered VNode tree (for focus management, etc.)

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

	// Phase 0: Apply focus state to Fiber tree before rendering
	// IMPORTANT: We must collect the focusable elements from the NEW Fiber tree first
	// then apply the focus manager's current index to set focus on the right element
	r.applyFocusStateToFiber(r.root)

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
	if os.Getenv("TUI_DEBUG_FOCUS") == "true" {
		var tag string
		if t, ok := fiber.VNode.(interface{ Tag() string }); ok {
			tag = t.Tag()
		}
		fmt.Fprintf(os.Stderr, "[renderFiberToBuffer] tag=%q, x=%d, y=%d, isHStack=%v\n", tag, x, y, r.isHStackFiber(fiber))
	}

	// Render this fiber based on its type
	r.renderFiber(fiber, x, y, buffer)

	// Render children with proper layout
	childX := x
	childY := y

	// Check if this is a LayoutNode to handle horizontal layout
	isHStack := false
	var gap int

	// Check the VNode type for layout direction
	if fiber.VNode != nil {
		if layoutNode, ok := fiber.VNode.(*rtui.LayoutNode); ok {
			isHStack = layoutNode.Direction() == 0 // DirectionRow = 0
			gap = layoutNode.Gap()
		}
		// Also check ElementVNode for hstack/vstack
		if elemNode, ok := fiber.VNode.(*rtui.ElementVNode); ok {
			tag := elemNode.Tag()
			if tag == "hstack" || tag == "row" {
				isHStack = true
				if g, ok := elemNode.Props()["gap"].(int); ok {
					gap = g
				}
			}
		}
		// Check for Tag() method on other types (e.g., LayoutBuilder)
		if tagger, ok := fiber.VNode.(interface{ Tag() string }); ok {
			tag := tagger.Tag()
			if tag == "hstack" || tag == "row" {
				isHStack = true
				// Try to get gap from Props
				if props := fiber.VNode.Props(); props != nil {
					if g, ok := props["gap"].(int); ok {
						gap = g
					}
				}
			}
		}
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

// isHStackFiber checks if a fiber is an HStack (for debug)
func (r *Reconciler) isHStackFiber(fiber *Fiber) bool {
	if fiber == nil || fiber.VNode == nil {
		return false
	}
	if layoutNode, ok := fiber.VNode.(*rtui.LayoutNode); ok {
		return layoutNode.Direction() == 0
	}
	if elemNode, ok := fiber.VNode.(*rtui.ElementVNode); ok {
		tag := elemNode.Tag()
		return tag == "hstack" || tag == "row"
	}
	// Try Tag() method for other types
	if tagger, ok := fiber.VNode.(interface{ Tag() string }); ok {
		tag := tagger.Tag()
		if tag == "hstack" || tag == "row" {
			return true
		}
	}
	return false
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
				return len(content)
			}
		}
		// Try Content() method for types that implement it
		if contenter, ok := fiber.VNode.(interface{ Content() string }); ok {
			return len(contenter.Content())
		}
		return 10
	}

	switch v := fiber.VNode.(type) {
	case *rtui.ButtonVNode:
		return len(v.Label()) + 2 // [label]
	case *rtui.InputVNode:
		return 22 // [ + 20 chars + ]
	case *rtui.SelectVNode:
		maxLen := 10
		for _, opt := range v.Options() {
			if len(opt.Label) > maxLen {
				maxLen = len(opt.Label)
			}
		}
		return maxLen + 4 // [label + ▼]
	case *rtui.CheckboxVNode:
		width := 4 // [X] or [ ]
		if v.Label() != "" {
			width += 1 + len(v.Label())
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
				return len(content)
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
	if os.Getenv("TUI_DEBUG_FOCUS") == "true" {
		fmt.Fprintf(os.Stderr, "[updateRenderedRoot] called, r.root=%v\n", r.root != nil)
		if r.root != nil && r.root.VNode != nil {
			fmt.Fprintf(os.Stderr, "[updateRenderedRoot] r.root.VNode type=%d\n", r.root.VNode.Type())
		}
	}

	if r.root == nil || r.root.VNode == nil {
		r.renderedRoot = nil
		if os.Getenv("TUI_DEBUG_FOCUS") == "true" {
			fmt.Fprintf(os.Stderr, "[updateRenderedRoot] root or VNode is nil, setting renderedRoot=nil\n")
		}
		return
	}

	// The root is a ComponentVNode wrapper
	// Its children contain the actual rendered VNode tree
	// We need to reconstruct the VNode tree from the Fiber tree
	r.renderedRoot = r.buildVNodeTree(r.root)

	if os.Getenv("TUI_DEBUG_FOCUS") == "true" {
		if r.renderedRoot == nil {
			fmt.Fprintf(os.Stderr, "[updateRenderedRoot] buildVNodeTree returned nil!\n")
		} else {
			fmt.Fprintf(os.Stderr, "[updateRenderedRoot] buildVNodeTree returned type=%d, children=%d\n",
				r.renderedRoot.Type(), len(r.renderedRoot.Children()))
		}
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

		// Find the first child fiber and expand each child
		childFiber := fiber.Child
		expandedChildren := make([]rtui.VNode, 0, len(originalChildren))
		for _, child := range originalChildren {
			expandedChild := r.expandVNodeTree(child, childFiber)
			if expandedChild != nil {
				expandedChildren = append(expandedChildren, expandedChild)
			}
			// Move to next sibling fiber
			if childFiber != nil {
				childFiber = childFiber.Sibling
			}
		}

		// Create a new VNode with expanded children
		if len(expandedChildren) == 0 {
			return vnode
		}
		if len(expandedChildren) == 1 && expandedChildren[0] == originalChildren[0] {
			// No expansion happened, return original
			return vnode
		}

		// Clone the VNode with expanded children
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
	focusable := r.collectFocusableFromFiber(fiber)

	if os.Getenv("TUI_DEBUG_FOCUS") == "true" {
		fmt.Fprintf(os.Stderr, "[applyFocus] focusedIndex=%d, totalFocusable=%d\n", focusedIndex, len(focusable))
	}

	// Set focus by index (not by ID, because multiple elements may have the same ID)
	for i, elem := range focusable {
		if i == focusedIndex {
			if os.Getenv("TUI_DEBUG_FOCUS") == "true" {
				fmt.Fprintf(os.Stderr, "[applyFocus] setting focus=true on index %d (%s)\n", i, elem.GetFocusID())
			}
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

// collectFocusableFromFiber collects all focusable VNodes from the Fiber tree in order
func (r *Reconciler) collectFocusableFromFiber(fiber *Fiber) []rtui.FocusableVNode {
	return r.collectFocusableFromFiberWithIndex(fiber, 0)
}

// collectFocusableFromFiberWithIndex collects focusable VNodes with a starting index
// This ensures consistent index assignment across recursive calls
func (r *Reconciler) collectFocusableFromFiberWithIndex(fiber *Fiber, startIndex int) []rtui.FocusableVNode {
	result := make([]rtui.FocusableVNode, 0)
	currentIndex := startIndex

	if fiber == nil || fiber.VNode == nil {
		return result
	}

	// Skip ComponentVNode wrappers
	if fiber.VNode.Type() == rtui.VNodeComponent {
		if fiber.Child != nil {
			return r.collectFocusableFromFiberWithIndex(fiber.Child, startIndex)
		}
		return result
	}

	// Check if current VNode is focusable
	if focusable, ok := fiber.VNode.(rtui.FocusableVNode); ok && focusable.IsFocusable() {
		// Set focusIndex for buttons to ensure unique FocusID
		if btn, ok := fiber.VNode.(interface{ SetFocusIndex(int) }); ok {
			btn.SetFocusIndex(currentIndex)
		}
		if os.Getenv("TUI_DEBUG_FOCUS") == "true" {
			fmt.Fprintf(os.Stderr, "[collectFocusable] adding focusable %d: %s\n", currentIndex, focusable.GetFocusID())
		}
		result = append(result, focusable)
		currentIndex++
	}

	// Recursively check children (pass the current index)
	if fiber.Child != nil {
		childResult := r.collectFocusableFromFiberWithIndex(fiber.Child, currentIndex)
		result = append(result, childResult...)
		currentIndex += len(childResult)
	}

	// Recursively check siblings (pass the current index)
	if fiber.Sibling != nil {
		siblingResult := r.collectFocusableFromFiberWithIndex(fiber.Sibling, currentIndex)
		result = append(result, siblingResult...)
	}

	return result
}

// setFocusOnFiber is deprecated - use applyFocusStateToFiber with index instead
func (r *Reconciler) setFocusOnFiber(fiber *Fiber, focusID string) {
	if fiber == nil || fiber.VNode == nil {
		return
	}

	// Check if this VNode matches the focus ID
	if focusable, ok := fiber.VNode.(rtui.FocusableVNode); ok {
		if focusable.GetFocusID() == focusID {
			focusable.SetFocus(true)
		} else {
			// Clear focus from non-focused elements
			focusable.SetFocus(false)
		}
	}

	// Recursively process children and siblings
	if fiber.Child != nil {
		r.setFocusOnFiber(fiber.Child, focusID)
	}
	if fiber.Sibling != nil {
		r.setFocusOnFiber(fiber.Sibling, focusID)
	}
}

// updateFocusManagerFromFiber updates the focus manager with the new Fiber tree's focusable elements
// This should be called AFTER rendering to ensure the next render has the correct focus state
func (r *Reconciler) updateFocusManagerFromFiber(fiber *Fiber) {
	if r.focusMgr == nil || fiber == nil {
		return
	}

	// Collect all focusable VNodes from the Fiber tree
	focusable := r.collectFocusableFromFiber(fiber)
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

// GetCurrentReconciler returns the current reconciler
func GetCurrentReconciler() *Reconciler {
	return currentReconciler
}
