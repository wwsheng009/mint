package popover

import (
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

const (
	propBody              = "body"
	propBorderStyle       = "borderStyle"
	propChangeIntent      = "changeIntent"
	propChangeIntentField = "changeIntentField"
	propComponentID       = "componentID"
	propDisabled          = "disabled"
	propGapRows           = "gapRows"
	propInitialOpen       = "initialOpen"
	propKey               = "key"
	propMaxWidth          = "maxWidth"
	propOpen              = "open"
	propOpenControlled    = "openControlled"
	propOverlayStyle      = "overlayStyle"
	propPlacement         = "placement"
	propShadowStyle       = "shadowStyle"
	propShowArrow         = "showArrow"
	propStyle             = "style"
	propTitle             = "title"
	propTitleStyle        = "titleStyle"
	propTrigger           = "trigger"
	propBodyStyle         = "bodyStyle"
)

// Placement controls where the popover card is displayed relative to the anchor.
type Placement int

const (
	PlacementAuto Placement = iota
	PlacementTop
	PlacementTopLeft
	PlacementTopRight
	PlacementBottom
	PlacementBottomLeft
	PlacementBottomRight
)

// TriggerMode controls how the popover is shown.
type TriggerMode int

const (
	TriggerClick TriggerMode = iota
	TriggerHover
	TriggerManual
)

// VNode is the declarative description of a Popover component.
type VNode struct {
	*rtui.ElementVNode

	key               string
	componentID       string
	child             rtui.VNode
	title             string
	body              string
	placement         Placement
	trigger           TriggerMode
	open              bool
	initialOpen       bool
	openControlled    bool
	disabled          bool
	showArrow         bool
	gapRows           int
	maxWidth          int
	rootStyle         style.Style
	overlayStyle      style.Style
	borderStyle       style.Style
	shadowStyle       style.Style
	titleStyle        style.Style
	bodyStyle         style.Style
	changeIntent      intent.Intent
	changeIntentField intent.FieldIntent
}

var (
	_ rtui.VNode           = (*VNode)(nil)
	_ rtui.InstanceFactory = (*VNode)(nil)
)

// New creates a new Popover VNode around the provided anchor child.
func New(child rtui.VNode) *VNode {
	return &VNode{
		ElementVNode: rtui.NewElement("popover"),
		child:        child,
		placement:    PlacementAuto,
		trigger:      TriggerClick,
		showArrow:    true,
		gapRows:      1,
		maxWidth:     32,
	}
}

func (v *VNode) Key() string { return v.key }

func (v *VNode) SetKey(key string) rtui.VNode {
	v.key = key
	return v
}

func (v *VNode) Tag() string { return "popover" }

func (v *VNode) Style() style.Style { return v.rootStyle }

func (v *VNode) SetStyle(s style.Style) rtui.VNode {
	v.rootStyle = s
	return v
}

func (v *VNode) Children() []rtui.VNode {
	if v.child == nil {
		return nil
	}
	return []rtui.VNode{v.child}
}

func (v *VNode) SetChildren(children []rtui.VNode) rtui.VNode {
	if len(children) > 0 {
		v.child = children[0]
	}
	return v
}

func (v *VNode) GetLayer() rtui.Layer { return rtui.LayerBase }

func (v *VNode) SetLayer(l rtui.Layer) rtui.VNode { return v }

func (v *VNode) Props() rtui.Props {
	return rtui.Props{
		propBody:              v.body,
		propBorderStyle:       v.borderStyle,
		propChangeIntent:      v.changeIntent,
		propChangeIntentField: v.changeIntentField,
		propComponentID:       v.componentID,
		propDisabled:          v.disabled,
		propGapRows:           v.gapRows,
		propInitialOpen:       v.initialOpen,
		propKey:               v.key,
		propMaxWidth:          v.maxWidth,
		propOpen:              v.open,
		propOpenControlled:    v.openControlled,
		propOverlayStyle:      v.overlayStyle,
		propPlacement:         v.placement,
		propShadowStyle:       v.shadowStyle,
		propShowArrow:         v.showArrow,
		propStyle:             v.rootStyle,
		propTitle:             v.title,
		propTitleStyle:        v.titleStyle,
		propTrigger:           v.trigger,
		propBodyStyle:         v.bodyStyle,
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
	if body, ok := props[propBody].(string); ok {
		v.body = body
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
	if gapRows, ok := props[propGapRows].(int); ok {
		v.gapRows = gapRows
	}
	if maxWidth, ok := props[propMaxWidth].(int); ok {
		v.maxWidth = maxWidth
	}
	if s, ok := props[propStyle].(style.Style); ok {
		v.rootStyle = s
	}
	if s, ok := props[propOverlayStyle].(style.Style); ok {
		v.overlayStyle = s
	}
	if s, ok := props[propBorderStyle].(style.Style); ok {
		v.borderStyle = s
	}
	if s, ok := props[propShadowStyle].(style.Style); ok {
		v.shadowStyle = s
	}
	if s, ok := props[propTitleStyle].(style.Style); ok {
		v.titleStyle = s
	}
	if s, ok := props[propBodyStyle].(style.Style); ok {
		v.bodyStyle = s
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

// SetComponentID sets the local intent routing ID.
func (v *VNode) SetComponentID(id string) *VNode {
	v.componentID = id
	return v
}

// SetChild sets the anchor child.
func (v *VNode) SetChild(child rtui.VNode) *VNode {
	v.child = child
	return v
}

// SetTitle sets the popover title.
func (v *VNode) SetTitle(title string) *VNode {
	v.title = title
	return v
}

// SetBody sets the popover body text.
func (v *VNode) SetBody(body string) *VNode {
	v.body = body
	return v
}

// SetPlacement sets the placement mode.
func (v *VNode) SetPlacement(placement Placement) *VNode {
	v.placement = placement
	return v
}

// SetTrigger sets the trigger mode.
func (v *VNode) SetTrigger(trigger TriggerMode) *VNode {
	v.trigger = trigger
	return v
}

// SetOpen sets the controlled open state.
func (v *VNode) SetOpen(open bool) *VNode {
	v.open = open
	v.openControlled = true
	return v
}

// SetInitialOpen sets the uncontrolled initial open state.
func (v *VNode) SetInitialOpen(open bool) *VNode {
	v.initialOpen = open
	return v
}

// SetDisabled toggles disabled state.
func (v *VNode) SetDisabled(disabled bool) *VNode {
	v.disabled = disabled
	return v
}

// SetShowArrow toggles arrow rendering.
func (v *VNode) SetShowArrow(show bool) *VNode {
	v.showArrow = show
	return v
}

// SetGapRows sets the vertical gap between anchor and card.
func (v *VNode) SetGapRows(rows int) *VNode {
	v.gapRows = rows
	return v
}

// SetMaxWidth sets the maximum wrapped content width.
func (v *VNode) SetMaxWidth(width int) *VNode {
	v.maxWidth = width
	return v
}

// SetOverlayStyle sets the overlay fill style.
func (v *VNode) SetOverlayStyle(s style.Style) *VNode {
	v.overlayStyle = s
	return v
}

// SetBorderStyle sets the border style.
func (v *VNode) SetBorderStyle(s style.Style) *VNode {
	v.borderStyle = s
	return v
}

// SetShadowStyle sets the shadow style.
func (v *VNode) SetShadowStyle(s style.Style) *VNode {
	v.shadowStyle = s
	return v
}

// SetTitleStyle sets the title text style.
func (v *VNode) SetTitleStyle(s style.Style) *VNode {
	v.titleStyle = s
	return v
}

// SetBodyStyle sets the body text style.
func (v *VNode) SetBodyStyle(s style.Style) *VNode {
	v.bodyStyle = s
	return v
}

// SetChangeIntent sets an additional custom change intent.
func (v *VNode) SetChangeIntent(i intent.Intent) *VNode {
	v.changeIntent = i
	return v
}

// SetChangeIntentField sets the field binding for open state.
func (v *VNode) SetChangeIntentField(i intent.FieldIntent) *VNode {
	v.changeIntentField = i
	return v
}
