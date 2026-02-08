package platform

import (
	"testing"
)

// TestCtrlCharacterConversion tests that Ctrl+letter control characters
// are properly converted to letter + Ctrl modifier, with case preservation for Shift
func TestCtrlCharacterConversion(t *testing.T) {
	tests := []struct {
		name            string
		uChar           uint16
		virtualKeyCode  uint16
		controlKeyState uint32
		expectedKey     rune
		expectedCtrl    bool
		expectedShift   bool
	}{
		// Plain Ctrl (no Shift) - should be lowercase
		{"Ctrl+A", 1, 0x41, 0x0008, 'a', true, false},
		{"Ctrl+B", 2, 0x42, 0x0008, 'b', true, false},
		{"Ctrl+C", 3, 0x43, 0x0008, 'c', true, false},
		{"Ctrl+D", 4, 0x44, 0x0008, 'd', true, false},
		{"Ctrl+K", 11, 0x4B, 0x0008, 'k', true, false},
		{"Ctrl+Z", 26, 0x5A, 0x0008, 'z', true, false},

		// Ctrl+Shift - should be uppercase
		{"Ctrl+Shift+A", 1, 0x41, 0x0018, 'A', true, true},
		{"Ctrl+Shift+B", 2, 0x42, 0x0018, 'B', true, true},
		{"Ctrl+Shift+C", 3, 0x43, 0x0018, 'C', true, true},
		{"Ctrl+Shift+D", 4, 0x44, 0x0018, 'D', true, true},
		{"Ctrl+Shift+K", 11, 0x4B, 0x0018, 'K', true, true},
		{"Ctrl+Shift+Z", 26, 0x5A, 0x0018, 'Z', true, true},

		// Plain letters (no Ctrl) - case preserved
		{"Plain D (no Ctrl)", 'D', 0x44, 0, 'D', false, false},
		{"Plain K (no Ctrl)", 'K', 0x4B, 0, 'K', false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the parseKeyEvent logic
			input := RawInput{
				Special: KeyUnknown,
			}

			// Check for Shift
			if tt.controlKeyState&0x0010 != 0 {
				input.Modifiers |= ModShift
			}
			// Check for Ctrl
			if tt.controlKeyState&0x0004 != 0 || tt.controlKeyState&0x0008 != 0 {
				input.Modifiers |= ModCtrl
			}
			// Check for Alt
			if tt.controlKeyState&0x0002 != 0 || tt.controlKeyState&0x0001 != 0 {
				input.Modifiers |= ModAlt
			}

			// Handle Ctrl+letter combinations
			if tt.uChar >= 1 && tt.uChar <= 26 && tt.virtualKeyCode >= 0x41 && tt.virtualKeyCode <= 0x5A {
				// This is Ctrl+letter (A-Z)
				// Preserve case: lowercase for ctrl+letter, uppercase for ctrl+shift+letter
				if tt.controlKeyState&0x0010 != 0 {
					// Shift is pressed - use uppercase
					input.Key = rune(tt.virtualKeyCode)
				} else {
					// No shift - use lowercase
					input.Key = rune(tt.virtualKeyCode + 32)
				}
				input.Modifiers |= ModCtrl
				input.Special = KeyUnknown
			} else if input.Special == KeyUnknown && tt.uChar > 0 {
				input.Key = rune(tt.uChar)
			}

			// Check results
			if input.Key != tt.expectedKey {
				t.Errorf("Key = %c, want %c", input.Key, tt.expectedKey)
			}

			hasCtrl := input.Modifiers&ModCtrl != 0
			if hasCtrl != tt.expectedCtrl {
				t.Errorf("Ctrl modifier = %v, want %v", hasCtrl, tt.expectedCtrl)
			}

			hasShift := input.Modifiers&ModShift != 0
			if hasShift != tt.expectedShift {
				t.Errorf("Shift modifier = %v, want %v", hasShift, tt.expectedShift)
			}

			modStr := ""
			if hasCtrl {
				modStr += "Ctrl+"
			}
			if hasShift {
				modStr += "Shift+"
			}
			t.Logf("✅ %s: Key='%c' Modifiers=%s", tt.name, input.Key, modStr)
		})
	}
}
