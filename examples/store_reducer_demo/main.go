// Store + Reducer Demo
//
// 演示如何使用 Store 和 Reducer 进行状态管理
// 这是一种更架构化的方式，遵循单向数据流原则：
//   Intent → Dispatcher → Reducer → Store → View
//
// 优势：
//  1. 单一真相源 - 所有状态集中在一个 Store
//  2. 可预测的状态变更 - 只有 Reducer 能修改状态
//  3. 易于测试 - Reducer 是纯函数
//  4. 可调试 - 可实现时间旅行调试
//
// 运行: go run ./examples/store_reducer_demo/

package main

import (
	"fmt"
	"strconv"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/reducer"
	"github.com/wwsheng009/mint/runtime/store"
	"github.com/wwsheng009/mint/ui"
)

// =============================================================================
// 应用状态定义
// =============================================================================

// AppState 是应用的单一状态源
type AppState struct {
	// 计数器
	Count int

	// 表单字段
	Username string
	Email    string

	// 表单验证错误
	UsernameErr string
	EmailErr    string

	// UI 状态
	ActiveTab int
}

// =============================================================================
// Intent 定义
// =============================================================================

// IncrementIntent 增加计数
type IncrementIntent struct{}

func (IncrementIntent) IntentType() string { return "Increment" }
func (IncrementIntent) StayPressed() bool   { return true }

// DecrementIntent 减少计数
type DecrementIntent struct{}

func (DecrementIntent) IntentType() string { return "Decrement" }
func (DecrementIntent) StayPressed() bool   { return true }

// SetUsernameIntent 设置用户名
type SetUsernameIntent struct {
	Value string
}

func (SetUsernameIntent) IntentType() string { return "SetUsername" }

// SetEmailIntent 设置邮箱
type SetEmailIntent struct {
	Value string
}

func (SetEmailIntent) IntentType() string { return "SetEmail" }

// SubmitFormIntent 提交表单
type SubmitFormIntent struct{}

func (SubmitFormIntent) IntentType() string { return "SubmitForm" }
func (SubmitFormIntent) StayPressed() bool   { return true }

// ResetFormIntent 重置表单
type ResetFormIntent struct{}

func (ResetFormIntent) IntentType() string { return "ResetForm" }
func (ResetFormIntent) StayPressed() bool   { return true }

// SwitchTabIntent 切换标签页
type SwitchTabIntent struct {
	Index int
}

func (SwitchTabIntent) IntentType() string { return "SwitchTab" }

// =============================================================================
// Reducer 定义 - 唯一的状态变更入口
// =============================================================================

// AppReducer 是应用的核心 Reducer
// 所有状态变更逻辑都集中在这里
var AppReducer = reducer.NewBuilder[AppState]().
	On(IncrementIntent{}, func(s AppState, i intent.Intent) AppState {
		s.Count++
		return s
	}).
	On(DecrementIntent{}, func(s AppState, i intent.Intent) AppState {
		s.Count--
		return s
	}).
	On(SetUsernameIntent{}, func(s AppState, i intent.Intent) AppState {
		if typed, ok := i.(SetUsernameIntent); ok {
			s.Username = typed.Value
			// 实时验证
			if len(s.Username) < 3 {
				s.UsernameErr = "用户名至少3个字符"
			} else {
				s.UsernameErr = ""
			}
		}
		return s
	}).
	On(SetEmailIntent{}, func(s AppState, i intent.Intent) AppState {
		if typed, ok := i.(SetEmailIntent); ok {
			s.Email = typed.Value
			// 实时验证（简化示例）
			if s.Email != "" && !containsAt(s.Email) {
				s.EmailErr = "邮箱格式不正确"
			} else {
				s.EmailErr = ""
			}
		}
		return s
	}).
	On(SubmitFormIntent{}, func(s AppState, i intent.Intent) AppState {
		// 提交时验证所有字段
		if len(s.Username) < 3 {
			s.UsernameErr = "用户名至少3个字符"
		}
		if s.Email == "" || !containsAt(s.Email) {
			s.EmailErr = "请输入有效邮箱"
		}
		// 如果没有错误，可以提交...
		return s
	}).
	On(ResetFormIntent{}, func(s AppState, i intent.Intent) AppState {
		s.Username = ""
		s.Email = ""
		s.UsernameErr = ""
		s.EmailErr = ""
		return s
	}).
	On(SwitchTabIntent{}, func(s AppState, i intent.Intent) AppState {
		if typed, ok := i.(SwitchTabIntent); ok {
			s.ActiveTab = typed.Index
		}
		return s
	}).
	Build()

func containsAt(s string) bool {
	for _, c := range s {
		if c == '@' {
			return true
		}
	}
	return false
}

// =============================================================================
// 全局 Store
// =============================================================================

var appStore *store.Store[AppState]

func initStore() {
	appStore = store.NewStore(AppState{
		Count:      0,
		Username:   "",
		Email:      "",
		ActiveTab:  0,
	})
}

// =============================================================================
// 主函数
// =============================================================================

func main() {
	initStore()

	err := ui.Run(App,
		ui.WithWidth(60),
		ui.WithHeight(30),
		ui.WithTitle("Store + Reducer Demo"),
		ui.WithInit(func() {
			// 注册 Intent 处理器
			// Intent → Reducer → Store → 触发重新渲染
			ui.RegisterIntent(func(ctx *intent.ActionContext, i intent.Intent) intent.IntentResult {
				// 从 Store 获取当前状态
				currentState := appStore.Get()

				// 通过 Reducer 计算新状态
				newState := AppReducer.Reduce(currentState, i)

				// 更新 Store
				appStore.Set(newState)

				return intent.HandledResult()
			})
		}),
	)
	if err != nil {
		panic(err)
	}
}

// =============================================================================
// 视图层 - 纯函数，从 Store 读取状态
// =============================================================================

func App() ui.VNode {
	// 从 Store 获取当前状态
	state := appStore.Get()

	return ui.VStack(
		// 标题
		ui.NewTextBuilder("📦 Store + Reducer Architecture Demo").
			Bold(true).
			FgColor("cyan").
			Build(),
		ui.Text(""),

		// 架构说明
		ui.NewTextBuilder("Intent → Dispatcher → Reducer → Store → View").
			FgColor("gray").
			Build(),
		ui.Text(""),

		// 分割线
		ui.NewTextBuilder("──────────────────────────────────────────").
			FgColor("bright-black").
			Build(),
		ui.Text(""),

		// 计数器演示
		ui.NewTextBuilder("Counter (State.Count = " + strconv.Itoa(state.Count) + ")").
			Bold(true).
			Build(),
		ui.HStack(
			ui.NewButtonBuilder(" - ").
				Variant(ui.ButtonVariantSecondary).
				OnPress(DecrementIntent{}).
				Build(),
			ui.Text(" "),
			ui.NewTextBuilder(strconv.Itoa(state.Count)).
				FgColor("yellow").
				Bold(true).
				Build(),
			ui.Text(" "),
			ui.NewButtonBuilder(" + ").
				Variant(ui.ButtonVariantPrimary).
				OnPress(IncrementIntent{}).
				Build(),
		),
		ui.Text(""),

		// 分割线
		ui.NewTextBuilder("──────────────────────────────────────────").
			FgColor("bright-black").
			Build(),
		ui.Text(""),

		// 表单演示
		ui.NewTextBuilder("Form (State in Store)").
			Bold(true).
			Build(),
		ui.HStack(
			ui.NewTextBuilder("Username:").
				FgColor("blue").
				Build(),
			ui.NewInputBuilder().
				ForField(intent.BindField("username")).
				Value(state.Username).
				Placeholder("3+ chars").
				Width(20).
				Build(),
		),
		showError(state.UsernameErr),
		ui.HStack(
			ui.NewTextBuilder("Email:").
				FgColor("blue").
				Build(),
			ui.NewInputBuilder().
				ForField(intent.BindField("email")).
				Value(state.Email).
				Placeholder("user@example.com").
				Width(20).
				Build(),
		),
		showError(state.EmailErr),

		// 表单按钮
		ui.HStack(
			ui.NewButtonBuilder(" Submit ").
				Variant(ui.ButtonVariantPrimary).
				OnPress(SubmitFormIntent{}).
				Build(),
			ui.Text(" "),
			ui.NewButtonBuilder(" Reset ").
				Variant(ui.ButtonVariantSecondary).
				OnPress(ResetFormIntent{}).
				Build(),
		),
		ui.Text(""),

		// 状态显示
		ui.NewTextBuilder("──────────────────────────────────────────").
			FgColor("bright-black").
			Build(),
		ui.NewTextBuilder("Current State:").Bold(true).Build(),
		ui.NewTextBuilder(fmt.Sprintf("  Count: %d", state.Count)).FgColor("bright-black").Build(),
		ui.NewTextBuilder(fmt.Sprintf("  Username: %q", state.Username)).FgColor("bright-black").Build(),
		ui.NewTextBuilder(fmt.Sprintf("  Email: %q", state.Email)).FgColor("bright-black").Build(),
	)
}

func showError(err string) ui.VNode {
	if err == "" {
		return ui.Text("")
	}
	return ui.NewTextBuilder("  ⚠ " + err).
		FgColor("red").
		Build()
}
