package ui

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/wwsheng009/mint/internal/log"
	"github.com/wwsheng009/mint/runtime/intent"
)

// =============================================================================
// Hook Types
// =============================================================================

// HookType represents the type of hook
type HookType int

const (
	// HookState is the useState hook type
	HookState HookType = iota
	// HookEffect is the useEffect hook type
	HookEffect
	// HookContext is the useContext hook type
	HookContext
	// HookMemo is the useMemo/useCallback hook type
	HookMemo
	// HookRef is the useRef hook type
	HookRef
)

// String returns the string representation of HookType
func (h HookType) String() string {
	switch h {
	case HookState:
		return "useState"
	case HookEffect:
		return "useEffect"
	case HookContext:
		return "useContext"
	case HookMemo:
		return "useMemo"
	case HookRef:
		return "useRef"
	default:
		return "unknown"
	}
}

// Hook represents a single hook instance
type Hook struct {
	Type        HookType
	Value       interface{}
	SetValue    func(interface{})
	Deps        []interface{}
	Cleanup     func()
	Initialized bool // Tracks if hook has been initialized
}

// ComponentContext holds the state for a component during rendering
type ComponentContext struct {
	ComponentID string
	Hooks       []Hook
	HookIndex   int
	Validator   *HookValidator
	RenderCount int

	// IntentRuntime is the intent runtime for this component context.
	// Initialized by DeclarativeNode and shared across root component tree.
	IntentRuntime *intent.Runtime

	// GlobalState is the component's global state store, accessible via Intent handlers.
	// This allows Intent handlers to update cross-component state without closures.
	// Example: ctx.SetState("username", "john")
	// Note: Renamed from State to GlobalState for clarity.
	//
	// Deprecated: Use Store + Reducer architecture instead.
	// Use UseStoreField or UseStoreSelector for type-safe state management.
	// Migration guide: https://github.com/wwsheng009/mint/docs/ui/store/guides/MIGRATION_GUIDE.md
	// Status: Will be removed in v1.0
	GlobalState map[string]interface{}

	// PendingUpdates is a batch queue for state updates.
	// This allows multiple SetState calls to be batched into a single re-render.
	// Example: ctx.SetState("field1", v1); ctx.SetState("field2", v2); // Only one re-render
	PendingUpdates map[string]interface{}

	// UpdateScheduled indicates whether a re-render has been scheduled.
	// Prevents multiple schedules for batches of updates.
	UpdateScheduled bool

	// scheduleUpdate is a callback to trigger Fiber re-render.
	// Set by DeclarativeNode during initialization.
	scheduleUpdate func()

	// StateMu protects concurrent access to GlobalState and PendingUpdates.
	StateMu sync.RWMutex
}

// =============================================================================
// Global Context Management
// =============================================================================

var (
	currentContext      *ComponentContext
	currentContextMu    sync.RWMutex
	globalComponentID   int
	globalComponentIDMu sync.Mutex
)

// NextComponentID generates a unique component ID
func NextComponentID() string {
	globalComponentIDMu.Lock()
	defer globalComponentIDMu.Unlock()
	globalComponentID++
	return fmt.Sprintf("comp-%d", globalComponentID)
}

// SetCurrentContext sets the current rendering context
func SetCurrentContext(ctx *ComponentContext) {
	currentContextMu.Lock()
	defer currentContextMu.Unlock()
	currentContext = ctx
}

// GetCurrentContext returns the current rendering context
func GetCurrentContext() *ComponentContext {
	currentContextMu.RLock()
	defer currentContextMu.RUnlock()
	return currentContext
}

// =============================================================================
// Ref
// =============================================================================

// Ref holds a mutable value that persists across renders
type Ref struct {
	Value interface{}
}

// =============================================================================
// Effect Types
// =============================================================================

// EffectCallback is the function passed to useEffect
type EffectCallback func() CleanupFunc

// CleanupFunc is the optional cleanup function returned by EffectCallback
type CleanupFunc func()

// =============================================================================
// ComponentContext Methods
// =============================================================================

// NewComponentContext creates a new component context
func NewComponentContext(name string) *ComponentContext {
	return &ComponentContext{
		ComponentID:    fmt.Sprintf("%s-%s", name, NextComponentID()),
		Hooks:          make([]Hook, 0),
		HookIndex:      0,
		Validator:      NewHookValidator(name),
		RenderCount:    0,
		GlobalState:    make(map[string]interface{}),
		PendingUpdates: make(map[string]interface{}),
	}
}

// NewComponentContextForRoot creates a new component context for the root
func NewComponentContextForRoot() *ComponentContext {
	return NewComponentContext("App")
}

// ResetContext resets the hook index for re-rendering
// IMPORTANT: This is called before each render, so we flush pending updates here.
func (ctx *ComponentContext) ResetContext() {
	// Flush any pending updates before re-rendering
	count := ctx.FlushUpdates()
	if count > 0 && log.UILogger.Enabled() {
		log.UILogger.IfEnabled().Debug("ResetContext: Flushed %d pending updates", count)
	}

	if len(ctx.Hooks) > 0 {
		log.UILogger.IfEnabled().Debug("ResetContext: BEFORE reset, Hooks[0].Value=%v, &ctx=%p", ctx.Hooks[0].Value, ctx)
		ctx.HookIndex = 0
		ctx.RenderCount++
		if len(ctx.Hooks) > 0 {
			log.UILogger.IfEnabled().Debug("ResetContext: AFTER reset, Hooks[0].Value=%v, &ctx=%p", ctx.Hooks[0].Value, ctx)
		}
	}
}

// FinishRender finishes the render and validates hooks
func (ctx *ComponentContext) FinishRender() error {
	return ctx.Validator.FinishRender()
}

// RunEffects executes all pending effects after render
// This should be called by the reconciler after committing changes
func (ctx *ComponentContext) RunEffects() {
	for i := range ctx.Hooks {
		hook := &ctx.Hooks[i]
		if hook.Type == HookEffect && hook.Value != nil {
			callback, ok := hook.Value.(EffectCallback)
			if ok && callback != nil {
				// Run the effect
				cleanup := callback()
				if cleanup != nil {
					hook.Cleanup = cleanup
				}
				// Clear Value to mark this effect as run
				hook.Value = nil
			}
		}
	}
}

// CleanupAll runs all cleanup functions (for unmounting)
func (ctx *ComponentContext) CleanupAll() {
	for i := range ctx.Hooks {
		hook := &ctx.Hooks[i]
		if hook.Cleanup != nil {
			hook.Cleanup()
			hook.Cleanup = nil
		}
	}
}

// GetOrCreateHook gets an existing hook or creates a new one
func (ctx *ComponentContext) GetOrCreateHook(hookType HookType) *Hook {
	if ctx.HookIndex < len(ctx.Hooks) {
		hook := &ctx.Hooks[ctx.HookIndex]
		if log.UILogger.Enabled() {
			log.UILogger.Debug("GetOrCreateHook: returning existing hook[%d], Type=%s, Value=%v, Initialized=%v, hook=%p, &ctx=%p, &Hooks=%p",
				ctx.HookIndex, hook.Type, hook.Value, hook.Initialized, hook, ctx, &ctx.Hooks)
		}
		if hook.Type != hookType {
			panic(fmt.Sprintf("hook order changed: expected %s, got %s at position %d",
				hook.Type, hookType, ctx.HookIndex))
		}
		ctx.HookIndex++
		return hook
	}

	// Create new hook
	hook := &Hook{
		Type: hookType,
	}
	ctx.Hooks = append(ctx.Hooks, *hook)
	log.UILogger.Debug("GetOrCreateHook: creating new hook[%d], Type=%s, &ctx=%p, &Hooks=%p",
		ctx.HookIndex, hookType, ctx, &ctx.Hooks)

	ctx.HookIndex++
	return &ctx.Hooks[len(ctx.Hooks)-1]
}

// =============================================================================
// Intent Support
// =============================================================================

// SetIntentRuntime sets the intent runtime for this context.
// This is called by DeclarativeNode during initialization.
func (ctx *ComponentContext) SetIntentRuntime(rt *intent.Runtime) {
	ctx.IntentRuntime = rt
}

// GetIntentRuntime returns the intent runtime for this context.
func (ctx *ComponentContext) GetIntentRuntime() *intent.Runtime {
	return ctx.IntentRuntime
}

// EmitIntent emits an intent through the intent runtime.
// This is the primary method for components to emit intents.
// Returns the result of dispatching the intent.
func (ctx *ComponentContext) EmitIntent(i intent.Intent) intent.IntentResult {
	if ctx.IntentRuntime == nil {
		// Fallback: if no runtime, log warning and mark as unhandled
		log.UILogger.IfEnabled().Debug("[ComponentContext] EmitIntent: no IntentRuntime set, intent=%s ignored", i.IntentType())
		return intent.ErrorResult(fmt.Errorf("IntentRuntime not initialized"))
	}
	return ctx.IntentRuntime.EmitFromSource(i, ctx.ComponentID)
}

// =============================================================================
// State Management (for Intent Handlers)
// =============================================================================

// SetState updates a state value by key with batch update support.
// Multiple SetState calls within the same render cycle will be batched
// into a single re-render for better performance.
// This implements the StateSetter interface for use by ActionContext.
// Example: ctx.SetState("username", "john")
//
// Deprecated: Use Store + Reducer architecture instead.
// Use UseStoreField for type-safe state management with automatic subscriptions.
// Migration guide: https://github.com/wwsheng009/mint/docs/ui/store/guides/MIGRATION_GUIDE.md
// Status: Will be removed in v1.0
func (ctx *ComponentContext) SetState(key string, value interface{}) {
	defer func() {
		// Recover from panics (e.g., uncomparable type comparison)
		if r := recover(); r != nil {
			log.UILogger.IfEnabled().Error("[ComponentContext] SetState panic recovered: %v (key=%s, value=%T)", r, key, value)
			// Schedule update anyway to ensure state is set
			if !ctx.UpdateScheduled && ctx.scheduleUpdate != nil {
				ctx.UpdateScheduled = true
				ctx.scheduleUpdate()
			}
		}
	}()

	ctx.StateMu.Lock()
	defer ctx.StateMu.Unlock()

	// Check if value actually changed
	// Use reflect.DeepEqual to handle uncomparable types (e.g., functions, slices, maps)
	changed := true
	if existing, exists := ctx.GlobalState[key]; exists {
		// Try direct comparison first for performance
		if existing == value {
			changed = false
		} else {
			// Fall back to DeepEqual for uncomparable types
			// This handles types like functions, slices, maps, etc.
			if reflect.DeepEqual(existing, value) {
				changed = false
			}
		}
	}

	if !changed {
		return
	}

	// Add to pending updates queue for batching
	ctx.PendingUpdates[key] = value

	log.UILogger.IfEnabled().Debug("[ComponentContext] SetState: %s = %v (queued)", key, value)

	// Schedule update only once per batch
	if !ctx.UpdateScheduled && ctx.scheduleUpdate != nil {
		ctx.UpdateScheduled = true
		ctx.scheduleUpdate()
	}
}

// GetState retrieves a state value by key.
// Returns (value, true) if key exists, (nil, false) otherwise.
// This implements the StateSetter interface for use by ActionContext.
//
// Deprecated: Use Store + Reducer architecture instead.
// Use UseStoreField for type-safe state management with automatic subscriptions.
// Migration guide: https://github.com/wwsheng009/mint/docs/ui/store/guides/MIGRATION_GUIDE.md
// Status: Will be removed in v1.0
func (ctx *ComponentContext) GetState(key string) (interface{}, bool) {
	ctx.StateMu.RLock()
	defer ctx.StateMu.RUnlock()

	// Check pending updates first (read-your-writes)
	if value, exists := ctx.PendingUpdates[key]; exists {
		return value, true
	}

	value, exists := ctx.GlobalState[key]
	return value, exists
}

// FlushUpdates applies all pending updates to GlobalState.
// This should be called before re-rendering to ensure the latest state is used.
// After flushing, the pending updates queue is cleared.
// Returns the number of updates applied.
func (ctx *ComponentContext) FlushUpdates() int {
	ctx.StateMu.Lock()
	defer ctx.StateMu.Unlock()

	if len(ctx.PendingUpdates) == 0 {
		return 0
	}

	count := len(ctx.PendingUpdates)
	for key, value := range ctx.PendingUpdates {
		ctx.GlobalState[key] = value
		log.UILogger.IfEnabled().Debug("[ComponentContext] FlushUpdates: %s = %v", key, value)
	}

	// Clear the pending queue
	ctx.PendingUpdates = make(map[string]interface{})
	ctx.UpdateScheduled = false

	return count
}

// GetStringState retrieves a string state value with a default.
//
// Deprecated: Use Store + Reducer architecture instead.
// Use UseStoreField for type-safe state management.
// Migration guide: https://github.com/wwsheng009/mint/docs/ui/store/guides/MIGRATION_GUIDE.md
// Status: Will be removed in v1.0
func (ctx *ComponentContext) GetStringState(key string, defaultValue string) string {
	if value, exists := ctx.GetState(key); exists {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return defaultValue
}

// GetIntState retrieves an int state value with a default.
//
// Deprecated: Use Store + Reducer architecture instead.
// Use UseStoreField for type-safe state management.
// Migration guide: https://github.com/wwsheng009/mint/docs/ui/store/guides/MIGRATION_GUIDE.md
// Status: Will be removed in v1.0
func (ctx *ComponentContext) GetIntState(key string, defaultValue int) int {
	if value, exists := ctx.GetState(key); exists {
		if i, ok := value.(int); ok {
			return i
		}
	}
	return defaultValue
}

// GetBoolState retrieves a boolean state value with a default.
//
// Deprecated: Use Store + Reducer architecture instead.
// Use UseStoreField for type-safe state management.
// Migration guide: https://github.com/wwsheng009/mint/docs/ui/store/guides/MIGRATION_GUIDE.md
// Status: Will be removed in v1.0
func (ctx *ComponentContext) GetBoolState(key string, defaultValue bool) bool {
	if value, exists := ctx.GetState(key); exists {
		if b, ok := value.(bool); ok {
			return b
		}
	}
	return defaultValue
}

// SetScheduleUpdate sets the callback to trigger Fiber re-render.
// This is called by DeclarativeNode during initialization.
func (ctx *ComponentContext) SetScheduleUpdate(fn func()) {
	ctx.scheduleUpdate = fn
}

// ScheduleUpdate triggers a Fiber re-render.
// This implements the StateSetter interface for use by ActionContext.
func (ctx *ComponentContext) ScheduleUpdate() {
	if ctx.scheduleUpdate != nil {
		ctx.scheduleUpdate()
	}
}

// =============================================================================
// Global Intent Runtime Management
// =============================================================================

// globalIntentRuntime is the global intent runtime shared across app.
// Set by ui.Run() when starting the declarative UI app.
var (
	globalIntentRuntime   *intent.Runtime
	globalIntentRuntimeMu sync.RWMutex
	pendingIntentMu       sync.Mutex
	pendingIntents        []*pendingIntent
)

type pendingIntent struct {
	mu         sync.Mutex
	register   func(*intent.Registry) func()
	unregister func()
	canceled   bool
}

// RegisterIntent registers an intent handler for a specific intent type.
// This is the public API for registering intent handlers in declarative UI components.
//
// Example:
//
//	RegisterIntent[OpenModalIntent](func(ctx *intent.ActionContext, i OpenModalIntent) intent.IntentResult {
//	    ctx.SetState("showModal", true)
//	    return intent.HandledResult()
//	})
func RegisterIntent[T intent.Intent](handler intent.TypedHandler[T]) func() {
	if handler == nil {
		return func() {}
	}

	globalIntentRuntimeMu.RLock()
	rt := globalIntentRuntime
	globalIntentRuntimeMu.RUnlock()
	if rt != nil {
		return intent.RegisterTyped(rt.Registry, handler)
	}

	pending := &pendingIntent{
		register: func(reg *intent.Registry) func() {
			return intent.RegisterTyped(reg, handler)
		},
	}

	pendingIntentMu.Lock()
	pendingIntents = append(pendingIntents, pending)
	pendingIntentMu.Unlock()

	log.UILogger.IfEnabled().Debug("[Intent] RegisterIntent queued until runtime is initialized")

	return func() {
		pending.mu.Lock()
		unregister := pending.unregister
		if unregister != nil {
			pending.mu.Unlock()
			unregister()
			return
		}
		if pending.canceled {
			pending.mu.Unlock()
			return
		}
		pending.canceled = true
		pending.mu.Unlock()

		pendingIntentMu.Lock()
		for i, item := range pendingIntents {
			if item == pending {
				pendingIntents = append(pendingIntents[:i], pendingIntents[i+1:]...)
				break
			}
		}
		pendingIntentMu.Unlock()
	}
}

// SetGlobalIntentRuntime sets the global intent runtime.
// This is called internally by ui.Run().
func SetGlobalIntentRuntime(rt *intent.Runtime) {
	globalIntentRuntimeMu.Lock()
	globalIntentRuntime = rt
	globalIntentRuntimeMu.Unlock()

	if rt == nil {
		return
	}

	pendingIntentMu.Lock()
	pending := pendingIntents
	pendingIntents = nil
	pendingIntentMu.Unlock()

	for _, item := range pending {
		item.mu.Lock()
		if item.canceled {
			item.mu.Unlock()
			continue
		}
		item.mu.Unlock()

		unregister := item.register(rt.Registry)
		item.mu.Lock()
		if item.canceled {
			item.mu.Unlock()
			unregister()
			continue
		}
		item.unregister = unregister
		item.mu.Unlock()
	}
}

// GetGlobalIntentRuntime returns the global intent runtime.
// This can be used to directly emit intents without a component context.
func GetGlobalIntentRuntime() *intent.Runtime {
	globalIntentRuntimeMu.RLock()
	defer globalIntentRuntimeMu.RUnlock()
	return globalIntentRuntime
}

// EmitIntentGlobal emits an intent through the global runtime.
// This can be used outside of component contexts.
func EmitIntentGlobal(i intent.Intent) intent.IntentResult {
	globalIntentRuntimeMu.RLock()
	rt := globalIntentRuntime
	globalIntentRuntimeMu.RUnlock()
	if rt == nil {
		return intent.ErrorResult(fmt.Errorf("Global IntentRuntime not initialized"))
	}
	return rt.Emit(i)
}
