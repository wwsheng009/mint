// Package actionbridge provides the bridge between Fiber tree and Action Router.
// It is the only module that knows about both Fiber and Router.
//
// Design principle (from fiber_action.md):
// - Fiber stores "who I am" (ActionTargetID)
// - ActionBridge generates "how to route" (Action)
// - Router decides "what to do" (dispatch to Target)
// - Component handles "the logic" (HandleAction)
//
// Architecture boundaries:
// - ❌ App should NOT access fiber.FocusableVNode directly
// - ❌ Dispatcher should NOT know Fiber structure
// - ✅ ActionBridge is the ONLY module that knows both Fiber and Router
package actionbridge

import (
	"fmt"

	"github.com/wwsheng009/mint/framework/action"
	rtuievent "github.com/wwsheng009/mint/runtime/event"
	"github.com/wwsheng009/mint/runtime/ui"
)

// Bridge connects Fiber tree with Action Router.
// It is the only component that knows about both.
type Bridge struct {
	router          *action.Router
	scopeDispatcher *action.ScopeDispatcher
}

// New creates a new ActionBridge.
func New(router *action.Router) *Bridge {
	return &Bridge{
		router: router,
	}
}

// SetScopeDispatcher sets the scope dispatcher for closure-based actions.
func (b *Bridge) SetScopeDispatcher(d *action.ScopeDispatcher) {
	b.scopeDispatcher = d
}

// DispatchFromFiber dispatches an Action based on Fiber.
// It supports three modes:
// 1. ScopeDispatcher mode: Fiber.ActionTargetID → ScopeDispatcher → registered closure
// 2. Semantic Action: Fiber.ActionTargetID → Router → registered handler
// 3. Closure mode (legacy): FocusableVNode.HandleAction → onClick callback
//
// Flow (bubble path):
//  1. Start from target Fiber
//  2. For each Fiber, try modes in order
//  3. Traverse up along Fiber.Return until handled
func (b *Bridge) DispatchFromFiber(
	start *ui.Fiber,
	actionType action.ActionType,
	payload interface{},
) bool {
	for f := start; f != nil; f = f.Return {
		// Mode 1: ScopeDispatcher mode (ActionTargetID → registered closure)
		// This is the new unified mode where closures are converted to ActionIDs
		if f.ActionTargetID != "" && b.scopeDispatcher != nil {
			// Convert ActionTargetID to uint64 for dispatch
			targetID := rtuievent.StringToNodeID(f.ActionTargetID)
			a := action.NewAction(actionType).
				WithTarget(targetID).
				WithPayload(payload)

			if b.scopeDispatcher.Dispatch(a) {
				return true
			}
		}

		// Mode 2: Semantic Action (ActionTargetID → Router)
		if f.ActionTargetID != "" {
			targetID := rtuievent.StringToNodeID(f.ActionTargetID)
			a := action.NewAction(actionType).
				WithTarget(targetID).
				WithPayload(payload)

			result := b.router.Dispatch(a)
			if result != nil && result.Handled {
				return true
			}
		}

		// Mode 3: Closure mode (FocusableVNode implements ActionTarget) - Legacy
		// This will be deprecated once all closures are converted to ScopeDispatcher
		if f.FocusableVNode != nil {
			if target, ok := f.FocusableVNode.(action.ActionTarget); ok {
				a := action.NewAction(actionType).WithPayload(payload)
				if target.HandleAction(a) {
					return true
				}
			}
		}
	}
	return false
}

// DispatchToTarget dispatches an Action directly to a specific target ID.
func (b *Bridge) DispatchToTarget(
	targetID string,
	actionType action.ActionType,
	payload interface{},
) bool {
	if targetID == "" {
		return false
	}

	nodeID := rtuievent.StringToNodeID(targetID)

	a := action.NewAction(actionType).
		WithTarget(nodeID).
		WithPayload(payload)

	// Try ScopeDispatcher first
	if b.scopeDispatcher != nil {
		if b.scopeDispatcher.Dispatch(a) {
			return true
		}
	}

	// Fallback to Router
	result := b.router.Dispatch(a)
	return result != nil && result.Handled
}

// DispatchToScopeTarget dispatches an Action to a scope-registered target.
// This is used when the target ID is already registered with ScopeDispatcher.
func (b *Bridge) DispatchToScopeTarget(
	actionID string,
	actionType action.ActionType,
	payload interface{},
) bool {
	if actionID == "" || b.scopeDispatcher == nil {
		return false
	}

	a := action.NewAction(actionType).WithPayload(payload)
	return b.scopeDispatcher.DispatchByID(actionID, a)
}

// String returns a string representation of the bridge state.
func (b *Bridge) String() string {
	var scopeInfo string
	if b.scopeDispatcher != nil {
		scopeInfo = fmt.Sprintf(", scopeDispatcher=%s", b.scopeDispatcher.ScopeName())
	}
	return fmt.Sprintf("Bridge{router=%p%s}", b.router, scopeInfo)
}
