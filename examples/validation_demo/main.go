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
			// 将 UseState 的 setter 保存到 GlobalState，供处理器调用
			ui.RegisterIntent(func(ctx *intent.ActionContext, i intent.FieldChangeIntent) intent.IntentResult {
				switch i.Field {
				case "username":
					if fn, ok := ctx.GetState("usernameSetter"); ok {
						if setter, ok := fn.(func(string)); ok {
							setter(i.Value)
						}
					}
				case "email":
					if fn, ok := ctx.GetState("emailSetter"); ok {
						if setter, ok := fn.(func(string)); ok {
							setter(i.Value)
						}
					}
				case "age":
					if fn, ok := ctx.GetState("ageSetter"); ok {
						if setter, ok := fn.(func(string)); ok {
							setter(i.Value)
						}
					}
				case "password":
					if fn, ok := ctx.GetState("passwordSetter"); ok {
						if setter, ok := fn.(func(string)); ok {
							setter(i.Value)
						}
					}
				case "confirmPwd":
					if fn, ok := ctx.GetState("confirmPwdSetter"); ok {
						if setter, ok := fn.(func(string)); ok {
							setter(i.Value)
						}
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
	// 使用 UseState 创建表单状态 - 这是单一事实源
	username, setUsername := ui.UseStateString("")
	email, setEmail := ui.UseStateString("")
	age, setAge := ui.UseStateString("")
	password, setPassword := ui.UseStateString("")
	confirmPwd, setConfirmPwd := ui.UseStateString("")

	// ✅ 这是解决闭包问题的关键：将状态存储到 GlobalState，handler 从 ActionContext 读取
	ctx := ui.GetCurrentContext()
	if ctx != nil {
		// 保存 setter
		ctx.GlobalState["usernameSetter"] = setUsername
		ctx.GlobalState["emailSetter"] = setEmail
		ctx.GlobalState["ageSetter"] = setAge
		ctx.GlobalState["passwordSetter"] = setPassword
		ctx.GlobalState["confirmPwdSetter"] = setConfirmPwd

		// 保存当前值（供 SubmitIntent 读取最新值）
		ctx.GlobalState["username"] = username
		ctx.GlobalState["email"] = email
		ctx.GlobalState["age"] = age
		ctx.GlobalState["password"] = password
		ctx.GlobalState["confirmPwd"] = confirmPwd
	}

	// 验证错误状态 - 使用 UseState 管理
	usernameErr, setUsernameErr := ui.UseStateString("")
	emailErr, setEmailErr := ui.UseStateString("")
	ageErr, setAgeErr := ui.UseStateString("")
	passwordErr, setPasswordErr := ui.UseStateString("")
	confirmErr, setConfirmErr := ui.UseStateString("")

	// 保存错误 setter 到 GlobalState
	if ctx != nil {
		ctx.GlobalState["setUsernameErr"] = setUsernameErr
		ctx.GlobalState["setEmailErr"] = setEmailErr
		ctx.GlobalState["setAgeErr"] = setAgeErr
		ctx.GlobalState["setPasswordErr"] = setPasswordErr
		ctx.GlobalState["setConfirmErr"] = setConfirmErr
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
	validateField := func(field string, value string, pwd string) string {
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
				ageVal, err := strconv.Atoi(value)
				if err != nil {
					return "请输入有效数字"
				}
				if ageVal < 1 || ageVal > 150 {
					return "年龄必须在1-150之间"
				}
			}
		case "password":
			if err := passwordValidator.Validate(value); err != nil {
				return err.Error()
			}
		case "confirmPwd":
			if value != pwd {
				return "两次密码输入不一致"
			}
		}
		return ""
	}

	// ✅ Submit 处理 - 使用 OnWithContext 从 ActionContext 读取最新状态
	// 这解决了闭包捕获旧值的问题
	ui.On(SubmitIntent{}, func(actx *intent.ActionContext) {
		// ✅ 从 ActionContext 获取最新的表单值
		currentUsername := actx.GetStringState("username", "")
		currentEmail := actx.GetStringState("email", "")
		currentAge := actx.GetStringState("age", "")
		currentPassword := actx.GetStringState("password", "")
		currentConfirmPwd := actx.GetStringState("confirmPwd", "")

		// 从 ActionContext 获取错误 setter
		setUNameErr, _ := actx.GetState("setUsernameErr")
		setEMailErr, _ := actx.GetState("setEmailErr")
		setAgeErrFn, _ := actx.GetState("setAgeErr")
		setPwdErr, _ := actx.GetState("setPasswordErr")
		setConfErr, _ := actx.GetState("setConfirmErr")

		// 验证所有字段并设置错误
		if fn, ok := setUNameErr.(func(string)); ok {
			fn(validateField("username", currentUsername, ""))
		}
		if fn, ok := setEMailErr.(func(string)); ok {
			fn(validateField("email", currentEmail, ""))
		}
		if fn, ok := setAgeErrFn.(func(string)); ok {
			fn(validateField("age", currentAge, ""))
		}
		if fn, ok := setPwdErr.(func(string)); ok {
			fn(validateField("password", currentPassword, ""))
		}
		if fn, ok := setConfErr.(func(string)); ok {
			fn(validateField("confirmPwd", currentConfirmPwd, currentPassword))
		}
	})

	// ✅ Reset 处理 - 使用 OnWithContext
	ui.On(ResetIntent{}, func(actx *intent.ActionContext) {
		// 从 ActionContext 获取 setter
		if fn, ok := actx.GetState("usernameSetter"); ok {
			if setter, ok := fn.(func(string)); ok {
				setter("")
			}
		}
		if fn, ok := actx.GetState("emailSetter"); ok {
			if setter, ok := fn.(func(string)); ok {
				setter("")
			}
		}
		if fn, ok := actx.GetState("ageSetter"); ok {
			if setter, ok := fn.(func(string)); ok {
				setter("")
			}
		}
		if fn, ok := actx.GetState("passwordSetter"); ok {
			if setter, ok := fn.(func(string)); ok {
				setter("")
			}
		}
		if fn, ok := actx.GetState("confirmPwdSetter"); ok {
			if setter, ok := fn.(func(string)); ok {
				setter("")
			}
		}
		// 清空错误
		if fn, ok := actx.GetState("setUsernameErr"); ok {
			if setter, ok := fn.(func(string)); ok {
				setter("")
			}
		}
		if fn, ok := actx.GetState("setEmailErr"); ok {
			if setter, ok := fn.(func(string)); ok {
				setter("")
			}
		}
		if fn, ok := actx.GetState("setAgeErr"); ok {
			if setter, ok := fn.(func(string)); ok {
				setter("")
			}
		}
		if fn, ok := actx.GetState("setPasswordErr"); ok {
			if setter, ok := fn.(func(string)); ok {
				setter("")
			}
		}
		if fn, ok := actx.GetState("setConfirmErr"); ok {
			if setter, ok := fn.(func(string)); ok {
				setter("")
			}
		}
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
				MaxLen(20).
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
				MaxLen(50).
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
				MaxLen(3).
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
				MaxLen(20).
				Password().
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
				MaxLen(20).
				Password().
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
