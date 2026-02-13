package data

import (
	"strings"
	"unicode/utf8"

	"github.com/wwsheng009/mint/framework/action"
	"github.com/wwsheng009/mint/framework/cmd"
	"github.com/wwsheng009/mint/framework/component"
	frameworkevent "github.com/wwsheng009/mint/framework/event"
	"github.com/wwsheng009/mint/runtime"
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
	"github.com/wwsheng009/mint/runtime/paint"
	runtimeplatform "github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/ui"
)

// Interface implementation assertions
var _ runtime.Measurable = (*ListVNode)(nil)
var _ frameworkevent.Component = (*ListVNode)(nil)
var _ component.Updater = (*ListVNode)(nil)
var _ action.ActionTarget = (*ListVNode)(nil)
var _ action.FocusableActionTarget = (*ListVNode)(nil)
var _ action.ScrollableActionTarget = (*ListVNode)(nil)
var _ action.SelectableActionTarget = (*ListVNode)(nil)

// ListVNode represents a general-purpose list component.
//
// Unlike VirtualList (which requires []interface{} items and a renderItem callback),
// ListVNode works directly with pre-formatted string rows and a column header.
// Each row is truncated to fit within the measured width from layout constraints,
// preventing horizontal overflow in constrained containers like the Inspector panel.
//
// Features:
//   - Respects BoxConstraints from the layout engine (Measure interface)
//   - Truncates each row to the available width with "…" suffix
//   - Optional header row with distinct style
//   - Optional separator line between header and data rows
//   - Supports per-row styling via RowStyle callback
//   - Keyboard navigation (up/down arrows, page up/down, home/end)
//   - Mouse click selection
//   - Scroll wheel support
//   - Focus and selection state tracking
//
// Usage:
//
//	list := data.ListBuilder().
//	    Header("Z   Node            Bounds       H  C").
//	    Rows(rows).
//	    OnSelect(func(index int) { ... }).
//	    Build()
type ListVNode struct {
	*ui.ElementVNode

	header      string                        // Optional column header text
	rows        []string                      // Pre-formatted data rows
	headerStyle style.Style                   // Style for the header row
	rowStyle    style.Style                   // Default style for data rows
	rowStyleFn  func(int, string) style.Style // Per-row style callback (index, text) → style
	showSep     bool                          // Show separator line between header and rows
	sepChar     rune                          // Separator character (default '─')
	emptyText   string                        // Text shown when rows is empty
	maxRows     int                           // Maximum visible rows (0 = unlimited)

	// Measured width from layout constraints (set during Measure)
	measuredWidth int

	// Event handling state
	focusIndex     int // Currently focused row index (-1 if none)
	selectedIndex  int // Currently selected row index (-1 if none)
	scrollOffset   int // Vertical scroll offset (in lines)
	viewportHeight int // Visible height for scrolling

	// Callbacks
	onChange func(int) // Called when selection changes (row index)
	onSelect func(int) // Called when a row is selected (click/enter)
	onScroll func(int) // Called when scroll position changes (new offset)

	// Layout bounds (for mouse hit testing)
	boundsX int
	boundsY int
	boundsW int
	boundsH int

	// ActionTarget support
	supportedActions []action.ActionType // Supported action types
}

// ListBuilderType provides a fluent API for building lists.
type ListBuilderType struct {
	node *ListVNode
}

// ListBuilder creates a new list builder.
func ListBuilder() *ListBuilderType {
	return &ListBuilderType{
		node: &ListVNode{
			ElementVNode:   ui.NewElement("list"),
			headerStyle:    style.Style{}.Bold(true),
			rowStyle:       style.Style{},
			showSep:        true,
			sepChar:        '─',
			emptyText:      "(empty)",
			maxRows:        0,
			focusIndex:     -1, // No focus initially
			selectedIndex:  -1, // No selection initially
			scrollOffset:   0,
			viewportHeight: 0,
			supportedActions: []action.ActionType{
				action.ActionNavigateUp,
				action.ActionNavigateDown,
				action.ActionNavigatePageUp,
				action.ActionNavigatePageDown,
				action.ActionNavigateHome,
				action.ActionNavigateEnd,
				action.ActionSelect,
				action.ActionEnter,
				action.ActionClick,
				action.ActionScroll,
			},
		},
	}
}

// Header sets the column header text.
func (b *ListBuilderType) Header(h string) *ListBuilderType {
	b.node.header = h
	return b
}

// Rows sets all data rows at once.
func (b *ListBuilderType) Rows(rows []string) *ListBuilderType {
	b.node.rows = rows
	return b
}

// AddRow appends a single row.
func (b *ListBuilderType) AddRow(row string) *ListBuilderType {
	b.node.rows = append(b.node.rows, row)
	return b
}

// HeaderStyle sets the style for the header row.
func (b *ListBuilderType) HeaderStyle(s style.Style) *ListBuilderType {
	b.node.headerStyle = s
	return b
}

// RowStyle sets the default style for data rows.
func (b *ListBuilderType) RowStyle(s style.Style) *ListBuilderType {
	b.node.rowStyle = s
	return b
}

// RowStyleFn sets a per-row style callback.
// The function receives the row index and text, and returns the style to use.
func (b *ListBuilderType) RowStyleFn(fn func(int, string) style.Style) *ListBuilderType {
	b.node.rowStyleFn = fn
	return b
}

// ShowSeparator controls whether a separator line is drawn between header and rows.
func (b *ListBuilderType) ShowSeparator(show bool) *ListBuilderType {
	b.node.showSep = show
	return b
}

// SepChar sets the separator character (default '─').
func (b *ListBuilderType) SepChar(ch rune) *ListBuilderType {
	b.node.sepChar = ch
	return b
}

// EmptyText sets the text displayed when there are no rows.
func (b *ListBuilderType) EmptyText(text string) *ListBuilderType {
	b.node.emptyText = text
	return b
}

// MaxRows limits the number of visible rows (0 = show all).
func (b *ListBuilderType) MaxRows(n int) *ListBuilderType {
	b.node.maxRows = n
	return b
}

// Width sets an explicit width hint (overridden by layout constraints).
func (b *ListBuilderType) Width(w int) *ListBuilderType {
	b.node.SetProp("width", w)
	return b
}

// Style sets the base visual style.
func (b *ListBuilderType) Style(s style.Style) *ListBuilderType {
	b.node.SetStyle(s)
	return b
}

// Key sets the key for reconciliation diffing.
func (b *ListBuilderType) Key(key string) *ListBuilderType {
	b.node.SetKey(key)
	return b
}

// OnChange sets the callback when selection changes (row index)
func (b *ListBuilderType) OnChange(fn func(int)) *ListBuilderType {
	b.node.onChange = fn
	return b
}

// OnSelect sets the callback when a row is selected via click or enter
func (b *ListBuilderType) OnSelect(fn func(int)) *ListBuilderType {
	b.node.onSelect = fn
	return b
}

// OnScroll sets the callback when scroll position changes
func (b *ListBuilderType) OnScroll(fn func(int)) *ListBuilderType {
	b.node.onScroll = fn
	return b
}

// Build returns the list VNode.
// Note: The list component must be directly accessible in the component tree
// to receive events. We return the underlying ListVNode directly
// instead of wrapping it in an ElementVNode.
func (b *ListBuilderType) Build() *ListVNode {
	return b.node
}

// =============================================================================
// Getters / Setters
// =============================================================================

func (l *ListVNode) Header() string                                 { return l.header }
func (l *ListVNode) Rows() []string                                 { return l.rows }
func (l *ListVNode) RowCount() int                                  { return len(l.rows) }
func (l *ListVNode) SetHeader(h string)                             { l.header = h }
func (l *ListVNode) SetRows(rows []string)                          { l.rows = rows }
func (l *ListVNode) SetRowStyleFn(fn func(int, string) style.Style) { l.rowStyleFn = fn }
func (l *ListVNode) EmptyText() string                              { return l.emptyText }

// =============================================================================
// Event State Getters / Setters
// =============================================================================

// FocusIndex returns the currently focused row index (-1 if none)
func (l *ListVNode) FocusIndex() int { return l.focusIndex }

// SetFocusIndex sets the focused row index
func (l *ListVNode) SetFocusIndex(index int) {
	if index < 0 || index >= len(l.rows) {
		return
	}
	l.focusIndex = index
	l.ensureVisible()
}

// SelectedIndex returns the currently selected row index (-1 if none)
func (l *ListVNode) SelectedIndex() int { return l.selectedIndex }

// SetSelectedIndex sets the selected row index
func (l *ListVNode) SetSelectedIndex(index int) {
	l.selectedIndex = index
	if l.onChange != nil && index >= 0 {
		l.onChange(index)
	}
}

// ScrollOffset returns the current scroll offset
func (l *ListVNode) ScrollOffset() int { return l.scrollOffset }

// SetScrollOffset sets the scroll offset
func (l *ListVNode) SetScrollOffset(offset int) {
	maxOffset := l.maxScrollOffset()
	if offset < 0 {
		offset = 0
	}
	if offset > maxOffset {
		offset = maxOffset
	}

	// Trigger callback whenever scroll offset is set (not just on change)
	// This ensures external code gets notified even for the same value
	l.scrollOffset = offset
	if l.onScroll != nil {
		l.onScroll(offset)
	}
}

// maxScrollOffset calculates the maximum scroll offset
func (l *ListVNode) maxScrollOffset() int {
	totalLines := l.visibleLineCount()
	if l.viewportHeight <= 0 {
		return 0
	}
	maxOffset := totalLines - l.viewportHeight
	if maxOffset < 0 {
		return 0
	}
	return maxOffset
}

// =============================================================================
// Measurable & Paintable Interface Implementation
// =============================================================================

// Measure implements runtime.Measurable.
//
// It calculates the natural size of the list (widest row × total lines)
// and then clamps to the constraints provided by the layout engine.
// The measured width is cached so Paint() can truncate rows accordingly.
func (l *ListVNode) Measure(constraints runtime.BoxConstraints) runtime.Size {
	if l == nil {
		return runtime.Size{Width: 0, Height: 0}
	}

	// Calculate natural width = max rune width of all lines
	naturalWidth := 0
	updateMax := func(s string) {
		if n := utf8.RuneCountInString(s); n > naturalWidth {
			naturalWidth = n
		}
	}
	if l.header != "" {
		updateMax(l.header)
	}
	for _, r := range l.rows {
		updateMax(r)
	}
	if naturalWidth < utf8.RuneCountInString(l.emptyText) && len(l.rows) == 0 {
		naturalWidth = utf8.RuneCountInString(l.emptyText)
	}

	// Explicit width prop overrides natural width
	if w, ok := l.Props()["width"].(int); ok && w > 0 {
		naturalWidth = w
	}

	// Calculate natural height
	naturalHeight := l.visibleLineCount()

	// Apply constraints
	width := naturalWidth
	height := naturalHeight

	if width < constraints.MinWidth {
		width = constraints.MinWidth
	}
	if constraints.HasBoundedWidth() && width > constraints.MaxWidth {
		width = constraints.MaxWidth
	}
	if height < constraints.MinHeight {
		height = constraints.MinHeight
	}
	if constraints.HasBoundedHeight() && height > constraints.MaxHeight {
		height = constraints.MaxHeight
	}

	// Cache measured width for Paint
	l.measuredWidth = width

	// Update viewport height if bounded
	oldViewportHeight := l.viewportHeight
	if constraints.HasBoundedHeight() {
		l.viewportHeight = constraints.MaxHeight
	} else {
		l.viewportHeight = 0 // No constraint, show all
	}

	// Trigger re-render if viewport height changed
	if oldViewportHeight != l.viewportHeight {
		l.ensureVisible()
	}

	return runtime.Size{Width: width, Height: height}
}

// Paint implements paint.Paintable.
//
// It renders the list as a series of DrawCmd, one per visible line.
// Each line is truncated to measuredWidth to prevent horizontal overflow.
// Supports scrolling, focus, and selection states.
func (l *ListVNode) Paint(x, y int) []paint.DrawCmd {
	if l == nil {
		return nil
	}

	// Determine effective width for truncation
	width := l.measuredWidth
	if width <= 0 {
		// Fallback: use explicit prop or default
		if w, ok := l.Props()["width"].(int); ok && w > 0 {
			width = w
		} else {
			width = 80
		}
	}

	// Store bounds for mouse hit testing
	l.boundsX = x
	l.boundsY = y
	l.boundsW = width
	l.boundsH = l.visibleLineCount()

	var cmds []paint.DrawCmd
	row := 0

	// Calculate visible range based on scroll offset
	startLine := 0
	endLine := l.visibleLineCount()

	if l.viewportHeight > 0 {
		// Virtual scrolling: only render visible lines
		startLine = l.scrollOffset
		endLine = startLine + l.viewportHeight
		if endLine > l.visibleLineCount() {
			endLine = l.visibleLineCount()
		}
	}

	// Header
	if l.header != "" && startLine == 0 {
		truncated := truncateRunes(l.header, width)
		cmds = append(cmds, paint.NewTextCmd(x, y+row, truncated, l.headerStyle))
		row++
	}

	// Separator (only if there are rows and we're showing it)
	if l.showSep && l.header != "" && len(l.rows) > 0 && startLine == 0 {
		sep := strings.Repeat(string(l.sepChar), width)
		cmds = append(cmds, paint.NewTextCmd(x, y+row, sep, l.Style()))
		row++
	}

	// Calculate which data rows to render
	visibleRows := l.rows
	overflow := false
	dataStartIdx := 0

	if l.header != "" {
		dataStartIdx = 1 // Skip header line
	}
	if l.showSep && l.header != "" && len(l.rows) > 0 && startLine == 0 {
		dataStartIdx = 2 // Skip header and separator
	}

	// Data rows
	if len(l.rows) == 0 {
		if row >= startLine && row < endLine {
			cmds = append(cmds, paint.NewTextCmd(x, y+row, truncateRunes(l.emptyText, width), l.Style()))
			row++
		}
	} else {
		// Apply scroll offset to row index
		scrolledDataStart := startLine - dataStartIdx
		if scrolledDataStart < 0 {
			scrolledDataStart = 0
		}

		visibleRows = l.rows
		if l.maxRows > 0 && len(visibleRows) > l.maxRows {
			visibleRows = visibleRows[:l.maxRows]
			overflow = true
		}

		// Apply scroll offset
		if scrolledDataStart < len(visibleRows) {
			endIdx := scrolledDataStart + (endLine - row)
			if endIdx > len(visibleRows) {
				endIdx = len(visibleRows)
			}
			visibleRows = visibleRows[scrolledDataStart:endIdx]
		} else {
			visibleRows = []string{}
		}

		for i, r := range visibleRows {
			actualRowIdx := scrolledDataStart + i

			// Build base style
			s := l.rowStyle
			if l.rowStyleFn != nil {
				s = l.rowStyleFn(actualRowIdx, r)
			}

			// Apply selection style
			if actualRowIdx == l.selectedIndex {
				s = s.Reverse(true).Bold(true)
			}

			// Apply focus style (if different from selection)
			if actualRowIdx == l.focusIndex && actualRowIdx != l.selectedIndex {
				s = s.Underline(true).Bold(true)
			}

			truncated := truncateRunes(r, width)
			cmds = append(cmds, paint.NewTextCmd(x, y+row, truncated, s))
			row++
		}

		if overflow && row < endLine {
			moreText := truncateRunes("... (more rows)", width)
			cmds = append(cmds, paint.NewTextCmd(x, y+row, moreText, l.Style()))
			row++
		}
	}

	return cmds
}

// =============================================================================
// Internal helpers
// =============================================================================

// visibleLineCount returns the total number of lines that will be rendered.
func (l *ListVNode) visibleLineCount() int {
	n := 0
	if l.header != "" {
		n++ // header
	}
	// Only show separator if there are rows (empty list doesn't need separator)
	if l.showSep && l.header != "" && len(l.rows) > 0 {
		n++ // separator
	}
	if len(l.rows) == 0 {
		n++ // empty text
	} else {
		visible := len(l.rows)
		if l.maxRows > 0 && visible > l.maxRows {
			visible = l.maxRows
			n++ // "... (more rows)" line
		}
		n += visible
	}
	return n
}

// truncateRunes truncates a string to at most maxRunes rune-width.
// If truncated, the last character is replaced with "…".
func truncateRunes(s string, maxRunes int) string {
	if maxRunes <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= maxRunes {
		// Pad with spaces to fill the width so the background style
		// covers the full row (consistent with Table behaviour).
		if len(runes) < maxRunes {
			return s + strings.Repeat(" ", maxRunes-len(runes))
		}
		return s
	}
	if maxRunes <= 1 {
		return "…"
	}
	// Keep maxRunes-1 characters and add ellipsis
	return string(runes[:maxRunes-1]) + "…"
}

// =============================================================================
// Event Handling - Component Interface Implementation
// =============================================================================

// SetBounds stores layout bounds for mouse hit testing.
func (l *ListVNode) SetBounds(x, y, width, height int) {
	l.boundsX = x
	l.boundsY = y
	l.boundsW = width
	l.boundsH = height
}

// GetBounds returns stored bounds (tuple form for hit testing).
func (l *ListVNode) GetBounds() (int, int, int, int) {
	return l.boundsX, l.boundsY, l.boundsW, l.boundsH
}

// Bounds returns bounds as array (used by some inspector utilities).
func (l *ListVNode) Bounds() [4]int {
	return [4]int{l.boundsX, l.boundsY, l.boundsW, l.boundsH}
}

// containsPoint reports whether screen coordinates are inside the list.
func (l *ListVNode) containsPoint(x, y int) bool {
	return x >= l.boundsX && x < l.boundsX+l.boundsW &&
		y >= l.boundsY && y < l.boundsY+l.boundsH
}

// HandleEvent handles mouse and keyboard events for the list.
// Implements frameworkevent.Component interface.
func (l *ListVNode) HandleEvent(ev frameworkevent.Event) bool {
	me, ok := ev.(*frameworkevent.MouseEvent)
	if !ok {
		return false
	}

	// Ignore events outside our bounds.
	if !l.containsPoint(me.X, me.Y) {
		return false
	}

	switch ev.Type() {
	case frameworkevent.EventMousePress:
		// Calculate which row was clicked
		localY := me.Y - l.boundsY
		if localY < 0 {
			return false
		}

		// Adjust for header and separator
		dataStartLine := 0
		if l.header != "" {
			dataStartLine = 1
		}
		if l.showSep && l.header != "" && len(l.rows) > 0 {
			dataStartLine++
		}

		target := l.scrollOffset + (localY - dataStartLine)
		if target < 0 || target >= len(l.rows) {
			return false
		}

		l.SetFocusIndex(target)
		l.SetSelectedIndex(target)
		if l.onSelect != nil {
			l.onSelect(target)
		}
		return true

	case frameworkevent.EventMouseWheel:
		// Scroll by wheel delta
		if delta, ok := me.Delta, me.Delta != 0; ok {
			l.ScrollBy(-delta) // Negative because wheel up is positive
			return true
		}
		return false

	case frameworkevent.EventMouseMove:
		// Hover focus (non-selecting) for visual cue
		localY := me.Y - l.boundsY
		if localY < 0 {
			return false
		}

		// Adjust for header and separator
		dataStartLine := 0
		if l.header != "" {
			dataStartLine = 1
		}
		if l.showSep && l.header != "" && len(l.rows) > 0 {
			dataStartLine++
		}

		target := l.scrollOffset + (localY - dataStartLine)
		if target < 0 || target >= len(l.rows) {
			return false
		}
		if target != l.focusIndex {
			l.SetFocusIndex(target)
			return true
		}
	}

	return false
}

// =============================================================================
// Msg/Cmd Architecture Support (component.Updater interface)
// =============================================================================

// Update implements component.Updater interface for Msg/Cmd architecture.
//
// Handles:
// - MouseMsg: List row selection, hover focus, scrolling
// - KeyMsg: Keyboard navigation (when focused)
func (l *ListVNode) Update(message runtimemsg.Msg) cmd.Cmd {
	switch msg := message.(type) {
	case *runtimemsg.MouseMsg:
		return l.updateMouse(msg)
	case *runtimemsg.KeyMsg:
		return l.updateKey(msg)
	}
	return nil
}

// updateMouse handles mouse messages for list interaction
func (l *ListVNode) updateMouse(mouseMsg *runtimemsg.MouseMsg) cmd.Cmd {
	switch mouseMsg.Action {
	case runtimemsg.MouseActionPress:
		// Calculate which row was clicked based on LocalY
		localY := mouseMsg.LocalY
		if localY < 0 {
			return nil
		}

		// Adjust for header and separator
		dataStartLine := 0
		if l.header != "" {
			dataStartLine = 1
		}
		if l.showSep && l.header != "" && len(l.rows) > 0 {
			dataStartLine++
		}

		target := l.scrollOffset + (localY - dataStartLine)
		if target < 0 || target >= len(l.rows) {
			return nil
		}

		l.SetFocusIndex(target)
		l.SetSelectedIndex(target)
		if l.onSelect != nil {
			l.onSelect(target)
		}
		return nil // TODO: Return Cmd to trigger re-render

	case runtimemsg.MouseActionWheel:
		// Scroll by wheel delta
		l.ScrollBy(-mouseMsg.Delta) // Negative because wheel up is positive
		return nil
	}

	return nil
}

// updateKey handles keyboard messages for navigation (when focused)
func (l *ListVNode) updateKey(keyMsg *runtimemsg.KeyMsg) cmd.Cmd {
	switch keyMsg.Special {
	case runtimeplatform.KeyUp:
		l.MoveUp()
		return nil

	case runtimeplatform.KeyDown:
		l.MoveDown()
		return nil

	case runtimeplatform.KeyPageUp:
		l.PageUp()
		return nil

	case runtimeplatform.KeyPageDown:
		l.PageDown()
		return nil

	case runtimeplatform.KeyHome:
		l.Home()
		return nil

	case runtimeplatform.KeyEnd:
		l.End()
		return nil

	case runtimeplatform.KeyEnter:
		// Select current focused row
		if l.focusIndex >= 0 && l.focusIndex < len(l.rows) {
			l.SetSelectedIndex(l.focusIndex)
			if l.onSelect != nil {
				l.onSelect(l.focusIndex)
			}
		}
		return nil
	}

	return nil
}

// =============================================================================
// Navigation Methods
// =============================================================================

// MoveUp moves focus to the previous row
func (l *ListVNode) MoveUp() {
	if l.focusIndex > 0 {
		l.SetFocusIndex(l.focusIndex - 1)
	}
}

// MoveDown moves focus to the next row
func (l *ListVNode) MoveDown() {
	if l.focusIndex < len(l.rows)-1 {
		l.SetFocusIndex(l.focusIndex + 1)
	}
}

// PageUp moves focus up by one page (viewport height)
func (l *ListVNode) PageUp() {
	pageSize := l.viewportHeight
	if pageSize <= 0 {
		pageSize = 10 // Default page size
	}

	newIndex := l.focusIndex - pageSize
	if newIndex < 0 {
		newIndex = 0
	}
	l.SetFocusIndex(newIndex)
}

// PageDown moves focus down by one page (viewport height)
func (l *ListVNode) PageDown() {
	pageSize := l.viewportHeight
	if pageSize <= 0 {
		pageSize = 10 // Default page size
	}

	newIndex := l.focusIndex + pageSize
	if newIndex >= len(l.rows) {
		newIndex = len(l.rows) - 1
	}
	l.SetFocusIndex(newIndex)
}

// Home moves focus to the first row
func (l *ListVNode) Home() {
	if len(l.rows) > 0 {
		l.SetFocusIndex(0)
	}
}

// End moves focus to the last row
func (l *ListVNode) End() {
	if len(l.rows) > 0 {
		l.SetFocusIndex(len(l.rows) - 1)
	}
}

// ScrollBy scrolls the list by delta lines
func (l *ListVNode) ScrollBy(delta int) {
	l.SetScrollOffset(l.scrollOffset + delta)
}

// ScrollTo scrolls to an absolute line
func (l *ListVNode) ScrollTo(index int) {
	l.SetScrollOffset(index)
}

// ensureVisible ensures the focused row is visible in the viewport
func (l *ListVNode) ensureVisible() {
	if l.viewportHeight <= 0 || l.focusIndex < 0 {
		return
	}

	// Scroll down if focus is below viewport
	if l.focusIndex >= l.scrollOffset+l.viewportHeight {
		l.scrollOffset = l.focusIndex - l.viewportHeight + 1
	}

	// Scroll up if focus is above viewport
	if l.focusIndex < l.scrollOffset {
		l.scrollOffset = l.focusIndex
	}

	// Clamp to valid range
	maxOffset := l.maxScrollOffset()
	if l.scrollOffset > maxOffset {
		l.scrollOffset = maxOffset
	}
	if l.scrollOffset < 0 {
		l.scrollOffset = 0
	}
}

// CanScrollUp checks if can scroll up
func (l *ListVNode) CanScrollUp() bool {
	return l.scrollOffset > 0
}

// CanScrollDown checks if can scroll down
func (l *ListVNode) CanScrollDown() bool {
	return l.scrollOffset < l.maxScrollOffset()
}

// =============================================================================
// ActionTarget Interface Implementation
// =============================================================================

// HandleAction implements ActionTarget interface
func (l *ListVNode) HandleAction(act *action.Action) bool {
	if act == nil {
		return false
	}

	// Handle action based on type
	switch act.Type {
	// Navigation actions
	case action.ActionNavigateUp:
		l.MoveUp()
		return true
	case action.ActionNavigateDown:
		l.MoveDown()
		return true
	case action.ActionNavigatePageUp:
		l.PageUp()
		return true
	case action.ActionNavigatePageDown:
		l.PageDown()
		return true
	case action.ActionNavigateHome:
		l.Home()
		return true
	case action.ActionNavigateEnd:
		l.End()
		return true

	// Selection actions
	case action.ActionSelect, action.ActionEnter:
		if l.focusIndex >= 0 && l.focusIndex < len(l.rows) {
			l.SetSelectedIndex(l.focusIndex)
			if l.onSelect != nil {
				l.onSelect(l.focusIndex)
			}
			return true
		}
		return false

	// Scroll action
	case action.ActionScroll:
		if delta, ok := act.GetPayloadInt(); ok {
			if delta > 0 && l.CanScrollDown() {
				l.ScrollBy(delta)
				return true
			} else if delta < 0 && l.CanScrollUp() {
				l.ScrollBy(delta)
				return true
			}
		}
		return false

	// Mouse click
	case action.ActionClick:
		// Click action already handled by HandleEvent
		if l.focusIndex >= 0 && l.focusIndex < len(l.rows) {
			l.SetSelectedIndex(l.focusIndex)
			return true
		}
		return false
	}

	return false
}

// GetSupportedActions implements ActionTarget interface
func (l *ListVNode) GetSupportedActions() []action.ActionType {
	if l.supportedActions == nil {
		return []action.ActionType{
			action.ActionNavigateUp,
			action.ActionNavigateDown,
			action.ActionNavigatePageUp,
			action.ActionNavigatePageDown,
			action.ActionNavigateHome,
			action.ActionNavigateEnd,
			action.ActionSelect,
			action.ActionEnter,
			action.ActionScroll,
			action.ActionClick,
		}
	}
	return l.supportedActions
}

// CanHandleAction implements ActionTarget interface
func (l *ListVNode) CanHandleAction(act *action.Action) bool {
	if act == nil {
		return false
	}

	// Check if action type is supported
	supported := l.GetSupportedActions()
	for _, supportedType := range supported {
		if supportedType == act.Type {
			return true
		}
	}

	return false
}

// =============================================================================
// FocusableActionTarget Interface Implementation
// =============================================================================

// Focus implements FocusableActionTarget interface
func (l *ListVNode) Focus() bool {
	if len(l.rows) == 0 {
		return false
	}
	// Focus on first row if no focus
	if l.focusIndex < 0 {
		l.SetFocusIndex(0)
		return true
	}
	return true
}

// Blur implements FocusableActionTarget interface
func (l *ListVNode) Blur() {
	// Clear visual focus indication (keep selection)
	l.focusIndex = -1
}

// IsFocused implements FocusableActionTarget interface
func (l *ListVNode) IsFocused() bool {
	return l.focusIndex >= 0 && l.focusIndex < len(l.rows)
}

// IsFocusable implements FocusableActionTarget interface
func (l *ListVNode) IsFocusable() bool {
	return len(l.rows) > 0
}

// =============================================================================
// ScrollableActionTarget Interface Implementation
// =============================================================================

// CanScroll implements ScrollableActionTarget interface
func (l *ListVNode) CanScroll(delta int) bool {
	if delta > 0 {
		return l.CanScrollDown()
	} else if delta < 0 {
		return l.CanScrollUp()
	}
	return false
}

// Scroll implements ScrollableActionTarget interface
func (l *ListVNode) Scroll(delta int) bool {
	if !l.CanScroll(delta) {
		return false
	}
	l.ScrollBy(delta)
	return true
}

// GetScrollPosition implements ScrollableActionTarget interface
func (l *ListVNode) GetScrollPosition() (int, int, int) {
	current := l.scrollOffset
	total := l.visibleLineCount()
	visible := l.viewportHeight
	if visible <= 0 {
		visible = total
	}
	return current, total, visible
}

// =============================================================================
// SelectableActionTarget Interface Implementation
// =============================================================================

// Select implements SelectableActionTarget interface
func (l *ListVNode) Select() bool {
	if !l.IsFocusable() {
		return false
	}
	if l.focusIndex >= 0 {
		l.SetSelectedIndex(l.focusIndex)
		if l.onSelect != nil {
			l.onSelect(l.focusIndex)
		}
	}
	return true
}

// IsSelected implements SelectableActionTarget interface
func (l *ListVNode) IsSelected() bool {
	return l.selectedIndex >= 0 && l.selectedIndex < len(l.rows)
}

// ToggleSelection implements SelectableActionTarget interface
func (l *ListVNode) ToggleSelection() bool {
	if l.HasSelection() {
		l.selectedIndex = -1
		return false
	}
	l.Select()
	return true
}

// GetSelectedCount implements SelectableActionTarget interface
func (l *ListVNode) GetSelectedCount() int {
	if l.HasSelection() {
		return 1
	}
	return 0
}

// HasSelection returns true if there's a selection
func (l *ListVNode) HasSelection() bool {
	return l.selectedIndex >= 0 && l.selectedIndex < len(l.rows)
}

// ClearSelection clears the current selection
func (l *ListVNode) ClearSelection() {
	l.selectedIndex = -1
}
