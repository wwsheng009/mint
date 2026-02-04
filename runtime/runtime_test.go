package runtime

import (
	"strings"
	"testing"

	"github.com/wwsheng009/mint/runtime/style"
)

// =============================================================================
// BoxConstraints Tests
// =============================================================================

func TestBoxConstraints_IsTight(t *testing.T) {
	tests := []struct {
		name     string
		c        BoxConstraints
		expected bool
	}{
		{
			name:     "tight constraints",
			c:        BoxConstraints{MinWidth: 100, MaxWidth: 100, MinHeight: 50, MaxHeight: 50},
			expected: true,
		},
		{
			name:     "loose width",
			c:        BoxConstraints{MinWidth: 0, MaxWidth: 100, MinHeight: 50, MaxHeight: 50},
			expected: false,
		},
		{
			name:     "loose height",
			c:        BoxConstraints{MinWidth: 100, MaxWidth: 100, MinHeight: 0, MaxHeight: 50},
			expected: false,
		},
		{
			name:     "loose both",
			c:        BoxConstraints{MinWidth: 0, MaxWidth: 100, MinHeight: 0, MaxHeight: 50},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.IsTight(); got != tt.expected {
				t.Errorf("IsTight() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestBoxConstraints_Constrain(t *testing.T) {
	c := BoxConstraints{MinWidth: 50, MaxWidth: 100, MinHeight: 30, MaxHeight: 60}

	tests := []struct {
		name               string
		width, height      int
		expectW, expectH int
	}{
		{"within bounds", 75, 45, 75, 45},
		{"below min", 25, 20, 50, 30},
		{"above max", 150, 80, 100, 60},
		{"mixed", 25, 80, 50, 60},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, h := c.Constrain(tt.width, tt.height)
			if w != tt.expectW || h != tt.expectH {
				t.Errorf("Constrain(%d,%d) = (%d,%d), want (%d,%d)",
					tt.width, tt.height, w, h, tt.expectW, tt.expectH)
			}
		})
	}
}

func TestBoxConstraints_Loosen(t *testing.T) {
	c := BoxConstraints{MinWidth: 50, MaxWidth: 100, MinHeight: 30, MaxHeight: 60}
	loose := c.Loosen()

	if loose.MinWidth != 0 {
		t.Errorf("Loosen() MinWidth = %d, want 0", loose.MinWidth)
	}
	if loose.MinHeight != 0 {
		t.Errorf("Loosen() MinHeight = %d, want 0", loose.MinHeight)
	}
	if loose.MaxWidth != 100 {
		t.Errorf("Loosen() MaxWidth = %d, want 100", loose.MaxWidth)
	}
	if loose.MaxHeight != 60 {
		t.Errorf("Loosen() MaxHeight = %d, want 60", loose.MaxHeight)
	}
}

func TestConstraints_HasInfiniteWidth(t *testing.T) {
	tests := []struct {
		name     string
		c        Constraints
		expected bool
	}{
		{"finite width", NewConstraints(100, 50), false},
		{"infinite width", BoxConstraints{MaxWidth: -1}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.HasInfiniteWidth(); got != tt.expected {
				t.Errorf("HasInfiniteWidth() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestConstraints_HasInfiniteHeight(t *testing.T) {
	tests := []struct {
		name     string
		c        Constraints
		expected bool
	}{
		{"finite height", NewConstraints(100, 50), false},
		{"infinite height", BoxConstraints{MaxHeight: -1}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.HasInfiniteHeight(); got != tt.expected {
				t.Errorf("HasInfiniteHeight() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// =============================================================================
// Position Tests
// =============================================================================

func TestNewPosition(t *testing.T) {
	p := NewPosition()
	if p.Type != PositionRelative {
		t.Errorf("NewPosition() Type = %s, want PositionRelative", p.Type)
	}
	if p.Top != nil || p.Left != nil || p.Right != nil || p.Bottom != nil {
		t.Error("NewPosition() offsets should be nil")
	}
}

func TestPosition_SetAbsolute(t *testing.T) {
	p := NewPosition()
	top, left, right, bottom := 10, 20, 30, 40
	p.SetAbsolute(&top, &left, &right, &bottom)

	if p.Type != PositionAbsolute {
		t.Errorf("SetAbsolute() Type = %s, want PositionAbsolute", p.Type)
	}
	if *p.Top != 10 || *p.Left != 20 || *p.Right != 30 || *p.Bottom != 40 {
		t.Error("SetAbsolute() offsets not set correctly")
	}
}

func TestPosition_SetAbsolute_NilValues(t *testing.T) {
	p := NewPosition()
	p.SetAbsolute(nil, nil, nil, nil)

	if p.Type != PositionAbsolute {
		t.Errorf("SetAbsolute() Type = %s, want PositionAbsolute", p.Type)
	}
	if p.Top != nil || p.Left != nil || p.Right != nil || p.Bottom != nil {
		t.Error("SetAbsolute() with nil should keep nil values")
	}
}

// =============================================================================
// Frame Tests
// =============================================================================

func TestFrame_String_NilBuffer(t *testing.T) {
	f := Frame{Buffer: nil}
	if got := f.String(); got != "" {
		t.Errorf("String() with nil buffer = %q, want empty string", got)
	}
}

func TestFrame_String_WithBuffer(t *testing.T) {
	f := Frame{
		Buffer: &CellBuffer{
			Width:  2,
			Height: 2,
			Cells:  [][]Cell{{{Cluster: "a"}, {Cluster: "b"}}, {{Cluster: "c"}, {Cluster: "d"}}},
		},
	}
	got := f.String()
	// Check that we get 2 lines with content
	lines := countLines(got)
	if lines != 2 {
		t.Errorf("String() should have 2 lines, got %d: %q", lines, got)
	}
	// Verify content
	if !contains(got, "ab") || !contains(got, "cd") {
		t.Errorf("String() = %q, should contain 'ab' and 'cd'", got)
	}
}

// =============================================================================
// SetContentRuntime Tests
// =============================================================================

func TestSetContentRuntime_OutOfBounds(t *testing.T) {
	b := &CellBuffer{Width: 10, Height: 10}
	// Should not panic with out of bounds coordinates
	SetContentRuntime(b, -1, 0, 0, 'a', false, false, false, "test")
	SetContentRuntime(b, 0, -1, 0, 'a', false, false, false, "test")
	SetContentRuntime(b, 10, 0, 0, 'a', false, false, false, "test")
	SetContentRuntime(b, 0, 10, 0, 'a', false, false, false, "test")
}

func TestSetContentRuntime_ZIndex(t *testing.T) {
	b := &CellBuffer{
		Width:  2,
		Height: 2,
		Cells:  make([][]Cell, 2),
	}
	b.Cells[0] = []Cell{
		{Cluster: "old", ZIndex: 5},
		{Cluster: "old", ZIndex: 0},
	}

	// Lower Z-index should not overwrite
	SetContentRuntime(b, 0, 0, 3, 'n', false, false, false, "test")
	if b.Cells[0][0].Cluster != "old" {
		t.Error("lower Z-index should not overwrite existing cell")
	}

	// Equal Z-index should overwrite
	SetContentRuntime(b, 0, 0, 5, 'n', false, false, false, "test")
	if b.Cells[0][0].Cluster != "n" {
		t.Error("equal Z-index should overwrite existing cell")
	}

	// Higher Z-index should overwrite
	SetContentRuntime(b, 1, 0, 1, 'x', false, false, false, "test")
	if b.Cells[0][1].Cluster != "x" {
		t.Error("higher Z-index should overwrite existing cell")
	}
}

func TestSetContentRuntime_ContentAndNodeID(t *testing.T) {
	b := &CellBuffer{
		Width:  1,
		Height: 1,
		Cells:  make([][]Cell, 1),
	}
	b.Cells[0] = make([]Cell, 1)

	SetContentRuntime(b, 0, 0, 0, 'a', true, true, true, "test-node")
	cell := b.Cells[0][0]

	if cell.Cluster != "a" {
		t.Errorf("Cluster = %s, want 'a'", cell.Cluster)
	}
	if cell.ZIndex != 0 {
		t.Errorf("ZIndex = %d, want 0", cell.ZIndex)
	}
	if cell.NodeID != "test-node" {
		t.Errorf("NodeID = %s, want 'test-node'", cell.NodeID)
	}
	// Verify style is set (we can't check unexported fields, but we can check the style is not default)
	s := style.Style{}
	if cell.Style == s {
		t.Error("Style should be set to non-default")
	}
}

func TestSetContentRuntime_DoesNotPanic(t *testing.T) {
	b := &CellBuffer{
		Width:  5,
		Height: 5,
		Cells:  make([][]Cell, 5),
	}
	for i := range b.Cells {
		b.Cells[i] = make([]Cell, 5)
	}

	// Should not panic with various combinations
	SetContentRuntime(b, 2, 2, 0, 'x', false, false, false, "node1")
	SetContentRuntime(b, 2, 2, 1, 'y', true, false, false, "node2")
	SetContentRuntime(b, 4, 4, 0, 'z', false, true, true, "node3")
}

// =============================================================================
// Helper functions
// =============================================================================

func countLines(s string) int {
	if s == "" {
		return 0
	}
	return strings.Count(s, "\n") + 1
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
