package progress

import "strings"

const (
	defaultUsageWarnAt     = 80
	defaultUsageCriticalAt = 95
)

// StatusForState maps common operational state strings to Progress status values.
func StatusForState(state string) Status {
	normalized := strings.ToLower(strings.TrimSpace(strings.NewReplacer(
		"\r\n", " ",
		"\n", " ",
		"\r", " ",
		"\t", " ",
	).Replace(state)))
	normalized = strings.ReplaceAll(normalized, " ", "_")

	switch normalized {
	case "healthy", "active", "available", "effective", "enabled", "success", "ok", "ready", "complete", "completed", "in_sync":
		return StatusSuccess
	case "processing", "loading", "syncing", "refreshing", "reloading", "running", "busy", "executing":
		return StatusActive
	case "degraded", "rate_limited", "limited", "pending_restart", "pending", "warning", "cooldown", "partial", "lagging", "retrying", "queued":
		return StatusWarning
	case "unhealthy", "disabled", "unauthorized", "unavailable", "failed", "failure", "error", "down", "blocked", "out_of_sync", "exception":
		return StatusException
	default:
		return StatusNormal
	}
}

// ForState creates a progress bar using common operational state semantics.
func ForState(label string, value, max int, state string) *VNode {
	status := StatusForState(state)
	builder := NewBuilder().
		Label(label).
		Value(value).
		Max(max).
		Status(status)
	if status == StatusActive && max <= 0 {
		builder.Indeterminate()
	}
	return builder.BuildTyped()
}

// Usage creates a progress bar for resource usage with default 80/95 thresholds.
func Usage(label string, used, total int) *VNode {
	return UsageWithThresholds(label, used, total, defaultUsageWarnAt, defaultUsageCriticalAt)
}

// UsageWithThresholds creates a resource usage progress bar.
func UsageWithThresholds(label string, used, total, warnAt, criticalAt int) *VNode {
	if warnAt <= 0 {
		warnAt = defaultUsageWarnAt
	}
	if criticalAt <= 0 {
		criticalAt = defaultUsageCriticalAt
	}
	if criticalAt < warnAt {
		criticalAt = warnAt
	}

	value, max := normalizeProgressRange(used, total)
	if max <= 0 {
		return NewBuilder().
			Label(label).
			Value(0).
			Max(100).
			Status(StatusNormal).
			BuildTyped()
	}

	status := StatusNormal
	percent := progressPercent(value, max)
	switch {
	case percent >= criticalAt:
		status = StatusException
	case percent >= warnAt:
		status = StatusWarning
	}

	return NewBuilder().
		Label(label).
		Value(value).
		Max(max).
		Status(status).
		BuildTyped()
}

// Busy creates an indeterminate progress bar for work without a known total.
func Busy(label string) *VNode {
	return NewBuilder().Label(label).Indeterminate().BuildTyped()
}

// Complete creates a complete success progress bar.
func Complete(label string) *VNode {
	return NewBuilder().Label(label).Value(100).Max(100).Success().BuildTyped()
}

// Failed creates a complete exception progress bar.
func Failed(label string) *VNode {
	return NewBuilder().Label(label).Value(100).Max(100).Exception().BuildTyped()
}
