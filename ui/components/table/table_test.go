package table

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// VNode Tests
// =============================================================================

func TestVNode_New(t *testing.T) {
	vnode := New()
	if vnode == nil {
		t.Fatal("New() returned nil")
	}
	if vnode.Tag() != "table" {
		t.Errorf("Expected tag 'table', got '%s'", vnode.Tag())
	}
}

func TestVNode_ImplementsInterfaces(t *testing.T) {
	vnode := New()

	// Test VNode interface
	var _ rtui.VNode = vnode

	// Test InstanceFactory interface
	var _ rtui.InstanceFactory = vnode
}

func TestVNode_FluentAPI(t *testing.T) {
	cols := []TableColumn{
		{Title: "ID", Width: 10},
		{Title: "Name", Width: 20},
		{Title: "Email", Width: 30},
	}
	rows := [][]string{
		{"1", "John", "john@example.com"},
		{"2", "Jane", "jane@example.com"},
	}

	vnode := New().
		SetColumns(cols).
		SetRows(rows).
		SetGap(1).
		SetHeaderStyle(style.Style{}.Bold(true).Foreground("cyan"))

	if len(vnode.columns) != 3 {
		t.Errorf("Expected 3 columns, got %d", len(vnode.columns))
	}
	if len(vnode.rows) != 2 {
		t.Errorf("Expected 2 rows, got %d", len(vnode.rows))
	}
	if vnode.gap != 1 {
		t.Errorf("Expected gap 1, got %d", vnode.gap)
	}
}

func TestVNode_AddRow(t *testing.T) {
	vnode := New().
		AddRow("1", "Name1").
		AddRow("2", "Name2").
		AddRow("3", "Name3")

	if len(vnode.rows) != 3 {
		t.Errorf("Expected 3 rows, got %d", len(vnode.rows))
	}
	if vnode.rows[0][0] != "1" {
		t.Errorf("Expected '1', got '%s'", vnode.rows[0][0])
	}
	if vnode.rows[2][1] != "Name3" {
		t.Errorf("Expected 'Name3', got '%s'", vnode.rows[2][1])
	}
}

func TestVNode_Props(t *testing.T) {
	cols := []TableColumn{{Title: "Col1"}}
	rows := [][]string{{"A", "B"}}

	vnode := New().
		SetColumns(cols).
		SetRows(rows).
		SetGap(2)

	props := vnode.Props()
	if !columnsEqual(props["columns"].([]TableColumn), cols) {
		t.Error("Columns mismatch")
	}
	if props["gap"] != 2 {
		t.Error("Gap mismatch")
	}
}

// =============================================================================
// Instance Tests
// =============================================================================

func TestInstance_NewInstance(t *testing.T) {
	cols := []TableColumn{
		{Title: "ID", Width: 10},
		{Title: "Name", Width: 20},
	}
	rows := [][]string{
		{"1", "John"},
		{"2", "Jane"},
	}

	inst := NewInstance(rtui.Props{
		"columns": cols,
		"rows":    rows,
		"gap":     1,
	})

	if inst == nil {
		t.Fatal("NewInstance() returned nil")
	}
	if len(inst.columns) != 2 {
		t.Errorf("Expected 2 columns, got %d", len(inst.columns))
	}
	if len(inst.rows) != 2 {
		t.Errorf("Expected 2 rows, got %d", len(inst.rows))
	}
	if inst.gap != 1 {
		t.Errorf("Expected gap 1, got %d", inst.gap)
	}
}

func TestInstance_Measure(t *testing.T) {
	cols := []TableColumn{
		{Title: "ID"},
		{Title: "Name"},
	}
	rows := [][]string{
		{"1", "John"},
		{"2", "Jane"},
	}

	inst := NewInstance(rtui.Props{
		"columns": cols,
		"rows":    rows,
	})

	size := inst.Measure(layout.Constraints{})
	
	// Expected: 2 chars (ID) + 3 spaces (" | ") + 4 chars (Name) = 9 width
	// Height: header(1) + separator(1) + 2 rows = 4
	if size.Width == 0 {
		t.Error("Expected non-zero width")
	}
	if size.Height != 4 {
		t.Errorf("Expected height 4, got %d", size.Height)
	}
}

func TestInstance_MeasureWithGap(t *testing.T) {
	cols := []TableColumn{{Title: "Col"}}
	rows := [][]string{{"Data"}}

	inst := NewInstance(rtui.Props{
		"columns": cols,
		"rows":    rows,
		"gap":     2,
	})

	size := inst.Measure(layout.Constraints{})
	if size.Height != 5 { // 1 header + 1 separator + 2 gap + 1 row
		t.Errorf("Expected height 5 with gap=2, got %d", size.Height)
	}
}

func TestInstance_Paint(t *testing.T) {
	cols := []TableColumn{
		{Title: "ID"},
		{Title: "Name"},
	}
	rows := [][]string{
		{"1", "John"},
		{"2", "Jane"},
	}

	inst := NewInstance(rtui.Props{
		"columns":     cols,
		"rows":        rows,
		"headerStyle": style.Style{}.Bold(true).Foreground("cyan"),
		"tableStyle":  style.Style{}.Foreground("white"),
	})

	cmds := inst.Paint(0, 0)
	if cmds == nil {
		t.Error("Paint() should return commands")
	}
	if len(cmds) == 0 {
		t.Error("Paint() should return non-empty commands")
	}
	// Should have: header + separator + 2 rows = 4 commands
	if len(cmds) != 4 {
		t.Errorf("Expected 4 commands, got %d", len(cmds))
	}
}

func TestInstance_PaintEmpty(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"columns": []TableColumn{},
		"rows":    [][]string{},
	})

	cmds := inst.Paint(0, 0)
	if cmds != nil {
		t.Error("Paint() with no columns should return nil")
	}
}

// =============================================================================
// Builder Tests
// =============================================================================

func TestBuilder_FluentAPI(t *testing.T) {
	cols := []TableColumn{
		{Title: "ID", Width: 10},
		{Title: "Name", Width: 20},
	}

	vnode := NewBuilder().
		Columns(cols).
		AddRow("1", "John").
		AddRow("2", "Jane").
		Gap(1).
		BuildVNode()

	if len(vnode.columns) != 2 {
		t.Errorf("Expected 2 columns, got %d", len(vnode.columns))
	}
	if len(vnode.rows) != 2 {
		t.Errorf("Expected 2 rows, got %d", len(vnode.rows))
	}
	if vnode.gap != 1 {
		t.Errorf("Expected gap 1, got %d", vnode.gap)
	}
}

func TestBuilder_AddRowMultiple(t *testing.T) {
	vnode := NewBuilder().
		AddRow("a", "b").
		AddRow("c", "d").
		AddRow("e", "f").
		BuildVNode()

	if len(vnode.rows) != 3 {
		t.Errorf("Expected 3 rows, got %d", len(vnode.rows))
	}
}

func TestBuilder_ConvenienceFunctions(t *testing.T) {
	cols := []TableColumn{{Title: "Col1"}}
	rows := [][]string{{"Data"}}

	// Test Of
	vnode := Of(cols, rows)
	if vnode.(*VNode).columns == nil {
		t.Error("Of() should set columns and rows")
	}

	// Test OfColumns
	vnode = OfColumns(cols)
	if vnode.(*VNode).columns == nil {
		t.Error("OfColumns() should set columns")
	}
}

// =============================================================================
// Helper Tests
// =============================================================================

func TestColumnsEqual(t *testing.T) {
	cols1 := []TableColumn{
		{Title: "ID", Width: 10},
		{Title: "Name", Width: 20},
	}
	cols2 := []TableColumn{
		{Title: "ID", Width: 10},
		{Title: "Name", Width: 20},
	}
	cols3 := []TableColumn{
		{Title: "ID", Width: 10},
		{Title: "Name", Width: 25},
	}

	if !columnsEqual(cols1, cols2) {
		t.Error("Expected equal columns")
	}
	if columnsEqual(cols1, cols3) {
		t.Error("Expected unequal columns")
	}
}

func TestColumnsEqual_DifferentLength(t *testing.T) {
	cols1 := []TableColumn{{Title: "A"}}
	cols2 := []TableColumn{{Title: "A"}, {Title: "B"}}

	if columnsEqual(cols1, cols2) {
		t.Error("Expected unequal columns with different length")
	}
}
