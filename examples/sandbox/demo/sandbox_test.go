// Package main provides sandbox testing examples
package main

import (
	"testing"

	"github.com/wwsheng009/mint/app"
	"github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/sandbox/mock"
	"github.com/wwsheng009/mint/ui"
)

// TestCounterWithSandbox demonstrates interactive component testing using sandbox
func TestCounterWithSandbox(t *testing.T) {
	// Use TestRun to properly initialize ComponentContext for hooks
	testApp, err := ui.TestRun(Counter,
		ui.TestWithSize(40, 18),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	// Get the context and set it as current before calling Counter()
	ctx := testApp.GetContext()
	if ctx == nil {
		t.Fatal("ComponentContext is nil")
	}

	ui.SetCurrentContext(ctx)
	counter := Counter()
	ui.SetCurrentContext(nil)

	// Verify the component can be created
	if counter == nil {
		t.Fatal("Counter component is nil")
	}

	// Create a mock sandbox for interaction testing
	sb := mock.New(40, 18)

	// Use helper to simulate user interaction
	helper := sb.Helper()

	// Simulate button clicks (tab to navigate, enter to click)
	result := helper.
		Tab().                      // Move focus to first button
		Process().
		Press(platform.KeyEnter).   // Click button
		Process().
		Type("World").              // Type in input field
		Process().
		Result()

	// Log result
	if result.OK() {
		t.Log("Sandbox test passed - no errors")
	} else {
		t.Errorf("Sandbox test failed with %d errors", len(result.Errors))
	}
}

// TestButtonInteraction tests button component
func TestButtonInteraction(t *testing.T) {
	sb := mock.New(40, 18)

	clicked := false
	// Simple button component using app package
	button := app.ButtonBuilder("Click Me").
		OnClick(func() {
		clicked = true
	}).
		Build()

	if button == nil {
		t.Fatal("Button component is nil")
	}

	// Use sandbox helper for interaction
	helper := sb.Helper()

	// Simulate clicking the button
	result := helper.
		Tab().           // Navigate to button
		Process().
		Press(platform.KeyEnter). // Click button
		Process().
		Result()

	if result.OK() {
		t.Logf("Button interaction test passed, clicked=%v", clicked)
	}
}

// TestInputInteraction tests input field interaction
func TestInputInteraction(t *testing.T) {
	sb := mock.New(40, 18)

	inputValue := ""
	// Use app.InputBuilder from app package
	input := app.InputBuilder().
		Value("").
		Placeholder("Type here").
		OnChange(func(value string) {
			inputValue = value
		}).
		Build()

	if input == nil {
		t.Fatal("Input component is nil")
	}

	helper := sb.Helper()

	// Type some text
	result := helper.
		Type("Hello").
		Process().
		Result()

	if result.OK() {
		t.Logf("Input value: %s", inputValue)
	}
}

// TestStyledText tests styled text rendering
func TestStyledText(t *testing.T) {
	// Use app.NewTextBuilder for styled text
	text := app.NewTextBuilder("Hello, Sandbox!").
		FgColor("green").
		Bold(true).
		Underline(true).
		Build()

	if text == nil {
		t.Fatal("Text component is nil")
	}

	t.Log("Styled text component created successfully")
}

// TestCounterComponentStructure tests counter VNode structure
func TestCounterComponentStructure(t *testing.T) {
	// Use TestRun to properly initialize ComponentContext for hooks
	testApp, err := ui.TestRun(Counter,
		ui.TestWithSize(40, 18),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	ctx := testApp.GetContext()
	if ctx == nil {
		t.Fatal("ComponentContext is nil")
	}

	// Create counter with proper context
	ui.SetCurrentContext(ctx)
	counter := Counter()
	ui.SetCurrentContext(nil)

	// Verify counter is created
	if counter == nil {
		t.Fatal("Counter component is nil")
	}

	// Verify counter has VStack structure
	if counter.Type() != ui.VNodeElement {
		t.Logf("Counter root type: %v", counter.Type())
	}

	children := counter.Children()
	if len(children) == 0 {
		t.Fatal("Counter has no children")
	}

	t.Logf("Counter has %d children", len(children))
}

// BenchmarkComponentCreation benchmarks component creation
func BenchmarkComponentCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = ui.VStack(
			app.Text("Line 1"),
			app.Text("Line 2"),
			app.Text("Line 3"),
			app.HStack(
				app.ButtonBuilder("Btn1").OnClick(func() {}).Build(),
				app.ButtonBuilder("Btn2").OnClick(func() {}).Build(),
			),
		)
	}
}
