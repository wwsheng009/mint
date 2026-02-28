package layer

import (
	"testing"

	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/compute"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// TestLayerManagerConstruction tests creating a new LayerManager
func TestLayerManagerConstruction(t *testing.T) {
	m := NewManager()

	if m == nil {
		t.Fatal("NewManager returned nil")
	}

	// Verify renderPlanes is initialized
	if m.renderPlanes == nil {
		t.Error("renderPlanes should not be nil")
	}

	// Verify layouts is initialized
	if m.layouts == nil {
		t.Error("layouts should not be nil")
	}

	// Verify collector is initialized
	if m.collector == nil {
		t.Error("collector should not be nil")
	}
}

// TestLayerManagerGetRenderPlanes tests getting RenderPlanes
func TestLayerManagerGetRenderPlanes(t *testing.T) {
	m := NewManager()

	rp := m.GetRenderPlanes()

	if rp == nil {
		t.Error("GetRenderPlanes returned nil")
	}

	// Verify it's the same instance each call
	rp2 := m.GetRenderPlanes()
	if rp != rp2 {
		t.Error("GetRenderPlanes should return the same instance")
	}
}

// TestLayerManagerQueries tests query methods using RenderPlanes
func TestLayerManagerQueries(t *testing.T) {
	m := NewManager()

	// Initially, all layers should be empty
	if m.HasModal() {
		t.Error("HasModal should return false initially")
	}

	if m.HasOverlay() {
		t.Error("HasOverlay should return false initially")
	}

	if m.GetHighestLayer() != rtui.LayerBase {
		t.Errorf("GetHighestLayer should return LayerBase when empty, got %d", m.GetHighestLayer())
	}

	// Add a modal box to RenderPlanes
	modalBox := &compute.ComputedBox{
		Box:    runtime.Box{X: 0, Y: 0, Width: 10, Height: 10},
		Layer:  rtui.LayerModal,
		NodeID: 1,
	}
	m.renderPlanes.AddToLayer(rtui.LayerModal, modalBox)

	// Now HasModal should return true
	if !m.HasModal() {
		t.Error("HasModal should return true after adding modal")
	}

	// GetHighestLayer should return LayerModal
	highest := m.GetHighestLayer()
	if highest != rtui.LayerModal {
		t.Errorf("GetHighestLayer should return LayerModal, got %d", highest)
	}

	// Add an overlay box
	overlayBox := &compute.ComputedBox{
		Box:    runtime.Box{X: 5, Y: 5, Width: 15, Height: 15},
		Layer:  rtui.LayerOverlay,
		NodeID: 2,
	}
	m.renderPlanes.AddToLayer(rtui.LayerOverlay, overlayBox)

	// HasOverlay should return true
	if !m.HasOverlay() {
		t.Error("HasOverlay should return true after adding overlay")
	}

	// GetHighestLayer should still be LayerModal (higher than Overlay)
	highest = m.GetHighestLayer()
	if highest != rtui.LayerModal {
		t.Errorf("GetHighestLayer should still be LayerModal, got %d", highest)
	}

	// Add a tooltip box (higher than Modal)
	tooltipBox := &compute.ComputedBox{
		Box:    runtime.Box{X: 2, Y: 2, Width: 8, Height: 8},
		Layer:  rtui.LayerTooltip,
		NodeID: 3,
	}
	m.renderPlanes.AddToLayer(rtui.LayerTooltip, tooltipBox)

	// GetHighestLayer should now be LayerTooltip
	highest = m.GetHighestLayer()
	if highest != rtui.LayerTooltip {
		t.Errorf("GetHighestLayer should return LayerTooltip, got %d", highest)
	}
}

// TestLayerManagerGetLayouts tests getting layer layouts
func TestLayerManagerGetLayouts(t *testing.T) {
	m := NewManager()

	layouts := m.GetLayouts()

	if layouts == nil {
		t.Error("GetLayouts returned nil")
	}

	// Initially should have no layouts
	if len(layouts) != 0 {
		t.Errorf("GetLayouts should return empty map initially, got %d entries", len(layouts))
	}

	// Add a layout
	m.layouts[rtui.LayerBase] = &compute.ComputedLayout{
		Root: &compute.ComputedBox{
			Box:    runtime.Box{X: 0, Y: 0, Width: 80, Height: 24},
			Layer:  rtui.LayerBase,
			NodeID: 1,
		},
	}

	layouts = m.GetLayouts()
	if len(layouts) != 1 {
		t.Errorf("GetLayouts should return 1 layout, got %d", len(layouts))
	}

	if _, ok := layouts[rtui.LayerBase]; !ok {
		t.Error("GetLayouts should contain LayerBase")
	}
}


// TestLayerManagerGetBaseLayout tests getting base layer layout
func TestLayerManagerGetBaseLayout(t *testing.T) {
	m := NewManager()

	// Initially nil
	if m.GetBaseLayout() != nil {
		t.Error("GetBaseLayout should return nil initially")
	}

	// Add base layout
	baseLayout := &compute.ComputedLayout{
		Root: &compute.ComputedBox{
			Box:    runtime.Box{X: 0, Y: 0, Width: 80, Height: 24},
			Layer:  rtui.LayerBase,
			NodeID: 1,
		},
	}
	m.layouts[rtui.LayerBase] = baseLayout

	// Should return the base layout
	if m.GetBaseLayout() != baseLayout {
		t.Error("GetBaseLayout should return the correct layout")
	}
}

// TestLayerManagerRenderOrder tests render order
func TestLayerManagerRenderOrder(t *testing.T) {
	m := NewManager()

	// Empty manager should have no layers
	order := m.RenderOrder()
	if len(order) != 0 {
		t.Errorf("RenderOrder should be empty initially, got %d layers", len(order))
	}

	// Add base layer
	m.layouts[rtui.LayerBase] = &compute.ComputedLayout{}

	order = m.RenderOrder()
	if len(order) != 1 {
		t.Errorf("RenderOrder should have 1 layer, got %d", len(order))
	}
	if order[0] != rtui.LayerBase {
		t.Errorf("RenderOrder[0] should be LayerBase, got %d", order[0])
	}

	// Add modal layer (should be after base)
	m.layouts[rtui.LayerModal] = &compute.ComputedLayout{}

	order = m.RenderOrder()
	if len(order) != 2 {
		t.Errorf("RenderOrder should have 2 layers, got %d", len(order))
	}
	if order[0] != rtui.LayerBase {
		t.Errorf("RenderOrder[0] should be LayerBase, got %d", order[0])
	}
	if order[1] != rtui.LayerModal {
		t.Errorf("RenderOrder[1] should be LayerModal, got %d", order[1])
	}

	// Add inspector (highest)
	m.layouts[rtui.LayerInspector] = &compute.ComputedLayout{}

	order = m.RenderOrder()
	if len(order) != 3 {
		t.Errorf("RenderOrder should have 3 layers, got %d", len(order))
	}
	// Inspector should be last (highest)
	if order[len(order)-1] != rtui.LayerInspector {
		t.Errorf("RenderOrder[%d] should be LayerInspector, got %d", len(order)-1, order[len(order)-1])
	}
}

// TestLayerManagerShouldBlockEvent tests event blocking
func TestLayerManagerShouldBlockEvent(t *testing.T) {
	m := NewManager()

	// Without modal, events should not be blocked
	if m.ShouldBlockEvent(10, 10) {
		t.Error("ShouldBlockEvent should return false without modal")
	}

	// Add modal at position (5, 5) with size 20x20
	modalLayout := &compute.ComputedLayout{
		Root: &compute.ComputedBox{
			Box: runtime.Box{
				X:      5,
				Y:      5,
				Width:  20,
				Height: 20,
			},
			Layer:  rtui.LayerModal,
			NodeID: 1,
		},
	}
	m.layouts[rtui.LayerModal] = modalLayout

	// Add modal to RenderPlanes so HasModal returns true
	m.renderPlanes.AddToLayer(rtui.LayerModal, modalLayout.Root)

	// Click inside modal should not be blocked
	if m.ShouldBlockEvent(10, 10) {
		t.Error("Click inside modal should not be blocked")
	}

	// Click outside modal should be blocked
	if !m.ShouldBlockEvent(0, 0) {
		t.Error("Click outside modal should be blocked")
	}
	if !m.ShouldBlockEvent(30, 30) {
		t.Error("Click far from modal should be blocked")
	}

	// Click on modal edge
	if m.ShouldBlockEvent(5, 10) {
		t.Error("Click on modal left edge should not be blocked")
	}
	if m.ShouldBlockEvent(24, 10) {
		t.Error("Click on modal right edge should not be blocked")
	}
}
