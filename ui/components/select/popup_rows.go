package selectcomp

import (
	"fmt"
	"strings"

	"github.com/wwsheng009/mint/runtime/paint"
)

const highlightCreateTag = -2

type OptionGroup struct {
	Label   string
	Options []Option
}

type popupRowKind int

const (
	popupRowCreateTag popupRowKind = iota
	popupRowGroup
	popupRowOption
	popupRowEmpty
)

type popupRow struct {
	kind        popupRowKind
	optionIndex int
	text        string
}

type popupRows struct {
	showFilter        bool
	filterQuery       string
	filterPlaceholder string
	scrollable        []popupRow
}

func flattenOptionGroups(groups []OptionGroup) []Option {
	if len(groups) == 0 {
		return nil
	}

	result := make([]Option, 0, len(groups)*2)
	for _, group := range groups {
		for _, opt := range group.Options {
			next := opt
			next.Group = group.Label
			result = append(result, next)
		}
	}
	return result
}

func mergeOptions(base, created []Option) []Option {
	if len(base) == 0 && len(created) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(base)+len(created))
	result := make([]Option, 0, len(base)+len(created))
	appendUnique := func(options []Option) {
		for _, opt := range options {
			key := strings.ToLower(opt.Value)
			if _, exists := seen[key]; exists {
				continue
			}
			seen[key] = struct{}{}
			result = append(result, opt)
		}
	}

	appendUnique(base)
	appendUnique(created)
	return result
}

func buildPopupRows(
	options []Option,
	mode SelectionMode,
	filterEnabled bool,
	filterPlaceholder string,
	filterQuery string,
) popupRows {
	rows := popupRows{
		showFilter:        filterEnabledFor(mode, filterEnabled),
		filterQuery:       sanitizeFilterText(filterQuery),
		filterPlaceholder: defaultFilterPlaceholder(filterPlaceholder),
	}

	if isTagsSelectionMode(mode) && canCreateTag(rows.filterQuery, options) {
		rows.scrollable = append(rows.scrollable, popupRow{
			kind: popupRowCreateTag,
			text: fmt.Sprintf("+ Create %q", strings.TrimSpace(rows.filterQuery)),
		})
	}

	filteredCount := 0
	lastGroup := ""
	haveGroup := false
	for idx, opt := range options {
		if !matchesFilter(opt, rows.filterQuery) {
			continue
		}

		if opt.Group != "" {
			if !haveGroup || lastGroup != opt.Group {
				rows.scrollable = append(rows.scrollable, popupRow{
					kind: popupRowGroup,
					text: "[" + opt.Group + "]",
				})
				lastGroup = opt.Group
				haveGroup = true
			}
		} else {
			lastGroup = ""
			haveGroup = false
		}

		rows.scrollable = append(rows.scrollable, popupRow{
			kind:        popupRowOption,
			optionIndex: idx,
			text:        opt.Label,
		})
		filteredCount++
	}

	if filteredCount == 0 && !hasCreateTagRow(rows) {
		emptyText := "(no options)"
		if strings.TrimSpace(rows.filterQuery) != "" {
			emptyText = "(no matches)"
		}
		if isTagsSelectionMode(mode) && strings.TrimSpace(rows.filterQuery) == "" {
			emptyText = "(type to add tags)"
		}
		rows.scrollable = append(rows.scrollable, popupRow{
			kind: popupRowEmpty,
			text: emptyText,
		})
	}

	return rows
}

func hasCreateTagRow(rows popupRows) bool {
	for _, row := range rows.scrollable {
		if row.kind == popupRowCreateTag {
			return true
		}
	}
	return false
}

func filterEnabledFor(mode SelectionMode, filterEnabled bool) bool {
	return filterEnabled || isTagsSelectionMode(mode)
}

func defaultFilterPlaceholder(placeholder string) string {
	if strings.TrimSpace(placeholder) == "" {
		return "type to filter"
	}
	return placeholder
}

func sanitizeFilterText(text string) string {
	text = strings.ReplaceAll(text, "\r", " ")
	text = strings.ReplaceAll(text, "\n", " ")
	return text
}

func matchesFilter(opt Option, query string) bool {
	query = strings.TrimSpace(strings.ToLower(query))
	if query == "" {
		return true
	}

	haystack := strings.ToLower(strings.Join([]string{
		opt.Label,
		opt.Value,
		opt.Group,
	}, "\n"))
	return strings.Contains(haystack, query)
}

func findExactOptionIndex(options []Option, query string) int {
	query = strings.TrimSpace(query)
	if query == "" {
		return -1
	}

	for idx, opt := range options {
		if strings.EqualFold(opt.Value, query) || strings.EqualFold(opt.Label, query) {
			return idx
		}
	}
	return -1
}

func canCreateTag(query string, options []Option) bool {
	query = strings.TrimSpace(query)
	return query != "" && findExactOptionIndex(options, query) < 0
}

func createTagOption(query string) Option {
	query = strings.TrimSpace(query)
	return Option{
		Value: query,
		Label: query,
	}
}

func selectableTargets(rows popupRows) []int {
	targets := make([]int, 0, len(rows.scrollable))
	for _, row := range rows.scrollable {
		switch row.kind {
		case popupRowCreateTag:
			targets = append(targets, highlightCreateTag)
		case popupRowOption:
			targets = append(targets, row.optionIndex)
		}
	}
	return targets
}

func firstSelectableTarget(rows popupRows) int {
	targets := selectableTargets(rows)
	if len(targets) == 0 {
		return -1
	}
	return targets[0]
}

func lastSelectableTarget(rows popupRows) int {
	targets := selectableTargets(rows)
	if len(targets) == 0 {
		return -1
	}
	return targets[len(targets)-1]
}

func defaultHighlightTarget(
	rows popupRows,
	mode SelectionMode,
	selectedIndex int,
	selectedIndices []int,
) int {
	if isTagsSelectionMode(mode) && hasCreateTagRow(rows) {
		return highlightCreateTag
	}

	if isMultiSelectionMode(mode) && len(selectedIndices) > 0 {
		candidate := selectedIndices[len(selectedIndices)-1]
		if rowPositionForHighlight(rows, candidate) >= 0 {
			return candidate
		}
	}

	if rowPositionForHighlight(rows, selectedIndex) >= 0 {
		return selectedIndex
	}
	return firstSelectableTarget(rows)
}

func rowPositionForHighlight(rows popupRows, highlight int) int {
	for idx, row := range rows.scrollable {
		switch row.kind {
		case popupRowCreateTag:
			if highlight == highlightCreateTag {
				return idx
			}
		case popupRowOption:
			if highlight == row.optionIndex {
				return idx
			}
		}
	}
	return -1
}

func visibleScrollableRowCount(rows popupRows, maxVisibleRows int) int {
	if len(rows.scrollable) == 0 {
		return 0
	}
	if maxVisibleRows <= 0 {
		maxVisibleRows = defaultMaxVisibleRows
	}
	return minInt(len(rows.scrollable), maxVisibleRows)
}

func maxScrollOffsetForRows(rows popupRows, maxVisibleRows int) int {
	return maxInt(0, len(rows.scrollable)-visibleScrollableRowCount(rows, maxVisibleRows))
}

func normalizePopupHighlight(
	rows popupRows,
	highlightedIndex int,
	scrollOffset int,
	maxVisibleRows int,
	mode SelectionMode,
	selectedIndex int,
	selectedIndices []int,
) (int, int) {
	if len(rows.scrollable) == 0 {
		return -1, 0
	}

	if rowPositionForHighlight(rows, highlightedIndex) < 0 {
		highlightedIndex = defaultHighlightTarget(rows, mode, selectedIndex, selectedIndices)
	}

	visible := visibleScrollableRowCount(rows, maxVisibleRows)
	if visible <= 0 {
		return highlightedIndex, 0
	}

	position := rowPositionForHighlight(rows, highlightedIndex)
	if position < 0 {
		return -1, 0
	}

	maxOffset := maxScrollOffsetForRows(rows, maxVisibleRows)
	scrollOffset = clampInt(scrollOffset, 0, maxOffset)
	if position < scrollOffset {
		scrollOffset = position
	}
	if position >= scrollOffset+visible {
		scrollOffset = position - visible + 1
	}
	return highlightedIndex, clampInt(scrollOffset, 0, maxOffset)
}

func nextHighlightTarget(rows popupRows, current, delta int) int {
	targets := selectableTargets(rows)
	if len(targets) == 0 {
		return -1
	}

	pos := indexOfInt(targets, current)
	if pos < 0 {
		if delta >= 0 {
			return targets[0]
		}
		return targets[len(targets)-1]
	}

	next := clampInt(pos+delta, 0, len(targets)-1)
	return targets[next]
}

func pageHighlightTarget(rows popupRows, current int, direction, pageSize int) int {
	if len(rows.scrollable) == 0 {
		return -1
	}
	if pageSize <= 0 {
		pageSize = 1
	}

	position := rowPositionForHighlight(rows, current)
	if position < 0 {
		return firstSelectableTarget(rows)
	}

	targetRow := clampInt(position+direction*pageSize, 0, len(rows.scrollable)-1)
	if direction >= 0 {
		for idx := targetRow; idx < len(rows.scrollable); idx++ {
			if target, ok := selectableTargetForRow(rows.scrollable[idx]); ok {
				return target
			}
		}
		return lastSelectableTarget(rows)
	}

	for idx := targetRow; idx >= 0; idx-- {
		if target, ok := selectableTargetForRow(rows.scrollable[idx]); ok {
			return target
		}
	}
	return firstSelectableTarget(rows)
}

func selectableTargetForRow(row popupRow) (int, bool) {
	switch row.kind {
	case popupRowCreateTag:
		return highlightCreateTag, true
	case popupRowOption:
		return row.optionIndex, true
	default:
		return 0, false
	}
}

func popupHitTarget(rows popupRows, scrollOffset, maxVisibleRows, localY int) (int, bool) {
	rowStart := 1
	if rows.showFilter {
		if localY == 1 {
			return 0, false
		}
		rowStart = 2
	}

	rowOffset := localY - rowStart
	if rowOffset < 0 || rowOffset >= visibleScrollableRowCount(rows, maxVisibleRows) {
		return 0, false
	}

	rowIndex := scrollOffset + rowOffset
	if rowIndex < 0 || rowIndex >= len(rows.scrollable) {
		return 0, false
	}
	return selectableTargetForRow(rows.scrollable[rowIndex])
}

func popupContentWidth(rows popupRows, mode SelectionMode) int {
	width := 0
	if rows.showFilter {
		width = maxInt(width, paint.StringWidth(popupFilterText(rows.filterQuery, rows.filterPlaceholder)))
	}

	markerWidth := markerWidthForMode(mode)
	for _, row := range rows.scrollable {
		switch row.kind {
		case popupRowOption:
			width = maxInt(width, markerWidth+1+paint.StringWidth(row.text))
		default:
			width = maxInt(width, paint.StringWidth(row.text))
		}
	}

	if width == 0 {
		return 4
	}
	return width
}

func popupFilterText(query, placeholder string) string {
	if strings.TrimSpace(query) != "" {
		return "/ " + query
	}
	return "/ " + defaultFilterPlaceholder(placeholder)
}

func popupRowText(
	row popupRow,
	width int,
	mode SelectionMode,
	selectedIndices []int,
) string {
	switch row.kind {
	case popupRowOption:
		marker := optionMarkerForMode(mode, selectedIndices, row.optionIndex)
		labelWidth := width
		content := ""
		if marker != "" {
			content = padDisplayWidth(marker, markerWidthForMode(mode))
			if width > markerWidthForMode(mode) {
				content += " "
				labelWidth = width - markerWidthForMode(mode) - 1
			} else {
				labelWidth = 0
			}
		}
		content += padDisplayWidth(truncateWithEllipsis(row.text, labelWidth), labelWidth)
		return padDisplayWidth(content, width)
	default:
		return padDisplayWidth(truncateWithEllipsis(row.text, width), width)
	}
}

func isHighlightedTarget(row popupRow, highlightedIndex int) bool {
	target, ok := selectableTargetForRow(row)
	return ok && target == highlightedIndex
}

func optionMarkerForMode(mode SelectionMode, selectedIndices []int, optionIndex int) string {
	selected := containsInt(selectedIndices, optionIndex)
	if isMultiSelectionMode(mode) {
		if selected {
			return "[x]"
		}
		return "[ ]"
	}
	if selected {
		return "●"
	}
	return "○"
}

func markerWidthForMode(mode SelectionMode) int {
	if isMultiSelectionMode(mode) {
		return 3
	}
	return 1
}

func joinSelectedValues(options []Option, selectedIndices []int) string {
	values := selectedValuesFor(options, selectedIndices)
	if len(values) == 0 {
		return ""
	}
	return strings.Join(values, ",")
}
