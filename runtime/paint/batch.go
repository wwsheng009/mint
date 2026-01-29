package paint

import (
	"bytes"
	"sort"

	"github.com/mattn/go-runewidth"
	"github.com/wwsheng009/mint/runtime/style"
)

// DrawCmd represents a single drawing command
type DrawCmd struct {
	X, Y  int
	Text  string
	Style style.Style
}

// CommandBatch batches draw commands to minimize terminal IO
type CommandBatch struct {
	cmds    []DrawCmd
	styleVM *StyleStateMachine
	// Cursor position tracking for optimization
	curX int
	curY int
}

// NewCommandBatch creates a new command batch
func NewCommandBatch() *CommandBatch {
	return &CommandBatch{
		cmds:    make([]DrawCmd, 0, 256),
		styleVM: NewStyleStateMachine(),
		curX:    -1, // Unknown initial position
		curY:    -1,
	}
}

// Add adds a draw command
func (b *CommandBatch) Add(x, y int, text string, st style.Style) {
	b.cmds = append(b.cmds, DrawCmd{
		X:     x,
		Y:     y,
		Text:  text,
		Style: st,
	})
}

// AddCell adds a single cell command
func (b *CommandBatch) AddCell(x, y int, char rune, st style.Style) {
	b.Add(x, y, string(char), st)
}

// Flush merges commands and generates the final output
func (b *CommandBatch) Flush() string {
	if len(b.cmds) == 0 {
		return ""
	}

	var buf bytes.Buffer
	b.styleVM.Reset()
	b.curX, b.curY = -1, -1 // Reset cursor tracking

	// Sort by Y then X for linear traversal
	b.sortCommands()

	// Merge adjacent commands with same style
	merged := b.mergeCommands()

	// Generate output with style state machine and cursor optimization
	for _, cmd := range merged {
		// Move cursor if needed (with optimization)
		cursorCmd := b.moveCursorOptimized(cmd.X, cmd.Y)
		if cursorCmd != "" {
			buf.WriteString(cursorCmd)
		}

		// Apply style if changed
		if b.styleVM.NeedsUpdate(cmd.Style) {
			buf.WriteString(b.styleVM.Update(cmd.Style))
		}

		// Write text and update cursor position
		buf.WriteString(cmd.Text)
		b.curX = cmd.X + runewidth.StringWidth(cmd.Text)
		b.curY = cmd.Y
	}

	// Reset style at end
	buf.WriteString("\x1b[0m")

	return buf.String()
}

// mergeCommands merges adjacent commands that can be combined
func (b *CommandBatch) mergeCommands() []DrawCmd {
	if len(b.cmds) == 0 {
		return nil
	}

	merged := make([]DrawCmd, 0, len(b.cmds))
	current := b.cmds[0]

	for i := 1; i < len(b.cmds); i++ {
		next := b.cmds[i]

		// Check if we can merge
		if b.canMerge(current, next) {
			current.Text += next.Text
		} else {
			merged = append(merged, current)
			current = next
		}
	}

	merged = append(merged, current)
	return merged
}

// canMerge checks if two commands can be merged
func (b *CommandBatch) canMerge(a, c DrawCmd) bool {
	// Must be on same line
	if a.Y != c.Y {
		return false
	}

	// Must be adjacent
	if a.X+len(a.Text) != c.X {
		return false
	}

	// Must have same style
	return a.Style == c.Style
}

// sortCommands sorts commands by Y then X
func (b *CommandBatch) sortCommands() {
	sort.Slice(b.cmds, func(i, j int) bool {
		if b.cmds[i].Y != b.cmds[j].Y {
			return b.cmds[i].Y < b.cmds[j].Y
		}
		return b.cmds[i].X < b.cmds[j].X
	})
}

// moveCursorOptimized generates optimized ANSI cursor movement
// Optimization rules:
// - Same position: no output
// - Same line, small step right: use \x1b[nC (forward)
// - Same line, large step: use absolute positioning
// - Different line: use absolute positioning
func (b *CommandBatch) moveCursorOptimized(x, y int) string {
	// Unknown initial position, use absolute
	if b.curX < 0 || b.curY < 0 {
		b.curX, b.curY = x, y
		return "\x1b[" + itoa(y+1) + ";" + itoa(x+1) + "H"
	}

	// Same position, no move needed
	if b.curX == x && b.curY == y {
		return ""
	}

	// Same line, moving right
	if b.curY == y && x > b.curX {
		delta := x - b.curX
		// Small step: use relative forward cursor
		if delta <= 5 {
			b.curX = x
			return "\x1b[" + itoa(delta) + "C"
		}
	}

	// Default: absolute positioning
	b.curX, b.curY = x, y
	return "\x1b[" + itoa(y+1) + ";" + itoa(x+1) + "H"
}

// moveCursor generates ANSI cursor movement (kept for compatibility)
func (b *CommandBatch) moveCursor(x, y int) string {
	return "\x1b[" + itoa(y+1) + ";" + itoa(x+1) + "H"
}

// Clear clears all commands
func (b *CommandBatch) Clear() {
	b.cmds = b.cmds[:0]
}

// Count returns the number of commands
func (b *CommandBatch) Count() int {
	return len(b.cmds)
}

// Reserve reserves space for additional commands
func (b *CommandBatch) Reserve(n int) {
	if cap(b.cmds) < len(b.cmds)+n {
		newCmds := make([]DrawCmd, len(b.cmds), cap(b.cmds)+n+256)
		copy(newCmds, b.cmds)
		b.cmds = newCmds
	}
}

// MergeFrom merges another batch into this one
func (b *CommandBatch) MergeFrom(other *CommandBatch) {
	if other == nil || len(other.cmds) == 0 {
		return
	}
	b.Reserve(len(other.cmds))
	b.cmds = append(b.cmds, other.cmds...)
}

// EstimateSize estimates the size of the flushed output
func (b *CommandBatch) EstimateSize() int {
	if len(b.cmds) == 0 {
		return 0
	}

	// Rough estimate: each character + cursor moves + style codes
	size := 0
	for _, cmd := range b.cmds {
		size += len(cmd.Text) + 20 // Approx for cursor and style
	}
	return size
}

// WriteToBuffer writes all commands to a paint Buffer
func (b *CommandBatch) WriteToBuffer(buf *Buffer) {
	for _, cmd := range b.cmds {
		buf.SetString(cmd.X, cmd.Y, cmd.Text, cmd.Style)
	}
}

// WriteToBufferWithOffset writes all commands to a paint Buffer with offset
func (b *CommandBatch) WriteToBufferWithOffset(buf *Buffer, offsetX, offsetY int) {
	for _, cmd := range b.cmds {
		buf.SetString(cmd.X+offsetX, cmd.Y+offsetY, cmd.Text, cmd.Style)
	}
}
