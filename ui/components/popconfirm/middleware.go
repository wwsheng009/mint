package popconfirm

import (
	"github.com/wwsheng009/mint/runtime/action"
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
)

type Middleware struct{}

func NewMiddleware() *Middleware {
	return &Middleware{}
}

func (m *Middleware) Name() string {
	return "PopconfirmMiddleware"
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
	for _, inst := range popconfirmRegistryGlobal.openInstances() {
		if inst != nil && inst.requestClose(inst.trigger) {
			return nil
		}
	}
	return act
}

func (m *Middleware) handleClickOutside(act *action.Action) *action.Action {
	mouseMsg, ok := popconfirmMousePayload(act.Payload)
	if !ok || mouseMsg == nil || mouseMsg.Action != runtimemsg.MouseActionPress {
		return act
	}

	popconfirms := popconfirmRegistryGlobal.openInstances()
	if len(popconfirms) == 0 {
		return act
	}
	if clickHitsOpenPopconfirm(mouseMsg, popconfirms) {
		return act
	}

	closed := false
	for _, inst := range popconfirms {
		if inst != nil {
			closed = inst.requestClose(inst.trigger) || closed
		}
	}
	if closed {
		return act
	}
	return act
}

func clickHitsOpenPopconfirm(mouseMsg *runtimemsg.MouseMsg, popconfirms []*Instance) bool {
	if mouseMsg == nil {
		return false
	}
	for _, inst := range popconfirms {
		if inst == nil {
			continue
		}
		if inst.containsAnchorPoint(mouseMsg.X, mouseMsg.Y) || inst.containsOverlayPoint(mouseMsg.X, mouseMsg.Y) {
			return true
		}
	}
	return false
}

func popconfirmMousePayload(payload any) (*runtimemsg.MouseMsg, bool) {
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
