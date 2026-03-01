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
	// Get all open modals from global registry
	modals := globalRegistry.getOpenModals()
	for _, modalInst := range modals {
		if modalInst.isOpen && modalInst.closeable {
			// Close the modal
			modalInst.isOpen = false
			modalInst.dirty = true
			modalInst.emitCloseIntent()
			// Return nil to intercept (stop) the action
			return nil
		}
	}
	return act
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

	// Get all open modals from global registry
	modals := globalRegistry.getOpenModals()
	if len(modals) == 0 {
		return act // No open modals, continue normally
	}

	// Check if the click is inside any open modal
	clickedInsideModal := false
	for _, modalInst := range modals {
		if modalInst.isOpen && modalInst.closeable {
			if modalInst.containsPoint(mouseMsg.X, mouseMsg.Y) {
				clickedInsideModal = true
				break
			}
		}
	}

	// If click is outside all modals, close the topmost one
	if !clickedInsideModal {
		for i := len(modals) - 1; i >= 0; i-- {
			modalInst := modals[i]
			if modalInst.isOpen && modalInst.closeable {
				// Close the modal
				modalInst.isOpen = false
				modalInst.dirty = true
				modalInst.emitCloseIntent()
				// Return nil to intercept (stop) the action
				return nil
			}
		}
	}

	return act
}

// After is called after action dispatch. Used for cleanup or logging.
func (m *ModalMiddleware) After(act *action.Action, result *action.RouterResult) {
	// Nothing to do after dispatch
}
