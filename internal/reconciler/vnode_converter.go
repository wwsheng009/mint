package reconciler

// =============================================================================
// VNode → runtime.LayoutNode Converter
// =============================================================================
// This module converts the declarative ui.VNode tree to the imperative
// runtime.LayoutNode IR used by the Layout Engine.
//
// Conversion Flow:
//   ui.VNode (declarative) → runtime.LayoutNode (IR) → Layout Engine → LayoutBox
// =============================================================================

import (
	"fmt"

	"github.com/wwsheng009/mint/runtime"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui"
) // Note: ui is still needed for component-specific VNode types (TextVNode, ButtonVNode, etc.)

// VNodeConverter converts ui.VNode trees to runtime.LayoutNode trees
type VNodeConverter struct {
	// nodeCounter generates unique IDs for LayoutNodes
	nodeCounter int
}

// NewVNodeConverter creates a new VNode converter
func NewVNodeConverter() *VNodeConverter {
	return &VNodeConverter{
		nodeCounter: 0,
	}
}

// Convert converts a ui.VNode tree to a runtime.LayoutNode tree
func (c *VNodeConverter) Convert(vnode rtui.VNode) *runtime.LayoutNode {
	if vnode == nil {
		return nil
	}
	return c.convertVNode(vnode, nil)
}

// convertVNode recursively converts a VNode to a LayoutNode
func (c *VNodeConverter) convertVNode(vnode ui.VNode, parent *runtime.LayoutNode) *runtime.LayoutNode {
	if vnode == nil {
		return nil
	}

	// Generate unique ID for this node
	id := c.generateID(vnode)

	// Convert based on VNode type
	switch v := vnode.(type) {
	case *ui.TextVNode:
		return c.convertText(v, parent, id)

	case *rtui.ComponentVNode:
		// Component nodes are expanded in the Fiber tree,
		// so we skip them here and convert children directly
		children := vnode.Children()
		if len(children) == 1 {
			return c.convertVNode(children[0], parent)
		}
		return c.convertFragment(vnode, parent, id)

	case *rtui.ElementVNode:
		return c.convertElement(v, parent, id)

	case *rtui.LayoutNode:
		return c.convertLayoutNode(v, parent, id)

	case *rtui.FragmentVNode:
		return c.convertFragment(v, parent, id)

	case *ui.ButtonVNode:
		return c.convertButton(v, parent, id)

	case *ui.InputVNode:
		return c.convertInput(v, parent, id)

	case *ui.TextareaVNode:
		return c.convertTextarea(v, parent, id)

	case *ui.CheckboxVNode:
		return c.convertCheckbox(v, parent, id)

	case *ui.SelectVNode:
		return c.convertSelect(v, parent, id)

	case *ui.ModalVNode:
		return c.convertModal(v, parent, id)

	case *ui.TabsVNode:
		return c.convertTabs(v, parent, id)

	case *ui.TableVNode:
		return c.convertTable(v, parent, id)

	case *ui.VirtualListVNode:
		return c.convertVirtualList(v, parent, id)

	case *ui.ProgressVNode:
		return c.convertProgress(v, parent, id)

	case *ui.SpinnerVNode:
		return c.convertSpinner(v, parent, id)

	default:
		// Fallback: try to determine type by tag/type name
		tag := getElementTag(vnode)
		return c.convertElementByTag(vnode, parent, id, tag)
	}
}

// generateID generates a unique ID for a LayoutNode
func (c *VNodeConverter) generateID(vnode ui.VNode) string {
	c.nodeCounter++
	typeName := getVNodeTypeName(vnode)
	if typeName == "" {
		typeName = "node"
	}
	return fmt.Sprintf("%s-%d", typeName, c.nodeCounter)
}

// getVNodeTypeName returns the type name of a VNode
func getVNodeTypeName(vnode ui.VNode) string {
	if vnode == nil {
		return ""
	}
	switch vnode.(type) {
	case *ui.TextVNode:
		return "text"
	case *ui.ButtonVNode:
		return "button"
	case *ui.InputVNode:
		return "input"
	case *ui.TextareaVNode:
		return "textarea"
	case *ui.CheckboxVNode:
		return "checkbox"
	case *ui.SelectVNode:
		return "select"
	case *ui.ModalVNode:
		return "modal"
	case *ui.TabsVNode:
		return "tabs"
	case *ui.TableVNode:
		return "table"
	case *ui.VirtualListVNode:
		return "virtuallist"
	case *ui.ProgressVNode:
		return "progress"
	case *ui.SpinnerVNode:
		return "spinner"
	case *rtui.LayoutNode:
		return "layout"
	case *rtui.ElementVNode:
		return getElementTag(vnode)
	case *rtui.ComponentVNode:
		return "component"
	case *rtui.FragmentVNode:
		return "fragment"
	default:
		return fmt.Sprintf("%T", vnode)
	}
}

// getElementTag extracts the tag from an ElementVNode
func getElementTag(vnode ui.VNode) string {
	if elem, ok := vnode.(*rtui.ElementVNode); ok {
		return elem.Tag()
	}
	// Try to get tag from Type() method
	typeStr := fmt.Sprintf("%v", vnode.Type())
	return typeStr
}

// =============================================================================
// Text Conversion
// =============================================================================

func (c *VNodeConverter) convertText(text *ui.TextVNode, parent *runtime.LayoutNode, id string) *runtime.LayoutNode {
	runtimeStyle := runtime.NewStyle()

	node := runtime.NewLayoutNode(id, runtime.NodeTypeText, runtimeStyle)
	node.Props = map[string]interface{}{
		"text": text.Content(),
	}

	// Store reference to original VNode for rendering
	node.Component = runtime.NewComponentRef(id, "text", text)

	if parent != nil {
		parent.AddChild(node)
	}

	return node
}

// =============================================================================
// Element Conversion
// =============================================================================

func (c *VNodeConverter) convertElement(elem *rtui.ElementVNode, parent *runtime.LayoutNode, id string) *runtime.LayoutNode {
	runtimeStyle := c.convertStyleFromProps(elem.Props())

	// Determine node type based on element tag
	nodeType := c.mapElementType(elem.Tag())

	node := runtime.NewLayoutNode(id, nodeType, runtimeStyle)

	// Copy props
	node.Props = make(map[string]interface{})
	for k, v := range elem.Props() {
		node.Props[k] = v
	}

	// Store reference to original VNode for rendering
	node.Component = runtime.NewComponentRef(id, elem.Tag(), elem)

	// Convert children
	for _, child := range elem.Children() {
		c.convertVNode(child, node)
	}

	if parent != nil {
		parent.AddChild(node)
	}

	return node
}

// convertElementByTag handles elements by their tag name
func (c *VNodeConverter) convertElementByTag(vnode ui.VNode, parent *runtime.LayoutNode, id, tag string) *runtime.LayoutNode {
	runtimeStyle := runtime.NewStyle()
	nodeType := c.mapElementType(tag)

	node := runtime.NewLayoutNode(id, nodeType, runtimeStyle)
	node.Props = map[string]interface{}{
		"tag": tag,
	}
	node.Component = runtime.NewComponentRef(id, tag, vnode)

	// Convert children
	for _, child := range vnode.Children() {
		c.convertVNode(child, node)
	}

	if parent != nil {
		parent.AddChild(node)
	}

	return node
}

// mapElementType maps element tag to runtime NodeType
func (c *VNodeConverter) mapElementType(tag string) runtime.NodeType {
	switch tag {
	case "hstack", "row":
		return runtime.NodeTypeRow
	case "vstack", "column":
		return runtime.NodeTypeColumn
	case "text":
		return runtime.NodeTypeText
	case "box", "container":
		return runtime.NodeTypeFlex
	default:
		return runtime.NodeTypeFlex
	}
}

// =============================================================================
// LayoutNode Conversion (ui.LayoutNode → runtime.LayoutNode)
// =============================================================================

func (c *VNodeConverter) convertLayoutNode(layout *rtui.LayoutNode, parent *runtime.LayoutNode, id string) *runtime.LayoutNode {
	runtimeStyle := runtime.NewStyle()

	// Map direction to node type and runtime style
	var nodeType runtime.NodeType
	switch layout.Direction() {
	case ui.DirectionRow:
		nodeType = runtime.NodeTypeRow
		runtimeStyle.Direction = runtime.DirectionRow
	case ui.DirectionColumn:
		nodeType = runtime.NodeTypeColumn
		runtimeStyle.Direction = runtime.DirectionColumn
	default:
		nodeType = runtime.NodeTypeFlex
		runtimeStyle.Direction = runtime.DirectionRow
	}

	node := runtime.NewLayoutNode(id, nodeType, runtimeStyle)

	// Copy layout props
	node.Props = map[string]interface{}{
		"gap":     layout.Gap(),
		"padding": layout.Padding(),
	}

	// Convert rtui.Align to runtime.Align/Justify
	align := layout.Align()
	runtimeStyle.AlignItems = mapUIAlignToRuntime(align)
	runtimeStyle.Justify = mapUIAlignToRuntimeJustify(align)

	// Set gap
	if gap := layout.Gap(); gap > 0 {
		runtimeStyle.Gap = gap
	}

	// Set padding
	padding := layout.Padding()
	runtimeStyle.Padding = runtime.Insets{
		Top:    padding[0],
		Right:  padding[1],
		Bottom: padding[2],
		Left:   padding[3],
	}

	// Store reference to original VNode
	node.Component = runtime.NewComponentRef(id, "layout", layout)

	// Convert children
	for _, child := range layout.Children() {
		c.convertVNode(child, node)
	}

	if parent != nil {
		parent.AddChild(node)
	}

	return node
}

// mapUIAlignToRuntime converts rtui.Align to runtime.Align
func mapUIAlignToRuntime(align rtui.Align) runtime.Align {
	switch align {
	case rtui.AlignStart:
		return runtime.AlignStart
	case rtui.AlignCenter:
		return runtime.AlignCenter
	case rtui.AlignEnd:
		return runtime.AlignEnd
	case rtui.AlignSpaceBetween:
		return runtime.AlignStart // No direct equivalent
	case rtui.AlignSpaceAround:
		return runtime.AlignStart // No direct equivalent
	default:
		return runtime.AlignStart
	}
}

// mapUIAlignToRuntimeJustify converts rtui.Align to runtime.Justify
func mapUIAlignToRuntimeJustify(align rtui.Align) runtime.Justify {
	switch align {
	case rtui.AlignStart:
		return runtime.JustifyStart
	case rtui.AlignCenter:
		return runtime.JustifyCenter
	case rtui.AlignEnd:
		return runtime.JustifyEnd
	case rtui.AlignSpaceBetween:
		return runtime.JustifyStart // No direct equivalent
	case rtui.AlignSpaceAround:
		return runtime.JustifyStart // No direct equivalent
	default:
		return runtime.JustifyStart
	}
}

// =============================================================================
// Fragment Conversion
// =============================================================================

func (c *VNodeConverter) convertFragment(frag ui.VNode, parent *runtime.LayoutNode, id string) *runtime.LayoutNode {
	// Fragments don't create their own node
	// Instead, they directly add children to parent
	children := frag.Children()
	if len(children) == 0 {
		return nil
	}

	// If no parent, create a flex container
	if parent == nil {
		runtimeStyle := runtime.NewStyle()
		node := runtime.NewLayoutNode(id, runtime.NodeTypeFlex, runtimeStyle)

		for _, child := range children {
			c.convertVNode(child, node)
		}

		return node
	}

	// Add children directly to parent
	for _, child := range children {
		c.convertVNode(child, parent)
	}

	return parent
}

// =============================================================================
// Component-Specific Conversions
// =============================================================================

func (c *VNodeConverter) convertButton(btn *ui.ButtonVNode, parent *runtime.LayoutNode, id string) *runtime.LayoutNode {
	runtimeStyle := runtime.NewStyle()

	node := runtime.NewLayoutNode(id, runtime.NodeTypeCustom, runtimeStyle)
	node.Props = map[string]interface{}{
		"label": btn.Label(),
		"type":  "button",
	}

	node.Component = runtime.NewComponentRef(id, "button", btn)

	if parent != nil {
		parent.AddChild(node)
	}

	return node
}

func (c *VNodeConverter) convertInput(input *ui.InputVNode, parent *runtime.LayoutNode, id string) *runtime.LayoutNode {
	runtimeStyle := runtime.NewStyle()
	// Default width for input
	runtimeStyle.Width = 20

	node := runtime.NewLayoutNode(id, runtime.NodeTypeCustom, runtimeStyle)
	node.Props = map[string]interface{}{
		"placeholder": input.Placeholder(),
		"value":       input.Value(),
		"type":        "input",
	}

	node.Component = runtime.NewComponentRef(id, "input", input)

	if parent != nil {
		parent.AddChild(node)
	}

	return node
}

func (c *VNodeConverter) convertTextarea(ta *ui.TextareaVNode, parent *runtime.LayoutNode, id string) *runtime.LayoutNode {
	runtimeStyle := runtime.NewStyle()
	// Default size for textarea
	runtimeStyle.Width = 40
	runtimeStyle.Height = 10

	node := runtime.NewLayoutNode(id, runtime.NodeTypeCustom, runtimeStyle)
	node.Props = map[string]interface{}{
		"placeholder": ta.Placeholder(),
		"value":       ta.Value(),
		"type":        "textarea",
	}

	node.Component = runtime.NewComponentRef(id, "textarea", ta)

	if parent != nil {
		parent.AddChild(node)
	}

	return node
}

func (c *VNodeConverter) convertCheckbox(cb *ui.CheckboxVNode, parent *runtime.LayoutNode, id string) *runtime.LayoutNode {
	runtimeStyle := runtime.NewStyle()

	node := runtime.NewLayoutNode(id, runtime.NodeTypeCustom, runtimeStyle)
	node.Props = map[string]interface{}{
		"label":   cb.Label(),
		"checked": cb.Checked(),
		"type":    "checkbox",
	}

	node.Component = runtime.NewComponentRef(id, "checkbox", cb)

	if parent != nil {
		parent.AddChild(node)
	}

	return node
}

func (c *VNodeConverter) convertSelect(sel *ui.SelectVNode, parent *runtime.LayoutNode, id string) *runtime.LayoutNode {
	runtimeStyle := runtime.NewStyle()

	node := runtime.NewLayoutNode(id, runtime.NodeTypeCustom, runtimeStyle)
	node.Props = map[string]interface{}{
		"options":  sel.Options(),
		"selected": sel.Selected(),
		"type":     "select",
	}

	node.Component = runtime.NewComponentRef(id, "select", sel)

	if parent != nil {
		parent.AddChild(node)
	}

	return node
}

func (c *VNodeConverter) convertModal(modal *ui.ModalVNode, parent *runtime.LayoutNode, id string) *runtime.LayoutNode {
	runtimeStyle := runtime.NewStyle()
	runtimeStyle.ZIndex = 1000 // Modals always on top

	// Set modal dimensions if available
	if w := modal.Width(); w > 0 {
		runtimeStyle.Width = w
	}
	if h := modal.Height(); h > 0 {
		runtimeStyle.Height = h
	}

	node := runtime.NewLayoutNode(id, runtime.NodeTypeCustom, runtimeStyle)
	node.Props = map[string]interface{}{
		"title":   modal.Title(),
		"visible": modal.IsOpen(),
		"type":    "modal",
	}

	node.Component = runtime.NewComponentRef(id, "modal", modal)

	// Convert modal content
	if content := modal.Content(); content != nil {
		c.convertVNode(content, node)
	}

	// Convert footer if present
	if footer := modal.Footer(); footer != nil {
		c.convertVNode(footer, node)
	}

	if parent != nil {
		parent.AddChild(node)
	}

	return node
}

func (c *VNodeConverter) convertTabs(tabs *ui.TabsVNode, parent *runtime.LayoutNode, id string) *runtime.LayoutNode {
	runtimeStyle := runtime.NewStyle()

	node := runtime.NewLayoutNode(id, runtime.NodeTypeCustom, runtimeStyle)
	node.Props = map[string]interface{}{
		"activeTab": tabs.ActiveTab(),
		"type":      "tabs",
	}

	node.Component = runtime.NewComponentRef(id, "tabs", tabs)

	// Convert children - tab content is stored as children
	for _, child := range tabs.Children() {
		c.convertVNode(child, node)
	}

	if parent != nil {
		parent.AddChild(node)
	}

	return node
}

func (c *VNodeConverter) convertTable(table *ui.TableVNode, parent *runtime.LayoutNode, id string) *runtime.LayoutNode {
	runtimeStyle := runtime.NewStyle()

	node := runtime.NewLayoutNode(id, runtime.NodeTypeCustom, runtimeStyle)
	node.Props = map[string]interface{}{
		"columns": table.Columns(),
		"rows":    table.Rows(),
		"type":    "table",
	}

	node.Component = runtime.NewComponentRef(id, "table", table)

	if parent != nil {
		parent.AddChild(node)
	}

	return node
}

func (c *VNodeConverter) convertVirtualList(vl *ui.VirtualListVNode, parent *runtime.LayoutNode, id string) *runtime.LayoutNode {
	runtimeStyle := runtime.NewStyle()

	// Set list height if available
	if h := vl.ListHeight(); h > 0 {
		runtimeStyle.Height = h
	}

	node := runtime.NewLayoutNode(id, runtime.NodeTypeCustom, runtimeStyle)
	node.Props = map[string]interface{}{
		"itemCount":  vl.ItemCount(),
		"itemHeight": vl.ItemHeight(),
		"type":       "virtuallist",
	}

	node.Component = runtime.NewComponentRef(id, "virtuallist", vl)

	if parent != nil {
		parent.AddChild(node)
	}

	return node
}

func (c *VNodeConverter) convertProgress(prog *ui.ProgressVNode, parent *runtime.LayoutNode, id string) *runtime.LayoutNode {
	runtimeStyle := runtime.NewStyle()

	// Set progress width if available
	if w := prog.Width(); w > 0 {
		runtimeStyle.Width = w
	}

	node := runtime.NewLayoutNode(id, runtime.NodeTypeCustom, runtimeStyle)
	node.Props = map[string]interface{}{
		"value":   prog.Value(),
		"max":     prog.Max(),
		"percent": prog.Percent(),
		"type":    "progress",
	}

	node.Component = runtime.NewComponentRef(id, "progress", prog)

	if parent != nil {
		parent.AddChild(node)
	}

	return node
}

func (c *VNodeConverter) convertSpinner(spinner *ui.SpinnerVNode, parent *runtime.LayoutNode, id string) *runtime.LayoutNode {
	runtimeStyle := runtime.NewStyle()

	node := runtime.NewLayoutNode(id, runtime.NodeTypeCustom, runtimeStyle)
	node.Props = map[string]interface{}{
		"type": "spinner",
	}

	node.Component = runtime.NewComponentRef(id, "spinner", spinner)

	if parent != nil {
		parent.AddChild(node)
	}

	return node
}

// =============================================================================
// Style Conversion
// =============================================================================

// convertStyleFromProps extracts layout properties from Props
func (c *VNodeConverter) convertStyleFromProps(props rtui.Props) runtime.Style {
	rs := runtime.NewStyle()

	// Extract width/height from props if present
	if w := props.GetInt("width"); w > 0 {
		rs.Width = w
	}
	if h := props.GetInt("height"); h > 0 {
		rs.Height = h
	}
	if flex := props.GetInt("flex"); flex > 0 {
		rs.FlexGrow = float64(flex)
	}

	return rs
}

// =============================================================================
// LayoutBox Generation
// =============================================================================

// GenerateLayoutBoxes generates LayoutBoxes from a LayoutNode tree
// This is called after layout calculation to produce the final layout boxes
func (c *VNodeConverter) GenerateLayoutBoxes(root *runtime.LayoutNode) []runtime.LayoutBox {
	if root == nil {
		return nil
	}

	boxes := make([]runtime.LayoutBox, 0)
	c.collectLayoutBoxes(root, &boxes)
	return boxes
}

// collectLayoutBoxes recursively collects layout boxes
func (c *VNodeConverter) collectLayoutBoxes(node *runtime.LayoutNode, boxes *[]runtime.LayoutBox) {
	if node == nil {
		return
	}

	// Only add boxes for nodes with components (leaf nodes)
	if node.Component != nil && node.Component.Instance != nil {
		box := runtime.NewLayoutBox(node)
		*boxes = append(*boxes, box)
	}

	// Recursively process children
	for _, child := range node.Children {
		c.collectLayoutBoxes(child, boxes)
	}
}
