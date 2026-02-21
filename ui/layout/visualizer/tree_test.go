package visualizer

import (
	"strings"
	"testing"

	"github.com/wwsheng009/mint/runtime/layout"
)

// =============================================================================
// Visualizer Core Tests
// =============================================================================

func TestVisualizer_New(t *testing.T) {
	vis := NewVisualizer()

	if vis == nil {
		t.Fatal("NewVisualizer() should not return nil")
	}
	if vis.nodes == nil {
		t.Fatal("nodes map should be initialized")
	}
	if vis.rootID != "" {
		t.Errorf("rootID should be empty, got '%s'", vis.rootID)
	}
}

func TestVisualizer_AddNode_Root(t *testing.T) {
	vis := NewVisualizer()

	vis.AddNode(
		"root",
		"panel",
		layout.Rect{X: 0, Y: 0, Width: 30, Height: 10},
		layout.Constraints{MinWidth: 0, MinHeight: 0, MaxWidth: 30, MaxHeight: 10},
		layout.Constraints{},
		layout.Size{Width: 30, Height: 10},
		"",
	)

	if vis.rootID != "root" {
		t.Errorf("rootID should be 'root', got '%s'", vis.rootID)
	}

	node := vis.GetNode("root")
	if node == nil {
		t.Fatal("root node should exist")
	}
	if node.ID != "root" {
		t.Errorf("expected ID 'root', got '%s'", node.ID)
	}
	if node.Tag != "panel" {
		t.Errorf("expected Tag 'panel', got '%s'", node.Tag)
	}
	if node.Bounds.Width != 30 {
		t.Errorf("expected width 30, got %d", node.Bounds.Width)
	}
	if node.Bounds.Height != 10 {
		t.Errorf("expected height 10, got %d", node.Bounds.Height)
	}
}

func TestVisualizer_AddNode_Child(t *testing.T) {
	vis := NewVisualizer()

	// Add root
	vis.AddNode(
		"root",
		"panel",
		layout.Rect{X: 0, Y: 0, Width: 30, Height: 10},
		layout.Constraints{},
		layout.Constraints{},
		layout.Size{Width: 30, Height: 10},
		"",
	)

	// Add child
	vis.AddNode(
		"child",
		"text",
		layout.Rect{X: 1, Y: 1, Width: 28, Height: 8},
		layout.Constraints{},
		layout.Constraints{},
		layout.Size{Width: 28, Height: 8},
		"root",
	)

	root := vis.GetNode("root")
	if len(root.Children) != 1 {
		t.Errorf("expected 1 child, got %d", len(root.Children))
	}
	if root.Children[0] != "child" {
		t.Errorf("expected child ID 'child', got '%s'", root.Children[0])
	}

	child := vis.GetNode("child")
	if child.ParentID != "root" {
		t.Errorf("expected ParentID 'root', got '%s'", child.ParentID)
	}
}

func TestVisualizer_SetNodeProperty(t *testing.T) {
	vis := NewVisualizer()

	vis.AddNode(
		"root",
		"panel",
		layout.Rect{},
		layout.Constraints{},
		layout.Constraints{},
		layout.Size{},
		"",
	)

	vis.SetNodeProperty("root", "title", "My Panel")
	vis.SetNodeProperty("root", "border", "single")

	node := vis.GetNode("root")
	if len(node.Props) != 2 {
		t.Errorf("expected 2 props, got %d", len(node.Props))
	}
	if node.Props["title"] != "My Panel" {
		t.Errorf("expected title 'My Panel', got '%v'", node.Props["title"])
	}
	if node.Props["border"] != "single" {
		t.Errorf("expected border 'single', got '%v'", node.Props["border"])
	}
}

// =============================================================================
// Print and Visualization Tests
// =============================================================================

func TestVisualizer_PrintTree_Simple(t *testing.T) {
	vis := NewVisualizer()

	vis.AddNode(
		"root",
		"panel",
		layout.Rect{X: 0, Y: 0, Width: 30, Height: 10},
		layout.Constraints{},
		layout.Constraints{},
		layout.Size{Width: 30, Height: 10},
		"",
	)

	output := vis.PrintTree()

	// Check that key information is present
	if !strings.Contains(output, "panel") {
		t.Error("PrintTree should contain 'panel'")
	}
	if !strings.Contains(output, "Position: (0, 0)") {
		t.Error("PrintTree should contain position")
	}
	if !strings.Contains(output, "Size: 30w x 10h") {
		t.Error("PrintTree should contain size")
	}
	if !strings.Contains(output, "Layout Tree:") {
		t.Error("PrintTree should contain header")
	}
}

func TestVisualizer_PrintTree_WithChildren(t *testing.T) {
	vis := NewVisualizer()

	vis.AddNode(
		"root",
		"panel",
		layout.Rect{X: 0, Y: 0, Width: 30, Height: 10},
		layout.Constraints{},
		layout.Constraints{},
		layout.Size{},
		"",
	)

	// Add two children
	vis.AddNode(
		"child1",
		"text",
		layout.Rect{X: 1, Y: 1, Width: 15, Height: 5},
		layout.Constraints{},
		layout.Constraints{},
		layout.Size{},
		"root",
	)

	vis.AddNode(
		"child2",
		"text",
		layout.Rect{X: 16, Y: 1, Width: 15, Height: 5},
		layout.Constraints{},
		layout.Constraints{},
		layout.Size{},
		"root",
	)

	output := vis.PrintTree()

	// Should show tree structure
	if !strings.Contains(output, "│") {
		t.Error("PrintTree should show children with branches")
	}
}

func TestVisualizer_PrintSummary(t *testing.T) {
	vis := NewVisualizer()

	// Build a small tree
	vis.AddNode(
		"root",
		"panel",
		layout.Rect{X: 0, Y: 0, Width: 30, Height: 10},
		layout.Constraints{},
		layout.Constraints{},
		layout.Size{},
		"",
	)

	vis.AddNode(
		"child1",
		"text",
		layout.Rect{},
		layout.Constraints{},
		layout.Constraints{},
		layout.Size{},
		"root",
	)

	vis.AddNode(
		"child2",
		"text",
		layout.Rect{},
		layout.Constraints{},
		layout.Constraints{},
		layout.Size{},
		"root",
	)

	summary := vis.PrintSummary()

	// Check summary contains key stats
	if !strings.Contains(summary, "Total Nodes: 3") {
		t.Error("Summary should show 3 nodes")
	}
	if !strings.Contains(summary, "Max Depth: 1") {
		t.Error("Summary should show depth 1")
	}
	if !strings.Contains(summary, "Root Size:") {
		t.Error("Summary should show root size")
	}
	if !strings.Contains(summary, "Node Types:") {
		t.Error("Summary should show node types")
	}
	if !strings.Contains(summary, "panel: 1") {
		t.Error("Summary should count panel nodes")
	}
	if !strings.Contains(summary, "text: 2") {
		t.Error("Summary should count text nodes")
	}
}

func TestVisualizer_FindProblems(t *testing.T) {
	vis := NewVisualizer()

	// Add a node that exceeds constraints
	vis.AddNode(
		"root",
		"panel",
		layout.Rect{},
		layout.Constraints{
			MinWidth:  0,
			MaxWidth:  20,
			MinHeight: 0,
			MaxHeight: 5,
		},
		layout.Constraints{},
		layout.Size{Width: 30, Height: 10}, // Exceeds constraints
		"",
	)

	problems := vis.FindProblems()

	if len(problems) != 2 {
		t.Errorf("expected 2 problems (width and height), got %d", len(problems))
	}

	// Check width problem
	widthProblemFound := false
	heightProblemFound := false
	for _, p := range problems {
		if strings.Contains(p, "width 30 exceeds MaxWidth 20") {
			widthProblemFound = true
		}
		if strings.Contains(p, "height 10 exceeds MaxHeight 5") {
			heightProblemFound = true
		}
	}

	if !widthProblemFound {
		t.Error("should report width exceeding MaxWidth")
	}
	if !heightProblemFound {
		t.Error("should report height exceeding MaxHeight")
	}
}

func TestVisualizer_FindProblems_BelowMin(t *testing.T) {
	vis := NewVisualizer()

	// Add a node below minimum
	vis.AddNode(
		"root",
		"panel",
		layout.Rect{},
		layout.Constraints{
			MinWidth:  20,
			MaxWidth:  50,
			MinHeight: 10,
			MaxHeight: 20,
		},
		layout.Constraints{},
		layout.Size{Width: 10, Height: 5}, // Below minimum
		"",
	)

	problems := vis.FindProblems()

	if len(problems) != 2 {
		t.Errorf("expected 2 problems, got %d", len(problems))
	}
}

func TestVisualizer_Clear(t *testing.T) {
	vis := NewVisualizer()

	// Add nodes
	vis.AddNode(
		"root",
		"panel",
		layout.Rect{},
		layout.Constraints{},
		layout.Constraints{},
		layout.Size{},
		"",
	)

	vis.AddNode(
		"child",
		"text",
		layout.Rect{},
		layout.Constraints{},
		layout.Constraints{},
		layout.Size{},
		"root",
	)

	// Clear
	vis.Clear()

	// Check state
	if len(vis.nodes) != 0 {
		t.Errorf("expected 0 nodes after clear, got %d", len(vis.nodes))
	}
	if vis.rootID != "" {
		t.Errorf("rootID should be empty after clear, got '%s'", vis.rootID)
	}
}

// =============================================================================
// VisualizerBuilder Tests
// =============================================================================

func TestVisualizeVNode_Simple(t *testing.T) {
	// This test is minimal since we need actual VNode implementations
	// Full integration tests would exist in the component test files

	vis := VisualizeVNode(nil, layout.Constraints{})

	if vis == nil {
		t.Fatal("VisualizeVNode should not return nil")
	}
}

// =============================================================================
// Helper Tests
// =============================================================================

func TestFormatConstraints(t *testing.T) {
	tests := []struct {
		name     string
		c        layout.Constraints
		contains []string
	}{
		{
			name: "Unbounded",
			c: layout.Constraints{
				MinWidth:  0,
				MinHeight: 0,
				MaxWidth:  layout.MaxInt,
				MaxHeight: layout.MaxInt,
			},
			contains: []string{"Unbounded"},
		},
		{
			name: "Bounded",
			c: layout.Constraints{
				MinWidth:  10,
				MinHeight: 5,
				MaxWidth:  30,
				MaxHeight: 20,
			},
			contains: []string{"{10..30", "{5..20"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := formatConstraints(tt.c)
			for _, substr := range tt.contains {
				if !strings.Contains(output, substr) {
					t.Errorf("formatConstraints output should contain '%s', got: %s", substr, output)
				}
			}
		})
	}
}

func TestShortID(t *testing.T) {
	tests := []struct {
		id     string
		expect string
	}{
		{"short", "short"},
		{"12345678", "12345678"},
		{"123456789", "...23456789"},
		{"very_long_id_here_123456789", "...23456789"},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			result := shortID(tt.id)
			if result != tt.expect {
				t.Errorf("shortID(%s) = %s, expected %s", tt.id, result, tt.expect)
			}
		})
	}
}



