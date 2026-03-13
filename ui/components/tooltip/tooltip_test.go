package tooltip

import (
	"testing"
	"time"

	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
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
		{"Bottom", NewBuilder(content, "text").Bottom(), PositionBottom},
		{"Left", NewBuilder(content, "text").Left(), PositionLeft},
		{"Right", NewBuilder(content, "text").Right(), PositionRight},
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
		name          string
		position      Position
		anchorX, anchorY, anchorW, anchorH int
		expectedX, expectedY int
	}{
		// Test text is "Test" (4 chars), tooltip width = 4 + 2 = 6
		{"Top", PositionTop, 10, 10, 20, 5, 17, 8},    // X = 10 + 10 - 3 = 17
		{"Bottom", PositionBottom, 10, 10, 20, 5, 17, 16}, // X = 10 + 10 - 3 = 17
		{"Left", PositionLeft, 10, 10, 20, 5, 3, 12},     // X = 10 - 6 - 1 = 3
		{"Right", PositionRight, 10, 10, 20, 5, 31, 12},  // X = 10 + 20 + 1 = 31
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

