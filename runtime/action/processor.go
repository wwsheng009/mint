package action

import (
	"github.com/wwsheng009/mint/internal/log"
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
// Usage:
//
//	processor := NewInputProcessor()
//	processor.SetKeyMap(defaultKeyMap)  // Optional
//	action := processor.ProcessMsg(mouseMsg)
type InputProcessor struct {
	keyMap *KeyMap // Keyboard mapping (can be nil for defaults)
}

// NewInputProcessor creates a new InputProcessor
func NewInputProcessor() *InputProcessor {
	return &InputProcessor{
		keyMap: nil,
	}
}

// SetKeyMap sets the keyboard mapping
func (p *InputProcessor) SetKeyMap(keyMap *KeyMap) {
	p.keyMap = keyMap
}

// GetKeyMap returns the current keyboard mapping
func (p *InputProcessor) GetKeyMap() *KeyMap {
	return p.keyMap
}

// ProcessMsg converts Msg to Action
// Returns nil if message cannot be recognized or converted
func (p *InputProcessor) ProcessMsg(message runtimemsg.Msg) *Action {
	if message == nil {
		return nil
	}

	switch m := message.(type) {
	case *runtimemsg.KeyMsg:
		return p.processKeyMsg(m)

	case *runtimemsg.PasteMsg:
		return p.processPasteMsg(m)

	case *runtimemsg.MouseMsg:
		return p.processMouseMsg(m)

	default:
		return nil
	}
}

// processKeyMsg handles keyboard messages
func (p *InputProcessor) processKeyMsg(keyMsg *runtimemsg.KeyMsg) *Action {
	// 1. Priority: Use KeyMap semantic mapping (if set)
	if p.keyMap != nil {
		if act := p.keyMap.LookupKeyMsg(keyMsg); act != nil {
			return act
		}
	}

	// 2. Default conversion rules
	return p.applyDefaultKeyMapping(keyMsg)
}

// processPasteMsg handles paste messages with complete text payloads.
func (p *InputProcessor) processPasteMsg(pasteMsg *runtimemsg.PasteMsg) *Action {
	if pasteMsg == nil || pasteMsg.Text == "" {
		return nil
	}
	return NewAction(ActionPaste).
		WithSource("paste").
		WithPayload(pasteMsg.Text)
}

// processMouseMsg handles mouse messages
func (p *InputProcessor) processMouseMsg(mouseMsg *runtimemsg.MouseMsg) *Action {
	switch mouseMsg.Action {
	case runtimemsg.MouseActionPress:
		if mouseMsg.Button == runtimemsg.MouseLeft {
			if log.RenderLogger.Enabled() {
				log.RenderLogger.Debug("[InputProcessor] mouse press -> ActionClick button=%v targetID=%d local=(%d,%d)", mouseMsg.Button, mouseMsg.TargetID, mouseMsg.LocalX, mouseMsg.LocalY)
			}
			act := NewAction(ActionClick).
				WithSource("mouse").
				WithPayload(mouseMsg)
			if mouseMsg.TargetID != 0 {
				act.WithTargetID(mouseMsg.TargetID)
			}
			return act
		} else if mouseMsg.Button == runtimemsg.MouseRight {
			act := NewAction(ActionRightClick).
				WithSource("mouse").
				WithPayload(mouseMsg)
			if mouseMsg.TargetID != 0 {
				act.WithTargetID(mouseMsg.TargetID)
			}
			return act
		} else if mouseMsg.Button == runtimemsg.MouseMiddle {
			act := NewAction(ActionMiddleClick).
				WithSource("mouse").
				WithPayload(mouseMsg)
			if mouseMsg.TargetID != 0 {
				act.WithTargetID(mouseMsg.TargetID)
			}
			return act
		}

	case runtimemsg.MouseActionRelease:
		// ✅ 修复：添加 Release 处理，让 pressed 状态能够正确重置
		// 根据 key_release.md 的设计思想，使用推断模式替代依赖事件
		if mouseMsg.Button == runtimemsg.MouseLeft {
			if log.RenderLogger.Enabled() {
				log.RenderLogger.Debug("[InputProcessor] mouse release -> ActionMouseRelease button=%v targetID=%d local=(%d,%d)", mouseMsg.Button, mouseMsg.TargetID, mouseMsg.LocalX, mouseMsg.LocalY)
			}
			act := NewAction(ActionMouseRelease).
				WithSource("mouse").
				WithPayload(mouseMsg)
			if mouseMsg.TargetID != 0 {
				act.WithTargetID(mouseMsg.TargetID)
			}
			return act
		} else if mouseMsg.Button == runtimemsg.MouseRight {
			act := NewAction(ActionRightRelease).
				WithSource("mouse").
				WithPayload(mouseMsg)
			if mouseMsg.TargetID != 0 {
				act.WithTargetID(mouseMsg.TargetID)
			}
			return act
		} else if mouseMsg.Button == runtimemsg.MouseMiddle {
			act := NewAction(ActionMiddleRelease).
				WithSource("mouse").
				WithPayload(mouseMsg)
			if mouseMsg.TargetID != 0 {
				act.WithTargetID(mouseMsg.TargetID)
			}
			return act
		}

	case runtimemsg.MouseActionWheel:
		act := NewAction(ActionScroll).
			WithSource("mouse").
			WithPayload(mouseMsg)
		if mouseMsg.TargetID != 0 {
			act.WithTargetID(mouseMsg.TargetID)
		}
		return act

	case runtimemsg.MouseActionMove:
		act := NewAction(ActionHover).
			WithSource("mouse").
			WithPayload(mouseMsg)
		if mouseMsg.TargetID != 0 {
			act.WithTargetID(mouseMsg.TargetID)
		}
		return act
	}

	return nil
}

// applyDefaultKeyMapping applies default keyboard mapping rules
func (p *InputProcessor) applyDefaultKeyMapping(keyMsg *runtimemsg.KeyMsg) *Action {
	// Navigation keys
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

	// Editing keys
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

	// Function keys
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

	// Ctrl combinations default handling
	if keyMsg.HasCtrl() {
		switch keyMsg.Special {
		case runtimeplatform.KeyUnknown: // Ctrl+letter keys
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
			case 'a', 'A':
				return NewActionFromKey(ActionNavigateHome, "keyboard")
			case 'e', 'E':
				return NewActionFromKey(ActionNavigateEnd, "keyboard")
			}
		}
	}
	// Printable characters → Input text
	// Shift is allowed (it changes the character, e.g., a→A, 1→!)
	// Ctrl and Alt are not allowed (they are shortcuts)
	if keyMsg.IsPrintable() && !keyMsg.HasCtrl() && !keyMsg.HasAlt() {
		return NewAction(ActionInputText).
			WithPayload(string(keyMsg.Rune))
	}

	return nil
}

// AddKeyMapping adds a custom key mapping
func (p *InputProcessor) AddKeyMapping(mapping KeyMapping) {
	if p.keyMap == nil {
		p.keyMap = NewKeyMap()
	}
	p.keyMap.Add(mapping)
}

// AddKeyMappings adds multiple key mappings
func (p *InputProcessor) AddKeyMappings(mappings []KeyMapping) {
	if p.keyMap == nil {
		p.keyMap = NewKeyMap()
	}
	p.keyMap.AddAll(mappings)
}
