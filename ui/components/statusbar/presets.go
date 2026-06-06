package statusbar

import (
	"fmt"
	"strings"
	"time"

	textcomp "github.com/wwsheng009/mint/ui/components/text"
)

// Tone describes an operational status severity for status bar presets.
type Tone string

const (
	ToneNormal  Tone = "normal"
	ToneWarn    Tone = "warn"
	ToneError   Tone = "error"
	ToneNeutral Tone = "neutral"
	ToneInfo    Tone = "info"
)

// KeyValue creates a compact "label: value" status section.
func KeyValue(label, value string) Section {
	label = strings.TrimSpace(normalizeStatusText(label))
	value = strings.TrimSpace(normalizeStatusText(value))
	if value == "" {
		value = "-"
	}
	if label == "" {
		return Text(value)
	}
	return Text(label + ": " + value)
}

// MutedKeyValue creates a compact low-emphasis "label: value" status section.
func MutedKeyValue(label, value string) Section {
	return KeyValue(label, value).WithForeground("bright-black")
}

// StateBadge creates a highlighted status section using common operational tone mapping.
func StateBadge(status string) Section {
	label := strings.TrimSpace(normalizeStatusText(status))
	if label == "" {
		label = "-"
	}
	fg, bg := toneColors(DefaultTone(label))
	return Badge(" "+label+" ", fg, bg)
}

// BusyBadge creates a warning-colored status section for running operations.
func BusyBadge(label string) Section {
	label = strings.TrimSpace(normalizeStatusText(label))
	if label == "" {
		label = "busy"
	}
	return Badge(" "+label+" ", "black", "yellow")
}

// ErrorBadge creates an error-colored status section.
func ErrorBadge(label string) Section {
	label = strings.TrimSpace(normalizeStatusText(label))
	if label == "" {
		label = "error"
	}
	return Badge(" "+label+" ", "white", "red")
}

// Endpoint creates a standard status section for the active API endpoint.
func Endpoint(value string) Section {
	return KeyValue("endpoint", value)
}

// Profile creates a standard status section for the active endpoint/profile name.
func Profile(value string) Section {
	return KeyValue("profile", value)
}

// User creates a standard status section for the active user/session label.
func User(value string) Section {
	return KeyValue("user", value)
}

// Role creates a standard status section for the active permission or role label.
func Role(value string) Section {
	return KeyValue("role", value)
}

// Page creates a standard status section for the active page or panel.
func Page(value string) Section {
	return KeyValue("page", value)
}

// Scope creates a low-emphasis status section for the active operational scope.
func Scope(value string) Section {
	return MutedKeyValue("scope", value)
}

// Target creates a low-emphasis status section for the operation target.
func Target(value string) Section {
	return MutedKeyValue("target", value)
}

// Selection creates a low-emphasis status section for the current selection.
func Selection(value string) Section {
	return MutedKeyValue("selection", value)
}

// SelectionTarget creates a compact selection value in "kind value" form. The
// caller owns sensitive-value masking; this helper only normalizes blank kind or
// value and applies display-width bounds to the value portion.
func SelectionTarget(kind, value string, maxValueWidth int) string {
	kind = strings.TrimSpace(normalizeStatusText(kind))
	value = strings.TrimSpace(normalizeStatusText(value))
	if value == "" {
		return "-"
	}
	value = textcomp.CompactFallbackText(value, "-", maxValueWidth)
	if kind == "" {
		return value
	}
	return kind + " " + value
}

// Filter creates a low-emphasis status section for active filters or search criteria.
func Filter(value string) Section {
	return MutedKeyValue("filter", value)
}

// Count creates a compact numeric "label: count" status section.
func Count(label string, count int) Section {
	if count < 0 {
		count = 0
	}
	return KeyValue(label, fmt.Sprintf("%d", count))
}

// Latency creates a low-emphasis status section for request or refresh latency.
func Latency(value time.Duration) Section {
	return MutedKeyValue("latency", formatDuration(value))
}

// Uptime creates a low-emphasis status section for process/runtime uptime.
func Uptime(value time.Duration) Section {
	return MutedKeyValue("uptime", formatDuration(value))
}

// Hotkey creates a low-emphasis status section that shows a keyboard shortcut hint.
func Hotkey(key, label string) Section {
	key = strings.TrimSpace(normalizeStatusText(key))
	label = strings.TrimSpace(normalizeStatusText(label))
	if key == "" {
		key = "-"
	}
	if label == "" {
		return Text(key).WithForeground("bright-black")
	}
	return Text(key + " " + label).WithForeground("bright-black")
}

// Separator creates a low-emphasis visual divider between compact status sections.
func Separator() Section {
	return Text(" | ").WithForeground("bright-black")
}

// LastSync creates a low-emphasis status section for the last successful sync time.
func LastSync(syncAt, now time.Time) Section {
	return MutedKeyValue("last sync", formatElapsed(syncAt, now))
}

// AutoRefresh creates a low-emphasis status section for a refresh countdown.
func AutoRefresh(remaining time.Duration) Section {
	return MutedKeyValue("refresh", formatRemaining(remaining))
}

// DefaultTone maps common operational status strings to semantic tones.
func DefaultTone(status string) Tone {
	normalized := strings.ToLower(strings.TrimSpace(normalizeStatusText(status)))
	normalized = strings.ReplaceAll(normalized, " ", "_")
	switch normalized {
	case "healthy", "active", "available", "effective", "enabled", "success", "ok", "ready", "running", "in_sync":
		return ToneNormal
	case "degraded", "rate_limited", "limited", "pending_restart", "pending", "warning", "cooldown", "partial", "lagging", "reloading":
		return ToneWarn
	case "unhealthy", "disabled", "unauthorized", "unavailable", "failed", "failure", "error", "down", "blocked", "out_of_sync":
		return ToneError
	case "info", "processing", "loading", "syncing", "refreshing":
		return ToneInfo
	default:
		return ToneNeutral
	}
}

func toneColors(tone Tone) (fgColor, bgColor string) {
	switch tone {
	case ToneNormal:
		return "black", "green"
	case ToneWarn:
		return "black", "yellow"
	case ToneError:
		return "white", "red"
	case ToneInfo:
		return "black", "cyan"
	default:
		return "bright-white", "bright-black"
	}
}

func formatElapsed(at, now time.Time) string {
	if at.IsZero() {
		return "never"
	}
	if now.IsZero() {
		now = time.Now()
	}
	elapsed := now.Sub(at)
	if elapsed < 0 {
		elapsed = 0
	}
	if elapsed < time.Second {
		return "just now"
	}
	if elapsed < time.Minute {
		return fmt.Sprintf("%ds ago", int(elapsed.Seconds()))
	}
	if elapsed < time.Hour {
		return fmt.Sprintf("%dm ago", int(elapsed.Minutes()))
	}
	if elapsed < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(elapsed.Hours()))
	}
	return fmt.Sprintf("%dd ago", int(elapsed.Hours()/24))
}

func formatRemaining(remaining time.Duration) string {
	if remaining <= 0 {
		return "now"
	}
	if remaining < time.Minute {
		return fmt.Sprintf("%ds", int(remaining.Seconds()))
	}
	if remaining < time.Hour {
		return fmt.Sprintf("%dm", int(remaining.Minutes()))
	}
	return fmt.Sprintf("%dh", int(remaining.Hours()))
}

func formatDuration(value time.Duration) string {
	if value < 0 {
		value = 0
	}
	if value < time.Millisecond {
		return fmt.Sprintf("%dus", value.Microseconds())
	}
	if value < time.Second {
		return fmt.Sprintf("%dms", value.Milliseconds())
	}
	if value < time.Minute {
		return fmt.Sprintf("%.1fs", value.Seconds())
	}
	if value < time.Hour {
		return fmt.Sprintf("%dm", int(value.Minutes()))
	}
	if value < 24*time.Hour {
		return fmt.Sprintf("%dh", int(value.Hours()))
	}
	return fmt.Sprintf("%dd", int(value.Hours()/24))
}
