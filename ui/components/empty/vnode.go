package empty

import (
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Prop Keys
// =============================================================================

const (
	propKey         = "key"
	propDescription = "description"
	propImage       = "image"
	propStyle       = "style"
)

// =============================================================================
// VNode - Description Only (No State, No Closures, No Paint)
// =============================================================================

// VNode is the immutable description of an Empty component.
type VNode struct {
	*rtui.ElementVNode

	key         string
	description string
	image       string // optional ASCII art / icon override
	emptyStyle  style.Style
}

var (
	_ rtui.VNode           = (*VNode)(nil)
	_ rtui.InstanceFactory = (*VNode)(nil)
)

// New creates a new Empty VNode with default description.
func New() *VNode {
	return &VNode{
		ElementVNode: rtui.NewElement("empty"),
		description:  "No Data",
	}
}

// =============================================================================
// Setters (return rtui.VNode for interface compliance)
// =============================================================================

func (v *VNode) SetKey(key string) rtui.VNode {
	v.key = key
	v.Props()[propKey] = key
	return v
}

func (v *VNode) SetDescription(desc string) *VNode {
	v.description = desc
	v.Props()[propDescription] = desc
	return v
}

func (v *VNode) SetImage(img string) *VNode {
	v.image = img
	v.Props()[propImage] = img
	return v
}

func (v *VNode) SetStyle(s style.Style) rtui.VNode {
	v.emptyStyle = s
	v.Props()[propStyle] = s
	return v
}

// =============================================================================
// InstanceFactory Interface
// =============================================================================

// CreateInstance creates a runtime Instance from this VNode's props.
func (v *VNode) CreateInstance() rtui.ComponentInstance {
	return NewInstance(v.Props())
}
