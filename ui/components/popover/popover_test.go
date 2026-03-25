package popover

import (
	"strings"
	"testing"

	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/runtime/intent"
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	textcomp "github.com/wwsheng009/mint/ui/components/text"
)

type fakeInstallerHost struct {
	middlewareCount int
}

func (h *fakeInstallerHost) AddMiddleware(_ action.ActionMiddleware) {
	h.middlewareCount++
}

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

func TestComputePopoverBoxUsesSharedPlacementCoordinates(t *testing.T) {
	anchor := [4]int{10, 8, 8, 1}

	topLeft := computePopoverBox("", "1234567890", PlacementTopLeft, true, 1, 16, anchor)
	if topLeft.x != 10 || topLeft.y != 4 {
		t.Fatalf("top-left box = (%d,%d), want (10,4)", topLeft.x, topLeft.y)
	}

	top := computePopoverBox("", "1234567890", PlacementTop, true, 1, 16, anchor)
	if top.x != 7 || top.y != 4 {
		t.Fatalf("top box = (%d,%d), want (7,4)", top.x, top.y)
	}

	topRight := computePopoverBox("", "1234567890", PlacementTopRight, true, 1, 16, anchor)
	if topRight.x != 4 || topRight.y != 4 {
		t.Fatalf("top-right box = (%d,%d), want (4,4)", topRight.x, topRight.y)
	}

	bottomRight := computePopoverBox("", "1234567890", PlacementBottomRight, true, 1, 16, anchor)
	if bottomRight.x != 4 || bottomRight.y != 10 {
		t.Fatalf("bottom-right box = (%d,%d), want (4,10)", bottomRight.x, bottomRight.y)
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

func TestPopoverMiddlewareEscapeClosesTopmostOpenPopover(t *testing.T) {
	popoverRegistryGlobal.reset()
	defer popoverRegistryGlobal.reset()

	first := NewInstance(rtui.Props{propTitle: "First", propBody: "Body"})
	first.OnMount()
	defer first.Destroy()
	if !first.setOpen(true, TriggerClick) {
		t.Fatal("expected first popover to open")
	}

	second := NewInstance(rtui.Props{propTitle: "Second", propBody: "Body"})
	second.OnMount()
	defer second.Destroy()
	if !second.setOpen(true, TriggerClick) {
		t.Fatal("expected second popover to open")
	}

	middleware := NewMiddleware()
	if next := middleware.Before(action.NewAction(action.ActionCancel)); next != nil {
		t.Fatal("escape should be intercepted when a popover closes")
	}
	if !first.open {
		t.Fatal("older popover should remain open after ESC closes topmost")
	}
	if second.open {
		t.Fatal("topmost popover should close after ESC")
	}
}

func TestPopoverMiddlewareClickOutsideClosesOpenPopover(t *testing.T) {
	popoverRegistryGlobal.reset()
	defer popoverRegistryGlobal.reset()

	inst := NewInstance(rtui.Props{propTitle: "Mint", propBody: "Body"})
	inst.SetBounds(10, 5, 8, 1)
	inst.OnMount()
	defer inst.Destroy()
	inst.setOpen(true, TriggerClick)

	middleware := NewMiddleware()
	act := action.NewAction(action.ActionClick).WithPayload(runtimemsg.NewMouseMsg(1, 1, runtimemsg.MouseLeft, runtimemsg.MouseActionPress))
	if next := middleware.Before(act); next == nil {
		t.Fatal("outside click should continue dispatch after closing popover")
	}
	if inst.open {
		t.Fatal("popover should close after outside click")
	}
}

func TestPopoverMiddlewareLeavesAnchorAndOverlayClicksAlone(t *testing.T) {
	popoverRegistryGlobal.reset()
	defer popoverRegistryGlobal.reset()

	inst := NewInstance(rtui.Props{propTitle: "Mint", propBody: "Popover body"})
	inst.SetBounds(10, 5, 8, 1)
	inst.OnMount()
	defer inst.Destroy()
	inst.setOpen(true, TriggerClick)

	middleware := NewMiddleware()
	anchorClick := action.NewAction(action.ActionClick).WithPayload(runtimemsg.NewMouseMsg(11, 5, runtimemsg.MouseLeft, runtimemsg.MouseActionPress))
	if next := middleware.Before(anchorClick); next == nil {
		t.Fatal("anchor click should continue dispatch")
	}
	if !inst.open {
		t.Fatal("anchor click should not be treated as outside click")
	}

	box := computePopoverBox(inst.title, inst.body, inst.placement, inst.showArrow, inst.gapRows, inst.maxWidth, inst.bounds)
	overlayClick := action.NewAction(action.ActionClick).WithPayload(runtimemsg.NewMouseMsg(box.x+1, box.y+1, runtimemsg.MouseLeft, runtimemsg.MouseActionPress))
	if next := middleware.Before(overlayClick); next == nil {
		t.Fatal("overlay click should continue dispatch")
	}
	if !inst.open {
		t.Fatal("overlay click should keep popover open")
	}
}

func TestInstallAddsMiddlewareOnce(t *testing.T) {
	host := &fakeInstallerHost{}

	Install(host)
	Install(host)

	if host.middlewareCount != 1 {
		t.Fatalf("middlewareCount = %d, want 1", host.middlewareCount)
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
