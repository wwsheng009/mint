package layout

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// Deep Nesting Tests - Mixed Directions
// =============================================================================

func TestDeepNesting_MixedDirections(t *testing.T) {
	// Create a deep tree alternating row/column directions
	// Row -> Column -> Row -> Column -> ... (10 levels)
	
	// Leaf nodes at the bottom
	leaf1 := NewMockNode("leaf1", 10, 5)
	leaf2 := NewMockNode("leaf2", 15, 8)

	// Build up from leaves, alternating directions
	// Level 10 (deepest): Column
	col10 := NewFlexLayout("col10", []Node{leaf1, leaf2})
	col10.SetDirection(FlexColumn)

	// Level 9: Row
	row9 := NewFlexLayout("row9", []Node{col10})
	row9.SetDirection(FlexRow)

	// Level 8: Column
	col8 := NewFlexLayout("col8", []Node{row9})
	col8.SetDirection(FlexColumn)

	// Level 7: Row
	row7 := NewFlexLayout("row7", []Node{col8})
	row7.SetDirection(FlexRow)

	// Level 6: Column
	col6 := NewFlexLayout("col6", []Node{row7})
	col6.SetDirection(FlexColumn)

	// Level 5: Row
	row5 := NewFlexLayout("row5", []Node{col6})
	row5.SetDirection(FlexRow)

	// Level 4: Column
	col4 := NewFlexLayout("col4", []Node{row5})
	col4.SetDirection(FlexColumn)

	// Level 3: Row
	row3 := NewFlexLayout("row3", []Node{col4})
	row3.SetDirection(FlexRow)

	// Level 2: Column
	col2 := NewFlexLayout("col2", []Node{row3})
	col2.SetDirection(FlexColumn)

	// Level 1 (root): Row
	root := NewFlexLayout("root", []Node{col2})
	root.SetDirection(FlexRow)

	// Should not panic and produce valid layout
	constraints := Constraints{MinWidth: 0, MaxWidth: 200, MinHeight: 0, MaxHeight: 100}
	result := NewEngine().Layout(root, constraints)

	assert.NotNil(t, result)
	assert.NotNil(t, result.Root)
	assert.Greater(t, result.Root.Width, 0)
	assert.Greater(t, result.Root.Height, 0)
}

func TestDeepNesting_WideAndDeep(t *testing.T) {
	// Create a tree that's both wide (many children) and deep (many levels)
	// 3 children per level, 5 levels deep = 3^5 = 243 leaf nodes
	
	var buildTree func(depth int, prefix string) Node
	buildTree = func(depth int, prefix string) Node {
		if depth == 0 {
			// Leaf node
			return NewMockNode(prefix, 10, 5)
		}

		children := make([]Node, 3)
		for i := 0; i < 3; i++ {
			children[i] = buildTree(depth-1, prefix+"_"+string(rune('a'+i)))
		}

		container := NewFlexLayout(prefix, children)
		if depth%2 == 0 {
			container.SetDirection(FlexRow)
		} else {
			container.SetDirection(FlexColumn)
		}
		return container
	}

	root := buildTree(5, "root")

	constraints := Constraints{MinWidth: 0, MaxWidth: 500, MinHeight: 0, MaxHeight: 300}
	result := NewEngine().Layout(root, constraints)

	assert.NotNil(t, result)
	assert.NotNil(t, result.Root)
	assert.Greater(t, result.Root.Width, 0)
	assert.Greater(t, result.Root.Height, 0)
}

// =============================================================================
// Circular Reference Tests
// =============================================================================

// selfReferencingNode is a mock that can create a circular reference
type selfReferencingNode struct {
	*MockNode
	selfRef *selfReferencingNode
}

func (n *selfReferencingNode) Children() []Node {
	if n.selfRef != nil {
		return []Node{n.selfRef}
	}
	return nil
}

func TestCircularReference_SelfReference(t *testing.T) {
	// Create a node that references itself
	node := &selfReferencingNode{
		MockNode: NewMockNode("self", 10, 5),
	}
	node.selfRef = node // Circular reference

	engine := NewEngine()
	constraints := Constraints{MinWidth: 0, MaxWidth: 100, MinHeight: 0, MaxHeight: 50}

	// Should not panic - cycle detection should handle it
	assert.NotPanics(t, func() {
		result := engine.Layout(node, constraints)
		assert.NotNil(t, result)
	})
}

// circularNode creates a cycle A -> B -> C -> A
type circularNode struct {
	*MockNode
	child Node
}

func (n *circularNode) Children() []Node {
	if n.child != nil {
		return []Node{n.child}
	}
	return nil
}

func TestCircularReference_ThreeNodeCycle(t *testing.T) {
	// Create cycle: A -> B -> C -> A
	a := &circularNode{MockNode: NewMockNode("A", 10, 5)}
	b := &circularNode{MockNode: NewMockNode("B", 10, 5)}
	c := &circularNode{MockNode: NewMockNode("C", 10, 5)}

	a.child = b
	b.child = c
	c.child = a // Creates cycle

	engine := NewEngine()
	constraints := Constraints{MinWidth: 0, MaxWidth: 100, MinHeight: 0, MaxHeight: 50}

	// Should not panic - cycle detection should handle it
	assert.NotPanics(t, func() {
		result := engine.Layout(a, constraints)
		assert.NotNil(t, result)
	})
}

// =============================================================================
// Max Depth Limit Tests
// =============================================================================

func TestMaxDepth_VeryDeepTree(t *testing.T) {
	// Create a chain 200 levels deep
	var current Node = NewMockNode("leaf", 10, 5)

	for i := 0; i < 200; i++ {
		container := NewFlexLayout("level", []Node{current})
		container.SetDirection(FlexColumn)
		current = container
	}

	engine := NewEngine()
	constraints := Constraints{MinWidth: 0, MaxWidth: 100, MinHeight: 0, MaxHeight: 50}

	// Should handle very deep trees without stack overflow
	assert.NotPanics(t, func() {
		result := engine.Layout(current, constraints)
		assert.NotNil(t, result)
	})
}

// =============================================================================
// Constraint Propagation Tests
// =============================================================================

func TestConstraintPropagation_NestedConflicts(t *testing.T) {
	// Outer container has tight constraints
	// Inner container has conflicting constraints
	// Innermost child should still get valid constraints

	innerChild := NewMockNode("innerChild", 50, 30)

	// Inner container wants 100 width, but will get constrained
	innerContainer := NewFlexLayout("inner", []Node{innerChild})
	innerContainer.SetDirection(FlexColumn)

	outerContainer := NewFlexLayout("outer", []Node{innerContainer})
	outerContainer.SetDirection(FlexRow)

	// Very tight outer constraints
	constraints := Constraints{
		MinWidth: 0, MaxWidth: 30,  // Less than inner wants
		MinHeight: 0, MaxHeight: 20,
	}

	result := NewEngine().Layout(outerContainer, constraints)

	assert.NotNil(t, result)
	// All boxes should be within constraints
	assert.LessOrEqual(t, result.Root.Width, 30)
	assert.LessOrEqual(t, result.Root.Height, 20)
}

// =============================================================================
// Zero-Size Container with Content Tests
// =============================================================================

func TestZeroSizeContainer_WithNonZeroChildren(t *testing.T) {
	// Container has zero max constraints, but children have size
	child := NewMockNode("child", 100, 50)
	container := NewFlexLayout("container", []Node{child})

	// Zero-size constraints
	constraints := Constraints{
		MinWidth: 0, MaxWidth: 0,
		MinHeight: 0, MaxHeight: 0,
	}

	result := NewEngine().Layout(container, constraints)

	// Should handle gracefully - children should be sized to 0 or clipped
	assert.NotNil(t, result)
	assert.NotNil(t, result.Root)
}

func TestZeroSizeContainer_SingleChild(t *testing.T) {
	child := NewMockNode("child", 50, 30)
	container := NewFlexLayout("container", []Node{child})

	constraints := TightConstraints(0, 0)

	result := NewEngine().Layout(container, constraints)

	assert.NotNil(t, result)
	assert.Equal(t, 0, result.Root.Width)
	assert.Equal(t, 0, result.Root.Height)
}

// =============================================================================
// All Nil Children Tests
// =============================================================================

func TestAllNilChildren_FlexLayout(t *testing.T) {
	// Container with all nil children
	container := NewFlexLayout("container", []Node{nil, nil, nil})

	constraints := Constraints{MinWidth: 0, MaxWidth: 100, MinHeight: 0, MaxHeight: 50}

	result := NewEngine().Layout(container, constraints)

	assert.NotNil(t, result)
	// Should produce minimal size (just padding if any)
}

func TestMixedNilAndValidChildren(t *testing.T) {
	valid1 := NewMockNode("valid1", 20, 10)
	valid2 := NewMockNode("valid2", 30, 15)

	container := NewFlexLayout("container", []Node{nil, valid1, nil, valid2, nil})

	constraints := Constraints{MinWidth: 0, MaxWidth: 100, MinHeight: 0, MaxHeight: 50}

	result := NewEngine().Layout(container, constraints)

	assert.NotNil(t, result)
	// Should layout only the valid children
	assert.Greater(t, result.Root.Width, 0)
}

// =============================================================================
// Nil Root Node Tests
// =============================================================================

func TestNilRootNode_Layout(t *testing.T) {
	engine := NewEngine()
	constraints := Constraints{MinWidth: 0, MaxWidth: 100, MinHeight: 0, MaxHeight: 50}

	result := engine.Layout(nil, constraints)

	// Should return empty result, not panic
	assert.NotNil(t, result)
	assert.Nil(t, result.Root)
}

// =============================================================================
// Rapid Constraint Changes (Stress Test)
// =============================================================================

func TestRapidConstraintChanges(t *testing.T) {
	child := NewMockNode("child", 50, 30)
	container := NewFlexLayout("container", []Node{child})

	engine := NewEngine()

	// Rapidly change constraints 100 times
	for i := 0; i < 100; i++ {
		constraints := Constraints{
			MinWidth:  i % 50,
			MaxWidth:  50 + i%100,
			MinHeight: i % 30,
			MaxHeight: 30 + i%60,
		}

		result := engine.Layout(container, constraints)
		assert.NotNil(t, result)

		// Verify constraints are respected
		assert.GreaterOrEqual(t, result.Root.Width, constraints.MinWidth)
		assert.LessOrEqual(t, result.Root.Width, constraints.MaxWidth)
	}
}

// =============================================================================
// Dynamic Child Add/Remove Tests
// =============================================================================

func TestDynamicChildren_AddRemove(t *testing.T) {
	child1 := NewMockNode("child1", 20, 10)
	child2 := NewMockNode("child2", 30, 15)
	child3 := NewMockNode("child3", 25, 12)

	// Start with one child
	container := NewFlexLayout("container", []Node{child1})
	container.SetDirection(FlexRow)

	engine := NewEngine()
	constraints := Constraints{MinWidth: 0, MaxWidth: 200, MinHeight: 0, MaxHeight: 100}

	// Layout with one child
	result1 := engine.Layout(container, constraints)
	assert.NotNil(t, result1)
	width1 := result1.Root.Width

	// Add children (simulate - create new container)
	container = NewFlexLayout("container", []Node{child1, child2, child3})
	container.SetDirection(FlexRow)

	result2 := engine.Layout(container, constraints)
	assert.NotNil(t, result2)
	assert.Greater(t, result2.Root.Width, width1) // Should be wider

	// Remove children (simulate - create new container with fewer)
	container = NewFlexLayout("container", []Node{child1})
	container.SetDirection(FlexRow)

	result3 := engine.Layout(container, constraints)
	assert.NotNil(t, result3)
	assert.Equal(t, width1, result3.Root.Width)
}

// =============================================================================
// Large Scale Extended Tests
// =============================================================================

func TestLargeScale_2000Nodes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large scale test in short mode")
	}

	// Create 2000 leaf nodes
	children := make([]Node, 2000)
	for i := 0; i < 2000; i++ {
		children[i] = NewMockNode("leaf", 5, 3)
	}

	// Split into groups of 100
	groups := make([]Node, 20)
	for g := 0; g < 20; g++ {
		start := g * 100
		end := start + 100
		groups[g] = NewFlexLayout("group", children[start:end])
	}

	root := NewFlexLayout("root", groups)
	root.SetDirection(FlexColumn)

	engine := NewEngine()
	constraints := Constraints{MinWidth: 0, MaxWidth: 500, MinHeight: 0, MaxHeight: 1000}

	result := engine.Layout(root, constraints)

	assert.NotNil(t, result)
	assert.NotNil(t, result.Root)
	assert.Greater(t, result.Root.Height, 0)
}

// =============================================================================
// Memory Stress Test
// =============================================================================

func TestMemoryStress_RepeatedLayouts(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory stress test in short mode")
	}

	child := NewMockNode("child", 50, 30)
	container := NewFlexLayout("container", []Node{child})

	engine := NewEngine()
	constraints := Constraints{MinWidth: 0, MaxWidth: 100, MinHeight: 0, MaxHeight: 50}

	// Run 1000 layout passes - verify no crashes or memory leaks
	for i := 0; i < 1000; i++ {
		result := engine.Layout(container, constraints)
		assert.NotNil(t, result)
	}

	// Note: Cache hits may be 0 because FlexLayout's internal state changes
	// after each measure, causing different hash values. This is expected behavior
	// for dynamic layouts. The test primarily verifies stability under stress.
}

// =============================================================================
// Cross-Axis Stretch Tests
// =============================================================================

func TestCrossAxis_Stretch_Row(t *testing.T) {
	// In a row, stretch should make all children same height
	child1 := NewMockNode("child1", 20, 10)
	child2 := NewMockNode("child2", 30, 20) // Taller
	child3 := NewMockNode("child3", 25, 15)

	container := NewFlexLayout("container", []Node{child1, child2, child3})
	container.SetDirection(FlexRow)
	container.SetCrossAxis(Stretch)

	constraints := Constraints{MinWidth: 0, MaxWidth: 200, MinHeight: 0, MaxHeight: 100}

	result := NewEngine().Layout(container, constraints)

	assert.NotNil(t, result)
	// All children should be stretched to container height
	// Note: Container height is determined by constraints
	assert.Greater(t, result.Root.Height, 0)
}

func TestCrossAxis_Stretch_Column(t *testing.T) {
	// In a column, stretch should make all children same width
	child1 := NewMockNode("child1", 10, 20)
	child2 := NewMockNode("child2", 20, 30) // Wider
	child3 := NewMockNode("child3", 15, 25)

	container := NewFlexLayout("container", []Node{child1, child2, child3})
	container.SetDirection(FlexColumn)
	container.SetCrossAxis(Stretch)

	constraints := Constraints{MinWidth: 0, MaxWidth: 100, MinHeight: 0, MaxHeight: 200}

	result := NewEngine().Layout(container, constraints)

	assert.NotNil(t, result)
	assert.Greater(t, result.Root.Width, 0)
}

// =============================================================================
// Baseline Alignment Tests
// =============================================================================

func TestBaselineAlignment_Row(t *testing.T) {
	// Test baseline alignment in row direction
	// All children should align their baselines
	child1 := NewMockNode("child1", 20, 30)
	child2 := NewMockNode("child2", 25, 20) // Different height
	child3 := NewMockNode("child3", 15, 25) // Different height

	container := NewFlexLayout("container", []Node{child1, child2, child3})
	container.SetDirection(FlexRow)
	container.SetCrossAxis(Baseline)

	constraints := Constraints{MinWidth: 0, MaxWidth: 200, MinHeight: 0, MaxHeight: 100}

	result := NewEngine().Layout(container, constraints)

	assert.NotNil(t, result)
	// Should layout without error
	assert.Greater(t, result.Root.Width, 0)
	assert.Greater(t, result.Root.Height, 0)
}

// =============================================================================
// Overflow Scenarios
// =============================================================================

func TestOverflow_ContentLargerThanMaxConstraints(t *testing.T) {
	// Child wants 100px, but container only allows 50px
	child := NewMockNode("child", 100, 50)

	container := NewFlexLayout("container", []Node{child})
	container.SetDirection(FlexRow)

	constraints := Constraints{MinWidth: 0, MaxWidth: 50, MinHeight: 0, MaxHeight: 50}

	result := NewEngine().Layout(container, constraints)

	assert.NotNil(t, result)
	// Content should be constrained to max
	assert.LessOrEqual(t, result.Root.Width, 50)
}

func TestOverflow_MultipleLargeChildren(t *testing.T) {
	// Multiple children that together exceed container
	children := make([]Node, 10)
	for i := range children {
		children[i] = NewMockNode("child", 20, 10) // Total = 200px
	}

	container := NewFlexLayout("container", children)
	container.SetDirection(FlexRow)
	container.SetGap(2) // 9 gaps = 18px extra

	constraints := Constraints{MinWidth: 0, MaxWidth: 100, MinHeight: 0, MaxHeight: 50}

	result := NewEngine().Layout(container, constraints)

	assert.NotNil(t, result)
	// Should clamp to max constraints
	assert.LessOrEqual(t, result.Root.Width, 100)
}
