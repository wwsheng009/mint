package form

import (
	"github.com/wwsheng009/mint/runtime/intent"
)

// =============================================================================
// Form Intents (Phase 6: Intent Bubble Migration)
// =============================================================================

// FormFieldChangeIntent is emitted when a field value changes in a Form.
// This intent bubbles up to the Form container for validation and state management.
type FormFieldChangeIntent struct {
	// FormID is the identifier of the form this field belongs to.
	FormID string

	// Field is the field name (e.g., "username", "email").
	Field string

	// Value is the new field value.
	Value interface{}

	// IsDirty indicates if the field value has changed since last submit.
	IsDirty bool
}

// IntentType implements the intent.Intent interface.
func (FormFieldChangeIntent) IntentType() string {
	return "Form:FieldChange"
}

// Priority implements the PriorityAware interface.
func (FormFieldChangeIntent) Priority() intent.ActionPriority {
	return intent.PriorityNormal
}

// IsTransition indicates this is NOT an async operation.
func (FormFieldChangeIntent) IsTransition() bool {
	return false
}

// IsGlobal implements intent.GlobalIntent.
// FormFieldChangeIntent bubbles locally through Parent() chain.
// Returns false to indicate local Intent Bubble behavior.
func (FormFieldChangeIntent) IsGlobal() bool {
	return false
}

// FormFieldBlurIntent is emitted when a field loses focus.
// This is typically used to trigger field validation.
type FormFieldBlurIntent struct {
	// FormID is the identifier of the form this field belongs to.
	FormID string

	// Field is the field name.
	Field string

	// Value is the current field value at the time of blur.
	Value interface{}
}

// IntentType implements the intent.Intent interface.
func (FormFieldBlurIntent) IntentType() string {
	return "Form:FieldBlur"
}

// Priority implements the PriorityAware interface.
func (FormFieldBlurIntent) Priority() intent.ActionPriority {
	return intent.PriorityNormal
}

// IsTransition indicates this is NOT an async operation.
func (FormFieldBlurIntent) IsTransition() bool {
	return false
}

// IsGlobal implements intent.GlobalIntent.
// FormFieldBlurIntent bubbles locally through Parent() chain.
// Returns false to indicate local Intent Bubble behavior.
func (FormFieldBlurIntent) IsGlobal() bool {
	return false
}

// FormValidateIntent is emitted to validate a form or specific field.
type FormValidateIntent struct {
	// FormID is the identifier of the form to validate.
	FormID string

	// Field is optional. If provided, only this field is validated.
	// If empty, the entire form is validated.
	Field string
}

// IntentType implements the intent.Intent interface.
func (FormValidateIntent) IntentType() string {
	return "Form:Validate"
}

// Priority implements the PriorityAware interface.
func (FormValidateIntent) Priority() intent.ActionPriority {
	return intent.PriorityUserBlocking
}

// IsTransition indicates this is NOT an async operation.
func (FormValidateIntent) IsTransition() bool {
	return false
}

// IsGlobal implements intent.GlobalIntent.
// FormValidateIntent bubbles locally through Parent() chain.
// Returns false to indicate local Intent Bubble behavior.
func (FormValidateIntent) IsGlobal() bool {
	return false
}

// FormSubmitIntent is emitted when the user submits the form.
// The form should validate all fields before proceeding with submission.
type FormSubmitIntent struct {
	// FormID is the identifier of the form being submitted.
	FormID string

	// Data is the form data to be submitted.
	Data map[string]interface{}
}

// IntentType implements the intent.Intent interface.
func (FormSubmitIntent) IntentType() string {
	return "Form:Submit"
}

// Priority implements the PriorityAware interface.
func (i FormSubmitIntent) Priority() intent.ActionPriority {
	return intent.PriorityUserBlocking
}

// IsTransition indicates this could be an async operation (e.g., HTTP request).
func (FormSubmitIntent) IsTransition() bool {
	return true
}

// IsGlobal implements intent.GlobalIntent.
// FormSubmitIntent bubbles locally through Parent() chain.
// Returns false to indicate local Intent Bubble behavior.
func (FormSubmitIntent) IsGlobal() bool {
	return false
}

// FormResetIntent is emitted to reset the form to its initial state.
type FormResetIntent struct {
	// FormID is the identifier of the form to reset.
	FormID string
}

// IntentType implements the intent.Intent interface.
func (FormResetIntent) IntentType() string {
	return "Form:Reset"
}

// Priority implements the PriorityAware interface.
func (FormResetIntent) Priority() intent.ActionPriority {
	return intent.PriorityNormal
}

// IsTransition indicates this is NOT an async operation.
func (FormResetIntent) IsTransition() bool {
	return false
}

// IsGlobal implements intent.GlobalIntent.
// FormResetIntent bubbles locally through Parent() chain.
// Returns false to indicate local Intent Bubble behavior.
func (FormResetIntent) IsGlobal() bool {
	return false
}
