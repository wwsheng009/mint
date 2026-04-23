package msg

import (
	"fmt"
	"strings"
	"time"
	"unicode"

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
			TypeValue:      MsgTypeKey,
			TimestampValue: time.Now(),
		},
		Rune:    rune,
		Special: special,
		Mod:     mod,
	}
}

// IsPrintable 检查是否为可打印字符
// 可打印字符是指可以直接显示的字符（字母、数字、符号等）
// 支持 Unicode 字符（包括中文、日文、韩文、emoji 等）
func (k *KeyMsg) IsPrintable() bool {
	// 如果有特殊键，不是可打印字符
	if k.Special != runtimeplatform.KeyUnknown && k.Special != runtimeplatform.KeyTab {
		return false
	}

	// 使用 unicode.IsPrint 判断是否为可打印字符
	// 这支持所有 Unicode 可打印字符，包括中文、日文、韩文、emoji 等
	return unicode.IsPrint(k.Rune)
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

// IsF1 检查是否为 F1 键
func (k *KeyMsg) IsF1() bool {
	return k.Special == runtimeplatform.KeyF1
}

// IsF2 检查是否为 F2 键
func (k *KeyMsg) IsF2() bool {
	return k.Special == runtimeplatform.KeyF2
}

// IsF3 检查是否为 F3 键
func (k *KeyMsg) IsF3() bool {
	return k.Special == runtimeplatform.KeyF3
}

// IsF4 检查是否为 F4 键
func (k *KeyMsg) IsF4() bool {
	return k.Special == runtimeplatform.KeyF4
}

// IsF5 检查是否为 F5 键
func (k *KeyMsg) IsF5() bool {
	return k.Special == runtimeplatform.KeyF5
}

// IsF6 检查是否为 F6 键
func (k *KeyMsg) IsF6() bool {
	return k.Special == runtimeplatform.KeyF6
}

// IsF7 检查是否为 F7 键
func (k *KeyMsg) IsF7() bool {
	return k.Special == runtimeplatform.KeyF7
}

// IsF8 检查是否为 F8 键
func (k *KeyMsg) IsF8() bool {
	return k.Special == runtimeplatform.KeyF8
}

// IsF9 检查是否为 F9 键
func (k *KeyMsg) IsF9() bool {
	return k.Special == runtimeplatform.KeyF9
}

// IsF10 检查是否为 F10 键
func (k *KeyMsg) IsF10() bool {
	return k.Special == runtimeplatform.KeyF10
}

// IsF11 检查是否为 F11 键
func (k *KeyMsg) IsF11() bool {
	return k.Special == runtimeplatform.KeyF11
}

// IsF12 检查是否为 F12 键
func (k *KeyMsg) IsF12() bool {
	return k.Special == runtimeplatform.KeyF12
}

// IsUp 检查是否为向上键
func (k *KeyMsg) IsUp() bool {
	return k.Special == runtimeplatform.KeyUp
}

// IsDown 检查是否为向下键
func (k *KeyMsg) IsDown() bool {
	return k.Special == runtimeplatform.KeyDown
}

// IsLeft 检查是否为向左键
func (k *KeyMsg) IsLeft() bool {
	return k.Special == runtimeplatform.KeyLeft
}

// IsRight 检查是否为向右键
func (k *KeyMsg) IsRight() bool {
	return k.Special == runtimeplatform.KeyRight
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
