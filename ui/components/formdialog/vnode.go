// Package formdialog provides a Fiber-first modal form composition.
package formdialog

import (
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/button"
	"github.com/wwsheng009/mint/ui/components/form"
)

const (
	propCancelIntent    = "cancelIntent"
	propCancelText      = "cancelText"
	propChildren        = "children"
	propCloseIntent     = "closeIntent"
	propCloseOnBackdrop = "closeOnBackdrop"
	propCloseOnEsc      = "closeOnEsc"
	propCloseable       = "closeable"
	propDescription     = "description"
	propDisabledReason  = "disabledReason"
	propFormID          = "formID"
	propHeight          = "height"
	propKey             = "key"
	propLayout          = "layout"
	propOpen            = "open"
	propStyle           = "style"
	propSubmitDisabled  = "submitDisabled"
	propSubmitIntent    = "submitIntent"
	propSubmitText      = "submitText"
	propSubmitVariant   = "submitVariant"
	propTitle           = "title"
	propValidateAll     = "validateAll"
	propValues          = "values"
	propWidth           = "width"
)

// VNode is the declarative description of a modal form dialog.
type VNode struct {
	*rtui.ElementVNode

	key             string
	title           string
	description     string
	open            bool
	width           int
	height          int
	formID          string
	layout          form.FormLayout
	values          map[string]interface{}
	validateAll     bool
	children        []rtui.VNode
	submitText      string
	cancelText      string
	submitVariant   button.Variant
	submitDisabled  bool
	disabledReason  string
	submitIntent    intent.Intent
	cancelIntent    intent.Intent
	closeIntent     intent.Intent
	closeable       bool
	closeOnEsc      bool
	closeOnBackdrop bool
	rootStyle       style.Style
}

var (
	_ rtui.VNode           = (*VNode)(nil)
	_ rtui.InstanceFactory = (*VNode)(nil)
)

// New creates a FormDialog VNode.
func New() *VNode {
	return &VNode{
		ElementVNode:    rtui.NewElement("formdialog"),
		width:           72,
		height:          18,
		layout:          form.LayoutVertical,
		validateAll:     true,
		values:          map[string]interface{}{},
		submitText:      "Submit",
		cancelText:      "Cancel",
		submitVariant:   button.VariantPrimary,
		closeable:       true,
		closeOnEsc:      true,
		closeOnBackdrop: true,
	}
}

func (v *VNode) Key() string { return v.key }

func (v *VNode) SetKey(key string) rtui.VNode {
	v.key = key
	return v
}

// ID returns the explicit business ID, or falls back to the key.
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

func (v *VNode) Tag() string { return "formdialog" }

func (v *VNode) Style() style.Style { return v.rootStyle }

func (v *VNode) SetStyle(s style.Style) rtui.VNode {
	v.rootStyle = s
	return v
}

func (v *VNode) Children() []rtui.VNode {
	return nil
}

func (v *VNode) SetChildren(children []rtui.VNode) rtui.VNode {
	v.children = cloneChildren(children)
	return v
}

func (v *VNode) GetLayer() rtui.Layer { return rtui.LayerBase }

func (v *VNode) SetLayer(l rtui.Layer) rtui.VNode { return v }

func (v *VNode) Props() rtui.Props {
	return rtui.Props{
		propCancelIntent:    v.cancelIntent,
		propCancelText:      v.cancelText,
		propChildren:        cloneChildren(v.children),
		propCloseIntent:     v.closeIntent,
		propCloseOnBackdrop: v.closeOnBackdrop,
		propCloseOnEsc:      v.closeOnEsc,
		propCloseable:       v.closeable,
		propDescription:     v.description,
		propDisabledReason:  v.disabledReason,
		propFormID:          v.formID,
		propHeight:          v.height,
		propKey:             v.key,
		propLayout:          v.layout,
		propOpen:            v.open,
		propStyle:           v.rootStyle,
		propSubmitDisabled:  v.submitDisabled,
		propSubmitIntent:    v.submitIntent,
		propSubmitText:      v.submitText,
		propSubmitVariant:   v.submitVariant,
		propTitle:           v.title,
		propValidateAll:     v.validateAll,
		propValues:          cloneValues(v.values),
		propWidth:           v.width,
	}
}

func (v *VNode) SetProps(props rtui.Props) rtui.VNode {
	v.key = getStringProp(props, propKey, v.key)
	v.title = getStringProp(props, propTitle, v.title)
	v.description = getStringProp(props, propDescription, v.description)
	v.open = getBoolProp(props, propOpen, v.open)
	v.width = getIntProp(props, propWidth, v.width)
	v.height = getIntProp(props, propHeight, v.height)
	v.formID = getStringProp(props, propFormID, v.formID)
	v.layout = getLayoutProp(props, propLayout, v.layout)
	v.values = getValuesProp(props, propValues, v.values)
	v.validateAll = getBoolProp(props, propValidateAll, v.validateAll)
	v.children = getChildrenProp(props, propChildren, v.children)
	v.submitText = getStringProp(props, propSubmitText, v.submitText)
	v.cancelText = getStringProp(props, propCancelText, v.cancelText)
	v.submitVariant = getButtonVariantProp(props, propSubmitVariant, v.submitVariant)
	v.submitDisabled = getBoolProp(props, propSubmitDisabled, v.submitDisabled)
	v.disabledReason = getStringProp(props, propDisabledReason, v.disabledReason)
	v.submitIntent = getIntentProp(props, propSubmitIntent, v.submitIntent)
	v.cancelIntent = getIntentProp(props, propCancelIntent, v.cancelIntent)
	v.closeIntent = getIntentProp(props, propCloseIntent, v.closeIntent)
	v.closeable = getBoolProp(props, propCloseable, v.closeable)
	v.closeOnEsc = getBoolProp(props, propCloseOnEsc, v.closeOnEsc)
	v.closeOnBackdrop = getBoolProp(props, propCloseOnBackdrop, v.closeOnBackdrop)
	v.rootStyle = getStyleProp(props, propStyle, v.rootStyle)
	v.normalize()
	return v
}

func (v *VNode) CreateInstance() rtui.ComponentInstance {
	return NewInstance(v.Props())
}

func (v *VNode) SetTitle(title string) *VNode {
	v.title = title
	return v
}

func (v *VNode) SetDescription(description string) *VNode {
	v.description = description
	return v
}

func (v *VNode) SetOpen(open bool) *VNode {
	v.open = open
	return v
}

func (v *VNode) SetWidth(width int) *VNode {
	v.width = width
	v.normalize()
	return v
}

func (v *VNode) SetHeight(height int) *VNode {
	v.height = height
	v.normalize()
	return v
}

func (v *VNode) SetFormID(formID string) *VNode {
	v.formID = formID
	return v
}

func (v *VNode) SetLayout(layout form.FormLayout) *VNode {
	v.layout = normalizeLayout(layout)
	return v
}

func (v *VNode) SetValues(values map[string]interface{}) *VNode {
	v.values = cloneValues(values)
	return v
}

func (v *VNode) SetValue(field string, value interface{}) *VNode {
	if v.values == nil {
		v.values = map[string]interface{}{}
	}
	v.values[field] = value
	return v
}

func (v *VNode) SetValidateAll(validate bool) *VNode {
	v.validateAll = validate
	return v
}

func (v *VNode) AddChild(child rtui.VNode) *VNode {
	if child != nil {
		v.children = append(v.children, child)
	}
	return v
}

func (v *VNode) AddChildren(children ...rtui.VNode) *VNode {
	for _, child := range children {
		v.AddChild(child)
	}
	return v
}

func (v *VNode) SetSubmitText(text string) *VNode {
	v.submitText = text
	return v
}

func (v *VNode) SetCancelText(text string) *VNode {
	v.cancelText = text
	return v
}

func (v *VNode) SetSubmitVariant(variant button.Variant) *VNode {
	v.submitVariant = variant
	return v
}

func (v *VNode) SetSubmitDisabled(disabled bool) *VNode {
	v.submitDisabled = disabled
	return v
}

func (v *VNode) SetDisabledReason(reason string) *VNode {
	v.disabledReason = reason
	return v
}

func (v *VNode) SetSubmitIntent(i intent.Intent) *VNode {
	v.submitIntent = i
	return v
}

func (v *VNode) SetCancelIntent(i intent.Intent) *VNode {
	v.cancelIntent = i
	return v
}

func (v *VNode) SetCloseIntent(i intent.Intent) *VNode {
	v.closeIntent = i
	return v
}

func (v *VNode) SetCloseable(closeable bool) *VNode {
	v.closeable = closeable
	return v
}

func (v *VNode) SetCloseOnEsc(closeOnEsc bool) *VNode {
	v.closeOnEsc = closeOnEsc
	return v
}

func (v *VNode) SetCloseOnBackdrop(closeOnBackdrop bool) *VNode {
	v.closeOnBackdrop = closeOnBackdrop
	return v
}

func (v *VNode) SetRootStyle(s style.Style) *VNode {
	v.rootStyle = s
	return v
}

func (v *VNode) normalize() {
	if v.width < 0 {
		v.width = 0
	}
	if v.height < 0 {
		v.height = 0
	}
	v.layout = normalizeLayout(v.layout)
	if v.values == nil {
		v.values = map[string]interface{}{}
	}
}
