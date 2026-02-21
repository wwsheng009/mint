package layout

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// Absolute 高级定位测试
// =============================================================================

// =============================================================================
// 所有锚点测试
// =============================================================================

func TestAbsolute_Anchor_AllPositions(t *testing.T) {
	container := NewMockAbsoluteNode("container", 100, 100)

	anchors := []Anchor{
		AnchorTopLeft,
		AnchorTop,
		AnchorTopRight,
		AnchorLeft,
		AnchorCenter,
		AnchorRight,
		AnchorBottomLeft,
		AnchorBottom,
		AnchorBottomRight,
	}

	for _, anchor := range anchors {
		t.Run(anchorName(anchor), func(t *testing.T) {
			container.SetPositionStyle(AbsolutePos(10), AbsolutePos(10), nil, nil, anchor)

			child := NewMockMeasurableNode("child", 20, 20)
			container.SetChildren([]Node{child})

			engine := NewEngine()
			result := engine.Layout(container, UnboundedConstraints())

			assert.NotNil(t, result)
			assert.Len(t, result.Root.Children, 1)
		})
	}
}

func anchorName(a Anchor) string {
	names := map[Anchor]string{
		AnchorTopLeft:     "TopLeft",
		AnchorTop:         "Top",
		AnchorTopRight:    "TopRight",
		AnchorLeft:        "Left",
		AnchorCenter:      "Center",
		AnchorRight:       "Right",
		AnchorBottomLeft:  "BottomLeft",
		AnchorBottom:      "Bottom",
		AnchorBottomRight: "BottomRight",
	}
	return names[a]
}

// =============================================================================
// 边距定位测试
// =============================================================================

func TestAbsolute_PositionWithRight(t *testing.T) {
	// 使用 right 定位
	container := NewMockAbsoluteNode("container", 100, 100)
	container.SetPositionStyle(nil, AbsolutePos(10), AbsolutePos(10), nil, AnchorTopRight)

	child := NewMockMeasurableNode("child", 20, 20)
	container.SetChildren([]Node{child})

	engine := NewEngine()
	result := engine.Layout(container, UnboundedConstraints())

	assert.NotNil(t, result)
}

func TestAbsolute_PositionWithBottom(t *testing.T) {
	// 使用 bottom 定位
	container := NewMockAbsoluteNode("container", 100, 100)
	container.SetPositionStyle(AbsolutePos(10), nil, nil, AbsolutePos(10), AnchorTopLeft)

	child := NewMockMeasurableNode("child", 20, 20)
	container.SetChildren([]Node{child})

	engine := NewEngine()
	result := engine.Layout(container, UnboundedConstraints())

	assert.NotNil(t, result)
}

func TestAbsolute_PositionWithRightAndBottom(t *testing.T) {
	// 使用 right + bottom 定位
	container := NewMockAbsoluteNode("container", 100, 100)
	container.SetPositionStyle(nil, nil, AbsolutePos(10), AbsolutePos(10), AnchorBottomRight)

	child := NewMockMeasurableNode("child", 20, 20)
	container.SetChildren([]Node{child})

	engine := NewEngine()
	result := engine.Layout(container, UnboundedConstraints())

	assert.NotNil(t, result)
}

func TestAbsolute_PositionAllFour(t *testing.T) {
	// 使用所有四个边定位 (拉伸)
	container := NewMockAbsoluteNode("container", 100, 100)
	container.SetPositionStyle(AbsolutePos(10), AbsolutePos(10), AbsolutePos(10), AbsolutePos(10), AnchorTopLeft)

	child := NewMockMeasurableNode("child", 80, 80)
	container.SetChildren([]Node{child})

	engine := NewEngine()
	result := engine.Layout(container, UnboundedConstraints())

	assert.NotNil(t, result)
}

func TestAbsolute_PositionConflicts(t *testing.T) {
	// 同时设置 left/right 或 top/bottom 的冲突
	container := NewMockAbsoluteNode("container", 100, 100)
	container.SetPositionStyle(AbsolutePos(10), AbsolutePos(10), AbsolutePos(10), AbsolutePos(10), AnchorCenter)

	child := NewMockMeasurableNode("child", 20, 20)
	container.SetChildren([]Node{child})

	engine := NewEngine()
	result := engine.Layout(container, UnboundedConstraints())

	assert.NotNil(t, result)
}

// =============================================================================
// 百分比定位测试
// =============================================================================

func TestAbsolute_PercentPosition(t *testing.T) {
	// 百分比定位
	container := NewMockAbsoluteNode("container", 100, 100)
	container.SetPositionStyle(RelativePos(50), RelativePos(50), nil, nil, AnchorTopLeft)

	child := NewMockMeasurableNode("child", 20, 20)
	container.SetChildren([]Node{child})

	engine := NewEngine()
	result := engine.Layout(container, UnboundedConstraints())

	assert.NotNil(t, result)
}

func TestAbsolute_PercentPosition_Zero(t *testing.T) {
	container := NewMockAbsoluteNode("container", 100, 100)
	container.SetPositionStyle(RelativePos(0), RelativePos(0), nil, nil, AnchorTopLeft)

	child := NewMockMeasurableNode("child", 20, 20)
	container.SetChildren([]Node{child})

	engine := NewEngine()
	result := engine.Layout(container, UnboundedConstraints())

	assert.NotNil(t, result)
}

func TestAbsolute_PercentPosition_100(t *testing.T) {
	container := NewMockAbsoluteNode("container", 100, 100)
	container.SetPositionStyle(RelativePos(100), RelativePos(100), nil, nil, AnchorTopLeft)

	child := NewMockMeasurableNode("child", 20, 20)
	container.SetChildren([]Node{child})

	engine := NewEngine()
	result := engine.Layout(container, UnboundedConstraints())

	assert.NotNil(t, result)
}

func TestAbsolute_PercentPosition_Exceed100(t *testing.T) {
	// 超过 100%
	container := NewMockAbsoluteNode("container", 100, 100)
	container.SetPositionStyle(RelativePos(150), RelativePos(150), nil, nil, AnchorTopLeft)

	child := NewMockMeasurableNode("child", 20, 20)
	container.SetChildren([]Node{child})

	engine := NewEngine()
	result := engine.Layout(container, UnboundedConstraints())

	assert.NotNil(t, result)
}

func TestAbsolute_PercentPosition_Negative(t *testing.T) {
	// 负百分比
	container := NewMockAbsoluteNode("container", 100, 100)
	container.SetPositionStyle(RelativePos(-50), RelativePos(-50), nil, nil, AnchorTopLeft)

	child := NewMockMeasurableNode("child", 20, 20)
	container.SetChildren([]Node{child})

	engine := NewEngine()
	result := engine.Layout(container, UnboundedConstraints())

	assert.NotNil(t, result)
}

// =============================================================================
// 混合定位测试
// =============================================================================

func TestAbsolute_MixedPosition(t *testing.T) {
	// 混合使用固定值和百分比
	container := NewMockAbsoluteNode("container", 100, 100)
	container.SetPositionStyle(AbsolutePos(10), RelativePos(50), nil, nil, AnchorTopLeft)

	child := NewMockMeasurableNode("child", 20, 20)
	container.SetChildren([]Node{child})

	engine := NewEngine()
	result := engine.Layout(container, UnboundedConstraints())

	assert.NotNil(t, result)
}

func TestAbsolute_MixedPosition_Right(t *testing.T) {
	// 左边固定，右边百分比
	container := NewMockAbsoluteNode("container", 100, 100)
	container.SetPositionStyle(AbsolutePos(10), nil, RelativePos(20), nil, AnchorTopLeft)

	child := NewMockMeasurableNode("child", 70, 20)
	container.SetChildren([]Node{child})

	engine := NewEngine()
	result := engine.Layout(container, UnboundedConstraints())

	assert.NotNil(t, result)
}

// =============================================================================
// 锚点 + 位置组合测试
// =============================================================================

func TestAbsolute_AnchorWithPosition_Center(t *testing.T) {
	// 中心锚点 + 偏移
	container := NewMockAbsoluteNode("container", 100, 100)
	container.SetPositionStyle(AbsolutePos(10), AbsolutePos(10), nil, nil, AnchorCenter)

	child := NewMockMeasurableNode("child", 20, 20)
	container.SetChildren([]Node{child})

	engine := NewEngine()
	result := engine.Layout(container, UnboundedConstraints())

	assert.NotNil(t, result)
}

func TestAbsolute_AnchorWithPosition_BottomRight(t *testing.T) {
	// 右下锚点 + 偏移
	container := NewMockAbsoluteNode("container", 100, 100)
	container.SetPositionStyle(nil, nil, AbsolutePos(5), AbsolutePos(5), AnchorBottomRight)

	child := NewMockMeasurableNode("child", 20, 20)
	container.SetChildren([]Node{child})

	engine := NewEngine()
	result := engine.Layout(container, UnboundedConstraints())

	assert.NotNil(t, result)
}

// =============================================================================
// 多子节点绝对定位测试
// =============================================================================

func TestAbsolute_MultipleChildren_NonOverlapping(t *testing.T) {
	// 多个非重叠子节点
	container := NewMockAbsoluteNode("container", 100, 100)

	children := []Node{
		createAbsoluteChild("tl", 0, 0, 40, 40),
		createAbsoluteChild("tr", 60, 0, 40, 40),
		createAbsoluteChild("bl", 0, 60, 40, 40),
		createAbsoluteChild("br", 60, 60, 40, 40),
	}
	container.SetChildren(children)

	engine := NewEngine()
	result := engine.Layout(container, UnboundedConstraints())

	assert.NotNil(t, result)
	assert.Len(t, result.Root.Children, 4)
}

func TestAbsolute_MultipleChildren_Overlapping(t *testing.T) {
	// 多个重叠子节点
	container := NewMockAbsoluteNode("container", 100, 100)

	children := []Node{
		createAbsoluteChild("bg", 0, 0, 100, 100),  // 背景
		createAbsoluteChild("mid", 10, 10, 80, 80), // 中间层
		createAbsoluteChild("fg", 20, 20, 60, 60),  // 前景
	}
	container.SetChildren(children)

	engine := NewEngine()
	result := engine.Layout(container, UnboundedConstraints())

	assert.NotNil(t, result)
	assert.Len(t, result.Root.Children, 3)
}

func createAbsoluteChild(id string, x, y, w, h int) *MockAbsoluteNode {
	node := NewMockAbsoluteNode(id, w, h)
	node.SetPositionStyle(AbsolutePos(x), AbsolutePos(y), nil, nil, AnchorTopLeft)
	return node
}

// =============================================================================
// Absolute 嵌套测试
// =============================================================================

func TestAbsolute_Nested(t *testing.T) {
	// 嵌套绝对定位
	outer := NewMockAbsoluteNode("outer", 100, 100)
	outer.SetPositionStyle(AbsolutePos(0), AbsolutePos(0), nil, nil, AnchorTopLeft)

	middle := NewMockAbsoluteNode("middle", 80, 80)
	middle.SetPositionStyle(AbsolutePos(10), AbsolutePos(10), nil, nil, AnchorTopLeft)

	inner := NewMockAbsoluteNode("inner", 60, 60)
	inner.SetPositionStyle(AbsolutePos(10), AbsolutePos(10), nil, nil, AnchorTopLeft)

	content := NewMockMeasurableNode("content", 60, 60)
	inner.SetChildren([]Node{content})
	middle.SetChildren([]Node{inner})
	outer.SetChildren([]Node{middle})

	engine := NewEngine()
	result := engine.Layout(outer, UnboundedConstraints())

	assert.NotNil(t, result)
}

func TestAbsolute_DeepNested(t *testing.T) {
	// 深度嵌套
	root := NewMockAbsoluteNode("root", 100, 100)
	root.SetPositionStyle(AbsolutePos(0), AbsolutePos(0), nil, nil, AnchorTopLeft)

	current := root
	for i := 0; i < 10; i++ {
		next := NewMockAbsoluteNode("level", 90-i*5, 90-i*5)
		next.SetPositionStyle(AbsolutePos(5), AbsolutePos(5), nil, nil, AnchorTopLeft)
		current.SetChildren([]Node{next})
		current = next
	}

	content := NewMockMeasurableNode("content", 40, 40)
	current.SetChildren([]Node{content})

	engine := NewEngine()
	result := engine.Layout(root, UnboundedConstraints())

	assert.NotNil(t, result)
}

// =============================================================================
// Absolute 边界条件测试
// =============================================================================

func TestAbsolute_PositionOutsideContainer(t *testing.T) {
	// 位置超出容器边界
	container := NewMockAbsoluteNode("container", 100, 100)
	container.SetPositionStyle(AbsolutePos(150), AbsolutePos(150), nil, nil, AnchorTopLeft)

	child := NewMockMeasurableNode("child", 20, 20)
	container.SetChildren([]Node{child})

	engine := NewEngine()
	result := engine.Layout(container, UnboundedConstraints())

	assert.NotNil(t, result)
}

func TestAbsolute_PositionNegative(t *testing.T) {
	// 负位置
	container := NewMockAbsoluteNode("container", 100, 100)
	container.SetPositionStyle(AbsolutePos(-10), AbsolutePos(-10), nil, nil, AnchorTopLeft)

	child := NewMockMeasurableNode("child", 20, 20)
	container.SetChildren([]Node{child})

	engine := NewEngine()
	result := engine.Layout(container, UnboundedConstraints())

	assert.NotNil(t, result)
}

func TestAbsolute_LargerThanContainer(t *testing.T) {
	// 子节点大于容器
	container := NewMockAbsoluteNode("container", 50, 50)
	container.SetPositionStyle(AbsolutePos(0), AbsolutePos(0), nil, nil, AnchorTopLeft)

	child := NewMockMeasurableNode("child", 100, 100)
	container.SetChildren([]Node{child})

	engine := NewEngine()
	result := engine.Layout(container, UnboundedConstraints())

	assert.NotNil(t, result)
}

func TestAbsolute_ZeroSizeChild(t *testing.T) {
	// 零尺寸子节点
	container := NewMockAbsoluteNode("container", 100, 100)
	container.SetPositionStyle(AbsolutePos(50), AbsolutePos(50), nil, nil, AnchorCenter)

	child := NewMockMeasurableNode("child", 0, 0)
	container.SetChildren([]Node{child})

	engine := NewEngine()
	result := engine.Layout(container, UnboundedConstraints())

	assert.NotNil(t, result)
}

func TestAbsolute_ZeroSizeContainer(t *testing.T) {
	// 零尺寸容器
	container := NewMockAbsoluteNode("container", 0, 0)
	container.SetPositionStyle(AbsolutePos(0), AbsolutePos(0), nil, nil, AnchorTopLeft)

	child := NewMockMeasurableNode("child", 10, 10)
	container.SetChildren([]Node{child})

	engine := NewEngine()
	result := engine.Layout(container, UnboundedConstraints())

	assert.NotNil(t, result)
}

// =============================================================================
// Absolute + Border 组合测试
// =============================================================================

func TestAbsolute_WithBorder(t *testing.T) {
	// 绝对定位容器有 Border
	container := NewMockAbsoluteNode("container", 100, 100)
	container.SetBorder(BorderSingle)
	container.SetPositionStyle(AbsolutePos(10), AbsolutePos(10), nil, nil, AnchorTopLeft)

	child := NewMockMeasurableNode("child", 98, 98)
	container.SetChildren([]Node{child})

	engine := NewEngine()
	result := engine.Layout(container, UnboundedConstraints())

	assert.NotNil(t, result)
}

func TestAbsolute_ChildWithBorder(t *testing.T) {
	// 绝对定位的子节点有 Border
	container := NewMockAbsoluteNode("container", 100, 100)
	container.SetPositionStyle(AbsolutePos(0), AbsolutePos(0), nil, nil, AnchorTopLeft)

	child := NewMockCompositeNode("child", 50, 50)
	child.SetBorder(BorderSingle)
	container.SetChildren([]Node{child})

	engine := NewEngine()
	result := engine.Layout(container, UnboundedConstraints())

	assert.NotNil(t, result)
}

func TestAbsolute_BothWithBorder(t *testing.T) {
	// 容器和子节点都有 Border
	container := NewMockAbsoluteNode("container", 100, 100)
	container.SetBorder(BorderSingle)
	container.SetPositionStyle(AbsolutePos(10), AbsolutePos(10), nil, nil, AnchorTopLeft)

	child := NewMockCompositeNode("child", 96, 96)
	child.SetBorder(BorderSingle)
	container.SetChildren([]Node{child})

	engine := NewEngine()
	result := engine.Layout(container, UnboundedConstraints())

	assert.NotNil(t, result)
}

// =============================================================================
// Absolute + 其他布局混合测试
// =============================================================================

func TestAbsolute_InsideGrid(t *testing.T) {
	// Grid 内的 Absolute
	grid := NewMockGridNode("grid", 100, 100)
	grid.SetGridColumns([]GridDimension{GridFixed(100)})
	grid.SetGridRows([]GridDimension{GridFixed(100)})

	abs := NewMockAbsoluteNode("abs", 100, 100)
	abs.SetPositionStyle(AbsolutePos(10), AbsolutePos(10), nil, nil, AnchorTopLeft)

	child := NewMockMeasurableNode("child", 80, 80)
	abs.SetChildren([]Node{child})

	grid.SetGridCells([]GridCell{{Child: abs, Row: 0, Col: 0}})

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	assert.NotNil(t, result)
}

func TestAbsolute_InsideFlex(t *testing.T) {
	// Flex 内的 Absolute
	flex := NewMockFlexNode("flex", FlexRow)

	abs := NewMockAbsoluteNode("abs", 50, 50)
	abs.SetPositionStyle(AbsolutePos(0), AbsolutePos(0), nil, nil, AnchorTopLeft)

	child := NewMockMeasurableNode("child", 50, 50)
	abs.SetChildren([]Node{child})

	flex.SetChildren([]Node{abs})

	engine := NewEngine()
	result := engine.Layout(flex, NewConstraints(50, 50, 50, 50))

	assert.NotNil(t, result)
}

func TestAbsolute_ContainingGrid(t *testing.T) {
	// Absolute 内包含 Grid
	abs := NewMockAbsoluteNode("abs", 100, 100)
	abs.SetPositionStyle(AbsolutePos(10), AbsolutePos(10), nil, nil, AnchorTopLeft)

	grid := NewMockGridNode("grid", 80, 80)
	grid.SetGridColumns([]GridDimension{GridFixed(40), GridFixed(40)})
	grid.SetGridRows([]GridDimension{GridFixed(40), GridFixed(40)})

	cells := make([]GridCell, 0, 4)
	for r := 0; r < 2; r++ {
		for c := 0; c < 2; c++ {
			cells = append(cells, GridCell{
				Child: NewMockMeasurableNode("cell", 40, 40),
				Row:   r,
				Col:   c,
			})
		}
	}
	grid.SetGridCells(cells)

	abs.SetChildren([]Node{grid})

	engine := NewEngine()
	result := engine.Layout(abs, UnboundedConstraints())

	assert.NotNil(t, result)
}

func TestAbsolute_ContainingFlex(t *testing.T) {
	// Absolute 内包含 Flex
	abs := NewMockAbsoluteNode("abs", 100, 100)
	abs.SetPositionStyle(AbsolutePos(10), AbsolutePos(10), nil, nil, AnchorTopLeft)

	flex := NewMockFlexNode("flex", FlexColumn)
	flex.SetChildren([]Node{
		NewMockMeasurableNode("c1", 80, 40),
		NewMockMeasurableNode("c2", 80, 40),
	})

	abs.SetChildren([]Node{flex})

	engine := NewEngine()
	result := engine.Layout(abs, UnboundedConstraints())

	assert.NotNil(t, result)
}

// =============================================================================
// 基准测试
// =============================================================================

func BenchmarkAbsolute_AllAnchors(b *testing.B) {
	engine := NewEngine()

	anchors := []Anchor{
		AnchorTopLeft, AnchorTop, AnchorTopRight,
		AnchorLeft, AnchorCenter, AnchorRight,
		AnchorBottomLeft, AnchorBottom, AnchorBottomRight,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, anchor := range anchors {
			container := NewMockAbsoluteNode("container", 100, 100)
			container.SetPositionStyle(AbsolutePos(10), AbsolutePos(10), nil, nil, anchor)

			child := NewMockMeasurableNode("child", 20, 20)
			container.SetChildren([]Node{child})

			engine.Layout(container, UnboundedConstraints())
		}
	}
}

func BenchmarkAbsolute_MultipleChildren(b *testing.B) {
	engine := NewEngine()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		container := NewMockAbsoluteNode("container", 100, 100)

		children := make([]Node, 10)
		for j := range children {
			children[j] = createAbsoluteChild("c", j*10, j*10, 10, 10)
		}
		container.SetChildren(children)

		engine.Layout(container, UnboundedConstraints())
	}
}

func BenchmarkAbsolute_Nested5Levels(b *testing.B) {
	engine := NewEngine()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		root := NewMockAbsoluteNode("root", 100, 100)
		root.SetPositionStyle(AbsolutePos(0), AbsolutePos(0), nil, nil, AnchorTopLeft)

		current := root
		for j := 0; j < 5; j++ {
			next := NewMockAbsoluteNode("level", 90-j*10, 90-j*10)
			next.SetPositionStyle(AbsolutePos(5), AbsolutePos(5), nil, nil, AnchorTopLeft)
			current.SetChildren([]Node{next})
			current = next
		}

		content := NewMockMeasurableNode("content", 40, 40)
		current.SetChildren([]Node{content})

		engine.Layout(root, UnboundedConstraints())
	}
}

func BenchmarkAbsolute_PercentPosition(b *testing.B) {
	engine := NewEngine()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		container := NewMockAbsoluteNode("container", 100, 100)
		container.SetPositionStyle(RelativePos(50), RelativePos(50), nil, nil, AnchorCenter)

		child := NewMockMeasurableNode("child", 20, 20)
		container.SetChildren([]Node{child})

		engine.Layout(container, UnboundedConstraints())
	}
}
