package runtime

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/style"
)

// TestSelectionManager_NewManager tests creating a new selection manager.
func TestSelectionManager_NewManager(t *testing.T) {
	manager := NewSelectionManager()

	if manager == nil {
		t.Fatal("NewSelectionManager returned nil")
	}

	if !manager.IsEnabled() {
		t.Error("New manager should be enabled by default")
	}

	if manager.IsActive() {
		t.Error("New manager should not be active")
	}
}

// TestSelectionManager_Start tests starting a selection.
func TestSelectionManager_Start(t *testing.T) {
	buffer := NewCellBuffer(20, 10)
	manager := NewSelectionManager()
	manager.SetBuffer(buffer)

	manager.Start(5, 3)

	if !manager.IsActive() {
		t.Error("Manager should be active after Start")
	}

	startX, endX, startY, endY := manager.GetSelectionRange()
	if startX != 5 || endX != 5 || startY != 3 || endY != 3 {
		t.Errorf("Expected range (5,5,3,3), got (%d,%d,%d,%d)", startX, endX, startY, endY)
	}
}

// TestSelectionManager_Update tests updating selection.
func TestSelectionManager_Update(t *testing.T) {
	buffer := NewCellBuffer(20, 10)
	manager := NewSelectionManager()
	manager.SetBuffer(buffer)

	manager.Start(2, 2)
	manager.Update(8, 5)

	if !manager.IsActive() {
		t.Error("Manager should be active after Update")
	}

	startX, endX, startY, endY := manager.GetSelectionRange()
	if startX != 2 || endX != 8 || startY != 2 || endY != 5 {
		t.Errorf("Expected range (2,8,2,5), got (%d,%d,%d,%d)", startX, endX, startY, endY)
	}
}

// TestSelectionManager_IsSelected tests checking if a cell is selected.
func TestSelectionManager_IsSelected(t *testing.T) {
	buffer := NewCellBuffer(20, 10)
	manager := NewSelectionManager()
	manager.SetBuffer(buffer)

	// Start a selection
	manager.Start(5, 3)
	manager.Update(10, 3)

	// Test selected cells
	if !manager.IsSelected(5, 3) {
		t.Error("Cell (5,3) should be selected")
	}
	if !manager.IsSelected(8, 3) {
		t.Error("Cell (8,3) should be selected")
	}
	if !manager.IsSelected(10, 3) {
		t.Error("Cell (10,3) should be selected")
	}

	// Test unselected cells
	if manager.IsSelected(4, 3) {
		t.Error("Cell (4,3) should not be selected")
	}
	if manager.IsSelected(11, 3) {
		t.Error("Cell (11,3) should not be selected")
	}
	if manager.IsSelected(5, 4) {
		t.Error("Cell (5,4) should not be selected")
	}
}

// TestSelectionManager_Clear tests clearing selection.
func TestSelectionManager_Clear(t *testing.T) {
	buffer := NewCellBuffer(20, 10)
	manager := NewSelectionManager()
	manager.SetBuffer(buffer)

	manager.Start(5, 3)
	manager.Update(10, 3)

	if !manager.IsActive() {
		t.Error("Manager should be active")
	}

	manager.Clear()

	if manager.IsActive() {
		t.Error("Manager should not be active after Clear")
	}

	if manager.IsSelected(5, 3) {
		t.Error("Cell should not be selected after Clear")
	}
}

// TestSelectionManager_SelectAll tests selecting all text.
func TestSelectionManager_SelectAll(t *testing.T) {
	buffer := NewCellBuffer(20, 10)
	manager := NewSelectionManager()
	manager.SetBuffer(buffer)

	manager.SelectAll()

	startX, endX, startY, endY := manager.GetSelectionRange()
	if startX != 0 || endX != 19 || startY != 0 || endY != 9 {
		t.Errorf("Expected range (0,19,0,9), got (%d,%d,%d,%d)", startX, endX, startY, endY)
	}

	if !manager.IsActive() {
		t.Error("Manager should be active after SelectAll")
	}
}

// TestSelectionManager_SelectWord tests word selection.
func TestSelectionManager_SelectWord(t *testing.T) {
	buffer := NewCellBuffer(30, 5)
	manager := NewSelectionManager()
	manager.SetBuffer(buffer)

	// Add text to buffer
	text := "hello world test"
	for x, ch := range text {
		buffer.SetContent(x, 2, 0, ch, CellStyle{}, "")
	}

	// Select the word "world"
	manager.SelectWord(7, 2)

	startX, endX, startY, endY := manager.GetSelectionRange()
	if startY != 2 || endY != 2 {
		t.Error("Word selection should be on the same line")
	}

	// Check that the selection covers "world" (positions 6-10)
	if startX > 6 || endX < 10 {
		t.Errorf("Expected selection to include 'world', got range (%d,%d)", startX, endX)
	}
}

// TestSelectionManager_SelectLine tests line selection.
func TestSelectionManager_SelectLine(t *testing.T) {
	buffer := NewCellBuffer(20, 5)
	manager := NewSelectionManager()
	manager.SetBuffer(buffer)

	manager.SelectLine(2)

	startX, endX, startY, endY := manager.GetSelectionRange()

	// Should select entire line 2
	if startY != 2 || endY != 2 {
		t.Error("Line selection should be on the same line")
	}

	if startX != 0 || endX != 19 {
		t.Errorf("Expected range (0,19,2,2), got (%d,%d,%d,%d)", startX, endX, startY, endY)
	}
}

// TestSelectionManager_SetEnabled tests enabling/disabling selection.
func TestSelectionManager_SetEnabled(t *testing.T) {
	buffer := NewCellBuffer(20, 10)
	manager := NewSelectionManager()
	manager.SetBuffer(buffer)

	// Try to start selection when disabled
	manager.SetEnabled(false)
	manager.Start(5, 3)

	if manager.IsActive() {
		t.Error("Selection should not start when disabled")
	}

	// Enable and try again
	manager.SetEnabled(true)
	manager.Start(5, 3)

	if !manager.IsActive() {
		t.Error("Selection should start when enabled")
	}
}

// TestSelectionManager_GetSelectedText tests getting selected text.
func TestSelectionManager_GetSelectedText(t *testing.T) {
	buffer := NewCellBuffer(20, 5)
	manager := NewSelectionManager()
	manager.SetBuffer(buffer)

	// Add text to buffer
	text := "test line"
	for x, ch := range text {
		buffer.SetContent(x, 1, 0, ch, CellStyle{}, "")
	}

	manager.Start(0, 1)
	manager.Update(len(text)-1, 1)

	selectedText := manager.GetSelectedText()
	if selectedText != "test line" {
		t.Errorf("Expected 'test line', got '%s'", selectedText)
	}
}

// TestSelectionManager_MultiLineSelection tests multi-line selection.
func TestSelectionManager_MultiLineSelection(t *testing.T) {
	buffer := NewCellBuffer(20, 5)
	manager := NewSelectionManager()
	manager.SetBuffer(buffer)

	// Add text to buffer
	line1 := "first line"
	line2 := "second line"
	for x, ch := range line1 {
		buffer.SetContent(x, 1, 0, ch, CellStyle{}, "")
	}
	for x, ch := range line2 {
		buffer.SetContent(x, 2, 0, ch, CellStyle{}, "")
	}

	manager.Start(0, 1)
	manager.Update(len(line2)-1, 2)

	selectedText := manager.GetSelectedText()
	// Should contain both lines with newline
	if selectedText != "first line\nsecond line" {
		t.Errorf("Expected 'first line\\nsecond line', got '%s'", selectedText)
	}
}

// TestSelectionManager_ApplyHighlight tests applying highlight to buffer.
func TestSelectionManager_ApplyHighlight(t *testing.T) {
	buffer := NewCellBuffer(20, 5)
	manager := NewSelectionManager()
	manager.SetBuffer(buffer)

	// Add text to buffer
	text := "highlight me"
	for x, ch := range text {
		buffer.SetContent(x, 1, 0, ch, CellStyle{}, "")
	}

	// Create selection
	manager.Start(0, 1)
	manager.Update(len(text)-1, 1)

	// Apply highlight
	manager.ApplyHighlight()

	// Check that selected cells have the highlight style
	for x := 0; x < len(text); x++ {
		cell := buffer.GetContent(x, 1)
		if !cell.Selected {
			t.Errorf("Cell (%d,1) should be marked as selected", x)
		}
	}

	// Check that non-selected cells are not marked
	cell := buffer.GetContent(0, 2)
	if cell.Selected {
		t.Error("Cell (0,2) should not be marked as selected")
	}
}

// TestSelectionManager_GetSelectedCells tests getting all selected cells.
func TestSelectionManager_GetSelectedCells(t *testing.T) {
	buffer := NewCellBuffer(10, 5)
	manager := NewSelectionManager()
	manager.SetBuffer(buffer)

	// Add text to buffer
	text := "test"
	for x, ch := range text {
		buffer.SetContent(x, 2, 0, ch, CellStyle{}, "")
	}

	manager.Start(0, 2)
	manager.Update(len(text)-1, 2)

	cells := manager.GetSelectedCells()

	if len(cells) != len(text) {
		t.Errorf("Expected %d cells, got %d", len(text), len(cells))
	}

	for i, cell := range cells {
		expectedX := i
		if cell.X != expectedX {
			t.Errorf("Expected cell at X=%d, got X=%d", expectedX, cell.X)
		}
		if cell.Y != 2 {
			t.Errorf("Expected cell at Y=2, got Y=%d", cell.Y)
		}
	}
}

// TestSelectionManager_GetRegion tests getting selection region.
func TestSelectionManager_GetRegion(t *testing.T) {
	buffer := NewCellBuffer(20, 10)
	manager := NewSelectionManager()
	manager.SetBuffer(buffer)

	manager.Start(5, 3)
	manager.Update(10, 7)

	region := manager.GetRegion()

	if region.StartX != 5 || region.StartY != 3 || region.EndX != 10 || region.EndY != 7 {
		t.Errorf("Expected region (5,3,10,7), got (%d,%d,%d,%d)", region.StartX, region.StartY, region.EndX, region.EndY)
	}
}

// TestDefaultSelectionHighlight tests the default highlight style.
func TestDefaultSelectionHighlight(t *testing.T) {
	highlight := DefaultSelectionHighlight()

	if !highlight.IsReverse() {
		t.Error("Default highlight should use reverse video")
	}
}

// TestLightSelectionHighlight tests the light theme highlight.
func TestLightSelectionHighlight(t *testing.T) {
	highlight := LightSelectionHighlight()

	// Check that the style is not empty (has some styling applied)
	if highlight == (style.Style{}) {
		t.Error("Light highlight should have styling")
	}
}

// TestDarkSelectionHighlight tests the dark theme highlight.
func TestDarkSelectionHighlight(t *testing.T) {
	highlight := DarkSelectionHighlight()

	// Check that the style is not empty (has some styling applied)
	if highlight == (style.Style{}) {
		t.Error("Dark highlight should have styling")
	}
}

// TestSelectionManager_StyledText tests selecting styled text with ANSI codes.
func TestSelectionManager_StyledText(t *testing.T) {
	buffer := NewCellBuffer(30, 5)
	manager := NewSelectionManager()
	manager.SetBuffer(buffer)

	// Directly set cells with styled text (ANSI codes)
	// Note: In real usage, styled text would be set by the rendering system
	styledText := "\x1b[31mred\x1b[0m"
	buffer.Cells[0][1] = Cell{
		Cluster:         styledText,
		IsContinuation: false,
		Style:           CellStyle{},
		ZIndex:          0,
	}

	manager.Start(0, 0)
	manager.Update(10, 0)

	selectedText := manager.GetSelectedText()
	// The extracted text should be " red" (with leading space from position 0, then "red" from position 1)
	if selectedText != " red" {
		t.Errorf("Expected ' red' (with leading space), got '%s'", selectedText)
	}
}

// TestSelectionManager_ContinuationCell tests handling continuation cells.
func TestSelectionManager_ContinuationCell(t *testing.T) {
	buffer := NewCellBuffer(20, 5)
	manager := NewSelectionManager()
	manager.SetBuffer(buffer)

	// Create a wide character followed by continuation cells
	// Position 0: main cell with wide character
	buffer.Cells[0][0] = Cell{
		Cluster:         "👨‍👩‍👧‍👦", // Emoji ZWJ sequence (wide character)
		Width:           2,
		IsContinuation:  false,
		Style:           CellStyle{},
		ZIndex:          0,
	}

	// Position 1: continuation cell
	buffer.Cells[1][0] = Cell{
		Cluster:         " ",
		Width:           0,
		IsContinuation:  true,
		Style:           CellStyle{},
		ZIndex:          0,
	}

	manager.Start(0, 0)
	manager.Update(1, 0)

	selectedText := manager.GetSelectedText()
	// Should handle continuation cells properly
	if len(selectedText) == 0 {
		t.Error("Should extract text from continuation cells")
	}
}

// TestFindStyledTextStart tests the findStyledTextStart function.
func TestFindStyledTextStart(t *testing.T) {
	buffer := NewCellBuffer(10, 3)

	// Test 1: Finding styled text from a continuation cell
	// Set up cells with styled text at position 2
	buffer.Cells[1][2] = Cell{
		Cluster:         "\x1b[31mred\x1b[0m",
		IsContinuation:  false,
		Style:           CellStyle{},
		ZIndex:          0,
	}

	// Add continuation cells at positions 3-4
	buffer.Cells[1][3] = Cell{
		Cluster:         " ",
		IsContinuation:  true,
		Style:           CellStyle{},
		ZIndex:          0,
	}
	buffer.Cells[1][4] = Cell{
		Cluster:         " ",
		IsContinuation:  true,
		Style:           CellStyle{},
		ZIndex:          0,
	}

	// Test finding the start from position 4
	styledText, offset := findStyledTextStart(buffer, 4, 1)

	if styledText == "" {
		t.Error("Should find styled text")
	}

	if styledText != "\x1b[31mred\x1b[0m" {
		t.Errorf("Expected styled text '\\x1b[31mred\\x1b[0m', got '%s'", styledText)
	}

	if offset != 2 {
		t.Errorf("Expected offset 2, got %d", offset)
	}

	// Test 2: Finding regular text from a continuation cell
	buffer2 := NewCellBuffer(10, 3)
	buffer2.Cells[1][3] = Cell{
		Cluster:         " ",
		IsContinuation:  true,
		Style:           CellStyle{},
		ZIndex:          0,
	}
	buffer2.Cells[1][2] = Cell{
		Cluster:         "hello",
		IsContinuation:  false,
		Style:           CellStyle{},
		ZIndex:          0,
	}

	// Position 3 is a continuation cell, should find the text at position 2
	styledText2, offset2 := findStyledTextStart(buffer2, 3, 1)

	if styledText2 != "hello" {
		t.Errorf("Expected 'hello', got '%s'", styledText2)
	}

	if offset2 != 1 {
		t.Errorf("Expected offset 1, got %d", offset2)
	}

	// Test 3: Empty buffer (all default cells with space character)
	buffer3 := NewCellBuffer(10, 3)
	styledText3, offset3 := findStyledTextStart(buffer3, 5, 1)

	if styledText3 != "" {
		t.Errorf("Expected empty string for empty buffer, got '%s'", styledText3)
	}

	if offset3 != 0 {
		t.Errorf("Expected offset 0, got %d", offset3)
	}

	// Test 4: Directly on a cell with content
	buffer4 := NewCellBuffer(10, 3)
	buffer4.Cells[1][5] = Cell{
		Cluster:         "x",
		IsContinuation:  false,
		Style:           CellStyle{},
		ZIndex:          0,
	}

	// Position 5 has a regular cell, should return that cell's content
	styledText4, offset4 := findStyledTextStart(buffer4, 5, 1)

	if styledText4 != "x" {
		t.Errorf("Expected 'x', got '%s'", styledText4)
	}

	if offset4 != 0 {
		t.Errorf("Expected offset 0, got %d", offset4)
	}
}

// TestExtractVisibleText tests extracting visible text from styled text.
func TestExtractVisibleText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"No ANSI codes", "hello", "hello"},
		{"Red text", "\x1b[31mhello\x1b[0m", "hello"},
		{"Bold red", "\x1b[1;31mhello\x1b[0m", "hello"},
		{"Multiple ANSI", "\x1b[31m\x1b[1mhello\x1b[0m\x1b[0m", "hello"},
		{"Empty string", "", ""},
		{"Only ANSI codes", "\x1b[31m\x1b[0m", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractVisibleText(tt.input)
			if result != tt.expected {
				t.Errorf("extractVisibleText(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestCountVisibleCharsInStyledText tests counting visible characters.
func TestCountVisibleCharsInStyledText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"No ANSI codes", "hello", 5},
		{"Red text", "\x1b[31mhello\x1b[0m", 5},
		{"Bold red", "\x1b[1;31mhello\x1b[0m", 5},
		{"Empty string", "", 0},
		{"Only ANSI codes", "\x1b[31m\x1b[0m", 0},
		{"Unicode characters", "👨‍👩‍👧‍👦", 7}, // ZWJ sequence
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := countVisibleCharsInStyledText(tt.input)
			if result != tt.expected {
				t.Errorf("countVisibleCharsInStyledText(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}
