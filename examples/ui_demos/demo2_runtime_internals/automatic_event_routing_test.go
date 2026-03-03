package main

import (
	"strings"
	"testing"
	"time"

	"github.com/wwsheng009/mint/internal/inspector"
	"github.com/wwsheng009/mint/runtime/platform"
	ui "github.com/wwsheng009/mint/ui"
)

// TestAutomaticEventRouting tests that Inspector automatically receives keyboard events
// from the framework's event routing system without manual HandleKeyEvent calls.
func TestAutomaticEventRouting(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping interactive test in short mode")
	}

	// Create Inspector
	insp := inspector.NewStandaloneInspector()
	insp.Enable()
	insp.ToggleVisibility()
	insp.SetOverlaySize(100, 40)

	// Create a test tree
	testRoot := ui.VStack(
		ui.Text("Root Node"),
		ui.Text("Node 1"),
		ui.Text("Node 2"),
		ui.Text("Node 3"),
	)

	// Attach to Inspector
	insp.AttachToApp(testRoot)

	// Render inspector overlay
	overlay := insp.RenderOverlay()
	if overlay == nil {
		t.Fatal("Inspector overlay is nil")
	}

	// Create testable app with Inspector overlay
	testApp, err := ui.RunTest(func() ui.VNode {
		return overlay
	},
		ui.WithWidth(120),
		ui.WithHeight(40),
		ui.WithTitle("Automatic Event Routing Test"),
	)
	if err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}
	defer testApp.Close()

	// CRITICAL: Register Inspector with framework app!
	// This enables automatic event routing from platform input → Inspector
	fwApp := testApp.GetFrameworkApp()
	fwApp.SetInspector(insp)
	fwApp.SetupInspectorShortcut()

	t.Logf("Inspector registered with framework app (visible=%v)", insp.IsVisible())

	// Wait for initial render
	time.Sleep(300 * time.Millisecond)

	initialRender := testApp.GetRenderString()
	t.Logf("=== Initial Render ===")
	if !strings.Contains(initialRender, "Layout Tree") {
		t.Error("Tree view should be displayed")
	}

	// Get TreeView component to verify state changes
	treeView := insp.GetTreeViewComponent()
	if treeView == nil {
		t.Fatal("TreeView component is nil")
	}

	initialFocus := treeView.GetSelectedIndex()
	t.Logf("Initial focus index: %d", initialFocus)

	// Test automatic event routing with Down Arrow
	t.Log("\n=== Testing Automatic Event Routing (Down Arrow) ===")
	for i := 0; i < 3; i++ {
		// Inject key - NO MANUAL HandleKeyEvent CALL!
		err = testApp.InjectSpecialKey(platform.KeyDown)
		if err != nil {
			t.Errorf("Failed to inject KeyDown: %v", err)
		}

		// Just wait - Inspector automatically receives the event
		time.Sleep(150 * time.Millisecond)
	}

	afterDown := treeView.GetSelectedIndex()
	t.Logf("After 3 Down arrows, focus index: %d", afterDown)

	if afterDown <= initialFocus {
		t.Errorf("Focus should have moved down from %d, but got %d", initialFocus, afterDown)
	}

	// Test automatic event routing with Up Arrow
	t.Log("\n=== Testing Automatic Event Routing (Up Arrow) ===")
	for i := 0; i < 2; i++ {
		err = testApp.InjectSpecialKey(platform.KeyUp)
		if err != nil {
			t.Errorf("Failed to inject KeyUp: %v", err)
		}
		time.Sleep(150 * time.Millisecond)
	}

	afterUp := treeView.GetSelectedIndex()
	t.Logf("After 2 Up arrows, focus index: %d", afterUp)

	if afterUp >= afterDown {
		t.Errorf("Focus should have moved up from %d, but got %d", afterDown, afterUp)
	}

	// Test PageDown
	t.Log("\n=== Testing Automatic Event Routing (PageDown) ===")
	err = testApp.InjectSpecialKey(platform.KeyPageDown)
	if err != nil {
		t.Errorf("Failed to inject PageDown: %v", err)
	}
	time.Sleep(200 * time.Millisecond)

	afterPageDown := treeView.GetSelectedIndex()
	t.Logf("After PageDown, focus index: %d", afterPageDown)

	if afterPageDown <= afterUp {
		t.Logf("Note: PageDown may not have moved focus if tree is small (current: %d)", afterPageDown)
	}

	// Test Home
	t.Log("\n=== Testing Automatic Event Routing (Home) ===")
	err = testApp.InjectSpecialKey(platform.KeyHome)
	if err != nil {
		t.Errorf("Failed to inject Home: %v", err)
	}
	time.Sleep(200 * time.Millisecond)

	afterHome := treeView.GetSelectedIndex()
	t.Logf("After Home, focus index: %d", afterHome)

	if afterHome != 0 {
		t.Errorf("Home should jump to top (focus=0), but got %d", afterHome)
	}

	// Test End
	t.Log("\n=== Testing Automatic Event Routing (End) ===")
	err = testApp.InjectSpecialKey(platform.KeyEnd)
	if err != nil {
		t.Errorf("Failed to inject End: %v", err)
	}
	time.Sleep(200 * time.Millisecond)

	afterEnd := treeView.GetSelectedIndex()
	t.Logf("After End, focus index: %d", afterEnd)

	if afterEnd < afterHome {
		t.Errorf("End should jump to bottom, but got %d", afterEnd)
	}

	t.Log("\n=== Automatic Event Routing Test Complete ===")
	t.Log("✅ Inspector automatically received all keyboard events!")
	t.Log("✅ No manual HandleKeyEvent() calls were needed!")
}

// TestAutomaticEventRoutingWithDemo2 tests automatic routing with actual demo2 app
func TestAutomaticEventRoutingWithDemo2(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping interactive test in short mode")
	}

	// This test will verify that demo2 works with automatic event routing
	// once demo2/main.go is updated to create and register the Inspector

	t.Skip("Test will be enabled after demo2/main.go adds Inspector setup")

	// Expected pattern for demo2/main.go:
	/*
		func main() {
			_ = theme.SetTheme("nord")

			// Create Inspector
			globalInspector := inspector.NewStandaloneInspector()
			globalInspector.Enable()

			// Create framework app
			fwApp := framework.NewApp()
			fwApp.Resize(100, 35)
			fwApp.InitTheme("dark")

			// Register Inspector
			fwApp.SetInspector(globalInspector)
			fwApp.SetupInspectorShortcut()

			// Set root with Fiber reconciler enabled
			declarativeRoot := render.NewDeclarativeNodeFromFuncWithFiber(RuntimeDemo)
    		declarativeRoot.SetApp(fwApp)
			fwApp.SetRoot(declarativeRoot)

			// Run
			fwApp.Run()
		}
	*/
}
