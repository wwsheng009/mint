package render

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/paint"
)

// =============================================================================
// PaintEngine New API Tests
// =============================================================================

func TestPaintEngine_PaintPaintableLayouts(t *testing.T) {
	engine := NewPaintEngine()
	buffer := paint.NewBuffer(80, 25)

	// 创建 PaintableLayouts
	layouts := make(paint.PaintableLayouts)

	// 创建一个简单的 root box
	rootBox := paint.NewPaintableBoxWithBounds(nil, 0, 0, 80, 25)
	layouts[paint.RenderLayerBase] = paint.NewPaintableLayout(rootBox)

	// 调用新 API
	err := engine.PaintPaintableLayouts(layouts, buffer)
	if err != nil {
		t.Fatalf("PaintPaintableLayouts() error = %v", err)
	}
}

func TestPaintEngine_PaintPaintableLayouts_Nil(t *testing.T) {
	engine := NewPaintEngine()
	buffer := paint.NewBuffer(80, 25)

	// 测试 nil 布局
	err := engine.PaintPaintableLayouts(nil, buffer)
	if err != nil {
		t.Fatalf("PaintPaintableLayouts(nil) error = %v", err)
	}

	// 测试空布局
	layouts := make(paint.PaintableLayouts)
	err = engine.PaintPaintableLayouts(layouts, buffer)
	if err != nil {
		t.Fatalf("PaintPaintableLayouts(empty) error = %v", err)
	}
}

func TestPaintEngine_PaintPaintablePlanes(t *testing.T) {
	engine := NewPaintEngine()
	buffer := paint.NewBuffer(80, 25)

	// 创建 PaintablePlanes
	planes := paint.NewPaintablePlanes()
	planes.AddToLayer(paint.RenderLayerBase, paint.NewPaintableBoxWithBounds(nil, 0, 0, 80, 25))
	planes.AddToLayer(paint.RenderLayerModal, paint.NewPaintableBoxWithBounds(nil, 20, 10, 40, 5))

	// 调用新 API
	err := engine.PaintPaintablePlanes(planes, buffer)
	if err != nil {
		t.Fatalf("PaintPaintablePlanes() error = %v", err)
	}
}

func TestPaintEngine_PaintPaintablePlanes_Nil(t *testing.T) {
	engine := NewPaintEngine()
	buffer := paint.NewBuffer(80, 25)

	// 测试 nil planes
	err := engine.PaintPaintablePlanes(nil, buffer)
	if err != nil {
		t.Fatalf("PaintPaintablePlanes(nil) error = %v", err)
	}

	// 测试空 planes
	planes := paint.NewPaintablePlanes()
	err = engine.PaintPaintablePlanes(planes, buffer)
	if err != nil {
		t.Fatalf("PaintPaintablePlanes(empty) error = %v", err)
	}
}

func TestPaintEngine_PaintPaintableLayouts_MultiLayer(t *testing.T) {
	engine := NewPaintEngine()
	buffer := paint.NewBuffer(80, 25)

	// 创建多层 PaintableLayouts
	layouts := make(paint.PaintableLayouts)

	// Base layer
	baseBox := paint.NewPaintableBoxWithBounds(nil, 0, 0, 80, 25)
	layouts[paint.RenderLayerBase] = paint.NewPaintableLayout(baseBox)

	// Modal layer
	modalBox := paint.NewPaintableBoxWithBounds(nil, 20, 10, 40, 5)
	layouts[paint.RenderLayerModal] = paint.NewPaintableLayout(modalBox)

	// Tooltip layer
	tooltipBox := paint.NewPaintableBoxWithBounds(nil, 30, 5, 20, 3)
	layouts[paint.RenderLayerTooltip] = paint.NewPaintableLayout(tooltipBox)

	// 调用新 API
	err := engine.PaintPaintableLayouts(layouts, buffer)
	if err != nil {
		t.Fatalf("PaintPaintableLayouts() error = %v", err)
	}
}

func TestPaintEngine_PaintLayout(t *testing.T) {
	engine := NewPaintEngine()
	buffer := paint.NewBuffer(80, 25)

	// 创建 PaintableLayout
	rootBox := paint.NewPaintableBoxWithBounds(nil, 0, 0, 80, 25)
	layout := paint.NewPaintableLayout(rootBox)

	// 调用 PaintLayout
	err := engine.PaintLayout(layout, buffer)
	if err != nil {
		t.Fatalf("PaintLayout() error = %v", err)
	}
}

func TestPaintEngine_PaintLayout_Nil(t *testing.T) {
	engine := NewPaintEngine()
	buffer := paint.NewBuffer(80, 25)

	// 测试 nil layout
	err := engine.PaintLayout(nil, buffer)
	if err != nil {
		t.Fatalf("PaintLayout(nil) error = %v", err)
	}

	// 测试空 layout
	layout := &paint.PaintableLayout{}
	err = engine.PaintLayout(layout, buffer)
	if err != nil {
		t.Fatalf("PaintLayout(empty) error = %v", err)
	}
}

// =============================================================================
// Modal Backdrop Tests
// =============================================================================

func TestPaintEngine_ModalBackdrop_Tracking(t *testing.T) {
	engine := NewPaintEngine()
	buffer := paint.NewBuffer(80, 25)

	// 第一次渲染：无 modal
	layoutsNoModal := make(paint.PaintableLayouts)
	layoutsNoModal[paint.RenderLayerBase] = paint.NewPaintableLayout(
		paint.NewPaintableBoxWithBounds(nil, 0, 0, 80, 25),
	)
	engine.PaintPaintableLayouts(layoutsNoModal, buffer)

	// 第二次渲染：有 modal
	layoutsWithModal := make(paint.PaintableLayouts)
	layoutsWithModal[paint.RenderLayerBase] = paint.NewPaintableLayout(
		paint.NewPaintableBoxWithBounds(nil, 0, 0, 80, 25),
	)
	// 有效的 modal box 需要有效的 Node
	modalBox := paint.NewPaintableBoxWithBounds(nil, 20, 10, 40, 5)
	layoutsWithModal[paint.RenderLayerModal] = paint.NewPaintableLayout(modalBox)
	engine.PaintPaintableLayouts(layoutsWithModal, buffer)

	// 第三次渲染：modal 消失
	engine.PaintPaintableLayouts(layoutsNoModal, buffer)
}
