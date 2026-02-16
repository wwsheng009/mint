// Package render provides new rendering pipeline with separated Layout and Paint phases
package render

import (
	"fmt"
	"os"

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
// Layout phase: LayoutSwitcher (can use compute.Engine or layout.Engine)
// Paint phase: PaintEngine renders using computed positions
type RenderingPipeline struct {
	layoutEngine *compute.Engine // Legacy: direct compute engine (kept for compatibility)
	switcher     *LayoutSwitcher // New: switchable layout engine
	paintEngine  *PaintEngine
	lastHitMap   *event.HitMap  // HitMap from the most recent RenderLayers call
	layerMgr     *layer.Manager // LayerManager from the most recent RenderLayers call
	useSwitcher  bool           // Whether to use the switcher (based on MINT_LAYOUT_ENGINE)
}

// NewRenderingPipeline creates a new rendering pipeline
func NewRenderingPipeline() *RenderingPipeline {
	pipeline := &RenderingPipeline{
		layoutEngine: compute.NewEngine(),
		paintEngine:  NewPaintEngine(),
	}

	// Check if we should use the switcher (new layout engine)
	envEngine := os.Getenv("MINT_LAYOUT_ENGINE")
	if envEngine != "" && envEngine != "compute" {
		pipeline.switcher = NewLayoutSwitcher()
		pipeline.useSwitcher = true
		log.PipelineLogger.Debug("[RenderingPipeline] Using LayoutSwitcher with engine: %s", pipeline.switcher.GetEngineType())
	} else {
		pipeline.useSwitcher = false
		log.PipelineLogger.Debug("[RenderingPipeline] Using legacy compute.Engine")
	}

	return pipeline
}

// SetLayoutDebug enables/disables layout debug output
func (p *RenderingPipeline) SetLayoutDebug(debug bool) {
	p.layoutEngine.SetDebug(debug)
	if p.switcher != nil {
		p.switcher.SetDebug(debug)
	}
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
		if os.Getenv("MINT_DEBUG_TEST") == "true" {
			fmt.Printf("[RenderingPipeline.Render] vnode is nil, returning\n")
		}
		return nil
	}

	if os.Getenv("MINT_DEBUG_TEST") == "true" {
		fmt.Printf("[RenderingPipeline.Render] START: vnode type=%d, tag=%s, buffer=%dx%d\n", vnode.Type(), vnode.Tag(), buffer.Width, buffer.Height)
		fmt.Printf("[RenderingPipeline.Render] useSwitcher=%v, switcher=%v\n", p.useSwitcher, p.switcher != nil)
	}

	log.PipelineLogger.Debug("Render started")

	var layout *compute.ComputedLayout
	var err error

	// Phase 1: Layout - calculate all positions
	if p.useSwitcher && p.switcher != nil {
		// Use the switcher (supports new layout engine)
		log.PipelineLogger.Debug("Using LayoutSwitcher with engine: %s", p.switcher.GetEngineType())
		result, layoutErr := p.switcher.Layout(vnode, fiber, constraints)
		if layoutErr != nil {
			log.PipelineLogger.Debug("❌ Layout FAILED: %v, falling back to legacy", layoutErr)
			if os.Getenv("MINT_DEBUG_TEST") == "true" {
				fmt.Printf("[RenderingPipeline.Render] Layout FAILED: %v\n", layoutErr)
			}
			return p.renderLegacy(vnode, 0, 0, buffer)
		}

		if os.Getenv("MINT_DEBUG_TEST") == "true" {
			fmt.Printf("[RenderingPipeline.Render] Layout result type=%T\n", result)
		}

		// Convert LayoutResult to ComputedLayout for PaintEngine
		if adapter, ok := result.(*computeLayoutResultAdapter); ok {
			layout = adapter.ComputedLayout
			if os.Getenv("MINT_DEBUG_TEST") == "true" {
				if layout != nil && layout.Root != nil {
					fmt.Printf("[RenderingPipeline.Render] computeLayoutResultAdapter: Root.Box=%dx%d\n", layout.Root.Box.Width, layout.Root.Box.Height)
				} else {
					fmt.Printf("[RenderingPipeline.Render] computeLayoutResultAdapter: layout.Root is nil\n")
				}
			}
		} else {
			// For new layout engine, we need different handling
			log.PipelineLogger.Debug("New layout engine result - converting for paint")
			if os.Getenv("MINT_DEBUG_TEST") == "true" {
				fmt.Printf("[RenderingPipeline.Render] NOT computeLayoutResultAdapter, using legacy fallback\n")
			}
			// For now, use the legacy engine as fallback for painting
			layout, err = p.layoutEngine.Layout(vnode, fiber, constraints)
			if err != nil {
				if os.Getenv("MINT_DEBUG_TEST") == "true" {
					fmt.Printf("[RenderingPipeline.Render] Legacy layout FAILED: %v\n", err)
				}
				return p.renderLegacy(vnode, 0, 0, buffer)
			}
			if os.Getenv("MINT_DEBUG_TEST") == "true" {
				if layout != nil && layout.Root != nil {
					fmt.Printf("[RenderingPipeline.Render] Legacy layout: Root.Box=%dx%d, Root.VNode=%T\n", layout.Root.Box.Width, layout.Root.Box.Height, layout.Root.VNode)
				} else {
					fmt.Printf("[RenderingPipeline.Render] Legacy layout: layout or Root is nil\n")
				}
			}
		}
	} else {
		// Use legacy compute engine
		log.PipelineLogger.Debug("Using legacy compute.Engine")
		layout, err = p.layoutEngine.Layout(vnode, fiber, constraints)
		if err != nil {
			log.PipelineLogger.Debug("❌ Layout FAILED: %v, falling back to legacy", err)
			return p.renderLegacy(vnode, 0, 0, buffer)
		}
	}

	log.PipelineLogger.Debug("✅ Layout complete, starting Paint phase")

	// Phase 2: Paint - render using computed positions
	log.PipelineLogger.Debug("Starting Paint phase...")
	if os.Getenv("MINT_DEBUG_TEST") == "true" {
		if layout != nil && layout.Root != nil {
			fmt.Printf("[RenderingPipeline.Render] Calling paintEngine.Paint, layout.Root.Box=%dx%d, buffer=%dx%d\n",
				layout.Root.Box.Width, layout.Root.Box.Height, buffer.Width, buffer.Height)
		} else {
			fmt.Printf("[RenderingPipeline.Render] layout or layout.Root is nil, NOT painting\n")
		}
	}
	err = p.paintEngine.Paint(layout, buffer)

	log.PipelineLogger.Debug("Paint complete, err=%v", err)

	// Save HitMap for event routing (hit testing)
	// This HitMap contains the FINAL positions from layout computation
	if layout != nil {
		p.lastHitMap = layout.HitMap
	}

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

// GetLayoutEngineType returns the current layout engine type
// Returns "compute" for legacy engine, "layout" for new engine, "both" for comparison mode
func (p *RenderingPipeline) GetLayoutEngineType() string {
	if p.useSwitcher && p.switcher != nil {
		return p.switcher.GetEngineType().String()
	}
	return "compute"
}

// GetSwitcher returns the LayoutSwitcher if available, nil otherwise
func (p *RenderingPipeline) GetSwitcher() *LayoutSwitcher {
	return p.switcher
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

	// Phase 5.5: NEW - Call LayoutFiber() to populate fiber.ComputedBox
	// This ensures BuildHitMapFromFiber() has complete data
	// Note: This is experimental until Phase 5 completes (modal centering moves to Layout Engine)
	if fiber != nil {
		_, err := p.layoutEngine.LayoutFiber(fiber, constraints)
		if err != nil {
			log.PipelineLogger.Debug("LayoutFiber failed: %v", err)
		} else {
			log.PipelineLogger.Debug("LayoutFiber completed - fiber.ComputedBox should now be populated")
		}
	}

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

	// Phase 5.5: NEW - Try BuildHitMapFromFiber first (if LayoutFiber was called and fiber.ComputedBox is populated)
	if fiber != nil {
		fiberHitMap := rtui.BuildHitMapFromFiber(fiber)
		if fiberHitMap.Size() > 0 {
			p.lastHitMap = fiberHitMap
			log.PipelineLogger.Debug("✅ BuildHitMapFromFiber succeeded: %d entries", fiberHitMap.Size())
		} else {
			log.PipelineLogger.Debug("⚠️ BuildHitMapFromFiber returned 0 entries, falling back to LayerManager")
			// Fallback to LayerManager if BuildHitMapFromFiber fails
			if layerMgr != nil {
				p.lastHitMap = layerMgr.GetMergedHitMap()
				log.PipelineLogger.Debug("Got HitMap from LayerManager: %d entries", p.lastHitMap.Size())
			}
		}
	} else {
		// No Fiber - use LayerManager fallback
		if layerMgr != nil {
			p.lastHitMap = layerMgr.GetMergedHitMap()
			log.PipelineLogger.Debug("Got HitMap from LayerManager (no Fiber): %d entries", p.lastHitMap.Size())
		}
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
