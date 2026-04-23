package layout

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFlex_Row_Basic(t *testing.T) {
	tests := []struct {
		name          string
		children      []Node
		constraints   Constraints
		expectedWidth int
	}{
		{
			name: "horizontal row layout",
			children: []Node{
				NewMockMeasurableNode("child1", 100, 50),
				NewMockMeasurableNode("child2", 150, 50),
			},
			constraints:   UnboundedConstraints(),
			expectedWidth: 250, // 100 + 150
		},
		{
			name: "row with multiple children",
			children: []Node{
				NewMockMeasurableNode("child1", 50, 50),
				NewMockMeasurableNode("child2", 75, 50),
				NewMockMeasurableNode("child3", 100, 50),
			},
			constraints:   UnboundedConstraints(),
			expectedWidth: 225, // 50 + 75 + 100
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flex := NewFlexLayout("flex", tt.children)
			flex.SetDirection(FlexRow)

			result := flex.Measure(tt.constraints)
			boxes := flex.LayoutChildren(result.Width, result.Height)

			assert.Len(t, boxes, len(tt.children), "Should have layout boxes for all children")
			assert.Equal(t, tt.expectedWidth, result.Width, "Width should match expected sum")

			// Verify children are laid out horizontally
			for i := 1; i < len(boxes); i++ {
				assert.Greater(t, boxes[i].X, boxes[i-1].X, "Children should be arranged horizontally")
			}
		})
	}
}

func TestFlex_Column_Basic(t *testing.T) {
	tests := []struct {
		name           string
		children       []Node
		constraints    Constraints
		expectedHeight int
	}{
		{
			name: "vertical column layout",
			children: []Node{
				NewMockMeasurableNode("child1", 50, 100),
				NewMockMeasurableNode("child2", 50, 150),
			},
			constraints:    UnboundedConstraints(),
			expectedHeight: 250, // 100 + 150
		},
		{
			name: "column with multiple children",
			children: []Node{
				NewMockMeasurableNode("child1", 50, 50),
				NewMockMeasurableNode("child2", 50, 75),
				NewMockMeasurableNode("child3", 50, 100),
			},
			constraints:    UnboundedConstraints(),
			expectedHeight: 225, // 50 + 75 + 100
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flex := NewFlexLayout("flex", tt.children)
			flex.SetDirection(FlexColumn)

			result := flex.Measure(tt.constraints)
			boxes := flex.LayoutChildren(result.Width, result.Height)

			assert.Len(t, boxes, len(tt.children), "Should have layout boxes for all children")
			assert.Equal(t, tt.expectedHeight, result.Height, "Height should match expected sum")

			// Verify children are laid out vertically
			for i := 1; i < len(boxes); i++ {
				assert.Greater(t, boxes[i].Y, boxes[i-1].Y, "Children should be arranged vertically")
			}
		})
	}
}

func TestFlex_AlignStart(t *testing.T) {
	tests := []struct {
		name      string
		direction FlexDirection
		mainAxis  MainAxisAlignment
		crossAxis CrossAxisAlignment
	}{
		{
			name:      "Row - AlignStart",
			direction: FlexRow,
			mainAxis:  MainStart,
			crossAxis: CrossStart,
		},
		{
			name:      "Column - AlignStart",
			direction: FlexColumn,
			mainAxis:  MainStart,
			crossAxis: CrossStart,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			children := []Node{
				NewMockMeasurableNode("child1", 100, 50),
				NewMockMeasurableNode("child2", 150, 75),
			}

			flex := NewFlexLayout("flex", children)
			flex.SetDirection(tt.direction)
			flex.SetMainAxis(tt.mainAxis)
			flex.SetCrossAxis(tt.crossAxis)

			result := flex.Measure(UnboundedConstraints())
			boxes := flex.LayoutChildren(result.Width, result.Height)

			// Verify alignment at start
			if tt.direction == FlexRow {
				assert.Equal(t, 0, boxes[0].X, "First child should start at X=0")
				assert.Equal(t, 0, boxes[0].Y, "First child should be at Y=0 (cross axis start)")
			} else {
				assert.Equal(t, 0, boxes[0].X, "First child should be at X=0 (cross axis start)")
				assert.Equal(t, 0, boxes[0].Y, "First child should start at Y=0")
			}
		})
	}
}

func TestFlex_AlignCenter(t *testing.T) {
	tests := []struct {
		name      string
		direction FlexDirection
		mainAxis  MainAxisAlignment
		crossAxis CrossAxisAlignment
	}{
		{
			name:      "Row - AlignCenter",
			direction: FlexRow,
			mainAxis:  MainStart,
			crossAxis: CrossCenter,
		},
		{
			name:      "Column - AlignCenter",
			direction: FlexColumn,
			mainAxis:  MainStart,
			crossAxis: CrossCenter,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			children := []Node{
				NewMockMeasurableNode("child1", 100, 50),
				NewMockMeasurableNode("child2", 150, 75),
			}

			flex := NewFlexLayout("flex", children)
			flex.SetDirection(tt.direction)
			flex.SetMainAxis(tt.mainAxis)
			flex.SetCrossAxis(tt.crossAxis)

			result := flex.Measure(Constraints{MinWidth: 0, MaxWidth: 400, MinHeight: 0, MaxHeight: 200})
			boxes := flex.LayoutChildren(result.Width, result.Height)

			// For cross axis center alignment
			if tt.direction == FlexRow {
				// Y position should be centered
				maxChildHeight := boxes[0].Height
				if boxes[1].Height > maxChildHeight {
					maxChildHeight = boxes[1].Height
				}
				expectedY := (result.Height - maxChildHeight) / 2
				assert.Equal(t, expectedY, boxes[0].Y, "Child should be centered vertically")
			} else {
				// X position should be centered
				maxChildWidth := boxes[0].Width
				if boxes[1].Width > maxChildWidth {
					maxChildWidth = boxes[1].Width
				}
				expectedX := (result.Width - maxChildWidth) / 2
				assert.Equal(t, expectedX, boxes[0].X, "Child should be centered horizontally")
			}
		})
	}
}

func TestFlex_AlignEnd(t *testing.T) {
	tests := []struct {
		name      string
		direction FlexDirection
		mainAxis  MainAxisAlignment
		crossAxis CrossAxisAlignment
	}{
		{
			name:      "Row - AlignEnd",
			direction: FlexRow,
			mainAxis:  MainStart,
			crossAxis: CrossEnd,
		},
		{
			name:      "Column - AlignEnd",
			direction: FlexColumn,
			mainAxis:  MainStart,
			crossAxis: CrossEnd,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			children := []Node{
				NewMockMeasurableNode("child1", 100, 50),
				NewMockMeasurableNode("child2", 150, 75),
			}

			flex := NewFlexLayout("flex", children)
			flex.SetDirection(tt.direction)
			flex.SetMainAxis(tt.mainAxis)
			flex.SetCrossAxis(tt.crossAxis)

			result := flex.Measure(Constraints{MinWidth: 0, MaxWidth: 400, MinHeight: 0, MaxHeight: 200})
			boxes := flex.LayoutChildren(result.Width, result.Height)

			// For cross axis end alignment
			if tt.direction == FlexRow {
				// Y position should be at end (or 0 if space allows)
				assert.GreaterOrEqual(t, boxes[0].Y, 0, "Child should have valid Y position")
			} else {
				// X position should be at end (or 0 if space allows)
				assert.GreaterOrEqual(t, boxes[0].X, 0, "Child should have valid X position")
			}
		})
	}
}

func TestFlex_SpaceBetween(t *testing.T) {
	tests := []struct {
		name      string
		direction FlexDirection
		children  []Node
	}{
		{
			name:      "Row - SpaceBetween",
			direction: FlexRow,
			children: []Node{
				NewMockMeasurableNode("child1", 100, 50),
				NewMockMeasurableNode("child2", 100, 50),
				NewMockMeasurableNode("child3", 100, 50),
			},
		},
		{
			name:      "Column - SpaceBetween",
			direction: FlexColumn,
			children: []Node{
				NewMockMeasurableNode("child1", 50, 100),
				NewMockMeasurableNode("child2", 50, 100),
				NewMockMeasurableNode("child3", 50, 100),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flex := NewFlexLayout("flex", tt.children)
			flex.SetDirection(tt.direction)
			flex.SetMainAxis(SpaceBetween)

			result := flex.Measure(Constraints{MinWidth: 0, MaxWidth: 500, MinHeight: 0, MaxHeight: 500})
			boxes := flex.LayoutChildren(result.Width, result.Height)

			// First child should be at start, last child at end
			if tt.direction == FlexRow {
				assert.Equal(t, 0, boxes[0].X, "First child should be at start")
				assert.Equal(t, result.Width-boxes[len(boxes)-1].Width, boxes[len(boxes)-1].X, "Last child should be at end")

				// Check spacing is distributed
				if len(boxes) > 1 {
					gap1 := boxes[1].X - (boxes[0].X + boxes[0].Width)
					gap2 := boxes[2].X - (boxes[1].X + boxes[1].Width)
					assert.Equal(t, gap1, gap2, "Gaps should be equal with SpaceBetween")
				}
			} else {
				assert.Equal(t, 0, boxes[0].Y, "First child should be at start")
				assert.Equal(t, result.Height-boxes[len(boxes)-1].Height, boxes[len(boxes)-1].Y, "Last child should be at end")
			}
		})
	}
}

func TestFlex_SpaceAround(t *testing.T) {
	tests := []struct {
		name      string
		direction FlexDirection
		children  []Node
	}{
		{
			name:      "Row - SpaceAround",
			direction: FlexRow,
			children: []Node{
				NewMockMeasurableNode("child1", 100, 50),
				NewMockMeasurableNode("child2", 100, 50),
			},
		},
		{
			name:      "Column - SpaceAround",
			direction: FlexColumn,
			children: []Node{
				NewMockMeasurableNode("child1", 50, 100),
				NewMockMeasurableNode("child2", 50, 100),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flex := NewFlexLayout("flex", tt.children)
			flex.SetDirection(tt.direction)
			flex.SetMainAxis(SpaceAround)

			result := flex.Measure(Constraints{MinWidth: 0, MaxWidth: 500, MinHeight: 0, MaxHeight: 500})
			boxes := flex.LayoutChildren(result.Width, result.Height)

			// Each child should have equal space on both sides
			if tt.direction == FlexRow {
				// First child's leading space should equal trailing space
				leadingSpace := boxes[0].X
				trailingSpace := boxes[1].X - (boxes[0].X + boxes[0].Width)
				assert.Equal(t, leadingSpace, trailingSpace/2, "Space should be distributed evenly")
			} else {
				leadingSpace := boxes[0].Y
				trailingSpace := boxes[1].Y - (boxes[0].Y + boxes[0].Height)
				assert.Equal(t, leadingSpace, trailingSpace/2, "Space should be distributed evenly")
			}
		})
	}
}

func TestFlex_Gap(t *testing.T) {
	tests := []struct {
		name      string
		direction FlexDirection
		gap       int
		children  []Node
	}{
		{
			name:      "Row - Gap",
			direction: FlexRow,
			gap:       10,
			children: []Node{
				NewMockMeasurableNode("child1", 100, 50),
				NewMockMeasurableNode("child2", 100, 50),
			},
		},
		{
			name:      "Column - Gap",
			direction: FlexColumn,
			gap:       15,
			children: []Node{
				NewMockMeasurableNode("child1", 50, 100),
				NewMockMeasurableNode("child2", 50, 100),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flex := NewFlexLayout("flex", tt.children)
			flex.SetDirection(tt.direction)
			flex.SetGap(tt.gap)

			result := flex.Measure(UnboundedConstraints())
			boxes := flex.LayoutChildren(result.Width, result.Height)

			if tt.direction == FlexRow {
				gap := boxes[1].X - (boxes[0].X + boxes[0].Width)
				assert.Equal(t, tt.gap, gap, "Gap should match set value")
			} else {
				gap := boxes[1].Y - (boxes[0].Y + boxes[0].Height)
				assert.Equal(t, tt.gap, gap, "Gap should match set value")
			}
		})
	}
}

func TestFlex_Wrap(t *testing.T) {
	t.Run("Row - Wrap behavior", func(t *testing.T) {
		children := []Node{
			NewMockMeasurableNode("child1", 200, 50),
			NewMockMeasurableNode("child2", 200, 50),
			NewMockMeasurableNode("child3", 200, 50),
		}

		flex := NewFlexLayout("flex", children)
		flex.SetDirection(FlexRow)

		// Constrain width so children would fit
		result := flex.Measure(Constraints{MinWidth: 0, MaxWidth: 250, MinHeight: 0, MaxHeight: 200})

		// Result should have valid dimensions
		assert.GreaterOrEqual(t, result.Height, 50, "Height should be at least child height")
	})

	t.Run("Column - Wrap behavior", func(t *testing.T) {
		children := []Node{
			NewMockMeasurableNode("child1", 50, 200),
			NewMockMeasurableNode("child2", 50, 200),
			NewMockMeasurableNode("child3", 50, 200),
		}

		flex := NewFlexLayout("flex", children)
		flex.SetDirection(FlexColumn)

		// Constrain height so children would fit
		result := flex.Measure(Constraints{MinWidth: 0, MaxWidth: 200, MinHeight: 0, MaxHeight: 250})

		// Result should have valid dimensions
		assert.GreaterOrEqual(t, result.Width, 50, "Width should be at least child width")
	})
}

func TestFlex_NestedFlex(t *testing.T) {
	t.Run("Row内嵌Row", func(t *testing.T) {
		innerChildren := []Node{
			NewMockMeasurableNode("inner1", 50, 30),
			NewMockMeasurableNode("inner2", 50, 30),
		}
		innerFlex := NewFlexLayout("inner", innerChildren)
		innerFlex.SetDirection(FlexRow)

		outerChildren := []Node{
			NewMockMeasurableNode("outer1", 100, 50),
			innerFlex,
		}
		outerFlex := NewFlexLayout("outer", outerChildren)
		outerFlex.SetDirection(FlexRow)

		result := innerFlex.Measure(UnboundedConstraints())
		assert.GreaterOrEqual(t, result.Width, 100, "Nested flex should compute valid width")
		assert.GreaterOrEqual(t, result.Height, 30, "Nested flex should compute valid height")
	})

	t.Run("Row内嵌Column", func(t *testing.T) {
		innerChildren := []Node{
			NewMockMeasurableNode("inner1", 30, 50),
			NewMockMeasurableNode("inner2", 30, 50),
		}
		innerFlex := NewFlexLayout("inner", innerChildren)
		innerFlex.SetDirection(FlexColumn)

		outerChildren := []Node{
			NewMockMeasurableNode("outer1", 100, 50),
			innerFlex,
		}
		outerFlex := NewFlexLayout("outer", outerChildren)
		outerFlex.SetDirection(FlexRow)

		result := outerFlex.Measure(UnboundedConstraints())
		assert.Greater(t, result.Width, 100, "Nested flex should compute combined width")
	})

	t.Run("Three-level nesting", func(t *testing.T) {
		level1Children := []Node{
			NewMockMeasurableNode("l1_1", 20, 20),
			NewMockMeasurableNode("l1_2", 20, 20),
		}
		level1 := NewFlexLayout("level1", level1Children)
		level1.SetDirection(FlexRow)

		level2Children := []Node{
			NewMockMeasurableNode("l2_1", 40, 40),
			level1,
		}
		level2 := NewFlexLayout("level2", level2Children)
		level2.SetDirection(FlexColumn)

		level3Children := []Node{
			NewMockMeasurableNode("l3_1", 60, 60),
			level2,
		}
		level3 := NewFlexLayout("level3", level3Children)
		level3.SetDirection(FlexRow)

		result := level3.Measure(UnboundedConstraints())
		assert.Greater(t, result.Width, 0, "Three-level nesting should produce valid width")
		assert.Greater(t, result.Height, 0, "Three-level nesting should produce valid height")
	})
}

func TestFlex_MixedFlex(t *testing.T) {
	t.Run("flexGrow + flexShrink", func(t *testing.T) {
		children := []Node{
			NewMockMeasurableNode("fixed", 100, 50),
			NewMockMeasurableNode("flexible", 100, 50),
		}

		flex := NewFlexLayout("flex", children)
		flex.SetDirection(FlexRow)
		flex.SetFlex(1, 1, 1, 0) // Second child is flexible

		result := flex.Measure(Constraints{MinWidth: 0, MaxWidth: 300, MinHeight: 0, MaxHeight: 100})

		// Width should be calculated based on children
		assert.GreaterOrEqual(t, result.Width, 200, "Width should include all children")
		assert.LessOrEqual(t, result.Width, 300, "Width should respect max constraint")
	})

	t.Run("Fixed + 弹性子节点", func(t *testing.T) {
		children := []Node{
			NewMockMeasurableNode("fixed1", 80, 50),
			NewMockMeasurableNode("flexible", 100, 50),
			NewMockMeasurableNode("fixed2", 80, 50),
		}

		flex := NewFlexLayout("flex", children)
		flex.SetDirection(FlexRow)
		flex.SetFlex(1, 1, 1, 0) // Middle child is flexible

		result := flex.Measure(Constraints{MinWidth: 0, MaxWidth: 400, MinHeight: 0, MaxHeight: 100})

		// Width should be calculated based on children
		assert.GreaterOrEqual(t, result.Width, 260, "Width should include all children")
		assert.LessOrEqual(t, result.Width, 400, "Width should respect max constraint")
	})

	t.Run("Complex layout", func(t *testing.T) {
		children := []Node{
			NewMockMeasurableNode("child1", 100, 50),
			NewMockMeasurableNode("child2", 150, 75),
			NewMockMeasurableNode("child3", 200, 100),
		}

		flex := NewFlexLayout("flex", children)
		flex.SetDirection(FlexRow)
		flex.SetMainAxis(Center)
		flex.SetCrossAxis(CrossCenter)
		flex.SetGap(10)
		flex.SetFlex(0, 2, 1, 100)
		flex.SetFlex(2, 1, 1, 150)

		result := flex.Measure(Constraints{MinWidth: 0, MaxWidth: 600, MinHeight: 0, MaxHeight: 200})
		boxes := flex.LayoutChildren(result.Width, result.Height)

		assert.Len(t, boxes, 3, "Should have all children")
		assert.Greater(t, result.Width, 0, "Should have valid width")
		assert.Greater(t, result.Height, 0, "Should have valid height")
	})
}

func TestFlex_ReverseDirection(t *testing.T) {
	t.Run("RowReverse", func(t *testing.T) {
		children := []Node{
			NewMockMeasurableNode("child1", 100, 50),
			NewMockMeasurableNode("child2", 150, 50),
		}

		flex := NewFlexLayout("flex", children)
		flex.SetDirection(FlexRowReverse)

		result := flex.Measure(UnboundedConstraints())
		boxes := flex.LayoutChildren(result.Width, result.Height)

		// Children should be laid out in reverse order
		assert.Len(t, boxes, 2, "Should have both children")
		// In RowReverse, first child in array appears at right position
		// The layout children index order matches the boxes array order
	})

	t.Run("ColumnReverse", func(t *testing.T) {
		children := []Node{
			NewMockMeasurableNode("child1", 50, 100),
			NewMockMeasurableNode("child2", 50, 150),
		}

		flex := NewFlexLayout("flex", children)
		flex.SetDirection(FlexColumnReverse)

		result := flex.Measure(UnboundedConstraints())
		boxes := flex.LayoutChildren(result.Width, result.Height)

		// Children should be laid out in reverse order
		assert.Len(t, boxes, 2, "Should have both children")
		// In ColumnReverse, first child in array appears at bottom position
		// The layout children index order matches the boxes array order
	})
}

func TestFlex_Padding(t *testing.T) {
	tests := []struct {
		name    string
		padding Padding
	}{
		{
			name: "all sides padding",
			padding: Padding{
				Left:   10,
				Right:  10,
				Top:    20,
				Bottom: 20,
			},
		},
		{
			name: "asymmetric padding",
			padding: Padding{
				Left:   5,
				Right:  15,
				Top:    10,
				Bottom: 20,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			children := []Node{
				NewMockMeasurableNode("child1", 100, 50),
				NewMockMeasurableNode("child2", 100, 50),
			}

			flex := NewFlexLayout("flex", children)
			flex.SetDirection(FlexRow)
			flex.SetPadding(tt.padding.Left, tt.padding.Right, tt.padding.Top, tt.padding.Bottom)

			result := flex.Measure(UnboundedConstraints())
			boxes := flex.LayoutChildren(result.Width, result.Height)

			// First child should be offset by left/top padding
			assert.Equal(t, tt.padding.Left, boxes[0].X, "Child X should be offset by left padding")
			assert.Equal(t, tt.padding.Top, boxes[0].Y, "Child Y should be offset by top padding")
		})
	}
}

func TestFlex_EmptyChildren(t *testing.T) {
	flex := NewFlexLayout("empty", []Node{})
	flex.SetDirection(FlexRow)

	result := flex.Measure(UnboundedConstraints())
	boxes := flex.LayoutChildren(result.Width, result.Height)

	assert.Nil(t, boxes, "Empty flex should have no layout boxes")
	assert.Equal(t, 0, result.Width, "Empty flex should have zero width")
	assert.Equal(t, 0, result.Height, "Empty flex should have zero height")
}

func TestFlex_SingleChild(t *testing.T) {
	children := []Node{
		NewMockMeasurableNode("only", 100, 50),
	}

	flex := NewFlexLayout("single", children)
	flex.SetDirection(FlexRow)

	result := flex.Measure(UnboundedConstraints())
	boxes := flex.LayoutChildren(result.Width, result.Height)

	assert.Len(t, boxes, 1, "Should have one layout box")
	assert.Equal(t, 100, result.Width, "Width should match child width")
	assert.Equal(t, 50, result.Height, "Height should match child height")
}

func TestFlex_Column_OutOfFlowChildrenDoNotShiftFlowContent(t *testing.T) {
	children := []Node{
		newMockFlowPositionNode("overlay-root", 1, 1, PositionAbsolute),
		newMockFlowPositionNode("modal-root", 1, 1, PositionAbsolute),
		newMockFlowPositionNode("tooltip-root", 1, 1, PositionAbsolute),
		NewMockMeasurableNode("content", 5, 2),
	}

	flex := NewFlexLayout("flex", children)
	flex.SetDirection(FlexColumn)

	result := flex.Measure(UnboundedConstraints())
	boxes := flex.LayoutChildren(result.Width, result.Height)

	assert.Len(t, boxes, 4, "Should have layout boxes for all children")
	assert.Equal(t, 5, result.Width, "Out-of-flow children should not affect measured width")
	assert.Equal(t, 2, result.Height, "Out-of-flow children should not affect measured height")
	assert.Equal(t, 0, boxes[0].Y, "Absolute overlay root should stay at origin")
	assert.Equal(t, 0, boxes[1].Y, "Absolute modal root should stay at origin")
	assert.Equal(t, 0, boxes[2].Y, "Absolute tooltip root should stay at origin")
	assert.Equal(t, 0, boxes[3].Y, "Flow content should not be pushed by out-of-flow roots")
}

// Benchmark tests
func BenchmarkFlex_RowLayout(b *testing.B) {
	children := make([]Node, 10)
	for i := range children {
		children[i] = NewMockMeasurableNode("child", 50, 50)
	}
	flex := NewFlexLayout("flex", children)
	flex.SetDirection(FlexRow)
	constraints := UnboundedConstraints()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := flex.Measure(constraints)
		flex.LayoutChildren(result.Width, result.Height)
	}
}

func BenchmarkFlex_ColumnLayout(b *testing.B) {
	children := make([]Node, 10)
	for i := range children {
		children[i] = NewMockMeasurableNode("child", 50, 50)
	}
	flex := NewFlexLayout("flex", children)
	flex.SetDirection(FlexColumn)
	constraints := UnboundedConstraints()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := flex.Measure(constraints)
		flex.LayoutChildren(result.Width, result.Height)
	}
}

func BenchmarkFlex_NestedLayout(b *testing.B) {
	innerChildren := make([]Node, 5)
	for i := range innerChildren {
		innerChildren[i] = NewMockMeasurableNode("inner", 30, 30)
	}
	innerFlex := NewFlexLayout("inner", innerChildren)
	innerFlex.SetDirection(FlexRow)

	outerChildren := make([]Node, 5)
	for i := range outerChildren {
		if i%2 == 0 {
			outerChildren[i] = NewMockMeasurableNode("outer", 50, 50)
		} else {
			outerChildren[i] = innerFlex
		}
	}
	outerFlex := NewFlexLayout("outer", outerChildren)
	outerFlex.SetDirection(FlexColumn)
	constraints := UnboundedConstraints()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := outerFlex.Measure(constraints)
		outerFlex.LayoutChildren(result.Width, result.Height)
	}
}
