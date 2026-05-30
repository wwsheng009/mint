// Package render provides VNode rendering implementations.
package render

import (
	"github.com/wwsheng009/mint/internal/reconciler"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/render"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Common Utilities
// =============================================================================

// getBuffer safely extracts a paint.Buffer from an interface{}
// Returns nil if the buffer is invalid or wrong type
func getBuffer(buffer interface{}) *paint.Buffer {
	if buf, ok := buffer.(*paint.Buffer); ok && buf != nil {
		return buf
	}
	return nil
}

// =============================================================================
// PipelineRendererAdapter - Adapts PipelineRenderer to VNodeRenderer interface
// =============================================================================

// PipelineRendererAdapter wraps the new PipelineRenderer to implement VNodeRenderer.
// This is now the DEFAULT renderer for all VNode rendering.
type PipelineRendererAdapter struct {
	pipeline *PipelineRenderer
}

// NewPipelineRendererAdapter creates a new adapter using the new rendering pipeline.
func NewPipelineRendererAdapter() *PipelineRendererAdapter {
	return &PipelineRendererAdapter{
		pipeline: NewPipelineRenderer(),
	}
}

// Render renders a VNode using the new Layout/Paint pipeline.
// This implements the VNodeRenderer interface.
func (r *PipelineRendererAdapter) Render(vnode rtui.VNode, x, y int, buffer interface{}) {
	buf := getBuffer(buffer)
	if buf == nil {
		return
	}
	r.pipeline.Render(vnode, x, y, buf)
}

// Measure returns the width and height of a VNode using the new pipeline.
// This implements the VNodeRenderer interface.
func (r *PipelineRendererAdapter) Measure(vnode rtui.VNode) (width, height int) {
	// Use the new pipeline's measure method
	// We use a large max value for unbounded measurement
	return r.pipeline.Measure(vnode, 1000, 1000)
}

// GetCacheStats returns cache statistics from the pipeline.
func (r *PipelineRendererAdapter) GetCacheStats() string {
	return r.pipeline.GetCacheStats()
}

// ClearCache clears the layout cache.
func (r *PipelineRendererAdapter) ClearCache() {
	r.pipeline.ClearCache()
}

// GetPipeline returns the underlying PipelineRenderer for direct access.
// This allows callers to access the rendering pipeline if needed.
func (r *PipelineRendererAdapter) GetPipeline() *PipelineRenderer {
	return r.pipeline
}

// GetRenderingPipeline returns the inner RenderingPipeline for constraint-based rendering.
// This is the recommended way to access the pipeline for layout-aware rendering.
func (r *PipelineRendererAdapter) GetRenderingPipeline() *RenderingPipeline {
	return r.pipeline.GetRenderingPipeline()
}

// GetHooks returns the HookManager for registering VNode transformation hooks.
// This allows framework to register Inspector and other overlay hooks.
func (r *PipelineRendererAdapter) GetHooks() *render.HookManager {
	return r.pipeline.GetHooks()
}

// SetFiber sets the Fiber tree for NodeID propagation
// Phase 8: Fiber to Layout Engine NodeID propagation
// This method allows the reconciler to update the fiber reference in PipelineRenderer
func (r *PipelineRendererAdapter) SetFiber(fiber *reconciler.Fiber) {
	r.pipeline.fiber = fiber
}
