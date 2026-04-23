package popconfirm

import (
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/button"
	"github.com/wwsheng009/mint/ui/components/popover"
)

const (
	propAnchorID          = "anchorID"
	propCancelIntent      = "cancelIntent"
	propCancelText        = "cancelText"
	propCancelVariant     = "cancelVariant"
	propChangeIntent      = "changeIntent"
	propChangeIntentField = "changeIntentField"
	propComponentID       = "componentID"
	propConfirmIntent     = "confirmIntent"
	propDescription       = "description"
	propDisabled          = "disabled"
	propFooterLayout      = "footerLayout"
	propGapRows           = "gapRows"
	propInitialOpen       = "initialOpen"
	propKey               = "key"
	propMaxWidth          = "maxWidth"
	propOkButtonStyle     = "okButtonStyle"
	propOkText            = "okText"
	propOkVariant         = "okVariant"
	propOpen              = "open"
	propOpenControlled    = "openControlled"
	propOverlayStyle      = "overlayStyle"
	propPlacement         = "placement"
	propRootStyle         = "style"
	propShowArrow         = "showArrow"
	propShowCancel        = "showCancel"
	propTextStyle         = "textStyle"
	propTitle             = "title"
	propTitleStyle        = "titleStyle"
	propTrigger           = "trigger"
)

type Placement = popover.Placement
type TriggerMode = popover.TriggerMode

const (
	PlacementAuto        = popover.PlacementAuto
	PlacementTop         = popover.PlacementTop
	PlacementTopLeft     = popover.PlacementTopLeft
	PlacementTopRight    = popover.PlacementTopRight
	PlacementBottom      = popover.PlacementBottom
	PlacementBottomLeft  = popover.PlacementBottomLeft
	PlacementBottomRight = popover.PlacementBottomRight

	TriggerClick  = popover.TriggerClick
	TriggerHover  = popover.TriggerHover
	TriggerManual = popover.TriggerManual
)

// FooterLayout controls how action buttons are arranged inside the overlay footer.
type FooterLayout int

const (
	FooterLayoutEnd FooterLayout = iota
	FooterLayoutCenter
	FooterLayoutStretch
	FooterLayoutVertical
)

// VNode is the declarative description of a Popconfirm component.
type VNode struct {
	*rtui.ElementVNode

	key               string
	componentID       string
	anchorID          string
	child             rtui.VNode
	title             string
	description       string
	placement         Placement
	trigger           TriggerMode
	open              bool
	initialOpen       bool
	openControlled    bool
	disabled          bool
	showArrow         bool
	showCancel        bool
	gapRows           int
	maxWidth          int
	okText            string
	cancelText        string
	okVariant         button.Variant
	cancelVariant     button.Variant
	footerLayout      FooterLayout
	rootStyle         style.Style
	overlayStyle      style.Style
	titleStyle        style.Style
	textStyle         style.Style
	okButtonStyle     style.Style
	confirmIntent     intent.Intent
	cancelIntent      intent.Intent
	changeIntent      intent.Intent
	changeIntentField intent.FieldIntent
}

var (
	_ rtui.VNode           = (*VNode)(nil)
	_ rtui.InstanceFactory = (*VNode)(nil)
)

// New creates a new Popconfirm VNode around an anchor child.
func New(child rtui.VNode) *VNode {
	node := &VNode{
		ElementVNode:  rtui.NewElement("popconfirm"),
		child:         child,
		placement:     PlacementTop,
		trigger:       TriggerClick,
		showArrow:     true,
		showCancel:    true,
		gapRows:       1,
		maxWidth:      36,
		okText:        "OK",
		cancelText:    "Cancel",
		okVariant:     button.VariantPrimary,
		cancelVariant: button.VariantSecondary,
		footerLayout:  FooterLayoutEnd,
	}
	node.captureButtonIntent(child)
	return node
}

func (v *VNode) Key() string { return v.key }

func (v *VNode) SetKey(key string) rtui.VNode {
	v.key = key
	return v
}

func (v *VNode) Tag() string { return "popconfirm" }

func (v *VNode) Style() style.Style { return v.rootStyle }

func (v *VNode) SetStyle(s style.Style) rtui.VNode {
	v.rootStyle = s
	return v
}

func (v *VNode) Children() []rtui.VNode {
	if v.child == nil {
		return nil
	}
	child := v.child
	if child.ID() == "" {
		child.SetID(v.resolvedAnchorID())
	}
	if btn, ok := child.(*button.VNode); ok && v.trigger == TriggerClick {
		btn.SetIntent(ToggleWithID(v.componentRef()))
	}
	return []rtui.VNode{child}
}

func (v *VNode) SetChildren(children []rtui.VNode) rtui.VNode {
	if len(children) > 0 {
		v.SetChild(children[0])
	}
	return v
}

func (v *VNode) GetLayer() rtui.Layer { return rtui.LayerBase }

func (v *VNode) SetLayer(l rtui.Layer) rtui.VNode { return v }

func (v *VNode) Props() rtui.Props {
	return rtui.Props{
		propAnchorID:          v.resolvedAnchorID(),
		propCancelIntent:      v.cancelIntent,
		propCancelText:        v.cancelText,
		propCancelVariant:     v.cancelVariant,
		propChangeIntent:      v.changeIntent,
		propChangeIntentField: v.changeIntentField,
		propComponentID:       v.componentRef(),
		propConfirmIntent:     v.confirmIntent,
		propDescription:       v.description,
		propDisabled:          v.disabled,
		propFooterLayout:      v.footerLayout,
		propGapRows:           v.gapRows,
		propInitialOpen:       v.initialOpen,
		propKey:               v.key,
		propMaxWidth:          v.maxWidth,
		propOkButtonStyle:     v.okButtonStyle,
		propOkText:            v.okText,
		propOkVariant:         v.okVariant,
		propOpen:              v.open,
		propOpenControlled:    v.openControlled,
		propOverlayStyle:      v.overlayStyle,
		propPlacement:         v.placement,
		propRootStyle:         v.rootStyle,
		propShowArrow:         v.showArrow,
		propShowCancel:        v.showCancel,
		propTextStyle:         v.textStyle,
		propTitle:             v.title,
		propTitleStyle:        v.titleStyle,
		propTrigger:           v.trigger,
	}
}

func (v *VNode) SetProps(props rtui.Props) rtui.VNode {
	if key, ok := props[propKey].(string); ok {
		v.key = key
	}
	if componentID, ok := props[propComponentID].(string); ok {
		v.componentID = componentID
	}
	if anchorID, ok := props[propAnchorID].(string); ok {
		v.anchorID = anchorID
	}
	if title, ok := props[propTitle].(string); ok {
		v.title = title
	}
	if description, ok := props[propDescription].(string); ok {
		v.description = description
	}
	if placement, ok := props[propPlacement].(Placement); ok {
		v.placement = placement
	}
	if trigger, ok := props[propTrigger].(TriggerMode); ok {
		v.trigger = trigger
	}
	if open, ok := props[propOpen].(bool); ok {
		v.open = open
	}
	if initialOpen, ok := props[propInitialOpen].(bool); ok {
		v.initialOpen = initialOpen
	}
	if openControlled, ok := props[propOpenControlled].(bool); ok {
		v.openControlled = openControlled
	}
	if disabled, ok := props[propDisabled].(bool); ok {
		v.disabled = disabled
	}
	if showArrow, ok := props[propShowArrow].(bool); ok {
		v.showArrow = showArrow
	}
	if showCancel, ok := props[propShowCancel].(bool); ok {
		v.showCancel = showCancel
	}
	if gapRows, ok := props[propGapRows].(int); ok {
		v.gapRows = gapRows
	}
	if maxWidth, ok := props[propMaxWidth].(int); ok {
		v.maxWidth = maxWidth
	}
	if okText, ok := props[propOkText].(string); ok {
		v.okText = okText
	}
	if cancelText, ok := props[propCancelText].(string); ok {
		v.cancelText = cancelText
	}
	if okVariant, ok := props[propOkVariant].(button.Variant); ok {
		v.okVariant = okVariant
	}
	if cancelVariant, ok := props[propCancelVariant].(button.Variant); ok {
		v.cancelVariant = cancelVariant
	}
	if footerLayout, ok := props[propFooterLayout].(FooterLayout); ok {
		v.footerLayout = footerLayout
	}
	if s, ok := props[propRootStyle].(style.Style); ok {
		v.rootStyle = s
	}
	if s, ok := props[propOverlayStyle].(style.Style); ok {
		v.overlayStyle = s
	}
	if s, ok := props[propTitleStyle].(style.Style); ok {
		v.titleStyle = s
	}
	if s, ok := props[propTextStyle].(style.Style); ok {
		v.textStyle = s
	}
	if s, ok := props[propOkButtonStyle].(style.Style); ok {
		v.okButtonStyle = s
	}
	if confirmIntent, ok := props[propConfirmIntent].(intent.Intent); ok {
		v.confirmIntent = confirmIntent
	}
	if cancelIntent, ok := props[propCancelIntent].(intent.Intent); ok {
		v.cancelIntent = cancelIntent
	}
	if changeIntent, ok := props[propChangeIntent].(intent.Intent); ok {
		v.changeIntent = changeIntent
	}
	if changeIntentField, ok := props[propChangeIntentField].(intent.FieldIntent); ok {
		v.changeIntentField = changeIntentField
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

func (v *VNode) SetAnchorID(id string) *VNode {
	v.anchorID = id
	return v
}

func (v *VNode) SetChild(child rtui.VNode) *VNode {
	v.child = child
	v.captureButtonIntent(child)
	return v
}

func (v *VNode) SetTitle(title string) *VNode {
	v.title = title
	return v
}

func (v *VNode) SetDescription(description string) *VNode {
	v.description = description
	return v
}

func (v *VNode) SetPlacement(placement Placement) *VNode {
	v.placement = placement
	return v
}

func (v *VNode) SetTrigger(trigger TriggerMode) *VNode {
	v.trigger = trigger
	return v
}

func (v *VNode) SetOpen(open bool) *VNode {
	v.open = open
	v.openControlled = true
	return v
}

func (v *VNode) SetInitialOpen(open bool) *VNode {
	v.initialOpen = open
	return v
}

func (v *VNode) SetDisabled(disabled bool) *VNode {
	v.disabled = disabled
	return v
}

func (v *VNode) SetShowArrow(show bool) *VNode {
	v.showArrow = show
	return v
}

func (v *VNode) SetShowCancel(show bool) *VNode {
	v.showCancel = show
	return v
}

func (v *VNode) SetGapRows(rows int) *VNode {
	v.gapRows = rows
	return v
}

func (v *VNode) SetMaxWidth(width int) *VNode {
	v.maxWidth = width
	return v
}

func (v *VNode) SetOkText(text string) *VNode {
	v.okText = text
	return v
}

func (v *VNode) SetCancelText(text string) *VNode {
	v.cancelText = text
	return v
}

func (v *VNode) SetOkVariant(variant button.Variant) *VNode {
	v.okVariant = variant
	return v
}

func (v *VNode) SetCancelVariant(variant button.Variant) *VNode {
	v.cancelVariant = variant
	return v
}

func (v *VNode) SetFooterLayout(layout FooterLayout) *VNode {
	v.footerLayout = layout
	return v
}

func (v *VNode) SetOverlayStyle(s style.Style) *VNode {
	v.overlayStyle = s
	return v
}

func (v *VNode) SetTitleStyle(s style.Style) *VNode {
	v.titleStyle = s
	return v
}

func (v *VNode) SetTextStyle(s style.Style) *VNode {
	v.textStyle = s
	return v
}

func (v *VNode) SetOkButtonStyle(s style.Style) *VNode {
	v.okButtonStyle = s
	return v
}

func (v *VNode) SetConfirmIntent(i intent.Intent) *VNode {
	v.confirmIntent = i
	return v
}

func (v *VNode) SetCancelIntent(i intent.Intent) *VNode {
	v.cancelIntent = i
	return v
}

func (v *VNode) SetChangeIntent(i intent.Intent) *VNode {
	v.changeIntent = i
	return v
}

func (v *VNode) SetChangeIntentField(i intent.FieldIntent) *VNode {
	v.changeIntentField = i
	return v
}

func (v *VNode) componentRef() string {
	if v.componentID != "" {
		return v.componentID
	}
	if v.ID() != "" {
		return v.ID()
	}
	if v.key != "" {
		return v.key
	}
	return "popconfirm"
}

func (v *VNode) resolvedAnchorID() string {
	if v.anchorID != "" {
		return v.anchorID
	}
	if v.child != nil && v.child.ID() != "" {
		return v.child.ID()
	}
	return v.componentRef() + "-anchor"
}

func (v *VNode) captureButtonIntent(child rtui.VNode) {
	if child == nil || v.confirmIntent != nil {
		return
	}
	if btn, ok := child.(*button.VNode); ok && btn.PressIntent() != nil {
		v.confirmIntent = btn.PressIntent()
	}
}
