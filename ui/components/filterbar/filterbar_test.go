package filterbar

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/intent"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/button"
)

type testIntent struct {
	name string
}

func (i testIntent) IntentType() string { return i.name }

func TestBuilderAndProps(t *testing.T) {
	bar := NewBuilder().
		Key("providers").
		Title("Provider Filters").
		Width(80).
		LabelWidth(8).
		Gap(2).
		RowGap(1).
		Field(Search("query", "Query", "openai").WithPlaceholder("provider").WithWidth(20).ForField("query")).
		Field(Select("status", "Status", []Option{
			{Value: "all", Label: "All"},
			{Value: "degraded", Label: "Degraded"},
		}).WithSelectedIndex(1).ForField("status")).
		Action(Button("refresh", "Refresh", testIntent{"refresh"}).Primary()).
		BuildVNode()

	if bar.Key() != "providers" {
		t.Fatalf("key = %q, want providers", bar.Key())
	}
	props := bar.Props()
	if got := props[propTitle]; got != "Provider Filters" {
		t.Fatalf("title = %v, want Provider Filters", got)
	}
	if got := props[propWidth]; got != 80 {
		t.Fatalf("width = %v, want 80", got)
	}
	if got := props[propLabelWidth]; got != 8 {
		t.Fatalf("label width = %v, want 8", got)
	}
	fields := props[propFields].([]Field)
	if len(fields) != 2 {
		t.Fatalf("fields len = %d, want 2", len(fields))
	}
	if fields[0].Kind != FieldSearch || fields[0].FieldName != "query" {
		t.Fatalf("search field = %#v", fields[0])
	}
	if fields[1].Kind != FieldSelect || fields[1].SelectedIndex != 1 || !fields[1].HasSelectedIndex {
		t.Fatalf("select field = %#v", fields[1])
	}
	actions := props[propActions].([]Action)
	if len(actions) != 1 || actions[0].Variant != button.VariantPrimary {
		t.Fatalf("actions = %#v", actions)
	}
}

func TestRuntimeChildrenBuildControlsWithBindings(t *testing.T) {
	inst := NewBuilder().
		Key("providers").
		Width(72).
		LabelWidth(10).
		Field(Search("query", "Query", "openai").ForField("query")).
		Field(Select("status", "Status", []Option{
			{Value: "all", Label: "All"},
			{Value: "failed", Label: "Failed"},
		}).WithSelectedIndex(1).ForField("status")).
		Action(Button("reset", "Reset", testIntent{"reset"}).Secondary()).
		BuildInstance()

	children := inst.RuntimeChildren()
	if len(children) != 1 {
		t.Fatalf("RuntimeChildren len = %d, want 1", len(children))
	}
	root := children[0]
	if root.Tag() != "vstack" {
		t.Fatalf("root tag = %q, want vstack", root.Tag())
	}

	search := findVNodeByKey(root, "providers-control-query")
	if search == nil {
		t.Fatal("search control not found")
	}
	if got := search.Props()["value"]; got != "openai" {
		t.Fatalf("search value = %v, want openai", got)
	}
	if _, ok := search.Props()["changeIntent"].(intent.FieldIntent); !ok {
		t.Fatalf("search changeIntent = %T, want FieldIntent", search.Props()["changeIntent"])
	}

	status := findVNodeByKey(root, "providers-control-status")
	if status == nil {
		t.Fatal("status control not found")
	}
	if got := status.Props()["selectedIndex"]; got != 1 {
		t.Fatalf("selectedIndex = %v, want 1", got)
	}
	if _, ok := status.Props()["changeIntent"].(intent.FieldIntent); !ok {
		t.Fatalf("select changeIntent = %T, want FieldIntent", status.Props()["changeIntent"])
	}

	reset := findVNodeByKey(root, "providers-action-reset")
	if reset == nil {
		t.Fatal("reset action not found")
	}
	if got := reset.Props()["variant"]; got != button.VariantSecondary {
		t.Fatalf("reset variant = %v, want secondary", got)
	}
}

func TestNormalizeFieldsAndActions(t *testing.T) {
	fields := normalizeFields([]Field{
		{Key: "dup", Width: -1, LabelWidth: -1, Kind: FieldKind("bad")},
		{Key: "dup", Kind: FieldSelect},
	})
	if fields[0].Key != "dup" || fields[1].Key != "dup-1" {
		t.Fatalf("field keys = %q, %q", fields[0].Key, fields[1].Key)
	}
	if fields[0].Width != 0 || fields[0].LabelWidth != 0 {
		t.Fatalf("field widths = %d/%d, want 0/0", fields[0].Width, fields[0].LabelWidth)
	}
	if fields[0].Kind != FieldText {
		t.Fatalf("field kind = %v, want text", fields[0].Kind)
	}
	if fields[1].SelectedIndex != -1 {
		t.Fatalf("select selected index = %d, want -1", fields[1].SelectedIndex)
	}

	actions := normalizeActions([]Action{{Key: "dup", Width: -1}, {Key: "dup"}})
	if actions[0].Key != "dup" || actions[1].Key != "dup-1" {
		t.Fatalf("action keys = %q, %q", actions[0].Key, actions[1].Key)
	}
	if actions[0].Width != 0 {
		t.Fatalf("action width = %d, want 0", actions[0].Width)
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
