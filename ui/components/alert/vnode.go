package alert

import (
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Prop Keys
// =============================================================================

const (
	propKey         = "key"
	propAlertType   = "alertType"
	propTitle       = "title"
	propMessage     = "message"
	propClosable    = "closable"
	propCloseIntent = "closeIntent"
	propStyle       = "style"
)

// =============================================================================
// Alert Type
// =============================================================================

// AlertType defines the type (severity) of the alert.
type AlertType int

const (
	AlertInfo    AlertType = iota // Informational
	AlertSuccess                  // Success
	AlertWarning                  // Warning
	AlertError                    // Error
)

// =============================================================================
// VNode - Description Only (No State, No Closures, No Paint)
// =============================================================================

// VNode is the immutable description of an Alert component.
type VNode struct {
	*rtui.ElementVNode

	key         string
	alertType   AlertType
	title       string
	message     string
	closable    bool
	closeIntent interface{}
	alertStyle  style.Style
}

var (
	_ rtui.VNode           = (*VNode)(nil)
	_ rtui.InstanceFactory = (*VNode)(nil)
)

// New creates a new Alert VNode.
func New() *VNode {
	return &VNode{
		ElementVNode: rtui.NewElement("alert"),
		alertType:    AlertInfo,
	}
}

// =============================================================================
// rtui.VNode Interface
// =============================================================================

func (v *VNode) Key() string                         { return v.key }
func (v *VNode) SetKey(key string) rtui.VNode        { v.key = key; return v }
func (v *VNode) Tag() string                         { return "alert" }
func (v *VNode) Style() style.Style                  { return v.alertStyle }
func (v *VNode) SetStyle(s style.Style) rtui.VNode   { v.alertStyle = s; return v }
func (v *VNode) Children() []rtui.VNode              { return nil }
func (v *VNode) SetChildren(children []rtui.VNode) rtui.VNode { return v }
func (v *VNode) GetLayer() rtui.Layer                { return rtui.LayerBase }
func (v *VNode) SetLayer(l rtui.Layer) rtui.VNode    { return v }

func (v *VNode) Props() rtui.Props {
	return rtui.Props{
		propKey:         v.key,
		propAlertType:   v.alertType,
		propTitle:       v.title,
		propMessage:     v.message,
		propClosable:    v.closable,
		propCloseIntent: v.closeIntent,
		propStyle:       v.alertStyle,
	}
}

func (v *VNode) SetProps(props rtui.Props) rtui.VNode {
	if s, ok := props[propKey].(string); ok {
		v.key = s
	}
	if t, ok := props[propAlertType].(AlertType); ok {
		v.alertType = t
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
	if s, ok := props[propStyle].(style.Style); ok {
		v.alertStyle = s
	}
	return v
}

// =============================================================================
// InstanceFactory
// =============================================================================

func (v *VNode) CreateInstance() rtui.ComponentInstance {
	return NewInstance(v.Props())
}

// =============================================================================
// Builder Methods
// =============================================================================

func (v *VNode) SetTitle(title string) *VNode            { v.title = title; return v }
func (v *VNode) SetMessage(message string) *VNode        { v.message = message; return v }
func (v *VNode) SetAlertType(t AlertType) *VNode         { v.alertType = t; return v }
func (v *VNode) SetClosable(closable bool) *VNode        { v.closable = closable; return v }
func (v *VNode) SetCloseIntent(ci interface{}) *VNode    { v.closeIntent = ci; return v }
func (v *VNode) SetAlertStyle(s style.Style) *VNode      { v.alertStyle = s; return v }

func (v *VNode) Info() *VNode    { v.alertType = AlertInfo; return v }
func (v *VNode) Success() *VNode { v.alertType = AlertSuccess; return v }
func (v *VNode) Warning() *VNode { v.alertType = AlertWarning; return v }
func (v *VNode) Error() *VNode   { v.alertType = AlertError; return v }

// =============================================================================
// Props Accessors
// =============================================================================

func (v *VNode) Title() string          { return v.title }
func (v *VNode) Message() string        { return v.message }
func (v *VNode) AlertType() AlertType   { return v.alertType }
func (v *VNode) Closable() bool         { return v.closable }
func (v *VNode) CloseIntent() interface{} { return v.closeIntent }
