package main

import (
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/reducer"
	"github.com/wwsheng009/mint/runtime/store"
	"github.com/wwsheng009/mint/ui"
)

// =============================================================================
// AppState - Toast 状态
// =============================================================================

type AppState struct {
	Info     bool // 是否显示 info toast
	Success  bool // 是否显示 success toast
	Warning  bool // 是否显示 warning toast
	Error    bool // 是否显示 error toast
}

// =============================================================================
// Intent Types
// =============================================================================

type ShowInfoToastIntent struct{}
func (ShowInfoToastIntent) IntentType() string     { return "ShowInfoToast" }
func (ShowInfoToastIntent) StayPressed() bool      { return true }

type ShowSuccessToastIntent struct{}
func (ShowSuccessToastIntent) IntentType() string  { return "ShowSuccessToast" }
func (ShowSuccessToastIntent) StayPressed() bool   { return true }

type ShowWarningToastIntent struct{}
func (ShowWarningToastIntent) IntentType() string  { return "ShowWarningToast" }
func (ShowWarningToastIntent) StayPressed() bool   { return true }

type ShowErrorToastIntent struct{}
func (ShowErrorToastIntent) IntentType() string    { return "ShowErrorToast" }
func (ShowErrorToastIntent) StayPressed() bool     { return true }

type ClearAllToastsIntent struct{}
func (ClearAllToastsIntent) IntentType() string    { return "ClearAllToasts" }
func (ClearAllToastsIntent) StayPressed() bool     { return true }

// =============================================================================
// Store 初始化
// =============================================================================

var toastStore = store.NewStore(AppState{
	Info:     false,
	Success:  false,
	Warning:  false,
	Error:    false,
})

// =============================================================================
// Reducer 注册
// =============================================================================

func init() {
	reducer.NewBuilder[AppState]().
		On(ShowInfoToastIntent{}, func(s AppState, i intent.Intent) AppState {
			s.Info = true
			return s
		}).
		On(ShowSuccessToastIntent{}, func(s AppState, i intent.Intent) AppState {
			s.Success = true
			return s
		}).
		On(ShowWarningToastIntent{}, func(s AppState, i intent.Intent) AppState {
			s.Warning = true
			return s
		}).
		On(ShowErrorToastIntent{}, func(s AppState, i intent.Intent) AppState {
			s.Error = true
			return s
		}).
		On(ClearAllToastsIntent{}, func(s AppState, i intent.Intent) AppState {
			s.Info = false
			s.Success = false
			s.Warning = false
			s.Error = false
			return s
		}).
		BuildAndRegister(intent.DefaultRegistry(), toastStore)
}

// =============================================================================
// Main
// =============================================================================

func main() {
	// ✅ 无需 ui.On，Reducer 已在 init 中注册
	ui.Run(MainComponent,
		ui.WithWidth(50),
		ui.WithHeight(20),
		ui.WithTitle("Toast Demo (Store 模式)"),
	)
}

// =============================================================================
// Main Component
// =============================================================================

func MainComponent() ui.VNode {
	// ✅ 订阅所有 toast 状态
	info := ui.UseStoreSelector(toastStore, func(s AppState) bool { return s.Info })
	success := ui.UseStoreSelector(toastStore, func(s AppState) bool { return s.Success })
	warning := ui.UseStoreSelector(toastStore, func(s AppState) bool { return s.Warning })
	error := ui.UseStoreSelector(toastStore, func(s AppState) bool { return s.Error })

	// 构建 toast 列表
	var toasts []ui.VNode
	if info {
		toasts = append(toasts, ui.NewToastBuilder("This is an info message").Info().Build())
	}
	if success {
		toasts = append(toasts, ui.NewToastBuilder("Operation completed successfully!").Success().Build())
	}
	if warning {
		toasts = append(toasts, ui.NewToastBuilder("Please check your input").Warning().Build())
	}
	if error {
		toasts = append(toasts, ui.NewToastBuilder("An error occurred!").Error().Build())
	}

	return ui.VStack(
		ui.NewTextBuilder("Toast Notifications Demo").Bold(true).FgColor("cyan").Build(),
		ui.Text(""),
		ui.Text("Click buttons below to show different toast types:"),
		ui.Text(""),
		ui.HStack(
			ui.NewButtonBuilder(" Info ").OnPress(ShowInfoToastIntent{}).Build(),
			ui.NewButtonBuilder(" Success ").OnPress(ShowSuccessToastIntent{}).Build(),
			ui.NewButtonBuilder(" Warning ").OnPress(ShowWarningToastIntent{}).Build(),
			ui.NewButtonBuilder(" Error ").OnPress(ShowErrorToastIntent{}).Build(),
		),
		ui.Text(""),
		ui.Text(""),
		ui.HStack(
			ui.NewButtonBuilder(" Clear All ").OnPress(ClearAllToastsIntent{}).Build(),
		),
		ui.Text(""),
		ui.NewTextBuilder("────────────────────────────").FgColor("blue").Build(),
		ui.Text(""),
		ui.VStack(toasts...),
	)
}
