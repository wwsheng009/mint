package main

import (
	"reflect"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/ui"
)

// CheckboxDemo demonstrates checkbox component (MVP 模式)
func CheckboxDemo() ui.VNode {
	// 定义状态键
	acceptTermsKey := intent.StateKey[bool]("acceptTerms")
	acceptTermsSetterKey := intent.StateKey[func(bool)]("acceptTermsSetter")
	acceptUpdatesKey := intent.StateKey[bool]("acceptUpdates")
	acceptUpdatesSetterKey := intent.StateKey[func(bool)]("acceptUpdatesSetter")
	acceptPrivacyKey := intent.StateKey[bool]("acceptPrivacy")
	acceptPrivacySetterKey := intent.StateKey[func(bool)]("acceptPrivacySetter")

	// 使用 UseState
	acceptTerms, setAcceptTerms := ui.UseStateBool(false)
	acceptUpdates, setAcceptUpdates := ui.UseStateBool(false)
	acceptPrivacy, setAcceptPrivacy := ui.UseStateBool(false)

	// 保存 setters 到 State
	ctx := ui.GetCurrentContext()
	if ctx != nil {
		ctx.GlobalState[acceptTermsSetterKey.String()] = setAcceptTerms
		ctx.GlobalState[acceptUpdatesSetterKey.String()] = setAcceptUpdates
		ctx.GlobalState[acceptPrivacySetterKey.String()] = setAcceptPrivacy
	}

	// Count checked checkboxes
	checkedCount := 0
	if acceptTerms {
		checkedCount++
	}
	if acceptUpdates {
		checkedCount++
	}
	if acceptPrivacy {
		checkedCount++
	}

	// 注册统一的 Intent Handler（在主函数中注册，只注册一次）
	// 在这里我们使用一个辅助函数来注册 handler，只在第一次渲染时调用
	registerCheckboxHandler := func() {
		if _, ok := ctx.GlobalState["checkboxHandlerRegistered"]; !ok {
			ctx.GlobalState["checkboxHandlerRegistered"] = true
			ui.RegisterIntent(func(ctx *intent.ActionContext, i intent.FieldChangeIntent) intent.IntentResult {
				field := i.Field
				// Checkbox 将 bool 转换为字符串 "true" 或 "false"
				// i.Value 已经是 string 类型，不需要类型断言
				value := i.Value == "true"

				switch field {
				case acceptTermsKey.String():
					setter, _ := ctx.GetState(acceptTermsSetterKey.String())
					callSetter(setter, value)
				case acceptUpdatesKey.String():
					setter, _ := ctx.GetState(acceptUpdatesSetterKey.String())
					callSetter(setter, value)
				case acceptPrivacyKey.String():
					setter, _ := ctx.GetState(acceptPrivacySetterKey.String())
					callSetter(setter, value)
				}
				return intent.HandledResult()
			})
		}
	}

	// 注册 handler（只在第一次渲染时）
	_ = registerCheckboxHandler

	return ui.VStack(
		ui.NewTextBuilder("Checkbox Demo").
			FgColor("cyan").
			Bold(true).
			Build(),
		ui.Text(""),
		ui.NewTextBuilder("Select your preferences:").
			FgColor("bright-black").
			Build(),
		ui.Text(""),
		ui.NewCheckboxBuilder().
			ForField(intent.ForField(acceptTermsKey)).
			Label("I accept the terms and conditions").
			Checked(acceptTerms).
			Build(),
		ui.NewCheckboxBuilder().
			ForField(intent.ForField(acceptUpdatesKey)).
			Label("Subscribe to updates").
			Checked(acceptUpdates).
			Build(),
		ui.NewCheckboxBuilder().
			ForField(intent.ForField(acceptPrivacyKey)).
			Label("I have read the privacy policy").
			Checked(acceptPrivacy).
			Build(),
		ui.Text(""),
		ui.NewTextBuilder("Checked: 1/3").
			FgColor("yellow").
			Build(),
		ui.Text(""),
		ui.NewTextBuilder("Tab: focus | Space/Enter: toggle | q: quit").
			FgColor("bright-black").
			Build(),
	)
}

// callSetter 使用反射调用 setter 函数
func callSetter(fn interface{}, arg interface{}) {
	if fn == nil {
		return
	}
	v := reflect.ValueOf(fn)
	if v.Kind() != reflect.Func {
		return
	}
	argV := reflect.ValueOf(arg)
	v.Call([]reflect.Value{argV})
}

func main() {
	err := ui.Run(CheckboxDemo,
		ui.WithWidth(50),
		ui.WithHeight(18),
		ui.WithTitle("Checkbox Demo (MVP)"),
		ui.WithInit(func() {
			// 注册统一的 FieldChangeIntent Handler
			acceptTermsKey := intent.StateKey[bool]("acceptTerms")
			acceptTermsSetterKey := intent.StateKey[func(bool)]("acceptTermsSetter")
			acceptUpdatesKey := intent.StateKey[bool]("acceptUpdates")
			acceptUpdatesSetterKey := intent.StateKey[func(bool)]("acceptUpdatesSetter")
			acceptPrivacyKey := intent.StateKey[bool]("acceptPrivacy")
			acceptPrivacySetterKey := intent.StateKey[func(bool)]("acceptPrivacySetter")

			ui.RegisterIntent(func(ctx *intent.ActionContext, i intent.FieldChangeIntent) intent.IntentResult {
				field := i.Field
				// Checkbox 将 bool 转换为字符串 "true" 或 "false"
				// i.Value 已经是 string 类型，不需要类型断言
				value := i.Value == "true"

				switch field {
				case acceptTermsKey.String():
					setter, _ := ctx.GetState(acceptTermsSetterKey.String())
					callSetter(setter, value)
				case acceptUpdatesKey.String():
					setter, _ := ctx.GetState(acceptUpdatesSetterKey.String())
					callSetter(setter, value)
				case acceptPrivacyKey.String():
					setter, _ := ctx.GetState(acceptPrivacySetterKey.String())
					callSetter(setter, value)
				}
				return intent.HandledResult()
			})
		}),
	)
	if err != nil {
		panic(err)
	}
}
