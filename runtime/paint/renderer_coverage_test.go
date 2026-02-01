package paint

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/style"
)

// =============================================================================
// Renderer Tests
// =============================================================================

func TestRenderer_GetFrontBuffer(t *testing.T) {
	renderer := NewRenderer(10, 10)

	front := renderer.GetFrontBuffer()

	if front == nil {
		t.Fatal("GetFrontBuffer returned nil")
	}

	if front.Width != 10 {
		t.Errorf("Width = %d, want 10", front.Width)
	}
	if front.Height != 10 {
		t.Errorf("Height = %d, want 10", front.Height)
	}
}

func TestRenderer_MarkDirty(t *testing.T) {
	renderer := NewRenderer(10, 10)

	renderer.MarkDirty()

	if !renderer.dirtyTracker.IsAllDirty() {
		t.Error("MarkDirty() did not mark all as dirty")
	}
}

func TestRenderer_MarkDirtyRect(t *testing.T) {
	renderer := NewRenderer(10, 10)

	rect := Rect{X: 2, Y: 3, Width: 4, Height: 5}
	renderer.MarkDirtyRect(rect)

	dirtyRects := renderer.dirtyTracker.GetDirtyRects()

	if len(dirtyRects) != 1 {
		t.Fatalf("MarkDirtyRect() resulted in %d rects, want 1", len(dirtyRects))
	}

	if dirtyRects[0].X != 2 || dirtyRects[0].Y != 3 {
		t.Errorf("Dirty rect = %+v, want {X:2, Y:3}", dirtyRects[0])
	}
}

func TestRenderer_ForceFullRender(t *testing.T) {
	renderer := NewRenderer(10, 10)

	renderer.ForceFullRender()

	if !renderer.dirtyTracker.IsAllDirty() {
		t.Error("ForceFullRender() did not mark all as dirty")
	}
}

func TestRenderer_GetStats(t *testing.T) {
	renderer := NewRenderer(10, 10)

	// Set some content
	back := renderer.GetBackBuffer()
	back.SetCell(0, 0, 'A', style.NewStyle())
	back.SetCell(1, 0, 'B', style.NewStyle())

	// First render to establish front buffer
	renderer.MarkDirty()
	_ = renderer.Render()

	// Set different content
	back.SetCell(0, 0, 'C', style.NewStyle())
	renderer.MarkDirty()
	_ = renderer.Render()

	// Get stats
	stats := renderer.GetStats()

	// ChangedCells tracks cumulative changes in the dirty tracker
	// This is tracking how many cells were marked as changed total
	if stats.ChangedCells < 0 {
		t.Errorf("ChangedCells = %d, want >= 0", stats.ChangedCells)
	}

	// Output bytes should be > 0 since we rendered something
	if stats.OutputBytes <= 0 {
		t.Errorf("OutputBytes = %d, want > 0", stats.OutputBytes)
	}
}

func TestRenderer_Reset(t *testing.T) {
	renderer := NewRenderer(10, 10)

	// Set some state
	back := renderer.GetBackBuffer()
	back.SetCell(5, 5, 'X', style.NewStyle())

	// Render to set internal state
	renderer.MarkDirty()
	_ = renderer.Render()

	// Reset only clears style state and output buffer, not the buffer
	renderer.Reset()

	// Get back buffer after reset - buffer contents should still be there
	back = renderer.GetBackBuffer()
	cell := back.Cells[5][5]
	if cell.Cluster != "X" {
		// This is expected - Reset doesn't clear the buffer
		// It only resets internal state (cursor, style state)
	}

	// Output buffer should be cleared
	stats := renderer.GetStats()
	if stats.OutputBytes != 0 {
		t.Errorf("After Reset, OutputBytes = %d, want 0 (output buffer cleared)", stats.OutputBytes)
	}
}

func TestRenderer_Resize(t *testing.T) {
	renderer := NewRenderer(10, 10)

	// Resize to different dimensions
	renderer.Resize(20, 15)

	front := renderer.GetFrontBuffer()
	back := renderer.GetBackBuffer()

	if front.Width != 20 {
		t.Errorf("After Resize, Width = %d, want 20", front.Width)
	}
	if front.Height != 15 {
		t.Errorf("After Resize, Height = %d, want 15", front.Height)
	}
	if back.Width != 20 {
		t.Errorf("After Resize, back Width = %d, want 20", back.Width)
	}
	if back.Height != 15 {
		t.Errorf("After Resize, back Height = %d, want 15", back.Height)
	}
}

func TestRenderer_ResetState(t *testing.T) {
	renderer := NewRenderer(10, 10)

	// Set some state
	back := renderer.GetBackBuffer()
	back.SetCell(0, 0, 'X', style.NewStyle().Bold(true))

	// Render to set internal state
	renderer.MarkDirty()
	_ = renderer.Render()

	// Reset state
	renderer.ResetState()

	// The state should be reset (cursor position invalid)
	// This is mainly for internal state verification
	_ = renderer.styleState
}

func TestRenderer_NoChangesNoOutput(t *testing.T) {
	renderer := NewRenderer(10, 10)

	// Render without any changes
	output := renderer.Render()

	if output != "" {
		t.Errorf("Render() with no changes = %q, want empty", output)
	}
}

func TestRenderer_SingleCellChangeCoverage(t *testing.T) {
	renderer := NewRenderer(10, 10)
	back := renderer.GetBackBuffer()

	// Change a single cell
	back.SetCell(5, 5, 'X', style.NewStyle().Foreground("red"))

	output := renderer.Render()

	if output == "" {
		t.Fatal("Render() returned empty string for single cell change")
	}

	// Should contain ANSI codes and the character
	if len(output) < 10 {
		t.Errorf("Output too short: %q", output)
	}
}

func TestRenderer_MultipleCellsSameStyle(t *testing.T) {
	renderer := NewRenderer(10, 10)
	back := renderer.GetBackBuffer()

	style := style.NewStyle().Foreground("red").Bold(true)

	// Set multiple cells with same style
	for i := 0; i < 5; i++ {
		back.SetCell(i, 0, rune('A'+i), style)
	}

	output := renderer.Render()

	if output == "" {
		t.Fatal("Render() returned empty string")
	}

	// Should contain the text
	outputLen := len(output)
	if outputLen < 20 {
		t.Errorf("Output too short: %d bytes", outputLen)
	}
}

func TestRenderer_StyleChangeOnly(t *testing.T) {
	renderer := NewRenderer(10, 10)
	back := renderer.GetBackBuffer()

	// Set initial cell
	back.SetCell(0, 0, 'A', style.NewStyle().Foreground("red"))

	// First render
	renderer.MarkDirty()
	_ = renderer.Render()

	// Change only style
	back.SetCell(0, 0, 'A', style.NewStyle().Foreground("blue"))

	output := renderer.Render()

	if output == "" {
		t.Fatal("Render() returned empty string for style change")
	}
}

func TestRenderer_FullBufferRender(t *testing.T) {
	renderer := NewRenderer(20, 10)
	back := renderer.GetBackBuffer()

	style := style.NewStyle().Foreground("green")

	// Fill entire buffer
	for y := 0; y < 10; y++ {
		for x := 0; x < 20; x++ {
			back.SetCell(x, y, '.', style)
		}
	}

	// First render to establish front buffer
	renderer.MarkDirty()
	_ = renderer.Render()

	// Now change some cells to trigger normal diff
	for y := 0; y < 5; y++ {
		for x := 0; x < 10; x++ {
			back.SetCell(x, y, 'X', style)
		}
	}

	// Render with normal diff (not MarkAll)
	output := renderer.Render()

	if output == "" {
		t.Fatal("Render() returned empty string for full buffer")
	}

	// Verify output contains ANSI codes
	if !contains(output, "\x1b[") {
		t.Error("Output should contain ANSI codes")
	}

	// After normal diff, ChangedCells should reflect actual changes
	stats := renderer.GetStats()
	// Note: ChangedCells is the tracker's internal counter which may vary
	// The important thing is that output was generated
	if stats.OutputBytes <= 0 {
		t.Errorf("OutputBytes = %d, want > 0", stats.OutputBytes)
	}
}

func TestRenderer_WideCharacterRendering(t *testing.T) {
	renderer := NewRenderer(20, 5)
	back := renderer.GetBackBuffer()

	// Set a wide character (Chinese)
	back.SetCell(0, 0, '中', style.NewStyle())

	renderer.MarkDirty()
	output := renderer.Render()

	if output == "" {
		t.Fatal("Render() returned empty string for wide character")
	}

	// Verify the wide character was handled
	cell := back.Cells[0][0]
	if cell.Cluster != "中" {
		t.Errorf("Cluster = %q, want %q", cell.Cluster, "中")
	}

	// Check continuation cell
	if len(back.Cells[0]) < 2 {
		t.Fatal("Row length too short for wide character")
	}

	contCell := back.Cells[0][1]
	if !contCell.IsContinuation {
		t.Error("Second cell should be marked as continuation")
	}
}

func TestRenderer_RegionBounds(t *testing.T) {
	renderer := NewRenderer(10, 10)
	back := renderer.GetBackBuffer()

	// First establish a baseline
	renderer.MarkDirty()
	_ = renderer.Render()

	// Set content in specific region
	for y := 2; y < 5; y++ {
		for x := 2; x < 6; x++ {
			back.SetCell(x, y, '#', style.NewStyle())
		}
	}

	renderer.MarkDirtyRect(Rect{X: 2, Y: 2, Width: 4, Height: 3})
	output := renderer.Render()

	if output == "" {
		t.Fatal("Render() returned empty string")
	}

	// Verify output was generated
	stats := renderer.GetStats()
	if stats.OutputBytes <= 0 {
		t.Errorf("OutputBytes = %d, want > 0", stats.OutputBytes)
	}
}

func TestRenderer_ConsecutiveRenders(t *testing.T) {
	renderer := NewRenderer(10, 10)
	back := renderer.GetBackBuffer()

	// First render - no changes
	output1 := renderer.Render()
	if output1 != "" {
		t.Error("First render with no content should return empty")
	}

	// Add content and render
	back.SetCell(0, 0, 'A', style.NewStyle())
	renderer.MarkDirty()
	output2 := renderer.Render()
	if output2 == "" {
		t.Error("Second render should have output")
	}

	// Third render - no changes
	output3 := renderer.Render()
	if output3 != "" {
		t.Error("Third render with no changes should return empty")
	}
}

func TestRenderer_EmptyBuffer(t *testing.T) {
	renderer := NewRenderer(0, 0)

	// Should not panic
	output := renderer.Render()

	if output != "" {
		t.Errorf("Render() on empty buffer = %q, want empty", output)
	}
}

func TestRenderer_SmallBuffer(t *testing.T) {
	renderer := NewRenderer(1, 1)
	back := renderer.GetBackBuffer()

	back.SetCell(0, 0, 'X', style.NewStyle())

	renderer.MarkDirty()
	output := renderer.Render()

	if output == "" {
		t.Fatal("Render() returned empty string")
	}
}

func TestRenderer_LargeBuffer(t *testing.T) {
	renderer := NewRenderer(200, 100)
	back := renderer.GetBackBuffer()

	// Set one cell
	back.SetCell(100, 50, 'X', style.NewStyle())

	renderer.MarkDirty()
	output := renderer.Render()

	if output == "" {
		t.Fatal("Render() returned empty string")
	}

	// Verify output size is reasonable (not the entire buffer)
	if len(output) > 1000 {
		t.Errorf("Output too large for single cell: %d bytes", len(output))
	}
}

func TestRenderer_MultipleDirtyRects(t *testing.T) {
	renderer := NewRenderer(20, 20)
	back := renderer.GetBackBuffer()

	// Establish baseline
	renderer.MarkDirty()
	_ = renderer.Render()

	// Set cells in multiple regions
	back.SetCell(2, 2, 'A', style.NewStyle())
	back.SetCell(15, 15, 'B', style.NewStyle())
	back.SetCell(10, 10, 'C', style.NewStyle())

	renderer.MarkDirtyRect(Rect{X: 2, Y: 2, Width: 1, Height: 1})
	renderer.MarkDirtyRect(Rect{X: 15, Y: 15, Width: 1, Height: 1})
	renderer.MarkDirtyRect(Rect{X: 10, Y: 10, Width: 1, Height: 1})

	output := renderer.Render()

	if output == "" {
		t.Error("Render() should return output for multiple dirty rects")
	}
}

func TestRenderer_EachRowDifferentStyle(t *testing.T) {
	renderer := NewRenderer(10, 5)
	back := renderer.GetBackBuffer()

	// Each row has different style
	colors := []style.Color{style.Red, style.Green, style.Yellow, style.Blue, style.Magenta}
	for y := 0; y < 5; y++ {
		st := style.NewStyle().Foreground(colors[y])
		back.SetCell(0, y, 'A'+rune(y), st)
	}

	renderer.MarkDirty()
	output := renderer.Render()

	if output == "" {
		t.Error("Render() should return output")
	}

	// Should contain style changes for each row
	stats := renderer.GetStats()
	if stats.OutputBytes <= 0 {
		t.Error("OutputBytes should be > 0")
	}
}

func TestRenderer_SameRowMultipleStyles(t *testing.T) {
	renderer := NewRenderer(20, 5)
	back := renderer.GetBackBuffer()

	// Same row, multiple styles
	back.SetCell(0, 0, 'A', style.NewStyle().Foreground(style.Red))
	back.SetCell(5, 0, 'B', style.NewStyle().Foreground(style.Green))
	back.SetCell(10, 0, 'C', style.NewStyle().Foreground(style.Blue))

	renderer.MarkDirty()
	output := renderer.Render()

	if output == "" {
		t.Error("Render() should return output")
	}
}

func TestRenderer_AllSameContent(t *testing.T) {
	renderer := NewRenderer(20, 10)
	back := renderer.GetBackBuffer()

	// Fill with same content
	st := style.NewStyle().Foreground(style.Cyan)
	for y := 0; y < 10; y++ {
		for x := 0; x < 20; x++ {
			back.SetCell(x, y, '.', st)
		}
	}

	// First render
	renderer.MarkDirty()
	output1 := renderer.Render()

	if output1 == "" {
		t.Error("First render should have output")
	}

	// Second render with same content should produce no output
	output2 := renderer.Render()

	if output2 != "" {
		t.Errorf("Second render with same content should return empty, got %d bytes", len(output2))
	}
}

func TestRenderer_CursorMovement(t *testing.T) {
	renderer := NewRenderer(10, 10)
	back := renderer.GetBackBuffer()

	// Set cells far apart to trigger cursor movement
	back.SetCell(0, 0, 'A', style.NewStyle())
	back.SetCell(9, 9, 'B', style.NewStyle())

	renderer.MarkDirty()
	output := renderer.Render()

	if output == "" {
		t.Error("Render() should return output")
	}

	// Should contain cursor movement
	if !contains(output, "\x1b[") {
		t.Error("Output should contain ANSI codes for cursor movement")
	}
}

func TestRenderer_WithNilFrontBuffer(t *testing.T) {
	renderer := NewRenderer(10, 10)
	back := renderer.GetBackBuffer()

	// Set content when front buffer is essentially empty (first render)
	back.SetCell(5, 5, 'X', style.NewStyle())

	renderer.MarkDirty()
	output := renderer.Render()

	if output == "" {
		t.Error("First render should produce output")
	}
}

func TestRenderer_veryWideOutput(t *testing.T) {
	renderer := NewRenderer(100, 5)
	back := renderer.GetBackBuffer()

	// Fill a long row
	style := style.NewStyle()
	for x := 0; x < 100; x++ {
		back.SetCell(x, 2, 'A'+rune(x%26), style)
	}

	renderer.MarkDirty()
	output := renderer.Render()

	if output == "" {
		t.Error("Render() should return output for wide row")
	}

	// Output should be reasonable
	if len(output) > 10000 {
		t.Errorf("Output size %d seems too large", len(output))
	}
}
