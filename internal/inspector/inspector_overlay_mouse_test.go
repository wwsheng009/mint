package inspector

import (
	"testing"

	frameworkevent "github.com/wwsheng009/mint/framework/event"
	"github.com/wwsheng009/mint/runtime/event"
)

// TestOverlayTabsMouse ensures overlay tab bar reacts to mouse press.
func TestOverlayTabsMouse(t *testing.T) {
	si := NewStandaloneInspector()
	si.Enable()
	si.ToggleVisibility() // make visible

	// Default position (0,0), overlayWidth/Height default from constructor (80,25).
	// Tab bar rendered at localY = 1. Click roughly on second tab label.
	x := 14 // position inside "Console" label (after [Elements])
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
