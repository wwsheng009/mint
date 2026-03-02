// Portal 示例程序 - 展示 Portal 组件的跨树挂载
// Phase 3: Portal 跨树挂载系统演示
//
// Portal 系统允许组件将其子元素渲染到 DOM 树的不同位置
// 主要用于 Modal、Tooltip、Toast 等需要浮层显示的组件
package main

import (
	"github.com/wwsheng009/mint/app"
	"github.com/wwsheng009/mint/runtime/intent"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui"
)

// =============================================================================
// Intent 类型定义
// =============================================================================

type OpenModalIntent struct {
	Content string
}

func (OpenModalIntent) IntentType() string { return "OpenModal" }

type CloseModalIntent struct{}

func (CloseModalIntent) IntentType() string { return "CloseModal" }

// =============================================================================
// Portal 组件
// =============================================================================

// ModalDialogPortal 使用 Portal 渲染到 app 顶层的 Modal
func ModalDialogPortal(content string, show bool) rtui.VNode {
	if !show || content == "" {
		// 返回空文本而不是 nil，避免 reconciler 处理 nil
		return ui.Text("")
	}

	// 使用 Portal 将 Modal 渲染到 "modal-root" PortalRoot
	return rtui.NewElement("box").
		SetProps(rtui.Props{
			"portalRoot": "modal-root", // 🔑 关键：指定 Portal 目标
			"width":     50,
			"height":    12,
			"position":  "fixed",
			"_layer":    rtui.LayerModal,
		}).
		SetChildren([]rtui.VNode{
			rtui.NewElement("box").
				SetProps(rtui.Props{
					"width":            80,
					"height":           25,
					"position":         "fixed",
					"_closeOnBackdrop": true,
					"_onClose":         CloseModalIntent{},
				}).
				SetChildren([]rtui.VNode{
					rtui.NewElement("box").
						SetProps(rtui.Props{
							"width":    40,
							"height":   10,
							"centered": true,
						}).
						SetChildren([]rtui.VNode{
							app.VStack(
								ui.NewTextBuilder("📦 Modal via Portal").
									FgColor("cyan").
									Bold(true).
									Build(),
								ui.Text(""),
								ui.Text(content),
								ui.Text(""),
								ui.NewTextBuilder("按 ESC 或点击外部关闭").
									FgColor("gray").
									Build(),
								ui.Text(""),
								app.ButtonBuilder("  [关闭]  ").
									Variant(app.ButtonVariantPrimary).
									OnPress(CloseModalIntent{}).
									Build(),
							),
						}),
				}),
		})
}

// =============================================================================
// 主应用组件
// =============================================================================

func App() rtui.VNode {
	modalContent, setModalContent := ui.UseStateString("")
	showModal, setShowModal := ui.UseStateBool(false)

	// 注册 Intent 处理器
	rtui.RegisterIntent(func(ctx *intent.ActionContext, i OpenModalIntent) intent.IntentResult {
		setModalContent(i.Content)
		setShowModal(true)
		return intent.HandledResult()
	})

	rtui.RegisterIntent(func(ctx *intent.ActionContext, i CloseModalIntent) intent.IntentResult {
		setShowModal(false)
		return intent.HandledResult()
	})

	// 使用 VStack 包装所有内容（避免 Fragment 的复杂性）
	return app.VStack(
		// ========================================
		// 🔑 PortalRoot - 定义在应用顶层（通过 Stack 隐藏，但存在于树中）
		// 这是所有 Portal 组件的挂载目标
		// ========================================

		// Modal PortalRoot
		rtui.NewElement("box").
			SetProps(rtui.Props{
				"portalRootId": "modal-root", // 🔑 标识为 PortalRoot
				"width":       80,
				"height":      25,
				"_layer":      rtui.LayerModal,
			}),

		// ========================================
		// 主内容区域 - 使用额外的 VStack 分隔
		// ========================================
		app.VStack(
			// 标题
			app.VStack(
				ui.NewTextBuilder("🌟 Portal 跨树挂载演示").
					FgColor("cyan").
					Bold(true).
					Build(),
				ui.Text(""),
				ui.NewTextBuilder("Portal 系统说明").
					FgColor("yellow").
					Bold(true).
					Build(),
				ui.Text(""),
				ui.NewTextBuilder("  • PortalRoot: 定义在应用顶层，作为挂载目标").
					FgColor("gray").
					Build(),
				ui.NewTextBuilder("  • Portal: 子组件通过 props[\"portalRoot\"] 指定目标").
					FgColor("gray").
					Build(),
				ui.NewTextBuilder("  • linkPortalsToRoots(): Commit 阶段自动建立链接").
					FgColor("gray").
					Build(),
				ui.Text(""),
				ui.NewTextBuilder("—").FgColor("gray").Build(),
				ui.Text(""),
			),

			// 交互按钮
			app.VStack(
				ui.HStack(
					ui.Text("  "),
					app.ButtonBuilder("  📦 打开 Modal  ").
						Variant(app.ButtonVariantPrimary).
						OnPress(OpenModalIntent{Content: "这是通过 Portal 渲染到 app 顶层的 Modal！"}).
						Disabled(showModal).
						Build(),
				),

				ui.Text(""),

				// 提示信息
				ui.NewTextBuilder("💡 提示: 按 ESC 或点击 Modal 外部区域可关闭").
					FgColor("gray").
					Italic(true).
					Build(),

				ui.Text(""),

				ui.NewTextBuilder("📋 Portal 渲染流程:").FgColor("cyan").Build(),
				ui.Text(""),
				ui.NewTextBuilder("  1. Render 阶段: VNode → Fiber").
					FgColor("gray").
					Build(),
				ui.NewTextBuilder("  2. Commit 阶段: linkPortalsToRoots() 建立链接").
					FgColor("gray").
					Build(),
				ui.NewTextBuilder("  3. Layout 阶段: 使用 PortalRoot 作为父级布局").
					FgColor("gray").
					Build(),
				ui.NewTextBuilder("  4. Paint 阶段: 在 PortalRoot 位置绘制").
					FgColor("gray").
					Build(),
			),

			ui.Text(""),

			// 状态显示
			rtui.NewElement("box").
				SetProps(rtui.Props{
					"width":  80,
					"height": 3,
				}).
				SetChildren([]rtui.VNode{
					ui.HStack(
						ui.Text("  "),
						ui.NewTextBuilder("Modal 状态: ").
							FgColor("blue").
							Build(),
						ui.NewTextBuilder(func() string {
							if showModal {
								return "🟢 已打开"
							}
							return "🔴 已关闭"
						}()).
							FgColor(func() string {
								if showModal {
									return "green"
								}
								return "gray"
							}()).
							Build(),
					),
				}),

			// ========================================
			// 🔑 Portal 子组件 - 声明在此，但渲染到 PortalRoot
			// ========================================

			ModalDialogPortal(modalContent, showModal),
		),
	)
}

// =============================================================================
// 主程序入口
// =============================================================================

func main() {
	err := ui.Run(App,
		ui.WithWidth(80),
		ui.WithHeight(25),
		ui.WithTitle("Portal Demo - 跨树挂载系统演示"),
	)
	if err != nil {
		panic(err)
	}
}
