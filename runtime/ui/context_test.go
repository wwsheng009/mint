package ui

import (
	"testing"

	fcontext "github.com/wwsheng009/mint/runtime/context"
)

// TestFiberContext_Propagation tests that Context is properly propagated through the Fiber tree
func TestFiberContext_Propagation(t *testing.T) {
	TestKey := fcontext.ContextKey("fiber-test")

	// Create a simple VNode tree using Element
	childVNode := NewElement("text").SetKey("child-test").SetProps(Props{"content": "Child"})
	parentVNode := NewElement("vstack").SetKey("parent-test").SetChildren([]VNode{childVNode})

	// Create Fiber tree
	root := CreateFiberFromVNode(parentVNode)
	if root == nil {
		t.Fatal("CreateFiberFromVNode should return a fiber")
	}

	// Verify root has a Context
	if root.Context == nil {
		t.Error("Root fiber should have Context")
	}

	// Provide value to root
	root.Context.Provide(TestKey, "root-value")

	// Navigate to child
	childFiber := root.Child
	if childFiber == nil {
		t.Fatal("Root should have a child fiber")
	}

	// Verify child inherited Context from parent
	if childFiber.Context == nil {
		t.Error("Child fiber should have Context inherited from parent")
	}

	// Verify child can access root's Context value
	value := childFiber.Context.UseContext(TestKey)
	if value != "root-value" {
		t.Errorf("Child should access root's context value: got %v", value)
	}
}

// TestFiberContext_Provider tests Provider component Context injection
func TestFiberContext_Provider(t *testing.T) {
	ProviderKey := fcontext.ContextKey("provider-key")
	ProviderValue := "provided-value"

	// Create a VNode tree with Provider
	childVNode := NewElement("text").SetKey("provider-child").SetProps(Props{"content": "Child"})
	providerVNode := NewProvider(ProviderKey, ProviderValue, childVNode)
	providerVNode.SetKey("provider-test")
	rootVNode := NewElement("vstack").SetKey("provider-root").SetChildren([]VNode{providerVNode})

	// Create Fiber tree
	root := CreateFiberFromVNode(rootVNode)
	if root == nil {
		t.Fatal("CreateFiberFromVNode should return a fiber")
	}

	// Navigate to provider
	providerFiber := root.Child
	if providerFiber == nil {
		t.Fatal("Root should have a child fiber (provider)")
	}

	if providerFiber.Tag != "provider" {
		t.Errorf("Expected provider tag, got %s", providerFiber.Tag)
	}

	// Verify provider has Context
	if providerFiber.Context == nil {
		t.Error("Provider fiber should have Context")
	}

	// Verify provider's Context has the provided value
	value := providerFiber.Context.UseContext(ProviderKey)
	if value != ProviderValue {
		t.Errorf("Provider should have provided value: got %v", value)
	}

	// Navigate to grandchild (actual content of provider)
	grandChildFiber := providerFiber.Child
	if grandChildFiber == nil {
		t.Fatal("Provider should have a child fiber")
	}

	// Verify grandchild inherited Context from provider
	if grandChildFiber.Context == nil {
		t.Error("Grandchild should have Context inherited from provider")
	}

	// Verify grandchild can access provider's Context value
	grandChildValue := grandChildFiber.Context.UseContext(ProviderKey)
	if grandChildValue != ProviderValue {
		t.Errorf("Grandchild should access provider's context value: got %v", grandChildValue)
	}
}

// TestFiberContext_NestedProviders tests that nested providers properly override values
func TestFiberContext_NestedProviders(t *testing.T) {
	InnerKey := fcontext.ContextKey("inner")
	OuterKey := fcontext.ContextKey("outer")

	// Create nested providers
	innerChildVNode := NewElement("text").SetKey("innermost").SetProps(Props{"content": "Inner Most"})
	innerProviderVNode := NewProvider(InnerKey, "inner-value", innerChildVNode)
	innerProviderVNode.SetKey("inner-provider")
	outerProviderVNode := NewProvider(OuterKey, "outer-value", innerProviderVNode)
	outerProviderVNode.SetKey("outer-provider")
	rootVNode := NewElement("vstack").SetKey("nested-root").SetChildren([]VNode{outerProviderVNode})

	// Create Fiber tree
	root := CreateFiberFromVNode(rootVNode)
	if root == nil {
		t.Fatal("CreateFiberFromVNode should return a fiber")
	}

	// Navigate through the tree
	outerProvider := root.Child
	if outerProvider == nil || outerProvider.Tag != "provider" {
		t.Fatal("First child should be outer provider")
	}

	innerProvider := outerProvider.Child
	if innerProvider == nil || innerProvider.Tag != "provider" {
		t.Fatal("Outer provider child should be inner provider")
	}

	innerMost := innerProvider.Child
	if innerMost == nil {
		t.Fatal("Inner provider should have a child (inner most)")
	}

	// Verify outer provider's context
	outerValue := outerProvider.Context.UseContext(OuterKey)
	if outerValue != "outer-value" {
		t.Errorf("Outer provider should have its value: got %v", outerValue)
	}

	// Verify inner provider inherits from outer
	inheritedValue := innerProvider.Context.UseContext(OuterKey)
	if inheritedValue != "outer-value" {
		t.Errorf("Inner provider should inherit from outer: got %v", inheritedValue)
	}

	// Verify inner provider has its own value
	innerValue := innerProvider.Context.UseContext(InnerKey)
	if innerValue != "inner-value" {
		t.Errorf("Inner provider should have its value: got %v", innerValue)
	}

	// Verify inner most inherits from both providers
	innerMostOuter := innerMost.Context.UseContext(OuterKey)
	if innerMostOuter != "outer-value" {
		t.Errorf("Inner most should inherit from outer: got %v", innerMostOuter)
	}

	innerMostInner := innerMost.Context.UseContext(InnerKey)
	if innerMostInner != "inner-value" {
		t.Errorf("Inner most should inherit from inner: got %v", innerMostInner)
	}
}

// TestFiberContext_UseContextValue tests type-safe context access through Fiber
func TestFiberContext_UseContextValue(t *testing.T) {
	IntKey := fcontext.ContextKey("int-key")
	StringKey := fcontext.ContextKey("string-key")

	childVNode := NewElement("text").SetKey("test-child").SetProps(Props{"content": "Test"})
	parentVNode := NewElement("vstack").SetKey("test-parent").SetChildren([]VNode{childVNode})

	root := CreateFiberFromVNode(parentVNode)
	if root == nil {
		t.Fatal("CreateFiberFromVNode should return a fiber")
	}

	// Provide typed values
	root.Context.Provide(IntKey, 42)
	root.Context.Provide(StringKey, "hello")

	// Access from root
	intVal, ok := fcontext.UseContextValue[int](root.Context, IntKey)
	if !ok || intVal != 42 {
		t.Errorf("Root should access int value: got %d, ok=%v", intVal, ok)
	}

	strVal, ok := fcontext.UseContextValue[string](root.Context, StringKey)
	if !ok || strVal != "hello" {
		t.Errorf("Root should access string value: got %s, ok=%v", strVal, ok)
	}

	// Access from child
	childFiber := root.Child
	if childFiber == nil {
		t.Fatal("Root should have a child")
	}

	childIntVal, ok := fcontext.UseContextValue[int](childFiber.Context, IntKey)
	if !ok || childIntVal != 42 {
		t.Errorf("Child should access parent's int value: got %d, ok=%v", childIntVal, ok)
	}

	// Type mismatch should fail
	_, ok = fcontext.UseContextValue[string](root.Context, IntKey)
	if ok {
		t.Error("Type mismatch should fail")
	}
}
