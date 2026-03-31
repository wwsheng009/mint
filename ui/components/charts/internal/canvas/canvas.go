package canvas

import "strings"

// RuneCanvas is a minimal character surface for chart rasterization.
type RuneCanvas struct {
	width  int
	height int
	cells  [][]rune
}

// NewRuneCanvas creates a new rune canvas filled with the provided rune.
func NewRuneCanvas(width, height int, fill rune) *RuneCanvas {
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}

	cells := make([][]rune, height)
	for y := 0; y < height; y++ {
		cells[y] = []rune(strings.Repeat(string(fill), width))
	}

	return &RuneCanvas{
		width:  width,
		height: height,
		cells:  cells,
	}
}

// Width returns the number of columns in the canvas.
func (c *RuneCanvas) Width() int {
	return c.width
}

// Height returns the number of rows in the canvas.
func (c *RuneCanvas) Height() int {
	return c.height
}

// Set writes a rune at the requested position. It returns false for out-of-range coordinates.
func (c *RuneCanvas) Set(x, y int, value rune) bool {
	if !c.InBounds(x, y) {
		return false
	}
	c.cells[y][x] = value
	return true
}

// Get returns the rune at the requested position. For out-of-range coordinates it returns zero.
func (c *RuneCanvas) Get(x, y int) rune {
	if !c.InBounds(x, y) {
		return 0
	}
	return c.cells[y][x]
}

// InBounds reports whether the coordinate can be addressed on this canvas.
func (c *RuneCanvas) InBounds(x, y int) bool {
	return x >= 0 && x < c.width && y >= 0 && y < c.height
}

// Rows returns the current canvas content as text rows.
func (c *RuneCanvas) Rows() []string {
	rows := make([]string, len(c.cells))
	for i, row := range c.cells {
		rows[i] = string(row)
	}
	return rows
}
