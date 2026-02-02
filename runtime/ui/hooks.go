package ui

import (
	"fmt"
	"os"
	"sync"
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
	}
}

// NewComponentContextForRoot creates a new component context for the root
func NewComponentContextForRoot() *ComponentContext {
	return NewComponentContext("App")
}

// ResetContext resets the hook index for re-rendering
func (ctx *ComponentContext) ResetContext() {
	if os.Getenv("TUI_DEBUG_UI") == "true" {
		if len(ctx.Hooks) > 0 {
			fmt.Fprintf(os.Stderr, "ResetContext: BEFORE reset, Hooks[0].Value=%v, &ctx=%p\n", ctx.Hooks[0].Value, ctx)
		}
	}
	ctx.HookIndex = 0
	ctx.RenderCount++
	if os.Getenv("TUI_DEBUG_UI") == "true" {
		if len(ctx.Hooks) > 0 {
			fmt.Fprintf(os.Stderr, "ResetContext: AFTER reset, Hooks[0].Value=%v, &ctx=%p\n", ctx.Hooks[0].Value, ctx)
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
		if os.Getenv("TUI_DEBUG_UI") == "true" {
			fmt.Fprintf(os.Stderr, "GetOrCreateHook: returning existing hook[%d], Type=%s, Value=%v, Initialized=%v, hook=%p, &ctx=%p, &Hooks=%p\n",
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
		fmt.Fprintf(os.Stderr, "GetOrCreateHook: creating new hook[%d], Type=%s, &ctx=%p, &Hooks=%p\n",
			ctx.HookIndex, hookType, ctx, &ctx.Hooks)
	}
	ctx.HookIndex++
	return &ctx.Hooks[len(ctx.Hooks)-1]
}
