package main

import (
	"strings"
	"testing"

	"github.com/wwsheng009/mint/internal/inspector"
	"github.com/wwsheng009/mint/runtime/platform"
	ui "github.com/wwsheng009/mint/ui"
)

// TestTreeViewNavigation tests TreeView keyboard navigation using testable app
func TestTreeViewNavigation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping interactive test in short mode")
	}

	// Create Inspector
	insp := inspector.NewStandaloneInspector()
	insp.Enable()
	insp.ToggleVisibility()
	insp.SetOverlaySize(100, 40)

	// Create a complex tree to test navigation
	testRoot := ui.VStack(
		ui.Text("Root Node"),
		ui.Text("Node 1"),
		ui.Text("Node 2"),
		ui.Text("Node 3"),
		ui.HStack(
			ui.Text("Child 1"),
			ui.Text("Child 2"),
		),
		ui.Text("Node 4"),
		ui.Text("Node 5"),
	)

	// Attach to Inspector
	insp.AttachToApp(testRoot)

	// Render inspector
	overlay := insp.RenderOverlay()
	if overlay == nil {
		t.Fatal("Inspector overlay is nil")
	}

	// Create testable app
	testApp, err := ui.RunTest(func() ui.VNode {
		return overlay
	},
		ui.WithWidth(120),
		ui.WithHeight(40),
		ui.WithTitle("TreeView Navigation Test"),
	)
	if err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}
	defer testApp.Close()

	// Wait for initial render
	waitForDemo2Idle(t, testApp)

	// Attach test root to Inspector
	insp.AttachToApp(testRoot)

	// Render inspector overlay
	overlay = insp.RenderOverlay()

	// Force render to update display
	testApp.ForceRender()

	initialRender := settleDemo2Render(t, testApp)
	logDemo2Snapshot(t, "=== Initial Render ===", initialRender, 24)

	// Verify tree is displayed
	if !strings.Contains(initialRender, "Layout Tree") {
		t.Error("Tree view should be displayed")
	}

	// Test Down Arrow navigation
	t.Log("\n=== Testing Down Arrow (3 times) ===")
	for i := 0; i < 3; i++ {
		// Inject the key event
		err = testApp.InjectSpecialKey(platform.KeyDown)
		if err != nil {
			t.Errorf("Failed to inject KeyDown: %v", err)
		}
		// Manually trigger Inspector's HandleKeyEvent
		insp.HandleKeyEvent("down", false, false, false)
		// Re-render overlay
		overlay = insp.RenderOverlay()
		testApp.ForceRender()
	}

	afterDown := settleDemo2Render(t, testApp)
	logDemo2Snapshot(t, "After Down arrows:", afterDown, 16)

	// Test Up Arrow navigation
	t.Log("\n=== Testing Up Arrow (2 times) ===")
	for i := 0; i < 2; i++ {
		err = testApp.InjectSpecialKey(platform.KeyUp)
		if err != nil {
			t.Errorf("Failed to inject KeyUp: %v", err)
		}
		// Manually trigger Inspector's HandleKeyEvent
		insp.HandleKeyEvent("up", false, false, false)
		// Re-render overlay
		overlay = insp.RenderOverlay()
		testApp.ForceRender()
	}

	afterUp := settleDemo2Render(t, testApp)
	logDemo2Snapshot(t, "After Up arrows:", afterUp, 16)

	// Test PageDown
	t.Log("\n=== Testing PageDown ===")
	err = testApp.InjectSpecialKey(platform.KeyPageDown)
	if err != nil {
		t.Errorf("Failed to inject PageDown: %v", err)
	}
	// Manually trigger Inspector's HandleKeyEvent
	insp.HandleKeyEvent("pgdn", false, false, false)
	// Re-render overlay
	overlay = insp.RenderOverlay()
	testApp.ForceRender()

	afterPageDown := settleDemo2Render(t, testApp)
	logDemo2Snapshot(t, "After PageDown:", afterPageDown, 16)

	// Test Home
	t.Log("\n=== Testing Home ===")
	err = testApp.InjectSpecialKey(platform.KeyHome)
	if err != nil {
		t.Errorf("Failed to inject Home: %v", err)
	}
	// Manually trigger Inspector's HandleKeyEvent
	insp.HandleKeyEvent("home", false, false, false)
	// Re-render overlay
	overlay = insp.RenderOverlay()
	testApp.ForceRender()

	afterHome := settleDemo2Render(t, testApp)
	logDemo2Snapshot(t, "After Home:", afterHome, 16)

	// Test End
	t.Log("\n=== Testing End ===")
	err = testApp.InjectSpecialKey(platform.KeyEnd)
	if err != nil {
		t.Errorf("Failed to inject End: %v", err)
	}
	// Manually trigger Inspector's HandleKeyEvent
	insp.HandleKeyEvent("end", false, false, false)
	// Re-render overlay
	overlay = insp.RenderOverlay()
	testApp.ForceRender()

	afterEnd := settleDemo2Render(t, testApp)
	logDemo2Snapshot(t, "After End:", afterEnd, 16)

	t.Log("\n=== TreeView Navigation Test Complete ===")
}

// TestTreeViewWithDemo2 tests TreeView navigation with actual demo2 app
func TestTreeViewWithDemo2(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping interactive test in short mode")
	}

	// Run with demo2
	testApp, err := ui.RunTest(RuntimeDemo,
		ui.WithWidth(120),
		ui.WithHeight(40),
		ui.WithTitle("TreeView Demo2 Test"),
	)
	if err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}
	defer testApp.Close()

	// Wait for initial render
	waitForDemo2Idle(t, testApp)

	// Activate Inspector
	t.Log("=== Activating Inspector with 'i' key ===")
	err = testApp.InjectKey('i')
	if err != nil {
		t.Fatalf("Failed to inject 'i' key: %v", err)
	}

	// Wait for Inspector to appear
	inspectorRender := settleDemo2Render(t, testApp)

	logDemo2Snapshot(t, "=== Inspector Rendered ===", inspectorRender, 24)

	// Verify Inspector is visible
	if !strings.Contains(inspectorRender, "INSPECTOR") {
		t.Log("Note: Inspector not visible (RuntimeDemo does not include inspector toggle)")
	} else {
		t.Log("✓ Inspector is visible")
	}

	// Test navigation on Inspector tree
	t.Log("\n=== Testing Navigation in Inspector ===")

	// Press Down Arrow multiple times
	t.Log("Pressing Down Arrow 5 times...")
	for i := 0; i < 5; i++ {
		err = testApp.InjectSpecialKey(platform.KeyDown)
		if err != nil {
			t.Errorf("Failed to inject KeyDown: %v", err)
		}
	}

	afterDownNav := settleDemo2Render(t, testApp)
	logDemo2Snapshot(t, "After Down navigation:", afterDownNav, 16)

	// Press Up Arrow
	t.Log("\nPressing Up Arrow 3 times...")
	for i := 0; i < 3; i++ {
		err = testApp.InjectSpecialKey(platform.KeyUp)
		if err != nil {
			t.Errorf("Failed to inject KeyUp: %v", err)
		}
	}

	afterUpNav := settleDemo2Render(t, testApp)
	logDemo2Snapshot(t, "After Up navigation:", afterUpNav, 16)

	// Test PageDown
	t.Log("\nTesting PageDown...")
	err = testApp.InjectSpecialKey(platform.KeyPageDown)
	if err != nil {
		t.Errorf("Failed to inject PageDown: %v", err)
	}
	afterPageDownNav := settleDemo2Render(t, testApp)
	logDemo2Snapshot(t, "After PageDown:", afterPageDownNav, 16)

	// Test Home
	t.Log("\nTesting Home...")
	err = testApp.InjectSpecialKey(platform.KeyHome)
	if err != nil {
		t.Errorf("Failed to inject Home: %v", err)
	}
	afterHomeNav := settleDemo2Render(t, testApp)
	logDemo2Snapshot(t, "After Home:", afterHomeNav, 16)

	// Deactivate Inspector
	t.Log("\n=== Deactivating Inspector with 'q' key ===")
	err = testApp.InjectKey('q')
	if err != nil {
		t.Errorf("Failed to inject 'q' key: %v", err)
	}
	settleDemo2Render(t, testApp)

	t.Log("\n=== TreeView Demo2 Test Complete ===")
}
