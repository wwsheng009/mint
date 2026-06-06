package ui

import "testing"

type testConfirmDialogIntent struct{}

func (testConfirmDialogIntent) IntentType() string { return "test.confirmdialog" }

func TestConfirmDialogShortcuts(t *testing.T) {
	dialog := NewConfirmDialogBuilder().
		Key("confirm").
		Title("Confirm").
		Message("Proceed?").
		Open(true).
		Target(ConfirmDialogTargetItem("target", "Target", "demo")).
		RequirePhrase("CONFIRM", "confirmPhrase", "CONFIRM").
		OnConfirm(testConfirmDialogIntent{}).
		OnCancel(testConfirmDialogIntent{}).
		Build()
	if dialog == nil {
		t.Fatal("NewConfirmDialogBuilder().Build() returned nil")
	}
	if dialog.Tag() != "confirmdialog" {
		t.Fatalf("dialog tag = %q, want confirmdialog", dialog.Tag())
	}
	if got := dialog.Props()["confirmPhrase"]; got != "CONFIRM" {
		t.Fatalf("confirm phrase = %v, want CONFIRM", got)
	}
}

func TestConfirmDialogDirectShortcut(t *testing.T) {
	dialog := ConfirmDialog("Confirm", "Proceed?", testConfirmDialogIntent{}, testConfirmDialogIntent{})
	if dialog == nil {
		t.Fatal("ConfirmDialog() returned nil")
	}
	if dialog.Tag() != "confirmdialog" {
		t.Fatalf("ConfirmDialog().Tag() = %q, want confirmdialog", dialog.Tag())
	}
}

func TestConfirmDialogDangerOperationShortcut(t *testing.T) {
	dialog := ConfirmDangerOperation(
		"disable-key.confirm",
		"Disable Provider Key",
		"Disable the selected key.",
		"Traffic may fail over.",
		"Disable",
		"actionReason",
		"maintenance",
		testConfirmDialogIntent{},
		testConfirmDialogIntent{},
		ConfirmDialogAPITarget("post", "/admin/loadbalancer/providers/{provider}/keys/{key}/disable"),
		ConfirmDialogImpactTarget("high", "Traffic may fail over."),
		ConfirmDialogTargetItem("provider", "Provider", "openai"),
		ConfirmDialogSensitiveTarget("key", "Key", "provider-key-demo"),
	)
	if dialog == nil {
		t.Fatal("ConfirmDangerOperation() returned nil")
	}
	if dialog.Tag() != "confirmdialog" {
		t.Fatalf("ConfirmDangerOperation().Tag() = %q, want confirmdialog", dialog.Tag())
	}
	props := dialog.Props()
	if got := props["reasonField"]; got != "actionReason" {
		t.Fatalf("reasonField = %v, want actionReason", got)
	}
	if got := props["reasonRequired"]; got != true {
		t.Fatalf("reasonRequired = %v, want true", got)
	}
	if got := props["confirmText"]; got != "Disable" {
		t.Fatalf("confirmText = %v, want Disable", got)
	}
	targets := props["targetItems"].([]ConfirmDialogTarget)
	if targets[0].Key != "api" || targets[0].Value != "POST /admin/loadbalancer/providers/{provider}/keys/{key}/disable" {
		t.Fatalf("api target = %+v", targets[0])
	}
	if targets[1].Key != "impact" || targets[1].Value != "high: Traffic may fail over." {
		t.Fatalf("impact target = %+v", targets[1])
	}
}
