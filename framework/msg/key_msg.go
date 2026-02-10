package msg

import (
	"fmt"
	"strings"
	"time"

	runtimeplatform "github.com/wwsheng009/mint/runtime/platform"
)

// KeyMsg 表示键盘输入消息
//
// KeyMsg 封装了键盘事件，包括：
// - 按键字符（Rune）
// - 特殊键（方向键、功能键等）
// - 修饰键（Ctrl、Alt、Shift）
type KeyMsg struct {
	BaseMsg

	// Rune 是按下的字符（对于可打印字符）
	// 如果是特殊键，Rune 为 0
	Rune rune

	// Special 是特殊键的类型（如果有的话）
	// 如果是普通字符输入，Special 为 KeyUnknown
	Special runtimeplatform.SpecialKey

	// Mod 按下的修饰键
	Mod Modifiers
}

// Modifiers 表示键盘修饰键的状态
type Modifiers struct {
	Alt   bool
	Ctrl  bool
	Shift bool
}

// NewKeyMsg 创建一个新的键盘消息
func NewKeyMsg(rune rune, special runtimeplatform.SpecialKey, mod Modifiers) *KeyMsg {
	return &KeyMsg{
		BaseMsg: BaseMsg{
			TypeValue:   MsgTypeKey,
			TimestampValue: time.Now(),
		},
		Rune:    rune,
		Special: special,
		Mod:     mod,
	}
}

// NewKeyMsgFromKeyEvent 从 runtime KeyEvent 创建 KeyMsg
func NewKeyMsgFromKeyEvent(keyEvent *KeyEvent) *KeyMsg {
	if keyEvent == nil {
		return nil
	}

	return &KeyMsg{
		BaseMsg: BaseMsg{
			TypeValue:   MsgTypeKey,
			TimestampValue: time.Now(),
		},
		Rune:    keyEvent.Key.Rune,
		Special: keyEvent.Special,
		Mod: Modifiers{
			Alt:   keyEvent.Key.Alt,
			Ctrl:  keyEvent.Key.Ctrl,
			Shift: keyEvent.Key.Shift,
		},
	}
}

// IsPrintable 检查是否为可打印字符
// 可打印字符是指可以直接显示的字符（字母、数字、符号等）
func (k *KeyMsg) IsPrintable() bool {
	// 如果有特殊键，不是可打印字符
	if k.Special != runtimeplatform.KeyUnknown && k.Special != runtimeplatform.KeyTab {
		return false
	}

	// 检查 rune 是否在可打印范围内
	// ASCII 可打印字符: 32-126
	// Unicode 可扩展这个范围
	return k.Rune >= 32 && k.Rune <= 126
}

// HasModifier 检查是否有修饰键
func (k *KeyMsg) HasModifier() bool {
	return k.Mod.Alt || k.Mod.Ctrl || k.Mod.Shift
}

// HasCtrl 检查是否有 Ctrl 键
func (k *KeyMsg) HasCtrl() bool {
	return k.Mod.Ctrl
}

// HasAlt 检查是否有 Alt 键
func (k *KeyMsg) HasAlt() bool {
	return k.Mod.Alt
}

// HasShift 检查是否有 Shift 键
func (k *KeyMsg) HasShift() bool {
	return k.Mod.Shift
}

// IsEnter 检查是否为 Enter 键
func (k *KeyMsg) IsEnter() bool {
	return k.Special == runtimeplatform.KeyEnter
}

// IsTab 检查是否为 Tab 键
func (k *KeyMsg) IsTab() bool {
	return k.Special == runtimeplatform.KeyTab
}

// IsEscape 检查是否为 Escape 键
func (k *KeyMsg) IsEscape() bool {
	return k.Special == runtimeplatform.KeyEscape
}

// IsBackspace 检查是否为 Backspace 键
func (k *KeyMsg) IsBackspace() bool {
	return k.Special == runtimeplatform.KeyBackspace
}

// IsDelete 检查是否为 Delete 键
func (k *KeyMsg) IsDelete() bool {
	return k.Special == runtimeplatform.KeyDelete
}

// IsNavigation 检查是否为导航键
func (k *KeyMsg) IsNavigation() bool {
	switch k.Special {
	case runtimeplatform.KeyUp, runtimeplatform.KeyDown,
	     runtimeplatform.KeyLeft, runtimeplatform.KeyRight,
	     runtimeplatform.KeyHome, runtimeplatform.KeyEnd,
	     runtimeplatform.KeyPageUp, runtimeplatform.KeyPageDown:
		return true
	default:
		return false
	}
}

// IsFunctionKey 检查是否为功能键（F1-F12）
func (k *KeyMsg) IsFunctionKey() bool {
	return k.Special >= runtimeplatform.KeyF1 && k.Special <= runtimeplatform.KeyF12
}

// String 返回 KeyMsg 的字符串表示
func (k *KeyMsg) String() string {
	var parts []string

	// 添加修饰键
	if k.Mod.Ctrl {
		parts = append(parts, "Ctrl")
	}
	if k.Mod.Alt {
		parts = append(parts, "Alt")
	}
	if k.Mod.Shift {
		parts = append(parts, "Shift")
	}

	// 添加按键
	if k.Special != runtimeplatform.KeyUnknown {
		parts = append(parts, k.specialKeyString())
	} else if k.Rune != 0 {
		parts = append(parts, fmt.Sprintf("'%c'", k.Rune))
	}

	if len(parts) == 0 {
		return "KeyMsg{<unknown>}"
	}

	return "KeyMsg{" + strings.Join(parts, "+") + "}"
}

// specialKeyString 返回特殊键的字符串表示
func (k *KeyMsg) specialKeyString() string {
	switch k.Special {
	case runtimeplatform.KeyUp:
		return "Up"
	case runtimeplatform.KeyDown:
		return "Down"
	case runtimeplatform.KeyLeft:
		return "Left"
	case runtimeplatform.KeyRight:
		return "Right"
	case runtimeplatform.KeyEnter:
		return "Enter"
	case runtimeplatform.KeyTab:
		return "Tab"
	case runtimeplatform.KeyBackspace:
		return "Backspace"
	case runtimeplatform.KeyDelete:
		return "Delete"
	case runtimeplatform.KeyEscape:
		return "Escape"
	case runtimeplatform.KeyHome:
		return "Home"
	case runtimeplatform.KeyEnd:
		return "End"
	case runtimeplatform.KeyPageUp:
		return "PageUp"
	case runtimeplatform.KeyPageDown:
		return "PageDown"
	case runtimeplatform.KeyF1:
		return "F1"
	case runtimeplatform.KeyF2:
		return "F2"
	case runtimeplatform.KeyF3:
		return "F3"
	case runtimeplatform.KeyF4:
		return "F4"
	case runtimeplatform.KeyF5:
		return "F5"
	case runtimeplatform.KeyF6:
		return "F6"
	case runtimeplatform.KeyF7:
		return "F7"
	case runtimeplatform.KeyF8:
		return "F8"
	case runtimeplatform.KeyF9:
		return "F9"
	case runtimeplatform.KeyF10:
		return "F10"
	case runtimeplatform.KeyF11:
		return "F11"
	case runtimeplatform.KeyF12:
		return "F12"
	default:
		return fmt.Sprintf("Special(%d)", k.Special)
	}
}

// KeyEvent 是兼容 runtime.KeyEvent 的结构
// 这个结构用于适配器从 runtime.Event 转换
type KeyEvent struct {
	Key     Key
	Special runtimeplatform.SpecialKey
}

// Key 表示一个按键
type Key struct {
	Rune  rune
	Alt   bool
	Ctrl  bool
	Shift bool
}
