package main

import (
	"fmt"

	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/ui/components/grid"
	"github.com/wwsheng009/mint/ui/layout/dsl"
)

// Grid DSL Demo
//
// This demo shows how to use the DSL to create Grid layouts.
// The DSL provides a declarative API for building layouts.

func main() {
	fmt.Println("=== Grid DSL Demo ===")
	fmt.Println()

	// Example 1: Simple Grid
	fmt.Println("Example 1: Simple Grid (2x2)")
	fmt.Println("---")
	simpleGrid := dsl.Grid(
		dsl.NewProps().
			Columns(dsl.FixedDim(20), dsl.FlexDim(1)).
			Gap(1).
			Build(),
		dsl.Text("Fixed"),
		dsl.Text("Flexible"),
		dsl.Text("Another"),
		dsl.Text("Cell"),
	)
	fmt.Println(simpleGrid.String())
	fmt.Println()

	// Example 2: Grid with Auto Rows
	fmt.Println("Example 2: Grid with Auto Rows")
	fmt.Println("---")
	autoGrid := dsl.Grid(
		dsl.NewProps().
			Columns(dsl.FlexDim(1), dsl.FlexDim(1), dsl.FlexDim(1)).
			Rows(dsl.AutoDim(), dsl.AutoDim(), dsl.AutoDim()).
			Gap(2).
			Build(),
		dsl.Text("1"), dsl.Text("2"), dsl.Text("3"),
		dsl.Text("4"), dsl.Text("5"), dsl.Text("6"),
		dsl.Text("7"), dsl.Text("8"), dsl.Text("9"),
	)
	fmt.Println(autoGrid.String())
	fmt.Println()

	// Example 3: Mixed Dimensions
	fmt.Println("Example 3: Mixed Dimensions (Fixed + Flex + Min/Max)")
	fmt.Println("---")
	mixedGrid := dsl.Grid(
		dsl.NewProps().
			Columns(
				dsl.FixedDim(20),
				dsl.FlexDim(1),
				dsl.MinDim(15, dsl.FlexDim(1)),
				dsl.MaxDim(25, dsl.FlexDim(1)),
			).
			Rows(dsl.AutoDim()).
			Gap(1).
			Build(),
		dsl.Text("Fixed=20"),
		dsl.Text("Flex"),
		dsl.Text("Min=15"),
		dsl.Text("Max=25"),
	)
	fmt.Println(mixedGrid.String())
	fmt.Println()

	// Example 4: Dashboard Layout
	fmt.Println("Example 4: Dashboard Layout (Panel + Grid + Column)")
	fmt.Println("---")
	dashboard := dsl.Panel(
		dsl.NewProps().
			Title("Dashboard").
			Width(80).
			Height(30).
			BorderStyle(layout.BorderDouble).
			Build(),
		dsl.Column(
			dsl.NewProps().Gap(1).Build(),
			dsl.Text("Stats Overview"),
			dsl.Grid(
				dsl.NewProps().
					Columns(dsl.FlexDim(1), dsl.FlexDim(1), dsl.FlexDim(1), dsl.FlexDim(1)).
					Gap(1).
					Build(),
				dsl.Text("Users: 100"),
				dsl.Text("Sales: $5K"),
				dsl.Text("Orders: 50"),
				dsl.Text("Rating: 4.5"),
			),
		),
	)
	fmt.Println(dashboard.String())
	fmt.Println()

	// Example 5: Asymmetric Grid
	fmt.Println("Example 5: Asymmetric Grid (3 columns, 2 rows)")
	fmt.Println("---")
	asymmetricGrid := dsl.Grid(
		dsl.NewProps().
			Columns(dsl.FixedDim(15), dsl.FlexDim(1), dsl.FixedDim(10)).
			Gap(1).
			Width(70).
			Build(),
		dsl.Text("Label 1"),
		dsl.Text("Content 1 fills remaining space"),
		dsl.Text("5%"),
		dsl.Text("Label 2"),
		dsl.Text("Content 2"),
		dsl.Text("10%"),
	)
	fmt.Println(asymmetricGrid.String())
	fmt.Println()

	// Example 6: Nested Containers
	fmt.Println("Example 6: Complex Layout (Column contains Row which contains Grid)")
	fmt.Println("---")
	complexLayout := dsl.Panel(
		dsl.NewProps().Title("Complex Layout").Build(),
		dsl.Column(
			dsl.NewProps().Gap(1).Build(),
			dsl.Text("Header Section"),
			dsl.Row(
				dsl.NewProps().Gap(2).Build(),
				dsl.Grid(
					dsl.NewProps().Columns(dsl.FlexDim(1), dsl.FlexDim(1)).Gap(2).Build(),
					dsl.Text("Menu Item 1"),
					dsl.Text("Menu Item 2"),
					dsl.Text("Menu Item 3"),
					dsl.Text("Menu Item 4"),
				),
				dsl.Grid(
					dsl.NewProps().Columns(dsl.FlexDim(1)).Gap(1).Build(),
					dsl.Text("Sidebar Item 1"),
					dsl.Text("Sidebar Item 2"),
				),
			),
			dsl.Text("Footer"),
		),
	)
	fmt.Println(complexLayout.String())
	fmt.Println()

	// ========================================
	// Example 7: Grid VNode Details
	// ========================================
	fmt.Println("Example 7: Grid VNode Details")
	fmt.Println("---")

	measureGrid := dsl.Grid(
		dsl.NewProps().
			Columns(dsl.FixedDim(20), dsl.FlexDim(1)).
			Rows(dsl.AutoDim(), dsl.AutoDim()).
			Gap(1).
			Width(60).
			Build(),
		dsl.Text("Cell 1"),
		dsl.Text("Cell 2"),
		dsl.Text("Cell 3"),
		dsl.Text("Cell 4"),
	)

	vnode := measureGrid.ToVNode()
	if gridVNode, ok := vnode.(*grid.VNode); ok {
		fmt.Printf("  Columns: %d\n", len(gridVNode.Columns()))
		fmt.Printf("  Rows: %d\n", len(gridVNode.Rows()))
		fmt.Printf("  Cells: %d\n", len(gridVNode.Cells()))
		fmt.Printf("  Column Gap: %d\n", gridVNode.ColumnGap())
		fmt.Printf("  Row Gap: %d\n", gridVNode.RowGap())
		fmt.Printf("  Width: %d\n", gridVNode.Width())
		fmt.Printf("  Height: %d\n", gridVNode.Height())
	}
	fmt.Println()

	// Example 8: VNode Props Inspection
	fmt.Println("Example 8: VNode Props Inspection")
	fmt.Println("---")
	propGrid := dsl.Grid(
		dsl.NewProps().
			Columns(dsl.FixedDim(15), dsl.FlexDim(1)).
			Gap(2).
			Build(),
		dsl.Text("A"),
		dsl.Text("B"),
	)

	vnode = propGrid.ToVNode()
	props := vnode.Props()

	fmt.Println("VNode Tag: " + vnode.Tag())
	if cols, ok := props["columns"].([]grid.Dimension); ok {
		fmt.Printf("  Props[columns]: %d dimensions\n", len(cols))
	}
	if gap, ok := props["gap"].(int); ok {
		fmt.Printf("  Props[gap]: %d\n", gap)
	}
	fmt.Println()

	fmt.Println("=== Demo Complete ===")
}

