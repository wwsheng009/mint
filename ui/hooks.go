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

// IntSetter is the type for setting int state
// Can be either an int value or a function that takes the previous value and returns a new one
type IntSetter interface{}

// SetIntFunc is a function that computes the new value from the old value
type SetIntFunc func(int) int

// UseStateInt is a type-safe version of useState for int
// Returns: (currentValue, setValue, getValue)
// Use getValue() in event handlers to get the latest value
//
// The setter accepts either an int value or a SetIntFunc:
//   setCount(5)                    // Set directly
//   setCount(func(c int) int {     // Functional update
//       return c + 1
//   })
func UseStateInt(initial int) (int, func(interface{}), func() int) {
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

	setInt := func(newValue interface{}) {
		switch v := newValue.(type) {
		case int:
			setValue(v)
		case SetIntFunc:
			// Functional update - get current value and apply function
			current := getValue()
			setValue(v(current))
		case func(int) int:
			// Also support raw func(int) int
			current := getValue()
			setValue(v(current))
		default:
			setValue(newValue)
		}
	}

	return value.(int), setInt, getValue
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

// =============================================================================
// useEffect Hook
// =============================================================================

// EffectCallback is the function passed to useEffect
type EffectCallback func() CleanupFunc

// CleanupFunc is the optional cleanup function returned by EffectCallback
type CleanupFunc func()

// useEffect runs side effects after render
// deps: nil = run once (mount), [] = run every render, [values] = run when values change
//
// Example:
//
//	count, setCount := UseStateInt(0)
//
//	UseEffect(func() func() {
//	    ticker := time.NewTicker(time.Second)
//	    go func() {
//	        for range ticker.C {
//	            setCount(func() int { return count + 1 })
//	        }
//	    }()
//	    return func() { ticker.Stop() } // cleanup
//	}, nil) // nil deps = run once on mount
func useEffect(callback EffectCallback, deps []interface{}) {
	ctx := getCurrentContext()
	if ctx == nil {
		panic("useEffect must be called within a component")
	}

	// Validate hook call
	if err := ctx.Validator.ValidateHookCall(HookEffect); err != nil {
		panic(err)
	}

	// Get or create hook
	hook := ctx.getOrCreateHook(HookEffect)

	// Check if dependencies changed
	shouldRun := false
	if !hook.Initialized {
		// First render - always run
		shouldRun = true
		hook.Initialized = true
	} else if deps == nil {
		// Previous effect had nil deps (run once) - don't run again
		shouldRun = false
	} else if len(deps) == 0 {
		// Empty deps array - run on every render
		shouldRun = true
	} else {
		// Check if dependencies changed
		shouldRun = !depsEqual(hook.Deps, deps)
	}

	if shouldRun {
		// Run cleanup from previous effect
		if hook.Cleanup != nil {
			hook.Cleanup()
			hook.Cleanup = nil
		}

		// Store callback for execution after render
		hook.Value = callback
		hook.Deps = deps
	}
	// If not running, don't set hook.Value - effect won't run again
}

// UseEffect is the public API for useEffect
func UseEffect(callback EffectCallback, deps []interface{}) {
	useEffect(callback, deps)
}

// runEffects executes all pending effects after render
// This should be called by the reconciler after committing changes
func (ctx *ComponentContext) runEffects() {
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

// cleanupAll runs all cleanup functions (for unmounting)
func (ctx *ComponentContext) cleanupAll() {
	for i := range ctx.Hooks {
		hook := &ctx.Hooks[i]
		if hook.Cleanup != nil {
			hook.Cleanup()
			hook.Cleanup = nil
		}
	}
}

// depsEqual compares two dependency arrays
func depsEqual(a, b []interface{}) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// =============================================================================
// useRef Hook
// =============================================================================

// Ref holds a mutable value that persists across renders
type Ref struct {
	Value interface{}
}

// useRef creates a ref that persists across renders
// Useful for:
// - Accessing DOM nodes (future)
// - Storing mutable values without triggering re-renders
// - Holding values that need to persist across effect runs
//
// Example:
//
//	countRef := UseRef(0)
//	// Update ref.Value directly - no re-render
//	countRef.Value = countRef.Value.(int) + 1
func UseRef(initial interface{}) *Ref {
	ctx := getCurrentContext()
	if ctx == nil {
		panic("useRef must be called within a component")
	}

	// Validate hook call
	if err := ctx.Validator.ValidateHookCall(HookRef); err != nil {
		panic(err)
	}

	// Get or create hook
	hook := ctx.getOrCreateHook(HookRef)

	// Initialize if first render
	if hook.Value == nil {
		hook.Value = &Ref{Value: initial}
	}

	return hook.Value.(*Ref)
}

// =============================================================================
// useMemo Hook
// =============================================================================

// useMemo memoizes a computed value
// Only recomputes when dependencies change
//
// Example:
//
//	expensiveValue := UseMemo(func() interface{} {
//	    return computeExpensive(a, b)
//	}, []interface{}{a, b})
func UseMemo(compute func() interface{}, deps []interface{}) interface{} {
	ctx := getCurrentContext()
	if ctx == nil {
		panic("useMemo must be called within a component")
	}

	// Validate hook call
	if err := ctx.Validator.ValidateHookCall(HookMemo); err != nil {
		panic(err)
	}

	// Get or create hook
	hook := ctx.getOrCreateHook(HookMemo)

	// Check if dependencies changed or first render
	shouldCompute := false
	if !hook.Initialized {
		// First time - compute
		shouldCompute = true
		hook.Initialized = true
	} else if deps == nil {
		// Previous had nil deps (compute once) - don't recompute
		shouldCompute = false
	} else if len(deps) == 0 {
		// Empty deps - always compute
		shouldCompute = true
	} else {
		// Check if dependencies changed
		shouldCompute = !depsEqual(hook.Deps, deps)
	}

	if shouldCompute {
		hook.Value = compute()
		hook.Deps = deps
	}

	return hook.Value
}

// =============================================================================
// useCallback Hook
// =============================================================================

// useCallback memoizes a function
// Only creates new function when dependencies change
// Equivalent to useMemo(() => callback, deps)
//
// Example:
//
//	handleClick := UseCallback(func() {
//	    setCount(count + 1)
//	}, []interface{}{count})
func UseCallback(callback func(), deps []interface{}) func() {
	return UseMemo(func() interface{} {
		return callback
	}, deps).(func())
}
