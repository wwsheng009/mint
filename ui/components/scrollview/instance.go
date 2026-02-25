package scrollview

import (
	"strings"

	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	newtext "github.com/wwsheng009/mint/ui/components/text"
)

// =============================================================================
// Instance - Runtime State & Behavior
// =============================================================================

// Instance manages scrollview runtime state and behavior.
type Instance struct {
	// === Props (from VNode) ===
	child         rtui.VNode
	width         int
	height        int
	scrollOffset  int
	showBorder    bool
	showIndicator bool
	instStyle     style.Style

	// === Runtime State ===
	dirty      bool
	bounds     [4]int // x, y, w, h
	totalLines int    // total content lines

	// === Cached Content ===
	contentLines []string // cached content lines
}

// Ensure Instance implements required interfaces
var (
	_ rtui.ComponentInstance        = (*Instance)(nil)
	_ rtui.PaintableInstance        = (*Instance)(nil)
	// Note: control.Instance intentionally not implemented - ScrollView doesn't need behaviors
	_ interface{ Measure(layout.Constraints) layout.Size } = (*Instance)(nil)
)

// =============================================================================
// Constructor
// =============================================================================

// NewInstance creates a new ScrollView Instance.
func NewInstance(props rtui.Props) *Instance {
	inst := &Instance{
		dirty:     true,
		instStyle: style.Style{},
	}
	inst.SetProps(props)
	return inst
}

// =============================================================================
// ComponentInstance Interface
// =============================================================================

// SetProps sets properties from VNode.
func (inst *Instance) SetProps(props rtui.Props) bool {
	inst.dirty = true

	if val, ok := props["style"].(style.Style); ok {
		inst.instStyle = val
	}
	if val, ok := props["width"].(int); ok {
		inst.width = val
	}
	if val, ok := props["height"].(int); ok {
		inst.height = val
	}
	if val, ok := props["scrollOffset"].(int); ok {
		inst.scrollOffset = val
	}
	if val, ok := props["showBorder"].(bool); ok {
		inst.showBorder = val
	}
	if val, ok := props["showIndicator"].(bool); ok {
		inst.showIndicator = val
	}
	if val, ok := props["child"].(rtui.VNode); ok {
		inst.child = val
		inst.extractContent()
	}
	return true
}

// GetStyle returns the instance style.
func (inst *Instance) GetStyle() style.Style {
	return inst.instStyle
}

// SetStyle sets the instance style.
func (inst *Instance) SetStyle(s style.Style) {
	inst.instStyle = s
	inst.dirty = true
}

// ClearDirty clears the dirty flag.
func (inst *Instance) ClearDirty() {
	inst.dirty = false
}

// IsDirty returns whether instance is dirty.
func (inst *Instance) IsDirty() bool {
	return inst.dirty
}

// Clone creates a copy of the instance.
func (inst *Instance) Clone() rtui.ComponentInstance {
	return &Instance{
		child:         inst.child,
		width:         inst.width,
		height:        inst.height,
		scrollOffset:  inst.scrollOffset,
		showBorder:    inst.showBorder,
		showIndicator: inst.showIndicator,
		instStyle:     inst.instStyle,
		dirty:         true,
		totalLines:    inst.totalLines,
		contentLines:  inst.contentLines,
	}
}

// =============================================================================
// ComponentInstance Interface - Additional Methods
// =============================================================================

// Key returns the component key.
func (inst *Instance) Key() string {
	return ""
}

// SetKey sets the component key.
func (inst *Instance) SetKey(key string) {
	// No-op for now
}

// Init initializes the instance with props.
func (inst *Instance) Init(props rtui.Props) {
	inst.SetProps(props)
}

// Destroy cleans up resources.
func (inst *Instance) Destroy() {
	// No cleanup needed
}

// OnMount is called when mounted.
func (inst *Instance) OnMount() {
	// No-op
}

// OnUnmount is called when unmounted.
func (inst *Instance) OnUnmount() {
	// No-op
}

// GetProps returns current props.
func (inst *Instance) GetProps() rtui.Props {
	return rtui.Props{
		"width":         inst.width,
		"height":        inst.height,
		"scrollOffset":  inst.scrollOffset,
		"showBorder":    inst.showBorder,
		"showIndicator": inst.showIndicator,
	}
}

// MarkDirty marks the instance as dirty.
func (inst *Instance) MarkDirty() {
	inst.dirty = true
}

// GetContext returns the component context (not used).
func (inst *Instance) GetContext() *rtui.ComponentContext {
	return nil
}

// =============================================================================
// Measure Implementation
// =============================================================================

// Measure calculates the natural size of the scrollview.
func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
	inst.extractContent()

	var w, h int

	// Width
	if inst.width > 0 {
		w = inst.width
	} else {
		// Auto-width: use max line length
		w = 20 // default
		for _, line := range inst.contentLines {
			if len(line) > w {
				w = len(line)
			}
		}
	}

	// Height
	if inst.height > 0 {
		h = inst.height
	} else {
		// Auto-height: show all content
		h = inst.totalLines
	}

	// Add border space if enabled
	if inst.showBorder {
		w += 2
		h += 2
	}

	return layout.Size{
		Width:  constraints.ConstrainWidth(w),
		Height: constraints.ConstrainHeight(h),
	}
}

// extractContent extracts text content from child for scrolling.
func (inst *Instance) extractContent() {
	if inst.child == nil {
		inst.contentLines = nil
		inst.totalLines = 0
		return
	}

	content := inst.extractTextContent(inst.child)
	if content == "" {
		inst.contentLines = nil
		inst.totalLines = 0
		return
	}

	inst.contentLines = strings.Split(content, "\n")
	inst.totalLines = len(inst.contentLines)
}

// extractTextContent recursively extracts text from VNode.
func (inst *Instance) extractTextContent(node rtui.VNode) string {
	if node == nil {
		return ""
	}

	// Check for content property
	if props := node.Props(); props != nil {
		if content, ok := props["content"]; ok && content != "" {
			if contentStr, ok := content.(string); ok {
				return contentStr
			}
		}
	}

	// Check if it's a TextVNode
	if textNode, ok := node.(*newtext.VNode); ok {
		return textNode.Content()
	}

	// Check if it implements Content()
	if contentNode, ok := node.(interface{ Content() string }); ok {
		return contentNode.Content()
	}

	// Recursively extract from children
	var result strings.Builder
	children := node.Children()
	for i, child := range children {
		if i > 0 {
			result.WriteString("\n")
		}
		result.WriteString(inst.extractTextContent(child))
	}

	return result.String()
}

// =============================================================================
// PaintableInstance Interface
// =============================================================================

// Paint renders the scrollview to draw commands.
func (inst *Instance) Paint(x, y int) []paint.DrawCmd {
	var cmds []paint.DrawCmd

	// Ensure content is extracted before painting
	inst.extractContent()

	// Content width and height (from user settings)
	contentW := inst.width
	contentH := inst.height

	// Auto-height: use total lines if height not specified
	if contentH == 0 {
		contentH = inst.totalLines
	}

	// If no border, just draw content
	if !inst.showBorder {
		// Clamp scroll offset
		maxOffset := inst.totalLines - contentH
		if maxOffset < 0 {
			maxOffset = 0
		}
		if inst.scrollOffset < 0 {
			inst.scrollOffset = 0
		}
		if inst.scrollOffset > maxOffset {
			inst.scrollOffset = maxOffset
		}

		startLine := inst.scrollOffset
		endLine := startLine + contentH
		if endLine > inst.totalLines {
			endLine = inst.totalLines
		}

		for i := 0; i < contentH; i++ {
			lineIdx := startLine + i
			if lineIdx < len(inst.contentLines) {
				line := inst.contentLines[lineIdx]
				if len(line) > contentW {
					line = line[:contentW]
				}
				cmds = append(cmds, paint.NewTextCmd(x, y+i, line, inst.instStyle))
			}
		}
		return cmds
	}

	// With border: draw border and content
	// Total visual size: contentW+2 x contentH+2
	borderStyle := style.Style{FG: inst.instStyle.FG}
	if borderStyle.FG == "" {
		borderStyle.FG = "white"
	}

	// Clamp scroll offset
	maxOffset := inst.totalLines - contentH
	if maxOffset < 0 {
		maxOffset = 0
	}
	if inst.scrollOffset < 0 {
		inst.scrollOffset = 0
	}
	if inst.scrollOffset > maxOffset {
		inst.scrollOffset = maxOffset
	}

	startLine := inst.scrollOffset
	endLine := startLine + contentH
	if endLine > inst.totalLines {
		endLine = inst.totalLines
	}

	// Top border
	topBorder := "┌" + strings.Repeat("─", contentW) + "┐"
	cmds = append(cmds, paint.NewTextCmd(x, y, topBorder, borderStyle))

	// Content lines with side borders
	contentX := x + 1
	contentY := y + 1
	for i := 0; i < contentH; i++ {
		// Left border
		cmds = append(cmds, paint.NewTextCmd(x, contentY+i, "│", borderStyle))

		// Content line
		var line string
		lineIdx := startLine + i
		if lineIdx < len(inst.contentLines) {
			line = inst.contentLines[lineIdx]
		}
		// Truncate or pad line to content width
		if len(line) > contentW {
			line = line[:contentW]
		}
		if len(line) < contentW {
			line += strings.Repeat(" ", contentW-len(line))
		}
		cmds = append(cmds, paint.NewTextCmd(contentX, contentY+i, line, inst.instStyle))

		// Right border with scroll indicator
		rightChar := "│"
		if inst.showIndicator && inst.totalLines > contentH {
			if i == contentH-1 {
				if inst.scrollOffset == 0 {
					rightChar = "↓"
				} else if inst.scrollOffset >= maxOffset {
					rightChar = "↑"
				} else {
					rightChar = "↕"
				}
			}
		}
		cmds = append(cmds, paint.NewTextCmd(x+contentW+1, contentY+i, rightChar, borderStyle))
	}

	// Bottom border
	bottomBorder := "└" + strings.Repeat("─", contentW) + "┘"
	cmds = append(cmds, paint.NewTextCmd(x, y+contentH+1, bottomBorder, borderStyle))

	return cmds
}

// SetBounds sets the layout bounds.
func (inst *Instance) SetBounds(x, y, w, h int) {
	inst.bounds = [4]int{x, y, w, h}
}

// GetBounds returns the layout bounds.
func (inst *Instance) GetBounds() (x, y, w, h int) {
	return inst.bounds[0], inst.bounds[1], inst.bounds[2], inst.bounds[3]
}

// =============================================================================
// Scroll Control Methods
// =============================================================================

// ScrollBy scrolls by the given delta.
func (inst *Instance) ScrollBy(delta int) int {
	viewportH := inst.height
	if inst.showBorder && viewportH > 2 {
		viewportH -= 2
	}

	maxOffset := inst.totalLines - viewportH
	if maxOffset < 0 {
		maxOffset = 0
	}

	newOffset := inst.scrollOffset + delta
	if newOffset < 0 {
		newOffset = 0
	}
	if newOffset > maxOffset {
		newOffset = maxOffset
	}

	inst.scrollOffset = newOffset
	inst.dirty = true
	return newOffset
}

// ScrollTo scrolls to an absolute position.
func (inst *Instance) ScrollTo(offset int) int {
	viewportH := inst.height
	if inst.showBorder && viewportH > 2 {
		viewportH -= 2
	}

	maxOffset := inst.totalLines - viewportH
	if maxOffset < 0 {
		maxOffset = 0
	}

	if offset < 0 {
		offset = 0
	}
	if offset > maxOffset {
		offset = maxOffset
	}

	inst.scrollOffset = offset
	inst.dirty = true
	return offset
}

// ScrollTop scrolls to the top.
func (inst *Instance) ScrollTop() {
	inst.scrollOffset = 0
	inst.dirty = true
}

// ScrollBottom scrolls to the bottom.
func (inst *Instance) ScrollBottom() {
	viewportH := inst.height
	if inst.showBorder && viewportH > 2 {
		viewportH -= 2
	}

	maxOffset := inst.totalLines - viewportH
	if maxOffset < 0 {
		inst.scrollOffset = 0
	} else {
		inst.scrollOffset = maxOffset
	}
	inst.dirty = true
}

// PageUp scrolls up by one page.
func (inst *Instance) PageUp() int {
	viewportH := inst.height
	if inst.showBorder && viewportH > 2 {
		viewportH -= 2
	}
	return inst.ScrollBy(-viewportH)
}

// PageDown scrolls down by one page.
func (inst *Instance) PageDown() int {
	viewportH := inst.height
	if inst.showBorder && viewportH > 2 {
		viewportH -= 2
	}
	return inst.ScrollBy(viewportH)
}

// CanScrollUp returns true if can scroll up.
func (inst *Instance) CanScrollUp() bool {
	return inst.scrollOffset > 0
}

// CanScrollDown returns true if can scroll down.
func (inst *Instance) CanScrollDown() bool {
	viewportH := inst.height
	if inst.showBorder && viewportH > 2 {
		viewportH -= 2
	}
	maxOffset := inst.totalLines - viewportH
	return inst.scrollOffset < maxOffset && maxOffset > 0
}

// GetScrollOffset returns current scroll offset.
func (inst *Instance) GetScrollOffset() int {
	return inst.scrollOffset
}

// GetTotalLines returns total content lines.
func (inst *Instance) GetTotalLines() int {
	return inst.totalLines
}

// GetViewportSize returns viewport height.
func (inst *Instance) GetViewportSize() int {
	viewportH := inst.height
	if inst.showBorder && viewportH > 2 {
		viewportH -= 2
	}
	return viewportH
}

// IsScrollable returns true if content is larger than viewport.
func (inst *Instance) IsScrollable() bool {
	viewportH := inst.height
	if inst.showBorder && viewportH > 2 {
		viewportH -= 2
	}
	return inst.totalLines > viewportH
}

// =============================================================================
// ActionTarget Interface
// =============================================================================

// HandleAction implements ActionTarget interface.
func (inst *Instance) HandleAction(act *action.Action) bool {
	if act == nil {
		return false
	}

	switch act.Type {
	case action.ActionScroll:
		if delta, ok := act.GetPayloadInt(); ok {
			inst.ScrollBy(delta)
			return true
		}

	case action.ActionNavigateUp:
		inst.ScrollBy(-1)
		return true

	case action.ActionNavigateDown:
		inst.ScrollBy(1)
		return true

	case action.ActionNavigatePageUp:
		inst.PageUp()
		return true

	case action.ActionNavigatePageDown:
		inst.PageDown()
		return true

	case action.ActionNavigateHome:
		inst.ScrollTop()
		return true

	case action.ActionNavigateEnd:
		inst.ScrollBottom()
		return true
	}

	return false
}

// GetSupportedActions implements ActionTarget interface.
func (inst *Instance) GetSupportedActions() []action.ActionType {
	return []action.ActionType{
		action.ActionScroll,
		action.ActionNavigateUp,
		action.ActionNavigateDown,
		action.ActionNavigatePageUp,
		action.ActionNavigatePageDown,
		action.ActionNavigateHome,
		action.ActionNavigateEnd,
	}
}

// CanHandleAction implements ActionTarget interface.
func (inst *Instance) CanHandleAction(act *action.Action) bool {
	if act == nil {
		return false
	}

	switch act.Type {
	case action.ActionScroll:
		return inst.IsScrollable()

	case action.ActionNavigateUp:
		return inst.CanScrollUp()

	case action.ActionNavigateDown:
		return inst.CanScrollDown()

	case action.ActionNavigatePageUp:
		return inst.CanScrollUp()

	case action.ActionNavigatePageDown:
		return inst.CanScrollDown()

	case action.ActionNavigateHome:
		return inst.scrollOffset > 0

	case action.ActionNavigateEnd:
		viewportH := inst.height
		if inst.showBorder && viewportH > 2 {
			viewportH -= 2
		}
		maxOffset := inst.totalLines - viewportH
		return inst.scrollOffset < maxOffset
	}

	return false
}

// =============================================================================
// ScrollableActionTarget Interface
// =============================================================================

// CanScroll implements ScrollableActionTarget interface.
func (inst *Instance) CanScroll(delta int) bool {
	if delta > 0 {
		return inst.scrollOffset > 0
	} else if delta < 0 {
		viewportH := inst.height
		if inst.showBorder && viewportH > 2 {
			viewportH -= 2
		}
		maxOffset := inst.totalLines - viewportH
		return inst.scrollOffset < maxOffset
	}
	return inst.IsScrollable()
}

// Scroll implements ScrollableActionTarget interface.
func (inst *Instance) Scroll(delta int) bool {
	oldOffset := inst.scrollOffset
	inst.ScrollBy(delta)
	return inst.scrollOffset != oldOffset
}

// GetScrollPosition implements ScrollableActionTarget interface.
func (inst *Instance) GetScrollPosition() (current, total, visible int) {
	viewportH := inst.height
	if inst.showBorder && viewportH > 2 {
		viewportH -= 2
	}
	return inst.scrollOffset, inst.totalLines, viewportH
}
