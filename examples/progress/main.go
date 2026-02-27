package main

import (
	"github.com/wwsheng009/mint/app"
	"github.com/wwsheng009/mint/ui"
)

// Intent Types
type IncrementProgressIntent struct{}
func (IncrementProgressIntent) IntentType() string { return "IncrementProgress" }
func (IncrementProgressIntent) StayPressed() bool  { return true }

// ProgressDemo demonstrates the progress bar component
func ProgressDemo() ui.VNode {
	progress, setProgress, _ := ui.UseStateInt(0)

	// Register intent handler
	ui.On(IncrementProgressIntent{}, func() {
		if progress >= 100 {
			return
		}
		setProgress(progress + 10)
	})

	return app.VStack(
		app.NewTextBuilder("Progress Bar Demo").
			FgColor("cyan").
			Bold(true).
			Build(),
		app.Text(""),
		// TODO: SpinnerBuilder 暂时不可用
		// app.SpinnerBuilder().
		// 	Message("Loading demo...").
		// 	FgColor("yellow").
		// 	Build(),
		app.NewTextBuilder("Spinner placeholder...").FgColor("yellow").Build(),
		app.Text(""),
		app.ProgressBuilder().
			Label("Download:").
			Value(progress).
			Max(100).
			Width(30).
			ShowPercent(true).
			Build(),
		app.Text(""),
		app.NewTextBuilder("Status:").
			FgColor("bright-black").
			Build(),
		func() ui.VNode {
			if progress < 30 {
				return app.NewTextBuilder("  Starting...").
					FgColor("bright-black").
					Build()
			} else if progress < 70 {
				return app.NewTextBuilder("  In progress...").
					FgColor("yellow").
					Build()
			} else if progress < 100 {
				return app.NewTextBuilder("  Almost done!").
					FgColor("cyan").
					Build()
			}
			return app.NewTextBuilder("  Complete!").
				FgColor("green").
				Bold(true).
				Build()
		}(),
		app.Text(""),
		app.ButtonBuilder("  +10%  ").
			OnPress(IncrementProgressIntent{}).
			Build(),
		app.Text(""),
		app.NewTextBuilder("Press button to increase progress").
			FgColor("bright-black").
			Build(),
		app.Text(""),
		app.NewTextBuilder("q: quit").
			FgColor("bright-black").
			Build(),
	)
}

func main() {
	err := ui.Run(ProgressDemo,
		ui.WithWidth(50),
		ui.WithHeight(20),
		ui.WithTitle("Progress Demo"),
	)
	if err != nil {
		panic(err)
	}
}
