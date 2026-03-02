package inspector

import (
	"fmt"
	"testing"

	"github.com/wwsheng009/mint/ui"
	"github.com/wwsheng009/mint/ui/components/stack"
)

// TestBorderedNodeStructure demonstrates how Stack border works
func TestBorderedNodeStructure(t *testing.T) {
	fmt.Println("\n=== Stack Border Structure Demo ===")

	// Create a bordered Stack with content
	bordered := stack.NewVStack().
		DoubleBorder("Title").
		SetChildrenList([]ui.VNode{
			ui.VStack(
				ui.Text("A"),
				ui.Text("B"),
				ui.Text("C"),
			),
		})

	fmt.Printf("Bordered Stack type: %T\n", bordered)
	fmt.Printf("Bordered Stack children: %d\n", len(bordered.Children()))

	// Get the child (should be a VStack)
	children := bordered.Children()
	if len(children) > 0 {
		child := children[0]
		fmt.Printf("Child type: %T\n", child)
		fmt.Printf("Child children: %d\n", len(child.Children()))

		// The VStack should have 3 Text children
		grandchildren := child.Children()
		fmt.Printf("Grandchildren (Text nodes): %d\n", len(grandchildren))

		for i, gc := range grandchildren {
			fmt.Printf("  Grandchild %d: %T\n", i, gc)
		}
	}

	// Build inspector tree
	tv := NewTreeView()
	tv.SetShowPaths(true)
	err := tv.SetRoot(bordered)
	if err != nil {
		t.Fatalf("SetRoot failed: %v", err)
	}

	fmt.Println("\n=== Inspector Tree ===")
	lines, _ := tv.GetTreeLines()
	for i, line := range lines {
		fmt.Printf("%2d: %s\n", i, line)
	}

	// Get all nodes
	allNodes := tv.GetFlatList()
	fmt.Printf("\n=== Node Details ===\n")
	fmt.Printf("Total nodes: %d\n\n", len(allNodes))

	for i, node := range allNodes {
		fmt.Printf("Node %d:\n", i)
		fmt.Printf("  Type: %s\n", node.Info.Type)
		fmt.Printf("  Path: %s\n", node.Path)
		fmt.Printf("  Children (in tree): %d\n", len(node.Children))

		// For BorderedNode, show if it has border properties
		if bn, ok := node.VNode.(*ui.BorderedNode); ok {
			fmt.Printf("  >>> Has Border Properties <<<\n")
			fmt.Printf("  Border Style: (check via reflection)\n")
			fmt.Printf("  Border Label: (check via reflection)\n")
			_ = bn // Use variable
		}

		fmt.Println()
	}

	fmt.Println("=== Key Points ===")
	fmt.Println("1. BorderedNode has 1 child (the content)")
	fmt.Println("2. Border is NOT a separate VNode - it's a RENDERING property")
	fmt.Println("3. During Render(), BorderedNode draws border characters around content")
	fmt.Println("4. This is like Text content - not a node, just rendering logic")
}

// TestBorderRenderingExplanation explains how borders are rendered
func TestBorderRenderingExplanation(t *testing.T) {
	fmt.Println("\n=== How Border Rendering Works ===")

	fmt.Println("VNode Tree Structure:")
	fmt.Println("  BorderedNode")
	fmt.Println("    └─ VStack (child)")
	fmt.Println("       ├─ Text")
	fmt.Println("       ├─ Text")
	fmt.Println("       └─ Text")
	fmt.Println()

	fmt.Println("Rendering Process:")
	fmt.Println("  1. BorderedNode.Render() is called")
	fmt.Println("  2. BorderedNode draws border characters: ┌───┐ │ │ └───┘")
	fmt.Println("  3. BorderedNode calls child.Render()")
	fmt.Println("  4. Child renders INSIDE the border area")
	fmt.Println()

	fmt.Println("Why border is NOT a VNode:")
	fmt.Println("  ❌ NOT like: BorderedNode → BorderVNode → Content")
	fmt.Println("  ✅ LIKE:    BorderedNode (draws border, then renders child)")
	fmt.Println()

	fmt.Println("Analogy:")
	fmt.Println("  Text content is not a separate node")
	fmt.Println("  Border decoration is not a separate node")
	fmt.Println("  Both are RENDERING ATTRIBUTES of their parent node")
}

// TestBorderedVsNormalNode compares BorderedNode with normal VStack
func TestBorderedVsNormalNode(t *testing.T) {
	fmt.Println("\n=== Comparing BorderedNode vs Normal VStack ===")

	// Normal VStack
	vstack := ui.VStack(
		ui.Text("A"),
		ui.Text("B"),
	)

	// Bordered VStack (migrated to Stack)
	bordered := stack.NewVStack().
		SingleBorder().
		SetChildrenList([]ui.VNode{
			ui.VStack(
				ui.Text("A"),
				ui.Text("B"),
			),
		})

	fmt.Println("Normal VStack:")
	fmt.Printf("  Type: %T\n", vstack)
	fmt.Printf("  Children: %d\n", len(vstack.Children()))
	fmt.Printf("  Tree would show: VStack → Text, Text\n")
	fmt.Println()

	fmt.Println("Bordered Stack:")
	fmt.Printf("  Type: %T\n", bordered)
	fmt.Printf("  Children: %d (the VStack inside border)\n", len(bordered.Children()))
	fmt.Printf("  Tree shows: Stack (bordered) → VStack → Text, Text\n")
	fmt.Println()

	fmt.Println("Visual Output:")
	fmt.Println("  Normal:  A  B")
	fmt.Println("  Bordered:")
	fmt.Println("           ┌───┐")
	fmt.Println("           │ A │")
	fmt.Println("           │ B │")
	fmt.Println("           └───┘")
	fmt.Println()
}
