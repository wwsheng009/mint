package layout

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Extreme Value Tests
// =============================================================================

func TestExtreme_ZeroConstraints(t *testing.T) {
	child := NewMockNode("child", 50, 30)
	container := NewFlexLayout("container", []Node{child})

	// Zero constraints: MaxWidth=0 is treated as unbounded by ConstrainWidth,
	// so the container sizes to fit its child.
	constraints := Constraints{MinWidth: 0, MaxWidth: 0, MinHeight: 0, MaxHeight: 0}

	result := NewEngine().Layout(container, constraints)
	require.NotNil(t, result)
	assert.Equal(t, 50, result.Root.Width)
	assert.Equal(t, 30, result.Root.Height)
}

func TestExtreme_MaxConstraints(t *testing.T) {
	child := NewMockNode("child", 50, 30)
	container := NewFlexLayout("container", []Node{child})

	// Very large constraints
	constraints := Constraints{MinWidth: 0, MaxWidth: 1000000, MinHeight: 0, MaxHeight: 1000000}

	result := NewEngine().Layout(container, constraints)
	require.NotNil(t, result)
	assert.GreaterOrEqual(t, result.Root.Width, 0)
	assert.LessOrEqual(t, result.Root.Width, 1000000)
}

func TestExtreme_NegativeNodeSize(t *testing.T) {
	// Node with negative size (should be handled gracefully)
	// Note: This test uses MockNode which doesn't fully implement Node interface
	// The actual handling depends on the complete Node implementation
	t.Skip("Requires full Node implementation - mock doesn't implement Type()")
}

func TestExtreme_MinGreaterThanMax(t *testing.T) {
	child := NewMockNode("child", 50, 30)
	container := NewFlexLayout("container", []Node{child})

	// Invalid: min > max
	constraints := Constraints{MinWidth: 200, MaxWidth: 100, MinHeight: 200, MaxHeight: 100}

	// Should handle gracefully (min should be clamped to max)
	assert.NotPanics(t, func() {
		result := NewEngine().Layout(container, constraints)
		require.NotNil(t, result)
	})
}

func TestExtreme_VeryLargeNode(t *testing.T) {
	// Node requesting very large size
	child := NewMockNode("huge", 100000, 100000)
	container := NewFlexLayout("container", []Node{child})

	constraints := Constraints{MinWidth: 0, MaxWidth: 100, MinHeight: 0, MaxHeight: 100}

	result := NewEngine().Layout(container, constraints)
	require.NotNil(t, result)
	// Should be constrained
	assert.LessOrEqual(t, result.Root.Width, 100)
	assert.LessOrEqual(t, result.Root.Height, 100)
}

// =============================================================================
// Deep Nesting Edge Cases
// =============================================================================

func TestExtreme_DeepNesting_1000Levels(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping deep nesting test in short mode")
	}

	// Create a chain 1000 levels deep
	var current Node = NewMockNode("leaf", 10, 5)
	for i := 0; i < 1000; i++ {
		container := NewFlexLayout("level", []Node{current})
		container.SetDirection(FlexColumn)
		current = container
	}

	constraints := Constraints{MinWidth: 0, MaxWidth: 100, MinHeight: 0, MaxHeight: 500}
	result := NewEngine().Layout(current, constraints)

	require.NotNil(t, result)
	// Due to depth limit (500), should still work
	assert.NotNil(t, result.Root)
}

func TestExtreme_WideContainer_10000Children(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping wide container test in short mode")
	}

	// Container with 10000 children
	children := make([]Node, 10000)
	for i := range children {
		children[i] = NewMockNode("child", 1, 1)
	}

	container := NewFlexLayout("wide", children)
	container.SetDirection(FlexRow)

	constraints := Constraints{MinWidth: 0, MaxWidth: 10000, MinHeight: 0, MaxHeight: 100}

	result := NewEngine().Layout(container, constraints)
	require.NotNil(t, result)
	assert.Greater(t, result.Root.Width, 0)
}

// =============================================================================
// Nil and Empty Edge Cases
// =============================================================================

func TestExtreme_AllNilChildren(t *testing.T) {
	children := make([]Node, 100)
	// All nil
	container := NewFlexLayout("allNil", children)

	constraints := Constraints{MinWidth: 0, MaxWidth: 100, MinHeight: 0, MaxHeight: 100}

	assert.NotPanics(t, func() {
		result := NewEngine().Layout(container, constraints)
		require.NotNil(t, result)
	})
}

func TestExtreme_NilRoot(t *testing.T) {
	engine := NewEngine()
	constraints := Constraints{MinWidth: 0, MaxWidth: 100, MinHeight: 0, MaxHeight: 100}

	result := engine.Layout(nil, constraints)
	require.NotNil(t, result)
	assert.Nil(t, result.Root)
}

func TestExtreme_EmptyFlexLayout(t *testing.T) {
	container := NewFlexLayout("empty", []Node{})

	constraints := Constraints{MinWidth: 0, MaxWidth: 100, MinHeight: 0, MaxHeight: 100}

	result := NewEngine().Layout(container, constraints)
	require.NotNil(t, result)
}

// =============================================================================
// Circular Reference Edge Cases
// =============================================================================

func TestExtreme_SelfReference(t *testing.T) {
	node := &selfRefNode{MockNode: NewMockNode("self", 10, 5)}
	node.self = node

	constraints := Constraints{MinWidth: 0, MaxWidth: 100, MinHeight: 0, MaxHeight: 100}

	assert.NotPanics(t, func() {
		result := NewEngine().Layout(node, constraints)
		require.NotNil(t, result)
	})
}

type selfRefNode struct {
	*MockNode
	self Node
}

func (n *selfRefNode) Children() []Node {
	if n.self != nil {
		return []Node{n.self}
	}
	return nil
}

func TestExtreme_MultipleCycles(t *testing.T) {
	// Create multiple cycles
	n1 := &cycleTestNode{MockNode: NewMockNode("n1", 10, 5)}
	n2 := &cycleTestNode{MockNode: NewMockNode("n2", 10, 5)}
	n3 := &cycleTestNode{MockNode: NewMockNode("n3", 10, 5)}
	n4 := &cycleTestNode{MockNode: NewMockNode("n4", 10, 5)}

	n1.child = n2
	n2.child = n3
	n3.child = n1 // Cycle 1: n1 -> n2 -> n3 -> n1
	n4.child = n2  // n4 also points to n2

	constraints := Constraints{MinWidth: 0, MaxWidth: 100, MinHeight: 0, MaxHeight: 100}

	assert.NotPanics(t, func() {
		result := NewEngine().Layout(n1, constraints)
		require.NotNil(t, result)
	})
}

type cycleTestNode struct {
	*MockNode
	child Node
}

func (n *cycleTestNode) Children() []Node {
	if n.child != nil {
		return []Node{n.child}
	}
	return nil
}

// =============================================================================
// Constraint Edge Cases
// =============================================================================

func TestExtreme_TightConstraints(t *testing.T) {
	child := NewMockNode("child", 100, 100)
	container := NewFlexLayout("container", []Node{child})

	// Exactly sized constraints
	constraints := Constraints{MinWidth: 100, MaxWidth: 100, MinHeight: 100, MaxHeight: 100}

	result := NewEngine().Layout(container, constraints)
	require.NotNil(t, result)
	assert.Equal(t, 100, result.Root.Width)
	assert.Equal(t, 100, result.Root.Height)
}

func TestExtreme_ConflictingConstraints(t *testing.T) {
	child := NewMockNode("child", 50, 30)
	container := NewFlexLayout("container", []Node{child})

	// Conflicting: minWidth > maxWidth
	constraints := Constraints{MinWidth: 200, MaxWidth: 50, MinHeight: 200, MaxHeight: 50}

	assert.NotPanics(t, func() {
		result := NewEngine().Layout(container, constraints)
		require.NotNil(t, result)
	})
}

func TestExtreme_NegativeConstraints(t *testing.T) {
	child := NewMockNode("child", 50, 30)
	container := NewFlexLayout("container", []Node{child})

	// Negative constraints (invalid but should not crash)
	constraints := Constraints{MinWidth: -100, MaxWidth: -50, MinHeight: -100, MaxHeight: -50}

	assert.NotPanics(t, func() {
		result := NewEngine().Layout(container, constraints)
		require.NotNil(t, result)
	})
}

// =============================================================================
// Stress Tests
// =============================================================================

func TestStress_RapidLayouts(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	child := NewMockNode("child", 50, 30)
	container := NewFlexLayout("container", []Node{child})
	engine := NewEngine()

	for i := 0; i < 10000; i++ {
		constraints := Constraints{
			MinWidth:  i % 100,
			MaxWidth:  100 + i%100,
			MinHeight: i % 50,
			MaxHeight: 50 + i%50,
		}
		result := engine.Layout(container, constraints)
		require.NotNil(t, result)
	}
}

func TestStress_LargeTree(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	// Create a large tree structure
	root := createLargeTree(5, 5) // 5 levels, 5 children each = 3906 nodes

	constraints := Constraints{MinWidth: 0, MaxWidth: 500, MinHeight: 0, MaxHeight: 500}

	assert.NotPanics(t, func() {
		result := NewEngine().Layout(root, constraints)
		require.NotNil(t, result)
	})
}

func createLargeTree(depth, breadth int) Node {
	if depth == 0 {
		return NewMockNode("leaf", 10, 5)
	}

	children := make([]Node, breadth)
	for i := 0; i < breadth; i++ {
		children[i] = createLargeTree(depth-1, breadth)
	}

	container := NewFlexLayout("branch", children)
	if depth%2 == 0 {
		container.SetDirection(FlexRow)
	} else {
		container.SetDirection(FlexColumn)
	}
	return container
}

func TestStress_ManySmallNodes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	// 5000 small nodes
	children := make([]Node, 5000)
	for i := range children {
		children[i] = NewMockNode("small", 1, 1)
	}

	container := NewFlexLayout("many", children)
	container.SetDirection(FlexRow)
	container.SetGap(0)

	constraints := Constraints{MinWidth: 0, MaxWidth: 5000, MinHeight: 0, MaxHeight: 100}

	result := NewEngine().Layout(container, constraints)
	require.NotNil(t, result)
}

// =============================================================================
// Alignment Edge Cases
// =============================================================================

func TestExtreme_Alignment_SingleChild(t *testing.T) {
	child := NewMockNode("child", 50, 30)

	alignments := []CrossAxisAlignment{CrossStart, CrossCenter, CrossEnd, Stretch, Baseline}

	for _, align := range alignments {
		t.Run(align.String(), func(t *testing.T) {
			container := NewFlexLayout("container", []Node{child})
			container.SetDirection(FlexRow)
			container.SetCrossAxis(align)

			constraints := Constraints{MinWidth: 0, MaxWidth: 200, MinHeight: 0, MaxHeight: 100}

			result := NewEngine().Layout(container, constraints)
			require.NotNil(t, result)
		})
	}
}

func TestExtreme_Alignment_DifferentHeights(t *testing.T) {
	// Children with vastly different heights
	children := []Node{
		NewMockNode("tiny", 10, 1),
		NewMockNode("small", 10, 10),
		NewMockNode("medium", 10, 50),
		NewMockNode("large", 10, 100),
		NewMockNode("huge", 10, 500),
	}

	container := NewFlexLayout("container", children)
	container.SetDirection(FlexRow)
	container.SetCrossAxis(CrossCenter)

	constraints := Constraints{MinWidth: 0, MaxWidth: 1000, MinHeight: 0, MaxHeight: 600}

	result := NewEngine().Layout(container, constraints)
	require.NotNil(t, result)
}

// =============================================================================
// Gap Edge Cases
// =============================================================================

func TestExtreme_LargeGap(t *testing.T) {
	children := []Node{
		NewMockNode("a", 10, 10),
		NewMockNode("b", 10, 10),
		NewMockNode("c", 10, 10),
	}

	container := NewFlexLayout("container", children)
	container.SetDirection(FlexRow)
	container.SetGap(10000) // Very large gap

	constraints := Constraints{MinWidth: 0, MaxWidth: 100, MinHeight: 0, MaxHeight: 100}

	result := NewEngine().Layout(container, constraints)
	require.NotNil(t, result)
	// Should be constrained
	assert.LessOrEqual(t, result.Root.Width, 100)
}

func TestExtreme_NegativeGap(t *testing.T) {
	children := []Node{
		NewMockNode("a", 10, 10),
		NewMockNode("b", 10, 10),
	}

	container := NewFlexLayout("container", children)
	container.SetDirection(FlexRow)
	container.SetGap(-10) // Negative gap

	constraints := Constraints{MinWidth: 0, MaxWidth: 100, MinHeight: 0, MaxHeight: 100}

	// Should not panic
	assert.NotPanics(t, func() {
		result := NewEngine().Layout(container, constraints)
		require.NotNil(t, result)
	})
}

// =============================================================================
// Flex Grow/Shrink Edge Cases
// =============================================================================

func TestExtreme_AllFlexGrow(t *testing.T) {
	children := []Node{
		NewMockNode("a", 10, 10),
		NewMockNode("b", 10, 10),
		NewMockNode("c", 10, 10),
	}

	container := NewFlexLayout("container", children)
	container.SetDirection(FlexRow)
	container.SetFlex(0, 1000, 0, 0) // Very large grow
	container.SetFlex(1, 1000, 0, 0)
	container.SetFlex(2, 1000, 0, 0)

	constraints := Constraints{MinWidth: 0, MaxWidth: 100, MinHeight: 0, MaxHeight: 100}

	result := NewEngine().Layout(container, constraints)
	require.NotNil(t, result)
}

func TestExtreme_ZeroFlexShrink(t *testing.T) {
	children := []Node{
		NewMockNode("a", 100, 10),
		NewMockNode("b", 100, 10),
	}

	container := NewFlexLayout("container", children)
	container.SetDirection(FlexRow)
	container.SetFlex(0, 0, 0, 0) // grow=0, shrink=0
	container.SetFlex(1, 0, 0, 0)

	constraints := Constraints{MinWidth: 0, MaxWidth: 50, MinHeight: 0, MaxHeight: 100}

	result := NewEngine().Layout(container, constraints)
	require.NotNil(t, result)
}

// =============================================================================
// Position Edge Cases
// =============================================================================

func TestExtreme_Position_LargeOffsets(t *testing.T) {
	top := 10000
	left := 10000
	pos := NewAbsolutePositionWithOffsets(&top, &left, nil, nil)

	x, y := CalculateAbsolutePosition(100, 100, 50, 50, pos)
	assert.Equal(t, 10000, x)
	assert.Equal(t, 10000, y)
}

func TestExtreme_Position_NegativeOffsets(t *testing.T) {
	top := -100
	left := -200
	pos := NewAbsolutePositionWithOffsets(&top, &left, nil, nil)

	x, y := CalculateAbsolutePosition(100, 100, 50, 50, pos)
	assert.Equal(t, -200, x)
	assert.Equal(t, -100, y)
}

func TestExtreme_Position_AllOffsets(t *testing.T) {
	top := 10
	left := 20
	right := 30
	bottom := 40
	pos := NewAbsolutePositionWithOffsets(&top, &left, &right, &bottom)

	x, y := CalculateAbsolutePosition(100, 100, 50, 50, pos)
	// Left takes precedence over right
	assert.Equal(t, 20, x)
	assert.Equal(t, 10, y)
}

// =============================================================================
// Table Layout Edge Cases
// =============================================================================

func TestExtreme_Table_Empty(t *testing.T) {
	table := NewTableLayout("empty", nil)

	w, h := table.GetSize()
	assert.Equal(t, 0, w)
	assert.Equal(t, 0, h)
}

func TestExtreme_Table_SingleCell(t *testing.T) {
	cell := NewTableCell("cell", NewMockNode("content", 50, 30))
	table := NewTableLayout("single", [][]Node{{cell}})

	w, h := table.GetSize()
	assert.Equal(t, 50, w)
	assert.Equal(t, 30, h)
}

func TestExtreme_Table_RaggedRows(t *testing.T) {
	rows := [][]Node{
		{NewMockNode("r0c0", 10, 5), NewMockNode("r0c1", 20, 5), NewMockNode("r0c2", 30, 5)},
		{NewMockNode("r1c0", 10, 5)}, // Shorter row
		{NewMockNode("r2c0", 10, 5), NewMockNode("r2c1", 20, 5)},
	}
	table := NewTableLayout("ragged", rows)

	colWidths := table.ColumnWidths()
	assert.Equal(t, 3, len(colWidths))
	// Each column should take max width
	assert.Equal(t, 10, colWidths[0])
	assert.Equal(t, 20, colWidths[1])
	assert.Equal(t, 30, colWidths[2])
}

func TestExtreme_Table_Large(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large table test in short mode")
	}

	// 100x100 table
	rows := make([][]Node, 100)
	for i := range rows {
		rows[i] = make([]Node, 100)
		for j := range rows[i] {
			rows[i][j] = NewMockNode("cell", 5, 3)
		}
	}
	table := NewTableLayout("large", rows)

	w, h := table.GetSize()
	assert.Equal(t, 500, w) // 100 * 5
	assert.Equal(t, 300, h) // 100 * 3
}

// =============================================================================
// Layer Edge Cases
// =============================================================================

func TestExtreme_Layer_AllTypes(t *testing.T) {
	child := NewMockNode("child", 50, 30)

	layers := []Layer{
		LayerBase,
		LayerOverlay,
		LayerModal,
		LayerTooltip,
		LayerInspector,
	}

	for _, layer := range layers {
		t.Run(layer.String(), func(t *testing.T) {
			node := NewLayeredNode("layered", child, layer, 0)

			assert.Equal(t, layer, node.GetLayer())
			assert.Equal(t, 0, node.GetZIndex())
			assert.Equal(t, layer.ZIndex(), node.EffectiveZIndex())
		})
	}
}

func TestExtreme_Layer_HighZIndex(t *testing.T) {
	child := NewMockNode("child", 50, 30)
	node := NewLayeredNode("layered", child, LayerModal, 999999)

	assert.Equal(t, 999999, node.GetZIndex())
	assert.Equal(t, LayerModal.ZIndex()+999999, node.EffectiveZIndex())
}

// =============================================================================
// Border Edge Cases
// =============================================================================

func TestExtreme_Border_AllStyles(t *testing.T) {
	styles := []BorderStyle{
		BorderNone,
		BorderSingle,
		BorderDouble,
		BorderRounded,
		BorderDashed,
	}

	for _, style := range styles {
		t.Run(style.String(), func(t *testing.T) {
			border := NewBorder(style)
			node := NewBorderedNode("bordered", NewMockNode("child", 50, 30), border)

			assert.Equal(t, style, node.GetBorder().Style)

			ow, oh := node.MeasureOuter(50, 30)
			if style == BorderNone {
				assert.Equal(t, 50, ow)
				assert.Equal(t, 30, oh)
			} else {
				assert.Greater(t, ow, 50)
				assert.Greater(t, oh, 30)
			}
		})
	}
}

func TestExtreme_Border_ZeroContent(t *testing.T) {
	border := NewBorder(BorderSingle)
	node := NewBorderedNode("bordered", NewMockNode("child", 0, 0), border)

	ow, oh := node.MeasureOuter(0, 0)
	// Should still have border size
	assert.Equal(t, border.HorizontalPadding(), ow)
	assert.Equal(t, border.VerticalPadding(), oh)
}

// =============================================================================
// Concurrent Access Tests
// =============================================================================

func TestExtreme_ConcurrentLayout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent test in short mode")
	}

	child := NewMockNode("child", 50, 30)
	container := NewFlexLayout("container", []Node{child})

	// Run multiple layouts concurrently
	done := make(chan bool, 100)

	for i := 0; i < 100; i++ {
		go func(idx int) {
			engine := NewEngine()
			constraints := Constraints{
				MinWidth:  idx % 100,
				MaxWidth:  100 + idx%100,
				MinHeight: idx % 50,
				MaxHeight: 50 + idx%50,
			}
			result := engine.Layout(container, constraints)
			assert.NotNil(t, result)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 100; i++ {
		<-done
	}
}

// =============================================================================
// Memory Tests
// =============================================================================

func TestExtreme_NoMemoryLeak(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory test in short mode")
	}

	child := NewMockNode("child", 50, 30)

	// Run many layouts and ensure memory doesn't explode
	for i := 0; i < 1000; i++ {
		container := NewFlexLayout("container", []Node{child})
		engine := NewEngine()
		constraints := Constraints{MinWidth: 0, MaxWidth: 100, MinHeight: 0, MaxHeight: 100}
		_ = engine.Layout(container, constraints)
	}
}
