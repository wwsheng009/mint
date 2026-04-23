package context

import (
	"sync"
	"testing"
)

func TestContext_ProvideAndUse(t *testing.T) {
	TestKey := ContextKey("test-key")
	TestValue := 123

	ctx := NewContext(nil)

	// Test Provide
	ctx.Provide(TestKey, TestValue)

	// Test UseContext
	value := ctx.UseContext(TestKey)
	if value != TestValue {
		t.Errorf("Expected %v, got %v", TestValue, value)
	}
}

func TestContext_Inheritance(t *testing.T) {
	ParentKey := ContextKey("parent-key")
	ChildKey := ContextKey("child-key")

	// Create parent context
	parent := NewContext(nil)
	parent.Provide(ParentKey, "parent-value")

	// Create child context (inherits from parent)
	child := NewContext(parent)
	child.Provide(ChildKey, "child-value")

	// Child should have parent's value
	if value := child.UseContext(ParentKey); value != "parent-value" {
		t.Errorf("Child should have parent's value: got %v", value)
	}

	// Child should have its own value
	if value := child.UseContext(ChildKey); value != "child-value" {
		t.Errorf("Child should have its own value: got %v", value)
	}

	// Parent should NOT have child's value
	if value := parent.UseContext(ChildKey); value != nil {
		t.Errorf("Parent should NOT have child's value: got %v", value)
	}
}

func TestContext_HasContext(t *testing.T) {
	TestKey := ContextKey("has-test")

	ctx := NewContext(nil)

	if ctx.HasContext(TestKey) {
		t.Error("empty context should not have key")
	}

	ctx.Provide(TestKey, "value")

	if !ctx.HasContext(TestKey) {
		t.Error("context should have key after provide")
	}
}

func TestContext_HasContextInHierarchy(t *testing.T) {
	ParentKey := ContextKey("parent-hierarchy")
	ChildKey := ContextKey("child-hierarchy")

	parent := NewContext(nil)
	parent.Provide(ParentKey, "parent-value")

	child := NewContext(parent)
	child.Provide(ChildKey, "child-value")

	// Child should find parent's value through hierarchy
	if !child.HasContextInHierarchy(ParentKey) {
		t.Error("child should find parent's key in hierarchy")
	}

	// Child should find its own key
	if !child.HasContextInHierarchy(ChildKey) {
		t.Error("child should find its own key")
	}

	// Parent should NOT find child's key
	if parent.HasContextInHierarchy(ChildKey) {
		t.Error("parent should not find child's key")
	}

	// Test key that doesn't exist
	NonExistKey := ContextKey("non-exist")
	if child.HasContextInHierarchy(NonExistKey) {
		t.Error("should not find non-existent key")
	}
}

func TestContext_ThreadSafety(t *testing.T) {
	TestKey := ContextKey("thread-test")
	ctx := NewContext(nil)

	var wg sync.WaitGroup
	concurrency := 100

	// Concurrent writes
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(val int) {
			defer wg.Done()
			ctx.Provide(TestKey, val)
		}(i)
	}

	wg.Wait()

	// Verify can read
	if ctx.UseContext(TestKey) == nil {
		t.Error("should be able to read after concurrent writes")
	}
}

func TestContext_UseContextNil(t *testing.T) {
	var ctx *FiberContext

	// Should not panic
	value := ctx.UseContext(ContextKey("test"))
	if value != nil {
		t.Error("nil context should return nil for any key")
	}
}

func TestContext_ProvideNil(t *testing.T) {
	var ctx *FiberContext

	// Should not panic
	ctx.Provide(ContextKey("test"), "value")
}

func TestNewContextWithNilParent(t *testing.T) {
	// Create context with nil parent
	ctx := NewContext(nil)
	if ctx == nil {
		t.Fatal("NewContext with nil parent should not return nil")
	}

	if ctx.parent != nil {
		t.Error("parent should be nil when nil is passed")
	}

	if ctx.values == nil {
		t.Error("values map should be initialized")
	}
}

func TestContext_MultipleValues(t *testing.T) {
	ctx := NewContext(nil)

	// Provide multiple values
	keys := []ContextKey{"key1", "key2", "key3"}
	values := []interface{}{"value1", 42, true}

	for i, key := range keys {
		ctx.Provide(key, values[i])
	}

	// Verify all values
	for i, key := range keys {
		value := ctx.UseContext(key)
		if value != values[i] {
			t.Errorf("Key %s: expected %v, got %v", key, values[i], value)
		}
	}
}
