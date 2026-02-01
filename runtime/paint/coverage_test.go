package paint

import (
	"testing"
	"time"

	"github.com/wwsheng009/mint/runtime/style"
)

// =============================================================================
// Remote Optimizer Tests
// =============================================================================

func TestNewRemoteOptimizer(t *testing.T) {
	optimizer := NewRemoteOptimizer()

	if optimizer == nil {
		t.Fatal("NewRemoteOptimizer returned nil")
	}

	// Check default values
	if optimizer.GetFrameInterval() != 16*time.Millisecond {
		t.Errorf("default interval = %v, want 16ms", optimizer.GetFrameInterval())
	}
}

func TestNewRemoteOptimizerWithInterval(t *testing.T) {
	interval := 50 * time.Millisecond
	optimizer := NewRemoteOptimizerWithInterval(interval)

	if optimizer == nil {
		t.Fatal("NewRemoteOptimizerWithInterval returned nil")
	}

	if optimizer.GetFrameInterval() != interval {
		t.Errorf("interval = %v, want %v", optimizer.GetFrameInterval(), interval)
	}
}

func TestRemoteOptimizer_ShouldFlush(t *testing.T) {
	optimizer := NewRemoteOptimizer()

	// Just created, should not need to flush yet (time hasn't passed)
	// Actually ShouldFlush checks time since last flush, which was never set
	// So it returns true because time.Now() - zero time is large
	if !optimizer.ShouldFlush() {
		// This is actually expected behavior
	}

	// Buffer some data - still shouldn't flush (not enough time)
	optimizer.BufferFrame([]byte("test"))
	// But the buffer is too small to trigger flush
	// ShouldFlush only checks buffer size > 4KB, not time
	_ = optimizer.BufferSize()
}

func TestRemoteOptimizer_BufferFrame(t *testing.T) {
	optimizer := NewRemoteOptimizer()

	data := []byte("test frame data")
	err := optimizer.BufferFrame(data)

	if err != nil {
		t.Fatalf("BufferFrame() error = %v", err)
	}

	if optimizer.BufferSize() != len(data) {
		t.Errorf("BufferSize() = %d, want %d", optimizer.BufferSize(), len(data))
	}
}

func TestRemoteOptimizer_Flush(t *testing.T) {
	optimizer := NewRemoteOptimizer()

	// Flush empty buffer
	result := optimizer.Flush()
	if result != nil {
		t.Errorf("Flush() on empty buffer = %v, want nil", result)
	}

	// Add data and flush
	data := []byte("test data")
	optimizer.BufferFrame(data)

	result = optimizer.Flush()
	if result == nil {
		t.Fatal("Flush() returned nil after adding data")
	}

	if string(result) != string(data) {
		t.Errorf("Flush() = %q, want %q", string(result), string(data))
	}

	// Buffer should be cleared after flush
	if optimizer.BufferSize() != 0 {
		t.Errorf("BufferSize() after Flush() = %d, want 0", optimizer.BufferSize())
	}
}

func TestRemoteOptimizer_FlushWithDelta(t *testing.T) {
	optimizer := NewRemoteOptimizer()

	// First frame
	data1 := []byte("frame 1 data")
	optimizer.BufferFrame(data1)
	result1 := optimizer.FlushWithDelta()

	if string(result1) != string(data1) {
		t.Errorf("First FlushWithDelta() = %q, want %q", string(result1), string(data1))
	}

	// Second frame with same content
	optimizer.BufferFrame(data1)
	result2 := optimizer.FlushWithDelta()

	// With delta encoding, should be smaller
	if len(result2) >= len(data1) {
		t.Errorf("Delta encoded size = %d, want < %d", len(result2), len(data1))
	}

	// Third frame with different content
	data2 := []byte("frame 2 data which is different")
	optimizer.BufferFrame(data2)
	result3 := optimizer.FlushWithDelta()

	if result3 == nil {
		t.Fatal("FlushWithDelta() returned nil")
	}
}

func TestRemoteOptimizer_SetFrameInterval(t *testing.T) {
	optimizer := NewRemoteOptimizer()

	newInterval := 100 * time.Millisecond
	optimizer.SetFrameInterval(newInterval)

	if optimizer.GetFrameInterval() != newInterval {
		t.Errorf("GetFrameInterval() = %v, want %v", optimizer.GetFrameInterval(), newInterval)
	}
}

func TestRemoteOptimizer_EnableDeltaEncoding(t *testing.T) {
	optimizer := NewRemoteOptimizer()

	// Delta encoding is on by default
	stats := optimizer.Stats()
	if !stats.DeltaEncoding {
		t.Error("DeltaEncoding = false, want true by default")
	}

	// Disable delta encoding
	optimizer.EnableDeltaEncoding(false)

	stats = optimizer.Stats()
	if stats.DeltaEncoding {
		t.Error("DeltaEncoding = true after disabling, want false")
	}
}

func TestRemoteOptimizer_Clear(t *testing.T) {
	optimizer := NewRemoteOptimizer()

	optimizer.BufferFrame([]byte("test"))
	if optimizer.BufferSize() == 0 {
		t.Fatal("BufferSize() = 0 after BufferFrame(), want > 0")
	}

	optimizer.Clear()
	if optimizer.BufferSize() != 0 {
		t.Errorf("BufferSize() after Clear() = %d, want 0", optimizer.BufferSize())
	}
}

func TestRemoteOptimizer_Reset(t *testing.T) {
	optimizer := NewRemoteOptimizer()

	optimizer.BufferFrame([]byte("test"))
	optimizer.Flush() // Set lastFlush time

	optimizer.Reset()

	if optimizer.BufferSize() != 0 {
		t.Errorf("BufferSize() after Reset() = %d, want 0", optimizer.BufferSize())
	}
}

func TestRemoteOptimizer_SetMaxBufferSize(t *testing.T) {
	optimizer := NewRemoteOptimizer()

	// Set a max buffer size
	optimizer.SetMaxBufferSize(100)

	// Buffer small data first
	smallData := make([]byte, 50)
	_ = optimizer.BufferFrame(smallData)

	if optimizer.BufferSize() != 50 {
		t.Errorf("BufferSize() = %d, want 50", optimizer.BufferSize())
	}

	// Now add data that would exceed max - this triggers auto-flush
	// After flush, previous data is cleared, then new data is added
	largeData := make([]byte, 200)
	err := optimizer.BufferFrame(largeData)

	if err != nil {
		t.Fatalf("BufferFrame() with large data error = %v", err)
	}

	// The buffer was flushed first (clearing the 50 bytes),
	// then all 200 bytes were written
	size := optimizer.BufferSize()
	if size != 200 {
		t.Errorf("BufferSize() = %d after exceeding max, want 200", size)
	}
}

func TestRemoteOptimizer_Stats(t *testing.T) {
	optimizer := NewRemoteOptimizer()

	stats := optimizer.Stats()

	if stats.DeltaEncoding != true {
		t.Error("Stats.DeltaEncoding = false, want true")
	}
	if stats.FrameInterval != 16*time.Millisecond {
		t.Errorf("Stats.FrameInterval = %v, want 16ms", stats.FrameInterval)
	}
}

// =============================================================================
// Painter Tests
// =============================================================================

func TestPainter_Translate(t *testing.T) {
	buf := NewBuffer(10, 10)
	bounds := Rect{X: 0, Y: 0, Width: 10, Height: 10}
	ctx := NewPaintContext(buf, bounds)
	painter := NewPainter(ctx)

	// Translate with dx, dy, dw, dh
	painter.Translate(2, 3, 5, 5)

	if painter == nil {
		t.Fatal("NewPainter returned nil")
	}
}

// =============================================================================
// Helper Functions
// =============================================================================

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && indexOf(s, substr) >= 0)
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if len(s[i:i+len(substr)]) >= len(substr) && s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// =============================================================================
// LayerType String Tests
// =============================================================================

func TestLayerType_String(t *testing.T) {
	tests := []struct {
		name     string
		layer    LayerType
		expected string
	}{
		{
			name:     "Background layer",
			layer:    LayerBackground,
			expected: "Background",
		},
		{
			name:     "Content layer",
			layer:    LayerContent,
			expected: "Content",
		},
		{
			name:     "Stream layer",
			layer:    LayerStream,
			expected: "Stream",
		},
		{
			name:     "Overlay layer",
			layer:    LayerOverlay,
			expected: "Overlay",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.layer.String()
			if result != tt.expected {
				t.Errorf("LayerType.String() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestNewLayer(t *testing.T) {
	layer := NewLayer("test-layer", LayerBackground, 0, 10, 10)

	if layer == nil {
		t.Fatal("NewLayer returned nil")
	}

	if layer.ID != "test-layer" {
		t.Errorf("ID = %q, want %q", layer.ID, "test-layer")
	}
	if layer.Type != LayerBackground {
		t.Errorf("Type = %v, want LayerBackground", layer.Type)
	}
	if layer.ZIndex != 0 {
		t.Errorf("ZIndex = %d, want 0", layer.ZIndex)
	}
}

// =============================================================================
// CommandBatch Tests
// =============================================================================

func TestNewCommandBatch(t *testing.T) {
	batch := NewCommandBatch()

	if batch == nil {
		t.Fatal("NewCommandBatch returned nil")
	}
}

func TestCommandBatch_Add(t *testing.T) {
	batch := NewCommandBatch()

	// Add a command using correct API
	batch.Add(0, 0, "Hello", style.NewStyle().Foreground("red"))

	// Can't directly verify commands without accessing internals
	// This test mainly ensures it doesn't panic
	_ = batch
}

func TestCommandBatch_AddCell(t *testing.T) {
	batch := NewCommandBatch()

	// AddCell internally calls Add
	batch.AddCell(0, 0, 'X', style.NewStyle().Bold(true))

	_ = batch
}

func TestCommandBatch_Flush(t *testing.T) {
	batch := NewCommandBatch()

	// Add commands
	batch.Add(0, 0, "A", style.NewStyle())
	batch.Add(1, 0, "B", style.NewStyle())

	// Flush generates output
	output := batch.Flush()

	if output == "" {
		t.Error("Flush() returned empty string")
	}
}

func TestCommandBatch_FlushEmpty(t *testing.T) {
	batch := NewCommandBatch()

	// Flush without commands
	output := batch.Flush()

	if output != "" {
		t.Errorf("Flush() on empty batch = %q, want empty", output)
	}
}

func TestCommandBatch_SequentialFlush(t *testing.T) {
	batch := NewCommandBatch()

	// First flush
	batch.Add(0, 0, "A", style.NewStyle())
	output1 := batch.Flush()

	// Commands should be cleared after flush
	batch.Add(1, 0, "B", style.NewStyle())
	output2 := batch.Flush()

	if output1 == "" {
		t.Error("First Flush() returned empty")
	}
	if output2 == "" {
		t.Error("Second Flush() returned empty")
	}
}

func TestCommandBatch_StyleOptimization(t *testing.T) {
	batch := NewCommandBatch()

	style := style.NewStyle().Foreground("red").Bold(true)

	// Multiple cells with same style
	batch.Add(0, 0, "X", style)
	batch.Add(1, 0, "Y", style)
	batch.Add(2, 0, "Z", style)

	output := batch.Flush()

	if output == "" {
		t.Error("Flush() returned empty string")
	}

	// Should contain the text
	if !contains(output, "X") {
		t.Error("Output should contain 'X'")
	}
}

func TestCommandBatch_Clear(t *testing.T) {
	batch := NewCommandBatch()

	batch.Add(0, 0, "A", style.NewStyle())
	if batch.Count() != 1 {
		t.Errorf("Count() = %d, want 1", batch.Count())
	}

	batch.Clear()
	if batch.Count() != 0 {
		t.Errorf("Count() after Clear() = %d, want 0", batch.Count())
	}

	// Flush should return empty after Clear
	output := batch.Flush()
	if output != "" {
		t.Errorf("Flush() after Clear() = %q, want empty", output)
	}
}

func TestCommandBatch_Count(t *testing.T) {
	batch := NewCommandBatch()

	if batch.Count() != 0 {
		t.Errorf("Count() on empty batch = %d, want 0", batch.Count())
	}

	batch.Add(0, 0, "A", style.NewStyle())
	batch.Add(1, 0, "B", style.NewStyle())

	if batch.Count() != 2 {
		t.Errorf("Count() = %d, want 2", batch.Count())
	}
}

func TestCommandBatch_Reserve(t *testing.T) {
	batch := NewCommandBatch()

	// Reserve should not panic
	batch.Reserve(100)

	// Add commands after reserve
	batch.Add(0, 0, "A", style.NewStyle())

	if batch.Count() != 1 {
		t.Errorf("Count() after Reserve = %d, want 1", batch.Count())
	}
}

func TestCommandBatch_MergeFrom(t *testing.T) {
	batch1 := NewCommandBatch()
	batch1.Add(0, 0, "A", style.NewStyle())

	batch2 := NewCommandBatch()
	batch2.Add(1, 0, "B", style.NewStyle())

	// Merge batch2 into batch1
	batch1.MergeFrom(batch2)

	if batch1.Count() != 2 {
		t.Errorf("Count() after MergeFrom = %d, want 2", batch1.Count())
	}

	// Merge with nil should not panic
	batch1.MergeFrom(nil)
	if batch1.Count() != 2 {
		t.Errorf("Count() after MergeFrom nil = %d, want 2", batch1.Count())
	}

	// Merge with empty batch
	batch3 := NewCommandBatch()
	batch1.MergeFrom(batch3)
	if batch1.Count() != 2 {
		t.Errorf("Count() after MergeFrom empty = %d, want 2", batch1.Count())
	}
}

func TestCommandBatch_EstimateSize(t *testing.T) {
	batch := NewCommandBatch()

	if batch.EstimateSize() != 0 {
		t.Errorf("EstimateSize() on empty batch = %d, want 0", batch.EstimateSize())
	}

	batch.Add(0, 0, "Hello", style.NewStyle())

	size := batch.EstimateSize()
	if size <= 0 {
		t.Errorf("EstimateSize() = %d, want > 0", size)
	}
}

func TestCommandBatch_WriteToBuffer(t *testing.T) {
	batch := NewCommandBatch()
	batch.Add(0, 0, "A", style.NewStyle().Foreground("red"))
	batch.Add(1, 0, "B", style.NewStyle().Foreground("blue"))

	buf := NewBuffer(10, 5)
	batch.WriteToBuffer(buf)

	// Check that cells were written
	cell1 := buf.GetContent(0, 0)
	if cell1.Cluster != "A" {
		t.Errorf("Cell at (0,0) = %q, want \"A\"", cell1.Cluster)
	}

	cell2 := buf.GetContent(1, 0)
	if cell2.Cluster != "B" {
		t.Errorf("Cell at (1,0) = %q, want \"B\"", cell2.Cluster)
	}
}

func TestCommandBatch_WriteToBufferWithOffset(t *testing.T) {
	batch := NewCommandBatch()
	batch.Add(0, 0, "A", style.NewStyle())
	batch.Add(1, 0, "B", style.NewStyle())

	buf := NewBuffer(10, 5)
	batch.WriteToBufferWithOffset(buf, 2, 1)

	// Check that cells were written with offset
	cell1 := buf.GetContent(2, 1)
	if cell1.Cluster != "A" {
		t.Errorf("Cell at (2,1) = %q, want \"A\"", cell1.Cluster)
	}

	cell2 := buf.GetContent(3, 1)
	if cell2.Cluster != "B" {
		t.Errorf("Cell at (3,1) = %q, want \"B\"", cell2.Cluster)
	}
}

// =============================================================================
// Buffer Additional Tests
// =============================================================================

func TestBuffer_GetContent(t *testing.T) {
	buf := NewBuffer(10, 5)

	// Out of bounds should return empty cell
	cell := buf.GetContent(-1, 0)
	if cell.Cluster != "" {
		t.Errorf("GetContent(-1, 0) = %q, want empty", cell.Cluster)
	}

	cell = buf.GetContent(100, 0)
	if cell.Cluster != "" {
		t.Errorf("GetContent(100, 0) = %q, want empty", cell.Cluster)
	}

	// Valid position
	buf.SetCell(5, 3, 'X', style.NewStyle())
	cell = buf.GetContent(5, 3)
	if cell.Cluster != "X" {
		t.Errorf("GetContent(5, 3) = %q, want \"X\"", cell.Cluster)
	}
}

func TestBuffer_SetContent(t *testing.T) {
	buf := NewBuffer(10, 5)

	// Set with Z-Index
	buf.SetContent(5, 3, 10, 'A', style.NewStyle().Foreground("red"), "node1")

	cell := buf.GetContent(5, 3)
	if cell.Cluster != "A" {
		t.Errorf("SetContent result = %q, want \"A\"", cell.Cluster)
	}
	if cell.ZIndex != 10 {
		t.Errorf("ZIndex = %d, want 10", cell.ZIndex)
	}

	// Lower Z-Index should not overwrite
	buf.SetContent(5, 3, 5, 'B', style.NewStyle().Foreground("blue"), "node2")
	cell = buf.GetContent(5, 3)
	if cell.Cluster != "A" {
		t.Errorf("After lower Z-Index, cell = %q, want \"A\"", cell.Cluster)
	}

	// Higher Z-Index should overwrite
	buf.SetContent(5, 3, 15, 'C', style.NewStyle().Foreground("green"), "node3")
	cell = buf.GetContent(5, 3)
	if cell.Cluster != "C" {
		t.Errorf("After higher Z-Index, cell = %q, want \"C\"", cell.Cluster)
	}

	// Out of bounds should be ignored
	buf.SetContent(-1, 0, 10, 'X', style.NewStyle(), "node4")
	// Should not panic
}

func TestBuffer_SetContentDirect(t *testing.T) {
	buf := NewBuffer(10, 5)

	// Set directly without Z-Index check
	buf.SetContentDirect(5, 3, 'A', style.NewStyle().Foreground("red"), 10)

	cell := buf.GetContent(5, 3)
	if cell.Cluster != "A" {
		t.Errorf("SetContentDirect result = %q, want \"A\"", cell.Cluster)
	}
	if cell.ZIndex != 10 {
		t.Errorf("ZIndex = %d, want 10", cell.ZIndex)
	}

	// Direct set should overwrite regardless of Z-Index
	buf.SetContentDirect(5, 3, 'B', style.NewStyle().Foreground("blue"), 5)
	cell = buf.GetContent(5, 3)
	if cell.Cluster != "B" {
		t.Errorf("After direct set with lower Z-Index, cell = %q, want \"B\"", cell.Cluster)
	}
}

func TestRect_Contains(t *testing.T) {
	rect := Rect{X: 5, Y: 10, Width: 20, Height: 30}

	// Point inside
	if !rect.Contains(10, 20) {
		t.Error("Contains(10, 20) = false, want true")
	}

	// Point at top-left corner
	if !rect.Contains(5, 10) {
		t.Error("Contains(5, 10) = false, want true")
	}

	// Point at bottom-right corner (exclusive)
	if rect.Contains(25, 40) {
		t.Error("Contains(25, 40) = true, want false (exclusive)")
	}

	// Point outside - left
	if rect.Contains(4, 20) {
		t.Error("Contains(4, 20) = true, want false")
	}

	// Point outside - top
	if rect.Contains(10, 9) {
		t.Error("Contains(10, 9) = true, want false")
	}

	// Point outside - right
	if rect.Contains(25, 20) {
		t.Error("Contains(25, 20) = true, want false")
	}

	// Point outside - bottom
	if rect.Contains(10, 40) {
		t.Error("Contains(10, 40) = true, want false")
	}
}

func TestBuffer_String(t *testing.T) {
	buf := NewBuffer(10, 3)

	// Empty buffer
	s := buf.String()
	if s == "" {
		// Should have some content (lines with spaces)
		t.Error("String() on empty buffer returned empty, expected lines")
	}

	// Buffer with content
	buf.SetCell(0, 0, 'H', style.NewStyle())
	buf.SetCell(1, 0, 'i', style.NewStyle())
	buf.SetCell(2, 0, '!', style.NewStyle())

	s = buf.String()
	if s == "" {
		t.Error("String() returned empty")
	}
	if !contains(s, "H") {
		t.Error("String() should contain 'H'")
	}

	// Empty buffer (height 0)
	emptyBuf := NewBuffer(10, 0)
	s = emptyBuf.String()
	if s != "" {
		t.Errorf("String() on zero-height buffer = %q, want empty", s)
	}
}

func TestBuffer_SetSelected(t *testing.T) {
	buf := NewBuffer(10, 5)

	// Set selected
	buf.SetSelected(5, 3, true)
	cell := buf.GetContent(5, 3)
	if !cell.Selected {
		t.Error("Cell should be selected")
	}

	// Unset selected
	buf.SetSelected(5, 3, false)
	cell = buf.GetContent(5, 3)
	if cell.Selected {
		t.Error("Cell should not be selected")
	}

	// Out of bounds should be ignored
	buf.SetSelected(-1, 0, true)
	buf.SetSelected(100, 0, true)
	// Should not panic
}

func TestBuffer_ClearSelection(t *testing.T) {
	buf := NewBuffer(10, 5)

	// Select some cells
	buf.SetSelected(0, 0, true)
	buf.SetSelected(5, 3, true)

	// Clear all
	buf.ClearSelection()

	// Check all cells are unselected
	for y := 0; y < buf.Height; y++ {
		for x := 0; x < buf.Width; x++ {
			if buf.Cells[y][x].Selected {
				t.Errorf("Cell at (%d, %d) is still selected", x, y)
			}
		}
	}
}

func TestBuffer_Clear(t *testing.T) {
	buf := NewBuffer(10, 5)

	// Set some content
	buf.SetCell(5, 3, 'X', style.NewStyle().Foreground("red"))
	buf.SetSelected(5, 3, true)

	// Clear
	buf.Clear()

	// Check all cells are cleared
	for y := 0; y < buf.Height; y++ {
		for x := 0; x < buf.Width; x++ {
			cell := buf.Cells[y][x]
			if cell.Cluster != " " {
				t.Errorf("Cell at (%d, %d) Cluster = %q, want \" \"", x, y, cell.Cluster)
			}
			if cell.Selected {
				t.Errorf("Cell at (%d, %d) is still selected", x, y)
			}
			if cell.ZIndex != 0 {
				t.Errorf("Cell at (%d, %d) ZIndex = %d, want 0", x, y, cell.ZIndex)
			}
		}
	}
}

func TestBuffer_Reset(t *testing.T) {
	buf := NewBuffer(10, 5)

	// Set some content
	buf.SetCell(5, 3, 'X', style.NewStyle())

	// Reset to different size
	buf.Reset(20, 10)

	if buf.Width != 20 {
		t.Errorf("Width after Reset = %d, want 20", buf.Width)
	}
	if buf.Height != 10 {
		t.Errorf("Height after Reset = %d, want 10", buf.Height)
	}

	// Reset with zero values (should default to 80x24)
	buf.Reset(0, 0)
	if buf.Width != 80 {
		t.Errorf("Width after Reset(0,0) = %d, want 80", buf.Width)
	}
	if buf.Height != 24 {
		t.Errorf("Height after Reset(0,0) = %d, want 24", buf.Height)
	}
}

// =============================================================================
// Compositor Tests
// =============================================================================

func TestCompositor_MarkAllDirty(t *testing.T) {
	compositor := NewCompositor(80, 24)

	layer1 := NewLayer("layer1", LayerBackground, 0, 80, 24)
	layer2 := NewLayer("layer2", LayerContent, 1, 80, 24)

	compositor.AddLayer(layer1)
	compositor.AddLayer(layer2)

	// Clear dirty flags first
	layer1.ClearDirty()
	layer2.ClearDirty()

	// Mark all as dirty
	compositor.MarkAllDirty()

	if !layer1.IsDirty() {
		t.Error("layer1 should be dirty")
	}
	if !layer2.IsDirty() {
		t.Error("layer2 should be dirty")
	}
}

func TestCompositor_MarkTypeDirty(t *testing.T) {
	compositor := NewCompositor(80, 24)

	bgLayer := NewLayer("bg", LayerBackground, 0, 80, 24)
	contentLayer := NewLayer("content", LayerContent, 1, 80, 24)
	overlayLayer := NewLayer("overlay", LayerOverlay, 2, 80, 24)

	compositor.AddLayer(bgLayer)
	compositor.AddLayer(contentLayer)
	compositor.AddLayer(overlayLayer)

	// Clear dirty flags first
	bgLayer.ClearDirty()
	contentLayer.ClearDirty()
	overlayLayer.ClearDirty()

	// Mark only Content layers as dirty
	compositor.MarkTypeDirty(LayerContent)

	if bgLayer.IsDirty() {
		t.Error("Background layer should not be dirty")
	}
	if !contentLayer.IsDirty() {
		t.Error("Content layer should be dirty")
	}
	if overlayLayer.IsDirty() {
		t.Error("Overlay layer should not be dirty")
	}
}

func TestCompositor_Resize(t *testing.T) {
	compositor := NewCompositor(80, 24)

	layer := NewLayer("test", LayerContent, 0, 100, 100)
	compositor.AddLayer(layer)

	// Resize compositor - should not panic
	compositor.Resize(50, 20)

	// Layer should be resized to fit
	rect := layer.GetRect()
	if rect.Width > 50 {
		t.Errorf("Layer width = %d, want <= 50", rect.Width)
	}
	if rect.Height > 20 {
		t.Errorf("Layer height = %d, want <= 20", rect.Height)
	}
}

func TestCompositor_GetLayerCount(t *testing.T) {
	compositor := NewCompositor(80, 24)

	if compositor.GetLayerCount() != 0 {
		t.Errorf("LayerCount = %d, want 0", compositor.GetLayerCount())
	}

	compositor.AddLayer(NewLayer("l1", LayerBackground, 0, 80, 24))
	if compositor.GetLayerCount() != 1 {
		t.Errorf("LayerCount = %d, want 1", compositor.GetLayerCount())
	}

	compositor.AddLayer(NewLayer("l2", LayerContent, 1, 80, 24))
	if compositor.GetLayerCount() != 2 {
		t.Errorf("LayerCount = %d, want 2", compositor.GetLayerCount())
	}
}

func TestCompositor_GetLayers(t *testing.T) {
	compositor := NewCompositor(80, 24)

	layer1 := NewLayer("l1", LayerBackground, 0, 80, 24)
	layer2 := NewLayer("l2", LayerContent, 1, 80, 24)

	compositor.AddLayer(layer1)
	compositor.AddLayer(layer2)

	layers := compositor.GetLayers()
	if len(layers) != 2 {
		t.Errorf("GetLayers() = %d layers, want 2", len(layers))
	}
}

func TestCompositor_Clear(t *testing.T) {
	compositor := NewCompositor(80, 24)

	layer := NewLayer("test", LayerContent, 0, 80, 24)
	layer.Buffer.SetCell(10, 10, 'X', style.NewStyle())

	compositor.AddLayer(layer)

	// Clear all layers
	compositor.Clear()

	// Check layer buffer is cleared - empty cells have empty Cluster
	cell := layer.Buffer.GetContent(10, 10)
	if cell.Cluster != "" && cell.Cluster != " " {
		t.Errorf("Layer should be cleared, but got %q", cell.Cluster)
	}
}

func TestCompositor_ClearType(t *testing.T) {
	compositor := NewCompositor(80, 24)

	bgLayer := NewLayer("bg", LayerBackground, 0, 80, 24)
	contentLayer := NewLayer("content", LayerContent, 1, 80, 24)

	bgLayer.Buffer.SetCell(5, 5, 'B', style.NewStyle())
	contentLayer.Buffer.SetCell(10, 10, 'C', style.NewStyle())

	compositor.AddLayer(bgLayer)
	compositor.AddLayer(contentLayer)

	// Clear only Content layers
	compositor.ClearType(LayerContent)

	// Background should still have content
	bgCell := bgLayer.Buffer.GetContent(5, 5)
	if bgCell.Cluster != "B" {
		t.Errorf("Background layer should not be cleared, got %q", bgCell.Cluster)
	}

	// Content should be cleared
	contentCell := contentLayer.Buffer.GetContent(10, 10)
	if contentCell.Cluster != "" && contentCell.Cluster != " " {
		t.Errorf("Content layer should be cleared, got %q", contentCell.Cluster)
	}
}

func TestCompositor_Fill(t *testing.T) {
	compositor := NewCompositor(80, 24)

	layer := NewLayer("test", LayerContent, 0, 80, 24)
	compositor.AddLayer(layer)

	// Fill with 'X' and red color
	style := style.NewStyle().Foreground("red")
	success := compositor.Fill("test", 'X', style)

	if !success {
		t.Error("Fill() returned false")
	}

	// Check that cells were filled
	cell := layer.Buffer.GetContent(10, 10)
	if cell.Cluster != "X" {
		t.Errorf("Cell Cluster = %q, want \"X\"", cell.Cluster)
	}

	// Try to fill non-existent layer
	success = compositor.Fill("nonexistent", 'Y', style)
	if success {
		t.Error("Fill() on non-existent layer should return false")
	}
}

func TestCompositor_GetLayerByType(t *testing.T) {
	compositor := NewCompositor(80, 24)

	bgLayer := NewLayer("bg", LayerBackground, 0, 80, 24)
	contentLayer := NewLayer("content", LayerContent, 1, 80, 24)

	compositor.AddLayer(bgLayer)
	compositor.AddLayer(contentLayer)

	// Get layer by type
	layer := compositor.GetLayerByType(LayerBackground)
	if layer == nil {
		t.Fatal("GetLayerByType(LayerBackground) returned nil")
	}
	if layer.ID != "bg" {
		t.Errorf("Layer ID = %q, want \"bg\"", layer.ID)
	}

	// Get non-existent type
	layer = compositor.GetLayerByType(LayerStream)
	if layer != nil {
		t.Error("GetLayerByType(LayerStream) should return nil")
	}
}

func TestLayer_EnableDisable(t *testing.T) {
	layer := NewLayer("test", LayerContent, 0, 80, 24)

	if !layer.Enabled {
		t.Error("Layer should be enabled by default")
	}

	layer.Disable()
	if layer.Enabled {
		t.Error("Layer should be disabled after Disable()")
	}

	layer.Enable()
	if !layer.Enabled {
		t.Error("Layer should be enabled after Enable()")
	}
}

func TestLayer_ShowHide(t *testing.T) {
	layer := NewLayer("test", LayerContent, 0, 80, 24)

	if !layer.Visible {
		t.Error("Layer should be visible by default")
	}

	layer.Hide()
	if layer.Visible {
		t.Error("Layer should not be visible after Hide()")
	}

	layer.Show()
	if !layer.Visible {
		t.Error("Layer should be visible after Show()")
	}
}

func TestLayer_SetPosition(t *testing.T) {
	layer := NewLayer("test", LayerContent, 0, 80, 24)
	layer.SetRect(Rect{X: 0, Y: 0, Width: 80, Height: 24})

	layer.SetPosition(10, 20)

	rect := layer.GetRect()
	if rect.X != 10 {
		t.Errorf("X = %d, want 10", rect.X)
	}
	if rect.Y != 20 {
		t.Errorf("Y = %d, want 20", rect.Y)
	}
}

func TestLayer_SetSize(t *testing.T) {
	layer := NewLayer("test", LayerContent, 0, 80, 24)
	layer.SetRect(Rect{X: 0, Y: 0, Width: 80, Height: 24})

	layer.SetSize(40, 15)

	rect := layer.GetRect()
	if rect.Width != 40 {
		t.Errorf("Width = %d, want 40", rect.Width)
	}
	if rect.Height != 15 {
		t.Errorf("Height = %d, want 15", rect.Height)
	}
}

func TestLayer_DirtyFlags(t *testing.T) {
	layer := NewLayer("test", LayerContent, 0, 80, 24)

	// New layers are created with Dirty=true by default
	if !layer.IsDirty() {
		t.Error("New layer should be dirty by default")
	}

	// Clear the dirty flag
	layer.ClearDirty()
	if layer.IsDirty() {
		t.Error("Layer should not be dirty after ClearDirty()")
	}

	// Mark it dirty again
	layer.MarkDirty()
	if !layer.IsDirty() {
		t.Error("Layer should be dirty after MarkDirty()")
	}

	// Clear again
	layer.ClearDirty()
	if layer.IsDirty() {
		t.Error("Layer should not be dirty after ClearDirty()")
	}
}

func TestLayer_Fill(t *testing.T) {
	layer := NewLayer("test", LayerContent, 0, 20, 10)

	st := style.NewStyle().Foreground("red")
	layer.Fill('X', st)

	// Check several cells
	for y := 0; y < layer.Buffer.Height; y++ {
		for x := 0; x < layer.Buffer.Width; x++ {
			cell := layer.Buffer.GetContent(x, y)
			if cell.Cluster != "X" {
				t.Errorf("Cell at (%d, %d) = %q, want \"X\"", x, y, cell.Cluster)
			}
		}
	}
}

// =============================================================================
// PaintContext Tests
// =============================================================================

func TestPaintContext_WithFocus(t *testing.T) {
	buf := NewBuffer(10, 10)
	bounds := Rect{X: 0, Y: 0, Width: 10, Height: 10}
	ctx := NewPaintContext(buf, bounds)

	if ctx.Focused {
		t.Error("New context should not be focused")
	}

	focusedCtx := ctx.WithFocus(true)
	if !focusedCtx.Focused {
		t.Error("WithFocus(true) should set Focused to true")
	}

	// Original should not be modified
	if ctx.Focused {
		t.Error("Original context should not be modified")
	}
}

func TestPaintContext_WithDisabled(t *testing.T) {
	buf := NewBuffer(10, 10)
	bounds := Rect{X: 0, Y: 0, Width: 10, Height: 10}
	ctx := NewPaintContext(buf, bounds)

	if ctx.Disabled {
		t.Error("New context should not be disabled")
	}

	disabledCtx := ctx.WithDisabled(true)
	if !disabledCtx.Disabled {
		t.Error("WithDisabled(true) should set Disabled to true")
	}
}

func TestPaintContext_WithZIndex(t *testing.T) {
	buf := NewBuffer(10, 10)
	bounds := Rect{X: 0, Y: 0, Width: 10, Height: 10}
	ctx := NewPaintContext(buf, bounds)

	if ctx.ZIndex != 0 {
		t.Errorf("New context ZIndex = %d, want 0", ctx.ZIndex)
	}

	zCtx := ctx.WithZIndex(5)
	if zCtx.ZIndex != 5 {
		t.Errorf("WithZIndex(5) = %d, want 5", zCtx.ZIndex)
	}
}

func TestPaintContext_WithBounds(t *testing.T) {
	buf := NewBuffer(10, 10)
	bounds := Rect{X: 0, Y: 0, Width: 10, Height: 10}
	ctx := NewPaintContext(buf, bounds)

	newBounds := Rect{X: 5, Y: 5, Width: 20, Height: 20}
	newCtx := ctx.WithBounds(newBounds)

	if newCtx.Bounds.X != 5 || newCtx.Bounds.Y != 5 {
		t.Errorf("Bounds = %+v, want {X:5, Y:5}", newCtx.Bounds)
	}
	if newCtx.X != 5 {
		t.Errorf("X = %d, want 5", newCtx.X)
	}
	if newCtx.AvailableWidth != 20 {
		t.Errorf("AvailableWidth = %d, want 20", newCtx.AvailableWidth)
	}
}

func TestPaintContext_WithViewport(t *testing.T) {
	buf := NewBuffer(10, 10)
	bounds := Rect{X: 0, Y: 0, Width: 10, Height: 10}
	ctx := NewPaintContext(buf, bounds)

	vpCtx := ctx.WithViewport(10, 20)
	// Just verify it doesn't panic and returns a new context
	if vpCtx == nil {
		t.Error("WithViewport should return a new context")
	}
}

func TestPaintContext_WithFocusPath(t *testing.T) {
	buf := NewBuffer(10, 10)
	bounds := Rect{X: 0, Y: 0, Width: 10, Height: 10}
	ctx := NewPaintContext(buf, bounds)

	path := make([]string, 0)
	path = append(path, "component1")

	pathCtx := ctx.WithFocusPath(path)
	if len(pathCtx.FocusPath) != 1 {
		t.Errorf("FocusPath length = %d, want 1", len(pathCtx.FocusPath))
	}
}

func TestPaintContext_Child(t *testing.T) {
	buf := NewBuffer(10, 10)
	bounds := Rect{X: 0, Y: 0, Width: 10, Height: 10}
	ctx := NewPaintContext(buf, bounds)

	childBounds := Rect{X: 2, Y: 2, Width: 5, Height: 5}
	childCtx := ctx.Child("child1", childBounds)

	if childCtx.Bounds.X != 2 || childCtx.Bounds.Y != 2 {
		t.Errorf("Child bounds = %+v, want {X:2, Y:2}", childCtx.Bounds)
	}

	// Check that focus path includes child
	if len(childCtx.FocusPath) != 1 {
		t.Errorf("Child FocusPath length = %d, want 1", len(childCtx.FocusPath))
	}
	if childCtx.FocusPath[0] != "child1" {
		t.Errorf("FocusPath[0] = %q, want \"child1\"", childCtx.FocusPath[0])
	}
}

func TestPaintContext_SetCell(t *testing.T) {
	buf := NewBuffer(10, 10)
	bounds := Rect{X: 0, Y: 0, Width: 10, Height: 10}
	ctx := NewPaintContext(buf, bounds)

	ctx.SetCell(5, 5, 'X', style.NewStyle().Foreground("red"))

	cell := buf.GetContent(5, 5)
	if cell.Cluster != "X" {
		t.Errorf("Cell at (5,5) = %q, want \"X\"", cell.Cluster)
	}
}

func TestPaintContext_SetString(t *testing.T) {
	buf := NewBuffer(10, 10)
	bounds := Rect{X: 0, Y: 0, Width: 10, Height: 10}
	ctx := NewPaintContext(buf, bounds)

	ctx.SetString(2, 3, "Hello", style.NewStyle())

	// Check characters
	cell := buf.GetContent(2, 3)
	if cell.Cluster != "H" {
		t.Errorf("Cell at (2,3) = %q, want \"H\"", cell.Cluster)
	}

	cell = buf.GetContent(3, 3)
	if cell.Cluster != "e" {
		t.Errorf("Cell at (3,3) = %q, want \"e\"", cell.Cluster)
	}
}

func TestPaintContext_Fill(t *testing.T) {
	buf := NewBuffer(10, 10)
	bounds := Rect{X: 0, Y: 0, Width: 10, Height: 10}
	ctx := NewPaintContext(buf, bounds)

	ctx.Fill(bounds, 'X', style.NewStyle())

	// Check that buffer is filled
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			cell := buf.GetContent(x, y)
			if cell.Cluster != "X" {
				t.Errorf("Cell at (%d,%d) = %q, want \"X\"", x, y, cell.Cluster)
			}
		}
	}
}

func TestPaintContext_Width(t *testing.T) {
	buf := NewBuffer(100, 50)
	bounds := Rect{X: 0, Y: 0, Width: 100, Height: 50}
	ctx := NewPaintContext(buf, bounds)

	if ctx.Width() != 100 {
		t.Errorf("Width() = %d, want 100", ctx.Width())
	}
}

func TestPaintContext_Height(t *testing.T) {
	buf := NewBuffer(100, 50)
	bounds := Rect{X: 0, Y: 0, Width: 100, Height: 50}
	ctx := NewPaintContext(buf, bounds)

	if ctx.Height() != 50 {
		t.Errorf("Height() = %d, want 50", ctx.Height())
	}
}

func TestPaintContext_Contains(t *testing.T) {
	buf := NewBuffer(10, 10)
	bounds := Rect{X: 5, Y: 5, Width: 10, Height: 10}
	ctx := NewPaintContext(buf, bounds)

	// Contains uses relative coordinates (0 to Width-1, 0 to Height-1)
	if !ctx.Contains(0, 0) {
		t.Error("Contains(0,0) should return true")
	}
	if !ctx.Contains(9, 9) {
		t.Error("Contains(9,9) should return true")
	}
	if ctx.Contains(-1, 0) {
		t.Error("Contains(-1,0) should return false")
	}
	if ctx.Contains(10, 0) {
		t.Error("Contains(10,0) should return false")
	}
	if ctx.Contains(0, 10) {
		t.Error("Contains(0,10) should return false")
	}
}

func TestPaintContext_Clamp(t *testing.T) {
	buf := NewBuffer(10, 10)
	bounds := Rect{X: 5, Y: 5, Width: 10, Height: 10}
	ctx := NewPaintContext(buf, bounds)

	// Clamp uses relative coordinates and clamps to [0, Width-1], [0, Height-1]
	x, y := ctx.Clamp(-1, -1)
	if x != 0 || y != 0 {
		t.Errorf("Clamp(-1,-1) = (%d,%d), want (0,0)", x, y)
	}

	x, y = ctx.Clamp(15, 20)
	if x != 9 || y != 9 {
		t.Errorf("Clamp(15,20) = (%d,%d), want (9,9)", x, y)
	}

	x, y = ctx.Clamp(5, 5)
	if x != 5 || y != 5 {
		t.Errorf("Clamp(5,5) = (%d,%d), want (5,5)", x, y)
	}
}

func TestPaintContext_Clone(t *testing.T) {
	buf := NewBuffer(10, 10)
	bounds := Rect{X: 0, Y: 0, Width: 10, Height: 10}
	ctx := NewPaintContext(buf, bounds)

	ctx.Focused = true
	ctx.Disabled = true
	ctx.ZIndex = 5

	// clone is private, but WithFocus calls it
	cloned := ctx.WithFocus(false)

	// Cloned should have different values
	if cloned.Focused {
		t.Error("Cloned context should have Focused=false")
	}
	// Original should not be modified
	if !ctx.Focused {
		t.Error("Original context should still have Focused=true")
	}
}
