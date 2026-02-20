package ui

import (
	"github.com/wwsheng009/mint/runtime/style"
)

// =============================================================================
// Memo Component
// =============================================================================
// Memo is a performance optimization component that skips re-rendering
// if its props haven't changed. Similar to React.memo().
//
// Usage:
//
//	Memo(
//		ExpensiveComponent,
//		func(oldProps, newProps Props) bool {
//			return !propsEqual(oldProps, newProps)
//		},
//	)
//
// If no comparison function is provided, a shallow equality check is used.

// PropsEqual is the default comparison function for memo
// It does a shallow comparison of two Props objects
type PropsEqual func(oldProps, newProps Props) bool

// MemoVNode represents a memoized component wrapper
type MemoVNode struct {
	component    VNode      // The component to memoize
	compare      PropsEqual // Optional custom comparison function
	lastProps    Props      // Cached props from last render
	memoizedVNode VNode     // Cached rendered result
	key          string
	style        style.Style
}

// NewMemo creates a new memo VNode with default shallow comparison
func NewMemo(component VNode) *MemoVNode {
	return &MemoVNode{
		component:    component,
		compare:      ShallowPropsEqual,
		memoizedVNode: nil, // Will be set on first render
	}
}

// NewMemoWithCompare creates a new memo VNode with custom comparison
func NewMemoWithCompare(component VNode, compare PropsEqual) *MemoVNode {
	return &MemoVNode{
		component:    component,
		compare:      compare,
		memoizedVNode: nil,
	}
}

// Type implements VNode
func (m *MemoVNode) Type() VNodeType {
	return VNodeComponent
}

// Props implements VNode
func (m *MemoVNode) Props() Props {
	return m.component.Props()
}

// SetProps implements VNode - returns VNode for chaining
func (m *MemoVNode) SetProps(p Props) VNode {
	m.component.SetProps(p)
	return m
}

// Children implements VNode
func (m *MemoVNode) Children() []VNode {
	return m.component.Children()
}

// SetChildren implements VNode - returns VNode for chaining
func (m *MemoVNode) SetChildren(children []VNode) VNode {
	m.component.SetChildren(children)
	return m
}

// Key implements VNode
func (m *MemoVNode) Key() string {
	if m.key != "" {
		return m.key
	}
	return m.component.Key()
}

// SetKey implements VNode - returns VNode for chaining
func (m *MemoVNode) SetKey(key string) VNode {
	m.key = key
	return m
}

// Style implements VNode
func (m *MemoVNode) Style() style.Style {
	return m.style
}

// SetStyle implements VNode - returns VNode for chaining
func (m *MemoVNode) SetStyle(s style.Style) VNode {
	m.style = s
	return m
}

// Tag implements VNode
func (m *MemoVNode) Tag() string {
	if comp, ok := m.component.(interface{ Tag() string }); ok {
		return "Memo:" + comp.Tag()
	}
	return "Memo:Component"
}

// GetLayer returns the layer from props (delegates to wrapped component)
func (m *MemoVNode) GetLayer() Layer {
	if m.component != nil {
		return m.component.GetLayer()
	}
	return LayerBase
}

// SetLayer sets the layer (delegates to wrapped component)
func (m *MemoVNode) SetLayer(l Layer) VNode {
	if m.component != nil {
		m.component = m.component.SetLayer(l)
	}
	return m
}

// Render returns the memoized component or re-renders if props changed
func (m *MemoVNode) Render() VNode {
	currentProps := m.component.Props()

	// First render - always render
	if m.lastProps == nil {
		m.lastProps = currentProps
		m.memoizedVNode = m.component
		return m.memoizedVNode
	}

	// Check if props changed using comparison function
	propsChanged := m.compare(m.lastProps, currentProps)

	if propsChanged {
		// Props changed - update cache and re-render
		m.lastProps = currentProps
		m.memoizedVNode = m.component
	}

	return m.memoizedVNode
}

// ShouldUpdate returns true if the component should update based on prop comparison
func (m *MemoVNode) ShouldUpdate(newProps Props) bool {
	if m.lastProps == nil {
		return true
	}
	return m.compare(m.lastProps, newProps)
}

// GetMemoizedChild returns the cached VNode
func (m *MemoVNode) GetMemoizedChild() VNode {
	return m.memoizedVNode
}

// GetComponent returns the wrapped component
func (m *MemoVNode) GetComponent() VNode {
	return m.component
}

// GetCompare returns the comparison function
func (m *MemoVNode) GetCompare() PropsEqual {
	return m.compare
}

// =============================================================================
// Default Comparison Functions
// =============================================================================

// ShallowPropsEqual performs a shallow comparison of two Props objects
// Returns true if they are equal (no update needed)
func ShallowPropsEqual(oldProps, newProps Props) bool {
	if oldProps == nil && newProps == nil {
		return true
	}
	if oldProps == nil || newProps == nil {
		return false
	}

	// Quick length check
	if len(oldProps) != len(newProps) {
		return false
	}

	// Check each key-value pair
	for key, newValue := range newProps {
		oldValue, exists := oldProps[key]
		if !exists {
			return false
		}

		// Use == for comparison (works for primitives, interfaces, pointers)
		if oldValue != newValue {
			return false
		}
	}

	return true
}

// CustomPropsEqual creates a comparison function from a predicate
// The predicate returns true if props are equal (no update needed)
func CustomPropsEqual(predicate func(oldProps, newProps Props) bool) PropsEqual {
	return predicate
}

// =============================================================================
// Memo Builder
// =============================================================================

// MemoBuilder provides fluent API for building memo components
type MemoBuilder struct {
	component VNode
	compare   PropsEqual
}

// Memo creates a new memo builder
func Memo(component VNode) *MemoBuilder {
	return &MemoBuilder{
		component: component,
		compare:   ShallowPropsEqual,
	}
}

// Compare sets a custom comparison function
func (b *MemoBuilder) Compare(compare PropsEqual) *MemoBuilder {
	b.compare = compare
	return b
}

// Key sets the key for diffing
func (b *MemoBuilder) Key(key string) *MemoBuilder {
	if vnode, ok := b.component.(*ComponentVNode); ok {
		vnode.SetKey(key)
	}
	return b
}

// Style sets the visual style
func (b *MemoBuilder) Style(s style.Style) *MemoBuilder {
	if vnode, ok := b.component.(interface{ SetStyle(style.Style) }); ok {
		vnode.SetStyle(s)
	}
	b.component.SetStyle(s)
	return b
}

// Build returns the memo VNode
func (b *MemoBuilder) Build() VNode {
	return NewMemoWithCompare(b.component, b.compare)
}

// =============================================================================
// Convenience Functions
// =============================================================================

// MemoComponent wraps a component function with memoization
// This is a convenience wrapper for the common case of memoizing a component function
func MemoComponent(name string, fn ComponentFunc) VNode {
	component := NewComponent(name, fn)
	return Memo(component).Build()
}

// MemoComponentWithProps wraps a component with props and memoization
func MemoComponentWithProps(name string, fn ComponentFuncWithProps, props Props) VNode {
	component := NewComponentWithProps(name, fn)
	component.SetProps(props)
	return Memo(component).Build()
}

// =============================================================================
// Advanced Comparison Functions
// =============================================================================

// PropsEqualExcept creates a comparison that ignores specific keys
func PropsEqualExcept(exceptKeys ...string) PropsEqual {
	return func(oldProps, newProps Props) bool {
		if oldProps == nil && newProps == nil {
			return true
		}
		if oldProps == nil || newProps == nil {
			return false
		}

		// Create sets for faster lookup
		exceptSet := make(map[string]bool)
		for _, key := range exceptKeys {
			exceptSet[key] = true
		}

		// Check all keys in newProps
		for key, newValue := range newProps {
			if exceptSet[key] {
				continue // Skip ignored keys
			}
			oldValue, exists := oldProps[key]
			if !exists || oldValue != newValue {
				return false
			}
		}

		// Check for removed keys (except ignored ones)
		for key := range oldProps {
			if !exceptSet[key] {
				if _, exists := newProps[key]; !exists {
					return false
				}
			}
		}

		return true
	}
}

// PropsEqualOnly creates a comparison that only checks specific keys
func PropsEqualOnly(keys ...string) PropsEqual {
	return func(oldProps, newProps Props) bool {
		if oldProps == nil && newProps == nil {
			return true
		}
		if oldProps == nil || newProps == nil {
			return false
		}

		// Only check the specified keys
		for _, key := range keys {
			oldValue, oldExists := oldProps[key]
			newValue, newExists := newProps[key]

			if oldExists != newExists {
				return false
			}

			if oldExists && oldValue != newValue {
				return false
			}
		}

		return true
	}
}
