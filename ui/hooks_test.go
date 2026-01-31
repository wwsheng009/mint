package ui

import (
	"testing"
)

// TestUseStateInt tests UseStateInt hook
func TestUseStateInt(t *testing.T) {
	ctx := newComponentContext("TestComponent")
	setCurrentContext(ctx)
	defer setCurrentContext(nil)

	// First call should initialize
	count, setCount, getCount := UseStateInt(0)
	if count != 0 {
		t.Errorf("Initial count = %v, want 0", count)
	}

	// Set new value
	setCount(5)

	// getValue should return new value
	if getCount() != 5 {
		t.Errorf("getCount() = %v, want 5", getCount())
	}

	// Reset context and render again - should preserve value
	ctx.resetContext()
	count2, _, getCount2 := UseStateInt(0)
	if count2 != 5 {
		t.Errorf("After re-render count = %v, want 5", count2)
	}
	if getCount2() != 5 {
		t.Errorf("After re-render getCount() = %v, want 5", getCount2())
	}
}

// TestUseStateIntFunctional tests functional updates
func TestUseStateIntFunctional(t *testing.T) {
	ctx := newComponentContext("TestComponent")
	setCurrentContext(ctx)
	defer setCurrentContext(nil)

	count, setCount, _ := UseStateInt(10)

	if count != 10 {
		t.Errorf("Initial count = %v, want 10", count)
	}

	// Functional update
	setCount(func(c int) int { return c + 5 })

	// After functional update, value should be 15
	ctx.resetContext()
	count2, _, _ := UseStateInt(0)
	if count2 != 15 {
		t.Errorf("After functional update count = %v, want 15", count2)
	}

	// Another functional update
	setCount(func(c int) int { return c * 2 })
	ctx.resetContext()
	count3, _, _ := UseStateInt(0)
	if count3 != 30 {
		t.Errorf("After second functional update count = %v, want 30", count3)
	}
}

// TestUseStateString tests UseStateString hook
func TestUseStateString(t *testing.T) {
	ctx := newComponentContext("TestComponent")
	setCurrentContext(ctx)
	defer setCurrentContext(nil)

	// Note: The returned value is a snapshot, not updated by setter
	_, setName := UseStateString("")

	// Setter updates internal state, not the returned variable
	setName("Alice")
	// The 'name' variable still holds the initial value
	// To get the updated value, we need to re-render
	ctx.resetContext()
	name2, _ := UseStateString("")

	if name2 != "Alice" {
		t.Errorf("After re-render name = %v, want Alice", name2)
	}
}

// TestUseStateBool tests UseStateBool hook
func TestUseStateBool(t *testing.T) {
	ctx := newComponentContext("TestComponent")
	setCurrentContext(ctx)
	defer setCurrentContext(nil)

	_, setVisible := UseStateBool(false)

	// Setter updates internal state
	setVisible(true)
	// The 'visible' variable still holds the initial value
	// To get the updated value, we need to re-render
	ctx.resetContext()
	visible2, _ := UseStateBool(false)

	if visible2 != true {
		t.Errorf("After re-render visible = %v, want true", visible2)
	}
}

// TestMultipleHooks tests multiple hook calls
func TestMultipleHooks(t *testing.T) {
	ctx := newComponentContext("TestComponent")
	setCurrentContext(ctx)
	defer setCurrentContext(nil)

	// Call multiple hooks
	count, _, _ := UseStateInt(0)
	name, _ := UseStateString("")
	flag, _ := UseStateBool(false)

	if count != 0 || name != "" || flag != false {
		t.Error("Initial values should be defaults")
	}

	// Re-render with same order
	ctx.resetContext()
	count2, _, _ := UseStateInt(0)
	name2, _ := UseStateString("")
	flag2, _ := UseStateBool(false)

	if count2 != 0 || name2 != "" || flag2 != false {
		t.Error("Re-render values should be defaults")
	}
}

// TestHookOutsideComponentPanic tests that hooks panic outside component
func TestHookOutsideComponentPanic(t *testing.T) {
	setCurrentContext(nil)
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when calling useState outside component")
		}
	}()
	UseStateInt(0)
}

// TestNewComponentContext tests component context creation
func TestNewComponentContext(t *testing.T) {
	ctx := newComponentContext("TestComponent")

	if ctx.ComponentID == "" {
		t.Error("ComponentID should not be empty")
	}
	if ctx.HookIndex != 0 {
		t.Errorf("HookIndex = %v, want 0", ctx.HookIndex)
	}
	if ctx.RenderCount != 0 {
		t.Errorf("RenderCount = %v, want 0", ctx.RenderCount)
	}
	if len(ctx.Hooks) != 0 {
		t.Errorf("len(Hooks) = %v, want 0", len(ctx.Hooks))
	}
}

// TestResetContext tests context reset
func TestResetContext(t *testing.T) {
	ctx := newComponentContext("TestComponent")
	setCurrentContext(ctx)
	defer setCurrentContext(nil)

	// Add a hook
	UseStateInt(42)
	if ctx.HookIndex != 1 {
		t.Errorf("HookIndex = %v, want 1", ctx.HookIndex)
	}

	// Reset
	ctx.resetContext()
	if ctx.HookIndex != 0 {
		t.Errorf("After reset HookIndex = %v, want 0", ctx.HookIndex)
	}
	if ctx.RenderCount != 1 {
		t.Errorf("After reset RenderCount = %v, want 1", ctx.RenderCount)
	}
}

// TestFinishRender tests hook validation on render finish
func TestFinishRender(t *testing.T) {
	ctx := newComponentContext("TestComponent")

	// Reset and add hooks
	ctx.resetContext()
	setCurrentContext(ctx)
	defer setCurrentContext(nil)

	UseStateInt(0)
	UseStateString("")

	// Should pass with consistent hooks
	if err := ctx.finishRender(); err != nil {
		t.Errorf("finishRender() = %v, want nil", err)
	}

	// Second render with same hook count should pass
	ctx.resetContext()
	UseStateInt(0)
	UseStateString("")

	if err := ctx.finishRender(); err != nil {
		t.Errorf("finishRender() = %v, want nil", err)
	}
}

// TestHookOrderValidation tests hook order validation
func TestHookOrderValidation(t *testing.T) {
	ctx := newComponentContext("TestComponent")
	setCurrentContext(ctx)
	defer setCurrentContext(nil)

	// First render - 2 hooks
	UseStateInt(0)
	UseStateString("")

	ctx.finishRender()

	// Second render - should have same order
	ctx.resetContext()
	UseStateInt(0)
	UseStateString("")

	if err := ctx.finishRender(); err != nil {
		t.Errorf("finishRender() = %v, want nil", err)
	}
}

// TestHookOrderViolation tests that hook order violations are detected
func TestHookOrderViolation(t *testing.T) {
	ctx := newComponentContext("TestComponent")
	setCurrentContext(ctx)
	defer setCurrentContext(nil)

	// First render - 2 hooks
	UseStateInt(0)
	UseStateString("")

	ctx.finishRender()

	// Second render - only 1 hook (order violation)
	ctx.resetContext()
	UseStateInt(0)
	// Missing UseStateString - this should cause error on finishRender

	if err := ctx.finishRender(); err == nil {
		t.Error("finishRender() should error when hook count changes")
	}
}

// BenchmarkUseStateInt benchmarks UseStateInt
func BenchmarkUseStateInt(b *testing.B) {
	ctx := newComponentContext("Benchmark")
	setCurrentContext(ctx)
	ctx.resetContext() // Initialize hooks slice

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if i == 0 {
			ctx.resetContext()
		}
		UseStateInt(0)
	}
}

// BenchmarkUseStateIntMultiple benchmarks multiple UseStateInt calls
func BenchmarkUseStateIntMultiple(b *testing.B) {
	ctx := newComponentContext("Benchmark")
	setCurrentContext(ctx)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx.resetContext()
		UseStateInt(0)
		UseStateInt(0)
		UseStateInt(0)
		UseStateInt(0)
		UseStateInt(0)
	}
}

// =============================================================================
// useEffect Tests
// =============================================================================

// TestUseEffectOnce tests useEffect with nil deps (runs once on mount)
func TestUseEffectOnce(t *testing.T) {
	ctx := newComponentContext("TestComponent")
	setCurrentContext(ctx)
	defer setCurrentContext(nil)

	runCount := 0

	// First render - effect should run
	UseEffect(func() CleanupFunc {
		runCount++
		return nil
	}, nil)

	if runCount != 0 {
		t.Errorf("Effect should not run during render, runCount = %d", runCount)
	}

	// Run effects explicitly
	ctx.runEffects()

	if runCount != 1 {
		t.Errorf("Effect should run once, runCount = %d", runCount)
	}

	// Second render - effect should NOT run again
	ctx.resetContext()
	UseEffect(func() CleanupFunc {
		runCount++
		return nil
	}, nil)
	ctx.runEffects()

	if runCount != 1 {
		t.Errorf("Effect should not run again, runCount = %d", runCount)
	}
}

// TestUseEffectWithDeps tests useEffect with dependency tracking
func TestUseEffectWithDeps(t *testing.T) {
	ctx := newComponentContext("TestComponent")
	setCurrentContext(ctx)
	defer setCurrentContext(nil)

	runCount := 0
	deps := []interface{}{1}

	// First render - should run
	UseEffect(func() CleanupFunc {
		runCount++
		return nil
	}, deps)
	ctx.runEffects()

	if runCount != 1 {
		t.Errorf("Effect should run on first render, runCount = %d", runCount)
	}

	// Second render with same deps - should NOT run
	ctx.resetContext()
	UseEffect(func() CleanupFunc {
		runCount++
		return nil
	}, deps)
	ctx.runEffects()

	if runCount != 1 {
		t.Errorf("Effect should not run with same deps, runCount = %d", runCount)
	}

	// Third render with different deps - should run
	ctx.resetContext()
	newDeps := []interface{}{2}
	UseEffect(func() CleanupFunc {
		runCount++
		return nil
	}, newDeps)
	ctx.runEffects()

	if runCount != 2 {
		t.Errorf("Effect should run with different deps, runCount = %d", runCount)
	}
}

// TestUseEffectCleanup tests useEffect cleanup function
func TestUseEffectCleanup(t *testing.T) {
	ctx := newComponentContext("TestComponent")
	setCurrentContext(ctx)
	defer setCurrentContext(nil)

	cleanupCount := 0

	// First render - effect with cleanup
	UseEffect(func() CleanupFunc {
		return func() {
			cleanupCount++
		}
	}, nil)
	ctx.runEffects()

	if cleanupCount != 0 {
		t.Errorf("Cleanup should not run during render, cleanupCount = %d", cleanupCount)
	}

	// Second render with nil deps - cleanup should NOT run (run-once behavior)
	ctx.resetContext()
	UseEffect(func() CleanupFunc {
		return func() {
			cleanupCount++
		}
	}, nil)
	ctx.runEffects()

	if cleanupCount != 0 {
		t.Errorf("Cleanup should not run with nil deps, cleanupCount = %d", cleanupCount)
	}

	// Third render with different deps - cleanup SHOULD run
	ctx.resetContext()
	UseEffect(func() CleanupFunc {
		return func() {
			cleanupCount++
		}
	}, []interface{}{1})
	ctx.runEffects()

	if cleanupCount != 1 {
		t.Errorf("Cleanup should run when deps change, cleanupCount = %d", cleanupCount)
	}
}

// TestUseEffectEmptyDeps tests useEffect with empty deps array
func TestUseEffectEmptyDeps(t *testing.T) {
	ctx := newComponentContext("TestComponent")
	setCurrentContext(ctx)
	defer setCurrentContext(nil)

	runCount := 0

	// Empty deps means run on every render
	UseEffect(func() CleanupFunc {
		runCount++
		return nil
	}, []interface{}{})
	ctx.runEffects()

	if runCount != 1 {
		t.Errorf("Effect should run, runCount = %d", runCount)
	}

	// Second render - should run again
	ctx.resetContext()
	UseEffect(func() CleanupFunc {
		runCount++
		return nil
	}, []interface{}{})
	ctx.runEffects()

	if runCount != 2 {
		t.Errorf("Effect should run again with empty deps, runCount = %d", runCount)
	}
}

// =============================================================================
// useRef Tests
// =============================================================================

// TestUseRef tests useRef hook
func TestUseRef(t *testing.T) {
	ctx := newComponentContext("TestComponent")
	setCurrentContext(ctx)
	defer setCurrentContext(nil)

	ref := UseRef(42)

	if ref.Value != 42 {
		t.Errorf("Initial ref.Value = %v, want 42", ref.Value)
	}

	// Mutate ref directly
	ref.Value = 100

	if ref.Value != 100 {
		t.Errorf("After mutation ref.Value = %v, want 100", ref.Value)
	}

	// Re-render - ref should persist
	ctx.resetContext()
	ref2 := UseRef(0) // Initial value ignored on re-render

	if ref2 != ref {
		t.Error("Ref should be same instance on re-render")
	}
	if ref.Value != 100 {
		t.Errorf("Ref value should persist, got %v, want 100", ref.Value)
	}
}

// TestUseRefOutsideComponentPanic tests that useRef panics outside component
func TestUseRefOutsideComponentPanic(t *testing.T) {
	setCurrentContext(nil)
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when calling UseRef outside component")
		}
	}()
	UseRef(0)
}

// =============================================================================
// useMemo Tests
// =============================================================================

// TestUseMemo tests useMemo hook
func TestUseMemo(t *testing.T) {
	ctx := newComponentContext("TestComponent")
	setCurrentContext(ctx)
	defer setCurrentContext(nil)

	computeCount := 0
	compute := func() interface{} {
		computeCount++
		return 42
	}

	// First render - should compute
	result := UseMemo(compute, []interface{}{1})

	if result != 42 {
		t.Errorf("result = %v, want 42", result)
	}
	if computeCount != 1 {
		t.Errorf("computeCount = %d, want 1", computeCount)
	}

	// Re-render with same deps - should NOT recompute
	ctx.resetContext()
	result2 := UseMemo(compute, []interface{}{1})

	if result2 != 42 {
		t.Errorf("result2 = %v, want 42", result2)
	}
	if computeCount != 1 {
		t.Errorf("Should not recompute, computeCount = %d, want 1", computeCount)
	}

	// Re-render with different deps - should recompute
	ctx.resetContext()
	result3 := UseMemo(compute, []interface{}{2})

	if result3 != 42 {
		t.Errorf("result3 = %v, want 42", result3)
	}
	if computeCount != 2 {
		t.Errorf("Should recompute, computeCount = %d, want 2", computeCount)
	}
}

// TestUseMemoNilDeps tests useMemo with nil deps (computes once)
func TestUseMemoNilDeps(t *testing.T) {
	ctx := newComponentContext("TestComponent")
	setCurrentContext(ctx)
	defer setCurrentContext(nil)

	computeCount := 0
	compute := func() interface{} {
		computeCount++
		return 42
	}

	// First render - should compute
	result := UseMemo(compute, nil)

	if computeCount != 1 {
		t.Errorf("computeCount = %d, want 1", computeCount)
	}

	// Re-render - should NOT recompute with nil deps
	ctx.resetContext()
	UseMemo(compute, nil)

	if computeCount != 1 {
		t.Errorf("Should not recompute with nil deps, computeCount = %d, want 1", computeCount)
	}

	_ = result
}

// =============================================================================
// useCallback Tests
// =============================================================================

// TestUseCallback tests useCallback hook
func TestUseCallback(t *testing.T) {
	ctx := newComponentContext("TestComponent")
	setCurrentContext(ctx)
	defer setCurrentContext(nil)

	computeCount := 0
	callback := func() {
		computeCount++
	}

	// First render
	cb1 := UseCallback(callback, []interface{}{1})
	if cb1 == nil {
		t.Error("UseCallback should not return nil")
	}
	cb1()

	if computeCount != 1 {
		t.Errorf("computeCount = %d, want 1", computeCount)
	}

	// Re-render with same deps - memoized value should be preserved
	// We can't compare functions directly, but we can test behavior
	ctx.resetContext()
	cb2 := UseCallback(callback, []interface{}{1})

	if cb2 == nil {
		t.Error("UseCallback should not return nil on re-render")
	}

	// Both callbacks should be functional
	cb1()
	cb2()

	if computeCount != 3 {
		t.Errorf("computeCount = %d, want 3", computeCount)
	}

	// Re-render with different deps - should create new callback
	// Again, we verify behavior rather than identity
	ctx.resetContext()
	cb3 := UseCallback(callback, []interface{}{2})

	if cb3 == nil {
		t.Error("UseCallback should not return nil with different deps")
	}
}

// =============================================================================
// Helper function tests
// =============================================================================

// TestDepsEqual tests depsEqual helper
func TestDepsEqual(t *testing.T) {
	tests := []struct {
		name string
		a    []interface{}
		b    []interface{}
		want bool
	}{
		{"Both nil", nil, nil, true},
		{"Both empty", []interface{}{}, []interface{}{}, true},
		{"Same length and values", []interface{}{1, "a"}, []interface{}{1, "a"}, true},
		{"Different length", []interface{}{1}, []interface{}{1, 2}, false},
		{"Different values", []interface{}{1}, []interface{}{2}, false},
		{"One nil one not", nil, []interface{}{1}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := depsEqual(tt.a, tt.b); got != tt.want {
				t.Errorf("depsEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}

// =============================================================================
// Benchmarks for new hooks
// =============================================================================

// BenchmarkUseEffect benchmarks useEffect
func BenchmarkUseEffect(b *testing.B) {
	ctx := newComponentContext("Benchmark")
	setCurrentContext(ctx)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx.resetContext()
		UseEffect(func() CleanupFunc {
			return func() {}
		}, nil)
	}
}

// BenchmarkUseRef benchmarks UseRef
func BenchmarkUseRef(b *testing.B) {
	ctx := newComponentContext("Benchmark")
	setCurrentContext(ctx)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx.resetContext()
		UseRef(0)
	}
}

// BenchmarkUseMemo benchmarks UseMemo
func BenchmarkUseMemo(b *testing.B) {
	ctx := newComponentContext("Benchmark")
	setCurrentContext(ctx)

	compute := func() interface{} {
		return 42
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx.resetContext()
		UseMemo(compute, []interface{}{1})
	}
}
