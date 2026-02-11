// Package render provides pipeline-based VNode renderer using the new Layout/Paint separation
package render

import (
	"fmt"
	"os"

	"github.com/wwsheng009/mint/internal/log"
	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/layer"
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
	pipeline    *RenderingPipeline
	layerMgr    *layer.Manager
	layerEvents *layer.EventHandler
	hooks       *render.HookManager
	debug       bool
}

// NewPipelineRenderer creates a new pipeline-based VNodeRenderer
func NewPipelineRenderer() *PipelineRenderer {
	layerMgr := layer.NewManager()
	return &PipelineRenderer{
		pipeline:    NewRenderingPipeline(),
		layerMgr:    layerMgr,
		layerEvents: layer.NewEventHandler(layerMgr),
		hooks:       render.NewHookManager(),
		debug:       os.Getenv("TUI_PIPELINE_DEBUG") == "true",
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
	if os.Getenv("TUI_DEBUG_RENDERING") == "true" || os.Getenv("TUI_LAYER_DEBUG") == "true" {
		fmt.Fprintf(os.Stderr, "[PipelineRenderer] Buffer size: %dx%d\n", buf.Width, buf.Height)
	}

	constraints := runtime.NewBoxConstraints(0, width, 0, height)

	// Check if VNode tree contains any layer nodes (Modal, Overlay, Tooltip)
	hasLayers := r.hasLayerNodes(vnode)

	if r.debug || os.Getenv("TUI_LAYER_DEBUG") == "true" {
		fmt.Fprintf(os.Stderr, "[PipelineRenderer] hasLayers=%v\n", hasLayers)
	}

	var err error
	if hasLayers {
		// Use multi-layer rendering for modals, overlays, tooltips
		if r.debug || os.Getenv("TUI_DEBUG_RENDERING") == "true" {
			fmt.Fprintf(os.Stderr, "[PipelineRenderer] Using RenderLayers for multi-layer rendering\n")
		}
		err = r.pipeline.RenderLayers(vnode, constraints, buf)
	} else {
		// Use standard rendering for simple VNode trees
		if r.debug || os.Getenv("TUI_DEBUG_RENDERING") == "true" {
			fmt.Fprintf(os.Stderr, "[PipelineRenderer] Using standard Render\n")
		}
		err = r.pipeline.Render(vnode, constraints, buf)
	}

	if err != nil {
		if r.debug || os.Getenv("TUI_DEBUG_RENDERING") == "true" {
			fmt.Fprintf(os.Stderr, "[PipelineRenderer] ❌ Render FAILED: %v, falling back to legacy\n", err)
		}
		// Fall back to legacy rendering if pipeline fails
		return r.renderLegacy(vnode, x, y, buf)
	}

	if r.debug || os.Getenv("TUI_DEBUG_RENDERING") == "true" {
		fmt.Fprintf(os.Stderr, "[PipelineRenderer] ✅ Render SUCCESS\n")
	}

	if r.debug {
		fmt.Fprintf(os.Stderr, "[PipelineRenderer] Render complete, cache stats: %s\n",
			r.pipeline.GetLayoutEngine().GetCacheStats().String())
	}

	return nil
}

// hasLayerNodes checks if the VNode tree contains any non-base layer nodes
func (r *PipelineRenderer) hasLayerNodes(vnode rtui.VNode) bool {
	if vnode == nil {
		return false
	}

	// Check this node
	layer := vnode.GetLayer()
	if os.Getenv("TUI_DEBUG_HITMAP") == "true" {
		log.RenderLogger.Debug("[hasLayerNodes] Node type=%T, Layer=%d, IsValid=%v",
			vnode, layer, layer.IsValid())
	}
	if layer != rtui.LayerBase && layer.IsValid() {
		if os.Getenv("TUI_DEBUG_HITMAP") == "true" {
			log.RenderLogger.Debug("[hasLayerNodes] ✅ Found layer node: Layer=%d", layer)
		}
		return true
	}

	// Recursively check children
	for _, child := range vnode.Children() {
		if r.hasLayerNodes(child) {
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
	// Use the compute engine to measure the VNode
	constraints := runtime.NewBoxConstraints(0, maxWidth, 0, maxHeight)
	layout, err := r.pipeline.GetLayoutEngine().Layout(vnode, constraints)
	if err != nil {
		if r.debug {
			fmt.Fprintf(os.Stderr, "[PipelineRenderer] Layout failed: %v\n", err)
		}
		return 0, 0
	}

	if layout.Root == nil {
		return 0, 0
	}

	return layout.Root.Box.Width, layout.Root.Box.Height
}

// renderLegacy provides fallback rendering using the legacy PaintVNode approach
func (r *PipelineRenderer) renderLegacy(vnode rtui.VNode, x, y int, buffer *paint.Buffer) error {
	// Create a temporary DeclarativeNode to use legacy rendering
	tempNode := NewDeclarativeNode(vnode)
	tempNode.PaintVNode(vnode, x, y, buffer)
	return nil
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

	if os.Getenv("TUI_DEBUG_RENDERING") == "true" || os.Getenv("TUI_LAYER_DEBUG") == "true" {
		fmt.Fprintf(os.Stderr, "[PipelineRenderer] Layout constraints: %dx%d (buffer: %dx%d)\n",
			layoutWidth, layoutHeight, buffer.Width, buffer.Height)
	}

	// Check if VNode tree contains any layer nodes (Modal, Overlay, Tooltip)
	hasLayers := r.hasLayerNodes(vnode)

	if r.debug || os.Getenv("TUI_LAYER_DEBUG") == "true" {
		fmt.Fprintf(os.Stderr, "[PipelineRenderer] hasLayers=%v\n", hasLayers)
	}

	var err error
	if hasLayers {
		err = r.pipeline.RenderLayers(vnode, constraints, buffer)
	} else {
		err = r.pipeline.Render(vnode, constraints, buffer)
	}

	if err != nil {
		if r.debug || os.Getenv("TUI_DEBUG_RENDERING") == "true" {
			fmt.Fprintf(os.Stderr, "[PipelineRenderer] ❌ Render FAILED: %v\n", err)
		}
		return err
	}

	if r.debug || os.Getenv("TUI_DEBUG_RENDERING") == "true" {
		fmt.Fprintf(os.Stderr, "[PipelineRenderer] ✅ Render SUCCESS\n")
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
	stats := r.pipeline.GetLayoutEngine().GetCacheStats()
	return stats.String()
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
