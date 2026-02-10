package sandbox

import (
	"github.com/wwsheng009/mint/framework/action"
	"github.com/wwsheng009/mint/framework/msg"
	"github.com/wwsheng009/mint/runtime/platform"
)

// Injector 是测试沙箱注入器
//
// Injector 提供了便捷的方法来注入各种类型的输入和操作，
// 用于测试和自动化场景。
type Injector struct {
	// messages 是注入的消息队列
	messages []msg.Msg

	// actions 是注入的 Action 队列
	actions []*action.Action
}

// NewInjector 创建一个新的注入器
func NewInjector() *Injector {
	return &Injector{
		messages: make([]msg.Msg, 0),
		actions: make([]*action.Action, 0),
	}
}

// ============================================================================
// 键盘注入
// ============================================================================

// InjectKey 注入单个按键
func (i *Injector) InjectKey(key rune, special platform.SpecialKey, mod msg.Modifiers) *Injector {
	keyMsg := msg.NewKeyMsg(key, special, mod)
	sandboxMsg := msg.NewSandboxKeyMsg(keyMsg)
	i.messages = append(i.messages, sandboxMsg)
	return i
}

// InjectKeySequence 注入按键序列
//
// 示例：
//   injector.InjectKeySequence("hello")
func (i *Injector) InjectKeySequence(keys string) *Injector {
	for _, ch := range keys {
		i.InjectKey(ch, platform.KeyUnknown, msg.Modifiers{})
	}
	return i
}

// InjectChar 注入单个字符
func (i *Injector) InjectChar(ch rune) *Injector {
	return i.InjectKey(ch, platform.KeyUnknown, msg.Modifiers{})
}

// InjectEnter 注入 Enter 键
func (i *Injector) InjectEnter() *Injector {
	act := action.NewAction(action.ActionEnter)
	i.actions = append(i.actions, act)
	return i
}

// InjectTab 注入 Tab 键
func (i *Injector) InjectTab() *Injector {
	act := action.NewAction(action.ActionNavigateNext)
	i.actions = append(i.actions, act)
	return i
}

// InjectBackspace 注入 Backspace 键
func (i *Injector) InjectBackspace() *Injector {
	act := action.NewAction(action.ActionBackspace)
	i.actions = append(i.actions, act)
	return i
}

// InjectDelete 注入 Delete 键
func (i *Injector) InjectDelete() *Injector {
	act := action.NewAction(action.ActionDeleteChar)
	i.actions = append(i.actions, act)
	return i
}

// InjectEscape 注入 Escape 键
func (i *Injector) InjectEscape() *Injector {
	act := action.NewAction(action.ActionCancel)
	i.actions = append(i.actions, act)
	return i
}

// InjectUp 注入向上键
func (i *Injector) InjectUp() *Injector {
	act := action.NewAction(action.ActionNavigateUp)
	i.actions = append(i.actions, act)
	return i
}

// InjectDown 注入向下键
func (i *Injector) InjectDown() *Injector {
	act := action.NewAction(action.ActionNavigateDown)
	i.actions = append(i.actions, act)
	return i
}

// InjectLeft 注入向左键
func (i *Injector) InjectLeft() *Injector {
	act := action.NewAction(action.ActionNavigateLeft)
	i.actions = append(i.actions, act)
	return i
}

// InjectRight 注入向右键
func (i *Injector) InjectRight() *Injector {
	act := action.NewAction(action.ActionNavigateRight)
	i.actions = append(i.actions, act)
	return i
}

// InjectCtrlKey 注入 Ctrl 组合键
//
// 示例：
//   injector.InjectCtrlKey('c') // Ctrl+C
func (i *Injector) InjectCtrlKey(key rune) *Injector {
	keyMsg := msg.NewKeyMsg(key, platform.KeyUnknown, msg.Modifiers{Ctrl: true})
	sandboxMsg := msg.NewSandboxKeyMsg(keyMsg)
	i.messages = append(i.messages, sandboxMsg)
	return i
}

// InjectAltKey 注入 Alt 组合键
func (i *Injector) InjectAltKey(key rune) *Injector {
	keyMsg := msg.NewKeyMsg(key, platform.KeyUnknown, msg.Modifiers{Alt: true})
	sandboxMsg := msg.NewSandboxKeyMsg(keyMsg)
	i.messages = append(i.messages, sandboxMsg)
	return i
}

// ============================================================================
// 鼠标注入
// ============================================================================

// InjectMouseClick 注入鼠标点击
//
// 示例：
//   injector.InjectMouseClick("button1", 10, 5)
func (i *Injector) InjectMouseClick(targetID string, localX, localY int) *Injector {
	act := action.NewActionFromMouse(action.ActionClick, targetID, localX, localY)
	i.actions = append(i.actions, act)
	return i
}

// InjectMouseRightClick 注入右键点击
func (i *Injector) InjectMouseRightClick(targetID string, localX, localY int) *Injector {
	act := action.NewActionFromMouse(action.ActionRightClick, targetID, localX, localY)
	i.actions = append(i.actions, act)
	return i
}

// InjectMouseMiddleClick 注入中键点击
func (i *Injector) InjectMouseMiddleClick(targetID string, localX, localY int) *Injector {
	act := action.NewActionFromMouse(action.ActionMiddleClick, targetID, localX, localY)
	i.actions = append(i.actions, act)
	return i
}

// InjectMouseWheel 注入鼠标滚轮
func (i *Injector) InjectMouseWheel(delta int) *Injector {
	act := action.NewActionWithPayload(action.ActionScroll, delta)
	i.actions = append(i.actions, act)
	return i
}

// ============================================================================
// Action 注入
// ============================================================================

// InjectAction 注入语义化 Action
//
// 示例：
//   act := action.NewAction(action.ActionNavigateDown)
//   injector.InjectAction(act)
func (i *Injector) InjectAction(act *action.Action) *Injector {
	if act != nil {
		i.actions = append(i.actions, act)
	}
	return i
}

// InjectNavigate 注入导航 Action
func (i *Injector) InjectNavigate(actionType action.ActionType) *Injector {
	act := action.NewAction(actionType)
	i.actions = append(i.actions, act)
	return i
}

// InjectSelect 注入选择 Action
func (i *Injector) InjectSelect() *Injector {
	act := action.NewAction(action.ActionSelect)
	i.actions = append(i.actions, act)
	return i
}

// InjectToggle 注入切换 Action
func (i *Injector) InjectToggle() *Injector {
	act := action.NewAction(action.ActionToggle)
	i.actions = append(i.actions, act)
	return i
}

// ============================================================================
// 状态注入
// ============================================================================

// InjectSetState 注入状态修改
//
// 示例：
//   injector.InjectSetState("button1", "disabled", true)
//   injector.InjectSetState("input1", "value", "test")
func (i *Injector) InjectSetState(targetID, path string, value interface{}) *Injector {
	stateMsg := msg.NewSandboxStateMsg(targetID, path, value)
	i.messages = append(i.messages, stateMsg)
	return i
}

// InjectSetValue 注入值设置（快捷方法）
func (i *Injector) InjectSetValue(targetID, value string) *Injector {
	return i.InjectSetState(targetID, "value", value)
}

// InjectSetDisabled 注入禁用状态
func (i *Injector) InjectSetDisabled(targetID string, disabled bool) *Injector {
	return i.InjectSetState(targetID, "disabled", disabled)
}

// InjectSetFocused 注入焦点状态
func (i *Injector) InjectSetFocused(targetID string, focused bool) *Injector {
	return i.InjectSetState(targetID, "focused", focused)
}

// ============================================================================
// 获取注入的内容
// ============================================================================

// GetMessages 获取所有注入的消息
func (i *Injector) GetMessages() []msg.Msg {
	return i.messages
}

// GetActions 获取所有注入的 Action
func (i *Injector) GetActions() []*action.Action {
	return i.actions
}

// Clear 清空所有注入的内容
func (i *Injector) Clear() {
	i.messages = make([]msg.Msg, 0)
	i.actions = make([]*action.Action, 0)
}

// Count 返回注入的总数量
func (i *Injector) Count() int {
	return len(i.messages) + len(i.actions)
}

// ============================================================================
// 辅助方法
// ============================================================================

// HasMessages 检查是否有消息
func (i *Injector) HasMessages() bool {
	return len(i.messages) > 0
}

// HasActions 检查是否有 Action
func (i *Injector) HasActions() bool {
	return len(i.actions) > 0
}

// IsEmpty 检查是否为空
func (i *Injector) IsEmpty() bool {
	return len(i.messages) == 0 && len(i.actions) == 0
}
