package intent

import (
	"context"
	"sync"

	mintlog "github.com/wwsheng009/mint/internal/log"
)

// =============================================================================
// Intent Emitter - For Components
// =============================================================================

// Emitter provides a simple interface for components to emit intents.
type Emitter struct {
	dispatcher *Dispatcher
	source     string
}

// NewEmitter creates a new intent emitter.
func NewEmitter(dispatcher *Dispatcher, source string) *Emitter {
	return &Emitter{
		dispatcher: dispatcher,
		source:     source,
	}
}

// Emit dispatches an intent.
func (e *Emitter) Emit(intent Intent) IntentResult {
	return e.dispatcher.DispatchWithSource(intent, e.source)
}

// EmitWithPriority dispatches an intent with explicit priority.
func (e *Emitter) EmitWithPriority(intent Intent, priority ActionPriority) IntentResult {
	return e.dispatcher.DispatchWithPriority(intent, e.source, priority)
}

// =============================================================================
// Intent-aware Component Interface
// =============================================================================

// IntentComponent is an interface for components that emit intents.
type IntentComponent interface {
	// EmitIntent emits an intent from this component.
	EmitIntent(intent Intent) IntentResult
}

// =============================================================================
// Simple State Store
// =============================================================================

// SimpleStore is a basic state store implementation for ActionContext.
type SimpleStore struct {
	mu    sync.RWMutex
	state map[string]interface{}
	dirty bool
}

// NewSimpleStore creates a new simple store.
func NewSimpleStore() *SimpleStore {
	return &SimpleStore{
		state: make(map[string]interface{}),
	}
}

// SetState implements StateSetter.
func (s *SimpleStore) SetState(key string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state[key] = value
	s.dirty = true
}

// GetState implements StateSetter.
func (s *SimpleStore) GetState(key string) (interface{}, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.state[key]
	return val, ok
}

// ScheduleUpdate implements StateSetter.
func (s *SimpleStore) ScheduleUpdate() {
	s.dirty = true
}

// IsDirty returns true if state has changed.
func (s *SimpleStore) IsDirty() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.dirty
}

// ClearDirty clears the dirty flag.
func (s *SimpleStore) ClearDirty() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.dirty = false
}

// GetAll returns all state.
func (s *SimpleStore) GetAll() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make(map[string]interface{}, len(s.state))
	for k, v := range s.state {
		result[k] = v
	}
	return result
}

// =============================================================================
// Runtime Helper
// =============================================================================

// Runtime provides a complete intent runtime setup.
type Runtime struct {
	Registry   *Registry
	Dispatcher *Dispatcher
	Store      *SimpleStore
}

// NewRuntime creates a new intent runtime using the global registry.
// This ensures all handlers registered via DefaultRegistry() are available.
// The Dispatcher and Registry both use IntentLogger by default for logging.
func NewRuntime() *Runtime {
	registry := DefaultRegistry()
	dispatcher := NewDispatcher(registry)
	store := NewSimpleStore()

	dispatcher.SetStateSetter(store)

	return &Runtime{
		Registry:   registry,
		Dispatcher: dispatcher,
		Store:      store,
	}
}

// NewRuntimeWithNewRegistry creates a new intent runtime with a fresh registry.
// Use this for testing or when you need isolation from the global registry.
// The Dispatcher and Registry both use IntentLogger by default for logging.
func NewRuntimeWithNewRegistry() *Runtime {
	registry := NewRegistry()
	dispatcher := NewDispatcher(registry)
	store := NewSimpleStore()

	dispatcher.SetStateSetter(store)

	return &Runtime{
		Registry:   registry,
		Dispatcher: dispatcher,
		Store:      store,
	}
}

// Emit dispatches an intent through the runtime.
func (r *Runtime) Emit(intent Intent) IntentResult {
	return r.Dispatcher.Dispatch(intent)
}

// EmitFromSource dispatches an intent with a source.
func (r *Runtime) EmitFromSource(intent Intent, source string) IntentResult {
	return r.Dispatcher.DispatchWithSource(intent, source)
}

// Register registers a handler for an intent type.
func (r *Runtime) Register(intentType string, handler Handler) func() {
	return r.Registry.Register(intentType, handler)
}

// RegisterTyped registers a type-safe handler.
func RegisterTypedRuntime[T Intent](rt *Runtime, handler TypedHandler[T]) func() {
	return RegisterTyped(rt.Registry, handler)
}

// RegisterTypedRuntimeWithOpts registers a type-safe handler with options.
// This allows overriding builtin handlers for specific intents.
func RegisterTypedRuntimeWithOpts[T Intent](rt *Runtime, handler TypedHandler[T], opts ...RegisterOption) func() {
	return RegisterTypedWithOpts(rt.Registry, handler, opts...)
}

// RegisterTypedWithOpts registers a type-safe handler with options.
// Example:
//
//	RegisterTypedWithOpts(rt, handleFieldChange, WithOverridable(true))
func RegisterTypedWithOpts[T Intent](registry *Registry, handler TypedHandler[T], opts ...RegisterOption) func() {
	// Create a zero value of T to get its type
	var zero T
	intentType := zero.IntentType()

	// Wrap the typed handler to implement Handler
	wrapper := &typedHandlerWrapper[T]{handler: handler}

	// Register with options
	return registry.Register(intentType, wrapper, opts...)
}

// =============================================================================
// ActionContext Builders
// =============================================================================

// NewContext creates an ActionContext with the runtime's store.
func (r *Runtime) NewContext(source string) *ActionContext {
	return NewActionContext(context.Background(), source, r.Store)
}

// GetLogger returns the current logger used by the dispatcher.
func (r *Runtime) GetLogger() *mintlog.Logger {
	return r.Dispatcher.GetLogger()
}

// SetLogger sets a custom logger for the dispatcher.
func (r *Runtime) SetLogger(logger *mintlog.Logger) {
	r.Dispatcher.SetLogger(logger)
}

// =============================================================================
// Intent Builder Pattern
// =============================================================================

// Builder provides a fluent API for creating intents.
type Builder struct {
	intent    Intent
	priority  ActionPriority
	source    string
	scheduler ScheduleFunc
}

// NewBuilder creates a new intent builder.
func NewBuilder(intent Intent) *Builder {
	return &Builder{
		intent:   intent,
		priority: PriorityNormal,
	}
}

// WithPriority sets the priority.
func (b *Builder) WithPriority(p ActionPriority) *Builder {
	b.priority = p
	return b
}

// WithSource sets the source.
func (b *Builder) WithSource(source string) *Builder {
	b.source = source
	return b
}

// WithScheduler sets the scheduler.
func (b *Builder) WithScheduler(fn ScheduleFunc) *Builder {
	b.scheduler = fn
	return b
}

// Dispatch dispatches the intent.
func (b *Builder) Dispatch(d *Dispatcher) IntentResult {
	return d.DispatchWithPriority(b.intent, b.source, b.priority)
}

// Build returns the intent.
func (b *Builder) Build() Intent {
	return b.intent
}

// =============================================================================
// Transition Wrapper
// =============================================================================

// Transition wraps an intent to mark it as a transition.
func Transition(intent Intent) TransitionIntent {
	return &transitionWrapper{intent: intent}
}

type transitionWrapper struct {
	intent Intent
}

func (w *transitionWrapper) IntentType() string {
	return w.intent.IntentType()
}

func (w *transitionWrapper) IsTransition() bool {
	return true
}

// =============================================================================
// Priority Override
// =============================================================================

// WithPriority wraps an intent to override its priority.
func WithPriority(intent Intent, p ActionPriority) PriorityAware {
	return &priorityWrapper{intent: intent, priority: p}
}

type priorityWrapper struct {
	intent   Intent
	priority ActionPriority
}

func (w *priorityWrapper) IntentType() string {
	return w.intent.IntentType()
}

func (w *priorityWrapper) Priority() ActionPriority {
	return w.priority
}
