package ui

import (
	"github.com/wwsheng009/mint/runtime/intent"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/filterbar"
)

type FilterBarBuilder = filterbar.Builder
type FilterBarVNode = filterbar.VNode
type FilterBarField = filterbar.Field
type FilterBarFieldKind = filterbar.FieldKind
type FilterBarAction = filterbar.Action
type FilterBarOption = filterbar.Option
type FilterBarSummaryPart = filterbar.SummaryPart
type FilterBarLookupSummaryConfig = filterbar.LookupSummaryConfig

const (
	FilterBarFieldText   = filterbar.FieldText
	FilterBarFieldSearch = filterbar.FieldSearch
	FilterBarFieldSelect = filterbar.FieldSelect
	FilterBarFieldCustom = filterbar.FieldCustom
)

// NewFilterBarBuilder creates a FilterBar builder.
func NewFilterBarBuilder() *filterbar.Builder {
	return filterbar.NewBuilder()
}

// FilterBar creates a FilterBar from fields.
func FilterBar(fields []filterbar.Field) rtui.VNode {
	return filterbar.Of(fields)
}

// FilterBarSearch creates a search filter field.
func FilterBarSearch(key, label, value string) filterbar.Field {
	return filterbar.Search(key, label, value)
}

// FilterBarText creates a text filter field.
func FilterBarText(key, label, value string) filterbar.Field {
	return filterbar.Text(key, label, value)
}

// FilterBarSelect creates a select filter field.
func FilterBarSelect(key, label string, options []filterbar.Option) filterbar.Field {
	return filterbar.Select(key, label, options)
}

// FilterBarCustom creates a custom filter field.
func FilterBarCustom(key, label string, node rtui.VNode) filterbar.Field {
	return filterbar.Custom(key, label, node)
}

// FilterBarButton creates a command action for a FilterBar.
func FilterBarButton(key, label string, pressIntent intent.Intent) filterbar.Action {
	return filterbar.Button(key, label, pressIntent)
}

// FilterBarJoinDisabledReasons joins non-empty disabled reasons for FilterBar actions.
func FilterBarJoinDisabledReasons(reasons ...string) string {
	return filterbar.JoinDisabledReasons(reasons...)
}

// FilterBarRefreshAction creates a primary Refresh action with standard loading disabled reason.
func FilterBarRefreshAction(pressIntent intent.Intent, busy bool, loadingReason string) filterbar.Action {
	return filterbar.RefreshAction(pressIntent, busy, loadingReason)
}

// FilterBarResetAction creates a secondary Reset action with standard loading disabled reason.
func FilterBarResetAction(pressIntent intent.Intent, busy bool, loadingReason string) filterbar.Action {
	return filterbar.ResetAction(pressIntent, busy, loadingReason)
}

// FilterBarResetActionWhenChanged creates a Reset action disabled until filters differ from defaults.
func FilterBarResetActionWhenChanged(pressIntent intent.Intent, changed bool, busy bool, loadingReason string) filterbar.Action {
	return filterbar.ResetActionWhenChanged(pressIntent, changed, busy, loadingReason)
}

// FilterBarClearFieldAction creates a secondary Clear action for a bound field.
func FilterBarClearFieldAction(fieldName, value string, busy bool, loadingReason string) filterbar.Action {
	return filterbar.ClearFieldAction(fieldName, value, busy, loadingReason)
}

// FilterBarClearFieldActionWithLabel creates a secondary Clear action with a
// caller-provided key and label for multi-field filter bars.
func FilterBarClearFieldActionWithLabel(key, label, fieldName, value string, busy bool, loadingReason string) filterbar.Action {
	return filterbar.ClearFieldActionWithLabel(key, label, fieldName, value, busy, loadingReason)
}

// FilterBarSummary joins operational filter summary segments with a stable separator.
func FilterBarSummary(parts ...filterbar.SummaryPart) string {
	return filterbar.Summary(parts...)
}

// FilterBarSummaryValue creates a "label value" filter summary segment.
func FilterBarSummaryValue(label, value string) filterbar.SummaryPart {
	return filterbar.SummaryValue(label, value)
}

// FilterBarSummaryValueUnless creates a filter summary segment only when value
// is non-empty and differs from defaultValue after trimming.
func FilterBarSummaryValueUnless(label, value, defaultValue string) filterbar.SummaryPart {
	return filterbar.SummaryValueUnless(label, value, defaultValue)
}

// FilterBarSummaryPresence creates a summary segment that reports whether value
// is present without displaying the value itself.
func FilterBarSummaryPresence(label, value, presentText, missingText string) filterbar.SummaryPart {
	return filterbar.SummaryPresence(label, value, presentText, missingText)
}

// FilterBarSummaryCount creates a non-negative integer count filter summary segment.
func FilterBarSummaryCount(label string, count int) filterbar.SummaryPart {
	return filterbar.SummaryCount(label, count)
}

// FilterBarSummaryRatio creates an "available/total" filter summary segment.
func FilterBarSummaryRatio(label string, available, total int) filterbar.SummaryPart {
	return filterbar.SummaryRatio(label, available, total)
}

// FilterBarSummarySearch creates a standard search filter summary segment.
func FilterBarSummarySearch(search string) filterbar.SummaryPart {
	return filterbar.SummarySearch(search)
}

// FilterBarSummaryCompactSearch creates a display-width-bounded search summary segment.
func FilterBarSummaryCompactSearch(search string, maxWidth int) filterbar.SummaryPart {
	return filterbar.SummaryCompactSearch(search, maxWidth)
}

// FilterBarSummaryCompact creates a display-width-bounded filter summary segment.
func FilterBarSummaryCompact(label, value string, maxWidth int) filterbar.SummaryPart {
	return filterbar.SummaryCompact(label, value, maxWidth)
}

// FilterBarSummaryCompactUnless creates a display-width-bounded filter summary
// segment only when value is non-empty and differs from defaultValue.
func FilterBarSummaryCompactUnless(label, value, defaultValue string, maxWidth int) filterbar.SummaryPart {
	return filterbar.SummaryCompactUnless(label, value, defaultValue, maxWidth)
}

// FilterBarPageSummary creates a standard page/total/search summary string for data-page filter bars.
func FilterBarPageSummary(page, total int, search string) string {
	return filterbar.PageSummary(page, total, search)
}

// FilterBarCompactPageSummary creates a standard page/total/search summary with
// a display-width-bounded search segment.
func FilterBarCompactPageSummary(page, total int, search string, searchMaxWidth int) string {
	return filterbar.CompactPageSummary(page, total, search, searchMaxWidth)
}

// FilterBarLookupSummary creates a standard lookup/source/resolved/counts
// summary for diagnostic filter bars.
func FilterBarLookupSummary(config filterbar.LookupSummaryConfig) string {
	return filterbar.LookupSummary(config)
}

// FilterBarSearchRefresh creates a standard data-page filter bar with one search field and Refresh.
func FilterBarSearchRefresh(key, summary string, width, labelWidth int, search filterbar.Field, refreshIntent intent.Intent, busy bool, loadingReason string) rtui.VNode {
	return filterbar.SearchRefresh(key, summary, width, labelWidth, search, refreshIntent, busy, loadingReason)
}

// FilterBarSearchRefreshClear creates a standard one-search filter bar with Refresh and Clear.
func FilterBarSearchRefreshClear(key, summary string, width, labelWidth int, search filterbar.Field, refreshIntent intent.Intent, busy bool, loadingReason string) rtui.VNode {
	return filterbar.SearchRefreshClear(key, summary, width, labelWidth, search, refreshIntent, busy, loadingReason)
}

// FilterBarSearchActions creates a standard one-search filter bar with contextual actions.
func FilterBarSearchActions(key, summary string, width, labelWidth int, search filterbar.Field, actions ...filterbar.Action) rtui.VNode {
	return filterbar.SearchActions(key, summary, width, labelWidth, search, actions...)
}

// FilterBarFieldsRefresh creates a standard wrapped multi-field filter bar with Refresh.
func FilterBarFieldsRefresh(key, summary string, width, labelWidth int, fields []filterbar.Field, refreshIntent intent.Intent, busy bool, loadingReason string) rtui.VNode {
	return filterbar.FieldsRefresh(key, summary, width, labelWidth, fields, refreshIntent, busy, loadingReason)
}

// FilterBarFieldsRefreshReset creates a standard wrapped multi-field filter bar with Refresh and Reset.
func FilterBarFieldsRefreshReset(key, summary string, width, labelWidth int, fields []filterbar.Field, refreshIntent, resetIntent intent.Intent, busy bool, loadingReason string) rtui.VNode {
	return filterbar.FieldsRefreshReset(key, summary, width, labelWidth, fields, refreshIntent, resetIntent, busy, loadingReason)
}

// FilterBarFieldsRefreshResetWhenChanged creates a wrapped multi-field filter bar with
// Refresh and a Reset action disabled until filters differ from defaults.
func FilterBarFieldsRefreshResetWhenChanged(key, summary string, width, labelWidth int, fields []filterbar.Field, refreshIntent, resetIntent intent.Intent, resetChanged bool, busy bool, loadingReason string) rtui.VNode {
	return filterbar.FieldsRefreshResetWhenChanged(key, summary, width, labelWidth, fields, refreshIntent, resetIntent, resetChanged, busy, loadingReason)
}
