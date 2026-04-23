package overlayposition

import "testing"

func TestPreferredVerticalPlacement(t *testing.T) {
	anchor := Rect{X: 10, Y: 6, Width: 4, Height: 1}

	if got := PreferredVerticalPlacement(anchor, 3, 1, 12, VerticalAutoPreferTop); got != PlacementTop {
		t.Fatalf("prefer top = %v, want %v", got, PlacementTop)
	}
	if got := PreferredVerticalPlacement(anchor, 3, 1, 12, VerticalAutoPreferBottom); got != PlacementBottom {
		t.Fatalf("prefer bottom = %v, want %v", got, PlacementBottom)
	}
	if got := PreferredVerticalPlacement(Rect{X: 10, Y: 1, Width: 4, Height: 1}, 3, 1, 12, VerticalAutoPreferTop); got != PlacementBottom {
		t.Fatalf("top edge fallback = %v, want %v", got, PlacementBottom)
	}
}

func TestVerticalFamilyCandidates(t *testing.T) {
	got := VerticalFamilyCandidates(PlacementTop, VerticalEdgesFirst)
	want := []Placement{PlacementTopLeft, PlacementTopRight, PlacementTop}
	assertPlacementsEqual(t, got, want)

	got = VerticalFamilyCandidates(PlacementBottom, VerticalCenterFirst)
	want = []Placement{PlacementBottom, PlacementBottomLeft, PlacementBottomRight}
	assertPlacementsEqual(t, got, want)
}

func TestVerticalPlacementCandidates(t *testing.T) {
	got := VerticalPlacementCandidates(PlacementTopRight)
	want := []Placement{
		PlacementTopRight,
		PlacementTop,
		PlacementTopLeft,
		PlacementBottomRight,
		PlacementBottom,
	}
	assertPlacementsEqual(t, got, want)

	got = VerticalPlacementCandidates(PlacementTopLeft)
	want = []Placement{
		PlacementTopLeft,
		PlacementTop,
		PlacementTopRight,
		PlacementBottomLeft,
		PlacementBottom,
	}
	assertPlacementsEqual(t, got, want)

	got = VerticalPlacementCandidates(PlacementBottomLeft)
	want = []Placement{
		PlacementBottomLeft,
		PlacementBottom,
		PlacementBottomRight,
		PlacementTopLeft,
		PlacementTop,
	}
	assertPlacementsEqual(t, got, want)
}

func TestVerticalResolveClampsTopRightAndStaysInTopFamilyInNarrowViewport(t *testing.T) {
	result := Resolve(Config{
		Anchor:     Rect{X: 4, Y: 4, Width: 2, Height: 1},
		Overlay:    Size{Width: 10, Height: 2},
		Viewport:   Size{Width: 8, Height: 10},
		Candidates: VerticalPlacementCandidates(PlacementTopRight),
		Gap:        1,
	})

	if result.Placement != PlacementTopRight {
		t.Fatalf("placement = %v, want %v", result.Placement, PlacementTopRight)
	}
	if result.X != 0 || result.Y != 1 {
		t.Fatalf("position = (%d,%d), want (0,1)", result.X, result.Y)
	}
	if !result.Clamped {
		t.Fatal("result should report left-edge clamping in narrow viewport")
	}
}

func TestVerticalResolveClampsTopLeftAndStaysInTopFamilyInNarrowViewport(t *testing.T) {
	result := Resolve(Config{
		Anchor:     Rect{X: 4, Y: 4, Width: 2, Height: 1},
		Overlay:    Size{Width: 10, Height: 2},
		Viewport:   Size{Width: 8, Height: 10},
		Candidates: VerticalPlacementCandidates(PlacementTopLeft),
		Gap:        1,
	})

	if result.Placement != PlacementTopLeft {
		t.Fatalf("placement = %v, want %v", result.Placement, PlacementTopLeft)
	}
	if result.X != 0 || result.Y != 1 {
		t.Fatalf("position = (%d,%d), want (0,1)", result.X, result.Y)
	}
	if !result.Clamped {
		t.Fatal("result should report right-edge clamping in narrow viewport")
	}
}

func TestVerticalResolveClampsTopLeftOnBothAxesAndStaysInTopFamilyWhenNothingFits(t *testing.T) {
	result := Resolve(Config{
		Anchor:     Rect{X: 4, Y: 1, Width: 2, Height: 1},
		Overlay:    Size{Width: 10, Height: 3},
		Viewport:   Size{Width: 8, Height: 4},
		Candidates: VerticalPlacementCandidates(PlacementTopLeft),
		Gap:        1,
	})

	if result.Placement != PlacementTopLeft {
		t.Fatalf("placement = %v, want %v", result.Placement, PlacementTopLeft)
	}
	if result.X != 0 || result.Y != 0 {
		t.Fatalf("position = (%d,%d), want (0,0)", result.X, result.Y)
	}
	if !result.Clamped {
		t.Fatal("result should report dual-axis clamping when no vertical candidate fits")
	}
}

func TestVerticalResolveFallsBelowWithinLeftFamilyNearTopLeftCorner(t *testing.T) {
	result := Resolve(Config{
		Anchor:     Rect{X: 2, Y: 1, Width: 4, Height: 1},
		Overlay:    Size{Width: 10, Height: 2},
		Viewport:   Size{Width: 40, Height: 10},
		Candidates: VerticalPlacementCandidates(PlacementTopLeft),
		Gap:        1,
	})

	if result.Placement != PlacementBottomLeft {
		t.Fatalf("placement = %v, want %v", result.Placement, PlacementBottomLeft)
	}
	if result.X != 2 || result.Y != 3 {
		t.Fatalf("position = (%d,%d), want (2,3)", result.X, result.Y)
	}
	if result.Clamped {
		t.Fatal("result should not be clamped when bottom-left fallback fits")
	}
}

func TestVerticalResolveFallsBelowWithinRightFamilyNearTopRightCorner(t *testing.T) {
	result := Resolve(Config{
		Anchor:     Rect{X: 34, Y: 1, Width: 4, Height: 1},
		Overlay:    Size{Width: 10, Height: 2},
		Viewport:   Size{Width: 40, Height: 10},
		Candidates: VerticalPlacementCandidates(PlacementTopRight),
		Gap:        1,
	})

	if result.Placement != PlacementBottomRight {
		t.Fatalf("placement = %v, want %v", result.Placement, PlacementBottomRight)
	}
	if result.X != 28 || result.Y != 3 {
		t.Fatalf("position = (%d,%d), want (28,3)", result.X, result.Y)
	}
	if result.Clamped {
		t.Fatal("result should not be clamped when bottom-right fallback fits")
	}
}

func TestVerticalResolveClampsBottomRightAndStaysInBottomFamilyInNarrowViewport(t *testing.T) {
	result := Resolve(Config{
		Anchor:     Rect{X: 4, Y: 4, Width: 2, Height: 1},
		Overlay:    Size{Width: 10, Height: 2},
		Viewport:   Size{Width: 8, Height: 10},
		Candidates: VerticalPlacementCandidates(PlacementBottomRight),
		Gap:        1,
	})

	if result.Placement != PlacementBottomRight {
		t.Fatalf("placement = %v, want %v", result.Placement, PlacementBottomRight)
	}
	if result.X != 0 || result.Y != 6 {
		t.Fatalf("position = (%d,%d), want (0,6)", result.X, result.Y)
	}
	if !result.Clamped {
		t.Fatal("result should report left-edge clamping in narrow viewport")
	}
}

func TestVerticalResolveClampsTopRightOnBothAxesAndStaysInTopFamilyWhenNothingFits(t *testing.T) {
	result := Resolve(Config{
		Anchor:     Rect{X: 4, Y: 1, Width: 2, Height: 1},
		Overlay:    Size{Width: 10, Height: 3},
		Viewport:   Size{Width: 8, Height: 4},
		Candidates: VerticalPlacementCandidates(PlacementTopRight),
		Gap:        1,
	})

	if result.Placement != PlacementTopRight {
		t.Fatalf("placement = %v, want %v", result.Placement, PlacementTopRight)
	}
	if result.X != 0 || result.Y != 0 {
		t.Fatalf("position = (%d,%d), want (0,0)", result.X, result.Y)
	}
	if !result.Clamped {
		t.Fatal("result should report dual-axis clamping when no vertical candidate fits")
	}
}

func TestVerticalResolveFallsAboveWithinLeftFamilyNearBottomLeftCorner(t *testing.T) {
	result := Resolve(Config{
		Anchor:     Rect{X: 2, Y: 8, Width: 4, Height: 1},
		Overlay:    Size{Width: 10, Height: 2},
		Viewport:   Size{Width: 40, Height: 10},
		Candidates: VerticalPlacementCandidates(PlacementBottomLeft),
		Gap:        1,
	})

	if result.Placement != PlacementTopLeft {
		t.Fatalf("placement = %v, want %v", result.Placement, PlacementTopLeft)
	}
	if result.X != 2 || result.Y != 5 {
		t.Fatalf("position = (%d,%d), want (2,5)", result.X, result.Y)
	}
	if result.Clamped {
		t.Fatal("result should not be clamped when top-left fallback fits")
	}
}

func TestVerticalResolveClampsBottomLeftAndStaysInBottomFamilyInNarrowViewport(t *testing.T) {
	result := Resolve(Config{
		Anchor:     Rect{X: 4, Y: 4, Width: 2, Height: 1},
		Overlay:    Size{Width: 10, Height: 2},
		Viewport:   Size{Width: 8, Height: 10},
		Candidates: VerticalPlacementCandidates(PlacementBottomLeft),
		Gap:        1,
	})

	if result.Placement != PlacementBottomLeft {
		t.Fatalf("placement = %v, want %v", result.Placement, PlacementBottomLeft)
	}
	if result.X != 0 || result.Y != 6 {
		t.Fatalf("position = (%d,%d), want (0,6)", result.X, result.Y)
	}
	if !result.Clamped {
		t.Fatal("result should report right-edge clamping in narrow viewport")
	}
}

func TestVerticalResolveClampsBottomLeftOnBothAxesAndStaysInBottomFamilyWhenNothingFits(t *testing.T) {
	result := Resolve(Config{
		Anchor:     Rect{X: 4, Y: 1, Width: 2, Height: 1},
		Overlay:    Size{Width: 10, Height: 3},
		Viewport:   Size{Width: 8, Height: 4},
		Candidates: VerticalPlacementCandidates(PlacementBottomLeft),
		Gap:        1,
	})

	if result.Placement != PlacementBottomLeft {
		t.Fatalf("placement = %v, want %v", result.Placement, PlacementBottomLeft)
	}
	if result.X != 0 || result.Y != 1 {
		t.Fatalf("position = (%d,%d), want (0,1)", result.X, result.Y)
	}
	if !result.Clamped {
		t.Fatal("result should report dual-axis clamping when no vertical candidate fits")
	}
}

func TestVerticalResolveFallsAboveWithinRightFamilyNearBottomRightCorner(t *testing.T) {
	result := Resolve(Config{
		Anchor:     Rect{X: 34, Y: 8, Width: 4, Height: 1},
		Overlay:    Size{Width: 10, Height: 2},
		Viewport:   Size{Width: 40, Height: 10},
		Candidates: VerticalPlacementCandidates(PlacementBottomRight),
		Gap:        1,
	})

	if result.Placement != PlacementTopRight {
		t.Fatalf("placement = %v, want %v", result.Placement, PlacementTopRight)
	}
	if result.X != 28 || result.Y != 5 {
		t.Fatalf("position = (%d,%d), want (28,5)", result.X, result.Y)
	}
	if result.Clamped {
		t.Fatal("result should not be clamped when top-right fallback fits")
	}
}

func TestVerticalResolveClampsBottomRightOnBothAxesAndStaysInBottomFamilyWhenNothingFits(t *testing.T) {
	result := Resolve(Config{
		Anchor:     Rect{X: 4, Y: 1, Width: 2, Height: 1},
		Overlay:    Size{Width: 10, Height: 3},
		Viewport:   Size{Width: 8, Height: 4},
		Candidates: VerticalPlacementCandidates(PlacementBottomRight),
		Gap:        1,
	})

	if result.Placement != PlacementBottomRight {
		t.Fatalf("placement = %v, want %v", result.Placement, PlacementBottomRight)
	}
	if result.X != 0 || result.Y != 1 {
		t.Fatalf("position = (%d,%d), want (0,1)", result.X, result.Y)
	}
	if !result.Clamped {
		t.Fatal("result should report dual-axis clamping when no vertical candidate fits")
	}
}

func assertPlacementsEqual(t *testing.T, got, want []Placement) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("placements[%d] = %v, want %v", i, got[i], want[i])
		}
	}
}
