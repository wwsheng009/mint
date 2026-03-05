// Validation Demo - Store + Reducer 版本
//
// 演示如何将 validation 包与 Input 组件结合使用
// 展示表单验证的完整流程：
//   1. 定义验证规则
//   2. 输入时实时验证
//   3. 提交时全面验证
//   4. 显示验证错误信息
//
// 架构：Store + Reducer（状态管理改进为单一状态源）
// - AppState: 单一事实源，包含所有表单状态
// - Reducer: 纯函数，处理所有 Intent
// - BuildAndRegister: 自动注册 handlers
//
// 运行: go run ./examples/validation_demo/

package main

import (
	"fmt"
	"strconv"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/reducer"
	"github.com/wwsheng009/mint/runtime/store"
	"github.com/wwsheng009/mint/ui"
	"github.com/wwsheng009/mint/ui/components/validation"
)

// =============================================================================
// 状态定义
// =============================================================================

// AppState 应用状态 - 单一事实源
type AppState struct {
	// 表单值
	Username  string
	Email     string
	Age       string
	Password  string
	ConfirmPwd string

	// 验证错误
	UsernameErr  string
	EmailErr     string
	AgeErr       string
	PasswordErr  string
	ConfirmErr   string
}

// =============================================================================
// Intent 定义
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
// 全局 Store
// =============================================================================

var appStore *store.Store[AppState]

func initStore() {
	appStore = store.NewStore(AppState{
		Username:   "",
		Email:      "",
		Age:        "",
		Password:   "",
		ConfirmPwd: "",

		UsernameErr:  "",
		EmailErr:     "",
		AgeErr:       "",
		PasswordErr:  "",
		ConfirmErr:   "",
	})
}

// =============================================================================
// Reducer 定义
// =============================================================================

// 定义验证器（在组件外避免重复创建）
var (
	usernameValidator = validation.NewChain().
			Required().
			MinLength(3).
			MaxLength(20).
			Build()

	emailValidator = validation.NewChain().
			Required().
			Email().
			Build()

	passwordValidator = validation.NewChain().
			Required().
			MinLength(6).
			MaxLength(20).
			Build()
)

// validateField 单个字段验证
func validateField(field string, value string, pwd string) string {
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

// 定义 Reducer
var appReducer = reducer.NewBuilder[AppState]().
	// 字段变更 - 自动更新状态
	On(intent.FieldChangeIntent{}, func(s AppState, i intent.Intent) AppState {
		fieldChange, ok := i.(intent.FieldChangeIntent)
		if !ok {
			return s
		}

		switch fieldChange.Field {
		case "username":
			s.Username = fieldChange.Value
			// 可以在这里进行实时验证
			// s.UsernameErr = validateField("username", s.Username, "")
		case "email":
			s.Email = fieldChange.Value
		case "age":
			s.Age = fieldChange.Value
		case "password":
			s.Password = fieldChange.Value
		case "confirmPwd":
			s.ConfirmPwd = fieldChange.Value
		}
		return s
	}).
	// 提交表单 - 验证所有字段
	On(SubmitIntent{}, func(s AppState, i intent.Intent) AppState {
		s.UsernameErr = validateField("username", s.Username, "")
		s.EmailErr = validateField("email", s.Email, "")
		s.AgeErr = validateField("age", s.Age, "")
		s.PasswordErr = validateField("password", s.Password, "")
		s.ConfirmErr = validateField("confirmPwd", s.ConfirmPwd, s.Password)
		return s
	}).
	// 重置表单 - 清空所有状态
	On(ResetIntent{}, func(s AppState, i intent.Intent) AppState {
		s.Username = ""
		s.Email = ""
		s.Age = ""
		s.Password = ""
		s.ConfirmPwd = ""
		s.UsernameErr = ""
		s.EmailErr = ""
		s.AgeErr = ""
		s.PasswordErr = ""
		s.ConfirmErr = ""
		return s
	})

// =============================================================================
// 主函数
// =============================================================================

func main() {
	initStore()
	appReducer.RegisterToGlobal(appStore)

	err := ui.Run(FormApp,
		ui.WithWidth(70),
		ui.WithHeight(35),
		ui.WithTitle("Validation Form Demo (Store + Reducer)"),
	)
	if err != nil {
		panic(err)
	}
}

// =============================================================================
// 主应用组件
// =============================================================================

func FormApp() ui.VNode {
	// 从 Store 读取最新状态（每次渲染时获取）
	state := appStore.Get()

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

		// 使用说明
		ui.NewTextBuilder("✅ 架构改进: Store + Reducer (无 UseState)").
			FgColor("green").
			Build(),
		ui.NewTextBuilder("   - 单一状态源: AppState").
			FgColor("gray").
			Build(),
		ui.NewTextBuilder("   - 无类型断言: 纯函数 Reducer").
			FgColor("gray").
			Build(),
		ui.NewTextBuilder("   - 自动注册: RegisterToGlobal()").
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
				Value(state.Username).
				Placeholder("3-20字符").
				Width(30).
				MaxLen(20).
				Build(),
		),
		showError(state.UsernameErr),

		// 邮箱
		ui.HStack(
			ui.NewTextBuilder("Email:").
				FgColor("blue").
				Build(),
			ui.NewInputBuilder().
				ForField(intent.BindField("email")).
				Value(state.Email).
				Placeholder("example@mail.com").
				Width(30).
				MaxLen(50).
				Build(),
		),
		showError(state.EmailErr),

		// 年龄
		ui.HStack(
			ui.NewTextBuilder("Age:").
				FgColor("blue").
				Build(),
			ui.NewInputBuilder().
				ForField(intent.BindField("age")).
				Value(state.Age).
				Placeholder("1-150").
				Width(30).
				MaxLen(3).
				Build(),
		),
		showError(state.AgeErr),

		// 密码
		ui.HStack(
			ui.NewTextBuilder("Password:").
				FgColor("blue").
				Build(),
			ui.NewInputBuilder().
				ForField(intent.BindField("password")).
				Value(state.Password).
				Placeholder("6-20字符").
				Width(30).
				MaxLen(20).
				Password().
				Build(),
		),
		showError(state.PasswordErr),

		// 确认密码
		ui.HStack(
			ui.NewTextBuilder("Confirm:").
				FgColor("blue").
				Build(),
			ui.NewInputBuilder().
				ForField(intent.BindField("confirmPwd")).
				Value(state.ConfirmPwd).
				Placeholder("再次输入密码").
				Width(30).
				MaxLen(20).
				Password().
				Build(),
		),
		showError(state.ConfirmErr),

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
		ui.NewTextBuilder(fmt.Sprintf("  Username: %s", state.Username)).Build(),
		ui.NewTextBuilder(fmt.Sprintf("  Email: %s", state.Email)).Build(),
		ui.NewTextBuilder(fmt.Sprintf("  Age: %s", state.Age)).Build(),
	)
}
