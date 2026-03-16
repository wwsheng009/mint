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
	viewportSize [2]int
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

// SetViewportSize sets the viewport size used for placement fallback.
func (inst *Instance) SetViewportSize(width, height int) {
	inst.viewportSize = [2]int{width, height}
}

// CalculatePosition calculates the tooltip position based on the anchor bounds.
// Returns (x, y) coordinates for the tooltip.
func (inst *Instance) CalculatePosition() (x, y int) {
	tooltipWidth := paint.StringWidth(inst.text) + 2 // +2 for padding
	tooltipHeight := 1
	return inst.calculatePositionWithViewport(tooltipWidth, tooltipHeight, inst.viewportSize[0], inst.viewportSize[1])
}

func (inst *Instance) calculatePositionWithViewport(tooltipWidth, tooltipHeight, viewportWidth, viewportHeight int) (x, y int) {
	if len(inst.anchorBounds) != 4 {
		return 0, 0
	}

	candidates := inst.positionCandidates()
	if viewportWidth > 0 && viewportHeight > 0 {
		for _, pos := range candidates {
			candidateX, candidateY := inst.positionCoordinates(pos, tooltipWidth, tooltipHeight)
			if fitsViewport(candidateX, candidateY, tooltipWidth, tooltipHeight, viewportWidth, viewportHeight) {
				return candidateX, candidateY
			}
		}
	}

	x, y = inst.positionCoordinates(candidates[0], tooltipWidth, tooltipHeight)
	if viewportWidth > 0 && viewportHeight > 0 {
		x = clamp(x, 0, maxInt(0, viewportWidth-tooltipWidth))
		y = clamp(y, 0, maxInt(0, viewportHeight-tooltipHeight))
	}
	return x, y
}

func (inst *Instance) positionCandidates() []Position {
	switch inst.position {
	case PositionTopLeft:
		return []Position{PositionTopLeft, PositionTop, PositionTopRight, PositionBottomLeft, PositionBottom}
	case PositionTopRight:
		return []Position{PositionTopRight, PositionTop, PositionTopLeft, PositionBottomRight, PositionBottom}
	case PositionBottomLeft:
		return []Position{PositionBottomLeft, PositionBottom, PositionBottomRight, PositionTopLeft, PositionTop}
	case PositionBottomRight:
		return []Position{PositionBottomRight, PositionBottom, PositionBottomLeft, PositionTopRight, PositionTop}
	case PositionLeftTop:
		return []Position{PositionLeftTop, PositionLeft, PositionLeftBottom, PositionTop, PositionBottom}
	case PositionLeftBottom:
		return []Position{PositionLeftBottom, PositionLeft, PositionLeftTop, PositionBottom, PositionTop}
	case PositionRightTop:
		return []Position{PositionRightTop, PositionRight, PositionRightBottom, PositionTop, PositionBottom}
	case PositionRightBottom:
		return []Position{PositionRightBottom, PositionRight, PositionRightTop, PositionBottom, PositionTop}
	case PositionBottom:
		return []Position{PositionBottom, PositionBottomLeft, PositionBottomRight, PositionTop}
	case PositionLeft:
		return []Position{PositionLeft, PositionLeftTop, PositionLeftBottom, PositionRight}
	case PositionRight:
		return []Position{PositionRight, PositionRightTop, PositionRightBottom, PositionLeft}
	case PositionAuto:
		return []Position{
			PositionTop,
			PositionBottom,
			PositionRight,
			PositionLeft,
			PositionTopLeft,
			PositionTopRight,
			PositionBottomLeft,
			PositionBottomRight,
			PositionRightTop,
			PositionRightBottom,
			PositionLeftTop,
			PositionLeftBottom,
		}
	default:
		return []Position{PositionTop, PositionTopLeft, PositionTopRight, PositionBottom}
	}
}

func (inst *Instance) positionCoordinates(position Position, tooltipWidth, tooltipHeight int) (x, y int) {
	anchorX, anchorY, anchorW, anchorH := inst.anchorBounds[0], inst.anchorBounds[1], inst.anchorBounds[2], inst.anchorBounds[3]

	centerX := anchorX + anchorW/2 - tooltipWidth/2
	topY := anchorY - tooltipHeight - 1
	bottomY := anchorY + anchorH + 1
	leftX := anchorX - tooltipWidth - 1
	rightX := anchorX + anchorW + 1
	centerY := anchorY + anchorH/2 - tooltipHeight/2
	topAlignedY := anchorY
	bottomAlignedY := anchorY + anchorH - tooltipHeight
	leftAlignedX := anchorX
	rightAlignedX := anchorX + anchorW - tooltipWidth

	switch position {
	case PositionTop:
		return centerX, topY
	case PositionTopLeft:
		return leftAlignedX, topY
	case PositionTopRight:
		return rightAlignedX, topY
	case PositionBottom:
		return centerX, bottomY
	case PositionBottomLeft:
		return leftAlignedX, bottomY
	case PositionBottomRight:
		return rightAlignedX, bottomY
	case PositionLeft:
		return leftX, centerY
	case PositionLeftTop:
		return leftX, topAlignedY
	case PositionLeftBottom:
		return leftX, bottomAlignedY
	case PositionRight:
		return rightX, centerY
	case PositionRightTop:
		return rightX, topAlignedY
	case PositionRightBottom:
		return rightX, bottomAlignedY
	case PositionAuto:
		return centerX, topY
	default:
		return centerX, topY
	}
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

func fitsViewport(x, y, width, height, viewportWidth, viewportHeight int) bool {
	if viewportWidth <= 0 || viewportHeight <= 0 {
		return true
	}
	return x >= 0 && y >= 0 && x+width <= viewportWidth && y+height <= viewportHeight
}

func clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
