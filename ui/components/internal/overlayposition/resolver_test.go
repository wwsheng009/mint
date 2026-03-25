package overlayposition

import "testing"

func TestCoordinates(t *testing.T) {
	anchor := Rect{X: 10, Y: 10, Width: 20, Height: 5}
	overlay := Size{Width: 6, Height: 1}

	tests := []struct {
		name      string
		placement Placement
		wantX     int
		wantY     int
	}{
		{name: "top", placement: PlacementTop, wantX: 17, wantY: 8},
		{name: "top left", placement: PlacementTopLeft, wantX: 10, wantY: 8},
		{name: "top right", placement: PlacementTopRight, wantX: 24, wantY: 8},
		{name: "bottom", placement: PlacementBottom, wantX: 17, wantY: 16},
		{name: "bottom left", placement: PlacementBottomLeft, wantX: 10, wantY: 16},
		{name: "bottom right", placement: PlacementBottomRight, wantX: 24, wantY: 16},
		{name: "left", placement: PlacementLeft, wantX: 3, wantY: 12},
		{name: "left top", placement: PlacementLeftTop, wantX: 3, wantY: 10},
		{name: "left bottom", placement: PlacementLeftBottom, wantX: 3, wantY: 14},
		{name: "right", placement: PlacementRight, wantX: 31, wantY: 12},
		{name: "right top", placement: PlacementRightTop, wantX: 31, wantY: 10},
		{name: "right bottom", placement: PlacementRightBottom, wantX: 31, wantY: 14},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x, y := Coordinates(anchor, overlay, tt.placement, 1)
			if x != tt.wantX || y != tt.wantY {
				t.Fatalf("Coordinates() = (%d,%d), want (%d,%d)", x, y, tt.wantX, tt.wantY)
			}
		})
	}
}

func TestRectFromBounds(t *testing.T) {
	rect := RectFromBounds([4]int{3, 4, 5, 6})
	if rect != (Rect{X: 3, Y: 4, Width: 5, Height: 6}) {
		t.Fatalf("RectFromBounds() = %+v", rect)
	}
}

func TestPointerXClampsToOverlayBorder(t *testing.T) {
	anchor := Rect{X: 10, Y: 0, Width: 6, Height: 1}
	if got := PointerX(anchor, 8, 10); got != 13 {
		t.Fatalf("PointerX() = %d, want 13", got)
	}

	offLeft := Rect{X: 0, Y: 0, Width: 2, Height: 1}
	if got := PointerX(offLeft, 8, 10); got != 9 {
		t.Fatalf("PointerX() left clamp = %d, want 9", got)
	}

	offRight := Rect{X: 30, Y: 0, Width: 4, Height: 1}
	if got := PointerX(offRight, 8, 10); got != 16 {
		t.Fatalf("PointerX() right clamp = %d, want 16", got)
	}
}

func TestResolveSelectsFirstCandidateThatFits(t *testing.T) {
	result := Resolve(Config{
		Anchor:   Rect{X: 1, Y: 1, Width: 4, Height: 2},
		Overlay:  Size{Width: 6, Height: 1},
		Viewport: Size{Width: 20, Height: 10},
		Candidates: []Placement{
			PlacementTopLeft,
			PlacementTop,
			PlacementTopRight,
			PlacementBottomLeft,
			PlacementBottom,
		},
		Gap: 1,
	})

	if result.Placement != PlacementBottomLeft {
		t.Fatalf("placement = %v, want %v", result.Placement, PlacementBottomLeft)
	}
	if result.X != 1 || result.Y != 4 {
		t.Fatalf("position = (%d,%d), want (1,4)", result.X, result.Y)
	}
	if result.Clamped {
		t.Fatal("result should not be clamped")
	}
}

func TestResolveClampsWhenNoCandidateFits(t *testing.T) {
	result := Resolve(Config{
		Anchor:     Rect{X: 8, Y: 4, Width: 2, Height: 2},
		Overlay:    Size{Width: 13, Height: 1},
		Viewport:   Size{Width: 8, Height: 5},
		Candidates: []Placement{PlacementRightBottom},
		Gap:        1,
	})

	if result.Placement != PlacementRightBottom {
		t.Fatalf("placement = %v, want %v", result.Placement, PlacementRightBottom)
	}
	if result.X != 0 || result.Y != 4 {
		t.Fatalf("position = (%d,%d), want (0,4)", result.X, result.Y)
	}
	if !result.Clamped {
		t.Fatal("result should report clamping")
	}
}

func TestResolveWithoutViewportKeepsPreferredPlacement(t *testing.T) {
	result := Resolve(Config{
		Anchor:     Rect{X: 1, Y: 1, Width: 4, Height: 2},
		Overlay:    Size{Width: 6, Height: 1},
		Candidates: []Placement{PlacementTop},
		Gap:        1,
	})

	if result.Placement != PlacementTop {
		t.Fatalf("placement = %v, want %v", result.Placement, PlacementTop)
	}
	if result.X != 0 || result.Y != -1 {
		t.Fatalf("position = (%d,%d), want (0,-1)", result.X, result.Y)
	}
	if result.Clamped {
		t.Fatal("result should not be clamped without viewport")
	}
}
