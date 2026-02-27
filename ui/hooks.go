package ui

import (
	"fmt"
	"strings"

	"github.com/wwsheng009/mint/internal/log"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Hook Types (re-exported from runtime/ui)
// =============================================================================

// HookType represents the type of hook
type HookType = rtui.HookType

const (
	// HookState is the useState hook type
	HookState = rtui.HookState
	// HookEffect is the useEffect hook type
	HookEffect = rtui.HookEffect
	// HookContext is the useContext hook type
	HookContext = rtui.HookContext
	// HookMemo is the useMemo/useCallback hook type
	HookMemo = rtui.HookMemo
	// HookRef is the useRef hook type
	HookRef = rtui.HookRef
)

// ComponentContext holds the state for a component during rendering
type ComponentContext = rtui.ComponentContext

// Ref holds a mutable value that persists across renders
type Ref = rtui.Ref

// EffectCallback is the function passed to useEffect
// It returns an optional cleanup function
type EffectCallback = rtui.EffectCallback

// CleanupFunc is the optional cleanup function returned by EffectCallback
type CleanupFunc = rtui.CleanupFunc

// =============================================================================
// Context Management (forwarded to runtime/ui)
// =============================================================================

// SetCurrentContext sets the current rendering context
func SetCurrentContext(ctx *ComponentContext) {
	rtui.SetCurrentContext(ctx)
}

// GetCurrentContext returns the current rendering context
func GetCurrentContext() *ComponentContext {
	return rtui.GetCurrentContext()
}

// NewComponentContextForRoot creates a new component context for the root
func NewComponentContextForRoot() *ComponentContext {
	return rtui.NewComponentContextForRoot()
}

// =============================================================================
// useState Hook
// =============================================================================

// useState creates a state hook
// Usage: count, setCount := useState(0)
func useState(initial interface{}) (interface{}, func(interface{})) {
	ctx := rtui.GetCurrentContext()
	if ctx == nil {
		panic("useState must be called within a component")
	}

	// Validate hook call
	if err := ctx.Validator.ValidateHookCall(rtui.HookState); err != nil {
		panic(err)
	}

	// Capture the hook index BEFORE getOrCreateHook increments it
	hookIndex := ctx.HookIndex

	// Get or create hook
	hook := ctx.GetOrCreateHook(rtui.HookState)

	// Initialize if first render (not just if Value is nil)
	if !hook.Initialized {
		hook.Value = initial
		hook.Initialized = true
	}

	// Get the current value to return
	currentValue := hook.Value

	if log.UILogger.Enabled() {
		log.UILogger.Debug("useState: componentID=%s, hookIndex=%d, value=%v, hook=%p, &ctx.Hooks[%d]=%p, &ctx=%p",
			ctx.ComponentID, hookIndex, currentValue, hook, hookIndex, &ctx.Hooks[hookIndex], ctx)
	}

	// Create setter function that captures ctx and hookIndex (not the hook pointer)
	// This ensures we always access the correct hook even if the slice is reallocated
	setState := func(newValue interface{}) {
		if log.UILogger.Enabled() {
			log.UILogger.Debug("setState BEFORE: componentID=%s, hookIndex=%d, value=%v, &ctx=%p, &ctx.Hooks=%p",
				ctx.ComponentID, hookIndex, ctx.Hooks[hookIndex].Value, ctx, &ctx.Hooks)
		}
		// Use index to access hook - this is safe even if slice was reallocated
		if hookIndex < len(ctx.Hooks) {
			ctx.Hooks[hookIndex].Value = newValue
			if log.UILogger.Enabled() {
				log.UILogger.Debug("setState AFTER: componentID=%s, hookIndex=%d, value=%v, &ctx=%p",
					ctx.ComponentID, hookIndex, ctx.Hooks[hookIndex].Value, ctx)
			}
		}
		// Schedule re-render
		scheduleRender(ctx.ComponentID)
	}

	// Also update SetValue on the hook for compatibility
	if hookIndex < len(ctx.Hooks) {
		ctx.Hooks[hookIndex].SetValue = setState
	}

	return currentValue, setState
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
//
//	setCount(5)                    // Set directly
//	setCount(func(c int) int {     // Functional update
//	    return c + 1
//	})
func UseStateInt(initial int) (int, func(interface{}), func() int) {
	// Get context BEFORE calling useState (useState will increment HookIndex)
	ctx := rtui.GetCurrentContext()
	hookIndex := ctx.HookIndex // Capture index before useState increments it

	value, setValue := useState(initial)

	// getValue uses the captured hookIndex to always get the latest value
	getValue := func() int {
		if hookIndex < len(ctx.Hooks) {
			if v, ok := ctx.Hooks[hookIndex].Value.(int); ok {
				return v
			}
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

// scheduleRender schedules a re-render by marking the app as dirty
func scheduleRender(componentID string) {
	// Access the global app instance to mark it dirty
	log.UILogger.Debug("scheduleRender: componentID=%s, appInstance=%v", componentID, appInstance != nil)
	if appInstance != nil {
		appInstance.MarkDirty()
		log.UILogger.Debug("scheduleRender: MarkDirty() called, state=%v", appInstance.GetState())
	} else {
		log.UILogger.Debug("scheduleRender: appInstance is nil, cannot MarkDirty()")
	}
}

// =============================================================================
// useEffect Hook
// =============================================================================

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
	ctx := rtui.GetCurrentContext()
	if ctx == nil {
		panic("useEffect must be called within a component")
	}

	// Validate hook call
	if err := ctx.Validator.ValidateHookCall(rtui.HookEffect); err != nil {
		panic(err)
	}

	// Get or create hook
	hook := ctx.GetOrCreateHook(rtui.HookEffect)

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

// =============================================================================
// useRef Hook
// =============================================================================

// UseRef creates a ref that persists across renders
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
	ctx := rtui.GetCurrentContext()
	if ctx == nil {
		panic("useRef must be called within a component")
	}

	// Validate hook call
	if err := ctx.Validator.ValidateHookCall(rtui.HookRef); err != nil {
		panic(err)
	}

	// Get or create hook
	hook := ctx.GetOrCreateHook(rtui.HookRef)

	// Initialize if first render
	if hook.Value == nil {
		hook.Value = &Ref{Value: initial}
	}

	return hook.Value.(*Ref)
}

// =============================================================================
// useMemo Hook
// =============================================================================

// UseMemo memoizes a computed value
// Only recomputes when dependencies change
//
// Example:
//
//	expensiveValue := UseMemo(func() interface{} {
//	    return computeExpensive(a, b)
//	}, []interface{}{a, b})
func UseMemo(compute func() interface{}, deps []interface{}) interface{} {
	ctx := rtui.GetCurrentContext()
	if ctx == nil {
		panic("useMemo must be called within a component")
	}

	// Validate hook call
	if err := ctx.Validator.ValidateHookCall(rtui.HookMemo); err != nil {
		panic(err)
	}

	// Get or create hook
	hook := ctx.GetOrCreateHook(rtui.HookMemo)

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

// UseCallback memoizes a function
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

// =============================================================================
// useHoverState Hook
// =============================================================================

// HoverStateChange is called when hover state changes
type HoverStateChange func(bool)

// useHoverState creates a hover state hook that persists across renders
// It uses a ref internally to maintain state without triggering re-renders
//
// This is useful for interactive elements that need to track hover state
// without causing full re-renders on every mouse move.
//
// Example:
//
//	isHovered, setHovered := UseHoverState()
//	isHovered()  // Returns current hover state
//	setHovered(true)  // Sets hover state
func useHoverState() (func() bool, func(bool)) {
	ctx := rtui.GetCurrentContext()
	if ctx == nil {
		panic("useHoverState must be called within a component")
	}

	// Use ref to persist state across renders
	ref := UseRef(false)

	// Getter returns current hover state
	getter := func() bool {
		if ref.Value == nil {
			return false
		}
		return ref.Value.(bool)
	}

	// Setter updates hover state without triggering re-render
	// The actual visual update will happen during next paint cycle
	setter := func(hovered bool) {
		ref.Value = hovered
		// We don't trigger a full re-render for hover changes
		// Instead, the component will check hover state during paint
	}

	return getter, setter
}

// UseHoverState is the public API for useHoverState
// Returns:
//   - getter: func() bool - returns current hover state
//   - setter: func(bool) - sets hover state
func UseHoverState() (func() bool, func(bool)) {
	return useHoverState()
}

// UseStateIntWithDebug is UseStateInt with additional debug output
// Returns:
//   - value: int - current state value
//   - setValue: func(interface{}) - sets new value
//   - getValue: func() int - gets current value
//   - hookIndex: int - the hook index (for debugging)
func UseStateIntWithDebug(initial int) (int, func(interface{}), func() int, int) {
	// Get context BEFORE calling useState (useState will increment HookIndex)
	ctx := rtui.GetCurrentContext()
	hookIndex := ctx.HookIndex // Capture index before useState increments it

	log.UILogger.Debug("[DEBUG] UseStateIntWithDebug: initial=%d, hookIndex=%d", initial, hookIndex)

	value, setValue := useState(initial)

	// getValue uses the captured hookIndex to always get the latest value
	getValue := func() int {
		if hookIndex < len(ctx.Hooks) {
			if v, ok := ctx.Hooks[hookIndex].Value.(int); ok {
				log.UILogger.Debug("[DEBUG] getValue: hookIndex=%d, value=%d, hook=%p",
					hookIndex, v, &ctx.Hooks[hookIndex])

				return v
			}
		}
		return initial
	}

	setInt := func(newValue interface{}) {
		switch v := newValue.(type) {
		case int:
			log.UILogger.Debug("[DEBUG] setInt: hookIndex=%d, oldValue=%v, newValue=%d",
				hookIndex, value, v)
			setValue(v)
		case func(int) int:
			// Also support raw func(int) int
			current := getValue()
			newVal := v(current)
			log.UILogger.Debug("[DEBUG] setInt(fn): hookIndex=%d, oldValue=%d, newValue=%d",
				hookIndex, current, newVal)
			setValue(newVal)
		default:
			setValue(newValue)
		}
	}

	intValue := value.(int)
	log.UILogger.Debug("[DEBUG] UseStateIntWithDebug RETURN: hookIndex=%d, returning value=%d, ptr=%p", hookIndex, intValue, &intValue)

	return intValue, setInt, getValue, hookIndex
}

// ==============================================================================
// Debug Functions for Sandbox Testing
// ==============================================================================

// DebugContextInfo returns debug information about the current context
func DebugContextInfo() map[string]interface{} {
	ctx := rtui.GetCurrentContext()

	if ctx == nil {
		return map[string]interface{}{
			"hasContext": false,
		}
	}

	hooksInfo := make([]map[string]interface{}, len(ctx.Hooks))
	for i, hook := range ctx.Hooks {
		hooksInfo[i] = map[string]interface{}{
			"type":    hook.Type.String(),
			"value":   hook.Value,
			"pointer": fmt.Sprintf("%p", &hook),
		}
	}

	return map[string]interface{}{
		"hasContext":     true,
		"componentID":    ctx.ComponentID,
		"hooksCount":     len(ctx.Hooks),
		"hookIndex":      ctx.HookIndex,
		"renderCount":    ctx.RenderCount,
		"hooks":          hooksInfo,
		"contextPointer": fmt.Sprintf("%p", ctx),
	}
}

// DebugHooksState returns a detailed dump of hooks state
func DebugHooksState() string {
	ctx := rtui.GetCurrentContext()

	if ctx == nil {
		return "No current context"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Context: %s\n", ctx.ComponentID))
	sb.WriteString(fmt.Sprintf("HookIndex: %d\n", ctx.HookIndex))
	sb.WriteString(fmt.Sprintf("Hooks Count: %d\n", len(ctx.Hooks)))
	sb.WriteString("Hooks:\n")

	for i, hook := range ctx.Hooks {
		sb.WriteString(fmt.Sprintf("  [%d] Type=%s, Value=%v, Ptr=%p\n",
			i, hook.Type, hook.Value, &hook))
	}

	return sb.String()
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
