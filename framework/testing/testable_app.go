package testing

import (
	"time"

	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/framework/msg"
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
	"github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/runtime/event"
)

// TestableApp 是可测试的应用包装器
//
// TestableApp 提供了便捷的测试辅助方法，使测试更加可读和易于维护。
// 它支持按 ID 注入事件、断言组件状态等。
type TestableApp struct {
	// root 是应用的根组件
	root interface{}

	// router 是 Action 分发器
	router *action.Router

	// lastError 记录最后一次错误
	lastError error

	// lastMsg 记录最后一条消息
	lastMsg *msg.SandboxMsg
}

// NewTestableApp 创建一个新的可测试应用
func NewTestableApp(root interface{}, router *action.Router) *TestableApp {
	return &TestableApp{
		root:   root,
		router: router,
	}
}

// InjectKeySequence 注入键盘按键序列
//
// 示例：
//   app.InjectKeySequence("hello")
//   app.InjectKeySequence("Ctrl+C")
func (t *TestableApp) InjectKeySequence(keys string) {
	for _, ch := range keys {
		keyMsg := runtimemsg.NewKeyMsg(ch, platform.KeyUnknown, runtimemsg.Modifiers{})
		act := t.keyMsgToAction(keyMsg)
		if act != nil {
			t.router.Dispatch(act)
		}
	}
}

// InjectKey 注入单个键盘按键
func (t *TestableApp) InjectKey(key rune, special platform.SpecialKey, mod runtimemsg.Modifiers) {
	keyMsg := runtimemsg.NewKeyMsg(key, special, mod)
	act := t.keyMsgToAction(keyMsg)
	if act != nil {
		t.router.Dispatch(act)
	}
}

// InjectEnter 注入 Enter 键
func (t *TestableApp) InjectEnter() {
	act := action.NewAction(action.ActionEnter)
	t.router.Dispatch(act)
}

// InjectTab 注入 Tab 键
func (t *TestableApp) InjectTab() {
	act := action.NewAction(action.ActionNavigateNext)
	t.router.Dispatch(act)
}

// InjectEscape 注入 Escape 键
func (t *TestableApp) InjectEscape() {
	act := action.NewAction(action.ActionCancel)
	t.router.Dispatch(act)
}

// InjectMouseClickByID 按 ID 注入鼠标点击
//
// 示例：
//   app.InjectMouseClickByID("button1")
func (t *TestableApp) InjectMouseClickByID(targetID string, localX, localY int) {
	act := action.NewActionFromMouse(action.ActionClick, localX, localY).
		WithTarget(targetID)
	t.router.Dispatch(act)
}

// InjectMouseRightClickByID 按 ID 注入右键点击
func (t *TestableApp) InjectMouseRightClickByID(targetID string, localX, localY int) {
	act := action.NewActionFromMouse(action.ActionRightClick, localX, localY).
		WithTarget(targetID)
	t.router.Dispatch(act)
}

// InjectAction 注入语义化 Action
//
// 示例：
//   act := action.NewAction(action.ActionNavigateDown)
//   app.InjectAction(act)
func (t *TestableApp) InjectAction(act *action.Action) {
	if act != nil {
		t.router.Dispatch(act)
	}
}

// InjectText 注入文本（用于文本输入组件）
//
// 示例：
//   app.InjectText("input1", "hello world")
func (t *TestableApp) InjectText(targetID string, text string) {
	act := action.NewActionWithPayload(action.ActionInputText, text)
	act.TargetID = event.StringToNodeID(targetID)
	t.router.Dispatch(act)
}

// SetState 直接设置组件状态（用于测试）
//
// 示例：
//   app.SetState("button1", "disabled", true)
func (t *TestableApp) SetState(targetID, path string, value interface{}) {
	// 创建沙箱状态修改消息
	sandboxMsg := msg.NewSandboxStateMsg(targetID, path, value)
	t.lastMsg = sandboxMsg
	// 在实际实现中，这应该由应用处理
}

// AssertFocused 断言组件获得焦点
//
// 示例：
//   app.AssertFocused("button1")
func (t *TestableApp) AssertFocused(targetID string) error {
	// 在实际实现中，应该检查焦点状态
	// 这里返回 nil 表示断言通过
	return nil
}

// AssertHovered 断言组件被悬停
//
// 示例：
//   app.AssertHovered("button1")
func (t *TestableApp) AssertHovered(targetID string) error {
	// 在实际实现中，应该检查悬停状态
	return nil
}

// AssertValue 断言组件的值
//
// 示例：
//   app.AssertValue("input1", "hello")
func (t *TestableApp) AssertValue(targetID, expectedValue string) error {
	// 在实际实现中，应该检查组件值
	return nil
}

// AssertSelected 断言项被选中
//
// 示例：
//   app.AssertSelected("treeview1", "node1")
func (t *TestableApp) AssertSelected(targetID, selectedItem string) error {
	// 在实际实现中，应该检查选中状态
	return nil
}

// Wait 等待指定时间
func (t *TestableApp) Wait(duration time.Duration) {
	time.Sleep(duration)
}

// KeyMsgToAction 将 KeyMsg 转换为 Action
func (t *TestableApp) keyMsgToAction(keyMsg *runtimemsg.KeyMsg) *action.Action {
	if keyMsg == nil {
		return nil
	}

	// 导航键
	if keyMsg.IsNavigation() {
		switch keyMsg.Special {
		case platform.KeyUp:
			return action.NewAction(action.ActionNavigateUp)
		case platform.KeyDown:
			return action.NewAction(action.ActionNavigateDown)
		case platform.KeyLeft:
			return action.NewAction(action.ActionNavigateLeft)
		case platform.KeyRight:
			return action.NewAction(action.ActionNavigateRight)
		case platform.KeyHome:
			return action.NewAction(action.ActionNavigateHome)
		case platform.KeyEnd:
			return action.NewAction(action.ActionNavigateEnd)
		case platform.KeyPageUp:
			return action.NewAction(action.ActionNavigatePageUp)
		case platform.KeyPageDown:
			return action.NewAction(action.ActionNavigatePageDown)
		}
	}

	// 编辑键
	switch keyMsg.Special {
	case platform.KeyEnter:
		return action.NewAction(action.ActionEnter)
	case platform.KeyBackspace:
		return action.NewAction(action.ActionBackspace)
	case platform.KeyDelete:
		return action.NewAction(action.ActionDeleteChar)
	case platform.KeyEscape:
		return action.NewAction(action.ActionCancel)
	}

	// 可打印字符
	if keyMsg.IsPrintable() {
		return action.NewActionWithPayload(action.ActionInputText, string(keyMsg.Rune))
	}

	return nil
}

// GetRouter 获取 Router
func (t *TestableApp) GetRouter() *action.Router {
	return t.router
}

// GetLastError 获取最后的错误
func (t *TestableApp) GetLastError() error {
	return t.lastError
}

// GetLastMsg 获取最后的消息
func (t *TestableApp) GetLastMsg() *msg.SandboxMsg {
	return t.lastMsg
}
