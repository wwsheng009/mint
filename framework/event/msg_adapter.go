package event

import (
	runtimeplatform "github.com/wwsheng009/mint/runtime/platform"
	runtimeevent "github.com/wwsheng009/mint/runtime/event"
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
)

// MsgToEvent converts a runtime Msg to a framework Event.
// This is a temporary adapter during the Msg/Cmd migration.
//
// TODO: Remove this once app.go is fully migrated to Msg.
func MsgToEvent(msg runtimemsg.Msg) Event {
	if msg == nil {
		return nil
	}

	switch m := msg.(type) {
	case *runtimemsg.KeyMsg:
		return keyMsgToKeyEvent(m)
	case *runtimemsg.MouseMsg:
		return mouseMsgToMouseEvent(m)
	case *runtimemsg.BaseMsg:
		// Handle generic messages like Resize
		if m.Type() == runtimemsg.MsgTypeResize {
			return &ResizeEvent{
				BaseEvent: NewBaseEvent(EventResize),
			}
		}
	}

	return nil
}

// keyMsgToKeyEvent converts a KeyMsg to KeyEvent.
func keyMsgToKeyEvent(keyMsg *runtimemsg.KeyMsg) *KeyEvent {
	// Build the Key structure
	key := Key{
		Rune:  keyMsg.Rune,
		Name:  specialKeyToName(keyMsg.Special),
		Alt:   keyMsg.Mod.Alt,
		Ctrl:  keyMsg.Mod.Ctrl,
		Shift: keyMsg.Mod.Shift,
	}

	// Convert runtime SpecialKey to framework SpecialKey
	frameworkSpecial := runtimeToFrameworkSpecialKey(keyMsg.Special)

	// Create KeyEvent
	ev := &KeyEvent{
		BaseEvent: NewBaseEvent(EventKeyPress),
		Key:       key,
		Special:   frameworkSpecial,
	}

	// Set modifiers
	if keyMsg.Mod.Ctrl {
		ev.Modifiers |= ModCtrl
	}
	if keyMsg.Mod.Alt {
		ev.Modifiers |= ModAlt
	}
	if keyMsg.Mod.Shift {
		ev.Modifiers |= ModShift
	}

	return ev
}

// mouseMsgToMouseEvent converts a MouseMsg to MouseEvent.
func mouseMsgToMouseEvent(mouseMsg *runtimemsg.MouseMsg) *MouseEvent {
	// Convert mouse button
	var button MouseButton
	switch mouseMsg.Button {
	case runtimemsg.MouseLeft:
		button = MouseLeft
	case runtimemsg.MouseMiddle:
		button = MouseMiddle
	case runtimemsg.MouseRight:
		button = MouseRight
	default:
		button = MouseNone
	}

	// Convert mouse action - runtime/msg and runtime/event use the same MouseAction type
	var action runtimeevent.MouseAction
	switch mouseMsg.Action {
	case runtimemsg.MouseActionPress:
		action = runtimeevent.MouseActionPress
	case runtimemsg.MouseActionRelease:
		action = runtimeevent.MouseActionRelease
	case runtimemsg.MouseActionMove:
		action = runtimeevent.MouseActionMove
	case runtimemsg.MouseActionWheel:
		action = runtimeevent.MouseActionWheel
	default:
		// No unknown constant, just use Press as default
		action = runtimeevent.MouseActionPress
	}

	// Determine EventType based on MouseAction
	var eventType EventType
	switch action {
	case runtimeevent.MouseActionPress:
		eventType = EventMousePress
	case runtimeevent.MouseActionRelease:
		eventType = EventMouseRelease
	case runtimeevent.MouseActionMove:
		eventType = EventMouseMove
	case runtimeevent.MouseActionWheel:
		eventType = EventMouseWheel
	}

	// Create MouseEvent
	ev := &MouseEvent{
		BaseEvent: NewBaseEvent(eventType),
		X:         mouseMsg.X,
		Y:         mouseMsg.Y,
		Button:    button,
		Action:    action,
		TargetID:  mouseMsg.TargetID,
		LocalX:    mouseMsg.LocalX,
		LocalY:    mouseMsg.LocalY,
		Delta:     mouseMsg.Delta,
	}

	return ev
}

// runtimeToFrameworkSpecialKey converts runtime platform SpecialKey to framework SpecialKey.
func runtimeToFrameworkSpecialKey(special runtimeplatform.SpecialKey) SpecialKey {
	switch special {
	case runtimeplatform.KeyUp:
		return KeyUp
	case runtimeplatform.KeyDown:
		return KeyDown
	case runtimeplatform.KeyLeft:
		return KeyLeft
	case runtimeplatform.KeyRight:
		return KeyRight
	case runtimeplatform.KeyEnter:
		return KeyEnter
	case runtimeplatform.KeyTab:
		return KeyTab
	case runtimeplatform.KeyBackspace:
		return KeyBackspace
	case runtimeplatform.KeyDelete:
		return KeyDelete
	case runtimeplatform.KeyEscape:
		return KeyEscape
	case runtimeplatform.KeyHome:
		return KeyHome
	case runtimeplatform.KeyEnd:
		return KeyEnd
	case runtimeplatform.KeyPageUp:
		return KeyPageUp
	case runtimeplatform.KeyPageDown:
		return KeyPageDown
	case runtimeplatform.KeyF1:
		return KeyF1
	case runtimeplatform.KeyF2:
		return KeyF2
	case runtimeplatform.KeyF3:
		return KeyF3
	case runtimeplatform.KeyF4:
		return KeyF4
	case runtimeplatform.KeyF5:
		return KeyF5
	case runtimeplatform.KeyF6:
		return KeyF6
	case runtimeplatform.KeyF7:
		return KeyF7
	case runtimeplatform.KeyF8:
		return KeyF8
	case runtimeplatform.KeyF9:
		return KeyF9
	case runtimeplatform.KeyF10:
		return KeyF10
	case runtimeplatform.KeyF11:
		return KeyF11
	case runtimeplatform.KeyF12:
		return KeyF12
	default:
		return KeyUnknown
	}
}

// specialKeyToName converts runtime platform SpecialKey to framework key name.
func specialKeyToName(special runtimeplatform.SpecialKey) string {
	switch special {
	case runtimeplatform.KeyUp:
		return "up"
	case runtimeplatform.KeyDown:
		return "down"
	case runtimeplatform.KeyLeft:
		return "left"
	case runtimeplatform.KeyRight:
		return "right"
	case runtimeplatform.KeyEnter:
		return "enter"
	case runtimeplatform.KeyTab:
		return "tab"
	case runtimeplatform.KeyBackspace:
		return "backspace"
	case runtimeplatform.KeyDelete:
		return "delete"
	case runtimeplatform.KeyEscape:
		return "escape"
	case runtimeplatform.KeyHome:
		return "home"
	case runtimeplatform.KeyEnd:
		return "end"
	case runtimeplatform.KeyPageUp:
		return "page up"
	case runtimeplatform.KeyPageDown:
		return "page down"
	case runtimeplatform.KeyF1:
		return "f1"
	case runtimeplatform.KeyF2:
		return "f2"
	case runtimeplatform.KeyF3:
		return "f3"
	case runtimeplatform.KeyF4:
		return "f4"
	case runtimeplatform.KeyF5:
		return "f5"
	case runtimeplatform.KeyF6:
		return "f6"
	case runtimeplatform.KeyF7:
		return "f7"
	case runtimeplatform.KeyF8:
		return "f8"
	case runtimeplatform.KeyF9:
		return "f9"
	case runtimeplatform.KeyF10:
		return "f10"
	case runtimeplatform.KeyF11:
		return "f11"
	case runtimeplatform.KeyF12:
		return "f12"
	default:
		return ""
	}
}
