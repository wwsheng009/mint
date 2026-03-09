package menu

import (
	"testing"
	"time"

	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/runtime/intent"
	rttypes "github.com/wwsheng009/mint/runtime/types"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

type testIntent struct{ name string }

func (i testIntent) IntentType() string { return i.name }

func TestNavigationSkipsNonSelectableItems(t *testing.T) {
	items := []MenuItem{
		LabelItem("label", "Group"),
		Separator(),
		Action("open", "Open", testIntent{"open"}),
		Action("save", "Save", testIntent{"save"}).WithDisabled(true),
		Action("quit", "Quit", testIntent{"quit"}),
	}
	if got := FirstSelectableIndex(items); got != 2 {
		t.Fatalf("FirstSelectableIndex() = %d, want 2", got)
	}
	if got := NextSelectableIndex(items, 2); got != 4 {
		t.Fatalf("NextSelectableIndex() = %d, want 4", got)
	}
	if got := PrevSelectableIndex(items, 4); got != 2 {
		t.Fatalf("PrevSelectableIndex() = %d, want 2", got)
	}
}

func TestCollectShortcutsIncludesNestedItems(t *testing.T) {
	items := []MenuItem{
		Submenu("file", "File",
			Action("new", "New", testIntent{"new"}).WithShortcut("ctrl+n"),
			Action("open", "Open", testIntent{"open"}).WithShortcut("ctrl+o"),
		),
	}
	bindings := CollectShortcuts(items)
	if len(bindings) != 2 {
		t.Fatalf("CollectShortcuts() len = %d, want 2", len(bindings))
	}
	if bindings[0].Path[0] != 0 || bindings[0].Path[1] != 0 {
		t.Fatalf("first shortcut path = %v, want [0 0]", bindings[0].Path)
	}
}

func TestMatchShortcutNormalizesCombo(t *testing.T) {
	items := []MenuItem{Action("save", "Save", testIntent{"save"}).WithShortcut("Ctrl + S")}
	binding, ok := MatchShortcut(items, "ctrl+s")
	if !ok {
		t.Fatal("MatchShortcut() should match normalized combo")
	}
	if binding.Item.Key != "save" {
		t.Fatalf("binding.Item.Key = %q, want save", binding.Item.Key)
	}
}

func TestBuilderBuildPopupCreatesOverlayVNode(t *testing.T) {
	vnode := NewPopup([]MenuItem{Action("open", "Open", testIntent{"open"})}).
		SetID("popup-1").
		Title("Menu").
		MaxHeight(8).
		Build()
	if vnode.Tag() != "menu-popup" {
		t.Fatalf("Tag() = %q, want menu-popup", vnode.Tag())
	}
	if vnode.GetLayer() != rtui.LayerOverlay {
		t.Fatalf("GetLayer() = %v, want LayerOverlay", vnode.GetLayer())
	}
	props := vnode.Props()
	model, ok := props["model"].(Model)
	if !ok {
		t.Fatal("Build() should store model prop")
	}
	if model.Title != "Menu" {
		t.Fatalf("model.Title = %q, want Menu", model.Title)
	}
}

func TestBuilderPreservesPortalProps(t *testing.T) {
	vnode := NewPopup([]MenuItem{Action("open", "Open", testIntent{"open"})}).
		SetID("popup-portal").
		PortalRoot("overlay-root").
		AnchorTo("toolbar.file", rttypes.AnchorBottomLeft).
		PortalPosition(rttypes.PositionFixed).
		PortalPriority(7).
		Build()

	props := vnode.Props()
	if got, _ := props["portalRoot"].(string); got != "overlay-root" {
		t.Fatalf("portalRoot = %q, want overlay-root", got)
	}
	if got, _ := props["anchorId"].(string); got != "toolbar.file" {
		t.Fatalf("anchorId = %q, want toolbar.file", got)
	}
	if got, _ := props["anchor"].(rttypes.Anchor); got != rttypes.AnchorBottomLeft {
		t.Fatalf("anchor = %v, want AnchorBottomLeft", got)
	}
	if got, _ := props["position"].(rttypes.PositionType); got != rttypes.PositionFixed {
		t.Fatalf("position = %v, want PositionFixed", got)
	}
	if got, _ := props["priority"].(int); got != 7 {
		t.Fatalf("priority = %d, want 7", got)
	}
}

func TestPopupInstanceNavigateAndActivate(t *testing.T) {
	vnode := NewPopup([]MenuItem{
		Action("open", "Open", testIntent{"open"}),
		Separator(),
		Action("quit", "Quit", testIntent{"quit"}),
	}).Build()
	inst := vnode.(rtui.InstanceFactory).CreateInstance().(*popupInstance)
	var emitted []intent.Intent
	inst.SetIntentEmitter(func(i intent.Intent) { emitted = append(emitted, i) })
	if inst.selectedIndex != 0 {
		t.Fatalf("selectedIndex = %d, want 0", inst.selectedIndex)
	}
	if handled := inst.HandleAction(action.NewAction(action.ActionNavigateDown)); !handled {
		t.Fatal("navigate down should be handled")
	}
	if inst.selectedIndex != 2 {
		t.Fatalf("selectedIndex = %d, want 2", inst.selectedIndex)
	}
	if handled := inst.HandleAction(action.NewAction(action.ActionEnter)); !handled {
		t.Fatal("enter should be handled")
	}
	if len(emitted) < 2 {
		t.Fatalf("emitted intents = %d, want at least 2", len(emitted))
	}
	if emitted[0].IntentType() != "menu.navigate" {
		t.Fatalf("first intent = %q, want menu.navigate", emitted[0].IntentType())
	}
	if emitted[1].IntentType() != "menu.activate_item" {
		t.Fatalf("second intent = %q, want menu.activate_item", emitted[1].IntentType())
	}
}

func TestPopupInstanceTypeaheadMovesSelection(t *testing.T) {
	vnode := NewPopup([]MenuItem{
		Action("alpha", "Alpha", testIntent{"alpha"}),
		Action("beta", "Beta", testIntent{"beta"}),
		Action("gamma", "Gamma", testIntent{"gamma"}),
	}).Typeahead(true).TypeaheadTimeout(50 * time.Millisecond).Build()

	inst := vnode.(rtui.InstanceFactory).CreateInstance().(*popupInstance)
	if inst.selectedIndex != 0 {
		t.Fatalf("selectedIndex = %d, want 0", inst.selectedIndex)
	}
	if handled := inst.HandleAction(action.NewAction(action.ActionInputText).WithPayload("g")); !handled {
		t.Fatal("typeahead should be handled")
	}
	if inst.selectedIndex != 2 {
		t.Fatalf("selectedIndex = %d, want 2", inst.selectedIndex)
	}
	time.Sleep(60 * time.Millisecond)
	if handled := inst.HandleAction(action.NewAction(action.ActionInputText).WithPayload("b")); !handled {
		t.Fatal("typeahead after timeout should be handled")
	}
	if inst.selectedIndex != 1 {
		t.Fatalf("selectedIndex = %d, want 1", inst.selectedIndex)
	}
}

func TestPathHelpers(t *testing.T) {
	base := []int{1, 2}
	child := ChildPath(base, 3)
	if !PathEqual(child, []int{1, 2, 3}) {
		t.Fatalf("ChildPath() = %v, want [1 2 3]", child)
	}
	parent := ParentPath(child)
	if !PathEqual(parent, base) {
		t.Fatalf("ParentPath() = %v, want %v", parent, base)
	}
}

func TestThemeDefaultsArePopulated(t *testing.T) {
	theme := DefaultTheme()
	if theme.SurfaceStyle.IsEmpty() {
		t.Fatal("DefaultTheme().SurfaceStyle should not be empty")
	}
	if theme.BarActiveStyle.IsEmpty() {
		t.Fatal("DefaultTheme().BarActiveStyle should not be empty")
	}
}
