package visualizer

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/mattn/go-runewidth"
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
		ID:                id,
		Tag:               tag,
		Bounds:            bounds,
		InputConstraints:  inputConstraints,
		OutputConstraints: outputConstraints,
		Dimension:         dimension,
		ParentID:          parentID,
		Children:          []string{},
		Props:             make(map[string]interface{}),
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
// Box Model Visualization (Chrome DevTools style)
// =============================================================================

// PrintBoxModel prints the layout tree in box model style (like Chrome DevTools)
func (v *Visualizer) PrintBoxModel() string {
	if v.rootID == "" {
		return "Empty layout tree"
	}

	// 1. Calculate total dimensions (content width/height)
	maxWidth := v.getContentWidth(v.rootID)
	maxHeight := v.getContentHeight(v.rootID)

	// 2. Create buffer
	buffer := make([][]rune, maxHeight)
	for y := 0; y < maxHeight; y++ {
		buffer[y] = make([]rune, maxWidth)
		for x := 0; x < maxWidth; x++ {
			buffer[y][x] = ' '
		}
	}

	// 3. Recursively fill buffer: first draw parent box complete borders, children drawn inside parent
	v.fillBoxModelBuffer(buffer, v.rootID, 0, 0)

	// 4. Output buffer
	var buf strings.Builder
	buf.WriteString("Layout Box Model (Chrome DevTools style)\n")
	buf.WriteString(strings.Repeat("=", 50))
	buf.WriteString("\n\n")
	for y := 0; y < maxHeight; y++ {
		for x := 0; x < maxWidth; x++ {
			buf.WriteRune(buffer[y][x])
		}
		buf.WriteRune('\n')
	}

	buf.WriteString("\n" + strings.Repeat("=", 50) + "\n")
	buf.WriteString("[BORDER] = Border component with padding\n")
	buf.WriteString("[TEXT]   = Text content element\n")
	return buf.String()
}

// generateInfoLines generates actual info lines for a node
func (v *Visualizer) generateInfoLines(node *NodeState, lines *[]string) {
	line := fmt.Sprintf(" %s (%dw x %dh)", node.Tag, node.Bounds.Width, node.Bounds.Height)
	if node.Tag == "border" || node.Tag == "bordered" {
		line += " [BORDER]"
	} else if node.Tag == "text" {
		line += " [TEXT]"
	}
	*lines = append(*lines, line)

	line = fmt.Sprintf(" Pos: (%d,%d)", node.Bounds.X, node.Bounds.Y)
	constStr := formatConstraints(node.InputConstraints)
	if len(constStr)+len(line) < 60 {
		line += fmt.Sprintf("  %s", constStr)
	}
	*lines = append(*lines, line)

	if node.Dimension.Height > node.InputConstraints.MaxHeight &&
		node.InputConstraints.MaxHeight > 0 &&
		node.InputConstraints.MaxHeight < layout.MaxInt {
		*lines = append(*lines, fmt.Sprintf(" ⚠️  Height %d > MaxHeight %d", node.Dimension.Height, node.InputConstraints.MaxHeight))
	}
	if node.Dimension.Width > node.InputConstraints.MaxWidth &&
		node.InputConstraints.MaxWidth > 0 &&
		node.InputConstraints.MaxWidth < layout.MaxInt {
		*lines = append(*lines, fmt.Sprintf(" ⚠️  Width %d > MaxWidth %d", node.Dimension.Width, node.InputConstraints.MaxWidth))
	}
}

// PrintGrid prints a 2D grid representation of the layout
func (v *Visualizer) PrintGrid() string {
	if v.rootID == "" {
		return "Empty layout tree"
	}

	var buf strings.Builder
	buf.WriteString("Layout Grid (2D Visualization)\n")
	buf.WriteString(strings.Repeat("=", 50) + "\n\n")

	// Find grid dimensions
	maxX, maxY := 0, 0
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

	// Cap grid size
	if maxX > 40 {
		maxX = 40
	}
	if maxY > 20 {
		maxY = 20
	}

	// Create grid (2D array of node tags)
	grid := make([][]string, maxY)
	for y := 0; y < maxY; y++ {
		grid[y] = make([]string, maxX)
		for x := 0; x < maxX; x++ {
			grid[y][x] = " "
		}
	}

	nodeChars := map[string]string{
		"panel":    "█",
		"border":   "█",
		"bordered": "▓",
		"text":     "·",
		"vstack":   "║",
		"hstack":   "═",
		"button":   "▒",
	}

	// Fill grid
	for _, node := range v.nodes {
		char := "░"
		if c, ok := nodeChars[node.Tag]; ok {
			char = c
		}

		for y := node.Bounds.Y; y < node.Bounds.Y+node.Bounds.Height && y < maxY; y++ {
			for x := node.Bounds.X; x < node.Bounds.X+node.Bounds.Width && x < maxX; x++ {
				if y >= 0 && y < maxY && x >= 0 && x < maxX {
					// Prefer more specific characters
					if grid[y][x] == " " || (grid[y][x] == "░" && char != "░") {
						grid[y][x] = char
					}
				}
			}
		}
	}

	// Print grid with coordinates
	buf.WriteString("  ")
	for x := 0; x < maxX; x += 5 {
		buf.WriteString(fmt.Sprintf("%5d", x))
	}
	buf.WriteString("\n  " + strings.Repeat("─", maxX) + "\n")

	for y := 0; y < maxY; y++ {
		buf.WriteString(fmt.Sprintf("%2d│", y))
		for x := 0; x < maxX; x++ {
			buf.WriteString(grid[y][x])
		}
		buf.WriteString("│\n")
	}

	buf.WriteString("  " + strings.Repeat("─", maxX) + "\n\n")
	buf.WriteString("Legend:\n")
	buf.WriteString("  █ = panel/border     ║ = vstack    ░ = unknown\n")
	buf.WriteString("  ▓ = bordered         ═ = hstack    · = text\n")
	buf.WriteString("  ▒ = button\n")

	return buf.String()
}

// =============================================================================
// Visualizer Builder
// =============================================================================

// VisualizeVNode creates a visualization from a VNode tree.
// Note: This only creates a structural visualization with estimated sizes.
// For accurate layout data, use VisualizeFromLayoutEngine() instead.
func VisualizeVNode(vnode ui.VNode, constraints layout.Constraints) *Visualizer {
	vis := NewVisualizer()
	vis.buildFromVNode(vnode, constraints, "", 0, 0)
	return vis
}

// VisualizeFromLayoutEngine creates a visualization from a computed layout.
// This captures actual position and size data from the layout engine.
func VisualizeFromLayoutEngine(computedLayout interface{}) *Visualizer {
	vis := NewVisualizer()
	vis.buildFromComputedLayout(computedLayout)
	return vis
}

// buildFromComputedLayout builds the visualizer from a computed layout.
// Uses reflection to avoid import cycles between ui/layout and runtime/compute.
func (v *Visualizer) buildFromComputedLayout(layout interface{}) {
	// Use reflection to access ComputedLayout fields without direct import
	// This avoids circular dependency: ui/layout/visualizer -> runtime/compute
	if layout == nil {
		return
	}

	// Get root box via reflection
	rv := reflect.ValueOf(layout).Elem()
	rootBox := rv.FieldByName("Root")
	if !rootBox.IsValid() || rootBox.IsNil() {
		return
	}

	// Recursively build from the computed box tree
	v.buildFromComputedBox(rootBox.Interface(), "", 0)
}

// buildFromComputedBox builds a node from a ComputedBox
func (v *Visualizer) buildFromComputedBox(box interface{}, parentID string, depth int) string {
	// Use reflection to extract box data
	rv := reflect.ValueOf(box).Elem()

	// Extract tag/node type via VNode field
	vnodeField := rv.FieldByName("VNode")
	var tag string
	var nodeID string
	if vnodeField.IsValid() && !vnodeField.IsNil() {
		vnodeRV := reflect.ValueOf(vnodeField.Interface())
		// Try Tag() method
		if tagMethod := vnodeRV.MethodByName("Tag"); tagMethod.IsValid() {
			results := tagMethod.Call(nil)
			if len(results) > 0 && results[0].Kind() == reflect.String {
				tag = results[0].String()
			}
		}
		// Try Key() method
		if keyMethod := vnodeRV.MethodByName("Key"); keyMethod.IsValid() {
			results := keyMethod.Call(nil)
			if len(results) > 0 && results[0].Kind() == reflect.String {
				nodeID = results[0].String()
			}
		}
	}

	if nodeID == "" {
		nodeID = fmt.Sprintf("box_%d", len(v.nodes))
	}
	if tag == "" {
		tag = "unknown"
	}

	// Extract Box fields (X, Y, Width, Height)
	boxField := rv.FieldByName("Box")
	x := int(boxField.FieldByName("X").Int())
	y := int(boxField.FieldByName("Y").Int())
	width := int(boxField.FieldByName("Width").Int())
	height := int(boxField.FieldByName("Height").Int())

	// Create NodeState
	nodeState := &NodeState{
		ID:     nodeID,
		Tag:    tag,
		Bounds: layout.Rect{X: x, Y: y, Width: width, Height: height},
		// InputConstraints and OutputConstraints need more complex reflection
		// For now, use placeholder values
		InputConstraints:  layout.Constraints{MinWidth: 0, MaxWidth: layout.MaxInt, MinHeight: 0, MaxHeight: layout.MaxInt},
		OutputConstraints: layout.Constraints{MinWidth: 0, MaxWidth: layout.MaxInt, MinHeight: 0, MaxHeight: layout.MaxInt},
		Dimension:         layout.Size{Width: width, Height: height},
		ParentID:          parentID,
		Children:          []string{},
		Props:             make(map[string]interface{}),
	}

	// Extract NodeID if available (from ComputedBox.NodeID)
	nodeIDField := rv.FieldByName("NodeID")
	if nodeIDField.IsValid() {
		nodeState.Props["ComputedNodeID"] = uint64(nodeIDField.Uint())
	}

	// Add to visualizer
	v.nodes[nodeID] = nodeState

	// Set root
	if parentID == "" && v.rootID == "" {
		v.rootID = nodeID
	}

	// Add to parent's children
	if parentID != "" {
		if parent, exists := v.nodes[parentID]; exists {
			parent.Children = append(parent.Children, nodeID)
		}
	}

	// Process children
	childrenField := rv.FieldByName("Children")
	if childrenField.IsValid() && childrenField.Kind() == reflect.Slice {
		for i := 0; i < childrenField.Len(); i++ {
			childBox := childrenField.Index(i).Interface()
			if !reflect.ValueOf(childBox).IsNil() {
				v.buildFromComputedBox(childBox, nodeID, depth+1)
			}
		}
	}

	return nodeID
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

// =============================================================================
	// Box Model Visualization using Buffer
// =============================================================================

// fillBoxModelBuffer draws only borders into the buffer
func (v *Visualizer) fillBoxModelBuffer(buffer [][]rune, nodeID string, x, y int) {
	node := v.GetNode(nodeID)
	if node == nil {
		return
	}

	// Calculate box width
	boxWidth := v.getContentWidth(nodeID)

	// Calculate info lines count
	infoLines := []string{}
	v.generateInfoLines(node, &infoLines)
	infoLinesCount := len(infoLines)

	// Calculate box height: top border + info lines + empty line (if children) + children + bottom
	boxHeight := 2 + infoLinesCount // top + info + bottom
	if len(node.Children) > 0 {
		boxHeight += 1 // empty line before children
		for _, childID := range node.Children {
			childHeight := v.getContentHeight(childID)
			boxHeight += childHeight - 2 // Remove top/bottom borders of child
		}
	}

	// Draw top border
	if y < len(buffer) {
		for i := x; i < x+boxWidth && i < len(buffer[y]); i++ {
			if i == x {
				buffer[y][i] = '┌'
			} else if i == x+boxWidth-1 {
				buffer[y][i] = '┐'
			} else {
				buffer[y][i] = '─'
			}
		}
	}

	// Draw info lines with borders
	for i := 0; i < infoLinesCount; i++ {
		lineY := y + 1 + i
		if lineY >= len(buffer) {
			break
		}
		if x < len(buffer[lineY]) {
			buffer[lineY][x] = '│'
		}
		if x+boxWidth-1 < len(buffer[lineY]) {
			buffer[lineY][x+boxWidth-1] = '│'
		}
		// Draw content using runewidth to handle multi-byte characters
		cursor := x + 1
		for _, r := range infoLines[i] {
			if cursor < x+boxWidth-1 && cursor < len(buffer[lineY]) {
				charWidth := runewidth.RuneWidth(r)
				// Check if character fits
				if cursor+charWidth <= x+boxWidth-1 {
					buffer[lineY][cursor] = r
					// For wide characters (2 cells), we need to account for next cell
					if charWidth > 1 {
						cursor += charWidth - 1
					}
				}
				cursor++
			}
		}
	}

	// Draw children
	contentY := y + 1 + infoLinesCount
	if len(node.Children) > 0 {
		// Empty line before children
		if contentY < len(buffer) {
			if x < len(buffer[contentY]) {
				buffer[contentY][x] = '│'
			}
			if x+boxWidth-1 < len(buffer[contentY]) {
				buffer[contentY][x+boxWidth-1] = '│'
			}
		}
		contentY++

		// Draw children INSIDE parent box
		childX := x + 2 // Indent: │ then space

		for _, childID := range node.Children {
			// Recursively draw child box
			v.fillBoxModelBuffer(buffer, childID, childX, contentY)

			// Move to next child position
			childHeight := v.getContentHeight(childID)
			contentY += childHeight - 2 // Remove top/bottom borders
		}

		// Update boxHeight to match actual content
		boxHeight = contentY - y
	}

	// Draw left and right borders for all lines
	for lineY := y + 1; lineY < y+boxHeight-1; lineY++ {
		if lineY >= len(buffer) {
			break
		}
		if x < len(buffer[lineY]) {
			buffer[lineY][x] = '│'
		}
		if x+boxWidth-1 < len(buffer[lineY]) {
			buffer[lineY][x+boxWidth-1] = '│'
		}
	}

	// Draw bottom border
	bottomY := y + boxHeight - 1
	if bottomY >= 0 && bottomY < len(buffer) {
		for i := x; i < x+boxWidth && i < len(buffer[bottomY]); i++ {
			if i == x {
				buffer[bottomY][i] = '└'
			} else if i == x+boxWidth-1 {
				buffer[bottomY][i] = '┘'
			} else {
				buffer[bottomY][i] = '─'
			}
		}
	}
}


// getContentWidth calculates the content width for a node (including borders)
func (v *Visualizer) getContentWidth(nodeID string) int {
	node := v.GetNode(nodeID)
	if node == nil {
		return 54
	}

	// Calculate content width
	contentWidth := node.Bounds.Width
	if contentWidth < 20 {
		contentWidth = 20
	}
	if contentWidth > 60 {
		contentWidth = 60
	}

	// Find max info line width using runewidth
	infoLines := []string{}
	v.generateInfoLines(node, &infoLines)

	maxInfoWidth := 0
	for _, line := range infoLines {
		displayW := runewidth.StringWidth(line)
		if displayW > maxInfoWidth {
			maxInfoWidth = displayW
		}
	}

	if maxInfoWidth > contentWidth {
		contentWidth = maxInfoWidth
	}

	// Add border width (2 columns for left and right borders)
	boxWidth := contentWidth + 2 // +2 for │ borders

	// Check children - they need to fit inside with indentation
	for _, childID := range node.Children {
		childWidth := v.getContentWidth(childID)
		// Child needs to be indented by 2 (│ )
		childTotalWidth := childWidth + 2
		if childTotalWidth > boxWidth {
			boxWidth = childTotalWidth
		}
	}

	// Cap at 80
	if boxWidth > 80 {
		boxWidth = 80
	}

	return boxWidth
}

// getContentHeight calculates the total content height for a node
func (v *Visualizer) getContentHeight(nodeID string) int {
	node := v.GetNode(nodeID)
	if node == nil {
		return 20
	}

	// Generate info lines
	infoLines := []string{}
	v.generateInfoLines(node, &infoLines)

	// Base height: top border + empty + info lines + bottom border
	infoLinesCount := len(infoLines)
	baseHeight := 3 + infoLinesCount // top+info+bottom (no empty line for children yet)

	// Add children height
	if len(node.Children) > 0 {
		// Empty line after info
		baseHeight += 1

		// For each child, get their height (minus top/bottom borders they would have)
		for _, childID := range node.Children {
			childHeight := v.getContentHeight(childID)
			// Each child takes its total height
			baseHeight += childHeight - 2 // Remove top/bottom borders
		}

		// Add separators between children (excluding first)
		if len(node.Children) > 1 {
			baseHeight += len(node.Children) - 1
		}
	}

	// Add top and bottom borders
	baseHeight += 2

	// Cap at 100
	if baseHeight > 100 {
		baseHeight = 100
	}

	return baseHeight + 2 // +2 for top and bottom border
}

// createEmptyBuffer creates an empty buffer initialized with spaces
func (v *Visualizer) createEmptyBuffer(width, height int) [][]rune {
	buffer := make([][]rune, height)
	for y := 0; y < height; y++ {
		buffer[y] = make([]rune, width)
		for x := 0; x < width; x++ {
			buffer[y][x] = ' '
		}
	}
	return buffer
}

