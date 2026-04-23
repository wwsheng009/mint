package anchor

import (
	"strings"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

const (
	propActiveKey           = "activeKey"
	propActiveKeyControlled = "activeKeyControlled"
	propChangeIntent        = "changeIntent"
	propComponentID         = "componentID"
	propFormID              = "formID"
	propItems               = "items"
	propKey                 = "key"
	propShowBorder          = "showBorder"
	propCurrentStyle        = "currentStyle"
	propStyle               = "style"
	propTitle               = "title"
	propViewportHeight      = "viewportHeight"
	propWidth               = "width"
)

// Item describes a single anchor node.
type Item struct {
	Key      string
	Title    string
	Href     string
	Disabled bool
	Children []Item
}

// VNode is the declarative description of an Anchor component.
type VNode struct {
	*rtui.ElementVNode

	key                 string
	componentID         string
	title               string
	items               []Item
	activeKey           string
	activeKeyControlled bool
	viewportHeight      int
	width               int
	showBorder          bool
	anchorStyle         style.Style
	currentStyle        style.Style
	changeIntent        intent.Intent
	formID              string
}

var (
	_ rtui.VNode           = (*VNode)(nil)
	_ rtui.InstanceFactory = (*VNode)(nil)
)

// New creates a new Anchor VNode.
func New() *VNode {
	return &VNode{
		ElementVNode:        rtui.NewElement("anchor"),
		activeKeyControlled: false,
		showBorder:          false,
		viewportHeight:      0,
	}
}

// NewItem creates an anchor item with optional children.
func NewItem(key, title string, children ...Item) Item {
	return Item{
		Key:      key,
		Title:    title,
		Children: append([]Item(nil), children...),
	}
}

func (v *VNode) Key() string { return v.key }

func (v *VNode) SetKey(key string) rtui.VNode {
	v.key = key
	return v
}

// ID returns the explicit business ID, or falls back to key.
func (v *VNode) ID() string {
	if id := v.ElementVNode.ID(); id != "" {
		return id
	}
	return v.key
}

func (v *VNode) SetID(id string) rtui.VNode {
	v.ElementVNode.SetID(id)
	return v
}

func (v *VNode) Tag() string { return "anchor" }

func (v *VNode) Style() style.Style { return v.anchorStyle }

func (v *VNode) SetStyle(s style.Style) rtui.VNode {
	v.anchorStyle = s
	return v
}

func (v *VNode) Children() []rtui.VNode { return nil }

func (v *VNode) SetChildren(children []rtui.VNode) rtui.VNode { return v }

func (v *VNode) GetLayer() rtui.Layer { return rtui.LayerBase }

func (v *VNode) SetLayer(l rtui.Layer) rtui.VNode { return v }

func (v *VNode) Props() rtui.Props {
	return rtui.Props{
		propActiveKey:           v.activeKey,
		propActiveKeyControlled: v.activeKeyControlled,
		propChangeIntent:        v.changeIntent,
		propComponentID:         v.componentID,
		propCurrentStyle:        v.currentStyle,
		propFormID:              v.formID,
		propItems:               cloneItems(v.items),
		propKey:                 v.key,
		propShowBorder:          v.showBorder,
		propStyle:               v.anchorStyle,
		propTitle:               v.title,
		propViewportHeight:      v.viewportHeight,
		propWidth:               v.width,
	}
}

func (v *VNode) SetProps(props rtui.Props) rtui.VNode {
	if key, ok := props[propKey].(string); ok {
		v.key = key
	}
	if componentID, ok := props[propComponentID].(string); ok {
		v.componentID = componentID
	}
	if title, ok := props[propTitle].(string); ok {
		v.title = title
	}
	if items, ok := props[propItems].([]Item); ok {
		v.items = cloneItems(items)
	}
	if activeKey, ok := props[propActiveKey].(string); ok {
		v.activeKey = activeKey
	}
	if controlled, ok := props[propActiveKeyControlled].(bool); ok {
		v.activeKeyControlled = controlled
	}
	if viewportHeight, ok := props[propViewportHeight].(int); ok {
		v.viewportHeight = viewportHeight
	}
	if width, ok := props[propWidth].(int); ok {
		v.width = width
	}
	if showBorder, ok := props[propShowBorder].(bool); ok {
		v.showBorder = showBorder
	}
	if anchorStyle, ok := props[propStyle].(style.Style); ok {
		v.anchorStyle = anchorStyle
	}
	if currentStyle, ok := props[propCurrentStyle].(style.Style); ok {
		v.currentStyle = currentStyle
	}
	if changeIntent, ok := props[propChangeIntent].(intent.Intent); ok {
		v.changeIntent = changeIntent
	}
	if formID, ok := props[propFormID].(string); ok {
		v.formID = formID
	}
	return v
}

func (v *VNode) CreateInstance() rtui.ComponentInstance {
	return NewInstance(v.Props())
}

func (v *VNode) SetComponentID(id string) *VNode {
	v.componentID = id
	return v
}

func (v *VNode) SetTitle(title string) *VNode {
	v.title = title
	return v
}

func (v *VNode) SetItems(items []Item) *VNode {
	v.items = cloneItems(items)
	return v
}

func (v *VNode) AddItem(item Item) *VNode {
	v.items = append(v.items, item)
	return v
}

func (v *VNode) SetActiveKey(key string) *VNode {
	v.activeKey = key
	v.activeKeyControlled = true
	return v
}

func (v *VNode) SetInitialActiveKey(key string) *VNode {
	v.activeKey = key
	v.activeKeyControlled = false
	return v
}

func (v *VNode) SetViewportHeight(height int) *VNode {
	v.viewportHeight = height
	return v
}

func (v *VNode) SetWidth(width int) *VNode {
	v.width = width
	return v
}

func (v *VNode) SetShowBorder(show bool) *VNode {
	v.showBorder = show
	return v
}

func (v *VNode) SetStyleProps(s style.Style) *VNode {
	v.anchorStyle = s
	return v
}

func (v *VNode) SetCurrentStyle(s style.Style) *VNode {
	v.currentStyle = s
	return v
}

func (v *VNode) SetChangeIntent(changeIntent intent.Intent) *VNode {
	v.changeIntent = changeIntent
	return v
}

func (v *VNode) SetFormID(formID string) *VNode {
	v.formID = formID
	return v
}

func cloneItems(items []Item) []Item {
	if len(items) == 0 {
		return nil
	}
	cloned := make([]Item, 0, len(items))
	for _, item := range items {
		normalized, ok := normalizeItem(item)
		if !ok {
			continue
		}
		cloned = append(cloned, normalized)
	}
	if len(cloned) == 0 {
		return nil
	}
	return cloned
}

func normalizeItem(item Item) (Item, bool) {
	key := strings.TrimSpace(item.Key)
	href := strings.TrimSpace(item.Href)
	title := strings.TrimSpace(item.Title)

	if strings.HasPrefix(key, "#") {
		if href == "" {
			href = key
		}
		key = strings.TrimPrefix(key, "#")
	}
	if key == "" && href != "" {
		key = strings.TrimPrefix(href, "#")
	}
	if key == "" {
		key = title
	}
	if key == "" {
		return Item{}, false
	}
	if title == "" {
		title = key
	}
	if href == "" {
		href = "#" + key
	}

	children := cloneItems(item.Children)
	return Item{
		Key:      key,
		Title:    title,
		Href:     href,
		Disabled: item.Disabled,
		Children: children,
	}, true
}
