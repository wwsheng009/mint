package paint

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/types"
)

// =============================================================================
// RenderLayer Tests
// =============================================================================

func TestRenderLayer_String(t *testing.T) {
	tests := []struct {
		layer    RenderLayer
		expected string
	}{
		{RenderLayerBase, "base"},
		{RenderLayerOverlay, "overlay"},
		{RenderLayerModal, "modal"},
		{RenderLayerTooltip, "tooltip"},
		{RenderLayerInspector, "inspector"},
		{RenderLayer(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.layer.String(); got != tt.expected {
				t.Errorf("RenderLayer(%d).String() = %q, want %q", tt.layer, got, tt.expected)
			}
		})
	}
}

func TestRenderLayer_Values(t *testing.T) {
	// 验证 RenderLayer 值与 types.Layer 常量一致
	if RenderLayerBase != types.LayerBase {
		t.Errorf("RenderLayerBase = %d, want %d", RenderLayerBase, types.LayerBase)
	}
	if RenderLayerOverlay != types.LayerOverlay {
		t.Errorf("RenderLayerOverlay = %d, want %d", RenderLayerOverlay, types.LayerOverlay)
	}
	if RenderLayerModal != types.LayerModal {
		t.Errorf("RenderLayerModal = %d, want %d", RenderLayerModal, types.LayerModal)
	}
	if RenderLayerTooltip != types.LayerTooltip {
		t.Errorf("RenderLayerTooltip = %d, want %d", RenderLayerTooltip, types.LayerTooltip)
	}
	if RenderLayerInspector != types.LayerInspector {
		t.Errorf("RenderLayerInspector = %d, want %d", RenderLayerInspector, types.LayerInspector)
	}
}

// =============================================================================
// PaintablePlanes Tests
// =============================================================================

func TestNewPaintablePlanes(t *testing.T) {
	pp := NewPaintablePlanes()
	if pp == nil {
		t.Fatal("NewPaintablePlanes() returned nil")
	}
	if pp.CountBoxes() != 0 {
		t.Errorf("New PaintablePlanes should have 0 boxes, got %d", pp.CountBoxes())
	}
}

func TestPaintablePlanes_AddToLayer(t *testing.T) {
	pp := NewPaintablePlanes()

	// 创建一个简单的 PaintableBox
	box := NewPaintableBoxWithBounds(nil, 0, 0, 10, 5)

	pp.AddToLayer(RenderLayerBase, box)

	if pp.CountBoxes() != 1 {
		t.Errorf("CountBoxes() = %d, want 1", pp.CountBoxes())
	}

	boxes := pp.GetLayer(RenderLayerBase)
	if len(boxes) != 1 {
		t.Errorf("GetLayer(Base) returned %d boxes, want 1", len(boxes))
	}
}

func TestPaintablePlanes_AddToLayer_MultipleLayers(t *testing.T) {
	pp := NewPaintablePlanes()

	baseBox := NewPaintableBoxWithBounds(nil, 0, 0, 80, 25)
	modalBox := NewPaintableBoxWithBounds(nil, 20, 10, 40, 5)
	tooltipBox := NewPaintableBoxWithBounds(nil, 30, 8, 20, 3)

	pp.AddToLayer(RenderLayerBase, baseBox)
	pp.AddToLayer(RenderLayerModal, modalBox)
	pp.AddToLayer(RenderLayerTooltip, tooltipBox)

	if pp.CountBoxes() != 3 {
		t.Errorf("CountBoxes() = %d, want 3", pp.CountBoxes())
	}

	if pp.IsLayerEmpty(RenderLayerBase) {
		t.Error("Base layer should not be empty")
	}
	if pp.IsLayerEmpty(RenderLayerModal) {
		t.Error("Modal layer should not be empty")
	}
	if pp.IsLayerEmpty(RenderLayerTooltip) {
		t.Error("Tooltip layer should not be empty")
	}
	if !pp.IsLayerEmpty(RenderLayerOverlay) {
		t.Error("Overlay layer should be empty")
	}
}

func TestPaintablePlanes_AddToLayer_Nil(t *testing.T) {
	pp := NewPaintablePlanes()

	pp.AddToLayer(RenderLayerBase, nil)

	if pp.CountBoxes() != 0 {
		t.Errorf("Adding nil should not increase count, got %d", pp.CountBoxes())
	}
}

func TestPaintablePlanes_GetHighestLayer(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(pp *PaintablePlanes)
		expected RenderLayer
	}{
		{
			name:     "empty planes",
			setup:    func(pp *PaintablePlanes) {},
			expected: RenderLayerBase,
		},
		{
			name: "only base",
			setup: func(pp *PaintablePlanes) {
				pp.AddToLayer(RenderLayerBase, NewPaintableBoxWithBounds(nil, 0, 0, 10, 5))
			},
			expected: RenderLayerBase,
		},
		{
			name: "base and modal",
			setup: func(pp *PaintablePlanes) {
				pp.AddToLayer(RenderLayerBase, NewPaintableBoxWithBounds(nil, 0, 0, 10, 5))
				pp.AddToLayer(RenderLayerModal, NewPaintableBoxWithBounds(nil, 5, 5, 10, 5))
			},
			expected: RenderLayerModal,
		},
		{
			name: "inspector is highest",
			setup: func(pp *PaintablePlanes) {
				pp.AddToLayer(RenderLayerBase, NewPaintableBoxWithBounds(nil, 0, 0, 10, 5))
				pp.AddToLayer(RenderLayerModal, NewPaintableBoxWithBounds(nil, 5, 5, 10, 5))
				pp.AddToLayer(RenderLayerInspector, NewPaintableBoxWithBounds(nil, 80, 5, 20, 10))
			},
			expected: RenderLayerInspector,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pp := NewPaintablePlanes()
			tt.setup(pp)
			if got := pp.GetHighestLayer(); got != tt.expected {
				t.Errorf("GetHighestLayer() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestPaintablePlanes_GetRenderOrder(t *testing.T) {
	pp := NewPaintablePlanes()
	order := pp.GetRenderOrder()

	expected := []RenderLayer{
		RenderLayerBase,
		RenderLayerOverlay,
		RenderLayerModal,
		RenderLayerTooltip,
		RenderLayerInspector,
	}

	if len(order) != len(expected) {
		t.Fatalf("GetRenderOrder() returned %d layers, want %d", len(order), len(expected))
	}

	for i, l := range order {
		if l != expected[i] {
			t.Errorf("GetRenderOrder()[%d] = %v, want %v", i, l, expected[i])
		}
	}
}

func TestPaintablePlanes_Clear(t *testing.T) {
	pp := NewPaintablePlanes()
	pp.AddToLayer(RenderLayerBase, NewPaintableBoxWithBounds(nil, 0, 0, 10, 5))
	pp.AddToLayer(RenderLayerModal, NewPaintableBoxWithBounds(nil, 5, 5, 10, 5))

	pp.Clear()

	if pp.CountBoxes() != 0 {
		t.Errorf("After Clear(), CountBoxes() = %d, want 0", pp.CountBoxes())
	}
}

func TestPaintablePlanes_Iterate(t *testing.T) {
	pp := NewPaintablePlanes()
	pp.AddToLayer(RenderLayerBase, NewPaintableBoxWithBounds(nil, 0, 0, 10, 5))
	pp.AddToLayer(RenderLayerModal, NewPaintableBoxWithBounds(nil, 5, 5, 10, 5))

	var layers []RenderLayer
	pp.Iterate(func(layer RenderLayer, box *PaintableBox) bool {
		layers = append(layers, layer)
		return true
	})

	// 应该按从低到高的顺序遍历
	if len(layers) != 2 {
		t.Fatalf("Iterate visited %d layers, want 2", len(layers))
	}
	if layers[0] != RenderLayerBase {
		t.Errorf("First layer = %v, want Base", layers[0])
	}
	if layers[1] != RenderLayerModal {
		t.Errorf("Second layer = %v, want Modal", layers[1])
	}
}

func TestPaintablePlanes_IterrateReverse(t *testing.T) {
	pp := NewPaintablePlanes()
	pp.AddToLayer(RenderLayerBase, NewPaintableBoxWithBounds(nil, 0, 0, 10, 5))
	pp.AddToLayer(RenderLayerModal, NewPaintableBoxWithBounds(nil, 5, 5, 10, 5))

	var layers []RenderLayer
	pp.IterateReverse(func(layer RenderLayer, box *PaintableBox) bool {
		layers = append(layers, layer)
		return true
	})

	// 应该按从高到低的顺序遍历
	if len(layers) != 2 {
		t.Fatalf("IterateReverse visited %d layers, want 2", len(layers))
	}
	if layers[0] != RenderLayerModal {
		t.Errorf("First layer = %v, want Modal", layers[0])
	}
	if layers[1] != RenderLayerBase {
		t.Errorf("Second layer = %v, want Base", layers[1])
	}
}

func TestPaintablePlanes_Iterrate_Stop(t *testing.T) {
	pp := NewPaintablePlanes()
	pp.AddToLayer(RenderLayerBase, NewPaintableBoxWithBounds(nil, 0, 0, 10, 5))
	pp.AddToLayer(RenderLayerModal, NewPaintableBoxWithBounds(nil, 5, 5, 10, 5))

	count := 0
	pp.Iterate(func(layer RenderLayer, box *PaintableBox) bool {
		count++
		return false // 停止遍历
	})

	if count != 1 {
		t.Errorf("Iterate with stop visited %d boxes, want 1", count)
	}
}

// =============================================================================
// PaintableLayouts Tests
// =============================================================================

func TestPaintableLayouts_Basic(t *testing.T) {
	layouts := make(PaintableLayouts)

	rootBox := NewPaintableBoxWithBounds(nil, 0, 0, 80, 25)
	layouts[RenderLayerBase] = NewPaintableLayout(rootBox)

	if layouts[RenderLayerBase] == nil {
		t.Error("PaintableLayouts should store layout")
	}
	if layouts[RenderLayerModal] != nil {
		t.Error("Non-existent layer should return nil")
	}
}

func TestBuildPaintablePlanesFromLayouts(t *testing.T) {
	layouts := make(PaintableLayouts)

	// 创建一个带有子节点的布局树
	rootBox := NewPaintableBoxWithBounds(nil, 0, 0, 80, 25)
	childBox := NewPaintableBoxWithBounds(nil, 0, 0, 40, 12)
	rootBox.AddChild(childBox)
	layouts[RenderLayerBase] = NewPaintableLayout(rootBox)

	pp := BuildPaintablePlanesFromLayouts(layouts)

	// 应该包含 rootBox 和 childBox
	if pp.CountBoxes() != 2 {
		t.Errorf("BuildPaintablePlanesFromLayouts returned %d boxes, want 2", pp.CountBoxes())
	}
}

// =============================================================================
// BuildFromPaintableBox Tests
// =============================================================================

func TestBuildFromPaintableBox(t *testing.T) {
	// 创建一个树结构
	rootBox := NewPaintableBoxWithBounds(nil, 0, 0, 80, 25)
	rootBox.Layer = 0

	child1 := NewPaintableBoxWithBounds(nil, 0, 0, 40, 25)
	child1.Layer = 0
	rootBox.AddChild(child1)

	child2 := NewPaintableBoxWithBounds(nil, 40, 0, 40, 25)
	child2.Layer = 2 // Modal
	rootBox.AddChild(child2)

	pp := BuildFromPaintableBox(rootBox)

	if pp.CountBoxes() != 3 {
		t.Errorf("BuildFromPaintableBox returned %d boxes, want 3", pp.CountBoxes())
	}

	if pp.CountBoxes() != 3 {
		t.Errorf("BuildFromPaintableBox CountBoxes = %d, want 3", pp.CountBoxes())
	}
}

func TestBuildFromPaintableBox_Nil(t *testing.T) {
	pp := BuildFromPaintableBox(nil)
	if pp == nil {
		t.Fatal("BuildFromPaintableBox(nil) returned nil")
	}
	if pp.CountBoxes() != 0 {
		t.Errorf("BuildFromPaintableBox(nil) should have 0 boxes, got %d", pp.CountBoxes())
	}
}
