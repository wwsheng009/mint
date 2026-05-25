package ui

import (
	"github.com/wwsheng009/mint/runtime/intent"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/formdialog"
)

type FormDialogBuilder = formdialog.Builder
type FormDialogVNode = formdialog.VNode
type FormDialogTargetItem = formdialog.TargetItem

// NewFormDialogBuilder creates a FormDialog builder.
func NewFormDialogBuilder() *formdialog.Builder {
	return formdialog.NewBuilder()
}

// FormDialog creates an open modal form with basic text and intents.
func FormDialog(title, description, formID string, submitIntent, cancelIntent intent.Intent, children ...rtui.VNode) rtui.VNode {
	return formdialog.NewBuilder().
		Title(title).
		Description(description).
		FormID(formID).
		Open(true).
		Children(children...).
		OnSubmit(submitIntent).
		OnCancel(cancelIntent).
		Build()
}

// NewFormDialogDangerReasonActionBuilder creates an opened danger form dialog builder with a required audit reason field.
func NewFormDialogDangerReasonActionBuilder(key, title, description, formID, reasonField, reasonValue, submitText string, submitIntent, cancelIntent intent.Intent) *formdialog.Builder {
	return formdialog.NewDangerReasonAction(key, title, description, formID, reasonField, reasonValue, submitText, submitIntent, cancelIntent)
}

// FormDialogDangerReasonAction creates an opened danger form dialog with a required audit reason field.
func FormDialogDangerReasonAction(key, title, description, formID, reasonField, reasonValue, submitText string, submitIntent, cancelIntent intent.Intent) rtui.VNode {
	return NewFormDialogDangerReasonActionBuilder(key, title, description, formID, reasonField, reasonValue, submitText, submitIntent, cancelIntent).Build()
}

// FormDialogTarget creates a target summary item for FormDialog.
func FormDialogTarget(key, label, value string) formdialog.TargetItem {
	return formdialog.Target(key, label, value)
}

// FormDialogSensitiveTarget creates a masked target summary item for FormDialog.
func FormDialogSensitiveTarget(key, label, value string) formdialog.TargetItem {
	return formdialog.SensitiveTarget(key, label, value)
}
