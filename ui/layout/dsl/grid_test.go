package dsl_test

import (
	"testing"

	"github.com/wwsheng009/mint/ui/components/grid"
	"github.com/wwsheng009/mint/ui/layout/dsl"
)

// TestGridDSL tests the Grid DSL support
func TestGridDSL(t *testing.T) {
	tests := []struct {
		name   string
		node   dsl.Node
		wantTag string
	}{
		{
			name: "Simple Grid",
			node: dsl.Grid(
				dsl.NewProps().Width(80).Height(20).Gap(1).Build(),
				dsl.Text("Cell 1"),
				dsl.Text("Cell 2"),
			),
			wantTag: "grid",
		},
		{
			name: "Grid with custom dimensions",
			node: dsl.Grid(
				dsl.NewProps().
					Columns(dsl.FixedDim(20), dsl.FlexDim(1)).
					Rows(dsl.AutoDim(), dsl.AutoDim()).
					Gap(1).
					Build(),
				dsl.Text("Fixed"),
				dsl.Text("Flexible"),
			),
			wantTag: "grid",
		},
		{
			name: "Nested Grid in Panel",
			node: dsl.Panel(
				dsl.NewProps().Title("Dashboard").Width(80).Height(30).Build(),
				dsl.Grid(
					dsl.NewProps().Columns(dsl.FlexDim(1), dsl.FlexDim(1)).Gap(2).Build(),
					dsl.Text("Item 1"),
					dsl.Text("Item 2"),
					dsl.Text("Item 3"),
					dsl.Text("Item 4"),
				),
			),
			wantTag: "panel",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vnode := tt.node.ToVNode()
			if vnode == nil {
				t.Fatal("ToVNode() returned nil")
			}

			if vnode.Tag() != tt.wantTag {
				t.Errorf("Tag() = %v, want %v", vnode.Tag(), tt.wantTag)
			}
		})
	}
}

// TestGridDSLConversion tests Grid DSL converts to actual Grid VNode
func TestGridDSLConversion(t *testing.T) {
	// Create a Grid DSL node
	gridNode := dsl.Grid(
		dsl.NewProps().
			Columns(dsl.FixedDim(20), dsl.FlexDim(1)).
			Rows(dsl.AutoDim(), dsl.AutoDim()).
			Gap(1).
			Width(80).
			Height(20).
			Build(),
		dsl.Text("Cell 1"),
		dsl.Text("Cell 2"),
		dsl.Text("Cell 3"),
		dsl.Text("Cell 4"),
	)

	// Convert to VNode
	vnode := gridNode.ToVNode()

	// Check if it's a Grid VNode
	gridVNode, ok := vnode.(*grid.VNode)
	if !ok {
		t.Fatal("Grid DSL node did not convert to grid.VNode")
	}

	// Check dimensions
	if len(gridVNode.Columns()) != 2 {
		t.Errorf("Expected 2 columns, got %d", len(gridVNode.Columns()))
	}
	if len(gridVNode.Rows()) != 2 {
		t.Errorf("Expected 2 rows, got %d", len(gridVNode.Rows()))
	}

	// Check sizes
	if gridVNode.Width() != 80 {
		t.Errorf("Expected width 80, got %d", gridVNode.Width())
	}
	if gridVNode.Height() != 20 {
		t.Errorf("Expected height 20, got %d", gridVNode.Height())
	}

	// Check gap
	if gridVNode.ColumnGap() != 1 || gridVNode.RowGap() != 1 {
		t.Errorf("Expected columnGap=1 and rowGap=1, got columnGap=%d, rowGap=%d",
			gridVNode.ColumnGap(), gridVNode.RowGap())
	}

	// Check children
	children := gridVNode.Children()
	if len(children) != 4 {
		t.Errorf("Expected 4 children, got %d", len(children))
	}
}

// TestGridDimensionHelpers tests the dimension helper functions
func TestGridDimensionHelpers(t *testing.T) {
	t.Run("FixedDim", func(t *testing.T) {
		dim := dsl.FixedDim(100)
		if _, ok := dim.(grid.Fixed); !ok {
			t.Errorf("FixedDim() did not return grid.Fixed")
		}
	})

	t.Run("FlexDim", func(t *testing.T) {
		dim := dsl.FlexDim(2)
		if f, ok := dim.(grid.Flex); ok {
			if f.Factor != 2 {
				t.Errorf("FlexDim(2) got Factor = %d, want 2", f.Factor)
			}
		} else {
			t.Errorf("FlexDim() did not return grid.Flex")
		}
	})

	t.Run("AutoDim", func(t *testing.T) {
		dim := dsl.AutoDim()
		if _, ok := dim.(grid.Auto); !ok {
			t.Errorf("AutoDim() did not return grid.Auto")
		}
	})

	t.Run("MinDim", func(t *testing.T) {
		dim := dsl.MinDim(100, dsl.AutoDim())
		if m, ok := dim.(grid.Min); ok {
			if m.Min != 100 {
				t.Errorf("MinDim(100, ...) got Min = %d, want 100", m.Min)
			}
		} else {
			t.Errorf("MinDim() did not return grid.Min")
		}
	})

	t.Run("MaxDim", func(t *testing.T) {
		dim := dsl.MaxDim(200, dsl.AutoDim())
		if m, ok := dim.(grid.Max); ok {
			if m.Max != 200 {
				t.Errorf("MaxDim(200, ...) got Max = %d, want 200", m.Max)
			}
		} else {
			t.Errorf("MaxDim() did not return grid.Max")
		}
	})
}

// TestGridDSLPropsBuilder tests the PropsBuilder for Grid props
func TestGridDSLPropsBuilder(t *testing.T) {
	props := dsl.NewProps().
		Columns(dsl.FixedDim(20), dsl.FlexDim(1)).
		Rows(dsl.AutoDim()).
		ColumnGap(2).
		RowGap(1).
		Gap(1).
		Build()

	// Check columns
	if cols, ok := props["columns"].([]grid.Dimension); ok {
		if len(cols) != 2 {
			t.Errorf("Expected 2 columns, got %d", len(cols))
		}
	} else {
		t.Error("props[\"columns\"] is not []grid.Dimension")
	}

	// Check gap
	if gap, ok := props["gap"].(int); ok {
		if gap != 1 {
			t.Errorf("Expected gap 1, got %d", gap)
		}
	} else {
		t.Error("props[\"gap\"] is not int")
	}
}

// TestGridComplexScenario tests a complex Grid layout scenario
func TestGridComplexScenario(t *testing.T) {
	layout := dsl.Panel(
		dsl.NewProps().Title("Complex Layout").Width(80).Height(40).Build(),
		dsl.Column(
			dsl.NewProps().Gap(1).Build(),
			dsl.Text("Header"),
			dsl.Row(
				dsl.NewProps().Columns(dsl.FlexDim(2), dsl.FlexDim(1)).Gap(1).Build(),
				dsl.Grid(
					dsl.NewProps().
						Columns(dsl.FixedDim(20), dsl.FlexDim(1)).
						Gap(1).
						Build(),
					dsl.Text("A"),
					dsl.Text("B"),
					dsl.Text("C"),
					dsl.Text("D"),
				),
				dsl.Column(
					dsl.NewProps().Build(),
					dsl.Text("Side 1"),
					dsl.Text("Side 2"),
				),
			),
			dsl.Text("Footer"),
		),
	)

	vnode := layout.ToVNode()
	if vnode == nil {
		t.Fatal("ToVNode() returned nil")
	}

	// Should be a panel
	if vnode.Tag() != "panel" {
		t.Errorf("Expected tag 'panel', got '%s'", vnode.Tag())
	}
}
