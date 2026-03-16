package breadcrumb

import (
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

const (
	propCurrentStyle   = "currentStyle"
	propItems          = "items"
	propItemStyle      = "itemStyle"
	propKey            = "key"
	propMaxWidth       = "maxWidth"
	propSeparator      = "separator"
	propSeparatorStyle = "separatorStyle"
	propStyle          = "style"
)

// Item describes a single breadcrumb segment.
type Item struct {
	Key     string
	Label   string
	Icon    string
	Current bool
}

// WithKey sets the item key.
func (i Item) WithKey(key string) Item {
	i.Key = key
	return i
}

// WithIcon sets the item icon prefix.
func (i Item) WithIcon(icon string) Item {
	i.Icon = icon
	return i
}

// AsCurrent marks the item as the active breadcrumb.
func (i Item) AsCurrent() Item {
	i.Current = true
	return i
}

// Crumb creates a breadcrumb item with the given label.
func Crumb(label string) Item {
	return Item{Label: label}
}

// Current creates an active breadcrumb item.
func Current(label string) Item {
	return Item{Label: label, Current: true}
}

// VNode is the immutable description of a Breadcrumb component.
type VNode struct {
	*rtui.ElementVNode

	key             string
	items           []Item
	separator       string
	maxWidth        int
	breadcrumbStyle style.Style
	itemStyle       style.Style
	currentStyle    style.Style
	separatorStyle  style.Style
}

var (
	_ rtui.VNode           = (*VNode)(nil)
	_ rtui.InstanceFactory = (*VNode)(nil)
)

// New creates a new Breadcrumb VNode.
func New(items []Item) *VNode {
	return &VNode{
		ElementVNode: rtui.NewElement("breadcrumb"),
		items:        cloneItems(items),
		separator:    " / ",
	}
}

func (v *VNode) Key() string { return v.key }

func (v *VNode) SetKey(key string) rtui.VNode {
	v.key = key
	return v
}

func (v *VNode) Tag() string { return "breadcrumb" }

func (v *VNode) Style() style.Style { return v.breadcrumbStyle }

func (v *VNode) SetStyle(s style.Style) rtui.VNode {
	v.breadcrumbStyle = s
	return v
}

func (v *VNode) Children() []rtui.VNode { return nil }

func (v *VNode) SetChildren(children []rtui.VNode) rtui.VNode { return v }

func (v *VNode) GetLayer() rtui.Layer { return rtui.LayerBase }

func (v *VNode) SetLayer(l rtui.Layer) rtui.VNode { return v }

func (v *VNode) Props() rtui.Props {
	return rtui.Props{
		propKey:            v.key,
		propItems:          cloneItems(v.items),
		propSeparator:      v.separator,
		propMaxWidth:       v.maxWidth,
		propStyle:          v.breadcrumbStyle,
		propItemStyle:      v.itemStyle,
		propCurrentStyle:   v.currentStyle,
		propSeparatorStyle: v.separatorStyle,
	}
}

func (v *VNode) SetProps(props rtui.Props) rtui.VNode {
	if key, ok := props[propKey].(string); ok {
		v.key = key
	}
	if items, ok := props[propItems].([]Item); ok {
		v.items = cloneItems(items)
	}
	if separator, ok := props[propSeparator].(string); ok {
		v.separator = separator
	}
	if maxWidth, ok := props[propMaxWidth].(int); ok {
		v.maxWidth = maxWidth
	}
	if s, ok := props[propStyle].(style.Style); ok {
		v.breadcrumbStyle = s
	}
	if s, ok := props[propItemStyle].(style.Style); ok {
		v.itemStyle = s
	}
	if s, ok := props[propCurrentStyle].(style.Style); ok {
		v.currentStyle = s
	}
	if s, ok := props[propSeparatorStyle].(style.Style); ok {
		v.separatorStyle = s
	}
	return v
}

func (v *VNode) CreateInstance() rtui.ComponentInstance {
	return NewInstance(v.Props())
}

// SetItems sets the breadcrumb items.
func (v *VNode) SetItems(items []Item) *VNode {
	v.items = cloneItems(items)
	return v
}

// AddItem appends a breadcrumb item.
func (v *VNode) AddItem(item Item) *VNode {
	v.items = append(v.items, item)
	return v
}

// SetSeparator sets the separator text between items.
func (v *VNode) SetSeparator(separator string) *VNode {
	v.separator = separator
	return v
}

// SetMaxWidth sets an optional preferred maximum width.
func (v *VNode) SetMaxWidth(width int) *VNode {
	v.maxWidth = width
	return v
}

// SetItemStyle sets the style for non-current breadcrumb items.
func (v *VNode) SetItemStyle(s style.Style) *VNode {
	v.itemStyle = s
	return v
}

// SetCurrentStyle sets the style for the current breadcrumb item.
func (v *VNode) SetCurrentStyle(s style.Style) *VNode {
	v.currentStyle = s
	return v
}

// SetSeparatorStyle sets the style for separator text.
func (v *VNode) SetSeparatorStyle(s style.Style) *VNode {
	v.separatorStyle = s
	return v
}

// Items returns the configured breadcrumb items.
func (v *VNode) Items() []Item { return cloneItems(v.items) }

// Separator returns the separator text.
func (v *VNode) Separator() string { return v.separator }

// MaxWidth returns the preferred maximum width.
func (v *VNode) MaxWidth() int { return v.maxWidth }

// ItemStyle returns the style for non-current items.
func (v *VNode) ItemStyle() style.Style { return v.itemStyle }

// CurrentStyle returns the style for the active item.
func (v *VNode) CurrentStyle() style.Style { return v.currentStyle }

// SeparatorStyle returns the style for separator text.
func (v *VNode) SeparatorStyle() style.Style { return v.separatorStyle }

func cloneItems(items []Item) []Item {
	if len(items) == 0 {
		return nil
	}
	cloned := make([]Item, len(items))
	copy(cloned, items)
	return cloned
}
