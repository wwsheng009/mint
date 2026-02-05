package runtime

// Core types for Yao TUI Runtime v1

// Infinity represents an unbounded constraint value
const Infinity = 1<<30 - 1

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
