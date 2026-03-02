// 06_comprehensive/main.go
// 综合示例
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

	"github.com/wwsheng009/mint/ui"
)

// Intent Types
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

// ComprehensiveApp 综合演示应用
func ComprehensiveApp() ui.VNode {
	count, setCount, _ := ui.UseStateInt(0)
	name, _ := ui.UseStateString("Guest")
	step, setStep, _ := ui.UseStateInt(1)

	// Register intent handlers
	ui.On(NextStepIntent{}, func() {
		if step < 3 {
			setStep(step + 1)
		}
	})
	ui.On(BackStepIntent{}, func() {
		if step > 1 {
			setStep(step - 1)
		}
	})
	ui.On(DecrementComprehensiveIntent{}, func() {
		if count > 0 {
			setCount(count - 1)
		}
	})
	ui.On(IncrementComprehensiveIntent{}, func() {
		setCount(count + 1)
	})

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

func main() {
	err := ui.Run(ComprehensiveApp,
		ui.WithWidth(40),
		ui.WithHeight(20),
		ui.WithTitle("Comprehensive Demo"),
	)
	if err != nil {
		panic(err)
	}
}
