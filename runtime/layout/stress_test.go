package layout

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// 极限测试 - Stress Tests
// =============================================================================

// =============================================================================
// 深度嵌套极限
// =============================================================================

func TestStress_DeepNesting_50Levels(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode")
	}

	// 50层嵌套 Border
	root := NewMockCompositeNode("level0", 200, 200)
	root.SetBorder(BorderSingle)
	current := root

	for i := 1; i < 50; i++ {
		next := NewMockCompositeNode("level"+string(rune('0'+i%10)), 198-i*2, 198-i*2)
		next.SetBorder(BorderSingle)
		current.SetChildren([]Node{next})
		current = next
	}

	// 最内层
	content := NewMockMeasurableNode("content", 50, 50)
	current.SetChildren([]Node{content})

	engine := NewEngine()
	result := engine.Layout(root, UnboundedConstraints())

	assert.NotNil(t, result)
}

func TestStress_DeepNesting_MixedLayouts(t *testing.T) {
	// 深度嵌套，混合使用不同布局
	root := NewMockCompositeNode("root", 1000, 1000)
	root.SetBorder(BorderSingle)

	level1 := NewMockAbsoluteNode("level1", 998, 998)
	level1.SetPositionStyle(AbsolutePos(1), AbsolutePos(1), nil, nil, AnchorTopLeft)

	level2 := NewMockGridNode("level2", 996, 996)
	level2.SetGridColumns([]GridDimension{GridFlex{Factor: 1}})
	level2.SetGridRows([]GridDimension{GridFlex{Factor: 1}})

	level3 := NewMockFlexNode("level3", FlexColumn)

	level4 := NewMockCompositeNode("level4", 990, 330)
	level4.SetBorder(BorderDouble)

	// 继续嵌套...
	inner := NewMockMeasurableNode("inner", 986, 326)

	level4.SetChildren([]Node{inner})
	level3.SetChildren([]Node{level4})
	level2.SetGridCells([]GridCell{{Child: level3, Row: 0, Col: 0}})
	level1.SetChildren([]Node{level2})
	root.SetChildren([]Node{level1})

	engine := NewEngine()
	result := engine.Layout(root, UnboundedConstraints())

	assert.NotNil(t, result)
}

// =============================================================================
// 大量子节点极限
// =============================================================================

func TestStress_ManyChildren_500(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode")
	}

	// 500个子节点
	flex := NewMockFlexNode("flex", FlexRow)
	children := make([]Node, 500)
	for i := range children {
		children[i] = NewMockMeasurableNode("child", 2, 10)
	}
	flex.SetChildren(children)

	engine := NewEngine()
	result := engine.Layout(flex, UnboundedConstraints())

	assert.NotNil(t, result)
	assert.Len(t, result.Root.Children, 500)
}

func TestStress_LargeGrid_20x20(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode")
	}

	// 20x20 网格 = 400 个单元格
	cols := make([]GridDimension, 20)
	rows := make([]GridDimension, 20)
	for i := range cols {
		cols[i] = GridFixed(5)
		rows[i] = GridFixed(2)
	}

	grid := NewMockGridNode("grid", 100, 40)
	grid.SetGridColumns(cols)
	grid.SetGridRows(rows)

	cells := make([]GridCell, 0, 400)
	for r := 0; r < 20; r++ {
		for c := 0; c < 20; c++ {
			cells = append(cells, GridCell{
				Child: NewMockMeasurableNode("cell", 5, 2),
				Row:   r,
				Col:   c,
			})
		}
	}
	grid.SetGridCells(cells)

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	assert.NotNil(t, result)
	assert.Len(t, result.Root.Children, 400)
}

func TestStress_ManyAbsoluteChildren(t *testing.T) {
	// 100个绝对定位的子节点
	container := NewMockAbsoluteNode("container", 1000, 1000)

	children := make([]Node, 100)
	for i := range children {
		children[i] = NewMockMeasurableNode("child", 10, 10)
	}
	container.SetChildren(children)

	engine := NewEngine()
	result := engine.Layout(container, UnboundedConstraints())

	assert.NotNil(t, result)
	assert.Len(t, result.Root.Children, 100)
}

// =============================================================================
// 大尺寸极限
// =============================================================================

func TestStress_LargeSize(t *testing.T) {
	// 非常大的容器尺寸
	container := NewMockCompositeNode("container", 10000, 5000)
	container.SetBorder(BorderSingle)

	child := NewMockMeasurableNode("child", 9998, 4998)
	container.SetChildren([]Node{child})

	engine := NewEngine()
	result := engine.Layout(container, UnboundedConstraints())

	assert.NotNil(t, result)
	assert.Equal(t, 10000, result.Root.Width)
	assert.Equal(t, 5000, result.Root.Height)
}

func TestStress_MaxConstraints(t *testing.T) {
	// 最大约束
	node := NewMockMeasurableNode("node", 100, 50)

	maxInt := int(^uint(0) >> 31)
	constraints := Constraints{
		MinWidth:  0,
		MaxWidth:  maxInt,
		MinHeight: 0,
		MaxHeight: maxInt,
	}

	engine := NewEngine()
	result := engine.Layout(node, constraints)

	assert.NotNil(t, result)
}

// =============================================================================
// 递归深度极限
// =============================================================================

func TestStress_RecursiveLayoutCalls(t *testing.T) {
	// 测试布局引擎在多次调用下的稳定性
	node := NewMockMeasurableNode("node", 100, 50)
	engine := NewEngine()
	constraints := UnboundedConstraints()

	for i := 0; i < 1000; i++ {
		result := engine.Layout(node, constraints)
		assert.NotNil(t, result)
	}
}

// =============================================================================
// 边界值组合测试
// =============================================================================

func TestStress_BoundaryCombinations(t *testing.T) {
	// 各种边界值的组合
	testCases := []struct {
		name        string
		width       int
		height      int
		constraints Constraints
	}{
		{"0x0 unbounded", 0, 0, UnboundedConstraints()},
		{"1x1 unbounded", 1, 1, UnboundedConstraints()},
		{"max unbounded", 10000, 10000, UnboundedConstraints()},
		{"0x0 tight", 0, 0, NewConstraints(0, 0, 0, 0)},
		{"1x1 tight", 1, 1, NewConstraints(1, 1, 1, 1)},
		{"exceed constraints", 100, 100, NewConstraints(10, 10, 10, 10)},
		{"below constraints", 5, 5, NewConstraints(10, 50, 10, 50)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			node := NewMockMeasurableNode("node", tc.width, tc.height)
			engine := NewEngine()
			result := engine.Layout(node, tc.constraints)
			assert.NotNil(t, result)
		})
	}
}

// =============================================================================
// 并发安全测试
// =============================================================================

func TestStress_ConcurrentLayout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode")
	}

	// 并发布局测试
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			node := NewMockMeasurableNode("node", 100, 50)
			engine := NewEngine()

			for j := 0; j < 100; j++ {
				result := engine.Layout(node, UnboundedConstraints())
				if result == nil {
					t.Error("Layout returned nil")
					return
				}
			}
			done <- true
		}(i)
	}

	// 等待所有 goroutine 完成
	for i := 0; i < 10; i++ {
		<-done
	}
}

// =============================================================================
// 内存压力测试
// =============================================================================

func TestStress_MemoryPressure(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode")
	}

	// 创建大量布局树
	for i := 0; i < 100; i++ {
		root := NewMockCompositeNode("root", 100, 100)
		root.SetBorder(BorderSingle)

		grid := NewMockGridNode("grid", 98, 98)
		grid.SetGridColumns([]GridDimension{GridFixed(49), GridFixed(49)})
		grid.SetGridRows([]GridDimension{GridFixed(49), GridFixed(49)})

		for r := 0; r < 2; r++ {
			for c := 0; c < 2; c++ {
				abs := NewMockAbsoluteNode("abs", 49, 49)
				abs.SetChildren([]Node{NewMockMeasurableNode("content", 47, 47)})
				grid.SetGridCells(append(grid.gridStyle.Cells, GridCell{Child: abs, Row: r, Col: c}))
			}
		}

		root.SetChildren([]Node{grid})

		engine := NewEngine()
		result := engine.Layout(root, UnboundedConstraints())
		assert.NotNil(t, result)
	}
}

// =============================================================================
// 特殊场景测试
// =============================================================================

func TestStress_EmptyToFullTransition(t *testing.T) {
	// 从空到满的过渡
	container := NewMockCompositeNode("container", 100, 100)

	engine := NewEngine()

	// 空容器
	result := engine.Layout(container, UnboundedConstraints())
	assert.NotNil(t, result)

	// 添加一个子节点
	container.SetChildren([]Node{NewMockMeasurableNode("child1", 100, 100)})
	result = engine.Layout(container, UnboundedConstraints())
	assert.NotNil(t, result)

	// 添加多个子节点
	children := make([]Node, 10)
	for i := range children {
		children[i] = NewMockMeasurableNode("child", 10, 10)
	}
	container.SetChildren(children)
	result = engine.Layout(container, UnboundedConstraints())
	assert.NotNil(t, result)

	// 清空
	container.SetChildren(nil)
	result = engine.Layout(container, UnboundedConstraints())
	assert.NotNil(t, result)
}

func TestStress_SizeChanges(t *testing.T) {
	// 尺寸连续变化
	node := NewMockNode("node", 100, 100)
	engine := NewEngine()

	sizes := []struct{ w, h int }{
		{100, 100},
		{50, 50},
		{200, 200},
		{1, 1},
		{1000, 1000},
		{0, 0},
		{100, 100},
	}

	for _, size := range sizes {
		node.size = Size{Width: size.w, Height: size.h}
		result := engine.Layout(node, UnboundedConstraints())
		assert.NotNil(t, result)
	}
}

func TestStress_ConstraintChanges(t *testing.T) {
	// 约束连续变化
	node := NewMockMeasurableNode("node", 100, 100)
	engine := NewEngine()

	constraints := []Constraints{
		UnboundedConstraints(),
		NewConstraints(50, 50, 25, 25),
		NewConstraints(200, 200, 100, 100),
		NewConstraints(0, 0, 0, 0),
		NewConstraints(1, 1000, 1, 1000),
	}

	for _, c := range constraints {
		result := engine.Layout(node, c)
		assert.NotNil(t, result)
	}
}

// =============================================================================
// 布局一致性测试
// =============================================================================

func TestStress_LayoutConsistency(t *testing.T) {
	// 相同输入应产生相同输出
	node := NewMockCompositeNode("node", 100, 100)
	node.SetBorder(BorderSingle)

	child1 := NewMockMeasurableNode("child1", 48, 48)
	child2 := NewMockMeasurableNode("child2", 48, 48)
	node.SetChildren([]Node{child1, child2})

	engine := NewEngine()
	constraints := UnboundedConstraints()

	var results []*LayoutResult
	for i := 0; i < 10; i++ {
		result := engine.Layout(node, constraints)
		results = append(results, result)
	}

	// 验证所有结果一致
	for i := 1; i < len(results); i++ {
		assert.Equal(t, results[0].Root.Width, results[i].Root.Width)
		assert.Equal(t, results[0].Root.Height, results[i].Root.Height)
		assert.Equal(t, results[0].Root.X, results[i].Root.X)
		assert.Equal(t, results[0].Root.Y, results[i].Root.Y)
	}
}

// =============================================================================
// 基准测试
// =============================================================================

func BenchmarkStress_DeepNesting_20Levels(b *testing.B) {
	root := NewMockCompositeNode("root", 200, 200)
	root.SetBorder(BorderSingle)
	current := root

	for i := 1; i < 20; i++ {
		next := NewMockCompositeNode("level", 198-i*2, 198-i*2)
		next.SetBorder(BorderSingle)
		current.SetChildren([]Node{next})
		current = next
	}

	content := NewMockMeasurableNode("content", 50, 50)
	current.SetChildren([]Node{content})

	engine := NewEngine()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.Layout(root, UnboundedConstraints())
	}
}

func BenchmarkStress_ManyChildren_100(b *testing.B) {
	flex := NewMockFlexNode("flex", FlexRow)
	children := make([]Node, 100)
	for i := range children {
		children[i] = NewMockMeasurableNode("child", 2, 10)
	}
	flex.SetChildren(children)

	engine := NewEngine()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.Layout(flex, UnboundedConstraints())
	}
}

func BenchmarkStress_Grid_10x10(b *testing.B) {
	grid := NewMockGridNode("grid", 100, 50)
	
	cols := make([]GridDimension, 10)
	rows := make([]GridDimension, 10)
	for i := range cols {
		cols[i] = GridFixed(10)
		rows[i] = GridFixed(5)
	}
	grid.SetGridColumns(cols)
	grid.SetGridRows(rows)

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

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.Layout(grid, UnboundedConstraints())
	}
}

func BenchmarkStress_ComplexLayout(b *testing.B) {
	// 复杂混合布局
	border := NewMockCompositeNode("border", 100, 100)
	border.SetBorder(BorderSingle)

	grid := NewMockGridNode("grid", 98, 98)
	grid.SetGridColumns([]GridDimension{GridFlex{Factor: 1}, GridFlex{Factor: 1}})
	grid.SetGridRows([]GridDimension{GridFlex{Factor: 1}})

	flex1 := NewMockFlexNode("flex1", FlexColumn)
	flex1.SetChildren([]Node{
		NewMockMeasurableNode("c1", 49, 49),
		NewMockMeasurableNode("c2", 49, 49),
	})

	flex2 := NewMockFlexNode("flex2", FlexRow)
	flex2.SetChildren([]Node{
		NewMockMeasurableNode("c3", 24, 98),
		NewMockMeasurableNode("c4", 24, 98),
	})

	grid.SetGridCells([]GridCell{
		{Child: flex1, Row: 0, Col: 0},
		{Child: flex2, Row: 0, Col: 1},
	})

	border.SetChildren([]Node{grid})

	engine := NewEngine()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.Layout(border, UnboundedConstraints())
	}
}
