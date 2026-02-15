package main

import (
	"fmt"
	"strings"

	"github.com/wwsheng009/mint/components/basic"
	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/compute"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

func main() {
	vnode := rtui.VStack(
		basic.NewText("Title"),
		rtui.HStack(
			basic.NewText("A"),
			basic.NewText("B"),
			basic.NewText("C"),
		),
		basic.NewText("X"),
		basic.NewText("Y"),
		basic.NewText("Z"),
	)

	engine := compute.NewEngine()
	fiber := rtui.CreateFiberFromVNode(vnode)

	fmt.Println("=== Fiber-First Layout Demo ===")
	fmt.Println(strings.Repeat("-", 60))

	fmt.Println("Fiber Tree:")
	printFiberTree(fiber, 0)

	layout, err := engine.BuildComputedBoxFiberOnly(fiber, runtime.BoxConstraints{MaxWidth: 60, MaxHeight: 20})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println()
	fmt.Println("ComputedBox Tree:")
	printComputedBoxTree(layout.Root, 0)

	fmt.Println()
	fmt.Println("Layout Statistics:")
	fmt.Println(strings.Repeat("-", 60))
	printStatistics(layout.Root)
}

func printFiberTree(fiber *rtui.Fiber, depth int) {
	if fiber == nil {
		return
	}
	indent := strings.Repeat("  ", depth)
	fmt.Printf("%s[NODE %d] Tag=%s Children=%d\n", indent, fiber.NodeID, fiber.Tag, len(fiber.GetChildFibers()))
	for _, child := range fiber.GetChildFibers() {
		printFiberTree(child, depth+1)
	}
}

func printComputedBoxTree(box *compute.ComputedBox, depth int) {
	if box == nil {
		return
	}
	indent := strings.Repeat("  ", depth)
	tag := "unknown"
	if f := box.GetFiber(); f != nil {
		tag = f.Tag
	}
	fmt.Printf("%s[NODE %d] Tag=%s Pos=%dx%d Size=%dx%d Children=%d\n",
		indent, box.NodeID, tag, box.Box.X, box.Box.Y, box.Box.Width, box.Box.Height, len(box.Children))
	for _, child := range box.Children {
		printComputedBoxTree(child, depth+1)
	}
}

func printStatistics(box *compute.ComputedBox) {
	fmt.Printf("  Total Nodes: %d\n", countNodes(box))
	fmt.Printf("  Root Size:   %dx%d\n", box.Box.Width, box.Box.Height)
	fmt.Printf("  Max Depth:   %d\n", maxDepth(box, 0))
}

func countNodes(box *compute.ComputedBox) int {
	if box == nil {
		return 0
	}
	count := 1
	for _, child := range box.Children {
		count += countNodes(child)
	}
	return count
}

func maxDepth(box *compute.ComputedBox, current int) int {
	if box == nil {
		return current
	}
	max := current
	for _, child := range box.Children {
		d := maxDepth(child, current+1)
		if d > max {
			max = d
		}
	}
	return max
}
