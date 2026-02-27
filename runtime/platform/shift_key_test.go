//go:build windows
// +build windows

package platform

import (
	"testing"
	"time"
	"unsafe"
)

// TestShiftPlusLetters 测试 Shift + 字母的处理
func TestShiftPlusLetters(t *testing.T) {
	tests := []struct {
		name            string
		virtualKeyCode  uint16
		uChar           uint16
		controlKeyState uint32
		expectedKey     rune
		expectedShift   bool
	}{
		// Shift + 字母键
		{"Shift+A", 0x41, 'A', 0x0010, 'A', true},
		{"Shift+B", 0x42, 'B', 0x0010, 'B', true},
		{"Shift+C", 0x43, 'C', 0x0010, 'C', true},
		{"Shift+Z", 0x5A, 'Z', 0x0010, 'Z', true},

		// Plain 字母键（无 Shift）
		{"Plain a", 0x41, 'a', 0, 'a', false},
		{"Plain b", 0x42, 'b', 0, 'b', false},
		{"Plain z", 0x5A, 'z', 0, 'z', false},

		// Shift + 数字键
		{"Shift+1 (@)", 0x31, '@', 0x0010, '@', true},
		{"Shift+2 (#)", 0x32, '#', 0x0010, '#', true},
		{"Shift+3 ($)", 0x33, '$', 0x0010, '$', true},
		{"Shift+4 (%)", 0x34, '%', 0x0010, '%', true},
	}

	r := &windowsInputReader{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建 INPUT_RECORD
			record := &INPUT_RECORD{
				EventType: KEY_EVENT,
				Padding:   0,
			}

			// 填充 KEY_EVENT_RECORD
			keyEvent := (*KEY_EVENT_RECORD)(unsafe.Pointer(&record.Event[0]))
			keyEvent.KeyDown = 1
			keyEvent.VirtualKeyCode = tt.virtualKeyCode
			keyEvent.UChar = tt.uChar
			keyEvent.ControlKeyState = tt.controlKeyState

			now := time.Now()
			input := r.parseKeyEvent(record, now)

			// 验证结果
			if input.Type != InputKeyPress {
				t.Errorf("Type = %v, want InputKeyPress", input.Type)
			}

			if input.Key != tt.expectedKey {
				t.Errorf("Key = %c (0x%02X), want %c (0x%02X)",
					input.Key, input.Key, tt.expectedKey, tt.expectedKey)
			}

			hasShift := input.Modifiers&ModShift != 0
			if hasShift != tt.expectedShift {
				t.Errorf("Shift modifier = %v, want %v", hasShift, tt.expectedShift)
			}

			if hasShift {
				t.Logf("✅ %s: Key='%c' Modifiers=Shift+", tt.name, input.Key)
			} else {
				t.Logf("✅ %s: Key='%c' Modifiers=none", tt.name, input.Key)
			}
		})
	}
}
