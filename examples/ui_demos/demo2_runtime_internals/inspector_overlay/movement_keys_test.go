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

// TestInspectorMovementKeys tests Alt+H/J/K/L movement keys
func TestInspectorMovementKeys(t *testing.T) {
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
	fwApp.SetupInspectorShortcut() // Enables Alt+H/J/K/L

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

	// Get initial position
	initialX, initialY := globalInspector.GetPosition()
	t.Logf("Initial Inspector position: (%d, %d)", initialX, initialY)

	// Test Alt+K (move up)
	t.Log("\n=== Testing Alt+K (move up) ===")
	rawAltK := platform.RawInput{
		Type:    platform.InputKeyPress,
		Special: platform.KeyK,
		Modifiers: platform.ModAlt,
	}
	if err := fwApp.InjectEvent(rawAltK); err != nil {
		t.Errorf("Failed to inject Alt+K: %v", err)
	}
	time.Sleep(200 * time.Millisecond)

	afterKX, afterKY := globalInspector.GetPosition()
	t.Logf("After Alt+K, position: (%d, %d)", afterKX, afterKY)

	if afterKY >= initialY {
		t.Errorf("Alt+K should move panel up (Y should decrease), but Y went from %d to %d", initialY, afterKY)
	} else {
		t.Logf("✅ Alt+K moved panel up: Y %d → %d", initialY, afterKY)
	}

	// Test Alt+J (move down)
	t.Log("\n=== Testing Alt+J (move down) ===")
	rawAltJ := platform.RawInput{
		Type:    platform.InputKeyPress,
		Special: platform.KeyJ,
		Modifiers: platform.ModAlt,
	}
	if err := fwApp.InjectEvent(rawAltJ); err != nil {
		t.Errorf("Failed to inject Alt+J: %v", err)
	}
	time.Sleep(200 * time.Millisecond)

	afterJX, afterJY := globalInspector.GetPosition()
	t.Logf("After Alt+J, position: (%d, %d)", afterJX, afterJY)

	if afterJY <= afterKY {
		t.Errorf("Alt+J should move panel down (Y should increase), but Y went from %d to %d", afterKY, afterJY)
	} else {
		t.Logf("✅ Alt+J moved panel down: Y %d → %d", afterKY, afterJY)
	}

	// Test Alt+H (move left)
	t.Log("\n=== Testing Alt+H (move left) ===")
	rawAltH := platform.RawInput{
		Type:    platform.InputKeyPress,
		Special: platform.KeyH,
		Modifiers: platform.ModAlt,
	}
	if err := fwApp.InjectEvent(rawAltH); err != nil {
		t.Errorf("Failed to inject Alt+H: %v", err)
	}
	time.Sleep(200 * time.Millisecond)

	afterHX, afterHY := globalInspector.GetPosition()
	t.Logf("After Alt+H, position: (%d, %d)", afterHX, afterHY)

	if afterHX >= afterJX {
		t.Errorf("Alt+H should move panel left (X should decrease), but X went from %d to %d", afterJX, afterHX)
	} else {
		t.Logf("✅ Alt+H moved panel left: X %d → %d", afterJX, afterHX)
	}

	// Test Alt+L (move right)
	t.Log("\n=== Testing Alt+L (move right) ===")
	rawAltL := platform.RawInput{
		Type:    platform.InputKeyPress,
		Special: platform.KeyL,
		Modifiers: platform.ModAlt,
	}
	if err := fwApp.InjectEvent(rawAltL); err != nil {
		t.Errorf("Failed to inject Alt+L: %v", err)
	}
	time.Sleep(200 * time.Millisecond)

	afterLX, afterLY := globalInspector.GetPosition()
	t.Logf("After Alt+L, position: (%d, %d)", afterLX, afterLY)

	if afterLX <= afterHX {
		t.Errorf("Alt+L should move panel right (X should increase), but X went from %d to %d", afterHX, afterLX)
	} else {
		t.Logf("✅ Alt+L moved panel right: X %d → %d", afterHX, afterLX)
	}

	t.Log("\n=== Movement Keys Test Complete ===")
	t.Log("✅ All Alt+H/J/K/L movement keys work!")

	fwApp.Quit()
}

// TestInspectorArrowMovement tests Alt+Arrow keys for movement
func TestInspectorArrowMovement(t *testing.T) {
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

	// Get initial position
	initialX, initialY := globalInspector.GetPosition()
	t.Logf("Initial position: (%d, %d)", initialX, initialY)

	// Test Alt+Up (move up)
	t.Log("\n=== Testing Alt+Up (move up) ===")
	rawAltUp := platform.RawInput{
		Type:    platform.InputKeyPress,
		Special: platform.KeyUp,
		Modifiers: platform.ModAlt,
	}
	fwApp.InjectEvent(rawAltUp)
	time.Sleep(200 * time.Millisecond)

	afterUpX, afterUpY := globalInspector.GetPosition()
	t.Logf("After Alt+Up: (%d, %d)", afterUpX, afterUpY)

	if afterUpY >= initialY {
		t.Errorf("Alt+Up should move panel up, but Y went from %d to %d", initialY, afterUpY)
	} else {
		t.Logf("✅ Alt+Up moved panel up: Y %d → %d", initialY, afterUpY)
	}

	// Test Alt+Down (move down)
	t.Log("\n=== Testing Alt+Down (move down) ===")
	rawAltDown := platform.RawInput{
		Type:    platform.InputKeyPress,
		Special: platform.KeyDown,
		Modifiers: platform.ModAlt,
	}
	fwApp.InjectEvent(rawAltDown)
	time.Sleep(200 * time.Millisecond)

	afterDownX, afterDownY := globalInspector.GetPosition()
	t.Logf("After Alt+Down: (%d, %d)", afterDownX, afterDownY)

	if afterDownY <= afterUpY {
		t.Errorf("Alt+Down should move panel down, but Y went from %d to %d", afterUpY, afterDownY)
	} else {
		t.Logf("✅ Alt+Down moved panel down: Y %d → %d", afterUpY, afterDownY)
	}

	t.Log("\n=== Arrow Movement Test Complete ===")
	t.Log("✅ Alt+Arrow keys work!")

	fwApp.Quit()
}
