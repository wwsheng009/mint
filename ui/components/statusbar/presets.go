package statusbar

import "strings"

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
