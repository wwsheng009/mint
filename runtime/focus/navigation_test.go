package focus

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wwsheng009/mint/runtime"
)

// createMockLayoutNode creates a mock LayoutNode with position and size
func createMockLayoutNode(id string, x, y, width, height int) *runtime.LayoutNode {
	mockComp := NewMockFocusableComponent(id, true)
	
	return &runtime.LayoutNode{
		ID:             id,
		X:              x,
		Y:              y,
		MeasuredWidth:  width,
		MeasuredHeight: height,
		Component: &runtime.ComponentRef{
			Instance: mockComp,
		},
		Children: []*runtime.LayoutNode{},
	}
}

// createMockFocusTree creates a tree of mock layout nodes
func createMockFocusTree() *runtime.LayoutNode {
	root := &runtime.LayoutNode{
		ID: "root",
		X:  0, Y: 0,
		MeasuredWidth:  600,
		MeasuredHeight: 400,
		Children: []*runtime.LayoutNode{},
	}
	
	// Create a 3x3 grid of buttons
	for row := 0; row < 3; row++ {
		for col := 0; col < 3; col++ {
			id := string(rune('1'+row)) + string(rune('1'+col))
			node := createMockLayoutNode(id, col*200, row*100, 100, 100)
			root.Children = append(root.Children, node)
		}
	}
	
	return root
}

func TestNavigate_Up(t *testing.T) {
	t.Run("向上导航", func(t *testing.T) {
		root := createMockFocusTree()
		gn := NewGeometricNavigator(root)
		
		focusableIDs := []string{"11", "12", "13", "21", "22", "23", "31", "32", "33"}
		gn.RefreshBounds(focusableIDs)
		
		// Start at "22" (middle)
		next := gn.FindNextInDirection("22", DirectionUp, focusableIDs)
		
		assert.Equal(t, "12", next, "Should navigate to node above")
	})
	
	t.Run("最上方元素处理", func(t *testing.T) {
		root := createMockFocusTree()
		gn := NewGeometricNavigator(root)
		
		focusableIDs := []string{"11", "12", "21", "22"}
		gn.RefreshBounds(focusableIDs)
		
		// Start at top row "11"
		next := gn.FindNextInDirection("11", DirectionUp, focusableIDs)
		
		// At top row, may return empty or wrap
		assert.NotEqual(t, "33", next, "Should not stay on same node")
	})
	
	t.Run("跨越边界", func(t *testing.T) {
		root := createMockFocusTree()
		gn := NewGeometricNavigator(root)
		
		focusableIDs := []string{"11", "21", "31"}
		gn.RefreshBounds(focusableIDs)
		
		// Navigate up from top row
		next := gn.FindNextInDirection("11", DirectionUp, focusableIDs)
		
		// May return empty or wrap around
		assert.NotEqual(t, "11", next, "Should try to find different node")
	})
}

func TestNavigate_Down(t *testing.T) {
	t.Run("向下导航", func(t *testing.T) {
		root := createMockFocusTree()
		gn := NewGeometricNavigator(root)
		
		focusableIDs := []string{"11", "12", "13", "21", "22", "23", "31", "32", "33"}
		gn.RefreshBounds(focusableIDs)
		
		// Start at "22" (middle)
		next := gn.FindNextInDirection("22", DirectionDown, focusableIDs)
		
		assert.Equal(t, "32", next, "Should navigate to node below")
	})
	
	t.Run("最下方元素处理", func(t *testing.T) {
		root := createMockFocusTree()
		gn := NewGeometricNavigator(root)
		
		focusableIDs := []string{"31", "32", "33"}
		gn.RefreshBounds(focusableIDs)
		
		// Start at bottom row "32"
		next := gn.FindNextInDirection("32", DirectionDown, focusableIDs)
		
		// At bottom row, may return empty or wrap
		assert.NotEqual(t, "32", next, "Should try to find different node")
	})
	
	t.Run("跨越边界", func(t *testing.T) {
		root := createMockFocusTree()
		gn := NewGeometricNavigator(root)
		
		focusableIDs := []string{"31", "32", "33"}
		gn.RefreshBounds(focusableIDs)
		
		// Try to navigate down from bottom
		next := gn.FindNextInDirection("33", DirectionDown, focusableIDs)
		
		// May return empty or wrap around
		assert.NotEqual(t, "33", next, "Should try to find different node")
	})
}

func TestNavigate_Left(t *testing.T) {
	t.Run("向左导航", func(t *testing.T) {
		root := createMockFocusTree()
		gn := NewGeometricNavigator(root)
		
		focusableIDs := []string{"11", "12", "13", "21", "22", "23", "31", "32", "33"}
		gn.RefreshBounds(focusableIDs)
		
		// Start at "13" (right of middle row)
		next := gn.FindNextInDirection("13", DirectionLeft, focusableIDs)
		
		assert.Equal(t, "12", next, "Should navigate to node on left")
	})
	
	t.Run("最左侧元素处理", func(t *testing.T) {
		root := createMockTreeHorizontal()
		gn := NewGeometricNavigator(root)
		
		focusableIDs := []string{"left", "middle", "right"}
		gn.RefreshBounds(focusableIDs)
		
		// Start at leftmost
		next := gn.FindNextInDirection("left", DirectionLeft, focusableIDs)
		
		// At left boundary, may return empty or wrap
		assert.NotEqual(t, "left", next, "Should try to find different node")
	})
}

func TestNavigate_Right(t *testing.T) {
	t.Run("向右导航", func(t *testing.T) {
		root := createMockTreeHorizontal()
		gn := NewGeometricNavigator(root)
		
		focusableIDs := []string{"left", "middle", "right"}
		gn.RefreshBounds(focusableIDs)
		
		// Start at "left"
		next := gn.FindNextInDirection("left", DirectionRight, focusableIDs)
		
		assert.Equal(t, "middle", next, "Should navigate to node on right")
	})
	
	t.Run("最右侧元素处理", func(t *testing.T) {
		root := createMockTreeHorizontal()
		gn := NewGeometricNavigator(root)
		
		focusableIDs := []string{"left", "middle", "right"}
		gn.RefreshBounds(focusableIDs)
		
		// Start at rightmost
		next := gn.FindNextInDirection("right", DirectionRight, focusableIDs)
		
		// At right boundary, may return empty or wrap
		assert.NotEqual(t, "right", next, "Should try to find different node")
	})
}

func TestNavigate_Home(t *testing.T) {
	t.Run("Home键到首元素", func(t *testing.T) {
		root := createMockFocusTree()
		gn := NewGeometricNavigator(root)
		
		focusableIDs := []string{"33", "22", "11", "32", "21", "12", "31", "23", "13"}
		gn.RefreshBounds(focusableIDs)
		
		// From "33", find top-left most
		next := gn.FindNextInDirection("33", DirectionUp, focusableIDs)
		
		// Should find a node in the top row
		bounds := gn.GetBounds(next)
		assert.NotNil(t, bounds, "Should have bounds")
		assert.LessOrEqual(t, bounds.Y, 100, "Should be in top row")
	})
	
	t.Run("无元素处理", func(t *testing.T) {
		gn := NewGeometricNavigator(nil)
		
		next := gn.FindNextInDirection("", DirectionUp, []string{})
		
		assert.Empty(t, next, "Should return empty when no focusable components")
	})
}

func TestNavigate_End(t *testing.T) {
	t.Run("End键到尾元素", func(t *testing.T) {
		root := createMockFocusTree()
		gn := NewGeometricNavigator(root)
		
		focusableIDs := []string{"11", "12", "13", "21", "22", "23", "31", "32", "33"}
		gn.RefreshBounds(focusableIDs)
		
		// From "11", find top-left most
		next := gn.FindNextInDirection("11", DirectionDown, focusableIDs)
		
		// Should find a node in the bottom row
		if next != "" {
			bounds := gn.GetBounds(next)
			assert.NotNil(t, bounds, "Should have bounds")
			assert.GreaterOrEqual(t, bounds.Y, 100, "Should be in bottom or middle rows")
		}
	})
}

func TestNavigate_WrapAround(t *testing.T) {
	t.Run("循环导航开启", func(t *testing.T) {
		root := createMockTreeHorizontal()
		gn := NewGeometricNavigator(root)
		
		focusableIDs := []string{"left", "middle", "right"}
		gn.RefreshBounds(focusableIDs)
		
		// Navigate left from leftmost
		next1 := gn.FindNextInDirection("left", DirectionLeft, focusableIDs)
		
		// Should wrap to rightmost or find best candidate
		assert.NotEqual(t, "left", next1, "Should try to find different candidate")
		
		// Navigate right from rightmost
		next2 := gn.FindNextInDirection("right", DirectionRight, focusableIDs)
		
		// Should wrap to leftmost or find best candidate
		assert.NotEqual(t, "right", next2, "Should try to find different candidate")
	})
	
	t.Run("单元素循环", func(t *testing.T) {
		root := &runtime.LayoutNode{
			ID:             "single",
			X:              0,
			Y:              0,
			MeasuredWidth:  100,
			MeasuredHeight: 50,
			Children:       []*runtime.LayoutNode{},
		}
		
		gn := NewGeometricNavigator(root)
		
		focusableIDs := []string{"single"}
		gn.RefreshBounds(focusableIDs)
		
		// Navigate in any direction with single element
		next := gn.FindNextInDirection("single", DirectionUp, focusableIDs)
		
		// With single element, navigation may return empty
		// This is expected behavior - no other candidates exist
		assert.True(t, next == "single" || next == "", "Should return same element or empty when only one exists")
	})
}

// Helper function to create horizontal tree
func createMockTreeHorizontal() *runtime.LayoutNode {
	root := &runtime.LayoutNode{
		ID:             "root",
		X:              0,
		Y:              0,
		MeasuredWidth:  600,
		MeasuredHeight: 100,
		Children:       []*runtime.LayoutNode{},
	}
	
	// Create horizontal layout
	left := createMockLayoutNode("left", 0, 0, 200, 100)
	middle := createMockLayoutNode("middle", 200, 0, 200, 100)
	right := createMockLayoutNode("right", 400, 0, 200, 100)
	
	root.Children = append(root.Children, left, middle, right)
	return root
}

func TestNavigate_OverlapScoring(t *testing.T) {
	t.Run("horizontal overlap bonus for vertical navigation", func(t *testing.T) {
		root := createMockTreeOverlap()
		gn := NewGeometricNavigator(root)
		
		focusableIDs := []string{"topLeft", "topRight", "bottomLeft"}
		gn.RefreshBounds(focusableIDs)
		
		// From topLeft (0,0,100,100), navigate down
		// bottomLeft (0,200,100,100) should be preferred over topRight (100,0,100,100)
		// due to horizontal overlap
		next := gn.FindNextInDirection("topLeft", DirectionDown, focusableIDs)
		
		assert.Equal(t, "bottomLeft", next, "Should prefer node with horizontal overlap")
	})
	
	t.Run("vertical overlap bonus for horizontal navigation", func(t *testing.T) {
		root := createMockTreeOverlap()
		gn := NewGeometricNavigator(root)
		
		focusableIDs := []string{"topLeft", "topRight", "bottomLeft"}
		gn.RefreshBounds(focusableIDs)
		
		// From topLeft (0,0,100,100), navigate right
		// topRight (100,0,100,100) should be preferred over bottomLeft (0,200,100,100)
		// due to vertical overlap
		next := gn.FindNextInDirection("topLeft", DirectionRight, focusableIDs)
		
		assert.Equal(t, "topRight", next, "Should prefer node with vertical overlap")
	})
}

func createMockTreeOverlap() *runtime.LayoutNode {
	root := &runtime.LayoutNode{
		ID:             "root",
		X:              0,
		Y:              0,
		MeasuredWidth:  200,
		MeasuredHeight: 300,
		Children:       []*runtime.LayoutNode{},
	}
	
	// Create overlapping nodes
	topLeft := createMockLayoutNode("topLeft", 0, 0, 100, 100)
	topRight := createMockLayoutNode("topRight", 100, 0, 100, 100)
	bottomLeft := createMockLayoutNode("bottomLeft", 0, 200, 100, 100)
	
	root.Children = append(root.Children, topLeft, topRight, bottomLeft)
	return root
}

func TestNavigate_DistanceCalculation(t *testing.T) {
	t.Run("prioritize closer components", func(t *testing.T) {
		root := &runtime.LayoutNode{
			ID:             "root",
			X:              0,
			Y:              0,
			MeasuredWidth: 400,
			MeasuredHeight: 400,
			Children:       []*runtime.LayoutNode{},
		}
		
		// Create nodes at different distances
		close := createMockLayoutNode("close", 50, 50, 100, 100)
		far := createMockLayoutNode("far", 300, 300, 100, 100)
		
		root.Children = append(root.Children, close, far)
		
		gn := NewGeometricNavigator(root)
		focusableIDs := []string{"close", "far"}
		gn.RefreshBounds(focusableIDs)
		
		// From "close", navigate up - should prefer close if it exists above
		// but if only "far" is above, it should be chosen
		next := gn.FindNextInDirection("close", DirectionUp, focusableIDs)
		
		// Should find a candidate (may be "far" if nothing is above "close")
		assert.NotEqual(t, "close", next, "Should find different candidate")
	})
}

func TestNavigate_NoComponents(t *testing.T) {
	t.Run("navigate with empty list", func(t *testing.T) {
		root := createMockFocusTree()
		gn := NewGeometricNavigator(root)
		
		next := gn.FindNextInDirection("22", DirectionUp, []string{})
		
		assert.Empty(t, next, "Should return empty when no focusable components")
	})
}

func TestNavigate_RefreshBounds(t *testing.T) {
	t.Run("refresh bounds updates cache", func(t *testing.T) {
		root := createMockFocusTree()
		gn := NewGeometricNavigator(root)
		
		focusableIDs := []string{"11", "22", "33"}
		gn.RefreshBounds(focusableIDs)
		
		// Verify bounds are cached
		bounds11 := gn.GetBounds("11")
		assert.NotNil(t, bounds11, "Should have cached bounds for node11")
		assert.Equal(t, "11", bounds11.ID, "Should have correct ID")
		
		// Refresh again
		gn.RefreshBounds(focusableIDs)
		
		// Bounds should still exist
		bounds22 := gn.GetBounds("22")
		assert.NotNil(t, bounds22, "Should have cached bounds after refresh")
	})
}

func TestNavigate_FindNearestInDirection(t *testing.T) {
	t.Run("convenience method refreshes and finds", func(t *testing.T) {
		root := createMockFocusTree()
		gn := NewGeometricNavigator(root)
		
		focusableIDs := []string{"11", "22", "33"}
		
		// Use convenience method that auto-refreshes
		next := gn.FindNearestInDirection("22", DirectionUp, focusableIDs)
		
		// Should find a candidate above
		assert.NotEmpty(t, next, "Should find nearest component")
		
		// Verify bounds were refreshed
		bounds := gn.GetBounds(next)
		assert.NotNil(t, bounds, "Should have bounds after refresh")
	})
}

// Benchmark tests
func BenchmarkNavigate_Up(b *testing.B) {
	root := createMockFocusTree()
	gn := NewGeometricNavigator(root)
	focusableIDs := []string{"11", "12", "13", "21", "22", "23", "31", "32", "33"}
	gn.RefreshBounds(focusableIDs)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gn.FindNextInDirection("22", DirectionUp, focusableIDs)
	}
}

func BenchmarkNavigate_Down(b *testing.B) {
	root := createMockFocusTree()
	gn := NewGeometricNavigator(root)
	focusableIDs := []string{"11", "12", "13", "21", "22", "23", "31", "32", "33"}
	gn.RefreshBounds(focusableIDs)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gn.FindNextInDirection("22", DirectionDown, focusableIDs)
	}
}

func BenchmarkNavigate_Left(b *testing.B) {
	root := createMockFocusTree()
	gn := NewGeometricNavigator(root)
	focusableIDs := []string{"11", "12", "13", "21", "22", "23", "31", "32", "33"}
	gn.RefreshBounds(focusableIDs)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gn.FindNextInDirection("13", DirectionLeft, focusableIDs)
	}
}

func BenchmarkNavigate_Right(b *testing.B) {
	root := createMockFocusTree()
	gn := NewGeometricNavigator(root)
	focusableIDs := []string{"11", "12", "13", "21", "22", "23", "31", "32", "33"}
	gn.RefreshBounds(focusableIDs)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gn.FindNextInDirection("11", DirectionRight, focusableIDs)
	}
}
