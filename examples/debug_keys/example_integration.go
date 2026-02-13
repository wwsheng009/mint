package main

import (
	"fmt"
	"strings"

	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui"
	"github.com/wwsheng009/mint/examples/debug_keys"
)

// 这是一个实际应用示例，展示如何集成 DebugKeyInspector

func main() {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("🔑 Debug Key Inspector - 集成示例")
	fmt.Println(strings.Repeat("=", 80))

	// 启用调试模式
	debugMode := true

	app := ui.NewApp()

	// 状态管理
	var modalOpen bool
	var overlayVisible bool
	var tooltipVisible bool

	// 创建 root 组件
	root := func() rtui.VNode {
		// Base layer content
		baseContent := rtui.NewElement("vstack").
			SetChildren(
				rtui.NewElement("text").
					SetProp("content", "这是一个演示应用，包含多层 layer"),
				rtui.NewElement("button").
					SetProp("label", "打开 Modal").
					OnClick(func() {
						modalOpen = true
						app.MarkDirty()
					}),
				rtui.NewElement("button").
					SetProp("label", "显示 Overlay").
					OnClick(func() {
						overlayVisible = true
						app.MarkDirty()
					}),
				rtui.NewElement("button").
					SetProp("label", "显示 Tooltip").
					OnClick(func() {
						tooltipVisible = true
						app.MarkDirty()
					}),
				rtui.NewElement("button").
					SetProp("label", "关闭所有").
					OnClick(func() {
						modalOpen = false
						overlayVisible = false
						tooltipVisible = false
						app.MarkDirty()
					}),
			)

		// Modal layer
		var modalNode rtui.VNode
		if modalOpen {
			modalContent := rtui.NewElement("vstack").
				SetChildren(
					rtui.NewElement("text").SetProp("content", "这是 Modal"),
					rtui.NewElement("text").SetProp("content", "你应该能看到这个节点有 [MODAL] 标记"),
					rtui.NewElement("button").
						SetProp("label", "Modal 中的按钮").
						SetKey("modal-btn").
						OnClick(func() {
							fmt.Println("✅ Modal 按钮点击成功！")
							if debugMode {
								// 每次点击后检查 KEY 信息
								inspectKeys(app)
							}
						}),
				)
			modalNode = ui.Modal(modalContent)
		}

		// Overlay layer
		var overlayNode rtui.VNode
		if overlayVisible {
			overlayContent := rtui.NewElement("text").SetProp("content", "这是 Overlay")
			overlayNode = ui.Overlay(overlayContent)
		}

		// Tooltip layer
		var tooltipNode rtui.VNode
		if tooltipVisible {
			tooltipNode = ui.Tooltip("这是 Tooltip 内容")
		}

		// 组合所有层
		children := []rtui.VNode{baseContent}
		if modalNode != nil {
			children = append(children, modalNode)
		}
		if overlayNode != nil {
			children = append(children, overlayNode)
		}
		if tooltipNode != nil {
			children = append(children, tooltipNode)
		}

		return rtui.NewElement("fragment").SetChildren(children...)
	}

	// 首次渲染
	app.Render(root)

	// 📊 初始 KEY 检查
	if debugMode {
		inspectKeys(app)
		fmt.Println("\n💡 提示: 点击上方按钮打开不同的 layer，观察 KEY 的变化")
		fmt.Println("   点击 Modal 中的按钮会触发 KEY 检查")
		fmt.Println()
	}
}

// inspectKeys 在渲染后检查 KEY 信息
func inspectKeys(app *ui.App) {
	fmt.Println("\n" + strings.Repeat("─", 80))
	fmt.Println("📊 当前渲染状态检查")
	fmt.Println(strings.Repeat("─", 80))

	// 创建 inspector
	inspector := debug_keys.NewDebugKeyInspector()
	inspector.MaxDepth = 15
	inspector.ShowKeys = true
	inspector.ShowPaths = true
	inspector.ShowLayers = true

	// 获取 VNode 树（简化版，实际应用中需要完整的渲染流程）
	vnodes := app.Root()

	// 显示 VNode 树
	inspector.InspectVNodes(vnodes)

	// 注意：实际应用中还需要获取 Fiber 树
	// fibers := app.GetCurrentFiber()  // 需要实现这个方法
	// inspector.InspectFibers(fibers)
	// inspector.CompareTrees(vnodes, fibers)

	// 显示统计信息
	inspector.PrintStatistics(vnodes, nil)  // fiber 为 nil 时只显示 VNode 统计
}
