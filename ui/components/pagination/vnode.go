package pagination

import (
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

const (
	propComponentID           = "componentID"
	propCurrentPage           = "currentPage"
	propCurrentPageControlled = "currentPageControlled"
	propDisabled              = "disabled"
	propDisabledStyle         = "disabledStyle"
	propKey                   = "key"
	propMaxButtons            = "maxButtons"
	propPageIntent            = "pageIntent"
	propPageIntentField       = "pageIntentField"
	propPageSize              = "pageSize"
	propPaginationStyle       = "paginationStyle"
	propSelectedStyle         = "selectedStyle"
	propShowTotal             = "showTotal"
	propTotal                 = "total"
)

// VNode is the immutable description of a Pagination component.
type VNode struct {
	*rtui.ElementVNode

	key                   string
	componentID           string
	total                 int
	pageSize              int
	currentPage           int
	currentPageControlled bool
	maxButtons            int
	showTotal             bool
	disabled              bool
	paginationStyle       style.Style
	selectedStyle         style.Style
	disabledStyle         style.Style
	pageIntent            intent.Intent
	pageIntentField       intent.FieldIntent
}

var (
	_ rtui.VNode           = (*VNode)(nil)
	_ rtui.InstanceFactory = (*VNode)(nil)
)

func New() *VNode {
	return &VNode{
		ElementVNode:    rtui.NewElement("pagination"),
		pageSize:        10,
		currentPage:     0,
		maxButtons:      5,
		showTotal:       true,
		paginationStyle: style.Style{},
		selectedStyle:   style.Style{}.Reverse(true).Bold(true),
		disabledStyle:   style.Style{}.Foreground(style.BrightBlack),
	}
}

func (v *VNode) Key() string { return v.key }

func (v *VNode) SetKey(key string) rtui.VNode {
	v.key = key
	return v
}

func (v *VNode) Tag() string { return "pagination" }

func (v *VNode) Type() rtui.VNodeType { return rtui.VNodeElement }

func (v *VNode) Children() []rtui.VNode { return nil }

func (v *VNode) SetChildren(children []rtui.VNode) rtui.VNode { return v }

func (v *VNode) GetLayer() rtui.Layer { return rtui.LayerBase }

func (v *VNode) SetLayer(l rtui.Layer) rtui.VNode { return v }

func (v *VNode) Style() style.Style { return v.paginationStyle }

func (v *VNode) SetStyle(s style.Style) rtui.VNode {
	v.paginationStyle = s
	return v
}

func (v *VNode) TextContent() string { return "" }

func (v *VNode) Props() rtui.Props {
	return rtui.Props{
		propKey:                   v.key,
		propComponentID:           v.componentID,
		propTotal:                 v.total,
		propPageSize:              v.pageSize,
		propCurrentPage:           v.currentPage,
		propCurrentPageControlled: v.currentPageControlled,
		propMaxButtons:            v.maxButtons,
		propShowTotal:             v.showTotal,
		propDisabled:              v.disabled,
		propPaginationStyle:       v.paginationStyle,
		propSelectedStyle:         v.selectedStyle,
		propDisabledStyle:         v.disabledStyle,
		propPageIntent:            v.pageIntent,
		propPageIntentField:       v.pageIntentField,
	}
}

func (v *VNode) SetProps(props rtui.Props) rtui.VNode {
	if value, ok := props[propKey].(string); ok {
		v.key = value
	}
	if value, ok := props[propComponentID].(string); ok {
		v.componentID = value
	}
	if value, ok := props[propTotal].(int); ok {
		v.total = value
	}
	if value, ok := props[propPageSize].(int); ok {
		v.pageSize = value
	}
	if value, ok := props[propCurrentPage].(int); ok {
		v.currentPage = value
	}
	if value, ok := props[propCurrentPageControlled].(bool); ok {
		v.currentPageControlled = value
	}
	if value, ok := props[propMaxButtons].(int); ok {
		v.maxButtons = value
	}
	if value, ok := props[propShowTotal].(bool); ok {
		v.showTotal = value
	}
	if value, ok := props[propDisabled].(bool); ok {
		v.disabled = value
	}
	if value, ok := props[propPaginationStyle].(style.Style); ok {
		v.paginationStyle = value
	}
	if value, ok := props[propSelectedStyle].(style.Style); ok {
		v.selectedStyle = value
	}
	if value, ok := props[propDisabledStyle].(style.Style); ok {
		v.disabledStyle = value
	}
	if value, ok := props[propPageIntent].(intent.Intent); ok {
		v.pageIntent = value
	}
	if value, ok := props[propPageIntentField].(intent.FieldIntent); ok {
		v.pageIntentField = value
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

func (v *VNode) SetTotal(total int) *VNode {
	v.total = total
	return v
}

func (v *VNode) SetPageSize(pageSize int) *VNode {
	v.pageSize = pageSize
	return v
}

func (v *VNode) SetCurrentPage(page int) *VNode {
	v.currentPage = page
	v.currentPageControlled = true
	return v
}

func (v *VNode) SetMaxButtons(maxButtons int) *VNode {
	v.maxButtons = maxButtons
	return v
}

func (v *VNode) SetShowTotal(show bool) *VNode {
	v.showTotal = show
	return v
}

func (v *VNode) SetDisabled(disabled bool) *VNode {
	v.disabled = disabled
	return v
}

func (v *VNode) SetSelectedStyle(s style.Style) *VNode {
	v.selectedStyle = s
	return v
}

func (v *VNode) SetDisabledStyle(s style.Style) *VNode {
	v.disabledStyle = s
	return v
}

func (v *VNode) SetPageIntent(i intent.Intent) *VNode {
	v.pageIntent = i
	return v
}

func (v *VNode) SetPageIntentField(i intent.FieldIntent) *VNode {
	v.pageIntentField = i
	return v
}
