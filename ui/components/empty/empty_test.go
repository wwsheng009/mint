package empty

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/layout"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

func TestEmptyInstance_DefaultMeasureAndPaint(t *testing.T) {
	inst := NewInstance(rtui.Props{})

	size := inst.Measure(layout.Constraints{})
	if size.Width != len(defaultImage) {
		t.Fatalf("expected width %d, got %d", len(defaultImage), size.Width)
	}
	if size.Height != 2 {
		t.Fatalf("expected height 2, got %d", size.Height)
	}

	inst.SetBounds(0, 0, size.Width, size.Height)
	cmds := inst.Paint(0, 0)
	if len(cmds) != 2 {
		t.Fatalf("expected 2 draw commands, got %d", len(cmds))
	}
	if cmds[0].Text != defaultImage {
		t.Fatalf("expected default image %q, got %q", defaultImage, cmds[0].Text)
	}
	if cmds[1].Text != "No Data" {
		t.Fatalf("expected default description %q, got %q", "No Data", cmds[1].Text)
	}
}

func TestEmptyInstance_SetPropsUpdatesKeyAndContent(t *testing.T) {
	inst := NewInstance(rtui.Props{"key": "empty-1"})

	changed := inst.SetProps(rtui.Props{
		"key":         "empty-2",
		"description": "Nothing here",
		"image":       "[x]",
	})
	if !changed {
		t.Fatal("expected SetProps to report change")
	}

	if inst.Key() != "empty-2" {
		t.Fatalf("expected key to update to empty-2, got %q", inst.Key())
	}

	props := inst.GetProps()
	if got, _ := props["description"].(string); got != "Nothing here" {
		t.Fatalf("expected description to update, got %q", got)
	}
	if got, _ := props["image"].(string); got != "[x]" {
		t.Fatalf("expected image to update, got %q", got)
	}
}
