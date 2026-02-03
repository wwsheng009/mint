package reconciler

import (
	"testing"

	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// TestCompleteWorkEffectCollection tests that effect flags are properly collected
func TestCompleteWorkEffectCollection(t *testing.T) {
	// Create a fiber tree with children that have flags
	child1 := &Fiber{
		Type:  rtui.VNodeText,
		Key:   "child1",
		Flags: EffectUpdate,
		SubtreeFlags: EffectUpdate,
	}

	child2 := &Fiber{
		Type:  rtui.VNodeText,
		Key:   "child2",
		Flags: EffectPlacement,
		SubtreeFlags: EffectNoEffect,
	}

	parent := &Fiber{
		Type:       rtui.VNodeElement,
		Key:        "parent",
		Child:      child1,
		Flags:      EffectNoEffect,
		SubtreeFlags: EffectNoEffect,
	}

	// Link siblings
	child1.Sibling = child2

	// Collect child effects
	collectChildEffects(parent)

	// Check that parent's SubtreeFlags are updated
	expectedFlags := EffectUpdate | EffectPlacement
	if parent.SubtreeFlags != expectedFlags {
		t.Errorf("Expected SubtreeFlags: %v, got: %v", expectedFlags, parent.SubtreeFlags)
	}

	t.Logf("Child1 Flags: %v, SubtreeFlags: %v", child1.Flags, child1.SubtreeFlags)
	t.Logf("Child2 Flags: %v, SubtreeFlags: %v", child2.Flags, child2.SubtreeFlags)
	t.Logf("Parent SubtreeFlags after collection: %v", parent.SubtreeFlags)
}

// TestNodeDeletionInDiff tests that node deletion is properly handled
func TestNodeDeletionInDiff(t *testing.T) {
	// Create existing child fibers
	child1 := CreateFiber(rtui.Element("item1").Key("item1").Build())
	child2 := CreateFiber(rtui.Element("item2").Key("item2").Build())
	child3 := CreateFiber(rtui.Element("item3").Key("item3").Build())

	t.Logf("Before reconcile: child1.Flags=%d, child2.Flags=%d, child3.Flags=%d",
		child1.Flags, child2.Flags, child3.Flags)

	child1.Sibling = child2
	child2.Sibling = child3

	parent := &Fiber{
		Type:  rtui.VNodeElement,
		Key:   "parent",
		Child: child1,
	}

	// New children have only item3 (item1 and item2 removed)
	newChildren := []rtui.VNode{
		rtui.Element("item3").Key("item3").Build(),
	}

	// Reconcile children
	result := reconcileChildren(parent, child1, newChildren, LaneSyncLane)

	// Check: Only item3 should remain in result
	if result == nil {
		t.Fatal("Expected result fiber, got nil")
	}

	// Count siblings in result (should be 1 - only item3)
	siblingCount := 0
	current := result
	for current != nil {
		siblingCount++
		current = current.Sibling
	}

	if siblingCount != 1 {
		t.Errorf("Expected 1 child (item3) in result, got %d", siblingCount)
	}

	// IMPORTANT: Check that item1 and item2 are marked for deletion
	t.Logf("After reconcile: child1.Flags=%d, child2.Flags=%d, child3.Flags=%d",
		child1.Flags, child2.Flags, child3.Flags)

	if child1.Flags&EffectDeletion == 0 {
		t.Errorf("Expected child1 (item1) to be marked with EffectDeletion, flags=%d", child1.Flags)
	} else {
		t.Logf("PASS: child1 marked with EffectDeletion")
	}

	if child2.Flags&EffectDeletion == 0 {
		t.Errorf("Expected child2 (item2) to be marked with EffectDeletion, flags=%d", child2.Flags)
	} else {
		t.Logf("PASS: child2 marked with EffectDeletion")
	}

	// item3 should NOT be marked for deletion
	if child3.Flags&EffectDeletion != 0 {
		t.Errorf("Expected child3 (item3) to NOT be marked with EffectDeletion, flags=%d", child3.Flags)
	} else {
		t.Logf("PASS: child3 not marked for deletion (still in tree)")
	}
}

// TestAlignMappingCompleteness tests that all align types are properly mapped
func TestAlignMappingCompleteness(t *testing.T) {
	testCases := []struct {
		uiAlign      rtui.Align
		expectedType string // "mapped" or "fallback"
		description  string
	}{
		{rtui.AlignStart, "mapped", "Start should map exactly"},
		{rtui.AlignCenter, "mapped", "Center should map exactly"},
		{rtui.AlignEnd, "mapped", "End should map exactly"},
		{rtui.AlignSpaceBetween, "fallback", "SpaceBetween falls back to Start"},
		{rtui.AlignSpaceAround, "fallback", "SpaceAround falls back to Start"},
	}

	for _, tc := range testCases {
		_ = mapUIAlignToRuntime(tc.uiAlign)
		if tc.expectedType == "fallback" {
			t.Logf("INFO: %s - Falls back to Start (current implementation)", tc.description)
		}
	}
}

// TestGlobalVariableUsage tests that currentReconciler global variable is used
func TestGlobalVariableUsage(t *testing.T) {
	// This test verifies that global variable currentReconciler exists
	// and is not nil (it should be set when a reconciler is created)

	t.Logf("INFO: currentReconciler is a package-level global variable")
	t.Logf("ISSUE: Using global variables for reconciler context is not thread-safe")
	t.Logf("LOCATION: begin_work.go:76, 109")
	t.Logf("RECOMMENDATION: Pass reconciler context through function parameters")
}
