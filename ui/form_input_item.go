package ui

import (
	"github.com/wwsheng009/mint/runtime/intent"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	formcomp "github.com/wwsheng009/mint/ui/components/form"
	"github.com/wwsheng009/mint/ui/components/input"
	"github.com/wwsheng009/mint/ui/components/validation"
)

// FormInputItemOption configures a field-bound input wrapped in a FormItem.
type FormInputItemOption func(*FormInputItemConfig)

// FormInputItemConfig stores configuration for FormInputItem helpers.
type FormInputItemConfig struct {
	Placeholder string
	Width       int
	Password    bool
	Search      bool
	InputType   input.Type
	HasType     bool
	Disabled    *bool
	ReadOnly    *bool
	MaxLen      int
	FormID      string
	Layout      formcomp.FormLayout
	Validators  []validation.Validator
}

// FormInputItem builds a FormItem containing a field-bound Input.
func FormInputItem(field, label, value string, opts ...FormInputItemOption) rtui.VNode {
	cfg := FormInputItemConfig{}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}

	inputBuilder := input.NewBuilder().
		Value(value).
		ForField(intent.BindField(field))
	if cfg.Placeholder != "" {
		inputBuilder.Placeholder(cfg.Placeholder)
	}
	if cfg.Width > 0 {
		inputBuilder.Width(cfg.Width)
	}
	if cfg.HasType {
		inputBuilder.Type(cfg.InputType)
	}
	if cfg.Password {
		inputBuilder.Password()
	}
	if cfg.Search {
		inputBuilder.Search()
	}
	if cfg.Disabled != nil {
		inputBuilder.Disabled(*cfg.Disabled)
	}
	if cfg.ReadOnly != nil {
		inputBuilder.ReadOnly(*cfg.ReadOnly)
	}
	if cfg.MaxLen > 0 {
		inputBuilder.MaxLen(cfg.MaxLen)
	}
	if cfg.FormID != "" {
		inputBuilder.ForForm(intent.BindForm(cfg.FormID))
	}

	itemBuilder := formcomp.NewItem(field, inputBuilder.Build()).Label(label)
	if cfg.FormID != "" {
		itemBuilder.ForForm(cfg.FormID)
	}
	if cfg.Layout != "" {
		itemBuilder.Layout(cfg.Layout)
	}
	if len(cfg.Validators) > 0 {
		itemBuilder.Validators(cfg.Validators...)
	}

	return itemBuilder.Build()
}

// FormPasswordItem builds a FormItem containing a password Input.
func FormPasswordItem(field, label, value string, opts ...FormInputItemOption) rtui.VNode {
	all := append([]FormInputItemOption{FormInputPassword()}, opts...)
	return FormInputItem(field, label, value, all...)
}

// FormSearchItem builds a FormItem containing a search Input.
func FormSearchItem(field, label, value string, opts ...FormInputItemOption) rtui.VNode {
	all := append([]FormInputItemOption{FormInputSearch()}, opts...)
	return FormInputItem(field, label, value, all...)
}

// FormInputPlaceholder sets the input placeholder.
func FormInputPlaceholder(text string) FormInputItemOption {
	return func(cfg *FormInputItemConfig) {
		cfg.Placeholder = text
	}
}

// FormInputWidth sets the input width.
func FormInputWidth(width int) FormInputItemOption {
	return func(cfg *FormInputItemConfig) {
		cfg.Width = width
	}
}

// FormInputPassword sets the input type to password.
func FormInputPassword() FormInputItemOption {
	return func(cfg *FormInputItemConfig) {
		cfg.Password = true
	}
}

// FormInputSearch enables the search input variant.
func FormInputSearch() FormInputItemOption {
	return func(cfg *FormInputItemConfig) {
		cfg.Search = true
	}
}

// FormInputType sets the input type.
func FormInputType(inputType InputType) FormInputItemOption {
	return func(cfg *FormInputItemConfig) {
		cfg.InputType = input.Type(inputType)
		cfg.HasType = true
	}
}

// FormInputDisabled sets the input disabled state.
func FormInputDisabled(disabled bool) FormInputItemOption {
	return func(cfg *FormInputItemConfig) {
		cfg.Disabled = &disabled
	}
}

// FormInputReadOnly sets the input read-only state.
func FormInputReadOnly(readOnly bool) FormInputItemOption {
	return func(cfg *FormInputItemConfig) {
		cfg.ReadOnly = &readOnly
	}
}

// FormInputMaxLen sets the maximum input length.
func FormInputMaxLen(maxLen int) FormInputItemOption {
	return func(cfg *FormInputItemConfig) {
		cfg.MaxLen = maxLen
	}
}

// FormInputForForm explicitly associates the item with a form.
func FormInputForForm(formID string) FormInputItemOption {
	return func(cfg *FormInputItemConfig) {
		cfg.FormID = formID
	}
}

// FormInputLayout overrides the form item layout.
func FormInputLayout(layout FormLayout) FormInputItemOption {
	return func(cfg *FormInputItemConfig) {
		cfg.Layout = formcomp.FormLayout(layout)
	}
}

// FormInputValidators registers validators for the field.
func FormInputValidators(validators ...Validator) FormInputItemOption {
	return func(cfg *FormInputItemConfig) {
		cfg.Validators = append([]validation.Validator(nil), validators...)
	}
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
