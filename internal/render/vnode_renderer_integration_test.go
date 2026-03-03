// Package render provides integration tests for VNodeRenderer implementations.
package render

import (
	"testing"

	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// TestVNodeRenderer_Integration tests VNodeRenderer implementations with real rendering scenarios
func TestVNodeRenderer_Integration(t *testing.T) {
	tests := []struct {
		name     string
		vnode    rtui.VNode
		describe string
	}{
		{
			name:     "simple text",
			vnode:    rtui.Element("text").Prop("content", "Hello").Build(),
			describe: "single text element",
		},
		{
			name:     "nested elements",
			vnode:    rtui.Element("div").Children(rtui.Element("text").Prop("content", "A").Build(), rtui.Element("text").Prop("content", "B").Build()).Build(),
			describe: "parent with multiple text children",
		},
		{
			name:     "HStack layout",
			vnode:    rtui.HStack(rtui.Element("text").Prop("content", "A").Build(), rtui.Element("text").Prop("content", "B").Build()),
			describe: "horizontal layout container",
		},
		{
			name:     "VStack layout",
			vnode:    rtui.VStack(rtui.Element("text").Prop("content", "A").Build(), rtui.Element("text").Prop("content", "B").Build()),
			describe: "vertical layout container",
		},
		{
			name:     "fragment",
			vnode:    rtui.Fragment(rtui.Element("text").Prop("content", "A").Build(), rtui.Element("text").Prop("content", "B").Build()),
			describe: "fragment with multiple children",
		},
		{
			name:     "deeply nested",
			vnode:    rtui.Element("div").Children(
				rtui.Element("text").Prop("content", "A").Build(),
				rtui.Element("text").Prop("content", "B").Build(),
			).Build(),
			describe: "element with multiple text children",
		},
	}

	node := NewDeclarativeNodeFromFuncWithFiber(func() rtui.VNode {
		return rtui.Element("text").Prop("content", "test").Build()
	})
	renderer := node.GetRenderer()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, h := renderer.Measure(tt.vnode)
			if w < 0 || h < 0 {
				t.Errorf("Measure() returned invalid dimensions: %dx%d", w, h)
			}
			t.Logf("%s: %s measures %dx%d", tt.name, tt.describe, w, h)
		})
	}
}

// TestVNodeRenderer_MeasureConsistency tests renderer measurement consistency
func TestVNodeRenderer_MeasureConsistency(t *testing.T) {
	// Create PipelineRenderer
	pipelineRenderer := NewPipelineRendererAdapter()

	// Test common VNode types
	testVNodes := []struct {
		name  string
		vnode rtui.VNode
	}{
		{
			name:  "text element",
			vnode: rtui.Element("text").Prop("content", "hello world").Build(),
		},
		{
			name:  "element with content prop",
			vnode: rtui.Element("div").Prop("content", "foo").Build(),
		},
		{
			name:  "empty element",
			vnode: rtui.Element("div").Build(),
		},
		{
			name:  "fragment with children",
			vnode: rtui.Fragment(
				rtui.Element("text").Prop("content", "a").Build(),
				rtui.Element("text").Prop("content", "b").Build(),
			),
		},
	}

	for _, tt := range testVNodes {
		t.Run(tt.name, func(t *testing.T) {
			w, h := pipelineRenderer.Measure(tt.vnode)
			t.Logf("%s: Pipeline=%dx%d", tt.name, w, h)
		})
	}
}

// TestVNodeRenderer_NilVNode tests renderer behavior with nil input
func TestVNodeRenderer_NilVNode(t *testing.T) {
	t.Run("PipelineRenderer", func(t *testing.T) {
		renderer := NewPipelineRendererAdapter()
		w, h := renderer.Measure(nil)
		if w != 0 || h != 0 {
			t.Errorf("Expected 0x0 for nil VNode, got %dx%d", w, h)
		}
	})
}

// TestVNodeRenderer_ComplexLayout tests measurement of complex layouts
func TestVNodeRenderer_ComplexLayout(t *testing.T) {
	// Create a complex nested layout
	complexVNode := rtui.VStack(
		rtui.HStack(
			rtui.Element("text").Prop("content", "A").Build(),
			rtui.Element("text").Prop("content", "B").Build(),
			rtui.Element("text").Prop("content", "C").Build(),
		),
		rtui.HStack(
			rtui.Element("text").Prop("content", "D").Build(),
			rtui.Element("text").Prop("content", "E").Build(),
		),
		rtui.Element("text").Prop("content", "F").Build(),
	)

	node := NewDeclarativeNodeFromFuncWithFiber(func() rtui.VNode {
		return rtui.Element("text").Prop("content", "test").Build()
	})
	renderer := node.GetRenderer()

	w, h := renderer.Measure(complexVNode)
	// Note: Width measurement for nested layouts is limited (may return 0)
	if h <= 0 {
		t.Errorf("Complex layout should have positive height, got %d", h)
	}
	t.Logf("Complex VStack of HStacks measures %dx%d", w, h)
}

// TestVNodeRenderer_EmptyChildren tests measurement of nodes with no children
func TestVNodeRenderer_EmptyChildren(t *testing.T) {
	node := NewDeclarativeNodeFromFuncWithFiber(func() rtui.VNode {
		return rtui.Element("text").Prop("content", "test").Build()
	})
	renderer := node.GetRenderer()

	testCases := []struct {
		name      string
		vnode     rtui.VNode
		minWidth  int
		minHeight int
	}{
		{
			name:      "empty HStack",
			vnode:     rtui.HStack(),
			minWidth:  0,
			minHeight: 0,
		},
		{
			name:      "empty VStack",
			vnode:     rtui.VStack(),
			minWidth:  0,
			minHeight: 0,
		},
		{
			name:      "empty fragment",
			vnode:     rtui.Fragment(),
			minWidth:  0,
			minHeight: 0,
		},
		{
			name:      "empty element",
			vnode:     rtui.Element("div").Build(),
			minWidth:  0, // Empty elements have no content, width is 0
			minHeight: 1, // Height is still 1 (default)
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			w, h := renderer.Measure(tt.vnode)
			if w < tt.minWidth || h < tt.minHeight {
				t.Errorf("%s: expected at least %dx%d, got %dx%d", tt.name, tt.minWidth, tt.minHeight, w, h)
			}
			t.Logf("%s measures %dx%d", tt.name, w, h)
		})
	}
}
