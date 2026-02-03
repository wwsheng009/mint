package ui

// LayoutInfo describes the layout properties of a VNode.
// This is used by renderers to determine how to position children.
type LayoutInfo struct {
	// IsHorizontal indicates if the layout is horizontal (HStack)
	IsHorizontal bool
	// Gap is the spacing between children
	Gap int
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
		return info
	}

	// Check for ElementVNode with layout tags
	if elemNode, ok := vnode.(*ElementVNode); ok {
		tag := elemNode.Tag()
		if tag == "hstack" || tag == "row" {
			info.IsHorizontal = true
			// Try to get gap from props
			if g, ok := elemNode.Props()["gap"].(int); ok {
				info.Gap = g
			} else {
				info.Gap = 1 // Default gap for hstack
			}
		} else if tag == "vstack" || tag == "column" {
			info.IsHorizontal = false
			if g, ok := elemNode.Props()["gap"].(int); ok {
				info.Gap = g
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
