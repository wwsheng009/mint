package platform

import (
	"bytes"
	"encoding/base64"
	"strings"
	"testing"
)

func testInlineImageDrawRequest() DrawImageRequest {
	return DrawImageRequest{
		PixelWidth:  2,
		PixelHeight: 2,
		CellX:       1,
		CellY:       2,
		CellWidth:   3,
		CellHeight:  4,
		RGBA: []byte{
			255, 0, 0, 255,
			0, 255, 0, 255,
			0, 0, 255, 255,
			255, 255, 255, 255,
		},
		AltText: "plot",
	}
}

func TestNewInlineImageGraphicsPresenterNormalizesCapabilities(t *testing.T) {
	presenter := NewInlineImageGraphicsPresenter(nil, GraphicsCapabilities{})
	caps := presenter.Capabilities()

	if caps.Mode != GraphicsModeInlineImage {
		t.Fatalf("mode = %v, want inline-image", caps.Mode)
	}
	if !caps.SupportsPlacement || !caps.SupportsReplace || caps.SupportsDelete {
		t.Fatalf("unexpected normalized caps: %+v", caps)
	}
	if caps.EffectivePresentationModel() != GraphicsPresentationModelOverlay {
		t.Fatalf("presentation model = %v, want overlay", caps.EffectivePresentationModel())
	}
}

func TestInlineImageGraphicsPresenterPresent(t *testing.T) {
	t.Setenv("TERM_PROGRAM", "")
	t.Setenv("WEZTERM_PANE", "")
	t.Setenv("WEZTERM_EXECUTABLE", "")
	t.Setenv("LC_TERMINAL", "")

	var out bytes.Buffer
	presenter := NewInlineImageGraphicsPresenter(&out, GraphicsCapabilities{})

	handle, err := presenter.Present(testInlineImageDrawRequest())
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
	for _, want := range []string{
		"\x1b[s",
		"\x1b[3;2H",
		"\x1b]1337;File=",
		"size=",
		"width=3",
		"height=4",
		"preserveAspectRatio=0",
		"inline=1",
		"\a",
		"\x1b[u",
	} {
		if !strings.Contains(command, want) {
			t.Fatalf("command %q does not contain %q", command, want)
		}
	}
	if strings.Contains(command, "doNotMoveCursor=1") {
		t.Fatalf("generic inline-image command should not contain wezterm-only flag: %q", command)
	}

	payload := inlineImagePayloadFromSequence(t, command)
	decoded, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		t.Fatalf("DecodeString() error = %v", err)
	}
	if !bytes.HasPrefix(decoded, []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}) {
		t.Fatalf("decoded payload is not a png: %x", decoded[:min(len(decoded), 8)])
	}
}

func TestInlineImageGraphicsPresenterWezTermAddsDoNotMoveCursor(t *testing.T) {
	t.Setenv("TERM_PROGRAM", "WezTerm")

	var out bytes.Buffer
	presenter := NewInlineImageGraphicsPresenter(&out, GraphicsCapabilities{})
	if _, err := presenter.Present(testInlineImageDrawRequest()); err != nil {
		t.Fatalf("Present() error = %v", err)
	}

	command := out.String()
	if !strings.Contains(command, "doNotMoveCursor=1") {
		t.Fatalf("wezterm inline-image command %q does not contain doNotMoveCursor=1", command)
	}
}

func TestInlineImageGraphicsPresenterPresentRejectsDuplicateHandle(t *testing.T) {
	t.Setenv("TERM_PROGRAM", "")
	t.Setenv("WEZTERM_PANE", "")
	t.Setenv("WEZTERM_EXECUTABLE", "")
	t.Setenv("LC_TERMINAL", "")

	presenter := NewInlineImageGraphicsPresenter(nil, GraphicsCapabilities{})
	req := testInlineImageDrawRequest()
	req.ID = "plot"

	if _, err := presenter.Present(req); err != nil {
		t.Fatalf("first Present() error = %v", err)
	}
	if _, err := presenter.Present(req); err == nil {
		t.Fatal("expected duplicate handle error")
	}
}

func TestInlineImageGraphicsPresenterPresentWithReplaceIfExists(t *testing.T) {
	t.Setenv("TERM_PROGRAM", "")
	t.Setenv("WEZTERM_PANE", "")
	t.Setenv("WEZTERM_EXECUTABLE", "")
	t.Setenv("LC_TERMINAL", "")

	presenter := NewInlineImageGraphicsPresenter(nil, GraphicsCapabilities{})
	req := testInlineImageDrawRequest()
	req.ID = "plot"

	handle, err := presenter.Present(req)
	if err != nil {
		t.Fatalf("first Present() error = %v", err)
	}

	req.ReplaceIfExists = true
	req.RGBA = []byte{
		0, 0, 0, 255,
		0, 0, 0, 255,
		0, 0, 0, 255,
		0, 0, 0, 255,
	}
	handle, err = presenter.Present(req)
	if err != nil {
		t.Fatalf("replace Present() error = %v", err)
	}
	if presenter.objects[handle].request.RGBA[0] != 0 {
		t.Fatalf("expected replaced payload, got %+v", presenter.objects[handle].request.RGBA)
	}
}

func TestInlineImageGraphicsPresenterReplaceDeleteClear(t *testing.T) {
	t.Setenv("TERM_PROGRAM", "")
	t.Setenv("WEZTERM_PANE", "")
	t.Setenv("WEZTERM_EXECUTABLE", "")
	t.Setenv("LC_TERMINAL", "")

	var out bytes.Buffer
	presenter := NewInlineImageGraphicsPresenter(&out, GraphicsCapabilities{})
	req := testInlineImageDrawRequest()
	req.ID = "plot"

	if _, err := presenter.Present(req); err != nil {
		t.Fatalf("Present() error = %v", err)
	}

	replaceReq := testInlineImageDrawRequest()
	replaceReq.RGBA = []byte{
		1, 2, 3, 255,
		1, 2, 3, 255,
		1, 2, 3, 255,
		1, 2, 3, 255,
	}
	if err := presenter.Replace("plot", replaceReq); err != nil {
		t.Fatalf("Replace() error = %v", err)
	}
	if err := presenter.Delete("plot"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if len(presenter.objects) != 0 {
		t.Fatalf("expected no objects after delete, got %d", len(presenter.objects))
	}

	if _, err := presenter.Present(testInlineImageDrawRequest()); err != nil {
		t.Fatalf("Present() error = %v", err)
	}
	if err := presenter.Clear(); err != nil {
		t.Fatalf("Clear() error = %v", err)
	}
	if len(presenter.objects) != 0 {
		t.Fatalf("expected no objects after clear, got %d", len(presenter.objects))
	}

	log := out.String()
	if !strings.Contains(log, "\x1b[2J\x1b[H") {
		t.Fatalf("expected clear command in output, got %q", log)
	}
}

func TestInlineImageGraphicsPresenterValidation(t *testing.T) {
	t.Setenv("TERM_PROGRAM", "")
	t.Setenv("WEZTERM_PANE", "")
	t.Setenv("WEZTERM_EXECUTABLE", "")
	t.Setenv("LC_TERMINAL", "")

	presenter := NewInlineImageGraphicsPresenter(nil, GraphicsCapabilities{})
	req := testInlineImageDrawRequest()
	req.PixelWidth = 0

	if _, err := presenter.Present(req); err == nil {
		t.Fatal("expected invalid request error")
	}

	shortReq := testInlineImageDrawRequest()
	shortReq.RGBA = []byte{255, 0, 0, 255}
	if _, err := presenter.Present(shortReq); err == nil {
		t.Fatal("expected short rgba payload error")
	}

	if err := presenter.Replace("missing", testInlineImageDrawRequest()); err == nil {
		t.Fatal("expected replace missing handle error")
	}
	if err := presenter.Delete("missing"); err == nil {
		t.Fatal("expected delete missing handle error")
	}
}

func TestEscapeInlineImageControlSequence(t *testing.T) {
	got := escapeInlineImageControlSequence("\x1b]1337;File=inline=1:abc\a\r\n")
	want := "\\x1b]1337;File=inline=1:abc\\a\\r\\n"
	if got != want {
		t.Fatalf("escapeInlineImageControlSequence() = %q, want %q", got, want)
	}
}

func inlineImagePayloadFromSequence(t *testing.T, sequence string) string {
	t.Helper()

	marker := "File="
	index := strings.Index(sequence, marker)
	if index < 0 {
		t.Fatalf("sequence %q does not contain %q", sequence, marker)
	}
	rest := sequence[index:]
	colon := strings.IndexByte(rest, ':')
	if colon < 0 {
		t.Fatalf("sequence %q does not contain payload separator", sequence)
	}
	payload := rest[colon+1:]
	end := strings.IndexByte(payload, '\a')
	if end < 0 {
		t.Fatalf("sequence %q does not contain BEL terminator", sequence)
	}
	return payload[:end]
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
