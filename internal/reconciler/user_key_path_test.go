package reconciler

import (
	"testing"

	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// TestUserKeyPathGeneration verifies that user keys get full paths
func TestUserKeyPathGeneration(t *testing.T) {
	pathGenerator = NewPathGenerator()

	parentFiber := &Fiber{
		Tag:     "base",
		Key:     "/root/base[0]",
		Path:    "/root/base[0]",
		DiffKey: "/root/base[0]",
		Type:    rtui.VNodeElement,
	}

	childVNode := rtui.NewElement("button")
	childVNode.SetKey("btn-event")

	childFiber := createChildFiber(parentFiber, childVNode, rtui.LaneSyncLane, 0)

	if childFiber.Key != "btn-event" {
		t.Errorf("Fiber.Key = %q, expected %q", childFiber.Key, "btn-event")
	}

	expectedPath := "/root/base[0]/button[0]/key[btn-event]"
	if childFiber.Path != expectedPath {
		t.Errorf("Fiber.Path = %q, expected %q", childFiber.Path, expectedPath)
	}

	if childVNode.Key() != "btn-event" {
		t.Errorf("VNode.Key() = %q, expected %q (VNode should NOT be modified)", childVNode.Key(), "btn-event")
	}
}

// TestUserKeyPathReuse verifies that cloned fibers preserve paths correctly
func TestUserKeyPathReuse(t *testing.T) {
	pathGenerator = NewPathGenerator()

	currentFiber := &Fiber{
		Tag:     "button",
		Key:     "btn-event",
		DiffKey: "btn-event",
		Path:    "/root/base[0]/button[0]/key[btn-event]",
		Type:    rtui.VNodeElement,
		Props:   rtui.Props{},
	}

	parentFiber := &Fiber{
		Tag:     "base",
		Key:     "/root/base[0]",
		Path:    "/root/base[0]",
		DiffKey: "/root/base[0]",
		Type:    rtui.VNodeElement,
	}

	newVNode := rtui.NewElement("button")
	newVNode.SetKey("btn-event")

	clonedFiber := cloneExistingFiber(parentFiber, currentFiber, newVNode, 0)

	if clonedFiber.Key != "btn-event" {
		t.Errorf("Cloned Fiber.Key = %q, expected %q", clonedFiber.Key, "btn-event")
	}

	if clonedFiber.DiffKey != "btn-event" {
		t.Errorf("Cloned Fiber.DiffKey = %q, expected %q", clonedFiber.DiffKey, "btn-event")
	}

	expectedPath := "/root/base[0]/button[0]/key[btn-event]"
	if clonedFiber.Path != expectedPath {
		t.Errorf("Cloned Fiber.Path = %q, expected %q", clonedFiber.Path, expectedPath)
	}

	if newVNode.Key() != "btn-event" {
		t.Errorf("New VNode.Key() = %q, expected %q (VNode should NOT be modified)", newVNode.Key(), "btn-event")
	}
}
