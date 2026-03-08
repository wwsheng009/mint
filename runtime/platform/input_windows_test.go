//go:build windows
// +build windows

package platform

import (
	"testing"
	"time"
	"unsafe"
)

func wheelButtonState(delta int16) uint32 {
	return uint32(uint16(delta)) << 16
}

func TestWindowsInputReader_ParseMouseEvent_WheelDirection(t *testing.T) {
	reader := &windowsInputReader{}

	tests := []struct {
		name       string
		flags      uint32
		delta      int16
		wantAction MouseAction
	}{
		{
			name:       "vertical wheel up",
			flags:      MOUSE_WHEELED,
			delta:      120,
			wantAction: MouseWheelUp,
		},
		{
			name:       "vertical wheel down",
			flags:      MOUSE_WHEELED,
			delta:      -120,
			wantAction: MouseWheelDown,
		},
		{
			name:       "horizontal wheel positive",
			flags:      MOUSE_HWHEELED,
			delta:      120,
			wantAction: MouseWheelUp,
		},
		{
			name:       "horizontal wheel negative",
			flags:      MOUSE_HWHEELED,
			delta:      -120,
			wantAction: MouseWheelDown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			record := &INPUT_RECORD{}
			mouseRecord := (*MOUSE_EVENT_RECORD)(unsafe.Pointer(&record.Event[0]))
			mouseRecord.EventFlags = tt.flags
			mouseRecord.ButtonState = wheelButtonState(tt.delta)

			input := reader.parseMouseEvent(record, time.Now())
			if input.MouseAction != tt.wantAction {
				t.Fatalf("mouse action = %v, want %v", input.MouseAction, tt.wantAction)
			}
		})
	}
}
