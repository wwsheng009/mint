package selection

import (
	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/style"
)

// Renderer handles rendering of selection highlights on a CellBuffer.
type Renderer struct {
	manager  *Manager
	buffer   *runtime.CellBuffer
	highlight CellStyle
}

// NewRenderer creates a new selection renderer.
func NewRenderer(manager *Manager, buffer *runtime.CellBuffer) *Renderer {
	return &Renderer{
		manager:  manager,
		buffer:   buffer,
		highlight: DefaultHighlightStyle(),
	}
}

// SetHighlightStyle sets the style used for selection highlighting.
func (r *Renderer) SetHighlightStyle(style CellStyle) {
	r.highlight = style
}

// GetHighlightStyle returns the current highlight style.
func (r *Renderer) GetHighlightStyle() CellStyle {
	return r.highlight
}

// ApplySelection applies the selection highlight to the buffer.
// This modifies the buffer in place, adding reverse styling to selected cells.
func (r *Renderer) ApplySelection() {
	if !r.manager.IsActive() || r.buffer == nil {
		return
	}

	cells := r.manager.GetSelectedCells()
	for _, cell := range cells {
		r.highlightCell(cell.X, cell.Y)
	}
}

// highlightCell applies the highlight style to a single cell.
func (r *Renderer) highlightCell(x, y int) {
	if r.buffer == nil {
		return
	}

	// Get the current cell
	currentCell := r.buffer.GetContent(x, y)

	// Create a new style that combines the original with the highlight
	newStyle := r.combineStyles(currentCell.Style, r.highlight)

	// Extract first rune from cluster for SetCell
	char := rune(0)
	for _, c := range currentCell.Cluster {
		char = c
		break
	}
	// Set the cell with the new style
	r.buffer.SetContent(x, y, currentCell.ZIndex, char, newStyle, "")
}

// combineStyles combines two cell styles, with the highlight taking precedence.
func (r *Renderer) combineStyles(original runtime.CellStyle, highlight CellStyle) runtime.CellStyle {
	result := original

	// If highlight uses reverse, that's our primary selection indicator
	if highlight.Reverse {
		result = result.Reverse(true)
	}

	// Add additional styling from highlight
	if highlight.Bold {
		result = result.Bold(true)
	}
	if highlight.Underline {
		result = result.Underline(true)
	}
	if highlight.Italic {
		result = result.Italic(true)
	}

	return result
}

// ApplySelectionToFrame applies selection highlighting to a frame's buffer.
// This is a convenience method that works directly with a Frame.
func ApplySelectionToFrame(frame *runtime.Frame, manager *Manager) {
	if !manager.IsActive() || frame == nil || frame.Buffer == nil {
		return
	}

	renderer := NewRenderer(manager, frame.Buffer)
	renderer.ApplySelection()
}

// DefaultHighlightStyle returns the default selection highlight style.
// Uses reverse video for maximum visibility across different color schemes.
func DefaultHighlightStyle() CellStyle {
	return CellStyle{
		Reverse: true,  // Reverse video is most visible
		Bold:    false,
		Underline: false,
	}
}

// LightHighlightStyle returns a light theme selection highlight style.
// Uses a subtle blue background.
func LightHighlightStyle() CellStyle {
	return CellStyle{
		Background: "#4A90E2",
		Foreground: "white",
		Bold:       true,
	}
}

// DarkHighlightStyle returns a dark theme selection highlight style.
// Uses a contrasting background.
func DarkHighlightStyle() CellStyle {
	return CellStyle{
		Background: "#607D8B",
		Foreground: "white",
		Bold:       true,
	}
}

// CellStyle represents the style for selection highlighting.
type CellStyle struct {
	Bold       bool
	Underline  bool
	Italic     bool
	Strikethrough bool
	Blink      bool
	Reverse    bool
	Foreground string
	Background string
}

// ToRuntimeStyle converts a selection.CellStyle to runtime.CellStyle.
func (s CellStyle) ToRuntimeStyle() runtime.CellStyle {
	st := runtime.CellStyle{}
	if s.Bold {
		st = st.Bold(true)
	}
	if s.Underline {
		st = st.Underline(true)
	}
	if s.Italic {
		st = st.Italic(true)
	}
	if s.Strikethrough {
		st = st.Strikethrough(true)
	}
	if s.Blink {
		st = st.Blink(true)
	}
	if s.Reverse {
		st = st.Reverse(true)
	}
	if s.Foreground != "" {
		st = st.Foreground(style.Color(s.Foreground))
	}
	if s.Background != "" {
		st = st.Background(style.Color(s.Background))
	}
	return st
}

// TextBufferAdapter adapts runtime.CellBuffer to selection.TextBuffer.
type TextBufferAdapter struct {
	buffer *runtime.CellBuffer
}

// NewTextBufferAdapter creates a new TextBufferAdapter.
func NewTextBufferAdapter(buffer *runtime.CellBuffer) TextBuffer {
	return &TextBufferAdapter{buffer: buffer}
}

// GetCell returns the cell at the given position.
func (a *TextBufferAdapter) GetCell(x, y int) Cell {
	if a.buffer == nil {
		return Cell{Empty: true}
	}

	runtimeCell := a.buffer.GetContent(x, y)
	// Extract first rune from cluster
	char := rune(0)
	for _, c := range runtimeCell.Cluster {
		char = c
		break
	}
	return Cell{
		Cluster: string(char),
		Empty: runtimeCell.Cluster == " " || runtimeCell.Cluster == "" || runtimeCell.Cluster == "\x00",
	}
}

// Width returns the buffer width.
func (a *TextBufferAdapter) Width() int {
	if a.buffer == nil {
		return 0
	}
	return a.buffer.Width
}

// Height returns the buffer height.
func (a *TextBufferAdapter) Height() int {
	if a.buffer == nil {
		return 0
	}
	return a.buffer.Height
}

// ManagerWithBuffer creates a new Manager with a CellBuffer as the text source.
func ManagerWithBuffer(buffer *runtime.CellBuffer) *Manager {
	adapter := NewTextBufferAdapter(buffer)
	return NewManager(adapter)
}

// CopySelectionToClipboard copies the current selection to the clipboard.
// Returns the copied text and any error that occurred.
func CopySelectionToClipboard(manager *Manager, clipboard *Clipboard) (string, error) {
	if !manager.IsActive() {
		return "", nil
	}

	text := manager.GetSelectedTextCompact()
	if text == "" {
		return "", nil
	}

	err := clipboard.Copy(text)
	return text, err
}

// CopySelectionToClipboardGlobal copies the selection using the global clipboard.
func CopySelectionToClipboardGlobal(manager *Manager) (string, error) {
	return CopySelectionToClipboard(manager, globalClipboard)
}
