//go:build windows
// +build windows

package platform

import (
	"testing"
	"time"
	"unsafe"
)

// TestStandaloneModifiers 测试单独的修饰键是否被正确过滤
func TestStandaloneModifiers(t *testing.T) {
	r := &windowsInputReader{}

	tests := []struct {
		name            string
		virtualKeyCode  uint16
		uChar           uint16
		controlKeyState uint32
		shouldBeFiltered bool  // true = 应该被过滤（返回Type=-1）
	}{
		// 单独的修饰键 - 应该被过滤
		{"Shift alone", 0x10, 0, 0x0010, true},
		{"Ctrl alone", 0x11, 0, 0x0008, true},
		{"Alt alone", 0x12, 0, 0x0002, true},

		// Shift组合键 - 不应该被过滤
		{"Shift + A", 0x41, 'A', 0x0010, false},
		{"Shift + a", 0x41, 'a', 0x0000, false},  // 小写a（无Shift）
		{"Shift + Z", 0x5A, 'Z', 0x0010, false},

		// Shift+特殊键组合 - 不应该被过滤
		{"Shift + Tab", 0x09, 0, 0x0010, false},
		{"Shift + Enter", 0x0D, 0, 0x0010, false},

		// Ctrl组合键 - 不应该被过滤
		{"Ctrl + A", 0x41, 1, 0x0008, false},  // Ctrl+A的UChar=1
		{"Ctrl + Z", 0x5A, 26, 0x0008, false},

		// Ctrl+Shift组合键 - 不应该被过滤
		{"Ctrl+Shift + A", 0x41, 1, 0x0018, false},
		{"Ctrl+Shift + Z", 0x5A, 26, 0x0018, false},

		// 纯字符 - 不应该被过滤
		{"Plain a", 0x41, 'a', 0, false},
		{"Plain 1", 0x31, '1', 0, false},

		// 特殊键（无修饰符） - 不应该被过滤
		{"Tab", 0x09, 0, 0, false},
		{"Enter", 0x0D, 0, 0, false},
		{"Escape", 0x1B, 0, 0, false},
	}

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

			if tt.shouldBeFiltered {
				// 应该被过滤 - Type应该是-1
				if input.Type != -1 {
					t.Errorf("%s: should be filtered (Type=-1), but got Type=%d, Key=%c, Modifiers=%d",
						tt.name, input.Type, input.Key, input.Modifiers)
				} else {
					t.Logf("✅ %s: correctly filtered (Type=-1)", tt.name)
				}
			} else {
				// 不应该被过滤 - Type应该是InputKeyPress
				if input.Type != InputKeyPress {
					t.Errorf("%s: should NOT be filtered, but got Type=%d",
						tt.name, input.Type)
				}

				modStr := ""
				if input.Modifiers&ModShift != 0 {
					modStr += "Shift+"
				}
				if input.Modifiers&ModCtrl != 0 {
					modStr += "Ctrl+"
				}
				if input.Modifiers&ModAlt != 0 {
					modStr += "Alt+"
				}

				t.Logf("✅ %s: NOT filtered, Key='%c' Modifiers=%s Special=%d",
					tt.name, input.Key, modStr, input.Special)
			}
		})
	}
}
