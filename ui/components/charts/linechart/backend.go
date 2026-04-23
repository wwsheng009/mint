package linechart

import rtui "github.com/wwsheng009/mint/runtime/ui"

const propRenderBackend = "renderBackend"

// RenderBackend controls how the plot area is rendered.
type RenderBackend int

const (
	RenderBackendText RenderBackend = iota
	RenderBackendImagePlot
)

const imagePlotBackendEnabled = false

// SupportsImagePlotBackend reports whether charts currently expose the
// experimental raster/image-plot backend through the component API.
func SupportsImagePlotBackend() bool {
	return imagePlotBackendEnabled
}

func (b RenderBackend) String() string {
	switch b {
	case RenderBackendText:
		return "text"
	case RenderBackendImagePlot:
		return "image-plot"
	default:
		return "unknown"
	}
}

func getRenderBackendProp(props rtui.Props, def RenderBackend) RenderBackend {
	if value, ok := props[propRenderBackend]; ok {
		if backend, ok := value.(RenderBackend); ok {
			return normalizeRenderBackend(backend)
		}
	}
	return normalizeRenderBackend(def)
}

func normalizeRenderBackend(backend RenderBackend) RenderBackend {
	if backend == RenderBackendImagePlot && !SupportsImagePlotBackend() {
		return RenderBackendText
	}
	return backend
}
