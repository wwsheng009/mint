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
	if root.Tag() != "panel" {
		t.Fatalf("root tag = %q, want panel", root.Tag())
	}
	if findVNodeByKey(root, "runtime-reload-form") == nil {
		t.Fatal("form not found")
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
