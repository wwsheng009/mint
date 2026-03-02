// MVP Form Demo
//
// 采用 ui.On 简化组件内状态管理
// 展示 Intents 的 MVP 架构数据流：
//   UI.Instance (缓冲) → Intent → State (事实源) → VNode → UI.Instance (渲染同步)
//
// 三种 Intent 管理模式：
//   1. 组件级状态 - ui.On + UseState + Simple* Intent（推荐组件内状态）
//   2. 全局状态 - runtime/intent 内置函数
//   3. 自定义 Intent + ui.On（本示例）
//
// 本示例演示：
// - 无需 WithInit：直接在组件内使用 RegisterIntent 注册
// - 无需反射：handler 闭包直接访问 setter 变量
// - 无需 GlobalState 临时保存 setter
//
// 表单字段使用 FieldChangeIntent（系统内置机制），
// 其他操作使用 ui.On 注册自定义 Intent。
//
// 详细说明请参考: docs/architecture/mvp/INTENT_MANAGEMENT_PATTERNS.md
//
// 运行: go run main.go

package main

import (
	"fmt"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/ui"
)

// =============================================================================
// 自定义 Intent 类型
// =============================================================================

// ResetIntent 重置表单
type ResetIntent struct{}
func (ResetIntent) IntentType() string { return "Reset" }
func (ResetIntent) StayPressed() bool  { return true }

// SubmitFormIntent 提交表单
type SubmitFormIntent struct{}
func (SubmitFormIntent) IntentType() string { return "SubmitForm" }
func (SubmitFormIntent) StayPressed() bool  { return true }

// ClearSubmittedIntent 清除提交状态
type ClearSubmittedIntent struct{}
func (ClearSubmittedIntent) IntentType() string { return "ClearSubmitted" }
func (ClearSubmittedIntent) StayPressed() bool  { return true }

// =============================================================================
// 主函数
// =============================================================================

func main() {
	err := ui.Run(App,
		ui.WithWidth(60),
		ui.WithHeight(25),
		ui.WithTitle("MVP Form Demo - ui.On + Custom Intents"),
		ui.WithInit(func() {
			// 注册 FieldChangeIntent 处理器（表单字段变更）
			// 在组件外注册避免重复注册
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
				case "agree":
					setAgree, _ := ctx.GetState("agreeSetter")
					if fn, ok := setAgree.(func(bool)); ok {
						agreeVal := i.Value == "true"
						fn(agreeVal)
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

func App() ui.VNode {
	// 使用 UseState 创建状态，state 是单一事实源
	username, setUsername := ui.UseStateString("")
	email, setEmail := ui.UseStateString("")
	agree, setAgree := ui.UseStateBool(false)
	submitted, setSubmitted := ui.UseStateBool(false)

	// 将 setter 保存到 GlobalState 供 Intent Handler 使用
	ctx := ui.GetCurrentContext()
	if ctx != nil {
		ctx.GlobalState["usernameSetter"] = setUsername
		ctx.GlobalState["emailSetter"] = setEmail
		ctx.GlobalState["agreeSetter"] = setAgree
		ctx.GlobalState["submittedSetter"] = setSubmitted
	}

	// 使用 ui.On 注册自定义 Intent 处理器（简化注册 API）
	// ui.On 有去重机制，多次渲染只注册一次

	// Reset 处理器
	ui.On(ResetIntent{}, func() {
		setUsername("")
		setEmail("")
		setAgree(false)
	})

	// Submit 处理器
	ui.On(SubmitFormIntent{}, func() {
		setSubmitted(true)
	})

	// Back 处理器
	ui.On(ClearSubmittedIntent{}, func() {
		setSubmitted(false)
	})

	if submitted {
		return SuccessView(username, email, agree)
	}

	return ui.VStack(
		ui.NewTextBuilder("📝 MVP Intent Data Flow Demo").
			Bold(true).
			FgColor("cyan").
			Build(),
		ui.Text(""),
		ui.NewTextBuilder("State is Single Source of Truth").
			FgColor("gray").
			Build(),
		ui.Text(""),
		ui.NewTextBuilder("─").
			FgColor("gray").
			Build(),

		ui.Text(""),

		ui.NewTextBuilder("Username:").
			FgColor("blue").
			Build(),

		ui.HStack(
			ui.Text("  "),
			ui.NewInputBuilder().
				ForField(intent.BindField("username")).
				Value(username).
				Placeholder("Enter username").
				Width(40).
				Build(),
		),

		ui.Text(""),

		ui.NewTextBuilder("Email:").
			FgColor("blue").
			Build(),

		ui.HStack(
			ui.Text("  "),
			ui.NewInputBuilder().
				ForField(intent.BindField("email")).
				Value(email).
				Placeholder("Enter email").
				Width(40).
				Build(),
		),

		ui.Text(""),

		ui.HStack(
			ui.Text("  "),
			ui.NewCheckboxBuilder().
				ForField(intent.BindField("agree")).
				Checked(agree).
				Label("I agree to the terms").
				Build(),
		),

		ui.Text(""),
		ui.NewTextBuilder("─").
			FgColor("gray").
			Build(),
		ui.Text(""),

		ui.HStack(
			ui.Text("  "),
			ui.NewButtonBuilder("  Submit  ").
				Variant(ui.ButtonVariantPrimary).
				OnPress(SubmitFormIntent{}).
				Disabled(username == "" || email == "" || !agree).
				Build(),
			ui.Text(" "),
			ui.NewButtonBuilder("  Reset  ").
				Variant(ui.ButtonVariantSecondary).
				OnPress(ResetIntent{}).
				Build(),
		),
	)
}

// SuccessView - 成功提交视图
func SuccessView(username, email string, agree bool) ui.VNode {
	return ui.VStack(
		ui.NewTextBuilder("✅ Form Submitted Successfully!").
			Bold(true).
			FgColor("green").
			Build(),
		ui.Text(""),
		ui.NewTextBuilder("─").
			FgColor("gray").
			Build(),
		ui.Text(""),
		ui.NewTextBuilder(fmt.Sprintf("Username: %s", username)).Build(),
		ui.NewTextBuilder(fmt.Sprintf("Email: %s", email)).Build(),
		ui.NewTextBuilder(fmt.Sprintf("Agreed: %v", agree)).Build(),
		ui.Text(""),
		ui.NewTextBuilder("─").
			FgColor("gray").
			Build(),
		ui.Text(""),
		ui.HStack(
			ui.Text("  "),
			ui.NewButtonBuilder("  Back to Form  ").
				Variant(ui.ButtonVariantSecondary).
				OnPress(ClearSubmittedIntent{}).
				Build(),
		),
	)
}
