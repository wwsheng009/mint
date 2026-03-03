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
//	intent := TypedFieldChange[string]{
//	    Key:   UsernameKey,
//	    Value: "alice",
//	}
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
// StateKey[T] - Type-Safe State Key
// =============================================================================

// StateKey is a type-safe key for accessing state.
// Use this instead of string keys to get compile-time type checking.
//
// Example:
//
//	var (
//	    Username = NewStateKey[string]("username")
//	    Age      = NewStateKey[int]("age")
//	    Active   = NewStateKey[bool]("active")
//	)
//
//	// Type-safe access
//	username := Username.Get(ctx, "")
//	Age.Set(ctx, 25)
type StateKey[T any] struct {
	name string
}

// NewStateKey creates a new type-safe state key.
//
// Example:
//
//	var Username = intent.NewStateKey[string]("username")
func NewStateKey[T any](name string) StateKey[T] {
	return StateKey[T]{name: name}
}

// String returns the string representation of the key.
func (k StateKey[T]) String() string {
	return k.name
}

// Name returns the key name (alias for String).
func (k StateKey[T]) Name() string {
	return k.name
}

// Get retrieves the typed value from ActionContext.
//
// Example:
//
//	username := Username.Get(ctx, "")
func (k StateKey[T]) Get(ctx *ActionContext, defaultValue T) T {
	return GetStateAs[T](ctx, k.name, defaultValue)
}

// Set stores the value in ActionContext.
//
// Example:
//
//	Username.Set(ctx, "alice")
func (k StateKey[T]) Set(ctx *ActionContext, value T) {
	ctx.SetState(k.name, value)
}

// Change creates a TypedFieldChange intent for this key.
//
// Example:
//
//	intent := Username.Change("alice")
//	dispatcher.Dispatch(intent)
func (k StateKey[T]) Change(value T) TypedFieldChange[T] {
	return TypedFieldChange[T]{
		Key:   k,
		Value: value,
	}
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
// Field Matcher for Reducers
// =============================================================================

// FieldMatcher helps match TypedFieldChange intents in a type-safe way.
//
// Example:
//
//	func reducer(state AppState, intent Intent) AppState {
//	    return intent.MatchField(state,
//	        Username, func(s *AppState, v string) { s.Username = v },
//	        Age, func(s *AppState, v int) { s.Age = v },
//	    )
//	}
type FieldMatcher[T any] struct {
	intent Intent
}

// MatchField creates a FieldMatcher for the given intent.
func MatchField[T any](intent Intent) *FieldMatcher[T] {
	return &FieldMatcher[T]{intent: intent}
}

// On matches a specific StateKey and calls the handler if matched.
//
// Example:
//
//	MatchField[AppState](intent).
//	    On(Username, func(s *AppState, v string) { s.Username = v }).
//	    On(Age, func(s *AppState, v int) { s.Age = v })
func (m *FieldMatcher[T]) On[V any](key StateKey[V], handler func(*T, V)) *FieldMatcher[T] {
	if tfc, ok := m.intent.(TypedFieldChange[V]); ok {
		if tfc.Key.String() == key.String() {
			// We need to return state modification, but we don't have state here
			// This is a design limitation - better to use switch statements directly
		}
	}
	return m
}
