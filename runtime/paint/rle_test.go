// Package paint provides tests for RLE optimization.
package paint

import (
	"strings"
	"testing"

	"github.com/wwsheng009/mint/runtime/style"
)

// =============================================================================
// RLE Tests
// =============================================================================

func TestEncodeRLE_Empty(t *testing.T) {
	runs := EncodeRLE([]Cell{}, 0)
	if runs != nil {
		t.Errorf("EncodeRLE() on empty = %v, want nil", runs)
	}
}

func TestEncodeRLE_SingleCell(t *testing.T) {
	row := []Cell{
		{Cluster: "A"},
	}
	runs := EncodeRLE(row, 1)

	if len(runs) != 1 {
		t.Fatalf("EncodeRLE() returned %d runs, want 1", len(runs))
	}

	if runs[0].Count != 1 {
		t.Errorf("runs[0].Count = %d, want 1", runs[0].Count)
	}
}

func TestEncodeRLE_MultipleRuns(t *testing.T) {
	style1 := style.Style{}.Foreground("red")
	style2 := style.Style{}.Foreground("blue")

	row := []Cell{
		{Cluster: "A", Style: style1},
		{Cluster: "A", Style: style1},
		{Cluster: "A", Style: style1},
		{Cluster: "B", Style: style2},
		{Cluster: "B", Style: style2},
	}

	runs := EncodeRLE(row, 5)

	if len(runs) != 2 {
		t.Fatalf("EncodeRLE() returned %d runs, want 2", len(runs))
	}

	if runs[0].Count != 3 {
		t.Errorf("runs[0].Count = %d, want 3", runs[0].Count)
	}
	if runs[1].Count != 2 {
		t.Errorf("runs[1].Count = %d, want 2", runs[1].Count)
	}
}

func TestEncodeRLE_SkipsContinuation(t *testing.T) {
	row := []Cell{
		{Cluster: "A", Width: 2, IsContinuation: false},
		{Cluster: "", Width: 0, IsContinuation: true},
		{Cluster: "B", Width: 1, IsContinuation: false},
	}

	runs := EncodeRLE(row, 3)

	// Should count non-continuation cells only
	if len(runs) != 2 {
		t.Fatalf("EncodeRLE() returned %d runs, want 2", len(runs))
	}
}

func TestRLERenderer_RenderRow(t *testing.T) {
	renderer := NewRLERenderer()

	// Use numeric color codes to avoid text interference
	style1 := style.Style{}.Foreground("#FF0000").Bold(true)
	style2 := style.Style{}.Foreground("#0000FF")

	row := []Cell{
		{Cluster: "H", Style: style1},
		{Cluster: "e", Style: style1},
		{Cluster: "l", Style: style1},
		{Cluster: "l", Style: style1},
		{Cluster: "o", Style: style2},
	}

	output := renderer.RenderRow(row, 5, 0)

	if output == "" {
		t.Fatal("RenderRow() returned empty string")
	}

	// Should contain ANSI codes
	if !strings.Contains(output, "\x1b[") {
		t.Error("Output should contain ANSI escape codes")
	}

	// Should contain the text (check each character separately)
	if !strings.Contains(output, "H") {
		t.Error("Output should contain 'H'")
	}
	if !strings.Contains(output, "e") {
		t.Error("Output should contain 'e'")
	}
	// Check for two 'l' characters
	lCount := strings.Count(output, "l")
	if lCount < 2 {
		t.Errorf("Output should contain at least 2 'l' characters, got %d", lCount)
	}
	if !strings.Contains(output, "o") {
		t.Error("Output should contain 'o'")
	}
}

func TestOptimizedOutput_NoChanges(t *testing.T) {
	buf := NewBuffer(10, 10)
	diff := &DiffResult{HasChanges: false}

	output := OptimizedOutput(buf, diff)

	if output != "" {
		t.Errorf("OptimizedOutput() with no changes = %q, want empty", output)
	}
}

func TestOptimizedOutput_WithChanges(t *testing.T) {
	buf := NewBuffer(10, 10)

	// Set some cells
	buf.SetCell(0, 0, 'A', style.Style{}.Foreground("red"))
	buf.SetCell(1, 0, 'B', style.Style{}.Foreground("blue"))
	buf.SetCell(2, 0, 'C', style.Style{}.Foreground("green"))

	diff := &DiffResult{
		HasChanges: true,
		DirtyRegions: []Rect{
			{X: 0, Y: 0, Width: 3, Height: 1},
		},
	}

	output := OptimizedOutput(buf, diff)

	if output == "" {
		t.Fatal("OptimizedOutput() returned empty string")
	}

	// Should contain ANSI codes
	if !strings.Contains(output, "\x1b[") {
		t.Error("Output should contain ANSI escape codes")
	}
}

func TestCursorMove_NoMovement(t *testing.T) {
	output := cursorMove(5, 5, 0)
	if output != "" {
		t.Errorf("cursorMove(5, 5, 0) = %q, want empty", output)
	}
}

func TestCursorMove_SmallForward(t *testing.T) {
	output := cursorMove(5, 7, 0)
	if !strings.Contains(output, "\x1b[2C") {
		t.Errorf("cursorMove(5, 7, 0) should contain 2C, got %q", output)
	}
}

func TestCursorMove_SmallBackward(t *testing.T) {
	output := cursorMove(7, 5, 0)
	if !strings.Contains(output, "\x1b[2D") {
		t.Errorf("cursorMove(7, 5, 0) should contain 2D, got %q", output)
	}
}

func TestCursorMove_Absolute(t *testing.T) {
	output := cursorMove(0, 20, 5)
	if !strings.Contains(output, "\x1b[6;21H") {
		t.Errorf("cursorMove(0, 20, 5) should contain absolute positioning, got %q", output)
	}
}

func TestStyleToANSI_Reset(t *testing.T) {
	output := styleToANSI(style.Style{})
	if output != "\x1b[0m" {
		t.Errorf("styleToANSI(empty) = %q, want \\x1b[0m", output)
	}
}

func TestStyleToANSI_Bold(t *testing.T) {
	output := styleToANSI(style.Style{}.Bold(true))
	if !strings.Contains(output, "1") {
		t.Errorf("styleToANSI(Bold) should contain '1', got %q", output)
	}
}

func TestStyleToANSI_Colors(t *testing.T) {
	output := styleToANSI(style.Style{}.
		Foreground("red").
		Background("blue"))

	if !strings.Contains(output, "red") {
		t.Error("styleToANSI with colors should contain 'red'")
	}
	if !strings.Contains(output, "blue") {
		t.Error("styleToANSI with colors should contain 'blue'")
	}
}

func TestAnalyzeBuffer_Empty(t *testing.T) {
	buf := NewBuffer(0, 0)
	stats := AnalyzeBuffer(buf)

	if stats.TotalCells != 0 {
		t.Errorf("TotalCells = %d, want 0", stats.TotalCells)
	}
}

func TestAnalyzeBuffer_Simple(t *testing.T) {
	buf := NewBuffer(10, 10)
	stats := AnalyzeBuffer(buf)

	if stats.TotalCells != 100 {
		t.Errorf("TotalCells = %d, want 100", stats.TotalCells)
	}
}

func TestRLEStats_RecordFrame(t *testing.T) {
	stats := &RLEStats{}

	stats.RecordFrame(100, 10, 500)

	if stats.FramesRendered != 1 {
		t.Errorf("FramesRendered = %d, want 1", stats.FramesRendered)
	}
	if stats.TotalCells != 100 {
		t.Errorf("TotalCells = %d, want 100", stats.TotalCells)
	}
	if stats.DirtyCells != 10 {
		t.Errorf("DirtyCells = %d, want 10", stats.DirtyCells)
	}
	if stats.OutputBytes != 500 {
		t.Errorf("OutputBytes = %d, want 500", stats.OutputBytes)
	}
}

func TestRLEStats_String(t *testing.T) {
	stats := &RLEStats{}
	stats.RecordFrame(100, 10, 500)

	output := stats.String()

	if !strings.Contains(output, "Frames:") {
		t.Error("String() should contain 'Frames:'")
	}
	if !strings.Contains(output, "Dirty/Total:") {
		t.Error("String() should contain 'Dirty/Total:'")
	}
}

// =============================================================================
// Integration Tests
// =============================================================================

func TestRLE_FullCycle(t *testing.T) {
	// Create a buffer
	buf := NewBuffer(20, 5)

	// Paint some content
	style1 := style.Style{}.Foreground("red").Bold(true)
	style2 := style.Style{}.Foreground("blue").Underline(true)

	for i := 0; i < 10; i++ {
		buf.SetCell(i, 0, 'X', style1)
	}
	for i := 10; i < 20; i++ {
		buf.SetCell(i, 0, 'Y', style2)
	}

	// Encode the row
	row := buf.Cells[0]
	runs := EncodeRLE(row, 20)

	// Should have 2 runs (one for X's, one for Y's)
	if len(runs) != 2 {
		t.Errorf("Got %d runs, want 2", len(runs))
	}

	// Test the renderer
	renderer := NewRLERenderer()
	output := renderer.RenderRow(row, 20, 0)

	if output == "" {
		t.Fatal("Renderer output is empty")
	}

	// Should contain both styles
	if !strings.Contains(output, "red") {
		t.Error("Output should contain 'red' style")
	}
	if !strings.Contains(output, "blue") {
		t.Error("Output should contain 'blue' style")
	}
}

func TestOptimizedOutput_Integration(t *testing.T) {
	buf := NewBuffer(10, 10)

	// Create content
	style1 := style.Style{}.Foreground("green")
	for i := 0; i < 10; i++ {
		buf.SetCell(i, 0, 'A', style1)
	}

	// Create diff result
	diff := &DiffResult{
		HasChanges: true,
		DirtyRegions: []Rect{
			{X: 0, Y: 0, Width: 10, Height: 1},
		},
		ChangedCells: 10,
	}

	// Generate output
	output := OptimizedOutput(buf, diff)

	if output == "" {
		t.Fatal("OptimizedOutput() returned empty")
	}

	// Verify it contains our content
	if !strings.Contains(output, "A") {
		t.Error("Output should contain 'A'")
	}
	if !strings.Contains(output, "green") {
		t.Error("Output should contain 'green' style")
	}
}

func TestRLE_CompressionRatio(t *testing.T) {
	// Create a row with many repeated cells
	buf := NewBuffer(100, 1)
	style1 := style.Style{}.Foreground("red")

	for i := 0; i < 100; i++ {
		buf.SetCell(i, 0, 'X', style1)
	}

	// Encode
	row := buf.Cells[0]
	runs := EncodeRLE(row, 100)

	// Should compress to a single run
	if len(runs) != 1 {
		t.Errorf("Got %d runs, want 1 (perfect compression)", len(runs))
	}

	if runs[0].Count != 100 {
		t.Errorf("Run.Count = %d, want 100", runs[0].Count)
	}
}
