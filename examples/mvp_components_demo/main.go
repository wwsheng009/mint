// MVP Components Demo - Store + Reducer 版本
//
// 采用 Store + Reducer 架构简化组件内状态管理
// 展示所有核心表单组件的 ForField() + FieldChangeIntent 模式：
//   - Input: 用户名、电子邮件
//   - Textarea: 个人简介
//   - Select: 国家选择
//   - Checkbox: 同意条款
//
// MVP 数据流（Store + Reducer）：
//   UI.Instance (缓冲) → FieldChangeIntent → Store (单一事实源) → VNode → UI.Instance (渲染同步)
//
// 架构改进：
// - 使用 Store[T] 代替 UseState + GlobalState（单一状态源）
// - 使用 Reducer[T] 纯函数处理所有 Intent
// - 无需 WithInit、无需反射、无类型断言
//
// 运行: go run main.go

package main

import (
	"fmt"
	"strconv"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/reducer"
	"github.com/wwsheng009/mint/runtime/store"
	"github.com/wwsheng009/mint/ui"
	selectcomp "github.com/wwsheng009/mint/ui/components/select"
)

// =============================================================================
// 自定义 Intent 类型
// =============================================================================

type ResetIntent struct{}

func (ResetIntent) IntentType() string { return "Reset" }
func (ResetIntent) StayPressed() bool  { return true }

type SubmitFormIntent struct{}

func (SubmitFormIntent) IntentType() string { return "SubmitForm" }
func (SubmitFormIntent) StayPressed() bool  { return true }

type BackFormIntent struct{}

func (BackFormIntent) IntentType() string { return "BackForm" }
func (BackFormIntent) StayPressed() bool  { return true }

// =============================================================================
// 状态定义
// =============================================================================

// AppState 应用状态 - 单一事实源
type AppState struct {
	// 表单字段
	Username string
	Email    string
	Bio      string
	Country  string  // 使用 string 存储索引
	Agree    string  // 使用 string 存储布尔值

	// 提交状态
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
		Bio:       "",
		Country:   "0",  // 默认选中第一个
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
		case "bio":
			s.Bio = fieldChange.Value
		case "country":
			s.Country = fieldChange.Value
		case "agree":
			s.Agree = fieldChange.Value
		}
		return s
	}).
	// 重置表单
	On(ResetIntent{}, func(s AppState, i intent.Intent) AppState {
		s.Username = ""
		s.Email = ""
		s.Bio = ""
		s.Country = "0"
		s.Agree = "false"
		s.Submitted = false
		return s
	}).
	// 提交表单
	On(SubmitFormIntent{}, func(s AppState, i intent.Intent) AppState {
		s.Submitted = true
		return s
	}).
	// 返回表单
	On(BackFormIntent{}, func(s AppState, i intent.Intent) AppState {
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
		ui.WithWidth(70),
		ui.WithHeight(35),
		ui.WithTitle("MVP Components Demo - Store + Reducer"),
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
		return SuccessView(state)
	}

	return FormView(state)
}

// FormView - 主表单视图
func FormView(state AppState) ui.VNode {
	return ui.VStack(
		ui.NewTextBuilder("🎨 MVP Components Demo").
			Bold(true).
			FgColor("cyan").
			Build(),
		ui.Text(""),
		ui.NewTextBuilder("ForField() + FieldChangeIntent Pattern").
			FgColor("gray").
			Build(),
		ui.NewTextBuilder("✅ 数据流: Instance → FieldChangeIntent → Store → VNode").
			FgColor("green").
			Build(),
		ui.NewTextBuilder("   - 单一状态源: Store[T]").
			FgColor("gray").
			Build(),
		ui.NewTextBuilder("   - 无类型断言: 纯函数 Reducer").
			FgColor("gray").
			Build(),
		ui.NewTextBuilder("   - 自动注册: RegisterToGlobal()").
			FgColor("gray").
			Build(),
		ui.Text(""),
		ui.NewTextBuilder("─").FgColor("gray").Build(),
		ui.Text(""),

		// 表单内容
		BasicFormFields(state),
		ui.NewTextBuilder("─").FgColor("gray").Build(),
		ui.Text(""),
		ProfileFormFields(state),

		ui.Text(""),
		ui.NewTextBuilder("─").FgColor("gray").Build(),
		ui.Text(""),

		// 底部按钮
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

// BasicFormFields - 基本信息字段 (Input + Checkbox)
func BasicFormFields(state AppState) ui.VNode {
	return ui.VStack(
		// Input 组件 - 用户名
		ui.NewTextBuilder("Username:").FgColor("blue").Build(),
		ui.HStack(
			ui.Text("  "),
			ui.NewInputBuilder().
				// ForField() 自动处理 FieldChangeIntent
				ForField(intent.BindField("username")).
				Value(state.Username).
				Placeholder("Enter username").
				Width(45).
				Build(),
		),

		ui.Text(""),

		// Input 组件 - 电子邮件
		ui.NewTextBuilder("Email:").FgColor("blue").Build(),
		ui.HStack(
			ui.Text("  "),
			ui.NewInputBuilder().
				ForField(intent.BindField("email")).
				Value(state.Email).
				Placeholder("Enter email").
				Width(45).
				Build(),
		),

		ui.Text(""),

		// Checkbox 组件 - 同意条款
		ui.HStack(
			ui.Text("  "),
			ui.NewCheckboxBuilder().
				ForField(intent.BindField("agree")).
				Checked(state.Agree == "true").
				Label("I agree to the terms and conditions").
				Build(),
		),

		ui.Text(""),
		ui.NewTextBuilder(fmt.Sprintf("✓ Store: username=%q, email=%q, agree=%v",
			state.Username, state.Email, state.Agree == "true")).FgColor("gray").Build(),
	)
}

// ProfileFormFields - 个人资料字段 (Select + Textarea)
func ProfileFormFields(state AppState) ui.VNode {
	countries := []selectcomp.Option{
		{Value: "us", Label: "United States"},
		{Value: "cn", Label: "China"},
		{Value: "jp", Label: "Japan"},
		{Value: "uk", Label: "United Kingdom"},
		{Value: "de", Label: "Germany"},
	}

	countryIdx := 0
	if idx, err := strconv.Atoi(state.Country); err == nil && idx >= 0 && idx < len(countries) {
		countryIdx = idx
	}

	var countryLabel string
	if countryIdx >= 0 && countryIdx < len(countries) {
		countryLabel = countries[countryIdx].Label
	} else {
		countryLabel = "Select a country"
	}

	return ui.VStack(
		// Select 组件 - 国家选择
		ui.NewTextBuilder("Country:").FgColor("blue").Build(),
		ui.HStack(
			ui.Text("  "),
			ui.NewSelectBuilder().
				Options(countries).
				Selected(countryIdx).
				// ForField() 会将选中的索引存储到 State
				ForField(intent.BindField("country")).
				Width(45).
				Build(),
		),

		ui.Text(""),

		// Textarea 组件 - 个人简介
		ui.NewTextBuilder("Bio:").FgColor("blue").Build(),
		ui.HStack(
			ui.Text("  "),
			ui.NewTextareaBuilder().
				ForField(intent.BindField("bio")).
				Value(state.Bio).
				Placeholder("Tell us about yourself...").
				Rows(5).
				Cols(45).
				Build(),
		),

		ui.Text(""),
		ui.NewTextBuilder(fmt.Sprintf("✓ Store: country=%s (%s), bio chars=%d",
			countryLabel, state.Country, len(state.Bio))).FgColor("gray").Build(),
	)
}

// SuccessView - 成功提交视图
func SuccessView(state AppState) ui.VNode {
	countries := []selectcomp.Option{
		{Value: "us", Label: "United States"},
		{Value: "cn", Label: "China"},
		{Value: "jp", Label: "Japan"},
		{Value: "uk", Label: "United Kingdom"},
		{Value: "de", Label: "Germany"},
	}

	countryIdx := 0
	if idx, err := strconv.Atoi(state.Country); err == nil && idx >= 0 && idx < len(countries) {
		countryIdx = idx
	}

	var countryLabel string
	if countryIdx >= 0 && countryIdx < len(countries) {
		countryLabel = countries[countryIdx].Label
	} else {
		countryLabel = "None"
	}

	return ui.VStack(
		ui.NewTextBuilder("✅ Form Submitted Successfully!").
			Bold(true).
			FgColor("green").
			Build(),
		ui.Text(""),
		ui.NewTextBuilder("─").FgColor("gray").Build(),
		ui.Text(""),

		ui.NewTextBuilder(fmt.Sprintf("Username: %s", state.Username)).Build(),
		ui.NewTextBuilder(fmt.Sprintf("Email: %s", state.Email)).Build(),
		ui.NewTextBuilder(fmt.Sprintf("Country: %s", countryLabel)).Build(),
		ui.NewTextBuilder(fmt.Sprintf("Bio: %s", state.Bio)).Build(),
		ui.NewTextBuilder(fmt.Sprintf("Agreed: %v", state.Agree == "true")).Build(),

		ui.Text(""),
		ui.NewTextBuilder("─").FgColor("gray").Build(),
		ui.Text(""),

		ui.HStack(
			ui.Text("  "),
			ui.NewButtonBuilder("  Back to Form  ").
				Variant(ui.ButtonVariantSecondary).
				OnPress(BackFormIntent{}).
				Build(),
			ui.Text(" "),
			ui.NewButtonBuilder("  Reset  ").
				Variant(ui.ButtonVariantDanger).
				OnPress(ResetIntent{}).
				Build(),
		),
	)
}
