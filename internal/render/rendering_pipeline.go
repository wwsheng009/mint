// Package render provides new rendering pipeline with separated Layout and Paint phases
package render

import (
	"fmt"
	"os"

	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/compute"
	"github.com/wwsheng009/mint/runtime/layer"
	"github.com/wwsheng009/mint/runtime/paint"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// RenderingPipeline is the new rendering pipeline with separated Layout and Paint phases
// Layout phase: compute.Engine calculates all positions
// Paint phase: PaintEngine renders using computed positions
type RenderingPipeline struct {
	layoutEngine *compute.Engine
	paintEngine  *PaintEngine
}

// NewRenderingPipeline creates a new rendering pipeline
func NewRenderingPipeline() *RenderingPipeline {
	return &RenderingPipeline{
		layoutEngine: compute.NewEngine(),
		paintEngine:  NewPaintEngine(),
	}
}

// SetLayoutDebug enables/disables layout debug output
func (p *RenderingPipeline) SetLayoutDebug(debug bool) {
	p.layoutEngine.SetDebug(debug)
}

// SetPaintDebug enables/disables paint debug output
func (p *RenderingPipeline) SetPaintDebug(debug bool) {
	p.paintEngine.SetDebug(debug)
}

// Render performs the complete rendering pipeline:
// 1. Layout phase: calculate positions for all nodes
// 2. Paint phase: render using computed positions
func (p *RenderingPipeline) Render(vnode rtui.VNode, constraints runtime.BoxConstraints, buffer *paint.Buffer) error {
	if vnode == nil {
		return nil
	}

	if os.Getenv("TUI_PIPELINE_DEBUG") == "true" || os.Getenv("TUI_DEBUG_RENDERING") == "true" {
		fmt.Fprintf(os.Stderr, "[RenderingPipeline] Render started\n")
	}

	// Phase 1: Layout - calculate all positions
	layout, err := p.layoutEngine.Layout(vnode, constraints)
	if err != nil {
		// Fallback to legacy rendering if layout fails
		fmt.Fprintf(os.Stderr, "[RenderingPipeline] ❌ Layout FAILED: %v, falling back to legacy\n", err)
		return p.renderLegacy(vnode, 0, 0, buffer)
	}

	if os.Getenv("TUI_PIPELINE_DEBUG") == "true" || os.Getenv("TUI_DEBUG_RENDERING") == "true" {
		fmt.Fprintf(os.Stderr, "[RenderingPipeline] ✅ Layout complete, starting Paint phase\n")
	}

	// Phase 2: Paint - render using computed positions
	if os.Getenv("TUI_DEBUG_RENDERING") == "true" {
		fmt.Fprintf(os.Stderr, "[RenderingPipeline] Starting Paint phase...\n")
	}
	err = p.paintEngine.Paint(layout, buffer)

	if os.Getenv("TUI_PIPELINE_DEBUG") == "true" || os.Getenv("TUI_PAINT_DEBUG") == "true" || os.Getenv("TUI_DEBUG_RENDERING") == "true" {
		fmt.Fprintf(os.Stderr, "[RenderingPipeline] Paint complete, err=%v\n", err)
	}

	return err
}

// RenderToSize renders with specific window size constraints
func (p *RenderingPipeline) RenderToSize(vnode rtui.VNode, width, height int, buffer *paint.Buffer) error {
	constraints := runtime.NewBoxConstraints(0, width, 0, height)
	return p.Render(vnode, constraints, buffer)
}

// renderLegacy fallback rendering for when the new pipeline fails
// This preserves backward compatibility
func (p *RenderingPipeline) renderLegacy(vnode rtui.VNode, x, y int, buffer *paint.Buffer) error {
	// Create a temporary DeclarativeNode to use legacy rendering
	// This bridges the new pipeline with the old PaintVNode approach
	tempNode := NewDeclarativeNode(vnode)
	tempNode.PaintVNode(vnode, x, y, buffer)
	return nil
}

// ComputeLayout performs only the layout phase, returning computed positions
// This can be useful for hit testing and other operations that need layout info without rendering
func (p *RenderingPipeline) ComputeLayout(vnode rtui.VNode, constraints runtime.BoxConstraints) (*compute.ComputedLayout, error) {
	return p.layoutEngine.Layout(vnode, constraints)
}

// GetLayoutEngine returns the layout engine for direct access
func (p *RenderingPipeline) GetLayoutEngine() *compute.Engine {
	return p.layoutEngine
}

// GetPaintEngine returns the paint engine for direct access
func (p *RenderingPipeline) GetPaintEngine() *PaintEngine {
	return p.paintEngine
}

// =============================================================================
// Cache Management
// =============================================================================

// GetCacheStats returns statistics about the layout cache
func (p *RenderingPipeline) GetCacheStats() compute.CacheStats {
	return p.layoutEngine.GetCacheStats()
}

// ResetCacheStats resets cache hit/miss counters
func (p *RenderingPipeline) ResetCacheStats() {
	p.layoutEngine.ResetCacheStats()
}

// ClearCache clears all cached layout results
func (p *RenderingPipeline) ClearCache() {
	p.layoutEngine.ClearCache()
}

// InvalidateCacheByType removes cached entries for a specific VNode type
func (p *RenderingPipeline) InvalidateCacheByType(vNodeType string) {
	p.layoutEngine.InvalidateCacheByType(vNodeType)
}

// InvalidateCacheByKey removes cached entries for a specific VNode key
func (p *RenderingPipeline) InvalidateCacheByKey(vnodeKey string) {
	p.layoutEngine.InvalidateCacheByKey(vnodeKey)
}

// =============================================================================
// Multi-Layer Rendering
// =============================================================================

// RenderLayers renders a VNode tree with multi-layer support
// This is the main entry point for layer-based rendering
func (p *RenderingPipeline) RenderLayers(
	vnode rtui.VNode,
	constraints runtime.BoxConstraints,
	buffer *paint.Buffer,
) error {
	if vnode == nil {
		return nil
	}

	if os.Getenv("TUI_PIPELINE_DEBUG") == "true" {
		fmt.Fprintf(os.Stderr, "[RenderingPipeline] RenderLayers started\n")
	}

	// Create a layer manager for this render pass
	layerMgr := layer.NewManager()

	// Collect and layout all layers
	if err := layerMgr.CollectAndLayout(vnode, constraints, p.layoutEngine); err != nil {
		if os.Getenv("TUI_PIPELINE_DEBUG") == "true" {
			fmt.Fprintf(os.Stderr, "[RenderingPipeline] Layer layout failed: %v\n", err)
		}
		// Fallback to regular rendering
		return p.Render(vnode, constraints, buffer)
	}

	// Get all layer layouts
	layouts := layerMgr.GetLayouts()

	if os.Getenv("TUI_PIPELINE_DEBUG") == "true" {
		fmt.Fprintf(os.Stderr, "[RenderingPipeline] Layer layouts complete, rendering %d layers\n", len(layouts))
	}

	// Paint all layers
	if err := p.paintEngine.PaintLayers(layouts, buffer); err != nil {
		if os.Getenv("TUI_PIPELINE_DEBUG") == "true" {
			fmt.Fprintf(os.Stderr, "[RenderingPipeline] PaintLayers failed: %v\n", err)
		}
		return err
	}

	// 验证 buffer 内容
	if os.Getenv("TUI_PIPELINE_DEBUG") == "true" || os.Getenv("TUI_DEBUG_RENDERING") == "true" {
		contentCount := 0
		for y := 0; y < buffer.Height; y++ {
			for x := 0; x < buffer.Width; x++ {
				if buffer.Cells[y][x].Cluster != "" {
					contentCount++
				}
			}
		}
		fmt.Fprintf(os.Stderr, "[RenderLayers] Buffer content after PaintLayers: %d cells (buffer size: %dx%d)\n",
			contentCount, buffer.Width, buffer.Height)

		if contentCount == 0 {
			fmt.Fprintf(os.Stderr, "[RenderLayers] ⚠️  WARNING: Buffer is empty!\n")
		} else {
			fmt.Fprintf(os.Stderr, "[RenderLayers] ✅ Buffer has content\n")
		}
	}

	return nil
}

// HasModalChecks returns whether the rendering pipeline detected any modal content
// This can be used to determine if events should be blocked
func (p *RenderingPipeline) HasModalChecks(vnode rtui.VNode, constraints runtime.BoxConstraints) bool {
	layerMgr := layer.NewManager()
	layerMgr.CollectAndLayout(vnode, constraints, p.layoutEngine)
	return layerMgr.HasModal()
}
