package overlayposition

// Placement describes where an overlay should appear relative to its anchor.
type Placement int

const (
	PlacementTop Placement = iota
	PlacementTopLeft
	PlacementTopRight
	PlacementBottom
	PlacementBottomLeft
	PlacementBottomRight
	PlacementLeft
	PlacementLeftTop
	PlacementLeftBottom
	PlacementRight
	PlacementRightTop
	PlacementRightBottom
)

// Rect is an integer bounds box.
type Rect struct {
	X      int
	Y      int
	Width  int
	Height int
}

// Size is an integer width/height pair.
type Size struct {
	Width  int
	Height int
}

// RectFromBounds converts an [x, y, width, height] tuple into a Rect.
func RectFromBounds(bounds [4]int) Rect {
	return Rect{
		X:      bounds[0],
		Y:      bounds[1],
		Width:  bounds[2],
		Height: bounds[3],
	}
}

// Config describes a placement resolution request.
type Config struct {
	Anchor     Rect
	Overlay    Size
	Viewport   Size
	Candidates []Placement
	Gap        int
}

// Result is the resolved overlay position.
type Result struct {
	X         int
	Y         int
	Placement Placement
	Clamped   bool
}

// Resolve returns the first candidate that fully fits the viewport.
// When none fit, the first candidate is clamped into the viewport.
func Resolve(cfg Config) Result {
	if cfg.Gap < 0 {
		cfg.Gap = 0
	}
	if len(cfg.Candidates) == 0 {
		cfg.Candidates = []Placement{PlacementTop}
	}

	for _, placement := range cfg.Candidates {
		x, y := Coordinates(cfg.Anchor, cfg.Overlay, placement, cfg.Gap)
		if fitsViewport(x, y, cfg.Overlay, cfg.Viewport) {
			return Result{
				X:         x,
				Y:         y,
				Placement: placement,
			}
		}
	}

	x, y := Coordinates(cfg.Anchor, cfg.Overlay, cfg.Candidates[0], cfg.Gap)
	clampedX, clampedY := clampToViewport(x, y, cfg.Overlay, cfg.Viewport)
	return Result{
		X:         clampedX,
		Y:         clampedY,
		Placement: cfg.Candidates[0],
		Clamped:   clampedX != x || clampedY != y,
	}
}

// Coordinates returns the raw position for a placement without viewport clamping.
func Coordinates(anchor Rect, overlay Size, placement Placement, gap int) (x, y int) {
	if gap < 0 {
		gap = 0
	}

	centerX := anchor.X + anchor.Width/2 - overlay.Width/2
	topY := anchor.Y - overlay.Height - gap
	bottomY := anchor.Y + anchor.Height + gap
	leftX := anchor.X - overlay.Width - gap
	rightX := anchor.X + anchor.Width + gap
	centerY := anchor.Y + anchor.Height/2 - overlay.Height/2
	topAlignedY := anchor.Y
	bottomAlignedY := anchor.Y + anchor.Height - overlay.Height
	leftAlignedX := anchor.X
	rightAlignedX := anchor.X + anchor.Width - overlay.Width

	switch placement {
	case PlacementTop:
		return centerX, topY
	case PlacementTopLeft:
		return leftAlignedX, topY
	case PlacementTopRight:
		return rightAlignedX, topY
	case PlacementBottom:
		return centerX, bottomY
	case PlacementBottomLeft:
		return leftAlignedX, bottomY
	case PlacementBottomRight:
		return rightAlignedX, bottomY
	case PlacementLeft:
		return leftX, centerY
	case PlacementLeftTop:
		return leftX, topAlignedY
	case PlacementLeftBottom:
		return leftX, bottomAlignedY
	case PlacementRight:
		return rightX, centerY
	case PlacementRightTop:
		return rightX, topAlignedY
	case PlacementRightBottom:
		return rightX, bottomAlignedY
	default:
		return centerX, topY
	}
}

// PointerX returns the anchor-centered X coordinate clamped within an overlay border.
func PointerX(anchor Rect, overlayX, overlayWidth int) int {
	x := anchor.X
	if anchor.Width > 0 {
		x = anchor.X + anchor.Width/2
	}
	if overlayWidth <= 0 {
		return overlayX
	}
	if overlayWidth <= 2 {
		return clamp(x, overlayX, overlayX+maxInt(0, overlayWidth-1))
	}
	return clamp(x, overlayX+1, overlayX+overlayWidth-2)
}

func fitsViewport(x, y int, overlay Size, viewport Size) bool {
	if viewport.Width <= 0 || viewport.Height <= 0 {
		return true
	}
	return x >= 0 &&
		y >= 0 &&
		x+overlay.Width <= viewport.Width &&
		y+overlay.Height <= viewport.Height
}

func clampToViewport(x, y int, overlay Size, viewport Size) (int, int) {
	if viewport.Width <= 0 || viewport.Height <= 0 {
		return x, y
	}

	maxX := maxInt(0, viewport.Width-overlay.Width)
	maxY := maxInt(0, viewport.Height-overlay.Height)
	return clamp(x, 0, maxX), clamp(y, 0, maxY)
}

func clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
