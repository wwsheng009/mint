package msg

import (
	"fmt"
	"time"

	runtimeevent "github.com/wwsheng009/mint/runtime/event"
)

// ToMsg 将 runtime.Event 转换为 Msg
//
// ToMsg 是 Event 到 Msg 的适配器函数，它将运行时的事件转换为
// 应用层的消息，实现 Event → Msg → Action 的流程。
func ToMsg(event runtimeevent.Event) Msg {
	if event == nil {
		return nil
	}

	switch e := event.(type) {
	case *runtimeevent.KeyEvent:
		return keyEventToMsg(e)

	case *runtimeevent.MouseEvent:
		return mouseEventToMsg(e)

	case *runtimeevent.ResizeEvent:
		return resizeEventToMsg(e)

	default:
		return NewBaseMsg(MsgTypeUnknown)
	}
}

// keyEventToMsg 将 KeyEvent 转换为 KeyMsg
func keyEventToMsg(keyEvent *runtimeevent.KeyEvent) Msg {
	if keyEvent == nil {
		return nil
	}

	return &KeyMsg{
		BaseMsg: BaseMsg{
			TypeValue:   MsgTypeKey,
			TimestampValue: time.Now(),
		},
		Rune: keyEvent.Key.Rune,
		Special: keyEvent.Special,
		Mod: Modifiers{
			Alt:   keyEvent.Key.Alt,
			Ctrl:  keyEvent.Key.Ctrl,
			Shift: keyEvent.Key.Shift,
		},
	}
}

// mouseEventToMsg 将 MouseEvent 转换为 MouseMsg
func mouseEventToMsg(mouseEvent *runtimeevent.MouseEvent) Msg {
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
	switch mouseEvent.Type {
	case runtimeevent.EventMousePress:
		action = MouseActionPress
	case runtimeevent.EventMouseRelease:
		action = MouseActionRelease
	case runtimeevent.EventMouseMove:
		action = MouseActionMove
	case runtimeevent.EventMouseWheel:
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
		LocalX:   0, // TODO: 从 HitMap 获取
		LocalY:   0, // TODO: 从 HitMap 获取
		TargetID: "", // TODO: 从 HitMap 获取
		Button:   button,
		Action:   action,
	}
}

// resizeEventToMsg 将 ResizeEvent 转换为 ResizeMsg
func resizeEventToMsg(resizeEvent *runtimeevent.ResizeEvent) Msg {
	if resizeEvent == nil {
		return nil
	}

	return &ResizeMsg{
		BaseMsg: BaseMsg{
			TypeValue:   MsgTypeResize,
			TimestampValue: time.Now(),
		},
		Width:  resizeEvent.Width,
		Height: resizeEvent.Height,
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
			TypeValue:   MsgTypeResize,
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
