package confirmdialog

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
	dialog := NewBuilder().
		Key("confirm-disable").
		Title("Disable Key").
		Message("Disable the selected provider key.").
		Warning("Requests may fail over to another key.").
		Open(true).
		Width(72).
		Height(20).
		Target(Target("provider", "Provider", "openai")).
		Target(SensitiveTarget("key", "Key", "provider-key-demo")).
		ReasonField("actionReason").
		ReasonValue("maintenance").
		ReasonRequired(true).
		ConfirmPhrase("DISABLE", "confirmText", "DISABLE").
		ConfirmText("Disable").
		ConfirmVariant(button.VariantDanger).
		OnConfirm(testIntent{"confirm"}).
		OnCancel(testIntent{"cancel"}).
		BuildVNode()

	if dialog.Key() != "confirm-disable" {
		t.Fatalf("key = %q, want confirm-disable", dialog.Key())
	}
	props := dialog.Props()
	if got := props[propTitle]; got != "Disable Key" {
		t.Fatalf("title = %v, want Disable Key", got)
	}
	if got := props[propOpen]; got != true {
		t.Fatalf("open = %v, want true", got)
	}
	if got := props[propWidth]; got != 72 {
		t.Fatalf("width = %v, want 72", got)
	}
	items := props[propTargetItems].([]TargetItem)
	if len(items) != 2 || !items[1].Sensitive {
		t.Fatalf("target items = %#v", items)
	}
	if got := props[propConfirmVariant]; got != button.VariantDanger {
		t.Fatalf("confirm variant = %v, want danger", got)
	}
	if got := props[propConfirmPhrase]; got != "DISABLE" {
		t.Fatalf("confirm phrase = %v, want DISABLE", got)
	}
	if got := props[propConfirmPhraseField]; got != "confirmText" {
		t.Fatalf("confirm phrase field = %v, want confirmText", got)
	}
	if got := props[propConfirmPhraseValue]; got != "DISABLE" {
		t.Fatalf("confirm phrase value = %v, want DISABLE", got)
	}
	if _, ok := props[propConfirmIntent].(intent.Intent); !ok {
		t.Fatalf("confirm intent = %T, want intent.Intent", props[propConfirmIntent])
	}
}

func TestRuntimeChildrenClosedDialogIsEmpty(t *testing.T) {
	inst := NewBuilder().Key("closed").Open(false).BuildInstance()
	if children := inst.RuntimeChildren(); len(children) != 0 {
		t.Fatalf("RuntimeChildren len = %d, want 0", len(children))
	}
}

func TestRuntimeChildrenBuildsModalWithReasonAndFooter(t *testing.T) {
	inst := NewBuilder().
		Key("confirm-disable").
		Title("Disable Key").
		Message("Disable the selected key.").
		Open(true).
		Target(Target("provider", "Provider", "openai")).
		ReasonField("actionReason").
		ReasonRequired(true).
		ConfirmText("Disable").
		OnConfirm(testIntent{"confirm"}).
		OnCancel(testIntent{"cancel"}).
		BuildInstance()

	children := inst.RuntimeChildren()
	if len(children) != 1 {
		t.Fatalf("RuntimeChildren len = %d, want 1", len(children))
	}
	root := children[0]
	if root.Tag() != "vstack" {
		t.Fatalf("root tag = %q, want vstack", root.Tag())
	}
	if findVNodeByKey(root, "confirm-disable-reason-input") == nil {
		t.Fatal("reason input not found")
	}
	confirm := findVNodeByKey(root, "confirm-disable-confirm")
	if confirm == nil {
		t.Fatal("confirm button not found")
	}
	if got := confirm.Props()["disabled"]; got != true {
		t.Fatalf("confirm disabled = %v, want true while reason is empty", got)
	}
	cancel := findVNodeByKey(root, "confirm-disable-cancel")
	if cancel == nil {
		t.Fatal("cancel button not found")
	}
}

func TestRuntimeChildrenConfirmPhraseDisablesUntilExactMatch(t *testing.T) {
	inst := NewBuilder().
		Key("confirm-delete").
		Title("Delete Rule").
		Message("Delete the selected alert rule.").
		Open(true).
		Height(20).
		ConfirmPhrase("DELETE", "confirmPhrase", "delete").
		ConfirmText("Delete").
		OnConfirm(testIntent{"confirm"}).
		BuildInstance()

	children := inst.RuntimeChildren()
	if len(children) != 1 {
		t.Fatalf("RuntimeChildren len = %d, want 1", len(children))
	}
	if findVNodeByKey(children[0], "confirm-delete-confirm-phrase-input") == nil {
		t.Fatal("confirm phrase input not found")
	}
	if got := children[0].Props()["height"]; got != minConfirmPhraseHeight {
		t.Fatalf("dialog height = %v, want min confirm phrase height", got)
	}
	confirm := findVNodeByKey(children[0], "confirm-delete-confirm")
	if confirm == nil {
		t.Fatal("confirm button not found")
	}
	if got := confirm.Props()["disabled"]; got != true {
		t.Fatalf("confirm disabled = %v, want true while phrase mismatches", got)
	}

	changed := inst.SetProps(NewBuilder().
		Key("confirm-delete").
		Open(true).
		ConfirmPhrase("DELETE", "confirmPhrase", "DELETE").
		BuildVNode().
		Props())
	if !changed {
		t.Fatal("SetProps should report changed confirm phrase value")
	}
	children = inst.RuntimeChildren()
	confirm = findVNodeByKey(children[0], "confirm-delete-confirm")
	if got := confirm.Props()["disabled"]; got != false {
		t.Fatalf("confirm disabled = %v, want false after exact phrase match", got)
	}
}

func TestNormalizeTargetItems(t *testing.T) {
	items := normalizeTargetItems([]TargetItem{{Key: "dup"}, {Key: "dup"}, {}})
	if items[0].Key != "dup" || items[1].Key != "dup-1" || items[2].Key != "target-2" {
		t.Fatalf("item keys = %q, %q, %q", items[0].Key, items[1].Key, items[2].Key)
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
