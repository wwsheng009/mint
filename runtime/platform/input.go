package platform

import (
	"fmt"
	"time"
)

// InputReader 输入读取抽象 (V3: 从 Terminal 拆分)
// Platform 只产生 RawInput，不产生语义化的 Action
// Action 转换由 Runtime 的 KeyMap 负责
type InputReader interface {
	// 读取单个输入
	ReadEvent() (RawInput, error)

	// 启动读取循环
	Start(events chan<- RawInput) error

	// 停止读取
	Stop() error
}

// RawInput 原始输入 (平台无关的表示)
type RawInput struct {
	Type RawInputType

	// 键盘
	Key      rune
	Special  SpecialKey
	Modifiers KeyModifier

	// 鼠标
	MouseX      int
	MouseY      int
	MouseButton MouseButton
	MouseAction MouseAction

	// 窗口大小
	Width     int
	Height    int

	// 其他
	Data     []byte
	Timestamp time.Time
}

// RawInputType 输入类型
type RawInputType int

const (
	InputKeyPress RawInputType = iota
	InputKeyRelease
	InputMouse
	InputResize
	InputPaste
	InputSignal
)

// SpecialKey 特殊键
type SpecialKey int

const (
	KeyUnknown SpecialKey = iota

	// 控制键
	KeyEscape
	KeyEnter
	KeyTab
	KeyBackspace
	KeyDelete
	KeyInsert
	KeyAlt
	KeyCtrl
	KeyShift
	KeyMeta

	// 光标键
	KeyUp
	KeyDown
	KeyLeft
	KeyRight
	KeyHome
	KeyEnd
	KeyPageUp
	KeyPageDown

	// 功能键
	KeyF1
	KeyF2
	KeyF3
	KeyF4
	KeyF5
	KeyF6
	KeyF7
	KeyF8
	KeyF9
	KeyF10
	KeyF11
	KeyF12

	// 组合键
	KeySpace

	// Vim 风格
	KeyK // vim up
	KeyJ // vim down
	KeyH // vim left
	KeyL // vim right
)

// String returns the name of the special key.
func (k SpecialKey) String() string {
	switch k {
	case KeyUnknown:
		return "Unknown"
	case KeyEscape:
		return "Escape"
	case KeyEnter:
		return "Enter"
	case KeyTab:
		return "Tab"
	case KeyBackspace:
		return "Backspace"
	case KeyDelete:
		return "Delete"
	case KeyInsert:
		return "Insert"
	case KeyAlt:
		return "Alt"
	case KeyCtrl:
		return "Ctrl"
	case KeyShift:
		return "Shift"
	case KeyMeta:
		return "Meta"
	case KeyUp:
		return "Up"
	case KeyDown:
		return "Down"
	case KeyLeft:
		return "Left"
	case KeyRight:
		return "Right"
	case KeyHome:
		return "Home"
	case KeyEnd:
		return "End"
	case KeyPageUp:
		return "PageUp"
	case KeyPageDown:
		return "PageDown"
	case KeyF1:
		return "F1"
	case KeyF2:
		return "F2"
	case KeyF3:
		return "F3"
	case KeyF4:
		return "F4"
	case KeyF5:
		return "F5"
	case KeyF6:
		return "F6"
	case KeyF7:
		return "F7"
	case KeyF8:
		return "F8"
	case KeyF9:
		return "F9"
	case KeyF10:
		return "F10"
	case KeyF11:
		return "F11"
	case KeyF12:
		return "F12"
	case KeySpace:
		return "Space"
	case KeyK:
		return "K"
	case KeyJ:
		return "J"
	case KeyH:
		return "H"
	case KeyL:
		return "L"
	default:
		return fmt.Sprintf("SpecialKey(%d)", k)
	}
}

// KeyModifier 修饰键
type KeyModifier uint8

const (
	ModShift      KeyModifier = 1 << iota // 1 (二进制: 0001)
	ModCtrl                               // 2 (二进制: 0010)
	ModAlt                                // 4 (二进制: 0100)
	ModMeta                         // 8 (二进制: 1000)
)

// MouseButton 鼠标按钮
type MouseButton int

const (
	MouseNone MouseButton = iota
	MouseLeft
	MouseMiddle
	MouseRight
)

// MouseAction 鼠标动作
type MouseAction int

const (
	MousePress MouseAction = iota
	MouseRelease
	MouseMotion
	MouseWheelUp
	MouseWheelDown
)

// NewInputReader 创建平台特定的输入读取器
func NewInputReader() (InputReader, error) {
	return newPlatformInputReader()
}

// newPlatformInputReader 根据平台创建输入读取器
func newPlatformInputReader() (InputReader, error) {
	return &defaultInputReaderWrapper{impl: newInputReaderImpl()}, nil
}

// defaultInputReaderWrapper 默认包装器，使用平台特定实现
type defaultInputReaderWrapper struct {
	impl inputReaderImpl
}

// Start 启动读取循环
func (w *defaultInputReaderWrapper) Start(events chan<- RawInput) error {
	if w.impl == nil {
		w.impl = newInputReaderImpl()
	}
	return w.impl.Start(events)
}

// Stop 停止读取
func (w *defaultInputReaderWrapper) Stop() error {
	if w.impl != nil {
		return w.impl.Stop()
	}
	return nil
}

// ReadEvent 读取单个事件
func (w *defaultInputReaderWrapper) ReadEvent() (RawInput, error) {
	if w.impl == nil {
		return RawInput{}, fmt.Errorf("input reader not started")
	}
	return w.impl.ReadEvent()
}

// RestoreTerminal 恢复终端到正常模式
//
// 🔥 这是应用层的恢复机制（多层防御系统的第 2 层）
//
// 恢复顺序（从内到外）：
//
// 1. Engine 层（第 1 层）：Engine.Run() 的 defer cleanup() → inputReader.Stop() → 恢复 originalMode
// 2. 应用层（第 2 层，这里）：main() 的 defer RestoreTerminal() → 强制恢复到安全模式（保险）
// 3. 进程层（第 3 层）：init() 的信号处理 → Ctrl+C 时强制恢复（兜底）
//
// 为什么需要多层防御？
//
// 终端模式污染是致命问题，一次污染就会导致 shell 永久损坏。
// 多层防御确保即使某一层失败，其他层仍能恢复终端。
//
// 使用示例：
//
//	func main() {
//	    defer platform.RestoreTerminal()  // 第 2 层防御
//	    // ... 你的代码
//	}
//
// 注意事项：
// - 这个函数会直接恢复到 Windows 安全模式，不依赖任何内部状态
// - 性能开销极小（只调用一次系统 API），但安全性提升显著
// - 即使 Engine 的恢复有 bug，这里也能兜底
//
// 工业级 TUI 程序的最佳实践：宁可过度保护，不可终端污染。
func RestoreTerminal() {
	restoreTerminalImpl()
}

// inputReaderImpl 平台特定实现接口
type inputReaderImpl interface {
	Start(events chan<- RawInput) error
	Stop() error
	ReadEvent() (RawInput, error)
}
