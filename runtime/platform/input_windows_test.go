//go:build windows
// +build windows

package platform

import (
	"testing"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
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

func TestWindowsInputReader_DedupeResizeEvent(t *testing.T) {
	reader := &windowsInputReader{
		lastWidth:  120,
		lastHeight: 40,
	}

	if input := reader.dedupeResizeInput(120, 40, time.Now()); input.Type != -1 {
		t.Fatalf("same-size resize type = %v, want invalid input", input.Type)
	}

	input := reader.dedupeResizeInput(121, 40, time.Now())
	if input.Type != InputResize {
		t.Fatalf("changed-size resize type = %v, want %v", input.Type, InputResize)
	}
	if input.Width != 121 || input.Height != 40 {
		t.Fatalf("resize dimensions = %dx%d, want 121x40", input.Width, input.Height)
	}
	if reader.lastWidth != 121 || reader.lastHeight != 40 {
		t.Fatalf("tracked size = %dx%d, want 121x40", reader.lastWidth, reader.lastHeight)
	}
}

func TestWindowsInputReader_PolledResizeRequiresStableConfirmation(t *testing.T) {
	reader := &windowsInputReader{
		lastWidth:  120,
		lastHeight: 40,
	}

	if input := reader.polledResizeInput(121, 40, time.Now()); input.Type != -1 {
		t.Fatalf("first polled resize type = %v, want invalid input", input.Type)
	}
	if !reader.pendingPollResize || reader.pendingPollWidth != 121 || reader.pendingPollHeight != 40 {
		t.Fatalf("pending polled size = enabled:%v %dx%d, want enabled: true 121x40", reader.pendingPollResize, reader.pendingPollWidth, reader.pendingPollHeight)
	}

	input := reader.polledResizeInput(121, 40, time.Now())
	if input.Type != InputResize {
		t.Fatalf("confirmed polled resize type = %v, want %v", input.Type, InputResize)
	}
	if reader.pendingPollResize {
		t.Fatal("pending polled resize should be cleared after confirmed emit")
	}
	if reader.lastWidth != 121 || reader.lastHeight != 40 {
		t.Fatalf("tracked size = %dx%d, want 121x40", reader.lastWidth, reader.lastHeight)
	}
}

func TestWindowsInputReader_PolledResizeOscillationDoesNotEmit(t *testing.T) {
	reader := &windowsInputReader{
		lastWidth:  120,
		lastHeight: 40,
	}

	if input := reader.polledResizeInput(121, 40, time.Now()); input.Type != -1 {
		t.Fatalf("first oscillating poll type = %v, want invalid input", input.Type)
	}
	if input := reader.polledResizeInput(120, 40, time.Now()); input.Type != -1 {
		t.Fatalf("return-to-stable poll type = %v, want invalid input", input.Type)
	}
	if reader.pendingPollResize {
		t.Fatal("pending polled resize should clear when size returns to stable baseline")
	}
	if reader.lastWidth != 120 || reader.lastHeight != 40 {
		t.Fatalf("tracked size = %dx%d, want baseline 120x40", reader.lastWidth, reader.lastHeight)
	}
}

func TestWindowsInputReader_RealResizeClearsPendingPolledCandidate(t *testing.T) {
	reader := &windowsInputReader{
		lastWidth:  120,
		lastHeight: 40,
	}

	_ = reader.polledResizeInput(121, 40, time.Now())
	reader.clearPendingPolledResize()
	if reader.pendingPollResize {
		t.Fatal("pending polled resize should be cleared")
	}
	input := reader.dedupeResizeInput(121, 40, time.Now())
	if input.Type != InputResize {
		t.Fatalf("real resize type = %v, want %v", input.Type, InputResize)
	}
}

func TestBuildWindowsInputConsoleMode_DoesNotLeakOutputFlags(t *testing.T) {
	original := uint32(ENABLE_LINE_INPUT | ENABLE_ECHO_INPUT | ENABLE_PROCESSED_INPUT)
	mode := buildWindowsInputConsoleMode(original, true)

	if mode&ENABLE_LINE_INPUT != 0 {
		t.Fatalf("ENABLE_LINE_INPUT still set in input mode: 0x%08X", mode)
	}
	if mode&ENABLE_ECHO_INPUT != 0 {
		t.Fatalf("ENABLE_ECHO_INPUT still set in input mode: 0x%08X", mode)
	}
	if mode&ENABLE_WINDOW_INPUT == 0 {
		t.Fatalf("ENABLE_WINDOW_INPUT missing from input mode: 0x%08X", mode)
	}
	if mode&ENABLE_MOUSE_INPUT == 0 {
		t.Fatalf("ENABLE_MOUSE_INPUT missing from input mode: 0x%08X", mode)
	}
	if mode&ENABLE_VIRTUAL_TERMINAL_INPUT != 0 {
		t.Fatalf("ENABLE_VIRTUAL_TERMINAL_INPUT should be disabled in raw input mode: 0x%08X", mode)
	}
	if mode&windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING != 0 {
		t.Fatalf("output VT flag leaked into input mode: 0x%08X", mode)
	}
}

func TestBuildWindowsOutputConsoleMode_EnablesVTProcessing(t *testing.T) {
	mode := buildWindowsOutputConsoleMode(0)
	if mode&windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING == 0 {
		t.Fatalf("ENABLE_VIRTUAL_TERMINAL_PROCESSING missing from output mode: 0x%08X", mode)
	}
}
