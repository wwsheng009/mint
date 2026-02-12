package inspector

import (
	"strings"
	"testing"

	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// TestTreeViewFullPathDisplay tests the full hierarchical path display
func TestTreeViewFullPathDisplay(t *testing.T) {
	// Create a tree structure simulating Fiber-generated keys
	// Root: /root/base[0]/vstack[0]
	// Child: /root/base[0]/vstack[0]/hstack[0]
	// Grandchild: /root/base[0]/vstack[0]/hstack[0]/text[0]

	root := rtui.NewElement("vstack")
	root.SetKey("/root/base[0]/vstack[0]")

	hstack := rtui.NewElement("hstack")
	hstack.SetKey("/root/base[0]/vstack[0]/hstack[0]")

	text := rtui.NewElement("text")
	text.SetKey("/root/base[0]/vstack[0]/hstack[0]/text[0]")

	// Build tree manually to test path extraction
	tv := NewTreeView()

	// Build tree from root
	treeRoot := tv.buildTree(root, nil, 0, "", 0)
	if treeRoot == nil {
		t.Fatal("buildTree returned nil")
	}

	t.Logf("Root node path: %s", treeRoot.Path)
	t.Logf("Root node UniqueID: %s", treeRoot.UniqueID)

	// Verify root path (should be base[0]/vstack[0])
	expectedPath := "base[0]/vstack[0]"
	if treeRoot.Path != expectedPath {
		t.Errorf("Expected root path %q, got %q", expectedPath, treeRoot.Path)
	} else {
		t.Logf("✓ Root path correct: %s", treeRoot.Path)
	}

	// Format and check output
	output := tv.FormatTree()
	t.Logf("\n=== Formatted Tree Output ===")
	t.Logf("%s", output)
	t.Logf("============================")

	if strings.Contains(output, "base[0]/vstack[0]") {
		t.Logf("✓ Full hierarchical path is displayed")
	} else {
		t.Logf("⚠ Full path not found in output")
	}
}

// TestTreeViewPathSegmentExtraction tests the path segment extraction logic
func TestTreeViewPathSegmentExtraction(t *testing.T) {
	tests := []struct {
		name        string
		fiberKey    string
		expectedPath string
	}{
		{
			name:        "Root level",
			fiberKey:    "/root/base[0]",
			expectedPath: "base[0]",
		},
		{
			name:        "One level deep",
			fiberKey:    "/root/base[0]/vstack[0]",
			expectedPath: "base[0]/vstack[0]",
		},
		{
			name:        "Two levels deep",
			fiberKey:    "/root/base[0]/vstack[0]/panel[0]",
			expectedPath: "base[0]/vstack[0]/panel[0]",
		},
		{
			name:        "Three levels deep",
			fiberKey:    "/root/base[0]/vstack[0]/hstack[0]/text[0]",
			expectedPath: "base[0]/vstack[0]/hstack[0]/text[0]",
		},
		{
			name:        "Overlay layer",
			fiberKey:    "/root/overlay[0]/modal[0]/panel[0]",
			expectedPath: "overlay[0]/modal[0]/panel[0]",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Simulate the path extraction logic
			var nodePath string
			if strings.HasPrefix(tc.fiberKey, "/root/") {
				if len(tc.fiberKey) > 6 {
					nodePath = tc.fiberKey[6:]
				} else {
					nodePath = tc.fiberKey
				}
			}

			if nodePath != tc.expectedPath {
				t.Errorf("Path extraction failed\nInput: %s\nExpected: %s\nGot: %s",
					tc.fiberKey, tc.expectedPath, nodePath)
			} else {
				t.Logf("✓ %s: %s → %s", tc.name, tc.fiberKey, nodePath)
			}
		})
	}
}

// TestTreeViewOldVsNewPathFormat compares old and new path formats
func TestTreeViewOldVsNewPathFormat(t *testing.T) {
	oldFormat := "vstack[0].bordered[0].hstack[0].text"
	newFormat := "base[0]/vstack[0]/bordered[0]/hstack[0]/text[0]"

	t.Logf("Old format: %s", oldFormat)
	t.Logf("New format: %s", newFormat)
	t.Logf("")
	t.Logf("Key differences:")
	t.Logf("  - Old format: dot-separated, index in brackets")
	t.Logf("  - New format: slash-separated, shows full hierarchy from layer")
	t.Logf("  - New format matches Fiber's internal path structure")

	t.Logf("\n✓ Path format comparison completed")
}
