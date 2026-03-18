package timeline

import (
	"fmt"

	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

const (
	propContentStyle = "contentStyle"
	propItems        = "items"
	propKey          = "key"
	propLabelStyle   = "labelStyle"
	propLineStyle    = "lineStyle"
	propPending      = "pending"
	propPendingStyle = "pendingStyle"
	propReverse      = "reverse"
	propStyle        = "style"
	propWidth        = "width"
)

// Status controls the timeline marker style.
type Status int

const (
	StatusDefault Status = iota
	StatusSuccess
	StatusWarning
	StatusError
	StatusPending
)

// Item describes one timeline entry.
type Item struct {
	Key         string
	Label       string
	Content     string
	Description string
	Dot         string
	Status      Status
	Color       style.Color
}

// Event creates a timeline item with content text.
func Event(content string) Item {
	return Item{Content: content}
}

// WithKey sets the item key.
func (i Item) WithKey(key string) Item {
	i.Key = key
	return i
}

// WithLabel sets the item label text.
func (i Item) WithLabel(label string) Item {
	i.Label = label
	return i
}

// WithDescription sets the item description text.
func (i Item) WithDescription(description string) Item {
	i.Description = description
	return i
}

// WithDot sets a custom marker glyph.
func (i Item) WithDot(dot string) Item {
	i.Dot = dot
	return i
}

// WithStatus sets the item status.
func (i Item) WithStatus(status Status) Item {
	i.Status = status
	return i
}

// WithColor sets an explicit marker color override.
func (i Item) WithColor(color style.Color) Item {
	i.Color = color
	return i
}

// VNode is the declarative description of a Timeline component.
type VNode struct {
	*rtui.ElementVNode

	key          string
	items        []Item
	pending      string
	reverse      bool
	width        int
	rootStyle    style.Style
	labelStyle   style.Style
	contentStyle style.Style
	pendingStyle style.Style
	lineStyle    style.Style
}

var (
	_ rtui.VNode           = (*VNode)(nil)
	_ rtui.InstanceFactory = (*VNode)(nil)
)

// New creates a new Timeline VNode.
func New(items []Item) *VNode {
	return &VNode{
		ElementVNode: rtui.NewElement("timeline"),
		items:        normalizeItems(items),
	}
}

func (v *VNode) Key() string { return v.key }

func (v *VNode) SetKey(key string) rtui.VNode {
	v.key = key
	return v
}

func (v *VNode) Tag() string { return "timeline" }

func (v *VNode) Style() style.Style { return v.rootStyle }

func (v *VNode) SetStyle(s style.Style) rtui.VNode {
	v.rootStyle = s
	return v
}

func (v *VNode) Children() []rtui.VNode { return nil }

func (v *VNode) SetChildren(children []rtui.VNode) rtui.VNode { return v }

func (v *VNode) GetLayer() rtui.Layer { return rtui.LayerBase }

func (v *VNode) SetLayer(l rtui.Layer) rtui.VNode { return v }

func (v *VNode) Props() rtui.Props {
	return rtui.Props{
		propContentStyle: v.contentStyle,
		propItems:        cloneItems(v.items),
		propKey:          v.key,
		propLabelStyle:   v.labelStyle,
		propLineStyle:    v.lineStyle,
		propPending:      v.pending,
		propPendingStyle: v.pendingStyle,
		propReverse:      v.reverse,
		propStyle:        v.rootStyle,
		propWidth:        v.width,
	}
}

func (v *VNode) SetProps(props rtui.Props) rtui.VNode {
	if key, ok := props[propKey].(string); ok {
		v.key = key
	}
	if items, ok := props[propItems].([]Item); ok {
		v.items = normalizeItems(items)
	}
	if pending, ok := props[propPending].(string); ok {
		v.pending = pending
	}
	if reverse, ok := props[propReverse].(bool); ok {
		v.reverse = reverse
	}
	if width, ok := props[propWidth].(int); ok {
		v.width = width
	}
	if s, ok := props[propStyle].(style.Style); ok {
		v.rootStyle = s
	}
	if s, ok := props[propLabelStyle].(style.Style); ok {
		v.labelStyle = s
	}
	if s, ok := props[propContentStyle].(style.Style); ok {
		v.contentStyle = s
	}
	if s, ok := props[propPendingStyle].(style.Style); ok {
		v.pendingStyle = s
	}
	if s, ok := props[propLineStyle].(style.Style); ok {
		v.lineStyle = s
	}
	return v
}

func (v *VNode) CreateInstance() rtui.ComponentInstance {
	return NewInstance(v.Props())
}

// SetItems replaces all entries.
func (v *VNode) SetItems(items []Item) *VNode {
	v.items = normalizeItems(items)
	return v
}

// AddItem appends an entry.
func (v *VNode) AddItem(item Item) *VNode {
	v.items = normalizeItems(append(v.items, item))
	return v
}

// SetPending sets the pending tail content.
func (v *VNode) SetPending(pending string) *VNode {
	v.pending = pending
	return v
}

// SetReverse toggles reverse order rendering.
func (v *VNode) SetReverse(reverse bool) *VNode {
	v.reverse = reverse
	return v
}

// SetWidth sets the preferred width.
func (v *VNode) SetWidth(width int) *VNode {
	v.width = width
	return v
}

// SetLabelStyle sets the label style.
func (v *VNode) SetLabelStyle(s style.Style) *VNode {
	v.labelStyle = s
	return v
}

// SetContentStyle sets the content style.
func (v *VNode) SetContentStyle(s style.Style) *VNode {
	v.contentStyle = s
	return v
}

// SetPendingStyle sets the pending content style.
func (v *VNode) SetPendingStyle(s style.Style) *VNode {
	v.pendingStyle = s
	return v
}

// SetLineStyle sets the connector line style.
func (v *VNode) SetLineStyle(s style.Style) *VNode {
	v.lineStyle = s
	return v
}

func cloneItems(items []Item) []Item {
	if len(items) == 0 {
		return nil
	}
	cloned := make([]Item, len(items))
	copy(cloned, items)
	return cloned
}

func normalizeItems(items []Item) []Item {
	if len(items) == 0 {
		return nil
	}
	cloned := cloneItems(items)
	for index := range cloned {
		if cloned[index].Key == "" {
			cloned[index].Key = fmt.Sprintf("item-%d", index)
		}
		if cloned[index].Dot == "" && cloned[index].Status == StatusPending {
			cloned[index].Dot = "○"
		}
	}
	return cloned
}
