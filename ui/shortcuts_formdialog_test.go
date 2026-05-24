package ui

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/ui/components/formdialog"
)

type formDialogShortcutIntent struct {
	name string
}

func (i formDialogShortcutIntent) IntentType() string { return i.name }

func TestFormDialogShortcuts(t *testing.T) {
	builder := NewFormDialogBuilder()
	if builder == nil {
		t.Fatal("NewFormDialogBuilder() returned nil")
	}

	node := builder.
		Key("runtime-reload-dialog").
		Title("Reload Runtime").
		Description("Reload runtime configuration.").
		FormID("runtime-reload-form").
		Opened().
		Children(FormInputItem("reason", "Reason", "maintenance", FormInputForForm("runtime-reload-form"))).
		SubmitText("Reload").
		OnSubmit(formDialogShortcutIntent{"formdialog.submit"}).
		OnCancel(formDialogShortcutIntent{"formdialog.cancel"}).
		BuildVNode()

	if node.Tag() != "formdialog" {
		t.Fatalf("node tag = %q, want formdialog", node.Tag())
	}
	props := node.Props()
	if got := props["formID"]; got != "runtime-reload-form" {
		t.Fatalf("formID = %v, want runtime-reload-form", got)
	}
	if _, ok := props["submitIntent"].(intent.Intent); !ok {
		t.Fatalf("submitIntent = %T, want intent.Intent", props["submitIntent"])
	}

	short := FormDialog(
		"Reload Runtime",
		"Reload runtime configuration.",
		"runtime-reload-form",
		formDialogShortcutIntent{"formdialog.submit"},
		formDialogShortcutIntent{"formdialog.cancel"},
		FormInputItem("reason", "Reason", "maintenance", FormInputForForm("runtime-reload-form")),
	)
	dialog, ok := short.(*formdialog.VNode)
	if !ok {
		t.Fatalf("FormDialog() = %T, want *formdialog.VNode", short)
	}
	if got := dialog.Props()["open"]; got != true {
		t.Fatalf("open = %v, want true", got)
	}
}
