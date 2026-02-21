package panel

import (
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	newborder "github.com/wwsheng009/mint/ui/components/border"
	newtext "github.com/wwsheng009/mint/ui/components/text"
)

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
	borderPadding := 2 * newborder.GetBorderWidth(v.borderStyle)
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
	borderPadding := 2 * newborder.GetBorderWidth(v.borderStyle)
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
	borderPadding := 2 * newborder.GetBorderWidth(v.borderStyle)
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
	padding := 2 * newborder.GetBorderWidth(v.borderStyle)
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
	return 2 * newborder.GetBorderWidth(borderStyle)
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
