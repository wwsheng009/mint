package action

import (
	"testing"

	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
	"github.com/wwsheng009/mint/runtime/platform"
)

func TestInputProcessor_ProcessPasteMsgMapsToPasteAction(t *testing.T) {
	processor := NewInputProcessor()

	act := processor.ProcessMsg(runtimemsg.NewPasteMsg("https://example.test/login"))
	if act == nil {
		t.Fatal("ProcessMsg returned nil for paste message")
	}
	if act.Type != ActionPaste {
		t.Fatalf("action type = %s, want %s", act.Type, ActionPaste)
	}
	text, ok := act.GetPayloadString()
	if !ok || text != "https://example.test/login" {
		t.Fatalf("payload = %q, want pasted text", text)
	}
}

func TestInputProcessor_CtrlVPasteMapsToPasteAction(t *testing.T) {
	processor := NewInputProcessor()

	act := processor.ProcessMsg(runtimemsg.NewKeyMsg('v', platform.KeyUnknown, runtimemsg.Modifiers{Ctrl: true}))
	if act == nil {
		t.Fatal("ProcessMsg returned nil for Ctrl+V")
	}
	if act.Type != ActionPaste {
		t.Fatalf("action type = %s, want %s", act.Type, ActionPaste)
	}
}
