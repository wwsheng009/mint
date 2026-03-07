package form

import (
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/runtime/intent"
)

// =============================================================================
// Builder - Form Fluent Builder
// =============================================================================

// Builder provides a fluent API for constructing Form VNodes.
// Builder is an alias for *VNode which implements the fluent API.
type Builder = *VNode

// NewBuilder creates a new Form builder.
func NewBuilder() Builder {
	return New()
}

// ============================================================
// Convenience Factory Functions
// ============================================================

// NewForm creates a form builder with the given key.
func NewForm(key string) Builder {
	return New().SetKey(key).(*VNode)
}

// WithLabel creates a form builder with a label.
func WithLabel(key string, label string) Builder {
	return NewForm(key).Label(label)
}

// ============================================================
// Field Binding Helpers
// ============================================================

// BindField creates a FieldBinding for a field.
// This can be passed to field components' ForField() method.
// Example:
//
//	form.BindField("username")
func BindField(field string) intent.FieldBinding {
	return intent.BindField(field)
}

// ============================================================
// Component Configuration Options
// ============================================================

// FieldOption is a functional option for configuring field components.
type FieldOption func(rtui.VNode) rtui.VNode

// WithPlaceholder is a field option that sets the placeholder text.
func WithPlaceholder(text string) FieldOption {
	return func(v rtui.VNode) rtui.VNode {
		if props := v.Props(); props != nil {
			props["placeholder"] = text
			v.SetProps(props)
		}
		return v
	}
}

// WithFieldLabel is a field option that sets the component label.
// Note: This is distinct from the Form's Label method.
func WithFieldLabel(text string) FieldOption {
	return func(v rtui.VNode) rtui.VNode {
		if props := v.Props(); props != nil {
			props["label"] = text
			v.SetProps(props)
		}
		return v
	}
}

// ============================================================
// Intent Creation Helpers
// ============================================================

// FieldChange creates a FormFieldChangeIntent.
// This is used by field components to emit change events.
// Example:
//
//	intent := form.FieldChange("loginForm", "username", newUserValue, true)
//	intent.Emit(formInstance)
func FieldChange(formID, field string, value interface{}, isDirty bool) FormFieldChangeIntent {
	return FormFieldChangeIntent{
		FormID: formID,
		Field:  field,
		Value:  value,
		IsDirty: isDirty,
	}
}

// FieldBlur creates a FormFieldBlurIntent.
// This is used by field components to emit blur events (for validation).
// Example:
//
//	intent := form.FieldBlur("loginForm", "username", currentUserValue)
//	intent.Emit(formInstance)
func FieldBlur(formID, field string, value interface{}) FormFieldBlurIntent {
	return FormFieldBlurIntent{
		FormID: formID,
		Field:  field,
		Value:  value,
	}
}

// Validate creates a FormValidateIntent.
// This is used to trigger validation for the form or a specific field.
// Example:
//
//	intent := form.Validate("loginForm", "")          // Validate entire form
//	intent := form.Validate("loginForm", "username") // Validate specific field
//	intent.Emit(formInstance)
func Validate(formID string, field string) FormValidateIntent {
	return FormValidateIntent{
		FormID: formID,
		Field:  field,
	}
}

// Submit creates a FormSubmitIntent with the provided data.
// Example:
//
//	intent := form.Submit("loginForm", map[string]interface{}{
//	    "username": user,
//	    "password": pass,
//	})
//	intent.Emit(formInstance)
func Submit(formID string, data map[string]interface{}) FormSubmitIntent {
	return FormSubmitIntent{
		FormID: formID,
		Data:   data,
	}
}

// Reset creates a FormResetIntent.
// Example:
//
//	intent := form.Reset("loginForm")
//	intent.Emit(formInstance)
func Reset(formID string) FormResetIntent {
	return FormResetIntent{
		FormID: formID,
	}
}
