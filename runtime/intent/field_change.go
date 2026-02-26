// Package intent provides type-safe Intent system for declarative UI actions.
package intent

import (
	"fmt"
	"strconv"
)

// FieldChangeIntent represents a field value change from form components.
// This intent carries both the field identifier and the runtime value from user input.
//
// This is the MVP Intent to replace the dual-Intent approach (UpdateField + SyncValue).
// It ensures:
//   - Single intent dispatch (no ordering dependencies)
//   - State becomes the single source of truth
//   - Instance is only a temporary buffer for input
//
// Example (string key):
//
//	intent := FieldChangeIntent{
//	    Field: "username",
//	    Value: "user typed text",
//	}
//	emitter.Emit(intent)
//
// Example (type-safe with StateKey[T]):
//
//	var username = StateKey[string]("username")
//	intent := intent.NewFieldChange(username, "user typed text")
//	emitter.Emit(intent)
//
// Handler:
//
//	RegisterIntent(func(ctx *ActionContext, i FieldChangeIntent) IntentResult {
//	    ctx.SetState(i.Field, i.Value)  // State becomes authority
//	    ctx.ScheduleUpdate()
//	    return HandledResult()
//	})
type FieldChangeIntent struct {
	// Field is the state key identifier for this field.
	Field string

	// Value is the runtime value from user input.
	// For Input: the text content
	// For Checkbox: boolean checked state
	// For Select: selected value
	Value string
}

// IntentType implements Intent interface.
func (FieldChangeIntent) IntentType() string {
	return "FieldChange"
}

// Priority implements PriorityAware interface.
// Field changes are high priority to ensure responsive UI.
func (FieldChangeIntent) Priority() ActionPriority {
	return PriorityUserBlocking
}

// GetField implements FieldIntent interface.
func (i FieldChangeIntent) GetField() string {
	return i.Field
}

// =============================================================================
// Generic Constructor Helpers
// =============================================================================

// NewFieldChange creates a FieldChangeIntent from a StateKey[T] and value.
// This is the type-safe way to create field change intents.
//
// Example:
//
//	var username = StateKey[string]("username")
//	intent := NewFieldChange(username, "john")
//
// Support for common types:
//   - string: value used directly
//   - bool: value converted to "true"/"false"
//   - int/int64: value converted using strconv.FormatInt
//   - float64: value converted using strconv.FormatFloat
func NewFieldChange[T any](key StateKey[T], value T) FieldChangeIntent {
	return FieldChangeIntent{
		Field: key.String(),
		Value: convertToString(value),
	}
}

// convertToString converts any type to string for Intent transport.
// FieldChangeIntent.Value is always string to maintain compatibility with
// existing Instance implementations.
func convertToString[T any](value T) string {
	switch v := any(value).(type) {
	case string:
		return v
	case bool:
		if v {
			return "true"
		}
		return "false"
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	default:
		// Fallback to simple string conversion
		return fmt.Sprintf("%v", v)
	}
}
