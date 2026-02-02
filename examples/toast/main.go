package main

import (
	"github.com/wwsheng009/mint/app"
	"github.com/wwsheng009/mint/ui"
)

func main() {
	ui.Run(func() ui.VNode {
		// Track toast visibility - hooks must be called inside component
		infoToast, setInfoToast, _ := ui.UseStateInt(0)
		successToast, setSuccessToast, _ := ui.UseStateInt(0)
		warningToast, setWarningToast, _ := ui.UseStateInt(0)
		errorToast, setErrorToast, _ := ui.UseStateInt(0)
		// Build toast notifications based on state
		var toasts []ui.VNode

		if infoToast == 1 {
			toasts = append(toasts, app.ToastBuilder().
				Message("This is an info message").
				Info().
				Visible(true).
				Build())
		}
		if successToast == 1 {
			toasts = append(toasts, app.ToastBuilder().
				Message("Operation completed successfully!").
				Success().
				Visible(true).
				Build())
		}
		if warningToast == 1 {
			toasts = append(toasts, app.ToastBuilder().
				Message("Please check your input").
				Warning().
				Visible(true).
				Build())
		}
		if errorToast == 1 {
			toasts = append(toasts, app.ToastBuilder().
				Message("An error occurred!").
				Error().
				Visible(true).
				Build())
		}

		return app.VStack(
			app.NewTextBuilder("Toast Notifications Demo").Bold(true).FgColor("cyan").Build(),
			app.Text(""),
			app.Text("Click buttons below to show different toast types:"),
			app.Text(""),
			app.HStack(
				app.ButtonBuilder(" Info ").
					OnClick(func() {
						setInfoToast(1)
					}).
					Build(),
				app.ButtonBuilder(" Success ").
					OnClick(func() {
						setSuccessToast(1)
					}).
					Build(),
				app.ButtonBuilder(" Warning ").
					OnClick(func() {
						setWarningToast(1)
					}).
					Build(),
				app.ButtonBuilder(" Error ").
					OnClick(func() {
						setErrorToast(1)
					}).
					Build(),
			),
			app.Text(""),
			app.Text(""),
			app.HStack(
				app.ButtonBuilder(" Clear All ").
					OnClick(func() {
						setInfoToast(0)
						setSuccessToast(0)
						setWarningToast(0)
						setErrorToast(0)
					}).
					Build(),
			),
			app.Text(""),
			app.NewTextBuilder("────────────────────────────").FgColor("blue").Build(),
			app.Text(""),
			// Display toasts
			app.VStack(toasts...),
		)
	},
		ui.WithWidth(50),
		ui.WithHeight(20),
		ui.WithTitle("Toast Demo"),
	)
}
