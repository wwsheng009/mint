package render

// =============================================================================
// Debug tests for buffer clearing
// =============================================================================

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/paint"
)

// TestPaintEngine_SimpleBoxMovement tests basic box movement with same ID
func TestPaintEngine_SimpleBoxMovement(t *testing.T) {
	engine := NewPaintEngine()
	buffer := paint.NewBuffer(80, 25)

	// Test debug mode

	// Frame 1: Box at x=10
	node1 := NewMockPaintableTextNode("box-1", "HELLO")
	box1 := paint.NewPaintableBoxWithBounds(node1, 10, 5, 8, 1)

	planes1 := paint.NewPaintablePlanes()
	planes1.AddToLayer(paint.RenderLayerBase, box1)

	t.Log("Frame 1: Painting box at x=10")
	err := engine.PaintPaintablePlanes(planes1, buffer)
	if err != nil {
		t.Fatalf("Frame 1 error = %v", err)
	}

	// Verify Frame 1
	if buffer.GetContent(10, 5).Cluster != "H" {
		t.Errorf("Frame 1: Expected 'H' at (10,5), got '%s'", buffer.GetContent(10, 5).Cluster)
	}

	// Frame 2: Same box ID, but moved to x=20
	node2 := NewMockPaintableTextNode("box-1", "HELLO")  // Same ID!
	box2 := paint.NewPaintableBoxWithBounds(node2, 20, 5, 8, 1)  // Different position

	planes2 := paint.NewPaintablePlanes()
	planes2.AddToLayer(paint.RenderLayerBase, box2)

	t.Log("Frame 2: Painting same box ID at x=20")
	err = engine.PaintPaintablePlanes(planes2, buffer)
	if err != nil {
		t.Fatalf("Frame 2 error = %v", err)
	}

	// Check old position is cleared
	if buffer.GetContent(10, 5).Cluster != "" && buffer.GetContent(10, 5).Cluster != " " {
		t.Errorf("BUG: Old position (10,5) still has '%s'", buffer.GetContent(10, 5).Cluster)
	}

	// Check new position has content
	if buffer.GetContent(20, 5).Cluster != "H" {
		t.Errorf("Frame 2: Expected 'H' at (20,5), got '%s'", buffer.GetContent(20, 5).Cluster)
	}

	t.Log("Frame 2: ✓ Box movement verified")
}

// TestPaintEngine_BoxRemoval tests box removal
func TestPaintEngine_BoxRemoval(t *testing.T) {
	engine := NewPaintEngine()
	buffer := paint.NewBuffer(80, 25)

	// Frame 1: Draw a box
	node1 := NewMockPaintableTextNode("removable-1", "DELETE")
	box1 := paint.NewPaintableBoxWithBounds(node1, 5, 10, 10, 1)

	planes1 := paint.NewPaintablePlanes()
	planes1.AddToLayer(paint.RenderLayerBase, box1)

	err := engine.PaintPaintablePlanes(planes1, buffer)
	if err != nil {
		t.Fatalf("Frame 1 error = %v", err)
	}

	// Verify Frame 1
	if buffer.GetContent(5, 10).Cluster != "D" {
		t.Errorf("Frame 1: Expected 'D'")
	}

	// Frame 2: Remove the box (draw nothing)
	planes2 := paint.NewPaintablePlanes()
	// Empty planes

	t.Log("Frame 2: Removing box")
	err = engine.PaintPaintablePlanes(planes2, buffer)
	if err != nil {
		t.Fatalf("Frame 2 error = %v", err)
	}

	// Check old position is cleared
	txt := buffer.GetContent(5, 10).Cluster
	if txt != "" && txt != " " {
		t.Errorf("BUG: Removed box still shows '%s' at (5,10)", txt)
	}

	t.Log("Frame 2: ✓ Box removal verified")
}
