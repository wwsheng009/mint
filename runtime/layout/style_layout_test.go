package layout

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// Style 影响测试
// =============================================================================

// MockStyleNode 实现 GetSize 从 style 获取尺寸的节点
type MockStyleNode struct {
	*MockNode
	styleWidth  int
	styleHeight int
	styleMargin Margin
	stylePad    Padding
}

func NewMockStyleNode(id string) *MockStyleNode {
	return &MockStyleNode{
		MockNode: NewMockNode(id, 0, 0),
	}
}

func (m *MockStyleNode) SetStyleSize(w, h int) {
	m.styleWidth = w
	m.styleHeight = h
}

func (m *MockStyleNode) SetStyleMargin(top, right, bottom, left int) {
	m.styleMargin = Margin{Top: top, Right: right, Bottom: bottom, Left: left}
}

func (m *MockStyleNode) SetStylePadding(top, right, bottom, left int) {
	m.stylePad = Padding{Top: top, Right: right, Bottom: bottom, Left: left}
}

func (m *MockStyleNode) Measure(constraints Constraints) Size {
	w := constraints.ConstrainWidth(m.styleWidth)
	h := constraints.ConstrainHeight(m.styleHeight)
	return Size{Width: w, Height: h}
}

func (m *MockStyleNode) GetMargin() Margin {
	return m.styleMargin
}

func (m *MockStyleNode) GetPadding() Padding {
	return m.stylePad
}

func (m *MockStyleNode) SetChildren(children []Node) {
	m.children = children
}

// =============================================================================
// Style Width/Height 测试
// =============================================================================

func TestStyle_ExplicitSize(t *testing.T) {
	tests := []struct {
		name          string
		width         int
		height        int
		constraints   Constraints
		expectedW     int
		expectedH     int
	}{
		{"fixed size", 100, 50, UnboundedConstraints(), 100, 50},
		{"constrained smaller", 100, 50, NewConstraints(50, 50, 30, 30), 50, 30},
		{"constrained larger (min)", 100, 50, NewConstraints(150, 150, 80, 80), 150, 80}, // minWidth=150 forces expansion
		{"zero size", 0, 0, UnboundedConstraints(), 0, 0},
		{"negative constraints", 100, 50, NewConstraints(-10, -10, -10, -10), 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := NewMockStyleNode("test")
			node.SetStyleSize(tt.width, tt.height)

			size := node.Measure(tt.constraints)
			assert.Equal(t, tt.expectedW, size.Width)
			assert.Equal(t, tt.expectedH, size.Height)
		})
	}
}

func TestStyle_SizeWithConstraints(t *testing.T) {
	// 测试尺寸在不同约束下的行为
	node := NewMockStyleNode("test")
	node.SetStyleSize(80, 24)

	tests := []struct {
		name        string
		constraints Constraints
		expectW     int
		expectH     int
	}{
		{"unbounded", UnboundedConstraints(), 80, 24},
		{"exact fit", NewConstraints(80, 80, 24, 24), 80, 24},
		{"tight horizontal", NewConstraints(60, 60, 24, 24), 60, 24},
		{"tight vertical", NewConstraints(80, 80, 20, 20), 80, 20},
		{"both tight", NewConstraints(60, 60, 20, 20), 60, 20},
		{"loose (min expansion)", LooseConstraints(100, 50), 100, 50}, // LooseConstraints sets min, expands to min
		{"min only", NewConstraints(50, 200, 10, 100), 80, 24},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			size := node.Measure(tt.constraints)
			assert.Equal(t, tt.expectW, size.Width, "Width mismatch for %s", tt.name)
			assert.Equal(t, tt.expectH, size.Height, "Height mismatch for %s", tt.name)
		})
	}
}

// =============================================================================
// Margin 影响测试
// =============================================================================

func TestMargin_LayoutPosition(t *testing.T) {
	// Margin 影响布局位置
	container := NewMockCompositeNode("container", 100, 100)

	child := NewMockStyleNode("child")
	child.SetStyleSize(50, 50)
	child.SetStyleMargin(10, 10, 10, 10) // all sides 10px margin

	container.SetChildren([]Node{child})

	engine := NewEngine()
	result := engine.Layout(container, UnboundedConstraints())

	assert.NotNil(t, result)
	assert.Len(t, result.Root.Children, 1)

	// 子节点位置应该考虑 margin
	childBox := result.Root.Children[0]
	// 根据布局引擎的实现，margin 可能影响位置
	assert.NotNil(t, childBox)
}

func TestMargin_Asymmetric(t *testing.T) {
	// 非对称 margin
	node := NewMockStyleNode("test")
	node.SetStyleMargin(5, 10, 15, 20) // top, right, bottom, left

	margin := node.GetMargin()
	assert.Equal(t, 5, margin.Top)
	assert.Equal(t, 10, margin.Right)
	assert.Equal(t, 15, margin.Bottom)
	assert.Equal(t, 20, margin.Left)
}

func TestMargin_Negative(t *testing.T) {
	// 负 margin（允许元素重叠）
	node := NewMockStyleNode("test")
	node.SetStyleMargin(-5, -5, -5, -5)

	margin := node.GetMargin()
	assert.Equal(t, -5, margin.Top)
}

func TestMargin_Large(t *testing.T) {
	// 大 margin 超过容器尺寸
	container := NewMockCompositeNode("container", 50, 50)

	child := NewMockStyleNode("child")
	child.SetStyleSize(20, 20)
	child.SetStyleMargin(100, 100, 100, 100) // margin 大于容器

	container.SetChildren([]Node{child})

	engine := NewEngine()
	result := engine.Layout(container, UnboundedConstraints())

	assert.NotNil(t, result)
	// 应该处理超大 margin
}

// =============================================================================
// Padding 影响测试
// =============================================================================

func TestPadding_Layout(t *testing.T) {
	// Padding 影响内容区域
	node := NewMockStyleNode("test")
	node.SetStylePadding(5, 10, 5, 10)

	padding := node.GetPadding()
	assert.Equal(t, 5, padding.Top)
	assert.Equal(t, 10, padding.Right)
	assert.Equal(t, 5, padding.Bottom)
	assert.Equal(t, 10, padding.Left)
}

func TestPadding_WithChildren(t *testing.T) {
	// 容器有 padding，子节点应该在 padding 内
	container := NewMockCompositeNode("container", 100, 100)
	container.padding = Padding{Top: 10, Right: 10, Bottom: 10, Left: 10}

	child := NewMockMeasurableNode("child", 80, 80)
	container.SetChildren([]Node{child})

	engine := NewEngine()
	result := engine.Layout(container, UnboundedConstraints())

	assert.NotNil(t, result)
	assert.Len(t, result.Root.Children, 1)
}

func TestPadding_Zero(t *testing.T) {
	node := NewMockStyleNode("test")
	node.SetStylePadding(0, 0, 0, 0)

	padding := node.GetPadding()
	assert.Equal(t, Padding{}, padding)
}

func TestPadding_Asymmetric(t *testing.T) {
	node := NewMockStyleNode("test")
	node.SetStylePadding(1, 2, 3, 4)

	padding := node.GetPadding()
	assert.Equal(t, 1, padding.Top)
	assert.Equal(t, 2, padding.Right)
	assert.Equal(t, 3, padding.Bottom)
	assert.Equal(t, 4, padding.Left)
}

// =============================================================================
// Border + Style 组合测试
// =============================================================================

func TestBorder_WithExplicitSize(t *testing.T) {
	// Border + 显式尺寸
	container := NewMockCompositeNode("container", 0, 0)
	container.SetBorder(BorderSingle)

	// 显式设置内部尺寸
	child := NewMockStyleNode("child")
	child.SetStyleSize(50, 30)
	container.SetChildren([]Node{child})

	engine := NewEngine()
	result := engine.Layout(container, NewConstraints(100, 100, 50, 50))

	assert.NotNil(t, result)
}

func TestBorder_WithPadding(t *testing.T) {
	// Border + Padding
	container := NewMockCompositeNode("container", 100, 100)
	container.SetBorder(BorderSingle)
	container.padding = Padding{Top: 5, Right: 5, Bottom: 5, Left: 5}

	child := NewMockMeasurableNode("child", 88, 88) // 100 - 2(border) - 10(padding) = 88
	container.SetChildren([]Node{child})

	engine := NewEngine()
	result := engine.Layout(container, UnboundedConstraints())

	assert.NotNil(t, result)
}

func TestBorder_WithMargin(t *testing.T) {
	// Border + Margin
	container := NewMockCompositeNode("container", 100, 100)
	container.SetBorder(BorderSingle)
	container.margin = Margin{Top: 5, Right: 5, Bottom: 5, Left: 5}

	child := NewMockMeasurableNode("child", 96, 96)
	container.SetChildren([]Node{child})

	engine := NewEngine()
	result := engine.Layout(container, UnboundedConstraints())

	assert.NotNil(t, result)
}

func TestBorder_FullBoxModel(t *testing.T) {
	// 完整盒模型：Border + Padding + Margin
	container := NewMockCompositeNode("container", 0, 0)
	container.SetBorder(BorderSingle)
	container.padding = Padding{Top: 5, Right: 5, Bottom: 5, Left: 5}
	container.margin = Margin{Top: 10, Right: 10, Bottom: 10, Left: 10}

	// 内容区域
	child := NewMockStyleNode("child")
	child.SetStyleSize(50, 30)
	container.SetChildren([]Node{child})

	engine := NewEngine()
	// 提供足够大的约束
	result := engine.Layout(container, NewConstraints(100, 100, 100, 100))

	assert.NotNil(t, result)
}

// =============================================================================
// Grid + Style 组合测试
// =============================================================================

func TestGrid_WithSizedChildren(t *testing.T) {
	// Grid 中的子节点有显式尺寸
	grid := NewMockGridNode("grid", 100, 50)
	grid.SetGridColumns([]GridDimension{GridFixed(50), GridFixed(50)})
	grid.SetGridRows([]GridDimension{GridFixed(50)})

	child1 := NewMockStyleNode("child1")
	child1.SetStyleSize(50, 50)

	child2 := NewMockStyleNode("child2")
	child2.SetStyleSize(50, 50)

	grid.SetGridCells([]GridCell{
		{Child: child1, Row: 0, Col: 0},
		{Child: child2, Row: 0, Col: 1},
	})

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	assert.NotNil(t, result)
	assert.Len(t, result.Root.Children, 2)
}

func TestGrid_WithPaddedCells(t *testing.T) {
	// Grid 单元格有 padding
	grid := NewMockGridNode("grid", 100, 50)
	grid.SetGridColumns([]GridDimension{GridFixed(50), GridFixed(50)})
	grid.SetGridRows([]GridDimension{GridFixed(50)})

	// 单元格内容有 padding
	cell1 := NewMockCompositeNode("cell1", 50, 50)
	cell1.padding = Padding{Top: 5, Right: 5, Bottom: 5, Left: 5}
	cell1.SetChildren([]Node{NewMockMeasurableNode("content", 40, 40)})

	grid.SetGridCells([]GridCell{{Child: cell1, Row: 0, Col: 0}})

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	assert.NotNil(t, result)
}

func TestGrid_FlexWithConstraints(t *testing.T) {
	// Flex 列 + 不同约束
	tests := []struct {
		name        string
		constraints Constraints
	}{
		{"small", NewConstraints(50, 50, 25, 25)},
		{"medium", NewConstraints(100, 100, 50, 50)},
		{"large", NewConstraints(200, 200, 100, 100)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			grid := NewMockGridNode("grid", 0, 0)
			grid.SetGridColumns([]GridDimension{
				GridFlex{Factor: 1},
				GridFlex{Factor: 2},
			})
			grid.SetGridRows([]GridDimension{GridFixed(25)})

			child1 := NewMockMeasurableNode("child1", 0, 25)
			child2 := NewMockMeasurableNode("child2", 0, 25)

			grid.SetGridCells([]GridCell{
				{Child: child1, Row: 0, Col: 0},
				{Child: child2, Row: 0, Col: 1},
			})

			engine := NewEngine()
			result := engine.Layout(grid, tt.constraints)

			assert.NotNil(t, result)
			assert.Len(t, result.Root.Children, 2)
		})
	}
}

// =============================================================================
// Absolute + Style 组合测试
// =============================================================================

func TestAbsolute_WithSizedChild(t *testing.T) {
	// Absolute 容器中的子节点有显式尺寸
	container := NewMockAbsoluteNode("container", 100, 100)
	container.SetPositionStyle(AbsolutePos(10), AbsolutePos(10), nil, nil, AnchorTopLeft)

	child := NewMockStyleNode("child")
	child.SetStyleSize(50, 30)
	container.SetChildren([]Node{child})

	engine := NewEngine()
	result := engine.Layout(container, UnboundedConstraints())

	assert.NotNil(t, result)
	assert.Len(t, result.Root.Children, 1)
}

func TestAbsolute_WithMargins(t *testing.T) {
	// Absolute 定位 + 子节点 margin
	container := NewMockAbsoluteNode("container", 100, 100)
	container.SetPositionStyle(AbsolutePos(0), AbsolutePos(0), nil, nil, AnchorTopLeft)

	child := NewMockStyleNode("child")
	child.SetStyleSize(50, 50)
	child.SetStyleMargin(10, 10, 10, 10)
	container.SetChildren([]Node{child})

	engine := NewEngine()
	result := engine.Layout(container, UnboundedConstraints())

	assert.NotNil(t, result)
}

func TestAbsolute_ExplicitSize(t *testing.T) {
	// Absolute 容器自身有显式尺寸
	container := NewMockAbsoluteNode("container", 80, 40)
	container.SetPositionStyle(AbsolutePos(10), AbsolutePos(10), nil, nil, AnchorTopLeft)

	child := NewMockMeasurableNode("child", 80, 40)
	container.SetChildren([]Node{child})

	engine := NewEngine()
	result := engine.Layout(container, UnboundedConstraints())

	assert.NotNil(t, result)
	assert.Equal(t, 80, result.Root.Width)
	assert.Equal(t, 40, result.Root.Height)
}

// =============================================================================
// Flex + Style 组合测试
// =============================================================================

func TestFlex_WithSizedChildren(t *testing.T) {
	// Flex 布局中的子节点有显式尺寸
	flex := NewMockFlexNode("flex", FlexRow)

	child1 := NewMockStyleNode("child1")
	child1.SetStyleSize(30, 50)

	child2 := NewMockStyleNode("child2")
	child2.SetStyleSize(40, 50)

	child3 := NewMockStyleNode("child3")
	child3.SetStyleSize(30, 50)

	flex.SetChildren([]Node{child1, child2, child3})

	engine := NewEngine()
	result := engine.Layout(flex, NewConstraints(100, 100, 50, 50))

	assert.NotNil(t, result)
	assert.Len(t, result.Root.Children, 3)
}

func TestFlex_WithMargins(t *testing.T) {
	// Flex 子节点有 margin
	flex := NewMockFlexNode("flex", FlexRow)

	child1 := NewMockStyleNode("child1")
	child1.SetStyleSize(30, 50)
	child1.SetStyleMargin(0, 5, 0, 0) // right margin 5

	child2 := NewMockStyleNode("child2")
	child2.SetStyleSize(30, 50)

	flex.SetChildren([]Node{child1, child2})

	engine := NewEngine()
	result := engine.Layout(flex, NewConstraints(100, 100, 50, 50))

	assert.NotNil(t, result)
}

func TestFlex_ColumnWithPadding(t *testing.T) {
	// Flex Column 容器有 padding
	flex := NewMockCompositeNode("flex", 100, 100)
	flex.flexStyle = &FlexStyle{Direction: FlexColumn}
	flex.padding = Padding{Top: 5, Right: 5, Bottom: 5, Left: 5}

	child1 := NewMockMeasurableNode("child1", 90, 30)
	child2 := NewMockMeasurableNode("child2", 90, 30)

	flex.SetChildren([]Node{child1, child2})

	engine := NewEngine()
	result := engine.Layout(flex, UnboundedConstraints())

	assert.NotNil(t, result)
}

// =============================================================================
// 复杂组合测试
// =============================================================================

func TestComplex_BorderGridAbsolute(t *testing.T) {
	// Border -> Grid -> Absolute
	border := NewMockCompositeNode("border", 102, 52)
	border.SetBorder(BorderSingle)

	grid := NewMockGridNode("grid", 100, 50)
	grid.SetGridColumns([]GridDimension{GridFixed(100)})
	grid.SetGridRows([]GridDimension{GridFixed(50)})

	abs := NewMockAbsoluteNode("abs", 100, 50)
	abs.SetPositionStyle(AbsolutePos(0), AbsolutePos(0), nil, nil, AnchorTopLeft)

	content := NewMockStyleNode("content")
	content.SetStyleSize(80, 30)
	abs.SetChildren([]Node{content})

	grid.SetGridCells([]GridCell{{Child: abs, Row: 0, Col: 0}})
	border.SetChildren([]Node{grid})

	engine := NewEngine()
	result := engine.Layout(border, UnboundedConstraints())

	assert.NotNil(t, result)
}

func TestComplex_MultipleBorders(t *testing.T) {
	// 多层嵌套 Border，每层有不同 padding
	outer := NewMockCompositeNode("outer", 100, 100)
	outer.SetBorder(BorderDouble)
	outer.padding = Padding{Top: 2, Right: 2, Bottom: 2, Left: 2}

	middle := NewMockCompositeNode("middle", 92, 92)
	middle.SetBorder(BorderSingle)
	middle.padding = Padding{Top: 3, Right: 3, Bottom: 3, Left: 3}

	inner := NewMockCompositeNode("inner", 82, 82)
	inner.SetBorder(BorderSingle)

	content := NewMockMeasurableNode("content", 80, 80)
	inner.SetChildren([]Node{content})
	middle.SetChildren([]Node{inner})
	outer.SetChildren([]Node{middle})

	engine := NewEngine()
	result := engine.Layout(outer, UnboundedConstraints())

	assert.NotNil(t, result)
}

func TestComplex_GridWithBorders(t *testing.T) {
	// Grid 每个单元格都有 Border
	grid := NewMockGridNode("grid", 100, 50)
	grid.SetGridColumns([]GridDimension{GridFixed(50), GridFixed(50)})
	grid.SetGridRows([]GridDimension{GridFixed(25), GridFixed(25)})

	for row := 0; row < 2; row++ {
		for col := 0; col < 2; col++ {
			cell := NewMockCompositeNode("cell", 48, 23)
			cell.SetBorder(BorderSingle)

			content := NewMockMeasurableNode("content", 46, 21)
			cell.SetChildren([]Node{content})

			grid.SetGridCells(append(grid.gridStyle.Cells, GridCell{
				Child: cell,
				Row:   row,
				Col:   col,
			}))
		}
	}

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	assert.NotNil(t, result)
	assert.Len(t, result.Root.Children, 4)
}

func TestComplex_FlexWithBorders(t *testing.T) {
	// Flex 布局中的每个子节点都有 Border
	flex := NewMockFlexNode("flex", FlexRow)

	for i := 0; i < 3; i++ {
		border := NewMockCompositeNode("border"+string(rune('0'+i)), 30, 50)
		border.SetBorder(BorderSingle)

		content := NewMockMeasurableNode("content", 28, 48)
		border.SetChildren([]Node{content})

		flex.SetChildren(append(flex.children, border))
	}

	engine := NewEngine()
	result := engine.Layout(flex, NewConstraints(100, 100, 50, 50))

	assert.NotNil(t, result)
	assert.Len(t, result.Root.Children, 3)
}

// =============================================================================
// Style 边界条件测试
// =============================================================================

func TestStyle_ZeroValues(t *testing.T) {
	node := NewMockStyleNode("test")
	node.SetStyleSize(0, 0)
	node.SetStyleMargin(0, 0, 0, 0)
	node.SetStylePadding(0, 0, 0, 0)

	size := node.Measure(UnboundedConstraints())
	assert.Equal(t, 0, size.Width)
	assert.Equal(t, 0, size.Height)

	margin := node.GetMargin()
	assert.Equal(t, Margin{}, margin)

	padding := node.GetPadding()
	assert.Equal(t, Padding{}, padding)
}

func TestStyle_MaxValues(t *testing.T) {
	// 最大整数值
	maxInt := int(^uint(0) >> 31)

	node := NewMockStyleNode("test")
	node.SetStyleSize(maxInt, maxInt)
	node.SetStyleMargin(maxInt, maxInt, maxInt, maxInt)

	// 在有限约束下
	size := node.Measure(NewConstraints(100, 100, 50, 50))
	assert.LessOrEqual(t, size.Width, 100)
	assert.LessOrEqual(t, size.Height, 50)
}

func TestStyle_NegativeSize(t *testing.T) {
	// 负数尺寸（应该被约束为 0）
	node := NewMockStyleNode("test")
	node.SetStyleSize(-100, -100)

	size := node.Measure(UnboundedConstraints())
	// 约束应该处理负数
	assert.LessOrEqual(t, size.Width, 0)
	assert.LessOrEqual(t, size.Height, 0)
}

func TestStyle_VeryLarge(t *testing.T) {
	// 非常大的尺寸
	node := NewMockStyleNode("test")
	node.SetStyleSize(10000, 10000)

	size := node.Measure(NewConstraints(100, 100, 50, 50))
	assert.LessOrEqual(t, size.Width, 100)
	assert.LessOrEqual(t, size.Height, 50)
}

// =============================================================================
// 约束边界测试
// =============================================================================

func TestConstraints_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		constraints Constraints
		inputW      int
		inputH      int
		expectW     int
		expectH     int
	}{
		{"zero max", Constraints{MinWidth: 0, MaxWidth: 0, MinHeight: 0, MaxHeight: 0}, 10, 10, 0, 0},
		{"min > max", Constraints{MinWidth: 100, MaxWidth: 50, MinHeight: 100, MaxHeight: 50}, 10, 10, 50, 50}, // Constrain applies min first, then max caps it
		{"negative min", Constraints{MinWidth: -10, MaxWidth: 100, MinHeight: -10, MaxHeight: 100}, 50, 50, 50, 50},
		{"equal min max", Constraints{MinWidth: 50, MaxWidth: 50, MinHeight: 25, MaxHeight: 25}, 100, 100, 50, 25},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, h := tt.constraints.Constrain(tt.inputW, tt.inputH)
			assert.Equal(t, tt.expectW, w)
			assert.Equal(t, tt.expectH, h)
		})
	}
}

func TestConstraints_Methods(t *testing.T) {
	c := NewConstraints(50, 100, 25, 50)

	assert.Equal(t, 50, c.ConstrainWidth(30))  // min
	assert.Equal(t, 100, c.ConstrainWidth(120)) // max
	assert.Equal(t, 75, c.ConstrainWidth(75))   // within

	assert.Equal(t, 25, c.ConstrainHeight(20))  // min
	assert.Equal(t, 50, c.ConstrainHeight(60))  // max
	assert.Equal(t, 40, c.ConstrainHeight(40))  // within
}

// =============================================================================
// 基准测试
// =============================================================================

func BenchmarkStyle_WithMargin(b *testing.B) {
	node := NewMockStyleNode("test")
	node.SetStyleSize(100, 50)
	node.SetStyleMargin(10, 10, 10, 10)

	constraints := UnboundedConstraints()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		node.Measure(constraints)
	}
}

func BenchmarkStyle_WithPadding(b *testing.B) {
	container := NewMockCompositeNode("container", 100, 100)
	container.padding = Padding{Top: 5, Right: 5, Bottom: 5, Left: 5}

	child := NewMockMeasurableNode("child", 90, 90)
	container.SetChildren([]Node{child})

	engine := NewEngine()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.Layout(container, UnboundedConstraints())
	}
}

func BenchmarkComplex_FullBoxModel(b *testing.B) {
	border := NewMockCompositeNode("border", 100, 100)
	border.SetBorder(BorderSingle)
	border.padding = Padding{Top: 5, Right: 5, Bottom: 5, Left: 5}
	border.margin = Margin{Top: 10, Right: 10, Bottom: 10, Left: 10}

	child := NewMockStyleNode("child")
	child.SetStyleSize(80, 80)
	border.SetChildren([]Node{child})

	engine := NewEngine()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.Layout(border, UnboundedConstraints())
	}
}
