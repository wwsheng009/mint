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

	"github.com/wwsheng009/mint/framework/component"
	"github.com/wwsheng009/mint/internal/log"
	"github.com/wwsheng009/mint/internal/state"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/paint"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// Scheduler 调度器接口，用于通知框架需要调度帧
// Reconciler 通过此接口请求调度，解耦对 framework.App 的直接依赖
type Scheduler interface {
	MarkDirty()
}

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
	scheduler            Scheduler                       // Scheduler for frame requests
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
	// layoutRoot     *runtime.LayoutNode // Root of the layout tree
	// layoutBoxes    []runtime.LayoutBox // Layout boxes for hit testing

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
func NewReconciler(scheduler Scheduler, rootComponent rtui.ComponentFunc, config ReconcilerConfig) *Reconciler {
	timeBudget := config.TimeBudget
	if timeBudget == 0 {
		timeBudget = 5 * time.Millisecond // Default 5ms budget
	}

	return &Reconciler{
		scheduler:            scheduler,
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
	// Wrap the root component as a ComponentVNode with props support
	// This ensures it goes through beginWorkComponent which manages Context properly
	//
	// NOTE: Using NewComponentWithProps to support future props passing to root.
	// Currently props are not used (state is in global ctx), but this provides
	// extensibility for features like theme, locale, etc.
	rootComponentVNode := rtui.NewComponentWithProps("RootComponent", func(p rtui.Props) rtui.VNode {
		return renderFunc()
	})
	rootComponentVNode.SetKey("root")

	// Create or update Fiber tree
	if r.root == nil {
		// First render - create new tree from the wrapped component
		r.root = CreateFiberFromVNode(rootComponentVNode)
		// ✨ Set root Fiber's Path for proper path generation in children
		// Children will inherit this path: /root + /segment = /root/segment
		r.root.Path = "/root"
		r.root.Key = "root"
		r.root.IsRoot = true // ✨ Mark as root fiber for explicit identification
		r.workInProgress = r.root
	} else {
		// Subsequent render - create work-in-progress tree
		r.workInProgress = r.createWorkInProgress(r.root, rootComponentVNode)
		// Ensure workInProgress also has the correct Path
		r.workInProgress.Path = "/root"
		r.workInProgress.IsRoot = true // ✨ Preserve root marker on work-in-progress
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

	// ⚠️ DEPRECATED: FocusableVNode field no longer updated.
	// Focus management uses Instance.(FocusableInstance) instead.
	// work.FocusableVNode = f

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

	// ✨ Phase 0.5: Link portals to their PortalRoot targets (Phase 3)
	// Collect all PortalRoot nodes, then link Portal nodes to their targets
	r.linkPortalsToRoots(r.root)

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
	// r.layoutRoot = r.buildLayoutTree(r.root)

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

// RenderFunc is a function to render a Fiber to the buffer
type RenderFunc func(fiber *Fiber, x, y int, buffer *paint.Buffer)

// SetRenderCallback sets the render callback
func (r *Reconciler) SetRenderCallback(cb RenderFunc) {
	r.renderCallback = cb
}


// =============================================================================
// Scheduling
// =============================================================================

// requestWork requests the framework to schedule a frame
func (r *Reconciler) requestWork() {
	if r.scheduler != nil {
		r.scheduler.MarkDirty()
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

// SetScheduler sets the scheduler for the reconciler
func (r *Reconciler) SetScheduler(scheduler Scheduler) {
	r.scheduler = scheduler
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

// linkPortalsToRoots links portal nodes to their PortalRoot targets (Phase 3)
// This function:
// 1. Collects all nodes with props["portalRootId"] as PortalRoot candidates
// 2. Builds a mapping of portalRootId -> Fiber
// 3. Links nodes with props["portalRoot"] to their target PortalRoot Fiber
func (r *Reconciler) linkPortalsToRoots(root *Fiber) {
	if root == nil {
		return
	}

	// Step 1: Collect all PortalRoot nodes
	// PortalRoot nodes have props["portalRootId"] non-empty string
	portalRoots := make(map[string]*Fiber)

	var collectPortalRoots func(fiber *Fiber)
	collectPortalRoots = func(fiber *Fiber) {
		if fiber == nil {
			return
		}

		// Check if this fiber is a PortalRoot
		if fiber.Props != nil {
			if portalRootID, ok := fiber.Props["portalRootId"].(string); ok && portalRootID != "" {
				portalRoots[portalRootID] = fiber
			}
		}

		// Recurse to children and siblings
		collectPortalRoots(fiber.Child)
		collectPortalRoots(fiber.Sibling)
	}
	collectPortalRoots(root)

	// Step 2: Link Portal nodes to their PortalRoot targets
	// Portal nodes have props["portalRoot"] referencing a portalRootId
	var linkPortalNodes func(fiber *Fiber)
	linkPortalNodes = func(fiber *Fiber) {
		if fiber == nil {
			return
		}

		// Check if this fiber is a Portal
		if fiber.Props != nil {
			if portalRootID, ok := fiber.Props["portalRoot"].(string); ok && portalRootID != "" {
				// Look up the target PortalRoot Fiber
				if target, exists := portalRoots[portalRootID]; exists {
					fiber.PortalRoot = target
				}
			}
		}

		// Recurse to children and siblings
		linkPortalNodes(fiber.Child)
		linkPortalNodes(fiber.Sibling)
	}
	linkPortalNodes(root)
}

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
		focused := (i == focusedIndex)
		if focused {
			focusID := fmt.Sprintf("node-%d", f.NodeID)
			log.FocusLogger.Debug("applyFocus setting focus=true on index %d (%s)", i, focusID)
		}

		// IMPORTANT: Use Instance.(FocusableInstance).SetFocus() instead of FocusableVNode.SetFocus()
		// FocusableVNode is DEPRECATED and its hasFocus field is not used during rendering.
		// Instance.state.Focused is the actual runtime state used by resolveStyle().
		if f.Instance != nil {
			if focusable, ok := f.Instance.(interface{ SetFocus(bool) }); ok {
				focusable.SetFocus(focused)
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
	if fiber.Instance != nil {
		if focusable, ok := fiber.Instance.(interface{ SetFocus(bool) }); ok {
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
	if fiber.Layer == layer {
		// Fiber-first: use Instance.(FocusableInstance) to check if focusable
		if fiber.Instance != nil {
			if _, ok := fiber.Instance.(interface{ IsDisabled() bool }); ok {
				return true
			}
		}
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

// SetIntentRuntime sets the Intent Runtime for this reconciler.
// This enables intent dispatching from component contexts.
func (r *Reconciler) SetIntentRuntime(runtime *intent.Runtime) {
	if r.ctx != nil {
		r.ctx.SetIntentRuntime(runtime)

		// Set StateSetter on dispatcher for intent handlers
		runtime.Dispatcher.SetStateSetter(r.ctx)

		// Set schedule update callback for state changes
		r.ctx.SetScheduleUpdate(func() {
			// Request schedule update through reconciler
			r.ScheduleUpdate(rtui.LaneDefaultLane)
		})
	}
}

// SetRootContext sets the root component context for this reconciler.
// This is used for global state management in Fiber-first mode.
// The root context is shared by all components for accessing global state.
func (r *Reconciler) SetRootContext(ctx *rtui.ComponentContext) {
	r.ctx = ctx
}

// GetCurrentReconciler returns the current reconciler
func GetCurrentReconciler() *Reconciler {
	return currentReconciler
}
