// Package render tests for DeclarativeNode.
package render

import (
	"testing"

	"github.com/wwsheng009/mint/framework"
	"github.com/wwsheng009/mint/framework/component"
	"github.com/wwsheng009/mint/runtime/paint"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// DeclarativeNode Constructor Tests
// =============================================================================

// TestNewDeclarativeNode tests creating a DeclarativeNode from a VNode
func TestNewDeclarativeNode_Constructor(t *testing.T) {
	vnode := rtui.Element("div").Build()
	node := NewDeclarativeNode(vnode)

	if node == nil {
		t.Fatal("NewDeclarativeNode should not return nil")
	}

	if node.root != vnode {
		t.Error("root should be set to the provided VNode")
	}

	if node.renderFn != nil {
		t.Error("renderFn should be nil for VNode constructor")
	}
}

// TestNewDeclarativeNodeFromFunc tests creating a DeclarativeNode from a function
func TestNewDeclarativeNode_FromFunc(t *testing.T) {
	fn := func() rtui.VNode {
		return rtui.Element("div").Build()
	}
	node := NewDeclarativeNodeFromFunc(fn)

	if node == nil {
		t.Fatal("NewDeclarativeNodeFromFunc should not return nil")
	}

	if node.renderFn == nil {
		t.Error("renderFn should be set")
	}

	if node.instance == nil {
		t.Error("instance should be initialized")
	}

	if node.focusMgr == nil {
		t.Error("focusMgr should be initialized")
	}

	if node.renderer == nil {
		t.Error("renderer should be initialized")
	}

	if node.useFiber {
		t.Error("useFiber should be false by default")
	}
}

// =============================================================================
// DeclarativeNode.ID Tests
// =============================================================================

func TestDeclarativeNode_ID(t *testing.T) {
	t.Run("with key", func(t *testing.T) {
		vnode := rtui.Element("div").Key("test-key").Build()
		node := NewDeclarativeNode(vnode)

		id := node.ID()
		if id != "declarative:test-key" {
			t.Errorf("ID() = %s, want declarative:test-key", id)
		}
	})

	t.Run("without key", func(t *testing.T) {
		vnode := rtui.Element("div").Build()
		node := NewDeclarativeNode(vnode)

		id := node.ID()
		if id != "declarative:node" {
			t.Errorf("ID() = %s, want declarative:node", id)
		}
	})

	t.Run("nil VNode", func(t *testing.T) {
		node := NewDeclarativeNode(nil)

		id := node.ID()
		if id != "declarative:node" {
			t.Errorf("ID() = %s, want declarative:node", id)
		}
	})
}

// =============================================================================
// DeclarativeNode.Type Tests
// =============================================================================

func TestDeclarativeNode_Type(t *testing.T) {
	node := NewDeclarativeNode(rtui.Element("div").Build())

	if typ := node.Type(); typ != "DeclarativeNode" {
		t.Errorf("Type() = %s, want DeclarativeNode", typ)
	}
}

// =============================================================================
// DeclarativeNode.Children Tests
// =============================================================================

func TestDeclarativeNode_Children(t *testing.T) {
	node := NewDeclarativeNode(rtui.Element("div").Build())

	children := node.Children()
	if children != nil {
		t.Errorf("Children() should return nil, got %v", children)
	}
}

// =============================================================================
// DeclarativeNode.Measure Tests
// =============================================================================

func TestDeclarativeNode_Measure(t *testing.T) {
	tests := []struct {
		name      string
		vnode     rtui.VNode
		maxWidth  int
		maxHeight int
		wantW     int
		wantH     int
	}{
		{
			name:      "no constraints",
			vnode:     rtui.Element("div").Build(),
			maxWidth:  100,
			maxHeight: 100,
			wantW:     100,
			wantH:     1,
		},
		{
			name:      "with explicit width prop",
			vnode:     rtui.Element("div").Prop("width", 50).Build(),
			maxWidth:  100,
			maxHeight: 100,
			wantW:     50,
			wantH:     1,
		},
		{
			name:      "with explicit height prop",
			vnode:     rtui.Element("div").Prop("height", 25).Build(),
			maxWidth:  100,
			maxHeight: 100,
			wantW:     100,
			wantH:     25,
		},
		{
			name:      "with explicit width and height",
			vnode:     rtui.Element("div").Prop("width", 30).Prop("height", 20).Build(),
			maxWidth:  100,
			maxHeight: 100,
			wantW:     30,
			wantH:     20,
		},
		{
			name:      "width clamped to maxWidth",
			vnode:     rtui.Element("div").Prop("width", 150).Build(),
			maxWidth:  100,
			maxHeight: 100,
			wantW:     100,
			wantH:     1,
		},
		{
			name:      "height clamped to maxHeight",
			vnode:     rtui.Element("div").Prop("height", 150).Build(),
			maxWidth:  100,
			maxHeight: 100,
			wantW:     100,
			wantH:     100,
		},
		{
			name:      "nil VNode",
			vnode:     nil,
			maxWidth:  100,
			maxHeight: 100,
			wantW:     100,
			wantH:     1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := NewDeclarativeNode(tt.vnode)
			w, h := node.Measure(tt.maxWidth, tt.maxHeight)

			if w != tt.wantW {
				t.Errorf("Measure() width = %d, want %d", w, tt.wantW)
			}
			if h != tt.wantH {
				t.Errorf("Measure() height = %d, want %d", h, tt.wantH)
			}
		})
	}
}

// =============================================================================
// DeclarativeNode.Paint Tests (Non-Fiber Mode)
// =============================================================================

func TestDeclarativeNode_Paint_NonFiber(t *testing.T) {
	tests := []struct {
		name  string
		vnode rtui.VNode
	}{
		{
			name:  "simple text element",
			vnode: rtui.Element("text").Prop("content", "Hello").Build(),
		},
		{
			name: "nested elements",
			vnode: rtui.Element("div").Children(
				rtui.Element("text").Prop("content", "A").Build(),
				rtui.Element("text").Prop("content", "B").Build(),
			).Build(),
		},
		{
			name: "HStack layout",
			vnode: rtui.HStack(
				rtui.Element("text").Prop("content", "A").Build(),
				rtui.Element("text").Prop("content", "B").Build(),
			),
		},
		{
			name: "VStack layout",
			vnode: rtui.VStack(
				rtui.Element("text").Prop("content", "A").Build(),
				rtui.Element("text").Prop("content", "B").Build(),
			),
		},
		{
			name: "fragment",
			vnode: rtui.Fragment(
				rtui.Element("text").Prop("content", "A").Build(),
				rtui.Element("text").Prop("content", "B").Build(),
			),
		},
		{
			name:  "nil VNode",
			vnode: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := NewDeclarativeNode(tt.vnode)
			buf := paint.NewBuffer(80, 24)

			ctx := component.PaintContext{
				Buffer: buf,
				Bounds: paint.Rect{X: 0, Y: 0, Width: 80, Height: 24},
			}

			// Should not panic
			node.Paint(ctx, buf)

			// Buffer should still be valid
			if buf.Width != 80 || buf.Height != 24 {
				t.Errorf("Buffer dimensions changed: %dx%d", buf.Width, buf.Height)
			}
		})
	}
}

// TestDeclarativeNode_Paint_NonFiber_RenderFunc tests Paint with render function
func TestDeclarativeNode_Paint_NonFiber_RenderFunc(t *testing.T) {
	callCount := 0
	fn := func() rtui.VNode {
		callCount++
		return rtui.Element("text").Prop("content", "Rendered").Build()
	}

	node := NewDeclarativeNodeFromFunc(fn)
	buf := paint.NewBuffer(80, 24)

	ctx := component.PaintContext{
		Buffer: buf,
		Bounds: paint.Rect{X: 0, Y: 0, Width: 80, Height: 24},
	}

	// First paint should call render function
	node.Paint(ctx, buf)
	if callCount != 1 {
		t.Errorf("Expected 1 render call, got %d", callCount)
	}

	// Second paint should also call render function (non-Fiber mode)
	node.Paint(ctx, buf)
	if callCount != 2 {
		t.Errorf("Expected 2 render calls, got %d", callCount)
	}
}

// =============================================================================
// DeclarativeNode.Paint Tests (Fiber Mode)
// =============================================================================

func TestDeclarativeNode_Paint_Fiber(t *testing.T) {
	// Create a Fiber-enabled node
	fn := func() rtui.VNode {
		return rtui.Element("text").Prop("content", "Fiber Render").Build()
	}

	app := framework.NewApp()
	node := NewDeclarativeNodeFromFuncWithFiber(fn, app)

	buf := paint.NewBuffer(80, 24)

	ctx := component.PaintContext{
		Buffer: buf,
		Bounds: paint.Rect{X: 0, Y: 0, Width: 80, Height: 24},
	}

	// Should not panic
	node.Paint(ctx, buf)

	// Buffer should still be valid
	if buf.Width != 80 || buf.Height != 24 {
		t.Errorf("Buffer dimensions changed: %dx%d", buf.Width, buf.Height)
	}

	// useFiber should be true
	if !node.useFiber {
		t.Error("useFiber should be true")
	}

	if node.reconciler == nil {
		t.Error("reconciler should be set")
	}
}

// =============================================================================
// DeclarativeNode.SetFrameworkApp Tests
// =============================================================================

func TestDeclarativeNode_SetFrameworkApp(t *testing.T) {
	node := NewDeclarativeNode(rtui.Element("div").Build())

	app := framework.NewApp()
	node.SetFrameworkApp(app)

	node.mu.RLock()
	defer node.mu.RUnlock()

	if node.fwApp != app {
		t.Error("fwApp should be set")
	}
}

// =============================================================================
// DeclarativeNode.SetReconciler Tests
// =============================================================================

func TestDeclarativeNode_SetReconciler(t *testing.T) {
	node := NewDeclarativeNode(rtui.Element("div").Build())

	// Create a mock reconciler
	app := framework.NewApp()
	reconciler := newFiberReconciler(app, func() rtui.VNode {
		return rtui.Element("div").Build()
	})

	node.SetReconciler(reconciler)

	node.mu.RLock()
	defer node.mu.RUnlock()

	if node.reconciler != reconciler {
		t.Error("reconciler should be set")
	}

	if !node.useFiber {
		t.Error("useFiber should be true when reconciler is set")
	}
}

func TestDeclarativeNode_SetReconciler_Nil(t *testing.T) {
	node := NewDeclarativeNodeFromFunc(func() rtui.VNode {
		return rtui.Element("div").Build()
	})

	// Set a reconciler first
	app := framework.NewApp()
	reconciler := newFiberReconciler(app, func() rtui.VNode {
		return rtui.Element("div").Build()
	})
	node.SetReconciler(reconciler)

	// Now clear it
	node.SetReconciler(nil)

	node.mu.RLock()
	defer node.mu.RUnlock()

	if node.reconciler != nil {
		t.Error("reconciler should be nil after setting nil")
	}

	if node.useFiber {
		t.Error("useFiber should be false when reconciler is nil")
	}
}

// =============================================================================
// DeclarativeNode.GetRenderer Tests
// =============================================================================

func TestDeclarativeNode_GetRenderer_Method(t *testing.T) {
	t.Run("non-Fiber mode", func(t *testing.T) {
		node := NewDeclarativeNodeFromFunc(func() rtui.VNode {
			return rtui.Element("div").Build()
		})

		renderer := node.GetRenderer()
		if renderer == nil {
			t.Fatal("GetRenderer() should not return nil")
		}

		_, ok := renderer.(*NonFiberRenderer)
		if !ok {
			t.Errorf("Expected *NonFiberRenderer, got %T", renderer)
		}
	})

	t.Run("Fiber mode", func(t *testing.T) {
		fn := func() rtui.VNode {
			return rtui.Element("div").Build()
		}
		app := framework.NewApp()
		node := NewDeclarativeNodeFromFuncWithFiber(fn, app)

		renderer := node.GetRenderer()
		if renderer == nil {
			t.Fatal("GetRenderer() should not return nil")
		}

		_, ok := renderer.(*FiberRenderer)
		if !ok {
			t.Errorf("Expected *FiberRenderer, got %T", renderer)
		}
	})
}

// =============================================================================
// DeclarativeNode.GetFocusManager Tests
// =============================================================================

func TestDeclarativeNode_GetFocusManager(t *testing.T) {
	node := NewDeclarativeNodeFromFunc(func() rtui.VNode {
		return rtui.Element("div").Build()
	})

	focusMgr := node.GetFocusManager()
	if focusMgr == nil {
		t.Fatal("GetFocusManager() should not return nil")
	}

	if focusMgr != node.focusMgr {
		t.Error("GetFocusManager() should return the internal focus manager")
	}
}

// =============================================================================
// DeclarativeNode.UpdateRoot Tests
// =============================================================================

func TestDeclarativeNode_UpdateRoot(t *testing.T) {
	node := NewDeclarativeNode(rtui.Element("div").Build())

	newRoot := rtui.Element("span").Build()
	node.UpdateRoot(newRoot)

	node.mu.RLock()
	defer node.mu.RUnlock()

	if node.root != newRoot {
		t.Error("root should be updated to newRoot")
	}
}

// =============================================================================
// DeclarativeNode.Mount/Unmount Tests
// =============================================================================

func TestDeclarativeNode_MountUnmount(t *testing.T) {
	node := NewDeclarativeNode(rtui.Element("div").Build())

	// Node with root is considered mounted
	if !node.IsMounted() {
		t.Error("node with root should be mounted")
	}

	// Mount is a no-op for DeclarativeNode (just checks parent context)
	// But should not panic
	node.Mount(nil)

	if !node.IsMounted() {
		t.Error("should still be mounted after Mount()")
	}

	// Unmount clears the root
	node.Unmount()

	if node.IsMounted() {
		t.Error("should not be mounted after Unmount()")
	}

	// Once unmounted, Mount doesn't restore the root (by design)
	// The node would need to be re-created or UpdateRoot called
	node.Mount(nil)

	if node.IsMounted() {
		t.Error("should remain unmounted (root not restored by Mount)")
	}
}

// =============================================================================
// DeclarativeNode Edge Cases
// =============================================================================

func TestDeclarativeNode_EdgeCases(t *testing.T) {
	t.Run("paint with zero-sized buffer", func(t *testing.T) {
		node := NewDeclarativeNode(rtui.Element("text").Prop("content", "Test").Build())
		buf := paint.NewBuffer(0, 0)

		ctx := component.PaintContext{
			Buffer: buf,
			Bounds: paint.Rect{X: 0, Y: 0, Width: 0, Height: 0},
		}

		// Should not panic
		node.Paint(ctx, buf)
	})

	t.Run("paint with very large buffer", func(t *testing.T) {
		node := NewDeclarativeNode(rtui.Element("text").Prop("content", "Test").Build())
		buf := paint.NewBuffer(1000, 1000)

		ctx := component.PaintContext{
			Buffer: buf,
			Bounds: paint.Rect{X: 0, Y: 0, Width: 1000, Height: 1000},
		}

		// Should not panic
		node.Paint(ctx, buf)
	})

	t.Run("paint with negative bounds", func(t *testing.T) {
		node := NewDeclarativeNode(rtui.Element("text").Prop("content", "Test").Build())
		buf := paint.NewBuffer(80, 24)

		ctx := component.PaintContext{
			Buffer: buf,
			Bounds: paint.Rect{X: -10, Y: -10, Width: 80, Height: 24},
		}

		// Should not panic
		node.Paint(ctx, buf)
	})

	t.Run("concurrent paint calls", func(t *testing.T) {
		node := NewDeclarativeNode(rtui.Element("text").Prop("content", "Test").Build())
		buf := paint.NewBuffer(80, 24)

		ctx := component.PaintContext{
			Buffer: buf,
			Bounds: paint.Rect{X: 0, Y: 0, Width: 80, Height: 24},
		}

		// Launch multiple goroutines painting the same node
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func() {
				defer func() { done <- true }()
				node.Paint(ctx, buf)
			}()
		}

		// Wait for all to complete
		for i := 0; i < 10; i++ {
			<-done
		}

		// Buffer should still be valid
		if buf.Width != 80 || buf.Height != 24 {
			t.Errorf("Buffer corrupted: %dx%d", buf.Width, buf.Height)
		}
	})
}

// =============================================================================
// DeclarativeNode.PaintVNode Tests
// =============================================================================

func TestDeclarativeNode_PaintVNode(t *testing.T) {
	node := NewDeclarativeNode(rtui.Element("div").Build())
	buf := paint.NewBuffer(80, 24)

	tests := []struct {
		name  string
		vnode rtui.VNode
	}{
		{
			name:  "text node",
			vnode: rtui.Element("text").Prop("content", "Hello").Build(),
		},
		{
			name:  "nil vnode",
			vnode: nil,
		},
		{
			name: "nested structure",
			vnode: rtui.VStack(
				rtui.HStack(
					rtui.Element("text").Prop("content", "A").Build(),
					rtui.Element("text").Prop("content", "B").Build(),
				),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			node.PaintVNode(tt.vnode, 0, 0, buf)
		})
	}
}

// =============================================================================
// DeclarativeNode.MeasureVNode Tests
// =============================================================================

func TestDeclarativeNode_MeasureVNode(t *testing.T) {
	node := NewDeclarativeNode(rtui.Element("div").Build())

	tests := []struct {
		name     string
		vnode    rtui.VNode
		minWidth int
	}{
		{
			name:     "text node",
			vnode:    rtui.Element("text").Prop("content", "Hello").Build(),
			minWidth: 5,
		},
		{
			name:     "nil vnode",
			vnode:    nil,
			minWidth: 0,
		},
		{
			name:     "empty element",
			vnode:    rtui.Element("div").Build(),
			minWidth: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := node.MeasureVNodeWidth(tt.vnode)
			if w < tt.minWidth {
				t.Errorf("MeasureVNodeWidth() = %d, want >= %d", w, tt.minWidth)
			}
		})
	}
}
