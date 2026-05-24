package formdialog

import (
	"strings"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/ui/components/button"
	"github.com/wwsheng009/mint/ui/components/forminputitem"
	"github.com/wwsheng009/mint/ui/components/validation"
)

const (
	defaultAuditReasonField       = "reason"
	defaultAuditReasonLabel       = "Reason"
	defaultAuditReasonPlaceholder = "Describe why this operation is necessary."
	defaultAuditReasonWidth       = 52
)

// NewDangerReasonAction builds an opened short form dialog for high-risk
// operational actions that need one required audit reason field.
func NewDangerReasonAction(key, title, description, formID, reasonField, reasonValue, submitText string, submitIntent, cancelIntent intent.Intent) *Builder {
	builder := NewBuilder().
		Key(key).
		Title(title).
		Description(description).
		Opened().
		SubmitVariant(button.VariantDanger).
		OnSubmit(submitIntent).
		OnCancel(cancelIntent)
	if strings.TrimSpace(formID) != "" {
		builder.FormID(formID)
	}
	builder.AuditReason(reasonField, reasonValue)
	if strings.TrimSpace(submitText) != "" {
		builder.SubmitText(submitText)
	}
	return builder
}

// AuditReason appends the standard required audit reason field to the form.
func (b *Builder) AuditReason(field, value string) *Builder {
	return b.AuditReasonWithLabel(field, defaultAuditReasonLabel, value)
}

// AuditReasonWithLabel appends a required audit reason field with a custom label.
func (b *Builder) AuditReasonWithLabel(field, label, value string) *Builder {
	field = strings.TrimSpace(field)
	if field == "" {
		field = defaultAuditReasonField
	}
	label = strings.TrimSpace(label)
	if label == "" {
		label = defaultAuditReasonLabel
	}
	formID := strings.TrimSpace(b.node.formID)
	if formID == "" {
		formID = defaultFormID(b.node.key)
		b.FormID(formID)
	}
	b.Value(field, value)
	return b.Child(forminputitem.New(
		field,
		label,
		value,
		forminputitem.ForForm(formID),
		forminputitem.Placeholder(defaultAuditReasonPlaceholder),
		forminputitem.Width(defaultAuditReasonWidth),
		forminputitem.Validators(validation.Required()),
	))
}

func defaultFormID(key string) string {
	key = strings.TrimSpace(key)
	if key == "" {
		return "formdialog-form"
	}
	return key + "-form"
}
