package ui

import (
	"testing"
)

// TestNodeIDGeneration verifies that Fiber NodeIDs are unique and sequential
func TestNodeIDGeneration(t *testing.T) {
	t.Log("=== Testing Fiber NodeID Generation ===")

	// Create a simple tree with 3 text nodes
	vnode := VStack(
		Text("Child 1"),
		Text("Child 2"),
		Text("Child 3"),
	)

	// Build Fiber tree
	root := CreateFiberFromVNode(vnode)

	if root == nil {
		t.Fatal("CreateFiberFromVNode returned nil")
	}

	t.Logf("Root Fiber: NodeID=%d, Type=%d", root.NodeID, root.Type)

	// Verify NodeIDs
	var nodeIDs []uint64
	var collectNodeIDs func(*Fiber)
	collectNodeIDs = func(f *Fiber) {
		if f == nil {
			return
		}
		nodeIDs = append(nodeIDs, f.NodeID)
		t.Logf("  Fiber: NodeID=%d, Type=%d", f.NodeID, f.Type)

		// Collect children
		for child := f.Child; child != nil; child = child.Sibling {
			collectNodeIDs(child)
		}
	}
	collectNodeIDs(root)

	// Verify we have 4 fibers (1 root + 3 children)
	if len(nodeIDs) != 4 {
		t.Errorf("Expected 4 NodeIDs, got %d", len(nodeIDs))
	}

	// Verify all NodeIDs are non-zero
	for i, nid := range nodeIDs {
		if nid == 0 {
			t.Errorf("NodeID %d is 0", i)
		}
	}

	// Verify all NodeIDs are unique
	uniqueIDs := make(map[uint64]bool)
	for _, nid := range nodeIDs {
		if uniqueIDs[nid] {
			t.Errorf("Duplicate NodeID=%d found", nid)
		}
		uniqueIDs[nid] = true
	}

	// Verify sequential generation (each NodeID is exactly 1 more than the previous)
	for i := 1; i < len(nodeIDs); i++ {
		if nodeIDs[i] != nodeIDs[i-1]+1 {
			t.Errorf("NodeID %d = %d, want %d (sequential after %d)", i, nodeIDs[i], nodeIDs[i-1]+1, nodeIDs[i-1])
		}
	}

	t.Logf("✅ All NodeIDs are unique and sequential: %v", nodeIDs)
}
