package tooltip

import (
	"time"

	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Prop Keys
// =============================================================================

// Prop key constants — shared by VNode and Instance to avoid magic strings.
const (
	propDelay = "delay"
	propKey = "key"
	propLayer = "layer"
	propPosition = "position"
	propStyle = "style"
	propText = "text"
)

// =============================================================================
// Tooltip Position
// =============================================================================

// Position defines where the tooltip appears relative to its anchor.
type Position int

const (
	PositionTop    Position = iota // Tooltip appears above anchor
	PositionBottom                 // Tooltip appears below anchor
	PositionLeft                   // Tooltip appears to the left of anchor
	PositionRight                  // Tooltip appears to the right of anchor
	PositionAuto                   // Position is automatically determined
)

// =============================================================================
// Tooltip VNode - Description Only (No State, No Closures, No Paint)
// =============================================================================

// VNode is the tooltip description.
// It contains ONLY declarative information - no state, no closures, no paint logic.
type VNode struct {
	*rtui.ElementVNode

	// === Identification ===
	key string

	// === Visual Props ===
	text     string
	style    style.Style
	position Position
	delay    time.Duration

	// === Rendering Layer ===
	layer rtui.Layer // Layer for Z-order: Base, Overlay, Modal, Tooltip, Inspector

	// === Content Props ===
	content rtui.VNode // Child content that triggers tooltip
}

// Ensure VNode implements required interfaces
var (
	_ rtui.VNode           = (*VNode)(nil)
	_ rtui.InstanceFactory = (*VNode)(nil)
)

// =============================================================================
// Constructor
// =============================================================================

// New creates a new Tooltip VNode wrapping content.
func New(content rtui.VNode, text string) *VNode {
	return &VNode{
		ElementVNode: rtui.NewElement("tooltip"),
		content:      content,
		text:         text,
		position:     PositionAuto,
		delay:        500 * time.Millisecond,
		layer:        rtui.LayerTooltip, // Default to Tooltip layer
	}
}

// =============================================================================
// rtui.VNode Interface Implementation
// =============================================================================

// Key returns the component key.
func (t *VNode) Key() string {
	return t.key
}

// SetKey sets the component key - returns VNode for chaining.
func (t *VNode) SetKey(key string) rtui.VNode {
	t.key = key
	return t
}

// Tag returns the tag name.
func (t *VNode) Tag() string {
	return "tooltip"
}

// Style returns the visual style.
func (t *VNode) Style() style.Style {
	return t.style
}

// SetStyle sets the visual style - returns VNode for chaining.
func (t *VNode) SetStyle(s style.Style) rtui.VNode {
	t.style = s
	return t
}

// Children returns the wrapped content as a child.
func (t *VNode) Children() []rtui.VNode {
	if t.content != nil {
		return []rtui.VNode{t.content}
	}
	return nil
}

// SetChildren sets the wrapped content - returns VNode for chaining.
func (t *VNode) SetChildren(children []rtui.VNode) rtui.VNode {
	if len(children) > 0 {
		t.content = children[0]
	}
	return t
}

// GetLayer returns the rendering layer (tooltips appear above content).
func (t *VNode) GetLayer() rtui.Layer {
	return t.layer
}

// SetLayer sets the rendering layer - returns VNode for chaining.
func (t *VNode) SetLayer(l rtui.Layer) rtui.VNode {
	t.layer = l
	return t
}

// Props returns the node properties.
func (t *VNode) Props() rtui.Props {
	return rtui.Props{
		propKey:      t.key,
		propText:     t.text,
		propPosition: t.position,
		propDelay:    t.delay,
		propStyle:    t.style,
		propLayer:    t.layer,
	}
}

// SetProps sets the node properties - returns VNode for chaining.
func (t *VNode) SetProps(p rtui.Props) rtui.VNode {
	if v, ok := p[propKey].(string); ok {
		t.key = v
	}
	if v, ok := p[propText].(string); ok {
		t.text = v
	}
	if v, ok := p[propPosition].(Position); ok {
		t.position = v
	}
	if v, ok := p[propDelay].(time.Duration); ok {
		t.delay = v
	}
	if v, ok := p[propStyle].(style.Style); ok {
		t.style = v
	}
	if v, ok := p[propLayer].(rtui.Layer); ok {
		t.layer = v
	}
	return t
}

// =============================================================================
// InstanceFactory Implementation
// =============================================================================

// CreateInstance creates a new Instance from this VNode description.
func (t *VNode) CreateInstance() rtui.ComponentInstance {
	props := rtui.Props{
		propKey:      t.key,
		propText:     t.text,
		propPosition: t.position,
		propDelay:    t.delay,
		propStyle:    t.style,
	}
	return NewInstance(props)
}

// =============================================================================
// Builder Methods - Fluent API
// =============================================================================

// SetText sets the tooltip text.
func (t *VNode) SetText(text string) *VNode {
	t.text = text
	return t
}

// SetPosition sets the tooltip position.
func (t *VNode) SetPosition(position Position) *VNode {
	t.position = position
	return t
}

// SetDelay sets the delay before showing the tooltip.
func (t *VNode) SetDelay(delay time.Duration) *VNode {
	t.delay = delay
	return t
}

// SetStyleProps sets the visual style.
func (t *VNode) SetStyleProps(s style.Style) *VNode {
	t.style = s
	return t
}

// =============================================================================
// Props Accessors
// =============================================================================

// Text returns the tooltip text.
func (t *VNode) Text() string {
	return t.text
}

// Position returns the tooltip position.
func (t *VNode) Position() Position {
	return t.position
}

// Delay returns the delay before showing.
func (t *VNode) Delay() time.Duration {
	return t.delay
}

// Content returns the wrapped content node.
func (t *VNode) Content() rtui.VNode {
	return t.content
}

