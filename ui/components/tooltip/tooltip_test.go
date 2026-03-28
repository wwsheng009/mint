package tooltip

import (
	"reflect"
	"testing"
	"time"

	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/control"
	"github.com/wwsheng009/mint/ui/components/internal/overlayposition"
	newtext "github.com/wwsheng009/mint/ui/components/text"
)

// =============================================================================
// Tooltip VNode Tests
// =============================================================================

func TestNewTooltip(t *testing.T) {
	content := newtext.New("button")
	tooltip := New(content, "Click me")

	if tooltip.Tag() != "tooltip" {
		t.Errorf("Expected tag 'tooltip', got '%s'", tooltip.Tag())
	}

	if tooltip.Text() != "Click me" {
		t.Errorf("Expected text 'Click me', got '%s'", tooltip.Text())
	}

	if tooltip.Position() != PositionAuto {
		t.Errorf("Expected position PositionAuto, got %v", tooltip.Position())
	}

	if tooltip.Delay() != 500*time.Millisecond {
		t.Errorf("Expected delay 500ms, got %v", tooltip.Delay())
	}
}

func TestTooltipBuilder(t *testing.T) {
	content := newtext.New("button")
	tooltip := NewBuilder(content, "Help").
		Key("tooltip1").
		Position(PositionTop).
		Delay(1000 * time.Millisecond).
		Build()

	tooltipVNode, ok := tooltip.(*VNode)
	if !ok {
		t.Fatal("Expected *VNode")
	}

	if tooltipVNode.Key() != "tooltip1" {
		t.Errorf("Expected key 'tooltip1', got '%s'", tooltipVNode.Key())
	}

	if tooltipVNode.Position() != PositionTop {
		t.Errorf("Expected position PositionTop, got %v", tooltipVNode.Position())
	}

	if tooltipVNode.Delay() != 1000*time.Millisecond {
		t.Errorf("Expected delay 1000ms, got %v", tooltipVNode.Delay())
	}
}

func TestTooltipPositionShortcuts(t *testing.T) {
	content := newtext.New("button")

	tests := []struct {
		name     string
		builder  *Builder
		expected Position
	}{
		{"Top", NewBuilder(content, "text").Top(), PositionTop},
		{"TopLeft", NewBuilder(content, "text").TopLeft(), PositionTopLeft},
		{"TopRight", NewBuilder(content, "text").TopRight(), PositionTopRight},
		{"Bottom", NewBuilder(content, "text").Bottom(), PositionBottom},
		{"BottomLeft", NewBuilder(content, "text").BottomLeft(), PositionBottomLeft},
		{"BottomRight", NewBuilder(content, "text").BottomRight(), PositionBottomRight},
		{"Left", NewBuilder(content, "text").Left(), PositionLeft},
		{"LeftTop", NewBuilder(content, "text").LeftTop(), PositionLeftTop},
		{"LeftBottom", NewBuilder(content, "text").LeftBottom(), PositionLeftBottom},
		{"Right", NewBuilder(content, "text").Right(), PositionRight},
		{"RightTop", NewBuilder(content, "text").RightTop(), PositionRightTop},
		{"RightBottom", NewBuilder(content, "text").RightBottom(), PositionRightBottom},
		{"Auto", NewBuilder(content, "text").Auto(), PositionAuto},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vnode := tt.builder.Build().(*VNode)
			if vnode.Position() != tt.expected {
				t.Errorf("Expected position %v, got %v", tt.expected, vnode.Position())
			}
		})
	}
}

func TestTooltipStyle(t *testing.T) {
	content := newtext.New("button")
	tooltip := NewBuilder(content, "Help").
		FgColor("white").
		BgColor("blue").
		Build()

	tooltipVNode := tooltip.(*VNode)
	s := tooltipVNode.Style()

	if s.FG != style.Color("white") {
		t.Errorf("Expected FG white, got %v", s.FG)
	}

	if s.BG != style.Color("blue") {
		t.Errorf("Expected BG blue, got %v", s.BG)
	}
}

func TestTooltipConvenienceFunc(t *testing.T) {
	content := newtext.New("button")
	tooltip := Tooltip(content, "Convenient")

	if tooltip == nil {
		t.Fatal("Expected non-nil tooltip")
	}

	if tooltip.Text() != "Convenient" {
		t.Errorf("Expected text 'Convenient', got '%s'", tooltip.Text())
	}
}

// =============================================================================
// Tooltip Instance Tests
// =============================================================================

func TestNewTooltipInstance(t *testing.T) {
	props := rtui.Props{
		"key":      "test",
		"text":     "Test tooltip",
		"position": PositionTop,
		"delay":    100 * time.Millisecond,
	}

	inst := NewInstance(props)

	if inst.Key() != "test" {
		t.Errorf("Expected key 'test', got '%s'", inst.Key())
	}

	if inst.text != "Test tooltip" {
		t.Errorf("Expected text 'Test tooltip', got '%s'", inst.text)
	}

	if inst.position != PositionTop {
		t.Errorf("Expected position PositionTop, got %v", inst.position)
	}
	if inst.WantsTick() {
		t.Fatal("new tooltip should not want ticks before activation")
	}
}

func TestTooltipShowHide(t *testing.T) {
	inst := NewInstance(rtui.Props{"text": "Test"})

	// Initially not visible
	if inst.visible {
		t.Error("Expected invisible initially")
	}

	// Show
	inst.Show()
	if !inst.visible {
		t.Error("Expected visible after Show()")
	}

	// Hide
	inst.Hide()
	if inst.visible {
		t.Error("Expected invisible after Hide()")
	}
}

func TestTooltipCalculatePosition(t *testing.T) {
	inst := NewInstance(rtui.Props{"text": "Test"})

	tests := []struct {
		name                               string
		position                           Position
		anchorX, anchorY, anchorW, anchorH int
		expectedX, expectedY               int
	}{
		// Test text is "Test" (4 chars), tooltip width = 4 + 2 = 6
		{"Top", PositionTop, 10, 10, 20, 5, 17, 8}, // X = 10 + 10 - 3 = 17
		{"TopLeft", PositionTopLeft, 10, 10, 20, 5, 10, 8},
		{"TopRight", PositionTopRight, 10, 10, 20, 5, 24, 8},
		{"Bottom", PositionBottom, 10, 10, 20, 5, 17, 16}, // X = 10 + 10 - 3 = 17
		{"BottomLeft", PositionBottomLeft, 10, 10, 20, 5, 10, 16},
		{"BottomRight", PositionBottomRight, 10, 10, 20, 5, 24, 16},
		{"Left", PositionLeft, 10, 10, 20, 5, 3, 12}, // X = 10 - 6 - 1 = 3
		{"LeftTop", PositionLeftTop, 10, 10, 20, 5, 3, 10},
		{"LeftBottom", PositionLeftBottom, 10, 10, 20, 5, 3, 14},
		{"Right", PositionRight, 10, 10, 20, 5, 31, 12}, // X = 10 + 20 + 1 = 31
		{"RightTop", PositionRightTop, 10, 10, 20, 5, 31, 10},
		{"RightBottom", PositionRightBottom, 10, 10, 20, 5, 31, 14},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inst.position = tt.position
			inst.SetAnchorBounds(tt.anchorX, tt.anchorY, tt.anchorW, tt.anchorH)
			x, y := inst.CalculatePosition()

			if x != tt.expectedX || y != tt.expectedY {
				t.Errorf("Expected position (%d, %d), got (%d, %d)", tt.expectedX, tt.expectedY, x, y)
			}
		})
	}
}

func TestTooltipCalculatePosition_AutoAndFallback(t *testing.T) {
	t.Run("auto falls back into viewport", func(t *testing.T) {
		inst := NewInstance(rtui.Props{
			"text":     "Test",
			"position": PositionAuto,
		})
		inst.SetAnchorBounds(1, 1, 4, 2)
		inst.SetViewportSize(20, 10)

		x, y := inst.CalculatePosition()
		if x != 0 || y != 4 {
			t.Fatalf("auto position = (%d,%d), want (0,4)", x, y)
		}
	})

	t.Run("explicit placement falls back when clipped", func(t *testing.T) {
		inst := NewInstance(rtui.Props{
			"text":     "Test",
			"position": PositionTopLeft,
		})
		inst.SetAnchorBounds(1, 1, 4, 2)
		inst.SetViewportSize(20, 10)

		x, y := inst.CalculatePosition()
		if x != 1 || y != 4 {
			t.Fatalf("fallback position = (%d,%d), want (1,4)", x, y)
		}
	})

	t.Run("top placement keeps falling back until below anchor fits", func(t *testing.T) {
		inst := NewInstance(rtui.Props{
			"text":     "Top edge tooltip fallback",
			"position": PositionTop,
		})
		inst.SetAnchorBounds(0, 1, 19, 1)
		inst.SetViewportSize(72, 12)

		x, y := inst.CalculatePosition()
		if x != 0 || y != 3 {
			t.Fatalf("fallback position = (%d,%d), want (0,3)", x, y)
		}
	})

	t.Run("top placement stays above and shifts right within top family near left edge", func(t *testing.T) {
		inst := NewInstance(rtui.Props{
			"text":     "1234567890",
			"position": PositionTop,
		})
		inst.SetAnchorBounds(2, 8, 4, 1)
		inst.SetViewportSize(40, 16)

		x, y := inst.CalculatePosition()
		if x != 2 || y != 6 {
			t.Fatalf("left-edge top-family position = (%d,%d), want (2,6)", x, y)
		}
	})

	t.Run("top right placement falls below within right family near top-right corner", func(t *testing.T) {
		inst := NewInstance(rtui.Props{
			"text":     "1234567890",
			"position": PositionTopRight,
		})
		inst.SetAnchorBounds(34, 1, 4, 1)
		inst.SetViewportSize(40, 10)

		x, y := inst.CalculatePosition()
		if x != 26 || y != 3 {
			t.Fatalf("top-right corner fallback position = (%d,%d), want (26,3)", x, y)
		}
	})

	t.Run("top right placement clamps left and stays above in narrow viewport", func(t *testing.T) {
		inst := NewInstance(rtui.Props{
			"text":     "12345678901234",
			"position": PositionTopRight,
		})
		inst.SetAnchorBounds(9, 7, 4, 1)
		inst.SetViewportSize(14, 14)

		x, y := inst.CalculatePosition()
		if x != 0 || y != 5 {
			t.Fatalf("narrow top-right clamped position = (%d,%d), want (0,5)", x, y)
		}
	})

	t.Run("top left placement clamps left and stays above in narrow viewport", func(t *testing.T) {
		inst := NewInstance(rtui.Props{
			"text":     "12345678901234",
			"position": PositionTopLeft,
		})
		inst.SetAnchorBounds(9, 7, 4, 1)
		inst.SetViewportSize(14, 14)

		x, y := inst.CalculatePosition()
		if x != 0 || y != 5 {
			t.Fatalf("narrow top-left clamped position = (%d,%d), want (0,5)", x, y)
		}
	})

	t.Run("top left placement clamps both axes and stays above when no vertical candidate fits", func(t *testing.T) {
		inst := NewInstance(rtui.Props{
			"text":     "12345678901234",
			"position": PositionTopLeft,
		})
		inst.SetAnchorBounds(9, 1, 4, 1)
		inst.SetViewportSize(14, 3)

		x, y := inst.CalculatePosition()
		if x != 0 || y != 0 {
			t.Fatalf("dual-axis top-left clamped position = (%d,%d), want (0,0)", x, y)
		}
	})

	t.Run("top right placement clamps both axes and stays above when no vertical candidate fits", func(t *testing.T) {
		inst := NewInstance(rtui.Props{
			"text":     "12345678901234",
			"position": PositionTopRight,
		})
		inst.SetAnchorBounds(9, 1, 4, 1)
		inst.SetViewportSize(14, 3)

		x, y := inst.CalculatePosition()
		if x != 0 || y != 0 {
			t.Fatalf("dual-axis top-right clamped position = (%d,%d), want (0,0)", x, y)
		}
	})

	t.Run("top left placement falls below within left family near top-left corner", func(t *testing.T) {
		inst := NewInstance(rtui.Props{
			"text":     "1234567890",
			"position": PositionTopLeft,
		})
		inst.SetAnchorBounds(2, 1, 4, 1)
		inst.SetViewportSize(40, 10)

		x, y := inst.CalculatePosition()
		if x != 2 || y != 3 {
			t.Fatalf("top-left corner fallback position = (%d,%d), want (2,3)", x, y)
		}
	})

	t.Run("bottom placement stays below and shifts left within bottom family near right edge", func(t *testing.T) {
		inst := NewInstance(rtui.Props{
			"text":     "1234567890",
			"position": PositionBottom,
		})
		inst.SetAnchorBounds(34, 8, 4, 1)
		inst.SetViewportSize(40, 16)

		x, y := inst.CalculatePosition()
		if x != 26 || y != 10 {
			t.Fatalf("right-edge bottom-family position = (%d,%d), want (26,10)", x, y)
		}
	})

	t.Run("bottom placement stays below and shifts right within bottom family near left edge", func(t *testing.T) {
		inst := NewInstance(rtui.Props{
			"text":     "1234567890",
			"position": PositionBottom,
		})
		inst.SetAnchorBounds(2, 8, 4, 1)
		inst.SetViewportSize(40, 16)

		x, y := inst.CalculatePosition()
		if x != 2 || y != 10 {
			t.Fatalf("left-edge bottom-family position = (%d,%d), want (2,10)", x, y)
		}
	})

	t.Run("bottom right placement falls above within right family near bottom-right corner", func(t *testing.T) {
		inst := NewInstance(rtui.Props{
			"text":     "1234567890",
			"position": PositionBottomRight,
		})
		inst.SetAnchorBounds(34, 8, 4, 1)
		inst.SetViewportSize(40, 10)

		x, y := inst.CalculatePosition()
		if x != 26 || y != 6 {
			t.Fatalf("bottom-right corner fallback position = (%d,%d), want (26,6)", x, y)
		}
	})

	t.Run("bottom right placement clamps left and stays below in narrow viewport", func(t *testing.T) {
		inst := NewInstance(rtui.Props{
			"text":     "12345678901234",
			"position": PositionBottomRight,
		})
		inst.SetAnchorBounds(9, 7, 4, 1)
		inst.SetViewportSize(14, 14)

		x, y := inst.CalculatePosition()
		if x != 0 || y != 9 {
			t.Fatalf("narrow bottom-right clamped position = (%d,%d), want (0,9)", x, y)
		}
	})

	t.Run("bottom left placement clamps left and stays below in narrow viewport", func(t *testing.T) {
		inst := NewInstance(rtui.Props{
			"text":     "12345678901234",
			"position": PositionBottomLeft,
		})
		inst.SetAnchorBounds(9, 7, 4, 1)
		inst.SetViewportSize(14, 14)

		x, y := inst.CalculatePosition()
		if x != 0 || y != 9 {
			t.Fatalf("narrow bottom-left clamped position = (%d,%d), want (0,9)", x, y)
		}
	})

	t.Run("bottom left placement clamps both axes and stays below when no vertical candidate fits", func(t *testing.T) {
		inst := NewInstance(rtui.Props{
			"text":     "12345678901234",
			"position": PositionBottomLeft,
		})
		inst.SetAnchorBounds(9, 1, 4, 1)
		inst.SetViewportSize(14, 3)

		x, y := inst.CalculatePosition()
		if x != 0 || y != 2 {
			t.Fatalf("dual-axis bottom-left clamped position = (%d,%d), want (0,2)", x, y)
		}
	})

	t.Run("bottom right placement clamps both axes and stays below when no vertical candidate fits", func(t *testing.T) {
		inst := NewInstance(rtui.Props{
			"text":     "12345678901234",
			"position": PositionBottomRight,
		})
		inst.SetAnchorBounds(9, 1, 4, 1)
		inst.SetViewportSize(14, 3)

		x, y := inst.CalculatePosition()
		if x != 0 || y != 2 {
			t.Fatalf("dual-axis bottom-right clamped position = (%d,%d), want (0,2)", x, y)
		}
	})

	t.Run("bottom left placement falls above within left family near bottom-left corner", func(t *testing.T) {
		inst := NewInstance(rtui.Props{
			"text":     "1234567890",
			"position": PositionBottomLeft,
		})
		inst.SetAnchorBounds(2, 8, 4, 1)
		inst.SetViewportSize(40, 10)

		x, y := inst.CalculatePosition()
		if x != 2 || y != 6 {
			t.Fatalf("bottom-left corner fallback position = (%d,%d), want (2,6)", x, y)
		}
	})

	t.Run("right bottom placement prefers mirrored horizontal family before vertical fallback", func(t *testing.T) {
		inst := NewInstance(rtui.Props{
			"text":     "Test",
			"position": PositionRightBottom,
		})
		inst.SetAnchorBounds(16, 1, 3, 3)
		inst.SetViewportSize(20, 10)

		x, y := inst.CalculatePosition()
		if x != 9 || y != 3 {
			t.Fatalf("fallback position = (%d,%d), want (9,3)", x, y)
		}
	})

	t.Run("right placement falls back to left family when right edge clips", func(t *testing.T) {
		inst := NewInstance(rtui.Props{
			"text":     "Test",
			"position": PositionRight,
		})
		inst.SetAnchorBounds(17, 4, 2, 1)
		inst.SetViewportSize(20, 10)

		x, y := inst.CalculatePosition()
		if x != 10 || y != 4 {
			t.Fatalf("fallback position = (%d,%d), want (10,4)", x, y)
		}
	})

	t.Run("left placement falls back to right family when left edge clips", func(t *testing.T) {
		inst := NewInstance(rtui.Props{
			"text":     "Test",
			"position": PositionLeft,
		})
		inst.SetAnchorBounds(1, 4, 2, 1)
		inst.SetViewportSize(20, 10)

		x, y := inst.CalculatePosition()
		if x != 4 || y != 4 {
			t.Fatalf("fallback position = (%d,%d), want (4,4)", x, y)
		}
	})

	t.Run("right top placement falls back to mirrored left top family near corner", func(t *testing.T) {
		inst := NewInstance(rtui.Props{
			"text":     "Corner",
			"position": PositionRightTop,
		})
		inst.SetAnchorBounds(15, 1, 2, 1)
		inst.SetViewportSize(20, 10)

		x, y := inst.CalculatePosition()
		if x != 6 || y != 1 {
			t.Fatalf("fallback position = (%d,%d), want (6,1)", x, y)
		}
	})

	t.Run("right top placement falls back to top before clamp when both horizontal families overflow", func(t *testing.T) {
		inst := NewInstance(rtui.Props{
			"text":     "123456",
			"position": PositionRightTop,
		})
		inst.SetAnchorBounds(4, 3, 2, 1)
		inst.SetViewportSize(10, 8)

		x, y := inst.CalculatePosition()
		if x != 1 || y != 1 {
			t.Fatalf("vertical fallback position = (%d,%d), want (1,1)", x, y)
		}
	})

	t.Run("left bottom placement falls back to mirrored right bottom family near corner", func(t *testing.T) {
		inst := NewInstance(rtui.Props{
			"text":     "Corner",
			"position": PositionLeftBottom,
		})
		inst.SetAnchorBounds(1, 7, 2, 2)
		inst.SetViewportSize(20, 10)

		x, y := inst.CalculatePosition()
		if x != 4 || y != 8 {
			t.Fatalf("fallback position = (%d,%d), want (4,8)", x, y)
		}
	})

	t.Run("left bottom placement falls back to bottom before clamp when both horizontal families overflow", func(t *testing.T) {
		inst := NewInstance(rtui.Props{
			"text":     "123456",
			"position": PositionLeftBottom,
		})
		inst.SetAnchorBounds(4, 3, 2, 1)
		inst.SetViewportSize(10, 8)

		x, y := inst.CalculatePosition()
		if x != 1 || y != 5 {
			t.Fatalf("vertical fallback position = (%d,%d), want (1,5)", x, y)
		}
	})

	t.Run("clamps when nothing fully fits", func(t *testing.T) {
		inst := NewInstance(rtui.Props{
			"text":     "LongTooltip",
			"position": PositionRightBottom,
		})
		inst.SetAnchorBounds(8, 4, 2, 2)
		inst.SetViewportSize(8, 5)

		x, y := inst.CalculatePosition()
		if x != 0 || y != 4 {
			t.Fatalf("clamped position = (%d,%d), want (0,4)", x, y)
		}
	})

	t.Run("corner placement clamps when mirrored families and vertical fallbacks still overflow", func(t *testing.T) {
		inst := NewInstance(rtui.Props{
			"text":     "LongTooltip",
			"position": PositionRightTop,
		})
		inst.SetAnchorBounds(6, 0, 1, 1)
		inst.SetViewportSize(8, 1)

		x, y := inst.CalculatePosition()
		if x != 0 || y != 0 {
			t.Fatalf("clamped position = (%d,%d), want (0,0)", x, y)
		}
	})
}

func TestTooltipPositionCandidatesReuseSharedPlacementHelpers(t *testing.T) {
	tests := []struct {
		name     string
		position Position
		want     []overlayposition.Placement
	}{
		{
			name:     "top",
			position: PositionTop,
			want:     overlayposition.VerticalPlacementCandidates(overlayposition.PlacementTop),
		},
		{
			name:     "top left",
			position: PositionTopLeft,
			want:     overlayposition.VerticalPlacementCandidates(overlayposition.PlacementTopLeft),
		},
		{
			name:     "top right",
			position: PositionTopRight,
			want:     overlayposition.VerticalPlacementCandidates(overlayposition.PlacementTopRight),
		},
		{
			name:     "bottom",
			position: PositionBottom,
			want:     overlayposition.VerticalPlacementCandidates(overlayposition.PlacementBottom),
		},
		{
			name:     "bottom left",
			position: PositionBottomLeft,
			want:     overlayposition.VerticalPlacementCandidates(overlayposition.PlacementBottomLeft),
		},
		{
			name:     "bottom right",
			position: PositionBottomRight,
			want:     overlayposition.VerticalPlacementCandidates(overlayposition.PlacementBottomRight),
		},
		{
			name:     "left",
			position: PositionLeft,
			want:     overlayposition.HorizontalPlacementCandidates(overlayposition.PlacementLeft),
		},
		{
			name:     "left top",
			position: PositionLeftTop,
			want:     overlayposition.HorizontalPlacementCandidates(overlayposition.PlacementLeftTop),
		},
		{
			name:     "left bottom",
			position: PositionLeftBottom,
			want:     overlayposition.HorizontalPlacementCandidates(overlayposition.PlacementLeftBottom),
		},
		{
			name:     "right",
			position: PositionRight,
			want:     overlayposition.HorizontalPlacementCandidates(overlayposition.PlacementRight),
		},
		{
			name:     "right top",
			position: PositionRightTop,
			want:     overlayposition.HorizontalPlacementCandidates(overlayposition.PlacementRightTop),
		},
		{
			name:     "right bottom",
			position: PositionRightBottom,
			want:     overlayposition.HorizontalPlacementCandidates(overlayposition.PlacementRightBottom),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inst := NewInstance(rtui.Props{"text": "Test", "position": tt.position})
			if got := inst.positionCandidates(); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("positionCandidates() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTooltipHandleActionHoverLifecycle(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"text":  "Test",
		"delay": time.Duration(0),
	})

	inst.HandleAction(action.NewAction(action.ActionMouseEnter))
	if !inst.visible {
		t.Fatal("tooltip should become visible on mouse enter when delay is zero")
	}

	inst.HandleAction(action.NewAction(action.ActionMouseLeave))
	if inst.visible {
		t.Fatal("tooltip should hide on mouse leave")
	}
}

func TestTooltipHandleActionDelayedShowWaitsForTimer(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"text":  "Test",
		"delay": 20 * time.Millisecond,
	})

	inst.HandleAction(action.NewAction(action.ActionMouseEnter))
	if inst.visible {
		t.Fatal("tooltip should stay hidden until delay elapses")
	}
	if !inst.WantsTick() {
		t.Fatal("tooltip should request ticks while waiting for delayed show")
	}

	start := time.Unix(0, 0)
	inst.beginDelayAt(start)
	if changed := inst.Tick(start.Add(10 * time.Millisecond)); changed {
		t.Fatal("tooltip should stay hidden before delay elapses")
	}
	if changed := inst.Tick(start.Add(30 * time.Millisecond)); !changed {
		t.Fatal("tooltip should become visible after delay elapses")
	}
	if !inst.visible {
		t.Fatal("tooltip should become visible after delay elapses")
	}
	if inst.WantsTick() {
		t.Fatal("visible tooltip should stop requesting ticks")
	}
}

func TestTooltipHandleActionMouseLeaveCancelsPendingShow(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"text":  "Test",
		"delay": 30 * time.Millisecond,
	})

	start := time.Unix(0, 0)
	inst.beginDelayAt(start)
	inst.triggerActive = true
	if changed := inst.Tick(start.Add(10 * time.Millisecond)); changed {
		t.Fatal("tooltip should stay hidden before the leave action")
	}
	inst.HandleAction(action.NewAction(action.ActionMouseLeave))
	if inst.WantsTick() {
		t.Fatal("tooltip should stop requesting ticks after mouse leave")
	}

	if inst.visible {
		t.Fatal("tooltip should remain hidden when hover exits before delay elapses")
	}
}

func TestTooltipRuntimeChildrenVisible(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"text":     "Test",
		"position": PositionTop,
	})
	inst.SetBounds(10, 4, 8, 1)
	inst.Show()

	children := inst.RuntimeChildren()
	if len(children) != 1 {
		t.Fatalf("runtime children = %d, want 1", len(children))
	}
	if children[0].GetLayer() != rtui.LayerTooltip {
		t.Fatalf("runtime child layer = %v, want %v", children[0].GetLayer(), rtui.LayerTooltip)
	}
}

func TestTooltipRuntimeChildrenFollowHoveredChildState(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"text":     "Test",
		"position": PositionTop,
		"delay":    time.Duration(0),
	})
	child := &tooltipMockChild{
		key:    "anchor",
		state:  control.InteractionState{Hovered: true},
		bounds: [4]int{12, 5, 8, 1},
	}
	inst.AddChild(child)

	children := inst.RuntimeChildren()
	if !inst.visible {
		t.Fatal("tooltip should become visible when a child control is hovered")
	}
	if len(children) != 1 {
		t.Fatalf("runtime children = %d, want 1", len(children))
	}
	if got := getBoundsProp(children[0].Props(), propAnchorBounds, [4]int{}); got != child.bounds {
		t.Fatalf("overlay anchor bounds = %v, want %v", got, child.bounds)
	}

	child.state.Hovered = false
	if children := inst.RuntimeChildren(); len(children) != 0 {
		t.Fatalf("runtime children = %d, want 0 after child hover clears", len(children))
	}
	if inst.visible {
		t.Fatal("tooltip should hide when hovered child clears")
	}
}

func TestTooltipRuntimeChildrenHoveredChildRespectsDelay(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"text":     "Test",
		"position": PositionTop,
		"delay":    20 * time.Millisecond,
	})

	child := &tooltipMockChild{
		key:    "anchor",
		state:  control.InteractionState{Hovered: true},
		bounds: [4]int{12, 5, 8, 1},
	}
	inst.AddChild(child)

	if children := inst.RuntimeChildren(); len(children) != 0 {
		t.Fatalf("runtime children = %d, want 0 before delay elapses", len(children))
	}
	if inst.visible {
		t.Fatal("tooltip should stay hidden before delayed child hover resolves")
	}
	if !inst.WantsTick() {
		t.Fatal("tooltip should request ticks while hovered child delay is pending")
	}

	start := time.Unix(0, 0)
	inst.beginDelayAt(start)
	if changed := inst.Tick(start.Add(10 * time.Millisecond)); changed {
		t.Fatal("tooltip should not become visible before delayed child hover resolves")
	}
	if changed := inst.Tick(start.Add(30 * time.Millisecond)); !changed {
		t.Fatal("tooltip should become visible after delayed child hover resolves")
	}
	children := inst.RuntimeChildren()
	if len(children) != 1 {
		t.Fatalf("runtime children = %d, want 1 after delay elapses", len(children))
	}
	if !inst.visible {
		t.Fatal("tooltip should become visible after delayed child hover resolves")
	}

	child.state.Hovered = false
	if children := inst.RuntimeChildren(); len(children) != 0 {
		t.Fatalf("runtime children = %d, want 0 after child hover clears", len(children))
	}
	if inst.visible {
		t.Fatal("tooltip should hide when delayed child hover clears")
	}
}

type tooltipMockChild struct {
	key    string
	props  rtui.Props
	state  control.InteractionState
	bounds [4]int
	dirty  bool
	parent rtui.ComponentInstance
}

func (m *tooltipMockChild) Key() string                        { return m.key }
func (m *tooltipMockChild) SetKey(key string)                  { m.key = key }
func (m *tooltipMockChild) Init(props rtui.Props)              { m.props = props }
func (m *tooltipMockChild) Destroy()                           {}
func (m *tooltipMockChild) OnMount()                           {}
func (m *tooltipMockChild) OnUnmount()                         {}
func (m *tooltipMockChild) SetProps(props rtui.Props) bool     { m.props = props; return true }
func (m *tooltipMockChild) GetProps() rtui.Props               { return m.props }
func (m *tooltipMockChild) MarkDirty()                         { m.dirty = true }
func (m *tooltipMockChild) IsDirty() bool                      { return m.dirty }
func (m *tooltipMockChild) GetContext() *rtui.ComponentContext { return nil }
func (m *tooltipMockChild) GetState() *control.InteractionState {
	return &m.state
}
func (m *tooltipMockChild) GetBounds() (int, int, int, int) {
	return m.bounds[0], m.bounds[1], m.bounds[2], m.bounds[3]
}
func (m *tooltipMockChild) SetParent(parent rtui.ComponentInstance) { m.parent = parent }
