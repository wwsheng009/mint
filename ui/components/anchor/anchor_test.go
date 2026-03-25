package anchor

import (
	"reflect"
	"testing"

	runtimeintent "github.com/wwsheng009/mint/runtime/intent"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/list"
)

func TestNewDefaults(t *testing.T) {
	node := New()
	if node.Tag() != "anchor" {
		t.Fatalf("Tag = %q, want anchor", node.Tag())
	}
	if node.showBorder {
		t.Fatal("showBorder should default to false")
	}
	if node.activeKeyControlled {
		t.Fatal("activeKeyControlled should default to false")
	}
}

func TestBuilderFluentAPI(t *testing.T) {
	node := NewBuilder().
		Key("doc-nav").
		ComponentID("doc-anchor").
		Title("Contents").
		InitialActiveKey("api").
		ViewportHeight(6).
		Width(28).
		ShowBorder(true).
		Items([]Item{
			NewItem("intro", "Introduction"),
			NewItem("api", "API"),
		}).
		BuildVNode()

	if node.Key() != "doc-nav" {
		t.Fatalf("Key = %q, want doc-nav", node.Key())
	}
	if node.componentID != "doc-anchor" {
		t.Fatalf("componentID = %q, want doc-anchor", node.componentID)
	}
	if node.title != "Contents" || node.activeKey != "api" {
		t.Fatalf("unexpected builder state: title=%q activeKey=%q", node.title, node.activeKey)
	}
	if node.viewportHeight != 6 || node.width != 28 || !node.showBorder {
		t.Fatalf("unexpected sizing/border: viewportHeight=%d width=%d showBorder=%v", node.viewportHeight, node.width, node.showBorder)
	}
}

func TestRuntimeChildrenFlattensHierarchy(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propTitle: "Contents",
		propItems: []Item{
			NewItem("intro", "Introduction"),
			NewItem("guide", "Guide",
				NewItem("api", "API"),
			),
		},
		propActiveKey: "api",
	})

	children := inst.RuntimeChildren()
	if len(children) != 1 {
		t.Fatalf("RuntimeChildren len = %d, want 1", len(children))
	}
	root := children[0]
	if root.Tag() != "list" {
		t.Fatalf("root tag = %q, want list", root.Tag())
	}

	rows, _ := root.Props()["rows"].([]string)
	expectedRows := []string{"Introduction", "Guide", "  API"}
	if !reflect.DeepEqual(rows, expectedRows) {
		t.Fatalf("rows = %#v, want %#v", rows, expectedRows)
	}
	if selectedIndex, _ := root.Props()["selectedIndex"].(int); selectedIndex != 2 {
		t.Fatalf("selectedIndex = %d, want 2", selectedIndex)
	}
}

func TestHandleIntentRowSelectUpdatesActiveKey(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propComponentID: "doc-anchor",
		propItems: []Item{
			NewItem("intro", "Introduction"),
			NewItem("api", "API"),
		},
	})

	var emitted []runtimeintent.Intent
	inst.SetIntentEmitter(func(i runtimeintent.Intent) {
		emitted = append(emitted, i)
	})

	if !inst.HandleIntent(list.RowSelectWithID(inst.listComponentID(), 1, "API")) {
		t.Fatal("row select intent should be handled")
	}
	if inst.activeKey != "api" {
		t.Fatalf("activeKey = %q, want api", inst.activeKey)
	}
	if len(emitted) == 0 {
		t.Fatal("expected change intent emission")
	}
	change, ok := emitted[0].(ChangeIntent)
	if !ok {
		t.Fatalf("first emitted intent type = %T, want anchor.ChangeIntent", emitted[0])
	}
	if change.Key != "api" || change.Href != "#api" || change.Title != "API" {
		t.Fatalf("unexpected change intent: %#v", change)
	}
}

func TestDisabledRowSelectResetsList(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propItems: []Item{
			NewItem("intro", "Introduction"),
			{Key: "api", Title: "API", Disabled: true},
		},
	})

	if !inst.HandleIntent(list.RowSelectWithID(inst.listComponentID(), 1, "API")) {
		t.Fatal("row select intent should be handled")
	}
	if inst.activeKey != "intro" {
		t.Fatalf("activeKey = %q, want intro", inst.activeKey)
	}
	if inst.listVersion != 1 {
		t.Fatalf("listVersion = %d, want 1", inst.listVersion)
	}
}

func TestActivateIntentControlledDoesNotMutateState(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propComponentID:         "doc-anchor",
		propItems:               []Item{NewItem("intro", "Introduction"), NewItem("api", "API")},
		propActiveKey:           "intro",
		propActiveKeyControlled: true,
	})

	var emitted []runtimeintent.Intent
	inst.SetIntentEmitter(func(i runtimeintent.Intent) {
		emitted = append(emitted, i)
	})

	if !inst.HandleIntent(ActivateWithID("doc-anchor", "api")) {
		t.Fatal("activate intent should be handled")
	}
	if inst.activeKey != "intro" {
		t.Fatalf("controlled activeKey should stay intro, got %q", inst.activeKey)
	}
	if inst.listVersion != 1 {
		t.Fatalf("listVersion = %d, want 1", inst.listVersion)
	}
	if len(emitted) == 0 {
		t.Fatal("expected change intent emission")
	}
}

func TestFieldBindingEmitsActiveKey(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propComponentID: "doc-anchor",
		propItems: []Item{
			NewItem("intro", "Introduction"),
			NewItem("api", "API"),
		},
		propChangeIntent: runtimeintent.BindField("section"),
	})

	var emitted []runtimeintent.Intent
	inst.SetIntentEmitter(func(i runtimeintent.Intent) {
		emitted = append(emitted, i)
	})

	inst.HandleIntent(list.RowSelectWithID(inst.listComponentID(), 1, "API"))

	if len(emitted) < 2 {
		t.Fatalf("emitted intents len = %d, want at least 2", len(emitted))
	}
	fieldChange, ok := emitted[1].(runtimeintent.FieldChangeIntent)
	if !ok {
		t.Fatalf("second emitted intent type = %T, want intent.FieldChangeIntent", emitted[1])
	}
	if fieldChange.Field != "section" || fieldChange.Value != "api" {
		t.Fatalf("field change = %#v, want field=section value=api", fieldChange)
	}
}
