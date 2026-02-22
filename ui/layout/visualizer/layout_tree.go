package visualizer

// TreeNodeLayout stores layout information for a tree node
type TreeNodeLayout struct {
	NodeID   string
	Width    int  // Width of this node's subtree
	X        int  // Calculated X position
	Y        int  // Calculated Y position
	Depth    int  // Depth in the tree
	Children []*TreeNodeLayout
	Parent   *TreeNodeLayout
}

// TreeLayoutCalculator calculates tree layout positions
type TreeLayoutCalculator struct {
	v             *Visualizer
	nodeWidth     int
	nodeHeight    int
	horizontalGap int
	verticalGap   int
}

// NewTreeLayoutCalculator creates a new tree layout calculator
func NewTreeLayoutCalculator(v *Visualizer) *TreeLayoutCalculator {
	return &TreeLayoutCalculator{
		v:             v,
		nodeWidth:     160,
		nodeHeight:    94, // Default height for nodes with minimal content
		horizontalGap: 10,
		verticalGap:   30,
	}
}

// CalculateLayout performs a complete tree layout calculation
func (calc *TreeLayoutCalculator) CalculateLayout() *TreeNodeLayout {
	if calc.v.rootID == "" {
		return nil
	}

	// First pass: calculate widths and build tree structure (post-order)
	root := calc.calculateWidths(calc.v.rootID, 0, nil)

	// Second pass: assign positions (pre-order)
	calc.assignPositions(root, 0, 0)

	return root
}

// calculateWidths calculates the required width for each node's subtree (post-order)
func (calc *TreeLayoutCalculator) calculateWidths(nodeID string, depth int, parent *TreeNodeLayout) *TreeNodeLayout {
	node := calc.v.nodes[nodeID]
	if node == nil {
		return nil
	}

	layout := &TreeNodeLayout{
		NodeID: nodeID,
		Depth:  depth,
		Parent: parent,
	}

	if len(node.Children) == 0 {
		// Leaf node
		layout.Width = calc.nodeWidth
	} else {
		// Internal node - calculate width based on children
		parentChildrenWidth := 0
		for i, childID := range node.Children {
			childLayout := calc.calculateWidths(childID, depth+1, layout)
			if childLayout != nil {
				layout.Children = append(layout.Children, childLayout)
				if i > 0 {
					parentChildrenWidth += calc.horizontalGap
				}
				parentChildrenWidth += childLayout.Width
			}
		}
		// Parent width is max of its own width and total children width
		if parentChildrenWidth > calc.nodeWidth {
			layout.Width = parentChildrenWidth
		} else {
			layout.Width = calc.nodeWidth
		}
	}

	return layout
}

// assignPositions assigns (X, Y) coordinates to each node (pre-order)
func (calc *TreeLayoutCalculator) assignPositions(node *TreeNodeLayout, x int, y int) {
	if node == nil {
		return
	}

	// Set position for this node
	node.X = x
	node.Y = y

	if len(node.Children) == 0 {
		// Leaf node - already positioned
		return
	}

	// Position children
	// Calculate starting X to center children under parent
	totalChildrenWidth := 0
	for i, child := range node.Children {
		if i > 0 {
			totalChildrenWidth += calc.horizontalGap
		}
		totalChildrenWidth += child.Width
	}

	childX := x
	if totalChildrenWidth < node.Width {
		// Center children under parent
		childX = x + (node.Width-totalChildrenWidth)/2
	}

	// Assign positions to all children
	childY := y + calc.nodeHeight + calc.verticalGap
	for i, child := range node.Children {
		calc.assignPositions(child, childX, childY)

		// Move to next child position
		childX += child.Width
		if i < len(node.Children)-1 {
			childX += calc.horizontalGap
		}
	}
}

// GetTreeLayout returns the calculated tree layout
func (v *Visualizer) GetTreeLayout() *TreeNodeLayout {
	calc := NewTreeLayoutCalculator(v)
	return calc.CalculateLayout()
}
