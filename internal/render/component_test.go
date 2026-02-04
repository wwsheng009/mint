// Package render tests for component types.
package render

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/paint"
)

// =============================================================================
// SimplePaintContext Tests
// =============================================================================

func TestNewSimplePaintContext(t *testing.T) {
	buf := paint.NewBuffer(80, 24)
	bounds := paint.Rect{X: 0, Y: 0, Width: 80, Height: 24}

	ctx := NewSimplePaintContext(buf, bounds)

	if ctx == nil {
		t.Fatal("NewSimplePaintContext should not return nil")
	}

	if ctx.buf != buf {
		t.Error("buf should be set")
	}

	if ctx.bounds != bounds {
		t.Error("bounds should be set")
	}

	if ctx.values == nil {
		t.Error("values should be initialized")
	}
}

func TestSimplePaintContext_Buffer(t *testing.T) {
	buf := paint.NewBuffer(100, 50)
	ctx := NewSimplePaintContext(buf, paint.Rect{})

	if ctx.Buffer() != buf {
		t.Error("Buffer() should return the original buffer")
	}
}

func TestSimplePaintContext_Bounds(t *testing.T) {
	bounds := paint.Rect{X: 10, Y: 20, Width: 100, Height: 50}
	ctx := NewSimplePaintContext(nil, bounds)

	if ctx.Bounds() != bounds {
		t.Errorf("Bounds() = %+v, want %+v", ctx.Bounds(), bounds)
	}
}

func TestSimplePaintContext_SetValue(t *testing.T) {
	ctx := NewSimplePaintContext(nil, paint.Rect{})

	ctx.SetValue("key1", "value1")
	ctx.SetValue("key2", 42)

	if len(ctx.values) != 2 {
		t.Errorf("expected 2 values, got %d", len(ctx.values))
	}

	v1, ok := ctx.values["key1"]
	if !ok || v1 != "value1" {
		t.Error("key1 not set correctly")
	}

	v2, ok := ctx.values["key2"]
	if !ok || v2 != 42 {
		t.Error("key2 not set correctly")
	}
}

func TestSimplePaintContext_GetValue(t *testing.T) {
	ctx := NewSimplePaintContext(nil, paint.Rect{})
	ctx.SetValue("existing", "value")

	t.Run("get existing value", func(t *testing.T) {
		v, ok := ctx.GetValue("existing")
		if !ok {
			t.Error("should return true for existing key")
		}
		if v != "value" {
			t.Errorf("got %v, want 'value'", v)
		}
	})

	t.Run("get non-existing value", func(t *testing.T) {
		v, ok := ctx.GetValue("non-existing")
		if ok {
			t.Error("should return false for non-existing key")
		}
		if v != nil {
			t.Errorf("got %v, want nil", v)
		}
	})
}

func TestSimplePaintContext_Painter(t *testing.T) {
	buf := paint.NewBuffer(80, 24)
	bounds := paint.Rect{X: 0, Y: 0, Width: 80, Height: 24}
	ctx := NewSimplePaintContext(buf, bounds)

	painter := ctx.Painter()

	if painter == nil {
		t.Fatal("Painter() should not return nil")
	}
}

// =============================================================================
// Constraints Tests
// =============================================================================

func TestNewConstraints(t *testing.T) {
	c := NewConstraints(10, 100, 20, 200)

	if c.MinWidth != 10 {
		t.Errorf("MinWidth = %d, want 10", c.MinWidth)
	}
	if c.MaxWidth != 100 {
		t.Errorf("MaxWidth = %d, want 100", c.MaxWidth)
	}
	if c.MinHeight != 20 {
		t.Errorf("MinHeight = %d, want 20", c.MinHeight)
	}
	if c.MaxHeight != 200 {
		t.Errorf("MaxHeight = %d, want 200", c.MaxHeight)
	}
}

func TestConstraints_Unbounded(t *testing.T) {
	c := Unbounded()

	const maxInt = int(^uint(0) >> 1)

	if c.MaxWidth != maxInt {
		t.Errorf("MaxWidth = %d, want %d", c.MaxWidth, maxInt)
	}
	if c.MaxHeight != maxInt {
		t.Errorf("MaxHeight = %d, want %d", c.MaxHeight, maxInt)
	}
	if c.MinWidth != 0 {
		t.Errorf("MinWidth = %d, want 0", c.MinWidth)
	}
	if c.MinHeight != 0 {
		t.Errorf("MinHeight = %d, want 0", c.MinHeight)
	}
}

func TestConstraints_Tight(t *testing.T) {
	c := Tight(50, 60)

	if c.MinWidth != 50 {
		t.Errorf("MinWidth = %d, want 50", c.MinWidth)
	}
	if c.MaxWidth != 50 {
		t.Errorf("MaxWidth = %d, want 50", c.MaxWidth)
	}
	if c.MinHeight != 60 {
		t.Errorf("MinHeight = %d, want 60", c.MinHeight)
	}
	if c.MaxHeight != 60 {
		t.Errorf("MaxHeight = %d, want 60", c.MaxHeight)
	}
}

func TestConstraints_Width(t *testing.T) {
	base := NewConstraints(0, 80, 10, 100)

	c := base.Width(20, 60)

	if c.MinWidth != 20 {
		t.Errorf("MinWidth = %d, want 20", c.MinWidth)
	}
	if c.MaxWidth != 60 {
		t.Errorf("MaxWidth = %d, want 60", c.MaxWidth)
	}
	if c.MinHeight != 10 {
		t.Errorf("MinHeight = %d, want 10", c.MinHeight)
	}
	if c.MaxHeight != 100 {
		t.Errorf("MaxHeight = %d, want 100", c.MaxHeight)
	}

	// Original should be unchanged
	if base.MinWidth != 0 {
		t.Error("original constraints should be unchanged")
	}
}

func TestConstraints_Height(t *testing.T) {
	base := NewConstraints(0, 80, 10, 100)

	c := base.Height(30, 70)

	if c.MinWidth != 0 {
		t.Errorf("MinWidth = %d, want 0", c.MinWidth)
	}
	if c.MaxWidth != 80 {
		t.Errorf("MaxWidth = %d, want 80", c.MaxWidth)
	}
	if c.MinHeight != 30 {
		t.Errorf("MinHeight = %d, want 30", c.MinHeight)
	}
	if c.MaxHeight != 70 {
		t.Errorf("MaxHeight = %d, want 70", c.MaxHeight)
	}

	// Original should be unchanged
	if base.MinHeight != 10 {
		t.Error("original constraints should be unchanged")
	}
}

// =============================================================================
// Size Tests
// =============================================================================

func TestSize_Zero(t *testing.T) {
	s := ZeroSize()

	if s.Width != 0 {
		t.Errorf("Width = %d, want 0", s.Width)
	}
	if s.Height != 0 {
		t.Errorf("Height = %d, want 0", s.Height)
	}
}

func TestSize_Infinite(t *testing.T) {
	s := InfiniteSize()

	const maxInt = int(^uint(0) >> 1)

	if s.Width != maxInt {
		t.Errorf("Width = %d, want %d", s.Width, maxInt)
	}
	if s.Height != maxInt {
		t.Errorf("Height = %d, want %d", s.Height, maxInt)
	}
}

func TestSize_NewSize(t *testing.T) {
	s := NewSize(100, 200)

	if s.Width != 100 {
		t.Errorf("Width = %d, want 100", s.Width)
	}
	if s.Height != 200 {
		t.Errorf("Height = %d, want 200", s.Height)
	}
}
