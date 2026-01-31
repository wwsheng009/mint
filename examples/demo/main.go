package main

import (
	"github.com/wwsheng009/mint/ui"
)

// DemoApp demonstrates all UI components
func DemoApp() ui.VNode {
	currentTab, setCurrentTab := ui.UseStateString("counter")
	counter, setCounter, _ := ui.UseStateInt(0)
	text, setText := ui.UseStateString("")
	checked1, setChecked1 := ui.UseStateBool(false)
	checked2, setChecked2 := ui.UseStateBool(false)
	checked3, setChecked3 := ui.UseStateBool(false)

	return ui.VStack(
		ui.NewTextBuilder("╔═══════════════════════════════════════╗").
			FgColor("cyan").
			Build(),
		ui.NewTextBuilder("║     Mint UI Declarative Framework     ║").
			FgColor("cyan").
			Build(),
		ui.NewTextBuilder("╚═══════════════════════════════════════╝").
			FgColor("cyan").
			Build(),
		ui.Text(""),
		// Tab navigation
		ui.HStack(
			ui.ButtonBuilder(" [1] Counter ").
				OnClick(func() { setCurrentTab("counter") }).
				Build(),
			ui.Text(" "),
			ui.ButtonBuilder(" [2] Input ").
				OnClick(func() { setCurrentTab("input") }).
				Build(),
			ui.Text(" "),
			ui.ButtonBuilder(" [3] Tasks ").
				OnClick(func() { setCurrentTab("tasks") }).
				Build(),
		),
		ui.Text(""),
		ui.NewTextBuilder("───────────────────────────────────────").
			FgColor("bright-black").
			Build(),
		ui.Text(""),
		// Content based on selected tab
		func() ui.VNode {
			if currentTab == "counter" {
				return ui.Fragment(
					ui.NewTextBuilder("📊 Counter Demo").
						FgColor("yellow").
						Bold(true).
						Build(),
					ui.Text(""),
					ui.NewTextBuilder("   Count:   ").
						FgColor("bright-black").
						Build(),
					ui.HStack(
						ui.NewTextBuilder("   ").
							FgColor("green").
							Bold(true).
							Build(),
						ui.NewTextBuilder("  ").
							FgColor("green").
							Bold(true).
							Build(),
						ui.ButtonBuilder("  [ - ]  ").
							OnClick(func() { setCounter(counter - 1) }).
							Build(),
						ui.Text(" "),
						ui.ButtonBuilder("  [ + ]  ").
							OnClick(func() { setCounter(counter + 1) }).
							Build(),
					),
				)
			} else if currentTab == "input" {
				return ui.Fragment(
					ui.NewTextBuilder("📝 Input Demo").
						FgColor("yellow").
						Bold(true).
						Build(),
					ui.Text(""),
					ui.HStack(
						ui.Text("Name: "),
						ui.InputBuilder().
							Value(text).
							Placeholder("Type here...").
							MaxLength(20).
							OnChange(setText).
							Build(),
					),
				)
			} else {
				return ui.Fragment(
					ui.NewTextBuilder("✓ Task List").
						FgColor("yellow").
						Bold(true).
						Build(),
					ui.Text(""),
					ui.CheckboxBuilder().
						Label("Review documentation").
						Checked(checked1).
						OnChange(setChecked1).
						Build(),
					ui.CheckboxBuilder().
						Label("Write tests").
						Checked(checked2).
						OnChange(setChecked2).
						Build(),
					ui.CheckboxBuilder().
						Label("Build release").
						Checked(checked3).
						OnChange(setChecked3).
						Build(),
				)
			}
		}(),
	)
}

func main() {
	err := ui.Run(DemoApp,
		ui.WithWidth(50),
		ui.WithHeight(20),
		ui.WithTitle("Mint UI Demo"),
	)
	if err != nil {
		panic(err)
	}
}
