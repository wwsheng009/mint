package main

import (
	"fmt"
	"strings"

	"github.com/wwsheng009/mint/app"
	"github.com/wwsheng009/mint/ui"
)

// KeyInspectorUI - UI Inspector，在界面中显示所有层次的 KEY 信息
func main() {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("🔑 UI Key Inspector - 交互式调试工具")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("\n这个工具会在 UI 中以 inspector 的形式显示所有 nodes 的 KEY 信息")
	fmt.Println("按 'q' 退出应用")
	fmt.Println("")

	// 状态管理
	var modalOpen bool
	var overlayVisible bool
	var inspectorEnabled bool = true
	var showKeys bool = true
	var showPaths bool = true
	var showLayers bool = true

	ui.Run(func() ui.VNode {
		// Base layer content
		baseContent := app.VStack(
			app.NewTextBuilder("🔑 UI Key Inspector 演示").Bold(true).FgColor("cyan").Build(),
			app.Text(""),
			app.NewTextBuilder("点击按钮打开不同的 layer，观察 Inspector 中的 KEY 变化").FgColor("gray").Build(),
			app.Text(""),
			app.ButtonBuilder("打开 Modal").
				OnClick(func() {
					modalOpen = true
				}).
				Build(),
			app.ButtonBuilder("显示 Overlay").
				OnClick(func() {
					overlayVisible = true
				}).
				Build(),
			app.ButtonBuilder("切换 Inspector").
				OnClick(func() {
					inspectorEnabled = !inspectorEnabled
				}).
				Build(),
			app.ButtonBuilder("关闭所有 Layers").
				OnClick(func() {
					modalOpen = false
					overlayVisible = false
				}).
				Build(),
			app.Text(""),
			app.NewTextBuilder("─────────────────────────────────").FgColor("gray").Build(),
			app.NewTextBuilder("显示选项:").FgColor("cyan").Build(),
			app.Text(""),
			buildCheckbox("显示 Keys", showKeys, func() {
				showKeys = !showKeys
			}),
			buildCheckbox("显示 Paths", showPaths, func() {
				showPaths = !showPaths
			}),
			buildCheckbox("显示 Layer 标记", showLayers, func() {
				showLayers = !showLayers
			}),
		)

		var modalContent ui.VNode
		if modalOpen {
			modalContent = app.VStack(
				app.NewTextBuilder("这是 Modal (LayerModal)").FgColor("cyan").Build(),
				app.NewTextBuilder("观察 Inspector 中这个节点的 KEY").FgColor("gray").Build(),
				app.Text(""),
				app.ButtonBuilder("Modal 内部的按钮").
					OnClick(func() {
						fmt.Println("Modal button clicked!")
					}).
					Build(),
				app.ButtonBuilder("关闭 Modal").
					OnClick(func() {
						modalOpen = false
					}).
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
					OnClick(func() {
						fmt.Println("Overlay button clicked!")
					}).
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
func buildCheckbox(label string, checked bool, onClick func()) ui.VNode {
	var status string
	if checked {
		status = "[X] "
	} else {
		status = "[ ] "
	}

	return app.HStack(
		app.NewTextBuilder(status+label).Build(),
		app.ButtonBuilder("切换").
			OnClick(onClick).
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
