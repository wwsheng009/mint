package descriptions

import (
	"fmt"
	"strings"

	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	textcomp "github.com/wwsheng009/mint/ui/components/text"
)

const (
	propBordered     = "bordered"
	propColon        = "colon"
	propColumn       = "column"
	propContentStyle = "contentStyle"
	propExtra        = "extra"
	propEmptyText    = "emptyText"
	propItems        = "items"
	propKey          = "key"
	propContentWidth = "contentWidth"
	propLabelWidth   = "labelWidth"
	propLabelStyle   = "labelStyle"
	propLayout       = "layout"
	propMaskText     = "maskText"
	propStyle        = "style"
	propTitle        = "title"
	propTitleStyle   = "titleStyle"
	propWidth        = "width"
)

// Layout controls how each description item is rendered.
type Layout int

const (
	LayoutHorizontal Layout = iota
	LayoutVertical
)

// Item describes a label/content pair in Descriptions.
type Item struct {
	Key          string
	Label        string
	Content      rtui.VNode
	Value        interface{}
	HasValue     bool
	Span         int
	LabelWidth   int
	ContentWidth int
	EmptyText    string
	Sensitive    bool
	MaskText     string
	LabelStyle   style.Style
	ContentStyle style.Style
}

// Entry creates a descriptions item from label and content.
func Entry(label string, content rtui.VNode) Item {
	return Item{
		Label:   label,
		Content: content,
		Span:    1,
	}
}

// Field creates a text descriptions item.
func Field(label, value string) Item {
	item := Entry(label, textcomp.New(value))
	item.Value = value
	item.HasValue = true
	return item
}

// Value creates a descriptions item from an arbitrary value.
func Value(label string, value interface{}) Item {
	return Item{
		Label:    label,
		Value:    value,
		HasValue: true,
		Span:     1,
	}
}

// SensitiveField creates a masked descriptions item for secrets and tokens.
func SensitiveField(label string, value interface{}) Item {
	return Value(label, value).WithSensitive(true)
}

// WithKey sets the item key.
func (i Item) WithKey(key string) Item {
	i.Key = key
	return i
}

// WithSpan sets the column span.
func (i Item) WithSpan(span int) Item {
	i.Span = span
	return i
}

// WithLabelWidth sets a fixed label width for this item.
func (i Item) WithLabelWidth(width int) Item {
	i.LabelWidth = width
	return i
}

// WithContentWidth sets a fixed content width for this item.
func (i Item) WithContentWidth(width int) Item {
	i.ContentWidth = width
	return i
}

// WithEmptyText sets the placeholder used when this item has an empty value.
func (i Item) WithEmptyText(emptyText string) Item {
	i.EmptyText = emptyText
	return i
}

// WithSensitive toggles masking for this item.
func (i Item) WithSensitive(sensitive bool) Item {
	i.Sensitive = sensitive
	return i
}

// WithMaskText sets the masked display text for this item.
func (i Item) WithMaskText(maskText string) Item {
	i.MaskText = maskText
	return i
}

// WithLabelStyle sets the item label style.
func (i Item) WithLabelStyle(s style.Style) Item {
	i.LabelStyle = s
	return i
}

// WithContentStyle sets the item content style.
func (i Item) WithContentStyle(s style.Style) Item {
	i.ContentStyle = s
	return i
}

// VNode is the declarative description of a Descriptions component.
type VNode struct {
	*rtui.ElementVNode

	key          string
	title        string
	extra        rtui.VNode
	items        []Item
	column       int
	bordered     bool
	colon        bool
	layout       Layout
	width        int
	labelWidth   int
	contentWidth int
	emptyText    string
	maskText     string
	rootStyle    style.Style
	titleStyle   style.Style
	labelStyle   style.Style
	contentStyle style.Style
}

var (
	_ rtui.VNode           = (*VNode)(nil)
	_ rtui.InstanceFactory = (*VNode)(nil)
)

// New creates a new Descriptions VNode.
func New(items []Item) *VNode {
	return &VNode{
		ElementVNode: rtui.NewElement("descriptions"),
		items:        normalizeItems(items),
		column:       3,
		bordered:     false,
		colon:        true,
		layout:       LayoutHorizontal,
		emptyText:    "-",
		maskText:     "****",
	}
}

func (v *VNode) Key() string { return v.key }

func (v *VNode) SetKey(key string) rtui.VNode {
	v.key = key
	return v
}

func (v *VNode) Tag() string { return "descriptions" }

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
		propBordered:     v.bordered,
		propColon:        v.colon,
		propColumn:       v.column,
		propContentStyle: v.contentStyle,
		propContentWidth: v.contentWidth,
		propEmptyText:    v.emptyText,
		propExtra:        v.extra,
		propItems:        cloneItems(v.items),
		propKey:          v.key,
		propLabelWidth:   v.labelWidth,
		propLabelStyle:   v.labelStyle,
		propLayout:       v.layout,
		propMaskText:     v.maskText,
		propStyle:        v.rootStyle,
		propTitle:        v.title,
		propTitleStyle:   v.titleStyle,
		propWidth:        v.width,
	}
}

func (v *VNode) SetProps(props rtui.Props) rtui.VNode {
	if key, ok := props[propKey].(string); ok {
		v.key = key
	}
	if title, ok := props[propTitle].(string); ok {
		v.title = title
	}
	if extra, ok := props[propExtra].(rtui.VNode); ok {
		v.extra = extra
	}
	if items, ok := props[propItems].([]Item); ok {
		v.items = normalizeItems(items)
	}
	if column, ok := props[propColumn].(int); ok {
		v.column = column
	}
	if bordered, ok := props[propBordered].(bool); ok {
		v.bordered = bordered
	}
	if colon, ok := props[propColon].(bool); ok {
		v.colon = colon
	}
	if layout, ok := props[propLayout].(Layout); ok {
		v.layout = layout
	}
	if width, ok := props[propWidth].(int); ok {
		v.width = width
	}
	if labelWidth, ok := props[propLabelWidth].(int); ok {
		v.labelWidth = labelWidth
	}
	if contentWidth, ok := props[propContentWidth].(int); ok {
		v.contentWidth = contentWidth
	}
	if emptyText, ok := props[propEmptyText].(string); ok {
		v.emptyText = emptyText
	}
	if maskText, ok := props[propMaskText].(string); ok {
		v.maskText = maskText
	}
	if s, ok := props[propStyle].(style.Style); ok {
		v.rootStyle = s
	}
	if s, ok := props[propTitleStyle].(style.Style); ok {
		v.titleStyle = s
	}
	if s, ok := props[propLabelStyle].(style.Style); ok {
		v.labelStyle = s
	}
	if s, ok := props[propContentStyle].(style.Style); ok {
		v.contentStyle = s
	}
	return v
}

func (v *VNode) CreateInstance() rtui.ComponentInstance {
	return NewInstance(v.Props())
}

// SetTitle sets the optional title text.
func (v *VNode) SetTitle(title string) *VNode {
	v.title = title
	return v
}

// SetExtra sets the optional header-side node.
func (v *VNode) SetExtra(extra rtui.VNode) *VNode {
	v.extra = extra
	return v
}

// SetItems replaces all items.
func (v *VNode) SetItems(items []Item) *VNode {
	v.items = normalizeItems(items)
	return v
}

// AddItem appends an item.
func (v *VNode) AddItem(item Item) *VNode {
	v.items = normalizeItems(append(v.items, item))
	return v
}

// SetColumn sets the target column count.
func (v *VNode) SetColumn(column int) *VNode {
	v.column = column
	return v
}

// SetBordered toggles the bordered presentation.
func (v *VNode) SetBordered(bordered bool) *VNode {
	v.bordered = bordered
	return v
}

// SetColon toggles label colon rendering.
func (v *VNode) SetColon(colon bool) *VNode {
	v.colon = colon
	return v
}

// SetLayout sets the item layout mode.
func (v *VNode) SetLayout(layout Layout) *VNode {
	v.layout = layout
	return v
}

// SetWidth sets a preferred width.
func (v *VNode) SetWidth(width int) *VNode {
	v.width = width
	return v
}

// SetLabelWidth sets the default fixed label width.
func (v *VNode) SetLabelWidth(width int) *VNode {
	v.labelWidth = width
	return v
}

// SetContentWidth sets the default fixed content width.
func (v *VNode) SetContentWidth(width int) *VNode {
	v.contentWidth = width
	return v
}

// SetEmptyText sets the placeholder for empty values.
func (v *VNode) SetEmptyText(emptyText string) *VNode {
	v.emptyText = emptyText
	return v
}

// SetMaskText sets the default masked text for sensitive values.
func (v *VNode) SetMaskText(maskText string) *VNode {
	v.maskText = maskText
	return v
}

// SetTitleStyle sets the title style.
func (v *VNode) SetTitleStyle(s style.Style) *VNode {
	v.titleStyle = s
	return v
}

// SetLabelStyle sets the default label style.
func (v *VNode) SetLabelStyle(s style.Style) *VNode {
	v.labelStyle = s
	return v
}

// SetContentStyle sets the default content style.
func (v *VNode) SetContentStyle(s style.Style) *VNode {
	v.contentStyle = s
	return v
}

// Items returns the configured items.
func (v *VNode) Items() []Item { return cloneItems(v.items) }

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
	seen := make(map[string]int, len(cloned))
	for index := range cloned {
		key := strings.TrimSpace(cloned[index].Key)
		if key == "" {
			key = fmt.Sprintf("item-%d", index)
		}
		base := key
		if count, exists := seen[base]; exists {
			count++
			seen[base] = count
			key = fmt.Sprintf("%s-%d", base, count)
		} else {
			seen[base] = 0
		}
		cloned[index].Key = key
		if cloned[index].Span < 1 {
			cloned[index].Span = 1
		}
		if cloned[index].LabelWidth < 0 {
			cloned[index].LabelWidth = 0
		}
		if cloned[index].ContentWidth < 0 {
			cloned[index].ContentWidth = 0
		}
	}
	return cloned
}
