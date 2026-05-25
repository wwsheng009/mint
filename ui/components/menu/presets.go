package menu

import (
	"strings"

	"github.com/wwsheng009/mint/runtime/intent"
)

// ActionWithDescription creates an action item with a secondary explanation.
func ActionWithDescription(key, label, description string, pressIntent intent.Intent) MenuItem {
	return Action(key, label, pressIntent).WithDescription(description)
}

// DangerAction creates a dangerous action item with danger styling and a description.
func DangerAction(key, label, description string, pressIntent intent.Intent) MenuItem {
	return ActionWithDescription(key, label, description, pressIntent).WithDanger(true)
}

// DisabledAction creates a disabled action item with a visible disabled reason.
func DisabledAction(key, label, reason string, pressIntent intent.Intent) MenuItem {
	return Action(key, label, pressIntent).WithDisabledReason(reason)
}

// RefreshAction creates a standard refresh menu item.
func RefreshAction(pressIntent intent.Intent) MenuItem {
	return ActionWithDescription("refresh", "Refresh", "Refresh current data", pressIntent).
		WithShortcut("r")
}

// ReloadRuntimeAction creates a standard runtime reload menu item.
func ReloadRuntimeAction(pressIntent intent.Intent) MenuItem {
	return DangerAction("reload-runtime", "Reload Runtime", "Reload runtime configuration", pressIntent).
		WithShortcut("ctrl+r")
}

// ResetRuntimeAction creates a standard runtime reset menu item.
func ResetRuntimeAction(pressIntent intent.Intent) MenuItem {
	return DangerAction("reset-runtime", "Reset Runtime", "Reset runtime state for the selected scope", pressIntent)
}

// ClearCircuitBreakersAction creates a standard clear-circuit-breakers menu item.
func ClearCircuitBreakersAction(pressIntent intent.Intent) MenuItem {
	return DangerAction("clear-circuit-breakers", "Clear Circuit Breakers", "Clear circuit breaker state", pressIntent)
}

// Group creates a label followed by a set of menu items.
func Group(label string, items ...MenuItem) []MenuItem {
	label = strings.TrimSpace(label)
	if label == "" {
		return Items(items...)
	}
	group := make([]MenuItem, 0, len(items)+1)
	group = append(group, LabelItem(slugKey(label), label))
	group = append(group, Items(items...)...)
	return group
}

// AppendGroup appends a labeled group to an existing menu item slice.
func AppendGroup(dst []MenuItem, label string, items ...MenuItem) []MenuItem {
	return append(dst, Group(label, items...)...)
}

func slugKey(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return "group"
	}
	replacer := strings.NewReplacer(" ", "-", "_", "-", "/", "-", "\\", "-", ":", "-", ".", "-")
	value = replacer.Replace(value)
	value = strings.Trim(value, "-")
	if value == "" {
		return "group"
	}
	return "group-" + value
}
