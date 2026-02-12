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

	// Verify VNode.Key() now contains the full path (for Inspector)
	if childVNode.Key() != expectedPath {
		t.Errorf("VNode.Key() = %q, expected %q", childVNode.Key(), expectedPath)
	}
}

// TestUserKeyPathReuse verifies that cloned fibers preserve paths correctly
func TestUserKeyPathReuse(t *testing.T) {
	// Initialize path generator
	pathGenerator = NewPathGenerator()

	// Create existing fiber with user key path
	existingVNode := rtui.NewElement("button")
	existingVNode.SetKey("/root/base[0]/button[0]/key[btn-event]")

	currentFiber := &Fiber{
		Tag:     "button",
		Key:     "btn-event",
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

	// Verify Fiber.Path is preserved
	expectedPath := "/root/base[0]/button[0]/key[btn-event]"
	if clonedFiber.Path != expectedPath {
		t.Errorf("Cloned Fiber.Path = %q, expected %q", clonedFiber.Path, expectedPath)
	}

	// Verify new VNode.Key() is synced with full path
	if newVNode.Key() != expectedPath {
		t.Errorf("New VNode.Key() = %q, expected %q", newVNode.Key(), expectedPath)
	}
}
