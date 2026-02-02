package main

import (
	"github.com/wwsheng009/mint/app"
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
		app.NewTextBuilder("╔═══════════════════════════════════════╗").
			FgColor("cyan").
			Build(),
		app.NewTextBuilder("║     Mint UI Declarative Framework     ║").
			FgColor("cyan").
			Build(),
		app.NewTextBuilder("╚═══════════════════════════════════════╝").
			FgColor("cyan").
			Build(),
		ui.Text(""),
		// Tab navigation
		ui.HStack(
			app.ButtonBuilder(" [1] Counter ").
				OnClick(func() { setCurrentTab("counter") }).
				Build(),
			ui.Text(" "),
			app.ButtonBuilder(" [2] Input ").
				OnClick(func() { setCurrentTab("input") }).
				Build(),
			ui.Text(" "),
			app.ButtonBuilder(" [3] Tasks ").
				OnClick(func() { setCurrentTab("tasks") }).
				Build(),
		),
		ui.Text(""),
		app.NewTextBuilder("───────────────────────────────────────").
			FgColor("bright-black").
			Build(),
		ui.Text(""),
		// Content based on selected tab
		func() ui.VNode {
			if currentTab == "counter" {
				return ui.Fragment(
					app.NewTextBuilder("📊 Counter Demo").
						FgColor("yellow").
						Bold(true).
						Build(),
					ui.Text(""),
					app.NewTextBuilder("   Count:   ").
						FgColor("bright-black").
						Build(),
					ui.HStack(
						app.NewTextBuilder("   ").
							FgColor("green").
							Bold(true).
							Build(),
						app.NewTextBuilder("  ").
							FgColor("green").
							Bold(true).
							Build(),
						app.ButtonBuilder("  [ - ]  ").
							OnClick(func() { setCounter(counter - 1) }).
							Build(),
						ui.Text(" "),
						app.ButtonBuilder("  [ + ]  ").
							OnClick(func() { setCounter(counter + 1) }).
							Build(),
					),
				)
			} else if currentTab == "input" {
				return ui.Fragment(
					app.NewTextBuilder("📝 Input Demo").
						FgColor("yellow").
						Bold(true).
						Build(),
					ui.Text(""),
					ui.HStack(
						ui.Text("Name: "),
						app.InputBuilder().
							Value(text).
							Placeholder("Type here...").
							MaxLength(20).
							OnChange(setText).
							Build(),
					),
				)
			} else {
				return ui.Fragment(
					app.NewTextBuilder("✓ Task List").
						FgColor("yellow").
						Bold(true).
						Build(),
					ui.Text(""),
					app.CheckboxBuilder().
						Label("Review documentation").
						Checked(checked1).
						OnChange(setChecked1).
						Build(),
					app.CheckboxBuilder().
						Label("Write tests").
						Checked(checked2).
						OnChange(setChecked2).
						Build(),
					app.CheckboxBuilder().
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
