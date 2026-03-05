package intent

import (
	"fmt"
	"reflect"
	"sync"

	mintlog "github.com/wwsheng009/mint/internal/log"
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

// HandlerRegistration stores handler metadata.
type HandlerRegistration struct {
	Handler     Handler
	Overridable bool
	Priority    ActionPriority
}

// RegisterOption configures handler registration.
type RegisterOption func(*HandlerRegistration)

// WithOverridable marks a handler as overridable.
// If a user registers a handler for the same intent type after this handler,
// the user's handler will replace the overridable handler.
func WithOverridable(overridable bool) RegisterOption {
	return func(reg *HandlerRegistration) {
		reg.Overridable = overridable
	}
}

// WithHandlerPriority sets the explicit priority for the handler registration.
// This overrides the PriorityAware interface on the intent if implemented.
func WithHandlerPriority(priority ActionPriority) RegisterOption {
	return func(reg *HandlerRegistration) {
		reg.Priority = priority
	}
}

// =============================================================================
// Registry
// =============================================================================

// Registry manages intent type registrations and handlers.
// It provides type-safe registration and dispatch.
type Registry struct {
	mu sync.RWMutex

	// handlers maps intent type to handler registration
	handlers map[string]*HandlerRegistration

	// typeMap maps reflect.Type to intent type string
	typeMap map[reflect.Type]string

	// priorities maps intent type to explicit priority (legacy, kept for compatibility)
	priorities map[string]ActionPriority

	// middleware that wraps all handlers
	middleware []Middleware

	// fallbackHandler is called when no specific handler is found
	fallbackHandler Handler

	// logger is the structured logger for registry events
	logger *mintlog.Logger
}

// Middleware wraps a handler for cross-cutting concerns.
type Middleware func(next Handler) Handler

// NewRegistry creates a new intent registry.
func NewRegistry() *Registry {
	return &Registry{
		handlers:   make(map[string]*HandlerRegistration),
		typeMap:    make(map[reflect.Type]string),
		priorities: make(map[string]ActionPriority),
		logger:     mintlog.IntentLogger, // Use dedicated IntentLogger by default
	}
}

// SetLogger sets the structured logger for registry events.
func (r *Registry) SetLogger(logger *mintlog.Logger) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.logger = logger
}

// GetLogger returns the current logger being used by the registry.
// Returns the IntentLogger by default, or a custom logger if SetLogger was called.
func (r *Registry) GetLogger() *mintlog.Logger {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.logger
}

// Register registers a handler for a specific intent type with optional configuration.
// This is the low-level registration method.
//
// Options:
//   - WithOverridable(true): Allow this handler to be overridden by later registrations
//   - WithHandlerPriority(priority): Set explicit priority for this handler
//
// Example:
//
//	registry.Register("Increment", handler, WithOverridable(true))
func (r *Registry) Register(intentType string, handler Handler, opts ...RegisterOption) func() {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check for existing handler
	existing, ok := r.handlers[intentType]

	// Create registration with default values
	reg := &HandlerRegistration{
		Handler: handler,
	}

	// Apply options
	for _, opt := range opts {
		opt(reg)
	}

	// Handle overridable logic
	if ok && existing != nil {
		// If existing handler is not overridable, warn and don't replace
		if !existing.Overridable {
			if r.logger != nil && r.logger.Enabled() {
				r.logger.Warn(
					"Cannot override protected handler: type=%s, currentHandler=%T, newHandler=%T. "+
						"Use WithOverridable(true) to allow overriding, or register the handler before the builtin one.",
					intentType, existing.Handler, handler,
				)
			}
			return func() {} // No-op unregister
		}
		// Existing handler is overridable, it will be replaced
		if r.logger != nil && r.logger.Enabled() {
			r.logger.Info(
				"Successfully overridable handler: type=%s, oldHandler=%T, newHandler=%T",
				intentType, existing.Handler, handler,
			)
		}
	} else {
		// New registration
		if r.logger != nil && r.logger.Enabled() {
			r.logger.Debug("Registering new handler: type=%s, handler=%T, overridable=%v",
				intentType, handler, reg.Overridable)
		}
	}

	r.handlers[intentType] = reg

	// Return unregister function
	return func() {
		r.mu.Lock()
		defer r.mu.Unlock()
		delete(r.handlers, intentType)
		if r.logger != nil && r.logger.Enabled() {
			r.logger.Debug(
				"[Register] Unregistered handler: type=%s, handler=%T",
				intentType, reg.Handler,
			)
		}
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

	// Check for existing handler
	existing, ok := r.handlers[intentType]

	// Create registration with default values (not overridable, default priority)
	reg := &HandlerRegistration{
		Handler: wrapper,
	}

	// Handle overridable logic
	if ok && existing != nil && !existing.Overridable {
		// If existing handler is not overridable, warn and don't replace
		if r.logger != nil && r.logger.Enabled() {
			r.logger.Warn(
				"[RegisterTyped] Cannot override protected handler: type=%s (from %T), newHandler=%T. "+
					"Handler is not marked as overridable. Use RegisterTypedWithOpts with WithOverridable(true) to override.",
				intentType, existing.Handler, wrapper,
			)
		}
		// For now, just return without replacing
		return func() {} // No-op unregister
	}

	// Log handler registration
	if ok && existing != nil {
		// Replacing existing handler
		if r.logger != nil && r.logger.Enabled() {
			r.logger.Info(
				"[RegisterTyped] Replacing handler: type=%s, oldHandler=%T, newHandler=%T",
				intentType, existing.Handler, wrapper,
			)
		}
	} else {
		// New handler registration
		if r.logger != nil && r.logger.Enabled() {
			r.logger.Debug(
				"[RegisterTyped] Registering new handler: type=%s, handler=%T (GoType=%s)",
				intentType, wrapper, typ.String(),
			)
		}
	}

	r.handlers[intentType] = reg
	r.typeMap[typ] = intentType

	// Return unregister function
	return func() {
		r.mu.Lock()
		defer r.mu.Unlock()
		delete(r.handlers, intentType)
		delete(r.typeMap, typ)
		if r.logger != nil && r.logger.Enabled() {
			r.logger.Debug("[RegisterTyped] Unregistered handler: type=%s", intentType)
		}
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

	reg, ok := r.handlers[intentType]
	if !ok {
		// Try fallback handler
		if r.fallbackHandler != nil {
			return r.fallbackHandler, true
		}
		return nil, false
	}

	handler := reg.Handler

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
