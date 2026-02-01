// Package paint provides Run-Length Encoding (RLE) optimization for rendering.
package paint

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/wwsheng009/mint/runtime/style"
)

// =============================================================================
// Run-Length Encoding (RLE)
// =============================================================================

// Run represents a sequence of identical cells.
type Run struct {
	Cell  Cell      // The cell value (style + cluster)
	Count int       // Number of consecutive cells
	X     int       // Starting X position
	Y     int       // Y position
}

// EncodeRLE encodes a buffer row using run-length encoding.
// This reduces output size when consecutive cells have the same style.
func EncodeRLE(row []Cell, width int) []Run {
	if len(row) == 0 || width == 0 {
		return nil
	}

	var runs []Run
	current := row[0]
	count := 1

	for i := 1; i < width && i < len(row); i++ {
		// Skip continuation cells for wide characters
		if row[i].IsContinuation {
			continue
		}

		// Check if we have the same style and cluster
		if row[i].Style == current.Style && row[i].Cluster == current.Cluster {
			count++
		} else {
			runs = append(runs, Run{
				Cell:  current,
				Count: count,
				X:     i - count,
			})
			current = row[i]
			count = 1
		}
	}

	// Add the last run
	runs = append(runs, Run{
		Cell:  current,
		Count: count,
		X:     width - count,
	})

	return runs
}

// =============================================================================
// RLE Renderer
// =============================================================================

// RLERenderer renders using run-length encoding for optimization.
type RLERenderer struct {
	buffer bytes.Buffer
}

// NewRLERenderer creates a new RLE renderer.
func NewRLERenderer() *RLERenderer {
	return &RLERenderer{}
}

// RenderRow renders a single row using RLE.
func (r *RLERenderer) RenderRow(row []Cell, width int, y int) string {
	runs := EncodeRLE(row, width)
	if len(runs) == 0 {
		return ""
	}

	r.buffer.Reset()
	lastStyle := style.Style{}
	lastX := 0

	for _, run := range runs {
		// Skip continuation cells
		if run.Cell.IsContinuation {
			continue
		}

		// Move cursor if needed
		if run.X != lastX {
			r.buffer.WriteString(cursorMove(lastX, run.X, y))
		}

		// Update style if needed
		if run.Cell.Style != lastStyle {
			r.buffer.WriteString(styleToANSI(run.Cell.Style))
			lastStyle = run.Cell.Style
		}

		// Write the character(s) - repeat for the run length
		if run.Cell.Cluster != "" {
			for i := 0; i < run.Count; i++ {
				r.buffer.WriteString(run.Cell.Cluster)
			}
		} else {
			for i := 0; i < run.Count; i++ {
				r.buffer.WriteString(" ")
			}
		}

		lastX = run.X + run.Count
	}

	return r.buffer.String()
}

// =============================================================================
// Optimized Buffer Output
// =============================================================================

// OptimizedOutput generates optimized ANSI output for a buffer.
// It uses RLE and style minimization to reduce output size.
func OptimizedOutput(buf *Buffer, diff *DiffResult) string {
	if !diff.HasChanges {
		return ""
	}

	var output bytes.Buffer
	lastStyle := style.Style{}
	lastY := -1
	lastX := 0

	for _, region := range diff.DirtyRegions {
		for y := region.Y; y < region.Y+region.Height && y < buf.Height; y++ {
			if y < 0 {
				continue
			}
			for x := region.X; x < region.X+region.Width && x < buf.Width; x++ {
				if x < 0 {
					continue
				}

				cell := buf.Cells[y][x]
				if cell.IsContinuation {
					continue
				}

				// Move cursor if position changed
				if y != lastY || x != lastX+1 {
					output.WriteString(cursorMove(lastX, x, y))
					lastY = y
					lastX = x
				}

				// Update style if needed
				if cell.Style != lastStyle {
					output.WriteString(styleToANSI(cell.Style))
					lastStyle = cell.Style
				}

				// Write the character
				if cell.Cluster != "" {
					output.WriteString(cell.Cluster)
				} else {
					output.WriteString(" ")
				}

				lastX = x
			}
		}
	}

	// Reset style at the end
	output.WriteString("\x1b[0m")

	return output.String()
}

// =============================================================================
// Cursor Movement
// =============================================================================

// cursorMove generates ANSI escape codes to move the cursor.
// It optimizes for the shortest path using relative movement when beneficial.
func cursorMove(fromX, toX, y int) string {
	if fromX == toX {
		return "" // Already at the right position
	}

	// Calculate the optimal movement
	dx := toX - fromX

	var builder strings.Builder

	// Use cursor forward/backward for small movements
	if dx > 0 && dx < 10 {
		builder.WriteString("\x1b[")
		builder.WriteString(fmt.Sprintf("%d", dx))
		builder.WriteString("C")
	} else if dx < 0 && dx > -10 {
		builder.WriteString("\x1b[")
		builder.WriteString(fmt.Sprintf("%d", -dx))
		builder.WriteString("D")
	} else {
		// Use absolute positioning for larger movements
		builder.WriteString("\x1b[")
		builder.WriteString(fmt.Sprintf("%d;%dH", y+1, toX+1))
	}

	return builder.String()
}

// =============================================================================
// Style to ANSI Conversion
// =============================================================================

// styleToANSI converts a style to ANSI escape codes.
// This is optimized to only output changed attributes.
func styleToANSI(s style.Style) string {
	if s == (style.Style{}) {
		return "\x1b[0m"
	}

	var codes []string

	// Attributes
	if s.IsBold() {
		codes = append(codes, "1")
	}
	if s.IsItalic() {
		codes = append(codes, "3")
	}
	if s.IsUnderline() {
		codes = append(codes, "4")
	}
	if s.IsBlink() {
		codes = append(codes, "5")
	}
	if s.IsReverse() {
		codes = append(codes, "7")
	}
	if s.IsStrikethrough() {
		codes = append(codes, "9")
	}

	// Foreground color
	if s.FG != "" {
		codes = append(codes, string(s.FG))
	}

	// Background color
	if s.BG != "" {
		codes = append(codes, string(s.BG))
	}

	if len(codes) == 0 {
		return ""
	}

	return "\x1b[" + strings.Join(codes, ";") + "m"
}

// =============================================================================
// Cell Statistics
// =============================================================================

// CellStats provides statistics about a buffer for optimization.
type CellStats struct {
	TotalCells    int
	EmptyCells    int
	StyleChanges  int
	Runs          int
	AvgRunLength  float64
}

// AnalyzeBuffer analyzes a buffer and returns optimization statistics.
func AnalyzeBuffer(buf *Buffer) CellStats {
	stats := CellStats{
		TotalCells: buf.Width * buf.Height,
	}

	if buf.Width == 0 || buf.Height == 0 {
		return stats
	}

	// Count empty cells and style changes
	lastStyle := style.Style{}
	runCount := 0
	totalRunLength := 0

	for y := 0; y < buf.Height; y++ {
		runLength := 1
		for x := 1; x < buf.Width; x++ {
			if buf.Cells[y][x].Style != lastStyle {
				stats.StyleChanges++
				runCount++
				totalRunLength += runLength
				runLength = 1
			} else {
				runLength++
			}
			lastStyle = buf.Cells[y][x].Style
		}

		// Count empty cells (space clusters with default style)
		for x := 0; x < buf.Width; x++ {
			if buf.Cells[y][x].Cluster == " " || buf.Cells[y][x].Cluster == "" {
				if buf.Cells[y][x].Style == (style.Style{}) {
					stats.EmptyCells++
				}
			}
		}
	}

	stats.Runs = runCount
	if runCount > 0 {
		stats.AvgRunLength = float64(totalRunLength) / float64(runCount)
	}

	return stats
}

// =============================================================================
// RLE Stats
// =============================================================================

// RLEStats tracks RLE rendering performance metrics.
type RLEStats struct {
	FramesRendered  int
	TotalCells      int64
	DirtyCells      int64
	OutputBytes     int64
	OutputReduction float64 // Percentage reduction from optimization
}

// RecordFrame records metrics for a rendered frame.
func (s *RLEStats) RecordFrame(totalCells, dirtyCells, outputBytes int) {
	s.FramesRendered++
	s.TotalCells += int64(totalCells)
	s.DirtyCells += int64(dirtyCells)
	s.OutputBytes += int64(outputBytes)

	// Calculate theoretical max (no optimization)
	maxBytes := totalCells * 20 // Approximate max per cell
	if s.TotalCells > 0 {
		s.OutputReduction = 100 * (1 - float64(s.OutputBytes)/float64(maxBytes*s.FramesRendered))
	}
}

// String returns a summary of render stats.
func (s *RLEStats) String() string {
	return strings.Join([]string{
		"RLE Stats:",
		"Frames: " + fmt.Sprintf("%d", s.FramesRendered),
		"Dirty/Total: " + fmt.Sprintf("%d/%d", s.DirtyCells, s.TotalCells),
		"Output: " + fmt.Sprintf("%d bytes", s.OutputBytes),
		"Reduction: " + fmt.Sprintf("%.1f%%", s.OutputReduction),
	}, " ")
}
