package popover

import (
	"fmt"
	"strings"

	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/internal/overlayposition"
	"github.com/wwsheng009/mint/ui/components/internal/proputil"
)

// Instance is the runtime entity for Popover components.
type Instance struct {
	key               string
	componentID       string
	title             string
	body              string
	placement         Placement
	trigger           TriggerMode
	open              bool
	openControlled    bool
	disabled          bool
	showArrow         bool
	gapRows           int
	maxWidth          int
	rootStyle         style.Style
	overlayStyle      style.Style
	borderStyle       style.Style
	shadowStyle       style.Style
	titleStyle        style.Style
	bodyStyle         style.Style
	changeIntent      intent.Intent
	changeIntentField intent.FieldIntent
	parent            rtui.ComponentInstance
	childInstances    []rtui.ComponentInstance
	intentEmitter     func(intent.Intent)
	bounds            [4]int
	viewportSize      [2]int
	dirty             bool
}

type overlayVNode struct {
	*rtui.ElementVNode
}

type overlayInstance struct {
	title        string
	body         string
	placement    Placement
	showArrow    bool
	gapRows      int
	maxWidth     int
	fillStyle    style.Style
	borderStyle  style.Style
	shadowStyle  style.Style
	titleStyle   style.Style
	bodyStyle    style.Style
	anchorBounds [4]int
	bounds       [4]int
	viewportSize [2]int
	dirty        bool
}

type popoverBox struct {
	x            int
	y            int
	width        int
	height       int
	contentW     int
	topBorder    string
	bottomBorder string
	bodyLines    []string
	titleLine    string
	divider      string
}

var (
	_ rtui.ComponentInstance       = (*Instance)(nil)
	_ rtui.RuntimeChildrenProvider = (*Instance)(nil)
	_ rtui.TreeNode                = (*Instance)(nil)
	_ rtui.TreeContainer           = (*Instance)(nil)
	_ rtui.ActionHandlerInstance   = (*Instance)(nil)
	_ intent.IntentHandler         = (*Instance)(nil)
	_ intent.TreeComponent         = (*Instance)(nil)

	_ rtui.ComponentInstance = (*overlayInstance)(nil)
	_ rtui.PaintableInstance = (*overlayInstance)(nil)
	_ interface {
		Measure(layout.Constraints) layout.Size
	} = (*overlayInstance)(nil)
)

// NewInstance creates a new Popover instance.
func NewInstance(props rtui.Props) *Instance {
	openControlled := proputil.GetBool(props, propOpenControlled, false)
	open := proputil.GetBool(props, propInitialOpen, false)
	if openControlled {
		open = proputil.GetBool(props, propOpen, false)
	}
	inst := &Instance{
		key:               proputil.GetString(props, propKey, ""),
		componentID:       proputil.GetString(props, propComponentID, ""),
		title:             proputil.GetString(props, propTitle, ""),
		body:              proputil.GetString(props, propBody, ""),
		placement:         getPlacementProp(props, PlacementAuto),
		trigger:           getTriggerProp(props, TriggerClick),
		open:              open,
		openControlled:    openControlled,
		disabled:          proputil.GetBool(props, propDisabled, false),
		showArrow:         proputil.GetBool(props, propShowArrow, true),
		gapRows:           proputil.GetInt(props, propGapRows, 1),
		maxWidth:          proputil.GetInt(props, propMaxWidth, 32),
		rootStyle:         proputil.GetStyle(props, propStyle, style.Style{}),
		overlayStyle:      proputil.GetStyle(props, propOverlayStyle, style.Style{}),
		borderStyle:       proputil.GetStyle(props, propBorderStyle, style.Style{}),
		shadowStyle:       proputil.GetStyle(props, propShadowStyle, style.Style{}),
		titleStyle:        proputil.GetStyle(props, propTitleStyle, style.Style{}),
		bodyStyle:         proputil.GetStyle(props, propBodyStyle, style.Style{}),
		changeIntent:      proputil.GetIntent(props, propChangeIntent, nil),
		changeIntentField: getFieldIntentProp(props, propChangeIntentField),
		dirty:             true,
	}
	inst.normalize()
	return inst
}

func (inst *Instance) Key() string { return inst.key }

func (inst *Instance) SetKey(key string) { inst.key = key }

func (inst *Instance) Init(props rtui.Props) { inst.SetProps(props) }

func (inst *Instance) Destroy() { popoverRegistryGlobal.unregister(inst) }

func (inst *Instance) OnMount() {
	popoverRegistryGlobal.register(inst)
	if inst.open {
		popoverRegistryGlobal.touch(inst)
	}
}

func (inst *Instance) OnUnmount() { popoverRegistryGlobal.unregister(inst) }

func (inst *Instance) SetProps(props rtui.Props) bool {
	old := *inst
	inst.key = proputil.GetString(props, propKey, inst.key)
	inst.componentID = proputil.GetString(props, propComponentID, inst.componentID)
	inst.title = proputil.GetString(props, propTitle, inst.title)
	inst.body = proputil.GetString(props, propBody, inst.body)
	inst.placement = getPlacementProp(props, inst.placement)
	inst.trigger = getTriggerProp(props, inst.trigger)
	nextControlled := proputil.GetBool(props, propOpenControlled, inst.openControlled)
	if nextControlled {
		inst.open = proputil.GetBool(props, propOpen, inst.open)
	} else if old.openControlled && !nextControlled {
		inst.open = proputil.GetBool(props, propInitialOpen, inst.open)
	}
	inst.openControlled = nextControlled
	inst.disabled = proputil.GetBool(props, propDisabled, inst.disabled)
	inst.showArrow = proputil.GetBool(props, propShowArrow, inst.showArrow)
	inst.gapRows = proputil.GetInt(props, propGapRows, inst.gapRows)
	inst.maxWidth = proputil.GetInt(props, propMaxWidth, inst.maxWidth)
	inst.rootStyle = proputil.GetStyle(props, propStyle, inst.rootStyle)
	inst.overlayStyle = proputil.GetStyle(props, propOverlayStyle, inst.overlayStyle)
	inst.borderStyle = proputil.GetStyle(props, propBorderStyle, inst.borderStyle)
	inst.shadowStyle = proputil.GetStyle(props, propShadowStyle, inst.shadowStyle)
	inst.titleStyle = proputil.GetStyle(props, propTitleStyle, inst.titleStyle)
	inst.bodyStyle = proputil.GetStyle(props, propBodyStyle, inst.bodyStyle)
	inst.changeIntent = proputil.GetIntent(props, propChangeIntent, inst.changeIntent)
	inst.changeIntentField = getFieldIntentProp(props, propChangeIntentField)
	inst.normalize()
	if !old.open && inst.open {
		popoverRegistryGlobal.touch(inst)
	}

	changed := old.key != inst.key ||
		old.componentID != inst.componentID ||
		old.title != inst.title ||
		old.body != inst.body ||
		old.placement != inst.placement ||
		old.trigger != inst.trigger ||
		old.open != inst.open ||
		old.openControlled != inst.openControlled ||
		old.disabled != inst.disabled ||
		old.showArrow != inst.showArrow ||
		old.gapRows != inst.gapRows ||
		old.maxWidth != inst.maxWidth ||
		old.rootStyle != inst.rootStyle ||
		old.overlayStyle != inst.overlayStyle ||
		old.borderStyle != inst.borderStyle ||
		old.shadowStyle != inst.shadowStyle ||
		old.titleStyle != inst.titleStyle ||
		old.bodyStyle != inst.bodyStyle ||
		old.changeIntent != inst.changeIntent ||
		old.changeIntentField != inst.changeIntentField
	if changed {
		inst.dirty = true
	}
	return changed
}

func (inst *Instance) GetProps() rtui.Props {
	return rtui.Props{
		propBody:              inst.body,
		propBorderStyle:       inst.borderStyle,
		propChangeIntent:      inst.changeIntent,
		propChangeIntentField: inst.changeIntentField,
		propComponentID:       inst.componentID,
		propDisabled:          inst.disabled,
		propGapRows:           inst.gapRows,
		propKey:               inst.key,
		propMaxWidth:          inst.maxWidth,
		propOpen:              inst.open,
		propOpenControlled:    inst.openControlled,
		propOverlayStyle:      inst.overlayStyle,
		propPlacement:         inst.placement,
		propShadowStyle:       inst.shadowStyle,
		propShowArrow:         inst.showArrow,
		propStyle:             inst.rootStyle,
		propTitle:             inst.title,
		propTitleStyle:        inst.titleStyle,
		propTrigger:           inst.trigger,
		propBodyStyle:         inst.bodyStyle,
	}
}

func (inst *Instance) MarkDirty() { inst.dirty = true }

func (inst *Instance) IsDirty() bool { return inst.dirty }

func (inst *Instance) GetContext() *rtui.ComponentContext { return nil }

func (inst *Instance) Parent() interface{} { return inst.parent }

func (inst *Instance) SetParent(parent rtui.ComponentInstance) { inst.parent = parent }

func (inst *Instance) Children() []rtui.ComponentInstance {
	return append([]rtui.ComponentInstance(nil), inst.childInstances...)
}

func (inst *Instance) AddChild(child rtui.ComponentInstance) {
	if child == nil {
		return
	}
	for i, existing := range inst.childInstances {
		if existing == child || existing.Key() == child.Key() {
			inst.childInstances[i] = child
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
	for i, existing := range inst.childInstances {
		if existing != child {
			continue
		}
		inst.childInstances = append(inst.childInstances[:i], inst.childInstances[i+1:]...)
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

func (inst *Instance) SetIntentEmitter(fn func(intent.Intent)) { inst.intentEmitter = fn }

func (inst *Instance) SetBounds(x, y, w, h int) {
	inst.bounds = [4]int{x, y, w, h}
}

func (inst *Instance) SetViewportSize(width, height int) {
	next := [2]int{width, height}
	if inst.viewportSize == next {
		return
	}
	inst.viewportSize = next
	if inst.open {
		inst.dirty = true
	}
}

func (inst *Instance) HandleAction(act *action.Action) bool {
	if act == nil || inst.disabled {
		return false
	}
	switch inst.trigger {
	case TriggerHover:
		switch act.Type {
		case action.ActionMouseEnter:
			return inst.setOpen(true, TriggerHover)
		case action.ActionMouseLeave, action.ActionCancel:
			return inst.setOpen(false, TriggerHover)
		}
	case TriggerManual:
		if act.Type == action.ActionCancel {
			return inst.setOpen(false, TriggerManual)
		}
	default:
		switch act.Type {
		case action.ActionClick, action.ActionEnter, action.ActionSelect:
			return inst.setOpen(!inst.open, TriggerClick)
		case action.ActionCancel:
			return inst.setOpen(false, TriggerClick)
		}
	}
	return false
}

func (inst *Instance) HandleIntent(i intent.Intent) bool {
	if !intent.ShouldHandleIntentWithID(inst.componentID, i) {
		return false
	}
	switch i.(type) {
	case PopoverToggleIntent:
		return inst.setOpen(!inst.open, inst.trigger)
	case PopoverOpenIntent:
		return inst.setOpen(true, inst.trigger)
	case PopoverCloseIntent:
		return inst.setOpen(false, inst.trigger)
	default:
		return false
	}
}

func (inst *Instance) RuntimeChildren() []rtui.VNode {
	if !inst.open || (strings.TrimSpace(inst.title) == "" && strings.TrimSpace(inst.body) == "") {
		return nil
	}
	overlay := newOverlayVNode(overlayProps{
		title:        inst.title,
		body:         inst.body,
		placement:    inst.placement,
		showArrow:    inst.showArrow,
		gapRows:      inst.gapRows,
		maxWidth:     inst.maxWidth,
		fillStyle:    inst.resolveOverlayStyle(),
		borderStyle:  inst.resolveBorderStyle(),
		shadowStyle:  inst.resolveShadowStyle(),
		titleStyle:   inst.resolveTitleStyle(),
		bodyStyle:    inst.resolveBodyStyle(),
		anchorBounds: inst.bounds,
		viewportSize: inst.viewportSize,
	})
	overlay.SetKey(inst.key + "-overlay")
	return []rtui.VNode{overlay}
}

func (inst *Instance) setOpen(next bool, trigger TriggerMode) bool {
	if inst.open == next && !inst.openControlled {
		return false
	}
	if !inst.openControlled {
		inst.open = next
		inst.dirty = true
		if next {
			popoverRegistryGlobal.touch(inst)
		}
	}
	inst.emitChange(next, trigger)
	return true
}

func (inst *Instance) requestClose(trigger TriggerMode) bool {
	if !inst.open {
		return false
	}
	return inst.setOpen(false, trigger)
}

func (inst *Instance) containsAnchorPoint(x, y int) bool {
	if inst.bounds[2] <= 0 || inst.bounds[3] <= 0 {
		return false
	}
	return x >= inst.bounds[0] && x < inst.bounds[0]+inst.bounds[2] &&
		y >= inst.bounds[1] && y < inst.bounds[1]+inst.bounds[3]
}

func (inst *Instance) containsOverlayPoint(x, y int) bool {
	box := computePopoverBox(inst.title, inst.body, inst.placement, inst.showArrow, inst.gapRows, inst.maxWidth, inst.bounds, inst.viewportSize)
	if box.width <= 0 || box.height <= 0 {
		return false
	}
	return x >= box.x && x < box.x+box.width &&
		y >= box.y && y < box.y+box.height
}

func (inst *Instance) emitChange(open bool, trigger TriggerMode) {
	if inst.intentEmitter == nil {
		return
	}
	inst.intentEmitter(PopoverChangeIntent{
		ComponentID: inst.componentID,
		Open:        open,
		Trigger:     trigger,
	})
	if inst.changeIntentField != nil {
		inst.intentEmitter(intent.FieldChangeIntent{
			Field: inst.changeIntentField.GetField(),
			Value: fmt.Sprintf("%t", open),
		})
	}
	if inst.changeIntent != nil {
		inst.intentEmitter(inst.changeIntent)
	}
}

func (inst *Instance) normalize() {
	if inst.gapRows < 0 {
		inst.gapRows = 0
	}
	if inst.maxWidth <= 0 {
		inst.maxWidth = 32
	}
}

func (inst *Instance) resolveOverlayStyle() style.Style {
	return style.NewStyle().
		Foreground(theme.Text()).
		Background(theme.Surface()).
		Merge(inst.overlayStyle)
}

func (inst *Instance) resolveBorderStyle() style.Style {
	base := style.NewStyle().Foreground(theme.Primary()).Background(theme.Surface()).Bold(true)
	if inst.borderStyle.IsEmpty() {
		return base
	}
	return base.Merge(inst.borderStyle)
}

func (inst *Instance) resolveShadowStyle() style.Style {
	base := style.NewStyle().Foreground(style.BrightBlack)
	if inst.shadowStyle.IsEmpty() {
		return base
	}
	return base.Merge(inst.shadowStyle)
}

func (inst *Instance) resolveTitleStyle() style.Style {
	base := style.NewStyle().Foreground(theme.Text()).Bold(true)
	if inst.titleStyle.IsEmpty() {
		return base
	}
	return base.Merge(inst.titleStyle)
}

func (inst *Instance) resolveBodyStyle() style.Style {
	base := style.NewStyle().Foreground(theme.Text())
	if inst.bodyStyle.IsEmpty() {
		return base
	}
	return base.Merge(inst.bodyStyle)
}

type overlayProps struct {
	title        string
	body         string
	placement    Placement
	showArrow    bool
	gapRows      int
	maxWidth     int
	fillStyle    style.Style
	borderStyle  style.Style
	shadowStyle  style.Style
	titleStyle   style.Style
	bodyStyle    style.Style
	anchorBounds [4]int
	viewportSize [2]int
}

func newOverlayVNode(model overlayProps) *overlayVNode {
	node := &overlayVNode{ElementVNode: rtui.NewElement("popover-overlay")}
	node.SetProp("title", model.title)
	node.SetProp("body", model.body)
	node.SetProp("placement", model.placement)
	node.SetProp("showArrow", model.showArrow)
	node.SetProp("gapRows", model.gapRows)
	node.SetProp("maxWidth", model.maxWidth)
	node.SetProp("overlayStyle", model.fillStyle)
	node.SetProp("borderStyle", model.borderStyle)
	node.SetProp("shadowStyle", model.shadowStyle)
	node.SetProp("titleStyle", model.titleStyle)
	node.SetProp("bodyStyle", model.bodyStyle)
	node.SetProp("anchorBounds", model.anchorBounds)
	node.SetProp("viewportSize", model.viewportSize)
	return node
}

func (v *overlayVNode) CreateInstance() rtui.ComponentInstance {
	return newOverlayInstance(v.Props())
}

func (v *overlayVNode) GetLayer() rtui.Layer { return rtui.LayerOverlay }

func (v *overlayVNode) SetLayer(l rtui.Layer) rtui.VNode { return v }

func newOverlayInstance(props rtui.Props) *overlayInstance {
	inst := &overlayInstance{
		title:        proputil.GetString(props, "title", ""),
		body:         proputil.GetString(props, "body", ""),
		placement:    getPlacementProp(props, PlacementAuto),
		showArrow:    proputil.GetBool(props, "showArrow", true),
		gapRows:      proputil.GetInt(props, "gapRows", 1),
		maxWidth:     proputil.GetInt(props, "maxWidth", 32),
		fillStyle:    proputil.GetStyle(props, "overlayStyle", style.Style{}),
		borderStyle:  proputil.GetStyle(props, "borderStyle", style.Style{}),
		shadowStyle:  proputil.GetStyle(props, "shadowStyle", style.Style{}),
		titleStyle:   proputil.GetStyle(props, "titleStyle", style.Style{}),
		bodyStyle:    proputil.GetStyle(props, "bodyStyle", style.Style{}),
		anchorBounds: getAnchorBoundsProp(props),
		viewportSize: getViewportSizeProp(props),
		dirty:        true,
	}
	if inst.maxWidth <= 0 {
		inst.maxWidth = 32
	}
	return inst
}

func (inst *overlayInstance) Key() string                        { return "" }
func (inst *overlayInstance) SetKey(key string)                  {}
func (inst *overlayInstance) Init(props rtui.Props)              { inst.SetProps(props) }
func (inst *overlayInstance) Destroy()                           {}
func (inst *overlayInstance) OnMount()                           {}
func (inst *overlayInstance) OnUnmount()                         {}
func (inst *overlayInstance) MarkDirty()                         { inst.dirty = true }
func (inst *overlayInstance) IsDirty() bool                      { return inst.dirty }
func (inst *overlayInstance) GetContext() *rtui.ComponentContext { return nil }

func (inst *overlayInstance) SetProps(props rtui.Props) bool {
	oldTitle := inst.title
	oldBody := inst.body
	oldPlacement := inst.placement
	oldShowArrow := inst.showArrow
	oldGapRows := inst.gapRows
	oldMaxWidth := inst.maxWidth
	oldFill := inst.fillStyle
	oldBorder := inst.borderStyle
	oldShadow := inst.shadowStyle
	oldTitleStyle := inst.titleStyle
	oldBodyStyle := inst.bodyStyle
	oldAnchor := inst.anchorBounds
	oldViewport := inst.viewportSize

	inst.title = proputil.GetString(props, "title", inst.title)
	inst.body = proputil.GetString(props, "body", inst.body)
	inst.placement = getPlacementProp(props, inst.placement)
	inst.showArrow = proputil.GetBool(props, "showArrow", inst.showArrow)
	inst.gapRows = proputil.GetInt(props, "gapRows", inst.gapRows)
	inst.maxWidth = proputil.GetInt(props, "maxWidth", inst.maxWidth)
	inst.fillStyle = proputil.GetStyle(props, "overlayStyle", inst.fillStyle)
	inst.borderStyle = proputil.GetStyle(props, "borderStyle", inst.borderStyle)
	inst.shadowStyle = proputil.GetStyle(props, "shadowStyle", inst.shadowStyle)
	inst.titleStyle = proputil.GetStyle(props, "titleStyle", inst.titleStyle)
	inst.bodyStyle = proputil.GetStyle(props, "bodyStyle", inst.bodyStyle)
	inst.anchorBounds = getAnchorBoundsPropWithDefault(props, inst.anchorBounds)
	inst.viewportSize = getViewportSizePropWithDefault(props, inst.viewportSize)
	changed := oldTitle != inst.title ||
		oldBody != inst.body ||
		oldPlacement != inst.placement ||
		oldShowArrow != inst.showArrow ||
		oldGapRows != inst.gapRows ||
		oldMaxWidth != inst.maxWidth ||
		oldFill != inst.fillStyle ||
		oldBorder != inst.borderStyle ||
		oldShadow != inst.shadowStyle ||
		oldTitleStyle != inst.titleStyle ||
		oldBodyStyle != inst.bodyStyle ||
		oldAnchor != inst.anchorBounds ||
		oldViewport != inst.viewportSize
	if changed {
		inst.dirty = true
	}
	return changed
}

func (inst *overlayInstance) GetProps() rtui.Props {
	return rtui.Props{
		"title":        inst.title,
		"body":         inst.body,
		"placement":    inst.placement,
		"showArrow":    inst.showArrow,
		"gapRows":      inst.gapRows,
		"maxWidth":     inst.maxWidth,
		"overlayStyle": inst.fillStyle,
		"borderStyle":  inst.borderStyle,
		"shadowStyle":  inst.shadowStyle,
		"titleStyle":   inst.titleStyle,
		"bodyStyle":    inst.bodyStyle,
		"anchorBounds": inst.anchorBounds,
		"viewportSize": inst.viewportSize,
	}
}

func (inst *overlayInstance) Measure(constraints layout.Constraints) layout.Size {
	return layout.Size{}
}

func (inst *overlayInstance) SetBounds(x, y, w, h int) {
	inst.bounds = [4]int{x, y, w, h}
}

func (inst *overlayInstance) SetViewportSize(width, height int) {
	next := [2]int{width, height}
	if inst.viewportSize == next {
		return
	}
	inst.viewportSize = next
	inst.dirty = true
}

func (inst *overlayInstance) Paint(x, y int) []paint.DrawCmd {
	box := inst.computeBox()
	if box.width <= 0 || box.height <= 0 {
		return nil
	}

	cmds := make([]paint.DrawCmd, 0, box.height*4)
	for row := 0; row < box.height; row++ {
		cmds = append(cmds, paint.DrawCmd{
			X:     box.x + 1,
			Y:     box.y + row + 1,
			Text:  strings.Repeat("░", box.width),
			Style: inst.shadowStyle,
		})
	}

	cmds = append(cmds,
		paint.DrawCmd{X: box.x, Y: box.y, Text: box.topBorder, Style: inst.borderStyle},
		paint.DrawCmd{X: box.x, Y: box.y + box.height - 1, Text: box.bottomBorder, Style: inst.borderStyle},
	)

	currentY := box.y + 1
	if box.titleLine != "" {
		cmds = append(cmds, inst.paintInteriorLine(box.x, currentY, box.contentW, box.titleLine, inst.titleStyle)...)
		currentY++
		if box.divider != "" {
			cmds = append(cmds, paint.DrawCmd{X: box.x, Y: currentY, Text: box.divider, Style: inst.borderStyle})
			currentY++
		}
	}
	for _, line := range box.bodyLines {
		cmds = append(cmds, inst.paintInteriorLine(box.x, currentY, box.contentW, line, inst.bodyStyle)...)
		currentY++
	}
	return cmds
}

func (inst *overlayInstance) paintInteriorLine(x, y, contentWidth int, line string, contentStyle style.Style) []paint.DrawCmd {
	padded := padDisplayWidth(line, contentWidth)
	return []paint.DrawCmd{
		{X: x, Y: y, Text: "│", Style: inst.borderStyle},
		{X: x + 1, Y: y, Text: " " + padded + " ", Style: inst.fillStyle},
		{X: x + contentWidth + 3, Y: y, Text: "│", Style: inst.borderStyle},
		{X: x + 2, Y: y, Text: padded, Style: contentStyle.Merge(inst.fillStyle)},
	}
}

func (inst *overlayInstance) computeBox() popoverBox {
	return computePopoverBox(inst.title, inst.body, inst.placement, inst.showArrow, inst.gapRows, inst.maxWidth, inst.anchorBounds, inst.viewportSize)
}

func computePopoverBox(title, body string, placement Placement, showArrow bool, gapRows, maxWidth int, anchorBounds [4]int, viewportSize [2]int) popoverBox {
	title = strings.TrimSpace(title)
	bodyLines := wrapTextLines(body, maxWidth)
	if len(bodyLines) == 0 {
		bodyLines = []string{""}
	}
	contentWidth := 1
	for _, line := range bodyLines {
		if w := paint.StringWidth(line); w > contentWidth {
			contentWidth = w
		}
	}
	titleLine := ""
	if title != "" {
		titleLine = truncateByDisplayWidth(title, maxInt(contentWidth, maxWidth))
		if w := paint.StringWidth(titleLine); w > contentWidth {
			contentWidth = w
		}
	}

	width := contentWidth + 4
	height := len(bodyLines) + 2
	divider := ""
	if titleLine != "" {
		height += 2
		divider = "├" + strings.Repeat("─", contentWidth+2) + "┤"
	}

	result := resolvePopoverLayout(anchorBounds, placement, width, height, gapRows, viewportSize)
	resolvedPlacement := popoverPlacementFromOverlay(result.Placement)
	x, y := result.X, result.Y
	arrowOffset := resolvePopoverArrowOffset(anchorBounds, x, width)
	topBorder := "┌" + strings.Repeat("─", contentWidth+2) + "┐"
	bottomBorder := "└" + strings.Repeat("─", contentWidth+2) + "┘"
	if showArrow {
		switch resolvedPlacement {
		case PlacementTop, PlacementTopLeft, PlacementTopRight:
			bottomBorder = replaceRune(bottomBorder, arrowOffset, '▼')
		default:
			topBorder = replaceRune(topBorder, arrowOffset, '▲')
		}
	}

	return popoverBox{
		x:            x,
		y:            y,
		width:        width,
		height:       height,
		contentW:     contentWidth,
		topBorder:    topBorder,
		bottomBorder: bottomBorder,
		bodyLines:    bodyLines,
		titleLine:    titleLine,
		divider:      divider,
	}
}

func resolvePopoverPlacement(placement Placement, gapRows int, anchorBounds [4]int, height int, viewportSize [2]int) Placement {
	if placement != PlacementAuto {
		return placement
	}
	return popoverPlacementFromOverlay(overlayposition.PreferredVerticalPlacement(
		overlayposition.RectFromBounds(anchorBounds),
		height,
		gapRows,
		viewportSize[1],
		overlayposition.VerticalAutoPreferTop,
	))
}

func resolvePopoverLayout(anchorBounds [4]int, placement Placement, width, height, gapRows int, viewportSize [2]int) overlayposition.Result {
	preferred := resolvePopoverPlacement(placement, gapRows, anchorBounds, height, viewportSize)
	return overlayposition.Resolve(overlayposition.Config{
		Anchor: overlayposition.RectFromBounds(anchorBounds),
		Overlay: overlayposition.Size{
			Width:  width,
			Height: height,
		},
		Viewport: overlayposition.Size{
			Width:  viewportSize[0],
			Height: viewportSize[1],
		},
		Candidates: resolvePopoverCandidates(preferred),
		Gap:        gapRows,
	})
}

func resolvePopoverCandidates(placement Placement) []overlayposition.Placement {
	return overlayposition.VerticalPlacementCandidates(popoverPlacementToOverlay(placement))
}

func popoverPlacementFromOverlay(placement overlayposition.Placement) Placement {
	switch placement {
	case overlayposition.PlacementTopLeft:
		return PlacementTopLeft
	case overlayposition.PlacementTopRight:
		return PlacementTopRight
	case overlayposition.PlacementBottom:
		return PlacementBottom
	case overlayposition.PlacementBottomLeft:
		return PlacementBottomLeft
	case overlayposition.PlacementBottomRight:
		return PlacementBottomRight
	default:
		return PlacementTop
	}
}

func popoverPlacementToOverlay(placement Placement) overlayposition.Placement {
	switch placement {
	case PlacementTopLeft:
		return overlayposition.PlacementTopLeft
	case PlacementTopRight:
		return overlayposition.PlacementTopRight
	case PlacementBottom:
		return overlayposition.PlacementBottom
	case PlacementBottomLeft:
		return overlayposition.PlacementBottomLeft
	case PlacementBottomRight:
		return overlayposition.PlacementBottomRight
	default:
		return overlayposition.PlacementTop
	}
}

func resolvePopoverArrowOffset(anchorBounds [4]int, boxX, width int) int {
	return overlayposition.PointerX(overlayposition.RectFromBounds(anchorBounds), boxX, width) - boxX
}

func getPlacementProp(props rtui.Props, def Placement) Placement {
	if value, ok := props[propPlacement]; ok {
		if placement, ok := value.(Placement); ok {
			return placement
		}
	}
	if value, ok := props["placement"]; ok {
		if placement, ok := value.(Placement); ok {
			return placement
		}
	}
	return def
}

func getTriggerProp(props rtui.Props, def TriggerMode) TriggerMode {
	if value, ok := props[propTrigger]; ok {
		if trigger, ok := value.(TriggerMode); ok {
			return trigger
		}
	}
	return def
}

func getFieldIntentProp(props rtui.Props, key string) intent.FieldIntent {
	if value, ok := props[key]; ok {
		if result, ok := value.(intent.FieldIntent); ok {
			return result
		}
	}
	return nil
}

func getAnchorBoundsProp(props rtui.Props) [4]int {
	if value, ok := props["anchorBounds"].([4]int); ok {
		return value
	}
	return [4]int{}
}

func getAnchorBoundsPropWithDefault(props rtui.Props, def [4]int) [4]int {
	if value, ok := props["anchorBounds"].([4]int); ok {
		return value
	}
	return def
}

func getViewportSizeProp(props rtui.Props) [2]int {
	if value, ok := props["viewportSize"].([2]int); ok {
		return value
	}
	return [2]int{}
}

func getViewportSizePropWithDefault(props rtui.Props, def [2]int) [2]int {
	if value, ok := props["viewportSize"].([2]int); ok {
		return value
	}
	return def
}

func wrapTextLines(text string, maxWidth int) []string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	parts := strings.Split(text, "\n")
	lines := make([]string, 0, len(parts))
	for _, part := range parts {
		if maxWidth <= 0 || paint.StringWidth(part) <= maxWidth {
			lines = append(lines, part)
			continue
		}
		lines = append(lines, wrapSingleLine(part, maxWidth)...)
	}
	if len(lines) == 0 {
		return nil
	}
	return lines
}

func wrapSingleLine(text string, maxWidth int) []string {
	if maxWidth <= 0 {
		return []string{text}
	}
	var lines []string
	var builder strings.Builder
	width := 0
	for _, r := range text {
		rw := paint.RuneWidth(r)
		if width+rw > maxWidth && builder.Len() > 0 {
			lines = append(lines, builder.String())
			builder.Reset()
			width = 0
		}
		builder.WriteRune(r)
		width += rw
	}
	if builder.Len() > 0 {
		lines = append(lines, builder.String())
	}
	if len(lines) == 0 {
		return []string{""}
	}
	return lines
}

func truncateByDisplayWidth(text string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	var builder strings.Builder
	width := 0
	for _, r := range text {
		rw := paint.RuneWidth(r)
		if width+rw > maxWidth {
			break
		}
		builder.WriteRune(r)
		width += rw
	}
	return builder.String()
}

func padDisplayWidth(content string, width int) string {
	content = truncateByDisplayWidth(content, width)
	padding := width - paint.StringWidth(content)
	if padding <= 0 {
		return content
	}
	return content + strings.Repeat(" ", padding)
}

func replaceRune(content string, index int, replacement rune) string {
	runes := []rune(content)
	if index <= 0 || index >= len(runes)-1 {
		return content
	}
	runes[index] = replacement
	return string(runes)
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
