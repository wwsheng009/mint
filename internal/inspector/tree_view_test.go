package inspector

import (
	"strings"
	"testing"

	"github.com/wwsheng009/mint/ui"
)

// TestNewTreeView tests creating a new tree view
func TestNewTreeView(t *testing.T) {
	treeView := NewTreeView()

	if treeView == nil {
		t.Fatal("Expected non-nil TreeView")
	}

	if treeView.maxDepth != 100 {
		t.Errorf("Expected maxDepth 100, got %d", treeView.maxDepth)
	}

	if treeView.maxNodes != 1000 {
		t.Errorf("Expected maxNodes 1000, got %d", treeView.maxNodes)
	}

	if !treeView.showIcons {
		t.Error("Show icons should be enabled by default")
	}
}

// TestSetRoot tests setting the root VNode
func TestSetRoot(t *testing.T) {
	treeView := NewTreeView()

	// Create a simple tree
	button := ui.NewButtonBuilder("Test").Build()
	text := ui.Text("Hello")

	container := ui.HStack(button, text)

	err := treeView.SetRoot(container)
	if err != nil {
		t.Fatalf("Failed to set root: %v", err)
	}

	if treeView.root == nil {
		t.Error("Root should be set")
	}
}

// TestSetRoot_Nil tests setting nil root
func TestSetRoot_Nil(t *testing.T) {
	treeView := NewTreeView()

	err := treeView.SetRoot(nil)
	if err != nil {
		t.Errorf("Setting nil root should not error, got %v", err)
	}

	if treeView.root != nil {
		t.Error("Root should be nil after setting nil")
	}
}

// TestFormatTree tests tree formatting
func TestFormatTree(t *testing.T) {
	treeView := NewTreeView()

	// Create a simple tree
	button1 := ui.NewButtonBuilder("Button1").Build()
	button2 := ui.NewButtonBuilder("Button2").Build()
	text := ui.Text("Hello")

	container := ui.HStack(
		ui.VStack(button1, button2),
		text,
	)

	err := treeView.SetRoot(container)
	if err != nil {
		t.Fatalf("Failed to set root: %v", err)
	}

	output := treeView.FormatTree()

	if output == "" {
		t.Error("Expected non-empty tree output")
	}

	// Check for tree structure indicators
	requiredStrings := []string{
		"Layout Tree",
		"├──",
		"└──",
		"LayoutNode",
		"ButtonVNode",
	}

	for _, s := range requiredStrings {
		if !contains(output, s) {
			t.Errorf("Output should contain '%s'", s)
		}
	}
}

// TestFormatTree_Empty tests formatting empty tree
func TestFormatTree_Empty(t *testing.T) {
	treeView := NewTreeView()

	output := treeView.FormatTree()

	if output != "No tree to display" {
		t.Errorf("Expected 'No tree to display', got '%s'", output)
	}
}

// TestToggleNode tests toggling node expansion
func TestToggleNode(t *testing.T) {
	treeView := NewTreeView()

	button := ui.NewButtonBuilder("Test").Build()
	treeView.SetRoot(button)

	// Get initial UniqueID
	uniqueID := treeView.root.UniqueID

	// Initially should be expanded by default? Let's check
	initialState := treeView.root.Expanded

	// Toggle
	treeView.ToggleNode(uniqueID)

	if treeView.root.Expanded == initialState {
		t.Error("Expansion state should toggle")
	}

	// Toggle back
	treeView.ToggleNode(uniqueID)

	if treeView.root.Expanded != initialState {
		t.Error("Expansion state should toggle back")
	}
}

// TestExpandAll tests expanding all nodes
func TestExpandAll(t *testing.T) {
	treeView := NewTreeView()

	button1 := ui.NewButtonBuilder("B1").Build()
	button2 := ui.NewButtonBuilder("B2").Build()
	container := ui.HStack(button1, button2)

	treeView.SetRoot(container)

	// Collapse all first
	treeView.CollapseAll()

	// Then expand all
	treeView.ExpandAll()

	// Check that all nodes are expanded
	flatList := treeView.GetFlatList()
	for _, node := range flatList {
		if !node.Expanded {
			t.Errorf("Node %s should be expanded", node.Path)
		}
	}
}

// TestCollapseAll tests collapsing all nodes
func TestCollapseAll(t *testing.T) {
	treeView := NewTreeView()

	button1 := ui.NewButtonBuilder("B1").Build()
	button2 := ui.NewButtonBuilder("B2").Build()
	container := ui.HStack(button1, button2)

	treeView.SetRoot(container)

	// Collapse all
	treeView.CollapseAll()

	// Check that all nodes are collapsed
	flatList := treeView.GetFlatList()
	for _, node := range flatList {
		if node.Expanded {
			t.Errorf("Node %s should be collapsed", node.Path)
		}
	}
}

// TestFindNodeByPath tests finding nodes by path
func TestFindNodeByPath(t *testing.T) {
	treeView := NewTreeView()

	button := ui.NewButtonBuilder("Test").Build()
	treeView.SetRoot(button)

	if treeView.root == nil {
		t.Fatal("Root should be set")
	}

	path := treeView.root.Path
	found := treeView.FindNodeByPath(path)

	if found == nil {
		t.Errorf("Should find node with path '%s'", path)
	}

	if found != treeView.root {
		t.Error("Should find root node")
	}
}

// TestFindNodeByPath_NotFound tests finding non-existent path
func TestFindNodeByPath_NotFound(t *testing.T) {
	treeView := NewTreeView()

	button := ui.NewButtonBuilder("Test").Build()
	treeView.SetRoot(button)

	found := treeView.FindNodeByPath("nonexistent.path")

	if found != nil {
		t.Error("Should not find node with non-existent path")
	}
}

// TestFindNodesByType tests finding nodes by type
func TestFindNodesByType(t *testing.T) {
	treeView := NewTreeView()

	button1 := ui.NewButtonBuilder("Button1").Build()
	button2 := ui.NewButtonBuilder("Button2").Build()
	text := ui.Text("Hello")

	container := ui.HStack(
		ui.VStack(button1, button2),
		text,
	)

	treeView.SetRoot(container)

	// Find all buttons
	buttons := treeView.FindNodesByType("Button")

	if len(buttons) < 2 {
		t.Errorf("Expected at least 2 buttons, got %d", len(buttons))
	}

	// Find LayoutNode
	layoutNodes := treeView.FindNodesByType("LayoutNode")

	if len(layoutNodes) < 2 {
		t.Errorf("Expected at least 2 layout nodes, got %d", len(layoutNodes))
	}
}

// TestFindNodesByLabel tests finding nodes by label
func TestFindNodesByLabel(t *testing.T) {
	treeView := NewTreeView()

	button1 := ui.NewButtonBuilder("Click Me").Build()
	button2 := ui.NewButtonBuilder("Click Me Too").Build()
	text := ui.Text("Search Text")

	container := ui.VStack(button1, button2, text)

	treeView.SetRoot(container)

	// Search for "Click Me"
	results := treeView.FindNodesByLabel("Click Me")

	if len(results) < 2 {
		t.Errorf("Expected at least 2 matching nodes, got %d", len(results))
	}

	// Search for "Text" (case-insensitive)
	textResults := treeView.FindNodesByLabel("text")

	if len(textResults) < 1 {
		t.Errorf("Expected at least 1 text node, got %d", len(textResults))
	}
}

// TestGetTreeStats tests getting tree statistics
func TestGetTreeStats(t *testing.T) {
	treeView := NewTreeView()

	// Create a simple tree
	button1 := ui.NewButtonBuilder("B1").Build()
	button2 := ui.Box().Child(ui.Text("B2")).Build()
	button3 := ui.NewButtonBuilder("B3").Build()
	text1 := ui.Text("T1")
	text2 := ui.Text("T2")

	container := ui.HStack(
		ui.VStack(
			ui.HStack(button1, button2),
			ui.HStack(button3, text1),
		),
		text2,
	)

	treeView.SetRoot(container)

	stats := treeView.GetTreeStats()

	if stats.TotalNodes == 0 {
		t.Error("Should have at least one node")
	}

	if stats.ParentNodes == 0 {
		t.Error("Should have at least one parent node")
	}

	if stats.LeafNodes == 0 {
		t.Error("Should have at least one leaf node")
	}

	if stats.MaxDepth < 1 {
		t.Errorf("Max depth should be at least 1, got %d", stats.MaxDepth)
	}
}

// TestGetFlatList tests getting flat list of nodes
func TestGetFlatList(t *testing.T) {
	treeView := NewTreeView()

	button := ui.NewButtonBuilder("Test").Build()
	treeView.SetRoot(button)

	flatList := treeView.GetFlatList()

	if len(flatList) == 0 {
		t.Error("Flat list should not be empty")
	}

	// Root should be first
	if flatList[0] != treeView.root {
		t.Error("First element should be root")
	}
}

// TestGetFlatList_Empty tests flat list with no root
func TestGetFlatList_Empty(t *testing.T) {
	treeView := NewTreeView()

	flatList := treeView.GetFlatList()

	if len(flatList) != 0 {
		t.Error("Flat list should be empty when no root set")
	}
}

// TestSetShowIcons tests show icons control
func TestSetShowIcons(t *testing.T) {
	treeView := NewTreeView()

	treeView.SetShowIcons(false)
	if treeView.showIcons {
		t.Error("Show icons should be disabled")
	}

	treeView.SetShowIcons(true)
	if !treeView.showIcons {
		t.Error("Show icons should be enabled")
	}
}

// TestSetShowPaths_TreeView tests show paths control
func TestSetShowPaths_TreeView(t *testing.T) {
	treeView := NewTreeView()

	treeView.SetShowPaths(true)
	if !treeView.showPaths {
		t.Error("Show paths should be enabled")
	}

	treeView.SetShowPaths(false)
	if treeView.showPaths {
		t.Error("Show paths should be disabled")
	}
}

// TestSetCompact tests compact mode
func TestSetCompact(t *testing.T) {
	treeView := NewTreeView()

	treeView.SetCompact(true)
	if !treeView.compact {
		t.Error("Compact mode should be enabled")
	}
}

// TestSetMaxDepth tests max depth setting
func TestSetMaxDepth(t *testing.T) {
	treeView := NewTreeView()

	treeView.SetMaxDepth(5)
	if treeView.maxDepth != 5 {
		t.Errorf("Expected maxDepth 5, got %d", treeView.maxDepth)
	}
}

// TestSetMaxNodes tests max nodes setting
func TestSetMaxNodes(t *testing.T) {
	treeView := NewTreeView()

	treeView.SetMaxNodes(100)
	if treeView.maxNodes != 100 {
		t.Errorf("Expected maxNodes 100, got %d", treeView.maxNodes)
	}
}

// TestGetIconForType tests icon generation
func TestGetIconForType(t *testing.T) {
	tests := []struct {
		typeName string
		expected string
	}{
		{"ButtonVNode", "🔵"},
		{"TextVNode", "📝"},
		{"HStack", "→"},
		{"VStack", "↓"},
		{"BoxVNode", "📦"},
		{"unknown", "📦"},
	}

	for _, tt := range tests {
		t.Run(tt.typeName, func(t *testing.T) {
			result := getIconForType(tt.typeName)
			if result != tt.expected {
				t.Errorf("Expected '%s' for type %s, got '%s'",
					tt.expected, tt.typeName, result)
			}
		})
	}
}

// TestBuildTree_Structure tests tree building structure
func TestBuildTree_Structure(t *testing.T) {
	treeView := NewTreeView()

	// Create a known structure:
	// HStack
	//   ├── VStack
	//   │   ├── Button1
	//   │   └── Button2
	//   └── Text

	button1 := ui.NewButtonBuilder("B1").Build()
	button2 := ui.NewButtonBuilder("B2").Build()
	vstack := ui.VStack(button1, button2)
	text := ui.Text("Hello")
	root := ui.HStack(vstack, text)

	treeView.SetRoot(root)

	// Verify root
	if treeView.root == nil {
		t.Fatal("Root should be set")
	}

	// Verify root has 2 children (vstack and text)
	if len(treeView.root.Children) != 2 {
		t.Fatalf("Expected root to have 2 children, got %d", len(treeView.root.Children))
	}

	// Verify first child is LayoutNode (VStack wrapper)
	vstackNode := treeView.root.Children[0]
	if !contains(vstackNode.Info.Type, "LayoutNode") {
		t.Errorf("Expected first child to be LayoutNode, got %s", vstackNode.Info.Type)
	}

	// Verify vstack has 2 children
	if len(vstackNode.Children) != 2 {
		t.Fatalf("Expected vstack to have 2 children, got %d", len(vstackNode.Children))
	}

	// Verify vstack contains Button children
	hasButton := false
	for _, child := range vstackNode.Children {
		if contains(child.Info.Type, "Button") {
			hasButton = true
			break
		}
	}
	if !hasButton {
		t.Error("Expected vstack to contain Button children")
	}

	// Verify second child is ElementVNode (Text wrapper)
	textNode := treeView.root.Children[1]
	if !contains(textNode.Info.Type, "ElementVNode") {
		t.Errorf("Expected second child to contain ElementVNode, got %s", textNode.Info.Type)
	}
}

// TestPathGeneration tests that paths are correctly generated
func TestPathGeneration(t *testing.T) {
	treeView := NewTreeView()

	button := ui.NewButtonBuilder("Test").Build()
	treeView.SetRoot(button)

	if treeView.root.Path == "" {
		t.Error("Root should have a path")
	}

	// Path should be simple type name
	if !contains(strings.ToLower(treeView.root.Path), "button") {
		t.Errorf("Expected path to contain 'button', got '%s'", treeView.root.Path)
	}
}
