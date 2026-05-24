package formdialog

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/intent"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/button"
)

type presetIntent struct {
	name string
}

func (i presetIntent) IntentType() string { return i.name }

func TestDangerReasonActionPreset(t *testing.T) {
	dialog := NewDangerReasonAction(
		"reload-runtime.dialog",
		"Reload Runtime",
		"Reload runtime configuration with an audit reason.",
		"reload-runtime-form",
		"actionReason",
		"maintenance",
		"Reload",
		presetIntent{"submit"},
		presetIntent{"cancel"},
	).BuildVNode()

	props := dialog.Props()
	if got := props[propOpen]; got != true {
		t.Fatalf("open = %v, want true", got)
	}
	if got := props[propFormID]; got != "reload-runtime-form" {
		t.Fatalf("form id = %v, want reload-runtime-form", got)
	}
	if got := props[propSubmitVariant]; got != button.VariantDanger {
		t.Fatalf("submit variant = %v, want danger", got)
	}
	if got := props[propSubmitText]; got != "Reload" {
		t.Fatalf("submit text = %v, want Reload", got)
	}
	values := props[propValues].(map[string]interface{})
	if got := values["actionReason"]; got != "maintenance" {
		t.Fatalf("actionReason value = %v, want maintenance", got)
	}
	children, ok := props[propChildren].([]rtui.VNode)
	if !ok {
		t.Fatalf("children = %T, want []rtui.VNode", props[propChildren])
	}
	if len(children) != 1 {
		t.Fatalf("children len = %d, want 1", len(children))
	}
}

func TestDangerReasonActionChildren(t *testing.T) {
	dialog := NewDangerReasonAction(
		"disable-key.dialog",
		"Disable Key",
		"Disable a key.",
		"",
		"",
		"",
		"Disable",
		presetIntent{"submit"},
		presetIntent{"cancel"},
	).BuildVNode()

	props := dialog.Props()
	if got := props[propFormID]; got != "disable-key.dialog-form" {
		t.Fatalf("form id = %v, want key-derived form id", got)
	}
	children, ok := props[propChildren].([]rtui.VNode)
	if !ok {
		t.Fatalf("children = %T, want []rtui.VNode", props[propChildren])
	}
	if len(children) != 1 {
		t.Fatalf("children len = %d, want 1", len(children))
	}
	if _, ok := props[propSubmitIntent].(intent.Intent); !ok {
		t.Fatalf("submit intent = %T, want intent.Intent", props[propSubmitIntent])
	}
}
