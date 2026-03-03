// Validation Demo
//
// 演示如何将 validation 包与 Input 组件结合使用
// 展示表单验证的完整流程：
//   1. 定义验证规则
//   2. 输入时实时验证
//   3. 提交时全面验证
//   4. 显示验证错误信息
//
// 运行: go run ./examples/validation_demo/

package main

import (
	"fmt"
	"strconv"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/ui"
	"github.com/wwsheng009/mint/ui/components/validation"
)

// =============================================================================
// 自定义 Intent
// =============================================================================

// SubmitIntent 提交表单
type SubmitIntent struct{}

func (SubmitIntent) IntentType() string { return "Submit" }
func (SubmitIntent) StayPressed() bool  { return true }

// ResetIntent 重置表单
type ResetIntent struct{}

func (ResetIntent) IntentType() string { return "Reset" }
func (ResetIntent) StayPressed() bool  { return true }

// =============================================================================
// 主函数
// =============================================================================

func main() {
	err := ui.Run(FormApp,
		ui.WithWidth(70),
		ui.WithHeight(35),
		ui.WithTitle("Validation Form Demo"),
		ui.WithInit(func() {
			// 注册 FieldChangeIntent 处理器
			ui.RegisterIntent(func(ctx *intent.ActionContext, i intent.FieldChangeIntent) intent.IntentResult {
				switch i.Field {
				case "username":
					setUsername, _ := ctx.GetState("usernameSetter")
					if fn, ok := setUsername.(func(string)); ok {
						fn(i.Value)
					}
				case "email":
					setEmail, _ := ctx.GetState("emailSetter")
					if fn, ok := setEmail.(func(string)); ok {
						fn(i.Value)
					}
				case "age":
					setAge, _ := ctx.GetState("ageSetter")
					if fn, ok := setAge.(func(string)); ok {
						fn(i.Value)
					}
				case "password":
					setPassword, _ := ctx.GetState("passwordSetter")
					if fn, ok := setPassword.(func(string)); ok {
						fn(i.Value)
					}
				case "confirmPwd":
					setConfirmPwd, _ := ctx.GetState("confirmPwdSetter")
					if fn, ok := setConfirmPwd.(func(string)); ok {
						fn(i.Value)
					}
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
// 主应用组件
// =============================================================================

func FormApp() ui.VNode {
	// 表单状态
	username, setUsername := ui.UseStateString("")
	email, setEmail := ui.UseStateString("")
	age, setAge := ui.UseStateString("")
	password, setPassword := ui.UseStateString("")
	confirmPwd, setConfirmPwd := ui.UseStateString("")

	// 验证错误状态
	usernameErr, setUsernameErr := ui.UseStateString("")
	emailErr, setEmailErr := ui.UseStateString("")
	ageErr, setAgeErr := ui.UseStateString("")
	passwordErr, setPasswordErr := ui.UseStateString("")
	confirmErr, setConfirmErr := ui.UseStateString("")

	// 将 setter 保存到 GlobalState
	ctx := ui.GetCurrentContext()
	if ctx != nil {
		ctx.GlobalState["usernameSetter"] = setUsername
		ctx.GlobalState["emailSetter"] = setEmail
		ctx.GlobalState["ageSetter"] = setAge
		ctx.GlobalState["passwordSetter"] = setPassword
		ctx.GlobalState["confirmPwdSetter"] = setConfirmPwd
		ctx.GlobalState["usernameErrSetter"] = setUsernameErr
		ctx.GlobalState["emailErrSetter"] = setEmailErr
		ctx.GlobalState["ageErrSetter"] = setAgeErr
		ctx.GlobalState["passwordErrSetter"] = setPasswordErr
		ctx.GlobalState["confirmErrSetter"] = setConfirmErr
	}

	// 验证器定义
	usernameValidator := validation.NewChain().
		Required().
		MinLength(3).
		MaxLength(20).
		Build()

	emailValidator := validation.NewChain().
		Required().
		Email().
		Build()

	passwordValidator := validation.NewChain().
		Required().
		MinLength(6).
		MaxLength(20).
		Build()

	// 验证单个字段
	validateField := func(field string, value string) string {
		switch field {
		case "username":
			if err := usernameValidator.Validate(value); err != nil {
				return err.Error()
			}
		case "email":
			if err := emailValidator.Validate(value); err != nil {
				return err.Error()
			}
		case "age":
			if value != "" {
				age, err := strconv.Atoi(value)
				if err != nil {
					return "请输入有效数字"
				}
				if age < 1 || age > 150 {
					return "年龄必须在1-150之间"
				}
			}
		case "password":
			if err := passwordValidator.Validate(value); err != nil {
				return err.Error()
			}
		case "confirmPwd":
			if value != password {
				return "两次密码输入不一致"
			}
		}
		return ""
	}

	// Submit 处理 - 提交时验证所有字段
	ui.On(SubmitIntent{}, func() {
		// 验证所有字段
		setUsernameErr(validateField("username", username))
		setEmailErr(validateField("email", email))
		setAgeErr(validateField("age", age))
		setPasswordErr(validateField("password", password))
		setConfirmErr(validateField("confirmPwd", confirmPwd))
	})

	// Reset 处理
	ui.On(ResetIntent{}, func() {
		setUsername("")
		setEmail("")
		setAge("")
		setPassword("")
		setConfirmPwd("")
		setUsernameErr("")
		setEmailErr("")
		setAgeErr("")
		setPasswordErr("")
		setConfirmErr("")
	})

	// 显示错误信息
	showError := func(err string) ui.VNode {
		if err == "" {
			return ui.Text("")
		}
		return ui.NewTextBuilder(err).
			FgColor("red").
			Build()
	}

	return ui.VStack(
		// 标题
		ui.NewTextBuilder("📝 Form Validation Demo").
			Bold(true).
			FgColor("cyan").
			Build(),
		ui.Text(""),

		// 说明
		ui.NewTextBuilder("使用 validation 包进行表单验证").
			FgColor("gray").
			Build(),
		ui.Text(""),

		// 分割线
		ui.NewTextBuilder("──────────────────────────────────────────").
			FgColor("bright-black").
			Build(),
		ui.Text(""),

		// 用户名
		ui.HStack(
			ui.NewTextBuilder("Username:").
				FgColor("blue").
				Build(),
			ui.NewInputBuilder().
				ForField(intent.BindField("username")).
				Value(username).
				Placeholder("3-20字符").
				Width(30).
				Build(),
		),
		showError(usernameErr),

		// 邮箱
		ui.HStack(
			ui.NewTextBuilder("Email:").
				FgColor("blue").
				Build(),
			ui.NewInputBuilder().
				ForField(intent.BindField("email")).
				Value(email).
				Placeholder("example@mail.com").
				Width(30).
				Build(),
		),
		showError(emailErr),

		// 年龄
		ui.HStack(
			ui.NewTextBuilder("Age:").
				FgColor("blue").
				Build(),
			ui.NewInputBuilder().
				ForField(intent.BindField("age")).
				Value(age).
				Placeholder("1-150").
				Width(30).
				Build(),
		),
		showError(ageErr),

		// 密码
		ui.HStack(
			ui.NewTextBuilder("Password:").
				FgColor("blue").
				Build(),
			ui.NewInputBuilder().
				ForField(intent.BindField("password")).
				Value(password).
				Placeholder("6-20字符").
				Width(30).
				Build(),
		),
		showError(passwordErr),

		// 确认密码
		ui.HStack(
			ui.NewTextBuilder("Confirm:").
				FgColor("blue").
				Build(),
			ui.NewInputBuilder().
				ForField(intent.BindField("confirmPwd")).
				Value(confirmPwd).
				Placeholder("再次输入密码").
				Width(30).
				Build(),
		),
		showError(confirmErr),

		ui.Text(""),

		// 按钮
		ui.HStack(
			ui.Text("  "),
			ui.NewButtonBuilder(" Submit ").
				Variant(ui.ButtonVariantPrimary).
				OnPress(SubmitIntent{}).
				Build(),
			ui.Text(" "),
			ui.NewButtonBuilder(" Reset ").
				Variant(ui.ButtonVariantSecondary).
				OnPress(ResetIntent{}).
				Build(),
		),

		ui.Text(""),

		// 验证规则说明
		ui.NewTextBuilder("验证规则:").Bold(true).Build(),
		ui.NewTextBuilder("  • Username: 必填, 3-20字符").FgColor("bright-black").Build(),
		ui.NewTextBuilder("  • Email: 必填, 有效邮箱格式").FgColor("bright-black").Build(),
		ui.NewTextBuilder("  • Age: 必填, 1-150").FgColor("bright-black").Build(),
		ui.NewTextBuilder("  • Password: 必填, 6-20字符").FgColor("bright-black").Build(),
		ui.NewTextBuilder("  • Confirm: 与密码一致").FgColor("bright-black").Build(),

		ui.Text(""),

		// 当前表单数据
		ui.NewTextBuilder("──────────────────────────────────────────").
			FgColor("bright-black").
			Build(),
		ui.Text(""),
		ui.NewTextBuilder("当前输入:").Bold(true).Build(),
		ui.NewTextBuilder(fmt.Sprintf("  Username: %s", username)).Build(),
		ui.NewTextBuilder(fmt.Sprintf("  Email: %s", email)).Build(),
		ui.NewTextBuilder(fmt.Sprintf("  Age: %s", age)).Build(),
	)
}
