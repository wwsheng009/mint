package ui

import (
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/progress"
)

type ProgressStatus = progress.Status

const (
	ProgressStatusNormal    = progress.StatusNormal
	ProgressStatusSuccess   = progress.StatusSuccess
	ProgressStatusException = progress.StatusException
	ProgressStatusActive    = progress.StatusActive
	ProgressStatusWarning   = progress.StatusWarning
)

// ProgressStatusForState maps common operational state strings to a Progress status.
func ProgressStatusForState(state string) ProgressStatus {
	return progress.StatusForState(state)
}

// ProgressForState creates a progress bar using common operational state semantics.
func ProgressForState(label string, value, max int, state string) rtui.VNode {
	return progress.ForState(label, value, max, state)
}

// ProgressUsage creates a resource usage progress bar with default 80/95 thresholds.
func ProgressUsage(label string, used, total int) rtui.VNode {
	return progress.Usage(label, used, total)
}

// ProgressUsageWithThresholds creates a resource usage progress bar.
func ProgressUsageWithThresholds(label string, used, total, warnAt, criticalAt int) rtui.VNode {
	return progress.UsageWithThresholds(label, used, total, warnAt, criticalAt)
}

// ProgressBusy creates an indeterminate progress bar for work without a known total.
func ProgressBusy(label string) rtui.VNode {
	return progress.Busy(label)
}

// ProgressComplete creates a complete success progress bar.
func ProgressComplete(label string) rtui.VNode {
	return progress.Complete(label)
}

// ProgressFailed creates a complete exception progress bar.
func ProgressFailed(label string) rtui.VNode {
	return progress.Failed(label)
}
