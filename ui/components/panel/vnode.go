// Package panel provides a Fiber-first Panel container component.
// Panel is a high-level container that manages borders, headers, and content layout.
// It is implemented as a composition of Border + Stack components.
package panel

import (
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	newborder "github.com/wwsheng009/mint/ui/components/border"
	newstack "github.com/wwsheng009/mint/ui/components/stack"
	newtext "github.com/wwsheng009/mint/ui/components/text"
)

// =============================================================================
// VNode - Composition-based Panel (no Instance needed)
// =============================================================================

// VNode is the panel container description.
// Panel is implemented as a composition: Border(VStack(header, content, footer))
type VNode struct {
	*rtui.ElementVNode

	// === Identification ===
	key string

	// === Panel Props ===
	title       string
	borderStyle layout.BorderStyle
	borderColor style.Color
	borderLabel string

	// === Layout Props ===
	width   int
	height  int
	flex    int
	padding int

	// === Content ===
	header  rtui.VNode
	content rtui.VNode
	footer  rtui.VNode

	// === Style ===
	instStyle style.Style

	// === Composed node (built on demand) ===
	composed rtui.VNode
}

// Ensure VNode implements required interfaces
var (
	_ rtui.VNode           = (*VNode)(nil)
	_ rtui.InstanceFactory = (*VNode)(nil)
)

// =============================================================================
// Constructors
// =============================================================================

// New creates a new panel container VNode.
func New() *VNode {
	return &VNode{
		ElementVNode: rtui.NewElement("panel"),
		borderStyle:  layout.BorderSingle,
		borderColor:  style.Color("blue"),
		padding:      0,
	}
}

// =============================================================================
// rtui.VNode Interface Implementation
// =============================================================================

// Key returns the component key.
func (v *VNode) Key() string {
	return v.key
}

// SetKey sets the component key - returns VNode for chaining.
func (v *VNode) SetKey(key string) rtui.VNode {
	v.key = key
	return v
}

// Tag returns the tag name.
func (v *VNode) Tag() string {
	return "panel"
}

// Type returns the node type.
func (v *VNode) Type() rtui.VNodeType {
	return rtui.VNodeElement
}

// Children returns child nodes.
// Panel returns the composed Border as its only child.
// This allows Fiber to render the complete Border(VStack) structure with correct positioning.
func (v *VNode) Children() []rtui.VNode {
	composed := v.getComposed()
	if composed != nil {
		return []rtui.VNode{composed}
	}
	return nil
}

// SetChildren sets child nodes - returns VNode for chaining.
func (v *VNode) SetChildren(children []rtui.VNode) rtui.VNode {
	// Panel uses SetContent/SetHeader/SetFooter, ignore generic SetChildren
	return v
}

// GetLayer returns the rendering layer.
func (v *VNode) GetLayer() rtui.Layer {
	return rtui.LayerBase
}

// SetLayer sets the rendering layer - returns VNode for chaining.
func (v *VNode) SetLayer(l rtui.Layer) rtui.VNode {
	return v
}

// Props returns the node properties.
func (v *VNode) Props() rtui.Props {
	return rtui.Props{
		"key":         v.key,
		"style":       v.instStyle,
		"title":       v.title,
		"width":       v.width,
		"height":      v.height,
		"flex":        v.flex,
		"padding":     v.padding,
		"borderStyle": v.borderStyle,
		"borderColor": v.borderColor,
		"borderLabel": v.borderLabel,
		"header":      v.header,
		"content":     v.content,
		"footer":      v.footer,
	}
}

// SetProps sets the node properties - returns VNode for chaining.
func (v *VNode) SetProps(p rtui.Props) rtui.VNode {
	if val, ok := p["key"].(string); ok {
		v.key = val
	}
	if val, ok := p["style"].(style.Style); ok {
		v.instStyle = val
	}
	if val, ok := p["title"].(string); ok {
		v.title = val
	}
	if val, ok := p["width"].(int); ok {
		v.width = val
	}
	if val, ok := p["height"].(int); ok {
		v.height = val
	}
	if val, ok := p["flex"].(int); ok {
		v.flex = val
	}
	if val, ok := p["padding"].(int); ok {
		v.padding = val
	}
	if val, ok := p["borderStyle"].(layout.BorderStyle); ok {
		v.borderStyle = val
	}
	if val, ok := p["borderColor"].(style.Color); ok {
		v.borderColor = val
	}
	if val, ok := p["borderLabel"].(string); ok {
		v.borderLabel = val
	}
	if val, ok := p["header"].(rtui.VNode); ok {
		v.header = val
	}
	if val, ok := p["content"].(rtui.VNode); ok {
		v.content = val
	}
	if val, ok := p["footer"].(rtui.VNode); ok {
		v.footer = val
	}
	// Reset composed to rebuild
	v.composed = nil
	return v
}

// Style returns the node style.
func (v *VNode) Style() style.Style {
	return v.instStyle
}

// SetStyle sets the node style - returns VNode for chaining.
func (v *VNode) SetStyle(s style.Style) rtui.VNode {
	v.instStyle = s
	return v
}

// TextContent returns the text content.
func (v *VNode) TextContent() string {
	return ""
}

// =============================================================================
// InstanceFactory Interface - Delegates to composed Border
// =============================================================================

// CreateInstance creates a new PanelInstance that wraps the Border instance
// and adds constraint tracing at the Panel level.
func (v *VNode) CreateInstance() rtui.ComponentInstance {
	// Create Panel instance with empty path (will be set by layout system)
	return newPanelInstance(v, "")
}

// =============================================================================
// Composition - Build the Border(VStack) structure
// =============================================================================

// getComposed builds and returns the composed node structure.
func (v *VNode) getComposed() rtui.VNode {
	if v.composed != nil {
		return v.composed
	}

	// Build the internal VStack structure
	var stackChildren []rtui.VNode

	// 1. Header (only if explicitly set - title is shown in border label, not as header)
	if v.header != nil {
		stackChildren = append(stackChildren, v.header)
	}

	// 2. Content (flex=1 to fill remaining space)
	contentNode := v.content
	if contentNode == nil {
		contentNode = newtext.New("")
	}
	// Wrap content with flex
	stackChildren = append(stackChildren, rtui.Flex(contentNode, 1))

	// 3. Footer
	if v.footer != nil {
		stackChildren = append(stackChildren, v.footer)
	}

	// Create VStack
	vstack := newstack.New(newstack.Column).
		SetChildrenList(stackChildren).
		SetGap(0)

	// Build border label from title (if not explicitly set)
	borderLabel := v.borderLabel
	if borderLabel == "" && v.title != "" {
		borderLabel = " " + v.title + " "
	}

	// Create Border wrapper
	border := newborder.New().
		SetBorderStyle(v.borderStyle).
		SetBorderColor(v.borderColor).
		SetBorderLabel(borderLabel).
		SetChild(vstack)

	// Panel width/height = total size including border
	// Border width/height = inner content size (border adds padding)
	// So we subtract border padding (2 for single-style borders)
	borderPadding := 2 * newborder.GetBorderWidth(v.borderStyle)
	if v.width > 0 {
		innerWidth := v.width - borderPadding
		if innerWidth > 0 {
			border.SetWidth(innerWidth)
		}
	}
	if v.height > 0 {
		innerHeight := v.height - borderPadding
		if innerHeight > 0 {
			border.SetHeight(innerHeight)
		}
	}
	if v.flex > 0 {
		border.SetFlex(v.flex)
	}
	if v.instStyle.FG != "" || v.instStyle.BG != "" {
		border.SetStyle(v.instStyle)
	}

	v.composed = border
	return v.composed
}

// =============================================================================
// Fluent Setters
// =============================================================================

// SetTitle sets the title - returns VNode for chaining.
func (v *VNode) SetTitle(title string) *VNode {
	v.title = title
	v.composed = nil
	return v
}

// SetHeader sets the header component.
func (v *VNode) SetHeader(header rtui.VNode) *VNode {
	v.header = header
	v.composed = nil
	return v
}

// SetContent sets the main content component.
func (v *VNode) SetContent(content rtui.VNode) *VNode {
	v.content = content
	v.composed = nil
	return v
}

// SetFooter sets the footer component.
func (v *VNode) SetFooter(footer rtui.VNode) *VNode {
	v.footer = footer
	v.composed = nil
	return v
}

// SetWidth sets the width.
func (v *VNode) SetWidth(w int) *VNode {
	v.width = w
	v.composed = nil
	return v
}

// SetHeight sets the height.
func (v *VNode) SetHeight(h int) *VNode {
	v.height = h
	v.composed = nil
	return v
}

// SetFlex sets the flex factor.
func (v *VNode) SetFlex(f int) *VNode {
	v.flex = f
	v.composed = nil
	return v
}

// SetPadding sets the inner padding.
func (v *VNode) SetPadding(p int) *VNode {
	v.padding = p
	v.composed = nil
	return v
}

// SetBorderStyle sets the border style.
func (v *VNode) SetBorderStyle(s layout.BorderStyle) *VNode {
	v.borderStyle = s
	v.composed = nil
	return v
}

// SetBorderColor sets the border color.
func (v *VNode) SetBorderColor(c style.Color) *VNode {
	v.borderColor = c
	v.composed = nil
	return v
}

// SetBorderLabel sets the border label.
func (v *VNode) SetBorderLabel(l string) *VNode {
	v.borderLabel = l
	v.composed = nil
	return v
}

// Rounded sets rounded border style.
func (v *VNode) Rounded() *VNode {
	return v.SetBorderStyle(layout.BorderRounded)
}

// Double sets double border style.
func (v *VNode) Double() *VNode {
	return v.SetBorderStyle(layout.BorderDouble)
}

// Single sets single border style.
func (v *VNode) Single() *VNode {
	return v.SetBorderStyle(layout.BorderSingle)
}

// NoBorder removes the border.
func (v *VNode) NoBorder() *VNode {
	return v.SetBorderStyle(layout.BorderNone)
}

// =============================================================================
// BoxModel Interface
// =============================================================================

// GetBorder returns BorderNone - Panel is a composition container.
// The actual border is handled by the internal Border component.
// This avoids double border calculation between Panel and Border.
func (v *VNode) GetBorder() layout.Border {
	return layout.Border{Style: layout.BorderNone}
}

// GetMargin returns zero margin.
func (v *VNode) GetMargin() layout.Margin {
	return layout.Margin{}
}

// GetPadding returns the panel padding.
func (v *VNode) GetPadding() layout.Padding {
	return layout.Padding{
		Top:    v.padding,
		Right:  v.padding,
		Bottom: v.padding,
		Left:   v.padding,
	}
}

// =============================================================================
// layout.BoxModelProvider Implementation
// =============================================================================

// GetBoxModel returns the box model for the Panel VNode.
// Implements layout.BoxModelProvider for unified padding/border handling.
// Note: Panel returns BorderNone to avoid double border calculation with internal Border.
func (v *VNode) GetBoxModel() layout.BoxModel {
	return layout.BoxModel{
		Padding: layout.Padding{
			Top:    v.padding,
			Right:  v.padding,
			Bottom: v.padding,
			Left:   v.padding,
		},
		Margin: layout.Margin{
			Left:   0,
			Right:  0,
			Top:    0,
			Bottom: 0,
		},
		Border: layout.Border{Style: layout.BorderNone},
	}
}
