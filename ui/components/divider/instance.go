package divider

import (
	"github.com/wwsheng009/mint/ui/components/internal/proputil"
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

// Instance is the runtime entity for Divider components.
// It persists across renders and holds all state.
type Instance struct {
	// === Identification ===
	key string

	// === Props (from VNode, may change each render) ===
	label        string
	dividerStyle Style
	orientation  Orientation
	thickness    int
	divStyle     style.Style
	fillWidth    bool

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

// NewInstance creates a new DividerInstance from props.
func NewInstance(props rtui.Props) *Instance {
	inst := &Instance{
		key:          proputil.GetString(props, "key", ""),
		label:        proputil.GetString(props, "label", ""),
		dividerStyle: getStyleEnumProp(props, StyleSolid),
		orientation:  getOrientationProp(props, Horizontal),
		thickness:    proputil.GetInt(props, "thickness", 1),
		divStyle:     proputil.GetStyle(props, "style", style.Style{}),
		fillWidth:    proputil.GetBool(props, "fillWidth", true),
		dirty:        true,
	}

	// Ensure thickness is at least 1
	if inst.thickness < 1 {
		inst.thickness = 1
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
	oldLabel := inst.label
	oldStyle := inst.dividerStyle
	oldThickness := inst.thickness

	inst.label = proputil.GetString(props, "label", inst.label)
	inst.dividerStyle = getStyleEnumProp(props, inst.dividerStyle)
	inst.orientation = getOrientationProp(props, inst.orientation)
	inst.thickness = proputil.GetInt(props, "thickness", inst.thickness)
	inst.divStyle = proputil.GetStyle(props, "style", style.Style{})
	inst.fillWidth = proputil.GetBool(props, "fillWidth", inst.fillWidth)

	if inst.thickness < 1 {
		inst.thickness = 1
	}

	// Check if props changed
	changed := oldLabel != inst.label ||
		oldStyle != inst.dividerStyle ||
		oldThickness != inst.thickness

	if changed {
		inst.dirty = true
	}
	return changed
}

// GetProps implements ComponentInstance.
func (inst *Instance) GetProps() rtui.Props {
	return rtui.Props{
		"key":          inst.key,
		"label":        inst.label,
		"dividerStyle": inst.dividerStyle,
		"orientation":  inst.orientation,
		"thickness":    inst.thickness,
		"fillWidth":    inst.fillWidth,
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

// GetContext implements ComponentInstance (no hooks for Divider).
func (inst *Instance) GetContext() *rtui.ComponentContext {
	return nil
}

// =============================================================================
// PaintableInstance Interface
// =============================================================================

// Paint implements PaintableInstance.
func (inst *Instance) Paint(x, y int) []paint.DrawCmd {
	if inst == nil {
		return nil
	}

	// Get width from bounds
	width := inst.bounds[2]
	if width <= 0 {
		width = 80 // Default width
	}

	// Build the divider string based on style
	var dividerStr string
	switch inst.dividerStyle {
	case StyleSolid:
		dividerStr = strings.Repeat("─", width)
	case StyleDashed:
		dividerStr = buildDashed(width)
	case StyleDotted:
		dividerStr = buildDotted(width)
	case StyleDouble:
		dividerStr = strings.Repeat("═", width)
	default:
		dividerStr = strings.Repeat("─", width)
	}

	// If there's a label, insert it in the middle
	if inst.label != "" {
		dividerStr = inst.insertLabel(dividerStr, width)
	}

	return []paint.DrawCmd{{
		X:     x,
		Y:     y,
		Text:  dividerStr,
		Style: inst.divStyle,
	}}
}

// buildDashed creates a dashed line pattern.
func buildDashed(width int) string {
	if width <= 0 {
		return ""
	}
	pattern := "- "
	repeats := (width + 1) / 2
	result := strings.Repeat(pattern, repeats)
	if len(result) > width {
		result = result[:width]
	}
	return result
}

// buildDotted creates a dotted line pattern.
func buildDotted(width int) string {
	if width <= 0 {
		return ""
	}
	pattern := "· "
	repeats := (width + 1) / 2
	result := strings.Repeat(pattern, repeats)
	if len(result) > width {
		result = result[:width]
	}
	return result
}

// insertLabel inserts the label in the center of the divider.
func (inst *Instance) insertLabel(dividerStr string, width int) string {
	textLen := utf8.RuneCountInString(inst.label)
	if textLen >= width {
		return inst.label[:width]
	}

	// Calculate padding
	totalPadding := width - textLen
	leftPadding := totalPadding / 2

	// Build result
	runes := []rune(dividerStr)
	if len(runes) < width {
		// Extend if needed
		for len(runes) < width {
			runes = append(runes, '─')
		}
	}

	// Create left part, label, right part
	leftPart := string(runes[:leftPadding])
	var rightPart string
	if leftPadding+textLen < len(runes) {
		rightPart = string(runes[leftPadding+textLen:])
	}

	return leftPart + inst.label + rightPart
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
// Calculates the divider's ideal size given the constraints.
func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
	if inst == nil {
		return layout.Size{}
	}

	var width, height int

	if inst.orientation == Horizontal {
		// Horizontal divider: fill width, use thickness for height
		if inst.fillWidth && constraints.MaxWidth > 0 {
			width = constraints.MaxWidth
		} else {
			// Use label width or minimum width
			minWidth := utf8.RuneCountInString(inst.label) + 4
			if minWidth < 20 {
				minWidth = 20
			}
			width = minWidth
		}
		height = inst.thickness

		// If label is longer than available width, need more space
		labelLen := utf8.RuneCountInString(inst.label)
		if labelLen > 0 && labelLen+4 > width {
			width = labelLen + 4
		}
	} else {
		// Vertical divider: use thickness for width, fill height
		width = inst.thickness
		if inst.fillWidth && constraints.MaxHeight > 0 {
			height = constraints.MaxHeight
		} else {
			height = 20 // Default height
		}
	}

	// Apply constraints
	width = constraints.ConstrainWidth(width)
	height = constraints.ConstrainHeight(height)

	// Apply explicit style dimensions if set
	if inst.divStyle.Width > 0 {
		width = constraints.ConstrainWidth(inst.divStyle.Width)
	}
	if inst.divStyle.Height > 0 {
		height = constraints.ConstrainHeight(inst.divStyle.Height)
	}

	return layout.Size{Width: width, Height: height}
}

// GetNaturalSize returns the natural (unconstrained) size of the divider.
func (inst *Instance) GetNaturalSize() (width, height int) {
	if inst.orientation == Horizontal {
		// Minimum useful width
		minWidth := utf8.RuneCountInString(inst.label) + 4
		if minWidth < 20 {
			minWidth = 20
		}
		return minWidth, inst.thickness
	}
	return inst.thickness, 20
}

// ClearDirty clears the dirty flag.
func (inst *Instance) ClearDirty() {
	inst.dirty = false
}

// GetStyle returns the divider style.
func (inst *Instance) GetStyle() style.Style {
	return inst.divStyle
}

// SetStyle sets the divider style.
func (inst *Instance) SetStyle(s style.Style) {
	inst.divStyle = s
}

// =============================================================================
// Prop Extraction Helpers
// =============================================================================

func getStyleEnumProp(props rtui.Props, def Style) Style {
	if v, ok := props["dividerStyle"]; ok {
		if s, ok := v.(Style); ok {
			return s
		}
	}
	return def
}

func getOrientationProp(props rtui.Props, def Orientation) Orientation {
	if v, ok := props["orientation"]; ok {
		if o, ok := v.(Orientation); ok {
			return o
		}
	}
	return def
}
