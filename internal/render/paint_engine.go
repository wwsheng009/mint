// Package render provides paint engine for rendering computed layouts
package render

import (
	"fmt"
	"os"

	"github.com/wwsheng009/mint/runtime/border"
	"github.com/wwsheng009/mint/runtime/compute"
	"github.com/wwsheng009/mint/runtime/paint"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/runtime/style"
)

// PaintEngine renders VNode trees using pre-computed layout information
// This is the paint-only phase of the new rendering pipeline
type PaintEngine struct {
	debug bool
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
	return e.paintNode(layout.Root, buffer)
}

// paintNode recursively paints a computed box and its children
func (e *PaintEngine) paintNode(box *compute.ComputedBox, buffer *paint.Buffer) error {
	if box == nil {
		return nil
	}

	if e.debug {
		fmt.Fprintf(os.Stderr, "[Paint] %s at (%d,%d) size %dx%d\n",
			box.VNode.Type().String(), box.Box.X, box.Box.Y, box.Box.Width, box.Box.Height)
	}

	// FIRST: Check if vnode implements Paintable interface (custom rendering like buttons)
	if paintable, ok := box.VNode.(interface{ Paint(int, int) []paint.DrawCmd }); ok {
		// Component has custom paint logic - use it
		commands := paintable.Paint(box.Box.X, box.Box.Y)
		for _, cmd := range commands {
			buffer.SetString(cmd.X, cmd.Y, cmd.Text, cmd.Style)
		}
		// Paintable components handle their own rendering, including children
		return nil
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
