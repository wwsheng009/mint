package reconciler

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/paint"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// TestBeginWork_Element tests beginWork for element nodes
func TestBeginWork_Element(t *testing.T) {
	elem := rtui.Element("div").Prop("id", "test").Build()
	current := CreateFiber(elem)
	workInProgress := CloneFiber(current)

	if workInProgress == nil {
		t.Fatal("CloneFiber returned nil")
	}

	nextUnitOfWork := BeginWork(current, workInProgress)

	if nextUnitOfWork == nil {
		t.Error("BeginWork should return a valid unit of work")
	}
}

// TestBeginWork_Component tests beginWork for component nodes
func TestBeginWork_Component(t *testing.T) {
	renderCalled := false
	comp := rtui.NewComponent("TestComponent", func() rtui.VNode {
		renderCalled = true
		return rtui.Element("span").Build()
	})
	comp.SetKey("test-comp")

	config := ReconcilerConfig{EnableFiber: true}
	reconciler := NewReconciler(nil, nil, config)
	currentReconciler = reconciler
	defer func() { currentReconciler = nil }()

	current := CreateFiber(comp)
	workInProgress := CloneFiber(current)

	if workInProgress == nil {
		t.Fatal("CloneFiber returned nil")
	}

	nextUnitOfWork := BeginWork(current, workInProgress)

	if nextUnitOfWork == nil {
		t.Error("BeginWork should return a valid unit of work for component")
	}

	// Component should have been rendered
	if !renderCalled {
		t.Error("Component render function should have been called")
	}
}

// TestBeginWork_Fragment tests beginWork for fragment nodes
func TestBeginWork_Fragment(t *testing.T) {
	frag := rtui.Fragment(
		rtui.Element("text").Prop("content", "A").Build(),
		rtui.Element("text").Prop("content", "B").Build(),
	)

	current := CreateFiber(frag)
	workInProgress := CloneFiber(current)

	if workInProgress == nil {
		t.Fatal("CloneFiber returned nil")
	}

	nextUnitOfWork := BeginWork(current, workInProgress)

	if nextUnitOfWork == nil {
		t.Error("BeginWork should return a valid unit of work for fragment")
	}
}

// TestBeginWork_Text tests beginWork for text nodes
func TestBeginWork_Text(t *testing.T) {
	text := rtui.Element("text").Prop("content", "Hello").Build()

	current := CreateFiber(text)
	workInProgress := CloneFiber(current)

	if workInProgress == nil {
		t.Fatal("CloneFiber returned nil")
	}

	nextUnitOfWork := BeginWork(current, workInProgress)

	// Text nodes return themselves (leaf nodes have no children)
	if nextUnitOfWork == nil {
		t.Error("BeginWork for text node should return workInProgress")
	}

	// Text nodes should have no children after BeginWork
	if nextUnitOfWork.Child != nil {
		t.Error("Text node should have no children")
	}
}

// TestCompleteWork_Element tests completeWork for element nodes
func TestCompleteWork_Element(t *testing.T) {
	elem := rtui.Element("div").Prop("id", "test").Build()
	current := CreateFiber(elem)
	workInProgress := CloneFiber(current)

	if workInProgress == nil {
		t.Fatal("CloneFiber returned nil")
	}

	// Process the element with BeginWork
	result := BeginWork(current, workInProgress)

	// After BeginWork, result should be workInProgress for leaf elements
	if result == nil {
		t.Error("BeginWork should return a result")
	}

	// Now complete work - CompleteWork processes effects and bubbles up
	_ = CompleteWork(current, workInProgress)
}

// TestCompleteWork_Component tests completeWork for component nodes
func TestCompleteWork_Component(t *testing.T) {
	comp := rtui.NewComponent("TestComponent", func() rtui.VNode {
		return rtui.Element("span").Build()
	})
	comp.SetKey("test-comp")

	config := ReconcilerConfig{EnableFiber: true}
	reconciler := NewReconciler(nil, nil, config)
	currentReconciler = reconciler
	defer func() { currentReconciler = nil }()

	current := CreateFiber(comp)
	workInProgress := CloneFiber(current)

	if workInProgress == nil {
		t.Fatal("CloneFiber returned nil")
	}

	// Process component
	nextUnitOfWork := BeginWork(current, workInProgress)
	if nextUnitOfWork == nil {
		t.Fatal("BeginWork should return valid unit of work")
	}

	// Complete component work
	next := CompleteWork(current, workInProgress)

	if next == nil {
		t.Error("CompleteWork should return next unit of work (child)")
	}
}

// TestReconcileChildren tests child reconciliation
func TestReconcileChildren(t *testing.T) {
	// Create parent fiber
	parent := &rtui.Fiber{
		Type: rtui.VNodeElement,
		Tag:  "div",
	}

	// Create new children
	newChildren := []rtui.VNode{
		rtui.Element("text").Prop("content", "A").Build(),
		rtui.Element("text").Prop("content", "B").Build(),
		rtui.Element("text").Prop("content", "C").Build(),
	}

	// reconcileChildren returns the first child - assign it to parent.Child
	parent.Child = reconcileChildren(parent, nil, newChildren, rtui.LaneSyncLane)

	if parent.Child == nil {
		t.Fatal("ReconcileChildren should create child fibers")
	}

	// Count children
	count := 0
	for child := parent.Child; child != nil; child = child.Sibling {
		count++
	}

	if count != 3 {
		t.Errorf("ReconcileChildren created %d children, want 3", count)
	}
}

// TestReconcileChildren_Update tests updating existing children
func TestReconcileChildren_Update(t *testing.T) {
	// Create parent with existing children
	oldChildren := []rtui.VNode{
		rtui.Element("text").Prop("content", "A").Build(),
		rtui.Element("text").Prop("content", "B").Build(),
	}

	parent := &rtui.Fiber{
		Type: rtui.VNodeElement,
		Tag:  "div",
	}

	// Set up existing children
	parent.Child = reconcileChildren(parent, nil, oldChildren, rtui.LaneSyncLane)

	if parent.Child == nil {
		t.Fatal("First reconcile should create children")
	}

	// Update with new children (same count, different content)
	newChildren := []rtui.VNode{
		rtui.Element("text").Prop("content", "X").Build(),
		rtui.Element("text").Prop("content", "Y").Build(),
	}

	parent.Child = reconcileChildren(parent, parent.Child, newChildren, rtui.LaneSyncLane)

	// Children should still exist
	if parent.Child == nil {
		t.Fatal("ReconcileChildren should preserve child structure")
	}
}

// TestReconcileChildren_Reduce tests reducing children
func TestReconcileChildren_Reduce(t *testing.T) {
	// Create parent with 3 children
	oldChildren := []rtui.VNode{
		rtui.Element("text").Prop("content", "A").Build(),
		rtui.Element("text").Prop("content", "B").Build(),
		rtui.Element("text").Prop("content", "C").Build(),
	}

	parent := &rtui.Fiber{
		Type: rtui.VNodeElement,
		Tag:  "div",
	}

	reconcileChildren(parent, nil, oldChildren, rtui.LaneSyncLane)

	// Reduce to 1 child
	newChildren := []rtui.VNode{
		rtui.Element("text").Prop("content", "X").Build(),
	}

	parent.Child = reconcileChildren(parent, parent.Child, newChildren, rtui.LaneSyncLane)

	// Count children
	count := 0
	for child := parent.Child; child != nil; child = child.Sibling {
		count++
	}

	if count != 1 {
		t.Errorf("ReconcileChildren should have 1 child after reduction, got %d", count)
	}
}

// TestReconcileChildren_Keys tests key-based reconciliation
func TestReconcileChildren_Keys(t *testing.T) {
	// Create parent with keyed children
	child1 := rtui.Element("text").Prop("content", "A").Build()
	child1.SetKey("key1")
	child2 := rtui.Element("text").Prop("content", "B").Build()
	child2.SetKey("key2")

	oldChildren := []rtui.VNode{child1, child2}

	parent := &rtui.Fiber{
		Type: rtui.VNodeElement,
		Tag:  "div",
	}

	parent.Child = reconcileChildren(parent, nil, oldChildren, rtui.LaneSyncLane)

	if parent.Child == nil {
		t.Fatal("First reconcile should create children")
	}

	// Update with reordered children (same keys, different order)
	newChild1 := rtui.Element("text").Prop("content", "C").Build()
	newChild1.SetKey("key2") // Swapped
	newChild2 := rtui.Element("text").Prop("content", "D").Build()
	newChild2.SetKey("key1") // Swapped

	newChildren := []rtui.VNode{newChild1, newChild2}

	parent.Child = reconcileChildren(parent, parent.Child, newChildren, rtui.LaneSyncLane)

	// Children should still exist with keys
	if parent.Child == nil {
		t.Fatal("ReconcileChildren should preserve keyed children")
	}
}

// TestRender tests full render pipeline
func TestRender(t *testing.T) {
	renderFn := func() rtui.VNode {
		return rtui.Element("div").Prop("id", "test").Build()
	}

	config := ReconcilerConfig{EnableFiber: true}
	reconciler := NewReconciler(nil, renderFn, config)

	// prepareFreshStack creates the fiber tree
	reconciler.prepareFreshStack(renderFn)

	if reconciler.workInProgress == nil {
		t.Error("prepareFreshStack should create workInProgress tree")
	}

	// Run the work loop
	reconciler.workLoopSync()

	if reconciler.root == nil {
		t.Error("workLoopSync should create root fiber tree")
	}
}

// TestPrepareFreshStack tests preparing fresh fiber stack
func TestPrepareFreshStack(t *testing.T) {
	renderFn := func() rtui.VNode {
		return rtui.Element("root").Build()
	}

	config := ReconcilerConfig{EnableFiber: true}
	reconciler := NewReconciler(nil, renderFn, config)

	reconciler.prepareFreshStack(renderFn)

	if reconciler.workInProgress == nil {
		t.Fatal("prepareFreshStack should create work-in-progress fiber")
	}
}

// TestCommitRoot tests committing the fiber tree
func TestCommitRoot(t *testing.T) {
	renderFn := func() rtui.VNode {
		return rtui.Element("div").Build()
	}

	config := ReconcilerConfig{EnableFiber: true}
	reconciler := NewReconciler(nil, renderFn, config)

	reconciler.prepareFreshStack(renderFn)
	reconciler.workLoopSync()

	// After workLoopSync, root should be set (swapped from workInProgress)
	if reconciler.root == nil {
		t.Error("workLoopSync should swap workInProgress to root")
	}

	// CommitRoot requires a buffer which is set up through Render()
	// For this test, we just verify the tree was built
	if reconciler.root != nil {
		// Root was successfully created and swapped
		t.Log("Root fiber successfully created")
	}
}

// TestStats_FullCycle tests reconciler statistics after full render cycle
func TestStats_FullCycle(t *testing.T) {
	renderFn := func() rtui.VNode {
		return rtui.Element("div").
			Child(rtui.Element("text").Prop("content", "A").Build()).
			Build()
	}

	config := ReconcilerConfig{EnableFiber: true}
	reconciler := NewReconciler(nil, renderFn, config)

	reconciler.prepareFreshStack(renderFn)
	reconciler.workLoopSync()

	stats := reconciler.Stats()

	// Should have fiber count
	if stats["fiberCount"] == 0 {
		t.Error("Stats should show fiberCount > 0")
	}
}

// TestShouldUpdate tests whether fibers should update
func TestShouldUpdate(t *testing.T) {
	oldVNode := rtui.Element("div").Prop("id", "test").Build()
	newVNode := rtui.Element("div").Prop("id", "test").Build()

	oldFiber := &rtui.Fiber{
		Type:          rtui.VNodeElement,
		Tag:           "div",
		DiffKey:       "_idx_0",
		SiblingIndex:  0,
		MemoizedProps: oldVNode.Props(),
	}

	if !shouldUpdate(oldFiber, newVNode, 0) {
		t.Error("shouldUpdate should return true for same props")
	}

	newVNodeDiff := rtui.Element("div").Prop("id", "different").Build()

	if !shouldUpdate(oldFiber, newVNodeDiff, 0) {
		t.Error("shouldUpdate should return true for different props")
	}
}

// TestCreateChildFiber tests creating child fibers
func TestCreateChildFiber(t *testing.T) {
	parent := &rtui.Fiber{
		Type: rtui.VNodeElement,
		Tag:  "div",
		Key:  "parent",
	}

	childVNode := rtui.Element("text").Prop("content", "Hello").Build()

	childFiber := createChildFiber(parent, childVNode, LaneSyncLane, 0)

	if childFiber == nil {
		t.Fatal("createChildFiber should return a fiber")
	}

	if childFiber.Return != parent {
		t.Error("childFiber.Return should be parent")
	}

	if childFiber.DiffKey == "" {
		t.Error("childFiber.DiffKey should not be empty")
	}
}

// TestCloneExistingFiber tests cloning existing fibers
func TestCloneExistingFiber(t *testing.T) {
	existing := &rtui.Fiber{
		Type:  rtui.VNodeElement,
		Tag:   "div",
		Key:   "test",
		Props: rtui.Props{"id": "test"},
	}

	newVNode := rtui.Element("div").Prop("id", "test").Build()

	// cloneExistingFiber requires (returnFiber, currentFiber, newVNode, siblingIndex)
	cloned := cloneExistingFiber(nil, existing, newVNode, 0)

	if cloned == nil {
		t.Fatal("cloneExistingFiber should return a fiber")
	}

	if cloned.Key != existing.Key {
		t.Error("cloned fiber should preserve key")
	}

	if cloned.Type != existing.Type {
		t.Error("cloned fiber should preserve type")
	}
}

// TestCompleteWork_TextNode tests completeWork with VNodeText type fiber
func TestCompleteWork_TextNode(t *testing.T) {
	// Create a fiber with VNodeText type directly
	// This tests the code path even though no current VNode impl returns this type
	current := &rtui.Fiber{
		Type: rtui.VNodeText,
		Tag:  "text",
	}
	workInProgress := &rtui.Fiber{
		Type: rtui.VNodeText,
		Tag:  "text",
	}

	result := CompleteWork(current, workInProgress)

	if result == nil {
		t.Error("CompleteWork should return workInProgress for text")
	}
}

// TestCompleteWork_FragmentNode tests completeWork with fragment
func TestCompleteWork_FragmentNode(t *testing.T) {
	frag := rtui.Fragment(
		rtui.Element("text").Prop("content", "A").Build(),
		rtui.Element("text").Prop("content", "B").Build(),
	)

	current := CreateFiber(frag)
	workInProgress := CloneFiber(current)

	if workInProgress == nil {
		t.Fatal("CloneFiber returned nil")
	}

	// Process children first
	_ = BeginWork(current, workInProgress)

	result := CompleteWork(current, workInProgress)

	if result == nil {
		t.Error("CompleteWork should return workInProgress for fragment")
	}
}

// TestCollectChildEffects tests effect collection from children
func TestCollectChildEffects(t *testing.T) {
	// Create a parent fiber
	parent := &rtui.Fiber{
		Type: rtui.VNodeElement,
		Tag:  "div",
	}

	// Create children with effect flags
	child1 := &rtui.Fiber{
		Type:         rtui.VNodeElement,
		Tag:          "span",
		Return:       parent,
		Flags:        rtui.EffectUpdate,
		SubtreeFlags: rtui.EffectPlacement,
	}

	child2 := &rtui.Fiber{
		Type:         rtui.VNodeElement,
		Tag:          "div",
		Return:       parent,
		Flags:        rtui.EffectRef,
		SubtreeFlags: rtui.EffectDeletion,
	}

	// Link children as siblings
	parent.Child = child1
	child1.Sibling = child2

	// Collect effects
	collectChildEffects(parent)

	// Parent should have collected child flags
	if parent.SubtreeFlags == 0 {
		t.Error("Parent should have collected subtree flags from children")
	}

	expectedFlags := rtui.EffectUpdate | rtui.EffectPlacement | rtui.EffectRef | rtui.EffectDeletion
	if parent.SubtreeFlags&expectedFlags == 0 {
		t.Errorf("Parent should have child effect flags, got %v", parent.SubtreeFlags)
	}
}

// TestCollectChildEffects_Nil tests collectChildEffects with nil fiber
func TestCollectChildEffects_Nil(t *testing.T) {
	// Should not panic
	collectChildEffects(nil)
}

// TestCollectChildEffects_NoChildren tests effect collection with no children
func TestCollectChildEffects_NoChildren(t *testing.T) {
	parent := &rtui.Fiber{
		Type: rtui.VNodeElement,
		Tag:  "div",
	}

	// Should not panic
	collectChildEffects(parent)

	if parent.SubtreeFlags != 0 {
		t.Error("Parent with no children should have no subtree flags")
	}
}

// TestBeginWork_VNodeText tests beginWork with VNodeText type fiber
func TestBeginWork_VNodeText(t *testing.T) {
	// Create a fiber with VNodeText type directly
	current := &rtui.Fiber{
		Type: rtui.VNodeText,
		Tag:  "text",
	}
	workInProgress := &rtui.Fiber{
		Type: rtui.VNodeText,
		Tag:  "text",
	}

	result := BeginWork(current, workInProgress)

	if result == nil {
		t.Error("BeginWork should return workInProgress for VNodeText")
	}

	// Text nodes should have no children
	if result.Child != nil {
		t.Error("VNodeText should have no children")
	}
}

// TestReconciler_SetScheduler tests setting the scheduler
func TestReconciler_SetScheduler(t *testing.T) {
	config := ReconcilerConfig{EnableFiber: true}
	reconciler := NewReconciler(nil, nil, config)

	// Create a mock scheduler - we can't use framework.NewApp() in tests
	// because it requires platform setup, but we can test with nil
	reconciler.SetScheduler(nil)

	// Should not panic
}

// TestReconciler_SetInstanceManager tests setting instance manager
func TestReconciler_SetInstanceManager(t *testing.T) {
	config := ReconcilerConfig{EnableFiber: true}
	reconciler := NewReconciler(nil, nil, config)

	_ = reconciler
}

// TestReconciler_SetRenderCallback tests setting render callback
func TestReconciler_SetRenderCallback(t *testing.T) {
	config := ReconcilerConfig{EnableFiber: true}
	reconciler := NewReconciler(nil, nil, config)

	callback := func(fiber *Fiber, x, y int, buffer *paint.Buffer) {
	}

	reconciler.SetRenderCallback(callback)

}

// TestReconciler_GetLayoutBoxes tests getting layout boxes
// func TestReconciler_GetLayoutBoxes(t *testing.T) {
// 	config := ReconcilerConfig{EnableFiber: true}
// 	reconciler := NewReconciler(nil, nil, config)

// 	// Initially should be nil (no layout computed yet)
// 	boxes := reconciler.GetLayoutBoxes()

// 	// Should return a nil slice before any rendering
// 	if boxes != nil {
// 		t.Log("GetLayoutBoxes returns non-nil slice (OK, may be initialized)")
// 	}
// }

// TestGetNextWorkUnit tests depth-first traversal
func TestGetNextWorkUnit(t *testing.T) {
	config := ReconcilerConfig{EnableFiber: true}
	reconciler := NewReconciler(nil, nil, config)

	// Create a simple fiber tree: root -> child -> grandchild
	//                                          -> child2
	grandchild := &rtui.Fiber{Key: "grandchild", Type: rtui.VNodeText}
	child2 := &rtui.Fiber{Key: "child2", Type: rtui.VNodeText}
	child := &rtui.Fiber{Key: "child", Type: rtui.VNodeElement, Child: grandchild, Sibling: child2}
	root := &rtui.Fiber{Key: "root", Type: rtui.VNodeElement, Child: child}

	grandchild.Return = child
	child.Return = root
	child2.Return = root

	// Start at root, should get child
	next := reconciler.getNextWorkUnit(root)
	if next != child {
		t.Errorf("getNextWorkUnit(root) = %v, want child", next)
	}

	// From child, should get grandchild (depth-first)
	next = reconciler.getNextWorkUnit(child)
	if next != grandchild {
		t.Errorf("getNextWorkUnit(child) = %v, want grandchild", next)
	}

	// From grandchild (no child, no sibling), should get child2 (uncle)
	next = reconciler.getNextWorkUnit(grandchild)
	if next != child2 {
		t.Errorf("getNextWorkUnit(grandchild) = %v, want child2", next)
	}

	// From child2 (no child, no sibling, no more siblings up), should get nil
	next = reconciler.getNextWorkUnit(child2)
	if next != nil {
		t.Errorf("getNextWorkUnit(child2) = %v, want nil", next)
	}
}

// TestGetNextWorkUnit_Nil tests getNextWorkUnit with nil
func TestGetNextWorkUnit_Nil(t *testing.T) {
	config := ReconcilerConfig{EnableFiber: true}
	reconciler := NewReconciler(nil, nil, config)

	next := reconciler.getNextWorkUnit(nil)
	if next != nil {
		t.Error("getNextWorkUnit(nil) should return nil")
	}
}

// TestGetNextWorkUnit_SiblingTraversal tests sibling traversal
func TestGetNextWorkUnit_SiblingTraversal(t *testing.T) {
	config := ReconcilerConfig{EnableFiber: true}
	reconciler := NewReconciler(nil, nil, config)

	// Create a tree with siblings only
	sibling1 := &rtui.Fiber{Key: "sibling1", Type: rtui.VNodeText}
	sibling2 := &rtui.Fiber{Key: "sibling2", Type: rtui.VNodeText}
	sibling3 := &rtui.Fiber{Key: "sibling3", Type: rtui.VNodeText}
	root := &rtui.Fiber{Key: "root", Type: rtui.VNodeElement, Child: sibling1}

	sibling1.Return = root
	sibling1.Sibling = sibling2
	sibling2.Return = root
	sibling2.Sibling = sibling3
	sibling3.Return = root

	// From root, should get first child
	next := reconciler.getNextWorkUnit(root)
	if next != sibling1 {
		t.Errorf("Expected sibling1, got %v", next)
	}

	// From sibling1, should get sibling2
	next = reconciler.getNextWorkUnit(sibling1)
	if next != sibling2 {
		t.Errorf("Expected sibling2, got %v", next)
	}

	// From sibling2, should get sibling3
	next = reconciler.getNextWorkUnit(sibling2)
	if next != sibling3 {
		t.Errorf("Expected sibling3, got %v", next)
	}

	// From sibling3, should get nil
	next = reconciler.getNextWorkUnit(sibling3)
	if next != nil {
		t.Errorf("Expected nil, got %v", next)
	}
}

// TestCreateWorkInProgress tests creating work-in-progress fibers
func TestCreateWorkInProgress(t *testing.T) {
	config := ReconcilerConfig{EnableFiber: true}
	reconciler := NewReconciler(nil, nil, config)

	// Create a fiber from a VNode
	vnode := rtui.Element("div").Prop("id", "test").Build()
	current := CreateFiberFromVNode(vnode)

	// Create work-in-progress from existing fiber
	work := reconciler.createWorkInProgress(current, vnode)

	if work == nil {
		t.Fatal("createWorkInProgress should return a fiber")
	}

	if work.Tag != vnode.Tag() {
		t.Error("work.Tag should match vnode.Tag()")
	}

	if work.Lanes != rtui.LaneNoLane {
		t.Errorf("work.Lanes should be NoLane, got %v", work.Lanes)
	}

	if work.Flags != rtui.EffectNoEffect {
		t.Errorf("work.Flags should be NoEffect, got %v", work.Flags)
	}

	if work.Alternate != current {
		t.Error("work.Alternate should point to current")
	}
}

// TestCreateWorkInProgress_New tests creating work-in-progress for new fibers
func TestCreateWorkInProgress_New(t *testing.T) {
	config := ReconcilerConfig{EnableFiber: true}
	reconciler := NewReconciler(nil, nil, config)

	// Create work-in-progress without current fiber (mount)
	vnode := rtui.Element("div").Prop("id", "test").Build()
	work := reconciler.createWorkInProgress(nil, vnode)

	if work == nil {
		t.Fatal("createWorkInProgress should return a fiber for nil current")
	}

	if work.Tag != vnode.Tag() {
		t.Error("work.Tag should match vnode.Tag()")
	}

	if work.Type != rtui.VNodeElement {
		t.Errorf("work.Type should be VNodeElement, got %v", work.Type)
	}

	if work.Tag != "div" {
		t.Errorf("work.Tag should be 'div', got %v", work.Tag)
	}
}

// TestShouldUpdate_KeyMismatch tests shouldUpdate with different keys
func TestShouldUpdate_KeyMismatch(t *testing.T) {
	elem1 := rtui.Element("div").Build()
	elem1.SetKey("key1")

	elem2 := rtui.Element("div").Build()
	elem2.SetKey("key2")

	fiber := CreateFiberFromVNode(elem1)

	// Different keys should not update
	if shouldUpdate(fiber, elem2, 0) {
		t.Error("shouldUpdate should return false for different keys")
	}
}

// TestShouldUpdate_TypeMismatch tests shouldUpdate with different types
func TestShouldUpdate_TypeMismatch(t *testing.T) {
	elem := rtui.Element("div").Build()
	elem.SetKey("same")

	comp := rtui.NewComponent("Test", func() rtui.VNode {
		return rtui.Element("span").Build()
	})
	comp.SetKey("same")

	fiber := CreateFiberFromVNode(elem)

	// Different types should not update even with same key
	if shouldUpdate(fiber, comp, 0) {
		t.Error("shouldUpdate should return false for different types")
	}
}

// TestShouldUpdate_NilValues tests shouldUpdate with nil values
func TestShouldUpdate_NilValues(t *testing.T) {
	elem := rtui.Element("div").Build()

	// Nil current fiber
	if shouldUpdate(nil, elem, 0) {
		t.Error("shouldUpdate should return false for nil current fiber")
	}

	// Create a fiber and test with nil VNode
	fiber := CreateFiberFromVNode(elem)
	if shouldUpdate(fiber, nil, 0) {
		t.Error("shouldUpdate should return false for nil VNode")
	}
}

// TestShouldUpdate_SameComponent tests shouldUpdate with same component
func TestShouldUpdate_SameComponent(t *testing.T) {
	comp1 := rtui.NewComponent("TestComponent", func() rtui.VNode {
		return rtui.Element("span").Build()
	})
	comp1.SetKey("test")

	comp2 := rtui.NewComponent("TestComponent", func() rtui.VNode {
		return rtui.Element("div").Build()
	})
	comp2.SetKey("test")

	fiber := CreateFiberFromVNode(comp1)

	// Same component name and key should update
	if !shouldUpdate(fiber, comp2, 0) {
		t.Error("shouldUpdate should return true for same component")
	}
}

// TestShouldUpdate_DifferentComponent tests shouldUpdate with different components
func TestShouldUpdate_DifferentComponent(t *testing.T) {
	comp1 := rtui.NewComponent("ComponentA", func() rtui.VNode {
		return rtui.Element("span").Build()
	})
	comp1.SetKey("test")

	comp2 := rtui.NewComponent("ComponentB", func() rtui.VNode {
		return rtui.Element("div").Build()
	})
	comp2.SetKey("test")

	fiber := CreateFiberFromVNode(comp1)

	// Different component names should not update
	if shouldUpdate(fiber, comp2, 0) {
		t.Error("shouldUpdate should return false for different components")
	}
}

// TestVNodeConverter_ConvertLayoutNode tests converting LayoutNode
func TestVNodeConverter_ConvertLayoutNode(t *testing.T) {
	converter := NewVNodeConverter()

	// Create a LayoutNode using HStack
	layoutNode := rtui.HStack(
		rtui.Element("text").Prop("content", "A").Build(),
		rtui.Element("text").Prop("content", "B").Build(),
	)

	result := converter.Convert(layoutNode)

	if result == nil {
		t.Error("Convert should return non-nil for LayoutNode")
	}

	// LayoutNode should have children from HStack
	if len(result.Children) == 0 {
		t.Error("HStack should have children after conversion")
	}
}

// TestVNodeConverter_ConvertNested tests converting nested VNodes
func TestVNodeConverter_ConvertNested(t *testing.T) {
	converter := NewVNodeConverter()

	// Create nested structure: VBox containing HStack
	hstack := rtui.HStack(
		rtui.Element("text").Prop("content", "A").Build(),
	)

	vbox := rtui.VStack(
		hstack,
		rtui.Element("text").Prop("content", "B").Build(),
	)

	result := converter.Convert(vbox)

	if result == nil {
		t.Error("Convert should return non-nil for nested VNodes")
	}

	// Should have children
	if len(result.Children) == 0 {
		t.Error("VStack should have children after conversion")
	}
}

// TestVNodeConverter_ConvertFragment tests converting Fragment
func TestVNodeConverter_ConvertFragment(t *testing.T) {
	converter := NewVNodeConverter()

	frag := rtui.Fragment(
		rtui.Element("text").Prop("content", "A").Build(),
		rtui.Element("text").Prop("content", "B").Build(),
	)

	result := converter.Convert(frag)

	if result == nil {
		t.Error("Convert should return non-nil for Fragment")
	}
}

// TestReconciler_HasMoreWork tests hasMoreWork
func TestReconciler_HasMoreWork(t *testing.T) {
	config := ReconcilerConfig{EnableFiber: true}
	reconciler := NewReconciler(nil, nil, config)

	// Initially no work
	if reconciler.hasMoreWork() {
		t.Error("Should have no work initially")
	}

	// Schedule some work
	reconciler.ScheduleUpdate(rtui.LaneSyncLane)

	if !reconciler.hasMoreWork() {
		t.Error("Should have work after scheduling")
	}
}

// TestReconciler_RequestWork tests requestWork
func TestReconciler_RequestWork(t *testing.T) {
	config := ReconcilerConfig{EnableFiber: true}
	reconciler := NewReconciler(nil, nil, config)

	// Request work with nil app should not panic
	reconciler.requestWork()

	// requestWork only does something if app != nil
	// This test verifies it doesn't panic with nil app
}

// TestReconciler_WorkLoopSync_FullTree tests complete work loop
func TestReconciler_WorkLoopSync_FullTree(t *testing.T) {
	renderFn := func() rtui.VNode {
		return rtui.VStack(
			rtui.Element("text").Prop("content", "A").Build(),
			rtui.Element("text").Prop("content", "B").Build(),
		)
	}

	config := ReconcilerConfig{EnableFiber: true}
	reconciler := NewReconciler(nil, renderFn, config)

	reconciler.prepareFreshStack(renderFn)
	reconciler.workLoopSync()

	// Root should be set after work loop
	if reconciler.root == nil {
		t.Error("root should be set after workLoopSync")
	}

	// WorkInProgress should be nil after completion
	if reconciler.workInProgress != nil {
		t.Error("workInProgress should be nil after workLoopSync")
	}
}

// TestReconciler_WorkLoopSync_NestedComponents tests nested component rendering
func TestReconciler_WorkLoopSync_NestedComponents(t *testing.T) {
	innerComp := rtui.NewComponent("Inner", func() rtui.VNode {
		return rtui.Element("text").Prop("content", "Inner").Build()
	})

	renderFn := func() rtui.VNode {
		return rtui.VStack(
			rtui.Element("text").Prop("content", "Outer").Build(),
			innerComp,
		)
	}

	config := ReconcilerConfig{EnableFiber: true}
	reconciler := NewReconciler(nil, renderFn, config)

	reconciler.prepareFreshStack(renderFn)
	reconciler.workLoopSync()

	if reconciler.root == nil {
		t.Error("root should be set for nested components")
	}
}

// TestReconciler_PrepareFreshStack_MultipleRenders tests multiple renders
func TestReconciler_PrepareFreshStack_MultipleRenders(t *testing.T) {
	renderFn1 := func() rtui.VNode {
		return rtui.Element("div").Prop("id", "v1").Build()
	}

	renderFn2 := func() rtui.VNode {
		return rtui.Element("div").Prop("id", "v2").Build()
	}

	config := ReconcilerConfig{EnableFiber: true}
	reconciler := NewReconciler(nil, renderFn1, config)

	// First render
	reconciler.prepareFreshStack(renderFn1)
	firstWIP := reconciler.workInProgress
	if firstWIP == nil {
		t.Fatal("First prepareFreshStack should create workInProgress")
	}

	// Run work loop
	reconciler.workLoopSync()

	// Second render
	reconciler.prepareFreshStack(renderFn2)
	secondWIP := reconciler.workInProgress
	if secondWIP == nil {
		t.Fatal("Second prepareFreshStack should create workInProgress")
	}

	// Should be different from first WIP
	if secondWIP == firstWIP {
		t.Error("Second workInProgress should be different from first")
	}
}

// TestVNodeConverter_ConvertNil tests converting nil VNode
func TestVNodeConverter_ConvertNil(t *testing.T) {
	converter := NewVNodeConverter()

	result := converter.Convert(nil)

	if result != nil {
		t.Error("Convert should return nil for nil VNode")
	}
}

// TestVNodeConverter_ConvertComponent tests converting component
func TestVNodeConverter_ConvertComponent(t *testing.T) {
	converter := NewVNodeConverter()

	comp := rtui.NewComponent("Test", func() rtui.VNode {
		return rtui.Element("text").Prop("content", "Hello").Build()
	})

	result := converter.Convert(comp)

	// Component expansion might return nil if it has no rendered children
	// Just verify it doesn't panic
	_ = result
}

// TestVNodeConverter_GenerateLayoutBoxes_Nil tests with nil node
func TestVNodeConverter_GenerateLayoutBoxes_Nil(t *testing.T) {
	converter := NewVNodeConverter()

	boxes := converter.GenerateLayoutBoxes(nil)

	// GenerateLayoutBoxes returns nil for nil input
	if boxes != nil {
		t.Error("GenerateLayoutBoxes should return nil for nil node")
	}
}

// TestBeginWork_NilFiber tests BeginWork with nil fiber
func TestBeginWork_NilFiber(t *testing.T) {
	result := BeginWork(nil, nil)

	if result != nil {
		t.Error("BeginWork should return nil for nil inputs")
	}
}

// TestCompleteWork_NilFiber tests CompleteWork with nil fiber
func TestCompleteWork_NilFiber(t *testing.T) {
	result := CompleteWork(nil, nil)

	if result != nil {
		t.Error("CompleteWork should return nil for nil inputs")
	}
}

// TestReconcileChildren_EmptyList tests reconciling empty children list
func TestReconcileChildren_EmptyList(t *testing.T) {
	parent := &rtui.Fiber{
		Type: rtui.VNodeElement,
		Tag:  "div",
	}

	// Empty children list
	result := reconcileChildren(parent, nil, []rtui.VNode{}, rtui.LaneSyncLane)

	if result != nil {
		t.Error("ReconcileChildren should return nil for empty children list")
	}
}

// TestCreateChildFiber_WithLanes tests creating child fiber with lanes
func TestCreateChildFiber_WithLanes(t *testing.T) {
	parent := &rtui.Fiber{
		Type: rtui.VNodeElement,
		Tag:  "div",
		Key:  "parent",
	}

	childVNode := rtui.Element("text").Prop("content", "Test").Build()

	// Create with specific lanes
	childFiber := createChildFiber(parent, childVNode, rtui.LaneInputContinuousLane, 0)

	if childFiber == nil {
		t.Fatal("createChildFiber should return a fiber")
	}

	if childFiber.Lanes != rtui.LaneInputContinuousLane {
		t.Errorf("Child fiber lanes should be InputContinuousLane, got %v", childFiber.Lanes)
	}
}

// TestCloneFiber_WithAlternate tests cloning with alternate pointer
func TestCloneFiber_WithAlternate(t *testing.T) {
	original := &rtui.Fiber{
		Key:  "test",
		Type: rtui.VNodeElement,
		Tag:  "div",
	}

	// Set an alternate on the original
	original.Alternate = &rtui.Fiber{Key: "old"}

	cloned := CloneFiber(original)

	if cloned == nil {
		t.Fatal("CloneFiber should return a fiber")
	}

	// CloneFiber copies the Alternate field from original
	if cloned.Alternate != original.Alternate {
		t.Error("Cloned fiber's Alternate should be copied from original")
	}
}

// TestReconciler_Stats_Fresh tests stats on fresh reconciler
func TestReconciler_Stats_Fresh(t *testing.T) {
	config := ReconcilerConfig{EnableFiber: true}
	reconciler := NewReconciler(nil, nil, config)

	stats := reconciler.Stats()

	// Check expected stats exist
	expectedKeys := []string{"hasWork", "lanes", "isWorking", "fiberCount", "instances"}
	for _, key := range expectedKeys {
		if _, ok := stats[key]; !ok {
			t.Errorf("Stats missing key: %s", key)
		}
	}

	// Initially should have no work
	if stats["hasWork"] != false {
		t.Error("Fresh reconciler should have no work")
	}

	// Initially should have 0 fibers
	if stats["fiberCount"] != 0 {
		t.Error("Fresh reconciler should have 0 fibers")
	}
}

// TestCreateFiberFromVNode_Props tests fiber creation with props
func TestCreateFiberFromVNode_Props(t *testing.T) {
	elem := rtui.Element("div").
		Prop("id", "test").
		Prop("class", "container").
		Build()

	fiber := CreateFiberFromVNode(elem)

	if fiber == nil {
		t.Fatal("CreateFiberFromVNode should return a fiber")
	}

	if fiber.Props == nil {
		t.Error("Fiber props should not be nil")
	}

	if fiber.Props.GetString("id") != "test" {
		t.Error("Fiber props should preserve id property")
	}

	if fiber.Props.GetString("class") != "container" {
		t.Error("Fiber props should preserve class property")
	}
}

// TestCreateFiberFromVNode_WithKey tests fiber creation with key
func TestCreateFiberFromVNode_WithKey(t *testing.T) {
	elem := rtui.Element("div").Build()
	elem.SetKey("unique-key")

	fiber := CreateFiberFromVNode(elem)

	if fiber == nil {
		t.Fatal("CreateFiberFromVNode should return a fiber")
	}

	if fiber.Key != "unique-key" {
		t.Errorf("Fiber key should be 'unique-key', got %s", fiber.Key)
	}
}

// TestCreateFiberFromVNode_ComponentWithKey tests component fiber with key
func TestCreateFiberFromVNode_ComponentWithKey(t *testing.T) {
	comp := rtui.NewComponent("MyComponent", func() rtui.VNode {
		return rtui.Element("span").Build()
	})
	comp.SetKey("my-component-key")

	fiber := CreateFiberFromVNode(comp)

	if fiber == nil {
		t.Fatal("CreateFiberFromVNode should return a fiber for component")
	}

	if fiber.Key != "my-component-key" {
		t.Errorf("Fiber key should be 'my-component-key', got %s", fiber.Key)
	}
}

// TestBeginWork_UnknownType tests BeginWork with unknown type
func TestBeginWork_UnknownType(t *testing.T) {
	// Create a fiber with unknown type (not one of the standard types)
	current := &rtui.Fiber{
		Type: rtui.VNodeType(99), // Unknown type
	}
	workInProgress := &rtui.Fiber{
		Type: rtui.VNodeType(99), // Unknown type
	}

	// Should not panic, return workInProgress
	result := BeginWork(current, workInProgress)

	if result != workInProgress {
		t.Error("BeginWork should return workInProgress for unknown type")
	}
}

// TestCompleteWork_UnknownType tests CompleteWork with unknown type
func TestCompleteWork_UnknownType(t *testing.T) {
	current := &rtui.Fiber{
		Type: rtui.VNodeType(99), // Unknown type
	}
	workInProgress := &rtui.Fiber{
		Type: rtui.VNodeType(99), // Unknown type
	}

	result := CompleteWork(current, workInProgress)

	if result != workInProgress {
		t.Error("CompleteWork should return workInProgress for unknown type")
	}
}

// TestMergeLanes_AllLanes tests merging all lane types
func TestMergeLanes_AllLanes(t *testing.T) {
	lanes := MergeLanes(
		rtui.LaneSyncLane,
		rtui.LaneInputContinuousLane,
	)
	lanes = MergeLanes(lanes, rtui.LaneDefaultLane)
	lanes = MergeLanes(lanes, rtui.LaneIdleLane)

	expected := rtui.Lane(1 | 2 | 4 | 8) // All lanes combined
	if lanes != expected {
		t.Errorf("MergeLanes should combine all lanes, got %v, want %v", lanes, expected)
	}
}

// TestVNodeConverter_ConvertElementWithChildren tests element with children
func TestVNodeConverter_ConvertElementWithChildren(t *testing.T) {
	converter := NewVNodeConverter()

	elem := rtui.Element("div").
		Child(
			rtui.Element("text").Prop("content", "Child 1").Build(),
		).
		Child(
			rtui.Element("text").Prop("content", "Child 2").Build(),
		).
		Build()

	result := converter.Convert(elem)

	if result == nil {
		t.Error("Convert should return non-nil for element with children")
	}

	// Should have children
	if len(result.Children) == 0 {
		t.Error("Element with children should have converted children")
	}
}

// TestReconciler_ScheduleUpdate_MultipleLanes tests scheduling multiple lanes
func TestReconciler_ScheduleUpdate_MultipleLanes(t *testing.T) {
	config := ReconcilerConfig{EnableFiber: true}
	reconciler := NewReconciler(nil, nil, config)

	// Schedule different lanes
	reconciler.ScheduleUpdate(rtui.LaneSyncLane)
	reconciler.ScheduleUpdate(rtui.LaneDefaultLane)
	reconciler.ScheduleUpdate(rtui.LaneIdleLane)

	// All lanes should be merged
	expectedLanes := rtui.LaneSyncLane | rtui.LaneDefaultLane | rtui.LaneIdleLane
	if reconciler.lanes != expectedLanes {
		t.Errorf("Lanes should be merged to %v, got %v", expectedLanes, reconciler.lanes)
	}
}

// TestVNodeConverter_IDGeneration tests unique ID generation
func TestVNodeConverter_IDGeneration(t *testing.T) {
	converter := NewVNodeConverter()

	elem1 := rtui.Element("div").Build()
	elem2 := rtui.Element("span").Build()

	// Convert twice - IDs should be different
	result1 := converter.Convert(elem1)
	result2 := converter.Convert(elem2)

	if result1 == nil || result2 == nil {
		t.Fatal("Convert should return non-nil")
	}

	// IDs should be different for different VNodes
	if result1.ID == result2.ID {
		t.Error("Different VNodes should have different IDs")
	}
}

// TestVNodeConverter_IDGeneration_SameVNode tests ID for same VNode
func TestVNodeConverter_IDGeneration_SameVNode(t *testing.T) {
	converter := NewVNodeConverter()

	elem := rtui.Element("div").Build()

	// Convert same VNode twice - IDs are sequential (counter-based)
	result1 := converter.Convert(elem)
	result2 := converter.Convert(elem)

	if result1 == nil || result2 == nil {
		t.Fatal("Convert should return non-nil")
	}

	// IDs should be different (counter increments)
	if result1.ID == result2.ID {
		t.Error("Each convert should generate a unique ID (counter-based)")
	}
}

// TestReconciler_MarkDirty_SchedulesWork tests MarkDirty schedules work
func TestReconciler_MarkDirty_SchedulesWork(t *testing.T) {
	config := ReconcilerConfig{EnableFiber: true}
	reconciler := NewReconciler(nil, nil, config)

	// MarkDirty should not panic (even with nil app)
	reconciler.MarkDirty()

	// With nil app, MarkDirty doesn't schedule work
	// Just verify it doesn't panic
}

// TestFiber_ChildSiblingChain tests fiber child/sibling chain
func TestFiber_ChildSiblingChain(t *testing.T) {
	// Create a chain: parent -> child1 -> child2 -> child3
	child1 := &rtui.Fiber{Key: "child1"}
	child2 := &rtui.Fiber{Key: "child2", Sibling: child1}
	child3 := &rtui.Fiber{Key: "child3", Sibling: child2}
	parent := &rtui.Fiber{Key: "parent", Child: child3}

	// Verify the chain
	if parent.Child != child3 {
		t.Error("Parent's child should be child3")
	}

	if child3.Sibling != child2 {
		t.Error("child3's sibling should be child2")
	}

	if child2.Sibling != child1 {
		t.Error("child2's sibling should be child1")
	}

	if child1.Sibling != nil {
		t.Error("child1 should have no sibling")
	}
}

// TestFiber_ReturnLink tests fiber return (parent) link
func TestFiber_ReturnLink(t *testing.T) {
	child := &rtui.Fiber{Key: "child"}
	parent := &rtui.Fiber{Key: "parent", Child: child}

	// Set return link manually (normally done by createChildFiber)
	child.Return = parent

	if child.Return != parent {
		t.Error("Child's Return should point to parent")
	}
}

// TestCloneExistingFiber_WithProps tests cloning with different props
func TestCloneExistingFiber_WithProps(t *testing.T) {
	existing := &rtui.Fiber{
		Type:  rtui.VNodeElement,
		Tag:   "div",
		Key:   "test",
		Props: rtui.Props{"id": "old"},
	}

	newVNode := rtui.Element("div").Prop("id", "new").Build()

	cloned := cloneExistingFiber(nil, existing, newVNode, 0)

	if cloned == nil {
		t.Fatal("cloneExistingFiber should return a fiber")
	}

	// Props should be updated to new VNode's props
	if cloned.Props.GetString("id") != "new" {
		t.Errorf("Cloned props should be updated, got %v", cloned.Props)
	}
}

// =============================================================================
// Mixed Key Strategy Integration Tests
// =============================================================================

// TestPathBasedKeyStrategy_Integration tests the full path-based key strategy
// with instance management and reconciliation
func TestPathBasedKeyStrategy_Integration(t *testing.T) {
	// Initialize path generator
	pg := NewPathGenerator()

	// Create parent fiber (simulating a vstack)
	parent := &Fiber{
		Key:  "/root/base[0]/vstack[0]",
		Path: "/root/base[0]/vstack[0]",
		Tag:  "vstack",
		Type: rtui.VNodeElement,
	}

	// Create children VNodes (static UI, no user keys)
	children := []rtui.VNode{
		rtui.Element("panel").Build(),
		rtui.Element("panel").Build(),
		rtui.Element("button").Build(),
	}

	// Use createAllNewChildren to generate paths
	pathGenerator = pg
	firstChild := createAllNewChildren(parent, children, LaneSyncLane)

	// Collect all children
	var allChildren []*Fiber
	for child := firstChild; child != nil; child = child.Sibling {
		allChildren = append(allChildren, child)
	}

	// Verify we have 3 children
	if len(allChildren) != 3 {
		t.Fatalf("Expected 3 children, got %d", len(allChildren))
	}

	panel1 := allChildren[0]
	expectedPanel1Path := "/root/base[0]/vstack[0]/panel[0]"
	if panel1.Key != "_idx_0" {
		t.Errorf("First panel key should be %q (index fallback), got %q", "_idx_0", panel1.Key)
	}
	if panel1.Path != expectedPanel1Path {
		t.Errorf("First panel path should be %q, got %q", expectedPanel1Path, panel1.Path)
	}

	panel2 := allChildren[1]
	expectedPanel2Path := "/root/base[0]/vstack[0]/panel[1]"
	if panel2.Key != "_idx_1" {
		t.Errorf("Second panel key should be %q (index fallback), got %q", "_idx_1", panel2.Key)
	}
	if panel2.Path != expectedPanel2Path {
		t.Errorf("Second panel path should be %q, got %q", expectedPanel2Path, panel2.Path)
	}

	button := allChildren[2]
	expectedButtonPath := "/root/base[0]/vstack[0]/button[0]"
	if button.Key != "_idx_2" {
		t.Errorf("Button key should be %q (index fallback for 3rd child), got %q", "_idx_2", button.Key)
	}
	if button.Path != expectedButtonPath {
		t.Errorf("Button path should be %q, got %q", expectedButtonPath, button.Path)
	}

	t.Logf("✅ Path-based key strategy integration test passed")
	t.Logf("   Panel1: Key=%s, Path=%s", panel1.Key, panel1.Path)
	t.Logf("   Panel2: Key=%s, Path=%s", panel2.Key, panel2.Path)
	t.Logf("   Button: Key=%s, Path=%s", button.Key, button.Path)
}

// TestUserKeyPriority_Integration tests that user keys override path-based keys
func TestUserKeyPriority_Integration(t *testing.T) {
	// Initialize path generator
	pg := NewPathGenerator()

	// Create parent fiber (simulating a vstack)
	parent := &Fiber{
		Key:  "/root/base[0]/vstack[0]",
		Path: "/root/base[0]/vstack[0]",
		Tag:  "vstack",
		Type: rtui.VNodeElement,
	}

	// Create children VNodes WITH user keys
	button1VNode := rtui.Element("button").Key("save-btn").Build()
	button2VNode := rtui.Element("button").Key("cancel-btn").Build()

	children := []rtui.VNode{button1VNode, button2VNode}

	// Use createAllNewChildren to generate paths
	pathGenerator = pg
	firstChild := createAllNewChildren(parent, children, LaneSyncLane)

	// Collect all children
	var allChildren []*Fiber
	for child := firstChild; child != nil; child = child.Sibling {
		allChildren = append(allChildren, child)
	}

	// Verify we have 2 children
	if len(allChildren) != 2 {
		t.Fatalf("Expected 2 children, got %d", len(allChildren))
	}

	// Verify first button uses user key
	button1 := allChildren[0]
	if button1.Key != "save-btn" {
		t.Errorf("First button key should be 'save-btn', got %q", button1.Key)
	}
	if button1.Path != "/root/base[0]/vstack[0]/button[0]/key[save-btn]" {
		t.Errorf("First button path should include user key, got %q", button1.Path)
	}

	// Verify second button uses user key
	button2 := allChildren[1]
	if button2.Key != "cancel-btn" {
		t.Errorf("Second button key should be 'cancel-btn', got %q", button2.Key)
	}
	// ✨ New Design: Path includes type index (button[1] for second button)
	if button2.Path != "/root/base[0]/vstack[0]/button[1]/key[cancel-btn]" {
		t.Errorf("Second button path should include user key, got %q", button2.Path)
	}

	t.Logf("✅ User key priority integration test passed")
	t.Logf("   Button1: Key=%s, Path=%s", button1.Key, button1.Path)
	t.Logf("   Button2: Key=%s, Path=%s", button2.Key, button2.Path)
}
