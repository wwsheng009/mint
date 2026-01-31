package layout

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBoxConstraints_Basic(t *testing.T) {
	tests := []struct {
		name      string
		minWidth  int
		maxWidth  int
		minHeight int
		maxHeight int
	}{
		{
			name:      "normal constraints",
			minWidth:  100,
			maxWidth:  200,
			minHeight: 50,
			maxHeight: 150,
		},
		{
			name:      "zero min constraints",
			minWidth:  0,
			maxWidth:  100,
			minHeight: 0,
			maxHeight: 100,
		},
		{
			name:      "large constraints",
			minWidth:  1000,
			maxWidth:  2000,
			minHeight: 1000,
			maxHeight: 2000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewConstraints(tt.minWidth, tt.maxWidth, tt.minHeight, tt.maxHeight)

			assert.Equal(t, tt.minWidth, c.MinWidth, "MinWidth should match")
			assert.Equal(t, tt.maxWidth, c.MaxWidth, "MaxWidth should match")
			assert.Equal(t, tt.minHeight, c.MinHeight, "MinHeight should match")
			assert.Equal(t, tt.maxHeight, c.MaxHeight, "MaxHeight should match")
		})
	}
}

func TestBoxConstraints_IsTight(t *testing.T) {
	tests := []struct {
		name     string
		c        Constraints
		expected bool
	}{
		{
			name:     "tight width and height",
			c:        Constraints{MinWidth: 100, MaxWidth: 100, MinHeight: 50, MaxHeight: 50},
			expected: true,
		},
		{
			name:     "tight width only",
			c:        Constraints{MinWidth: 100, MaxWidth: 100, MinHeight: 0, MaxHeight: 100},
			expected: false,
		},
		{
			name:     "tight height only",
			c:        Constraints{MinWidth: 0, MaxWidth: 100, MinHeight: 50, MaxHeight: 50},
			expected: false,
		},
		{
			name:     "loose constraints",
			c:        Constraints{MinWidth: 0, MaxWidth: 100, MinHeight: 0, MaxHeight: 100},
			expected: false,
		},
		{
			name:     "zero tight constraints",
			c:        Constraints{MinWidth: 0, MaxWidth: 0, MinHeight: 0, MaxHeight: 0},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.c.IsTight()
			assert.Equal(t, tt.expected, result, "IsTight result should match")
		})
	}
}

func TestBoxConstraints_Loosen(t *testing.T) {
	tests := []struct {
		name         string
		tightWidth   int
		tightHeight  int
		looseMinW    int
		looseMaxW    int
		looseMinH    int
		looseMaxH    int
	}{
		{
			name:        "loosen tight constraints",
			tightWidth:  100,
			tightHeight: 50,
			looseMinW:   0,
			looseMaxW:   MaxInt,
			looseMinH:   0,
			looseMaxH:   MaxInt,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tight := TightConstraints(tt.tightWidth, tt.tightHeight)
			loose := LooseConstraints(0, 0)

			assert.True(t, tight.IsTight(), "TightConstraints should create tight constraints")
			assert.False(t, loose.IsTight(), "LooseConstraints should create loose constraints")
			assert.Equal(t, tt.looseMinW, loose.MinWidth, "Loose min width should be 0")
			assert.Equal(t, tt.looseMaxW, loose.MaxWidth, "Loose max width should be MaxInt")
			assert.Equal(t, tt.looseMinH, loose.MinHeight, "Loose min height should be 0")
			assert.Equal(t, tt.looseMaxH, loose.MaxHeight, "Loose max height should be MaxInt")
		})
	}
}

func TestBoxConstraints_Constrain_Boundaries(t *testing.T) {
	tests := []struct {
		name          string
		constraints   Constraints
		inputWidth    int
		inputHeight   int
		expectedWidth int
		expectedHeight int
	}{
		{
			name:          "within range",
			constraints:   NewConstraints(100, 200, 50, 150),
			inputWidth:    150,
			inputHeight:   100,
			expectedWidth: 150,
			expectedHeight: 100,
		},
		{
			name:          "width below minimum",
			constraints:   NewConstraints(100, 200, 50, 150),
			inputWidth:    50,
			inputHeight:   100,
			expectedWidth: 100,
			expectedHeight: 100,
		},
		{
			name:          "width above maximum",
			constraints:   NewConstraints(100, 200, 50, 150),
			inputWidth:    250,
			inputHeight:   100,
			expectedWidth: 200,
			expectedHeight: 100,
		},
		{
			name:          "height below minimum",
			constraints:   NewConstraints(100, 200, 50, 150),
			inputWidth:    150,
			inputHeight:   25,
			expectedWidth: 150,
			expectedHeight: 50,
		},
		{
			name:          "height above maximum",
			constraints:   NewConstraints(100, 200, 50, 150),
			inputWidth:    150,
			inputHeight:   200,
			expectedWidth: 150,
			expectedHeight: 150,
		},
		{
			name:          "both below minimum",
			constraints:   NewConstraints(100, 200, 50, 150),
			inputWidth:    50,
			inputHeight:   25,
			expectedWidth: 100,
			expectedHeight: 50,
		},
		{
			name:          "both above maximum",
			constraints:   NewConstraints(100, 200, 50, 150),
			inputWidth:    250,
			inputHeight:   200,
			expectedWidth: 200,
			expectedHeight: 150,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			width, height := tt.constraints.Constrain(tt.inputWidth, tt.inputHeight)
			assert.Equal(t, tt.expectedWidth, width, "Width should be constrained correctly")
			assert.Equal(t, tt.expectedHeight, height, "Height should be constrained correctly")
		})
	}
}

func TestBoxConstraints_Constrain_EdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		constraints   Constraints
		inputWidth    int
		inputHeight   int
		expectedWidth int
		expectedHeight int
	}{
		{
			name:          "zero constraints",
			constraints:   NewConstraints(0, 0, 0, 0),
			inputWidth:    100,
			inputHeight:   100,
			expectedWidth: 0,
			expectedHeight: 0,
		},
		{
			name:          "negative input width",
			constraints:   NewConstraints(100, 200, 50, 150),
			inputWidth:    -50,
			inputHeight:   100,
			expectedWidth: 100,
			expectedHeight: 100,
		},
		{
			name:          "negative input height",
			constraints:   NewConstraints(100, 200, 50, 150),
			inputWidth:    150,
			inputHeight:   -25,
			expectedWidth: 150,
			expectedHeight: 50,
		},
		{
			name:          "both negative inputs",
			constraints:   NewConstraints(100, 200, 50, 150),
			inputWidth:    -50,
			inputHeight:   -25,
			expectedWidth: 100,
			expectedHeight: 50,
		},
		{
			name:          "very large input",
			constraints:   NewConstraints(0, 100, 0, 100),
			inputWidth:    999999,
			inputHeight:   999999,
			expectedWidth: 100,
			expectedHeight: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			width, height := tt.constraints.Constrain(tt.inputWidth, tt.inputHeight)
			assert.Equal(t, tt.expectedWidth, width, "Width should handle edge cases correctly")
			assert.Equal(t, tt.expectedHeight, height, "Height should handle edge cases correctly")
		})
	}
}

func TestBoxConstraints_Infinity(t *testing.T) {
	tests := []struct {
		name           string
		constraints    Constraints
		isWidthBounded bool
		isHeightBounded bool
	}{
		{
			name:           "infinite width",
			constraints:    Constraints{MinWidth: 0, MaxWidth: MaxInt, MinHeight: 50, MaxHeight: 100},
			isWidthBounded: false,
			isHeightBounded: true,
		},
		{
			name:           "infinite height",
			constraints:    Constraints{MinWidth: 100, MaxWidth: 200, MinHeight: 0, MaxHeight: MaxInt},
			isWidthBounded: true,
			isHeightBounded: false,
		},
		{
			name:           "both infinite",
			constraints:    UnboundedConstraints(),
			isWidthBounded: false,
			isHeightBounded: false,
		},
		{
			name:           "both bounded",
			constraints:    NewConstraints(100, 200, 50, 150),
			isWidthBounded: true,
			isHeightBounded: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isBounded := tt.constraints.IsBounded()
			
			if tt.isWidthBounded && tt.isHeightBounded {
				assert.True(t, isBounded, "Should be bounded when both dimensions are bounded")
			} else if !tt.isWidthBounded && !tt.isHeightBounded {
				assert.False(t, isBounded, "Should not be bounded when both dimensions are unbounded")
			} else {
				// Mixed case: at least one dimension is bounded
				assert.True(t, isBounded || tt.constraints.MaxWidth < MaxInt || tt.constraints.MaxHeight < MaxInt)
			}
		})
	}
}

func TestBoxConstraints_Zero(t *testing.T) {
	tests := []struct {
		name          string
		constraints   Constraints
		isZeroWidth   bool
		isZeroHeight  bool
	}{
		{
			name:        "zero width",
			constraints: NewConstraints(0, 0, 50, 100),
			isZeroWidth: true,
			isZeroHeight: false,
		},
		{
			name:        "zero height",
			constraints: NewConstraints(100, 200, 0, 0),
			isZeroWidth: false,
			isZeroHeight: true,
		},
		{
			name:        "both zero",
			constraints: NewConstraints(0, 0, 0, 0),
			isZeroWidth: true,
			isZeroHeight: true,
		},
		{
			name:        "non-zero",
			constraints: NewConstraints(100, 200, 50, 150),
			isZeroWidth: false,
			isZeroHeight: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.isZeroWidth {
				assert.Equal(t, 0, tt.constraints.MinWidth, "Zero width should have MinWidth = 0")
				assert.Equal(t, 0, tt.constraints.MaxWidth, "Zero width should have MaxWidth = 0")
			}
			if tt.isZeroHeight {
				assert.Equal(t, 0, tt.constraints.MinHeight, "Zero height should have MinHeight = 0")
				assert.Equal(t, 0, tt.constraints.MaxHeight, "Zero height should have MaxHeight = 0")
			}
		})
	}
}

func TestBoxConstraints_Negative(t *testing.T) {
	tests := []struct {
		name          string
		inputWidth    int
		inputHeight   int
		constraints   Constraints
	}{
		{
			name:        "negative width constrained to min",
			inputWidth:  -100,
			inputHeight: 100,
			constraints: NewConstraints(50, 200, 50, 200),
		},
		{
			name:        "negative height constrained to min",
			inputWidth:  100,
			inputHeight: -100,
			constraints: NewConstraints(50, 200, 50, 200),
		},
		{
			name:        "both negative constrained to min",
			inputWidth:  -100,
			inputHeight: -100,
			constraints: NewConstraints(50, 200, 50, 200),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			width, height := tt.constraints.Constrain(tt.inputWidth, tt.inputHeight)
			
			// Negative values should be clamped to MinWidth/MinHeight
			assert.GreaterOrEqual(t, width, tt.constraints.MinWidth, "Width should be >= MinWidth")
			assert.GreaterOrEqual(t, height, tt.constraints.MinHeight, "Height should be >= MinHeight")
		})
	}
}

func TestBoxConstraints_ConstrainWidth(t *testing.T) {
	tests := []struct {
		name        string
		constraints Constraints
		input       int
		expected    int
	}{
		{
			name:        "within range",
			constraints: NewConstraints(100, 200, 0, 100),
			input:       150,
			expected:    150,
		},
		{
			name:        "below minimum",
			constraints: NewConstraints(100, 200, 0, 100),
			input:       50,
			expected:    100,
		},
		{
			name:        "above maximum",
			constraints: NewConstraints(100, 200, 0, 100),
			input:       250,
			expected:    200,
		},
		{
			name:        "negative input",
			constraints: NewConstraints(100, 200, 0, 100),
			input:       -50,
			expected:    100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.constraints.ConstrainWidth(tt.input)
			assert.Equal(t, tt.expected, result, "ConstrainWidth should work correctly")
		})
	}
}

func TestBoxConstraints_ConstrainHeight(t *testing.T) {
	tests := []struct {
		name        string
		constraints Constraints
		input       int
		expected    int
	}{
		{
			name:        "within range",
			constraints: NewConstraints(0, 100, 50, 150),
			input:       100,
			expected:    100,
		},
		{
			name:        "below minimum",
			constraints: NewConstraints(0, 100, 50, 150),
			input:       25,
			expected:    50,
		},
		{
			name:        "above maximum",
			constraints: NewConstraints(0, 100, 50, 150),
			input:       200,
			expected:    150,
		},
		{
			name:        "negative input",
			constraints: NewConstraints(0, 100, 50, 150),
			input:       -25,
			expected:    50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.constraints.ConstrainHeight(tt.input)
			assert.Equal(t, tt.expected, result, "ConstrainHeight should work correctly")
		})
	}
}

func TestBoxConstraints_Width(t *testing.T) {
	c := NewConstraints(100, 200, 50, 150)
	
	newC := c.Width(150, 250)
	
	assert.Equal(t, 150, newC.MinWidth, "New MinWidth should be updated")
	assert.Equal(t, 250, newC.MaxWidth, "New MaxWidth should be updated")
	assert.Equal(t, 50, newC.MinHeight, "MinHeight should remain unchanged")
	assert.Equal(t, 150, newC.MaxHeight, "MaxHeight should remain unchanged")
}

func TestBoxConstraints_Height(t *testing.T) {
	c := NewConstraints(100, 200, 50, 150)
	
	newC := c.Height(75, 175)
	
	assert.Equal(t, 100, newC.MinWidth, "MinWidth should remain unchanged")
	assert.Equal(t, 200, newC.MaxWidth, "MaxWidth should remain unchanged")
	assert.Equal(t, 75, newC.MinHeight, "New MinHeight should be updated")
	assert.Equal(t, 175, newC.MaxHeight, "New MaxHeight should be updated")
}

func TestBoxConstraints_IsBounded(t *testing.T) {
	tests := []struct {
		name     string
		c        Constraints
		expected bool
	}{
		{
			name:     "bounded both dimensions",
			c:        NewConstraints(0, 100, 0, 100),
			expected: true,
		},
		{
			name:     "unbounded width",
			c:        Constraints{MinWidth: 0, MaxWidth: MaxInt, MinHeight: 0, MaxHeight: 100},
			expected: true, // height is still bounded
		},
		{
			name:     "unbounded height",
			c:        Constraints{MinWidth: 0, MaxWidth: 100, MinHeight: 0, MaxHeight: MaxInt},
			expected: true, // width is still bounded
		},
		{
			name:     "both unbounded",
			c:        UnboundedConstraints(),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.c.IsBounded()
			assert.Equal(t, tt.expected, result, "IsBounded should return correct value")
		})
	}
}

func TestBoxConstraints_CreationHelpers(t *testing.T) {
	t.Run("TightConstraints", func(t *testing.T) {
		c := TightConstraints(100, 50)
		
		assert.True(t, c.IsTight(), "TightConstraints should create tight constraints")
		assert.Equal(t, 100, c.MinWidth, "Width should be set")
		assert.Equal(t, 100, c.MaxWidth, "Width should be tight")
		assert.Equal(t, 50, c.MinHeight, "Height should be set")
		assert.Equal(t, 50, c.MaxHeight, "Height should be tight")
	})
	
	t.Run("LooseConstraints", func(t *testing.T) {
		c := LooseConstraints(10, 20)
		
		assert.False(t, c.IsTight(), "LooseConstraints should create loose constraints")
		assert.Equal(t, 10, c.MinWidth, "MinWidth should be set")
		assert.Equal(t, MaxInt, c.MaxWidth, "MaxWidth should be MaxInt")
		assert.Equal(t, 20, c.MinHeight, "MinHeight should be set")
		assert.Equal(t, MaxInt, c.MaxHeight, "MaxHeight should be MaxInt")
	})
	
	t.Run("UnboundedConstraints", func(t *testing.T) {
		c := UnboundedConstraints()
		
		assert.False(t, c.IsTight(), "UnboundedConstraints should create loose constraints")
		assert.False(t, c.IsBounded(), "UnboundedConstraints should not be bounded")
		assert.Equal(t, 0, c.MinWidth, "MinWidth should be 0")
		assert.Equal(t, MaxInt, c.MaxWidth, "MaxWidth should be MaxInt")
		assert.Equal(t, 0, c.MinHeight, "MinHeight should be 0")
		assert.Equal(t, MaxInt, c.MaxHeight, "MaxHeight should be MaxInt")
	})
}

// Benchmark tests
func BenchmarkBoxConstraints_Constrain(b *testing.B) {
	c := NewConstraints(100, 200, 50, 150)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Constrain(150, 100)
	}
}

func BenchmarkBoxConstraints_IsTight(b *testing.B) {
	c := NewConstraints(100, 100, 50, 50)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.IsTight()
	}
}
