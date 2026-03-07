package render

// =============================================================================
// Buffer Cleaning Tests - Reproduce buffer wipeout bug
// =============================================================================
//
// This test suite reproduces a critical bug where previous frame content
// is not properly cleared when new content is shorter or at a different
// position.
//
// Bug Description:
//   Frame 1: "计数: 0    [  +  ]   *[  -  ]"
//   Frame 2: "计数: 1   *[  +  ]Rt [  -  ]"  ← "Rt" is leftover from previous render
//
// Root Cause:
//   - SetStringAligned only clears within box.Width
//   - If previous content was wider, leftover characters remain
//   - No mechanism to detect and clear outdated painted regions
//
// Solution: Smart buffer clearing with previous frame tracking
//

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
)

// MockPaintableTextNode is a simple PaintableNode that holds text content
type MockPaintableTextNode struct {
	id          string
	text        string
	nodeType    paint.NodeType
	style       style.Style
}

func (m *MockPaintableTextNode) ID() string                              { return m.id }
func (m *MockPaintableTextNode) Type() string                            { return "MockText" }
func (m *MockPaintableTextNode) Tag() string                             { return "text" }
func (m *MockPaintableTextNode) TextContent() string                     { return m.text }
func (m *MockPaintableTextNode) NodeType() paint.NodeType                { return m.nodeType }
func (m *MockPaintableTextNode) Style() style.Style                      { return m.style }
func (m *MockPaintableTextNode) SetStyle(s style.Style)                  { m.style = s }
func (m *MockPaintableTextNode) Paint(x, y int) []paint.DrawCmd       { return nil }
func (m *MockPaintableTextNode) Children() []paint.PaintableNode         { return nil }

// NewMockPaintableTextNode creates a new mock text node
func NewMockPaintableTextNode(id, text string) *MockPaintableTextNode {
	return &MockPaintableTextNode{
		id:       id,
		text:     text,
		nodeType: paint.NodeTypeText,
		style:    style.Style{FG: "white"},
	}
}

// SetText updates the text content (simulating state update)
func (m *MockPaintableTextNode) SetText(text string) {
	m.text = text
}

// =============================================================================
// Test Case 1: Text Length Change - Reproduce the exact bug
// =============================================================================

// TestPaintEngine_TextLengthChange_BugReproduction reproduces the exact bug
// where a shorter text doesn't clear the previous longer text.
func TestPaintEngine_TextLengthChange_BugReproduction(t *testing.T) {
	engine := NewPaintEngine()
	buffer := paint.NewBuffer(80, 25)

	// Create a text node with initial long text (using ASCII for reliable testing)
	longText := "Count: 0    [  +  ]   *[  -  ]"
	textNode1 := NewMockPaintableTextNode("count-0", longText)
	box1 := paint.NewPaintableBoxWithBounds(textNode1, 0, 0, 30, 1)  // box width = 30
	layout1 := paint.NewPaintableLayout(box1)

	// Frame 1: Paint long text
	t.Log("Frame 1: Painting long text...")
	err := engine.PaintLayout(layout1, buffer)
	if err != nil {
		t.Fatalf("Frame 1 PaintLayout error = %v", err)
	}

	// Verify Frame 1 content
	runes := []rune(longText)
	for i := 0; i < min(20, len(runes)); i++ {
		cell := buffer.GetContent(i, 0)
		expected := string(runes[i])
		if cell.Cluster != expected {
			t.Errorf("Frame 1: Expected '%s' at (%d,0), got '%s'", expected, i, cell.Cluster)
		}
	}
	t.Log("Frame 1: ✓ Verified long text")

	// Frame 2: Update text to shorter content
	shortText := "Count: 1   *[  +  ]"
	textNode2 := NewMockPaintableTextNode("count-1", shortText)  // Same ID to simulate update
	box2 := paint.NewPaintableBoxWithBounds(textNode2, 0, 0, 30, 1)  // Same position and width
	layout2 := paint.NewPaintableLayout(box2)

	// Use the same engine instance to enable smart buffer clearing
	t.Log("Frame 2: Painting shorter text...")
	err = engine.PaintLayout(layout2, buffer)
	if err != nil {
		t.Fatalf("Frame 2 PaintLayout error = %v", err)
	}

	// Verify Frame 2 content - THIS IS WHERE THE BUG OCCURS
	shortRunes := []rune(shortText)
	for i := 0; i < 30; i++ {
		cell := buffer.GetContent(i, 0)

		var expectedString string
		if i < len(shortRunes) {
			expectedString = string(shortRunes[i])
		} else {
			expectedString = " "  // Should be space after text ends
		}

		if cell.Cluster != expectedString {
			// BUG DETECTED: Leftover character from previous frame
			t.Errorf("BUG DETECTED at (%d,0): Expected '%s', got '%s' (leftover from Frame 1)",
				i, expectedString, cell.Cluster)
		}
	}
	t.Log("Frame 2: Verification complete")
}

// =============================================================================
// Test Case 2: Multi-component Scene - Real bug scenario
// =============================================================================

// TestPaintEngine_MultiComponent_BugReproduction reproduces the real bug scenario
// where multiple components on the same line have state changes.
func TestPaintEngine_MultiComponent_BugReproduction(t *testing.T) {
	engine := NewPaintEngine()
	buffer := paint.NewBuffer(80, 25)

	// Frame 1: Box at x=10 (will move to x=20 in Frame 2)
	textNode1 := NewMockPaintableTextNode("movable-box", "HELLO")
	box1 := paint.NewPaintableBoxWithBounds(textNode1, 10, 5, 8, 1)

	// Frame 1: Static box (won't move)
	textNode2 := NewMockPaintableTextNode("static-box", "WORLD")
	box2 := paint.NewPaintableBoxWithBounds(textNode2, 30, 5, 8, 1)

	planes1 := paint.NewPaintablePlanes()
	planes1.AddToLayer(paint.RenderLayerBase, box1)
	planes1.AddToLayer(paint.RenderLayerBase, box2)

	t.Log("Frame 1: Painting boxes at x=10 and x=30...")
	err := engine.PaintPaintablePlanes(planes1, buffer)
	if err != nil {
		t.Fatalf("Frame 1 error = %v", err)
	}

	// Verify Frame 1
	if buffer.GetContent(10, 5).Cluster != "H" {
		t.Errorf("Frame 1: Expected 'H' at (10,5)")
	}
	if buffer.GetContent(30, 5).Cluster != "W" {
		t.Errorf("Frame 1: Expected 'W' at (30,5)")
	}

	// Frame 2: Move box1 from x=10 to x=20
	// Reuse the same node ID with same text to simulate component update with layout change
	textNode3 := NewMockPaintableTextNode("movable-box", "HELLO")  // Same ID as Frame 1!
	box3 := paint.NewPaintableBoxWithBounds(textNode3, 20, 5, 8, 1)   // Moved to x=20

	// Static box stays in place
	box4 := paint.NewPaintableBoxWithBounds(textNode2, 30, 5, 8, 1)   // Same box2

	planes2 := paint.NewPaintablePlanes()
	planes2.AddToLayer(paint.RenderLayerBase, box3)
	planes2.AddToLayer(paint.RenderLayerBase, box4)

	// Use the same engine instance to enable smart buffer clearing
	t.Log("Frame 2: Painting with box moved to x=20...")
	err = engine.PaintPaintablePlanes(planes2, buffer)
	if err != nil {
		t.Fatalf("Frame 2 error = %v", err)
	}

	// BUG CHECK: Old position (x=10) should be cleared
	oldPosChar := buffer.GetContent(10, 5).Cluster
	if oldPosChar != "" && oldPosChar != " " {
		t.Errorf("BUG DETECTED: Old position (10,5) still has '%s', should be empty",
			oldPosChar)
	}

	// Check entire old region
	for x := 10; x < 18; x++ {
		ch := buffer.GetContent(x, 5).Cluster
		if ch != "" && ch != " " {
			t.Errorf("BUG DETECTED: Old region (%d,5) has '%s', should be empty", x, ch)
		}
	}

	// Check new position has the content
	if buffer.GetContent(20, 5).Cluster != "H" {
		t.Errorf("Frame 2: New position: Expected 'H' at (20,5), got '%s'",
			buffer.GetContent(20, 5).Cluster)
	}

	// Static box should remain unchanged
	if buffer.GetContent(30, 5).Cluster != "W" {
		t.Errorf("Frame 2: Static box: Expected 'W' at (30,5), got '%s'",
			buffer.GetContent(30, 5).Cluster)
	}

	t.Log("Frame 2: ✓ Verified box movement and static box preservation")
}

// =============================================================================
// Test Case 2b: Component Position Shift
// =============================================================================

// TestPaintEngine_ComponentPositionShift_BugReproduction tests when
// a component moves to a different position (simulating focus state change).
func TestPaintEngine_ComponentPositionShift_BugReproduction(t *testing.T) {
	engine := NewPaintEngine()
	buffer := paint.NewBuffer(80, 25)

	// Frame 1: Button NOT focused - "[  +  ]"
	textNode1 := NewMockPaintableTextNode("btn-1-1", "[  +  ]")
	box1 := paint.NewPaintableBoxWithBounds(textNode1, 10, 5, 7, 1)

	// Next button IS focused - "*[  -  ]"
	textNode2 := NewMockPaintableTextNode("btn-2-1", "*[  -  ]")
	box2 := paint.NewPaintableBoxWithBounds(textNode2, 20, 5, 7, 1)

	rootBox1 := paint.NewPaintableBox(textNode1)
	rootBox1.X, rootBox1.Y = 0, 0
	rootBox1.Width, rootBox1.Height = 30, 1
	rootBox1.Children = []*paint.PaintableBox{box1, box2}

	planes1 := paint.NewPaintablePlanes()
	planes1.AddToLayer(paint.RenderLayerBase, rootBox1)

	err := engine.PaintPaintablePlanes(planes1, buffer)
	if err != nil {
		t.Fatalf("Frame 1 error = %v", err)
	}

	// Verify Frame 1
	if buffer.GetContent(10, 5).Cluster != "[" {
		t.Errorf("Frame 1: Expected '[' at (10,5), got '%s'", buffer.GetContent(10, 5).Cluster)
	}
	if buffer.GetContent(20, 5).Cluster != "*" {
		t.Errorf("Frame 1: Expected '*' at (20,5), got '%s'", buffer.GetContent(20, 5).Cluster)
	}

	// Frame 2: Focus shifts to first button
	// First button IS focused - "*[  +  ]"
	textNode3 := NewMockPaintableTextNode("btn-1-2", "*[  +  ]")
	box1_2 := paint.NewPaintableBoxWithBounds(textNode3, 10, 5, 7, 1)

	// Second button NOT focused - "[  -  ]"
	textNode4 := NewMockPaintableTextNode("btn-2-2", "[  -  ]")
	box2_2 := paint.NewPaintableBoxWithBounds(textNode4, 20, 5, 7, 1)

	rootBox2 := paint.NewPaintableBox(textNode3)
	rootBox2.X, rootBox2.Y = 0, 0
	rootBox2.Width, rootBox2.Height = 30, 1
	rootBox2.Children = []*paint.PaintableBox{box1_2, box2_2}

	planes2 := paint.NewPaintablePlanes()
	planes2.AddToLayer(paint.RenderLayerBase, rootBox2)

	// Use the same engine instance to enable smart buffer clearing
	err = engine.PaintPaintablePlanes(planes2, buffer)
	if err != nil {
		t.Fatalf("Frame 2 error = %v", err)
	}

	// Verify Frame 2
	if buffer.GetContent(10, 5).Cluster != "*" {
		t.Errorf("Frame 2: Expected '*' at (10,5), got '%s'", buffer.GetContent(10, 5).Cluster)
	}
	if buffer.GetContent(20, 5).Cluster != "[" {
		t.Errorf("Frame 2: Expected '[' at (20,5), got '%s'", buffer.GetContent(20, 5).Cluster)
	}

	// BUG CHECK: Check for any characters outside the expected range
	for x := 0; x < 30; x++ {
		cell := buffer.GetContent(x, 5)
		isExpected := (x >= 10 && x <= 16) || (x >= 20 && x <= 26)
		if isExpected {
			// Should be within expected range
			continue
		}
		if cell.Cluster != "" && cell.Cluster != " " {
			t.Errorf("BUG DETECTED: Unexpected '%s' at (%d,5), should be empty",
				cell.Cluster, x)
		}
	}

	t.Log("Frame 2: ✓ Verified focus shift")
}

// =============================================================================
// Test Case 3: Component Removal - Content should be removed
// =============================================================================

// TestPaintEngine_ComponentRemoval_BugReproduction reproduces the bug when
// a component is removed, leaving its content in the buffer.
func TestPaintEngine_ComponentRemoval_BugReproduction(t *testing.T) {
	engine := NewPaintEngine()
	buffer := paint.NewBuffer(80, 25)

	// Frame 1: Render a component
	textNode := NewMockPaintableTextNode("removable", "DELETE ME")
	box := paint.NewPaintableBoxWithBounds(textNode, 5, 10, 15, 1)
	layout1 := paint.NewPaintableLayout(box)

	err := engine.PaintLayout(layout1, buffer)
	if err != nil {
		t.Fatalf("Frame 1 error = %v", err)
	}

	// Verify Frame 1
	txt := buffer.GetContent(5, 10).Cluster
	if txt != "D" {
		t.Errorf("Frame 1: Expected 'D' at (5,10), got '%s'", txt)
	}

	// Frame 2: Remove component (render empty layout)
	// In production, component removed from tree
	engine2 := NewPaintEngine()  // Fresh engine

	emptyNode := NewMockPaintableTextNode("empty", "")
	emptyBox := paint.NewPaintableBoxWithBounds(emptyNode, 0, 0, 0, 0)
	layout2 := paint.NewPaintableLayout(emptyBox)

	err = engine2.PaintLayout(layout2, buffer)
	if err != nil {
		t.Fatalf("Frame 2 error = %v", err)
	}

	// BUG CHECK: Old content should be removed
	txt = buffer.GetContent(5, 10).Cluster
	if txt != "" && txt != " " {
		// BUG DETECTED: Deleted component content still visible
		t.Errorf("BUG DETECTED: Removed component at (5,10) still shows '%s', should be empty",
			txt)
	}

	// Check entire row 10 for any leftover
	for x := 0; x < 15; x++ {
		cell := buffer.GetContent(x, 10)
		if cell.Cluster != "" && cell.Cluster != " " {
			t.Errorf("BUG DETECTED: Row 10 has leftover '%s' at column %d, should be empty",
				cell.Cluster, x)
		}
	}

	t.Log("Frame 2: ✓ Verified removal zone is clear")
}

// =============================================================================
// Test Case 4: Planes - Multi-layer rendering bug
// =============================================================================

// TestPaintEngine_Planes_BufferWipeout tests buffer cleaning with PaintablePlanes
func TestPaintEngine_Planes_BufferWipeout(t *testing.T) {
	engine := NewPaintEngine()
	buffer := paint.NewBuffer(80, 25)

	// Frame 1: Paint content
	textNode1 := NewMockPaintableTextNode("plane-1", "LONG TEXT that should be cleared")
	box1 := paint.NewPaintableBoxWithBounds(textNode1, 0, 0, 50, 1)

	planes1 := paint.NewPaintablePlanes()
	planes1.AddToLayer(paint.RenderLayerBase, box1)

	err := engine.PaintPaintablePlanes(planes1, buffer)
	if err != nil {
		t.Fatalf("Frame 1 error = %v", err)
	}

	// Frame 2: Paint shorter content at same position
	textNode2 := NewMockPaintableTextNode("plane-1", "SHORT")  // Same ID!
	box2 := paint.NewPaintableBoxWithBounds(textNode2, 0, 0, 20, 1)

	planes2 := paint.NewPaintablePlanes()
	planes2.AddToLayer(paint.RenderLayerBase, box2)

	// Use the same engine instance to enable smart buffer clearing
	err = engine.PaintPaintablePlanes(planes2, buffer)
	if err != nil {
		t.Fatalf("Frame 2 error = %v", err)
	}

	// Check for leftover characters
	expectedText := "SHORT"
	for i := 0; i < 50; i++ {
		cell := buffer.GetContent(i, 0)

		var expectedChar rune
		if i < len(expectedText) {
			expectedChar = []rune(expectedText)[i]
		} else {
			expectedChar = ' '  // Should be space after text ends
		}

		if cell.Cluster != string(expectedChar) {
			t.Errorf("Plane test: BUG at (%d,0): Expected '%c', got '%s'",
				i, expectedChar, cell.Cluster)
		}
	}

	t.Log("Plane test: ✓ Short text painted, verifying no leftovers...")
}

// =============================================================================
// Helper Functions
// =============================================================================

// countRowNonEmpty counts non-empty cells in a row
func countRowNonEmpty(buffer *paint.Buffer, y int) int {
	count := 0
	for x := 0; x < buffer.Width; x++ {
		cell := buffer.GetContent(x, y)
		if cell.Cluster != "" && cell.Cluster != " " {
			count++
		}
	}
	return count
}
