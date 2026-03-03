// Package validation provides form validation for TUI components.
package validation

import (
	"fmt"
)

// =============================================================================
// Validator Interface
// =============================================================================

// Validator is the interface for all validators.
type Validator interface {
	// Validate validates the given value.
	Validate(value interface{}) error

	// Message returns the error message.
	Message() string

	// WithMessage sets a custom error message.
	WithMessage(msg string) Validator
}

// ValidatorFunc is a function type for validation.
type ValidatorFunc func(value interface{}) error

// =============================================================================
// FuncValidator
// =============================================================================

// FuncValidator wraps a validation function.
type FuncValidator struct {
	fn      ValidatorFunc
	message string
}

// NewFuncValidator creates a new function validator.
func NewFuncValidator(fn ValidatorFunc, message string) *FuncValidator {
	return &FuncValidator{
		fn:      fn,
		message: message,
	}
}

// Validate validates the value using the wrapped function.
func (v *FuncValidator) Validate(value interface{}) error {
	if err := v.fn(value); err != nil {
		if v.message != "" {
			return fmt.Errorf("%s: %w", v.message, err)
		}
		return err
	}
	return nil
}

// Message returns the error message.
func (v *FuncValidator) Message() string {
	return v.message
}

// WithMessage sets a custom error message.
func (v *FuncValidator) WithMessage(msg string) Validator {
	v.message = msg
	return v
}

// =============================================================================
// Composite Validator
// =============================================================================

// CompositeMode defines how multiple validators are combined.
type CompositeMode int

const (
	// ModeAll requires all validators to pass (AND logic).
	ModeAll CompositeMode = iota

	// ModeAny requires at least one validator to pass (OR logic).
	ModeAny
)

// CompositeValidator combines multiple validators.
type CompositeValidator struct {
	validators []Validator
	message    string
	mode       CompositeMode
}

// NewAllValidator creates an AND validator.
func NewAllValidator(validators ...Validator) *CompositeValidator {
	return &CompositeValidator{
		validators: validators,
		mode:       ModeAll,
	}
}

// NewAnyValidator creates an OR validator.
func NewAnyValidator(validators ...Validator) *CompositeValidator {
	return &CompositeValidator{
		validators: validators,
		mode:       ModeAny,
	}
}

// Validate runs all validators according to the composite mode.
func (v *CompositeValidator) Validate(value interface{}) error {
	var errors []error

	for _, validator := range v.validators {
		err := validator.Validate(value)
		if v.mode == ModeAll {
			if err != nil {
				return err
			}
		} else { // ModeAny
			if err == nil {
				return nil
			}
			errors = append(errors, err)
		}
	}

	if v.mode == ModeAny && len(errors) > 0 {
		return fmt.Errorf("none of the validators passed")
	}

	return nil
}

// Message returns the error message.
func (v *CompositeValidator) Message() string {
	return v.message
}

// WithMessage sets a custom error message.
func (v *CompositeValidator) WithMessage(msg string) Validator {
	v.message = msg
	return v
}

// =============================================================================
// Validator Chain (Fluent API)
// =============================================================================

// Chain creates a validator chain for fluent validation.
type Chain struct {
	validators []Validator
}

// NewChain creates a new validator chain.
func NewChain() *Chain {
	return &Chain{}
}

// Add adds a validator to the chain.
func (c *Chain) Add(v Validator) *Chain {
	c.validators = append(c.validators, v)
	return c
}

// Required adds a required validator.
func (c *Chain) Required() *Chain {
	return c.Add(Required())
}

// MinLength adds a min length validator.
func (c *Chain) MinLength(min int) *Chain {
	return c.Add(MinLength(min))
}

// MaxLength adds a max length validator.
func (c *Chain) MaxLength(max int) *Chain {
	return c.Add(MaxLength(max))
}

// Email adds an email validator.
func (c *Chain) Email() *Chain {
	return c.Add(Email())
}

// URL adds a URL validator.
func (c *Chain) URL() *Chain {
	return c.Add(URL())
}

// Pattern adds a pattern validator.
func (c *Chain) Pattern(pattern string) *Chain {
	return c.Add(Pattern(pattern))
}

// Validate validates the value against all validators in the chain.
func (c *Chain) Validate(value interface{}) error {
	for _, v := range c.validators {
		if err := v.Validate(value); err != nil {
			return err
		}
	}
	return nil
}

// Build creates a composite validator from the chain.
func (c *Chain) Build() Validator {
	return NewAllValidator(c.validators...)
}
