package main

import (
	"github.com/wwsheng009/mint/ui"
)

// ProgressDemo demonstrates the progress bar component
func ProgressDemo() ui.VNode {
	progress, setProgress, _ := ui.UseStateInt(0)

	return ui.VStack(
		ui.NewTextBuilder("Progress Bar Demo").
			FgColor("cyan").
			Bold(true).
			Build(),
		ui.Text(""),
		ui.SpinnerBuilder().
			Message("Loading demo...").
			FgColor("yellow").
			Build(),
		ui.Text(""),
		ui.ProgressBuilder().
			Label("Download:").
			Value(progress).
			Max(100).
			Width(30).
			ShowPercent(true).
			FgColor("green").
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
		ui.ButtonBuilder("  +10%  ").
			OnClick(func() {
				if progress >= 100 {
					return
				}
				setProgress(progress + 10)
			}).
			Disabled(progress >= 100).
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
