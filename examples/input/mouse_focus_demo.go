// Mouse Focus Demo - Demonstrates mouse click focus switching for Input components
package main

import (
	"fmt"

	"github.com/wwsheng009/mint/app"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/ui"
)

// SubmitFormIntent 提交表单
type SubmitFormIntent struct{}
func (SubmitFormIntent) IntentType() string { return "SubmitForm" }
func (SubmitFormIntent) StayPressed() bool  { return true }

// ClearSubmittedStateIntent 清除提交状态
type ClearSubmittedStateIntent struct{}
func (ClearSubmittedStateIntent) IntentType() string { return "ClearSubmitted" }
func (ClearSubmittedStateIntent) StayPressed() bool  { return true }

// MouseFocusDemo demonstrates mouse click focus switching between multiple input fields
func MouseFocusDemo() ui.VNode {
	// Use UseState to create component-level state (single source of truth)
	name, setName := ui.UseStateString("")
	email, setEmail := ui.UseStateString("")
	password, setPassword := ui.UseStateString("")
	submitted, setSubmitted := ui.UseStateBool(false)

	// Save setters to GlobalState for Intent Handler usage
	ctx := ui.GetCurrentContext()
	if ctx != nil {
		ctx.GlobalState["nameSetter"] = setName
		ctx.GlobalState["emailSetter"] = setEmail
		ctx.GlobalState["passwordSetter"] = setPassword
		ctx.GlobalState["submittedSetter"] = setSubmitted
	}

	// Register Submit intent handler
	ui.On(SubmitFormIntent{}, func() {
		setSubmitted(true)
	})

	// Register ClearSubmitted intent handler
	ui.On(ClearSubmittedStateIntent{}, func() {
		setSubmitted(false)
	})

	// Show submitted view
	if submitted {
		return SubmittedView(name, email, password)
	}

	return app.VStack(
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
			app.InputBuilder().
				ForField(intent.BindField("name")).
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
			app.InputBuilder().
				ForField(intent.BindField("email")).
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
			app.InputBuilder().
				ForField(intent.BindField("password")).
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
			app.ButtonBuilder("  Submit  ").
				Variant(app.ButtonVariantPrimary).
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
	return app.VStack(
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
			app.ButtonBuilder("  Back  ").
				Variant(app.ButtonVariantSecondary).
				Key("back-btn").
				OnPress(ClearSubmittedStateIntent{}).
				Build(),
		),
	)
}

func main() {
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║   Input Mouse Focus Demo                                   ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println("")
	fmt.Println("This demo demonstrates:")
	fmt.Println("  - Mouse click to switch focus between inputs")
	fmt.Println("  - Real-time text input with MVP data flow")
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
		ui.WithTitle("Mouse Focus Demo"),
		ui.WithInit(func() {
			// Register FieldChangeIntent handler (form field change)
			ui.RegisterIntent(func(ctx *intent.ActionContext, i intent.FieldChangeIntent) intent.IntentResult {
				switch i.Field {
				case "name":
					setName, _ := ctx.GetState("nameSetter")
					if fn, ok := setName.(func(string)); ok {
						fn(i.Value)
					}
				case "email":
					setEmail, _ := ctx.GetState("emailSetter")
					if fn, ok := setEmail.(func(string)); ok {
						fn(i.Value)
					}
				case "password":
					setPassword, _ := ctx.GetState("passwordSetter")
					if fn, ok := setPassword.(func(string)); ok {
						fn(i.Value)
					}
				}
				return intent.HandledResult()
			})
		}),
	)
	if err != nil {
		fmt.Printf("Error running app: %v\n", err)
	}
}
