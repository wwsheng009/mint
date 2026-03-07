package ui

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/wwsheng009/mint/internal/log"
	"github.com/wwsheng009/mint/runtime/intent"
	strstore "github.com/wwsheng009/mint/runtime/store"
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
// useState Hook - DEPRECATED
// =============================================================================

// Deprecated: Use Store + Reducer architecture instead.
//
// useState is deprecated in favor of the Store + Reducer architecture.
// Using Store + Reducer provides better code organization and predictability.
//
// Migration Guide:
//   1. Define your application state as a struct
//   2. Create a global store: `appStore := store.NewStore(AppState{})`
//   3. Define a reducer to handle state changes
//   4. Use `state := appStore.Get()` to read state in components
//
// Example:
//
// Before (Old):
//   username, setUsername := ui.UseStateString("")
//   ctx.GlobalState["setter"] = setUsername
//
// After (New):
//   type AppState struct { Username string }
//   appStore := store.NewStore(AppState{Username: ""})
//   // In component:
//   state := appStore.Get()
//
// See migration guide: docs/architecture/store/MIGRATION_GUIDE.md
// See examples/store_reducer_demo/main.go for usage example

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
//
// Deprecated: Use Store + Reducer architecture instead.
// Migration Guide: See docs/architecture/store/MIGRATION_GUIDE.md
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
//
// Deprecated: Use Store + Reducer architecture instead.
// Migration Guide: See docs/architecture/store/MIGRATION_GUIDE.md
func UseStateString(initial string) (string, func(string)) {
	value, setValue := useState(initial)
	return value.(string), func(newValue string) {
		setValue(newValue)
	}
}

// UseStateBool is a type-safe version of useState for bool
//
// Deprecated: Use Store + Reducer architecture instead.
// Migration Guide: See docs/architecture/store/MIGRATION_GUIDE.md
func UseStateBool(initial bool) (bool, func(bool)) {
	value, setValue := useState(initial)
	return value.(bool), func(newValue bool) {
		setValue(newValue)
	}
}

// scheduleRender schedules a re-render by marking the app as dirty
func scheduleRender(componentID string) {
	// Access the global app instance to mark it dirty
	log.UILogger.IfEnabled().Debug("scheduleRender: componentID=%s, appInstance=%v", componentID, appInstance != nil)
	if appInstance != nil {
		appInstance.MarkDirty()
		log.UILogger.IfEnabled().Debug("scheduleRender: MarkDirty() called, state=%v", appInstance.GetState())
	} else {
		log.UILogger.IfEnabled().Debug("scheduleRender: appInstance is nil, cannot MarkDirty()")
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
//
// Deprecated: Use Store + Reducer architecture instead.
// Migration Guide: See docs/architecture/store/MIGRATION_GUIDE.md
func UseStateIntWithDebug(initial int) (int, func(interface{}), func() int, int) {
	// Get context BEFORE calling useState (useState will increment HookIndex)
	ctx := rtui.GetCurrentContext()
	hookIndex := ctx.HookIndex // Capture index before useState increments it

	log.UILogger.IfEnabled().Debug("[DEBUG] UseStateIntWithDebug: initial=%d, hookIndex=%d", initial, hookIndex)

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
	log.UILogger.IfEnabled().Debug("[DEBUG] UseStateIntWithDebug RETURN: hookIndex=%d, returning value=%d, ptr=%p", hookIndex, intValue, &intValue)

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

// deepEqual performs deep equality comparison for two values.
// It supports:
//   - Nil comparisons
//   - Basic types (int, float, string, bool, etc.)
//   - Slices (recursive)
//   - Maps (recursive)
//   - Structs (by comparing exported fields)
//   - Pointers (comparing pointed values)
//
// This is used by UseStoreSelector and UseStoreField to determine if a value
// has truly changed, avoiding unnecessary re-renders.
//
// Limitations:
//   - Struct comparison uses reflection (may have performance overhead)
//   - Unexported struct fields are not compared
//   - Circular references may cause stack overflow (unlikely in typical use)
func deepEqual(a, b interface{}) bool {
	// Handle nil cases first
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// Use reflection for complex types
	va := reflect.ValueOf(a)
	vb := reflect.ValueOf(b)

	// If types don't match, they're not equal
	if va.Kind() != vb.Kind() {
		return false
	}

	// Handle slices and maps specially (nil vs empty is different)
	switch va.Kind() {
	case reflect.Slice, reflect.Map:
		// nil vs empty is different
		if va.IsNil() != vb.IsNil() {
			return false
		}
	}

	return deepEqualValue(va, vb)
}

// deepEqualValue recursively compares two reflect.Values
func deepEqualValue(a, b reflect.Value) bool {
	// Handle basic types efficiently
	switch a.Kind() {
	case reflect.Bool:
		return a.Bool() == b.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return a.Int() == b.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return a.Uint() == b.Uint()
	case reflect.Float32, reflect.Float64:
		return a.Float() == b.Float()
	case reflect.Complex64, reflect.Complex128:
		return a.Complex() == b.Complex()
	case reflect.String:
		return a.String() == b.String()
	}

	// Handle pointers and interfaces
	switch a.Kind() {
	case reflect.Ptr:
		// nil check
		if a.IsNil() != b.IsNil() {
			return false
		}
		if a.IsNil() {
			return true
		}
		// For pointers, we compare the pointed values (deep comparison)
		return deepEqualValue(a.Elem(), b.Elem())

	case reflect.Interface:
		if a.IsNil() != b.IsNil() {
			return false
		}
		if a.IsNil() {
			return true
		}
		return deepEqualValue(a.Elem(), b.Elem())

	case reflect.Slice, reflect.Array:
		if a.Len() != b.Len() {
			return false
		}
		for i := 0; i < a.Len(); i++ {
			if !deepEqualValue(a.Index(i), b.Index(i)) {
				return false
			}
		}
		return true

	case reflect.Map:
		if a.Len() != b.Len() {
			return false
		}
		// Compare keys and values
		for _, key := range a.MapKeys() {
			va := a.MapIndex(key)
			vb := b.MapIndex(key)
			if !vb.IsValid() {
				return false // Key doesn't exist in b
			}
			if !deepEqualValue(va, vb) {
				return false
			}
		}
		return true

	case reflect.Struct:
		// Compare exported struct fields
		t := a.Type()
		for i := 0; i < a.NumField(); i++ {
			field := t.Field(i)
			// Skip unexported fields
			if field.PkgPath != "" {
				continue
			}
			if !deepEqualValue(a.Field(i), b.Field(i)) {
				return false
			}
		}
		return true

	case reflect.Func:
		// Functions are equal only if they have the same address or both nil
		return a.Pointer() == b.Pointer()

	default:
		// For other types, try direct equality
		return a.Interface() == b.Interface()
	}
}

// =============================================================================
// Store Hooks - Hybrid Mode (useState + Store Architecture)
// =============================================================================

// These hooks provide React-style useState API while using the Store + Reducer
// architecture underneath. This enables a gradual migration path from the old
// useState-based state management to the superior Store + Reducer pattern.
//
// Benefits:
//   - Simple useState-like API for component-level usage
//   - Unified Store architecture for app-level state management
//   - Type-safe with Go generics
//   - No closure capture issues
//   - Automatic subscription and re-rendering

// UseStoreField subscribes to a specific field from a Store.
//
// It provides useState-like API while using the Store architecture:
//   - currentValue: Current field value
//   - setter: Function to update the field (triggers Store update)
//
// This is useful for migrating from useState to Store gradually, or when you
// prefer the hook-style API in components while maintaining Store architecture.
//
// Example:
//
//	type AppState struct {
//	    Username string
//	    Count    int
//	}
//	appStore := store.NewStore(AppState{})
//
//	// In component:
//	username, setUsername := ui.UseStoreField(
//	    appStore,
//	    func(s AppState) string { return s.Username },
//	    func(s AppState, v string) AppState { s.Username = v; return s },
//	)
//
// Usage:
//   setFieldValue(newValue)        // Set directly
//
// Note: This auto-subscribes to Store changes and triggers re-render.
func UseStoreField[S any, T any](
	store *strstore.Store[S],
	selector func(S) T,
	setter func(S, T) S,
) (T, func(T)) {
	ctx := rtui.GetCurrentContext()
	if ctx == nil {
		panic("UseStoreField must be called within a component")
	}

	// Get current value from store
	currentValue := strstore.SelectWith(store, selector)

	// Create setter that updates the store
	setFieldValue := func(newValue T) {
		store.Update(func(state S) S {
			return setter(state, newValue)
		})
	}

	// Subscribe to store changes and trigger re-render
	// This uses useEffect to manage the subscription lifecycle
	UseEffect(func() CleanupFunc {
		unsubscribe := store.Subscribe(func(state S) {
			currentCtx := rtui.GetCurrentContext()
			if currentCtx != nil {
				scheduleRender(currentCtx.ComponentID)
			}
		})

		return unsubscribe
	}, nil) // Subscribe once on mount

	return currentValue, setFieldValue
}

// UseStoreComputed subscribes to a computed value from a Store.
//
// Computed values are derived from state and automatically update when
// dependencies change. The Store's Computed already handles caching.
//
// This is useful for expensive calculations or derived state that doesn't
// need to be stored separately.
//
// Example:
//
//	type AppState struct {
//	    Items []Item
//	}
//	appStore := store.NewStore(AppState{})
//
//	// Create computed value
//	totalPrice := store.NewComputed(appStore, func(s AppState) float64 {
//	    total := 0.0
//	    for _, item := range s.Items {
//	        total += item.Price
//	    }
//	    return total
//	})
//
//	// In component:
//	total := ui.UseStoreComputed(totalPrice)
//
// Note: This auto-subscribes to the computed value and triggers re-render.
func UseStoreComputed[S any, R any] (computed *strstore.Computed[S, R]) R {
	ctx := rtui.GetCurrentContext()
	if ctx == nil {
		panic("UseStoreComputed must be called within a component")
	}

	// Get current computed value
	currentValue := computed.Get()

	// Subscribe to the underlying store and trigger re-render
	UseEffect(func() CleanupFunc {
		// Note: Computed values automatically subscribe to their store
		// We don't need to do additional subscription here
		// The cleanup will be handled by Dispose if needed

		return func() {
			// Cleanup if needed
		}
	}, nil) // Subscribe once on mount

	return currentValue
}

// UseStoreSelector provides a flexible way to subscribe to derived state.
//
// Unlike UseStoreField which focuses on a single field, UseStoreSelector
// allows you to compute derived values and subscribe to changes.
//
// This is similar to React's useSelector from Redux, adapted for Mint's Store.
//
// Example:
//
//	type AppState struct {
//	    Items []Item
//	}
//	appStore := store.NewStore(AppState{})
//
//	// In component:
//	itemCount := ui.UseStoreSelector(
//	    appStore,
//	    func(s AppState) int {
//	        return len(s.Items)
//	    },
//	)
//
// Note: This auto-subscribes to all Store changes and triggers re-render.
// For performance-critical components, consider using UseStoreField with
// a specific field to avoid unnecessary re-renders.
func UseStoreSelector[S any, T any](
	store *strstore.Store[S],
	selector func(S) T,
) T {
	ctx := rtui.GetCurrentContext()
	if ctx == nil {
		panic("UseStoreSelector must be called within a component")
	}

	// Get current selected value
	currentValue := strstore.SelectWith(store, selector)

	// Track the last selected value to avoid unnecessary re-renders
	hook := ctx.GetOrCreateHook(rtui.HookMemo)

	// Check if value changed
	valueChanged := false
	if !hook.Initialized {
		valueChanged = true
		hook.Initialized = true
	} else {
		// Use deep equality comparison instead of string representation
		// This handles slices, maps, and structs correctly
		valueChanged = !deepEqual(hook.Value, currentValue)
	}

	if valueChanged {
		hook.Value = currentValue
	}

	// Subscribe to store changes and trigger re-render
	UseEffect(func() CleanupFunc {
		unsubscribe := store.Subscribe(func(state S) {
			currentCtx := rtui.GetCurrentContext()
			if currentCtx != nil {
				scheduleRender(currentCtx.ComponentID)
			}
		})

		return unsubscribe
	}, nil) // Subscribe once on mount

	return currentValue
}

// UseStore is a lower-level hook that provides direct access to a Store.
//
// This exposes the underlying Store methods for advanced use cases.
// Most components should prefer UseStoreField or UseStoreSelector.
//
// Example:
//
//	// In component:
//	appStore := ui.UseStore(appStore)
//	state := appStore.Get()
//
// Use with caution: This requires manual subscription management and
// doesn't automatically trigger re-renders.
func UseStore[S any](store *strstore.Store[S]) *strstore.Store[S] {
	ctx := rtui.GetCurrentContext()
	if ctx == nil {
		panic("UseStore must be called within a component")
	}

	return store
}

// =============================================================================
// UseStoreFieldFunctional - Functional Update Support
// =============================================================================

// FunctionalUpdateFunc[T] is the functional update type.
// It takes the current value and returns the new value.
//
// Example:
//
//	count, setCount := ui.UseStoreFieldFunctional(...)
//	setCount(func(old int) int {
//	    return old + 1  // Functional update
//	})
//
// Or directly:
//	setCount(10)  // Direct update
//
// UseStoreFieldFunctional subscribes to a Store field with functional update support.
//
// Unlike UseStoreField, the setter supports both direct value updates and functional updates:
//
//	Direct:   setField(newValue)
//	Functional: setField(func(old T) T { return old + 1 })
//
// This is useful for counters, increment/decrement operations, and cases where
// the new value depends on the current value.
//
// Example:
//
//	type AppState struct {
//	    Count    int
//	    Username string
//	}
//	appStore := store.NewStore(AppState{})
//
//	// In component:
//	count, setCount := ui.UseStoreFieldFunctional(
//	    appStore,
//	    func(s AppState) int { return s.Count },
//	    func(s AppState, v int) AppState { s.Count = v; return s },
//	)
//
//	// Direct update
//	setCount(10)
//
//	// Functional update
//	setCount(func(old int) int {
//	    return old + 1
//	})
func UseStoreFieldFunctional[S any, T any](
	store *strstore.Store[S],
	selector func(S) T,
	setter func(S, T) S,
) (T, func(interface{})) {
	ctx := rtui.GetCurrentContext()
	if ctx == nil {
		panic("UseStoreFieldFunctional must be called within a component")
	}

	// Get current value from store
	currentValue := strstore.SelectWith(store, selector)

	// Create setter that supports both direct and functional updates
	setFieldValue := func(newValue interface{}) {
		switch v := newValue.(type) {
		case T:
			// Direct value update
			store.Update(func(state S) S {
				return setter(state, v)
			})
		case func(T) T:
			// Functional update - apply function to current value
			store.Update(func(state S) S {
				oldValue := selector(state)
				newValue := v(oldValue)
				return setter(state, newValue)
			})
		}
	}

	// Subscribe to store changes and trigger re-render
	UseEffect(func() CleanupFunc {
		unsubscribe := store.Subscribe(func(state S) {
			currentCtx := rtui.GetCurrentContext()
			if currentCtx != nil {
				scheduleRender(currentCtx.ComponentID)
			}
		})

		return unsubscribe
	}, nil) // Subscribe once on mount

	return currentValue, setFieldValue
}

// =============================================================================
// UseStoreFieldForBinding - Simplified API for FieldBinding
// =============================================================================

// UseStoreFieldForBinding provides a simplified API for using FieldBinding with Store.
//
// This hook combines the ForField binding pattern with Store subscriptions, eliminating
// the need to manually manage FieldChangeIntent and Reducer registration.
//
// Usage:
//
//	// Setup FieldBinding helper
//	type BindHelper struct {
//	    Username string
//	}
//	binding := ui.UseStoreFieldForBinding(appStore, BindHelper{})
//
//	// Use in components
//	username, setUsername := binding.String("username")
//	email, setEmail := binding.String("email")
//	count, setCount := binding.Int("count")
//
// This is especially useful for forms where you have many fields.
func UseStoreFieldForBinding[S any](store *strstore.Store[S], _ intent.FieldBinding) StoreFieldBinding[S] {
	ctx := rtui.GetCurrentContext()
	if ctx == nil {
		panic("UseStoreFieldForBinding must be called within a component")
	}

	return StoreFieldBinding[S]{
		store: store,
		ctx:   ctx,
	}
}

// StoreFieldBinding provides a fluent API for binding Store fields.
type StoreFieldBinding[S any] struct {
	store *strstore.Store[S]
	ctx   *rtui.ComponentContext
}

// String creates a string field binding.
func (b StoreFieldBinding[S]) String(fieldName string, selector func(S) string, setter func(S, string) S) (string, func(string)) {
	currentValue := selector(b.store.Get())

	setValue := func(newValue string) {
		b.store.Update(func(state S) S {
			return setter(state, newValue)
		})
	}

	b.subscribe()

	return currentValue, setValue
}

// Int creates an int field binding.
func (b StoreFieldBinding[S]) Int(fieldName string, selector func(S) int, setter func(S, int) S) (int, func(interface{})) {
	currentValue := selector(b.store.Get())

	setValue := func(newValue interface{}) {
		switch v := newValue.(type) {
		case int:
			b.store.Update(func(state S) S {
				return setter(state, v)
			})
		case func(int) int:
			// Functional update
			b.store.Update(func(state S) S {
				oldValue := selector(b.store.Get())
				newVal := v(oldValue)
				return setter(state, newVal)
			})
		}
	}

	b.subscribe()

	return currentValue, setValue
}

// Bool creates a bool field binding.
func (b StoreFieldBinding[S]) Bool(fieldName string, selector func(S) bool, setter func(S, bool) S) (bool, func(bool)) {
	currentValue := selector(b.store.Get())

	setValue := func(newValue bool) {
		b.store.Update(func(state S) S {
			return setter(state, newValue)
		})
	}

	b.subscribe()

	return currentValue, setValue
}

// subscribe subscribes to store changes and triggers re-render.
// This is called automatically for each field binding.
func (b *StoreFieldBinding[S]) subscribe() {
	// Use a ref to track if we've already subscribed
	subscribed := UseRef(false)

	UseEffect(func() CleanupFunc {
		if subscribed.Value.(bool) {
			return nil // Already subscribed
		}
		subscribed.Value = true

		unsubscribe := b.store.Subscribe(func(state S) {
			if ctx := rtui.GetCurrentContext(); ctx != nil {
				scheduleRender(ctx.ComponentID)
			}
		})

		return unsubscribe
	}, nil)
}
