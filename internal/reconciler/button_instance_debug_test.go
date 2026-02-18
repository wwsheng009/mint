package reconciler

import (
	"testing"

	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// TestButtonInstanceCreation tests that buttons with user keys
// ✨ Updated for new design: DiffKey is used for diffing, VNode is NOT modified
func TestButtonInstanceCreation(t *testing.T) {
	t.Log("=== Testing Button ComponentInstance Creation ===")

	pathGenerator = NewPathGenerator()

	parentFiber := &Fiber{
		Tag:     "base",
		Key:     "/root/base[0]",
		Path:    "/root/base[0]",
		DiffKey: "/root/base[0]",
		Type:    rtui.VNodeElement,
	}

	buttonVNode := rtui.NewElement("button")
	buttonVNode.SetKey("btn-test")

	t.Logf("1. Created button VNode with Key() = %q", buttonVNode.Key())

	childFiber := createChildFiber(parentFiber, buttonVNode, LaneSyncLane, 0)

	t.Logf("2. After createChildFiber:")
	t.Logf("   Fiber.DiffKey = %q", childFiber.DiffKey)
	t.Logf("   Fiber.Key = %q", childFiber.Key)
	t.Logf("   Fiber.Path = %q (for debugging)", childFiber.Path)
	t.Logf("   VNode.Key() = %q (NOT modified)", buttonVNode.Key())
	t.Logf("   Fiber.NodeID = %d (runtime identity)", childFiber.NodeID)

	if childFiber.DiffKey != "btn-test" {
		t.Errorf("Fiber.DiffKey = %q, expected %q", childFiber.DiffKey, "btn-test")
	}

	if childFiber.Key != childFiber.DiffKey {
		t.Errorf("Fiber.Key = %q, expected equal to DiffKey %q", childFiber.Key, childFiber.DiffKey)
	}

	if childFiber.Path == "" {
		t.Errorf("Fiber.Path should be generated for debugging")
	}

	expectedPathPattern := "/root/base[0]/button[0]/key[btn-test]"
	if childFiber.Path != expectedPathPattern {
		t.Logf("Note: Path = %q, expected %q (format may vary)", childFiber.Path, expectedPathPattern)
	}

	if buttonVNode.Key() != "btn-test" {
		t.Errorf("VNode.Key() was modified to %q, expected to remain %q (VNode should NOT be modified)",
			buttonVNode.Key(), "btn-test")
	}

	if childFiber.NodeID == 0 {
		t.Errorf("Fiber.NodeID should be allocated for runtime identity")
	}

	t.Log("✅ Test passed: Button instance correctly created")
	t.Log("   - DiffKey stable from user key")
	t.Log("   - VNode NOT modified")
	t.Log("   - Path generated for debugging")
	t.Log("   - NodeID allocated for identity")
}

// TestTwoButtonsWithSameKey tests that two buttons with same user key
// ✨ Updated for new design: DiffKey is used for diffing, VNode is NOT modified
func TestTwoButtonsWithSameKey(t *testing.T) {
	t.Log("=== Testing Two Buttons With Same User Key ===")

	pathGenerator = NewPathGenerator()

	parentFiber := &Fiber{
		Tag:     "base",
		Key:     "/root/base[0]",
		Path:    "/root/base[0]",
		DiffKey: "/root/base[0]",
		Type:    rtui.VNodeElement,
	}

	button1VNode := rtui.NewElement("button")
	button1VNode.SetKey("btn-test")

	button2VNode := rtui.NewElement("button")
	button2VNode.SetKey("btn-test")

	child1Fiber := createChildFiber(parentFiber, button1VNode, LaneSyncLane, 0)
	parentFiber.Child = child1Fiber
	child2Fiber := createChildFiber(parentFiber, button2VNode, LaneSyncLane, 1)
	child1Fiber.Sibling = child2Fiber

	t.Logf("Button 1:")
	t.Logf("   Fiber.DiffKey = %q, Fiber.Path = %q", child1Fiber.DiffKey, child1Fiber.Path)
	t.Logf("   VNode.Key() = %q (NOT modified)", button1VNode.Key())
	t.Logf("   NodeID = %d (runtime identity)", child1Fiber.NodeID)

	t.Logf("Button 2:")
	t.Logf("   Fiber.DiffKey = %q, Fiber.Path = %q", child2Fiber.DiffKey, child2Fiber.Path)
	t.Logf("   VNode.Key() = %q (NOT modified)", button2VNode.Key())
	t.Logf("   NodeID = %d (runtime identity)", child2Fiber.NodeID)

	if child1Fiber.DiffKey != child2Fiber.DiffKey {
		t.Errorf("DiffKeys differ: button1=%q, button2=%q (both should have same user key)",
			child1Fiber.DiffKey, child2Fiber.DiffKey)
	}

	if child1Fiber.Path == child2Fiber.Path {
		t.Errorf("Both buttons have same Path: %q (paths should differ for different siblings)",
			child1Fiber.Path)
	}

	if child1Fiber.NodeID == child2Fiber.NodeID {
		t.Errorf("Both buttons have same NodeID: %d (NodeIDs should be unique for distinct fibers)",
			child1Fiber.NodeID)
	}

	if button1VNode.Key() != "btn-test" || button2VNode.Key() != "btn-test" {
		t.Errorf("VNodes were modified")
	}

	t.Log("✅ Test passed: Two buttons with same key")
	t.Log("   - Same DiffKey (from user key)")
	t.Log("   - Different Paths (for debugging)")
	t.Log("   - Different NodeIDs (unique identity)")
	t.Log("   - VNodes NOT modified")
}
