package popover

import (
	"strings"
	"testing"

	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	textcomp "github.com/wwsheng009/mint/ui/components/text"
)

func TestNew(t *testing.T) {
	v := New(textcomp.New("anchor"))
	if v == nil {
		t.Fatal("New returned nil")
	}
	if v.Tag() != "popover" {
		t.Fatalf("Tag = %q, want popover", v.Tag())
	}
	if v.placement != PlacementAuto {
		t.Fatalf("placement = %v, want auto", v.placement)
	}
	if v.trigger != TriggerClick {
		t.Fatalf("trigger = %v, want click", v.trigger)
	}
	if !v.showArrow {
		t.Fatal("showArrow should default true")
	}
}

func TestBuilderFluent(t *testing.T) {
	child := textcomp.New("anchor")
	v := NewBuilder(child).
		Key("help").
		ComponentID("help.popover").
		Title("Mint").
		Body("Popover body").
		Placement(PlacementBottomRight).
		Trigger(TriggerHover).
		InitialOpen(true).
		ShowArrow(false).
		GapRows(2).
		MaxWidth(40).
		OpenForField(intent.BindField("helpOpen")).
		BuildVNode()

	if v.Key() != "help" || v.componentID != "help.popover" {
		t.Fatalf("key/componentID = (%q,%q)", v.Key(), v.componentID)
	}
	if v.title != "Mint" || v.body != "Popover body" {
		t.Fatalf("title/body = (%q,%q)", v.title, v.body)
	}
	if v.placement != PlacementBottomRight || v.trigger != TriggerHover {
		t.Fatalf("placement/trigger = (%v,%v)", v.placement, v.trigger)
	}
	if !v.initialOpen || v.showArrow {
		t.Fatalf("initialOpen/showArrow = (%v,%v)", v.initialOpen, v.showArrow)
	}
	if v.gapRows != 2 || v.maxWidth != 40 {
		t.Fatalf("gap/maxWidth = (%d,%d)", v.gapRows, v.maxWidth)
	}
}

func TestHandleActionClickTogglesOpen(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propTitle: "Mint",
		propBody:  "Body",
	})
	var emitted []intent.Intent
	inst.SetIntentEmitter(func(i intent.Intent) {
		emitted = append(emitted, i)
	})

	if !inst.HandleAction(action.NewAction(action.ActionClick)) {
		t.Fatal("expected click to toggle popover")
	}
	if !inst.open {
		t.Fatal("popover should open after click")
	}
	if len(emitted) != 1 {
		t.Fatalf("emitted len = %d, want 1", len(emitted))
	}
	change, ok := emitted[0].(PopoverChangeIntent)
	if !ok {
		t.Fatalf("emitted[0] = %T, want PopoverChangeIntent", emitted[0])
	}
	if !change.Open {
		t.Fatal("change intent should report open=true")
	}
}

func TestHandleActionHoverOpensAndCloses(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propTitle:   "Mint",
		propBody:    "Body",
		propTrigger: TriggerHover,
	})

	if !inst.HandleAction(action.NewAction(action.ActionMouseEnter)) || !inst.open {
		t.Fatal("hover enter should open popover")
	}
	if !inst.HandleAction(action.NewAction(action.ActionMouseLeave)) || inst.open {
		t.Fatal("hover leave should close popover")
	}
}

func TestHandleIntentRespectsComponentID(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propComponentID: "help.popover",
		propTitle:       "Mint",
		propBody:        "Body",
	})
	if !inst.HandleIntent(ToggleWithID("help.popover")) {
		t.Fatal("expected matching toggle intent to be handled")
	}
	if !inst.open {
		t.Fatal("popover should open")
	}
	if inst.HandleIntent(CloseWithID("other.popover")) {
		t.Fatal("expected other componentID to be ignored")
	}
}

func TestRuntimeChildrenIncludesOverlayWhenOpen(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propTitle: "Mint",
		propBody:  "Popover body",
	})
	inst.bounds = [4]int{10, 5, 8, 1}
	inst.open = true
	children := inst.RuntimeChildren()
	if len(children) != 1 {
		t.Fatalf("RuntimeChildren len = %d, want 1", len(children))
	}
	if children[0].GetLayer() != rtui.LayerOverlay {
		t.Fatalf("overlay layer = %v, want %v", children[0].GetLayer(), rtui.LayerOverlay)
	}
}

func TestOverlayPaintRendersTitleBodyAndArrow(t *testing.T) {
	v := newOverlayVNode(overlayProps{
		title:        "Mint",
		body:         "Popover body",
		placement:    PlacementBottom,
		showArrow:    true,
		gapRows:      1,
		maxWidth:     20,
		fillStyle:    style.Style{},
		borderStyle:  style.Style{},
		shadowStyle:  style.Style{},
		titleStyle:   style.Style{},
		bodyStyle:    style.Style{},
		anchorBounds: [4]int{10, 5, 8, 1},
	})
	inst := v.CreateInstance().(*overlayInstance)
	cmds := inst.Paint(0, 0)
	if len(cmds) == 0 {
		t.Fatal("expected overlay paint commands")
	}
	if !containsDrawText(cmds, "Mint") {
		t.Fatal("expected title draw command")
	}
	if !containsDrawText(cmds, "Popover body") {
		t.Fatal("expected body draw command")
	}
	if !containsDrawText(cmds, "▲") {
		t.Fatal("expected arrow on top border")
	}
}

func TestSetOpenEmitsFieldChange(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propTitle:             "Mint",
		propBody:              "Body",
		propChangeIntentField: intent.BindField("popoverOpen"),
	})
	var emitted []intent.Intent
	inst.SetIntentEmitter(func(i intent.Intent) {
		emitted = append(emitted, i)
	})
	inst.setOpen(true, TriggerClick)
	if len(emitted) != 2 {
		t.Fatalf("emitted len = %d, want 2", len(emitted))
	}
	fieldChange, ok := emitted[1].(intent.FieldChangeIntent)
	if !ok {
		t.Fatalf("emitted[1] = %T, want FieldChangeIntent", emitted[1])
	}
	if fieldChange.Field != "popoverOpen" || fieldChange.Value != "true" {
		t.Fatalf("unexpected field change: %+v", fieldChange)
	}
}

func containsDrawText(cmds []paint.DrawCmd, want string) bool {
	for _, cmd := range cmds {
		if strings.Contains(cmd.Text, want) {
			return true
		}
	}
	return false
}
