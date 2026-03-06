package context

import (
	"testing"
)

func TestUseContextValue_TypeSafety(t *testing.T) {
	StringKey := ContextKey("string")
	IntKey := ContextKey("int")

	ctx := NewContext(nil)
	ctx.Provide(StringKey, "hello")
	ctx.Provide(IntKey, 42)

	// Type-safe access
	str, ok := UseContextValue[string](ctx, StringKey)
	if !ok {
		t.Error("should retrieve string value")
	}
	if str != "hello" {
		t.Errorf("string value mismatch: got %s", str)
	}

	num, ok := UseContextValue[int](ctx, IntKey)
	if !ok {
		t.Error("should retrieve int value")
	}
	if num != 42 {
		t.Errorf("int value mismatch: got %d", num)
	}

	// Type mismatch - should fail
	_, ok = UseContextValue[string](ctx, IntKey)
	if ok {
		t.Error("type mismatch should return false")
	}

	// Non-existent key - should fail
	_, ok = UseContextValue[int](ctx, "non-existent")
	if ok {
		t.Error("non-existent key should return false")
	}
}

func TestUseContextValue_PointerTypes(t *testing.T) {
	PtrKey := ContextKey("pointer")

	value := "hello"
	ctx := NewContext(nil)
	ctx.Provide(PtrKey, &value)

	// Retrieve pointer
	ptr, ok := UseContextValue[*string](ctx, PtrKey)
	if !ok {
		t.Error("should retrieve pointer value")
	}
	if *ptr != "hello" {
		t.Errorf("pointer value mismatch: got %s", *ptr)
	}

	// Type mismatch (non-pointer) - should fail
	_, ok = UseContextValue[string](ctx, PtrKey)
	if ok {
		t.Error("non-pointer type should fail for pointer value")
	}
}

func TestUseContextValue_CustomType(t *testing.T) {
	type TestStruct struct {
		Name string
		Age  int
	}

	TestKey := ContextKey("struct")

	ctx := NewContext(nil)
	value := TestStruct{
		Name: "test",
		Age:  30,
	}
	ctx.Provide(TestKey, value)

	// Retrieve custom type
	result, ok := UseContextValue[TestStruct](ctx, TestKey)
	if !ok {
		t.Error("should retrieve custom type")
	}
	if result.Name != "test" || result.Age != 30 {
		t.Errorf("custom type value mismatch: got %+v", result)
	}
}

func TestUseContextValue_ZeroValue(t *testing.T) {
	ctx := NewContext(nil)

	// Non-existent key should return zero value
	result, ok := UseContextValue[int](ctx, "non-existent")
	if ok {
		t.Error("non-existent key should return false")
	}
	if result != 0 {
		t.Errorf("zero value mismatch: got %d", result)
	}

	// String zero value
	str, ok := UseContextValue[string](ctx, "non-existent")
	if ok {
		t.Error("non-existent key should return false for string")
	}
	if str != "" {
		t.Errorf("string zero value mismatch: got %q", str)
	}
}

func TestUseContextValue_Inheritance(t *testing.T) {
	ParentKey := ContextKey("parent")
	ChildKey := ContextKey("child")

	parent := NewContext(nil)
	parent.Provide(ParentKey, "parent-value")

	child := NewContext(parent)
	child.Provide(ChildKey, "child-value")

	// Child should retrieve parent's value
	str, ok := UseContextValue[string](child, ParentKey)
	if !ok {
		t.Error("should retrieve parent's value through inheritance")
	}
	if str != "parent-value" {
		t.Errorf("parent value mismatch: got %s", str)
	}
}

func TestUseContextValue_NilContext(t *testing.T) {
	var ctx *FiberContext

	// Should not panic and return zero value with false
	result, ok := UseContextValue[string](ctx, "test")
	if ok {
		t.Error("nil context should return false")
	}
	if result != "" {
		t.Errorf("zero value for nil context mismatch: got %q", result)
	}
}

func TestUseContextValue_Override(t *testing.T) {
	TestKey := ContextKey("override")

	parent := NewContext(nil)
	parent.Provide(TestKey, "parent")

	child := NewContext(parent)
	child.Provide(TestKey, "child")

	// Child should get its own value, not parent's
	value, ok := UseContextValue[string](child, TestKey)
	if !ok {
		t.Error("should retrieve child's own value")
	}
	if value != "child" {
		t.Errorf("child's own value should override: got %s", value)
	}

	// Parent should still have its own value
	parentValue, ok := UseContextValue[string](parent, TestKey)
	if !ok {
		t.Error("parent should still have its value")
	}
	if parentValue != "parent" {
		t.Errorf("parent value should be unchanged: got %s", parentValue)
	}
}
