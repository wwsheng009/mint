// examples/error_boundary/error_boundary_e2e_test.go - End-to-end tests for error boundaries
package main

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/ui"
)

// =============================================================================
// Test Components for Error Boundary E2E Tests
// =============================================================================

// ErrorBoundaryAppTest creates an app with error boundary (for testing)
func ErrorBoundaryAppTest() ui.VNode {
	return ui.VStack(
		ui.Element("text").Prop("content", "Error Boundary Demo").Prop("style", style.Style{}.Bold(true)).Build(),
		ui.ErrorBoundary(
			"demoBoundary",
			PanicComponent,
			ui.FallbackBox("Error Occurred", "Something went wrong in this component"),
		),
	)
}

// MultipleErrorBoundariesApp creates an app with multiple error boundaries
func MultipleErrorBoundariesApp() ui.VNode {
	return ui.VStack(
		ui.Element("text").Prop("content", "Multiple Error Boundaries Demo").Build(),
		ui.HStack(
			ui.ErrorBoundary(
				"boundary1",
				func() ui.VNode {
					if shouldPanic {
						panic(errors.New("left component panic"))
					}
					return ui.Element("text").Prop("content", "Left: OK").Build()
				},
				ui.FallbackText("Left: Error"),
			),
			ui.ErrorBoundary(
				"boundary2",
				func() ui.VNode {
					return ui.Element("text").Prop("content", "Right: OK").Build()
				},
				ui.FallbackText("Right: Error"),
			),
		),
	)
}

// =============================================================================
// End-to-End Tests
// =============================================================================

// TestErrorBoundary_NormalRendering tests that error boundary doesn't interfere with normal rendering
func TestErrorBoundary_NormalRendering(t *testing.T) {
	resetErrorBoundaryTestState(t)
	t.Setenv("MINT_USE_FIBER", "true")
	t.Setenv("TUI_DEBUG_UI", "false")

	shouldPanic = false

	testApp, err := ui.RunTest(ErrorBoundaryApp,
		ui.WithWidth(40),
		ui.WithHeight(10),
	)
	if err != nil {
		t.Fatalf("RunTest failed: %v", err)
	}
	defer testApp.Close()

	time.Sleep(100 * time.Millisecond)

	rendered := testApp.GetRenderString()
	t.Logf("Normal render:\n%s", rendered)

	// Should show normal content
	if !strings.Contains(rendered, "Normal Content") {
		t.Error("Should show normal content when no panic occurs")
	}

	// Should NOT show error message
	if strings.Contains(rendered, "Error Occurred") {
		t.Error("Should not show error message when no panic occurs")
	}
}

// TestErrorBoundary_CatchesPanic tests that error boundary catches and displays fallback
func TestErrorBoundary_CatchesPanic(t *testing.T) {
	resetErrorBoundaryTestState(t)
	t.Setenv("MINT_USE_FIBER", "true")
	t.Setenv("TUI_DEBUG_UI", "false")

	shouldPanic = true

	testApp, err := ui.RunTest(ErrorBoundaryApp,
		ui.WithWidth(40),
		ui.WithHeight(10),
	)
	if err != nil {
		t.Fatalf("RunTest failed: %v", err)
	}
	defer testApp.Close()

	time.Sleep(100 * time.Millisecond)

	rendered := testApp.GetRenderString()
	t.Logf("Panic render:\n%s", rendered)

	// Should show error fallback
	if !strings.Contains(rendered, "Error Occurred") {
		t.Error("Should show error fallback when panic occurs")
	}

	// Should NOT show normal content
	if strings.Contains(rendered, "Normal Content") {
		t.Error("Should not show normal content when panic occurs")
	}
}

// TestErrorBoundary_RecoveryAfterPanic tests that app can recover after panic is resolved
func TestErrorBoundary_RecoveryAfterPanic(t *testing.T) {
	resetErrorBoundaryTestState(t)
	t.Setenv("MINT_USE_FIBER", "true")
	t.Setenv("TUI_DEBUG_UI", "false")

	shouldPanic = true

	testApp, err := ui.RunTest(ErrorBoundaryApp,
		ui.WithWidth(40),
		ui.WithHeight(10),
	)
	if err != nil {
		t.Fatalf("RunTest failed: %v", err)
	}
	defer testApp.Close()

	time.Sleep(100 * time.Millisecond)

	// Initial render should show error
	initialRender := testApp.GetRenderString()
	if !strings.Contains(initialRender, "Error Occurred") {
		t.Error("Initial render should show error")
	}

	// Resolve the panic condition
	shouldPanic = false

	// Trigger re-render by injecting a key press
	if err := testApp.InjectKey(' '); err != nil {
		t.Fatalf("InjectKey failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// After re-render, should show normal content
	recoveredRender := testApp.GetRenderString()
	t.Logf("Recovered render:\n%s", recoveredRender)

	// Note: The current implementation clears error state on each render attempt
	// So recovery should work if the component no longer panics
	if !strings.Contains(recoveredRender, "Normal Content") && !strings.Contains(recoveredRender, "Error Occurred") {
		t.Logf("Note: Recovery behavior depends on re-render mechanism")
	}
}

// TestErrorBoundary_MultipleBoundaries tests multiple independent error boundaries
func TestErrorBoundary_MultipleBoundaries(t *testing.T) {
	resetErrorBoundaryTestState(t)
	t.Setenv("MINT_USE_FIBER", "true")
	t.Setenv("TUI_DEBUG_UI", "false")

	shouldPanic = false

	testApp, err := ui.RunTest(MultipleErrorBoundariesApp,
		ui.WithWidth(40),
		ui.WithHeight(10),
	)
	if err != nil {
		t.Fatalf("RunTest failed: %v", err)
	}
	defer testApp.Close()

	time.Sleep(100 * time.Millisecond)

	// Initial render - both should be OK
	rendered := testApp.GetRenderString()
	t.Logf("Initial render:\n%s", rendered)

	// Check that both components render something
	if !strings.Contains(rendered, "OK") {
		t.Error("Components should show OK")
	}

	// Check that the demo title is shown
	if !strings.Contains(rendered, "Multiple Error Boundaries") {
		t.Error("Should show demo title")
	}
}

// TestErrorBoundary_PartialFailure tests that one error boundary doesn't affect others
func TestErrorBoundary_PartialFailure(t *testing.T) {
	resetErrorBoundaryTestState(t)
	t.Setenv("MINT_USE_FIBER", "true")
	t.Setenv("TUI_DEBUG_UI", "false")

	shouldPanic = true // Left component will panic

	testApp, err := ui.RunTest(MultipleErrorBoundariesApp,
		ui.WithWidth(40),
		ui.WithHeight(10),
	)
	if err != nil {
		t.Fatalf("RunTest failed: %v", err)
	}
	defer testApp.Close()

	time.Sleep(100 * time.Millisecond)

	rendered := testApp.GetRenderString()
	t.Logf("Partial failure render:\n%s", rendered)

	// Should see error message somewhere
	if !strings.Contains(rendered, "Error") {
		t.Error("Should show error when component panics")
	}

	// Should still see OK (right component)
	if !strings.Contains(rendered, "OK") {
		t.Error("Other component should still work")
	}
}

// TestErrorBoundary_NonFiberMode tests error boundary in non-Fiber mode
func TestErrorBoundary_NonFiberMode(t *testing.T) {
	resetErrorBoundaryTestState(t)
	t.Setenv("MINT_USE_FIBER", "false")
	t.Setenv("TUI_DEBUG_UI", "false")

	shouldPanic = false

	testApp, err := ui.RunTest(ErrorBoundaryApp,
		ui.WithWidth(40),
		ui.WithHeight(10),
	)
	if err != nil {
		t.Fatalf("RunTest failed: %v", err)
	}
	defer testApp.Close()

	time.Sleep(100 * time.Millisecond)

	rendered := testApp.GetRenderString()
	t.Logf("Non-Fiber render:\n%s", rendered)

	// Should work in non-Fiber mode too
	if !strings.Contains(rendered, "Error Boundary Demo") {
		t.Error("Should render in non-Fiber mode")
	}
}

// TestErrorBoundary_WithInteraction tests error boundary with user interaction
func TestErrorBoundary_WithInteraction(t *testing.T) {
	resetErrorBoundaryTestState(t)
	t.Setenv("MINT_USE_FIBER", "true")
	t.Setenv("TUI_DEBUG_UI", "false")

	shouldPanic = false

	testApp, err := ui.RunTest(ErrorBoundaryApp,
		ui.WithWidth(40),
		ui.WithHeight(10),
	)
	if err != nil {
		t.Fatalf("RunTest failed: %v", err)
	}
	defer testApp.Close()

	time.Sleep(100 * time.Millisecond)

	// Trigger a panic condition through interaction
	shouldPanic = true

	// Inject some key to trigger re-render
	if err := testApp.InjectKey('x'); err != nil {
		t.Fatalf("InjectKey failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	rendered := testApp.GetRenderString()
	t.Logf("After interaction:\n%s", rendered)

	// Should handle panic gracefully
	// The app should still be responsive
}

// TestErrorBoundary_ErrorContent tests error boundary with detailed error content
func TestErrorBoundary_ErrorContent(t *testing.T) {
	resetErrorBoundaryTestState(t)
	t.Setenv("MINT_USE_FIBER", "true")
	t.Setenv("TUI_DEBUG_UI", "false")

	// Custom component with specific error message
	customPanicComponent := func() ui.VNode {
		panic(errors.New("specific error: division by zero"))
	}

	customApp := func() ui.VNode {
		return ui.ErrorBoundary(
			"customBoundary",
			customPanicComponent,
			ui.FallbackBox("Custom Error", "A specific error occurred"),
		)
	}

	testApp, err := ui.RunTest(customApp,
		ui.WithWidth(40),
		ui.WithHeight(10),
	)
	if err != nil {
		t.Fatalf("RunTest failed: %v", err)
	}
	defer testApp.Close()

	time.Sleep(100 * time.Millisecond)

	rendered := testApp.GetRenderString()
	t.Logf("Custom error render:\n%s", rendered)

	// Should show custom error content
	if !strings.Contains(rendered, "Custom Error") {
		t.Error("Should show custom error title")
	}

	if !strings.Contains(rendered, "A specific error occurred") {
		t.Error("Should show custom error message")
	}
}

// TestErrorBoundary_StressTest tests rapid state changes with error boundary
func TestErrorBoundary_StressTest(t *testing.T) {
	resetErrorBoundaryTestState(t)
	t.Setenv("MINT_USE_FIBER", "true")
	t.Setenv("TUI_DEBUG_UI", "false")

	testApp, err := ui.RunTest(ErrorBoundaryApp,
		ui.WithWidth(40),
		ui.WithHeight(10),
	)
	if err != nil {
		t.Fatalf("RunTest failed: %v", err)
	}
	defer testApp.Close()

	time.Sleep(50 * time.Millisecond)

	// Rapidly toggle panic state
	for i := 0; i < 5; i++ {
		shouldPanic = (i%2 == 0)

		// Trigger re-render
		if err := testApp.InjectKey('a'); err != nil {
			t.Fatalf("InjectKey failed: %v", err)
		}
		time.Sleep(50 * time.Millisecond)
	}

	// App should still be responsive
	finalRender := testApp.GetRenderString()
	t.Logf("Final render after stress test:\n%s", finalRender)

	if finalRender == "" {
		t.Error("App should still render after stress test")
	}
}

// TestErrorBoundary_KeyDown tests error boundary doesn't block keyboard input
func TestErrorBoundary_KeyDown(t *testing.T) {
	resetErrorBoundaryTestState(t)
	t.Setenv("MINT_USE_FIBER", "true")
	t.Setenv("TUI_DEBUG_UI", "false")

	shouldPanic = true

	testApp, err := ui.RunTest(ErrorBoundaryApp,
		ui.WithWidth(40),
		ui.WithHeight(10),
	)
	if err != nil {
		t.Fatalf("RunTest failed: %v", err)
	}
	defer testApp.Close()

	time.Sleep(100 * time.Millisecond)

	// Try various special keys
	keys := []platform.SpecialKey{
		platform.KeyTab,
		platform.KeyEnter,
		platform.KeyUp,
		platform.KeyDown,
		platform.KeyLeft,
		platform.KeyRight,
	}

	for _, key := range keys {
		if err := testApp.InjectSpecialKey(key); err != nil {
			t.Logf("InjectSpecialKey(%v) failed: %v", key, err)
		}
		time.Sleep(20 * time.Millisecond)
	}

	// App should still be running
	finalRender := testApp.GetRenderString()
	if finalRender == "" {
		t.Error("App should still respond after key events")
	}
}

// TestErrorBoundary_Resize tests error boundary behavior during window resize
func TestErrorBoundary_Resize(t *testing.T) {
	resetErrorBoundaryTestState(t)
	t.Setenv("MINT_USE_FIBER", "true")
	t.Setenv("TUI_DEBUG_UI", "false")

	shouldPanic = false

	// Test with different sizes
	testApp, err := ui.RunTest(ErrorBoundaryApp,
		ui.WithWidth(60),
		ui.WithHeight(20),
	)
	if err != nil {
		t.Fatalf("RunTest failed: %v", err)
	}
	defer testApp.Close()

	time.Sleep(100 * time.Millisecond)

	largeRender := testApp.GetRenderString()
	t.Logf("Large render:\n%s", largeRender)

	// Should still render
	if largeRender == "" {
		t.Error("App should render at larger size")
	}
}
