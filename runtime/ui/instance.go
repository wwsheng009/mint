package ui

import "fmt"

// ComponentInstance represents a persistent component instance that maintains
// state across renders.
//
// Architecture:
//
//	VNode (created each render)  →  ComponentInstance (persists across renders)
//	                                    ├─ state: { isHovered, value, ... }
//	                                    ├─ props: { label, onClick, ... }
//	                                    ├─ hooks: [ useState, useEffect, ... ]
//	                                    └─ key: string (for matching)
//
// The instance is created on first render and reused on subsequent renders
// when a VNode with the same key is encountered.
type ComponentInstance interface {
	// Key returns the unique identifier for this instance
	// Used to match VNodes to instances across renders
	Key() string

	// SetKey sets the instance key
	SetKey(key string)

	// SetProps updates the component properties
	// Returns true if props changed (may trigger re-render)
	SetProps(props Props) bool

	// GetProps returns the current properties
	GetProps() Props

	// GetState returns the component's state map
	// This includes all state from useState hooks
	GetState() map[string]interface{}

	// SetState updates a specific state value and triggers re-render
	SetState(key string, value interface{})

	// GetContext returns the component's hook context
	// This persists hooks across renders
	GetContext() *ComponentContext

	// Render calls the component function and returns the new VNode
	// This is called each render cycle with the persistent context
	Render() VNode

	// OnMount is called when the instance is first created
	OnMount()

	// OnUpdate is called before rendering with new props
	// Returns false to cancel the update
	OnUpdate(newProps, oldProps Props) bool

	// OnUnmount is called when the instance is being destroyed
	OnUnmount()

	// MarkDirty marks the instance as needing a re-render
	MarkDirty()

	// IsDirty returns whether the instance needs re-rendering
	IsDirty() bool
}

// BaseComponentInstance provides a base implementation of ComponentInstance
// that can be embedded by specific component instances.
type BaseComponentInstance struct {
	key        string
	props      Props
	context    *ComponentContext
	fn         ComponentFunc
	fnWithProps ComponentFuncWithProps
	dirty      bool
	mounted    bool
}

// NewBaseComponentInstance creates a new base component instance
func NewBaseComponentInstance(key string, fn ComponentFunc) *BaseComponentInstance {
	return &BaseComponentInstance{
		key:     key,
		fn:      fn,
		props:   make(Props),
		context: NewComponentContext(key),
		dirty:   true,
		mounted: false,
	}
}

// NewBaseComponentInstanceWithProps creates a new base component instance with props
func NewBaseComponentInstanceWithProps(key string, fn ComponentFuncWithProps, props Props) *BaseComponentInstance {
	return &BaseComponentInstance{
		key:         key,
		fnWithProps: fn,
		props:       props,
		context:     NewComponentContext(key),
		dirty:       true,
		mounted:     false,
	}
}

// Key implements ComponentInstance
func (b *BaseComponentInstance) Key() string {
	return b.key
}

// SetKey implements ComponentInstance
func (b *BaseComponentInstance) SetKey(key string) {
	b.key = key
}

// SetProps implements ComponentInstance
func (b *BaseComponentInstance) SetProps(props Props) bool {
	// Check if props actually changed
	if propsEqual(b.props, props) {
		return false
	}
	oldProps := b.props
	b.props = props
	// Call OnUpdate if implemented by embedded type
	if b.OnUpdate(props, oldProps) {
		b.MarkDirty()
		return true
	}
	return false
}

// GetProps implements ComponentInstance
func (b *BaseComponentInstance) GetProps() Props {
	return b.props
}

// GetState implements ComponentInstance
func (b *BaseComponentInstance) GetState() map[string]interface{} {
	state := make(map[string]interface{})
	// Collect state from hooks
	for i, hook := range b.context.Hooks {
		if hook.Type == HookState {
			state[fmt.Sprintf("hook_%d", i)] = hook.Value
		}
	}
	return state
}

// SetState implements ComponentInstance
func (b *BaseComponentInstance) SetState(key string, value interface{}) {
	// This is a low-level API - normally users use useState
	// But it allows direct state manipulation if needed
	b.MarkDirty()
}

// GetContext implements ComponentInstance
func (b *BaseComponentInstance) GetContext() *ComponentContext {
	return b.context
}

// Render implements ComponentInstance
func (b *BaseComponentInstance) Render() VNode {
	// Reset hook index for re-render
	b.context.ResetContext()

	// Set this instance's context as current
	SetCurrentContext(b.context)

	// Call the component function
	var vnode VNode
	if b.fn != nil {
		vnode = b.fn()
	} else if b.fnWithProps != nil {
		vnode = b.fnWithProps(b.props)
	}

	// Clear current context
	SetCurrentContext(nil)

	return vnode
}

// OnMount implements ComponentInstance
func (b *BaseComponentInstance) OnMount() {
	b.mounted = true
}

// OnUpdate implements ComponentInstance
// Returns true to allow the update, false to cancel
func (b *BaseComponentInstance) OnUpdate(newProps, oldProps Props) bool {
	return true // Allow update by default
}

// OnUnmount implements ComponentInstance
func (b *BaseComponentInstance) OnUnmount() {
	// Run cleanup functions
	b.context.CleanupAll()
	b.mounted = false
}

// MarkDirty implements ComponentInstance
func (b *BaseComponentInstance) MarkDirty() {
	b.dirty = true
}

// IsDirty implements ComponentInstance
func (b *BaseComponentInstance) IsDirty() bool {
	return b.dirty
}

// ClearDirty clears the dirty flag
func (b *BaseComponentInstance) ClearDirty() {
	b.dirty = false
}

// IsMounted returns whether the component is mounted
func (b *BaseComponentInstance) IsMounted() bool {
	return b.mounted
}

// propsEqual compares two Props for equality
func propsEqual(a, b Props) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}
