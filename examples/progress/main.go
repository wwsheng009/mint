package main

import (
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

	return ui.VStack(
		ui.NewTextBuilder("Progress Bar Demo").
			FgColor("cyan").
			Bold(true).
			Build(),
		ui.Text(""),
		// TODO: SpinnerBuilder 暂时不可用
		// ui.NewSpinnerBuilder().
		// 	Message("Loading demo...").
		// 	FgColor("yellow").
		// 	Build(),
		ui.NewTextBuilder("Spinner placeholder...").FgColor("yellow").Build(),
		ui.Text(""),
		ui.NewProgressBuilder().
			Label("Download:").
			Value(progress).
			Max(100).
			Width(30).
			ShowPercent(true).
			Build(),
		ui.Text(""),
		ui.NewTextBuilder("Status:").
			FgColor("bright-black").
			Build(),
		func() ui.VNode {
			if progress < 30 {
				return ui.NewTextBuilder("  Starting...").
					FgColor("bright-black").
					Build()
			} else if progress < 70 {
				return ui.NewTextBuilder("  In progress...").
					FgColor("yellow").
					Build()
			} else if progress < 100 {
				return ui.NewTextBuilder("  Almost done!").
					FgColor("cyan").
					Build()
			}
			return ui.NewTextBuilder("  Complete!").
				FgColor("green").
				Bold(true).
				Build()
		}(),
		ui.Text(""),
		ui.NewButtonBuilder("  +10%  ").
			OnPress(IncrementProgressIntent{}).
			Build(),
		ui.Text(""),
		ui.NewTextBuilder("Press button to increase progress").
			FgColor("bright-black").
			Build(),
		ui.Text(""),
		ui.NewTextBuilder("q: quit").
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
