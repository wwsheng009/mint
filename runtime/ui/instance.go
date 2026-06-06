package ui

import (
	"reflect"

	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	"time"
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

// ===== Instance Tree (Mint Runtime 2.0 - Phase 1) =====
// 注意：这不是 ComponentInstance 的一部分
// 只有需要管理子树的组件才实现这些接口

// TreeNode is an optional interface for components that can be in the instance tree
// Most components are tree nodes, but pure functional components might not be
// Parent() returns interface{} to be compatible with intent.TreeComponent and avoid import cycles
type TreeNode interface {
	ComponentInstance
	// Parent returns the parent component instance, or nil if this is root
	Parent() interface{}
	// Children returns the list of child component instances
	Children() []ComponentInstance
}

// TreeContainer is an optional interface for components that manage their child tree
// Components that have children should implement this
type TreeContainer interface {
	TreeNode
	// AddChild adds a child component instance
	AddChild(child ComponentInstance)
	// RemoveChild removes a child component instance
	RemoveChild(child ComponentInstance)
	// ClearChildren removes all child instances
	ClearChildren()
}

// =============================================================================
// BaseComponentInstance - Base implementation
// =============================================================================

// BaseComponentInstance provides a base implementation of ComponentInstance
// It also implements TreeNode and TreeContainer for components that need instance tree support
type BaseComponentInstance struct {
	key         string
	props       Props
	context     *ComponentContext
	fn          ComponentFunc
	fnWithProps ComponentFuncWithProps
	dirty       bool
	mounted     bool

	// ===== Instance Tree (Mint Runtime 2.0 - Phase 1) =====
	// These fields enable TreeNode and TreeContainer interfaces
	// Only relevant for components that manage children
	parent   ComponentInstance
	children []ComponentInstance
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
	b.context.CleanupAll()
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

// ===== Instance Tree Methods (Mint Runtime 2.0 - Phase 1) =====

// Parent implements TreeNode/TreeComponent interface.
// Returns interface{} to be compatible with both interfaces.
func (b *BaseComponentInstance) Parent() interface{} {
	return b.parent
}

// Children implements ComponentInstance
func (b *BaseComponentInstance) Children() []ComponentInstance {
	return b.children
}

// SetParent sets the parent reference (for use by TreeContainer implementations)
// This method is needed by components (like Form) to set parent references on child instances
// that are BaseComponentInstance types but are in different packages.
// Phase 2 fix: Added for Intent Bubble support (INTENT_BUBBLE_AUDIT_REPORT.md P0-2)
func (b *BaseComponentInstance) SetParent(parent ComponentInstance) {
	b.parent = parent
}

// AddChild adds a child component instance
// This is called by the fiber reconciler during mount/updates
func (b *BaseComponentInstance) AddChild(child ComponentInstance) {
	if child == nil {
		return
	}

	// Check if already in children list
	for _, existing := range b.children {
		if existing == child {
			return // Already added
		}
	}

	// Check for circular dependency (prevent adding parent as grandchild)
	// If child would create a cycle, reject it
	if wouldCauseCycle(b, child) {
		return
	}

	// Remove child from its current parent (if any) for re-parenting
	// Skip if child is already the current node's child in another tree
	if child != b {
		if treeNode, ok := child.(TreeNode); ok {
			if oldParent := treeNode.Parent(); oldParent != nil && oldParent != b {
				if oldParentContainer, ok := oldParent.(TreeContainer); ok {
					oldParentContainer.RemoveChild(child)
				}
			}
		}
	}

	// Add to children list
	b.children = append(b.children, child)

	setInstanceParent(child, b)
}

// wouldCauseCycle checks if adding newChild as a child of parent would create a cycle
func wouldCauseCycle(parent, newChild ComponentInstance) bool {
	// Check 1: newChild is already an ancestor of parent?
	// This would create: ... -> newChild -> ... -> parent -> newChild (cycle)
	current := newChild
	visited := make(map[ComponentInstance]bool)

	for current != nil {
		if visited[current] {
			// Cycle detected without adding newChild, this shouldn't happen
			return true
		}
		visited[current] = true

		if current == parent {
			// newChild is already an ancestor of parent, would create cycle
			return true
		}

		// Move up to parent
		if treeNode, ok := current.(TreeNode); ok {
			if parentPtr := treeNode.Parent(); parentPtr != nil {
				current = parentPtr.(ComponentInstance)
			} else {
				break
			}
		} else if current == parent {
			// Direct circular reference
			return true
		} else {
			break
		}
	}

	// Check 2: parent is already an ancestor of newChild?
	// This would create: parent -> ... -> newChild, adding newChild -> parent creates cycle
	current = parent
	visited = make(map[ComponentInstance]bool)

	for current != nil {
		if visited[current] {
			// Cycle detected without adding newChild, this shouldn't happen
			return true
		}
		visited[current] = true

		if current == newChild {
			// parent is already an ancestor of newChild, would create cycle
			return true
		}

		// Move up to parent
		if treeNode, ok := current.(TreeNode); ok {
			if parentPtr := treeNode.Parent(); parentPtr != nil {
				current = parentPtr.(ComponentInstance)
			} else {
				break
			}
		} else if current == newChild {
			// Direct circular reference
			return true
		} else {
			break
		}
	}

	return false
}

// RemoveChild removes a child component instance
// This is called by the fiber reconciler during unmounts
func (b *BaseComponentInstance) RemoveChild(child ComponentInstance) {
	if child == nil {
		return
	}

	// Find and remove child
	for i, existing := range b.children {
		if existing == child {
			b.children = append(b.children[:i], b.children[i+1:]...)
			setInstanceParent(child, nil)
			break
		}
	}
}

// ClearChildren removes all child instances
// Used during subtree unmounts
func (b *BaseComponentInstance) ClearChildren() {
	// Clear parent references on all children
	for _, child := range b.children {
		setInstanceParent(child, nil)
	}
	b.children = b.children[:0]
}

func setInstanceParent(child ComponentInstance, parent ComponentInstance) {
	if child == nil {
		return
	}
	if setter, ok := child.(interface{ SetParent(ComponentInstance) }); ok {
		setter.SetParent(parent)
		return
	}
	if childBase, ok := child.(*BaseComponentInstance); ok {
		childBase.parent = parent
	}
}

// Render calls the component function
func (b *BaseComponentInstance) Render() VNode {
	b.context.SetOwnerInstance(b)
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
		if !reflect.DeepEqual(b[k], v) {
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

// PostPaintableInstance is implemented by native components that need to draw
// lightweight overlays after their child subtree has painted. It is intended for
// visual affordances such as scroll indicators; it must not affect layout,
// focus, hit testing, or event routing.
type PostPaintableInstance interface {
	ComponentInstance
	PostPaint(x, y int) []paint.DrawCmd
}

// ScenePaintableInstance is an optional extension for native instances that can
// contribute raster image layers after the regular text paint has completed.
//
// Implementations should return layers in absolute cell coordinates based on
// the most recent layout bounds pushed through SetBounds().
type ScenePaintableInstance interface {
	ComponentInstance
	SceneLayers() []paint.ImageLayer
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

// RuntimeChildrenProvider is implemented by composite instances that need to
// synthesize additional child VNodes from runtime state.
// This is primarily used by popup/overlay style components where the child tree
// depends on persistent instance state rather than pure VNode props.
type RuntimeChildrenProvider interface {
	ComponentInstance
	RuntimeChildren() []VNode
}

// TickableInstance is implemented by components that need time-driven updates
// such as blinking cursors or lightweight animations.
type TickableInstance interface {
	ComponentInstance
	WantsTick() bool
	Tick(now time.Time) bool
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

// AsPostPaintableInstance attempts to cast to PostPaintableInstance.
func AsPostPaintableInstance(inst ComponentInstance) (PostPaintableInstance, bool) {
	p, ok := inst.(PostPaintableInstance)
	return p, ok
}

// AsScenePaintableInstance attempts to cast to ScenePaintableInstance.
func AsScenePaintableInstance(inst ComponentInstance) (ScenePaintableInstance, bool) {
	s, ok := inst.(ScenePaintableInstance)
	return s, ok
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

// AsTickableInstance attempts to cast to TickableInstance.
func AsTickableInstance(inst ComponentInstance) (TickableInstance, bool) {
	t, ok := inst.(TickableInstance)
	return t, ok
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
