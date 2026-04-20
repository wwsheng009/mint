package linechart

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

func TestBuilderImagePlotBackendFallsBackToText(t *testing.T) {
	vnode := NewBuilder([]float64{1, 3, 2, 5}).
		Title("Trend").
		ImagePlotBackend().
		BuildTyped()

	if vnode.RenderBackend() != RenderBackendText {
		t.Fatalf("RenderBackend() = %v, want normalized text backend", vnode.RenderBackend())
	}

	inst := vnode.CreateInstance().(*Instance)
	if got := inst.GetProps()[propRenderBackend]; got != RenderBackendText {
		t.Fatalf("instance prop renderBackend = %v, want normalized text backend", got)
	}

	propsVNode := New(nil).SetProps(rtui.Props{
		propRenderBackend: RenderBackendImagePlot,
	})
	if propsVNode.(*VNode).RenderBackend() != RenderBackendText {
		t.Fatalf("SetProps renderBackend = %v, want normalized text backend", propsVNode.(*VNode).RenderBackend())
	}
}

func TestInstanceSceneLayersDisabledEvenWhenImagePlotRequested(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propTitle:         "Trend",
		propData:          []float64{1, 3, 2, 5},
		propWidth:         5,
		propHeight:        4,
		propShowAxis:      true,
		propShowPoints:    true,
		propRenderBackend: RenderBackendImagePlot,
	})
	inst.SetBounds(2, 3, 5, 6)

	if got := inst.GetProps()[propRenderBackend]; got != RenderBackendText {
		t.Fatalf("instance renderBackend = %v, want normalized text backend", got)
	}
	if layers := inst.SceneLayers(); len(layers) != 0 {
		t.Fatalf("SceneLayers() len = %d, want 0 while chart image backend is paused", len(layers))
	}
}

func TestRenderPlotImageLayerGeneratesRaster(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propTitle:      "Trend",
		propData:       []float64{1, 3, 2, 5, 4},
		propWidth:      5,
		propHeight:     4,
		propShowAxis:   true,
		propShowGrid:   true,
		propShowLegend: false,
		propShowPoints: true,
	})
	inst.SetBounds(3, 4, 5, 6)

	layer, ok := inst.renderPlotImageLayer()
	if !ok {
		t.Fatal("renderPlotImageLayer() = false, want raster layer")
	}
	if layer.Bounds.X != 3 || layer.Bounds.Y != 5 {
		t.Fatalf("layer bounds origin = (%d,%d), want (3,5)", layer.Bounds.X, layer.Bounds.Y)
	}
	if layer.Bounds.Width != 5 || layer.Bounds.Height != 4 {
		t.Fatalf("layer bounds size = %dx%d, want 5x4", layer.Bounds.Width, layer.Bounds.Height)
	}
	if layer.PixelWidth != 40 || layer.PixelHeight != 48 {
		t.Fatalf("layer pixel size = %dx%d, want 40x48", layer.PixelWidth, layer.PixelHeight)
	}
	if !layer.HasPixels() {
		t.Fatal("expected layer to carry pixel payload")
	}
	if len(layer.RGBA) != layer.PixelWidth*layer.PixelHeight*4 {
		t.Fatalf("len(layer.RGBA) = %d, want %d", len(layer.RGBA), layer.PixelWidth*layer.PixelHeight*4)
	}
	if layer.RGBA[3] != 255 {
		t.Fatalf("expected opaque background alpha, got %d", layer.RGBA[3])
	}
	if layer.ID == "" {
		t.Fatal("expected stable layer id")
	}
}

func TestRenderPlotImageLayerUsesConfiguredCellPixels(t *testing.T) {
	t.Setenv("MINT_CELL_PIXELS", "8x16")

	inst := NewInstance(rtui.Props{
		propTitle:      "Trend",
		propData:       []float64{1, 3, 2, 5, 4},
		propWidth:      5,
		propHeight:     4,
		propShowAxis:   true,
		propShowGrid:   true,
		propShowLegend: false,
		propShowPoints: true,
	})
	inst.SetBounds(3, 4, 5, 6)

	layer, ok := inst.renderPlotImageLayer()
	if !ok {
		t.Fatal("renderPlotImageLayer() = false, want raster layer")
	}
	if layer.PixelWidth != 40 || layer.PixelHeight != 64 {
		t.Fatalf("layer pixel size = %dx%d, want 40x64", layer.PixelWidth, layer.PixelHeight)
	}
}

func TestRenderPlotImageLayerUsesChartBackgroundColor(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propTitle:      "Trend",
		propData:       []float64{2, 5, 3, 4},
		propWidth:      5,
		propHeight:     4,
		propShowAxis:   false,
		propShowGrid:   false,
		propShowLegend: false,
		propShowPoints: false,
		propStyle:      style.Style{BG: style.Color("red")},
	})
	inst.SetBounds(0, 0, 5, 4)

	layer, ok := inst.renderPlotImageLayer()
	if !ok {
		t.Fatal("renderPlotImageLayer() = false, want raster layer")
	}
	if len(layer.RGBA) < 4 {
		t.Fatalf("len(layer.RGBA) = %d, want >= 4", len(layer.RGBA))
	}
	if layer.RGBA[0] != 205 || layer.RGBA[1] != 49 || layer.RGBA[2] != 49 || layer.RGBA[3] != 255 {
		t.Fatalf(
			"top-left pixel rgba = [%d %d %d %d], want [205 49 49 255]",
			layer.RGBA[0],
			layer.RGBA[1],
			layer.RGBA[2],
			layer.RGBA[3],
		)
	}
}

func TestRenderPlotImageLayerDefaultsToBlackBackground(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propTitle:      "Trend",
		propData:       []float64{2, 5, 3, 4},
		propWidth:      5,
		propHeight:     4,
		propShowAxis:   false,
		propShowGrid:   false,
		propShowLegend: false,
		propShowPoints: false,
	})
	inst.SetBounds(0, 0, 5, 4)

	layer, ok := inst.renderPlotImageLayer()
	if !ok {
		t.Fatal("renderPlotImageLayer() = false, want raster layer")
	}
	if len(layer.RGBA) < 4 {
		t.Fatalf("len(layer.RGBA) = %d, want >= 4", len(layer.RGBA))
	}
	if layer.RGBA[0] != 0 || layer.RGBA[1] != 0 || layer.RGBA[2] != 0 || layer.RGBA[3] != 255 {
		t.Fatalf(
			"default top-left pixel rgba = [%d %d %d %d], want [0 0 0 255]",
			layer.RGBA[0],
			layer.RGBA[1],
			layer.RGBA[2],
			layer.RGBA[3],
		)
	}
}

func TestPlotImageContentRectUsesCellInsets(t *testing.T) {
	t.Setenv("MINT_CELL_PIXELS", "8x16")

	rect := plotImageContentRect(248, 96)
	if rect.X != 2 || rect.Y != 4 || rect.Width != 244 || rect.Height != 88 {
		t.Fatalf("plotImageContentRect() = %+v, want {X:2 Y:4 Width:244 Height:88}", rect)
	}
}
