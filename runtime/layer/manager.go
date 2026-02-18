// Package layer provides layer management for multi-layer TUI rendering
package layer

import (
	"sort"

	"github.com/wwsheng009/mint/internal/log"
	"github.com/wwsheng009/mint/internal/reconciler"
	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/compute"
	"github.com/wwsheng009/mint/runtime/event"
	runtimelayout "github.com/wwsheng009/mint/runtime/layout"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// LayerManager
// =============================================================================

// Manager manages multiple rendering layers
// It handles collection, layout, and coordination of layer-based rendering
type Manager struct {
	collector    *Collector
	layouts      LayerLayouts
	renderPlanes *RenderPlanes // Phase 3: RenderPlanes for unified layer management
}

// LayerLayouts maps each Layer to its computed layout
type LayerLayouts map[rtui.Layer]*compute.ComputedLayout

// NewManager creates a new layer manager
func NewManager() *Manager {
	return &Manager{
		collector:    NewCollector(),
		layouts:      make(LayerLayouts),
		renderPlanes: NewRenderPlanes(), // Phase 3: Initialize RenderPlanes
	}
}

// =============================================================================
// Main API
// =============================================================================

// CollectAndLayout performs Fiber-first single-pass layout with layer grouping.
// This is the main entry point for layer-based rendering.
//
// Fiber-First Architecture (per docs/fiber/diff_layer.md):
// 1. Single layout pass on entire Fiber tree (NO StripLayers, NO VNode clone)
// 2. Layer is just a grouping attribute on ComputedBox
// 3. Build RenderPlanes by grouping ComputedBoxes by Layer
// 4. Apply layer-specific transforms (modal centering) as post-processing
//
// Workflow:
//  1. Single layout pass on Fiber tree (BuildComputedBoxFiberOnly)
//  2. Build RenderPlanes by grouping ComputedBoxes by their Layer field
//  3. Apply layer-specific transforms (modal centering, inspector positioning)
//  4. Build merged HitMap from transformed positions
//
// Parameters:
//
//	vnode: The VNode tree (used when Fiber is nil for backward compatibility)
//	fiber: The Fiber tree with NodeID and Layer info (PREFERRED)
//	constraints: Box constraints for layout
//	engine: Layout engine for computation
func (m *Manager) CollectAndLayout(
	vnode rtui.VNode,
	fiber *reconciler.Fiber,
	constraints runtime.BoxConstraints,
	engine *compute.Engine,
) error {
	m.layouts = make(LayerLayouts)
	m.renderPlanes = NewRenderPlanes()

	if fiber != nil {
		log.LayerLogger.Debug("[CollectAndLayout] ✅ Using Fiber-first single-pass layout")
		return m.collectAndLayoutFiberFirst(fiber, constraints, engine)
	}

	log.LayerLogger.Debug("[CollectAndLayout] ⚠️ Using legacy VNode-based layout (no Fiber)")
	return m.collectAndLayoutLegacy(vnode, constraints, engine)
}

// collectAndLayoutFiberFirst implements Fiber-first layout with proper layer handling.
//
// Fiber-First Architecture (per docs/fiber/diff_layer.md):
// 1. Single layout pass on entire tree (using engine.Layout with Fiber for NodeID propagation)
// 2. Build RenderPlanes by grouping ComputedBoxes by their Layer field
// 3. Apply layer-specific transforms (modal centering) as post-processing
// 4. Build merged HitMap from transformed positions
//
// Key Principle: Layer is just a grouping attribute, not a structural change.
// Modal nodes stay in the Fiber tree but need centering transform applied.
func (m *Manager) collectAndLayoutFiberFirst(
	fiber *reconciler.Fiber,
	constraints runtime.BoxConstraints,
	engine *compute.Engine,
) error {
	// Use engine.Layout with Fiber for NodeID propagation
	// This is the stable path that correctly handles all node types
	// BuildComputedBoxFiberOnly is still experimental and has issues with child measurement
	
	// Get VNode from Fiber for layout
	vnode := rtui.NewFiberVNode(fiber)
	
	layout, err := engine.Layout(vnode, fiber, constraints)
	if err != nil {
		log.LayerLogger.Debug("[CollectAndLayoutFiberFirst] Layout failed: %v", err)
		return err
	}

	if layout.Root == nil {
		log.LayerLogger.Debug("[CollectAndLayoutFiberFirst] Empty layout result")
		return nil
	}

	log.LayerLogger.Debug("[CollectAndLayoutFiberFirst] Layout complete, root size=%dx%d",
		layout.Root.Box.Width, layout.Root.Box.Height)

	m.layouts[rtui.LayerBase] = layout

	// Build RenderPlanes from the computed layout
	m.renderPlanes = BuildRenderPlane(layout.Root)
	boxCount := m.renderPlanes.CountBoxes()
	log.LayerLogger.Debug("[CollectAndLayoutFiberFirst] Built RenderPlanes with %d boxes", boxCount)

	// Apply layer-specific transforms (modal centering)
	m.applyLayerTransforms(layout.Root, constraints)

	// Rebuild HitMap with final transformed positions
	if layout.HitMap != nil {
		layout.HitMap = m.buildMergedHitMapFromRenderPlanes()
	}

	return nil
}

// collectAndLayoutLegacy implements legacy VNode-based layout for backward compatibility.
// This path uses StripLayers and is DEPRECATED - only used when Fiber is nil.
func (m *Manager) collectAndLayoutLegacy(
	vnode rtui.VNode,
	constraints runtime.BoxConstraints,
	engine *compute.Engine,
) error {
	m.collector.Collect(vnode)
	log.LayerLogger.Debug("[CollectAndLayoutLegacy] collected %d modal nodes", len(m.collector.GetModalNodes()))

	baseTree := m.collector.StripLayers(vnode)
	baseChildren := baseTree.Children()
	log.LayerLogger.Debug("[CollectAndLayoutLegacy] baseTree has %d children (after stripping)", len(baseChildren))

	baseLayout, err := engine.Layout(baseTree, nil, constraints)
	if err != nil {
		return err
	}
	m.layouts[rtui.LayerBase] = baseLayout

	for layer, nodes := range m.collector.GetLayers() {
		if len(nodes) == 0 {
			continue
		}
		for _, node := range nodes {
			if !node.Visible {
				continue
			}
			layerLayout, err := m.layoutLayer(node, layer, constraints, engine, nil)
			if err != nil {
				return err
			}
			m.layouts[layer] = layerLayout
			break
		}
	}

	m.renderPlanes = BuildRenderPlanesFromLayouts(m.layouts)
	log.LayerLogger.Debug("[CollectAndLayoutLegacy] Built RenderPlanes with %d boxes", m.renderPlanes.CountBoxes())

	return nil
}

// applyLayerTransforms applies layer-specific transformations after layout.
// Modal: centered in viewport
// Inspector: positioned at specified coordinates
//
// Fiber-First Architecture (per docs/fiber/diff_layer.md):
// Layer is just a grouping attribute, not a structural change.
// Modal nodes stay in the Fiber tree but need centering transform applied.
func (m *Manager) applyLayerTransforms(root *compute.ComputedBox, constraints runtime.BoxConstraints) {
	// Use RenderPlanes to find layer root nodes for centering
	// Modal centering: find the TOP-MOST modal node (first one added to LayerModal plane)
	// This node and all its children should be shifted together
	modalBoxes := m.renderPlanes.GetLayer(rtui.LayerModal)
	if len(modalBoxes) > 0 {
		// Find the modal root (the bordered box that contains the modal content)
		// It's the one that has no parent with LayerModal
		for _, box := range modalBoxes {
			if m.isLayerRoot(box, rtui.LayerModal) {
				m.centerModalBox(box, constraints)
				break // Only center the first modal root
			}
		}
	}

	// Inspector positioning
	inspectorBoxes := m.renderPlanes.GetLayer(rtui.LayerInspector)
	if len(inspectorBoxes) > 0 {
		for _, box := range inspectorBoxes {
			if m.isLayerRoot(box, rtui.LayerInspector) {
				// Position at specified coordinates (from props)
				// TODO: Implement inspector positioning
				break
			}
		}
	}
}

// isLayerRoot checks if a box is the root of its layer.
// A layer root is a box whose parent is either nil or in a different layer.
func (m *Manager) isLayerRoot(box *compute.ComputedBox, layer rtui.Layer) bool {
	if box == nil {
		return false
	}
	if box.Parent == nil {
		return true
	}
	// Parent exists but is in a different layer
	return box.Parent.Layer != layer
}

// transformBoxesByLayer recursively applies transforms based on Layer type.
func (m *Manager) transformBoxesByLayer(box *compute.ComputedBox, constraints runtime.BoxConstraints, offsetX, offsetY int) {
	if box == nil {
		return
	}

	box.Box.X += offsetX
	box.Box.Y += offsetY

	if box.Layer == rtui.LayerModal && box.Parent == nil {
		m.centerModalBox(box, constraints)
	}

	for _, child := range box.Children {
		m.transformBoxesByLayer(child, constraints, offsetX, offsetY)
	}
}

// centerModalBox centers a modal box in the viewport.
func (m *Manager) centerModalBox(box *compute.ComputedBox, constraints runtime.BoxConstraints) {
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

	log.LayerLogger.Debug("[centerModalBox] modal=%dx%d container=%dx%d offset=(%d,%d)",
		modalWidth, modalHeight, containerWidth, containerHeight, offsetX, offsetY)

	m.shiftBoxTree(box, offsetX, offsetY)
}

// shiftBoxTree shifts all boxes in a ComputedBox tree by the given offset.
func (m *Manager) shiftBoxTree(box *compute.ComputedBox, offsetX, offsetY int) {
	if box == nil {
		return
	}

	box.Box.X += offsetX
	box.Box.Y += offsetY

	if box.VNode != nil {
		if boundsAware, ok := box.VNode.(interface{ SetBounds(int, int, int, int) }); ok {
			boundsAware.SetBounds(box.Box.X, box.Box.Y, box.Box.Width, box.Box.Height)
		}
	}

	for _, child := range box.Children {
		m.shiftBoxTree(child, offsetX, offsetY)
	}
}

// buildMergedHitMapFromRenderPlanes builds HitMap from RenderPlanes.
func (m *Manager) buildMergedHitMapFromRenderPlanes() *event.HitMap {
	var entries []event.HitMapEntryInternal

	renderOrder := []rtui.Layer{
		rtui.LayerBase,
		rtui.LayerOverlay,
		rtui.LayerModal,
		rtui.LayerTooltip,
		rtui.LayerInspector,
	}

	zOrder := 0
	for _, layer := range renderOrder {
		boxes := m.renderPlanes.GetLayer(layer)
		for _, box := range boxes {
			if box.NodeID == 0 {
				continue
			}
			entries = append(entries, event.HitMapEntryInternal{
				NodeID:  box.NodeID,
				Node:    rtui.AsLayoutNode(box.VNode),
				Bounds:  runtimelayout.Rect{X: box.Box.X, Y: box.Box.Y, Width: box.Box.Width, Height: box.Box.Height},
				LocalXY: func(screenX, screenY int) (int, int) { return screenX - box.Box.X, screenY - box.Box.Y },
				ZOrder:  zOrder,
			})
		}
		if len(boxes) > 0 {
			zOrder++
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].ZOrder < entries[j].ZOrder
	})

	log.LayerLogger.Debug("[buildMergedHitMapFromRenderPlanes] Built HitMap with %d entries", len(entries))
	return event.BuildHitMapFromEntries(entries)
}

// BuildRenderPlane creates RenderPlanes from a single ComputedBox tree.
func BuildRenderPlane(root *compute.ComputedBox) *RenderPlanes {
	rp := NewRenderPlanes()

	var walk func(box *compute.ComputedBox)
	walk = func(box *compute.ComputedBox) {
		if box == nil {
			return
		}
		rp.AddToLayer(box.Layer, box)
		for _, child := range box.Children {
			walk(child)
		}
	}

	walk(root)
	return rp
}

// layoutLayer performs layout for a single layer node
// Phase 8: Added fiber parameter for NodeID propagation
func (m *Manager) layoutLayer(
	node *LayerNode,
	layer rtui.Layer,
	constraints runtime.BoxConstraints,
	engine *compute.Engine,
	fiber *reconciler.Fiber,
) (*compute.ComputedLayout, error) {
	log.LayerLogger.Debug("[layoutLayer] Layer=%d, constraints.Max=%dx%d",
		layer, constraints.MaxWidth, constraints.MaxHeight)
	log.HitMapLogger.Debug("[layoutLayer] Layer=%d, constraints.Max=%dx%d",
		layer, constraints.MaxWidth, constraints.MaxHeight)

	var layerConstraints runtime.BoxConstraints

	// Different layers have different constraints
	switch layer {
	case rtui.LayerModal:
		// Modals use full-screen constraints and will be centered
		layerConstraints = runtime.BoxConstraints{
			MinWidth:  0,
			MaxWidth:  constraints.MaxWidth,
			MinHeight: 0,
			MaxHeight: constraints.MaxHeight,
		}

	case rtui.LayerOverlay:
		// Overlays (dropdowns, etc.) position themselves
		layerConstraints = runtime.BoxConstraints{
			MinWidth:  0,
			MaxWidth:  constraints.MaxWidth,
			MinHeight: 0,
			MaxHeight: constraints.MaxHeight,
		}

	case rtui.LayerTooltip:
		// Tooltips are small and self-positioning
		layerConstraints = runtime.BoxConstraints{
			MinWidth:  0,
			MaxWidth:  constraints.MaxWidth,
			MinHeight: 0,
			MaxHeight: constraints.MaxHeight,
		}

	case rtui.LayerInspector:
		// Inspector overlay positions itself (typically in a corner)
		layerConstraints = runtime.BoxConstraints{
			MinWidth:  0,
			MaxWidth:  constraints.MaxWidth,
			MinHeight: 0,
			MaxHeight: constraints.MaxHeight,
		}

	default:
		layerConstraints = constraints
	}

	// Perform layout
	// Phase 8: Pass Fiber for NodeID propagation to modal/overlay/tooltip layers
	layout, err := engine.Layout(node.Content, fiber, layerConstraints)
	if err != nil {
		return nil, err
	}

	// Post-process layout for modal (center it)
	if layer == rtui.LayerModal && layout.Root != nil {
		m.centerModal(layout.Root, constraints)
	}

	// Post-process layout for inspector (position it at specified coordinates)
	if layer == rtui.LayerInspector && layout.Root != nil {
		m.positionInspector(node, layout.Root)
	}

	// IMPORTANT: Rebuild HitMap AFTER post-processing (centering, etc.)
	// The HitMap built in Engine.Layout() was before centering, so it has wrong positions
	// We need to rebuild it now with the final transformed positions
	if layout.Root != nil {
		layout.HitMap = m.buildHitMapFromComputedBox(layout.Root)

		// DEBUG: Output modal position after centering
		if log.HitMapLogger.Enabled() {
			log.RenderLogger.Debug("[layoutLayer] Layer=%d, Root pos=(%d,%d) size=%dx%d, HitMap entries=%d",
				layer, layout.Root.Box.X, layout.Root.Box.Y, layout.Root.Box.Width, layout.Root.Box.Height,
				layout.HitMap.Size())
		}
	}

	return layout, nil
}

// centerModal adjusts the modal position to be centered in the viewport
func (m *Manager) centerModal(root *compute.ComputedBox, constraints runtime.BoxConstraints) {
	if root == nil {
		return
	}

	// Calculate centering offset
	modalWidth := root.Box.Width
	modalHeight := root.Box.Height
	originalX := root.Box.X
	originalY := root.Box.Y

	containerWidth := constraints.MaxWidth
	containerHeight := constraints.MaxHeight

	// Handle infinite dimensions
	if containerWidth == runtime.Infinity {
		containerWidth = modalWidth
	}
	if containerHeight == runtime.Infinity {
		containerHeight = modalHeight
	}

	offsetX := (containerWidth - modalWidth) / 2
	offsetY := (containerHeight - modalHeight) / 2

	// Ensure non-negative offsets
	if offsetX < 0 {
		offsetX = 0
	}
	if offsetY < 0 {
		offsetY = 0
	}

	log.RenderLogger.Debug("[centerModal] modal=(%d,%d) size=%dx%d container=%dx%d offset=(%d,%d)",
		originalX, originalY, modalWidth, modalHeight, containerWidth, containerHeight, offsetX, offsetY)

	// Shift the entire layout tree
	m.shiftPositions(root, offsetX, offsetY)

	log.RenderLogger.Debug("[centerModal] after shift: modal=(%d,%d)",
		root.Box.X, root.Box.Y)
}

// shiftPositions shifts all boxes in a layout tree by the given offset
func (m *Manager) shiftPositions(box *compute.ComputedBox, offsetX, offsetY int) {
	if box == nil {
		return
	}

	box.Box.X += offsetX
	box.Box.Y += offsetY

	// Update VNode bounds after position shift
	// This ensures Button.bounds and other components have correct coordinates for hit testing
	if box.VNode != nil {
		if boundsAware, ok := box.VNode.(interface{ SetBounds(int, int, int, int) }); ok {
			boundsAware.SetBounds(box.Box.X, box.Box.Y, box.Box.Width, box.Box.Height)
		}
	}

	for _, child := range box.Children {
		m.shiftPositions(child, offsetX, offsetY)
	}
}

// positionInspector positions the inspector overlay at its specified coordinates
// Inspector overlays use "x" and "y" props to specify their position
func (m *Manager) positionInspector(node *LayerNode, root *compute.ComputedBox) {
	if root == nil {
		return
	}

	// Get the specified position from props
	var targetX, targetY int
	props := node.Content.Props()

	if x, ok := props["x"].(int); ok {
		targetX = x
	} else {
		targetX = 0 // Default to left edge
	}

	if y, ok := props["y"].(int); ok {
		targetY = y
	} else {
		targetY = 0 // Default to top edge
	}

	// Clamp negative coordinates to 0
	if targetX < 0 {
		targetX = 0
	}
	if targetY < 0 {
		targetY = 0
	}

	originalX := root.Box.X
	originalY := root.Box.Y

	log.RenderLogger.Debug("[positionInspector] original=(%d,%d) target=(%d,%d)\n",
		originalX, originalY, targetX, targetY)

	// Calculate offset
	offsetX := targetX - originalX
	offsetY := targetY - originalY

	// Shift the entire layout tree
	m.shiftPositions(root, offsetX, offsetY)

	log.RenderLogger.Debug("[positionInspector] after shift: inspector=(%d,%d) size=%dx%d\n",
		root.Box.X, root.Box.Y, root.Box.Width, root.Box.Height)

}

// =============================================================================
// Query Methods
// =============================================================================

// GetRenderPlanes returns the RenderPlanes for unified layer management
// Phase 3: Provides access to the new RenderPlanes-based layer system
func (m *Manager) GetRenderPlanes() *RenderPlanes {
	return m.renderPlanes
}

// GetLayouts returns all layer layouts
func (m *Manager) GetLayouts() LayerLayouts {
	return m.layouts
}

// GetLayout returns the layout for a specific layer
func (m *Manager) GetLayout(layer rtui.Layer) (*compute.ComputedLayout, bool) {
	layout, ok := m.layouts[layer]
	return layout, ok
}

// GetBaseLayout returns the base layer layout
func (m *Manager) GetBaseLayout() *compute.ComputedLayout {
	return m.layouts[rtui.LayerBase]
}

// HasModal returns true if there is a modal layer
func (m *Manager) HasModal() bool {
	// Phase 3: Prefer RenderPlanes-based query
	return m.renderPlanes.HasLayer(rtui.LayerModal)
}

// HasOverlay returns true if there is any overlay content
func (m *Manager) HasOverlay() bool {
	// Phase 3: Prefer RenderPlanes-based query
	return m.renderPlanes.HasLayer(rtui.LayerOverlay)
}

// GetHighestLayer returns the highest layer with content
func (m *Manager) GetHighestLayer() rtui.Layer {
	// Phase 3: Use RenderPlanes-based query
	return m.renderPlanes.GetHighestLayer()
}

// GetModalNodes returns all modal layer nodes
func (m *Manager) GetModalNodes() []*LayerNode {
	return m.collector.GetModalNodes()
}

// GetOverlayNodes returns all overlay layer nodes
func (m *Manager) GetOverlayNodes() []*LayerNode {
	return m.collector.GetOverlayNodes()
}

// GetTooltipNodes returns all tooltip layer nodes
func (m *Manager) GetTooltipNodes() []*LayerNode {
	return m.collector.GetTooltipNodes()
}

// GetMergedHitMap merges HitMaps from all layers into a single HitMap
// This combines hit test information from base, modal, overlay, tooltip, and inspector layers
// The merged HitMap respects layer Z-order (upper layers have higher Z-order)
func (m *Manager) GetMergedHitMap() *event.HitMap {
	var entries []event.HitMapEntryInternal

	// Render order: from lowest (base) to highest (inspector)
	renderOrder := []rtui.Layer{
		rtui.LayerBase,
		rtui.LayerOverlay,
		rtui.LayerModal,
		rtui.LayerTooltip,
		rtui.LayerInspector,
	}

	zOrder := 0
	for _, layer := range renderOrder {
		layout, ok := m.layouts[layer]
		if !ok || layout.HitMap == nil {
			if !ok {
				log.RenderLogger.Debug("[GetMergedHitMap] Layer %d: no layout", layer)
			} else {
				log.RenderLogger.Debug("[GetMergedHitMap] Layer %d: layout has nil HitMap", layer)
			}
			continue
		}

		log.RenderLogger.Debug("[GetMergedHitMap] Layer %d: HitMap has %d entries", layer, layout.HitMap.Size())

		// Append all entries from this layer's HitMap
		// Update their Z-order to reflect the layer hierarchy
		for _, entry := range layout.HitMap.AllEntries() {
			// Log modal button positions
			if layer == rtui.LayerModal {
				log.RenderLogger.Debug("[GetMergedHitMap] Modal entry: ID=%d, Bounds=(%d,%d,%dx%d)",
					entry.NodeID, entry.Bounds.X, entry.Bounds.Y, entry.Bounds.Width, entry.Bounds.Height)
			}

			// Create a new entry with updated Z-order
			newEntry := event.HitMapEntryInternal{
				NodeID:  entry.NodeID,
				Node:    entry.Node,
				Bounds:  entry.Bounds,
				LocalXY: entry.LocalXY,
				ZOrder:  zOrder,
			}
			entries = append(entries, newEntry)
		}

		zOrder++
	}

	// Sort by Z-order (ascending)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].ZOrder < entries[j].ZOrder
	})

	log.RenderLogger.Debug("[GetMergedHitMap] Merged HitMap: %d entries from %d layers",
		len(entries), len(m.layouts))

	// Build HitMap from entries
	return event.BuildHitMapFromEntries(entries)
}

// buildHitMapFromComputedBox builds a HitMap from a ComputedBox tree
// This is called AFTER layer transforms (centering, etc.) to capture the final positions
func (m *Manager) buildHitMapFromComputedBox(root *compute.ComputedBox) *event.HitMap {
	if root == nil {
		return event.NewHitMap()
	}

	var entries []event.HitMapEntryInternal

	// Recursively walk the ComputedBox tree
	var walk func(box *compute.ComputedBox, zOrder int)
	walk = func(box *compute.ComputedBox, zOrder int) {
		if box == nil {
			return
		}

		// Skip nodes with zero size
		if box.Box.Width <= 0 || box.Box.Height <= 0 {
			for _, child := range box.Children {
				walk(child, zOrder+1)
			}
			return
		}

		// Get NodeID from ComputedBox (now has uint64 NodeID field)
		// Phase 3: Use box.NodeID directly for stable identity
		// Fallback to converting string key to uint64 for compatibility during transition
		nodeID := box.NodeID
		if nodeID == 0 && box.VNode != nil {
			// Convert VNode key to NodeID using hash for compatibility
			// This maintains backward compatibility with non-Fiber mode
			if key := box.VNode.Key(); key != "" {
				nodeID = event.StringToNodeID(key)
			}
		}

		// Create entry with FINAL positions (after layer transforms)
		entry := event.HitMapEntryInternal{
			NodeID: nodeID,
			Node:   rtui.AsLayoutNode(box.VNode),
			Bounds: runtimelayout.Rect{
				X:      box.Box.X, // ✅ Final position AFTER centering
				Y:      box.Box.Y,
				Width:  box.Box.Width,
				Height: box.Box.Height,
			},
			LocalXY: func(screenX, screenY int) (int, int) {
				return screenX - box.Box.X, screenY - box.Box.Y
			},
			ZOrder: zOrder,
		}

		entries = append(entries, entry)

		// Log entry positions for debugging
		if entry.NodeID != 0 {
			log.RenderLogger.Debug("[buildHitMapFromComputedBox] Entry: ID=%d, Bounds=(%d,%d,%dx%d)",
				entry.NodeID, entry.Bounds.X, entry.Bounds.Y, entry.Bounds.Width, entry.Bounds.Height)
		}

		// Recursively process children
		for _, child := range box.Children {
			walk(child, zOrder+1)
		}
	}

	walk(root, 0)

	// Sort by Z-order
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].ZOrder < entries[j].ZOrder
	})

	log.RenderLogger.Debug("[LayerManager] Built HitMap: %d entries", len(entries))

	return event.BuildHitMapFromEntries(entries)
}

// GetInspectorNodes returns all inspector layer nodes
func (m *Manager) GetInspectorNodes() []*LayerNode {
	return m.collector.GetInspectorNodes()
}

// HasInspector returns true if there is an inspector layer
func (m *Manager) HasInspector() bool {
	return m.collector.HasInspector()
}

// =============================================================================
// Layer Ordering
// =============================================================================

// RenderOrder returns the layers in render order (lowest to highest)
func (m *Manager) RenderOrder() []rtui.Layer {
	var layers []rtui.Layer

	// Always include base layer
	if _, ok := m.layouts[rtui.LayerBase]; ok {
		layers = append(layers, rtui.LayerBase)
	}

	// Add overlay
	if _, ok := m.layouts[rtui.LayerOverlay]; ok {
		layers = append(layers, rtui.LayerOverlay)
	}

	// Add modal
	if _, ok := m.layouts[rtui.LayerModal]; ok {
		layers = append(layers, rtui.LayerModal)
	}

	// Add tooltip
	if _, ok := m.layouts[rtui.LayerTooltip]; ok {
		layers = append(layers, rtui.LayerTooltip)
	}

	// Add inspector (highest layer)
	if _, ok := m.layouts[rtui.LayerInspector]; ok {
		layers = append(layers, rtui.LayerInspector)
	}

	return layers
}

// =============================================================================
// Event Handling Support
// =============================================================================

// ShouldBlockEvent returns true if events should be blocked at the given position
// This is used to prevent clicks on background content when a modal is open
func (m *Manager) ShouldBlockEvent(x, y int) bool {
	// If there's a modal, it blocks all background events
	if m.HasModal() {
		// Check if the click is within the modal bounds
		if modalLayout, ok := m.layouts[rtui.LayerModal]; ok && modalLayout.Root != nil {
			box := modalLayout.Root.Box
			// Click outside modal bounds should be blocked
			// (or could be used to close the modal)
			return x < box.X || x >= box.X+box.Width ||
				y < box.Y || y >= box.Y+box.Height
		}
		return true
	}

	// Other layers don't block events
	return false
}

// GetNodeAtPosition returns the layer node at the given position
// This is used for event dispatching
func (m *Manager) GetNodeAtPosition(x, y int) (*LayerNode, rtui.Layer) {
	// Check from highest to lowest layer
	for l := rtui.LayerTooltip; l >= rtui.LayerBase; l-- {
		layout, ok := m.layouts[l]
		if !ok || layout.Root == nil {
			continue
		}

		if m.containsPosition(layout.Root, x, y) {
			// Find the layer node for this layer
			nodes := m.collector.GetLayerNodes(l)
			if len(nodes) > 0 {
				return nodes[0], l
			}
		}
	}

	return nil, rtui.LayerBase
}

// containsPosition checks if a box tree contains the given position
func (m *Manager) containsPosition(box *compute.ComputedBox, x, y int) bool {
	if box == nil {
		return false
	}

	if x >= box.Box.X && x < box.Box.X+box.Box.Width &&
		y >= box.Box.Y && y < box.Box.Y+box.Box.Height {
		return true
	}

	for _, child := range box.Children {
		if m.containsPosition(child, x, y) {
			return true
		}
	}

	return false
}
