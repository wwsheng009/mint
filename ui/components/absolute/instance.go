package absolute

import (
	"github.com/wwsheng009/mint/ui/components/internal/proputil"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Instance - Runtime Entity
// =============================================================================

// Instance is the runtime entity for Absolute components.
// It persists across renders and holds all state.
type Instance struct {
	// === Identification ===
	key string

	// === Child ===
	child rtui.VNode

	// === Position Props ===
	left   PositionValue
	top    PositionValue
	right  PositionValue
	bottom PositionValue
	anchor Anchor

	// === Sizing Props ===
	width  int
	height int
	zIndex int
	flex   int

	// === Style ===
	instStyle style.Style

	// === Runtime State ===
	bounds [4]int // x, y, w, h
	absX   int    // calculated absolute X
	absY   int    // calculated absolute Y
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

// NewInstance creates a new AbsoluteInstance from props.
func NewInstance(props rtui.Props) *Instance {
	inst := &Instance{
		key:       proputil.GetString(props, "key", ""),
		anchor:    AnchorTopLeft,
		width:     proputil.GetInt(props, "width", 0),
		height:    proputil.GetInt(props, "height", 0),
		zIndex:    proputil.GetInt(props, "zIndex", 0),
		flex:      proputil.GetInt(props, "flex", 0),
		instStyle: proputil.GetStyle(props, "style", style.Style{}),
		dirty:     true,
	}

	// Parse child
	if v, ok := props["child"].(rtui.VNode); ok {
		inst.child = v
	}

	// Parse positions
	if v, ok := props["left"].(PositionValue); ok {
		inst.left = v
	}
	if v, ok := props["top"].(PositionValue); ok {
		inst.top = v
	}
	if v, ok := props["right"].(PositionValue); ok {
		inst.right = v
	}
	if v, ok := props["bottom"].(PositionValue); ok {
		inst.bottom = v
	}
	if v, ok := props["anchor"].(Anchor); ok {
		inst.anchor = v
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
func (inst *Instance) OnMount() {}

// OnUnmount implements ComponentInstance.
func (inst *Instance) OnUnmount() {}

// SetProps implements ComponentInstance.
func (inst *Instance) SetProps(props rtui.Props) bool {
	oldWidth := inst.width
	oldHeight := inst.height
	oldChild := inst.child
	oldAnchor := inst.anchor

	inst.key = proputil.GetString(props, "key", inst.key)
	inst.width = proputil.GetInt(props, "width", inst.width)
	inst.height = proputil.GetInt(props, "height", inst.height)
	inst.zIndex = proputil.GetInt(props, "zIndex", inst.zIndex)
	inst.flex = proputil.GetInt(props, "flex", inst.flex)
	inst.instStyle = proputil.GetStyle(props, "style", style.Style{})

	if v, ok := props["child"].(rtui.VNode); ok {
		inst.child = v
	}
	if v, ok := props["left"].(PositionValue); ok {
		inst.left = v
	}
	if v, ok := props["top"].(PositionValue); ok {
		inst.top = v
	}
	if v, ok := props["right"].(PositionValue); ok {
		inst.right = v
	}
	if v, ok := props["bottom"].(PositionValue); ok {
		inst.bottom = v
	}
	if v, ok := props["anchor"].(Anchor); ok {
		inst.anchor = v
	}

	changed := oldWidth != inst.width ||
		oldHeight != inst.height ||
		oldChild != inst.child ||
		oldAnchor != inst.anchor

	if changed {
		inst.dirty = true
	}
	return changed
}

// GetProps implements ComponentInstance.
func (inst *Instance) GetProps() rtui.Props {
	return rtui.Props{
		"key":    inst.key,
		"child":  inst.child,
		"left":   inst.left,
		"top":    inst.top,
		"right":  inst.right,
		"bottom": inst.bottom,
		"anchor": inst.anchor,
		"width":  inst.width,
		"height": inst.height,
		"zIndex": inst.zIndex,
		"flex":   inst.flex,
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

// GetContext implements ComponentInstance (no hooks for Absolute).
func (inst *Instance) GetContext() *rtui.ComponentContext {
	return nil
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

// GetAbsolutePosition returns the calculated absolute position.
func (inst *Instance) GetAbsolutePosition() (x, y int) {
	return inst.absX, inst.absY
}

// SetAbsolutePosition sets the calculated absolute position.
func (inst *Instance) SetAbsolutePosition(x, y int) {
	inst.absX = x
	inst.absY = y
}

// CalculatePosition calculates the absolute position based on container size.
func (inst *Instance) CalculatePosition(containerWidth, containerHeight int) (int, int) {
	x := 0
	y := 0

	// Get child dimensions first (needed for Right/Bottom calculations)
	childWidth := inst.width
	if childWidth == 0 {
		childWidth = 20 // default
	}

	childHeight := inst.height
	if childHeight == 0 {
		childHeight = 1 // default
	}

	// Calculate X position
	if inst.left != nil {
		x = inst.left.Resolve(containerWidth)
	} else if inst.right != nil {
		rightPos := inst.right.Resolve(containerWidth)
		x = containerWidth - rightPos - childWidth
	}

	// Calculate Y position
	if inst.top != nil {
		y = inst.top.Resolve(containerHeight)
	} else if inst.bottom != nil {
		bottomPos := inst.bottom.Resolve(containerHeight)
		y = containerHeight - bottomPos - childHeight
	}

	// Adjust based on anchor
	switch inst.anchor {
	case AnchorTop, AnchorTopLeft:
		// No adjustment needed
	case AnchorTopRight:
		x = x - childWidth
	case AnchorLeft:
		// No adjustment needed
	case AnchorCenter:
		x = x - childWidth/2
		y = y - childHeight/2
	case AnchorRight:
		x = x - childWidth
	case AnchorBottom:
		y = y - childHeight
	case AnchorBottomLeft:
		y = y - childHeight
	case AnchorBottomRight:
		x = x - childWidth
		y = y - childHeight
	}

	return x, y
}

// =============================================================================
// Measurable Interface
// =============================================================================

// Measure implements layout.Measurable interface.
func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
	if inst == nil {
		return layout.Size{}
	}

	width := inst.width
	height := inst.height

	// If no explicit size, use defaults
	if width == 0 {
		width = 20
	}
	if height == 0 {
		height = 1
	}

	// Check explicit style dimensions
	if inst.instStyle.Width > 0 {
		width = inst.instStyle.Width
	}
	if inst.instStyle.Height > 0 {
		height = inst.instStyle.Height
	}

	// Apply constraints
	width = constraints.ConstrainWidth(width)
	height = constraints.ConstrainHeight(height)

	return layout.Size{Width: width, Height: height}
}

// =============================================================================
// PaintableInstance Interface
// =============================================================================

// Paint implements PaintableInstance.
// Absolute container doesn't have visual representation - child is painted by reconciler.
func (inst *Instance) Paint(x, y int) []paint.DrawCmd {
	// Pure layout container has no visual representation
	return nil
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
// Prop Extraction Helpers
// =============================================================================

// =============================================================================
// Layout Interfaces Implementation
// =============================================================================

// GetZIndex returns the z-index for this instance.
// Implements layout.Layered interface.
func (inst *Instance) GetZIndex() int {
	return inst.zIndex
}

// GetAbsoluteStyle returns the absolute positioning style.
// Implements layout.AbsoluteStyleProvider interface.
func (inst *Instance) GetAbsoluteStyle() *layout.AbsoluteStyle {
	return &layout.AbsoluteStyle{
		Left:   inst.left,
		Top:    inst.top,
		Right:  inst.right,
		Bottom: inst.bottom,
		Anchor: inst.anchor,
		Width:  inst.width,
		Height: inst.height,
		ZIndex: inst.zIndex,
	}
}

// GetPositionType returns absolute positioning type.
// Implements layout.Positionable interface.
func (inst *Instance) GetPositionType() layout.Position {
	// Build absolute position from component fields
	p := layout.NewAbsolutePosition()

	// Set left offset
	if inst.left != nil {
		var left int
		if absPos, ok := inst.left.(layout.AbsolutePos); ok {
			left = int(absPos)
		} else if relPos, ok := inst.left.(layout.RelativePos); ok {
			// Convert relative position to percentage representation
			// Negative values represent percentages (-2 to -101 = 2% to 101%)
			left = -int(relPos)
		}
		p.Left = &left
	}

	// Set top offset
	if inst.top != nil {
		var top int
		if absPos, ok := inst.top.(layout.AbsolutePos); ok {
			top = int(absPos)
		} else if relPos, ok := inst.top.(layout.RelativePos); ok {
			top = -int(relPos)
		}
		p.Top = &top
	}

	// Set right offset
	if inst.right != nil {
		var right int
		if absPos, ok := inst.right.(layout.AbsolutePos); ok {
			right = int(absPos)
		} else if relPos, ok := inst.right.(layout.RelativePos); ok {
			right = -int(relPos)
		}
		p.Right = &right
	}

	// Set bottom offset
	if inst.bottom != nil {
		var bottom int
		if absPos, ok := inst.bottom.(layout.AbsolutePos); ok {
			bottom = int(absPos)
		} else if relPos, ok := inst.bottom.(layout.RelativePos); ok {
			bottom = -int(relPos)
		}
		p.Bottom = &bottom
	}

	// Anchor is handled in GetAbsoluteStyle(), not in Position
	return p
}

