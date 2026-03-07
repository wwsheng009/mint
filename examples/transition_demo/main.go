// Package main demonstrates Transition Intent pattern for async operations.
//
// This example shows:
//   1. Using Intent to trigger async operations
//   2. Showing loading states while operation runs
//   3. Updating UI with results when operation completes
//
// Key Concept: Async operations follow this pattern:
//   1. User clicks → Intent emitted
//   2. Handler sets loading state → UI shows spinner
//   3. Background work runs in goroutine
//   4. Goroutine updates state with result → UI re-renders
package main

import (
	"fmt"
	"time"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/ui"
	buttoncomp "github.com/wwsheng009/mint/ui/components/button"
)

// =============================================================================
// Async Operation Intents
// =============================================================================

// LoadDataIntent starts an async data loading operation.
type LoadDataIntent struct {
	Source string // Where to load data from (e.g., "API", "Database", "File")
}

func (LoadDataIntent) IntentType() string { return "LoadData" }
func (LoadDataIntent) StayPressed() bool  { return true }

// =============================================================================
// Main Application Component
// =============================================================================

func App() ui.VNode {
	// 1. Define state using hooks (must be at top)
	isLoading, setIsLoading := ui.UseStateBool(false)
	lastResult, setLastResult := ui.UseStateString("")
	status, setStatus := ui.UseStateString("")

	// 2. Save setters to GlobalState for access from Intent handlers
	ctx := ui.GetCurrentContext()
	if ctx != nil {
		ctx.GlobalState["setIsLoading"] = setIsLoading
		ctx.GlobalState["setLastResult"] = setLastResult
		ctx.GlobalState["setStatus"] = setStatus
	}

	// 3. Register Intent handler using RegisterIntent to access intent parameters
	// We use RegisterIntent instead of ui.On because we need access to the intent's Source field
	ui.RegisterIntent(func(actx *intent.ActionContext, i LoadDataIntent) intent.IntentResult {
		source := i.Source

		// Step 1: Update UI to show loading state
		if fn, ok := actx.GetState("setIsLoading"); ok {
			if setter, ok := fn.(func(bool)); ok {
				setter(true)
			}
		}
		if fn, ok := actx.GetState("setStatus"); ok {
			if setter, ok := fn.(func(string)); ok {
				setter(fmt.Sprintf("Loading from %s...", source))
			}
		}

		// Step 2: Perform async work in background goroutine
		go func() {
			// Simulate async operation (e.g., API call, database query, file I/O)
			// Random duration between 1-3 seconds
			duration := time.Duration(1000 + (time.Now().UnixNano() % 2000))
			time.Sleep(duration)

			// Step 3: Update UI with result
			result := fmt.Sprintf("Data from %s (loaded in %.1fs)", source, duration.Seconds())

			if fn, ok := actx.GetState("setIsLoading"); ok {
				if setter, ok := fn.(func(bool)); ok {
					setter(false)
				}
			}
			if fn, ok := actx.GetState("setLastResult"); ok {
				if setter, ok := fn.(func(string)); ok {
					setter(result)
				}
			}
			if fn, ok := actx.GetState("setStatus"); ok {
				if setter, ok := fn.(func(string)); ok {
					setter("")
				}
			}
		}()

		return intent.HandledResult()
	})

	// 4. Get animation frame for loading spinner (alternates every 500ms)
	animationFrame := int(time.Now().UnixNano() / 500000000) % 4
	spinners := []string{"|", "/", "-", "\\"}
	spinChar := spinners[animationFrame]

	// 5. Build UI
	return ui.VStack(
		// Title
		ui.NewTextBuilder("╔══════════════════════════════════════╗").
			FgColor("cyan").
			Build(),
		ui.NewTextBuilder("║     Transition Intent Demo           ║").
			FgColor("cyan").
			Build(),
		ui.NewTextBuilder("╚══════════════════════════════════════╝").
			FgColor("cyan").
			Build(),
		ui.Text(""),

		// Description
		ui.NewTextBuilder("Async Operation Pattern").
			FgColor("yellow").
			Bold(true).
			Build(),
		ui.Text(""),
		ui.Text("Flow: Click → Loading → Background work → Result"),

		ui.Text(""),
		ui.NewTextBuilder("────────────────────────────────────").
			FgColor("bright-black").
			Build(),
		ui.Text(""),

		// Display loading spinner when isLoading
		func() ui.VNode {
			if isLoading {
				return ui.HStack(
					ui.NewTextBuilder("[").
						FgColor("yellow").
						Build(),
					ui.NewTextBuilder(spinChar).
						FgColor("yellow").
						Bold(true).
						Build(),
					ui.NewTextBuilder("Loading").
						FgColor("yellow").
						Build(),
					ui.Text("] "),
					ui.Text(status),
				)
			}
			return ui.HStack(
				ui.NewTextBuilder("[Ready] ").
					FgColor("green").
					Build(),
				ui.Text("Click a button to load data"),
			)
		}(),

		ui.Text(""),

		// Display last result
		func() ui.VNode {
			if lastResult != "" {
				return ui.HStack(
					ui.NewTextBuilder("✓ Result:").
						FgColor("bright-black").
						Bold(true).
						Build(),
					ui.Text(lastResult),
				)
			}
			return ui.Text("")
		}(),

		ui.Text(""),
		ui.NewTextBuilder("────────────────────────────────────").
			FgColor("bright-black").
			Build(),
		ui.Text(""),

		// Load buttons
		ui.NewTextBuilder("Trigger Async Operations:").
			FgColor("gray").
			Build(),
		ui.Text(""),
		ui.HStack(
			buttoncomp.NewBuilder("[API]").
				OnPress(LoadDataIntent{Source: "API Server"}).
				Variant(buttoncomp.VariantPrimary).
				Disabled(isLoading).
				Build(),
			ui.Text(" "),
			buttoncomp.NewBuilder("[DB]").
				OnPress(LoadDataIntent{Source: "Database"}).
				Variant(buttoncomp.VariantSecondary).
				Disabled(isLoading).
				Build(),
			ui.Text(" "),
			buttoncomp.NewBuilder("[File]").
				OnPress(LoadDataIntent{Source: "File System"}).
				Variant(buttoncomp.VariantSecondary).
				Disabled(isLoading).
				Build(),
		),
		ui.Text(""),
		ui.Text(""),
		ui.NewTextBuilder("[Tip] Buttons disabled during load").
			FgColor("gray").
			Build(),
	)
}

// =============================================================================
// Main Function
// =============================================================================

func main() {
	err := ui.Run(App,
		ui.WithWidth(60),
		ui.WithHeight(22),
		ui.WithTitle("Transition Demo"),
	)
	if err != nil {
		panic(err)
	}
}
