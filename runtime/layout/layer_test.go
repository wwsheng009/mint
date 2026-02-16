package layout

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// Layer Tests
// =============================================================================

func TestLayer_String(t *testing.T) {
	tests := []struct {
		layer    Layer
		expected string
	}{
		{LayerBase, "base"},
		{LayerDropdown, "dropdown"},
		{LayerSticky, "sticky"},
		{LayerFixed, "fixed"},
		{LayerModalBackdrop, "modalBackdrop"},
		{LayerModal, "modal"},
		{LayerPopover, "popover"},
		{LayerTooltip, "tooltip"},
		{Layer(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.layer.String())
		})
	}
}

func TestLayer_ZIndex(t *testing.T) {
	tests := []struct {
		layer    Layer
		expected int
	}{
		{LayerBase, 0},
		{LayerDropdown, 1000},
		{LayerSticky, 2000},
		{LayerFixed, 3000},
		{LayerModalBackdrop, 4000},
		{LayerModal, 5000},
		{LayerPopover, 6000},
		{LayerTooltip, 7000},
	}

	for _, tt := range tests {
		t.Run(tt.layer.String(), func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.layer.ZIndex())
		})
	}
}

func TestLayer_Comparison(t *testing.T) {
	assert.True(t, LayerModal.IsHigher(LayerBase))
	assert.True(t, LayerBase.IsLower(LayerModal))
	assert.False(t, LayerBase.IsHigher(LayerBase))
	assert.False(t, LayerBase.IsLower(LayerBase))
}

// =============================================================================
// LayeredNode Tests
// =============================================================================

func TestNewLayeredNode(t *testing.T) {
	child := NewMockNode("child", 50, 30)
	node := NewLayeredNode("layered", child, LayerModal, 10)

	assert.Equal(t, "layered", node.ID())
	assert.Equal(t, "layered", node.Type())
	assert.Equal(t, LayerModal, node.GetLayer())
	assert.Equal(t, 10, node.GetZIndex())
	assert.Equal(t, child, node.GetChild())
}

func TestLayeredNode_Children(t *testing.T) {
	child := NewMockNode("child", 50, 30)
	node := NewLayeredNode("layered", child, LayerModal, 10)

	children := node.Children()
	assert.Len(t, children, 1)
	assert.Equal(t, child, children[0])

	// Nil child
	nilNode := NewLayeredNode("nil", nil, LayerBase, 0)
	assert.Nil(t, nilNode.Children())
}

func TestLayeredNode_GetSize(t *testing.T) {
	child := NewMockNode("child", 50, 30)
	node := NewLayeredNode("layered", child, LayerModal, 10)

	w, h := node.GetSize()
	assert.Equal(t, 50, w)
	assert.Equal(t, 30, h)

	// Nil child
	nilNode := NewLayeredNode("nil", nil, LayerBase, 0)
	w, h = nilNode.GetSize()
	assert.Equal(t, 0, w)
	assert.Equal(t, 0, h)
}

func TestLayeredNode_EffectiveZIndex(t *testing.T) {
	child := NewMockNode("child", 50, 30)
	node := NewLayeredNode("layered", child, LayerModal, 10)

	// LayerModal = 5000 + 10 = 5010
	assert.Equal(t, 5010, node.EffectiveZIndex())

	// Base layer
	baseNode := NewLayeredNode("base", child, LayerBase, 5)
	assert.Equal(t, 5, baseNode.EffectiveZIndex())
}

// =============================================================================
// LayeredLayoutResult Tests
// =============================================================================

func TestNewLayeredLayoutResult(t *testing.T) {
	root := &LayoutBox{ID: "root"}
	result := NewLayeredLayoutResult(root)

	assert.Equal(t, root, result.Root)
	assert.NotNil(t, result.Layers)
	assert.NotNil(t, result.AllBoxes)
}

func TestLayeredLayoutResult_AddBox(t *testing.T) {
	root := &LayoutBox{ID: "root"}
	result := NewLayeredLayoutResult(root)

	box1 := &LayoutBox{ID: "box1", Layer: LayerBase}
	box2 := &LayoutBox{ID: "box2", Layer: LayerModal}

	result.AddBox(box1, LayerBase)
	result.AddBox(box2, LayerModal)

	assert.Len(t, result.Layers[LayerBase], 1)
	assert.Len(t, result.Layers[LayerModal], 1)
	assert.Len(t, result.AllBoxes, 2)
}

func TestLayeredLayoutResult_GetLayer(t *testing.T) {
	root := &LayoutBox{ID: "root"}
	result := NewLayeredLayoutResult(root)

	box1 := &LayoutBox{ID: "box1"}
	box2 := &LayoutBox{ID: "box2"}

	result.AddBox(box1, LayerBase)
	result.AddBox(box2, LayerBase)

	boxes := result.GetLayer(LayerBase)
	assert.Len(t, boxes, 2)

	// Empty layer
	boxes = result.GetLayer(LayerModal)
	assert.Len(t, boxes, 0)
}

func TestLayeredLayoutResult_GetLayers(t *testing.T) {
	root := &LayoutBox{ID: "root"}
	result := NewLayeredLayoutResult(root)

	result.AddBox(&LayoutBox{ID: "box1"}, LayerBase)
	result.AddBox(&LayoutBox{ID: "box2"}, LayerModal)

	layers := result.GetLayers()
	assert.Len(t, layers, 2)
}

func TestLayeredLayoutResult_SortByZIndex(t *testing.T) {
	root := &LayoutBox{ID: "root"}
	result := NewLayeredLayoutResult(root)

	// Add boxes with different layers and z-indices
	result.AddBox(&LayoutBox{ID: "modal2", Layer: LayerModal, ZIndex: 2}, LayerModal)
	result.AddBox(&LayoutBox{ID: "base1", Layer: LayerBase, ZIndex: 1}, LayerBase)
	result.AddBox(&LayoutBox{ID: "modal1", Layer: LayerModal, ZIndex: 1}, LayerModal)
	result.AddBox(&LayoutBox{ID: "base2", Layer: LayerBase, ZIndex: 2}, LayerBase)

	sorted := result.SortByZIndex()

	// Base layer should come before modal
	assert.Equal(t, "base1", sorted[0].ID)
	assert.Equal(t, "base2", sorted[1].ID)
	assert.Equal(t, "modal1", sorted[2].ID)
	assert.Equal(t, "modal2", sorted[3].ID)
}

// =============================================================================
// Helper Function Tests
// =============================================================================

func TestIsLayered(t *testing.T) {
	child := NewMockNode("child", 50, 30)
	layered := NewLayeredNode("layered", child, LayerModal, 10)
	regular := NewMockNode("regular", 10, 5)

	assert.True(t, isLayered(layered))
	assert.False(t, isLayered(regular))
}

func TestGetLayerFromNode(t *testing.T) {
	child := NewMockNode("child", 50, 30)
	layered := NewLayeredNode("layered", child, LayerModal, 10)
	regular := NewMockNode("regular", 10, 5)

	assert.Equal(t, LayerModal, GetLayerFromNode(layered))
	assert.Equal(t, LayerBase, GetLayerFromNode(regular))
	assert.Equal(t, LayerBase, GetLayerFromNode(nil))
}

func TestGetZIndexFromNode(t *testing.T) {
	child := NewMockNode("child", 50, 30)
	layered := NewLayeredNode("layered", child, LayerModal, 10)
	regular := NewMockNode("regular", 10, 5)

	assert.Equal(t, 10, GetZIndexFromNode(layered))
	assert.Equal(t, 0, GetZIndexFromNode(regular))
	assert.Equal(t, 0, GetZIndexFromNode(nil))
}

func TestCompareZOrder(t *testing.T) {
	child := NewMockNode("child", 50, 30)
	
	// Higher layer wins
	a := NewLayeredNode("a", child, LayerModal, 0)
	b := NewLayeredNode("b", child, LayerBase, 100)
	assert.Equal(t, 1, CompareZOrder(a, b))  // a is above b
	assert.Equal(t, -1, CompareZOrder(b, a)) // b is below a

	// Same layer, z-index wins
	c := NewLayeredNode("c", child, LayerBase, 10)
	d := NewLayeredNode("d", child, LayerBase, 5)
	assert.Equal(t, 1, CompareZOrder(c, d))  // c is above d

	// Same layer and z-index
	e := NewLayeredNode("e", child, LayerBase, 5)
	assert.Equal(t, 0, CompareZOrder(d, e))
}

func TestIsInHigherLayer(t *testing.T) {
	child := NewMockNode("child", 50, 30)
	a := NewLayeredNode("a", child, LayerModal, 0)
	b := NewLayeredNode("b", child, LayerBase, 100)

	assert.True(t, IsInHigherLayer(a, b))
	assert.False(t, IsInHigherLayer(b, a))
}

func TestParseLayer(t *testing.T) {
	tests := []struct {
		input    string
		expected Layer
	}{
		{"base", LayerBase},
		{"", LayerBase},
		{"dropdown", LayerDropdown},
		{"sticky", LayerSticky},
		{"fixed", LayerFixed},
		{"modalBackdrop", LayerModalBackdrop},
		{"modal-backdrop", LayerModalBackdrop},
		{"modal", LayerModal},
		{"popover", LayerPopover},
		{"tooltip", LayerTooltip},
		{"unknown", LayerBase},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, ParseLayer(tt.input))
		})
	}
}

// =============================================================================
// LayoutBox with Layer Tests
// =============================================================================

func TestLayoutBox_Layer(t *testing.T) {
	box := &LayoutBox{
		ID:      "test",
		X:       10,
		Y:       20,
		Width:   100,
		Height:  50,
		Layer:   LayerModal,
		ZIndex:  5,
	}

	assert.Equal(t, LayerModal, box.Layer)
	assert.Equal(t, 5, box.ZIndex)
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkLayer_ZIndex(b *testing.B) {
	layer := LayerModal
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = layer.ZIndex()
	}
}

func BenchmarkGetLayerFromNode(b *testing.B) {
	child := NewMockNode("child", 50, 30)
	layered := NewLayeredNode("layered", child, LayerModal, 10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetLayerFromNode(layered)
	}
}

func BenchmarkCompareZOrder(b *testing.B) {
	child := NewMockNode("child", 50, 30)
	a := NewLayeredNode("a", child, LayerModal, 10)
	d := NewLayeredNode("d", child, LayerBase, 5)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CompareZOrder(a, d)
	}
}

func BenchmarkLayeredLayoutResult_SortByZIndex(b *testing.B) {
	root := &LayoutBox{ID: "root"}
	result := NewLayeredLayoutResult(root)

	for i := 0; i < 10; i++ {
		result.AddBox(&LayoutBox{ID: "base", Layer: LayerBase, ZIndex: i}, LayerBase)
		result.AddBox(&LayoutBox{ID: "modal", Layer: LayerModal, ZIndex: i}, LayerModal)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = result.SortByZIndex()
	}
}
