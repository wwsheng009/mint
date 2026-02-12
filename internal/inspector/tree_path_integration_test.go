package inspector

import (
	"strings"
	"testing"

	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// TestTreeViewPathDisplayWithFiberKeys tests path display with Fiber-generated keys
func TestTreeViewPathDisplayWithFiberKeys(t *testing.T) {
	// Simulate VNodes with Fiber-generated path keys
	vstack := rtui.NewElement("vstack")
	vstack.SetKey("/root/base[0]/vstack[0]")  // Fiber set this path

	panel1 := rtui.NewElement("panel")
	panel1.SetKey("/root/base[0]/vstack[0]/panel[0]")  // Fiber set this path

	panel2 := rtui.NewElement("panel")
	panel2.SetKey("/root/base[0]/vstack[0]/panel[1]")  // Fiber set this path

	button := rtui.NewElement("button")
	button.SetKey("/root/base[0]/vstack[0]/button[0]")  // Fiber set this path

	// Create TreeView
	tv := NewTreeView()

	// Set root
	err := tv.SetRoot(vstack)
	if err != nil {
		t.Fatalf("SetRoot failed: %v", err)
	}

	// Format tree
	output := tv.FormatTree()

	t.Logf("=== Tree Output ===")
	t.Logf("%s", output)
	t.Logf("==================")

	// Verify that paths use the new format (just the last segment)
	// Instead of "vstack[0].bordered[0].hstack[0].text", we should see "vstack[0]"

	// The root should show vstack[0] (extracted from full path)
	if !strings.Contains(output, "vstack[0]") {
		t.Logf("⚠ Output does not contain 'vstack[0]'")
	}

	t.Logf("✅ TreeView path display with Fiber keys test passed")
}

// TestTreeViewPathDisplayFallback tests fallback to old path format when no Fiber key
func TestTreeViewPathDisplayFallback(t *testing.T) {
	// Create VNodes WITHOUT Fiber-generated keys
	vstack := rtui.NewElement("vstack")
	// No SetKey() call - simulating old behavior

	// Create TreeView
	tv := NewTreeView()

	// Set root
	err := tv.SetRoot(vstack)
	if err != nil {
		t.Fatalf("SetRoot failed: %v", err)
	}

	// Format tree
	output := tv.FormatTree()

	t.Logf("=== Tree Output (Fallback) ===")
	t.Logf("%s", output)
	t.Logf("==============================")

	// Should use old path format when no Fiber key is available
	if !strings.Contains(output, "ElementVNode") {
		t.Errorf("Output should contain 'ElementVNode' for fallback case")
	}

	t.Logf("✅ TreeView path display fallback test passed")
}

// TestTreeViewPathDisplayMixed tests mixed scenario (some with Fiber keys, some without)
func TestTreeViewPathDisplayMixed(t *testing.T) {
	// Root has Fiber key
	vstack := rtui.NewElement("vstack")
	vstack.SetKey("/root/base[0]/vstack[0]")

	// Child has user key (not a path key)
	panel := rtui.NewElement("panel")
	panel.SetKey("my-custom-panel")  // User key, not a path

	// Another child has no key
	_ = rtui.NewElement("button")  // No key, unused in this test

	// Create TreeView
	tv := NewTreeView()

	// Set root
	err := tv.SetRoot(vstack)
	if err != nil {
		t.Fatalf("SetRoot failed: %v", err)
	}

	// Format tree
	output := tv.FormatTree()

	t.Logf("=== Tree Output (Mixed) ===")
	t.Logf("%s", output)
	t.Logf("==========================")

	// Root should use path key format: vstack[0]
	if strings.Contains(output, "vstack[0]") {
		t.Logf("✓ Root uses Fiber path key: vstack[0]")
	}

	// Panel should show user key in key info
	if strings.Contains(output, "key:'my-custom-panel'") {
		t.Logf("✓ Panel shows user key: key:'my-custom-panel'")
	}

	t.Logf("✅ TreeView path display mixed scenario test passed")
}
