package visualizer

import (
	"fmt"
	"strings"

	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Layout Tree Visualizer
// =============================================================================

// Visualizer visualizes the layout tree structure with constraints and dimensions.
type Visualizer struct {
	// nodes stores the state of each node in the layout tree
	nodes map[string]*NodeState

	// rootID is the root node ID
	rootID string
}

// NodeState represents the state of a layout node.
type NodeState struct {
	// ID is the node identifier
	ID string

	// Tag is the node type (e.g., "panel", "border", "text")
	Tag string

	// Bounds is the node's position and size
	Bounds layout.Rect

	// InputConstraints is the constraints passed to this node
	InputConstraints layout.Constraints

	// OutputConstraints is the constraints passed to children
	OutputConstraints layout.Constraints

	// Dimension is the measured size
	Dimension layout.Size

	// ParentID is the parent node ID
	ParentID string

	// Children IDs of children nodes
	Children []string

	// Additional metadata
	Props map[string]interface{}
}

// NewVisualizer creates a new layout visualizer.
func NewVisualizer() *Visualizer {
	return &Visualizer{
		nodes:  make(map[string]*NodeState),
		rootID: "",
	}
}

// AddNode adds a node to the visualizer.
func (v *Visualizer) AddNode(
	id string,
	tag string,
	bounds layout.Rect,
	inputConstraints layout.Constraints,
	outputConstraints layout.Constraints,
	dimension layout.Size,
	parentID string,
) {
	v.nodes[id] = &NodeState{
		ID:               id,
		Tag:              tag,
		Bounds:           bounds,
		InputConstraints: inputConstraints,
		OutputConstraints: outputConstraints,
		Dimension:        dimension,
		ParentID:         parentID,
		Children:         []string{},
		Props:            make(map[string]interface{}),
	}

	// Set as root if no parent
	if parentID == "" && v.rootID == "" {
		v.rootID = id
	}

	// Add to parent's children list
	if parentID != "" {
		if parent, exists := v.nodes[parentID]; exists {
			parent.Children = append(parent.Children, id)
		}
	}
}

// SetNodeProperty sets a property on a node.
func (v *Visualizer) SetNodeProperty(id string, key string, value interface{}) {
	if node, exists := v.nodes[id]; exists {
		if node.Props == nil {
			node.Props = make(map[string]interface{})
		}
		node.Props[key] = value
	}
}

// GetNode returns a node state by ID.
func (v *Visualizer) GetNode(id string) *NodeState {
	return v.nodes[id]
}

// PrintTree prints the layout tree to a string.
func (v *Visualizer) PrintTree() string {
	if v.rootID == "" {
		return "Empty layout tree"
	}

	var buf strings.Builder
	buf.WriteString("Layout Tree:\n")
	buf.WriteString("════════════\n\n")
	v.printRecursive(&buf, v.rootID, "")
	return buf.String()
}

func (v *Visualizer) printRecursive(buf *strings.Builder, nodeID string, indent string) {
	node := v.nodes[nodeID]
	if node == nil {
		return
	}

	// Print child's line with tree branch symbol first
	buf.WriteString(fmt.Sprintf("%s┌─ %s (%s)\n", indent, node.Tag, shortID(node.ID)))
	buf.WriteString(fmt.Sprintf("%s│  Position: (%d, %d)\n", indent, node.Bounds.X, node.Bounds.Y))
	buf.WriteString(fmt.Sprintf("%s│  Size: %dw x %dh\n", indent, node.Bounds.Width, node.Bounds.Height))
	buf.WriteString(fmt.Sprintf("%s│  Input: %s\n", indent, formatConstraints(node.InputConstraints)))
	if node.OutputConstraints != (layout.Constraints{}) {
		buf.WriteString(fmt.Sprintf("%s│  To Children: %s\n", indent, formatConstraints(node.OutputConstraints)))
	}

	// Print additional properties
	if len(node.Props) > 0 {
		buf.WriteString(fmt.Sprintf("%s│  Props:", indent))
		for k, val := range node.Props {
			buf.WriteString(fmt.Sprintf(" %s=%v", k, val))
		}
		buf.WriteString("\n")
	}

	// Print warning if dimension exceeds constraints
	if node.Dimension.Height > node.InputConstraints.MaxHeight &&
		node.InputConstraints.MaxHeight > 0 &&
		node.InputConstraints.MaxHeight < layout.MaxInt {
		buf.WriteString(fmt.Sprintf("%s│  ⚠️  Height %d exceeds MaxHeight %d\n",
			indent, node.Dimension.Height, node.InputConstraints.MaxHeight))
	}

	buf.WriteString("\n")

	// Print children
	for i, childID := range node.Children {
		isLast := i == len(node.Children)-1
		// For children, increase indentation with vertical lines
		newIndent := indent + "│  "
		if isLast {
			newIndent = indent + "   "
		}
		v.printRecursive(buf, childID, newIndent)
	}
}

// PrintConstraintChain prints the constraint propagation chain from root to a node.
func (v *Visualizer) PrintConstraintChain(nodeID string) string {
	node := v.nodes[nodeID]
	if node == nil {
		return fmt.Sprintf("Node %s not found", nodeID)
	}

	var chain []string

	// Walk up the tree from node to root
	currentID := nodeID
	for currentID != "" {
		currentNode := v.nodes[currentID]
		if currentNode == nil {
			break
		}

		chain = append(chain, fmt.Sprintf("%s (%s)", currentNode.Tag, shortID(currentNode.ID)))
		chain = append(chain, fmt.Sprintf("  Input: %s", formatConstraints(currentNode.InputConstraints)))
		if currentNode.OutputConstraints != (layout.Constraints{}) {
			chain = append(chain, fmt.Sprintf("  Output: %s", formatConstraints(currentNode.OutputConstraints)))
		}
		chain = append(chain, "  ↓")
		chain = append(chain, "")

		currentID = currentNode.ParentID
	}

	// Reverse to get root-to-leaf order
	for i, j := 0, len(chain)-1; i < j; i, j = i+1, j-1 {
		chain[i], chain[j] = chain[j], chain[i]
	}

	return strings.Join(chain, "\n")
}

// PrintSummary prints a summary of the layout tree.
func (v *Visualizer) PrintSummary() string {
	if v.rootID == "" {
		return "Empty layout tree"
	}

	var buf strings.Builder
	buf.WriteString("Layout Summary:\n")
	buf.WriteString("═══════════════\n\n")

	totalNodes := len(v.nodes)
	maxDepth := v.calculateDepth(v.rootID, 0)
	rootNode := v.nodes[v.rootID]

	buf.WriteString(fmt.Sprintf("Total Nodes: %d\n", totalNodes))
	buf.WriteString(fmt.Sprintf("Max Depth: %d\n", maxDepth))
	buf.WriteString(fmt.Sprintf("Root Size: %dw × %dh\n", rootNode.Bounds.Width, rootNode.Bounds.Height))
	buf.WriteString(fmt.Sprintf("Root Position: (%d, %d)\n", rootNode.Bounds.X, rootNode.Bounds.Y))
	buf.WriteString(fmt.Sprintf("Root Constraints: %s\n", formatConstraints(rootNode.InputConstraints)))

	// Count nodes by type
	typeCounts := make(map[string]int)
	for _, node := range v.nodes {
		typeCounts[node.Tag]++
	}

	buf.WriteString("\nNode Types:\n")
	for tag, count := range typeCounts {
		buf.WriteString(fmt.Sprintf("  %s: %d\n", tag, count))
	}

	return buf.String()
}

// FindProblems finds potential layout problems.
func (v *Visualizer) FindProblems() []string {
	var problems []string

	for _, node := range v.nodes {
		// Check if dimension exceeds constraints
		if node.Dimension.Height > node.InputConstraints.MaxHeight &&
			node.InputConstraints.MaxHeight > 0 &&
			node.InputConstraints.MaxHeight < layout.MaxInt {
			problems = append(problems,
				fmt.Sprintf("Node %s (%s): height %d exceeds MaxHeight %d",
					shortID(node.ID), node.Tag, node.Dimension.Height, node.InputConstraints.MaxHeight))
		}

		if node.Dimension.Width > node.InputConstraints.MaxWidth &&
			node.InputConstraints.MaxWidth > 0 &&
			node.InputConstraints.MaxWidth < layout.MaxInt {
			problems = append(problems,
				fmt.Sprintf("Node %s (%s): width %d exceeds MaxWidth %d",
					shortID(node.ID), node.Tag, node.Dimension.Width, node.InputConstraints.MaxWidth))
		}

		// Check if dimension is below minimum
		if node.Dimension.Width < node.InputConstraints.MinWidth {
			problems = append(problems,
				fmt.Sprintf("Node %s (%s): width %d below MinWidth %d",
					shortID(node.ID), node.Tag, node.Dimension.Width, node.InputConstraints.MinWidth))
		}

		if node.Dimension.Height < node.InputConstraints.MinHeight {
			problems = append(problems,
				fmt.Sprintf("Node %s (%s): height %d below MinHeight %d",
					shortID(node.ID), node.Tag, node.Dimension.Height, node.InputConstraints.MinHeight))
		}
	}

	return problems
}

// Clear clears all nodes from the visualizer.
func (v *Visualizer) Clear() {
	v.nodes = make(map[string]*NodeState)
	v.rootID = ""
}

// =============================================================================
// Visualizer Builder
// =============================================================================

// VisualizeVNode creates a visualization from a VNode tree.
func VisualizeVNode(vnode ui.VNode, constraints layout.Constraints) *Visualizer {
	vis := NewVisualizer()
	vis.buildFromVNode(vnode, constraints, "", 0, 0)
	return vis
}

func (v *Visualizer) buildFromVNode(
	vnode ui.VNode,
	constraints layout.Constraints,
	parentID string,
	x int,
	y int,
) string {
	if vnode == nil {
		return ""
	}

	// Get node key and tag
	nodeID := vnode.Key()
	if nodeID == "" {
		nodeID = fmt.Sprintf("node_%d", len(v.nodes))
	}
	tag := vnode.Tag()

	// Get dimensions from props if available
	props := vnode.Props()
	width := 0
	height := 0
	if props != nil {
		if w, ok := props["width"].(int); ok {
			width = w
		}
		if h, ok := props["height"].(int); ok {
			height = h
		}
	}

	// Measure if possible
	dimension := layout.Size{Width: width, Height: height}
	if measurable, ok := vnode.(interface{ Measure(layout.Constraints) layout.Size }); ok {
		dimension = measurable.Measure(constraints)
	}

	// Add node
	v.AddNode(
		nodeID,
		tag,
		layout.Rect{X: x, Y: y, Width: dimension.Width, Height: dimension.Height},
		constraints,
		layout.Constraints{}, // Will be filled by children
		dimension,
		parentID,
	)

	// Build children recursively
	children := vnode.Children()
	outputConstraints := constraints // Default: pass same constraints to children
	if len(children) > 0 {
		for _, child := range children {
			v.buildFromVNode(child, outputConstraints, nodeID, x, y)
		}
	}

	// Update output constraints after building children
	if node, ok := v.nodes[nodeID]; ok {
		node.OutputConstraints = outputConstraints
	}

	return nodeID
}

// =============================================================================
// Helper Functions
// =============================================================================

func formatConstraints(c layout.Constraints) string {
	if c.MaxWidth == layout.MaxInt && c.MaxHeight == layout.MaxInt {
		if c.MinWidth == 0 && c.MinHeight == 0 {
			return "Unbounded"
		}
		return fmt.Sprintf("Min: {%dw, %dh}", c.MinWidth, c.MinHeight)
	}
	return fmt.Sprintf("{%d..%d} x {%d..%d}",
		c.MinWidth, c.MaxWidth, c.MinHeight, c.MaxHeight)
}

func shortID(id string) string {
	if len(id) <= 8 {
		return id
	}
	return "..." + id[len(id)-8:]
}

func (v *Visualizer) calculateDepth(nodeID string, currentDepth int) int {
	node := v.nodes[nodeID]
	if node == nil || len(node.Children) == 0 {
		return currentDepth
	}

	maxChildDepth := currentDepth
	for _, childID := range node.Children {
		childDepth := v.calculateDepth(childID, currentDepth+1)
		if childDepth > maxChildDepth {
			maxChildDepth = childDepth
		}
	}

	return maxChildDepth
}
