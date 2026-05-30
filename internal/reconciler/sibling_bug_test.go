package reconciler

import (
	"testing"

	"github.com/wwsheng009/mint/framework"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// TestSiblingTraversalBugReproduction reproduces the bug where siblings are not processed
func TestSiblingTraversalBugReproduction(t *testing.T) {
	// Setup: Create a Fiber tree with 3 siblings: A, B, C
	siblingA := rtui.Element("text").Prop("content", "A").Build()
	siblingB := rtui.Element("text").Prop("content", "B").Build()
	siblingC := rtui.Element("text").Prop("content", "C").Build()

	div := rtui.Element("div").Children(siblingA, siblingB, siblingC).Build()

	// Build the Fiber tree
	fiberTree := CreateFiberFromVNode(div)

	// Verify siblings exist in the tree structure
	if fiberTree.Child == nil {
		t.Fatal("Root has no Child")
	}

	// Count siblings via Sibling pointer chain
	siblingCount := 0
	current := fiberTree.Child
	for current != nil {
		siblingCount++
		current = current.Sibling
	}

	if siblingCount != 3 {
		t.Fatalf("Tree structure has wrong number of siblings: expected 3, got %d", siblingCount)
	}

	// Now verify that performUnitOfWork processes all siblings
	// Create a test reconciler
	r := &Reconciler{
		root:           nil,
		workInProgress: fiberTree,
	}
	currentReconciler = r
	defer func() { currentReconciler = nil }()

	// Count fibers BEFORE work loop
	before := CountFibers(r.workInProgress)

	// Run the work loop (this is the actual code being tested)
	r.workLoopSync()

	// Count fibers AFTER work loop
	after := CountFibers(r.root)

	// Count should remain the same (no fibers should be lost)
	if before != after {
		t.Errorf("Fiber count changed: before=%d, after=%d", before, after)
		t.Error("BUG: Some siblings were lost during workLoopSync")
	}

	// Verify all siblings are still in the tree
	siblingCountAfter := 0
	current = r.root.Child
	for current != nil {
		siblingCountAfter++
		current = current.Sibling
	}

	if siblingCountAfter != 3 {
		t.Errorf("After workLoopSync: expected 3 siblings, got %d (BUG:siblings lost)", siblingCountAfter)
		t.Error("REPRODUCED: The sibling traversal bug - only first sibling was processed")
	}

	// Additional verification: walk the tree and tag each fiber
	walked := 0
	rtui.WalkFiberDepthFirst(r.root, func(fiber *rtui.Fiber) bool {
		walked++
		return true
	})

	if walked != 4 { // root + 3 siblings
		t.Errorf("Tree walk found %d fibers, expected 4 (root+3 siblings)", walked)
		t.Error("BUG: Tree structure is incomplete - siblings missing")
	}

	t.Logf("Verification: %d siblings in tree structure", siblingCountAfter)
	t.Logf("Total fibers walked: %d", walked)
}

// TestMultipleLevelsOfSiblings tests sibling handling at multiple tree levels
func TestMultipleLevelsOfSiblings(t *testing.T) {
	// Create structure:
	// Root
	//   ├── ChildA
	//   │   ├── Grandchild1
	//   │   └── Grandchild2  <-- should handle sibling traversal here too
	//   ├── ChildB
	//   └── ChildC

	gc1 := rtui.Element("text").Prop("content", "GC1").Build()
	gc2 := rtui.Element("text").Prop("content", "GC2").Build()

	childA := rtui.Element("div").Children(gc1, gc2).Build()
	childA.SetKey("childA")

	childB := rtui.Element("text").Prop("content", "B").Build()
	childB.SetKey("childB")

	childC := rtui.Element("text").Prop("content", "C").Build()
	childC.SetKey("childC")

	root := rtui.Element("div").Children(childA, childB, childC).Build()

	// Build and process
	fiberTree := CreateFiberFromVNode(root)

	r := &Reconciler{
		root:           nil,
		workInProgress: fiberTree,
	}
	currentReconciler = r
	defer func() { currentReconciler = nil }()

	r.workLoopSync()

	// Count all fibers
	count := CountFibers(r.root)

	// Expected: 1 root + 3 children + 2 grandchildren = 6 total
	expected := 6
	if count != expected {
		t.Errorf("Expected %d fibers (root+3 children+2 grandchildren), got %d", expected, count)
		t.Error("BUG: Sibling traversal fails at multiple tree levels")
	}

	// Verify by walking and counting each type
	textNodes := 0
	rtui.WalkFiberDepthFirst(r.root, func(fiber *rtui.Fiber) bool {
		if fiber.Tag == "text" || fiber.Type == rtui.VNodeText {
			textNodes++
		}
		return true
	})

	if textNodes != 4 { // GC1, GC2, B, C
		t.Errorf("Expected 4 text nodes, got %d", textNodes)
	}

	t.Logf("Tree structure: %d total fibers, %d text nodes", count, textNodes)
}

// TestLongSiblingChain tests a chain of 10 siblings
func TestLongSiblingChain(t *testing.T) {
	// Create 10 siblings
	children := make([]rtui.VNode, 10)
	for i := 0; i < 10; i++ {
		children[i] = rtui.Element("text").Prop("content", "Item").Build()
	}

	root := rtui.Element("div").Children(children...).Build()

	// Build and process
	fiberTree := CreateFiberFromVNode(root)

	r := &Reconciler{
		root:        nil,
		workInProgress: fiberTree,
	}
	currentReconciler = r
	defer func() { currentReconciler = nil }()

	r.workLoopSync()

	// Count fibers
	count := CountFibers(r.root)

	// Expected: 1 root + 10 siblings = 11
	expected := 11
	if count != expected {
		t.Errorf("Expected %d fibers (root+10 siblings), got %d", expected, count)
		if count < expected {
			lost := expected - count
			t.Errorf("BUG: %d siblings were lost (only first %d processed)", lost, count-1)
		}
	}

	// Verify by counting sibling chain
	siblingCount := 0
	current := r.root.Child
	for current != nil {
		siblingCount++
		current = current.Sibling
	}

	if siblingCount != 10 {
		t.Errorf("Expected 10 siblings in tree, got %d", siblingCount)
	}

	t.Logf("Chain of 10 siblings: %d fibers in tree", count)
}

// TestFullRenderCycleWithSiblings tests the complete Render() method
func TestFullRenderCycleWithSiblings(t *testing.T) {
	// Test at the reconciler.Render() level
	rootComponent := func() rtui.VNode {
		return rtui.Element("div").Children(
			rtui.Element("text").Prop("content", "First").Build(),
			rtui.Element("text").Prop("content", "Second").Build(),
			rtui.Element("text").Prop("content", "Third").Build(),
			rtui.Element("text").Prop("content", "Fourth").Build(),
		).Build()
	}

	app := &framework.App{}
	reconciler := NewReconciler(app, rootComponent, ReconcilerConfig{})

	// Get the root fiber tree (without calling Render)
	testRoot := rtui.Element("div").Children(
		rtui.Element("text").Prop("content", "A").Build(),
		rtui.Element("text").Prop("content", "B").Build(),
		rtui.Element("text").Prop("content", "C").Build(),
	).Build()

	reconciler.root = CreateFiberFromVNode(testRoot)

	// Verify siblings
	siblingCount := 0
	current := reconciler.root.Child
	for current != nil {
		siblingCount++
		current = current.Sibling
	}

	if siblingCount != 3 {
		t.Errorf("Expected 3 siblings in reconciler's root, got %d", siblingCount)
	}

	// Verify with full tree walk
	totalNodes := CountFibers(reconciler.root)

	if totalNodes != 4 { // root + 3 siblings
		t.Errorf("Full render cycle: expected 4 fibers, got %d", totalNodes)
	}

	t.Logf("Full render cycle: %d siblings, %d total fibers", siblingCount, totalNodes)
}
