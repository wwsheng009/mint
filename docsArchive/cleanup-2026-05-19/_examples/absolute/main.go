// Absolute Demo demonstrates absolute positioning (Store + Reducer 架构)
package main

import (
	"fmt"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/reducer"
	"github.com/wwsheng009/mint/runtime/store"
	"github.com/wwsheng009/mint/ui"
	"github.com/wwsheng009/mint/ui/components/absolute"
	"github.com/wwsheng009/mint/ui/components/button"
)

// =============================================================================
// AppState (Single Source of Truth)
// =============================================================================

// AppState represents the absolute demo state.
type AppState struct {
	Count int
}

// =============================================================================
// Custom Intent Type
// =============================================================================

// IncrementIntent increments the count.
type IncrementIntent struct{}

func (IncrementIntent) IntentType() string { return "Increment" }
func (IncrementIntent) StayPressed() bool  { return true }

// =============================================================================
// Reducer (Pure Function)
// =============================================================================

// appReducer handles all state transitions.
var appReducer = reducer.NewBuilder[AppState]()

// Initialize the reducer.
func init() {
	// Handle IncrementIntent - increment the count
	appReducer.On(IncrementIntent{}, func(s AppState, i intent.Intent) AppState {
		s.Count++
		return s
	})
}

// =============================================================================
// Store (Single State Source)
// =============================================================================

// appStore holds the absolute demo state.
var appStore = store.NewStore(AppState{
	Count: 0,
})

// =============================================================================
// Main Function
// =============================================================================

func main() {
	// Register all handlers automatically
	appReducer.RegisterToGlobal(appStore)

	ui.Run(App,
		ui.WithWidth(50),
		ui.WithHeight(15),
		ui.WithTitle("Absolute Demo (Store+Reducer)"),
	)
}

// =============================================================================
// App Component
// =============================================================================

func App() ui.VNode {
	// Get current state snapshot
	state := appStore.Get()

	return ui.VStack(
		ui.NewTextBuilder("Absolute Positioning Demo").Bold(true).FgColor("cyan").Build(),
		ui.Text(""),
		ui.Text("Button with notification badge:"),
		ui.Text(""),
		ui.HStack(
			ui.NewButtonBuilder("  Messages  ").
				OnPress(IncrementIntent{}).
				Variant(button.VariantPrimary).
				Build(),
			// Badge positioned absolutely relative to parent
			ui.NewAbsoluteBuilder(
				ui.NewTextBuilder("New!").
					FgColor("red").
					BgColor("white").
					Bold(true).
					Build(),
			).
				Left(absolute.AbsolutePos(16)).
				Top(absolute.AbsolutePos(10)).
				Build(),
		),
		ui.Text(""),
		ui.NewTextBuilder("Stacked Elements").FgColor("yellow").Build(),
		ui.Text(""),
		ui.VStack(
			ui.Text("Background layer"),
			ui.HStack(
				ui.Text("Middle layer"),
				ui.NewAbsoluteBuilder(
					ui.NewTextBuilder("OVERLAY").FgColor("white").BgColor("red").Build(),
				).
					Left(absolute.AbsolutePos(10)).
					Top(absolute.AbsolutePos(5)).
					Build(),
			),
		),
		ui.Text(""),
		ui.NewTextBuilder(fmt.Sprintf("Click count: %d", state.Count)).Build(),
	)
}
