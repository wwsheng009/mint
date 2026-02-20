package border

import (
	"strings"

	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Instance - Runtime Entity
// =============================================================================

// Instance is the runtime entity for border components.
// It persists across renders and holds all state.
type Instance struct {
	// === Identification ===
	key string

	// === Border Props ===
	borderStyle BorderStyle
	borderColor style.Color
	borderLabel string

	// === Layout Props ===
	width  int
	height int
	flex   int

	// === Content ===
	child rtui.VNode

	// === Style ===
	instStyle style.Style

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

// NewInstance creates a new BorderInstance from props.
func NewInstance(props rtui.Props) *Instance {
	inst := &Instance{
		key:         getStringProp(props, "key", ""),
		borderStyle: getBorderStyleProp(props, "borderStyle", BorderSingle),
		borderColor: style.Color(getStringProp(props, "borderColor", "blue")),
		borderLabel: getStringProp(props, "borderLabel", ""),
		width:       getIntProp(props, "width", 0),
		height:      getIntProp(props, "height", 0),
		flex:        getIntProp(props, "flex", 0),
		child:       getChildProp(props),
		instStyle:   getStyleProp(props),
		dirty:       true,
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
	inst.child = nil
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
	oldStyle := inst.borderStyle
	oldWidth := inst.width
	oldHeight := inst.height

	if val, ok := props["borderStyle"].(BorderStyle); ok {
		inst.borderStyle = val
	}
	if val, ok := props["borderColor"].(string); ok {
		inst.borderColor = style.Color(val)
	}
	if val, ok := props["borderLabel"].(string); ok {
		inst.borderLabel = val
	}
	if val, ok := props["width"].(int); ok {
		inst.width = val
	}
	if val, ok := props["height"].(int); ok {
		inst.height = val
	}
	if val, ok := props["flex"].(int); ok {
		inst.flex = val
	}
	if val, ok := props["child"].(rtui.VNode); ok {
		inst.child = val
	}
	if val, ok := props["style"].(style.Style); ok {
		inst.instStyle = val
	}

	changed := oldStyle != inst.borderStyle || oldWidth != inst.width || oldHeight != inst.height
	if changed {
		inst.dirty = true
	}
	return changed
}

// GetProps implements ComponentInstance.
func (inst *Instance) GetProps() rtui.Props {
	return rtui.Props{
		"key":         inst.key,
		"borderStyle": inst.borderStyle,
		"borderColor": string(inst.borderColor),
		"borderLabel": inst.borderLabel,
		"width":       inst.width,
		"height":      inst.height,
		"flex":        inst.flex,
		"child":       inst.child,
		"style":       inst.instStyle,
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

// GetContext implements ComponentInstance (no hooks for Border).
func (inst *Instance) GetContext() *rtui.ComponentContext {
	return nil
}

// =============================================================================
// Measurable Interface
// =============================================================================

// Measure calculates the natural size of the bordered container.
// Border adds 2 * borderWidth to each dimension.
func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
	borderWidth := GetBorderWidth(inst.borderStyle)

	// Calculate inner constraints (reduced by border)
	innerMinWidth := max(0, constraints.MinWidth-borderWidth*2)
	innerMaxWidth := max(0, constraints.MaxWidth-borderWidth*2)
	innerMinHeight := max(0, constraints.MinHeight-borderWidth*2)
	innerMaxHeight := max(0, constraints.MaxHeight-borderWidth*2)

	innerConstraints := layout.Constraints{
		MinWidth:  innerMinWidth,
		MaxWidth:  innerMaxWidth,
		MinHeight: innerMinHeight,
		MaxHeight: innerMaxHeight,
	}

	var innerWidth, innerHeight int

	// Measure child if present
	if inst.child != nil {
		if measurable, ok := inst.child.(interface {
			Measure(layout.Constraints) layout.Size
		}); ok {
			innerSize := measurable.Measure(innerConstraints)
			innerWidth = innerSize.Width
			innerHeight = innerSize.Height
		}
	}

	// Use explicit dimensions if set
	if inst.width > 0 {
		innerWidth = inst.width
	}
	if inst.height > 0 {
		innerHeight = inst.height
	}

	// Add border
	totalWidth := innerWidth + borderWidth*2
	totalHeight := innerHeight + borderWidth*2

	// Apply constraints using Constrain method
	totalWidth, totalHeight = constraints.Constrain(totalWidth, totalHeight)

	return layout.Size{Width: totalWidth, Height: totalHeight}
}

// =============================================================================
// PaintableInstance Interface
// =============================================================================

// Paint generates draw commands for the border.
// Border is rendered AROUND the content area, not inside it.
func (inst *Instance) Paint(x, y int) []paint.DrawCmd {
	if inst.borderStyle == BorderNone {
		return nil
	}

	// Get border dimensions
	width := inst.width
	height := inst.height
	if width == 0 {
		width = 10 // Default minimum
	}
	if height == 0 {
		height = 3 // Default minimum
	}

	// Calculate total size including border
	borderWidth := GetBorderWidth(inst.borderStyle)

	// Get border characters
	cornerTL, cornerTR, cornerBL, cornerBR, horizontal, vertical := inst.getBorderChars()

	var cmds []paint.DrawCmd
	borderStyle := style.Style{FG: inst.borderColor}

	// Top border
	topBorder := inst.buildTopBorder(cornerTL, cornerTR, horizontal, width)
	cmds = append(cmds, paint.NewTextCmd(x, y, topBorder, borderStyle))

	// Middle rows (vertical borders)
	contentY := y + borderWidth
	for i := 0; i < height; i++ {
		// Left border
		cmds = append(cmds, paint.NewTextCmd(x, contentY+i, string(vertical), borderStyle))
		// Right border
		cmds = append(cmds, paint.NewTextCmd(x+width+borderWidth, contentY+i, string(vertical), borderStyle))
	}

	// Bottom border
	bottomY := y + height + borderWidth
	bottomBorder := inst.buildBottomBorder(cornerBL, cornerBR, horizontal, width)
	cmds = append(cmds, paint.NewTextCmd(x, bottomY, bottomBorder, borderStyle))

	return cmds
}

// =============================================================================
// Bounds Management
// =============================================================================

// SetBounds sets the layout bounds.
func (inst *Instance) SetBounds(x, y, w, h int) {
	inst.bounds = [4]int{x, y, w, h}
}

// GetBounds returns the layout bounds.
func (inst *Instance) GetBounds() (x, y, w, h int) {
	return inst.bounds[0], inst.bounds[1], inst.bounds[2], inst.bounds[3]
}

// =============================================================================
// Border Configuration
// =============================================================================

// GetBorder returns the border configuration for the layout engine.
func (inst *Instance) GetBorder() layout.Border {
	return layout.Border{
		Style: inst.borderStyle,
		Width: GetBorderWidth(inst.borderStyle),
		Label: inst.borderLabel,
	}
}

// =============================================================================
// Style Management
// =============================================================================

// GetStyle returns the instance style.
func (inst *Instance) GetStyle() style.Style {
	return inst.instStyle
}

// SetStyle sets the instance style.
func (inst *Instance) SetStyle(s style.Style) {
	inst.instStyle = s
}

// ClearDirty clears the dirty flag.
func (inst *Instance) ClearDirty() {
	inst.dirty = false
}

// =============================================================================
// Border Character Helpers
// =============================================================================

// getBorderChars returns the characters for each border style.
func (inst *Instance) getBorderChars() (cornerTL, cornerTR, cornerBL, cornerBR, horizontal, vertical rune) {
	switch inst.borderStyle {
	case BorderDouble:
		return '╔', '╗', '╚', '╝', '═', '║'
	case BorderRounded:
		return '╭', '╮', '╰', '╯', '─', '│'
	case BorderDashed:
		return '+', '+', '+', '+', '-', '│'
	default: // BorderSingle
		return '┌', '┐', '└', '┘', '─', '│'
	}
}

// buildTopBorder builds the top border line with optional label.
func (inst *Instance) buildTopBorder(cornerTL, cornerTR, horizontal rune, contentWidth int) string {
	if inst.borderLabel == "" {
		// Simple top border: cornerTL + horizontal fill + cornerTR
		fill := strings.Repeat(string(horizontal), contentWidth)
		return string(cornerTL) + fill + string(cornerTR)
	}

	// Top border with label: "┌─ Label ─┐"
	label := inst.borderLabel
	labelWidth := len(label) + 2 // +1 for space on each side
	availableWidth := contentWidth
	if availableWidth < labelWidth {
		availableWidth = labelWidth
	}

	leftPadding := (availableWidth - labelWidth) / 2
	rightPadding := availableWidth - labelWidth - leftPadding

	left := string(cornerTL) + strings.Repeat(string(horizontal), leftPadding+1)
	right := strings.Repeat(string(horizontal), rightPadding+1) + string(cornerTR)

	return left + " " + label + " " + right
}

// buildBottomBorder builds the bottom border line.
func (inst *Instance) buildBottomBorder(cornerBL, cornerBR, horizontal rune, contentWidth int) string {
	fill := strings.Repeat(string(horizontal), contentWidth)
	return string(cornerBL) + fill + string(cornerBR)
}

// =============================================================================
// Helper Functions
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

func getBorderStyleProp(props rtui.Props, key string, def BorderStyle) BorderStyle {
	if v, ok := props[key]; ok {
		if s, ok := v.(BorderStyle); ok {
			return s
		}
	}
	return def
}

func getChildProp(props rtui.Props) rtui.VNode {
	if v, ok := props["child"].(rtui.VNode); ok {
		return v
	}
	return nil
}

func getStyleProp(props rtui.Props) style.Style {
	if v, ok := props["style"].(style.Style); ok {
		return v
	}
	return style.Style{}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
