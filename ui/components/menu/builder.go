package menu

import (
	"time"

	"github.com/wwsheng009/mint/runtime/style"
	rttypes "github.com/wwsheng009/mint/runtime/types"
	rtui "github.com/wwsheng009/mint/runtime/ui"
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

