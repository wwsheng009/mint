package render

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

type sceneTestVNode struct {
	*rtui.ElementVNode
}

func newSceneTestVNode() *sceneTestVNode {
	v := &sceneTestVNode{ElementVNode: rtui.NewElement("scene-test")}
	v.SetKey("scene-test")
	v.SetProp("width", 6)
	v.SetProp("height", 2)
	return v
}

func (v *sceneTestVNode) CreateInstance() rtui.ComponentInstance {
	inst := &sceneTestInstance{
		BaseComponentInstance: rtui.NewBaseComponentInstance("scene-test", nil),
	}
	inst.Init(v.Props())
	return inst
}

type sceneTestInstance struct {
	*rtui.BaseComponentInstance
	bounds [4]int
}

func (i *sceneTestInstance) Paint(x, y int) []paint.DrawCmd {
	return []paint.DrawCmd{{
		X:     x,
		Y:     y,
		Text:  "IMG",
		Style: style.NewStyle().Foreground(style.Green),
	}}
}

func (i *sceneTestInstance) SetBounds(x, y, w, h int) {
	i.bounds = [4]int{x, y, w, h}
}

func (i *sceneTestInstance) SceneLayers() []paint.ImageLayer {
	return []paint.ImageLayer{{
		ID:          "scene-test-layer",
		Bounds:      paint.Rect{X: i.bounds[0], Y: i.bounds[1], Width: i.bounds[2], Height: i.bounds[3]},
		PixelWidth:  12,
		PixelHeight: 4,
		RGBA: []byte{
			0xff, 0x00, 0x00, 0xff,
			0x00, 0xff, 0x00, 0xff,
			0x00, 0x00, 0xff, 0xff,
			0xff, 0xff, 0x00, 0xff,
		},
		AltText: "scene test layer",
	}}
}

func TestDeclarativeNodePaintSceneCollectsImageLayers(t *testing.T) {
	node := NewDeclarativeNodeFromFuncWithFiber(func() rtui.VNode {
		return newSceneTestVNode()
	})

	buf := paint.NewBuffer(12, 6)
	ctx := *paint.NewPaintContext(buf, paint.Rect{X: 0, Y: 0, Width: 12, Height: 6})

	scene := node.PaintScene(ctx, buf)
	if scene == nil {
		t.Fatal("PaintScene() returned nil, want scene with image layers")
	}
	if scene.Buffer != buf {
		t.Fatal("scene buffer should reuse the painted buffer")
	}
	if len(scene.ImageLayers) != 1 {
		t.Fatalf("len(scene.ImageLayers) = %d, want 1", len(scene.ImageLayers))
	}

	layer := scene.ImageLayers[0]
	if !layer.HasPixels() {
		t.Fatal("expected collected image layer to carry raster payload")
	}
	if layer.Bounds.Width <= 0 || layer.Bounds.Height <= 0 {
		t.Fatalf("layer bounds = %+v, want positive size", layer.Bounds)
	}
	if layer.Bounds.X != 0 || layer.Bounds.Y != 0 {
		t.Fatalf("layer origin = (%d,%d), want (0,0)", layer.Bounds.X, layer.Bounds.Y)
	}

	if got := buf.GetContent(0, 0).Cluster; got != "I" {
		t.Fatalf("buffer content at (0,0) = %q, want %q", got, "I")
	}
}

func TestDeclarativeNodePaintSceneReturnsNilForTextOnlyTrees(t *testing.T) {
	node := NewDeclarativeNodeFromFuncWithFiber(func() rtui.VNode {
		return rtui.Element("text").Prop("content", "plain").Build()
	})

	buf := paint.NewBuffer(12, 3)
	ctx := *paint.NewPaintContext(buf, paint.Rect{X: 0, Y: 0, Width: 12, Height: 3})

	scene := node.PaintScene(ctx, buf)
	if scene != nil {
		t.Fatalf("PaintScene() = %+v, want nil for text-only tree", scene)
	}
}
