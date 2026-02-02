package ui

import (
	"testing"
)

// TestCreateFiber tests basic fiber creation
func TestCreateFiber(t *testing.T) {
	vnode := Text("Hello")
	fiber := CreateFiber(vnode)

	if fiber == nil {
		t.Fatal("CreateFiber() returned nil")
	}

	if fiber.VNode != vnode {
		t.Error("VNode not set correctly")
	}

	if fiber.Type != VNodeElement {
		t.Errorf("Type = %v, want %v", fiber.Type, VNodeElement)
	}

	if fiber.Tag != "text" {
		t.Errorf("Tag = %v, want 'text'", fiber.Tag)
	}

	if fiber.Return != nil {
		t.Error("Return should be nil for root fiber")
	}

	if fiber.Child != nil {
		t.Error("Child should be nil for leaf fiber")
	}

	if fiber.Sibling != nil {
		t.Error("Sibling should be nil for single fiber")
	}
}

// TestCreateFiberFromVNode tests fiber tree creation
func TestCreateFiberFromVNode(t *testing.T) {
	vnode := VStack(
		Text("Child 1"),
		Text("Child 2"),
		Text("Child 3"),
	)

	root := CreateFiberFromVNode(vnode)

	if root == nil {
		t.Fatal("CreateFiberFromVNode() returned nil")
	}

	// Should have children
	if root.Child == nil {
		t.Error("Root fiber should have children")
	}

	// Count fibers
	count := CountFibers(root)
	if count != 4 { // 1 layout + 3 text
		t.Errorf("CountFibers() = %v, want 4", count)
	}
}

// TestFiberTreeStructure tests fiber tree parent-child relationships
func TestFiberTreeStructure(t *testing.T) {
	child1 := Text("A")
	child2 := Text("B")
	parent := VStack(child1, child2)

	root := CreateFiberFromVNode(parent)

	// Root should be layout
	if root.Tag != "layout" {
		t.Errorf("Root Tag = %v, want 'layout'", root.Tag)
	}

	// First child
	firstChild := root.Child
	if firstChild == nil {
		t.Fatal("First child should not be nil")
	}

	if firstChild.Return != root {
		t.Error("First child's Return should point to root")
	}

	// Second child (sibling of first)
	secondChild := firstChild.Sibling
	if secondChild == nil {
		t.Fatal("Second child should not be nil")
	}

	if secondChild.Return != root {
		t.Error("Second child's Return should point to root")
	}

	// Second child should have no more siblings
	if secondChild.Sibling != nil {
		t.Error("Second child should have no sibling")
	}
}

// TestWalkFiberDepthFirst tests depth-first traversal
func TestWalkFiberDepthFirst(t *testing.T) {
	vnode := VStack(
		Text("A"),
		HStack(
			Text("B1"),
			Text("B2"),
		),
		Text("C"),
	)

	root := CreateFiberFromVNode(vnode)

	var visited []string
	WalkFiberDepthFirst(root, func(fiber *Fiber) bool {
		visited = append(visited, fiber.Tag)
		return true
	})

	// Depth-first should visit: root, A, hstack, B1, B2, C
	expectedCount := 6
	if len(visited) != expectedCount {
		t.Errorf("Visited %d nodes, want %d", len(visited), expectedCount)
	}
}

// TestWalkFiberBreadthFirst tests breadth-first traversal
func TestWalkFiberBreadthFirst(t *testing.T) {
	vnode := VStack(
		Text("A"),
		Text("B"),
		Text("C"),
	)

	root := CreateFiberFromVNode(vnode)

	var visited []string
	WalkFiberBreadthFirst(root, func(fiber *Fiber) bool {
		visited = append(visited, fiber.Tag)
		return true
	})

	// Breadth-first should visit: root, A, B, C
	// Note: With nested structures, this would be different
	if len(visited) < 4 {
		t.Errorf("Visited %d nodes, want at least 4", len(visited))
	}
}

// TestWalkFiberEarlyExit tests early exit from traversal
func TestWalkFiberEarlyExit(t *testing.T) {
	vnode := VStack(
		Text("A"),
		Text("B"),
		Text("C"),
	)

	root := CreateFiberFromVNode(vnode)

	visitCount := 0
	WalkFiberDepthFirst(root, func(fiber *Fiber) bool {
		visitCount++
		// Stop after visiting 2 nodes
		return visitCount < 2
	})

	if visitCount != 2 {
		t.Errorf("VisitCount = %d, want 2", visitCount)
	}
}

// TestCloneFiber tests fiber cloning
func TestCloneFiber(t *testing.T) {
	vnode := Text("Hello")
	original := CreateFiber(vnode)

	original.Flags = EffectUpdate
	original.Lanes = LaneSyncLane

	cloned := CloneFiber(original)

	if cloned == nil {
		t.Fatal("CloneFiber() returned nil")
	}

	if cloned.VNode != original.VNode {
		t.Error("Cloned fiber should have same VNode")
	}

	if cloned.Flags != original.Flags {
		t.Errorf("Cloned Flags = %v, want %v", cloned.Flags, original.Flags)
	}

	if cloned.Lanes != original.Lanes {
		t.Errorf("Cloned Lanes = %v, want %v", cloned.Lanes, original.Lanes)
	}

	// Modifying clone should not affect original
	cloned.Flags = EffectPlacement
	if original.Flags == EffectPlacement {
		t.Error("Modifying clone should not affect original")
	}
}

// TestFiberKey tests fiber key handling
func TestFiberKey(t *testing.T) {
	vnode := Text("Hello")
	vnode.SetKey("test-key")

	fiber := CreateFiber(vnode)

	if fiber.Key != "test-key" {
		t.Errorf("Key = %v, want 'test-key'", fiber.Key)
	}
}

// TestFindFiberByKey tests finding fibers by key
func TestFindFiberByKey(t *testing.T) {
	vnode := VStack(
		func() VNode {
			n := Text("A")
			n.SetKey("a")
			return n
		}(),
		func() VNode {
			n := Text("B")
			n.SetKey("b")
			return n
		}(),
		func() VNode {
			n := Text("C")
			n.SetKey("c")
			return n
		}(),
	)

	root := CreateFiberFromVNode(vnode)

	found := FindFiberByKey(root, "b")
	if found == nil {
		t.Fatal("FindFiberByKey() returned nil")
	}

	if found.Tag != "text" {
		t.Errorf("Found Tag = %v, want 'text'", found.Tag)
	}

	// Find non-existent key
	notFound := FindFiberByKey(root, "xyz")
	if notFound != nil {
		t.Error("FindFiberByKey() should return nil for non-existent key")
	}
}

// TestMarkUpdate tests marking fibers for update
func TestMarkUpdate(t *testing.T) {
	parent := CreateFiber(NewElement("div"))
	child := CreateFiber(Text("Hello"))
	child.Return = parent
	parent.Child = child

	// Mark child for update
	child.MarkUpdate(LaneSyncLane)

	if child.Lanes != LaneSyncLane {
		t.Errorf("Child Lanes = %v, want %v", child.Lanes, LaneSyncLane)
	}

	if child.Flags&EffectUpdate == 0 {
		t.Error("Child should have EffectUpdate flag")
	}

	// Parent should have ChildLanes set
	if parent.ChildLanes != LaneSyncLane {
		t.Errorf("Parent ChildLanes = %v, want %v", parent.ChildLanes, LaneSyncLane)
	}
}

// TestHasNoPendingWork tests pending work detection
func TestHasNoPendingWork(t *testing.T) {
	fiber := CreateFiber(Text("Hello"))

	if !fiber.HasNoPendingWork() {
		t.Error("New fiber should have no pending work")
	}

	fiber.Lanes = LaneSyncLane

	if fiber.HasNoPendingWork() {
		t.Error("Fiber with lanes should have pending work")
	}
}

// TestGetFiberDepth tests fiber depth calculation
func TestGetFiberDepth(t *testing.T) {
	vnode := VStack(
		HStack(
			Text("Deep"),
		),
	)

	root := CreateFiberFromVNode(vnode)

	depth := GetFiberDepth(root)
	if depth != 0 {
		t.Errorf("Root depth = %d, want 0", depth)
	}

	child := root.Child
	childDepth := GetFiberDepth(child)
	if childDepth != 1 {
		t.Errorf("Child depth = %d, want 1", childDepth)
	}

	if child.Child != nil {
		grandchild := child.Child
		grandchildDepth := GetFiberDepth(grandchild)
		if grandchildDepth != 2 {
			t.Errorf("Grandchild depth = %d, want 2", grandchildDepth)
		}
	}
}

// TestCollectFibersWithFlags tests collecting fibers with specific flags
func TestCollectFibersWithFlags(t *testing.T) {
	vnode := VStack(
		Text("A"),
		Text("B"),
	)

	root := CreateFiberFromVNode(vnode)

	// Mark some fibers for update
	root.Child.Flags = EffectUpdate
	if root.Child.Sibling != nil {
		root.Child.Sibling.Flags = EffectPlacement
	}

	updated := CollectFibersWithFlags(root, EffectUpdate)

	if len(updated) != 1 {
		t.Errorf("Collected %d fibers, want 1", len(updated))
	}
}

// TestLaneOperations tests lane utility functions
func TestLaneOperations(t *testing.T) {
	// MergeLanes is the only exported lane function now
	merged := MergeLanes(LaneSyncLane, LaneDefaultLane)
	if merged != (LaneSyncLane | LaneDefaultLane) {
		t.Error("MergeLanes failed")
	}
}

// BenchmarkCreateFiber benchmarks fiber creation
func BenchmarkCreateFiber(b *testing.B) {
	vnode := Text("Hello")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CreateFiber(vnode)
	}
}

// BenchmarkCreateFiberFromVNode benchmarks fiber tree creation
func BenchmarkCreateFiberFromVNode(b *testing.B) {
	vnode := VStack(
		Text("A"),
		Text("B"),
		Text("C"),
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CreateFiberFromVNode(vnode)
	}
}
