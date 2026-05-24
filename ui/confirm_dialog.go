package ui

import (
	"github.com/wwsheng009/mint/runtime/intent"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/button"
	"github.com/wwsheng009/mint/ui/components/confirmdialog"
)

type ConfirmDialogBuilder = confirmdialog.Builder
type ConfirmDialogVNode = confirmdialog.VNode
type ConfirmDialogTarget = confirmdialog.TargetItem

// NewConfirmDialogBuilder creates a ConfirmDialog builder.
func NewConfirmDialogBuilder() *confirmdialog.Builder {
	return confirmdialog.NewBuilder()
}

// ConfirmDialog creates an open confirmation dialog with basic text and intents.
func ConfirmDialog(title, message string, confirmIntent, cancelIntent intent.Intent) rtui.VNode {
	return confirmdialog.NewBuilder().
		Title(title).
		Message(message).
		Open(true).
		OnConfirm(confirmIntent).
		OnCancel(cancelIntent).
		Build()
}

// ConfirmDialogTarget creates a target summary item.
func ConfirmDialogTargetItem(key, label, value string) confirmdialog.TargetItem {
	return confirmdialog.Target(key, label, value)
}

// ConfirmDialogSensitiveTarget creates a masked target summary item.
func ConfirmDialogSensitiveTarget(key, label, value string) confirmdialog.TargetItem {
	return confirmdialog.SensitiveTarget(key, label, value)
}

// ConfirmDialogDanger selects the danger confirm button variant.
func ConfirmDialogDanger() button.Variant {
	return button.VariantDanger
}
