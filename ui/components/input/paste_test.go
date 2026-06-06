package input

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/action"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

func TestInstance_HandleAction_PasteBypassesWidthLimit(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"width": 4,
	})

	pasted := "https://example.test/login"
	if !inst.HandleAction(action.NewActionWithPayload(action.ActionPaste, pasted)) {
		t.Fatal("HandleAction(ActionPaste) should return true")
	}
	if got := inst.GetValue(); got != pasted {
		t.Fatalf("value = %q, want pasted URL", got)
	}
}
