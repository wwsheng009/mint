package collapse

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	textcomp "github.com/wwsheng009/mint/ui/components/text"
)

func TestNew(t *testing.T) {
	v := New([]Item{
		Section("General", textcomp.New("Body")).WithKey("general"),
	})
	if v == nil {
		t.Fatal("New returned nil")
	}
	if v.Tag() != "collapse" {
		t.Fatalf("Tag = %q, want collapse", v.Tag())
	}
	if !v.bordered {
		t.Fatal("Collapse should be bordered by default")
	}
	if len(v.Items()) != 1 {
		t.Fatalf("Items len = %d, want 1", len(v.Items()))
	}
}

func TestBuilderFluent(t *testing.T) {
	v := NewBuilder().
		Key("settings").
		ComponentID("settings.collapse").
		Item(Section("General", textcomp.New("Body")).WithKey("general")).
		Item(Section("Advanced", textcomp.New("More")).WithKey("advanced")).
		AccordionMode().
		ActiveKeys("general").
		Width(40).
		HeaderStyle(style.NewStyle().Bold(true)).
		ActiveHeaderStyle(style.NewStyle().Underline(true)).
		ContentStyle(style.NewStyle().Foreground(style.Color("cyan"))).
		OnChange(testIntent{}).
		BuildVNode()

	if v.Key() != "settings" {
		t.Fatalf("Key = %q, want settings", v.Key())
	}
	if v.componentID != "settings.collapse" {
		t.Fatalf("componentID = %q, want settings.collapse", v.componentID)
	}
	if !v.accordion {
		t.Fatal("Accordion should be enabled")
	}
	if !v.activeKeysControlled {
		t.Fatal("ActiveKeys should enable controlled mode")
	}
	if got := v.ActiveKeys(); len(got) != 1 || got[0] != "general" {
		t.Fatalf("ActiveKeys = %v, want [general]", got)
	}
	if v.width != 40 {
		t.Fatalf("width = %d, want 40", v.width)
	}
}

func TestNormalizeItemsDeduplicatesKeys(t *testing.T) {
	items := normalizeItems([]Item{
		Section("A", nil).WithKey("dup"),
		Section("B", nil).WithKey("dup"),
		Section("C", nil),
	})
	if items[0].Key != "dup" {
		t.Fatalf("items[0].Key = %q, want dup", items[0].Key)
	}
	if items[1].Key != "dup-1" {
		t.Fatalf("items[1].Key = %q, want dup-1", items[1].Key)
	}
	if items[2].Key != "panel-2" {
		t.Fatalf("items[2].Key = %q, want panel-2", items[2].Key)
	}
}

func TestInstanceToggleMultiple(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propItems: []Item{
			Section("General", textcomp.New("Body")).WithKey("general"),
			Section("Advanced", textcomp.New("More")).WithKey("advanced"),
		},
		propInitialActiveKeys: []string{"general"},
	})
	var emitted []intent.Intent
	inst.SetIntentEmitter(func(i intent.Intent) {
		emitted = append(emitted, i)
	})

	if !inst.HandleIntent(Toggle("advanced", "Advanced", 1)) {
		t.Fatal("expected toggle to be handled")
	}
	if got := inst.ActiveKeys(); len(got) != 2 || got[0] != "general" || got[1] != "advanced" {
		t.Fatalf("ActiveKeys = %v, want [general advanced]", got)
	}
	if len(emitted) != 1 {
		t.Fatalf("emitted len = %d, want 1", len(emitted))
	}
	change, ok := emitted[0].(CollapseChangeIntent)
	if !ok {
		t.Fatalf("emitted[0] = %T, want CollapseChangeIntent", emitted[0])
	}
	if !change.Expanded || change.ToggledKey != "advanced" || change.Index != 1 {
		t.Fatalf("unexpected change intent: %+v", change)
	}
}

func TestInstanceAccordionKeepsSingleExpanded(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propAccordion: true,
		propItems: []Item{
			Section("General", textcomp.New("Body")).WithKey("general"),
			Section("Advanced", textcomp.New("More")).WithKey("advanced"),
		},
		propInitialActiveKeys: []string{"general"},
	})

	if !inst.HandleIntent(Toggle("advanced", "Advanced", 1)) {
		t.Fatal("expected accordion toggle to be handled")
	}
	if got := inst.ActiveKeys(); len(got) != 1 || got[0] != "advanced" {
		t.Fatalf("ActiveKeys = %v, want [advanced]", got)
	}
	if !inst.HandleIntent(Toggle("advanced", "Advanced", 1)) {
		t.Fatal("expected second toggle to collapse active item")
	}
	if got := inst.ActiveKeys(); len(got) != 0 {
		t.Fatalf("ActiveKeys = %v, want []", got)
	}
}

func TestInstanceSetPropsControlledActiveKeys(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propItems: []Item{
			Section("General", textcomp.New("Body")).WithKey("general"),
			Section("Advanced", textcomp.New("More")).WithKey("advanced"),
		},
		propActiveKeysControl: true,
		propActiveKeys:        []string{"general"},
	})

	inst.SetProps(rtui.Props{
		propItems: []Item{
			Section("General", textcomp.New("Body")).WithKey("general"),
			Section("Advanced", textcomp.New("More")).WithKey("advanced"),
		},
		propActiveKeysControl: true,
		propActiveKeys:        []string{"advanced"},
	})

	if got := inst.ActiveKeys(); len(got) != 1 || got[0] != "advanced" {
		t.Fatalf("ActiveKeys = %v, want [advanced]", got)
	}
}

func TestInstanceToggleDisabledIgnored(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propItems: []Item{
			Section("General", textcomp.New("Body")).WithKey("general"),
			Section("Advanced", textcomp.New("More")).WithKey("advanced").WithDisabled(true),
		},
	})

	if inst.HandleIntent(Toggle("advanced", "Advanced", 1)) {
		t.Fatal("expected disabled panel toggle to be ignored")
	}
	if got := inst.ActiveKeys(); len(got) != 0 {
		t.Fatalf("ActiveKeys = %v, want []", got)
	}
}

func TestInstanceToggleEmitsFieldChange(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propItems: []Item{
			Section("General", textcomp.New("Body")).WithKey("general"),
			Section("Advanced", textcomp.New("More")).WithKey("advanced"),
		},
		propChangeIntentField: intent.BindField("sections"),
	})
	var emitted []intent.Intent
	inst.SetIntentEmitter(func(i intent.Intent) {
		emitted = append(emitted, i)
	})

	if !inst.HandleIntent(Toggle("general", "General", 0)) {
		t.Fatal("expected toggle to be handled")
	}
	if len(emitted) != 2 {
		t.Fatalf("emitted len = %d, want 2", len(emitted))
	}
	fieldChange, ok := emitted[1].(intent.FieldChangeIntent)
	if !ok {
		t.Fatalf("emitted[1] = %T, want FieldChangeIntent", emitted[1])
	}
	if fieldChange.Field != "sections" || fieldChange.Value != "general" {
		t.Fatalf("unexpected field change: %+v", fieldChange)
	}
}

func TestInstanceRuntimeChildrenReflectExpandedState(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propItems: []Item{
			Section("General", textcomp.New("General body")).WithKey("general"),
			Section("Advanced", textcomp.New("Advanced body")).WithKey("advanced"),
		},
		propInitialActiveKeys: []string{"general"},
	})

	children := inst.RuntimeChildren()
	if len(children) != 1 {
		t.Fatalf("RuntimeChildren len = %d, want 1", len(children))
	}
	if !containsVNodeText(children[0], "General body") {
		t.Fatal("expanded body should be rendered")
	}
	if containsVNodeText(children[0], "Advanced body") {
		t.Fatal("collapsed body should not be rendered")
	}
}

func TestHandleIntentRespectsComponentID(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propComponentID: "settings.collapse",
		propItems: []Item{
			Section("General", textcomp.New("Body")).WithKey("general"),
		},
	})

	if !inst.HandleIntent(ToggleWithID("settings.collapse", "general", "General", 0)) {
		t.Fatal("expected matching componentID to be handled")
	}
	if inst.HandleIntent(ToggleWithID("other.collapse", "general", "General", 0)) {
		t.Fatal("expected other componentID to be ignored")
	}
}

func containsVNodeText(node rtui.VNode, want string) bool {
	if node == nil {
		return false
	}
	if contentProvider, ok := node.(interface{ Content() string }); ok && contentProvider.Content() == want {
		return true
	}
	for _, child := range node.Children() {
		if containsVNodeText(child, want) {
			return true
		}
	}
	return false
}

type testIntent struct{}

func (testIntent) IntentType() string { return "collapse.testIntent" }
