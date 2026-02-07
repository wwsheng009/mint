package ui

// =============================================================================
// Box Model Interface - CSS-like box model for all components
// =============================================================================

// BoxModel defines the CSS box model contract
// Any component implementing this interface automatically supports:
// - Padding: inner spacing [top, right, bottom, left]
// - Margin: outer spacing [top, right, bottom, left]
// - TextAlign: text alignment within component
//
// Example:
//
//   type ButtonVNode struct {
//       *ui.ElementVNode
//       ui.BoxModelMixin  // Embed to automatically implement interface
//       // ... other fields ...
//   }
//
// Now ButtonVNode automatically satisfies the BoxModel interface!
type BoxModel interface {
	VNode

	// Padding returns the inner spacing [top, right, bottom, left]
	Padding() [4]int

	// Margin returns the outer spacing [top, right, bottom, left]
	Margin() [4]int

	// TextAlign returns the text alignment within the component
	TextAlign() Align
}

// =============================================================================
// Box Model Mixin - Default Implementation
// =============================================================================

// BoxModelMixin provides default implementation for box model properties
// Embed this in your component to automatically support padding, margin, and text align
//
// Example:
//
//   type ButtonVNode struct {
//       *ui.ElementVNode
//       ui.BoxModelMixin  // ← Embed this
//       label string
//       // ... other fields ...
//   }
type BoxModelMixin struct {
	padding   [4]int
	margin    [4]int
	textAlign Align
}

// Padding returns the inner spacing
func (b *BoxModelMixin) Padding() [4]int {
	return b.padding
}

// Margin returns the outer spacing
func (b *BoxModelMixin) Margin() [4]int {
	return b.margin
}

// TextAlign returns the text alignment
func (b *BoxModelMixin) TextAlign() Align {
	return b.textAlign
}

// =============================================================================
// Mixin Setters
// =============================================================================

// SetPadding sets the padding
func (b *BoxModelMixin) SetPadding(top, right, bottom, left int) {
	b.padding = [4]int{top, right, bottom, left}
}

// SetMargin sets the margin
func (b *BoxModelMixin) SetMargin(top, right, bottom, left int) {
	b.margin = [4]int{top, right, bottom, left}
}

// SetTextAlign sets the text alignment
func (b *BoxModelMixin) SetTextAlign(align Align) {
	b.textAlign = align
}

// =============================================================================
// Interface Helpers
// =============================================================================

// HasBoxModel checks if a VNode implements BoxModel interface
func HasBoxModel(vnode VNode) bool {
	_, ok := vnode.(BoxModel)
	return ok
}

// GetBoxModel safely gets BoxModel interface
// Returns nil if vnode doesn't implement BoxModel
func GetBoxModel(vnode VNode) BoxModel {
	if boxModel, ok := vnode.(BoxModel); ok {
		return boxModel
	}
	return nil
}
