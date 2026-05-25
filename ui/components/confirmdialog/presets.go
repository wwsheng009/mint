package confirmdialog

import (
	"strings"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/ui/components/button"
)

const (
	defaultAuditReasonField       = "reason"
	defaultAuditReasonLabel       = "Reason"
	defaultAuditReasonPlaceholder = "Describe why this operation is necessary."
	defaultConfirmPhraseField     = "confirmPhrase"
	defaultConfirmPhraseLabel     = "Confirmation"
)

// NewDangerOperation builds an opened confirmation dialog for high-risk
// operational actions that require a selected target and an audit reason.
func NewDangerOperation(key, title, message, warning, confirmText, reasonField, reasonValue string, confirmIntent, cancelIntent intent.Intent, targets ...TargetItem) *Builder {
	builder := NewBuilder().
		Key(key).
		Title(title).
		Message(message).
		Warning(warning).
		Open(true).
		Danger(confirmText).
		AuditReason(reasonField, reasonValue).
		OnConfirm(confirmIntent).
		OnCancel(cancelIntent)
	for _, target := range targets {
		builder.Target(target)
	}
	return builder
}

// Danger applies the danger button variant and optional command text.
func (b *Builder) Danger(confirmText string) *Builder {
	b.ConfirmVariant(button.VariantDanger)
	if strings.TrimSpace(confirmText) != "" {
		b.ConfirmText(confirmText)
	}
	return b
}

// AuditReason configures the standard required audit reason field.
func (b *Builder) AuditReason(field, value string) *Builder {
	return b.AuditReasonWithLabel(field, defaultAuditReasonLabel, value)
}

// AuditReasonWithLabel configures a required audit reason field with a custom label.
func (b *Builder) AuditReasonWithLabel(field, label, value string) *Builder {
	field = strings.TrimSpace(field)
	if field == "" {
		field = defaultAuditReasonField
	}
	label = strings.TrimSpace(label)
	if label == "" {
		label = defaultAuditReasonLabel
	}
	return b.
		ReasonLabel(label).
		ReasonField(field).
		ReasonValue(value).
		ReasonRequired(true).
		ReasonPlaceholder(defaultAuditReasonPlaceholder)
}

// RequirePhrase adds a second confirmation layer by requiring the operator to
// type an exact short phrase before the confirm button can be enabled.
func (b *Builder) RequirePhrase(expected, field, value string) *Builder {
	expected = strings.TrimSpace(expected)
	field = strings.TrimSpace(field)
	if field == "" {
		field = defaultConfirmPhraseField
	}
	return b.
		ConfirmPhrase(expected, field, value).
		ConfirmPhraseLabel(defaultConfirmPhraseLabel).
		ConfirmPhrasePlaceholder(confirmPhrasePlaceholder(expected))
}

func confirmPhrasePlaceholder(expected string) string {
	if expected == "" {
		return "Type the required confirmation text."
	}
	return "Type " + expected + " to confirm."
}
