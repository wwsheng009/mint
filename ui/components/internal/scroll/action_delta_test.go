package scroll

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/action"
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
)

func TestDeltaFromAction_IntPayload(t *testing.T) {
	act := action.NewActionWithPayload(action.ActionScroll, 3)
	delta, ok := DeltaFromAction(act)
	if !ok {
		t.Fatal("delta should be extracted")
	}
	if delta != 3 {
		t.Fatalf("delta = %d, want 3", delta)
	}
}

func TestDeltaFromAction_MouseWheelPayload(t *testing.T) {
	mouseMsg := runtimemsg.NewMouseMsgWithDelta(0, 0, 1, runtimemsg.MouseActionWheel)
	act := action.NewActionWithPayload(action.ActionScroll, mouseMsg)
	delta, ok := DeltaFromAction(act)
	if !ok {
		t.Fatal("delta should be extracted")
	}
	if delta != -1 {
		t.Fatalf("delta = %d, want -1", delta)
	}
}
