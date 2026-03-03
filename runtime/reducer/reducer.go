// Package reducer provides state transformation functions for the Mint UI runtime.
//
// Reducer is a pure function that takes the current state and an intent,
// and returns a new state. It is the only place where state changes happen.
//
// Architecture:
//
//	┌─────────────────────────────────────────────────────────────────────┐
//	│                      逻辑层 (Reducer)                                │
//	│  Reducer(state, intent) → newState                                 │
//	│  📍 唯一状态变更入口：验证、业务逻辑、副作用管理                      │
//	│                                                                      │
//	│  Key Principles:                                                     │
//	│  1. Pure function - no side effects                                 │
//	│  2. Returns new state - never mutates input                         │
//	│  3. Single entry point - all state changes go through here          │
//	└─────────────────────────────────────────────────────────────────────┘
//
// Example:
//
//	type AppState struct {
//		Count int
//	}
//
//	var AppReducer = reducer.New[AppState](
//		func(state AppState, intent intent.Intent) AppState {
//			switch i := intent.(type) {
//			case IncrementIntent:
//				state.Count++
//			case DecrementIntent:
//				state.Count--
//			}
//			return state
//		},
//	)
package reducer

import (
	"github.com/wwsheng009/mint/runtime/intent"
)

// Reducer[T] is a function that transforms state based on an intent.
// It should be a pure function - no side effects, no mutation of input.
type Reducer[T any] func(state T, i intent.Intent) T

// New creates a new Reducer from a function.
func New[T any](fn func(state T, i intent.Intent) T) Reducer[T] {
	return Reducer[T](fn)
}

// Reduce applies the reducer to the current state and intent.
func (r Reducer[T]) Reduce(state T, i intent.Intent) T {
	return r(state, i)
}

// =============================================================================
// Reducer Composition
// =============================================================================

// Compose combines multiple reducers into one.
// Reducers are applied in order, each receiving the state from the previous.
//
// Example:
//
//	combined := reducer.Compose(
//		counterReducer,
//		formReducer,
//		uiReducer,
//	)
func Compose[T any](reducers ...Reducer[T]) Reducer[T] {
	return func(state T, i intent.Intent) T {
		for _, r := range reducers {
			state = r(state, i)
		}
		return state
	}
}

// =============================================================================
// Typed Reducer Builder
// =============================================================================

// Builder provides a fluent API for building reducers.
type Builder[T any] struct {
	handlers map[string]func(T, intent.Intent) T
}

// NewBuilder creates a new reducer builder.
func NewBuilder[T any]() *Builder[T] {
	return &Builder[T]{
		handlers: make(map[string]func(T, intent.Intent) T),
	}
}

// On registers a handler for a specific intent type.
//
// Example:
//
//	reducer.NewBuilder[AppState]().
//		On(IncrementIntent{}, func(s AppState, i intent.Intent) AppState {
//			s.Count++
//			return s
//		}).
//		On(DecrementIntent{}, func(s AppState, i intent.Intent) AppState {
//			s.Count--
//			return s
//		}).
//		Build()
func (b *Builder[T]) On(intentType intent.Intent, handler func(T, intent.Intent) T) *Builder[T] {
	b.handlers[intentType.IntentType()] = handler
	return b
}

// OnTyped registers a type-safe handler for a specific intent type.
// The handler receives the typed intent instead of the base Intent interface.
func OnTyped[T any, I intent.Intent](b *Builder[T], intentType I, handler func(T, I) T) *Builder[T] {
	b.handlers[intentType.IntentType()] = func(state T, i intent.Intent) T {
		if typed, ok := i.(I); ok {
			return handler(state, typed)
		}
		return state
	}
	return b
}

// Build creates a Reducer from the builder.
func (b *Builder[T]) Build() Reducer[T] {
	return func(state T, i intent.Intent) T {
		if handler, ok := b.handlers[i.IntentType()]; ok {
			return handler(state, i)
		}
		return state
	}
}

// =============================================================================
// Middleware Support
// =============================================================================

// Middleware wraps a reducer for cross-cutting concerns.
type Middleware[T any] func(next Reducer[T]) Reducer[T]

// WithMiddleware applies middleware to a reducer.
//
// Example:
//
//	reducer.WithMiddleware(
//		myReducer,
//		loggingMiddleware,
//		validationMiddleware,
//	)
func WithMiddleware[T any](r Reducer[T], middlewares ...Middleware[T]) Reducer[T] {
	// Apply middlewares in reverse order so they execute in order
	for i := len(middlewares) - 1; i >= 0; i-- {
		r = middlewares[i](r)
	}
	return r
}

// LoggingMiddleware logs all intent processing.
func LoggingMiddleware[T any](logFn func(state T, i intent.Intent, newState T)) Middleware[T] {
	return func(next Reducer[T]) Reducer[T] {
		return func(state T, i intent.Intent) T {
			newState := next(state, i)
			if logFn != nil {
				logFn(state, i, newState)
			}
			return newState
		}
	}
}

// =============================================================================
// Common Reducer Patterns
// =============================================================================

// FilterReducer creates a reducer that only processes intents matching a predicate.
func FilterReducer[T any](predicate func(intent.Intent) bool, r Reducer[T]) Reducer[T] {
	return func(state T, i intent.Intent) T {
		if predicate(i) {
			return r(state, i)
		}
		return state
	}
}

// ChainReducer chains multiple reducers for the same intent.
// Each reducer sees the state as modified by previous reducers.
func ChainReducer[T any](reducers ...Reducer[T]) Reducer[T] {
	return func(state T, i intent.Intent) T {
		for _, r := range reducers {
			state = r(state, i)
		}
		return state
	}
}

// =============================================================================
// Immutability Helpers
// =============================================================================

// Clone creates a shallow copy of a struct.
// Useful for ensuring immutability in reducers.
//
// Note: This is a shallow copy. Nested pointers will still reference the same objects.
func Clone[T any](state T) T {
	return state
}

// Update returns a copy of state with updated fields.
// This is a convenience function for updating structs.
//
// Example:
//
//	newState := reducer.Update(state, func(s *AppState) {
//		s.Count++
//		s.LastUpdated = time.Now()
//	})
func Update[T any](state T, updateFn func(*T)) T {
	newState := state // Copy
	updateFn(&newState)
	return newState
}
