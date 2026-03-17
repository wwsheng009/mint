package badge

import (
	"testing"

	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

func TestNew(t *testing.T) {
	v := New("Inbox")
	if v == nil {
		t.Fatal("New returned nil")
	}
	if v.Tag() != "badge" {
		t.Fatalf("Tag = %q, want badge", v.Tag())
	}
	if v.label != "Inbox" {
		t.Fatalf("label = %q, want Inbox", v.label)
	}
}

func TestBuilderFluent(t *testing.T) {
	v := NewBuilder("Inbox").
		Key("inbox").
		Count(120).
		OverflowCount(99).
		Primary().
		ShowZero(true).
		Build()

	if v.Key() != "inbox" {
		t.Fatalf("key = %q, want inbox", v.Key())
	}
	if v.count != 120 || v.overflowCount != 99 || !v.showZero || v.status != StatusPrimary {
		t.Fatalf("unexpected builder values: %+v", v)
	}
}

func TestInstanceMeasureAndPaintCount(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propLabel:         "Inbox",
		propCount:         120,
		propOverflowCount: 99,
		propStatus:        StatusError,
	})
	size := inst.Measure(layout.Constraints{MaxWidth: 80, MaxHeight: 1})
	wantWidth := paint.StringWidth("Inbox [99+]")
	if size.Width != wantWidth || size.Height != 1 {
		t.Fatalf("size = %+v, want width=%d height=1", size, wantWidth)
	}
	inst.SetBounds(0, 0, 80, 1)
	cmds := inst.Paint(0, 0)
	if len(cmds) != 3 {
		t.Fatalf("cmd count = %d, want 3", len(cmds))
	}
	if cmds[0].Text != "Inbox" || cmds[2].Text != "[99+]" {
		t.Fatalf("painted text = %q %q", cmds[0].Text, cmds[2].Text)
	}
	if cmds[2].Style.BG != theme.Error() {
		t.Fatalf("badge BG = %q, want %q", cmds[2].Style.BG, theme.Error())
	}
}

func TestInstancePaintDotStatus(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propLabel:  "Live",
		propDot:    true,
		propStatus: StatusSuccess,
	})
	inst.SetBounds(0, 0, 20, 1)
	cmds := inst.Paint(0, 0)
	if len(cmds) != 3 {
		t.Fatalf("cmd count = %d, want 3", len(cmds))
	}
	if cmds[2].Text != "●" {
		t.Fatalf("dot text = %q, want ●", cmds[2].Text)
	}
	if cmds[2].Style.FG != theme.Success() {
		t.Fatalf("dot FG = %q, want %q", cmds[2].Style.FG, theme.Success())
	}
}

func TestInstanceShowZeroAndHide(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propLabel: "Inbox",
		propCount: 0,
	})
	inst.SetBounds(0, 0, 20, 1)
	cmds := inst.Paint(0, 0)
	if len(cmds) != 1 || cmds[0].Text != "Inbox" {
		t.Fatalf("paint without showZero = %+v", cmds)
	}

	inst = NewInstance(rtui.Props{
		propLabel:    "Inbox",
		propCount:    0,
		propShowZero: true,
	})
	inst.SetBounds(0, 0, 20, 1)
	cmds = inst.Paint(0, 0)
	if len(cmds) != 3 || cmds[2].Text != "[0]" {
		t.Fatalf("paint with showZero = %+v", cmds)
	}
}

func TestInstanceCustomText(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propText:   "NEW",
		propStatus: StatusPrimary,
	})
	inst.SetBounds(0, 0, 20, 1)
	cmds := inst.Paint(0, 0)
	if len(cmds) != 1 || cmds[0].Text != "[NEW]" {
		t.Fatalf("paint = %+v, want [NEW]", cmds)
	}
}

func TestVNodeImplementsInterfaces(t *testing.T) {
	var _ rtui.VNode = New("Inbox")
	var _ rtui.InstanceFactory = New("Inbox")
}
