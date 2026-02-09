package inspector

import (
	"strings"
	"testing"

	"github.com/wwsheng009/mint/ui"
)

// TestTreeViewPathIndexBasedIDs tests that path+index IDs are unique and stable
func TestTreeViewPathIndexBasedIDs(t *testing.T) {
	tv := NewTreeView()
	tv.maxDepth = 5

	// Create a tree with multiple siblings of same type
	// This tests that index-based IDs prevent collisions
	root := ui.VStack(
		ui.Text("A"),
		ui.Text("B"),
		ui.Text("C"),
	)

	err := tv.SetRoot(root)
	if err != nil {
		t.Fatalf("SetRoot failed: %v", err)
	}

	// Get all nodes
	allNodes := tv.GetFlatList()

	// Collect all UniqueIDs
	uniqueIDs := make(map[string]bool)
	var textNodeUIDs []string

	for _, node := range allNodes {
		// Look for text nodes (check path contains "text" since type might be ElementVNode)
		if strings.Contains(strings.ToLower(node.Path), "text") {
			textNodeUIDs = append(textNodeUIDs, node.UniqueID)
		}

		// Check for duplicates
		if uniqueIDs[node.UniqueID] {
			t.Errorf("❌ DUPLICATE UniqueID found: %s", node.UniqueID)
		}
		uniqueIDs[node.UniqueID] = true
	}

	// Verify we have 3 Text nodes with different IDs
	if len(textNodeUIDs) != 3 {
		t.Fatalf("Expected 3 Text nodes, got %d", len(textNodeUIDs))
	}

	// Verify all three Text nodes have different UniqueIDs
	uidSet := make(map[string]bool)
	for _, uid := range textNodeUIDs {
		if uidSet[uid] {
			t.Errorf("❌ Duplicate Text node UniqueID: %s", uid)
		}
		uidSet[uid] = true
		t.Logf("✓ Text node UniqueID: %s", uid)
	}

	if len(uidSet) != 3 {
		t.Errorf("Expected 3 unique IDs, got %d", len(uidSet))
	} else {
		t.Logf("✓ All 3 Text nodes have unique IDs")
	}
}

// TestTreeViewNestedUniqueness tests that nested structures have unique IDs
func TestTreeViewNestedUniqueness(t *testing.T) {
	tv := NewTreeView()
	tv.maxDepth = 5

	// Create nested structures with same type at different levels
	// This tests that different levels create unique IDs
	root := ui.VStack(
		ui.VStack(
			ui.Text("A"),
			ui.Text("B"),
		),
		ui.VStack(
			ui.Text("C"),
			ui.Text("D"),
		),
	)

	err := tv.SetRoot(root)
	if err != nil {
		t.Fatalf("SetRoot failed: %v", err)
	}

	// Get all nodes
	allNodes := tv.GetFlatList()

	// Collect all UniqueIDs
	uniqueIDs := make(map[string]bool)
	uidList := make([]string, 0, len(allNodes))

	for _, node := range allNodes {
		uid := node.UniqueID
		uidList = append(uidList, uid)

		// Check for duplicates
		if uniqueIDs[uid] {
			t.Errorf("❌ DUPLICATE UniqueID found: %s", uid)
		}
		uniqueIDs[uid] = true

		t.Logf("Node: %-30s UniqueID: %s", node.Info.Type, uid)
	}

	// Verify we have no duplicates
	if len(uniqueIDs) != len(allNodes) {
		t.Errorf("❌ Expected %d unique IDs, got %d", len(allNodes), len(uniqueIDs))
	} else {
		t.Logf("✓ All %d nodes have unique IDs", len(allNodes))
	}

	// Verify that nodes at the same level but different branches have unique IDs
	// Collect all nodes with vstack in their path
	var vstackPathNodes []string
	for _, node := range allNodes {
		if strings.Contains(node.Path, "vstack") {
			vstackPathNodes = append(vstackPathNodes, node.UniqueID)
		}
	}

	if len(vstackPathNodes) > 0 {
		t.Logf("✓ Found %d nodes with vstack in path", len(vstackPathNodes))
		for i, uid := range vstackPathNodes {
			t.Logf("  VStackPath[%d]: %s", i, uid)
		}

		// Verify all are unique
		uidSet := make(map[string]bool)
		for _, uid := range vstackPathNodes {
			if uidSet[uid] {
				t.Errorf("❌ Duplicate vstack path node ID: %s", uid)
			}
			uidSet[uid] = true
		}
		if len(uidSet) == len(vstackPathNodes) {
			t.Logf("✓ All vstack path nodes have unique IDs")
		}
	}
}

// TestTreeViewIDStability tests that IDs remain stable across rebuilds
func TestTreeViewIDStability(t *testing.T) {
	tv := NewTreeView()
	tv.maxDepth = 5

	// Create a tree
	root := ui.VStack(
		ui.Text("A"),
		ui.Text("B"),
	)

	err := tv.SetRoot(root)
	if err != nil {
		t.Fatalf("SetRoot failed: %v", err)
	}

	// Get UniqueIDs from first build
	nodes1 := tv.GetFlatList()
	uids1 := make(map[string]string)
	for _, node := range nodes1 {
		if strings.Contains(node.Info.Type, "Text") {
			uids1[node.Info.Label] = node.UniqueID
		}
	}

	// Rebuild the same tree
	err = tv.SetRoot(root)
	if err != nil {
		t.Fatalf("SetRoot failed on second build: %v", err)
	}

	// Get UniqueIDs from second build
	nodes2 := tv.GetFlatList()
	uids2 := make(map[string]string)
	for _, node := range nodes2 {
		if strings.Contains(node.Info.Type, "Text") {
			uids2[node.Info.Label] = node.UniqueID
		}
	}

	// Verify IDs are stable (same after rebuild)
	for label, uid1 := range uids1 {
		uid2, exists := uids2[label]
		if !exists {
			t.Errorf("Label %s not found in second build", label)
			continue
		}
		if uid1 != uid2 {
			t.Errorf("❌ ID not stable for label %q: %s -> %s", label, uid1, uid2)
		} else {
			t.Logf("✓ ID stable for label %q: %s", label, uid1)
		}
	}
}
