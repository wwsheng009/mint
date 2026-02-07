package ui

// Internal helpers to set padding/margin props on any VNode
// These functions implement the CSS-like box model for TUI

func setPadding(vnode VNode, top, right, bottom, left int) {
	// First, try to use BoxModel interface if available
	if boxModel, ok := vnode.(interface {
		SetPadding(top, right, bottom, left int)
	}); ok {
		// Component has BoxModelMixin - use its setter
		boxModel.SetPadding(top, right, bottom, left)
		return
	}

	// Fallback: store in props for components not yet migrated
	props := vnode.Props()
	if props == nil {
		props = make(Props)
		vnode.SetProps(props)
	}
	props["padding"] = [4]int{top, right, bottom, left}
}

func setMargin(vnode VNode, top, right, bottom, left int) {
	// First, try to use BoxModel interface if available
	if boxModel, ok := vnode.(interface {
		SetMargin(top, right, bottom, left int)
	}); ok {
		// Component has BoxModelMixin - use its setter
		boxModel.SetMargin(top, right, bottom, left)
		return
	}

	// Fallback: store in props for components not yet migrated
	props := vnode.Props()
	if props == nil {
		props = make(Props)
		vnode.SetProps(props)
	}
	props["margin"] = [4]int{top, right, bottom, left}
}

// setTextAlign sets text alignment for any VNode
func setTextAlign(vnode VNode, align Align) {
	// First, try to use BoxModel interface if available
	if boxModel, ok := vnode.(interface {
		SetTextAlign(align Align)
	}); ok {
		// Component has BoxModelMixin - use its setter
		boxModel.SetTextAlign(align)
		return
	}

	// Fallback: store in props for components not yet migrated
	props := vnode.Props()
	if props == nil {
		props = make(Props)
		vnode.SetProps(props)
	}
	props["textAlign"] = int(align)
}

// =============================================================================
// Padding Helpers - Inner spacing
// =============================================================================

// Padding adds padding to any VNode (top, right, bottom, left)
// Returns the same VNode for chaining
// CSS equivalent: padding: top right bottom left
func Padding(vnode VNode, top, right, bottom, left int) VNode {
	setPadding(vnode, top, right, bottom, left)
	return vnode
}

// PaddingH sets horizontal padding (left, right)
// CSS equivalent: padding: 0 right 0 left
func PaddingH(vnode VNode, left, right int) VNode {
	return Padding(vnode, 0, right, 0, left)
}

// PaddingV sets vertical padding (top, bottom)
// CSS equivalent: padding: top 0 bottom 0
func PaddingV(vnode VNode, top, bottom int) VNode {
	return Padding(vnode, top, 0, bottom, 0)
}

// PaddingAll sets same padding on all sides
// CSS equivalent: padding: value
func PaddingAll(vnode VNode, p int) VNode {
	return Padding(vnode, p, p, p, p)
}

// =============================================================================
// Margin Helpers - Outer spacing
// =============================================================================

// Margin adds margin to any VNode (top, right, bottom, left)
// Returns the same VNode for chaining
// CSS equivalent: margin: top right bottom left
func Margin(vnode VNode, top, right, bottom, left int) VNode {
	setMargin(vnode, top, right, bottom, left)
	return vnode
}

// MarginH sets horizontal margin (left, right)
// CSS equivalent: margin: 0 right 0 left
func MarginH(vnode VNode, left, right int) VNode {
	return Margin(vnode, 0, right, 0, left)
}

// MarginV sets vertical margin (top, bottom)
// CSS equivalent: margin: top 0 bottom 0
func MarginV(vnode VNode, top, bottom int) VNode {
	return Margin(vnode, top, 0, bottom, 0)
}

// MarginAll sets same margin on all sides
// CSS equivalent: margin: value
func MarginAll(vnode VNode, m int) VNode {
	return Margin(vnode, m, m, m, m)
}

// =============================================================================
// Text Alignment Helper
// =============================================================================

// SetTextAlign sets text alignment for any VNode
// Note: Renamed from TextAlign to avoid conflict with ui.TextAlign() which creates text nodes
// CSS equivalent: text-align: left|center|right
func SetTextAlign(vnode VNode, align Align) VNode {
	setTextAlign(vnode, align)
	return vnode
}
