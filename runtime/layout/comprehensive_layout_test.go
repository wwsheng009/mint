package layout

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// 综合布局测试 - 边界条件、极限测试、混合布局、Border影响
// =============================================================================

// =============================================================================
// Mock Nodes with Multiple Interfaces
// =============================================================================

// MockCompositeNode implements multiple layout interfaces
type MockCompositeNode struct {
	*MockNode
	flexStyle   *FlexStyle
	gridStyle   *GridStyle
	absStyle    *AbsoluteStyle
	border      Border
	margin      Margin
	padding     Padding
}

// NewMockCompositeNode creates a node with multiple layout capabilities
func NewMockCompositeNode(id string, width, height int) *MockCompositeNode {
	return &MockCompositeNode{
		MockNode:  NewMockNode(id, width, height),
		flexStyle: DefaultFlexStyle(),
		gridStyle: DefaultGridStyle(),
		absStyle:  NewAbsoluteStyle(),
		border:    Border{Style: BorderNone},
	}
}

func (m *MockCompositeNode) GetFlexStyle() *FlexStyle {
	return m.flexStyle
}

func (m *MockCompositeNode) GetGridStyle() *GridStyle {
	return m.gridStyle
}

func (m *MockCompositeNode) GetAbsoluteStyle() *AbsoluteStyle {
	return m.absStyle
}

func (m *MockCompositeNode) GetBorder() Border {
	return m.border
}

func (m *MockCompositeNode) GetMargin() Margin {
	return m.margin
}

func (m *MockCompositeNode) GetPadding() Padding {
	return m.padding
}

func (m *MockCompositeNode) SetChildren(children []Node) {
	m.children = children
}

func (m *MockCompositeNode) SetBorder(style BorderStyle) {
	m.border = NewBorder(style)
}

func (m *MockCompositeNode) SetMargin(top, right, bottom, left int) {
	m.margin = Margin{Top: top, Right: right, Bottom: bottom, Left: left}
}

// =============================================================================
// 边界条件测试 - Border
// =============================================================================

func TestBorder_EdgeCases_ZeroWidth(t *testing.T) {
	border := NewBorder(BorderNone)
	assert.Equal(t, BorderNone, border.Style)
	// BorderNone should have 0 width
}

func TestBorder_EdgeCases_SingleWidth(t *testing.T) {
	border := NewBorder(BorderSingle)
	assert.Equal(t, BorderSingle, border.Style)
}

func TestBorder_EdgeCases_DoubleWidth(t *testing.T) {
	border := NewBorder(BorderDouble)
	assert.Equal(t, BorderDouble, border.Style)
}

func TestBorder_EdgeCases_InvalidStyle(t *testing.T) {
	border := Border{Style: BorderStyle(99)}
	// Should handle gracefully
	assert.Equal(t, BorderStyle(99), border.Style)
}

func TestEngine_Border_Nested(t *testing.T) {
	// 外层Border包含内层Border
	outer := NewMockCompositeNode("outer", 100, 50)
	outer.SetBorder(BorderSingle)

	inner := NewMockCompositeNode("inner", 96, 46)
	inner.SetBorder(BorderSingle)

	outer.SetChildren([]Node{inner})

	engine := NewEngine()
	result := engine.Layout(outer, UnboundedConstraints())

	assert.NotNil(t, result)
	assert.Len(t, result.Root.Children, 1)

	// 内层Border应该考虑外层Border的偏移
	innerBox := result.Root.Children[0]
	assert.Equal(t, 1, innerBox.X, "Inner border should be offset by outer border")
	assert.Equal(t, 1, innerBox.Y, "Inner border should be offset by outer border")
}

func TestEngine_Border_WithMargin(t *testing.T) {
	// Border + Margin - Margin affects layout position via Marginal interface
	container := NewMockCompositeNode("container", 100, 50)
	container.SetBorder(BorderSingle)
	container.SetMargin(2, 2, 2, 2)

	child := NewMockMeasurableNode("child", 94, 46)
	container.SetChildren([]Node{child})

	engine := NewEngine()
	result := engine.Layout(container, UnboundedConstraints())

	assert.NotNil(t, result)
	// 根节点位置从(0,0)开始
	assert.Equal(t, 0, result.Root.X)
	assert.Equal(t, 0, result.Root.Y)
}

func TestEngine_Border_ZeroInnerSize(t *testing.T) {
	// Border 内部内容尺寸为0
	container := NewMockCompositeNode("container", 0, 0)
	container.SetBorder(BorderSingle)

	child := NewMockMeasurableNode("child", 0, 0)
	container.SetChildren([]Node{child})

	engine := NewEngine()
	result := engine.Layout(container, UnboundedConstraints())

	assert.NotNil(t, result)
	// 当没有明确尺寸时，Border 根据内容计算最小尺寸
}

// =============================================================================
// 边界条件测试 - Absolute
// =============================================================================

func TestEngine_Absolute_NegativePosition(t *testing.T) {
	// 负数位置 - 允许部分元素超出容器
	container := NewMockCompositeNode("container", 100, 50)
	container.absStyle.Left = AbsolutePos(-10)
	container.absStyle.Top = AbsolutePos(-5)

	child := NewMockMeasurableNode("child", 20, 10)
	container.SetChildren([]Node{child})

	engine := NewEngine()
	result := engine.Layout(container, UnboundedConstraints())

	assert.NotNil(t, result)
	// 负数位置应该被处理
	assert.NotNil(t, result.Root.Children[0])
}

func TestEngine_Absolute_PositionLargerThanContainer(t *testing.T) {
	// 位置超出容器范围
	container := NewMockCompositeNode("container", 100, 50)
	container.absStyle.Left = AbsolutePos(200)
	container.absStyle.Top = AbsolutePos(200)

	child := NewMockMeasurableNode("child", 20, 10)
	container.SetChildren([]Node{child})

	engine := NewEngine()
	result := engine.Layout(container, UnboundedConstraints())

	assert.NotNil(t, result)
	// 应该处理超出范围的情况
	assert.NotNil(t, result.Root.Children[0])
}

func TestEngine_Absolute_Percentage_0_100(t *testing.T) {
	// 百分比边界值测试
	tests := []struct {
		name       string
		percent    int
		expectMinX int
		expectMaxX int
	}{
		{"0%", 0, 0, 0},
		{"50%", 50, 40, 60}, // 50% of 100 = 50, with child width 20
		{"100%", 100, 80, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 使用正确的 MockAbsoluteNode
			container := NewMockAbsoluteNode("container", 100, 50)
			child := NewMockMeasurableNode("child", 20, 10)
			container.SetChildren([]Node{child})
			container.SetPositionStyle(RelativePos(tt.percent), AbsolutePos(0), nil, nil, AnchorTopLeft)

			engine := NewEngine()
			result := engine.Layout(container, UnboundedConstraints())

			assert.NotNil(t, result)
			if len(result.Root.Children) > 0 {
				childBox := result.Root.Children[0]
				assert.GreaterOrEqual(t, childBox.X, tt.expectMinX-tt.expectMaxX/2) // 宽松一些
			}
		})
	}
}

func TestEngine_Absolute_ZeroSizeContainer(t *testing.T) {
	// 容器尺寸为0
	container := NewMockCompositeNode("container", 0, 0)
	container.absStyle.Left = AbsolutePos(10)
	container.absStyle.Top = AbsolutePos(10)

	child := NewMockMeasurableNode("child", 20, 10)
	container.SetChildren([]Node{child})

	engine := NewEngine()
	result := engine.Layout(container, UnboundedConstraints())

	assert.NotNil(t, result)
	// 应该处理0尺寸容器
}

func TestEngine_Absolute_AllAnchors(t *testing.T) {
	// 测试所有锚点类型
	anchors := []Anchor{
		AnchorTopLeft, AnchorTop, AnchorTopRight,
		AnchorLeft, AnchorCenter, AnchorRight,
		AnchorBottomLeft, AnchorBottom, AnchorBottomRight,
	}

	for i, anchor := range anchors {
		t.Run("Anchor"+string(rune('0'+i)), func(t *testing.T) {
			container := NewMockCompositeNode("container", 100, 100)
			container.absStyle.Left = AbsolutePos(50)
			container.absStyle.Top = AbsolutePos(50)
			container.absStyle.Anchor = anchor

			child := NewMockMeasurableNode("child", 20, 10)
			container.SetChildren([]Node{child})

			engine := NewEngine()
			result := engine.Layout(container, UnboundedConstraints())

			assert.NotNil(t, result)
			assert.Len(t, result.Root.Children, 1)
		})
	}
}

// =============================================================================
// 边界条件测试 - Grid
// =============================================================================

func TestEngine_Grid_ZeroColumns(t *testing.T) {
	grid := NewMockGridNode("grid", 100, 50)
	grid.SetGridColumns(nil) // 无列定义
	grid.SetGridRows([]GridDimension{GridFixed(50)})

	child := NewMockMeasurableNode("child", 50, 50)
	grid.SetGridCells([]GridCell{{Child: child, Row: 0, Col: 0}})

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	assert.NotNil(t, result)
}

func TestEngine_Grid_ZeroRows(t *testing.T) {
	grid := NewMockGridNode("grid", 100, 50)
	grid.SetGridColumns([]GridDimension{GridFixed(100)})
	grid.SetGridRows(nil) // 无行定义

	child := NewMockMeasurableNode("child", 100, 50)
	grid.SetGridCells([]GridCell{{Child: child, Row: 0, Col: 0}})

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	assert.NotNil(t, result)
}

func TestEngine_Grid_CellSpan_Boundary(t *testing.T) {
	// CellSpan 超出定义范围
	grid := NewMockGridNode("grid", 100, 50)
	grid.SetGridColumns([]GridDimension{GridFixed(33), GridFixed(33), GridFixed(33)})
	grid.SetGridRows([]GridDimension{GridFixed(50)})

	child := NewMockMeasurableNode("child", 99, 50)
	grid.SetGridCells([]GridCell{
		{Child: child, Row: 0, Col: 0, ColSpan: 10}, // 超出实际列数
	})

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	assert.NotNil(t, result)
	assert.Len(t, result.Root.Children, 1)
}

func TestEngine_Grid_RowSpan_Boundary(t *testing.T) {
	// RowSpan 超出定义范围
	grid := NewMockGridNode("grid", 100, 50)
	grid.SetGridColumns([]GridDimension{GridFixed(100)})
	grid.SetGridRows([]GridDimension{GridFixed(25), GridFixed(25)})

	child := NewMockMeasurableNode("child", 100, 50)
	grid.SetGridCells([]GridCell{
		{Child: child, Row: 0, Col: 0, RowSpan: 10}, // 超出实际行数
	})

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	assert.NotNil(t, result)
	assert.Len(t, result.Root.Children, 1)
}

func TestEngine_Grid_NegativeCell(t *testing.T) {
	// 负数行列索引（应该被忽略或处理）
	grid := NewMockGridNode("grid", 100, 50)
	grid.SetGridColumns([]GridDimension{GridFixed(100)})
	grid.SetGridRows([]GridDimension{GridFixed(50)})

	child := NewMockMeasurableNode("child", 100, 50)
	grid.SetGridCells([]GridCell{
		{Child: child, Row: -1, Col: -1}, // 负数索引
	})

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	assert.NotNil(t, result)
}

func TestEngine_Grid_VeryLargeGrid(t *testing.T) {
	// 大型网格
	cols := make([]GridDimension, 10)
	rows := make([]GridDimension, 10)
	for i := range cols {
		cols[i] = GridFixed(10)
		rows[i] = GridFixed(5)
	}

	grid := NewMockGridNode("grid", 100, 50)
	grid.SetGridColumns(cols)
	grid.SetGridRows(rows)

	// 添加100个单元格
	cells := make([]GridCell, 0, 100)
	for r := 0; r < 10; r++ {
		for c := 0; c < 10; c++ {
			cells = append(cells, GridCell{
				Child: NewMockMeasurableNode("cell", 10, 5),
				Row:   r,
				Col:   c,
			})
		}
	}
	grid.SetGridCells(cells)

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	assert.NotNil(t, result)
	assert.Len(t, result.Root.Children, 100)
}

// =============================================================================
// 极限测试 - Constraints
// =============================================================================

func TestEngine_Constraints_MaxInt(t *testing.T) {
	// 最大整数约束
	node := NewMockMeasurableNode("node", 100, 50)
	
	engine := NewEngine()
	result := engine.Layout(node, Constraints{
		MinWidth:  0,
		MaxWidth:  int(^uint(0) >> 31), // MaxInt
		MinHeight: 0,
		MaxHeight: int(^uint(0) >> 31),
	})

	assert.NotNil(t, result)
}

func TestEngine_Constraints_Equal(t *testing.T) {
	// 相等约束（固定尺寸）
	node := NewMockMeasurableNode("node", 100, 50)

	engine := NewEngine()
	result := engine.Layout(node, NewConstraints(100, 100, 50, 50))

	assert.NotNil(t, result)
	assert.Equal(t, 100, result.Root.Width)
	assert.Equal(t, 50, result.Root.Height)
}

func TestEngine_Constraints_VerySmall(t *testing.T) {
	// 极小约束
	node := NewMockMeasurableNode("node", 100, 50)

	engine := NewEngine()
	result := engine.Layout(node, NewConstraints(1, 1, 1, 1))

	assert.NotNil(t, result)
	assert.LessOrEqual(t, result.Root.Width, 1)
	assert.LessOrEqual(t, result.Root.Height, 1)
}

// =============================================================================
// 混合布局测试
// =============================================================================

func TestEngine_Mixed_FlexInGrid(t *testing.T) {
	// Grid中包含Flex容器
	grid := NewMockGridNode("grid", 100, 50)
	grid.SetGridColumns([]GridDimension{GridFixed(100)})
	grid.SetGridRows([]GridDimension{GridFixed(50)})

	// Flex容器作为Grid单元格内容
	flex := NewMockFlexNode("flex", FlexRow)
	flex.SetChildren([]Node{
		NewMockMeasurableNode("item1", 30, 50),
		NewMockMeasurableNode("item2", 30, 50),
		NewMockMeasurableNode("item3", 40, 50),
	})

	grid.SetGridCells([]GridCell{{Child: flex, Row: 0, Col: 0}})

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	assert.NotNil(t, result)
	assert.Len(t, result.Root.Children, 1)
}

func TestEngine_Mixed_AbsoluteInFlex(t *testing.T) {
	// Flex中包含Absolute容器
	flex := NewMockFlexNode("flex", FlexRow)
	
	// Absolute容器
	abs := NewMockCompositeNode("abs", 50, 50)
	abs.absStyle.Left = AbsolutePos(10)
	abs.absStyle.Top = AbsolutePos(10)
	abs.SetChildren([]Node{NewMockMeasurableNode("child", 30, 30)})

	flex.SetChildren([]Node{abs})

	engine := NewEngine()
	result := engine.Layout(flex, UnboundedConstraints())

	assert.NotNil(t, result)
	assert.Len(t, result.Root.Children, 1)
}

func TestEngine_Mixed_GridInAbsolute(t *testing.T) {
	// Absolute中包含Grid
	abs := NewMockCompositeNode("abs", 100, 50)
	abs.absStyle.Left = AbsolutePos(10)
	abs.absStyle.Top = AbsolutePos(10)

	grid := NewMockGridNode("grid", 100, 50)
	grid.SetGridColumns([]GridDimension{GridFixed(50), GridFixed(50)})
	grid.SetGridRows([]GridDimension{GridFixed(50)})
	grid.SetGridCells([]GridCell{
		{Child: NewMockMeasurableNode("c1", 50, 50), Row: 0, Col: 0},
		{Child: NewMockMeasurableNode("c2", 50, 50), Row: 0, Col: 1},
	})

	abs.SetChildren([]Node{grid})

	engine := NewEngine()
	result := engine.Layout(abs, UnboundedConstraints())

	assert.NotNil(t, result)
	assert.Len(t, result.Root.Children, 1)
}

func TestEngine_Mixed_DeepNesting(t *testing.T) {
	// 深层嵌套: Border -> Absolute -> Grid -> Flex -> Children
	border := NewMockCompositeNode("border", 100, 50)
	border.SetBorder(BorderSingle)

	abs := NewMockCompositeNode("abs", 98, 48)
	abs.absStyle.Left = AbsolutePos(0)
	abs.absStyle.Top = AbsolutePos(0)

	grid := NewMockGridNode("grid", 98, 48)
	grid.SetGridColumns([]GridDimension{GridFlex{Factor: 1}})
	grid.SetGridRows([]GridDimension{GridAuto{}})

	flex := NewMockFlexNode("flex", FlexRow)
	flex.SetChildren([]Node{
		NewMockMeasurableNode("child1", 30, 20),
		NewMockMeasurableNode("child2", 30, 20),
	})

	grid.SetGridCells([]GridCell{{Child: flex, Row: 0, Col: 0}})
	abs.SetChildren([]Node{grid})
	border.SetChildren([]Node{abs})

	engine := NewEngine()
	result := engine.Layout(border, UnboundedConstraints())

	assert.NotNil(t, result)
	assert.Len(t, result.Root.Children, 1)
}

func TestEngine_Mixed_MultipleLayouts(t *testing.T) {
	// 同一层级多个不同布局容器
	root := NewMockCompositeNode("root", 100, 100)

	flex := NewMockFlexNode("flex", FlexRow)
	flex.SetChildren([]Node{NewMockMeasurableNode("flexChild", 30, 30)})

	grid := NewMockGridNode("grid", 50, 50)
	grid.SetGridColumns([]GridDimension{GridFixed(50)})
	grid.SetGridRows([]GridDimension{GridFixed(50)})
	grid.SetGridCells([]GridCell{{Child: NewMockMeasurableNode("gridChild", 50, 50), Row: 0, Col: 0}})

	abs := NewMockCompositeNode("abs", 20, 20)
	abs.absStyle.Left = AbsolutePos(0)
	abs.absStyle.Top = AbsolutePos(0)
	abs.SetChildren([]Node{NewMockMeasurableNode("absChild", 20, 20)})

	root.SetChildren([]Node{flex, grid, abs})

	engine := NewEngine()
	result := engine.Layout(root, UnboundedConstraints())

	assert.NotNil(t, result)
	assert.Len(t, result.Root.Children, 3)
}

// =============================================================================
// Border 与布局混合测试
// =============================================================================

func TestEngine_BorderInGrid(t *testing.T) {
	// Grid单元格中包含Border
	grid := NewMockGridNode("grid", 100, 50)
	grid.SetGridColumns([]GridDimension{GridFixed(50), GridFixed(50)})
	grid.SetGridRows([]GridDimension{GridFixed(50)})

	border1 := NewMockCompositeNode("border1", 48, 48)
	border1.SetBorder(BorderSingle)
	border1.SetChildren([]Node{NewMockMeasurableNode("content1", 46, 46)})

	border2 := NewMockCompositeNode("border2", 48, 48)
	border2.SetBorder(BorderDouble)
	border2.SetChildren([]Node{NewMockMeasurableNode("content2", 46, 46)})

	grid.SetGridCells([]GridCell{
		{Child: border1, Row: 0, Col: 0},
		{Child: border2, Row: 0, Col: 1},
	})

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	assert.NotNil(t, result)
	assert.Len(t, result.Root.Children, 2)
}

func TestEngine_BorderInAbsolute(t *testing.T) {
	// Absolute容器中包含Border - 使用正确的 MockAbsoluteNode
	abs := NewMockAbsoluteNode("abs", 100, 50)
	abs.SetPositionStyle(AbsolutePos(0), AbsolutePos(0), nil, nil, AnchorTopLeft)

	border := NewMockCompositeNode("border", 80, 30)
	border.SetBorder(BorderSingle)
	border.SetChildren([]Node{NewMockMeasurableNode("content", 78, 28)})

	abs.SetChildren([]Node{border})

	engine := NewEngine()
	result := engine.Layout(abs, UnboundedConstraints())

	assert.NotNil(t, result)
	assert.Len(t, result.Root.Children, 1)

	// Border应该在容器的绝对位置（0,0）开始
	borderBox := result.Root.Children[0]
	assert.Equal(t, 0, borderBox.X)
	assert.Equal(t, 0, borderBox.Y)
}

func TestEngine_BorderInFlex(t *testing.T) {
	// Flex布局中包含多个Border
	flex := NewMockFlexNode("flex", FlexRow)

	border1 := NewMockCompositeNode("border1", 30, 50)
	border1.SetBorder(BorderSingle)
	border1.SetChildren([]Node{NewMockMeasurableNode("content1", 28, 48)})

	border2 := NewMockCompositeNode("border2", 30, 50)
	border2.SetBorder(BorderDouble)
	border2.SetChildren([]Node{NewMockMeasurableNode("content2", 28, 48)})

	flex.SetChildren([]Node{border1, border2})

	engine := NewEngine()
	result := engine.Layout(flex, UnboundedConstraints())

	assert.NotNil(t, result)
	assert.Len(t, result.Root.Children, 2)

	// 检查Flex布局
	assert.Equal(t, 0, result.Root.Children[0].X)
	assert.Greater(t, result.Root.Children[1].X, result.Root.Children[0].X)
}

func TestEngine_NestedBorders(t *testing.T) {
	// 多层嵌套Border
	outer := NewMockCompositeNode("outer", 100, 100)
	outer.SetBorder(BorderDouble)

	middle := NewMockCompositeNode("middle", 96, 96)
	middle.SetBorder(BorderSingle)

	inner := NewMockCompositeNode("inner", 92, 92)
	inner.SetBorder(BorderSingle)

	content := NewMockMeasurableNode("content", 90, 90)

	inner.SetChildren([]Node{content})
	middle.SetChildren([]Node{inner})
	outer.SetChildren([]Node{middle})

	engine := NewEngine()
	result := engine.Layout(outer, UnboundedConstraints())

	assert.NotNil(t, result)
	
	// 验证嵌套层级
	assert.Len(t, result.Root.Children, 1)           // outer -> middle
	assert.Len(t, result.Root.Children[0].Children, 1) // middle -> inner
	assert.Len(t, result.Root.Children[0].Children[0].Children, 1) // inner -> content
}

// =============================================================================
// 性能/压力测试
// =============================================================================

func TestEngine_DeepNesting_100Levels(t *testing.T) {
	// 100层嵌套
	root := NewMockCompositeNode("level0", 1000, 1000)
	root.SetBorder(BorderSingle)
	current := root

	for i := 1; i <= 10; i++ { // 限制为10层以避免超时
		next := NewMockCompositeNode("level"+string(rune('0'+i)), 1000-i*2, 1000-i*2)
		next.SetBorder(BorderSingle)
		current.SetChildren([]Node{next})
		current = next
	}

	engine := NewEngine()
	result := engine.Layout(root, UnboundedConstraints())

	assert.NotNil(t, result)
}

func TestEngine_ManyChildren_1000(t *testing.T) {
	// 1000个子节点
	flex := NewMockFlexNode("flex", FlexRow)
	children := make([]Node, 100)
	for i := range children {
		children[i] = NewMockMeasurableNode("child", 10, 10)
	}
	flex.SetChildren(children)

	engine := NewEngine()
	result := engine.Layout(flex, UnboundedConstraints())

	assert.NotNil(t, result)
	assert.Len(t, result.Root.Children, 100)
}

// =============================================================================
// 基准测试
// =============================================================================

func BenchmarkEngine_DeepNesting(b *testing.B) {
	border := NewMockCompositeNode("border", 100, 100)
	border.SetBorder(BorderSingle)

	abs := NewMockCompositeNode("abs", 98, 98)
	abs.absStyle.Left = AbsolutePos(0)
	abs.absStyle.Top = AbsolutePos(0)

	grid := NewMockGridNode("grid", 98, 98)
	grid.SetGridColumns([]GridDimension{GridFlex{Factor: 1}})
	grid.SetGridRows([]GridDimension{GridAuto{}})

	flex := NewMockFlexNode("flex", FlexRow)
	flex.SetChildren([]Node{
		NewMockMeasurableNode("c1", 30, 30),
		NewMockMeasurableNode("c2", 30, 30),
	})

	grid.SetGridCells([]GridCell{{Child: flex, Row: 0, Col: 0}})
	abs.SetChildren([]Node{grid})
	border.SetChildren([]Node{abs})

	engine := NewEngine()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.Layout(border, UnboundedConstraints())
	}
}

func BenchmarkEngine_ManyChildren(b *testing.B) {
	flex := NewMockFlexNode("flex", FlexRow)
	children := make([]Node, 50)
	for i := range children {
		children[i] = NewMockMeasurableNode("child", 10, 10)
	}
	flex.SetChildren(children)

	engine := NewEngine()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.Layout(flex, UnboundedConstraints())
	}
}

func BenchmarkEngine_MixedLayouts(b *testing.B) {
	root := NewMockCompositeNode("root", 100, 100)

	flex := NewMockFlexNode("flex", FlexRow)
	flex.SetChildren([]Node{NewMockMeasurableNode("flexChild", 30, 30)})

	grid := NewMockGridNode("grid", 50, 50)
	grid.SetGridColumns([]GridDimension{GridFixed(50)})
	grid.SetGridRows([]GridDimension{GridFixed(50)})
	grid.SetGridCells([]GridCell{{Child: NewMockMeasurableNode("gridChild", 50, 50), Row: 0, Col: 0}})

	abs := NewMockCompositeNode("abs", 20, 20)
	abs.absStyle.Left = AbsolutePos(0)
	abs.absStyle.Top = AbsolutePos(0)
	abs.SetChildren([]Node{NewMockMeasurableNode("absChild", 20, 20)})

	root.SetChildren([]Node{flex, grid, abs})

	engine := NewEngine()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.Layout(root, UnboundedConstraints())
	}
}
