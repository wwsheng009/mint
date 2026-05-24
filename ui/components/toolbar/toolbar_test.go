package toolbar

import (
	"testing"

	rttypes "github.com/wwsheng009/mint/runtime/types"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/button"
	menucomp "github.com/wwsheng009/mint/ui/components/menu"
)

type testIntent struct {
	name string
}

func (i testIntent) IntentType() string { return i.name }

func TestBuilderAndProps(t *testing.T) {
	bar := NewBuilder().
		Key("ops-toolbar").
		Title("Load Balancer").
		TitleWidth(16).
		Width(96).
		Gap(2).
		Dense(true).
		Left(Text("scope", "group: default").WithWidth(18)).
		Center(Badge("state", "degraded").WithColors("black", "yellow")).
		Action(Button("refresh", "Refresh", testIntent{"refresh"}).Primary().WithHelp("Reload current page")).
		Action(Button("reset", "Reset", testIntent{"reset"}).Danger().WithDisabled(true)).
		BuildVNode()

	if bar.Key() != "ops-toolbar" {
		t.Fatalf("key = %q, want ops-toolbar", bar.Key())
	}
	props := bar.Props()
	if got := props[propTitle]; got != "Load Balancer" {
		t.Fatalf("title = %v, want Load Balancer", got)
	}
	if got := props[propTitleWidth]; got != 16 {
		t.Fatalf("titleWidth = %v, want 16", got)
	}
	if got := props[propWidth]; got != 96 {
		t.Fatalf("width = %v, want 96", got)
	}
	if got := props[propDense]; got != true {
		t.Fatalf("dense = %v, want true", got)
	}
	left := props[propLeftItems].([]Item)
	if len(left) != 1 || left[0].Key != "scope" || left[0].Width != 18 {
		t.Fatalf("left items = %#v", left)
	}
	center := props[propCenterItems].([]Item)
	if len(center) != 1 || center[0].Kind != ItemBadge {
		t.Fatalf("center items = %#v", center)
	}
	right := props[propRightItems].([]Item)
	if len(right) != 2 || right[0].Variant != button.VariantPrimary || right[1].Variant != button.VariantDanger {
		t.Fatalf("right items = %#v", right)
	}
}

func TestRuntimeChildrenBuildToolbarControls(t *testing.T) {
	inst := NewBuilder().
		Key("ops-toolbar").
		Title("Runtime").
		Left(Text("focus", "F2 Load Balancer")).
		Right(Button("refresh", "Refresh", testIntent{"refresh"}).Primary()).
		Right(Button("reload", "Reload", testIntent{"reload"}).Danger().WithDisabled(true)).
		BuildInstance()

	children := inst.RuntimeChildren()
	if len(children) != 1 {
		t.Fatalf("RuntimeChildren len = %d, want 1", len(children))
	}
	root := children[0]
	if root.Tag() != "hstack" {
		t.Fatalf("root tag = %q, want hstack", root.Tag())
	}
	if findVNodeByKey(root, "ops-toolbar-title") == nil {
		t.Fatal("title node not found")
	}
	refresh := findVNodeByKey(root, "ops-toolbar-right-refresh")
	if refresh == nil {
		t.Fatal("refresh action not found")
	}
	if got := refresh.Props()["variant"]; got != button.VariantPrimary {
		t.Fatalf("refresh variant = %v, want primary", got)
	}
	reload := findVNodeByKey(root, "ops-toolbar-right-reload")
	if reload == nil {
		t.Fatal("reload action not found")
	}
	if got := reload.Props()["disabled"]; got != true {
		t.Fatalf("reload disabled = %v, want true", got)
	}
}

func TestRuntimeChildrenBuildStatusBarMode(t *testing.T) {
	inst := NewBuilder().
		Key("ops-status").
		Title("Manager").
		UseStatusBar(true).
		Left(Badge("mode", "ADMIN").WithHelp("Admin mode")).
		Right(Button("help", "F10 Help", testIntent{"help"}).WithHelp("Open help")).
		BuildInstance()

	children := inst.RuntimeChildren()
	if len(children) != 1 {
		t.Fatalf("RuntimeChildren len = %d, want 1", len(children))
	}
	root := children[0]
	if root.Tag() != "fragment" && root.Tag() != "hstack" && root.Tag() != "vstack" {
		t.Fatalf("unexpected statusbar root tag = %q", root.Tag())
	}
	if findVNodeByKey(root, "ops-status-left-mode") == nil {
		t.Fatal("statusbar badge section not found")
	}
	if findVNodeByKey(root, "ops-status-right-help") == nil {
		t.Fatal("statusbar help action not found")
	}
}

func TestRuntimeChildrenBuildControlledDropdownMenu(t *testing.T) {
	items := []menucomp.MenuItem{
		menucomp.Action("reload", "Reload", testIntent{"reload"}),
		menucomp.Action("inspect", "Inspect", testIntent{"inspect"}).WithDescription("Open diagnostics"),
	}
	inst := NewBuilder().
		Key("ops-toolbar").
		Right(Dropdown("more", "More", items, true).
			WithMenuID("ops-actions").
			WithMenuPlacement(menucomp.PlacementBottomEnd).
			WithMenuMinWidth(24).
			WithMenuMaxHeight(8).
			WithMenuDescriptions(true)).
		BuildInstance()

	children := inst.RuntimeChildren()
	if len(children) != 1 {
		t.Fatalf("RuntimeChildren len = %d, want 1", len(children))
	}
	root := children[0]
	if root.Tag() != "fragment" {
		t.Fatalf("root tag = %q, want fragment when dropdown is open", root.Tag())
	}
	buttonNode := findVNodeByKey(root, "ops-toolbar-right-more")
	if buttonNode == nil {
		t.Fatal("dropdown button not found")
	}
	if buttonNode.ID() != "ops-toolbar-right-more" {
		t.Fatalf("dropdown button ID = %q, want anchorable ID", buttonNode.ID())
	}
	pressIntent, ok := buttonNode.Props()["pressIntent"].(menucomp.OpenMenuIntent)
	if !ok {
		t.Fatalf("press intent = %T, want menu.OpenMenuIntent", buttonNode.Props()["pressIntent"])
	}
	if pressIntent.MenuID != "ops-actions" {
		t.Fatalf("OpenMenuIntent.MenuID = %q, want ops-actions", pressIntent.MenuID)
	}

	portal := findVNodeByKey(root, "ops-actions-portal")
	if portal == nil {
		t.Fatal("dropdown menu portal not found")
	}
	if got, _ := portal.Props()["anchorId"].(string); got != "ops-toolbar-right-more" {
		t.Fatalf("portal anchorId = %q, want toolbar button ID", got)
	}
	if got, _ := portal.Props()["anchor"].(rttypes.Anchor); got != rttypes.AnchorBottomRight {
		t.Fatalf("portal anchor = %v, want AnchorBottomRight", got)
	}
	if got, _ := portal.Props()["popupPlacement"].(string); got != string(menucomp.PlacementBottomEnd) {
		t.Fatalf("popupPlacement = %q, want bottom-end", got)
	}
}

func TestDropdownClosedOnlyRendersButton(t *testing.T) {
	inst := NewBuilder().
		Key("ops-toolbar").
		Right(Dropdown("more", "More", []menucomp.MenuItem{
			menucomp.Action("reload", "Reload", testIntent{"reload"}),
		}, false)).
		BuildInstance()

	children := inst.RuntimeChildren()
	if len(children) != 1 {
		t.Fatalf("RuntimeChildren len = %d, want 1", len(children))
	}
	root := children[0]
	if root.Tag() != "hstack" {
		t.Fatalf("root tag = %q, want hstack when dropdown is closed", root.Tag())
	}
	buttonNode := findVNodeByKey(root, "ops-toolbar-right-more")
	if buttonNode == nil {
		t.Fatal("dropdown button not found")
	}
	pressIntent, ok := buttonNode.Props()["pressIntent"].(menucomp.OpenMenuIntent)
	if !ok {
		t.Fatalf("press intent = %T, want menu.OpenMenuIntent", buttonNode.Props()["pressIntent"])
	}
	if pressIntent.MenuID != "ops-toolbar-right-more-menu" {
		t.Fatalf("OpenMenuIntent.MenuID = %q, want generated menu id", pressIntent.MenuID)
	}
}

func TestNormalizeItems(t *testing.T) {
	items := normalizeItems([]Item{
		{Key: "dup", Kind: ItemKind("bad"), Width: -1},
		{Key: "dup", Kind: ItemButton},
		{Kind: ItemSeparator},
	})
	if items[0].Key != "dup" || items[1].Key != "dup-1" || items[2].Key != "item-2" {
		t.Fatalf("item keys = %q, %q, %q", items[0].Key, items[1].Key, items[2].Key)
	}
	if items[0].Kind != ItemText {
		t.Fatalf("first item kind = %v, want text", items[0].Kind)
	}
	if items[0].Width != 0 {
		t.Fatalf("width = %d, want 0", items[0].Width)
	}
}

func findVNodeByKey(node rtui.VNode, key string) rtui.VNode {
	if node == nil {
		return nil
	}
	if node.Key() == key {
		return node
	}
	for _, child := range node.Children() {
		if found := findVNodeByKey(child, key); found != nil {
			return found
		}
	}
	return nil
}
