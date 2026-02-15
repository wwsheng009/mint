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
//
// Action 到 Event 的映射：
// - ActionClick → EventClick
// - ActionDoubleClick → EventDoubleClick
// - ActionRightClick → EventContextMenu
// - ActionHover → EventMouseEnter (鼠标悬停)
// - ActionScroll → EventMouseWheel
// - ActionDragStart/DragMove/DragEnd → 不直接映射（需要特殊处理）
// - ActionInputText → EventChange (内容改变)
// - ActionEnter → EventSubmit (提交)
// - ActionBackspace → EventChange (删除字符)
// - ActionNavigateUp/Down → 无直接映射（导航由 FocusManager 处理）
func ActionToEvent(act *Action) frameworkevent.Event {
	if act == nil {
		return nil
	}

	// 根据 Action 类型创建对应的 Event
	var eventType frameworkevent.EventType

	switch act.Type {
	case ActionClick:
		eventType = frameworkevent.EventClick
	case ActionDoubleClick:
		eventType = frameworkevent.EventDoubleClick
	case ActionRightClick:
		eventType = frameworkevent.EventContextMenu
	case ActionHover:
		eventType = frameworkevent.EventMouseEnter
	case ActionScroll:
		eventType = frameworkevent.EventMouseWheel
	case ActionInputText, ActionDeleteChar, ActionDeleteWord,
	     ActionDeleteLine, ActionBackspace:
		// 文本输入/删除操作对应 Change 事件
		eventType = frameworkevent.EventChange
	case ActionEnter:
		// 回车键可能表示提交
		eventType = frameworkevent.EventSubmit
	case ActionToggle:
		// 切换状态
		eventType = frameworkevent.EventChange
	case ActionSelect:
		eventType = frameworkevent.EventSelect
	case ActionExpand:
		eventType = frameworkevent.EventExpand
	case ActionCollapse:
		eventType = frameworkevent.EventCollapse
	case ActionSubmit:
		eventType = frameworkevent.EventSubmit
	case ActionCancel:
		eventType = frameworkevent.EventCancel
	case ActionFocus:
		eventType = frameworkevent.EventFocus
	case ActionBlur:
		eventType = frameworkevent.EventBlur
	default:
		// 不支持的 Action 类型，返回 nil
		return nil
	}

	// 创建并返回 Framework Event
	// 注意：由于 Action 不包含 Component 目标信息，
	// 调用者需要设置适当的 Target
	evt := frameworkevent.NewBaseEvent(eventType)

	return evt
}
