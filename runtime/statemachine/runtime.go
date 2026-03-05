// Package statemachine provides a state-machine-driven runtime for UI applications.
//
// AppRuntime[T] is the main entry point for applications using the Store + Reducer pattern.
// It integrates:
//   - Store[T]: Single source of truth for application state
//   - Reducer[T]: Pure function for state transformations
//   - Dispatcher: Intent handling and scheduling
//
// Architecture:
//
//	┌─────────────────────────────────────────────────────────────────────┐
//	│                        AppRuntime[T]                                │
//	│                                                                      │
//	│  Intent → Dispatcher → Reducer → Store → View                       │
//	│                                                                      │
//	│  Features:                                                           │
//	│  - Single direction data flow                                        │
//	│  - Type-safe state management                                        │
//	│  - Subscription-based rendering                                      │
//	│  - Time-travel debugging support                                     │
//	└─────────────────────────────────────────────────────────────────────┘
//
// Example:
//
//	type AppState struct {
//		Count    int
//		Username string
//	}
//
//	func AppView(state AppState) ui.VNode {
//		return ui.VStack(
//			ui.Text(fmt.Sprintf("Count: %d", state.Count)),
//			ui.NewButtonBuilder("+").OnPress(IncrementIntent{}).Build(),
//		)
//	}
//
//	func main() {
//		rt := statemachine.NewAppRuntime(AppState{}, AppView, AppReducer)
//		ui.RunApp(rt)
//	}
package statemachine

import (
	"context"
	"fmt"
	"sync"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/reducer"
	"github.com/wwsheng009/mint/runtime/store"
)

// ViewFunction[T] is a pure function that renders state to a renderable result.
// It should be stateless and deterministic - same state = same output.
//
// NOTE: Returns 'any' to avoid circular dependency with ui package.
// For type safety, use AppRuntime.NewTyped() or wrap your view function.
//
// Example:
//
//	// Direct usage (less type-safe):
//	func AppView(state AppState) any {
//		return ui.VStack(...)
//	}
//
//	// Type-safe usage with wrapper:
//	func renderAppView(state AppState) ui.VNode {
//		return ui.VStack(...)
//	}
//	func AppView(state AppState) any {
//		return renderAppView(state)
//	}
type ViewFunction[T any] func(state T) any

// AppRuntime[T] is the main runtime for state-machine-driven applications.
// It manages the complete lifecycle of state, intents, and rendering.
type AppRuntime[T any] struct {
	mu sync.RWMutex

	// Core components
	store   *store.Store[T]
	reducer reducer.Reducer[T]
	view    ViewFunction[T]

	// Context for cancellation
	ctx    context.Context
	cancel context.CancelFunc

	// State change callback
	onStateChange func(T)

	// Debug support
	history      []T
	currentIndex int // Current position in history index
	maxHistory   int
	skipHistory  bool // Flag to skip history recording (for Undo/JumpTo)
}

// RuntimeConfig holds configuration for AppRuntime.
type RuntimeConfig struct {
	// MaxHistory is the maximum number of states to keep for time-travel debugging.
	// Set to 0 to disable history.
	MaxHistory int
}

// DefaultRuntimeConfig returns the default configuration.
func DefaultRuntimeConfig() RuntimeConfig {
	return RuntimeConfig{
		MaxHistory: 100,
	}
}

// RuntimeOption is a function that configures the runtime.
type RuntimeOption func(*RuntimeConfig)

// WithMaxHistory sets the maximum history size for time-travel debugging.
func WithMaxHistory(n int) RuntimeOption {
	return func(c *RuntimeConfig) {
		c.MaxHistory = n
	}
}

// NewAppRuntime creates a new AppRuntime with the given initial state,
// view function, and reducer.
//
// Example:
//
//	rt := NewAppRuntime(
//		AppState{Count: 0},
//		AppView,
//		AppReducer,
//		WithMaxHistory(50),
//	)
func NewAppRuntime[T any](
	initial T,
	view ViewFunction[T],
	red reducer.Reducer[T],
	opts ...RuntimeOption,
) *AppRuntime[T] {
	config := DefaultRuntimeConfig()
	for _, opt := range opts {
		opt(&config)
	}

	ctx, cancel := context.WithCancel(context.Background())

	rt := &AppRuntime[T]{
		store:      store.NewStore(initial),
		reducer:    red,
		view:       view,
		ctx:        ctx,
		cancel:     cancel,
		maxHistory: config.MaxHistory,
	}

	// Subscribe to state changes
	rt.store.Subscribe(rt.handleStateChange)

	// Record initial state
	if config.MaxHistory > 0 {
		rt.history = []T{initial}
	}

	return rt
}

// handleStateChange is called when state changes.
func (rt *AppRuntime[T]) handleStateChange(state T) {
	rt.mu.Lock()
	// Record history (only if not skipping)
	if !rt.skipHistory && rt.maxHistory > 0 {
		// If we're in the middle of history and perform a new action,
		// truncate history from current position forward
		if rt.currentIndex < len(rt.history)-1 {
			rt.history = rt.history[:rt.currentIndex+1]
		}
		rt.history = append(rt.history, state)
		rt.currentIndex = len(rt.history) - 1
		if len(rt.history) > rt.maxHistory {
			rt.history = rt.history[1:]
			rt.currentIndex--
		}
	}
	callback := rt.onStateChange
	rt.mu.Unlock()

	// Call callback outside of lock
	if callback != nil {
		callback(state)
	}
}

// =============================================================================
// Core API
// =============================================================================

// GetState returns the current state.
func (rt *AppRuntime[T]) GetState() T {
	return rt.store.Get()
}

// Dispatch dispatches an intent to be processed by the reducer.
func (rt *AppRuntime[T]) Dispatch(i intent.Intent) {
	prev := rt.store.Get()
	next := rt.reducer(prev, i)
	rt.store.Set(next)
}

// Subscribe registers a callback to be called when state changes.
// Returns an unsubscribe function.
func (rt *AppRuntime[T]) Subscribe(callback func(T)) func() {
	return rt.store.Subscribe(callback)
}

// OnStateChange sets the callback for state changes.
// This is a convenience method for setting up rendering.
func (rt *AppRuntime[T]) OnStateChange(callback func(T)) {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	rt.onStateChange = callback
}

// View renders the current state.
func (rt *AppRuntime[T]) View() any {
	return rt.view(rt.store.Get())
}

// GetStore returns the store for this runtime.
// This is useful for registering intent handlers.
func (rt *AppRuntime[T]) GetStore() *store.Store[T] {
	return rt.store
}

// GetReducer returns the reducer for this runtime.
func (rt *AppRuntime[T]) GetReducer() reducer.Reducer[T] {
	return rt.reducer
}

// =============================================================================
// Lifecycle
// =============================================================================

// Close stops the runtime and releases resources.
func (rt *AppRuntime[T]) Close() error {
	rt.cancel()
	return nil
}

// Context returns the runtime's context.
func (rt *AppRuntime[T]) Context() context.Context {
	return rt.ctx
}

// =============================================================================
// Debug Support (Time Travel)
// =============================================================================

// History returns the state history for time-travel debugging.
func (rt *AppRuntime[T]) History() []T {
	rt.mu.RLock()
	defer rt.mu.RUnlock()
	return append([]T{}, rt.history...)
}

// JumpTo jumps to a specific state in history.
// This is useful for time-travel debugging.
func (rt *AppRuntime[T]) JumpTo(index int) error {
	rt.mu.Lock()

	if index < 0 || index >= len(rt.history) {
		rt.mu.Unlock()
		return fmt.Errorf("index out of range: %d (history size: %d)", index, len(rt.history))
	}

	state := rt.history[index]
	rt.currentIndex = index

	// Skip history recording during time jump
	rt.skipHistory = true
	rt.mu.Unlock()

	rt.store.Set(state)

	rt.mu.Lock()
	rt.skipHistory = false
	rt.mu.Unlock()

	return nil
}

// Undo reverts to the previous state.
func (rt *AppRuntime[T]) Undo() error {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	if rt.currentIndex <= 0 {
		return fmt.Errorf("no previous state to undo to")
	}

	// Move to previous state index
	rt.currentIndex--

	// Get previous state
	prev := rt.history[rt.currentIndex]

	// Skip history recording during undo
	rt.skipHistory = true
	rt.mu.Unlock()

	rt.store.Set(prev)

	rt.mu.Lock()
	rt.skipHistory = false

	return nil
}

// Redo reinstates the next state that was previously undone.
func (rt *AppRuntime[T]) Redo() error {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	if rt.currentIndex >= len(rt.history)-1 {
		return fmt.Errorf("no next state to redo to")
	}

	// Move to next state index
	rt.currentIndex++

	// Get next state
	next := rt.history[rt.currentIndex]

	// Skip history recording during redo
	rt.skipHistory = true
	rt.mu.Unlock()

	rt.store.Set(next)

	rt.mu.Lock()
	rt.skipHistory = false

	return nil
}



// Reset resets the state to the initial state.
func (rt *AppRuntime[T]) Reset(initial T) {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	rt.store.Set(initial)
	rt.history = []T{initial}
}

// HistoryIndex returns the current index in history.
func (rt *AppRuntime[T]) HistoryIndex() int {
	rt.mu.RLock()
	defer rt.mu.RUnlock()
	return rt.currentIndex
}

// CanUndo returns true if undo is possible.
func (rt *AppRuntime[T]) CanUndo() bool {
	rt.mu.RLock()
	defer rt.mu.RUnlock()
	return rt.currentIndex > 0
}

// CanRedo returns true if redo is possible.
func (rt *AppRuntime[T]) CanRedo() bool {
	rt.mu.RLock()
	defer rt.mu.RUnlock()
	return rt.currentIndex < len(rt.history)-1
}
