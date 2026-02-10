package msg

import (
	"fmt"
	"time"

	runtimeevent "github.com/wwsheng009/mint/runtime/event"
)

// KeyEventToMsg 将 KeyEvent 转换为 KeyMsg
func KeyEventToMsg(keyEvent *runtimeevent.KeyEvent) Msg {
	if keyEvent == nil {
		return nil
	}

	return &KeyMsg{
		BaseMsg: BaseMsg{
			TypeValue:       MsgTypeKey,
			TimestampValue: time.Now(),
		},
		Rune:    keyEvent.Key,
		Special: keyEvent.Special,
		Mod: Modifiers{
			Alt:   keyEvent.Mod&runtimeevent.ModAlt != 0,
			Ctrl:  keyEvent.Mod&runtimeevent.ModCtrl != 0,
			Shift: keyEvent.Mod&runtimeevent.ModShift != 0,
		},
	}
}

// MouseEventToMsg 将 MouseEvent 转换为 MouseMsg
func MouseEventToMsg(mouseEvent *runtimeevent.MouseEvent) Msg {
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
		// Fallback to Type field for backward compatibility
		switch mouseEvent.Type {
		case "press":
			action = MouseActionPress
		case "release":
			action = MouseActionRelease
		case "move", "motion":
			action = MouseActionMove
		case "wheel":
			action = MouseActionWheel
		default:
			action = MouseActionUnknown
		}
	}

	return &MouseMsg{
		BaseMsg: BaseMsg{
			TypeValue:       MsgTypeMouse,
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

// ResizeMsg 表示窗口大小改变消息
type ResizeMsg struct {
	BaseMsg

	// Width 是新的窗口宽度
	Width int

	// Height 是新的窗口高度
	Height int
}

// NewResizeMsg 创建一个 Resize 消息
func NewResizeMsg(width, height int) *ResizeMsg {
	return &ResizeMsg{
		BaseMsg: BaseMsg{
			TypeValue:       MsgTypeResize,
			TimestampValue: time.Now(),
		},
		Width:  width,
		Height: height,
	}
}

// String 返回 ResizeMsg 的字符串表示
func (r *ResizeMsg) String() string {
	return fmt.Sprintf("ResizeMsg{%dx%d}", r.Width, r.Height)
}
