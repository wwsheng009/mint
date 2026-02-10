package action

import (
	"fmt"

	frameworkevent "github.com/wwsheng009/mint/framework/event"
	runtimeevent "github.com/wwsheng009/mint/runtime/event"
)

// InputProcessor 将原始 Event 转换为语义化的 Action
//
// InputProcessor 是 Event 和 Action 之间的桥梁，它：
// 1. 接收原始的键盘/鼠标事件
// 2. 应用转换规则（包括 KeyMap 查找）
// 3. 返回语义化的 Action
//
// 使用示例：
//   processor := NewInputProcessor()
//   processor.SetKeyMap(defaultKeyMap)  // 可选
//   action := processor.Process(mouseEvent)
type InputProcessor struct {
	keyMap *KeyMap // 键盘映射（可以为 nil，使用默认规则）
}

// NewInputProcessor 创建新的 InputProcessor
func NewInputProcessor() *InputProcessor {
	return &InputProcessor{
		keyMap: nil,
	}
}

// SetKeyMap 设置键盘映射
func (p *InputProcessor) SetKeyMap(keyMap *KeyMap) {
	p.keyMap = keyMap
}

// GetKeyMap 获取当前的键盘映射
func (p *InputProcessor) GetKeyMap() *KeyMap {
	return p.keyMap
}

// Process 将 Event 转换为 Action
//
// 如果事件无法识别或转换为 Action，返回 nil
func (p *InputProcessor) Process(event frameworkevent.Event) *Action {
	if event == nil {
		return nil
	}

	switch e := event.(type) {
	case *frameworkevent.KeyEvent:
		return p.processKeyEvent(e)

	case *frameworkevent.MouseEvent:
		return p.processMouseEvent(e)

	default:
		// 不支持的事件类型
		return nil
	}
}

// processKeyEvent 处理键盘事件
func (p *InputProcessor) processKeyEvent(ev *frameworkevent.KeyEvent) *Action {
	// 1. 优先使用 KeyMap 语义化（如果设置了）
	if p.keyMap != nil {
		if act := p.keyMap.LookupKeyEvent(ev); act != nil {
			return act
		}
	}

	// 2. 默认转换规则
	return p.applyDefaultKeyMapping(ev)
}

// applyDefaultKeyMapping 应用默认的键盘映射规则
func (p *InputProcessor) applyDefaultKeyMapping(ev *frameworkevent.KeyEvent) *Action {
	// 导航键
	switch ev.Special {
	case frameworkevent.KeyUp:
		return NewActionFromKey(ActionNavigateUp, "keyboard")
	case frameworkevent.KeyDown:
		return NewActionFromKey(ActionNavigateDown, "keyboard")
	case frameworkevent.KeyLeft:
		return NewActionFromKey(ActionNavigateLeft, "keyboard")
	case frameworkevent.KeyRight:
		return NewActionFromKey(ActionNavigateRight, "keyboard")
	case frameworkevent.KeyPageUp:
		return NewActionFromKey(ActionNavigatePageUp, "keyboard")
	case frameworkevent.KeyPageDown:
		return NewActionFromKey(ActionNavigatePageDown, "keyboard")
	case frameworkevent.KeyHome:
		return NewActionFromKey(ActionNavigateHome, "keyboard")
	case frameworkevent.KeyEnd:
		return NewActionFromKey(ActionNavigateEnd, "keyboard")
	}

	// 编辑键
	switch ev.Special {
	case frameworkevent.KeyEnter:
		return NewActionFromKey(ActionEnter, "keyboard")
	case frameworkevent.KeyTab:
		return NewActionFromKey(ActionNavigateNext, "keyboard")
	case frameworkevent.KeyBackspace:
		return NewActionFromKey(ActionBackspace, "keyboard")
	case frameworkevent.KeyDelete:
		return NewActionFromKey(ActionDeleteChar, "keyboard")
	case frameworkevent.KeyEscape:
		return NewActionFromKey(ActionCancel, "keyboard")
	}

	// 功能键
	switch ev.Special {
	case frameworkevent.KeyF1:
		return NewActionFromKey(ActionInspect, "keyboard") // F1 切换 Inspector
	case frameworkevent.KeyF5:
		return NewActionFromKey(ActionRefresh, "keyboard") // F5 刷新
	case frameworkevent.KeyF10:
		return NewActionFromKey(ActionQuit, "keyboard") // F10 退出
	}

	// Ctrl/Alt 组合键
	if ev.Key.Alt || ev.Key.Ctrl {
		return p.applyModifierMapping(ev)
	}

	// 可打印字符（文本输入）
	if ev.Key.Rune >= 32 && ev.Key.Rune <= 126 {
		// 普通字符作为文本输入
		return NewActionWithPayload(
			ActionInputText,
			string(ev.Key.Rune),
		).WithSource("keyboard")
	}

	// 无法识别的按键
	return nil
}

// applyModifierMapping 应用修饰键组合映射
func (p *InputProcessor) applyModifierMapping(ev *frameworkevent.KeyEvent) *Action {
	key := ev.Key.Rune

	// Ctrl 组合
	if ev.Key.Ctrl {
		switch key {
		case 'c': // Ctrl+C - 复制
			return NewActionFromKey(ActionCopy, "keyboard")
		case 'v': // Ctrl+V - 粘贴
			return NewActionFromKey(ActionPaste, "keyboard")
		case 'x': // Ctrl+X - 剪切
			return NewActionFromKey(ActionCut, "keyboard")
		case 'f': // Ctrl+F - 搜索
			return NewActionFromKey(ActionSearch, "keyboard")
		case 'q': // Ctrl+Q - 退出
			return NewActionFromKey(ActionQuit, "keyboard")
		case 's': // Ctrl+S - 保存/提交
			return NewActionFromKey(ActionSubmit, "keyboard")
		case 'a': // Ctrl+A - 全选（这里映射到 NavigateHome）
			return NewActionFromKey(ActionNavigateHome, "keyboard")
		case 'e': // Ctrl+E - 跳到行尾
			return NewActionFromKey(ActionNavigateEnd, "keyboard")
		}
	}

	// Alt 组合
	if ev.Key.Alt {
		switch key {
		case ' ': // Alt+Space - 显示菜单（这里映射到 Toggle）
			return NewActionFromKey(ActionToggle, "keyboard")
		}
	}

	return nil
}

// processMouseEvent 处理鼠标事件
func (p *InputProcessor) processMouseEvent(ev *frameworkevent.MouseEvent) *Action {
	// 如果没有命中目标，返回 nil
	if ev.TargetID == "" {
		return nil
	}

	// 根据鼠标动作类型转换（Action 字段是 runtime/event.MouseAction 类型）
	switch ev.Action {
	case runtimeevent.MouseActionPress:
		return p.processMousePress(ev)
	case runtimeevent.MouseActionRelease:
		return p.processMouseRelease(ev)
	case runtimeevent.MouseActionMove:
		return p.processMouseMove(ev)
	case runtimeevent.MouseActionWheel:
		return p.processMouseWheel(ev)
	default:
		return nil
	}
}

// processMousePress 处理鼠标按下
func (p *InputProcessor) processMousePress(ev *frameworkevent.MouseEvent) *Action {
	switch ev.Button {
	case frameworkevent.MouseLeft:
		return NewActionFromMouse(
			ActionClick,
			ev.TargetID,
			ev.LocalX,
			ev.LocalY,
		)

	case frameworkevent.MouseRight:
		return NewActionFromMouse(
			ActionRightClick,
			ev.TargetID,
			ev.LocalX,
			ev.LocalY,
		)

	case frameworkevent.MouseMiddle:
		return NewActionFromMouse(
			ActionMiddleClick,
			ev.TargetID,
			ev.LocalX,
			ev.LocalY,
		)

	default:
		return nil
	}
}

// processMouseRelease 处理鼠标释放
func (p *InputProcessor) processMouseRelease(ev *frameworkevent.MouseEvent) *Action {
	// 通常鼠标释放不产生独立的 Action
	// 特殊情况可以在需要时添加
	return nil
}

// processMouseMove 处理鼠标移动
func (p *InputProcessor) processMouseMove(ev *frameworkevent.MouseEvent) *Action {
	// 鼠标移动产生 Hover Action
	return NewActionFromMouse(
		ActionHover,
		ev.TargetID,
		ev.LocalX,
		ev.LocalY,
	)
}

// processMouseWheel 处理鼠标滚轮
func (p *InputProcessor) processMouseWheel(ev *frameworkevent.MouseEvent) *Action {
	// Delta 已经是 int 类型
	return &Action{
		Type:     ActionScroll,
		Payload:  ev.Delta,
		Source:   "mouse",
		TargetID: ev.TargetID,
	}
}

// ============================================================================
// 批量处理支持
// ============================================================================

// ProcessBatch 批量处理事件
// 返回成功转换的 Action 列表
func (p *InputProcessor) ProcessBatch(events []frameworkevent.Event) []*Action {
	actions := make([]*Action, 0, len(events))

	for _, ev := range events {
		if act := p.Process(ev); act != nil {
			actions = append(actions, act)
		}
	}

	return actions
}

// ProcessWithCallback 处理事件并使用回调函数
// 回调函数返回 true 表示停止处理
func (p *InputProcessor) ProcessWithCallback(
	events []frameworkevent.Event,
	callback func(*Action) bool,
) {
	for _, ev := range events {
		act := p.Process(ev)
		if act == nil {
			continue
		}

		if callback(act) {
			break
		}
	}
}

// ============================================================================
// 调试和诊断
// ============================================================================

// String 返回处理器的字符串表示
func (p *InputProcessor) String() string {
	hasKeyMap := p.keyMap != nil
	return fmt.Sprintf("InputProcessor{KeyMap=%v}", hasKeyMap)
}

// Stats 返回处理器统计信息（用于调试）
func (p *InputProcessor) Stats() map[string]interface{} {
	return map[string]interface{}{
		"has_keymap": p.keyMap != nil,
	}
}
