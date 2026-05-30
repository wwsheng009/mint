package formdialog

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/intent"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/button"
	"github.com/wwsheng009/mint/ui/components/form"
	"github.com/wwsheng009/mint/ui/components/input"
)

type testIntent struct {
	name string
}

func (i testIntent) IntentType() string { return i.name }

func TestBuilderAndProps(t *testing.T) {
	dialog := NewBuilder().
		Key("runtime-reload-dialog").
		Title("Reload Runtime").
		Description("Reload runtime configuration.").
		Open(true).
		Width(76).
		Height(22).
		FormID("runtime-reload-form").
		Layout(form.LayoutHorizontal).
		Value("reason", "maintenance").
		ValidateAll(true).
		Target(Target("group", "Group", "default")).
		Target(SensitiveTarget("key", "Key", "provider-key-demo")).
		Child(input.NewBuilder().Key("reason-input").Value("maintenance").Build()).
		SubmitText("Reload").
		CancelText("Close").
		SubmitVariant(button.VariantDanger).
		SubmitDisabled(true).
		DisabledReason("Reason is required.").
		OnSubmit(testIntent{"submit"}).
		OnCancel(testIntent{"cancel"}).
		OnClose(testIntent{"close"}).
		CloseOnEsc(false).
		CloseOnBackdrop(false).
		BuildVNode()

	if dialog.Key() != "runtime-reload-dialog" {
		t.Fatalf("key = %q, want runtime-reload-dialog", dialog.Key())
	}
	props := dialog.Props()
	if got := props[propTitle]; got != "Reload Runtime" {
		t.Fatalf("title = %v, want Reload Runtime", got)
	}
	if got := props[propOpen]; got != true {
		t.Fatalf("open = %v, want true", got)
	}
	if got := props[propWidth]; got != 76 {
		t.Fatalf("width = %v, want 76", got)
	}
	if got := props[propLayout]; got != form.LayoutHorizontal {
		t.Fatalf("layout = %v, want horizontal", got)
	}
	if got := props[propSubmitVariant]; got != button.VariantDanger {
		t.Fatalf("submit variant = %v, want danger", got)
	}
	if got := props[propSubmitDisabled]; got != true {
		t.Fatalf("submit disabled = %v, want true", got)
	}
	targets := props[propTargetItems].([]TargetItem)
	if len(targets) != 2 {
		t.Fatalf("target items len = %d, want 2", len(targets))
	}
	if got := targets[1].Sensitive; got != true {
		t.Fatalf("sensitive target = %v, want true", got)
	}
	if _, ok := props[propSubmitIntent].(intent.Intent); !ok {
		t.Fatalf("submit intent = %T, want intent.Intent", props[propSubmitIntent])
	}
	if children := props[propChildren].([]rtui.VNode); len(children) != 1 {
		t.Fatalf("children len = %d, want 1", len(children))
	}
}

func TestRuntimeChildrenClosedDialogIsEmpty(t *testing.T) {
	inst := NewBuilder().Key("closed").Closed().BuildInstance()
	if children := inst.RuntimeChildren(); len(children) != 0 {
		t.Fatalf("RuntimeChildren len = %d, want 0", len(children))
	}
}

func TestRuntimeChildrenBuildsModalFormAndFooter(t *testing.T) {
	inst := NewBuilder().
		Key("runtime-reload-dialog").
		Title("Reload Runtime").
		Description("Reload runtime configuration.").
		Opened().
		FormID("runtime-reload-form").
		Target(Target("group", "Group", "default")).
		Target(SensitiveTarget("key", "Key", "provider-key-demo")).
		Child(input.NewBuilder().Key("reason-input").Value("maintenance").Build()).
		SubmitText("Reload").
		SubmitDisabled(true).
		DisabledReason("Reason is required.").
		OnSubmit(testIntent{"submit"}).
		OnCancel(testIntent{"cancel"}).
		BuildInstance()

	children := inst.RuntimeChildren()
	if len(children) != 1 {
		t.Fatalf("RuntimeChildren len = %d, want 1", len(children))
	}
	root := children[0]
	if root.Tag() != "modal" {
		t.Fatalf("root tag = %q, want modal", root.Tag())
	}
	if got := root.GetLayer(); got != rtui.LayerModal {
		t.Fatalf("root layer = %v, want %v", got, rtui.LayerModal)
	}
	props := root.Props()
	if got := props["isOpen"]; got != true {
		t.Fatalf("modal isOpen = %v, want true", got)
	}
	if got := props["centered"]; got != true {
		t.Fatalf("modal centered = %v, want true", got)
	}
	if findVNodeByKey(root, "runtime-reload-form") == nil {
		t.Fatal("form not found")
	}
	targets := findVNodeByKey(root, "runtime-reload-dialog-targets")
	if targets == nil {
		t.Fatal("target summary not found")
	}
	if findVNodeByKey(root, "reason-input") == nil {
		t.Fatal("form child not found")
	}
	submit := findVNodeByKey(root, "runtime-reload-dialog-submit")
	if submit == nil {
		t.Fatal("submit button not found")
	}
	if got := submit.Props()["disabled"]; got != true {
		t.Fatalf("submit disabled = %v, want true", got)
	}
	cancel := findVNodeByKey(root, "runtime-reload-dialog-cancel")
	if cancel == nil {
		t.Fatal("cancel button not found")
	}
}

func TestTargetItemsNormalizeDuplicateAndMultilineKeys(t *testing.T) {
	dialog := NewBuilder().
		Key("target-dialog").
		Target(Target("", "Group\nName", "default\tgroup")).
		Target(Target("", "Group", "canary")).
		BuildVNode()

	targets := dialog.Props()[propTargetItems].([]TargetItem)
	if len(targets) != 2 {
		t.Fatalf("target items len = %d, want 2", len(targets))
	}
	if targets[0].Key != "target-0" || targets[1].Key != "target-1" {
		t.Fatalf("target keys = %#v, want generated stable keys", []string{targets[0].Key, targets[1].Key})
	}
	if targets[0].Label != "Group Name" || targets[0].Value != "default group" {
		t.Fatalf("normalized target = %#v", targets[0])
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
