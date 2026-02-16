package layout

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBoundsValidator(t *testing.T) {
	validator := NewBoundsValidator()

	assert.NotNil(t, validator)
	assert.Equal(t, 100, validator.maxOverlaps)
	assert.Equal(t, 1000, validator.maxDepth)
	assert.False(t, validator.strict)
}

func TestBoundsValidator_SetMaxOverlaps(t *testing.T) {
	validator := NewBoundsValidator()

	validator.SetMaxOverlaps(50)
	assert.Equal(t, 50, validator.maxOverlaps)
}

func TestBoundsValidator_SetMaxDepth(t *testing.T) {
	validator := NewBoundsValidator()

	validator.SetMaxDepth(500)
	assert.Equal(t, 500, validator.maxDepth)
}

func TestBoundsValidator_SetStrict(t *testing.T) {
	validator := NewBoundsValidator()

	validator.SetStrict(true)
	assert.True(t, validator.strict)
}

func TestBoundsValidator_ValidateBox_NonPositiveSize(t *testing.T) {
	validator := NewBoundsValidator()

	t.Run("zero width", func(t *testing.T) {
		box := &LayoutBox{
			ID:      "box1",
			Width:   0,
			Height:  10,
		}

		problems := validator.ValidateBox(box)
		assert.Len(t, problems, 1)
		assert.Equal(t, "NonPositiveWidth", problems[0].Problem)
		assert.Equal(t, SeverityError, problems[0].Severity)
	})

	t.Run("negative width", func(t *testing.T) {
		box := &LayoutBox{
			ID:      "box2",
			Width:   -10,
			Height:  10,
		}

		problems := validator.ValidateBox(box)
		assert.Len(t, problems, 1)
		assert.Equal(t, "NonPositiveWidth", problems[0].Problem)
	})

	t.Run("zero height", func(t *testing.T) {
		box := &LayoutBox{
			ID:      "box3",
			Width:   10,
			Height:  0,
		}

		problems := validator.ValidateBox(box)
		assert.Len(t, problems, 1)
		assert.Equal(t, "NonPositiveHeight", problems[0].Problem)
	})

	t.Run("negative height", func(t *testing.T) {
		box := &LayoutBox{
			ID:      "box4",
			Width:   10,
			Height:  -10,
		}

		problems := validator.ValidateBox(box)
		assert.Len(t, problems, 1)
		assert.Equal(t, "NonPositiveHeight", problems[0].Problem)
	})

	t.Run("both zero", func(t *testing.T) {
		box := &LayoutBox{
			ID:      "box5",
			Width:   0,
			Height:  0,
		}

		problems := validator.ValidateBox(box)
		assert.Len(t, problems, 2)
	})
}

func TestBoundsValidator_ValidateBox_NegativePosition(t *testing.T) {
	validator := NewBoundsValidator()
	validator.SetStrict(true)

	t.Run("negative X", func(t *testing.T) {
		box := &LayoutBox{
			ID:      "box1",
			X:       -10,
			Y:       0,
			Width:   10,
			Height:  10,
		}

		problems := validator.ValidateBox(box)
		assert.Len(t, problems, 1)
		assert.Equal(t, "NegativePosition", problems[0].Problem)
		assert.Equal(t, SeverityWarning, problems[0].Severity)
	})

	t.Run("negative Y", func(t *testing.T) {
		box := &LayoutBox{
			ID:      "box2",
			X:       0,
			Y:       -10,
			Width:   10,
			Height:  10,
		}

		problems := validator.ValidateBox(box)
		assert.Len(t, problems, 1)
		assert.Equal(t, "NegativePosition", problems[0].Problem)
	})

	t.Run("both negative", func(t *testing.T) {
		box := &LayoutBox{
			ID:      "box3",
			X:       -10,
			Y:       -10,
			Width:   10,
			Height:  10,
		}

		problems := validator.ValidateBox(box)
		assert.Len(t, problems, 2)
	})

	t.Run("not strict mode", func(t *testing.T) {
		validator := NewBoundsValidator()
		box := &LayoutBox{
			ID:      "box4",
			X:       -10,
			Y:       0,
			Width:   10,
			Height:  10,
		}

		problems := validator.ValidateBox(box)
		assert.Len(t, problems, 0)
	})
}

func TestBoundsValidator_ValidateBox_Nil(t *testing.T) {
	validator := NewBoundsValidator()

	problems := validator.ValidateBox(nil)
	assert.Len(t, problems, 0)
}

func TestBoundsValidator_ValidateTree_Overlaps(t *testing.T) {
	validator := NewBoundsValidator()

	t.Run("overlapping siblings", func(t *testing.T) {
		child1 := &LayoutBox{
			ID:      "child1",
			X:       10,
			Y:       10,
			Width:   50,
			Height:  30,
			Children: []*LayoutBox{},
		}

		child2 := &LayoutBox{
			ID:      "child2",
			X:       20,
			Y:       20,
			Width:   50,
			Height:  30,
			Children: []*LayoutBox{},
		}

		root := &LayoutBox{
			ID:      "root",
			X:       0,
			Y:       0,
			Width:   100,
			Height:  100,
			Children: []*LayoutBox{child1, child2},
		}

		problems := validator.ValidateTree(root)
		
		// Should have overlap warning
		found := false
		for _, p := range problems {
			if p.Problem == "Overlap" {
				found = true
				assert.Equal(t, SeverityWarning, p.Severity)
				break
			}
		}
		assert.True(t, found, "Should find overlap problem")
	})

	t.Run("no overlapping siblings", func(t *testing.T) {
		child1 := &LayoutBox{
			ID:      "child1",
			X:       0,
			Y:       0,
			Width:   50,
			Height:  50,
			Children: []*LayoutBox{},
		}

		child2 := &LayoutBox{
			ID:      "child2",
			X:       50,
			Y:       0,
			Width:   50,
			Height:  50,
			Children: []*LayoutBox{},
		}

		root := &LayoutBox{
			ID:      "root",
			X:       0,
			Y:       0,
			Width:   100,
			Height:  50,
			Children: []*LayoutBox{child1, child2},
		}

		problems := validator.ValidateTree(root)
		
		// Should not have overlap warnings
		for _, p := range problems {
			assert.NotEqual(t, "Overlap", p.Problem)
		}
	})
}

func TestBoundsValidator_ValidateTree_Depth(t *testing.T) {
	validator := NewBoundsValidator()
	validator.SetMaxDepth(5)

	t.Run("exceeds max depth", func(t *testing.T) {
		// Create a deep tree
		leaf := &LayoutBox{
			ID:      "leaf",
			Width:   10,
			Height:  10,
			Children: []*LayoutBox{},
		}

		// Build 6 levels
		current := leaf
		for i := 0; i < 6; i++ {
			current = &LayoutBox{
				ID:      "level" + string(rune('0'+i)),
				Width:   20,
				Height:  20,
				Children: []*LayoutBox{current},
			}
		}

		problems := validator.ValidateTree(current)
		
		found := false
		for _, p := range problems {
			if p.Problem == "MaxDepthExceeded" {
				found = true
				assert.Equal(t, SeverityError, p.Severity)
				break
			}
		}
		assert.True(t, found, "Should find max depth exceeded problem")
	})

	t.Run("within max depth", func(t *testing.T) {
		leaf := &LayoutBox{
			ID:      "leaf",
			Width:   10,
			Height:  10,
			Children: []*LayoutBox{},
		}

		// Build 3 levels
		current := leaf
		for i := 0; i < 3; i++ {
			current = &LayoutBox{
				ID:      "level" + string(rune('0'+i)),
				Width:   20,
				Height:  20,
				Children: []*LayoutBox{current},
			}
		}

		problems := validator.ValidateTree(current)
		
		// Should not have depth problems
		for _, p := range problems {
			assert.NotEqual(t, "MaxDepthExceeded", p.Problem)
		}
	})
}

func TestBoundsValidator_ValidateTree_Nil(t *testing.T) {
	validator := NewBoundsValidator()

	problems := validator.ValidateTree(nil)
	assert.Len(t, problems, 0)
}

func TestBoundsValidator_ValidateWithinConstraints(t *testing.T) {
	validator := NewBoundsValidator()

	t.Run("width exceeds max", func(t *testing.T) {
		box := &LayoutBox{
			ID:      "box1",
			Width:   150,
			Height:  50,
		}

		constraints := NewConstraints(0, 100, 0, 100)
		problems := validator.ValidateWithinConstraints(box, constraints)

		assert.Len(t, problems, 1)
		assert.Equal(t, "MaxWidthViolation", problems[0].Problem)
		assert.Equal(t, SeverityError, problems[0].Severity)
	})

	t.Run("width below min", func(t *testing.T) {
		box := &LayoutBox{
			ID:      "box2",
			Width:   50,
			Height:  50,
		}

		constraints := NewConstraints(100, 200, 0, 100)
		problems := validator.ValidateWithinConstraints(box, constraints)

		assert.Len(t, problems, 1)
		assert.Equal(t, "MinWidthViolation", problems[0].Problem)
	})

	t.Run("height exceeds max", func(t *testing.T) {
		box := &LayoutBox{
			ID:      "box3",
			Width:   50,
			Height:  150,
		}

		constraints := NewConstraints(0, 100, 0, 100)
		problems := validator.ValidateWithinConstraints(box, constraints)

		assert.Len(t, problems, 1)
		assert.Equal(t, "MaxHeightViolation", problems[0].Problem)
	})

	t.Run("height below min", func(t *testing.T) {
		box := &LayoutBox{
			ID:      "box4",
			Width:   50,
			Height:  50,
		}

		constraints := NewConstraints(0, 100, 100, 200)
		problems := validator.ValidateWithinConstraints(box, constraints)

		assert.Len(t, problems, 1)
		assert.Equal(t, "MinHeightViolation", problems[0].Problem)
	})

	t.Run("within constraints", func(t *testing.T) {
		box := &LayoutBox{
			ID:      "box5",
			Width:   50,
			Height:  50,
		}

		constraints := NewConstraints(0, 100, 0, 100)
		problems := validator.ValidateWithinConstraints(box, constraints)

		assert.Len(t, problems, 0)
	})

	t.Run("unbounded constraints", func(t *testing.T) {
		box := &LayoutBox{
			ID:      "box6",
			Width:   1000,
			Height:  1000,
		}

		constraints := UnboundedConstraints()
		problems := validator.ValidateWithinConstraints(box, constraints)

		assert.Len(t, problems, 0)
	})
}

func TestBoundsValidator_ValidateTree_Combined(t *testing.T) {
	validator := NewBoundsValidator()
	validator.SetStrict(true)

	// Create a tree with multiple issues
	child1 := &LayoutBox{
		ID:      "child1",
		X:       -10, // Negative position (warning in strict mode)
		Y:       0,
		Width:   0,  // Non-positive size (error)
		Height:  10,
		Children: []*LayoutBox{},
	}

	child2 := &LayoutBox{
		ID:      "child2",
		X:       20,
		Y:       20,
		Width:   50,
		Height:  30,
		Children: []*LayoutBox{},
	}

	root := &LayoutBox{
		ID:      "root",
		X:       0,
		Y:       0,
		Width:   100,
		Height:  100,
		Children: []*LayoutBox{child1, child2},
	}

	problems := validator.ValidateTree(root)

	// Should have multiple problems
	assert.Greater(t, len(problems), 0)

	// Count by severity
	errorCount := 0
	warningCount := 0
	for _, p := range problems {
		if p.Severity == SeverityError {
			errorCount++
		} else if p.Severity == SeverityWarning {
			warningCount++
		}
	}

	assert.Greater(t, errorCount, 0, "Should have at least one error")
	assert.Greater(t, warningCount, 0, "Should have at least one warning")
}

func TestSeverity_String(t *testing.T) {
	tests := []struct {
		severity   Severity
		expected string
	}{
		{SeverityInfo, "Info"},
		{SeverityWarning, "Warning"},
		{SeverityError, "Error"},
		{Severity(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.severity.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBoundsValidator_isAncestor(t *testing.T) {
	validator := NewBoundsValidator()

	t.Run("direct parent", func(t *testing.T) {
		child := &LayoutBox{
			ID:      "child",
			Children: []*LayoutBox{},
		}

		parent := &LayoutBox{
			ID:      "parent",
			Children: []*LayoutBox{child},
		}

		assert.True(t, validator.isAncestor(parent, child))
		assert.False(t, validator.isAncestor(child, parent))
	})

	t.Run("grandparent", func(t *testing.T) {
		leaf := &LayoutBox{
			ID:      "leaf",
			Children: []*LayoutBox{},
		}

		mid := &LayoutBox{
			ID:      "mid",
			Children: []*LayoutBox{leaf},
		}

		root := &LayoutBox{
			ID:      "root",
			Children: []*LayoutBox{mid},
		}

		assert.True(t, validator.isAncestor(root, leaf))
		assert.True(t, validator.isAncestor(mid, leaf))
		assert.False(t, validator.isAncestor(leaf, root))
	})

	t.Run("no relationship", func(t *testing.T) {
		box1 := &LayoutBox{
			ID:      "box1",
			Children: []*LayoutBox{},
		}

		box2 := &LayoutBox{
			ID:      "box2",
			Children: []*LayoutBox{},
		}

		assert.False(t, validator.isAncestor(box1, box2))
		assert.False(t, validator.isAncestor(box2, box1))
	})
}

// Benchmark tests
func BenchmarkBoundsValidator_ValidateBox(b *testing.B) {
	validator := NewBoundsValidator()
	box := &LayoutBox{
		ID:      "box",
		X:       10,
		Y:       10,
		Width:   100,
		Height:  50,
		Children: []*LayoutBox{},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.ValidateBox(box)
	}
}

func BenchmarkBoundsValidator_ValidateTree(b *testing.B) {
	validator := NewBoundsValidator()

	// Create a large tree
	children := make([]*LayoutBox, 100)
	for i := 0; i < 100; i++ {
		children[i] = &LayoutBox{
			ID:      "child" + string(rune('0'+i%10)),
			X:       0,
			Y:       i * 10,
			Width:   100,
			Height:  10,
			Children: []*LayoutBox{},
		}
	}

	root := &LayoutBox{
		ID:      "root",
		X:       0,
		Y:       0,
		Width:   100,
		Height:  1000,
		Children: children,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.ValidateTree(root)
	}
}

func BenchmarkBoundsValidator_ValidateWithinConstraints(b *testing.B) {
	validator := NewBoundsValidator()
	box := &LayoutBox{
		ID:      "box",
		Width:   50,
		Height:  50,
	}

	constraints := NewConstraints(0, 100, 0, 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.ValidateWithinConstraints(box, constraints)
	}
}
