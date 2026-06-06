package ui

import (
	"strings"
	"time"

	"github.com/wwsheng009/mint/runtime/intent"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	menucomp "github.com/wwsheng009/mint/ui/components/menu"
	"github.com/wwsheng009/mint/ui/components/toolbar"
)

type ToolbarBuilder = toolbar.Builder
type ToolbarVNode = toolbar.VNode
type ToolbarItem = toolbar.Item
type ToolbarItemKind = toolbar.ItemKind
type ToolbarSelectionConfig = toolbar.SelectionConfig
type ToolbarPaginationConfig = toolbar.PaginationConfig
type ToolbarPaginationScope = toolbar.PaginationScope
type ToolbarActionGroupConfig = toolbar.ActionGroupConfig
type ToolbarPageSummaryPart = toolbar.PageSummaryPart
type ToolbarActionTargetPart = toolbar.ActionTargetPart

const (
	ToolbarItemText      = toolbar.ItemText
	ToolbarItemBadge     = toolbar.ItemBadge
	ToolbarItemButton    = toolbar.ItemButton
	ToolbarItemMenu      = toolbar.ItemMenu
	ToolbarItemSeparator = toolbar.ItemSeparator
	ToolbarItemCustom    = toolbar.ItemCustom
)

// NewToolbarBuilder creates a Toolbar builder.
func NewToolbarBuilder() *toolbar.Builder {
	return toolbar.NewBuilder()
}

// Toolbar creates a Toolbar from left, center, and right items.
func Toolbar(left, center, right []toolbar.Item) rtui.VNode {
	return toolbar.Of(left, center, right)
}

// ToolbarText creates a plain toolbar item.
func ToolbarText(key, label string) toolbar.Item {
	return toolbar.Text(key, label)
}

// ToolbarBadge creates a highlighted toolbar item.
func ToolbarBadge(key, label string) toolbar.Item {
	return toolbar.Badge(key, label)
}

// ToolbarButton creates a command toolbar item.
func ToolbarButton(key, label string, pressIntent intent.Intent) toolbar.Item {
	return toolbar.Button(key, label, pressIntent)
}

// ToolbarDropdown creates a controlled toolbar menu item.
func ToolbarDropdown(key, label string, items []menucomp.MenuItem, open bool) toolbar.Item {
	return toolbar.Dropdown(key, label, items, open)
}

// ToolbarSeparator creates a toolbar separator item.
func ToolbarSeparator(key string) toolbar.Item {
	return toolbar.Separator(key)
}

// ToolbarCustom creates a custom toolbar item.
func ToolbarCustom(key string, node rtui.VNode) toolbar.Item {
	return toolbar.Custom(key, node)
}

// ToolbarKeyValue creates a compact "label: value" toolbar item.
func ToolbarKeyValue(key, label, value string) toolbar.Item {
	return toolbar.KeyValue(key, label, value)
}

// ToolbarMutedKeyValue creates a low-emphasis "label: value" toolbar item.
func ToolbarMutedKeyValue(key, label, value string) toolbar.Item {
	return toolbar.MutedKeyValue(key, label, value)
}

// ToolbarStateBadge creates a highlighted toolbar item using operational tone mapping.
func ToolbarStateBadge(key, status string) toolbar.Item {
	return toolbar.StateBadge(key, status)
}

// ToolbarBusyBadge creates a warning-colored toolbar item for running operations.
func ToolbarBusyBadge(key, label string) toolbar.Item {
	return toolbar.BusyBadge(key, label)
}

// ToolbarErrorBadge creates an error-colored toolbar item.
func ToolbarErrorBadge(key, label string) toolbar.Item {
	return toolbar.ErrorBadge(key, label)
}

// ToolbarEndpoint creates a standard toolbar item for the active API endpoint.
func ToolbarEndpoint(value string) toolbar.Item {
	return toolbar.Endpoint(value)
}

// ToolbarScope creates a standard toolbar item for the current page scope.
func ToolbarScope(value string) toolbar.Item {
	return toolbar.Scope(value)
}

// ToolbarSelection creates a low-emphasis toolbar item for the current selection.
func ToolbarSelection(value string) toolbar.Item {
	return toolbar.Selection(value)
}

// ToolbarJoinDisabledReasons joins non-empty disabled reasons for operation controls.
func ToolbarJoinDisabledReasons(reasons ...string) string {
	return toolbar.JoinDisabledReasons(reasons...)
}

// ToolbarShellHeader creates a standard top-level operations shell header.
func ToolbarShellHeader(key, title, endpoint, auth string, actions ...toolbar.Item) rtui.VNode {
	return toolbar.ShellHeader(key, title, endpoint, auth, actions...)
}

// ToolbarShellNav creates a dense top-level navigation toolbar.
func ToolbarShellNav(key string, items []toolbar.Item) rtui.VNode {
	return toolbar.ShellNav(key, items)
}

// ToolbarPageSummary creates a standard pagination summary toolbar item.
func ToolbarPageSummary(page, totalPages, total, pageSize int, extraLabel, extraValue string, width int) toolbar.Item {
	if strings.TrimSpace(extraLabel) == "" {
		return toolbar.PageSummary(page, totalPages, total, pageSize, width)
	}
	return toolbar.PageSummary(page, totalPages, total, pageSize, width, toolbar.PageSummaryPart{Label: extraLabel, Value: extraValue})
}

// ToolbarPageSummaryWithParts creates a pagination summary with multiple scope
// context segments, for example provider plus search.
func ToolbarPageSummaryWithParts(page, totalPages, total, pageSize, width int, parts ...toolbar.PageSummaryPart) toolbar.Item {
	return toolbar.PageSummary(page, totalPages, total, pageSize, width, parts...)
}

// ToolbarCompactPageSummaryPart creates a display-width-bounded pagination
// scope segment for search/provider/status context.
func ToolbarCompactPageSummaryPart(label, value string, maxWidth int) toolbar.PageSummaryPart {
	return toolbar.CompactPageSummaryPart(label, value, maxWidth)
}

// ToolbarPageSummaryPartIfValue creates a pagination scope segment only when
// value is non-empty after trimming.
func ToolbarPageSummaryPartIfValue(label, value string) toolbar.PageSummaryPart {
	return toolbar.PageSummaryPartIfValue(label, value)
}

// ToolbarPageSummaryPartUnless creates a pagination scope segment only when
// value is non-empty and differs from defaultValue after trimming.
func ToolbarPageSummaryPartUnless(label, value, defaultValue string) toolbar.PageSummaryPart {
	return toolbar.PageSummaryPartUnless(label, value, defaultValue)
}

// ToolbarCompactPageSummaryPartIfValue creates a display-width-bounded
// pagination scope segment only when value is non-empty after trimming.
func ToolbarCompactPageSummaryPartIfValue(label, value string, maxWidth int) toolbar.PageSummaryPart {
	return toolbar.CompactPageSummaryPartIfValue(label, value, maxWidth)
}

// ToolbarCompactPageSummaryPartUnless creates a display-width-bounded
// pagination scope segment only when value is non-empty and differs from
// defaultValue.
func ToolbarCompactPageSummaryPartUnless(label, value, defaultValue string, maxWidth int) toolbar.PageSummaryPart {
	return toolbar.CompactPageSummaryPartUnless(label, value, defaultValue, maxWidth)
}

// ToolbarActionTargetSummary creates a compact operation target summary for
// action groups and dangerous operation surfaces.
func ToolbarActionTargetSummary(parts ...toolbar.ActionTargetPart) string {
	return toolbar.ActionTargetSummary(parts...)
}

// ToolbarActionTargetSummaryWithScope creates a compact operation target
// summary with an explicit active operation scope prefix.
func ToolbarActionTargetSummaryWithScope(scope string, parts ...toolbar.ActionTargetPart) string {
	return toolbar.ActionTargetSummaryWithScope(scope, parts...)
}

// ToolbarActionTargetSummaryWithScopes creates a compact operation target
// summary for action areas that expose one or more operation scopes.
func ToolbarActionTargetSummaryWithScopes(scopes []string, parts ...toolbar.ActionTargetPart) string {
	return toolbar.ActionTargetSummaryWithScopes(scopes, parts...)
}

// ToolbarCompactActionTargetPart creates a display-width-bounded operation
// target summary segment.
func ToolbarCompactActionTargetPart(label, value string, maxWidth int) toolbar.ActionTargetPart {
	return toolbar.CompactActionTargetPart(label, value, maxWidth)
}

// ToolbarPaginationScopeOf normalizes pagination numbers and applies a fallback
// total when the server total is not available.
func ToolbarPaginationScopeOf(page, pageSize, total, totalPages, fallbackTotal int) toolbar.PaginationScope {
	return toolbar.NormalizePaginationScope(page, pageSize, total, totalPages, fallbackTotal)
}

// ToolbarPaginationControlsWithScope creates standard pagination controls from
// normalized page numbers and scope summary parts.
func ToolbarPaginationControlsWithScope(key string, width int, scope toolbar.PaginationScope, prevIntent, nextIntent intent.Intent, busy bool, loadingReason string, summaryWidth int, parts ...toolbar.PageSummaryPart) rtui.VNode {
	return toolbar.PaginationControlsWithScope(key, width, scope, prevIntent, nextIntent, busy, loadingReason, summaryWidth, parts...)
}

// ToolbarLastSync creates a muted toolbar item for the latest successful sync time.
func ToolbarLastSync(syncAt time.Time) toolbar.Item {
	value := "-"
	if !syncAt.IsZero() {
		value = syncAt.Format("15:04:05")
	}
	return ToolbarText("last-sync", "last sync: "+value).WithWidth(22).WithForeground("bright-black")
}

// ToolbarSyncAge creates a toolbar item for elapsed time since the latest sync.
func ToolbarSyncAge(key, label string, syncAt, now time.Time, width int) toolbar.Item {
	if key == "" {
		key = "sync-age"
	}
	if label == "" {
		label = "age"
	}
	if width <= 0 {
		width = 13
	}
	if syncAt.IsZero() {
		return ToolbarText(key, label+": -").WithWidth(width).WithForeground("bright-black")
	}
	return ToolbarCustom(key, OperationElapsedTimerWithKey(key, label, syncAt, now, width))
}

// ToolbarPageHeader creates a standard page-level operations toolbar.
//
// It is intended for data and operations pages that need the same first-step
// context: title, short scope/subtitle, latest successful sync time, and sync age.
func ToolbarPageHeader(key, title, subtitle string, syncAt, now time.Time) rtui.VNode {
	key = strings.TrimSpace(key)
	if key == "" {
		key = "page.toolbar"
	}
	subtitle = strings.TrimSpace(subtitle)
	if subtitle == "" {
		subtitle = "-"
	}
	return NewToolbarBuilder().
		Key(key).
		Title(title).
		TitleWidth(18).
		Left(ToolbarText("subtitle", subtitle).WithWidth(46).WithForeground("bright-black")).
		Right(ToolbarLastSync(syncAt)).
		Right(ToolbarSyncAge(toolbarPageSyncAgeKey(key), "age", syncAt, now, 13)).
		Build()
}

func toolbarPageSyncAgeKey(key string) string {
	base := strings.TrimSuffix(strings.TrimSpace(key), ".toolbar")
	if base == "" {
		return "page.sync-age"
	}
	return base + ".sync-age"
}

// ToolbarPaginationPrev creates a Prev pagination action with standard disabled reasons.
func ToolbarPaginationPrev(pressIntent intent.Intent, busy bool, page int, loadingReason string) toolbar.Item {
	return toolbar.PaginationPrev(pressIntent, busy, page, loadingReason)
}

// ToolbarPaginationNext creates a Next pagination action with standard disabled reasons.
func ToolbarPaginationNext(pressIntent intent.Intent, busy bool, page, totalPages int, loadingReason string) toolbar.Item {
	return toolbar.PaginationNext(pressIntent, busy, page, totalPages, loadingReason)
}

// ToolbarPaginationControls creates a standard pagination toolbar.
func ToolbarPaginationControls(key string, width int, summary toolbar.Item, prevIntent, nextIntent intent.Intent, busy bool, page, totalPages int, loadingReason string) rtui.VNode {
	return toolbar.PaginationControls(key, width, summary, prevIntent, nextIntent, busy, page, totalPages, loadingReason)
}

// ToolbarSelectionPrev creates an Up selection action with standard disabled reasons.
func ToolbarSelectionPrev(pressIntent intent.Intent, busy bool, index, total int, itemLabel, loadingReason string) toolbar.Item {
	return toolbar.SelectionPrev(pressIntent, busy, index, total, itemLabel, loadingReason)
}

// ToolbarSelectionNext creates a Down selection action with standard disabled reasons.
func ToolbarSelectionNext(pressIntent intent.Intent, busy bool, index, total int, itemLabel, loadingReason string) toolbar.Item {
	return toolbar.SelectionNext(pressIntent, busy, index, total, itemLabel, loadingReason)
}

// ToolbarSelectionControls creates a standard Up/Down selection toolbar.
func ToolbarSelectionControls(key, title string, prevIntent, nextIntent intent.Intent, busy bool, index, total int, itemLabel, loadingReason string) rtui.VNode {
	return toolbar.SelectionControls(key, title, prevIntent, nextIntent, busy, index, total, itemLabel, loadingReason)
}

// ToolbarSelectionActionControls creates a standard selection toolbar with contextual actions.
func ToolbarSelectionActionControls(key, title string, width int, prevIntent, nextIntent intent.Intent, busy bool, index, total int, itemLabel, loadingReason string, actions ...toolbar.Item) rtui.VNode {
	return toolbar.SelectionActionControls(key, title, width, prevIntent, nextIntent, busy, index, total, itemLabel, loadingReason, actions...)
}

// ToolbarActionControls creates a compact action toolbar from caller-supplied items.
func ToolbarActionControls(key string, width int, items ...toolbar.Item) rtui.VNode {
	return toolbar.ActionControls(key, width, items...)
}

// ToolbarActionGroup creates a standard titled action toolbar.
func ToolbarActionGroup(key, title string, items []toolbar.Item) rtui.VNode {
	return toolbar.ActionGroup(key, title, items)
}

// ToolbarActionGroupWithLayout creates a titled action toolbar with explicit
// width and title width.
func ToolbarActionGroupWithLayout(key, title string, width, titleWidth int, items []toolbar.Item) rtui.VNode {
	return toolbar.ActionGroupWithLayout(key, title, width, titleWidth, items)
}

// ToolbarActionGroups creates a grouped operation surface with optional summary text.
func ToolbarActionGroups(key string, groups []toolbar.ActionGroupConfig, summary, emptyText string) rtui.VNode {
	return toolbar.ActionGroups(key, groups, summary, emptyText)
}
