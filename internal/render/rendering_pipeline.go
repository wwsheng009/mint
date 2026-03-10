// Package render provides new rendering pipeline with separated Layout and Paint phases
package render

import (
	"github.com/wwsheng009/mint/internal/log"
	"github.com/wwsheng009/mint/internal/reconciler"
	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/event"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// RenderingPipeline is the rendering pipeline with separated Layout and Paint phases.
//
// Architecture:
//   - Layout phase: layout.Engine calculates positions
//   - Paint phase: PaintEngine renders using computed positions
//
// For Fiber-first rendering, use fiberFirstPaint() in DeclarativeNode which
// uses NewLayoutEngineAdapter directly for better performance.
type RenderingPipeline struct {
	layoutEngine *layout.Engine // Layout engine
	paintEngine  *PaintEngine
	lastHitMap   *event.HitMap // HitMap from the most recent RenderLayers call
}

// NewRenderingPipeline creates a new rendering pipeline.
func NewRenderingPipeline() *RenderingPipeline {
	pipeline := &RenderingPipeline{
		layoutEngine: layout.NewEngine(),
		paintEngine:  NewPaintEngine(),
	}

	log.PipelineLogger.IfEnabled().Debug("[RenderingPipeline] Initialized with layout.Engine")
	return pipeline
}

// SetLayoutDebug enables/disables layout debug output
func (p *RenderingPipeline) SetLayoutDebug(debug bool) {
	// Layout engine debug is handled internally
}

// SetPaintDebug enables/disables paint debug output
func (p *RenderingPipeline) SetPaintDebug(debug bool) {
	p.paintEngine.SetDebug(debug)
}

// Render performs the complete rendering pipeline:
// 1. Layout phase: calculate positions for all nodes using layout.Engine
// 2. Paint phase: render using computed positions
func (p *RenderingPipeline) Render(vnode rtui.VNode, fiber *reconciler.Fiber, constraints runtime.BoxConstraints, buffer *paint.Buffer) error {
	if vnode == nil {
		log.PaintLogger.IfEnabled().Debug("[RenderingPipeline.Render] vnode is nil, returning")
		return nil
	}
	log.PaintLogger.IfEnabled().Debug("[RenderingPipeline.Render] START: vnode type=%d, tag=%s, buffer=%dx%d", vnode.Type(), vnode.Tag(), buffer.Width, buffer.Height)
	// Convert constraints to layout.Constraints
	layoutConstraints := layout.Constraints{
		MinWidth:  constraints.MinWidth,
		MaxWidth:  constraints.MaxWidth,
		MinHeight: constraints.MinHeight,
		MaxHeight: constraints.MaxHeight,
	}

	// Choose adapter based on whether we have Fiber or VNode
	var node layout.Node
	var converter PaintableConverter

	if fiber != nil {
		// Fiber-first path: use FiberToNodeAdapterPure
		node = NewFiberToNodeAdapterPure(fiber)
		converter = NewFiberToPaintableConverter(fiber)
	} else {
		// Legacy VNode path: use VNodeToNodeAdapter
		node = NewVNodeToNodeAdapter(vnode)
		converter = NewVNodeToPaintableConverter(vnode)
	}

	// Phase 1: Layout using layout.Engine
	result := p.layoutEngine.Layout(node, layoutConstraints)
	if result == nil || result.Root == nil {
		log.PipelineLogger.IfEnabled().Debug("❌ Layout returned nil result, falling back to legacy")
		return p.renderLegacy(vnode, 0, 0, buffer)
	}

	log.PaintLogger.IfEnabled().Debug("[RenderingPipeline.Render] Layout: Root=%dx%d", result.Root.Width, result.Root.Height)
	log.PipelineLogger.IfEnabled().Debug("✅ Layout complete, starting Paint phase")

	// Convert LayoutBox to PaintableBox
	paintableLayout := converter.ConvertToLayout(result.Root)
	if paintableLayout == nil || paintableLayout.Root == nil {
		log.PipelineLogger.IfEnabled().Debug("❌ PaintableLayout conversion failed, falling back to legacy")
		return p.renderLegacy(vnode, 0, 0, buffer)
	}

	// Phase 2: Paint using PaintableLayout
	err := p.paintEngine.PaintLayout(paintableLayout, buffer)

	log.PipelineLogger.IfEnabled().Debug("Paint complete, err=%v", err)

	// Build HitMap from layout result with TargetFiber enrichment
	p.lastHitMap = convertLayoutHitMap(result.HitMap, fiber)

	if p.lastHitMap != nil {
		log.PipelineLogger.IfEnabled().Debug("Saved HitMap: %d entries", p.lastHitMap.Size())
	}

	return err
}

// RenderToSize renders with specific window size constraints
func (p *RenderingPipeline) RenderToSize(vnode rtui.VNode, width, height int, buffer *paint.Buffer) error {
	constraints := runtime.NewBoxConstraints(0, width, 0, height)
	return p.Render(vnode, nil, constraints, buffer)
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

// GetLayoutEngine returns the layout engine for direct access
func (p *RenderingPipeline) GetLayoutEngine() *layout.Engine {
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
func (p *RenderingPipeline) GetCacheStats() CacheStats {
	stats := p.layoutEngine.GetStats()
	return CacheStats{
		Hits:   int(stats.CacheHits),
		Misses: int(stats.CacheMisses),
	}
}

// ResetCacheStats resets cache hit/miss counters
func (p *RenderingPipeline) ResetCacheStats() {
	// Layout engine doesn't have this method yet
}

// ClearCache clears all cached layout results
func (p *RenderingPipeline) ClearCache() {
	p.layoutEngine.Invalidate()
}

// =============================================================================
// Multi-Layer Rendering
// =============================================================================

// RenderLayers renders a VNode tree with multi-layer support.
// This is the main entry point for layer-based rendering.
func (p *RenderingPipeline) RenderLayers(
	vnode rtui.VNode,
	fiber *reconciler.Fiber,
	constraints runtime.BoxConstraints,
	buffer *paint.Buffer,
) error {
	if vnode == nil {
		return nil
	}

	log.PipelineLogger.IfEnabled().Debug("RenderLayers started (layout.Engine path)")

	// Convert constraints to layout.Constraints
	layoutConstraints := layout.Constraints{
		MinWidth:  constraints.MinWidth,
		MaxWidth:  constraints.MaxWidth,
		MinHeight: constraints.MinHeight,
		MaxHeight: constraints.MaxHeight,
	}

	// Choose adapter based on whether we have Fiber or VNode
	var node layout.Node
	var converter PaintableConverter

	if fiber != nil {
		// Fiber-first path: use FiberToNodeAdapterPure
		node = NewFiberToNodeAdapterPure(fiber)
		converter = NewFiberToPaintableConverter(fiber)
	} else {
		// Legacy VNode path: use VNodeToNodeAdapter
		node = NewVNodeToNodeAdapter(vnode)
		converter = NewVNodeToPaintableConverter(vnode)
	}

	// Perform layout using layout.Engine
	result := p.layoutEngine.Layout(node, layoutConstraints)
	if result == nil || result.Root == nil {
		log.PipelineLogger.IfEnabled().Debug("Layout returned nil result")
		return nil
	}

	log.PipelineLogger.Debug("Layout complete, root size=%dx%d",
		result.Root.Width, result.Root.Height)

	// Convert LayoutBox to PaintableBox
	paintableLayout := converter.ConvertToLayout(result.Root)
	if paintableLayout == nil || paintableLayout.Root == nil {
		log.PipelineLogger.IfEnabled().Debug("PaintableLayout conversion failed")
		return nil
	}

	// ✨ Phase 1.1: Modal 居中逻辑已移到 Layout 阶段
	// applyLayerTransformsToPaintable 已不再需要，因为居中现在在 layoutNodeWithDepth() 中完成
	// p.applyLayerTransformsToPaintable(paintableLayout.Root, layoutConstraints)

	// Build PaintablePlanes from PaintableBox tree
	paintablePlanes := p.buildPaintablePlanes(paintableLayout.Root)
	log.PipelineLogger.IfEnabled().Debug("PaintablePlanes: %d boxes", paintablePlanes.CountBoxes())

	// Paint using PaintablePlanes
	if err := p.paintEngine.PaintPaintablePlanes(paintablePlanes, buffer); err != nil {
		log.PipelineLogger.IfEnabled().Debug("PaintPaintablePlanes failed: %v", err)
		return err
	}

	// Build HitMap from PaintablePlanes
	p.lastHitMap = p.buildHitMapFromPaintablePlanes(paintablePlanes)

	log.PipelineLogger.IfEnabled().Debug("RenderLayers complete")
	return nil
}

// PaintableConverter is the interface for converting LayoutBox to PaintableBox
type PaintableConverter interface {
	ConvertToLayout(lbox *layout.LayoutBox) *paint.PaintableLayout
}

// buildPaintablePlanes builds PaintablePlanes from PaintableBox tree
func (p *RenderingPipeline) buildPaintablePlanes(root *paint.PaintableBox) *paint.PaintablePlanes {
	pp := paint.NewPaintablePlanes()

	var walk func(box *paint.PaintableBox)
	walk = func(box *paint.PaintableBox) {
		if box == nil {
			return
		}

		pp.AddToLayer(paint.RenderLayer(box.Layer), box)

		for _, child := range box.Children {
			walk(child)
		}
	}

	walk(root)
	return pp
}

// buildHitMapFromPaintablePlanes builds HitMap from PaintablePlanes
func (p *RenderingPipeline) buildHitMapFromPaintablePlanes(pp *paint.PaintablePlanes) *event.HitMap {
	if pp == nil || pp.CountBoxes() == 0 {
		return nil
	}

	entries := make([]event.HitMapEntryInternal, 0)
	zOrder := 0

	// Iterate from highest to lowest layer for event handling
	pp.IterateReverse(func(layer paint.RenderLayer, box *paint.PaintableBox) bool {
		if box == nil || box.Node == nil {
			return true
		}
		boxX, boxY := box.X, box.Y

		// Get NodeID from PaintableNode
		nodeID := uint64(0)
		if id := box.Node.ID(); id != "" {
			nodeID = event.StringToNodeID(id)
		}

		// Create entry
		entries = append(entries, event.HitMapEntryInternal{
			NodeID: nodeID,
			Node:   nil, // PaintableBox doesn't have direct LayoutNode access
			Bounds: layout.Rect{
				X:      boxX,
				Y:      boxY,
				Width:  box.Width,
				Height: box.Height,
			},
			LocalXY: func(screenX, screenY int) (int, int) {
				return screenX - boxX, screenY - boxY
			},
			ZOrder: zOrder,
		})
		zOrder++

		return true
	})

	if len(entries) == 0 {
		return nil
	}

	return event.BuildHitMapFromEntries(entries)
}

// GetHitMap returns the HitMap from the most recent RenderLayers call
// This HitMap contains the FINAL positions after all layer transforms (centering, etc.)
// Returns nil if RenderLayers has not been called yet
func (p *RenderingPipeline) GetHitMap() *event.HitMap {
	return p.lastHitMap
}
