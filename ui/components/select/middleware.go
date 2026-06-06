package selectcomp

import (
	"github.com/wwsheng009/mint/runtime/action"
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

type Middleware struct{}

func NewMiddleware() *Middleware {
	return &Middleware{}
}

func (m *Middleware) Name() string {
	return "SelectMiddleware"
}

func (m *Middleware) Before(act *action.Action) *action.Action {
	if act == nil {
		return nil
	}
	switch act.Type {
	case action.ActionCancel, action.ActionQuit:
		return m.handleCancel(act)
	case action.ActionClick:
		return m.handleClickOutside(act)
	case action.ActionNavigateUp,
		action.ActionNavigateDown,
		action.ActionNavigateHome,
		action.ActionNavigateEnd,
		action.ActionNavigatePageUp,
		action.ActionNavigatePageDown,
		action.ActionScrollUp,
		action.ActionScrollDown,
		action.ActionEnter,
		action.ActionSubmit,
		action.ActionInputChar,
		action.ActionInputText,
		action.ActionBackspace,
		action.ActionDeleteChar,
		action.ActionClear,
		action.ActionCursorUp,
		action.ActionCursorDown,
		action.ActionCursorHome,
		action.ActionCursorEnd:
		return m.handleOpenPopupKeyboard(act)
	}
	return act
}

func (m *Middleware) After(act *action.Action, result *action.RouterResult) {}

func (m *Middleware) handleCancel(act *action.Action) *action.Action {
	for _, popup := range selectPopupRegistry.openPopups() {
		if popup != nil && popup.requestClose() {
			return nil
		}
	}
	return act
}

func (m *Middleware) handleOpenPopupKeyboard(act *action.Action) *action.Action {
	popupAct := popupKeyboardAction(act)
	for _, popup := range selectPopupRegistry.openPopups() {
		if popup == nil {
			continue
		}
		if popup.HandleAction(popupAct) {
			return nil
		}
	}
	return act
}

func popupKeyboardAction(act *action.Action) *action.Action {
	if act == nil {
		return nil
	}
	switch act.Type {
	case action.ActionCursorUp:
		return cloneActionWithType(act, action.ActionNavigateUp)
	case action.ActionCursorDown:
		return cloneActionWithType(act, action.ActionNavigateDown)
	case action.ActionCursorHome:
		return cloneActionWithType(act, action.ActionNavigateHome)
	case action.ActionCursorEnd:
		return cloneActionWithType(act, action.ActionNavigateEnd)
	default:
		return act
	}
}

func cloneActionWithType(act *action.Action, actionType action.ActionType) *action.Action {
	if act == nil {
		return nil
	}
	clone := act.Clone()
	clone.Type = actionType
	return clone
}

func (m *Middleware) handleClickOutside(act *action.Action) *action.Action {
	mouseMsg, ok := act.Payload.(*runtimemsg.MouseMsg)
	if !ok || mouseMsg == nil || mouseMsg.Action != runtimemsg.MouseActionPress {
		return act
	}

	popups := selectPopupRegistry.openPopups()
	if len(popups) == 0 {
		return act
	}
	if clickHitsOpenSelect(mouseMsg, popups) {
		return act
	}

	closed := false
	for _, popup := range popups {
		if popup != nil && popup.closeOnOutside {
			closed = popup.requestClose() || closed
		}
	}
	if !closed {
		return act
	}

	if fiber := selectTargetFiber(mouseMsg); fiber != nil && fiberTargetsSelect(fiber) {
		return act
	}
	return nil
}

func clickHitsOpenSelect(mouseMsg *runtimemsg.MouseMsg, popups []*popupInstance) bool {
	if mouseMsg == nil || len(popups) == 0 {
		return false
	}

	openIDs := make(map[string]struct{}, len(popups))
	for _, popup := range popups {
		if popup == nil {
			continue
		}
		openIDs[popup.selectID] = struct{}{}
		if popup.containsPoint(mouseMsg.X, mouseMsg.Y) {
			return true
		}
	}

	if fiber := selectTargetFiber(mouseMsg); fiber != nil && fiberBelongsToOpenSelect(fiber, openIDs) {
		return true
	}
	return false
}

func selectTargetFiber(mouseMsg *runtimemsg.MouseMsg) *rtui.Fiber {
	if mouseMsg == nil || mouseMsg.TargetFiber == nil {
		return nil
	}
	fiber, ok := mouseMsg.TargetFiber.(*rtui.Fiber)
	if !ok || fiber == nil {
		return nil
	}
	return fiber
}

func fiberBelongsToOpenSelect(fiber *rtui.Fiber, openIDs map[string]struct{}) bool {
	for node := fiber; node != nil; node = node.Return {
		switch inst := node.Instance.(type) {
		case *Instance:
			if _, ok := openIDs[inst.selectIdentity()]; ok {
				return true
			}
		case *popupInstance:
			if _, ok := openIDs[inst.selectID]; ok {
				return true
			}
		}
	}
	return false
}

func fiberTargetsSelect(fiber *rtui.Fiber) bool {
	for node := fiber; node != nil; node = node.Return {
		switch node.Instance.(type) {
		case *Instance, *popupInstance:
			return true
		}
	}
	return false
}
