package forminputitem

import (
	"github.com/wwsheng009/mint/runtime/intent"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	formcomp "github.com/wwsheng009/mint/ui/components/form"
	"github.com/wwsheng009/mint/ui/components/input"
	"github.com/wwsheng009/mint/ui/components/validation"
)

const (
	defaultOperationReasonLabel       = "Reason"
	defaultOperationReasonPlaceholder = "maintenance window, incident response, provider recovery"
	defaultOperationReasonHelp        = "Required before preparing any operation; shown in the confirmation dialog and sent where the API supports it."
)

// Option configures a field-bound input wrapped in a FormItem.
type Option func(*Config)

// Config stores configuration for FormInputItem helpers.
type Config struct {
	Placeholder string
	Width       int
	LabelWidth  int
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
	Help        string
	Required    bool
}

// New builds a FormItem containing a field-bound Input.
func New(field, label, value string, opts ...Option) rtui.VNode {
	cfg := Config{}
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
	if cfg.LabelWidth > 0 {
		itemBuilder.LabelWidth(cfg.LabelWidth)
	}
	if cfg.Help != "" {
		itemBuilder.Help(cfg.Help)
	}
	if cfg.Required {
		itemBuilder.Required(true)
	}
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

// Password builds a FormItem containing a password Input.
func Password(field, label, value string, opts ...Option) rtui.VNode {
	all := append([]Option{InputPassword()}, opts...)
	return New(field, label, value, all...)
}

// Search builds a FormItem containing a search Input.
func Search(field, label, value string, opts ...Option) rtui.VNode {
	all := append([]Option{InputSearch()}, opts...)
	return New(field, label, value, all...)
}

// OperationReason builds a standard required audit reason input for operational actions.
func OperationReason(field, value string, opts ...Option) rtui.VNode {
	all := append([]Option{
		Placeholder(defaultOperationReasonPlaceholder),
		Required(),
		Help(defaultOperationReasonHelp),
	}, opts...)
	return New(field, defaultOperationReasonLabel, value, all...)
}

// Placeholder sets the input placeholder.
func Placeholder(text string) Option {
	return func(cfg *Config) {
		cfg.Placeholder = text
	}
}

// Width sets the input width.
func Width(width int) Option {
	return func(cfg *Config) {
		cfg.Width = width
	}
}

// LabelWidth sets a fixed display width for the FormItem label.
func LabelWidth(width int) Option {
	return func(cfg *Config) {
		cfg.LabelWidth = width
	}
}

// Horizontal uses horizontal FormItem layout with an optional fixed label width.
func Horizontal(labelWidth int) Option {
	return func(cfg *Config) {
		cfg.Layout = formcomp.LayoutHorizontal
		if labelWidth > 0 {
			cfg.LabelWidth = labelWidth
		}
	}
}

// InputPassword sets the input type to password.
func InputPassword() Option {
	return func(cfg *Config) {
		cfg.Password = true
	}
}

// InputSearch enables the search input variant.
func InputSearch() Option {
	return func(cfg *Config) {
		cfg.Search = true
	}
}

// InputType sets the input type.
func InputType(inputType input.Type) Option {
	return func(cfg *Config) {
		cfg.InputType = inputType
		cfg.HasType = true
	}
}

// Disabled sets the input disabled state.
func Disabled(disabled bool) Option {
	return func(cfg *Config) {
		cfg.Disabled = &disabled
	}
}

// ReadOnly sets the input read-only state.
func ReadOnly(readOnly bool) Option {
	return func(cfg *Config) {
		cfg.ReadOnly = &readOnly
	}
}

// MaxLen sets the maximum input length.
func MaxLen(maxLen int) Option {
	return func(cfg *Config) {
		cfg.MaxLen = maxLen
	}
}

// ForForm explicitly associates the item with a form.
func ForForm(formID string) Option {
	return func(cfg *Config) {
		cfg.FormID = formID
	}
}

// Layout overrides the form item layout.
func Layout(layout formcomp.FormLayout) Option {
	return func(cfg *Config) {
		cfg.Layout = layout
	}
}

// Validators registers validators for the field.
func Validators(validators ...validation.Validator) Option {
	return func(cfg *Config) {
		cfg.Validators = append([]validation.Validator(nil), validators...)
	}
}

// Help sets helper text for the field.
func Help(text string) Option {
	return func(cfg *Config) {
		cfg.Help = text
	}
}

// Required marks the field label as required.
func Required() Option {
	return func(cfg *Config) {
		cfg.Required = true
	}
}
