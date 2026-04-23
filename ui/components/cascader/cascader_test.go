package cascader

import (
	"strings"
	"testing"

	"github.com/wwsheng009/mint/runtime/action"
	runtimeintent "github.com/wwsheng009/mint/runtime/intent"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

func testOptions() []Option {
	return []Option{
		Node("zj", "Zhejiang",
			Leaf("hz", "Hangzhou"),
			Leaf("nb", "Ningbo"),
		),
		Node("js", "Jiangsu",
			Leaf("nj", "Nanjing"),
		),
	}
}

func TestNew(t *testing.T) {
	v := New()
	if v == nil {
		t.Fatal("New returned nil")
	}
	if v.Tag() != "cascader" {
		t.Fatalf("Tag = %q, want cascader", v.Tag())
	}
	if v.placeholder != defaultPlaceholder {
		t.Fatalf("placeholder = %q, want %q", v.placeholder, defaultPlaceholder)
	}
}

func TestBuilderFluent(t *testing.T) {
	v := NewBuilder().
		Key("location").
		SetID("location-input").
		ComponentID("filters.location").
		Options(testOptions()).
		Value("zj", "hz").
		ChangeOnSelect(true).
		BuildVNode()

	if v.Key() != "location" || v.ID() != "location-input" {
		t.Fatalf("key/id = (%q,%q)", v.Key(), v.ID())
	}
	if v.componentID != "filters.location" {
		t.Fatalf("componentID = %q", v.componentID)
	}
	if !v.valueControlled || !v.changeOnSelect {
		t.Fatal("valueControlled/changeOnSelect should be true")
	}
	if len(v.value) != 2 || v.value[0] != "zj" || v.value[1] != "hz" {
		t.Fatalf("value = %v", v.value)
	}
}

func TestInstanceKeyboardNavigationCommitsLeaf(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propOptions: testOptions(),
	})

	if !inst.HandleAction(action.NewAction(action.ActionEnter)) {
		t.Fatal("enter should open popup")
	}
	if !inst.open {
		t.Fatal("popup should be open")
	}
	if !inst.HandleAction(action.NewAction(action.ActionEnter)) {
		t.Fatal("enter on branch should move to child level")
	}
	if inst.activeLevel != 1 {
		t.Fatalf("activeLevel = %d, want 1", inst.activeLevel)
	}
	if !inst.HandleAction(action.NewAction(action.ActionNavigateDown)) {
		t.Fatal("down should move to second child")
	}
	if !inst.HandleAction(action.NewAction(action.ActionEnter)) {
		t.Fatal("enter on leaf should commit selection")
	}

	if got := inst.SelectedValue(); got != "zj/nb" {
		t.Fatalf("SelectedValue = %q, want zj/nb", got)
	}
	if inst.open {
		t.Fatal("popup should close after leaf commit")
	}
}

func TestInstanceChangeOnSelectEmitsFieldChangeAndCascaderChange(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propOptions:        testOptions(),
		propChangeOnSelect: true,
		propChangeIntent:   runtimeintent.BindField("filters.location"),
	})

	var emitted []runtimeintent.Intent
	inst.SetIntentEmitter(func(i runtimeintent.Intent) {
		emitted = append(emitted, i)
	})

	if !inst.HandleAction(action.NewAction(action.ActionEnter)) {
		t.Fatal("enter should open popup")
	}
	if !inst.HandleAction(action.NewAction(action.ActionEnter)) {
		t.Fatal("enter should commit branch when changeOnSelect is true")
	}

	if got := inst.SelectedValue(); got != "zj" {
		t.Fatalf("SelectedValue = %q, want zj", got)
	}
	if len(emitted) != 2 {
		t.Fatalf("emitted len = %d, want 2", len(emitted))
	}
	fieldChange, ok := emitted[0].(runtimeintent.FieldChangeIntent)
	if !ok {
		t.Fatalf("first emitted = %T, want FieldChangeIntent", emitted[0])
	}
	if fieldChange.Field != "filters.location" || fieldChange.Value != "zj" {
		t.Fatalf("field change = %+v", fieldChange)
	}
	change, ok := emitted[1].(ChangeIntent)
	if !ok {
		t.Fatalf("second emitted = %T, want cascader.ChangeIntent", emitted[1])
	}
	if change.Value != "zj" || change.Label != "Zhejiang" {
		t.Fatalf("change intent = %+v", change)
	}
}

func TestInstanceControlledValueSync(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propOptions:         testOptions(),
		propValueControlled: true,
		propValue:           []string{"zj", "hz"},
	})

	inst.SetProps(rtui.Props{
		propOptions:         testOptions(),
		propValueControlled: true,
		propValue:           []string{"js", "nj"},
	})

	if got := inst.SelectedValue(); got != "js/nj" {
		t.Fatalf("SelectedValue = %q, want js/nj", got)
	}
}

func TestInstancePaintOpenPopupIncludesColumns(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propOptions: testOptions(),
	})
	inst.openPopup()

	cmds := inst.Paint(0, 0)
	var joined []string
	for _, cmd := range cmds {
		joined = append(joined, cmd.Text)
	}
	text := strings.Join(joined, "\n")
	if !strings.Contains(text, "Zhejiang") || !strings.Contains(text, "Hangzhou") {
		t.Fatalf("paint output missing cascader labels:\n%s", text)
	}
}
