package main

import (
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/reducer"
	"github.com/wwsheng009/mint/runtime/store"
	"github.com/wwsheng009/mint/ui"
)

// =============================================================================
// AppState - Modal 状态
// =============================================================================

type AppState struct {
	IsOpen bool // true = modal is open, false = modal is closed
}

// =============================================================================
// Intent Types
// =============================================================================

type OpenModalIntent struct{}
func (OpenModalIntent) IntentType() string { return "OpenModal" }
func (OpenModalIntent) StayPressed() bool  { return true }

type CloseModalIntent struct{}
func (CloseModalIntent) IntentType() string { return "CloseModal" }
func (CloseModalIntent) StayPressed() bool  { return true }

// =============================================================================
// Store 初始化
// =============================================================================

var modalStore = store.NewStore(AppState{
	IsOpen: false,
})

// =============================================================================
// Reducer 注册
// =============================================================================

func init() {
	reducer.NewBuilder[AppState]().
		On(OpenModalIntent{}, func(s AppState, i intent.Intent) AppState {
			s.IsOpen = true
			return s
		}).
		On(CloseModalIntent{}, func(s AppState, i intent.Intent) AppState {
			s.IsOpen = false
			return s
		}).
		BuildAndRegister(intent.DefaultRegistry(), modalStore)
}

// =============================================================================
// Main
// =============================================================================

func main() {
	// ✅ 无需 WithInit，Reducer 已在 init 中注册
	ui.Run(
		MainComponent,
		ui.WithWidth(50),
		ui.WithHeight(20),
		ui.WithTitle("Modal Demo (Store 模式)"),
	)
}

// =============================================================================
// Main Component
// =============================================================================

func MainComponent() ui.VNode {
	// ✅ 订阅 IsOpen 状态
	isOpen := ui.UseStoreSelector(
		modalStore,
		func(s AppState) bool { return s.IsOpen },
	)

	// 如果 modal 打开，显示 modal 内容
	if isOpen {
		return ModalContent()
	}

	// Modal 关闭 - 显示主内容
	return MainContent()
}

// =============================================================================
// Modal Content (打开状态)
// =============================================================================

func ModalContent() ui.VNode {
	return ui.VStack(
		ui.NewTextBuilder("┌───────────────────────────────────────┐").FgColor("cyan").Build(),
		ui.NewTextBuilder("│           MODAL IS OPEN               │").FgColor("cyan").Build(),
		ui.NewTextBuilder("│                                       │").FgColor("cyan").Build(),
		ui.NewTextBuilder("│  Do you want to proceed?              │").FgColor("white").Build(),
		ui.NewTextBuilder("│                                       │").FgColor("cyan").Build(),
		ui.HStack(
			ui.NewTextBuilder("│  ").FgColor("cyan").Build(),
			ui.NewButtonBuilder(" Yes ").
				// ✅ 使用自定义 Intent - 由 Reducer 处理
				OnPress(CloseModalIntent{}).
				Build(),
			ui.NewTextBuilder("  ").FgColor("cyan").Build(),
			ui.NewButtonBuilder(" No ").
				OnPress(CloseModalIntent{}).
				Build(),
			ui.NewTextBuilder("               │").FgColor("cyan").Build(),
		),
		ui.NewTextBuilder("│                                       │").FgColor("cyan").Build(),
		ui.NewTextBuilder("└───────────────────────────────────────┘").FgColor("cyan").Build(),
		ui.Text(""),
		ui.NewTextBuilder("Press Tab to focus, Enter to close").FgColor("gray").Build(),
	)
}

// =============================================================================
// Main Content (关闭状态)
// =============================================================================

func MainContent() ui.VNode {
	return ui.VStack(
		ui.NewTextBuilder("Modal Demo").Bold(true).FgColor("cyan").Build(),
		ui.Text(""),
		ui.NewTextBuilder("Click the button below to open a modal dialog").FgColor("gray").Build(),
		ui.Text(""),
		ui.NewButtonBuilder("  Show Modal  ").
			// ✅ 使用自定义 Intent - 由 Reducer 处理
			OnPress(OpenModalIntent{}).
			Build(),
		ui.Text(""),
		ui.NewTextBuilder("Tab/Arrows: focus | Enter/Space: click").FgColor("gray").Build(),
	)
}
