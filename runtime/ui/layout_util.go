package ui

// VNodeRenderer defines the contract for rendering VNodes.
// Both Fiber and non-Fiber rendering paths should implement this interface
// to ensure consistent behavior across different rendering modes.
//
// The interface is intentionally simple to allow flexibility in implementation
// while maintaining a clear contract for VNode rendering.
type VNodeRenderer interface {
	// Render renders a VNode tree to a buffer at the specified position.
	// The renderer is responsible for:
	// - Handling custom Paintable VNodes
	// - Rendering text content
	// - Applying layout (HStack/VStack positioning)
	// - Rendering children with proper spacing
	Render(vnode VNode, x, y int, buffer interface{})

	// Measure returns the width and height of a VNode.
	// This is used for layout calculations, particularly for HStack
	// horizontal positioning.
	Measure(vnode VNode) (width, height int)
}

// LayoutInfo describes the layout properties of a VNode.
// This is used by renderers to determine how to position children.
type LayoutInfo struct {
	// IsHorizontal indicates if the layout is horizontal (HStack)
	IsHorizontal bool
	// Gap is the spacing between children
	Gap int
	// Flex is the flex factor (0 = fixed size, >0 = grows to fill space)
	Flex int
	// Align is the main axis alignment
	Align Align
	// CrossAlign is the cross axis alignment
	CrossAlign Align
	// StretchCross makes all children stretch to fill cross axis
	StretchCross bool
	// Padding is the inner spacing [top, right, bottom, left]
	Padding [4]int
}

// GetLayoutInfo extracts layout information from a VNode.
// This is a shared utility used by both Fiber and non-Fiber rendering paths.
//
// It handles:
// - LayoutNode (from HStack/VStack)
// - ElementVNode with "hstack"/"row"/"vstack"/"column" tags
// - Any VNode with a Tag() method returning layout tags
func GetLayoutInfo(vnode VNode) LayoutInfo {
	info := LayoutInfo{
		IsHorizontal: false, // Default to vertical
		Gap:          0,     // Default gap
	}

	if vnode == nil {
		return info
	}

	// Check for LayoutNode (from ui.HStack, ui.VStack)
	if layoutNode, ok := vnode.(*LayoutNode); ok {
		info.IsHorizontal = layoutNode.Direction() == DirectionRow
		info.Gap = layoutNode.Gap()
		info.Flex = layoutNode.Flex()
		info.Align = layoutNode.Align()
		info.CrossAlign = layoutNode.CrossAlign()
		info.StretchCross = layoutNode.StretchCross()
		info.Padding = layoutNode.Padding()
		// Check props for flex override (from ui.Flex wrapper)
		if props := vnode.Props(); props != nil {
			if f, ok := props["flex"].(int); ok {
				info.Flex = f
			}
		}
		return info
	}

	// Check for BorderedNode (from ui.Bordered)
	if _, ok := vnode.(*BorderedNode); ok {
		// BorderedNode can have flex
		if props := vnode.Props(); props != nil {
			if f, ok := props["flex"].(int); ok {
				info.Flex = f
			}
		}
		return info
	}

	// Check for ElementVNode with layout tags
	if elemNode, ok := vnode.(*ElementVNode); ok {
		tag := elemNode.Tag()
		if tag == "hstack" || tag == "row" {
			info.IsHorizontal = true
			// Try to get gap from props
			if props := vnode.Props(); props != nil {
				if g, ok := props["gap"].(int); ok {
					info.Gap = g
				} else {
					info.Gap = 1 // Default gap for hstack
				}
				if f, ok := props["flex"].(int); ok {
					info.Flex = f
				}
				// Read align and crossAlign props for ElementVNode
				if a, ok := props["align"].(int); ok {
					info.Align = Align(a)
				}
				if a, ok := props["crossAlign"].(int); ok {
					info.CrossAlign = Align(a)
				}
			}
		} else if tag == "vstack" || tag == "column" {
			info.IsHorizontal = false
			if props := vnode.Props(); props != nil {
				if g, ok := props["gap"].(int); ok {
					info.Gap = g
				}
				if f, ok := props["flex"].(int); ok {
					info.Flex = f
				}
				// Read align and crossAlign props for ElementVNode
				if a, ok := props["align"].(int); ok {
					info.Align = Align(a)
				}
				if a, ok := props["crossAlign"].(int); ok {
					info.CrossAlign = Align(a)
				}
			}
		} else {
			// For other element types (bordered, etc.), check for flex prop
			if props := vnode.Props(); props != nil {
				if f, ok := props["flex"].(int); ok {
					info.Flex = f
				}
			}
		}
		return info
	}

	// Check for any VNode with a Tag() method (e.g., LayoutBuilder)
	if tagger, ok := vnode.(interface{ Tag() string }); ok {
		tag := tagger.Tag()
		if tag == "hstack" || tag == "row" {
			info.IsHorizontal = true
			// Try to get gap from Props
			if props := vnode.Props(); props != nil {
				if g, ok := props["gap"].(int); ok {
					info.Gap = g
				} else {
					info.Gap = 1 // Default gap for hstack
				}
				if f, ok := props["flex"].(int); ok {
					info.Flex = f
				}
				// Read align and crossAlign props
				if a, ok := props["align"].(int); ok {
					info.Align = Align(a)
				}
				if a, ok := props["crossAlign"].(int); ok {
					info.CrossAlign = Align(a)
				}
			}
		} else if tag == "vstack" || tag == "column" {
			info.IsHorizontal = false
			if props := vnode.Props(); props != nil {
				if g, ok := props["gap"].(int); ok {
					info.Gap = g
				}
				if f, ok := props["flex"].(int); ok {
					info.Flex = f
				}
				// Read align and crossAlign props
				if a, ok := props["align"].(int); ok {
					info.Align = Align(a)
				}
				if a, ok := props["crossAlign"].(int); ok {
					info.CrossAlign = Align(a)
				}
			}
		}
	}

	return info
}

// GetTextContent extracts text content from a VNode.
// This is a shared utility used by both Fiber and non-Fiber rendering paths.
//
// It handles:
// - VNodeText with "content" prop
// - VNodeElement with "content" prop (e.g., ui.Text elements)
//
// Returns the text content string, or empty string if no content is found.
func GetTextContent(vnode VNode) string {
	if vnode == nil {
		return ""
	}

	// Check for Content() method first (for TextVNode)
	if contenter, ok := vnode.(interface{ Content() string }); ok {
		return contenter.Content()
	}

	// Fall back to checking Props
	if props := vnode.Props(); props != nil {
		if content, ok := props["content"].(string); ok {
			return content
		}
	}

	return ""
}
