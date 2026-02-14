package layer

import (
	"testing"

	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/compute"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// TestNewRenderPlanes tests creating a new RenderPlanes
func TestNewRenderPlanes(t *testing.T) {
	rp := NewRenderPlanes()

	if rp == nil {
		t.Fatal("NewRenderPlanes returned nil")
	}

	if rp.planes == nil {
		t.Fatal("planes map not initialized")
	}

	if rp.renderOrder == nil {
		t.Fatal("renderOrder not initialized")
	}

	expectedLayers := []rtui.Layer{
		rtui.LayerBase,
		rtui.LayerOverlay,
		rtui.LayerModal,
		rtui.LayerTooltip,
		rtui.LayerInspector,
	}

	if len(rp.renderOrder) != len(expectedLayers) {
		t.Errorf("Expected %d render orders, got %d", len(expectedLayers), len(rp.renderOrder))
	}

	for i, expected := range expectedLayers {
		if rp.renderOrder[i] != expected {
			t.Errorf("renderOrder[%d]: expected %v, got %v", i, expected, rp.renderOrder[i])
		}
	}
}

// TestRenderPlanesAddToLayer tests adding ComputedBox to layers
func TestRenderPlanesAddToLayer(t *testing.T) {
	rp := NewRenderPlanes()

	box := &compute.ComputedBox{
		Box:    runtime.Box{X: 0, Y: 0, Width: 10, Height: 10},
		Layer:  rtui.LayerBase,
		NodeID: 1,
	}

	rp.AddToLayer(rtui.LayerBase, box)

	if rp.IsLayerEmpty(rtui.LayerBase) {
		t.Error("LayerBase should not be empty")
	}

	boxes := rp.GetLayer(rtui.LayerBase)
	if len(boxes) != 1 {
		t.Errorf("Expected 1 box in LayerBase, got %d", len(boxes))
	}

	if boxes[0] != box {
		t.Error("Added box not found in layer")
	}
}

// TestRenderPlanesMultipleLayers tests boxes in multiple layers
func TestRenderPlanesMultipleLayers(t *testing.T) {
	rp := NewRenderPlanes()

	baseBox := &compute.ComputedBox{
		Box:    runtime.Box{X: 0, Y: 0, Width: 10, Height: 10},
		Layer:  rtui.LayerBase,
		NodeID: 1,
	}

	modalBox := &compute.ComputedBox{
		Box:    runtime.Box{X: 5, Y: 5, Width: 20, Height: 20},
		Layer:  rtui.LayerModal,
		NodeID: 2,
	}

	overlayBox := &compute.ComputedBox{
		Box:    runtime.Box{X: 10, Y: 10, Width: 15, Height: 15},
		Layer:  rtui.LayerOverlay,
		NodeID: 3,
	}

	rp.AddToLayer(rtui.LayerBase, baseBox)
	rp.AddToLayer(rtui.LayerModal, modalBox)
	rp.AddToLayer(rtui.LayerOverlay, overlayBox)

	// Check counts
	if rp.CountBoxes() != 3 {
		t.Errorf("Expected 3 boxes total, got %d", rp.CountBoxes())
	}

	// Check each layer
	if len(rp.GetLayer(rtui.LayerBase)) != 1 {
		t.Error("LayerBase should have 1 box")
	}

	if len(rp.GetLayer(rtui.LayerModal)) != 1 {
		t.Error("LayerModal should have 1 box")
	}

	if len(rp.GetLayer(rtui.LayerOverlay)) != 1 {
		t.Error("LayerOverlay should have 1 box")
	}

	if !rp.IsLayerEmpty(rtui.LayerTooltip) {
		t.Error("LayerTooltip should be empty")
	}
}

// TestRenderPlanesAddToLayerNilBox tests adding nil box
func TestRenderPlanesAddToLayerNilBox(t *testing.T) {
	rp := NewRenderPlanes()

	rp.AddToLayer(rtui.LayerBase, nil)

	if !rp.IsLayerEmpty(rtui.LayerBase) {
		t.Error("LayerBase should still be empty after adding nil box")
	}
}

// TestBuildRenderPlanes tests building RenderPlanes from ComputedBox tree
func TestBuildRenderPlanes(t *testing.T) {
	// Create a simple tree with boxes in different layers
	baseBox := &compute.ComputedBox{
		Box:    runtime.Box{X: 0, Y: 0, Width: 10, Height: 10},
		Layer:  rtui.LayerBase,
		NodeID: 1,
		Children: []*compute.ComputedBox{
			{
				Box:    runtime.Box{X: 0, Y: 0, Width: 5, Height: 5},
				Layer:  rtui.LayerBase,
				NodeID: 2,
			},
			{
				Box:    runtime.Box{X: 5, Y: 0, Width: 5, Height: 5},
				Layer:  rtui.LayerModal,
				NodeID: 3,
			},
		},
	}

	rp := BuildRenderPlanes(baseBox)

	// Should have 3 boxes total (1 base root + 2 children)
	if rp.CountBoxes() != 3 {
		t.Errorf("Expected 3 boxes total, got %d", rp.CountBoxes())
	}

	// LayerBase should have 2 boxes (root + 1 child)
	if len(rp.GetLayer(rtui.LayerBase)) != 2 {
		t.Errorf("Expected 2 boxes in LayerBase, got %d", len(rp.GetLayer(rtui.LayerBase)))
	}

	// LayerModal should have 1 box
	if len(rp.GetLayer(rtui.LayerModal)) != 1 {
		t.Errorf("Expected 1 box in LayerModal, got %d", len(rp.GetLayer(rtui.LayerModal)))
	}
}

// TestBuildRenderPlanesNilRoot tests building from nil root
func TestBuildRenderPlanesNilRoot(t *testing.T) {
	rp := BuildRenderPlanes(nil)

	if rp == nil {
		t.Fatal("BuildRenderPlanes returned nil for nil root")
	}

	if rp.CountBoxes() != 0 {
		t.Errorf("Expected 0 boxes from nil root, got %d", rp.CountBoxes())
	}
}

// TestRenderPlanesIterate tests iterating over planes
func TestRenderPlanesIterate(t *testing.T) {
	rp := NewRenderPlanes()

	baseBox := &compute.ComputedBox{
		Box:    runtime.Box{X: 0, Y: 0, Width: 10, Height: 10},
		Layer:  rtui.LayerBase,
		NodeID: 1,
	}

	modalBox := &compute.ComputedBox{
		Box:    runtime.Box{X: 5, Y: 5, Width: 20, Height: 20},
		Layer:  rtui.LayerModal,
		NodeID: 2,
	}

	rp.AddToLayer(rtui.LayerBase, baseBox)
	rp.AddToLayer(rtui.LayerModal, modalBox)

	// Test forward iteration (low to high)
	var visited []uint64
	rp.Iterate(func(layer rtui.Layer, box *compute.ComputedBox) bool {
		visited = append(visited, box.NodeID)
		return true
	})

	if len(visited) != 2 {
		t.Errorf("Expected to visit 2 boxes, visited %d", len(visited))
	}

	if visited[0] != 1 && visited[1] != 1 {
		t.Error("LayerBase box should be visited first (low to high)")
	}
}

// TestRenderPlanesIterateReverse tests iterating in reverse order
func TestRenderPlanesIterateReverse(t *testing.T) {
	rp := NewRenderPlanes()

	baseBox := &compute.ComputedBox{
		Box:    runtime.Box{X: 0, Y: 0, Width: 10, Height: 10},
		Layer:  rtui.LayerBase,
		NodeID: 1,
	}

	modalBox := &compute.ComputedBox{
		Box:    runtime.Box{X: 5, Y: 5, Width: 20, Height: 20},
		Layer:  rtui.LayerModal,
		NodeID: 2,
	}

	rp.AddToLayer(rtui.LayerBase, baseBox)
	rp.AddToLayer(rtui.LayerModal, modalBox)

	// Test reverse iteration (high to low)
	var visited []uint64
	rp.IterateReverse(func(layer rtui.Layer, box *compute.ComputedBox) bool {
		visited = append(visited, box.NodeID)
		return true
	})

	if len(visited) != 2 {
		t.Errorf("Expected to visit 2 boxes, visited %d", len(visited))
	}

	if visited[0] != 2 {
		t.Error("LayerModal box should be visited first (high to low)")
	}
}

// TestRenderPlanesClear tests clearing all layers
func TestRenderPlanesClear(t *testing.T) {
	rp := NewRenderPlanes()

	box := &compute.ComputedBox{
		Box:    runtime.Box{X: 0, Y: 0, Width: 10, Height: 10},
		Layer:  rtui.LayerBase,
		NodeID: 1,
	}

	rp.AddToLayer(rtui.LayerBase, box)

	if rp.CountBoxes() != 1 {
		t.Errorf("Expected 1 box before clear, got %d", rp.CountBoxes())
	}

	rp.Clear()

	if rp.CountBoxes() != 0 {
		t.Errorf("Expected 0 boxes after clear, got %d", rp.CountBoxes())
	}

	if !rp.IsLayerEmpty(rtui.LayerBase) {
		t.Error("LayerBase should be empty after clear")
	}
}

// TestRenderPlanesValidate tests validation of RenderPlanes
func TestRenderPlanesValidate(t *testing.T) {
	rp := NewRenderPlanes()

	// Add box with matching layer
	validBox := &compute.ComputedBox{
		Box:    runtime.Box{X: 0, Y: 0, Width: 10, Height: 10},
		Layer:  rtui.LayerBase,
		NodeID: 1,
	}
	rp.AddToLayer(rtui.LayerBase, validBox)

	// Add box with mismatched layer (simulate bug)
	mismatchedBox := &compute.ComputedBox{
		Box:    runtime.Box{X: 0, Y: 0, Width: 10, Height: 10},
		Layer:  rtui.LayerModal, // Box claims to be Modal
		NodeID: 2,
	}
	rp.AddToLayer(rtui.LayerBase, mismatchedBox) // But added to Base layer

	errors := rp.Validate()

	if len(errors) != 1 {
		t.Errorf("Expected 1 validation error, got %d", len(errors))
	}
}

// TestBuildFromFiberBasic tests building RenderPlanes from a simple Fiber tree (single layer)
func TestBuildFromFiberBasic(t *testing.T) {
	// Create a simple Fiber tree with ComputedBox
	box1 := &compute.ComputedBox{
		Box:    runtime.Box{X: 0, Y: 0, Width: 10, Height: 10},
		Layer:  rtui.LayerBase,
		NodeID: 1,
	}

	box2 := &compute.ComputedBox{
		Box:    runtime.Box{X: 0, Y: 0, Width: 5, Height: 5},
		Layer:  rtui.LayerBase,
		NodeID: 2,
	}

	root := &rtui.Fiber{
		Layer:       rtui.LayerBase,
		NodeID:      1,
		ComputedBox: box1,
		Child: &rtui.Fiber{
			Layer:       rtui.LayerBase,
			NodeID:      2,
			ComputedBox: box2,
		},
	}

	rp := BuildFromFiber(root)

	// Verify both boxes are collected
	if rp.CountBoxes() != 2 {
		t.Errorf("Expected 2 boxes, got %d", rp.CountBoxes())
	}

	// Verify LayerBase has both boxes
	baseBoxes := rp.GetLayer(rtui.LayerBase)
	if len(baseBoxes) != 2 {
		t.Errorf("Expected 2 boxes in LayerBase, got %d", len(baseBoxes))
	}

	// Verify content
	foundIDs := make(map[uint64]bool)
	for _, box := range baseBoxes {
		foundIDs[box.NodeID] = true
	}

	if !foundIDs[1] {
		t.Error("NodeID 1 not found in LayerBase")
	}
	if !foundIDs[2] {
		t.Error("NodeID 2 not found in LayerBase")
	}
}

// TestBuildFromFiberMultipleLayers tests building RenderPlanes from Fiber tree with multiple layers
func TestBuildFromFiberMultipleLayers(t *testing.T) {
	// Create Fiber tree with boxes in different layers
	box1 := &compute.ComputedBox{
		Box:    runtime.Box{X: 0, Y: 0, Width: 10, Height: 10},
		Layer:  rtui.LayerBase,
		NodeID: 1,
	}

	box2 := &compute.ComputedBox{
		Box:    runtime.Box{X: 0, Y: 0, Width: 5, Height: 5},
		Layer:  rtui.LayerBase,
		NodeID: 2,
	}

	box3 := &compute.ComputedBox{
		Box:    runtime.Box{X: 5, Y: 5, Width: 20, Height: 20},
		Layer:  rtui.LayerModal,
		NodeID: 3,
	}

	box4 := &compute.ComputedBox{
		Box:    runtime.Box{X: 10, Y: 10, Width: 15, Height: 15},
		Layer:  rtui.LayerOverlay,
		NodeID: 4,
	}

	box5 := &compute.ComputedBox{
		Box:    runtime.Box{X: 2, Y: 2, Width: 8, Height: 8},
		Layer:  rtui.LayerTooltip,
		NodeID: 5,
	}

	// Fiber tree structure:
	//   root (Base, box1)
	//   ├── child1 (Base, box2)
	//   ├── child2 (Modal, box3)
	//   │   └── grandchild (Overlay, box4)
	//   └── child3 (Tooltip, box5)

	root := &rtui.Fiber{
		Layer:       rtui.LayerBase,
		NodeID:      1,
		ComputedBox: box1,
		Child: &rtui.Fiber{
			Layer:       rtui.LayerBase,
			NodeID:      2,
			ComputedBox: box2,
			Sibling: &rtui.Fiber{
				Layer:       rtui.LayerModal,
				NodeID:      3,
				ComputedBox: box3,
				Child: &rtui.Fiber{
					Layer:       rtui.LayerOverlay,
					NodeID:      4,
					ComputedBox: box4,
				},
				Sibling: &rtui.Fiber{
					Layer:       rtui.LayerTooltip,
					NodeID:      5,
					ComputedBox: box5,
				},
			},
		},
	}

	rp := BuildFromFiber(root)

	// Verify total count
	if rp.CountBoxes() != 5 {
		t.Errorf("Expected 5 boxes total, got %d", rp.CountBoxes())
	}

	// Verify each layer
	baseBoxes := rp.GetLayer(rtui.LayerBase)
	if len(baseBoxes) != 2 {
		t.Errorf("Expected 2 boxes in LayerBase, got %d", len(baseBoxes))
	}

	modalBoxes := rp.GetLayer(rtui.LayerModal)
	if len(modalBoxes) != 1 {
		t.Errorf("Expected 1 box in LayerModal, got %d", len(modalBoxes))
	}

	overlayBoxes := rp.GetLayer(rtui.LayerOverlay)
	if len(overlayBoxes) != 1 {
		t.Errorf("Expected 1 box in LayerOverlay, got %d", len(overlayBoxes))
	}

	tooltipBoxes := rp.GetLayer(rtui.LayerTooltip)
	if len(tooltipBoxes) != 1 {
		t.Errorf("Expected 1 box in LayerTooltip, got %d", len(tooltipBoxes))
	}

	// Verify NodeIDs are correct
	if modalBoxes[0].NodeID != 3 {
		t.Errorf("Modal box should have NodeID 3, got %d", modalBoxes[0].NodeID)
	}
	if overlayBoxes[0].NodeID != 4 {
		t.Errorf("Overlay box should have NodeID 4, got %d", overlayBoxes[0].NodeID)
	}
	if tooltipBoxes[0].NodeID != 5 {
		t.Errorf("Tooltip box should have NodeID 5, got %d", tooltipBoxes[0].NodeID)
	}
}

// TestBuildFromFiberNil tests building from nil Fiber
func TestBuildFromFiberNil(t *testing.T) {
	rp := BuildFromFiber(nil)

	if rp == nil {
		t.Fatal("BuildFromFiber returned nil for nil root")
	}

	if rp.CountBoxes() != 0 {
		t.Errorf("Expected 0 boxes from nil root, got %d", rp.CountBoxes())
	}
}

// TestBuildFromFiberNilComputedBox tests building from Fiber with nil ComputedBox
func TestBuildFromFiberNilComputedBox(t *testing.T) {
	root := &rtui.Fiber{
		Layer:       rtui.LayerBase,
		NodeID:      1,
		ComputedBox: nil, // No ComputedBox
		Child: &rtui.Fiber{
			Layer:       rtui.LayerModal,
			NodeID:      2,
			ComputedBox: nil,
		},
	}

	rp := BuildFromFiber(root)

	if rp.CountBoxes() != 0 {
		t.Errorf("Expected 0 boxes when all ComputedBox are nil, got %d", rp.CountBoxes())
	}
}

// TestBuildFromFiberEmptyTree tests building from Fiber with no children
func TestBuildFromFiberEmptyTree(t *testing.T) {
	box := &compute.ComputedBox{
		Box:    runtime.Box{X: 0, Y: 0, Width: 10, Height: 10},
		Layer:  rtui.LayerBase,
		NodeID: 1,
	}

	root := &rtui.Fiber{
		Layer:       rtui.LayerBase,
		NodeID:      1,
		ComputedBox: box,
		Child:       nil, // No children
	}

	rp := BuildFromFiber(root)

	if rp.CountBoxes() != 1 {
		t.Errorf("Expected 1 box from empty tree, got %d", rp.CountBoxes())
	}

	baseBoxes := rp.GetLayer(rtui.LayerBase)
	if len(baseBoxes) != 1 {
		t.Errorf("Expected 1 box in LayerBase, got %d", len(baseBoxes))
	}

	if baseBoxes[0].NodeID != 1 {
		t.Errorf("Expected NodeID 1, got %d", baseBoxes[0].NodeID)
	}
}
