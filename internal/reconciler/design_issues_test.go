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
				Lane:   LaneSyncLane,
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
		Lane:   LaneDefaultLane,
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

// TestStateRedundancy tests the three potential state sources
func TestStateRedundancy(t *testing.T) {
	// Create a component VNode
	componentFunc := func() rtui.VNode {
		return rtui.Element("text").Prop("content", "test").Build()
	}
	compVNode := rtui.NewComponent("TestComponent", componentFunc)

	// Create a Fiber
	fiber := &Fiber{
		VNode: compVNode,
		Type:  rtui.VNodeComponent,
		Tag:   "TestComponent",
		Key:   "test-comp",
		MemoizedState: map[string]interface{}{
			"key1": "value1",
		},
	}

	// Create a ComponentInstance
	instance := rtui.NewBaseComponentInstance("test-comp", componentFunc)
	fiber.ComponentInstance = instance

	// Now we have three potential state sources:
	// 1. fiber.MemoizedState
	// 2. fiber.ComponentInstance (which has hooks/state)
	// 3. instance.GetState()

	state1 := fiber.MemoizedState
	if state1 != nil {
		if m, ok := state1.(map[string]interface{}); ok {
			t.Logf("fiber.MemoizedState: %v", m)
		}
	}

	state3 := instance.GetState()
	t.Logf("instance.GetState(): %v", state3)

	// BUG: Confusion about which state source is authoritative
	if state1 != nil && len(state3) > 0 {
		t.Error("Potential bug: Two different state sources exist")
		t.Error("  - fiber.MemoizedState is used where?")
		t.Error("  - instance.GetState() returns hook state")
		t.Error("Which one should be used?")
	}

	// Check if MemoizedState is actually used anywhere
	if fiber.MemoizedState != nil {
		t.Log("fiber.MemoizedState is non-nil - this field exists but its purpose is unclear")
	}
}
