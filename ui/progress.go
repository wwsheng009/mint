package ui

import (
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/progress"
)

type ProgressStatus = progress.Status
type ProgressGlyphStyle = progress.GlyphStyle

const (
	ProgressStatusNormal    = progress.StatusNormal
	ProgressStatusSuccess   = progress.StatusSuccess
	ProgressStatusException = progress.StatusException
	ProgressStatusActive    = progress.StatusActive
	ProgressStatusWarning   = progress.StatusWarning

	ProgressGlyphStyleUnicode = progress.GlyphStyleUnicode
	ProgressGlyphStyleASCII   = progress.GlyphStyleASCII
)

// ProgressStatusForState maps common operational state strings to a Progress status.
func ProgressStatusForState(state string) ProgressStatus {
	return progress.StatusForState(state)
}

// ProgressForState creates a progress bar using common operational state semantics.
func ProgressForState(label string, value, max int, state string) rtui.VNode {
	return progress.ForState(label, value, max, state)
}

// ProgressForStateWithValue creates a semantic progress bar that also displays value/max.
func ProgressForStateWithValue(label string, value, max int, state, unit string) rtui.VNode {
	return progress.ForStateWithValue(label, value, max, state, unit)
}

// ProgressOperationalValue creates a keyed, fixed-width operational progress bar with value labels.
func ProgressOperationalValue(key, label string, value, max int, state, unit string, width int) rtui.VNode {
	return progress.OperationalValue(key, label, value, max, state, unit, width)
}

// ProgressUsage creates a resource usage progress bar with default 80/95 thresholds.
func ProgressUsage(label string, used, total int) rtui.VNode {
	return progress.Usage(label, used, total)
}

// ProgressUsageWithThresholds creates a resource usage progress bar.
func ProgressUsageWithThresholds(label string, used, total, warnAt, criticalAt int) rtui.VNode {
	return progress.UsageWithThresholds(label, used, total, warnAt, criticalAt)
}

// ProgressUsageWithValue creates a resource usage progress bar that also displays used/total.
func ProgressUsageWithValue(label string, used, total int, unit string) rtui.VNode {
	return progress.UsageWithValue(label, used, total, unit)
}

// ProgressUsageWithValueThresholds creates a resource usage progress bar with custom thresholds and value labels.
func ProgressUsageWithValueThresholds(label string, used, total int, unit string, warnAt, criticalAt int) rtui.VNode {
	return progress.UsageWithValueThresholds(label, used, total, unit, warnAt, criticalAt)
}

// ProgressBusy creates an indeterminate progress bar for work without a known total.
func ProgressBusy(label string) rtui.VNode {
	return progress.Busy(label)
}

// ProgressBusyWithKey creates a keyed, fixed-width indeterminate progress bar.
func ProgressBusyWithKey(key, label string, width int) rtui.VNode {
	return progress.BusyWithKey(key, label, width)
}

// ProgressComplete creates a complete success progress bar.
func ProgressComplete(label string) rtui.VNode {
	return progress.Complete(label)
}

// ProgressFailed creates a complete exception progress bar.
func ProgressFailed(label string) rtui.VNode {
	return progress.Failed(label)
}
