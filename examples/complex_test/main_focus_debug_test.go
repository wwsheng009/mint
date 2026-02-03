package main

import (
	"testing"
	"time"

	ui "github.com/wwsheng009/mint/ui"
)

// TestFocusIDStability tests if focus IDs are stable across renders
func TestFocusIDStability(t *testing.T) {
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

	// Get buttons twice
	firstSet := testApp.GetButtons()

	// Trigger a re-render
	testApp.InjectKey(' ')
	time.Sleep(100 * time.Millisecond)

	secondSet := testApp.GetButtons()

	t.Logf("First set: %d buttons", len(firstSet))
	t.Logf("Second set: %d buttons", len(secondSet))

	// Compare focus IDs
	maxCompare := 3
	if len(firstSet) < maxCompare {
		maxCompare = len(firstSet)
	}
	if len(secondSet) < maxCompare {
		maxCompare = len(secondSet)
	}
	for i := 0; i < maxCompare; i++ {
		id1, ok1 := firstSet[i].(interface{ GetFocusID() string })
		id2, ok2 := secondSet[i].(interface{ GetFocusID() string })

		if ok1 && ok2 {
			sameID := id1.GetFocusID() == id2.GetFocusID()
			t.Logf("Button %d ID1='%s', ID2='%s', same=%v", i, id1.GetFocusID(), id2.GetFocusID(), sameID)
		}
	}

	// Test FocusNext and re-render
	t.Log("\n=== Testing FocusNext + re-render ===")
	if len(firstSet) > 0 {
		// Get current focus index via reflection
		if fm, ok := focusMgr.(interface{ CurrentIndex() int }); ok {
			t.Logf("Before FocusNext: current=%d", fm.CurrentIndex())
		}

		// Get the focus ID of button 0 before FocusNext
		if id, ok := firstSet[0].(interface{ GetFocusID() string }); ok {
			t.Logf("Button 0 focus ID: %s", id.GetFocusID())
		}

		if fm, ok := focusMgr.(interface{ FocusNext() bool }); ok {
			fm.FocusNext()
		}
		if fm, ok := focusMgr.(interface{ CurrentIndex() int }); ok {
			t.Logf("After FocusNext: current=%d", fm.CurrentIndex())
		}

		// Check if focus state was set
		if fb, ok := firstSet[0].(interface{ IsFocused() bool }); ok {
			t.Logf("Button 0 IsFocused after FocusNext: %v", fb.IsFocused())
		}
		if len(firstSet) > 1 {
			if fb, ok := firstSet[1].(interface{ IsFocused() bool }); ok {
				t.Logf("Button 1 IsFocused after FocusNext: %v", fb.IsFocused())
			}
		}

		// Now trigger re-render
		testApp.InjectKey(' ')
		time.Sleep(100 * time.Millisecond)

		thirdSet := testApp.GetButtons()
		t.Logf("After re-render: %d buttons", len(thirdSet))

		// Check if focus index was preserved
		if fm, ok := focusMgr.(interface{ CurrentIndex() int }); ok {
			t.Logf("Focus index after re-render: current=%d", fm.CurrentIndex())
		}

		if len(thirdSet) > 1 {
			if fb, ok := thirdSet[1].(interface{ IsFocused() bool }); ok {
				t.Logf("New button 1 IsFocused: %v", fb.IsFocused())
			}
		}
	}
}
