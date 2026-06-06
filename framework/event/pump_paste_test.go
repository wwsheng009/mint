package event

import (
	"testing"

	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
	"github.com/wwsheng009/mint/runtime/platform"
)

func TestPump_ConvertToMsg_ConvertsPastePayload(t *testing.T) {
	pump := NewPumpWithSource(NewChannelEventSource(make(chan platform.RawInput)))

	msg := pump.convertToMsg(platform.RawInput{
		Type: platform.InputPaste,
		Data: []byte("https://example.test/login"),
	})
	if msg == nil {
		t.Fatal("convertToMsg returned nil for paste input")
	}

	pasteMsg, ok := msg.(*runtimemsg.PasteMsg)
	if !ok {
		t.Fatalf("convertToMsg() = %T, want *msg.PasteMsg", msg)
	}
	if pasteMsg.Text != "https://example.test/login" {
		t.Fatalf("paste text = %q, want pasted payload", pasteMsg.Text)
	}
}

func TestPump_ConvertToMsg_IgnoresEmptyPastePayload(t *testing.T) {
	pump := NewPumpWithSource(NewChannelEventSource(make(chan platform.RawInput)))

	if msg := pump.convertToMsg(platform.RawInput{Type: platform.InputPaste}); msg != nil {
		t.Fatalf("convertToMsg() = %T, want nil for empty paste payload", msg)
	}
}
