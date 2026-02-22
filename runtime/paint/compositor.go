package paint

import (
	"bytes"
	"sort"

	"github.com/wwsheng009/mint/runtime/style"
)

// Compositor manages multiple layers and composites them
type Compositor struct {
	layers      []*Layer
	width       int
	height      int
	batch       *CommandBatch
	styleState  *StyleStateMachine
}

// NewCompositor creates a new compositor
func NewCompositor(width, height int) *Compositor {
	return &Compositor{
		layers:     make([]*Layer, 0, 4),
		width:      width,
		height:     height,
		batch:      NewCommandBatch(),
		styleState: NewStyleStateMachine(),
	}
}

// AddLayer adds a layer to the compositor
func (c *Compositor) AddLayer(layer *Layer) {
	c.layers = append(c.layers, layer)
	c.sortLayers()
}

// RemoveLayer removes a layer by ID
func (c *Compositor) RemoveLayer(id string) bool {
	for i, layer := range c.layers {
		if layer.ID == id {
			c.layers = append(c.layers[:i], c.layers[i+1:]...)
			return true
		}
	}
	return false
}

// GetLayer returns a layer by ID
func (c *Compositor) GetLayer(id string) *Layer {
	for _, layer := range c.layers {
		if layer.ID == id {
			return layer
		}
	}
	return nil
}

// GetLayerByType returns the first layer of the given type
func (c *Compositor) GetLayerByType(layerType LayerType) *Layer {
	for _, layer := range c.layers {
		if layer.Type == layerType {
			return layer
		}
	}
	return nil
}

// sortLayers sorts layers by Z-index
func (c *Compositor) sortLayers() {
	sort.Slice(c.layers, func(i, j int) bool {
		return c.layers[i].ZIndex < c.layers[j].ZIndex
	})
}

// RenderDirty renders only dirty layers and returns the composited output
func (c *Compositor) RenderDirty() string {
	var buf bytes.Buffer

	for _, layer := range c.layers {
		// Skip layers that are not dirty, not enabled, or not visible
		if !layer.IsDirty() || !layer.Enabled || !layer.Visible {
			continue
		}

		// Skip layers completely outside the viewport (culling)
		if !c.layerInViewport(layer) {
			continue
		}

		// Output layer content
		buf.WriteString(c.renderLayer(layer))

		layer.ClearDirty()
	}

	return buf.String()
}

// RenderDirtyRect renders only dirty layers within a clip region
// This is useful for partial screen updates
func (c *Compositor) RenderDirtyRect(clipRect Rect) string {
	var buf bytes.Buffer

	// Clip the rect to viewport
	clipped := c.clipRectToViewport(clipRect)

	for _, layer := range c.layers {
		// Skip layers not matching criteria
		if !layer.IsDirty() || !layer.Enabled || !layer.Visible {
			continue
		}

		// Skip layers completely outside the viewport (culling)
		if !c.layerInViewport(layer) {
			continue
		}

		// Skip layers not intersecting with clip region
		if !rectIntersect(layer.Rect, clipped) {
			continue
		}

		// Output only the clipped region
		buf.WriteString(c.renderLayerClipped(layer, clipped))

		layer.ClearDirty()
	}

	return buf.String()
}

// clipRectToViewport clips a rectangle to the compositor's viewport
func (c *Compositor) clipRectToViewport(rect Rect) Rect {
	// Clamp X and Y
	cx := maxInt(0, rect.X)
	cy := maxInt(0, rect.Y)

	// Clamp right and bottom
	right := minInt(rect.X+rect.Width, c.width)
	bottom := minInt(rect.Y+rect.Height, c.height)

	// Calculate width and height
	width := right - cx
	height := bottom - cy

	// Ensure non-negative dimensions
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}

	return Rect{X: cx, Y: cy, Width: width, Height: height}
}

// renderLayerClipped renders a single layer within a clipped region
func (c *Compositor) renderLayerClipped(layer *Layer, clipRect Rect) string {
	c.batch.Clear()
	c.styleState.Reset()

	buf := layer.Buffer

	// Calculate intersection between layer and clip region
	intersection := rectIntersection(layer.Rect, clipRect)

	// Convert world coordinates to layer-local coordinates
	startX := intersection.X - layer.Rect.X
	startY := intersection.Y - layer.Rect.Y
	endX := startX + intersection.Width
	endY := startY + intersection.Height

	// Iterate only over the clipped region
	for y := startY; y < endY; y++ {
		for x := startX; x < endX; x++ {
			// Skip if position is out of layer bounds
			if x < 0 || y < 0 || x >= buf.Width || y >= buf.Height {
				continue
			}

			cell := buf.Cells[y][x]

			// Skip continuation cells and empty cells
			if cell.IsContinuation || cell.Cluster == "" {
				continue
			}

			c.batch.Add(
				layer.Rect.X+x,
				layer.Rect.Y+y,
				cell.Cluster,
				cell.Style,
			)
		}
	}

	return c.batch.Flush()
}

// renderLayer renders a single layer with region scrolling
func (c *Compositor) renderLayer(layer *Layer) string {
	var output string

	// For stream layers, use scroll optimization
	if layer.Type == LayerStream {
		// Set scroll region
		output = "\x1b[" + itoa(layer.Rect.Y+1) + ";" +
			itoa(layer.Rect.Y+layer.Rect.Height) + "r"
	}

	// Output layer buffer content
	output += c.bufferToString(layer)

	// Reset scroll region
	if layer.Type == LayerStream {
		output += "\x1b[r"
	}

	return output
}

// bufferToString converts a layer's buffer to ANSI output
func (c *Compositor) bufferToString(layer *Layer) string {
	c.batch.Clear()
	c.styleState.Reset()

	buf := layer.Buffer
	for y := 0; y < buf.Height; y++ {
		for x := 0; x < buf.Width; x++ {
			cell := buf.Cells[y][x]
			// 跳过延续单元格和空单元格
			if cell.IsContinuation || cell.Cluster == "" {
				continue
			}
			c.batch.Add(
				layer.Rect.X+x,
				layer.Rect.Y+y,
				cell.Cluster,
				cell.Style,
			)
		}
	}

	return c.batch.Flush()
}

// layerInViewport checks if a layer has any part visible within the compositor viewport
func (c *Compositor) layerInViewport(layer *Layer) bool {
	// Viewport: [0, 0, c.width, c.height]
	// Layer rect: [layer.Rect.X, layer.Rect.Y, layer.Rect.Width, layer.Rect.Height]

	// Check if layer is completely to the right
	if layer.Rect.X >= c.width {
		return false
	}

	// Check if layer is completely below
	if layer.Rect.Y >= c.height {
		return false
	}

	// Check if layer is completely to the left
	if layer.Rect.X+layer.Rect.Width <= 0 {
		return false
	}

	// Check if layer is completely above
	if layer.Rect.Y+layer.Rect.Height <= 0 {
		return false
	}

	// At least part of the layer is visible
	return true
}

// IsLayerVisible checks if a layer is currently visible (enabled + visible + in viewport)
func (c *Compositor) IsLayerVisible(layer *Layer) bool {
	if !layer.Enabled || !layer.Visible {
		return false
	}
	return c.layerInViewport(layer)
}

// Composite creates a composite buffer from all layers
func (c *Compositor) Composite() *Buffer {
	buffer := NewBuffer(c.width, c.height)

	for _, layer := range c.layers {
		if !layer.Enabled || !layer.Visible {
			continue
		}

		c.blitLayer(buffer, layer)
	}

	return buffer
}

// blitLayer blits a layer onto the composite buffer with clipping optimization
func (c *Compositor) blitLayer(dst *Buffer, src *Layer) {
	// Calculate clipping bounds to minimize iteration
	startX := maxInt(0, -src.Rect.X)
	startY := maxInt(0, -src.Rect.Y)
	endX := minInt(src.Rect.Width, dst.Width-src.Rect.X)
	endY := minInt(src.Rect.Height, dst.Height-src.Rect.Y)

	// Early exit if layer is completely outside
	if startX >= endX || startY >= endY {
		return
	}

	for y := startY; y < endY; y++ {
		for x := startX; x < endX; x++ {
			// Skip if layer buffer is too small
			if x >= src.Buffer.Width || y >= src.Buffer.Height {
				continue
			}

			srcX := src.Rect.X + x
			srcY := src.Rect.Y + y

			// Skip if destination is out of bounds
			if srcX >= dst.Width || srcY >= dst.Height || srcX < 0 || srcY < 0 {
				continue
			}

			cell := src.Buffer.Cells[y][x]
			if cell.Cluster != "" {
				dst.Cells[srcY][srcX] = cell
			}
		}
	}
}

// MarkAllDirty marks all layers as dirty
func (c *Compositor) MarkAllDirty() {
	for _, layer := range c.layers {
		layer.MarkDirty()
	}
}

// MarkTypeDirty marks all layers of a given type as dirty
func (c *Compositor) MarkTypeDirty(layerType LayerType) {
	for _, layer := range c.layers {
		if layer.Type == layerType {
			layer.MarkDirty()
		}
	}
}

// Resize handles window resize
func (c *Compositor) Resize(width, height int) {
	c.width = width
	c.height = height

	for _, layer := range c.layers {
		if layer.Rect.Width > width {
			layer.Rect.Width = width
		}
		if layer.Rect.Height > height {
			layer.Rect.Height = height
		}
		layer.SetRect(layer.Rect)
	}
}

// GetLayerCount returns the number of layers
func (c *Compositor) GetLayerCount() int {
	return len(c.layers)
}

// GetLayers returns all layers
func (c *Compositor) GetLayers() []*Layer {
	return c.layers
}

// Clear clears all layers
func (c *Compositor) Clear() {
	for _, layer := range c.layers {
		layer.Clear()
	}
}

// ClearType clears all layers of a given type
func (c *Compositor) ClearType(layerType LayerType) {
	for _, layer := range c.layers {
		if layer.Type == layerType {
			layer.Clear()
		}
	}
}

// Fill fills a layer with a character and style
func (c *Compositor) Fill(id string, char rune, st style.Style) bool {
	layer := c.GetLayer(id)
	if layer == nil {
		return false
	}
	layer.Fill(char, st)
	return true
}

// =============================================================================
// Rectangle Helper Functions
// =============================================================================

// rectIntersect returns true if two rectangles intersect
func rectIntersect(r1, r2 Rect) bool {
	return !(r1.X >= r2.X+r2.Width ||
		r1.X+r1.Width <= r2.X ||
		r1.Y >= r2.Y+r2.Height ||
		r1.Y+r1.Height <= r2.Y)
}

// rectIntersection returns the intersection of two rectangles
func rectIntersection(r1, r2 Rect) Rect {
	x := maxInt(r1.X, r2.X)
	y := maxInt(r1.Y, r2.Y)
	width := minInt(r1.X+r1.Width, r2.X+r2.Width) - x
	height := minInt(r1.Y+r1.Height, r2.Y+r2.Height) - y

	// Handle non-intersecting case
	if width <= 0 || height <= 0 {
		return Rect{X: x, Y: y, Width: 0, Height: 0}
	}

	return Rect{X: x, Y: y, Width: width, Height: height}
}
