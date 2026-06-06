package text

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/wwsheng009/mint/runtime/paint"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

const defaultSensitiveMask = "****"

// CountPart describes one labeled non-negative count in an operational summary.
type CountPart struct {
	Label string
	Count int
}

// KeyValuePart describes one compact key/value display segment.
type KeyValuePart struct {
	Label string
	Value string
}

// Muted creates a muted helper text node.
func Muted(content string) rtui.VNode {
	return NewBuilder(content).FgColor("gray").Build()
}

// MutedLines creates a vertical stack of muted helper text lines.
func MutedLines(lines ...string) rtui.VNode {
	nodes := make([]rtui.VNode, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		nodes = append(nodes, Muted(line))
	}
	if len(nodes) == 0 {
		return nil
	}
	return rtui.VStackBuilder(nodes...).Gap(0).AlignCross(rtui.AlignStart).Build()
}

// Subtle creates a lower-contrast helper text node.
func Subtle(content string) rtui.VNode {
	return NewBuilder(content).FgColor("bright-black").Build()
}

// Success creates a success helper text node.
func Success(content string) rtui.VNode {
	return NewBuilder(content).FgColor("green").Build()
}

// Warning creates a warning helper text node.
func Warning(content string) rtui.VNode {
	return NewBuilder(content).FgColor("yellow").Build()
}

// Danger creates an error helper text node.
func Danger(content string) rtui.VNode {
	return NewBuilder(content).FgColor("red").Build()
}

// State creates a helper text node using common operational state colors.
func State(content, state string) rtui.VNode {
	switch normalizeOperationalState(state) {
	case "healthy", "active", "available", "effective", "enabled", "success", "ok", "ready", "complete", "completed", "in_sync":
		return Success(content)
	case "processing", "loading", "syncing", "refreshing", "reloading", "running", "busy", "executing", "degraded", "rate_limited", "limited", "pending_restart", "pending", "warning", "cooldown", "partial", "lagging", "retrying", "queued", "attention", "risk":
		return Warning(content)
	case "unhealthy", "disabled", "unauthorized", "unavailable", "failed", "failure", "error", "down", "blocked", "out_of_sync", "exception", "critical", "firing":
		return Danger(content)
	default:
		return Muted(content)
	}
}

// DisplayText normalizes a display string and replaces blank content with "-".
func DisplayText(content string) string {
	return FallbackText(content, "-")
}

// FallbackText normalizes a display string and uses fallback for blank content.
func FallbackText(content, fallback string) string {
	fallback = normalizeDisplayText(fallback)
	if fallback == "" {
		fallback = "-"
	}
	content = normalizeDisplayText(content)
	if content == "" {
		return fallback
	}
	return content
}

// FirstNonEmptyText returns the first normalized non-empty display string.
func FirstNonEmptyText(values ...string) string {
	for _, value := range values {
		if text := normalizeDisplayText(value); text != "" {
			return text
		}
	}
	return ""
}

// CompactText returns a display-width-bounded string with "-" for blank content.
func CompactText(content string, maxWidth int) string {
	return CompactFallbackText(content, "-", maxWidth)
}

// CompactFallbackText returns a fallback-backed display-width-bounded string.
func CompactFallbackText(content, fallback string, maxWidth int) string {
	return compactDisplayText(FallbackText(content, fallback), maxWidth)
}

// OptionalCompactText returns a display-width-bounded string and preserves blank content as empty.
func OptionalCompactText(content string, maxWidth int) string {
	content = normalizeDisplayText(content)
	if content == "" {
		return ""
	}
	return compactDisplayText(content, maxWidth)
}

// IntText formats a non-negative integer for operational display.
func IntText(value int) string {
	if value < 0 {
		value = 0
	}
	return fmt.Sprintf("%d", value)
}

// NonZeroIntText formats a positive integer and leaves zero or negative values blank.
func NonZeroIntText(value int) string {
	if value <= 0 {
		return ""
	}
	return IntText(value)
}

// RatioText formats an available/total pair for operational display.
func RatioText(available, total int) string {
	if available < 0 {
		available = 0
	}
	if total < 0 {
		total = 0
	}
	return fmt.Sprintf("%d/%d", available, total)
}

// SlashText joins already-formatted operational display segments with "/".
func SlashText(values ...string) string {
	if len(values) == 0 {
		return "-"
	}
	segments := make([]string, 0, len(values))
	for _, value := range values {
		segments = append(segments, FallbackText(value, "-"))
	}
	return strings.Join(segments, "/")
}

// ListText joins display list values with ", ", skipping blank values.
func ListText(values ...string) string {
	segments := make([]string, 0, len(values))
	for _, value := range values {
		if text := FirstNonEmptyText(value); text != "" {
			segments = append(segments, text)
		}
	}
	if len(segments) == 0 {
		return "-"
	}
	return strings.Join(segments, ", ")
}

// BoolText formats a boolean as a compact yes/no operational value.
func BoolText(value bool) string {
	if value {
		return "yes"
	}
	return "no"
}

// EnabledText formats a boolean as an enabled/disabled operational state.
func EnabledText(value bool) string {
	if value {
		return "enabled"
	}
	return "disabled"
}

// ShortTimeText formats a time-of-day value for compact operational display.
func ShortTimeText(value time.Time) string {
	if value.IsZero() {
		return "-"
	}
	return value.Format("15:04:05")
}

// SortScopeText formats a sort scope such as "server runtime desc" or
// "current page latency asc". It returns blank when label is blank so callers
// can omit inactive sort summaries without rendering a misleading scope.
func SortScopeText(scope, label string, descending bool) string {
	label = normalizeDisplayText(label)
	if label == "" {
		return ""
	}
	scope = normalizeDisplayText(scope)
	direction := "asc"
	if descending {
		direction = "desc"
	}
	if scope == "" {
		return label + " " + direction
	}
	return scope + " " + label + " " + direction
}

// ServerSortScopeText formats a sort scope known to be applied by the server.
func ServerSortScopeText(label string, descending bool) string {
	return SortScopeText("server", label, descending)
}

// CurrentPageSortScopeText formats a sort scope that only applies to the
// current loaded page or rows.
func CurrentPageSortScopeText(label string, descending bool) string {
	return SortScopeText("current page", label, descending)
}

// SortLabelForColumn returns the normalized display label for a sortable
// column index. Negative, out-of-range, or blank labels return blank so callers
// can omit inactive or unsupported sort summaries.
func SortLabelForColumn(columnIndex int, labels ...string) string {
	if columnIndex < 0 || columnIndex >= len(labels) {
		return ""
	}
	return normalizeDisplayText(labels[columnIndex])
}

// ColumnSortScopeText formats a sort scope from a column index and ordered
// labels. It returns blank when the column index is not mapped.
func ColumnSortScopeText(scope string, columnIndex int, descending bool, labels ...string) string {
	return SortScopeText(scope, SortLabelForColumn(columnIndex, labels...), descending)
}

// ServerColumnSortScopeText formats a column sort scope applied by the server.
func ServerColumnSortScopeText(columnIndex int, descending bool, labels ...string) string {
	return ColumnSortScopeText("server", columnIndex, descending, labels...)
}

// CurrentPageColumnSortScopeText formats a column sort scope that only applies
// to the current loaded page or rows.
func CurrentPageColumnSortScopeText(columnIndex int, descending bool, labels ...string) string {
	return ColumnSortScopeText("current page", columnIndex, descending, labels...)
}

// PercentText formats a ratio or percentage value with one decimal place.
func PercentText(value float64) string {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return "-"
	}
	if value <= 1 {
		value *= 100
	}
	return fmt.Sprintf("%.1f%%", value)
}

// MillisecondsText formats a positive millisecond value.
func MillisecondsText(value float64) string {
	if value <= 0 || math.IsNaN(value) || math.IsInf(value, 0) {
		return "-"
	}
	return fmt.Sprintf("%.0fms", value)
}

// DecimalText formats a finite decimal value with one decimal place.
func DecimalText(value float64) string {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return "-"
	}
	return fmt.Sprintf("%.1f", value)
}

// TrimmedDecimalText formats a finite decimal value up to maxDecimals places and removes trailing zeroes.
func TrimmedDecimalText(value float64, maxDecimals int) string {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return "-"
	}
	if maxDecimals < 0 {
		maxDecimals = 0
	}
	if maxDecimals > 9 {
		maxDecimals = 9
	}
	if maxDecimals > 0 {
		scale := math.Pow10(maxDecimals)
		if math.Round(value*scale)/scale == 0 {
			value = 0
		}
	}
	formatted := fmt.Sprintf("%."+strconv.Itoa(maxDecimals)+"f", value)
	if maxDecimals == 0 {
		return formatted
	}
	formatted = strings.TrimRight(strings.TrimRight(formatted, "0"), ".")
	if formatted == "-0" {
		return "0"
	}
	return formatted
}

// OptionalTrimmedDecimalText formats a value only when it is present.
func OptionalTrimmedDecimalText(value float64, present bool, maxDecimals int) string {
	if !present {
		return "-"
	}
	return TrimmedDecimalText(value, maxDecimals)
}

// UnitText formats a finite decimal value with a compact operational unit.
func UnitText(value float64, unit string) string {
	unit = normalizeDisplayText(unit)
	switch unit {
	case "%":
		return PercentText(value)
	case "ms":
		return MillisecondsText(value)
	case "":
		return DecimalText(value)
	default:
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return "-"
		}
		return fmt.Sprintf("%.1f%s", value, unit)
	}
}

// OptionalUnitText formats a unit value only when it is present.
func OptionalUnitText(value float64, present bool, unit string) string {
	if !present {
		return "-"
	}
	return UnitText(value, unit)
}

// SecondsText formats a non-negative seconds count.
func SecondsText(seconds int) string {
	if seconds < 0 {
		seconds = 0
	}
	return fmt.Sprintf("%ds", seconds)
}

// DurationSecondsText formats a positive duration in seconds, or "-" when absent.
func DurationSecondsText(seconds int) string {
	if seconds <= 0 {
		return "-"
	}
	return fmt.Sprintf("%ds", seconds)
}

// CountSummaryText joins labeled non-negative counts as "1 active / 2 idle".
func CountSummaryText(parts ...CountPart) string {
	segments := make([]string, 0, len(parts))
	for _, part := range parts {
		label := normalizeDisplayText(part.Label)
		if label == "" {
			continue
		}
		segments = append(segments, IntText(part.Count)+" "+label)
	}
	if len(segments) == 0 {
		return "-"
	}
	return strings.Join(segments, " / ")
}

// KeyValueSummaryText joins compact key/value display segments as "key=value / other=-".
func KeyValueSummaryText(parts ...KeyValuePart) string {
	segments := make([]string, 0, len(parts))
	for _, part := range parts {
		label := normalizeDisplayText(part.Label)
		if label == "" {
			continue
		}
		segments = append(segments, label+"="+FallbackText(part.Value, "-"))
	}
	if len(segments) == 0 {
		return "-"
	}
	return strings.Join(segments, " / ")
}

// MaskSensitive returns a partially masked display value for secrets,
// account identifiers, tokens, and provider keys.
func MaskSensitive(content string, visiblePrefix, visibleSuffix int) string {
	content = normalizeDisplayText(content)
	if content == "" {
		return "-"
	}
	if visiblePrefix < 0 {
		visiblePrefix = 0
	}
	if visibleSuffix < 0 {
		visibleSuffix = 0
	}
	if visiblePrefix == 0 && visibleSuffix == 0 {
		return defaultSensitiveMask
	}
	runes := []rune(content)
	if len(runes) <= visiblePrefix+visibleSuffix {
		return defaultSensitiveMask
	}
	return string(runes[:visiblePrefix]) + "..." + string(runes[len(runes)-visibleSuffix:])
}

func normalizeOperationalState(state string) string {
	normalized := strings.ToLower(strings.TrimSpace(strings.NewReplacer(
		"\r\n", " ",
		"\n", " ",
		"\r", " ",
		"\t", " ",
	).Replace(state)))
	return strings.ReplaceAll(normalized, " ", "_")
}

func normalizeDisplayText(content string) string {
	return strings.TrimSpace(strings.NewReplacer(
		"\r\n", " ",
		"\n", " ",
		"\r", " ",
		"\t", " ",
	).Replace(content))
}

func compactDisplayText(content string, maxWidth int) string {
	if maxWidth <= 0 || paint.StringWidth(content) <= maxWidth {
		return content
	}
	if maxWidth <= 3 {
		return trimDisplayText(content, maxWidth)
	}
	prefix := strings.TrimRight(trimDisplayText(content, maxWidth-3), " ")
	if prefix == "" {
		return "..."
	}
	return prefix + "..."
}

func trimDisplayText(content string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	var builder strings.Builder
	width := 0
	for _, r := range content {
		runeWidth := paint.RuneWidth(r)
		if runeWidth <= 0 {
			continue
		}
		if width+runeWidth > maxWidth {
			break
		}
		builder.WriteRune(r)
		width += runeWidth
	}
	return builder.String()
}
