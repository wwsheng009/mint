// Package action provides action dispatching with scope support.
// Each component subtree can have its own Dispatcher for action isolation.
//
// Design principle (from fiber_confict.md):
// - Closure is converted to ActionID at build time
// - Handler is registered to Scope Dispatcher
// - Fiber only stores ActionTargetID, not function references
// - Runtime uses ActionTargetID to dispatch via Dispatcher
package action

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/wwsheng009/mint/runtime/event"
)

// ScopeActionHandler is a function that handles an Action.
type ScopeActionHandler func(*Action) bool

// ScopeDispatcher handles action dispatching within a scope.
// It supports parent chain for bubbling across scope boundaries.
type ScopeDispatcher struct {
	mu sync.RWMutex

	// handlers maps ActionID to handler function
	handlers map[string]ScopeActionHandler

	// parent is the parent dispatcher (for scope bubbling)
	parent *ScopeDispatcher

	// id is a unique identifier for this dispatcher
	id uint64

	// scopeName is a human-readable name for debugging
	scopeName string
}

// scopeDispatcherIDCounter is a global counter for generating unique dispatcher IDs
var scopeDispatcherIDCounter uint64

// NewScopeDispatcher creates a new ScopeDispatcher.
func NewScopeDispatcher(parent *ScopeDispatcher) *ScopeDispatcher {
	return &ScopeDispatcher{
		handlers:  make(map[string]ScopeActionHandler),
		parent:    parent,
		id:        atomic.AddUint64(&scopeDispatcherIDCounter, 1),
		scopeName: "",
	}
}

// NewScopeDispatcherWithName creates a new ScopeDispatcher with a name.
func NewScopeDispatcherWithName(parent *ScopeDispatcher, name string) *ScopeDispatcher {
	return &ScopeDispatcher{
		handlers:  make(map[string]ScopeActionHandler),
		parent:    parent,
		id:        atomic.AddUint64(&scopeDispatcherIDCounter, 1),
		scopeName: name,
	}
}

// Register registers a handler for an ActionID.
// Returns the ActionID that was registered.
func (d *ScopeDispatcher) Register(actionID string, handler ScopeActionHandler) string {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.handlers[actionID] = handler
	return actionID
}

// Unregister removes a handler for an ActionID.
func (d *ScopeDispatcher) Unregister(actionID string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	delete(d.handlers, actionID)
}

// Dispatch dispatches an Action to the appropriate handler.
// If the Action is not handled locally, it bubbles up to the parent dispatcher.
func (d *ScopeDispatcher) Dispatch(act *Action) bool {
	if act == nil {
		return false
	}

	// Get TargetID as string for lookup
	targetIDStr := fmt.Sprintf("%d", act.TargetID)

	d.mu.RLock()
	handler, exists := d.handlers[targetIDStr]
	d.mu.RUnlock()

	if exists && handler != nil {
		if handler(act) {
			return true
		}
	}

	// Bubble to parent
	if d.parent != nil {
		return d.parent.Dispatch(act)
	}

	return false
}

// DispatchByID dispatches an Action by string ID (for convenience).
func (d *ScopeDispatcher) DispatchByID(actionID string, act *Action) bool {
	act = act.WithTarget(event.StringToNodeID(actionID))
	return d.Dispatch(act)
}

// GetParent returns the parent dispatcher.
func (d *ScopeDispatcher) GetParent() *ScopeDispatcher {
	return d.parent
}

// SetParent sets the parent dispatcher.
func (d *ScopeDispatcher) SetParent(parent *ScopeDispatcher) {
	d.parent = parent
}

// ID returns the unique identifier of this dispatcher.
func (d *ScopeDispatcher) ID() uint64 {
	return d.id
}

// ScopeName returns the human-readable name of this dispatcher.
func (d *ScopeDispatcher) ScopeName() string {
	return d.scopeName
}

// HandlerCount returns the number of registered handlers.
func (d *ScopeDispatcher) HandlerCount() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return len(d.handlers)
}

// HasHandler checks if a handler exists for an ActionID.
func (d *ScopeDispatcher) HasHandler(actionID string) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	_, exists := d.handlers[actionID]
	return exists
}

// Clear removes all handlers.
func (d *ScopeDispatcher) Clear() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.handlers = make(map[string]ScopeActionHandler)
}

// ============================================================================
// Global Scope Management
// ============================================================================

// currentScopeDispatcher is a thread-local storage for the current dispatcher
// This allows builders to register handlers without explicit dispatcher reference
var currentScopeDispatcher atomic.Value

// SetCurrentScopeDispatcher sets the current dispatcher for the goroutine.
// This is typically called when entering a component scope.
func SetCurrentScopeDispatcher(d *ScopeDispatcher) {
	currentScopeDispatcher.Store(d)
}

// GetCurrentScopeDispatcher gets the current dispatcher.
// Returns nil if no dispatcher is set.
func GetCurrentScopeDispatcher() *ScopeDispatcher {
	v := currentScopeDispatcher.Load()
	if v == nil {
		return nil
	}
	return v.(*ScopeDispatcher)
}

// WithScopeDispatcher sets the current dispatcher for the duration of a function.
// This is useful for building components with their own scope.
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

// GenerateScopeActionID generates a unique ActionID.
// The format is "action_<counter>" for simplicity.
func GenerateScopeActionID() string {
	id := atomic.AddUint64(&scopeActionIDCounter, 1)
	return fmt.Sprintf("action_%d", id)
}

// GenerateScopeActionIDWithPrefix generates a unique ActionID with a prefix.
// The format is "<prefix>_<counter>".
func GenerateScopeActionIDWithPrefix(prefix string) string {
	id := atomic.AddUint64(&scopeActionIDCounter, 1)
	return fmt.Sprintf("%s_%d", prefix, id)
}

// ============================================================================
// Helper Functions for Closure Registration
// ============================================================================

// RegisterScopeClosure registers a closure as an Action handler.
// It generates a unique ActionID and registers the handler to the current dispatcher.
// Returns the ActionID for use in Fiber.ActionTargetID.
//
// Usage:
//
//	actionID := RegisterScopeClosure(func() {
//	    // handle click
//	})
//	fiber.ActionTargetID = actionID
func RegisterScopeClosure(handler func()) string {
	return RegisterScopeClosureWithAction(func(act *Action) bool {
		handler()
		return true
	})
}

// RegisterScopeClosureWithAction registers a closure that receives the Action.
// This allows the handler to inspect the Action payload.
func RegisterScopeClosureWithAction(handler ScopeActionHandler) string {
	d := GetCurrentScopeDispatcher()
	if d == nil {
		// No dispatcher set, return empty
		// This shouldn't happen in normal usage
		return ""
	}

	actionID := GenerateScopeActionID()
	d.Register(actionID, handler)
	return actionID
}

// RegisterScopeClosureToDispatcher registers a closure to a specific dispatcher.
// Use this when you have an explicit dispatcher reference.
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
