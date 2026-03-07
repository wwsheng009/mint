package intent

// =============================================================================
// Form Intent Binding (Phase 6: Form Integration)
// =============================================================================

// FormBinding represents a form binding for form field components.
// This is used with ForForm() methods in component builders to enable
// Form-specific intent emission (FormFieldChangeIntent, FormFieldBlurIntent).
//
// Example:
//
//	inputBuilder := input.NewBuilder()
//	    .ForField(intent.BindField("username"))
//	    .ForForm(intent.BindForm("loginForm"))
//
// The Instance will emit FormFieldChangeIntent/FormFieldBlurIntent
// instead of FieldChangeIntent when the user interacts with the field.
type FormBinding struct {
	// FormID is the identifier of the form this field belongs to.
	FormID string
}

// BindForm creates a FormBinding for the given form ID.
// Use this with component builders' ForForm() methods.
//
// Example:
//
//	input.NewBuilder().ForForm(intent.BindForm("loginForm"))
//	checkbox.NewBuilder().ForForm(intent.BindForm("signupForm"))
//	selectField.NewBuilder().ForForm(intent.BindForm("settingsForm"))
func BindForm(formID string) FormBinding {
	return FormBinding{FormID: formID}
}

// IntentType implements Intent interface.
// FormBinding is not dispatched - it's a static metadata used by components.
func (FormBinding) IntentType() string {
	return "FormBinding"
}

// GetFormID returns the form identifier.
func (f FormBinding) GetFormID() string {
	return f.FormID
}
