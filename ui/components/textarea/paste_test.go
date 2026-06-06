package textarea

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/action"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

func TestInstance_HandleAction_PasteInsertsFullText(t *testing.T) {
	inst := NewInstance(rtui.Props{})

	pasted := "https://example.test/login"
	if !inst.HandleAction(action.NewActionWithPayload(action.ActionPaste, pasted)) {
		t.Fatal("HandleAction(ActionPaste) should return true")
	}
	if got := inst.GetValue(); got != pasted {
		t.Fatalf("value = %q, want pasted URL", got)
	}
}
