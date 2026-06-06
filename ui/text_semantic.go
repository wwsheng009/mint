package ui

import (
	"time"

	rtui "github.com/wwsheng009/mint/runtime/ui"
	textcomp "github.com/wwsheng009/mint/ui/components/text"
)

// CountTextPart describes one labeled non-negative count in an operational summary.
type CountTextPart = textcomp.CountPart

// KeyValueTextPart describes one compact key/value display segment.
type KeyValueTextPart = textcomp.KeyValuePart

// TextMuted creates a muted helper text node.
func TextMuted(content string) rtui.VNode {
	return textcomp.Muted(content)
}

// TextMutedLines creates a vertical stack of muted helper text lines.
func TextMutedLines(lines ...string) rtui.VNode {
	return textcomp.MutedLines(lines...)
}

// TextSubtle creates a lower-contrast helper text node.
func TextSubtle(content string) rtui.VNode {
	return textcomp.Subtle(content)
}

// TextSuccess creates a success helper text node.
func TextSuccess(content string) rtui.VNode {
	return textcomp.Success(content)
}

// TextWarning creates a warning helper text node.
func TextWarning(content string) rtui.VNode {
	return textcomp.Warning(content)
}

// TextDanger creates an error helper text node.
func TextDanger(content string) rtui.VNode {
	return textcomp.Danger(content)
}

// TextState creates a helper text node using common operational state colors.
func TextState(content, state string) rtui.VNode {
	return textcomp.State(content, state)
}

// DisplayText normalizes a display string and replaces blank content with "-".
func DisplayText(content string) string {
	return textcomp.DisplayText(content)
}

// DisplayFallbackText normalizes a display string and uses fallback for blank content.
func DisplayFallbackText(content, fallback string) string {
	return textcomp.FallbackText(content, fallback)
}

// FirstNonEmptyText returns the first normalized non-empty display string.
func FirstNonEmptyText(values ...string) string {
	return textcomp.FirstNonEmptyText(values...)
}

// CompactText returns a display-width-bounded string with "-" for blank content.
func CompactText(content string, maxWidth int) string {
	return textcomp.CompactText(content, maxWidth)
}

// CompactFallbackText returns a fallback-backed display-width-bounded string.
func CompactFallbackText(content, fallback string, maxWidth int) string {
	return textcomp.CompactFallbackText(content, fallback, maxWidth)
}

// OptionalCompactText returns a display-width-bounded string and preserves blank content as empty.
func OptionalCompactText(content string, maxWidth int) string {
	return textcomp.OptionalCompactText(content, maxWidth)
}

// IntDisplayText formats a non-negative integer for operational display.
func IntDisplayText(value int) string {
	return textcomp.IntText(value)
}

// NonZeroIntText formats a positive integer and leaves zero or negative values blank.
func NonZeroIntText(value int) string {
	return textcomp.NonZeroIntText(value)
}

// RatioText formats an available/total pair for operational display.
func RatioText(available, total int) string {
	return textcomp.RatioText(available, total)
}

// SlashText joins already-formatted operational display segments with "/".
func SlashText(values ...string) string {
	return textcomp.SlashText(values...)
}

// ListText joins display list values with ", ", skipping blank values.
func ListText(values ...string) string {
	return textcomp.ListText(values...)
}

// BoolText formats a boolean as a compact yes/no operational value.
func BoolText(value bool) string {
	return textcomp.BoolText(value)
}

// EnabledText formats a boolean as an enabled/disabled operational state.
func EnabledText(value bool) string {
	return textcomp.EnabledText(value)
}

// ShortTimeText formats a time-of-day value for compact operational display.
func ShortTimeText(value time.Time) string {
	return textcomp.ShortTimeText(value)
}

// SortScopeText formats a sort scope such as "server runtime desc" or
// "current page latency asc".
func SortScopeText(scope, label string, descending bool) string {
	return textcomp.SortScopeText(scope, label, descending)
}

// ServerSortScopeText formats a sort scope known to be applied by the server.
func ServerSortScopeText(label string, descending bool) string {
	return textcomp.ServerSortScopeText(label, descending)
}

// CurrentPageSortScopeText formats a sort scope that only applies to the
// current loaded page or rows.
func CurrentPageSortScopeText(label string, descending bool) string {
	return textcomp.CurrentPageSortScopeText(label, descending)
}

// SortLabelForColumn returns the normalized display label for a sortable column index.
func SortLabelForColumn(columnIndex int, labels ...string) string {
	return textcomp.SortLabelForColumn(columnIndex, labels...)
}

// ColumnSortScopeText formats a sort scope from a column index and ordered labels.
func ColumnSortScopeText(scope string, columnIndex int, descending bool, labels ...string) string {
	return textcomp.ColumnSortScopeText(scope, columnIndex, descending, labels...)
}

// ServerColumnSortScopeText formats a column sort scope known to be applied by the server.
func ServerColumnSortScopeText(columnIndex int, descending bool, labels ...string) string {
	return textcomp.ServerColumnSortScopeText(columnIndex, descending, labels...)
}

// CurrentPageColumnSortScopeText formats a column sort scope that only applies to the current loaded page or rows.
func CurrentPageColumnSortScopeText(columnIndex int, descending bool, labels ...string) string {
	return textcomp.CurrentPageColumnSortScopeText(columnIndex, descending, labels...)
}

// PercentText formats a ratio or percentage value with one decimal place.
func PercentText(value float64) string {
	return textcomp.PercentText(value)
}

// MillisecondsText formats a positive millisecond value.
func MillisecondsText(value float64) string {
	return textcomp.MillisecondsText(value)
}

// DecimalText formats a finite decimal value with one decimal place.
func DecimalText(value float64) string {
	return textcomp.DecimalText(value)
}

// TrimmedDecimalText formats a finite decimal value up to maxDecimals places and removes trailing zeroes.
func TrimmedDecimalText(value float64, maxDecimals int) string {
	return textcomp.TrimmedDecimalText(value, maxDecimals)
}

// OptionalTrimmedDecimalText formats a value only when it is present.
func OptionalTrimmedDecimalText(value float64, present bool, maxDecimals int) string {
	return textcomp.OptionalTrimmedDecimalText(value, present, maxDecimals)
}

// UnitText formats a finite decimal value with a compact operational unit.
func UnitText(value float64, unit string) string {
	return textcomp.UnitText(value, unit)
}

// OptionalUnitText formats a unit value only when it is present.
func OptionalUnitText(value float64, present bool, unit string) string {
	return textcomp.OptionalUnitText(value, present, unit)
}

// SecondsText formats a non-negative seconds count.
func SecondsText(seconds int) string {
	return textcomp.SecondsText(seconds)
}

// DurationSecondsText formats a positive duration in seconds, or "-" when absent.
func DurationSecondsText(seconds int) string {
	return textcomp.DurationSecondsText(seconds)
}

// CountSummaryText joins labeled non-negative counts as "1 active / 2 idle".
func CountSummaryText(parts ...CountTextPart) string {
	return textcomp.CountSummaryText(parts...)
}

// KeyValueSummaryText joins compact key/value display segments as "key=value / other=-".
func KeyValueSummaryText(parts ...KeyValueTextPart) string {
	return textcomp.KeyValueSummaryText(parts...)
}

// MaskSensitiveText returns a partially masked display value for sensitive text.
func MaskSensitiveText(content string, visiblePrefix, visibleSuffix int) string {
	return textcomp.MaskSensitive(content, visiblePrefix, visibleSuffix)
}
