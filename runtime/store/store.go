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
	"reflect"
	"sync"
	"sync/atomic"
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
	mu             sync.RWMutex
	state          T
	listeners      []func(T)
	listenersRef   atomic.Value // Stores []func(T) snapshot for Set/Update
}

// NewStore creates a new Store with the given initial state.
func NewStore[T any](initial T) *Store[T] {
	s := &Store[T]{
		state:     initial,
		listeners: make([]func(T), 0),
	}
	// Initialize listeners snapshot with empty slice
	s.listenersRef.Store(make([]func(T), 0))
	return s
}

// updateListenersSnapshot updates the listeners snapshot.
// This must be called with mu.Lock() held.
func (s *Store[T]) updateListenersSnapshot() {
	listeners := make([]func(T), len(s.listeners))
	copy(listeners, s.listeners)
	s.listenersRef.Store(listeners)
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
	s.mu.Unlock()

	// Get listeners snapshot (lock-free read)
	if listeners, ok := s.listenersRef.Load().([]func(T)); ok {
		// Notify subscribers outside of lock
		for _, listener := range listeners {
			listener(next)
		}
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
	s.listeners = append(s.listeners, callback)
	s.updateListenersSnapshot()
	s.mu.Unlock()

	// Capture the function pointer (address) for comparison
	callbackPtr := reflect.ValueOf(callback).Pointer()

	return func() {
		s.mu.Lock()
		defer s.mu.Unlock()

		// Find the listener by function pointer comparison
		for i, listener := range s.listeners {
			if reflect.ValueOf(listener).Pointer() == callbackPtr {
				// Remove listener at this index
				s.listeners = append(s.listeners[:i], s.listeners[i+1:]...)
				s.updateListenersSnapshot()
				return
			}
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
	s.mu.Unlock()

	// Get listeners snapshot (lock-free read)
	if listeners, ok := s.listenersRef.Load().([]func(T)); ok {
		// Notify subscribers outside of lock
		for _, listener := range listeners {
			listener(next)
		}
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
// Note: This method does not cache results. If you need caching, use NewComputed instead.
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
//
// This method reads the current state and applies the selector to it.
// It's equivalent to calling Get() and then applying the selector manually.
//
// Example:
//
//	usernameLength := store.SelectWith(func(state AppState) int {
//		return len(state.Username)
//	})
//	// Equivalent to:
//	// state := store.Get()
//	// usernameLength := len(state.Username)
func SelectWith[T any, R any](s *Store[T], selector func(T) R) R {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return selector(s.state)
}

// SelectCache is a cached selector that remembers the last result.
type SelectCache[T any, R any] struct {
	store    *Store[T]
	selector func(T) R
	cache    *Computed[T, R]
}

// NewSelectCache creates a cached selector with automatic invalidation.
// This is useful when a selector is called frequently and the result is expensive to compute.
//
// The cache is automatically invalidated when the store state changes.
// For fine-grained control, use NewSelectCacheWithInvalidator.
//
// Example:
//
//	usernameLength := store.NewSelectCache(func(state AppState) int {
//		return len(state.Username) // Expensive operation
//	})
//
//	// Multiple calls return cached value
//	_ = usernameLength.Get() // Computes once
//	_ = usernameLength.Get() // Returns cached value
func NewSelectCache[T any, R any](store *Store[T], selector func(T) R) *SelectCache[T, R] {
	return NewSelectCacheWithInvalidator(store, selector, nil)
}

// NewSelectCacheWithInvalidator creates a cached selector with custom invalidation logic.
// The invalidator determines when to invalidate the cache based on state changes.
//
// Example:
//
//	usernameLength := NewSelectCacheWithInvalidator(store,
//		func(state AppState) int {
//			return len(state.Username)
//		},
//		func(prevState, newState AppState) bool {
//			// Only invalidate when username actually changes
//			return prevState.Username != newState.Username
//		})
func NewSelectCacheWithInvalidator[T any, R any](
	store *Store[T],
	selector func(T) R,
	invalidator Invalidator[T],
) *SelectCache[T, R] {
	cache := NewComputedWithInvalidator(store, selector, invalidator)
	return &SelectCache[T, R]{
		store:    store,
		selector: selector,
		cache:    cache,
	}
}

// Get returns the cached value, recomputing if necessary.
func (sc *SelectCache[T, R]) Get() R {
	return sc.cache.Get()
}

// Invalidate forces the cache to be invalidated.
func (sc *SelectCache[T, R]) Invalidate() {
	sc.cache.Invalidate()
}

// Dispose removes the subscription to the store.
func (sc *SelectCache[T, R]) Dispose() {
	sc.cache.Dispose()
}

// InvalidateNow invalidates the cache and returns the new value immediately.
// This is a convenience method for Invalidate() followed by Get().
func (sc *SelectCache[T, R]) InvalidateNow() R {
	sc.Invalidate()
	return sc.Get()
}

// =============================================================================
// Computed Values
// =============================================================================

// Invalidator determines whether a change event should invalidate the computed value.
// Returns true if the cache should be invalidated.
type Invalidator[T any] func(oldState, newState T) bool

// Computed[T, R] represents a derived value that is computed from state.
type Computed[T any, R any] struct {
	store     *Store[T]
	selector  func(T) R
	cache     R
	dirty     bool
	lastState T
	mu        sync.RWMutex
	cancelFn  func() // unsubscribe function
}

// NewComputed creates a computed value that is automatically updated
// when the store state changes.
//
// Example:
//
//	itemCount := store.NewComputed(func(s AppState) int {
//		return len(s.Items)
//	})
//	count := itemCount.Get()
func NewComputed[T any, R any](store *Store[T], selector func(T) R) *Computed[T, R] {
	return NewComputedWithInvalidator(store, selector, nil)
}

// NewComputedWithInvalidator creates a computed value with a custom invalidator.
// The invalidator is called when the store state changes.
// If it returns true, the cache is invalidated; otherwise, the cached value is kept.
//
// This allows for fine-grained cache invalidation, e.g., only invalidate when specific fields change.
//
// Example:
//
//	itemCount := NewComputedWithInvalidator(store,
//		func(s AppState) int { return len(s.Items) },
//		func(old, new AppState) bool {
//			// Only invalidate if Items count changed
//			return len(old.Items) != len(new.Items)
//		})
func NewComputedWithInvalidator[T any, R any](
	store *Store[T],
	selector func(T) R,
	invalidator Invalidator[T],
) *Computed[T, R] {
	c := &Computed[T, R]{
		store:    store,
		selector: selector,
		dirty:    true,
		lastState: store.Get(),
	}

	// Subscribe to store changes
	unsubscribe := store.Subscribe(func(state T) {
		c.mu.Lock()
		defer c.mu.Unlock()

		// Check if we should invalidate the cache
		if invalidator == nil {
			// Default behavior: always invalidate on state change
			c.dirty = true
		} else {
			// Use custom invalidator for fine-grained control
			if invalidator(c.lastState, state) {
				c.dirty = true
			}
			c.lastState = state
		}
	})

	c.cancelFn = unsubscribe
	return c
}

// Get returns the computed value, recalculating if necessary.
//
// This method is thread-safe and can be called from multiple goroutines.
func (c *Computed[T, R]) Get() R {
	c.mu.RLock()
	if !c.dirty {
		defer c.mu.RUnlock()
		return c.cache
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	// Double check after acquiring write lock (double-checked locking pattern)
	if !c.dirty {
		return c.cache
	}

	// Recompute the value
	c.cache = c.selector(c.store.Get())
	c.dirty = false
	return c.cache
}

// Invalidate forces the cache to be invalidated.
// The next call to Get() will recompute the value.
//
// This is useful for manual cache invalidation when you know the underlying
// data has changed but don't want to wait for the next state change.
func (c *Computed[T, R]) Invalidate() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.dirty = true
}

// Dispose removes the subscription to the store.
// After calling Dispose, the computed value will no longer update automatically.
//
// This should be called when the computed value is no longer needed to prevent
// memory leaks.
func (c *Computed[T, R]) Dispose() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cancelFn != nil {
		c.cancelFn()
		c.cancelFn = nil
	}
}
