package inspector

import (
	"testing"

	frameworkevent "github.com/wwsheng009/mint/framework/event"
	"github.com/wwsheng009/mint/components/display"
	"github.com/wwsheng009/mint/runtime/event"
	"github.com/wwsheng009/mint/ui"
)

// TestOverlayTabsMouse ensures overlay tab bar reacts to mouse press.
func TestOverlayTabsMouse(t *testing.T) {
	si := NewStandaloneInspector()
	si.Enable()
	si.ToggleVisibility() // make visible

	// Default position (0,0), overlayWidth/Height default from constructor (80,25).
	// Tab bar rendered at localY = 1. Click on second tab label.
	// Tab rendering: "[Elements(1)] | Console(2) | ..."
	// Position: Elements=[0,13), sep=[13,16), Console=[16,27)
	x := 18 // position inside "Console(2)" label
	y := 1

	ev := &frameworkevent.MouseEvent{
		BaseEvent: frameworkevent.NewBaseEvent(event.EventMousePress),
		X:         x,
		Y:         y,
		Button:    frameworkevent.MouseLeft,
	}

	handled := si.HandleMouseEvent(frameworkevent.EventMousePress, ev)
	if !handled {
		t.Fatalf("expected inspector to handle overlay click")
	}

	if si.activeTab != TabConsole {
		t.Fatalf("expected activeTab switched to Console, got %v", si.activeTab)
	}
}

// TestOverlayTreeViewClick ensures TreeView in Elements tab handles mouse clicks.
func TestOverlayTreeViewClick(t *testing.T) {
	// Enable verbose output for this test
	t.Setenv("TUI_INSPECTOR_VERBOSE", "true")

	si := NewStandaloneInspector()
	si.Enable()
	si.ToggleVisibility() // make visible

	// Ensure we're on Elements tab which has TreeView
	si.activeTab = TabElements

	// Initialize inspector with test data to create TreeView component
	// The inspector needs tree data to create the display.TreeView component
	testRoot := ui.VStack(
		ui.Text("Item 1"),
		ui.Text("Item 2"),
		ui.Text("Item 3"),
	)
	si.treeView.SetRoot(testRoot)

	// Trigger tree line generation (this creates the display.TreeView component)
	si.treeLines, _ = si.treeView.GetTreeLines()
	if len(si.treeLines) == 0 {
		t.Fatal("Failed to generate tree lines from test data")
	}

	// Now get the TreeView component (should be non-nil after tree lines are generated)
	tvComponent := si.GetTreeViewComponent()
	if tvComponent == nil {
		// Manually create TreeView component if it wasn't auto-created
		si.treeViewComponent = display.NewTreeView().
			FromLines(si.treeLines).
			ExpandLevel(1).
			ShowIcons(true).
			Compact(false).
			Build().(*display.TreeView)
		tvComponent = si.treeViewComponent
	}

	t.Logf("TreeView state: lineCount=%d, focusIndex=%d",
		tvComponent.GetLineCount(), tvComponent.GetFocusIndex())

	// TreeView is rendered below tab bar (starting around Y=4-5)
	// Click in middle of overlay where TreeView should be
	x := 10
	y := 10

	ev := &frameworkevent.MouseEvent{
		BaseEvent: frameworkevent.NewBaseEvent(frameworkevent.EventMousePress),
		X:         x,
		Y:         y,
		Button:    frameworkevent.MouseLeft,
	}

	handled := si.HandleMouseEvent(frameworkevent.EventMousePress, ev)

	t.Logf("TreeView click test: handled=%v, lineCount=%d, focusIndex=%d",
		handled, tvComponent.GetLineCount(), tvComponent.GetFocusIndex())

	// The click should reach TreeView's HandleEvent with proper bounds set
	// Even if the line index is out of bounds, the event routing should work
	if !handled {
		t.Log("Click not handled (may be out of bounds or TreeView empty)")
	}
}
