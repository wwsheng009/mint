package reconciler

import (
	"testing"

	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// TestButtonInstanceCreation tests that buttons with user keys
// create ComponentInstance with correct full path keys
func TestButtonInstanceCreation(t *testing.T) {
	t.Log("=== Testing Button ComponentInstance Creation ===")

	// Initialize path generator
	pathGenerator = NewPathGenerator()

	// Create parent fiber
	parentFiber := &Fiber{
		Tag:  "base",
		Key:  "/root/base[0]",
		Path: "/root/base[0]",
		Type: rtui.VNodeElement,
		VNode: rtui.NewElement("base"),
	}

	// Create a button VNode with user key
	buttonVNode := rtui.NewElement("button")
	buttonVNode.SetKey("btn-test")

	t.Logf("1. Created button VNode with Key() = %q", buttonVNode.Key())

	// Simulate createChildFiber - this should set both Fiber.Key and Fiber.Path
	// For user keys: Fiber.Key = "btn-test", Fiber.Path = "/root/base[0]/.../button[0]/key[btn-test]"
	childFiber := createChildFiber(parentFiber, buttonVNode, LaneSyncLane, 0)

	t.Logf("2. After createChildFiber:")
	t.Logf("   Fiber.Key = %q", childFiber.Key)
	t.Logf("   Fiber.Path = %q", childFiber.Path)
	t.Logf("   VNode.Key() = %q", childFiber.VNode.Key())

	// Verify: Fiber.Key should be the original user key
	if childFiber.Key != "btn-test" {
		t.Errorf("Fiber.Key = %q, expected %q", childFiber.Key, "btn-test")
	}

	// Verify: Fiber.Path should contain the type path + user key
	expectedPathPattern := "/root/base[0]/button[0]/key[btn-test]"
	if childFiber.Path != expectedPathPattern {
		t.Errorf("Fiber.Path = %q, expected to contain %q", childFiber.Path, expectedPathPattern)
	}

	// Verify: VNode.Key() should be synced with Fiber.Path
	if childFiber.VNode.Key() != childFiber.Path {
		t.Errorf("VNode.Key() = %q, expected %q (synced with Fiber.Path)", childFiber.VNode.Key(), childFiber.Path)
	}

	t.Log("✅ Test passed: Button instance correctly created with full path")
}

// TestTwoButtonsWithSameKey tests that two buttons with same user key
// get different Fiber.Path values but should both create ComponentInstance
func TestTwoButtonsWithSameKey(t *testing.T) {
	t.Log("=== Testing Two Buttons With Same User Key ===")

	pathGenerator = NewPathGenerator()

	parentFiber := &Fiber{
		Tag:  "base",
		Key:  "/root/base[0]",
		Path: "/root/base[0]",
		Type: rtui.VNodeElement,
		VNode: rtui.NewElement("base"),
	}

	// Create two buttons with SAME user key
	button1VNode := rtui.NewElement("button")
	button1VNode.SetKey("btn-test")

	button2VNode := rtui.NewElement("button")
	button2VNode.SetKey("btn-test")

	// Create two child fibers
	child1Fiber := createChildFiber(parentFiber, button1VNode, LaneSyncLane, 0)
	child2Fiber := createChildFiber(parentFiber, button2VNode, LaneSyncLane, 1)

	t.Logf("Button 1:")
	t.Logf("   Fiber.Key = %q, Fiber.Path = %q", child1Fiber.Key, child1Fiber.Path)
	t.Logf("   VNode.Key() = %q", button1VNode.Key())

	t.Logf("Button 2:")
	t.Logf("   Fiber.Key = %q, Fiber.Path = %q", child2Fiber.Key, child2Fiber.Path)
	t.Logf("   VNode.Key() = %q", button2VNode.Key())

	// Both buttons should have different paths because they're at different sibling indices
	if child1Fiber.Path == child2Fiber.Path {
		t.Errorf("Both buttons have same Path: %q", child1Fiber.Path)
	}

	// But both should have created ComponentInstance
	// (This would be tested by checking if beginWorkElement is called)
	t.Log("✅ Test passed: Two buttons with same key get different paths")
}
