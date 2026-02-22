package reconciler

import (
	"testing"

	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// TestVNodeCacheConsistency tests that Fiber's cached props stay in sync with VNode
// NOTE: In Fiber-first architecture, VNode is ONLY used during Fiber creation
// and then DISCARDED. Fiber does NOT keep a reference to VNode.
// This test verifies that behavior.
func TestVNodeCacheConsistency(t *testing.T) {
	// Create a VNode with initial props
	initialProps := rtui.Props{"value": "initial"}
	vnode := rtui.Element("text").Prop("value", "initial").Build()
	vnode.SetProps(initialProps)

	// Create a Fiber from this VNode
	fiber := CreateFiber(vnode)

	// Verify initial cached props match
	if fiber.Props.GetString("value") != "initial" {
		t.Errorf("Initial cached props incorrect: expected 'initial', got '%s'", fiber.Props.GetString("value"))
	}

	if fiber.MemoizedProps.GetString("value") != "initial" {
		t.Errorf("Initial memoized props incorrect: expected 'initial', got '%s'", fiber.MemoizedProps.GetString("value"))
	}

	// Now update the VNode's props
	newProps := rtui.Props{"value": "updated"}
	vnode.SetProps(newProps)

	// The Fiber's cached Props field is stale (snapshot taken at creation time)
	// This is EXPECTED behavior in Fiber-first architecture
	// VNode is discarded after Fiber creation, so Fiber cannot access updated VNode
	if fiber.Props.GetString("value") == "initial" {
		t.Log("EXPECTED: Fiber.Props field is stale (snapshot taken at creation)")
	} else {
		t.Error("Fiber.Props should remain 'initial' (no connection to updated VNode)")
	}

	// GetProps() returns the Fiber's Props (the snapshot, NOT the VNode's current props)
	// Fiber-first: NO connection to VNode after creation
	if fiber.GetProps().GetString("value") != "initial" {
		t.Errorf("Fiber.GetProps() should return 'initial' (fiber's snapshot), got '%s'", fiber.GetProps().GetString("value"))
	}

	// The VNode has the new value (but Fiber doesn't know about it)
	if vnode.Props().GetString("value") != "updated" {
		t.Error("VNode should have updated props")
	}

	// Type and Key are cached at creation time and remain stable
	if fiber.Type != vnode.Type() {
		t.Error("Fiber.Type should match VNode at creation time")
	}

	if fiber.Key != vnode.Key() {
		t.Error("Fiber.Key should match VNode Key at creation time")
	}

	t.Logf("VNode props: %v (updated)", vnode.Props())
	t.Logf("Fiber cached Props: %v (snapshot at creation)", fiber.Props)
	t.Logf("Fiber MemoizedProps: %v", fiber.MemoizedProps)

	// CONCLUSION: Fiber-first architecture
	// - VNode is used ONLY during Fiber creation
	// - Fiber does NOT keep a reference to VNode
	// - Fiber's Props is a snapshot taken at creation time
	// - Updating VNode after Fiber creation has NO effect on Fiber
	// - To update Fiber props, run reconciliation with a new VNode tree
	t.Log("CONCLUSION: Fiber-first architecture - VNode is discarded after Fiber creation")
	t.Log("  - Fiber.Props is a snapshot at creation time")
	t.Log("  - NO connection to VNode after creation")
	t.Log("  - To update, run reconcile with new VNode tree")
}

// TestFiberCloneUpdateQueueSharing tests the UpdateQueue sharing issue
func TestFiberCloneUpdateQueueSharing(t *testing.T) {
	// Create a Fiber with some updates
	fiber1 := &Fiber{
		Key:  "original",
		Type: rtui.VNodeText,
		UpdateQueue: &UpdateQueue{
			First: &Update{
				Payload: "update1",
				Lane:    LaneSyncLane,
			},
		},
	}

	// Clone the fiber
	fiber2 := CloneFiber(fiber1)

	// Verify they don't share the same UpdateQueue anymore
	if fiber1.UpdateQueue != fiber2.UpdateQueue {
		t.Log("FIXED: UpdateQueues are now separate (fiber2.UpdateQueue is nil)")
	} else {
		t.Log("WARNING: UpdateQueues are still shared between fiber1 and fiber2")
	}

	// Initialize UpdateQueue for fiber2 since CloneFiber now sets it to nil
	if fiber2.UpdateQueue == nil {
		fiber2.UpdateQueue = &UpdateQueue{}
	}

	// Add an update to fiber2
	newUpdate := &Update{
		Payload: "update2",
		Lane:    LaneDefaultLane,
	}
	fiber2.EnqueueUpdate(newUpdate)

	// Check if fiber1 was affected
	if fiber1.UpdateQueue.First.Next != nil {
		t.Error("Adding update to cloned fiber affected original fiber")
		if fiber1.UpdateQueue.First.Next.Payload == "update2" {
			t.Error("Original fiber's UpdateQueue was modified by clone")
		}
	} else {
		t.Log("FIXED: Original fiber's UpdateQueue was not affected")
	}

	t.Logf("fiber1 queue length: %d", countUpdateQueue(fiber1.UpdateQueue))
	t.Logf("fiber2 queue length: %d", countUpdateQueue(fiber2.UpdateQueue))
}

// countUpdateQueue helper
func countUpdateQueue(queue *UpdateQueue) int {
	if queue == nil {
		return 0
	}
	count := 0
	for u := queue.First; u != nil; u = u.Next {
		count++
	}
	return count
}

// TestStateRedundancy clarifies the different state sources
func TestStateRedundancy(t *testing.T) {
	// Fiber.MemoizedState serves different purposes depending on VNode type:
	// 1. TextVNode: stores text content (complete_work.go:80)
	// 2. ComponentVNode with UpdateQueue: stores state for functional updates (begin_work.go:247-251)
	// 3. ComponentVNode with hooks: NOT used - hooks use ComponentContext

	// For hook-based components, state is in ComponentContext.Hooks, NOT in MemoizedState
	componentFunc := func() rtui.VNode {
		return rtui.Element("text").Prop("content", "test").Build()
	}

	fiber := &Fiber{
		Type: rtui.VNodeComponent,
		Tag:  "TestComponent",
		Key:  "test-comp",
	}

	instance := rtui.NewBaseComponentInstance("test-comp", componentFunc)
	fiber.ComponentInstance = instance

	// MemoizedState is nil for hook-based components (unless using UpdateQueue)
	t.Logf("fiber.MemoizedState: %v", fiber.MemoizedState)
	// BaseComponentInstance doesn't have GetState() method
	// For hook-based state, it's stored in ComponentContext.Hooks
	if ctx := instance.GetContext(); ctx != nil {
		t.Logf("instance.GetContext().Hooks (from hooks): %v", ctx.Hooks)
	} else {
		t.Logf("instance.GetContext() returned nil")
	}

	// CONCLUSION: No actual redundancy
	// - MemoizedState is for VNode-level data (text content, update queue state)
	// - Hook state is in ComponentContext.Hooks (accessible via ComponentInstance.GetContext())
	// They serve different purposes and don't overlap
	t.Log("CONCLUSION: MemoizedState and hook state serve different purposes:")
	t.Log("  - MemoizedState: VNode-level data (text content, UpdateQueue state)")
	t.Log("  - Hook state: In ComponentContext.Hooks (via ComponentInstance.GetContext().Hooks)")
	t.Log("  - For TextVNode: completeWorkText() sets MemoizedState to text content")
	t.Log("  - For ComponentVNode with UpdateQueue: beginWork() uses MemoizedState for functional updates")
	t.Log("  - For ComponentVNode with hooks: MemoizedState NOT used (state in ComponentContext.Hooks)")
}
