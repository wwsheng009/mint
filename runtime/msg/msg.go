package msg

import (
	"fmt"
	"time"
)

// Msg 是 TUI 应用中消息的统一接口
//
// Msg 代表应用中的各种事件和状态变化，类似于 Elm Architecture 或 Bubble Tea 的 Msg 概念。
// 所有进入应用的消息都应该实现这个接口。
//
// Msg 是不可变的，应该通过值传递而不是指针传递。
type Msg interface {
	// Type 返回消息的类型
	// 用于快速判断消息类别而不需要类型断言
	Type() MsgType

	// Timestamp 返回消息创建的时间
	Timestamp() time.Time

	// String 返回消息的字符串表示（用于调试）
	String() string
}

// MsgType 表示消息的类型
type MsgType string

const (
	// MsgTypeUnknown 表示未知类型的消息
	MsgTypeUnknown MsgType = "unknown"

	// ============================================================================
	// 输入类消息
	// ============================================================================

	// MsgTypeKey 键盘输入消息
	MsgTypeKey MsgType = "key"

	// MsgTypeMouse 鼠标输入消息
	MsgTypeMouse MsgType = "mouse"

	// MsgTypePaste 粘贴输入消息，携带完整粘贴文本
	MsgTypePaste MsgType = "paste"

	// ============================================================================
	// 系统类消息
	// ============================================================================

	// MsgTypeResize 窗口大小改变消息
	MsgTypeResize MsgType = "resize"

	// MsgTypeQuit 退出应用消息
	MsgTypeQuit MsgType = "quit"

	// MsgTypeTick 定时器消息
	MsgTypeTick MsgType = "tick"

	// ============================================================================
	// 组件类消息
	// ============================================================================

	// MsgTypeAction 组件操作消息（通过 ActionTarget 处理）
	MsgTypeAction MsgType = "action"

	// MsgTypeState 状态变化消息
	MsgTypeState MsgType = "state"

	// ============================================================================
	// 测试类消息
	// ============================================================================

	// MsgTypeSandbox 测试沙箱注入消息
	MsgTypeSandbox MsgType = "sandbox"
)

// BaseMsg 提供了 Msg 接口的基础实现
// 其他 Msg 类型可以嵌入 BaseMsg 来自动实现通用方法
type BaseMsg struct {
	TypeValue      MsgType
	TimestampValue time.Time
}

// Type 返回消息类型
func (m *BaseMsg) Type() MsgType {
	return m.TypeValue
}

// Timestamp 返回消息创建时间
func (m *BaseMsg) Timestamp() time.Time {
	return m.TimestampValue
}

// String 返回基础字符串表示
// 具体的 Msg 类型应该覆盖这个方法以提供更有用的信息
func (m *BaseMsg) String() string {
	return fmt.Sprintf("Msg{Type=%s, Time=%s}", m.TypeValue, m.TimestampValue.Format(time.RFC3339))
}

// NewBaseMsg 创建一个新的基础消息
func NewBaseMsg(msgType MsgType) *BaseMsg {
	return &BaseMsg{
		TypeValue:      msgType,
		TimestampValue: time.Now(),
	}
}

// ============================================================================
// 辅助函数
// ============================================================================

// IsInputMsg 检查消息是否为输入类消息（键盘或鼠标）
func IsInputMsg(m Msg) bool {
	switch m.Type() {
	case MsgTypeKey, MsgTypeMouse, MsgTypePaste:
		return true
	default:
		return false
	}
}

// IsSystemMsg 检查消息是否为系统类消息
func IsSystemMsg(m Msg) bool {
	switch m.Type() {
	case MsgTypeResize, MsgTypeQuit, MsgTypeTick:
		return true
	default:
		return false
	}
}

// IsComponentMsg 检查消息是否为组件类消息
func IsComponentMsg(m Msg) bool {
	switch m.Type() {
	case MsgTypeAction, MsgTypeState:
		return true
	default:
		return false
	}
}

// IsTestMsg 检查消息是否为测试类消息
func IsTestMsg(m Msg) bool {
	return m.Type() == MsgTypeSandbox
}

// MsgAge 返回消息的年龄（从创建到现在的时间）
func MsgAge(m Msg) time.Duration {
	return time.Since(m.Timestamp())
}

// FormatMsg 格式化消息用于日志输出
func FormatMsg(m Msg) string {
	if m == nil {
		return "Msg{nil}"
	}
	return fmt.Sprintf("[%s] %s", m.Type(), m.String())
}
