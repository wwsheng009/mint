package runtime

import (
	"github.com/wwsheng009/mint/runtime/layout"
)

// BuildLayoutNodeTree converts a layout.Node interface tree to a runtime.LayoutNode struct tree.
// This is needed for the ActionRouter to perform Capture/Bubble phase traversal.
//
// The layout.Node interface is used by the layout engine, but ActionRouter needs
// the concrete runtime.LayoutNode struct to traverse Children[] for event propagation.
func BuildLayoutNodeTree(root layout.Node) *LayoutNode {
	if root == nil {
		return nil
	}

	return buildLayoutNodeRecursive(root, nil)
}

// buildLayoutNodeRecursive recursively converts layout.Node to runtime.LayoutNode
func buildLayoutNodeRecursive(node layout.Node, parent *LayoutNode) *LayoutNode {
	if node == nil {
		return nil
	}

	// Create the runtime.LayoutNode
	layoutNode := &LayoutNode{
		ID:   node.ID(),
		Type: NodeType(node.Type()), // Convert string to NodeType
		Parent: parent,
	}

	// Get position and size from the layout.Node
	x, y := node.GetPosition()
	width, height := node.GetSize()
	layoutNode.X = x
	layoutNode.Y = y
	layoutNode.MeasuredWidth = width
	layoutNode.MeasuredHeight = height

	// Recursively convert children
	children := node.Children()
	layoutNode.Children = make([]*LayoutNode, 0, len(children))

	for _, child := range children {
		childLayoutNode := buildLayoutNodeRecursive(child, layoutNode)
		if childLayoutNode != nil {
			layoutNode.Children = append(layoutNode.Children, childLayoutNode)
		}
	}

	return layoutNode
}

// BuildLayoutNodeTreeWithBounds converts a layout.Node tree to runtime.LayoutNode
// with explicit position/size overrides. This is useful when you have computed
// layout positions that differ from the node's current values.
func BuildLayoutNodeTreeWithBounds(
	root layout.Node,
	bounds map[string]struct{ X, Y, Width, Height int },
) *LayoutNode {
	if root == nil {
		return nil
	}

	return buildLayoutNodeRecursiveWithBounds(root, nil, bounds)
}

// buildLayoutNodeRecursiveWithBounds recursively converts with position overrides
func buildLayoutNodeRecursiveWithBounds(
	node layout.Node,
	parent *LayoutNode,
	bounds map[string]struct{ X, Y, Width, Height int },
) *LayoutNode {
	if node == nil {
		return nil
	}

	layoutNode := &LayoutNode{
		ID:     node.ID(),
		Type:   NodeType(node.Type()),
		Parent: parent,
	}

	// Use bounds if available, otherwise get from node
	if bound, ok := bounds[node.ID()]; ok {
		layoutNode.X = bound.X
		layoutNode.Y = bound.Y
		layoutNode.MeasuredWidth = bound.Width
		layoutNode.MeasuredHeight = bound.Height
	} else {
		x, y := node.GetPosition()
		width, height := node.GetSize()
		layoutNode.X = x
		layoutNode.Y = y
		layoutNode.MeasuredWidth = width
		layoutNode.MeasuredHeight = height
	}

	// Recursively convert children
	children := node.Children()
	layoutNode.Children = make([]*LayoutNode, 0, len(children))

	for _, child := range children {
		childLayoutNode := buildLayoutNodeRecursiveWithBounds(child, layoutNode, bounds)
		if childLayoutNode != nil {
			layoutNode.Children = append(layoutNode.Children, childLayoutNode)
		}
	}

	return layoutNode
}

// BuildLayoutNodeTreeFromHitMap creates a runtime.LayoutNode tree from a HitMap.
// This extracts the tree structure from the HitMap entries.
//
// Note: HitMap is in runtime/event package, so we pass the root layout.Node
// as a parameter to avoid circular imports.
func BuildLayoutNodeTreeFromHitMap(hitmapRoot layout.Node) *LayoutNode {
	if hitmapRoot == nil {
		return nil
	}

	return buildLayoutNodeRecursive(hitmapRoot, nil)
}
