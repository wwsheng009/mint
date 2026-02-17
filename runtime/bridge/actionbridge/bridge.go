// Package actionbridge provides the bridge between Fiber tree and Action Router.
// It is the only module that knows about both Fiber and Router.
//
// Design principle (from fiber_action.md):
// - Fiber stores "who I am" (ActionTargetID)
// - ActionBridge generates "how to route" (Action)
// - Router decides "what to do" (dispatch to Target)
// - Component handles "the logic" (HandleAction)
package actionbridge

import (
	"github.com/wwsheng009/mint/framework/action"
	rtuievent "github.com/wwsheng009/mint/runtime/event"
	"github.com/wwsheng009/mint/runtime/ui"
)

// Bridge connects Fiber tree with Action Router.
// It is the only component that knows about both.
type Bridge struct {
	router *action.Router
}

// New creates a new ActionBridge.
func New(router *action.Router) *Bridge {
	return &Bridge{
		router: router,
	}
}

// DispatchFromFiber dispatches an Action based on Fiber's ActionTargetID.
// It traverses the Fiber tree (bubble path) to find a handler.
//
// Flow:
//  1. Start from target Fiber
//  2. Traverse up (bubble) along Fiber.Return
//  3. For each Fiber with ActionTargetID, create Action and dispatch
//  4. Stop if handled
func (b *Bridge) DispatchFromFiber(
	start *ui.Fiber,
	actionType action.ActionType,
	payload interface{},
) bool {
	for f := start; f != nil; f = f.Return {
		if f.ActionTargetID == "" {
			continue
		}

		// Convert ActionTargetID (string) to NodeID (uint64)
		targetID := rtuievent.StringToNodeID(f.ActionTargetID)

		a := action.NewAction(actionType).
			WithTarget(targetID).
			WithPayload(payload)

		result := b.router.Dispatch(a)
		if result != nil && result.Handled {
			return true
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

	result := b.router.Dispatch(a)
	return result != nil && result.Handled
}
