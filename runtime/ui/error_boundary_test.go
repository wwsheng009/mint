package ui

import (
	"errors"
	"testing"

	"github.com/wwsheng009/mint/runtime/style"
)

// TestErrorBoundary_BasicPanicRecovery tests that error boundaries catch panics
func TestErrorBoundary_BasicPanicRecovery(t *testing.T) {
	// Create a component that will panic
	panicComponent := func() VNode {
		panic(errors.New("test panic"))
	}

	// Create a fallback
	fallback := FallbackText("Error occurred")

	// Create error boundary
	boundary := NewErrorBoundary("testBoundary", panicComponent, fallback)

	// Call Render to trigger panic recovery
	result := boundary.Render()

	// Should have caught the error
	if !boundary.HadError() {
		t.Error("Expected boundary to have caught an error")
	}

	// Check error message
	if boundary.GetErrorMsg() == "" {
		t.Error("Expected error message to be set")
	}

	// Check stack trace is captured
	if boundary.GetStack() == "" {
		t.Error("Expected stack trace to be captured")
	}

	// Render should return nil when panic occurs (fallback is handled by reconciler)
	if result != nil {
		t.Errorf("Expected nil result after panic, got %v", result)
	}
}

// TestErrorBoundary_NoPanic tests that components work normally without panics
func TestErrorBoundary_NoPanic(t *testing.T) {
	// Create a component that won't panic
	normalComponent := func() VNode {
		return Element("text").Prop("content", "Hello, World!").Build()
	}

	fallback := FallbackText("Error occurred")
	boundary := NewErrorBoundary("testBoundary", normalComponent, fallback)

	// Call Render
	result := boundary.Render()

	// Should not have caught an error
	if boundary.HadError() {
		t.Error("Expected boundary to not have caught an error")
	}

	// Should return the component's result
	if result == nil {
		t.Error("Expected non-nil result from normal component")
	}
}

// TestErrorBoundary_ResetError tests that errors can be reset
func TestErrorBoundary_ResetError(t *testing.T) {
	panicComponent := func() VNode {
		panic(errors.New("test panic"))
	}

	fallback := FallbackText("Error occurred")
	boundary := NewErrorBoundary("testBoundary", panicComponent, fallback)

	// Trigger panic
	boundary.Render()

	if !boundary.HadError() {
		t.Fatal("Expected boundary to have caught an error")
	}

	// Reset error
	boundary.ResetError()

	if boundary.HadError() {
		t.Error("Expected error to be reset")
	}

	if boundary.GetError() != nil {
		t.Error("Expected error to be nil after reset")
	}

	if boundary.GetErrorMsg() != "" {
		t.Error("Expected error message to be empty after reset")
	}

	if boundary.GetStack() != "" {
		t.Error("Expected stack to be empty after reset")
	}
}

// TestErrorBoundary_AccessorMethods tests the accessor methods
func TestErrorBoundary_AccessorMethods(t *testing.T) {
	component := func() VNode {
		return Element("text").Prop("content", "test").Build()
	}
	fallback := FallbackText("fallback")

	boundary := NewErrorBoundary("test", component, fallback)

	// Test Component() returns the component function
	if boundary.Component() == nil {
		t.Error("Expected Component() to return the component function")
	}

	// Test Fallback() returns the fallback VNode
	if boundary.Fallback() == nil {
		t.Error("Expected Fallback() to return the fallback VNode")
	}

	// Test Name() returns the name
	if boundary.Name() != "test" {
		t.Errorf("Expected Name() to return 'test', got '%s'", boundary.Name())
	}
}

// TestErrorBoundary_Tag tests the Tag method
func TestErrorBoundary_Tag(t *testing.T) {
	boundary := NewErrorBoundary("myBoundary", func() VNode {
		return Element("text").Prop("content", "").Build()
	}, FallbackText("err"))

	expected := "ErrorBoundary:myBoundary"
	if boundary.Tag() != expected {
		t.Errorf("Expected Tag() to return '%s', got '%s'", expected, boundary.Tag())
	}
}

// TestErrorBoundary_VNodeInterface tests that ErrorBoundaryVNode implements VNode
func TestErrorBoundary_VNodeInterface(t *testing.T) {
	boundary := NewErrorBoundary("test", func() VNode {
		return Element("text").Prop("content", "").Build()
	}, FallbackText("err"))

	var vnode VNode = boundary

	// Test all VNode methods are implemented
	_ = vnode.Type()
	_ = vnode.Props()
	_ = vnode.Children()
	vnode.SetProps(make(Props))
	vnode.SetChildren(nil)
	_ = vnode.Key()
	vnode.SetKey("test")
	_ = vnode.Style()
	vnode.SetStyle(style.Style{})
	_ = vnode.Tag()
}

// TestFallbackText tests the FallbackText helper
func TestFallbackText(t *testing.T) {
	fallback := FallbackText("Something went wrong")

	if fallback == nil {
		t.Error("Expected FallbackText to return a non-nil VNode")
	}
}

// TestFallbackError tests the FallbackError helper
func TestFallbackError(t *testing.T) {
	fallback := FallbackError("MyComponent")

	if fallback == nil {
		t.Error("Expected FallbackError to return a non-nil VNode")
	}
}

// TestFallbackBox tests the FallbackBox helper
func TestFallbackBox(t *testing.T) {
	fallback := FallbackBox("Error Title", "Error message")

	if fallback == nil {
		t.Error("Expected FallbackBox to return a non-nil VNode")
	}
}

// TestErrorBoundary_Function tests the ErrorBoundary builder function
func TestErrorBoundary_Function(t *testing.T) {
	component := func() VNode {
		return Element("text").Prop("content", "").Build()
	}
	fallback := FallbackText("error")

	boundary := ErrorBoundary("test", component, fallback)

	if boundary == nil {
		t.Error("Expected ErrorBoundary to return a non-nil VNode")
	}

	// Should be an ErrorBoundaryVNode
	errorBoundary, ok := boundary.(*ErrorBoundaryVNode)
	if !ok {
		t.Error("Expected ErrorBoundary to return *ErrorBoundaryVNode")
	}

	// Test the wrapper has the correct name
	if errorBoundary.Name() != "test" {
		t.Errorf("Expected name to be 'test', got '%s'", errorBoundary.Name())
	}
}

// TestErrorBoundary_SetError tests the SetError method
func TestErrorBoundary_SetError(t *testing.T) {
	boundary := NewErrorBoundary("test", func() VNode {
		return Element("text").Prop("content", "").Build()
	}, FallbackText("err"))

	testErr := errors.New("test error")
	testMsg := "test message"
	testStack := "test stack"

	boundary.SetError(testErr, testMsg, testStack)

	if !boundary.HadError() {
		t.Error("Expected HadError() to return true after SetError")
	}

	if boundary.GetError() != testErr {
		t.Error("Expected GetError() to return the test error")
	}

	if boundary.GetErrorMsg() != testErr.Error() {
		t.Errorf("Expected GetErrorMsg() to return '%s', got '%s'", testErr.Error(), boundary.GetErrorMsg())
	}

	if boundary.GetStack() != testStack {
		t.Error("Expected GetStack() to return the test stack")
	}
}

// TestErrorBoundary_RenderRetry tests that Render retries after a previous error
func TestErrorBoundary_RenderRetry(t *testing.T) {
	panicCount := 0
	panicComponent := func() VNode {
		panicCount++
		if panicCount == 1 {
			panic(errors.New("first panic"))
		}
		return Element("text").Prop("content", "Recovered!").Build()
	}

	fallback := FallbackText("Error occurred")
	boundary := NewErrorBoundary("testBoundary", panicComponent, fallback)

	// First render should catch the panic
	boundary.Render()
	if !boundary.HadError() {
		t.Error("Expected first render to catch error")
	}

	// Second render should clear the error and retry
	result := boundary.Render()
	if boundary.HadError() {
		t.Error("Expected second render to not have error (retried successfully)")
	}
	if result == nil {
		t.Error("Expected non-nil result on successful retry")
	}

	if panicCount != 2 {
		t.Errorf("Expected component to be called twice, got %d calls", panicCount)
	}
}
