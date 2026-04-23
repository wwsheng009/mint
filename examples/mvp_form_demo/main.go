// MVP Form Demo - Store + Reducer 版本
//
// 采用 Store + Reducer 架构简化组件内状态管理
// 展示 Intents 的数据流：
//   UI.Instance (缓冲) → Intent → Store (单一事实源) → VNode → UI.Instance (渲染同步)
//
// 架构改进：
// - 使用 Store[T] 作为单一状态源（代替 UseState + GlobalState）
// - 使用 Reducer[T] 纯函数处理所有 Intent
// - 无需 WithInit、无需反射、无需 GlobalState 临时保存
//
// 表单字段使用 FieldChangeIntent（系统内置机制），
// 其他操作使用自定义 Intent（SubmitForm, Reset, ClearSubmitted）。
//
// 详细说明请参考: docs/architecture/store/MIGRATION_GUIDE.md
//
// 运行: go run main.go

package main

import (
	"fmt"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/reducer"
	"github.com/wwsheng009/mint/runtime/store"
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
// 状态定义
// =============================================================================

// AppState 应用状态 - 单一事实源
type AppState struct {
	Username  string
	Email     string
	Agree     string  // 使用 string 存储布尔值（"true"/"false"）
	Submitted bool
}

// =============================================================================
// 全局 Store
// =============================================================================

var appStore *store.Store[AppState]

func initStore() {
	appStore = store.NewStore(AppState{
		Username:  "",
		Email:     "",
		Agree:     "false",
		Submitted: false,
	})
}

// =============================================================================
// Reducer 定义
// =============================================================================

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
		case "email":
			s.Email = fieldChange.Value
		case "agree":
			s.Agree = fieldChange.Value
		}
		return s
	}).
	// 重置表单
	On(ResetIntent{}, func(s AppState, i intent.Intent) AppState {
		s.Username = ""
		s.Email = ""
		s.Agree = "false"
		return s
	}).
	// 提交表单
	On(SubmitFormIntent{}, func(s AppState, i intent.Intent) AppState {
		s.Submitted = true
		return s
	}).
	// 清除提交状态
	On(ClearSubmittedIntent{}, func(s AppState, i intent.Intent) AppState {
		s.Submitted = false
		return s
	})

// =============================================================================
// 主函数
// =============================================================================

func main() {
	initStore()
	appReducer.RegisterToGlobal(appStore)

	err := ui.Run(App,
		ui.WithWidth(60),
		ui.WithHeight(25),
		ui.WithTitle("MVP Form Demo - Store + Reducer"),
	)
	if err != nil {
		panic(err)
	}
}

// =============================================================================
// 主应用组件
// =============================================================================

func App() ui.VNode {
	// 从 Store 读取最新状态（每次渲染时获取）
	state := appStore.Get()

	if state.Submitted {
		return SuccessView(state.Username, state.Email, state.Agree == "true")
	}

	return FormView(state)
}

// FormView - 主表单视图
func FormView(state AppState) ui.VNode {
	return ui.VStack(
		ui.NewTextBuilder("📝 MVP Intent Data Flow Demo").
			Bold(true).
			FgColor("cyan").
			Build(),
		ui.Text(""),
		ui.NewTextBuilder("✅ State is Single Source of Truth (Store[T])").
			FgColor("green").
			Build(),
		ui.NewTextBuilder("   - 无 UseState").
			FgColor("gray").
			Build(),
		ui.NewTextBuilder("   - 无类型断言").
			FgColor("gray").
			Build(),
		ui.NewTextBuilder("   - BuildAndRegister 自动注册").
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
				Value(state.Username).
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
				Value(state.Email).
				Placeholder("Enter email").
				Width(40).
				Build(),
		),

		ui.Text(""),

		ui.HStack(
			ui.Text("  "),
			ui.NewCheckboxBuilder().
				ForField(intent.BindField("agree")).
				Checked(state.Agree == "true").
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
				Disabled(state.Username == "" || state.Email == "" || state.Agree != "true").
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
