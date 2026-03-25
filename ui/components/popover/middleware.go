package popover

import (
	"github.com/wwsheng009/mint/runtime/action"
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
)

type Middleware struct{}

func NewMiddleware() *Middleware {
	return &Middleware{}
}

func (m *Middleware) Name() string {
	return "PopoverMiddleware"
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
	for _, inst := range popoverRegistryGlobal.openInstances() {
		if inst != nil && inst.requestClose(inst.trigger) {
			return nil
		}
	}
	return act
}

func (m *Middleware) handleClickOutside(act *action.Action) *action.Action {
	mouseMsg, ok := popoverMousePayload(act.Payload)
	if !ok || mouseMsg == nil || mouseMsg.Action != runtimemsg.MouseActionPress {
		return act
	}

	popovers := popoverRegistryGlobal.openInstances()
	if len(popovers) == 0 {
		return act
	}
	if clickHitsOpenPopover(mouseMsg, popovers) {
		return act
	}

	for _, inst := range popovers {
		if inst != nil {
			inst.requestClose(inst.trigger)
		}
	}
	return act
}

func clickHitsOpenPopover(mouseMsg *runtimemsg.MouseMsg, popovers []*Instance) bool {
	if mouseMsg == nil {
		return false
	}
	for _, inst := range popovers {
		if inst == nil {
			continue
		}
		if inst.containsAnchorPoint(mouseMsg.X, mouseMsg.Y) || inst.containsOverlayPoint(mouseMsg.X, mouseMsg.Y) {
			return true
		}
	}
	return false
}

func popoverMousePayload(payload any) (*runtimemsg.MouseMsg, bool) {
	switch value := payload.(type) {
	case *runtimemsg.MouseMsg:
		if value != nil {
			return value, true
		}
	case runtimemsg.MouseMsg:
		copy := value
		return &copy, true
	}
	return nil, false
}
