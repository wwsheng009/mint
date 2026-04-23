package paint

import "fmt"

// ImageLayer carries the minimal image-layer semantics needed by Phase 1.
// It stays protocol-agnostic and only describes raster content plus its logical
// cell placement inside the rendered scene.
type ImageLayer struct {
	ID          string
	Bounds      Rect
	PixelWidth  int
	PixelHeight int
	RGBA        []byte
	AltText     string
	ZIndex      int
}

// HasPixels reports whether the layer carries a non-empty raster payload.
func (l ImageLayer) HasPixels() bool {
	return l.PixelWidth > 0 && l.PixelHeight > 0 && len(l.RGBA) > 0
}

// Clone returns a copy of the layer with its raster payload detached from the
// source slice to avoid aliasing between callers.
func (l ImageLayer) Clone() ImageLayer {
	out := l
	if len(l.RGBA) > 0 {
		out.RGBA = append([]byte(nil), l.RGBA...)
	}
	return out
}

// SceneDiagnostics stores lightweight debugging metadata about a scene frame.
type SceneDiagnostics struct {
	Summary string
	Notes   []string
}

// WithNotes returns a copy with notes appended.
func (d SceneDiagnostics) WithNotes(notes ...string) SceneDiagnostics {
	if len(notes) == 0 {
		return d
	}

	out := d
	out.Notes = append(append([]string(nil), d.Notes...), notes...)
	return out
}

// SceneFrame wraps the regular text buffer with optional image layers and
// diagnostics, without changing Buffer semantics.
type SceneFrame struct {
	Buffer      *Buffer
	ImageLayers []ImageLayer
	Diagnostics SceneDiagnostics
}

// NewSceneFrame creates a new scene frame around an existing text buffer.
func NewSceneFrame(buffer *Buffer) *SceneFrame {
	return &SceneFrame{Buffer: buffer}
}

// HasImageLayers reports whether the scene contains any image layers.
func (f *SceneFrame) HasImageLayers() bool {
	return f != nil && len(f.ImageLayers) > 0
}

// CloneImageLayers returns a detached copy of the current image layers.
func (f *SceneFrame) CloneImageLayers() []ImageLayer {
	if f == nil {
		return nil
	}
	return CloneImageLayers(f.ImageLayers)
}

// WithImageLayers returns a copy of the scene with detached image-layer data.
func (f SceneFrame) WithImageLayers(layers ...ImageLayer) SceneFrame {
	f.ImageLayers = CloneImageLayers(layers)
	return f
}

// WithDiagnostics returns a copy of the scene with copied diagnostics notes.
func (f SceneFrame) WithDiagnostics(diag SceneDiagnostics) SceneFrame {
	diag.Notes = append([]string(nil), diag.Notes...)
	f.Diagnostics = diag
	return f
}

// Summary returns a stable lightweight summary for diagnostics and tests.
func (f *SceneFrame) Summary() string {
	if f == nil {
		return "scene=nil"
	}
	if f.Diagnostics.Summary != "" {
		return f.Diagnostics.Summary
	}
	if f.Buffer == nil {
		return fmt.Sprintf("buffer=nil images=%d", len(f.ImageLayers))
	}
	return fmt.Sprintf("buffer=%dx%d images=%d", f.Buffer.Width, f.Buffer.Height, len(f.ImageLayers))
}

// CloneImageLayers returns a detached copy of image layers, including RGBA
// payload slices, so later caller mutation cannot affect stored scene state.
func CloneImageLayers(layers []ImageLayer) []ImageLayer {
	if len(layers) == 0 {
		return nil
	}

	out := make([]ImageLayer, len(layers))
	for i := range layers {
		out[i] = layers[i].Clone()
	}
	return out
}
