package steps

import (
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

const (
	propComponentID        = "componentID"
	propCurrent            = "current"
	propCurrentControlled  = "currentControlled"
	propCurrentIntent      = "currentIntent"
	propCurrentIntentField = "currentIntentField"
	propDescriptionStyle   = "descriptionStyle"
	propDisabled           = "disabled"
	propDirection          = "direction"
	propErrorStyle         = "errorStyle"
	propFinishStyle        = "finishStyle"
	propInitialCurrent     = "initialCurrent"
	propItems              = "items"
	propKey                = "key"
	propPercent            = "percent"
	propProgressDot        = "progressDot"
	propProcessStyle       = "processStyle"
	propSeparatorStyle     = "separatorStyle"
	propStyle              = "style"
	propTitleStyle         = "titleStyle"
	propWaitStyle          = "waitStyle"
)

// Direction controls the steps layout direction.
type Direction int

const (
	DirectionHorizontal Direction = iota
	DirectionVertical
)

// Status controls the visual state of one step.
type Status int

const (
	StatusAuto Status = iota
	StatusWait
	StatusProcess
	StatusFinish
	StatusError
)

// Item describes one step entry.
type Item struct {
	Key         string
	Title       string
	Description string
	Icon        string
	Status      Status
}

// WithKey sets the item key.
func (i Item) WithKey(key string) Item {
	i.Key = key
	return i
}

// WithDescription sets the item description.
func (i Item) WithDescription(description string) Item {
	i.Description = description
	return i
}

// WithIcon sets the item icon.
func (i Item) WithIcon(icon string) Item {
	i.Icon = icon
	return i
}

// WithStatus sets the item status.
func (i Item) WithStatus(status Status) Item {
	i.Status = status
	return i
}

// AsWait marks the item as waiting.
func (i Item) AsWait() Item {
	i.Status = StatusWait
	return i
}

// AsProcess marks the item as the active/current step.
func (i Item) AsProcess() Item {
	i.Status = StatusProcess
	return i
}

// AsFinish marks the item as finished.
func (i Item) AsFinish() Item {
	i.Status = StatusFinish
	return i
}

// AsError marks the item as errored.
func (i Item) AsError() Item {
	i.Status = StatusError
	return i
}

// Step creates a step item with the given title.
func Step(title string) Item {
	return Item{Title: title}
}

// VNode is the immutable description of a Steps component.
type VNode struct {
	*rtui.ElementVNode

	key                string
	componentID        string
	items              []Item
	current            int
	initialCurrent     int
	currentControlled  bool
	disabled           bool
	percent            int
	progressDot        bool
	direction          Direction
	stepsStyle         style.Style
	titleStyle         style.Style
	descriptionStyle   style.Style
	separatorStyle     style.Style
	waitStyle          style.Style
	processStyle       style.Style
	finishStyle        style.Style
	errorStyle         style.Style
	currentIntent      intent.Intent
	currentIntentField intent.FieldIntent
}

var (
	_ rtui.VNode           = (*VNode)(nil)
	_ rtui.InstanceFactory = (*VNode)(nil)
)

// New creates a new Steps VNode.
func New(items []Item) *VNode {
	return &VNode{
		ElementVNode: rtui.NewElement("steps"),
		items:        cloneItems(items),
		current:      0,
		percent:      -1,
		direction:    DirectionHorizontal,
	}
}

func (v *VNode) Key() string { return v.key }

func (v *VNode) SetKey(key string) rtui.VNode {
	v.key = key
	return v
}

func (v *VNode) Tag() string { return "steps" }

func (v *VNode) Style() style.Style { return v.stepsStyle }

func (v *VNode) SetStyle(s style.Style) rtui.VNode {
	v.stepsStyle = s
	return v
}

func (v *VNode) Children() []rtui.VNode { return nil }

func (v *VNode) SetChildren(children []rtui.VNode) rtui.VNode { return v }

func (v *VNode) GetLayer() rtui.Layer { return rtui.LayerBase }

func (v *VNode) SetLayer(l rtui.Layer) rtui.VNode { return v }

func (v *VNode) Props() rtui.Props {
	return rtui.Props{
		propComponentID:        v.componentID,
		propCurrent:            v.current,
		propCurrentControlled:  v.currentControlled,
		propCurrentIntent:      v.currentIntent,
		propCurrentIntentField: v.currentIntentField,
		propDescriptionStyle:   v.descriptionStyle,
		propDisabled:           v.disabled,
		propDirection:          v.direction,
		propErrorStyle:         v.errorStyle,
		propFinishStyle:        v.finishStyle,
		propInitialCurrent:     v.initialCurrent,
		propItems:              cloneItems(v.items),
		propKey:                v.key,
		propPercent:            v.percent,
		propProgressDot:        v.progressDot,
		propProcessStyle:       v.processStyle,
		propSeparatorStyle:     v.separatorStyle,
		propStyle:              v.stepsStyle,
		propTitleStyle:         v.titleStyle,
		propWaitStyle:          v.waitStyle,
	}
}

func (v *VNode) SetProps(props rtui.Props) rtui.VNode {
	if key, ok := props[propKey].(string); ok {
		v.key = key
	}
	if componentID, ok := props[propComponentID].(string); ok {
		v.componentID = componentID
	}
	if items, ok := props[propItems].([]Item); ok {
		v.items = cloneItems(items)
	}
	if current, ok := props[propCurrent].(int); ok {
		v.current = current
	}
	if initialCurrent, ok := props[propInitialCurrent].(int); ok {
		v.initialCurrent = initialCurrent
	}
	if currentControlled, ok := props[propCurrentControlled].(bool); ok {
		v.currentControlled = currentControlled
	}
	if currentIntent, ok := props[propCurrentIntent].(intent.Intent); ok {
		v.currentIntent = currentIntent
	}
	if currentIntentField, ok := props[propCurrentIntentField].(intent.FieldIntent); ok {
		v.currentIntentField = currentIntentField
	}
	if disabled, ok := props[propDisabled].(bool); ok {
		v.disabled = disabled
	}
	if percent, ok := props[propPercent].(int); ok {
		v.percent = percent
	}
	if progressDot, ok := props[propProgressDot].(bool); ok {
		v.progressDot = progressDot
	}
	if direction, ok := props[propDirection].(Direction); ok {
		v.direction = direction
	}
	if s, ok := props[propStyle].(style.Style); ok {
		v.stepsStyle = s
	}
	if s, ok := props[propTitleStyle].(style.Style); ok {
		v.titleStyle = s
	}
	if s, ok := props[propDescriptionStyle].(style.Style); ok {
		v.descriptionStyle = s
	}
	if s, ok := props[propSeparatorStyle].(style.Style); ok {
		v.separatorStyle = s
	}
	if s, ok := props[propWaitStyle].(style.Style); ok {
		v.waitStyle = s
	}
	if s, ok := props[propProcessStyle].(style.Style); ok {
		v.processStyle = s
	}
	if s, ok := props[propFinishStyle].(style.Style); ok {
		v.finishStyle = s
	}
	if s, ok := props[propErrorStyle].(style.Style); ok {
		v.errorStyle = s
	}
	return v
}

func (v *VNode) CreateInstance() rtui.ComponentInstance {
	return NewInstance(v.Props())
}

// SetItems replaces the step items.
func (v *VNode) SetItems(items []Item) *VNode {
	v.items = cloneItems(items)
	return v
}

// AddItem appends a step item.
func (v *VNode) AddItem(item Item) *VNode {
	v.items = append(v.items, item)
	return v
}

// SetComponentID sets the event component ID.
func (v *VNode) SetComponentID(id string) *VNode {
	v.componentID = id
	return v
}

// SetCurrent sets the current active step.
func (v *VNode) SetCurrent(index int) *VNode {
	v.current = index
	v.currentControlled = true
	return v
}

// SetInitialCurrent sets the initial active step in uncontrolled mode.
func (v *VNode) SetInitialCurrent(index int) *VNode {
	v.initialCurrent = index
	return v
}

// SetDisabled toggles component interactivity.
func (v *VNode) SetDisabled(disabled bool) *VNode {
	v.disabled = disabled
	return v
}

// SetProgressDot toggles dot-style indicators.
func (v *VNode) SetProgressDot(enabled bool) *VNode {
	v.progressDot = enabled
	return v
}

// SetPercent sets the current process percentage.
func (v *VNode) SetPercent(percent int) *VNode {
	v.percent = percent
	return v
}

// SetDirection sets the layout direction.
func (v *VNode) SetDirection(direction Direction) *VNode {
	v.direction = direction
	return v
}

// SetTitleStyle sets the title style.
func (v *VNode) SetTitleStyle(s style.Style) *VNode {
	v.titleStyle = s
	return v
}

// SetDescriptionStyle sets the description style.
func (v *VNode) SetDescriptionStyle(s style.Style) *VNode {
	v.descriptionStyle = s
	return v
}

// SetSeparatorStyle sets the separator style.
func (v *VNode) SetSeparatorStyle(s style.Style) *VNode {
	v.separatorStyle = s
	return v
}

// SetWaitStyle sets the wait-step style override.
func (v *VNode) SetWaitStyle(s style.Style) *VNode {
	v.waitStyle = s
	return v
}

// SetProcessStyle sets the current-step style override.
func (v *VNode) SetProcessStyle(s style.Style) *VNode {
	v.processStyle = s
	return v
}

// SetFinishStyle sets the finish-step style override.
func (v *VNode) SetFinishStyle(s style.Style) *VNode {
	v.finishStyle = s
	return v
}

// SetErrorStyle sets the error-step style override.
func (v *VNode) SetErrorStyle(s style.Style) *VNode {
	v.errorStyle = s
	return v
}

// SetCurrentIntent sets the change intent.
func (v *VNode) SetCurrentIntent(i intent.Intent) *VNode {
	v.currentIntent = i
	return v
}

// SetCurrentIntentField sets the change field binding.
func (v *VNode) SetCurrentIntentField(i intent.FieldIntent) *VNode {
	v.currentIntentField = i
	return v
}

// Items returns the step items.
func (v *VNode) Items() []Item { return cloneItems(v.items) }

// Current returns the current index.
func (v *VNode) Current() int { return v.current }

// Direction returns the layout direction.
func (v *VNode) Direction() Direction { return v.direction }

func cloneItems(items []Item) []Item {
	if len(items) == 0 {
		return nil
	}
	cloned := make([]Item, len(items))
	copy(cloned, items)
	return cloned
}
