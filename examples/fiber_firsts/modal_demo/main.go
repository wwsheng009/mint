// Fiber-First Modal Component Demo - Interactive
// Demonstrates the new Modal component with full interactivity
package main

import (
	"github.com/wwsheng009/mint/framework"
	"github.com/wwsheng009/mint/runtime/intent"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui"
	"github.com/wwsheng009/mint/ui/components/modal"
	newtext "github.com/wwsheng009/mint/ui/components/text"
)

func main() {
	err := ui.Run(App,
		ui.WithWidth(70),
		ui.WithHeight(40),
		ui.WithTitle("Modal Demo - Interactive"),
		ui.WithPluginSetup(func(app *framework.App) {
			// Register modal support for ESC key and click-outside-to-close
			app.AddMiddleware(modal.NewModalMiddleware())
		}),
	)
	if err != nil {
		panic(err)
	}
}

// =============================================================================
// Custom Intent Types
// =============================================================================

type OpenModalIntent struct {
	ModalType string
}

func (OpenModalIntent) IntentType() string { return "OpenModal" }

type CloseModalIntent struct{}

func (CloseModalIntent) IntentType() string { return "CloseModal" }

// =============================================================================
// Main App
// =============================================================================

func App() ui.VNode {
	modalType, setModalType := ui.UseStateString("")

	// Register Intent handlers in App function
	rtui.RegisterIntent(func(ctx *intent.ActionContext, i OpenModalIntent) intent.IntentResult {
		setModalType(i.ModalType)
		return intent.HandledResult()
	})

	rtui.RegisterIntent(func(ctx *intent.ActionContext, i CloseModalIntent) intent.IntentResult {
		setModalType("")
		return intent.HandledResult()
	})

	return ui.VStack(
		ui.VStack(
			ui.NewTextBuilder("🎨 Modal Component Demo").
				Bold(true).
				FgColor("cyan").
				Build(),
			ui.Text(""),
			ui.NewTextBuilder("Interactive Modal Dialog System").
				FgColor("gray").
				Build(),
			ui.NewTextBuilder("ESCAPE key or click outside to close").
				FgColor("gray").
				Build(),
			ui.Text(""),
		),

		ui.NewTextBuilder("─").FgColor("gray").Build(),
		ui.Text(""),

		// Button Grid
		ui.HStack(
			ui.Text("  "),
			ui.NewButtonBuilder("  Basic  ").
				Variant(ui.ButtonVariantPrimary).
				OnPress(OpenModalIntent{ModalType: "basic"}).
				Disabled(modalType != "").
				Build(),
			ui.Text(" "),
			ui.NewButtonBuilder("  Border  ").
				Variant(ui.ButtonVariantSecondary).
				OnPress(OpenModalIntent{ModalType: "border"}).
				Disabled(modalType != "").
				Build(),
			ui.Text(" "),
			ui.NewButtonBuilder("  Footer  ").
				Variant(ui.ButtonVariantSecondary).
				OnPress(OpenModalIntent{ModalType: "footer"}).
				Disabled(modalType != "").
				Build(),
		),

		ui.Text(""),
		ui.HStack(
			ui.Text("  "),
			ui.NewButtonBuilder("  Alert  ").
				Variant(ui.ButtonVariantSecondary).
				OnPress(OpenModalIntent{ModalType: "alert"}).
				Disabled(modalType != "").
				Build(),
			ui.Text(" "),
			ui.NewButtonBuilder("  Sizes  ").
				Variant(ui.ButtonVariantSecondary).
				OnPress(OpenModalIntent{ModalType: "sizes"}).
				Disabled(modalType != "").
				Build(),
			ui.Text(" "),
			ui.NewButtonBuilder("  Locked  ").
				Variant(ui.ButtonVariantDanger).
				OnPress(OpenModalIntent{ModalType: "locked"}).
				Disabled(modalType != "").
				Build(),
		),

		ui.Text(""),
		ui.NewTextBuilder("─").FgColor("gray").Build(),
		ui.Text(""),

		// Status Display
		ui.HStack(
			ui.Text("  "),
			ui.NewTextBuilder("Status: ").FgColor("blue").Build(),
			ui.NewTextBuilder(getStatusText(modalType)).
				FgColor(getStatusColor(modalType)).
				Build(),
		),

		ui.Text(""),
		ui.NewTextBuilder("─").FgColor("gray").Build(),

		// Only render the active modal
		getModal(modalType),
	)
}

func getStatusText(modalType string) string {
	if modalType == "" {
		return "No modal open"
	}
	return "Modal: " + modalType
}

func getStatusColor(modalType string) string {
	if modalType == "" {
		return "gray"
	}
	return "green"
}

// getModal returns the modal component for the given type
func getModal(modalType string) ui.VNode {
	switch modalType {
	case "basic":
		// Add a close button in the modal content
		return modal.NewBuilder().
			Key("modal-basic").
			Title("Basic Modal").
			Content(ui.VStack(
				ui.NewTextBuilder("This is a basic modal dialog.").
					FgColor("white").
					Build(),
				ui.Text(""),
				ui.NewTextBuilder("Try pressing ESC").
					FgColor("gray").
					Build(),
				ui.NewTextBuilder("or clicking outside").
					FgColor("gray").
					Build(),
				ui.NewTextBuilder("or the button below").
					FgColor("gray").
					Build(),
				ui.Text(""),
				ui.NewButtonBuilder("  [Close]  ").
					Variant(ui.ButtonVariantPrimary).
					OnPress(CloseModalIntent{}).
					Build(),
			)).
			Width(45).
			Height(12).
			Rounded().
			Open(true).
			OnClose(CloseModalIntent{}).
			Build()

	case "border":
		return modal.NewBuilder().
			Key("modal-border").
			Title("Border Styles Demo").
			Content(newtext.New("Compare different border styles:\n\n• Single: clean and minimal\n• Double: bold and prominent\n• Rounded: friendly and modern\n• Dashed: subtle and light")).
			Width(50).
			Height(12).
			Rounded().
			Open(true).
			OnClose(CloseModalIntent{}).
			Build()

	case "footer":
		return modal.NewBuilder().
			Key("modal-footer").
			Title("Confirmation Dialog").
			Content(newtext.New("Are you sure you want to proceed?\nThis action cannot be undone.")).
			Footer(newtext.New("  [Esc] Cancel     [Enter] Confirm  ")).
			Width(40).
			Height(10).
			Rounded().
			Open(true).
			OnClose(CloseModalIntent{}).
			Build()

	case "alert":
		return modal.NewBuilder().
			Key("modal-alert").
			Title("Alert").
			Content(newtext.New("⚠️  Important notification!\n\nPlease review this message before continuing.")).
			Width(40).
			Height(10).
			Rounded().
			Open(true).
			OnClose(CloseModalIntent{}).
			Build()

	case "sizes":
		return modal.NewBuilder().
			Key("modal-sizes").
			Title("Modal Sizes").
			Content(newtext.New("Different modal sizes for different use cases:\n\n• Small: 25x6 - Quick messages\n• Medium: 30x8 - Standard dialogs\n• Large: 35x10 - Complex forms")).
			Width(40).
			Height(12).
			Rounded().
			Open(true).
			OnClose(CloseModalIntent{}).
			Build()

	case "locked":
		return modal.NewBuilder().
			Key("modal-locked").
			Title("⚠️  Critical Alert").
			Content(newtext.New("This modal is locked and cannot be closed.\n\nUsed for critical alerts that require user attention.\n\nClick outside or press ESC to simulate a system dismiss.")).
			Width(45).
			Height(10).
			Rounded().
			Closeable(true).
			Open(true).
			OnClose(CloseModalIntent{}).
			Build()

	default:
		return ui.Text("")
	}
}
