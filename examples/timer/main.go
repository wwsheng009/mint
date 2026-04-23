package main

import (
	"fmt"
	"time"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/reducer"
	"github.com/wwsheng009/mint/runtime/store"
	"github.com/wwsheng009/mint/ui"
)

// =============================================================================
// AppState - Timer 状态
// =============================================================================

type AppState struct {
	Elapsed int  // 经过的秒数
	Running bool // 是否正在运行
}

// =============================================================================
// Intent Types
// =============================================================================

type StartTimerIntent struct{}
func (StartTimerIntent) IntentType() string { return "StartTimer" }
func (StartTimerIntent) StayPressed() bool  { return true }

type StopTimerIntent struct{}
func (StopTimerIntent) IntentType() string  { return "StopTimer" }
func (StopTimerIntent) StayPressed() bool   { return true }

type ResetTimerIntent struct{}
func (ResetTimerIntent) IntentType() string { return "ResetTimer" }
func (ResetTimerIntent) StayPressed() bool  { return true }

type TickTimerIntent struct{}
func (TickTimerIntent) IntentType() string  { return "TickTimer" }
func (TickTimerIntent) StayPressed() bool   { return false } // 非用户操作

// =============================================================================
// Store 初始化
// =============================================================================

var timerStore = store.NewStore(AppState{
	Elapsed: 0,
	Running: false,
})

// =============================================================================
// Reducer 注册
// =============================================================================

func init() {
	reducer.NewBuilder[AppState]().
		On(StartTimerIntent{}, func(s AppState, i intent.Intent) AppState {
			s.Running = true
			return s
		}).
		On(StopTimerIntent{}, func(s AppState, i intent.Intent) AppState {
			s.Running = false
			return s
		}).
		On(ResetTimerIntent{}, func(s AppState, i intent.Intent) AppState {
			s.Running = false
			s.Elapsed = 0
			return s
		}).
		On(TickTimerIntent{}, func(s AppState, i intent.Intent) AppState {
			if s.Running {
				s.Elapsed++
			}
			return s
		}).
		BuildAndRegister(intent.DefaultRegistry(), timerStore)
}

// =============================================================================
// Helper Functions
// =============================================================================

func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
	}
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}

// =============================================================================
// Main
// =============================================================================

func main() {
	err := ui.Run(TimerDemo,
		ui.WithWidth(50),
		ui.WithHeight(14),
		ui.WithTitle("Timer Demo (Store 模式)"),
	)
	if err != nil {
		panic(err)
	}
}

// =============================================================================
// Timer Demo Component
// =============================================================================

func TimerDemo() ui.VNode {
	// ✅ 订阅 timer 状态
	elapsed := ui.UseStoreSelector(timerStore, func(s AppState) int { return s.Elapsed })
	running := ui.UseStoreSelector(timerStore, func(s AppState) bool { return s.Running })

	// ✅ UseEffect 实现定时器 - 依赖 running 状态
	ui.UseEffect(func() ui.CleanupFunc {
		if !running {
			return func() {}
		}

		ticker := time.NewTicker(time.Second)
		done := make(chan struct{})

		go func() {
			for {
				select {
				case <-ticker.C:
					// ✅ 通过 Store 更新状态
					timerStore.Update(func(s AppState) AppState {
						s.Elapsed++
						return s
					})
				case <-done:
					ticker.Stop()
					return
				}
			}
		}()

		return func() {
			close(done)
		}
	}, []interface{}{running})

	// 计算显示状态
	elapsedDuration := time.Duration(elapsed) * time.Second
	timeStr := formatDuration(elapsedDuration)

	statusText := "Stopped"
	statusColor := "yellow"
	if running {
		statusText = "Running"
		statusColor = "green"
	}

	return ui.VStack(
		ui.NewTextBuilder("⏱ Timer").Bold(true).FgColor("cyan").Build(),
		ui.Text(""),
		ui.NewTextBuilder(timeStr).Bold(true).FgColor("bright-white").Build(),
		ui.Text(""),
		ui.NewTextBuilder(fmt.Sprintf("Status: %s", statusText)).FgColor(statusColor).Build(),
		ui.Text(""),
		ui.HStack(
			ui.NewButtonBuilder(" ▶ Start ").OnPress(StartTimerIntent{}).Build(),
			ui.Text(" "),
			ui.NewButtonBuilder(" ⏹ Stop ").OnPress(StopTimerIntent{}).Build(),
			ui.Text(" "),
			ui.NewButtonBuilder(" ↺ Reset ").OnPress(ResetTimerIntent{}).Build(),
		),
		ui.Text(""),
		ui.NewTextBuilder("Tab: focus | Enter: click | q: quit").FgColor("bright-black").Build(),
	)
}
