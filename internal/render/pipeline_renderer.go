// Package render provides Fiber-backed declarative rendering with separated layout and paint phases.
package render

import (
	"fmt"

	"github.com/wwsheng009/mint/internal/log"
	"github.com/wwsheng009/mint/internal/reconciler"
	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/render"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// PipelineRenderer owns the Fiber-backed RenderingPipeline.
//
// It still satisfies the VNodeRenderer compatibility API, but the active
// application render path is Fiber-first: DeclarativeNode reconciles VNodes
// into Fiber, then this renderer lays out and paints the Fiber tree.
//
// This renderer provides:
// - Constraint-driven layout calculation
// - Independent paint phase using pre-computed positions
// - Caching for leaf nodes
// - Multi-layer rendering support (Modal, Overlay, Tooltip)
// - Better separation of concerns
// - Hook manager for VNode transforms before Fiber reconciliation
type PipelineRenderer struct {
	pipeline *RenderingPipeline
	hooks    *render.HookManager
	fiber    *reconciler.Fiber // Fiber tree for NodeID propagation
	debug    bool
}

// NewPipelineRenderer creates a new Fiber-backed pipeline renderer.
func NewPipelineRenderer() *PipelineRenderer {
	return &PipelineRenderer{
		pipeline: NewRenderingPipeline(),
		hooks:    render.NewHookManager(),
		debug:    log.PipelineLogger.Enabled(),
	}
}

// Render implements the VNodeRenderer compatibility interface.
//
// The VNode and offset arguments are kept for API compatibility. Rendering uses
// the current Fiber tree supplied by the reconciler via SetFiber.
func (r *PipelineRenderer) Render(_ rtui.VNode, _, _ int, buffer interface{}) error {
	buf, ok := buffer.(*paint.Buffer)
	if !ok {
		// Try to convert from other buffer types
		if b, ok := buffer.(interface{ GetBuffer() *paint.Buffer }); ok {
			buf = b.GetBuffer()
		} else {
			return fmt.Errorf("invalid buffer type for PipelineRenderer")
		}
	}

	if buf == nil {
		return nil
	}

	// Compatibility rendering uses the buffer size as constraints.
	width := buf.Width
	height := buf.Height

	// Debug: Log buffer size
	log.RenderLogger.IfEnabled().Debug("Buffer size: %dx%d", buf.Width, buf.Height)
	log.LayerLogger.IfEnabled().Debug("Buffer size: %dx%d", buf.Width, buf.Height)

	constraints := runtime.NewBoxConstraints(0, width, 0, height)

	// Check if Fiber tree contains any layer nodes (Modal, Overlay, Tooltip)
	hasLayers := r.hasLayerNodes()

	log.LayerLogger.IfEnabled().Debug("hasLayers=%v", hasLayers)

	var err error
	if hasLayers {
		// Use multi-layer rendering for modals, overlays, tooltips
		log.RenderLogger.IfEnabled().Debug("Using RenderLayers for multi-layer rendering")
		err = r.pipeline.RenderLayers(r.fiber, constraints, buf)
	} else {
		// Use standard rendering for simple Fiber trees
		log.RenderLogger.IfEnabled().Debug("Using standard Render")
		err = r.pipeline.Render(r.fiber, constraints, buf)
	}

	if err != nil {
		log.RenderLogger.IfEnabled().Debug("X[Render FAILED: %v", err)
		return err
	}

	log.RenderLogger.IfEnabled().Debug("✅ Render SUCCESS")

	if r.debug {
		log.RenderLogger.IfEnabled().Debug("Render complete, cache stats: %s", r.GetCacheStats())
	}

	return nil
}

// hasLayerNodes checks the Fiber tree for non-base layer nodes.
func (r *PipelineRenderer) hasLayerNodes() bool {
	return r.hasLayerNodesFromFiber(r.fiber)
}

// hasLayerNodesFromFiber checks if the Fiber tree contains any non-base layer nodes
// This is the preferred method in Fiber mode
func (r *PipelineRenderer) hasLayerNodesFromFiber(fiber *rtui.Fiber) bool {
	if fiber == nil {
		return false
	}

	// Check this node
	layer := fiber.Layer
	log.HitMapLogger.Debug("[hasLayerNodes] Node type=%T, Layer=%d, IsValid=%v",
		fiber, layer, layer.IsValid())
	if layer != rtui.LayerBase && layer.IsValid() {
		log.HitMapLogger.IfEnabled().Debug("[hasLayerNodes] ✅ Found layer node: Layer=%d", layer)
		return true
	}

	// Recursively check children (Fiber tree: Child -> Sibling)
	for child := fiber.Child; child != nil; child = child.Sibling {
		if r.hasLayerNodesFromFiber(child) {
			return true
		}
	}

	return false
}

// GetRenderingPipeline returns the inner RenderingPipeline for direct access.
// This allows calling the constraint-based Render method directly.
func (r *PipelineRenderer) GetRenderingPipeline() *RenderingPipeline {
	return r.pipeline
}

// GetHooks returns the HookManager for registering VNode transformation hooks.
// DeclarativeNode applies these hooks before Fiber reconciliation.
func (r *PipelineRenderer) GetHooks() *render.HookManager {
	return r.hooks
}

// Measure implements the VNodeRenderer compatibility measurement API.
// It measures the current Fiber root, not the VNode argument.
func (r *PipelineRenderer) Measure(_ rtui.VNode, maxWidth, maxHeight int) (width, height int) {
	if r.fiber == nil {
		log.RenderLogger.IfEnabled().Debug("[PipelineRenderer.Measure] no Fiber root available")
		return 0, 0
	}

	constraints := runtime.NewBoxConstraints(0, maxWidth, 0, maxHeight)
	result, err := NewNewLayoutEngineAdapter().LayoutFiber(r.fiber, constraints)
	if err != nil {
		log.RenderLogger.IfEnabled().Debug("[PipelineRenderer.Measure] layout failed: %v", err)
		return 0, 0
	}
	adapter, ok := result.(*newLayoutResultAdapter)
	if !ok || adapter.result == nil || adapter.result.Root == nil {
		return 0, 0
	}

	root := adapter.result.Root
	return root.Width, root.Height
}

// GetPipeline returns the underlying RenderingPipeline for advanced usage
func (r *PipelineRenderer) GetPipeline() *RenderingPipeline {
	return r.pipeline
}

// SetDebug enables/disables debug output
func (r *PipelineRenderer) SetDebug(debug bool) {
	r.debug = debug
	r.pipeline.SetLayoutDebug(debug)
	r.pipeline.SetPaintDebug(debug)
}

// GetCacheStats returns cache statistics from the layout engine
func (r *PipelineRenderer) GetCacheStats() string {
	stats := r.pipeline.GetCacheStats()
	return fmt.Sprintf("hits=%d, misses=%d", stats.Hits, stats.Misses)
}

// ClearCache clears the layout cache
func (r *PipelineRenderer) ClearCache() {
	r.pipeline.ClearCache()
}

// RenderWithFiber renders with explicit Fiber tree for NodeID propagation
// Phase 8: This method allows passing Fiber tree directly to layout engine
// so that NodeIDs can be propagated to ComputedBox for proper identity tracking
//
// Parameters:
//
//	fiber: The actual Fiber tree with NodeIDs.
//	buffer: Paint buffer for rendering.
func (r *PipelineRenderer) RenderWithFiber(fiber *reconciler.Fiber, buffer *paint.Buffer) error {
	if buffer == nil {
		return nil
	}

	// Get buffer dimensions for layout constraints
	width := buffer.Width
	height := buffer.Height

	log.RenderLogger.IfEnabled().Debug("RenderWithFiber: Buffer size: %dx%d", width, height)
	log.LayerLogger.IfEnabled().Debug("RenderWithFiber: Buffer size: %dx%d", width, height)

	constraints := runtime.NewBoxConstraints(0, width, 0, height)

	// Check if Fiber tree contains any layer nodes (Modal, Overlay, Tooltip)
	hasLayers := r.hasLayerNodesFromFiber(fiber)

	log.LayerLogger.IfEnabled().Debug("RenderWithFiber: hasLayers=%v, fiber=%v", hasLayers, fiber != nil)

	var err error
	if hasLayers {
		// Use multi-layer rendering for modals, overlays, tooltips
		log.RenderLogger.IfEnabled().Debug("RenderWithFiber: Using RenderLayers for multi-layer rendering")
		err = r.pipeline.RenderLayers(fiber, constraints, buffer)
	} else {
		// Use standard rendering for simple Fiber trees
		log.RenderLogger.IfEnabled().Debug("RenderWithFiber: Using standard Render")
		err = r.pipeline.Render(fiber, constraints, buffer)
	}

	if err != nil {
		log.RenderLogger.IfEnabled().Debug("X[RenderWithFiber] FAILED: %v", err)
	}

	return err
}

// RenderWithFiber renders with explicit Fiber tree for NodeID propagation
// Phase 8: Adapter method that delegates to pipeline.RenderWithFiber
func (a *PipelineRendererAdapter) RenderWithFiber(buffer *paint.Buffer) error {
	return a.pipeline.RenderWithFiber(a.pipeline.fiber, buffer)
}
