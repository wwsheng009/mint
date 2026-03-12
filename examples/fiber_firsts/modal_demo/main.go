// Fiber-First Modal Component Demo - Interactive
// Demonstrates the new Modal component with full interactivity
//
// Architecture: Store + Reducer + Custom Intent (Single Source of Truth)

package main

import (
	"github.com/wwsheng009/mint/framework"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/reducer"
	"github.com/wwsheng009/mint/runtime/store"
	"github.com/wwsheng009/mint/ui"
	"github.com/wwsheng009/mint/ui/components/modal"
	newtext "github.com/wwsheng009/mint/ui/components/text"
)

// =============================================================================
// AppState (Single Source of Truth)
// =============================================================================

// AppState represents the modal demo state.
type AppState struct {
	ModalType string
}

// =============================================================================
// Custom Intent Types
// =============================================================================

// OpenModalIntent opens a modal of the specified type.
type OpenModalIntent struct {
	ModalType string
}

func (OpenModalIntent) IntentType() string { return "OpenModal" }
func (OpenModalIntent) StayPressed() bool  { return true }

// CloseModalIntent closes the current modal.
type CloseModalIntent struct{}

func (CloseModalIntent) IntentType() string { return "CloseModal" }
func (CloseModalIntent) StayPressed() bool  { return false }

// =============================================================================
// Reducer (Pure Function)
// =============================================================================

// appReducer handles all state transitions.
var appReducer = reducer.NewBuilder[AppState]()

// Initialize the reducer.
func init() {
	// Handle OpenModalIntent
	appReducer.On(OpenModalIntent{}, func(s AppState, i intent.Intent) AppState {
		omi := i.(OpenModalIntent)
		s.ModalType = omi.ModalType
		return s
	})

	// Handle CloseModalIntent
	appReducer.On(CloseModalIntent{}, func(s AppState, i intent.Intent) AppState {
		s.ModalType = ""
		return s
	})
}

// =============================================================================
// Store (Single State Source)
// =============================================================================

// appStore holds the modal demo state.
var appStore = store.NewStore(AppState{
	ModalType: "",
})

// =============================================================================
// Main Entry Point
// =============================================================================

func main() {
	// Register reducer handlers to store
	appReducer.RegisterToGlobal(appStore)

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
// Main App
// =============================================================================

func App() ui.VNode {
	state := appStore.Get()

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
			ui.NewTextBuilder("Styled header, shadow, padding, and close policies").
				FgColor("gray").
				Build(),
			ui.NewTextBuilder("Background clicks are blocked by the modal middleware while open.").
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
				Build(),
			ui.Text(" "),
			ui.NewButtonBuilder("  Border  ").
				Variant(ui.ButtonVariantSecondary).
				OnPress(OpenModalIntent{ModalType: "border"}).
				Build(),
			ui.Text(" "),
			ui.NewButtonBuilder("  Footer  ").
				Variant(ui.ButtonVariantSecondary).
				OnPress(OpenModalIntent{ModalType: "footer"}).
				Build(),
			ui.Text(" "),
			ui.NewButtonBuilder("  Padded  ").
				Variant(ui.ButtonVariantSecondary).
				OnPress(OpenModalIntent{ModalType: "padded"}).
				Build(),
		),

		ui.Text(""),
		ui.HStack(
			ui.Text("  "),
			ui.NewButtonBuilder("  Alert  ").
				Variant(ui.ButtonVariantSecondary).
				OnPress(OpenModalIntent{ModalType: "alert"}).
				Build(),
			ui.Text(" "),
			ui.NewButtonBuilder("  Sizes  ").
				Variant(ui.ButtonVariantSecondary).
				OnPress(OpenModalIntent{ModalType: "sizes"}).
				Build(),
			ui.Text(" "),
			ui.NewButtonBuilder("  Sticky  ").
				Variant(ui.ButtonVariantDanger).
				OnPress(OpenModalIntent{ModalType: "sticky"}).
				Build(),
		),

		ui.Text(""),
		ui.NewTextBuilder("─").FgColor("gray").Build(),
		ui.Text(""),

		// Status Display
		ui.HStack(
			ui.Text("  "),
			ui.NewTextBuilder("Status: ").FgColor("blue").Build(),
			ui.NewTextBuilder(getStatusText(state.ModalType)).
				FgColor(getStatusColor(state.ModalType)).
				Build(),
		),

		ui.Text(""),
		ui.NewTextBuilder("─").FgColor("gray").Build(),

		// Only render the active modal
		getModal(state.ModalType),
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
			Padding(1).
			Content(ui.VStack(
				ui.NewTextBuilder("This is a basic modal dialog.").
					FgColor("white").
					Build(),
				ui.Text(""),
				ui.NewTextBuilder("Try pressing ESC, clicking outside,").
					FgColor("gray").
					Build(),
				ui.NewTextBuilder("or the button below.").
					FgColor("gray").
					Build(),
				ui.Text(""),
				ui.NewButtonBuilder("  [Close]  ").
					Variant(ui.ButtonVariantPrimary).
					OnPress(CloseModalIntent{}).
					Build(),
			)).
			InnerSize(39, 6).
			Rounded().
			Open(true).
			OnClose(CloseModalIntent{}).
			Build()

	case "border":
		return modal.NewBuilder().
			Key("modal-border").
			Title("Border Styles Demo").
			Padding(1).
			Content(newtext.New("Compare different border styles:\n\n- Single: clean and minimal\n- Double: bold and prominent\n- Rounded: friendly and modern\n- Dashed: subtle and light")).
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
			Padding(1).
			Content(newtext.New("Are you sure you want to proceed?\nThis action cannot be undone.")).
			Footer(
				ui.HStack(
					ui.NewTextBuilder("Press Enter on a button to continue.").FgColor("gray").Build(),
					ui.Spacer().Flex(1).Build(),
					ui.NewButtonBuilder(" Cancel ").
						Variant(ui.ButtonVariantSecondary).
						OnPress(CloseModalIntent{}).
						Build(),
					ui.Text(" "),
					ui.NewButtonBuilder(" Confirm ").
						Variant(ui.ButtonVariantPrimary).
						OnPress(CloseModalIntent{}).
						Build(),
				),
			).
			InnerSize(38, 5).
			Double().
			Open(true).
			OnClose(CloseModalIntent{}).
			Build()

	case "padded":
		return modal.NewBuilder().
			Key("modal-padded").
			Title("Padded Surface").
			Padding(2).
			Content(ui.VStack(
				ui.NewTextBuilder("This modal uses InnerSize + Padding.").FgColor("white").Build(),
				ui.Text(""),
				ui.NewTextBuilder("The title is rendered in a dedicated header row,").FgColor("gray").Build(),
				ui.NewTextBuilder("and the body gets a filled surface background.").FgColor("gray").Build(),
				ui.Text(""),
				ui.NewButtonBuilder(" Close ").Variant(ui.ButtonVariantPrimary).OnPress(CloseModalIntent{}).Build(),
			)).
			InnerSize(36, 6).
			Shadow(true).
			Rounded().
			Open(true).
			OnClose(CloseModalIntent{}).
			Build()

	case "alert":
		return modal.NewBuilder().
			Key("modal-alert").
			Title("Alert").
			Padding(1).
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
			Padding(1).
			Content(newtext.New("Different modal sizes for different use cases:\n\n- Small: quick messages\n- Medium: standard dialogs\n- Large: multi-step content")).
			Width(40).
			Height(12).
			Rounded().
			Open(true).
			OnClose(CloseModalIntent{}).
			Build()

	case "sticky":
		return modal.NewBuilder().
			Key("modal-sticky").
			Title("Sticky Modal").
			Padding(1).
			Content(ui.VStack(
				ui.NewTextBuilder("ESC and backdrop closing are disabled here.").FgColor("white").Build(),
				ui.Text(""),
				ui.NewTextBuilder("Background controls stay blocked while the modal is open.").FgColor("gray").Build(),
				ui.NewTextBuilder("Use the footer action to dismiss it.").FgColor("gray").Build(),
			)).
			Footer(
				ui.HStack(
					ui.NewTextBuilder("Use the button to close this dialog.").FgColor("gray").Build(),
					ui.Spacer().Flex(1).Build(),
					ui.NewButtonBuilder(" Acknowledge ").
						Variant(ui.ButtonVariantPrimary).
						OnPress(CloseModalIntent{}).
						Build(),
				),
			).
			InnerSize(52, 5).
			Rounded().
			CloseOnEsc(false).
			CloseOnBackdrop(false).
			Open(true).
			OnClose(CloseModalIntent{}).
			Build()

	default:
		return ui.Text("")
	}
}
