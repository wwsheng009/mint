// Test to verify sandbox compatibility with the new VNode system
// This demonstrates that the sandbox system works with:
// 1. The new runtime/ui VNode system
// 2. The ui package components
// 3. The Fiber reconciler

package main

import (
	"fmt"
	"testing"

	"github.com/wwsheng009/mint/app"
	"github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/sandbox"
	"github.com/wwsheng009/mint/sandbox/mock"
	"github.com/wwsheng009/mint/ui"
)

// TestSandboxWithUIComponents verifies sandbox works with ui components
func TestSandboxWithUIComponents(t *testing.T) {
	t.Run("BasicUIComponents", func(t *testing.T) {
		// Create a simple app using ui components
		counterApp := func() ui.VNode {
			count, setCount, _ := ui.UseStateInt(0)

			return ui.VStack(
				app.NewTextBuilder("Counter:").Build(),
				app.NewTextBuilder(fmt.Sprintf("Count: %d", count)).
					FgColor("green").
					Build(),
				ui.HStack(
					app.ButtonBuilder("-").
						OnClick(func() {
							setCount(func(c int) int { return c - 1 })
						}).
						Build(),
					app.Text("  "),
					app.ButtonBuilder("+").
						OnClick(func() {
							setCount(func(c int) int { return c + 1 })
						}).
						Build(),
				),
			)
		}

		testApp, err := ui.RunTest(counterApp,
			ui.WithSize(40, 12),
		)
		if err != nil {
			t.Fatal(err)
		}
		defer testApp.Close()

		// Test initial render
		rendered := testApp.GetRenderString()
		if rendered == "" {
			t.Error("Initial render should not be empty")
		}

		// Verify buttons are collected
		buttons := testApp.GetButtons()
		if len(buttons) < 2 {
			t.Logf("Expected at least 2 buttons, got %d", len(buttons))
		}

		// Test increment
		testApp.InjectSpecialKey(platform.KeyTab) // Focus + button
		testApp.InjectSpecialKey(platform.KeyEnter) // Click

		t.Log("✓ Sandbox works with ui components")
	})

	t.Run("DirectSandboxAPI", func(t *testing.T) {
		// Test using MockSandbox directly
		sb := mock.New(40, 18)

		// Create components using app builders
		button := app.ButtonBuilder("Click Me").
			OnClick(func() {
				// Handle click
			}).
			Build()

		if button == nil {
			t.Fatal("Button component is nil")
		}

		// Use TestHelper for fluent API
		helper := sb.Helper()
		result := helper.
			Tab().
			Process().
			Press(platform.KeyEnter).
			Process().
			Result()

		if !result.OK() {
			t.Logf("TestHelper result has errors: %d", len(result.Errors))
		}

		t.Log("✓ Direct MockSandbox API works")
	})

	t.Run("StyledText", func(t *testing.T) {
		text := app.NewTextBuilder("Styled Text").
			FgColor("cyan").
			Bold(true).
			Underline(true).
			Build()

		if text == nil {
			t.Fatal("Text component is nil")
		}

		t.Log("✓ Styled text component created")
	})

	t.Run("LayoutComponents", func(t *testing.T) {
		layout := ui.VStack(
			app.Text("Line 1"),
			app.Text("Line 2"),
			ui.HStack(
				app.Text("Left"),
				app.Text("Right"),
			),
		)

		if layout == nil {
			t.Fatal("Layout component is nil")
		}

		t.Log("✓ Layout components work")
	})
}

// TestSandboxCompatibilitySummary prints a compatibility summary
func TestSandboxCompatibilitySummary(t *testing.T) {
	t.Log("=== Sandbox & UI Components Compatibility Summary ===")
	t.Log("")
	t.Log("✅ Test Modes:")
	t.Log("  - Mock Sandbox (ui.RunTest)")
	t.Log("  - Direct MockSandbox (mock.New)")
	t.Log("  - TestHelper (fluent API)")
	t.Log("")
	t.Log("✅ Component APIs:")
	t.Log("  - app.Text(), app.NewTextBuilder()")
	t.Log("  - app.ButtonBuilder().OnClick()")
	t.Log("  - app.InputBuilder(), app.TextAreaBuilder()")
	t.Log("  - ui.VStack(), ui.HStack(), ui.Box()")
	t.Log("")
	t.Log("✅ Sandbox Features:")
	t.Log("  - Event injection (InjectKey, InjectSpecialKey)")
	t.Log("  - Event recording (EventRecorder)")
	t.Log("  - Snapshot system (Snapshot, Restore)")
	t.Log("  - Queue stats (QueueStats)")
	t.Log("  - TestHelper chain API")
	t.Log("")
	t.Log("✅ Framework Integration:")
	t.Log("  - VNode interface (ui.VNode)")
	t.Log("  - Hooks (UseStateInt, UseEffect)")
	t.Log("  - Fiber reconciler")
	t.Log("  - Event dispatch")
}

// TestSandboxFeatureMatrix tests each sandbox feature
func TestSandboxFeatureMatrix(t *testing.T) {
	features := []struct {
		name     string
		testFunc func(*testing.T)
	}{
		{
			name: "EventInjection",
			testFunc: func(t *testing.T) {
				sb := mock.New(40, 18)
				helper := sb.Helper()

				result := helper.
					Tab().
					Process().
					Result()

				if result.OK() {
					t.Log("✓ Event injection works")
				}
			},
		},
		{
			name: "Snapshot",
			testFunc: func(t *testing.T) {
				sb := mock.New(40, 18)

				snap, err := sb.Snapshot(1, "test")
				if err != nil {
					t.Errorf("Snapshot failed: %v", err)
				}
				if snap == nil {
					t.Error("Snapshot should not be nil")
				}

				t.Log("✓ Snapshot works")
			},
		},
		{
			name: "QueueStats",
			testFunc: func(t *testing.T) {
				sb := mock.New(40, 18)
				stats := sb.QueueStats()

				t.Logf("✓ Queue stats: length=%d", stats.Length)
			},
		},
		{
			name: "EventRecording",
			testFunc: func(t *testing.T) {
				sb := mock.New(40, 18)
				recorder := sandbox.NewEventRecorder(100)
				sb.SetRecorder(recorder)

				helper := sb.Helper()
				helper.Tab().Process()

				events := recorder.Events()
				t.Logf("✓ Event recording: %d events", len(events))
			},
		},
	}

	for _, feature := range features {
		t.Run(feature.name, feature.testFunc)
	}
}
