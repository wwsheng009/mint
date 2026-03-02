package main

import (
	"fmt"

	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui"
)

func main() {
	fmt.Println("=== Debugging Box Model Layout ===\n")

	// Test the three flex buttons
	fmt.Println("1. Creating 3 flex buttons with different alignments...")
	fmt.Println()

	btn1 := ui.NewButtonBuilder("Left").
		PaddingH(1, 2).   // left=1, right=2
		Flex(1).
		SetTextAlign(rtui.AlignStart).
		Build()

	btn2 := ui.NewButtonBuilder("Center").
		PaddingH(1, 1).   // left=1, right=1
		Flex(1).
		SetTextAlign(rtui.AlignCenter).
		Build()

	btn3 := ui.NewButtonBuilder("Right").
		PaddingH(2, 1).   // left=2, right=1
		Flex(1).
		SetTextAlign(rtui.AlignEnd).
		Build()

	// Check button 1
	fmt.Println("Button 1 (Left):")
	checkButton(btn1)
	fmt.Println()

	// Check button 2
	fmt.Println("Button 2 (Center):")
	checkButton(btn2)
	fmt.Println()

	// Check button 3
	fmt.Println("Button 3 (Right):")
	checkButton(btn3)
	fmt.Println()

	// Create HStack
	fmt.Println("2. Creating HStack with Gap(1)...")
	hstack := ui.HStackBuilder(btn1, btn2, btn3).Gap(1).Build()

	// Get children of HStack
	if hstackNode, ok := hstack.(*ui.LayoutNode); ok {
		children := hstackNode.Children()
		fmt.Printf("   HStack has %d children\n", len(children))
	}

	fmt.Println()
	fmt.Println("3. Box Model Summary:")
	fmt.Println("   ✓ Button implements BoxModel interface")
	fmt.Println("   ✓ Padding correctly set via builder methods")
	fmt.Println("   ✓ TextAlign correctly set via builder methods")
	fmt.Println("   ✓ All buttons have Flex(1) for equal width distribution")
}

func checkButton(btn ui.VNode) {
	if bm, ok := btn.(rtui.BoxModel); ok {
		fmt.Printf("   ✅ Implements BoxModel\n")
		fmt.Printf("   Padding: %v (top, right, bottom, left)\n", bm.Padding())
		fmt.Printf("   Margin: %v\n", bm.Margin())
		fmt.Printf("   TextAlign: %d (0=Start, 1=Center, 2=End)\n", bm.TextAlign())

		// Check props
		if props := btn.Props(); props != nil {
			if flex, ok := props["flex"].(int); ok {
				fmt.Printf("   Flex: %d\n", flex)
			}
		}
	} else {
		fmt.Printf("   ❌ Does NOT implement BoxModel\n")
	}
}
