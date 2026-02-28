package ui_test

import (
	"testing"

	"github.com/wwsheng009/mint/internal/render"
	"github.com/wwsheng009/mint/runtime/layout"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// TestFiberAdapter_FlexAutoFill 测试 flex 容器自动填充父容器宽度
func TestFiberAdapter_FlexAutoFill(t *testing.T) {
	t.Log("=== 测试 flex 容器自动填充 ===")

	// 创建 HStack fiber（不直接依赖 app 包，模拟已有 flex 子节点）
	flexChildren := []rtui.Fiber{
		{Tag: "button", LayoutFlex: 1, Props: map[string]interface{}{"content": "Left"}},
		{Tag: "button", LayoutFlex: 1, Props: map[string]interface{}{"content": "Center"}},
		{Tag: "button", LayoutFlex: 1, Props: map[string]interface{}{"content": "Right"}},
	}

	hstackFiber := &rtui.Fiber{
		Tag:            "hstack",
		LayoutGap:      1,
		LayoutDirection: rtui.DirectionRow,
		Child:          &flexChildren[0],
	}
	flexChildren[0].Sibling = &flexChildren[1]
	flexChildren[1].Sibling = &flexChildren[2]

	// 创建 Adapter
	adapter := render.NewFiberToNodeAdapterPure(hstackFiber)

	// 场景 1: 无界约束 - 使用自然宽度
	unboundedConstraints := layout.UnboundedConstraints()
	naturalSize := adapter.Measure(unboundedConstraints)
	t.Logf("无界约束: 尺寸 = %dx%d", naturalSize.Width, naturalSize.Height)

	if naturalSize.Width <= 0 {
		t.Errorf("无界约束宽度应该 > 0, got %d", naturalSize.Width)
	}

	// 场景 2: 有界约束 (80) - 应该自动填充到父容器宽度
	boundedConstraints := layout.NewConstraints(0, 80, 0, 15)
	filledSize := adapter.Measure(boundedConstraints)
	t.Logf("有界约束 (80): 尺寸 = %dx%d", filledSize.Width, filledSize.Height)

	// 有界约束时，HStack 应该填充到父容器宽度 (80)
	if filledSize.Width != 80 {
		t.Errorf("有界 (80) 时宽度 = %d, want 80 (应该自动填充)", filledSize.Width)
	}

	t.Log("✅ flex 容器自动填充测试通过")
}

// TestFiberAdapter_FlexAutoFillVStack 测试 VStack 下 HStack 的自动填充
func TestFiberAdapter_FlexAutoFillVStack(t *testing.T) {
	t.Log("=== 测试 VStack 下 HStack 的自动填充 ===")

	// 创建 HStack fiber（包含 flex 子节点）
	flexChildren := []rtui.Fiber{
		{Tag: "button", LayoutFlex: 1, Props: map[string]interface{}{"content": "Btn1"}},
		{Tag: "button", LayoutFlex: 1, Props: map[string]interface{}{"content": "Btn2"}},
	}

	hstackFiber := &rtui.Fiber{
		Tag:              "hstack",
		LayoutGap:        1,
		LayoutDirection:  rtui.DirectionRow,
		Child:            &flexChildren[0],
	}
	flexChildren[0].Sibling = &flexChildren[1]

	// 创建 VStack fiber
	vstackFiber := &rtui.Fiber{
		Tag:              "vstack",
		LayoutGap:        1,
		LayoutDirection:  rtui.DirectionColumn,
		Child:            hstackFiber,
	}

	// 创建 Adapter
	adapter := render.NewFiberToNodeAdapterPure(vstackFiber)

	// 有界约束 (80, 15)
	boundedConstraints := layout.NewConstraints(0, 80, 0, 15)
	size := adapter.Measure(boundedConstraints)
	t.Logf("VStack 有界约束 (80x15): 尺寸 = %dx%d", size.Width, size.Height)

	// HStack 应该填充 VStack 的宽度 (80)
	if size.Width != 80 {
		t.Errorf("VStack 有界 (80x15) 时宽度 = %d, want 80", size.Width)
	}

	t.Log("✅ VStack 下 HStack 的自动填充测试通过")
}

// TestFiberAdapter_NoFlexChildren_NoAutoFill 测试不含 flex 子节点的容器不自动填充
func TestFiberAdapter_NoFlexChildren_NoAutoFill(t *testing.T) {
	t.Log("=== 测试不含 flex 子节点的容器不自动填充 ===")

	// 创建 HStack fiber（不含 flex 子节点）
	nonFlexChildren := []rtui.Fiber{
		{Tag: "button", LayoutFlex: 0, Props: map[string]interface{}{"content": "Left"}},
		{Tag: "button", LayoutFlex: 0, Props: map[string]interface{}{"content": "Center"}},
		{Tag: "button", LayoutFlex: 0, Props: map[string]interface{}{"content": "Right"}},
	}

	hstackFiber := &rtui.Fiber{
		Tag:              "hstack",
		LayoutGap:        1,
		LayoutDirection:  rtui.DirectionRow,
		Child:            &nonFlexChildren[0],
	}
	nonFlexChildren[0].Sibling = &nonFlexChildren[1]
	nonFlexChildren[1].Sibling = &nonFlexChildren[2]

	// 创建 Adapter
	adapter := render.NewFiberToNodeAdapterPure(hstackFiber)

	// 有界约束 (80)
	boundedConstraints := layout.NewConstraints(0, 80, 0, 15)
	size := adapter.Measure(boundedConstraints)
	t.Logf("有界约束 (80), 无 flex 子节点: 尺寸 = %dx%d", size.Width, size.Height)

	// 无 flex 子节点时，HStack 应该使用自然宽度（不超过 40）
	if size.Width > 40 {
		t.Errorf("无 flex 子节点时不应该自动填充，宽度 = %d (应该紧凑, <=40)", size.Width)
	}

	t.Log("✅ 无 flex 子节点的容器不自动填充测试通过")
}
