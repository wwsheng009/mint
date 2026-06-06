package framework

import (
	"testing"

	frameworkevent "github.com/wwsheng009/mint/framework/event"
	runtimepkg "github.com/wwsheng009/mint/runtime"
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/platform"
)

type mockMouseCaptureInputReader struct {
	mouseCaptureEnabled bool
}

func (m *mockMouseCaptureInputReader) ReadEvent() (platform.RawInput, error) {
	return platform.RawInput{}, nil
}

func (m *mockMouseCaptureInputReader) Start(events chan<- platform.RawInput) error {
	return nil
}

func (m *mockMouseCaptureInputReader) Stop() error {
	return nil
}

func (m *mockMouseCaptureInputReader) SetMouseCaptureEnabled(enabled bool) error {
	m.mouseCaptureEnabled = enabled
	return nil
}

func (m *mockMouseCaptureInputReader) MouseCaptureEnabled() bool {
	return m.mouseCaptureEnabled
}

func TestApp_SetInteractionMode_ConfiguresMouseCaptureAndSelection(t *testing.T) {
	app := NewApp()
	reader := &mockMouseCaptureInputReader{mouseCaptureEnabled: true}
	app.inputReader = reader
	app.pump = frameworkevent.NewPumpWithSource(
		frameworkevent.NewChannelEventSource(make(chan platform.RawInput, 1)),
	)

	if err := app.SetInteractionMode(InteractionModeTerminalSelection); err != nil {
		t.Fatalf("SetInteractionMode(terminal) error: %v", err)
	}
	if reader.mouseCaptureEnabled {
		t.Fatal("mouse capture should be disabled in terminal selection mode")
	}
	if app.GetInteractionMode() != InteractionModeTerminalSelection {
		t.Fatalf("mode mismatch: got %v", app.GetInteractionMode())
	}

	if err := app.SetInteractionMode(InteractionModeAppSelection); err != nil {
		t.Fatalf("SetInteractionMode(app_selection) error: %v", err)
	}
	if !reader.mouseCaptureEnabled {
		t.Fatal("mouse capture should be enabled in app selection mode")
	}
	if app.selectionAdapter == nil || !app.selectionAdapter.IsEnabled() {
		t.Fatal("selection adapter should be enabled in app selection mode")
	}
}

func TestApp_DispatchSelectionEvent_ModeAware(t *testing.T) {
	app := NewApp()
	keyMsg := runtimemsg.NewKeyMsg('a', platform.KeyUnknown, runtimemsg.Modifiers{Ctrl: true})

	if handled := app.dispatchSelectionEvent(keyMsg); handled {
		t.Fatal("selection event should not be handled outside app selection mode")
	}

	if err := app.SetInteractionMode(InteractionModeAppSelection); err != nil {
		t.Fatalf("SetInteractionMode(app_selection) error: %v", err)
	}
	buf := paint.NewBuffer(16, 4)
	app.ensureSelectionAdapter().OnRender(&runtimepkg.Frame{
		Buffer: buf,
		Width:  16,
		Height: 4,
		Dirty:  true,
	})
	if handled := app.dispatchSelectionEvent(keyMsg); !handled {
		t.Fatal("selection event should be handled in app selection mode")
	}
}

func TestApp_DispatchSelectionEvent_MousePressNotConsumed(t *testing.T) {
	app := NewApp()
	if err := app.SetInteractionMode(InteractionModeAppSelection); err != nil {
		t.Fatalf("SetInteractionMode(app_selection) error: %v", err)
	}

	buf := paint.NewBuffer(16, 4)
	app.ensureSelectionAdapter().OnRender(&runtimepkg.Frame{
		Buffer: buf,
		Width:  16,
		Height: 4,
		Dirty:  true,
	})

	mouseMsg := runtimemsg.NewMouseMsg(1, 1, runtimemsg.MouseLeft, runtimemsg.MouseActionPress)
	if consumed := app.dispatchSelectionEvent(mouseMsg); consumed {
		t.Fatal("mouse press should not be consumed to keep UI buttons clickable")
	}
}

func TestApp_HandleGlobalKeyShortcut_InActionPath(t *testing.T) {
	app := NewApp()
	triggered := false
	app.OnKeyCombo("f6", func() {
		triggered = true
	})

	keyMsg := runtimemsg.NewKeyMsg(0, platform.KeyF6, runtimemsg.Modifiers{})
	if handled := app.handleGlobalKeyShortcut(keyMsg); !handled {
		t.Fatal("global key shortcut should be handled in action path")
	}
	if !triggered {
		t.Fatal("global key shortcut handler should be triggered")
	}
}

func TestApp_ProcessMsg_GlobalKeyShortcutWithoutDefaultAction(t *testing.T) {
	app := NewApp()
	triggered := false
	app.OnKeyCombo("f6", func() {
		triggered = true
	})

	app.processMsg(runtimemsg.NewKeyMsg(0, platform.KeyF6, runtimemsg.Modifiers{}))

	if !triggered {
		t.Fatal("global key shortcut should be triggered even when the key has no default action")
	}
}

func TestApp_ProcessMsg_GlobalNavigationShortcutAfterMiddleware(t *testing.T) {
	app := NewApp()
	triggered := false
	app.OnKeyCombo("down", func() {
		triggered = true
	})

	app.processMsg(runtimemsg.NewKeyMsg(0, platform.KeyDown, runtimemsg.Modifiers{}))

	if !triggered {
		t.Fatal("global down shortcut should be triggered when no component middleware consumes it")
	}
}

func TestApp_HandleGlobalKeyShortcut_CtrlComboNormalized(t *testing.T) {
	app := NewApp()
	triggered := false
	app.OnKeyCombo("Ctrl + O", func() {
		triggered = true
	})

	keyMsg := runtimemsg.NewKeyMsg('O', platform.KeyUnknown, runtimemsg.Modifiers{Ctrl: true})
	if handled := app.handleGlobalKeyShortcut(keyMsg); !handled {
		t.Fatal("ctrl+o shortcut should be handled after combo normalization")
	}
	if !triggered {
		t.Fatal("ctrl+o shortcut handler should be triggered")
	}
}

func TestApp_HandleGlobalKeyShortcut_FunctionKeyNormalized(t *testing.T) {
	app := NewApp()
	triggered := false
	app.OnKeyCombo("F5", func() {
		triggered = true
	})

	keyMsg := runtimemsg.NewKeyMsg(0, platform.KeyF5, runtimemsg.Modifiers{})
	if handled := app.handleGlobalKeyShortcut(keyMsg); !handled {
		t.Fatal("f5 shortcut should be handled after combo normalization")
	}
	if !triggered {
		t.Fatal("f5 shortcut handler should be triggered")
	}
}
