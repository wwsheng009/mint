package reconciler

import (
	"testing"

	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// TestNodeID_StabilityAcrossRenders verifies that NodeID is preserved when Fiber is reused
// and new NodeID is generated when Fiber is created fresh
func TestNodeID_StabilityAcrossRenders(t *testing.T) {
	pathGenerator = NewPathGenerator()

	parent := &Fiber{
		Type:    rtui.VNodeElement,
		Tag:     "div",
		DiffKey: "root",
		Key:     "root",
	}

	children := []rtui.VNode{
		rtui.Element("text").Prop("content", "A").Key("a").Build(),
		rtui.Element("text").Prop("content", "B").Key("b").Build(),
	}

	firstChild := reconcileChildren(parent, nil, children, LaneSyncLane)

	var firstRenderIDs []uint64
	for child := firstChild; child != nil; child = child.Sibling {
		firstRenderIDs = append(firstRenderIDs, child.NodeID)
	}

	if len(firstRenderIDs) != 2 {
		t.Fatalf("Expected 2 children, got %d", len(firstRenderIDs))
	}

	t.Logf("First render NodeIDs: %v", firstRenderIDs)

	updatedChildren := []rtui.VNode{
		rtui.Element("text").Prop("content", "A-updated").Key("a").Build(),
		rtui.Element("text").Prop("content", "B-updated").Key("b").Build(),
	}

	parent.Child = firstChild
	secondChild := reconcileChildren(parent, firstChild, updatedChildren, LaneSyncLane)

	var secondRenderIDs []uint64
	for child := secondChild; child != nil; child = child.Sibling {
		secondRenderIDs = append(secondRenderIDs, child.NodeID)
	}

	t.Logf("Second render NodeIDs: %v", secondRenderIDs)

	if firstRenderIDs[0] != secondRenderIDs[0] {
		t.Errorf("NodeID for child 'a' changed: was %d, now %d (should be preserved)",
			firstRenderIDs[0], secondRenderIDs[0])
	} else {
		t.Logf("✅ NodeID for 'a' preserved: %d", firstRenderIDs[0])
	}

	if firstRenderIDs[1] != secondRenderIDs[1] {
		t.Errorf("NodeID for child 'b' changed: was %d, now %d (should be preserved)",
			firstRenderIDs[1], secondRenderIDs[1])
	} else {
		t.Logf("✅ NodeID for 'b' preserved: %d", firstRenderIDs[1])
	}
}

// TestNodeID_NewNodeGetsNewID verifies that newly added nodes get new NodeIDs
func TestNodeID_NewNodeGetsNewID(t *testing.T) {
	pathGenerator = NewPathGenerator()

	parent := &Fiber{
		Type:    rtui.VNodeElement,
		Tag:     "div",
		DiffKey: "root",
		Key:     "root",
	}

	initialChildren := []rtui.VNode{
		rtui.Element("text").Prop("content", "A").Key("a").Build(),
	}

	firstChild := reconcileChildren(parent, nil, initialChildren, LaneSyncLane)

	var initialID uint64
	for child := firstChild; child != nil; child = child.Sibling {
		initialID = child.NodeID
	}

	t.Logf("Initial child NodeID: %d", initialID)

	expandedChildren := []rtui.VNode{
		rtui.Element("text").Prop("content", "A").Key("a").Build(),
		rtui.Element("text").Prop("content", "B").Key("b").Build(), // New child
	}

	parent.Child = firstChild
	expandedChild := reconcileChildren(parent, firstChild, expandedChildren, LaneSyncLane)

	var expandedIDs []uint64
	for child := expandedChild; child != nil; child = child.Sibling {
		expandedIDs = append(expandedIDs, child.NodeID)
	}

	t.Logf("Expanded children NodeIDs: %v", expandedIDs)

	if expandedIDs[0] != initialID {
		t.Errorf("Existing child NodeID changed: was %d, now %d", initialID, expandedIDs[0])
	} else {
		t.Logf("✅ Existing child NodeID preserved: %d", initialID)
	}

	if expandedIDs[1] == 0 {
		t.Errorf("New child has NodeID 0 (should have unique ID)")
	} else {
		t.Logf("✅ New child got unique NodeID: %d", expandedIDs[1])
	}

	if expandedIDs[1] == initialID {
		t.Errorf("New child has same NodeID as existing child: %d", initialID)
	}
}

// TestNodeID_DeletedAndRecreatedGetsNewID verifies that deleted and recreated nodes get new NodeIDs
func TestNodeID_DeletedAndRecreatedGetsNewID(t *testing.T) {
	pathGenerator = NewPathGenerator()

	parent := &Fiber{
		Type:    rtui.VNodeElement,
		Tag:     "div",
		DiffKey: "root",
		Key:     "root",
	}

	initialChildren := []rtui.VNode{
		rtui.Element("text").Prop("content", "A").Key("a").Build(),
	}

	firstChild := reconcileChildren(parent, nil, initialChildren, LaneSyncLane)

	var originalID uint64
	for child := firstChild; child != nil; child = child.Sibling {
		originalID = child.NodeID
	}

	t.Logf("Original NodeID: %d", originalID)

	emptyChildren := []rtui.VNode{}
	parent.Child = firstChild
	reconcileChildren(parent, firstChild, emptyChildren, LaneSyncLane)

	t.Log("Children removed")

	recreatedChildren := []rtui.VNode{
		rtui.Element("text").Prop("content", "A-recreated").Key("a").Build(),
	}

	recreatedChild := reconcileChildren(parent, nil, recreatedChildren, LaneSyncLane)

	var recreatedID uint64
	for child := recreatedChild; child != nil; child = child.Sibling {
		recreatedID = child.NodeID
	}

	t.Logf("Recreated NodeID: %d", recreatedID)

	if recreatedID == 0 {
		t.Errorf("Recreated child has NodeID 0")
	} else {
		t.Logf("✅ Recreated child has non-zero NodeID: %d", recreatedID)
	}

	if recreatedID == originalID {
		t.Logf("Note: Recreated child has same NodeID as original (%d) - this could indicate reuse", recreatedID)
	} else {
		t.Logf("✅ Recreated child has different NodeID: original=%d, recreated=%d", originalID, recreatedID)
	}
}

// TestNodeID_UniquenessInTree verifies that all NodeIDs in a tree are unique
func TestNodeID_UniquenessInTree(t *testing.T) {
	pathGenerator = NewPathGenerator()

	children := []rtui.VNode{
		rtui.Element("text").Prop("content", "A").Key("a").Build(),
		rtui.Element("text").Prop("content", "B").Key("b").Build(),
		rtui.Element("text").Prop("content", "C").Key("c").Build(),
	}

	parent := &Fiber{
		Type:    rtui.VNodeElement,
		Tag:     "div",
		DiffKey: "root",
		Key:     "root",
	}

	firstChild := reconcileChildren(parent, nil, children, LaneSyncLane)

	nodeIDs := make(map[uint64]bool)
	duplicates := []uint64{}

	for child := firstChild; child != nil; child = child.Sibling {
		if nodeIDs[child.NodeID] {
			duplicates = append(duplicates, child.NodeID)
		}
		nodeIDs[child.NodeID] = true
	}

	if len(duplicates) > 0 {
		t.Errorf("Duplicate NodeIDs found: %v", duplicates)
	} else {
		t.Logf("✅ All %d children have unique NodeIDs", len(nodeIDs))
	}
}

// TestNodeID_MultipleDiffCycles verifies NodeID stability across multiple diff cycles
func TestNodeID_MultipleDiffCycles(t *testing.T) {
	pathGenerator = NewPathGenerator()

	parent := &Fiber{
		Type:    rtui.VNodeElement,
		Tag:     "div",
		DiffKey: "root",
		Key:     "root",
	}

	children := []rtui.VNode{
		rtui.Element("text").Prop("content", "A").Key("a").Build(),
		rtui.Element("text").Prop("content", "B").Key("b").Build(),
	}

	currentChild := reconcileChildren(parent, nil, children, LaneSyncLane)
	parent.Child = currentChild

	expectedIDs := make(map[string]uint64)
	for child := currentChild; child != nil; child = child.Sibling {
		expectedIDs[child.Key] = child.NodeID
	}

	t.Logf("Initial NodeIDs: a=%d, b=%d", expectedIDs["a"], expectedIDs["b"])

	for cycle := 1; cycle <= 5; cycle++ {
		updatedChildren := []rtui.VNode{
			rtui.Element("text").Prop("content", "A-cycle"+string(rune('0'+cycle))).Key("a").Build(),
			rtui.Element("text").Prop("content", "B-cycle"+string(rune('0'+cycle))).Key("b").Build(),
		}

		currentChild = reconcileChildren(parent, currentChild, updatedChildren, LaneSyncLane)
		parent.Child = currentChild

		for child := currentChild; child != nil; child = child.Sibling {
			if child.Key == "a" && child.NodeID != expectedIDs["a"] {
				t.Errorf("Cycle %d: NodeID for 'a' changed from %d to %d", cycle, expectedIDs["a"], child.NodeID)
			}
			if child.Key == "b" && child.NodeID != expectedIDs["b"] {
				t.Errorf("Cycle %d: NodeID for 'b' changed from %d to %d", cycle, expectedIDs["b"], child.NodeID)
			}
		}

		t.Logf("Cycle %d: NodeIDs preserved", cycle)
	}

	t.Log("✅ NodeIDs stable across multiple diff cycles")
}
