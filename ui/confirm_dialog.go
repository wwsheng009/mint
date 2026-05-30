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

// NewConfirmDangerOperationBuilder creates an opened high-risk operation dialog builder.
func NewConfirmDangerOperationBuilder(key, title, message, warning, confirmText, reasonField, reasonValue string, confirmIntent, cancelIntent intent.Intent, targets ...confirmdialog.TargetItem) *confirmdialog.Builder {
	return confirmdialog.NewDangerOperation(key, title, message, warning, confirmText, reasonField, reasonValue, confirmIntent, cancelIntent, targets...)
}

// ConfirmDangerOperation creates an opened high-risk operation confirmation dialog.
func ConfirmDangerOperation(key, title, message, warning, confirmText, reasonField, reasonValue string, confirmIntent, cancelIntent intent.Intent, targets ...confirmdialog.TargetItem) rtui.VNode {
	return NewConfirmDangerOperationBuilder(key, title, message, warning, confirmText, reasonField, reasonValue, confirmIntent, cancelIntent, targets...).Build()
}

// ConfirmDialogTarget creates a target summary item.
func ConfirmDialogTargetItem(key, label, value string) confirmdialog.TargetItem {
	return confirmdialog.Target(key, label, value)
}

// ConfirmDialogSensitiveTarget creates a masked target summary item.
func ConfirmDialogSensitiveTarget(key, label, value string) confirmdialog.TargetItem {
	return confirmdialog.SensitiveTarget(key, label, value)
}

// ConfirmDialogAPITarget creates a standard API route target summary item.
func ConfirmDialogAPITarget(method, path string) confirmdialog.TargetItem {
	return confirmdialog.APITarget(method, path)
}

// ConfirmDialogImpactTarget creates a standard operational impact target summary item.
func ConfirmDialogImpactTarget(level, text string) confirmdialog.TargetItem {
	return confirmdialog.ImpactTarget(level, text)
}

// ConfirmDialogDanger selects the danger confirm button variant.
func ConfirmDialogDanger() button.Variant {
	return button.VariantDanger
}
