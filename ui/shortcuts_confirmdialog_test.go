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
		OnConfirm(testConfirmDialogIntent{}).
		OnCancel(testConfirmDialogIntent{}).
		Build()
	if dialog == nil {
		t.Fatal("NewConfirmDialogBuilder().Build() returned nil")
	}
	if dialog.Tag() != "confirmdialog" {
		t.Fatalf("dialog tag = %q, want confirmdialog", dialog.Tag())
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
