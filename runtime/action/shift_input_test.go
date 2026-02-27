package action

import (
	"bytes"
	"fmt"
	"testing"

	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
	runtimeplatform "github.com/wwsheng009/mint/runtime/platform"
)

// TestShiftInputText 测试 Shift + 字符输入转换
func TestShiftInputText(t *testing.T) {
	processor := NewInputProcessor()

	tests := []struct {
		name         string
		keyMsg       *runtimemsg.KeyMsg
		expectAction bool        // true 表示期望得到 ActionInputText
		expectedChar string      // 期望的字符
	}{
		// Shift + 字母 → 大写字母
		{"Shift + a", runtimemsg.NewKeyMsg('A', runtimeplatform.KeyUnknown, runtimemsg.Modifiers{Shift: true}), true, "A"},
		{"Shift + b", runtimemsg.NewKeyMsg('B', runtimeplatform.KeyUnknown, runtimemsg.Modifiers{Shift: true}), true, "B"},
		{"Shift + z", runtimemsg.NewKeyMsg('Z', runtimeplatform.KeyUnknown, runtimemsg.Modifiers{Shift: true}), true, "Z"},

		// Shift + 数字 → 符号
		{"Shift + 1 (!)", runtimemsg.NewKeyMsg('!', runtimeplatform.KeyUnknown, runtimemsg.Modifiers{Shift: true}), true, "!"},
		{"Shift + 2 (@)", runtimemsg.NewKeyMsg('@', runtimeplatform.KeyUnknown, runtimemsg.Modifiers{Shift: true}), true, "@"},
		{"Shift + 3 (#)", runtimemsg.NewKeyMsg('#', runtimeplatform.KeyUnknown, runtimemsg.Modifiers{Shift: true}), true, "#"},
		{"Shift + 4 ($)", runtimemsg.NewKeyMsg('$', runtimeplatform.KeyUnknown, runtimemsg.Modifiers{Shift: true}), true, "$"},
		{"Shift + 5 (%)", runtimemsg.NewKeyMsg('%', runtimeplatform.KeyUnknown, runtimemsg.Modifiers{Shift: true}), true, "%"},
		{"Shift + 6 (^)", runtimemsg.NewKeyMsg('^', runtimeplatform.KeyUnknown, runtimemsg.Modifiers{Shift: true}), true, "^"},
		{"Shift + 7 (&)", runtimemsg.NewKeyMsg('&', runtimeplatform.KeyUnknown, runtimemsg.Modifiers{Shift: true}), true, "&"},
		{"Shift + 8 (*)", runtimemsg.NewKeyMsg('*', runtimeplatform.KeyUnknown, runtimemsg.Modifiers{Shift: true}), true, "*"},
		{"Shift + 9 (()", runtimemsg.NewKeyMsg('(', runtimeplatform.KeyUnknown, runtimemsg.Modifiers{Shift: true}), true, "("},
		{"Shift + 0 ())", runtimemsg.NewKeyMsg(')', runtimeplatform.KeyUnknown, runtimemsg.Modifiers{Shift: true}), true, ")"},

		// Shift + 其他符号
		{"Shift + - (_)", runtimemsg.NewKeyMsg('_', runtimeplatform.KeyUnknown, runtimemsg.Modifiers{Shift: true}), true, "_"},
		{"Shift + = (+)", runtimemsg.NewKeyMsg('+', runtimeplatform.KeyUnknown, runtimemsg.Modifiers{Shift: true}), true, "+"},
		{"Shift + [ ({)", runtimemsg.NewKeyMsg('{', runtimeplatform.KeyUnknown, runtimemsg.Modifiers{Shift: true}), true, "{"},
		{"Shift + ] (})", runtimemsg.NewKeyMsg('}', runtimeplatform.KeyUnknown, runtimemsg.Modifiers{Shift: true}), true, "}"},
		{"Shift + \\ (|)", runtimemsg.NewKeyMsg('|', runtimeplatform.KeyUnknown, runtimemsg.Modifiers{Shift: true}), true, "|"},
		{"Shift + ' (\")", runtimemsg.NewKeyMsg('"', runtimeplatform.KeyUnknown, runtimemsg.Modifiers{Shift: true}), true, "\""},
		{"Shift + ; (:)", runtimemsg.NewKeyMsg(':', runtimeplatform.KeyUnknown, runtimemsg.Modifiers{Shift: true}), true, ":"},

		// 普通字符（无修饰符）
		{"Plain a", runtimemsg.NewKeyMsg('a', runtimeplatform.KeyUnknown, runtimemsg.Modifiers{}), true, "a"},
		{"Plain 1", runtimemsg.NewKeyMsg('1', runtimeplatform.KeyUnknown, runtimemsg.Modifiers{}), true, "1"},
		{"Plain !", runtimemsg.NewKeyMsg('!', runtimeplatform.KeyUnknown, runtimemsg.Modifiers{}), true, "!"},

		// Ctrl 组合键 - 不应该返回文本输入
		{"Ctrl + c", runtimemsg.NewKeyMsg('c', runtimeplatform.KeyUnknown, runtimemsg.Modifiers{Ctrl: true}), false, ""},
		{"Ctrl + v", runtimemsg.NewKeyMsg('v', runtimeplatform.KeyUnknown, runtimemsg.Modifiers{Ctrl: true}), false, ""},

		// Ctrl+Shift 组合键 - 不应该返回文本输入
		{"Ctrl+Shift + C", runtimemsg.NewKeyMsg('C', runtimeplatform.KeyUnknown, runtimemsg.Modifiers{Ctrl: true, Shift: true}), false, ""},

		// Alt 组合键 - 不应该返回文本输入
		{"Alt + a", runtimemsg.NewKeyMsg('a', runtimeplatform.KeyUnknown, runtimemsg.Modifiers{Alt: true}), false, ""},

		// 特殊键 - 不应该返回文本输入
		{"Enter", runtimemsg.NewKeyMsg(0, runtimeplatform.KeyEnter, runtimemsg.Modifiers{}), false, ""},
		{"Tab", runtimemsg.NewKeyMsg(0, runtimeplatform.KeyTab, runtimemsg.Modifiers{}), false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action := processor.ProcessMsg(tt.keyMsg)

			if tt.expectAction {
				// 期望得到 ActionInputText
				if action == nil {
					t.Errorf("%s: expected ActionInputText, got nil", tt.name)
					return
				}
				if action.Type != ActionInputText {
					t.Errorf("%s: expected ActionInputText type, got %s", tt.name, action.Type)
					return
				}

				// 检查字符是否正确
				payload, ok := action.Payload.(string)
				if !ok {
					t.Errorf("%s: expected string payload, got %T", tt.name, action.Payload)
					return
				}
				if payload != tt.expectedChar {
					t.Errorf("%s: expected char '%s', got '%s'", tt.name, tt.expectedChar, payload)
					return
				}

				t.Logf("✅ %s → %s", tt.name, tt.expectedChar)
			} else {
				// 不期望得到 ActionInputText
				if action != nil && action.Type == ActionInputText {
					t.Errorf("%s: should NOT return ActionInputText, but got %s", tt.name, action.Type)
				}
			}
		})
	}
}

// TestProcessorToBuffer 测试 InputProcessor 是否能正确处理连续输入并写入缓冲区
func TestProcessorToBuffer(t *testing.T) {
	processor := NewInputProcessor()
	var buf bytes.Buffer

	inputs := []struct {
		name  string
		key   rune
		shift bool
	}{
		{"H", 'H', true},
		{"e", 'e', false},
		{"l", 'l', false},
		{"l", 'l', false},
		{"o", 'o', false},
		{" ", ' ', false},
		{"W", 'W', true},
		{"o", 'o', false},
		{"r", 'r', false},
		{"l", 'l', false},
		{"d", 'd', false},
		{"!", '!', false},
	}

	for _, input := range inputs {
		keyMsg := runtimemsg.NewKeyMsg(input.key, runtimeplatform.KeyUnknown, runtimemsg.Modifiers{Shift: input.shift})
		action := processor.ProcessMsg(keyMsg)

		if action != nil && action.Type == ActionInputText {
			if payload, ok := action.Payload.(string); ok {
				buf.WriteString(payload)
			}
		}
	}

	result := buf.String()
	expected := "Hello World!"

	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	} else {
		t.Logf("✅ Combined input: %s", result)
	}
}

// Example demonstrating the fix
func ExampleInputProcessor_shiftKeys() {
	processor := NewInputProcessor()

	// 模拟用户输入 "HELLO!" (使用 Shift 键)
	inputs := []rune{'H', 'E', 'L', 'L', 'O', '!'}
	var result string

	for _, r := range inputs {
		// 创建带有 Shift 修饰符的 KeyMsg
		keyMsg := runtimemsg.NewKeyMsg(r, runtimeplatform.KeyUnknown, runtimemsg.Modifiers{Shift: true})

		action := processor.ProcessMsg(keyMsg)
		if action != nil && action.Type == ActionInputText {
			if payload, ok := action.Payload.(string); ok {
				result += payload
			}
		}
	}

	fmt.Println(result)
	// Output: HELLO!
}
