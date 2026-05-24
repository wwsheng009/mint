package image

import (
	"bytes"
	"encoding/base64"
	stdimage "image"
	"image/color"
	"image/png"
	"strings"
	"testing"

	"github.com/wwsheng009/mint/runtime/layout"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

func TestImageSourceRGBAEmitsSceneLayer(t *testing.T) {
	inst := NewBuilder().
		ID("captcha").
		Alt("captcha code").
		SourceRGBA([]byte{
			255, 0, 0, 255,
			0, 255, 0, 255,
			0, 0, 255, 255,
			255, 255, 255, 255,
		}, 2, 2).
		Size(6, 3).
		Build().
		CreateInstance()

	scene, ok := inst.(rtui.ScenePaintableInstance)
	if !ok {
		t.Fatal("image instance must implement ScenePaintableInstance")
	}
	layers := scene.SceneLayers()
	if len(layers) != 1 {
		t.Fatalf("SceneLayers len = %d, want 1", len(layers))
	}
	layer := layers[0]
	if layer.ID != "captcha" || layer.PixelWidth != 2 || layer.PixelHeight != 2 {
		t.Fatalf("unexpected layer metadata: %+v", layer)
	}
	if layer.Bounds.Width != 6 || layer.Bounds.Height != 3 {
		t.Fatalf("unexpected layer bounds: %+v", layer.Bounds)
	}
	if layer.AltText != "captcha code" {
		t.Fatalf("AltText = %q, want captcha code", layer.AltText)
	}
}

func TestImageDataURIDecodesPNG(t *testing.T) {
	var buf bytes.Buffer
	src := stdimage.NewRGBA(stdimage.Rect(0, 0, 1, 1))
	src.Set(0, 0, color.RGBA{R: 12, G: 34, B: 56, A: 255})
	if err := png.Encode(&buf, src); err != nil {
		t.Fatal(err)
	}

	dataURI := "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())
	inst := NewBuilder().SourceDataURI(dataURI).Alt("pixel").Build().CreateInstance().(*Instance)
	layers := inst.SceneLayers()
	if len(layers) != 1 {
		t.Fatalf("SceneLayers len = %d, want 1", len(layers))
	}
	if layers[0].PixelWidth != 1 || layers[0].PixelHeight != 1 || len(layers[0].RGBA) != 4 {
		t.Fatalf("unexpected decoded layer: %+v", layers[0])
	}
}

func TestImageSVGDataURIUsesFallbackOnly(t *testing.T) {
	payload := base64.StdEncoding.EncodeToString([]byte(`<svg xmlns="http://www.w3.org/2000/svg"/>`))
	inst := NewBuilder().SourceDataURI("data:image/svg+xml;base64," + payload).Alt("captcha").Build().CreateInstance().(*Instance)
	if layers := inst.SceneLayers(); len(layers) != 0 {
		t.Fatalf("SceneLayers len = %d, want 0 for SVG fallback", len(layers))
	}
	size := inst.Measure(layout.Constraints{})
	if size.Width <= 0 || size.Height != 1 {
		t.Fatalf("unexpected fallback size: %+v", size)
	}
	cmds := inst.Paint(0, 0)
	if len(cmds) != 1 || !strings.Contains(cmds[0].Text, "captcha") {
		t.Fatalf("unexpected fallback paint: %+v", cmds)
	}
}
