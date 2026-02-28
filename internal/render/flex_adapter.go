package render

import (
	"github.com/wwsheng009/mint/runtime/layout"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// FlexLayoutAdapter - Creates FlexLayout from Fiber/VNode
// =============================================================================

// FlexLayoutAdapter creates layout.FlexLayout nodes from Fiber/VNode trees
// This handles the conversion of flex properties
//
// Deprecated: Use FiberToNodeAdapterPure with Fiber-first architecture instead.
// This adapter mixes VNode and Fiber data, which is against Fiber-first principles.
type FlexLayoutAdapter struct {
	*FiberToNodeAdapter
	style *layout.FlexStyle
}

// NewFlexLayoutAdapter creates a flex layout adapter
//
// Deprecated: Use NewFiberToNodeAdapterPure with Fiber-first architecture instead.
func NewFlexLayoutAdapter(fiber *rtui.Fiber, vnode rtui.VNode) *FlexLayoutAdapter {
	adapter := &FlexLayoutAdapter{
		FiberToNodeAdapter: NewFiberToNodeAdapter(fiber, vnode),
		style:              extractFlexStyle(fiber, vnode),
	}
	return adapter
}

// GetFlexStyle returns the flex style
func (a *FlexLayoutAdapter) GetFlexStyle() *layout.FlexStyle {
	return a.style
}