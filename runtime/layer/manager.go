// Package layer provides layer management for multi-layer TUI rendering
package layer

import (
	"sort"

	"github.com/wwsheng009/mint/internal/log"
	"github.com/wwsheng009/mint/runtime/compute"
	"github.com/wwsheng009/mint/runtime/event"
	"github.com/wwsheng009/mint/runtime/paint"
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
// Query Methods
// =============================================================================

// GetRenderPlanes returns the RenderPlanes for unified layer management
// Phase 3: Provides access to the new RenderPlanes-based layer system
func (m *Manager) GetRenderPlanes() *RenderPlanes {
	return m.renderPlanes
}

// GetPaintablePlanes returns the PaintablePlanes for decoupled rendering.
// This is the preferred API for PaintEngine.PaintPaintablePlanes().
func (m *Manager) GetPaintablePlanes() *paint.PaintablePlanes {
	if m.renderPlanes == nil {
		return paint.NewPaintablePlanes()
	}
	return m.renderPlanes.AsPaintablePlanes()
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