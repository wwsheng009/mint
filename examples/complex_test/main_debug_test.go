package main

import (
	"testing"
	"time"

	ui "github.com/wwsheng009/mint/ui"
)

// TestTabEventFlow tests Tab key event flow in detail
func TestTabEventFlow(t *testing.T) {
	testApp, err := ui.RunTestWithSandbox(RootApp,
		ui.WithWidth(100),
		ui.WithHeight(40),
	)
	if err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}
	defer testApp.Close()

	time.Sleep(200 * time.Millisecond)

	focusMgr := testApp.GetFocusManager()

	focusable := testApp.GetButtons()
	t.Logf("Initial: %d buttons", len(focusable))
	if fm, ok := focusMgr.(interface{ CurrentIndex() int }); ok {
		t.Logf("Initial focus index: %d", fm.CurrentIndex())
	}

	// Inject Tab key
	t.Log("\n=== Injecting Tab key ===")
	err = testApp.InjectSpecialKey(3) // KeyTab = 3 (KeyUnknown=0, KeyEscape=1, KeyEnter=2, KeyTab=3)
	t.Logf("InjectSpecialKey result: %v", err)
	time.Sleep(100 * time.Millisecond)

	if fm, ok := focusMgr.(interface{ CurrentIndex() int }); ok {
		t.Logf("After Tab - Focus index: %d", fm.CurrentIndex())
	}

	// Check button focus states
	maxCheck := 3
	if len(focusable) < maxCheck {
		maxCheck = len(focusable)
	}
	for i := 0; i < maxCheck; i++ {
		if fb, ok := focusable[i].(interface{ IsFocused() bool }); ok {
			t.Logf("Button %d IsFocused: %v", i, fb.IsFocused())
		}
	}

	// Check if re-render happened
	newFocusable := testApp.GetButtons()
	t.Logf("After Tab - Button count: %d", len(newFocusable))

	// Check if VNode objects are the same
	if len(focusable) > 0 && len(newFocusable) > 0 {
		// Compare pointers
		t.Logf("Button 0 address changed: %v", &focusable[0] != &newFocusable[0])
	}

	// Try direct FocusNext
	t.Log("\n=== Direct FocusNext call ===")
	if fm, ok := focusMgr.(interface{ FocusNext() bool }); ok {
		fm.FocusNext()
	}
	time.Sleep(100 * time.Millisecond)

	if fm, ok := focusMgr.(interface{ CurrentIndex() int }); ok {
		t.Logf("After FocusNext - Focus index: %d", fm.CurrentIndex())
	}

	rendered := testApp.GetRenderString()
	t.Logf("Render after FocusNext:\n%s", rendered)
}
