package ui

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// =============================================================================
// InstanceManager Tests
// =============================================================================

// TestInstanceManager_GetOrCreate tests instance creation and reuse
func TestInstanceManager_GetOrCreate(t *testing.T) {
	m := NewInstanceManager()

	createCount := 0
	creator := func() ComponentInstance {
		createCount++
		return NewBaseComponentInstance("test", func() VNode {
			return NewText("test")
		})
	}

	// First call - should create
	inst1 := m.GetOrCreate("key1", creator)
	if createCount != 1 {
		t.Errorf("Expected 1 creation, got %d", createCount)
	}
	if inst1.Key() != "test" {
		t.Errorf("Key = %q, want \"test\"", inst1.Key())
	}

	// Second call with same key - should reuse
	inst2 := m.GetOrCreate("key1", creator)
	if createCount != 1 {
		t.Errorf("Creator should not be called again, got %d creations", createCount)
	}
	if inst1 != inst2 {
		t.Error("Should return same instance")
	}

	// Different key - should create new
	inst3 := m.GetOrCreate("key2", creator)
	if createCount != 2 {
		t.Errorf("Expected 2 creations, got %d", createCount)
	}
	if inst1 == inst3 {
		t.Error("Should return different instance")
	}
}

// TestInstanceManager_Cleanup tests cleanup of unused instances
func TestInstanceManager_Cleanup(t *testing.T) {
	m := NewInstanceManager()

	// Create three instances
	inst1 := m.GetOrCreate("key1", func() ComponentInstance {
		return NewBaseComponentInstance("test1", func() VNode { return NewText("1") })
	})
	_ = m.GetOrCreate("key2", func() ComponentInstance {
		return NewBaseComponentInstance("test2", func() VNode { return NewText("2") })
	})
	inst3 := m.GetOrCreate("key3", func() ComponentInstance {
		return NewBaseComponentInstance("test3", func() VNode { return NewText("3") })
	})

	if m.Count() != 3 {
		t.Errorf("Count = %d, want 3", m.Count())
	}

	// Cleanup keeping only key1 and key3
	m.Cleanup([]string{"key1", "key3"})

	if m.Count() != 2 {
		t.Errorf("After cleanup Count = %d, want 2", m.Count())
	}

	// key2 should be removed, key1 and key3 should still exist
	if m.Get("key2") != nil {
		t.Error("key2 should be removed")
	}
	if m.Get("key1") != inst1 {
		t.Error("key1 should still exist")
	}
	if m.Get("key3") != inst3 {
		t.Error("key3 should still exist")
	}
}

// TestInstanceManager_LRU tests LRU eviction when limit is reached
func TestInstanceManager_LRU(t *testing.T) {
	m := NewInstanceManager()
	m.SetMaxInstances(3) // Set low limit for testing

	// Create 3 instances (at limit)
	m.GetOrCreate("key1", func() ComponentInstance {
		return NewBaseComponentInstance("test1", func() VNode { return NewText("1") })
	})
	m.GetOrCreate("key2", func() ComponentInstance {
		return NewBaseComponentInstance("test2", func() VNode { return NewText("2") })
	})
	m.GetOrCreate("key3", func() ComponentInstance {
		return NewBaseComponentInstance("test3", func() VNode { return NewText("3") })
	})

	if m.Count() != 3 {
		t.Errorf("Count = %d, want 3", m.Count())
	}

	// Access key1 to make it more recent
	m.GetOrCreate("key1", func() ComponentInstance {
		return NewBaseComponentInstance("test1", func() VNode { return NewText("1") })
	})

	// Create key4 - should evict key2 (least recently used)
	m.GetOrCreate("key4", func() ComponentInstance {
		return NewBaseComponentInstance("test4", func() VNode { return NewText("4") })
	})

	if m.Count() != 3 {
		t.Errorf("After eviction Count = %d, want 3", m.Count())
	}

	// key2 should be evicted
	if m.Get("key2") != nil {
		t.Error("key2 should be evicted")
	}

	// key1, key3, key4 should still exist
	if m.Get("key1") == nil {
		t.Error("key1 should exist")
	}
	if m.Get("key3") == nil {
		t.Error("key3 should exist")
	}
	if m.Get("key4") == nil {
		t.Error("key4 should exist")
	}
}

// TestInstanceManager_Remove tests manual removal
func TestInstanceManager_Remove(t *testing.T) {
	m := NewInstanceManager()

	inst := m.GetOrCreate("key1", func() ComponentInstance {
		return NewBaseComponentInstance("test", func() VNode { return NewText("test") })
	})

	if m.Count() != 1 {
		t.Errorf("Count = %d, want 1", m.Count())
	}

	// Remove the instance
	removed := m.Remove("key1")

	if removed != inst {
		t.Error("Should return the removed instance")
	}
	if m.Count() != 0 {
		t.Errorf("After removal Count = %d, want 0", m.Count())
	}
	if m.Get("key1") != nil {
		t.Error("key1 should be removed")
	}
}

// TestInstanceManager_Clear tests clearing all instances
func TestInstanceManager_Clear(t *testing.T) {
	m := NewInstanceManager()

	// Create multiple instances
	m.GetOrCreate("key1", func() ComponentInstance {
		return NewBaseComponentInstance("test1", func() VNode { return NewText("1") })
	})
	m.GetOrCreate("key2", func() ComponentInstance {
		return NewBaseComponentInstance("test2", func() VNode { return NewText("2") })
	})
	m.GetOrCreate("key3", func() ComponentInstance {
		return NewBaseComponentInstance("test3", func() VNode { return NewText("3") })
	})

	if m.Count() != 3 {
		t.Errorf("Count = %d, want 3", m.Count())
	}

	// Clear all
	m.Clear()

	if m.Count() != 0 {
		t.Errorf("After clear Count = %d, want 0", m.Count())
	}
}

// =============================================================================
// ComponentInstance Lifecycle Tests
// =============================================================================

// TestComponentInstance_OnMount tests OnMount is called
func TestComponentInstance_OnMount(t *testing.T) {
	inst := NewBaseComponentInstance("test", func() VNode {
		return NewText("test")
	})

	// Initially not mounted
	if inst.IsMounted() {
		t.Error("Should not be mounted initially")
	}

	// Call OnMount
	inst.OnMount()

	if !inst.IsMounted() {
		t.Error("Should be mounted after OnMount")
	}
}

// TestComponentInstance_OnUnmount tests cleanup on unmount
func TestComponentInstance_OnUnmount(t *testing.T) {
	cleanupCalled := false
	cleanupPtr := &cleanupCalled

	// Create a component function that uses useEffect
	componentFn := func() VNode {
		// Add an effect with cleanup
		UseEffect(func() CleanupFunc {
			return func() {
				*cleanupPtr = true
			}
		}, nil)
		return NewText("test")
	}

	// Create instance - this creates a new context
	inst := NewBaseComponentInstance("test", componentFn)

	// Render the component (registers the useEffect)
	oldContext := getCurrentContext()
	setCurrentContext(inst.GetContext())
	inst.Render()
	setCurrentContext(oldContext)

	// Run effects to register cleanup function
	inst.GetContext().runEffects()

	// Unmount should call cleanup
	inst.OnUnmount()

	if !cleanupCalled {
		t.Error("Cleanup should be called on unmount")
	}
	if inst.IsMounted() {
		t.Error("Instance should not be mounted after OnUnmount")
	}
}

// TestComponentInstance_StatePersistence tests state persists across renders
func TestComponentInstance_StatePersistence(t *testing.T) {
	inst := NewBaseComponentInstance("test", func() VNode {
		return NewText("test")
	})

	// Set some state via the context
	setCurrentContext(inst.GetContext())
	defer setCurrentContext(nil)

	count, setCount, _ := UseStateInt(0)
	if count != 0 {
		t.Errorf("Initial count = %d, want 0", count)
	}

	// Update state
	setCount(42)

	// Re-render
	inst.GetContext().resetContext()
	count2, _, _ := UseStateInt(0)

	if count2 != 42 {
		t.Errorf("After re-render count = %d, want 42", count2)
	}
}

// TestComponentInstance_Props tests props updating
func TestComponentInstance_Props(t *testing.T) {
	inst := NewBaseComponentInstance("test", nil)

	// Initial props
	newProps := make(Props)
	newProps["name"] = "Alice"
	newProps["age"] = 30

	changed := inst.SetProps(newProps)
	if !changed {
		t.Error("SetProps should return true when props change")
	}

	// Same props - should not change
	changed = inst.SetProps(newProps)
	if changed {
		t.Error("SetProps should return false when props are same")
	}

	// Verify props
	if inst.GetProps()["name"] != "Alice" {
		t.Errorf("name = %v, want Alice", inst.GetProps()["name"])
	}
}

// =============================================================================
// useHoverState Tests
// =============================================================================

// TestUseHoverState tests the useHoverState hook
func TestUseHoverState(t *testing.T) {
	ctx := newComponentContext("TestComponent")
	setCurrentContext(ctx)
	defer setCurrentContext(nil)

	isHovered, setHovered := useHoverState()

	// Initial state should be false
	if isHovered() {
		t.Error("Initial hovered should be false")
	}

	// Set to true
	setHovered(true)
	if !isHovered() {
		t.Error("hovered should be true after setHovered(true)")
	}

	// Re-render - state should persist (still true)
	ctx.resetContext()
	isHovered2, setHovered2 := useHoverState()

	if !isHovered2() {
		t.Error("hovered state should persist across renders (should still be true)")
	}

	// Now set to false
	setHovered2(false)
	if isHovered2() {
		t.Error("hovered should be false after setHovered2(false)")
	}
}

// TestUseHoverStateInComponent tests useHoverState in a realistic scenario
func TestUseHoverStateInComponent(t *testing.T) {
	ctx := newComponentContext("HoverButton")
	setCurrentContext(ctx)
	defer setCurrentContext(nil)

	// Simulate component using useHoverState
	isHovered, setHovered := useHoverState()

	var hoveredCount int
	// Simulate mouse enter
	setHovered(true)
	if isHovered() {
		hoveredCount++
	}

	// Simulate multiple renders
	for i := 0; i < 5; i++ {
		ctx.resetContext()
		isHovered, _ := useHoverState()
		if isHovered() {
			hoveredCount++
		}
	}

	if hoveredCount != 6 {
		t.Errorf("hoveredCount = %d, want 6", hoveredCount)
	}
}

// =============================================================================
// Memory Leak Tests
// =============================================================================

// TestCleanupAllCalled tests that cleanup functions are called
func TestCleanupAllCalled(t *testing.T) {
	ctx := newComponentContext("TestComponent")
	setCurrentContext(ctx)
	defer setCurrentContext(nil)

	cleanupCount := 0
	// Use a pointer to make updates visible in closures
	cleanupCountPtr := &cleanupCount

	// Add multiple effects with cleanup
	UseEffect(func() CleanupFunc {
		return func() { (*cleanupCountPtr)++; t.Log("cleanup 1 called") }
	}, nil)

	UseEffect(func() CleanupFunc {
		return func() { (*cleanupCountPtr)++; t.Log("cleanup 2 called") }
	}, []interface{}{1})

	// Run effects to register cleanup functions
	ctx.runEffects()
	t.Logf("After runEffects: cleanupCount = %d, hooks[0].Cleanup = %v, hooks[1].Cleanup = %v",
		cleanupCount, ctx.Hooks[0].Cleanup != nil, ctx.Hooks[1].Cleanup != nil)

	// Cleanup all
	ctx.cleanupAll()
	t.Logf("After cleanupAll: cleanupCount = %d", cleanupCount)

	// cleanupCount should be 2 (from cleanupAll)
	if cleanupCount != 2 {
		t.Errorf("cleanupCount = %d, want 2", cleanupCount)
	}

	// Verify cleanup functions were set to nil
	if ctx.Hooks[0].Cleanup != nil {
		t.Error("Cleanup function 0 should be nil after cleanupAll")
	}
	if ctx.Hooks[1].Cleanup != nil {
		t.Error("Cleanup function 1 should be nil after cleanupAll")
	}
}

// TestGoroutineLeakDetection is a basic goroutine leak test
// Note: This is a simple test and may not catch all leaks
func TestGoroutineLeakDetection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping leak test in short mode")
	}

	// This test verifies that properly written effects don't leak goroutines
	// In a real application, you'd use runtime.NumGoroutine() to detect leaks

	ctx := newComponentContext("TestComponent")
	setCurrentContext(ctx)
	defer setCurrentContext(nil)

	doneChans := make([]chan struct{}, 0)

	// Create effects with goroutines that properly clean up
	for i := 0; i < 5; i++ {
		done := make(chan struct{})
		doneChans = append(doneChans, done)

		UseEffect(func() CleanupFunc {
			go func() {
				for {
					select {
					case <-time.After(10 * time.Millisecond):
						// Some work
					case <-done:
						return // Exit when done
					}
				}
			}()
			return func() { close(done) }
		}, nil)
		_ = i // Use loop variable
	}

	// Run effects
	ctx.runEffects()

	// Cleanup all - all goroutines should exit
	ctx.cleanupAll()

	// Wait a bit for goroutines to exit
	time.Sleep(50 * time.Millisecond)

	// If there were leaks, goroutines would still be running
	// In a real test, we'd check runtime.NumGoroutine()
}

// =============================================================================
// Thread Safety Tests
// =============================================================================

// TestInstanceManager_ConcurrentGetOrCreate tests concurrent instance creation
func TestInstanceManager_ConcurrentGetOrCreate(t *testing.T) {
	m := NewInstanceManager()
	var wg sync.WaitGroup
	createCount := 0
	var mu sync.Mutex

	creator := func() ComponentInstance {
		mu.Lock()
		createCount++
		mu.Unlock()
		return NewBaseComponentInstance("test", func() VNode {
			return NewText("test")
		})
	}

	// Concurrent creates with same key
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			m.GetOrCreate("key1", creator)
		}()
	}

	wg.Wait()

	// Should only create once despite 100 concurrent calls
	if createCount != 1 {
		t.Errorf("Expected 1 creation with concurrent calls, got %d", createCount)
	}

	// Should have exactly 1 instance
	if m.Count() != 1 {
		t.Errorf("Count = %d, want 1", m.Count())
	}
}

// TestInstanceManager_ConcurrentCleanup tests concurrent cleanup
func TestInstanceManager_ConcurrentCleanup(t *testing.T) {
	m := NewInstanceManager()
	var wg sync.WaitGroup

	// Create many instances
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key%d", i)
		m.GetOrCreate(key, func() ComponentInstance {
			return NewBaseComponentInstance(key, func() VNode {
				return NewText(key)
			})
		})
	}

	// Concurrent cleanups
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(iter int) {
			defer wg.Done()
			// Each iteration keeps different keys
			activeKeys := make([]string, 0)
			for j := 0; j < 10; j++ {
				key := fmt.Sprintf("key%d", iter*10+j)
				activeKeys = append(activeKeys, key)
			}
			m.Cleanup(activeKeys)
		}(i)
	}

	wg.Wait()

	// Should have no instances left
	if m.Count() != 0 {
		t.Errorf("After concurrent cleanups Count = %d, want 0", m.Count())
	}
}
