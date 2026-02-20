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

	// === Runtime State ===
	bounds [4]int // x, y, w, h
	dirty  bool
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

	inst.content = getStringProp(props, "content", inst.content)
	inst.textStyle = getStyleProp(props)
	inst.padding = getPaddingProp(props)
	inst.textAlign = getTextAlignProp(props, inst.textAlign)
	inst.maxWidth = getIntProp(props, "maxWidth", inst.maxWidth)

	// Check if props changed
	changed := oldContent != inst.content || oldMaxWidth != inst.maxWidth

	if changed {
		inst.dirty = true
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

// Paint implements PaintableInstance.
func (inst *Instance) Paint(x, y int) []paint.DrawCmd {
	if inst == nil || inst.content == "" {
		return nil
	}

	// Get padding
	paddingLeft := inst.padding[3] // left
	paddingRight := inst.padding[1] // right

	// Build the text with padding
	text := inst.content
	contentWidth := utf8.RuneCountInString(text)

	// Calculate natural width and layout width
	naturalWidth := contentWidth + paddingLeft + paddingRight
	layoutWidth := naturalWidth
	if inst.bounds[2] > 0 && inst.bounds[2] > naturalWidth {
		layoutWidth = inst.bounds[2]
	}

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
		Y:     y,
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

	// Width: count rune width
	contentWidth := utf8.RuneCountInString(content)

	// Height is always 1 for single-line text
	contentHeight := 1

	// Apply user-specified padding
	horizontalPadding := inst.padding[1] + inst.padding[3] // right + left
	verticalPadding := inst.padding[0] + inst.padding[2]   // top + bottom

	width := contentWidth + horizontalPadding
	height := contentHeight + verticalPadding

	// Apply maxWidth constraint if set
	if inst.maxWidth > 0 && width > inst.maxWidth {
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
