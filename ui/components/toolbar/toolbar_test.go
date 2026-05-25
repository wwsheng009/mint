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
		Action(Button("reset", "Reset", testIntent{"reset"}).Danger().WithDisabledReason("Select a provider before reset")).
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
	if !right[1].Disabled || right[1].DisabledReason != "Select a provider before reset" {
		t.Fatalf("reset disabled reason = disabled:%v reason:%q", right[1].Disabled, right[1].DisabledReason)
	}
}

func TestRuntimeChildrenBuildToolbarControls(t *testing.T) {
	inst := NewBuilder().
		Key("ops-toolbar").
		Title("Runtime").
		Left(Text("focus", "F2 Load Balancer")).
		Right(Button("refresh", "Refresh", testIntent{"refresh"}).Primary()).
		Right(Button("reload", "Reload", testIntent{"reload"}).Danger().WithDisabledReason("Runtime is already reloading")).
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
	tooltip := findVNodeByKey(root, "ops-toolbar-right-reload-tooltip")
	if tooltip == nil {
		t.Fatal("reload disabled reason tooltip not found")
	}
	if got := tooltip.Props()["text"]; got != "Runtime is already reloading" {
		t.Fatalf("tooltip text = %v, want disabled reason", got)
	}
}

func TestRuntimeChildrenBuildStatusBarMode(t *testing.T) {
	inst := NewBuilder().
		Key("ops-status").
		Title("Manager").
		UseStatusBar(true).
		Left(Badge("mode", "ADMIN").WithHelp("Admin mode")).
		Right(Button("reload", "Reload", testIntent{"reload"}).WithDisabledReason("Select a target first")).
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
	if findVNodeByKey(root, "ops-status-right-reload") == nil {
		t.Fatal("statusbar disabled action not found")
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

func TestOperationalPresets(t *testing.T) {
	tests := []struct {
		name string
		got  Item
		want Item
	}{
		{
			name: "key value",
			got:  KeyValue("endpoint", "endpoint", "http://localhost:8080"),
			want: Item{Key: "endpoint", Label: "endpoint: http://localhost:8080", Kind: ItemText},
		},
		{
			name: "key value empty",
			got:  KeyValue("scope", "scope", ""),
			want: Item{Key: "scope", Label: "scope: -", Kind: ItemText},
		},
		{
			name: "scope",
			got:  Scope("group: default"),
			want: Item{Key: "scope", Label: "scope: group: default", Kind: ItemText},
		},
		{
			name: "selection",
			got:  Selection("provider/openai"),
			want: Item{Key: "selection", Label: "selection: provider/openai", Kind: ItemText, FgColor: "bright-black"},
		},
	}

	for _, tt := range tests {
		if tt.got.Key != tt.want.Key || tt.got.Label != tt.want.Label || tt.got.Kind != tt.want.Kind || tt.got.FgColor != tt.want.FgColor {
			t.Fatalf("%s item = %+v, want %+v", tt.name, tt.got, tt.want)
		}
	}

	state := StateBadge("state", "degraded")
	if state.Kind != ItemBadge || state.Label != "degraded" || state.FgColor != "black" || state.BgColor != "yellow" {
		t.Fatalf("state badge = %+v", state)
	}
	busy := BusyBadge("busy", "")
	if busy.Label != "busy" || busy.BgColor != "yellow" {
		t.Fatalf("busy badge = %+v", busy)
	}
	err := ErrorBadge("error", "failed")
	if err.Label != "failed" || err.BgColor != "red" {
		t.Fatalf("error badge = %+v", err)
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
