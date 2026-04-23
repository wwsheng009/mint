package overlayposition

// HorizontalCandidateOrder controls the order of candidates within one horizontal family.
type HorizontalCandidateOrder int

const (
	HorizontalCenterFirst HorizontalCandidateOrder = iota
	HorizontalEdgesFirst
)

// HorizontalFamilyCandidates returns candidates for a left/right family without crossing to the opposite side.
func HorizontalFamilyCandidates(preferred Placement, order HorizontalCandidateOrder) []Placement {
	center, start, end := horizontalFamily(preferred)
	if order == HorizontalEdgesFirst {
		return []Placement{start, end, center}
	}
	return []Placement{center, start, end}
}

// HorizontalPlacementCandidates returns the ordered fallback list for left/right explicit placements.
func HorizontalPlacementCandidates(placement Placement) []Placement {
	switch placement {
	case PlacementLeft:
		return []Placement{
			PlacementLeft,
			PlacementLeftTop,
			PlacementLeftBottom,
			PlacementRight,
			PlacementRightTop,
			PlacementRightBottom,
		}
	case PlacementLeftTop:
		return []Placement{
			PlacementLeftTop,
			PlacementLeft,
			PlacementLeftBottom,
			PlacementRightTop,
			PlacementRight,
			PlacementRightBottom,
			PlacementTop,
			PlacementBottom,
		}
	case PlacementLeftBottom:
		return []Placement{
			PlacementLeftBottom,
			PlacementLeft,
			PlacementLeftTop,
			PlacementRightBottom,
			PlacementRight,
			PlacementRightTop,
			PlacementBottom,
			PlacementTop,
		}
	case PlacementRight:
		return []Placement{
			PlacementRight,
			PlacementRightTop,
			PlacementRightBottom,
			PlacementLeft,
			PlacementLeftTop,
			PlacementLeftBottom,
		}
	case PlacementRightTop:
		return []Placement{
			PlacementRightTop,
			PlacementRight,
			PlacementRightBottom,
			PlacementLeftTop,
			PlacementLeft,
			PlacementLeftBottom,
			PlacementTop,
			PlacementBottom,
		}
	case PlacementRightBottom:
		return []Placement{
			PlacementRightBottom,
			PlacementRight,
			PlacementRightTop,
			PlacementLeftBottom,
			PlacementLeft,
			PlacementLeftTop,
			PlacementBottom,
			PlacementTop,
		}
	default:
		return []Placement{
			PlacementLeft,
			PlacementLeftTop,
			PlacementLeftBottom,
			PlacementRight,
			PlacementRightTop,
			PlacementRightBottom,
		}
	}
}

func horizontalFamily(preferred Placement) (center, start, end Placement) {
	switch preferred {
	case PlacementRight, PlacementRightTop, PlacementRightBottom:
		return PlacementRight, PlacementRightTop, PlacementRightBottom
	default:
		return PlacementLeft, PlacementLeftTop, PlacementLeftBottom
	}
}
