package selectcomp

import (
	"strings"
	"testing"

	"github.com/wwsheng009/mint/framework"
	"github.com/wwsheng009/mint/framework/component"
	"github.com/wwsheng009/mint/internal/render"
	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/layout"
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
	"github.com/wwsheng009/mint/runtime/paint"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// VNode Tests
// =============================================================================

func TestVNode_New(t *testing.T) {
	s := New()
	if s == nil {
		t.Fatal("New returned nil")
	}
	if s.Tag() != "select" {
		t.Errorf("Tag = %q, want %q", s.Tag(), "select")
	}
}

func TestVNode_Defaults(t *testing.T) {
	s := New()
	if len(s.Options()) != 0 {
		t.Error("Default options should be empty")
	}
	if s.SelectedIndex() != -1 {
		t.Errorf("Default selectedIndex = %d, want -1", s.SelectedIndex())
	}
	if s.Disabled() {
		t.Error("Default disabled should be false")
	}
}

func TestVNode_Builder(t *testing.T) {
	opts := []Option{
		{Value: "a", Label: "Option A"},
		{Value: "b", Label: "Option B"},
	}

	s := NewBuilder().
		Options(opts).
		Selected(1).
		Key("myselect").
		Build()

	vnode := s.(*VNode)
	if len(vnode.Options()) != 2 {
		t.Errorf("Options count = %d, want 2", len(vnode.Options()))
	}
	if vnode.SelectedIndex() != 1 {
		t.Errorf("SelectedIndex = %d, want 1", vnode.SelectedIndex())
	}
	if vnode.Key() != "myselect" {
		t.Errorf("Key = %q, want %q", vnode.Key(), "myselect")
	}
}

func TestVNode_AddOption(t *testing.T) {
	s := New().
		AddOption("a", "Option A").
		AddOption("b", "Option B")

	if len(s.Options()) != 2 {
		t.Errorf("Options count = %d, want 2", len(s.Options()))
	}
	if s.Options()[0].Value != "a" {
		t.Errorf("First option value = %q, want %q", s.Options()[0].Value, "a")
	}
}

func TestVNode_CreateInstance(t *testing.T) {
	s := New().
		AddOption("a", "Option A").
		SetSelectedIndex(0)

	inst := s.CreateInstance()
	if inst == nil {
		t.Fatal("CreateInstance returned nil")
	}

	ci, ok := inst.(*Instance)
	if !ok {
		t.Fatal("Instance is not *Instance")
	}
	if ci.SelectedIndex() != 0 {
		t.Errorf("SelectedIndex = %d, want 0", ci.SelectedIndex())
	}
}

// =============================================================================
// Instance Tests
// =============================================================================

func TestInstance_New(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"options": []Option{
			{Value: "a", Label: "Option A"},
			{Value: "b", Label: "Option B"},
		},
		"selectedIndex": 1,
	})

	if len(inst.options) != 2 {
		t.Errorf("Options count = %d, want 2", len(inst.options))
	}
	if inst.SelectedIndex() != 1 {
		t.Errorf("SelectedIndex = %d, want 1", inst.SelectedIndex())
	}
}

func TestInstance_Measure(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"options": []Option{
			{Value: "a", Label: "Short"},
			{Value: "b", Label: "Longer Option"},
		},
	})

	size := inst.Measure(layout.UnboundedConstraints())

	// "Longer Option" = 13 chars + 4 for "< >" = 17
	if size.Width < 17 {
		t.Errorf("Width = %d, want >= 17", size.Width)
	}
	if size.Height != 1 {
		t.Errorf("Height = %d, want 1", size.Height)
	}
}

func TestInstance_SelectNext(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"options": []Option{
			{Value: "a", Label: "A"},
			{Value: "b", Label: "B"},
			{Value: "c", Label: "C"},
		},
		"selectedIndex": 0,
	})

	inst.SelectNext()
	if inst.SelectedIndex() != 1 {
		t.Errorf("SelectedIndex = %d, want 1", inst.SelectedIndex())
	}

	inst.SelectNext()
	if inst.SelectedIndex() != 2 {
		t.Errorf("SelectedIndex = %d, want 2", inst.SelectedIndex())
	}

	// Wrap around
	inst.SelectNext()
	if inst.SelectedIndex() != 0 {
		t.Errorf("SelectedIndex = %d, want 0 (wrap)", inst.SelectedIndex())
	}
}

func TestInstance_SelectPrev(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"options": []Option{
			{Value: "a", Label: "A"},
			{Value: "b", Label: "B"},
			{Value: "c", Label: "C"},
		},
		"selectedIndex": 0,
	})

	inst.SelectPrev()
	if inst.SelectedIndex() != 2 {
		t.Errorf("SelectedIndex = %d, want 2 (wrap)", inst.SelectedIndex())
	}

	inst.SelectPrev()
	if inst.SelectedIndex() != 1 {
		t.Errorf("SelectedIndex = %d, want 1", inst.SelectedIndex())
	}
}

func TestInstance_SelectedValue(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"options": []Option{
			{Value: "val_a", Label: "Label A"},
			{Value: "val_b", Label: "Label B"},
		},
		"selectedIndex": 1,
	})

	if inst.SelectedValue() != "val_b" {
		t.Errorf("SelectedValue = %q, want %q", inst.SelectedValue(), "val_b")
	}
	if inst.SelectedLabel() != "Label B" {
		t.Errorf("SelectedLabel = %q, want %q", inst.SelectedLabel(), "Label B")
	}
}

func TestInstance_Disabled(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"options": []Option{
			{Value: "a", Label: "A"},
		},
		"disabled": true,
	})

	if !inst.IsDisabled() {
		t.Error("Should be disabled")
	}

	// Disabled instance should not handle actions
	if inst.HandleAction(action.NewAction(action.ActionSelect)) {
		t.Error("Disabled instance should not handle select action")
	}
}

func TestInstance_HandleAction(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"options": []Option{
			{Value: "a", Label: "A"},
			{Value: "b", Label: "B"},
		},
		"selectedIndex": 0,
	})

	// Select/Click/Enter/Down action
	handled := inst.HandleAction(action.NewAction(action.ActionSelect))
	if !handled {
		t.Error("Select action should be handled")
	}
	if inst.SelectedIndex() != 1 {
		t.Errorf("SelectedIndex = %d, want 1", inst.SelectedIndex())
	}

	// Up action
	handled = inst.HandleAction(action.NewAction(action.ActionNavigateUp))
	if !handled {
		t.Error("Up action should be handled")
	}
	if inst.SelectedIndex() != 0 {
		t.Errorf("SelectedIndex = %d, want 0", inst.SelectedIndex())
	}

	// Unknown action
	handled = inst.HandleAction(action.NewActionWithPayload("unknown", nil))
	if handled {
		t.Error("Unknown action should not be handled")
	}
}

func TestInstance_Focus(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"options": []Option{
			{Value: "a", Label: "A"},
		},
	})

	if inst.HasFocus() {
		t.Error("Should not have focus initially")
	}

	inst.SetFocus(true)
	if !inst.HasFocus() {
		t.Error("Should have focus after SetFocus(true)")
	}

	inst.SetFocus(false)
	if inst.HasFocus() {
		t.Error("Should not have focus after SetFocus(false)")
	}
}

func TestInstance_Paint(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"options": []Option{
			{Value: "a", Label: "Option A"},
			{Value: "b", Label: "Option B"},
		},
		"selectedIndex": 1,
	})

	cmds := inst.Paint(0, 0)
	if len(cmds) != 1 {
		t.Fatalf("Paint returned %d commands, want 1", len(cmds))
	}

	expected := "< Option B >"
	if cmds[0].Text != expected {
		t.Errorf("Text = %q, want %q", cmds[0].Text, expected)
	}
}

func TestInstance_Paint_NoSelection(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"options": []Option{
			{Value: "a", Label: "Option A"},
		},
		"selectedIndex": -1,
	})

	cmds := inst.Paint(0, 0)
	if len(cmds) != 1 {
		t.Fatalf("Paint returned %d commands, want 1", len(cmds))
	}

	// No selection shows first option
	expected := "< Option A >"
	if cmds[0].Text != expected {
		t.Errorf("Text = %q, want %q", cmds[0].Text, expected)
	}
}

func TestInstance_HandleAction_ClickOpensDropdown(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"options": []Option{
			{Value: "a", Label: "A"},
			{Value: "b", Label: "B"},
		},
	})

	if handled := inst.HandleAction(action.NewAction(action.ActionClick)); !handled {
		t.Fatal("click should open dropdown")
	}
	if !inst.open {
		t.Fatal("dropdown should be open after click")
	}
}

func TestInstance_HandleAction_MouseReleaseDoesNotCloseOverlayTrigger(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"overlayPopup": true,
		"options": []Option{
			{Value: "a", Label: "A"},
			{Value: "b", Label: "B"},
		},
	})

	press := runtimemsg.NewMouseMsg(1, 1, runtimemsg.MouseLeft, runtimemsg.MouseActionPress)
	if handled := inst.HandleAction(action.NewAction(action.ActionClick).WithPayload(press)); !handled {
		t.Fatal("mouse press should open dropdown")
	}
	if !inst.open {
		t.Fatal("overlay select should be open after mouse press")
	}

	release := runtimemsg.NewMouseMsg(1, 1, runtimemsg.MouseLeft, runtimemsg.MouseActionRelease)
	if handled := inst.HandleAction(action.NewAction(action.ActionSelect).WithPayload(release)); !handled {
		t.Fatal("mouse release should be consumed")
	}
	if !inst.open {
		t.Fatal("mouse release should not close overlay select")
	}
}

func TestInstance_HandleAction_ActionMouseReleaseDoesNotCloseOverlayTrigger(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"overlayPopup": true,
		"options": []Option{
			{Value: "a", Label: "A"},
			{Value: "b", Label: "B"},
		},
	})
	inst.open = true

	release := runtimemsg.NewMouseMsg(1, 1, runtimemsg.MouseLeft, runtimemsg.MouseActionRelease)
	if handled := inst.HandleAction(action.NewAction(action.ActionMouseRelease).WithPayload(release)); !handled {
		t.Fatal("ActionMouseRelease should be consumed by trigger")
	}
	if !inst.open {
		t.Fatal("ActionMouseRelease should not close overlay select")
	}
}

func TestInstance_Paint_OpenDropdownIncludesPopup(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"options": []Option{
			{Value: "a", Label: "Option A"},
			{Value: "b", Label: "Option B"},
			{Value: "c", Label: "Option C"},
		},
		"selectedIndex": 1,
	})

	inst.openDropdown()
	cmds := inst.Paint(0, 0)
	if len(cmds) < 5 {
		t.Fatalf("Paint returned %d commands, want popup commands too", len(cmds))
	}

	var texts []string
	for _, cmd := range cmds {
		texts = append(texts, cmd.Text)
	}
	joined := strings.Join(texts, "\n")
	if !strings.Contains(joined, "┌") || !strings.Contains(joined, "Option A") || !strings.Contains(joined, "Option B") {
		t.Fatalf("open popup paint missing expected content:\n%s", joined)
	}
}

func TestInstance_HandleAction_MultiSelectTogglesIndices(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"options": []Option{
			{Value: "a", Label: "A"},
			{Value: "b", Label: "B"},
			{Value: "c", Label: "C"},
		},
		"selectionMode": SelectionMultiple,
	})

	if !inst.HandleAction(action.NewAction(action.ActionEnter)) {
		t.Fatal("enter should open multi-select dropdown")
	}
	if !inst.open {
		t.Fatal("dropdown should be open")
	}

	if !inst.HandleAction(action.NewAction(action.ActionEnter)) {
		t.Fatal("enter should toggle highlighted option")
	}
	if got := inst.SelectedIndices(); len(got) != 1 || got[0] != 0 {
		t.Fatalf("SelectedIndices = %v, want [0]", got)
	}
	if !inst.open {
		t.Fatal("multi-select should stay open after toggle")
	}

	if !inst.HandleAction(action.NewAction(action.ActionNavigateDown)) {
		t.Fatal("down should move highlight")
	}
	if !inst.HandleAction(action.NewAction(action.ActionEnter)) {
		t.Fatal("enter should toggle second option")
	}
	if got := inst.SelectedIndices(); len(got) != 2 || got[0] != 0 || got[1] != 1 {
		t.Fatalf("SelectedIndices = %v, want [0 1]", got)
	}
}

func TestInstance_HandleAction_MultiSelectEmitsFieldChangeIntent(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"options": []Option{
			{Value: "a", Label: "A"},
			{Value: "b", Label: "B"},
		},
		"selectionMode": SelectionMultiple,
		"changeIntent":  intent.BindField("choices"),
	})

	var emitted []intent.Intent
	inst.SetIntentEmitter(func(i intent.Intent) {
		emitted = append(emitted, i)
	})

	inst.openDropdown()
	inst.HandleAction(action.NewAction(action.ActionEnter))
	inst.HandleAction(action.NewAction(action.ActionNavigateDown))
	inst.HandleAction(action.NewAction(action.ActionEnter))

	var lastField intent.FieldChangeIntent
	found := false
	for _, emittedIntent := range emitted {
		fieldChange, ok := emittedIntent.(intent.FieldChangeIntent)
		if !ok {
			continue
		}
		lastField = fieldChange
		found = true
	}
	if !found {
		t.Fatal("expected FieldChangeIntent to be emitted")
	}
	if lastField.Field != "choices" {
		t.Fatalf("Field = %q, want %q", lastField.Field, "choices")
	}
	if lastField.Value != "0,1" {
		t.Fatalf("Value = %q, want %q", lastField.Value, "0,1")
	}
}

func TestInstance_RuntimeChildren_OverlayPopupChild(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"selectID":       "country-select",
		"portalRoot":     rtui.DefaultOverlayPortalRootID,
		"overlayPopup":   true,
		"maxVisibleRows": 6,
		"options": []Option{
			{Value: "us", Label: "United States"},
			{Value: "cn", Label: "China"},
		},
	})

	if children := inst.RuntimeChildren(); len(children) != 0 {
		t.Fatalf("closed RuntimeChildren len = %d, want 0", len(children))
	}

	inst.openDropdown()
	if children := inst.RuntimeChildren(); len(children) != 0 {
		t.Fatalf("overlay RuntimeChildren len = %d, want 0", len(children))
	}
}

func TestPopupInstance_Measure_UsesPopupProps(t *testing.T) {
	popup := newPopupInstance(rtui.Props{
		"selectID":       "country-select",
		"componentID":    "country-select",
		"options":        []Option{{Value: "us", Label: "United States"}, {Value: "cn", Label: "China"}},
		"maxVisibleRows": 6,
		"minWidth":       20,
	})
	size := popup.Measure(layout.UnboundedConstraints())
	if size.Width <= 0 || size.Height <= 0 {
		t.Fatalf("popup Measure = %+v, want non-zero size", size)
	}
}

func TestPopupInstance_HandleAction_CommitsSelection(t *testing.T) {
	state := overlayControllerState{
		selectedIndex:    -1,
		selectedIndices:  nil,
		open:             true,
		highlightedIndex: 0,
		scrollOffset:     0,
	}
	callbacks := &overlayCallbacks{
		setOpen: func(open bool) overlayControllerState {
			state.open = open
			return state
		},
		setHighlight: func(index int) overlayControllerState {
			state.highlightedIndex = index
			return state
		},
		commit: func(index int) overlayControllerState {
			nextIndex, nextIndices, _, shouldClose := applyOverlayCommit(
				SelectionSingle,
				2,
				state.selectedIndex,
				state.selectedIndices,
				index,
			)
			state.selectedIndex = nextIndex
			state.selectedIndices = nextIndices
			state.highlightedIndex = index
			state.open = !shouldClose
			return state
		},
	}
	emitted := make([]intent.Intent, 0, 2)
	popup := newPopupInstance(rtui.Props{
		"selectID":         "country-select",
		"componentID":      "country-select",
		"options":          []Option{{Value: "us", Label: "United States"}, {Value: "cn", Label: "China"}},
		"selectedIndex":    -1,
		"selectedIndices":  []int{},
		"highlightedIndex": 0,
		"maxVisibleRows":   6,
		"minWidth":         20,
		"changeIntent":     intent.BindField("country"),
		overlayCallbacksProp: callbacks,
	})
	popup.SetIntentEmitter(func(i intent.Intent) {
		emitted = append(emitted, i)
	})

	mouse := runtimemsg.NewMouseMsg(1, 2, runtimemsg.MouseLeft, runtimemsg.MouseActionPress)
	mouse.LocalX = 1
	mouse.LocalY = 1
	if !popup.HandleAction(action.NewAction(action.ActionClick).WithPayload(mouse)) {
		t.Fatal("popup click should be handled")
	}
	if state.selectedIndex != 0 {
		t.Fatalf("selectedIndex = %d, want 0", state.selectedIndex)
	}
	if state.open {
		t.Fatal("single-select popup should close after commit")
	}
	if len(emitted) == 0 {
		t.Fatal("expected popup commit to emit intents")
	}
}

func TestPopupInstance_HandleAction_MouseReleaseCommitsSelection(t *testing.T) {
	state := overlayControllerState{
		selectedIndex:    -1,
		selectedIndices:  nil,
		open:             true,
		highlightedIndex: 0,
		scrollOffset:     0,
	}
	callbacks := &overlayCallbacks{
		setOpen: func(open bool) overlayControllerState {
			state.open = open
			return state
		},
		setHighlight: func(index int) overlayControllerState {
			state.highlightedIndex = index
			return state
		},
		commit: func(index int) overlayControllerState {
			nextIndex, nextIndices, _, shouldClose := applyOverlayCommit(
				SelectionSingle,
				2,
				state.selectedIndex,
				state.selectedIndices,
				index,
			)
			state.selectedIndex = nextIndex
			state.selectedIndices = nextIndices
			state.highlightedIndex = index
			state.open = !shouldClose
			return state
		},
	}
	popup := newPopupInstance(rtui.Props{
		"selectID":           "country-select",
		"componentID":        "country-select",
		"options":            []Option{{Value: "us", Label: "United States"}, {Value: "cn", Label: "China"}},
		"selectedIndex":      -1,
		"selectedIndices":    []int{},
		"highlightedIndex":   0,
		"maxVisibleRows":     6,
		"minWidth":           20,
		overlayCallbacksProp: callbacks,
	})

	mouse := runtimemsg.NewMouseMsg(1, 2, runtimemsg.MouseLeft, runtimemsg.MouseActionRelease)
	mouse.LocalX = 1
	mouse.LocalY = 1
	if !popup.HandleAction(action.NewAction(action.ActionMouseRelease).WithPayload(mouse)) {
		t.Fatal("popup mouse release should be handled")
	}
	if state.selectedIndex != 0 {
		t.Fatalf("selectedIndex = %d, want 0", state.selectedIndex)
	}
	if state.open {
		t.Fatal("single-select popup should close after mouse release commit")
	}
}

func TestPopupInstance_HandleAction_NavigateDownUpdatesHighlight(t *testing.T) {
	state := overlayControllerState{
		selectedIndex:    -1,
		selectedIndices:  nil,
		open:             true,
		highlightedIndex: 0,
		scrollOffset:     0,
	}
	callbacks := &overlayCallbacks{
		setOpen: func(open bool) overlayControllerState {
			state.open = open
			return state
		},
		setHighlight: func(index int) overlayControllerState {
			state.highlightedIndex = index
			return state
		},
		commit: func(index int) overlayControllerState {
			state.highlightedIndex = index
			return state
		},
	}
	popup := newPopupInstance(rtui.Props{
		"selectID":         "country-select",
		"componentID":      "country-select",
		"options":          []Option{{Value: "us", Label: "United States"}, {Value: "cn", Label: "China"}, {Value: "jp", Label: "Japan"}},
		"selectedIndex":    -1,
		"selectedIndices":  []int{},
		"highlightedIndex": 0,
		"maxVisibleRows":   6,
		"minWidth":         20,
		overlayCallbacksProp: callbacks,
	})

	if !popup.HandleAction(action.NewAction(action.ActionNavigateDown)) {
		t.Fatal("popup down should be handled")
	}
	if state.highlightedIndex != 1 {
		t.Fatalf("highlightedIndex = %d, want 1", state.highlightedIndex)
	}
}

func TestPopupInstance_HandleAction_NavigateDownUpdatesHighlight_EmptyComponentID(t *testing.T) {
	state := overlayControllerState{
		selectedIndex:    -1,
		selectedIndices:  nil,
		open:             true,
		highlightedIndex: 0,
		scrollOffset:     0,
	}
	callbacks := &overlayCallbacks{
		setOpen: func(open bool) overlayControllerState {
			state.open = open
			return state
		},
		setHighlight: func(index int) overlayControllerState {
			state.highlightedIndex = index
			return state
		},
		commit: func(index int) overlayControllerState {
			state.highlightedIndex = index
			return state
		},
	}
	popup := newPopupInstance(rtui.Props{
		"selectID":         "country-select",
		"options":          []Option{{Value: "us", Label: "United States"}, {Value: "cn", Label: "China"}, {Value: "jp", Label: "Japan"}},
		"selectedIndex":    -1,
		"selectedIndices":  []int{},
		"highlightedIndex": 0,
		"maxVisibleRows":   6,
		"minWidth":         20,
		overlayCallbacksProp: callbacks,
	})

	if !popup.HandleAction(action.NewAction(action.ActionNavigateDown)) {
		t.Fatal("popup down should be handled")
	}
	if state.highlightedIndex != 1 {
		t.Fatalf("highlightedIndex = %d, want 1", state.highlightedIndex)
	}
}

func TestMiddleware_ClickOutsideClosesOpenOverlaySelect(t *testing.T) {
	open := true
	callbacks := &overlayCallbacks{
		setOpen: func(next bool) overlayControllerState {
			open = next
			return overlayControllerState{open: open}
		},
		setHighlight: func(index int) overlayControllerState {
			return overlayControllerState{open: open, highlightedIndex: index}
		},
		commit: func(index int) overlayControllerState {
			return overlayControllerState{open: open, highlightedIndex: index}
		},
	}
	popup := newPopupInstance(rtui.Props{
		"selectID":         "country-select",
		"componentID":      "country-select",
		"closeOnOutside":   true,
		"options":          []Option{{Value: "us", Label: "United States"}, {Value: "cn", Label: "China"}},
		overlayCallbacksProp: callbacks,
	})
	popup.SetBounds(2, 3, 12, 4)
	popup.OnMount()
	defer popup.OnUnmount()

	middleware := NewMiddleware()
	act := action.NewAction(action.ActionClick).WithPayload(
		runtimemsg.NewMouseMsg(40, 20, runtimemsg.MouseLeft, runtimemsg.MouseActionPress),
	)

	if result := middleware.Before(act); result != nil {
		t.Fatal("outside click should be intercepted after overlay select closes")
	}
	if open {
		t.Fatal("overlay select should be closed after outside click")
	}
}

func TestMiddleware_CancelClosesOpenOverlaySelect(t *testing.T) {
	open := true
	callbacks := &overlayCallbacks{
		setOpen: func(next bool) overlayControllerState {
			open = next
			return overlayControllerState{open: open}
		},
		setHighlight: func(index int) overlayControllerState {
			return overlayControllerState{open: open, highlightedIndex: index}
		},
		commit: func(index int) overlayControllerState {
			return overlayControllerState{open: open, highlightedIndex: index}
		},
	}
	popup := newPopupInstance(rtui.Props{
		"selectID":         "country-select",
		"componentID":      "country-select",
		"options":          []Option{{Value: "us", Label: "United States"}, {Value: "cn", Label: "China"}},
		overlayCallbacksProp: callbacks,
	})
	popup.OnMount()
	defer popup.OnUnmount()

	middleware := NewMiddleware()
	act := action.NewAction(action.ActionCancel)

	if result := middleware.Before(act); result != nil {
		t.Fatal("cancel should be intercepted after overlay select closes")
	}
	if open {
		t.Fatal("overlay select should be closed after cancel")
	}
}

func TestOverlayPopup_AnchorsToSelectBounds(t *testing.T) {
	selectNode := NewBuilder().
		SetID("country-select").
		OverlayPopup(true).
		Width(20).
		Options([]Option{
			{Value: "us", Label: "United States"},
			{Value: "cn", Label: "China"},
			{Value: "jp", Label: "Japan"},
		}).
		Build()

	app := func() rtui.VNode {
		return rtui.VStack(
			rtui.NewElement("box").SetProps(rtui.Props{
				"portalRootId": rtui.DefaultOverlayPortalRootID,
				"position":     "absolute",
				"left":         0,
				"top":          0,
				"width":        1,
				"height":       1,
			}),
			rtui.NewElement("text").SetProps(rtui.Props{"content": "Country:"}),
			rtui.HStack(
				rtui.NewElement("text").SetProps(rtui.Props{"content": "  "}),
				selectNode,
			),
		)
	}

	node := render.NewDeclarativeNodeFromFuncWithFiber(app)
	node.SetApp(framework.NewApp())
	node.SetRenderMode(render.RenderModeFiberFirst)

	ctx := component.PaintContext{
		Bounds:          paint.Rect{X: 0, Y: 0, Width: 50, Height: 12},
		AvailableWidth:  50,
		AvailableHeight: 12,
	}

	node.Paint(ctx, paint.NewBuffer(50, 12))
	fiberRoot := node.GetFiberRoot()
	if fiberRoot == nil || fiberRoot.Child == nil {
		t.Fatalf("unexpected fiber tree shape after render")
	}
	var selectFiber *rtui.Fiber
	var overlayFiber *rtui.Fiber
	var findSelect func(*rtui.Fiber)
	findSelect = func(f *rtui.Fiber) {
		if f == nil {
			return
		}
		if f.Tag == "SelectOverlay" && overlayFiber == nil {
			overlayFiber = f
		}
		if f.Tag == "select" {
			selectFiber = f
		}
		findSelect(f.Child)
		findSelect(f.Sibling)
	}
	findSelect(fiberRoot)
	if selectFiber == nil || overlayFiber == nil {
		t.Fatalf("expected select overlay fibers in tree")
	}
	if selectFiber.Tag != "select" {
		t.Fatalf("select fiber tag = %q", selectFiber.Tag)
	}
	triggerInst, ok := selectFiber.Instance.(*Instance)
	if !ok || triggerInst == nil {
		t.Fatal("expected select fiber instance")
	}
	if !triggerInst.HandleAction(action.NewAction(action.ActionEnter)) {
		t.Fatal("select enter should open popup")
	}
	node.Paint(ctx, paint.NewBuffer(50, 12))
	fiberRoot = node.GetFiberRoot()
	selectFiber = nil
	overlayFiber = nil
	findSelect(fiberRoot)
	if selectFiber == nil {
		t.Fatalf("expected select fiber after opening popup")
	}
	if selectFiber.ID != "country-select" {
		t.Fatalf("select fiber ID = %q, want %q", selectFiber.ID, "country-select")
	}

	portals := node.GetPortalRoots()
	if len(portals) != 1 {
		t.Fatalf("portal roots len = %d, want 1\nlayout:\n%s\nportal:\n%s", len(portals), node.GetLayoutTreeString(), node.GetPortalTreeString())
	}

	triggerInst = selectFiber.Instance.(*Instance)
	triggerX, triggerY, _, triggerH := triggerInst.GetBounds()
	if triggerX == 0 && triggerY == 0 {
		t.Fatalf("trigger bounds stayed at origin; test would not validate anchoring\nlayout:\n%s\npaintable:\n%s", node.GetLayoutTreeString(), node.GetPaintableTreeString())
	}
	portal := portals[0]
	if portal.X != triggerX {
		t.Fatalf("portal X = %d, want %d\nlayout:\n%s\nportal:\n%s\npaintable:\n%s", portal.X, triggerX, node.GetLayoutTreeString(), node.GetPortalTreeString(), node.GetPaintableTreeString())
	}
	expectedPopupY := triggerY + triggerH
	if portal.Y != expectedPopupY {
		t.Fatalf("portal Y = %d, want %d\nlayout:\n%s\nportal:\n%s\npaintable:\n%s", portal.Y, expectedPopupY, node.GetLayoutTreeString(), node.GetPortalTreeString(), node.GetPaintableTreeString())
	}

	var popupBox *paint.PaintableBox
	var findPopup func(*paint.PaintableBox)
	findPopup = func(box *paint.PaintableBox) {
		if box == nil || popupBox != nil {
			return
		}
		if box.Node != nil && box.Node.Tag() == "select-popup" {
			popupBox = box
			return
		}
		for _, child := range box.Children {
			findPopup(child)
		}
	}
	findPopup(node.GetPaintableRoot())
	if popupBox == nil {
		t.Fatalf("select-popup paintable box not found\npaintable:\n%s", node.GetPaintableTreeString())
	}
	if popupBox.X != triggerX {
		t.Fatalf("popup paintable X = %d, want %d\nlayout:\n%s\nportal:\n%s\npaintable:\n%s", popupBox.X, triggerX, node.GetLayoutTreeString(), node.GetPortalTreeString(), node.GetPaintableTreeString())
	}
	if popupBox.Y != expectedPopupY {
		t.Fatalf("popup paintable Y = %d, want %d\nlayout:\n%s\nportal:\n%s\npaintable:\n%s", popupBox.Y, expectedPopupY, node.GetLayoutTreeString(), node.GetPortalTreeString(), node.GetPaintableTreeString())
	}
}
