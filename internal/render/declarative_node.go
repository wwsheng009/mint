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
	"github.com/wwsheng009/mint/runtime/event"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/render"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// DeclarativeNode - Bridges VNode with framework Component
// =============================================================================
// DeclarativeNode allows a VNode tree to be used as a framework Component.
// This enables mixing declarative UI (VNode) with imperative Components.

// RenderMode specifies the declarative rendering path.
//
// Fiber-first is the only active path. The legacy value is retained only for
// source compatibility with older callers that still pass it to SetRenderMode.
type RenderMode int

const (
	// RenderModeLegacy is deprecated and no longer selects a rendering path.
	// SetRenderMode normalizes it to RenderModeFiberFirst.
	RenderModeLegacy RenderMode = iota
	// RenderModeFiberFirst uses the new Fiber-first rendering pipeline
	// This is the default and recommended rendering mode.
	RenderModeFiberFirst
)

// DeclarativeNode wraps a VNode function for use as a framework Component
type DeclarativeNode struct {
	mu       sync.RWMutex
	root     rtui.VNode              // The root VNode of this tree (legacy, will be removed)
	renderFn rtui.ComponentFunc      // Function that renders the VNode
	instance *rtui.ComponentContext  // Component instance for hooks
	focusMgr *rtui.FiberFocusManager // Focus manager for keyboard navigation (Fiber-first)

	// Framework integration
	reconciler rtui.Reconciler    // Fiber reconciler (if enabled) - use interface to avoid import cycle
	renderer   rtui.VNodeRenderer // VNode renderer (implements VNodeRenderer interface)

	// Scheduler fallback for global render requests before the reconciler is available.
	scheduler reconciler.Scheduler

	// === Intent Integration ===
	intentRuntime *intent.Runtime // Intent runtime for dispatching intents

	// === Fiber-first Rendering Pipeline (Phase 4) ===
	newLayoutEngine *NewLayoutEngineAdapter // New layout engine (Fiber-first only)
	paintEngine     *PaintEngine            // Paint engine for Fiber-first
	converter       *FiberToPaintableConverter

	// === Portal Support (Two-Phase Layout) ===
	portalLayoutEngine *PortalAwareLayoutEngine // Portal-aware layout engine (Phase 5)
	usePortalLayout    bool                     // Whether to use Portal-aware layout

	// === HitMap Storage ===
	// Fiber-first mode stores HitMap here (separate from RenderingPipeline)
	fiberLastHitMap *event.HitMap // HitMap from fiberFirstPaint() path

	// === Layout Result Storage ===
	lastLayoutResult *layout.LayoutResult // Last layout computation result (for GetLayoutBoxes)

	// === Paintable Result Storage ===
	lastPaintableRoot    *paint.PaintableBox // Last paintable layout result (for GetPaintableBoxes)
	lastPaintDirtyRects  []paint.Rect        // Dirty rect hints from last paintable tree
	lastSceneImageLayers []paint.ImageLayer  // Raster image layers collected during paint traversal

	// === Portal Box Debug Storage ===
	lastPortalBoxes []*layout.LayoutBox // Portal boxes from last layout (for debugging)
}

// SetScheduler sets the scheduler for requesting frame updates
// The Fiber reconciler owns normal update scheduling; this fallback scheduler
// is kept for global render requests before a reconciler update can be issued.
func (n *DeclarativeNode) SetScheduler(scheduler reconciler.Scheduler) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.scheduler = scheduler
}

// SetApp sets the framework app (scheduler) on the DeclarativeNode and its reconciler.
//
// OPTIONAL - This method is only required when:
//   - Running an interactive application (e.g., ui.Run)
//   - State-driven re-rendering is needed (e.g., setState, Intent handlers)
//   - The reconciler's requestWork() needs to trigger MarkDirty()
//
// NOT REQUIRED for:
//   - Static rendering / one-time painting
//   - Measurement / layout calculation
//   - Unit tests that only call Paint() once
//
// The reconciler internally handles nil scheduler safely (requestWorks() checks before calling MarkDirty)
//
// This is called from ui.Run to enable frame scheduling.
func (n *DeclarativeNode) SetApp(app interface{}) {
	// Keep a scheduler fallback for global render requests.
	if scheduler, ok := app.(reconciler.Scheduler); ok {
		n.SetScheduler(scheduler)
	}

	// Set app on the reconciler (for Fiber mode)
	if n.reconciler != nil {
		if reconciler, ok := n.reconciler.(interface{ SetApp(interface{}) }); ok {
			reconciler.SetApp(app)
		}
	}

	rtui.SetGlobalRenderScheduler(func() {
		n.mu.RLock()
		reconciler := n.reconciler
		scheduler := n.scheduler
		n.mu.RUnlock()

		if reconciler != nil {
			if adapter, ok := reconciler.(*fiberReconcilerAdapter); ok {
				adapter.r.ScheduleUpdate(rtui.LaneSyncLane)
				return
			}
		}
		if scheduler != nil {
			scheduler.MarkDirty()
		}
	})
}

// NewDeclarativeNodeFromFuncWithFiber creates a new declarative node with Fiber reconciler enabled
// This is the default and recommended way to create declarative nodes.
//
// Default enabled features:
//   - Fiber-first mode (reconciliation + layout + paint pipeline)
//   - Portal-aware layout (two-phase layout for Portal components)
//
// Portal-aware layout can be disabled via environment variable for compatibility:
//   - MINT_PORTAL_LAYOUT=false or MINT_PORTAL_LAYOUT=0
//
// Scheduler must be set later via SetApp() to enable frame scheduling (optional for static rendering).
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
	}

	// Initialize Fiber-first pipeline components
	node.initFiberFirstPipeline()
	node.initPortalLayoutSupport()

	return node
}

// initFiberFirstPipeline initializes the Fiber-first rendering pipeline components
// Fiber-first mode is always enabled; no environment variable check needed.
func (n *DeclarativeNode) initFiberFirstPipeline() {
	log.RenderLogger.IfEnabled().Debug("[DeclarativeNode] Fiber-first mode ENABLED (default)")

	// Use the new layout engine directly (runtime/layout), bypassing LayoutSwitcher
	// This ensures we never go through the compute path
	n.newLayoutEngine = NewNewLayoutEngineAdapter()
	n.paintEngine = NewPaintEngine()
	// converter will be created with Fiber root during render
}

// initPortalLayoutSupport initializes Portal-aware layout configuration
// Portal-aware layout uses two-phase layout: main tree first, then portal overlays separately
func (n *DeclarativeNode) initPortalLayoutSupport() {
	// Portal-aware layout is now enabled by default
	// Can be disabled via MINT_PORTAL_LAYOUT=false or MINT_PORTAL_LAYOUT=0 for backward compatibility
	portalLayoutEnv := os.Getenv("MINT_PORTAL_LAYOUT")
	if portalLayoutEnv == "false" || portalLayoutEnv == "0" {
		n.usePortalLayout = false
		log.RenderLogger.IfEnabled().Debug("[DeclarativeNode] Portal-aware layout DISABLED (MINT_PORTAL_LAYOUT=%s)", portalLayoutEnv)
	} else {
		n.usePortalLayout = true
		log.RenderLogger.IfEnabled().Debug("[DeclarativeNode] Portal-aware layout ENABLED (default or MINT_PORTAL_LAYOUT=%s)", portalLayoutEnv)
	}
}

// SetRenderMode preserves the historical API while keeping the active renderer
// Fiber-first. Legacy mode has been removed, so non-Fiber values are ignored.
func (n *DeclarativeNode) SetRenderMode(mode RenderMode) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if mode != RenderModeFiberFirst {
		log.RenderLogger.IfEnabled().Warn("[DeclarativeNode.SetRenderMode] legacy render mode %d is no longer supported; using Fiber-first", mode)
	}
	// Use the new layout engine directly (runtime/layout), bypassing LayoutSwitcher.
	if n.newLayoutEngine == nil {
		n.newLayoutEngine = NewNewLayoutEngineAdapter()
	}
	if n.paintEngine == nil {
		n.paintEngine = NewPaintEngine()
	}
}

// GetRenderMode returns the active rendering mode.
func (n *DeclarativeNode) GetRenderMode() RenderMode {
	return RenderModeFiberFirst
}

// IsFiberFirstEnabled returns whether Fiber-first mode is enabled.
func (n *DeclarativeNode) IsFiberFirstEnabled() bool {
	return true
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

// Paint renders the declarative tree using the Fiber-first pipeline.
func (n *DeclarativeNode) Paint(ctx paint.PaintContext, buf *paint.Buffer) {
	// Acquire read lock to read state needed for rendering.
	n.mu.RLock()
	reconciler := n.reconciler
	n.mu.RUnlock()

	// Debug logging
	log.PaintLogger.Debug("[DeclarativeNode.Paint] START: ctx.X=%d, ctx.Y=%d, buf=%dx%d, hasReconciler=%v",
		ctx.Bounds.X, ctx.Bounds.Y, buf.Width, buf.Height, reconciler != nil)

	if reconciler != nil {
		n.fiberFirstPaint(ctx, buf)
		return
	}

	log.PaintLogger.IfEnabled().Warn("[DeclarativeNode.Paint] Fiber-first renderer is not initialized; render skipped")
}

// PaintScene renders the text buffer using the existing Fiber-first paint
// pipeline and then collects any raster image layers exposed by runtime
// instances. Returning nil preserves text-only behavior for callers.
func (n *DeclarativeNode) PaintScene(ctx paint.PaintContext, buf *paint.Buffer) *paint.SceneFrame {
	n.Paint(ctx, buf)

	layers := n.collectSceneImageLayers()
	if len(layers) == 0 {
		return nil
	}

	scene := paint.NewSceneFrame(buf)
	scene.ImageLayers = layers
	scene.Diagnostics = paint.SceneDiagnostics{
		Summary: fmt.Sprintf("declarative scene images=%d", len(layers)),
		Notes:   []string{"source=declarative-fiber"},
	}
	return scene
}

// fiberFirstPaint renders using the new Fiber-first pipeline
// Phase 1: Reconcile (VNode -> Fiber, VNode discarded)
// Phase 2: Layout (Fiber -> LayoutBox, no VNode access)
// Phase 3: Paint (LayoutBox -> PaintableBox -> Buffer)
func (n *DeclarativeNode) fiberFirstPaint(ctx paint.PaintContext, buf *paint.Buffer) {
	log.RenderLogger.IfEnabled().Debug("[DeclarativeNode.fiberFirstPaint] === STARTING Fiber-first render ===")
	// Phase 1: Fiber Reconciliation
	// The reconciler updates the Fiber tree, VNode is discarded after this
	// Use a minimal buffer for reconciliation (actual painting happens later)
	nullBuf := paint.NewBuffer(1, 1)
	renderFn := n.renderFn
	if hooks := n.GetHooks(); hooks != nil && hooks.VNodeHookCount() > 0 {
		base := renderFn
		renderFn = func() rtui.VNode {
			return hooks.ApplyVNodeHooks(base())
		}
	}
	n.reconciler.Render(paint.PaintContext{
		Bounds: paint.Rect{X: 0, Y: 0, Width: 1, Height: 1},
	}, nullBuf, renderFn)

	// Get the Fiber root from reconciler
	fiberRoot := n.getFiberRoot()
	if fiberRoot == nil {
		log.PaintLogger.IfEnabled().Error("[DeclarativeNode.fiberFirstPaint] Fiber root is nil, render aborted")
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
		log.RenderLogger.IfEnabled().Debug("[DeclarativeNode.fiberFirstPaint] ✅ Using Portal-aware layout engine")

		if n.portalLayoutEngine == nil {
			n.portalLayoutEngine = NewPortalAwareLayoutEngine()
		}

		// Perform two-phase layout using PortalAwareLayoutEngine
		mainResult, portalBoxes, layoutErr := n.portalLayoutEngine.Layout(fiberRoot, constraints)

		if layoutErr != nil {
			log.PaintLogger.IfEnabled().Error("[DeclarativeNode.fiberFirstPaint] Portal layout FAILED: %v, render aborted", layoutErr)
			return
		}

		// Merge portal boxes into main result and store for debugging
		n.mu.Lock()
		if len(portalBoxes) > 0 {
			log.RenderLogger.IfEnabled().Debug("[DeclarativeNode.fiberFirstPaint] Merged %d Portal boxes into layout", len(portalBoxes))
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
			log.PaintLogger.IfEnabled().Error("[DeclarativeNode.fiberFirstPaint] Layout FAILED: %v, render aborted", err)
			return
		}
	}

	log.RenderLogger.IfEnabled().Debug("[DeclarativeNode.fiberFirstPaint] Layout complete")

	// Get the layout result
	if adapter, ok := layoutResult.(*newLayoutResultAdapter); ok {
		// Save layout result for GetLayoutBoxes()
		innerResult := adapter.result
		n.mu.Lock()
		n.lastLayoutResult = innerResult
		n.mu.Unlock()

		if innerResult == nil || innerResult.Root == nil {
			log.PaintLogger.IfEnabled().Error("[DeclarativeNode.fiberFirstPaint] Layout root is nil, render aborted")
			return
		}

		log.RenderLogger.Debug("[DeclarativeNode.fiberFirstPaint] LayoutBox root: children=%d",
			len(innerResult.Root.Children))

		// Phase 3: Paint - Convert LayoutResult to PaintablePlanes
		converter := NewFiberToPaintableConverter(fiberRoot)
		paintableLayout := converter.ConvertToLayout(innerResult.Root)

		if paintableLayout == nil || paintableLayout.Root == nil {
			log.PaintLogger.IfEnabled().Error("[DeclarativeNode.fiberFirstPaint] PaintableLayout result is nil, render aborted")
			return
		}

		// Build PaintablePlanes from PaintableLayout
		planes := paint.NewPaintablePlanes()
		dirtyRects := make([]paint.Rect, 0, 8)
		sceneLayers := make([]paint.ImageLayer, 0, 4)
		walkPaintableBoxesByEffectiveLayer(paintableLayout.Root, func(layer paint.RenderLayer, box *paint.PaintableBox) {
			planes.AddToLayer(layer, box)
			collectPaintableDirtyRectFromBox(box, &dirtyRects)
			collectSceneImageLayerFromBox(box, &sceneLayers)
		})

		// Save paintable root for GetPaintableBoxes().
		n.mu.Lock()
		n.lastPaintableRoot = paintableLayout.Root
		n.lastPaintDirtyRects = dirtyRects
		n.lastSceneImageLayers = sceneLayers
		n.mu.Unlock()

		log.RenderLogger.IfEnabled().Debug("[DeclarativeNode.fiberFirstPaint] PaintablePlanes: %d boxes", planes.CountBoxes())

		// Paint using PaintablePlanes for proper layer Z-Ordering
		if err := n.paintEngine.PaintPaintablePlanes(planes, buf); err != nil {
			log.PaintLogger.IfEnabled().Error("[DeclarativeNode.fiberFirstPaint] Paint FAILED: %v, render aborted", err)
			return
		}

		// Save HitMap for event routing
		// GetHitMap() converts layout.HitMap to event.HitMap with TargetFiber enrichment
		if hitMap := layoutResult.GetHitMap(); hitMap != nil {
			n.mu.Lock()
			n.fiberLastHitMap = hitMap
			n.mu.Unlock()
			log.RenderLogger.IfEnabled().Debug("[DeclarativeNode.fiberFirstPaint] ✅ Saved HitMap with %d entries", hitMap.Size())
		} else {
			log.RenderLogger.IfEnabled().Debug("[DeclarativeNode.fiberFirstPaint] ⚠️  HitMap is nil from layoutResult")
			n.mu.Lock()
			n.fiberLastHitMap = nil
			n.mu.Unlock()
		}

		log.RenderLogger.IfEnabled().Debug("[DeclarativeNode.fiberFirstPaint] ✅ Render complete")
	} else {
		log.PaintLogger.IfEnabled().Error("[DeclarativeNode.fiberFirstPaint] Layout result type mismatch, render aborted")
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
	if disposable, ok := n.reconciler.(interface{ Dispose() }); ok {
		disposable.Dispose()
	}
	if n.instance != nil {
		n.instance.CleanupAll()
	}
	if n.paintEngine != nil {
		n.paintEngine.InvalidateCache()
	}
	if n.newLayoutEngine != nil {
		n.newLayoutEngine.ClearCache()
	}
	if n.converter != nil {
		n.converter = nil
	}
	n.fiberLastHitMap = nil
	n.lastLayoutResult = nil
	n.lastPaintableRoot = nil
	n.lastPaintDirtyRects = nil
	n.lastSceneImageLayers = nil
	n.lastPortalBoxes = nil
	rtui.SetGlobalRenderScheduler(nil)
	n.root = nil
	n.renderFn = nil
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
		log.RenderLogger.IfEnabled().Debug("[DeclarativeNode.GetHitMap] Returning fiberFirstPaint HitMap with %d entries", n.fiberLastHitMap.Size())
		return n.fiberLastHitMap
	}

	// Priority 2: Check the renderer pipeline for compatibility callers that
	// invoked PipelineRenderer directly.
	if adapter, ok := n.renderer.(*PipelineRendererAdapter); ok {
		// Get the RenderingPipeline from the adapter
		pipeline := adapter.GetRenderingPipeline()
		if pipeline != nil {
			hitMap := pipeline.GetHitMap()
			if hitMap != nil {
				log.RenderLogger.IfEnabled().Debug("[DeclarativeNode.GetHitMap] Returning RenderingPipeline HitMap with %d entries", hitMap.Size())
			} else {
				log.RenderLogger.IfEnabled().Debug("[DeclarativeNode.GetHitMap] RenderingPipeline returned nil HitMap")
			}
			return hitMap
		}
	}

	// No HitMap available
	log.RenderLogger.IfEnabled().Debug("[DeclarativeNode.GetHitMap] No HitMap available (fiberFirstPaint=%v, renderer=%T)", n.fiberLastHitMap != nil, n.renderer)
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

// requestRender triggers a re-render.
func (n *DeclarativeNode) requestRender(reconciler rtui.Reconciler) {
	if reconciler != nil {
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
		log.RenderLogger.IfEnabled().Debug("[RenderWithFiber] No reconciler or renderer available")
		return fmt.Errorf("no reconciler or renderer for Fiber-based rendering")
	}

	// Get the renderer and call its RenderWithFiber method
	// This allows NodeID propagation from Fiber tree to ComputedBox
	if adapter, ok := n.renderer.(*PipelineRendererAdapter); ok {
		return adapter.RenderWithFiber(n.root, buffer)
	}

	// Fallback for non-PipelineRenderer types
	log.RenderLogger.IfEnabled().Debug("[RenderWithFiber] No PipelineRendererAdapter available")
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

// GetPaintDirtyRects returns dirty rect hints collected from the last paintable tree.
// These hints are optional and used by the terminal renderer for incremental repaint planning.
func (n *DeclarativeNode) GetPaintDirtyRects() []paint.Rect {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if len(n.lastPaintDirtyRects) == 0 {
		return nil
	}
	rects := make([]paint.Rect, len(n.lastPaintDirtyRects))
	copy(rects, n.lastPaintDirtyRects)
	return rects
}

func collectPaintableDirtyRectFromBox(box *paint.PaintableBox, rects *[]paint.Rect) {
	if box == nil || rects == nil {
		return
	}
	if box.LayoutDirty && box.Width > 0 && box.Height > 0 {
		*rects = append(*rects, paint.Rect{
			X:      box.X,
			Y:      box.Y,
			Width:  box.Width,
			Height: box.Height,
		})
	}
}

func (n *DeclarativeNode) collectSceneImageLayers() []paint.ImageLayer {
	n.mu.RLock()
	layers := cloneImageLayers(n.lastSceneImageLayers)
	n.mu.RUnlock()

	return layers
}

func collectSceneImageLayerFromBox(box *paint.PaintableBox, layers *[]paint.ImageLayer) {
	if box == nil || layers == nil {
		return
	}
	fiberNode, ok := box.Node.(*FiberPaintableNode)
	if !ok || fiberNode == nil || fiberNode.fiber == nil {
		return
	}
	sceneInst, ok := rtui.AsScenePaintableInstance(fiberNode.fiber.Instance)
	if !ok {
		return
	}
	if boundsSetter, ok := box.Node.(interface{ SetBounds(int, int, int, int) }); ok {
		boundsSetter.SetBounds(box.X, box.Y, box.Width, box.Height)
	}
	for _, layer := range sceneInst.SceneLayers() {
		if !layer.HasPixels() || layer.Bounds.Width <= 0 || layer.Bounds.Height <= 0 {
			continue
		}
		*layers = append(*layers, layer.Clone())
	}
}

func cloneImageLayers(layers []paint.ImageLayer) []paint.ImageLayer {
	if len(layers) == 0 {
		return nil
	}
	cloned := make([]paint.ImageLayer, 0, len(layers))
	for _, layer := range layers {
		cloned = append(cloned, layer.Clone())
	}
	return cloned
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
		if isPortalFiber(f) {
			found = true
			return
		}
		if f.Props != nil {
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
// Note: Portal-aware layout is enabled by default when using NewDeclarativeNodeFromFuncWithFiber.
// Can be disabled via MINT_PORTAL_LAYOUT=false or MINT_PORTAL_LAYOUT=0 for backward compatibility.
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
