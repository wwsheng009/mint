// Package intent provides a type-safe Intent system for declarative UI actions.
//
// Intent replaces closure-based callbacks with structured, schedulable actions.
// This enables:
//   - Type-safe action definitions
//   - Priority-based scheduling via Lane mapping
//   - Transition support for async operations
//   - Suspense integration
//   - DevTools tracing
//
// Architecture:
//
//	Component → Emit Intent[T] → Registry → Dispatcher → Lane → Scheduler → Handler
package intent

import (
	"context"
	"time"

	"github.com/wwsheng009/mint/runtime/priority"
)

// =============================================================================
// Core Intent Types
// =============================================================================

// Intent is the base interface for all intents.
// Every intent type must implement this interface to be dispatchable.
//
// Example:
//
//	type OpenModalIntent struct{}
//	func (OpenModalIntent) IntentType() string { return "OpenModal" }
type Intent interface {
	// IntentType returns the unique type identifier for this intent.
	// This is used for routing and debugging.
	IntentType() string
}

// PriorityAware is an optional interface for intents that declare their priority.
// If an intent implements this, the dispatcher will use this priority.
//
// Example:
//
//	func (OpenModalIntent) Priority() ActionPriority {
//	    return PriorityUserBlocking
//	}
type PriorityAware interface {
	Intent
	Priority() ActionPriority
}

// TransitionIntent is an interface for intents that support async execution.
// Transition intents can be:
//   - Deferred to lower priority lanes
//   - Interrupted by higher priority work
//   - Show loading states via Suspense
//
// Example:
//
//	type LoadDataIntent struct {
//	    URL string
//	}
//	func (LoadDataIntent) IntentType() string { return "LoadData" }
//	func (LoadDataIntent) IsTransition() bool { return true }
type TransitionIntent interface {
	Intent
	IsTransition() bool
}

// ActionPriority represents the priority level of an intent.
// Higher priority intents are processed first.
type ActionPriority int

const (
	// PriorityImmediate is for urgent, blocking operations (e.g., focus changes).
	// These are processed synchronously and cannot be interrupted.
	PriorityImmediate ActionPriority = iota

	// PriorityUserBlocking is for user-initiated actions (e.g., button clicks, input).
	// These should be processed quickly to maintain responsiveness.
	PriorityUserBlocking

	// PriorityNormal is for standard updates (default).
	PriorityNormal

	// PriorityTransition is for async operations that can be deferred.
	// These can be interrupted by higher priority work.
	PriorityTransition

	// PriorityIdle is for background tasks that can wait.
	PriorityIdle
)

// String returns the string representation of the priority.
func (p ActionPriority) String() string {
	switch p {
	case PriorityImmediate:
		return "Immediate"
	case PriorityUserBlocking:
		return "UserBlocking"
	case PriorityNormal:
		return "Normal"
	case PriorityTransition:
		return "Transition"
	case PriorityIdle:
		return "Idle"
	default:
		return "Unknown"
	}
}

// ToLane converts ActionPriority to a priority.DirtyLevel.
func (p ActionPriority) ToLane() priority.DirtyLevel {
	switch p {
	case PriorityImmediate:
		return priority.DirtyHigh
	case PriorityUserBlocking:
		return priority.DirtyHigh
	case PriorityNormal:
		return priority.DirtyNormal
	case PriorityTransition:
		return priority.DirtyLow
	case PriorityIdle:
		return priority.DirtyLow
	default:
		return priority.DirtyNormal
	}
}

// =============================================================================
// ActionContext
// =============================================================================

// ActionContext provides context and utilities for intent handlers.
type ActionContext struct {
	// Context is the standard Go context for cancellation.
	context.Context

	// Source is the component that emitted this intent.
	Source string

	// Timestamp is when the intent was dispatched.
	Timestamp time.Time

	// stateSetter allows handlers to update state.
	stateSetter StateSetter
}

// StateSetter is the interface for updating state.
type StateSetter interface {
	// SetState updates a state value by key.
	SetState(key string, value interface{})

	// GetState retrieves a state value by key.
	GetState(key string) (interface{}, bool)

	// ScheduleUpdate schedules a component update.
	ScheduleUpdate()
}

// NewActionContext creates a new ActionContext.
func NewActionContext(ctx context.Context, source string, setter StateSetter) *ActionContext {
	return &ActionContext{
		Context:     ctx,
		Source:      source,
		Timestamp:   time.Now(),
		stateSetter: setter,
	}
}

// SetState updates a state value.
func (c *ActionContext) SetState(key string, value interface{}) {
	if c.stateSetter != nil {
		c.stateSetter.SetState(key, value)
	}
}

// GetState retrieves a state value.
func (c *ActionContext) GetState(key string) (interface{}, bool) {
	if c.stateSetter != nil {
		return c.stateSetter.GetState(key)
	}
	return nil, false
}

// ScheduleUpdate schedules a component update.
func (c *ActionContext) ScheduleUpdate() {
	if c.stateSetter != nil {
		c.stateSetter.ScheduleUpdate()
	}
}

// =============================================================================
// Type-safe State Accessors (Phase 1.2 of Refactoring)
// =============================================================================

// GetStringState retrieves a string state value with a default.
// This is the preferred way for handlers to read string state.
func (c *ActionContext) GetStringState(key string, defaultValue string) string {
	if c.stateSetter != nil {
		if v, ok := c.stateSetter.GetState(key); ok {
			if s, ok := v.(string); ok {
				return s
			}
		}
	}
	return defaultValue
}

// GetIntState retrieves an int state value with a default.
// This is the preferred way for handlers to read int state.
func (c *ActionContext) GetIntState(key string, defaultValue int) int {
	if c.stateSetter != nil {
		if v, ok := c.stateSetter.GetState(key); ok {
			if i, ok := v.(int); ok {
				return i
			}
		}
	}
	return defaultValue
}

// GetBoolState retrieves a boolean state value with a default.
// This is the preferred way for handlers to read boolean state.
func (c *ActionContext) GetBoolState(key string, defaultValue bool) bool {
	if c.stateSetter != nil {
		if v, ok := c.stateSetter.GetState(key); ok {
			if b, ok := v.(bool); ok {
				return b
			}
		}
	}
	return defaultValue
}

// GetFloat64State retrieves a float64 state value with a default.
func (c *ActionContext) GetFloat64State(key string, defaultValue float64) float64 {
	if c.stateSetter != nil {
		if v, ok := c.stateSetter.GetState(key); ok {
			if f, ok := v.(float64); ok {
				return f
			}
		}
	}
	return defaultValue
}

// GetStateAs retrieves a typed state value with a default.
// This is a generic version for custom types.
func GetStateAs[T any](c *ActionContext, key string, defaultValue T) T {
	if c.stateSetter != nil {
		if v, ok := c.stateSetter.GetState(key); ok {
			if t, ok := v.(T); ok {
				return t
			}
		}
	}
	return defaultValue
}

// =============================================================================
// Intent Result
// =============================================================================

// IntentResult represents the result of an intent handler.
type IntentResult struct {
	// Handled indicates whether the intent was handled.
	Handled bool

	// Error contains any error that occurred during handling.
	Error error

	// Async indicates if this intent requires async processing.
	// If true, the handler returns a channel that will be closed when done.
	Async bool

	// Done is closed when async processing is complete.
	Done chan struct{}
}

// HandledResult creates a simple handled result.
func HandledResult() IntentResult {
	return IntentResult{Handled: true}
}

// ErrorResult creates an error result.
func ErrorResult(err error) IntentResult {
	return IntentResult{Handled: false, Error: err}
}

// AsyncResult creates an async result.
func AsyncResult(done chan struct{}) IntentResult {
	return IntentResult{Handled: true, Async: true, Done: done}
}
