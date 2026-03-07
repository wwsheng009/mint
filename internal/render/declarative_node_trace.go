package render

import (
	"fmt"
	"strings"

	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
)

// GetPortalRoots returns the portal box roots from the last layout computation
// Each portal root represents an independent tree structure
func (n *DeclarativeNode) GetPortalRoots() []*layout.LayoutBox {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if n.lastPortalBoxes == nil {
		return make([]*layout.LayoutBox, 0)
	}
	// Return a copy to avoid external modification
	result := make([]*layout.LayoutBox, len(n.lastPortalBoxes))
	copy(result, n.lastPortalBoxes)
	return result
}

// GetPortalTreeString returns the portal trees as string (multiple hierarchical structures)
func (n *DeclarativeNode) GetPortalTreeString() string {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if len(n.lastPortalBoxes) == 0 {
		return "No portal trees found!"
	}

	var sb strings.Builder
	sb.WriteString("Portal Trees (hierarchical):\n")
	sb.WriteString(strings.Repeat("=", 70))
	sb.WriteString(fmt.Sprintf("\nTotal Portals: %d\n\n", len(n.lastPortalBoxes)))

	// Each portal is an independent tree root
	for i, portalRoot := range n.lastPortalBoxes {
		sb.WriteString(fmt.Sprintf("=== Portal %d ===\n", i+1))
		buildLayoutTreeNodeString(portalRoot, 0, &sb)
		sb.WriteString("\n")
	}

	return sb.String()
}

// buildLayoutTreeNodeString recursively builds the string representation of a layout tree node
// Helper function for GetPortalTreeString
func buildLayoutTreeNodeString(box *layout.LayoutBox, depth int, sb *strings.Builder) {
	if box == nil {
		return
	}

	indent := strings.Repeat("  ", depth)
	propsID := box.PropsID
	if len(propsID) > 15 {
		propsID = propsID[:12] + "..."
	}
	if propsID == "" {
		propsID = "-"
	}

	// Append node with hierarchical relationship
	sb.WriteString(fmt.Sprintf("%s└─ [%s] %s (ID:%s, Size:%dx%d, Pos:%d,%d)\n",
		indent, box.Tag, propsID, box.ID, box.Width, box.Height, box.X, box.Y))

	// Append children recursively
	for _, child := range box.Children {
		buildLayoutTreeNodeString(child, depth+1, sb)
	}
}

// GetPortalBoxes returns the portal boxes from the last layout computation
// DEPRECATED: Use GetPortalRoots() instead to get tree structure
func (n *DeclarativeNode) GetPortalBoxes() []layout.LayoutBox {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if n.lastPortalBoxes == nil {
		return make([]layout.LayoutBox, 0)
	}
	// Convert pointer slice to value slice for backward compatibility
	result := make([]layout.LayoutBox, len(n.lastPortalBoxes))
	for i, box := range n.lastPortalBoxes {
		result[i] = *box
	}
	return result
}


// GetPaintableTreeString returns the paintable tree as string (hierarchical structure)
func (n *DeclarativeNode) GetPaintableTreeString() string {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if n.lastPaintableRoot != nil {
		return n.lastPaintableRoot.TreeString()
	}
	return "No paintable tree found!"
}

// GetPaintableBoxes returns the flattened paintable boxes from the last paint computation
// DEPRECATED: Use GetPaintableRoot() instead to get tree structure
// This provides flattened list for backward compatibility
func (n *DeclarativeNode) GetPaintableBoxes() []*paint.PaintableBox {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if n.lastPaintableRoot == nil {
		return make([]*paint.PaintableBox, 0)
	}

	// Collect all paintable boxes recursively
	boxes := make([]*paint.PaintableBox, 0)
	var collect func(box *paint.PaintableBox)
	collect = func(box *paint.PaintableBox) {
		if box == nil {
			return
		}
		boxes = append(boxes, box)
		for _, child := range box.Children {
			collect(child)
		}
	}
	collect(n.lastPaintableRoot)
	return boxes
}


// GetLayoutRoot returns the root of the layout tree from the last layout computation
// This provides access to the computed hierarchical layout (tree structure)
func (n *DeclarativeNode) GetLayoutRoot() *layout.LayoutBox {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if n.lastLayoutResult != nil && n.lastLayoutResult.Root != nil {
		return n.lastLayoutResult.Root
	}
	return nil
}

// GetLayoutTreeString returns the layout tree as string (hierarchical structure)
func (n *DeclarativeNode) GetLayoutTreeString() string {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if n.lastLayoutResult != nil {
		return n.lastLayoutResult.TreeString()
	}
	return "No layout tree found!"
}

// GetLayoutBoxes returns the flattened layout boxes from the last layout computation
// DEPRECATED: Use GetLayoutRoot() instead to get tree structure
// This provides flattened list for backward compatibility
func (n *DeclarativeNode) GetLayoutBoxes() []*layout.LayoutBox {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if n.lastLayoutResult != nil && n.lastLayoutResult.Root != nil {
		// Collect all boxes recursively
		boxes := make([]*layout.LayoutBox, 0)
		var collect func(box *layout.LayoutBox)
		collect = func(box *layout.LayoutBox) {
			if box == nil {
				return
			}
			boxes = append(boxes, box)
			for _, child := range box.Children {
				collect(child)
			}
		}
		collect(n.lastLayoutResult.Root)
		return boxes
	}
	return make([]*layout.LayoutBox, 0)
}
