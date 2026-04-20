package platform

import (
	"bytes"
	"strings"
	"testing"
)

func testDrawImageRequest() DrawImageRequest {
	return DrawImageRequest{
		PixelWidth:  2,
		PixelHeight: 2,
		CellX:       1,
		CellY:       2,
		CellWidth:   3,
		CellHeight:  4,
		RGBA:        []byte{255, 0, 0, 255},
		AltText:     "plot",
	}
}

func TestNewKittyGraphicsPresenterNormalizesCapabilities(t *testing.T) {
	presenter := NewKittyGraphicsPresenter(nil, GraphicsCapabilities{})
	caps := presenter.Capabilities()

	if caps.Mode != GraphicsModeKitty {
		t.Fatalf("mode = %v, want kitty", caps.Mode)
	}
	if !caps.SupportsPlacement || !caps.SupportsReplace || !caps.SupportsDelete {
		t.Fatalf("unexpected normalized caps: %+v", caps)
	}
	if caps.EffectivePresentationModel() != GraphicsPresentationModelOverlay {
		t.Fatalf("presentation model = %v, want overlay", caps.EffectivePresentationModel())
	}
}

func TestKittyGraphicsPresenterPresent(t *testing.T) {
	var out bytes.Buffer
	presenter := NewKittyGraphicsPresenter(&out, GraphicsCapabilities{})

	handle, err := presenter.Present(testDrawImageRequest())
	if err != nil {
		t.Fatalf("Present() error = %v", err)
	}
	if handle == "" {
		t.Fatal("expected non-empty handle")
	}
	if len(presenter.objects) != 1 {
		t.Fatalf("expected 1 object, got %d", len(presenter.objects))
	}

	command := out.String()
	for _, want := range []string{"\x1b[s", "\x1b[3;2H", "\x1b_G", "a=T", "q=2", "f=32", "i=1", "c=3", "r=4", "C=1", "m=0", "/wAA/w==", "\x1b[u"} {
		if !strings.Contains(command, want) {
			t.Fatalf("command %q does not contain %q", command, want)
		}
	}
	if strings.Contains(command, "x=1") || strings.Contains(command, "y=2") {
		t.Fatalf("command %q should not encode cell placement via x/y crop keys", command)
	}
}

func TestKittyGraphicsPresenterPresentChunksLargePayload(t *testing.T) {
	var out bytes.Buffer
	presenter := NewKittyGraphicsPresenter(&out, GraphicsCapabilities{})
	req := testDrawImageRequest()
	req.PixelWidth = 128
	req.PixelHeight = 128
	req.RGBA = bytes.Repeat([]byte{0xff, 0x00, 0x00, 0xff}, 128*128)

	if _, err := presenter.Present(req); err != nil {
		t.Fatalf("Present() error = %v", err)
	}

	command := out.String()
	if !strings.Contains(command, "m=1") {
		t.Fatalf("chunked command %q does not contain continuation chunk marker", command)
	}
	if !strings.Contains(command, "m=0") {
		t.Fatalf("chunked command %q does not contain final chunk marker", command)
	}
	if count := strings.Count(command, "\x1b_G"); count < 2 {
		t.Fatalf("chunked command APC count = %d, want >= 2", count)
	}
}

func TestKittyGraphicsPresenterPresentRejectsDuplicateHandle(t *testing.T) {
	presenter := NewKittyGraphicsPresenter(nil, GraphicsCapabilities{})
	req := testDrawImageRequest()
	req.ID = "plot"

	if _, err := presenter.Present(req); err != nil {
		t.Fatalf("first Present() error = %v", err)
	}
	if _, err := presenter.Present(req); err == nil {
		t.Fatal("expected duplicate handle error")
	}
}

func TestKittyGraphicsPresenterPresentWithReplaceIfExists(t *testing.T) {
	presenter := NewKittyGraphicsPresenter(nil, GraphicsCapabilities{})
	req := testDrawImageRequest()
	req.ID = "plot"

	handle, err := presenter.Present(req)
	if err != nil {
		t.Fatalf("first Present() error = %v", err)
	}

	originalKittyID := presenter.objects[handle].kittyID
	req.ReplaceIfExists = true
	req.RGBA = []byte{0, 255, 0, 255}

	handle, err = presenter.Present(req)
	if err != nil {
		t.Fatalf("replace Present() error = %v", err)
	}
	if presenter.objects[handle].kittyID != originalKittyID {
		t.Fatalf("expected replace to keep kitty id %d, got %d", originalKittyID, presenter.objects[handle].kittyID)
	}
	if presenter.objects[handle].request.RGBA[0] != 0 {
		t.Fatalf("expected replaced payload, got %+v", presenter.objects[handle].request.RGBA)
	}
}

func TestKittyGraphicsPresenterReplaceDeleteClear(t *testing.T) {
	var out bytes.Buffer
	presenter := NewKittyGraphicsPresenter(&out, GraphicsCapabilities{})
	req := testDrawImageRequest()
	req.ID = "plot"

	if _, err := presenter.Present(req); err != nil {
		t.Fatalf("Present() error = %v", err)
	}

	replaceReq := testDrawImageRequest()
	replaceReq.RGBA = []byte{1, 2, 3, 4}
	if err := presenter.Replace("plot", replaceReq); err != nil {
		t.Fatalf("Replace() error = %v", err)
	}
	if err := presenter.Delete("plot"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if len(presenter.objects) != 0 {
		t.Fatalf("expected no objects after delete, got %d", len(presenter.objects))
	}

	if _, err := presenter.Present(testDrawImageRequest()); err != nil {
		t.Fatalf("Present() error = %v", err)
	}
	if err := presenter.Clear(); err != nil {
		t.Fatalf("Clear() error = %v", err)
	}
	if len(presenter.objects) != 0 {
		t.Fatalf("expected no objects after clear, got %d", len(presenter.objects))
	}

	log := out.String()
	if !strings.Contains(log, "a=d,d=I") {
		t.Fatalf("expected delete command in output, got %q", log)
	}
	if !strings.Contains(log, "a=d,d=A") {
		t.Fatalf("expected clear command in output, got %q", log)
	}
}

func TestKittyGraphicsPresenterValidation(t *testing.T) {
	presenter := NewKittyGraphicsPresenter(nil, GraphicsCapabilities{})
	req := testDrawImageRequest()
	req.PixelWidth = 0

	if _, err := presenter.Present(req); err == nil {
		t.Fatal("expected invalid request error")
	}
	if err := presenter.Replace("missing", testDrawImageRequest()); err == nil {
		t.Fatal("expected replace missing handle error")
	}
	if err := presenter.Delete("missing"); err == nil {
		t.Fatal("expected delete missing handle error")
	}
}

func TestEscapeKittyControlSequence(t *testing.T) {
	got := escapeKittyControlSequence("\x1b_Ga=T;abc\x1b\\\r\n")
	want := "\\x1b_Ga=T;abc\\x1b\\\\\\r\\n"
	if got != want {
		t.Fatalf("escapeKittyControlSequence() = %q, want %q", got, want)
	}
}
