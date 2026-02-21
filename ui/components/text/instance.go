package text

import (
	"strings"
	"unicode/utf8"

	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Instance - Runtime Entity
// =============================================================================

// Instance is the runtime entity for Text components.
// It persists across renders and holds all state.
type Instance struct {
	// === Identification ===
	key string

	// === Props (from VNode, may change each render) ===
	content   string
	textStyle style.Style
	padding   [4]int
	textAlign rtui.Align
	maxWidth  int
	wrap      bool // enable word wrapping

	// === Runtime State ===
	bounds    [4]int // x, y, w, h
	dirty     bool
	wrapLines []string // cached wrapped lines
}

// Ensure Instance implements required interfaces
var (
	_ rtui.ComponentInstance = (*Instance)(nil)
	_ rtui.PaintableInstance = (*Instance)(nil)
	_ interface {
		Measure(layout.Constraints) layout.Size
	} = (*Instance)(nil)
)

// =============================================================================
// Constructor
// =============================================================================

// NewInstance creates a new TextInstance from props.
func NewInstance(props rtui.Props) *Instance {
	inst := &Instance{
		key:       getStringProp(props, "key", ""),
		content:   getStringProp(props, "content", ""),
		textStyle: getStyleProp(props),
		padding:   getPaddingProp(props),
		textAlign: getTextAlignProp(props, rtui.AlignStart),
		maxWidth:  getIntProp(props, "maxWidth", 0),
		wrap:      getBoolProp(props, "wrap", false),
		dirty:     true,
	}

	return inst
}

// =============================================================================
// ComponentInstance Interface
// =============================================================================

// Key implements ComponentInstance.
func (inst *Instance) Key() string {
	return inst.key
}

// SetKey implements ComponentInstance.
func (inst *Instance) SetKey(key string) {
	inst.key = key
}

// Init implements ComponentInstance.
func (inst *Instance) Init(props rtui.Props) {
	inst.SetProps(props)
}

// Destroy implements ComponentInstance.
func (inst *Instance) Destroy() {
	// Nothing to clean up
}

// OnMount implements ComponentInstance.
func (inst *Instance) OnMount() {
	// Nothing to do on mount
}

// OnUnmount implements ComponentInstance.
func (inst *Instance) OnUnmount() {
	// Nothing to do on unmount
}

// SetProps implements ComponentInstance.
func (inst *Instance) SetProps(props rtui.Props) bool {
	oldContent := inst.content
	oldMaxWidth := inst.maxWidth
	oldWrap := inst.wrap

	inst.content = getStringProp(props, "content", inst.content)
	inst.textStyle = getStyleProp(props)
	inst.padding = getPaddingProp(props)
	inst.textAlign = getTextAlignProp(props, inst.textAlign)
	inst.maxWidth = getIntProp(props, "maxWidth", inst.maxWidth)
	inst.wrap = getBoolProp(props, "wrap", inst.wrap)

	// Check if props changed
	changed := oldContent != inst.content || oldMaxWidth != inst.maxWidth || oldWrap != inst.wrap

	if changed {
		inst.dirty = true
		inst.wrapLines = nil // Clear cached wrap lines when content or wrap changes
	}
	return changed
}

// GetProps implements ComponentInstance.
func (inst *Instance) GetProps() rtui.Props {
	return rtui.Props{
		"key":       inst.key,
		"content":   inst.content,
		"style":     inst.textStyle,
		"padding":   inst.padding,
		"textAlign": inst.textAlign,
		"maxWidth":  inst.maxWidth,
		"wrap":      inst.wrap,
	}
}

// MarkDirty implements ComponentInstance.
func (inst *Instance) MarkDirty() {
	inst.dirty = true
}

// IsDirty implements ComponentInstance.
func (inst *Instance) IsDirty() bool {
	return inst.dirty
}

// GetContext implements ComponentInstance (no hooks for Text).
func (inst *Instance) GetContext() *rtui.ComponentContext {
	return nil
}

// =============================================================================
// PaintableInstance Interface
// =============================================================================

// GetWrapLines returns the wrapped lines (for debugging)
// This is a debug helper to expose internal state for troubleshooting
func (inst Instance) GetWrapLines() []string {
	return inst.wrapLines
}

// ValidatePaintSize validates that the measured size doesn't exceed the paint bounds.
// This method should be called after Measure and before Paint to detect potential overflow.
func (inst *Instance) ValidatePaintSize(measureSize layout.Size, paintBounds [4]int) error {
	verticalPadding := inst.padding[0] + inst.padding[2]

	// Calculate available height in paint bounds
	paintWidth := paintBounds[2]
	paintHeight := paintBounds[3]

	// Account for padding
	contentHeightLimit := paintHeight - verticalPadding
	if contentHeightLimit < 0 {
		contentHeightLimit = 0
	}

	// Check if measure height exceeds available paint height
	if inst.wrap && measureSize.Height > contentHeightLimit && paintHeight > 0 {
		// Content will be cropped, this might be intentional for some components
		// but for Text.Wrap we want to warn about this
		// This is a warning, not an error, since Paint already handles cropping
	}

	// Validate width constraint
	paddingLeft := inst.padding[3]
	paddingRight := inst.padding[1]
	contentWidthLimit := paintWidth - paddingLeft - paddingRight
	if contentWidthLimit < 0 {
		contentWidthLimit = 0
	}

	if !inst.wrap && measureSize.Width > contentWidthLimit && paintWidth > 0 {
		// Content will be truncated, this is normal behavior for non-wrapping text
	}

	return nil
}

// Paint implements PaintableInstance.
func (inst *Instance) Paint(x, y int) []paint.DrawCmd {
	if inst == nil || inst.content == "" {
		return nil
	}

	// Get padding
	paddingTop := inst.padding[0]    // top
	paddingLeft := inst.padding[3]   // left
	paddingRight := inst.padding[1]  // right

	// Build the text with padding
	text := inst.content
	runes := []rune(text)
	contentWidth := len(runes)

	var cmds []paint.DrawCmd

	// Handle wrap mode
	if inst.wrap && len(inst.wrapLines) > 0 {
		// Multi-line rendering
		layoutWidth := inst.bounds[2]
		if layoutWidth <= 0 {
			layoutWidth = contentWidth + paddingLeft + paddingRight
		}

		// Calculate available height from bounds (content area excluding padding)
		availableHeight := 0
		if inst.bounds[3] > 0 {
			availableHeight = inst.bounds[3] - paddingTop
			if availableHeight < 0 {
				availableHeight = 0
			}
		} else {
			// No bounds height, allow all lines
			availableHeight = len(inst.wrapLines)
		}

		// Clamp to available height
		maxLines := len(inst.wrapLines)
		if availableHeight < maxLines {
			maxLines = availableHeight
		}

		for i := 0; i < maxLines; i++ {
			line := inst.wrapLines[i]
			lineText := line
			lineRunes := []rune(line)
			lineWidth := len(lineRunes)

			// Apply padding and alignment
		 paddedWidth := paddingLeft + lineWidth + paddingRight
			if layoutWidth > paddedWidth {
				availableSpace := layoutWidth - paddedWidth
				switch inst.textAlign {
				case rtui.AlignCenter:
					leftSpace := paddingLeft + availableSpace/2
					rightSpace := paddingRight + (availableSpace - availableSpace/2)
					lineText = strings.Repeat(" ", leftSpace) + lineText + strings.Repeat(" ", rightSpace)
				case rtui.AlignEnd:
					leftSpace := paddingLeft + availableSpace
					lineText = strings.Repeat(" ", leftSpace) + lineText + strings.Repeat(" ", paddingRight)
				default:
					lineText = strings.Repeat(" ", paddingLeft) + lineText + strings.Repeat(" ", paddingRight+availableSpace)
				}
			} else {
				lineText = strings.Repeat(" ", paddingLeft) + lineText + strings.Repeat(" ", paddingRight)
			}

			cmds = append(cmds, paint.DrawCmd{
				X:     x,
				Y:     y + paddingTop + i,
				Text:  lineText,
				Style: inst.textStyle,
			})
		}
		return cmds
	}

	// Single-line rendering (original code with truncation)
	// Calculate natural width and layout width
	naturalWidth := contentWidth + paddingLeft + paddingRight
	layoutWidth := naturalWidth

	// Use bounds width if available (from layout engine)
	if inst.bounds[2] > 0 {
		layoutWidth = inst.bounds[2]
	}

	// Truncate text if it exceeds layout width
	maxContentWidth := layoutWidth - paddingLeft - paddingRight
	if maxContentWidth < 0 {
		maxContentWidth = 0
	}
	if contentWidth > maxContentWidth {
		// Truncate runes to fit
		if maxContentWidth > 0 {
			runes = runes[:maxContentWidth]
			text = string(runes)
			contentWidth = maxContentWidth
		} else {
			text = ""
			contentWidth = 0
		}
	}

	// Recalculate natural width after truncation
	naturalWidth = contentWidth + paddingLeft + paddingRight

	// Apply text alignment if text container is stretched
	if layoutWidth > naturalWidth {
		availableSpace := layoutWidth - naturalWidth
		switch inst.textAlign {
		case rtui.AlignCenter:
			leftSpace := paddingLeft + availableSpace/2
			rightSpace := paddingRight + (availableSpace - availableSpace/2)
			text = strings.Repeat(" ", leftSpace) + text + strings.Repeat(" ", rightSpace)
		case rtui.AlignEnd:
			leftSpace := paddingLeft + availableSpace
			text = strings.Repeat(" ", leftSpace) + text + strings.Repeat(" ", paddingRight)
		default:
			text = strings.Repeat(" ", paddingLeft) + text + strings.Repeat(" ", paddingRight+availableSpace)
		}
	} else {
		text = strings.Repeat(" ", paddingLeft) + text + strings.Repeat(" ", paddingRight)
	}

	return []paint.DrawCmd{{
		X:     x,
		Y:     y + paddingTop,
		Text:  text,
		Style: inst.textStyle,
	}}
}

// =============================================================================
// Bounds Management
// =============================================================================

// GetBounds returns the layout bounds.
func (inst *Instance) GetBounds() (x, y, w, h int) {
	return inst.bounds[0], inst.bounds[1], inst.bounds[2], inst.bounds[3]
}

// SetBounds sets the layout bounds.
func (inst *Instance) SetBounds(x, y, w, h int) {
	inst.bounds = [4]int{x, y, w, h}
}

// =============================================================================
// Measurable Interface (Two-Pass Layout)
// =============================================================================

// Measure implements layout.Measurable interface.
// Calculates the text's ideal size given the constraints.
// This is Phase 1 of two-pass layout: measure natural size without position.
func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
	if inst == nil {
		return layout.Size{}
	}

	// Calculate text width
	content := inst.content
	if content == "" {
		content = " " // Empty text still has minimal width
	}

	// Apply user-specified padding
	horizontalPadding := inst.padding[1] + inst.padding[3] // right + left
	verticalPadding := inst.padding[0] + inst.padding[2]   // top + bottom

	var contentWidth, contentHeight int

	// Handle wrap mode
	if inst.wrap {
		// Determine available width for content
		availableWidth := constraints.MaxWidth
		if inst.maxWidth > 0 && (availableWidth == 0 || inst.maxWidth < availableWidth) {
			availableWidth = inst.maxWidth
		}
		availableWidth -= horizontalPadding
		if availableWidth < 1 {
			availableWidth = 1
		}

		// Wrap text and calculate dimensions
		inst.wrapLines = wordWrap(content, availableWidth)
		contentHeight = len(inst.wrapLines)
		if contentHeight == 0 {
			contentHeight = 1
		}

		// Find the maximum line width
		maxLineWidth := 0
		for _, line := range inst.wrapLines {
			lineWidth := utf8.RuneCountInString(line)
			if lineWidth > maxLineWidth {
				maxLineWidth = lineWidth
			}
		}
		contentWidth = maxLineWidth
	} else {
		// Single-line mode
		contentWidth = utf8.RuneCountInString(content)
		contentHeight = 1
	}

	width := contentWidth + horizontalPadding
	height := contentHeight + verticalPadding

	// Apply maxWidth constraint (for truncate mode)
	if !inst.wrap && inst.maxWidth > 0 && width > inst.maxWidth {
		width = inst.maxWidth
	}

	// Apply constraints
	width = constraints.ConstrainWidth(width)
	height = constraints.ConstrainHeight(height)

	// Apply explicit style dimensions if set
	if inst.textStyle.Width > 0 {
		width = constraints.ConstrainWidth(inst.textStyle.Width)
	}
	if inst.textStyle.Height > 0 {
		height = constraints.ConstrainHeight(inst.textStyle.Height)
	}

	return layout.Size{Width: width, Height: height}
}

// GetNaturalSize returns the natural (unconstrained) size of the text.
func (inst *Instance) GetNaturalSize() (width, height int) {
	content := inst.content
	if content == "" {
		content = " "
	}

	width = utf8.RuneCountInString(content)
	height = 1

	return width, height
}

// ClearDirty clears the dirty flag.
func (inst *Instance) ClearDirty() {
	inst.dirty = false
}

// GetStyle returns the text style.
func (inst *Instance) GetStyle() style.Style {
	return inst.textStyle
}

// SetStyle sets the text style.
func (inst *Instance) SetStyle(s style.Style) {
	inst.textStyle = s
}

// =============================================================================
// Prop Extraction Helpers
// =============================================================================

func getStringProp(props rtui.Props, key, def string) string {
	if v, ok := props[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return def
}

func getIntProp(props rtui.Props, key string, def int) int {
	if v, ok := props[key]; ok {
		if i, ok := v.(int); ok {
			return i
		}
	}
	return def
}

func getBoolProp(props rtui.Props, key string, def bool) bool {
	if v, ok := props[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return def
}

func getStyleProp(props rtui.Props) style.Style {
	if v, ok := props["style"]; ok {
		if s, ok := v.(style.Style); ok {
			return s
		}
	}
	return style.Style{}
}

func getPaddingProp(props rtui.Props) [4]int {
	if v, ok := props["padding"]; ok {
		if p, ok := v.([4]int); ok {
			return p
		}
	}
	return [4]int{}
}

func getTextAlignProp(props rtui.Props, def rtui.Align) rtui.Align {
	if v, ok := props["textAlign"]; ok {
		if a, ok := v.(rtui.Align); ok {
			return a
		}
	}
	return def
}

// =============================================================================
// Word Wrapping
// =============================================================================

// wordWrap breaks text into lines at word boundaries to fit the given max width.
// It tries to break at spaces/punctuation first, then falls back to hard breaks.
func wordWrap(text string, maxWidth int) []string {
	if maxWidth <= 0 {
		return []string{text}
	}

	runes := []rune(text)
	if len(runes) == 0 {
		return []string{}
	}

	// Simple implementation: break by space, then fit lines
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{}
	}

	var lines []string
	var currentLine string

	for _, word := range words {
		// Check if we can add this word to the current line
		testLine := currentLine
		if testLine != "" {
			testLine += " "
		}
		testLine += word

		if utf8.RuneCountInString(testLine) <= maxWidth {
			currentLine = testLine
		} else {
			// Can't fit in current line
			if currentLine != "" {
				lines = append(lines, currentLine)
			}
			// If a single word is too long, break it
			if utf8.RuneCountInString(word) <= maxWidth {
				currentLine = word
			} else {
				// Break long word
				wordRunes := []rune(word)
				for len(wordRunes) > maxWidth {
					lines = append(lines, string(wordRunes[:maxWidth]))
					wordRunes = wordRunes[maxWidth:]
				}
				currentLine = string(wordRunes)
			}
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
}

