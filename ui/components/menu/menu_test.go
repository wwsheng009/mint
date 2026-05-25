package menu

import (
	"testing"
	"time"

	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/runtime/intent"
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
	rttypes "github.com/wwsheng009/mint/runtime/types"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	overlayposition "github.com/wwsheng009/mint/ui/components/internal/overlayposition"
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
	userData        map[string]interface{}
}

func (h *fakeInstallerHost) AddMiddleware(_ action.ActionMiddleware) {
	h.middlewareCount++
}

func (h *fakeInstallerHost) SetUserData(key string, value interface{}) {
	if h.userData == nil {
		h.userData = map[string]interface{}{}
	}
	h.userData[key] = value
}

func (h *fakeInstallerHost) GetUserData(key string) interface{} {
	if h.userData == nil {
		return nil
	}
	return h.userData[key]
}

func buildPopupSurface(items []MenuItem) *popupVNode {
	model := NewPopup(items).BuildModel()
	return newPopupVNode(clearPortalModel(model))
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

func TestOperationalMenuPresets(t *testing.T) {
	refresh := RefreshAction(testIntent{"refresh"})
	if refresh.Key != "refresh" || refresh.Label != "Refresh" || refresh.Description == "" {
		t.Fatalf("refresh preset = %+v", refresh)
	}
	if refresh.Shortcut.Combo != "r" {
		t.Fatalf("refresh shortcut = %q, want r", refresh.Shortcut.Combo)
	}

	reload := ReloadRuntimeAction(testIntent{"reload"})
	if !reload.Danger || reload.Key != "reload-runtime" || reload.Description == "" {
		t.Fatalf("reload preset = %+v", reload)
	}

	disabled := DisabledAction("reset-key", "Reset Key", "Select a key first", testIntent{"reset"})
	if !disabled.Disabled {
		t.Fatal("disabled action should be disabled")
	}
	if disabled.Description != "Select a key first" {
		t.Fatalf("disabled description = %q", disabled.Description)
	}
	if disabled.Metadata["disabledReason"] != "Select a key first" {
		t.Fatalf("disabled reason metadata = %#v", disabled.Metadata["disabledReason"])
	}

	group := Group("Runtime Actions", refresh, reload)
	if len(group) != 3 {
		t.Fatalf("group len = %d, want 3", len(group))
	}
	if !group[0].IsLabel() || group[0].Key != "group-runtime-actions" {
		t.Fatalf("group label = %+v", group[0])
	}
}

func TestBuilderBuildPopupCreatesOverlayVNode(t *testing.T) {
	vnode := NewPopup([]MenuItem{Action("open", "Open", testIntent{"open"})}).
		SetID("popup-1").
		Title("Menu").
		MaxHeight(8).
		Build()
	if vnode.Tag() != "box" {
		t.Fatalf("Tag() = %q, want box portal wrapper", vnode.Tag())
	}
	if vnode.GetLayer() != rtui.LayerOverlay {
		t.Fatalf("GetLayer() = %v, want LayerOverlay", vnode.GetLayer())
	}
	props := vnode.Props()
	if got, _ := props["portalRoot"].(string); got != rtui.DefaultOverlayPortalRootID {
		t.Fatalf("portalRoot = %q, want %q", got, rtui.DefaultOverlayPortalRootID)
	}
	children := vnode.Children()
	if len(children) != 1 {
		t.Fatalf("wrapper children = %d, want 1", len(children))
	}
	if children[0].Tag() != "menu-popup" {
		t.Fatalf("child Tag() = %q, want menu-popup", children[0].Tag())
	}
}

func TestBuilderAnchorDefaultsToAbsolutePortalPosition(t *testing.T) {
	model := NewPopup([]MenuItem{Action("open", "Open", testIntent{"open"})}).
		AnchorTo("toolbar.file", rttypes.AnchorTopLeft).
		BuildModel()
	if model.PortalPosition != rttypes.PositionAbsolute {
		t.Fatalf("PortalPosition = %v, want PositionAbsolute", model.PortalPosition)
	}
}

func TestBuilderDefaultsToFrameworkOverlayPortalRoot(t *testing.T) {
	model := NewPopup([]MenuItem{Action("open", "Open", testIntent{"open"})}).BuildModel()
	if model.PortalRoot != rtui.DefaultOverlayPortalRootID {
		t.Fatalf("PortalRoot = %q, want %q", model.PortalRoot, rtui.DefaultOverlayPortalRootID)
	}
}

func TestBuilderPreservesPortalProps(t *testing.T) {
	vnode := NewPopup([]MenuItem{Action("open", "Open", testIntent{"open"})}).
		SetID("popup-portal").
		PortalRoot("overlay-root").
		AnchorTo("toolbar.file", rttypes.AnchorBottomLeft).
		PortalPosition(rttypes.PositionFixed).
		PortalPriority(7).
		PortalOffset(11, 2).
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
	if got, _ := props["left"].(int); got != 11 {
		t.Fatalf("left = %d, want 11", got)
	}
	if got, _ := props["top"].(int); got != 2 {
		t.Fatalf("top = %d, want 2", got)
	}
}

func TestBuilderPlacementBottomEndUsesRootPopupMetrics(t *testing.T) {
	vnode := NewPopup([]MenuItem{Action("open", "Open", testIntent{"open"})}).
		AnchorTo("toolbar.file", rttypes.AnchorBottomLeft).
		Placement(PlacementBottomEnd).
		Build()

	props := vnode.Props()
	if got, _ := props["anchor"].(rttypes.Anchor); got != rttypes.AnchorBottomRight {
		t.Fatalf("anchor = %v, want AnchorBottomRight", got)
	}

	child := vnode.Children()[0].(*popupVNode)
	inst := child.CreateInstance().(*popupInstance)
	root := inst.popupSurfaces()[0]
	outerWidth := root.metrics.surfaceWidth + root.metrics.shadowWidth
	outerHeight := root.metrics.surfaceHeight + root.metrics.shadowHeight

	if got, _ := props["left"].(int); got != 0 {
		t.Fatalf("left = %d, want 0", got)
	}
	if got, _ := props["top"].(int); got != outerHeight {
		t.Fatalf("top = %d, want %d", got, outerHeight)
	}
	if got, _ := props["width"].(int); got != outerWidth {
		t.Fatalf("width = %d, want %d", got, outerWidth)
	}
	if got, _ := props["height"].(int); got != outerHeight {
		t.Fatalf("height = %d, want %d", got, outerHeight)
	}
	if got, _ := props["positioningWidth"].(int); got != outerWidth {
		t.Fatalf("positioningWidth = %d, want %d", got, outerWidth)
	}
	if got, _ := props["positioningHeight"].(int); got != outerHeight {
		t.Fatalf("positioningHeight = %d, want %d", got, outerHeight)
	}
}

func TestBuilderPlacementRightStartUsesRootWidthOffset(t *testing.T) {
	vnode := NewPopup([]MenuItem{Action("open", "Open", testIntent{"open"})}).
		AnchorTo("toolbar.file", rttypes.AnchorBottomLeft).
		Placement(PlacementRightStart).
		Build()

	props := vnode.Props()
	if got, _ := props["anchor"].(rttypes.Anchor); got != rttypes.AnchorTopRight {
		t.Fatalf("anchor = %v, want AnchorTopRight", got)
	}

	child := vnode.Children()[0].(*popupVNode)
	inst := child.CreateInstance().(*popupInstance)
	root := inst.popupSurfaces()[0]
	outerWidth := root.metrics.surfaceWidth + root.metrics.shadowWidth
	outerHeight := root.metrics.surfaceHeight + root.metrics.shadowHeight

	if got, _ := props["left"].(int); got != outerWidth {
		t.Fatalf("left = %d, want %d", got, outerWidth)
	}
	if got, _ := props["top"].(int); got != 0 {
		t.Fatalf("top = %d, want 0", got)
	}
	if got, _ := props["width"].(int); got != outerWidth {
		t.Fatalf("width = %d, want %d", got, outerWidth)
	}
	if got, _ := props["height"].(int); got != outerHeight {
		t.Fatalf("height = %d, want %d", got, outerHeight)
	}
	if got, _ := props["positioningWidth"].(int); got != outerWidth {
		t.Fatalf("positioningWidth = %d, want %d", got, outerWidth)
	}
	if got, _ := props["positioningHeight"].(int); got != outerHeight {
		t.Fatalf("positioningHeight = %d, want %d", got, outerHeight)
	}
}

func TestBuilderPlacementAutoDefaultsToBottomStart(t *testing.T) {
	vnode := NewPopup([]MenuItem{Action("open", "Open", testIntent{"open"})}).
		AnchorTo("toolbar.file", rttypes.AnchorBottomLeft).
		Placement(PlacementAuto).
		Build()

	props := vnode.Props()
	if got, _ := props["anchor"].(rttypes.Anchor); got != rttypes.AnchorBottomLeft {
		t.Fatalf("anchor = %v, want AnchorBottomLeft", got)
	}

	child := vnode.Children()[0].(*popupVNode)
	inst := child.CreateInstance().(*popupInstance)
	root := inst.popupSurfaces()[0]
	outerHeight := root.metrics.surfaceHeight + root.metrics.shadowHeight

	if got, _ := props["left"].(int); got != 0 {
		t.Fatalf("left = %d, want 0", got)
	}
	if got, _ := props["top"].(int); got != outerHeight {
		t.Fatalf("top = %d, want %d", got, outerHeight)
	}
}

func TestBuilderPlacementAutoTracksTopRightAnchor(t *testing.T) {
	vnode := NewPopup([]MenuItem{Action("open", "Open", testIntent{"open"})}).
		AnchorTo("toolbar.file", rttypes.AnchorTopRight).
		Placement(PlacementAuto).
		Build()

	props := vnode.Props()
	if got, _ := props["anchor"].(rttypes.Anchor); got != rttypes.AnchorTopRight {
		t.Fatalf("anchor = %v, want AnchorTopRight", got)
	}

	child := vnode.Children()[0].(*popupVNode)
	inst := child.CreateInstance().(*popupInstance)
	root := inst.popupSurfaces()[0]
	outerHeight := root.metrics.surfaceHeight + root.metrics.shadowHeight

	if got, _ := props["left"].(int); got != 0 {
		t.Fatalf("left = %d, want 0", got)
	}
	if got, _ := props["top"].(int); got != -outerHeight {
		t.Fatalf("top = %d, want %d", got, -outerHeight)
	}
}

func TestBuilderPlacementAutoTracksRightAnchor(t *testing.T) {
	vnode := NewPopup([]MenuItem{Action("open", "Open", testIntent{"open"})}).
		AnchorTo("toolbar.file", rttypes.AnchorRight).
		Placement(PlacementAuto).
		Build()

	props := vnode.Props()
	if got, _ := props["anchor"].(rttypes.Anchor); got != rttypes.AnchorTopRight {
		t.Fatalf("anchor = %v, want AnchorTopRight", got)
	}

	child := vnode.Children()[0].(*popupVNode)
	inst := child.CreateInstance().(*popupInstance)
	root := inst.popupSurfaces()[0]
	outerWidth := root.metrics.surfaceWidth + root.metrics.shadowWidth

	if got, _ := props["left"].(int); got != outerWidth {
		t.Fatalf("left = %d, want %d", got, outerWidth)
	}
	if got, _ := props["top"].(int); got != 0 {
		t.Fatalf("top = %d, want 0", got)
	}
}

func TestBuilderPlacementStoresPopupPlacementMetadata(t *testing.T) {
	vnode := NewPopup([]MenuItem{Action("open", "Open", testIntent{"open"})}).
		AnchorTo("toolbar.file", rttypes.AnchorTopRight).
		Placement(PlacementAuto).
		PortalOffset(3, 2).
		Build()

	props := vnode.Props()
	if got, _ := props["popupPlacement"].(string); got != string(PlacementTopEnd) {
		t.Fatalf("popupPlacement = %q, want %q", got, PlacementTopEnd)
	}
	if got, _ := props["popupOffsetX"].(int); got != 3 {
		t.Fatalf("popupOffsetX = %d, want 3", got)
	}
	if got, _ := props["popupOffsetY"].(int); got != 2 {
		t.Fatalf("popupOffsetY = %d, want 2", got)
	}
}

func TestContextMenuStoresViewportClampMetadata(t *testing.T) {
	vnode := NewContextMenu([]MenuItem{Action("open", "Open", testIntent{"open"})}).
		PortalOffset(58, 16).
		Build()

	props := vnode.Props()
	child := vnode.Children()[0].(*popupVNode)
	inst := child.CreateInstance().(*popupInstance)
	root := inst.popupSurfaces()[0]
	outerWidth := root.metrics.surfaceWidth + root.metrics.shadowWidth
	outerHeight := root.metrics.surfaceHeight + root.metrics.shadowHeight

	if got, _ := props["popupClampToViewport"].(bool); !got {
		t.Fatalf("popupClampToViewport = %v, want true", got)
	}
	if got, _ := props["positioningWidth"].(int); got != outerWidth {
		t.Fatalf("positioningWidth = %d, want %d", got, outerWidth)
	}
	if got, _ := props["positioningHeight"].(int); got != outerHeight {
		t.Fatalf("positioningHeight = %d, want %d", got, outerHeight)
	}
}

func TestPopupInstanceNavigateAndActivate(t *testing.T) {
	menuRegistryGlobal.reset()
	defer menuRegistryGlobal.reset()
	vnode := buildPopupSurface([]MenuItem{
		Action("open", "Open", testIntent{"open"}),
		Separator(),
		Action("quit", "Quit", testIntent{"quit"}),
	})
	inst := vnode.CreateInstance().(*popupInstance)
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
	model := NewPopup([]MenuItem{
		Action("alpha", "Alpha", testIntent{"alpha"}),
		Action("beta", "Beta", testIntent{"beta"}),
		Action("gamma", "Gamma", testIntent{"gamma"}),
	}).Typeahead(true).TypeaheadTimeout(50 * time.Millisecond).BuildModel()
	vnode := newPopupVNode(clearPortalModel(model))

	inst := vnode.CreateInstance().(*popupInstance)
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
	vnode := buildPopupSurface([]MenuItem{
		Submenu("file", "File",
			Action("new", "New", testIntent{"new"}),
			Action("open", "Open", testIntent{"open"}),
		),
		Action("quit", "Quit", testIntent{"quit"}),
	})

	inst := vnode.CreateInstance().(*popupInstance)
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
	model := NewPopup([]MenuItem{
		Submenu("file", "File",
			Action("new", "New", testIntent{"new"}),
			Action("open", "Open", testIntent{"open"}),
		),
		Action("quit", "Quit", testIntent{"quit"}),
	}).ActivePath(0, 1).BuildModel()
	vnode := newPopupVNode(clearPortalModel(model))

	inst := vnode.CreateInstance().(*popupInstance)
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

func TestPopupInstanceSubmenuFlipsLeftWithinViewport(t *testing.T) {
	menuRegistryGlobal.reset()
	defer menuRegistryGlobal.reset()

	model := NewPopup([]MenuItem{
		Submenu("tools", "More Tools",
			Action("details", "Tool Details", testIntent{"details"}),
			Action("inspect", "Inspect Tool", testIntent{"inspect"}),
		),
		Action("quit", "Quit", testIntent{"quit"}),
	}).ActivePath(0, 0).BuildModel()
	inst := newPopupVNode(clearPortalModel(model)).CreateInstance().(*popupInstance)

	rootMetrics := inst.popupMetricsFor(inst.model.Items)
	childMetrics := inst.popupMetricsFor(inst.model.Items[0].Normalize().Children)
	rootOuterWidth := popupSurfaceOuterWidth(rootMetrics)
	rootOuterHeight := popupSurfaceOuterHeight(rootMetrics)
	childOuterWidth := popupSurfaceOuterWidth(childMetrics)
	viewportWidth := rootOuterWidth + childOuterWidth + 2

	inst.SetViewportSize(viewportWidth, 18)
	inst.SetBounds(viewportWidth-rootOuterWidth, 1, rootOuterWidth, rootOuterHeight)

	surfaces := inst.popupSurfaces()
	if len(surfaces) != 2 {
		t.Fatalf("popupSurfaces len = %d, want 2", len(surfaces))
	}
	if surfaces[1].x != -childOuterWidth {
		t.Fatalf("submenu x = %d, want %d for left-start flip", surfaces[1].x, -childOuterWidth)
	}

	boundsX, _, boundsWidth, _ := inst.GetBounds()
	if boundsX != viewportWidth-rootOuterWidth-childOuterWidth {
		t.Fatalf("hit bounds x = %d, want %d", boundsX, viewportWidth-rootOuterWidth-childOuterWidth)
	}
	if boundsWidth != rootOuterWidth+childOuterWidth {
		t.Fatalf("hit bounds width = %d, want %d", boundsWidth, rootOuterWidth+childOuterWidth)
	}
}

func TestPopupInstanceSubmenuClampsVerticallyWithinViewport(t *testing.T) {
	menuRegistryGlobal.reset()
	defer menuRegistryGlobal.reset()

	model := NewPopup([]MenuItem{
		Action("new", "New", testIntent{"new"}),
		Action("open", "Open", testIntent{"open"}),
		Action("save", "Save", testIntent{"save"}),
		Action("export", "Export", testIntent{"export"}),
		Submenu("tools", "More Tools",
			Action("details", "Tool Details", testIntent{"details"}),
			Action("inspect", "Inspect Tool", testIntent{"inspect"}),
			Action("history", "History", testIntent{"history"}),
			Action("recent", "Recent", testIntent{"recent"}),
			Action("backup", "Backup", testIntent{"backup"}),
		),
	}).ActivePath(4, 0).BuildModel()
	inst := newPopupVNode(clearPortalModel(model)).CreateInstance().(*popupInstance)

	rootMetrics := inst.popupMetricsFor(inst.model.Items)
	childMetrics := inst.popupMetricsFor(inst.model.Items[4].Normalize().Children)
	rootOuterWidth := popupSurfaceOuterWidth(rootMetrics)
	rootOuterHeight := popupSurfaceOuterHeight(rootMetrics)
	childOuterHeight := popupSurfaceOuterHeight(childMetrics)
	viewportHeight := childOuterHeight + 1

	inst.SetViewportSize(rootOuterWidth*2+4, viewportHeight)
	inst.SetBounds(4, viewportHeight-rootOuterHeight, rootOuterWidth, rootOuterHeight)

	surfaces := inst.popupSurfaces()
	if len(surfaces) != 2 {
		t.Fatalf("popupSurfaces len = %d, want 2", len(surfaces))
	}

	baseY := inst.layoutBounds[1]
	submenuTop := baseY + surfaces[1].y
	if submenuTop != viewportHeight-childOuterHeight {
		t.Fatalf("submenu top = %d, want %d after bottom-edge clamp", submenuTop, viewportHeight-childOuterHeight)
	}
	if submenuTop+childOuterHeight > viewportHeight {
		t.Fatalf("submenu bottom edge = %d, want <= %d", submenuTop+childOuterHeight, viewportHeight)
	}
}

func TestPopupInstanceNestedSubmenusPreserveFlippedDirectionWhenViewportAllows(t *testing.T) {
	menuRegistryGlobal.reset()
	defer menuRegistryGlobal.reset()

	model := NewPopup([]MenuItem{
		Submenu("tools", "More Tools",
			Submenu("advanced", "Advanced Tools",
				Action("deep", "Deep Action", testIntent{"deep"}),
			),
		),
		Action("quit", "Quit", testIntent{"quit"}),
	}).ActivePath(0, 0, 0).BuildModel()
	inst := newPopupVNode(clearPortalModel(model)).CreateInstance().(*popupInstance)

	rootMetrics := inst.popupMetricsFor(inst.model.Items)
	firstMetrics := inst.popupMetricsFor(inst.model.Items[0].Normalize().Children)
	secondMetrics := inst.popupMetricsFor(inst.model.Items[0].Normalize().Children[0].Normalize().Children)
	rootOuterWidth := popupSurfaceOuterWidth(rootMetrics)
	rootOuterHeight := popupSurfaceOuterHeight(rootMetrics)
	firstOuterWidth := popupSurfaceOuterWidth(firstMetrics)
	secondOuterWidth := popupSurfaceOuterWidth(secondMetrics)
	viewportWidth := rootOuterWidth + firstOuterWidth + secondOuterWidth + 2

	inst.SetViewportSize(viewportWidth, 18)
	inst.SetBounds(viewportWidth-rootOuterWidth, 1, rootOuterWidth, rootOuterHeight)

	surfaces := inst.popupSurfaces()
	if len(surfaces) != 3 {
		t.Fatalf("popupSurfaces len = %d, want 3", len(surfaces))
	}
	if surfaces[1].x != -firstOuterWidth {
		t.Fatalf("first submenu x = %d, want %d", surfaces[1].x, -firstOuterWidth)
	}
	if surfaces[2].x != -firstOuterWidth-secondOuterWidth {
		t.Fatalf("second submenu x = %d, want %d to continue left cascade", surfaces[2].x, -firstOuterWidth-secondOuterWidth)
	}

	boundsX, _, boundsWidth, _ := inst.GetBounds()
	if boundsX != viewportWidth-rootOuterWidth-firstOuterWidth-secondOuterWidth {
		t.Fatalf("hit bounds x = %d, want %d", boundsX, viewportWidth-rootOuterWidth-firstOuterWidth-secondOuterWidth)
	}
	if boundsWidth != rootOuterWidth+firstOuterWidth+secondOuterWidth {
		t.Fatalf("hit bounds width = %d, want %d", boundsWidth, rootOuterWidth+firstOuterWidth+secondOuterWidth)
	}
}

func TestPopupInstanceNestedSubmenusInferDirectionFromResolvedClampSide(t *testing.T) {
	menuRegistryGlobal.reset()
	defer menuRegistryGlobal.reset()

	model := NewPopup([]MenuItem{
		Submenu("ops", "Operations",
			Submenu("massive", "Extremely Wide Branch Options",
				Action("deep", "Ultra Recovery Action Path", testIntent{"deep"}),
			),
		),
		Action("quit", "Quit", testIntent{"quit"}),
	}).ActivePath(0, 0, 0).BuildModel()
	inst := newPopupVNode(clearPortalModel(model)).CreateInstance().(*popupInstance)

	rootMetrics := inst.popupMetricsFor(inst.model.Items)
	secondMetrics := inst.popupMetricsFor(inst.model.Items[0].Normalize().Children[0].Normalize().Children)
	rootOuterWidth := popupSurfaceOuterWidth(rootMetrics)
	rootOuterHeight := popupSurfaceOuterHeight(rootMetrics)
	secondOuterWidth := popupSurfaceOuterWidth(secondMetrics)
	viewportWidth := rootOuterWidth + 18

	inst.SetViewportSize(viewportWidth, 18)
	inst.SetBounds(viewportWidth-rootOuterWidth, 1, rootOuterWidth, rootOuterHeight)

	surfaces := inst.popupSurfaces()
	if len(surfaces) != 3 {
		t.Fatalf("popupSurfaces len = %d, want 3", len(surfaces))
	}
	firstLeft := inst.layoutBounds[0] + surfaces[1].x
	if firstLeft >= inst.layoutBounds[0] {
		t.Fatalf("first submenu left = %d, want < root left = %d after clamp", firstLeft, inst.layoutBounds[0])
	}
	if surfaces[1].direction != overlayposition.CascadeLeft {
		t.Fatalf("first submenu direction = %v, want overlayposition.CascadeLeft after resolved clamp", surfaces[1].direction)
	}
	if surfaces[2].x != -inst.layoutBounds[0] {
		t.Fatalf("second submenu x = %d, want %d to clamp against viewport left edge", surfaces[2].x, -inst.layoutBounds[0])
	}
	if secondOuterWidth <= firstLeft {
		t.Fatalf("test setup invalid: second width = %d should exceed first left margin = %d", secondOuterWidth, firstLeft)
	}
}

func TestPopupInstanceNestedSubmenusClampUpwardNearBottomRightCorner(t *testing.T) {
	menuRegistryGlobal.reset()
	defer menuRegistryGlobal.reset()

	model := NewPopup([]MenuItem{
		Action("new", "New", testIntent{"new"}),
		Action("open", "Open", testIntent{"open"}),
		Action("save", "Save", testIntent{"save"}),
		Action("export", "Export", testIntent{"export"}),
		Submenu("tools", "More Tools",
			Action("scan", "Scan", testIntent{"scan"}),
			Action("repair", "Repair", testIntent{"repair"}),
			Action("archive", "Archive", testIntent{"archive"}),
			Action("cleanup", "Cleanup", testIntent{"cleanup"}),
			Submenu("advanced", "Advanced Tools",
				Action("deep", "Deep Action", testIntent{"deep"}),
				Action("verify", "Verify", testIntent{"verify"}),
				Action("history", "History", testIntent{"history"}),
				Action("reindex", "Reindex", testIntent{"reindex"}),
				Action("recover", "Recover", testIntent{"recover"}),
			),
		),
	}).ActivePath(4, 4, 0).BuildModel()
	inst := newPopupVNode(clearPortalModel(model)).CreateInstance().(*popupInstance)

	rootMetrics := inst.popupMetricsFor(inst.model.Items)
	firstMetrics := inst.popupMetricsFor(inst.model.Items[4].Normalize().Children)
	secondMetrics := inst.popupMetricsFor(inst.model.Items[4].Normalize().Children[4].Normalize().Children)
	rootOuterWidth := popupSurfaceOuterWidth(rootMetrics)
	rootOuterHeight := popupSurfaceOuterHeight(rootMetrics)
	firstOuterWidth := popupSurfaceOuterWidth(firstMetrics)
	secondOuterWidth := popupSurfaceOuterWidth(secondMetrics)
	secondOuterHeight := popupSurfaceOuterHeight(secondMetrics)
	viewportWidth := rootOuterWidth + firstOuterWidth + secondOuterWidth + 2
	viewportHeight := secondOuterHeight + 1

	inst.SetViewportSize(viewportWidth, viewportHeight)
	inst.SetBounds(viewportWidth-rootOuterWidth, viewportHeight-rootOuterHeight, rootOuterWidth, rootOuterHeight)

	surfaces := inst.popupSurfaces()
	if len(surfaces) != 3 {
		t.Fatalf("popupSurfaces len = %d, want 3", len(surfaces))
	}
	if surfaces[1].x != -firstOuterWidth {
		t.Fatalf("first submenu x = %d, want %d", surfaces[1].x, -firstOuterWidth)
	}
	if surfaces[2].x != -firstOuterWidth-secondOuterWidth {
		t.Fatalf("second submenu x = %d, want %d", surfaces[2].x, -firstOuterWidth-secondOuterWidth)
	}

	baseY := inst.layoutBounds[1]
	firstTop := baseY + surfaces[1].y
	secondTop := baseY + surfaces[2].y
	if firstTop+popupSurfaceOuterHeight(firstMetrics) > viewportHeight {
		t.Fatalf("first submenu bottom edge = %d, want <= %d", firstTop+popupSurfaceOuterHeight(firstMetrics), viewportHeight)
	}
	if secondTop != viewportHeight-secondOuterHeight {
		t.Fatalf("second submenu top = %d, want %d after bottom-edge clamp", secondTop, viewportHeight-secondOuterHeight)
	}
}

func TestPopupInstanceNestedSubmenusClampLeftAndUpwardInNarrowBottomCorner(t *testing.T) {
	menuRegistryGlobal.reset()
	defer menuRegistryGlobal.reset()

	model := NewPopup([]MenuItem{
		Action("new", "New", testIntent{"new"}),
		Action("open", "Open", testIntent{"open"}),
		Action("save", "Save", testIntent{"save"}),
		Action("export", "Export", testIntent{"export"}),
		Submenu("ops", "Operations",
			Action("scan", "Scan", testIntent{"scan"}),
			Action("repair", "Repair", testIntent{"repair"}),
			Action("archive", "Archive", testIntent{"archive"}),
			Action("cleanup", "Cleanup", testIntent{"cleanup"}),
			Submenu("massive", "Extremely Wide Branch Options",
				Action("deep", "Ultra Recovery Action Path", testIntent{"deep"}),
				Action("verify", "Verify Recovery Journal", testIntent{"verify"}),
				Action("history", "Inspect Historical Recovery Entries", testIntent{"history"}),
				Action("reindex", "Reindex Recovery Snapshots", testIntent{"reindex"}),
				Action("recover", "Recover Snapshot Chain", testIntent{"recover"}),
			),
		),
	}).ActivePath(4, 4, 0).BuildModel()
	inst := newPopupVNode(clearPortalModel(model)).CreateInstance().(*popupInstance)

	rootMetrics := inst.popupMetricsFor(inst.model.Items)
	secondMetrics := inst.popupMetricsFor(inst.model.Items[4].Normalize().Children[4].Normalize().Children)
	rootOuterWidth := popupSurfaceOuterWidth(rootMetrics)
	rootOuterHeight := popupSurfaceOuterHeight(rootMetrics)
	secondOuterWidth := popupSurfaceOuterWidth(secondMetrics)
	secondOuterHeight := popupSurfaceOuterHeight(secondMetrics)
	viewportWidth := rootOuterWidth + 18
	viewportHeight := secondOuterHeight + 1

	inst.SetViewportSize(viewportWidth, viewportHeight)
	inst.SetBounds(viewportWidth-rootOuterWidth, viewportHeight-rootOuterHeight, rootOuterWidth, rootOuterHeight)

	surfaces := inst.popupSurfaces()
	if len(surfaces) != 3 {
		t.Fatalf("popupSurfaces len = %d, want 3", len(surfaces))
	}

	firstLeft := inst.layoutBounds[0] + surfaces[1].x
	if firstLeft >= inst.layoutBounds[0] {
		t.Fatalf("first submenu left = %d, want < root left = %d after clamp", firstLeft, inst.layoutBounds[0])
	}
	if surfaces[1].direction != overlayposition.CascadeLeft {
		t.Fatalf("first submenu direction = %v, want overlayposition.CascadeLeft after resolved clamp", surfaces[1].direction)
	}
	if surfaces[2].x != -inst.layoutBounds[0] {
		t.Fatalf("second submenu x = %d, want %d to clamp against viewport left edge", surfaces[2].x, -inst.layoutBounds[0])
	}
	secondTop := inst.layoutBounds[1] + surfaces[2].y
	if secondTop != viewportHeight-secondOuterHeight {
		t.Fatalf("second submenu top = %d, want %d after bottom-edge clamp", secondTop, viewportHeight-secondOuterHeight)
	}
	if secondOuterWidth <= firstLeft {
		t.Fatalf("test setup invalid: second width = %d should exceed first left margin = %d", secondOuterWidth, firstLeft)
	}

	boundsX, boundsY, _, boundsHeight := inst.GetBounds()
	if boundsX != 0 {
		t.Fatalf("hit bounds x = %d, want 0 after left-edge clamp", boundsX)
	}
	if boundsY > secondTop {
		t.Fatalf("hit bounds y = %d, want <= second submenu top = %d", boundsY, secondTop)
	}
	if boundsY+boundsHeight > viewportHeight {
		t.Fatalf("hit bounds bottom edge = %d, want <= %d", boundsY+boundsHeight, viewportHeight)
	}
}

func TestPopupInstanceNestedSubmenusMirrorRightAfterLeftEdgeClampDirection(t *testing.T) {
	menuRegistryGlobal.reset()
	defer menuRegistryGlobal.reset()

	model := NewPopup([]MenuItem{
		Action("new", "New", testIntent{"new"}),
		Action("open", "Open", testIntent{"open"}),
		Action("save", "Save", testIntent{"save"}),
		Action("export", "Export", testIntent{"export"}),
		Submenu("ops", "Operations",
			Action("pad", "Inspect Intermediate Recovery Ledger", testIntent{"pad"}),
			Action("repair", "Repair", testIntent{"repair"}),
			Action("archive", "Archive", testIntent{"archive"}),
			Submenu("branch", "Branch",
				Action("node", "Moderately Wide Inner Recovery Journal", testIntent{"node"}),
				Submenu("pivot", "Pivot",
					Action("deep", "Deep Action", testIntent{"deep"}),
				),
			),
		),
	}).ActivePath(4, 3, 1, 0).BuildModel()
	inst := newPopupVNode(clearPortalModel(model)).CreateInstance().(*popupInstance)

	rootMetrics := inst.popupMetricsFor(inst.model.Items)
	firstMetrics := inst.popupMetricsFor(inst.model.Items[4].Normalize().Children)
	secondMetrics := inst.popupMetricsFor(inst.model.Items[4].Normalize().Children[3].Normalize().Children)
	thirdMetrics := inst.popupMetricsFor(inst.model.Items[4].Normalize().Children[3].Normalize().Children[1].Normalize().Children)
	rootOuterWidth := popupSurfaceOuterWidth(rootMetrics)
	rootOuterHeight := popupSurfaceOuterHeight(rootMetrics)
	firstOuterWidth := popupSurfaceOuterWidth(firstMetrics)
	secondOuterWidth := popupSurfaceOuterWidth(secondMetrics)
	thirdOuterWidth := popupSurfaceOuterWidth(thirdMetrics)
	viewportWidth := rootOuterWidth + firstOuterWidth + 2
	viewportHeight := 18

	inst.SetViewportSize(viewportWidth, viewportHeight)
	inst.SetBounds(viewportWidth-rootOuterWidth, 2, rootOuterWidth, rootOuterHeight)

	surfaces := inst.popupSurfaces()
	if len(surfaces) != 4 {
		t.Fatalf("popupSurfaces len = %d, want 4", len(surfaces))
	}
	if surfaces[1].direction != overlayposition.CascadeLeft {
		t.Fatalf("first submenu direction = %v, want overlayposition.CascadeLeft", surfaces[1].direction)
	}
	if surfaces[2].x != -inst.layoutBounds[0] {
		t.Fatalf("second submenu x = %d, want %d after left-edge clamp", surfaces[2].x, -inst.layoutBounds[0])
	}
	if surfaces[2].direction != overlayposition.CascadeLeft {
		t.Fatalf("second submenu direction = %v, want overlayposition.CascadeLeft after resolved clamp", surfaces[2].direction)
	}
	if secondOuterWidth+thirdOuterWidth > viewportWidth {
		t.Fatalf("test setup invalid: second+third widths = %d exceed viewport width = %d", secondOuterWidth+thirdOuterWidth, viewportWidth)
	}
	if surfaces[3].x <= surfaces[2].x {
		t.Fatalf("third submenu x = %d, want > second submenu x = %d after mirrored right fallback", surfaces[3].x, surfaces[2].x)
	}
	if surfaces[3].direction != overlayposition.CascadeRight {
		t.Fatalf("third submenu direction = %v, want overlayposition.CascadeRight after mirrored fallback", surfaces[3].direction)
	}

	boundsX, _, boundsWidth, _ := inst.GetBounds()
	expectedRight := inst.layoutBounds[0] + max(
		rootOuterWidth,
		max(surfaces[1].x+firstOuterWidth, max(surfaces[2].x+secondOuterWidth, surfaces[3].x+thirdOuterWidth)),
	)
	if boundsX != 0 {
		t.Fatalf("hit bounds x = %d, want 0 after second submenu clamps to left edge", boundsX)
	}
	if boundsX+boundsWidth != expectedRight {
		t.Fatalf("hit bounds right edge = %d, want %d to include mirrored-right third submenu", boundsX+boundsWidth, expectedRight)
	}
}

func TestPopupInstanceNestedSubmenusMirrorLeftAfterRightEdgeClampDirection(t *testing.T) {
	menuRegistryGlobal.reset()
	defer menuRegistryGlobal.reset()

	model := NewPopup([]MenuItem{
		Action("new", "New", testIntent{"new"}),
		Action("open", "Open", testIntent{"open"}),
		Action("save", "Save", testIntent{"save"}),
		Action("export", "Export", testIntent{"export"}),
		Submenu("ops", "Operations",
			Action("wide", "Extremely Wide Recovery Workspace Ledger", testIntent{"wide"}),
			Action("repair", "Repair", testIntent{"repair"}),
			Action("archive", "Archive", testIntent{"archive"}),
			Submenu("branch", "Branch",
				Action("node", "Moderately Wide Inner Recovery Journal", testIntent{"node"}),
				Submenu("pivot", "Pivot",
					Action("deep", "Deep Action", testIntent{"deep"}),
				),
			),
		),
	}).ActivePath(4, 3, 1, 0).BuildModel()
	inst := newPopupVNode(clearPortalModel(model)).CreateInstance().(*popupInstance)

	rootMetrics := inst.popupMetricsFor(inst.model.Items)
	firstMetrics := inst.popupMetricsFor(inst.model.Items[4].Normalize().Children)
	secondMetrics := inst.popupMetricsFor(inst.model.Items[4].Normalize().Children[3].Normalize().Children)
	thirdMetrics := inst.popupMetricsFor(inst.model.Items[4].Normalize().Children[3].Normalize().Children[1].Normalize().Children)
	rootOuterWidth := popupSurfaceOuterWidth(rootMetrics)
	rootOuterHeight := popupSurfaceOuterHeight(rootMetrics)
	firstOuterWidth := popupSurfaceOuterWidth(firstMetrics)
	secondOuterWidth := popupSurfaceOuterWidth(secondMetrics)
	thirdOuterWidth := popupSurfaceOuterWidth(thirdMetrics)
	viewportWidth := rootOuterWidth + firstOuterWidth + 2
	viewportHeight := 18

	inst.SetViewportSize(viewportWidth, viewportHeight)
	inst.SetBounds(0, 2, rootOuterWidth, rootOuterHeight)

	surfaces := inst.popupSurfaces()
	if len(surfaces) != 4 {
		t.Fatalf("popupSurfaces len = %d, want 4", len(surfaces))
	}
	if surfaces[1].x != rootOuterWidth {
		t.Fatalf("first submenu x = %d, want %d for right cascade", surfaces[1].x, rootOuterWidth)
	}
	if surfaces[1].direction != overlayposition.CascadeRight {
		t.Fatalf("first submenu direction = %v, want overlayposition.CascadeRight", surfaces[1].direction)
	}
	expectedSecondX := viewportWidth - secondOuterWidth
	if surfaces[2].x != expectedSecondX {
		t.Fatalf("second submenu x = %d, want %d after right-edge clamp", surfaces[2].x, expectedSecondX)
	}
	if surfaces[2].direction != overlayposition.CascadeRight {
		t.Fatalf("second submenu direction = %v, want overlayposition.CascadeRight after resolved clamp", surfaces[2].direction)
	}
	if thirdOuterWidth > surfaces[2].x {
		t.Fatalf("test setup invalid: third width = %d should fit to the left of second submenu x = %d", thirdOuterWidth, surfaces[2].x)
	}
	if surfaces[3].x >= surfaces[2].x {
		t.Fatalf("third submenu x = %d, want < second submenu x = %d after mirrored left fallback", surfaces[3].x, surfaces[2].x)
	}
	if surfaces[3].direction != overlayposition.CascadeLeft {
		t.Fatalf("third submenu direction = %v, want overlayposition.CascadeLeft after mirrored fallback", surfaces[3].direction)
	}

	boundsX, _, boundsWidth, _ := inst.GetBounds()
	expectedRight := max(
		rootOuterWidth,
		max(surfaces[1].x+firstOuterWidth, max(surfaces[2].x+secondOuterWidth, surfaces[3].x+thirdOuterWidth)),
	)
	if boundsX != 0 {
		t.Fatalf("hit bounds x = %d, want 0 with root anchored at left edge", boundsX)
	}
	if boundsX+boundsWidth != expectedRight {
		t.Fatalf("hit bounds right edge = %d, want %d to include right-edge clamped second submenu", boundsX+boundsWidth, expectedRight)
	}
}

func TestPopupInstanceNestedSubmenusMirrorLeftAndClampUpwardNearBottomRight(t *testing.T) {
	menuRegistryGlobal.reset()
	defer menuRegistryGlobal.reset()

	model := NewPopup([]MenuItem{
		Action("new", "New", testIntent{"new"}),
		Action("open", "Open", testIntent{"open"}),
		Action("save", "Save", testIntent{"save"}),
		Action("export", "Export", testIntent{"export"}),
		Submenu("ops", "Operations",
			Action("scan", "Scan", testIntent{"scan"}),
			Action("repair", "Repair", testIntent{"repair"}),
			Action("archive", "Archive", testIntent{"archive"}),
			Action("cleanup", "Cleanup", testIntent{"cleanup"}),
			Submenu("branch", "Branch",
				Action("wide", "Extremely Wide Recovery Workspace Ledger", testIntent{"wide"}),
				Action("repair-node", "Repair Recovery Nodes", testIntent{"repair-node"}),
				Action("audit", "Audit", testIntent{"audit"}),
				Action("history", "Historical Recovery Timeline Browser", testIntent{"history"}),
				Submenu("pivot", "Pivot",
					Action("deep", "Deep Action", testIntent{"deep"}),
					Action("verify", "Verify", testIntent{"verify"}),
					Action("recover", "Recover", testIntent{"recover"}),
				),
			),
		),
	}).ActivePath(4, 4, 4, 0).BuildModel()
	inst := newPopupVNode(clearPortalModel(model)).CreateInstance().(*popupInstance)

	rootMetrics := inst.popupMetricsFor(inst.model.Items)
	firstMetrics := inst.popupMetricsFor(inst.model.Items[4].Normalize().Children)
	secondMetrics := inst.popupMetricsFor(inst.model.Items[4].Normalize().Children[4].Normalize().Children)
	thirdMetrics := inst.popupMetricsFor(inst.model.Items[4].Normalize().Children[4].Normalize().Children[4].Normalize().Children)
	rootOuterWidth := popupSurfaceOuterWidth(rootMetrics)
	rootOuterHeight := popupSurfaceOuterHeight(rootMetrics)
	firstOuterWidth := popupSurfaceOuterWidth(firstMetrics)
	secondOuterWidth := popupSurfaceOuterWidth(secondMetrics)
	secondOuterHeight := popupSurfaceOuterHeight(secondMetrics)
	thirdOuterWidth := popupSurfaceOuterWidth(thirdMetrics)
	thirdOuterHeight := popupSurfaceOuterHeight(thirdMetrics)
	viewportWidth := max(rootOuterWidth+firstOuterWidth+2, secondOuterWidth+thirdOuterWidth)
	viewportHeight := max(rootOuterHeight, thirdOuterHeight) + 2

	inst.SetViewportSize(viewportWidth, viewportHeight)
	inst.SetBounds(0, viewportHeight-rootOuterHeight, rootOuterWidth, rootOuterHeight)

	surfaces := inst.popupSurfaces()
	if len(surfaces) != 4 {
		t.Fatalf("popupSurfaces len = %d, want 4", len(surfaces))
	}
	if surfaces[1].x != rootOuterWidth {
		t.Fatalf("first submenu x = %d, want %d for right cascade", surfaces[1].x, rootOuterWidth)
	}
	if surfaces[1].direction != overlayposition.CascadeRight {
		t.Fatalf("first submenu direction = %v, want overlayposition.CascadeRight", surfaces[1].direction)
	}
	expectedSecondX := viewportWidth - secondOuterWidth
	if surfaces[2].x != expectedSecondX {
		t.Fatalf("second submenu x = %d, want %d after right-edge clamp", surfaces[2].x, expectedSecondX)
	}
	if surfaces[2].direction != overlayposition.CascadeRight {
		t.Fatalf("second submenu direction = %v, want overlayposition.CascadeRight after resolved clamp", surfaces[2].direction)
	}
	if thirdOuterWidth > surfaces[2].x {
		t.Fatalf("test setup invalid: third width = %d should fit to the left of second submenu x = %d", thirdOuterWidth, surfaces[2].x)
	}
	if surfaces[3].x >= surfaces[2].x {
		t.Fatalf("third submenu x = %d, want < second submenu x = %d after mirrored left fallback", surfaces[3].x, surfaces[2].x)
	}
	if surfaces[3].direction != overlayposition.CascadeLeft {
		t.Fatalf("third submenu direction = %v, want overlayposition.CascadeLeft after mirrored fallback", surfaces[3].direction)
	}

	baseY := inst.layoutBounds[1]
	secondTop := baseY + surfaces[2].y
	thirdTop := baseY + surfaces[3].y
	if secondTop+secondOuterHeight > viewportHeight {
		t.Fatalf("second submenu bottom edge = %d, want <= %d", secondTop+secondOuterHeight, viewportHeight)
	}
	if thirdTop != viewportHeight-thirdOuterHeight {
		t.Fatalf("third submenu top = %d, want %d after bottom-edge clamp", thirdTop, viewportHeight-thirdOuterHeight)
	}
	if thirdTop >= secondTop+4 {
		t.Fatalf("third submenu top = %d, want < %d after upward clamp from pivot row geometry", thirdTop, secondTop+4)
	}

	boundsX, boundsY, boundsWidth, boundsHeight := inst.GetBounds()
	expectedRight := max(
		rootOuterWidth,
		max(surfaces[1].x+firstOuterWidth, max(surfaces[2].x+secondOuterWidth, surfaces[3].x+thirdOuterWidth)),
	)
	if boundsX != 0 {
		t.Fatalf("hit bounds x = %d, want 0 with root anchored at left edge", boundsX)
	}
	if boundsY > thirdTop {
		t.Fatalf("hit bounds y = %d, want <= third submenu top = %d", boundsY, thirdTop)
	}
	if boundsX+boundsWidth != expectedRight {
		t.Fatalf("hit bounds right edge = %d, want %d to include mirrored-left third submenu", boundsX+boundsWidth, expectedRight)
	}
	if boundsY+boundsHeight > viewportHeight {
		t.Fatalf("hit bounds bottom edge = %d, want <= %d", boundsY+boundsHeight, viewportHeight)
	}
}

func TestPopupInstanceNestedSubmenusMirrorRightAndClampUpwardNearBottomLeft(t *testing.T) {
	menuRegistryGlobal.reset()
	defer menuRegistryGlobal.reset()

	model := NewPopup([]MenuItem{
		Action("new", "New", testIntent{"new"}),
		Action("open", "Open", testIntent{"open"}),
		Action("save", "Save", testIntent{"save"}),
		Action("export", "Export", testIntent{"export"}),
		Submenu("ops", "Operations",
			Action("scan", "Scan", testIntent{"scan"}),
			Action("repair", "Repair", testIntent{"repair"}),
			Action("archive", "Archive", testIntent{"archive"}),
			Action("cleanup", "Cleanup", testIntent{"cleanup"}),
			Submenu("branch", "Branch",
				Action("wide", "Extremely Wide Recovery Workspace Ledger", testIntent{"wide"}),
				Action("repair-node", "Repair Recovery Nodes", testIntent{"repair-node"}),
				Action("audit", "Audit", testIntent{"audit"}),
				Action("history", "Historical Recovery Timeline Browser", testIntent{"history"}),
				Submenu("pivot", "Pivot",
					Action("deep", "Deep Action", testIntent{"deep"}),
					Action("verify", "Verify", testIntent{"verify"}),
					Action("recover", "Recover", testIntent{"recover"}),
				),
			),
		),
	}).ActivePath(4, 4, 4, 0).BuildModel()
	inst := newPopupVNode(clearPortalModel(model)).CreateInstance().(*popupInstance)

	rootMetrics := inst.popupMetricsFor(inst.model.Items)
	firstMetrics := inst.popupMetricsFor(inst.model.Items[4].Normalize().Children)
	secondMetrics := inst.popupMetricsFor(inst.model.Items[4].Normalize().Children[4].Normalize().Children)
	thirdMetrics := inst.popupMetricsFor(inst.model.Items[4].Normalize().Children[4].Normalize().Children[4].Normalize().Children)
	rootOuterWidth := popupSurfaceOuterWidth(rootMetrics)
	rootOuterHeight := popupSurfaceOuterHeight(rootMetrics)
	firstOuterWidth := popupSurfaceOuterWidth(firstMetrics)
	secondOuterWidth := popupSurfaceOuterWidth(secondMetrics)
	secondOuterHeight := popupSurfaceOuterHeight(secondMetrics)
	thirdOuterWidth := popupSurfaceOuterWidth(thirdMetrics)
	thirdOuterHeight := popupSurfaceOuterHeight(thirdMetrics)
	viewportWidth := max(rootOuterWidth+firstOuterWidth+2, secondOuterWidth+thirdOuterWidth)
	viewportHeight := max(rootOuterHeight, thirdOuterHeight) + 2

	inst.SetViewportSize(viewportWidth, viewportHeight)
	inst.SetBounds(viewportWidth-rootOuterWidth, viewportHeight-rootOuterHeight, rootOuterWidth, rootOuterHeight)

	surfaces := inst.popupSurfaces()
	if len(surfaces) != 4 {
		t.Fatalf("popupSurfaces len = %d, want 4", len(surfaces))
	}
	if surfaces[1].direction != overlayposition.CascadeLeft {
		t.Fatalf("first submenu direction = %v, want overlayposition.CascadeLeft", surfaces[1].direction)
	}
	if surfaces[2].x != -inst.layoutBounds[0] {
		t.Fatalf("second submenu x = %d, want %d after left-edge clamp", surfaces[2].x, -inst.layoutBounds[0])
	}
	if surfaces[2].direction != overlayposition.CascadeLeft {
		t.Fatalf("second submenu direction = %v, want overlayposition.CascadeLeft after resolved clamp", surfaces[2].direction)
	}
	if secondOuterWidth+thirdOuterWidth > viewportWidth {
		t.Fatalf("test setup invalid: second+third widths = %d exceed viewport width = %d", secondOuterWidth+thirdOuterWidth, viewportWidth)
	}
	if surfaces[3].x <= surfaces[2].x {
		t.Fatalf("third submenu x = %d, want > second submenu x = %d after mirrored right fallback", surfaces[3].x, surfaces[2].x)
	}
	if surfaces[3].direction != overlayposition.CascadeRight {
		t.Fatalf("third submenu direction = %v, want overlayposition.CascadeRight after mirrored fallback", surfaces[3].direction)
	}

	baseY := inst.layoutBounds[1]
	secondTop := baseY + surfaces[2].y
	thirdTop := baseY + surfaces[3].y
	if secondTop+secondOuterHeight > viewportHeight {
		t.Fatalf("second submenu bottom edge = %d, want <= %d", secondTop+secondOuterHeight, viewportHeight)
	}
	if thirdTop != viewportHeight-thirdOuterHeight {
		t.Fatalf("third submenu top = %d, want %d after bottom-edge clamp", thirdTop, viewportHeight-thirdOuterHeight)
	}
	if thirdTop >= secondTop+4 {
		t.Fatalf("third submenu top = %d, want < %d after upward clamp from pivot row geometry", thirdTop, secondTop+4)
	}

	boundsX, boundsY, boundsWidth, boundsHeight := inst.GetBounds()
	expectedRight := inst.layoutBounds[0] + max(
		rootOuterWidth,
		max(surfaces[1].x+firstOuterWidth, max(surfaces[2].x+secondOuterWidth, surfaces[3].x+thirdOuterWidth)),
	)
	if boundsX != 0 {
		t.Fatalf("hit bounds x = %d, want 0 after second submenu clamps to left edge", boundsX)
	}
	if boundsY > thirdTop {
		t.Fatalf("hit bounds y = %d, want <= third submenu top = %d", boundsY, thirdTop)
	}
	if boundsX+boundsWidth != expectedRight {
		t.Fatalf("hit bounds right edge = %d, want %d to include mirrored-right third submenu", boundsX+boundsWidth, expectedRight)
	}
	if boundsY+boundsHeight > viewportHeight {
		t.Fatalf("hit bounds bottom edge = %d, want <= %d", boundsY+boundsHeight, viewportHeight)
	}
}

func TestPopupInstanceNestedSubmenusZigZagMirrorAndClampUpwardNearBottom(t *testing.T) {
	menuRegistryGlobal.reset()
	defer menuRegistryGlobal.reset()

	model := NewPopup([]MenuItem{
		Action("new", "New", testIntent{"new"}),
		Action("open", "Open", testIntent{"open"}),
		Action("save", "Save", testIntent{"save"}),
		Action("export", "Export", testIntent{"export"}),
		Submenu("ops", "Operations",
			Action("scan", "Scan", testIntent{"scan"}),
			Action("repair", "Repair", testIntent{"repair"}),
			Action("archive", "Archive", testIntent{"archive"}),
			Action("cleanup", "Cleanup", testIntent{"cleanup"}),
			Submenu("branch", "Branch",
				Action("wide", "Extremely Wide Recovery Workspace Ledger", testIntent{"wide"}),
				Action("repair-node", "Repair Recovery Nodes", testIntent{"repair-node"}),
				Action("audit", "Audit", testIntent{"audit"}),
				Action("history", "Historical Recovery Timeline Browser", testIntent{"history"}),
				Submenu("pivot", "Pivot",
					Action("verify", "Verify", testIntent{"verify"}),
					Action("compare", "Compare", testIntent{"compare"}),
					Action("recover", "Recover", testIntent{"recover"}),
					Submenu("rebound", "Rebound",
						Action("deep", "Deep Action", testIntent{"deep"}),
						Action("secondary", "Secondary Recovery Marker", testIntent{"secondary"}),
						Action("rollback", "Rollback", testIntent{"rollback"}),
						Action("archive-tail", "Archive Tail", testIntent{"archive-tail"}),
					),
				),
			),
		),
	}).ActivePath(4, 4, 4, 3, 0).BuildModel()
	inst := newPopupVNode(clearPortalModel(model)).CreateInstance().(*popupInstance)

	rootMetrics := inst.popupMetricsFor(inst.model.Items)
	firstMetrics := inst.popupMetricsFor(inst.model.Items[4].Normalize().Children)
	secondMetrics := inst.popupMetricsFor(inst.model.Items[4].Normalize().Children[4].Normalize().Children)
	thirdMetrics := inst.popupMetricsFor(inst.model.Items[4].Normalize().Children[4].Normalize().Children[4].Normalize().Children)
	fourthMetrics := inst.popupMetricsFor(inst.model.Items[4].Normalize().Children[4].Normalize().Children[4].Normalize().Children[3].Normalize().Children)
	rootOuterWidth := popupSurfaceOuterWidth(rootMetrics)
	rootOuterHeight := popupSurfaceOuterHeight(rootMetrics)
	firstOuterWidth := popupSurfaceOuterWidth(firstMetrics)
	secondOuterWidth := popupSurfaceOuterWidth(secondMetrics)
	thirdOuterWidth := popupSurfaceOuterWidth(thirdMetrics)
	fourthOuterWidth := popupSurfaceOuterWidth(fourthMetrics)
	fourthOuterHeight := popupSurfaceOuterHeight(fourthMetrics)
	viewportWidth := max(rootOuterWidth+firstOuterWidth+2, secondOuterWidth+thirdOuterWidth)
	viewportHeight := max(rootOuterHeight, fourthOuterHeight) + 2

	inst.SetViewportSize(viewportWidth, viewportHeight)
	inst.SetBounds(0, viewportHeight-rootOuterHeight, rootOuterWidth, rootOuterHeight)

	surfaces := inst.popupSurfaces()
	if len(surfaces) != 5 {
		t.Fatalf("popupSurfaces len = %d, want 5", len(surfaces))
	}
	if surfaces[1].direction != overlayposition.CascadeRight {
		t.Fatalf("first submenu direction = %v, want overlayposition.CascadeRight", surfaces[1].direction)
	}
	expectedSecondX := viewportWidth - secondOuterWidth
	if surfaces[2].x != expectedSecondX {
		t.Fatalf("second submenu x = %d, want %d after right-edge clamp", surfaces[2].x, expectedSecondX)
	}
	if surfaces[2].direction != overlayposition.CascadeRight {
		t.Fatalf("second submenu direction = %v, want overlayposition.CascadeRight after resolved clamp", surfaces[2].direction)
	}
	expectedThirdX := max(0, surfaces[2].x-thirdOuterWidth)
	if surfaces[3].x != expectedThirdX {
		t.Fatalf("third submenu x = %d, want %d after mirrored left fallback", surfaces[3].x, expectedThirdX)
	}
	if surfaces[3].direction != overlayposition.CascadeLeft {
		t.Fatalf("third submenu direction = %v, want overlayposition.CascadeLeft after mirrored fallback", surfaces[3].direction)
	}
	if surfaces[4].x <= surfaces[3].x {
		t.Fatalf("fourth submenu x = %d, want > third submenu x = %d after mirrored right fallback", surfaces[4].x, surfaces[3].x)
	}
	if surfaces[4].direction != overlayposition.CascadeRight {
		t.Fatalf("fourth submenu direction = %v, want overlayposition.CascadeRight after rebound fallback", surfaces[4].direction)
	}
	if surfaces[3].x != 0 {
		t.Fatalf("third submenu x = %d, want 0 to prove mirrored-left branch hit the viewport edge", surfaces[3].x)
	}
	if fourthOuterWidth > secondOuterWidth {
		t.Fatalf("test setup invalid: fourth width = %d should fit to the right of third within second width = %d", fourthOuterWidth, secondOuterWidth)
	}

	baseY := inst.layoutBounds[1]
	thirdTop := baseY + surfaces[3].y
	fourthTop := baseY + surfaces[4].y
	if fourthTop != viewportHeight-fourthOuterHeight {
		t.Fatalf("fourth submenu top = %d, want %d after bottom-edge clamp", fourthTop, viewportHeight-fourthOuterHeight)
	}
	if fourthTop >= thirdTop+3 {
		t.Fatalf("fourth submenu top = %d, want < %d after upward clamp from rebound row geometry", fourthTop, thirdTop+3)
	}

	boundsX, boundsY, boundsWidth, boundsHeight := inst.GetBounds()
	expectedRight := max(
		rootOuterWidth,
		max(surfaces[1].x+firstOuterWidth,
			max(surfaces[2].x+secondOuterWidth,
				max(surfaces[3].x+thirdOuterWidth, surfaces[4].x+fourthOuterWidth))),
	)
	if boundsX != 0 {
		t.Fatalf("hit bounds x = %d, want 0 after mirrored-left branch reaches viewport edge", boundsX)
	}
	if boundsY > fourthTop {
		t.Fatalf("hit bounds y = %d, want <= fourth submenu top = %d", boundsY, fourthTop)
	}
	if boundsX+boundsWidth != expectedRight {
		t.Fatalf("hit bounds right edge = %d, want %d to include mirrored-right fourth submenu", boundsX+boundsWidth, expectedRight)
	}
	if boundsY+boundsHeight > viewportHeight {
		t.Fatalf("hit bounds bottom edge = %d, want <= %d", boundsY+boundsHeight, viewportHeight)
	}
}

func TestPopupInstancePathPrefixIsIncludedInEmittedPath(t *testing.T) {
	menuRegistryGlobal.reset()
	defer menuRegistryGlobal.reset()

	model := NewPopup([]MenuItem{
		Submenu("theme", "Theme",
			Action("dark", "Dark", testIntent{"dark"}),
			Action("light", "Light", testIntent{"light"}),
		),
		Action("refresh", "Refresh", testIntent{"refresh"}),
	}).PathPrefix(1).BuildModel()

	inst := newPopupVNode(clearPortalModel(model)).CreateInstance().(*popupInstance)
	var emitted []intent.Intent
	inst.SetIntentEmitter(func(i intent.Intent) { emitted = append(emitted, i) })

	if handled := inst.HandleAction(action.NewAction(action.ActionNavigateRight)); !handled {
		t.Fatal("navigate right should be handled")
	}
	if len(emitted) == 0 {
		t.Fatal("expected OpenMenuIntent to be emitted")
	}

	openIntent, ok := emitted[len(emitted)-1].(OpenMenuIntent)
	if !ok {
		t.Fatalf("last emitted intent = %T, want OpenMenuIntent", emitted[len(emitted)-1])
	}
	if !PathEqual(openIntent.Path, []int{1, 0}) {
		t.Fatalf("OpenMenuIntent.Path = %v, want [1 0]", openIntent.Path)
	}
	if got := inst.GetProps()["model"].(Model).ActivePath; !PathEqual(got, []int{1, 0, 0}) {
		t.Fatalf("GetProps().model.ActivePath = %v, want [1 0 0]", got)
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
	vnode := buildPopupSurface([]MenuItem{Action("open", "Open", testIntent{"open"})})
	inst := vnode.CreateInstance().(*popupInstance)
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

func TestMenuMiddlewareClickOutsideClosesOpenPopupWithValuePayload(t *testing.T) {
	menuRegistryGlobal.reset()
	defer menuRegistryGlobal.reset()
	vnode := buildPopupSurface([]MenuItem{Action("open", "Open", testIntent{"open"})})
	inst := vnode.CreateInstance().(*popupInstance)
	inst.SetBounds(10, 5, 24, 8)
	inst.OnMount()
	defer inst.Destroy()

	middleware := NewMiddleware()
	mouse := *runtimemsg.NewMouseMsg(1, 1, runtimemsg.MouseLeft, runtimemsg.MouseActionPress)
	act := action.NewAction(action.ActionClick).WithPayload(mouse)
	if next := middleware.Before(act); next != nil {
		t.Fatal("outside click should be intercepted when popup closes")
	}
	if inst.open {
		t.Fatal("popup should be closed after outside click")
	}
}

func TestMenuMiddlewareOutsideClickIgnoresTargetFiberAncestry(t *testing.T) {
	menuRegistryGlobal.reset()
	defer menuRegistryGlobal.reset()

	vnode := buildPopupSurface([]MenuItem{Action("open", "Open", testIntent{"open"})})
	inst := vnode.CreateInstance().(*popupInstance)
	inst.SetBounds(10, 5, 24, 8)
	inst.OnMount()
	defer inst.Destroy()

	// Simulate a stale/incorrect TargetFiber chain that points at menu instances.
	// Outside-click handling should rely on geometry and still close the popup.
	target := &rtui.Fiber{Instance: inst}
	msg := runtimemsg.NewMouseMsg(1, 1, runtimemsg.MouseLeft, runtimemsg.MouseActionPress)
	msg.TargetFiber = target

	middleware := NewMiddleware()
	act := action.NewAction(action.ActionClick).WithPayload(msg)
	if next := middleware.Before(act); next != nil {
		t.Fatal("outside click should be intercepted when popup closes")
	}
	if inst.open {
		t.Fatal("popup should be closed even when TargetFiber points to menu")
	}
}

func TestMenuMiddlewareLeavesInsideClickAlone(t *testing.T) {
	menuRegistryGlobal.reset()
	defer menuRegistryGlobal.reset()
	vnode := buildPopupSurface([]MenuItem{Action("open", "Open", testIntent{"open"})})
	inst := vnode.CreateInstance().(*popupInstance)
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

func TestMenuMiddlewareLeavesMenuBarClickAlone(t *testing.T) {
	menuRegistryGlobal.reset()
	defer menuRegistryGlobal.reset()

	barModel := NewMenuBar([]MenuItem{
		Submenu("file", "File", Action("open", "Open", testIntent{"open"})),
	}).ComponentID("main-menu").BuildModel()
	bar := newBarVNode(barModel).CreateInstance().(*barInstance)
	bar.SetBounds(0, 0, 20, 1)
	bar.OnMount()
	defer bar.Destroy()

	popupModel := NewPopup([]MenuItem{
		Action("open", "Open", testIntent{"open"}),
	}).ComponentID("main-menu").BuildModel()
	inst := newPopupVNode(clearPortalModel(popupModel)).CreateInstance().(*popupInstance)
	inst.SetBounds(0, 1, 24, 8)
	inst.OnMount()
	defer inst.Destroy()

	middleware := NewMiddleware()
	act := action.NewAction(action.ActionClick).WithPayload(runtimemsg.NewMouseMsg(5, 0, runtimemsg.MouseLeft, runtimemsg.MouseActionPress))
	if next := middleware.Before(act); next == nil {
		t.Fatal("menu bar click should continue dispatch")
	}
	if !inst.open {
		t.Fatal("popup should remain open after clicking menu bar")
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

func TestInstallStateIsScopedToHostInstance(t *testing.T) {
	first := &fakeInstallerHost{}
	second := &fakeInstallerHost{}

	Install(first, nil)
	Install(second, nil)

	if first.middlewareCount != 1 {
		t.Fatalf("first middlewareCount = %d, want 1", first.middlewareCount)
	}
	if second.middlewareCount != 1 {
		t.Fatalf("second middlewareCount = %d, want 1", second.middlewareCount)
	}
}

func TestRegisterBuiltinHandlers_RegistersNavigateMenuIntentAsOverridable(t *testing.T) {
	registry := intent.NewRegistry()
	registerBuiltinHandlers(registry)

	handler, ok := registry.GetHandler((NavigateMenuIntent{}).IntentType())
	if !ok || handler == nil {
		t.Fatal("expected builtin handler for menu.navigate")
	}

	overrideCalled := false
	intent.RegisterTypedWithOpts(registry, func(ctx *intent.ActionContext, i NavigateMenuIntent) intent.IntentResult {
		overrideCalled = true
		return intent.HandledResult()
	})

	result := handler.Handle(intent.NewActionContext(nil, "test", nil), NavigateMenuIntent{})
	if result.Error != nil {
		t.Fatalf("builtin handler returned error: %v", result.Error)
	}

	overridden, ok := registry.GetHandler((NavigateMenuIntent{}).IntentType())
	if !ok || overridden == nil {
		t.Fatal("expected overridden handler for menu.navigate")
	}

	overridden.Handle(intent.NewActionContext(nil, "test", nil), NavigateMenuIntent{})
	if !overrideCalled {
		t.Fatal("expected custom menu.navigate handler to override builtin handler")
	}
}
