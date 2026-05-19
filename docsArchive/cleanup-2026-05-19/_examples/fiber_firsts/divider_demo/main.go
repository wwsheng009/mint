// Package main demonstrates the Fiber-first Divider component.
package main

import (
	"fmt"
	"os"

	"github.com/wwsheng009/mint/ui/components/divider"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

func main() {
	fmt.Println("=== Fiber-first Divider Demo ===")
	fmt.Println()

	// 1. Create divider components using *VNode methods
	solidDivider := divider.New().Solid()
	dashedDivider := divider.New().Dashed()
	dottedDivider := divider.New().Dotted()
	doubleDivider := divider.New().Double()
	labeledDivider := divider.New().SetLabel(" Section Title ").Solid()

	// 2. Create instances from *VNode (concrete type has CreateInstance method)
	instances := []struct {
		name string
		inst rtui.ComponentInstance
	}{
		{"Solid", solidDivider.CreateInstance()},
		{"Dashed", dashedDivider.CreateInstance()},
		{"Dotted", dottedDivider.CreateInstance()},
		{"Double", doubleDivider.CreateInstance()},
		{"Labeled", labeledDivider.CreateInstance()},
	}

	// 3. Measure with constraints
	constraints := layout.Constraints{
		MinWidth:  0,
		MaxWidth:  50,
		MinHeight: 0,
		MaxHeight: 10,
	}

	for _, item := range instances {
		// Measure
		if measurable, ok := item.inst.(interface{ Measure(layout.Constraints) layout.Size }); ok {
			size := measurable.Measure(constraints)
			fmt.Printf("%s divider size: %dx%d\n", item.name, size.Width, size.Height)
		}
	}

	fmt.Println()
	fmt.Println("=== Rendered Output ===")
	fmt.Println()

	// 4. Render each divider
	for i, item := range instances {
		inst := item.inst
		y := i * 2 // spacing between dividers

		// Set bounds
		if p, ok := inst.(interface{ SetBounds(x, y, w, h int) }); ok {
			p.SetBounds(0, y, 50, 1)
		}

		fmt.Printf("[%s]\n", item.name)

		// Paint
		if paintable, ok := inst.(interface{ Paint(x, y int) []paint.DrawCmd }); ok {
			cmds := paintable.Paint(0, y)
			for _, cmd := range cmds {
				fmt.Printf("  %s\n", cmd.Text)
			}
		}

		fmt.Println()
	}

	// 5. Demo Builder API
	fmt.Println("=== Builder API Demo ===")
	fmt.Println()

	customVNode := divider.NewBuilder().
		Key("custom-1").
		Label("  CUSTOM  ").
		Double().
		Build()

	// Builder.Build() returns rtui.VNode, need type assertion for InstanceFactory
	customInst := customVNode.(*divider.VNode).CreateInstance()
	if measurable, ok := customInst.(interface{ Measure(layout.Constraints) layout.Size }); ok {
		size := measurable.Measure(constraints)
		fmt.Printf("Custom divider size: %dx%d\n", size.Width, size.Height)
	}

	if p, ok := customInst.(interface{ SetBounds(x, y, w, h int) }); ok {
		p.SetBounds(0, 0, 50, 1)
	}

	if paintable, ok := customInst.(interface{ Paint(x, y int) []paint.DrawCmd }); ok {
		cmds := paintable.Paint(0, 0)
		for _, cmd := range cmds {
			fmt.Printf("  %s\n", cmd.Text)
		}
	}

	fmt.Println()

	// 6. Demo convenience functions
	fmt.Println("=== Convenience Functions ===")
	fmt.Println()

	quickDividers := []struct {
		name string
		vnode *divider.VNode
	}{
		{"D()", divider.D().(*divider.VNode)},
		{"H(\"Header\")", divider.H("Header").(*divider.VNode)},
		{"V()", divider.V().(*divider.VNode)},
		{"WithLabel(\"Label\")", divider.WithLabel("Label").(*divider.VNode)},
	}

	for _, item := range quickDividers {
		inst := item.vnode.CreateInstance()
		if measurable, ok := inst.(interface{ Measure(layout.Constraints) layout.Size }); ok {
			size := measurable.Measure(constraints)
			fmt.Printf("%s -> size: %dx%d\n", item.name, size.Width, size.Height)
		}
	}

	fmt.Println()
	fmt.Println("=== All divider styles: Solid, Dashed, Dotted, Double ===")
	fmt.Println("=== With optional centered labels ===")

	os.Exit(0)
}
