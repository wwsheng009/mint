package form

import (
	"github.com/wwsheng009/mint/runtime/intent"
	rtui "github.com/wwsheng009/mint/runtime/ui"
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

// BindForm creates a FormBinding for a form.
// This can be passed to field components' ForForm() method.
// Example:
//
//	form.BindForm("loginForm")
func BindForm(formID string) intent.FormBinding {
	return intent.BindForm(formID)
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
		FormID:  formID,
		Field:   field,
		Value:   value,
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

// =============================================================================
// Context Support (Phase 2)
// =============================================================================

// ProvideForm is a compatibility helper that attaches children to a Form VNode.
// Form runtime linkage is currently resolved by explicit formID plus ancestor
// Form lookup at render time; this helper remains for call-site readability.
//
// Example:
//
//	formBuilder := form.NewForm("loginForm")
//	formBuilder.Label("Login Form")
//
//	selectField := selectComp.NewBuilder()
//	selectField.SetKey("username")
//	selectField.ForForm(form.BindForm("loginForm"))
//	selectField.SetOptions([]selectComp.Option{
//		{Value: "user1", Label: "User 1"},
//		{Value: "user2", Label: "User 2"},
//	})
//
//	formWithChildren := form.ProvideForm(formBuilder, selectField)
func ProvideForm(formBuilder Builder, children ...rtui.VNode) rtui.VNode {
	formBuilder.AddChildren(children...)
	return formBuilder
}

// BuildWithConfig builds a Form with explicit configuration.
func BuildWithConfig(key string, opts ...FormConfigOption) rtui.VNode {
	builder := NewForm(key)
	for _, opt := range opts {
		builder = opt(builder)
	}
	return builder
}

// FormConfigOption is a functional option for configuring forms
type FormConfigOption func(Builder) Builder

// WithFormLabel sets the form label
func WithFormLabel(label string) FormConfigOption {
	return func(b Builder) Builder {
		return b.Label(label)
	}
}

// WithFormValues sets the initial field values
func WithFormValues(values map[string]interface{}) FormConfigOption {
	return func(b Builder) Builder {
		return b.SetValues(values)
	}
}

// WithFormValidation enables/disables validation
func WithFormValidation(validate bool) FormConfigOption {
	return func(b Builder) Builder {
		return b.ValidateAll(validate)
	}
}

// WithFormLayout sets the default layout for nested FormItems.
func WithFormLayout(layout FormLayout) FormConfigOption {
	return func(b Builder) Builder {
		return b.Layout(layout)
	}
}

// WithFormSubmit sets the submit handler
func WithFormSubmit(handler intent.Intent) FormConfigOption {
	return func(b Builder) Builder {
		return b.OnSubmit(handler)
	}
}

// WithFormReset sets the reset handler
func WithFormReset(handler intent.Intent) FormConfigOption {
	return func(b Builder) Builder {
		return b.OnReset(handler)
	}
}
