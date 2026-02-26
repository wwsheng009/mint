// Ant Design Style Demo
//
// 这个示例展示如何在 Mint TUI 中应用 Ant Design 的设计理念
// 包括：表单布局、按钮类型、颜色使用、键盘交互等
//
// 运行: go run main.go

package main

import (
	"fmt"
	"strings"

	"github.com/wwsheng009/mint/app"
	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/internal/log"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui"
	"github.com/wwsheng009/mint/ui/components/input"
)

// =============================================================================
// Intent Definitions
// =============================================================================

// UpdateFieldIntent 更新表单字段
type UpdateFieldIntent struct {
	Field string // 字段名: username, email, password, age
	Value string
}

func (UpdateFieldIntent) IntentType() string { return "UpdateField" }
func (UpdateFieldIntent) StayPressed() bool  { return true } // 保持按下视觉反馈

// UpdateStepIntent 更新步骤
type UpdateStepIntent struct {
	Step int
}

func (UpdateStepIntent) IntentType() string { return "UpdateStep" }
func (UpdateStepIntent) StayPressed() bool  { return true } // 保持按下视觉反馈

// UpdateAgreedIntent 更新同意状态
type UpdateAgreedIntent struct {
	Agreed bool
}

func (UpdateAgreedIntent) IntentType() string { return "UpdateAgreed" }
func (UpdateAgreedIntent) StayPressed() bool  { return true } // 保持按下视觉反馈

// ShowModalIntent 显示 Modal
type ShowModalIntent struct{}

func (ShowModalIntent) IntentType() string { return "ShowModal" }
func (ShowModalIntent) StayPressed() bool  { return true } // 保持按下视觉反馈

// QuitIntent 退出应用
type QuitIntent struct{}

func (QuitIntent) IntentType() string { return "Quit" }
func (QuitIntent) StayPressed() bool  { return false } // 立即重置（应用即将退出）

func main() {
	// 设置主题为 Ant Design 推荐的配色
	_ = theme.SetTheme("nord")

	// 定义 App 并在 ui.Run 之后注册 Intent Handlers
	// 因为 Intent Runtime 在 ui.Run 内部创建
	err := ui.Run(App,
		ui.WithWidth(80),
		ui.WithHeight(30),
		ui.WithTitle("Mint TUI - Ant Design Style"),
		ui.WithInit(func() {
			// 注册 Intent Handlers
			ui.RegisterIntent(func(ctx *intent.ActionContext, i UpdateFieldIntent) intent.IntentResult {
				ctx.SetState(i.Field, i.Value)
				return intent.HandledResult()
			})

			ui.RegisterIntent(func(ctx *intent.ActionContext, i UpdateStepIntent) intent.IntentResult {
				ctx.SetState("step", i.Step)
				log.TempLogger.Debug("ctx.SetState:%d", i.Step)
				return intent.HandledResult()
			})

			ui.RegisterIntent(func(ctx *intent.ActionContext, i UpdateAgreedIntent) intent.IntentResult {
				ctx.SetState("agreed", i.Agreed)
				return intent.HandledResult()
			})

			ui.RegisterIntent(func(ctx *intent.ActionContext, i ShowModalIntent) intent.IntentResult {
				// Modal 显示逻辑（省略实现）
				ctx.SetState("showModal", true)
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

// App - 主应用组件
func App() ui.VNode {
	// 状态管理 - 使用 ComponentContext 的状态管理（非闭包）
	ctx := rtui.GetCurrentContext()

	// 从 State 中读取状态
	username := ctx.GetStringState("username", "")
	email := ctx.GetStringState("email", "")
	password := ctx.GetStringState("password", "")
	age := ctx.GetStringState("age", "")
	step := ctx.GetIntState("step", 1)
	agreed := ctx.GetBoolState("agreed", false)
	log.TempLogger.Debug("ctx.GetIntState:%d", step)
	return ui.VStackBuilder(
		Header(),
		StepIndicator(step),
		ProgressBar(step),
		FormContent(username, email, password, age, agreed, step),
		ActionButtons(step),
		Footer(),
	).Gap(1).Build()
}

// FormData - 表单数据结构（仅用于组织数据，状态由Hook单独管理）
type FormData struct {
	Username string
	Email    string
	Password string
	Age      string
	Agreed   bool
}

// getFormData 从各个状态组合成FormData
func getFormData(username, email, password, age string, agreed bool) *FormData {
	return &FormData{
		Username: username,
		Email:    email,
		Password: password,
		Age:      age,
		Agreed:   agreed,
	}
}

// Header - 页面头部
func Header() ui.VNode {
	return ui.Bordered().
		Color(string(theme.Primary())).
		Child(
			ui.HStackBuilder(
				app.NewTextBuilder("📝 User Registration Form").
					Style(style.Style{}.
						Foreground(theme.BG()). // Ant Design: 主色背景用 BG 作为文字
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
			// 已完成 - Success 绿色
			itemStyle = style.Style{}.
				Foreground(theme.Success()).
				Bold(true)
		} else if i == step {
			// 当前 - Primary 蓝色
			itemStyle = style.Style{}.
				Foreground(theme.Primary()).
				Bold(true)
		} else {
			// 未完成 - Muted 灰色
			itemStyle = style.Style{}.
				Foreground(theme.Muted())
		}

		items[i] = ui.HStackBuilder(
			app.NewTextBuilder(fmt.Sprintf("%d. %s", i+1, s)).
				Style(itemStyle).
				Build(),
		).Align(ui.AlignCenter).Build()
	}

	return ui.HStackBuilder(items...).Gap(4).Build()
}

// ProgressBar - 进度条（Ant Design Progress 组件）
func ProgressBar(step int) ui.VNode {
	const totalSteps = 3
	progress := step * 30 / totalSteps // 每步 30%

	// 轨道
	track := app.NewTextBuilder("┌" + strings.Repeat("─", 30) + "┐").
		Style(style.Style{}.Foreground(theme.Border())).
		Build()

	// 进度填充
	fill := app.NewTextBuilder("│" + strings.Repeat("━", progress) + strings.Repeat("─", 30-progress) + "│").
		Style(style.Style{}.Foreground(theme.Primary())).
		Build()

	return ui.VStack(
		track,
		fill,
		app.NewTextBuilder("└"+strings.Repeat("─", 30)+"┘").
			Style(style.Style{}.Foreground(theme.Border())).
			Build(),
	)
}

// FormContent - 表单内容
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
					// Form Item: Username
					FormItem(
						"Username:",
						"Enter your username",
						24,
						username,
						"username", // field name
						"",
						true,
					),
					// Form Item: Email
					FormItem(
						"Email:",
						"example@domain.com",
						24,
						email,
						"email", // field name
						"We'll never share your email",
						true,
					),
					// Form Item: Password
					FormItemPassword(
						"Password:",
						"Enter your password",
						24,
						password,
						"password", // field name
						"At least 8 characters",
					),
				).Gap(2).Build(),
			).
			Build()
	} else if step == 2 {
		// Step 2: Profile
		return ui.Bordered().
			Child(
				ui.VStackBuilder(
					FormItem(
						"Age:",
						"Your age",
						10,
						age,
						"age", // field name
						"",
						true,
					),
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
						app.CheckboxBuilder().
							Checked(agreed).
							Label("I agree to the Terms and Conditions").
							// 使用 Intent 模式
							OnToggle(UpdateAgreedIntent{}).
							Build(),
					).Build(),
				).Gap(2).Build(),
			).
			Build()
	}
}

// FormItem - Ant Design 风格的表单项
func FormItem(
	label string,
	placeholder string,
	width int,
	value string,
	field string, // 字段名，用于创建 Intent
	helpText string,
	required bool,
) ui.VNode {
	labelWidth := 10

	// Required 标记
	var requiredMark ui.VNode
	if required {
		requiredMark = app.NewTextBuilder("*").
			Style(style.Style{}.Foreground(theme.Error())).
			Build()
	} else {
		requiredMark = ui.Text("")
	}

	// Help/Error 文本行
	var helpNode ui.VNode
	if helpText != "" {
		helpNode = app.NewTextBuilder(helpText).
			Style(style.Style{}.Foreground(theme.Muted())).
			Build()
	} else {
		helpNode = ui.Text("")
	}

	// 使用 Intent 模式
	changeIntent := UpdateFieldIntent{
		Field: field,
	}

	return ui.VStackBuilder(
		// Label 行
		ui.HStackBuilder(
			app.NewTextBuilder(fmt.Sprintf("%-*s", labelWidth, label)).
				Style(style.Style{}.
					Foreground(theme.Text()). // Ant Design: Label 使用 TEXT
					Bold(true)).
				Build(),
			ui.Text(" "),
			app.InputBuilder().
				Value(value).
				Placeholder(placeholder).
				Type(input.TypeText).
				OnChange(changeIntent).
				Build(),
			requiredMark,
		).Build(),
		// Help/Error 文本行
		ui.HStackBuilder(
			ui.Text(strings.Repeat(" ", labelWidth+1)),
			helpNode,
		).Build(),
	).Gap(1).Build()
}

// FormItemPassword - 密码输入框表单项
func FormItemPassword(
	label string,
	placeholder string,
	width int,
	value string,
	field string, // 字段名，用于创建 Intent
	helpText string,
) ui.VNode {
	labelWidth := 10

	// Help/Error 文本行
	var helpNode ui.VNode
	if helpText != "" {
		helpNode = app.NewTextBuilder(helpText).
			Style(style.Style{}.Foreground(theme.Muted())).
			Build()
	} else {
		helpNode = ui.Text("")
	}

	// 使用 Intent 模式
	changeIntent := UpdateFieldIntent{
		Field: field,
	}

	return ui.VStackBuilder(
		ui.HStackBuilder(
			app.NewTextBuilder(fmt.Sprintf("%-*s", labelWidth, label)).
				Style(style.Style{}.
					Foreground(theme.Text()).
					Bold(true)).
				Build(),
			ui.Text(" "),
			app.InputBuilder().
				Value(value).
				Password().
				Placeholder(placeholder).
				OnChange(changeIntent).
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
		app.NewTextBuilder(fmt.Sprintf("%-*s", labelWidth, label)).
			Style(style.Style{}.
				Foreground(theme.Muted()). // Label 使用 MUTED
				Bold(true)).
			Build(),
		app.NewTextBuilder(value).
			Style(style.Style{}.
				Foreground(theme.Text())). // Value 使用 TEXT
			Build(),
	).Build()
}

// ActionButtons - 操作按钮组
func ActionButtons(step int) ui.VNode {
	const totalSteps = 3
	log.TempLogger.Debug("ActionButtons Called step:%d", step)
	var buttons []ui.VNode

	// Previous 按钮 - 使用 Intent
	if step > 1 {
		buttons = append(buttons,
			app.ButtonBuilder("[ Previous ]").
				Variant(app.ButtonVariantSecondary). // Ant Design: 次要操作
				// 使用 Intent 模式
				OnPress(UpdateStepIntent{
					Step: step - 1,
				}).
				Build(),
		)
	}

	// Next / Submit 按钮 - 使用 Intent
	if step < totalSteps {
		buttons = append(buttons,
			app.ButtonBuilder("[ Next ]").
				Variant(app.ButtonVariantPrimary). // Ant Design: 主要操作
				// 使用 Intent 模式
				OnPress(UpdateStepIntent{
					Step: step + 1,
				}).
				Build(),
		)
	} else {
		buttons = append(buttons,
			app.ButtonBuilder("[ Submit ]").
				Variant(app.ButtonVariantPrimary).
				// 使用 Intent 模式
				OnPress(ShowModalIntent{}).
				Build(),
		)
	}

	// Cancel 按钮 - 使用 Intent
	buttons = append(buttons,
		app.ButtonBuilder("[ Cancel ]").
			Variant(app.ButtonVariantDefault). // Ant Design: 默认按钮
			// 使用 Intent 模式
			OnPress(QuitIntent{}).
			Build(),
	)

	return ui.HStackBuilder(buttons...).Gap(2).Build()
}

// Footer - 页脚
func Footer() ui.VNode {
	return ui.HStackBuilder(
		app.NewTextBuilder("Press Tab to navigate • Enter to select • Esc to quit").
			Style(style.Style{}.Foreground(theme.Placeholder())).
			Build(),
	).Align(ui.AlignCenter).Build()
}
