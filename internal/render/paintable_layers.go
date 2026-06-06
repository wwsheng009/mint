package render

import "github.com/wwsheng009/mint/runtime/paint"

func walkPaintableBoxesByEffectiveLayer(root *paint.PaintableBox, visit func(paint.RenderLayer, *paint.PaintableBox)) {
	var walk func(*paint.PaintableBox, paint.RenderLayer)
	walk = func(box *paint.PaintableBox, inherited paint.RenderLayer) {
		if box == nil {
			return
		}

		layer := paint.RenderLayer(box.Layer)
		if layer < inherited {
			layer = inherited
		}
		visit(layer, box)

		for _, child := range box.Children {
			walk(child, layer)
		}
	}
	walk(root, paint.RenderLayerBase)
}
