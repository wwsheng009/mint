package dsl

import (
	"fmt"

	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/grid"
	"github.com/wwsheng009/mint/ui/components/panel"
	"github.com/wwsheng009/mint/ui/components/text"
)

// =============================================================================
// Layout DSL - Declarative Layout Building
// =============================================================================

// Node represents a declarative layout node.
type Node struct {
	tag      string
	props    ui.Props
	children []Node
}

// Tag returns the node type.
func (n *Node) Tag() string {
	return n.tag
}

// Props returns the node properties.
func (n *Node) Props() ui.Props {
	return n.props
}

// Children returns the child nodes.
func (n *Node) Children() []Node {
	return n.children
}

// ToVNode converts the DSL Node to a VNode.
func (n *Node) ToVNode() ui.VNode {
	// Convert children
	vnodeChildren := make([]ui.VNode, len(n.children))
	for i, child := range n.children {
		vnodeChildren[i] = child.ToVNode()
	}

	// Create TextVNode for text nodes
	if n.tag == "text" {
		content := n.props.GetString("content")
		return text.New(content)
	}

	// Create Panel for panel nodes
	if n.tag == "panel" {
		p := panel.NewBuilder()
		if title, ok := n.props["title"].(string); ok {
			p.Title(title)
		}
		if width, ok := n.props["width"].(int); ok {
			p.Width(width)
		}
		if height, ok := n.props["height"].(int); ok {
			p.Height(height)
		}
		if flex, ok := n.props["flex"].(int); ok {
			p.Flex(flex)
		}
		if pad, ok := n.props["padding"].(int); ok {
			p.Padding(pad)
		}
		if len(vnodeChildren) > 0 {
			p.Content(vnodeChildren[0])
		}
		return p.Build()
	}

	// For row/column, create a panel with layout tag marker
	if n.tag == "row" || n.tag == "column" {
		p := panel.NewBuilder()
		if width, ok := n.props["width"].(int); ok {
			p.Width(width)
		}
		if height, ok := n.props["height"].(int); ok {
			p.Height(height)
		}
		if flex, ok := n.props["flex"].(int); ok {
			p.Flex(flex)
		}
		if len(vnodeChildren) > 0 {
			p.Content(vnodeChildren[0])
		}
		result := p.Build()
		// Mark with props for consumers that care about layout type
		if result.Props() == nil {
			result.SetProps(make(ui.Props))
		}
		result.Props()["_layout"] = n.tag
		return result
	}

	// Create Grid for grid nodes
	if n.tag == "grid" {
		g := grid.New()

		// Set dimensions if specified
		if columns, ok := n.props["columns"].([]grid.Dimension); ok {
			g.SetColumns(columns...)
		}
		if rows, ok := n.props["rows"].([]grid.Dimension); ok {
			g.SetRows(rows...)
		}

		// Set sizing props
		if width, ok := n.props["width"].(int); ok {
			g.SetWidth(width)
		}
		if height, ok := n.props["height"].(int); ok {
			g.SetHeight(height)
		}
		if flex, ok := n.props["flex"].(int); ok {
			g.SetFlex(flex)
		}

		// Set gap
		if colGap, ok := n.props["columnGap"].(int); ok {
			if rowGap, ok := n.props["rowGap"].(int); ok {
				g.SetGap(colGap, rowGap)
			}
		}
		if gap, ok := n.props["gap"].(int); ok {
			g.SetGap(gap, gap)
		}

		// Set children - auto-position in row-major order
		if len(vnodeChildren) > 0 {
			g.SetChildrenAuto(vnodeChildren)
		}

		return g
	}

	// Default: return nil or create a placeholder
	return nil
}

// =============================================================================
// Factory Functions
// =============================================================================

// Panel creates a panel node.
func Panel(props ui.Props, children ...Node) Node {
	return Node{
		tag:      "panel",
		props:    props,
		children: children,
	}
}

// Text creates a text node.
func Text(content string) Node {
	return Node{
		tag:   "text",
		props: ui.Props{"content": content},
	}
}

// Row creates a horizontal container node.
func Row(props ui.Props, children ...Node) Node {
	if props == nil {
		props = make(ui.Props)
	}
	props["_layout"] = "row"
	return Node{
		tag:      "row",
		props:    props,
		children: children,
	}
}

// Column creates a vertical container node.
func Column(props ui.Props, children ...Node) Node {
	if props == nil {
		props = make(ui.Props)
	}
	props["_layout"] = "column"
	return Node{
		tag:      "column",
		props:    props,
		children: children,
	}
}

// Grid creates a grid container node.
func Grid(props ui.Props, children ...Node) Node {
	if props == nil {
		props = make(ui.Props)
	}
	return Node{
		tag:      "grid",
		props:    props,
		children: children,
	}
}

// =============================================================================
// Grid Dimension Factory Functions (convenience helpers)
// =============================================================================

// FixedDim creates a fixed-size grid dimension.
func FixedDim(size int) grid.Dimension {
	return grid.Fixed(size)
}

// FlexDim creates a flexible grid dimension.
func FlexDim(factor int) grid.Dimension {
	return grid.Flex{Factor: factor}
}

// AutoDim creates an auto-sized grid dimension.
func AutoDim() grid.Dimension {
	return grid.Auto{}
}

// MinDim creates a dimension with minimum size.
func MinDim(min int, content grid.Dimension) grid.Dimension {
	return grid.Min{Min: min, Content: content}
}

// MaxDim creates a dimension with maximum size.
func MaxDim(max int, content grid.Dimension) grid.Dimension {
	return grid.Max{Max: max, Content: content}
}

// =============================================================================
// Property Builder Functions
// =============================================================================

// Props creates a new Props map with fluent builder pattern.
type PropsBuilder struct {
	props ui.Props
}

// NewProps creates a new PropsBuilder.
func NewProps() *PropsBuilder {
	return &PropsBuilder{
		props: make(ui.Props),
	}
}

// Width sets width property.
func (pb *PropsBuilder) Width(w int) *PropsBuilder {
	pb.props["width"] = w
	return pb
}

// Height sets height property.
func (pb *PropsBuilder) Height(h int) *PropsBuilder {
	pb.props["height"] = h
	return pb
}

// Flex sets flex property.
func (pb *PropsBuilder) Flex(f int) *PropsBuilder {
	pb.props["flex"] = f
	return pb
}

// Padding sets padding property.
func (pb *PropsBuilder) Padding(p int) *PropsBuilder {
	pb.props["padding"] = p
	return pb
}

// Title sets title property.
func (pb *PropsBuilder) Title(t string) *PropsBuilder {
	pb.props["title"] = t
	return pb
}

// BorderStyle sets border style.
func (pb *PropsBuilder) BorderStyle(s layout.BorderStyle) *PropsBuilder {
	pb.props["borderStyle"] = s
	return pb
}

// BorderColor sets border color.
func (pb *PropsBuilder) BorderColor(c style.Color) *PropsBuilder {
	pb.props["borderColor"] = c
	return pb
}

// Color sets foreground color.
func (pb *PropsBuilder) Color(c style.Color) *PropsBuilder {
	pb.props["color"] = c
	return pb
}

// Background sets background color.
func (pb *PropsBuilder) Background(c style.Color) *PropsBuilder {
	pb.props["background"] = c
	return pb
}

// Columns sets grid column definitions.
func (pb *PropsBuilder) Columns(cols ...grid.Dimension) *PropsBuilder {
	pb.props["columns"] = cols
	return pb
}

// Rows sets grid row definitions.
func (pb *PropsBuilder) Rows(rows ...grid.Dimension) *PropsBuilder {
	pb.props["rows"] = rows
	return pb
}

// ColumnGap sets the gap between columns.
func (pb *PropsBuilder) ColumnGap(gap int) *PropsBuilder {
	pb.props["columnGap"] = gap
	return pb
}

// RowGap sets the gap between rows.
func (pb *PropsBuilder) RowGap(gap int) *PropsBuilder {
	pb.props["rowGap"] = gap
	return pb
}

// Gap sets both column and row gaps.
func (pb *PropsBuilder) Gap(gap int) *PropsBuilder {
	pb.props["gap"] = gap
	return pb
}

// Set sets a custom property.
func (pb *PropsBuilder) Set(key string, value interface{}) *PropsBuilder {
	pb.props[key] = value
	return pb
}

// Build returns the Props map.
func (pb *PropsBuilder) Build() ui.Props {
	return pb.props
}

// =============================================================================
// Layout Shortcut Functions
// =============================================================================

// FlexWidth creates a width flex property.
func FlexWidth(amount int) ui.Props {
	return ui.Props{"flex": amount}
}

// FlexHeight creates a height flex property.
func FlexHeight(amount int) ui.Props {
	return ui.Props{"flex": amount}
}

// FixedWidth creates a fixed width property.
func FixedWidth(w int) ui.Props {
	return ui.Props{"width": w}
}

// FixedHeight creates a fixed height property.
func FixedHeight(h int) ui.Props {
	return ui.Props{"height": h}
}

// FixedSize creates a fixed size property.
func FixedSize(w, h int) ui.Props {
	return ui.Props{"width": w, "height": h}
}

// AutoWidth creates an auto width property.
func AutoWidth() ui.Props {
	return ui.Props{}
}

// AutoHeight creates an auto height property.
func AutoHeight() ui.Props {
	return ui.Props{}
}

// AutoSize creates an auto size property.
func AutoSize() ui.Props {
	return ui.Props{}
}

// =============================================================================
// Component Shortcut Functions
// =============================================================================

// InfoBox creates an info panel with text.
func InfoBox(title, content string) Node {
	return Panel(
		NewProps().Title(title).Build(),
		Text(content),
	)
}

// ErrorBox creates an error panel with text.
func ErrorBox(title, content string) Node {
	return Panel(
		NewProps().Title(title).BorderStyle(layout.BorderSingle).Build(),
		Text(content),
	)
}

// SuccessBox creates a success panel with text.
func SuccessBox(title, content string) Node {
	return Panel(
		NewProps().Title(title).BorderStyle(layout.BorderDouble).Build(),
		Text(content),
	)
}

// WarningBox creates a warning panel with text.
func WarningBox(title, content string) Node {
	return Panel(
		NewProps().Title(title).BorderStyle(layout.BorderRounded).Build(),
		Text(content),
	)
}

// =============================================================================
// String Representation
// =============================================================================

// String returns a string representation of the Node.
func (n *Node) String() string {
	return n.stringify(0)
}

func (n *Node) stringify(depth int) string {
	indent := ""
	for i := 0; i < depth; i++ {
		indent += "  "
	}

	result := fmt.Sprintf("%s%s", indent, n.tag)

	// Add key properties
	if n.props != nil {
		if title, ok := n.props["title"].(string); ok {
			result += fmt.Sprintf(" title=%q", title)
		}
		if width, ok := n.props["width"].(int); ok && width > 0 {
			result += fmt.Sprintf(" width=%d", width)
		}
		if height, ok := n.props["height"].(int); ok && height > 0 {
			result += fmt.Sprintf(" height=%d", height)
		}
		if flex, ok := n.props["flex"].(int); ok && flex > 0 {
			result += fmt.Sprintf(" flex=%d", flex)
		}
	}

	// Add children
	if len(n.children) > 0 {
		for _, child := range n.children {
			result += "\n" + child.stringify(depth + 1)
		}
	}

	return result
}
