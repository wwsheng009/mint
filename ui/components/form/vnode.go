package form

import (
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
)

// =============================================================================
// Prop Keys
// =============================================================================

// Prop key constants — shared by VNode and Instance to avoid magic strings.
const (
	propKey = "key"
	propLabel = "label"
	propOnReset = "onReset"
	propOnSubmit = "onSubmit"
	propStyle = "style"
	propValidateAll = "validateAll"
	propValues = "values"
)

// =============================================================================
// VNode - Form Component Description
// =============================================================================

// VNode represents a Form component.
type VNode struct {
	*rtui.ElementVNode

	// === Identification ===
	key string

	// === Props ===
	label       string
	formStyle   style.Style
	values      map[string]interface{} // Initial field values
	validateAll bool                    // Whether to validate all on submit
	onSubmit    intent.Intent           // Intent to emit on submit
	onReset     intent.Intent           // Intent to emit on reset

	// === Children ===
	children []rtui.VNode

	// === Box Model ===
	rtui.BoxModelMixin
}

var (
	_ rtui.VNode           = (*VNode)(nil)
	_ rtui.InstanceFactory = (*VNode)(nil)
	_ rtui.BoxModel        = (*VNode)(nil)
)

// =============================================================================
// VNode Factory
// =============================================================================

// New creates a new Form VNode.
func New() *VNode {
	return &VNode{
		ElementVNode: rtui.NewElement("form"),
		key:          "form",
		validateAll:  true,
		values:       make(map[string]interface{}),
	}
}

// =============================================================================
// VNode Interface Implementation
// =============================================================================

func (f *VNode) Key() string        { return f.key }
func (f *VNode) SetKey(key string) rtui.VNode {
	f.key = key
	return f
}
func (f *VNode) Tag() string        { return "form" }
func (f *VNode) Style() style.Style { return f.formStyle }
func (f *VNode) SetStyle(s style.Style) rtui.VNode {
	f.formStyle = s
	return f
}
func (f *VNode) Children() []rtui.VNode { return f.children }
func (f *VNode) SetChildren(children []rtui.VNode) rtui.VNode {
	f.children = children
	return f
}
func (f *VNode) GetLayer() rtui.Layer { return rtui.LayerBase }
func (f *VNode) SetLayer(l rtui.Layer) rtui.VNode { return f }

func (f *VNode) Props() rtui.Props {
	props := rtui.Props{
		propKey:         f.key,
		propLabel:       f.label,
		propValidateAll: f.validateAll,
		propValues:      f.values,
	}

	if f.formStyle != (style.Style{}) {
		props[propStyle] = f.formStyle
	}
	if f.onSubmit != nil {
		props[propOnSubmit] = f.onSubmit
	}
	if f.onReset != nil {
		props[propOnReset] = f.onReset
	}

	return props
}

func (f *VNode) SetProps(p rtui.Props) rtui.VNode {
	if v, ok := p[propKey].(string); ok {
		f.key = v
	}
	if v, ok := p[propLabel].(string); ok {
		f.label = v
	}
	if v, ok := p[propStyle].(style.Style); ok {
		f.formStyle = v
	}
	if v, ok := p[propValidateAll].(bool); ok {
		f.validateAll = v
	}
	if v, ok := p[propValues].(map[string]interface{}); ok {
		f.values = v
	}
	if v, ok := p[propOnSubmit].(intent.Intent); ok {
		f.onSubmit = v
	}
	if v, ok := p[propOnReset].(intent.Intent); ok {
		f.onReset = v
	}
	return f
}

// =============================================================================
// InstanceFactory Implementation
// =============================================================================

func (f *VNode) CreateInstance() rtui.ComponentInstance {
	return NewInstance(f.Props())
}

// =============================================================================
// Fluent Builder API
// =============================================================================

// Label sets the form label.
func (f *VNode) Label(label string) *VNode {
	f.label = label
	return f
}

// SetValues sets the initial field values.
func (f *VNode) SetValues(values map[string]interface{}) *VNode {
	f.values = values
	return f
}

// SetValue sets an initial field value.
func (f *VNode) SetValue(field string, value interface{}) *VNode {
	if f.values == nil {
		f.values = make(map[string]interface{})
	}
	f.values[field] = value
	return f
}

// ValidateAll sets whether to validate all fields on submit.
func (f *VNode) ValidateAll(validate bool) *VNode {
	f.validateAll = validate
	return f
}

// OnSubmit sets the intent to emit on successful form submission.
func (f *VNode) OnSubmit(intent intent.Intent) *VNode {
	f.onSubmit = intent
	return f
}

// OnReset sets the intent to emit on form reset.
func (f *VNode) OnReset(intent intent.Intent) *VNode {
	f.onReset = intent
	return f
}

// WithStyle sets the form style.
func (f *VNode) WithStyle(s style.Style) *VNode {
	f.formStyle = s
	return f
}

// AddChild adds a child component (field) to the form.
func (f *VNode) AddChild(child rtui.VNode) *VNode {
	f.children = append(f.children, child)
	return f
}

// AddChildren adds multiple child components (fields) to the form.
func (f *VNode) AddChildren(children ...rtui.VNode) *VNode {
	f.children = append(f.children, children...)
	return f
}
