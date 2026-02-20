// Package render provides paint engine for rendering computed layouts
package render

import (
	"fmt"
	"os"

	"github.com/wwsheng009/mint/internal/log"
	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/border"
	"github.com/wwsheng009/mint/runtime/compute"
	"github.com/wwsheng009/mint/runtime/layer"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// PaintEngine renders layout trees using pre-computed layout information
// This is the paint-only phase of the new rendering pipeline
type PaintEngine struct {
	debug             bool
	lastHadModal      bool                               // Track if modal was present in last frame (for backdrop restoration)
	forceFullRender   bool                               // Flag to force full buffer render on next frame
	parentBackground  map[*paint.PaintableBox]style.Color // Track parent background for inheritance (refactored)
	parentBackgroundLegacy map[*compute.ComputedBox]style.Color // Legacy: for backward compatibility
	lastLayersPresent map[rtui.Layer]bool                // Track which layers were present in last frame
	lastLayerBounds   map[rtui.Layer]runtime.Box         // Track last bounds of each layer for cleanup
}

// NewPaintEngine creates a new paint engine
func NewPaintEngine() *PaintEngine {
	return &PaintEngine{
		debug:             log.PaintLogger.Enabled(),
		lastLayersPresent: make(map[rtui.Layer]bool),
		lastLayerBounds:   make(map[rtui.Layer]runtime.Box),
	}
}

// SetDebug enables/disables debug output
func (e *PaintEngine) SetDebug(debug bool) {
	e.debug = debug
}

// =============================================================================
// New API: PaintableLayout (Decoupled from VNode/Fiber)
// =============================================================================

// PaintLayout renders a PaintableLayout to a buffer.
// This is the new decoupled API that operates on paint.PaintableLayout.
func (e *PaintEngine) PaintLayout(layout *paint.PaintableLayout, buffer *paint.Buffer) error {
	if layout == nil || layout.Root == nil {
		return nil
	}

	// Clear parent background map at the start of each frame
	e.parentBackground = make(map[*paint.PaintableBox]style.Color)

	if e.forceFullRender {
		e.forceFullRender = false
		for y := 0; y < buffer.Height; y++ {
			for x := 0; x < buffer.Width; x++ {
				buffer.Cells[y][x] = paint.Cell{}
			}
		}
	}

	return e.paintBox(layout.Root, buffer)
}

// paintBox recursively paints a PaintableBox and its children
func (e *PaintEngine) paintBox(box *paint.PaintableBox, buffer *paint.Buffer) error {
	if box == nil || box.Node == nil {
		return nil
	}

	if e.debug || log.PaintLogger.Enabled() {
		log.PaintLogger.Debug("[Paint.paintBox] %s at (%d,%d) size %dx%d",
			box.Node.Tag(), box.X, box.Y, box.Width, box.Height)
	}

	// Check if we have a parent background to inherit
	var parentBG style.Color
	if e.parentBackground != nil {
		if inheritedBG, ok := e.parentBackground[box]; ok && inheritedBG != "" {
			parentBG = inheritedBG
		}
	}

	// FIRST: Check if node has custom paint logic
	// For container components (with children), we skip Paint and use LayoutBox coordinates
	// For leaf components (no children), we use Paint method
	commands := box.Node.Paint(box.X, box.Y)

	// Only use Paint method for leaf nodes (no children)
	// Container nodes use LayoutBox coordinates for children
	if len(commands) > 0 && len(box.Children) == 0 {
		// Apply commands with potential background inheritance
		for _, cmd := range commands {
			styleToApply := cmd.Style
			if parentBG != "" && (styleToApply.BG == "" || styleToApply.BG == style.NoColor) {
				styleToApply.BG = parentBG
			}
			buffer.SetString(cmd.X, cmd.Y, cmd.Text, styleToApply)
		}
		// Leaf component rendered, no children to process
		return nil
	}

	// Inherit parent background for non-Paintable nodes
	if parentBG != "" {
		nodeStyle := box.Node.Style()
		if nodeStyle.BG == "" || nodeStyle.BG == style.NoColor {
			nodeStyle.BG = parentBG
			box.Node.SetStyle(nodeStyle)
		}
	}

	// Paint based on node type
	switch box.Node.NodeType() {
	case paint.NodeTypeText:
		e.paintTextBox(box, buffer)
	case paint.NodeTypeElement:
		e.paintElementBox(box, buffer)
	case paint.NodeTypeFragment:
		return e.paintBoxChildren(box, buffer)
	}

	// Handle bordered elements
	if bs, bc, bl := box.GetBorderInfo(); bs != paint.BorderStyleNone {
		e.paintBorderedBox(box, buffer, bs, bc, bl)
		return nil
	}

	// Paint children
	return e.paintBoxChildren(box, buffer)
}

// paintTextBox paints a text node (PaintableBox version)
func (e *PaintEngine) paintTextBox(box *paint.PaintableBox, buffer *paint.Buffer) {
	text := box.RenderedText
	if text == "" {
		text = box.Node.TextContent()
	}
	if text != "" {
		// Use box.Width as the max width (relative), not absolute maxX
		buffer.SetStringAligned(box.X, box.Y, text, box.Node.Style(), box.Width)
	}
}

// paintElementBox paints an element node (PaintableBox version)
func (e *PaintEngine) paintElementBox(box *paint.PaintableBox, buffer *paint.Buffer) {
	content := box.RenderedText
	if content == "" {
		content = box.Node.TextContent()
	}
	if content != "" {
		// Use box.Width as the max width (relative), not absolute maxX
		buffer.SetStringAligned(box.X, box.Y, content, box.Node.Style(), box.Width)
		return
	}

	// Paint container background if set
	nodeStyle := box.Node.Style()
	if nodeStyle.BG != "" && nodeStyle.BG != style.NoColor {
		e.paintBoxContainerBackground(box, buffer, nodeStyle)
		
		// Store parent background for child inheritance
		if e.parentBackground == nil {
			e.parentBackground = make(map[*paint.PaintableBox]style.Color)
		}
		for _, childBox := range box.Children {
			e.parentBackground[childBox] = nodeStyle.BG
		}
	}
}

// paintBoxContainerBackground fills the container area with background color
func (e *PaintEngine) paintBoxContainerBackground(box *paint.PaintableBox, buffer *paint.Buffer, bgStyle style.Style) {
	backgroundStyle := style.Style{}.Background(bgStyle.BG)
	for y := 0; y < box.Height; y++ {
		for x := 0; x < box.Width; x++ {
			buffer.SetCell(box.X+x, box.Y+y, ' ', backgroundStyle)
		}
	}
}

// paintBoxChildren paints children of a PaintableBox
func (e *PaintEngine) paintBoxChildren(box *paint.PaintableBox, buffer *paint.Buffer) error {
	for _, childBox := range box.Children {
		if err := e.paintBox(childBox, buffer); err != nil {
			return err
		}
	}
	return nil
}

// paintBorderedBox paints a bordered PaintableBox
func (e *PaintEngine) paintBorderedBox(box *paint.PaintableBox, buffer *paint.Buffer, bs paint.BorderStyle, bc, bl string) {
	// Convert border style
	var borderStyle border.Style
	switch bs {
	case paint.BorderStyleDouble:
		borderStyle = border.StyleDouble
	case paint.BorderStyleRounded:
		borderStyle = border.StyleRounded
	default:
		borderStyle = border.StyleSingle
	}

	config := border.Config{
		Style: borderStyle,
		Color: bc,
		Label: bl,
	}
	renderer := border.WithConfig(config)

	contentWidth := box.Width - 2
	contentHeight := box.Height - 2
	if contentWidth < 0 {
		contentWidth = 0
	}
	if contentHeight < 0 {
		contentHeight = 0
	}

	renderer.Paint(box.X, box.Y, contentWidth, contentHeight,
		func(px, py int, ch rune, s style.Style) {
			buffer.SetCell(px, py, ch, s)
		})

	for _, childBox := range box.Children {
		if err := e.paintBox(childBox, buffer); err != nil && e.debug {
			log.PaintLogger.Debug("[paintBorderedBox] error: %v", err)
		}
	}
}

// paintModalBackdropBox draws modal backdrop (PaintableBox version)
func (e *PaintEngine) paintModalBackdropBox(root *paint.PaintableBox, buffer *paint.Buffer) {
	if root == nil {
		return
	}

	width, height := buffer.Width, buffer.Height
	modalX := root.X
	modalY := root.Y
	modalWidth := root.Width
	modalHeight := root.Height

	dimmedFG := style.Color("bright-black")
	dimmedBG := style.Color("#1e2028")

	applyDimmed := func(x, y int) {
		cell := buffer.GetContent(x, y)
		if cell.Cluster == "" || cell.Cluster == " " {
			buffer.SetCell(x, y, ' ', style.Style{BG: dimmedBG})
		} else {
			dimmedStyle := style.Style{FG: dimmedFG, BG: dimmedBG}
			runeStr := cell.Cluster
			if len(runeStr) > 0 {
				buffer.SetCell(x, y, []rune(runeStr)[0], dimmedStyle)
			}
		}
	}

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

// =============================================================================
// Legacy API: ComputedLayout (Backward Compatibility)
// =============================================================================

// Paint renders a computed layout to a buffer
// Deprecated: Use PaintLayout for decoupled API
func (e *PaintEngine) Paint(layout *compute.ComputedLayout, buffer *paint.Buffer) error {
	if layout == nil || layout.Root == nil {
		if os.Getenv("MINT_DEBUG_TEST") == "true" {
			fmt.Printf("[PaintEngine.Paint] layout or layout.Root is nil\n")
		}
		return nil
	}

	if os.Getenv("MINT_DEBUG_TEST") == "true" {
		fmt.Printf("[PaintEngine.Paint] START: layout.Root.Box=(%d,%d,%dx%d)\n",
			layout.Root.Box.X, layout.Root.Box.Y, layout.Root.Box.Width, layout.Root.Box.Height)
	}

	if log.PaintLogger.Enabled() {
		log.PaintLogger.Debug("[PaintEngine.Paint] START: box=(%d,%d,%dx%d)",
			layout.Root.Box.X, layout.Root.Box.Y, layout.Root.Box.Width, layout.Root.Box.Height)
	}

	// Convert to PaintableLayout and use new API
	paintableLayout := layout.AsPaintableLayout()
	return e.PaintLayout(paintableLayout, buffer)
}

// paintNode recursively paints a computed box and its children (Legacy)
// This method is kept for backward compatibility and delegates to paintBox.
func (e *PaintEngine) paintNode(box *compute.ComputedBox, buffer *paint.Buffer) error {
	if box == nil {
		return nil
	}

	// Convert to PaintableBox and use new paintBox method
	paintableBox := box.AsPaintable()
	return e.paintBox(paintableBox, buffer)
}

// paintText paints a text node (Legacy)
// Deprecated: Use paintTextBox
func (e *PaintEngine) paintText(box *compute.ComputedBox, buffer *paint.Buffer) {
	paintableBox := box.AsPaintable()
	e.paintTextBox(paintableBox, buffer)
}

// paintElement paints an element node (Legacy)
// Deprecated: Use paintElementBox
func (e *PaintEngine) paintElement(box *compute.ComputedBox, buffer *paint.Buffer) {
	paintableBox := box.AsPaintable()
	e.paintElementBox(paintableBox, buffer)
}

// paintContainerBackground fills the container area with background color (Legacy)
func (e *PaintEngine) paintContainerBackground(box *compute.ComputedBox, buffer *paint.Buffer, bgStyle style.Style) {
	paintableBox := box.AsPaintable()
	e.paintBoxContainerBackground(paintableBox, buffer, bgStyle)
}

// paintChildren paints children of a node (Legacy)
// Deprecated: Use paintBoxChildren
func (e *PaintEngine) paintChildren(box *compute.ComputedBox, buffer *paint.Buffer) error {
	paintableBox := box.AsPaintable()
	return e.paintBoxChildren(paintableBox, buffer)
}

// paintBordered paints a bordered node (Legacy)
// Deprecated: Use paintBorderedBox
func (e *PaintEngine) paintBordered(box *compute.ComputedBox, buffer *paint.Buffer) {
	bs, bc, bl := box.AsPaintable().GetBorderInfo()
	if bs != paint.BorderStyleNone {
		paintableBox := box.AsPaintable()
		e.paintBorderedBox(paintableBox, buffer, bs, bc, bl)
	}
}

// paintTable paints a table element (Legacy)
func (e *PaintEngine) paintTable(box *compute.ComputedBox, buffer *paint.Buffer) error {
	return e.paintChildren(box, buffer)
}

// paintModalBackdrop draws modal backdrop (Legacy)
// Deprecated: Use paintModalBackdropBox
func (e *PaintEngine) paintModalBackdrop(root *compute.ComputedBox, buffer *paint.Buffer) {
	paintableBox := root.AsPaintable()
	e.paintModalBackdropBox(paintableBox, buffer)
}

// clearRegion clears a rectangular region of the buffer
func (e *PaintEngine) clearRegion(bounds runtime.Box, buffer *paint.Buffer) {
	maxX := buffer.Width
	maxY := buffer.Height

	for y := bounds.Y; y < bounds.Y+bounds.Height && y < maxY; y++ {
		for x := bounds.X + bounds.Width - 1; x >= bounds.X && x < maxX; x-- {
			buffer.SetCell(x, y, ' ', style.Style{})
		}
	}

	if e.debug || log.PaintLogger.Enabled() {
		log.PaintLogger.Debug("[Paint.clearRegion] Cleared %dx%d region at (%d,%d)",
			bounds.Width, bounds.Height, bounds.X, bounds.Y)
	}
}

// =============================================================================
// Multi-Layer Rendering
// =============================================================================

// PaintLayers renders multiple layers in order (from lowest to highest)
func (e *PaintEngine) PaintLayers(
	layouts layer.LayerLayouts,
	buffer *paint.Buffer,
) error {
	_, hasModal := layouts[rtui.LayerModal]
	hadModal := e.lastHadModal

	if hasModal != hadModal {
		e.forceFullRender = true
	}
	e.lastHadModal = hasModal

	renderOrder := []rtui.Layer{
		rtui.LayerBase,
		rtui.LayerOverlay,
		rtui.LayerModal,
		rtui.LayerTooltip,
		rtui.LayerInspector,
	}

	for _, l := range renderOrder {
		hasLayer := false
		var currentBounds runtime.Box = runtime.Box{}
		if layout, ok := layouts[l]; ok && layout.Root != nil {
			hasLayer = true
			currentBounds = layout.Root.Box
		}
		hadLayer := e.lastLayersPresent[l]
		prevBounds := e.lastLayerBounds[l]

		if hadLayer && !hasLayer {
			log.PaintLogger.Debug("[PaintLayers] Layer %s disappeared, clearing region", l.String())
			e.clearRegion(prevBounds, buffer)
			e.forceFullRender = true
		}

		if hasLayer && hadLayer && currentBounds != prevBounds {
			e.forceFullRender = true
		}

		e.lastLayersPresent[l] = hasLayer
		if hasLayer {
			e.lastLayerBounds[l] = currentBounds
		} else {
			delete(e.lastLayerBounds, l)
		}
	}

	for _, l := range renderOrder {
		layout, ok := layouts[l]
		if !ok || layout.Root == nil {
			continue
		}

		if e.debug {
			log.PaintLogger.Debug("[PaintLayers] Rendering layer: %s", l.String())
		}

		if err := e.Paint(layout, buffer); err != nil {
			return fmt.Errorf("error painting layer %s: %w", l.String(), err)
		}

		if l == rtui.LayerModal {
			e.paintModalBackdrop(layout.Root, buffer)
		}
	}

	return nil
}

// PaintRenderPlanes paints RenderPlanes to buffer directly
func (e *PaintEngine) PaintRenderPlanes(
	renderPlanes *layer.RenderPlanes,
	buffer *paint.Buffer,
) error {
	if renderPlanes == nil {
		return nil
	}

	log.PaintLogger.Debug("[PaintEngine.PaintRenderPlanes] START: boxes=%d", renderPlanes.CountBoxes())

	for _, layer := range renderPlanes.GetRenderOrder() {
		boxes := renderPlanes.GetLayer(layer)
		if boxes == nil || len(boxes) == 0 {
			continue
		}

		log.PaintLogger.Debug("[PaintEngine.PaintRenderPlanes] Layer %s: %d boxes", layer.String(), len(boxes))

		for _, box := range boxes {
			// Convert ComputedBox to PaintableBox and use new API
			paintableBox := box.AsPaintable()
			layout := paint.NewPaintableLayout(paintableBox)
			if err := e.PaintLayout(layout, buffer); err != nil {
				return fmt.Errorf("error painting box in layer %s: %w", layer.String(), err)
			}
		}

		if layer == rtui.LayerModal && len(boxes) > 0 {
			paintableBox := boxes[0].AsPaintable()
			e.paintModalBackdropBox(paintableBox, buffer)
		}
	}

	log.PaintLogger.Debug("[PaintEngine.PaintRenderPlanes] END")
	return nil
}
