package ui

import (
	"github.com/wwsheng009/mint/runtime/intent"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/formdialog"
)

type FormDialogBuilder = formdialog.Builder
type FormDialogVNode = formdialog.VNode

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
