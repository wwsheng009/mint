package reconciler

// =============================================================================
// Mixed Key Strategy Tests
// =============================================================================
// Tests for the mixed key generation strategy:
// 1. User-provided keys (highest priority)
// 2. Dynamic list detection with mandatory key requirements
// 3. Automatic path-based key generation for static UI
// =============================================================================

import (
	"testing"

	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// TestMixedKeyStrategy_UserKeyPriority tests that user-provided keys take priority
func TestMixedKeyStrategy_UserKeyPriority(t *testing.T) {
	parent := &Fiber{
		Path: "/root/base[0]",
		Tag:  "vstack",
	}

	// Create VNode with user key
	vnode := rtui.NewElement("button")
	vnode.SetKey("my-button")

	// Create child fiber
	fiber := createChildFiber(parent, vnode, LaneSyncLane, 0)

	// Verify user key is used
	if fiber.Key != "my-button" {
		t.Errorf("Expected key 'my-button', got '%s'", fiber.Key)
	}

	// ✨ New Design: Path includes type index "/button[0]/key[my-button]"
	expectedPath := "/root/base[0]/button[0]/key[my-button]"
	if fiber.Path != expectedPath {
		t.Errorf("Expected path '%s', got '%s'", expectedPath, fiber.Path)
	}

	// ✨ New Design: PathSegment is the last part of the type path
	if fiber.PathSegment != "button[0]" {
		t.Errorf("Expected path segment 'button[0]', got '%s'", fiber.PathSegment)
	}

	t.Logf("✅ User key priority test passed")
	t.Logf("   Key: %s", fiber.Key)
	t.Logf("   Path: %s", fiber.Path)
}

// TestMixedKeyStrategy_StaticUIAutoKey tests automatic key generation for static UI
func TestMixedKeyStrategy_StaticUIAutoKey(t *testing.T) {
	parent := &Fiber{
		Path: "/root/base[0]/vstack[0]",
		Tag:  "vstack",
	}

	// Create VNode without user key (static UI)
	vnode := rtui.NewElement("panel")

	// Create child fiber
	fiber := createChildFiber(parent, vnode, LaneSyncLane, 0)

	// Verify automatic path key is generated
	if fiber.Key == "" {
		t.Error("Expected automatic key to be generated")
	}

	// Verify path format
	expectedPathPrefix := "/root/base[0]/vstack[0]/panel["
	if len(fiber.Path) < len(expectedPathPrefix) || fiber.Path[:len(expectedPathPrefix)] != expectedPathPrefix {
		t.Errorf("Expected path to start with '%s', got '%s'", expectedPathPrefix, fiber.Path)
	}

	t.Logf("✅ Static UI auto-key test passed")
	t.Logf("   Key: %s", fiber.Key)
	t.Logf("   Path: %s", fiber.Path)
	t.Logf("   PathSegment: %s", fiber.PathSegment)
}

// TestMixedKeyStrategy_MultipleStaticChildren tests index increment for multiple children
func TestMixedKeyStrategy_MultipleStaticChildren(t *testing.T) {
	parent := &Fiber{
		Path: "/root/base[0]",
		Tag:  "vstack",
	}

	// Create multiple panel children (no keys)
	children := []rtui.VNode{
		rtui.NewElement("panel"),
		rtui.NewElement("panel"),
		rtui.NewElement("button"),
		rtui.NewElement("button"),
	}

	// Create fibers using createAllNewChildren (which properly handles type indexing)
	firstChild := createAllNewChildren(parent, children, LaneSyncLane)

	// Collect fibers into slice
	fibers := []*Fiber{}
	for child := firstChild; child != nil; child = child.Sibling {
		fibers = append(fibers, child)
	}

	// Verify we have 4 children
	if len(fibers) != 4 {
		t.Fatalf("Expected 4 children, got %d", len(fibers))
	}

	// ✨ New Design: Key uses index fallback, Path is for debugging
	// Verify first panel (index 0)
	if fibers[0].Key != "0" {
		t.Errorf("Expected key '0' (index fallback), got '%s'", fibers[0].Key)
	}
	if fibers[0].Path != "/root/base[0]/panel[0]" {
		t.Errorf("Expected path '/root/base[0]/panel[0]', got '%s'", fibers[0].Path)
	}

	// Verify second panel (index 1)
	if fibers[1].Key != "1" {
		t.Errorf("Expected key '1' (index fallback), got '%s'", fibers[1].Key)
	}
	if fibers[1].Path != "/root/base[0]/panel[1]" {
		t.Errorf("Expected path '/root/base[0]/panel[1]', got '%s'", fibers[1].Path)
	}

	// Verify first button (index 2)
	if fibers[2].Key != "2" {
		t.Errorf("Expected key '2' (index fallback), got '%s'", fibers[2].Key)
	}
	if fibers[2].Path != "/root/base[0]/button[0]" {
		t.Errorf("Expected path '/root/base[0]/button[0]', got '%s'", fibers[2].Path)
	}

	// Verify second button (index 3)
	if fibers[3].Key != "3" {
		t.Errorf("Expected key '3' (index fallback), got '%s'", fibers[3].Key)
	}
	if fibers[3].Path != "/root/base[0]/button[1]" {
		t.Errorf("Expected path '/root/base[0]/button[1]', got '%s'", fibers[3].Path)
	}

	t.Logf("✅ Multiple static children test passed")
	for _, f := range fibers {
		t.Logf("   Key: %s, Segment: %s", f.Key, f.PathSegment)
	}
}

// TestMixedKeyStrategy_DynamicListRequireKey tests that dynamic lists require keys
func TestMixedKeyStrategy_DynamicListRequireKey(t *testing.T) {
	parent := &Fiber{
		Path: "/root/base[0]",
		Tag:  "List", // This is a dynamic list
	}

	// Create VNode without key (should panic)
	vnode := rtui.NewElement("item")

	// Should panic with detailed error message
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for missing key in dynamic list")
		} else {
			t.Logf("✅ Dynamic list panic test passed")
			t.Logf("   Panic message: %v", r)
		}
	}()

	createChildFiber(parent, vnode, LaneSyncLane, 0)
}

// TestMixedKeyStrategy_DynamicListWithKey tests that dynamic lists work with keys
func TestMixedKeyStrategy_DynamicListWithKey(t *testing.T) {
	parent := &Fiber{
		Path: "/root/base[0]",
		Tag:  "List", // This is a dynamic list
	}

	// Create VNode WITH key (should not panic)
	vnode := rtui.NewElement("item")
	vnode.SetKey("item-123")

	// Should NOT panic
	fiber := createChildFiber(parent, vnode, LaneSyncLane, 0)

	if fiber.Key != "item-123" {
		t.Errorf("Expected key 'item-123', got '%s'", fiber.Key)
	}

	t.Logf("✅ Dynamic list with key test passed")
	t.Logf("   Key: %s", fiber.Key)
	t.Logf("   Path: %s", fiber.Path)
}

// TestMixedKeyStrategy_RootNode tests root node path generation
func TestMixedKeyStrategy_RootNode(t *testing.T) {
	// Root node has no parent
	vnode := rtui.NewElement("div")
	vnode.SetLayer(rtui.LayerBase)

	// Create root fiber
	fiber := createChildFiber(nil, vnode, LaneSyncLane, 0)

	// Verify root path
	expectedPath := "/root/base[0]"
	if fiber.Path != expectedPath {
		t.Errorf("Expected path '%s', got '%s'", expectedPath, fiber.Path)
	}

	t.Logf("✅ Root node test passed")
	t.Logf("   Key: %s", fiber.Key)
	t.Logf("   Path: %s", fiber.Path)
}

// TestMixedKeyStrategy_DifferentLayers tests path generation for different layers
func TestMixedKeyStrategy_DifferentLayers(t *testing.T) {
	layers := []struct {
		layer    rtui.Layer
		expected string
	}{
		{rtui.LayerBase, "/root/base[0]"},
		{rtui.LayerModal, "/root/modal[0]"},
		{rtui.LayerOverlay, "/root/overlay[0]"},
		{rtui.LayerInspector, "/root/inspector[0]"},
	}

	for _, tc := range layers {
		t.Run(tc.layer.String(), func(t *testing.T) {
			vnode := rtui.NewElement("div")
			vnode.SetLayer(tc.layer)

			fiber := createChildFiber(nil, vnode, LaneSyncLane, 0)

			if fiber.Path != tc.expected {
				t.Errorf("Expected path '%s', got '%s'", tc.expected, fiber.Path)
			}

			t.Logf("✅ Layer %s: %s", tc.layer.String(), fiber.Path)
		})
	}
}

// TestMixedKeyStrategy_SiblingIndex tests that sibling index is correctly set
func TestMixedKeyStrategy_SiblingIndex(t *testing.T) {
	parent := &Fiber{
		Path: "/root/base[0]",
		Tag:  "vstack",
	}

	children := []rtui.VNode{
		rtui.NewElement("panel"),
		rtui.NewElement("panel"),
		rtui.NewElement("panel"),
	}

	// Create fibers using createAllNewChildren
	firstChild := createAllNewChildren(parent, children, LaneSyncLane)

	// Verify each child's SiblingIndex
	i := 0
	for child := firstChild; child != nil; child = child.Sibling {
		if child.SiblingIndex != i {
			t.Errorf("Expected SiblingIndex %d, got %d", i, child.SiblingIndex)
		}
		i++
	}

	t.Logf("✅ Sibling index test passed")
}

// TestMixedKeyStrategy_NestedStructure tests nested structure path generation
func TestMixedKeyStrategy_NestedStructure(t *testing.T) {
	// Create root
	root := createChildFiber(nil, rtui.NewElement("div"), LaneSyncLane, 0)
	t.Logf("Root: %s", root.Path)

	// Create first level child
	child1 := createChildFiber(root, rtui.NewElement("vstack"), LaneSyncLane, 0)
	t.Logf("Child1: %s", child1.Path)

	// Create second level child (nested)
	child2 := createChildFiber(child1, rtui.NewElement("panel"), LaneSyncLane, 0)
	t.Logf("Child2: %s", child2.Path)

	// Verify path depth
	if child2.Path == "/root/base[0]/vstack[0]/panel[0]" {
		t.Logf("✅ Nested structure test passed")
	} else {
		t.Errorf("Expected '/root/base[0]/vstack[0]/panel[0]', got '%s'", child2.Path)
	}
}

// TestExtractPathSegment tests the path segment extraction utility
func TestExtractPathSegment(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"/root/base[0]", "base[0]"},
		{"/root/base[0]/vstack[0]", "vstack[0]"},
		{"/root/base[0]/vstack[0]/panel[1]", "panel[1]"},
		{"", ""},
		{"single", "single"},
	}

	for _, tc := range tests {
		result := extractPathSegment(tc.path)
		if result != tc.expected {
			t.Errorf("extractPathSegment(%q) = %q, expected %q",
				tc.path, result, tc.expected)
		}
	}

	t.Logf("✅ Path segment extraction test passed")
}

// TestGetTypeIDFromSegment tests type ID extraction from path segment
func TestGetTypeIDFromSegment(t *testing.T) {
	tests := []struct {
		segment  string
		expected string
	}{
		{"button[0]", "button"},
		{"panel[1]", "panel"},
		{"vstack[0]", "vstack"},
		{"text", "text"},
	}

	for _, tc := range tests {
		result := getTypeIDFromSegment(tc.segment)
		if result != tc.expected {
			t.Errorf("getTypeIDFromSegment(%q) = %q, expected %q",
				tc.segment, result, tc.expected)
		}
	}

	t.Logf("✅ Type ID extraction test passed")
}
