package ui

import (
	"fmt"
	"runtime/debug"

	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/runtime/types"
)

// =============================================================================
// Error Boundary
// =============================================================================
// ErrorBoundary catches errors from component rendering and displays
// a fallback UI instead. This prevents the entire app from crashing due to
// a single component error.
//
// Usage:
//
//	ErrorBoundary(
//		"myBoundary",
//		func() VNode { return MyComponent() },
//		FallbackText("Something went wrong"),
//	)
//
// Error Boundary API:
//   - Wrap components that might throw panics
//   - Provide fallback UI to display when errors occur
//   - Access error information via GetError() method
//   - Reset error state to retry rendering

// ErrorBoundaryVNode represents an error boundary wrapper
type ErrorBoundaryVNode struct {
	key      string
	id       string
	name     string
	component ComponentFunc
	fallback VNode
	props    Props
	layer    Layer
	// Error state
	hadError bool
	error    error
	errorMsg string
	stack     string
}

// NewErrorBoundary creates a new error boundary VNode
func NewErrorBoundary(name string, component ComponentFunc, fallback VNode) *ErrorBoundaryVNode {
	return &ErrorBoundaryVNode{
		name:     name,
		component: component,
		fallback: fallback,
	}
}

// Type implements VNode
func (e *ErrorBoundaryVNode) Type() VNodeType {
	return VNodeComponent
}

// Props implements VNode
func (e *ErrorBoundaryVNode) Props() Props {
	return e.props
}

// SetProps implements VNode - returns VNode for chaining
func (e *ErrorBoundaryVNode) SetProps(p Props) VNode {
	// Error boundary doesn't use props
	return e
}

// Children implements VNode
func (e *ErrorBoundaryVNode) Children() []VNode {
	// Children are rendered by the component function
	return nil
}

// SetChildren implements VNode - returns VNode for chaining
func (e *ErrorBoundaryVNode) SetChildren(children []VNode) VNode {
	// Error boundary doesn't have direct children
	return e
}

// Key implements VNode
func (e *ErrorBoundaryVNode) Key() string {
	return e.key
}

// SetKey implements VNode - returns VNode for chaining
func (e *ErrorBoundaryVNode) SetKey(key string) VNode {
	e.key = key
	return e
}

// ID implements VNode - returns the business identifier for error boundary reference/positioning
func (e *ErrorBoundaryVNode) ID() string {
	return e.id
}

// SetID implements VNode - sets the business identifier and returns VNode for chaining
func (e *ErrorBoundaryVNode) SetID(id string) VNode {
	e.id = id
	return e
}

// Style implements VNode
func (e *ErrorBoundaryVNode) Style() style.Style {
	return style.Style{}
}

// SetStyle implements VNode - returns VNode for chaining
func (e *ErrorBoundaryVNode) SetStyle(s style.Style) VNode {
	// Error boundary doesn't use style directly
	return e
}

// Tag implements VNode
func (e *ErrorBoundaryVNode) Tag() string {
	return "ErrorBoundary:" + e.name
}

// GetLayer returns the layer for this error boundary
func (e *ErrorBoundaryVNode) GetLayer() Layer {
	return e.layer
}

// SetLayer sets the layer for this error boundary
func (e *ErrorBoundaryVNode) SetLayer(l Layer) VNode {
	e.layer = l
	return e
}

// =============================================================================
// Portal Methods - Chainable methods for Portal configuration
// =============================================================================

// SetPortalRoot implements VNode - sets the portalRoot property
func (e *ErrorBoundaryVNode) SetPortalRoot(portalRootID string) VNode {
	if e.props == nil {
		e.props = make(Props)
	}
	e.props["portalRoot"] = portalRootID
	return e
}

// SetAnchorTo implements VNode - sets anchorId and anchor properties
func (e *ErrorBoundaryVNode) SetAnchorTo(anchorID string, anchor types.Anchor) VNode {
	if e.props == nil {
		e.props = make(Props)
	}
	e.props["anchorId"] = anchorID
	e.props["anchor"] = anchor
	return e
}

// SetPortalPosition implements VNode - sets the position property
func (e *ErrorBoundaryVNode) SetPortalPosition(position types.PositionType) VNode {
	if e.props == nil {
		e.props = make(Props)
	}
	e.props["position"] = position
	return e
}

// SetPortalPriority implements VNode - sets the priority property
func (e *ErrorBoundaryVNode) SetPortalPriority(priority int) VNode {
	if e.props == nil {
		e.props = make(Props)
	}
	e.props["priority"] = priority
	return e
}

// SetPortalRootId implements VNode - sets the portalRootId property
func (e *ErrorBoundaryVNode) SetPortalRootId(portalRootId string) VNode {
	if e.props == nil {
		e.props = make(Props)
	}
	e.props["portalRootId"] = portalRootId
	return e
}

// Name returns the error boundary name
func (e *ErrorBoundaryVNode) Name() string {
	return e.name
}

// Render implements component rendering
func (e *ErrorBoundaryVNode) Render() VNode {
	// If there was a previous error, try to recover
	if e.hadError {
		// Clear error and retry
		e.hadError = false
		e.error = nil
		e.errorMsg = ""
		e.stack = ""
	}

	// Call the component function with panic recovery
	defer func() {
		if r := recover(); r != nil {
			e.hadError = true
			e.error = r.(error)
			e.errorMsg = fmt.Sprintf("panic in %s: %v", e.name, r)
			e.stack = string(debug.Stack())
		}
	}()

	// Render the component
	if e.component != nil {
		return e.component()
	}

	return nil
}

// GetError returns the current error state
func (e *ErrorBoundaryVNode) GetError() error {
	return e.error
}

// GetErrorMsg returns the error message
func (e *ErrorBoundaryVNode) GetErrorMsg() string {
	if e.error != nil {
		return e.error.Error()
	}
	return e.errorMsg
}

// GetStack returns the stack trace from the error
func (e *ErrorBoundaryVNode) GetStack() string {
	return e.stack
}

// HadError returns true if this boundary caught an error
func (e *ErrorBoundaryVNode) HadError() bool {
	return e.hadError
}

// ResetError clears the error state and allows retry
func (e *ErrorBoundaryVNode) ResetError() {
	e.hadError = false
	e.error = nil
	e.errorMsg = ""
	e.stack = ""
}

// Component returns the wrapped component function
func (e *ErrorBoundaryVNode) Component() ComponentFunc {
	return e.component
}

// Fallback returns the fallback VNode to render on error
func (e *ErrorBoundaryVNode) Fallback() VNode {
	return e.fallback
}

// SetError sets the error state (used by the reconciler)
func (e *ErrorBoundaryVNode) SetError(err error, msg string, stack string) {
	e.hadError = true
	e.error = err
	e.errorMsg = msg
	e.stack = stack
}

// ErrorBoundary creates a new error boundary
//   - name: identifier for this boundary (for debugging)
//   - component: the component function to wrap
//   - fallback: the VNode to render when errors occur
func ErrorBoundary(name string, component ComponentFunc, fallback VNode) VNode {
	boundary := NewErrorBoundary(name, component, fallback)
	return boundary
}

// =============================================================================
// Fallback Helpers
// =============================================================================

// FallbackText creates a simple text fallback
func FallbackText(text string) VNode {
	return Element("text").Prop("content", text).Build()
}

// FallbackError creates an error message fallback with details
func FallbackError(prefix string) VNode {
	errorText := fmt.Sprintf("%s: an error occurred. See logs for details.", prefix)
	return FallbackText(errorText)
}

// FallbackBox creates a boxed error message
func FallbackBox(title, message string) VNode {
	return VStack(
		Element("text").Prop("content", title).Prop("style", style.Style{}.Bold(true)).Build(),
		Element("text").Prop("content", message).Build(),
	)
}
