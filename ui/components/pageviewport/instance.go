package pageviewport

import (
	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	scrollutil "github.com/wwsheng009/mint/ui/components/internal/scroll"
)

// Instance stores PageViewport runtime state and layout metrics.
type Instance struct {
	key           string
	child         rtui.VNode
	width         int
	height        int
	scrollOffset  int
	controlled    bool
	showIndicator bool
	instStyle     style.Style
	dirty         bool
	bounds        [4]int
	metrics       [4]int // contentW, contentH, viewportW, viewportH
}

var (
	_ rtui.ComponentInstance     = (*Instance)(nil)
	_ rtui.PostPaintableInstance = (*Instance)(nil)
	_ interface {
		Measure(layout.Constraints) layout.Size
	} = (*Instance)(nil)
	_ rtui.ActionHandlerInstance                             = (*Instance)(nil)
	_ interface{ GetScrollViewport() layout.ScrollViewport } = (*Instance)(nil)
	_ interface {
		SetScrollViewportMetrics(contentWidth, contentHeight, viewportWidth, viewportHeight int)
	} = (*Instance)(nil)
)

// NewInstance creates a PageViewport instance.
func NewInstance(props rtui.Props) *Instance {
	inst := &Instance{showIndicator: true, dirty: true}
	inst.SetProps(props)
	return inst
}

func (inst *Instance) Key() string                        { return inst.key }
func (inst *Instance) SetKey(key string)                  { inst.key = key }
func (inst *Instance) Init(props rtui.Props)              { inst.SetProps(props) }
func (inst *Instance) Destroy()                           { inst.child = nil }
func (inst *Instance) OnMount()                           {}
func (inst *Instance) OnUnmount()                         {}
func (inst *Instance) MarkDirty()                         { inst.dirty = true }
func (inst *Instance) IsDirty() bool                      { return inst.dirty }
func (inst *Instance) GetContext() *rtui.ComponentContext { return nil }

func (inst *Instance) SetProps(props rtui.Props) bool {
	oldKey := inst.key
	oldChild := inst.child
	oldWidth := inst.width
	oldHeight := inst.height
	oldOffset := inst.scrollOffset
	oldControlled := inst.controlled
	oldIndicator := inst.showIndicator
	oldStyle := inst.instStyle

	if val, ok := props[propKey].(string); ok {
		inst.key = val
	}
	if val, ok := props[propChild].(rtui.VNode); ok {
		inst.child = val
	}
	if val, ok := props[propWidth].(int); ok {
		inst.width = val
	}
	if val, ok := props[propHeight].(int); ok {
		inst.height = val
	}
	if val, ok := props[propScrollOffset].(int); ok {
		inst.scrollOffset = val
		inst.controlled = true
	} else {
		inst.controlled = false
	}
	if inst.scrollOffset < 0 {
		inst.scrollOffset = 0
	}
	if val, ok := props[propShowIndicator].(bool); ok {
		inst.showIndicator = val
	}
	if val, ok := props[propStyle].(style.Style); ok {
		inst.instStyle = val
	}
	changed := oldKey != inst.key || oldChild != inst.child || oldWidth != inst.width ||
		oldHeight != inst.height || oldOffset != inst.scrollOffset ||
		oldControlled != inst.controlled || oldIndicator != inst.showIndicator || oldStyle != inst.instStyle
	if changed {
		inst.dirty = true
	}
	return changed
}

func (inst *Instance) GetProps() rtui.Props {
	props := rtui.Props{
		propKey:           inst.key,
		propChild:         inst.child,
		propWidth:         inst.width,
		propHeight:        inst.height,
		propShowIndicator: inst.showIndicator,
		propStyle:         inst.instStyle,
	}
	if inst.controlled {
		props[propScrollOffset] = inst.scrollOffset
	}
	return props
}

func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
	w := inst.width
	h := inst.height
	if w <= 0 && constraints.MaxWidth > 0 && constraints.MaxWidth < layout.MaxInt {
		w = constraints.MaxWidth
	}
	if h <= 0 && constraints.MaxHeight > 0 && constraints.MaxHeight < layout.MaxInt {
		h = constraints.MaxHeight
	}
	if w < 0 {
		w = 0
	}
	if h < 0 {
		h = 0
	}
	return layout.Size{Width: constraints.ConstrainWidth(w), Height: constraints.ConstrainHeight(h)}
}

func (inst *Instance) GetScrollViewport() layout.ScrollViewport {
	return layout.ScrollViewport{Enabled: true, Width: inst.width, Height: inst.height, ScrollOffset: inst.scrollOffset}
}

// PostPaint draws a non-interactive vertical scroll indicator after children have
// painted, keeping dense pages visibly scrollable without affecting hit testing.
func (inst *Instance) PostPaint(x, y int) []paint.DrawCmd {
	if !inst.showIndicator {
		return nil
	}
	width := inst.bounds[2]
	height := inst.bounds[3]
	if width <= 0 {
		width = inst.width
	}
	if height <= 0 {
		height = inst.viewportHeight()
	}
	if width <= 0 || height <= 0 {
		return nil
	}
	contentHeight := inst.contentHeight()
	if contentHeight <= height {
		return nil
	}
	maxOffset := contentHeight - height
	if maxOffset <= 0 {
		return nil
	}

	trackStyle := style.NewStyle().Foreground(style.BrightBlack)
	thumbStyle := style.NewStyle().Foreground(style.Cyan).Bold(true)
	cmds := make([]paint.DrawCmd, 0, height)
	drawX := x + width - 1
	thumbHeight := height * height / contentHeight
	if thumbHeight < 1 {
		thumbHeight = 1
	}
	if thumbHeight > height {
		thumbHeight = height
	}
	thumbTop := 0
	if maxThumbTop := height - thumbHeight; maxThumbTop > 0 {
		thumbTop = inst.scrollOffset * maxThumbTop / maxOffset
	}
	thumbBottom := thumbTop + thumbHeight

	for row := 0; row < height; row++ {
		text := "|"
		cmdStyle := trackStyle
		if row >= thumbTop && row < thumbBottom {
			text = "#"
			cmdStyle = thumbStyle
		}
		if row == 0 && inst.scrollOffset > 0 {
			text = "^"
			cmdStyle = thumbStyle
		}
		if row == height-1 && inst.scrollOffset < maxOffset {
			text = "v"
			cmdStyle = thumbStyle
		}
		cmds = append(cmds, paint.DrawCmd{X: drawX, Y: y + row, Text: text, Style: cmdStyle})
	}
	return cmds
}

func (inst *Instance) SetScrollViewportMetrics(contentWidth, contentHeight, viewportWidth, viewportHeight int) {
	inst.metrics = [4]int{contentWidth, contentHeight, viewportWidth, viewportHeight}
	if maxOffset := inst.maxOffset(); inst.scrollOffset > maxOffset {
		inst.scrollOffset = maxOffset
	}
}

func (inst *Instance) HandleAction(act *action.Action) bool {
	if act == nil || inst.controlled {
		return false
	}

	switch act.Type {
	case action.ActionScroll:
		if delta, ok := scrollutil.DeltaFromAction(act); ok {
			return inst.ScrollBy(delta)
		}
	case action.ActionNavigatePageUp:
		return inst.PageUp()
	case action.ActionNavigatePageDown:
		return inst.PageDown()
	case action.ActionNavigateHome:
		return inst.ScrollTo(0)
	case action.ActionNavigateEnd:
		return inst.ScrollTo(inst.maxOffset())
	}

	return false
}

func (inst *Instance) ScrollBy(delta int) bool {
	return inst.ScrollTo(inst.scrollOffset + delta)
}

func (inst *Instance) PageUp() bool {
	return inst.ScrollBy(-inst.viewportHeight())
}

func (inst *Instance) PageDown() bool {
	return inst.ScrollBy(inst.viewportHeight())
}

func (inst *Instance) ScrollTo(offset int) bool {
	if offset < 0 {
		offset = 0
	}
	if maxOffset := inst.maxOffset(); offset > maxOffset {
		offset = maxOffset
	}
	if offset == inst.scrollOffset {
		return false
	}
	inst.scrollOffset = offset
	inst.dirty = true
	return true
}

func (inst *Instance) viewportHeight() int {
	if inst.metrics[3] > 0 {
		return inst.metrics[3]
	}
	if inst.height > 0 {
		return inst.height
	}
	return 1
}

func (inst *Instance) contentHeight() int {
	if inst.metrics[1] > 0 {
		return inst.metrics[1]
	}
	return inst.height
}

func (inst *Instance) maxOffset() int {
	maxOffset := inst.contentHeight() - inst.viewportHeight()
	if maxOffset < 0 {
		return 0
	}
	return maxOffset
}

func (inst *Instance) SetBounds(x, y, w, h int) { inst.bounds = [4]int{x, y, w, h} }
func (inst *Instance) GetBounds() (x, y, w, h int) {
	return inst.bounds[0], inst.bounds[1], inst.bounds[2], inst.bounds[3]
}
