package platform

import (
	"bytes"
	"strings"
	"testing"
)

func testSixelDrawImageRequest() DrawImageRequest {
	req := testDrawImageRequest()
	req.RGBA = bytes.Repeat([]byte{255, 0, 0, 255}, req.PixelWidth*req.PixelHeight)
	return req
}

func TestNewSixelGraphicsPresenterNormalizesCapabilities(t *testing.T) {
	presenter := NewSixelGraphicsPresenter(nil, GraphicsCapabilities{})
	caps := presenter.Capabilities()

	if caps.Mode != GraphicsModeSixel {
		t.Fatalf("mode = %v, want sixel", caps.Mode)
	}
	if !caps.SupportsPlacement || !caps.SupportsReplace {
		t.Fatalf("unexpected normalized caps: %+v", caps)
	}
	if caps.SupportsDelete {
		t.Fatalf("unexpected delete support for sixel caps: %+v", caps)
	}
	if caps.EffectivePresentationModel() != GraphicsPresentationModelTerminalFrame {
		t.Fatalf("presentation model = %v, want terminal-frame", caps.EffectivePresentationModel())
	}
}

func TestSixelGraphicsPresenterPresent(t *testing.T) {
	var out bytes.Buffer
	presenter := NewSixelGraphicsPresenter(&out, GraphicsCapabilities{
		Mode:            GraphicsModeSixel,
		Reliable:        true,
		CellPixelWidth:  2,
		CellPixelHeight: 3,
		ProbeSource:     "env-override",
	})
	req := DrawImageRequest{
		ID:          "plot",
		PixelWidth:  1,
		PixelHeight: 1,
		CellX:       1,
		CellY:       2,
		CellWidth:   2,
		CellHeight:  2,
		RGBA:        []byte{255, 0, 0, 255},
	}

	handle, err := presenter.Present(req)
	if err != nil {
		t.Fatalf("Present() error = %v", err)
	}
	if handle != "plot" {
		t.Fatalf("handle = %q, want plot", handle)
	}

	command := out.String()
	for _, want := range []string{"\x1b[s", "\x1b[3;2H", "\x1b[?80l", "\x1bPq", "\"1;1;4;6", "#0;2;100;0;0", "\x1b\\", "\x1b[u"} {
		if !strings.Contains(command, want) {
			t.Fatalf("command %q does not contain %q", command, want)
		}
	}
}

func TestSixelGraphicsPresenterPresentRejectsDuplicateHandle(t *testing.T) {
	presenter := NewSixelGraphicsPresenter(nil, GraphicsCapabilities{})
	req := testSixelDrawImageRequest()
	req.ID = "plot"

	if _, err := presenter.Present(req); err != nil {
		t.Fatalf("first Present() error = %v", err)
	}
	if _, err := presenter.Present(req); err == nil {
		t.Fatal("expected duplicate handle error")
	}
}

func TestSixelGraphicsPresenterPresentWithReplaceIfExists(t *testing.T) {
	presenter := NewSixelGraphicsPresenter(nil, GraphicsCapabilities{})
	req := testSixelDrawImageRequest()
	req.ID = "plot"

	handle, err := presenter.Present(req)
	if err != nil {
		t.Fatalf("first Present() error = %v", err)
	}

	req.ReplaceIfExists = true
	req.RGBA = bytes.Repeat([]byte{0, 255, 0, 255}, req.PixelWidth*req.PixelHeight)
	handle, err = presenter.Present(req)
	if err != nil {
		t.Fatalf("replace Present() error = %v", err)
	}

	if handle != "plot" {
		t.Fatalf("handle = %q, want plot", handle)
	}
	if presenter.objects[handle].request.RGBA[1] != 255 {
		t.Fatalf("expected replaced payload, got %+v", presenter.objects[handle].request.RGBA)
	}
}

func TestSixelGraphicsPresenterClear(t *testing.T) {
	var out bytes.Buffer
	presenter := NewSixelGraphicsPresenter(&out, GraphicsCapabilities{})

	if _, err := presenter.Present(testSixelDrawImageRequest()); err != nil {
		t.Fatalf("Present() error = %v", err)
	}
	if err := presenter.Clear(); err != nil {
		t.Fatalf("Clear() error = %v", err)
	}
	if len(presenter.objects) != 0 {
		t.Fatalf("expected no objects after clear, got %d", len(presenter.objects))
	}
	if !strings.Contains(out.String(), "\x1b[2J\x1b[H") {
		t.Fatalf("expected terminal clear in output, got %q", out.String())
	}
}

func TestEscapeSixelControlSequence(t *testing.T) {
	got := escapeSixelControlSequence("\x1bPq#0?\x1b\\\r\n")
	want := "\\x1bPq#0?\\x1b\\\\\\r\\n"
	if got != want {
		t.Fatalf("escapeSixelControlSequence() = %q, want %q", got, want)
	}
}
