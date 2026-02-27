package main

import (
	"fmt"
	"strings"

	"github.com/wwsheng009/mint/app"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/ui"
)

// =============================================================================
// Intent Types
// =============================================================================

type OpenModalIntent struct{}
func (OpenModalIntent) IntentType() string { return "OpenModal" }
func (OpenModalIntent) StayPressed() bool  { return true }

type ShowOverlayIntent struct{}
func (ShowOverlayIntent) IntentType() string { return "ShowOverlay" }
func (ShowOverlayIntent) StayPressed() bool  { return true }

type ToggleInspectorIntent struct{}
func (ToggleInspectorIntent) IntentType() string { return "ToggleInspector" }
func (ToggleInspectorIntent) StayPressed() bool  { return true }

type CloseAllLayersIntent struct{}
func (CloseAllLayersIntent) IntentType() string { return "CloseAllLayers" }
func (CloseAllLayersIntent) StayPressed() bool  { return true }

type ToggleShowKeysIntent struct{}
func (ToggleShowKeysIntent) IntentType() string { return "ToggleShowKeys" }
func (ToggleShowKeysIntent) StayPressed() bool  { return true }

type ToggleShowPathsIntent struct{}
func (ToggleShowPathsIntent) IntentType() string { return "ToggleShowPaths" }
func (ToggleShowPathsIntent) StayPressed() bool  { return true }

type ToggleShowLayersIntent struct{}
func (ToggleShowLayersIntent) IntentType() string { return "ToggleShowLayers" }
func (ToggleShowLayersIntent) StayPressed() bool  { return true }

type ModalButtonClickIntent struct{}
func (ModalButtonClickIntent) IntentType() string { return "ModalButtonClick" }
func (ModalButtonClickIntent) StayPressed() bool  { return true }

type CloseModalIntent struct{}
func (CloseModalIntent) IntentType() string { return "CloseModal" }
func (CloseModalIntent) StayPressed() bool  { return true }

type OverlayButtonClickIntent struct{}
func (OverlayButtonClickIntent) IntentType() string { return "OverlayButtonClick" }
func (OverlayButtonClickIntent) StayPressed() bool  { return true }

// =============================================================================
// KeyInspectorUI - UI Inspector
// =============================================================================

func main() {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("🔑 UI Key Inspector - 交互式调试工具")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("\n这个工具会在 UI 中以 inspector 的形式显示所有 nodes 的 KEY 信息")
	fmt.Println("按 'q' 退出应用")
	fmt.Println("")

	ui.Run(func() ui.VNode {
		// 使用 UseState 管理状态
		modalOpen, setModalOpen := ui.UseStateBool(false)
		overlayVisible, setOverlayVisible := ui.UseStateBool(false)
		inspectorEnabled, setInspectorEnabled := ui.UseStateBool(true)
		showKeys, setShowKeys := ui.UseStateBool(true)
		showPaths, setShowPaths := ui.UseStateBool(true)
		showLayers, setShowLayers := ui.UseStateBool(true)

		// 使用 ui.On 注册 Intent handler（闭包捕获最新状态值）
		ui.On(OpenModalIntent{}, func() {
			setModalOpen(true)
		})
		ui.On(ShowOverlayIntent{}, func() {
			setOverlayVisible(true)
		})
		ui.On(ToggleInspectorIntent{}, func() {
			setInspectorEnabled(!inspectorEnabled)
		})
		ui.On(CloseAllLayersIntent{}, func() {
			setModalOpen(false)
			setOverlayVisible(false)
		})
		ui.On(ToggleShowKeysIntent{}, func() {
			setShowKeys(!showKeys)
		})
		ui.On(ToggleShowPathsIntent{}, func() {
			setShowPaths(!showPaths)
		})
		ui.On(ToggleShowLayersIntent{}, func() {
			setShowLayers(!showLayers)
		})
		ui.On(ModalButtonClickIntent{}, func() {
			fmt.Println("Modal button clicked!")
		})
		ui.On(CloseModalIntent{}, func() {
			setModalOpen(false)
		})
		ui.On(OverlayButtonClickIntent{}, func() {
			fmt.Println("Overlay button clicked!")
		})

		// Base layer content
		baseContent := app.VStack(
			app.NewTextBuilder("🔑 UI Key Inspector 演示").Bold(true).FgColor("cyan").Build(),
			app.Text(""),
			app.NewTextBuilder("点击按钮打开不同的 layer，观察 Inspector 中的 KEY 变化").FgColor("gray").Build(),
			app.Text(""),
			app.ButtonBuilder("打开 Modal").
				OnPress(OpenModalIntent{}).
				Build(),
			app.ButtonBuilder("显示 Overlay").
				OnPress(ShowOverlayIntent{}).
				Build(),
			app.ButtonBuilder("切换 Inspector").
				OnPress(ToggleInspectorIntent{}).
				Build(),
			app.ButtonBuilder("关闭所有 Layers").
				OnPress(CloseAllLayersIntent{}).
				Build(),
			app.Text(""),
			app.NewTextBuilder("─────────────────────────────────").FgColor("gray").Build(),
			app.NewTextBuilder("显示选项:").FgColor("cyan").Build(),
			app.Text(""),
			buildCheckbox("显示 Keys", showKeys, ToggleShowKeysIntent{}),
			buildCheckbox("显示 Paths", showPaths, ToggleShowPathsIntent{}),
			buildCheckbox("显示 Layer 标记", showLayers, ToggleShowLayersIntent{}),
		)

		var modalContent ui.VNode
		if modalOpen {
			modalContent = app.VStack(
				app.NewTextBuilder("这是 Modal (LayerModal)").FgColor("cyan").Build(),
				app.NewTextBuilder("观察 Inspector 中这个节点的 KEY").FgColor("gray").Build(),
				app.Text(""),
				app.ButtonBuilder("Modal 内部的按钮").
					OnPress(ModalButtonClickIntent{}).
					Build(),
				app.ButtonBuilder("关闭 Modal").
					OnPress(CloseModalIntent{}).
					Build(),
			)
		}

		var overlayContent ui.VNode
		if overlayVisible {
			overlayContent = app.VStack(
				app.NewTextBuilder("这是 Overlay (LayerOverlay)").FgColor("yellow").Build(),
				app.NewTextBuilder("观察 Inspector 中这个节点的 KEY").FgColor("gray").Build(),
				app.Text(""),
				app.ButtonBuilder("Overlay 按钮").
					OnPress(OverlayButtonClickIntent{}).
					Build(),
			)
		}

		// 创建 Inspector overlay
		var inspectorContent ui.VNode
		if inspectorEnabled {
			inspectorContent = createInspectorOverlay(showKeys, showPaths, showLayers)
		}

		// 组合所有 children
		children := []ui.VNode{baseContent}
		if modalContent != nil {
			children = append(children, modalContent)
		}
		if overlayContent != nil {
			children = append(children, overlayContent)
		}
		if inspectorContent != nil {
			children = append(children, inspectorContent)
		}

		return app.VStack(children...)
	},
		ui.WithWidth(80),
		ui.WithHeight(40),
		ui.WithTitle("UI Key Inspector"),
	)
}

// buildCheckbox 创建一个简单的 checkbox (使用 text + button 模拟)
func buildCheckbox(label string, checked bool, toggleIntent intent.Intent) ui.VNode {
	var status string
	if checked {
		status = "[X] "
	} else {
		status = "[ ] "
	}

	return app.HStack(
		app.NewTextBuilder(status+label).Build(),
		app.ButtonBuilder("切换").
			OnPress(toggleIntent).
			Build(),
	)
}

// createInspectorOverlay 创建 Inspector overlay，显示 KEY 信息
func createInspectorOverlay(showKeys, showPaths, showLayers bool) ui.VNode {
	// TODO: 这里应该获取当前的 VNode 和 Fiber 树
	// 由于当前架构限制，我们创建一个示例演示

	lines := []string{
		"══════════════════════════════════════════",
		"🔑 KEY INSPECTOR",
		"══════════════════════════════════════════",
		"",
		"示例 KEY 信息:",
		"│─ vstack",
		"  │─ text",
		"  │─ button",
		"  │─ vstack [MODAL]",
		"    │─ text",
		"    │─ button",
		"",
		fmt.Sprintf("显示 Keys: %v", showKeys),
		fmt.Sprintf("显示 Paths: %v", showPaths),
		fmt.Sprintf("显示 Layers: %v", showLayers),
		"",
		"注意: 这是一个演示 overlay",
		"要获取真实数据，需要访问",
		"reconciler 的内部状态",
		"══════════════════════════════════════════",
	}

	textNodes := make([]ui.VNode, len(lines))
	for i, line := range lines {
		textNodes[i] = app.Text(line)
	}

	return app.VStack(textNodes...)
}
