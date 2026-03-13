// Package panel provides a Fiber-first Panel container component.
// Panel is a high-level container that manages borders, headers, and content layout.
// It is implemented using native Stack border properties (no border wrapper needed).
package panel

import (
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	newtext "github.com/wwsheng009/mint/ui/components/text"
)

// =============================================================================
// VNode - Composition-based Panel (no Instance needed)
// =============================================================================

// VNode is the panel container description.
// Panel is implemented as: VStack with native border properties (Border(VStack(...)) no longer needed)
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
	_ rtui.VNode = (*VNode)(nil)
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
// Composition - Build the VStack with native border properties
// =============================================================================

// getComposed builds and returns the composed node structure.
// Now uses VStack's native border properties instead of wrapping with Border component.
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

	// Build border label from title (if not explicitly set)
	borderLabel := v.borderLabel
	if borderLabel == "" && v.title != "" {
		borderLabel = " " + v.title + " "
	}

	// Convert layout.BorderStyle to string
	borderStyleStr := "none"
	switch v.borderStyle {
	case layout.BorderSingle:
		borderStyleStr = "single"
	case layout.BorderDouble:
		borderStyleStr = "double"
	case layout.BorderRounded:
		borderStyleStr = "rounded"
	case layout.BorderDashed:
		borderStyleStr = "dashed"
	}

	// Create VStack with border properties
	vstack := rtui.VStackBuilder().
		SetChildrenList(stackChildren).
		SetGap(0).
		SetBorder(borderStyleStr, borderLabel).
		SetBorderColor(v.borderColor)

	// 传递 width/height 给 VStack
	if v.width > 0 {
		vstack = vstack.SetWidth(v.width)
	}
	if v.height > 0 {
		vstack = vstack.SetHeight(v.height)
	}
	if v.flex > 0 {
		vstack = vstack.SetFlex(v.flex)
	}
	if v.instStyle.FG != "" || v.instStyle.BG != "" {
		vstack = vstack.SetStyleProps(v.instStyle)
	}

	// Build the LayoutNode to ensure border properties are set in Props
	v.composed = vstack.Build()
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
// The actual border is handled by the internal VStack component.
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
// Panel returns BorderNone because the internal VStack handles the border.
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

// ============================================================================
// VNode Enhanced API - Dimension Semantics
// ============================================================================

// SetOuterWidth sets the outer width including border padding.
// This is an alias for SetWidth(), providing semantic clarity.
func (v *VNode) SetOuterWidth(w int) *VNode {
	return v.SetWidth(w)
}

// SetInnerWidth sets the inner content width (excluding border padding).
// Automatically calculates outer width = innerWidth + borderPadding.
// This is easier than manually adding border padding.
func (v *VNode) SetInnerWidth(w int) *VNode {
	borderPadding := 0
	if v.borderStyle != layout.BorderNone {
		borderPadding = 2
	}
	v.width = w + borderPadding
	v.composed = nil
	return v
}

// SetContentWidth sets the content width (alias for SetInnerWidth).
func (v *VNode) SetContentWidth(w int) *VNode {
	return v.SetInnerWidth(w)
}

// SetOuterHeight sets the outer height including border padding.
// This is an alias for SetHeight(), providing semantic clarity.
func (v *VNode) SetOuterHeight(h int) *VNode {
	return v.SetHeight(h)
}

// SetInnerHeight sets the inner content height (excluding border padding).
// Automatically calculates outer height = innerHeight + borderPadding.
func (v *VNode) SetInnerHeight(h int) *VNode {
	borderPadding := 0
	if v.borderStyle != layout.BorderNone {
		borderPadding = 2
	}
	v.height = h + borderPadding
	v.composed = nil
	return v
}

// SetContentSize sets the content text line count.
// Automatically calculates height = lineCount + borderPadding.
func (v *VNode) SetContentSize(lineCount int) *VNode {
	return v.SetInnerHeight(lineCount)
}

// ============================================================================
// Convenience Methods for Content
// ============================================================================

// SetTextContent sets text content with auto wrap.
// Automatically creates a Text component with Wrap=true.
func (v *VNode) SetTextContent(content string) *VNode {
	textNode := newtext.New(content).SetWrap(true)
	v.content = textNode
	v.composed = nil
	return v
}

// SetWrappedTextContent sets wrapped text with fixed width.
// Automatically adjusts Panel width to accommodate the text with border padding.
func (v *VNode) SetWrappedTextContent(content string, width int) *VNode {
	v.SetContentWidth(width)
	textNode := newtext.New(content).SetWrap(true)
	v.content = textNode
	v.composed = nil
	return v
}

// SetPlainContent sets plain (non-wrapped) text content.
// Automatically creates a Text component with Wrap=false.
func (v *VNode) SetPlainContent(content string) *VNode {
	textNode := newtext.New(content).SetWrap(false)
	v.content = textNode
	v.composed = nil
	return v
}

// ============================================================================
// Dimension Query Methods
// ============================================================================

// GetOuterDimensions returns the outer dimensions (including border).
func (v *VNode) GetOuterDimensions() (width, height int) {
	return v.width, v.height
}

// GetInnerDimensions returns the inner content dimensions (excluding border).
func (v *VNode) GetInnerDimensions() (width, height int) {
	borderPadding := 0
	if v.borderStyle != layout.BorderNone {
		borderPadding = 2
	}
	innerWidth := max(0, v.width-borderPadding)
	innerHeight := max(0, v.height-borderPadding)
	return innerWidth, innerHeight
}

// GetContentWidth returns the content width (alias for GetInnerDimensions width).
func (v *VNode) GetContentWidth() int {
	w, _ := v.GetInnerDimensions()
	return w
}

// GetContentHeight returns the content height (alias for GetInnerDimensions height).
func (v *VNode) GetContentHeight() int {
	_, h := v.GetInnerDimensions()
	return h
}

// GetBorderPadding returns the total border padding added to dimensions.
func (v *VNode) GetBorderPadding() (widthPadding, heightPadding int) {
	padding := 0
	if v.borderStyle != layout.BorderNone {
		padding = 2
	}
	return padding, padding
}

// ============================================================================
// Combination Methods (Multiple Properties at Once)
// ============================================================================

// SetOuterSize sets both outer width and height.
func (v *VNode) SetOuterSize(w, h int) *VNode {
	return v.SetWidth(w).SetHeight(h)
}

// SetInnerSize sets both inner width and height.
func (v *VNode) SetInnerSize(w, h int) *VNode {
	return v.SetInnerWidth(w).SetInnerHeight(h)
}

// SetContentSize2D sets both content width and height.
func (v *VNode) SetContentSize2D(width, height int) *VNode {
	return v.SetInnerSize(width, height)
}

// ============================================================================
// Style-Aware Methods (adjust dimensions based on style)
// ============================================================================

// SetInnerWidthForStyle sets inner width but first updates border style if needed.
// Useful when you want to control dimensions and style together.
func (v *VNode) SetInnerWidthForStyle(w int, borderStyle layout.BorderStyle) *VNode {
	v.borderStyle = borderStyle
	return v.SetInnerWidth(w)
}

// SetInnerHeightForStyle sets inner height but first updates border style if needed.
func (v *VNode) SetInnerHeightForStyle(h int, borderStyle layout.BorderStyle) *VNode {
	v.borderStyle = borderStyle
	return v.SetInnerHeight(h)
}

// ============================================================================
// Builder-style Methods for Common Patterns
// ============================================================================

// FixedSize sets fixed outer dimensions.
func (v *VNode) FixedSize(w, h int) *VNode {
	return v.SetOuterSize(w, h)
}

// AutoHeight sets height to 0 (auto-measure from content).
func (v *VNode) AutoHeight() *VNode {
	return v.SetHeight(0)
}

// AutoWidth sets width to 0 (auto-measure from content).
func (v *VNode) AutoWidth() *VNode {
	return v.SetWidth(0)
}

// AutoSize sets both width and height to 0 (auto-measure from content).
func (v *VNode) AutoSize() *VNode {
	return v.AutoWidth().AutoHeight()
}

// FixedWidthAutoHeight sets fixed width but uses auto height.
func (v *VNode) FixedWidthAutoHeight(w int) *VNode {
	return v.SetWidth(w).AutoHeight()
}

// FixedHeightAutoWidth sets fixed height but uses auto width.
func (v *VNode) FixedHeightAutoWidth(h int) *VNode {
	return v.SetHeight(h).AutoWidth()
}

// FixedContentWidth sets fixed content width (accounting for border padding).
func (v *VNode) FixedContentWidth(w int) *VNode {
	return v.SetContentWidth(w)
}

// FixedContentHeight sets fixed content height (accounting for border padding).
func (v *VNode) FixedContentHeight(h int) *VNode {
	return v.SetContentSize(h)
}

// ============================================================================
// Convenience Functions for Common Use Cases
// ============================================================================

// InfoPanel creates an info panel with standard styling.
func InfoPanel(title, content string) *VNode {
	return New().
		SetTitle(title).
		SetBorderStyle(layout.BorderSingle).
		SetBorderColor(style.Color("blue")).
		SetTextContent(content)
}

// WarningPanel creates a warning panel with standard styling.
func WarningPanel(title, content string) *VNode {
	return New().
		SetTitle(title).
		SetBorderStyle(layout.BorderSingle).
		SetBorderColor(style.Color("yellow")).
		SetTextContent(content)
}

// ErrorPanel creates an error panel with standard styling.
func ErrorPanel(title, content string) *VNode {
	return New().
		SetTitle(title).
		SetBorderStyle(layout.BorderDouble).
		SetBorderColor(style.Color("red")).
		SetTextContent(content)
}

// SuccessPanel creates a success panel with standard styling.
func SuccessPanel(title, content string) *VNode {
	return New().
		SetTitle(title).
		SetBorderStyle(layout.BorderDouble).
		SetBorderColor(style.Color("green")).
		SetTextContent(content)
}

// TextPanel creates a simple panel with wrapped text and auto dimensions.
func TextPanel(title, content string, contentWidth int) *VNode {
	return New().
		SetTitle(title).
		SetBorderStyle(layout.BorderSingle).
		SetBorderColor(style.Color("blue")).
		SetWrappedTextContent(content, contentWidth)
}

// ============================================================================
// VNode Enhancement - With Methods (for optional parameters)
// ============================================================================

// WithTitle sets the title and returns the VNode.
func (v *VNode) WithTitle(title string) *VNode {
	return v.SetTitle(title)
}

// WithContent sets the content node and returns the VNode.
func (v *VNode) WithContent(content rtui.VNode) *VNode {
	return v.SetContent(content)
}

// WithHeader sets the header node and returns the VNode.
func (v *VNode) WithHeader(header rtui.VNode) *VNode {
	return v.SetHeader(header)
}

// WithFooter sets the footer node and returns the VNode.
func (v *VNode) WithFooter(footer rtui.VNode) *VNode {
	return v.SetFooter(footer)
}

// WithOuterDimensions sets outer dimensions and returns the VNode.
func (v *VNode) WithOuterDimensions(w, h int) *VNode {
	return v.SetOuterSize(w, h)
}

// WithInnerDimensions sets inner dimensions and returns the VNode.
func (v *VNode) WithInnerDimensions(w, h int) *VNode {
	return v.SetInnerSize(w, h)
}

// WithContentText sets text content with wrap and returns the VNode.
func (v *VNode) WithContentText(text string) *VNode {
	return v.SetTextContent(text)
}

// WithWrappedText sets wrapped text with width and returns the VNode.
func (v *VNode) WithWrappedText(text string, width int) *VNode {
	return v.SetWrappedTextContent(text, width)
}

// WithBorderStyleAndColor sets border style and color at once.
func (v *VNode) WithBorderStyleAndColor(style layout.BorderStyle, color style.Color) *VNode {
	return v.SetBorderStyle(style).SetBorderColor(color)
}

// ============================================================================
// Utility Functions
// ============================================================================

// helper function to get border padding for a given style.
func getBorderPaddingForStyle(borderStyle layout.BorderStyle) int {
	if borderStyle == layout.BorderNone {
		return 0
	}
	return 2
}

// CalculateOuterWidth calculates outer width from inner width and border style.
func CalculateOuterWidth(innerWidth int, borderStyle layout.BorderStyle) int {
	return innerWidth + getBorderPaddingForStyle(borderStyle)
}

// CalculateOuterHeight calculates outer height from inner height and border style.
func CalculateOuterHeight(innerHeight int, borderStyle layout.BorderStyle) int {
	return innerHeight + getBorderPaddingForStyle(borderStyle)
}

// CalculateInnerWidth calculates inner width from outer width and border style.
func CalculateInnerWidth(outerWidth int, borderStyle layout.BorderStyle) int {
	padding := getBorderPaddingForStyle(borderStyle)
	innerWidth := outerWidth - padding
	if innerWidth < 0 {
		innerWidth = 0
	}
	return innerWidth
}

// CalculateInnerHeight calculates inner height from outer height and border style.
func CalculateInnerHeight(outerHeight int, borderStyle layout.BorderStyle) int {
	padding := getBorderPaddingForStyle(borderStyle)
	innerHeight := outerHeight - padding
	if innerHeight < 0 {
		innerHeight = 0
	}
	return innerHeight
}
