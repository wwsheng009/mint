package inspector

import (
	"fmt"
	"strings"

	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// Overlay provides visual highlighting of inspected elements
type Overlay struct {
	// Border styles
	selectedBorder []rune
	hoveredBorder  []rune
	flexBorder     []rune

	// Display options
	showDimensions bool
	showBorders    bool
}

// NewOverlay creates a new Overlay instance
func NewOverlay() *Overlay {
	return &Overlay{
		selectedBorder: []rune("│"),  // Vertical bars for sides
		hoveredBorder:  []rune("┃"),  // Double vertical bars
		flexBorder:     []rune("║"),  // Hash marks for flex children
		showDimensions: true,
		showBorders:    true,
	}
}

// Paint renders the overlay on the given buffer
func (o *Overlay) Paint(buf *paint.Buffer, selected, hovered rtui.VNode) error {
	if !o.showBorders {
		return nil
	}

	// Draw border around selected element
	if selected != nil {
		o.drawElementBorder(buf, selected, o.selectedBorder, true)
	}

	// Draw border around hovered element
	if hovered != nil && hovered != selected {
		o.drawElementBorder(buf, hovered, o.hoveredBorder, false)
	}

	return nil
}

// drawElementBorder draws a border around the given VNode
func (o *Overlay) drawElementBorder(buf *paint.Buffer, vnode rtui.VNode, borderStyle []rune, isSelected bool) {
	if vnode == nil {
		return
	}

	// Get bounds
	boundsAware, ok := vnode.(interface{ GetBounds() [4]int })
	if !ok {
		return
	}

	bounds := boundsAware.GetBounds()
	if bounds == [4]int{0, 0, 0, 0} {
		return
	}

	x, y, w, h := bounds[0], bounds[1], bounds[2], bounds[3]

	// Draw border
	for i := 0; i < w; i++ {
		// Top and bottom edges
		if x+i < buf.Width && y < buf.Height {
			buf.SetCell(x+i, y, borderStyle[0], style.Style{})
			if h > 1 && y+h-1 < buf.Height {
				buf.SetCell(x+i, y+h-1, borderStyle[0], style.Style{})
			}
		}
	}

	for i := 0; i < h; i++ {
		// Left and right edges
		if y+i < buf.Height {
			if x < buf.Width {
				buf.SetCell(x, y+i, borderStyle[0], style.Style{})
			}
			if x+w-1 < buf.Width {
				buf.SetCell(x+w-1, y+i, borderStyle[0], style.Style{})
			}
		}
	}

	// Show dimensions if enabled
	if o.showDimensions && isSelected {
		o.drawDimensions(buf, x, y, w, h)
	}
}

// drawDimensions draws size annotations
func (o *Overlay) drawDimensions(buf *paint.Buffer, x, y, w, h int) {
	dimText := strings.TrimSpace(fmt.Sprintf("%dx%d", w, h))

	// Try to draw above the element
	if y > 0 && x+len(dimText) < buf.Width {
		for i, ch := range dimText {
			buf.SetCell(x+i, y-1, ch, style.Style{})
		}
	}
}

// SetShowDimensions controls whether to show dimension annotations
func (o *Overlay) SetShowDimensions(show bool) {
	o.showDimensions = show
}

// SetShowBorders controls whether to show borders
func (o *Overlay) SetShowBorders(show bool) {
	o.showBorders = show
}

// GetBorderStyle returns the border style for a given element type
func (o *Overlay) GetBorderStyle(vnode rtui.VNode) []rune {
	if vnode == nil {
		return o.hoveredBorder
	}

	// Check if it's a flex child
	if props := vnode.Props(); props != nil {
		if flex, ok := props["flex"].(int); ok && flex > 0 {
			return o.flexBorder
		}
	}

	// Check element type
	tag := ""
	if tagger, ok := vnode.(interface{ Tag() string }); ok {
		tag = tagger.Tag()
	}

	switch tag {
	case "button":
		return []rune("▓") // Diamond-ish for buttons
	case "text":
		return []rune("•") // Bullet for text
	default:
		return o.hoveredBorder
	}
}

// PaintHighlight paints a simple highlight (single character at corners)
func (o *Overlay) PaintHighlight(buf *paint.Buffer, vnode rtui.VNode, char rune) error {
	if vnode == nil {
		return nil
	}

	boundsAware, ok := vnode.(interface{ GetBounds() [4]int })
	if !ok {
		return nil
	}

	bounds := boundsAware.GetBounds()
	if bounds == [4]int{0, 0, 0, 0} {
		return nil
	}

	x, y, w, h := bounds[0], bounds[1], bounds[2], bounds[3]

	// Draw corners
	if x < buf.Width && y < buf.Height {
		buf.SetCell(x, y, char, style.Style{})
	}
	if x+w-1 < buf.Width && y < buf.Height {
		buf.SetCell(x+w-1, y, char, style.Style{})
	}
	if x < buf.Width && y+h-1 < buf.Height {
		buf.SetCell(x, y+h-1, char, style.Style{})
	}
	if x+w-1 < buf.Width && y+h-1 < buf.Height {
		buf.SetCell(x+w-1, y+h-1, char, style.Style{})
	}

	return nil
}
