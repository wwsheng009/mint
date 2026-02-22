package visualizer

import (
	"fmt"
	"html"
	"strings"

	"github.com/wwsheng009/mint/runtime/layout"
)

// PrintSVG prints the layout tree in SVG format.
func (v *Visualizer) PrintSVG() string {
	if v.rootID == "" {
		return v.printEmptySVG()
	}

	return v.printFullSVG()
}

// PrintSVGNestedBox prints layout as nested boxes (actual layout positions).
// This visualizes the actual spatial relationships, not just the tree structure.
func (v *Visualizer) PrintSVGNestedBox() string {
	if v.rootID == "" {
		return v.printEmptySVG()
	}

	return v.printNestedBoxSVG()
}

// printNestedBoxSVG prints nested box SVG showing actual layout positions.
func (v *Visualizer) printNestedBoxSVG() string {
	var buf strings.Builder

	// Find bounds of all nodes
	maxX, maxY := v.findMaxBounds()

	// Add padding
	padding := 20
	width := maxX + padding
	height := maxY + padding

	buf.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	buf.WriteString("\n")
	buf.WriteString(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d">`, width, height))
	buf.WriteString("\n")

	// CSS styles for nested box visualization
	buf.WriteString(`  <style>
    .node-box { fill: rgba(255,255,255,0.9); stroke: #666; stroke-width: 1; }
    .node-box-panel { fill: rgba(227,242,253,0.9); stroke: #1976D2; stroke-width: 2; }
    .node-box-border { fill: rgba(255,243,224,0.9); stroke: #F57C00; stroke-width: 2; }
    .node-box-text { fill: rgba(243,229,245,0.9); stroke: #7B1FA2; stroke-width: 2; }
    .node-box-stack-v { fill: rgba(232,245,233,0.9); stroke: #388E3C; stroke-width: 2; }
    .node-box-stack-h { fill: rgba(252,228,236,0.9); stroke: #C2185B; stroke-width: 2; }
    .node-label { font-family: monospace; font-size: 10px; fill: #333; }
    .node-label-pos { font-size: 8px; fill: #666; }
  </style>`)
	buf.WriteString("\n")

	// Background
	buf.WriteString(fmt.Sprintf(`  <rect x="0" y="0" width="%d" height="%d" fill="#fafafa"/>`, width, height))
	buf.WriteString("\n")

	// Title
	buf.WriteString(fmt.Sprintf(`  <text x="%d" y="20" text-anchor="middle" font-size="16" font-weight="bold" fill="#4CAF50">Nested Box Layout (Actual Positions)</text>`, width/2))
	buf.WriteString("\n")

	// Draw all nested boxes (in tree order, so parents draw first)
	v.drawNestedBoxes(&buf, v.rootID, 0)

	buf.WriteString("</svg>")
	return buf.String()
}

// drawNestedBoxes draws nested boxes showing actual layout positions.
func (v *Visualizer) drawNestedBoxes(buf *strings.Builder, nodeID string, depth int) {
	node := v.nodes[nodeID]
	if node == nil {
		return
	}

	// Use Bounds directly - ComputedBox positions are already absolute
	// in the layout engine coordinate space
	absX := node.Bounds.X
	absY := node.Bounds.Y

	// Node box class
	boxClass := v.getNodeBoxClass(node)

	// Adjust position for padding and title offset
	displayX := absX + 10
	displayY := absY + 30 // offset for title
	displayWidth := node.Bounds.Width
	displayHeight := node.Bounds.Height

	// Skip rendering zero-size boxes
	if displayWidth <= 0 || displayHeight <= 0 {
		return
	}

	// Draw node box (actual size and position from layout)
	buf.WriteString(fmt.Sprintf(`  <g transform="translate(%d, %d)">`+"\n", displayX, displayY))

	// Draw box with actual layout dimensions
	buf.WriteString(fmt.Sprintf(`    <rect x="0" y="0" width="%d" height="%d" class="%s" rx="3" ry="3"/>`+"\n",
		displayWidth, displayHeight, boxClass))

	// Add label inside the box (if large enough)
	if displayWidth >= 30 && displayHeight >= 20 {
		label := node.Tag
		if len(label) > 10 {
			label = label[:10]
		}
		buf.WriteString(fmt.Sprintf(`    <text x="5" y="12" class="node-label">%s</text>`+"\n", html.EscapeString(label)))
		buf.WriteString(fmt.Sprintf(`    <text x="5" y="22" class="node-label-pos">(%d,%d)</text>`+"\n",
			absX, absY))
	}

	buf.WriteString("  </g>\n")
	buf.WriteString("\n")

	// Recursively draw children (they will be drawn on top of parent)
	for _, childID := range node.Children {
		v.drawNestedBoxes(buf, childID, depth+1)
	}
}

// findMaxBounds finds the maximum extents of all nodes in the tree.
func (v *Visualizer) findMaxBounds() (maxX, maxY int) {
	maxX, maxY = 0, 0
	for _, node := range v.nodes {
		right := node.Bounds.X + node.Bounds.Width
		bottom := node.Bounds.Y + node.Bounds.Height
		if right > maxX {
			maxX = right
		}
		if bottom > maxY {
			maxY = bottom
		}
	}
	return maxX, maxY
}

// printEmptySVG prints an empty SVG document.
func (v *Visualizer) printEmptySVG() string {
	var buf strings.Builder

	buf.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	buf.WriteString("\n")
	buf.WriteString(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 800 600">`)
	buf.WriteString("\n")
	buf.WriteString(`  <style>
    .text { font-family: 'Segoe UI', sans-serif; font-size: 14px; fill: #333; }
    .title { font-size: 20px; font-weight: bold; fill: #4CAF50; }
    .node-box { fill: #fff; stroke: #ddd; stroke-width: 1; rx: 4; }
    .node-box-panel { fill: #e3f2fd; stroke: #2196F3; stroke-width: 2; }
    .node-box-border { fill: #fff3e0; stroke: #f57c00; stroke-width: 2; }
    .node-box-text { fill: #f3e5f5; stroke: #7b1fa2; stroke-width: 2; }
    .node-box-error { fill: #ffebee; stroke: #f44336; stroke-width: 2; stroke-dasharray: 4; }
  </style>`)
	buf.WriteString("\n")
	buf.WriteString(`  <text x="400" y="300" text-anchor="middle" class="title">Empty layout tree</text>`)
	buf.WriteString("\n")
	buf.WriteString("</svg>")

	return buf.String()
}

// printFullSVG prints the full SVG document with layout nodes.
func (v *Visualizer) printFullSVG() string {
	var buf strings.Builder

	// Calculate tree layout first
	treeLayout := v.GetTreeLayout()
	if treeLayout == nil {
		return v.printEmptySVG()
	}

	// Calculate overall size from tree layout
	maxWidth, maxHeight := v.calculateTreeDimensions(treeLayout)

	buf.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	buf.WriteString("\n")

	buf.WriteString(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d">`, maxWidth, maxHeight))
	buf.WriteString("\n")

	// CSS styles
	buf.WriteString(v.getSVGStyles())

	// Background
	buf.WriteString(fmt.Sprintf(`  <rect x="0" y="0" width="%d" height="%d" fill="#f5f5f5"/>`, maxWidth, maxHeight))
	buf.WriteString("\n")

	// Title
	buf.WriteString("  <text x=\"400\" y=\"30\" text-anchor=\"middle\" class=\"title\">Layout Tree Visualization</text>")
	buf.WriteString("\n")

	// Legend
	buf.WriteString(v.getSVGLegend())

	// Draw nodes using pre-calculated layout
	v.drawSVGTree(&buf, treeLayout, 50)

	buf.WriteString("</svg>")

	return buf.String()
}

// getSVGStyles returns the CSS styles for SVG.
func (v *Visualizer) getSVGStyles() string {
	var buf strings.Builder

	buf.WriteString(`  <style>
    .text { font-family: 'Segoe UI', sans-serif; font-size: 12px; fill: #333; }
    .title { font-size: 16px; font-weight: bold; fill: #4CAF50; }
    .legend-title { font-size: 14px; font-weight: bold; fill: #666; }
    .legend-text { font-size: 12px; fill: #555; }
    .node-title { font-size: 13px; font-weight: bold; fill: #333; }
    .node-id { font-size: 11px; fill: #999; }
    .node-size { font-size: 11px; fill: #666; font-family: monospace; }
    .node-constraint { font-size: 10px; fill: #666; font-family: monospace; }
    .connection-line { stroke: #bbb; stroke-width: 1; fill: none; }
    .node-box { fill: #fff; stroke: #ddd; stroke-width: 1; rx: 6; ry: 6; }
    .node-box-panel { fill: #e3f2fd; stroke: #2196F3; stroke-width: 2; rx: 6; ry: 6; }
    .node-box-border { fill: #fff3e0; stroke: #f57c00; stroke-width: 2; rx: 6; ry: 6; }
    .node-box-text { fill: #f3e5f5; stroke: #7b1fa2; stroke-width: 2; rx: 6; ry: 6; }
    .node-box-stack-v { fill: #e8f5e9; stroke: #388e3c; stroke-width: 2; rx: 6; ry: 6; }
    .node-box-stack-h { fill: #fce4ec; stroke: #c2185b; stroke-width: 2; rx: 6; ry: 6; }
    .node-box-grid { fill: #e0f7fa; stroke: #0097a7; stroke-width: 2; rx: 6; ry: 6; }
    .node-box-warning { fill: #fff3e0; stroke: #ff9800; stroke-width: 2; stroke-dasharray: 4; rx: 6; ry: 6; }
    .node-box-error { fill: #ffebee; stroke: #f44336; stroke-width: 2; stroke-dasharray: 4; rx: 6; ry: 6; }
  </style>`)
	buf.WriteString("\n")

	return buf.String()
}

// getSVGLegend returns a legend for the SVG.
func (v *Visualizer) getSVGLegend() string {
	var buf strings.Builder

	buf.WriteString("  <g transform=\"translate(50, 50)\">\n")
	buf.WriteString("    <!-- Legend box -->\n")
	buf.WriteString("    <rect x=\"0\" y=\"0\" width=\"140\" height=\"150\" fill=\"white\" stroke=\"#ddd\" rx=\"4\"/>\n")
	buf.WriteString("    <text x=\"70\" y=\"20\" text-anchor=\"middle\" class=\"legend-title\">Legend</text>\n")

	// Legend items
	items := []struct {
		class   string
		fill    string
		label   string
		yOffset int
	}{
		{"node-box-panel", "#e3f2fd", "Panel", 40},
		{"node-box-border", "#fff3e0", "Border", 60},
		{"node-box-text", "#f3e5f5", "Text", 80},
		{"node-box-stack-v", "#e8f5e9", "VStack", 100},
		{"node-box-stack-h", "#fce4ec", "HStack", 120},
		{"node-box-grid", "#e0f7fa", "Grid", 140},
	}

	for _, item := range items {
		buf.WriteString(fmt.Sprintf("    <rect x=\"20\" y=\"%d\" width=\"20\" height=\"15\" class=\"%s\"/>\n", item.yOffset, item.class))
		buf.WriteString(fmt.Sprintf("    <text x=\"50\" y=\"%d\" class=\"legend-text\">%s</text>\n", item.yOffset+12, item.label))
	}

	buf.WriteString("  </g>\n")
	buf.WriteString("\n")

	return buf.String()
}

// drawSVGTree draws the complete tree using pre-calculated layout
func (v *Visualizer) drawSVGTree(buf *strings.Builder, treeLayout *TreeNodeLayout, offsetX int) {
	if treeLayout == nil {
		return
	}

	// First pass: draw all connection lines
	v.drawTreeConnections(buf, treeLayout, offsetX)

	// Second pass: draw all nodes
	v.drawTreeNode(buf, treeLayout, offsetX)
}

// drawTreeConnections draws connection lines for all nodes
func (v *Visualizer) drawTreeConnections(buf *strings.Builder, node *TreeNodeLayout, offsetX int) {
	if node == nil || len(node.Children) == 0 {
		return
	}

	nodeData := v.nodes[node.NodeID]
	if nodeData == nil {
		return
	}

	nodeHeight := 80 + v.calculateNodeExtraHeight(nodeData)
	parentBottomX := offsetX + node.X + 160/2 // nodeWidth/2
	parentBottomY := 60 + node.Y + nodeHeight

	for _, child := range node.Children {
		childData := v.nodes[child.NodeID]
		if childData == nil {
			continue
		}

		childTopX := offsetX + child.X + 160/2
		childTopY := 60 + child.Y

		// Draw curved connection path
		controlY := parentBottomY + (childTopY-parentBottomY)/2
		buf.WriteString(fmt.Sprintf("  <path d=\"M %d,%d C %d,%d %d,%d %d,%d\" class=\"connection-line\"/>\n",
			parentBottomX, parentBottomY,
			parentBottomX, controlY,
			childTopX, controlY,
			childTopX, childTopY))

		// Recursively draw child connections
		v.drawTreeConnections(buf, child, offsetX)
	}
}

// drawTreeNode draws a single node and its children
func (v *Visualizer) drawTreeNode(buf *strings.Builder, node *TreeNodeLayout, offsetX int) {
	if node == nil {
		return
	}

	nodeData := v.nodes[node.NodeID]
	if nodeData == nil {
		return
	}

	// Draw this node
	v.drawSingleNode(buf, nodeData, offsetX+node.X, 60+node.Y)

	// Recursively draw children
	for _, child := range node.Children {
		v.drawTreeNode(buf, child, offsetX)
	}
}

// drawSingleNode draws a single node at the given position
func (v *Visualizer) drawSingleNode(buf *strings.Builder, node *NodeState, x int, y int) {
	// Calculate node dimensions
	nodeWidth := 160
	nodeHeight := 80 + v.calculateNodeExtraHeight(node)

	// Node box class
	boxClass := v.getNodeBoxClass(node)
	if v.hasConstraintError(node) {
		boxClass = "node-box-error"
	} else if v.hasConstraintWarning(node) {
		boxClass = "node-box-warning"
	}

	// Draw node box
	buf.WriteString(fmt.Sprintf(`  <g transform="translate(%d, %d)">`+"\n", x, y))

	buf.WriteString(fmt.Sprintf("    <rect x=\"0\" y=\"0\" width=\"%d\" height=\"%d\" class=\"%s\"/>\n", nodeWidth, nodeHeight, boxClass))

	// Node title (type)
	buf.WriteString(fmt.Sprintf("    <text x=\"10\" y=\"20\" class=\"node-title\">%s</text>\n", html.EscapeString(node.Tag)))

	// Node ID
	buf.WriteString(fmt.Sprintf("    <text x=\"%d\" y=\"20\" class=\"node-id\">%s</text>\n", nodeWidth-10, html.EscapeString(shortID(node.ID))))

	// Node size
	buf.WriteString(fmt.Sprintf("    <text x=\"10\" y=\"38\" class=\"node-size\">%dw × %dh</text>\n", node.Bounds.Width, node.Bounds.Height))

	// Node position
	buf.WriteString(fmt.Sprintf("    <text x=\"10\" y=\"54\" class=\"node-size\">pos: (%d, %d)</text>\n", node.Bounds.X, node.Bounds.Y))

	// Input constraints
	constStr := formatShortConstraints(node.InputConstraints)
	buf.WriteString(fmt.Sprintf("    <text x=\"10\" y=\"70\" class=\"node-constraint\">in: %s</text>\n", html.EscapeString(constStr)))

	// Output constraints (if present)
	currentY := 70
	if node.OutputConstraints != (layout.Constraints{}) {
		currentY += 14
		outputStr := formatShortConstraints(node.OutputConstraints)
		buf.WriteString(fmt.Sprintf("    <text x=\"10\" y=\"%d\" class=\"node-constraint\">out: %s</text>\n", currentY, html.EscapeString(outputStr)))
	}

	// Additional properties
	if len(node.Props) > 0 {
		currentY += 14
		propStr := v.formatNodeProps(node.Props, 140) // max width constraint
		buf.WriteString(fmt.Sprintf("    <text x=\"10\" y=\"%d\" class=\"node-constraint\">%s</text>\n", currentY, html.EscapeString(propStr)))
	}

	// Error indicator
	if v.hasConstraintError(node) {
		buf.WriteString(fmt.Sprintf("    <text x=\"%d\" y=\"%d\" class=\"node-constraint\" fill=\"#f44336\">!</text>\n", nodeWidth-25, nodeHeight-15))
	}

	buf.WriteString("  </g>\n")
	buf.WriteString("\n")
}

// calculateTreeDimensions calculates the overall dimensions from tree layout
func (v *Visualizer) calculateTreeDimensions(treeLayout *TreeNodeLayout) (width, height int) {
	width = max(treeLayout.Width+100, 800)
	height = 600

	// Calculate max depth and height
	maxDepth := v.calculateMaxDepth(treeLayout, 0)
	height = max(60 + (maxDepth+1)*(94+30) + 50, height)

	// Cap dimensions
	if width > 3000 {
		width = 3000
	}
	if height > 3000 {
		height = 3000
	}

	return width, height
}

// calculateMaxDepth calculates the maximum depth of the tree
func (v *Visualizer) calculateMaxDepth(node *TreeNodeLayout, currentDepth int) int {
	if node == nil {
		return currentDepth
	}

	maxDepth := currentDepth
	for _, child := range node.Children {
		childDepth := v.calculateMaxDepth(child, currentDepth+1)
		if childDepth > maxDepth {
			maxDepth = childDepth
		}
	}

	return maxDepth
}

// getNodeBoxClass returns the CSS class for a node based on its type.
func (v *Visualizer) getNodeBoxClass(node *NodeState) string {
	switch node.Tag {
	case "panel":
		return "node-box-panel"
	case "border", "bordered":
		return "node-box-border"
	case "text":
		return "node-box-text"
	case "vstack":
		return "node-box-stack-v"
	case "hstack":
		return "node-box-stack-h"
	case "grid":
		return "node-box-grid"
	default:
		return "node-box"
	}
}

// calculateNodeExtraHeight calculates additional height for constraints and props.
func (v *Visualizer) calculateNodeExtraHeight(node *NodeState) int {
	extraHeight := 0

	if node.OutputConstraints != (layout.Constraints{}) {
		extraHeight += 14
	}

	if len(node.Props) > 0 {
		extraHeight += 14
	}

	return extraHeight
}

// formatNodeProps formats node properties for display.
func (v *Visualizer) formatNodeProps(props map[string]interface{}, maxWidth int) string {
	if len(props) == 0 {
		return ""
	}

	var parts []string
	totalWidth := 0

	for k, val := range props {
		part := fmt.Sprintf("%s=%v", k, val)
		if totalWidth+len(part) > maxWidth && len(parts) > 0 {
			parts = append(parts, "...")
			break
		}
		parts = append(parts, part)
		totalWidth += len(part) + 2
	}

	return strings.Join(parts, ", ")
}

// formatShortConstraints formats constraints in a shorter format.
func formatShortConstraints(c layout.Constraints) string {
	if c.MaxWidth == layout.MaxInt && c.MaxHeight == layout.MaxInt {
		if c.MinWidth == 0 && c.MinHeight == 0 {
			return "unbounded"
		}
		return fmt.Sprintf(">={%d,%d}", c.MinWidth, c.MinHeight)
	}
	return fmt.Sprintf("{%d..%d}x{%d..%d}",
		c.MinWidth, c.MaxWidth, c.MinHeight, c.MaxHeight)
}

// PrintSVGSimple prints a simplified SVG tree diagram.
func (v *Visualizer) PrintSVGSimple() string {
	if v.rootID == "" {
		return v.printEmptySVG()
	}

	var buf strings.Builder

	// Use fixed size for simple diagrams
	maxWidth, maxHeight := 980, 620
	if len(v.nodes) > 10 {
		maxWidth = 2000
		maxHeight = 1200
	}

	buf.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	buf.WriteString("\n")
	buf.WriteString(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d">`, maxWidth, maxHeight))
	buf.WriteString("\n")

	buf.WriteString(v.getSimpleSVGStyles())

	// Draw simple nodes
	v.drawSimpleSVGNodes(&buf, v.rootID, 400, 80, maxWidth/2)

	buf.WriteString("</svg>")

	return buf.String()
}

// getSimpleSVGStyles returns simplified CSS styles.
func (v *Visualizer) getSimpleSVGStyles() string {
	return `  <style>
    .node-circle { fill: #fff; stroke: #333; stroke-width: 2; }
    .node-text { font-family: sans-serif; font-size: 11px; fill: #333; text-anchor: middle; }
    .link { stroke: #999; stroke-width: 1; fill: none; }
  </style>
` + "\n"
}

// drawSimpleSVGNodes draws nodes as simple circles.
func (v *Visualizer) drawSimpleSVGNodes(buf *strings.Builder, nodeID string, x, y, maxWidth int) {
	node := v.nodes[nodeID]
	if node == nil {
		return
	}

	// Draw link from parent (not for root)
	if node.ParentID != "" {
		parent := v.nodes[node.ParentID]
		if parent != nil {
			// Calculate parent position (simplified)
			buf.WriteString(fmt.Sprintf("  <line x1=\"%d\" y1=\"%d\" x2=\"%d\" y2=\"%d\" class=\"link\"/>\n",
				x, y-50, x, y))
		}
	}

	// Draw node circle
	buf.WriteString(fmt.Sprintf("  <circle cx=\"%d\" cy=\"%d\" r=\"25\" class=\"node-circle\"/>\n", x, y))

	// Draw node type (first 3 characters)
	label := node.Tag
	if len(label) > 3 {
		label = label[:3]
	}
	buf.WriteString(fmt.Sprintf("  <text x=\"%d\" y=\"%d\" class=\"node-text\">%s</text>\n", x, y+4, label))

	// Draw children
	if len(node.Children) > 0 {
		childY := y + 80
		totalWidth := len(node.Children) * 60
		startX := x - totalWidth/2 + 30

		for i, childID := range node.Children {
			childX := startX + i*60
			v.drawSimpleSVGNodes(buf, childID, childX, childY, maxWidth)
		}
	}
}

// PrintSVGTreeMap prints a tree-map style SVG visualization.
func (v *Visualizer) PrintSVGTreeMap() string {
	if v.rootID == "" {
		return v.printEmptySVG()
	}

	var buf strings.Builder
	maxWidth, maxHeight := 1000, 800

	buf.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	buf.WriteString("\n")
	buf.WriteString(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d">`, maxWidth, maxHeight))
	buf.WriteString("\n")

	buf.WriteString(`  <style>
    .treemap-rect { stroke: #fff; stroke-width: 2; }
    .treemap-text { font-family: sans-serif; font-size: 14px; fill: #fff; text-anchor: middle; }
  </style>
`)

	// Draw background
	buf.WriteString(fmt.Sprintf("  <rect x=\"0\" y=\"0\" width=\"%d\" height=\"%d\" fill=\"#333\"/>", maxWidth, maxHeight))
	buf.WriteString("\n")

	// Draw tree-map starting from position (20, 20)
	v.drawTreeMapNodes(&buf, v.rootID, 20, 20, maxWidth-40, maxHeight-40, 0)

	buf.WriteString("</svg>")

	return buf.String()
}

// drawTreeMapNodes draws nodes in tree-map style (nested rectangles).
func (v *Visualizer) drawTreeMapNodes(buf *strings.Builder, nodeID string, x, y, width, height int, depth int) {
	node := v.nodes[nodeID]
	if node == nil || width <= 0 || height <= 0 {
		return
	}

	// Get color based on node type
	color := v.getTreeMapColor(node.Tag, depth)

	// Draw rectangle
	buf.WriteString(fmt.Sprintf("  <rect x=\"%d\" y=\"%d\" width=\"%d\" height=\"%d\" fill=\"%s\" class=\"treemap-rect\"/>\n",
		x, y, width, height-4, color))

	// Draw label if large enough
	if width > 60 && height > 30 {
		buf.WriteString(fmt.Sprintf("  <text x=\"%d\" y=\"%d\" class=\"treemap-text\">%s</text>\n",
			x+width/2, y+height/2, html.EscapeString(node.Tag)))
	}

	// Draw children
	if len(node.Children) > 0 {
		// Simple layout: horizontal split
		childWidth := width / len(node.Children)
		childY := y + 20
		childHeight := height - 24

		for i, childID := range node.Children {
			v.drawTreeMapNodes(buf, childID, x+i*childWidth, childY, childWidth, childHeight, depth+1)
		}
	}
}

// getTreeMapColor returns a color for tree-map visualization.
func (v *Visualizer) getTreeMapColor(tag string, depth int) string {
	colors := map[string]string{
		"panel": "#2196F3",
		"border": "#f57c00",
		"bordered": "#f57c00",
		"text": "#7b1fa2",
		"vstack": "#388e3c",
		"hstack": "#c2185b",
		"grid": "#0097a7",
	}

	if color, ok := colors[tag]; ok {
		return color
	}

	return "#666"
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
