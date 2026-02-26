// Package main demonstrates Transition Intent pattern for async operations.
//
// This example shows:
//   1. Using ShowPendingIntent to display loading states
//   2. Using CompleteTransitionIntent to update UI with async results
//   3. Understanding the async flow with pending/complete pattern
//
// Key Concept: Async operations are handled by showing a pending state,
// performing the operation in background, then emitting complete intent.
package main

import (
	"fmt"
	"time"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/ui"
	"github.com/wwsheng009/mint/ui/components/button"
)

// =============================================================================
// Async Operation Intents
// =============================================================================

// LoadDataIntent demonstrates an async data loading operation.
// When clicked, it starts a simulated async operation.
type LoadDataIntent struct {
	Source string // Where to load data from
}

func (LoadDataIntent) IntentType() string { return "LoadData" }

// =============================================================================
// Main Application
// =============================================================================

func App() ui.VNode {
	// We'll use a simple state management for this demo
	// In a real app, you'd use the proper State setters

	// LoadDataIntent handler
	ui.RegisterIntent(func(actx *intent.ActionContext, i LoadDataIntent) intent.IntentResult {
		fmt.Printf("[Async] Starting to load from %s...\n", i.Source)

		// Emit a ShowPendingIntent to indicate loading state
		// Note: In a full implementation, this would trigger UI updates
		// For this demo, we just simulate the flow

		// Simulate async work
		go func(source string) {
			fmt.Printf("[Background] Loading data from %s...\n", source)
			time.Sleep(1 * time.Second)

			// Emit CompleteTransitionIntent with results
			fmt.Printf("[Complete] Loaded data from %s\n", source)

			// In full implementation: would emit CompleteTransitionIntent here
		}(i.Source)

		return intent.HandledResult()
	})

	// ShowPendingIntent handler (would update UI with loading spinner)
	ui.RegisterIntent(func(actx *intent.ActionContext, i intent.ShowPendingIntent) intent.IntentResult {
		// Update state to show pending state
		actx.SetState("status", i.Label)
		actx.ScheduleUpdate()
		return intent.HandledResult()
	})

	// CompleteTransitionIntent handler (would update UI with results)
	ui.RegisterIntent(func(actx *intent.ActionContext, i intent.CompleteTransitionIntent) intent.IntentResult {
		actx.SetState("lastResult", fmt.Sprintf("Completed: %s", i.Name))
		actx.ScheduleUpdate()
		return intent.HandledResult()
	})

	// Build UI - simple static layout for demo
	loadButton := button.NewBuilder("Load Data").
		OnPress(LoadDataIntent{Source: "API Server"}).
		Variant(button.VariantPrimary).
		Build()

	return ui.VStack(
		ui.Text("=== Transition Intent Demo ==="),
		ui.Text(""),
		ui.Text("Click the button to trigger an async operation."),
		ui.Text(""),
		ui.Text("Pattern:"),
		ui.Text("  1. User clicks → LoadDataIntent emitted"),
		ui.Text("  2. Handler shows pending state"),
		ui.Text("  3. Background work runs"),
		ui.Text("  4. CompleteTransitionIntent updates UI"),
		ui.Text(""),
		loadButton,
	)
}

func main() {
	ui.Run(App,
		ui.WithWidth(60),
		ui.WithHeight(16),
		ui.WithTitle("Transition Demo"),
	)
}
