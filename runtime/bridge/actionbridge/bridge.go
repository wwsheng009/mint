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

	"github.com/wwsheng009/mint/internal/log"
	"github.com/wwsheng009/mint/runtime/action"
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
// 3. ActionHandlerInstance mode: Fiber.Instance.HandleAction → Fiber-first components
//
// Flow (bubble path):
//  1. Start from target Fiber
//  2. For each Fiber, try modes in order
//  3. Traverse up along Fiber.Return until handled
//
// Note: payload should be the raw payload value (e.g., string for text input), NOT the Action object.
// The Action object is reconstructed for Mode 1 and Mode 2.
func (b *Bridge) DispatchFromFiber(
	start *ui.Fiber,
	actionType action.ActionType,
	payload interface{},
) bool {
	for f := start; f != nil; f = f.Return {
		// Mode 3: ActionHandlerInstance mode (Fiber-first components)
		// New UI components (ui/components/*) implement rtui.ActionHandlerInstance
		// This bypasses the Action system and directly calls Instance.HandleAction
		if f.Instance != nil {
			if handler, ok := f.Instance.(ui.ActionHandlerInstance); ok {
				// Create Action object and pass to handler
				a := action.NewAction(actionType).
					WithTargetID(f.NodeID). // Use uint64 ID for performance
					WithPayload(payload)
				if handler.HandleAction(a) {
					return true
				}
			}
		} else {
			log.RenderLogger.Debug("[%s] %s ui.ComponentInstance is nil!", start.Tag, start.Key)
		}

		// Mode 1: ScopeDispatcher mode (ActionTargetID → registered closure)
		// This is the new unified mode where closures are converted to ActionIDs
		// Reconstruct Action object with metadata for dispatch
		if f.ActionTargetID != "" && b.scopeDispatcher != nil {
			a := action.NewAction(actionType).
				WithTarget(f.ActionTargetID).
				WithPayload(payload)

			if b.scopeDispatcher.Dispatch(a) {
				return true
			}
		}

		// Mode 2: Semantic Action (ActionTargetID → Router)
		// Reconstruct Action object with metadata for dispatch
		if f.ActionTargetID != "" {
			a := action.NewAction(actionType).
				WithTarget(f.ActionTargetID).
				WithPayload(payload)

			result := b.router.Dispatch(a)
			if result != nil && result.Handled {
				return true
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

	a := action.NewAction(actionType).
		WithTarget(targetID).
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
