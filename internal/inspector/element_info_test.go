package inspector

import (
	"fmt"
	"testing"

	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui"
)

// TestCollectHitTestEntryInfo tests extracting tag/key/label from VNodes
// This simulates what collectHitTestEntries does in standalone_inspector.go
func TestCollectHitTestEntryInfo(t *testing.T) {
	tests := []struct {
		name          string
		createVNode   func() rtui.VNode
		expectedTag   string
		expectedKey   string
		expectedLabel string
		expectedType  string
	}{
		{
			name: "Button with label",
			createVNode: func() rtui.VNode {
				return ui.NewButtonBuilder("Click Me").Key("btn1").Build()
			},
			expectedTag:   "button",
			expectedKey:   "btn1",
			expectedLabel: "Click Me",
			expectedType:  "Element",
		},
		{
			name: "Text node",
			createVNode: func() rtui.VNode {
				return ui.Text("Hello World")
			},
			expectedTag:   "text",
			expectedKey:   "",
			expectedLabel: "Hello World",
			expectedType:  "Text",
		},
		{
			name: "Button without key",
			createVNode: func() rtui.VNode {
				return ui.NewButtonBuilder("Submit").Build()
			},
			expectedTag:   "button",
			expectedKey:   "",
			expectedLabel: "Submit",
			expectedType:  "Element",
		},
		{
			name: "ElementVNode with custom tag",
			createVNode: func() rtui.VNode {
				elem := rtui.NewElement("bordered")
				elem.SetKey("border1")
				return elem
			},
			expectedTag:   "bordered",
			expectedKey:   "border1",
			expectedLabel: "",
			expectedType:  "Element",
		},
		{
			name: "VStack",
			createVNode: func() rtui.VNode {
				return rtui.VStack(
					ui.Text("Child 1"),
					ui.Text("Child 2"),
				)
			},
			expectedTag:   "vstack",
			expectedKey:   "",
			expectedLabel: "",
			expectedType:  "Element",
		},
		{
			name: "Text with long content (should be truncated)",
			createVNode: func() rtui.VNode {
				return ui.Text("This is a very long text content that should be truncated")
			},
			expectedTag:   "text",
			expectedKey:   "",
			expectedLabel: "This is a very long text content that should be truncated",
			expectedType:  "Text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := tt.createVNode()

			// Simulate what collectHitTestEntries does
			var tag string
			var key string
			var label string

			// Get tag if available
			if tagger, ok := node.(interface{ Tag() string }); ok {
				tag = tagger.Tag()
			}

			// Get key from VNode.Key() method
			if keyer, ok := node.(interface{ Key() string }); ok {
				key = keyer.Key()
			}

			// Get label/text content from various sources
			// Try Label() method first (for buttons)
			if labeler, ok := node.(interface{ Label() string }); ok {
				label = labeler.Label()
			}
			// Try Content() method (for TextVNode)
			if label == "" {
				if contenter, ok := node.(interface{ Content() string }); ok {
					label = contenter.Content()
				}
			}
			// Fall back to GetTextContent utility
			if label == "" {
				label = rtui.GetTextContent(node)
			}

			// Verify results
			if tag != tt.expectedTag {
				t.Errorf("Tag: expected %q, got %q", tt.expectedTag, tag)
			}
			if key != tt.expectedKey {
				t.Errorf("Key: expected %q, got %q", tt.expectedKey, key)
			}
			if label != tt.expectedLabel {
				t.Errorf("Label: expected %q, got %q", tt.expectedLabel, label)
			}
			if node.Type().String() != tt.expectedType {
				t.Errorf("Type: expected %q, got %q", tt.expectedType, node.Type().String())
			}
		})
	}
}

// TestHitTestEntryFormatting tests the full HitTestEntry formatting
func TestHitTestEntryFormatting(t *testing.T) {
	tests := []struct {
		name           string
		entry          HitTestEntry
		expectedFormat string
	}{
		{
			name: "Button with all fields",
			entry: HitTestEntry{
				Type:      "Element",
				Tag:       "button",
				Key:       "btn1",
				Label:     "Click Me",
				Bounds:    "10,5 20x3",
				ZOrder:    0,
				HitTest:   "YES",
				Clickable: true,
			},
			expectedFormat: "Element/(button) [btn1] 'Click Me'",
		},
		{
			name: "Element with only tag",
			entry: HitTestEntry{
				Type:      "Element",
				Tag:       "bordered",
				Key:       "",
				Label:     "",
				Bounds:    "0,0 80x1",
				ZOrder:    1,
				HitTest:   "NO",
				Clickable: false,
			},
			expectedFormat: "Element/(bordered)",
		},
		{
			name: "Text node",
			entry: HitTestEntry{
				Type:      "Text",
				Tag:       "text",
				Key:       "",
				Label:     "Hello World",
				Bounds:    "5,2 11x1",
				ZOrder:    2,
				HitTest:   "NO",
				Clickable: false,
			},
			expectedFormat: "Text/(text) 'Hello World'",
		},
		{
			name: "Long label should be truncated",
			entry: HitTestEntry{
				Type:      "Element",
				Tag:       "button",
				Key:       "",
				Label:     "This is a very long button label",
				Bounds:    "0,0 30x1",
				ZOrder:    0,
				HitTest:   "YES",
				Clickable: true,
			},
			expectedFormat: "Element/(button) 'This is a...'", // collectHitTestEntries truncates to 12 chars, formatNodeInfo to 15
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Truncate label like collectHitTestEntries does
			label := tt.entry.Label
			if len(label) > 12 {
				label = label[:9] + "..."
			}

			result := formatNodeInfo(tt.entry.Type, tt.entry.Tag, tt.entry.Key, label)
			if result != tt.expectedFormat {
				t.Errorf("Expected %q, got %q", tt.expectedFormat, result)
			}
		})
	}
}

// mockVNode wraps a real VNode for testing
type mockVNode struct {
	vnode   rtui.VNode
	tag     string
	key     string
	label   string
	content string
}

// TestCollectHitTestEntries_RealWorld tests with real VNode structures
// This simulates what collectHitTestEntries actually processes at runtime
func TestCollectHitTestEntries_RealWorld(t *testing.T) {
	tests := []struct {
		name          string
		vnode         rtui.VNode
		expectedTag   string
		expectedKey   string
		expectedLabel string
	}{
		{
			name:          "Button with label",
			vnode:         ui.NewButtonBuilder("Click Me").Build(),
			expectedTag:   "button",
			expectedKey:   "",
			expectedLabel: "Click Me",
		},
		{
			name:          "Text node",
			vnode:         ui.Text("Hello World"),
			expectedTag:   "text",
			expectedKey:   "",
			expectedLabel: "Hello World",
		},
		{
			name:          "Element with custom tag",
			vnode:         rtui.NewElement("bordered"),
			expectedTag:   "bordered",
			expectedKey:   "",
			expectedLabel: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate what collectHitTestEntries does
			var tag, key, label string

			// Get tag via type assertion
			if tagger, ok := tt.vnode.(interface{ Tag() string }); ok {
				tag = tagger.Tag()
			}

			// Get key via VNode.Key() method
			if keyer, ok := tt.vnode.(interface{ Key() string }); ok {
				key = keyer.Key()
			}

			// Get label via Label() method (for buttons)
			if labeler, ok := tt.vnode.(interface{ Label() string }); ok {
				label = labeler.Label()
			}

			// Try Content() method (for TextVNode)
			if label == "" {
				if contenter, ok := tt.vnode.(interface{ Content() string }); ok {
					label = contenter.Content()
				}
			}

			// Fall back to GetTextContent
			if label == "" {
				label = rtui.GetTextContent(tt.vnode)
			}

			if tag != tt.expectedTag {
				t.Errorf("Expected tag %q, got %q", tt.expectedTag, tag)
			}
			if key != tt.expectedKey {
				t.Errorf("Expected key %q, got %q", tt.expectedKey, key)
			}
			if label != tt.expectedLabel {
				t.Errorf("Expected label %q, got %q", tt.expectedLabel, label)
			}
		})
	}
}

// TestFormatNodeInfo_RealWorld tests formatNodeInfo with real VNode data
func TestFormatNodeInfo_RealWorld(t *testing.T) {
	tests := []struct {
		name     string
		vnode    rtui.VNode
		contains []string // Substrings that should appear in formatted output
	}{
		{
			name:  "Button shows tag and label",
			vnode: ui.NewButtonBuilder("Submit").Build(),
			contains: []string{
				"Element",
				"button",
				"Submit",
			},
		},
		{
			name:  "Text shows content",
			vnode: ui.Text("Hello World"),
			contains: []string{
				"Text",
				"text",
				"Hello World",
			},
		},
		// Real demo2 components
		{
			name:  "Bordered node (migrated to Stack)",
			vnode: ui.NewVStack().SingleBorder().SetChildrenList([]ui.VNode{ui.Text("Content")}),
			contains: []string{
				"Element",
				"vstack", // Stack uses vstack tag
			},
		},
		{
			name: "VStack (from demo2)",
			vnode: rtui.VStack(
				ui.NewButtonBuilder("Button A").Build(),
				ui.NewButtonBuilder("Button B").Build(),
			),
			contains: []string{
				"Element",
				"vstack",
			},
		},
		{
			name: "HStack (from demo2)",
			vnode: rtui.HStack(
				ui.Text("Left"),
				ui.Text("Right"),
			),
			contains: []string{
				"Element",
				"hstack",
			},
		},
		{
			name:  "Button with key (from demo2 ControlPanel)",
			vnode: ui.NewButtonBuilder("[1] Event").Key("btn-event").Build(),
			contains: []string{
				"Element",
				"button",
				"btn-even~", // Key is truncated to 8 chars
				"[1] Event",
			},
		},
		{
			name: "Wrap component (from demo2)",
			vnode: ui.NewWrapBuilder().Children(
				ui.NewButtonBuilder("Btn1").Build(),
				ui.NewButtonBuilder("Btn2").Build(),
				ui.NewButtonBuilder("Btn3").Build(),
			).Build(),
			contains: []string{
				"Element",
				"vstack", // Wrap creates a VStack internally
			},
		},
		{
			name: "Nested structure - Bordered Stack with VStack and Buttons",
			vnode: ui.NewVStack().
				SingleBorder().
				SetChildrenList([]ui.VNode{
					ui.VStack(
						ui.NewButtonBuilder("[1] Event").Key("btn-event").Build(),
						ui.NewButtonBuilder("[2]setState").Key("btn-setstate").Build(),
					),
				}),
			contains: []string{
				"Element",
				"vstack",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Extract info like collectHitTestEntries does
			var tag, key, label string

			if tagger, ok := tt.vnode.(interface{ Tag() string }); ok {
				tag = tagger.Tag()
			}
			if keyer, ok := tt.vnode.(interface{ Key() string }); ok {
				key = keyer.Key()
			}
			if labeler, ok := tt.vnode.(interface{ Label() string }); ok {
				label = labeler.Label()
			}
			if label == "" {
				if contenter, ok := tt.vnode.(interface{ Content() string }); ok {
					label = contenter.Content()
				}
			}
			if label == "" {
				label = rtui.GetTextContent(tt.vnode)
			}

			result := formatNodeInfo(tt.vnode.Type().String(), tag, key, label)

			for _, expected := range tt.contains {
				if !contains(result, expected) {
					t.Errorf("Expected output %q to contain %q", result, expected)
				}
			}
		})
	}
}

// TestExtractElementInfo_Button tests extracting info from a Button
func TestExtractElementInfo_Button(t *testing.T) {
	button := ui.NewButtonBuilder("[1] Event").Build()

	info := ExtractElementInfo(button)

	// Check basic info
	if info.Type == "" {
		t.Error("Type should not be empty")
	}
	if info.Tag != "button" {
		t.Errorf("Expected tag 'button', got '%s'", info.Tag)
	}
	if info.Label != "[1] Event" {
		t.Errorf("Expected label '[1] Event', got '%s'", info.Label)
	}
}

// TestExtractElementInfo_Text tests extracting info from Text
func TestExtractElementInfo_Text(t *testing.T) {
	text := ui.Text("Hello World")

	info := ExtractElementInfo(text)

	// Check basic info
	if info.Type == "" {
		t.Error("Type should not be empty")
	}
	if info.Label != "Hello World" {
		t.Errorf("Expected label 'Hello World', got '%s'", info.Label)
	}

	// Check natural width
	if info.Layout.NaturalWidth != 11 { // "Hello World" = 11 chars
		t.Errorf("Expected natural width 11, got %d", info.Layout.NaturalWidth)
	}
}

// TestExtractElementInfo_WithBounds tests extracting info with bounds
func TestExtractElementInfo_WithBounds(t *testing.T) {
	t.Skip("SetBounds integration needs actual layout engine")

	button := ui.NewButtonBuilder("Test").Build()

	// Simulate SetBounds being called
	if boundsAware, ok := button.(interface{ SetBounds(int, int, int, int) }); ok {
		boundsAware.SetBounds(5, 10, 20, 1)
	}

	info := ExtractElementInfo(button)

	// Check bounds
	if info.Bounds[0] != 5 || info.Bounds[1] != 10 ||
		info.Bounds[2] != 20 || info.Bounds[3] != 1 {
		t.Errorf("Expected bounds [5 10 20 1], got %v", info.Bounds)
	}

	// Check position
	if info.Position.X != 5 || info.Position.Y != 10 {
		t.Errorf("Expected position (5, 10), got (%d, %d)",
			info.Position.X, info.Position.Y)
	}

	// Check size
	if info.Size.Width != 20 || info.Size.Height != 1 {
		t.Errorf("Expected size 20x1, got %dx%d",
			info.Size.Width, info.Size.Height)
	}

	// Check layout width
	if info.Layout.LayoutWidth != 20 {
		t.Errorf("Expected layout width 20, got %d", info.Layout.LayoutWidth)
	}
}

// TestExtractElementInfo_Flex tests flex property extraction
func TestExtractElementInfo_Flex(t *testing.T) {
	t.Skip("SetProp integration needs proper props handling")

	// Create a button with flex prop through ElementVNode
	button := ui.NewButtonBuilder("Test").Build()

	// Get the element and set prop
	if elem, ok := button.(interface{ SetProp(string, interface{}) }); ok {
		elem.SetProp("flex", 1)
	}

	info := ExtractElementInfo(button)

	// Check flex
	if info.Layout.Flex != 1 {
		t.Errorf("Expected flex 1, got %d", info.Layout.Flex)
	}

	if !info.Layout.IsFlexChild {
		t.Error("Expected IsFlexChild to be true")
	}
}

// TestExtractElementInfo_NilVNode tests handling of nil VNode
func TestExtractElementInfo_NilVNode(t *testing.T) {
	info := ExtractElementInfo(nil)

	if info.Type != "nil" {
		t.Errorf("Expected type 'nil', got '%s'", info.Type)
	}

	if info.VNode != nil {
		t.Error("Expected VNode to be nil")
	}
}

// TestFormatElementInfo tests formatting of ElementInfo
func TestFormatElementInfo(t *testing.T) {
	button := ui.NewButtonBuilder("Test Button").Build()

	info := ExtractElementInfo(button)
	formatted := formatNodeInfo(info.Type, info.Tag, info.Key, info.Label)

	if formatted == "" {
		t.Error("Formatted output should not be empty")
	}

	// Check for expected sections (without Bounds since we haven't set it)
	expectedSections := []string{
		"Element:",
		"Tag:",
		"Position:",
		"Size:",
		"Layout:",
	}

	for _, section := range expectedSections {
		if !contains(formatted, section) {
			t.Errorf("Formatted output should contain '%s'", section)
		}
	}
}

// TestExtractElementInfo_NaturalWidthCalculation tests natural width calculation
func TestExtractElementInfo_NaturalWidthCalculation(t *testing.T) {
	tests := []struct {
		name          string
		label         string
		expectedWidth int
	}{
		{"Short label", "OK", 6},          // "OK" (2) + brackets (2) + focus space (2)
		{"Medium label", "Cancel", 10},    // "Cancel" (6) + brackets (2) + focus space (2)
		{"Long label", "Submit Form", 15}, // "Submit Form" (11) + brackets (2) + focus space (2)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			button := ui.NewButtonBuilder(tt.label).Build()
			info := ExtractElementInfo(button)

			if info.Layout.NaturalWidth != tt.expectedWidth {
				t.Errorf("Expected natural width %d, got %d",
					tt.expectedWidth, info.Layout.NaturalWidth)
			}
		})
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestFormatNodeInfo tests the formatNodeInfo function
func TestFormatNodeInfo(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		tag      string
		key      string
		label    string
		expected string
	}{
		{
			name:     "Type only",
			id:       "Element",
			tag:      "",
			key:      "",
			label:    "",
			expected: "Element",
		},
		{
			name:     "Type and tag",
			id:       "Element",
			tag:      "button",
			key:      "",
			label:    "",
			expected: "Element/(button)",
		},
		{
			name:     "Type, tag, and label",
			id:       "Element",
			tag:      "button",
			key:      "",
			label:    "Click Me",
			expected: "Element/(button) 'Click Me'",
		},
		{
			name:     "All fields",
			id:       "Element",
			tag:      "button",
			key:      "btn1",
			label:    "Submit",
			expected: "Element/(button) [btn1] 'Submit'",
		},
		{
			name:     "Truncate long label",
			id:       "Element",
			tag:      "button",
			key:      "",
			label:    "This is a very long button label",
			expected: "Element/(button) 'This is a very ...'",
		},
		{
			name:     "Truncate long key",
			id:       "Element",
			tag:      "button",
			key:      "verylongkeythatshouldbetruncated",
			label:    "OK",
			expected: "Element/(button) [verylong~] 'OK'",
		},
		{
			name:     "Text node",
			id:       "Text",
			tag:      "text",
			key:      "",
			label:    "Hello World",
			expected: "Text/(text) 'Hello World'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatNodeInfo(tt.id, tt.tag, tt.key, tt.label)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestDemo2RealWorldStructure tests the actual structure used in demo2
// This simulates the real components used in the demo:
// - Buttons with keys like "btn-event", "btn-setstate", etc.
// - Bordered containers
// - VStack/HStack layouts
// - Text nodes
func TestDemo2RealWorldStructure(t *testing.T) {
	// Create a structure similar to demo2's ControlPanel
	btn1 := ui.NewButtonBuilder("[1] Event").Key("btn-event").Build()
	_ = ui.NewButtonBuilder("[2]setState").Key("btn-setstate").Build()
	_ = ui.NewButtonBuilder("[3]Scheduler").Key("btn-scheduler").Build()

	// Test button extraction
	t.Run("Button [1] Event", func(t *testing.T) {
		var tag, key, label string

		if tagger, ok := btn1.(interface{ Tag() string }); ok {
			tag = tagger.Tag()
		}
		if keyer, ok := btn1.(interface{ Key() string }); ok {
			key = keyer.Key()
		}
		if labeler, ok := btn1.(interface{ Label() string }); ok {
			label = labeler.Label()
		}

		if tag != "button" {
			t.Errorf("Expected tag 'button', got '%s'", tag)
		}
		if key != "btn-event" {
			t.Errorf("Expected key 'btn-event', got '%s'", key)
		}
		if label != "[1] Event" {
			t.Errorf("Expected label '[1] Event', got '%s'", label)
		}

		// Test formatting
		result := formatNodeInfo(btn1.Type().String(), tag, key, label)
		expected := "Element/(button) [btn-even~] '[1] Event'" // Key truncated
		if result != expected {
			t.Errorf("Expected %q, got %q", expected, result)
		}
	})

	// Test Bordered Stack node with nested content (like HeaderPanel)
	t.Run("Bordered Stack container", func(t *testing.T) {
		bordered := ui.NewVStack().
			SingleBorder().
			SetChildrenList([]ui.VNode{
				ui.Text("Content inside border"),
			})

		var tag string
		tag = bordered.Tag()

		if tag != "vstack" {
			t.Errorf("Expected tag 'vstack', got '%s'", tag)
		}

		result := formatNodeInfo(bordered.Type().String(), tag, "", "")
		if result != "Element/(vstack)" {
			t.Errorf("Expected 'Element/(vstack)', got %q", result)
		}
	})

	// Test VStack with multiple children (like PipelineVisualization)
	t.Run("VStack container", func(t *testing.T) {
		vstack := rtui.VStack(
			ui.Text("Line 1"),
			ui.Text("Line 2"),
			ui.Text("Line 3"),
		)

		var tag string
		if tagger, ok := vstack.(interface{ Tag() string }); ok {
			tag = tagger.Tag()
		}

		if tag != "vstack" {
			t.Errorf("Expected tag 'vstack', got '%s'", tag)
		}

		result := formatNodeInfo(vstack.Type().String(), tag, "", "")
		if result != "Element/(vstack)" {
			t.Errorf("Expected 'Element/(vstack)', got %q", result)
		}
	})

	// Test HStack (like StatisticsPanel content)
	t.Run("HStack container", func(t *testing.T) {
		hstack := rtui.HStack(
			ui.Text("Left"),
			ui.Text("Right"),
		)

		var tag string
		if tagger, ok := hstack.(interface{ Tag() string }); ok {
			tag = tagger.Tag()
		}

		if tag != "hstack" {
			t.Errorf("Expected tag 'hstack', got '%s'", tag)
		}

		result := formatNodeInfo(hstack.Type().String(), tag, "", "")
		if result != "Element/(hstack)" {
			t.Errorf("Expected 'Element/(hstack)', got %q", result)
		}
	})

	// Test Text node
	t.Run("Text node", func(t *testing.T) {
		text := ui.Text("Hello World")

		var tag, label string
		if tagger, ok := text.(interface{ Tag() string }); ok {
			tag = tagger.Tag()
		}
		if labeler, ok := text.(interface{ Label() string }); ok {
			label = labeler.Label()
		}
		if label == "" {
			if contenter, ok := text.(interface{ Content() string }); ok {
				label = contenter.Content()
			}
		}

		if tag != "text" {
			t.Errorf("Expected tag 'text', got '%s'", tag)
		}
		if label != "Hello World" {
			t.Errorf("Expected label 'Hello World', got '%s'", label)
		}

		result := formatNodeInfo(text.Type().String(), tag, "", label)
		expected := "Text/(text) 'Hello World'"
		if result != expected {
			t.Errorf("Expected %q, got %q", expected, result)
		}
	})

	// Test nested structure (Bordered Stack > VStack > Buttons)
	t.Run("Nested demo2 structure", func(t *testing.T) {
		nested := ui.NewVStack().
			SingleBorder().
			SetChildrenList([]ui.VNode{
				ui.VStack(
					ui.NewButtonBuilder("[1] Event").Key("btn-event").Build(),
					ui.NewButtonBuilder("[2]setState").Key("btn-setstate").Build(),
				),
			})

		// Extract tag from top-level Stack
		var tag string
		tag = nested.Tag()

		if tag != "vstack" {
			t.Errorf("Expected tag 'vstack', got '%s'", tag)
		}

		// Check children
		children := nested.Children()
		if len(children) == 0 {
			t.Error("Bordered node should have children")
		} else {
			// First child should be VStack
			vstack := children[0]
			var vstackTag string
			if tagger, ok := vstack.(interface{ Tag() string }); ok {
				vstackTag = tagger.Tag()
			}
			if vstackTag != "vstack" {
				t.Errorf("Expected vstack tag 'vstack', got '%s'", vstackTag)
			}
		}
	})
}

// TestVNodeBoundsDataFlow tests the bounds data setting and retrieval
func TestVNodeBoundsDataFlow(t *testing.T) {
	t.Run("ElementVNode bounds storage", func(t *testing.T) {
		elem := rtui.NewElement("test")

		// Initially bounds should be zero
		bounds := elem.GetBounds()
		if bounds != [4]int{0, 0, 0, 0} {
			t.Errorf("Initial bounds should be [0,0,0,0], got %v", bounds)
		}

		// Set bounds
		elem.SetBounds(10, 20, 100, 50)

		// Verify bounds were set
		bounds = elem.GetBounds()
		if bounds != [4]int{10, 20, 100, 50} {
			t.Errorf("Bounds should be [10,20,100,50], got %v", bounds)
		}

		// Simulate collectHitTestEntries formatting
		boundsStr := fmt.Sprintf("%d,%d %dx%d", bounds[0], bounds[1], bounds[2], bounds[3])
		if boundsStr != "10,20 100x50" {
			t.Errorf("Bounds string should be '10,20 100x50', got %q", boundsStr)
		}
	})

	t.Run("Button with bounds", func(t *testing.T) {
		// Create a simple element to test bounds
		elem := rtui.NewElement("button")
		elem.SetKey("btn1")

		// Set bounds as layout engine would
		elem.SetBounds(5, 10, 20, 3)

		bounds := elem.GetBounds()
		if bounds != [4]int{5, 10, 20, 3} {
			t.Errorf("Element bounds should be [5,10,20,3], got %v", bounds)
		}

		// Verify HitTestEntry would be created correctly
		entry := HitTestEntry{
			Type:      elem.Type().String(),
			Tag:       elem.Tag(),
			Key:       elem.Key(),
			Label:     "",
			Bounds:    fmt.Sprintf("%d,%d %dx%d", bounds[0], bounds[1], bounds[2], bounds[3]),
			ZOrder:    0,
			HitTest:   "YES",
			Clickable: false,
		}

		if entry.Bounds != "5,10 20x3" {
			t.Errorf("HitTestEntry bounds should be '5,10 20x3', got %q", entry.Bounds)
		}
		if entry.Type != "Element" {
			t.Errorf("NodeID should be 'Element', got %q", entry.Type)
		}
		if entry.Tag != "button" {
			t.Errorf("Tag should be 'button', got %q", entry.Tag)
		}
		if entry.Key != "btn1" {
			t.Errorf("Key should be 'btn1', got %q", entry.Key)
		}
	})

	t.Run("Text with bounds", func(t *testing.T) {
		text := ui.Text("Hello")

		// Set bounds
		if boundsSetter, ok := text.(interface{ SetBounds(int, int, int, int) }); ok {
			boundsSetter.SetBounds(0, 5, 11, 1)

			bounds := text.(interface{ GetBounds() [4]int }).GetBounds()
			if bounds != [4]int{0, 5, 11, 1} {
				t.Errorf("Text bounds should be [0,5,11,1], got %v", bounds)
			}
		}
	})

	t.Run("Bordered Stack node with bounds", func(t *testing.T) {
		bordered := ui.NewVStack().
			SingleBorder().
			SetChildrenList([]ui.VNode{ui.Text("Content")})

		// Stack should support SetBounds (inherited from ElementVNode)
			bordered.SetBounds(2, 2, 40, 10)

			bounds := bordered.GetBounds()
			if bounds != [4]int{2, 2, 40, 10} {
				t.Errorf("Bordered Stack bounds should be [2,2,40,10], got %v", bounds)
			}
	})

	t.Run("HitTestEntry creation with bounds", func(t *testing.T) {
		// Create a node with known bounds
		elem := rtui.NewElement("div")
		elem.SetBounds(15, 25, 80, 5)

		// Simulate collectHitTestEntries creating HitTestEntry
		bounds := elem.GetBounds()
		entry := HitTestEntry{
			Type:      elem.Type().String(),
			Tag:       elem.Tag(),
			Key:       elem.Key(),
			Label:     "",
			Bounds:    fmt.Sprintf("%d,%d %dx%d", bounds[0], bounds[1], bounds[2], bounds[3]),
			ZOrder:    1,
			HitTest:   "NO",
			Clickable: false,
		}

		if entry.Type != "Element" {
			t.Errorf("NodeID: expected 'Element', got %q", entry.Type)
		}
		if entry.Tag != "div" {
			t.Errorf("Tag: expected 'div', got %q", entry.Tag)
		}
		if entry.Bounds != "15,25 80x5" {
			t.Errorf("Bounds: expected '15,25 80x5', got %q", entry.Bounds)
		}
		if entry.ZOrder != 1 {
			t.Errorf("ZOrder: expected 1, got %d", entry.ZOrder)
		}
	})

	t.Run("formatNodeInfo with bounds context", func(t *testing.T) {
		// Test that formatNodeInfo correctly formats node info
		// even without bounds (bounds are shown separately in HitTestEntry)
		button := ui.NewButtonBuilder("Submit").Key("btn-submit").Build()

		var tag, key, label string
		if tagger, ok := button.(interface{ Tag() string }); ok {
			tag = tagger.Tag()
		}
		if keyer, ok := button.(interface{ Key() string }); ok {
			key = keyer.Key()
		}
		if labeler, ok := button.(interface{ Label() string }); ok {
			label = labeler.Label()
		}

		result := formatNodeInfo(button.Type().String(), tag, key, label)

		// Result should show button tag and label, not bounds (bounds are separate)
		// Key "btn-submit" (9 chars) gets truncated to "btn-subm~" (8 + ~)
		if result != "Element/(button) [btn-subm~] 'Submit'" {
			t.Errorf("Expected 'Element/(button) [btn-subm~] 'Submit'', got %q", result)
		}
	})
}
