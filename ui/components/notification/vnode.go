package notification

import (
	"time"

	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Prop Keys
// =============================================================================

const (
	propKey              = "key"
	propNotificationType = "notificationType"
	propTitle            = "title"
	propMessage          = "message"
	propClosable         = "closable"
	propCloseIntent      = "closeIntent"
	propDuration         = "duration"
	propPlacement        = "placement"
	propStyle            = "style"
)

// =============================================================================
// NotificationType
// =============================================================================

// NotificationType defines the severity level of the notification.
type NotificationType int

const (
	NotificationInfo    NotificationType = iota // Informational
	NotificationSuccess                         // Success
	NotificationWarning                         // Warning
	NotificationError                           // Error
)

// =============================================================================
// Placement
// =============================================================================

// Placement controls where the notification appears on screen.
type Placement int

const (
	PlacementTopRight    Placement = iota // Default
	PlacementTopLeft
	PlacementBottomRight
	PlacementBottomLeft
)

// =============================================================================
// VNode - Description Only (No State, No Closures, No Paint)
// =============================================================================

// VNode is the immutable description of a Notification component.
type VNode struct {
	*rtui.ElementVNode

	key              string
	notificationType NotificationType
	title            string
	message          string
	closable         bool
	closeIntent      interface{}
	duration         time.Duration // 0 = persistent
	placement        Placement
	notifyStyle      style.Style
}

var (
	_ rtui.VNode           = (*VNode)(nil)
	_ rtui.InstanceFactory = (*VNode)(nil)
)

// New creates a new Notification VNode (closable, persistent by default).
func New() *VNode {
	return &VNode{
		ElementVNode: rtui.NewElement("notification"),
		closable:     true,
		placement:    PlacementTopRight,
	}
}

// =============================================================================
// rtui.VNode Interface
// =============================================================================

func (v *VNode) Key() string                         { return v.key }
func (v *VNode) SetKey(key string) rtui.VNode        { v.key = key; return v }
func (v *VNode) Tag() string                         { return "notification" }
func (v *VNode) Style() style.Style                  { return v.notifyStyle }
func (v *VNode) SetStyle(s style.Style) rtui.VNode   { v.notifyStyle = s; return v }
func (v *VNode) Children() []rtui.VNode              { return nil }
func (v *VNode) SetChildren(children []rtui.VNode) rtui.VNode { return v }
func (v *VNode) GetLayer() rtui.Layer                { return rtui.LayerOverlay }
func (v *VNode) SetLayer(l rtui.Layer) rtui.VNode    { return v }

func (v *VNode) Props() rtui.Props {
	return rtui.Props{
		propKey:              v.key,
		propNotificationType: v.notificationType,
		propTitle:            v.title,
		propMessage:          v.message,
		propClosable:         v.closable,
		propCloseIntent:      v.closeIntent,
		propDuration:         v.duration,
		propPlacement:        v.placement,
		propStyle:            v.notifyStyle,
	}
}

func (v *VNode) SetProps(props rtui.Props) rtui.VNode {
	if s, ok := props[propKey].(string); ok {
		v.key = s
	}
	if t, ok := props[propNotificationType].(NotificationType); ok {
		v.notificationType = t
	}
	if s, ok := props[propTitle].(string); ok {
		v.title = s
	}
	if s, ok := props[propMessage].(string); ok {
		v.message = s
	}
	if b, ok := props[propClosable].(bool); ok {
		v.closable = b
	}
	if ci, ok := props[propCloseIntent]; ok {
		v.closeIntent = ci
	}
	if d, ok := props[propDuration].(time.Duration); ok {
		v.duration = d
	}
	if p, ok := props[propPlacement].(Placement); ok {
		v.placement = p
	}
	if s, ok := props[propStyle].(style.Style); ok {
		v.notifyStyle = s
	}
	return v
}

func (v *VNode) SetProp(key string, value interface{}) rtui.VNode {
	return v.SetProps(rtui.Props{key: value})
}

// =============================================================================
// Typed Setters (for builder/fluent use)
// =============================================================================

func (v *VNode) SetTitle(title string) *VNode            { v.title = title; return v }
func (v *VNode) SetMessage(message string) *VNode        { v.message = message; return v }
func (v *VNode) SetType(t NotificationType) *VNode       { v.notificationType = t; return v }
func (v *VNode) SetClosable(closable bool) *VNode        { v.closable = closable; return v }
func (v *VNode) SetCloseIntent(intent interface{}) *VNode { v.closeIntent = intent; return v }
func (v *VNode) SetDuration(d time.Duration) *VNode      { v.duration = d; return v }
func (v *VNode) SetPlacement(p Placement) *VNode         { v.placement = p; return v }

// Convenience type setters
func (v *VNode) Info() *VNode    { v.notificationType = NotificationInfo; return v }
func (v *VNode) Success() *VNode { v.notificationType = NotificationSuccess; return v }
func (v *VNode) Warning() *VNode { v.notificationType = NotificationWarning; return v }
func (v *VNode) Error() *VNode   { v.notificationType = NotificationError; return v }

// =============================================================================
// InstanceFactory
// =============================================================================

func (v *VNode) CreateInstance() rtui.ComponentInstance {
	return NewInstance(v.Props())
}
