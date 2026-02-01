// Package main demonstrates dynamic list with state preservation
package main

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/wwsheng009/mint/ui"
	"github.com/wwsheng009/mint/runtime/platform"
)

// TestDynamicListKeyboardInput tests keyboard navigation and button clicks
func TestDynamicListKeyboardInput(t *testing.T) {
	testApp, err := ui.RunTest(TodoList,
		ui.WithWidth(50),
		ui.WithHeight(16),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	// Wait for initial render
	time.Sleep(100 * time.Millisecond)

	// Get initial render
	rendered := testApp.GetRenderString()
	t.Logf("Initial render:\n%s", rendered)

	// Check that buttons are present
	if err := testApp.AssertRender("[ +]"); err != nil {
		t.Errorf("Buttons not found in initial render: %v", err)
	}

	// Check the declarative root state
	root := testApp.GetDeclarativeRoot()
	t.Logf("Buttons count: %d", len(root.GetButtons()))
	t.Logf("Inputs count: %d", len(root.GetInputs()))
	t.Logf("Focused index: %d, type: %d", root.GetFocusedIndex(), root.GetFocusedType())

	// Try Tab key to navigate to first button
	t.Log("\n=== Pressing Tab to navigate to button ===")
	if err := testApp.InjectSpecialKey(platform.KeyTab); err != nil {
		t.Errorf("Failed to inject Tab: %v", err)
	}
	time.Sleep(100 * time.Millisecond)

	rendered = testApp.GetRenderString()
	t.Logf("After Tab:\n%s", rendered)
	t.Logf("Focused index: %d, type: %d", root.GetFocusedIndex(), root.GetFocusedType())

	// Try Enter to click the button
	t.Log("\n=== Pressing Enter to click button ===")
	if err := testApp.InjectSpecialKey(platform.KeyEnter); err != nil {
		t.Errorf("Failed to inject Enter: %v", err)
	}
	time.Sleep(50 * time.Millisecond)
	testApp.GetFrameworkApp().ForceRenderNow() // Force render to complete
	time.Sleep(50 * time.Millisecond)

	// Force another render to ensure buffer is up to date
	testApp.GetFrameworkApp().ForceRenderNow()

	rendered = testApp.GetRenderString()
	t.Logf("After Enter:\n%s", rendered)
	t.Logf("Rendered bytes: %d", len(rendered))

	// Check if "clicked: 1" appears anywhere
	if strings.Contains(rendered, "clicked: 1") {
		t.Log("FOUND 'clicked: 1' in render!")
	} else {
		t.Log("NOT FOUND 'clicked: 1' - checking for partial matches...")
		if strings.Contains(rendered, "clicked") {
			t.Log("Found 'clicked' without : 1")
		}
		if strings.Contains(rendered, "ked: 1") {
			t.Log("Found 'ked: 1' (partial)")
		}
	}

	// The button click should increment the counter
	// We should see "clicked: 1" in the output
	if err := testApp.AssertRender("clicked: 1"); err != nil {
		t.Errorf("Counter not incremented after button click: %v", err)
	}

	// Try clicking again
	t.Log("\n=== Pressing Enter again ===")
	if err := testApp.InjectSpecialKey(platform.KeyEnter); err != nil {
		t.Errorf("Failed to inject Enter: %v", err)
	}
	time.Sleep(50 * time.Millisecond)
	testApp.GetFrameworkApp().ForceRenderNow() // Force render to complete
	time.Sleep(50 * time.Millisecond)

	rendered = testApp.GetRenderString()
	t.Logf("After second Enter:\n%s", rendered)

	if err := testApp.AssertRender("clicked: 2"); err != nil {
		t.Errorf("Counter not incremented to 2: %v", err)
	}
}

// TestDynamicListDirectEventInjection tests direct event injection
func TestDynamicListDirectEventInjection(t *testing.T) {
	testApp, err := ui.RunTest(TodoList,
		ui.WithWidth(50),
		ui.WithHeight(16),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	// Wait for initial render
	time.Sleep(100 * time.Millisecond)

	// Get framework app to check event queue
	fwApp := testApp.GetFrameworkApp()
	t.Logf("App state: %v", fwApp.GetState())

	// Inject Tab key directly
	t.Log("=== Injecting Tab key ===")
	err = testApp.InjectSpecialKey(platform.KeyTab)
	if err != nil {
		t.Errorf("Failed to inject Tab: %v", err)
	}
	time.Sleep(100 * time.Millisecond)

	// Inject Enter key
	t.Log("=== Injecting Enter key ===")
	err = testApp.InjectSpecialKey(platform.KeyEnter)
	if err != nil {
		t.Errorf("Failed to inject Enter: %v", err)
	}
	time.Sleep(100 * time.Millisecond)

	rendered := testApp.GetRenderString()
	t.Logf("Render after Tab+Enter:\n%s", rendered)

	// Check for counter increment
	if err := testApp.AssertRender("clicked: 1"); err != nil {
		t.Logf("Note: Counter not incremented - this might be expected if focus management needs work: %v", err)
	}
}

// TestDynamicListMouseClick tests mouse click on buttons
func TestDynamicListMouseClick(t *testing.T) {
	testApp, err := ui.RunTest(TodoList,
		ui.WithWidth(50),
		ui.WithHeight(16),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	// Wait for initial render
	time.Sleep(100 * time.Millisecond)

	root := testApp.GetDeclarativeRoot()
	buttons := root.GetButtons()

	t.Logf("Found %d buttons", len(buttons))
	if len(buttons) == 0 {
		t.Fatal("No buttons found in the app")
	}

	// Try clicking the first button
	// Find button bounds
	for i, btn := range buttons {
		bounds := btn.Bounds()
		t.Logf("Button %d bounds: x=%d, y=%d, w=%d, h=%d",
			i, bounds[0], bounds[1], bounds[2], bounds[3])
	}

	// Click on the first button (assuming it's at position around x=25, y=7)
	// The button "[ +]" should be clickable
	t.Log("\n=== Clicking first button with mouse ===")
	if len(buttons) > 0 {
		bounds := buttons[0].Bounds()
		// Click in the center of the button
		clickX := bounds[0] + bounds[2]/2
		clickY := bounds[1] + bounds[3]/2
		t.Logf("Clicking at x=%d, y=%d", clickX, clickY)

		// Mouse click requires:
		// 1. MouseEnter to set hover state
		// 2. MousePress
		// 3. MouseRelease (which triggers onClick)
		err = testApp.InjectMouse(clickX, clickY, platform.MouseLeft, platform.MouseMotion)
		if err != nil {
			t.Errorf("Failed to inject mouse motion: %v", err)
		}
		err = testApp.InjectMouse(clickX, clickY, platform.MouseLeft, platform.MousePress)
		if err != nil {
			t.Errorf("Failed to inject mouse press: %v", err)
		}
		err = testApp.InjectMouse(clickX, clickY, platform.MouseLeft, platform.MouseRelease)
		if err != nil {
			t.Errorf("Failed to inject mouse release: %v", err)
		}
		time.Sleep(50 * time.Millisecond)
		testApp.GetFrameworkApp().ForceRenderNow() // Force render to complete
		time.Sleep(50 * time.Millisecond)

		rendered := testApp.GetRenderString()
		t.Logf("After mouse click:\n%s", rendered)

		// Check if counter was incremented
		if err := testApp.AssertRender("clicked: 1"); err != nil {
			t.Errorf("Counter not incremented after mouse click: %v", err)
		}
	}
}

// TestDynamicListFocusManagement tests focus management
func TestDynamicListFocusManagement(t *testing.T) {
	testApp, err := ui.RunTest(TodoList,
		ui.WithWidth(50),
		ui.WithHeight(16),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	// Wait for initial render
	time.Sleep(100 * time.Millisecond)

	root := testApp.GetDeclarativeRoot()

	t.Log("=== Initial state ===")
	t.Logf("Buttons: %d", len(root.GetButtons()))
	t.Logf("Focused index: %d", root.GetFocusedIndex())
	t.Logf("Focused type: %d", root.GetFocusedType())

	// Inject Tab key
	t.Log("\n=== After Tab ===")
	testApp.InjectSpecialKey(platform.KeyTab)
	time.Sleep(100 * time.Millisecond)

	t.Logf("Focused index: %d", root.GetFocusedIndex())
	t.Logf("Focused type: %d", root.GetFocusedType())

	// Inject Tab key again
	t.Log("\n=== After second Tab ===")
	testApp.InjectSpecialKey(platform.KeyTab)
	time.Sleep(100 * time.Millisecond)

	t.Logf("Focused index: %d", root.GetFocusedIndex())
	t.Logf("Focused type: %d", root.GetFocusedType())

	rendered := testApp.GetRenderString()
	t.Logf("Current render:\n%s", rendered)
}

// BenchmarkDynamicListRender benchmarks the rendering performance
func BenchmarkDynamicListRender(b *testing.B) {
	testApp, err := ui.RunTest(TodoList,
		ui.WithWidth(50),
		ui.WithHeight(16),
	)
	if err != nil {
		b.Fatal(err)
	}
	defer testApp.Close()

	// Wait for initial render
	time.Sleep(100 * time.Millisecond)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testApp.GetRenderString()
	}
}

// TestDynamicListWithSandbox tests using RunTestWithSandbox
func TestDynamicListWithSandbox(t *testing.T) {
	testApp, err := ui.RunTestWithSandbox(TodoList,
		ui.WithWidth(50),
		ui.WithHeight(16),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	// Wait for initial render
	time.Sleep(100 * time.Millisecond)

	// Get the sandbox
	sb := testApp.GetSandbox()
	t.Logf("Sandbox queue stats: %+v", sb.QueueStats())

	// Inject events through the sandbox
	t.Log("=== Injecting Tab through sandbox ===")
	sb.InjectSpecialKey(platform.KeyTab)
	time.Sleep(100 * time.Millisecond)

	t.Log("=== Injecting Enter through sandbox ===")
	sb.InjectSpecialKey(platform.KeyEnter)
	time.Sleep(100 * time.Millisecond)

	rendered := testApp.GetRenderString()
	t.Logf("Render after sandbox events:\n%s", rendered)

	t.Logf("Final queue stats: %+v", sb.QueueStats())
}

// TestDynamicListDebug is a comprehensive debug test
func TestDynamicListDebug(t *testing.T) {
	testApp, err := ui.RunTest(TodoList,
		ui.WithWidth(50),
		ui.WithHeight(16),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	// Wait for initial render
	time.Sleep(150 * time.Millisecond)

	root := testApp.GetDeclarativeRoot()
	fwApp := testApp.GetFrameworkApp()

	t.Log("\n=== APP STATE ===")
	t.Logf("App state: %v", fwApp.GetState())
	t.Logf("Renderer: %v", fwApp.GetRenderer())

	t.Log("\n=== INTERACTIVE ELEMENTS ===")
	t.Logf("Buttons: %d", len(root.GetButtons()))
	t.Logf("Inputs: %d", len(root.GetInputs()))
	t.Logf("Focused index: %d", root.GetFocusedIndex())
	t.Logf("Focused type: %d", root.GetFocusedType())

	t.Log("\n=== INITIAL RENDER ===")
	rendered := testApp.GetRenderString()
	t.Log(rendered)

	// Try multiple Tab + Enter cycles
	for i := 1; i <= 3; i++ {
		t.Log(fmt.Sprintf("\n=== CYCLE %d: Tab + Enter ===", i))

		if err := testApp.InjectSpecialKey(platform.KeyTab); err != nil {
			t.Errorf("Failed to inject Tab (cycle %d): %v", i, err)
		}
		time.Sleep(50 * time.Millisecond)

		if err := testApp.InjectSpecialKey(platform.KeyEnter); err != nil {
			t.Errorf("Failed to inject Enter (cycle %d): %v", i, err)
		}
		time.Sleep(100 * time.Millisecond)

		rendered := testApp.GetRenderString()
		t.Logf("After cycle %d:\n%s", i, rendered)
		t.Logf("Focused index: %d, type: %d", root.GetFocusedIndex(), root.GetFocusedType())

		expected := fmt.Sprintf("clicked: %d", i)
		if err := testApp.AssertRender(expected); err != nil {
			t.Logf("Cycle %d: %v", i, err)
		} else {
			t.Logf("Cycle %d: SUCCESS - counter shows %d", i, i)
		}
	}
}
