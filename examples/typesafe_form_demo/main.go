// Package main demonstrates type-safe form handling with Store + Reducer.
//
// 优化版本：使用字段映射减少硬编码
//
package main

import (
	"fmt"

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
// Optimized Reducer (Using Field Mapping)
// =============================================================================

// FieldType 表示字段类型
type FieldType int

const (
	TypeString FieldType = iota
	TypeBool
	TypeInt
)

// FieldUpdate 表示字段更新器
type FieldUpdate struct {
	Name  string
	Type  FieldType
	Update func(s *AppState, val string) // val 是字符串表示
}

// 字段映射表 - 避免硬编码 switch
var fieldUpdates = []FieldUpdate{
	{
		Name: "Username",
		Type: TypeString,
		Update: func(s *AppState, val string) {
			s.Username = val
		},
	},
	{
		Name: "Email",
		Type: TypeString,
		Update: func(s *AppState, val string) {
			s.Email = val
		},
	},
	{
		Name: "Age",
		Type: TypeInt,
		Update: func(s *AppState, val string) {
			var ageVal int
			if _, err := fmt.Sscanf(val, "%d", &ageVal); err == nil {
				s.Age = ageVal
			}
		},
	},
	{
		Name: "AcceptTerms",
		Type: TypeBool,
		Update: func(s *AppState, val string) {
			s.AcceptTerms = val == "true"
		},
	},
	{
		Name: "Subscribe",
		Type: TypeBool,
		Update: func(s *AppState, val string) {
			s.Subscribe = val == "true"
		},
	},
}

// 查找字段更新器（避免遍历，可以用 map 优化）
func findFieldUpdate(fieldName string) *FieldUpdate {
	for i := range fieldUpdates {
		if fieldUpdates[i].Name == fieldName {
			return &fieldUpdates[i]
		}
	}
	return nil
}

// =============================================================================
// Reducer (Using Field Mapping) / 纯函数
// =============================================================================

// appReducer handles all state transitions for the form.
var appReducer = reducer.NewBuilder[AppState]()

// Initialize the reducer - ForField components emit FieldChangeIntent automatically
func init() {
	// 方案 1: 使用字段映射表（推荐）
	appReducer.On(intent.FieldChangeIntent{}, func(s AppState, i intent.Intent) AppState {
		fci := i.(intent.FieldChangeIntent)
		if update := findFieldUpdate(fci.Field); update != nil {
			update.Update(&s, fci.Value)
		}
		return s
	})

	// 方案 2: 处理 SubmitIntent - log the form data
	appReducer.On(SubmitIntent{}, func(s AppState, i intent.Intent) AppState {
		fmt.Println("=== Form Submission ===")
		fmt.Printf("Username:   %v\n", s.Username)
		fmt.Printf("Email:      %v\n", s.Email)
		fmt.Printf("Age:        %v\n", s.Age)
		fmt.Printf("Accept T&C: %v\n", s.AcceptTerms)
		fmt.Printf("Subscribe:  %v\n", s.Subscribe)
		fmt.Println("========================")
		return s
	})
}

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
		Value(fmt.Sprintf("%d", state.Age)).
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
	appReducer.RegisterToGlobal(appStore)

	ui.Run(App,
		ui.WithWidth(50),
		ui.WithHeight(18),
		ui.WithTitle("Type-Safe Form Demo (Store+Reducer)"),
	)
}
