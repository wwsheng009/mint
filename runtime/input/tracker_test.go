package input

import (
	"testing"
	"time"

	runtimeplatform "github.com/wwsheng009/mint/runtime/platform"
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
)

func TestInputTracker_MousePressRelease(t *testing.T) {
	tracker := NewInputTracker()

	// 初始化快照（位置设置为 10,10，无按钮）
	initSnap := &InputSnapshot{
		MouseX:       10,
		MouseY:       10,
		MouseButton:  runtimemsg.MouseButtonUnknown,
		MouseAction:  runtimemsg.MouseActionUnknown,
		Timestamp:    time.Now().UnixNano(),
	}
	intents := tracker.Update(initSnap)

	// 按下
	pressSnap := &InputSnapshot{
		MouseX:       10,
		MouseY:       10,
		MouseButton:  runtimemsg.MouseLeft,
		MouseAction:  runtimemsg.MouseActionPress,
		Timestamp:    time.Now().UnixNano(),
	}
	intents = tracker.Update(pressSnap)

	if len(intents) != 1 {
		t.Errorf("Expected 1 intent on press, got %d", len(intents))
	} else {
		pressIntent, ok := intents[0].(InputPressIntent)
		if !ok {
			t.Error("Expected InputPressIntent, got different type")
		} else {
			if pressIntent.X != 10 || pressIntent.Y != 10 {
				t.Errorf("Expected position (10,10), got (%d,%d)", pressIntent.X, pressIntent.Y)
			}
			if pressIntent.Source != "mouse" {
				t.Errorf("Expected source 'mouse', got '%s'", pressIntent.Source)
			}
		}
	}

	// 保持状态（无新输入）
	sameSnap := &InputSnapshot{
		MouseX:       10,
		MouseY:       10,
		MouseButton:  runtimemsg.MouseLeft,
		MouseAction:  runtimemsg.MouseActionUnknown,
		Timestamp:    time.Now().UnixNano(),
	}
	intents = tracker.Update(sameSnap)
	if len(intents) != 0 {
		t.Errorf("Expected 0 intents when state unchanged, got %d", len(intents))
	}

	// 释放
	releaseSnap := &InputSnapshot{
		MouseX:       10,
		MouseY:       10,
		MouseButton:  runtimemsg.MouseButtonUnknown,
		MouseAction:  runtimemsg.MouseActionRelease,
		Timestamp:    time.Now().UnixNano(),
	}
	intents = tracker.Update(releaseSnap)

	if len(intents) != 1 {
		t.Errorf("Expected 1 intent on release, got %d", len(intents))
	} else {
		releaseIntent, ok := intents[0].(InputReleaseIntent)
		if !ok {
			t.Error("Expected InputReleaseIntent, got different type")
		} else {
			if releaseIntent.Source != "mouse" {
				t.Errorf("Expected source 'mouse', got '%s'", releaseIntent.Source)
			}
		}
	}
}

func TestInputTracker_MouseMove(t *testing.T) {
	tracker := NewInputTracker()

	snap1 := &InputSnapshot{
		MouseX:       10,
		MouseY:       10,
		MouseButton:  runtimemsg.MouseButtonUnknown,
		MouseAction:  runtimemsg.MouseActionUnknown,
		Timestamp:    time.Now().UnixNano(),
	}
	intents := tracker.Update(snap1)
	// 第一次更新会触发一个 InputMoveIntent（从 0,0 移动到 10,10）
	if len(intents) != 1 {
		t.Errorf("Expected 1 intent on initial snapshot, got %d", len(intents))
	} else {
		moveIntent, ok := intents[0].(InputMoveIntent)
		if !ok {
			t.Error("Expected InputMoveIntent, got different type")
		} else {
			if moveIntent.X != 10 || moveIntent.Y != 10 {
				t.Errorf("Expected position (10,10), got (%d,%d)", moveIntent.X, moveIntent.Y)
			}
		}
	}

	snap2 := &InputSnapshot{
		MouseX:       20,
		MouseY:       30,
		MouseButton:  runtimemsg.MouseButtonUnknown,
		MouseAction:  runtimemsg.MouseActionUnknown,
		Timestamp:    time.Now().UnixNano(),
	}
	intents = tracker.Update(snap2)

	if len(intents) != 1 {
		t.Errorf("Expected 1 intent on move, got %d", len(intents))
	} else {
		moveIntent, ok := intents[0].(InputMoveIntent)
		if !ok {
			t.Error("Expected InputMoveIntent, got different type")
		} else {
			if moveIntent.X != 20 || moveIntent.Y != 30 {
				t.Errorf("Expected position (20,30), got (%d,%d)", moveIntent.X, moveIntent.Y)
			}
		}
	}
}

func TestInputTracker_KeyboardInput(t *testing.T) {
	tracker := NewInputTracker()

	// 可打印字符
	snap1 := &InputSnapshot{
		KeyboardKey:  'a',
		SpecialKey:   runtimeplatform.KeyUnknown,
		Modifiers:    runtimemsg.Modifiers{},
		Timestamp:    time.Now().UnixNano(),
	}
	intents := tracker.Update(snap1)

	if len(intents) != 1 {
		t.Errorf("Expected 1 intent on key input, got %d", len(intents))
	} else {
		keyIntent, ok := intents[0].(InputKeyboardIntent)
		if !ok {
			t.Error("Expected InputKeyboardIntent, got different type")
		} else {
			if keyIntent.Key != 'a' {
				t.Errorf("Expected key 'a', got '%c'", keyIntent.Key)
			}
		}
	}

	// 特殊键
	snap2 := &InputSnapshot{
		KeyboardKey:  0,
		SpecialKey:   runtimeplatform.KeyEnter,
		Modifiers:    runtimemsg.Modifiers{},
		Timestamp:    time.Now().UnixNano(),
	}
	intents = tracker.Update(snap2)

	if len(intents) != 1 {
		t.Errorf("Expected 1 intent on special key, got %d", len(intents))
	} else {
		keyIntent, ok := intents[0].(InputKeyboardIntent)
		if !ok {
			t.Error("Expected InputKeyboardIntent, got different type")
		} else {
			if keyIntent.Special != runtimeplatform.KeyEnter {
				t.Errorf("Expected KeyEnter, got %v", keyIntent.Special)
			}
		}
	}
}

func TestInputSnapshot_Clone(t *testing.T) {
	original := &InputSnapshot{
		MouseX:      10,
		MouseY:      20,
		MouseButton: runtimemsg.MouseLeft,
		KeyboardKey: 'a',
		SpecialKey:  runtimeplatform.KeyEnter,
		Modifiers:   runtimemsg.Modifiers{},
		Timestamp:   12345,
	}

	cloned := original.Clone()

	if cloned.MouseX != original.MouseX {
		t.Errorf("Clone failed: MouseX %d != %d", cloned.MouseX, original.MouseX)
	}
	if cloned.MouseY != original.MouseY {
		t.Errorf("Clone failed: MouseY %d != %d", cloned.MouseY, original.MouseY)
	}
	if cloned.MouseButton != original.MouseButton {
		t.Errorf("Clone failed: MouseButton %d != %d", cloned.MouseButton, original.MouseButton)
	}
	if cloned.KeyboardKey != original.KeyboardKey {
		t.Errorf("Clone failed: KeyboardKey %c != %c", cloned.KeyboardKey, original.KeyboardKey)
	}
	if cloned.SpecialKey != original.SpecialKey {
		t.Errorf("Clone failed: SpecialKey %v != %v", cloned.SpecialKey, original.SpecialKey)
	}
}


