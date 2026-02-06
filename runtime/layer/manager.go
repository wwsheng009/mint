// Package layer provides layer management for multi-layer TUI rendering
package layer

import (
	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/compute"
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
func (m *Manager) CollectAndLayout(
	vnode rtui.VNode,
	constraints runtime.BoxConstraints,
	engine *compute.Engine,
) error {
	// Clear previous state
	m.layouts = make(LayerLayouts)

	// 1. Collect layer nodes from the VNode tree
	m.collector.Collect(vnode)

	// 2. Strip layer nodes from the main tree to get clean base content
	baseTree := m.collector.StripLayers(vnode)

	// 3. Layout the base layer
	baseLayout, err := engine.Layout(baseTree, constraints)
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

			layerLayout, err := m.layoutLayer(node, layer, constraints, engine)
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
func (m *Manager) layoutLayer(
	node *LayerNode,
	layer rtui.Layer,
	constraints runtime.BoxConstraints,
	engine *compute.Engine,
) (*compute.ComputedLayout, error) {
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

	default:
		layerConstraints = constraints
	}

	// Perform layout
	layout, err := engine.Layout(node.Content, layerConstraints)
	if err != nil {
		return nil, err
	}

	// Post-process layout for modal (center it)
	if layer == rtui.LayerModal && layout.Root != nil {
		m.centerModal(layout.Root, constraints)
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

	// Shift the entire layout tree
	m.shiftPositions(root, offsetX, offsetY)
}

// shiftPositions shifts all boxes in a layout tree by the given offset
func (m *Manager) shiftPositions(box *compute.ComputedBox, offsetX, offsetY int) {
	if box == nil {
		return
	}

	box.Box.X += offsetX
	box.Box.Y += offsetY

	for _, child := range box.Children {
		m.shiftPositions(child, offsetX, offsetY)
	}
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
