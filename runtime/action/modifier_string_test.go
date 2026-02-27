package action

import (
	"testing"
)

// TestModifierString 测试 Modifier.String() 方法在位掩码定义下是否正确工作
func TestModifierString(t *testing.T) {
	tests := []struct {
		name     string
		modifier Modifier
		expected string
	}{
		{"None", ModNone, "none"},
		{"Shift only", ModShift, "shift"},
		{"Ctrl only", ModCtrl, "ctrl"},
		{"Alt only", ModAlt, "alt"},
		{"Meta only", ModMeta, "meta"},
		{"Ctrl+Shift", ModCtrl | ModShift, "shift+ctrl"},
		{"Ctrl+Alt", ModCtrl | ModAlt, "ctrl+alt"},
		{"Ctrl+Meta", ModCtrl | ModMeta, "ctrl+meta"},
		{"Shift+Alt", ModShift | ModAlt, "shift+alt"},
		{"Shift+Meta", ModShift | ModMeta, "shift+meta"},
		{"Alt+Meta", ModAlt | ModMeta, "alt+meta"},
		{"Ctrl+Shift+Alt", ModCtrl | ModShift | ModAlt, "shift+ctrl+alt"},
		{"Ctrl+Shift+Meta", ModCtrl | ModShift | ModMeta, "shift+ctrl+meta"},
		{"Ctrl+Alt+Meta", ModCtrl | ModAlt | ModMeta, "ctrl+alt+meta"},
		{"Shift+Alt+Meta", ModShift | ModAlt | ModMeta, "shift+alt+meta"},
		{"All modifiers", ModCtrl | ModShift | ModAlt | ModMeta, "shift+ctrl+alt+meta"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.modifier.String()
			if result != tt.expected {
				t.Errorf("Modifier.String() = %s, want %s", result, tt.expected)
			} else {
				t.Logf("✅ %s: %s", tt.name, result)
			}
		})
	}
}
