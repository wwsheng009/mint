package main

import (
	"github.com/wwsheng009/mint/app"
	"github.com/wwsheng009/mint/ui"
)

// 自定义 Intent 类型 - 用于 toast 控制
type ShowInfoToastIntent struct{}
func (ShowInfoToastIntent) IntentType() string { return "ShowInfoToast" }
func (ShowInfoToastIntent) StayPressed() bool  { return true }

type ShowSuccessToastIntent struct{}
func (ShowSuccessToastIntent) IntentType() string { return "ShowSuccessToast" }
func (ShowSuccessToastIntent) StayPressed() bool  { return true }

type ShowWarningToastIntent struct{}
func (ShowWarningToastIntent) IntentType() string { return "ShowWarningToast" }
func (ShowWarningToastIntent) StayPressed() bool  { return true }

type ShowErrorToastIntent struct{}
func (ShowErrorToastIntent) IntentType() string { return "ShowErrorToast" }
func (ShowErrorToastIntent) StayPressed() bool  { return true }

type ClearAllToastsIntent struct{}
func (ClearAllToastsIntent) IntentType() string { return "ClearAllToasts" }
func (ClearAllToastsIntent) StayPressed() bool  { return true }

func main() {
	ui.Run(func() ui.VNode {
		// 1. 定义状态（hooks 必须在顶部）
		infoToast, setInfoToast, _ := ui.UseStateInt(0)
		successToast, setSuccessToast, _ := ui.UseStateInt(0)
		warningToast, setWarningToast, _ := ui.UseStateInt(0)
		errorToast, setErrorToast, _ := ui.UseStateInt(0)

		// 2. 注册 Intent handler（闭包捕获 setter）
		ui.On(ShowInfoToastIntent{}, func() {
			setInfoToast(1)
		})
		ui.On(ShowSuccessToastIntent{}, func() {
			setSuccessToast(1)
		})
		ui.On(ShowWarningToastIntent{}, func() {
			setWarningToast(1)
		})
		ui.On(ShowErrorToastIntent{}, func() {
			setErrorToast(1)
		})
		ui.On(ClearAllToastsIntent{}, func() {
			setInfoToast(0)
			setSuccessToast(0)
			setWarningToast(0)
			setErrorToast(0)
		})

		// 3. 构建 toast 列表
		var toasts []ui.VNode
		if infoToast == 1 {
			toasts = append(toasts, app.ToastBuilder("This is an info message").Info().Build())
		}
		if successToast == 1 {
			toasts = append(toasts, app.ToastBuilder("Operation completed successfully!").Success().Build())
		}
		if warningToast == 1 {
			toasts = append(toasts, app.ToastBuilder("Please check your input").Warning().Build())
		}
		if errorToast == 1 {
			toasts = append(toasts, app.ToastBuilder("An error occurred!").Error().Build())
		}

		// 4. 返回 VNode
		return app.VStack(
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
			app.VStack(toasts...),
		)
	},
		ui.WithWidth(50),
		ui.WithHeight(20),
		ui.WithTitle("Toast Demo"),
	)
}
