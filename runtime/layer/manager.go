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
	collector *Collector
	layouts   LayerLayouts
}

// LayerLayouts maps each Layer to its computed layout
type LayerLayouts map[rtui.Layer]*compute.ComputedLayout

// NewManager creates a new layer manager
func NewManager() *Manager {
	return &Manager{
		collector: NewCollector(),
		layouts:   make(LayerLayouts),
	}
}

// =============================================================================
// Main API
// =============================================================================

// CollectAndLayout collects layer nodes and performs layout for all layers
// This is the main entry point for layer-based rendering
//
// Phase 8: Added optional fiber parameter for NodeID propagation
func (m *Manager) CollectAndLayout(
	vnode rtui.VNode,
	fiber *reconciler.Fiber,
	constraints runtime.BoxConstraints,
	engine *compute.Engine,
) error {
	// Clear previous state
	m.layouts = make(LayerLayouts)

	// 1. Collect layer nodes from the VNode tree
	m.collector.Collect(vnode)

	log.LayerLogger.Debug("[CollectAndLayout] collected %d modal nodes", len(m.collector.GetModalNodes()))

	// 2. Strip layer nodes from the main tree to get clean base content
	baseTree := m.collector.StripLayers(vnode)

	baseChildren := baseTree.Children()
	log.LayerLogger.Debug("[CollectAndLayout] baseTree has %d children (after stripping)", len(baseChildren))
	for i, child := range baseChildren {
		log.LayerLogger.Debug("[CollectAndLayout]   child %d: layer=%d type=%s", i, child.GetLayer(), child.Type().String())
	}

	// 3. Layout the base layer
	// Phase 8: Pass Fiber to layout engine for NodeID propagation
	baseLayout, err := engine.Layout(baseTree, fiber, constraints)
	if err != nil {
		return err
	}
	m.layouts[rtui.LayerBase] = baseLayout

	// 4. Layout each collected layer
	for layer, nodes := range m.collector.GetLayers() {
		if len(nodes) == 0 {
			continue
		}

		// Layout nodes for this layer
		// For now, we only support the first visible node per layer
		for _, node := range nodes {
			if !node.Visible {
				continue
			}

			// Phase 8: Pass fiber to layoutLayer for NodeID propagation
			layerLayout, err := m.layoutLayer(node, layer, constraints, engine, fiber)
			if err != nil {
				return err
			}
			m.layouts[layer] = layerLayout

			// Only layout the first visible node per layer
			break
		}
	}

	return nil
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
	return m.collector.HasModal()
}

// HasOverlay returns true if there is any overlay content
func (m *Manager) HasOverlay() bool {
	return m.collector.HasOverlay()
}

// GetHighestLayer returns the highest layer with content
func (m *Manager) GetHighestLayer() rtui.Layer {
	return m.collector.GetHighestLayer()
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
