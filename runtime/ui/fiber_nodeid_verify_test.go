package ui

import (
	"testing"
)

// TestFiberNodeIDGeneration verifies that generateNodeID creates unique sequential IDs
func TestFiberNodeIDGeneration(t *testing.T) {
	t.Log("=== Testing NodeID Generation ===")

	// Generate 5 NodeIDs and verify they are sequential
	start := nodeIDGenerator
	for i := 0; i < 5; i++ {
		expected := start + uint64(i) + 1
		nodeID := generateNodeID()
		t.Logf("Generated NodeID %d: expected %d", nodeID, expected)
		if nodeID != expected {
			t.Errorf("NodeID %d = %d, want %d", i, nodeID, expected)
		}
	}

	t.Logf("All NodeIDs are sequential and unique!")
}

// TestFiberNodeIDsInFiberTree verifies that CreateFiberFromVNode creates unique NodeIDs
func TestFiberNodeIDsInFiberTree(t *testing.T) {
	t.Log("=== Testing NodeID in Fiber Tree ===")

	// Create simple VNode tree
	child1 := NewElement("text")
	child1.SetProps(map[string]interface{}{"content": "A"})

	child2 := NewElement("text")
	child2.SetProps(map[string]interface{}{"content": "B"})

	root := NewElement("vstack")
	root.SetChildren([]VNode{child1, child2})

	t.Logf("Created VNode tree with %d children", len(root.Children()))

	// Build Fiber tree
	fiberRoot := CreateFiberFromVNode(root)
	if fiberRoot == nil {
		t.Fatal("Fiber root is nil")
	}

	t.Logf("Fiber root NodeID: %d", fiberRoot.NodeID)

	// Collect all NodeIDs
	var nodeIDs []uint64
	var collectFunc func(*Fiber)
	collectFunc = func(fiber *Fiber) {
		if fiber == nil {
			return
		}
		nodeIDs = append(nodeIDs, fiber.NodeID)
		t.Logf("  Fiber node: NodeID=%d", fiber.NodeID)

		for child := fiber.Child; child != nil; child = child.Sibling {
			collectFunc(child)
		}
	}

	collectFunc(fiberRoot)

	t.Logf("Total fiber nodes: %d", len(nodeIDs))
	t.Logf("All NodeIDs: %v", nodeIDs)

	// Verify count
	if len(nodeIDs) != 3 {
		t.Errorf("Expected 3 fiber nodes, got %d", len(nodeIDs))
	}

	// Verify all non-zero
	for i, nodeID := range nodeIDs {
		if nodeID == 0 {
			t.Errorf("NodeID %d is 0", i)
		}
	}

	// Verify unique
	nodeIDMap := make(map[uint64]bool)
	for _, nodeID := range nodeIDs {
		if nodeIDMap[nodeID] {
			t.Errorf("Duplicate NodeID: %d", nodeID)
		}
		nodeIDMap[nodeID] = true
	}

	t.Logf("All fiber nodes have unique non-zero NodeIDs!")
}
