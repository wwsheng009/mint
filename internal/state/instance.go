package state

import "fmt"

// ComponentInstance represents a persistent component instance that maintains
// state across renders. This is similar to React's component instance model.
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
	SetProps(props map[string]interface{}) bool

	// GetProps returns the current properties
	GetProps() map[string]interface{}

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
	Render() interface{}

	// OnMount is called when the instance is first created
	OnMount()

	// OnUpdate is called before rendering with new props
	// Returns false to cancel the update
	OnUpdate(newProps, oldProps map[string]interface{}) bool

	// OnUnmount is called when the instance is being destroyed
	OnUnmount()

	// MarkDirty marks the instance as needing a re-render
	MarkDirty()

	// IsDirty returns whether the instance needs re-rendering
	IsDirty() bool
}

// ComponentContext manages hook state for a component instance
// This will be imported from ui/ to avoid duplication
type ComponentContext struct {
	Key             string
	Hooks           []Hook
	EffectQueue     []Effect
	Active          bool
	HookIndex       int
	RenderCount     int
	StateVersion    uint64
	PreviousContext *ComponentContext
}

// Hook represents a single hook call
type Hook struct {
	Type    HookType
	Value   interface{}
	Cleanup func()
}

// HookType represents the type of hook
type HookType int

const (
	HookState HookType = iota
	HookEffect
	HookMemo
	HookCallback
	HookRef
)

// Effect represents a useEffect effect
type Effect struct {
	Create func() func()
	Deps   []interface{}
	Clean  func()
}

// BaseComponentInstance provides a base implementation of ComponentInstance
// that can be embedded by specific component instances.
type BaseComponentInstance struct {
	key        string
	props      map[string]interface{}
	context    *ComponentContext
	fn         func() interface{}
	fnWithProps func(map[string]interface{}) interface{}
	dirty      bool
	mounted    bool
}

// NewBaseComponentInstance creates a new base component instance
func NewBaseComponentInstance(key string, fn func() interface{}) *BaseComponentInstance {
	return &BaseComponentInstance{
		key:     key,
		fn:      fn,
		props:   make(map[string]interface{}),
		context: NewComponentContext(key),
		dirty:   true,
		mounted: false,
	}
}

// NewBaseComponentInstanceWithProps creates a new base component instance with props
func NewBaseComponentInstanceWithProps(key string, fn func(map[string]interface{}) interface{}, props map[string]interface{}) *BaseComponentInstance {
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
func (b *BaseComponentInstance) SetProps(props map[string]interface{}) bool {
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
func (b *BaseComponentInstance) GetProps() map[string]interface{} {
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

// SetContext sets the component context
func (b *BaseComponentInstance) SetContext(ctx *ComponentContext) {
	b.context = ctx
}

// Render implements ComponentInstance
func (b *BaseComponentInstance) Render() interface{} {
	// Reset hook index for re-render
	b.context.ResetContext()

	// Set this instance's context as current
	SetCurrentContext(b.context)

	// Call the component function
	var vnode interface{}
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
func (b *BaseComponentInstance) OnUpdate(newProps, oldProps map[string]interface{}) bool {
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

// propsEqual compares two Props maps for equality
func propsEqual(a, b map[string]interface{}) bool {
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

// =============================================================================
// ComponentContext Management
// =============================================================================

// NewComponentContext creates a new component context
func NewComponentContext(key string) *ComponentContext {
	return &ComponentContext{
		Key:         key,
		Hooks:       make([]Hook, 0),
		EffectQueue: make([]Effect, 0),
		Active:      true,
		HookIndex:   0,
		RenderCount: 0,
	}
}

// ResetContext resets the context for a new render
func (c *ComponentContext) ResetContext() {
	c.HookIndex = 0
	c.RenderCount++
	c.StateVersion++
}

// AddHook adds a hook to the context
func (c *ComponentContext) AddHook(hook Hook) {
	if c.HookIndex == len(c.Hooks) {
		c.Hooks = append(c.Hooks, hook)
	} else {
		c.Hooks[c.HookIndex] = hook
	}
	c.HookIndex++
}

// GetHook returns the hook at the current index
func (c *ComponentContext) GetHook() Hook {
	if c.HookIndex < len(c.Hooks) {
		return c.Hooks[c.HookIndex]
	}
	return Hook{}
}

// CleanupAll runs all cleanup functions
func (c *ComponentContext) CleanupAll() {
	for _, hook := range c.Hooks {
		if hook.Cleanup != nil {
			hook.Cleanup()
		}
	}
}

// =============================================================================
// Global Context Management
// =============================================================================

var currentContext *ComponentContext

// GetCurrentContext returns the current component context
func GetCurrentContext() *ComponentContext {
	return currentContext
}

// SetCurrentContext sets the current component context
func SetCurrentContext(ctx *ComponentContext) {
	currentContext = ctx
}
