// Mouse Focus Demo - Demonstrates mouse click focus switching for Input components (Store 模式)
package main

import (
	"fmt"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/reducer"
	"github.com/wwsheng009/mint/runtime/store"
	"github.com/wwsheng009/mint/ui"
)

// =============================================================================
// AppState - 定义应用状态
// =============================================================================

type AppState struct {
	Name      string // 姓名
	Email     string // 邮箱
	Password  string // 密码
	Submitted bool   // 是否已提交
}

// =============================================================================
// Intent Types
// =============================================================================

type SubmitFormIntent struct{}
func (SubmitFormIntent) IntentType() string { return "SubmitForm" }
func (SubmitFormIntent) StayPressed() bool  { return true }

type ClearSubmittedStateIntent struct{}
func (ClearSubmittedStateIntent) IntentType() string { return "ClearSubmitted" }
func (ClearSubmittedStateIntent) StayPressed() bool  { return true }

type SetInputNameIntent struct {
	Name string
}
func (SetInputNameIntent) IntentType() string { return "SetInputName" }
func (SetInputNameIntent) StayPressed() bool  { return false }

type SetInputEmailIntent struct {
	Email string
}
func (SetInputEmailIntent) IntentType() string { return "SetInputEmail" }
func (SetInputEmailIntent) StayPressed() bool  { return false }

type SetInputPasswordIntent struct {
	Password string
}
func (SetInputPasswordIntent) IntentType() string { return "SetInputPassword" }
func (SetInputPasswordIntent) StayPressed() bool  { return false }

// =============================================================================
// Store 初始化
// ============================================================================

var inputDemoStore = store.NewStore(AppState{
	Name:      "",
	Email:     "",
	Password:  "",
	Submitted: false,
})

// =============================================================================
// Reducer 注册
// ============================================================================

func init() {
	reducer.NewBuilder[AppState]().
		On(SubmitFormIntent{}, func(s AppState, i intent.Intent) AppState {
			s.Submitted = true
			return s
		}).
		On(ClearSubmittedStateIntent{}, func(s AppState, i intent.Intent) AppState {
			s.Submitted = false
			return s
		}).
		On(SetInputNameIntent{}, func(s AppState, i intent.Intent) AppState {
			s.Name = i.(SetInputNameIntent).Name
			return s
		}).
		On(SetInputEmailIntent{}, func(s AppState, i intent.Intent) AppState {
			s.Email = i.(SetInputEmailIntent).Email
			return s
		}).
		On(SetInputPasswordIntent{}, func(s AppState, i intent.Intent) AppState {
			s.Password = i.(SetInputPasswordIntent).Password
			return s
		}).
		BuildAndRegister(intent.DefaultRegistry(), inputDemoStore)
}

// =============================================================================
// MouseFocusDemo - 演示多个输入字段的鼠标点击焦点切换
// ============================================================================

func MouseFocusDemo() ui.VNode {
	// ✅ 订阅存储的状态
	name := ui.UseStoreSelector(inputDemoStore, func(s AppState) string { return s.Name })
	email := ui.UseStoreSelector(inputDemoStore, func(s AppState) string { return s.Email })
	password := ui.UseStoreSelector(inputDemoStore, func(s AppState) string { return s.Password })
	submitted := ui.UseStoreSelector(inputDemoStore, func(s AppState) bool { return s.Submitted })

	// Show submitted view
	if submitted {
		return SubmittedView(name, email, password)
	}

	return ui.VStack(
		ui.NewTextBuilder("=== Mouse Focus Demo ===").
			Bold(true).
			Build(),
		ui.Text(""),
		ui.NewTextBuilder("Click on any input to set focus").
			FgColor("cyan").
			Build(),
		ui.Text(""),

		// Multiple input fields with MVP data flow
		ui.NewTextBuilder("Name:").
			Build(),
		ui.HStack(
			ui.Text("  "),
			ui.NewInputBuilder().
				Value(name).
				Placeholder("Enter your name").
				Width(30).
				Key("name-input").
				Build(),
		),

		ui.NewTextBuilder("Email:").
			Build(),
		ui.HStack(
			ui.Text("  "),
			ui.NewInputBuilder().
				Value(email).
				Placeholder("Enter your email").
				Width(30).
				Key("email-input").
				Build(),
		),

		ui.NewTextBuilder("Password:").
			Build(),
		ui.HStack(
			ui.Text("  "),
			ui.NewInputBuilder().
				Value(password).
				Placeholder("Enter password").
				Password().
				Width(30).
				Key("password-input").
				Build(),
		),

		ui.Text(""),

		// Submit button
		ui.HStack(
			ui.Text("  "),
			ui.NewButtonBuilder("  Submit  ").
				Variant(ui.ButtonVariantPrimary).
				Key("submit-btn").
				OnPress(SubmitFormIntent{}).
				Disabled(name == "" || email == "" || password == "").
				Build(),
		),

		ui.Text(""),

		// Instructions
		ui.NewTextBuilder("Instructions:").
			FgColor("yellow").
			Build(),
		ui.Text("  • Mouse Click:  Click an input/button to focus"),
		ui.Text("  • Tab:          Navigate to next focusable"),
		ui.Text("  • SHIFT+Tab:    Navigate to previous"),
		ui.Text("  • Type:         Enter text in focused input"),
		ui.Text("  • Backspace:    Delete character"),
		ui.Text("  • q:            Quit"),
	)
}

// SubmittedView - 显示提交成功的视图
func SubmittedView(name, email, password string) ui.VNode {
	return ui.VStack(
		ui.NewTextBuilder("✅ Form Submitted!").
			Bold(true).
			FgColor("green").
			Build(),
		ui.Text(""),
		ui.NewTextBuilder("─").FgColor("gray").Build(),
		ui.Text(""),
		ui.NewTextBuilder(fmt.Sprintf("Name: %s", name)).Build(),
		ui.NewTextBuilder(fmt.Sprintf("Email: %s", email)).Build(),
		ui.Text(""),
		ui.NewTextBuilder("─").FgColor("gray").Build(),
		ui.Text(""),
		ui.HStack(
			ui.Text("  "),
			ui.NewButtonBuilder("  Back  ").
				Variant(ui.ButtonVariantSecondary).
				Key("back-btn").
				OnPress(ClearSubmittedStateIntent{}).
				Build(),
		),
	)
}

// =============================================================================
// Main
// ============================================================================

func main() {
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║   Input Mouse Focus Demo                                   ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println("")
	fmt.Println("This demo demonstrates:")
	fmt.Println("  - Mouse click to switch focus between inputs")
	fmt.Println("  - Real-time text input with Store-based state")
	fmt.Println("")
	fmt.Println("Expected Behavior:")
	fmt.Println("  ✓ Click on Name/Email/Password - focus moves to clicked field")
	fmt.Println("  ✓ Focused field shows cyan foreground, underline, and bold text")
	fmt.Println("  ✓ Type to enter text - value is stored in component state")
	fmt.Println("  ✓ Tab key also works for keyboard navigation")
	fmt.Println("  ✓ Submit button enables when all fields are filled")
	fmt.Println("")
	fmt.Println("Starting demo...")

	err := ui.Run(MouseFocusDemo,
		ui.WithWidth(50),
		ui.WithHeight(35),
		ui.WithTitle("Mouse Focus Demo (Store 模式)"),
	)
	if err != nil {
		fmt.Printf("Error running app: %v\n", err)
	}
}
