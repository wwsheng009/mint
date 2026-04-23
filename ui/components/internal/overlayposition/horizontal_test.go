package overlayposition

import (
	"reflect"
	"testing"
)

func TestHorizontalFamilyCandidates(t *testing.T) {
	got := HorizontalFamilyCandidates(PlacementLeft, HorizontalCenterFirst)
	want := []Placement{PlacementLeft, PlacementLeftTop, PlacementLeftBottom}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("HorizontalFamilyCandidates(left, center-first) = %v, want %v", got, want)
	}

	got = HorizontalFamilyCandidates(PlacementRightBottom, HorizontalEdgesFirst)
	want = []Placement{PlacementRightTop, PlacementRightBottom, PlacementRight}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("HorizontalFamilyCandidates(right-bottom, edges-first) = %v, want %v", got, want)
	}
}

func TestHorizontalPlacementCandidates(t *testing.T) {
	got := HorizontalPlacementCandidates(PlacementLeft)
	want := []Placement{
		PlacementLeft,
		PlacementLeftTop,
		PlacementLeftBottom,
		PlacementRight,
		PlacementRightTop,
		PlacementRightBottom,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("HorizontalPlacementCandidates(left) = %v, want %v", got, want)
	}

	got = HorizontalPlacementCandidates(PlacementRightTop)
	want = []Placement{
		PlacementRightTop,
		PlacementRight,
		PlacementRightBottom,
		PlacementLeftTop,
		PlacementLeft,
		PlacementLeftBottom,
		PlacementTop,
		PlacementBottom,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("HorizontalPlacementCandidates(right-top) = %v, want %v", got, want)
	}

	got = HorizontalPlacementCandidates(PlacementRightBottom)
	want = []Placement{
		PlacementRightBottom,
		PlacementRight,
		PlacementRightTop,
		PlacementLeftBottom,
		PlacementLeft,
		PlacementLeftTop,
		PlacementBottom,
		PlacementTop,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("HorizontalPlacementCandidates(right-bottom) = %v, want %v", got, want)
	}
}

func TestHorizontalResolveFallsBackToTopAfterBothHorizontalFamiliesOverflow(t *testing.T) {
	result := Resolve(Config{
		Anchor:     Rect{X: 4, Y: 3, Width: 2, Height: 1},
		Overlay:    Size{Width: 8, Height: 2},
		Viewport:   Size{Width: 10, Height: 8},
		Candidates: HorizontalPlacementCandidates(PlacementRightTop),
		Gap:        1,
	})

	if result.Placement != PlacementTop {
		t.Fatalf("placement = %v, want %v", result.Placement, PlacementTop)
	}
	if result.X != 1 || result.Y != 0 {
		t.Fatalf("position = (%d,%d), want (1,0)", result.X, result.Y)
	}
	if result.Clamped {
		t.Fatal("result should not be clamped when top fallback fits")
	}
}

func TestHorizontalResolveFallsBackToBottomAfterRightBottomFamiliesOverflow(t *testing.T) {
	result := Resolve(Config{
		Anchor:     Rect{X: 4, Y: 3, Width: 2, Height: 1},
		Overlay:    Size{Width: 8, Height: 2},
		Viewport:   Size{Width: 10, Height: 8},
		Candidates: HorizontalPlacementCandidates(PlacementRightBottom),
		Gap:        1,
	})

	if result.Placement != PlacementBottom {
		t.Fatalf("placement = %v, want %v", result.Placement, PlacementBottom)
	}
	if result.X != 1 || result.Y != 5 {
		t.Fatalf("position = (%d,%d), want (1,5)", result.X, result.Y)
	}
	if result.Clamped {
		t.Fatal("result should not be clamped when bottom fallback fits")
	}
}

func TestHorizontalResolveFallsBackToBottomAfterBothHorizontalFamiliesOverflow(t *testing.T) {
	result := Resolve(Config{
		Anchor:     Rect{X: 4, Y: 3, Width: 2, Height: 1},
		Overlay:    Size{Width: 8, Height: 2},
		Viewport:   Size{Width: 10, Height: 8},
		Candidates: HorizontalPlacementCandidates(PlacementLeftBottom),
		Gap:        1,
	})

	if result.Placement != PlacementBottom {
		t.Fatalf("placement = %v, want %v", result.Placement, PlacementBottom)
	}
	if result.X != 1 || result.Y != 5 {
		t.Fatalf("position = (%d,%d), want (1,5)", result.X, result.Y)
	}
	if result.Clamped {
		t.Fatal("result should not be clamped when bottom fallback fits")
	}
}

func TestHorizontalResolveFallsBackToTopAfterLeftTopFamiliesOverflow(t *testing.T) {
	result := Resolve(Config{
		Anchor:     Rect{X: 4, Y: 3, Width: 2, Height: 1},
		Overlay:    Size{Width: 8, Height: 2},
		Viewport:   Size{Width: 10, Height: 8},
		Candidates: HorizontalPlacementCandidates(PlacementLeftTop),
		Gap:        1,
	})

	if result.Placement != PlacementTop {
		t.Fatalf("placement = %v, want %v", result.Placement, PlacementTop)
	}
	if result.X != 1 || result.Y != 0 {
		t.Fatalf("position = (%d,%d), want (1,0)", result.X, result.Y)
	}
	if result.Clamped {
		t.Fatal("result should not be clamped when top fallback fits")
	}
}
