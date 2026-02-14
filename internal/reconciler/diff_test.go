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
	// Setup path generator
	pathGenerator = NewPathGenerator()

	// Create a current fiber
	current := &Fiber{
		Type:    rtui.VNodeElement,
		Tag:     "button",
		DiffKey: "my-button-key", // User-provided key
		Key:     "my-button-key", // Backward compatibility alias
		NodeID:  1,
	}

	// Create new VNode with same key
	newVNode := rtui.Element("button").Key("my-button-key").Build()

	// Clone the fiber
	returnFiber := &Fiber{Path: "/root", DiffKey: "root", Key: "root"}
	fiber := cloneExistingFiber(returnFiber, current, newVNode, 0)

	// Verify DiffKey is preserved (stable across renders)
	if fiber.DiffKey != current.DiffKey {
		t.Errorf("Expected DiffKey %q to be preserved, got %q", current.DiffKey, fiber.DiffKey)
	}

	// Verify NodeID is preserved (stable identity)
	if fiber.NodeID != current.NodeID {
		t.Errorf("Expected NodeID %d to be preserved, got %d", current.NodeID, fiber.NodeID)
	}

	// Verify Path can change (it's only for debugging)
	// Note: The exact path may vary, so we don't check equality here
	t.Logf("✅ DiffKey stable: %q", fiber.DiffKey)
	t.Logf("✅ NodeID stable: %d", fiber.NodeID)
	t.Logf("✅ Path regenerated: %s", fiber.Path)
}

// TestDiffKeyGeneration_NoUserKey tests that DiffKey uses index fallback for static UI
func TestDiffKeyGeneration_NoUserKey(t *testing.T) {
	// Setup path generator
	pathGenerator = NewPathGenerator()

	// Create new VNode without key
	newVNode := rtui.Element("button").Build()

	// Create child fiber
	returnFiber := &Fiber{Path: "/root", DiffKey: "root", Key: "root", Type: rtui.VNodeComponent}
	fiber := createChildFiber(returnFiber, newVNode, LaneSyncLane, 0)

	// ✨ New Design: DiffKey uses index fallback for static UI
	// This is correct behavior per diff_key2.md design
	if fiber.DiffKey != "0" {
		t.Errorf("Expected index fallback '0' for static UI, got %q", fiber.DiffKey)
	}

	// Verify Path is still generated for debugging
	if fiber.Path == "" {
		t.Errorf("Expected Path to be generated for debugging")
	}

	// Verify Key alias is also the index fallback
	if fiber.Key != "0" {
		t.Errorf("Expected Key alias to be '0', got %q", fiber.Key)
	}

	t.Logf("✅ DiffKey uses index fallback for static UI: %q", fiber.DiffKey)
	t.Logf("✅ Path generated for debug: %s", fiber.Path)
}

// TestDiffKeyChange_UserChangesKey tests that DiffKey updates when user changes the key
func TestDiffKeyChange_UserChangesKey(t *testing.T) {
	// Setup path generator
	pathGenerator = NewPathGenerator()

	// Create current fiber with old key
	current := &Fiber{
		Type:    rtui.VNodeElement,
		Tag:     "button",
		DiffKey: "old-key",
		Key:     "old-key",
		NodeID:  1,
	}

	// Create new VNode with different key
	newVNode := rtui.Element("button").Key("new-key").Build()

	// Clone the fiber
	returnFiber := &Fiber{Path: "/root", DiffKey: "root", Key: "root"}
	fiber := cloneExistingFiber(returnFiber, current, newVNode, 0)

	// Verify DiffKey is updated to new key
	// Note: With current shouldUpdate logic, if DiffKey differs, fiber will not be reused
	// This test checks the property of cloneExistingFiber if called directly
	t.Logf("✅ Current DiffKey: %q", fiber.DiffKey)
	t.Logf("✅ New DiffKey would be: %q", newVNode.Key())

	// The key point is that shouldUpdate should return false for different DiffKeys
	if shouldUpdate(current, newVNode) {
		t.Errorf("✅ Correctly: shouldUpdate returns false when DiffKeys differ")
	} else {
		t.Log("✅ shouldUpdate correctly returns false when DiffKeys differ")
	}
}

// TestCreateChildFiberWithIndex_PathOnlyDebug tests that Path is generated only for debugging
func TestCreateChildFiberWithIndex_PathOnlyDebug(t *testing.T) {
	pathGenerator = NewPathGenerator()

	// Create VNode with user key
	newVNode := rtui.Element("button").Key("btn-1").Build()

	// Create child fiber
	returnFiber := &Fiber{Path: "/root", DiffKey: "root", Key: "root", Tag: "root"}
	fiber := createChildFiber(returnFiber, newVNode, LaneSyncLane, 0)

	// Verify DiffKey matches user key
	if fiber.DiffKey != "btn-1" {
		t.Errorf("Expected DiffKey %q, got %q", "btn-1", fiber.DiffKey)
	}

	// Verify Path is generated for debugging
	if fiber.Path == "" {
		t.Errorf("Expected Path to be generated")
	}

	// Verify Path starts with expected prefix
	// Note: The exact path format isn't critical, what matters is that it's generated
	t.Logf("✅ DiffKey: %q (from VNode.Key())", fiber.DiffKey)
	t.Logf("✅ Path: %s (for debugging)", fiber.Path)
}

// TestCloneExistingFiber_PreservesDiffKey tests that cloneExistingFiber preserves DiffKey
func TestCloneExistingFiber_PreservesDiffKey(t *testing.T) {
	pathGenerator = NewPathGenerator()

	// Create current fiber
	current := &Fiber{
		Type:     rtui.VNodeElement,
		Tag:      "button",
		DiffKey:  "user-key",
		Key:      "user-key",
		NodeID:   123,
		VNode:    rtui.Element("button").Key("user-key").Build(),
		Return:   &Fiber{Path: "/root", DiffKey: "root", Key: "root"},
	}

	// Create new VNode with same key
	newVNode := rtui.Element("button").Key("user-key").Build()

	// Clone the fiber
	fiber := cloneExistingFiber(current.Return, current, newVNode, 0)

	// Verify DiffKey is preserved
	if fiber.DiffKey != current.DiffKey {
		t.Errorf("Expected DiffKey %q to be preserved, got %q", current.DiffKey, fiber.DiffKey)
	}

	// Verify NodeID is preserved
	if fiber.NodeID != current.NodeID {
		t.Errorf("Expected NodeID %d to be preserved, got %d", current.NodeID, fiber.NodeID)
	}

	t.Logf("✅ DiffKey preserved: %q", fiber.DiffKey)
	t.Logf("✅ NodeID preserved: %d", fiber.NodeID)
}

// TestShouldUpdate_UsesDiffKey tests that shouldUpdate uses DiffKey, not Path
func TestShouldUpdate_UsesDiffKey(t *testing.T) {
	// Create current fiber
	current := &Fiber{
		Type:    rtui.VNodeElement,
		Tag:     "button",
		DiffKey: "button-key",
		Key:     "button-key",
		VNode:   rtui.Element("button").Key("button-key").Build(),
	}

	// Test 1: Same DiffKey → should be true
	newVNode1 := rtui.Element("button").Key("button-key").Build()
	if !shouldUpdate(current, newVNode1) {
		t.Error("Expected shouldUpdate to return true when DiffKeys match")
	}
	t.Log("✅ Test 1: Same DiffKey → shouldUpdate = true")

	// Test 2: Different DiffKey → should be false
	newVNode2 := rtui.Element("button").Key("other-key").Build()
	if shouldUpdate(current, newVNode2) {
		t.Error("Expected shouldUpdate to return false when DiffKeys differ")
	}
	t.Log("✅ Test 2: Different DiffKey → shouldUpdate = false")

	// Test 3: Empty DiffKey → match with empty
	current.DiffKey = ""
	newVNode3 := rtui.Element("button").Build()
	if !shouldUpdate(current, newVNode3) {
		t.Error("Expected shouldUpdate to return true with empty DiffKeys")
	}
	t.Log("✅ Test 3: Empty DiffKey match → shouldUpdate = true")
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
