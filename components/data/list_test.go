package data

import (
	"strings"
	"testing"

	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/style"
)

func TestListBuilder_Basic(t *testing.T) {
	list := ListBuilder().
		Header("Name").
		Rows([]string{"Alice", "Bob", "Charlie"}).
		Build()

	if list == nil {
		t.Fatal("List should not be nil")
	}

	listVNode := list.(*ListVNode)
	if listVNode == nil {
		t.Fatal("ListVNode should not be nil")
	}

	t.Logf("Header: '%s'", listVNode.Header())
	t.Logf("Rows: %v", listVNode.Rows())

	if listVNode.Header() != "Name" {
		t.Errorf("Expected header 'Name', got '%s'", listVNode.Header())
	}

	rows := listVNode.Rows()
	if len(rows) != 3 {
		t.Errorf("Expected 3 rows, got %d", len(rows))
	}
	if rows[0] != "Alice" || rows[1] != "Bob" || rows[2] != "Charlie" {
		t.Errorf("Unexpected row data")
	}
}

func TestListBuilder_HeaderAndSeparator(t *testing.T) {
	list := ListBuilder().
		Header("ID\tName").
		Rows([]string{"1\tAlice", "2\tBob"}).
		ShowSeparator(true).
		Build()

	listVNode := list.(*ListVNode)

	// Test measurement with unbounded constraints (natural width)
	constraints := runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth:  runtime.Infinity,
		MinHeight: 0,
		MaxHeight: runtime.Infinity,
	}
	size := listVNode.Measure(constraints)

	
	// Height should be: header (1) + separator (1) + 2 rows = 4
	if size.Height != 4 {
		t.Errorf("Expected height 4, got %d", size.Height)
	}

	// Width should be natural width (max of all lines)
	expectedWidth := 7 // "ID\tName" is 7 characters (tab counts as 1)
	if size.Width != expectedWidth {
		t.Errorf("Expected width %d (natural), got %d", expectedWidth, size.Width)
	}

	// Test painting
	cmds := listVNode.Paint(0, 0)
	if len(cmds) != 4 {
		t.Errorf("Expected 4 draw commands (header + sep + 2 rows), got %d", len(cmds))
	}

	// First command should be header
	if cmds[0].Text != "ID\tName" {
		t.Errorf("Expected header text 'ID\\tName', got '%s'", cmds[0].Text)
	}

	// Second command should be separator (width matches measured width)
	expectedSep := strings.Repeat(string(listVNode.sepChar), size.Width)
	if cmds[1].Text != expectedSep {
		t.Errorf("Expected separator line of width %d, got '%s'", size.Width, cmds[1].Text)
	}
}

func TestListBuilder_EmptyList(t *testing.T) {
	list := ListBuilder().
		Header("Empty").
		Rows([]string{}).
		EmptyText("No data available").
		Build()

	listVNode := list.(*ListVNode)
	size := listVNode.Measure(runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth:  runtime.Infinity,
		MinHeight: 0,
		MaxHeight: runtime.Infinity,
	})
	if size.Height != 2 { // header + empty text
		t.Errorf("Expected height 2, got %d", size.Height)
	}

	cmds := listVNode.Paint(0, 0)
	if len(cmds) != 2 {
		t.Errorf("Expected 2 draw commands (header + empty text), got %d", len(cmds))
	}

	if cmds[1].Text != "No data available" {
		t.Errorf("Expected empty text 'No data available', got '%s'", cmds[1].Text)
	}
}

func TestListBuilder_MaxRows(t *testing.T) {
	rows := []string{"1", "2", "3", "4", "5"}
	list := ListBuilder().
		Rows(rows).
		MaxRows(3).
		Build()

	listVNode := list.(*ListVNode)
	size := listVNode.Measure(runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth:  runtime.Infinity,
		MinHeight: 0,
		MaxHeight: runtime.Infinity,
	})
	if size.Height != 4 { // 3 rows + "... (more rows)"
		t.Errorf("Expected height 4, got %d", size.Height)
	}

	cmds := listVNode.Paint(0, 0)
	if len(cmds) != 4 {
		t.Errorf("Expected 4 draw commands (3 rows + more indicator), got %d", len(cmds))
	}

	// Last command should be overflow indicator (truncated if needed)
	expectedText := "... (more rows)"
	if size.Width < len(expectedText) {
		expectedText = expectedText[:size.Width-1] + "…"
	} else {
		expectedText += strings.Repeat(" ", size.Width-len(expectedText))
	}
	if cmds[3].Text != expectedText {
		t.Errorf("Expected overflow indicator, got '%s' (width=%d)", cmds[3].Text, size.Width)
	}
}

func TestListBuilder_Styles(t *testing.T) {
	headerStyle := style.Style{}.Bold(true).Foreground("red")
	rowStyle := style.Style{}.Italic(true)

	list := ListBuilder().
		Header("Header").
		Rows([]string{"Row1", "Row2"}).
		HeaderStyle(headerStyle).
		RowStyle(rowStyle).
		Build()

	listVNode := list.(*ListVNode)
	cmds := listVNode.Paint(0, 0)

	// Header should use headerStyle
	if cmds[0].Style != headerStyle {
		t.Error("Header should use headerStyle")
	}

	// Rows should use rowStyle
	if cmds[2].Style != rowStyle {
		t.Error("Rows should use rowStyle")
	}
}

func TestListBuilder_PerRowStyle(t *testing.T) {
	list := ListBuilder().
		Header("Data").
		Rows([]string{"First", "Second", "Third"}).
		RowStyleFn(func(index int, text string) style.Style {
			if index%2 == 0 {
				return style.Style{}.Bold(true)
			}
			return style.Style{}.Italic(true)
		}).
		Build()

	listVNode := list.(*ListVNode)
	cmds := listVNode.Paint(0, 0)

	// Even index rows should be bold
	if cmds[2].Style.IsBold() {
		t.Log("Even index row is bold - good")
	} else {
		t.Error("Even index row should be bold")
	}

	// Odd index rows should be italic
	if cmds[3].Style.IsItalic() {
		t.Log("Odd index row is italic - good")
	} else {
		t.Error("Odd index row should be italic")
	}
}

func TestListBuilder_Truncation(t *testing.T) {
	list := ListBuilder().
		Header("LongHeader").
		Rows([]string{"VeryLongRowThatExceedsTheWidth"}).
		Build()

	listVNode := list.(*ListVNode)
	// Set a tight constraint
	constraints := runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth:  10,
		MinHeight: 0,
		MaxHeight: runtime.Infinity,
	}
	size := listVNode.Measure(constraints)

	if size.Width != 10 {
		t.Errorf("Expected width 10, got %d", size.Width)
	}

	cmds := listVNode.Paint(0, 0)

	// Header should not be truncated (9 chars < 10 max)
	if cmds[0].Text != "LongHeader" {
		t.Errorf("Expected header 'LongHeader', got '%s'", cmds[0].Text)
	}

	// Row should be truncated (10 max, so 9 chars + ellipsis)
	if cmds[2].Text != "VeryLongR…" {
		t.Errorf("Expected truncated row 'VeryLongR…', got '%s'", cmds[2].Text)
	}
}

func TestListBuilder_NoSeparator(t *testing.T) {
	list := ListBuilder().
		Header("Header").
		Rows([]string{"Row1"}).
		ShowSeparator(false).
		Build()

	listVNode := list.(*ListVNode)
	cmds := listVNode.Paint(0, 0)
	if len(cmds) != 2 { // Only header + row, no separator
		t.Errorf("Expected 2 draw commands, got %d", len(cmds))
	}

	// No separator should be present
	if cmds[1].Text == "────────────────────────────────────" {
		t.Error("Separator should not be present when ShowSeparator(false)")
	}
}

func TestListBuilder_SeparatorCharacter(t *testing.T) {
	list := ListBuilder().
		Header("Header").
		Rows([]string{"Row1"}).
		SepChar('=').
		Build()

	listVNode := list.(*ListVNode)
	// Measure first to get the width
	size := listVNode.Measure(runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth:  runtime.Infinity,
		MinHeight: 0,
		MaxHeight: runtime.Infinity,
	})
	cmds := listVNode.Paint(0, 0)
	expectedSep := strings.Repeat("=", size.Width)
	if cmds[1].Text != expectedSep {
		t.Errorf("Expected separator with '=', got '%s'", cmds[1].Text)
	}
}

func TestListBuilder_WidthProperty(t *testing.T) {
	list := ListBuilder().
		Rows([]string{"Row1", "Row2"}).
		Width(50).
		Build()

	listVNode := list.(*ListVNode)
	// Explicit width should override natural width
	constraints := runtime.BoxConstraints{
		MinWidth:  10,
		MaxWidth:  100,
		MinHeight: 5,
		MaxHeight: runtime.Infinity,
	}
	size := listVNode.Measure(constraints)

	if size.Width != 50 {
		t.Errorf("Expected width 50 from explicit prop, got %d", size.Width)
	}
}

func TestListBuilder_Chaining(t *testing.T) {
	list := ListBuilder().
		Header("Title").
		AddRow("Row1").
		AddRow("Row2").
		Rows([]string{"Row3", "Row4"}).
		AddRow("Row5").
		Build()

	listVNode := list.(*ListVNode)
	rows := listVNode.Rows()
	if len(rows) != 3 {
		t.Errorf("Expected 3 rows (Rows() sets rows, AddRow appends), got %d", len(rows))
	}
	if rows[0] != "Row3" || rows[1] != "Row4" || rows[2] != "Row5" {
		t.Errorf("Unexpected row data after chaining: %v", rows)
	}
}

func TestListVNode_Getters(t *testing.T) {
	list := ListBuilder().
		Header("Test").
		Rows([]string{"A", "B"}).
		Build()

	listVNode := list.(*ListVNode)

	// Test getters
	if listVNode.Header() != "Test" {
		t.Errorf("GetHeader failed")
	}

	if listVNode.RowCount() != 2 {
		t.Errorf("RowCount failed")
	}

	// Test setters
	listVNode.SetHeader("NewHeader")
	listVNode.SetRows([]string{"X", "Y", "Z"})

	if listVNode.Header() != "NewHeader" {
		t.Errorf("SetHeader failed")
	}

	if len(listVNode.Rows()) != 3 {
		t.Errorf("SetRows failed")
	}
}

func TestTruncateRunes(t *testing.T) {
	tests := []struct {
		input    string
		maxRunes int
		expected string
	}{
		{"Hello", 5, "Hello"},
		{"Hello", 3, "He…"},
		{"Hello", 0, ""},
		{"Hello", 1, "…"},
		{"", 10, "          "}, // Empty string gets padded
		{"VeryLongString", 5, "Very…"},
		{"日本語", 2, "日…"}, // Unicode test with truncation
	}

	for _, test := range tests {
		result := truncateRunes(test.input, test.maxRunes)
		if result != test.expected {
			t.Errorf("truncateRunes(%q, %d) = %q, want %q", test.input, test.maxRunes, result, test.expected)
		}
	}
}

func TestListBuilder_MeasureConstraints(t *testing.T) {
	list := ListBuilder().
		Header("Short").
		Rows([]string{"A bit longer row"}).
		Build()

	listVNode := list.(*ListVNode)

	// Test minimum width constraint
	minConstraints := runtime.BoxConstraints{
		MinWidth:  20,
		MaxWidth:  runtime.Infinity,
		MinHeight: 5,
		MaxHeight: runtime.Infinity,
	}
	size := listVNode.Measure(minConstraints)
	if size.Width < 20 {
		t.Errorf("Width should respect MinWidth: got %d, expected >= 20", size.Width)
	}

	// Test maximum width constraint
	maxConstraints := runtime.BoxConstraints{
		MinWidth:  5,
		MaxWidth:  15,
		MinHeight: 5,
		MaxHeight: runtime.Infinity,
	}
	size = listVNode.Measure(maxConstraints)
	if size.Width > 15 {
		t.Errorf("Width should respect MaxWidth: got %d, expected <= 15", size.Width)
	}

	// Test height constraint
	heightConstraints := runtime.BoxConstraints{MinWidth: 10, MinHeight: 10, MaxWidth: runtime.Infinity, MaxHeight: 3}
	size = listVNode.Measure(heightConstraints)
	if size.Height > 3 {
		t.Errorf("Height should respect MaxHeight: got %d, expected <= 3", size.Height)
	}
}