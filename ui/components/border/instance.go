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
	bounds            [4]int // x, y, w, h
	measuredChildSize layout.Size // cached child measurement
	dirty             bool
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

// measureChild measures a single child VNode by creating a temporary instance.
func (inst *Instance) measureChild(child rtui.VNode, constraints layout.Constraints) layout.Size {
	if child == nil {
		return layout.Size{}
	}

	// Try InstanceFactory -> CreateInstance -> Measure
	if factory, ok := child.(rtui.InstanceFactory); ok {
		tempInst := factory.CreateInstance()
		if measurable, ok := tempInst.(interface{ Measure(layout.Constraints) layout.Size }); ok {
			return measurable.Measure(constraints)
		}
	}

	// Try direct Measurable interface
	if measurable, ok := child.(interface{ Measure(layout.Constraints) layout.Size }); ok {
		return measurable.Measure(constraints)
	}

	// Fallback: estimate from content
	return inst.estimateChildSize(child, constraints)
}

// estimateChildSize estimates child size when Measure is not available.
func (inst *Instance) estimateChildSize(child rtui.VNode, constraints layout.Constraints) layout.Size {
	w := 10 // default width
	h := 1  // default height

	if props := child.Props(); props != nil {
		if pw := props.GetInt("width"); pw > 0 {
			w = pw
		}
		if ph := props.GetInt("height"); ph > 0 {
			h = ph
		}
		if content := props.GetString("content"); content != "" {
			w = len([]rune(content))
		}
	}

	// Handle LayoutNode (VStack/HStack/Wrap) with children
	if nodeWithChildren, ok := child.(interface{ Children() []rtui.VNode }); ok {
		children := nodeWithChildren.Children()
		if len(children) > 0 {
			// Check if this is a VStack (vertical layout) or HStack (horizontal layout)
			isVertical := true // default to VStack
			if tagger, ok := child.(interface{ Tag() string }); ok {
				tag := tagger.Tag()
				isVertical = (tag == "vstack" || tag == "VStack")
			}

			if isVertical {
				// VStack: width = max child width, height = sum of child heights
				maxWidth := 0
				totalHeight := 0
				for _, c := range children {
					cw, ch := inst.estimateSingleChildSize(c, constraints)
					if cw > maxWidth {
						maxWidth = cw
					}
					totalHeight += ch
				}
				w = maxWidth
				h = totalHeight
			} else {
				// HStack: width = sum of child widths, height = max child height
				totalWidth := 0
				maxHeight := 0
				for _, c := range children {
					cw, ch := inst.estimateSingleChildSize(c, constraints)
					totalWidth += cw
					if ch > maxHeight {
						maxHeight = ch
					}
				}
				w = totalWidth
				h = maxHeight
			}
		}
	}

	return layout.Size{
		Width:  constraints.ConstrainWidth(w),
		Height: constraints.ConstrainHeight(h),
	}
}

// estimateSingleChildSize estimates size for a single child node.
func (inst *Instance) estimateSingleChildSize(child rtui.VNode, constraints layout.Constraints) (w, h int) {
	w = 10 // default width
	h = 1  // default height

	if props := child.Props(); props != nil {
		if pw := props.GetInt("width"); pw > 0 {
			w = pw
		}
		if ph := props.GetInt("height"); ph > 0 {
			h = ph
		}
		if content := props.GetString("content"); content != "" {
			w = len([]rune(content))
		}
	}

	// Recursively handle LayoutNode children
	if nodeWithChildren, ok := child.(interface{ Children() []rtui.VNode }); ok {
		children := nodeWithChildren.Children()
		if len(children) > 0 {
			isVertical := true
			if tagger, ok := child.(interface{ Tag() string }); ok {
				tag := tagger.Tag()
				isVertical = (tag == "vstack" || tag == "VStack")
			}

			if isVertical {
				maxWidth := 0
				totalHeight := 0
				for _, c := range children {
					cw, ch := inst.estimateSingleChildSize(c, constraints)
					if cw > maxWidth {
						maxWidth = cw
					}
					totalHeight += ch
				}
				w = maxWidth
				h = totalHeight
			} else {
				totalWidth := 0
				maxHeight := 0
				for _, c := range children {
					cw, ch := inst.estimateSingleChildSize(c, constraints)
					totalWidth += cw
					if ch > maxHeight {
						maxHeight = ch
					}
				}
				w = totalWidth
				h = maxHeight
			}
		}
	}

	return w, h
}

// Measure calculates the natural size of the bordered container.
// If width or height is not explicitly set, it automatically measures the child.
// Border adds 2 * borderWidth to each dimension.
func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
	borderWidth := GetBorderWidth(inst.borderStyle)

	var innerWidth, innerHeight int

	// Use explicit dimensions if set
	if inst.width > 0 {
		innerWidth = inst.width
	}
	if inst.height > 0 {
		innerHeight = inst.height
	}

	// Auto-measure child if dimensions not explicitly set
	needMeasureWidth := inst.width == 0
	needMeasureHeight := inst.height == 0

	if (needMeasureWidth || needMeasureHeight) && inst.child != nil {
		// Calculate inner constraints (subtract border width)
		innerConstraints := layout.Constraints{
			MinWidth:  0,
			MaxWidth:  constraints.MaxWidth - 2*borderWidth,
			MinHeight: 0,
			MaxHeight: constraints.MaxHeight - 2*borderWidth,
		}
		if innerConstraints.MaxWidth < 0 {
			innerConstraints.MaxWidth = layout.MaxInt
		}
		if innerConstraints.MaxHeight < 0 {
			innerConstraints.MaxHeight = layout.MaxInt
		}

		// Measure child and cache result
		inst.measuredChildSize = inst.measureChild(inst.child, innerConstraints)

		if needMeasureWidth {
			innerWidth = inst.measuredChildSize.Width
		}
		if needMeasureHeight {
			innerHeight = inst.measuredChildSize.Height
		}
	}

	// Consider label width - label may be wider than child
	if inst.borderLabel != "" {
		// Label format: "┌─ Label ─┐" = corner + dash + space + label + space + dash + corner
		// Minimum: 1 + 1 + 1 + len(label) + 1 + 1 + 1 = len(label) + 6
		labelWidth := len(inst.borderLabel) + 6
		if labelWidth > innerWidth {
			innerWidth = labelWidth
		}
	}

	// Add border
	totalWidth := innerWidth + borderWidth*2
	totalHeight := innerHeight + borderWidth*2

	// Apply constraints
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

	borderWidth := GetBorderWidth(inst.borderStyle)

	// Determine content dimensions (width/height inside the border)
	// Priority: explicit props > bounds (from layout engine) > defaults
	var contentWidth, contentHeight int

	// 1. Use explicit dimensions if set
	if inst.width > 0 {
		contentWidth = inst.width
	}
	if inst.height > 0 {
		contentHeight = inst.height
	}

	// 2. If not set, try bounds from layout engine (total size minus border)
	if contentWidth == 0 || contentHeight == 0 {
		_, _, boundsW, boundsH := inst.GetBounds()
		if boundsW > 0 || boundsH > 0 {
			// Bounds contains total size, subtract border to get content size
			if contentWidth == 0 && boundsW > 0 {
				contentWidth = boundsW - 2*borderWidth
				if contentWidth < 0 {
					contentWidth = 0
				}
			}
			if contentHeight == 0 && boundsH > 0 {
				contentHeight = boundsH - 2*borderWidth
				if contentHeight < 0 {
					contentHeight = 0
				}
			}
		}
	}

	// 3. Fallback to defaults (when no bounds or explicit props)
	if contentWidth == 0 {
		contentWidth = 10
	}
	if contentHeight == 0 {
		contentHeight = 3
	}

	// Get border characters
	cornerTL, cornerTR, cornerBL, cornerBR, horizontal, vertical := inst.getBorderChars()

	var cmds []paint.DrawCmd
	borderStyle := style.Style{FG: inst.borderColor}

	// Top border
	topBorder := inst.buildTopBorder(cornerTL, cornerTR, horizontal, contentWidth)
	cmds = append(cmds, paint.NewTextCmd(x, y, topBorder, borderStyle))

	// Middle rows (vertical borders)
	contentY := y + borderWidth
	for i := 0; i < contentHeight; i++ {
		// Left border
		cmds = append(cmds, paint.NewTextCmd(x, contentY+i, string(vertical), borderStyle))
		// Right border
		cmds = append(cmds, paint.NewTextCmd(x+contentWidth+borderWidth, contentY+i, string(vertical), borderStyle))
	}

	// Bottom border
	bottomY := y + contentHeight + borderWidth
	bottomBorder := inst.buildBottomBorder(cornerBL, cornerBR, horizontal, contentWidth)
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
	labelSpace := len(label) + 2 // label + space on each side

	// If contentWidth is too small for label, truncate label
	if contentWidth < labelSpace {
		maxLabelLen := contentWidth - 2
		if maxLabelLen < 0 {
			maxLabelLen = 0
		}
		if maxLabelLen < len(label) {
			label = label[:maxLabelLen]
			labelSpace = len(label) + 2
		}
	}

	leftPadding := (contentWidth - labelSpace) / 2
	rightPadding := contentWidth - labelSpace - leftPadding

	left := string(cornerTL) + strings.Repeat(string(horizontal), leftPadding)
	right := strings.Repeat(string(horizontal), rightPadding) + string(cornerTR)

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
