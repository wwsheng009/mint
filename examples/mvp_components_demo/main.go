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

	"github.com/wwsheng009/mint/framework"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/reducer"
	"github.com/wwsheng009/mint/runtime/store"
	"github.com/wwsheng009/mint/runtime/style"
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

type SetInteractionModeIntent struct {
	Mode string
}

func (SetInteractionModeIntent) IntentType() string { return "SetInteractionMode" }
func (SetInteractionModeIntent) StayPressed() bool  { return true }

// =============================================================================
// 状态定义
// =============================================================================

// AppState 应用状态 - 单一事实源
type AppState struct {
	// 表单字段
	Username string
	Email    string
	Bio      string
	Country  string // 使用 string 存储索引
	Agree    string // 使用 string 存储布尔值

	// 提交状态
	Submitted bool

	// 交互模式
	InteractionMode string // interactive | app_selection | terminal_selection
}

// =============================================================================
// 全局 Store
// =============================================================================

var appStore *store.Store[AppState]
var runtimeApp *framework.App

func initStore() {
	appStore = store.NewStore(AppState{
		Username:        "",
		Email:           "",
		Bio:             "",
		Country:         "0", // 默认选中第一个
		Agree:           "false",
		Submitted:       false,
		InteractionMode: "interactive",
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
	}).
	// 切换交互模式（由 Store 状态驱动）
	On(SetInteractionModeIntent{}, func(s AppState, i intent.Intent) AppState {
		modeIntent, ok := i.(SetInteractionModeIntent)
		if !ok {
			return s
		}
		s.InteractionMode = normalizeModeString(modeIntent.Mode)
		return s
	})

// =============================================================================
// 主函数
// =============================================================================

func main() {
	initStore()
	appReducer.RegisterToGlobal(appStore)
	appStore.Subscribe(func(s AppState) {
		applyRuntimeInteractionMode(s.InteractionMode)
	})

	err := ui.Run(App,
		ui.WithWidth(70),
		ui.WithHeight(35),
		ui.WithTitle("MVP Components Demo - Store + Reducer"),
		ui.WithInteractionMode(ui.InteractionModeInteractive),
		ui.WithPluginSetup(func(app *framework.App) {
			runtimeApp = app
			selectcomp.Install(app)
			applyRuntimeInteractionMode(appStore.Get().InteractionMode)

			// F6: 循环切换三种交互模式
			app.OnKeyCombo("f6", func() {
				nextMode, err := app.CycleInteractionMode()
				if err != nil {
					return
				}
				appStore.Update(func(s AppState) AppState {
					s.InteractionMode = modeToString(nextMode)
					return s
				})
			})

			// Ctrl+1/2/3: 直接切换模式
			app.OnKeyCombo("ctrl+1", func() { setModeFromShortcut("interactive") })
			app.OnKeyCombo("ctrl+2", func() { setModeFromShortcut("app_selection") })
			app.OnKeyCombo("ctrl+3", func() { setModeFromShortcut("terminal_selection") })
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
		InteractionStatusBar(state),
		ui.Text(""),

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
		ui.Text(""),
		ui.HStack(
			ui.Text("  "),
			ui.NewButtonBuilder("Interactive").
				Variant(modeButtonVariant(state.InteractionMode, "interactive")).
				OnPress(SetInteractionModeIntent{Mode: "interactive"}).
				Build(),
			ui.Text(" "),
			ui.NewButtonBuilder("App Selection").
				Variant(modeButtonVariant(state.InteractionMode, "app_selection")).
				OnPress(SetInteractionModeIntent{Mode: "app_selection"}).
				Build(),
			ui.Text(" "),
			ui.NewButtonBuilder("Terminal Selection").
				Variant(modeButtonVariant(state.InteractionMode, "terminal_selection")).
				OnPress(SetInteractionModeIntent{Mode: "terminal_selection"}).
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

func setModeFromShortcut(mode string) {
	normalized := normalizeModeString(mode)
	appStore.Update(func(s AppState) AppState {
		s.InteractionMode = normalized
		return s
	})
}

func applyRuntimeInteractionMode(mode string) {
	if runtimeApp == nil {
		return
	}
	target := stringToMode(mode)
	if runtimeApp.GetInteractionMode() == target {
		return
	}
	_ = runtimeApp.SetInteractionMode(target)
}

func normalizeModeString(mode string) string {
	switch mode {
	case "interactive", "app_selection", "terminal_selection":
		return mode
	default:
		return "interactive"
	}
}

func stringToMode(mode string) framework.InteractionMode {
	switch normalizeModeString(mode) {
	case "app_selection":
		return framework.InteractionModeAppSelection
	case "terminal_selection":
		return framework.InteractionModeTerminalSelection
	default:
		return framework.InteractionModeInteractive
	}
}

func modeToString(mode framework.InteractionMode) string {
	switch mode {
	case framework.InteractionModeAppSelection:
		return "app_selection"
	case framework.InteractionModeTerminalSelection:
		return "terminal_selection"
	default:
		return "interactive"
	}
}

func modeLabel(mode string) string {
	switch normalizeModeString(mode) {
	case "app_selection":
		return "App Selection"
	case "terminal_selection":
		return "Terminal Selection"
	default:
		return "Interactive"
	}
}

func modeButtonVariant(current, target string) ui.ButtonVariant {
	if normalizeModeString(current) == normalizeModeString(target) {
		return ui.ButtonVariantPrimary
	}
	return ui.ButtonVariantSecondary
}

func InteractionStatusBar(state AppState) ui.VNode {
	modeColor := modeSectionColor(state.InteractionMode)
	nextMode := nextInteractionMode(state.InteractionMode)

	bar := ui.NewStatusBarBuilder().
		Theme(
			ui.StatusBarThemeDefault().
				WithTooltipBorderStyle(style.NewStyle().Foreground(style.BrightWhite).Background(style.Blue).Bold(true)).
				WithTooltipShadowStyle(style.NewStyle().Foreground(style.BrightBlack).Background(style.Blue)).
				WithTooltipArrowStyle(ui.StatusBarTooltipArrowRounded),
		).
		HelpDisplayMode(ui.StatusBarHelpOverlay).
		TooltipPlacement(ui.StatusBarTooltipAuto).
		TooltipGapRows(0).
		TooltipMaxWidth(38).
		HelpPrefix("> ").
		HelpFallback("Hover actions for overlay help | Select supports Enter, Up/Down, and mouse click | F6 and Ctrl+1/2/3 still work").
		Left(
			ui.StatusBarActionBadge(" MODE ", "black", modeColor, SetInteractionModeIntent{Mode: nextMode}).
				WithHelp("Cycle to the next interaction mode"),
		).
		Left(
			ui.StatusBarActionText(" "+modeLabel(state.InteractionMode)+" ", SetInteractionModeIntent{Mode: nextMode}).
				WithWidth(20).
				WithEllipsis().
				WithHelp("Current mode: click or press Enter to cycle"),
		).
		Center(
			ui.StatusBarText(" Tab/Enter/Up/Down | Mouse | F6 | Ctrl+1-3 ").
				WithWidth(40).
				WithAlign(ui.AlignCenter).
				WithBold(true).
				WithEllipsis().
				WithTooltip("Select: Enter opens, Up/Down moves, mouse click selects"),
		).
		Right(modeStatusSection(" UI ", "interactive", state.InteractionMode)).
		Right(modeStatusSection(" APP ", "app_selection", state.InteractionMode)).
		Right(modeStatusSection(" TERM ", "terminal_selection", state.InteractionMode))

	return ui.Padding(bar.BuildWithHelp(), 0, 1, 0, 1)
}

func modeSectionColor(mode string) string {
	switch normalizeModeString(mode) {
	case "app_selection":
		return "cyan"
	case "terminal_selection":
		return "green"
	default:
		return "yellow"
	}
}

func nextInteractionMode(mode string) string {
	switch normalizeModeString(mode) {
	case "app_selection":
		return "terminal_selection"
	case "terminal_selection":
		return "interactive"
	default:
		return "app_selection"
	}
}

func modeStatusSection(label, targetMode, currentMode string) ui.StatusBarSection {
	active := normalizeModeString(currentMode) == normalizeModeString(targetMode)
	fgColor := "bright-white"
	bgColor := "bright-black"
	if active {
		fgColor = "black"
		bgColor = modeSectionColor(targetMode)
	}
	return ui.StatusBarActionBadge(label, fgColor, bgColor, SetInteractionModeIntent{Mode: targetMode}).
		WithKey("mode-" + targetMode).
		WithHelp(modeStatusHelp(targetMode, active))
}

func modeStatusHelp(mode string, active bool) string {
	prefix := "Switch to "
	if active {
		prefix = "Already in "
	}
	switch normalizeModeString(mode) {
	case "app_selection":
		return prefix + "App Selection: app-managed text copy mode"
	case "terminal_selection":
		return prefix + "Terminal Selection: native terminal selection mode"
	default:
		return prefix + "Interactive: regular UI mouse and keyboard mode"
	}
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
				SetID("profile.country").
				OverlayPopup(true).
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
		InteractionStatusBar(state),
		ui.Text(""),

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
