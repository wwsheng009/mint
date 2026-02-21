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
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// RenderingPipeline is the rendering pipeline with separated Layout and Paint phases.
//
// Architecture:
//   - Layout phase: compute.Engine calculates positions
//   - Paint phase: PaintEngine renders using computed positions
//
// For Fiber-first rendering, use fiberFirstPaint() in DeclarativeNode which
// uses NewLayoutEngineAdapter directly for better performance.
type RenderingPipeline struct {
	layoutEngine *compute.Engine // Layout engine (compute)
	paintEngine  *PaintEngine
	lastHitMap   *event.HitMap  // HitMap from the most recent RenderLayers call
	layerMgr     *layer.Manager // LayerManager from the most recent RenderLayers call
}

// NewRenderingPipeline creates a new rendering pipeline.
func NewRenderingPipeline() *RenderingPipeline {
	pipeline := &RenderingPipeline{
		layoutEngine: compute.NewEngine(),
		paintEngine:  NewPaintEngine(),
	}

	log.PipelineLogger.Debug("[RenderingPipeline] Initialized with compute.Engine")
	return pipeline
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
func (p *RenderingPipeline) Render(vnode rtui.VNode, fiber *reconciler.Fiber, constraints runtime.BoxConstraints, buffer *paint.Buffer) error {
	if vnode == nil {
		if os.Getenv("MINT_DEBUG_TEST") == "true" {
			fmt.Printf("[RenderingPipeline.Render] vnode is nil, returning\n")
		}
		return nil
	}

	if os.Getenv("MINT_DEBUG_TEST") == "true" {
		fmt.Printf("[RenderingPipeline.Render] START: vnode type=%d, tag=%s, buffer=%dx%d\n", vnode.Type(), vnode.Tag(), buffer.Width, buffer.Height)
	}

	log.PipelineLogger.Debug("Render started")

	// Phase 1: Layout - calculate all positions using compute.Engine
	layout, err := p.layoutEngine.Layout(vnode, fiber, constraints)
	if err != nil {
		log.PipelineLogger.Debug("❌ Layout FAILED: %v, falling back to legacy", err)
		if os.Getenv("MINT_DEBUG_TEST") == "true" {
			fmt.Printf("[RenderingPipeline.Render] Layout FAILED: %v\n", err)
		}
		return p.renderLegacy(vnode, 0, 0, buffer)
	}

	if os.Getenv("MINT_DEBUG_TEST") == "true" {
		if layout != nil && layout.Root != nil {
			fmt.Printf("[RenderingPipeline.Render] Layout: Root.Box=%dx%d, Root.VNode=%T\n", layout.Root.Box.Width, layout.Root.Box.Height, layout.Root.VNode)
		} else {
			fmt.Printf("[RenderingPipeline.Render] Layout: layout or Root is nil\n")
		}
	}

	log.PipelineLogger.Debug("✅ Layout complete, starting Paint phase")

	// Phase 2: Paint - render using computed positions
	log.PipelineLogger.Debug("Starting Paint phase...")
	err = p.paintEngine.Paint(layout, buffer)

	log.PipelineLogger.Debug("Paint complete, err=%v", err)

	// Save HitMap for event routing (hit testing)
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

	log.PipelineLogger.Debug("RenderLayers started")

	// Use legacy compute.Engine path via LayerManager
	layerMgr := layer.NewManager()
	if err := layerMgr.CollectAndLayout(vnode, fiber, constraints, p.layoutEngine); err != nil {
		log.PipelineLogger.Debug("Layer layout failed: %v, falling back to single-layer render", err)
		return p.Render(vnode, fiber, constraints, buffer)
	}

	renderPlanes := layerMgr.GetRenderPlanes()
	log.PipelineLogger.Debug("RenderPlanes: %d boxes", renderPlanes.CountBoxes())

	if err := p.paintEngine.PaintRenderPlanes(renderPlanes, buffer); err != nil {
		log.PipelineLogger.Debug("PaintRenderPlanes failed: %v", err)
		return err
	}

	p.lastHitMap = layerMgr.GetMergedHitMap()
	p.layerMgr = layerMgr

	if p.lastHitMap != nil {
		log.PipelineLogger.Debug("Merged HitMap: %d entries", p.lastHitMap.Size())
	}

	return nil
}

// applyLayerTransforms applies layer-specific transforms to a ComputedBox tree.
// This includes Modal centering and Inspector positioning.
// Mirrors the logic in layer/manager.go applyLayerTransforms.
func (p *RenderingPipeline) applyLayerTransforms(root *compute.ComputedBox, constraints runtime.BoxConstraints) {
	if root == nil {
		return
	}

	// Walk the ComputedBox tree and apply transforms based on Layer
	var walk func(box *compute.ComputedBox)
	walk = func(box *compute.ComputedBox) {
		if box == nil {
			return
		}

		// Apply Modal centering
		if box.Layer == rtui.LayerModal {
			p.centerModalBox(box, constraints)
		}

		// Recursively process children
		for _, child := range box.Children {
			walk(child)
		}
	}

	walk(root)
}

// centerModalBox centers a Modal box in the viewport.
// This shifts the entire box tree by (offsetX, offsetY).
func (p *RenderingPipeline) centerModalBox(box *compute.ComputedBox, constraints runtime.BoxConstraints) {
	if box == nil {
		return
	}

	modalWidth := box.Box.Width
	modalHeight := box.Box.Height
	containerWidth := constraints.MaxWidth
	containerHeight := constraints.MaxHeight

	if containerWidth == runtime.Infinity {
		containerWidth = modalWidth
	}
	if containerHeight == runtime.Infinity {
		containerHeight = modalHeight
	}

	offsetX := (containerWidth - modalWidth) / 2
	offsetY := (containerHeight - modalHeight) / 2

	if offsetX < 0 {
		offsetX = 0
	}
	if offsetY < 0 {
		offsetY = 0
	}

	log.PipelineLogger.Debug("centerModalBox: modal=%dx%d container=%dx%d offset=(%d,%d)",
		modalWidth, modalHeight, containerWidth, containerHeight, offsetX, offsetY)

	// Shift the entire box tree
	p.shiftBoxTree(box, offsetX, offsetY)
}

// shiftBoxTree shifts all boxes in a ComputedBox tree by the given offset.
func (p *RenderingPipeline) shiftBoxTree(box *compute.ComputedBox, offsetX, offsetY int) {
	if box == nil {
		return
	}

	box.Box.X += offsetX
	box.Box.Y += offsetY

	for _, child := range box.Children {
		p.shiftBoxTree(child, offsetX, offsetY)
	}
}

// applyModalCenteringToRenderPlanes applies centering transform to Modal layer boxes in RenderPlanes.
// This is used for runtime/layout engine results where we have RenderPlanes but not the original ComputedBox tree.
func (p *RenderingPipeline) applyModalCenteringToRenderPlanes(renderPlanes *layer.RenderPlanes, constraints runtime.BoxConstraints) {
	if renderPlanes == nil {
		return
	}

	// Get all Modal layer boxes
	modalBoxes := renderPlanes.GetLayer(rtui.LayerModal)
	if len(modalBoxes) == 0 {
		return
	}

	// Calculate container dimensions
	containerWidth := constraints.MaxWidth
	containerHeight := constraints.MaxHeight

	// Apply centering to each modal box
	for _, box := range modalBoxes {
		if box == nil {
			continue
		}

		modalWidth := box.Box.Width
		modalHeight := box.Box.Height

		if containerWidth == runtime.Infinity {
			containerWidth = modalWidth
		}
		if containerHeight == runtime.Infinity {
			containerHeight = modalHeight
		}

		offsetX := (containerWidth - modalWidth) / 2
		offsetY := (containerHeight - modalHeight) / 2

		if offsetX < 0 {
			offsetX = 0
		}
		if offsetY < 0 {
			offsetY = 0
		}

		log.PipelineLogger.Debug("applyModalCenteringToRenderPlanes: modal=%dx%d container=%dx%d offset=(%d,%d)",
			modalWidth, modalHeight, containerWidth, containerHeight, offsetX, offsetY)

		// Apply offset to this box and all its children
		p.shiftBoxTree(box, offsetX, offsetY)
	}
}

// applyModalCenteringToLayoutBox applies Modal centering to the layout.LayoutBox tree.
// This is used for runtime/layout engine results.
func (p *RenderingPipeline) applyModalCenteringToLayoutBox(result *layout.LayoutResult, constraints runtime.BoxConstraints) {
	if result == nil || result.Root == nil {
		return
	}

	// Walk the LayoutBox tree and apply centering to Modal layer nodes
	var walk func(box *layout.LayoutBox)
	walk = func(box *layout.LayoutBox) {
		if box == nil {
			return
		}

		// Check if this box is on Modal layer
		if box.Layer == layout.LayerModal {
			// Calculate centering offset
			containerWidth := constraints.MaxWidth
			containerHeight := constraints.MaxHeight

			if containerWidth == runtime.Infinity {
				containerWidth = box.Width
			}
			if containerHeight == runtime.Infinity {
				containerHeight = box.Height
			}

			offsetX := (containerWidth - box.Width) / 2
			offsetY := (containerHeight - box.Height) / 2

			if offsetX < 0 {
				offsetX = 0
			}
			if offsetY < 0 {
				offsetY = 0
			}

			log.PipelineLogger.Debug("applyModalCenteringToLayoutBox: modal=%dx%d container=%dx%d offset=(%d,%d)",
				box.Width, box.Height, containerWidth, containerHeight, offsetX, offsetY)

			// Shift this box and all its children
			p.shiftLayoutBoxTree(box, offsetX, offsetY)
		} else {
			// Only recurse if not a Modal box (Modal children are shifted with parent)
			for _, child := range box.Children {
				walk(child)
			}
		}
	}

	walk(result.Root)
}

// shiftLayoutBoxTree shifts a layout.LayoutBox and all its children by the given offset.
func (p *RenderingPipeline) shiftLayoutBoxTree(box *layout.LayoutBox, offsetX, offsetY int) {
	if box == nil {
		return
	}

	box.X += offsetX
	box.Y += offsetY

	for _, child := range box.Children {
		p.shiftLayoutBoxTree(child, offsetX, offsetY)
	}
}

// convertLayoutBoxToComputedLayout converts a layout.LayoutResult to compute.ComputedLayout.
// This is the key bridge between runtime/layout engine and PaintEngine.
//
// Fiber-First Architecture:
// - LayoutBox contains position and Layer info
// - ComputedBox needs VNode reference for PaintEngine
// - We traverse Fiber tree in parallel with LayoutBox tree to match positions with VNodes
func (p *RenderingPipeline) convertLayoutBoxToComputedLayout(
	result *layout.LayoutResult,
	fiber *reconciler.Fiber,
	constraints runtime.BoxConstraints,
) *compute.ComputedLayout {
	if result == nil || result.Root == nil {
		return nil
	}

	// Build a map from LayoutBox ID to LayoutBox for quick lookup
	// The ID in LayoutBox comes from FiberToNodeAdapter.ID() which uses Fiber.DiffKey
	boxMap := make(map[string]*layout.LayoutBox)
	var collectBoxes func(box *layout.LayoutBox)
	collectBoxes = func(box *layout.LayoutBox) {
		if box == nil {
			return
		}
		if box.ID != "" {
			boxMap[box.ID] = box
		}
		for _, child := range box.Children {
			collectBoxes(child)
		}
	}
	collectBoxes(result.Root)

	// Build a map from Fiber key to Fiber for matching
	// We use multiple keys for matching:
	// 1. DiffKey (primary)
	// 2. Key (alias for DiffKey)
	// 3. NodeID as string (for matching with LayoutBox.ID from FiberToNodeAdapter)
	fiberMap := make(map[string]*reconciler.Fiber)
	var collectFibers func(f *reconciler.Fiber)
	collectFibers = func(f *reconciler.Fiber) {
		if f == nil {
			return
		}
		// Add by DiffKey
		if f.DiffKey != "" {
			fiberMap[f.DiffKey] = f
		}
		// Add by Key (if different from DiffKey)
		if f.Key != "" && f.Key != f.DiffKey {
			fiberMap[f.Key] = f
		}
		// Add by NodeID string (for matching with FiberToNodeAdapter.ID())
		nodeIDKey := fmt.Sprintf("%d", f.NodeID)
		fiberMap[nodeIDKey] = f
		
		// Recursively collect children
		for child := f.Child; child != nil; child = child.Sibling {
			collectFibers(child)
		}
	}
	collectFibers(fiber)

	// Convert LayoutBox tree to ComputedBox tree
	rootBox := p.convertLayoutBoxToComputedBox(result.Root, fiber, boxMap, fiberMap, nil)

	return compute.NewComputedLayout(rootBox)
}

// convertLayoutBoxToComputedBox converts a single LayoutBox to ComputedBox with VNode reference.
func (p *RenderingPipeline) convertLayoutBoxToComputedBox(
	lbox *layout.LayoutBox,
	fiber *reconciler.Fiber,
	boxMap map[string]*layout.LayoutBox,
	fiberMap map[string]*reconciler.Fiber,
	parent *compute.ComputedBox,
) *compute.ComputedBox {
	if lbox == nil {
		return nil
	}

	// Convert layout.Layer to rtui.Layer
	rtuiLayer := convertLayoutLayerToRTUI(lbox.Layer)

	// Create ComputedBox with position from LayoutBox
	cbox := &compute.ComputedBox{
		Box: runtime.Box{
			X:      lbox.X,
			Y:      lbox.Y,
			Width:   lbox.Width,
			Height:  lbox.Height,
		},
		Layer:      rtuiLayer,
		Parent:     parent,
		Children:   make([]*compute.ComputedBox, 0),
		LayoutDirty: false,
	}

	// Try to find the Fiber node by multiple matching strategies
	// Strategy 1: Direct ID match (lbox.ID == fiber.DiffKey/Key)
	// Strategy 2: NodeID match (lbox.ID == fmt.Sprintf("%d", fiber.NodeID))
	var matchingFiber *reconciler.Fiber
	
	// First try direct ID match
	if f, ok := fiberMap[lbox.ID]; ok {
		matchingFiber = f
	} else {
		// Fallback: search by NodeID
		// LayoutBox.ID may be NodeID format ("95") while fiberMap uses DiffKey
		for _, f := range fiberMap {
			if fmt.Sprintf("%d", f.NodeID) == lbox.ID {
				matchingFiber = f
				break
			}
		}
	}

	if matchingFiber != nil {
		cbox.NodeID = matchingFiber.NodeID
		cbox.DiffKey = matchingFiber.DiffKey // Fiber-first: copy DiffKey
		cbox.ChildFiber = matchingFiber
		// Wrap Fiber as VNode using FiberVNode (transitional approach)
		// Fiber-first: VNode reference is not stored in Fiber, we wrap it
		cbox.VNode = rtui.NewFiberVNode(matchingFiber)
	}

	// Convert children
	for _, childLBox := range lbox.Children {
		childCBox := p.convertLayoutBoxToComputedBox(childLBox, fiber, boxMap, fiberMap, cbox)
		if childCBox != nil {
			cbox.Children = append(cbox.Children, childCBox)
		}
	}

	return cbox
}

// convertLayoutLayerToRTUI converts layout.Layer to rtui.Layer.
func convertLayoutLayerToRTUI(l layout.Layer) rtui.Layer {
	switch l {
	case layout.LayerBase:
		return rtui.LayerBase
	case layout.LayerDropdown, layout.LayerSticky, layout.LayerFixed:
		return rtui.LayerOverlay
	case layout.LayerModalBackdrop, layout.LayerModal:
		return rtui.LayerModal
	case layout.LayerPopover, layout.LayerTooltip:
		return rtui.LayerTooltip
	default:
		return rtui.LayerBase
	}
}

// =============================================================================
// New Simplified Conversion: LayoutBox → PaintableBox
// =============================================================================

// convertLayoutBoxToPaintableLayout converts a layout.LayoutResult directly to PaintableLayout.
// This is the NEW simplified path that bypasses ComputedBox entirely.
//
// Data Flow:
//
//	Fiber → LayoutBox → PaintableBox → Paint
//	        (layout)    (paint)
//
// This method uses FiberToPaintableConverter for the conversion.
func (p *RenderingPipeline) convertLayoutBoxToPaintableLayout(
	result *layout.LayoutResult,
	fiber *reconciler.Fiber,
) *paint.PaintableLayout {
	if result == nil || result.Root == nil {
		return nil
	}

	// Create converter with Fiber index
	converter := NewFiberToPaintableConverter(fiber)

	// Convert LayoutBox tree to PaintableBox tree
	return converter.ConvertToLayout(result.Root)
}

// buildHitMapFromLayoutResult builds an event.HitMap from layout.LayoutResult.
func (p *RenderingPipeline) buildHitMapFromLayoutResult(result *layout.LayoutResult) *event.HitMap {
	if result == nil {
		return nil
	}

	var entries []event.HitMapEntryInternal

	// Build HitMap from LayoutBox tree
	var walk func(box *layout.LayoutBox)
	walk = func(box *layout.LayoutBox) {
		if box == nil {
			return
		}

		// Add entry for this box if it has valid bounds
		if box.Width > 0 && box.Height > 0 {
			// Convert layout.Layer to rtui.Layer for ZOrder calculation
			rtuiLayer := convertLayoutLayerToRTUI(box.Layer)
			zOrder := int(rtuiLayer) * 1000 + box.ZIndex

			entry := event.HitMapEntryInternal{
				NodeID: event.StringToNodeID(box.ID),
				Bounds: layout.Rect{
					X:      box.X,
					Y:      box.Y,
					Width:  box.Width,
					Height: box.Height,
				},
				ZOrder: zOrder,
				LocalXY: func(screenX, screenY int) (int, int) {
					return screenX - box.X, screenY - box.Y
				},
			}
			entries = append(entries, entry)
		}

		// Recurse into children
		for _, child := range box.Children {
			walk(child)
		}
	}

	walk(result.Root)

	hitMap := event.BuildHitMapFromEntries(entries)

	log.PipelineLogger.Debug("buildHitMapFromLayoutResult: Built HitMap with %d entries", hitMap.Size())

	return hitMap
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
