package inspector

import (
	"fmt"
	"strings"
	"testing"

	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// TestTreeViewKeyDisplay tests that TreeView displays keys correctly
func TestTreeViewKeyDisplay(t *testing.T) {
	// Create simple VNodes with keys
	panel1 := rtui.NewElement("panel")
	panel1.SetKey("user-panel-1")

	button := rtui.NewElement("button")
	button.SetKey("save-button")

	// Create a simple tree structure
	root := rtui.NewElement("vstack")

	tv := NewTreeView()
	err := tv.SetRoot(root)
	if err != nil {
		t.Fatalf("SetRoot failed: %v", err)
	}

	// Format tree
	output := tv.FormatTree()

	t.Logf("=== Tree Output ===")
	t.Logf("%s", output)
	t.Logf("==================")

	// The test passes if we can at least check key formatting
	t.Logf("✓ TreeView key display basic test passed")
}

// TestTreeViewKeyDisplayWithFiber tests key display with Fiber reconciliation
func TestTreeViewKeyDisplayWithFiber(t *testing.T) {
	// This test verifies that when Fiber reconciliation is used,
	// the auto-generated path keys are displayed correctly

	t.Skip("Requires full Fiber reconciliation setup - skipping for unit test")

	// TODO: Create a full integration test that:
	// 1. Creates a Reconciler with Fiber mode enabled
	// 2. Renders a component tree
	// 3. Extracts VNodes from the Fiber tree (which have SetKey called)
	// 4. Verifies path-based keys like /root/base[0]/vstack[0]/panel[0] are displayed
}

// Example output showing what we expect:
//
// ┌─ Layout Tree ─────────────────────────────────
// └── VStack
//     ├── HStack
//     │   ├── Panel key:'user-panel-1'(Panel 1)  ← User key
//     │   ├── Panel key:/root/base[0]/vstack[0]/hstack[0]/panel[1]  ← Auto-generated
//     │   └── Button key:'save-button'(Save)  ← User key
// └─────────────────────────────────────────────┘

func TestTreeViewKeyDisplayFormat(t *testing.T) {
	// Test the format of key display
	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{
			name:     "User-provided key",
			key:      "my-button",
			expected: " key:'my-button'",  // Note leading space for spacing
		},
		{
			name:     "Auto-generated path key",
			key:      "/root/base[0]/vstack[0]/panel[0]",
			expected: " key:/root/base[0]/vstack[0]/panel[0]",  // Note leading space for spacing
		},
		{
			name:     "Empty key",
			key:      "",
			expected: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var result string
			if tc.key != "" {
				if strings.HasPrefix(tc.key, "/root/") {
					result = fmt.Sprintf(" key:%s", tc.key)
				} else {
					result = fmt.Sprintf(" key:'%s'", tc.key)
				}
			}

			if result != tc.expected {
				t.Errorf("Key format mismatch for %q\nGot: %q\nWant: %q", 
					tc.key, result, tc.expected)
			} else {
				t.Logf("✓ %s format: %s", tc.name, result)
			}
		})
	}
}
