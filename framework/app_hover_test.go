package framework

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/action"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

type hoverActionRecorder struct {
	actions []action.ActionType
}

func (m *hoverActionRecorder) Key() string                        { return "" }
func (m *hoverActionRecorder) SetKey(key string)                  {}
func (m *hoverActionRecorder) Init(props rtui.Props)              {}
func (m *hoverActionRecorder) Destroy()                           {}
func (m *hoverActionRecorder) OnMount()                           {}
func (m *hoverActionRecorder) OnUnmount()                         {}
func (m *hoverActionRecorder) SetProps(props rtui.Props) bool     { return false }
func (m *hoverActionRecorder) GetProps() rtui.Props               { return nil }
func (m *hoverActionRecorder) MarkDirty()                         {}
func (m *hoverActionRecorder) IsDirty() bool                      { return false }
func (m *hoverActionRecorder) GetContext() *rtui.ComponentContext { return nil }
func (m *hoverActionRecorder) HandleAction(act *action.Action) bool {
	m.actions = append(m.actions, act.Type)
	switch act.Type {
	case action.ActionMouseEnter, action.ActionMouseLeave:
		return true
	default:
		return false
	}
}

func TestApp_UpdateHoveredFiberDispatchesEnterLeave(t *testing.T) {
	app := NewApp()

	first := &hoverActionRecorder{}
	second := &hoverActionRecorder{}
	firstFiber := &rtui.Fiber{NodeID: 1, Instance: first}
	secondFiber := &rtui.Fiber{NodeID: 2, Instance: second}

	app.updateHoveredFiber(firstFiber, nil)
	app.updateHoveredFiber(secondFiber, nil)
	app.updateHoveredFiber(nil, nil)

	if len(first.actions) != 2 || first.actions[0] != action.ActionMouseEnter || first.actions[1] != action.ActionMouseLeave {
		t.Fatalf("first hover actions = %#v, want [mouse_enter mouse_leave]", first.actions)
	}
	if len(second.actions) != 2 || second.actions[0] != action.ActionMouseEnter || second.actions[1] != action.ActionMouseLeave {
		t.Fatalf("second hover actions = %#v, want [mouse_enter mouse_leave]", second.actions)
	}
}
