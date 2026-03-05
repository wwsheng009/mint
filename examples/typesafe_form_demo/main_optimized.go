// Package main demonstrates type-safe form handling with Store + Reducer.
//
// 优化版本：使用 FieldMap 消除硬编码
//
// 架构优势：
//   - 类型安全（泛型）
//   - 消除 switch-case 硬编码
//   - 使用映射表替代重复的 On FieldChangeIntent
//   - 单一 FieldChangeIntent 处理器处理所有字段
//
package main

import (
	"fmt"
	"strconv"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/reducer"
	"github.com/wwsheng009/mint/runtime/store"
	"github.com/wwsheng009/mint/ui"
	"github.com/wwsheng009/mint/ui/components/button"
	"github.com/wwsheng009/mint/ui/components/checkbox"
	"github.com/wwsheng009/mint/ui/components/input"
)

// =============================================================================
// AppState (Single Source of Truth)
// =============================================================================

// AppState represents the form state with all fields.
type AppState struct {
	Username    string
	Email       string
	Age         int
	AcceptTerms bool
	Subscribe   bool
}

// =============================================================================
// Submit Intent
// =============================================================================

type SubmitIntent struct{}

func (SubmitIntent) IntentType() string { return "Submit" }
func (SubmitIntent) StayPressed() bool  { return true }

// =============================================================================
// Field Map (消除硬编码)
// =============================================================================

// fieldMap 定义所有字段的更新逻辑，避免 switch-case
var fieldMap = reducer.BindField(reducer.NewBuilder[AppState]()).
	BindFieldMap(map[string]func(AppState, string) AppState{
		// 字符串字段
		"Username": func(s AppState, val string) AppState {
			s.Username = val
			return s
		},
		"Email": func(s AppState, val string) AppState {
			s.Email = val
			return s
		},
		// 整型字段
		"Age": func(s AppState, val string) AppState {
			if v, err := strconv.Atoi(val); err == nil {
				s.Age = v
			}
			return s
		},
		// 布尔字段
		"AcceptTerms": func(s AppState, val string) AppState {
			s.AcceptTerms = val == "true"
			return s
		},
		"Subscribe": func(s AppState, val string) AppState {
			s.Subscribe = val == "true"
			return s
		},
	}).
	GetBuilder().
	On(SubmitIntent{}, func(s AppState, i intent.Intent) AppState {
		fmt.Println("\n=== Form Submission ===")
		fmt.Printf("Username:   %v\n", s.Username)
		fmt.Printf("Email:      %v\n", s.Email)
		fmt.Printf("Age:        %v\n", s.Age)
		fmt.Printf("Accept T&C: %v\n", s.AcceptTerms)
		fmt.Printf("Subscribe:  %v\n", s.Subscribe)
		fmt.Println("========================\n")
		return s
	})

// =============================================================================
// Store (Single State Source)
// =============================================================================

// appStore holds the application state.
var appStore = store.NewStore(AppState{
	Username:    "",
	Email:       "",
	Age:         0,
	AcceptTerms: false,
	Subscribe:   false,
})

// =============================================================================
// App Component
// =============================================================================

func App() ui.VNode {
	// Get current state snapshot
	state := appStore.Get()

	// Build form components with ForField binding
	usernameInput := input.NewBuilder().
		Placeholder("Enter your username").
		ForField(intent.BindField("Username")).
		Value(state.Username).
		Width(30).
		Build()

	emailInput := input.NewBuilder().
		Placeholder("Enter your email").
		ForField(intent.BindField("Email")).
		Value(state.Email).
		Width(30).
		Build()

	ageInput := input.NewBuilder().
		Placeholder("Enter your age").
		ForField(intent.BindField("Age")).
		Value(reducer.FormatInt(state.Age)).
		Width(10).
		Build()

	termsCheckbox := checkbox.NewBuilder().
		Label("I accept the terms and conditions").
		ForField(intent.BindField("AcceptTerms")).
		Checked(state.AcceptTerms).
		Build()

	subscribeCheckbox := checkbox.NewBuilder().
		Label("Subscribe to newsletter").
		ForField(intent.BindField("Subscribe")).
		Checked(state.Subscribe).
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
	// Register all handlers automatically
	fieldMap.RegisterToGlobal(appStore)

	ui.Run(App,
		ui.WithWidth(50),
		ui.WithHeight(18),
		ui.WithTitle("Type-Safe Form Demo (Store+Reducer Optimized)"),
	)
}
