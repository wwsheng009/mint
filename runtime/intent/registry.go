package intent

import (
	"fmt"
	"reflect"
	"sync"
)

// =============================================================================
// Intent Handler Types
// =============================================================================

// Handler is the base interface for intent handlers.
type Handler interface {
	// Handle processes the intent.
	Handle(ctx *ActionContext, intent Intent) IntentResult
}

// HandlerFunc is a function adapter for Handler.
type HandlerFunc func(ctx *ActionContext, intent Intent) IntentResult

// Handle implements Handler.
func (f HandlerFunc) Handle(ctx *ActionContext, intent Intent) IntentResult {
	return f(ctx, intent)
}

// TypedHandler is a type-safe handler for a specific intent type.
type TypedHandler[T Intent] func(ctx *ActionContext, intent T) IntentResult

// =============================================================================
// Registry
// =============================================================================

// Registry manages intent type registrations and handlers.
// It provides type-safe registration and dispatch.
type Registry struct {
	mu sync.RWMutex

	// handlers maps intent type to handler
	handlers map[string]Handler

	// typeMap maps reflect.Type to intent type string
	typeMap map[reflect.Type]string

	// priorities maps intent type to explicit priority
	priorities map[string]ActionPriority

	// middleware that wraps all handlers
	middleware []Middleware

	// fallbackHandler is called when no specific handler is found
	fallbackHandler Handler
}

// Middleware wraps a handler for cross-cutting concerns.
type Middleware func(next Handler) Handler

// NewRegistry creates a new intent registry.
func NewRegistry() *Registry {
	return &Registry{
		handlers:   make(map[string]Handler),
		typeMap:    make(map[reflect.Type]string),
		priorities: make(map[string]ActionPriority),
	}
}

// Register registers a handler for a specific intent type.
// This is the low-level registration method.
func (r *Registry) Register(intentType string, handler Handler) func() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.handlers[intentType] = handler

	// Return unregister function
	return func() {
		r.mu.Lock()
		defer r.mu.Unlock()
		delete(r.handlers, intentType)
	}
}

// RegisterTyped registers a type-safe handler for a specific intent type T.
// This provides compile-time type safety.
//
// Example:
//
//	registry.RegisterTyped(func(ctx *ActionContext, intent OpenModalIntent) IntentResult {
//	    ctx.SetState("showModal", true)
//	    return HandledResult()
//	})
func RegisterTyped[T Intent](r *Registry, handler TypedHandler[T]) func() {
	// Create a zero value of T to get its type
	var zero T

	// Check if T is an interface type using reflection
	// When T is an interface type, reflect.TypeOf(zero) returns nil because zero is nil
	// For concrete struct types, reflect.TypeOf(zero) returns the struct's type
	typ := reflect.TypeOf(zero)
	if typ == nil {
		panic(fmt.Sprintf(
			"RegisterTyped: type parameter T cannot be an interface type (e.g., intent.Intent). " +
				"Use a concrete intent struct type instead. " +
				"For generic handlers, use Registry.Register() with an explicit intentType string.",
		))
	}

	intentType := zero.IntentType()

	// Wrap the typed handler to implement Handler
	wrapper := &typedHandlerWrapper[T]{handler: handler}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.handlers[intentType] = wrapper
	r.typeMap[typ] = intentType

	// Return unregister function
	return func() {
		r.mu.Lock()
		defer r.mu.Unlock()
		delete(r.handlers, intentType)
		delete(r.typeMap, typ)
	}
}

// typedHandlerWrapper adapts a TypedHandler to Handler.
type typedHandlerWrapper[T Intent] struct {
	handler TypedHandler[T]
}

// Handle implements Handler.
func (w *typedHandlerWrapper[T]) Handle(ctx *ActionContext, intent Intent) IntentResult {
	if typed, ok := intent.(T); ok {
		return w.handler(ctx, typed)
	}
	return ErrorResult(fmt.Errorf("intent type mismatch: expected %T, got %T", *new(T), intent))
}

// RegisterPriority explicitly sets the priority for an intent type.
// This overrides the PriorityAware interface if implemented.
func (r *Registry) RegisterPriority(intentType string, priority ActionPriority) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.priorities[intentType] = priority
}

// GetPriority returns the priority for an intent.
// Priority resolution order:
//  1. Explicitly registered priority
//  2. PriorityAware interface
//  3. Default (PriorityNormal)
func (r *Registry) GetPriority(intent Intent) ActionPriority {
	intentType := intent.IntentType()

	// Check explicit registration
	r.mu.RLock()
	if p, ok := r.priorities[intentType]; ok {
		r.mu.RUnlock()
		return p
	}
	r.mu.RUnlock()

	// Check PriorityAware
	if pa, ok := intent.(PriorityAware); ok {
		return pa.Priority()
	}

	// Default
	return PriorityNormal
}

// IsTransition returns true if the intent is a transition intent.
func (r *Registry) IsTransition(intent Intent) bool {
	if ti, ok := intent.(TransitionIntent); ok {
		return ti.IsTransition()
	}
	return false
}

// GetHandler returns the handler for an intent type.
// If no specific handler is found but a fallback handler is registered,
// returns the fallback handler.
func (r *Registry) GetHandler(intentType string) (Handler, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	handler, ok := r.handlers[intentType]
	if !ok {
		// Try fallback handler
		if r.fallbackHandler != nil {
			handler = r.fallbackHandler
		} else {
			return nil, false
		}
	}

	// Apply middleware
	for i := len(r.middleware) - 1; i >= 0; i-- {
		handler = r.middleware[i](handler)
	}

	return handler, true
}

// HasHandler returns true if a handler is registered for the intent type.
func (r *Registry) HasHandler(intentType string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.handlers[intentType]
	return ok
}

// Use adds middleware that wraps all handlers.
func (r *Registry) Use(middleware Middleware) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.middleware = append(r.middleware, middleware)
}

// RegisterFallback registers a handler that will be called when no specific
// handler is found for an intent type. This is useful for implementing
// reducer-style handlers that process all intents.
//
// Example:
//
//	registry.RegisterFallback(HandlerFunc(func(ctx *ActionContext, intent Intent) IntentResult {
//		// Process all intents through a reducer
//		newState := reducer.Reduce(currentState, intent)
//		store.Set(newState)
//		return HandledResult()
//	}))
func (r *Registry) RegisterFallback(handler Handler) func() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.fallbackHandler = handler

	// Return unregister function
	return func() {
		r.mu.Lock()
		defer r.mu.Unlock()
		r.fallbackHandler = nil
	}
}

// HasFallback returns true if a fallback handler is registered.
func (r *Registry) HasFallback() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.fallbackHandler != nil
}

// GetRegisteredTypes returns all registered intent types.
func (r *Registry) GetRegisteredTypes() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	types := make([]string, 0, len(r.handlers))
	for t := range r.handlers {
		types = append(types, t)
	}
	return types
}

// =============================================================================
// Global Registry
// =============================================================================

var globalRegistry = NewRegistry()

// DefaultRegistry returns the global default registry.
func DefaultRegistry() *Registry {
	return globalRegistry
}

// RegisterTypedGlobally registers a handler in the global registry.
func RegisterTypedGlobally[T Intent](handler TypedHandler[T]) func() {
	return RegisterTyped(globalRegistry, handler)
}

// RegisterGlobally registers a handler in the global registry.
func RegisterGlobally(intentType string, handler Handler) func() {
	return globalRegistry.Register(intentType, handler)
}

// RegisterFallbackGlobally registers a fallback handler in the global registry.
// This handler will be called when no specific handler is found for an intent type.
//
// Example:
//
//	intent.RegisterFallbackGlobally(intent.HandlerFunc(func(ctx *intent.ActionContext, i intent.Intent) intent.IntentResult {
//		newState := reducer.Reduce(store.Get(), i)
//		store.Set(newState)
//		return intent.HandledResult()
//	}))
func RegisterFallbackGlobally(handler Handler) func() {
	return globalRegistry.RegisterFallback(handler)
}
