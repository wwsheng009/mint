// Package render provides pipeline-based VNode renderer using the new Layout/Paint separation
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

// PipelineRenderer is a VNodeRenderer that uses the new RenderingPipeline
// with separated Layout and Paint phases.
//
// This renderer provides:
// - Constraint-driven layout calculation
// - Independent paint phase using pre-computed positions
// - Caching for leaf nodes
// - Multi-layer rendering support (Modal, Overlay, Tooltip)
// - Better separation of concerns
// - Hook system for VNode transformation (e.g., Inspector injection)
type PipelineRenderer struct {
	pipeline *RenderingPipeline
	hooks    *render.HookManager
	fiber    *reconciler.Fiber // Fiber tree for NodeID propagation
	debug    bool
}

// NewPipelineRenderer creates a new pipeline-based VNodeRenderer
func NewPipelineRenderer() *PipelineRenderer {
	return &PipelineRenderer{
		pipeline: NewRenderingPipeline(),
		hooks:    render.NewHookManager(),
		debug:    log.PipelineLogger.Enabled(),
	}
}

// Render implements the VNodeRenderer interface
func (r *PipelineRenderer) Render(vnode rtui.VNode, x, y int, buffer interface{}) error {
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

	// Apply VNode hooks (e.g., Inspector injection)
	// Hooks can modify the VNode tree before rendering
	vnode = r.hooks.ApplyVNodeHooks(vnode)

	// For the new pipeline, we use the buffer size as constraints
	// Note: This is the legacy behavior - use RenderWithConstraints for proper layout sizing
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
		// Use standard rendering for simple VNode trees
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
// This allows external code (like framework) to register hooks for features
// like Inspector injection.
func (r *PipelineRenderer) GetHooks() *render.HookManager {
	return r.hooks
}

// Measure implements the VNodeRenderer Measure interface
// This allows the renderer to be used for size calculation
func (r *PipelineRenderer) Measure(vnode rtui.VNode, maxWidth, maxHeight int) (width, height int) {
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

// RenderWithConstraints renders with explicit layout constraints (not buffer size)
// This is important when the buffer size differs from the desired layout constraints
// For example: terminal is 156x44 but user configured 80x24 for layout
func (r *PipelineRenderer) RenderWithConstraints(vnode rtui.VNode, layoutWidth, layoutHeight int, buffer *paint.Buffer) error {
	if buffer == nil {
		return nil
	}

	// Apply VNode hooks (e.g., Inspector injection)
	vnode = r.hooks.ApplyVNodeHooks(vnode)

	// Use explicit layout constraints instead of buffer size
	constraints := runtime.NewBoxConstraints(0, layoutWidth, 0, layoutHeight)

	log.RenderLogger.Debug("Layout constraints: %dx%d (buffer: %dx%d)",
		layoutWidth, layoutHeight, buffer.Width, buffer.Height)
	log.LayerLogger.Debug("Layout constraints: %dx%d (buffer: %dx%d)",
		layoutWidth, layoutHeight, buffer.Width, buffer.Height)

	// Check if Fiber tree contains any layer nodes (Modal, Overlay, Tooltip)
	hasLayers := r.hasLayerNodes()

	log.LayerLogger.IfEnabled().Debug("hasLayers=%v", hasLayers)

	var err error
	if hasLayers {
		err = r.pipeline.RenderLayers(r.fiber, constraints, buffer)
	} else {
		err = r.pipeline.Render(r.fiber, constraints, buffer)
	}

	if err != nil {
		log.RenderLogger.IfEnabled().Debug("❌ Render FAILED: %v", err)
		return err
	}

	log.RenderLogger.IfEnabled().Debug("✅ Render SUCCESS")

	if r.debug {
		log.RenderLogger.IfEnabled().Debug("Render complete, cache stats: %s", r.GetCacheStats())
	}

	return nil
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

// =============================================================================
// Helper Functions for Using Pipeline Renderer
// =============================================================================

// UsePipelineRendererOption creates an option that configures a DeclarativeNode
// to use the new pipeline-based renderer
func UsePipelineRendererOption() func(*DeclarativeNode) {
	return func(node *DeclarativeNode) {
		// Note: This would require adding a SetRenderer method to DeclarativeNode
		// For now, this is a placeholder for future enhancement
		_ = node
	}
}

// RenderWithFiber renders with explicit Fiber tree for NodeID propagation
// Phase 8: This method allows passing Fiber tree directly to layout engine
// so that NodeIDs can be propagated to ComputedBox for proper identity tracking
//
// Parameters:
//   vnode: The rendered VNode tree
//   fiber: The actual Fiber tree with NodeIDs (nil for non-Fiber mode)
//   buffer: Paint buffer for rendering
func (r *PipelineRenderer) RenderWithFiber(vnode rtui.VNode, fiber *reconciler.Fiber, buffer *paint.Buffer) error {
	if buffer == nil {
		return nil
	}

	// Apply VNode hooks (e.g., Inspector injection)
	vnode = r.hooks.ApplyVNodeHooks(vnode)

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
		// Use standard rendering for simple VNode trees
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
func (a *PipelineRendererAdapter) RenderWithFiber(vnode rtui.VNode, buffer *paint.Buffer) error {
	return a.pipeline.RenderWithFiber(vnode, a.pipeline.fiber, buffer)
}

// =============================================================================
// Error Types
// =============================================================================

// ErrInvalidBuffer is returned when an invalid buffer is provided
type ErrInvalidBuffer struct {
	Msg string
}

func (e *ErrInvalidBuffer) Error() string {
	return e.Msg
}
