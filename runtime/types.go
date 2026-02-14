package runtime

import "fmt"

// Core types for Yao TUI Runtime v1

// Infinity represents an unbounded constraint value
const Infinity = 1 << 30

// BoxConstraints defines the min/max constraints for layout
type BoxConstraints struct {
	MinWidth, MaxWidth   int
	MinHeight, MaxHeight int
}

// NewBoxConstraints creates a new BoxConstraints
func NewBoxConstraints(minWidth, maxWidth, minHeight, maxHeight int) BoxConstraints {
	return BoxConstraints{
		MinWidth:  minWidth,
		MaxWidth:  maxWidth,
		MinHeight: minHeight,
		MaxHeight: maxHeight,
	}
}

// IsTight returns true if width and height are both fixed
func (bc BoxConstraints) IsTight() bool {
	return bc.MinWidth == bc.MaxWidth && bc.MinHeight == bc.MaxHeight
}

// Constrain clamps a width and height within the constraints
func (bc BoxConstraints) Constrain(width, height int) (int, int) {
	w := clamp(width, bc.MinWidth, bc.MaxWidth)
	h := clamp(height, bc.MinHeight, bc.MaxHeight)
	return w, h
}

// Loosen returns a new BoxConstraints with min values set to 0
func (bc BoxConstraints) Loosen() BoxConstraints {
	return BoxConstraints{
		MinWidth:  0,
		MaxWidth:  bc.MaxWidth,
		MinHeight: 0,
		MaxHeight: bc.MaxHeight,
	}
}

// TightConstraints creates constraints with fixed width and height
func TightConstraints(width, height int) BoxConstraints {
	return BoxConstraints{
		MinWidth:  width,
		MaxWidth:  width,
		MinHeight: height,
		MaxHeight: height,
	}
}

// LooseConstraints creates constraints with only maximum bounds
func LooseConstraints(maxWidth, maxHeight int) BoxConstraints {
	return BoxConstraints{
		MinWidth:  0,
		MaxWidth:  maxWidth,
		MinHeight: 0,
		MaxHeight: maxHeight,
	}
}

// UnboundedConstraints creates constraints with no bounds (except Infinity)
func UnboundedConstraints() BoxConstraints {
	return BoxConstraints{
		MinWidth:  0,
		MaxWidth:  Infinity,
		MinHeight: 0,
		MaxHeight: Infinity,
	}
}

// IsBounded returns true if the constraints have an upper bound
func (bc BoxConstraints) IsBounded() bool {
	return bc.MaxWidth < Infinity || bc.MaxHeight < Infinity
}

// HasBoundedWidth returns true if MaxWidth is not Infinity
func (bc BoxConstraints) HasBoundedWidth() bool {
	return bc.MaxWidth < Infinity
}

// HasBoundedHeight returns true if MaxHeight is not Infinity
func (bc BoxConstraints) HasBoundedHeight() bool {
	return bc.MaxHeight < Infinity
}

// SubtractPadding returns new constraints with padding subtracted (only if bounded)
func (bc BoxConstraints) SubtractPadding(horizontal, vertical int) BoxConstraints {
	result := bc
	// Subtract padding from MaxWidth (if bounded)
	if bc.HasBoundedWidth() && bc.MaxWidth > horizontal {
		result.MaxWidth = bc.MaxWidth - horizontal
		if result.MaxWidth < 0 {
			result.MaxWidth = 0
		}
	}
	// Also subtract from MinWidth (but don't go below 0)
	if bc.MinWidth > horizontal {
		result.MinWidth = bc.MinWidth - horizontal
	} else {
		result.MinWidth = 0
	}
	// Subtract padding from MaxHeight (if bounded)
	if bc.HasBoundedHeight() && bc.MaxHeight > vertical {
		result.MaxHeight = bc.MaxHeight - vertical
		if result.MaxHeight < 0 {
			result.MaxHeight = 0
		}
	}
	// Also subtract from MinHeight (but don't go below 0)
	if bc.MinHeight > vertical {
		result.MinHeight = bc.MinHeight - vertical
	} else {
		result.MinHeight = 0
	}
	// Ensure MinWidth <= MaxWidth and MinHeight <= MaxHeight
	if result.MinWidth > result.MaxWidth {
		result.MinWidth = result.MaxWidth
	}
	if result.MinHeight > result.MaxHeight {
		result.MinHeight = result.MaxHeight
	}
	return result
}

// ConstrainWidth clamps a width within the constraints
func (bc BoxConstraints) ConstrainWidth(width int) int {
	if width < bc.MinWidth {
		return bc.MinWidth
	}
	if width > bc.MaxWidth {
		return bc.MaxWidth
	}
	return width
}

// ConstrainHeight clamps a height within the constraints
func (bc BoxConstraints) ConstrainHeight(height int) int {
	if height < bc.MinHeight {
		return bc.MinHeight
	}
	if height > bc.MaxHeight {
		return bc.MaxHeight
	}
	return height
}

// clamp clamps a value between min and max
func clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// Size represents a 2D size
type Size struct {
	Width  int
	Height int
}

// Box represents a rectangular area with position and size
type Box struct {
	X      int
	Y      int
	Width  int
	Height int
}

// Empty returns true if the box has zero area
func (b Box) Empty() bool {
	return b.Width <= 0 || b.Height <= 0
}

// Contains returns true if the point (x, y) is inside the box
func (b Box) Contains(x, y int) bool {
	return x >= b.X && x < b.X+b.Width && y >= b.Y && y < b.Y+b.Height
}

// Intersects returns true if two boxes intersect
func (b Box) Intersects(other Box) bool {
	return b.X < other.X+other.Width &&
		b.X+b.Width > other.X &&
		b.Y < other.Y+other.Height &&
		b.Y+b.Height > other.Y
}

// String returns a string representation of the box
func (b Box) String() string {
	return fmt.Sprintf("(%d,%d,%d×%d)", b.X, b.Y, b.Width, b.Height)
}

// GetX returns the X coordinate
func (b Box) GetX() int {
	return b.X
}

// GetY returns the Y coordinate
func (b Box) GetY() int {
	return b.Y
}

// GetWidth returns the width
func (b Box) GetWidth() int {
	return b.Width
}

// GetHeight returns the height
func (b Box) GetHeight() int {
	return b.Height
}

// ConstraintsAlias is an alias for backward compatibility with legacy code
// Legacy code uses Constraints with MaxW and MaxH fields
type Constraints = BoxConstraints

// NewConstraints creates a Constraints with only MaxW and MaxH set
func NewConstraints(maxW, maxH int) Constraints {
	return BoxConstraints{
		MinWidth:  0,
		MaxWidth:  maxW,
		MinHeight: 0,
		MaxHeight: maxH,
	}
}

// IsInfinite checks if constraint is infinite (-1 means infinite)
func (c Constraints) HasInfiniteWidth() bool {
	return c.MaxWidth < 0
}

func (c Constraints) HasInfiniteHeight() bool {
	return c.MaxHeight < 0
}

// PositionType defines the positioning scheme
type PositionType string

const (
	PositionRelative PositionType = "relative"
	PositionAbsolute PositionType = "absolute"
)

// Position defines absolute positioning properties
type Position struct {
	Type   PositionType
	Top    *int // nil means auto
	Left   *int
	Right  *int
	Bottom *int
}

// TextAlign defines horizontal text alignment within a container
type TextAlign int

const (
	TextAlignLeft TextAlign = iota
	TextAlignCenter
	TextAlignRight
)

// NewPosition creates a new Position with relative positioning (default)
func NewPosition() Position {
	return Position{
		Type: PositionRelative,
	}
}

// SetAbsolute sets absolute positioning with offsets
func (p *Position) SetAbsolute(top, left, right, bottom *int) {
	p.Type = PositionAbsolute
	p.Top = top
	p.Left = left
	p.Right = right
	p.Bottom = bottom
}
