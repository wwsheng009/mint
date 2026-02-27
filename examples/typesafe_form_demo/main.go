// Package main demonstrates type-safe form handling with StateKey[T].
//
// 采用【模式3：自定义 Intent】+ FieldChangeIntent + 类型安全 StateKey[T]
//
// 三种 Intent 管理模式：
//   1. 组件级状态 - ui.On + UseState + Simple* Intent（推荐组件内状态）
//   2. 全局状态 - runtime/intent 内置函数
//   3. 自定义 Intent + ui.On（本示例）
//
// This example shows the full MVP data flow with compile-time type safety:
//   1. Define StateKey[T] package-level keys (type-safe field identifiers)
//   2. Use ForField() to bind components to StateKey[T]
//   3. Register a FieldChangeIntent handler in WithInit
//   4. Instances emit intents carrying values
//   5. State updates become the single source of truth
//
// Benefits:
//   - No string key typos
//   - IDE autocomplete support
//   - Refactor-safe field renames
//   - Type-checked values
package main

import (
	"fmt"

	"github.com/wwsheng009/mint/runtime/intent"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui"
	"github.com/wwsheng009/mint/ui/components/button"
	"github.com/wwsheng009/mint/ui/components/checkbox"
	"github.com/wwsheng009/mint/ui/components/input"
)

// =============================================================================
// Type-Safe State Keys (Package Level)
// =============================================================================

// Form state keys - these provide compile-time type safety.
var (
	keyUsername = intent.StateKey[string]("username")
	keyEmail    = intent.StateKey[string]("email")

	keyAcceptTerms = intent.StateKey[bool]("acceptTerms")
	keySubscribe   = intent.StateKey[bool]("subscribe")

	keyAge = intent.StateKey[int]("age")
)

// =============================================================================
// Submit Intent
// =============================================================================

type SubmitIntent struct{}
func (SubmitIntent) IntentType() string { return "Submit" }
func (SubmitIntent) StayPressed() bool  { return true }

// =============================================================================
// Form Component Builders
// =============================================================================

var submitHandlerUnregister func()

func App() ui.VNode {
	// 使用 UseState 管理表单状态
	username, setUsername := ui.UseStateString("")
	email, setEmail := ui.UseStateString("")
	age, setAge, _ := ui.UseStateInt(0)
	acceptTerms, setAcceptTerms := ui.UseStateBool(false)
	subscribe, setSubscribe := ui.UseStateBool(false)

	// 保存 setters 到 GlobalState 供 Intent Handler 使用
	ctx := ui.GetCurrentContext()
	if ctx != nil {
		ctx.GlobalState[keyUsername.String()+"Setter"] = setUsername
		ctx.GlobalState[keyEmail.String()+"Setter"] = setEmail
		ctx.GlobalState[keyAge.String()+"Setter"] = setAge
		ctx.GlobalState[keyAcceptTerms.String()+"Setter"] = setAcceptTerms
		ctx.GlobalState[keySubscribe.String()+"Setter"] = setSubscribe
	}

	// Unregister previous handler (if exists) to avoid memory leak
	if submitHandlerUnregister != nil {
		submitHandlerUnregister()
	}

	// Register SubmitIntent handler - captures current values via closure
	// Using rtui.RegisterIntent instead of ui.On to re-register on each render
	submitHandlerUnregister = rtui.RegisterIntent(func(ctx *intent.ActionContext, i SubmitIntent) intent.IntentResult {
		fmt.Println("\n=== Form Submission ===")
		fmt.Printf("Username:   %v\n", username)
		fmt.Printf("Email:      %v\n", email)
		fmt.Printf("Age:        %v\n", age)
		fmt.Printf("Accept T&C: %v\n", acceptTerms)
		fmt.Printf("Subscribe:  %v\n", subscribe)
		fmt.Println("========================\n")
		return intent.HandledResult()
	})

	// Define form components with current values
	usernameInput := input.NewBuilder().
		Placeholder("Enter your username").
		ForField(intent.ForField(keyUsername)).
		Value(username).
		Width(30).
		Build()

	emailInput := input.NewBuilder().
		Placeholder("Enter your email").
		ForField(intent.ForField(keyEmail)).
		Value(email).
		Width(30).
		Build()

	ageInput := input.NewBuilder().
		Placeholder("Enter your age").
		ForField(intent.ForField(keyAge)).
		Value(fmt.Sprintf("%d", age)).
		Width(10).
		Build()

	termsCheckbox := checkbox.NewBuilder().
		Label("I accept the terms and conditions").
		ForField(intent.ForField(keyAcceptTerms)).
		Checked(acceptTerms).
		Build()

	subscribeCheckbox := checkbox.NewBuilder().
		Label("Subscribe to newsletter").
		ForField(intent.ForField(keySubscribe)).
		Checked(subscribe).
		Build()

	submitButton := button.NewBuilder("Submit Form").
		OnPress(SubmitIntent{}).
		Variant(button.VariantPrimary).
		Build()

	// Build the form layout
	form := ui.VStack(
		ui.Text("=== Type-Safe Registration Form ==="),
		ui.Text(""),

		ui.HStack(ui.Text("Username: "), ui.Text("  "), usernameInput),
		ui.HStack(ui.Text("Email:    "), ui.Text("  "), emailInput),
		ui.HStack(ui.Text("Age:      "), ui.Text("  "), ageInput),

		ui.Text(""),

		termsCheckbox,
		subscribeCheckbox,

		ui.Text(""),

		ui.HStack(ui.Text("  "), ui.Text("  "), submitButton),
	)

	return form
}

func main() {
	ui.Run(App,
		ui.WithWidth(50),
		ui.WithHeight(18),
		ui.WithTitle("Type-Safe Form Demo"),
		ui.WithInit(func() {
			// FieldChangeIntent handler registered in WithInit as recommended
			ui.RegisterIntent(func(actx *intent.ActionContext, i intent.FieldChangeIntent) intent.IntentResult {
				fmt.Printf("Intent Received: Field='%s', Value='%s'\n", i.Field, i.Value)

				switch i.Field {
				case keyUsername.String():
					val, _ := actx.GetState(keyUsername.String() + "Setter")
					if fn, ok := val.(func(string)); ok {
						fn(i.Value)
					}
				case keyEmail.String():
					val, _ := actx.GetState(keyEmail.String() + "Setter")
					if fn, ok := val.(func(string)); ok {
						fn(i.Value)
					}
				case keyAge.String():
					val, _ := actx.GetState(keyAge.String() + "Setter")
					if fn, ok := val.(func(interface{})); ok {
						var ageVal int
						if _, err := fmt.Sscanf(i.Value, "%d", &ageVal); err == nil {
							fn(ageVal)
						}
					}
				case keyAcceptTerms.String():
					val, _ := actx.GetState(keyAcceptTerms.String() + "Setter")
					if fn, ok := val.(func(bool)); ok {
						agreeVal := i.Value == "true"
						fn(agreeVal)
					}
				case keySubscribe.String():
					val, _ := actx.GetState(keySubscribe.String() + "Setter")
					if fn, ok := val.(func(bool)); ok {
						subscribeVal := i.Value == "true"
						fn(subscribeVal)
					}
				}
				return intent.HandledResult()
			})
		}),
	)
}
