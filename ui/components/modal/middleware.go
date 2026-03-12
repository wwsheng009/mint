package modal

import (
	"github.com/wwsheng009/mint/runtime/action"
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
)

// ModalMiddleware handles modal interactions including:
// - Closing modal on ESC key (ActionCancel, ActionQuit)
// - Closing modal on click outside (ActionClick)
//
// This middleware is registered with the action system and intercepts
// actions before they reach target components.
type ModalMiddleware struct{}

// NewModalMiddleware creates a new modal middleware instance.
func NewModalMiddleware() *ModalMiddleware {
	return &ModalMiddleware{}
}

// Name returns the middleware name for debugging and tracking.
func (m *ModalMiddleware) Name() string {
	return "ModalMiddleware"
}

// Before is called before action dispatch. This is where we intercept
// actions that should affect open modals.
//
// Returns nil to intercept (stop) the action, or the action (possibly modified)
// to continue with normal dispatch.
func (m *ModalMiddleware) Before(act *action.Action) *action.Action {
	// Handle ESC key (ActionCancel, ActionQuit)
	if act.Type == action.ActionCancel || act.Type == action.ActionQuit {
		return m.handleKeyboardClose(act)
	}

	// Handle mouse click (check for click outside)
	if act.Type == action.ActionClick {
		return m.handleClickOutside(act)
	}

	return act
}

// handleKeyboardClose handles ESC key to close modals.
func (m *ModalMiddleware) handleKeyboardClose(act *action.Action) *action.Action {
	topmost := globalRegistry.getTopmostOpenModal()
	if topmost == nil {
		return act
	}

	if topmost.closeable && topmost.closeOnEsc {
		topmost.requestClose()
	}

	// Modal is open, so ESC should never propagate to the app below it.
	return nil
}

// handleClickOutside handles mouse clicks outside of modals.
func (m *ModalMiddleware) handleClickOutside(act *action.Action) *action.Action {
	// Check if payload is MouseMsg
	mouseMsg, ok := act.Payload.(*runtimemsg.MouseMsg)
	if !ok {
		return act
	}

	// Only handle mouse press (not release or move)
	if mouseMsg.Action != runtimemsg.MouseActionPress {
		return act
	}

	topmost := globalRegistry.getTopmostOpenModal()
	if topmost == nil {
		return act // No open modals, continue normally
	}

	// Clicks inside the topmost modal should continue to normal dispatch.
	if topmost.containsPoint(mouseMsg.X, mouseMsg.Y) {
		return act
	}

	if topmost.closeable && topmost.closeOnBackdrop {
		topmost.requestClose()
	}

	// Even when outside-click close is disabled, modal should still block the background.
	return nil
}

// After is called after action dispatch. Used for cleanup or logging.
func (m *ModalMiddleware) After(act *action.Action, result *action.RouterResult) {
	// Nothing to do after dispatch
}
