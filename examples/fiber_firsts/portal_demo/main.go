// Portal 示例程序 - 展示 Modal 和 Tooltips 的跨树挂载
// Phase 3: Portal 跨树挂载系统演示
package main

import (
	"fmt"
	"time"

	"github.com/wwsheng009/mint/framework/app"
	"github.com/wwsheng009/mint/framework/component"
	ui "github.com/wwsheng009/mint/runtime/ui"
)

// ========================================
// 应用状态
// ========================================

type AppState struct {
	showModal     bool
	showTooltip   string // 空表示隐藏，非空表示显示的文本
	modalContent  string
}

// ========================================
// Portal 组件
// ========================================

// ModalDialog 模态对话框组件 - 使用 Portal 渲染到应用顶层
type ModalDialog struct {
	component.ComponentBase
	Content  string
	OnClose  func()
	Show     bool
}

// Render 实现 Component 接口
func (m *ModalDialog) Render(ctx *component.Context) ui.VNode {
	if !m.Show {
		return nil
	}

	// Modal 使用 Portal 渲染到 "modal-root"
	return ui.Box(
		ui.Props{
			"portalRoot": "modal-root",  // 🔑 指定 Portal 目标
			"width":     50,
			"height":    15,
			"style":     "modal",
			"position":  "fixed",        // 固定定位
			"_layer":    ui.LayerModal,  // Modal 层级
		},
		// Modal 背景遮罩
		ui.Box(
			ui.Props{
				"width":              80,
				"height":             25,
				"style":              "modalBackdrop",
				"position":           "fixed",
				"_closeOnBackdrop":   true,
				"_onClose":           m.OnClose,
			},
			// Modal 内容
			ui.Box(
				ui.Props{
					"width":     40,
					"height":    10,
					"centered":  true,
				},
				ui.Text(m.Content),
				// 确认按钮
				ui.Box(
					ui.Props{
						"width":   10,
						"height":  3,
						"style":   "button",
					},
					ui.Text("确认"),
				),
			),
		),
	)
}

// Tooltip 提示框组件 - 使用 Portal 渲染到应用顶层
type Tooltip struct {
	component.ComponentBase
	Content string
	Target  string // 锚定目标的 ID
}

// Render 实现 Component 接口
func (t *Tooltip) Render(ctx *component.Context) ui.VNode {
	if t.Content == "" {
		return nil
	}

	// Tooltip 使用 Portal 渲染到 "tooltip-root"
	return ui.Box(
		ui.Props{
			"portalRoot": "tooltip-root",  // 🔑 指定 Portal 目标
			"width":     len(t.Content) + 4,
			"height":    3,
			"style":     "tooltip",
			"position":  "fixed",
			"_layer":    ui.LayerTooltip,  // Tooltip 层级
		},
		ui.Text(t.Content),
	)
}

// ========================================
// 普通组件
// ========================================

// MainContent 主内容区域组件
type MainContent struct {
	component.ComponentBase
	ShowModal    func(string)
	ShowTooltip  func(string)
}

// Render 实现 Component 接口
func (mc *MainContent) Render(ctx *component.Context) ui.VNode {
	return ui.Box(
		ui.Props{
			"width":  80,
			"height": 20,
			"style":  "mainContent",
		},
		ui.Text("🌟 Portal 系统演示"),
		ui.Box(ui.Props{"height": 1}),
		ui.Text("这是一个使用 Portal 跨树挂载的示例应用"),
		ui.Box(ui.Props{"height": 1}),
		// 打开 Modal 按钮
		ui.Button(
			ui.Props{
				"width":   20,
				"height":  3,
				"onEnter": func(...interface{}) { mc.ShowModal("这是通过 Portal 渲染的 Modal 对话框！") },
			},
			ui.Text("📦 打开 Modal"),
		),
		ui.Box(ui.Props{"height": 1}),
		// 显示 Tooltip 按钮
		ui.Button(
			ui.Props{
				"width":   20,
				"height":  3,
				"onEnter": func(...interface{}) { mc.ShowTooltip("这是通过 Portal 渲染的 Tooltip！") },
				"onLeave": func(...interface{}) { mc.ShowTooltip("") },
			},
			ui.Text("💬 显示 Tooltip"),
		),
	)
}

// ========================================
// 根应用组件
// ========================================

type PortalDemoApp struct {
	component.ComponentBase
}

// Render 实现 Component 接口
func (app *PortalDemoApp) Render(ctx *component.Context) ui.VNode {
	state := ctx.GetState().(*AppState)

	// 使用 Fragment 返回多个子节点（包括 PortalRoot）
	return ui.Fragment(
		// ========================================
		// 🔑 PortalRoot - 放在应用顶层
		// ========================================

		// Modal PortalRoot
		ui.Box(
			ui.Props{
				"portalRootId": "modal-root",   // 🔑 标识为 PortalRoot
				"width":       80,
				"height":      25,
				"_layer":      ui.LayerModal,
			},
			// Modal 子组件将通过 Portal 渲染到这里
		),

		// Tooltip PortalRoot
		ui.Box(
			ui.Props{
				"portalRootId": "tooltip-root",  // 🔑 标识为 PortalRoot
				"width":       80,
				"height":      25,
				"_layer":      ui.LayerTooltip,
			},
			// Tooltip 子组件将通过 Portal 渲染到这里
		),

		// ========================================
		// 主内容区域
		// ========================================

		ui.Box(
			ui.Props{
				"width":  80,
				"height": 25,
				"style":  "app",
			},
			component.New(&MainContent{
				ShowModal: func(content string) {
					state.showModal = true
					state.modalContent = content
					ctx.RequestRender()
				},
				ShowTooltip: func(content string) {
					state.showTooltip = content
					ctx.RequestRender()
				},
			}),

			// ========================================
			// Portal 子组件 - 声明在主内容中，但会渲染到 PortalRoot
			// ========================================

			// Modal Portal（通过 props["portalRoot"] 指定目标）
			component.New(&ModalDialog{
				Content: state.modalContent,
				Show:    state.showModal,
				OnClose: func() {
					state.showModal = false
					ctx.RequestRender()
				},
			}),

			// Tooltip Portal
			component.New(&Tooltip{
				Content: state.showTooltip,
			}),
		),
	)
}

// ========================================
// 主程序
// ========================================

func main() {
	fmt.Println("🚀 Mint TUI - Portal 系统演示")
	fmt.Println("=" + "=======================================")
	fmt.Println()
	fmt.Println("📋 说明：")
	fmt.Println("  1. 应用顶层定义了两个 PortalRoot (modal-root, tooltip-root)")
	fmt.Println("  2. Modal 和 Tooltip 子组件使用 props[\"portalRoot\"] 指定目标")
	fmt.Println("  3. Reconciler.linkPortalsToRoots() 在 Commit 阶段自动建立链接")
	fmt.Println("  4. Modal/Tooltip 虽然声明在主内容中，但实际渲染到 PortalRoot 位置")
	fmt.Println()
	fmt.Println("⌨️ 交互：")
	fmt.Println("  - Tab/Shift+Tab: 在按钮间导航")
	fmt.Println("  - Enter: 打开 Modal 或显示 Tooltip")
	fmt.Println("  - ESC: 关闭 Modal")
	fmt.Println()
	fmt.Println("按任意键启动程序...")
	time.Sleep(2 * time.Second)

	// 创建应用
	appInstance := app.New(
		80, 25,
		"Portal Demo",
		component.New(&PortalDemoApp{}),
	)

	// 初始化应用状态
	appInstance.SetState(&AppState{
		showModal:    false,
		showTooltip:  "",
		modalContent: "",
	})

	// 运行应用
	appInstance.Run()
}
