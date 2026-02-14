// Package render provides new rendering pipeline with separated Layout and Paint phases
package render

import (
	"github.com/wwsheng009/mint/internal/log"
	"github.com/wwsheng009/mint/internal/reconciler"
	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/compute"
	"github.com/wwsheng009/mint/runtime/event"
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
	lastHitMap   *event.HitMap  // HitMap from the most recent RenderLayers call
	layerMgr     *layer.Manager // LayerManager from the most recent RenderLayers call
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
//
// Phase 8: Added optional fiber parameter for NodeID propagation
func (p *RenderingPipeline) Render(vnode rtui.VNode, fiber *reconciler.Fiber, constraints runtime.BoxConstraints, buffer *paint.Buffer) error {
	if vnode == nil {
		return nil
	}

	log.PipelineLogger.Debug("Render started")

	// Phase 1: Layout - calculate all positions
	// Phase 8: Pass Fiber to layout engine for NodeID propagation
	layout, err := p.layoutEngine.Layout(vnode, fiber, constraints)
	if err != nil {
		// Fallback to legacy rendering if layout fails
		log.PipelineLogger.Debug("❌ Layout FAILED: %v, falling back to legacy", err)
		return p.renderLegacy(vnode, 0, 0, buffer)
	}

	log.PipelineLogger.Debug("✅ Layout complete, starting Paint phase")

	// Phase 2: Paint - render using computed positions
	log.PipelineLogger.Debug("Starting Paint phase...")
	err = p.paintEngine.Paint(layout, buffer)

	log.PipelineLogger.Debug("Paint complete, err=%v", err)

	// Save HitMap for event routing (hit testing)
	// This HitMap contains the FINAL positions from layout computation
	p.lastHitMap = layout.HitMap

	if p.lastHitMap != nil {
		log.PipelineLogger.Debug("Saved HitMap: %d entries", p.lastHitMap.Size())
	} else {
		log.PipelineLogger.Debug("⚠️ Layout.HitMap is nil")
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

// ComputeLayout performs only the layout phase, returning computed positions
// This can be useful for hit testing and other operations that need layout info without rendering
// Phase 8: Accept fiber parameter for NodeID propagation
func (p *RenderingPipeline) ComputeLayout(vnode rtui.VNode, fiber *reconciler.Fiber, constraints runtime.BoxConstraints) (*compute.ComputedLayout, error) {
	return p.layoutEngine.Layout(vnode, fiber, constraints)
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
// RenderWithFiber renders with explicit Fiber tree for NodeID propagation
// =============================================================================
// Phase 8: This method allows passing the Fiber tree directly to the layout engine
// so that NodeIDs can be propagated to ComputedBox for proper identity tracking
//
// This is the primary entry point for Fiber-based rendering, where:
// - vnode: The rendered VNode tree (may differ from Fiber tree structure)
// - fiber: The actual Fiber tree with NodeIDs assigned during reconciliation
func (p *RenderingPipeline) RenderWithFiber(
	vnode rtui.VNode,
	fiber *reconciler.Fiber,
	constraints runtime.BoxConstraints,
	buffer *paint.Buffer,
) error {
	if vnode == nil {
		return nil
	}

	log.PipelineLogger.Debug("RenderWithFiber started")

	// Phase 1: Layout - calculate all positions with Fiber
	layout, err := p.layoutEngine.Layout(vnode, fiber, constraints)
	if err != nil {
		// Fallback to legacy rendering if layout fails
		log.PipelineLogger.Debug("❌ Layout FAILED: %v, falling back to legacy", err)
		return p.renderLegacy(vnode, 0, 0, buffer)
	}

	log.PipelineLogger.Debug("✅ Layout complete, starting Paint phase")

	// Phase 2: Paint - render using computed positions
	log.PipelineLogger.Debug("Starting Paint phase...")
	err = p.paintEngine.Paint(layout, buffer)

	log.PipelineLogger.Debug("Paint complete, err=%v", err)

	// Save HitMap for event routing (hit testing)
	// This HitMap contains FINAL positions from layout computation
	p.lastHitMap = layout.HitMap

	if p.lastHitMap != nil {
		log.PipelineLogger.Debug("Saved HitMap: %d entries", p.lastHitMap.Size())
	} else {
		log.PipelineLogger.Debug("⚠️  Layout.HitMap is nil")
	}

	return err
}

// =============================================================================
// Multi-Layer Rendering
// =============================================================================

// RenderLayers renders a VNode tree with multi-layer support
// This is the main entry point for layer-based rendering
//
// Phase 8: Added optional fiber parameter for NodeID propagation
func (p *RenderingPipeline) RenderLayers(
	vnode rtui.VNode,
	fiber *reconciler.Fiber,
	constraints runtime.BoxConstraints,
	buffer *paint.Buffer,
) error {
	if vnode == nil {
		return nil
	}

	log.PipelineLogger.Debug("RenderLayers started")

	// Create a layer manager for this render pass
	layerMgr := layer.NewManager()

	// Phase 5: New unified architecture - Build RenderPlanes directly from Fiber
	// Note: Still using CollectAndLayout temporarily for modal centering until
	// it moves to Layout Engine in Phase 5-7
	if err := layerMgr.CollectAndLayout(vnode, fiber, constraints, p.layoutEngine); err != nil {
		log.PipelineLogger.Debug("Layer layout failed: %v", err)
		// Fallback to regular rendering
		return p.Render(vnode, fiber, constraints, buffer)
	}

	// Phase 5: Get the RenderPlanes already built by CollectAndLayout
	// DO NOT call BuildRenderPlanes(fiber) - it will overwrite the 58 boxes with 0 boxes
	renderPlanes := layerMgr.GetRenderPlanes()
	log.PipelineLogger.Debug("Got RenderPlanes: %d boxes", renderPlanes.CountBoxes())

	// Phase 5: Paint all layers from RenderPlanes (new PaintRenderPlanes method)
	if err := p.paintEngine.PaintRenderPlanes(renderPlanes, buffer); err != nil {
		log.PipelineLogger.Debug("PaintRenderPlanes failed: %v", err)
		return err
	}

	// Phase 6: Get HitMap from LayerManager (already built by CollectAndLayout)
	// DO NOT use BuildHitMapFromFiber - it has 0 entries because fiber.ComputedBox is nil
	if layerMgr != nil {
		p.lastHitMap = layerMgr.GetMergedHitMap()
		log.PipelineLogger.Debug("Got HitMap from LayerManager")
	}

	// Save layerMgr reference for event handling
	// This allows DeclarativeNode to access modal nodes for mouse event distribution
	p.layerMgr = layerMgr

	if p.lastHitMap != nil {
		log.PipelineLogger.Debug("Merged HitMap: %d entries", p.lastHitMap.Size())
	}
	if p.layerMgr != nil {
		modalNodes := p.layerMgr.GetModalNodes()
		log.PipelineLogger.Debug("Saved layerMgr with %d modal nodes", len(modalNodes))
	}

	// 验证 buffer 内容
	if log.PipelineLogger.Enabled() {
		contentCount := 0
		for y := 0; y < buffer.Height; y++ {
			for x := 0; x < buffer.Width; x++ {
				if buffer.Cells[y][x].Cluster != "" {
					contentCount++
				}
			}
		}
		log.PipelineLogger.Debug("Buffer content after PaintLayers: %d cells (buffer size: %dx%d)",
			contentCount, buffer.Width, buffer.Height)

		if contentCount == 0 {
			log.PipelineLogger.Debug("⚠️  WARNING: Buffer is empty!")
		} else {
			log.PipelineLogger.Debug("✅ Buffer has content")
		}
	}

	return nil
}

// HasModalChecks returns whether the rendering pipeline detected any modal content
// This can be used to determine if events should be blocked
// Phase 8: Accept fiber parameter for NodeID propagation consistency
func (p *RenderingPipeline) HasModalChecks(vnode rtui.VNode, fiber *reconciler.Fiber, constraints runtime.BoxConstraints) bool {
	layerMgr := layer.NewManager()
	layerMgr.CollectAndLayout(vnode, fiber, constraints, p.layoutEngine)
	return layerMgr.HasModal()
}

// GetHitMap returns the HitMap from the most recent RenderLayers call
// This HitMap contains the FINAL positions after all layer transforms (centering, etc.)
// Returns nil if RenderLayers has not been called yet
func (p *RenderingPipeline) GetHitMap() *event.HitMap {
	return p.lastHitMap
}

// GetLayerMgr returns the LayerManager from the most recent RenderLayers call
// This LayerManager contains modal nodes for event distribution
// Returns nil if RenderLayers has not been called yet
func (p *RenderingPipeline) GetLayerMgr() *layer.Manager {
	return p.layerMgr
}
