package layout

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// Margin Struct Tests
// =============================================================================

func TestMargin_Horizontal(t *testing.T) {
	tests := []struct {
		name     string
		margin   Margin
		expected int
	}{
		{"zero margin", Margin{}, 0},
		{"left only", Margin{Left: 5}, 5},
		{"right only", Margin{Right: 3}, 3},
		{"both", Margin{Left: 5, Right: 3}, 8},
		{"all sides", Margin{Left: 2, Right: 3, Top: 4, Bottom: 5}, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.margin.Horizontal())
		})
	}
}

func TestMargin_Vertical(t *testing.T) {
	tests := []struct {
		name     string
		margin   Margin
		expected int
	}{
		{"zero margin", Margin{}, 0},
		{"top only", Margin{Top: 5}, 5},
		{"bottom only", Margin{Bottom: 3}, 3},
		{"both", Margin{Top: 5, Bottom: 3}, 8},
		{"all sides", Margin{Left: 2, Right: 3, Top: 4, Bottom: 5}, 9},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.margin.Vertical())
		})
	}
}

// =============================================================================
// Marginal Interface Tests
// =============================================================================

// mockMarginalNode 实现 Marginal 接口
type mockMarginalNode struct {
	*MockNode
	margin Margin
}

func newMockMarginalNode(id string, width, height int, margin Margin) *mockMarginalNode {
	return &mockMarginalNode{
		MockNode: NewMockNode(id, width, height),
		margin:   margin,
	}
}

func (n *mockMarginalNode) GetMargin() Margin {
	return n.margin
}

func TestMarginal_Interface(t *testing.T) {
	node := newMockMarginalNode("test", 100, 50, Margin{Left: 5, Right: 5, Top: 10, Bottom: 10})

	// Test interface assertion
	var _ Marginal = node

	// Test GetMargin
	margin := node.GetMargin()
	assert.Equal(t, 5, margin.Left)
	assert.Equal(t, 5, margin.Right)
	assert.Equal(t, 10, margin.Top)
	assert.Equal(t, 10, margin.Bottom)
}

// =============================================================================
// MarginBox Tests
// =============================================================================

func TestMarginBox_ContentBox(t *testing.T) {
	tests := []struct {
		name     string
		box      *LayoutBox
		margin   Margin
		expected *LayoutBox
	}{
		{
			name: "no margin",
			box:  &LayoutBox{ID: "test", X: 0, Y: 0, Width: 100, Height: 50},
			margin: Margin{},
			expected: &LayoutBox{ID: "test", X: 0, Y: 0, Width: 100, Height: 50},
		},
		{
			name: "with margin",
			box:  &LayoutBox{ID: "test", X: 0, Y: 0, Width: 100, Height: 50},
			margin: Margin{Left: 5, Right: 5, Top: 10, Bottom: 10},
			expected: &LayoutBox{ID: "test", X: 5, Y: 10, Width: 90, Height: 30},
		},
		{
			name: "asymmetric margin",
			box:  &LayoutBox{ID: "test", X: 0, Y: 0, Width: 100, Height: 50},
			margin: Margin{Left: 10, Right: 5, Top: 5, Bottom: 15},
			expected: &LayoutBox{ID: "test", X: 10, Y: 5, Width: 85, Height: 30},
		},
		{
			name: "nil box",
			box:  nil,
			margin: Margin{Left: 5, Right: 5, Top: 10, Bottom: 10},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mb := &MarginBox{Box: tt.box, Margin: tt.margin}
			result := mb.ContentBox()

			if tt.expected == nil {
				assert.Nil(t, result)
				return
			}

			assert.Equal(t, tt.expected.ID, result.ID)
			assert.Equal(t, tt.expected.X, result.X)
			assert.Equal(t, tt.expected.Y, result.Y)
			assert.Equal(t, tt.expected.Width, result.Width)
			assert.Equal(t, tt.expected.Height, result.Height)
		})
	}
}

func TestMarginBox_BorderBox(t *testing.T) {
	box := &LayoutBox{ID: "test", X: 10, Y: 20, Width: 100, Height: 50}
	mb := &MarginBox{Box: box, Margin: Margin{Left: 5, Right: 5, Top: 10, Bottom: 10}}

	result := mb.BorderBox()
	assert.Equal(t, box, result)
}

// =============================================================================
// FlexStyle with Margin Tests
// =============================================================================

func TestFlexStyle_WithMargin(t *testing.T) {
	style := DefaultFlexStyle()
	assert.Equal(t, Margin{}, style.Margin)

	// Set margin
	style.Margin = Margin{Left: 5, Right: 5, Top: 10, Bottom: 10}
	assert.Equal(t, 10, style.Margin.Horizontal())
	assert.Equal(t, 20, style.Margin.Vertical())
}

// =============================================================================
// Margin in Layout Calculation Tests
// =============================================================================

func TestLayout_WithMargin(t *testing.T) {
	// Create a child with margin
	child := newMockMarginalNode("child", 50, 30, Margin{Left: 5, Right: 5, Top: 10, Bottom: 10})

	// Verify the child implements Marginal
	assert.True(t, isMarginal(child))

	// Get margin from child
	margin := getMargin(child)
	assert.Equal(t, 10, margin.Horizontal())
	assert.Equal(t, 20, margin.Vertical())
}

// Helper functions for margin handling

// isMarginal checks if a node implements Marginal interface
func isMarginal(node Node) bool {
	_, ok := node.(Marginal)
	return ok
}

// getMargin returns the margin of a node, or zero margin if not implemented
func getMargin(node Node) Margin {
	if m, ok := node.(Marginal); ok {
		return m.GetMargin()
	}
	return Margin{}
}

// =============================================================================
// Edge Cases
// =============================================================================

func TestMargin_NegativeValues(t *testing.T) {
	// Negative margins are technically allowed (overlap behavior)
	margin := Margin{Left: -5, Right: -3, Top: -2, Bottom: -1}
	assert.Equal(t, -8, margin.Horizontal())
	assert.Equal(t, -3, margin.Vertical())
}

func TestMargin_LargeValues(t *testing.T) {
	margin := Margin{Left: 1000, Right: 1000, Top: 500, Bottom: 500}
	assert.Equal(t, 2000, margin.Horizontal())
	assert.Equal(t, 1000, margin.Vertical())
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkMargin_Horizontal(b *testing.B) {
	margin := Margin{Left: 10, Right: 20, Top: 5, Bottom: 5}
	for i := 0; i < b.N; i++ {
		_ = margin.Horizontal()
	}
}

func BenchmarkMargin_Vertical(b *testing.B) {
	margin := Margin{Left: 10, Right: 20, Top: 5, Bottom: 5}
	for i := 0; i < b.N; i++ {
		_ = margin.Vertical()
	}
}

func BenchmarkMarginBox_ContentBox(b *testing.B) {
	box := &LayoutBox{ID: "test", X: 0, Y: 0, Width: 100, Height: 50}
	margin := Margin{Left: 5, Right: 5, Top: 10, Bottom: 10}
	mb := &MarginBox{Box: box, Margin: margin}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mb.ContentBox()
	}
}

func BenchmarkGetMargin(b *testing.B) {
	node := newMockMarginalNode("test", 100, 50, Margin{Left: 5, Right: 5, Top: 10, Bottom: 10})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = getMargin(node)
	}
}
