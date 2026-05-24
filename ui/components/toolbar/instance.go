package toolbar

import (
	"reflect"
	"strings"

	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/button"
	menucomp "github.com/wwsheng009/mint/ui/components/menu"
	"github.com/wwsheng009/mint/ui/components/statusbar"
	"github.com/wwsheng009/mint/ui/components/tag"
	"github.com/wwsheng009/mint/ui/components/text"
)

// Instance is the runtime entity for Toolbar.
type Instance struct {
	key          string
	title        string
	titleWidth   int
	width        int
	gap          int
	dense        bool
	separator    string
	useStatusBar bool
	leftItems    []Item
	centerItems  []Item
	rightItems   []Item
	rootStyle    style.Style
	dirty        bool
}

var (
	_ rtui.ComponentInstance       = (*Instance)(nil)
	_ rtui.RuntimeChildrenProvider = (*Instance)(nil)
)

// NewInstance creates a Toolbar instance from props.
func NewInstance(props rtui.Props) *Instance {
	inst := &Instance{
		key:          getStringProp(props, propKey, ""),
		title:        getStringProp(props, propTitle, ""),
		titleWidth:   getIntProp(props, propTitleWidth, 0),
		width:        getIntProp(props, propWidth, 0),
		gap:          getIntProp(props, propGap, 1),
		dense:        getBoolProp(props, propDense, false),
		separator:    getStringProp(props, propSeparator, "|"),
		useStatusBar: getBoolProp(props, propUseStatusBar, false),
		leftItems:    normalizeItems(getItemsProp(props, propLeftItems, nil)),
		centerItems:  normalizeItems(getItemsProp(props, propCenterItems, nil)),
		rightItems:   normalizeItems(getItemsProp(props, propRightItems, getItemsProp(props, propActions, nil))),
		rootStyle:    getStyleProp(props, propStyle, style.Style{}),
		dirty:        true,
	}
	inst.normalize()
	return inst
}

func (inst *Instance) Key() string       { return inst.key }
func (inst *Instance) SetKey(key string) { inst.key = key }
func (inst *Instance) Init(props rtui.Props) {
	inst.SetProps(props)
}
func (inst *Instance) Destroy()   {}
func (inst *Instance) OnMount()   {}
func (inst *Instance) OnUnmount() {}
func (inst *Instance) MarkDirty() { inst.dirty = true }
func (inst *Instance) IsDirty() bool {
	return inst.dirty
}
func (inst *Instance) GetContext() *rtui.ComponentContext {
	return nil
}

func (inst *Instance) SetProps(props rtui.Props) bool {
	old := inst.snapshot()

	inst.key = getStringProp(props, propKey, inst.key)
	inst.title = getStringProp(props, propTitle, inst.title)
	inst.titleWidth = getIntProp(props, propTitleWidth, inst.titleWidth)
	inst.width = getIntProp(props, propWidth, inst.width)
	inst.gap = getIntProp(props, propGap, inst.gap)
	inst.dense = getBoolProp(props, propDense, inst.dense)
	inst.separator = getStringProp(props, propSeparator, inst.separator)
	inst.useStatusBar = getBoolProp(props, propUseStatusBar, inst.useStatusBar)
	inst.leftItems = normalizeItems(getItemsProp(props, propLeftItems, inst.leftItems))
	inst.centerItems = normalizeItems(getItemsProp(props, propCenterItems, inst.centerItems))
	inst.rightItems = normalizeItems(getItemsProp(props, propRightItems, getItemsProp(props, propActions, inst.rightItems)))
	inst.rootStyle = getStyleProp(props, propStyle, inst.rootStyle)
	inst.normalize()

	changed := !reflect.DeepEqual(old, inst.snapshot())
	if changed {
		inst.dirty = true
	}
	return changed
}

func (inst *Instance) GetProps() rtui.Props {
	return rtui.Props{
		propActions:      cloneItems(inst.rightItems),
		propCenterItems:  cloneItems(inst.centerItems),
		propDense:        inst.dense,
		propGap:          inst.gap,
		propKey:          inst.key,
		propLeftItems:    cloneItems(inst.leftItems),
		propRightItems:   cloneItems(inst.rightItems),
		propSeparator:    inst.separator,
		propStyle:        inst.rootStyle,
		propTitle:        inst.title,
		propTitleWidth:   inst.titleWidth,
		propUseStatusBar: inst.useStatusBar,
		propWidth:        inst.width,
	}
}

// RuntimeChildren synthesizes the Toolbar controls used by Fiber.
func (inst *Instance) RuntimeChildren() []rtui.VNode {
	if inst.empty() {
		return nil
	}
	var root rtui.VNode
	if inst.useStatusBar {
		root = inst.buildStatusBar()
	} else {
		root = inst.buildToolbarRow()
	}
	if root == nil {
		return nil
	}
	return []rtui.VNode{root}
}

func (inst *Instance) buildToolbarRow() rtui.VNode {
	children := make([]rtui.VNode, 0, 8+len(inst.leftItems)+len(inst.centerItems)+len(inst.rightItems))
	if title := inst.buildTitle(); title != nil {
		children = append(children, title)
	}
	children = append(children, inst.buildItems(inst.leftItems, "left")...)
	if len(inst.centerItems) > 0 {
		children = append(children, rtui.Spacer().Flex(1).Build())
		children = append(children, inst.buildItems(inst.centerItems, "center")...)
	}
	if len(inst.rightItems) > 0 {
		children = append(children, rtui.Spacer().Flex(1).Build())
		children = append(children, inst.buildItems(inst.rightItems, "right")...)
	}
	if len(children) == 0 {
		return nil
	}

	root := rtui.HStackBuilder(children...).Gap(inst.gap).AlignCross(rtui.AlignCenter)
	if inst.width > 0 {
		root.Width(inst.width)
	}
	if !inst.rootStyle.IsEmpty() {
		root.SetStyleProps(inst.rootStyle)
	}
	root.SetKey(inst.rootKey())
	row := root.Build()

	overlays := inst.buildMenuOverlays()
	if len(overlays) == 0 {
		return row
	}
	nodes := make([]rtui.VNode, 0, 1+len(overlays))
	nodes = append(nodes, row)
	nodes = append(nodes, overlays...)
	return rtui.Fragment(nodes...)
}

func (inst *Instance) buildTitle() rtui.VNode {
	if strings.TrimSpace(inst.title) == "" {
		return nil
	}
	node := rtui.VNode(text.NewBuilder(inst.title).
		Key(inst.childKey("title")).
		Style(style.NewStyle().Foreground(theme.Text()).Bold(true)).
		Build())
	if inst.titleWidth > 0 {
		box := rtui.Box().Width(inst.titleWidth).Child(node).Build()
		box.SetKey(inst.childKey("title-box"))
		node = box
	}
	return node
}

func (inst *Instance) buildItems(items []Item, slot string) []rtui.VNode {
	children := make([]rtui.VNode, 0, len(items))
	for _, item := range items {
		if node := inst.buildItem(item, slot); node != nil {
			children = append(children, node)
		}
	}
	return children
}

func (inst *Instance) buildItem(item Item, slot string) rtui.VNode {
	switch item.Kind {
	case ItemMenu:
		return inst.buildMenuButton(item, slot)
	case ItemButton:
		return inst.buildButton(item, slot)
	case ItemBadge:
		return inst.wrapWidthIfNeeded(item, slot, tag.NewBuilder(item.Label).
			Key(inst.itemKey(slot, item)).
			Style(inst.itemStyle(item, style.NewStyle().Foreground(style.Black).Background(style.BrightBlack).Bold(true))).
			Build())
	case ItemSeparator:
		label := inst.separator
		if label == "" {
			label = "|"
		}
		return text.NewBuilder(label).
			Key(inst.itemKey(slot, item)).
			Style(inst.itemStyle(item, style.NewStyle().Foreground(theme.Muted()))).
			Build()
	case ItemCustom:
		if item.Custom == nil {
			return nil
		}
		if item.Custom.Key() == "" {
			item.Custom.SetKey(inst.itemKey(slot, item))
		}
		return inst.wrapWidthIfNeeded(item, slot, item.Custom)
	default:
		return inst.wrapWidthIfNeeded(item, slot, text.NewBuilder(item.Label).
			Key(inst.itemKey(slot, item)).
			Style(inst.itemStyle(item, style.NewStyle().Foreground(theme.Text()))).
			Build())
	}
}

func (inst *Instance) buildButton(item Item, slot string) rtui.VNode {
	if strings.TrimSpace(item.Label) == "" {
		return nil
	}
	builder := button.NewBuilder(item.Label).
		Key(inst.itemKey(slot, item)).
		SetID(inst.itemKey(slot, item)).
		Variant(item.Variant).
		Disabled(item.Disabled)
	if inst.dense {
		builder.Small()
	}
	if item.PressIntent != nil {
		builder.OnPress(item.PressIntent)
	}
	if !inst.explicitItemStyle(item).IsEmpty() {
		builder.Style(inst.explicitItemStyle(item))
	}
	return inst.wrapWidthIfNeeded(item, slot, builder.Build())
}

func (inst *Instance) buildMenuButton(item Item, slot string) rtui.VNode {
	if strings.TrimSpace(item.Label) == "" {
		return nil
	}
	pressIntent := item.PressIntent
	if pressIntent == nil {
		pressIntent = menucomp.OpenMenuIntent{MenuID: inst.menuID(slot, item)}
	}
	return inst.buildButton(item.OnPress(pressIntent), slot)
}

func (inst *Instance) buildMenuOverlays() []rtui.VNode {
	overlays := make([]rtui.VNode, 0)
	for _, entry := range []struct {
		slot  string
		items []Item
	}{
		{"left", inst.leftItems},
		{"center", inst.centerItems},
		{"right", inst.rightItems},
	} {
		for _, item := range entry.items {
			if item.Kind != ItemMenu || !item.MenuOpen || len(item.MenuItems) == 0 {
				continue
			}
			overlay := menucomp.NewPopup(item.MenuItems).
				Key(inst.menuID(entry.slot, item)).
				ComponentID(inst.menuID(entry.slot, item)).
				Open(true).
				AnchorTo(inst.itemKey(entry.slot, item), menuAnchorForPlacement(item.MenuPlacement)).
				Placement(item.MenuPlacement).
				ShowShortcuts(item.MenuShowShortcuts).
				ShowDescriptions(item.MenuShowDescriptions)
			if len(item.MenuActivePath) > 0 {
				overlay.ActivePath(item.MenuActivePath...)
			}
			if item.MenuMinWidth > 0 {
				overlay.MinWidth(item.MenuMinWidth)
			}
			if item.MenuMaxHeight > 0 {
				overlay.MaxHeight(item.MenuMaxHeight)
			}
			overlays = append(overlays, overlay.Build())
		}
	}
	return overlays
}

func (inst *Instance) buildStatusBar() rtui.VNode {
	builder := statusbar.NewBuilder().
		Theme(statusbar.MutedTheme()).
		Gap(inst.gap).
		HelpFallback("").
		HelpDisplayMode(statusbar.HelpDisplayOverlay)

	left := inst.itemsToSections(inst.leftItems, "left")
	if title := strings.TrimSpace(inst.title); title != "" {
		titleSection := statusbar.Text(title).WithKey(inst.childKey("title")).WithBold(true)
		if inst.titleWidth > 0 {
			titleSection = titleSection.WithWidth(inst.titleWidth)
		}
		left = append([]statusbar.Section{titleSection}, left...)
	}
	builder.LeftSections(left...)
	builder.CenterSections(inst.itemsToSections(inst.centerItems, "center")...)
	builder.RightSections(inst.itemsToSections(inst.rightItems, "right")...)

	root := builder.BuildWithHelp()
	if inst.width > 0 || !inst.rootStyle.IsEmpty() {
		box := rtui.Box().Child(root)
		if inst.width > 0 {
			box.Width(inst.width)
		}
		root = box.Build()
		root.SetKey(inst.rootKey())
		if !inst.rootStyle.IsEmpty() {
			root.SetStyle(inst.rootStyle)
		}
		return root
	}
	root.SetKey(inst.rootKey())
	return root
}

func (inst *Instance) itemsToSections(items []Item, slot string) []statusbar.Section {
	sections := make([]statusbar.Section, 0, len(items))
	for _, item := range items {
		section, ok := inst.itemToSection(item, slot)
		if ok {
			sections = append(sections, section)
		}
	}
	return sections
}

func (inst *Instance) itemToSection(item Item, slot string) (statusbar.Section, bool) {
	label := item.Label
	if item.Kind == ItemSeparator {
		label = inst.separator
		if label == "" {
			label = "|"
		}
	}
	if strings.TrimSpace(label) == "" {
		return statusbar.Section{}, false
	}

	var section statusbar.Section
	switch item.Kind {
	case ItemBadge:
		section = statusbar.Badge(label, firstNonEmpty(item.FgColor, "black"), firstNonEmpty(item.BgColor, "bright-black"))
	case ItemButton, ItemMenu:
		pressIntent := item.PressIntent
		if item.Kind == ItemMenu && pressIntent == nil {
			pressIntent = menucomp.OpenMenuIntent{MenuID: inst.menuID(slot, item)}
		}
		section = statusbar.ActionText(label, pressIntent)
	default:
		section = statusbar.Text(label)
	}
	section = section.WithKey(inst.itemKey(slot, item))
	if item.Width > 0 {
		section = section.WithWidth(item.Width).WithEllipsis()
	}
	if item.Bold {
		section = section.WithBold(true)
	}
	if item.FgColor != "" || item.BgColor != "" {
		section = section.WithColors(item.FgColor, item.BgColor)
	}
	if item.Disabled {
		section = section.WithDisabled(true)
	}
	if item.PressIntent != nil && item.Kind != ItemButton {
		section = section.OnPress(item.PressIntent)
	}
	if item.HelpText != "" {
		section = section.WithHelp(item.HelpText)
	}
	return section, true
}

func (inst *Instance) wrapWidthIfNeeded(item Item, slot string, node rtui.VNode) rtui.VNode {
	if node == nil || item.Width <= 0 {
		return node
	}
	box := rtui.Box().Width(item.Width).Child(node).Build()
	box.SetKey(inst.itemKey(slot, item) + "-box")
	return box
}

func (inst *Instance) itemStyle(item Item, fallback style.Style) style.Style {
	explicit := inst.explicitItemStyle(item)
	if explicit.IsEmpty() {
		return fallback
	}
	return fallback.Merge(explicit)
}

func (inst *Instance) explicitItemStyle(item Item) style.Style {
	s := style.NewStyle()
	if item.FgColor != "" {
		s = s.Foreground(style.Color(item.FgColor))
	}
	if item.BgColor != "" {
		s = s.Background(style.Color(item.BgColor))
	}
	if item.Bold {
		s = s.Bold(true)
	}
	return s
}

func (inst *Instance) empty() bool {
	return strings.TrimSpace(inst.title) == "" &&
		len(inst.leftItems) == 0 &&
		len(inst.centerItems) == 0 &&
		len(inst.rightItems) == 0
}

func (inst *Instance) rootKey() string {
	if inst.key == "" {
		return "toolbar-root"
	}
	return inst.key + "-root"
}

func (inst *Instance) childKey(suffix string) string {
	if inst.key == "" {
		return "toolbar-" + suffix
	}
	return inst.key + "-" + suffix
}

func (inst *Instance) itemKey(slot string, item Item) string {
	return inst.childKey(slot + "-" + item.Key)
}

func (inst *Instance) menuID(slot string, item Item) string {
	if strings.TrimSpace(item.MenuID) != "" {
		return item.MenuID
	}
	return inst.itemKey(slot, item) + "-menu"
}

func (inst *Instance) normalize() {
	inst.leftItems = normalizeItems(inst.leftItems)
	inst.centerItems = normalizeItems(inst.centerItems)
	inst.rightItems = normalizeItems(inst.rightItems)
	if inst.width < 0 {
		inst.width = 0
	}
	if inst.titleWidth < 0 {
		inst.titleWidth = 0
	}
	if inst.gap < 0 {
		inst.gap = 0
	}
}

type instanceSnapshot struct {
	key          string
	title        string
	titleWidth   int
	width        int
	gap          int
	dense        bool
	separator    string
	useStatusBar bool
	leftItems    []Item
	centerItems  []Item
	rightItems   []Item
	rootStyle    style.Style
}

func (inst *Instance) snapshot() instanceSnapshot {
	return instanceSnapshot{
		key:          inst.key,
		title:        inst.title,
		titleWidth:   inst.titleWidth,
		width:        inst.width,
		gap:          inst.gap,
		dense:        inst.dense,
		separator:    inst.separator,
		useStatusBar: inst.useStatusBar,
		leftItems:    cloneItems(inst.leftItems),
		centerItems:  cloneItems(inst.centerItems),
		rightItems:   cloneItems(inst.rightItems),
		rootStyle:    inst.rootStyle,
	}
}

func getItemsProp(props rtui.Props, key string, def []Item) []Item {
	if items, ok := props[key].([]Item); ok {
		return cloneItems(items)
	}
	return cloneItems(def)
}

func getStringProp(props rtui.Props, key, def string) string {
	if value, ok := props[key]; ok {
		if text, ok := value.(string); ok {
			return text
		}
	}
	return def
}

func getIntProp(props rtui.Props, key string, def int) int {
	if value, ok := props[key]; ok {
		if number, ok := value.(int); ok {
			return number
		}
	}
	return def
}

func getBoolProp(props rtui.Props, key string, def bool) bool {
	if value, ok := props[key]; ok {
		if flag, ok := value.(bool); ok {
			return flag
		}
	}
	return def
}

func getStyleProp(props rtui.Props, key string, def style.Style) style.Style {
	if value, ok := props[key]; ok {
		if s, ok := value.(style.Style); ok {
			return s
		}
	}
	return def
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
