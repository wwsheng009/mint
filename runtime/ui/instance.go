package ui

import (
	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
)

// =============================================================================
// ComponentInstance - Fiber-first Runtime Instance
// =============================================================================
// ComponentInstance is the runtime entity for components.
// It persists across renders and holds all state.
//
// Architecture (from fiber_paint.md):
//
//	VNode (描述) ──创建──→ Instance (运行期实体)
//	                       ↓
//	                 Fiber.Instance = instance
//	                       ↓
//	                 生命周期: Mount → Update* → Unmount
//
// Key Points:
//   - VNode is created every render (description only)
//   - Instance is created once and persists (stateful)
//   - Fiber holds Instance reference, NOT VNode reference
//   - Instance.Paint() is called during commit phase

// ComponentInstance is the core runtime interface for all components.
// This matches the design in fiber_paint.md.
type ComponentInstance interface {
	// === Identification ===
	Key() string
	SetKey(key string)

	// === Lifecycle ===
	Init(props Props)
	Destroy()
	OnMount()
	OnUnmount()

	// === Props Management ===
	SetProps(props Props) bool
	GetProps() Props

	// === State ===
	MarkDirty()
	IsDirty() bool

	// === Context (for hooks) ===
	GetContext() *ComponentContext
}

// =============================================================================
// BaseComponentInstance - Base implementation
// =============================================================================

// BaseComponentInstance provides a base implementation of ComponentInstance
type BaseComponentInstance struct {
	key         string
	props       Props
	context     *ComponentContext
	fn          ComponentFunc
	fnWithProps ComponentFuncWithProps
	dirty       bool
	mounted     bool
}

// NewBaseComponentInstance creates a new base instance
func NewBaseComponentInstance(key string, fn ComponentFunc) *BaseComponentInstance {
	return &BaseComponentInstance{
		key:     key,
		fn:      fn,
		props:   make(Props),
		context: NewComponentContext(key),
		dirty:   true,
	}
}

// NewBaseComponentInstanceWithProps creates a new base instance with props
func NewBaseComponentInstanceWithProps(key string, fn ComponentFuncWithProps, props Props) *BaseComponentInstance {
	return &BaseComponentInstance{
		key:         key,
		fnWithProps: fn,
		props:       props,
		context:     NewComponentContext(key),
		dirty:       true,
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

// Init implements ComponentInstance
func (b *BaseComponentInstance) Init(props Props) {
	b.props = props
	b.dirty = true
}

// Destroy implements ComponentInstance
func (b *BaseComponentInstance) Destroy() {
	b.context.CleanupAll()
	b.mounted = false
}

// OnMount implements ComponentInstance
func (b *BaseComponentInstance) OnMount() {
	b.mounted = true
}

// OnUnmount implements ComponentInstance
func (b *BaseComponentInstance) OnUnmount() {
	b.mounted = false
}

// SetProps implements ComponentInstance
func (b *BaseComponentInstance) SetProps(props Props) bool {
	if propsEqual(b.props, props) {
		return false
	}
	b.props = props
	b.dirty = true
	return true
}

// GetProps implements ComponentInstance
func (b *BaseComponentInstance) GetProps() Props {
	return b.props
}

// MarkDirty implements ComponentInstance
func (b *BaseComponentInstance) MarkDirty() {
	b.dirty = true
}

// IsDirty implements ComponentInstance
func (b *BaseComponentInstance) IsDirty() bool {
	return b.dirty
}

// GetContext implements ComponentInstance
func (b *BaseComponentInstance) GetContext() *ComponentContext {
	return b.context
}

// ClearDirty clears the dirty flag
func (b *BaseComponentInstance) ClearDirty() {
	b.dirty = false
}

// IsMounted returns whether the component is mounted
func (b *BaseComponentInstance) IsMounted() bool {
	return b.mounted
}

// Render calls the component function
func (b *BaseComponentInstance) Render() VNode {
	b.context.ResetContext()
	SetCurrentContext(b.context)

	var vnode VNode
	if b.fn != nil {
		vnode = b.fn()
	} else if b.fnWithProps != nil {
		vnode = b.fnWithProps(b.props)
	}

	SetCurrentContext(nil)
	return vnode
}

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

// =============================================================================
// PaintableInstance - Instance with Paint capability
// =============================================================================

// PaintableInstance is implemented by native components that can paint themselves.
type PaintableInstance interface {
	ComponentInstance
	Paint(x, y int) []paint.DrawCmd
}

// =============================================================================
// FocusableInstance - Instance with Focus capability
// =============================================================================

// FocusableInstance is implemented by components that can receive focus.
type FocusableInstance interface {
	ComponentInstance
	SetFocus(focused bool)
	HasFocus() bool
	IsDisabled() bool
}

// =============================================================================
// ActionHandlerInstance - Instance that handles actions
// =============================================================================

// ActionHandlerInstance is implemented by components that handle actions.
// Fiber-first architecture: components receive concrete Action objects.
type ActionHandlerInstance interface {
	ComponentInstance
	HandleAction(action *action.Action) bool
}

// =============================================================================
// InstanceFactory - Creates instances from VNodes
// =============================================================================

// InstanceFactory is implemented by VNode types that can create instances.
type InstanceFactory interface {
	CreateInstance() ComponentInstance
}

// =============================================================================
// Helper Functions
// =============================================================================

// AsPaintableInstance attempts to cast to PaintableInstance
func AsPaintableInstance(inst ComponentInstance) (PaintableInstance, bool) {
	p, ok := inst.(PaintableInstance)
	return p, ok
}

// AsFocusableInstance attempts to cast to FocusableInstance
func AsFocusableInstance(inst ComponentInstance) (FocusableInstance, bool) {
	f, ok := inst.(FocusableInstance)
	return f, ok
}

// AsActionHandler attempts to cast to ActionHandlerInstance
func AsActionHandler(inst ComponentInstance) (ActionHandlerInstance, bool) {
	a, ok := inst.(ActionHandlerInstance)
	return a, ok
}

// TryInstancePaint attempts to paint an instance
func TryInstancePaint(inst ComponentInstance, x, y int) []paint.DrawCmd {
	if p, ok := AsPaintableInstance(inst); ok {
		return p.Paint(x, y)
	}
	return nil
}

// IsInstanceFactory checks if a VNode implements InstanceFactory
func IsInstanceFactory(vnode VNode) bool {
	_, ok := vnode.(InstanceFactory)
	return ok
}

// =============================================================================
// Simple Paint Registry (for simple components without state)
// =============================================================================

// PaintFunc is a simple paint function for stateless components
type PaintFunc func(props Props, st style.Style, x, y int) []paint.DrawCmd

var paintRegistry = make(map[string]PaintFunc)

// RegisterPaint registers a paint function for a tag
func RegisterPaint(tag string, fn PaintFunc) {
	paintRegistry[tag] = fn
}

// GetPaint returns the paint function for a tag
func GetPaint(tag string) PaintFunc {
	return paintRegistry[tag]
}

// =============================================================================
// State Map (for simple state storage)
// =============================================================================

// StateMap is a simple string-keyed state container
type StateMap map[string]interface{}

// Get retrieves a value from the state map
func (s StateMap) Get(key string) interface{} {
	if s == nil {
		return nil
	}
	return s[key]
}

// Set stores a value in the state map
func (s StateMap) Set(key string, value interface{}) {
	s[key] = value
}

// GetString retrieves a string value
func (s StateMap) GetString(key string) string {
	if v, ok := s[key].(string); ok {
		return v
	}
	return ""
}

// GetInt retrieves an int value
func (s StateMap) GetInt(key string) int {
	switch v := s[key].(type) {
	case int:
		return v
	case int32:
		return int(v)
	case int64:
		return int(v)
	}
	return 0
}

// GetBool retrieves a bool value
func (s StateMap) GetBool(key string) bool {
	if v, ok := s[key].(bool); ok {
		return v
	}
	return false
}

// Clone creates a shallow copy of the state map
func (s StateMap) Clone() StateMap {
	result := make(StateMap, len(s))
	for k, v := range s {
		result[k] = v
	}
	return result
}
