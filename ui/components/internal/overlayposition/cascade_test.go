package overlayposition

import "testing"

func TestResolveCascadePrefersRightWhenViewportAllows(t *testing.T) {
	result := ResolveCascade(CascadeConfig{
		Parent:             Rect{X: 30, Y: 4, Width: 12, Height: 5},
		Overlay:            Size{Width: 10, Height: 6},
		Viewport:           Size{Width: 80, Height: 24},
		Top:                8,
		PreferredDirection: CascadeRight,
	})

	if result.X != 42 {
		t.Fatalf("result.X = %d, want 42", result.X)
	}
	if result.Y != 8 {
		t.Fatalf("result.Y = %d, want 8", result.Y)
	}
	if result.Direction != CascadeRight {
		t.Fatalf("result.Direction = %v, want CascadeRight", result.Direction)
	}
	if result.ClampedX || result.ClampedY {
		t.Fatalf("result clamp flags = (%v,%v), want false,false", result.ClampedX, result.ClampedY)
	}
}

func TestResolveCascadeFallsBackLeftWhenRightDoesNotFit(t *testing.T) {
	result := ResolveCascade(CascadeConfig{
		Parent:             Rect{X: 60, Y: 2, Width: 12, Height: 5},
		Overlay:            Size{Width: 14, Height: 6},
		Viewport:           Size{Width: 72, Height: 18},
		Top:                3,
		PreferredDirection: CascadeRight,
	})

	if result.X != 46 {
		t.Fatalf("result.X = %d, want 46", result.X)
	}
	if result.Direction != CascadeLeft {
		t.Fatalf("result.Direction = %v, want CascadeLeft", result.Direction)
	}
}

func TestResolveCascadePrefersLeftWhenViewportAllows(t *testing.T) {
	result := ResolveCascade(CascadeConfig{
		Parent:             Rect{X: 32, Y: 3, Width: 10, Height: 4},
		Overlay:            Size{Width: 12, Height: 5},
		Viewport:           Size{Width: 80, Height: 24},
		Top:                4,
		PreferredDirection: CascadeLeft,
	})

	if result.X != 20 {
		t.Fatalf("result.X = %d, want 20", result.X)
	}
	if result.Direction != CascadeLeft {
		t.Fatalf("result.Direction = %v, want CascadeLeft", result.Direction)
	}
	if result.ClampedX || result.ClampedY {
		t.Fatalf("result clamp flags = (%v,%v), want false,false", result.ClampedX, result.ClampedY)
	}
}

func TestResolveCascadeFallsBackRightWhenLeftDoesNotFit(t *testing.T) {
	result := ResolveCascade(CascadeConfig{
		Parent:             Rect{X: 4, Y: 2, Width: 12, Height: 5},
		Overlay:            Size{Width: 14, Height: 6},
		Viewport:           Size{Width: 72, Height: 18},
		Top:                3,
		PreferredDirection: CascadeLeft,
	})

	if result.X != 16 {
		t.Fatalf("result.X = %d, want 16", result.X)
	}
	if result.Direction != CascadeRight {
		t.Fatalf("result.Direction = %v, want CascadeRight", result.Direction)
	}
}

func TestResolveCascadeInfersLeftDirectionFromClampedPreferredSide(t *testing.T) {
	result := ResolveCascade(CascadeConfig{
		Parent:             Rect{X: 13, Y: 1, Width: 33, Height: 3},
		Overlay:            Size{Width: 28, Height: 3},
		Viewport:           Size{Width: 48, Height: 12},
		Top:                9,
		PreferredDirection: CascadeRight,
	})

	if result.X != 20 {
		t.Fatalf("result.X = %d, want 20 after clamp", result.X)
	}
	if result.Direction != CascadeRight {
		t.Fatalf("result.Direction = %v, want CascadeRight", result.Direction)
	}

	narrow := ResolveCascade(CascadeConfig{
		Parent:             Rect{X: 13, Y: 1, Width: 33, Height: 3},
		Overlay:            Size{Width: 35, Height: 3},
		Viewport:           Size{Width: 48, Height: 12},
		Top:                9,
		PreferredDirection: CascadeRight,
	})
	if narrow.X != 13 {
		t.Fatalf("narrow.X = %d, want 13 after clamp", narrow.X)
	}
	if narrow.Direction != CascadeRight {
		t.Fatalf("narrow.Direction = %v, want CascadeRight when resolved left == parent left", narrow.Direction)
	}

	leftClamped := ResolveCascade(CascadeConfig{
		Parent:             Rect{X: 18, Y: 1, Width: 22, Height: 3},
		Overlay:            Size{Width: 28, Height: 3},
		Viewport:           Size{Width: 40, Height: 12},
		Top:                9,
		PreferredDirection: CascadeRight,
	})
	if leftClamped.X != 12 {
		t.Fatalf("leftClamped.X = %d, want 12", leftClamped.X)
	}
	if leftClamped.Direction != CascadeLeft {
		t.Fatalf("leftClamped.Direction = %v, want CascadeLeft when clamp lands left of parent", leftClamped.Direction)
	}
}

func TestResolveCascadeClampsVerticalTopIntoViewport(t *testing.T) {
	result := ResolveCascade(CascadeConfig{
		Parent:             Rect{X: 50, Y: 10, Width: 14, Height: 5},
		Overlay:            Size{Width: 12, Height: 8},
		Viewport:           Size{Width: 72, Height: 18},
		Top:                15,
		PreferredDirection: CascadeLeft,
	})

	if result.Y != 10 {
		t.Fatalf("result.Y = %d, want 10 after bottom-edge clamp", result.Y)
	}
	if !result.ClampedY {
		t.Fatal("result.ClampedY = false, want true")
	}
	if result.Direction != CascadeLeft {
		t.Fatalf("result.Direction = %v, want CascadeLeft", result.Direction)
	}
}

func TestResolveCascadeClampsNegativeTopIntoViewport(t *testing.T) {
	result := ResolveCascade(CascadeConfig{
		Parent:             Rect{X: 40, Y: 1, Width: 12, Height: 4},
		Overlay:            Size{Width: 10, Height: 7},
		Viewport:           Size{Width: 72, Height: 18},
		Top:                -3,
		PreferredDirection: CascadeRight,
	})

	if result.Y != 0 {
		t.Fatalf("result.Y = %d, want 0 after top-edge clamp", result.Y)
	}
	if !result.ClampedY {
		t.Fatal("result.ClampedY = false, want true")
	}
	if result.Direction != CascadeRight {
		t.Fatalf("result.Direction = %v, want CascadeRight", result.Direction)
	}
}

func TestResolveCascadeClampsBothAxesWhenNeitherSideFits(t *testing.T) {
	result := ResolveCascade(CascadeConfig{
		Parent:             Rect{X: 18, Y: 8, Width: 16, Height: 4},
		Overlay:            Size{Width: 28, Height: 9},
		Viewport:           Size{Width: 30, Height: 12},
		Top:                9,
		PreferredDirection: CascadeLeft,
	})

	if result.X != 0 {
		t.Fatalf("result.X = %d, want 0 after horizontal clamp", result.X)
	}
	if result.Y != 3 {
		t.Fatalf("result.Y = %d, want 3 after vertical clamp", result.Y)
	}
	if !result.ClampedX || !result.ClampedY {
		t.Fatalf("result clamp flags = (%v,%v), want true,true", result.ClampedX, result.ClampedY)
	}
	if result.Direction != CascadeLeft {
		t.Fatalf("result.Direction = %v, want CascadeLeft when final box stays left of parent", result.Direction)
	}
}

func TestResolveCascadeFallsBackRightAfterLeftEdgeClampDirection(t *testing.T) {
	result := ResolveCascade(CascadeConfig{
		Parent:             Rect{X: 0, Y: 10, Width: 18, Height: 6},
		Overlay:            Size{Width: 12, Height: 5},
		Viewport:           Size{Width: 36, Height: 18},
		Top:                11,
		PreferredDirection: CascadeLeft,
	})

	if result.X != 18 {
		t.Fatalf("result.X = %d, want 18 when mirrored right side fits", result.X)
	}
	if result.Y != 11 {
		t.Fatalf("result.Y = %d, want 11 without vertical clamp", result.Y)
	}
	if result.Direction != CascadeRight {
		t.Fatalf("result.Direction = %v, want CascadeRight after mirrored fallback from left edge", result.Direction)
	}
	if result.ClampedX || result.ClampedY {
		t.Fatalf("result clamp flags = (%v,%v), want false,false", result.ClampedX, result.ClampedY)
	}
}
