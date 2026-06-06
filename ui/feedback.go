package ui

import (
	rtui "github.com/wwsheng009/mint/runtime/ui"
	alertcomp "github.com/wwsheng009/mint/ui/components/alert"
	"github.com/wwsheng009/mint/ui/components/progress"
	textcomp "github.com/wwsheng009/mint/ui/components/text"
)

// OperationalFeedback renders a compact global operation feedback line.
//
// Priority is busy, error, notice, empty. This matches operational shells where
// in-flight work should be visible before stale messages.
func OperationalFeedback(key string, busy bool, busyLabel, errorText, noticeText string, width int) rtui.VNode {
	if busy {
		if busyLabel == "" {
			busyLabel = "loading"
		}
		if width <= 0 {
			width = 28
		}
		return progress.BusyWithKey(key, busyLabel, width)
	}
	if errorText != "" {
		return alertcomp.NewBuilder(errorText).Error().Build()
	}
	if noticeText != "" {
		return textcomp.NewBuilder(noticeText).FgColor("green").Build()
	}
	return textcomp.Text("")
}
