package msg

import (
	"testing"

	runtimeplatform "github.com/wwsheng009/mint/runtime/platform"
)

func TestIsPrintable_ChineseCharacters(t *testing.T) {
	tests := []struct {
		name string
		msg  *KeyMsg
		want bool
	}{
		{
			name: "ASCII letter",
			msg:  NewKeyMsg('A', runtimeplatform.KeyUnknown, Modifiers{}),
			want: true,
		},
		{
			name: "ASCII number",
			msg:  NewKeyMsg('5', runtimeplatform.KeyUnknown, Modifiers{}),
			want: true,
		},
		{
			name: "Chinese character (你)",
			msg:  NewKeyMsg('你', runtimeplatform.KeyUnknown, Modifiers{}),
			want: true,
		},
		{
			name: "Chinese character (好)",
			msg:  NewKeyMsg('好', runtimeplatform.KeyUnknown, Modifiers{}),
			want: true,
		},
		{
			name: "Chinese character (世)",
			msg:  NewKeyMsg('世', runtimeplatform.KeyUnknown, Modifiers{}),
			want: true,
		},
		{
			name: "Chinese character (界)",
			msg:  NewKeyMsg('界', runtimeplatform.KeyUnknown, Modifiers{}),
			want: true,
		},
		{
			name: "Japanese character (あ)",
			msg:  NewKeyMsg('あ', runtimeplatform.KeyUnknown, Modifiers{}),
			want: true,
		},
		{
			name: "Korean character (가)",
			msg:  NewKeyMsg('가', runtimeplatform.KeyUnknown, Modifiers{}),
			want: true,
		},
		{
			name: "Emoji (😊)",
			msg:  NewKeyMsg('😊', runtimeplatform.KeyUnknown, Modifiers{}),
			want: true,
		},
		{
			name: "Enter special key",
			msg:  NewKeyMsg(0, runtimeplatform.KeyEnter, Modifiers{}),
			want: false, // Enter is not printable
		},
		{
			name: "Space character",
			msg:  NewKeyMsg(' ', runtimeplatform.KeyUnknown, Modifiers{}),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.msg.IsPrintable(); got != tt.want {
				t.Errorf("KeyMsg.IsPrintable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsPrintable_WithModifiers(t *testing.T) {
	tests := []struct {
		name       string
		msg        *KeyMsg
		wantBefore bool // IsPrintable result
	}{
		{
			name:       "Chinese character with Ctrl modifier",
			msg:        NewKeyMsg('你', runtimeplatform.KeyUnknown, Modifiers{Ctrl: true}),
			wantBefore: true, // IsPrintable should still return true
		},
		{
			name:       "ASCII letter with Alt modifier",
			msg:        NewKeyMsg('A', runtimeplatform.KeyUnknown, Modifiers{Alt: true}),
			wantBefore: true, // IsPrintable should still return true
		},
		{
			name:       "Chinese character with Shift modifier",
			msg:        NewKeyMsg('你', runtimeplatform.KeyUnknown, Modifiers{Shift: true}),
			wantBefore: true, // IsPrintable should still return true
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.msg.IsPrintable(); got != tt.wantBefore {
				t.Errorf("KeyMsg.IsPrintable() = %v, want %v", got, tt.wantBefore)
			}
		})
	}
}
