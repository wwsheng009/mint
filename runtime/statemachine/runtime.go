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

// ViewFunction[T] is a pure function that renders state to VNode.
// It should be stateless and deterministic - same state = same output.
type ViewFunction[T any] func(state T) any // Using any to avoid importing ui.VNode

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
	history    []T
	maxHistory int
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
	// Record history
	if rt.maxHistory > 0 {
		rt.history = append(rt.history, state)
		if len(rt.history) > rt.maxHistory {
			rt.history = rt.history[1:]
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
	defer rt.mu.Unlock()

	if index < 0 || index >= len(rt.history) {
		return fmt.Errorf("index out of range: %d (history size: %d)", index, len(rt.history))
	}

	state := rt.history[index]
	rt.store.Set(state)
	return nil
}

// Undo reverts to the previous state.
func (rt *AppRuntime[T]) Undo() error {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	if len(rt.history) <= 1 {
		return fmt.Errorf("no previous state to undo to")
	}

	// Remove current state from history
	rt.history = rt.history[:len(rt.history)-1]

	// Get previous state
	prev := rt.history[len(rt.history)-1]
	rt.store.Set(prev)
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
	return len(rt.history) - 1
}

// CanUndo returns true if undo is possible.
func (rt *AppRuntime[T]) CanUndo() bool {
	rt.mu.RLock()
	defer rt.mu.RUnlock()
	return len(rt.history) > 1
}
