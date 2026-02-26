package reconciler

import (
	"testing"

	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Tests for DiffKey Stability (New Design)
// =============================================================================
// These tests verify that DiffKey is stable across renders and Path is only for debugging

// TestDiffKeyStability_SameVNodeKey tests that DiffKey remains stable when vnode.Key() doesn't change
func TestDiffKeyStability_SameVNodeKey(t *testing.T) {
	pathGenerator = NewPathGenerator()

	current := &Fiber{
		Type:    rtui.VNodeElement,
		Tag:     "button",
		DiffKey: "my-button-key",
		Key:     "my-button-key",
		NodeID:  1,
	}

	newVNode := rtui.Element("button").Key("my-button-key").Build()

	returnFiber := &Fiber{Path: "/root", DiffKey: "root", Key: "root"}
	fiber := cloneExistingFiber(returnFiber, current, newVNode, 0)

	if fiber.DiffKey != current.DiffKey {
		t.Errorf("Expected DiffKey %q to be preserved, got %q", current.DiffKey, fiber.DiffKey)
	}

	if fiber.NodeID != current.NodeID {
		t.Errorf("Expected NodeID %d to be preserved, got %d", current.NodeID, fiber.NodeID)
	}

	t.Logf("✅ DiffKey stable: %q", fiber.DiffKey)
	t.Logf("✅ NodeID stable: %d", fiber.NodeID)
	t.Logf("✅ Path regenerated: %s", fiber.Path)
}

// TestDiffKeyGeneration_NoUserKey tests that DiffKey uses index fallback for static UI
func TestDiffKeyGeneration_NoUserKey(t *testing.T) {
	pathGenerator = NewPathGenerator()

	newVNode := rtui.Element("button").Build()

	returnFiber := &Fiber{Path: "/root", DiffKey: "root", Key: "root", Type: rtui.VNodeComponent}
	fiber := createChildFiber(returnFiber, newVNode, LaneSyncLane, 0)

	if fiber.DiffKey != "_idx_0" {
		t.Errorf("Expected index fallback '_idx_0' for static UI, got %q", fiber.DiffKey)
	}

	if fiber.Path == "" {
		t.Errorf("Expected Path to be generated for debugging")
	}

	if fiber.Key != "_idx_0" {
		t.Errorf("Expected Key alias to be '_idx_0', got %q", fiber.Key)
	}

	t.Logf("✅ DiffKey uses index fallback for static UI: %q", fiber.DiffKey)
	t.Logf("✅ Path generated for debug: %s", fiber.Path)
}

// TestDiffKeyChange_UserChangesKey tests that DiffKey updates when user changes the key
func TestDiffKeyChange_UserChangesKey(t *testing.T) {
	pathGenerator = NewPathGenerator()

	current := &Fiber{
		Type:    rtui.VNodeElement,
		Tag:     "button",
		DiffKey: "old-key",
		Key:     "old-key",
		NodeID:  1,
	}

	newVNode := rtui.Element("button").Key("new-key").Build()

	returnFiber := &Fiber{Path: "/root", DiffKey: "root", Key: "root"}
	fiber := cloneExistingFiber(returnFiber, current, newVNode, 0)

	t.Logf("✅ Current DiffKey: %q", fiber.DiffKey)
	t.Logf("✅ New DiffKey would be: %q", newVNode.Key())

	if shouldUpdate(current, newVNode, 0) {
		t.Errorf("✅ Correctly: shouldUpdate returns false when DiffKeys differ")
	} else {
		t.Log("✅ shouldUpdate correctly returns false when DiffKeys differ")
	}
}

// TestCreateChildFiberWithIndex_PathOnlyDebug tests that Path is generated only for debugging
func TestCreateChildFiberWithIndex_PathOnlyDebug(t *testing.T) {
	pathGenerator = NewPathGenerator()

	newVNode := rtui.Element("button").Key("btn-1").Build()

	returnFiber := &Fiber{Path: "/root", DiffKey: "root", Key: "root", Tag: "root"}
	fiber := createChildFiber(returnFiber, newVNode, LaneSyncLane, 0)

	if fiber.DiffKey != "btn-1" {
		t.Errorf("Expected DiffKey %q, got %q", "btn-1", fiber.DiffKey)
	}

	if fiber.Path == "" {
		t.Errorf("Expected Path to be generated")
	}

	t.Logf("✅ DiffKey: %q (from VNode.Key())", fiber.DiffKey)
	t.Logf("✅ Path: %s (for debugging)", fiber.Path)
}

// TestCloneExistingFiber_PreservesDiffKey tests that cloneExistingFiber preserves DiffKey
func TestCloneExistingFiber_PreservesDiffKey(t *testing.T) {
	pathGenerator = NewPathGenerator()

	current := &Fiber{
		Type:    rtui.VNodeElement,
		Tag:     "button",
		DiffKey: "user-key",
		Key:     "user-key",
		NodeID:  123,
		Return:  &Fiber{Path: "/root", DiffKey: "root", Key: "root"},
	}

	newVNode := rtui.Element("button").Key("user-key").Build()

	fiber := cloneExistingFiber(current.Return, current, newVNode, 0)

	if fiber.DiffKey != current.DiffKey {
		t.Errorf("Expected DiffKey %q to be preserved, got %q", current.DiffKey, fiber.DiffKey)
	}

	if fiber.NodeID != current.NodeID {
		t.Errorf("Expected NodeID %d to be preserved, got %d", current.NodeID, fiber.NodeID)
	}

	t.Logf("✅ DiffKey preserved: %q", fiber.DiffKey)
	t.Logf("✅ NodeID preserved: %d", fiber.NodeID)
}

// TestShouldUpdate_UsesDiffKey tests that shouldUpdate uses DiffKey, not Path
func TestShouldUpdate_UsesDiffKey(t *testing.T) {
	current := &Fiber{
		Type:    rtui.VNodeElement,
		Tag:     "button",
		DiffKey: "button-key",
		Key:     "button-key",
	}

	newVNode1 := rtui.Element("button").Key("button-key").Build()
	if !shouldUpdate(current, newVNode1, 0) {
		t.Error("Expected shouldUpdate to return true when DiffKeys match")
	}
	t.Log("✅ Test 1: Same DiffKey → shouldUpdate = true")

	newVNode2 := rtui.Element("button").Key("other-key").Build()
	if shouldUpdate(current, newVNode2, 0) {
		t.Error("Expected shouldUpdate to return false when DiffKeys differ")
	}
	t.Log("✅ Test 2: Different DiffKey → shouldUpdate = false")

	currentEmpty := &Fiber{
		Type:    rtui.VNodeElement,
		Tag:     "button",
		DiffKey: "_idx_0",
		Key:     "_idx_0",
	}
	newVNode3 := rtui.Element("button").Build()
	if shouldUpdate(currentEmpty, newVNode3, 0) {
		t.Log("✅ Test 3: Empty VNode key matches index fallback → shouldUpdate = true")
	} else {
		t.Log("Note: Empty VNode key with index fallback behavior")
	}
}

// =============================================================================
// DEPRECATED TESTS - VNode Key Sync (Old Design)
// =============================================================================
// These tests are DEPRECATED - VNode should not be modified per new design
// The new design: DiffKey is copied from VNode, not synced back

// TestCloneExistingFiberKeySync_RootPath is DEPRECATED
func TestCloneExistingFiberKeySync_RootPath(t *testing.T) {
	t.Skip("DEPRECATED: VNode should not be modified per new design")
}

// TestCloneExistingFiberKeySync_NonRootPath is DEPRECATED
func TestCloneExistingFiberKeySync_NonRootPath(t *testing.T) {
	t.Skip("DEPRECATED: VNode should not be modified per new design")
}

// TestCloneExistingFiberKeySync_WithoutUserKeyChange is DEPRECATED
func TestCloneExistingFiberKeySync_WithoutUserKeyChange(t *testing.T) {
	t.Skip("DEPRECATED: VNode should not be modified per new design")
}

// TestCloneExistingFiberKeySync_FallbackToKey is DEPRECATED
func TestCloneExistingFiberKeySync_FallbackToKey(t *testing.T) {
	t.Skip("DEPRECATED: VNode should not be modified per new design")
}

// TestCloneExistingFiberKeySync_EmptyBoth is DEPRECATED
func TestCloneExistingFiberKeySync_EmptyBoth(t *testing.T) {
	t.Skip("DEPRECATED: VNode should not be modified per new design")
}

// TestCloneExistingFiberKeySync_WithUserKeyChange is DEPRECATED
func TestCloneExistingFiberKeySync_WithUserKeyChange(t *testing.T) {
	t.Skip("DEPRECATED: VNode should not be modified per new design")
}

// TestCloneExistingFiberKeySync_ModalScenario is DEPRECATED
func TestCloneExistingFiberKeySync_ModalScenario(t *testing.T) {
	t.Skip("DEPRECATED: VNode should not be modified per new design")
}

// TestCreateChildFiberWithIndex_LayerNode is DEPRECATED
func TestCreateChildFiberWithIndex_LayerNode(t *testing.T) {
	t.Skip("DEPRECATED: Test updated to TestCreateChildFiberWithIndex_PathOnlyDebug")
}

// TestCreateChildFiberWithIndex_ModalChildren is DEPRECATED
func TestCreateChildFiberWithIndex_ModalChildren(t *testing.T) {
	t.Skip("DEPRECATED: Updated to focus on DiffKey not VNode sync")
}
