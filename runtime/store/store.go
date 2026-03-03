// Package store provides a type-safe state container for the Mint UI runtime.
//
// Store[T] is the single source of truth for application state.
// It follows the unidirectional data flow pattern:
//
//	Intent → Dispatcher → Reducer → Store → View
//
// Architecture:
//
//	┌─────────────────────────────────────────────────────────────────────┐
//	│                      状态层 (Store)                                  │
//	│  Store[T] <- 单一真相源                                              │
//	│                                                                      │
//	│  Features:                                                           │
//	│  - Type-safe state access                                            │
//	│  - Subscription-based change notification                            │
//	│  - Immutable state updates (via Reducer)                             │
//	│  - Thread-safe operations                                            │
//	└─────────────────────────────────────────────────────────────────────┘
package store

import (
	"sync"
)

// Store[T] is a type-safe state container.
// It holds a single state value and notifies subscribers when state changes.
//
// Key principles:
//  1. Single source of truth - there should be only one Store per application
//  2. Immutable updates - state changes only through Reducer
//  3. Subscription-based - components subscribe to state changes
//
// Example:
//
//	type AppState struct {
//		Count    int
//		Username string
//	}
//
//	store := store.NewStore(AppState{Count: 0, Username: ""})
//	store.Subscribe(func(state AppState) {
//		fmt.Printf("State changed: %+v\n", state)
//	})
//	store.Set(AppState{Count: 1, Username: "john"})
type Store[T any] struct {
	mu        sync.RWMutex
	state     T
	listeners []func(T)
}

// NewStore creates a new Store with the given initial state.
func NewStore[T any](initial T) *Store[T] {
	return &Store[T]{
		state:     initial,
		listeners: make([]func(T), 0),
	}
}

// Get returns the current state.
// This is a read-only operation and is thread-safe.
func (s *Store[T]) Get() T {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state
}

// Set updates the state and notifies all subscribers.
// This should typically be called by the Reducer, not directly by components.
func (s *Store[T]) Set(next T) {
	s.mu.Lock()
	s.state = next
	listeners := make([]func(T), len(s.listeners))
	copy(listeners, s.listeners)
	s.mu.Unlock()

	// Notify subscribers outside of lock
	for _, listener := range listeners {
		listener(next)
	}
}

// Subscribe registers a callback to be called when state changes.
// Returns an unsubscribe function.
//
// Example:
//
//	unsubscribe := store.Subscribe(func(state AppState) {
//		render(state)
//	})
//	// Later...
//	unsubscribe()
func (s *Store[T]) Subscribe(callback func(T)) func() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.listeners = append(s.listeners, callback)
	index := len(s.listeners) - 1

	return func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		// Remove listener by index
		if index < len(s.listeners) {
			s.listeners = append(s.listeners[:index], s.listeners[index+1:]...)
		}
	}
}

// Update applies a function to the current state and sets the result.
// This is useful for atomic updates.
//
// Example:
//
//	store.Update(func(state AppState) AppState {
//		state.Count++
//		return state
//	})
func (s *Store[T]) Update(fn func(T) T) {
	s.mu.Lock()
	next := fn(s.state)
	s.state = next
	listeners := make([]func(T), len(s.listeners))
	copy(listeners, s.listeners)
	s.mu.Unlock()

	// Notify subscribers outside of lock
	for _, listener := range listeners {
		listener(next)
	}
}

// ListenerCount returns the number of registered listeners.
// Useful for debugging and testing.
func (s *Store[T]) ListenerCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.listeners)
}

// =============================================================================
// Selector Support
// =============================================================================

// Selector[T, R] is a function that extracts a derived value from state.
type Selector[T any, R any] func(state T) R

// Select applies a selector to the current state.
// Useful for derived data that doesn't need to be stored.
//
// Example:
//
//	usernameLength := store.Select(func(state AppState) int {
//		return len(state.Username)
//	})
func (s *Store[T]) Select(selector Selector[T, any]) any {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return selector(s.state)
}

// SelectWith performs a type-safe selection.
func SelectWith[T any, R any](s *Store[T], selector func(T) R) R {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return selector(s.state)
}

// =============================================================================
// Computed Values
// =============================================================================

// Computed[T, R] represents a derived value that is computed from state.
type Computed[T any, R any] struct {
	store    *Store[T]
	selector func(T) R
	cache    R
	dirty    bool
	mu       sync.RWMutex
}

// NewComputed creates a computed value that is automatically updated
// when the store state changes.
func NewComputed[T any, R any](store *Store[T], selector func(T) R) *Computed[T, R] {
	c := &Computed[T, R]{
		store:    store,
		selector: selector,
		dirty:    true,
	}

	// Subscribe to store changes
	store.Subscribe(func(state T) {
		c.mu.Lock()
		c.dirty = true
		c.mu.Unlock()
	})

	return c
}

// Get returns the computed value, recalculating if necessary.
func (c *Computed[T, R]) Get() R {
	c.mu.RLock()
	if !c.dirty {
		defer c.mu.RUnlock()
		return c.cache
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	// Double check after acquiring write lock
	if !c.dirty {
		return c.cache
	}

	c.cache = c.selector(c.store.Get())
	c.dirty = false
	return c.cache
}
