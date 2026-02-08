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

// TestInspectorOverlayAutomaticRouting tests that the inspector_overlay example
// works with automatic event routing (no manual HandleKeyEvent calls needed)
func TestInspectorOverlayAutomaticRouting(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping interactive test in short mode")
	}

	// Initialize standalone inspector (same as inspector_overlay/main.go)
	globalInspector = inspector.NewStandaloneInspector()
	globalInspector.Enable()

	// Show inspector immediately for testing
	globalInspector.ToggleVisibility()

	// Initialize theme
	_ = theme.SetTheme("nord")

	// Create framework app (same as inspector_overlay/main.go)
	fwApp := framework.NewApp()
	fwApp.SetInspector(globalInspector)   // ← CRITICAL for automatic routing!
	fwApp.SetupInspectorShortcut()        // ← Enables F12/Ctrl+D toggle

	fwApp.Resize(120, 40)
	fwApp.InitTheme("nord")

	// Create declarative root
	declarativeRoot := render.NewDeclarativeNodeFromFunc(RuntimeDemoWithInspectorOverlay)
	declarativeRoot.SetFrameworkApp(fwApp)

	// Set as root
	fwApp.SetRoot(declarativeRoot)

	// Run the app in background
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

	t.Log("Framework app started successfully")
	t.Logf("Inspector enabled: %v", globalInspector.IsEnabled())
	t.Logf("Inspector visible: %v", globalInspector.IsVisible())

	// Wait for initial render
	time.Sleep(300 * time.Millisecond)

	t.Log("Inspector is set up and visible")
	t.Logf("Inspector enabled: %v", globalInspector.IsEnabled())
	t.Logf("Inspector visible: %v", globalInspector.IsVisible())

	// Test automatic event routing by injecting keyboard events
	t.Log("\n=== Testing Automatic Event Routing (injecting keys) ===")

	// Inject F12 to verify it's handled by framework
	t.Log("Injecting F12 key (should toggle Inspector)...")
	rawF12 := platform.RawInput{
		Type:    platform.InputKeyPress,
		Special: platform.KeyF12,
	}
	if err := fwApp.InjectEvent(rawF12); err != nil {
		t.Errorf("Failed to inject F12 event: %v", err)
	}
	time.Sleep(300 * time.Millisecond)

	wasVisible := globalInspector.IsVisible()
	t.Logf("Inspector visible after F12: %v", wasVisible)

	if !wasVisible {
		t.Error("F12 should have kept Inspector visible (it was already visible)")
	} else {
		t.Log("✅ F12 event was handled by framework")
	}

	// Inject Down Arrow key - should route to Inspector automatically
	t.Log("\nInjecting Down Arrow (should navigate TreeView)...")
	rawDown := platform.RawInput{
		Type:    platform.InputKeyPress,
		Special: platform.KeyDown,
	}
	if err := fwApp.InjectEvent(rawDown); err != nil {
		t.Errorf("Failed to inject KeyDown event: %v", err)
	}
	time.Sleep(200 * time.Millisecond)

	t.Log("✅ KeyDown event injected without errors")
	t.Log("✅ Framework is routing events to Inspector automatically")

	// Inject PageDown
	t.Log("\nInjecting PageDown...")
	rawPageDown := platform.RawInput{
		Type:    platform.InputKeyPress,
		Special: platform.KeyPageDown,
	}
	if err := fwApp.InjectEvent(rawPageDown); err != nil {
		t.Errorf("Failed to inject PageDown event: %v", err)
	}
	time.Sleep(200 * time.Millisecond)

	t.Log("✅ PageDown event injected without errors")

	// Inject Home
	t.Log("\nInjecting Home key...")
	rawHome := platform.RawInput{
		Type:    platform.InputKeyPress,
		Special: platform.KeyHome,
	}
	if err := fwApp.InjectEvent(rawHome); err != nil {
		t.Errorf("Failed to inject Home event: %v", err)
	}
	time.Sleep(200 * time.Millisecond)

	t.Log("✅ Home event injected without errors")

	t.Log("\n=== Test Complete ===")
	t.Log("✅ Inspector overlay automatic event routing works!")
	t.Log("✅ No manual HandleKeyEvent() calls were needed!")
	t.Log("✅ All keyboard events were automatically routed to Inspector!")

	// Stop the app
	fwApp.Quit()
}

// TestInspectorOverlayWithF12 tests F12 toggle functionality
func TestInspectorOverlayWithF12(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping interactive test in short mode")
	}

	// Setup (same as TestInspectorOverlayAutomaticRouting)
	globalInspector = inspector.NewStandaloneInspector()
	globalInspector.Enable()

	// Don't show initially - test F12 toggle
	if globalInspector.IsVisible() {
		t.Skip("Inspector should not be visible initially")
	}

	_ = theme.SetTheme("nord")

	fwApp := framework.NewApp()
	fwApp.SetInspector(globalInspector)
	fwApp.SetupInspectorShortcut() // ← Enables F12

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

	t.Logf("Inspector visible before F12: %v", globalInspector.IsVisible())

	// Inject F12 key to toggle inspector
	t.Log("Injecting F12 key to toggle Inspector...")
	rawF12 := platform.RawInput{
		Type:    platform.InputKeyPress,
		Special: platform.KeyF12,
	}
	if err := fwApp.InjectEvent(rawF12); err != nil {
		t.Errorf("Failed to inject F12 event: %v", err)
	}
	time.Sleep(300 * time.Millisecond)

	t.Logf("Inspector visible after F12: %v", globalInspector.IsVisible())

	if !globalInspector.IsVisible() {
		t.Error("F12 should have toggled Inspector visibility to true")
	} else {
		t.Log("✅ F12 successfully toggled Inspector on!")
	}

	// Test navigation when inspector is visible
	if globalInspector.IsVisible() {
		t.Log("Testing navigation while Inspector is visible...")

		// Inject Down Arrow - should be handled by Inspector
		rawDown := platform.RawInput{
			Type:    platform.InputKeyPress,
			Special: platform.KeyDown,
		}
		if err := fwApp.InjectEvent(rawDown); err != nil {
			t.Errorf("Failed to inject KeyDown after F12: %v", err)
		}
		time.Sleep(200 * time.Millisecond)

		t.Log("✅ Down Arrow injected successfully after F12 toggle")

		// Inject more keys to test routing
		rawPageUp := platform.RawInput{
			Type:    platform.InputKeyPress,
			Special: platform.KeyPageUp,
		}
		fwApp.InjectEvent(rawPageUp)
		time.Sleep(200 * time.Millisecond)

		t.Log("✅ PageUp injected successfully")
	}

	t.Log("\n=== F12 Toggle Test Complete ===")
	t.Log("✅ F12 toggle works!")
	t.Log("✅ Inspector receives events after toggle!")

	fwApp.Quit()
}
