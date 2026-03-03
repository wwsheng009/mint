// Package intent provides TypedFieldChange[T] for type-safe field updates.
//
// TypedFieldChange[T] eliminates string keys and provides compile-time
// type safety for form field updates and state changes.
package intent

import "fmt"

// =============================================================================
// TypedFieldChange[T] - Type-Safe Field Update Intent
// =============================================================================

// TypedFieldChange is a type-safe intent for updating a typed field.
// Use this instead of string-based FieldChangeIntent for compile-time safety.
//
// Example:
//
//	var UsernameKey = NewStateKey[string]("username")
//
//	// Create a typed field change
//	intent := UsernameKey.Change("alice")
//
//	// In Reducer:
//	switch v := intent.(type) {
//	case TypedFieldChange[string]:
//	    switch v.Key {
//	    case UsernameKey:
//	        state.Username = v.Value  // Type-safe!
//	    }
//	}
type TypedFieldChange[T any] struct {
	Key   StateKey[T]
	Value T
}

// IntentType implements the Intent interface.
func (TypedFieldChange[T]) IntentType() string {
	return "TypedFieldChange"
}

// String returns a human-readable representation.
func (t TypedFieldChange[T]) String() string {
	return fmt.Sprintf("TypedFieldChange[%T]{Key: %s, Value: %v}", t.Value, t.Key.String(), t.Value)
}

// =============================================================================
// Helper Functions
// =============================================================================

// SetField creates a TypedFieldChange intent for a state key.
// This is a convenience function for creating type-safe intents.
//
// Example:
//
//	intent := intent.SetField(Username, "alice")
func SetField[T any](key StateKey[T], value T) TypedFieldChange[T] {
	return TypedFieldChange[T]{
		Key:   key,
		Value: value,
	}
}

// UpdateField creates a TypedFieldChange intent using a key name.
// This is useful when you have the key name as a string but want type safety.
//
// Example:
//
//	intent := intent.UpdateField[string]("username", "alice")
func UpdateField[T any](keyName string, value T) TypedFieldChange[T] {
	return TypedFieldChange[T]{
		Key:   NewStateKey[T](keyName),
		Value: value,
	}
}

// =============================================================================
// TypedFieldChange with Multiple Values
// =============================================================================

// MultiFieldChange represents multiple field changes in a single intent.
// Use this when updating multiple fields atomically.
//
// Example:
//
//	intent := intent.MultiFieldChange{
//	    Username.Change("alice"),
//	    Age.Change(25),
//	    Active.Change(true),
//	}
type MultiFieldChange []Intent

// IntentType implements the Intent interface.
func (MultiFieldChange) IntentType() string {
	return "MultiFieldChange"
}

// Add appends a TypedFieldChange to the MultiFieldChange.
//
// Example:
//
//	changes := intent.MultiFieldChange{}
//	changes = changes.Add(Username.Change("alice"))
func (m MultiFieldChange) Add(intent Intent) MultiFieldChange {
	return append(m, intent)
}

// AddTyped is a typed helper for adding field changes.
//
// Example:
//
//	changes := intent.MultiFieldChange{}
//	changes = changes.AddTyped(Username, "alice")
func AddTyped[T any](m MultiFieldChange, key StateKey[T], value T) MultiFieldChange {
	return append(m, key.Change(value))
}

// =============================================================================
// Field Matching Helper Functions
// =============================================================================

// MatchTypedField is a helper function to match a TypedFieldChange in reducers.
// This provides a cleaner API than raw type assertions.
//
// Example:
//
//	func reducer(state AppState, intent Intent) AppState {
//	    if change, ok := intent.MatchTypedField(Username); ok {
//	        state.Username = change.Value
//	    }
//	    if change, ok := intent.MatchTypedField(Age); ok {
//	        state.Age = change.Value
//	    }
//	    return state
//	}
//
// Note: Due to Go's type system limitations, use switch statements for best type safety:
//
//	switch v := intent.(type) {
//	case TypedFieldChange[string]:
//	    switch v.Key.String() {
//	    case "username":
//	        state.Username = v.Value
//	    }
//	case TypedFieldChange[int]:
//	    switch v.Key.String() {
//	    case "age":
//	        state.Age = v.Value
//	    }
//	}
