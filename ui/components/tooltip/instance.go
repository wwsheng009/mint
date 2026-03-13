package tooltip

import (
	"github.com/wwsheng009/mint/ui/components/internal/proputil"
	"time"

	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Tooltip Instance - Runtime Entity
// =============================================================================

// Instance is the runtime entity for Tooltip components.
// It persists across renders and holds all state.
type Instance struct {
	// === Identification ===
	key string

	// === Props (from VNode, may change each render) ===
	text         string
	position     Position
	delay        time.Duration
	tooltipStyle style.Style

	// === Runtime State ===
	visible    bool
	showTimer  *time.Timer
	hoverTimer *time.Timer
	bounds     [4]int // x, y, w, h
	dirty      bool

	// === Content ===
	content rtui.VNode

	// === Tracking ===
	mouseOver    bool
	anchorBounds [4]int
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

// NewInstance creates a new TooltipInstance from props.
func NewInstance(props rtui.Props) *Instance {
	inst := &Instance{
		key:          proputil.GetString(props, "key", ""),
		text:         proputil.GetString(props, "text", ""),
		position:     getPositionProp(props, PositionAuto),
		delay:        getDurationProp(props, 500*time.Millisecond),
		tooltipStyle: proputil.GetStyle(props, "style", style.Style{}),
		visible:      false,
		dirty:        true,
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
	if inst.showTimer != nil {
		inst.showTimer.Stop()
	}
	if inst.hoverTimer != nil {
		inst.hoverTimer.Stop()
	}
}

// OnMount implements ComponentInstance.
func (inst *Instance) OnMount() {
	// Tooltip instances are managed by hit testing, not mount hooks
}

// OnUnmount implements ComponentInstance.
func (inst *Instance) OnUnmount() {
	inst.Destroy()
}

// SetProps implements ComponentInstance.
func (inst *Instance) SetProps(props rtui.Props) bool {
	oldText := inst.text
	oldPosition := inst.position
	oldDelay := inst.delay

	inst.text = proputil.GetString(props, "text", inst.text)
	inst.position = getPositionProp(props, inst.position)
	inst.delay = getDurationProp(props, inst.delay)
	inst.tooltipStyle = proputil.GetStyle(props, "style", style.Style{})

	changed := oldText != inst.text ||
		oldPosition != inst.position ||
		oldDelay != inst.delay

	if changed {
		inst.dirty = true
	}
	return changed
}

// GetProps implements ComponentInstance.
func (inst *Instance) GetProps() rtui.Props {
	return rtui.Props{
		propKey:      inst.key,
		propText:     inst.text,
		propPosition: inst.position,
		propDelay:    inst.delay,
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

// GetContext implements ComponentInstance (no hooks for Tooltip).
func (inst *Instance) GetContext() *rtui.ComponentContext {
	return nil
}

// =============================================================================
// PaintableInstance Interface
// =============================================================================

// Paint implements PaintableInstance.
func (inst *Instance) Paint(x, y int) []paint.DrawCmd {
	if !inst.visible || inst.text == "" {
		return nil
	}

	// Resolve tooltip style
	tooltipStyle := inst.resolveStyle()

	// Simple tooltip rendering with padding
	tooltipText := " " + inst.text + " "

	return []paint.DrawCmd{
		{
			X:     x,
			Y:     y,
			Text:  tooltipText,
			Style: tooltipStyle,
		},
	}
}

// resolveStyle resolves the visual style for the tooltip.
func (inst *Instance) resolveStyle() style.Style {
	s := inst.tooltipStyle

	// Apply default styling if not explicitly set
	if s.FG == "" && s.BG == "" {
		s = s.Foreground(theme.BG()).Background(theme.Primary()).Bold(true)
	}

	return s
}

// =============================================================================
// Public Methods
// =============================================================================

// Show displays the tooltip.
func (inst *Instance) Show() {
	inst.visible = true
	inst.dirty = true
}

// Hide hides the tooltip.
func (inst *Instance) Hide() {
	inst.visible = false
	inst.dirty = true
	if inst.showTimer != nil {
		inst.showTimer.Stop()
		inst.showTimer = nil
	}
	if inst.hoverTimer != nil {
		inst.hoverTimer.Stop()
		inst.hoverTimer = nil
	}
}

// SetAnchorBounds sets the bounds of the anchor element for positioning.
func (inst *Instance) SetAnchorBounds(x, y, w, h int) {
	inst.anchorBounds = [4]int{x, y, w, h}
}

// CalculatePosition calculates the tooltip position based on the anchor bounds.
// Returns (x, y) coordinates for the tooltip.
func (inst *Instance) CalculatePosition() (x, y int) {
	if len(inst.anchorBounds) != 4 {
		return 0, 0
	}

	anchorX, anchorY, anchorW, anchorH := inst.anchorBounds[0], inst.anchorBounds[1], inst.anchorBounds[2], inst.anchorBounds[3]
	tooltipWidth := paint.StringWidth(inst.text) + 2 // +2 for padding
	tooltipHeight := 1

	switch inst.position {
	case PositionTop:
		x = anchorX + anchorW/2 - tooltipWidth/2
		y = anchorY - tooltipHeight - 1
	case PositionBottom:
		x = anchorX + anchorW/2 - tooltipWidth/2
		y = anchorY + anchorH + 1
	case PositionLeft:
		x = anchorX - tooltipWidth - 1
		y = anchorY + anchorH/2 - tooltipHeight/2
	case PositionRight:
		x = anchorX + anchorW + 1
		y = anchorY + anchorH/2 - tooltipHeight/2
	case PositionAuto:
		// Default to top
		x = anchorX + anchorW/2 - tooltipWidth/2
		y = anchorY - tooltipHeight - 1
	}

	return x, y
}

// =============================================================================
// Measurable Interface
// =============================================================================

// Measure implements layout.Measurable interface.
func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
	if inst == nil || inst.text == "" {
		return layout.Size{}
	}

	width := paint.StringWidth(inst.text) + 2 // +2 for padding
	height := 1

	// Apply constraints
	width = constraints.ConstrainWidth(width)
	height = constraints.ConstrainHeight(height)

	return layout.Size{Width: width, Height: height}
}

// =============================================================================
// Prop Extraction Helpers
// =============================================================================

func getPositionProp(props rtui.Props, def Position) Position {
	if v, ok := props[propPosition]; ok {
		if pos, ok := v.(Position); ok {
			return pos
		}
	}
	return def
}

func getDurationProp(props rtui.Props, def time.Duration) time.Duration {
	if v, ok := props[propDelay]; ok {
		if d, ok := v.(time.Duration); ok {
			return d
		}
	}
	return def
}

func getToastTypeProp(props rtui.Props, def ToastType) ToastType {
	if v, ok := props["toastType"]; ok {
		if tt, ok := v.(ToastType); ok {
			return tt
		}
	}
	return def
}

