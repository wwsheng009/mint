// Package render provides VNode rendering implementations.
package render

import (
	"github.com/wwsheng009/mint/internal/reconciler"
	"github.com/wwsheng009/mint/runtime"
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

// measureExplicitDimensions checks for explicit width/height in props
// Returns (width, height) with 0 meaning not explicitly set
func measureExplicitDimensions(vnode rtui.VNode) (width, height int) {
	if vnode == nil {
		return 0, 0
	}
	props := vnode.Props()
	if props == nil {
		return 0, 0
	}
	return props.GetInt("width"), props.GetInt("height")
}

// =============================================================================
// NonFiberRenderer
// =============================================================================

// NonFiberRenderer implements VNodeRenderer for traditional (non-Fiber) rendering.
// It walks the VNode tree directly and renders each node to the buffer.
//
// Deprecated: Use FiberRenderer with Fiber-first architecture instead.
// NonFiber mode does not support persistent component instances or hooks.
type NonFiberRenderer struct {
	// owner is the DeclarativeNode that owns this renderer
	// This allows access to methods like measureVNodeWidth
	owner *DeclarativeNode
}

// NewNonFiberRenderer creates a new NonFiberRenderer.
//
// Deprecated: Use NewFiberRenderer with Fiber-first architecture instead.
func NewNonFiberRenderer(owner *DeclarativeNode) *NonFiberRenderer {
	return &NonFiberRenderer{owner: owner}
}

// Render renders a VNode tree to a buffer at the specified position.
// This implements the VNodeRenderer interface.
func (r *NonFiberRenderer) Render(vnode rtui.VNode, x, y int, buffer interface{}) {
	buf := getBuffer(buffer)
	if buf == nil {
		return
	}
	r.owner.PaintVNode(vnode, x, y, buf)
}

// Measure returns the width and height of a VNode.
// This implements the VNodeRenderer interface.
func (r *NonFiberRenderer) Measure(vnode rtui.VNode) (width, height int) {
	if vnode == nil {
		return 0, 0
	}

	// Check if VNode implements Measurable interface (Phase 4 improvement)
	type measurable interface {
		Measure(constraints runtime.BoxConstraints) runtime.Size
	}
	if m, ok := vnode.(measurable); ok {
		// Use UnboundedConstraints to get natural size
		// Components will report their content-based size, not expand to fill arbitrary max widths
		size := m.Measure(runtime.UnboundedConstraints())
		return size.Width, size.Height
	}

	// Fallback: Get the width using the existing measurement logic
	width = r.owner.MeasureVNodeWidth(vnode)

	// Height is typically 1 for leaf nodes, calculated for containers
	height = r.measureVNodeHeight(vnode)

	return width, height
}

// measureVNodeHeight measures the height of a VNode.
func (r *NonFiberRenderer) measureVNodeHeight(vnode rtui.VNode) int {
	if vnode == nil {
		return 0
	}

	// Check for explicit height in props
	if props := vnode.Props(); props != nil {
		if h := props.GetInt("height"); h > 0 {
			return h
		}
	}

	// For containers, count children
	switch vnode.Type() {
	case rtui.VNodeText, rtui.VNodeElement:
		// Leaf nodes are typically 1 line
		return 1

	case rtui.VNodeFragment:
		// Fragments span their children
		children := vnode.Children()
		height := 0
		for _, child := range children {
			height += r.measureVNodeHeight(child)
		}
		return height

	default:
		return 1
	}
}

// =============================================================================
// FiberRenderer - Implements VNodeRenderer for Fiber mode
// =============================================================================

// FiberRenderer implements VNodeRenderer for Fiber-based rendering.
// It renders through the Fiber reconciler's render callback.
type FiberRenderer struct {
	// renderCallback is the function that actually renders VNodes to the buffer
	renderCallback func(rtui.VNode, int, int, *paint.Buffer)
}

// NewFiberRenderer creates a new FiberRenderer with the given render callback.
func NewFiberRenderer(renderCallback func(rtui.VNode, int, int, *paint.Buffer)) *FiberRenderer {
	return &FiberRenderer{
		renderCallback: renderCallback,
	}
}

// Render renders a VNode to a buffer at the specified position.
// This implements the VNodeRenderer interface.
func (r *FiberRenderer) Render(vnode rtui.VNode, x, y int, buffer interface{}) {
	buf := getBuffer(buffer)
	if buf == nil {
		return
	}
	if r.renderCallback != nil {
		r.renderCallback(vnode, x, y, buf)
	}
}

// Measure returns the width and height of a VNode.
// This implements the VNodeRenderer interface.
func (r *FiberRenderer) Measure(vnode rtui.VNode) (width, height int) {
	if vnode == nil {
		return 0, 0
	}

	// Check for explicit dimensions first
	w, h := measureExplicitDimensions(vnode)
	if w > 0 {
		width = w
	}
	if h > 0 {
		height = h
	}

	// Check if VNode implements Measurable interface (Phase 4 improvement)
	type measurable interface {
		Measure(constraints runtime.BoxConstraints) runtime.Size
	}
	if m, ok := vnode.(measurable); ok {
		// Use UnboundedConstraints to get natural size
		// Components will report their content-based size, not expand to fill arbitrary max widths
		size := m.Measure(runtime.UnboundedConstraints())
		return size.Width, size.Height
	}

	// Fallback: For text content, measure the text length
	if width == 0 {
		if text := rtui.GetTextContent(vnode); text != "" {
			width = len(text)
			height = 1
		}
	}

	// For buttons, measure label + brackets
	if width == 0 {
		if btn, ok := vnode.(interface{ Label() string }); ok {
			width = len(btn.Label()) + 2
			height = 1
		}
	}

	// Default dimensions
	if width == 0 {
		width = 1
	}
	if height == 0 {
		height = 1
	}

	return width, height
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

