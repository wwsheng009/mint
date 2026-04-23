package scroll

import (
	"github.com/wwsheng009/mint/runtime/action"
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
)

func DeltaFromAction(act *action.Action) (int, bool) {
	if act == nil {
		return 0, false
	}

	if delta, ok := act.GetPayloadInt(); ok {
		return delta, true
	}

	switch payload := act.Payload.(type) {
	case *runtimemsg.MouseMsg:
		if payload != nil && payload.Delta != 0 {
			return -payload.Delta, true
		}
	case runtimemsg.MouseMsg:
		if payload.Delta != 0 {
			return -payload.Delta, true
		}
	}

	return 0, false
}
