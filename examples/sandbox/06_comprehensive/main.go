// 06_comprehensive/main.go
// 综合示例 (Store 模式)
//
// 演示多个高级功能的组合使用：
// - 事件录制与回放
// - 快照系统
// - TestHelper 链式 API
// - 队列统计
// - 事件注入策略

package main

import (
	"fmt"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/reducer"
	"github.com/wwsheng009/mint/runtime/store"
	"github.com/wwsheng009/mint/ui"
)

// ============================================================================
// AppState - 定义应用状态
// ============================================================================

type AppState struct {
	Count int    // 计数器值
	Name  string // 用户名称
	Step  int    // 当前步骤 (1-3)
}

// ============================================================================
// Intent Types
// ============================================================================

type NextStepIntent struct{}
func (NextStepIntent) IntentType() string { return "NextStep" }
func (NextStepIntent) StayPressed() bool  { return true }

type BackStepIntent struct{}
func (BackStepIntent) IntentType() string { return "BackStep" }
func (BackStepIntent) StayPressed() bool  { return true }

type DecrementComprehensiveIntent struct{}
func (DecrementComprehensiveIntent) IntentType() string { return "Decrement" }
func (DecrementComprehensiveIntent) StayPressed() bool  { return true }

type IncrementComprehensiveIntent struct{}
func (IncrementComprehensiveIntent) IntentType() string { return "Increment" }
func (IncrementComprehensiveIntent) StayPressed() bool  { return true }

type SetComprehensiveNameIntent struct {
	Name string
}
func (SetComprehensiveNameIntent) IntentType() string { return "SetComprehensiveName" }
func (SetComprehensiveNameIntent) StayPressed() bool  { return false }

// ============================================================================
// Store 初始化
// ============================================================================

var comprehensiveStore = store.NewStore(AppState{
	Count: 0,
	Name:  "Guest",
	Step:  1,
})

// ============================================================================
// Reducer 注册
// ============================================================================

func init() {
	reducer.NewBuilder[AppState]().
		On(NextStepIntent{}, func(s AppState, i intent.Intent) AppState {
			if s.Step < 3 {
				s.Step++
			}
			return s
		}).
		On(BackStepIntent{}, func(s AppState, i intent.Intent) AppState {
			if s.Step > 1 {
				s.Step--
			}
			return s
		}).
		On(DecrementComprehensiveIntent{}, func(s AppState, i intent.Intent) AppState {
			if s.Count > 0 {
				s.Count--
			}
			return s
		}).
		On(IncrementComprehensiveIntent{}, func(s AppState, i intent.Intent) AppState {
			s.Count++
			return s
		}).
		On(SetComprehensiveNameIntent{}, func(s AppState, i intent.Intent) AppState {
			s.Name = i.(SetComprehensiveNameIntent).Name
			return s
		}).
		BuildAndRegister(intent.DefaultRegistry(), comprehensiveStore)
}

// ============================================================================
// ComprehensiveApp - 综合演示应用
// ============================================================================

func ComprehensiveApp() ui.VNode {
	// ✅ 订阅存储的状态
	count := ui.UseStoreSelector(comprehensiveStore, func(s AppState) int { return s.Count })
	name := ui.UseStoreSelector(comprehensiveStore, func(s AppState) string { return s.Name })
	step := ui.UseStoreSelector(comprehensiveStore, func(s AppState) int { return s.Step })

	steps := []string{
		"Step 1: Enter name",
		"Step 2: Adjust counter",
		"Step 3: Confirm",
	}

	return ui.VStack(
		ui.NewTextBuilder("╔══════════════════════════════╗").
			FgColor("cyan").
			Build(),
		ui.NewTextBuilder("║     Comprehensive Demo         ║").
			FgColor("cyan").
			Build(),
		ui.NewTextBuilder("╚══════════════════════════════╝").
			FgColor("cyan").
			Build(),
		ui.Text(""),
		ui.HStack(
			ui.Text("Step: "),
			ui.NewTextBuilder(fmt.Sprintf("%d/3", step)).
				FgColor("yellow").
				Bold(true).
				Build(),
		),
		ui.Text(""),
		ui.NewTextBuilder(steps[step-1]).
			FgColor("bright-black").
			Build(),
		ui.Text(""),

		// Step 1: Name input
		func() ui.VNode {
			if step == 1 {
				return ui.HStack(
					ui.Text("Name: "),
					ui.NewInputBuilder().
						Value(name).
						Placeholder("Your name").
						Build(), // TODO: integrate with FieldChangeIntent
				)
			}
			return ui.Text("")
		}(),

		// Step 2 & 3: Counter and buttons
		func() ui.VNode {
			if step >= 2 {
				return ui.VStack(
					ui.HStack(
						ui.Text("Hello, "),
						ui.NewTextBuilder(name).
							FgColor("magenta").
							Bold(true).
							Build(),
						ui.Text("!"),
					),
					ui.Text(""),
					ui.NewTextBuilder(fmt.Sprintf("Count: %d", count)).
						FgColor("green").
						Bold(true).
						Build(),
					ui.Text(""),
					ui.HStack(
						ui.NewButtonBuilder("  [ - ]  ").
							OnPress(DecrementComprehensiveIntent{}).
							Build(),
						ui.Text(" "),
						ui.NewButtonBuilder("  [ + ]  ").
							OnPress(IncrementComprehensiveIntent{}).
							Build(),
					),
				)
			}
			return ui.Text("")
		}(),

		ui.Text(""),
		ui.NewButtonBuilder("  [ Next ]  ").
			OnPress(NextStepIntent{}).
			Build(),
		ui.NewButtonBuilder("  [ Back ]  ").
			OnPress(BackStepIntent{}).
			Build(),
	)
}

// ============================================================================
// Main
// ============================================================================

func main() {
	err := ui.Run(ComprehensiveApp,
		ui.WithWidth(40),
		ui.WithHeight(20),
		ui.WithTitle("Comprehensive Demo (Store 模式)"),
	)
	if err != nil {
		panic(err)
	}
}
