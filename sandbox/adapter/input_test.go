// adapter/input_test.go - 输入适配器测试
package adapter

import (
	"testing"
	"time"

	"github.com/wwsheng009/mint/runtime/platform"
)

func TestBuildKeyEvent(t *testing.T) {
	event := BuildKeyEvent('a')

	if event.Type != platform.InputKeyPress {
		t.Errorf("BuildKeyEvent() Type = %v, want %v", event.Type, platform.InputKeyPress)
	}

	if event.Key != 'a' {
		t.Errorf("BuildKeyEvent() Key = %v, want 'a'", event.Key)
	}

	if event.Timestamp.IsZero() {
		t.Error("BuildKeyEvent() Timestamp should not be zero")
	}
}

func TestBuildSpecialKeyEvent(t *testing.T) {
	event := BuildSpecialKeyEvent(platform.KeyEnter, platform.ModCtrl)

	if event.Type != platform.InputKeyPress {
		t.Errorf("BuildSpecialKeyEvent() Type = %v, want %v", event.Type, platform.InputKeyPress)
	}

	if event.Special != platform.KeyEnter {
		t.Errorf("BuildSpecialKeyEvent() Special = %v, want %v", event.Special, platform.KeyEnter)
	}

	if event.Modifiers != platform.ModCtrl {
		t.Errorf("BuildSpecialKeyEvent() Modifiers = %v, want %v", event.Modifiers, platform.ModCtrl)
	}
}

func TestBuildMouseEvent(t *testing.T) {
	event := BuildMouseEvent(10, 20, platform.MouseLeft, platform.MousePress)

	if event.Type != platform.InputMouse {
		t.Errorf("BuildMouseEvent() Type = %v, want %v", event.Type, platform.InputMouse)
	}

	if event.MouseX != 10 {
		t.Errorf("BuildMouseEvent() MouseX = %v, want 10", event.MouseX)
	}

	if event.MouseY != 20 {
		t.Errorf("BuildMouseEvent() MouseY = %v, want 20", event.MouseY)
	}

	if event.MouseButton != platform.MouseLeft {
		t.Errorf("BuildMouseEvent() MouseButton = %v, want %v", event.MouseButton, platform.MouseLeft)
	}
}

func TestBuildResizeEvent(t *testing.T) {
	event := BuildResizeEvent(80, 24)

	if event.Type != platform.InputResize {
		t.Errorf("BuildResizeEvent() Type = %v, want %v", event.Type, platform.InputResize)
	}

	if event.Width != 80 {
		t.Errorf("BuildResizeEvent() Width = %v, want 80", event.Width)
	}

	if event.Height != 24 {
		t.Errorf("BuildResizeEvent() Height = %v, want 24", event.Height)
	}
}

func TestBuildPasteEvent(t *testing.T) {
	text := "test content"
	event := BuildPasteEvent(text)

	if event.Type != platform.InputPaste {
		t.Errorf("BuildPasteEvent() Type = %v, want %v", event.Type, platform.InputPaste)
	}

	if string(event.Data) != text {
		t.Errorf("BuildPasteEvent() Data = %v, want %v", string(event.Data), text)
	}
}

func TestToSandboxEvent(t *testing.T) {
	raw := platform.RawInput{
		Type:      platform.InputKeyPress,
		Key:       'x',
		Timestamp: time.Now(),
	}

	sandEvent := ToSandboxEvent(raw, true)

	if sandEvent.Raw.Type != raw.Type {
		t.Error("ToSandboxEvent() Raw event mismatch")
	}

	if !sandEvent.Injected {
		t.Error("ToSandboxEvent() Injected = true, want false")
	}

	if sandEvent.Timestamp.IsZero() {
		t.Error("ToSandboxEvent() Timestamp should not be zero")
	}
}
