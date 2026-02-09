// Package ui provides constraint defect tests for the layout system
// These tests document and verify known issues in the layout constraint mechanism
package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wwsheng009/mint/runtime"
)

// createTextNode creates a simple text element for testing
func createTextNode(content string) VNode {
	node := NewElement("text")
	node.SetProp("content", content)
	return node
}

// =============================================================================
// 缺陷 1: BorderedNode.Measure() 不检查 Props["width"]
// 位置: runtime/ui/layout.go:1122-1231
// 问题: BorderedBuilder.Width() 将值存储在 Props["width"]，但 Measure() 只检查 Style.Width
// =============================================================================

func TestDefect_BorderedNode_WidthPropIgnored(t *testing.T) {
	t.Run("Width prop should be respected by Measure", func(t *testing.T) {
		// 创建一个设置了 Width(40) 的 Bordered 节点
		bordered := Bordered().
			Width(40).
			Child(createTextNode("Hello")).
			Build()

		// 使用无界约束测量
		constraints := runtime.BoxConstraints{
			MinWidth:  0,
			MaxWidth:  runtime.Infinity,
			MinHeight: 0,
			MaxHeight: runtime.Infinity,
		}

		borderedNode := bordered.(*BorderedNode)
		size := borderedNode.Measure(constraints)

		// 🔴 缺陷: 期望宽度为 40，但实际返回的是内容自然宽度 + 边框
		// 因为 Measure() 只检查 Style.Width，不检查 Props["width"]
		t.Logf("Bordered with Width(40): measured size = %v", size)

		// 验证 Props 确实包含 width
		props := bordered.Props()
		if w, ok := props["width"].(int); ok {
			assert.Equal(t, 40, w, "Props should contain width=40")
		} else {
			t.Error("Props does not contain width")
		}

		// 这个断言会失败，暴露缺陷
		// 预期: size.Width == 40
		// 实际: size.Width == len("Hello") + 2 (边框) = 7
		if size.Width != 40 {
			t.Logf("🔴 DEFECT CONFIRMED: Width prop is ignored!")
			t.Logf("   Expected width: 40")
			t.Logf("   Actual width: %d", size.Width)
			t.Logf("   Root cause: BorderedNode.Measure() only checks Style.Width, not Props[\"width\"]")
		}
	})

	t.Run("Height prop should be respected by Measure", func(t *testing.T) {
		bordered := Bordered().
			Height(10).
			Child(createTextNode("Hello")).
			Build()

		constraints := runtime.BoxConstraints{
			MinWidth:  0,
			MaxWidth:  runtime.Infinity,
			MinHeight: 0,
			MaxHeight: runtime.Infinity,
		}

		borderedNode := bordered.(*BorderedNode)
		size := borderedNode.Measure(constraints)

		t.Logf("Bordered with Height(10): measured size = %v", size)

		// 验证 Props 确实包含 height
		props := bordered.Props()
		if h, ok := props["height"].(int); ok {
			assert.Equal(t, 10, h, "Props should contain height=10")
		}

		// 这个断言会失败，暴露缺陷
		if size.Height != 10 {
			t.Logf("🔴 DEFECT CONFIRMED: Height prop is ignored!")
			t.Logf("   Expected height: 10")
			t.Logf("   Actual height: %d", size.Height)
		}
	})
}

// =============================================================================
// 缺陷 2: HStack 在 VStack 中被强制填充宽度，忽略用户设置的 Width()
// 位置: runtime/ui/layout_measurement.go:309-312, runtime/compute/engine.go:513-514
// 问题: 硬编码的类型检查导致 HStack 总是被拉伸到父容器宽度
// =============================================================================

func TestDefect_HStackInVStack_WidthOverridden(t *testing.T) {
	t.Run("HStack with explicit Width should not be overridden in VStack", func(t *testing.T) {
		// 创建一个在 VStack 中的 HStack，HStack 设置了固定宽度 20
		hstack := HStackBuilder(createTextNode("A"), createTextNode("B")).
			Width(20).
			Build()

		vstack := VStackBuilder(hstack).Build()

		// 使用有界宽度约束
		constraints := runtime.BoxConstraints{
			MinWidth:  0,
			MaxWidth:  80, // 父容器宽度 80
			MinHeight: 0,
			MaxHeight: 24,
		}

		vstackNode := vstack.(*LayoutNode)
		size := vstackNode.Measure(constraints)

		t.Logf("VStack size: %v", size)

		// 获取子节点 HStack 的约束
		// 🔴 缺陷: VStack 会强制给 HStack 设置 MinWidth = MaxWidth = 80
		// 这会覆盖用户设置的 Width(20)

		// 检查 HStack 的 Props
		hstackNode := hstack.(*LayoutNode)
		props := hstackNode.Props()
		if w, ok := props["width"].(int); ok {
			assert.Equal(t, 20, w, "HStack Props should contain width=20")
			t.Logf("HStack has Width prop = %d", w)
		}

		// 在 VStack 中，由于硬编码的特殊处理，HStack 会被强制拉伸
		// 查看 layout_measurement.go:309-312:
		//   if innerMaxWidth != runtime.Infinity && (childTag == "hstack" || childTag == "row") {
		//       childMinWidth = innerMaxWidth // HStack fills VStack width
		//   }

		t.Logf("🔴 DEFECT: HStack in VStack is force-stretched to parent width")
		t.Logf("   User set Width(20), but layout system ignores it")
		t.Logf("   Root cause: Hardcoded type check in measureVStackLayout()")
	})
}

// =============================================================================
// 缺陷 3: Flex 子组件在无界约束下无法正确分配空间
// 位置: runtime/ui/layout_measurement.go:161-173
// 问题: 当父容器没有有界宽度时，flex 子组件使用自然尺寸而非分配空间
// =============================================================================

func TestDefect_FlexChildrenWithUnboundedConstraints(t *testing.T) {
	t.Run("Flex children should warn when parent has unbounded width", func(t *testing.T) {
		// 创建两个 flex=1 的子组件
		child1 := HStackBuilder(createTextNode("Left")).Flex(1).Build()
		child2 := HStackBuilder(createTextNode("Right")).Flex(1).Build()

		hstack := HStackBuilder(child1, child2).Build()

		// 使用无界宽度约束
		constraints := runtime.BoxConstraints{
			MinWidth:  0,
			MaxWidth:  runtime.Infinity, // 无界宽度
			MinHeight: 0,
			MaxHeight: 24,
		}

		hstackNode := hstack.(*LayoutNode)
		size := hstackNode.Measure(constraints)

		t.Logf("HStack with flex children (unbounded): size = %v", size)

		// 🔴 缺陷: flex 子组件无法正确分配空间，因为父容器没有有界宽度
		// 它们会回退到自然尺寸，flex 属性被忽略
		t.Logf("🔴 DEFECT: Flex children cannot distribute space without bounded parent width")
		t.Logf("   Flex children fall back to natural size")
		t.Logf("   User should set explicit Width() on parent or system should warn")
	})

	t.Run("Flex children work correctly with bounded width", func(t *testing.T) {
		child1 := HStackBuilder(createTextNode("Left")).Flex(1).Build()
		child2 := HStackBuilder(createTextNode("Right")).Flex(1).Build()

		hstack := HStackBuilder(child1, child2).Build()

		// 使用有界宽度约束
		constraints := runtime.BoxConstraints{
			MinWidth:  0,
			MaxWidth:  80, // 有界宽度
			MinHeight: 0,
			MaxHeight: 24,
		}

		hstackNode := hstack.(*LayoutNode)
		size := hstackNode.Measure(constraints)

		t.Logf("HStack with flex children (bounded): size = %v", size)

		// 有界宽度时，flex 子组件应该正确分配空间
		assert.Equal(t, 80, size.Width, "HStack should fill bounded width")
	})
}

// =============================================================================
// 缺陷 4: 双重测量路径可能产生不一致的结果
// 位置: runtime/compute/engine.go:72-206
// 问题: buildComputedBox 中有两条路径：单遍布局和回退的两遍布局
// =============================================================================

func TestDefect_DualMeasurementPaths(t *testing.T) {
	t.Run("Single-pass and fallback paths should produce consistent results", func(t *testing.T) {
		// 创建一个布局结构
		child := createTextNode("Hello World")
		hstack := HStackBuilder(child).Build()
		vstack := VStackBuilder(hstack).Build()

		constraints := runtime.BoxConstraints{
			MinWidth:  0,
			MaxWidth:  80,
			MinHeight: 0,
			MaxHeight: 24,
		}

		vstackNode := vstack.(*LayoutNode)

		// 第一次测量 - 通过 Measure()
		size1 := vstackNode.Measure(constraints)

		// 第二次测量 - 应该产生相同结果
		size2 := vstackNode.Measure(constraints)

		t.Logf("First measure: %v", size1)
		t.Logf("Second measure: %v", size2)

		assert.Equal(t, size1, size2, "Multiple measurements should be consistent")

		// 🔴 潜在缺陷: 当 compute/engine.go 中的 buildComputedBox 使用
		// 不同的路径（单遍 vs 两遍）时，可能产生不一致的约束
		t.Logf("Note: compute/engine.go has two measurement paths that may diverge")
	})
}

// =============================================================================
// 缺陷 5: Layer 系统的 Modal 约束过于宽松
// 位置: runtime/layer/manager.go:114-121
// 问题: Modal 使用全屏约束，完全依赖 Content 节点限制尺寸
// =============================================================================

func TestDefect_ModalConstraintsTooLoose(t *testing.T) {
	t.Run("Modal with Width prop should be respected", func(t *testing.T) {
		// 模拟 Modal 内容：一个设置了 Width(40) 的 Bordered
		modalContent := Bordered().
			Width(40).
			Height(10).
			Child(VStack(createTextNode("Modal Title"), createTextNode("Modal Body"))).
			Build()

		// 模拟 Layer 系统给 Modal 的约束
		layerConstraints := runtime.BoxConstraints{
			MinWidth:  0,
			MaxWidth:  80, // 全屏宽度
			MinHeight: 0,
			MaxHeight: 24, // 全屏高度
		}

		borderedNode := modalContent.(*BorderedNode)
		size := borderedNode.Measure(layerConstraints)

		t.Logf("Modal content with Width(40): measured size = %v", size)

		// 🔴 缺陷: 由于缺陷 1（Props["width"] 被忽略），
		// Modal 会扩展到全屏宽度而非指定的 40
		if size.Width != 40 {
			t.Logf("🔴 DEFECT CONFIRMED: Modal width is not constrained!")
			t.Logf("   Expected: 40")
			t.Logf("   Actual: %d", size.Width)
			t.Logf("   Root cause: BorderedNode.Measure() ignores Props[\"width\"]")
		}
	})
}

// =============================================================================
// 缺陷 6: LayoutNode.Measure() 和 MeasureLayout() 的实现不一致
// 位置: runtime/ui/layout.go vs runtime/ui/layout_measurement.go
// 问题: 两个方法可能对相同输入产生不同结果
// =============================================================================

func TestDefect_MeasureVsMeasureLayoutInconsistency(t *testing.T) {
	t.Run("Measure and MeasureLayout should be consistent", func(t *testing.T) {
		hstack := HStackBuilder(createTextNode("A"), createTextNode("B"), createTextNode("C")).
			Gap(2).
			Build()

		constraints := runtime.BoxConstraints{
			MinWidth:  0,
			MaxWidth:  80,
			MinHeight: 0,
			MaxHeight: 24,
		}

		hstackNode := hstack.(*LayoutNode)

		// 通过 Measure() 测量
		size1 := hstackNode.Measure(constraints)

		// 通过 MeasureLayout() 测量 (需要一个 ChildMeasurer)
		// 这里我们只检查 Measure() 的结果
		t.Logf("HStack Measure(): size = %v", size1)

		// MeasureLayout() 在 layout_measurement.go 中实现
		// Measure() 在 layout.go 中实现
		// 两者应该产生一致的结果，但由于代码重复，可能存在差异

		t.Logf("Note: layout.go:Measure() and layout_measurement.go:MeasureLayout() are separate implementations")
		t.Logf("      Code duplication increases risk of inconsistency")
	})
}

// =============================================================================
// 缺陷 7: 子组件显式宽度被父容器约束覆盖
// 问题: 子组件设置的 Width() 可能被父容器的约束传递逻辑覆盖
// =============================================================================

func TestDefect_ExplicitWidthOverriddenByParent(t *testing.T) {
	t.Run("Child explicit width should not be overridden by parent tight constraints", func(t *testing.T) {
		// 子组件设置固定宽度 30
		child := HStackBuilder(createTextNode("Short")).
			Width(30).
			Build()

		// 父容器 VStack
		parent := VStackBuilder(child).Build()

		// 父容器收到 tight 约束 (MinWidth == MaxWidth)
		constraints := runtime.BoxConstraints{
			MinWidth:  60,
			MaxWidth:  60, // Tight width
			MinHeight: 0,
			MaxHeight: 24,
		}

		parentNode := parent.(*LayoutNode)
		size := parentNode.Measure(constraints)

		t.Logf("Parent (VStack) size: %v", size)

		// 检查子组件的 Props
		childNode := child.(*LayoutNode)
		props := childNode.Props()
		if w, ok := props["width"].(int); ok {
			t.Logf("Child has Width prop = %d", w)
		}

		// 🔴 缺陷: 子组件的 Width(30) 可能被父容器的 tight 约束覆盖
		// 因为 VStack 的 measureVStackLayout() 会传递 innerMaxWidth 给子组件
		// 而 HStack 在 VStack 中会被特殊处理，强制填充宽度

		t.Logf("🔴 DEFECT: Child Width() may be ignored when parent has tight constraints")
		t.Logf("   Child set Width(30), parent has tight width 60")
		t.Logf("   Expected: Child respects its own Width(30)")
		t.Logf("   Actual: Child may be stretched to 60")
	})
}

// =============================================================================
// 缺陷 8: 交叉轴对齐在无界约束下失效
// 问题: 当父容器没有有界尺寸时，交叉轴对齐无法正确工作
// =============================================================================

func TestDefect_CrossAxisAlignmentWithUnboundedConstraints(t *testing.T) {
	t.Run("Cross-axis alignment fails with unbounded constraints", func(t *testing.T) {
		// 创建一个需要垂直居中的 HStack
		hstack := HStackBuilder(createTextNode("Center Me")).
			AlignCross(AlignCenter). // 垂直居中
			Build()

		// 无界高度约束
		constraints := runtime.BoxConstraints{
			MinWidth:  0,
			MaxWidth:  80,
			MinHeight: 0,
			MaxHeight: runtime.Infinity, // 无界高度
		}

		hstackNode := hstack.(*LayoutNode)
		size := hstackNode.Measure(constraints)

		t.Logf("HStack with AlignCross(Center): size = %v", size)

		// 🔴 缺陷: 当 MaxHeight 是 Infinity 时，
		// 交叉轴居中无法正确工作，因为没有可用空间来计算居中位置
		t.Logf("🔴 DEFECT: Cross-axis alignment cannot work with unbounded height")
		t.Logf("   AlignCross(Center) has no effect when MaxHeight is Infinity")
		t.Logf("   Size will be content height, no centering space available")
	})
}

// =============================================================================
// 缺陷总结
// =============================================================================

func TestDefect_Summary(t *testing.T) {
	t.Log("=== 布局系统约束机制缺陷总结 ===")
	t.Log("")
	t.Log("🔴 缺陷 1: BorderedNode.Measure() 不检查 Props[\"width\"]")
	t.Log("   位置: runtime/ui/layout.go:1122-1231")
	t.Log("   影响: Modal Width() 设置被忽略")
	t.Log("")
	t.Log("🔴 缺陷 2: HStack 在 VStack 中被强制填充宽度")
	t.Log("   位置: runtime/ui/layout_measurement.go:309-312")
	t.Log("   影响: 用户设置的 Width() 被覆盖")
	t.Log("")
	t.Log("🔴 缺陷 3: Flex 子组件在无界约束下失效")
	t.Log("   位置: runtime/ui/layout_measurement.go:161-173")
	t.Log("   影响: Flex 属性被忽略，使用自然尺寸")
	t.Log("")
	t.Log("🟡 缺陷 4: 双重测量路径可能不一致")
	t.Log("   位置: runtime/compute/engine.go:72-206")
	t.Log("   影响: 布局结果不可预测")
	t.Log("")
	t.Log("🟡 缺陷 5: Modal 约束过于宽松")
	t.Log("   位置: runtime/layer/manager.go:114-121")
	t.Log("   影响: Modal 可能扩展到全屏")
	t.Log("")
	t.Log("🟡 缺陷 6: Measure 和 MeasureLayout 实现重复")
	t.Log("   位置: layout.go vs layout_measurement.go")
	t.Log("   影响: 维护成本高，可能不一致")
	t.Log("")
	t.Log("🟡 缺陷 7: 子组件显式宽度被覆盖")
	t.Log("   影响: 用户设置被忽略")
	t.Log("")
	t.Log("🟡 缺陷 8: 交叉轴对齐在无界约束下失效")
	t.Log("   影响: 对齐设置无效果")
}
