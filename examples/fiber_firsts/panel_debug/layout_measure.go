package main

import (
	"fmt"

	"github.com/wwsheng009/mint/internal/render"
	"github.com/wwsheng009/mint/runtime/layout"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/text"
	"github.com/wwsheng009/mint/ui/components/stack"
)

func main() {
	fmt.Println("========================================")
	fmt.Println("Direct Layout.Measure Test")
	fmt.Println("========================================")

	// Test Panel 2 directly (auto-height)
	textNode := text.New("This text is too long and will be wrapped to multiple lines.").SetWrap(true)

	// Wrap in VStack (which is what Panel uses)
	vstackNode := stack.New(stack.Column)
	vstackNode.SetChildrenList([]rtui.VNode{textNode})

	// Wrap in Border (what Panel delegates to)
	borderNode := render.NewVNodeToNodeAdapter(vstackNode)

	fmt.Println("\nTest 1: VNodeToNodeAdapter -> Measure MaxInt constraints")
	constraints1 := layout.Constraints{
		MinWidth:  0,
		MaxWidth:  layout.MaxInt,
		MinHeight: 0,
		MaxHeight: layout.MaxInt,
	}

	engine := layout.NewEngine()
	size1 := engine.Measure(borderNode, constraints1)
	fmt.Printf("Constraints: MaxWidth=%d, MaxHeight=%d\n", constraints1.MaxWidth, constraints1.MaxHeight)
	fmt.Printf("Measured Size: %dx%d\n", size1.Width, size1.Height)

	fmt.Println("\nTest 2: VNode.ToNodeAdapter -> Measure bounded constraints")
	constraints2 := layout.Constraints{
		MinWidth:  0,
		MaxWidth:  50,
		MinHeight: 0,
		MaxHeight: 50,
	}

	size2 := engine.Measure(borderNode, constraints2)
	fmt.Printf("Constraints: MaxWidth=%d, MaxHeight=%d\n", constraints2.MaxWidth, constraints2.MaxHeight)
	fmt.Printf("Measured Size: %dx%d\n", size2.Width, size2.Height)

	fmt.Println("\nTest 3: Direct Stack.Measure")
	stackInst := vstackNode.CreateInstance()
	if measurable, ok := stackInst.(interface{ Measure(layout.Constraints) layout.Size }); ok {
		size3 := measurable.Measure(constraints2)
		fmt.Printf("Constraints: MaxWidth=%d, MaxHeight=%d\n", constraints2.MaxWidth, constraints2.MaxHeight)
		fmt.Printf("Stack.Measured Size: %dx%d\n", size3.Width, size3.Height)
	}

	fmt.Println("\nTest 4: Direct Text.Measure")
	textInst := textNode.CreateInstance()
	if measurable, ok := textInst.(interface{ Measure(layout.Constraints) layout.Size }); ok {
		size4 := measurable.Measure(constraints2)
		fmt.Printf("Constraints: MaxWidth=%d, MaxHeight=%d\n", constraints2.MaxWidth, constraints2.MaxHeight)
		fmt.Printf("Text.Measured Size: %dx%d\n", size4.Width, size4.Height)
	}
}
