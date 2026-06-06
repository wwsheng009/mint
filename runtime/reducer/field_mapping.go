// Package reducer provides field mapping utilities for Store + Reducer architecture.
//
// This module helps reduce hardcoding when processing FieldChangeIntent
// by providing type-safe field binding helpers.
package reducer

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/wwsheng009/mint/runtime/intent"
)

// =============================================================================
// Field Mapping Helper
// =============================================================================

// FieldBinder helps bind fields to state without manual switch statements.
// Example:
//
//	appReducer := Reducer[AppState]().
//		BindStringField(&AppState{}.Username, "username").
//		BindIntField(&AppState{}.Age, "age").
//		BindBoolField(&AppState{}.Agreed, "agreed").
//		Build()
type FieldBinder[T any] struct {
	*Builder[T]
}

// BindField creates a field binder from a reducer builder.
func BindField[T any](b *Builder[T]) *FieldBinder[T] {
	return &FieldBinder[T]{Builder: b}
}

// =============================================================================
// String Field Binding
// =============================================================================

// BindStringField binds a string field to FieldChangeIntent.
//
// Example:
//
//	BindField(appReducer).
//		BindStringField("username", func(s AppState, val string) AppState {
//			s.Username = val
//			return s
//		})
func (fb *FieldBinder[T]) BindStringField(
	fieldName string,
	update func(T, string) T,
) *FieldBinder[T] {
	fb.On(intent.FieldChangeIntent{}, func(s T, i intent.Intent) T {
		fci, ok := i.(intent.FieldChangeIntent)
		if !ok || fci.Field != fieldName {
			return s
		}
		return update(s, fci.Value)
	})
	return fb
}

// =============================================================================
// Bool Field Binding
// =============================================================================

// BindBoolField binds a boolean field to FieldChangeIntent.
//
// Example:
//
//	BindField(appReducer).
//		BindBoolField("agreed", func(s AppState, val bool) AppState {
//			s.Agreed = val
//			return s
//		})
func (fb *FieldBinder[T]) BindBoolField(
	fieldName string,
	update func(T, bool) T,
) *FieldBinder[T] {
	fb.On(intent.FieldChangeIntent{}, func(s T, i intent.Intent) T {
		fci, ok := i.(intent.FieldChangeIntent)
		if !ok || fci.Field != fieldName {
			return s
		}
		// Convert "true"/"false" string to bool
		boolVal := strings.ToLower(fci.Value) == "true"
		return update(s, boolVal)
	})
	return fb
}

// =============================================================================
// Int Field Binding
// =============================================================================

// BindIntField binds an integer field to FieldChangeIntent.
//
// Example:
//
//	BindField(appReducer).
//		BindIntField("age", func(s AppState, val int) AppState {
//			s.Age = val
//			return s
//		})
func (fb *FieldBinder[T]) BindIntField(
	fieldName string,
	update func(T, int) T,
) *FieldBinder[T] {
	fb.On(intent.FieldChangeIntent{}, func(s T, i intent.Intent) T {
		fci, ok := i.(intent.FieldChangeIntent)
		if !ok || fci.Field != fieldName {
			return s
		}
		// Convert string to int
		intVal, err := strconv.Atoi(fci.Value)
		if err != nil {
			// Keep original value if conversion fails
			return s
		}
		return update(s, intVal)
	})
	return fb
}

// =============================================================================
// Float Field Binding
// =============================================================================

// BindFloatField binds a float64 field to FieldChangeIntent.
//
// Example:
//
//	BindField(appReducer).
//		BindFloatField("price", func(s AppState, val float64) AppState {
//			s.Price = val
//			return s
//		})
func (fb *FieldBinder[T]) BindFloatField(
	fieldName string,
	update func(T, float64) T,
) *FieldBinder[T] {
	fb.On(intent.FieldChangeIntent{}, func(s T, i intent.Intent) T {
		fci, ok := i.(intent.FieldChangeIntent)
		if !ok || fci.Field != fieldName {
			return s
		}
		// Convert string to float64
		floatVal, err := strconv.ParseFloat(fci.Value, 64)
		if err != nil {
			// Keep original value if conversion fails
			return s
		}
		return update(s, floatVal)
	})
	return fb
}

// =============================================================================
// Generic Field Binding
// =============================================================================

// BindFieldGeneric binds a field with custom parsing logic.
//
// Example:
//
//	BindField(appReducer).
//		BindFieldGeneric("username", func(s AppState, val string) AppState {
//			s.Username = val
//			return s
//		})
func (fb *FieldBinder[T]) BindFieldGeneric(
	fieldName string,
	update func(T, string) T,
) *FieldBinder[T] {
	fb.On(intent.FieldChangeIntent{}, func(s T, i intent.Intent) T {
		fci, ok := i.(intent.FieldChangeIntent)
		if !ok || fci.Field != fieldName {
			return s
		}
		return update(s, fci.Value)
	})
	return fb
}

// =============================================================================
// Multi-Field Binding (Using a map) / 多字段绑定（使用映射表）
// =============================================================================

// FieldValidator validates a field value before updating the state.
// Returns true if the value is valid, false otherwise.
type FieldValidator func(fieldName, value string) bool

// FieldError is returned when field validation fails.
type FieldError struct {
	Field   string
	Value   string
	Message string
}

func (e FieldError) Error() string {
	return fmt.Sprintf("field %q validation failed: %s (value: %q)", e.Field, e.Message, e.Value)
}

// ValidationError is a slice of FieldError.
type ValidationError []FieldError

func (ve ValidationError) Error() string {
	if len(ve) == 0 {
		return "no validation errors"
	}
	msg := fmt.Sprintf("validation failed: %d error(s)", len(ve))
	if len(ve) <= 3 {
		for _, e := range ve {
			msg += fmt.Sprintf("\n  - %s", e.Error())
		}
	} else {
		for i := 0; i < 3; i++ {
			msg += fmt.Sprintf("\n  - %s", ve[i].Error())
		}
		msg += fmt.Sprintf("\n  ... and %d more error(s)", len(ve)-3)
	}
	return msg
}

// FieldMap maps field names to update functions with optional validation.
// This allows binding multiple fields without duplicate FieldChangeIntent handlers.
//
// For advanced usage with validation and error handling, see FieldEntry.
//
// Example:
//
//	type AppState struct {
//		Username string
//		Email    string
//		Age      int
//	}
//
//	fieldMap := reducer.FieldMap[AppState]{
//		"username": func(s AppState, val string) AppState {
//			s.Username = val
//			return s
//		},
//		"email": func(s AppState, val string) AppState {
//			s.Email = val
//			return s
//		},
//		"age": func(s AppState, val string) AppState {
//			if v, err := strconv.Atoi(val); err == nil {
//				s.Age = v
//			}
//			return s
//		},
//	}
//
//	appReducer := BindField(reducer.NewBuilder[AppState]()).
//		BindFieldMap(fieldMap).
//		Build()
type FieldMap[T any] map[string]func(T, string) T

// FieldEntry represents a complete field binding configuration with validation and error handling.
type FieldEntry[T any] struct {
	// Updater updates the state with the field value
	Updater func(T, string) T

	// Validator validates the field value before updating
	// If nil, no validation is performed
	Validator FieldValidator

	// Required indicates if the field is required
	// If true and the value is empty, validation fails
	Required bool

	// Transform transforms the field value before validation
	// If nil, no transformation is performed
	Transform func(string) string
}

// BindFieldMap binds multiple fields using a single FieldChangeIntent handler.
//
// Example:
//
//	BindField(appReducer).
//		BindFieldMap(map[string]func(s AppState, val string) AppState{
//			"username": func(s AppState, val string) AppState {
//				s.Username = val
//				return s
//			},
//			"email": func(s AppState, val string) AppState {
//				s.Email = val
//				return s
//			},
//		})
func (fb *FieldBinder[T]) BindFieldMap(fieldMap FieldMap[T]) *FieldBinder[T] {
	fb.On(intent.FieldChangeIntent{}, func(s T, i intent.Intent) T {
		fci, ok := i.(intent.FieldChangeIntent)
		if !ok {
			return s
		}
		if updater, exists := fieldMap[fci.Field]; exists {
			return updater(s, fci.Value)
		}
		return s
	})
	return fb
}

// UpdateStringFieldIfChanged applies update only when next differs from the
// current string value. It is useful for filter/search fields where changing the
// scope should reset page or selection state, while repeated identical
// FieldChangeIntent values should be a no-op.
func UpdateStringFieldIfChanged[T any](s T, current, next string, update func(T, string) T) T {
	if current == next || update == nil {
		return s
	}
	return update(s, next)
}

// BindFieldMapWithEntries binds multiple fields with advanced validation and error handling.
//
// FieldEntry allows for:
// - Custom validation logic
// - Required field checking
// - Value transformation (e.g., trimming, case conversion)
//
// Example:
//
//	entries := map[string]*FieldEntry[AppState]{
//		"username": {
//			Updater: func(s GameState, val string) AppState {
//				s.Username = val
//				return s
//			},
//			Validator: func(field, val string) bool {
//				return len(val) >= 3
//			},
//			Required: true,
//		},
//		"email": {
//			Updater: func(s AppState, val string) AppState {
//				s.Email = strings.ToLower(val)
//				return s
//			},
//			Validator: func(field, val string) bool {
//				return strings.Contains(val, "@")
//			},
//			Transform: func(val string) string {
//				return strings.TrimSpace(val)
//			},
//		},
//	}
//
//	BindField(appReducer).BindFieldMapWithEntries(entries)
func (fb *FieldBinder[T]) BindFieldMapWithEntries(entries map[string]*FieldEntry[T]) *FieldBinder[T] {
	fb.On(intent.FieldChangeIntent{}, func(s T, i intent.Intent) T {
		fci, ok := i.(intent.FieldChangeIntent)
		if !ok {
			return s
		}

		entry, exists := entries[fci.Field]
		if !exists {
			return s
		}

		var value = fci.Value

		// Apply transformation if provided
		if entry.Transform != nil {
			value = entry.Transform(value)
		}

		// Check required fields
		if entry.Required && value == "" {
			// Could also store validation errors in state if needed
			return s
		}

		// Apply validation if provided
		if entry.Validator != nil {
			if !entry.Validator(fci.Field, value) {
				// Could also store validation errors in state if needed
				return s
			}
		}

		// Update the state
		return entry.Updater(s, value)
	})
	return fb
}

// =============================================================================
// Direct Field Binding (Reflection-based - Optional)
// =============================================================================

// BindFieldsAuto automatically binds all fields using struct tags and reflection.
// Requires struct tag: `field:"fieldname"`
//
// Example:
//
//	type AppState struct {
//		Username string `field:"username"`
//		Email    string `field:"email"`
//		Age      int    `field:"age"`
//	}
//
//	BindField(appReducer).BindFieldsAuto(&AppState{})
func (fb *FieldBinder[T]) BindFieldsAuto(statePtr *T) *FieldBinder[T] {
	// TODO: Implement reflection-based field binding
	// This would require the runtime/reflection package
	return fb
}

// =============================================================================
// Builder integration / Builder 集成
// =============================================================================

// GetBuilder returns the underlying Builder for further operations.
func (fb *FieldBinder[T]) GetBuilder() *Builder[T] {
	return fb.Builder
}

// Build completes the field binding and returns the Reducer.
func (fb *FieldBinder[T]) Build() Reducer[T] {
	return fb.Builder.Build()
}

// RegisterToGlobal registers the reducer to the global registry.
func (fb *FieldBinder[T]) RegisterToGlobal(store StoreSetter[T]) Reducer[T] {
	return fb.Builder.RegisterToGlobal(store)
}

// =============================================================================
// Conveniences Helper Functions / 便捷辅助函数
// =============================================================================

// ParseBool parses a bool from string representation.
func ParseBool(s string) bool {
	return strings.ToLower(s) == "true"
}

// ParseInt parses an int from string representation.
func ParseInt(s string) (int, error) {
	return strconv.Atoi(s)
}

// ParseFloat parses a float64 from string representation.
func ParseFloat(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

// FormatBool formats a bool to string representation.
func FormatBool(b bool) string {
	return fmt.Sprintf("%v", b)
}

// FormatInt formats an int to string representation.
func FormatInt(i int) string {
	return fmt.Sprintf("%d", i)
}
