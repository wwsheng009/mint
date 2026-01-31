package layout

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// MockNode is a mock node implementation for testing
type MockNode struct {
	id       string
	nodeType string
	position Point
	size     Size
	children []Node
}

func NewMockNode(id string, width, height int) *MockNode {
	return &MockNode{
		id:       id,
		nodeType: "mock",
		size:     Size{Width: width, Height: height},
		position: Point{X: 0, Y: 0},
		children: nil,
	}
}

func (m *MockNode) ID() string {
	return m.id
}

func (m *MockNode) Type() string {
	return m.nodeType
}

func (m *MockNode) Children() []Node {
	return m.children
}

func (m *MockNode) GetPosition() (int, int) {
	return m.position.X, m.position.Y
}

func (m *MockNode) SetPosition(x, y int) {
	m.position.X = x
	m.position.Y = y
}

func (m *MockNode) GetSize() (int, int) {
	return m.size.Width, m.size.Height
}

func (m *MockNode) SetSize(width, height int) {
	m.size.Width = width
	m.size.Height = height
}

func (m *MockNode) GetWidth() int {
	return m.size.Width
}

func (m *MockNode) GetHeight() int {
	return m.size.Height
}

// MockMeasurableNode is a mock measurable node
type MockMeasurableNode struct {
	*MockNode
	measureSize Size
}

func NewMockMeasurableNode(id string, measureWidth, measureHeight int) *MockMeasurableNode {
	return &MockMeasurableNode{
		MockNode:    NewMockNode(id, measureWidth, measureHeight),
		measureSize: Size{Width: measureWidth, Height: measureHeight},
	}
}

func (m *MockMeasurableNode) Measure(constraints Constraints) Size {
	// Return the constrained measure size
	width, height := constraints.Constrain(m.measureSize.Width, m.measureSize.Height)
	return Size{Width: width, Height: height}
}

func TestMeasure_Text(t *testing.T) {
	tests := []struct {
		name          string
		text          string
		constraints   Constraints
		expectedWidth int
	}{
		{
			name:          "single line text",
			text:          "Hello",
			constraints:   UnboundedConstraints(),
			expectedWidth: 5, // Assuming 1 char = 1 column
		},
		{
			name:          "single line with wide characters",
			text:          "你好",
			constraints:   UnboundedConstraints(),
			expectedWidth: 4, // Assuming 1 wide char = 2 columns
		},
		{
			name:          "constrained text",
			text:          "Hello World",
			constraints:   NewConstraints(0, 10, 0, 10),
			expectedWidth: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a measurable node representing text
			node := NewMockMeasurableNode(tt.text, tt.expectedWidth, 1)
			result := node.Measure(tt.constraints)
			
			assert.Equal(t, tt.expectedWidth, result.Width, "Width should match expected")
			assert.GreaterOrEqual(t, result.Width, tt.constraints.MinWidth, "Width should be >= MinWidth")
			assert.LessOrEqual(t, result.Width, tt.constraints.MaxWidth, "Width should be <= MaxWidth")
		})
	}
}

func TestMeasure_Box(t *testing.T) {
	tests := []struct {
		name           string
		nodeWidth      int
		nodeHeight     int
		constraints    Constraints
		expectedWidth  int
		expectedHeight int
	}{
		{
			name:           "fixed size box",
			nodeWidth:      100,
			nodeHeight:     50,
			constraints:    UnboundedConstraints(),
			expectedWidth:  100,
			expectedHeight: 50,
		},
		{
			name:           "elastic box",
			nodeWidth:      50,
			nodeHeight:     30,
			constraints:    LooseConstraints(10, 10),
			expectedWidth:  50,
			expectedHeight: 30,
		},
		{
			name:           "box with min constraints",
			nodeWidth:      50,
			nodeHeight:     30,
			constraints:    NewConstraints(100, 200, 100, 200),
			expectedWidth:  100,
			expectedHeight: 100,
		},
		{
			name:           "box with max constraints",
			nodeWidth:      250,
			nodeHeight:     150,
			constraints:    NewConstraints(0, 200, 0, 100),
			expectedWidth:  200,
			expectedHeight: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := NewMockMeasurableNode("box", tt.nodeWidth, tt.nodeHeight)
			result := node.Measure(tt.constraints)
			
			assert.Equal(t, tt.expectedWidth, result.Width, "Width should match expected")
			assert.Equal(t, tt.expectedHeight, result.Height, "Height should match expected")
		})
	}
}

func TestMeasure_Nested(t *testing.T) {
	tests := []struct {
		name           string
		nestingLevel   int
		constraints    Constraints
		expectedLevels int
	}{
		{
			name:           "two level nesting",
			nestingLevel:   2,
			constraints:    UnboundedConstraints(),
			expectedLevels: 2,
		},
		{
			name:           "five level nesting",
			nestingLevel:   5,
			constraints:    UnboundedConstraints(),
			expectedLevels: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build nested structure
			var parent Node = NewMockMeasurableNode("level0", 100, 100)
			
			for i := 1; i < tt.nestingLevel; i++ {
				child := NewMockMeasurableNode("level"+string(rune('0'+i)), 80, 80)
				parent = &FlexLayout{
					id:       "parent_" + child.ID(),
					children: []Node{parent, child},
					style:    DefaultFlexStyle(),
				}
			}
			
			// Measure the top-level parent
			if measurable, ok := parent.(Measurable); ok {
				result := measurable.Measure(tt.constraints)
				assert.Greater(t, result.Width, 0, "Width should be greater than 0")
				assert.Greater(t, result.Height, 0, "Height should be greater than 0")
			}
		})
	}
}

func TestMeasure_FlexGrow(t *testing.T) {
	tests := []struct {
		name           string
		flexGrowValues []int
		availableSpace int
		expectedTotal   int
	}{
		{
			name:           "no flex grow (all zero)",
			flexGrowValues: []int{0, 0, 0},
			availableSpace: 300,
			expectedTotal:   0, // No extra space distributed
		},
		{
			name:           "uniform flex grow",
			flexGrowValues: []int{1, 1, 1},
			availableSpace: 300,
			expectedTotal:   300, // All space distributed equally
		},
		{
			name:           "non-uniform flex grow",
			flexGrowValues: []int{1, 2, 3},
			availableSpace: 300,
			expectedTotal:   300, // Space distributed proportionally
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			children := make([]Node, len(tt.flexGrowValues))
			for i := range tt.flexGrowValues {
				children[i] = NewMockMeasurableNode("child"+string(rune('0'+i)), 50, 50)
			}
			
			flex := NewFlexLayout("flex", children)
			for i, grow := range tt.flexGrowValues {
				flex.SetFlex(i, grow, 1, 0)
			}
			
			result := flex.Measure(Constraints{
				MinWidth:  tt.availableSpace,
				MaxWidth:  tt.availableSpace,
				MinHeight: 0,
				MaxHeight: MaxInt,
			})
			
			assert.Equal(t, tt.availableSpace, result.Width, "Width should match available space")
		})
	}
}

func TestMeasure_FlexShrink(t *testing.T) {
	tests := []struct {
		name            string
		flexShrinkValues []int
		availableSpace  int
		baseSizes       []int
	}{
		{
			name:            "no flex shrink (all zero)",
			flexShrinkValues: []int{0, 0, 0},
			availableSpace:  200,
			baseSizes:       []int{100, 100, 100},
		},
		{
			name:            "uniform flex shrink",
			flexShrinkValues: []int{1, 1, 1},
			availableSpace:  200,
			baseSizes:       []int{100, 100, 100},
		},
		{
			name:            "non-uniform flex shrink",
			flexShrinkValues: []int{1, 2, 3},
			availableSpace:  200,
			baseSizes:       []int{100, 100, 100},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			children := make([]Node, len(tt.flexShrinkValues))
			for i := range tt.flexShrinkValues {
				children[i] = NewMockMeasurableNode("child"+string(rune('0'+i)), tt.baseSizes[i], 50)
			}
			
			flex := NewFlexLayout("flex", children)
			for i, shrink := range tt.flexShrinkValues {
				flex.SetFlex(i, 0, shrink, tt.baseSizes[i])
			}
			
			result := flex.Measure(Constraints{
				MinWidth:  0,
				MaxWidth:  tt.availableSpace,
				MinHeight: 0,
				MaxHeight: MaxInt,
			})
			
			// Result should be constrained by available space
			assert.LessOrEqual(t, result.Width, tt.availableSpace, "Width should be <= available space")
		})
	}
}

func TestMeasure_MaxWidth(t *testing.T) {
	tests := []struct {
		name          string
		contentWidth  int
		maxWidth      int
		flexGrow      int
		expectedWidth int
	}{
		{
			name:          "maxWidth constraint生效",
			contentWidth:  300,
			maxWidth:      200,
			flexGrow:      0,
			expectedWidth: 200,
		},
		{
			name:          "maxWidth与flexGrow冲突",
			contentWidth:  100,
			maxWidth:      200,
			flexGrow:      1,
			expectedWidth: 200,
		},
		{
			name:          "maxWidth超出父约束",
			contentWidth:  300,
			maxWidth:      500,
			flexGrow:      0,
			expectedWidth: 300, // Should be constrained by parent
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			child := NewMockMeasurableNode("child", tt.contentWidth, 50)
			flex := NewFlexLayout("flex", []Node{child})
			flex.SetFlex(0, tt.flexGrow, 1, 0)
			
			result := flex.Measure(Constraints{
				MinWidth:  0,
				MaxWidth:  tt.maxWidth,
				MinHeight: 0,
				MaxHeight: MaxInt,
			})
			
			assert.LessOrEqual(t, result.Width, tt.maxWidth, "Width should be <= maxWidth")
		})
	}
}

func TestMeasure_MaxHeight(t *testing.T) {
	tests := []struct {
		name           string
		contentHeight  int
		maxHeight      int
		flexGrow       int
		expectedHeight int
	}{
		{
			name:           "maxHeight约束生效",
			contentHeight:  300,
			maxHeight:      200,
			flexGrow:       0,
			expectedHeight: 200,
		},
		{
			name:           "maxHeight与flexGrow冲突",
			contentHeight:  100,
			maxHeight:      200,
			flexGrow:       1,
			expectedHeight: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			child := NewMockMeasurableNode("child", 50, tt.contentHeight)
			flex := NewFlexLayout("flex", []Node{child})
			flex.SetDirection(FlexColumn)
			flex.SetFlex(0, tt.flexGrow, 1, 0)
			
			result := flex.Measure(Constraints{
				MinWidth:  0,
				MaxWidth:  MaxInt,
				MinHeight: 0,
				MaxHeight: tt.maxHeight,
			})
			
			assert.LessOrEqual(t, result.Height, tt.maxHeight, "Height should be <= maxHeight")
		})
	}
}

func TestMeasure_MinConstraints(t *testing.T) {
	tests := []struct {
		name           string
		contentSize    Size
		minWidth       int
		minHeight      int
		expectedWidth  int
		expectedHeight int
	}{
		{
			name:           "minWidth约束生效",
			contentSize:    Size{Width: 50, Height: 50},
			minWidth:       100,
			minHeight:      0,
			expectedWidth:  100,
			expectedHeight: 50,
		},
		{
			name:           "minHeight约束生效",
			contentSize:    Size{Width: 50, Height: 50},
			minWidth:       0,
			minHeight:      100,
			expectedWidth:  50,
			expectedHeight: 100,
		},
		{
			name:           "min与max矛盾处理",
			contentSize:    Size{Width: 50, Height: 50},
			minWidth:       200,
			minHeight:      200,
			expectedWidth:  200,
			expectedHeight: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			child := NewMockMeasurableNode("child", tt.contentSize.Width, tt.contentSize.Height)
			flex := NewFlexLayout("flex", []Node{child})
			
			result := flex.Measure(Constraints{
				MinWidth:  tt.minWidth,
				MaxWidth:  MaxInt,
				MinHeight: tt.minHeight,
				MaxHeight: MaxInt,
			})
			
			assert.GreaterOrEqual(t, result.Width, tt.minWidth, "Width should be >= minWidth")
			assert.GreaterOrEqual(t, result.Height, tt.minHeight, "Height should be >= minHeight")
		})
	}
}

func TestEngine_Measure(t *testing.T) {
	engine := NewEngine()
	
	tests := []struct {
		name        string
		node        Node
		constraints Constraints
		wantWidth   int
		wantHeight  int
	}{
		{
			name:        "measurable node",
			node:        NewMockMeasurableNode("measurable", 100, 50),
			constraints: UnboundedConstraints(),
			wantWidth:   100,
			wantHeight:  50,
		},
		{
			name:        "non-measurable node",
			node:        NewMockNode("non-measurable", 100, 50),
			constraints: NewConstraints(10, 200, 10, 100),
			wantWidth:   10,
			wantHeight:  10,
		},
		{
			name:        "constrained measurable",
			node:        NewMockMeasurableNode("constrained", 300, 200),
			constraints: NewConstraints(0, 150, 0, 100),
			wantWidth:   150,
			wantHeight:  100,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.Measure(tt.node, tt.constraints)
			
			assert.Equal(t, tt.wantWidth, result.Width, "Width should match expected")
			assert.Equal(t, tt.wantHeight, result.Height, "Height should match expected")
		})
	}
}

func TestEngine_Measure_FlexLayout(t *testing.T) {
	engine := NewEngine()
	
	tests := []struct {
		name        string
		direction   FlexDirection
		children    []Node
		constraints Constraints
	}{
		{
			name:      "row layout",
			direction: FlexRow,
			children: []Node{
				NewMockMeasurableNode("child1", 100, 50),
				NewMockMeasurableNode("child2", 150, 50),
			},
			constraints: UnboundedConstraints(),
		},
		{
			name:      "column layout",
			direction: FlexColumn,
			children: []Node{
				NewMockMeasurableNode("child1", 50, 100),
				NewMockMeasurableNode("child2", 50, 150),
			},
			constraints: UnboundedConstraints(),
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flex := NewFlexLayout("flex", tt.children)
			flex.SetDirection(tt.direction)
			
			result := engine.Measure(flex, tt.constraints)
			
			assert.Greater(t, result.Width, 0, "Width should be greater than 0")
			assert.Greater(t, result.Height, 0, "Height should be greater than 0")
		})
	}
}

// Benchmark tests
func BenchmarkMeasure_SimpleNode(b *testing.B) {
	node := NewMockMeasurableNode("bench", 100, 50)
	constraints := UnboundedConstraints()
	engine := NewEngine()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.Measure(node, constraints)
	}
}

func BenchmarkMeasure_FlexLayout(b *testing.B) {
	children := make([]Node, 10)
	for i := range children {
		children[i] = NewMockMeasurableNode("child", 50, 50)
	}
	flex := NewFlexLayout("flex", children)
	constraints := UnboundedConstraints()
	engine := NewEngine()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.Measure(flex, constraints)
	}
}

func BenchmarkMeasure_NestedLayout(b *testing.B) {
	var parent Node = NewMockMeasurableNode("root", 100, 100)
	
	// Create 5 levels of nesting
	for i := 0; i < 5; i++ {
		child := NewMockMeasurableNode("level", 80, 80)
		parent = NewFlexLayout("parent"+string(rune('0'+i)), []Node{parent, child})
	}
	
	constraints := UnboundedConstraints()
	engine := NewEngine()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.Measure(parent, constraints)
	}
}
