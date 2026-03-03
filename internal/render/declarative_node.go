// Package render provides declarative node implementation for bridging VNode and framework Component.
package render

import (
	"fmt"
	"os"
	"sync"

	"github.com/wwsheng009/mint/framework/component"
	"github.com/wwsheng009/mint/internal/log"
	"github.com/wwsheng009/mint/internal/reconciler"
	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/border"
	"github.com/wwsheng009/mint/runtime/event"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/render"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// DeclarativeNode - Bridges VNode with framework Component
// =============================================================================
// DeclarativeNode allows a VNode tree to be used as a framework Component.
// This enables mixing declarative UI (VNode) with imperative Components.

// RenderMode specifies which rendering path to use
type RenderMode int

const (
	// RenderModeLegacy uses the old VNode-based rendering
	RenderModeLegacy RenderMode = iota
	// RenderModeFiberFirst uses the new Fiber-first rendering pipeline
	RenderModeFiberFirst
	// RenderModeBoth runs both paths and compares results (for testing)
	RenderModeBoth
)

// DeclarativeNode wraps a VNode function for use as a framework Component
type DeclarativeNode struct {
	mu       sync.RWMutex
	root     rtui.VNode              // The root VNode of this tree (legacy, will be removed)
	renderFn rtui.ComponentFunc      // Function that renders the VNode
	instance *rtui.ComponentContext  // Component instance for hooks
	focusMgr *rtui.FiberFocusManager // Focus manager for keyboard navigation (Fiber-first)

	// Framework integration
	reconciler rtui.Reconciler   // Fiber reconciler (if enabled) - use interface to avoid import cycle
	renderer   rtui.VNodeRenderer // VNode renderer (implements VNodeRenderer interface)
	useFiber   bool               // Whether Fiber mode is enabled

	// scheduler 用于非 Fiber 模式下请求帧调度
	scheduler reconciler.Scheduler // Scheduler for frame requests (non-Fiber mode only)

	// === Intent Integration ===
	intentRuntime *intent.Runtime // Intent runtime for dispatching intents

	// === Fiber-first Rendering Pipeline (Phase 4) ===
	renderMode        RenderMode                 // Current rendering mode
	newLayoutEngine   *NewLayoutEngineAdapter    // New layout engine (Fiber-first only)
	paintEngine       *PaintEngine               // Paint engine for Fiber-first
	converter         *FiberToPaintableConverter // Fiber to Paintable converter
	fiberFirstEnabled bool                       // Whether Fiber-first mode is enabled

	// === Portal Support (Two-Phase Layout) ===
	portalLayoutEngine *PortalAwareLayoutEngine // Portal-aware layout engine (Phase 5)
	usePortalLayout    bool                     // Whether to use Portal-aware layout

	// === HitMap Storage ===
	// Fiber-first mode stores HitMap here (separate from RenderingPipeline)
	fiberLastHitMap *event.HitMap // HitMap from fiberFirstPaint() path

	// === Layout Result Storage ===
	lastLayoutResult *layout.LayoutResult // Last layout computation result (for GetLayoutBoxes)

	// === Paintable Result Storage ===
	lastPaintableRoot *paint.PaintableBox // Last paintable layout result (for GetPaintableBoxes)

	// === Portal Box Debug Storage ===
	lastPortalBoxes []*layout.LayoutBox // Portal boxes from last layout (for debugging)
}

// NewDeclarativeNode creates a new declarative node from a VNode
func NewDeclarativeNode(vnode rtui.VNode) *DeclarativeNode {
	return &DeclarativeNode{
		root: vnode,
	}
}

// SetScheduler sets the scheduler for requesting frame updates
// Only used in non-Fiber mode. In Fiber mode, the reconciler handles scheduling internally.
func (n *DeclarativeNode) SetScheduler(scheduler reconciler.Scheduler) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.scheduler = scheduler
}

// SetApp sets the framework app (scheduler) on the DeclarativeNode and its reconciler
// This is called from ui.Run to enable frame scheduling
func (n *DeclarativeNode) SetApp(app interface{}) {
	// Set scheduler on the DeclarativeNode (for non-Fiber mode)
	if scheduler, ok := app.(reconciler.Scheduler); ok {
		n.SetScheduler(scheduler)
	}

	// Set app on the reconciler (for Fiber mode)
	if n.reconciler != nil {
		if reconciler, ok := n.reconciler.(interface{ SetApp(interface{}) }); ok {
			reconciler.SetApp(app)
		}
	}
}

// NewDeclarativeNodeFromFuncWithFiber creates a new declarative node with Fiber reconciler enabled
// This function is called from ui.Run when MINT_USE_FIBER is set
// Scheduler must be set later via SetApp() to enable frame scheduling
func NewDeclarativeNodeFromFuncWithFiber(fn rtui.ComponentFunc) *DeclarativeNode {
	// Create root component context (shared global state)
	rootCtx := rtui.NewComponentContextForRoot()

	// Import the reconciler package here to avoid import cycles in ui package
	// This is safe because internal/render can import internal/reconciler
	r := newFiberReconciler(nil, fn, rootCtx) // scheduler will be set later via SetApp

	// Create a new FiberFocusManager (Fiber-first architecture)
	// This replaces the VNodeFocusManager to ensure focus state is managed on Fiber nodes
	focusMgr := rtui.NewFiberFocusManager()

	// Set the focus manager on the reconciler so it can apply focus state before rendering
	if adapter, ok := r.(*fiberReconcilerAdapter); ok {
		adapter.SetFocusManager(focusMgr)
	}

	// Use the new PipelineRenderer with Layout/Paint separation
	// The Fiber reconciler handles the update/reconciliation logic,
	// while PipelineRenderer handles the actual rendering
	renderer := NewPipelineRendererAdapter()

	// Phase 8: Set renderer on reconciler for NodeID propagation
	// This enables SetFiber() to be called after reconciliation completes
	if adapter, ok := r.(*fiberReconcilerAdapter); ok {
		adapter.SetRenderer(renderer)
	}

	node := &DeclarativeNode{
		renderFn:   fn,
		instance:   rootCtx, // Use the same context for global state
		focusMgr:   focusMgr,
		reconciler: r,
		renderer:   renderer,
		useFiber:   true,
	}

	// Initialize Fiber-first pipeline components
	node.initFiberFirstPipeline()

	return node
}

// initFiberFirstPipeline initializes the Fiber-first rendering pipeline components
func (n *DeclarativeNode) initFiberFirstPipeline() {
	// Check if Fiber-first mode is enabled via environment variable
	fiberFirstEnv := os.Getenv("MINT_FIBER_FIRST")
	n.fiberFirstEnabled = fiberFirstEnv == "true" || fiberFirstEnv == "1"

	if n.fiberFirstEnabled {
		log.RenderLogger.Debug("[DeclarativeNode] Fiber-first mode ENABLED")
		n.renderMode = RenderModeFiberFirst
		// Use the new layout engine directly (runtime/layout), bypassing LayoutSwitcher
		// This ensures we never go through the compute path
		n.newLayoutEngine = NewNewLayoutEngineAdapter()
		n.paintEngine = NewPaintEngine()
		// converter will be created with Fiber root during render
	} else {
		// Default to legacy mode for now
		n.renderMode = RenderModeLegacy
		log.RenderLogger.Debug("[DeclarativeNode] Using legacy render mode (MINT_FIBER_FIRST not set)")
	}
}

// SetRenderMode sets the rendering mode
func (n *DeclarativeNode) SetRenderMode(mode RenderMode) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.renderMode = mode
	if mode == RenderModeFiberFirst {
		n.fiberFirstEnabled = true
		// Use the new layout engine directly (runtime/layout), bypassing LayoutSwitcher
		if n.newLayoutEngine == nil {
			n.newLayoutEngine = NewNewLayoutEngineAdapter()
		}
		if n.paintEngine == nil {
			n.paintEngine = NewPaintEngine()
		}
	}
}

// GetRenderMode returns the current rendering mode
func (n *DeclarativeNode) GetRenderMode() RenderMode {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.renderMode
}

// IsFiberFirstEnabled returns whether Fiber-first mode is enabled
func (n *DeclarativeNode) IsFiberFirstEnabled() bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.fiberFirstEnabled
}

// =============================================================================
// Intent Runtime Integration
// =============================================================================

// SetDeclarativeNodeIntentRuntime sets the Intent Runtime for this declarative node.
// This is called from ui.Run() after creating the declarative node.
func SetDeclarativeNodeIntentRuntime(node *DeclarativeNode, rt *intent.Runtime) {
	node.mu.Lock()
	defer node.mu.Unlock()
	node.intentRuntime = rt

	// Also set on reconciler if present (Fiber mode)
	if node.reconciler != nil {
		if adapter, ok := node.reconciler.(*fiberReconcilerAdapter); ok {
			adapter.r.SetIntentRuntime(rt)
		}
	}

	// CRITICAL: Set StateSetter on dispatcher to use node.instance (root ComponentContext)
	// This enables Intent handlers to update the global component state.
	// In Fiber mode, node.instance is the ComponentContextForRoot that acts as global state.
	node.instance.SetIntentRuntime(rt)          // Set runtime for context
	rt.Dispatcher.SetStateSetter(node.instance) // Use root context as state setter

	// Set schedule update callback - trigger Fiber reconciler update in Fiber mode
	node.instance.SetScheduleUpdate(func() {
		if node.reconciler != nil {
			// Fiber mode: schedule reconciler update
			if adapter, ok := node.reconciler.(*fiberReconcilerAdapter); ok {
				// Use LaneSyncLane for synchronous updating
				adapter.r.ScheduleUpdate(rtui.LaneSyncLane)
			}
		} else if node.scheduler != nil {
			// Non-Fiber mode: mark framework app as dirty
			node.scheduler.MarkDirty()
		}
	})
}

// GetIntentRuntime returns the Intent Runtime for this declarative node.
func (n *DeclarativeNode) GetIntentRuntime() *intent.Runtime {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.intentRuntime
}

// =============================================================================
// SetReconciler
// =============================================================================

// SetReconciler sets the Fiber reconciler for this node
// This is called by ui.Run when Fiber mode is enabled
// Phase 8: Set renderer on reconciler for NodeID propagation
func (n *DeclarativeNode) SetReconciler(r rtui.Reconciler) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.reconciler = r
	n.useFiber = r != nil

	// Phase 8: Set renderer on reconciler so it can call SetFiber after reconciliation
	// This enables NodeID propagation from Fiber tree to LayoutEngine
	if setter, ok := r.(interface{ SetRenderer(rtui.VNodeRenderer) }); ok {
		setter.SetRenderer(n.renderer)
	}
}

// GetReconciler returns the Fiber reconciler for this node
func (n *DeclarativeNode) GetReconciler() rtui.Reconciler {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.reconciler
}

// GetRenderedRoot returns the rendered VNode tree from the Fiber reconciler
// This is called by Framework to get the tree after reconciliation for Inspector
func (n *DeclarativeNode) GetRenderedRoot() rtui.VNode {
	n.mu.RLock()
	defer n.mu.RUnlock()

	// If using Fiber reconciler, get the rendered tree from reconciler
	if n.reconciler != nil {
		if provider, ok := n.reconciler.(interface{ GetRenderedRoot() rtui.VNode }); ok {
			return provider.GetRenderedRoot()
		}
	}

	// Fallback: return the current root
	return n.root
}

// =============================================================================
// framework.Node interface implementation
// =============================================================================

// ID returns the node ID
func (n *DeclarativeNode) ID() string {
	n.mu.RLock()
	defer n.mu.RUnlock()
	if n.root != nil {
		if key := n.root.Key(); key != "" {
			return "declarative:" + key
		}
	}
	return "declarative:node"
}

// Type returns the node type
func (n *DeclarativeNode) Type() string {
	return "DeclarativeNode"
}

// Children returns child VNodes
func (n *DeclarativeNode) Children() []runtime.Node {
	// DeclarativeNode doesn't expose children as framework Nodes
	// They are managed internally through the VNode tree
	return nil
}

// =============================================================================
// framework.Measurable interface implementation
// =============================================================================

// Measure calculates the ideal size for this declarative node
func (n *DeclarativeNode) Measure(maxWidth, maxHeight int) (width, height int) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	// Default size if not measured
	width, height = maxWidth, 1

	// Try to get size from VNode if it has layout info
	if n.root != nil {
		if props := n.root.Props(); props != nil {
			if w := props.GetInt("width"); w > 0 {
				width = minInt(w, maxWidth)
			}
			if h := props.GetInt("height"); h > 0 {
				height = minInt(h, maxHeight)
			}
		}
	}

	return width, height
}

// =============================================================================
// framework.Paintable interface implementation
// =============================================================================

// Paint renders the VNode tree to the buffer
// UNIFIED RENDERING: Supports both Legacy and Fiber-first rendering paths
func (n *DeclarativeNode) Paint(ctx paint.PaintContext, buf *paint.Buffer) {
	// Acquire read lock to read state for determining render mode
	n.mu.RLock()
	fiberFirstEnabled := n.fiberFirstEnabled
	useFiber := n.useFiber
	renderMode := n.renderMode
	hasReconciler := (n.reconciler != nil)
	n.mu.RUnlock()

	// Debug logging
	log.PaintLogger.Debug("[DeclarativeNode.Paint] START: ctx.X=%d, ctx.Y=%d, buf=%dx%d, useFiber=%v, renderMode=%v, fiberFirstEnabled=%v",
		ctx.Bounds.X, ctx.Bounds.Y, buf.Width, buf.Height, useFiber, renderMode, fiberFirstEnabled)

	// Check if Fiber-first mode is enabled
	if fiberFirstEnabled && useFiber && hasReconciler {
		switch renderMode {
		case RenderModeFiberFirst:
			// Release locks before calling fiberFirstPaint (it will acquire its own locks)
			n.fiberFirstPaint(ctx, buf)
			return
		case RenderModeBoth:
			// Release locks before calling comparePaint (it will acquire its own locks)
			n.comparePaint(ctx, buf)
			return
		default:
			// Fall through to legacy path
		}
	}

	// Acquire write lock for legacy rendering mode
	n.mu.Lock()
	defer n.mu.Unlock()

	// === Legacy Rendering Path ===
	// NOT HITMAP
	n.legacyPaint(ctx, buf)
}

// fiberFirstPaint renders using the new Fiber-first pipeline
// Phase 1: Reconcile (VNode -> Fiber, VNode discarded)
// Phase 2: Layout (Fiber -> LayoutBox, no VNode access)
// Phase 3: Paint (LayoutBox -> PaintableBox -> Buffer)
func (n *DeclarativeNode) fiberFirstPaint(ctx paint.PaintContext, buf *paint.Buffer) {
	log.RenderLogger.Debug("[DeclarativeNode.fiberFirstPaint] === STARTING Fiber-first render ===")
	// Phase 1: Fiber Reconciliation
	// The reconciler updates the Fiber tree, VNode is discarded after this
	// Use a minimal buffer for reconciliation (actual painting happens later)
	nullBuf := paint.NewBuffer(1, 1)
	n.reconciler.Render(paint.PaintContext{
		Bounds: paint.Rect{X: 0, Y: 0, Width: 1, Height: 1},
	}, nullBuf, n.renderFn)

	// Get the Fiber root from reconciler
	fiberRoot := n.getFiberRoot()
	if fiberRoot == nil {
		log.PaintLogger.Debug("[DeclarativeNode.fiberFirstPaint] Fiber root is nil, falling back to legacy")
		n.legacyPaint(ctx, buf)
		return
	}

	log.PaintLogger.Debug("[DeclarativeNode.fiberFirstPaint] Fiber root OK: type=%d, tag=%s, children=%d",
		fiberRoot.Type, fiberRoot.Tag, countFiberChildren(fiberRoot))

	// Phase 2: Fiber-based Layout
	// 方案A: 单树共享布局 - 所有layer在一个布局树中计算
	// 移除了LayerManager的坐标归一化，保留原始坐标用于正确渲染

	// Ensure layout engine has adapter
	if n.newLayoutEngine == nil {
		n.newLayoutEngine = NewNewLayoutEngineAdapter()
	}

	constraints := runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth:  ctx.AvailableWidth,
		MinHeight: 0,
		MaxHeight: ctx.AvailableHeight,
	}

	// Check if the tree contains Portals (check for PortalRoot/Portal props)
	hasPortals := n.hasPortals(fiberRoot)

	var layoutResult LayoutResult
	var err error

	if hasPortals && n.usePortalLayout {
		// Use Portal-aware layout engine for two-phase layout
		log.RenderLogger.Debug("[DeclarativeNode.fiberFirstPaint] ✅ Using Portal-aware layout engine")

		if n.portalLayoutEngine == nil {
			n.portalLayoutEngine = NewPortalAwareLayoutEngine()
		}

		// Perform two-phase layout using PortalAwareLayoutEngine
		mainResult, portalBoxes, layoutErr := n.portalLayoutEngine.Layout(fiberRoot, constraints)

		if layoutErr != nil {
			log.RenderLogger.Debug("[DeclarativeNode.fiberFirstPaint] Portal Layout FAILED: %v, falling back to legacy", layoutErr)
			n.legacyPaint(ctx, buf)
			return
		}

		// Merge portal boxes into main result and store for debugging
		n.mu.Lock()
		if len(portalBoxes) > 0 {
			log.RenderLogger.Debug("[DeclarativeNode.fiberFirstPaint] Merged %d Portal boxes into layout", len(portalBoxes))
			mainResult.Boxes = append(mainResult.Boxes, portalBoxes...)
			// Convert to pointer slice to preserve tree structure
			n.lastPortalBoxes = make([]*layout.LayoutBox, len(portalBoxes))
			for i := range portalBoxes {
				n.lastPortalBoxes[i] = &portalBoxes[i]
			}

			// Build a combined hit map that includes main tree + portals
			// For now, portals are added to the main hit map
			// TODO: Refine hit map building for portal z-ordering
		} else {
			n.lastPortalBoxes = nil
		}
		n.mu.Unlock()

		layoutResult = &newLayoutResultAdapter{result: mainResult, fiberRoot: fiberRoot}
	} else {
		// Use standard layout engine (single-phase)
		layoutResult, err = n.newLayoutEngine.LayoutFiber(fiberRoot, constraints)
		if err != nil {
			log.RenderLogger.Debug("[DeclarativeNode.fiberFirstPaint] Layout FAILED: %v, falling back to legacy", err)
			n.legacyPaint(ctx, buf)
			return
		}
	}

	log.RenderLogger.Debug("[DeclarativeNode.fiberFirstPaint] Layout complete")

	// Get the layout result
	if adapter, ok := layoutResult.(*newLayoutResultAdapter); ok {
		// Save layout result for GetLayoutBoxes()
		innerResult := adapter.result
		n.mu.Lock()
		n.lastLayoutResult = innerResult
		n.mu.Unlock()

		if innerResult == nil || innerResult.Root == nil {
			log.RenderLogger.Debug("[DeclarativeNode.fiberFirstPaint] Layout root is nil")
			n.legacyPaint(ctx, buf)
			return
		}

		log.RenderLogger.Debug("[DeclarativeNode.fiberFirstPaint] LayoutBox root: children=%d",
			len(innerResult.Root.Children))

		// Phase 3: Paint - Convert LayoutResult to PaintablePlanes
		converter := NewFiberToPaintableConverter(fiberRoot)
		paintableLayout := converter.ConvertToLayout(innerResult.Root)

		if paintableLayout == nil || paintableLayout.Root == nil {
			log.RenderLogger.Debug("[DeclarativeNode.fiberFirstPaint] PaintableLayout result is nil")
			n.legacyPaint(ctx, buf)
			return
		}

		// Save paintable root for GetPaintableBoxes()
		n.mu.Lock()
		n.lastPaintableRoot = paintableLayout.Root
		n.mu.Unlock()

		// Build PaintablePlanes from PaintableLayout
		planes := paint.NewPaintablePlanes()
		var walkPaintable func(box *paint.PaintableBox)
		walkPaintable = func(box *paint.PaintableBox) {
			if box == nil {
				return
			}
			planes.AddToLayer(paint.RenderLayer(box.Layer), box)
			for _, child := range box.Children {
				walkPaintable(child)
			}
		}
		walkPaintable(paintableLayout.Root)

		log.RenderLogger.Debug("[DeclarativeNode.fiberFirstPaint] PaintablePlanes: %d boxes", planes.CountBoxes())

		// Paint using PaintablePlanes for proper layer Z-Ordering
		if err := n.paintEngine.PaintPaintablePlanes(planes, buf); err != nil {
			log.RenderLogger.Debug("[DeclarativeNode.fiberFirstPaint] Paint FAILED: %v, falling back", err)
			n.legacyPaint(ctx, buf)
			return
		}

		// Save HitMap for event routing
		// GetHitMap() converts layout.HitMap to event.HitMap with TargetFiber enrichment
		if hitMap := layoutResult.GetHitMap(); hitMap != nil {
			n.mu.Lock()
			n.fiberLastHitMap = hitMap
			n.mu.Unlock()
			log.RenderLogger.Debug("[DeclarativeNode.fiberFirstPaint] ✅ Saved HitMap with %d entries", hitMap.Size())
		} else {
			log.RenderLogger.Debug("[DeclarativeNode.fiberFirstPaint] ⚠️  HitMap is nil from layoutResult")
			n.mu.Lock()
			n.fiberLastHitMap = nil
			n.mu.Unlock()
		}

		log.RenderLogger.Debug("[DeclarativeNode.fiberFirstPaint] ✅ Render complete")
	} else {
		log.RenderLogger.Debug("[DeclarativeNode.fiberFirstPaint] Layout result type mismatch")
		n.legacyPaint(ctx, buf)
	}
}

// countFiberChildren counts the number of children in a Fiber tree
func countFiberChildren(fiber *rtui.Fiber) int {
	count := 0
	for child := fiber.Child; child != nil; child = child.Sibling {
		count++
	}
	return count
}

// comparePaint runs both legacy and Fiber-first paths and compares results
// Used for testing and validation during migration
//
// Deprecated: This method is only for testing purposes during migration to Fiber-first.
// It will be removed once migration is complete. Use fiberFirstPaint instead.
func (n *DeclarativeNode) comparePaint(ctx paint.PaintContext, buf *paint.Buffer) {
	log.RenderLogger.Debug("[DeclarativeNode.comparePaint] Running both paths for comparison")

	// Create separate buffers for each path
	legacyBuf := paint.NewBuffer(buf.Width, buf.Height)
	fiberFirstBuf := paint.NewBuffer(buf.Width, buf.Height)

	// Run legacy path
	n.legacyPaint(ctx, legacyBuf)

	// Run Fiber-first path
	n.fiberFirstPaint(ctx, fiberFirstBuf)

	// TODO: Compare buffers and log differences
	// For now, use the Fiber-first result
	copyBuffer(buf, fiberFirstBuf)

	log.RenderLogger.Debug("[DeclarativeNode.comparePaint] Comparison complete, using Fiber-first result")
}

// legacyPaint is the original VNode-based rendering path
//
// Deprecated: Use fiberFirstPaint with Fiber-first architecture instead.
// This method is kept for backward compatibility but should not be used in new code.
// The legacy path has the following issues:
//   - VNode is kept in memory during rendering (not discarded after Fiber reconciliation)
//   - Uses LayoutSwitcher which is deprecated
//   - Does not follow Fiber-first architecture where VNode is discarded after Fiber creation
//
// Migration: Set MINT_FIBER_FIRST=true and use NewDeclarativeNodeFromFuncWithFiber
func (n *DeclarativeNode) legacyPaint(ctx paint.PaintContext, buf *paint.Buffer) {
	// Debug logging
	log.PaintLogger.Debug("[DeclarativeNode.legacyPaint] START: useFiber=%v, reconciler=%v",
		n.useFiber, n.reconciler != nil)

	// Phase 1: Get the VNode tree
	if n.useFiber && n.reconciler != nil {
		// Fiber mode: just call render function directly for now
		// The reconciler's state management still happens through hooks
		log.PaintLogger.Debug("[DeclarativeNode.legacyPaint] ✅ Calling renderWithFiberContext (useFiber=%v, reconciler=%v)",
			n.useFiber, n.reconciler != nil)

		n.root = n.renderWithFiberContext()

		// Debug: Check if root is valid
		if n.root != nil {
			log.PaintLogger.Debug("[DeclarativeNode.Paint] n.root type=%d, tag=%s, children=%d",
				n.root.Type(), n.root.Tag(), len(n.root.Children()))
		} else {
			log.PaintLogger.Debug("[DeclarativeNode.Paint] n.root is NIL after renderWithFiberContext")
		}

	} else {
		// Non-Fiber mode
		log.RenderLogger.Debug("[DeclarativeNode.Paint] ⚠️  Using nonFiberRender (useFiber=%v, reconciler=%v)",
			n.useFiber, n.reconciler != nil)

		n.root = n.nonFiberRender()
	}

	if n.root == nil {
		log.RenderLogger.Debug("DeclarativeNode.Paint: root is nil, returning")

		return
	}

	// Phase 2: Apply focus state
	n.applyFocusState()

	// Phase 3: UNIFIED RENDERING - use PipelineRenderer with constraint-based layout
	log.RenderLogger.Debug("[DeclarativeNode.Paint] n.renderer = %v", n.renderer)
	if n.renderer != nil {
		log.RenderLogger.Debug("[DeclarativeNode.Paint] renderer type = %T", n.renderer)
	}

	log.PaintLogger.Debug("[DeclarativeNode.Paint] n.renderer=%v, isAdapter=%v", n.renderer != nil, n.renderer != nil && fmt.Sprintf("%T", n.renderer) == "*render.PipelineRendererAdapter")

	if n.renderer != nil {
		// Use the PaintContext dimensions as layout constraints (not buffer size)
		// The PaintContext.AvailableWidth/Height contains the user's configured layout size
		// while the buffer size may be larger (actual terminal size)
		if adapter, ok := n.renderer.(*PipelineRendererAdapter); ok {
			log.RenderLogger.Debug("[DeclarativeNode.Paint] ✅ Using PipelineRendererAdapter")
			log.RenderLogger.Debug("[DeclarativeNode.Paint] Layout constraints: %dx%d (buffer: %dx%d)",
				ctx.AvailableWidth, ctx.AvailableHeight, buf.Width, buf.Height)

			// Call RenderWithConstraints which will:
			// 1. Use PaintContext dimensions as BoxConstraints (user's configured layout size)
			// 2. Detect layer nodes and call RenderLayers() if needed
			// 3. Apply modal centering for LayerModal nodes using the correct layout size
			pipeline := adapter.GetPipeline()
			if err := pipeline.RenderWithConstraints(n.root, ctx.AvailableWidth, ctx.AvailableHeight, buf); err != nil {
				// Fallback to legacy rendering if pipeline fails
				log.RenderLogger.Debug("[DeclarativeNode.Paint] ❌ Pipeline render FAILED: %v, falling back to legacy", err)
				n.PaintVNode(n.root, ctx.Bounds.X, ctx.Bounds.Y, buf)
			} else {
				log.RenderLogger.Debug("[DeclarativeNode.Paint] ✅ Pipeline render SUCCESS")
				// NOTE: Inspector attachment is now handled by application layer
				// The demo calls inspector.AttachToApp() after reconciliation completes
				// This avoids circular dependency between render and framework packages
			}
		} else {
			// Use the generic renderer interface (old path)
			log.RenderLogger.Debug("[DeclarativeNode.Paint] ⚠️ Using generic renderer interface (old path)")

			n.renderer.Render(n.root, ctx.Bounds.X, ctx.Bounds.Y, buf)
		}
	} else {
		// Fallback to legacy painting
		log.RenderLogger.Debug("[DeclarativeNode.Paint] ⚠️ No renderer, using legacy PaintVNode")

		n.PaintVNode(n.root, ctx.Bounds.X, ctx.Bounds.Y, buf)
	}

	log.RenderLogger.Debug("DeclarativeNode.Paint: painting complete")

}

// renderWithFiberContext renders the VNode tree with Fiber hook context
// This ensures hooks work correctly in Fiber mode
func (n *DeclarativeNode) renderWithFiberContext() rtui.VNode {
	if n.renderFn == nil {
		return n.root
	}

	log.RenderLogger.Debug("[renderWithFiberContext] reconciler=%v", n.reconciler != nil)

	// The reconciler manages hook context through its render cycle
	// We capture the VNode tree during the render to avoid calling renderFn twice
	var capturedVNode rtui.VNode
	callCount := 0

	nullBuf := paint.NewBuffer(1, 1)
	n.reconciler.Render(paint.PaintContext{
		Bounds: paint.Rect{X: 0, Y: 0, Width: 1, Height: 1},
	}, nullBuf, func() rtui.VNode {
		callCount++
		vnode := n.renderFn()
		capturedVNode = vnode // Capture for PipelineRenderer
		if vnode != nil {
			log.PaintLogger.Debug("[renderWithFiberContext] renderFn call #%d: returned type=%d tag=%s children=%d", callCount, vnode.Type(), vnode.Tag(), len(vnode.Children()))
		} else {
			log.PaintLogger.Debug("[renderWithFiberContext] renderFn call #%d: returned nil", callCount)
		}

		return vnode
	})

	if capturedVNode != nil {
		log.PaintLogger.Debug("[renderWithFiberContext] FINAL capturedVNode type=%d tag=%s children=%d (total calls: %d)", capturedVNode.Type(), capturedVNode.Tag(), len(capturedVNode.Children()), callCount)
	} else {
		log.PaintLogger.Debug("[renderWithFiberContext] FINAL capturedVNode is nil (total calls: %d)", callCount)
	}

	// Return the captured VNode tree for PipelineRenderer
	return capturedVNode
}

// nonFiberRender renders the VNode tree in non-Fiber mode
func (n *DeclarativeNode) nonFiberRender() rtui.VNode {
	// Initialize component context if needed
	instanceCreated := false
	if n.instance == nil && n.renderFn != nil {
		n.instance = rtui.NewComponentContextForRoot()
		instanceCreated = true
	}

	if n.renderFn == nil {
		return n.root
	}

	// Set Intent Runtime to component context (enables intent.EmitIntents)
	n.mu.RLock()
	intentRuntime := n.intentRuntime
	n.mu.RUnlock()
	n.instance.SetIntentRuntime(intentRuntime)

	// Set StateSetter on dispatcher if this is a new instance
	if instanceCreated && intentRuntime != nil {
		intentRuntime.Dispatcher.SetStateSetter(n.instance)
		// Set schedule update callback
		n.instance.SetScheduleUpdate(func() {
			// Request re-render through scheduler (framework app)
			if n.scheduler != nil {
				n.scheduler.MarkDirty()
			}
		})
	}

	// Set component context for hooks
	n.instance.ResetContext()
	rtui.SetCurrentContext(n.instance)

	// Call render function
	n.root = n.renderFn()

	// Expand all ComponentVNodes recursively
	n.root = n.expandComponents(n.root)

	// Clear context
	rtui.SetCurrentContext(nil)

	return n.root
}

// applyFocusState applies focus state to the VNode tree
// In Fiber mode, focus is managed by the reconciler's FiberFocusManager
// IMPORTANT: Also syncs focusManager to framework.App for event routing
func (n *DeclarativeNode) applyFocusState() {
	if n.focusMgr == nil {
		return
	}

	// In Fiber mode, focus is managed by reconciler during commit phase
	// The FiberFocusManager is updated by reconciler.updateFocusManagerFromFiber()
	if n.useFiber {
		log.RenderLogger.Debug("DeclarativeNode.applyFocusState: Fiber mode, focus managed by reconciler")
		return
	}

	// Non-Fiber mode: legacy focus management via VNode tree
	if n.root == nil {
		return
	}

	var focusable []rtui.FocusableVNode

	// Check if there's a modal open - if so, trap focus in modal
	hasModal := rtui.HasModalInTree(n.root)

	if hasModal {
		// Focus trap: only collect focusable elements from modal layer
		focusable = rtui.CollectFocusableInLayer(n.root, rtui.LayerModal)
		log.RenderLogger.Debug("DeclarativeNode.Paint: modal detected, collected %d modal focusable nodes", len(focusable))
	} else {
		// No modal: collect all focusable elements
		focusable = rtui.CollectFocusable(n.root)
		log.RenderLogger.Debug("DeclarativeNode.Paint: no modal, collected %d focusable nodes", len(focusable))
	}

	// Legacy: This path is only for non-Fiber mode
	// For Fiber mode, focus is managed by FiberFocusManager which stores []*Fiber
	// We skip this update in Fiber mode
	if !n.useFiber {
		// Cast to VNodeFocusManager for legacy mode
		type legacyFocusManager interface {
			UpdateFocusableList([]rtui.FocusableVNode)
		}
		if legacyMgr, ok := interface{}(n.focusMgr).(legacyFocusManager); ok {
			legacyMgr.UpdateFocusableList(focusable)

			// Clamp focus index
			currentIndex := n.focusMgr.CurrentIndex()
			if currentIndex >= len(focusable) {
				currentIndex = len(focusable) - 1
			}
			if currentIndex < 0 && len(focusable) > 0 {
				currentIndex = 0
			}
			if currentIndex >= 0 {
				n.focusMgr.SetFocusByIndex(currentIndex)
			}
		}

		// Apply focus state
		n.applyFocus(focusable)
	}
}

// PaintVNode recursively paints a VNode and its children.
//
// DEPRECATED: This method accesses VNode during the Paint phase, which violates
// the Fiber-first architecture. Use PaintLayout with PaintableLayout instead.
//
// Migration path:
//  1. Use LayoutEngine.LayoutFiber() to get LayoutResult
//  2. Use FiberToPaintableConverter to convert to PaintableLayout
//  3. Use PaintEngine.PaintLayout() to render
//
// This method is kept for backward compatibility during migration.
// Set MINT_WARN_LEGACY=true environment variable to see deprecation warnings.
//
// Rendering strategy:
// 1. If node implements Paintable: use Paint(x, y) → []DrawCmd → write to buffer
// 2. Otherwise: handle by node type (Text/Element/Fragment)
// 3. Always process children for container nodes
func (n *DeclarativeNode) PaintVNode(vnode rtui.VNode, x, y int, buf *paint.Buffer) {
	if vnode == nil {
		return
	}

	// Deprecation warning (can be enabled via environment variable)
	log.RenderLogger.Warn("[DEPRECATED] PaintVNode is deprecated, use PaintLayout instead. VNode type=%d", vnode.Type())

	// Debug logging
	log.RenderLogger.Debug("[PaintVNode] vnode type=%d (%s), x=%d, y=%d, actual type=%T",
		vnode.Type(), vnode.Type(), x, y, vnode)

	// Set component bounds for mouse hit testing
	// Check if vnode implements SetBounds method
	if boundsSetter, ok := vnode.(interface{ SetBounds(x, y, width, height int) }); ok {
		// Measure the node to get its width and height
		width := n.MeasureVNodeWidth(vnode)
		height := n.MeasureVNodeHeight(vnode)
		boundsSetter.SetBounds(x, y, width, height)
		log.RenderLogger.Debug("[PaintVNode] SetBounds: x=%d, y=%d, width=%d, height=%d, type=%T",
			x, y, width, height, vnode)

	} else {
		log.RenderLogger.Debug("[PaintVNode] vnode does not implement SetBounds: type=%T", vnode)

	}

	// Check if vnode implements Paintable interface (custom rendering)
	if paintable, ok := vnode.(interface {
		Paint(int, int) []paint.DrawCmd
	}); ok {
		// Component has custom paint logic - use it
		commands := paintable.Paint(x, y)
		for _, cmd := range commands {
			buf.SetString(cmd.X, cmd.Y, cmd.Text, cmd.Style)
		}
		// Paintable components handle their own rendering, including children
		return
	}

	// Handle built-in VNode types
	switch vnode.Type() {
	case rtui.VNodeText:
		n.paintText(vnode, x, y, buf)

	case rtui.VNodeElement:
		n.paintElement(vnode, x, y, buf)

	case rtui.VNodeComponent:
		// Component nodes should be expanded before painting
		// In non-Fiber mode, expandComponents() handles this
		// In Fiber mode, components are already expanded in the Fiber tree

	case rtui.VNodeFragment:
		// Fragment - just paint children, no self-rendering
		n.paintChildren(vnode, x, y, buf)
	}

	// For non-Paintable elements, paint children after self-rendering
	if vnode.Type() == rtui.VNodeElement {
		// Check if this is a bordered element - render border + content
		if borderedNode, ok := vnode.(interface{ RenderBorder(int, int) []rtui.VNode }); ok {
			n.paintBordered(vnode, borderedNode, x, y, buf)
			return
		}

		// Check if this is a table element
		if tagger, ok := vnode.(interface{ Tag() string }); ok && tagger.Tag() == "table" {
			n.paintTable(vnode, x, y, buf)
			return
		}

		children := vnode.Children()
		if len(children) > 0 {
			// Use shared layout detection utility
			layoutInfo := rtui.GetLayoutInfo(vnode)
			isHStack := layoutInfo.IsHorizontal
			gap := layoutInfo.Gap

			if isHStack {
				// Horizontal layout: paint children on the same line with x offset
				childX := x
				for i, child := range children {
					childWidth := n.MeasureVNodeWidth(child)
					if log.BorderLogger.Enabled() {
						label := "?"
						if l, ok := child.(interface{ Tag() string }); ok {
							label = fmt.Sprintf("tag=%s", l.Tag())
						}
						log.RenderLogger.Debug("[HSTACK] child %d (%s): x=%d, width=%d, nextX=%d",
							i, label, childX, childWidth, childX+childWidth+gap)
					}
					n.PaintVNode(child, childX, y, buf)
					// Move X by the width of this child + gap
					childX += childWidth + gap
				}
			} else {
				// Vertical layout: paint children on different lines
				childY := y
				for _, child := range children {
					n.PaintVNode(child, x, childY, buf)
					childY += n.MeasureVNodeHeight(child)
				}
			}
		}
	}
}

// paintText paints a text VNode
func (n *DeclarativeNode) paintText(vnode rtui.VNode, x, y int, buf *paint.Buffer) {
	text := rtui.GetTextContent(vnode)
	if text != "" {
		buf.SetString(x, y, text, vnode.Style())
	}
}

// paintElement paints an element VNode
func (n *DeclarativeNode) paintElement(vnode rtui.VNode, x, y int, buf *paint.Buffer) {
	// Check if element has text content (for text elements created with ui.Text)
	if content := rtui.GetTextContent(vnode); content != "" {
		buf.SetString(x, y, content, vnode.Style())
		return // Don't paint children for text elements
	}
	// For non-text elements, children will be painted after the switch
}

// paintChildren paints children of a VNode
func (n *DeclarativeNode) paintChildren(vnode rtui.VNode, x, y int, buf *paint.Buffer) {
	children := vnode.Children()
	if len(children) == 0 {
		return
	}

	childY := y
	for _, child := range children {
		n.PaintVNode(child, x, childY, buf)
		childY += n.MeasureVNodeHeight(child)
	}
}

// paintTable paints a table element row by row
// Tables use row-based layout where each row (tr) contains cells (td)
// Cells in each row are painted horizontally, and rows are stacked vertically
func (n *DeclarativeNode) paintTable(vnode rtui.VNode, x, y int, buf *paint.Buffer) {
	rows := vnode.Children()
	if len(rows) == 0 {
		return
	}

	// Calculate column widths by measuring all cells in each column
	// First, find the maximum number of columns
	maxCols := 0
	for _, row := range rows {
		if tagger, ok := row.(interface{ Tag() string }); ok && tagger.Tag() == "tr" {
			cells := row.Children()
			if len(cells) > maxCols {
				maxCols = len(cells)
			}
		}
	}

	// Measure the width of each column
	colWidths := make([]int, maxCols)
	for _, row := range rows {
		if tagger, ok := row.(interface{ Tag() string }); ok && tagger.Tag() == "tr" {
			cells := row.Children()
			for colIdx, cell := range cells {
				if colIdx < maxCols {
					cellWidth := n.MeasureVNodeWidth(cell)
					if cellWidth > colWidths[colIdx] {
						colWidths[colIdx] = cellWidth
					}
				}
			}
		}
	}

	// Paint each row
	rowY := y
	for _, row := range rows {
		// Check if this is a table row (tr)
		if tagger, ok := row.(interface{ Tag() string }); ok && tagger.Tag() == "tr" {
			cells := row.Children()
			cellX := x
			for colIdx, cell := range cells {
				// Paint the cell
				n.PaintVNode(cell, cellX, rowY, buf)
				// Move X by the column width
				if colIdx < len(colWidths) {
					cellX += colWidths[colIdx]
				} else {
					cellX += n.MeasureVNodeWidth(cell)
				}
			}
			// Move to next row
			rowY += n.MeasureVNodeHeight(row)
		} else {
			// Non-row children are painted vertically
			n.PaintVNode(row, x, rowY, buf)
			rowY += n.MeasureVNodeHeight(row)
		}
	}
}

// paintBordered paints a bordered element with auto-rendered borders
// The border is rendered outside the content area
func (n *DeclarativeNode) paintBordered(vnode rtui.VNode, _ interface{ RenderBorder(int, int) []rtui.VNode }, x, y int, buf *paint.Buffer) {
	children := vnode.Children()
	if len(children) == 0 {
		return
	}

	child := children[0]
	if child == nil {
		return
	}

	// Measure content size
	contentHeight := n.MeasureVNodeHeight(child)

	// Get the measured total width of the BorderedNode (including border)
	borderedNodeWidth := 0
	if measurable, ok := vnode.(interface {
		Measure(runtime.BoxConstraints) runtime.Size
	}); ok {
		borderedNodeWidth = measurable.Measure(runtime.UnboundedConstraints()).Width
	}

	// Calculate inner content width
	// BorderedNode.Measure uses: totalWidth = max(childWidth, labelWidth) + borderWidth + (labelPresent ? 2 : 0)
	// So innerWidth = totalWidth - borderWidth - (labelPresent ? 2 : 0)
	hasLabel := false
	if l, ok := vnode.(interface{ GetBorderLabel() string }); ok && l.GetBorderLabel() != "" {
		hasLabel = true
	}

	innerWidth := borderedNodeWidth - 2 // Subtract border (1 left + 1 right)
	if hasLabel {
		innerWidth -= 2 // Subtract the extra padding added when label is present
	}
	contentWidth := max(0, innerWidth)

	// Also ensure contentWidth is at least the child width and label width
	childWidth := n.MeasureVNodeWidth(child)
	if childWidth > contentWidth {
		contentWidth = childWidth
	}
	if hasLabel {
		if l, ok := vnode.(interface{ GetBorderLabel() string }); ok {
			labelWidth := len(" " + l.GetBorderLabel() + " ")
			if labelWidth > contentWidth {
				contentWidth = labelWidth
			}
		}
	}

	// DEBUG: Log border painting info
	if log.BorderLogger.Enabled() {
		// Try to get a label for debugging
		label := "?"
		if l, ok := child.(interface{ Tag() string }); ok {
			label = l.Tag()
		}
		if l, ok := child.(interface{ GetBorderLabel() string }); ok && l.GetBorderLabel() != "" {
			label = l.GetBorderLabel()
		}
		log.RenderLogger.Debug("[BORDER] %s: x=%d, y=%d, contentW=%d, contentH=%d, totalH=%d",
			label, x, y, contentWidth, contentHeight, contentHeight+2)
		log.RenderLogger.Debug("[BORDER] Left border should be at col %d, rows %d-%d",
			x, y, y+contentHeight+1)
	}

	// Get border config from BorderedNode
	config := border.DefaultConfig()

	// Get style from BorderedNode (it returns rtui.BorderStyle)
	if bn, ok := vnode.(interface{ GetBorderStyle() rtui.BorderStyle }); ok {
		// Convert rtui.BorderStyle to border.Style
		// They have the same int values, so we can cast
		config.Style = border.Style(bn.GetBorderStyle())
	}
	if bn, ok := vnode.(interface{ GetBorderColor() string }); ok {
		config.Color = bn.GetBorderColor()
	}
	if bn, ok := vnode.(interface{ GetBorderLabel() string }); ok {
		config.Label = bn.GetBorderLabel()
	}

	// Create border renderer and paint
	renderer := border.WithConfig(config)

	// Paint border cells
	renderer.Paint(x, y, contentWidth, contentHeight, func(px, py int, ch rune, s style.Style) {
		if log.BorderLogger.Enabled() {
			// Log first few border cells for debugging
			if ch == '┌' || (px == x && py == y) {
				log.RenderLogger.Debug("[BORDER.Paint] cornerTL at (%d,%d): '%c'", px, py, ch)
			}
		}
		buf.SetCell(px, py, ch, s)
	})

	// Paint content inside border
	// Just paint once at the offset - the child (e.g., VStack) will handle
	// painting all its children at the correct relative positions
	offsetX, offsetY := renderer.GetContentOffset()
	n.PaintVNode(child, x+offsetX, y+offsetY, buf)
}

// MeasureVNodeWidth measures the width of a VNode for horizontal layout
// This method now prioritizes the Measurable interface over fallback logic
func (n *DeclarativeNode) MeasureVNodeWidth(vnode rtui.VNode) int {
	if vnode == nil {
		return 0
	}

	// PRIORITY 1: Use Measurable interface if available (new constraint-based measurement)
	if measurable, ok := vnode.(interface {
		Measure(constraints runtime.BoxConstraints) runtime.Size
	}); ok {
		// Use UnboundedConstraints to get natural size
		size := measurable.Measure(runtime.UnboundedConstraints())
		return size.Width
	}

	// PRIORITY 2: Check for table cell (td) - measure its content
	if tagger, ok := vnode.(interface{ Tag() string }); ok && tagger.Tag() == "td" {
		children := vnode.Children()
		if len(children) > 0 {
			return n.MeasureVNodeWidth(children[0])
		}
		return 0
	}

	// PRIORITY 3: Check if it's a bordered element - measure total width including border
	if borderedNode, ok := vnode.(interface{ GetBorderLabel() string }); ok {
		// BorderedNode now implements Measurable, but handle it here for compatibility
		children := vnode.Children()
		if len(children) > 0 {
			contentWidth := n.MeasureVNodeWidth(children[0])
			label := borderedNode.GetBorderLabel()
			// Calculate inner width (content or label, whichever is wider)
			innerWidth := contentWidth
			if label != "" {
				labelWidth := len(" " + label + " ")
				if labelWidth > innerWidth {
					innerWidth = labelWidth
				}
			}
			// Total width = border (left + right = 2) + inner width
			return innerWidth + 2
		}
	}

	// PRIORITY 4: Fallback logic for legacy components
	// Check the VNode type first
	switch vnode.Type() {
	case rtui.VNodeFragment:
		// Fragment is a virtual container - doesn't contribute its own width
		return 0

	case rtui.VNodeText:
		return len(rtui.GetTextContent(vnode))

	case rtui.VNodeElement:
		// Check if it's a button
		if btn, ok := vnode.(interface{ Label() string }); ok {
			// Button width matches Button.Measure:
			// - Label length
			// - +2 for brackets "[]"
			// - +2 for medium button padding (1 on each side)
			// - +1 for focus indicator
			return len(btn.Label()) + 5
		}
		// For elements, try text content
		if content := rtui.GetTextContent(vnode); content != "" {
			return len(content)
		}
		// Check for label in props
		if props := vnode.Props(); props != nil {
			if label, ok := props["label"].(string); ok {
				return len(label) + 2 // [label]
			}
		}
		// Don't return 0 - fall through to container handling for VStack/HStack
	}

	// PRIORITY 5: For containers, calculate width based on layout type
	// This handles VStack/HStack that don't have direct text content
	layoutInfo := rtui.GetLayoutInfo(vnode)
	if layoutInfo.IsHorizontal {
		// HStack: sum of children's widths
		width := 0
		for _, child := range vnode.Children() {
			width += n.MeasureVNodeWidth(child)
		}
		return width
	}
	// VStack: return maximum width of all children
	maxWidth := 0
	for _, child := range vnode.Children() {
		childWidth := n.MeasureVNodeWidth(child)
		if childWidth > maxWidth {
			maxWidth = childWidth
		}
	}
	return maxWidth
}

// MeasureVNodeHeight measures the height of a VNode for vertical layout
// This method now prioritizes the Measurable interface over fallback logic
func (n *DeclarativeNode) MeasureVNodeHeight(vnode rtui.VNode) int {
	if vnode == nil {
		return 0
	}

	// PRIORITY 1: Use Measurable interface if available (new constraint-based measurement)
	if measurable, ok := vnode.(interface {
		Measure(constraints runtime.BoxConstraints) runtime.Size
	}); ok {
		// Use UnboundedConstraints to get natural size
		size := measurable.Measure(runtime.UnboundedConstraints())
		return size.Height
	}

	// PRIORITY 2: Check for table row (tr) - height is max of cell heights
	if tagger, ok := vnode.(interface{ Tag() string }); ok && tagger.Tag() == "tr" {
		maxHeight := 0
		for _, cell := range vnode.Children() {
			cellHeight := n.MeasureVNodeHeight(cell)
			if cellHeight > maxHeight {
				maxHeight = cellHeight
			}
		}
		return maxHeight
	}

	// PRIORITY 3: Check for table cell (td) - height of its content
	if tagger, ok := vnode.(interface{ Tag() string }); ok && tagger.Tag() == "td" {
		children := vnode.Children()
		if len(children) > 0 {
			return n.MeasureVNodeHeight(children[0])
		}
		return 1
	}

	// PRIORITY 4: Check if it's a bordered element - measure total height including border
	if _, ok := vnode.(interface{ GetBorderLabel() string }); ok {
		// BorderedNode now implements Measurable, but handle it here for compatibility
		children := vnode.Children()
		if len(children) > 0 {
			contentHeight := n.MeasureVNodeHeight(children[0])
			// Total height = border (top + bottom = 2) + content height
			return contentHeight + 2
		}
	}

	// PRIORITY 5: Check for explicit height in props
	if props := vnode.Props(); props != nil {
		if h := props.GetInt("height"); h > 0 {
			return h
		}
	}

	// PRIORITY 6: Fallback logic for legacy components
	// Check if it's a button - buttons are single line
	if btn, ok := vnode.(interface{ Label() string }); ok && btn.Label() != "" {
		return 1
	}

	// Check the VNode type
	switch vnode.Type() {
	case rtui.VNodeText, rtui.VNodeElement:
		// Text elements are single line
		if rtui.GetTextContent(vnode) != "" {
			return 1
		}
		// For elements without text content, check if they're containers
		layoutInfo := rtui.GetLayoutInfo(vnode)
		if layoutInfo.IsHorizontal {
			// HStack: height is max of children's heights
			maxHeight := 0
			for _, child := range vnode.Children() {
				childHeight := n.MeasureVNodeHeight(child)
				if childHeight > maxHeight {
					maxHeight = childHeight
				}
			}
			return maxHeight
		} else {
			// VStack: height is sum of children's heights
			totalHeight := 0
			for _, child := range vnode.Children() {
				totalHeight += n.MeasureVNodeHeight(child)
			}
			return totalHeight
		}
	case rtui.VNodeFragment:
		// Fragment: height is sum of children's heights
		totalHeight := 0
		for _, child := range vnode.Children() {
			totalHeight += n.MeasureVNodeHeight(child)
		}
		return totalHeight
	}

	// Default height for leaf nodes
	return 1
}

// applyFocus applies focus state to VNodes based on the current focus index.
// This is called in non-Fiber mode to ensure the focused element is visually highlighted.
func (n *DeclarativeNode) applyFocus(focusable []rtui.FocusableVNode) {
	if n.focusMgr == nil {
		return
	}

	focusedIndex := n.focusMgr.CurrentIndex()
	if focusedIndex < 0 || focusedIndex >= len(focusable) {
		// No valid focus, clear all focus
		for _, elem := range focusable {
			elem.SetFocus(false)
		}
		return
	}

	// Set focus by index
	for i, elem := range focusable {
		if i == focusedIndex {
			log.RenderLogger.Debug("[applyFocus] setting focus=true on index %d (%s)", i, elem.GetFocusID())

			elem.SetFocus(true)
		} else {
			elem.SetFocus(false)
		}
	}
}

// expandComponents recursively expands all ComponentVNodes in the tree
// This is used in non-Fiber mode to ensure components are rendered with proper hook context
// The hook context must be set before calling this function
func (n *DeclarativeNode) expandComponents(vnode rtui.VNode) rtui.VNode {
	if vnode == nil {
		return nil
	}

	// If this is a ComponentVNode, expand it by calling its render function
	if vnode.Type() == rtui.VNodeComponent {
		if componentVNode, ok := vnode.(*rtui.ComponentVNode); ok {
			rendered := componentVNode.Render()
			if rendered != nil {
				// Recursively expand the rendered VNode
				return n.expandComponents(rendered)
			}
		}
		return nil
	}

	// For non-component nodes, we need to expand their children
	// Create a new VNode with expanded children
	switch v := vnode.(type) {
	case *rtui.ElementVNode:
		// Expand children of ElementVNode
		children := vnode.Children()
		if len(children) == 0 {
			return vnode
		}
		expandedChildren := make([]rtui.VNode, 0, len(children))
		for _, child := range children {
			expanded := n.expandComponents(child)
			if expanded != nil {
				expandedChildren = append(expandedChildren, expanded)
			}
		}
		// Create a clone with expanded children AND layer info
		cloned := rtui.NewElement(v.Tag())
		cloned.SetProps(vnode.Props().Clone()) // Clone to preserve _layer
		cloned.SetKey(vnode.Key())
		cloned.SetStyle(vnode.Style())
		cloned.SetChildren(expandedChildren)
		// Preserve layer information
		if layer := vnode.GetLayer(); layer != rtui.LayerBase {
			cloned.SetLayer(layer)
		}
		return cloned

	case *rtui.LayoutNode:
		// LayoutNode embeds ElementVNode, so similar handling
		children := vnode.Children()
		if len(children) == 0 {
			return vnode
		}
		expandedChildren := make([]rtui.VNode, 0, len(children))
		for _, child := range children {
			expanded := n.expandComponents(child)
			if expanded != nil {
				expandedChildren = append(expandedChildren, expanded)
			}
		}
		// For LayoutNode, we can't directly clone it because private fields
		// Instead, modify the existing node's children in place
		vnode.SetChildren(expandedChildren)
		return vnode

	case *rtui.FragmentVNode:
		// Fragment - expand children
		children := vnode.Children()
		if len(children) == 0 {
			return vnode
		}
		expandedChildren := make([]rtui.VNode, 0, len(children))
		for _, child := range children {
			expanded := n.expandComponents(child)
			if expanded != nil {
				expandedChildren = append(expandedChildren, expanded)
			}
		}
		return rtui.Fragment(expandedChildren...)

	case *rtui.BorderedNode:
		// BorderedNode embeds ElementVNode, need to expand children
		children := vnode.Children()
		if len(children) == 0 {
			return vnode
		}
		expandedChildren := make([]rtui.VNode, 0, len(children))
		for _, child := range children {
			expanded := n.expandComponents(child)
			if expanded != nil {
				expandedChildren = append(expandedChildren, expanded)
			}
		}
		// Modify in place like LayoutNode since we can't clone BorderedNode easily
		// Layer info should be preserved in the props
		vnode.SetChildren(expandedChildren)
		return vnode

	default:
		// For TextVNode and other leaf nodes, return as-is
		return vnode
	}
}

// =============================================================================
// framework.Mountable interface implementation
// =============================================================================

// Mount mounts this node to a parent container
func (n *DeclarativeNode) Mount(parent component.Container) {
	n.mu.Lock()
	defer n.mu.Unlock()
	// DeclarativeNode doesn't need explicit mount handling
	// The VNode tree is self-contained
}

// Unmount unmounts this node from its parent
func (n *DeclarativeNode) Unmount() {
	n.mu.Lock()
	defer n.mu.Unlock()
	// Cleanup any resources
	n.root = nil
}

// IsMounted returns whether this node is mounted
func (n *DeclarativeNode) IsMounted() bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.root != nil
}

// =============================================================================
// VNode access
// =============================================================================

// Root returns the root VNode
func (n *DeclarativeNode) Root() rtui.VNode {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.root
}

// UpdateRoot updates the root VNode
func (n *DeclarativeNode) UpdateRoot(vnode rtui.VNode) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.root = vnode
}



// =============================================================================
// Test Helper Methods
// =============================================================================

// GetFocusManager returns the focus manager for this declarative node
// Returns *FiberFocusManager in Fiber mode, interface{} for compatibility
func (n *DeclarativeNode) GetFocusManager() *rtui.FiberFocusManager {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.focusMgr
}

// GetRenderer returns the VNode renderer for this declarative node
func (n *DeclarativeNode) GetRenderer() rtui.VNodeRenderer {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.renderer
}

// GetHitMap returns the HitMap from the most recent render
// This HitMap contains the FINAL positions after all layout transforms (including layer centering)
// Returns nil if the node hasn't been rendered yet or doesn't use the RenderingPipeline
func (n *DeclarativeNode) GetHitMap() *event.HitMap {
	n.mu.RLock()
	defer n.mu.RUnlock()

	// Priority 1: Check fiberFirstPaint path (Fiber-first mode)
	if n.fiberLastHitMap != nil {
		log.RenderLogger.Debug("[DeclarativeNode.GetHitMap] Returning fiberFirstPaint HitMap with %d entries", n.fiberLastHitMap.Size())
		return n.fiberLastHitMap
	}

	// Priority 2: Check if renderer is PipelineRendererAdapter (Legacy path)
	if adapter, ok := n.renderer.(*PipelineRendererAdapter); ok {
		// Get the RenderingPipeline from the adapter
		pipeline := adapter.GetRenderingPipeline()
		if pipeline != nil {
			hitMap := pipeline.GetHitMap()
			if hitMap != nil {
				log.RenderLogger.Debug("[DeclarativeNode.GetHitMap] Returning RenderingPipeline HitMap with %d entries", hitMap.Size())
			} else {
				log.RenderLogger.Debug("[DeclarativeNode.GetHitMap] RenderingPipeline returned nil HitMap")
			}
			return hitMap
		}
	}

	// No HitMap available
	log.RenderLogger.Debug("[DeclarativeNode.GetHitMap] No HitMap available (fiberFirstPaint=%v, renderer=%T)", n.fiberLastHitMap != nil, n.renderer)
	return nil
}

// GetFocusedIndex returns the index of the currently focused element
func (n *DeclarativeNode) GetFocusedIndex() int {
	n.mu.RLock()
	defer n.mu.RUnlock()
	if n.focusMgr == nil {
		return -1
	}
	return n.focusMgr.CurrentIndex()
}

// GetFocusedType returns the type of the currently focused element
func (n *DeclarativeNode) GetFocusedType() int {
	n.mu.RLock()
	defer n.mu.RUnlock()
	if n.focusMgr == nil {
		return 0
	}
	current := n.focusMgr.GetCurrent()
	if current == nil {
		return 0
	}
	// Return VNodeType as int (Fiber-first: current is *Fiber)
	return int(current.Type)
}

// getFrameworkApp returns the framework app (for triggering re-renders)
// Deprecated: Use framework.App to access FocusManager directly
func (n *DeclarativeNode) getFrameworkApp() reconciler.Scheduler {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.scheduler
}


func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// renderFiberToBuffer renders a single Fiber to the buffer (Fiber-first)
func renderFiberToBuffer(fiber *reconciler.Fiber, x, y int, buffer *paint.Buffer) {
	if fiber == nil {
		return
	}

	// For text nodes, render text content
	if fiber.Type == rtui.VNodeText {
		content := ""
		if fiber.MemoizedState != nil {
			if s, ok := fiber.MemoizedState.(string); ok {
				content = s
			}
		} else if fiber.Props != nil {
			if s, ok := fiber.Props["content"].(string); ok {
				content = s
			}
		}
		if content != "" {
			buffer.SetString(x, y, content, fiber.Style)
		}
		return
	}

	// For elements with children, render would be handled by the reconciler
	// This callback is for leaf nodes mainly
}

// =============================================================================
// Layer Event Handling
// =============================================================================

// handleLayerKeyEvent handles keyboard events for layer components (e.g., ESC to close modal)
// Returns true if the event was handled
func (n *DeclarativeNode) handleLayerKeyEvent(root rtui.VNode) bool {
	modalNode := n.findModalNode(root)
	if modalNode == nil {
		return false
	}

	// Check if this modal should close on ESC
	props := modalNode.Props()
	if props == nil {
		return false
	}

	closeOnESC := true // Default to true
	if v, ok := props["_closeOnESC"].(bool); ok {
		closeOnESC = v
	}

	if !closeOnESC {
		return false
	}

	// Trigger the OnClose callback
	if onClose, ok := props["_onClose"].(func()); ok {
		onClose()
		return true
	}

	return false
}

// findModalNode recursively searches for a modal node in the VNode tree
func (n *DeclarativeNode) findModalNode(vnode rtui.VNode) rtui.VNode {
	if vnode == nil {
		return nil
	}

	// Check if this node is a modal
	if vnode.GetLayer() == rtui.LayerModal {
		return vnode
	}

	// Recursively check children
	for _, child := range vnode.Children() {
		if modal := n.findModalNode(child); modal != nil {
			return modal
		}
	}

	return nil
}

// requestRender triggers a re-render
func (n *DeclarativeNode) requestRender(useFiber bool, reconciler rtui.Reconciler) {
	if useFiber && reconciler != nil {
		// Fiber mode: schedule reconciler update
		if r, ok := reconciler.(*fiberReconcilerAdapter); ok {
			r.r.ScheduleUpdate(rtui.LaneSyncLane)
		}
	} else {
		// Non-Fiber mode: mark framework app as dirty to trigger re-render
		if fwApp := n.getFrameworkApp(); fwApp != nil {
			fwApp.MarkDirty()
		}
	}
}

// GetHooks returns the HookManager for registering VNode transformation hooks.
// This delegates to the underlying PipelineRendererAdapter.
func (n *DeclarativeNode) GetHooks() *render.HookManager {
	n.mu.RLock()
	defer n.mu.RUnlock()

	// Get the renderer (should be PipelineRendererAdapter)
	if adapter, ok := n.renderer.(*PipelineRendererAdapter); ok {
		return adapter.GetHooks()
	}

	return nil
}

// GetFiberRoot returns the Fiber root from the Fiber reconciler
// This allows the rendering pipeline to access the Fiber tree for NodeID propagation
// Phase 8: Fiber to Layout Engine NodeID propagation
func (n *DeclarativeNode) GetFiberRoot() *reconciler.Fiber {
	n.mu.RLock()
	defer n.mu.RUnlock()

	// Access the underlying reconciler to get Fiber root
	if adapter, ok := n.reconciler.(*fiberReconcilerAdapter); ok {
		return adapter.GetFiberRoot()
	}

	return nil
}

// RenderWithFiber renders with explicit Fiber tree for NodeID propagation
// Phase 8: Entry point for DeclarativeNode Fiber-based rendering
// This method wraps the renderer's RenderWithFiber, providing access to the Fiber tree
//
// Parameters:
//
//	buffer: Paint buffer for rendering
func (n *DeclarativeNode) RenderWithFiber(buffer *paint.Buffer) error {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if n.reconciler == nil || n.renderer == nil {
		// No reconciler or renderer - can't use Fiber-based rendering
		log.RenderLogger.Debug("[RenderWithFiber] No reconciler or renderer available")
		return fmt.Errorf("no reconciler or renderer for Fiber-based rendering")
	}

	// Get the renderer and call its RenderWithFiber method
	// This allows NodeID propagation from Fiber tree to ComputedBox
	if adapter, ok := n.renderer.(*PipelineRendererAdapter); ok {
		return adapter.RenderWithFiber(n.root, buffer)
	}

	// Fallback for non-PipelineRenderer types
	log.RenderLogger.Debug("[RenderWithFiber] No PipelineRendererAdapter available")
	return fmt.Errorf("no PipelineRendererAdapter for Fiber-based rendering")
}

// GetInstanceManager returns the Fiber Reconciler's InstanceManager
// GetInstanceManager returns the InstanceManager (deprecated: use GetComponentInstances instead)
// This allows external code to access ComponentInstances managed by the Fiber Reconciler
func (n *DeclarativeNode) GetInstanceManager() interface{} {
	n.mu.RLock()
	defer n.mu.RUnlock()

	// The reconciler is a rtui.Reconciler interface
	// We need to access the underlying *reconciler.Reconciler to get InstanceManager
	if adapter, ok := n.reconciler.(*fiberReconcilerAdapter); ok {
		return adapter.r.GetInstanceManager()
	}

	return nil
}

// GetComponentInstances returns a map of NodeID to ComponentInstance
// This provides direct access to component instances for HitMap enrichment
// without using reflection
func (n *DeclarativeNode) GetComponentInstances() map[uint64]rtui.ComponentInstance {
	n.mu.RLock()
	defer n.mu.RUnlock()

	// Access the underlying reconciler to get InstanceManager
	if adapter, ok := n.reconciler.(*fiberReconcilerAdapter); ok {
		instanceMgr := adapter.r.GetInstanceManager()
		if instanceMgr != nil {
			return instanceMgr.GetAllInstancesByID()
		}
	}

	return nil
}

// GetAllInteractionInstances returns all instances that implement interaction interfaces
// This is used by App to register instances with InteractionContext
func (n *DeclarativeNode) GetAllInteractionInstances() map[int]interface{} {
	n.mu.RLock()
	defer n.mu.RUnlock()

	// Access the underlying reconciler to get instances
	if adapter, ok := n.reconciler.(*fiberReconcilerAdapter); ok {
		return adapter.GetAllInteractionInstances()
	}

	return nil
}

// getFiberRoot returns the Fiber root from the reconciler (internal, no lock)
// This is used by fiberFirstPaint which already holds the lock
func (n *DeclarativeNode) getFiberRoot() *reconciler.Fiber {
	if adapter, ok := n.reconciler.(*fiberReconcilerAdapter); ok {
		return adapter.GetFiberRoot()
	}
	return nil
}

// copyBuffer copies content from source buffer to destination buffer
func copyBuffer(dst, src *paint.Buffer) {
	if dst == nil || src == nil {
		return
	}
	// Copy cell by cell
	for y := 0; y < minInt(dst.Height, src.Height); y++ {
		for x := 0; x < minInt(dst.Width, src.Width); x++ {
			cell := src.GetContent(x, y)
			// SetCell takes rune, but Cell.Cluster is a string
			// Use first rune of cluster, or space if empty
			var char rune
			if len(cell.Cluster) > 0 {
				char = []rune(cell.Cluster)[0]
			} else {
				char = ' '
			}
			dst.SetCell(x, y, char, cell.Style)
		}
	}
}

// GetPaintableRoot returns the root of the paintable boxes tree from the last paint computation
// This provides access to the computed hierarchical paintable structure
func (n *DeclarativeNode) GetPaintableRoot() *paint.PaintableBox {
	n.mu.RLock()
	defer n.mu.RUnlock()

	return n.lastPaintableRoot
}



// =============================================================================
// Portal Support - Helper Methods
// =============================================================================

// hasPortals checks if the Fiber tree contains any Portal nodes
func (n *DeclarativeNode) hasPortals(fiber *reconciler.Fiber) bool {
	if fiber == nil {
		return false
	}

	found := false

	var checkNode func(f *reconciler.Fiber)
	checkNode = func(f *reconciler.Fiber) {
		if f == nil || found {
			return
		}

		// Check if this is a Portal node
		if f.Props != nil {
			if _, ok := f.Props["portalRoot"].(string); ok {
				found = true
				return
			}
			// Also check for PortalRoot
			if _, ok := f.Props["portalRootId"].(string); ok {
				found = true
				return
			}
		}

		checkNode(f.Child)
		checkNode(f.Sibling)
	}

	checkNode(fiber)
	return found
}

// SetUsePortalLayout enables or disables Portal-aware layout
func (n *DeclarativeNode) SetUsePortalLayout(enabled bool) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.usePortalLayout = enabled
}

// IsPortalLayoutEnabled returns whether Portal-aware layout is enabled
func (n *DeclarativeNode) IsPortalLayoutEnabled() bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.usePortalLayout
}