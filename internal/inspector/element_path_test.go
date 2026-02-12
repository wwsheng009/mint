package inspector

import (
	"strings"
	"testing"

	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// TestVNodeWithKey is a mock VNode that implements Key() method
type TestVNodeWithPathKey struct {
	*rtui.ElementVNode
	key string
}

func (n *TestVNodeWithPathKey) Key() string {
	return n.key
}

// TestGetPath_FiberKey verifies that getPath correctly extracts Fiber-generated paths
func TestGetPath_FiberKey(t *testing.T) {
	tests := []struct {
		name           string
		fiberKey       string
		expectedPath   string
		expectedEmpty  bool
	}{
		{
			name:          "Full hierarchical path",
			fiberKey:      "/root/base[0]/vstack[0]/panel[0]",
			expectedPath:  "base[0]/vstack[0]/panel[0]",
			expectedEmpty: false,
		},
		{
			name:          "Single layer path",
			fiberKey:      "/root/base[0]",
			expectedPath:  "base[0]",
			expectedEmpty: false,
		},
		{
			name:          "Deep hierarchy",
			fiberKey:      "/root/base[0]/vstack[0]/hstack[0]/text[0]",
			expectedPath:  "base[0]/vstack[0]/hstack[0]/text[0]",
			expectedEmpty: false,
		},
		{
			name:          "User-provided key (not a path)",
			fiberKey:      "my-button",
			expectedPath:  "",
			expectedEmpty: true,
		},
		{
			name:          "Empty key",
			fiberKey:      "",
			expectedPath:  "",
			expectedEmpty: true,
		},
		{
			name:          "Path without /root/ prefix",
			fiberKey:      "base[0]/vstack[0]",
			expectedPath:  "",
			expectedEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vnode := &TestVNodeWithPathKey{
				ElementVNode: rtui.NewElement("test"),
				key:          tt.fiberKey,
			}

			path := getPath(vnode)

			if tt.expectedEmpty {
				if path != "" {
					t.Errorf("getPath() = %q, expected empty string", path)
				}
			} else {
				if path != tt.expectedPath {
					t.Errorf("getPath() = %q, expected %q", path, tt.expectedPath)
				}
			}
		})
	}
}

// TestExtractElementInfo_WithPath verifies that ExtractElementInfo populates the Path field
func TestExtractElementInfo_WithPath(t *testing.T) {
	// Create a VNode with a Fiber-generated path key
	vnode := &TestVNodeWithPathKey{
		ElementVNode: rtui.NewElement("test"),
		key:          "/root/base[0]/vstack[0]/panel[0]",
	}

	info := ExtractElementInfo(vnode)

	// Verify Path field is populated
	if info.Path != "base[0]/vstack[0]/panel[0]" {
		t.Errorf("ExtractElementInfo().Path = %q, expected %q", info.Path, "base[0]/vstack[0]/panel[0]")
	}

	// Verify Key field still has the full Fiber key
	if info.Key != "/root/base[0]/vstack[0]/panel[0]" {
		t.Errorf("ExtractElementInfo().Key = %q, expected %q", info.Key, "/root/base[0]/vstack[0]/panel[0]")
	}
}

// TestExtractElementInfo_NoPath verifies behavior when VNode has no Fiber key
func TestExtractElementInfo_NoPath(t *testing.T) {
	vnode := &TestVNodeWithPathKey{
		ElementVNode: rtui.NewElement("test"),
		key:          "", // No key
	}

	info := ExtractElementInfo(vnode)

	// Path should be empty
	if info.Path != "" {
		t.Errorf("ExtractElementInfo().Path = %q, expected empty string", info.Path)
	}

	// Key should also be empty
	if info.Key != "" {
		t.Errorf("ExtractElementInfo().Key = %q, expected empty string", info.Key)
	}
}

// TestExtractElementInfo_UserKey verifies behavior with user-provided key (now with full path)
func TestExtractElementInfo_UserKey(t *testing.T) {
	// After the fix, user keys are stored with full path in VNode.Key()
	// e.g., "/root/base[0]/vstack[0]/key[my-custom-button]"
	userKeyPath := "/root/base[0]/vstack[0]/key[my-custom-button]"
	vnode := &TestVNodeWithPathKey{
		ElementVNode: rtui.NewElement("button"),
		key:          userKeyPath, // Full path with user key suffix
	}

	info := ExtractElementInfo(vnode)

	// Path should now show the full path (minus /root/ prefix)
	expectedPath := "base[0]/vstack[0]/key[my-custom-button]"
	if info.Path != expectedPath {
		t.Errorf("ExtractElementInfo().Path = %q, expected %q", info.Path, expectedPath)
	}

	// Key should have the full path
	if info.Key != userKeyPath {
		t.Errorf("ExtractElementInfo().Key = %q, expected %q", info.Key, userKeyPath)
	}
}

// TestFormatSidebarWithPath verifies the sidebar displays the path correctly
func TestFormatSidebarWithPath(t *testing.T) {
	vnode := &TestVNodeWithPathKey{
		ElementVNode: rtui.NewElement("panel"),
		key:          "/root/base[0]/vstack[0]/panel[0]",
	}

	info := ExtractElementInfo(vnode)
	sidebar := NewSidebar()
	formatted := sidebar.FormatSidebar(info)

	// Verify the formatted text contains the path (the sidebar shows it with header + indented value)
	if !strings.Contains(formatted, "base[0]/vstack[0]/panel[0]") {
		t.Errorf("Sidebar format missing expected path.\nGot:\n%s\n\nExpected to contain: base[0]/vstack[0]/panel[0]", formatted)
	}

	// Verify it contains "Path" header (with box drawing characters)
	if !strings.Contains(formatted, "Path") {
		t.Errorf("Sidebar format missing Path header.\nGot:\n%s", formatted)
	}

	// Verify the new slash-based format is used
	if !strings.Contains(formatted, "[0]/") {
		t.Errorf("Sidebar format should use slash-based path format.\nGot:\n%s", formatted)
	}

	// Verify it doesn't use the old dot-notation format (like vstack[0].bordered[0])
	if strings.Contains(formatted, "[0].bordered[0]") ||
	   strings.Contains(formatted, "[0].hstack[0]") ||
	   strings.Contains(formatted, "[0].vstack[0]") {
		t.Errorf("Sidebar format contains old dot-notation path format.\nGot:\n%s", formatted)
	}
}
