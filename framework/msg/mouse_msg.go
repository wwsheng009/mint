package msg

import (
	"fmt"
	"time"

	runtimeevent "github.com/wwsheng009/mint/runtime/event"
)

// MouseMsg 表示鼠标输入消息
//
// MouseMsg 封装了鼠标事件，包括：
// - 鼠标位置（全局和本地坐标）
// - 鼠标按钮（左、中、右）
// - 鼠标动作（按下、释放、移动、滚轮）
// - 目标组件信息（来自 HitMap）
type MouseMsg struct {
	BaseMsg

	// X, Y 是鼠标的全局坐标
	X, Y int

	// LocalX, LocalY 是相对于目标组件的本地坐标
	LocalX, LocalY int

	// TargetID 是目标组件的 ID（来自 HitMap）
	TargetID string

	// Button 是按下的鼠标按钮
	Button MouseButton

	// Action 是鼠标动作类型
	Action MouseAction
}

// MouseButton 表示鼠标按钮
type MouseButton int

const (
	MouseButtonUnknown MouseButton = iota
	MouseLeft                    // 左键
	MouseMiddle                  // 中键
	MouseRight                   // 右键
)

// MouseAction 表示鼠标动作
type MouseAction int

const (
	MouseActionUnknown MouseAction = iota
	MouseActionPress              // 按下
	MouseActionRelease            // 释放
	MouseActionMove               // 移动
	MouseActionWheel              // 滚轮
)

// NewMouseMsg 创建一个新的鼠标消息
func NewMouseMsg(x, y int, button MouseButton, action MouseAction) *MouseMsg {
	return &MouseMsg{
		BaseMsg: BaseMsg{
			TypeValue:   MsgTypeMouse,
			TimestampValue: time.Now(),
		},
		X:      x,
		Y:      y,
		Button: button,
		Action: action,
	}
}

// NewMouseMsgFromMouseEvent 从 runtime MouseEvent 创建 MouseMsg
func NewMouseMsgFromMouseEvent(mouseEvent *runtimeevent.MouseEvent) *MouseMsg {
	if mouseEvent == nil {
		return nil
	}

	// 转换 MouseButton
	var button MouseButton
	switch mouseEvent.Button {
	case runtimeevent.MouseLeft:
		button = MouseLeft
	case runtimeevent.MouseMiddle:
		button = MouseMiddle
	case runtimeevent.MouseRight:
		button = MouseRight
	default:
		button = MouseButtonUnknown
	}

	// 转换 MouseAction
	var action MouseAction
	switch mouseEvent.Action {
	case runtimeevent.MouseActionPress:
		action = MouseActionPress
	case runtimeevent.MouseActionRelease:
		action = MouseActionRelease
	case runtimeevent.MouseActionMove:
		action = MouseActionMove
	case runtimeevent.MouseActionWheel:
		action = MouseActionWheel
	default:
		action = MouseActionUnknown
	}

	return &MouseMsg{
		BaseMsg: BaseMsg{
			TypeValue:   MsgTypeMouse,
			TimestampValue: time.Now(),
		},
		X:        mouseEvent.X,
		Y:        mouseEvent.Y,
		LocalX:   mouseEvent.LocalX,
		LocalY:   mouseEvent.LocalY,
		TargetID: mouseEvent.TargetID,
		Button:   button,
		Action:   action,
	}
}

// IsClick 检查是否为点击事件（左键按下）
func (m *MouseMsg) IsClick() bool {
	return m.Action == MouseActionPress && m.Button == MouseLeft
}

// IsRightClick 检查是否为右键点击
func (m *MouseMsg) IsRightClick() bool {
	return m.Action == MouseActionPress && m.Button == MouseRight
}

// IsMiddleClick 检查是否为中键点击
func (m *MouseMsg) IsMiddleClick() bool {
	return m.Action == MouseActionPress && m.Button == MouseMiddle
}

// IsScroll 检查是否为滚轮事件
func (m *MouseMsg) IsScroll() bool {
	return m.Action == MouseActionWheel
}

// IsMove 检查是否为鼠标移动事件
func (m *MouseMsg) IsMove() bool {
	return m.Action == MouseActionMove
}

// IsPress 检查是否为鼠标按下事件
func (m *MouseMsg) IsPress() bool {
	return m.Action == MouseActionPress
}

// IsRelease 检查是否为鼠标释放事件
func (m *MouseMsg) IsRelease() bool {
	return m.Action == MouseActionRelease
}

// HasTarget 检查是否有目标组件（通过 HitMap 命中）
func (m *MouseMsg) HasTarget() bool {
	return m.TargetID != ""
}

// GetPosition 返回鼠标的全局位置
func (m *MouseMsg) GetPosition() (x, y int) {
	return m.X, m.Y
}

// GetLocalPosition 返回相对于目标的本地位置
func (m *MouseMsg) GetLocalPosition() (x, y int) {
	return m.LocalX, m.LocalY
}

// String 返回 MouseMsg 的字符串表示
func (m *MouseMsg) String() string {
	return fmt.Sprintf("MouseMsg{%s %s at (%d,%d) local(%d,%d) target=%s}",
		m.buttonString(),
		m.actionString(),
		m.X, m.Y,
		m.LocalX, m.LocalY,
		m.TargetID,
	)
}

// buttonString 返回按钮的字符串表示
func (m *MouseMsg) buttonString() string {
	switch m.Button {
	case MouseLeft:
		return "Left"
	case MouseMiddle:
		return "Middle"
	case MouseRight:
		return "Right"
	default:
		return "Unknown"
	}
}

// actionString 返回动作的字符串表示
func (m *MouseMsg) actionString() string {
	switch m.Action {
	case MouseActionPress:
		return "Press"
	case MouseActionRelease:
		return "Release"
	case MouseActionMove:
		return "Move"
	case MouseActionWheel:
		return "Wheel"
	default:
		return "Unknown"
	}
}
