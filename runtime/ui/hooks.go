package ui

import (
	"fmt"
	"os"
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

	// State is the component's global state store, accessible via Intent handlers.
	// This allows Intent handlers to update state without closures.
	// Example: ctx.SetState("username", "john")
	State map[string]interface{}

	// scheduleUpdate is a callback to trigger Fiber re-render.
	// Set by DeclarativeNode during initialization.
	scheduleUpdate func()

	// StateMu protects concurrent access to State.
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
		ComponentID: fmt.Sprintf("%s-%s", name, NextComponentID()),
		Hooks:       make([]Hook, 0),
		HookIndex:   0,
		Validator:   NewHookValidator(name),
		RenderCount: 0,
		State:       make(map[string]interface{}),
	}
}

// NewComponentContextForRoot creates a new component context for the root
func NewComponentContextForRoot() *ComponentContext {
	return NewComponentContext("App")
}

// ResetContext resets the hook index for re-rendering
func (ctx *ComponentContext) ResetContext() {
	if log.UILogger.Enabled() {
		if len(ctx.Hooks) > 0 {
			log.UILogger.Debug("ResetContext: BEFORE reset, Hooks[0].Value=%v, &ctx=%p", ctx.Hooks[0].Value, ctx)
		}
	}
	ctx.HookIndex = 0
	ctx.RenderCount++
	if log.UILogger.Enabled() {
		if len(ctx.Hooks) > 0 {
			log.UILogger.Debug("ResetContext: AFTER reset, Hooks[0].Value=%v, &ctx=%p", ctx.Hooks[0].Value, ctx)
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
	if os.Getenv("TUI_DEBUG_UI") == "true" {
		log.UILogger.Debug("GetOrCreateHook: creating new hook[%d], Type=%s, &ctx=%p, &Hooks=%p",
			ctx.HookIndex, hookType, ctx, &ctx.Hooks)
	}
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
		log.UILogger.Debug("[ComponentContext] EmitIntent: no IntentRuntime set, intent=%s ignored", i.IntentType())
		return intent.ErrorResult(fmt.Errorf("IntentRuntime not initialized"))
	}
	return ctx.IntentRuntime.EmitFromSource(i, ctx.ComponentID)
}

// =============================================================================
// State Management (for Intent Handlers)
// =============================================================================

// SetState updates a state value by key.
// This implements the StateSetter interface for use by ActionContext.
// Example: ctx.SetState("username", "john")
func (ctx *ComponentContext) SetState(key string, value interface{}) {
	ctx.StateMu.Lock()
	defer ctx.StateMu.Unlock()

	changed := true
	if existing, exists := ctx.State[key]; exists && existing == value {
		changed = false
	}

	if changed {
		ctx.State[key] = value
		if log.UILogger.Enabled() {
			log.UILogger.Debug("[ComponentContext] SetState: %s = %v", key, value)
		}

		// Trigger Fiber update if schedule callback is set
		if ctx.scheduleUpdate != nil {
			ctx.scheduleUpdate()
		}
	}
}

// GetState retrieves a state value by key.
// Returns (value, true) if key exists, (nil, false) otherwise.
// This implements the StateSetter interface for use by ActionContext.
func (ctx *ComponentContext) GetState(key string) (interface{}, bool) {
	ctx.StateMu.RLock()
	defer ctx.StateMu.RUnlock()
	value, exists := ctx.State[key]
	return value, exists
}

// GetStringState retrieves a string state value with a default.
func (ctx *ComponentContext) GetStringState(key string, defaultValue string) string {
	if value, exists := ctx.GetState(key); exists {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return defaultValue
}

// GetIntState retrieves an int state value with a default.
func (ctx *ComponentContext) GetIntState(key string, defaultValue int) int {
	if value, exists := ctx.GetState(key); exists {
		if i, ok := value.(int); ok {
			return i
		}
	}
	return defaultValue
}

// GetBoolState retrieves a boolean state value with a default.
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
var globalIntentRuntime *intent.Runtime

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
	if globalIntentRuntime == nil {
		panic("RegisterIntent called before app initialized. Call ui.Run() first.")
	}
	return intent.RegisterTyped(globalIntentRuntime.Registry, handler)
}

// SetGlobalIntentRuntime sets the global intent runtime.
// This is called internally by ui.Run().
func SetGlobalIntentRuntime(rt *intent.Runtime) {
	globalIntentRuntime = rt
}

// GetGlobalIntentRuntime returns the global intent runtime.
// This can be used to directly emit intents without a component context.
func GetGlobalIntentRuntime() *intent.Runtime {
	return globalIntentRuntime
}

// EmitIntentGlobal emits an intent through the global runtime.
// This can be used outside of component contexts.
func EmitIntentGlobal(i intent.Intent) intent.IntentResult {
	if globalIntentRuntime == nil {
		return intent.ErrorResult(fmt.Errorf("Global IntentRuntime not initialized"))
	}
	return globalIntentRuntime.Emit(i)
}
