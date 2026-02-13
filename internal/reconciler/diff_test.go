package reconciler

import (
	"testing"

	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Tests for cloneExistingFiber VNode Key Sync Fix
// =============================================================================

// TestCloneExistingFiberKeySync_RootPath tests that VNode keys are synced
// when the path starts with "/root/" (existing behavior)
func TestCloneExistingFiberKeySync_RootPath(t *testing.T) {
	// Setup path generator
	pathGenerator = NewPathGenerator()

	// Create a current fiber with a root path
	current := &Fiber{
		Type:  rtui.VNodeElement,
		Tag:   "button",
		Path:  "/root/base[0]/button[0]",
		Key:   "/root/base[0]/button[0]",
		VNode: rtui.Element("button").Build(),
	}

	// Create a new VNode
	newVNode := rtui.Element("button").Build()

	// Clone the fiber
	returnFiber := &Fiber{Path: "/root"}
	fiber := cloneExistingFiber(returnFiber, current, newVNode, 0)

	// Verify VNode key was synced to fiber path
	expectedKey := "/root/base[0]/button[0]"
	if newVNode.Key() != expectedKey {
		t.Errorf("Expected VNode key %q, got %q", expectedKey, newVNode.Key())
	}
	if fiber.Path != expectedKey {
		t.Errorf("Expected fiber path %q, got %q", expectedKey, fiber.Path)
	}
}

// TestCloneExistingFiberKeySync_NonRootPath tests that VNode keys are synced
// even when path doesn't start with "/root/" (FIX CRITICAL)
func TestCloneExistingFiberKeySync_NonRootPath(t *testing.T) {
	// Setup path generator
	pathGenerator = NewPathGenerator()

	// Create a current fiber with a non-root path (e.g., modal)
	current := &Fiber{
		Type:  rtui.VNodeElement,
		Tag:   "button",
		Path:  "/modal[0]/button[0]",
		Key:   "/modal[0]/button[0]",
		VNode: rtui.Element("button").Build(),
	}

	// Create a new VNode
	newVNode := rtui.Element("button").Build()

	// Clone the fiber
	returnFiber := &Fiber{Path: "/root"}
	fiber := cloneExistingFiber(returnFiber, current, newVNode, 0)

	// Verify VNode key was synced to fiber path (FIX: this should work now!)
	expectedKey := "/modal[0]/button[0]"
	if newVNode.Key() != expectedKey {
		t.Errorf("Expected VNode key %q, got %q", expectedKey, newVNode.Key())
	}
	if fiber.Path != expectedKey {
		t.Errorf("Expected fiber path %q, got %q", expectedKey, fiber.Path)
	}
}

// TestCloneExistingFiberKeySync_WithoutUserKeyChange tests that path is preserved
// when user key is not changed (and no user key on vnode)
func TestCloneExistingFiberKeySync_WithoutUserKeyChange(t *testing.T) {
	// Setup path generator
	pathGenerator = NewPathGenerator()

	// Create a current fiber with a complex path
	current := &Fiber{
		Type:  rtui.VNodeElement,
		Tag:   "button",
		Path:  "/root/base[0]/hstack[0]/modal[0]/button[0]",
		Key:   "/root/base[0]/hstack[0]/modal[0]/button[0]",
		VNode: rtui.Element("button").Build(),
	}

	// Create a new VNode WITHOUT user key
	newVNode := rtui.Element("button").Build()

	// Clone the fiber
	returnFiber := &Fiber{Path: "/root/base[0]/hstack[0]"}
	fiber := cloneExistingFiber(returnFiber, current, newVNode, 0)

	// Verify path and key are preserved
	expectedPath := "/root/base[0]/hstack[0]/modal[0]/button[0]"
	if fiber.Path != expectedPath {
		t.Errorf("Expected fiber path %q, got %q", expectedPath, fiber.Path)
	}
	if fiber.Key != expectedPath {
		t.Errorf("Expected fiber key %q, got %q", expectedPath, fiber.Key)
	}
	if newVNode.Key() != expectedPath {
		t.Errorf("Expected VNode key %q, got %q", expectedPath, newVNode.Key())
	}
}

// TestCloneExistingFiberKeySync_FallbackToKey tests that when path is empty,
// the key is used as fallback
func TestCloneExistingFiberKeySync_FallbackToKey(t *testing.T) {
	// Setup path generator
	pathGenerator = NewPathGenerator()

	// Create a current fiber with empty path but with key
	current := &Fiber{
		Type:  rtui.VNodeElement,
		Tag:   "button",
		Path:  "",
		Key:   "my-button-key",
		VNode: rtui.Element("button").Build(),
	}

	// Create a new VNode
	newVNode := rtui.Element("button").Build()

	// Clone the fiber
	returnFiber := &Fiber{Path: "/root"}
	fiber := cloneExistingFiber(returnFiber, current, newVNode, 0)

	// Verify VNode key falls back to fiber key
	if newVNode.Key() != "my-button-key" {
		t.Errorf("Expected VNode key %q, got %q", "my-button-key", newVNode.Key())
	}
	if fiber.Key != "my-button-key" {
		t.Errorf("Expected fiber key %q, got %q", "my-button-key", fiber.Key)
	}
}

// TestCloneExistingFiberKeySync_EmptyBoth tests that when both path and key are empty,
// nothing is synced (no panic)
func TestCloneExistingFiberKeySync_EmptyBoth(t *testing.T) {
	// Setup path generator
	pathGenerator = NewPathGenerator()

	// Create a current fiber with empty path and key
	current := &Fiber{
		Type:  rtui.VNodeElement,
		Tag:   "button",
		Path:  "",
		Key:   "",
		VNode: rtui.Element("button").Build(),
	}

	// Create a new VNode
	newVNode := rtui.Element("button").Build()

	// Clone the fiber (should not panic)
	returnFiber := &Fiber{Path: "/root"}
	fiber := cloneExistingFiber(returnFiber, current, newVNode, 0)

	// Verify no key was set
	if newVNode.Key() != "" {
		t.Errorf("Expected empty VNode key, got %q", newVNode.Key())
	}
	if fiber.Path != "" {
		t.Errorf("Expected empty fiber path, got %q", fiber.Path)
	}
}

// TestCloneExistingFiberKeySync_WithUserKeyChange tests that when user changes key,
// new path is generated with user key
func TestCloneExistingFiberKeySync_WithUserKeyChange(t *testing.T) {
	// Setup path generator
	pathGenerator = NewPathGenerator()

	// Create a current fiber
	current := &Fiber{
		Type:  rtui.VNodeElement,
		Tag:   "button",
		Path:  "/root/base[0]/button[0]",
		Key:   "/root/base[0]/button[0]",
		VNode: rtui.Element("button").Build(),
	}

	// Create a new VNode WITH user key (user changed it)
	newVNode := rtui.Element("button").Build()
	newVNode.SetKey("new-user-key")

	// Clone the fiber
	returnFiber := &Fiber{Path: "/root"}
	fiber := cloneExistingFiber(returnFiber, current, newVNode, 0)

	// Verify new path is generated with user key
	expectedPath := "/root/button[0]/key[new-user-key]"
	if fiber.Path != expectedPath {
		t.Errorf("Expected fiber path %q, got %q", expectedPath, fiber.Path)
	}
	if fiber.Key != "new-user-key" {
		t.Errorf("Expected fiber key %q, got %q", "new-user-key", fiber.Key)
	}
	// VNode key should be synced to the new path
	if newVNode.Key() != expectedPath {
		t.Errorf("Expected VNode key %q, got %q", expectedPath, newVNode.Key())
	}
}

// TestCloneExistingFiberKeySync_ModalScenario tests the real-world scenario:
// A button in a modal that doesn't have a "/root/" prefix
func TestCloneExistingFiberKeySync_ModalScenario(t *testing.T) {
	// Setup path generator
	pathGenerator = NewPathGenerator()

	// Simulate a modal button fiber (without /root/ prefix)
	current := &Fiber{
		Type:  rtui.VNodeElement,
		Tag:   "button",
		Path:  "/modal[0]/vstack[0]/button[0]",
		Key:   "/modal[0]/vstack[0]/button[0]",
		VNode: rtui.Element("button").Build(),
	}

	// Create a new VNode
	newVNode := rtui.Element("button").Build()

	// Clone the fiber
	returnFiber := &Fiber{Path: "/modal[0]"}
	fiber := cloneExistingFiber(returnFiber, current, newVNode, 0)

	// Verify VNode key is synced (CRITICAL: this ensures HitMap works!)
	expectedKey := "/modal[0]/vstack[0]/button[0]"
	// This is the key that will be used for Instance: "vnode:" + fiber.Path
	// And this must match HitMap NodeID which is: box.VNode.Key()
	instanceKey := "vnode:" + fiber.Path
	hitMapNodeID := newVNode.Key()

	if fiber.Path != expectedKey {
		t.Errorf("Expected fiber path %q, got %q", expectedKey, fiber.Path)
	}
	if newVNode.Key() != expectedKey {
		t.Errorf("Expected VNode key %q for HitMap, got %q", expectedKey, newVNode.Key())
	}

	// Verify Instance key matches HitMap NodeID (this is what the fix enables!)
	expectedInstanceKey := "vnode:" + expectedKey
	if instanceKey != expectedInstanceKey {
		t.Errorf("Instance key %q doesn't match expected %q", instanceKey, expectedInstanceKey)
	}
	if hitMapNodeID != expectedKey {
		t.Errorf("HitMap NodeID %q doesn't match fiber path %q", hitMapNodeID, fiber.Path)
	}
	t.Logf("✅ Instance key and HitMap NodeID are now synchronized!")
	t.Logf("   Instance key: %s", instanceKey)
	t.Logf("   HitMap NodeID: %s", hitMapNodeID)
}

// =============================================================================
// Tests for createChildFiberWithIndex layer-based path generation
// =============================================================================

// TestCreateChildFiberWithIndex_LayerNode tests that layer nodes (modal, overlay, etc.)
// get layer-based paths even when not root's direct child
func TestCreateChildFiberWithIndex_LayerNode(t *testing.T) {
	// Setup path generator
	pathGenerator = NewPathGenerator()

	// Create a parent fiber (simulating a base layer node)
	parentFiber := &Fiber{
		Type:  rtui.VNodeElement,
		Tag:   "vstack",
		Key:   "/root/base[0]/vstack[0]",
		Path:  "/root/base[0]/vstack[0]",
		VNode: rtui.Element("vstack").Build(),
	}

	// Create a Modal VNode with layer property
	modalVNode := rtui.Element("div").Build()
	modalVNode.SetLayer(rtui.LayerModal)

	// Create child fiber - should use layer-based path because it's a layer node
	fiber := createChildFiber(parentFiber, modalVNode, LaneSyncLane, 0)

	// Verify: Modal should get layer-based path, not parent-based path
	expectedPath := "/root/modal[0]"
	if fiber.Path != expectedPath {
		t.Errorf("Expected fiber path %q (layer-based), got %q", expectedPath, fiber.Path)
	}

	// VNode key should be synced
	if modalVNode.Key() != expectedPath {
		t.Errorf("Expected VNode key %q, got %q", expectedPath, modalVNode.Key())
	}

	t.Logf("✅ Modal node got layer-based path correctly!")
	t.Logf("   Fiber.Path: %s", fiber.Path)
	t.Logf("   VNode.Key(): %s", modalVNode.Key())
}

// TestCreateChildFiberWithIndex_ModalTooltipNesting tests that nested layer nodes
// (e.g., tooltip inside modal) each get their own layer-based paths
func TestCreateChildFiberWithIndex_ModalTooltipNesting(t *testing.T) {
	// Setup path generator
	pathGenerator = NewPathGenerator()

	// Create a base parent
	baseParent := &Fiber{
		Type:  rtui.VNodeElement,
		Tag:   "root",
		Key:   "root",
		Path:  "/root",
		VNode: rtui.Element("root").Build(),
	}

	// Create a Modal VNode
	modalVNode := rtui.Element("div").Build()
	modalVNode.SetLayer(rtui.LayerModal)

	// Create modal fiber
	modalFiber := createChildFiber(baseParent, modalVNode, LaneSyncLane, 0)

	// Modal should get /root/modal[0]
	if modalFiber.Path != "/root/modal[0]" {
		t.Errorf("Expected modal path %q, got %q", "/root/modal[0]", modalFiber.Path)
	}

	// Now create a tooltip inside the modal
	tooltipVNode := rtui.Element("div").Build()
	tooltipVNode.SetLayer(rtui.LayerTooltip)

	tooltipFiber := createChildFiber(modalFiber, tooltipVNode, LaneSyncLane, 0)

	// Tooltip should get /root/tooltip[0] (layer-based), NOT /root/modal[0]/tooltip[0]
	if tooltipFiber.Path != "/root/tooltip[0]" {
		t.Errorf("Expected tooltip path %q (layer-based), got %q", "/root/tooltip[0]", tooltipFiber.Path)
	}

	t.Logf("✅ Nested layer nodes each get their own layer-based path!")
	t.Logf("   Modal path: %s", modalFiber.Path)
	t.Logf("   Tooltip path: %s", tooltipFiber.Path)
}

// TestCreateChildFiberWithIndex_ModalChildren tests that modal's children
// (non-layer nodes) get paths relative to modal, not their grandparents
func TestCreateChildFiberWithIndex_ModalChildren(t *testing.T) {
	// Setup path generator
	pathGenerator = NewPathGenerator()

	// Create a Modal parent fiber
	modalParent := &Fiber{
		Type:  rtui.VNodeElement,
		Tag:   "div",
		Key:   "/root/modal[0]",
		Path:  "/root/modal[0]",
		VNode: rtui.Element("div").Build(),
	}

	// Create a normal button child (no layer property = LayerBase)
	buttonVNode := rtui.Element("button").Build()

	// Create child fiber - should use parent-based path (NOT layer-based)
	buttonFiber := createChildFiber(modalParent, buttonVNode, LaneSyncLane, 0)

	// Verify: Button should get parent-based path, not layer-based
	expectedPath := "/root/modal[0]/button[0]"
	if buttonFiber.Path != expectedPath {
		t.Errorf("Expected fiber path %q (parent-based), got %q", expectedPath, buttonFiber.Path)
	}

	// VNode key should be synced
	if buttonVNode.Key() != expectedPath {
		t.Errorf("Expected VNode key %q, got %q", expectedPath, buttonVNode.Key())
	}

	t.Logf("✅ Modal's children get parent-based paths correctly!")
	t.Logf("   Button fiber.Path: %s", buttonFiber.Path)
	t.Logf("   Button VNode.Key(): %s", buttonVNode.Key())
}

// TestCreateChildFiberWithIndex_OverlayNode tests that overlay nodes get layer-based paths
func TestCreateChildFiberWithIndex_OverlayNode(t *testing.T) {
	// Setup path generator
	pathGenerator = NewPathGenerator()

	// Create a parent in base layer
	parentFiber := &Fiber{
		Type:  rtui.VNodeElement,
		Tag:   "vstack",
		Key:   "/root/base[0]/vstack[0]",
		Path:  "/root/base[0]/vstack[0]",
		VNode: rtui.Element("vstack").Build(),
	}

	// Create an Overlay VNode
	overlayVNode := rtui.Element("div").Build()
	overlayVNode.SetLayer(rtui.LayerOverlay)

	// Create child fiber
	overlayFiber := createChildFiber(parentFiber, overlayVNode, LaneSyncLane, 0)

	// Verify: Overlay should get layer-based path
	expectedPath := "/root/overlay[0]"
	if overlayFiber.Path != expectedPath {
		t.Errorf("Expected overlay path %q, got %q", expectedPath, overlayFiber.Path)
	}

	t.Logf("✅ Overlay node got layer-based path: %s", overlayFiber.Path)
}

// TestCreateChildFiberWithIndex_InspectorNode tests that inspector nodes get layer-based paths
func TestCreateChildFiberWithIndex_InspectorNode(t *testing.T) {
	// Setup path generator
	pathGenerator = NewPathGenerator()

	// Create a parent
	parentFiber := &Fiber{
		Type:  rtui.VNodeElement,
		Tag:   "root",
		Key:   "root",
		Path:  "/root",
		VNode: rtui.Element("root").Build(),
	}

	// Create an Inspector VNode
	inspectorVNode := rtui.Element("div").Build()
	inspectorVNode.SetLayer(rtui.LayerInspector)

	// Create child fiber
	inspectorFiber := createChildFiber(parentFiber, inspectorVNode, LaneSyncLane, 0)

	// Verify: Inspector should get layer-based path
	expectedPath := "/root/inspector[0]"
	if inspectorFiber.Path != expectedPath {
		t.Errorf("Expected inspector path %q, got %q", expectedPath, inspectorFiber.Path)
	}

	t.Logf("✅ Inspector node got layer-based path: %s", inspectorFiber.Path)
}

// TestCreateChildFiberWithIndex_AllLayers tests all layer types get correct paths
func TestCreateChildFiberWithIndex_AllLayers(t *testing.T) {
	// Setup path generator
	pathGenerator = NewPathGenerator()

	rootFiber := &Fiber{
		Type:  rtui.VNodeElement,
		Tag:   "root",
		Key:   "root",
		Path:  "/root",
		VNode: rtui.Element("root").Build(),
	}

	// Test all layer types
	tests := []struct {
		name           string
		layer          rtui.Layer
		expectedPath   string
	}{
		{"Modal", rtui.LayerModal, "/root/modal[0]"},
		{"Overlay", rtui.LayerOverlay, "/root/overlay[0]"},
		{"Tooltip", rtui.LayerTooltip, "/root/tooltip[0]"},
		{"Inspector", rtui.LayerInspector, "/root/inspector[0]"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			vnode := rtui.Element("div").Build()
			vnode.SetLayer(tc.layer)

			fiber := createChildFiber(rootFiber, vnode, LaneSyncLane, 0)

			if fiber.Path != tc.expectedPath {
				t.Errorf("Expected path %q, got %q", tc.expectedPath, fiber.Path)
			}

			t.Logf("✅ %s layer node got path: %s", tc.name, fiber.Path)
		})
	}
}
