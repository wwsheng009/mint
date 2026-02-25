package paint

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/style"
)

// =============================================================================
// Mock PaintableNode for Testing
// =============================================================================

type mockPaintableNode struct {
	id       string
	nodeType NodeType
	tag      string
	style    style.Style
	text     string
}

func newMockPaintableNode(id, tag string) *mockPaintableNode {
	return &mockPaintableNode{
		id:       id,
		nodeType: NodeTypeElement,
		tag:      tag,
		style:    style.Style{},
		text:     "",
	}
}

func (m *mockPaintableNode) ID() string               { return m.id }
func (m *mockPaintableNode) NodeType() NodeType      { return m.nodeType }
func (m *mockPaintableNode) Tag() string             { return m.tag }
func (m *mockPaintableNode) Style() style.Style      { return m.style }
func (m *mockPaintableNode) SetStyle(s style.Style)  { m.style = s }
func (m *mockPaintableNode) TextContent() string     { return m.text }
func (m *mockPaintableNode) Paint(x, y int) []DrawCmd { return nil }

// =============================================================================
// Basic Configuration Tests
// =============================================================================

func TestPaintableBox_New(t *testing.T) {
	node := newMockPaintableNode("test-node", "div")
	box := NewPaintableBox(node)

	if box == nil {
		t.Fatal("NewPaintableBox returned nil")
	}
	if box.Node == nil {
		t.Error("Node should not be nil")
	}
	if box.NodeID != 0 {
		t.Errorf("NodeID should be 0, got %d", box.NodeID)
	}
	if box.Layer != 0 {
		t.Errorf("Layer should be 0, got %d", box.Layer)
	}
	if box.ZIndex != 0 {
		t.Errorf("ZIndex should be 0, got %d", box.ZIndex)
	}
	if len(box.Children) != 0 {
		t.Errorf("Children should be empty, got %d", len(box.Children))
	}
}

func TestPaintableBox_NewWithBounds(t *testing.T) {
	node := newMockPaintableNode("test-node", "div")
	box := NewPaintableBoxWithBounds(node, 10, 20, 30, 40)

	if box.X != 10 {
		t.Errorf("X should be 10, got %d", box.X)
	}
	if box.Y != 20 {
		t.Errorf("Y should be 20, got %d", box.Y)
	}
	if box.Width != 30 {
		t.Errorf("Width should be 30, got %d", box.Width)
	}
	if box.Height != 40 {
		t.Errorf("Height should be 40, got %d", box.Height)
	}
}

func TestPaintableBox_Bounds(t *testing.T) {
	node := newMockPaintableNode("test-node", "div")
	box := NewPaintableBoxWithBounds(node, 5, 10, 15, 20)

	x, y, w, h := box.Bounds()
	if x != 5 || y != 10 || w != 15 || h != 20 {
		t.Errorf("Bounds() returned (%d, %d, %d, %d), expected (5, 10, 15, 20)", x, y, w, h)
	}
}

func TestPaintableBox_SetBounds(t *testing.T) {
	node := newMockPaintableNode("test-node", "div")
	box := NewPaintableBox(node)

	box.SetBounds(100, 200, 300, 400)
	if box.X != 100 || box.Y != 200 || box.Width != 300 || box.Height != 400 {
		t.Errorf("After SetBounds: got (%d, %d, %d, %d), expected (100, 200, 300, 400)",
			box.X, box.Y, box.Width, box.Height)
	}
}

func TestPaintableBox_Rect(t *testing.T) {
	node := newMockPaintableNode("test-node", "div")
	box := NewPaintableBoxWithBounds(node, 10, 20, 30, 40)

	rect := box.Rect()
	if rect.X != 10 || rect.Y != 20 || rect.Width != 30 || rect.Height != 40 {
		t.Errorf("Rect() returned %+v, expected Rect{X:10, Y:20, W:30, H:40}", rect)
	}
}

// =============================================================================
// Tree Structure Tests
// =============================================================================

func TestPaintableBox_AddChild(t *testing.T) {
	parentNode := newMockPaintableNode("parent", "div")
	childNode := newMockPaintableNode("child", "span")

	parent := NewPaintableBox(parentNode)
	child := NewPaintableBox(childNode)

	parent.AddChild(child)

	if len(parent.Children) != 1 {
		t.Errorf("Parent should have 1 child, got %d", len(parent.Children))
	}
	if parent.Children[0] != child {
		t.Error("Child should be the one added")
	}
	if child.Parent != parent {
		t.Error("Child's parent reference should be set to parent")
	}
}

func TestPaintableBox_AddChild_Nil(t *testing.T) {
	parentNode := newMockPaintableNode("parent", "div")
	parent := NewPaintableBox(parentNode)

	// Adding nil child should not panic
	parent.AddChild(nil)

	if len(parent.Children) != 0 {
		t.Errorf("Parent should have 0 children after adding nil, got %d", len(parent.Children))
	}
}

func TestPaintableBox_AddChild_Multiple(t *testing.T) {
	parentNode := newMockPaintableNode("parent", "div")
	parent := NewPaintableBox(parentNode)

	for i := 0; i < 5; i++ {
		child := NewPaintableBox(newMockPaintableNode("child", "span"))
		parent.AddChild(child)
	}

	if len(parent.Children) != 5 {
		t.Errorf("Parent should have 5 children, got %d", len(parent.Children))
	}

	// Verify all children have correct parent reference
	for i, child := range parent.Children {
		if child.Parent != parent {
			t.Errorf("Child %d parent reference is incorrect", i)
		}
	}
}

func TestPaintableBox_Depth_Root(t *testing.T) {
	node := newMockPaintableNode("root", "div")
	box := NewPaintableBox(node)

	if depth := box.Depth(); depth != 0 {
		t.Errorf("Root depth should be 0, got %d", depth)
	}
}

func TestPaintableBox_Depth_Nested(t *testing.T) {
	root := NewPaintableBox(newMockPaintableNode("root", "div"))
	level1 := NewPaintableBox(newMockPaintableNode("l1", "div"))
	level2 := NewPaintableBox(newMockPaintableNode("l2", "span"))
	level3 := NewPaintableBox(newMockPaintableNode("l3", "text"))

	root.AddChild(level1)
	level1.AddChild(level2)
	level2.AddChild(level3)

	if depth := root.Depth(); depth != 0 {
		t.Errorf("Root depth should be 0, got %d", depth)
	}
	if depth := level1.Depth(); depth != 1 {
		t.Errorf("Level 1 depth should be 1, got %d", depth)
	}
	if depth := level2.Depth(); depth != 2 {
		t.Errorf("Level 2 depth should be 2, got %d", depth)
	}
	if depth := level3.Depth(); depth != 3 {
		t.Errorf("Level 3 depth should be 3, got %d", depth)
	}
}

func TestPaintableBox_Count_Single(t *testing.T) {
	node := newMockPaintableNode("test", "div")
	box := NewPaintableBox(node)

	if count := box.Count(); count != 1 {
		t.Errorf("Single box count should be 1, got %d", count)
	}
}

func TestPaintableBox_Count_Tree(t *testing.T) {
	root := NewPaintableBox(newMockPaintableNode("root", "div"))

	// Add children
	for i := 0; i < 3; i++ {
		child := NewPaintableBox(newMockPaintableNode("child", "span"))
		root.AddChild(child)

		// Add grandchildren
		for j := 0; j < 2; j++ {
			grandchild := NewPaintableBox(newMockPaintableNode("grandchild", "text"))
			child.AddChild(grandchild)
		}
	}

	// Root + 3 children + 6 grandchildren = 10
	if count := root.Count(); count != 10 {
		t.Errorf("Tree count should be 10, got %d", count)
	}

	// Child count should be 1 (itself) + 2 grandchildren = 3
	if count := root.Children[0].Count(); count != 3 {
		t.Errorf("Child count should be 3, got %d", count)
	}
}

// =============================================================================
// Search and Position Tests
// =============================================================================

func TestPaintableBox_FindByPosition_NoMatch(t *testing.T) {
	node := newMockPaintableNode("test", "div")
	box := NewPaintableBoxWithBounds(node, 10, 10, 20, 20)

	// Position outside the box
	if found := box.FindByPosition(0, 0); found != nil {
		t.Error("Should not find a box outside the bounds")
	}
	if found := box.FindByPosition(100, 100); found != nil {
		t.Error("Should not find a box far outside the bounds")
	}
}

func TestPaintableBox_FindByPosition_InBox(t *testing.T) {
	node := newMockPaintableNode("test", "div")
	box := NewPaintableBoxWithBounds(node, 10, 10, 20, 20)

	// Position inside the box
	if found := box.FindByPosition(15, 15); found != box {
		t.Error("Should find the box itself")
	}
	if found := box.FindByPosition(10, 10); found != box {
		t.Error("Should find the box at top-left corner")
	}
	if found := box.FindByPosition(29, 29); found != box {
		t.Error("Should find the box at bottom-right corner (inclusive)")
	}

	// Position on edges (should be inside)
	if found := box.FindByPosition(20, 15); found != box {
		t.Error("Should find the box at middle edge")
	}
}

func TestPaintableBox_FindByPosition_Tree(t *testing.T) {
	root := NewPaintableBoxWithBounds(newMockPaintableNode("root", "div"), 0, 0, 100, 100)
	child1 := NewPaintableBoxWithBounds(newMockPaintableNode("child1", "div"), 10, 10, 30, 30)
	child2 := NewPaintableBoxWithBounds(newMockPaintableNode("child2", "div"), 40, 10, 30, 30)
	grandchild := NewPaintableBoxWithBounds(newMockPaintableNode("grandchild", "span"), 15, 15, 10, 10)

	root.AddChild(child1)
	root.AddChild(child2)
	child1.AddChild(grandchild)

	// Find in grandchild
	if found := root.FindByPosition(17, 17); found != grandchild {
		t.Errorf("Should find grandchild, got %+v", found)
	}

	// Find in child2 (no children overlap)
	if found := root.FindByPosition(45, 20); found != child2 {
		t.Errorf("Should find child2, got %+v", found)
	}

	// Find in root (no children match)
	if found := root.FindByPosition(5, 5); found != root {
		t.Errorf("Should find root, got %+v", found)
	}
}

func TestPaintableBox_Contains(t *testing.T) {
	node := newMockPaintableNode("test", "div")
	box := NewPaintableBoxWithBounds(node, 10, 10, 20, 20)

	testCases := []struct {
		x, y     int
		expected bool
	}{
		{5, 5, false},    // Outside, top-left
		{10, 10, true},   // Top-left corner
		{15, 15, true},   // Inside
		{29, 10, true},   // Right edge
		{10, 29, true},   // Bottom edge
		{29, 29, true},   // Bottom-right corner
		{30, 29, false},  // Right of box
		{29, 30, false},  // Below box
		{0, 0, false},    // Far outside
	}

	for _, tc := range testCases {
		if result := box.Contains(tc.x, tc.y); result != tc.expected {
			t.Errorf("Contains(%d, %d) = %v, expected %v", tc.x, tc.y, result, tc.expected)
		}
	}
}

func TestPaintableBox_ContainsRect(t *testing.T) {
	node := newMockPaintableNode("test", "div")
	box := NewPaintableBoxWithBounds(node, 10, 10, 100, 100)

	testCases := []struct {
		rect     Rect
		expected bool
		desc     string
	}{
		{Rect{X: 15, Y: 15, Width: 10, Height: 10}, true, "Small rect inside"},
		{Rect{X: 10, Y: 10, Width: 90, Height: 90}, true, "Rect touching edges"},
		{Rect{X: 8, Y: 15, Width: 10, Height: 10}, false, "Rect extends left"},
		{Rect{X: 105, Y: 15, Width: 10, Height: 10}, false, "Rect extends right"},
		{Rect{X: 15, Y: 8, Width: 10, Height: 10}, false, "Rect extends top"},
		{Rect{X: 15, Y: 105, Width: 10, Height: 10}, false, "Rect extends bottom"},
		{Rect{X: 0, Y: 0, Width: 5, Height: 5}, false, "Rect far outside"},
	}

	for _, tc := range testCases {
		if result := box.ContainsRect(tc.rect); result != tc.expected {
			t.Errorf("%s: ContainsRect(%+v) = %v, expected %v", tc.desc, tc.rect, result, tc.expected)
		}
	}
}

func TestPaintableBox_Intersects(t *testing.T) {
	node1 := newMockPaintableNode("box1", "div")
	box1 := NewPaintableBoxWithBounds(node1, 10, 10, 30, 30)

	testCases := []struct {
		otherX, otherY, otherW, otherH int
		expected                       bool
		desc                           string
	}{
		{0, 0, 20, 20, true, "Overlap top-left"},
		{25, 25, 30, 30, true, "Overlap bottom-right"},
		{0, 0, 5, 5, false, "Touch top-left corner"},
		{40, 10, 20, 20, false, "Touch right edge"},
		{10, 40, 20, 20, false, "Touch bottom edge"},
		{50, 50, 10, 10, false, "Far away"},
		{15, 15, 20, 20, true, "Inside"},
	}

	for _, tc := range testCases {
		other := NewPaintableBoxWithBounds(newMockPaintableNode("other", "div"),
			tc.otherX, tc.otherY, tc.otherW, tc.otherH)
		if result := box1.Intersects(other); result != tc.expected {
			t.Errorf("%s: Intersects(%d,%d,%d,%d) = %v, expected %v",
				tc.desc, tc.otherX, tc.otherY, tc.otherW, tc.otherH, result, tc.expected)
		}
	}

	// Test with nil
	if result := box1.Intersects(nil); result != false {
		t.Error("Intersects(nil) should return false")
	}
}

func TestPaintableBox_FindByID(t *testing.T) {
	root := NewPaintableBoxWithBounds(newMockPaintableNode("root", "div"), 0, 0, 100, 100)
	root.NodeID = 1

	child1 := NewPaintableBoxWithBounds(newMockPaintableNode("child1", "div"), 10, 10, 30, 30)
	child1.NodeID = 2

	child2 := NewPaintableBoxWithBounds(newMockPaintableNode("child2", "div"), 40, 10, 30, 30)
	child2.NodeID = 3

	root.AddChild(child1)
	root.AddChild(child2)

	// Find existing IDs
	if found := root.FindByID(1); found != root {
		t.Error("Should find root by ID 1")
	}
	if found := root.FindByID(2); found != child1 {
		t.Error("Should find child1 by ID 2")
	}
	if found := root.FindByID(3); found != child2 {
		t.Error("Should find child2 by ID 3")
	}

	// Find non-existent ID
	if found := root.FindByID(999); found != nil {
		t.Error("Should return nil for non-existent ID")
	}
}

func TestPaintableBox_FindByNodeID(t *testing.T) {
	node1 := newMockPaintableNode("node-1", "div")
	node2 := newMockPaintableNode("node-2", "span")
	node3 := newMockPaintableNode("node-3", "text")

	root := NewPaintableBox(node1)
	child := NewPaintableBox(node2)
	grandchild := NewPaintableBox(node3)

	root.AddChild(child)
	child.AddChild(grandchild)

	// Find by Node's ID() method
	if found := root.FindByNodeID("node-1"); found != root {
		t.Error("Should find root by node ID 'node-1'")
	}
	if found := root.FindByNodeID("node-2"); found != child {
		t.Error("Should find child by node ID 'node-2'")
	}
	if found := root.FindByNodeID("node-3"); found != grandchild {
		t.Error("Should find grandchild by node ID 'node-3'")
	}

	// Find non-existent ID
	if found := root.FindByNodeID("non-existent"); found != nil {
		t.Error("Should return nil for non-existent node ID")
	}
}

// =============================================================================
// Dirty State Tests
// =============================================================================

func TestPaintableBox_MarkDirty(t *testing.T) {
	root := NewPaintableBox(newMockPaintableNode("root", "div"))
	level1 := NewPaintableBox(newMockPaintableNode("l1", "div"))
	level2 := NewPaintableBox(newMockPaintableNode("l2", "div"))
	level3 := NewPaintableBox(newMockPaintableNode("l3", "div"))

	root.AddChild(level1)
	level1.AddChild(level2)
	level2.AddChild(level3)

	// Initially clean
	if root.LayoutDirty || level1.LayoutDirty || level2.LayoutDirty || level3.LayoutDirty {
		t.Error("All should initially be clean")
	}

	// Mark level3 dirty - should propagate to root
	level3.MarkDirty()

	if !root.LayoutDirty {
		t.Error("Root should be marked dirty")
	}
	if !level1.LayoutDirty {
		t.Error("Level1 should be marked dirty")
	}
	if !level2.LayoutDirty {
		t.Error("Level2 should be marked dirty")
	}
	if !level3.LayoutDirty {
		t.Error("Level3 should be marked dirty")
	}
}

func TestPaintableBox_Mark_DoubleMark(t *testing.T) {
	root := NewPaintableBox(newMockPaintableNode("root", "div"))
	child := NewPaintableBox(newMockPaintableNode("child", "div"))

	root.AddChild(child)

	// Mark root dirty first
	root.MarkDirty()
	if !root.LayoutDirty {
		t.Error("Root should be dirty")
	}

	// Mark child dirty - should not affect root since it's already dirty
	child.MarkDirty()

	if !root.LayoutDirty {
		t.Error("Root should still be dirty")
	}
	if !child.LayoutDirty {
		t.Error("Child should be dirty")
	}
}

func TestPaintableBox_ClearDirty(t *testing.T) {
	root := NewPaintableBox(newMockPaintableNode("root", "div"))
	level1 := NewPaintableBox(newMockPaintableNode("l1", "div"))
	level2 := NewPaintableBox(newMockPaintableNode("l2", "div"))

	root.AddChild(level1)
	level1.AddChild(level2)

	// Mark all dirty
	root.MarkDirty()
	level2.MarkDirty()

	if !root.LayoutDirty || !level1.LayoutDirty || !level2.LayoutDirty {
		t.Error("All should be dirty")
	}

	// Clear from root
	root.ClearDirty()

	if root.LayoutDirty || level1.LayoutDirty || level2.LayoutDirty {
		t.Error("All should be clean after ClearDirty()")
	}
}

func TestPaintableBox_ClearDirty_SingleNode(t *testing.T) {
	node := newMockPaintableNode("test", "div")
	box := NewPaintableBox(node)

	// Clear clean box should not panic
	box.ClearDirty()
	if box.LayoutDirty {
		t.Error("Box should not be dirty after clearing")
	}
}

// =============================================================================
// Border Tests
// =============================================================================

func TestPaintableBox_HasBorder(t *testing.T) {
	node := newMockPaintableNode("test", "div")

	// No border
	box1 := NewPaintableBox(node)
	if box1.HasBorder() {
		t.Error("Box with BorderStyleNone should not have border")
	}

	// Single border
	boxStyleSingle := NewPaintableBox(node)
	boxStyleSingle.BorderStyle = BorderStyleSingle
	if !boxStyleSingle.HasBorder() {
		t.Error("Box with BorderStyleSingle should have border")
	}

	// Double border
	boxStyleDouble := NewPaintableBox(node)
	boxStyleDouble.BorderStyle = BorderStyleDouble
	if !boxStyleDouble.HasBorder() {
		t.Error("Box with BorderStyleDouble should have border")
	}
}

func TestPaintableBox_GetBorderInfo_Direct(t *testing.T) {
	node := newMockPaintableNode("test", "div")
	box := NewPaintableBox(node)

	box.BorderStyle = BorderStyleDouble
	box.BorderColor = "cyan"
	box.BorderLabel = "Title"

	style, color, label := box.GetBorderInfo()
	if style != BorderStyleDouble {
		t.Errorf("Style should be BorderStyleDouble, got %v", style)
	}
	if color != "cyan" {
		t.Errorf("Color should be 'cyan', got '%s'", color)
	}
	if label != "Title" {
		t.Errorf("Label should be 'Title', got '%s'", label)
	}
}

func TestPaintableBox_GetBorderInfo_None(t *testing.T) {
	node := newMockPaintableNode("test", "div")
	box := NewPaintableBox(node)

	style, color, label := box.GetBorderInfo()
	if style != BorderStyleNone {
		t.Errorf("Style should be BorderStyleNone, got %v", style)
	}
	if color != "" {
		t.Errorf("Color should be empty, got '%s'", color)
	}
	if label != "" {
		t.Errorf("Label should be empty, got '%s'", label)
	}
}

// =============================================================================
// Clone Tests
// =============================================================================

func TestPaintableBox_Clone(t *testing.T) {
	node := newMockPaintableNode("test", "div")
	box := NewPaintableBoxWithBounds(node, 10, 20, 30, 40)
	box.NodeID = 100
	box.DiffKey = "test-key"
	box.Layer = 2
	box.ZIndex = 5
	box.RenderedText = "Hello"
	box.NaturalWidth = 20
	box.BorderStyle = BorderStyleSingle

	clone := box.Clone()

	if clone == box {
		t.Error("Clone should return a new instance")
	}
	if clone.Node != box.Node {
		t.Error("Node reference should be shared (shallow clone)")
	}
	if clone.X != 10 || clone.Y != 20 || clone.Width != 30 || clone.Height != 40 {
		t.Errorf("Clone bounds incorrect: %+v", clone)
	}
	if clone.NodeID != 100 {
		t.Errorf("Clone NodeID should be 100, got %d", clone.NodeID)
	}
	if clone.DiffKey != "test-key" {
		t.Errorf("Clone DiffKey should be 'test-key', got '%s'", clone.DiffKey)
	}
	if len(clone.Children) != 0 {
		t.Errorf("Clone should have no children (shallow), got %d", len(clone.Children))
	}
	if clone.Parent != nil {
		t.Error("Clone should have nil parent")
	}
}

func TestPaintableBox_CloneDeep(t *testing.T) {
	rootNode := newMockPaintableNode("root", "div")
	childNode := newMockPaintableNode("child", "span")
	grandchildNode := newMockPaintableNode("grandchild", "text")

	root := NewPaintableBoxWithBounds(rootNode, 0, 0, 100, 100)
	root.NodeID = 1

	child := NewPaintableBoxWithBounds(childNode, 10, 10, 50, 50)
	child.NodeID = 2

	grandchild := NewPaintableBoxWithBounds(grandchildNode, 15, 15, 20, 20)
	grandchild.NodeID = 3

	root.AddChild(child)
	child.AddChild(grandchild)

	clone := root.CloneDeep()

	// Verify structure
	if clone == root {
		t.Error("CloneDeep should return a new instance")
	}
	if len(clone.Children) != 1 {
		t.Errorf("Clone should have 1 child, got %d", len(clone.Children))
	}
	if len(clone.Children[0].Children) != 1 {
		t.Errorf("Clone's child should have 1 child, got %d", len(clone.Children[0].Children))
	}

	// Verify parent references
	if clone.Parent != nil {
		t.Error("Clone root should have nil parent")
	}
	if clone.Children[0].Parent != clone {
		t.Error("Clone child's parent should be clone root")
	}
	if clone.Children[0].Children[0].Parent != clone.Children[0] {
		t.Error("Clone grandchild's parent should be clone child")
	}

	// Verify data
	if clone.NodeID != 1 {
		t.Errorf("Clone root NodeID should be 1, got %d", clone.NodeID)
	}
	if clone.Children[0].NodeID != 2 {
		t.Errorf("Clone child NodeID should be 2, got %d", clone.Children[0].NodeID)
	}
	if clone.Children[0].Children[0].NodeID != 3 {
		t.Errorf("Clone grandchild NodeID should be 3, got %d", clone.Children[0].Children[0].NodeID)
	}

	// Verify original is unchanged
	if root.Children[0].Parent != root {
		t.Error("Original child's parent should still be root")
	}

	// Add a child to clone - should not affect original
	newChild := NewPaintableBox(newMockPaintableNode("new", "text"))
	clone.AddChild(newChild)

	if len(root.Children) != 1 {
		t.Errorf("Original root should still have 1 child, got %d", len(root.Children))
	}
	if len(clone.Children) != 2 {
		t.Errorf("Cloned root should have 2 children, got %d", len(clone.Children))
	}
}

// =============================================================================
// PaintableLayout Tests
// =============================================================================

func TestPaintableLayout_New(t *testing.T) {
	rootNode := newMockPaintableNode("root", "div")
	root := NewPaintableBox(rootNode)

	layout := NewPaintableLayout(root)

	if layout.Root != root {
		t.Error("Layout root should be set")
	}
	if layout.HitMap != nil {
		t.Error("HitMap should be nil when not specified")
	}
}

func TestPaintableLayout_NewWithHitMap(t *testing.T) {
	rootNode := newMockPaintableNode("root", "div")
	root := NewPaintableBox(rootNode)

	mockHitMap := "mock-hitmap"
	layout := NewPaintableLayoutWithHitMap(root, mockHitMap)

	if layout.Root != root {
		t.Error("Layout root should be set")
	}
	if layout.HitMap != mockHitMap {
		t.Error("HitMap should be set")
	}
}

func TestPaintableLayout_FindByPosition(t *testing.T) {
	root := NewPaintableBoxWithBounds(newMockPaintableNode("root", "div"), 0, 0, 100, 100)

	layout := NewPaintableLayout(root)

	if found := layout.FindByPosition(50, 50); found != root {
		t.Error("Should find root at (50, 50)")
	}
	if found := layout.FindByPosition(200, 200); found != nil {
		t.Error("Should return nil for position outside layout")
	}
}

func TestPaintableLayout_FindByPosition_NilRoot(t *testing.T) {
	layout := NewPaintableLayout(nil)

	if found := layout.FindByPosition(50, 50); found != nil {
		t.Error("Should return nil when root is nil")
	}
}

func TestPaintableLayout_FindByID(t *testing.T) {
	root := NewPaintableBox(newMockPaintableNode("root", "div"))
	root.NodeID = 123

	layout := NewPaintableLayout(root)

	if found := layout.FindByID(123); found != root {
		t.Error("Should find root by ID")
	}
	if found := layout.FindByID(999); found != nil {
		t.Error("Should return nil for non-existent ID")
	}
}

func TestPaintableLayout_FindByID_NilRoot(t *testing.T) {
	layout := NewPaintableLayout(nil)

	if found := layout.FindByID(123); found != nil {
		t.Error("Should return nil when root is nil")
	}
}

func TestPaintableLayout_Count(t *testing.T) {
	root := NewPaintableBox(newMockPaintableNode("root", "div"))
	for i := 0; i < 5; i++ {
		child := NewPaintableBox(newMockPaintableNode("child", "div"))
		root.AddChild(child)
	}

	layout := NewPaintableLayout(root)

	if count := layout.Count(); count != 6 { // 1 root + 5 children
		t.Errorf("Count should be 6, got %d", count)
	}
}

func TestPaintableLayout_Count_NilRoot(t *testing.T) {
	layout := NewPaintableLayout(nil)

	if count := layout.Count(); count != 0 {
		t.Errorf("Count should be 0 when root is nil, got %d", count)
	}
}

func TestPaintableLayout_IsEmpty(t *testing.T) {
	layout1 := NewPaintableLayout(nil)
	layout2 := NewPaintableLayout(NewPaintableBox(newMockPaintableNode("root", "div")))

	if !layout1.IsEmpty() {
		t.Error("Layout with nil root should be empty")
	}
	if layout2.IsEmpty() {
		t.Error("Layout with root should not be empty")
	}
}

func TestPaintableLayout_Bounds(t *testing.T) {
	root := NewPaintableBoxWithBounds(newMockPaintableNode("root", "div"), 10, 20, 100, 200)
	layout := NewPaintableLayout(root)

	x, y, w, h := layout.Bounds()
	if x != 10 || y != 20 || w != 100 || h != 200 {
		t.Errorf("Bounds should be (10, 20, 100, 200), got (%d, %d, %d, %d)", x, y, w, h)
	}
}

func TestPaintableLayout_Bounds_NilRoot(t *testing.T) {
	layout := NewPaintableLayout(nil)

	x, y, w, h := layout.Bounds()
	if x != 0 || y != 0 || w != 0 || h != 0 {
		t.Errorf("Bounds should be (0, 0, 0, 0) when root is nil, got (%d, %d, %d, %d)", x, y, w, h)
	}
}

// =============================================================================
// PaintableBoxBuilder Tests
// =============================================================================

func TestPaintableBoxBuilder_Build(t *testing.T) {
	node := newMockPaintableNode("test", "div")

	box := NewPaintableBoxBuilder().
		WithNode(node).
		WithBounds(10, 20, 30, 40).
		WithLayer(2).
		WithZIndex(5).
		WithNodeID(123).
		WithDiffKey("test-key").
		WithBorder(BorderStyleDouble, "cyan", "Title").
		WithRenderedText("Hello").
		Build()

	if box == nil {
		t.Fatal("Build should return a PaintableBox")
	}
	if box.Node != node {
		t.Error("Node should be set")
	}
	if box.X != 10 || box.Y != 20 || box.Width != 30 || box.Height != 40 {
		t.Errorf("Bounds should be (10, 20, 30, 40), got (%d, %d, %d, %d)", box.X, box.Y, box.Width, box.Height)
	}
	if box.Layer != 2 {
		t.Errorf("Layer should be 2, got %d", box.Layer)
	}
	if box.ZIndex != 5 {
		t.Errorf("ZIndex should be 5, got %d", box.ZIndex)
	}
	if box.NodeID != 123 {
		t.Errorf("NodeID should be 123, got %d", box.NodeID)
	}
	if box.DiffKey != "test-key" {
		t.Errorf("DiffKey should be 'test-key', got '%s'", box.DiffKey)
	}
	if box.BorderStyle != BorderStyleDouble {
		t.Errorf("BorderStyle should be BorderStyleDouble, got %v", box.BorderStyle)
	}
	if box.BorderColor != "cyan" {
		t.Errorf("BorderColor should be 'cyan', got '%s'", box.BorderColor)
	}
	if box.BorderLabel != "Title" {
		t.Errorf("BorderLabel should be 'Title', got '%s'", box.BorderLabel)
	}
	if box.RenderedText != "Hello" {
		t.Errorf("RenderedText should be 'Hello', got '%s'", box.RenderedText)
	}
}

func TestPaintableBoxBuilder_AddChild(t *testing.T) {
	childNode := newMockPaintableNode("child", "div")

	child := NewPaintableBoxBuilder().WithNode(childNode).Build()

	builtParent := NewPaintableBoxBuilder().AddChild(child).Build()

	if len(builtParent.Children) != 1 {
		t.Errorf("Builder should have 1 child, got %d", len(builtParent.Children))
	}
	if builtParent.Children[0] != child {
		t.Error("Child should be the one added")
	}
	if child.Parent != builtParent {
		t.Error("Child's parent should be set to built parent")
	}
}

func TestPaintableBoxBuilder_MultipleAddChild(t *testing.T) {
	builder := NewPaintableBoxBuilder()

	for i := 0; i < 3; i++ {
		child := NewPaintableBoxBuilder().
			WithNode(newMockPaintableNode("child", "div")).
			Build()
		builder = builder.AddChild(child)
	}

	box := builder.Build()

	if len(box.Children) != 3 {
		t.Errorf("Box should have 3 children, got %d", len(box.Children))
	}

	// Verify parent references
	for i, child := range box.Children {
		if child.Parent != box {
			t.Errorf("Child %d parent reference is incorrect", i)
		}
	}
}

// =============================================================================
// Edge Cases and Boundary Tests
// =============================================================================

func TestPaintableBox_NegativeBounds(t *testing.T) {
	node := newMockPaintableNode("test", "div")

	// Negative positions should be allowed
	box := NewPaintableBoxWithBounds(node, -10, -20, 30, 40)
	if box.X != -10 || box.Y != -20 {
		t.Error("Negative positions should be stored")
	}

	// Negative sizes should be allowed (though unusual)
	box2 := NewPaintableBoxWithBounds(node, 0, 0, -30, -40)
	if box2.Width != -30 || box2.Height != -40 {
		t.Error("Negative sizes should be stored")
	}
}

func TestPaintableBox_ZeroBounds(t *testing.T) {
	node := newMockPaintableNode("test", "div")
	box := NewPaintableBoxWithBounds(node, 0, 0, 0, 0)

	if box.X != 0 || box.Y != 0 || box.Width != 0 || box.Height != 0 {
		t.Error("Zero bounds should be stored correctly")
	}

	// Zero-size box should not contain any point
	if box.Contains(0, 0) {
		t.Error("Zero-size box at (0,0) should not contain (0,0)")
	}
}

func TestPaintableBox_MaxInt_Values(t *testing.T) {
	node := newMockPaintableNode("test", "div")

	box := NewPaintableBoxWithBounds(node, 1<<30, 1<<30, 1<<30, 1<<30)

	if box.X != 1<<30 || box.Y != 1<<30 || box.Width != 1<<30 || box.Height != 1<<30 {
		t.Error("Large values should be stored correctly")
	}
}

func TestPaintableBox_DeepTree(t *testing.T) {
	// Create a deep tree (100 levels)
	root := NewPaintableBox(newMockPaintableNode("root", "div"))
	current := root

	for i := 0; i < 100; i++ {
		child := NewPaintableBox(newMockPaintableNode("level", "div"))
		current.AddChild(child)
		current = child
	}

	// Verify depth
	if depth := root.Depth(); depth != 0 {
		t.Errorf("Root depth should be 0, got %d", depth)
	}
	if depth := current.Depth(); depth != 100 {
		t.Errorf("Leaf depth should be 100, got %d", depth)
	}

	// Verify count (1 root + 100 children)
	if count := root.Count(); count != 101 {
		t.Errorf("Tree count should be 101, got %d", count)
	}
}

func TestPaintableBox_WideTree(t *testing.T) {
	// Create a wide tree (root with many children)
	root := NewPaintableBox(newMockPaintableNode("root", "div"))

	for i := 0; i < 1000; i++ {
		child := NewPaintableBox(newMockPaintableNode("child", "div"))
		root.AddChild(child)
	}

	if len(root.Children) != 1000 {
		t.Errorf("Root should have 1000 children, got %d", len(root.Children))
	}
	if count := root.Count(); count != 1001 { // 1 root + 1000 children
		t.Errorf("Tree count should be 1001, got %d", count)
	}
}

func TestPaintableBox_FindByPosition_ReverseOrder(t *testing.T) {
	// Test that FindByPosition returns the topmost child (last in array)
	root := NewPaintableBoxWithBounds(newMockPaintableNode("root", "div"), 0, 0, 100, 100)

	// Add overlapping children - last one should be "on top"
	child1 := NewPaintableBoxWithBounds(newMockPaintableNode("child1", "div"), 10, 10, 50, 50)
	child2 := NewPaintableBoxWithBounds(newMockPaintableNode("child2", "div"), 20, 20, 50, 50)

	root.AddChild(child1)
	root.AddChild(child2)

	// Find at overlapping position - should return child2 (added last)
	if found := root.FindByPosition(25, 25); found != child2 {
		t.Error("Should find child2 (topmost at overlapping position)")
	}
}

func TestPaintableBox_CloneDeep_WithDirtyState(t *testing.T) {
	root := NewPaintableBox(newMockPaintableNode("root", "div"))
	child := NewPaintableBox(newMockPaintableNode("child", "div"))

	root.AddChild(child)

	// Mark root dirty - this only marks root itself (not children)
	root.MarkDirty()

	clone := root.CloneDeep()

	// Clone should have same dirty state as original
	if !clone.LayoutDirty {
		t.Error("Clone should have LayoutDirty = true")
	}
	if clone.Children[0].LayoutDirty {
		t.Error("Clone's child should NOT have LayoutDirty = true (MarkDirty only marks parent path)")
	}

	// Clear original dirty state should NOT affect clone (they are independent)
	root.ClearDirty()

	if root.LayoutDirty || root.Children[0].LayoutDirty {
		t.Error("Original should be clean after ClearDirty()")
	}

	// Clone should STILL be dirty (independent from original)
	if !clone.LayoutDirty {
		t.Error("Clone should remain dirty even after original is cleared")
	}
	if clone.Children[0].LayoutDirty {
		t.Error("Clone's child should still NOT have LayoutDirty")
	}
}

func TestPaintableBox_MarkDirty_CycleDetection(t *testing.T) {
	root := NewPaintableBox(newMockPaintableNode("root", "div"))
	child := NewPaintableBox(newMockPaintableNode("child", "div"))

	root.AddChild(child)

	// Create a cycle (should not happen in practice but test for safety)
	child.Parent = root
	root.AddChild(child) // child is already a child

	// Marking dirty should still work (should stop when it sees already-marked parent)
	root.LayoutDirty = false
	child.MarkDirty()

	if !root.LayoutDirty {
		t.Error("Root should be marked dirty")
	}
}

// =============================================================================
// Performance Tests
// =============================================================================

func TestPaintableBox_Performance_FindByPosition_DeepTree(t *testing.T) {
	// Build a tree with many levels
	root := NewPaintableBoxWithBounds(newMockPaintableNode("root", "div"), 0, 0, 1000, 1000)
	current := root

	for i := 0; i < 100; i++ {
		child := NewPaintableBoxWithBounds(newMockPaintableNode("level", "div"), 10, 10+i, 10, 10)
		current.AddChild(child)
		current = child
	}

	// Find should visit each level at most once
	if found := root.FindByPosition(15, 50); found == nil {
		t.Error("Should find a box at (15, 50)")
	}
}

func TestPaintableBox_Performance_FindByID_LargeTree(t *testing.T) {
	// Build a tree with many nodes
	root := NewPaintableBox(newMockPaintableNode("root", "div"))
	root.NodeID = 0

	for i := 1; i < 1000; i++ {
		child := NewPaintableBox(newMockPaintableNode("child", "div"))
		child.NodeID = uint64(i)
		root.AddChild(child)
	}

	// Find first and last IDs
	if found := root.FindByID(0); found != root {
		t.Error("Should find root by ID 0")
	}
	if found := root.FindByID(999); found == nil {
		t.Error("Should find box by ID 999")
	}
}

func TestPaintableBox_Performance_CloneDeep_LargeTree(t *testing.T) {
	// Build a tree with many nodes
	root := NewPaintableBox(newMockPaintableNode("root", "div"))

	for i := 0; i < 100; i++ {
		child := NewPaintableBox(newMockPaintableNode("child", "div"))
		root.AddChild(child)
	}

	// Clone should work without issues
	clone := root.CloneDeep()

	if count := clone.Count(); count != 101 {
		t.Errorf("Cloned tree should have 101 nodes, got %d", count)
	}
}

// =============================================================================
// Rect Type Tests (Edge Cases)
// =============================================================================

func TestRect_Equal(t *testing.T) {
	r1 := Rect{X: 10, Y: 20, Width: 30, Height: 40}
	r2 := Rect{X: 10, Y: 20, Width: 30, Height: 40}
	r3 := Rect{X: 10, Y: 20, Width: 30, Height: 41}

	if r1 != r2 {
		t.Error("Equal rects should be equal")
	}
	if r1 == r3 {
		t.Error("Different rects should not be equal")
	}
}
