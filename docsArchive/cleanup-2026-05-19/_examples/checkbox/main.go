// Checkbox Demo demonstrates checkbox component (Store + Reducer 架构)
package main

import (
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/reducer"
	"github.com/wwsheng009/mint/runtime/store"
	"github.com/wwsheng009/mint/ui"
)

// =============================================================================
// AppState (Single Source of Truth)
// =============================================================================

// AppState represents the checkbox demo state.
type AppState struct {
	AcceptTerms   bool
	AcceptUpdates bool
	AcceptPrivacy bool
}

// =============================================================================
// Reducer (Pure Function)
// =============================================================================

// appReducer handles all state transitions for checkboxes.
var appReducer = reducer.NewBuilder[AppState]()

// Initialize the reducer - ForField components emit FieldChangeIntent automatically
func init() {
	// Handle FieldChangeIntent - update checkbox state
	appReducer.On(intent.FieldChangeIntent{}, func(s AppState, i intent.Intent) AppState {
		fci := i.(intent.FieldChangeIntent)
		value := fci.Value == "true"
		switch fci.Field {
		case "acceptTerms":
			s.AcceptTerms = value
		case "acceptUpdates":
			s.AcceptUpdates = value
		case "acceptPrivacy":
			s.AcceptPrivacy = value
		}
		return s
	})
}

// =============================================================================
// Store (Single State Source)
// =============================================================================

// appStore holds the checkbox demo state.
var appStore = store.NewStore(AppState{
	AcceptTerms:   false,
	AcceptUpdates: false,
	AcceptPrivacy: false,
})

// =============================================================================
// Checkbox Demo Component
// =============================================================================

func CheckboxDemo() ui.VNode {
	// Get current state snapshot
	state := appStore.Get()

	// Count checked checkboxes
	checkedCount := 0
	if state.AcceptTerms {
		checkedCount++
	}
	if state.AcceptUpdates {
		checkedCount++
	}
	if state.AcceptPrivacy {
		checkedCount++
	}

	return ui.VStack(
		ui.NewTextBuilder("Checkbox Demo").
			FgColor("cyan").
			Bold(true).
			Build(),
		ui.Text(""),
		ui.NewTextBuilder("Select your preferences:").
			FgColor("bright-black").
			Build(),
		ui.Text(""),
		ui.NewCheckboxBuilder().
			ForField(intent.BindField("acceptTerms")).
			Label("I accept the terms and conditions").
			Checked(state.AcceptTerms).
			Build(),
		ui.NewCheckboxBuilder().
			ForField(intent.BindField("acceptUpdates")).
			Label("Subscribe to updates").
			Checked(state.AcceptUpdates).
			Build(),
		ui.NewCheckboxBuilder().
			ForField(intent.BindField("acceptPrivacy")).
			Label("I have read the privacy policy").
			Checked(state.AcceptPrivacy).
			Build(),
		ui.Text(""),
		ui.NewTextBuilder("Checked: 1/3").
			FgColor("yellow").
			Build(),
		ui.Text(""),
		ui.NewTextBuilder("Tab: focus | Space/Enter: toggle | q: quit").
			FgColor("bright-black").
			Build(),
	)
}

func main() {
	// Register all handlers automatically
	appReducer.RegisterToGlobal(appStore)

	err := ui.Run(CheckboxDemo,
		ui.WithWidth(50),
		ui.WithHeight(18),
		ui.WithTitle("Checkbox Demo (Store+Reducer)"),
	)
	if err != nil {
		panic(err)
	}
}
