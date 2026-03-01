// Package render tests for VNodeRenderer implementations.
package render

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/paint"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// TestDeclarativeNode_GetRenderer tests GetRenderer method
func TestDeclarativeNode_GetRenderer(t *testing.T) {
	t.Run("pipeline renderer", func(t *testing.T) {
		node := NewDeclarativeNodeFromFunc(func() rtui.VNode {
			return rtui.Element("text").Prop("content", "test").Build()
		})

		renderer := node.GetRenderer()
		if renderer == nil {
			t.Fatal("GetRenderer() should not return nil")
		}

		_, ok := renderer.(*PipelineRendererAdapter)
		if !ok {
			t.Errorf("Expected *PipelineRendererAdapter, got %T", renderer)
		}
	})
}

// TestVNodeRendererInterface verifies renderers implement the interface
func TestVNodeRendererInterface(t *testing.T) {
	// Verify that PipelineRendererAdapter implements VNodeRenderer
	var _ rtui.VNodeRenderer = &PipelineRendererAdapter{}
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
