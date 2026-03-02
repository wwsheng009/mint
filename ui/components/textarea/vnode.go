package textarea

import (
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// VNode - Description Only (No State, No Closures, No Paint)
// =============================================================================

// VNode is the textarea description.
type VNode struct {
	*rtui.ElementVNode

	// === Identification ===
	key string

	// === Visual Props ===
	placeholder string
	style       style.Style

	// === Layout Props ===
	rows int
	cols int

	// === Intent Props (no closures!) ===
	changeIntent intent.Intent
	submitIntent intent.Intent

	// === State Props (declarative) ===
	value    string
	maxLen   int
	disabled bool

	// === Box Model ===
	rtui.BoxModelMixin
}

// Ensure VNode implements required interfaces
var (
	_ rtui.VNode           = (*VNode)(nil)
	_ rtui.InstanceFactory = (*VNode)(nil)
	_ rtui.BoxModel        = (*VNode)(nil)
)

// New creates a new Textarea VNode.
func New() *VNode {
	return &VNode{
		ElementVNode: rtui.NewElement("textarea"),
		rows:         3,
		cols:         40,
	}
}

// =============================================================================
// rtui.VNode Interface Implementation
// =============================================================================

func (t *VNode) Key() string                    { return t.key }
func (t *VNode) SetKey(key string) rtui.VNode   { t.key = key; return t }
func (t *VNode) Tag() string                    { return "textarea" }
func (t *VNode) Style() style.Style             { return t.style }
func (t *VNode) SetStyle(s style.Style) rtui.VNode { t.style = s; return t }
func (t *VNode) Children() []rtui.VNode         { return nil }
func (t *VNode) SetChildren(children []rtui.VNode) rtui.VNode { return t }
func (t *VNode) GetLayer() rtui.Layer           { return rtui.LayerBase }
func (t *VNode) SetLayer(l rtui.Layer) rtui.VNode { return t }

func (t *VNode) Props() rtui.Props {
	return rtui.Props{
		"key":          t.key,
		"placeholder":  t.placeholder,
		"style":        t.style,
		"rows":         t.rows,
		"cols":         t.cols,
		"changeIntent": t.changeIntent,
		"submitIntent": t.submitIntent,
		"value":        t.value,
		"maxLen":       t.maxLen,
		"disabled":     t.disabled,
	}
}

func (t *VNode) SetProps(p rtui.Props) rtui.VNode {
	if v, ok := p["key"].(string); ok {
		t.key = v
	}
	if v, ok := p["placeholder"].(string); ok {
		t.placeholder = v
	}
	if v, ok := p["style"].(style.Style); ok {
		t.style = v
	}
	if v, ok := p["rows"].(int); ok {
		t.rows = v
	}
	if v, ok := p["cols"].(int); ok {
		t.cols = v
	}
	if v, ok := p["changeIntent"].(intent.Intent); ok {
		t.changeIntent = v
	}
	if v, ok := p["submitIntent"].(intent.Intent); ok {
		t.submitIntent = v
	}
	if v, ok := p["value"].(string); ok {
		t.value = v
	}
	if v, ok := p["maxLen"].(int); ok {
		t.maxLen = v
	}
	if v, ok := p["disabled"].(bool); ok {
		t.disabled = v
	}
	return t
}

// =============================================================================
// InstanceFactory Implementation
// =============================================================================

func (t *VNode) CreateInstance() rtui.ComponentInstance {
	return NewInstance(t.Props())
}

// =============================================================================
// Builder Methods
// =============================================================================

func (t *VNode) SetPlaceholder(text string) *VNode { t.placeholder = text; return t }
func (t *VNode) SetValue(value string) *VNode      { t.value = value; return t }
func (t *VNode) SetRows(rows int) *VNode           { t.rows = rows; return t }
func (t *VNode) SetCols(cols int) *VNode           { t.cols = cols; return t }
func (t *VNode) SetMaxLen(len int) *VNode          { t.maxLen = len; return t }
func (t *VNode) SetDisabled(disabled bool) *VNode  { t.disabled = disabled; return t }
func (t *VNode) SetChangeIntent(i intent.Intent) *VNode { t.changeIntent = i; return t }
func (t *VNode) SetSubmitIntent(i intent.Intent) *VNode { t.submitIntent = i; return t }

// =============================================================================
// Props Accessors
// =============================================================================

func (t *VNode) Placeholder() string   { return t.placeholder }
func (t *VNode) Value() string         { return t.value }
func (t *VNode) Rows() int             { return t.rows }
func (t *VNode) Cols() int             { return t.cols }
func (t *VNode) MaxLen() int           { return t.maxLen }
func (t *VNode) Disabled() bool        { return t.disabled }
func (t *VNode) ChangeIntent() intent.Intent { return t.changeIntent }
func (t *VNode) SubmitIntent() intent.Intent { return t.submitIntent }

// =============================================================================
// layout.BoxModelProvider Implementation
// =============================================================================

// GetBoxModel returns the box model for the TextArea VNode.
// Implements layout.BoxModelProvider for unified padding/border handling.
// Note: TextArea uses BoxModelMixin for padding/margin, and has no border.
func (t *VNode) GetBoxModel() layout.BoxModel {
	return layout.BoxModel{
		Padding: layout.Padding{
			Left:   t.BoxModelMixin.Padding()[3],
			Right:  t.BoxModelMixin.Padding()[1],
			Top:    t.BoxModelMixin.Padding()[0],
			Bottom: t.BoxModelMixin.Padding()[2],
		},
		Margin: layout.Margin{
			Left:   t.BoxModelMixin.Margin()[3],
			Right:  t.BoxModelMixin.Margin()[1],
			Top:    t.BoxModelMixin.Margin()[0],
			Bottom: t.BoxModelMixin.Margin()[2],
		},
		Border: layout.Border{Style: layout.BorderNone},
	}
}