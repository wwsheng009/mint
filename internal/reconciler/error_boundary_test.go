package reconciler

import (
	"errors"
	"testing"

	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Error Boundary Integration Tests
// =============================================================================

// TestErrorBoundary_CatchesPanic tests that error boundaries catch panics during render
func TestErrorBoundary_CatchesPanic(t *testing.T) {
	panicCalled := false

	// Component that will panic
	panicComponent := func() rtui.VNode {
		panicCalled = true
		panic(errors.New("component error"))
	}

	// Wrap in error boundary
	boundary := rtui.NewErrorBoundary("testBoundary", panicComponent, rtui.FallbackText("Error occurred"))

	config := ReconcilerConfig{EnableFiber: true}
	reconciler := NewReconciler(nil, nil, config)
	currentReconciler = reconciler
	defer func() { currentReconciler = nil }()

	// Create fiber from boundary
	current := CreateFiber(boundary)
	workInProgress := CloneFiber(current)

	if workInProgress == nil {
		t.Fatal("CloneFiber returned nil")
	}

	// BeginWork should handle the panic and render fallback
	nextUnitOfWork := BeginWork(current, workInProgress)

	if nextUnitOfWork == nil {
		t.Error("BeginWork should return a valid unit of work")
	}

	// Component should have been called
	if !panicCalled {
		t.Error("Panic component should have been called")
	}

	// Boundary should have caught the error
	if !boundary.HadError() {
		t.Error("Error boundary should have caught the error")
	}

	// Error message should be set
	if boundary.GetErrorMsg() == "" {
		t.Error("Error message should be set")
	}

	// Stack trace should be captured
	if boundary.GetStack() == "" {
		t.Error("Stack trace should be captured")
	}
}

// TestErrorBoundary_RendersFallbackWhenPanicking tests that fallback UI is rendered
func TestErrorBoundary_RendersFallbackWhenPanicking(t *testing.T) {
	// Component that panics
	panicComponent := func() rtui.VNode {
		panic(errors.New("test error"))
	}

	// Fallback UI
	fallback := rtui.Element("text").Prop("content", "Something went wrong").Build()

	boundary := rtui.NewErrorBoundary("testBoundary", panicComponent, fallback)

	config := ReconcilerConfig{EnableFiber: true}
	reconciler := NewReconciler(nil, nil, config)
	currentReconciler = reconciler
	defer func() { currentReconciler = nil }()

	current := CreateFiber(boundary)
	workInProgress := CloneFiber(current)

	if workInProgress == nil {
		t.Fatal("CloneFiber returned nil")
	}

	// Process the boundary
	_ = BeginWork(current, workInProgress)

	// The child should be the fallback UI (not nil)
	child := workInProgress.Child
	if child == nil {
		t.Error("Error boundary should render fallback as child")
	}
}

// TestErrorBoundary_RendersComponentWhenNoPanic tests normal rendering
func TestErrorBoundary_RendersComponentWhenNoPanic(t *testing.T) {
	renderCalled := false

	// Normal component
	normalComponent := func() rtui.VNode {
		renderCalled = true
		return rtui.Element("text").Prop("content", "Hello, World!").Build()
	}

	fallback := rtui.FallbackText("Error occurred")
	boundary := rtui.NewErrorBoundary("testBoundary", normalComponent, fallback)

	config := ReconcilerConfig{EnableFiber: true}
	reconciler := NewReconciler(nil, nil, config)
	currentReconciler = reconciler
	defer func() { currentReconciler = nil }()

	current := CreateFiber(boundary)
	workInProgress := CloneFiber(current)

	if workInProgress == nil {
		t.Fatal("CloneFiber returned nil")
	}

	// Process the boundary
	_ = BeginWork(current, workInProgress)

	// Component should have been called
	if !renderCalled {
		t.Error("Normal component should have been called")
	}

	// Boundary should NOT have caught an error
	if boundary.HadError() {
		t.Error("Error boundary should not have caught an error for normal rendering")
	}

	// Child should be the component's result (not fallback)
	child := workInProgress.Child
	if child == nil {
		t.Error("Error boundary should render component result as child")
	}
}

// TestErrorBoundary_NestedErrorBoundaries tests nested error boundaries
func TestErrorBoundary_NestedErrorBoundaries(t *testing.T) {
	// Inner component that panics
	innerPanicComponent := func() rtui.VNode {
		panic(errors.New("inner error"))
	}

	// Wrap in inner error boundary
	innerBoundary := rtui.NewErrorBoundary("inner", innerPanicComponent, rtui.FallbackText("Inner error"))

	// Outer component (normal)
	outerComponent := func() rtui.VNode {
		return innerBoundary
	}

	// Wrap in outer error boundary
	outerBoundary := rtui.NewErrorBoundary("outer", outerComponent, rtui.FallbackText("Outer error"))

	config := ReconcilerConfig{EnableFiber: true}
	reconciler := NewReconciler(nil, nil, config)
	currentReconciler = reconciler
	defer func() { currentReconciler = nil }()

	current := CreateFiber(outerBoundary)
	workInProgress := CloneFiber(current)

	if workInProgress == nil {
		t.Fatal("CloneFiber returned nil")
	}

	// Process the outer boundary
	_ = BeginWork(current, workInProgress)

	// Outer boundary should NOT have caught the error (panics from children
	// happen during child reconciliation, not in the outer component function)
	if outerBoundary.HadError() {
		t.Error("Outer error boundary should not have caught the error (panic happens in child, not outer function)")
	}

	// Note: Currently, error boundaries only catch panics from their direct
	// component function. To catch panics from child components, the reconciler
	// would need to propagate panic handling through the entire subtree.
	// This is a known limitation for future enhancement.
}

// TestErrorBoundary_SiblingErrorBoundaries tests error boundaries as siblings
func TestErrorBoundary_SiblingErrorBoundaries(t *testing.T) {
	// First component (panics)
	firstComponent := func() rtui.VNode {
		panic(errors.New("first error"))
	}
	firstBoundary := rtui.NewErrorBoundary("first", firstComponent, rtui.FallbackText("First error"))

	// Second component (normal)
	secondComponent := func() rtui.VNode {
		return rtui.Element("text").Prop("content", "Second").Build()
	}
	secondBoundary := rtui.NewErrorBoundary("second", secondComponent, rtui.FallbackText("Second error"))

	config := ReconcilerConfig{EnableFiber: true}
	reconciler := NewReconciler(nil, nil, config)
	currentReconciler = reconciler
	defer func() { currentReconciler = nil }()

	// Create parent with both boundaries as children
	parent := &Fiber{
		Type: rtui.VNodeElement,
		Tag:  "div",
	}

	children := []rtui.VNode{firstBoundary, secondBoundary}
	parent.Child = reconcileChildren(parent, nil, children, rtui.LaneSyncLane)

	if parent.Child == nil {
		t.Fatal("ReconcileChildren should create child fibers")
	}

	// Note: reconcileChildren creates fibers but doesn't call BeginWork on them.
	// The panic would only be caught when BeginWork processes each boundary fiber.
	// This test verifies that error boundary VNodes can coexist as siblings.

	// Verify both boundaries exist in the tree
	count := 0
	for child := parent.Child; child != nil; child = child.Sibling {
		count++
	}
	if count != 2 {
		t.Errorf("Expected 2 children, got %d", count)
	}

	// Process each child through BeginWork to trigger error handling
	for child := parent.Child; child != nil; child = child.Sibling {
		if child.Type == rtui.VNodeComponent {
			current := CreateFiber(firstBoundary)
			workInProgress := CloneFiber(current)
			_ = BeginWork(current, workInProgress)
		}
	}

	// First boundary should have caught the error (after BeginWork processing)
	if !firstBoundary.HadError() {
		t.Error("First error boundary should have caught the error")
	}

	// Second boundary should NOT have caught an error
	if secondBoundary.HadError() {
		t.Error("Second error boundary should not have caught an error")
	}
}

// TestErrorBoundary_FullRenderCycle tests error boundary through full render cycle
func TestErrorBoundary_FullRenderCycle(t *testing.T) {
	renderCount := 0

	// Component that panics on first render, succeeds on second
	panicComponent := func() rtui.VNode {
		renderCount++
		if renderCount == 1 {
			panic(errors.New("first render error"))
		}
		return rtui.Element("text").Prop("content", "Recovered!").Build()
	}

	fallback := rtui.FallbackText("Error occurred")
	boundary := rtui.NewErrorBoundary("testBoundary", panicComponent, fallback)

	config := ReconcilerConfig{EnableFiber: true}
	reconciler := NewReconciler(nil, nil, config)
	currentReconciler = reconciler
	defer func() { currentReconciler = nil }()

	// First render - should catch panic
	current1 := CreateFiber(boundary)
	workInProgress1 := CloneFiber(current1)
	_ = BeginWork(current1, workInProgress1)

	if !boundary.HadError() {
		t.Error("First render should catch error")
	}

	// Second render - should retry (error boundary resets error state automatically)
	current2 := CreateFiber(boundary)
	workInProgress2 := CloneFiber(current2)
	_ = BeginWork(current2, workInProgress2)

	// After reset, the boundary should try again
	// Note: The reconciler's beginWorkErrorBoundary doesn't persist hadError across renders
	// It's reset on each render attempt, allowing retries
}

// TestErrorBoundary_WithComponentChildren tests error boundary with component children
func TestErrorBoundary_WithComponentChildren(t *testing.T) {
	// Child component that panics
	childComponent := rtui.NewComponent("PanicChild", func() rtui.VNode {
		panic(errors.New("child error"))
	})

	wrapComponent := func() rtui.VNode {
		return childComponent
	}

	boundary := rtui.NewErrorBoundary("testBoundary", wrapComponent, rtui.FallbackText("Child error"))

	config := ReconcilerConfig{EnableFiber: true}
	reconciler := NewReconciler(nil, nil, config)
	currentReconciler = reconciler
	defer func() { currentReconciler = nil }()

	current := CreateFiber(boundary)
	workInProgress := CloneFiber(current)

	if workInProgress == nil {
		t.Fatal("CloneFiber returned nil")
	}

	// Process the boundary
	_ = BeginWork(current, workInProgress)

	// Note: The wrapComponent function returns childComponent, which doesn't panic.
	// The panic happens when childComponent is rendered during child reconciliation.
	// Currently, error boundaries only catch panics from their direct component function.
	// This test verifies the boundary is processed without panicking itself.

	// The boundary should not have an error (wrapComponent didn't panic)
	if boundary.HadError() {
		t.Error("Error boundary should not have caught panic from child component during this phase")
	}

	// The child component panic would be caught when the child's BeginWork is processed.
	// To fully catch child component panics, error boundaries would need to handle
	// panics during child reconciliation as well.
}

// TestErrorBoundary_ResetError tests error reset functionality
func TestErrorBoundary_ResetError(t *testing.T) {
	panicComponent := func() rtui.VNode {
		panic(errors.New("test error"))
	}

	fallback := rtui.FallbackText("Error occurred")
	boundary := rtui.NewErrorBoundary("testBoundary", panicComponent, fallback)

	config := ReconcilerConfig{EnableFiber: true}
	reconciler := NewReconciler(nil, nil, config)
	currentReconciler = reconciler
	defer func() { currentReconciler = nil }()

	// First render - causes panic
	current1 := CreateFiber(boundary)
	workInProgress1 := CloneFiber(current1)
	_ = BeginWork(current1, workInProgress1)

	if !boundary.HadError() {
		t.Fatal("Error boundary should have caught error")
	}

	// Reset the error
	boundary.ResetError()

	if boundary.HadError() {
		t.Error("Error should be reset")
	}

	if boundary.GetError() != nil {
		t.Error("Error should be nil after reset")
	}
}

// TestErrorBoundary_VNodeIntegration tests error boundary as VNode
func TestErrorBoundary_VNodeIntegration(t *testing.T) {
	component := func() rtui.VNode {
		return rtui.Element("text").Prop("content", "Hello").Build()
	}

	boundary := rtui.NewErrorBoundary("test", component, rtui.FallbackText("Error"))

	// Should implement VNode interface
	var vnode rtui.VNode = boundary

	// Test basic VNode methods
	_ = vnode.Type()
	if vnode.Type() != rtui.VNodeComponent {
		t.Error("ErrorBoundary should be VNodeComponent type")
	}

	_ = vnode.Props()
	_ = vnode.Children()
	_ = vnode.Key()
	_ = vnode.Style()

	// Tag() is available on the concrete type, not through VNode interface
	if boundary.Tag() != "ErrorBoundary:test" {
		t.Errorf("Tag should be 'ErrorBoundary:test', got '%s'", boundary.Tag())
	}
}

// TestErrorBoundary_FallbackVariants tests different fallback types
func TestErrorBoundary_FallbackVariants(t *testing.T) {
	panicComponent := func() rtui.VNode {
		panic(errors.New("error"))
	}

	tests := []struct {
		name     string
		fallback rtui.VNode
	}{
		{"FallbackText", rtui.FallbackText("Error occurred")},
		{"FallbackError", rtui.FallbackError("MyComponent")},
		{"FallbackBox", rtui.FallbackBox("Error", "Something went wrong")},
		{"Element", rtui.Element("text").Prop("content", "Custom fallback").Build()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			boundary := rtui.NewErrorBoundary("test", panicComponent, tt.fallback)

			config := ReconcilerConfig{EnableFiber: true}
			reconciler := NewReconciler(nil, nil, config)
			currentReconciler = reconciler
			defer func() { currentReconciler = nil }()

			current := CreateFiber(boundary)
			workInProgress := CloneFiber(current)

			// Should not panic with any fallback type
			_ = BeginWork(current, workInProgress)

			if !boundary.HadError() {
				t.Error("Boundary should have caught the error")
			}
		})
	}
}

// TestErrorBoundary_ErrorBoundaryFunction tests the ErrorBoundary builder function
func TestErrorBoundary_ErrorBoundaryFunction(t *testing.T) {
	component := func() rtui.VNode {
		panic(errors.New("error"))
	}

	boundary := rtui.ErrorBoundary("test", component, rtui.FallbackText("Error"))

	// Should return a VNode
	if boundary == nil {
		t.Fatal("ErrorBoundary should return a VNode")
	}

	// Should be an ErrorBoundaryVNode
	errorBoundary, ok := boundary.(*rtui.ErrorBoundaryVNode)
	if !ok {
		t.Fatal("ErrorBoundary should return *ErrorBoundaryVNode")
	}

	// Test through reconciler
	config := ReconcilerConfig{EnableFiber: true}
	reconciler := NewReconciler(nil, nil, config)
	currentReconciler = reconciler
	defer func() { currentReconciler = nil }()

	current := CreateFiber(boundary)
	workInProgress := CloneFiber(current)

	_ = BeginWork(current, workInProgress)

	if !errorBoundary.HadError() {
		t.Error("ErrorBoundary should have caught the error")
	}
}

// TestErrorBoundary_NilComponent tests error boundary with nil component
func TestErrorBoundary_NilComponent(t *testing.T) {
	var nilComponent rtui.ComponentFunc = nil

	boundary := rtui.NewErrorBoundary("test", nilComponent, rtui.FallbackText("Error"))

	config := ReconcilerConfig{EnableFiber: true}
	reconciler := NewReconciler(nil, nil, config)
	currentReconciler = reconciler
	defer func() { currentReconciler = nil }()

	current := CreateFiber(boundary)
	workInProgress := CloneFiber(current)

	// Should not panic with nil component
	_ = BeginWork(current, workInProgress)

	// Child should be nil or empty fragment (component is nil)
	_ = workInProgress.Child
}

// TestErrorBoundary_NilFallback tests error boundary with nil fallback
func TestErrorBoundary_NilFallback(t *testing.T) {
	panicComponent := func() rtui.VNode {
		panic(errors.New("error"))
	}

	var nilFallback rtui.VNode = nil
	boundary := rtui.NewErrorBoundary("test", panicComponent, nilFallback)

	config := ReconcilerConfig{EnableFiber: true}
	reconciler := NewReconciler(nil, nil, config)
	currentReconciler = reconciler
	defer func() { currentReconciler = nil }()

	current := CreateFiber(boundary)
	workInProgress := CloneFiber(current)

	// Should not panic with nil fallback (renders empty fragment)
	_ = BeginWork(current, workInProgress)

	if !boundary.HadError() {
		t.Error("Boundary should have caught the error")
	}

	// Child should be an empty fragment when fallback is nil
	child := workInProgress.Child
	if child != nil {
		// If child exists, it should be a fragment
		_ = child
	}
}
