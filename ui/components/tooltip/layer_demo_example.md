package tooltip

import (
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// 这是指南，展示如何使用 tooltip 组件的多层 layer 功能

// =============================================================================
// Tooltip Layer 使用示例
// =============================================================================

// Example 1: 使用 SetLayer() 设置层级
func ExampleVNode_SetLayer() {
	content := rtui.NewText("Hover me")

	// 设置为 Tooltip 层（默认）
	t := New(content, "Help info")
	t.SetLayer(rtui.LayerTooltip)

	// 设置为 Overlay 层
	t.SetLayer(rtui.LayerOverlay)

	// 设置为 Modal 层（紧急提示）
	t.SetLayer(rtui.LayerModal)

	// 设置为 Inspector 层（调试专用）
	t.SetLayer(rtui.LayerInspector)
}

// Example 2: 使用 Builder API 设置层级
func ExampleBuilder_Layer() {
	content := rtui.NewText("Hover me")

	// 使用 Builder.Layer()
	tooltip.NewBuilder(content, "Help info").
		Layer(rtui.LayerTooltip).
		Build()

	// 使用便捷方法
	tooltip.NewBuilder(content, "Important").
		ModalLayer().      // Modal 层
		Build()

	tooltip.NewBuilder(content, "Normal").
		OverlayLayer().    // Overlay 层
		Build()

	tooltip.NewBuilder(content, "Debug").
		InspectorLayer().  // Inspector 层
		Build()

	tooltip.NewBuilder(content, "Base").
		BaseLayer().       // Base 层
		Build()
}

// =============================================================================
// Toast Layer 使用示例
// =============================================================================

// Example 3: Toast SetLayer()
func ExampleToastVNode_SetLayer() {
	t := NewToast("Operation successful")

	// 设置为 Overlay 层（默认）
	t.SetLayer(rtui.LayerOverlay)

	// 设置为 Tooltip 层
	t.SetLayer(rtui.LayerTooltip)

	// 设置为 Modal 层（紧急通知）
	t.SetLayer(rtui.LayerModal)
}

// Example 4: Toast Builder API
func ExampleToastBuilder_Layer() {
	// 使用 Builder.Layer()
	tooltip.NewToastBuilder("Saved!").
		Layer(rtui.LayerOverlay).
		Build()

	// 使用便捷方法
	tooltip.NewToastBuilder("Error!").
		Error().
		ModalLayer().      // Modal 层
		Build()

	tooltip.NewToastBuilder("Info").
		Info().
		OverlayLayer().    // Overlay 层
		Build()

	tooltip.NewToastBuilder("Debug").
		InspectorLayer().  // Inspector 层
		Build()
}

// =============================================================================
// 层级选择指南
// =============================================================================

// Layer Guide:
// - LayerBase:     (0)  基础层，普通 UI 内容
// - LayerOverlay:  (1)  覆盖层，下拉菜单、弹出框、普通 Toast
// - LayerModal:    (2)  模态层，需要用户关注的模态对话框、紧急通知
// - LayerTooltip:  (3)  提示层，Tooltip、Help 文本
// - LayerInspector:(4)  检查器层，UI 检查器调试覆盖层

// 使用建议：
// - Tooltip: 使用 LayerTooltip (默认) 或 LayerOverlay
// - Toast:   使用 LayerOverlay (默认)
// - 紧急提示: 使用 LayerModal
// - 调试信息: 使用 LayerInspector

// =============================================================================
// 完整示例：创建不同层级的提示组件
// =============================================================================

func Example_CompleteUsage() {
	button := rtui.NewText("Click me")

	// 1. 普通 Tooltip - Tooltip 层（默认）
	normalT := tooltip.NewBuilder(button, "Click to submit").
		Position(PositionTop).
		Build()

	// 2. 重要提示 - Modal 层
	importantT := tooltip.New(button, "This action cannot be undone")
	importantT.SetLayer(rtui.LayerModal)

	// 3. 调试信息 - Inspector 层
	debugT := tooltip.New(button, "DEBUG: onClick handler registered")
	debugT.SetLayer(rtui.LayerInspector)

	// 4. 普通 Toast - Overlay 层
	infoToast := tooltip.NewToastBuilder("Saved successfully").
		Info().
		Build()

	// 5. 错误 Toast - Overlay 层（也可用 Modal）
	errorToast := tooltip.NewToastBuilder("Failed to save").
		Error().
		Build()

	// 6. 紧急 Toast - Modal 层
	urgentToast := tooltip.NewToastBuilder("Server error! Contact support.")
	urgentToast.SetLayer(rtui.LayerModal)

	_, _, _, _, _, _ = normalT, importantT, debugT, infoToast, errorToast, urgentToast
}
