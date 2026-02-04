// Package render tests for VNodeRenderer implementations.
package render

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/paint"
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

// =============================================================================
// getBuffer Tests
// =============================================================================

func TestGetBuffer(t *testing.T) {
	t.Run("returns buffer for valid *paint.Buffer", func(t *testing.T) {
		buf := paint.NewBuffer(80, 24)
		result := getBuffer(buf)
		if result != buf {
			t.Error("getBuffer should return the same buffer")
		}
	})

	t.Run("returns nil for nil buffer", func(t *testing.T) {
		result := getBuffer(nil)
		if result != nil {
			t.Error("getBuffer should return nil for nil input")
		}
	})

	t.Run("returns nil for wrong type", func(t *testing.T) {
		result := getBuffer("not a buffer")
		if result != nil {
			t.Error("getBuffer should return nil for wrong type")
		}
	})
}

// =============================================================================
// NonFiberRenderer.Render Tests
// =============================================================================

func TestNonFiberRenderer_Render_Method(t *testing.T) {
	node := NewDeclarativeNodeFromFunc(func() rtui.VNode {
		return rtui.Element("text").Prop("content", "Hello").Build()
	})

	renderer, ok := node.GetRenderer().(*NonFiberRenderer)
	if !ok {
		t.Fatalf("Expected *NonFiberRenderer, got %T", node.GetRenderer())
	}

	t.Run("renders to valid buffer", func(t *testing.T) {
		buf := paint.NewBuffer(80, 24)
		vnode := rtui.Element("text").Prop("content", "Hi").Build()

		// Should not panic
		renderer.Render(vnode, 0, 0, buf)
	})

	t.Run("handles nil buffer gracefully", func(t *testing.T) {
		vnode := rtui.Element("text").Prop("content", "Hi").Build()

		// Should not panic
		renderer.Render(vnode, 0, 0, nil)
	})

	t.Run("handles wrong buffer type gracefully", func(t *testing.T) {
		vnode := rtui.Element("text").Prop("content", "Hi").Build()

		// Should not panic
		renderer.Render(vnode, 0, 0, "not a buffer")
	})
}

// =============================================================================
// FiberRenderer.Render Tests
// =============================================================================

func TestFiberRenderer_Render_Method(t *testing.T) {
	t.Run("calls renderCallback when set", func(t *testing.T) {
		called := false
		renderer := NewFiberRenderer(func(vnode rtui.VNode, x, y int, buf *paint.Buffer) {
			called = true
		})

		buf := paint.NewBuffer(80, 24)
		vnode := rtui.Element("text").Prop("content", "Hi").Build()

		renderer.Render(vnode, 0, 0, buf)

		if !called {
			t.Error("renderCallback should have been called")
		}
	})

	t.Run("handles nil buffer gracefully", func(t *testing.T) {
		called := false
		renderer := NewFiberRenderer(func(vnode rtui.VNode, x, y int, buf *paint.Buffer) {
			called = true
		})

		vnode := rtui.Element("text").Prop("content", "Hi").Build()
		renderer.Render(vnode, 0, 0, nil)

		if called {
			t.Error("renderCallback should not have been called with nil buffer")
		}
	})

	t.Run("handles nil callback gracefully", func(t *testing.T) {
		renderer := NewFiberRenderer(nil)
		buf := paint.NewBuffer(80, 24)
		vnode := rtui.Element("text").Prop("content", "Hi").Build()

		// Should not panic
		renderer.Render(vnode, 0, 0, buf)
	})
}
