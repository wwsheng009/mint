package ui

import (
	rtui "github.com/wwsheng009/mint/runtime/ui"
	formcomp "github.com/wwsheng009/mint/ui/components/form"
	"github.com/wwsheng009/mint/ui/components/forminputitem"
	"github.com/wwsheng009/mint/ui/components/input"
	"github.com/wwsheng009/mint/ui/components/validation"
)

// FormInputItemOption configures a field-bound input wrapped in a FormItem.
type FormInputItemOption = forminputitem.Option

// FormInputItemConfig stores configuration for FormInputItem helpers.
type FormInputItemConfig = forminputitem.Config

// FormInputItem builds a FormItem containing a field-bound Input.
func FormInputItem(field, label, value string, opts ...FormInputItemOption) rtui.VNode {
	return forminputitem.New(field, label, value, opts...)
}

// FormPasswordItem builds a FormItem containing a password Input.
func FormPasswordItem(field, label, value string, opts ...FormInputItemOption) rtui.VNode {
	return forminputitem.Password(field, label, value, opts...)
}

// FormSearchItem builds a FormItem containing a search Input.
func FormSearchItem(field, label, value string, opts ...FormInputItemOption) rtui.VNode {
	return forminputitem.Search(field, label, value, opts...)
}

// FormInputPlaceholder sets the input placeholder.
func FormInputPlaceholder(text string) FormInputItemOption {
	return forminputitem.Placeholder(text)
}

// FormInputWidth sets the input width.
func FormInputWidth(width int) FormInputItemOption {
	return forminputitem.Width(width)
}

// FormInputPassword sets the input type to password.
func FormInputPassword() FormInputItemOption {
	return forminputitem.InputPassword()
}

// FormInputSearch enables the search input variant.
func FormInputSearch() FormInputItemOption {
	return forminputitem.InputSearch()
}

// FormInputType sets the input type.
func FormInputType(inputType InputType) FormInputItemOption {
	return forminputitem.InputType(input.Type(inputType))
}

// FormInputDisabled sets the input disabled state.
func FormInputDisabled(disabled bool) FormInputItemOption {
	return forminputitem.Disabled(disabled)
}

// FormInputReadOnly sets the input read-only state.
func FormInputReadOnly(readOnly bool) FormInputItemOption {
	return forminputitem.ReadOnly(readOnly)
}

// FormInputMaxLen sets the maximum input length.
func FormInputMaxLen(maxLen int) FormInputItemOption {
	return forminputitem.MaxLen(maxLen)
}

// FormInputForForm explicitly associates the item with a form.
func FormInputForForm(formID string) FormInputItemOption {
	return forminputitem.ForForm(formID)
}

// FormInputLayout overrides the form item layout.
func FormInputLayout(layout FormLayout) FormInputItemOption {
	return forminputitem.Layout(formcomp.FormLayout(layout))
}

// FormInputValidators registers validators for the field.
func FormInputValidators(validators ...Validator) FormInputItemOption {
	return forminputitem.Validators(validators...)
}

// FormInputHelp sets helper text for the field.
func FormInputHelp(text string) FormInputItemOption {
	return forminputitem.Help(text)
}

// FormInputRequired marks the field label as required.
func FormInputRequired() FormInputItemOption {
	return forminputitem.Required()
}

// Validator is the public validation contract used by form helpers.
type Validator = validation.Validator

// Required creates a required field validator.
func Required() Validator {
	return validation.Required()
}

// MinLength creates a minimum length validator.
func MinLength(min int) Validator {
	return validation.MinLength(min)
}

// MaxLength creates a maximum length validator.
func MaxLength(max int) Validator {
	return validation.MaxLength(max)
}

// Min creates a minimum numeric value validator.
func Min(min float64) Validator {
	return validation.Min(min)
}

// Max creates a maximum numeric value validator.
func Max(max float64) Validator {
	return validation.Max(max)
}

// Range creates a numeric range validator.
func Range(min, max float64) Validator {
	return validation.Range(min, max)
}

// Pattern creates a regular expression validator.
func Pattern(pattern string) Validator {
	return validation.Pattern(pattern)
}

// Email creates an email validator.
func Email() Validator {
	return validation.Email()
}

// URL creates a URL validator.
func URL() Validator {
	return validation.URL()
}

// OneOf creates an enum validator.
func OneOf(values ...interface{}) Validator {
	return validation.OneOf(values...)
}
