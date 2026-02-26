// Package intent provides type-safe Intent system for declarative UI actions.
package intent

// =============================================================================
// Type-Safe State Keys
// =============================================================================

// StateKey[T] is a type-safe wrapper for state keys.
// It provides compile-time type safety and eliminates string key typos.
//
// Usage:
//
//	// Define keys (typically at package level)
//	var (
//		username = intent.StateKey[string]("username")
//		email    = intent.StateKey[string]("email")
//		agree    = intent.StateKey[bool]("agree")
//		checked  = intent.StateKey[int]("checked")
//	)
//
//	// Use in FieldChangeIntent
//	intent := intent.FieldChangeIntent{
//	    Field: username.String(),  // or use FieldKey extension
//	    Value: "john",
//	}
//
//	// Get field from Binding
//	binding := intent.ForField(username)
//
// This design offers:
//   - Compile-time type checking: StateKey[string] vs StateKey[int]
//   - IDE autocomplete: package-level keys
//   - Refactor safety: rename key updates all references
type StateKey[T any] string

// String returns the underlying string key.
func (k StateKey[T]) String() string {
	return string(k)
}

// Type returns a string representation of the type (for debugging).
func (k StateKey[T]) Type() string {
	var zero T
	typeName := getTypeName(zero)
	return typeName
}

// FieldIntent implements FieldIntent interface.
func (k StateKey[T]) GetField() string {
	return string(k)
}

// IntentType implements Intent interface.
// StateKey represents a field identifier, not a dispatchable intent.
func (k StateKey[T]) IntentType() string {
	return "StateKey"
}

// =============================================================================
// Helper Functions
// =============================================================================

// ForField creates a FieldBinding from a StateKey[T].
// This is a convenience method for component builders.
//
// Example:
//
//	var username = intent.StateKey[string]("username")
//	input.NewBuilder().ForField(intent.ForField(username))
func ForField[T any](key StateKey[T]) FieldBinding {
	return FieldBinding(key.String())
}

// =============================================================================
// Reflection Helper
// =============================================================================

// getTypeName returns the type name of a value using reflection.
// This is used for Type() method's debug output.
func getTypeName(v interface{}) string {
	// Simple type name extraction
	switch v.(type) {
	case string:
		return "string"
	case bool:
		return "bool"
	case int:
		return "int"
	case int64:
		return "int64"
	case float64:
		return "float64"
	default:
		// For complex types, use reflect if available
		// Note: We use a simple switch to avoid reflect dependency
		return "any"
	}
}

// =============================================================================
// Predefined Common Keys
// =============================================================================

// Common state keys for form fields.
// Users can define their own package-level keys.

var (
	// String keys
	KeyUsername  = StateKey[string]("username")
	KeyEmail     = StateKey[string]("email")
	KeyPassword  = StateKey[string]("password")
	KeyText      = StateKey[string]("text")
	KeySearch    = StateKey[string]("search")
	KeyMessage   = StateKey[string]("message")

	// Boolean keys
	KeyAgree     = StateKey[bool]("agree")
	KeyChecked   = StateKey[bool]("checked")
	KeyEnabled   = StateKey[bool]("enabled")
	KeyVisible   = StateKey[bool]("visible")
	KeySelected  = StateKey[bool]("selected")
	KeyLoading   = StateKey[bool]("loading")
	KeyDisabled  = StateKey[bool]("disabled")
	KeyReadOnly  = StateKey[bool]("readOnly")

	// Numeric keys
	KeyCount     = StateKey[int]("count")
	KeyIndex     = StateKey[int]("index")
	KeyPage      = StateKey[int]("page")
	KeyOffset    = StateKey[int]("offset")
	KeyLimit     = StateKey[int]("limit")
)
