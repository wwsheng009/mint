package action

import (
	"github.com/wwsheng009/mint/framework/component"
	frameworkevent "github.com/wwsheng009/mint/framework/event"
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
	runtimeplatform "github.com/wwsheng009/mint/runtime/platform"
)

// ============================================================================
// Phase 0: 适配器 - 将旧接口转换为 ActionTarget
// ============================================================================

// UpdaterAdapter 将 Updater 接口适配为 ActionTarget
// 用于过渡期，允许旧组件继续使用 Update(msg) 接口
type UpdaterAdapter struct {
	updater component.Updater
	id      uint64
}

// NewUpdaterAdapter 创建 Updater 适配器
func NewUpdaterAdapter(updater component.Updater, id uint64) *UpdaterAdapter {
	return &UpdaterAdapter{
		updater: updater,
		id:      id,
	}
}

// HandleAction 将 Action 转换为 Msg 并调用 Update
func (a *UpdaterAdapter) HandleAction(act *Action) bool {
	// 将 Action 转换为 Msg
	msg := ActionToMsg(act)
	if msg == nil {
		return false
	}

	// 调用旧的 Update 方法
	cmd := a.updater.Update(msg)
	// TODO: 执行 Cmd
	return cmd != nil
}

// GetSupportedActions 返回支持的 Action 类型
func (a *UpdaterAdapter) GetSupportedActions() []ActionType {
	// 保守估计：假设支持常见 Action
	return []ActionType{
		ActionClick,
		ActionInputText,
		ActionNavigateUp,
		ActionNavigateDown,
		ActionNavigateLeft,
		ActionNavigateRight,
		ActionEnter,
		ActionBackspace,
	}
}

// CanHandleAction 检查是否能处理该 Action
func (a *UpdaterAdapter) CanHandleAction(act *Action) bool {
	return true // 保守估计
}

// EventHandlerAdapter 将 EventHandler 接口适配为 ActionTarget
// 用于过渡期，允许旧组件继续使用 HandleEvent(Event) 接口
type EventHandlerAdapter struct {
	handler frameworkevent.EventHandler
	id      uint64
}

// NewEventHandlerAdapter 创建 EventHandler 适配器
func NewEventHandlerAdapter(handler frameworkevent.EventHandler, id uint64) *EventHandlerAdapter {
	return &EventHandlerAdapter{
		handler: handler,
		id:      id,
	}
}

// HandleAction 将 Action 转换为 Event 并调用 HandleEvent
func (a *EventHandlerAdapter) HandleAction(act *Action) bool {
	// 将 Action 转换为 Event
	event := ActionToEvent(act)
	if event == nil {
		return false
	}

	// 调用旧的 HandleEvent 方法
	return a.handler.HandleEvent(event)
}

// GetSupportedActions 返回支持的 Action 类型
func (a *EventHandlerAdapter) GetSupportedActions() []ActionType {
	return []ActionType{
		ActionClick,
		ActionDoubleClick,
		ActionRightClick,
		ActionHover,
		ActionScroll,
	}
}

// CanHandleAction 检查是否能处理该 Action
func (a *EventHandlerAdapter) CanHandleAction(act *Action) bool {
	return act.IsMouse()
}

// ============================================================================
// Action 与其他类型之间的转换
// ============================================================================

// ActionToMsg 将 Action 转换为 Msg
// 这是 UpdaterAdapter 的辅助函数
func ActionToMsg(act *Action) runtimemsg.Msg {
	emptyMod := runtimemsg.Modifiers{}

	switch act.Type {
	case ActionInputText:
		if s, ok := act.GetPayloadString(); ok && len(s) > 0 {
			return runtimemsg.NewKeyMsg([]rune(s)[0], runtimeplatform.KeyUnknown, emptyMod)
		}
	case ActionNavigateUp:
		return runtimemsg.NewKeyMsg(0, runtimeplatform.KeyUp, emptyMod)
	case ActionNavigateDown:
		return runtimemsg.NewKeyMsg(0, runtimeplatform.KeyDown, emptyMod)
	case ActionNavigateLeft:
		return runtimemsg.NewKeyMsg(0, runtimeplatform.KeyLeft, emptyMod)
	case ActionNavigateRight:
		return runtimemsg.NewKeyMsg(0, runtimeplatform.KeyRight, emptyMod)
	case ActionEnter:
		return runtimemsg.NewKeyMsg(0, runtimeplatform.KeyEnter, emptyMod)
	case ActionBackspace:
		return runtimemsg.NewKeyMsg(0, runtimeplatform.KeyBackspace, emptyMod)
	case ActionClick:
		if x, y, ok := act.GetPayloadPoint(); ok {
			return runtimemsg.NewMouseMsg(x, y, runtimemsg.MouseLeft, runtimemsg.MouseActionPress)
		}
	}
	return nil
}

// ActionToEvent 将 Action 转换为 Event
// 这是 EventHandlerAdapter 的辅助函数
func ActionToEvent(act *Action) frameworkevent.Event {
	// 注意：这里需要创建 Event 接口的实现
	// 由于 Event 接口的具体实现可能在 event 包中，这里返回 nil
	// 实际使用时可能需要根据具体的 Event 类型进行转换
	return nil
}
