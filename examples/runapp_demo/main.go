// RunApp Demo
//
// 演示如何使用 ui.RunApp[T] 与 AppRuntime 结合使用
// 这是最推荐的 Store + Reducer 架构用法，提供了最佳的开发体验
//
// 优势：
//  1. 更简洁的 API - 不需要手动创建全局 Store
//  2. 自动状态订阅和重新渲染
//  3. 支持时间旅行调试（通过 AppRuntime 的 History API）
//  4. 类型安全
//
// 运行: go run ./examples/runapp_demo/

package main

import (
	"strconv"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/reducer"
	"github.com/wwsheng009/mint/runtime/statemachine"
	"github.com/wwsheng009/mint/ui"
)

// =============================================================================
// 应用状态定义
// =============================================================================

// AppState 是应用的单一状态源
type AppState struct {
	// 计数器
	Count int

	// 计数器历史（用于展示）
	History []int

	// 最大计数值
	MaxCount int
}

// =============================================================================
// Intent 定义
// =============================================================================

// IncrementIntent 增加计数
type IncrementIntent struct{}

func (IncrementIntent) IntentType() string { return "DemoIncrement" }
func (IncrementIntent) StayPressed() bool   { return true }

// DecrementIntent 减少计数
type DecrementIntent struct{}

func (DecrementIntent) IntentType() string { return "DemoDecrement" }
func (DecrementIntent) StayPressed() bool   { return true }

// ResetIntent 重置计数
type ResetIntent struct{}

func (ResetIntent) IntentType() string { return "DemoReset" }
func (ResetIntent) StayPressed() bool   { return true }

// =============================================================================
// Reducer 定义 - 唯一的状态变更入口
// =============================================================================

// AppReducer 是应用的 Reducer，纯函数，处理所有状态变更
//
// 注意：这个 Reducer 需要注册到 Intent Runtime 才能工作。
// 在 ui.RunApp 模式下，我们需要使用 WithInit 来注册 handlers，
// 或者使用传统的全局 Store 模式（见 store_reducer_demo）。
var appReducerBuilder = reducer.NewBuilder[AppState]().
	On(IncrementIntent{}, func(s AppState, i intent.Intent) AppState {
		s.Count++
		// 添加到历史记录
		s.History = append(s.History, s.Count)
		// 更新最大值
		if s.Count > s.MaxCount {
			s.MaxCount = s.Count
		}
		return s
	}).
	On(DecrementIntent{}, func(s AppState, i intent.Intent) AppState {
		if s.Count > 0 {
			s.Count--
			s.History = append(s.History, s.Count)
		}
		return s
	}).
	On(ResetIntent{}, func(s AppState, i intent.Intent) AppState {
		s.Count = 0
		s.History = []int{0}
		return s
	})

// AppReducer 是构建后的 Reducer（未注册）
var AppReducer = appReducerBuilder.Build()

// =============================================================================
// 视图层 - 纯函数，从状态渲染 UI
// =============================================================================

// AppView 接收当前状态并返回 UI 树
// 这是纯函数，不包含任何状态，只进行渲染
// 注意：ViewFunction 返回类型是 `any` 以避免循环依赖
func AppView(state AppState) any {
	return renderAppView(state)
}

// renderAppView 实际的渲染函数，返回 ui.VNode
func renderAppView(state AppState) ui.VNode {
	// 计算历史进度条
	historyUI := renderHistory(state.History)

	return ui.VStack(
		// 标题
		ui.NewTextBuilder("🚀 ui.RunApp[T] Demo").
			Bold(true).
			FgColor("cyan").
			Build(),
		ui.Text(""),

		// 架构说明
		ui.NewTextBuilder("Store + Reducer + AppRuntime = 🎉").
			FgColor("gray").
			Build(),
		ui.NewTextBuilder("Automatic state subscription and re-rendering!").
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

		// 重置按钮
		ui.HStack(
			ui.NewButtonBuilder(" Reset ").
				Variant(ui.ButtonVariantSecondary).
				OnPress(ResetIntent{}).
				Build(),
		),
		ui.Text(""),

		// 分割线
		ui.NewTextBuilder("──────────────────────────────────────────").
			FgColor("bright-black").
			Build(),

		// 状态显示
		ui.NewTextBuilder("Current State:").Bold(true).Build(),
		ui.NewTextBuilder("  Count: "+strconv.Itoa(state.Count)).FgColor("bright-black").Build(),
		ui.NewTextBuilder("  Max Count: "+strconv.Itoa(state.MaxCount)).FgColor("bright-black").Build(),
		ui.NewTextBuilder("  History Length: "+strconv.Itoa(len(state.History))).FgColor("bright-black").Build(),

		// 历史记录进度条
		ui.Text(""),
		ui.NewTextBuilder("History:").Bold(true).Build(),
		historyUI,
	)
}

// renderHistory 渲染历史记录的可视化进度条
func renderHistory(history []int) ui.VNode {
	if len(history) == 0 {
		return ui.Text("No history yet")
	}

	// 只显示最近的 20 个历史记录
	start := len(history) - 20
	if start < 0 {
		start = 0	}
	recent := history[start:len(history)]

	// 计算最大值用于归一化
	maxVal := 0
	for _, v := range recent {
		if v > maxVal {
			maxVal = v
		}
	}

	// 构建进度条 UI
	var bars []ui.VNode
	for _, v := range recent {
		// 归一化到 1-10 个字符
		width := 1
		if maxVal > 0 {
			width = (v * 10) / maxVal
			if width < 1 {
				width = 1
			}
		}

		// 根据值设置颜色
		color := "green"
		if v > 70 {
			color = "red"
		} else if v > 40 {
			color = "yellow"
		}

		// 创建进度条
		bar := ""
		for i := 0; i < width; i++ {
			bar += "█"
		}
		for i := width; i < 10; i++ {
			bar += "░"
		}
		bar += " " + strconv.Itoa(v)

		bars = append(bars, ui.NewTextBuilder(bar).FgColor(color).Build())
	}

	return ui.VStack(bars...)
}

// =============================================================================
// 主函数
// =============================================================================

func main() {
	// 创建初始状态
	initialState := AppState{
		Count:    0,
		History:  []int{0},
		MaxCount: 0,
	}

	// 创建 AppRuntime - 整合 Store、Reducer 和 View
	rt := statemachine.NewAppRuntime(
		initialState,
		AppView,
		AppReducer,
		statemachine.WithMaxHistory(100),
	)

	// 使用 ui.RunApp[T] 启动应用
	// 这会自动：
	//   1. 订阅状态变化
	//   2. 触发重新渲染
	//   3. 支持时间旅行调试
	//
	// 注意：需要使用 ui.WithInit 注册 Intent handlers
	err := ui.RunApp(rt,
		ui.WithWidth(70),
		ui.WithHeight(40),
		ui.WithTitle("RunApp[T] Demo - Store + Reducer"),
		ui.WithInit(func() {
			// 注册 Intent handlers 到全局 Intent Runtime
			// 这连接了 AppRuntime 的 Store 和 Intent 系统
			appReducerBuilder.RegisterToGlobal(rt.GetStore())
		}),
	)
	if err != nil {
		panic(err)
	}
}
