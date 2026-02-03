package main

import (
	"testing"
	"time"

	"github.com/wwsheng009/mint/runtime/platform"
	ui "github.com/wwsheng009/mint/ui"
)

// TestTabNavigation tests Tab key navigation between buttons
func TestTabNavigation(t *testing.T) {
	testApp, err := ui.RunTestWithSandbox(Counter,
		ui.WithWidth(40),
		ui.WithHeight(12),
	)
	if err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}
	defer testApp.Close()

	time.Sleep(200 * time.Millisecond)

	// Get initial state
	focusMgr := testApp.GetFocusManager()
	buttons := testApp.GetButtons()

	t.Logf("Initial: %d buttons", len(buttons))
	if fm, ok := focusMgr.(interface{ CurrentIndex() int }); ok {
		t.Logf("Initial focus index: %d", fm.CurrentIndex())
	}

	// Get initial render
	initialRender := testApp.GetRenderString()
	t.Logf("Initial render:\n%s", initialRender)

	// Inject Tab key
	t.Log("\n=== Injecting Tab key ===")
	err = testApp.InjectSpecialKey(3) // KeyTab = 3
	if err != nil {
		t.Logf("InjectSpecialKey error: %v", err)
	}
	time.Sleep(300 * time.Millisecond) // Wait for re-render

	// Check focus index after Tab
	if fm, ok := focusMgr.(interface{ CurrentIndex() int }); ok {
		afterTabIndex := fm.CurrentIndex()
		t.Logf("After Tab - Focus index: %d", afterTabIndex)

		// Tab should have moved focus to the next button
		if afterTabIndex == 0 {
			t.Error("Tab key did not change focus index (still 0)")
		} else if afterTabIndex == 1 {
			t.Log("SUCCESS: Tab key moved focus to button 1")
		}
	}

	// Get render after Tab
	afterTabRender := testApp.GetRenderString()
	t.Logf("Render after Tab:\n%s", afterTabRender)

	// Check button focus states
	newButtons := testApp.GetButtons()
	t.Logf("After Tab - Button count: %d", len(newButtons))

	// Check focus state of buttons
	for i := 0; i < len(newButtons) && i < 2; i++ {
		if fb, ok := newButtons[i].(interface{ IsFocused() bool }); ok {
			t.Logf("Button %d IsFocused: %v", i, fb.IsFocused())
		}
	}
}

// TestShiftTabNavigation tests Shift+Tab key navigation
func TestShiftTabNavigation(t *testing.T) {
	testApp, err := ui.RunTestWithSandbox(Counter,
		ui.WithWidth(40),
		ui.WithHeight(12),
	)
	if err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}
	defer testApp.Close()

	time.Sleep(200 * time.Millisecond)

	focusMgr := testApp.GetFocusManager()

	// First move to second button
	t.Log("=== Moving to button 1 with Tab ===")
	testApp.InjectSpecialKey(3) // KeyTab
	time.Sleep(300 * time.Millisecond)

	if fm, ok := focusMgr.(interface{ CurrentIndex() int }); ok {
		t.Logf("After first Tab - Focus index: %d", fm.CurrentIndex())
	}

	// Now use Shift+Tab to go back
	t.Log("\n=== Injecting Shift+Tab key ===")
	// Inject Shift+Tab
	testApp.InjectSpecialKeyWithMod(3, platform.ModShift) // KeyTab=3 with Shift
	time.Sleep(300 * time.Millisecond)

	if fm, ok := focusMgr.(interface{ CurrentIndex() int }); ok {
		afterShiftTabIndex := fm.CurrentIndex()
		t.Logf("After Shift+Tab - Focus index: %d", afterShiftTabIndex)

		if afterShiftTabIndex == 0 {
			t.Log("SUCCESS: Shift+Tab moved focus back to button 0")
		} else {
			t.Errorf("Shift+Tab did not move focus back to button 0, got index %d", afterShiftTabIndex)
		}
	}
}
