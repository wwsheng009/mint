// Package render provides paint engine for rendering computed layouts
package render

import (
	"fmt"

	"github.com/wwsheng009/mint/internal/log"
	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/border"
	"github.com/wwsheng009/mint/runtime/compute"
	"github.com/wwsheng009/mint/runtime/layer"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// PaintEngine renders VNode trees using pre-computed layout information
// This is the paint-only phase of the new rendering pipeline
type PaintEngine struct {
	debug             bool
	lastHadModal      bool                                 // Track if modal was present in last frame (for backdrop restoration)
	forceFullRender   bool                                 // Flag to force full buffer render on next frame
	parentBackground  map[*compute.ComputedBox]style.Color // Track parent background for inheritance
	lastLayersPresent map[rtui.Layer]bool                  // Track which layers were present in last frame
	lastLayerBounds   map[rtui.Layer]runtime.Box           // Track last bounds of each layer for cleanup
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

// Paint renders a computed layout to a buffer
func (e *PaintEngine) Paint(layout *compute.ComputedLayout, buffer *paint.Buffer) error {
	if layout == nil || layout.Root == nil {
		return nil
	}

	if log.PaintLogger.Enabled() {
		log.PaintLogger.Debug("[PaintEngine.Paint] START: layout.Root=%T, box=(%d,%d,%dx%d)",
			layout.Root.GetVNode(), layout.Root.Box.X, layout.Root.Box.Y, layout.Root.Box.Width, layout.Root.Box.Height)
	}

	// Clear parent background map at the start of each frame
	// This prevents stale background inheritance from previous frames
	e.parentBackground = make(map[*compute.ComputedBox]style.Color)

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
	if log.PaintLogger.Enabled() {
		log.PaintLogger.Debug("[PaintEngine.Paint] END: err=%v", err)
	}
	return err
}

// paintNode recursively paints a computed box and its children
func (e *PaintEngine) paintNode(box *compute.ComputedBox, buffer *paint.Buffer) error {
	if box == nil {
		return nil
	}

	if e.debug || log.PaintLogger.Enabled() || log.PaintLogger.Enabled() {
		vnode := box.GetVNode()
		var vnodeType string
		if vnode != nil {
			vnodeType = vnode.Type().String()
		}
		log.PaintLogger.Debug("[Paint.paintNode] %s at (%d,%d) size %dx%d, vnode_type=%T",
			vnodeType, box.Box.X, box.Box.Y, box.Box.Width, box.Box.Height, vnode)
	}

	// Check if we have a parent background to inherit
	var parentBG style.Color
	cleanUpParentBG := false
	if e.parentBackground != nil {
		if inheritedBG, ok := e.parentBackground[box]; ok && inheritedBG != "" {
			parentBG = inheritedBG
			cleanUpParentBG = true
		}
	}

	// FIRST: Check if vnode implements Paintable interface (custom rendering like buttons)
	vnode := box.GetVNode()
	if vnode == nil {
		// No VNode - skip paintable handling
		return e.paintChildren(box, buffer)
	}
	paintable, ok := vnode.(interface {
		Paint(int, int) []paint.DrawCmd
	})
	if ok {
		if e.debug || log.PaintLogger.Enabled() || log.PaintLogger.Enabled() {
			log.PaintLogger.Debug("[Paint.paintNode]   ✅ Paintable: YES, calling Paint(%d, %d)", box.Box.X, box.Box.Y)
		}
		// Component has custom paint logic - use it
		commands := paintable.Paint(box.Box.X, box.Box.Y)

		// Apply commands with potential background inheritance
		for _, cmd := range commands {
			styleToApply := cmd.Style
			// If command has no background and parent has one, inherit it
			if parentBG != "" && (styleToApply.BG == "" || styleToApply.BG == style.NoColor) {
				styleToApply.BG = parentBG
				if e.debug || log.PaintLogger.Enabled() {
					log.PaintLogger.Debug("[Paint.paintNode]   🎨 Paintable inherited parent BG=%s", parentBG)
				}
			}
			buffer.SetString(cmd.X, cmd.Y, cmd.Text, styleToApply)
		}

		// Clean up parent background entry
		if cleanUpParentBG {
			delete(e.parentBackground, box)
		}

		// Paintable components handle their own rendering, including children
		return nil
	} else {
		if e.debug || log.PaintLogger.Enabled() || log.PaintLogger.Enabled() {
			log.PaintLogger.Debug("[Paint.paintNode]   ❌ Paintable: NO (type assertion failed)")
		}
	}

	// ENHANCEMENT: Inherit parent background for non-Paintable nodes
	// This ensures child controls blend with parent container's background
	if parentBG != "" {
		vnode := box.GetVNode()
		if vnode == nil {
			// No VNode to inherit from
			return nil
		}
		nodeStyle := vnode.Style()
		// Only inherit if node doesn't have its own background
		if nodeStyle.BG == "" || nodeStyle.BG == style.NoColor {
			// Inherit parent background
			inheritedStyle := nodeStyle
			inheritedStyle.BG = parentBG
			vnode.SetStyle(inheritedStyle)

			if e.debug || log.PaintLogger.Enabled() {
				log.PaintLogger.Debug("[Paint.paintNode]   🎨 Inherited parent BG=%s", parentBG)
			}
		}
	}

	// Clean up parent background entry
	if cleanUpParentBG {
		delete(e.parentBackground, box)
	}

	// Paint the node based on its type
	// Reuse vnode from earlier (line 102) - it's already in scope
	if vnode == nil {
		// No VNode associated - this can happen when Fiber is nil
		// Just paint children without special handling
		return e.paintChildren(box, buffer)
	}
	switch vnode.Type() {
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
	// Reuse vnode from earlier (line 102) - it's already in scope
	if vnode == nil {
		// No VNode - skip border decoration
		return e.paintChildren(box, buffer)
	}
	if _, ok := vnode.(interface{ GetBorderLabel() string }); ok {
		e.paintBordered(box, buffer)
		return nil
	}

	// For non-bordered elements, paint children after self-rendering
	children := vnode.Children()
	if len(children) > 0 {
		// Check if this is a table element
		if tagger, ok := vnode.(interface{ Tag() string }); ok && tagger.Tag() == "table" {
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
	var vnode rtui.VNode
	if text == "" {
		vnode = box.GetVNode()
		if vnode == nil {
			return
		}
		text = rtui.GetTextContent(vnode)
	}
	if e.debug {
		log.PaintLogger.Debug("[Paint.paintText] box=(%d,%d,%dx%d) renderedText=%q text=%q",
			box.Box.X, box.Box.Y, box.Box.Width, box.Box.Height, box.RenderedText, text)
	}
	if text != "" {
		// Use SetStringAligned to pad row and prevent leftover characters (TUI_BUFFER_FIX2.md)
		// Pad to box.Box.X + box.Box.Width to fill entire row within component boundary
		maxX := box.Box.X + box.Box.Width
		buffer.SetStringAligned(box.Box.X, box.Box.Y, text, vnode.Style(), maxX)
	}
}

// paintElement paints an element node
func (e *PaintEngine) paintElement(box *compute.ComputedBox, buffer *paint.Buffer) {
	// Use RenderedText calculated during layout phase if available
	content := box.RenderedText
	var vnode rtui.VNode
	if content == "" {
		vnode = box.GetVNode()
		if vnode == nil {
			return
		}
		content = rtui.GetTextContent(vnode)
	}
	if content != "" {
		// Use SetStringAligned to pad row and prevent leftover characters (TUI_BUFFER_FIX2.md)
		// Pad to box.Box.X + box.Box.Width to fill entire row within component boundary
		maxX := box.Box.X + box.Box.Width
		buffer.SetStringAligned(box.Box.X, box.Box.Y, content, vnode.Style(), maxX)
		return // Don't paint children for text elements
	}

	// ENHANCEMENT: Paint container background if set
	// This allows elements like Inspector panels to have solid backgrounds
	if vnode == nil {
		vnode = box.GetVNode()
		if vnode == nil {
			return
		}
	}
	nodeStyle := vnode.Style()
	if nodeStyle.BG != "" && nodeStyle.BG != style.NoColor {
		e.paintContainerBackground(box, buffer, nodeStyle)

		// IMPORTANT: Store parent background for child inheritance
		// Children without explicit background will inherit this background
		if e.parentBackground == nil {
			e.parentBackground = make(map[*compute.ComputedBox]style.Color)
		}
		for _, childBox := range box.Children {
			e.parentBackground[childBox] = nodeStyle.BG
		}
	}

	// For non-text elements, children will be painted after the switch
}

// paintContainerBackground fills the entire container area with background color
// This is used to create solid backgrounds for panels like Inspector
// IMPORTANT: This must be called BEFORE painting children to occlude underlying content
func (e *PaintEngine) paintContainerBackground(box *compute.ComputedBox, buffer *paint.Buffer, bgStyle style.Style) {
	// Create background style (only BG, no foreground)
	backgroundStyle := style.Style{}.Background(bgStyle.BG)

	// CRITICAL: Unconditionally fill entire container area with background color
	// This occludes any content rendered underneath (e.g., from lower layers)
	// Children will be painted on top of this background
	for y := 0; y < box.Box.Height; y++ {
		for x := 0; x < box.Box.Width; x++ {
			// Unconditionally set background to occlude underlying content
			// Use space character ' ' to clear any existing content
			buffer.SetCell(box.Box.X+x, box.Box.Y+y, ' ', backgroundStyle)
		}
	}

	if e.debug || log.PaintLogger.Enabled() {
		log.PaintLogger.Debug("[Paint.paintContainerBackground] Occluded %dx%d area at (%d,%d) with BG=%s",
			box.Box.Width, box.Box.Height, box.Box.X, box.Box.Y, bgStyle.BG)
	}
}

// clearRegion clears a rectangular region of the buffer
// This is used to clean up areas that were previously occupied by disappeared layers
func (e *PaintEngine) clearRegion(bounds runtime.Box, buffer *paint.Buffer) {
	// Clamp bounds to buffer size
	maxX := buffer.Width
	maxY := buffer.Height

	for y := bounds.Y; y < bounds.Y+bounds.Height && y < maxY; y++ {
		// CRITICAL: Clear from right to left to properly handle wide characters
		// If we clear from left to right, when we hit a continuation cell,
		// it will clear the head at (x-1), which might already have been processed
		for x := bounds.X + bounds.Width - 1; x >= bounds.X && x < maxX; x-- {
			// Clear the cell by setting to empty
			buffer.SetCell(x, y, ' ', style.Style{})
		}
	}

	if e.debug || log.PaintLogger.Enabled() {
		log.PaintLogger.Debug("[Paint.clearRegion] Cleared %dx%d region at (%d,%d)",
			bounds.Width, bounds.Height, bounds.X, bounds.Y)
	}
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
	// Debug: print box info
	log.PaintLogger.Debug("[PaintBordered] START: box.Box.X=%d, box.Box.Y=%d, box.Box.Width=%d, box.Box.Height=%d",
		box.Box.X, box.Box.Y, box.Box.Width, box.Box.Height)
	log.PaintLogger.Debug("[PaintBordered] box.Children count = %d", len(box.Children))

	// Get VNode for border configuration
	vnode := box.GetVNode()
	if vnode == nil {
		// No VNode - cannot get border config, just paint children
		log.PaintLogger.Debug("[PaintBordered] vnode is nil, painting children only")
		e.paintChildren(box, buffer)
		return
	}
	log.PaintLogger.Debug("[PaintBordered] vnode.Type=%v", vnode.Type())
	borderStyle := border.StyleSingle
	if labeled, ok := vnode.(interface{ GetBorderStyle() rtui.BorderStyle }); ok {
		borderStyle = border.Style(labeled.GetBorderStyle())
	}
	borderColor := "blue"
	if colored, ok := vnode.(interface{ GetBorderColor() string }); ok {
		borderColor = colored.GetBorderColor()
	}
	borderLabel := ""
	if labeled, ok := vnode.(interface{ GetBorderLabel() string }); ok {
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
	childCount := len(box.Children)
	if e.debug || log.PaintLogger.Enabled() {
		log.PaintLogger.Debug("[PaintBordered] box.Children count = %d", childCount)
	}
	if childCount == 0 {
		log.PaintLogger.Debug("[PaintBordered] WARNING: No children to paint!")
	}
	for _, childBox := range box.Children {
		log.PaintLogger.Debug("[PaintBordered] Painting child: X=%d Y=%d W=%d H=%d",
			childBox.Box.X, childBox.Box.Y, childBox.Box.Width, childBox.Box.Height)
		if err := e.paintNode(childBox, buffer); err != nil && e.debug {
			log.PaintLogger.Debug("[PaintBordered] error: %v", err)
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

	// Check if any layer's presence changed
	// This is important for clearing inspector content when it's hidden
	// Also track layer bounds to clear specific regions when layers disappear
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

		// If layer disappeared, clear its previous region
		if hadLayer && !hasLayer {
			log.PaintLogger.Debug("[PaintLayers] Layer %s disappeared, clearing region: (%d,%d) %dx%d",
				l.String(), prevBounds.X, prevBounds.Y, prevBounds.Width, prevBounds.Height)

			// Clear the region that was previously occupied by this layer
			e.clearRegion(prevBounds, buffer)
			e.forceFullRender = true // Force full render to ensure proper repaint
		}

		// If layer layer bounds changed significantly, also force full render
		if hasLayer && hadLayer && currentBounds != prevBounds {
			log.PaintLogger.Debug("[PaintLayers] Layer %s bounds changed: (%d,%d) %dx%d -> (%d,%d) %dx%d",
				l.String(), prevBounds.X, prevBounds.Y, prevBounds.Width, prevBounds.Height,
				currentBounds.X, currentBounds.Y, currentBounds.Width, currentBounds.Height)

			e.forceFullRender = true
		}

		e.lastLayersPresent[l] = hasLayer
		if hasLayer {
			e.lastLayerBounds[l] = currentBounds
		} else {
			delete(e.lastLayerBounds, l)
		}
	}

	// Render layers in order from lowest (base) to highest (inspector)
	// This ensures proper z-ordering

	for _, l := range renderOrder {
		layout, ok := layouts[l]
		if !ok || layout.Root == nil {
			continue
		}

		if e.debug {
			log.PaintLogger.Debug("[PaintLayers] Rendering layer: %s root=(%d,%d) size=%dx%d",
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

// PaintRenderPlanes paints RenderPlanes to buffer directly
// This is the new unified API replacing PaintLayers for the Fiber architecture
// It iterates through RenderPlanes by layer and paints each ComputedBox
//
// Parameters:
//   renderPlanes - RenderPlanes containing all boxes organized by layer
//   buffer - Paint buffer to render to
//
// Returns:
//   error - Any painting error
func (e *PaintEngine) PaintRenderPlanes(
	renderPlanes *layer.RenderPlanes,
	buffer *paint.Buffer,
) error {
	if renderPlanes == nil {
		return nil
	}

	log.PaintLogger.Debug("[PaintEngine.PaintRenderPlanes] START: boxes=%d", renderPlanes.CountBoxes())

	// Iterate through layers in render order (low to high)
	for _, layer := range renderPlanes.GetRenderOrder() {
		boxes := renderPlanes.GetLayer(layer)
		if boxes == nil || len(boxes) == 0 {
			continue
		}

		log.PaintLogger.Debug("[PaintEngine.PaintRenderPlanes] Layer %s: %d boxes", layer.String(), len(boxes))

		// Paint each box in this layer
		for _, box := range boxes {
			// Create a temporary ComputedLayout for each box
			// This allows us to reuse the existing Paint() method
			layout := &compute.ComputedLayout{
				Root: box,
			}
			if err := e.Paint(layout, buffer); err != nil {
				log.PaintLogger.Debug("[PaintEngine.PaintRenderPlanes] Paint failed: %v", err)
				return fmt.Errorf("error painting box in layer %s: %w", layer.String(), err)
			}
		}

		// Special handling for modal layer - draw backdrop
		if layer == rtui.LayerModal && len(boxes) > 0 {
			e.paintModalBackdrop(boxes[0], buffer)
		}
	}

	log.PaintLogger.Debug("[PaintEngine.PaintRenderPlanes] END")
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
	dimmedFG := style.Color("bright-black") // Dimmed text color
	dimmedBG := style.Color("#1e2028")      // Dark overlay background (nord0 darker)

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
