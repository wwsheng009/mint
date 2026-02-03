// Package render tests for VNodeRenderer implementations.
package render

import (
	"testing"

	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// TestNonFiberRenderer_Render tests rendering with NonFiberRenderer
func TestNonFiberRenderer_Render(t *testing.T) {
	node := NewDeclarativeNodeFromFunc(func() rtui.VNode {
		return rtui.Element("text").Prop("content", "Hello, World!").Build()
	})

	renderer, ok := node.GetRenderer().(*NonFiberRenderer)
	if !ok {
		t.Fatalf("Expected *NonFiberRenderer, got %T", node.GetRenderer())
	}

	if renderer == nil {
		t.Fatal("Renderer should not be nil")
	}

	if renderer.owner != node {
		t.Error("Renderer owner should be the DeclarativeNode")
	}
}

// TestNonFiberRenderer_Measure tests measuring with NonFiberRenderer
func TestNonFiberRenderer_Measure(t *testing.T) {
	tests := []struct {
		name     string
		vnode    rtui.VNode
		wantW    int
		wantH    int
	}{
		{
			name:  "text node",
			vnode: rtui.Element("text").Prop("content", "hello").Build(),
			wantW: 5,
			wantH: 1,
		},
		{
			name:  "empty text",
			vnode: rtui.Element("text").Prop("content", "").Build(),
			wantW: 0,
			wantH: 1,
		},
		{
			name:  "button",
			vnode: rtui.Element("button").Prop("label", "OK").Build(),
			wantW: 4, // [OK] = 2 + 2 brackets
			wantH: 1,
		},
		{
			name:  "nil vnode",
			vnode: nil,
			wantW: 0,
			wantH: 0,
		},
	}

	node := NewDeclarativeNodeFromFunc(func() rtui.VNode {
		return rtui.Element("text").Prop("content", "test").Build()
	})
	renderer := node.GetRenderer()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, h := renderer.Measure(tt.vnode)
			if w != tt.wantW {
				t.Errorf("Measure() width = %v, want %v", w, tt.wantW)
			}
			if h != tt.wantH {
				t.Errorf("Measure() height = %v, want %v", h, tt.wantH)
			}
		})
	}
}

// TestFiberRenderer_Measure tests measuring with FiberRenderer
func TestFiberRenderer_Measure(t *testing.T) {
	renderer := NewFiberRenderer(nil)

	tests := []struct {
		name     string
		vnode    rtui.VNode
		wantW    int
		wantH    int
	}{
		{
			name:  "text node",
			vnode: rtui.Element("text").Prop("content", "hello").Build(),
			wantW: 5,
			wantH: 1,
		},
		{
			name:  "element with label prop (not a real button)",
			vnode: rtui.Element("button").Prop("label", "OK").Build(),
			wantW: 1, // Basic element without Label() interface returns default width
			wantH: 1,
		},
		{
			name:  "nil vnode",
			vnode: nil,
			wantW: 0,
			wantH: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, h := renderer.Measure(tt.vnode)
			if w != tt.wantW {
				t.Errorf("Measure() width = %v, want %v", w, tt.wantW)
			}
			if h != tt.wantH {
				t.Errorf("Measure() height = %v, want %v", h, tt.wantH)
			}
		})
	}
}

// TestDeclarativeNode_GetRenderer tests GetRenderer method
func TestDeclarativeNode_GetRenderer(t *testing.T) {
	t.Run("non-Fiber mode", func(t *testing.T) {
		node := NewDeclarativeNodeFromFunc(func() rtui.VNode {
			return rtui.Element("text").Prop("content", "test").Build()
		})

		renderer := node.GetRenderer()
		if renderer == nil {
			t.Fatal("GetRenderer() should not return nil")
		}

		_, ok := renderer.(*NonFiberRenderer)
		if !ok {
			t.Errorf("Expected *NonFiberRenderer in non-Fiber mode, got %T", renderer)
		}
	})
}

// TestVNodeRendererInterface verifies both renderers implement the interface
func TestVNodeRendererInterface(t *testing.T) {
	// This test verifies that both renderer types implement VNodeRenderer
	var _ rtui.VNodeRenderer = &NonFiberRenderer{}
	var _ rtui.VNodeRenderer = &FiberRenderer{}
}
