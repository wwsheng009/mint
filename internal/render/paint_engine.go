// Package render provides paint engine for rendering computed layouts
package render

import (
	"fmt"
	"sync"

	"github.com/wwsheng009/mint/internal/log"
	cachepkg "github.com/wwsheng009/mint/internal/render/cache"
	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/border"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// PaintEngine renders layout trees using pre-computed layout information
// This is the paint-only phase of the new rendering pipeline
type PaintEngine struct {
	mu                sync.Mutex
	debug             bool
	lastHadModal      bool                                // Track if modal was present in last frame (for backdrop restoration)
	forceFullRender   bool                                // Flag to force full buffer render on next frame
	parentBackground  map[*paint.PaintableBox]style.Color // Track parent background for inheritance
	lastLayersPresent map[rtui.Layer]bool                 // Track which layers were present in last frame
	lastLayerBounds   map[rtui.Layer]runtime.Box          // Track last bounds of each layer for cleanup

	// Frame-to-frame tracking for smart buffer clearing
	// These track individual PaintableBox positions to detect and clear outdated regions
	previousFrameBoxes map[string]runtime.Box // key: box.Node.ID(), value: box bounds
	currentFrameBoxes  map[string]runtime.Box // key: box.Node.ID(), value: box bounds (being rendered)

	// Performance optimization: Paint cache
	cache        *cachepkg.PaintCache      // Cache for rendered paintable boxes
	enableCache  bool                      // Enable cache (true by default)
	version      int                       // Current render version (for cache invalidation)
	paintContext *cachepkg.PaintingContext // Context for cache-aware painting
}

// NewPaintEngine creates a new paint engine
func NewPaintEngine() *PaintEngine {
	return &PaintEngine{
		debug:              log.PaintLogger.Enabled(),
		lastLayersPresent:  make(map[rtui.Layer]bool),
		lastLayerBounds:    make(map[rtui.Layer]runtime.Box),
		previousFrameBoxes: make(map[string]runtime.Box),
		currentFrameBoxes:  make(map[string]runtime.Box),
		enableCache:        true,
		version:            0,
	}
}

// InitCache initializes the paint cache with the given buffer
func (e *PaintEngine) InitCache(buffer *paint.Buffer) {
	if !e.enableCache {
		return
	}
	if e.cache == nil {
		e.cache = cachepkg.NewPaintCache()
	}
	e.version++
	// PaintEngine uses renderer-level dirty tracking and does not call
	// PaintingContext.GetDirtyRects during normal painting. Avoid cloning the
	// full frame buffer here; callers that need dirty rects can still create a
	// PaintingContext with a buffer directly.
	e.paintContext = cachepkg.NewPaintingContext(e.cache, nil, e.version)
}

// EnableCache enables or disables the paint cache
func (e *PaintEngine) EnableCache(enable bool) {
	e.enableCache = enable
	if !enable && e.cache != nil {
		// Clear cache when disabling
		e.cache.Clear()
	}
}

// InvalidateCache invalidates all cached entries
func (e *PaintEngine) InvalidateCache() {
	if e.cache != nil {
		e.cache.InvalidateAll()
	}
}

// GetCacheStats returns paint cache statistics
func (e *PaintEngine) GetCacheStats() cachepkg.CacheStats {
	if e.cache == nil {
		return cachepkg.CacheStats{}
	}
	return e.cache.Stats()
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
	return e.paintLayout(layout, buffer, true)
}

// paintLayout renders one paintable layout.
// clearOnEmptyRoot should only be true for top-level/full-frame calls.
// When painting individual boxes (e.g. PaintPaintablePlanes), a zero-sized box
// must not clear the entire target buffer.
func (e *PaintEngine) paintLayout(layout *paint.PaintableLayout, buffer *paint.Buffer, clearOnEmptyRoot bool) error {
	if layout == nil || layout.Root == nil {
		return nil
	}

	// Empty/zero-sized roots can represent "no content" for top-level layout calls.
	// But they may also intentionally rely on custom Paint() commands (for example,
	// overlay/tooltip helpers that paint outside their measured layout box). Do not
	// return early here; let paintBox decide whether the node actually emits content.
	if layout.Root.Width <= 0 || layout.Root.Height <= 0 {
		if clearOnEmptyRoot && len(layout.Root.Children) == 0 {
			buffer.Clear()
		}
	}

	if buffer == nil {
		return nil
	}

	e.beginPaintFrame(buffer)

	return e.paintBox(layout.Root, buffer)
}

// beginPaintFrame initializes per-frame paint state once for a top-level paint
// operation. PaintPaintablePlanes may paint many boxes in one frame, so this
// must not be repeated for each individual box.
func (e *PaintEngine) beginPaintFrame(buffer *paint.Buffer) {
	if buffer == nil {
		return
	}

	if e.enableCache {
		e.InitCache(buffer)
		// e.paintContext.UpdateBufferCopy(buffer)
		// REMOVED: This buffer copy was causing 33% CPU overhead (cloneBuffer).
		// The system already uses DirtyTracker for diff tracking (runtime/paint/dirty.go).
		// See DIFF_REDUNDANCY_ANALYSIS.md for details.
	}

	// Clear parent background map at the start of each frame
	if e.parentBackground == nil {
		e.parentBackground = make(map[*paint.PaintableBox]style.Color)
	} else {
		clear(e.parentBackground)
	}

	if e.forceFullRender {
		e.forceFullRender = false
		for y := 0; y < buffer.Height; y++ {
			for x := 0; x < buffer.Width; x++ {
				buffer.Cells[y][x] = paint.Cell{}
			}
		}
		// Invalidate cache on full render
		if e.cache != nil {
			e.cache.InvalidateAll()
		}
	}
}

// paintBox recursively paints a PaintableBox and its children.
func (e *PaintEngine) paintBox(box *paint.PaintableBox, buffer *paint.Buffer) error {
	return e.paintBoxWithMode(box, buffer, true)
}

// paintBoxOnly paints only the current PaintableBox. PaintablePlanes already
// contains every box in layer order, so recursive painting there would repaint
// descendants once for every ancestor.
func (e *PaintEngine) paintBoxOnly(box *paint.PaintableBox, buffer *paint.Buffer) error {
	return e.paintBoxWithMode(box, buffer, false)
}

func (e *PaintEngine) paintBoxWithMode(box *paint.PaintableBox, buffer *paint.Buffer, paintChildren bool) error {
	if box == nil || box.Node == nil {
		return nil
	}

	log.PaintLogger.Debug("[Paint.paintBox] paint element: [%s]%s at (%d,%d) size %dx%d",
		box.Node.ID(), box.Node.Tag(), box.X, box.Y, box.Width, box.Height)

	if viewportSetter, ok := box.Node.(interface{ SetViewportSize(width, height int) }); ok {
		viewportSetter.SetViewportSize(buffer.Width, buffer.Height)
	}

	// IMPORTANT: Set bounds before Paint (Fiber-first architecture)
	// This allows Instance to access layout-computed dimensions
	if boundsSetter, ok := box.Node.(interface{ SetBounds(x, y, w, h int) }); ok {
		boundsSetter.SetBounds(box.X, box.Y, box.Width, box.Height)
	}

	// Query custom paint once. Many component instances allocate draw commands
	// in Paint(), so probing cacheability and then painting again doubles work.
	commands := box.Node.Paint(box.X, box.Y)

	// Check cache first (for leaf nodes without custom paint commands)
	// Skip caching for nodes with dynamic content (like text inputs, animations)
	cacheKey := e.boxCacheKey(box)
	if e.enableCache && e.paintContext != nil && cacheKey != "" {
		// Only try caching for nodes that are likely cacheable (simple layout nodes)
		// Skip nodes with custom paint commands, children, or dynamic content
		if len(box.Children) == 0 && len(commands) == 0 {
			// Try to paint from cache
			if e.paintContext.TryPaintFromCache(buffer, cacheKey, box.X, box.Y) {
				return nil // Successfully painted from cache
			}
		}
	}

	// Check if we have a parent background to inherit
	var parentBG style.Color
	if e.parentBackground != nil {
		if inheritedBG, ok := e.parentBackground[box]; ok && inheritedBG != "" {
			parentBG = inheritedBG
		}
	}

	// FIRST: Check if node has custom paint logic
	// Apply custom paint commands if present (for both leaf and container nodes)
	// Container components like Border return border drawing commands
	if len(commands) > 0 {
		for _, cmd := range commands {
			styleToApply := cmd.Style
			if parentBG != "" && (styleToApply.BG == "" || styleToApply.BG == style.NoColor) {
				styleToApply.BG = parentBG
			}
			buffer.SetString(cmd.X, cmd.Y, cmd.Text, styleToApply)
		}
		// For leaf nodes (no children), we're done
		if len(box.Children) == 0 {
			// Update cache for leaf nodes with custom paint
			if e.enableCache && e.paintContext != nil && cacheKey != "" {
				rect := layout.Rect{X: box.X, Y: box.Y, Width: box.Width, Height: box.Height}
				e.paintContext.UpdateCache(cacheKey, rect, buffer)
			}
			return nil
		}
		// For container nodes, continue to paint children
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
		if paintChildren {
			return e.paintBoxChildren(box, buffer)
		}
		return nil
	}

	// Handle bordered elements (legacy path - for nodes without custom Paint)
	if bs, bc, bl := box.GetBorderInfo(); bs != paint.BorderStyleNone && len(commands) == 0 {
		e.paintBorderedBox(box, buffer, bs, bc, bl, paintChildren)
		return nil
	}

	// Paint children
	if paintChildren {
		return e.paintBoxChildren(box, buffer)
	}
	return nil
}

// paintTextBox paints a text node (PaintableBox version)
func (e *PaintEngine) paintTextBox(box *paint.PaintableBox, buffer *paint.Buffer) {
	cacheKey := e.boxCacheKey(box)
	text := box.RenderedText
	if text == "" {
		text = box.Node.TextContent()
	}
	if text != "" {
		// SetStringAligned expects an absolute maxX; clip text to box bounds.
		buffer.SetStringAligned(box.X, box.Y, text, box.Node.Style(), box.X+box.Width)

		// Update cache for text box (if cacheable)
		if e.enableCache && e.paintContext != nil && cacheKey != "" {
			rect := layout.Rect{X: box.X, Y: box.Y, Width: box.Width, Height: box.Height}
			e.paintContext.UpdateCache(cacheKey, rect, buffer)
		}
	}
}

// paintElementBox paints an element node (PaintableBox version)
func (e *PaintEngine) paintElementBox(box *paint.PaintableBox, buffer *paint.Buffer) {
	cacheKey := e.boxCacheKey(box)
	content := box.RenderedText
	if content == "" {
		content = box.Node.TextContent()
	}
	if content != "" {
		// SetStringAligned expects an absolute maxX; clip text to box bounds.
		buffer.SetStringAligned(box.X, box.Y, content, box.Node.Style(), box.X+box.Width)

		// Update cache for element with content (if cacheable)
		if e.enableCache && e.paintContext != nil && cacheKey != "" {
			rect := layout.Rect{X: box.X, Y: box.Y, Width: box.Width, Height: box.Height}
			e.paintContext.UpdateCache(cacheKey, rect, buffer)
		}
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

// paintBorderedBox paints a bordered PaintableBox.
func (e *PaintEngine) paintBorderedBox(box *paint.PaintableBox, buffer *paint.Buffer, bs paint.BorderStyle, bc, bl string, paintChildren bool) {
	// Convert border style
	var borderStyle border.Style
	switch bs {
	case paint.BorderStyleDouble:
		borderStyle = border.StyleDouble
	case paint.BorderStyleRounded:
		borderStyle = border.StyleRounded
	case paint.BorderStyleDashed:
		borderStyle = border.StyleDashed
	default:
		borderStyle = border.StyleSingle
	}

	// IMPORTANT: All border styles (including Double) occupy only 1 character cell width
	// Double border lines (═, ║) are single characters, just styled differently
	borderWidth := 1

	config := border.Config{
		Style: borderStyle,
		Color: bc,
		Label: bl,
	}
	renderer := border.WithConfig(config)

	// Calculate content area: subtract border padding only
	// LabelPadding is only for visual balance in the border line, NOT for constraining content area
	// The label is rendered inside the top border line, doesn't reduce horizontal content space
	contentWidth := box.Width - (borderWidth * 2)
	contentHeight := box.Height - (borderWidth * 2)
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

	if paintChildren {
		for _, childBox := range box.Children {
			if err := e.paintBox(childBox, buffer); err != nil && e.debug {
				log.PaintLogger.IfEnabled().Debug("[paintBorderedBox] error: %v", err)
			}
		}
	}
}

// paintModalBackdropBox draws modal backdrop (PaintableBox version)
// 智能版本：按行遍历，对空白区域用空格填充，对有内容区域用 SetString 灰染（保留内容并改色）
// 关键优化：跨单元格时跳过延续单元格，避免破坏中文连续性
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
	dimmedStyle := style.Style{FG: dimmedFG, BG: dimmedBG}

	// 对每个需要进行灰化的区域，按行处理
	//
	// ┌─────────────────────────────────────────────────────────────────────┐
	// 1. 上方区域 (0, 0, width, modalY)                                    │
	// │                    ↓                                                 │
	// │              ┌────────────────────────────────────────────────────┐
	// │              2.│          Modal 内容区域          │    │
	// │              └────────────────────────────────────────────────────┘│
	// │                    ↓                                                 │
	// │ 3. 下方区域 (0, modalY+modalHeight, width, height-modalY-modalHeight)│
	// └─────────────────────────────────────────────────────────────────────┘
	// 4. 左侧区域 (0, modalY, modalX, modalHeight)
	// 5. 右侧区域 (modalX+modalWidth, modalY, width-modalX-modalWidth, modalHeight)

	// 辅助函数：对指定区域进行灰化处理
	dimRegion := func(startX, startY, endX, endY int) {
		if startX >= endX || startY >= endY {
			return
		}
		for y := startY; y < endY; y++ {
			x := startX
			for x < endX {
				cell := buffer.GetContent(x, y)
				// 如果是延续单元格，跳过（属于主字符的一部分）
				if cell.IsContinuation {
					x++
					continue
				}

				// 空白或空格：用灰色空格填充
				if cell.Cluster == "" || cell.Cluster == " " {
					buffer.SetCell(x, y, ' ', dimmedStyle)
					x++
				} else {
					// 有内容：用 SetString 保留内容并改色
					buffer.SetString(x, y, cell.Cluster, dimmedStyle)
					// 跳过字符的所有延续单元格
					x += cell.Width
				}
			}
		}
	}

	// 矩形 1: 上方区域
	dimRegion(0, 0, width, modalY)

	// 矩形 2: 下方区域
	bottomY := modalY + modalHeight
	dimRegion(0, bottomY, width, height)

	// 矩形 3: 左侧区域
	dimRegion(0, modalY, modalX, modalY+modalHeight)

	// 矩形 4: 右侧区域
	rightX := modalX + modalWidth
	dimRegion(rightX, modalY, width, modalY+modalHeight)
}

func (e *PaintEngine) findBackdropModalBox(root *paint.PaintableBox) *paint.PaintableBox {
	if root == nil {
		return nil
	}

	var best *paint.PaintableBox
	var walk func(box *paint.PaintableBox)
	walk = func(box *paint.PaintableBox) {
		if box == nil {
			return
		}
		if e.isBackdropTargetBox(box) {
			if best == nil || box.ZIndex > best.ZIndex || (box.ZIndex == best.ZIndex && box.Width*box.Height > best.Width*best.Height) {
				best = box
			}
		}
		for _, child := range box.Children {
			walk(child)
		}
	}

	walk(root)
	return best
}

func (e *PaintEngine) findBackdropModalBoxInLayer(boxes []*paint.PaintableBox) *paint.PaintableBox {
	var best *paint.PaintableBox
	for _, box := range boxes {
		if !e.isBackdropTargetBox(box) {
			continue
		}
		if best == nil || box.ZIndex > best.ZIndex || (box.ZIndex == best.ZIndex && box.Width*box.Height > best.Width*best.Height) {
			best = box
		}
	}
	return best
}

func (e *PaintEngine) isBackdropTargetBox(box *paint.PaintableBox) bool {
	if box == nil || box.Node == nil {
		return false
	}
	if box.Width <= 0 || box.Height <= 0 {
		return false
	}
	return !e.isPortalRootBox(box)
}

func (e *PaintEngine) isPortalRootBox(box *paint.PaintableBox) bool {
	if box == nil || box.Node == nil {
		return false
	}

	switch node := box.Node.(type) {
	case *FiberPaintableNode:
		return hasPortalRootID(node.fiber.Props)
	case *VNodePaintableNode:
		return hasPortalRootID(node.vnode.Props())
	default:
		return false
	}
}

func hasPortalRootID(props rtui.Props) bool {
	if props == nil {
		return false
	}
	portalRootID, ok := props["portalRootId"].(string)
	return ok && portalRootID != ""
}

// Fill fills a rectangular region of the buffer with a specific character and style
func (e *PaintEngine) Fill(buffer *paint.Buffer, bounds runtime.Box, ch rune, s style.Style) {
	if bounds.Width <= 0 || bounds.Height <= 0 {
		return
	}

	// 填充矩形区域
	for y := bounds.Y; y < bounds.Y+bounds.Height; y++ {
		for x := bounds.X; x < bounds.X+bounds.Width; x++ {
			buffer.SetCell(x, y, ch, s)
		}
	}
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
// Multi-Layer Rendering (New API - PaintableBox based)
// =============================================================================

// PaintPaintableLayouts renders multiple PaintableLayouts in order (from lowest to highest).
// This is the new decoupled API that operates on paint.PaintableLayouts.
func (e *PaintEngine) PaintPaintableLayouts(
	layouts paint.PaintableLayouts,
	buffer *paint.Buffer,
) error {
	var modalBackdropBox *paint.PaintableBox
	if modalLayout, ok := layouts[paint.RenderLayerModal]; ok {
		modalBackdropBox = e.findBackdropModalBox(modalLayout.Root)
	}
	hasModal := modalBackdropBox != nil
	hadModal := e.lastHadModal

	if hasModal != hadModal {
		e.forceFullRender = true
	}
	e.lastHadModal = hasModal

	renderOrder := []paint.RenderLayer{
		paint.RenderLayerBase,
		paint.RenderLayerOverlay,
		paint.RenderLayerModal,
		paint.RenderLayerTooltip,
		paint.RenderLayerInspector,
	}

	for _, l := range renderOrder {
		hasLayer := false
		var currentBounds runtime.Box = runtime.Box{}
		if layout, ok := layouts[l]; ok && layout.Root != nil {
			target := layout.Root
			if l == paint.RenderLayerModal {
				target = modalBackdropBox
			}
			if target != nil {
				hasLayer = true
				currentBounds = runtime.Box{
					X:      target.X,
					Y:      target.Y,
					Width:  target.Width,
					Height: target.Height,
				}
			}
		}
		hadLayer := e.lastLayersPresent[rtui.Layer(l)]
		prevBounds := e.lastLayerBounds[rtui.Layer(l)]

		if hadLayer && !hasLayer {
			log.PaintLogger.IfEnabled().Debug("[PaintPaintableLayouts] Layer %s disappeared, clearing region", l.String())
			e.clearRegion(prevBounds, buffer)
			e.forceFullRender = true
		}

		if hasLayer && hadLayer && currentBounds != prevBounds {
			e.forceFullRender = true
		}

		e.lastLayersPresent[rtui.Layer(l)] = hasLayer
		if hasLayer {
			e.lastLayerBounds[rtui.Layer(l)] = currentBounds
		} else {
			delete(e.lastLayerBounds, rtui.Layer(l))
		}
	}

	for _, l := range renderOrder {
		layout, ok := layouts[l]
		if !ok || layout.Root == nil {
			continue
		}

		if e.debug {
			log.PaintLogger.IfEnabled().Debug("[PaintPaintableLayouts] Rendering layer: %s", l.String())
		}

		if err := e.PaintLayout(layout, buffer); err != nil {
			return fmt.Errorf("error painting layer %s: %w", l.String(), err)
		}

		if l == paint.RenderLayerModal && modalBackdropBox != nil {
			e.paintModalBackdropBox(modalBackdropBox, buffer)
		}
	}

	return nil
}

// PaintPaintablePlanes paints PaintablePlanes to buffer directly.
// This is the new decoupled API that operates on paint.PaintablePlanes.
func (e *PaintEngine) PaintPaintablePlanes(
	planes *paint.PaintablePlanes,
	buffer *paint.Buffer,
) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if planes == nil {
		return nil
	}
	if buffer == nil {
		return nil
	}

	log.PaintLogger.IfEnabled().Debug("[PaintEngine.PaintPaintablePlanes] START: boxes=%d", planes.CountBoxes())

	modalBackdropBox := e.findBackdropModalBoxInLayer(planes.GetLayer(paint.RenderLayerModal))
	hasModal := modalBackdropBox != nil
	hadModal := e.lastHadModal
	if hasModal != hadModal {
		e.forceFullRender = true
	}
	e.lastHadModal = hasModal
	e.beginPaintFrame(buffer)

	// Phase 1: Collect all boxes in current frame for tracking
	clear(e.currentFrameBoxes)
	for _, layer := range planes.GetRenderOrder() {
		boxes := planes.GetLayer(layer)
		for _, box := range boxes {
			e.collectFrameBoxes(box)
		}
	}

	// Phase 2: Clear outdated painted regions (smart buffer clearing)
	e.clearOutdatedRegions(buffer)

	paintChildren := !e.paintablePlanesContainDescendants(planes)

	// Phase 3: Paint all boxes in current frame
	for _, layer := range planes.GetRenderOrder() {
		boxes := planes.GetLayer(layer)
		if len(boxes) == 0 {
			continue
		}

		log.PaintLogger.IfEnabled().Debug("[PaintEngine.PaintPaintablePlanes] Layer %s: %d boxes", layer.String(), len(boxes))

		for _, box := range boxes {
			var err error
			if paintChildren {
				err = e.paintBox(box, buffer)
			} else {
				err = e.paintBoxOnly(box, buffer)
			}
			if err != nil {
				return fmt.Errorf("error painting box in layer %s: %w", layer.String(), err)
			}
		}

		if layer == paint.RenderLayerModal && modalBackdropBox != nil {
			e.paintModalBackdropBox(modalBackdropBox, buffer)
		}
	}

	// Phase 4: Update tracking maps for next frame
	reusableFrameBoxes := e.previousFrameBoxes
	e.previousFrameBoxes = e.currentFrameBoxes
	e.currentFrameBoxes = reusableFrameBoxes
	clear(e.currentFrameBoxes)

	log.PaintLogger.IfEnabled().Debug("[PaintEngine.PaintPaintablePlanes] END")
	return nil
}

func (e *PaintEngine) paintablePlanesContainDescendants(planes *paint.PaintablePlanes) bool {
	if planes == nil {
		return false
	}

	boxSet := make(map[*paint.PaintableBox]struct{}, planes.CountBoxes())
	for _, layer := range planes.GetRenderOrder() {
		for _, box := range planes.GetLayer(layer) {
			if box != nil {
				boxSet[box] = struct{}{}
			}
		}
	}
	if len(boxSet) == 0 {
		return false
	}

	for box := range boxSet {
		for _, child := range box.Children {
			if child == nil {
				continue
			}
			if _, ok := boxSet[child]; !ok {
				return false
			}
		}
	}
	return true
}

// clearOutdatedRegions smartly clears regions that need refreshing
// It compares previous frame boxes with current frame boxes and clears:
// - Boxes that were in previous frame but are gone now
// - Boxes that have moved to different positions
func (e *PaintEngine) clearOutdatedRegions(buffer *paint.Buffer) {
	// Check each box from previous frame
	for id, prevBounds := range e.previousFrameBoxes {
		currentBounds, exists := e.currentFrameBoxes[id]

		if !exists {
			// Case 1: Box removed - clear its previous region
			log.PaintLogger.IfEnabled().Debug("[clearOutdatedRegions] Box %s removed, clearing %v", id, prevBounds)
			e.clearRegion(prevBounds, buffer)
		} else if prevBounds != currentBounds {
			// Case 2: Box moved - clear both old and new positions
			log.PaintLogger.IfEnabled().Debug("[clearOutdatedRegions] Box %s moved from %v to %v",
				id, prevBounds, currentBounds)
			e.clearRegion(prevBounds, buffer) // Clear old position
		}
	}
}

func (e *PaintEngine) collectFrameBoxes(box *paint.PaintableBox) {
	if box == nil {
		return
	}
	key := e.boxTrackingKey(box)
	if key != "" {
		e.currentFrameBoxes[key] = runtime.Box{
			X:      box.X,
			Y:      box.Y,
			Width:  box.Width,
			Height: box.Height,
		}
	}
	for _, child := range box.Children {
		e.collectFrameBoxes(child)
	}
}

func (e *PaintEngine) boxTrackingKey(box *paint.PaintableBox) string {
	if box == nil || box.Node == nil {
		return ""
	}
	return fmt.Sprintf("%s@%d,%d,%d,%d", box.Node.ID(), box.X, box.Y, box.Width, box.Height)
}

func (e *PaintEngine) boxCacheKey(box *paint.PaintableBox) string {
	if box == nil || box.Node == nil {
		return ""
	}
	return fmt.Sprintf("%s@%d,%d,%d,%d", box.Node.ID(), box.X, box.Y, box.Width, box.Height)
}

// PrintBoxTrackingLogs prints current and previous frame box tracking for debugging
func (e *PaintEngine) PrintBoxTrackingLogs() {
	log.PaintLogger.IfEnabled().Debug("=== Frame Box Tracking ===")
	log.PaintLogger.IfEnabled().Debug("Previous frame boxes:")
	for id, bounds := range e.previousFrameBoxes {
		log.PaintLogger.IfEnabled().Debug("  %s: %v", id, bounds)
	}
	log.PaintLogger.IfEnabled().Debug("Current frame boxes:")
	for id, bounds := range e.currentFrameBoxes {
		log.PaintLogger.IfEnabled().Debug("  %s: %v", id, bounds)
	}
	log.PaintLogger.IfEnabled().Debug("=========================")
}

// =============================================================================
// Cache Helper Functions
// =============================================================================

// boxRect converts a PaintableBox bounds to a layout.Rect for caching
func (e *PaintEngine) boxRect(box *paint.PaintableBox) layout.Rect {
	return layout.Rect{
		X:      box.X,
		Y:      box.Y,
		Width:  box.Width,
		Height: box.Height,
	}
}
