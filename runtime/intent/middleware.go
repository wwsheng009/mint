package intent

import (
	"fmt"
	"time"
)

// =============================================================================
// Intent Middleware System (Phase 12)
// =============================================================================

// IntentMiddleware defines the interface for intent middleware.
// Middleware can intercept intent emission before and after routing.
type IntentMiddleware interface {
	// BeforeEmit is called before intent is routed.
	// Can be used for logging, validation, modification, or filtering.
	// Returns error to abort emission.
	BeforeEmit(ctx *MiddlewareContext) error

	// AfterEmit is called after intent is routed.
	// Can be used for logging, metrics, or cleanup.
	AfterEmit(ctx *MiddlewareContext, result *EmitResult)
}

// MiddlewareFunc is a function adapter for IntentMiddleware interface.
type MiddlewareFunc func(ctx *MiddlewareContext, next func(MiddlewareContext) EmitResult) EmitResult

// MiddlewareContext contains context information for middleware processing.
type MiddlewareContext struct {
	// Intent being processed
	Intent Intent

	// Component emitting the intent (may be nil)
	Source ComponentInstance

	// Route type determined by system
	Route RouteType

	// Metadata for middleware use
	Metadata map[string]interface{}

	// Timestamp when intent was emitted
	Timestamp time.Time

	// TraceID for tracking intent through middleware chain
	TraceID string
}

// ComponentInstance represents a component instance that can emit intents.
type ComponentInstance interface {
	// Can be implemented by any component instance
	Parent() interface{}
}

// RouteType indicates where the intent is routed.
type RouteType int

const (
	// RouteGlobal means intent is sent to global Intent Runtime
	RouteGlobal RouteType = iota

	// RouteLocal means intent bubbles locally through Parent() chain
	RouteLocal
)

// RouteTypeString returns string representation of RouteType.
func (r RouteType) String() string {
	switch r {
	case RouteGlobal:
		return "Global"
	case RouteLocal:
		return "Local"
	default:
		return "Unknown"
	}
}

// EmitResult represents the result of intent emission.
type EmitResult struct {
	// Error if emission failed
	Error error

	// Duration of intent processing
	Duration time.Duration

	// Whether intent was handled
	Handled bool
}

// Success creates a successful EmitResult.
func Success(duration time.Duration) EmitResult {
	return EmitResult{
		Duration: duration,
		Handled:  true,
	}
}

// Failure creates a failed EmitResult.
func Failure(err error) EmitResult {
	return EmitResult{
		Error:   err,
		Handled: false,
	}
}

// =============================================================================
// Middleware Chain
// =============================================================================

// MiddlewareChain manages a list of middleware in execution order.
type MiddlewareChain struct {
	middlewares []IntentMiddleware
}

// NewMiddlewareChain creates a new middleware chain.
func NewMiddlewareChain() *MiddlewareChain {
	return &MiddlewareChain{
		middlewares: make([]IntentMiddleware, 0),
	}
}

// Add adds a middleware to the end of the chain.
func (c *MiddlewareChain) Add(mw IntentMiddleware) *MiddlewareChain {
	c.middlewares = append(c.middlewares, mw)
	return c
}

// Use adds a middleware using a function adapter.
func (c *MiddlewareChain) Use(fn MiddlewareFunc) *MiddlewareChain {
	return c.Add(&middlewareFuncAdapter{fn: fn})
}

// Execute executes the middleware chain.
func (c *MiddlewareChain) Execute(ctx *MiddlewareContext, emitFunc func() EmitResult) EmitResult {
	start := time.Now()

	// Run BeforeEmit for each middleware
	for _, mw := range c.middlewares {
		if err := mw.BeforeEmit(ctx); err != nil {
			return Failure(fmt.Errorf("middleware BeforeEmit failed: %w", err))
		}
	}

	// Execute the actual intent emission
	result := emitFunc()
	result.Duration = time.Since(start)

	// Run AfterEmit for each middleware (in reverse order)
	for i := len(c.middlewares) - 1; i >= 0; i-- {
		c.middlewares[i].AfterEmit(ctx, &result)
	}

	return result
}

// Len returns the number of middleware in the chain.
func (c *MiddlewareChain) Len() int {
	return len(c.middlewares)
}

// middlewareFuncAdapter adapts MiddlewareFunc to Middleware interface.
type middlewareFuncAdapter struct {
	fn MiddlewareFunc
}

func (a *middlewareFuncAdapter) BeforeEmit(ctx *MiddlewareContext) error {
	// Not used in adapter mode
	return nil
}

func (a *middlewareFuncAdapter) AfterEmit(ctx *MiddlewareContext, result *EmitResult) {
	// Not used in adapter mode
}

// =============================================================================
// Global Middleware Registry
// =============================================================================

var (
	// globalMiddlewareChain is the global middleware chain
	globalMiddlewareChain = NewMiddlewareChain()
)

// AddGlobalMiddleware adds a middleware to the global chain.
func AddGlobalMiddleware(mw IntentMiddleware) {
	globalMiddlewareChain.Add(mw)
}

// GetGlobalMiddleware returns the global middleware chain.
func GetGlobalMiddleware() *MiddlewareChain {
	return globalMiddlewareChain
}

// =============================================================================
// Utility Functions
// =============================================================================

// NewContext creates a new MiddlewareContext.
func NewContext(intent Intent, source ComponentInstance, route RouteType) *MiddlewareContext {
	return &MiddlewareContext{
		Intent:    intent,
		Source:    source,
		Route:     route,
		Metadata:  make(map[string]interface{}),
		Timestamp: time.Now(),
		TraceID:   generateTraceID(),
	}
}

// generateTraceID generates a unique trace ID.
func generateTraceID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
