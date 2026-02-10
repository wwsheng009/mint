package action

import (
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
	runtimeplatform "github.com/wwsheng009/mint/runtime/platform"
)

// InputProcessor 将 Msg 转换为语义化的 Action
//
// InputProcessor 是 Msg 和 Action 之间的桥梁，它：
// 1. 接收键盘/鼠标消息（Msg）
// 2. 应用转换规则（包括 KeyMap 查找）
// 3. 返回语义化的 Action
//
// 使用示例：
//   processor := NewInputProcessor()
//   processor.SetKeyMap(defaultKeyMap)  // 可选
//   action := processor.ProcessMsg(mouseMsg)
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

// ProcessMsg 将 Msg 转换为 Action
//
// 如果消息无法识别或转换为 Action，返回 nil
func (p *InputProcessor) ProcessMsg(message runtimemsg.Msg) *Action {
	if message == nil {
		return nil
	}

	switch m := message.(type) {
	case *runtimemsg.KeyMsg:
		return p.processKeyMsg(m)

	case *runtimemsg.MouseMsg:
		return p.processMouseMsg(m)

	default:
		// 不支持的消息类型
		return nil
	}
}

// processKeyMsg 处理键盘消息
func (p *InputProcessor) processKeyMsg(keyMsg *runtimemsg.KeyMsg) *Action {
	// 1. 优先使用 KeyMap 语义化（如果设置了）
	if p.keyMap != nil {
		if act := p.keyMap.LookupKeyMsg(keyMsg); act != nil {
			return act
		}
	}

	// 2. 默认转换规则
	return p.applyDefaultKeyMapping(keyMsg)
}

// processMouseMsg 处理鼠标消息
func (p *InputProcessor) processMouseMsg(mouseMsg *runtimemsg.MouseMsg) *Action {
	// 鼠标事件转换为 Action 语义
	switch mouseMsg.Action {
	case runtimemsg.MouseActionPress:
		if mouseMsg.Button == runtimemsg.MouseLeft {
			return NewAction(ActionClick).
				WithTarget(mouseMsg.TargetID).
				WithPayload(struct{ X, Y int }{mouseMsg.LocalX, mouseMsg.LocalY})
		} else if mouseMsg.Button == runtimemsg.MouseRight {
			return NewAction(ActionRightClick).
				WithTarget(mouseMsg.TargetID)
		}

	case runtimemsg.MouseActionWheel:
		return NewAction(ActionScroll).
			WithTarget(mouseMsg.TargetID).
			WithPayload(mouseMsg.Delta)

	case runtimemsg.MouseActionMove:
		return NewAction(ActionHover).
			WithTarget(mouseMsg.TargetID).
			WithPayload(struct{ X, Y int }{mouseMsg.LocalX, mouseMsg.LocalY})
	}

	return nil
}

// applyDefaultKeyMapping 应用默认的键盘映射规则
func (p *InputProcessor) applyDefaultKeyMapping(keyMsg *runtimemsg.KeyMsg) *Action {
	// 导航键
	switch keyMsg.Special {
	case runtimeplatform.KeyUp:
		return NewActionFromKey(ActionNavigateUp, "keyboard")
	case runtimeplatform.KeyDown:
		return NewActionFromKey(ActionNavigateDown, "keyboard")
	case runtimeplatform.KeyLeft:
		return NewActionFromKey(ActionNavigateLeft, "keyboard")
	case runtimeplatform.KeyRight:
		return NewActionFromKey(ActionNavigateRight, "keyboard")
	case runtimeplatform.KeyPageUp:
		return NewActionFromKey(ActionNavigatePageUp, "keyboard")
	case runtimeplatform.KeyPageDown:
		return NewActionFromKey(ActionNavigatePageDown, "keyboard")
	case runtimeplatform.KeyHome:
		return NewActionFromKey(ActionNavigateHome, "keyboard")
	case runtimeplatform.KeyEnd:
		return NewActionFromKey(ActionNavigateEnd, "keyboard")
	}

	// 编辑键
	switch keyMsg.Special {
	case runtimeplatform.KeyEnter:
		return NewActionFromKey(ActionEnter, "keyboard")
	case runtimeplatform.KeyTab:
		if keyMsg.Mod.Shift {
			return NewActionFromKey(ActionNavigatePrev, "keyboard")
		}
		return NewActionFromKey(ActionNavigateNext, "keyboard")
	case runtimeplatform.KeyBackspace:
		return NewActionFromKey(ActionBackspace, "keyboard")
	case runtimeplatform.KeyDelete:
		return NewActionFromKey(ActionDeleteChar, "keyboard")
	case runtimeplatform.KeyEscape:
		return NewActionFromKey(ActionCancel, "keyboard")
	}

	// 功能键
	switch keyMsg.Special {
	case runtimeplatform.KeyF1:
		return NewActionFromKey(ActionInspect, "keyboard")
	case runtimeplatform.KeyF5:
		return NewActionFromKey(ActionRefresh, "keyboard")
	case runtimeplatform.KeyF10:
		return NewActionFromKey(ActionQuit, "keyboard")
	case runtimeplatform.KeyF12:
		return NewActionFromKey(ActionInspect, "keyboard")
	}

	// 可打印字符 → 输入文本
	if keyMsg.IsPrintable() && !keyMsg.HasModifier() {
		return NewAction(ActionInputText).
			WithPayload(string(keyMsg.Rune))
	}

	// Ctrl 组合键的默认处理
	if keyMsg.HasCtrl() {
		switch keyMsg.Special {
		case runtimeplatform.KeyUnknown: // Ctrl+字母键
			switch keyMsg.Rune {
			case 'c', 'C':
				return NewActionFromKey(ActionCopy, "keyboard")
			case 'v', 'V':
				return NewActionFromKey(ActionPaste, "keyboard")
			case 'x', 'X':
				return NewActionFromKey(ActionCut, "keyboard")
			case 'f', 'F':
				return NewActionFromKey(ActionSearch, "keyboard")
			case 'q', 'Q':
				return NewActionFromKey(ActionQuit, "keyboard")
			case 's', 'S':
				return NewActionFromKey(ActionSubmit, "keyboard")
			}
		}
	}

	return nil
}
