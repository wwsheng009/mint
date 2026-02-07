// Package render provides paint engine for rendering computed layouts
package render

import (
	"fmt"
	"os"

	"github.com/wwsheng009/mint/runtime/border"
	"github.com/wwsheng009/mint/runtime/compute"
	"github.com/wwsheng009/mint/runtime/layer"
	"github.com/wwsheng009/mint/runtime/paint"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/runtime/style"
)

// PaintEngine renders VNode trees using pre-computed layout information
// This is the paint-only phase of the new rendering pipeline
type PaintEngine struct {
	debug         bool
	lastHadModal  bool  // Track if modal was present in last frame (for backdrop restoration)
	forceFullRender bool // Flag to force full buffer render on next frame
}

// NewPaintEngine creates a new paint engine
func NewPaintEngine() *PaintEngine {
	return &PaintEngine{
		debug: os.Getenv("TUI_PAINT_DEBUG") == "true",
	}
}

// SetDebug enables/disables debug output
func (e *PaintEngine) SetDebug(debug bool) {
	e.debug = debug
}

// Paint renders a computed layout to a buffer
func (e *PaintEngine) Paint(layout *compute.ComputedLayout, buffer *paint.Buffer) error {
	if layout == nil || layout.Root == nil {
		return nil
	}

	if os.Getenv("TUI_DEBUG_RENDERING") == "true" {
		fmt.Fprintf(os.Stderr, "[PaintEngine.Paint] START: layout.Root=%T, box=(%d,%d,%dx%d)\n",
			layout.Root.VNode, layout.Root.Box.X, layout.Root.Box.Y, layout.Root.Box.Width, layout.Root.Box.Height)
	}

	// If force full render is set (e.g., modal appeared/disappeared), clear buffer
	if e.forceFullRender {
		e.forceFullRender = false
		// Clear the entire buffer to force re-render of all cells
		for y := 0; y < buffer.Height; y++ {
			for x := 0; x < buffer.Width; x++ {
				buffer.Cells[y][x] = paint.Cell{}
			}
		}
	}

	err := e.paintNode(layout.Root, buffer)
	if os.Getenv("TUI_DEBUG_RENDERING") == "true" {
		fmt.Fprintf(os.Stderr, "[PaintEngine.Paint] END: err=%v\n", err)
	}
	return err
}

// paintNode recursively paints a computed box and its children
func (e *PaintEngine) paintNode(box *compute.ComputedBox, buffer *paint.Buffer) error {
	if box == nil {
		return nil
	}

	if e.debug || os.Getenv("TUI_PAINT_DEBUG") == "true" || os.Getenv("TUI_DEBUG_RENDERING") == "true" {
		fmt.Fprintf(os.Stderr, "[Paint.paintNode] %s at (%d,%d) size %dx%d, vnode_type=%T\n",
			box.VNode.Type().String(), box.Box.X, box.Box.Y, box.Box.Width, box.Box.Height, box.VNode)
	}

	// FIRST: Check if vnode implements Paintable interface (custom rendering like buttons)
	paintable, ok := box.VNode.(interface{ Paint(int, int) []paint.DrawCmd })
	if ok {
		if e.debug || os.Getenv("TUI_PAINT_DEBUG") == "true" || os.Getenv("TUI_DEBUG_RENDERING") == "true" {
			fmt.Fprintf(os.Stderr, "[Paint.paintNode]   ✅ Paintable: YES, calling Paint(%d, %d)\n", box.Box.X, box.Box.Y)
		}
		// Component has custom paint logic - use it
		commands := paintable.Paint(box.Box.X, box.Box.Y)
		for _, cmd := range commands {
			buffer.SetString(cmd.X, cmd.Y, cmd.Text, cmd.Style)
		}
		// Paintable components handle their own rendering, including children
		return nil
	} else {
		if e.debug || os.Getenv("TUI_PAINT_DEBUG") == "true" || os.Getenv("TUI_DEBUG_RENDERING") == "true" {
			fmt.Fprintf(os.Stderr, "[Paint.paintNode]   ❌ Paintable: NO (type assertion failed)\n")
		}
	}

	// Paint the node based on its type
	switch box.VNode.Type() {
	case rtui.VNodeText:
		e.paintText(box, buffer)

	case rtui.VNodeElement:
		e.paintElement(box, buffer)

	case rtui.VNodeComponent:
		// Component nodes should be expanded before painting
		// In non-Fiber mode, expandComponents() handles this
		// In Fiber mode, components are already expanded in the Fiber tree

	case rtui.VNodeFragment:
		// Fragment - just paint children, no self-rendering
		return e.paintChildren(box, buffer)
	}

	// Handle bordered elements - paint border decoration
	if _, ok := box.VNode.(interface{ GetBorderLabel() string }); ok {
		e.paintBordered(box, buffer)
		return nil
	}

	// For non-bordered elements, paint children after self-rendering
	children := box.VNode.Children()
	if len(children) > 0 {
		// Check if this is a table element
		if tagger, ok := box.VNode.(interface{ Tag() string }); ok && tagger.Tag() == "table" {
			e.paintTable(box, buffer)
			return nil
		}
	}

	// Paint children using their computed positions
	return e.paintChildren(box, buffer)
}

// paintText paints a text node
func (e *PaintEngine) paintText(box *compute.ComputedBox, buffer *paint.Buffer) {
	// Use RenderedText calculated during layout phase if available
	text := box.RenderedText
	if text == "" {
		text = rtui.GetTextContent(box.VNode)
	}
	if e.debug {
		fmt.Fprintf(os.Stderr, "[Paint.paintText] box=(%d,%d,%dx%d) renderedText=%q text=%q\n",
			box.Box.X, box.Box.Y, box.Box.Width, box.Box.Height, box.RenderedText, text)
	}
	if text != "" {
		buffer.SetString(box.Box.X, box.Box.Y, text, box.VNode.Style())
	}
}

// paintElement paints an element node
func (e *PaintEngine) paintElement(box *compute.ComputedBox, buffer *paint.Buffer) {
	// Use RenderedText calculated during layout phase if available
	content := box.RenderedText
	if content == "" {
		content = rtui.GetTextContent(box.VNode)
	}
	if content != "" {
		buffer.SetString(box.Box.X, box.Box.Y, content, box.VNode.Style())
		return // Don't paint children for text elements
	}
	// For non-text elements, children will be painted after the switch
}

// paintChildren paints children of a node using their computed positions
func (e *PaintEngine) paintChildren(box *compute.ComputedBox, buffer *paint.Buffer) error {
	for _, childBox := range box.Children {
		if err := e.paintNode(childBox, buffer); err != nil {
			return err
		}
	}
	return nil
}

// paintBordered paints a bordered node with border decoration
func (e *PaintEngine) paintBordered(box *compute.ComputedBox, buffer *paint.Buffer) {
	// Get border configuration
	borderStyle := border.StyleSingle
	if labeled, ok := box.VNode.(interface{ GetBorderStyle() rtui.BorderStyle }); ok {
		borderStyle = border.Style(labeled.GetBorderStyle())
	}
	borderColor := "blue"
	if colored, ok := box.VNode.(interface{ GetBorderColor() string }); ok {
		borderColor = colored.GetBorderColor()
	}
	borderLabel := ""
	if labeled, ok := box.VNode.(interface{ GetBorderLabel() string }); ok {
		borderLabel = labeled.GetBorderLabel()
	}

	// Create border renderer
	config := border.Config{
		Style: borderStyle,
		Color: borderColor,
		Label: borderLabel,
	}

	renderer := border.WithConfig(config)

	// Content dimensions (without border)
	contentWidth := box.Box.Width - 2
	contentHeight := box.Box.Height - 2
	if contentWidth < 0 {
		contentWidth = 0
	}
	if contentHeight < 0 {
		contentHeight = 0
	}

	// Paint border
	renderer.Paint(box.Box.X, box.Box.Y, contentWidth, contentHeight,
		func(px, py int, ch rune, s style.Style) {
			buffer.SetCell(px, py, ch, s)
		})

	// Paint children (content inside border)
	for _, childBox := range box.Children {
		if err := e.paintNode(childBox, buffer); err != nil && e.debug {
			fmt.Fprintf(os.Stderr, "[PaintBordered] error: %v\n", err)
		}
	}
}

// paintTable paints a table element
func (e *PaintEngine) paintTable(box *compute.ComputedBox, buffer *paint.Buffer) error {
	// Tables use the computed positions for their cells
	// Just need to paint children at their computed positions
	return e.paintChildren(box, buffer)
}

// =============================================================================
// Multi-Layer Rendering
// =============================================================================

// PaintLayers renders multiple layers in order (from lowest to highest)
// This is the main entry point for layer-based rendering
func (e *PaintEngine) PaintLayers(
	layouts layer.LayerLayouts,
	buffer *paint.Buffer,
) error {
	// Check if modal layer exists (for backdrop restoration)
	_, hasModal := layouts[rtui.LayerModal]
	hadModal := e.lastHadModal

	// If modal state changed, force full render to restore/clear backdrop
	if hasModal != hadModal {
		e.forceFullRender = true
	}
	e.lastHadModal = hasModal

	// Render layers in order from lowest (base) to highest (tooltip)
	// This ensures proper z-ordering
	renderOrder := []rtui.Layer{
		rtui.LayerBase,
		rtui.LayerOverlay,
		rtui.LayerModal,
		rtui.LayerTooltip,
	}

	for _, l := range renderOrder {
		layout, ok := layouts[l]
		if !ok || layout.Root == nil {
			continue
		}

		if e.debug {
			fmt.Fprintf(os.Stderr, "[PaintLayers] Rendering layer: %s root=(%d,%d) size=%dx%d\n",
				l.String(), layout.Root.Box.X, layout.Root.Box.Y, layout.Root.Box.Width, layout.Root.Box.Height)
		}

		// Paint this layer
		if err := e.Paint(layout, buffer); err != nil {
			return fmt.Errorf("error painting layer %s: %w", l.String(), err)
		}

		// Special handling for modal layer - draw backdrop
		if l == rtui.LayerModal {
			e.paintModalBackdrop(layout.Root, buffer)
		}
	}

	return nil
}

// paintModalBackdrop draws a semi-transparent backdrop behind the modal
// In TUI, we simulate this by:
// 1. Setting a dimmed background color for all areas outside the modal
// 2. Dimming the foreground text to gray
func (e *PaintEngine) paintModalBackdrop(root *compute.ComputedBox, buffer *paint.Buffer) {
	if root == nil {
		return
	}

	// Get buffer dimensions
	width, height := buffer.Width, buffer.Height

	// Modal bounds
	modalX := root.Box.X
	modalY := root.Box.Y
	modalWidth := root.Box.Width
	modalHeight := root.Box.Height

	// Dimmed style: gray foreground on dark background (simulates transparency)
	dimmedFG := style.Color("bright-black")  // Dimmed text color
	dimmedBG := style.Color("#1e2028")       // Dark overlay background (nord0 darker)

	// Helper function to apply dimmed effect to a cell
	applyDimmed := func(x, y int) {
		cell := buffer.GetContent(x, y)

		if cell.Cluster == "" || cell.Cluster == " " {
			// Empty cell: just set dimmed background
			buffer.SetCell(x, y, ' ', style.Style{BG: dimmedBG})
		} else {
			// Cell with content: dimmed foreground + dimmed background
			dimmedStyle := style.Style{
				FG: dimmedFG,
				BG: dimmedBG,
			}
			runeStr := cell.Cluster
			if len(runeStr) > 0 {
				buffer.SetCell(x, y, []rune(runeStr)[0], dimmedStyle)
			}
		}
	}

	// Apply dimmed effect to all areas outside the modal
	// Area above modal
	for y := 0; y < modalY && y < height; y++ {
		for x := 0; x < width; x++ {
			applyDimmed(x, y)
		}
	}

	// Area below modal
	for y := modalY + modalHeight; y < height; y++ {
		for x := 0; x < width; x++ {
			applyDimmed(x, y)
		}
	}

	// Area left of modal
	for y := modalY; y < modalY+modalHeight && y < height; y++ {
		for x := 0; x < modalX; x++ {
			applyDimmed(x, y)
		}
	}

	// Area right of modal
	for y := modalY; y < modalY+modalHeight && y < height; y++ {
		for x := modalX + modalWidth; x < width; x++ {
			applyDimmed(x, y)
		}
	}
}
