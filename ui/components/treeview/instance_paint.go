package treeview

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	scrollutil "github.com/wwsheng009/mint/ui/components/internal/scroll"
)

// =============================================================================
// PaintableInstance Interface
// =============================================================================

// Paint implements drawing logic for the tree view
func (inst *Instance) Paint(x, y int) []paint.DrawCmd {
	visible, _ := inst.visibleEntries()
	visibleCount := len(visible)
	viewSize := inst.effectiveViewportHeight(visibleCount)
	viewport := scrollutil.NewVerticalViewport(visibleCount, viewSize, inst.scrollOffset)
	scrollbarVisible := inst.showScrollbar && viewport.IsScrollable()
	startLine, endLine := viewport.VisibleRange()

	contentWidth := inst.calculateContentWidth(visible)
	statsLine := inst.searchStatsLine()
	if inst.showSearchStats {
		statsWidth := paint.StringWidth(statsLine)
		if statsWidth > contentWidth {
			contentWidth = statsWidth
		}
	}
	if contentWidth < 1 {
		contentWidth = 1
	}
	width := contentWidth
	if inst.showBorder {
		width += 4
	} else if scrollbarVisible {
		width += 1
	}

	cmds := make([]paint.DrawCmd, 0, viewSize+2)
	borderStyle := inst.treeStyle
	borderHorizontal := "─"
	borderVertical := "│"
	borderTopLeft := "┌"
	borderTopRight := "┐"
	borderBottomLeft := "└"
	borderBottomRight := "┘"
	focusLabel := ""
	if inst.focused {
		borderStyle = borderStyle.Bold(true)
		borderHorizontal = "═"
		borderVertical = "║"
		borderTopLeft = "╔"
		borderTopRight = "╗"
		borderBottomLeft = "╚"
		borderBottomRight = "╝"
		focusLabel = " [FOCUS] "
	}

	if inst.showBorder {
		topBorder := borderTopLeft + inst.borderInnerText(width-2, borderHorizontal, focusLabel) + borderTopRight
		cmds = append(cmds, paint.NewTextCmd(x, y, topBorder, borderStyle))
	}

	rowOffset := 0
	if inst.showBorder {
		rowOffset = 1
	}
	if inst.showSearchStats {
		statsText := padRightToWidth(truncateText(statsLine, contentWidth), contentWidth)
		statsStyle := inst.searchStatsStyle
		if statsStyle == (style.Style{}) {
			statsStyle = inst.treeStyle
		}
		if inst.showBorder {
			statsText = borderVertical + " " + statsText + " " + borderVertical
		}
		cmds = append(cmds, paint.NewTextCmd(x, y+rowOffset, statsText, statsStyle))
		rowOffset++
	}

	for i := startLine; i < endLine; i++ {
		entry := visible[i]
		rowY := y + rowOffset + (i - startLine)
		selected := i == inst.selectedIndex

		line, prefixWidth, icon, iconWidth, content := inst.composeLine(entry, contentWidth)
		rowStyle := inst.treeStyle
		if inst.rowStyleFn != nil {
			if override := inst.rowStyleFn(entry.Index, entry.Node); override != (style.Style{}) {
				rowStyle = override
			}
		}
		if entry.Match && inst.matchStyle != (style.Style{}) {
			rowStyle = inst.matchStyle
		}
		if selected {
			rowStyle = inst.selectedStyle
		}

		if inst.showBorder {
			line = borderVertical + " " + line + " " + borderVertical
		}
		cmds = append(cmds, paint.NewTextCmd(x, rowY, line, rowStyle))

		if entry.Match && inst.matchStyle != (style.Style{}) && inst.searchFn == nil {
			if highlightStart, highlightText, ok := inst.matchHighlight(content); ok {
				highlightX := x + prefixWidth + iconWidth + highlightStart
				if inst.showBorder {
					highlightX += 2
				}
				cmds = append(cmds, paint.NewTextCmd(highlightX, rowY, highlightText, inst.matchStyle))
			}
		}

		if !selected && inst.showIcons && iconWidth > 0 && inst.iconStyle != (style.Style{}) {
			iconX := x + prefixWidth
			if inst.showBorder {
				iconX += 2
			}
			// Only overlay icon if it fully fits in the content width.
			if contentWidth >= prefixWidth+iconWidth {
				cmds = append(cmds, paint.NewTextCmd(iconX, rowY, icon, inst.iconStyle))
			}
		}
	}

	// Fill remaining viewport rows (to clear previous content)
	for i := endLine; i < startLine+viewSize; i++ {
		rowY := y + rowOffset + (i - startLine)
		blank := strings.Repeat(" ", contentWidth)
		if inst.showBorder {
			blank = borderVertical + " " + blank + " " + borderVertical
		}
		cmds = append(cmds, paint.NewTextCmd(x, rowY, blank, inst.treeStyle))
	}

	if inst.showBorder {
		bottomBorder := borderBottomLeft + strings.Repeat(borderHorizontal, max(1, width-2)) + borderBottomRight
		cmds = append(cmds, paint.NewTextCmd(x, y+rowOffset+viewSize, bottomBorder, borderStyle))
	}

	if inst.showScrollbar {
		scrollbarStyle := inst.scrollbarStyle
		if scrollbarStyle == (style.Style{}) {
			scrollbarStyle = borderStyle
		} else if scrollbarStyle.FG == "" {
			scrollbarStyle = scrollbarStyle.Foreground(borderStyle.FG)
		}
		scrollbarX := x + max(1, width) - 1
		scrollbarY := y
		if inst.showBorder {
			scrollbarY++
		}
		scrollbarY += inst.statsHeight()
		cmds = append(cmds, scrollutil.DrawVerticalScrollbar(
			scrollbarX,
			scrollbarY,
			viewSize,
			viewport,
			scrollbarStyle,
			scrollutil.DefaultVerticalScrollbarConfig(),
		)...)
	}

	return cmds
}

func (inst *Instance) lineParts(entry nodeEntry) (prefix, icon, content string) {
	prefix = inst.indentPrefix(entry.Depth)
	prefix += inst.selectionMarker(entry)
	icon = inst.iconFor(entry)
	content = entry.Node.Content
	if inst.showLineNums {
		content += fmt.Sprintf(" [%d]", entry.Node.NodeID)
	}
	if suffix := inst.statusSuffix(entry); suffix != "" {
		content += suffix
	}
	return prefix, icon, content
}

func (inst *Instance) indentPrefix(depth int) string {
	if depth <= 0 {
		return ""
	}
	if inst.compact {
		return strings.Repeat("  ", depth)
	}
	return strings.Repeat("│   ", depth)
}

func (inst *Instance) iconFor(entry nodeEntry) string {
	if !inst.showIcons {
		return "  "
	}
	isFolder := entry.HasChildren || entry.Node.NodeType == "folder"
	if isFolder {
		if inst.isExpanded(entry.Index) {
			return "📂 "
		}
		return "📁 "
	}
	return "📄 "
}

func (inst *Instance) statusSuffix(entry nodeEntry) string {
	node := entry.Node
	if node.LoadError != "" {
		msg := trimToWidth(node.LoadError, 24)
		if msg != "" {
			return " [error: " + msg + "] [retry:R]"
		}
		return " [error] [retry:R]"
	}
	if node.Loading {
		return " [loading]"
	}
	if node.Lazy && !entry.HasDescendants {
		return " [load:R]"
	}
	return ""
}

func (inst *Instance) matchHighlight(content string) (int, string, bool) {
	query := strings.TrimSpace(inst.searchQuery)
	if query == "" || content == "" {
		return 0, "", false
	}

	searchable := content
	if strings.HasSuffix(searchable, "...") {
		searchable = strings.TrimSuffix(searchable, "...")
	}
	start, length, ok := matchSpanRunes(searchable, query)
	if !ok || length <= 0 {
		return 0, "", false
	}
	runes := []rune(searchable)
	if start < 0 || start+length > len(runes) {
		return 0, "", false
	}
	prefix := string(runes[:start])
	match := string(runes[start : start+length])
	return paint.StringWidth(prefix), match, true
}

func matchSpanRunes(content, query string) (int, int, bool) {
	contentRunes := []rune(content)
	queryRunes := []rune(query)
	if len(queryRunes) == 0 || len(contentRunes) == 0 || len(queryRunes) > len(contentRunes) {
		return 0, 0, false
	}
	for i := 0; i <= len(contentRunes)-len(queryRunes); i++ {
		match := true
		for j := range queryRunes {
			if unicode.ToLower(contentRunes[i+j]) != unicode.ToLower(queryRunes[j]) {
				match = false
				break
			}
		}
		if match {
			return i, len(queryRunes), true
		}
	}
	return 0, 0, false
}

func (inst *Instance) selectionMarker(entry nodeEntry) string {
	if inst.selectionMode == SelectionNone {
		return ""
	}
	if inst.isChecked(entry) {
		return "[x] "
	}
	return "[ ] "
}

func (inst *Instance) selectionMarkerWidth() int {
	if inst.selectionMode == SelectionNone {
		return 0
	}
	return paint.StringWidth("[ ] ")
}

func (inst *Instance) isChecked(entry nodeEntry) bool {
	if inst.selectionMode == SelectionNone {
		return false
	}
	return inst.checkedKeys != nil && inst.checkedKeys[entry.Key]
}

func (inst *Instance) composeLine(entry nodeEntry, maxWidth int) (line string, prefixWidth int, icon string, iconWidth int, content string) {
	prefix, icon, content := inst.lineParts(entry)
	prefixWidth = paint.StringWidth(prefix)
	iconWidth = paint.StringWidth(icon)

	available := maxWidth - prefixWidth - iconWidth
	if available < 0 {
		trimmed := trimToWidth(prefix+icon, maxWidth)
		return padRightToWidth(trimmed, maxWidth), prefixWidth, icon, iconWidth, ""
	}

	content = truncateText(content, available)
	line = prefix + icon + content
	line = padRightToWidth(line, maxWidth)
	return line, prefixWidth, icon, iconWidth, content
}

// =============================================================================
// FocusableInstance Interface
// =============================================================================

func (inst *Instance) SetFocus(focused bool) {
	if inst.focused == focused {
		return
	}
	inst.focused = focused
	inst.dirty = true
}

func (inst *Instance) HasFocus() bool { return inst.focused }

func (inst *Instance) IsDisabled() bool { return false }

