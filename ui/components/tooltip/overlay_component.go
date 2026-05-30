package tooltip

import (
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/internal/proputil"
)

const propAnchorBounds = "anchorBounds"
const propViewportSize = "viewportSize"

type overlayVNode struct {
	*rtui.ElementVNode
	layer rtui.Layer
}

func newOverlayVNode(text string, position Position, tooltipStyle style.Style, layer rtui.Layer, anchorBounds [4]int, viewportSize [2]int) *overlayVNode {
	if !layer.IsValid() {
		layer = rtui.LayerTooltip
	}
	node := &overlayVNode{ElementVNode: rtui.NewElement("tooltip-overlay"), layer: layer}
	node.SetStyle(tooltipStyle)
	node.SetProp(propText, text)
	node.SetProp(propPosition, position)
	node.SetProp(propStyle, tooltipStyle)
	node.SetProp(propLayer, layer)
	node.SetProp(propAnchorBounds, anchorBounds)
	node.SetProp(propViewportSize, viewportSize)
	return node
}

func (v *overlayVNode) CreateInstance() rtui.ComponentInstance {
	props := v.Props().Clone()
	props[propStyle] = v.Style()
	return newOverlayInstance(props)
}

func (v *overlayVNode) GetLayer() rtui.Layer {
	return v.layer
}

func (v *overlayVNode) SetLayer(l rtui.Layer) rtui.VNode {
	if l.IsValid() {
		v.layer = l
		v.SetProp(propLayer, l)
	}
	return v
}

type overlayInstance struct {
	text         string
	position     Position
	tooltipStyle style.Style
	anchorBounds [4]int
	viewportSize [2]int
	bounds       [4]int
	dirty        bool
}

func newOverlayInstance(props rtui.Props) *overlayInstance {
	inst := &overlayInstance{}
	inst.SetProps(props)
	return inst
}

func (inst *overlayInstance) Key() string                        { return "" }
func (inst *overlayInstance) SetKey(key string)                  {}
func (inst *overlayInstance) Init(props rtui.Props)              { inst.SetProps(props) }
func (inst *overlayInstance) Destroy()                           {}
func (inst *overlayInstance) OnMount()                           {}
func (inst *overlayInstance) OnUnmount()                         {}
func (inst *overlayInstance) GetContext() *rtui.ComponentContext { return nil }
func (inst *overlayInstance) MarkDirty()                         { inst.dirty = true }
func (inst *overlayInstance) IsDirty() bool                      { return inst.dirty }

func (inst *overlayInstance) SetProps(props rtui.Props) bool {
	oldText := inst.text
	oldPosition := inst.position
	oldStyle := inst.tooltipStyle
	oldAnchor := inst.anchorBounds
	oldViewport := inst.viewportSize

	inst.text = proputil.GetString(props, propText, inst.text)
	inst.position = getPositionProp(props, inst.position)
	inst.tooltipStyle = proputil.GetStyle(props, propStyle, inst.tooltipStyle)
	inst.anchorBounds = getBoundsProp(props, propAnchorBounds, inst.anchorBounds)
	inst.viewportSize = getViewportSizeProp(props, propViewportSize, inst.viewportSize)

	changed := oldText != inst.text ||
		oldPosition != inst.position ||
		oldStyle != inst.tooltipStyle ||
		oldAnchor != inst.anchorBounds ||
		oldViewport != inst.viewportSize
	if changed {
		inst.dirty = true
	}
	return changed
}

func (inst *overlayInstance) GetProps() rtui.Props {
	return rtui.Props{
		propText:         inst.text,
		propPosition:     inst.position,
		propStyle:        inst.tooltipStyle,
		propAnchorBounds: inst.anchorBounds,
		propViewportSize: inst.viewportSize,
	}
}

func (inst *overlayInstance) Paint(x, y int) []paint.DrawCmd {
	if inst.text == "" {
		return nil
	}

	viewportWidth := inst.viewportSize[0]
	viewportHeight := inst.viewportSize[1]
	if viewportWidth <= 0 {
		viewportWidth = inst.bounds[2]
	}
	if viewportHeight <= 0 {
		viewportHeight = inst.bounds[3]
	}

	measure := &Instance{
		text:         inst.text,
		position:     inst.position,
		anchorBounds: inst.anchorBounds,
		viewportSize: [2]int{viewportWidth, viewportHeight},
	}
	tooltipWidth := paint.StringWidth(inst.text) + 2
	tooltipHeight := 1
	tooltipX, tooltipY := measure.calculatePositionWithViewport(tooltipWidth, tooltipHeight, viewportWidth, viewportHeight)

	return []paint.DrawCmd{{
		X:     tooltipX,
		Y:     tooltipY,
		Text:  " " + inst.text + " ",
		Style: resolveTooltipStyle(inst.tooltipStyle),
	}}
}

func (inst *overlayInstance) Measure(constraints layout.Constraints) layout.Size {
	return layout.Size{}
}

func (inst *overlayInstance) SetBounds(x, y, w, h int) {
	inst.bounds = [4]int{x, y, w, h}
}

func getBoundsProp(props rtui.Props, key string, fallback [4]int) [4]int {
	if value, ok := props[key].([4]int); ok {
		return value
	}
	return fallback
}

func getViewportSizeProp(props rtui.Props, key string, fallback [2]int) [2]int {
	if value, ok := props[key].([2]int); ok {
		return value
	}
	return fallback
}
