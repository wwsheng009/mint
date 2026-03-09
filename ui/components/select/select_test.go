package selectcomp

import (
	"strings"
	"testing"

	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/layout"
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
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

func TestVNode_Children_OverlayPopupChild(t *testing.T) {
	vnode := NewBuilder().
		SetID("country-select").
		OverlayPopup(true).
		AddOption("us", "United States").
		AddOption("cn", "China").
		BuildTyped()

	children := vnode.Children()
	if len(children) != 1 {
		t.Fatalf("Children len = %d, want 1", len(children))
	}
	if children[0].Tag() != "box" {
		t.Fatalf("child Tag() = %q, want %q", children[0].Tag(), "box")
	}
	props := children[0].Props()
	if got, _ := props["portalRoot"].(string); got != rtui.DefaultOverlayPortalRootID {
		t.Fatalf("portalRoot = %q, want %q", got, rtui.DefaultOverlayPortalRootID)
	}
	grandChildren := children[0].Children()
	if len(grandChildren) != 1 || grandChildren[0].Tag() != "select-popup" {
		t.Fatalf("expected popup surface child, got %#v", grandChildren)
	}
}

func TestPopupInstance_Measure_UsesOwnerOpenState(t *testing.T) {
	owner := NewInstance(rtui.Props{
		"key":          "country",
		"ownerID":      "country-select",
		"overlayPopup": true,
		"options": []Option{
			{Value: "us", Label: "United States"},
			{Value: "cn", Label: "China"},
		},
	})
	owner.syncOverlayRegistration()
	defer owner.unregisterOverlay()

	popup := newPopupInstance(rtui.Props{
		"ownerID": "country-select",
	})
	defer popup.unregister()

	size := popup.Measure(layout.UnboundedConstraints())
	if size.Width != 0 || size.Height != 0 {
		t.Fatalf("closed popup Measure = %+v, want zero size", size)
	}

	owner.openDropdown()
	size = popup.Measure(layout.UnboundedConstraints())
	if size.Width <= 0 || size.Height <= 0 {
		t.Fatalf("open popup Measure = %+v, want non-zero size", size)
	}
}

func TestMiddleware_ClickOutsideClosesOpenOverlaySelect(t *testing.T) {
	owner := NewInstance(rtui.Props{
		"key":            "country",
		"ownerID":        "country-select",
		"overlayPopup":   true,
		"closeOnOutside": true,
		"options": []Option{
			{Value: "us", Label: "United States"},
			{Value: "cn", Label: "China"},
		},
	})
	owner.SetBounds(2, 2, 12, 1)
	owner.syncOverlayRegistration()
	defer owner.unregisterOverlay()
	owner.openDropdown()

	popup := newPopupInstance(rtui.Props{"ownerID": "country-select"})
	popup.SetBounds(2, 3, owner.popupWidth(), owner.popupHeight())
	selectOverlayRegistry.registerPopup("country-select", popup)
	defer popup.unregister()

	middleware := NewMiddleware()
	act := action.NewAction(action.ActionClick).WithPayload(
		runtimemsg.NewMouseMsg(40, 20, runtimemsg.MouseLeft, runtimemsg.MouseActionPress),
	)

	if result := middleware.Before(act); result != nil {
		t.Fatal("outside click should be intercepted when overlay select closes")
	}
	if owner.open {
		t.Fatal("overlay select should be closed after outside click")
	}
}
