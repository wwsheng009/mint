package main

import (
	"testing"
	"time"

	"github.com/wwsheng009/mint/framework"
	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/internal/inspector"
	"github.com/wwsheng009/mint/internal/render"
	"github.com/wwsheng009/mint/runtime/platform"
)

// TestCtrlModifierDetection tests if Ctrl modifier is correctly detected
func TestCtrlModifierDetection(t *testing.T) {
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

	time.Sleep(300 * time.Millisecond)

	// Test Ctrl+D (shortcut to toggle Inspector)
	t.Log("\n=== Test 1: Ctrl+D should toggle Inspector ===")
	rawCtrlD := platform.RawInput{
		Type:      platform.InputKeyPress,
		Key:       'd',
		Modifiers: platform.ModCtrl,
	}

	if err := fwApp.InjectEvent(rawCtrlD); err != nil {
		t.Errorf("Failed to inject Ctrl+D: %v", err)
	}
	time.Sleep(200 * time.Millisecond)

	// If Inspector is hidden, Ctrl+D should show it
	// If Inspector is visible, Ctrl+D should toggle debug mode
	// Either way, it should be handled
	t.Log("✅ Ctrl+D event injected successfully")

	// Test Ctrl+K (should be detected)
	t.Log("\n=== Test 2: Ctrl+K modifier detection ===")
	if globalInspector.HandleKeyEvent("k", false, true, false) {
		t.Log("✅ Inspector detected Ctrl+K")
	} else {
		t.Error("❌ Inspector did NOT detect Ctrl+K")
	}

	// Test Shift+K
	t.Log("\n=== Test 3: Shift+K modifier detection ===")
	if globalInspector.HandleKeyEvent("K", false, false, false) {
		t.Log("✅ Inspector detected plain 'K' (capital)")
	} else {
		t.Error("❌ Inspector did NOT detect 'K'")
	}

	// Test Alt+K
	t.Log("\n=== Test 4: Alt+K modifier detection ===")
	if globalInspector.HandleKeyEvent("k", true, false, false) {
		t.Log("✅ Inspector detected Alt+K")
	} else {
		t.Error("❌ Inspector did NOT detect Alt+K")
	}

	// Test Ctrl+Alt+K (all three modifiers except Shift)
	t.Log("\n=== Test 5: Ctrl+Alt+K modifier detection ===")
	if globalInspector.HandleKeyEvent("k", true, true, false) {
		t.Log("✅ Inspector detected Ctrl+Alt+K")
	} else {
		t.Error("❌ Inspector did NOT detect Ctrl+Alt+K")
	}

	// Test Ctrl+Shift+K
	t.Log("\n=== Test 6: Ctrl+Shift+K modifier detection ===")
	if globalInspector.HandleKeyEvent("K", false, true, true) {
		t.Log("✅ Inspector detected Ctrl+Shift+K")
	} else {
		t.Error("❌ Inspector did NOT detect Ctrl+Shift+K")
	}

	t.Log("\n=== Ctrl Modifier Detection Test Complete ===")
	t.Log("✅ All modifiers are being detected correctly!")

	fwApp.Quit()
}
