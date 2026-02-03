package reconciler

import (
	"testing"

	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// TestEnqueueUpdateDeepCopy tests if EnqueueUpdate creates new queue or modifies existing
func TestEnqueueUpdateDeepCopy(t *testing.T) {
	// Create fiber with existing update queue
	fiber1 := &Fiber{
		Key:  "fiber1",
		Type: rtui.VNodeText,
		UpdateQueue: &UpdateQueue{
			First: &Update{
				Payload: "update1",
				Lane:   LaneSyncLane,
			},
			Last: nil, // Initially same as First
		},
	}
	fiber1.UpdateQueue.Last = fiber1.UpdateQueue.First

	// Clone the fiber
	fiber2 := CloneFiber(fiber1)

	// Check if they share the same queue
	t.Logf("fiber1.UpdateQueue == fiber2.UpdateQueue: %v", fiber1.UpdateQueue == fiber2.UpdateQueue)
	t.Logf("fiber1 queue length before: %d", countUpdateQueueDetail(fiber1.UpdateQueue))
	t.Logf("fiber2 queue length before: %d", countUpdateQueueDetail(fiber2.UpdateQueue))

	// Add update to fiber2
	newUpdate := &Update{
		Payload: "update2",
		Lane:   LaneDefaultLane,
	}

	t.Logf("Before EnqueueUpdate:")
	t.Logf("  fiber1.UpdateQueue.First = %v", fiber1.UpdateQueue.First)
	t.Logf("  fiber1.UpdateQueue.Last = %v", fiber1.UpdateQueue.Last)
	t.Logf("  fiber1.UpdateQueue.First.Next = %v", fiber1.UpdateQueue.First.Next)

	fiber2.EnqueueUpdate(newUpdate)

	t.Logf("After EnqueueUpdate:")
	t.Logf("  fiber1.UpdateQueue.First = %v", fiber1.UpdateQueue.First)
	t.Logf("  fiber1.UpdateQueue.Last = %v", fiber1.UpdateQueue.Last)
	t.Logf("  fiber1.UpdateQueue.First.Next = %v", fiber1.UpdateQueue.First.Next)
	t.Logf("  fiber2.UpdateQueue.Last = fiber1.UpdateQueue.Last: %v", fiber2.UpdateQueue.Last == fiber1.UpdateQueue.Last)

	t.Logf("fiber1 queue length after: %d", countUpdateQueueDetail(fiber1.UpdateQueue))
	t.Logf("fiber2 queue length after: %d", countUpdateQueueDetail(fiber2.UpdateQueue))

	// Check if fiber2 created its own queue
	if fiber1.UpdateQueue != fiber2.UpdateQueue {
		t.Log("fiber2 created its own UpdateQueue")
	}

	// Check if fiber1.Queue was modified
	if fiber1.UpdateQueue.First.Next != nil {
		t.Logf("fiber1.UpdateQueue.First.Next = %v", fiber1.UpdateQueue.First.Next)
		if fiber1.UpdateQueue.First.Next.Payload == "update2" {
			t.Log("fiber1 was affected! (shared queue)")
		}
	}
}

// countUpdateQueueDetail counts queue and logs details
func countUpdateQueueDetail(queue *UpdateQueue) int {
	if queue == nil {
		return 0
	}
	count := 0
	for u := queue.First; u != nil; u = u.Next {
		count++
	}
	return count
}
