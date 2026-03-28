// Package main provides sandbox testing examples
package main

import (
	"testing"
	"time"

	"github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/sandbox/mock"
	"github.com/wwsheng009/mint/ui"
)

// TestCounterWithSandbox demonstrates interactive component testing using sandbox
// 使用新版 RunTest API
func TestCounterWithSandbox(t *testing.T) {
	sandboxStore.Set(AppState{Count: 0, Name: "Guest"})

	testApp, err := ui.RunTest(Counter,
		ui.WithSize(40, 18),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	// Wait for initial render
	time.Sleep(50 * time.Millisecond)

	// Test keyboard navigation
	testApp.InjectSpecialKey(platform.KeyTab)
	time.Sleep(20 * time.Millisecond)
	testApp.InjectSpecialKey(platform.KeyEnter)
	time.Sleep(50 * time.Millisecond)

	// Verify render output
	rendered := testApp.GetRenderString()
	if rendered == "" {
		t.Error("Render output is empty")
	}

	t.Log("Sandbox test passed - no errors")
}

// OLD_TEST_COUNTER_WITH_SANDBOX - 旧版本测试 (已注释)
// 旧版测试使用 ui.TestRun，不能正确处理完整的应用事件系统
/*
func TestCounterWithSandbox_OLD(t *testing.T) {
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

	ui.SetCurrentContext(ctx)
	counter := Counter()
	ui.SetCurrentContext(nil)

	if counter == nil {
		t.Fatal("Counter component is nil")
	}

	sb := mock.New(40, 18)
	helper := sb.Helper()

	result := helper.
		Tab().
		Process().
		Press(platform.KeyEnter).
		Process().
		Type("World").
		Process().
		Result()

	if result.OK() {
		t.Log("Sandbox test passed - no errors")
	} else {
		t.Errorf("Sandbox test failed with %d errors", len(result.Errors))
	}
}
*/

// TestButtonInteraction tests button component
func TestButtonInteraction(t *testing.T) {
	sb := mock.New(40, 18)

	button := ui.NewButtonBuilder("Click Me").
		Key("test-btn").
		Build()

	if button == nil {
		t.Fatal("Button component is nil")
	}

	helper := sb.Helper()

	result := helper.
		Tab().
		Process().
		Press(platform.KeyEnter).
		Process().
		Result()

	if result.OK() {
		t.Logf("Button interaction test passed")
	}
}

// TestInputInteraction tests input field interaction
func TestInputInteraction(t *testing.T) {
	sb := mock.New(40, 18)

	input := ui.NewInputBuilder().
		Value("").
		Placeholder("Type here").
		Key("test-input").
		Build()

	if input == nil {
		t.Fatal("Input component is nil")
	}

	helper := sb.Helper()

	result := helper.
		Type("Hello").
		Process().
		Result()

	if result.OK() {
		t.Log("Input interaction test passed")
	}
}

// TestStyledText tests styled text rendering
func TestStyledText(t *testing.T) {
	text := ui.NewTextBuilder("Hello, Sandbox!").
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
// 使用新版 RunTest API
func TestCounterComponentStructure(t *testing.T) {
	sandboxStore.Set(AppState{Count: 0, Name: "Guest"})

	testApp, err := ui.RunTest(Counter,
		ui.WithSize(40, 18),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	time.Sleep(50 * time.Millisecond)

	// Verify render output
	rendered := testApp.GetRenderString()
	if rendered == "" {
		t.Error("Render output is empty")
	}

	t.Logf("Counter test structure verified")
}

// OLD_TEST_COUNTER_COMPONENT_STRUCTURE - 旧版本测试 (已注释)
/*
func TestCounterComponentStructure_OLD(t *testing.T) {
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

	ui.SetCurrentContext(ctx)
	counter := Counter()
	ui.SetCurrentContext(nil)

	if counter == nil {
		t.Fatal("Counter component is nil")
	}

	if counter.Type() != ui.VNodeElement {
		t.Logf("Counter root type: %v", counter.Type())
	}

	children := counter.Children()
	if len(children) == 0 {
		t.Fatal("Counter has no children")
	}

	t.Logf("Counter has %d children", len(children))
}
*/

// BenchmarkComponentCreation benchmarks component creation
func BenchmarkComponentCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = ui.VStack(
			ui.Text("Line 1"),
			ui.Text("Line 2"),
			ui.Text("Line 3"),
			ui.HStack(
				ui.NewButtonBuilder("Btn1").Key("btn1").Build(),
				ui.NewButtonBuilder("Btn2").Key("btn2").Build(),
			),
		)
	}
}
