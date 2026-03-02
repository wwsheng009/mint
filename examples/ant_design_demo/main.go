// Ant Design Style Demo (MVP 迁移版)
//
// 这个示例展示如何在 Mint TUI 中应用 Ant Design 的设计理念
// 包括：表单布局、按钮类型、颜色使用、键盘交互等
//
// MVP 架构迁移:
// - 使用 StateKey[T] 类型安全字段键
// - 使用 ForField() 统一字段绑定
// - 统一的 FieldChangeIntent Handler
//
// 运行: go run main.go

package main

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/wwsheng009/mint/app"
	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/internal/log"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/ui"
	"github.com/wwsheng009/mint/ui/components/input"
)

// =============================================================================
// 类型安全 StateKey 定义 (MVP 模式)
// =============================================================================

var (
	// 表单字段键
	usernameKey        = intent.StateKey[string]("username")
	usernameSetterKey  = intent.StateKey[func(string)]("usernameSetter")
	emailKey          = intent.StateKey[string]("email")
	emailSetterKey    = intent.StateKey[func(string)]("emailSetter")
	passwordKey       = intent.StateKey[string]("password")
	passwordSetterKey = intent.StateKey[func(string)]("passwordSetter")
	ageKey            = intent.StateKey[string]("age")
	ageSetterKey      = intent.StateKey[func(string)]("ageSetter")

	// 状态字段键
	agreedKey        = intent.StateKey[bool]("agreed")
	agreedSetterKey  = intent.StateKey[func(bool)]("agreedSetter")
	stepKey          = intent.StateKey[int]("step")
	stepSetterKey    = intent.StateKey[func(int)]("stepSetter")
	showModalKey     = intent.StateKey[bool]("showModal")
	showModalSetterKey = intent.StateKey[func(bool)]("showModalSetter")
)

// =============================================================================
// 自定义 Intent 类型（非字段变更的 Action Intent）
// =============================================================================

// UpdateStepIntent 更新步骤 (非表单字段，仍是自定义 Intent)
type UpdateStepIntent struct {
	Step int
}

func (UpdateStepIntent) IntentType() string { return "UpdateStep" }
func (UpdateStepIntent) StayPressed() bool  { return true }

// ShowModalIntent 显示 Modal
type ShowModalIntent struct{}

func (ShowModalIntent) IntentType() string { return "ShowModal" }
func (ShowModalIntent) StayPressed() bool  { return true }

// QuitIntent 退出应用
type QuitIntent struct{}

func (QuitIntent) IntentType() string { return "Quit" }
func (QuitIntent) StayPressed() bool  { return false }

// CloseModalIntent 关闭 Modal
type CloseModalIntent struct{}

func (CloseModalIntent) IntentType() string { return "CloseModal" }
func (CloseModalIntent) StayPressed() bool  { return true }

// =============================================================================
// 主函数
// =============================================================================

func main() {
	// 设置主题为 Ant Design 推荐的配色
	_ = theme.SetTheme("nord")

	err := ui.Run(App,
		ui.WithWidth(80),
		ui.WithHeight(30),
		ui.WithTitle("Mint TUI - Ant Design Style (MVP)"),
		ui.WithInit(func() {
			// 注册统一的 FieldChangeIntent Handler
			ui.RegisterIntent(func(ctx *intent.ActionContext, i intent.FieldChangeIntent) intent.IntentResult {
				field := i.Field
				value := i.Value

				// 字符串字段
				switch field {
				case usernameKey.String():
					setter, _ := ctx.GetState(usernameSetterKey.String())
					callSetter(setter, value)
				case emailKey.String():
					setter, _ := ctx.GetState(emailSetterKey.String())
					callSetter(setter, value)
				case passwordKey.String():
					setter, _ := ctx.GetState(passwordSetterKey.String())
					callSetter(setter, value)
				case ageKey.String():
					setter, _ := ctx.GetState(ageSetterKey.String())
					callSetter(setter, value)
				}

				return intent.HandledResult()
			})

			// 注册其他自定义 Intent Handler
			ui.RegisterIntent(func(ctx *intent.ActionContext, i UpdateStepIntent) intent.IntentResult {
				setter, _ := ctx.GetState(stepSetterKey.String())
				callSetter(setter, i.Step)
				log.TempLogger.Debug("Step updated to: %d", i.Step)
				return intent.HandledResult()
			})

			ui.RegisterIntent(func(ctx *intent.ActionContext, i ShowModalIntent) intent.IntentResult {
				setter, _ := ctx.GetState(showModalSetterKey.String())
				callSetter(setter, true)
				return intent.HandledResult()
			})

			ui.RegisterIntent(func(ctx *intent.ActionContext, i CloseModalIntent) intent.IntentResult {
				setter, _ := ctx.GetState(showModalSetterKey.String())
				callSetter(setter, false)
				return intent.HandledResult()
			})

			ui.RegisterIntent(func(ctx *intent.ActionContext, i QuitIntent) intent.IntentResult {
				ui.Quit()
				return intent.HandledResult()
			})
		}),
	)
	if err != nil {
		panic(err)
	}
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

// =============================================================================
// 主应用组件
// =============================================================================

func App() ui.VNode {
	// 使用 UseState 获取状态
	username, setUsername := ui.UseStateString("")
	email, setEmail := ui.UseStateString("")
	password, setPassword := ui.UseStateString("")
	age, setAge := ui.UseStateString("")
	step, setStep, _ := ui.UseStateInt(1)
	agreed, setAgreed := ui.UseStateBool(false)
	showModal, setShowModal := ui.UseStateBool(false)

	// 保存 setters 到 State 供 Intent Handler 使用
	ctx := ui.GetCurrentContext()
	if ctx != nil {
		ctx.GlobalState[usernameSetterKey.String()] = setUsername
		ctx.GlobalState[emailSetterKey.String()] = setEmail
		ctx.GlobalState[passwordSetterKey.String()] = setPassword
		ctx.GlobalState[ageSetterKey.String()] = setAge
		ctx.GlobalState[stepSetterKey.String()] = setStep
		ctx.GlobalState[agreedSetterKey.String()] = setAgreed
		ctx.GlobalState[showModalSetterKey.String()] = setShowModal
	}

	// 如果应该显示 Modal，返回 Modal 视图
	if showModal {
		return ModalView()
	}

	// 否则返回主表单视图
	return ui.VStackBuilder(
		Header(),
		StepIndicator(step),
		ProgressBar(step),
		FormContent(username, email, password, age, agreed, step),
		ActionButtons(step),
		Footer(),
	).Gap(1).Build()
}

// ModalView - 成功提交后的 Modal 视图
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
			app.ButtonBuilder("[ Close ]").
				Variant(app.ButtonVariantPrimary).
				OnPress(CloseModalIntent{}). // 使用 CloseModalIntent 关闭 Modal
				Build(),
			ui.Text(""),
		),
	).Gap(1).Build()
}

// =============================================================================
// UI 组件
// =============================================================================

// Header - 页面头部
func Header() ui.VNode {
	return ui.Bordered().
		Color(string(theme.Primary())).
		Child(
			ui.HStackBuilder(
				ui.NewTextBuilder("📝 User Registration Form").
					Style(style.Style{}.
						Foreground(theme.BG()).
						Background(theme.Primary()).
						Bold(true)).
					Build(),
			).Align(ui.AlignCenter).Build(),
		).
		Build()
}

// StepIndicator - 步骤指示器（Ant Design Steps 组件）
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

// ProgressBar - 进度条（Ant Design Progress 组件）
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

// FormContent - 表单内容（根据步骤变化）
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
		return ui.Bordered().
			Child(
				ui.VStackBuilder(
					FormItem(username, "Enter your username", 24, usernameKey, "", true),
					FormItem(email, "example@domain.com", 24, emailKey, "We'll never share your email", true),
					FormItemPassword(password, "Enter your password", 24, passwordKey, "At least 8 characters"),
				).Gap(2).Build(),
			).
			Build()
	} else if step == 2 {
		// Step 2: Profile
		return ui.Bordered().
			Child(
				ui.VStackBuilder(
					FormItem(age, "Your age", 10, ageKey, "", true),
				).Gap(1).Build(),
			).
			Build()
	} else {
		// Step 3: Confirm
		return ui.Bordered().
			Child(
				ui.VStackBuilder(
					ConfirmInfo("Username:", username),
					ConfirmInfo("Email:", email),
					ConfirmInfo("Age:", age),
					ui.HStackBuilder(
						ui.Text("         "),
						app.CheckboxBuilder().
							ForField(intent.ForField(agreedKey)).
							Checked(agreed).
							Label("I agree to the Terms and Conditions").
							Build(),
					).Build(),
				).Gap(2).Build(),
			).
			Build()
	}
}

// FormItem - Ant Design 风格的表单项（MVP 模式）
func FormItem(
	value string,
	placeholder string,
	width int,
	fieldKey intent.StateKey[string],
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

	// 提取字段标签（从 Key 中获取）
	label := strings.Replace(fieldKey.String(), "username", "Username:", 1)
	label = strings.Replace(label, "email", "Email:", 1)
	label = strings.Replace(label, "password", "Password:", 1)
	label = strings.Replace(label, "age", "Age:", 1)

	return ui.VStackBuilder(
		ui.HStackBuilder(
			ui.NewTextBuilder(fmt.Sprintf("%-*s", labelWidth, label)).
				Style(style.Style{}.
					Foreground(theme.Text()).
					Bold(true)).
				Build(),
			ui.Text(" "),
			app.InputBuilder().
				ForField(intent.ForField(fieldKey)).
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

// FormItemPassword - 密码输入框表单项（MVP 模式）
func FormItemPassword(
	value string,
	placeholder string,
	width int,
	fieldKey intent.StateKey[string],
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
			app.InputBuilder().
				ForField(intent.ForField(fieldKey)).
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

// ConfirmInfo - 确认页面信息显示
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

// ActionButtons - 操作按钮组
func ActionButtons(step int) ui.VNode {
	const totalSteps = 3
	var buttons []ui.VNode

	if step > 1 {
		buttons = append(buttons,
			app.ButtonBuilder("[ Previous ]").
				Variant(app.ButtonVariantSecondary).
				OnPress(UpdateStepIntent{Step: step - 1}).
				Build(),
		)
	}

	if step < totalSteps {
		buttons = append(buttons,
			app.ButtonBuilder("[ Next ]").
				Variant(app.ButtonVariantPrimary).
				OnPress(UpdateStepIntent{Step: step + 1}).
				Build(),
		)
	} else {
		buttons = append(buttons,
			app.ButtonBuilder("[ Submit ]").
				Variant(app.ButtonVariantPrimary).
				OnPress(ShowModalIntent{}).
				Build(),
		)
	}

	buttons = append(buttons,
		app.ButtonBuilder("[ Cancel ]").
			Variant(app.ButtonVariantDefault).
			OnPress(QuitIntent{}).
			Build(),
	)

	return ui.HStackBuilder(buttons...).Gap(2).Build()
}

// Footer - 页脚
func Footer() ui.VNode {
	return ui.HStackBuilder(
		ui.NewTextBuilder("Press Tab to navigate • Enter to select • Esc to quit").
			Style(style.Style{}.Foreground(theme.Placeholder())).
			Build(),
	).Align(ui.AlignCenter).Build()
}
