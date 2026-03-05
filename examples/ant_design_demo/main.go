// Ant Design Style Demo (Store + Reducer 版本)
//
// 这个示例展示如何在 Mint TUI 中应用 Ant Design 的设计理念
// 包括：表单布局、按钮类型、颜色使用、键盘交互等
//
// Store + Reducer 架构:
// - 使用 AppState 结构体（单一状态源）
// - 使用 Reducer 纯函数处理状态更新
// - 使用 ForField() 统一字段绑定
// - 自动注册所有 Intent Handler（BuildAndRegister）
//
// 架构优势:
//   - 单一状态源
//   - 纯函数 Reducer
//   - 编译期类型检查（无类型断言）
//   - 自动注册（无需手动注册 Handler）
//   - 数据流清晰（UI.Instance → Intent → Reducer → Store → VNode）
//
// 运行: go run main.go

package main

import (
	"fmt"
	"strings"

	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/internal/log"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/reducer"
	"github.com/wwsheng009/mint/runtime/store"
	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/ui"
	"github.com/wwsheng009/mint/ui/components/input"
)

// =============================================================================
// AppState (Single Source of Truth)
// =============================================================================

// AppState represents the application state with all form fields and UI state.
type AppState struct {
	// Form fields
	Username string
	Email    string
	Password string
	Age      string

	// UI state
	Step      int
	Agreed    bool
	ShowModal bool
}

// =============================================================================
// Custom Intent Types (Non-Action Intents)
// =============================================================================

// UpdateStepIntent updates the step number.
type UpdateStepIntent struct {
	Step int
}

func (UpdateStepIntent) IntentType() string { return "UpdateStep" }
func (UpdateStepIntent) StayPressed() bool  { return true }

// ShowModalIntent shows the success modal.
type ShowModalIntent struct{}

func (ShowModalIntent) IntentType() string { return "ShowModal" }
func (ShowModalIntent) StayPressed() bool  { return true }

// QuitIntent quits the application.
type QuitIntent struct{}

func (QuitIntent) IntentType() string { return "Quit" }
func (QuitIntent) StayPressed() bool  { return false }

// CloseModalIntent closes the success modal.
type CloseModalIntent struct{}

func (CloseModalIntent) IntentType() string { return "CloseModal" }
func (CloseModalIntent) StayPressed() bool  { return true }

// =============================================================================
// Reducer (Pure Function)
// =============================================================================

// appReducer handles all state transitions.
var appReducer = reducer.NewBuilder[AppState]()

// Initialize the reducer with all handlers.
func init() {
	// Handle FieldChangeIntent - update form fields automatically
	appReducer.On(intent.FieldChangeIntent{}, func(s AppState, i intent.Intent) AppState {
		fci := i.(intent.FieldChangeIntent)
		switch fci.Field {
		case "username":
			s.Username = fci.Value
		case "email":
			s.Email = fci.Value
		case "password":
			s.Password = fci.Value
		case "age":
			s.Age = fci.Value
		case "agreed":
			s.Agreed = fci.Value == "true"
		}
		return s
	})

	// Handle UpdateStepIntent - update the step number
	appReducer.On(UpdateStepIntent{}, func(s AppState, i intent.Intent) AppState {
		usi := i.(UpdateStepIntent)
		s.Step = usi.Step
		log.TempLogger.Debug("Step updated to: %d", usi.Step)
		return s
	})

	// Handle ShowModalIntent - show the success modal
	appReducer.On(ShowModalIntent{}, func(s AppState, i intent.Intent) AppState {
		s.ShowModal = true
		return s
	})

	// Handle CloseModalIntent - close the modal
	appReducer.On(CloseModalIntent{}, func(s AppState, i intent.Intent) AppState {
		s.ShowModal = false
		return s
	})

	// Handle QuitIntent - quit the application
	appReducer.On(QuitIntent{}, func(s AppState, i intent.Intent) AppState {
		ui.Quit()
		return s
	})
}

// =============================================================================
// Store (Single State Source)
// =============================================================================

// appStore holds the application state.
var appStore = store.NewStore(AppState{
	Username: "",
	Email:    "",
	Password: "",
	Age:      "",
	Step:     1,
	Agreed:   false,
	ShowModal: false,
})

// =============================================================================
// Main Function
// =============================================================================

func main() {
	// Set the theme to Ant Design's recommended color scheme
	_ = theme.SetTheme("nord")

	// Register all handlers automatically
	appReducer.RegisterToGlobal(appStore)

	err := ui.Run(App,
		ui.WithWidth(80),
		ui.WithHeight(30),
		ui.WithTitle("Mint TUI - Ant Design Style (Store+Reducer)"),
	)
	if err != nil {
		panic(err)
	}
}

// =============================================================================
// Main App Component
// =============================================================================

func App() ui.VNode {
	// Get current state snapshot
	state := appStore.Get()

	// If showing modal, return ModalView
	if state.ShowModal {
		return ModalView()
	}

	// Otherwise return the main form view
	return ui.VStackBuilder(
		Header(),
		StepIndicator(state.Step),
		ProgressBar(state.Step),
		FormContent(state.Username, state.Email, state.Password, state.Age, state.Agreed, state.Step),
		ActionButtons(state.Step),
		Footer(),
	).Gap(1).Build()
}

// ModalView - Success modal view
func ModalView() ui.VNode {
	return ui.VStackBuilder(
		ui.HStackBuilder(
			ui.Text(""),
			ui.NewTextBuilder("✓ Registration Successful!").
				Style(style.Style{}.Foreground(theme.Success()).Bold(true)).
				Build(),
			ui.Text(""),
		).Build(),
		ui.Text(""),
		ui.NewTextBuilder("Thank you for registering with us.").
			Style(style.Style{}).
			Build(),
		ui.Text(""),
		ui.HStack(
			ui.Text(""),
			ui.NewButtonBuilder("[ Close ]").
				Variant(ui.ButtonVariantPrimary).
				OnPress(CloseModalIntent{}).
				Build(),
			ui.Text(""),
		),
	).Gap(1).Build()
}

// =============================================================================
// UI Components
// =============================================================================

// Header - Page header
func Header() ui.VNode {
	return ui.NewVStack().
		SingleBorder().
		BorderColor(theme.Primary()).
		SetChildrenList([]ui.VNode{
			ui.HStackBuilder(
				ui.NewTextBuilder("📝 User Registration Form").
					Style(style.Style{}.
						Foreground(theme.BG()).
						Background(theme.Primary()).
						Bold(true)).
					Build(),
			).Align(ui.AlignCenter).Build(),
		})
}

// StepIndicator - Step indicator (Ant Design Steps component)
func StepIndicator(step int) ui.VNode {
	steps := []string{"Account", "Profile", "Confirm"}

	items := make([]ui.VNode, len(steps))
	for i, s := range steps {
		var itemStyle style.Style
		if i < step {
			itemStyle = style.Style{}.
				Foreground(theme.Success()).
				Bold(true)
		} else if i == step {
			itemStyle = style.Style{}.
				Foreground(theme.Primary()).
				Bold(true)
		} else {
			itemStyle = style.Style{}.
				Foreground(theme.Muted())
		}

		items[i] = ui.HStackBuilder(
			ui.NewTextBuilder(fmt.Sprintf("%d. %s", i+1, s)).
				Style(itemStyle).
				Build(),
		).Align(ui.AlignCenter).Build()
	}

	return ui.HStackBuilder(items...).Gap(4).Build()
}

// ProgressBar - Progress bar (Ant Design Progress component)
func ProgressBar(step int) ui.VNode {
	const totalSteps = 3
	progress := step * 30 / totalSteps

	track := ui.NewTextBuilder("┌" + strings.Repeat("─", 30) + "┐").
		Style(style.Style{}.Foreground(theme.Border())).
		Build()

	fill := ui.NewTextBuilder("│" + strings.Repeat("━", progress) + strings.Repeat("─", 30-progress) + "│").
		Style(style.Style{}.Foreground(theme.Primary())).
		Build()

	return ui.VStack(
		track,
		fill,
		ui.NewTextBuilder("└"+strings.Repeat("─", 30)+"┐").
			Style(style.Style{}.Foreground(theme.Border())).
			Build(),
	)
}

// FormContent - Form content (changes based on step)
func FormContent(
	username string,
	email string,
	password string,
	age string,
	agreed bool,
	step int,
) ui.VNode {
	if step == 1 {
		// Step 1: Account Information
		return ui.NewVStack().
			SingleBorder().
			SetChildrenList([]ui.VNode{
				ui.VStackBuilder(
					FormItem(username, "Enter your username", 24, "username", "", true),
					FormItem(email, "example@domain.com", 24, "email", "We'll never share your email", true),
					FormItemPassword(password, "Enter your password", 24, "password", "At least 8 characters"),
				).Gap(2).Build(),
			})
	} else if step == 2 {
		// Step 2: Profile
		return ui.NewVStack().
			SingleBorder().
			SetChildrenList([]ui.VNode{
				ui.VStackBuilder(
					FormItem(age, "Your age", 10, "age", "", true),
				).Gap(1).Build(),
			})
	} else {
		// Step 3: Confirm
		return ui.NewVStack().
			SingleBorder().
			SetChildrenList([]ui.VNode{
				ui.VStackBuilder(
					ConfirmInfo("Username:", username),
					ConfirmInfo("Email:", email),
					ConfirmInfo("Age:", age),
					ui.HStackBuilder(
						ui.Text("         "),
						ui.NewCheckboxBuilder().
							ForField(intent.BindField("agreed")).
							Checked(agreed).
							Label("I agree to the Terms and Conditions").
							Build(),
					).Build(),
				).Gap(2).Build(),
			})
	}
}

// FormItem - Ant Design style form item
func FormItem(
	value string,
	placeholder string,
	width int,
	fieldName string,
	helpText string,
	required bool,
) ui.VNode {
	labelWidth := 10

	var requiredMark ui.VNode
	if required {
		requiredMark = ui.NewTextBuilder("*").
			Style(style.Style{}.Foreground(theme.Error())).
			Build()
	} else {
		requiredMark = ui.Text("")
	}

	var helpNode ui.VNode
	if helpText != "" {
		helpNode = ui.NewTextBuilder(helpText).
			Style(style.Style{}.Foreground(theme.Muted())).
			Build()
	} else {
		helpNode = ui.Text("")
	}

	// Create label from field name
	var label string
	switch fieldName {
	case "username":
		label = "Username:"
	case "email":
		label = "Email:"
	case "password":
		label = "Password:"
	case "age":
		label = "Age:"
	default:
		label = fieldName + ":"
	}

	return ui.VStackBuilder(
		ui.HStackBuilder(
			ui.NewTextBuilder(fmt.Sprintf("%-*s", labelWidth, label)).
				Style(style.Style{}.
					Foreground(theme.Text()).
					Bold(true)).
				Build(),
			ui.Text(" "),
			ui.NewInputBuilder().
				ForField(intent.BindField(fieldName)).
				Value(value).
				Placeholder(placeholder).
				Type(input.TypeText).
				Build(),
			requiredMark,
		).Build(),
		ui.HStackBuilder(
			ui.Text(strings.Repeat(" ", labelWidth+1)),
			helpNode,
		).Build(),
	).Gap(1).Build()
}

// FormItemPassword - Password input form item
func FormItemPassword(
	value string,
	placeholder string,
	width int,
	fieldName string,
	helpText string,
) ui.VNode {
	labelWidth := 10

	var helpNode ui.VNode
	if helpText != "" {
		helpNode = ui.NewTextBuilder(helpText).
			Style(style.Style{}.Foreground(theme.Muted())).
			Build()
	} else {
		helpNode = ui.Text("")
	}

	return ui.VStackBuilder(
		ui.HStackBuilder(
			ui.NewTextBuilder(fmt.Sprintf("%-*s", labelWidth, "Password:")).
				Style(style.Style{}.
					Foreground(theme.Text()).
					Bold(true)).
				Build(),
			ui.Text(" "),
			ui.NewInputBuilder().
				ForField(intent.BindField(fieldName)).
				Value(value).
				Password().
				Placeholder(placeholder).
				Build(),
		).Build(),
		ui.HStackBuilder(
			ui.Text(strings.Repeat(" ", labelWidth+1)),
			helpNode,
		).Build(),
	).Gap(1).Build()
}

// ConfirmInfo - Confirm page info display
func ConfirmInfo(label, value string) ui.VNode {
	labelWidth := 10

	return ui.HStackBuilder(
		ui.NewTextBuilder(fmt.Sprintf("%-*s", labelWidth, label)).
			Style(style.Style{}.
				Foreground(theme.Muted()).
				Bold(true)).
			Build(),
		ui.NewTextBuilder(value).
			Style(style.Style{}.
				Foreground(theme.Text())).
			Build(),
	).Build()
}

// ActionButtons - Action buttons group
func ActionButtons(step int) ui.VNode {
	const totalSteps = 3
	var buttons []ui.VNode

	if step > 1 {
		buttons = append(buttons,
			ui.NewButtonBuilder("[ Previous ]").
				Variant(ui.ButtonVariantSecondary).
				OnPress(UpdateStepIntent{Step: step - 1}).
				Build(),
		)
	}

	if step < totalSteps {
		buttons = append(buttons,
			ui.NewButtonBuilder("[ Next ]").
				Variant(ui.ButtonVariantPrimary).
				OnPress(UpdateStepIntent{Step: step + 1}).
				Build(),
		)
	} else {
		buttons = append(buttons,
			ui.NewButtonBuilder("[ Submit ]").
				Variant(ui.ButtonVariantPrimary).
				OnPress(ShowModalIntent{}).
				Build(),
		)
	}

	buttons = append(buttons,
		ui.NewButtonBuilder("[ Cancel ]").
			Variant(ui.ButtonVariantDefault).
			OnPress(QuitIntent{}).
			Build(),
	)

	return ui.HStackBuilder(buttons...).Gap(2).Build()
}

// Footer - Page footer
func Footer() ui.VNode {
	return ui.HStackBuilder(
		ui.NewTextBuilder("Press Tab to navigate • Enter to select • Esc to quit").
			Style(style.Style{}.Foreground(theme.Placeholder())).
			Build(),
	).Align(ui.AlignCenter).Build()
}
