package grid

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/layout"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	newtext "github.com/wwsheng009/mint/ui/components/text"
)

// =============================================================================
// VNode Tests
// =============================================================================

func TestVNode_New(t *testing.T) {
	g := New()

	if g == nil {
		t.Fatal("New() returned nil")
	}

	if g.Tag() != "grid" {
		t.Errorf("Expected tag 'grid', got '%s'", g.Tag())
	}

	// Default columns should have one Flex
	if len(g.Columns()) != 1 {
		t.Errorf("Expected 1 default column, got %d", len(g.Columns()))
	}

	// Default rows should have one Auto
	if len(g.Rows()) != 1 {
		t.Errorf("Expected 1 default row, got %d", len(g.Rows()))
	}
}

func TestVNode_SetColumns(t *testing.T) {
	g := New().
		SetColumns(Fixed(10), Flex{Factor: 1}, Auto{})

	cols := g.Columns()
	if len(cols) != 3 {
		t.Fatalf("Expected 3 columns, got %d", len(cols))
	}

	// Check first column is Fixed
	if _, ok := cols[0].(Fixed); !ok {
		t.Error("First column should be Fixed")
	}

	// Check second column is Flex
	if f, ok := cols[1].(Flex); !ok {
		t.Error("Second column should be Flex")
	} else if f.Factor != 1 {
		t.Errorf("Flex factor should be 1, got %d", f.Factor)
	}

	// Check third column is Auto
	if _, ok := cols[2].(Auto); !ok {
		t.Error("Third column should be Auto")
	}
}

func TestVNode_SetRows(t *testing.T) {
	g := New().
		SetRows(Fixed(3), Flex{Factor: 2})

	rows := g.Rows()
	if len(rows) != 2 {
		t.Fatalf("Expected 2 rows, got %d", len(rows))
	}

	// Check first row is Fixed(3)
	if f, ok := rows[0].(Fixed); !ok {
		t.Error("First row should be Fixed")
	} else if int(f) != 3 {
		t.Errorf("Fixed row should be 3, got %d", int(f))
	}

	// Check second row is Flex(2)
	if f, ok := rows[1].(Flex); !ok {
		t.Error("Second row should be Flex")
	} else if f.Factor != 2 {
		t.Errorf("Flex factor should be 2, got %d", f.Factor)
	}
}

func TestVNode_AddCell(t *testing.T) {
	child1 := newtext.New("Cell 1")
	child2 := newtext.New("Cell 2")

	g := New().
		AddCell(0, 0, child1).
		AddCell(1, 2, child2)

	cells := g.Cells()
	if len(cells) != 2 {
		t.Fatalf("Expected 2 cells, got %d", len(cells))
	}

	// Check first cell
	if cells[0].Row != 0 || cells[0].Col != 0 {
		t.Errorf("Cell 0: expected (0,0), got (%d,%d)", cells[0].Row, cells[0].Col)
	}

	// Check second cell
	if cells[1].Row != 1 || cells[1].Col != 2 {
		t.Errorf("Cell 1: expected (1,2), got (%d,%d)", cells[1].Row, cells[1].Col)
	}
}

func TestVNode_SetGap(t *testing.T) {
	g := New().SetGap(2, 1)

	if g.ColumnGap() != 2 {
		t.Errorf("ColumnGap: expected 2, got %d", g.ColumnGap())
	}

	if g.RowGap() != 1 {
		t.Errorf("RowGap: expected 1, got %d", g.RowGap())
	}
}

func TestVNode_SetPadding(t *testing.T) {
	g := New().SetPadding(1, 2, 3, 4)

	p := g.Padding()
	if p[0] != 1 || p[1] != 2 || p[2] != 3 || p[3] != 4 {
		t.Errorf("Padding: expected [1,2,3,4], got %v", p)
	}
}

func TestVNode_SetDimensions(t *testing.T) {
	g := New().
		SetWidth(100).
		SetHeight(30).
		SetFlex(2)

	if g.Width() != 100 {
		t.Errorf("Width: expected 100, got %d", g.Width())
	}

	if g.Height() != 30 {
		t.Errorf("Height: expected 30, got %d", g.Height())
	}

	if g.Flex() != 2 {
		t.Errorf("Flex: expected 2, got %d", g.Flex())
	}
}

func TestVNode_Children(t *testing.T) {
	child1 := newtext.New("A")
	child2 := newtext.New("B")
	child3 := newtext.New("C")

	g := New().
		SetColumns(Flex{Factor: 1}, Flex{Factor: 1}).
		SetChildrenAuto([]rtui.VNode{child1, child2, child3})

	children := g.Children()
	if len(children) != 3 {
		t.Fatalf("Expected 3 children, got %d", len(children))
	}

	cells := g.Cells()
	// Check auto-positioning (row-major order with 2 columns)
	// Cell 0: row=0, col=0
	// Cell 1: row=0, col=1
	// Cell 2: row=1, col=0
	if cells[0].Row != 0 || cells[0].Col != 0 {
		t.Errorf("Cell 0: expected (0,0), got (%d,%d)", cells[0].Row, cells[0].Col)
	}
	if cells[1].Row != 0 || cells[1].Col != 1 {
		t.Errorf("Cell 1: expected (0,1), got (%d,%d)", cells[1].Row, cells[1].Col)
	}
	if cells[2].Row != 1 || cells[2].Col != 0 {
		t.Errorf("Cell 2: expected (1,0), got (%d,%d)", cells[2].Row, cells[2].Col)
	}
}

func TestVNode_InstanceFactory(t *testing.T) {
	g := New().SetWidth(50)

	inst := g.CreateInstance()
	if inst == nil {
		t.Fatal("CreateInstance returned nil")
	}

	// Verify it's a GridInstance
	gridInst, ok := inst.(*Instance)
	if !ok {
		t.Fatal("Instance is not a *Instance")
	}

	if gridInst.width != 50 {
		t.Errorf("Instance width: expected 50, got %d", gridInst.width)
	}
}

// =============================================================================
// Instance Tests
// =============================================================================

func TestInstance_Measure_Empty(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"columns": []Dimension{Flex{Factor: 1}},
		"rows":    []Dimension{Auto{}},
	})

	constraints := layout.NewConstraints(0, 100, 0, 50)
	size := inst.Measure(constraints)

	// Empty grid with Flex column and Auto row
	// Should have minimum size
	if size.Width < 0 {
		t.Errorf("Width should be non-negative, got %d", size.Width)
	}
	if size.Height < 0 {
		t.Errorf("Height should be non-negative, got %d", size.Height)
	}
}

func TestInstance_Measure_FixedDimensions(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"columns": []Dimension{Fixed(10), Fixed(20)},
		"rows":    []Dimension{Fixed(2), Fixed(3)},
	})

	constraints := layout.NewConstraints(0, 100, 0, 50)
	size := inst.Measure(constraints)

	// Fixed: 10 + 20 = 30 width
	// Fixed: 2 + 3 = 5 height
	if size.Width != 30 {
		t.Errorf("Width: expected 30, got %d", size.Width)
	}
	if size.Height != 5 {
		t.Errorf("Height: expected 5, got %d", size.Height)
	}
}

func TestInstance_Measure_WithGaps(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"columns":    []Dimension{Fixed(10), Fixed(20)},
		"rows":       []Dimension{Fixed(2), Fixed(3)},
		"columnGap":  2,
		"rowGap":     1,
	})

	constraints := layout.NewConstraints(0, 100, 0, 50)
	size := inst.Measure(constraints)

	// Fixed: 10 + 20 + 1 gap = 32 width
	// Fixed: 2 + 3 + 1 gap = 6 height
	if size.Width != 32 {
		t.Errorf("Width: expected 32, got %d", size.Width)
	}
	if size.Height != 6 {
		t.Errorf("Height: expected 6, got %d", size.Height)
	}
}

func TestInstance_Measure_WithPadding(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"columns": []Dimension{Fixed(10)},
		"rows":    []Dimension{Fixed(2)},
		"padding": [4]int{1, 2, 3, 4}, // top, right, bottom, left
	})

	constraints := layout.NewConstraints(0, 100, 0, 50)
	size := inst.Measure(constraints)

	// Fixed: 10 + 2 + 4 padding = 16 width
	// Fixed: 2 + 1 + 3 padding = 6 height
	if size.Width != 16 {
		t.Errorf("Width: expected 16, got %d", size.Width)
	}
	if size.Height != 6 {
		t.Errorf("Height: expected 6, got %d", size.Height)
	}
}

func TestInstance_Measure_FlexColumns(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"columns": []Dimension{Flex{Factor: 1}, Flex{Factor: 2}},
		"rows":    []Dimension{Fixed(1)},
	})

	constraints := layout.NewConstraints(0, 60, 0, 20)
	size := inst.Measure(constraints)

	// Total width = 60, distribute by flex factor (1:2)
	// Col 0: 60 * 1/3 = 20
	// Col 1: 60 * 2/3 = 40
	if size.Width != 60 {
		t.Errorf("Width: expected 60, got %d", size.Width)
	}

	// Check column widths are calculated
	widths := inst.GetColumnWidths()
	if len(widths) != 2 {
		t.Fatalf("Expected 2 column widths, got %d", len(widths))
	}

	if widths[0] != 20 {
		t.Errorf("Col 0 width: expected 20, got %d", widths[0])
	}
	if widths[1] != 40 {
		t.Errorf("Col 1 width: expected 40, got %d", widths[1])
	}
}

func TestInstance_Measure_ExplicitSize(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"columns": []Dimension{Fixed(10)},
		"rows":    []Dimension{Fixed(1)},
		"width":   50,
		"height":  20,
	})

	constraints := layout.NewConstraints(0, 100, 0, 50)
	size := inst.Measure(constraints)

	// Explicit size should override calculated size
	if size.Width != 50 {
		t.Errorf("Width: expected 50, got %d", size.Width)
	}
	if size.Height != 20 {
		t.Errorf("Height: expected 20, got %d", size.Height)
	}
}

func TestInstance_Measure_Constraints(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"columns": []Dimension{Fixed(100)},
		"rows":    []Dimension{Fixed(50)},
	})

	// Tight constraints
	constraints := layout.TightConstraints(30, 10)
	size := inst.Measure(constraints)

	// Should be constrained to 30x10
	if size.Width != 30 {
		t.Errorf("Width: expected 30, got %d", size.Width)
	}
	if size.Height != 10 {
		t.Errorf("Height: expected 10, got %d", size.Height)
	}
}

func TestInstance_Bounds(t *testing.T) {
	inst := NewInstance(rtui.Props{})

	inst.SetBounds(5, 10, 100, 50)
	x, y, w, h := inst.GetBounds()

	if x != 5 || y != 10 || w != 100 || h != 50 {
		t.Errorf("Bounds: expected (5,10,100,50), got (%d,%d,%d,%d)", x, y, w, h)
	}
}

func TestInstance_SetProps(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"columns": []Dimension{Fixed(10)},
	})

	// Update props
	changed := inst.SetProps(rtui.Props{
		"columns": []Dimension{Fixed(20), Fixed(30)},
		"width":   50,
	})

	if !changed {
		t.Error("SetProps should return true when columns change")
	}

	if len(inst.columns) != 2 {
		t.Errorf("Expected 2 columns, got %d", len(inst.columns))
	}

	if inst.width != 50 {
		t.Errorf("Width: expected 50, got %d", inst.width)
	}
}

// =============================================================================
// Builder Tests
// =============================================================================

func TestBuilder_Full(t *testing.T) {
	child := newtext.New("Test")

	g := NewBuilder().
		Key("test-grid").
		Columns(Fixed(10), Flex{Factor: 1}).
		Rows(Fixed(2), Auto{}).
		Cell(0, 0, child).
		Gap(1, 0).
		Padding(1, 1, 1, 1).
		Width(50).
		Height(20).
		Build()

	// Type assert to *VNode
	grid, ok := g.(*VNode)
	if !ok {
		t.Fatal("Build() should return *VNode")
	}

	if grid.Key() != "test-grid" {
		t.Errorf("Key: expected 'test-grid', got '%s'", grid.Key())
	}

	if len(grid.Columns()) != 2 {
		t.Errorf("Expected 2 columns, got %d", len(grid.Columns()))
	}

	if len(grid.Rows()) != 2 {
		t.Errorf("Expected 2 rows, got %d", len(grid.Rows()))
	}

	if len(grid.Cells()) != 1 {
		t.Errorf("Expected 1 cell, got %d", len(grid.Cells()))
	}

	if grid.Width() != 50 || grid.Height() != 20 {
		t.Errorf("Dimensions: expected (50,20), got (%d,%d)", grid.Width(), grid.Height())
	}
}

func TestBuilder_Children(t *testing.T) {
	children := []rtui.VNode{
		newtext.New("A"),
		newtext.New("B"),
		newtext.New("C"),
		newtext.New("D"),
	}

	g := NewBuilder().
		Columns(Flex{Factor: 1}, Flex{Factor: 1}).
		Children(children).
		Build()

	grid, ok := g.(*VNode)
	if !ok {
		t.Fatal("Build() should return *VNode")
	}

	cells := grid.Cells()
	if len(cells) != 4 {
		t.Fatalf("Expected 4 cells, got %d", len(cells))
	}

	// Check auto-positioning with 2 columns
	// A (0,0), B (0,1), C (1,0), D (1,1)
	expected := []struct{ row, col int }{
		{0, 0}, {0, 1}, {1, 0}, {1, 1},
	}

	for i, exp := range expected {
		if cells[i].Row != exp.row || cells[i].Col != exp.col {
			t.Errorf("Cell %d: expected (%d,%d), got (%d,%d)",
				i, exp.row, exp.col, cells[i].Row, cells[i].Col)
		}
	}
}

// =============================================================================
// Convenience Function Tests
// =============================================================================

func TestSimpleGrid(t *testing.T) {
	children := []rtui.VNode{
		newtext.New("A"),
		newtext.New("B"),
		newtext.New("C"),
	}

	g := SimpleGrid(3, children...)

	grid, ok := g.(*VNode)
	if !ok {
		t.Fatal("SimpleGrid should return *VNode")
	}

	cols := grid.Columns()
	if len(cols) != 3 {
		t.Errorf("Expected 3 columns, got %d", len(cols))
	}

	// All columns should be Flex{Factor: 1}
	for i, col := range cols {
		if f, ok := col.(Flex); !ok {
			t.Errorf("Column %d should be Flex", i)
		} else if f.Factor != 1 {
			t.Errorf("Column %d: expected factor 1, got %d", i, f.Factor)
		}
	}
}

func TestTwoColumnGrid(t *testing.T) {
	children := []rtui.VNode{
		newtext.New("A"),
		newtext.New("B"),
	}

	g := TwoColumnGrid(children...)

	grid, ok := g.(*VNode)
	if !ok {
		t.Fatal("TwoColumnGrid should return *VNode")
	}

	cols := grid.Columns()
	if len(cols) != 2 {
		t.Errorf("Expected 2 columns, got %d", len(cols))
	}
}

func TestThreeColumnGrid(t *testing.T) {
	children := []rtui.VNode{
		newtext.New("A"),
		newtext.New("B"),
		newtext.New("C"),
	}

	g := ThreeColumnGrid(children...)

	grid, ok := g.(*VNode)
	if !ok {
		t.Fatal("ThreeColumnGrid should return *VNode")
	}

	cols := grid.Columns()
	if len(cols) != 3 {
		t.Errorf("Expected 3 columns, got %d", len(cols))
	}
}

func TestFixedGrid(t *testing.T) {
	children := []rtui.VNode{
		newtext.New("A"),
		newtext.New("B"),
	}

	g := FixedGrid([]int{10, 20}, children...)

	grid, ok := g.(*VNode)
	if !ok {
		t.Fatal("FixedGrid should return *VNode")
	}

	cols := grid.Columns()
	if len(cols) != 2 {
		t.Errorf("Expected 2 columns, got %d", len(cols))
	}

	// First column should be Fixed(10)
	if f, ok := cols[0].(Fixed); !ok || int(f) != 10 {
		t.Errorf("Column 0 should be Fixed(10)")
	}

	// Second column should be Fixed(20)
	if f, ok := cols[1].(Fixed); !ok || int(f) != 20 {
		t.Errorf("Column 1 should be Fixed(20)")
	}
}

// =============================================================================
// Dimension Type Tests
// =============================================================================

func TestDimensionTypes(t *testing.T) {
	// Fixed
	f := Fixed(10)
	if int(f) != 10 {
		t.Errorf("Fixed: expected 10, got %d", int(f))
	}

	// Flex
	flex := Flex{Factor: 2}
	if flex.Factor != 2 {
		t.Errorf("Flex factor: expected 2, got %d", flex.Factor)
	}

	// Auto
	auto := Auto{}
	_ = auto // just verify it compiles

	// Min
	min := Min{Min: 5, Content: Fixed(10)}
	if min.Min != 5 {
		t.Errorf("Min: expected 5, got %d", min.Min)
	}

	// Max
	max := Max{Max: 20, Content: Flex{Factor: 1}}
	if max.Max != 20 {
		t.Errorf("Max: expected 20, got %d", max.Max)
	}
}
