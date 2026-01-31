package ui

import "testing"

// =============================================================================
// Select Tests
// =============================================================================

// TestNewSelect tests select creation
func TestNewSelect(t *testing.T) {
	selectNode := NewSelect()

	if selectNode == nil {
		t.Fatal("NewSelect() returned nil")
	}

	if selectNode.Selected() != -1 {
		t.Errorf("Initial selected = %v, want -1", selectNode.Selected())
	}

	if selectNode.Disabled() != false {
		t.Errorf("Initial disabled = %v, want false", selectNode.Disabled())
	}
}

// TestSelectBuilder tests select builder
func TestSelectBuilder(t *testing.T) {
	selectNode := SelectBuilder().
		AddOption("1", "Option 1").
		AddOption("2", "Option 2").
		Selected(0).
		Disabled(true).
		Build()

	selectVNode, ok := selectNode.(*SelectVNode)
	if !ok {
		t.Fatal("SelectBuilder() did not return *SelectVNode")
	}

	if selectVNode.Selected() != 0 {
		t.Errorf("Selected = %v, want 0", selectVNode.Selected())
	}

	if !selectVNode.Disabled() {
		t.Error("Disabled should be true")
	}

	if len(selectVNode.Options()) != 2 {
		t.Errorf("Options length = %v, want 2", len(selectVNode.Options()))
	}
}

// TestSelectOptions tests setting options
func TestSelectOptions(t *testing.T) {
	selectNode := NewSelect()

	options := []SelectOption{
		{Value: "a", Label: "A"},
		{Value: "b", Label: "B"},
		{Value: "c", Label: "C"},
	}
	selectNode.SetOptions(options)

	if len(selectNode.Options()) != 3 {
		t.Errorf("Options length = %v, want 3", len(selectNode.Options()))
	}

	if selectNode.Options()[0].Label != "A" {
		t.Errorf("First option label = %v, want 'A'", selectNode.Options()[0].Label)
	}
}

// TestSelectAddOption tests adding options
func TestSelectAddOption(t *testing.T) {
	selectNode := NewSelect()

	selectNode.AddOption("1", "One")
	selectNode.AddOption("2", "Two")

	if len(selectNode.Options()) != 2 {
		t.Errorf("Options length = %v, want 2", len(selectNode.Options()))
	}
}

// TestSelectSetSelected tests setting selected index
func TestSelectSetSelected(t *testing.T) {
	selectNode := NewSelect()
	selectNode.AddOption("1", "One")
	selectNode.AddOption("2", "Two")

	selectNode.SetSelected(1)
	if selectNode.Selected() != 1 {
		t.Errorf("Selected = %v, want 1", selectNode.Selected())
	}

	if selectNode.SelectedValue() != "2" {
		t.Errorf("SelectedValue = %v, want '2'", selectNode.SelectedValue())
	}

	if selectNode.SelectedLabel() != "Two" {
		t.Errorf("SelectedLabel = %v, want 'Two'", selectNode.SelectedLabel())
	}
}

// TestSelectByValue tests selecting by value
func TestSelectByValue(t *testing.T) {
	selectNode := NewSelect()
	selectNode.AddOption("1", "One")
	selectNode.AddOption("2", "Two")
	selectNode.AddOption("3", "Three")

	selectNode.SelectByValue("2")

	if selectNode.Selected() != 1 {
		t.Errorf("Selected = %v, want 1", selectNode.Selected())
	}

	if selectNode.SelectedValue() != "2" {
		t.Errorf("SelectedValue = %v, want '2'", selectNode.SelectedValue())
	}
}

// TestSelectOnChange tests change handler
func TestSelectOnChange(t *testing.T) {
	called := false
	var receivedValue string

	selectNode := SelectBuilder().
		AddOption("1", "One").
		AddOption("2", "Two").
		OnChange(func(val string) {
			called = true
			receivedValue = val
		}).
		Build()

	selectVNode, ok := selectNode.(*SelectVNode)
	if !ok {
		t.Fatal("Build() did not return *SelectVNode")
	}

	if selectVNode.OnChange() == nil {
		t.Error("OnChange should not be nil")
	} else {
		selectVNode.OnChange()("test")
		if !called {
			t.Error("OnChange handler was not called")
		}
		if receivedValue != "test" {
			t.Errorf("OnChange received %v, want 'test'", receivedValue)
		}
	}
}

// =============================================================================
// Table Tests
// =============================================================================

// TestNewTable tests table creation
func TestNewTable(t *testing.T) {
	table := NewTable()

	if table == nil {
		t.Fatal("NewTable() returned nil")
	}

	if len(table.Columns()) != 0 {
		t.Errorf("Initial columns length = %v, want 0", len(table.Columns()))
	}

	if len(table.Rows()) != 0 {
		t.Errorf("Initial rows length = %v, want 0", len(table.Rows()))
	}
}

// TestTableBuilder tests table builder
func TestTableBuilder(t *testing.T) {
	table := TableBuilder().
		Columns([]TableColumn{
			{Title: "Name", Width: 20},
			{Title: "Age", Width: 5},
		}).
		AddRow("Alice", "30").
		AddRow("Bob", "25").
		Build()

	tableVNode, ok := table.(*TableVNode)
	if !ok {
		t.Fatal("TableBuilder() did not return *TableVNode")
	}

	if len(tableVNode.Columns()) != 2 {
		t.Errorf("Columns length = %v, want 2", len(tableVNode.Columns()))
	}

	if len(tableVNode.Rows()) != 2 {
		t.Errorf("Rows length = %v, want 2", len(tableVNode.Rows()))
	}
}

// TestTableSetColumns tests setting columns
func TestTableSetColumns(t *testing.T) {
	table := NewTable()

	columns := []TableColumn{
		{Title: "ID", Width: 5},
		{Title: "Name", Width: 20},
	}
	table.SetColumns(columns)

	if len(table.Columns()) != 2 {
		t.Errorf("Columns length = %v, want 2", len(table.Columns()))
	}

	if table.Columns()[0].Title != "ID" {
		t.Errorf("First column title = %v, want 'ID'", table.Columns()[0].Title)
	}
}

// TestTableAddRow tests adding rows
func TestTableAddRow(t *testing.T) {
	table := NewTable()

	table.AddRow([]string{"1", "Alice"})
	table.AddRow([]string{"2", "Bob"})

	if len(table.Rows()) != 2 {
		t.Errorf("Rows length = %v, want 2", len(table.Rows()))
	}

	if table.Rows()[0][0] != "1" {
		t.Errorf("First row first cell = %v, want '1'", table.Rows()[0][0])
	}
}

// TestTableSetRows tests setting rows
func TestTableSetRows(t *testing.T) {
	table := NewTable()

	rows := [][]string{
		{"1", "Alice"},
		{"2", "Bob"},
	}
	table.SetRows(rows)

	if len(table.Rows()) != 2 {
		t.Errorf("Rows length = %v, want 2", len(table.Rows()))
	}
}

// TestTableHeaderStyle tests header style
func TestTableHeaderStyle(t *testing.T) {
	table := NewTable()

	style := table.HeaderStyle()
	if !style.IsBold() {
		t.Error("Header style should be bold by default")
	}

	newStyle := style.Foreground("cyan")
	table.SetHeaderStyle(newStyle)

	if table.HeaderStyle().FG != "cyan" {
		t.Errorf("Header style FG = %v, want 'cyan'", table.HeaderStyle().FG)
	}
}

// BenchmarkSelectBuilder benchmarks select builder
func BenchmarkSelectBuilder(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SelectBuilder().
			AddOption("1", "One").
			AddOption("2", "Two").
			AddOption("3", "Three").
			Build()
	}
}

// BenchmarkTableBuilder benchmarks table builder
func BenchmarkTableBuilder(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		TableBuilder().
			Columns([]TableColumn{{Title: "Name"}, {Title: "Age"}}).
			AddRow("Alice", "30").
			AddRow("Bob", "25").
			Build()
	}
}
