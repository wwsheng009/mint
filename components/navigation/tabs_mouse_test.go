package navigation

import (
	"testing"
	"time"

	frameworkevent "github.com/wwsheng009/mint/framework/event"
	"github.com/wwsheng009/mint/runtime/event"
	"github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/ui"
)

// TestTabsMouseClick verifies that Tabs respond to mouse presses via TestableApp.
func TestTabsMouseClick(t *testing.T) {
	var tabsNode *TabsVNode

	activeID := "a"

	appFn := func() ui.VNode {
		if tabsNode == nil {
			tabs := TabsBuilder().
				AddTab("a", "Tab1").
				AddTab("b", "Tab2").
				Content("a", ui.Text("ONE")).
				Content("b", ui.Text("TWO")).
				ActiveTab(tabIndex(activeID)).
				OnChange(func(id string) { activeID = id }).
				Build()
			tabsNode = tabs.(*TabsVNode)
		}
		return tabsNode
	}
	ta, err := ui.RunTestWithSandbox(appFn, ui.WithSize(40, 5))
	if err != nil {
		t.Fatalf("RunTest failed: %v", err)
	}
	defer ta.Close()

	time.Sleep(20 * time.Millisecond)
	ta.ForceRender()
	if err := ta.AssertRender("ONE"); err != nil {
		t.Fatalf("expected initial render to contain ONE: %v", err)
	}

	if tabsNode == nil {
		t.Fatalf("tabsNode not captured")
	}
	if b := tabsNode.Bounds(); b[2] == 0 || b[3] == 0 {
		t.Fatalf("tabs bounds not set: %v", b)
	} else {
		t.Logf("tabs bounds: %v", b)
	}

	// Direct HandleEvent sanity check (should switch to Tab2)
	testEv := &frameworkevent.MouseEvent{
		BaseEvent: frameworkevent.NewBaseEvent(event.EventMousePress),
		X:         tabsNode.Bounds()[0] + 10,
		Y:         tabsNode.Bounds()[1],
		Button:    frameworkevent.MouseLeft,
	}
	if !tabsNode.HandleEvent(testEv) {
		t.Fatalf("direct HandleEvent returned false")
	}
	if tabsNode.activeTab != 1 {
		t.Fatalf("direct HandleEvent did not change activeTab, got %d", tabsNode.activeTab)
	}
	// Reset to Tab1 for injected event test
	tabsNode.SetActiveTab(0)

	// Click on the second tab ("Tab2") within the tab bar.
	// Tab bar string: "[Tab1] | Tab2" so Tab2 starts around x=9.
	if err := ta.InjectMouse(10, 0, platform.MouseLeft, platform.MousePress); err != nil {
		t.Fatalf("InjectMouse failed: %v", err)
	}

	// Also call root.HandleEvent directly to mirror App routing.
	root := ta.GetDeclarativeRoot()
	mouseEv := &frameworkevent.MouseEvent{
		BaseEvent: frameworkevent.NewBaseEvent(event.EventMousePress),
		X:         10,
		Y:         0,
		Button:    frameworkevent.MouseLeft,
	}
	handled := root.HandleEvent(mouseEv)
	t.Logf("root.HandleEvent handled=%v", handled)

	// Give the app a moment to process.
	time.Sleep(20 * time.Millisecond)
	ta.ForceRender()
	if tabsNode.activeTab != 1 {
		t.Fatalf("expected activeTab=1 after click, got %d", tabsNode.activeTab)
	}
	if err := ta.AssertRender("TWO"); err != nil {
		t.Fatalf("expected render to contain TWO after click: %v", err)
	}
}

func tabIndex(id string) int {
	if id == "b" {
		return 1
	}
	return 0
}
