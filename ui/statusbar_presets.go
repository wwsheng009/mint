package ui

import (
	"time"

	"github.com/wwsheng009/mint/ui/components/statusbar"
)

// StatusBarEndpoint creates a standard status bar section for the active API endpoint.
func StatusBarEndpoint(value string) statusbar.Section {
	return statusbar.Endpoint(value)
}

// StatusBarUser creates a standard status bar section for the active user/session label.
func StatusBarUser(value string) statusbar.Section {
	return statusbar.User(value)
}

// StatusBarPage creates a standard status bar section for the active page or panel.
func StatusBarPage(value string) statusbar.Section {
	return statusbar.Page(value)
}

// StatusBarSelection creates a low-emphasis status bar section for the current selection.
func StatusBarSelection(value string) statusbar.Section {
	return statusbar.Selection(value)
}

// StatusBarLastSync creates a low-emphasis status bar section for the last successful sync time.
func StatusBarLastSync(syncAt, now time.Time) statusbar.Section {
	return statusbar.LastSync(syncAt, now)
}

// StatusBarAutoRefresh creates a low-emphasis status bar section for a refresh countdown.
func StatusBarAutoRefresh(remaining time.Duration) statusbar.Section {
	return statusbar.AutoRefresh(remaining)
}
