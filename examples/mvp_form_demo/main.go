// MVP Form Demo
//
// 这个示例演示 Intents 的 MVP 架构数据流：
//   Instance (缓冲) → FieldChangeIntent → State (事实源) → Setter → VNode → Instance (渲染同步)
//
// MVP 原则：
//   1. State (通过 Setter) 是单一事实源
//   2. Intent 携带最少数据 (Field, Value)
//   3. Instance 不能决定状态
//
// 运行: go run main.go

package main

import (
	"fmt"
	"reflect"

	"github.com/wwsheng009/mint/app"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/ui"
)

func main() {
	err := ui.Run(App,
		ui.WithWidth(60),
		ui.WithHeight(25),
		ui.WithTitle("MVP Form Demo - Intents Architecture"),
		ui.WithInit(func() {
			// 注册 FieldChangeIntent handler
			ui.RegisterIntent(func(ctx *intent.ActionContext, i intent.FieldChangeIntent) intent.IntentResult {
				// 通过 setter 更新状态
				switch i.Field {
				case "username":
					setUsername, _ := ctx.GetState("usernameSetter")
					callSetter(setUsername, i.Value)
				case "email":
					setEmail, _ := ctx.GetState("emailSetter")
					callSetter(setEmail, i.Value)
				case "agree":
					setAgree, _ := ctx.GetState("agreeSetter")
					agreeVal := i.Value == "true"
					callSetter(setAgree, agreeVal)
				}
				return intent.HandledResult()
			})

			// 注册 Reset 意图
			ui.RegisterIntent(func(ctx *intent.ActionContext, i ResetIntent) intent.IntentResult {
				setUsername, _ := ctx.GetState("usernameSetter")
				callSetter(setUsername, "")
				setEmail, _ := ctx.GetState("emailSetter")
				callSetter(setEmail, "")
				setAgree, _ := ctx.GetState("agreeSetter")
				callSetter(setAgree, false)
				return intent.HandledResult()
			})

			// 注册 Submit 意图
			ui.RegisterIntent(func(ctx *intent.ActionContext, i SubmitFormIntent) intent.IntentResult {
				setSubmitted, _ := ctx.GetState("submittedSetter")
				callSetter(setSubmitted, true)
				return intent.HandledResult()
			})

			// 注册 ClearSubmitted 意图
			ui.RegisterIntent(func(ctx *intent.ActionContext, i ClearSubmittedIntent) intent.IntentResult {
				setSubmitted, _ := ctx.GetState("submittedSetter")
				callSetter(setSubmitted, false)
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

// ResetIntent 重置表单
type ResetIntent struct{}

func (ResetIntent) IntentType() string { return "Reset" }

// SubmitFormIntent 提交表单
type SubmitFormIntent struct{}

func (SubmitFormIntent) IntentType() string { return "SubmitForm" }

// ClearSubmittedIntent 清除提交状态
type ClearSubmittedIntent struct{}

func (ClearSubmittedIntent) IntentType() string { return "ClearSubmitted" }

// =============================================================================

// App - 主应用组件
func App() ui.VNode {
	// MVP: 使用 UseState，state 是单一事实源
	username, setUsername := ui.UseStateString("")
	email, setEmail := ui.UseStateString("")
	agree, setAgree := ui.UseStateBool(false)
	submitted, setSubmitted := ui.UseStateBool(false)

	// 保存 setters 到 State 供 Intent Handler 使用
	ctx := ui.GetCurrentContext()
	if ctx != nil {
		ctx.GlobalState["usernameSetter"] = setUsername
		ctx.GlobalState["emailSetter"] = setEmail
		ctx.GlobalState["agreeSetter"] = setAgree
		ctx.GlobalState["submittedSetter"] = setSubmitted
	}

	if submitted {
		return SuccessView(username, email, agree)
	}

	return ui.VStack(
		app.NewTextBuilder("📝 MVP Intent Data Flow Demo").
			Bold(true).
			FgColor("cyan").
			Build(),
		app.Text(""),
		app.NewTextBuilder("State is Single Source of Truth").
			FgColor("gray").
			Build(),
		app.Text(""),
		app.NewTextBuilder("─").
			FgColor("gray").
			Build(),

		app.Text(""),

		app.NewTextBuilder("Username:").
			FgColor("blue").
			Build(),

		app.HStack(
			app.Text("  "),
			app.InputBuilder().
				ForField(intent.BindField("username")).
				Value(username).
				Placeholder("Enter username").
				Width(40).
				Build(),
		),

		app.Text(""),

		app.NewTextBuilder("Email:").
			FgColor("blue").
			Build(),

		app.HStack(
			app.Text("  "),
			app.InputBuilder().
				ForField(intent.BindField("email")).
				Value(email).
				Placeholder("Enter email").
				Width(40).
				Build(),
		),

		app.Text(""),

		app.HStack(
			app.Text("  "),
			app.CheckboxBuilder().
				ForField(intent.BindField("agree")).
				Checked(agree).
				Label("I agree to the terms").
				Build(),
		),

		app.Text(""),
		app.NewTextBuilder("─").
			FgColor("gray").
			Build(),
		app.Text(""),

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

// SuccessView - 成功提交视图
func SuccessView(username, email string, agree bool) ui.VNode {
	return ui.VStack(
		app.NewTextBuilder("✅ Form Submitted Successfully!").
			Bold(true).
			FgColor("green").
			Build(),
		app.Text(""),
		app.NewTextBuilder("─").
			FgColor("gray").
			Build(),
		app.Text(""),
		app.NewTextBuilder(fmt.Sprintf("Username: %s", username)).Build(),
		app.NewTextBuilder(fmt.Sprintf("Email: %s", email)).Build(),
		app.NewTextBuilder(fmt.Sprintf("Agreed: %v", agree)).Build(),
		app.Text(""),
		app.NewTextBuilder("─").
			FgColor("gray").
			Build(),
		app.Text(""),
		app.HStack(
			app.Text("  "),
			app.ButtonBuilder("  Back to Form  ").
				Variant(app.ButtonVariantSecondary).
				OnPress(ClearSubmittedIntent{}).
				Build(),
		),
	)
}
