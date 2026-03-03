// Package main tests the select dropdown component
package main

import (
	"strings"
	"testing"
	"time"

	"github.com/wwsheng009/mint/ui"
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

// TestSelectKeyboardNavigation tests using keyboard to navigate options
func TestSelectKeyboardNavigation(t *testing.T) {
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

	// Press Tab to focus the select component
	testApp.InjectKey('\t')
	time.Sleep(50 * time.Millisecond)
	testApp.GetFrameworkApp().ForceRenderNow()
	time.Sleep(50 * time.Millisecond)

	// Press Down arrow to cycle to next option (Light Theme)
	testApp.InjectSpecialKey(0x102) // KeyDown (platform.KeyDown)
	time.Sleep(50 * time.Millisecond)
	testApp.GetFrameworkApp().ForceRenderNow()
	time.Sleep(50 * time.Millisecond)

	rendered = testApp.GetRenderString()
	t.Logf("After Down arrow:\n%s", rendered)

	// Should now show "Light Theme"
	if err := testApp.AssertRender("Light Theme"); err != nil {
		t.Errorf("Selection didn't change to 'Light Theme': %v", err)
	}
	if err := testApp.AssertRender("Selected: Light Theme"); err != nil {
		t.Errorf("Selected text not updated: %v", err)
	}

	// Press Down again to cycle to Dracula Theme
	testApp.InjectSpecialKey(0x102) // KeyDown
	time.Sleep(50 * time.Millisecond)
	testApp.GetFrameworkApp().ForceRenderNow()
	time.Sleep(50 * time.Millisecond)

	rendered = testApp.GetRenderString()

	if err := testApp.AssertRender("Dracula Theme"); err != nil {
		t.Errorf("Selection didn't change to 'Dracula Theme': %v", err)
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

// TestSelectTableRender tests that the table is rendered alongside select
func TestSelectTableRender(t *testing.T) {
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
	t.Logf("Render:\n%s", rendered)

	// Check for table elements
	if !strings.Contains(rendered, "ID") || !strings.Contains(rendered, "Name") {
		t.Error("Table headers not found in render")
	}
	if !strings.Contains(rendered, "Alice") || !strings.Contains(rendered, "Bob") {
		t.Error("Table content not found in render")
	}
}
