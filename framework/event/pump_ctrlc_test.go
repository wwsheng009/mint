package event

import (
	"testing"
	"time"

	"github.com/wwsheng009/mint/runtime/platform"
)

func TestPump_CtrlCAsQuit_Disabled(t *testing.T) {
	rawCh := make(chan platform.RawInput, 4)
	pump := NewPumpWithSource(NewChannelEventSource(rawCh))
	pump.SetCtrlCAsQuit(false)

	if err := pump.Start(); err != nil {
		t.Fatalf("pump.Start() error: %v", err)
	}
	defer pump.Stop()

	rawCh <- platform.RawInput{
		Type:      platform.InputKeyPress,
		Key:       'c',
		Modifiers: platform.ModCtrl,
	}

	select {
	case <-pump.QuitAppRequested():
		t.Fatal("QuitAppRequested should not close when Ctrl+C quit is disabled")
	case <-time.After(80 * time.Millisecond):
		// expected
	}
}

func TestPump_CtrlCAsQuit_Enabled(t *testing.T) {
	rawCh := make(chan platform.RawInput, 4)
	pump := NewPumpWithSource(NewChannelEventSource(rawCh))
	pump.SetCtrlCAsQuit(true)

	if err := pump.Start(); err != nil {
		t.Fatalf("pump.Start() error: %v", err)
	}
	defer pump.Stop()

	rawCh <- platform.RawInput{
		Type:      platform.InputKeyPress,
		Key:       'C',
		Modifiers: platform.ModCtrl,
	}

	select {
	case <-pump.QuitAppRequested():
		// expected
	case <-time.After(200 * time.Millisecond):
		t.Fatal("QuitAppRequested should close when Ctrl+C quit is enabled")
	}
}
