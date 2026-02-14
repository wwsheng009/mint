package reconciler

import (
	"testing"

	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// TestUserKeyPathGeneration verifies that user keys get full paths
func TestUserKeyPathGeneration(t *testing.T) {
	// Initialize path generator
	pathGenerator = NewPathGenerator()

	// Create parent fiber (simulating /root/base[0])
	parentFiber := &Fiber{
		Tag:  "base",
		Key:  "/root/base[0]",
		Path: "/root/base[0]",
		Type: rtui.VNodeElement,
		VNode: rtui.NewElement("base"),
	}

	// Create a child VNode with user key
	childVNode := rtui.NewElement("button")
	childVNode.SetKey("btn-event")

	// Create child fiber (this should generate full path with user key)
	childFiber := createChildFiber(parentFiber, childVNode, rtui.LaneSyncLane, 0)

	// Verify Fiber.Key is the original user key (for reconciliation)
	if childFiber.Key != "btn-event" {
		t.Errorf("Fiber.Key = %q, expected %q", childFiber.Key, "btn-event")
	}

	// Verify Fiber.Path contains full path with type + user key
	// Format: /root/base[0]/button[0]/key[btn-event]
	expectedPath := "/root/base[0]/button[0]/key[btn-event]"
	if childFiber.Path != expectedPath {
		t.Errorf("Fiber.Path = %q, expected %q", childFiber.Path, expectedPath)
	}

	// ✨ New Design: VNode is NOT modified
	// VNode.Key() should remain the original user key, not the full path
	if childVNode.Key() != "btn-event" {
		t.Errorf("VNode.Key() = %q, expected %q (VNode should NOT be modified)", childVNode.Key(), "btn-event")
	}
}

// TestUserKeyPathReuse verifies that cloned fibers preserve paths correctly
func TestUserKeyPathReuse(t *testing.T) {
	// Initialize path generator
	pathGenerator = NewPathGenerator()

	// Create existing fiber - VNode has user key (not full path)
	existingVNode := rtui.NewElement("button")
	existingVNode.SetKey("btn-event") // User key only

	currentFiber := &Fiber{
		Tag:     "button",
		Key:     "btn-event",
		DiffKey: "btn-event", // ✨ New Design: DiffKey is preserved
		Path:    "/root/base[0]/button[0]/key[btn-event]",
		Type:    rtui.VNodeElement,
		VNode:   existingVNode,
		Props:   rtui.Props{},
	}

	// Create parent fiber
	parentFiber := &Fiber{
		Tag:  "base",
		Key:  "/root/base[0]",
		Path: "/root/base[0]",
		Type: rtui.VNodeElement,
		VNode: rtui.NewElement("base"),
	}

	// Create new VNode with same user key
	newVNode := rtui.NewElement("button")
	newVNode.SetKey("btn-event")

	// Clone the fiber
	clonedFiber := cloneExistingFiber(parentFiber, currentFiber, newVNode, 0)

	// Verify Fiber.Key is preserved (user key)
	if clonedFiber.Key != "btn-event" {
		t.Errorf("Cloned Fiber.Key = %q, expected %q", clonedFiber.Key, "btn-event")
	}

	// Verify ✨ New Design: DiffKey is preserved across renders
	if clonedFiber.DiffKey != "btn-event" {
		t.Errorf("Cloned Fiber.DiffKey = %q, expected %q", clonedFiber.DiffKey, "btn-event")
	}

	// Verify Fiber.Path is regenerated for debugging
	expectedPath := "/root/base[0]/button[0]/key[btn-event]"
	if clonedFiber.Path != expectedPath {
		t.Errorf("Cloned Fiber.Path = %q, expected %q", clonedFiber.Path, expectedPath)
	}

	// ✨ New Design: VNodes are NEVER modified
	// Previous VNode key should remain as user key
	if existingVNode.Key() != "btn-event" {
		t.Errorf("Existing VNode.Key() = %q, expected %q (VNode should NOT be modified)", existingVNode.Key(), "btn-event")
	}
	// New VNode key should also remain as user key
	if newVNode.Key() != "btn-event" {
		t.Errorf("New VNode.Key() = %q, expected %q (VNode should NOT be modified)", newVNode.Key(), "btn-event")
	}
}
