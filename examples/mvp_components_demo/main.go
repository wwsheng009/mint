// MVP Components Demo
//
// 采用【模式3：自定义 Intent】+ FieldChangeIntent + 类型安全 StateKey
// 展示所有核心表单组件的 ForField() + FieldChangeIntent 模式：
//   - Input: 用户名、电子邮件
//   - Textarea: 个人简介
//   - Select: 国家选择
//   - Checkbox: 同意条款
//
// MVP 数据流：
//   UI.Instance (缓冲) → FieldChangeIntent → State (事实源) → VNode → UI.Instance (渲染同步)
//
// 三种 Intent 管理模式：
//   1. 组件级状态 - ui.On + UseState + Simple* Intent（推荐组件内状态）
//   2. 全局状态 - runtime/intent 内置函数
//   3. 自定义 Intent + ui.On（本示例）
//
// 运行: go run main.go

package main

import (
	"fmt"
	"strconv"

	"github.com/wwsheng009/mint/app"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/ui"
	selectcomp "github.com/wwsheng009/mint/ui/components/select"
)

func main() {
	err := ui.Run(App,
		ui.WithWidth(70),
		ui.WithHeight(35),
		ui.WithTitle("MVP Components Demo - FieldChangeIntent"),
		ui.WithInit(func() {
			// 注册 FieldChangeIntent handler - 统一处理所有字段变更
			// 在 WithInit 中注册，通过 GlobalState 动态获取 setter
			ui.RegisterIntent(func(ctx *intent.ActionContext, i intent.FieldChangeIntent) intent.IntentResult {
				field := i.Field
				value := i.Value

				switch field {
				case usernameKey.String():
					val, _ := ctx.GetState(usernameSetterKey.String())
					if fn, ok := val.(func(string)); ok {
						fn(value)
					}
				case emailKey.String():
					val, _ := ctx.GetState(emailSetterKey.String())
					if fn, ok := val.(func(string)); ok {
						fn(value)
					}
				case bioKey.String():
					val, _ := ctx.GetState(bioSetterKey.String())
					if fn, ok := val.(func(string)); ok {
						fn(value)
					}
				case countryKey.String():
					// Select 的 value 是索引字符串，转换为 int
					val, _ := ctx.GetState(countrySetterKey.String())
					if fn, ok := val.(func(int)); ok {
						if idx, err := strconv.Atoi(value); err == nil {
							fn(idx)
						}
					}
				case agreeKey.String():
					val, _ := ctx.GetState(agreeSetterKey.String())
					if fn, ok := val.(func(bool)); ok {
						agreeVal := value == "true"
						fn(agreeVal)
					}
				}
				return intent.HandledResult()
			})

			// 注册 Reset 意图 - 重置所有状态
			ui.RegisterIntent(func(ctx *intent.ActionContext, i ResetIntent) intent.IntentResult {
				val, _ := ctx.GetState(usernameSetterKey.String())
				if fn, ok := val.(func(string)); ok {
					fn("")
				}
				val, _ = ctx.GetState(emailSetterKey.String())
				if fn, ok := val.(func(string)); ok {
					fn("")
				}
				val, _ = ctx.GetState(bioSetterKey.String())
				if fn, ok := val.(func(string)); ok {
					fn("")
				}
				val, _ = ctx.GetState(countrySetterKey.String())
				if fn, ok := val.(func(int)); ok {
					fn(0)
				}
				val, _ = ctx.GetState(agreeSetterKey.String())
				if fn, ok := val.(func(bool)); ok {
					fn(false)
				}
				return intent.HandledResult()
			})

			// 注册 Submit 意图
			ui.RegisterIntent(func(ctx *intent.ActionContext, i SubmitFormIntent) intent.IntentResult {
				val, _ := ctx.GetState(submittedSetterKey.String())
				if fn, ok := val.(func(bool)); ok {
					fn(true)
				}
				return intent.HandledResult()
			})

			// 注册 Back 意图
			ui.RegisterIntent(func(ctx *intent.ActionContext, i BackFormIntent) intent.IntentResult {
				val, _ := ctx.GetState(submittedSetterKey.String())
				if fn, ok := val.(func(bool)); ok {
					fn(false)
				}
				return intent.HandledResult()
			})
		}),
	)
	if err != nil {
		panic(err)
	}
}

// =============================================================================
// 类型安全 StateKey 定义
// =============================================================================

var (
	usernameKey        = intent.StateKey[string](usernameField)
	usernameSetterKey  = intent.StateKey[func(string)](usernameField + "Setter")
	emailKey          = intent.StateKey[string](emailField)
	emailSetterKey    = intent.StateKey[func(string)](emailField + "Setter")
	bioKey            = intent.StateKey[string](bioField)
	bioSetterKey      = intent.StateKey[func(string)](bioField + "Setter")
	countryKey        = intent.StateKey[int](countryField)
	countrySetterKey  = intent.StateKey[func(int)](countryField + "Setter")
	agreeKey          = intent.StateKey[bool](agreeField)
	agreeSetterKey    = intent.StateKey[func(bool)](agreeField + "Setter")
	submittedKey      = intent.StateKey[bool](submittedField)
	submittedSetterKey= intent.StateKey[func(bool)](submittedField + "Setter")
)

const (
	usernameField  = "username"
	emailField     = "email"
	bioField       = "bio"
	countryField   = "country"
	agreeField     = "agree"
	submittedField = "submitted"
)

// =============================================================================
// 自定义 Intent 类型
// =============================================================================

type ResetIntent struct{}

func (ResetIntent) IntentType() string { return "Reset" }

type SubmitFormIntent struct{}

func (SubmitFormIntent) IntentType() string { return "SubmitForm" }

type BackFormIntent struct{}

func (BackFormIntent) IntentType() string { return "BackForm" }

// =============================================================================
// 主应用组件
// =============================================================================

func App() ui.VNode {
	// 使用 UseState 获取状态和 setter
	username, setUsername := ui.UseStateString("")
	email, setEmail := ui.UseStateString("")
	bio, setBio := ui.UseStateString("")
	country, setCountry, _ := ui.UseStateInt(0)
	agree, setAgree := ui.UseStateBool(false)
	submitted, setSubmitted := ui.UseStateBool(false)

	// 保存 setters 到 State 供 Intent Handler 使用
	ctx := ui.GetCurrentContext()
	if ctx != nil {
		ctx.GlobalState[usernameSetterKey.String()] = setUsername
		ctx.GlobalState[emailSetterKey.String()] = setEmail
		ctx.GlobalState[bioSetterKey.String()] = setBio
		ctx.GlobalState[countrySetterKey.String()] = setCountry
		ctx.GlobalState[agreeSetterKey.String()] = setAgree
		ctx.GlobalState[submittedSetterKey.String()] = setSubmitted
	}

	// 如果已提交，显示成功视图
	if submitted {
		return SuccessView(username, email, bio, country, agree)
	}

	// 否则显示表单视图
	return FormView(username, email, bio, country, agree)
}

// FormView - 主表单视图
func FormView(username, email, bio string, country int, agree bool) ui.VNode {
	return ui.VStack(
		app.NewTextBuilder("🎨 MVP Components Demo").
			Bold(true).
			FgColor("cyan").
			Build(),
		app.Text(""),
		app.NewTextBuilder("ForField() + FieldChangeIntent Pattern").
			FgColor("gray").
			Build(),
		app.NewTextBuilder("数据流: Instance → FieldChangeIntent → State → VNode").
			FgColor("gray").
			Build(),
		app.Text(""),
		app.NewTextBuilder("─").FgColor("gray").Build(),
		app.Text(""),

		// 表单内容
		BasicFormFields(username, email, agree),
		app.NewTextBuilder("─").FgColor("gray").Build(),
		app.Text(""),
		ProfileFormFields(bio, country),

		app.Text(""),
		app.NewTextBuilder("─").FgColor("gray").Build(),
		app.Text(""),

		// 底部按钮
		app.HStack(
			app.Text("  "),
			app.ButtonBuilder("  Submit  ").
				Variant(app.ButtonVariantPrimary).
				OnPress(SubmitFormIntent{}).
				Disabled(username == "" || email == "" || !agree).
				Build(),
			app.Text(" "),
			app.ButtonBuilder("  Reset  ").
				Variant(app.ButtonVariantSecondary).
				OnPress(ResetIntent{}).
				Build(),
		),
	)
}

// BasicFormFields - 基本信息字段 (Input + Checkbox)
func BasicFormFields(username, email string, agree bool) ui.VNode {
	return ui.VStack(
		// Input 组件 - 用户名
		app.NewTextBuilder("Username:").FgColor("blue").Build(),
		app.HStack(
			app.Text("  "),
			app.InputBuilder().
				// ForField() 自动处理 FieldChangeIntent
				ForField(intent.ForField(usernameKey)).
				Value(username).
				Placeholder("Enter username").
				Width(45).
				Build(),
		),

		app.Text(""),

		// Input 组件 - 电子邮件
		app.NewTextBuilder("Email:").FgColor("blue").Build(),
		app.HStack(
			app.Text("  "),
			app.InputBuilder().
				ForField(intent.ForField(emailKey)).
				Value(email).
				Placeholder("Enter email").
				Width(45).
				Build(),
		),

		app.Text(""),

		// Checkbox 组件 - 同意条款
		app.HStack(
			app.Text("  "),
			app.CheckboxBuilder().
				ForField(intent.ForField(agreeKey)).
				Checked(agree).
				Label("I agree to the terms and conditions").
				Build(),
		),

		app.Text(""),
		app.NewTextBuilder(fmt.Sprintf("✓ State: username=%q, email=%q, agree=%v",
			username, email, agree)).FgColor("gray").Build(),
	)
}

// ProfileFormFields - 个人资料字段 (Select + Textarea)
func ProfileFormFields(bio string, country int) ui.VNode {
	countries := []selectcomp.Option{
		{Value: "us", Label: "United States"},
		{Value: "cn", Label: "China"},
		{Value: "jp", Label: "Japan"},
		{Value: "uk", Label: "United Kingdom"},
		{Value: "de", Label: "Germany"},
	}

	var countryLabel string
	if country >= 0 && country < len(countries) {
		countryLabel = countries[country].Label
	} else {
		countryLabel = "Select a country"
	}

	return ui.VStack(
		// Select 组件 - 国家选择
		app.NewTextBuilder("Country:").FgColor("blue").Build(),
		app.HStack(
			app.Text("  "),
			app.SelectBuilder().
				Options(countries).
				Selected(country).
				// ForField() 会将选中的索引存储到 State
				ForField(intent.ForField(countryKey)).
				Width(45).
				Build(),
		),

		app.Text(""),

		// Textarea 组件 - 个人简介
		app.NewTextBuilder("Bio:").FgColor("blue").Build(),
		app.HStack(
			app.Text("  "),
			app.TextareaBuilder().
				ForField(intent.ForField(bioKey)).
				Value(bio).
				Placeholder("Tell us about yourself...").
				Rows(5).
				Cols(45).
				Build(),
		),

		app.Text(""),
		app.NewTextBuilder(fmt.Sprintf("✓ State: country=%s (%d), bio chars=%d",
			countryLabel, country, len(bio))).FgColor("gray").Build(),
	)
}

// SuccessView - 成功提交视图
func SuccessView(username, email, bio string, country int, agree bool) ui.VNode {
	countries := []selectcomp.Option{
		{Value: "us", Label: "United States"},
		{Value: "cn", Label: "China"},
		{Value: "jp", Label: "Japan"},
		{Value: "uk", Label: "United Kingdom"},
		{Value: "de", Label: "Germany"},
	}

	var countryLabel string
	if country >= 0 && country < len(countries) {
		countryLabel = countries[country].Label
	} else {
		countryLabel = "None"
	}

	return ui.VStack(
		app.NewTextBuilder("✅ Form Submitted Successfully!").
			Bold(true).
			FgColor("green").
			Build(),
		app.Text(""),
		app.NewTextBuilder("─").FgColor("gray").Build(),
		app.Text(""),

		app.NewTextBuilder(fmt.Sprintf("Username: %s", username)).Build(),
		app.NewTextBuilder(fmt.Sprintf("Email: %s", email)).Build(),
		app.NewTextBuilder(fmt.Sprintf("Country: %s", countryLabel)).Build(),
		app.NewTextBuilder(fmt.Sprintf("Bio: %s", bio)).Build(),
		app.NewTextBuilder(fmt.Sprintf("Agreed: %v", agree)).Build(),

		app.Text(""),
		app.NewTextBuilder("─").FgColor("gray").Build(),
		app.Text(""),

		app.HStack(
			app.Text("  "),
			app.ButtonBuilder("  Back to Form  ").
				Variant(app.ButtonVariantSecondary).
				OnPress(BackFormIntent{}).
				Build(),
			app.Text(" "),
			app.ButtonBuilder("  Reset  ").
				Variant(app.ButtonVariantDanger).
				OnPress(ResetIntent{}).
				Build(),
		),
	)
}
