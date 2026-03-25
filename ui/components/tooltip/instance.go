package tooltip

import (
	"time"

	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/control"
	"github.com/wwsheng009/mint/ui/components/internal/overlayposition"
	"github.com/wwsheng009/mint/ui/components/internal/proputil"
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
	parent         rtui.ComponentInstance
	childInstances []rtui.ComponentInstance
	content        rtui.VNode

	// === Tracking ===
	mouseOver     bool
	triggerActive bool
	anchorBounds  [4]int
	viewportSize  [2]int
}

// Ensure Instance implements required interfaces
var (
	_ rtui.ComponentInstance       = (*Instance)(nil)
	_ rtui.PaintableInstance       = (*Instance)(nil)
	_ rtui.ActionHandlerInstance   = (*Instance)(nil)
	_ rtui.RuntimeChildrenProvider = (*Instance)(nil)
	_ rtui.TreeNode                = (*Instance)(nil)
	_ rtui.TreeContainer           = (*Instance)(nil)
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

func (inst *Instance) Parent() interface{} { return inst.parent }

func (inst *Instance) SetParent(parent rtui.ComponentInstance) { inst.parent = parent }

func (inst *Instance) Children() []rtui.ComponentInstance {
	return append([]rtui.ComponentInstance(nil), inst.childInstances...)
}

func (inst *Instance) AddChild(child rtui.ComponentInstance) {
	if child == nil {
		return
	}
	for index, existing := range inst.childInstances {
		if existing == child || existing.Key() == child.Key() {
			inst.childInstances[index] = child
			if setter, ok := child.(interface{ SetParent(rtui.ComponentInstance) }); ok {
				setter.SetParent(inst)
			}
			return
		}
	}
	inst.childInstances = append(inst.childInstances, child)
	if setter, ok := child.(interface{ SetParent(rtui.ComponentInstance) }); ok {
		setter.SetParent(inst)
	}
}

func (inst *Instance) RemoveChild(child rtui.ComponentInstance) {
	if child == nil {
		return
	}
	for index, existing := range inst.childInstances {
		if existing != child {
			continue
		}
		inst.childInstances = append(inst.childInstances[:index], inst.childInstances[index+1:]...)
		if setter, ok := child.(interface{ SetParent(rtui.ComponentInstance) }); ok {
			setter.SetParent(nil)
		}
		return
	}
}

func (inst *Instance) ClearChildren() {
	for _, child := range inst.childInstances {
		if setter, ok := child.(interface{ SetParent(rtui.ComponentInstance) }); ok {
			setter.SetParent(nil)
		}
	}
	inst.childInstances = inst.childInstances[:0]
}

// =============================================================================
// PaintableInstance Interface
// =============================================================================

// Paint implements PaintableInstance.
func (inst *Instance) Paint(x, y int) []paint.DrawCmd {
	return nil
}

// resolveStyle resolves the visual style for the tooltip.
func (inst *Instance) resolveStyle() style.Style {
	return resolveTooltipStyle(inst.tooltipStyle)
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

func (inst *Instance) RuntimeChildren() []rtui.VNode {
	inst.syncTriggerFromChildren()
	if !inst.visible || inst.text == "" {
		return nil
	}
	overlay := newOverlayVNode(inst.text, inst.position, inst.tooltipStyle, inst.anchorBounds, inst.viewportSize)
	if inst.key != "" {
		overlay.SetKey(inst.key + "-overlay")
	}
	return []rtui.VNode{overlay}
}

func (inst *Instance) SetBounds(x, y, w, h int) {
	next := [4]int{x, y, w, h}
	if inst.bounds == next && inst.anchorBounds == next {
		return
	}
	inst.bounds = next
	inst.anchorBounds = next
	if inst.visible {
		inst.dirty = true
	}
}

func (inst *Instance) HandleAction(act *action.Action) bool {
	if act == nil {
		return false
	}

	switch act.Type {
	case action.ActionMouseEnter, action.ActionHover:
		inst.mouseOver = true
		inst.syncTriggerActive(true)
		return true
	case action.ActionMouseLeave, action.ActionCancel:
		inst.mouseOver = false
		inst.syncTriggerActive(false)
		return true
	}
	return false
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
	result := overlayposition.Resolve(overlayposition.Config{
		Anchor: overlayposition.RectFromBounds(inst.anchorBounds),
		Overlay: overlayposition.Size{
			Width:  tooltipWidth,
			Height: tooltipHeight,
		},
		Viewport: overlayposition.Size{
			Width:  viewportWidth,
			Height: viewportHeight,
		},
		Candidates: inst.positionCandidates(),
		Gap:        1,
	})
	return result.X, result.Y
}

func (inst *Instance) positionCandidates() []overlayposition.Placement {
	switch inst.position {
	case PositionTop:
		return []overlayposition.Placement{
			overlayposition.PlacementTop,
			overlayposition.PlacementTopLeft,
			overlayposition.PlacementTopRight,
			overlayposition.PlacementBottom,
			overlayposition.PlacementBottomLeft,
			overlayposition.PlacementBottomRight,
		}
	case PositionTopLeft:
		return []overlayposition.Placement{
			overlayposition.PlacementTopLeft,
			overlayposition.PlacementTop,
			overlayposition.PlacementTopRight,
			overlayposition.PlacementBottomLeft,
			overlayposition.PlacementBottom,
		}
	case PositionTopRight:
		return []overlayposition.Placement{
			overlayposition.PlacementTopRight,
			overlayposition.PlacementTop,
			overlayposition.PlacementTopLeft,
			overlayposition.PlacementBottomRight,
			overlayposition.PlacementBottom,
		}
	case PositionBottomLeft:
		return []overlayposition.Placement{
			overlayposition.PlacementBottomLeft,
			overlayposition.PlacementBottom,
			overlayposition.PlacementBottomRight,
			overlayposition.PlacementTopLeft,
			overlayposition.PlacementTop,
		}
	case PositionBottomRight:
		return []overlayposition.Placement{
			overlayposition.PlacementBottomRight,
			overlayposition.PlacementBottom,
			overlayposition.PlacementBottomLeft,
			overlayposition.PlacementTopRight,
			overlayposition.PlacementTop,
		}
	case PositionLeftTop:
		return []overlayposition.Placement{
			overlayposition.PlacementLeftTop,
			overlayposition.PlacementLeft,
			overlayposition.PlacementLeftBottom,
			overlayposition.PlacementTop,
			overlayposition.PlacementBottom,
		}
	case PositionLeftBottom:
		return []overlayposition.Placement{
			overlayposition.PlacementLeftBottom,
			overlayposition.PlacementLeft,
			overlayposition.PlacementLeftTop,
			overlayposition.PlacementBottom,
			overlayposition.PlacementTop,
		}
	case PositionRightTop:
		return []overlayposition.Placement{
			overlayposition.PlacementRightTop,
			overlayposition.PlacementRight,
			overlayposition.PlacementRightBottom,
			overlayposition.PlacementTop,
			overlayposition.PlacementBottom,
		}
	case PositionRightBottom:
		return []overlayposition.Placement{
			overlayposition.PlacementRightBottom,
			overlayposition.PlacementRight,
			overlayposition.PlacementRightTop,
			overlayposition.PlacementBottom,
			overlayposition.PlacementTop,
		}
	case PositionBottom:
		return []overlayposition.Placement{
			overlayposition.PlacementBottom,
			overlayposition.PlacementBottomLeft,
			overlayposition.PlacementBottomRight,
			overlayposition.PlacementTop,
		}
	case PositionLeft:
		return []overlayposition.Placement{
			overlayposition.PlacementLeft,
			overlayposition.PlacementLeftTop,
			overlayposition.PlacementLeftBottom,
			overlayposition.PlacementRight,
		}
	case PositionRight:
		return []overlayposition.Placement{
			overlayposition.PlacementRight,
			overlayposition.PlacementRightTop,
			overlayposition.PlacementRightBottom,
			overlayposition.PlacementLeft,
		}
	case PositionAuto:
		return []overlayposition.Placement{
			overlayposition.PlacementTop,
			overlayposition.PlacementBottom,
			overlayposition.PlacementRight,
			overlayposition.PlacementLeft,
			overlayposition.PlacementTopLeft,
			overlayposition.PlacementTopRight,
			overlayposition.PlacementBottomLeft,
			overlayposition.PlacementBottomRight,
			overlayposition.PlacementRightTop,
			overlayposition.PlacementRightBottom,
			overlayposition.PlacementLeftTop,
			overlayposition.PlacementLeftBottom,
		}
	default:
		return []overlayposition.Placement{
			overlayposition.PlacementTop,
			overlayposition.PlacementTopLeft,
			overlayposition.PlacementTopRight,
			overlayposition.PlacementBottom,
		}
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

func (inst *Instance) scheduleShow() {
	if inst.visible {
		return
	}
	if inst.delay <= 0 {
		inst.Show()
		return
	}
	if inst.showTimer != nil {
		inst.showTimer.Stop()
	}
	inst.showTimer = time.AfterFunc(inst.delay, func() {
		if !inst.triggerActive {
			return
		}
		inst.Show()
	})
}

func (inst *Instance) syncTriggerFromChildren() {
	childHovered, childBounds, hasChildBounds := tooltipChildHoverState(inst.childInstances)
	if hasChildBounds && inst.anchorBounds != childBounds {
		inst.anchorBounds = childBounds
		if inst.visible {
			inst.dirty = true
		}
	}
	inst.syncTriggerActive(inst.mouseOver || childHovered)
}

func (inst *Instance) syncTriggerActive(active bool) {
	if active == inst.triggerActive {
		return
	}
	inst.triggerActive = active
	if active {
		inst.scheduleShow()
		return
	}
	inst.Hide()
}

func tooltipChildHoverState(children []rtui.ComponentInstance) (hovered bool, bounds [4]int, hasBounds bool) {
	for _, child := range children {
		if child == nil {
			continue
		}

		if !hasBounds {
			if childBounds, ok := tooltipInstanceBounds(child); ok {
				bounds = childBounds
				hasBounds = true
			}
		}

		if stateProvider, ok := child.(interface {
			GetState() *control.InteractionState
		}); ok {
			if state := stateProvider.GetState(); state != nil && state.Hovered {
				if childBounds, ok := tooltipInstanceBounds(child); ok {
					return true, childBounds, true
				}
				return true, bounds, hasBounds
			}
		}

		if node, ok := child.(rtui.TreeNode); ok {
			childHovered, childBounds, childHasBounds := tooltipChildHoverState(node.Children())
			if childHovered {
				return true, childBounds, childHasBounds
			}
			if !hasBounds && childHasBounds {
				bounds = childBounds
				hasBounds = true
			}
		}
	}
	return false, bounds, hasBounds
}

func tooltipInstanceBounds(inst rtui.ComponentInstance) ([4]int, bool) {
	reader, ok := inst.(interface{ GetBounds() (int, int, int, int) })
	if !ok {
		return [4]int{}, false
	}
	x, y, w, h := reader.GetBounds()
	if w <= 0 || h <= 0 {
		return [4]int{}, false
	}
	return [4]int{x, y, w, h}, true
}

func resolveTooltipStyle(s style.Style) style.Style {
	if s.FG == "" && s.BG == "" {
		return s.Foreground(theme.BG()).Background(theme.Primary()).Bold(true)
	}
	return s
}
