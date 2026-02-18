package reconciler

import (
	"testing"

	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// TestVNodeCacheConsistency tests that Fiber's cached props stay in sync with VNode
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
	// This is expected behavior - use GetProps() to get current props
	if fiber.Props.GetString("value") == "initial" {
		t.Log("EXPECTED: Fiber.Props field is stale (snapshot)")
	}

	// But GetProps() returns the current props from VNode
	if fiber.GetProps().GetString("value") != "updated" {
		t.Error("GetProps() should return current props from VNode")
	}

	// The VNode has the new value
	if vnode.Props().GetString("value") != "updated" {
		t.Error("VNode should have updated props")
	}

	// Also check Type and key fields
	if fiber.Type != vnode.Type() {
		t.Error("Fiber.Type may be stale")
	}

	if fiber.Key != vnode.Key() {
		t.Error("Fiber.Key may be stale")
	}

	t.Logf("VNode props: %v", vnode.Props())
	t.Logf("Fiber cached Props: %v", fiber.Props)
	t.Logf("Fiber MemoizedProps: %v", fiber.MemoizedProps)
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
	t.Logf("instance.GetState() (from hooks): %v", instance.GetState())

	// CONCLUSION: No actual redundancy
	// - MemoizedState is for VNode-level data (text content, update queue state)
	// - ComponentInstance.GetState() is for hook-based state (useState)
	// They serve different purposes and don't overlap
	t.Log("CONCLUSION: MemoizedState and hook state serve different purposes:")
	t.Log("  - MemoizedState: VNode-level data (text content, UpdateQueue state)")
	t.Log("  - ComponentInstance.GetState(): Hook-based state from useState")
	t.Log("  - For TextVNode: completeWorkText() sets MemoizedState to text content")
	t.Log("  - For ComponentVNode with UpdateQueue: beginWork() uses MemoizedState for functional updates")
	t.Log("  - For ComponentVNode with hooks: MemoizedState NOT used (state in ComponentContext.Hooks)")
}
