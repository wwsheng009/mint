// Package action provides action dispatching with scope support.
// Each component subtree can have its own Dispatcher for action isolation.
//
// Design principle:
// - Closure is converted to ActionID at build time
// - Handler is registered to Scope Dispatcher
// - Component stores only ActionTargetID, not function references
// - Runtime uses ActionTargetID to dispatch via Dispatcher
package action

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
)

// ScopeActionHandler is a function that handles an Action within a scope
type ScopeActionHandler func(*Action) bool

// ScopeDispatcher handles action dispatching within a scope
// Supports parent chain for bubbling across scope boundaries
type ScopeDispatcher struct {
	mu sync.RWMutex

	// handlers maps ActionID to handler function
	// Supports both string IDs and uint64 IDs
	handlers     map[string]ScopeActionHandler
	handlersU64  map[uint64]ScopeActionHandler

	// parent is the parent dispatcher (for scope bubbling)
	parent *ScopeDispatcher

	// id is a unique identifier for this dispatcher
	id uint64

	// scopeName is a human-readable name for debugging
	scopeName string
}

// scopeDispatcherIDCounter is a global counter for generating unique dispatcher IDs
var scopeDispatcherIDCounter uint64

// NewScopeDispatcher creates a new ScopeDispatcher
func NewScopeDispatcher(parent *ScopeDispatcher) *ScopeDispatcher {
	return &ScopeDispatcher{
		handlers:    make(map[string]ScopeActionHandler),
		handlersU64: make(map[uint64]ScopeActionHandler),
		parent:      parent,
		id:          atomic.AddUint64(&scopeDispatcherIDCounter, 1),
		scopeName:   "",
	}
}

// NewScopeDispatcherWithName creates a new ScopeDispatcher with a name
func NewScopeDispatcherWithName(parent *ScopeDispatcher, name string) *ScopeDispatcher {
	return &ScopeDispatcher{
		handlers:    make(map[string]ScopeActionHandler),
		handlersU64: make(map[uint64]ScopeActionHandler),
		parent:      parent,
		id:          atomic.AddUint64(&scopeDispatcherIDCounter, 1),
		scopeName:   name,
	}
}

// Register registers a handler for an ActionID
// Returns the ActionID that was registered
func (d *ScopeDispatcher) Register(actionID string, handler ScopeActionHandler) string {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.handlers[actionID] = handler

	// Also register by uint64 ID for performance
	u64ID := StringToTargetID(actionID)
	d.handlersU64[u64ID] = handler

	return actionID
}

// RegisterU64 registers a handler for a uint64 ActionID directly
func (d *ScopeDispatcher) RegisterU64(actionID uint64, handler ScopeActionHandler) uint64 {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.handlersU64[actionID] = handler

	// Also register by string representation
	strID := fmt.Sprintf("%d", actionID)
	d.handlers[strID] = handler

	return actionID
}

// Unregister removes a handler for an ActionID
func (d *ScopeDispatcher) Unregister(actionID string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Remove from both maps
	u64ID := StringToTargetID(actionID)
	delete(d.handlers, actionID)
	delete(d.handlersU64, u64ID)
}

// UnregisterU64 removes a handler by uint64 ID
func (d *ScopeDispatcher) UnregisterU64(actionID uint64) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Remove from both maps
	strID := fmt.Sprintf("%d", actionID)
	delete(d.handlers, strID)
	delete(d.handlersU64, actionID)
}

// Dispatch dispatches an Action to the appropriate handler
// If the Action is not handled locally, it bubbles up to the parent dispatcher
func (d *ScopeDispatcher) Dispatch(act *Action) bool {
	if act == nil {
		return false
	}

	// Try string ID first
	if act.Target != "" {
		d.mu.RLock()
		handler, exists := d.handlers[act.Target]
		d.mu.RUnlock()

		if exists && handler != nil {
			if handler(act) {
				return true
			}
		}
	}

	// Try uint64 ID
	if act.TargetID != 0 {
		d.mu.RLock()
		handler, exists := d.handlersU64[act.TargetID]
		d.mu.RUnlock()

		if exists && handler != nil {
			if handler(act) {
				return true
			}
		}
	}

	// Bubble to parent
	if d.parent != nil {
		return d.parent.Dispatch(act)
	}

	return false
}

// DispatchByID dispatches an Action by string ID
func (d *ScopeDispatcher) DispatchByID(actionID string, act *Action) bool {
	act = act.WithTarget(actionID)
	return d.Dispatch(act)
}

// DispatchByIDU64 dispatches an Action by uint64 ID
func (d *ScopeDispatcher) DispatchByIDU64(actionID uint64, act *Action) bool {
	act = act.WithTargetID(actionID)
	return d.Dispatch(act)
}

// GetParent returns the parent dispatcher
func (d *ScopeDispatcher) GetParent() *ScopeDispatcher {
	return d.parent
}

// SetParent sets the parent dispatcher
func (d *ScopeDispatcher) SetParent(parent *ScopeDispatcher) {
	d.parent = parent
}

// ID returns the unique identifier of this dispatcher
func (d *ScopeDispatcher) ID() uint64 {
	return d.id
}

// ScopeName returns the human-readable name of this dispatcher
func (d *ScopeDispatcher) ScopeName() string {
	return d.scopeName
}

// SetScopeName sets the scope name
func (d *ScopeDispatcher) SetScopeName(name string) {
	d.scopeName = name
}

// HandlerCount returns the number of registered handlers
func (d *ScopeDispatcher) HandlerCount() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return len(d.handlers)
}

// HasHandler checks if a handler exists for an ActionID
func (d *ScopeDispatcher) HasHandler(actionID string) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	_, exists := d.handlers[actionID]
	return exists
}

// HasHandlerU64 checks if a handler exists for a uint64 ActionID
func (d *ScopeDispatcher) HasHandlerU64(actionID uint64) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	_, exists := d.handlersU64[actionID]
	return exists
}

// Clear removes all handlers
func (d *ScopeDispatcher) Clear() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.handlers = make(map[string]ScopeActionHandler)
	d.handlersU64 = make(map[uint64]ScopeActionHandler)
}

// String returns string representation
func (d *ScopeDispatcher) String() string {
	parentName := "nil"
	if d.parent != nil {
		parentName = d.parent.ScopeName()
	}
	return fmt.Sprintf("ScopeDispatcher{id=%d, name=%s, handlers=%d, parent=%s}",
		d.id, d.scopeName, d.HandlerCount(), parentName)
}

// ============================================================================
// Global Scope Management
// ============================================================================

// currentScopeDispatcher is a thread-local storage for the current dispatcher
// This allows builders to register handlers without explicit dispatcher reference
var currentScopeDispatcher atomic.Value

// SetCurrentScopeDispatcher sets the current dispatcher for the goroutine
// This is typically called when entering a component scope
func SetCurrentScopeDispatcher(d *ScopeDispatcher) {
	currentScopeDispatcher.Store(d)
}

// GetCurrentScopeDispatcher gets the current dispatcher
// Returns nil if no dispatcher is set
func GetCurrentScopeDispatcher() *ScopeDispatcher {
	v := currentScopeDispatcher.Load()
	if v == nil {
		return nil
	}
	return v.(*ScopeDispatcher)
}

// WithScopeDispatcher sets the current dispatcher for the duration of a function
// This is useful for building components with their own scope
func WithScopeDispatcher(d *ScopeDispatcher, fn func()) {
	old := GetCurrentScopeDispatcher()
	SetCurrentScopeDispatcher(d)
	defer SetCurrentScopeDispatcher(old)
	fn()
}

// ============================================================================
// ActionID Generation
// ============================================================================

// scopeActionIDCounter is a global counter for generating unique ActionIDs
var scopeActionIDCounter uint64

// GenerateScopeActionID generates a unique ActionID
// The format is "action_<counter>" for simplicity
func GenerateScopeActionID() string {
	id := atomic.AddUint64(&scopeActionIDCounter, 1)
	return fmt.Sprintf("action_%d", id)
}

// GenerateScopeActionIDWithPrefix generates a unique ActionID with a prefix
// The format is "<prefix>_<counter>"
func GenerateScopeActionIDWithPrefix(prefix string) string {
	id := atomic.AddUint64(&scopeActionIDCounter, 1)
	return fmt.Sprintf("%s_%d", prefix, id)
}

// ============================================================================
// Helper Functions for Closure Registration
// ============================================================================

// RegisterScopeClosure registers a closure as an Action handler
// It generates a unique ActionID and registers the handler to the current dispatcher
// Returns the ActionID for use in component ActionTargetID
//
// Usage:
//
//	actionID := RegisterScopeClosure(func() {
//	    // handle click
//	})
//	component.ActionTargetID = actionID
func RegisterScopeClosure(handler func()) string {
	return RegisterScopeClosureWithAction(func(act *Action) bool {
		handler()
		return true
	})
}

// RegisterScopeClosureWithAction registers a closure that receives the Action
// This allows the handler to inspect the Action payload
func RegisterScopeClosureWithAction(handler ScopeActionHandler) string {
	d := GetCurrentScopeDispatcher()
	if d == nil {
		return "" // No dispatcher set
	}

	actionID := GenerateScopeActionID()
	d.Register(actionID, handler)
	return actionID
}

// RegisterScopeClosureToDispatcher registers a closure to a specific dispatcher
// Use this when you have an explicit dispatcher reference
func RegisterScopeClosureToDispatcher(d *ScopeDispatcher, handler func()) string {
	if d == nil {
		return ""
	}

	actionID := GenerateScopeActionID()
	d.Register(actionID, func(act *Action) bool {
		handler()
		return true
	})
	return actionID
}

// RegisterTargetToDispatcher registers a Target to a specific dispatcher
// This is a convenience wrapper for registering components
func RegisterTargetToDispatcher(d *ScopeDispatcher, target Target) string {
	if d == nil || target == nil {
		return ""
	}

	actionID := target.ID()
	d.Register(actionID, target.HandleAction)
	return actionID
}

// CreateChildScope creates a child scope dispatcher
func (d *ScopeDispatcher) CreateChildScope(name string) *ScopeDispatcher {
	return NewScopeDispatcherWithName(d, name)
}

// DispatchBubble dispatches action with automatic bubbling to parent
// This is a convenience method that always tries bubbling
func (d *ScopeDispatcher) DispatchBubble(act *Action) bool {
	// Try local first
	if d.Dispatch(act) {
		return true
	}

	// Force bubble to parent
	if d.parent != nil {
		return d.parent.DispatchBubble(act)
	}

	return false
}

// WalkUp walks up the dispatcher chain and calls fn for each
// Stops if fn returns true
func (d *ScopeDispatcher) WalkUp(fn func(dispatcher *ScopeDispatcher) bool) {
	current := d
	for current != nil {
		if fn(current) {
			return
		}
		current = current.parent
	}
}

// GetDepth returns the depth of this dispatcher in the chain (0 = root)
func (d *ScopeDispatcher) GetDepth() int {
	depth := 0
	current := d.parent
	for current != nil {
		depth++
		current = current.parent
	}
	return depth
}

// IsRoot returns true if this dispatcher has no parent
func (d *ScopeDispatcher) IsRoot() bool {
	return d.parent == nil
}

// DumpTree dumps the entire dispatcher tree hierarchy
func (d *ScopeDispatcher) DumpTree() string {
	var sb strings.Builder
	d.dumpTree(&sb, 0)
	return sb.String()
}

func (d *ScopeDispatcher) dumpTree(sb *strings.Builder, depth int) {
	indent := strings.Repeat("  ", depth)
	sb.WriteString(fmt.Sprintf("%sScopeDispatcher{id=%d, name=%s, handlers=%d}\n",
		indent, d.id, d.scopeName, d.HandlerCount()))

	// Note: We can't easily traverse children since they only have parent refs
	// This is a limitation - we'd need to maintain a child list if we want full tree traversal
}
