package action

import "fmt"

// ============================================================================
// Core Target Interface (Simplified from runtime/action)
// ============================================================================

// Target is the interface that components implement to receive actions.
// Implements only two methods for simplicity.
type Target interface {
	// ID returns the component's unique identifier
	ID() string

	// HandleAction processes an action and returns true if handled
	HandleAction(a *Action) bool
}

// ============================================================================
// Capability Interfaces (From framework/action)
// ============================================================================
// Components can implement these optional capability interfaces to declare
// semantic abilities beyond basic action handling.

// Focusable indicates the component can receive and manage focus
type Focusable interface {
	// Focus sets focus to this component, returns true if successful
	Focus() bool

	// Blur removes focus from this component
	Blur()

	// IsFocused checks if this component currently has focus
	IsFocused() bool

	// IsFocusable checks if this component can receive focus
	IsFocusable() bool
}

// Scrollable indicates the component supports scrolling
type Scrollable interface {
	// CanScroll checks if scrolling is possible in the specified direction
	// delta > 0 means scroll up, delta < 0 means scroll down
	CanScroll(delta int) bool

	// Scroll performs scrolling, returns true if successful
	Scroll(delta int) bool

	// GetScrollPosition returns current, total, and visible range
	GetScrollPosition() (current, total, visible int)
}

// Editable indicates the component supports text editing
type Editable interface {
	// InsertText inserts text at cursor position
	InsertText(text string) bool

	// DeleteText deletes text in direction (-1=backspace, 1=delete)
	DeleteText(direction int) bool

	// GetText returns the current text content
	GetText() string

	// SetCursorPosition moves the cursor to position
	SetCursorPosition(pos int) bool

	// GetCursorPosition returns the current cursor position
	GetCursorPosition() int
}

// Selectable indicates the component supports selection
type Selectable interface {
	// Select selects the current item
	Select() bool

	// ToggleSelection toggles selection state
	ToggleSelection() bool

	// IsSelected checks if current item is selected
	IsSelected() bool

	// GetSelectedCount returns the number of selected items
	GetSelectedCount() int
}

// Expandable indicates the component supports expand/collapse
type Expandable interface {
	// Expand expands the component
	Expand() bool

	// Collapse collapses the component
	Collapse() bool

	// Toggle toggles expand/collapse state
	Toggle() bool

	// IsExpanded checks if component is expanded
	IsExpanded() bool
}

// Draggable indicates the component supports dragging
type Draggable interface {
	// StartDrag begins drag operation
	StartDrag(act *Action) bool

	// Drag handles drag movement
	Drag(act *Action) bool

	// EndDrag ends drag operation
	EndDrag(act *Action) bool

	// IsDragging checks if currently dragging
	IsDragging() bool
}

// ============================================================================
// Composite Interfaces
// ============================================================================
// These combine Target with specific capabilities for type safety.

// FocusableTarget combines Target with Focusable
type FocusableTarget interface {
	Target
	Focusable
}

// ScrollableTarget combines Target with Scrollable
type ScrollableTarget interface {
	Target
	Scrollable
}

// EditableTarget combines Target with Editable
type EditableTarget interface {
	Target
	Editable
}

// SelectableTarget combines Target with Selectable
type SelectableTarget interface {
	Target
	Selectable
}

// ExpandableTarget combines Target with Expandable
type ExpandableTarget interface {
	Target
	Expandable
}

// ============================================================================
// Functional Adapters
// ============================================================================

// TargetFunc is a functional implementation of Target
type TargetFunc struct {
	id      string
	handler func(*Action) bool
}

// NewTargetFunc creates a functional Target
func NewTargetFunc(id string, handler func(*Action) bool) *TargetFunc {
	return &TargetFunc{
		id:      id,
		handler: handler,
	}
}

// ID returns the target's ID
func (t *TargetFunc) ID() string {
	return t.id
}

// HandleAction calls the handler function
func (t *TargetFunc) HandleAction(a *Action) bool {
	if t.handler != nil {
		return t.handler(a)
	}
	return false
}

// TargetChain implements responsibility chain pattern for targets
type TargetChain struct {
	id      string
	targets []Target
}

// NewTargetChain creates a chain of targets
func NewTargetChain(id string, targets ...Target) *TargetChain {
	return &TargetChain{
		id:      id,
		targets: targets,
	}
}

// ID returns the chain's ID
func (c *TargetChain) ID() string {
	return c.id
}

// HandleAction tries each target in sequence until one handles the action
func (c *TargetChain) HandleAction(a *Action) bool {
	for _, target := range c.targets {
		if target.HandleAction(a) {
			return true
		}
	}
	return false
}

// AddTarget adds a target to the chain
func (c *TargetChain) AddTarget(target Target) {
	c.targets = append(c.targets, target)
}

// NoOpTarget is a no-op target for testing/placeholder
type NoOpTarget struct {
	id string
}

// NewNoOpTarget creates a no-op target
func NewNoOpTarget(id string) *NoOpTarget {
	return &NoOpTarget{id: id}
}

// ID returns the target's ID
func (n *NoOpTarget) ID() string {
	return n.id
}

// HandleAction does nothing, always returns false
func (n *NoOpTarget) HandleAction(a *Action) bool {
	return false
}

// ============================================================================
// Base ActionTarget (From framework/action)
// ============================================================================

// BaseActionTarget provides a base implementation for ActionTarget
type BaseActionTarget struct {
	supportedActions []ActionType
}

// NewBaseActionTarget creates a base action target
func NewBaseActionTarget(supportedActions ...ActionType) *BaseActionTarget {
	return &BaseActionTarget{
		supportedActions: supportedActions,
	}
}

// HandleAction always returns false (placeholder)
func (b *BaseActionTarget) HandleAction(a *Action) bool {
	return false
}

// ID returns empty string (should be overridden)
func (b *BaseActionTarget) ID() string {
	return ""
}

// GetSupportedActions returns supported action types
func (b *BaseActionTarget) GetSupportedActions() []ActionType {
	return b.supportedActions
}

// CanHandleAction checks if action is supported
func (b *BaseActionTarget) CanHandleAction(a *Action) bool {
	for _, supported := range b.supportedActions {
		if supported == a.Type {
			return true
		}
	}
	return false
}

// AddSupportedActions adds more supported actions
func (b *BaseActionTarget) AddSupportedActions(actions ...ActionType) {
	b.supportedActions = append(b.supportedActions, actions...)
}

// ============================================================================
// Composite ActionTarget (From framework/action)
// ============================================================================

// CompositeActionTarget combines multiple targets into one
type CompositeActionTarget struct {
	targets []Target
}

// NewCompositeActionTarget creates a composite target
func NewCompositeActionTarget(targets ...Target) *CompositeActionTarget {
	return &CompositeActionTarget{
		targets: targets,
	}
}

// ID returns the first target's ID or empty
func (c *CompositeActionTarget) ID() string {
	if len(c.targets) > 0 {
		return c.targets[0].ID()
	}
	return ""
}

// HandleAction tries each target until one handles the action
func (c *CompositeActionTarget) HandleAction(a *Action) bool {
	for _, target := range c.targets {
		if target.HandleAction(a) {
			return true
		}
	}
	return false
}

// GetSupportedActions returns union of all targets' supported actions
func (c *CompositeActionTarget) GetSupportedActions() []ActionType {
	if base, ok := c.targets[0].(*BaseActionTarget); ok {
		return base.GetSupportedActions()
	}
	return nil
}

// CanHandleAction checks if any target can handle the action
func (c *CompositeActionTarget) CanHandleAction(a *Action) bool {
	for _, target := range c.targets {
		if base, ok := target.(*BaseActionTarget); ok {
			if base.CanHandleAction(a) {
				return true
			}
		}
	}
	return false
}

// AddTarget adds a target to the composite
func (c *CompositeActionTarget) AddTarget(target Target) {
	c.targets = append(c.targets, target)
}

// String returns debug info
func (c *CompositeActionTarget) String() string {
	return fmt.Sprintf("CompositeActionTarget{%d targets}", len(c.targets))
}

// ============================================================================
// ActionTarget Adapter (From framework/action)
// ============================================================================

// ActionHandlerFunc is a function type for handling actions
type ActionHandlerFunc func(action *Action) bool

// ActionTargetAdapter adapts a function to Target interface
type ActionTargetAdapter struct {
	*BaseActionTarget
	handler ActionHandlerFunc
}

// NewActionTargetAdapter creates an action target adapter
func NewActionTargetAdapter(supportedActions []ActionType, handler ActionHandlerFunc) *ActionTargetAdapter {
	return &ActionTargetAdapter{
		BaseActionTarget: NewBaseActionTarget(supportedActions...),
		handler:          handler,
	}
}

// HandleAction calls the handler function
func (a *ActionTargetAdapter) HandleAction(action *Action) bool {
	if a.handler != nil {
		return a.handler(action)
	}
	return false
}

// ============================================================================
// Utility Functions for Target Management
// ============================================================================

// GetActionTargets extracts all Target instances from a slice of interfaces
func GetActionTargets(nodes []interface{}) []Target {
	targets := make([]Target, 0)
	for _, node := range nodes {
		if target, ok := node.(Target); ok {
			targets = append(targets, target)
		}
	}
	return targets
}

// FilterActionTargets filters targets that support a specific action
func FilterActionTargets(targets []Target, actionType ActionType) []Target {
	filtered := make([]Target, 0)
	for _, target := range targets {
		if base, ok := target.(*BaseActionTarget); ok {
			for _, supported := range base.GetSupportedActions() {
				if supported == actionType {
					filtered = append(filtered, target)
					break
				}
			}
		}
	}
	return filtered
}

// FindActionTarget finds first target that can handle an action
func FindActionTargets(targets []Target, actionType ActionType) Target {
	for _, target := range targets {
		if base, ok := target.(*BaseActionTarget); ok {
			if base.CanHandleAction(NewAction(actionType)) {
				return target
			}
		}
	}
	return nil
}

// DispatchActionToTargets dispatches action to targets in sequence
func DispatchActionToTargets(action *Action, targets ...Target) bool {
	for _, target := range targets {
		if target.HandleAction(action) {
			return true
		}
	}
	return false
}

// DispatchActionToTargetsWithFallback dispatches with fallback handler
func DispatchActionToTargetsWithFallback(action *Action, fallback func(*Action) bool, targets ...Target) bool {
	if DispatchActionToTargets(action, targets...) {
		return true
	}
	if fallback != nil {
		return fallback(action)
	}
	return false
}
