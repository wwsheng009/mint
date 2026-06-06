package ui

import (
	"strings"
	"time"

	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/statusbar"
)

// StatusBarEndpoint creates a standard status bar section for the active API endpoint.
func StatusBarEndpoint(value string) statusbar.Section {
	return statusbar.Endpoint(value)
}

// StatusBarProfile creates a standard status bar section for the active endpoint/profile name.
func StatusBarProfile(value string) statusbar.Section {
	return statusbar.Profile(value)
}

// StatusBarUser creates a standard status bar section for the active user/session label.
func StatusBarUser(value string) statusbar.Section {
	return statusbar.User(value)
}

// StatusBarRole creates a standard status bar section for the active permission or role label.
func StatusBarRole(value string) statusbar.Section {
	return statusbar.Role(value)
}

// StatusBarPage creates a standard status bar section for the active page or panel.
func StatusBarPage(value string) statusbar.Section {
	return statusbar.Page(value)
}

// StatusBarScope creates a low-emphasis status bar section for the active operational scope.
func StatusBarScope(value string) statusbar.Section {
	return statusbar.Scope(value)
}

// StatusBarTarget creates a low-emphasis status bar section for the operation target.
func StatusBarTarget(value string) statusbar.Section {
	return statusbar.Target(value)
}

// StatusBarSelection creates a low-emphasis status bar section for the current selection.
func StatusBarSelection(value string) statusbar.Section {
	return statusbar.Selection(value)
}

// StatusBarSelectionTarget creates a compact "kind value" selection label.
func StatusBarSelectionTarget(kind, value string, maxValueWidth int) string {
	return statusbar.SelectionTarget(kind, value, maxValueWidth)
}

// StatusBarFilter creates a low-emphasis status bar section for active filters or search criteria.
func StatusBarFilter(value string) statusbar.Section {
	return statusbar.Filter(value)
}

// StatusBarCount creates a compact numeric status bar section.
func StatusBarCount(label string, count int) statusbar.Section {
	return statusbar.Count(label, count)
}

// StatusBarLatency creates a low-emphasis status bar section for request or refresh latency.
func StatusBarLatency(value time.Duration) statusbar.Section {
	return statusbar.Latency(value)
}

// StatusBarUptime creates a low-emphasis status bar section for process/runtime uptime.
func StatusBarUptime(value time.Duration) statusbar.Section {
	return statusbar.Uptime(value)
}

// StatusBarHotkey creates a low-emphasis keyboard shortcut hint section.
func StatusBarHotkey(key, label string) statusbar.Section {
	return statusbar.Hotkey(key, label)
}

// StatusBarSeparator creates a low-emphasis visual divider section.
func StatusBarSeparator() statusbar.Section {
	return statusbar.Separator()
}

// StatusBarLastSync creates a low-emphasis status bar section for the last successful sync time.
func StatusBarLastSync(syncAt, now time.Time) statusbar.Section {
	return statusbar.LastSync(syncAt, now)
}

// StatusBarAutoRefresh creates a low-emphasis status bar section for a refresh countdown.
func StatusBarAutoRefresh(remaining time.Duration) statusbar.Section {
	return statusbar.AutoRefresh(remaining)
}

// OperationalStatusBar creates a standard footer for operational pages.
//
// The layout follows the common operator sequence and inserts visible dividers
// so adjacent sections remain readable even when terminal copy or rendering
// collapses layout-only gaps.
func OperationalStatusBar(mode statusbar.Section, endpoint, user, page string, lastSyncAt, now time.Time, selection string, hints ...statusbar.Section) rtui.VNode {
	if mode.Text == "" {
		mode = StatusBarStateBadge("ready")
	}
	leftSections := []statusbar.Section{mode}
	if strings.TrimSpace(endpoint) != "" {
		leftSections = append(leftSections, StatusBarEndpoint(endpoint).WithHelp("Current Admin API endpoint"))
	}
	leftSections = append(leftSections,
		StatusBarUser(user).WithHelp("Current local session label"),
		StatusBarPage(page).WithHelp("Current manager page"),
	)
	rightSections := StatusBarSections(
		StatusBarSelection(selection).WithHelp("Current selected operation target"),
	)
	rightSections = append(rightSections, hints...)
	return NewStatusBarBuilder().
		Theme(StatusBarThemeDefault()).
		LeftSections(statusBarSeparatedSections(leftSections...)...).
		CenterSections(statusBarLeadingSeparatedSections(
			StatusBarLastSync(lastSyncAt, now).WithHelp("Latest successful Admin API refresh time"),
		)...).
		RightSections(statusBarLeadingSeparatedSections(rightSections...)...).
		Build()
}

func statusBarSeparatedSections(sections ...statusbar.Section) []statusbar.Section {
	if len(sections) == 0 {
		return nil
	}
	result := make([]statusbar.Section, 0, len(sections)*2-1)
	for _, section := range sections {
		if section.Text == "" {
			continue
		}
		if len(result) > 0 {
			result = append(result, StatusBarSeparator())
		}
		result = append(result, section)
	}
	return result
}

func statusBarLeadingSeparatedSections(sections ...statusbar.Section) []statusbar.Section {
	result := statusBarSeparatedSections(sections...)
	if len(result) == 0 {
		return nil
	}
	return append([]statusbar.Section{StatusBarSeparator()}, result...)
}
