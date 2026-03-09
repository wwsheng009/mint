package menu

import (
	"strings"
	"time"

	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/layout"
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rttypes "github.com/wwsheng009/mint/runtime/types"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	scrollutil "github.com/wwsheng009/mint/ui/components/internal/scroll"
)

const (
	defaultPopupMinWidth = 18
	popupLeftPadding     = 1
	popupRightPadding    = 1
)

type ThemeableModel struct {
	Model
	Theme Theme
	Style style.Style
}

type Builder struct {
	model ThemeableModel
}

func NewBuilder() *Builder {
	return &Builder{
		model: ThemeableModel{
			Model: Model{
				Variant:          VariantPopup,
				Open:             true,
				Layer:            rtui.LayerOverlay,
				ShowShortcuts:    true,
				ShowCheckMarks:   true,
				ShowIcons:        true,
				Scrollable:       true,
				CloseOnOutside:   true,
				CloseOnEscape:    true,
				Typeahead:        true,
				TypeaheadTimeout: 750 * time.Millisecond,
				ActiveIndex:      -1,
				SelectedIndex:    -1,
			},
			Theme: DefaultTheme(),
		},
	}
}

func NewMenuBar(items []MenuItem) *Builder {
	return NewBuilder().Variant(VariantMenuBar).Items(items).Layer(rtui.LayerBase)
}

func NewPopup(items []MenuItem) *Builder {
	return NewBuilder().Variant(VariantPopup).Items(items).Layer(rtui.LayerOverlay)
}

func NewContextMenu(items []MenuItem) *Builder {
	return NewBuilder().Variant(VariantContext).Items(items).Layer(rtui.LayerOverlay)
}

func (b *Builder) Key(key string) *Builder         { return b.SetID(key) }
func (b *Builder) SetID(id string) *Builder        { b.model.ID = id; return b }
func (b *Builder) ID(id string) *Builder           { return b.SetID(id) }
func (b *Builder) ComponentID(id string) *Builder  { b.model.ComponentID = id; return b }
func (b *Builder) Variant(v Variant) *Builder      { b.model.Variant = v; return b }
func (b *Builder) Title(title string) *Builder     { b.model.Title = title; return b }
func (b *Builder) Theme(theme Theme) *Builder      { b.model.Theme = theme; return b }
func (b *Builder) Style(s style.Style) *Builder    { b.model.Style = s; return b }
func (b *Builder) Layer(layer rtui.Layer) *Builder { b.model.Layer = layer; return b }
func (b *Builder) PathPrefix(path ...int) *Builder {
	b.model.PathPrefix = append([]int(nil), path...)
	return b
}
func (b *Builder) PortalRoot(root string) *Builder { b.model.PortalRoot = root; return b }
func (b *Builder) AnchorTo(anchorID string, anchor rttypes.Anchor) *Builder {
	b.model.AnchorID = anchorID
	b.model.Anchor = anchor
	return b
}
func (b *Builder) PortalPosition(position rttypes.PositionType) *Builder {
	b.model.PortalPosition = position
	return b
}
func (b *Builder) PortalPriority(priority int) *Builder {
	b.model.PortalPriority = priority
	return b
}
func (b *Builder) PortalOffset(offsetX, offsetY int) *Builder {
	b.model.PortalOffsetX = offsetX
	b.model.PortalOffsetY = offsetY
	return b
}
func (b *Builder) Open(open bool) *Builder        { b.model.Open = open; return b }
func (b *Builder) Placement(p Placement) *Builder { b.model.Placement = p; return b }
func (b *Builder) ActivePath(path ...int) *Builder {
	b.model.ActivePath = append([]int(nil), path...)
	return b
}
func (b *Builder) ActiveIndex(index int) *Builder   { b.model.ActiveIndex = index; return b }
func (b *Builder) SelectedIndex(index int) *Builder { b.model.SelectedIndex = index; return b }
func (b *Builder) ScrollOffset(offset int) *Builder {
	b.model.ScrollOffset = clampNonNegative(offset)
	return b
}
func (b *Builder) MaxWidth(width int) *Builder { b.model.MaxWidth = clampNonNegative(width); return b }
func (b *Builder) MaxHeight(height int) *Builder {
	b.model.MaxHeight = clampNonNegative(height)
	return b
}
func (b *Builder) MinWidth(width int) *Builder         { b.model.MinWidth = clampNonNegative(width); return b }
func (b *Builder) ShowShortcuts(show bool) *Builder    { b.model.ShowShortcuts = show; return b }
func (b *Builder) ShowDescriptions(show bool) *Builder { b.model.ShowDescriptions = show; return b }
func (b *Builder) ShowCheckMarks(show bool) *Builder   { b.model.ShowCheckMarks = show; return b }
func (b *Builder) ShowIcons(show bool) *Builder        { b.model.ShowIcons = show; return b }
func (b *Builder) Scrollable(scrollable bool) *Builder { b.model.Scrollable = scrollable; return b }
func (b *Builder) CloseOnOutside(close bool) *Builder  { b.model.CloseOnOutside = close; return b }
func (b *Builder) CloseOnEscape(close bool) *Builder   { b.model.CloseOnEscape = close; return b }
func (b *Builder) Typeahead(enabled bool) *Builder     { b.model.Typeahead = enabled; return b }
func (b *Builder) TypeaheadTimeout(timeout time.Duration) *Builder {
	b.model.TypeaheadTimeout = timeout
	return b
}
func (b *Builder) RegisterShortcuts(register bool) *Builder {
	b.model.RegisterShortcuts = register
	return b
}
func (b *Builder) Items(items []MenuItem) *Builder { b.model.Items = cloneItems(items); return b }
func (b *Builder) AddItem(item MenuItem) *Builder {
	b.model.Items = append(b.model.Items, item.Normalize())
	return b
}
func (b *Builder) Shortcuts() []ShortcutBinding { return CollectShortcuts(b.model.Items) }

func (b *Builder) BuildModel() ThemeableModel {
	model := b.model
	model.Items = cloneItems(model.Items)
	model.PathPrefix = append([]int(nil), model.PathPrefix...)
	if isZeroTheme(model.Theme) {
		model.Theme = DefaultTheme()
	}
	if model.MinWidth <= 0 {
		model.MinWidth = defaultPopupMinWidth
	}
	if model.TypeaheadTimeout <= 0 {
		model.TypeaheadTimeout = 750 * time.Millisecond
	}
	if model.Variant != VariantMenuBar && model.PortalRoot == "" {
		model.PortalRoot = rtui.DefaultOverlayPortalRootID
	}
	if model.Variant == VariantMenuBar && model.Layer == rtui.LayerOverlay {
		model.Layer = rtui.LayerBase
	}
	if model.AnchorID != "" && model.PortalPosition == rttypes.PositionRelative {
		model.PortalPosition = rttypes.PositionAbsolute
	}
	return model
}

func (b *Builder) Build() rtui.VNode {
	model := b.BuildModel()
	switch model.Variant {
	case VariantMenuBar:
		return newBarVNode(model)
	default:
		surface := newPopupVNode(clearPortalModel(model))
		if model.PortalRoot == "" {
			return surface
		}
		return newPopupPortalVNode(model, surface)
	}
}

func newPopupPortalVNode(model ThemeableModel, child rtui.VNode) rtui.VNode {
	portal := rtui.NewElement("box")
	if model.ID != "" {
		portal.SetKey(model.ID + "-portal")
	}
	portal.SetProps(rtui.Props{
		"position": "absolute",
		"left":     0,
		"top":      0,
		"width":    1,
		"height":   1,
		"_layer":   model.Layer,
	})
	applyPortalProps(portal, model.Model)
	portal.SetChildren([]rtui.VNode{child})
	return portal
}

func clearPortalModel(model ThemeableModel) ThemeableModel {
	model.PortalRoot = ""
	model.AnchorID = ""
	model.Anchor = rttypes.AnchorTopLeft
	model.PortalPosition = rttypes.PositionRelative
	model.PortalPriority = 0
	model.PortalOffsetX = 0
	model.PortalOffsetY = 0
	return model
}

type barVNode struct{ *rtui.ElementVNode }
type popupVNode struct{ *rtui.ElementVNode }

func newBarVNode(model ThemeableModel) *barVNode {
	node := &barVNode{ElementVNode: rtui.NewElement("menu-bar")}
	if model.ID != "" {
		node.SetID(model.ID)
		node.SetKey(model.ID)
	}
	applyPortalProps(node.ElementVNode, model.Model)
	node.SetProp("model", model.Model)
	node.SetProp("theme", model.Theme)
	node.SetProp("style", model.Style)
	node.SetProp("layer", model.Layer)
	return node
}

func newPopupVNode(model ThemeableModel) *popupVNode {
	tag := "menu-popup"
	if model.Variant == VariantContext {
		tag = "context-menu"
	}
	node := &popupVNode{ElementVNode: rtui.NewElement(tag)}
	if model.ID != "" {
		node.SetID(model.ID)
		node.SetKey(model.ID)
	}
	node.SetProp("model", model.Model)
	node.SetProp("theme", model.Theme)
	node.SetProp("style", model.Style)
	node.SetProp("layer", model.Layer)
	return node
}

func (v *barVNode) SetProps(p rtui.Props) rtui.VNode {
	v.ElementVNode.SetProps(v.ElementVNode.Props().Merge(p))
	return v
}

func (v *popupVNode) SetProps(p rtui.Props) rtui.VNode {
	v.ElementVNode.SetProps(v.ElementVNode.Props().Merge(p))
	return v
}

func (v *barVNode) GetLayer() rtui.Layer {
	if layer, ok := v.Props()["layer"].(rtui.Layer); ok {
		return layer
	}
	return rtui.LayerBase
}

func (v *barVNode) SetLayer(l rtui.Layer) rtui.VNode {
	v.SetProp("layer", l)
	return v
}

func (v *popupVNode) GetLayer() rtui.Layer {
	if layer, ok := v.Props()["layer"].(rtui.Layer); ok {
		return layer
	}
	return rtui.LayerOverlay
}

func (v *popupVNode) SetLayer(l rtui.Layer) rtui.VNode {
	v.SetProp("layer", l)
	return v
}

func (v *barVNode) CreateInstance() rtui.ComponentInstance {
	props := v.Props().Clone()
	props["key"] = v.Key()
	props["style"] = getNodeStyle(v)
	return newBarInstance(props)
}

func (v *popupVNode) CreateInstance() rtui.ComponentInstance {
	props := v.Props().Clone()
	props["key"] = v.Key()
	props["style"] = getNodeStyle(v)
	return newPopupInstance(props)
}

type barInstance struct {
	key           string
	model         Model
	theme         Theme
	rootStyle     style.Style
	activeIndex   int
	focused       bool
	bounds        [4]int
	dirty         bool
	intentEmitter func(intent.Intent)
}

type popupInstance struct {
	key           string
	model         Model
	theme         Theme
	rootStyle     style.Style
	selectedIndex int
	scrollOffset  int
	submenuPath   []int
	submenuScroll []int
	typeahead     string
	typeaheadAt   time.Time
	open          bool
	registered    bool
	focused       bool
	bounds        [4]int
	dirty         bool
	intentEmitter func(intent.Intent)
}

type barSegment struct {
	index int
	x     int
	width int
	text  string
}

type popupMetrics struct {
	visibleIndices []int
	innerWidth     int
	contentWidth   int
	markWidth      int
	iconWidth      int
	shortcutWidth  int
	arrowWidth     int
	viewportRows   int
	surfaceWidth   int
	surfaceHeight  int
	shadowWidth    int
	shadowHeight   int
}

type popupSurface struct {
	depth        int
	parentPath   []int
	items        []MenuItem
	selectedIdx  int
	scrollOffset int
	x            int
	y            int
	metrics      popupMetrics
}

var (
	_ rtui.ComponentInstance     = (*barInstance)(nil)
	_ rtui.PaintableInstance     = (*barInstance)(nil)
	_ rtui.ActionHandlerInstance = (*barInstance)(nil)
	_ rtui.FocusableInstance     = (*barInstance)(nil)
	_ interface {
		Measure(layout.Constraints) layout.Size
	} = (*barInstance)(nil)

	_ rtui.ComponentInstance     = (*popupInstance)(nil)
	_ rtui.PaintableInstance     = (*popupInstance)(nil)
	_ rtui.ActionHandlerInstance = (*popupInstance)(nil)
	_ rtui.FocusableInstance     = (*popupInstance)(nil)
	_ interface {
		Measure(layout.Constraints) layout.Size
	} = (*popupInstance)(nil)
)

func newBarInstance(props rtui.Props) *barInstance {
	inst := &barInstance{dirty: true}
	inst.SetProps(props)
	if inst.activeIndex < 0 {
		inst.activeIndex = FirstSelectableIndex(inst.model.Items)
	}
	return inst
}

func newPopupInstance(props rtui.Props) *popupInstance {
	inst := &popupInstance{dirty: true}
	inst.SetProps(props)
	if inst.selectedIndex < 0 {
		inst.selectedIndex = FirstSelectableIndex(inst.model.Items)
	}
	inst.ensureSelectionVisible()
	return inst
}

func (inst *barInstance) Key() string                        { return inst.key }
func (inst *barInstance) SetKey(key string)                  { inst.key = key }
func (inst *barInstance) Init(props rtui.Props)              { inst.SetProps(props) }
func (inst *barInstance) Destroy()                           { menuRegistryGlobal.unregisterBar(inst) }
func (inst *barInstance) OnMount()                           { menuRegistryGlobal.registerBar(inst); inst.dirty = true }
func (inst *barInstance) OnUnmount()                         { menuRegistryGlobal.unregisterBar(inst) }
func (inst *barInstance) MarkDirty()                         { inst.dirty = true }
func (inst *barInstance) IsDirty() bool                      { return inst.dirty }
func (inst *barInstance) GetContext() *rtui.ComponentContext { return nil }

func (inst *popupInstance) Key() string                        { return inst.key }
func (inst *popupInstance) SetKey(key string)                  { inst.key = key }
func (inst *popupInstance) Init(props rtui.Props)              { inst.SetProps(props) }
func (inst *popupInstance) Destroy()                           { inst.unregister() }
func (inst *popupInstance) OnMount()                           { inst.syncRegistration(); inst.dirty = true }
func (inst *popupInstance) OnUnmount()                         { inst.unregister() }
func (inst *popupInstance) MarkDirty()                         { inst.dirty = true }
func (inst *popupInstance) IsDirty() bool                      { return inst.dirty }
func (inst *popupInstance) GetContext() *rtui.ComponentContext { return nil }

func (inst *barInstance) SetProps(props rtui.Props) bool {
	inst.key = getStringProp(props, "key", inst.key)
	inst.model = getModelProp(props, inst.model)
	inst.model.Items = cloneItems(inst.model.Items)
	inst.theme = getThemeProp(props, inst.theme)
	if isZeroTheme(inst.theme) {
		inst.theme = DefaultTheme()
	}
	inst.rootStyle = getStyleProp(props, "style")
	if inst.model.ActiveIndex >= 0 {
		inst.activeIndex = clampIndex(inst.model.ActiveIndex, len(inst.model.Items))
	}
	if inst.activeIndex < 0 {
		inst.activeIndex = FirstSelectableIndex(inst.model.Items)
	}
	inst.dirty = true
	return true
}

func (inst *popupInstance) SetProps(props rtui.Props) bool {
	inst.key = getStringProp(props, "key", inst.key)
	inst.model = getModelProp(props, inst.model)
	inst.model.Items = cloneItems(inst.model.Items)
	localActivePath := trimPathPrefix(inst.model.ActivePath, inst.model.PathPrefix)
	inst.theme = getThemeProp(props, inst.theme)
	if isZeroTheme(inst.theme) {
		inst.theme = DefaultTheme()
	}
	inst.rootStyle = getStyleProp(props, "style")
	inst.open = inst.model.Open
	if len(localActivePath) > 0 {
		inst.selectedIndex = clampIndex(localActivePath[0], len(inst.model.Items))
		inst.submenuPath = append([]int(nil), localActivePath[1:]...)
		inst.submenuScroll = ensureLength(inst.submenuScroll, len(inst.submenuPath))
	} else if inst.model.SelectedIndex >= 0 {
		inst.selectedIndex = clampIndex(inst.model.SelectedIndex, len(inst.model.Items))
	}
	if inst.selectedIndex < 0 {
		inst.selectedIndex = FirstSelectableIndex(inst.model.Items)
	}
	inst.scrollOffset = clampNonNegative(inst.model.ScrollOffset)
	inst.normalizeCascadeState()
	inst.syncRegistration()
	inst.ensureSelectionVisible()
	inst.dirty = true
	return true
}

func (inst *barInstance) GetProps() rtui.Props {
	return rtui.Props{"key": inst.key, "model": inst.model, "theme": inst.theme, "style": inst.rootStyle}
}

func (inst *popupInstance) GetProps() rtui.Props {
	model := inst.model
	model.Open = inst.open
	model.ActivePath = inst.prefixedPath(inst.activePath())
	model.SelectedIndex = inst.selectedIndex
	model.ScrollOffset = inst.scrollOffset
	return rtui.Props{"key": inst.key, "model": model, "theme": inst.theme, "style": inst.rootStyle}
}

func (inst *barInstance) Measure(constraints layout.Constraints) layout.Size {
	segments := inst.barSegments()
	width := 0
	for _, segment := range segments {
		width = max(width, segment.x+segment.width)
	}
	if width == 0 {
		width = 1
	}
	width = applyWidthConstraints(width, constraints)
	return layout.Size{Width: width, Height: 1}
}

func (inst *popupInstance) Measure(constraints layout.Constraints) layout.Size {
	if !inst.open {
		return layout.Size{}
	}
	surfaces := inst.popupSurfaces()
	width := 1
	height := 1
	for _, surface := range surfaces {
		width = max(width, surface.x+surface.metrics.surfaceWidth+surface.metrics.shadowWidth)
		height = max(height, surface.y+surface.metrics.surfaceHeight+surface.metrics.shadowHeight)
	}
	width = applyWidthConstraints(width, constraints)
	height = applyHeightConstraints(height, constraints)
	return layout.Size{Width: width, Height: height}
}

func (inst *barInstance) Paint(x, y int) []paint.DrawCmd {
	segments := inst.barSegments()
	cmds := make([]paint.DrawCmd, 0, len(segments))
	baseStyle := inst.theme.BarStyle.Merge(inst.rootStyle)
	for _, segment := range segments {
		item := inst.model.Items[segment.index].Normalize()
		segmentStyle := baseStyle
		if item.Disabled {
			segmentStyle = segmentStyle.Merge(inst.theme.ItemDisabledStyle)
		} else if segment.index == inst.activeIndex {
			segmentStyle = segmentStyle.Merge(inst.theme.BarActiveStyle)
			if inst.focused {
				segmentStyle = segmentStyle.Merge(inst.theme.ItemFocusStyle)
			}
		}
		cmds = append(cmds, paint.DrawCmd{X: x + segment.x, Y: y, Text: segment.text, Style: segmentStyle})
	}
	return cmds
}

func (inst *popupInstance) Paint(x, y int) []paint.DrawCmd {
	if !inst.open {
		return nil
	}
	surfaces := inst.popupSurfaces()
	if len(surfaces) == 0 {
		return nil
	}
	cmds := make([]paint.DrawCmd, 0, len(surfaces)*32)
	for _, surface := range surfaces {
		cmds = append(cmds, inst.paintSurface(surface, x, y)...)
	}
	return cmds
}

func (inst *barInstance) HandleAction(act *action.Action) bool {
	if act == nil {
		return false
	}
	switch act.Type {
	case action.ActionNavigateLeft, action.ActionNavigatePrev:
		previous := inst.activeIndex
		inst.activeIndex = PrevSelectableIndex(inst.model.Items, inst.activeIndex)
		if inst.activeIndex != previous {
			inst.EmitIntent(NavigateMenuIntent{MenuID: inst.menuID(), Direction: "left", FromIndex: previous, ToIndex: inst.activeIndex})
			inst.dirty = true
		}
		return true
	case action.ActionNavigateRight, action.ActionNavigateNext:
		previous := inst.activeIndex
		inst.activeIndex = NextSelectableIndex(inst.model.Items, inst.activeIndex)
		if inst.activeIndex != previous {
			inst.EmitIntent(NavigateMenuIntent{MenuID: inst.menuID(), Direction: "right", FromIndex: previous, ToIndex: inst.activeIndex})
			inst.dirty = true
		}
		return true
	case action.ActionNavigateHome:
		inst.activeIndex = FirstSelectableIndex(inst.model.Items)
		inst.dirty = true
		return true
	case action.ActionNavigateEnd:
		inst.activeIndex = LastSelectableIndex(inst.model.Items)
		inst.dirty = true
		return true
	case action.ActionHover:
		if mouse, ok := mousePayload(act.Payload); ok && mouse.LocalY == 0 {
			if idx, hit := inst.barIndexAt(mouse.LocalX); hit {
				inst.activeIndex = idx
				inst.dirty = true
				return true
			}
		}
	case action.ActionClick, action.ActionEnter, action.ActionSubmit, action.ActionNavigateDown:
		if act.Type == action.ActionClick {
			if mouse, ok := mousePayload(act.Payload); ok {
				if idx, hit := inst.barIndexAt(mouse.LocalX); hit {
					inst.activeIndex = idx
				}
			}
		}
		inst.activateBarIndex(inst.activeIndex)
		return true
	case action.ActionCancel:
		inst.EmitIntent(CloseMenuIntent{MenuID: inst.menuID()})
		return true
	}
	return false
}

func (inst *popupInstance) HandleAction(act *action.Action) bool {
	if act == nil || !inst.open {
		return false
	}
	currentDepth := inst.currentDepth()
	switch act.Type {
	case action.ActionNavigateUp, action.ActionNavigatePrev:
		items := inst.itemsAtDepth(currentDepth)
		previous := inst.selectedIndexAtDepth(currentDepth)
		next := PrevSelectableIndex(items, previous)
		if next != previous {
			inst.setSelectedIndexAtDepth(currentDepth, next)
			inst.trimCascadeAfter(currentDepth)
			inst.openCurrentSubmenuIfNeeded(currentDepth, false)
			inst.EmitIntent(NavigateMenuIntent{MenuID: inst.menuID(), Direction: "up", FromIndex: previous, ToIndex: next})
			inst.ensureSelectionVisible()
			inst.dirty = true
		}
		return true
	case action.ActionNavigateDown, action.ActionNavigateNext:
		items := inst.itemsAtDepth(currentDepth)
		previous := inst.selectedIndexAtDepth(currentDepth)
		next := NextSelectableIndex(items, previous)
		if next != previous {
			inst.setSelectedIndexAtDepth(currentDepth, next)
			inst.trimCascadeAfter(currentDepth)
			inst.openCurrentSubmenuIfNeeded(currentDepth, false)
			inst.EmitIntent(NavigateMenuIntent{MenuID: inst.menuID(), Direction: "down", FromIndex: previous, ToIndex: next})
			inst.ensureSelectionVisible()
			inst.dirty = true
		}
		return true
	case action.ActionNavigateHome:
		inst.setSelectedIndexAtDepth(currentDepth, FirstSelectableIndex(inst.itemsAtDepth(currentDepth)))
		inst.trimCascadeAfter(currentDepth)
		inst.ensureSelectionVisible()
		inst.dirty = true
		return true
	case action.ActionNavigateEnd:
		inst.setSelectedIndexAtDepth(currentDepth, LastSelectableIndex(inst.itemsAtDepth(currentDepth)))
		inst.trimCascadeAfter(currentDepth)
		inst.ensureSelectionVisible()
		inst.dirty = true
		return true
	case action.ActionNavigatePageUp:
		inst.setSelectedIndexAtDepth(currentDepth, inst.pageMove(currentDepth, -1))
		inst.trimCascadeAfter(currentDepth)
		inst.ensureSelectionVisible()
		inst.dirty = true
		return true
	case action.ActionNavigatePageDown:
		inst.setSelectedIndexAtDepth(currentDepth, inst.pageMove(currentDepth, 1))
		inst.trimCascadeAfter(currentDepth)
		inst.ensureSelectionVisible()
		inst.dirty = true
		return true
	case action.ActionNavigateRight:
		return inst.openCurrentSubmenu(currentDepth)
	case action.ActionNavigateLeft, action.ActionCancel:
		if currentDepth > 0 {
			inst.trimCascadeAfter(currentDepth - 1)
			inst.ensureSelectionVisible()
			inst.dirty = true
			return true
		}
		inst.EmitIntent(CloseMenuIntent{MenuID: inst.menuID()})
		return true
	case action.ActionHover:
		if mouse, ok := mousePayload(act.Payload); ok {
			if depth, idx, hit := inst.popupIndexAt(mouse.LocalX, mouse.LocalY); hit {
				inst.setSelectedIndexAtDepth(depth, idx)
				inst.trimCascadeAfter(depth)
				inst.ensureSelectionVisible()
				inst.dirty = true
				inst.openCurrentSubmenuIfNeeded(depth, true)
				return true
			}
		}
	case action.ActionClick, action.ActionEnter, action.ActionSubmit:
		if act.Type == action.ActionClick {
			if mouse, ok := mousePayload(act.Payload); ok {
				if depth, idx, hit := inst.popupIndexAt(mouse.LocalX, mouse.LocalY); hit {
					inst.setSelectedIndexAtDepth(depth, idx)
					inst.trimCascadeAfter(depth)
					currentDepth = depth
				}
			}
		}
		inst.activatePopupIndex(currentDepth, inst.selectedIndexAtDepth(currentDepth))
		return true
	case action.ActionInputText:
		if !inst.model.Typeahead {
			return false
		}
		if text, ok := act.GetPayloadString(); ok {
			return inst.applyTypeahead(currentDepth, text)
		}
	case action.ActionInputChar:
		if !inst.model.Typeahead {
			return false
		}
		if r, ok := act.GetPayloadRune(); ok {
			return inst.applyTypeahead(currentDepth, string(r))
		}
	case action.ActionScroll:
		if delta, ok := scrollutil.DeltaFromAction(act); ok {
			if mouse, ok := mousePayload(act.Payload); ok {
				if depth, _, hit := inst.popupIndexAt(mouse.LocalX, mouse.LocalY); hit {
					inst.adjustScroll(depth, delta)
					return true
				}
			}
			inst.adjustScroll(currentDepth, delta)
			return true
		}
	}
	return false
}

func (inst *barInstance) SetFocus(focused bool) {
	inst.focused = focused
	inst.dirty = true
}

func (inst *barInstance) HasFocus() bool   { return inst.focused }
func (inst *barInstance) IsDisabled() bool { return false }

func (inst *popupInstance) SetFocus(focused bool) {
	inst.focused = focused
	inst.dirty = true
}

func (inst *popupInstance) HasFocus() bool   { return inst.focused }
func (inst *popupInstance) IsDisabled() bool { return !inst.open }

func (inst *barInstance) SetIntentEmitter(fn func(intent.Intent))   { inst.intentEmitter = fn }
func (inst *popupInstance) SetIntentEmitter(fn func(intent.Intent)) { inst.intentEmitter = fn }

func (inst *barInstance) EmitIntent(i intent.Intent) {
	if inst.intentEmitter != nil && i != nil {
		inst.intentEmitter(i)
	}
}

func (inst *popupInstance) EmitIntent(i intent.Intent) {
	if inst.intentEmitter != nil && i != nil {
		inst.intentEmitter(i)
	}
}

func (inst *popupInstance) syncRegistration() {
	if !inst.open {
		inst.unregister()
		return
	}
	if inst.registered {
		return
	}
	menuRegistryGlobal.register(inst)
	inst.registered = true
}

func (inst *popupInstance) unregister() {
	if !inst.registered {
		return
	}
	menuRegistryGlobal.unregister(inst)
	inst.registered = false
}

func (inst *popupInstance) close() {
	if !inst.open {
		return
	}
	inst.open = false
	inst.unregister()
	inst.dirty = true
	inst.EmitIntent(CloseMenuIntent{MenuID: inst.menuID()})
}

func (inst *barInstance) GetBounds() (x, y, w, h int) {
	return inst.bounds[0], inst.bounds[1], inst.bounds[2], inst.bounds[3]
}

func (inst *barInstance) SetBounds(x, y, w, h int) {
	inst.bounds = [4]int{x, y, w, h}
}

func (inst *barInstance) containsPoint(screenX, screenY int) bool {
	left, top, width, height := inst.bounds[0], inst.bounds[1], inst.bounds[2], inst.bounds[3]
	if width <= 0 || height <= 0 {
		return false
	}
	return screenX >= left && screenX < left+width && screenY >= top && screenY < top+height
}

func (inst *popupInstance) GetBounds() (x, y, w, h int) {
	return inst.bounds[0], inst.bounds[1], inst.bounds[2], inst.bounds[3]
}

func (inst *popupInstance) SetBounds(x, y, w, h int) {
	inst.bounds = [4]int{x, y, w, h}
}

func (inst *popupInstance) containsPoint(screenX, screenY int) bool {
	if !inst.open {
		return false
	}
	baseX, baseY := inst.bounds[0], inst.bounds[1]
	for _, surface := range inst.popupSurfaces() {
		left := baseX + surface.x
		top := baseY + surface.y
		right := left + surface.metrics.surfaceWidth + surface.metrics.shadowWidth
		bottom := top + surface.metrics.surfaceHeight + surface.metrics.shadowHeight
		if screenX >= left && screenX < right && screenY >= top && screenY < bottom {
			return true
		}
	}
	return false
}

func (inst *barInstance) barSegments() []barSegment {
	items := inst.model.Items
	segments := make([]barSegment, 0, len(items))
	cursor := 0
	for index, item := range items {
		item = item.Normalize()
		if !item.IsVisible() {
			continue
		}
		text := " " + item.Label + " "
		segments = append(segments, barSegment{index: index, x: cursor, width: paint.StringWidth(text), text: text})
		cursor += paint.StringWidth(text) + 1
	}
	return segments
}

func (inst *barInstance) barIndexAt(localX int) (int, bool) {
	for _, segment := range inst.barSegments() {
		if localX >= segment.x && localX < segment.x+segment.width {
			return segment.index, true
		}
	}
	return -1, false
}

func (inst *barInstance) activateBarIndex(index int) {
	if index < 0 || index >= len(inst.model.Items) {
		return
	}
	item := inst.model.Items[index].Normalize()
	if !item.IsSelectable() {
		return
	}
	if item.HasSubmenu() {
		inst.EmitIntent(OpenMenuIntent{MenuID: inst.menuID(), Path: []int{index}})
		return
	}
	inst.EmitIntent(ActivateMenuItemIntent{MenuID: inst.menuID(), ItemKey: item.Key, Path: []int{index}})
	inst.EmitIntent(item.EffectiveIntent())
}

func (inst *popupInstance) popupMetricsFor(items []MenuItem) popupMetrics {
	visible := make([]int, 0, len(items))
	contentWidth := 0
	markWidth := 0
	iconWidth := 0
	shortcutWidth := 0
	arrowWidth := 0
	for index, raw := range items {
		item := raw.Normalize()
		if !item.IsVisible() {
			continue
		}
		visible = append(visible, index)
		if item.IsSeparator() {
			continue
		}
		contentWidth = max(contentWidth, paint.StringWidth(item.Label))
		if inst.model.ShowDescriptions && item.Description != "" {
			contentWidth = max(contentWidth, paint.StringWidth(item.Label)+paint.StringWidth(" — ")+paint.StringWidth(item.Description))
		}
		if !inst.model.ShowDescriptions && item.SecondaryText != "" {
			contentWidth = max(contentWidth, paint.StringWidth(item.Label)+1+paint.StringWidth(item.SecondaryText))
		}
		if inst.model.ShowCheckMarks && (item.Kind == ItemCheckbox || item.Kind == ItemRadio || item.Checked) {
			markWidth = max(markWidth, 1)
		}
		if inst.model.ShowIcons && item.Icon != "" {
			iconWidth = max(iconWidth, max(1, paint.StringWidth(item.Icon)))
		}
		if inst.model.ShowShortcuts && strings.TrimSpace(item.Shortcut.DisplayText()) != "" {
			shortcutWidth = max(shortcutWidth, paint.StringWidth(item.Shortcut.DisplayText()))
		}
		if item.HasSubmenu() {
			arrowWidth = max(arrowWidth, 1)
		}
	}
	contentWidth = max(contentWidth, paint.StringWidth(inst.model.Title))
	fixedWidth := popupLeftPadding + popupRightPadding
	if markWidth > 0 {
		fixedWidth += markWidth + 1
	}
	if iconWidth > 0 {
		fixedWidth += iconWidth + 1
	}
	if shortcutWidth > 0 {
		fixedWidth += 2 + shortcutWidth
	}
	if arrowWidth > 0 {
		fixedWidth += 1 + arrowWidth
	}
	innerWidth := fixedWidth + max(contentWidth, 4)
	minInner := max(defaultPopupMinWidth-2, inst.model.MinWidth-2)
	if minInner > innerWidth {
		innerWidth = minInner
	}
	if inst.model.MaxWidth > 0 {
		maxInner := max(1, inst.model.MaxWidth-2)
		if innerWidth > maxInner {
			innerWidth = maxInner
		}
	}
	contentWidth = max(4, innerWidth-fixedWidth)
	surfaceWidth := innerWidth + 2
	totalRows := len(visible)
	viewportRows := totalRows
	if viewportRows == 0 {
		viewportRows = 1
	}
	if inst.model.MaxHeight > 0 {
		maxRows := max(1, inst.model.MaxHeight-2)
		if viewportRows > maxRows {
			viewportRows = maxRows
		}
	}
	surfaceHeight := viewportRows + 2
	shadowWidth := 0
	shadowHeight := 0
	if !inst.theme.SurfaceShadowStyle.IsEmpty() {
		shadowWidth = 1
		shadowHeight = 1
	}
	return popupMetrics{visibleIndices: visible, innerWidth: innerWidth, contentWidth: contentWidth, markWidth: markWidth, iconWidth: iconWidth, shortcutWidth: shortcutWidth, arrowWidth: arrowWidth, viewportRows: viewportRows, surfaceWidth: surfaceWidth, surfaceHeight: surfaceHeight, shadowWidth: shadowWidth, shadowHeight: shadowHeight}
}

func (inst *popupInstance) popupIndexAt(localX, localY int) (int, int, bool) {
	surfaces := inst.popupSurfaces()
	for _, surface := range surfaces {
		if localX < surface.x || localX >= surface.x+surface.metrics.surfaceWidth {
			continue
		}
		if localY < surface.y+1 || localY >= surface.y+1+surface.metrics.viewportRows {
			continue
		}
		row := localY - surface.y - 1
		itemPos := surface.scrollOffset + row
		if itemPos < 0 || itemPos >= len(surface.metrics.visibleIndices) {
			return 0, -1, false
		}
		itemIndex := surface.metrics.visibleIndices[itemPos]
		item := surface.items[itemIndex].Normalize()
		if !item.IsSelectable() {
			return 0, -1, false
		}
		return surface.depth, itemIndex, true
	}
	return 0, -1, false
}

func (inst *popupInstance) resolvePopupRowStyle(item MenuItem, selected bool) style.Style {
	rowStyle := inst.theme.ItemStyle.Merge(inst.theme.SurfaceStyle).Merge(inst.rootStyle)
	if item.Danger {
		rowStyle = rowStyle.Merge(inst.theme.ItemDangerStyle)
	}
	if item.Checked {
		rowStyle = rowStyle.Merge(inst.theme.ItemCheckedStyle)
	}
	if selected {
		rowStyle = rowStyle.Merge(inst.theme.ItemActiveStyle)
		if inst.focused {
			rowStyle = rowStyle.Merge(inst.theme.ItemFocusStyle)
		} else {
			rowStyle = rowStyle.Merge(inst.theme.ItemHoverStyle)
		}
	}
	if item.Disabled {
		rowStyle = rowStyle.Merge(inst.theme.ItemDisabledStyle)
	}
	if item.IsLabel() {
		rowStyle = rowStyle.Merge(inst.theme.TitleStyle)
	}
	return rowStyle
}

func (inst *popupInstance) currentItem(depth int) (MenuItem, bool) {
	items := inst.itemsAtDepth(depth)
	index := inst.selectedIndexAtDepth(depth)
	if index < 0 || index >= len(items) {
		return MenuItem{}, false
	}
	return items[index].Normalize(), true
}

func (inst *popupInstance) activatePopupIndex(depth, index int) {
	items := inst.itemsAtDepth(depth)
	if index < 0 || index >= len(items) {
		return
	}
	item := items[index].Normalize()
	if !item.IsSelectable() {
		return
	}
	if item.HasSubmenu() {
		inst.openCurrentSubmenu(depth)
		return
	}
	itemPath := inst.pathAtDepth(depth, index)
	inst.EmitIntent(ActivateMenuItemIntent{MenuID: inst.menuID(), ItemKey: item.Key, Path: itemPath})
	inst.EmitIntent(item.EffectiveIntent())
	if item.EffectiveCloseOnSelect() {
		inst.EmitIntent(CloseMenuIntent{MenuID: inst.menuID()})
	}
}

func (inst *popupInstance) ensureSelectionVisible() {
	inst.ensureSelectionVisibleAtDepth(0)
	for depth := 1; depth <= len(inst.submenuPath); depth++ {
		inst.ensureSelectionVisibleAtDepth(depth)
	}
}

func (inst *popupInstance) maxScrollOffset(depth int) int {
	metrics := inst.popupMetricsFor(inst.itemsAtDepth(depth))
	return max(0, len(metrics.visibleIndices)-metrics.viewportRows)
}

func (inst *popupInstance) pageMove(depth, direction int) int {
	items := inst.itemsAtDepth(depth)
	metrics := inst.popupMetricsFor(items)
	if len(metrics.visibleIndices) == 0 {
		return -1
	}
	position := 0
	for i, idx := range metrics.visibleIndices {
		if idx == inst.selectedIndexAtDepth(depth) {
			position = i
			break
		}
	}
	position += direction * max(1, metrics.viewportRows)
	position = clamp(position, 0, len(metrics.visibleIndices)-1)
	for position >= 0 && position < len(metrics.visibleIndices) {
		idx := metrics.visibleIndices[position]
		if items[idx].Normalize().IsSelectable() {
			return idx
		}
		position += direction
	}
	return inst.selectedIndexAtDepth(depth)
}

func (inst *popupInstance) applyTypeahead(depth int, text string) bool {
	text = strings.TrimSpace(text)
	if text == "" {
		return false
	}
	now := time.Now()
	timeout := inst.model.TypeaheadTimeout
	if timeout <= 0 {
		timeout = 750 * time.Millisecond
	}
	if now.Sub(inst.typeaheadAt) > timeout {
		inst.typeahead = ""
	}
	inst.typeaheadAt = now
	inst.typeahead += strings.ToLower(text)
	items := inst.itemsAtDepth(depth)
	currentIndex := inst.selectedIndexAtDepth(depth)
	if match := MatchTypeahead(items, inst.typeahead, currentIndex); match >= 0 {
		inst.setSelectedIndexAtDepth(depth, match)
		inst.trimCascadeAfter(depth)
		inst.ensureSelectionVisible()
		inst.dirty = true
		return true
	}
	last := strings.ToLower(text)
	inst.typeahead = last
	if match := MatchTypeahead(items, inst.typeahead, currentIndex); match >= 0 {
		inst.setSelectedIndexAtDepth(depth, match)
		inst.trimCascadeAfter(depth)
		inst.ensureSelectionVisible()
		inst.dirty = true
		return true
	}
	return false
}

func (inst *popupInstance) normalizeCascadeState() {
	if inst.selectedIndex < 0 {
		inst.selectedIndex = FirstSelectableIndex(inst.model.Items)
	}
	items := inst.model.Items
	depth := 0
	trimmedPath := make([]int, 0, len(inst.submenuPath))
	trimmedScroll := make([]int, 0, len(inst.submenuScroll))
	currentIndex := inst.selectedIndex
	for depth < len(inst.submenuPath) {
		if currentIndex < 0 || currentIndex >= len(items) {
			break
		}
		item := items[currentIndex].Normalize()
		if !item.HasSubmenu() || len(item.Children) == 0 {
			break
		}
		childIndex := inst.submenuPath[depth]
		if childIndex < 0 || childIndex >= len(item.Children) {
			childIndex = FirstSelectableIndex(item.Children)
			if childIndex < 0 {
				break
			}
		}
		trimmedPath = append(trimmedPath, childIndex)
		if depth < len(inst.submenuScroll) {
			trimmedScroll = append(trimmedScroll, clampNonNegative(inst.submenuScroll[depth]))
		} else {
			trimmedScroll = append(trimmedScroll, 0)
		}
		items = item.Children
		currentIndex = childIndex
		depth++
	}
	inst.submenuPath = trimmedPath
	inst.submenuScroll = trimmedScroll
}

func (inst *popupInstance) popupSurfaces() []popupSurface {
	rootMetrics := inst.popupMetricsFor(inst.model.Items)
	surfaces := []popupSurface{{
		depth:        0,
		items:        inst.model.Items,
		selectedIdx:  inst.selectedIndex,
		scrollOffset: inst.scrollOffset,
		x:            0,
		y:            0,
		metrics:      rootMetrics,
	}}
	prefix := []int{inst.selectedIndex}
	for depth := 1; depth <= len(inst.submenuPath); depth++ {
		parentSurface := surfaces[depth-1]
		parentItem, ok := ItemAtPath(inst.model.Items, prefix[:depth])
		if !ok || !parentItem.HasSubmenu() || len(parentItem.Children) == 0 {
			break
		}
		selectedIdx := inst.submenuPath[depth-1]
		metrics := inst.popupMetricsFor(parentItem.Children)
		rowPos := visiblePosition(parentSurface.metrics.visibleIndices, prefix[depth-1], parentSurface.scrollOffset)
		if rowPos < 0 {
			rowPos = 0
		}
		surface := popupSurface{
			depth:        depth,
			parentPath:   append([]int(nil), prefix[:depth]...),
			items:        parentItem.Children,
			selectedIdx:  selectedIdx,
			scrollOffset: inst.scrollAtDepth(depth),
			x:            parentSurface.x + parentSurface.metrics.surfaceWidth + parentSurface.metrics.shadowWidth,
			y:            parentSurface.y + rowPos,
			metrics:      metrics,
		}
		surfaces = append(surfaces, surface)
		prefix = append(prefix, selectedIdx)
	}
	return surfaces
}

func (inst *popupInstance) paintSurface(surface popupSurface, offsetX, offsetY int) []paint.DrawCmd {
	metrics := surface.metrics
	x := offsetX + surface.x
	y := offsetY + surface.y
	borderStyle := inst.theme.SurfaceBorderStyle.Merge(inst.rootStyle)
	fillStyle := inst.theme.SurfaceStyle.Merge(inst.rootStyle)
	shadowStyle := inst.theme.SurfaceShadowStyle
	innerWidth := metrics.innerWidth
	cmds := make([]paint.DrawCmd, 0, metrics.surfaceHeight*6)
	top := "┌" + strings.Repeat("─", max(0, metrics.surfaceWidth-2)) + "┐"
	bottom := "└" + strings.Repeat("─", max(0, metrics.surfaceWidth-2)) + "┘"
	cmds = append(cmds, paint.DrawCmd{X: x, Y: y, Text: top, Style: borderStyle})
	if surface.depth == 0 && strings.TrimSpace(inst.model.Title) != "" && metrics.surfaceWidth > 4 {
		title := " " + truncateWithEllipsis(inst.model.Title, metrics.surfaceWidth-4) + " "
		cmds = append(cmds, paint.DrawCmd{X: x + 2, Y: y, Text: title, Style: inst.theme.TitleStyle.Merge(fillStyle)})
	}
	for row := 0; row < metrics.viewportRows; row++ {
		yPos := y + 1 + row
		cmds = append(cmds,
			paint.DrawCmd{X: x, Y: yPos, Text: "│", Style: borderStyle},
			paint.DrawCmd{X: x + 1, Y: yPos, Text: strings.Repeat(" ", innerWidth), Style: fillStyle},
			paint.DrawCmd{X: x + metrics.surfaceWidth - 1, Y: yPos, Text: "│", Style: borderStyle},
		)
		itemPos := surface.scrollOffset + row
		if itemPos < 0 || itemPos >= len(metrics.visibleIndices) {
			continue
		}
		itemIndex := metrics.visibleIndices[itemPos]
		item := surface.items[itemIndex].Normalize()
		if item.IsSeparator() {
			cmds = append(cmds, paint.DrawCmd{X: x + 1, Y: yPos, Text: strings.Repeat("─", innerWidth), Style: inst.theme.SeparatorStyle.Merge(fillStyle)})
			continue
		}
		rowStyle := inst.resolvePopupRowStyle(item, itemIndex == surface.selectedIdx)
		cmds = append(cmds, paint.DrawCmd{X: x + 1, Y: yPos, Text: strings.Repeat(" ", innerWidth), Style: rowStyle})
		leftX := x + 1 + popupLeftPadding
		if metrics.markWidth > 0 {
			cmds = append(cmds, paint.DrawCmd{X: leftX, Y: yPos, Text: padDisplayWidth(markForItem(item), metrics.markWidth), Style: rowStyle.Merge(inst.theme.CheckmarkStyle)})
			leftX += metrics.markWidth + 1
		}
		if metrics.iconWidth > 0 {
			cmds = append(cmds, paint.DrawCmd{X: leftX, Y: yPos, Text: padDisplayWidth(item.Icon, metrics.iconWidth), Style: rowStyle.Merge(inst.theme.IconStyle)})
			leftX += metrics.iconWidth + 1
		}
		label := truncateWithEllipsis(item.Label, metrics.contentWidth)
		labelW := paint.StringWidth(label)
		cmds = append(cmds, paint.DrawCmd{X: leftX, Y: yPos, Text: label, Style: rowStyle})
		if inst.model.ShowDescriptions && item.Description != "" {
			remaining := metrics.contentWidth - labelW
			if remaining > 0 {
				desc := truncateWithEllipsis(" — "+item.Description, remaining)
				cmds = append(cmds, paint.DrawCmd{X: leftX + labelW, Y: yPos, Text: desc, Style: rowStyle.Merge(inst.theme.DescriptionStyle)})
			}
		} else if item.SecondaryText != "" {
			remaining := metrics.contentWidth - labelW
			if remaining > 0 {
				secondary := truncateWithEllipsis(" "+item.SecondaryText, remaining)
				cmds = append(cmds, paint.DrawCmd{X: leftX + labelW, Y: yPos, Text: secondary, Style: rowStyle.Merge(inst.theme.DescriptionStyle)})
			}
		}
		if metrics.shortcutWidth > 0 {
			shortcutText := padDisplayWidth(truncateWithEllipsis(item.Shortcut.DisplayText(), metrics.shortcutWidth), metrics.shortcutWidth)
			shortcutX := x + 1 + innerWidth - popupRightPadding - metrics.arrowWidth - metrics.shortcutWidth
			if metrics.arrowWidth > 0 {
				shortcutX--
			}
			cmds = append(cmds, paint.DrawCmd{X: shortcutX, Y: yPos, Text: shortcutText, Style: rowStyle.Merge(inst.theme.ShortcutStyle)})
		}
		if metrics.arrowWidth > 0 && item.HasSubmenu() {
			arrowX := x + 1 + innerWidth - popupRightPadding - metrics.arrowWidth
			cmds = append(cmds, paint.DrawCmd{X: arrowX, Y: yPos, Text: padDisplayWidth("›", metrics.arrowWidth), Style: rowStyle.Merge(inst.theme.SubmenuArrowStyle)})
		}
	}
	cmds = append(cmds, paint.DrawCmd{X: x, Y: y + metrics.surfaceHeight - 1, Text: bottom, Style: borderStyle})
	if metrics.shadowWidth > 0 && !shadowStyle.IsEmpty() {
		for row := 0; row < metrics.surfaceHeight; row++ {
			cmds = append(cmds, paint.DrawCmd{X: x + metrics.surfaceWidth, Y: y + row + metrics.shadowHeight - 1, Text: strings.Repeat(" ", metrics.shadowWidth), Style: shadowStyle})
		}
	}
	if metrics.shadowHeight > 0 && !shadowStyle.IsEmpty() {
		cmds = append(cmds, paint.DrawCmd{X: x + metrics.shadowWidth, Y: y + metrics.surfaceHeight, Text: strings.Repeat(" ", metrics.surfaceWidth), Style: shadowStyle})
	}
	return cmds
}

func (inst *popupInstance) itemsAtDepth(depth int) []MenuItem {
	if depth <= 0 {
		return inst.model.Items
	}
	path := inst.activePath()
	if depth > len(path) {
		return nil
	}
	children, ok := ChildrenAtPath(inst.model.Items, path[:depth])
	if !ok {
		return nil
	}
	return children
}

func (inst *popupInstance) selectedIndexAtDepth(depth int) int {
	if depth <= 0 {
		return inst.selectedIndex
	}
	if depth-1 >= len(inst.submenuPath) {
		return -1
	}
	return inst.submenuPath[depth-1]
}

func (inst *popupInstance) setSelectedIndexAtDepth(depth, index int) {
	if depth <= 0 {
		inst.selectedIndex = index
		return
	}
	if len(inst.submenuPath) < depth {
		inst.submenuPath = append(inst.submenuPath, make([]int, depth-len(inst.submenuPath))...)
	}
	inst.submenuPath[depth-1] = index
	inst.submenuScroll = ensureLength(inst.submenuScroll, len(inst.submenuPath))
}

func (inst *popupInstance) scrollAtDepth(depth int) int {
	if depth <= 0 {
		return inst.scrollOffset
	}
	if depth-1 >= len(inst.submenuScroll) {
		return 0
	}
	return inst.submenuScroll[depth-1]
}

func (inst *popupInstance) setScrollAtDepth(depth, offset int) {
	if depth <= 0 {
		inst.scrollOffset = offset
		return
	}
	inst.submenuScroll = ensureLength(inst.submenuScroll, depth)
	inst.submenuScroll[depth-1] = offset
}

func (inst *popupInstance) currentDepth() int {
	return len(inst.submenuPath)
}

func (inst *popupInstance) activePath() []int {
	if inst.selectedIndex < 0 {
		return nil
	}
	path := []int{inst.selectedIndex}
	path = append(path, inst.submenuPath...)
	return path
}

func (inst *popupInstance) prefixedPath(path []int) []int {
	fullPath := append([]int(nil), inst.model.PathPrefix...)
	fullPath = append(fullPath, path...)
	return fullPath
}

func trimPathPrefix(path, prefix []int) []int {
	if len(prefix) == 0 || len(path) < len(prefix) {
		return append([]int(nil), path...)
	}
	for index, value := range prefix {
		if path[index] != value {
			return append([]int(nil), path...)
		}
	}
	return append([]int(nil), path[len(prefix):]...)
}

func (inst *popupInstance) pathAtDepth(depth, index int) []int {
	if depth <= 0 {
		return inst.prefixedPath([]int{index})
	}
	path := []int{inst.selectedIndex}
	path = append(path, inst.submenuPath[:depth-1]...)
	path = append(path, index)
	return inst.prefixedPath(path)
}

func (inst *popupInstance) trimCascadeAfter(depth int) {
	if depth < 0 {
		depth = 0
	}
	if depth == 0 {
		inst.submenuPath = nil
		inst.submenuScroll = nil
		return
	}
	if len(inst.submenuPath) > depth {
		inst.submenuPath = append([]int(nil), inst.submenuPath[:depth]...)
	}
	if len(inst.submenuScroll) > depth {
		inst.submenuScroll = append([]int(nil), inst.submenuScroll[:depth]...)
	}
}

func (inst *popupInstance) ensureSelectionVisibleAtDepth(depth int) {
	items := inst.itemsAtDepth(depth)
	metrics := inst.popupMetricsFor(items)
	if len(metrics.visibleIndices) == 0 {
		inst.setScrollAtDepth(depth, 0)
		return
	}
	selectedIndex := inst.selectedIndexAtDepth(depth)
	if selectedIndex < 0 {
		selectedIndex = FirstSelectableIndex(items)
		inst.setSelectedIndexAtDepth(depth, selectedIndex)
	}
	visiblePosition := visiblePosition(metrics.visibleIndices, selectedIndex, 0)
	if visiblePosition < 0 {
		return
	}
	scrollOffset := inst.scrollAtDepth(depth)
	maxOffset := inst.maxScrollOffset(depth)
	if visiblePosition < scrollOffset {
		scrollOffset = visiblePosition
	} else if visiblePosition >= scrollOffset+metrics.viewportRows {
		scrollOffset = visiblePosition - metrics.viewportRows + 1
	}
	inst.setScrollAtDepth(depth, clamp(scrollOffset, 0, maxOffset))
}

func (inst *popupInstance) openCurrentSubmenu(depth int) bool {
	item, ok := inst.currentItem(depth)
	if !ok || !item.HasSubmenu() {
		return false
	}
	childIndex := FirstSelectableIndex(item.Children)
	if childIndex < 0 {
		inst.EmitIntent(OpenMenuIntent{MenuID: inst.menuID(), Path: inst.pathAtDepth(depth, inst.selectedIndexAtDepth(depth))})
		return true
	}
	inst.submenuPath = ensureLength(inst.submenuPath, depth+1)
	inst.submenuScroll = ensureLength(inst.submenuScroll, depth+1)
	inst.submenuPath[depth] = childIndex
	inst.submenuScroll[depth] = 0
	inst.ensureSelectionVisible()
	inst.EmitIntent(OpenMenuIntent{MenuID: inst.menuID(), Path: inst.pathAtDepth(depth, inst.selectedIndexAtDepth(depth))})
	inst.dirty = true
	return true
}

func (inst *popupInstance) openCurrentSubmenuIfNeeded(depth int, viaHover bool) {
	item, ok := inst.currentItem(depth)
	if !ok {
		return
	}
	if !item.HasSubmenu() {
		inst.trimCascadeAfter(depth)
		return
	}
	if viaHover || depth < inst.currentDepth() {
		inst.openCurrentSubmenu(depth)
	}
}

func (inst *popupInstance) adjustScroll(depth, delta int) {
	maxOffset := inst.maxScrollOffset(depth)
	next := clamp(inst.scrollAtDepth(depth)+delta, 0, maxOffset)
	inst.setScrollAtDepth(depth, next)
	inst.dirty = true
}

func (inst *barInstance) menuID() string {
	return firstNonEmpty(inst.model.ComponentID, inst.model.ID, inst.key, "menu")
}

func (inst *popupInstance) menuID() string {
	return firstNonEmpty(inst.model.ComponentID, inst.model.ID, inst.key, "menu")
}

func mousePayload(payload any) (*runtimemsg.MouseMsg, bool) {
	switch value := payload.(type) {
	case *runtimemsg.MouseMsg:
		if value != nil {
			return value, true
		}
	case runtimemsg.MouseMsg:
		copy := value
		return &copy, true
	}
	return nil, false
}

func getModelProp(props rtui.Props, def Model) Model {
	if value, ok := props["model"].(Model); ok {
		return value
	}
	return def
}

func getThemeProp(props rtui.Props, def Theme) Theme {
	if value, ok := props["theme"].(Theme); ok {
		return value
	}
	return def
}

func getStyleProp(props rtui.Props, key string) style.Style {
	if value, ok := props[key].(style.Style); ok {
		return value
	}
	return style.Style{}
}

func getStringProp(props rtui.Props, key, def string) string {
	if value, ok := props[key].(string); ok {
		return value
	}
	return def
}

func getBoolProp(props rtui.Props, key string, def bool) bool {
	if value, ok := props[key].(bool); ok {
		return value
	}
	return def
}

func getNodeStyle(v interface{ Style() style.Style }) style.Style {
	if v == nil {
		return style.Style{}
	}
	return v.Style()
}

func applyPortalProps(node *rtui.ElementVNode, model Model) {
	if node == nil {
		return
	}
	if model.PortalRoot != "" {
		node.SetPortalRoot(model.PortalRoot)
	}
	if model.AnchorID != "" {
		node.SetAnchorTo(model.AnchorID, model.Anchor)
	}
	if model.PortalPosition != rttypes.PositionRelative {
		node.SetPortalPosition(model.PortalPosition)
	}
	if model.PortalPriority != 0 {
		node.SetPortalPriority(model.PortalPriority)
	}
	if model.PortalOffsetX != 0 {
		node.SetProp("left", model.PortalOffsetX)
	}
	if model.PortalOffsetY != 0 {
		node.SetProp("top", model.PortalOffsetY)
	}
}

func markForItem(item MenuItem) string {
	switch item.Kind {
	case ItemCheckbox:
		if item.Checked {
			return "✓"
		}
		return ""
	case ItemRadio:
		if item.Checked {
			return "●"
		}
		return ""
	default:
		if item.Checked {
			return "✓"
		}
		return ""
	}
}

func truncateWithEllipsis(content string, width int) string {
	if width <= 0 {
		return ""
	}
	const ellipsis = "…"
	if paint.StringWidth(content) <= width {
		return content
	}
	ellipsisWidth := paint.StringWidth(ellipsis)
	if width <= ellipsisWidth {
		return ellipsis
	}
	trimmed := strings.TrimRight(truncateByDisplayWidth(content, width-ellipsisWidth), " ")
	if trimmed == "" {
		return truncateByDisplayWidth(content, width)
	}
	return trimmed + ellipsis
}

func truncateByDisplayWidth(content string, width int) string {
	if width <= 0 {
		return ""
	}
	var builder strings.Builder
	currentWidth := 0
	for _, r := range content {
		runeWidth := paint.RuneWidth(r)
		if currentWidth+runeWidth > width {
			break
		}
		builder.WriteRune(r)
		currentWidth += runeWidth
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

func clamp(value, low, high int) int {
	if value < low {
		return low
	}
	if value > high {
		return high
	}
	return value
}

func clampIndex(index, length int) int {
	if length <= 0 {
		return -1
	}
	if index < 0 {
		return 0
	}
	if index >= length {
		return length - 1
	}
	return index
}

func clampNonNegative(value int) int {
	if value < 0 {
		return 0
	}
	return value
}

func applyWidthConstraints(width int, constraints layout.Constraints) int {
	if constraints.MinWidth > 0 {
		width = max(width, constraints.MinWidth)
	}
	if constraints.MaxWidth > 0 && constraints.MaxWidth < layout.MaxInt {
		width = min(width, constraints.MaxWidth)
	}
	return width
}

func applyHeightConstraints(height int, constraints layout.Constraints) int {
	if constraints.MinHeight > 0 {
		height = max(height, constraints.MinHeight)
	}
	if constraints.MaxHeight > 0 && constraints.MaxHeight < layout.MaxInt {
		height = min(height, constraints.MaxHeight)
	}
	return height
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func ensureLength(values []int, length int) []int {
	if length <= 0 {
		return nil
	}
	if len(values) >= length {
		return append([]int(nil), values[:length]...)
	}
	out := append([]int(nil), values...)
	out = append(out, make([]int, length-len(out))...)
	return out
}

func visiblePosition(indices []int, selected, scrollOffset int) int {
	for i, idx := range indices {
		if idx == selected {
			return i - scrollOffset
		}
	}
	return -1
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
