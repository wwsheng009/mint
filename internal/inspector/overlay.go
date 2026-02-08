package inspector

import (
	"fmt"
	"strings"

	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// OverlayColor defines color schemes for inspector overlays
type OverlayColor struct {
	Foreground style.Color
	Background style.Color
}

// ColorScheme defines colors for different element types and states
type ColorScheme struct {
	Selected     OverlayColor
	Hovered      OverlayColor
	Flex         OverlayColor
	Button       OverlayColor
	Text         OverlayColor
	Input        OverlayColor
	Container    OverlayColor
	Dimension    OverlayColor
	CornerTag    OverlayColor
}

// DefaultColorScheme returns the default color scheme
func DefaultColorScheme() *ColorScheme {
	return &ColorScheme{
		Selected: OverlayColor{
			Foreground: style.Yellow,
			Background: style.Black,
		},
		Hovered: OverlayColor{
			Foreground: style.Cyan,
			Background: style.Black,
		},
		Flex: OverlayColor{
			Foreground: style.Magenta,
			Background: style.Black,
		},
		Button: OverlayColor{
			Foreground: style.Green,
			Background: style.Black,
		},
		Text: OverlayColor{
			Foreground: style.White,
			Background: style.Black,
		},
		Input: OverlayColor{
			Foreground: style.Blue,
			Background: style.Black,
		},
		Container: OverlayColor{
			Foreground: style.BrightBlack,
			Background: style.Black,
		},
		Dimension: OverlayColor{
			Foreground: style.Yellow,
			Background: style.Black,
		},
		CornerTag: OverlayColor{
			Foreground: style.Magenta,
			Background: style.Black,
		},
	}
}

// Overlay provides visual highlighting of inspected elements
type Overlay struct {
	// Border styles
	selectedBorder []rune
	hoveredBorder  []rune
	flexBorder     []rune

	// Color scheme
	colors *ColorScheme

	// Display options
	showDimensions   bool
	showBorders      bool
	showCornerTags   bool
	showElementTypes bool
	showPadding      bool
}

// NewOverlay creates a new Overlay instance with default settings
func NewOverlay() *Overlay {
	return &Overlay{
		selectedBorder:  []rune("│"),  // Vertical bars for sides
		hoveredBorder:   []rune("┃"),  // Double vertical bars
		flexBorder:      []rune("║"),  // Hash marks for flex children
		colors:          DefaultColorScheme(),
		showDimensions:  true,
		showBorders:     true,
		showCornerTags:  true,
		showElementTypes: true,
		showPadding:     false,
	}
}

// SetColorScheme sets a custom color scheme
func (o *Overlay) SetColorScheme(scheme *ColorScheme) {
	o.colors = scheme
}

// GetColorScheme returns the current color scheme
func (o *Overlay) GetColorScheme() *ColorScheme {
	return o.colors
}

// SetShowCornerTags controls whether to show corner type indicators
func (o *Overlay) SetShowCornerTags(show bool) {
	o.showCornerTags = show
}

// SetShowElementTypes controls whether to show element type labels
func (o *Overlay) SetShowElementTypes(show bool) {
	o.showElementTypes = show
}

// SetShowPadding controls whether to show padding visualization
func (o *Overlay) SetShowPadding(show bool) {
	o.showPadding = show
}

// Paint renders the overlay on the given buffer
func (o *Overlay) Paint(buf *paint.Buffer, selected, hovered rtui.VNode) error {
	if !o.showBorders {
		return nil
	}

	// Draw border around selected element
	if selected != nil {
		o.drawElementBorder(buf, selected, o.selectedBorder, true)
		if o.showCornerTags {
			o.drawCornerTags(buf, selected)
		}
		if o.showElementTypes {
			o.drawElementType(buf, selected)
		}
		if o.showPadding {
			o.drawPadding(buf, selected)
		}
	}

	// Draw border around hovered element
	if hovered != nil && hovered != selected {
		o.drawElementBorder(buf, hovered, o.hoveredBorder, false)
		if o.showCornerTags {
			o.drawCornerTags(buf, hovered)
		}
	}

	return nil
}

// drawElementBorder draws a colored border around the given VNode
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

	// Determine color
	color := o.getColorForVNode(vnode, isSelected)

	// Create styled border
	borderStyleObj := style.Style{
		FG: color.Foreground,
		BG: color.Background,
	}

	// Draw border
	for i := 0; i < w; i++ {
		// Top and bottom edges
		if x+i < buf.Width && y < buf.Height {
			buf.SetCell(x+i, y, borderStyle[0], borderStyleObj)
			if h > 1 && y+h-1 < buf.Height {
				buf.SetCell(x+i, y+h-1, borderStyle[0], borderStyleObj)
			}
		}
	}

	for i := 0; i < h; i++ {
		// Left and right edges
		if y+i < buf.Height {
			if x < buf.Width {
				buf.SetCell(x, y+i, borderStyle[0], borderStyleObj)
			}
			if x+w-1 < buf.Width {
				buf.SetCell(x+w-1, y+i, borderStyle[0], borderStyleObj)
			}
		}
	}

	// Show dimensions if enabled
	if o.showDimensions && isSelected {
		o.drawDimensions(buf, x, y, w, h)
	}
}

// drawDimensions draws enhanced size annotations with color
func (o *Overlay) drawDimensions(buf *paint.Buffer, x, y, w, h int) {
	dimText := strings.TrimSpace(fmt.Sprintf("%dx%d", w, h))

	// Try to draw above the element
	if y > 0 && x+len(dimText) < buf.Width {
		dimStyle := style.Style{
			FG: o.colors.Dimension.Foreground,
			BG: o.colors.Dimension.Background,
		}
		for i, ch := range dimText {
			buf.SetCell(x+i, y-1, ch, dimStyle)
		}
	}
}

// drawCornerTags draws element type indicators at corners
func (o *Overlay) drawCornerTags(buf *paint.Buffer, vnode rtui.VNode) {
	if vnode == nil {
		return
	}

	boundsAware, ok := vnode.(interface{ GetBounds() [4]int })
	if !ok {
		return
	}

	bounds := boundsAware.GetBounds()
	if bounds == [4]int{0, 0, 0, 0} {
		return
	}

	x, y, _, _ := bounds[0], bounds[1], bounds[2], bounds[3]

	// Get corner indicator for element type
	tag := o.getCornerIndicator(vnode)
	if tag == 0 {
		return
	}

	tagStyle := style.Style{
		FG: o.colors.CornerTag.Foreground,
		BG: o.colors.CornerTag.Background,
	}

	// Draw tag at top-left corner (inside border)
	if x+1 < buf.Width && y < buf.Height {
		buf.SetCell(x+1, y, tag, tagStyle)
	}
}

// drawElementType draws element type label
func (o *Overlay) drawElementType(buf *paint.Buffer, vnode rtui.VNode) {
	if vnode == nil {
		return
	}

	boundsAware, ok := vnode.(interface{ GetBounds() [4]int })
	if !ok {
		return
	}

	bounds := boundsAware.GetBounds()
	if bounds == [4]int{0, 0, 0, 0} {
		return
	}

	x, y, w, h := bounds[0], bounds[1], bounds[2], bounds[3]

	// Get element type name
	typeName := getElementTypeName(vnode)
	if typeName == "" {
		return
	}

	// Truncate if too long
	maxLen := w - 2 // Leave space for borders
	if len(typeName) > maxLen {
		typeName = typeName[:maxLen]
	}

	// Draw below the element
	if y+h < buf.Height {
		typeStyle := style.Style{
			FG: o.colors.Dimension.Foreground,
			BG: o.colors.Dimension.Background,
		}
		for i, ch := range typeName {
			if x+i < buf.Width {
				buf.SetCell(x+i, y+h, ch, typeStyle)
			}
		}
	}
}

// drawPadding visualizes padding if available
func (o *Overlay) drawPadding(buf *paint.Buffer, vnode rtui.VNode) {
	// Check for padding information
	if props := vnode.Props(); props != nil {
		if padding, ok := props["padding"].(int); ok && padding > 0 {
			boundsAware, ok := vnode.(interface{ GetBounds() [4]int })
			if !ok {
				return
			}

			bounds := boundsAware.GetBounds()
			if bounds == [4]int{0, 0, 0, 0} {
				return
			}

			x, y, w, _ := bounds[0], bounds[1], bounds[2], bounds[3]

			// Draw padding indicators (dots)
			paddingStyle := style.Style{
				FG: style.BrightBlack,
				BG: style.Black,
			}

			// Draw dots at padding boundaries
			for i := 0; i < padding && i < w; i++ {
				if y+1 < buf.Height && x+i < buf.Width {
					buf.SetCell(x+i, y+1, '·', paddingStyle)
				}
			}
		}
	}
}

// getColorForVNode returns the appropriate color for a VNode
func (o *Overlay) getColorForVNode(vnode rtui.VNode, isSelected bool) OverlayColor {
	if isSelected {
		return o.colors.Selected
	}

	// Check if it's a flex child
	if props := vnode.Props(); props != nil {
		if flex, ok := props["flex"].(int); ok && flex > 0 {
			return o.colors.Flex
		}
	}

	// Check element type
	tag := ""
	if tagger, ok := vnode.(interface{ Tag() string }); ok {
		tag = tagger.Tag()
	}

	switch tag {
	case "button":
		return o.colors.Button
	case "text":
		return o.colors.Text
	case "input":
		return o.colors.Input
	case "hstack", "vstack", "box", "border":
		return o.colors.Container
	default:
		return o.colors.Hovered
	}
}

// getCornerIndicator returns a corner indicator character for element type
func (o *Overlay) getCornerIndicator(vnode rtui.VNode) rune {
	if vnode == nil {
		return 0
	}

	// Check element type
	tag := ""
	if tagger, ok := vnode.(interface{ Tag() string }); ok {
		tag = tagger.Tag()
	}

	switch tag {
	case "button":
		return '█' // Solid block for buttons
	case "text":
		return '▪' // Small square for text
	case "input":
		return '▬' // Rectangle for input
	case "hstack":
		return '→' // Arrow for HStack
	case "vstack":
		return '↓' // Arrow for VStack
	case "box":
		return '■' // Box for Box
	case "border":
		return '╔' // Border corner
	default:
		return 0
	}
}

// getElementTypeName returns a short type name for the element
func getElementTypeName(vnode rtui.VNode) string {
	if vnode == nil {
		return ""
	}

	// Check tag
	tag := ""
	if tagger, ok := vnode.(interface{ Tag() string }); ok {
		tag = tagger.Tag()
	}

	// Return short names
	typeNames := map[string]string{
		"button":    "BTN",
		"text":      "TXT",
		"input":     "IN",
		"hstack":    "H",
		"vstack":    "V",
		"box":       "BOX",
		"border":    "BORDER",
		"checkbox":  "CHK",
		"select":    "SEL",
		"textarea":  "TXTA",
	}

	if name, ok := typeNames[tag]; ok {
		return name
	}

	return ""
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

	highlightStyle := style.Style{
		FG: o.colors.Selected.Foreground,
		BG: o.colors.Selected.Background,
	}

	// Draw corners
	if x < buf.Width && y < buf.Height {
		buf.SetCell(x, y, char, highlightStyle)
	}
	if x+w-1 < buf.Width && y < buf.Height {
		buf.SetCell(x+w-1, y, char, highlightStyle)
	}
	if x < buf.Width && y+h-1 < buf.Height {
		buf.SetCell(x, y+h-1, char, highlightStyle)
	}
	if x+w-1 < buf.Width && y+h-1 < buf.Height {
		buf.SetCell(x+w-1, y+h-1, char, highlightStyle)
	}

	return nil
}
