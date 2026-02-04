package runtime_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/style"
)

func TestBoxConstraints(t *testing.T) {
	// Test creating constraints
	c := runtime.NewBoxConstraints(10, 100, 20, 200)

	assert.Equal(t, 10, c.MinWidth)
	assert.Equal(t, 100, c.MaxWidth)
	assert.Equal(t, 20, c.MinHeight)
	assert.Equal(t, 200, c.MaxHeight)

	// Test IsTight
	tight := runtime.NewBoxConstraints(50, 50, 50, 50)
	assert.True(t, tight.IsTight())

	loose := runtime.NewBoxConstraints(0, 100, 0, 100)
	assert.False(t, loose.IsTight())

	// Test Constrain
	width, height := c.Constrain(150, 250)
	assert.Equal(t, 100, width)  // MaxWidth
	assert.Equal(t, 200, height) // MaxHeight

	width, height = c.Constrain(5, 15)
	assert.Equal(t, 10, width)  // MinWidth
	assert.Equal(t, 20, height) // MinHeight

	// Test Loosen
	looseC := loose.Loosen()
	assert.Equal(t, 0, looseC.MinWidth)
	assert.Equal(t, 0, looseC.MinHeight)
	assert.Equal(t, 100, looseC.MaxWidth)
	assert.Equal(t, 100, looseC.MaxHeight)
}

func TestConstraintsAlias(t *testing.T) {
	// Test that Constraints type alias works
	c := runtime.NewConstraints(80, 24)

	assert.Equal(t, 0, c.MinWidth)
	assert.Equal(t, 80, c.MaxWidth)
	assert.Equal(t, 0, c.MinHeight)
	assert.Equal(t, 24, c.MaxHeight)
}

func TestStyle(t *testing.T) {
	// Test default style
	style := runtime.NewStyle()
	assert.Equal(t, -1, style.Width)
	assert.Equal(t, -1, style.Height)
	assert.Equal(t, float64(0), style.FlexGrow)
	assert.Equal(t, runtime.DirectionRow, style.Direction)
	assert.Equal(t, runtime.AlignStart, style.AlignItems)
	assert.Equal(t, runtime.JustifyStart, style.Justify)
	assert.Equal(t, 0, style.Gap)

	// Test builder pattern
	style = style.WithWidth(100).
		WithHeight(50).
		WithFlexGrow(1.5).
		WithDirection(runtime.DirectionColumn).
		WithAlignItems(runtime.AlignCenter)

	assert.Equal(t, 100, style.Width)
	assert.Equal(t, 50, style.Height)
	assert.Equal(t, float64(1.5), style.FlexGrow)
	assert.Equal(t, runtime.DirectionColumn, style.Direction)
	assert.Equal(t, runtime.AlignCenter, style.AlignItems)
}

func TestInsets(t *testing.T) {
	insets := runtime.NewInsets(1, 2, 3, 4)
	assert.Equal(t, 1, insets.Top)
	assert.Equal(t, 2, insets.Right)
	assert.Equal(t, 3, insets.Bottom)
	assert.Equal(t, 4, insets.Left)
}

func TestCellBuffer(t *testing.T) {
	// Test creating buffer
	buf := runtime.NewCellBuffer(10, 5)

	assert.Equal(t, 10, buf.Width)
	assert.Equal(t, 5, buf.Height)

	// Test default cell
	cell := buf.GetContent(5, 3)
	assert.Equal(t, " ", cell.Cluster)
	assert.Equal(t, 0, cell.ZIndex)

	// Test setting content
	boldStyle := style.NewStyle().Bold(true)
	buf.SetContent(5, 3, 10, 'A', boldStyle, "test-node")

	cell = buf.GetContent(5, 3)
	assert.Equal(t, "A", cell.Cluster)
	assert.Equal(t, 10, cell.ZIndex)
	assert.Equal(t, "test-node", cell.NodeID)
	assert.True(t, cell.Style.IsBold())

	// Test Z-Index (overwrites lower Z-Index)
	buf.SetContent(5, 3, 5, 'B', style.Style{}, "low-node")
	cell = buf.GetContent(5, 3)
	assert.Equal(t, "A", cell.Cluster) // Higher Z-Index wins

	// Test Clear
	buf.Clear()
	cell = buf.GetContent(5, 3)
	assert.Equal(t, " ", cell.Cluster)
	assert.Equal(t, 0, cell.ZIndex)

	// Test String output
	buf.SetContent(0, 0, 0, 'H', style.Style{}, "")
	buf.SetContent(1, 0, 0, 'i', style.Style{}, "")
	buf.SetContent(2, 0, 0, '!', style.Style{}, "")

	buf.SetContent(0, 1, 0, 'B', style.Style{}, "")
	buf.SetContent(1, 1, 0, 'y', style.Style{}, "")
	buf.SetContent(2, 1, 0, 'e', style.Style{}, "")

	str := buf.String()
	assert.Contains(t, str, "Hi!")
	assert.Contains(t, str, "Bye")
}

func TestLayoutNode(t *testing.T) {
	// Test creating node
	style := runtime.NewStyle().WithWidth(100).WithHeight(50)
	node := runtime.NewLayoutNode("test-node", runtime.NodeTypeText, style)

	assert.Equal(t, "test-node", node.ID)
	assert.Equal(t, runtime.NodeTypeText, node.Type)
	assert.Equal(t, 100, node.Style.Width)
	assert.Equal(t, 50, node.Style.Height)
	assert.True(t, node.IsDirty())

	// Test AddChild
	child := runtime.NewLayoutNode("child", runtime.NodeTypeText, runtime.NewStyle())
	node.AddChild(child)

	assert.Equal(t, 1, len(node.Children))
	assert.Equal(t, node, child.Parent)

	// Test MarkDirty
	child.ClearDirty()
	assert.False(t, child.IsDirty())
	node.MarkDirty()
	assert.True(t, child.IsDirty())

	// Test ContainsPoint
	node.X = 10
	node.Y = 20
	node.MeasuredWidth = 30
	node.MeasuredHeight = 40

	assert.True(t, node.ContainsPoint(15, 30))
	assert.True(t, node.ContainsPoint(10, 20))
	assert.False(t, node.ContainsPoint(5, 30))
	assert.False(t, node.ContainsPoint(50, 70))
}

func TestSize(t *testing.T) {
	size := runtime.Size{Width: 100, Height: 50}
	assert.Equal(t, 100, size.Width)
	assert.Equal(t, 50, size.Height)
}

// =============================================================================
// LayoutNode Dirty Flag Tests
// =============================================================================

func TestLayoutNode_MarkLayoutDirty(t *testing.T) {
	parent := runtime.NewLayoutNode("parent", runtime.NodeTypeFlex, runtime.NewStyle())
	child := runtime.NewLayoutNode("child", runtime.NodeTypeText, runtime.NewStyle())
	parent.AddChild(child)

	// Clear all dirty flags
	parent.ClearDirty()
	child.ClearDirty()

	// Mark child as layout dirty
	child.MarkLayoutDirty()

	// Child should be layout dirty
	assert.True(t, child.IsLayoutDirty())
	assert.False(t, child.IsPaintDirty())

	// Parent should also be layout dirty (propagates)
	assert.True(t, parent.IsLayoutDirty())

	// LayoutVersion should be incremented on first call
	assert.Greater(t, child.LayoutVersion, uint32(0))

	// Call again when already dirty - should skip increment (early return)
	versionAfterSecondCall := child.LayoutVersion
	child.MarkLayoutDirty()
	assert.Equal(t, versionAfterSecondCall, child.LayoutVersion)
}

func TestLayoutNode_MarkPaintDirty(t *testing.T) {
	parent := runtime.NewLayoutNode("parent", runtime.NodeTypeFlex, runtime.NewStyle())
	child := runtime.NewLayoutNode("child", runtime.NodeTypeText, runtime.NewStyle())
	parent.AddChild(child)

	// Clear all dirty flags
	parent.ClearDirty()
	child.ClearDirty()

	// Mark child as paint dirty
	child.MarkPaintDirty()

	// Child should be paint dirty
	assert.True(t, child.IsPaintDirty())
	assert.False(t, child.IsLayoutDirty())

	// Parent should NOT be affected (paint dirtiness doesn't propagate)
	assert.False(t, parent.IsPaintDirty())
	assert.False(t, parent.IsLayoutDirty())
}

func TestLayoutNode_IsDirtyFlags(t *testing.T) {
	node := runtime.NewLayoutNode("test", runtime.NodeTypeText, runtime.NewStyle())

	// Initially only layout dirty is true (from constructor)
	assert.True(t, node.IsLayoutDirty())
	assert.False(t, node.IsPaintDirty()) // paintDirty is false initially
	assert.True(t, node.IsDirty())        // But overall dirty because layoutDirty=true

	// Clear layout dirty only
	node.ClearLayoutDirty()
	assert.False(t, node.IsLayoutDirty())
	assert.False(t, node.IsPaintDirty())
	assert.False(t, node.IsDirty())

	// Mark paint dirty only
	node.MarkPaintDirty()
	assert.False(t, node.IsLayoutDirty())
	assert.True(t, node.IsPaintDirty())
	assert.True(t, node.IsDirty())

	// Clear paint dirty only
	node.ClearPaintDirty()
	assert.False(t, node.IsLayoutDirty())
	assert.False(t, node.IsPaintDirty())
	assert.False(t, node.IsDirty())

	// MarkDirty sets both flags
	node.MarkDirty()
	assert.True(t, node.IsLayoutDirty())
	assert.True(t, node.IsPaintDirty())
	assert.True(t, node.IsDirty())
}

func TestLayoutNode_ClearDirty(t *testing.T) {
	node := runtime.NewLayoutNode("test", runtime.NodeTypeText, runtime.NewStyle())

	// Mark as dirty
	node.MarkDirty()
	assert.True(t, node.IsDirty())

	// Clear all dirty flags
	node.ClearDirty()
	assert.False(t, node.IsLayoutDirty())
	assert.False(t, node.IsPaintDirty())
	assert.False(t, node.IsDirty())
}

func TestLayoutNode_ClearDirty_NilNode(t *testing.T) {
	var node *runtime.LayoutNode

	// Should not panic on nil node
	node.ClearDirty()
	node.ClearLayoutDirty()
	node.ClearPaintDirty()

	// Is functions should return false for nil
	assert.False(t, node.IsLayoutDirty())
	assert.False(t, node.IsPaintDirty())
	assert.False(t, node.IsDirty())
}

// =============================================================================
// LayoutNode Priority Tests
// =============================================================================

func TestLayoutNode_Priority(t *testing.T) {
	node := runtime.NewLayoutNode("test", runtime.NodeTypeText, runtime.NewStyle())

	// Default priority is DirtyNormal (1)
	assert.Equal(t, 1, int(node.GetPriority()))

	// Set priority
	node.SetPriority(2)
	assert.Equal(t, 2, int(node.GetPriority()))

	node.SetPriority(0)
	assert.Equal(t, 0, int(node.GetPriority()))
}

func TestLayoutNode_Priority_NilNode(t *testing.T) {
	var node *runtime.LayoutNode

	// Should return DirtyNormal (1) for nil node
	assert.Equal(t, 1, int(node.GetPriority()))

	// Should not panic on nil node
	node.SetPriority(2)
}

// =============================================================================
// LayoutNode Measure Tests
// =============================================================================

func TestLayoutNode_Measure(t *testing.T) {
	// Node without component returns zero size
	node := runtime.NewLayoutNode("test", runtime.NodeTypeText, runtime.NewStyle())
	constraints := runtime.NewBoxConstraints(10, 100, 10, 100)

	size := node.Measure(constraints)
	assert.Equal(t, 0, size.Width)
	assert.Equal(t, 0, size.Height)
}

func TestLayoutNode_Measure_WithComponent(t *testing.T) {
	// Create a mock component
	mockComp := &mockMeasurableComponent{
		width:  50,
		height: 30,
	}

	node := runtime.NewLayoutNode("test", runtime.NodeTypeText, runtime.NewStyle())
	node.Component = runtime.NewComponentRef("test", "mock", mockComp)

	constraints := runtime.NewBoxConstraints(10, 100, 10, 100)
	size := node.Measure(constraints)

	assert.Equal(t, 50, size.Width)
	assert.Equal(t, 30, size.Height)
}

// mockMeasurableComponent is a test component that implements Measure
type mockMeasurableComponent struct {
	width, height int
}

func (m *mockMeasurableComponent) Measure(c runtime.BoxConstraints) runtime.Size {
	return runtime.Size{Width: m.width, Height: m.height}
}

// =============================================================================
// LayoutNode Bounds Tests
// =============================================================================

func TestLayoutNode_GetBounds(t *testing.T) {
	node := runtime.NewLayoutNode("test", runtime.NodeTypeText, runtime.NewStyle())
	node.X = 10
	node.Y = 20
	node.MeasuredWidth = 30
	node.MeasuredHeight = 40

	x, y, w, h := node.GetBounds()
	assert.Equal(t, 10, x)
	assert.Equal(t, 20, y)
	assert.Equal(t, 30, w)
	assert.Equal(t, 40, h)
}

func TestLayoutNode_GetBounds_NilNode(t *testing.T) {
	var node *runtime.LayoutNode
	x, y, w, h := node.GetBounds()
	assert.Equal(t, 0, x)
	assert.Equal(t, 0, y)
	assert.Equal(t, 0, w)
	assert.Equal(t, 0, h)
}

func TestLayoutNode_GetInnerBounds(t *testing.T) {
	node := runtime.NewLayoutNode("test", runtime.NodeTypeText, runtime.NewStyle())
	node.X = 10
	node.Y = 20
	node.MeasuredWidth = 100
	node.MeasuredHeight = 80
	node.Style.Padding = runtime.NewInsets(5, 10, 15, 20) // top, right, bottom, left

	x, y, w, h := node.GetInnerBounds()
	assert.Equal(t, 30, x)  // 10 + 20 (left)
	assert.Equal(t, 25, y)  // 20 + 5 (top)
	assert.Equal(t, 70, w)  // 100 - 20 - 10 (left - right)
	assert.Equal(t, 60, h)  // 80 - 5 - 15 (top - bottom)
}

func TestLayoutNode_GetInnerBounds_NilNode(t *testing.T) {
	var node *runtime.LayoutNode
	x, y, w, h := node.GetInnerBounds()
	assert.Equal(t, 0, x)
	assert.Equal(t, 0, y)
	assert.Equal(t, 0, w)
	assert.Equal(t, 0, h)
}

// =============================================================================
// LayoutNode AddChildren Tests
// =============================================================================

func TestLayoutNode_AddChildren(t *testing.T) {
	parent := runtime.NewLayoutNode("parent", runtime.NodeTypeFlex, runtime.NewStyle())
	child1 := runtime.NewLayoutNode("child1", runtime.NodeTypeText, runtime.NewStyle())
	child2 := runtime.NewLayoutNode("child2", runtime.NodeTypeText, runtime.NewStyle())
	child3 := runtime.NewLayoutNode("child3", runtime.NodeTypeText, runtime.NewStyle())

	parent.AddChildren(child1, child2, child3)

	assert.Equal(t, 3, len(parent.Children))
	assert.Equal(t, parent, child1.Parent)
	assert.Equal(t, parent, child2.Parent)
	assert.Equal(t, parent, child3.Parent)
}

func TestLayoutNode_AddChildren_WithNil(t *testing.T) {
	parent := runtime.NewLayoutNode("parent", runtime.NodeTypeFlex, runtime.NewStyle())
	child1 := runtime.NewLayoutNode("child1", runtime.NodeTypeText, runtime.NewStyle())

	// Should handle nil children gracefully
	parent.AddChildren(child1, nil, (*runtime.LayoutNode)(nil))

	assert.Equal(t, 1, len(parent.Children))
}

// =============================================================================
// LayoutBox Tests
// =============================================================================

func TestNewLayoutBox(t *testing.T) {
	node := runtime.NewLayoutNode("test", runtime.NodeTypeText, runtime.NewStyle())
	node.X = 10
	node.Y = 20
	node.MeasuredWidth = 30
	node.MeasuredHeight = 40
	node.Style.ZIndex = 5

	box := runtime.NewLayoutBox(node)

	assert.Equal(t, "test", box.NodeID)
	assert.Equal(t, 10, box.X)
	assert.Equal(t, 20, box.Y)
	assert.Equal(t, 30, box.W)
	assert.Equal(t, 40, box.H)
	assert.Equal(t, 5, box.ZIndex)
	assert.Equal(t, node, box.Node)
}

func TestNewLayoutBox_NilNode(t *testing.T) {
	// Note: NewLayoutBox panics on nil node
	// This test documents that behavior
	assert.Panics(t, func() {
		runtime.NewLayoutBox(nil)
	})
}

// =============================================================================
// LayoutResult Tests
// =============================================================================

func TestLayoutResult_FindBoxByID(t *testing.T) {
	node1 := runtime.NewLayoutNode("node1", runtime.NodeTypeText, runtime.NewStyle())
	node1.X = 10
	node1.Y = 20
	node1.MeasuredWidth = 30
	node1.MeasuredHeight = 40

	node2 := runtime.NewLayoutNode("node2", runtime.NodeTypeText, runtime.NewStyle())
	node2.X = 50
	node2.Y = 60
	node2.MeasuredWidth = 20
	node2.MeasuredHeight = 30

	box1 := runtime.NewLayoutBox(node1)
	box2 := runtime.NewLayoutBox(node2)

	result := runtime.LayoutResult{
		Boxes: []runtime.LayoutBox{box1, box2},
	}

	// Find existing box
	found := result.FindBoxByID("node1")
	assert.NotNil(t, found)
	assert.Equal(t, "node1", found.NodeID)

	// Find non-existent box
	notFound := result.FindBoxByID("node3")
	assert.Nil(t, notFound)

	// Empty result
	emptyResult := runtime.LayoutResult{}
	assert.Nil(t, emptyResult.FindBoxByID("node1"))
}
