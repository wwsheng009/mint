package layer

import (
	"github.com/wwsheng009/mint/runtime/compute"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/types"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// 转换函数：compute → paint
// =============================================================================
// 这些函数将 compute 包的类型转换为 paint 包的类型，
// 使 PaintEngine 可以使用完全解耦的 API。
//
// 注意：由于 ui.Layer、layout.Layer 和 paint.RenderLayer 现在都是
// types.Layer 的类型别名，转换实际上是直接类型转换。

// toPaintRenderLayer 将 rtui.Layer 转换为 paint.RenderLayer
// 由于两者都是 types.Layer 的别名，这是直接类型转换
func toPaintRenderLayer(l rtui.Layer) paint.RenderLayer {
	return paint.RenderLayer(l)
}

// toTypesLayer 将 rtui.Layer 转换为 types.Layer
func toTypesLayer(l rtui.Layer) types.Layer {
	return types.Layer(l)
}

// ConvertToPaintableLayouts 将 LayerLayouts 转换为 PaintableLayouts。
func ConvertToPaintableLayouts(layouts LayerLayouts) paint.PaintableLayouts {
	result := make(paint.PaintableLayouts)
	for layer, layout := range layouts {
		if layout != nil {
			result[toPaintRenderLayer(layer)] = layout.AsPaintableLayout()
		}
	}
	return result
}

// ConvertToPaintablePlanes 将 RenderPlanes 转换为 PaintablePlanes。
func ConvertToPaintablePlanes(rp *RenderPlanes) *paint.PaintablePlanes {
	if rp == nil {
		return paint.NewPaintablePlanes()
	}

	pp := paint.NewPaintablePlanes()

	for _, layer := range rp.GetRenderOrder() {
		boxes := rp.GetLayer(layer)
		for _, box := range boxes {
			pp.AddToLayer(toPaintRenderLayer(layer), box.AsPaintable())
		}
	}

	return pp
}

// =============================================================================
// RenderPlanes 扩展方法
// =============================================================================

// AsPaintablePlanes 将 RenderPlanes 转换为 PaintablePlanes。
func (rp *RenderPlanes) AsPaintablePlanes() *paint.PaintablePlanes {
	return ConvertToPaintablePlanes(rp)
}

// =============================================================================
// LayerLayouts 扩展方法
// =============================================================================

// AsPaintableLayouts 将 LayerLayouts 转换为 PaintableLayouts。
func (ll LayerLayouts) AsPaintableLayouts() paint.PaintableLayouts {
	return ConvertToPaintableLayouts(ll)
}

// =============================================================================
// 从 ComputedBox 直接构建 PaintablePlanes
// =============================================================================

// BuildPaintablePlanesFromComputedBox 从 ComputedBox 树构建 PaintablePlanes。
func BuildPaintablePlanesFromComputedBox(root *compute.ComputedBox) *paint.PaintablePlanes {
	if root == nil {
		return paint.NewPaintablePlanes()
	}

	pp := paint.NewPaintablePlanes()

	var walk func(box *compute.ComputedBox)
	walk = func(box *compute.ComputedBox) {
		if box == nil {
			return
		}

		pp.AddToLayer(paint.RenderLayer(box.Layer), box.AsPaintable())

		for _, child := range box.Children {
			walk(child)
		}
	}

	walk(root)
	return pp
}

// =============================================================================
// 从 Fiber 直接构建 PaintablePlanes (Fiber-first)
// =============================================================================

// BuildPaintablePlanesFromFiber 从 Fiber 树构建 PaintablePlanes。
func BuildPaintablePlanesFromFiber(root *rtui.Fiber) *paint.PaintablePlanes {
	if root == nil {
		return paint.NewPaintablePlanes()
	}

	pp := paint.NewPaintablePlanes()

	var walk func(fiber *rtui.Fiber)
	walk = func(fiber *rtui.Fiber) {
		if fiber == nil {
			return
		}

		// 获取 ComputedBox 并转换
		if fiber.ComputedBox != nil {
			if box, ok := fiber.ComputedBox.(*compute.ComputedBox); ok && box != nil {
				pp.AddToLayer(paint.RenderLayer(fiber.Layer), box.AsPaintable())
			}
		}

		// 递归处理子节点
		for childFiber := fiber.Child; childFiber != nil; childFiber = childFiber.Sibling {
			walk(childFiber)
		}
	}

	walk(root)
	return pp
}
