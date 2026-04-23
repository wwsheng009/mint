// Type-Safe Form Demo (RunApp + FieldMap Optimized)
//
// 演示如何使用 ui.RunApp[T] 与 AppRuntime + FieldMap 组合
//
// 优势：
//   1. 更简洁的 API - 不需要手动创建全局 Store
//   2. 自动状态订阅和重新渲染
//   3. FieldMap 消除 switch-case 硬编码
//   4. 类型安全
//
// 运行: go run ./examples/typesafe_form_demo/main_runapp.go

package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/wwsheng009/mint/framework"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/reducer"
	"github.com/wwsheng009/mint/runtime/scheduler"
	"github.com/wwsheng009/mint/runtime/statemachine"
	"github.com/wwsheng009/mint/runtime/store"
	mintui "github.com/wwsheng009/mint/ui"
	"github.com/wwsheng009/mint/ui/components/button"
	"github.com/wwsheng009/mint/ui/components/checkbox"
	"github.com/wwsheng009/mint/ui/components/input"
	modal "github.com/wwsheng009/mint/ui/components/modal"
)

// =============================================================================
// 应用状态定义
// =============================================================================

// AppState represents the form state with all fields.
type AppState struct {
	Username    string
	Email       string
	Age         int
	AcceptTerms bool
	Subscribe   bool
	// 表单提交状态
	Submitted   bool
	SubmittedAt time.Time

	// ========== Advanced Features State ==========

	// Lane Scheduler
	LaneStatus   string
	LastTaskTime time.Time

	// Modal
	ShowModal     bool
	ModalPosition string

	// Computed values (cached in state for display)
	EmailValid   bool
	FormComplete bool
	FormStrength int // 0-100

	// Middleware - display action count
	ActionCount int
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

// UndoIntent 撤销上一次状态更改
type UndoIntent struct{}

func (UndoIntent) IntentType() string { return "Undo" }

// RedoIntent 重做
type RedoIntent struct{}

func (RedoIntent) IntentType() string { return "Redo" }

// ShowModalIntent 显示模态框
type ShowModalIntent struct{}

func (ShowModalIntent) IntentType() string { return "ShowModal" }

// BackgroundTaskIntent 后台任务（演示Lane Scheduler）
type BackgroundTaskIntent struct{}

func (BackgroundTaskIntent) IntentType() string { return "BackgroundTask" }

// UpdateComputedIntent 更新计算值
type UpdateComputedIntent struct{}

func (UpdateComputedIntent) IntentType() string { return "UpdateComputed" }

// =============================================================================
// FieldMap - 消除硬编码
// =============================================================================

// fieldMap 定义所有字段的更新逻辑，避免 switch-case
var fieldMapBuilder = reducer.BindField(reducer.NewBuilder[AppState]()).
	BindFieldMap(map[string]func(AppState, string) AppState{
		// 字符串字段
		"Username": func(s AppState, val string) AppState {
			s.Username = val
			return s
		},
		"Email": func(s AppState, val string) AppState {
			s.Email = val
			return s
		},
		// 整型字段
		"Age": func(s AppState, val string) AppState {
			if v, err := strconv.Atoi(val); err == nil {
				s.Age = v
			}
			return s
		},
		// 布尔字段
		"AcceptTerms": func(s AppState, val string) AppState {
			s.AcceptTerms = val == "true"
			return s
		},
		"Subscribe": func(s AppState, val string) AppState {
			s.Subscribe = val == "true"
			return s
		},
	})

// Global computed values for derived state
var (
	emailValidComputed   *store.Computed[AppState, bool]
	formCompleteComputed *store.Computed[AppState, bool]
	formStrengthComputed *store.Computed[AppState, int]
)

func initComputed(appStore *store.Store[AppState]) {
	// =============================================================================
	// Computed/Selector
	// =============================================================================
	// Email validation computed value
	emailValidComputed = store.NewComputedWithInvalidator(
		appStore,
		func(s AppState) bool {
			// 简单的 email 验证：包含 @ 且格式基本正确
			return strings.Contains(s.Email, "@") &&
				len(s.Email) > 3 &&
				strings.Contains(s.Email, ".")
		},
		func(old, new AppState) bool {
			// 只有 Email 改变时才重新计算
			return old.Email != new.Email
		},
	)

	// Form completeness computed value
	formCompleteComputed = store.NewComputedWithInvalidator(
		appStore,
		func(s AppState) bool {
			return s.Username != "" &&
				s.Email != "" &&
				s.Age > 0 &&
				s.AcceptTerms
		},
		func(old, new AppState) bool {
			// 任何表单字段改变都检查
			return old.Username != new.Username ||
				old.Email != new.Email ||
				old.Age != new.Age ||
				old.AcceptTerms != new.AcceptTerms
		},
	)

	// Form strength computed value (0-100)
	formStrengthComputed = store.NewComputedWithInvalidator(
		appStore,
		func(s AppState) int {
			strength := 0
			if s.Username != "" {
				strength += 20
				if len(s.Username) >= 3 {
					strength += 10
				}
			}
			if s.Email != "" {
				strength += 20
				if strings.Contains(s.Email, "@") {
					strength += 10
				}
			}
			if s.Age > 0 {
				strength += 10
			}
			if s.AcceptTerms {
				strength += 20
			}
			if s.Subscribe {
				strength += 10
			}
			return strength
		},
		func(old, new AppState) bool {
			// 任何表单字段改变都重新计算
			return old.Username != new.Username ||
				old.Email != new.Email ||
				old.Age != new.Age ||
				old.AcceptTerms != new.AcceptTerms ||
				old.Subscribe != new.Subscribe
		},
	)
}

// ==================== Middleware ====================

// actionLog 记录所有 Intent 处理（用于 Middleware 演示）
var actionLog = []string{}

// loggingMiddleware 日志中间件
func loggingMiddleware(next reducer.Reducer[AppState]) reducer.Reducer[AppState] {
	return func(state AppState, i intent.Intent) AppState {
		now := time.Now().Format("15:04:05")
		entry := fmt.Sprintf("[%s] %s", now, i.IntentType())

		oldState := state
		newState := next(state, i)

		// 记录日志
		actionLog = append(actionLog, entry)
		if len(actionLog) > 10 {
			actionLog = actionLog[len(actionLog)-10:]
		}

		// 在 UI 中显示变化
		if oldState.Username != newState.Username ||
			oldState.Email != newState.Email ||
			oldState.AcceptTerms != newState.AcceptTerms {
			newState.LaneStatus = fmt.Sprintf("Action: %s", i.IntentType())
		}

		return newState
	}
}

// appReducerBuilder 扩展 FieldMap，添加表单特定的 Intent 处理
// 这里添加所有高级功能的 Intent 处理
var appReducerBuilder = fieldMapBuilder.GetBuilder().
	On(SubmitIntent{}, func(s AppState, i intent.Intent) AppState {
		// 设置提交状态，UI 会自动重新渲染
		s.Submitted = true
		s.SubmittedAt = time.Now()
		s.ActionCount++
		return s
	}).
	On(ResetIntent{}, func(s AppState, i intent.Intent) AppState {
		s.ActionCount++
		return AppState{
			Username:    "",
			Email:       "",
			Age:         0,
			AcceptTerms: false,
			Subscribe:   false,
			Submitted:   false,
			SubmittedAt: time.Time{},
			LaneStatus:  "Form Reset",
			ActionCount: s.ActionCount + 1,
		}
	}).
	// Modal - Show Modal
	On(ShowModalIntent{}, func(s AppState, i intent.Intent) AppState {
		s.ShowModal = true
		s.ActionCount++
		return s
	}).
	// Lane Scheduler - Background Task
	On(BackgroundTaskIntent{}, func(s AppState, i intent.Intent) AppState {
		s.LastTaskTime = time.Now()
		s.LaneStatus = fmt.Sprintf("Background task at %s", s.LastTaskTime.Format("15:04:05.000"))
		s.ActionCount++
		return s
	}).
	// Update computed values display
	On(UpdateComputedIntent{}, func(s AppState, i intent.Intent) AppState {
		// 更新状态中的计算值缓存
		if emailValidComputed != nil {
			s.EmailValid = emailValidComputed.Get()
		}
		if formCompleteComputed != nil {
			s.FormComplete = formCompleteComputed.Get()
		}
		if formStrengthComputed != nil {
			s.FormStrength = formStrengthComputed.Get()
		}
		return s
	})

// AppReducer 是构建后的 Reducer，应用了日志中间件
var AppReducer = reducer.WithMiddleware(
	appReducerBuilder.Build(),
	loggingMiddleware,
)

// =============================================================================
// 视图层 - 纯函数，从状态渲染 UI
// =============================================================================

// AppView 接收当前状态并返回 UI 树（返回 any 避免循环依赖）
func AppView(state AppState) any {
	return renderAppView(state)
}

// renderAppView 实际的渲染函数，返回 ui.VNode（完全类型安全）
func renderAppView(state AppState) mintui.VNode {
	// Build form components with ForField binding
	usernameInput := input.NewBuilder().
		Placeholder("Enter your username").
		ForField(intent.BindField("Username")).
		Value(state.Username).
		Width(30).
		Build()

	emailInput := input.NewBuilder().
		Placeholder("Enter your email").
		ForField(intent.BindField("Email")).
		Value(state.Email).
		Width(30).
		Build()

	ageInput := input.NewBuilder().
		Placeholder("Enter your age").
		ForField(intent.BindField("Age")).
		Value(reducer.FormatInt(state.Age)).
		Width(10).
		Build()

	termsCheckbox := checkbox.NewBuilder().
		Label("I accept the terms and conditions").
		ForField(intent.BindField("AcceptTerms")).
		Checked(state.AcceptTerms).
		Build()

	subscribeCheckbox := checkbox.NewBuilder().
		Label("Subscribe to newsletter").
		ForField(intent.BindField("Subscribe")).
		Checked(state.Subscribe).
		Build()

	submitButton := button.NewBuilder("Submit Form").
		OnPress(SubmitIntent{}).
		Variant(button.VariantPrimary).
		Build()

	resetButton := button.NewBuilder("Reset").
		OnPress(ResetIntent{}).
		Variant(button.VariantSecondary).
		Build()

	// Time Travel - Undo/Redo buttons
	undoButton := button.NewBuilder("⟲ Undo").
		OnPress(UndoIntent{}).
		Variant(button.VariantSecondary).
		Build()

	redoButton := button.NewBuilder("Redo ⟳").
		OnPress(RedoIntent{}).
		Variant(button.VariantSecondary).
		Build()

	// Modal button
	modalButton := button.NewBuilder("📦 Modal").
		OnPress(ShowModalIntent{}).
		Variant(button.VariantDanger).
		Build()

	// Lane Scheduler button
	bgTaskButton := button.NewBuilder("⚙️ BG Task").
		OnPress(BackgroundTaskIntent{}).
		Variant(button.VariantSuccess).
		Build()

	// Form layout with advanced features
	form := mintui.VStack(
		// Header
		mintui.NewTextBuilder("🚀 Advanced Features Demo").
			Bold(true).
			FgColor("cyan").
			Build(),
		mintui.NewTextBuilder("RunApp + FieldMap + 6 Advanced Features").
			FgColor("gray").
			Build(),
		mintui.Text(""),

		// Advanced Features Status Bar
		showAdvancedStatus(state),

		mintui.NewTextBuilder("─").FgColor("gray").Build(),
		mintui.Text(""),

		// 表单字段
		mintui.HStack(mintui.Text("Username: "), mintui.Text("  "), usernameInput),
		mintui.HStack(mintui.Text("Email:    "), mintui.Text("  "), emailInput),
		mintui.HStack(mintui.Text("Age:      "), mintui.Text("  "), ageInput),

		mintui.Text(""),

		termsCheckbox,
		subscribeCheckbox,

		mintui.Text(""),

		// Computed Values Display
		showComputedValues(state),

		mintui.Text(""),

		// 按钮行 - Time Travel + Modal + Lane Scheduler
		mintui.HStack(
			mintui.Text("  "),
			submitButton,
			mintui.Text(" "),
			resetButton,
			mintui.Text(" "),
			undoButton,
			mintui.Text(" "),
			redoButton,
			mintui.Text(" "),
			modalButton,
			mintui.Text(" "),
			bgTaskButton,
		),

		mintui.Text(""),
		mintui.NewTextBuilder("─").FgColor("gray").Build(),
		mintui.Text(""),

		// 状态显示
		showFormState(state),

		// Modal (conditionally shown)
		showModal(state),
	)

	return form
}

// showFormState 显示表单状态（包括提交信息）
func showFormState(state AppState) mintui.VNode {
	var lines []mintui.VNode

	if state.Submitted {
		lines = append(lines,
			mintui.Text(""),
			mintui.NewTextBuilder("──────────────────────────────────────────").
				FgColor("gray").
				Build(),
			mintui.NewTextBuilder("✅ Form Submitted Successfully!").
				FgColor("green").
				Bold(true).
				Build(),
			mintui.NewTextBuilder("──────────────────────────────────────────").
				FgColor("gray").
				Build(),
			mintui.NewTextBuilder("Username:  "+state.Username).FgColor("white").Build(),
			mintui.NewTextBuilder("Email:     "+state.Email).FgColor("white").Build(),
			mintui.NewTextBuilder("Age:       "+strconv.Itoa(state.Age)).FgColor("white").Build(),
			mintui.NewTextBuilder("Terms:     "+fmt.Sprintf("%v", state.AcceptTerms)).FgColor("white").Build(),
			mintui.NewTextBuilder("Subscribe: "+fmt.Sprintf("%v", state.Subscribe)).FgColor("white").Build(),
			mintui.NewTextBuilder("Submitted: "+state.SubmittedAt.Format("15:04:05")).FgColor("gray").Build(),
		)
	}

	// 实时状态显示（折叠表单内容时不显示）
	lines = append(lines,
		mintui.NewTextBuilder("──────────────────────────────────────────").
			FgColor("gray").
			Build(),
	)

	return mintui.VStack(lines...)
}

// showAdvancedStatus 显示高级功能状态栏
func showAdvancedStatus(state AppState) mintui.VNode {
	strengthColor := "red"
	if state.FormStrength >= 70 {
		strengthColor = "green"
	} else if state.FormStrength >= 40 {
		strengthColor = "yellow"
	}

	strengthBar := ""
	for i := 0; i < state.FormStrength/10; i++ {
		strengthBar += "█"
	}
	for i := state.FormStrength / 10; i < 10; i++ {
		strengthBar += "░"
	}

	return mintui.HStack(
		mintui.NewTextBuilder("Actions:").FgColor("gray").Build(),
		mintui.NewTextBuilder(fmt.Sprintf(" %d", state.ActionCount)).FgColor("cyan").Build(),
		mintui.Text("  |  "),
		mintui.NewTextBuilder("Status:").FgColor("gray").Build(),
		mintui.NewTextBuilder(fmt.Sprintf(" %s", state.LaneStatus)).FgColor("yellow").Build(),
		mintui.Text("  |  "),
		mintui.NewTextBuilder("Strength:").FgColor("gray").Build(),
		mintui.NewTextBuilder(fmt.Sprintf(" %s", strengthBar)).FgColor(strengthColor).Build(),
		mintui.NewTextBuilder(fmt.Sprintf(" %d%%", state.FormStrength)).FgColor(strengthColor).Build(),
	)
}

// showComputedValues 显示计算值
func showComputedValues(state AppState) mintui.VNode {
	emailColor := "red"
	if state.EmailValid {
		emailColor = "green"
	}

	completeColor := "gray"
	completeText := "Incomplete"
	if state.FormComplete {
		completeColor = "green"
		completeText = "Complete ✓"
	}

	return mintui.VStack(
		mintui.NewTextBuilder("📊 Computed Values (Auto-updated):").FgColor("bright-cyan").Build(),
		mintui.HStack(
			mintui.Text("  "),
			mintui.NewTextBuilder("Email:").FgColor("gray").Build(),
			mintui.NewTextBuilder(fmt.Sprintf("%v", state.EmailValid)).FgColor(emailColor).Build(),
			mintui.Text("  "),
			mintui.NewTextBuilder("Form:").FgColor("gray").Build(),
			mintui.NewTextBuilder(completeText).FgColor(completeColor).Build(),
		),
		mintui.HStack(
			mintui.Text("  "),
			mintui.NewTextBuilder("Computed from:").FgColor("gray").Build(),
			mintui.NewTextBuilder("Email, Username, Age, Terms").FgColor("bright-black").Build(),
		),
	)
}

// showModal 显示模态框
func showModal(state AppState) mintui.VNode {
	if !state.ShowModal {
		return mintui.Text("")
	}

	return modal.NewBuilder().
		Key("demo-modal").
		Title("📦 Advanced Features Modal").
		Content(mintui.VStack(
			mintui.NewTextBuilder("This modal demonstrates:").
				FgColor("cyan").
				Bold(true).
				Build(),
			mintui.Text(""),
			mintui.NewTextBuilder("1. PluginSetup - Modal middleware").
				FgColor("white").
				Build(),
			mintui.NewTextBuilder("2. Lane Scheduler - BG tasks").
				FgColor("white").
				Build(),
			mintui.NewTextBuilder("3. Time Travel - Undo/Redo").
				FgColor("white").
				Build(),
			mintui.NewTextBuilder("4. Computed/Selector values").
				FgColor("white").
				Build(),
			mintui.NewTextBuilder("5. Middleware - Action logging").
				FgColor("white").
				Build(),
			mintui.NewTextBuilder("6. WithDefaultLane priority").
				FgColor("white").
				Build(),
			mintui.Text(""),
			mintui.NewTextBuilder("Press ESC or click outside to close").
				FgColor("gray").
				Build(),
		)).
		Width(40).
		Height(14).
		Center().
		Open(true).
		OnClose(intent.CloseModalIntent{}).
		Rounded().
		Build()
}

// =============================================================================
// 主函数
// =============================================================================

func main() {
	// 创建初始状态
	initialState := AppState{
		Username:    "",
		Email:       "",
		Age:         0,
		AcceptTerms: false,
		Subscribe:   false,
		SubmittedAt: time.Time{},
		LaneStatus:  "Ready",
	}

	// 创建 AppRuntime - 整合 Store、Reducer 和 View
	// Time Travel Debugging - WithMaxHistory(100)
	rt := statemachine.NewAppRuntime(
		initialState,
		AppView,
		AppReducer,
		statemachine.WithMaxHistory(100),
	)

	// 初始化计算值
	initComputed(rt.GetStore())

	// 使用 ui.RunApp[T] 启动应用 - 包含所有 6 个高级功能
	err := mintui.RunApp(rt,
		// 1. Lane Scheduler (通过Lane调度优化渲染)
		mintui.WithLaneScheduler(),
		// 2. WithDefaultLane (设置默认优先级)
		mintui.WithDefaultLane(uint32(scheduler.InputLane)),
		// 3. PluginSetup (插件设置)
		mintui.WithPluginSetup(func(app *framework.App) {
			// 注册 modal 中间件支持 ESC 键和点击外部关闭
			app.AddMiddleware(modal.NewModalMiddleware())
		}),
		// 4. Time Travel Debugging (WithMaxHistory 已在上面的 statemachine.NewAppRuntime 中启用)
		// 5. Computed/Selector (已在 initComputed 中设置)
		// 6. Middleware (已在 AppReducer 中通过 reducer.WithMiddleware 设置)

		// UI 配置
		mintui.WithWidth(80),
		mintui.WithHeight(50),
		mintui.WithTitle("RunApp + Advanced Features Demo"),
		mintui.WithInit(func() {
			// 注册 Intent handlers 到全局 Intent Runtime
			// 连接了 AppRuntime 的 Store 和 Intent 系统
			appReducerBuilder.RegisterToGlobal(rt.GetStore())

			// 覆盖builtin的CloseModalIntent handler
			// Modal middleware emit CloseModalIntent，现在由我们的Store处理
			intent.RegisterTypedWithOpts(
				intent.DefaultRegistry(),
				func(ctx *intent.ActionContext, i intent.CloseModalIntent) intent.IntentResult {
					// 更新state.ShowModal = false（而非GlobalState的__modal_visible）
					current := rt.GetStore().Get()
					if current.ShowModal {
						newState := current
						newState.ShowModal = false
						newState.ActionCount++
						rt.GetStore().Set(newState)
						ctx.ScheduleUpdate()
					}
					return intent.HandledResult()
				},
				intent.WithOverridable(true), // 允许覆盖builtin handler
			)

			// Time Travel - Undo/Redo handlers
			// 这些handler直接调用AppRuntime的Undo()/Redo()，而不是通过reducer
			intent.RegisterTypedWithOpts(
				intent.DefaultRegistry(),
				func(ctx *intent.ActionContext, i UndoIntent) intent.IntentResult {
					// 检查是否可以Undo
					if !rt.CanUndo() {
						// 什么都不做，直接返回HandledResult
						return intent.HandledResult()
					}

					// 执行Undo（rt.Undo()内部已经调用了store.Set并触发更新）
					// 我们什么都不做，只返回HandledResult
					if err := rt.Undo(); err != nil {
						// Undo失败，但不返回ErrorResult，避免阻塞其他输入
						return intent.HandledResult()
					}
					return intent.HandledResult()
				},
				intent.WithOverridable(true),
			)

			intent.RegisterTypedWithOpts(
				intent.DefaultRegistry(),
				func(ctx *intent.ActionContext, i RedoIntent) intent.IntentResult {
					// 检查是否可以Redo
					if !rt.CanRedo() {
						// 什么都做，直接返回HandledResult
						return intent.HandledResult()
					}

					// 执行Redo（rt.Redo()内部已经调用了store.Set并触发更新）
					// 我们什么都不做，只返回HandledResult
					if err := rt.Redo(); err != nil {
						// Redo失败，但不返回ErrorResult，避免阻塞其他输入
						return intent.HandledResult()
					}
					return intent.HandledResult()
				},
				intent.WithOverridable(true),
			)
		}),
	)
	if err != nil {
		panic(err)
	}
}
