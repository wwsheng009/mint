package filterbar

import (
	"fmt"
	"strings"

	"github.com/wwsheng009/mint/runtime/paint"
)

const summarySeparator = " · "

// SummaryPart describes one "label value" segment in a filter bar summary.
type SummaryPart struct {
	Label string
	Value string
}

// LookupSummaryConfig describes a diagnostic lookup summary where the target
// may come from typed input, current selection, or another page context.
type LookupSummaryConfig struct {
	LookupLabel      string
	Lookup           string
	LookupFallback   string
	LookupWidth      int
	SourceLabel      string
	Source           string
	SourceFallback   string
	ResolvedLabel    string
	Resolved         string
	ResolvedFallback string
	ResolvedWidth    int
	ItemsLabel       string
	Items            int
	ErrorsLabel      string
	Errors           int
}

// Summary joins operational filter summary segments with a stable separator.
func Summary(parts ...SummaryPart) string {
	if len(parts) == 0 {
		return ""
	}
	segments := make([]string, 0, len(parts))
	for _, part := range parts {
		label := normalizeInlineText(part.Label)
		value := normalizeInlineText(part.Value)
		if label == "" && value == "" {
			continue
		}
		if value == "" {
			value = "-"
		}
		if label == "" {
			segments = append(segments, value)
			continue
		}
		segments = append(segments, label+" "+value)
	}
	return strings.Join(segments, summarySeparator)
}

// SummaryValue creates a summary segment and normalizes blank values to "-".
func SummaryValue(label, value string) SummaryPart {
	return SummaryPart{Label: label, Value: summaryValueText(value)}
}

// SummaryValueUnless creates a summary segment only when value is non-empty
// and differs from defaultValue after trimming.
func SummaryValueUnless(label, value, defaultValue string) SummaryPart {
	if summaryValueIsDefault(value, defaultValue) {
		return SummaryPart{}
	}
	return SummaryValue(label, value)
}

// SummaryPresence creates a summary segment that reports whether value is
// present without displaying the value itself. It is useful for required
// operation fields such as reason, confirmation text, or ticket id where the
// filter bar only needs to show readiness.
func SummaryPresence(label, value, presentText, missingText string) SummaryPart {
	presentText = normalizeInlineText(presentText)
	if presentText == "" {
		presentText = "ready"
	}
	missingText = normalizeInlineText(missingText)
	if missingText == "" {
		missingText = "missing"
	}
	if normalizeInlineText(value) == "" {
		return SummaryValue(label, missingText)
	}
	return SummaryValue(label, presentText)
}

// SummaryCount creates a non-negative integer count summary segment.
func SummaryCount(label string, count int) SummaryPart {
	if count < 0 {
		count = 0
	}
	return SummaryValue(label, fmt.Sprintf("%d", count))
}

// SummaryRatio creates an "available/total" summary segment.
func SummaryRatio(label string, available, total int) SummaryPart {
	if available < 0 {
		available = 0
	}
	if total < 0 {
		total = 0
	}
	return SummaryValue(label, fmt.Sprintf("%d/%d", available, total))
}

// SummarySearch creates a standard search summary segment.
func SummarySearch(search string) SummaryPart {
	return SummaryValue("search", search)
}

// SummaryCompactSearch creates a display-width-bounded search summary segment.
func SummaryCompactSearch(search string, maxWidth int) SummaryPart {
	return SummaryCompact("search", search, maxWidth)
}

// SummaryCompact creates a display-width-bounded summary segment.
func SummaryCompact(label, value string, maxWidth int) SummaryPart {
	return SummaryValue(label, compactSummaryText(summaryValueText(value), maxWidth))
}

// SummaryCompactUnless creates a display-width-bounded summary segment only
// when value is non-empty and differs from defaultValue after trimming.
func SummaryCompactUnless(label, value, defaultValue string, maxWidth int) SummaryPart {
	if summaryValueIsDefault(value, defaultValue) {
		return SummaryPart{}
	}
	return SummaryCompact(label, value, maxWidth)
}

// PageSummary creates a standard page/total/search summary.
func PageSummary(page, total int, search string) string {
	return CompactPageSummary(page, total, search, 0)
}

// CompactPageSummary creates a standard page/total/search summary with a
// display-width-bounded search segment.
func CompactPageSummary(page, total int, search string, searchMaxWidth int) string {
	if page <= 0 {
		page = 1
	}
	if total < 0 {
		total = 0
	}
	searchPart := SummarySearch(search)
	if searchMaxWidth > 0 {
		searchPart = SummaryCompactSearch(search, searchMaxWidth)
	}
	return Summary(
		SummaryCount("page", page),
		SummaryCount("total", total),
		searchPart,
	)
}

// LookupSummary creates a standard lookup/source/resolved/counts summary for
// diagnostic pages such as traces, audit events, or drilldown lookups.
func LookupSummary(cfg LookupSummaryConfig) string {
	lookupLabel := normalizeInlineText(cfg.LookupLabel)
	if lookupLabel == "" {
		lookupLabel = "lookup"
	}
	sourceLabel := normalizeInlineText(cfg.SourceLabel)
	if sourceLabel == "" {
		sourceLabel = "source"
	}
	resolvedLabel := normalizeInlineText(cfg.ResolvedLabel)
	if resolvedLabel == "" {
		resolvedLabel = "resolved"
	}
	itemsLabel := normalizeInlineText(cfg.ItemsLabel)
	if itemsLabel == "" {
		itemsLabel = "items"
	}
	errorsLabel := normalizeInlineText(cfg.ErrorsLabel)
	if errorsLabel == "" {
		errorsLabel = "errors"
	}

	lookup := SummaryValue(lookupLabel, lookupSummaryValue(cfg.Lookup, cfg.LookupFallback))
	if cfg.LookupWidth > 0 {
		lookup = SummaryCompact(lookupLabel, lookupSummaryValue(cfg.Lookup, cfg.LookupFallback), cfg.LookupWidth)
	}
	resolved := SummaryValue(resolvedLabel, lookupSummaryValue(cfg.Resolved, cfg.ResolvedFallback))
	if cfg.ResolvedWidth > 0 {
		resolved = SummaryCompact(resolvedLabel, lookupSummaryValue(cfg.Resolved, cfg.ResolvedFallback), cfg.ResolvedWidth)
	}

	return Summary(
		lookup,
		SummaryValue(sourceLabel, lookupSummaryValue(cfg.Source, cfg.SourceFallback)),
		resolved,
		SummaryCount(itemsLabel, cfg.Items),
		SummaryCount(errorsLabel, cfg.Errors),
	)
}

func lookupSummaryValue(value, fallback string) string {
	value = normalizeInlineText(value)
	if value != "" {
		return value
	}
	fallback = normalizeInlineText(fallback)
	if fallback != "" {
		return fallback
	}
	return ""
}

func summaryValueText(value string) string {
	value = normalizeInlineText(value)
	if value == "" {
		return "-"
	}
	return value
}

func summaryValueIsDefault(value, defaultValue string) bool {
	value = normalizeInlineText(value)
	defaultValue = normalizeInlineText(defaultValue)
	return value == "" || value == defaultValue
}

func compactSummaryText(text string, maxWidth int) string {
	if maxWidth <= 0 || paint.StringWidth(text) <= maxWidth {
		return text
	}
	if maxWidth <= 3 {
		return trimSummaryText(text, maxWidth)
	}
	prefix := strings.TrimRight(trimSummaryText(text, maxWidth-3), " ")
	if prefix == "" {
		return "..."
	}
	return prefix + "..."
}

func trimSummaryText(text string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	var builder strings.Builder
	width := 0
	for _, r := range text {
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
