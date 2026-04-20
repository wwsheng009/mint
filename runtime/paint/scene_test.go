package paint

import "testing"

func TestNewSceneFrame(t *testing.T) {
	buf := NewBuffer(10, 4)
	scene := NewSceneFrame(buf)

	if scene == nil {
		t.Fatal("NewSceneFrame returned nil")
	}
	if scene.Buffer != buf {
		t.Fatal("scene buffer was not preserved")
	}
	if scene.HasImageLayers() {
		t.Fatal("new scene should not have image layers")
	}
}

func TestSceneFrameWithImageLayersCopiesInput(t *testing.T) {
	buf := NewBuffer(8, 3)
	scene := NewSceneFrame(buf)
	layers := []ImageLayer{{
		ID:          "plot",
		Bounds:      Rect{X: 1, Y: 1, Width: 4, Height: 2},
		PixelWidth:  40,
		PixelHeight: 20,
		RGBA:        []byte{1, 2, 3, 4},
	}}

	updated := scene.WithImageLayers(layers...)
	if !updated.HasImageLayers() {
		t.Fatal("expected scene to have image layers")
	}

	layers[0].RGBA[0] = 9
	if updated.ImageLayers[0].RGBA[0] != 1 {
		t.Fatalf("expected RGBA payload copy, got %+v", updated.ImageLayers[0].RGBA)
	}
}

func TestCloneImageLayers(t *testing.T) {
	src := []ImageLayer{{
		ID:          "a",
		PixelWidth:  2,
		PixelHeight: 2,
		RGBA:        []byte{1, 2, 3, 4},
	}}

	cloned := CloneImageLayers(src)
	if len(cloned) != 1 {
		t.Fatalf("expected 1 cloned layer, got %d", len(cloned))
	}

	src[0].RGBA[1] = 9
	if cloned[0].RGBA[1] != 2 {
		t.Fatalf("expected detached RGBA payload, got %+v", cloned[0].RGBA)
	}
}

func TestSceneFrameWithDiagnosticsCopiesNotes(t *testing.T) {
	scene := NewSceneFrame(NewBuffer(4, 2))
	diag := SceneDiagnostics{
		Summary: "plot-only",
		Notes:   []string{"first"},
	}

	updated := scene.WithDiagnostics(diag)
	diag.Notes[0] = "mutated"

	if updated.Diagnostics.Notes[0] != "first" {
		t.Fatalf("expected diagnostics notes to be copied, got %+v", updated.Diagnostics.Notes)
	}
	if updated.Summary() != "plot-only" {
		t.Fatalf("unexpected scene summary: %q", updated.Summary())
	}
}

func TestImageLayerHasPixels(t *testing.T) {
	tests := []struct {
		name  string
		layer ImageLayer
		want  bool
	}{
		{name: "valid", layer: ImageLayer{PixelWidth: 2, PixelHeight: 2, RGBA: []byte{1}}, want: true},
		{name: "missing rgba", layer: ImageLayer{PixelWidth: 2, PixelHeight: 2}, want: false},
		{name: "missing width", layer: ImageLayer{PixelHeight: 2, RGBA: []byte{1}}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.layer.HasPixels(); got != tt.want {
				t.Fatalf("HasPixels() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSceneFrameSummaryFallbacks(t *testing.T) {
	var nilScene *SceneFrame
	if got := nilScene.Summary(); got != "scene=nil" {
		t.Fatalf("nil summary = %q", got)
	}

	scene := &SceneFrame{}
	if got := scene.Summary(); got != "buffer=nil images=0" {
		t.Fatalf("empty scene summary = %q", got)
	}

	scene = NewSceneFrame(NewBuffer(7, 5))
	scene.ImageLayers = []ImageLayer{{ID: "plot"}}
	if got := scene.Summary(); got != "buffer=7x5 images=1" {
		t.Fatalf("buffer summary = %q", got)
	}
}
