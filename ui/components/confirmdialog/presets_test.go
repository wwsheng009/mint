package confirmdialog

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/ui/components/button"
)

type presetIntent struct {
	name string
}

func (i presetIntent) IntentType() string { return i.name }

func TestDangerOperationPreset(t *testing.T) {
	dialog := NewDangerOperation(
		"disable-key.confirm",
		"Disable Provider Key",
		"Disable the selected provider key.",
		"Traffic may fail over to another available key.",
		"Disable",
		"actionReason",
		"maintenance",
		presetIntent{"confirm"},
		presetIntent{"cancel"},
		Target("provider", "Provider", "openai"),
		SensitiveTarget("key", "Key", "provider-key-demo"),
	).BuildVNode()

	props := dialog.Props()
	if got := props[propOpen]; got != true {
		t.Fatalf("open = %v, want true", got)
	}
	if got := props[propConfirmVariant]; got != button.VariantDanger {
		t.Fatalf("confirm variant = %v, want danger", got)
	}
	if got := props[propConfirmText]; got != "Disable" {
		t.Fatalf("confirm text = %v, want Disable", got)
	}
	if got := props[propReasonField]; got != "actionReason" {
		t.Fatalf("reason field = %v, want actionReason", got)
	}
	if got := props[propReasonValue]; got != "maintenance" {
		t.Fatalf("reason value = %v, want maintenance", got)
	}
	if got := props[propReasonRequired]; got != true {
		t.Fatalf("reason required = %v, want true", got)
	}
	items := props[propTargetItems].([]TargetItem)
	if len(items) != 2 || !items[1].Sensitive {
		t.Fatalf("target items = %#v", items)
	}
	if _, ok := props[propConfirmIntent].(intent.Intent); !ok {
		t.Fatalf("confirm intent = %T, want intent.Intent", props[propConfirmIntent])
	}
}

func TestAuditReasonDefaults(t *testing.T) {
	dialog := NewBuilder().AuditReason("", "").BuildVNode()
	props := dialog.Props()
	if got := props[propReasonField]; got != defaultAuditReasonField {
		t.Fatalf("reason field = %v, want %s", got, defaultAuditReasonField)
	}
	if got := props[propReasonLabel]; got != defaultAuditReasonLabel {
		t.Fatalf("reason label = %v, want %s", got, defaultAuditReasonLabel)
	}
	if got := props[propReasonRequired]; got != true {
		t.Fatalf("reason required = %v, want true", got)
	}
	if got := props[propReasonPlaceholder]; got != defaultAuditReasonPlaceholder {
		t.Fatalf("reason placeholder = %v, want default", got)
	}
}

func TestRequirePhraseDefaults(t *testing.T) {
	dialog := NewBuilder().RequirePhrase("DELETE", "", "DEL").BuildVNode()
	props := dialog.Props()
	if got := props[propConfirmPhrase]; got != "DELETE" {
		t.Fatalf("confirm phrase = %v, want DELETE", got)
	}
	if got := props[propConfirmPhraseField]; got != defaultConfirmPhraseField {
		t.Fatalf("confirm phrase field = %v, want %s", got, defaultConfirmPhraseField)
	}
	if got := props[propConfirmPhraseValue]; got != "DEL" {
		t.Fatalf("confirm phrase value = %v, want DEL", got)
	}
	if got := props[propConfirmPhraseLabel]; got != defaultConfirmPhraseLabel {
		t.Fatalf("confirm phrase label = %v, want %s", got, defaultConfirmPhraseLabel)
	}
	if got := props[propConfirmPhrasePlaceholder]; got != "Type DELETE to confirm." {
		t.Fatalf("confirm phrase placeholder = %v, want phrase placeholder", got)
	}
}
