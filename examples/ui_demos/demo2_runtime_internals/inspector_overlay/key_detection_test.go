package main

import (
	"testing"
	"time"

	"github.com/wwsheng009/mint/framework"
	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/internal/inspector"
	"github.com/wwsheng009/mint/internal/render"
)

// TestKeyDetection tests if Inspector receives keys correctly
func TestKeyDetection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping interactive test in short mode")
	}

	// Setup inspector
	globalInspector = inspector.NewStandaloneInspector()
	globalInspector.Enable()
	globalInspector.ToggleVisibility()

	_ = theme.SetTheme("nord")

	fwApp := framework.NewApp()
	fwApp.SetInspector(globalInspector)
	fwApp.SetupInspectorShortcut()

	fwApp.Resize(120, 40)
	fwApp.InitTheme("nord")

	declarativeRoot := render.NewDeclarativeNodeFromFunc(RuntimeDemoWithInspectorOverlay)
	declarativeRoot.SetFrameworkApp(fwApp)
	fwApp.SetRoot(declarativeRoot)

	go func() {
		if err := fwApp.Run(); err != nil {
			t.Errorf("fwApp.Run() error: %v", err)
		}
	}()

	// Wait for app to start
	for i := 0; i < 200; i++ {
		if fwApp.GetState() == framework.StateRunning {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	if fwApp.GetState() != framework.StateRunning {
		t.Fatal("Framework app did not start")
	}

	time.Sleep(300 * time.Millisecond)

	// Test 1: Plain "k" key
	t.Log("\n=== Test 1: Pressing 'k' ===")
	if globalInspector.HandleKeyEvent("k", false, false, false) {
		t.Log("✅ Inspector handled 'k' key")
	} else {
		t.Error("❌ Inspector did NOT handle 'k' key")
	}

	// Test 2: Alt+K
	t.Log("\n=== Test 2: Pressing Alt+K ===")
	if globalInspector.HandleKeyEvent("k", true, false, false) {
		t.Log("✅ Inspector handled Alt+K")
	} else {
		t.Error("❌ Inspector did NOT handle Alt+K")
	}

	// Test 3: Ctrl+D (toggles debug mode)
	t.Log("\n=== Test 3: Pressing Ctrl+D ===")
	if globalInspector.HandleKeyEvent("d", false, true, false) {
		t.Log("✅ Inspector handled Ctrl+D")
	} else {
		t.Error("❌ Inspector did NOT handle Ctrl+D")
	}

	// Test 4: Shift+K
	t.Log("\n=== Test 4: Pressing Shift+K ===")
	if globalInspector.HandleKeyEvent("K", false, false, true) {
		t.Log("✅ Inspector handled Shift+K")
	} else {
		t.Error("❌ Inspector did NOT handle Shift+K")
	}

	// Test 5: F12
	t.Log("\n=== Test 5: Pressing F12 ===")
	if globalInspector.HandleKeyEvent("f12", false, false, false) {
		t.Log("✅ Inspector handled F12")
	} else {
		t.Error("❌ Inspector did NOT handle F12")
	}

	t.Log("\n=== Key Detection Test Complete ===")
	t.Log("All keys were received by Inspector!")

	fwApp.Quit()
}
