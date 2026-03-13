package panel

import (
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Builder - Fluent API
// =============================================================================

// Builder provides a fluent API for constructing Panel VNodes.
type Builder struct {
	vnode *VNode
}

// NewBuilder creates a new Panel builder.
func NewBuilder() *Builder {
	return &Builder{
		vnode: New(),
	}
}

// Key sets the component key.
func (b *Builder) Key(key string) *Builder {
	b.vnode.key = key
	return b
}

// SetID sets the business identifier for positioning and Portal anchoring.
// This is separate from Key() which is used for list diffing.
func (b *Builder) SetID(id string) *Builder {
	b.vnode.SetID(id)
	return b
}

// Title sets the panel title.
func (b *Builder) Title(title string) *Builder {
	b.vnode.SetTitle(title)
	return b
}

// Header sets the header component.
func (b *Builder) Header(header rtui.VNode) *Builder {
	b.vnode.SetHeader(header)
	return b
}

// Content sets the main content component.
func (b *Builder) Content(content rtui.VNode) *Builder {
	b.vnode.SetContent(content)
	return b
}

// Child sets the main content (alias for Content).
func (b *Builder) Child(child rtui.VNode) *Builder {
	return b.Content(child)
}

// Footer sets the footer component.
func (b *Builder) Footer(footer rtui.VNode) *Builder {
	b.vnode.SetFooter(footer)
	return b
}

// Width sets the width.
func (b *Builder) Width(w int) *Builder {
	b.vnode.SetWidth(w)
	return b
}

// Height sets the height.
func (b *Builder) Height(h int) *Builder {
	b.vnode.SetHeight(h)
	return b
}

// Size sets both width and height.
func (b *Builder) Size(w, h int) *Builder {
	return b.Width(w).Height(h)
}

// Flex sets the flex factor.
func (b *Builder) Flex(f int) *Builder {
	b.vnode.SetFlex(f)
	return b
}

// Padding sets the inner padding.
func (b *Builder) Padding(p int) *Builder {
	b.vnode.SetPadding(p)
	return b
}

// Style sets the style.
func (b *Builder) Style(s style.Style) *Builder {
	b.vnode.SetStyle(s)
	return b
}

// BorderStyle sets the border style.
func (b *Builder) BorderStyle(s layout.BorderStyle) *Builder {
	b.vnode.SetBorderStyle(s)
	return b
}

// BorderColor sets the border color.
func (b *Builder) BorderColor(c style.Color) *Builder {
	b.vnode.SetBorderColor(c)
	return b
}

// Color sets the border color (alias for BorderColor).
func (b *Builder) Color(c style.Color) *Builder {
	return b.BorderColor(c)
}

// BorderLabel sets the border label.
func (b *Builder) BorderLabel(l string) *Builder {
	b.vnode.SetBorderLabel(l)
	return b
}

// Label sets the border label (alias for BorderLabel).
func (b *Builder) Label(l string) *Builder {
	return b.BorderLabel(l)
}

// Rounded sets rounded border style.
func (b *Builder) Rounded() *Builder {
	b.vnode.Rounded()
	return b
}

// Double sets double border style.
func (b *Builder) Double() *Builder {
	b.vnode.Double()
	return b
}

// Single sets single border style.
func (b *Builder) Single() *Builder {
	b.vnode.Single()
	return b
}

// NoBorder removes the border.
func (b *Builder) NoBorder() *Builder {
	b.vnode.NoBorder()
	return b
}

// Build returns the configured VNode.
func (b *Builder) Build() rtui.VNode {
	return b.vnode
}

// BuildVNode returns the configured *VNode.
func (b *Builder) BuildVNode() *VNode {
	return b.vnode
}

// =============================================================================
// Convenience Functions
// =============================================================================

// Of creates a Panel with the given content.
func Of(content rtui.VNode) rtui.VNode {
	return NewBuilder().Content(content).Build()
}

// OfSize creates a Panel with explicit dimensions.
func OfSize(content rtui.VNode, width, height int) rtui.VNode {
	return NewBuilder().Content(content).Width(width).Height(height).Build()
}

// Titled creates a Panel with a title.
func Titled(title string, content rtui.VNode) rtui.VNode {
	return NewBuilder().Title(title).Content(content).Rounded().Build()
}

// Bordered creates a Panel with border.
func Bordered(content rtui.VNode, width, height int) rtui.VNode {
	return NewBuilder().Content(content).Width(width).Height(height).Rounded().Build()
}

// =============================================================================
// Fluent Global Functions
// =============================================================================

// Panel creates a new Panel builder.
func Panel() *Builder {
	return NewBuilder()
}

// WithTitle creates a titled Panel builder.
func WithTitle(title string) *Builder {
	return NewBuilder().Title(title).Rounded()
}

// HeaderFooter creates a Panel with header and footer.
func HeaderFooter(header, content, footer rtui.VNode) rtui.VNode {
	return NewBuilder().Header(header).Content(content).Footer(footer).Rounded().Build()
}

// ============================================================================
// Builder Enhanced API - Outer/Inner Dimension Methods
// ============================================================================

// OuterWidth sets the outer width (including border padding).
// Alias for Width() with clearer semantic meaning.
func (b *Builder) OuterWidth(w int) *Builder {
	return b.Width(w)
}

// OuterHeight sets the outer height (including border padding).
// Alias for Height() with clearer semantic meaning.
func (b *Builder) OuterHeight(h int) *Builder {
	return b.Height(h)
}

// OuterSize sets both outer width and height.
func (b *Builder) OuterSize(w, h int) *Builder {
	return b.Size(w, h)
}

// InnerWidth sets the inner content width (excluding border padding).
// Automatically calculates outer width = innerWidth + borderPadding.
func (b *Builder) InnerWidth(w int) *Builder {
	borderPadding := 0
	if b.vnode.borderStyle != layout.BorderNone {
		borderPadding = 2
	}
	b.vnode.SetWidth(w + borderPadding)
	return b
}

// ContentWidth sets the content width (alias for InnerWidth).
func (b *Builder) ContentWidth(w int) *Builder {
	return b.InnerWidth(w)
}

// InnerHeight sets the inner content height (excluding border padding).
// Automatically calculates outer height = innerHeight + borderPadding.
func (b *Builder) InnerHeight(h int) *Builder {
	borderPadding := 0
	if b.vnode.borderStyle != layout.BorderNone {
		borderPadding = 2
	}
	b.vnode.SetHeight(h + borderPadding)
	return b
}

// ContentHeight sets the content height in lines (alias for InnerHeight).
func (b *Builder) ContentHeight(h int) *Builder {
	return b.InnerHeight(h)
}

// ContentSize sets both content width and height.
func (b *Builder) ContentSize(w, h int) *Builder {
	return b.InnerWidth(w).InnerHeight(h)
}

// InnerSize sets both inner width and height (alias for ContentSize).
func (b *Builder) InnerSize(w, h int) *Builder {
	return b.ContentSize(w, h)
}

// ============================================================================
// Builder Convenience Methods for Auto Sizing
// ============================================================================

// AutoWidth sets width to 0 (auto-measure from content).
func (b *Builder) AutoWidth() *Builder {
	b.vnode.SetWidth(0)
	return b
}

// AutoHeight sets height to 0 (auto-measure from content).
func (b *Builder) AutoHeight() *Builder {
	b.vnode.SetHeight(0)
	return b
}

// AutoSize sets both width and height to 0 (auto-measure from content).
func (b *Builder) AutoSize() *Builder {
	return b.AutoWidth().AutoHeight()
}

// Fixed sets fixed outer dimensions.
func (b *Builder) Fixed(w, h int) *Builder {
	return b.Size(w, h)
}

// FixedInner sets fixed inner dimensions.
func (b *Builder) FixedInner(w, h int) *Builder {
	return b.InnerSize(w, h)
}

// FixedWidthAutoHeight sets fixed width but uses auto height.
func (b *Builder) FixedWidthAutoHeight(w int) *Builder {
	return b.Width(w).AutoHeight()
}

// FixedHeightAutoWidth sets fixed height but uses auto width.
func (b *Builder) FixedHeightAutoWidth(h int) *Builder {
	return b.Height(h).AutoWidth()
}

// ============================================================================
// Builder Convenience Methods for Text Content
// ============================================================================

// WithTextContent sets text content with auto wrap.
// Automatically creates a Text component with Wrap=true.
func (b *Builder) WithTextContent(text string) *Builder {
	b.vnode.SetTextContent(text)
	return b
}

// WithWrappedText sets wrapped text with fixed content width.
// Automatically adjusts Panel width to accommodate the text with border padding.
func (b *Builder) WithWrappedText(text string, width int) *Builder {
	b.ContentWidth(width)
	b.vnode.SetTextContent(text)
	return b
}

// WithTitle sets title for the panel (alias for Title).
func (b *Builder) WithTitle(title string) *Builder {
	return b.Title(title)
}

// WithPlainText sets plain (non-wrapped) text content.
func (b *Builder) WithPlainContent(text string) *Builder {
	b.vnode.SetPlainContent(text)
	return b
}

// TextPanel quickly creates a Panel with wrapped text and title.
func (b *Builder) TextPanel(title, text string, contentWidth int) *Builder {
	return b.Title(title).WithWrappedText(text, contentWidth)
}

// ============================================================================
// Builder Style-Aware Methods
// ============================================================================

// WithInnerWidthForStyle sets inner width after updating border style.
func (b *Builder) WithInnerWidthForStyle(w int, borderStyle layout.BorderStyle) *Builder {
	b.vnode.SetBorderStyle(borderStyle)
	return b.InnerWidth(w)
}

// WithInnerHeightForStyle sets inner height after updating border style.
func (b *Builder) WithInnerHeightForStyle(h int, borderStyle layout.BorderStyle) *Builder {
	b.vnode.SetBorderStyle(borderStyle)
	return b.InnerHeight(h)
}

// WithBorder sets both border style and color at once.
func (b *Builder) WithBorder(style layout.BorderStyle, color style.Color) *Builder {
	return b.BorderStyle(style).BorderColor(color)
}

// ============================================================================
// Builder Chain Methods for Common Patterns
// ============================================================================

// WithContentOnly sets only content, enabling auto sizing.
func (b *Builder) WithContentOnly(content rtui.VNode) *Builder {
	b.vnode.content = content
	return b
}

// WithHeaderContent sets header and content.
func (b *Builder) WithHeaderContent(header, content rtui.VNode) *Builder {
	return b.Header(header).Content(content)
}

// WithContentFooter sets content and footer.
func (b *Builder) WithContentFooter(content, footer rtui.VNode) *Builder {
	return b.Content(content).Footer(footer)
}

// WithFullContent sets header, content, and footer.
func (b *Builder) WithFullContent(header, content, footer rtui.VNode) *Builder {
	return b.Header(header).Content(content).Footer(footer)
}

// ============================================================================
// Enhanced Convenience Functions
// ============================================================================

// AutoContent creates a Panel with auto dimensions.
func AutoContent(content rtui.VNode) rtui.VNode {
	return NewBuilder().
		AutoSize().
		Content(content).
		Build()
}

// TitledAuto creates a titled Panel with auto dimensions.
func TitledAuto(title string, content rtui.VNode) rtui.VNode {
	return NewBuilder().
		Title(title).
		AutoSize().
		Content(content).
		Rounded().
		Build()
}

// Text creates a Panel with text content using auto wrap.
func Text(title, text string) rtui.VNode {
	return NewBuilder().
		Title(title).
		AutoSize().
		WithTextContent(text).
		Rounded().
		Build()
}

// TextSize creates a Panel with fixed content dimensions for wrapped text.
func TextSize(title, text string, width, height int) rtui.VNode {
	return NewBuilder().
		Title(title).
		ContentSize(width, height).
		WithWrappedText(text, width).
		Rounded().
		Build()
}

// TextWidth creates a Panel with fixed content width (auto height).
func TextWidth(title, text string, width int) rtui.VNode {
	return NewBuilder().
		Title(title).
		ContentWidth(width).
		AutoHeight().
		WithWrappedText(text, width).
		Rounded().
		Build()
}

// Info creates a standard info panel.
func Info(title, message string) rtui.VNode {
	return NewBuilder().
		Title(title).
		WithBorder(layout.BorderSingle, style.Color("blue")).
		WithTextContent(message).
		Build()
}

// Warning creates a warning panel.
func Warning(title, message string) rtui.VNode {
	return NewBuilder().
		Title(title).
		WithBorder(layout.BorderSingle, style.Color("yellow")).
		WithTextContent(message).
		Build()
}

// Error creates an error panel.
func Error(title, message string) rtui.VNode {
	return NewBuilder().
		Title(title).
		WithBorder(layout.BorderDouble, style.Color("red")).
		WithTextContent(message).
		Build()
}

// Success creates a success panel.
func Success(title, message string) rtui.VNode {
	return NewBuilder().
		Title(title).
		WithBorder(layout.BorderDouble, style.Color("green")).
		WithTextContent(message).
		Build()
}

// Box creates a simple bordered box with fixed outer dimensions.
func Box(content rtui.VNode, width, height int) rtui.VNode {
	return NewBuilder().
		Content(content).
		Fixed(width, height).
		Single().
		Build()
}

// BoxInner creates a bordered box with fixed inner dimensions.
func BoxInner(content rtui.VNode, innerWidth, innerHeight int) rtui.VNode {
	return NewBuilder().
		Content(content).
		FixedInner(innerWidth, innerHeight).
		Single().
		Build()
}

// BoxAuto creates a bordered box with auto dimensions.
func BoxAuto(content rtui.VNode) rtui.VNode {
	return NewBuilder().
		Content(content).
		AutoSize().
		Single().
		Build()
}

// ============================================================================
// Panel Presets (Quick Start Options)
// ============================================================================

// Minimal creates a minimal Panel with content only.
func Minimal(content rtui.VNode) rtui.VNode {
	return NewBuilder().
		Content(content).
		NoBorder().
		Build()
}

// Simple creates a simple Panel with content and single border.
func Simple(content rtui.VNode) rtui.VNode {
	return NewBuilder().
		Content(content).
		Single().
		Build()
}

// Card creates a card-style Panel with rounded border.
func Card(content rtui.VNode) rtui.VNode {
	return NewBuilder().
		Content(content).
		Rounded().
		Build()
}

// Modal creates a modal-style Panel with double border.
func Modal(content rtui.VNode) rtui.VNode {
	return NewBuilder().
		Content(content).
		Double().
		Build()
}

// ============================================================================
// Fluent Methods for Optional Components
// ============================================================================

// MaybeTitle sets title only if it's non-empty.
func (b *Builder) MaybeTitle(title string) *Builder {
	if title != "" {
		b.vnode.SetTitle(title)
	}
	return b
}

// MaybeHeader sets header only if it's not nil.
func (b *Builder) MaybeHeader(header rtui.VNode) *Builder {
	if header != nil {
		b.vnode.SetHeader(header)
	}
	return b
}

// MaybeFooter sets footer only if it's not nil.
func (b *Builder) MaybeFooter(footer rtui.VNode) *Builder {
	if footer != nil {
		b.vnode.SetFooter(footer)
	}
	return b
}

// MaybeBorder sets border only if style is set.
func (b *Builder) MaybeBorder(style layout.BorderStyle, color style.Color) *Builder {
	if style != layout.BorderNone {
		b.BorderStyle(style).BorderColor(color)
	}
	return b
}

// MaybePadding sets padding only if positive.
func (b *Builder) MaybePadding(p int) *Builder {
	if p > 0 {
		b.vnode.SetPadding(p)
	}
	return b
}

// ============================================================================
// Conditional Methods
// ============================================================================

// IfTitle conditionally sets title based on predicate.
func (b *Builder) IfTitle(title string, predicate func() bool) *Builder {
	if predicate() {
		b.vnode.SetTitle(title)
	}
	return b
}

// IfWidth conditionally sets width based on predicate.
func (b *Builder) IfWidth(w int, predicate func() bool) *Builder {
	if predicate() {
		b.Width(w)
	}
	return b
}

// IfHeight conditionally sets height based on predicate.
func (b *Builder) IfHeight(h int, predicate func() bool) *Builder {
	if predicate() {
		b.Height(h)
	}
	return b
}

// IfBorderStyle conditionally sets border style based on predicate.
func (b *Builder) IfBorderStyle(style layout.BorderStyle, predicate func() bool) *Builder {
	if predicate() {
		b.BorderStyle(style)
	}
	return b
}

// ============================================================================
// Builder Utility Functions
// ============================================================================

// NewPanelBuilder creates a new Panel builder with default settings.
// This is an alias for NewBuilder() for convenience.
func NewPanelBuilder() *Builder {
	return NewBuilder()
}

// BuildFromVNode creates a builder from an existing VNode.
// This allows modifying an existing panel configuration.
func BuildFromVNode(vnode *VNode) *Builder {
	return &Builder{vnode: vnode}
}

// ============================================================================
// Global Factory Functions (Return Builder for chaining)
// ============================================================================

// Colored returns a builder with border color set.
func Colored(color style.Color) *Builder {
	return NewBuilder().BorderColor(color)
}

// Styled returns a builder with both border style and color set.
func Styled(style layout.BorderStyle, color style.Color) *Builder {
	return NewBuilder().
		BorderStyle(style).
		BorderColor(color)
}

// FixedSize returns a builder with fixed dimensions set.
func FixedSize(w, h int) *Builder {
	return NewBuilder().Fixed(w, h)
}

// FixedContentSize returns a builder with fixed content dimensions set.
func FixedContentSize(w, h int) *Builder {
	return NewBuilder().ContentSize(w, h)
}

// Auto returns a builder with auto dimensions set.
func Auto() *Builder {
	return NewBuilder().AutoSize()
}
