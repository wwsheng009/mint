// Package intent provides type-safe Intent system for declarative UI actions.
package intent

// FieldBinding represents a simple field binding for form components.
// This is used with ForField() helper methods in component builders.
//
// Example:
//
//	inputBuilder := input.NewBuilder()
//	    .ForField(intent.BindField("username"))
//
// The Instance will use the Field key when emitting FieldChangeIntent.
type FieldBinding string

// BindField creates a FieldBinding for the given state key.
// Use this with component builders' ForField() methods.
//
// Example:
//
//	input.NewBuilder().ForField(intent.BindField("email"))
//	checkbox.NewBuilder().ForField(intent.BindField("acceptTerms"))
func BindField(key string) FieldBinding {
	return FieldBinding(key)
}

// IntentType implements Intent interface.
// FieldBinding is not dispatched - it's a static metadata used by components.
func (FieldBinding) IntentType() string {
	return "FieldBinding"
}

// GetField implements FieldIntent interface.
func (f FieldBinding) GetField() string {
	return string(f)
}
