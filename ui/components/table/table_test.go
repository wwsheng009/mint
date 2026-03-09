package table

import (
	"strings"
	"testing"

	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/layout"
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
	"github.com/wwsheng009/mint/runtime/paint"
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

func TestInstance_PaintAppliesSearchFilterAndFooter(t *testing.T) {
	cols := []TableColumn{
		{Title: "ID", Width: 4, Sortable: true},
		{Title: "Name", Width: 10, Sortable: true},
		{Title: "Role", Width: 8},
	}
	rows := [][]string{
		{"2", "Bob", "Admin"},
		{"1", "Alice", "User"},
		{"3", "Alex", "Admin"},
	}

	inst := NewInstance(rtui.Props{
		"columns":        cols,
		"rows":           rows,
		"searchQuery":    "al",
		"filters":        map[int]string{2: "admin"},
		"sortColumn":     0,
		"sortDescending": false,
		"showFooter":     true,
	})

	cmds := inst.Paint(0, 0)
	if got := textAtY(cmds, 0); !strings.Contains(got, "ID ↑") {
		t.Fatalf("header = %q, want sort marker", got)
	}
	if got := textAtY(cmds, 2); !strings.Contains(got, "3") || !strings.Contains(got, "Alex") {
		t.Fatalf("first row = %q, want filtered/sorted Alex row", got)
	}
	if got := textAtY(cmds, 3); !strings.Contains(got, "Rows 1/3") || !strings.Contains(got, "Search \"al\"") || !strings.Contains(got, "Filters 1") {
		t.Fatalf("footer = %q, want search/filter summary", got)
	}
}

func TestInstance_HandleAction_ClickSortableHeaderTogglesSort(t *testing.T) {
	cols := []TableColumn{
		{Title: "ID", Width: 4},
		{Title: "Name", Width: 8, Sortable: true},
	}
	rows := [][]string{
		{"2", "Bob"},
		{"1", "Alice"},
	}

	inst := NewInstance(rtui.Props{
		"columns":    cols,
		"rows":       rows,
		"showBorder": true,
	})

	mouseMsg := runtimemsg.NewMouseMsg(0, 0, runtimemsg.MouseLeft, runtimemsg.MouseActionPress)
	mouseMsg.LocalX = 10
	mouseMsg.LocalY = 1
	act := action.NewAction(action.ActionClick).WithPayload(mouseMsg)

	if !inst.HandleAction(act) {
		t.Fatal("expected header click to be handled")
	}
	if inst.sortColumn != 1 || inst.sortDescending {
		t.Fatalf("sort state = (%d,%v), want column 1 ascending", inst.sortColumn, inst.sortDescending)
	}
	if got := textAtY(inst.Paint(0, 0), 3); !strings.Contains(got, "Alice") {
		t.Fatalf("first sorted row = %q, want Alice", got)
	}

	if !inst.HandleAction(act) {
		t.Fatal("expected second header click to be handled")
	}
	if inst.sortColumn != 1 || !inst.sortDescending {
		t.Fatalf("sort state = (%d,%v), want column 1 descending", inst.sortColumn, inst.sortDescending)
	}
	if got := textAtY(inst.Paint(0, 0), 3); !strings.Contains(got, "Bob") {
		t.Fatalf("first sorted row after toggle = %q, want Bob", got)
	}
}

func TestInstance_HandleAction_PageNavigation(t *testing.T) {
	cols := []TableColumn{
		{Title: "ID", Width: 4},
		{Title: "Name", Width: 8},
	}
	rows := [][]string{
		{"1", "Alice"},
		{"2", "Bob"},
		{"3", "Carol"},
		{"4", "Dave"},
	}

	inst := NewInstance(rtui.Props{
		"columns":  cols,
		"rows":     rows,
		"pageSize": 2,
	})

	if !inst.HandleAction(action.NewAction(action.ActionNavigatePageDown)) {
		t.Fatal("expected page down to be handled")
	}
	if inst.currentPage != 1 {
		t.Fatalf("currentPage = %d, want 1", inst.currentPage)
	}
	if inst.selectedIndex != 2 {
		t.Fatalf("selectedIndex = %d, want 2", inst.selectedIndex)
	}
	if got := textAtY(inst.Paint(0, 0), 2); !strings.Contains(got, "3") || !strings.Contains(got, "Carol") {
		t.Fatalf("first row on page 2 = %q, want Carol", got)
	}
}

func TestBuilder_FluentEnhancements(t *testing.T) {
	vnode := NewBuilder().
		Columns([]TableColumn{{Title: "ID", Sortable: true}}).
		ComponentID("orders.table").
		SearchQuery("alice").
		Filter(0, "1").
		PageSize(5).
		CurrentPage(1).
		SortBy(0, true).
		ShowBorder(true).
		ShowFooter(true).
		ShowScrollbar(true).
		BuildVNode()

	if vnode.searchQuery != "alice" {
		t.Fatalf("searchQuery = %q, want alice", vnode.searchQuery)
	}
	if vnode.pageSize != 5 {
		t.Fatalf("pageSize = %d, want 5", vnode.pageSize)
	}
	if vnode.currentPage != 1 || !vnode.currentPageControlled {
		t.Fatalf("currentPage state = (%d,%v), want (1,true)", vnode.currentPage, vnode.currentPageControlled)
	}
	if vnode.sortColumn != 0 || !vnode.sortDescending || !vnode.sortControlled {
		t.Fatalf("sort state = (%d,%v,%v), want (0,true,true)", vnode.sortColumn, vnode.sortDescending, vnode.sortControlled)
	}
	if !vnode.showBorder || !vnode.showFooter {
		t.Fatal("expected border and footer to be enabled")
	}
	if !vnode.showScrollbar {
		t.Fatal("expected scrollbar to be enabled")
	}
	if vnode.componentID != "orders.table" {
		t.Fatalf("componentID = %q, want orders.table", vnode.componentID)
	}
}

func TestInstance_HandleAction_SelectEmitsFieldChangeIntent(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"columns":      []TableColumn{{Title: "ID", Width: 4}, {Title: "Name", Width: 8}},
		"rows":         [][]string{{"10", "Alice"}, {"20", "Bob"}},
		"changeIntent": intent.BindField("selected_row"),
	})

	var emitted []intent.Intent
	inst.SetIntentEmitter(func(i intent.Intent) { emitted = append(emitted, i) })

	if !inst.HandleAction(action.NewAction(action.ActionNavigateDown)) {
		t.Fatal("expected navigate down to be handled")
	}
	if len(emitted) != 1 {
		t.Fatalf("emitted len = %d, want 1", len(emitted))
	}
	fieldChange, ok := emitted[0].(intent.FieldChangeIntent)
	if !ok {
		t.Fatalf("emitted intent = %T, want FieldChangeIntent", emitted[0])
	}
	if fieldChange.Field != "selected_row" || fieldChange.Value != "0" {
		t.Fatalf("field change = %#v, want selected_row=0", fieldChange)
	}
}

func TestInstance_EmitStateChangeIntent(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"componentID": "orders.table",
		"columns":     []TableColumn{{Title: "ID", Width: 4}, {Title: "Name", Width: 8}},
		"rows":        [][]string{{"10", "Alice"}, {"20", "Bob"}},
	})
	var emitted []intent.Intent
	inst.SetIntentEmitter(func(i intent.Intent) { emitted = append(emitted, i) })

	if !inst.HandleAction(action.NewAction(action.ActionNavigateDown)) {
		t.Fatal("expected navigate down to be handled")
	}
	if len(emitted) != 1 {
		t.Fatalf("emitted len = %d, want 1", len(emitted))
	}
	change, ok := emitted[0].(StateChangeIntent)
	if !ok {
		t.Fatalf("emitted intent = %T, want StateChangeIntent", emitted[0])
	}
	if change.ComponentID != "orders.table" || change.SelectedSourceIndex != 0 || change.PageCount != 1 || change.FilteredRows != 2 || change.TotalRows != 2 {
		t.Fatalf("state change = %#v, want componentID orders.table, source index 0, pageCount 1, filteredRows 2, totalRows 2", change)
	}
}

func TestInstance_PaintHonorsColumnAlignment(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"columns": []TableColumn{
			{Title: "Qty", Width: 5, Align: rtui.AlignEnd},
			{Title: "Name", Width: 8, Align: rtui.AlignCenter},
		},
		"rows": [][]string{{"12", "Bolt"}},
	})

	row := textAtY(inst.Paint(0, 0), 2)
	if !strings.HasPrefix(row, "   12") {
		t.Fatalf("row = %q, want right-aligned quantity", row)
	}
	if !strings.Contains(row, "  Bolt  ") {
		t.Fatalf("row = %q, want centered name cell", row)
	}
}

func TestInstance_PaintShowsScrollbarWhenPaginated(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"columns":       []TableColumn{{Title: "ID", Width: 4}, {Title: "Name", Width: 8}},
		"rows":          [][]string{{"1", "Alice"}, {"2", "Bob"}, {"3", "Carol"}, {"4", "Dave"}},
		"pageSize":      2,
		"showScrollbar": true,
	})

	cmds := inst.Paint(0, 0)
	hasThumb := false
	hasRail := false
	for _, cmd := range cmds {
		if cmd.Text == "█" {
			hasThumb = true
		}
		if cmd.Text == "│" {
			hasRail = true
		}
	}
	if !hasThumb || !hasRail {
		t.Fatalf("scrollbar cmds missing, thumb=%t rail=%t", hasThumb, hasRail)
	}
}

func textAtY(cmds []paint.DrawCmd, y int) string {
	for _, cmd := range cmds {
		if cmd.Y == y {
			return cmd.Text
		}
	}
	return ""
}
