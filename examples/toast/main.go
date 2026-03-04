package main

import (
	"github.com/wwsheng009/mint/runtime/intent"
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

		// 将 setter 保存到 GlobalState，供 handler 从 ActionContext 读取
		ctx := ui.GetCurrentContext()
		if ctx != nil {
			ctx.GlobalState["setInfoToast"] = setInfoToast
			ctx.GlobalState["setSuccessToast"] = setSuccessToast
			ctx.GlobalState["setWarningToast"] = setWarningToast
			ctx.GlobalState["setErrorToast"] = setErrorToast
		}

		// 2. 注册 Intent handler（从 ActionContext 读取状态）
		ui.On(ShowInfoToastIntent{}, func(actx *intent.ActionContext) {
			if fn, ok := actx.GetState("setInfoToast"); ok {
				if setter, ok := fn.(func(int)); ok {
					setter(1)
				}
			}
		})
		ui.On(ShowSuccessToastIntent{}, func(actx *intent.ActionContext) {
			if fn, ok := actx.GetState("setSuccessToast"); ok {
				if setter, ok := fn.(func(int)); ok {
					setter(1)
				}
			}
		})
		ui.On(ShowWarningToastIntent{}, func(actx *intent.ActionContext) {
			if fn, ok := actx.GetState("setWarningToast"); ok {
				if setter, ok := fn.(func(int)); ok {
					setter(1)
				}
			}
		})
		ui.On(ShowErrorToastIntent{}, func(actx *intent.ActionContext) {
			if fn, ok := actx.GetState("setErrorToast"); ok {
				if setter, ok := fn.(func(int)); ok {
					setter(1)
				}
			}
		})
		ui.On(ClearAllToastsIntent{}, func(actx *intent.ActionContext) {
			if fn, ok := actx.GetState("setInfoToast"); ok {
				if setter, ok := fn.(func(int)); ok {
					setter(0)
				}
			}
			if fn, ok := actx.GetState("setSuccessToast"); ok {
				if setter, ok := fn.(func(int)); ok {
					setter(0)
				}
			}
			if fn, ok := actx.GetState("setWarningToast"); ok {
				if setter, ok := fn.(func(int)); ok {
					setter(0)
				}
			}
			if fn, ok := actx.GetState("setErrorToast"); ok {
				if setter, ok := fn.(func(int)); ok {
					setter(0)
				}
			}
		})

		// 3. 构建 toast 列表
		var toasts []ui.VNode
		if infoToast == 1 {
			toasts = append(toasts, ui.NewToastBuilder("This is an info message").Info().Build())
		}
		if successToast == 1 {
			toasts = append(toasts, ui.NewToastBuilder("Operation completed successfully!").Success().Build())
		}
		if warningToast == 1 {
			toasts = append(toasts, ui.NewToastBuilder("Please check your input").Warning().Build())
		}
		if errorToast == 1 {
			toasts = append(toasts, ui.NewToastBuilder("An error occurred!").Error().Build())
		}

		// 4. 返回 VNode
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
	},
		ui.WithWidth(50),
		ui.WithHeight(20),
		ui.WithTitle("Toast Demo"),
	)
}
