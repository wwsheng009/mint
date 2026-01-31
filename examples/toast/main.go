package main

import (
	"github.com/wwsheng009/mint/ui"
)

func main() {
	// Track toast visibility
	infoToast, setInfoToast, _ := ui.UseStateInt(0)
	successToast, setSuccessToast, _ := ui.UseStateInt(0)
	warningToast, setWarningToast, _ := ui.UseStateInt(0)
	errorToast, setErrorToast, _ := ui.UseStateInt(0)

	ui.Run(func() ui.VNode {
		// Build toast notifications based on state
		var toasts []ui.VNode

		if infoToast == 1 {
			toasts = append(toasts, ui.ToastBuilder().
				Message("This is an info message").
				Info().
				Visible(true).
				Build())
		}
		if successToast == 1 {
			toasts = append(toasts, ui.ToastBuilder().
				Message("Operation completed successfully!").
				Success().
				Visible(true).
				Build())
		}
		if warningToast == 1 {
			toasts = append(toasts, ui.ToastBuilder().
				Message("Please check your input").
				Warning().
				Visible(true).
				Build())
		}
		if errorToast == 1 {
			toasts = append(toasts, ui.ToastBuilder().
				Message("An error occurred!").
				Error().
				Visible(true).
				Build())
		}

		return ui.VStack(
			ui.NewTextBuilder("Toast Notifications Demo").Bold(true).FgColor("cyan").Build(),
			ui.Text(""),
			ui.Text("Click buttons below to show different toast types:"),
			ui.Text(""),
			ui.HStack(
				ui.ButtonBuilder(" Info ").
					OnClick(func() {
						setInfoToast(1)
					}).
					Build(),
				ui.ButtonBuilder(" Success ").
					OnClick(func() {
						setSuccessToast(1)
					}).
					Build(),
				ui.ButtonBuilder(" Warning ").
					OnClick(func() {
						setWarningToast(1)
					}).
					Build(),
				ui.ButtonBuilder(" Error ").
					OnClick(func() {
						setErrorToast(1)
					}).
					Build(),
			),
			ui.Text(""),
			ui.Text(""),
			ui.HStack(
				ui.ButtonBuilder(" Clear All ").
					OnClick(func() {
						setInfoToast(0)
						setSuccessToast(0)
						setWarningToast(0)
						setErrorToast(0)
					}).
					Build(),
			),
			ui.Text(""),
			ui.NewTextBuilder("────────────────────────────").FgColor("blue").Build(),
			ui.Text(""),
			// Display toasts
			ui.VStack(toasts...),
		)
	},
		ui.WithWidth(50),
		ui.WithHeight(20),
		ui.WithTitle("Toast Demo"),
	)
}
