package layer

import (
	"testing"

	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/compute"
	"github.com/wwsheng009/mint/runtime/paint"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Convert Paintable Tests
// =============================================================================

func TestToPaintRenderLayer(t *testing.T) {
	tests := []struct {
		uiLayer  rtui.Layer
		expected paint.RenderLayer
	}{
		{rtui.LayerBase, paint.RenderLayerBase},
		{rtui.LayerOverlay, paint.RenderLayerOverlay},
		{rtui.LayerModal, paint.RenderLayerModal},
		{rtui.LayerTooltip, paint.RenderLayerTooltip},
		{rtui.LayerInspector, paint.RenderLayerInspector},
	}

	for _, tt := range tests {
		t.Run(tt.uiLayer.String(), func(t *testing.T) {
			got := toPaintRenderLayer(tt.uiLayer)
			if got != tt.expected {
				t.Errorf("toPaintRenderLayer(%v) = %v, want %v", tt.uiLayer, got, tt.expected)
			}
		})
	}
}

func TestConvertToPaintableLayouts(t *testing.T) {
	// 创建 LayerLayouts
	layouts := make(LayerLayouts)

	// 创建一个简单的 ComputedLayout
	rootBox := &compute.ComputedBox{
		Box:     runtime.Box{X: 0, Y: 0, Width: 80, Height: 25},
		NodeID:  1,
		Layer:   rtui.LayerBase,
	}
	layout := &compute.ComputedLayout{Root: rootBox}
	layouts[rtui.LayerBase] = layout

	// 转换
	paintableLayouts := ConvertToPaintableLayouts(layouts)

	if paintableLayouts == nil {
		t.Fatal("ConvertToPaintableLayouts returned nil")
	}

	if paintableLayouts[paint.RenderLayerBase] == nil {
		t.Error("Expected Base layer in PaintableLayouts")
	}
}

func TestConvertToPaintablePlanes(t *testing.T) {
	// 创建 RenderPlanes
	rp := NewRenderPlanes()

	box1 := &compute.ComputedBox{
		Box:    runtime.Box{X: 0, Y: 0, Width: 80, Height: 25},
		NodeID: 1,
		Layer:  rtui.LayerBase,
	}
	box2 := &compute.ComputedBox{
		Box:    runtime.Box{X: 20, Y: 10, Width: 40, Height: 5},
		NodeID: 2,
		Layer:  rtui.LayerModal,
	}

	rp.AddToLayer(rtui.LayerBase, box1)
	rp.AddToLayer(rtui.LayerModal, box2)

	// 转换
	pp := ConvertToPaintablePlanes(rp)

	if pp == nil {
		t.Fatal("ConvertToPaintablePlanes returned nil")
	}

	if pp.CountBoxes() != 2 {
		t.Errorf("PaintablePlanes.CountBoxes() = %d, want 2", pp.CountBoxes())
	}

	if pp.IsLayerEmpty(paint.RenderLayerBase) {
		t.Error("Base layer should not be empty")
	}
	if pp.IsLayerEmpty(paint.RenderLayerModal) {
		t.Error("Modal layer should not be empty")
	}
}

func TestConvertToPaintablePlanes_Nil(t *testing.T) {
	pp := ConvertToPaintablePlanes(nil)

	if pp == nil {
		t.Fatal("ConvertToPaintablePlanes(nil) returned nil")
	}

	if pp.CountBoxes() != 0 {
		t.Errorf("ConvertToPaintablePlanes(nil) should have 0 boxes, got %d", pp.CountBoxes())
	}
}

func TestRenderPlanes_AsPaintablePlanes(t *testing.T) {
	rp := NewRenderPlanes()
	rp.AddToLayer(rtui.LayerBase, &compute.ComputedBox{
		Box:    runtime.Box{X: 0, Y: 0, Width: 10, Height: 5},
		NodeID: 1,
		Layer:  rtui.LayerBase,
	})

	pp := rp.AsPaintablePlanes()

	if pp == nil {
		t.Fatal("AsPaintablePlanes() returned nil")
	}

	if pp.CountBoxes() != 1 {
		t.Errorf("PaintablePlanes.CountBoxes() = %d, want 1", pp.CountBoxes())
	}
}

func TestLayerLayouts_AsPaintableLayouts(t *testing.T) {
	layouts := make(LayerLayouts)
	layouts[rtui.LayerBase] = &compute.ComputedLayout{
		Root: &compute.ComputedBox{
			Box:    runtime.Box{X: 0, Y: 0, Width: 10, Height: 5},
			NodeID: 1,
			Layer:  rtui.LayerBase,
		},
	}

	paintableLayouts := layouts.AsPaintableLayouts()

	if paintableLayouts == nil {
		t.Fatal("AsPaintableLayouts() returned nil")
	}

	if paintableLayouts[paint.RenderLayerBase] == nil {
		t.Error("Expected Base layer in PaintableLayouts")
	}
}

func TestBuildPaintablePlanesFromComputedBox(t *testing.T) {
	root := &compute.ComputedBox{
		Box:    runtime.Box{X: 0, Y: 0, Width: 80, Height: 25},
		NodeID: 1,
		Layer:  rtui.LayerBase,
		Children: []*compute.ComputedBox{
			{
				Box:    runtime.Box{X: 0, Y: 0, Width: 40, Height: 25},
				NodeID: 2,
				Layer:  rtui.LayerBase,
			},
			{
				Box:    runtime.Box{X: 40, Y: 0, Width: 40, Height: 25},
				NodeID: 3,
				Layer:  rtui.LayerModal,
			},
		},
	}

	pp := BuildPaintablePlanesFromComputedBox(root)

	if pp == nil {
		t.Fatal("BuildPaintablePlanesFromComputedBox returned nil")
	}

	if pp.CountBoxes() != 3 {
		t.Errorf("PaintablePlanes.CountBoxes() = %d, want 3", pp.CountBoxes())
	}

	baseBoxes := pp.GetLayer(paint.RenderLayerBase)
	if len(baseBoxes) != 2 {
		t.Errorf("Base layer has %d boxes, want 2", len(baseBoxes))
	}

	modalBoxes := pp.GetLayer(paint.RenderLayerModal)
	if len(modalBoxes) != 1 {
		t.Errorf("Modal layer has %d boxes, want 1", len(modalBoxes))
	}
}

func TestBuildPaintablePlanesFromComputedBox_Nil(t *testing.T) {
	pp := BuildPaintablePlanesFromComputedBox(nil)

	if pp == nil {
		t.Fatal("BuildPaintablePlanesFromComputedBox(nil) returned nil")
	}

	if pp.CountBoxes() != 0 {
		t.Errorf("BuildPaintablePlanesFromComputedBox(nil) should have 0 boxes, got %d", pp.CountBoxes())
	}
}

// =============================================================================
// Manager.GetPaintablePlanes Tests
// =============================================================================

func TestManager_GetPaintablePlanes(t *testing.T) {
	// Test with nil RenderPlanes
	m := &Manager{renderPlanes: nil}
	pp := m.GetPaintablePlanes()

	if pp == nil {
		t.Fatal("GetPaintablePlanes() returned nil for nil RenderPlanes")
	}

	if pp.CountBoxes() != 0 {
		t.Errorf("GetPaintablePlanes() should have 0 boxes for nil RenderPlanes, got %d", pp.CountBoxes())
	}
}

func TestManager_GetPaintablePlanes_WithData(t *testing.T) {
	m := NewManager()
	rp := m.GetRenderPlanes()

	// Add some boxes
	rp.AddToLayer(rtui.LayerBase, &compute.ComputedBox{
		Box:    runtime.Box{X: 0, Y: 0, Width: 10, Height: 5},
		NodeID: 1,
		Layer:  rtui.LayerBase,
	})
	rp.AddToLayer(rtui.LayerModal, &compute.ComputedBox{
		Box:    runtime.Box{X: 5, Y: 5, Width: 20, Height: 10},
		NodeID: 2,
		Layer:  rtui.LayerModal,
	})

	pp := m.GetPaintablePlanes()

	if pp == nil {
		t.Fatal("GetPaintablePlanes() returned nil")
	}

	if pp.CountBoxes() != 2 {
		t.Errorf("PaintablePlanes.CountBoxes() = %d, want 2", pp.CountBoxes())
	}

	baseBoxes := pp.GetLayer(paint.RenderLayerBase)
	if len(baseBoxes) != 1 {
		t.Errorf("Base layer has %d boxes, want 1", len(baseBoxes))
	}

	modalBoxes := pp.GetLayer(paint.RenderLayerModal)
	if len(modalBoxes) != 1 {
		t.Errorf("Modal layer has %d boxes, want 1", len(modalBoxes))
	}
}
