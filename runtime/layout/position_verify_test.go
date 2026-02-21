package layout

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// 绝对定位计算验证测试
// 验证各种偏移定位是否能正确计算位置
// =============================================================================

// =============================================================================
// 基础定位验证 - Left/Top
// =============================================================================

func TestAbsolutePosition_Verify_LeftTop(t *testing.T) {
	// 容器 100x100，子节点 20x20
	// Left=10, Top=15
	style := NewAbsoluteStyle()
	style.Left = AbsolutePos(10)
	style.Top = AbsolutePos(15)
	style.Anchor = AnchorTopLeft

	x, y := style.CalculatePosition(100, 100, 20, 20)

	assert.Equal(t, 10, x, "Left offset should be 10")
	assert.Equal(t, 15, y, "Top offset should be 15")
}

func TestAbsolutePosition_Verify_LeftTopZero(t *testing.T) {
	// 左上角 (0, 0)
	style := NewAbsoluteStyle()
	style.Left = AbsolutePos(0)
	style.Top = AbsolutePos(0)
	style.Anchor = AnchorTopLeft

	x, y := style.CalculatePosition(100, 100, 20, 20)

	assert.Equal(t, 0, x)
	assert.Equal(t, 0, y)
}

// =============================================================================
// Right/Bottom 定位验证
// =============================================================================

func TestAbsolutePosition_Verify_Right(t *testing.T) {
	// 容器 100x100，子节点 20x20
	// Right=10 表示子节点右边缘距离容器右边缘10
	// X = 100 - 10 - 20 = 70
	style := NewAbsoluteStyle()
	style.Right = AbsolutePos(10)
	style.Anchor = AnchorTopLeft

	x, y := style.CalculatePosition(100, 100, 20, 20)

	assert.Equal(t, 70, x, "Right=10 should place node at x=70 (100-10-20)")
	assert.Equal(t, 0, y, "No top offset, y should be 0")
}

func TestAbsolutePosition_Verify_Bottom(t *testing.T) {
	// 容器 100x100，子节点 20x20
	// Bottom=15 表示子节点底边距离容器底边15
	// Y = 100 - 15 - 20 = 65
	style := NewAbsoluteStyle()
	style.Bottom = AbsolutePos(15)
	style.Anchor = AnchorTopLeft

	x, y := style.CalculatePosition(100, 100, 20, 20)

	assert.Equal(t, 0, x, "No left offset, x should be 0")
	assert.Equal(t, 65, y, "Bottom=15 should place node at y=65 (100-15-20)")
}

func TestAbsolutePosition_Verify_RightBottom(t *testing.T) {
	// 容器 100x100，子节点 20x20
	// Right=10, Bottom=15
	// X = 100 - 10 - 20 = 70
	// Y = 100 - 15 - 20 = 65
	style := NewAbsoluteStyle()
	style.Right = AbsolutePos(10)
	style.Bottom = AbsolutePos(15)
	style.Anchor = AnchorTopLeft

	x, y := style.CalculatePosition(100, 100, 20, 20)

	assert.Equal(t, 70, x, "Right=10 should place node at x=70")
	assert.Equal(t, 65, y, "Bottom=15 should place node at y=65")
}

func TestAbsolutePosition_Verify_RightBottomZero(t *testing.T) {
	// 右下角 (Right=0, Bottom=0)
	// 容器 100x100，子节点 20x20
	// X = 100 - 0 - 20 = 80
	// Y = 100 - 0 - 20 = 80
	style := NewAbsoluteStyle()
	style.Right = AbsolutePos(0)
	style.Bottom = AbsolutePos(0)
	style.Anchor = AnchorTopLeft

	x, y := style.CalculatePosition(100, 100, 20, 20)

	assert.Equal(t, 80, x, "Right=0 should place node at x=80")
	assert.Equal(t, 80, y, "Bottom=0 should place node at y=80")
}

// =============================================================================
// 混合定位验证 - Left + Bottom, Right + Top
// =============================================================================

func TestAbsolutePosition_Verify_LeftBottom(t *testing.T) {
	// Left=10, Bottom=15
	// 容器 100x100，子节点 20x20
	// X = 10 (from Left)
	// Y = 100 - 15 - 20 = 65 (from Bottom)
	style := NewAbsoluteStyle()
	style.Left = AbsolutePos(10)
	style.Bottom = AbsolutePos(15)
	style.Anchor = AnchorTopLeft

	x, y := style.CalculatePosition(100, 100, 20, 20)

	assert.Equal(t, 10, x, "Left=10")
	assert.Equal(t, 65, y, "Bottom=15 should place at y=65")
}

func TestAbsolutePosition_Verify_RightTop(t *testing.T) {
	// Right=10, Top=15
	// 容器 100x100，子节点 20x20
	// X = 100 - 10 - 20 = 70 (from Right)
	// Y = 15 (from Top)
	style := NewAbsoluteStyle()
	style.Right = AbsolutePos(10)
	style.Top = AbsolutePos(15)
	style.Anchor = AnchorTopLeft

	x, y := style.CalculatePosition(100, 100, 20, 20)

	assert.Equal(t, 70, x, "Right=10 should place at x=70")
	assert.Equal(t, 15, y, "Top=15")
}

// =============================================================================
// 百分比定位验证
// =============================================================================

func TestAbsolutePosition_Verify_Percent50(t *testing.T) {
	// 50% 定位
	// 容器 100x100，子节点 20x20
	// Left=50% => X = 100 * 50 / 100 = 50
	// Top=50% => Y = 100 * 50 / 100 = 50
	style := NewAbsoluteStyle()
	style.Left = RelativePos(50)
	style.Top = RelativePos(50)
	style.Anchor = AnchorTopLeft

	x, y := style.CalculatePosition(100, 100, 20, 20)

	assert.Equal(t, 50, x, "Left=50% should be 50")
	assert.Equal(t, 50, y, "Top=50% should be 50")
}

func TestAbsolutePosition_Verify_Percent25_75(t *testing.T) {
	// 25% 和 75% 定位
	// 容器 100x100，子节点 20x20
	// Left=25% => X = 25
	// Top=75% => Y = 75
	style := NewAbsoluteStyle()
	style.Left = RelativePos(25)
	style.Top = RelativePos(75)
	style.Anchor = AnchorTopLeft

	x, y := style.CalculatePosition(100, 100, 20, 20)

	assert.Equal(t, 25, x, "Left=25% should be 25")
	assert.Equal(t, 75, y, "Top=75% should be 75")
}

func TestAbsolutePosition_Verify_PercentRight(t *testing.T) {
	// 百分比 Right 定位
	// 容器 100x100，子节点 20x20
	// Right=30% => rightPos = 100 * 30 / 100 = 30
	// X = 100 - 30 - 20 = 50
	style := NewAbsoluteStyle()
	style.Right = RelativePos(30)
	style.Anchor = AnchorTopLeft

	x, y := style.CalculatePosition(100, 100, 20, 20)

	assert.Equal(t, 50, x, "Right=30% should place at x=50 (100-30-20)")
	assert.Equal(t, 0, y)
}

func TestAbsolutePosition_Verify_PercentBottom(t *testing.T) {
	// 百分比 Bottom 定位
	// 容器 100x100，子节点 20x20
	// Bottom=40% => bottomPos = 100 * 40 / 100 = 40
	// Y = 100 - 40 - 20 = 40
	style := NewAbsoluteStyle()
	style.Bottom = RelativePos(40)
	style.Anchor = AnchorTopLeft

	x, y := style.CalculatePosition(100, 100, 20, 20)

	assert.Equal(t, 0, x)
	assert.Equal(t, 40, y, "Bottom=40% should place at y=40 (100-40-20)")
}

// =============================================================================
// 锚点定位验证
// =============================================================================

func TestAbsolutePosition_Verify_AnchorTopLeft(t *testing.T) {
	// AnchorTopLeft - 无偏移调整
	style := NewAbsoluteStyle()
	style.Left = AbsolutePos(50)
	style.Top = AbsolutePos(50)
	style.Anchor = AnchorTopLeft

	x, y := style.CalculatePosition(100, 100, 20, 20)

	assert.Equal(t, 50, x, "AnchorTopLeft: no adjustment")
	assert.Equal(t, 50, y, "AnchorTopLeft: no adjustment")
}

func TestAbsolutePosition_Verify_AnchorTop(t *testing.T) {
	// AnchorTop - 水平居中
	// X = 50 - 20/2 = 40
	style := NewAbsoluteStyle()
	style.Left = AbsolutePos(50)
	style.Top = AbsolutePos(50)
	style.Anchor = AnchorTop

	x, y := style.CalculatePosition(100, 100, 20, 20)

	assert.Equal(t, 40, x, "AnchorTop: should center horizontally (50-10)")
	assert.Equal(t, 50, y, "AnchorTop: no vertical adjustment")
}

func TestAbsolutePosition_Verify_AnchorTopRight(t *testing.T) {
	// AnchorTopRight - 右对齐
	// X = 50 - 20 = 30
	style := NewAbsoluteStyle()
	style.Left = AbsolutePos(50)
	style.Top = AbsolutePos(50)
	style.Anchor = AnchorTopRight

	x, y := style.CalculatePosition(100, 100, 20, 20)

	assert.Equal(t, 30, x, "AnchorTopRight: should align right (50-20)")
	assert.Equal(t, 50, y, "AnchorTopRight: no vertical adjustment")
}

func TestAbsolutePosition_Verify_AnchorLeft(t *testing.T) {
	// AnchorLeft - 垂直居中
	// Y = 50 - 20/2 = 40
	style := NewAbsoluteStyle()
	style.Left = AbsolutePos(50)
	style.Top = AbsolutePos(50)
	style.Anchor = AnchorLeft

	x, y := style.CalculatePosition(100, 100, 20, 20)

	assert.Equal(t, 50, x, "AnchorLeft: no horizontal adjustment")
	assert.Equal(t, 40, y, "AnchorLeft: should center vertically (50-10)")
}

func TestAbsolutePosition_Verify_AnchorCenter(t *testing.T) {
	// AnchorCenter - 居中
	// X = 50 - 20/2 = 40
	// Y = 50 - 20/2 = 40
	style := NewAbsoluteStyle()
	style.Left = AbsolutePos(50)
	style.Top = AbsolutePos(50)
	style.Anchor = AnchorCenter

	x, y := style.CalculatePosition(100, 100, 20, 20)

	assert.Equal(t, 40, x, "AnchorCenter: should center horizontally (50-10)")
	assert.Equal(t, 40, y, "AnchorCenter: should center vertically (50-10)")
}

func TestAbsolutePosition_Verify_AnchorRight(t *testing.T) {
	// AnchorRight - 右对齐 + 垂直居中
	// X = 50 - 20 = 30
	// Y = 50 - 20/2 = 40
	style := NewAbsoluteStyle()
	style.Left = AbsolutePos(50)
	style.Top = AbsolutePos(50)
	style.Anchor = AnchorRight

	x, y := style.CalculatePosition(100, 100, 20, 20)

	assert.Equal(t, 30, x, "AnchorRight: should align right (50-20)")
	assert.Equal(t, 40, y, "AnchorRight: should center vertically (50-10)")
}

func TestAbsolutePosition_Verify_AnchorBottomLeft(t *testing.T) {
	// AnchorBottomLeft - 底对齐
	// Y = 50 - 20 = 30
	style := NewAbsoluteStyle()
	style.Left = AbsolutePos(50)
	style.Top = AbsolutePos(50)
	style.Anchor = AnchorBottomLeft

	x, y := style.CalculatePosition(100, 100, 20, 20)

	assert.Equal(t, 50, x, "AnchorBottomLeft: no horizontal adjustment")
	assert.Equal(t, 30, y, "AnchorBottomLeft: should align bottom (50-20)")
}

func TestAbsolutePosition_Verify_AnchorBottom(t *testing.T) {
	// AnchorBottom - 水平居中 + 底对齐
	// X = 50 - 20/2 = 40
	// Y = 50 - 20 = 30
	style := NewAbsoluteStyle()
	style.Left = AbsolutePos(50)
	style.Top = AbsolutePos(50)
	style.Anchor = AnchorBottom

	x, y := style.CalculatePosition(100, 100, 20, 20)

	assert.Equal(t, 40, x, "AnchorBottom: should center horizontally (50-10)")
	assert.Equal(t, 30, y, "AnchorBottom: should align bottom (50-20)")
}

func TestAbsolutePosition_Verify_AnchorBottomRight(t *testing.T) {
	// AnchorBottomRight - 右对齐 + 底对齐
	// X = 50 - 20 = 30
	// Y = 50 - 20 = 30
	style := NewAbsoluteStyle()
	style.Left = AbsolutePos(50)
	style.Top = AbsolutePos(50)
	style.Anchor = AnchorBottomRight

	x, y := style.CalculatePosition(100, 100, 20, 20)

	assert.Equal(t, 30, x, "AnchorBottomRight: should align right (50-20)")
	assert.Equal(t, 30, y, "AnchorBottomRight: should align bottom (50-20)")
}

// =============================================================================
// 特殊场景：完全居中（容器中心）
// =============================================================================

func TestAbsolutePosition_Verify_CenterInContainer(t *testing.T) {
	// 在容器正中心
	// 容器 100x100，子节点 20x20
	// 使用 50% + AnchorCenter
	style := NewAbsoluteStyle()
	style.Left = RelativePos(50)
	style.Top = RelativePos(50)
	style.Anchor = AnchorCenter

	x, y := style.CalculatePosition(100, 100, 20, 20)

	// 50% of 100 = 50, then center anchor subtracts half of node size
	// X = 50 - 10 = 40
	// Y = 50 - 10 = 40
	// This means the node center is at (50, 50) which is the container center
	assert.Equal(t, 40, x, "Center positioning: x should be 40")
	assert.Equal(t, 40, y, "Center positioning: y should be 40")
}

func TestAbsolutePosition_Verify_CenterUsingAbsolute(t *testing.T) {
	// 使用绝对值在容器中心
	// 容器 100x100，子节点 20x20
	// 要让子节点居中，需要 left = (100-20)/2 = 40
	style := NewAbsoluteStyle()
	style.Left = AbsolutePos(40)
	style.Top = AbsolutePos(40)
	style.Anchor = AnchorTopLeft

	x, y := style.CalculatePosition(100, 100, 20, 20)

	assert.Equal(t, 40, x)
	assert.Equal(t, 40, y)
}

// =============================================================================
// 四边拉伸定位验证
// =============================================================================

func TestAbsolutePosition_Verify_StretchAllSides(t *testing.T) {
	// Left=10, Top=10, Right=10, Bottom=10
	// 这通常用于拉伸子节点填满剩余空间
	// 但在当前实现中，Left 优先于 Right，Top 优先于 Bottom
	style := NewAbsoluteStyle()
	style.Left = AbsolutePos(10)
	style.Top = AbsolutePos(10)
	style.Right = AbsolutePos(10)
	style.Bottom = AbsolutePos(10)
	style.Anchor = AnchorTopLeft

	x, y := style.CalculatePosition(100, 100, 20, 20)

	// Left and Top take precedence
	assert.Equal(t, 10, x, "Left takes precedence over Right")
	assert.Equal(t, 10, y, "Top takes precedence over Bottom")
}

// =============================================================================
// 边界条件验证
// =============================================================================

func TestAbsolutePosition_Verify_OutOfBounds(t *testing.T) {
	// 定位超出容器边界
	// 容器 100x100，子节点 20x20
	// Left=90 会把子节点放到 x=90，但宽度20会超出
	style := NewAbsoluteStyle()
	style.Left = AbsolutePos(90)
	style.Top = AbsolutePos(90)
	style.Anchor = AnchorTopLeft

	x, y := style.CalculatePosition(100, 100, 20, 20)

	// 当前实现不会限制到边界内
	assert.Equal(t, 90, x)
	assert.Equal(t, 90, y)
}

func TestAbsolutePosition_Verify_NegativePosition(t *testing.T) {
	// 负定位
	style := NewAbsoluteStyle()
	style.Left = AbsolutePos(-10)
	style.Top = AbsolutePos(-10)
	style.Anchor = AnchorTopLeft

	x, y := style.CalculatePosition(100, 100, 20, 20)

	assert.Equal(t, -10, x, "Negative positions should be allowed")
	assert.Equal(t, -10, y, "Negative positions should be allowed")
}

func TestAbsolutePosition_Verify_LargeOffset(t *testing.T) {
	// 大偏移量
	style := NewAbsoluteStyle()
	style.Left = AbsolutePos(1000)
	style.Top = AbsolutePos(1000)
	style.Anchor = AnchorTopLeft

	x, y := style.CalculatePosition(100, 100, 20, 20)

	assert.Equal(t, 1000, x)
	assert.Equal(t, 1000, y)
}

func TestAbsolutePosition_Verify_ZeroContainer(t *testing.T) {
	// 零尺寸容器
	style := NewAbsoluteStyle()
	style.Left = AbsolutePos(10)
	style.Top = AbsolutePos(10)
	style.Anchor = AnchorTopLeft

	x, y := style.CalculatePosition(0, 0, 20, 20)

	assert.Equal(t, 10, x)
	assert.Equal(t, 10, y)
}

func TestAbsolutePosition_Verify_ZeroNode(t *testing.T) {
	// 零尺寸子节点
	style := NewAbsoluteStyle()
	style.Left = AbsolutePos(50)
	style.Top = AbsolutePos(50)
	style.Anchor = AnchorCenter

	x, y := style.CalculatePosition(100, 100, 0, 0)

	// With zero size node, anchor adjustments should still work but with 0
	assert.Equal(t, 50, x)
	assert.Equal(t, 50, y)
}

// =============================================================================
// 实际布局引擎验证
// =============================================================================

func TestAbsolutePosition_Verify_EngineLayout(t *testing.T) {
	// 使用布局引擎验证实际布局结果
	container := NewMockAbsoluteNode("container", 100, 100)
	container.SetPositionStyle(AbsolutePos(10), AbsolutePos(15), nil, nil, AnchorTopLeft)

	child := NewMockMeasurableNode("child", 20, 20)
	container.SetChildren([]Node{child})

	engine := NewEngine()
	result := engine.Layout(container, UnboundedConstraints())

	assert.NotNil(t, result)
	assert.Len(t, result.Root.Children, 1)

	childBox := result.Root.Children[0]
	// 子节点应该在容器内的 (10, 15) 位置
	assert.Equal(t, 10, childBox.X, "Child should be at x=10")
	assert.Equal(t, 15, childBox.Y, "Child should be at y=15")
}

func TestAbsolutePosition_Verify_EngineLayoutRightBottom(t *testing.T) {
	// 使用 Right/Bottom 定位
	container := NewMockAbsoluteNode("container", 100, 100)
	container.SetPositionStyle(nil, nil, AbsolutePos(10), AbsolutePos(15), AnchorTopLeft)

	child := NewMockMeasurableNode("child", 20, 20)
	container.SetChildren([]Node{child})

	engine := NewEngine()
	result := engine.Layout(container, UnboundedConstraints())

	assert.NotNil(t, result)
	assert.Len(t, result.Root.Children, 1)

	childBox := result.Root.Children[0]
	// Right=10, Bottom=15
	// X = 100 - 10 - 20 = 70
	// Y = 100 - 15 - 20 = 65
	assert.Equal(t, 70, childBox.X, "Child should be at x=70")
	assert.Equal(t, 65, childBox.Y, "Child should be at y=65")
}

func TestAbsolutePosition_Verify_EngineLayoutCenter(t *testing.T) {
	// 居中定位
	container := NewMockAbsoluteNode("container", 100, 100)
	container.SetPositionStyle(RelativePos(50), RelativePos(50), nil, nil, AnchorCenter)

	child := NewMockMeasurableNode("child", 20, 20)
	container.SetChildren([]Node{child})

	engine := NewEngine()
	result := engine.Layout(container, UnboundedConstraints())

	assert.NotNil(t, result)
	assert.Len(t, result.Root.Children, 1)

	childBox := result.Root.Children[0]
	// 50% of 100 = 50, center anchor subtracts half of node size
	// X = 50 - 10 = 40
	// Y = 50 - 10 = 40
	assert.Equal(t, 40, childBox.X, "Child should be centered at x=40")
	assert.Equal(t, 40, childBox.Y, "Child should be centered at y=40")
}

// =============================================================================
// 表格驱动测试 - 综合验证
// =============================================================================

func TestAbsolutePosition_Verify_Table(t *testing.T) {
	tests := []struct {
		name      string
		left      PositionValue
		top       PositionValue
		right     PositionValue
		bottom    PositionValue
		anchor    Anchor
		expectX   int
		expectY   int
	}{
		// 基础定位
		{"Left=0, Top=0", AbsolutePos(0), AbsolutePos(0), nil, nil, AnchorTopLeft, 0, 0},
		{"Left=10, Top=20", AbsolutePos(10), AbsolutePos(20), nil, nil, AnchorTopLeft, 10, 20},
		{"Right=10", nil, nil, AbsolutePos(10), nil, AnchorTopLeft, 70, 0},
		{"Bottom=15", nil, nil, nil, AbsolutePos(15), AnchorTopLeft, 0, 65},
		{"Right=10, Bottom=15", nil, nil, AbsolutePos(10), AbsolutePos(15), AnchorTopLeft, 70, 65},

		// 混合定位
		{"Left=10, Bottom=15", AbsolutePos(10), nil, nil, AbsolutePos(15), AnchorTopLeft, 10, 65},
		{"Right=10, Top=20", nil, AbsolutePos(20), AbsolutePos(10), nil, AnchorTopLeft, 70, 20},

		// 百分比定位
		{"Left=25%", RelativePos(25), nil, nil, nil, AnchorTopLeft, 25, 0},
		{"Top=75%", nil, RelativePos(75), nil, nil, AnchorTopLeft, 0, 75},
		{"Right=30%", nil, nil, RelativePos(30), nil, AnchorTopLeft, 50, 0},
		{"Bottom=40%", nil, nil, nil, RelativePos(40), AnchorTopLeft, 0, 40},

		// 锚点定位
		{"Center anchor", RelativePos(50), RelativePos(50), nil, nil, AnchorCenter, 40, 40},
		{"TopRight anchor", AbsolutePos(50), AbsolutePos(50), nil, nil, AnchorTopRight, 30, 50},
		{"BottomRight anchor", AbsolutePos(50), AbsolutePos(50), nil, nil, AnchorBottomRight, 30, 30},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			style := NewAbsoluteStyle()
			style.Left = tt.left
			style.Top = tt.top
			style.Right = tt.right
			style.Bottom = tt.bottom
			style.Anchor = tt.anchor

			x, y := style.CalculatePosition(100, 100, 20, 20)

			assert.Equal(t, tt.expectX, x, "X position mismatch")
			assert.Equal(t, tt.expectY, y, "Y position mismatch")
		})
	}
}

// =============================================================================
// 基准测试
// =============================================================================

func BenchmarkAbsolutePosition_Calculate(b *testing.B) {
	style := NewAbsoluteStyle()
	style.Left = RelativePos(50)
	style.Top = RelativePos(50)
	style.Anchor = AnchorCenter

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		style.CalculatePosition(100, 100, 20, 20)
	}
}

func BenchmarkAbsolutePosition_AllAnchors(b *testing.B) {
	anchors := []Anchor{
		AnchorTopLeft, AnchorTop, AnchorTopRight,
		AnchorLeft, AnchorCenter, AnchorRight,
		AnchorBottomLeft, AnchorBottom, AnchorBottomRight,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, anchor := range anchors {
			style := NewAbsoluteStyle()
			style.Left = RelativePos(50)
			style.Top = RelativePos(50)
			style.Anchor = anchor
			style.CalculatePosition(100, 100, 20, 20)
		}
	}
}
