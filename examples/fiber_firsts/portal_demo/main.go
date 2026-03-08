// Portal 示例程序 - 展示 Portal 组件的跨树挂载
// Phase 3: Portal 跨树挂载系统演示
//
// Portal 系统允许组件将其子元素渲染到 DOM 树的不同位置
// 主要用于 Modal、Tooltip、Toast 等需要浮层显示的组件
//
// Architecture: Store + Reducer + Custom Intent (Single Source of Truth)

package main

import (
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/reducer"
	"github.com/wwsheng009/mint/runtime/store"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui"
)

// =============================================================================
// AppState (Single Source of Truth)
// =============================================================================

// AppState represents the portal demo state.
type AppState struct {
	ModalContent string
	ShowModal    bool
}

// =============================================================================
// Custom Intent Types
// =============================================================================

// OpenModalIntent opens a modal with the specified content.
type OpenModalIntent struct {
	Content string
}

func (OpenModalIntent) IntentType() string { return "OpenModal" }
func (OpenModalIntent) StayPressed() bool  { return true }

// CloseModalIntent closes the modal.
type CloseModalIntent struct{}

func (CloseModalIntent) IntentType() string { return "CloseModal" }
func (CloseModalIntent) StayPressed() bool  { return false }

// =============================================================================
// Reducer (Pure Function)
// =============================================================================

// appReducer handles all state transitions.
var appReducer = reducer.NewBuilder[AppState]()

// Initialize the reducer.
func init() {
	// Handle OpenModalIntent
	appReducer.On(OpenModalIntent{}, func(s AppState, i intent.Intent) AppState {
		omi := i.(OpenModalIntent)
		s.ModalContent = omi.Content
		s.ShowModal = true
		return s
	})

	// Handle CloseModalIntent
	appReducer.On(CloseModalIntent{}, func(s AppState, i intent.Intent) AppState {
		s.ShowModal = false
		return s
	})
}

// =============================================================================
// Store (Single State Source)
// =============================================================================

// appStore holds the portal demo state.
var appStore = store.NewStore(AppState{
	ModalContent: "",
	ShowModal:    false,
})

// =============================================================================
// Portal Component
// =============================================================================

// ModalDialogPortal uses Portal to render a modal to the app's top layer.
func ModalDialogPortal(content string, show bool) rtui.VNode {
	if !show || content == "" {
		// Return empty text instead of nil to avoid reconciler issues
		return ui.Text("")
	}

	// Use Portal to render Modal to "modal-root" PortalRoot
	return rtui.NewElement("box").
		SetProps(rtui.Props{
			"portalRoot": "modal-root", // 🔑 Key: specify Portal target
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
							ui.VStack(
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
								ui.NewButtonBuilder("  [关闭]  ").
									Variant(ui.ButtonVariantPrimary).
									OnPress(CloseModalIntent{}).
									Build(),
							),
						}),
				}),
		})
}

// =============================================================================
// Main Application Component
// =============================================================================

func App() rtui.VNode {
	// Get current state snapshot from Store
	state := appStore.Get()

	// Wrap all content with VStack (to avoid Fragment complexity)
	return ui.VStack(
		// ========================================
		// 🔑 PortalRoot - defined at app top layer
		// This is the mount target for all Portal components
		// ========================================

		// Modal PortalRoot
		rtui.NewElement("box").
			SetProps(rtui.Props{
				"portalRootId": "modal-root", // 🔑 Identified as PortalRoot
				"width":       80,
				"height":      25,
				"_layer":      rtui.LayerModal,
			}),

		// ========================================
		// Main content area - separated with extra VStack
		// ========================================
		ui.VStack(
			// Title
			ui.VStack(
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

			// Interactive buttons
			ui.VStack(
				ui.HStack(
					ui.Text("  "),
					ui.NewButtonBuilder("  📦 打开 Modal  ").
						Variant(ui.ButtonVariantPrimary).
						OnPress(OpenModalIntent{Content: "这是通过 Portal 渲染到 app 顶层的 Modal！"}).
						Disabled(state.ShowModal).
						Build(),
				),

				ui.Text(""),

				// Hint
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

			// Status display
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
							if state.ShowModal {
								return "🟢 已打开"
							}
							return "🔴 已关闭"
						}()).
							FgColor(func() string {
								if state.ShowModal {
									return "green"
								}
								return "gray"
							}()).
							Build(),
					),
				}),

			// ========================================
			// 🔑 Portal child component - declared here but rendered to PortalRoot
			// ========================================

			ModalDialogPortal(state.ModalContent, state.ShowModal),
		),
	)
}

// =============================================================================
// Main Entry Point
// =============================================================================

func main() {
	// Register reducer handlers to store
	appReducer.RegisterToGlobal(appStore)

	err := ui.Run(App,
		ui.WithWidth(80),
		ui.WithHeight(25),
		ui.WithTitle("Portal Demo - 跨树挂载系统演示"),
	)
	if err != nil {
		panic(err)
	}
}
