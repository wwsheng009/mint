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
	"github.com/wwsheng009/mint/components/form"
	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/ui"
)

func main() {
	// 设置主题为 Ant Design 推荐的配色
	_ = theme.SetTheme("nord")

	err := ui.Run(App,
		ui.WithWidth(80),
		ui.WithHeight(25),
		ui.WithTitle("Mint TUI - Ant Design Style"),
	)
	if err != nil {
		panic(err)
	}
}

// App - 主应用组件
func App() ui.VNode {
	// 状态管理
	formData, setFormData := ui.UseStatePtr(&FormData{
		Username: "",
		Email:    "",
		Password: "",
		Age:      "",
		Agreed:   false,
	})

	showModal, setShowModal := ui.UseStateBool(false)
	step, setStep := ui.UseStateInt(1)

	return ui.VStack(
		Header(),
		ui.Text(""),
		StepIndicator(step),
		ui.Text(""),
		ProgressBar(step),
		ui.Text(""),
		FormContent(formData, setFormData, step, setStep, setShowModal),
		ui.Text(""),
		ActionButtons(step, setStep, setShowModal),
		ui.Text(""),
		Footer(),
	)
}

// FormData - 表单数据结构
type FormData struct {
	Username string
	Email    string
	Password string
	Age      string
	Agreed   bool
}

// Header - 页面头部
func Header() ui.VNode {
	return ui.Bordered().
		Style(string(theme.Primary())).
		Child(
			ui.HStackBuilder(
				app.NewTextBuilder("📝 User Registration Form").
					Style(style.Style{}.
						Foreground(theme.BG()).  // Ant Design: 主色背景用 BG 作为文字
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

	return ui.HStack(items...).Gap(4).Build()
}

// ProgressBar - 进度条（Ant Design Progress 组件）
func ProgressBar(step int) ui.VNode {
	const totalSteps = 3
	progress := step * 30 / totalSteps  // 每步 30%

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
		app.NewTextBuilder("└" + strings.Repeat("─", 30) + "┘").
			Style(style.Style{}.Foreground(theme.Border())).
			Build(),
	)
}

// FormContent - 表单内容
func FormContent(data *FormData, setData func(*FormData), step int, setStep func(int), setShowModal func(bool)) ui.VNode {

	if step == 1 {
		// Step 1: Account Information
		return ui.Bordered().
			Style(string(theme.Border())).
			Child(
				ui.VStackBuilder(
					ui.Text(""),
					// Form Item: Username
					FormItem(
						"Username:",
						"Enter your username",
						24,
						data.Username,
						func(s string) { data.Username = s; setData(data) },
						"",
						true,
					),
					ui.Text(""),
					// Form Item: Email
					FormItem(
						"Email:",
						"example@domain.com",
						24,
						data.Email,
						func(s string) { data.Email = s; setData(data) },
						"We'll never share your email",
						true,
					),
					ui.Text(""),
					// Form Item: Password
					FormItemPassword(
						"Password:",
						"Enter your password",
						24,
						data.Password,
						func(s string) { data.Password = s; setData(data) },
						"At least 8 characters",
					),
				).Gap(1).Build(),
			).
			Build()
	} else if step == 2 {
		// Step 2: Profile
		return ui.Bordered().
			Style(string(theme.Border())).
			Child(
				ui.VStackBuilder(
					ui.Text(""),
					FormItem(
						"Age:",
						"Your age",
						10,
						data.Age,
						func(s string) { data.Age = s; setData(data) },
						"",
						true,
					),
					ui.Text(""),
					ui.Text(""),
				).Gap(1).Build(),
			).
			Build()
	} else {
		// Step 3: Confirm
		return ui.Bordered().
			Style(string(theme.Border())).
			Child(
				ui.VStackBuilder(
					ui.Text(""),
					ConfirmInfo("Username:", data.Username),
					ConfirmInfo("Email:", data.Email),
					ConfirmInfo("Age:", data.Age),
					ui.Text(""),
					ui.HStackBuilder(
						ui.Text("").Width(11).Build(),
						app.CheckboxBuilder().
							Checked(data.Agreed).
							Label("I agree to the Terms and Conditions").
							OnChange(func(checked bool) {
								data.Agreed = checked
								setData(data)
							}).
							Build(),
					).Build(),
				).Gap(1).Build(),
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
	onChange func(string),
	helpText string,
	required bool,
) ui.VNode {
	labelWidth := 10

	return ui.VStackBuilder(
		// Label 行
		ui.HStackBuilder(
			app.NewTextBuilder(label).
				Style(style.Style{}.
					Foreground(theme.Text()).  // Ant Design: Label 使用 TEXT
					Bold(true)).
				Width(labelWidth).
				Build(),
			ui.Text(" "),
			app.InputBuilder().
				Value(value).
				Placeholder(placeholder).
				Width(width).
				OnChange(onChange).
				Build(),
			// Required 标记
			app.NewTextBuilder("*").
				Style(style.Style{}.Foreground(theme.Error())).
				BuildCondition(required),
		).Build(),
		// Help/Error 文本行
		ui.HStackBuilder(
			ui.Text("").Width(labelWidth + 1).Build(),
			app.NewTextBuilder(helpText).
				Style(style.Style{}.Foreground(theme.Muted())).  // Ant Design: Help 使用 MUTED
				BuildCondition(helpText != ""),
		).Build(),
	).Gap(1).Build()
}

// FormItemPassword - 密码输入框表单项
func FormItemPassword(
	label string,
	placeholder string,
	width int,
	value string,
	onChange func(string),
	helpText string,
) ui.VNode {
	labelWidth := 10

	return ui.VStackBuilder(
		ui.HStackBuilder(
			app.NewTextBuilder(label).
				Style(style.Style{}.
					Foreground(theme.Text()).
					Bold(true)).
				Width(labelWidth).
				Build(),
			ui.Text(" "),
			app.InputBuilder().
				Value(value).
				InputType(form.InputTypePassword).
				Placeholder(placeholder).
				Width(width).
				OnChange(onChange).
				Build(),
		).Build(),
		ui.HStackBuilder(
			ui.Text("").Width(labelWidth + 1).Build(),
			app.NewTextBuilder(helpText).
				Style(style.Style{}.Foreground(theme.Muted())).
				BuildCondition(helpText != ""),
		).Build(),
	).Gap(1).Build()
}

// ConfirmInfo - 确认页面信息显示
func ConfirmInfo(label, value string) ui.VNode {
	labelWidth := 10

	return ui.HStackBuilder(
		app.NewTextBuilder(label).
			Style(style.Style{}.
				Foreground(theme.Muted()).  // Label 使用 MUTED
				Bold(true)).
			Width(labelWidth).
			Build(),
		app.NewTextBuilder(value).
			Style(style.Style{}.
				Foreground(theme.Text())).  // Value 使用 TEXT
			Build(),
	).Build()
}

// ActionButtons - 操作按钮组
func ActionButtons(step int, setStep func(int), setShowModal func(bool)) ui.VNode {
	const totalSteps = 3

	var buttons []ui.VNode

	// Previous 按钮
	if step > 1 {
		buttons = append(buttons,
			app.ButtonBuilder("[ Previous ]").
				Variant(app.ButtonVariantSecondary).  // Ant Design: 次要操作
				OnClick(func() {
					setStep(step - 1)
				}).
				Build(),
		)
	}

	// Next / Submit 按钮
	if step < totalSteps {
		buttons = append(buttons,
			app.ButtonBuilder("[ Next ]").
				Variant(app.ButtonVariantPrimary).  // Ant Design: 主要操作
				OnClick(func() {
					setStep(step + 1)
				}).
				Build(),
		)
	} else {
		buttons = append(buttons,
			app.ButtonBuilder("[ Submit ]").
				Variant(app.ButtonVariantPrimary).
				OnClick(func() {
					setShowModal(true)
				}).
				Build(),
		)
	}

	// Cancel 按钮
	buttons = append(buttons,
		app.ButtonBuilder("[ Cancel ]").
			Variant(app.ButtonVariantDefault).  // Ant Design: 默认按钮
			OnClick(func() {
				ui.Quit()
			}).
			Build(),
	)

	return ui.HStack(buttons...).Gap(2).Build()
}

// Footer - 页脚
func Footer() ui.VNode {
	return ui.HStackBuilder(
		app.NewTextBuilder("Press Tab to navigate • Enter to select • Esc to quit").
			Style(style.Style{}.Foreground(theme.Placeholder())).
			Build(),
	).Align(ui.AlignCenter).Build()
}
