package ui

import (
	"fmt"
	"sync"
)

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
	Type     HookType
	Value    interface{}
	SetValue func(interface{})
	Deps     []interface{}
	Cleanup  func()
}

// ComponentContext holds the state for a component during rendering
type ComponentContext struct {
	ComponentID string
	Hooks       []Hook
	HookIndex   int
	Validator   *HookValidator
	RenderCount int
}

// Global context for the currently rendering component
var (
	currentContext      *ComponentContext
	currentContextMu    sync.RWMutex
	globalComponentID   int
	globalComponentIDMu sync.Mutex
)

// nextComponentID generates a unique component ID
func nextComponentID() string {
	globalComponentIDMu.Lock()
	defer globalComponentIDMu.Unlock()
	globalComponentID++
	return fmt.Sprintf("comp-%d", globalComponentID)
}

// setCurrentContext sets the current rendering context
func setCurrentContext(ctx *ComponentContext) {
	currentContextMu.Lock()
	defer currentContextMu.Unlock()
	currentContext = ctx
}

// getCurrentContext returns the current rendering context
func getCurrentContext() *ComponentContext {
	currentContextMu.RLock()
	defer currentContextMu.RUnlock()
	return currentContext
}

// useState creates a state hook
// Usage: count, setCount := useState(0)
func useState(initial interface{}) (interface{}, func(interface{})) {
	ctx := getCurrentContext()
	if ctx == nil {
		panic("useState must be called within a component")
	}

	// Validate hook call
	if err := ctx.Validator.ValidateHookCall(HookState); err != nil {
		panic(err)
	}

	// Get or create hook
	hook := ctx.getOrCreateHook(HookState)

	// Initialize if first render
	if hook.Value == nil {
		hook.Value = initial
	}

	// Create setter function
	setState := func(newValue interface{}) {
		hook.Value = newValue
		// Schedule re-render (will be implemented in reconciler)
		scheduleRender(ctx.ComponentID)
	}

	hook.SetValue = setState

	return hook.Value, setState
}

// UseStateInt is a type-safe version of useState for int
// Returns: (currentValue, setValue, getValue)
// Use getValue() in event handlers to get the latest value
func UseStateInt(initial int) (int, func(int), func() int) {
	value, setValue := useState(initial)
	// Capture the hook for getValue
	ctx := getCurrentContext()
	hookIndex := ctx.HookIndex - 1 // The hook was just created/accessed

	getValue := func() int {
		if hookIndex < len(ctx.Hooks) {
			return ctx.Hooks[hookIndex].Value.(int)
		}
		return initial
	}

	return value.(int), func(newValue int) {
		setValue(newValue)
	}, getValue
}

// UseStateString is a type-safe version of useState for string
func UseStateString(initial string) (string, func(string)) {
	value, setValue := useState(initial)
	return value.(string), func(newValue string) {
		setValue(newValue)
	}
}

// UseStateBool is a type-safe version of useState for bool
func UseStateBool(initial bool) (bool, func(bool)) {
	value, setValue := useState(initial)
	return value.(bool), func(newValue bool) {
		setValue(newValue)
	}
}

// getOrCreateHook gets an existing hook or creates a new one
func (ctx *ComponentContext) getOrCreateHook(hookType HookType) *Hook {
	if ctx.HookIndex < len(ctx.Hooks) {
		hook := &ctx.Hooks[ctx.HookIndex]
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
	ctx.HookIndex++
	return &ctx.Hooks[len(ctx.Hooks)-1]
}

// scheduleRender schedules a re-render by marking the app as dirty
func scheduleRender(componentID string) {
	// Access the global app instance to mark it dirty
	if appInstance != nil {
		appInstance.MarkDirty()
	}
}

// newComponentContext creates a new component context
func newComponentContext(name string) *ComponentContext {
	return &ComponentContext{
		ComponentID: fmt.Sprintf("%s-%s", name, nextComponentID()),
		Hooks:       make([]Hook, 0),
		HookIndex:   0,
		Validator:   NewHookValidator(name),
		RenderCount: 0,
	}
}

// resetContext resets the hook index for re-rendering
func (ctx *ComponentContext) resetContext() {
	ctx.HookIndex = 0
	ctx.RenderCount++
}

// finishRender finishes the render and validates hooks
func (ctx *ComponentContext) finishRender() error {
	return ctx.Validator.FinishRender()
}
