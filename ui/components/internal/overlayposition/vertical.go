package overlayposition

// VerticalAutoBias controls which side wins when both top and bottom fit.
type VerticalAutoBias int

const (
	VerticalAutoPreferTop VerticalAutoBias = iota
	VerticalAutoPreferBottom
)

// VerticalCandidateOrder controls the order of candidates within one vertical family.
type VerticalCandidateOrder int

const (
	VerticalCenterFirst VerticalCandidateOrder = iota
	VerticalEdgesFirst
)

// PreferredVerticalPlacement returns the preferred top/bottom family for an auto-placement overlay.
func PreferredVerticalPlacement(anchor Rect, overlayHeight, gap, viewportHeight int, bias VerticalAutoBias) Placement {
	if gap < 0 {
		gap = 0
	}

	aboveBoxY := anchor.Y - overlayHeight - gap
	belowBoxY := anchor.Y + anchor.Height + gap
	fitsBelow := viewportHeight <= 0 || belowBoxY+overlayHeight <= viewportHeight
	fitsAbove := aboveBoxY >= 0 && (viewportHeight <= 0 || aboveBoxY+overlayHeight <= viewportHeight)
	switch {
	case fitsAbove && !fitsBelow:
		return PlacementTop
	case !fitsAbove && fitsBelow:
		return PlacementBottom
	case fitsAbove:
		if bias == VerticalAutoPreferBottom {
			return PlacementBottom
		}
		return PlacementTop
	case viewportHeight <= 0 && anchor.Y > overlayHeight+gap:
		return PlacementTop
	default:
		return PlacementBottom
	}
}

// VerticalFamilyCandidates returns candidates for a top/bottom family without crossing to the opposite side.
func VerticalFamilyCandidates(preferred Placement, order VerticalCandidateOrder) []Placement {
	center, start, end := verticalFamily(preferred)
	if order == VerticalEdgesFirst {
		return []Placement{start, end, center}
	}
	return []Placement{center, start, end}
}

// VerticalPlacementCandidates returns the ordered fallback list for top/bottom explicit placements.
func VerticalPlacementCandidates(placement Placement) []Placement {
	switch placement {
	case PlacementTop:
		return []Placement{
			PlacementTop,
			PlacementTopLeft,
			PlacementTopRight,
			PlacementBottom,
			PlacementBottomLeft,
			PlacementBottomRight,
		}
	case PlacementTopLeft:
		return []Placement{
			PlacementTopLeft,
			PlacementTop,
			PlacementTopRight,
			PlacementBottomLeft,
			PlacementBottom,
		}
	case PlacementTopRight:
		return []Placement{
			PlacementTopRight,
			PlacementTop,
			PlacementTopLeft,
			PlacementBottomRight,
			PlacementBottom,
		}
	case PlacementBottomLeft:
		return []Placement{
			PlacementBottomLeft,
			PlacementBottom,
			PlacementBottomRight,
			PlacementTopLeft,
			PlacementTop,
		}
	case PlacementBottomRight:
		return []Placement{
			PlacementBottomRight,
			PlacementBottom,
			PlacementBottomLeft,
			PlacementTopRight,
			PlacementTop,
		}
	case PlacementBottom:
		return []Placement{
			PlacementBottom,
			PlacementBottomLeft,
			PlacementBottomRight,
			PlacementTop,
			PlacementTopLeft,
			PlacementTopRight,
		}
	default:
		return []Placement{
			PlacementTop,
			PlacementTopLeft,
			PlacementTopRight,
			PlacementBottom,
			PlacementBottomLeft,
			PlacementBottomRight,
		}
	}
}

func verticalFamily(preferred Placement) (center, start, end Placement) {
	switch preferred {
	case PlacementBottom, PlacementBottomLeft, PlacementBottomRight:
		return PlacementBottom, PlacementBottomLeft, PlacementBottomRight
	default:
		return PlacementTop, PlacementTopLeft, PlacementTopRight
	}
}
