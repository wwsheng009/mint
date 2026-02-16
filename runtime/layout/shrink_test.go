package layout

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// CalculateShrinkDistribution Tests
// =============================================================================

func TestCalculateShrinkDistribution_NoDeficit(t *testing.T) {
	children := []ShrinkInfo{
		{Index: 0, OriginalSize: 100, ShrinkFactor: 1},
	}
	result := CalculateShrinkDistribution(0, children)
	assert.Empty(t, result)
}

func TestCalculateShrinkDistribution_NoChildren(t *testing.T) {
	result := CalculateShrinkDistribution(50, nil)
	assert.Empty(t, result)

	result = CalculateShrinkDistribution(50, []ShrinkInfo{})
	assert.Empty(t, result)
}

func TestCalculateShrinkDistribution_SingleChild(t *testing.T) {
	children := []ShrinkInfo{
		{Index: 0, OriginalSize: 100, ShrinkFactor: 1, MinSize: 0},
	}
	result := CalculateShrinkDistribution(30, children)

	assert.Equal(t, 30, result[0])
}

func TestCalculateShrinkDistribution_MultipleChildren(t *testing.T) {
	children := []ShrinkInfo{
		{Index: 0, OriginalSize: 100, ShrinkFactor: 1, MinSize: 0},
		{Index: 1, OriginalSize: 100, ShrinkFactor: 1, MinSize: 0},
	}
	result := CalculateShrinkDistribution(60, children)

	// Each should shrink by 30 (equal weight)
	assert.Equal(t, 30, result[0])
	assert.Equal(t, 30, result[1])
}

func TestCalculateShrinkDistribution_WeightedBySize(t *testing.T) {
	children := []ShrinkInfo{
		{Index: 0, OriginalSize: 200, ShrinkFactor: 1, MinSize: 0}, // 2x weight
		{Index: 1, OriginalSize: 100, ShrinkFactor: 1, MinSize: 0}, // 1x weight
	}
	result := CalculateShrinkDistribution(90, children)

	// Total weight = 200 + 100 = 300
	// Child 0: (90 * 200) / 300 = 60
	// Child 1: (90 * 100) / 300 = 30
	assert.Equal(t, 60, result[0])
	assert.Equal(t, 30, result[1])
}

func TestCalculateShrinkDistribution_WeightedByFactor(t *testing.T) {
	children := []ShrinkInfo{
		{Index: 0, OriginalSize: 100, ShrinkFactor: 2, MinSize: 0}, // 2x weight
		{Index: 1, OriginalSize: 100, ShrinkFactor: 1, MinSize: 0}, // 1x weight
	}
	result := CalculateShrinkDistribution(60, children)

	// Total weight = 2*100 + 1*100 = 300
	// Child 0: (60 * 200) / 300 = 40
	// Child 1: (60 * 100) / 300 = 20
	assert.Equal(t, 40, result[0])
	assert.Equal(t, 20, result[1])
}

func TestCalculateShrinkDistribution_RespectMinSize(t *testing.T) {
	children := []ShrinkInfo{
		{Index: 0, OriginalSize: 100, ShrinkFactor: 1, MinSize: 80}, // Can only shrink 20
		{Index: 1, OriginalSize: 100, ShrinkFactor: 1, MinSize: 0},
	}
	result := CalculateShrinkDistribution(60, children)

	// Child 0 would get 30, but max shrink is 20
	assert.Equal(t, 20, result[0])
	// Child 1 should get remaining
	assert.GreaterOrEqual(t, result[1], 30)
}

func TestCalculateShrinkDistribution_AllAtMinSize(t *testing.T) {
	children := []ShrinkInfo{
		{Index: 0, OriginalSize: 100, ShrinkFactor: 1, MinSize: 100}, // Already at min
	}
	result := CalculateShrinkDistribution(50, children)

	// Should return 0 since child can't shrink
	assert.Equal(t, 0, result[0])
}

func TestCalculateShrinkDistribution_ZeroShrinkFactor(t *testing.T) {
	children := []ShrinkInfo{
		{Index: 0, OriginalSize: 100, ShrinkFactor: 0, MinSize: 0}, // Can't shrink
	}
	result := CalculateShrinkDistribution(50, children)

	assert.Empty(t, result)
}

// =============================================================================
// ApplyShrinkToSizes Tests
// =============================================================================

func TestApplyShrinkToSizes_Horizontal(t *testing.T) {
	sizes := []Size{
		{Width: 100, Height: 50},
		{Width: 200, Height: 50},
	}
	shrinkAmounts := map[int]int{
		0: 20,
		1: 40,
	}

	ApplyShrinkToSizes(sizes, shrinkAmounts, true)

	assert.Equal(t, 80, sizes[0].Width)
	assert.Equal(t, 50, sizes[0].Height)
	assert.Equal(t, 160, sizes[1].Width)
	assert.Equal(t, 50, sizes[1].Height)
}

func TestApplyShrinkToSizes_Vertical(t *testing.T) {
	sizes := []Size{
		{Width: 50, Height: 100},
		{Width: 50, Height: 200},
	}
	shrinkAmounts := map[int]int{
		0: 20,
		1: 40,
	}

	ApplyShrinkToSizes(sizes, shrinkAmounts, false)

	assert.Equal(t, 50, sizes[0].Width)
	assert.Equal(t, 80, sizes[0].Height)
	assert.Equal(t, 50, sizes[1].Width)
	assert.Equal(t, 160, sizes[1].Height)
}

func TestApplyShrinkToSizes_NoNegative(t *testing.T) {
	sizes := []Size{
		{Width: 10, Height: 50},
	}
	shrinkAmounts := map[int]int{
		0: 50, // More than size
	}

	ApplyShrinkToSizes(sizes, shrinkAmounts, true)

	assert.Equal(t, 0, sizes[0].Width) // Capped at 0
}

func TestApplyShrinkToSizes_EmptyMap(t *testing.T) {
	sizes := []Size{
		{Width: 100, Height: 50},
	}
	shrinkAmounts := map[int]int{}

	ApplyShrinkToSizes(sizes, shrinkAmounts, true)

	assert.Equal(t, 100, sizes[0].Width)
}

// =============================================================================
// GetShrinkableChildren Tests
// =============================================================================

func TestGetShrinkableChildren_Basic(t *testing.T) {
	children := []Node{
		NewMockNode("a", 100, 50),
		NewMockNode("b", 100, 50),
	}
	flexConfig := map[int]*Flex{
		0: {Grow: 0, Shrink: 1, Basis: 0},
		1: {Grow: 0, Shrink: 2, Basis: 0},
	}
	sizes := []Size{
		{Width: 100, Height: 50},
		{Width: 100, Height: 50},
	}

	shrinkable := GetShrinkableChildren(children, flexConfig, sizes, true)

	assert.Len(t, shrinkable, 2)
	assert.Equal(t, 0, shrinkable[0].Index)
	assert.Equal(t, 1, shrinkable[1].Index)
	assert.Equal(t, 1, shrinkable[0].ShrinkFactor)
	assert.Equal(t, 2, shrinkable[1].ShrinkFactor)
}

func TestGetShrinkableChildren_SomeNonShrinkable(t *testing.T) {
	children := []Node{
		NewMockNode("a", 100, 50),
		NewMockNode("b", 100, 50),
		NewMockNode("c", 100, 50),
	}
	flexConfig := map[int]*Flex{
		0: {Grow: 0, Shrink: 1, Basis: 0},
		1: {Grow: 0, Shrink: 0, Basis: 0}, // Not shrinkable
		2: {Grow: 0, Shrink: 1, Basis: 0},
	}
	sizes := []Size{
		{Width: 100, Height: 50},
		{Width: 100, Height: 50},
		{Width: 100, Height: 50},
	}

	shrinkable := GetShrinkableChildren(children, flexConfig, sizes, true)

	assert.Len(t, shrinkable, 2)
	assert.Equal(t, 0, shrinkable[0].Index)
	assert.Equal(t, 2, shrinkable[1].Index)
}

func TestGetShrinkableChildren_NoFlexConfig(t *testing.T) {
	children := []Node{
		NewMockNode("a", 100, 50),
	}
	sizes := []Size{
		{Width: 100, Height: 50},
	}

	shrinkable := GetShrinkableChildren(children, nil, sizes, true)

	assert.Empty(t, shrinkable)
}

func TestGetShrinkableChildren_NilChild(t *testing.T) {
	children := []Node{
		nil,
		NewMockNode("b", 100, 50),
	}
	flexConfig := map[int]*Flex{
		0: {Grow: 0, Shrink: 1, Basis: 0},
		1: {Grow: 0, Shrink: 1, Basis: 0},
	}
	sizes := []Size{
		{Width: 100, Height: 50},
		{Width: 100, Height: 50},
	}

	shrinkable := GetShrinkableChildren(children, flexConfig, sizes, true)

	assert.Len(t, shrinkable, 1)
	assert.Equal(t, 1, shrinkable[0].Index)
}

// =============================================================================
// Integration Tests
// =============================================================================

func TestShrinkIntegration(t *testing.T) {
	// Simulate a scenario where we have 2 children that need to shrink
	children := []Node{
		NewMockNode("a", 150, 50),
		NewMockNode("b", 100, 50),
	}
	flexConfig := map[int]*Flex{
		0: {Grow: 0, Shrink: 1, Basis: 0},
		1: {Grow: 0, Shrink: 2, Basis: 0},
	}
	sizes := []Size{
		{Width: 150, Height: 50},
		{Width: 100, Height: 50},
	}

	// Get shrinkable children
	shrinkable := GetShrinkableChildren(children, flexConfig, sizes, true)
	assert.Len(t, shrinkable, 2)

	// Calculate distribution
	deficit := 75
	distribution := CalculateShrinkDistribution(deficit, shrinkable)

	// Apply shrink
	ApplyShrinkToSizes(sizes, distribution, true)

	// Verify total shrink equals deficit (approximately)
	totalShrink := 0
	for _, shrink := range distribution {
		totalShrink += shrink
	}
	assert.GreaterOrEqual(t, totalShrink, deficit-2) // Allow small rounding error
}

// =============================================================================
// Edge Cases
// =============================================================================

func TestCalculateShrinkDistribution_LargeDeficit(t *testing.T) {
	children := []ShrinkInfo{
		{Index: 0, OriginalSize: 100, ShrinkFactor: 1, MinSize: 0},
	}
	result := CalculateShrinkDistribution(500, children)

	// Should only shrink to min size (0)
	assert.Equal(t, 100, result[0])
}

func TestCalculateShrinkDistribution_MixedMinSizes(t *testing.T) {
	children := []ShrinkInfo{
		{Index: 0, OriginalSize: 100, ShrinkFactor: 1, MinSize: 50}, // Can shrink 50
		{Index: 1, OriginalSize: 100, ShrinkFactor: 1, MinSize: 90}, // Can shrink 10
	}
	result := CalculateShrinkDistribution(100, children)

	// Child 0 can shrink more
	assert.GreaterOrEqual(t, result[0], 50)
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkCalculateShrinkDistribution(b *testing.B) {
	children := []ShrinkInfo{
		{Index: 0, OriginalSize: 100, ShrinkFactor: 1, MinSize: 0},
		{Index: 1, OriginalSize: 100, ShrinkFactor: 2, MinSize: 0},
		{Index: 2, OriginalSize: 100, ShrinkFactor: 1, MinSize: 0},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CalculateShrinkDistribution(50, children)
	}
}

func BenchmarkApplyShrinkToSizes(b *testing.B) {
	sizes := []Size{
		{Width: 100, Height: 50},
		{Width: 100, Height: 50},
		{Width: 100, Height: 50},
	}
	shrinkAmounts := map[int]int{
		0: 10,
		1: 20,
		2: 15,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Reset sizes
		sizes[0].Width = 100
		sizes[1].Width = 100
		sizes[2].Width = 100
		ApplyShrinkToSizes(sizes, shrinkAmounts, true)
	}
}

func BenchmarkGetShrinkableChildren(b *testing.B) {
	children := []Node{
		NewMockNode("a", 100, 50),
		NewMockNode("b", 100, 50),
		NewMockNode("c", 100, 50),
	}
	flexConfig := map[int]*Flex{
		0: {Shrink: 1},
		1: {Shrink: 2},
		2: {Shrink: 1},
	}
	sizes := []Size{
		{Width: 100, Height: 50},
		{Width: 100, Height: 50},
		{Width: 100, Height: 50},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetShrinkableChildren(children, flexConfig, sizes, true)
	}
}
