package demo

import (
	"strings"
	"testing"

	runtimeplatform "github.com/wwsheng009/mint/runtime/platform"
	linechartcomp "github.com/wwsheng009/mint/ui/components/charts/linechart"
)

func TestPrototypeDiagnosticsIncludeSceneSummary(t *testing.T) {
	lines := PrototypeDiagnostics()
	if len(lines) < 4 {
		t.Fatalf("len(PrototypeDiagnostics()) = %d, want >= 4", len(lines))
	}

	joined := strings.Join(lines, "\n")
	for _, want := range []string{
		"Graphics:",
		"Display:",
		"Backends: text vs requested-image-plot(paused)",
		"Scene: images=0 backend=text requested=image-plot-disabled",
	} {
		if !strings.Contains(joined, want) {
			t.Fatalf("diagnostics = %q, want substring %q", joined, want)
		}
	}
}

func TestPrototypeChartInstanceUsesRequestedBackend(t *testing.T) {
	textInst := prototypeChartInstance(linechartcomp.RenderBackendText)
	if got := textInst.GetProps()["renderBackend"]; got != linechartcomp.RenderBackendText {
		t.Fatalf("text renderBackend = %v, want %v", got, linechartcomp.RenderBackendText)
	}

	imageInst := prototypeChartInstance(linechartcomp.RenderBackendImagePlot)
	if got := imageInst.GetProps()["renderBackend"]; got != linechartcomp.RenderBackendText {
		t.Fatalf("requested image renderBackend = %v, want normalized text backend", got)
	}
}

func TestPrototypeStatusFallbackBanner(t *testing.T) {
	t.Setenv("MINT_GRAPHICS", "off")
	t.Setenv("MINT_GRAPHICS_ALLOW_UNVERIFIED_INLINE_IMAGE", "")

	status := PrototypeStatus()
	if !strings.Contains(status.Banner, "Chart image backend paused") {
		t.Fatalf("Banner = %q, want chart image backend paused message", status.Banner)
	}
	if status.Display != "charts render text only; chart pixel backend paused" {
		t.Fatalf("Display = %q, want chart pixel backend paused display", status.Display)
	}
	if status.ImageTitle != "Requested Image Plot Backend (paused to text)" {
		t.Fatalf("ImageTitle = %q, want paused image title", status.ImageTitle)
	}
}

func TestPrototypeStatusStillReportsGraphicsCapabilitiesWhileChartsStayTextOnly(t *testing.T) {
	t.Setenv("MINT_GRAPHICS", "inline-image")
	t.Setenv("MINT_GRAPHICS_ALLOW_UNVERIFIED_INLINE_IMAGE", "")
	t.Setenv("TERM_PROGRAM", "WezTerm")
	t.Setenv("WEZTERM_PANE", "pane-1")
	t.Setenv("WEZTERM_EXECUTABLE", "")
	t.Setenv("LC_TERMINAL", "")

	status := PrototypeStatus()
	if !strings.Contains(status.Banner, "Chart image backend paused") {
		t.Fatalf("Banner = %q, want chart image backend paused message", status.Banner)
	}
	if status.Display != "charts render text only; chart pixel backend paused" {
		t.Fatalf("Display = %q, want chart pixel backend paused display", status.Display)
	}
	if status.Capabilities.Mode != runtimeplatform.GraphicsModeInlineImage {
		t.Fatalf("Mode = %v, want inline-image", status.Capabilities.Mode)
	}
}
