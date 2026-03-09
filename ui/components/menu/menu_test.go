package menu

import (
	"testing"
	"time"

	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/runtime/intent"
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
	rttypes "github.com/wwsheng009/mint/runtime/types"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

type testIntent struct{ name string }

func (i testIntent) IntentType() string { return i.name }

type fakeRegistrar struct {
	handlers map[string]func()
}

func (r *fakeRegistrar) OnKeyCombo(combo string, handler func()) {
	if r.handlers == nil {
		r.handlers = map[string]func(){}
	}
	r.handlers[combo] = handler
}

type fakeInstallerHost struct {
	fakeRegistrar
	middlewareCount int
}

func (h *fakeInstallerHost) AddMiddleware(_ action.ActionMiddleware) {
	h.middlewareCount++
}

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
	menuRegistryGlobal.reset()
	defer menuRegistryGlobal.reset()
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
	menuRegistryGlobal.reset()
	defer menuRegistryGlobal.reset()
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

func TestPopupInstanceNavigateRightOpensSubmenuCascade(t *testing.T) {
	menuRegistryGlobal.reset()
	defer menuRegistryGlobal.reset()
	vnode := NewPopup([]MenuItem{
		Submenu("file", "File",
			Action("new", "New", testIntent{"new"}),
			Action("open", "Open", testIntent{"open"}),
		),
		Action("quit", "Quit", testIntent{"quit"}),
	}).Build()

	inst := vnode.(rtui.InstanceFactory).CreateInstance().(*popupInstance)
	if handled := inst.HandleAction(action.NewAction(action.ActionNavigateRight)); !handled {
		t.Fatal("navigate right should be handled")
	}
	if len(inst.submenuPath) != 1 {
		t.Fatalf("submenuPath len = %d, want 1", len(inst.submenuPath))
	}
	if inst.submenuPath[0] != 0 {
		t.Fatalf("submenuPath[0] = %d, want 0", inst.submenuPath[0])
	}
	surfaces := inst.popupSurfaces()
	if len(surfaces) != 2 {
		t.Fatalf("popupSurfaces len = %d, want 2", len(surfaces))
	}
}

func TestPopupInstanceControlledActivePath(t *testing.T) {
	menuRegistryGlobal.reset()
	defer menuRegistryGlobal.reset()
	vnode := NewPopup([]MenuItem{
		Submenu("file", "File",
			Action("new", "New", testIntent{"new"}),
			Action("open", "Open", testIntent{"open"}),
		),
		Action("quit", "Quit", testIntent{"quit"}),
	}).ActivePath(0, 1).Build()

	inst := vnode.(rtui.InstanceFactory).CreateInstance().(*popupInstance)
	if inst.selectedIndex != 0 {
		t.Fatalf("selectedIndex = %d, want 0", inst.selectedIndex)
	}
	if len(inst.submenuPath) != 1 || inst.submenuPath[0] != 1 {
		t.Fatalf("submenuPath = %v, want [1]", inst.submenuPath)
	}
	if got := inst.GetProps()["model"].(Model).ActivePath; !PathEqual(got, []int{0, 1}) {
		t.Fatalf("ActivePath = %v, want [0 1]", got)
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

func TestMenuMiddlewareClickOutsideClosesOpenPopup(t *testing.T) {
	menuRegistryGlobal.reset()
	defer menuRegistryGlobal.reset()
	vnode := NewPopup([]MenuItem{Action("open", "Open", testIntent{"open"})}).Build()
	inst := vnode.(rtui.InstanceFactory).CreateInstance().(*popupInstance)
	inst.SetBounds(10, 5, 24, 8)
	inst.OnMount()
	defer inst.Destroy()

	middleware := NewMiddleware()
	act := action.NewAction(action.ActionClick).WithPayload(runtimemsg.NewMouseMsg(1, 1, runtimemsg.MouseLeft, runtimemsg.MouseActionPress))
	if next := middleware.Before(act); next != nil {
		t.Fatal("outside click should be intercepted when popup closes")
	}
	if inst.open {
		t.Fatal("popup should be closed after outside click")
	}
}

func TestMenuMiddlewareLeavesInsideClickAlone(t *testing.T) {
	menuRegistryGlobal.reset()
	defer menuRegistryGlobal.reset()
	vnode := NewPopup([]MenuItem{Action("open", "Open", testIntent{"open"})}).Build()
	inst := vnode.(rtui.InstanceFactory).CreateInstance().(*popupInstance)
	inst.SetBounds(10, 5, 24, 8)
	inst.OnMount()
	defer inst.Destroy()

	middleware := NewMiddleware()
	act := action.NewAction(action.ActionClick).WithPayload(runtimemsg.NewMouseMsg(12, 6, runtimemsg.MouseLeft, runtimemsg.MouseActionPress))
	if next := middleware.Before(act); next == nil {
		t.Fatal("inside click should continue dispatch")
	}
	if !inst.open {
		t.Fatal("popup should remain open after inside click")
	}
}

func TestRegisterGlobalShortcutsRegistersGlobalBindings(t *testing.T) {
	registrar := &fakeRegistrar{}
	var emitted []intent.Intent
	count := RegisterGlobalShortcuts(registrar, "main-menu", []MenuItem{
		Action("open", "Open", testIntent{"open"}).WithShortcut("ctrl+o"),
		Action("save", "Save", testIntent{"save"}).WithShortcut("ctrl+s").WithShortcutScope(ShortcutGlobal),
		Action("local", "Local", testIntent{"local"}).WithShortcut("alt+l").WithShortcutScope(ShortcutLocal),
	}, func(i intent.Intent) {
		emitted = append(emitted, i)
	})
	if count != 2 {
		t.Fatalf("RegisterGlobalShortcuts() = %d, want 2", count)
	}
	if len(registrar.handlers) != 2 {
		t.Fatalf("registered handlers = %d, want 2", len(registrar.handlers))
	}
	handler := registrar.handlers["ctrl+o"]
	if handler == nil {
		t.Fatal("ctrl+o should be registered")
	}
	handler()
	if len(emitted) != 2 {
		t.Fatalf("emitted intents = %d, want 2", len(emitted))
	}
	if emitted[0].IntentType() != "menu.activate_item" {
		t.Fatalf("first emitted intent = %q, want menu.activate_item", emitted[0].IntentType())
	}
	if emitted[1].IntentType() != "open" {
		t.Fatalf("second emitted intent = %q, want open", emitted[1].IntentType())
	}
}

func TestInstallAddsMiddlewareOnceAndDedupsShortcuts(t *testing.T) {
	host := &fakeInstallerHost{}
	var emitted []intent.Intent
	builderA := NewPopup([]MenuItem{
		Action("open", "Open", testIntent{"open"}).WithShortcut("ctrl+o"),
	}).RegisterShortcuts(true)
	builderB := NewPopup([]MenuItem{
		Action("other-open", "Other Open", testIntent{"other-open"}).WithShortcut("ctrl+o"),
		Action("save", "Save", testIntent{"save"}).WithShortcut("ctrl+s"),
	}).RegisterShortcuts(true)

	count := Install(host, func(i intent.Intent) {
		emitted = append(emitted, i)
	}, builderA, builderB)
	if count != 2 {
		t.Fatalf("Install() = %d, want 2", count)
	}
	if host.middlewareCount != 1 {
		t.Fatalf("middlewareCount = %d, want 1", host.middlewareCount)
	}
	if len(host.handlers) != 2 {
		t.Fatalf("handlers len = %d, want 2", len(host.handlers))
	}

	count = Install(host, func(i intent.Intent) {
		emitted = append(emitted, i)
	}, builderA, builderB)
	if count != 0 {
		t.Fatalf("second Install() = %d, want 0", count)
	}
	if host.middlewareCount != 1 {
		t.Fatalf("middlewareCount after second install = %d, want 1", host.middlewareCount)
	}
}
