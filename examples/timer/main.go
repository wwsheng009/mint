package main

import (
	"fmt"
	"time"

	"github.com/wwsheng009/mint/ui"
)

// 自定义 Intent 类型 - 用于计时器控制
type StartTimerIntent struct{}
func (StartTimerIntent) IntentType() string { return "StartTimer" }
func (StartTimerIntent) StayPressed() bool  { return true }

type StopTimerIntent struct{}
func (StopTimerIntent) IntentType() string { return "StopTimer" }
func (StopTimerIntent) StayPressed() bool  { return true }

type ResetTimerIntent struct{}
func (ResetTimerIntent) IntentType() string { return "ResetTimer" }
func (ResetTimerIntent) StayPressed() bool  { return true }

func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
	}
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}

func TimerDemo() ui.VNode {
	// 1. 定义状态（hooks 必须在顶部）
	elapsed, setElapsed, _ := ui.UseStateInt(0) // 经过的秒数
	running, setRunning := ui.UseStateBool(false)

	// 2. 使用 Effect 实现定时器
	ui.UseEffect(func() ui.CleanupFunc {
		if !running {
			// 定时器未运行，返回空清理函数
			return func() {}
		}

		// 启动 ticker
		ticker := time.NewTicker(time.Second)
		done := make(chan struct{})

		go func() {
			for {
				select {
				case <-ticker.C:
					// 每秒更新 elapsed
					setElapsed(func(e int) int { return e + 1 })
				case <-done:
					ticker.Stop()
					return
				}
			}
		}()

		// 清理函数：停止 ticker
		return func() {
			close(done)
		}
	}, []interface{}{running}) // 依赖 running 状态

	// 3. 注册 Intent handler
	ui.On(StartTimerIntent{}, func() {
		setRunning(true)
	})
	ui.On(StopTimerIntent{}, func() {
		setRunning(false)
	})
	ui.On(ResetTimerIntent{}, func() {
		setRunning(false)
		setElapsed(0)
	})

	// 4. 计算显示状态
	elapsedDuration := time.Duration(elapsed) * time.Second
	timeStr := formatDuration(elapsedDuration)

	statusText := "Stopped"
	statusColor := "yellow"
	if running {
		statusText = "Running"
		statusColor = "green"
	}

	// 5. 返回 VNode
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

func main() {
	err := ui.Run(TimerDemo,
		ui.WithWidth(50),
		ui.WithHeight(14),
		ui.WithTitle("Timer Demo"),
	)
	if err != nil {
		panic(err)
	}
}
