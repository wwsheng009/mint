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
	if act.Type == action.ActionClick {
		return m.handleClickOutside(act)
	}
	return act
}

func (m *Middleware) After(act *action.Action, result *action.RouterResult) {}

func (m *Middleware) handleClickOutside(act *action.Action) *action.Action {
	mouseMsg, ok := act.Payload.(*runtimemsg.MouseMsg)
	if !ok || mouseMsg == nil || mouseMsg.Action != runtimemsg.MouseActionPress {
		return act
	}

	open := selectOverlayRegistry.openTriggers()
	if len(open) == 0 {
		return act
	}
	if clickHitsOpenSelect(mouseMsg, open) {
		return act
	}

	closed := false
	for _, inst := range open {
		if inst != nil && inst.closeOnOutside {
			closed = inst.closeDropdown() || closed
		}
	}
	if closed {
		return nil
	}
	return act
}

func clickHitsOpenSelect(mouseMsg *runtimemsg.MouseMsg, selects []*Instance) bool {
	if mouseMsg == nil || len(selects) == 0 {
		return false
	}

	openOwners := make(map[string]struct{}, len(selects))
	for _, inst := range selects {
		if inst == nil {
			continue
		}
		openOwners[inst.ownerID] = struct{}{}
		if inst.containsPoint(mouseMsg.X, mouseMsg.Y) {
			return true
		}
		if popup := selectOverlayRegistry.popup(inst.ownerID); popup != nil && popup.containsPoint(mouseMsg.X, mouseMsg.Y) {
			return true
		}
	}

	if fiber := selectTargetFiber(mouseMsg); fiber != nil && fiberBelongsToOpenSelect(fiber, openOwners) {
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

func fiberBelongsToOpenSelect(fiber *rtui.Fiber, openOwners map[string]struct{}) bool {
	for node := fiber; node != nil; node = node.Return {
		switch inst := node.Instance.(type) {
		case *Instance:
			if _, ok := openOwners[inst.ownerID]; ok {
				return true
			}
		case *popupInstance:
			if _, ok := openOwners[inst.ownerID]; ok {
				return true
			}
		}
	}
	return false
}
