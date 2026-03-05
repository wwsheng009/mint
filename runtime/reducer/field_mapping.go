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

// FieldMap maps field names to update functions.
// This allows binding multiple fields without duplicate FieldChangeIntent handlers.
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
