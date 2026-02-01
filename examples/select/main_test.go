// Package main tests the select dropdown component
package main

import (
	"strings"
	"testing"
	"time"

	"github.com/wwsheng009/mint/ui"
	"github.com/wwsheng009/mint/runtime/platform"
)

// TestSelectInitialRender tests the initial render of the select component
func TestSelectInitialRender(t *testing.T) {
	testApp, err := ui.RunTest(SelectDemo,
		ui.WithWidth(50),
		ui.WithHeight(22),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	// Wait for initial render
	time.Sleep(100 * time.Millisecond)

	rendered := testApp.GetRenderString()
	t.Logf("Initial render:\n%s", rendered)

	// Check for select component elements
	if err := testApp.AssertRender("Theme:"); err != nil {
		t.Errorf("Theme label not found: %v", err)
	}
	if err := testApp.AssertRender("Dark Theme"); err != nil {
		t.Errorf("Initial selection 'Dark Theme' not found: %v", err)
	}
	if err := testApp.AssertRender("Selected: Dark Theme"); err != nil {
		t.Errorf("Selected text not found: %v", err)
	}
}

// TestSelectMouseClick tests clicking on the select to cycle options
func TestSelectMouseClick(t *testing.T) {
	testApp, err := ui.RunTest(SelectDemo,
		ui.WithWidth(50),
		ui.WithHeight(22),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	// Wait for initial render
	time.Sleep(100 * time.Millisecond)

	// Get initial state
	rendered := testApp.GetRenderString()
	t.Logf("Initial state:\n%s", rendered)

	// Should show "Dark Theme" initially
	if !strings.Contains(rendered, "Dark Theme") {
		t.Error("Initial selection should be 'Dark Theme'")
	}

	// Get the declarative root to access the select component
	root := testApp.GetDeclarativeRoot()

	// Find select components by checking bounds
	selects := root.GetSelects()
	if len(selects) == 0 {
		t.Fatal("No select components found")
	}

	t.Logf("Found %d select component(s)", len(selects))
	selectComp := selects[0]
	bounds := selectComp.Bounds()
	t.Logf("Select bounds: x=%d, y=%d, w=%d, h=%d", bounds[0], bounds[1], bounds[2], bounds[3])

	// Click on the select to cycle to next option (Light Theme)
	clickX := bounds[0] + bounds[2]/2
	clickY := bounds[1] + bounds[3]/2
	t.Logf("Clicking at x=%d, y=%d", clickX, clickY)

	// Mouse click: Motion -> Press -> Release
	testApp.InjectMouse(clickX, clickY, platform.MouseLeft, platform.MouseMotion)
	time.Sleep(50 * time.Millisecond)
	testApp.InjectMouse(clickX, clickY, platform.MouseLeft, platform.MousePress)
	time.Sleep(50 * time.Millisecond)
	testApp.InjectMouse(clickX, clickY, platform.MouseLeft, platform.MouseRelease)
	time.Sleep(50 * time.Millisecond)
	testApp.GetFrameworkApp().ForceRenderNow()
	time.Sleep(50 * time.Millisecond)

	rendered = testApp.GetRenderString()
	t.Logf("After first click:\n%s", rendered)

	// Should now show "Light Theme"
	if err := testApp.AssertRender("Light Theme"); err != nil {
		t.Errorf("Selection didn't change to 'Light Theme': %v", err)
	}
	if err := testApp.AssertRender("Selected: Light Theme"); err != nil {
		t.Errorf("Selected text not updated: %v", err)
	}

	// Click again to cycle to Dracula Theme
	// Re-fetch bounds in case they changed after re-render
	selects = root.GetSelects()
	if len(selects) == 0 {
		t.Fatal("Select component disappeared after first click")
	}
	selectComp = selects[0]
	bounds = selectComp.Bounds()
	clickX = bounds[0] + bounds[2]/2
	clickY = bounds[1] + bounds[3]/2

	testApp.InjectMouse(clickX, clickY, platform.MouseLeft, platform.MouseMotion)
	time.Sleep(50 * time.Millisecond)
	testApp.InjectMouse(clickX, clickY, platform.MouseLeft, platform.MousePress)
	time.Sleep(50 * time.Millisecond)
	testApp.InjectMouse(clickX, clickY, platform.MouseLeft, platform.MouseRelease)
	time.Sleep(50 * time.Millisecond)
	testApp.GetFrameworkApp().ForceRenderNow()
	time.Sleep(50 * time.Millisecond)

	rendered = testApp.GetRenderString()

	if err := testApp.AssertRender("Dracula Theme"); err != nil {
		t.Errorf("Selection didn't change to 'Dracula Theme': %v", err)
	}
	if err := testApp.AssertRender("Selected: Dracula Theme"); err != nil {
		t.Errorf("Selected text not updated: %v", err)
	}
}

// TestSelectCycleAllOptions tests cycling through all options
func TestSelectCycleAllOptions(t *testing.T) {
	testApp, err := ui.RunTest(SelectDemo,
		ui.WithWidth(50),
		ui.WithHeight(22),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	// Wait for initial render
	time.Sleep(100 * time.Millisecond)

	root := testApp.GetDeclarativeRoot()
	selects := root.GetSelects()
	if len(selects) == 0 {
		t.Fatal("No select components found")
	}

	selectComp := selects[0]
	bounds := selectComp.Bounds()
	clickX := bounds[0] + bounds[2]/2
	clickY := bounds[1] + bounds[3]/2

	// Expected sequence: Dark -> Light -> Dracula -> Nord -> Dark (wraps)
	expected := []string{"Dark Theme", "Light Theme", "Dracula Theme", "Nord Theme", "Dark Theme"}

	for i, exp := range expected {
		t.Logf("Cycle %d: expecting %s", i, exp)

		rendered := testApp.GetRenderString()

		// Verify current selection
		if !strings.Contains(rendered, exp) {
			t.Errorf("Cycle %d: expected %s but not found in render", i, exp)
			t.Logf("Render:\n%s", rendered)
		}

		// Click to next option (except for last one)
		if i < len(expected)-1 {
			testApp.InjectMouse(clickX, clickY, platform.MouseLeft, platform.MouseMotion)
			testApp.InjectMouse(clickX, clickY, platform.MouseLeft, platform.MousePress)
			testApp.InjectMouse(clickX, clickY, platform.MouseLeft, platform.MouseRelease)
			time.Sleep(50 * time.Millisecond)
			testApp.GetFrameworkApp().ForceRenderNow()
			time.Sleep(50 * time.Millisecond)
		}
	}
}

// TestSelectWithSandbox tests using RunTestWithSandbox
func TestSelectWithSandbox(t *testing.T) {
	testApp, err := ui.RunTestWithSandbox(SelectDemo,
		ui.WithWidth(50),
		ui.WithHeight(22),
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

	rendered := testApp.GetRenderString()
	t.Logf("Initial render with sandbox:\n%s", rendered)

	if err := testApp.AssertRender("Dark Theme"); err != nil {
		t.Errorf("Initial selection not found: %v", err)
	}
}

// TestSelectComponentsFound tests that select components are properly collected
func TestSelectComponentsFound(t *testing.T) {
	testApp, err := ui.RunTest(SelectDemo,
		ui.WithWidth(50),
		ui.WithHeight(22),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	// Wait for initial render
	time.Sleep(100 * time.Millisecond)

	root := testApp.GetDeclarativeRoot()

	// Check selects
	selects := root.GetSelects()
	t.Logf("Selects count: %d", len(selects))
	if len(selects) == 0 {
		t.Error("No select components found")
	} else {
		s := selects[0]
		t.Logf("Select: selected=%d, options=%d, focused=%v",
			s.Selected(), len(s.Options()), s.IsFocused())
		if s.Selected() != 0 {
			t.Errorf("Expected selected index 0, got %d", s.Selected())
		}
		if len(s.Options()) != 4 {
			t.Errorf("Expected 4 options, got %d", len(s.Options()))
		}
	}
}
