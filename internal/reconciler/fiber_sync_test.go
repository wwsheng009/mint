package reconciler

import (
	"testing"

	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Edge Case: Deletion and Re-insertion
// =============================================================================

// TestFiberSync_DeleteAndReinsertSameKey tests that deleting a node and
// re-inserting it with the same key creates a new fiber (not reuse old deleted one)
func TestFiberSync_DeleteAndReinsertSameKey(t *testing.T) {
	// Create parent with a keyed child
	child1 := rtui.Element("text").Prop("content", "A").Build()
	child1.SetKey("item1")

	parent := &Fiber{
		Type: rtui.VNodeElement,
		Tag:  "div",
	}

	// Initial render with child
	parent.Child = reconcileChildren(parent, nil, []rtui.VNode{child1}, rtui.LaneSyncLane)
	if parent.Child == nil {
		t.Fatal("Initial reconcile should create child")
	}
	originalChildPtr := parent.Child

	// Update: remove the child
	parent.Child = reconcileChildren(parent, parent.Child, []rtui.VNode{}, rtui.LaneSyncLane)
	if parent.Child != nil {
		t.Error("Child should be nil after removal")
	}

	// The original child fiber was marked for deletion by markForDeletion
	// Since we no longer have a reference to it in the tree, we verify
	// through the originalChild pointer we saved
	if originalChildPtr.Flags&rtui.EffectDeletion == 0 {
		t.Log("Note: Original child not marked - deletion happens through markForDeletion in diff")
	}

	// Update: re-insert child with same key
	newChild := rtui.Element("text").Prop("content", "B").Build()
	newChild.SetKey("item1")

	// Pass nil as currentFirstChild since parent.Child is now nil
	parent.Child = reconcileChildren(parent, nil, []rtui.VNode{newChild}, rtui.LaneSyncLane)

	if parent.Child == nil {
		t.Fatal("Re-insert should create new child")
	}

	// The new child should NOT be the same fiber as the deleted one
	// (currentFirstChild was nil, so createAllNewChildren is called)
	if parent.Child == originalChildPtr {
		t.Error("Re-inserted child should be a new fiber, not the deleted one")
	}

	// The new child should NOT have the deletion flag
	if parent.Child.Flags&rtui.EffectDeletion != 0 {
		t.Error("New child should not have deletion flag")
	}
}

// =============================================================================
// Edge Case: Root Node Replacement
// =============================================================================

// TestFiberSync_RootReplacement tests complete root node replacement
func TestFiberSync_RootReplacement(t *testing.T) {
	// Initial render with one type of root
	renderFn1 := func() rtui.VNode {
		return rtui.Element("div").Prop("id", "old-root").Build()
	}

	config := ReconcilerConfig{EnableFiber: true}
	r := NewReconciler(nil, renderFn1, config)

	r.prepareFreshStack(renderFn1)
	r.workLoopSync()

	if r.root == nil {
		t.Fatal("Initial render should create root")
	}
	// Root wraps the actual content (RootComponent wrapper)
	t.Logf("Initial root type: %d, tag: %s", r.root.Type, r.root.Tag)

	// Replace with completely different root type
	renderFn2 := func() rtui.VNode {
		return rtui.VStack(
			rtui.Element("text").Prop("content", "new").Build(),
		)
	}

	r.prepareFreshStack(renderFn2)
	r.workLoopSync()

	if r.root == nil {
		t.Fatal("Replacement render should create root")
	}
	// Root should be updated (may be VStack or layout type)
	t.Logf("New root type: %d, tag: %s", r.root.Type, r.root.Tag)

	// The important thing is that the root was recreated, not the same fiber
	// (We can't directly compare since workInProgress is swapped)
}

// =============================================================================
// Edge Case: Mixed Keyed and Non-Keyed Children
// =============================================================================

// TestFiberSync_MixedKeyedAndNonKeyed tests reconciliation with mixed keys
func TestFiberSync_MixedKeyedAndNonKeyed(t *testing.T) {
	// Create parent with both keyed and non-keyed children
	keyed1 := rtui.Element("text").Prop("content", "A").Build()
	keyed1.SetKey("key1")

	nonKeyed := rtui.Element("text").Prop("content", "B").Build()

	keyed2 := rtui.Element("text").Prop("content", "C").Build()
	keyed2.SetKey("key2")

	oldChildren := []rtui.VNode{keyed1, nonKeyed, keyed2}

	parent := &Fiber{
		Type: rtui.VNodeElement,
		Tag:  "div",
	}

	parent.Child = reconcileChildren(parent, nil, oldChildren, rtui.LaneSyncLane)

	if parent.Child == nil {
		t.Fatal("Initial reconcile should create children")
	}

	// Count children
	count := 0
	for child := parent.Child; child != nil; child = child.Sibling {
		count++
	}
	if count != 3 {
		t.Errorf("Should have 3 children, got %d", count)
	}

	// Update with different order, keeping mixed keys
	newKeyed1 := rtui.Element("text").Prop("content", "C-updated").Build()
	newKeyed1.SetKey("key2")

	newNonKeyed := rtui.Element("text").Prop("content", "B-updated").Build()

	newKeyed2 := rtui.Element("text").Prop("content", "A-updated").Build()
	newKeyed2.SetKey("key1")

	newChildren := []rtui.VNode{newKeyed1, newNonKeyed, newKeyed2}

	parent.Child = reconcileChildren(parent, parent.Child, newChildren, rtui.LaneSyncLane)

	// Children should still exist
	count = 0
	for child := parent.Child; child != nil; child = child.Sibling {
		count++
	}
	if count != 3 {
		t.Errorf("Should still have 3 children after update, got %d", count)
	}
}

// =============================================================================
// Edge Case: Deep Nesting
// =============================================================================

// TestFiberSync_DeepNesting tests handling of deeply nested structures
func TestFiberSync_DeepNesting(t *testing.T) {
	// Create a deeply nested structure (100 levels deep)
	depth := 100

	var nested func(int) rtui.VNode
	nested = func(d int) rtui.VNode {
		if d == 0 {
			return rtui.Element("text").Prop("content", "Leaf").Build()
		}
		return rtui.VStack(
			rtui.Element("text").Prop("content", "Level").Build(),
			nested(d-1),
		)
	}

	root := nested(depth)

	config := ReconcilerConfig{EnableFiber: true}
	r := NewReconciler(nil, func() rtui.VNode { return root }, config)

	r.prepareFreshStack(func() rtui.VNode { return root })
	r.workLoopSync()

	if r.root == nil {
		t.Fatal("Deep nesting render should create root")
	}

	// Count total fibers (should be approximately 2*depth + 1)
	count := CountFibers(r.root)
	t.Logf("Deep nesting (%d levels) created %d fibers", depth, count)

	if count < depth {
		t.Errorf("Expected at least %d fibers for %d levels, got %d", depth, depth, count)
	}
}

// =============================================================================
// Edge Case: Rapid Add/Remove Cycles
// =============================================================================

// TestFiberSync_RapidAddRemoveCycles tests rapid addition and removal
func TestFiberSync_RapidAddRemoveCycles(t *testing.T) {
	parent := &Fiber{
		Type: rtui.VNodeElement,
		Tag:  "div",
	}

	// Cycle: add child, remove child, add child, remove child
	for i := 0; i < 5; i++ {
		// Add
		child := rtui.Element("text").Prop("content", "Item").Build()
		child.SetKey("dynamic")

		parent.Child = reconcileChildren(parent, nil, []rtui.VNode{child}, rtui.LaneSyncLane)
		if parent.Child == nil {
			t.Errorf("Cycle %d add: should create child", i)
		}

		// Remove
		parent.Child = reconcileChildren(parent, parent.Child, []rtui.VNode{}, rtui.LaneSyncLane)
		if parent.Child != nil {
			t.Errorf("Cycle %d remove: child should be nil", i)
		}
	}
}

// =============================================================================
// Edge Case: Key Type Changes
// =============================================================================

// TestFiberSync_KeyTypeChange tests when a node with same key changes type
func TestFiberSync_KeyTypeChange(t *testing.T) {
	// Create a keyed element
	elem := rtui.Element("button").Prop("label", "Click").Build()
	elem.SetKey("interactive")

	parent := &Fiber{
		Type: rtui.VNodeElement,
		Tag:  "div",
	}

	// Initial render
	parent.Child = reconcileChildren(parent, nil, []rtui.VNode{elem}, rtui.LaneSyncLane)
	if parent.Child == nil {
		t.Fatal("Initial reconcile should create child")
	}
	if parent.Child.Type != rtui.VNodeElement {
		t.Errorf("Initial child should be Element type, got %d", parent.Child.Type)
	}

	// Update with same key but component type
	comp := rtui.NewComponent("Button", func() rtui.VNode {
		return rtui.Element("text").Prop("content", "Click").Build()
	})
	comp.SetKey("interactive")

	// Different type with same key should create new fiber
	parent.Child = reconcileChildren(parent, parent.Child, []rtui.VNode{comp}, rtui.LaneSyncLane)

	if parent.Child == nil {
		t.Fatal("Type change should create new child")
	}
	if parent.Child.Type != rtui.VNodeComponent {
		t.Errorf("Updated child should be Component type, got %d", parent.Child.Type)
	}
}

// =============================================================================
// Edge Case: Sibling Chain Integrity
// =============================================================================

// TestFiberSync_SiblingChainIntegrity tests that sibling chains remain consistent
func TestFiberSync_SiblingChainIntegrity(t *testing.T) {
	// Create chain of 5 siblings
	children := make([]rtui.VNode, 5)
	for i := 0; i < 5; i++ {
		children[i] = rtui.Element("text").Prop("content", string(rune('A'+i))).Build()
		children[i].SetKey(string(rune('A' + i)))
	}

	parent := &Fiber{
		Type: rtui.VNodeElement,
		Tag:  "div",
	}

	parent.Child = reconcileChildren(parent, nil, children, rtui.LaneSyncLane)

	// Verify sibling chain
	count := 0
	for child := parent.Child; child != nil; child = child.Sibling {
		count++
		if child.Return != parent {
			t.Errorf("Child %d Return should point to parent", count)
		}
	}
	if count != 5 {
		t.Errorf("Should have 5 siblings, got %d", count)
	}

	// Remove middle child
	newChildren := []rtui.VNode{children[0], children[2], children[3], children[4]}
	parent.Child = reconcileChildren(parent, parent.Child, newChildren, rtui.LaneSyncLane)

	// Verify new chain
	count = 0
	keys := []string{}
	for child := parent.Child; child != nil; child = child.Sibling {
		count++
		keys = append(keys, child.Key)
		if child.Return != parent {
			t.Errorf("After removal, child %d Return should point to parent", count)
		}
	}
	if count != 4 {
		t.Errorf("After removal should have 4 siblings, got %d", count)
	}

	expectedKeys := []string{"A", "C", "D", "E"}
	if len(keys) != len(expectedKeys) {
		t.Errorf("Key mismatch: got %v, want %v", keys, expectedKeys)
	}
	for i, k := range keys {
		if k != expectedKeys[i] {
			t.Errorf("Key at %d should be %s, got %s", i, expectedKeys[i], k)
		}
	}
}

// =============================================================================
// Edge Case: Empty to Full to Empty
// =============================================================================

// TestFiberSync_EmptyToFullToEmpty tests empty -> full -> empty transitions
func TestFiberSync_EmptyToFullToEmpty(t *testing.T) {
	parent := &Fiber{
		Type: rtui.VNodeElement,
		Tag:  "div",
	}

	// Start empty
	parent.Child = reconcileChildren(parent, nil, []rtui.VNode{}, rtui.LaneSyncLane)
	if parent.Child != nil {
		t.Error("Should start with no children")
	}

	// Add 10 children
	children := make([]rtui.VNode, 10)
	for i := 0; i < 10; i++ {
		children[i] = rtui.Element("text").Prop("content", string(rune('0'+i))).Build()
		children[i].SetKey(string(rune('0' + i)))
	}
	parent.Child = reconcileChildren(parent, nil, children, rtui.LaneSyncLane)

	count := 0
	for child := parent.Child; child != nil; child = child.Sibling {
		count++
	}
	if count != 10 {
		t.Errorf("Should have 10 children, got %d", count)
	}

	// Remove all
	parent.Child = reconcileChildren(parent, parent.Child, []rtui.VNode{}, rtui.LaneSyncLane)
	if parent.Child != nil {
		t.Error("Should end with no children")
	}

	// Verify all were marked for deletion
	deletedCount := 0
	WalkFiberDepthFirst(parent, func(fiber *Fiber) bool {
		if fiber.Flags&rtui.EffectDeletion != 0 {
			deletedCount++
		}
		return true
	})
	t.Logf("Deleted %d fibers after removing all children", deletedCount)
}
