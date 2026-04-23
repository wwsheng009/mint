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
