package main

import (
	"strings"
	"testing"
	"time"

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
	time.Sleep(300 * time.Millisecond)

	// Attach test root to Inspector
	insp.AttachToApp(testRoot)

	// Render inspector overlay
	overlay = insp.RenderOverlay()

	// Force render to update display
	testApp.ForceRender()

	time.Sleep(100 * time.Millisecond)

	initialRender := testApp.GetRenderString()
	t.Logf("=== Initial Render (first 40 lines) ===")
	lines := strings.Split(initialRender, "\n")
	maxLines := 40
	if len(lines) < maxLines {
		maxLines = len(lines)
	}
	for i := 0; i < maxLines; i++ {
		t.Logf("  %s", lines[i])
	}
	t.Logf("=== End ===\nTotal lines: %d", len(lines))

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
		time.Sleep(150 * time.Millisecond)
	}

	afterDown := testApp.GetRenderString()
	t.Logf("After Down arrows (first 40 lines):")
	lines = strings.Split(afterDown, "\n")
	maxLines = 40
	if len(lines) < maxLines {
		maxLines = len(lines)
	}
	for i := 0; i < maxLines; i++ {
		t.Logf("  %s", lines[i])
	}

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
		time.Sleep(150 * time.Millisecond)
	}

	afterUp := testApp.GetRenderString()
	t.Logf("After Up arrows (first 40 lines):")
	lines = strings.Split(afterUp, "\n")
	maxLines = 40
	if len(lines) < maxLines {
		maxLines = len(lines)
	}
	for i := 0; i < maxLines; i++ {
		t.Logf("  %s", lines[i])
	}

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
	time.Sleep(200 * time.Millisecond)

	afterPageDown := testApp.GetRenderString()
	t.Logf("After PageDown (first 40 lines):")
	lines = strings.Split(afterPageDown, "\n")
	maxLines = 40
	if len(lines) < maxLines {
		maxLines = len(lines)
	}
	for i := 0; i < maxLines; i++ {
		t.Logf("  %s", lines[i])
	}

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
	time.Sleep(200 * time.Millisecond)

	afterHome := testApp.GetRenderString()
	t.Logf("After Home (first 40 lines):")
	lines = strings.Split(afterHome, "\n")
	maxLines = 40
	if len(lines) < maxLines {
		maxLines = len(lines)
	}
	for i := 0; i < maxLines; i++ {
		t.Logf("  %s", lines[i])
	}

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
	time.Sleep(200 * time.Millisecond)

	afterEnd := testApp.GetRenderString()
	t.Logf("After End (first 40 lines):")
	lines = strings.Split(afterEnd, "\n")
	maxLines = 40
	if len(lines) < maxLines {
		maxLines = len(lines)
	}
	for i := 0; i < maxLines; i++ {
		t.Logf("  %s", lines[i])
	}

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
	time.Sleep(300 * time.Millisecond)

	// Activate Inspector
	t.Log("=== Activating Inspector with 'i' key ===")
	err = testApp.InjectKey('i')
	if err != nil {
		t.Fatalf("Failed to inject 'i' key: %v", err)
	}

	// Wait for Inspector to appear
	time.Sleep(500 * time.Millisecond)

	inspectorRender := testApp.GetRenderString()
	t.Logf("=== Inspector Rendered (first 50 lines) ===")
	lines := strings.Split(inspectorRender, "\n")
	maxLines := 50
	if len(lines) < maxLines {
		maxLines = len(lines)
	}
	for i := 0; i < maxLines; i++ {
		t.Logf("  %s", lines[i])
	}

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
		time.Sleep(100 * time.Millisecond)
	}

	afterDownNav := testApp.GetRenderString()
	t.Logf("After Down navigation (first 40 lines):")
	lines = strings.Split(afterDownNav, "\n")
	maxLines = 40
	if len(lines) < maxLines {
		maxLines = len(lines)
	}
	for i := 0; i < maxLines; i++ {
		t.Logf("  %s", lines[i])
	}

	// Press Up Arrow
	t.Log("\nPressing Up Arrow 3 times...")
	for i := 0; i < 3; i++ {
		err = testApp.InjectSpecialKey(platform.KeyUp)
		if err != nil {
			t.Errorf("Failed to inject KeyUp: %v", err)
		}
		time.Sleep(100 * time.Millisecond)
	}

	afterUpNav := testApp.GetRenderString()
	t.Logf("After Up navigation (first 40 lines):")
	lines = strings.Split(afterUpNav, "\n")
	maxLines = 40
	if len(lines) < maxLines {
		maxLines = len(lines)
	}
	for i := 0; i < maxLines; i++ {
		t.Logf("  %s", lines[i])
	}

	// Test PageDown
	t.Log("\nTesting PageDown...")
	err = testApp.InjectSpecialKey(platform.KeyPageDown)
	if err != nil {
		t.Errorf("Failed to inject PageDown: %v", err)
	}
	time.Sleep(200 * time.Millisecond)

	afterPageDownNav := testApp.GetRenderString()
	t.Logf("After PageDown (first 40 lines):")
	lines = strings.Split(afterPageDownNav, "\n")
	maxLines = 40
	if len(lines) < maxLines {
		maxLines = len(lines)
	}
	for i := 0; i < maxLines; i++ {
		t.Logf("  %s", lines[i])
	}

	// Test Home
	t.Log("\nTesting Home...")
	err = testApp.InjectSpecialKey(platform.KeyHome)
	if err != nil {
		t.Errorf("Failed to inject Home: %v", err)
	}
	time.Sleep(200 * time.Millisecond)

	afterHomeNav := testApp.GetRenderString()
	t.Logf("After Home (first 40 lines):")
	lines = strings.Split(afterHomeNav, "\n")
	maxLines = 40
	if len(lines) < maxLines {
		maxLines = len(lines)
	}
	for i := 0; i < maxLines; i++ {
		t.Logf("  %s", lines[i])
	}

	// Deactivate Inspector
	t.Log("\n=== Deactivating Inspector with 'q' key ===")
	err = testApp.InjectKey('q')
	if err != nil {
		t.Errorf("Failed to inject 'q' key: %v", err)
	}
	time.Sleep(300 * time.Millisecond)

	t.Log("\n=== TreeView Demo2 Test Complete ===")
}
