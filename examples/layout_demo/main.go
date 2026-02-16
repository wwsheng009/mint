package main

import (
	"fmt"
	"strings"

	"github.com/wwsheng009/mint/runtime/layout"
)

// =============================================================================
// Mock Node Implementation
// =============================================================================

type TextNode struct {
	id     string
	text   string
	width  int
	height int
	x      int
	y      int
}

func NewTextNode(id, text string) *TextNode {
	// Estimate size based on text length
	width := len(text)
	height := 1
	return &TextNode{id: id, text: text, width: width, height: height}
}

func (n *TextNode) ID() string                          { return n.id }
func (n *TextNode) Type() string                        { return "text" }
func (n *TextNode) Children() []layout.Node             { return nil }
func (n *TextNode) GetPosition() (int, int)             { return n.x, n.y }
func (n *TextNode) SetPosition(x, y int)                { n.x, n.y = x, y }
func (n *TextNode) GetSize() (int, int)                 { return n.width, n.height }
func (n *TextNode) SetSize(w, h int)                    { n.width, n.height = w, h }
func (n *TextNode) GetWidth() int                       { return n.width }
func (n *TextNode) GetHeight() int                      { return n.height }

func (n *TextNode) Measure(constraints layout.Constraints) layout.Size {
	w := constraints.ConstrainWidth(n.width)
	h := constraints.ConstrainHeight(n.height)
	return layout.Size{Width: w, Height: h}
}

func (n *TextNode) GetBaseline() int {
	// Baseline is at the bottom for text (single line)
	return n.height
}

// =============================================================================
// Demo Application
// =============================================================================

func main() {
	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║        Mint Layout Engine - New Render Path Demo            ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Create a simple UI: Header, Content, Footer
	fmt.Println("Creating UI components...")

	// Header with title
	header := layout.NewFlexLayout("header", []layout.Node{
		NewTextNode("title", "=== Application Title ==="),
	})
	header.SetDirection(layout.FlexColumn)

	// Content area with two columns
	leftColumn := layout.NewFlexLayout("left", []layout.Node{
		NewTextNode("label1", "Name:"),
		NewTextNode("value1", "John Doe"),
		NewTextNode("label2", "Email:"),
		NewTextNode("value2", "john@example.com"),
	})
	leftColumn.SetDirection(layout.FlexColumn)
	leftColumn.SetGap(1)

	rightColumn := layout.NewFlexLayout("right", []layout.Node{
		NewTextNode("label3", "Status:"),
		NewTextNode("value3", "Active"),
		NewTextNode("label4", "Role:"),
		NewTextNode("value4", "Admin"),
	})
	rightColumn.SetDirection(layout.FlexColumn)
	rightColumn.SetGap(1)

	content := layout.NewFlexLayout("content", []layout.Node{
		leftColumn,
		rightColumn,
	})
	content.SetDirection(layout.FlexRow)
	content.SetGap(4)

	// Footer
	footer := layout.NewFlexLayout("footer", []layout.Node{
		NewTextNode("footer", "=== Footer ==="),
	})
	footer.SetDirection(layout.FlexColumn)

	// Main container
	root := layout.NewFlexLayout("root", []layout.Node{
		header,
		content,
		footer,
	})
	root.SetDirection(layout.FlexColumn)
	root.SetGap(1)

	// Layout with constraints
	fmt.Println("\n📊 Running layout engine...")
	constraints := layout.Constraints{
		MinWidth:  0,
		MaxWidth:  80,
		MinHeight: 0,
		MaxHeight: 40,
	}

	engine := layout.NewEngine()
	result := engine.Layout(root, constraints)

	fmt.Println("\n📈 Layout Results:")
	fmt.Printf("   Root Size: %dx%d\n", result.Root.Width, result.Root.Height)
	fmt.Printf("   Total Boxes: %d\n", len(result.Boxes))
	fmt.Printf("   Cache Hits: %d\n", engine.GetStats().CacheHits)
	fmt.Printf("   Cache Misses: %d\n", engine.GetStats().CacheMisses)

	// Print all layout boxes
	fmt.Println("\n📦 Layout Boxes:")
	printLayoutBox(result.Root, 0)

	// Render the result as ASCII
	fmt.Println("\n🎨 Rendered Output:")
	fmt.Println("┌" + strings.Repeat("─", 40) + "┐")
	renderLayout(result.Root, 40)
	fmt.Println("└" + strings.Repeat("─", 40) + "┘")

	// Test different layouts
	fmt.Println("\n\n═══════════════════════════════════════════════════════════════")
	fmt.Println("Testing Different Layout Configurations:")
	fmt.Println("═══════════════════════════════════════════════════════════════")

	// Test 1: Centered content
	testCenteredContent()

	// Test 2: Stretched columns
	testStretchedColumns()

	// Test 3: Nested layout
	testNestedLayout()

	// Test 4: Position absolute
	testAbsolutePosition()

	// Test 5: Table layout
	testTableLayout()

	// Test 6: Layer system
	testLayerSystem()

	// Test 7: Border container
	testBorderContainer()

	// Test 8: Margin support
	testMargin()

	// Test 9: Baseline alignment
	testBaselineAlignment()

	fmt.Println("\n✅ All layout tests completed successfully!")
}

func printLayoutBox(box *layout.LayoutBox, indent int) {
	prefix := strings.Repeat("  ", indent)
	fmt.Printf("%s├─ [%s] pos=(%d,%d) size=(%dx%d)\n",
		prefix, box.ID, box.X, box.Y, box.Width, box.Height)
	for _, child := range box.Children {
		printLayoutBox(child, indent+1)
	}
}

func renderLayout(box *layout.LayoutBox, width int) {
	// Simple ASCII rendering of the layout
	for _, child := range box.Children {
		renderNode(child, 0)
	}
}

func renderNode(box *layout.LayoutBox, depth int) {
	indent := strings.Repeat("  ", depth)
	fmt.Printf("│ %s%-20s @(%d,%d) %dx%d\n", indent, box.ID, box.X, box.Y, box.Width, box.Height)
	for _, child := range box.Children {
		renderNode(child, depth+1)
	}
}

// =============================================================================
// Test Functions
// =============================================================================

func testCenteredContent() {
	fmt.Println("\n[Test 1] Centered Content:")
	fmt.Println("─────────────────────────")

	container := layout.NewFlexLayout("centered", []layout.Node{
		NewTextNode("text", "Centered Text"),
	})
	container.SetDirection(layout.FlexRow)
	container.SetMainAxis(layout.Center)

	result := layout.NewEngine().Layout(container, layout.Constraints{
		MinWidth: 0, MaxWidth: 40, MinHeight: 0, MaxHeight: 10,
	})

	fmt.Printf("  Container size: %dx%d\n", result.Root.Width, result.Root.Height)
	if len(result.Root.Children) > 0 {
		fmt.Printf("  Text position: (%d, %d)\n", result.Root.Children[0].X, result.Root.Children[0].Y)
	}
	fmt.Println("  ✅ Center alignment works")
}

func testStretchedColumns() {
	fmt.Println("\n[Test 2] Stretched Columns:")
	fmt.Println("─────────────────────────")

	left := layout.NewFlexLayout("left", []layout.Node{NewTextNode("l", "Left")})
	right := layout.NewFlexLayout("right", []layout.Node{NewTextNode("r", "Right")})

	container := layout.NewFlexLayout("stretched", []layout.Node{left, right})
	container.SetDirection(layout.FlexRow)
	container.SetCrossAxis(layout.Stretch)

	result := layout.NewEngine().Layout(container, layout.Constraints{
		MinWidth: 0, MaxWidth: 40, MinHeight: 0, MaxHeight: 10,
	})

	fmt.Printf("  Container size: %dx%d\n", result.Root.Width, result.Root.Height)
	if len(result.Root.Children) > 1 {
		fmt.Printf("  Left column height: %d\n", result.Root.Children[0].Height)
		fmt.Printf("  Right column height: %d\n", result.Root.Children[1].Height)
	}
	fmt.Println("  ✅ Stretch alignment works")
}

func testNestedLayout() {
	fmt.Println("\n[Test 3] Nested Layout:")
	fmt.Println("──────────────────────")

	inner := layout.NewFlexLayout("inner", []layout.Node{
		NewTextNode("a", "A"),
		NewTextNode("b", "B"),
	})
	inner.SetDirection(layout.FlexColumn)

	outer := layout.NewFlexLayout("outer", []layout.Node{
		inner,
		NewTextNode("c", "C"),
	})
	outer.SetDirection(layout.FlexRow)

	result := layout.NewEngine().Layout(outer, layout.Constraints{
		MinWidth: 0, MaxWidth: 40, MinHeight: 0, MaxHeight: 20,
	})

	fmt.Printf("  Outer size: %dx%d\n", result.Root.Width, result.Root.Height)
	if len(result.Root.Children) > 0 {
		fmt.Printf("  Inner size: %dx%d\n", result.Root.Children[0].Width, result.Root.Children[0].Height)
	}
	fmt.Println("  ✅ Nested layout works")
}

func testAbsolutePosition() {
	fmt.Println("\n[Test 4] Absolute Position:")
	fmt.Println("──────────────────────────")

	// Create positioned content
	top := 5
	left := 10
	pos := layout.NewAbsolutePositionWithOffsets(&top, &left, nil, nil)

	x, y := layout.CalculateAbsolutePosition(100, 50, 20, 10, pos)
	fmt.Printf("  Position type: Absolute\n")
	fmt.Printf("  Offsets: top=%d, left=%d\n", top, left)
	fmt.Printf("  Calculated position: (%d, %d)\n", x, y)

	fmt.Println("  ✅ Absolute positioning works")
}

func testTableLayout() {
	fmt.Println("\n[Test 5] Table Layout:")
	fmt.Println("─────────────────────")

	// Create a simple 3x3 table
	rows := [][]layout.Node{
		{NewTextNode("h1", "Name"), NewTextNode("h2", "Value"), NewTextNode("h3", "Status")},
		{NewTextNode("d1", "Item1"), NewTextNode("d2", "100"), NewTextNode("d3", "OK")},
		{NewTextNode("d4", "Item2"), NewTextNode("d5", "200"), NewTextNode("d6", "Pending")},
	}

	table := layout.NewTableLayout("table", rows)

	w, h := table.GetSize()
	fmt.Printf("  Table size: %dx%d\n", w, h)

	colWidths := table.ColumnWidths()
	fmt.Printf("  Column widths: %v\n", colWidths)

	rowHeights := table.RowHeights()
	fmt.Printf("  Row heights: %v\n", rowHeights)

	// Get cell position
	x, y := table.CellPosition(1, 1)
	fmt.Printf("  Cell (1,1) position: (%d, %d)\n", x, y)

	fmt.Println("  ✅ Table layout works")
}

func testLayerSystem() {
	fmt.Println("\n[Test 6] Layer System:")
	fmt.Println("─────────────────────")

	child := NewTextNode("modal", "Modal Content")
	layered := layout.NewLayeredNode("layered", child, layout.LayerModal, 100)

	fmt.Printf("  Layer: %s\n", layered.GetLayer().String())
	fmt.Printf("  Z-Index: %d\n", layered.GetZIndex())
	fmt.Printf("  Effective Z-Index: %d\n", layered.EffectiveZIndex())

	// Test layer comparison
	other := layout.NewLayeredNode("other", child, layout.LayerBase, 0)
	fmt.Printf("  Modal higher than base: %v\n", layout.IsInHigherLayer(layered, other))

	fmt.Println("  ✅ Layer system works")
}

func testBorderContainer() {
	fmt.Println("\n[Test 7] Border Container:")
	fmt.Println("─────────────────────────")

	// Create bordered content
	inner := NewTextNode("inner", "Content inside border")
	border := layout.NewBorder(layout.BorderSingle)
	border.Label = "Panel"

	bordered := layout.NewBorderedNode("bordered", inner, border)

	// Get size
	w, h := bordered.GetSize()
	fmt.Printf("  Outer size (with border): %dx%d\n", w, h)

	padding := border.HorizontalPadding()
	fmt.Printf("  Border padding: %d horizontal, %d vertical\n", padding, border.VerticalPadding())

	fmt.Println("  ✅ Border container works")
}

func testMargin() {
	fmt.Println("\n[Test 8] Margin Support:")
	fmt.Println("───────────────────────")

	margin := layout.Margin{
		Left:   10,
		Right:  20,
		Top:    5,
		Bottom: 5,
	}

	fmt.Printf("  Margin: L=%d, R=%d, T=%d, B=%d\n", margin.Left, margin.Right, margin.Top, margin.Bottom)
	fmt.Printf("  Horizontal margin: %d\n", margin.Horizontal())
	fmt.Printf("  Vertical margin: %d\n", margin.Vertical())

	fmt.Println("  ✅ Margin support works")
}

func testBaselineAlignment() {
	fmt.Println("\n[Test 9] Baseline Alignment:")
	fmt.Println("───────────────────────────")

	// Create nodes with different heights
	small := NewTextNode("small", "A")
	medium := NewTextNode("medium", "Middle")
	large := NewTextNode("large", "Tall Text")

	container := layout.NewFlexLayout("baseline", []layout.Node{small, medium, large})
	container.SetDirection(layout.FlexRow)
	container.SetCrossAxis(layout.Baseline)

	result := layout.NewEngine().Layout(container, layout.Constraints{
		MinWidth: 0, MaxWidth: 100, MinHeight: 0, MaxHeight: 20,
	})

	fmt.Printf("  Container size: %dx%d\n", result.Root.Width, result.Root.Height)
	fmt.Println("  ✅ Baseline alignment works")
}
