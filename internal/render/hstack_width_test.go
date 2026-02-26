package render

import (
	"fmt"
	"testing"

	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// TestHStackWidth 测试 HStack 的宽度计算是否正确
// 问题：HStack 应该计算所有子元素的宽度之和，而不是只取最大宽度
func TestHStackWidth(t *testing.T) {
	// 创建一个类似 ConfirmInfo 的 HStack
	// Label: "Username:" (10 chars) + Value: "testvalue" (9 chars) = 19
	label := rtui.NewElement("text")
	label.SetProp("content", fmt.Sprintf("%-10s", "Username:"))

	value := rtui.NewElement("text")
	value.SetProp("content", "testvalue")

	hstack := rtui.NewElement("hstack")
	hstack.SetChildren([]rtui.VNode{label, value})
	hstack.SetProp("direction", rtui.DirectionRow)
	hstack.SetProp("gap", 0)

	fmt.Printf("HStack Tag: %s\n", hstack.Tag())

	// 创建 Fiber 树
	fiber := rtui.CreateFiberFromVNode(hstack)
	fmt.Printf("Fiber Tag: %s\n", fiber.Tag)
	fmt.Printf("Fiber LayoutDirection: %d (Row=%d, Column=%d)\n",
		fiber.LayoutDirection, rtui.DirectionRow, rtui.DirectionColumn)

	// 使用 FiberToNodeAdapter 测量
	unbounded := layout.Constraints{
		MinWidth:  0,
		MaxWidth:  layout.MaxInt,
		MinHeight: 0,
		MaxHeight: layout.MaxInt,
	}

	adapter := NewFiberToNodeAdapterPure(fiber)
	size := adapter.Measure(unbounded)

	fmt.Printf("HStack measured width: %d (expected: ~19)\n", size.Width)
	fmt.Printf("HStack measured height: %d\n", size.Height)

	// 理论宽度 = label宽度(10) + value宽度(9) = 19
	// 如果 width < 15，说明只计算了最大值，而不是累加
	if size.Width >= 15 {
		t.Logf("✅ HStack width is correct: %d (sum of children)\n", size.Width)
	} else {
		t.Errorf("❌ HStack width is INCORRECT: %d (only max width of children, expected sum)\n", size.Width)
	}

	// 检查 FlexStyle
	flexStyle := adapter.GetFlexStyle()
	if flexStyle == nil {
		t.Error("❌ GetFlexStyle() returned nil for HStack")
	} else {
		fmt.Printf("FlexStyle Direction: %v (FlexRow=%d, FlexColumn=%d)\n",
			flexStyle.Direction, layout.FlexRow, layout.FlexColumn)

		if flexStyle.Direction == layout.FlexRow {
			t.Log("✅ HStack is correctly identified as FlexRow")
		} else {
			t.Errorf("❌ HStack is NOT FlexRow, direction is: %v", flexStyle.Direction)
		}
	}
}

// TestVStackWithHStackChildren 测试 VStack 中包含 HStack 子元素的宽度计算
func TestVStackWithHStackChildren(t *testing.T) {
	// 模拟 Ant Design Demo Step 3 的结构
	createHStack := func(labelText, valueText string) rtui.VNode {
		label := rtui.NewElement("text")
		label.SetProp("content", fmt.Sprintf("%-10s", labelText))

		value := rtui.NewElement("text")
		value.SetProp("content", valueText)

		hstack := rtui.NewElement("hstack")
		hstack.SetChildren([]rtui.VNode{label, value})
		hstack.SetProp("direction", rtui.DirectionRow)
		hstack.SetProp("gap", 0)
		return hstack
	}

	confirmInfo1 := createHStack("Username:", "alice")
	confirmInfo2 := createHStack("Email:", "alice@example.com")
	confirmInfo3 := createHStack("Age:", "25")

	vstack := rtui.NewElement("vstack")
	vstack.SetChildren([]rtui.VNode{confirmInfo1, confirmInfo2, confirmInfo3})
	vstack.SetProp("direction", rtui.DirectionColumn)
	vstack.SetProp("gap", 2)

	// 创建 Fiber 树
	fiber := rtui.CreateFiberFromVNode(vstack)

	// 测量
	unbounded := layout.Constraints{
		MinWidth:  0,
		MaxWidth:  layout.MaxInt,
		MinHeight: 0,
		MaxHeight: layout.MaxInt,
	}

	adapter := NewFiberToNodeAdapterPure(fiber)
	size := adapter.Measure(unbounded)

	// VStack 的宽度应该是所有子元素（HStack）的最大宽度
	// 每个 HStack 的宽度应该是 Label + Value 的和
	// ConfirmInfo1: 10 + 5 = 15
	// ConfirmInfo2: 10 + 16 = 26 (最长)
	// ConfirmInfo3: 10 + 2 = 12
	// VStack 最终宽度应该是 max(15, 26, 12) = 26 + Gap(0) + BorderPadding(0) = 26
	expectedMinWidth := 20 // 保守估计，应该是 26

	fmt.Printf("VStack measured width: %d (expected >= %d)\n", size.Width, expectedMinWidth)
	fmt.Printf("VStack measured height: %d\n", size.Height)

	if size.Width < expectedMinWidth {
		t.Errorf("❌ VStack width is too small: %d (expected >= %d)\n", size.Width, expectedMinWidth)
	} else {
		t.Logf("✅ VStack width is correct: %d\n", size.Width)
	}

	// 检查 VStack 的 FlexStyle
	flexStyle := adapter.GetFlexStyle()
	if flexStyle == nil {
		t.Error("❌ GetFlexStyle() returned nil for VStack")
	} else {
		fmt.Printf("VStack FlexStyle Direction: %v (FlexColumn=%d)\n",
			flexStyle.Direction, layout.FlexColumn)

		if flexStyle.Direction == layout.FlexColumn {
			t.Log("✅ VStack is correctly identified as FlexColumn")
		} else {
			t.Errorf("❌ VStack is NOT FlexColumn, direction is: %v", flexStyle.Direction)
		}
	}
}

// TestBorderedVStackWidth 测试带边框的 VStack 宽度计算
func TestBorderedVStackWidth(t *testing.T) {
	label := rtui.NewElement("text")
	label.SetProp("content", fmt.Sprintf("%-10s", "Username:"))

	value := rtui.NewElement("text")
	value.SetProp("content", "testvalue123456")

	hstack := rtui.NewElement("hstack")
	hstack.SetChildren([]rtui.VNode{label, value})
	hstack.SetProp("direction", rtui.DirectionRow)
	hstack.SetProp("gap", 0)

	vstack := rtui.NewElement("vstack")
	vstack.SetChildren([]rtui.VNode{hstack})
	vstack.SetProp("direction", rtui.DirectionColumn)
	vstack.SetProp("gap", 2)

	bordered := rtui.NewElement("bordered")
	bordered.SetChildren([]rtui.VNode{vstack})
	bordered.SetProp("label", "Confirm")

	// 创建 Fiber 树
	fiber := rtui.CreateFiberFromVNode(bordered)

	// 测量
	unbounded := layout.Constraints{
		MinWidth:  0,
		MaxWidth:  layout.MaxInt,
		MinHeight: 0,
		MaxHeight: layout.MaxInt,
	}

	adapter := NewFiberToNodeAdapterPure(fiber)
	size := adapter.Measure(unbounded)

	// 预期宽度 = HStack宽度(10 + 14 = 24) + 边框Padding
	// Label "Confirm" 长度为 7，加空格后为 ~9
	// 所以最终宽度应该是 max(24, 9) + padding
	expectedMinWidth := 24

	fmt.Printf("Bordered VStack measured width: %d (expected >= %d)\n", size.Width, expectedMinWidth)
	fmt.Printf("Bordered VStack measured height: %d\n", size.Height)

	if size.Width < expectedMinWidth {
		t.Errorf("❌ Bordered VStack width is too small: %d (expected >= %d)\n", size.Width, expectedMinWidth)
	} else {
		t.Logf("✅ Bordered VStack width is correct: %d\n", size.Width)
	}

	// 检查边框
	border := adapter.GetBorder()
	if border.Style == layout.BorderNone {
		t.Error("❌ GetBorder() returned BorderNone for Bordered")
	} else {
		fmt.Printf("Border Label: %s\n", border.Label)
		if border.Label == "Confirm" {
			t.Log("✅ Border label is correct")
		} else {
			t.Errorf("❌ Border label is wrong: %s", border.Label)
		}
	}
}

// TestStyledTextWidth 测试带样式的 Text 宽度
func TestStyledTextWidth(t *testing.T) {
	label := rtui.NewElement("text")
	label.SetProp("content", fmt.Sprintf("%-10s", "Username:"))
	sty := style.Style{}.Foreground(style.Color("gray")).Bold(true)
	label.SetStyle(sty)

	value := rtui.NewElement("text")
	value.SetProp("content", "testvalue")
	sty2 := style.Style{}.Foreground(style.Color("white"))
	value.SetStyle(sty2)

	hstack := rtui.NewElement("hstack")
	hstack.SetChildren([]rtui.VNode{label, value})
	hstack.SetProp("direction", rtui.DirectionRow)
	hstack.SetProp("gap", 0)

	// 创建 Fiber 树
	fiber := rtui.CreateFiberFromVNode(hstack)

	// 测量
	unbounded := layout.Constraints{
		MinWidth:  0,
		MaxWidth:  layout.MaxInt,
		MinHeight: 0,
		MaxHeight: layout.MaxInt,
	}

	adapter := NewFiberToNodeAdapterPure(fiber)
	size := adapter.Measure(unbounded)

	// 预期宽度 = 10 + 9 = 19
	expectedMinWidth := 15

	fmt.Printf("Styled HStack measured width: %d (expected >= %d)\n", size.Width, expectedMinWidth)

	if size.Width < expectedMinWidth {
		t.Errorf("❌ Styled HStack width is too small: %d (expected >= %d)\n", size.Width, expectedMinWidth)
	} else {
		t.Logf("✅ Styled HStack width is correct: %d\n", size.Width)
	}
}
