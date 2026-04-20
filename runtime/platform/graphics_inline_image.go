package platform

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"io"
	"os"
	"strings"

	"github.com/wwsheng009/mint/internal/log"
)

const inlineImageGraphicsDebugEnvVar = "MINT_DEBUG_INLINE_IMAGE_GRAPHICS"

var inlineImageGraphicsLogger = log.NewLoggerWithEnv("InlineImageGraphics", "PLATFORM", inlineImageGraphicsDebugEnvVar)

type inlineImageFlavor string

const (
	inlineImageFlavorGeneric inlineImageFlavor = "generic"
	inlineImageFlavorWezTerm inlineImageFlavor = "wezterm"
	inlineImageFlavorITerm2  inlineImageFlavor = "iterm2"
)

type inlineImageObject struct {
	handle  string
	request DrawImageRequest
}

type inlineImageEmission struct {
	sequence string
	summary  string
}

// InlineImageGraphicsPresenter renders images using the iTerm2 inline image
// protocol, which is also implemented by WezTerm.
type InlineImageGraphicsPresenter struct {
	writer     io.Writer
	caps       GraphicsCapabilities
	objects    map[string]inlineImageObject
	nextHandle int
	flavor     inlineImageFlavor
}

// NewInlineImageGraphicsPresenter creates a new iTerm2/WezTerm inline-image presenter.
func NewInlineImageGraphicsPresenter(writer io.Writer, caps GraphicsCapabilities) *InlineImageGraphicsPresenter {
	inlineImageGraphicsLogger.SetEnabled(inlineImageGraphicsDebugEnabled())
	normalized := normalizeInlineImageCapabilities(caps)
	return &InlineImageGraphicsPresenter{
		writer:     writer,
		caps:       normalized,
		objects:    make(map[string]inlineImageObject),
		nextHandle: 1,
		flavor:     detectInlineImageFlavor(normalized),
	}
}

// Capabilities reports normalized inline-image capabilities.
func (p *InlineImageGraphicsPresenter) Capabilities() GraphicsCapabilities {
	if p == nil {
		return normalizeInlineImageCapabilities(GraphicsCapabilities{})
	}
	return p.caps
}

// Present draws or replaces an inline image at the requested cell position.
func (p *InlineImageGraphicsPresenter) Present(req DrawImageRequest) (string, error) {
	if p == nil {
		return "", fmt.Errorf("inline-image presenter is nil")
	}
	if err := validateDrawImageRequest(req); err != nil {
		return "", err
	}

	handle := req.ID
	if handle == "" {
		handle = p.newHandle()
	}
	if existing, ok := p.objects[handle]; ok {
		if !req.ReplaceIfExists {
			return "", fmt.Errorf("image handle %q already exists", handle)
		}
		req.ID = handle
		if err := p.replaceObject(existing, req); err != nil {
			return "", err
		}
		return handle, nil
	}

	req.ID = handle
	obj := inlineImageObject{
		handle:  handle,
		request: cloneDrawImageRequest(req),
	}
	emission, err := p.presentEmission(obj, "present")
	if err != nil {
		return "", err
	}
	if err := p.emit(emission); err != nil {
		return "", err
	}
	p.objects[handle] = obj
	return handle, nil
}

// Replace updates an existing inline image by redrawing it in place.
func (p *InlineImageGraphicsPresenter) Replace(id string, req DrawImageRequest) error {
	if p == nil {
		return fmt.Errorf("inline-image presenter is nil")
	}
	if id == "" {
		return fmt.Errorf("image handle is required")
	}
	if err := validateDrawImageRequest(req); err != nil {
		return err
	}
	existing, ok := p.objects[id]
	if !ok {
		return fmt.Errorf("image handle %q not found", id)
	}
	req.ID = id
	return p.replaceObject(existing, req)
}

// Delete removes tracked lifecycle state. The inline image protocol has no
// stable object deletion primitive, so actual cleanup falls back to terminal clears.
func (p *InlineImageGraphicsPresenter) Delete(id string) error {
	if p == nil {
		return fmt.Errorf("inline-image presenter is nil")
	}
	if id == "" {
		return fmt.Errorf("image handle is required")
	}
	if _, ok := p.objects[id]; !ok {
		return fmt.Errorf("image handle %q not found", id)
	}
	delete(p.objects, id)
	return nil
}

// Clear clears the terminal so callers can repaint text before drawing the next frame.
func (p *InlineImageGraphicsPresenter) Clear() error {
	if p == nil {
		return fmt.Errorf("inline-image presenter is nil")
	}
	if len(p.objects) == 0 {
		return nil
	}
	if err := p.emit(inlineImageEmission{
		sequence: "\x1b[2J\x1b[H",
		summary:  fmt.Sprintf("clear tracked=%d mode=terminal-clear", len(p.objects)),
	}); err != nil {
		return err
	}
	for handle := range p.objects {
		delete(p.objects, handle)
	}
	return nil
}

func (p *InlineImageGraphicsPresenter) replaceObject(existing inlineImageObject, req DrawImageRequest) error {
	req.ID = existing.handle
	updated := inlineImageObject{
		handle:  existing.handle,
		request: cloneDrawImageRequest(req),
	}
	emission, err := p.presentEmission(updated, "replace")
	if err != nil {
		return err
	}
	if err := p.emit(emission); err != nil {
		return err
	}
	p.objects[existing.handle] = updated
	return nil
}

func (p *InlineImageGraphicsPresenter) newHandle() string {
	handle := fmt.Sprintf("inline-img-%d", p.nextHandle)
	p.nextHandle++
	return handle
}

func (p *InlineImageGraphicsPresenter) emit(emission inlineImageEmission) error {
	if p.writer == nil {
		return nil
	}
	_, err := io.WriteString(p.writer, emission.sequence)
	if err != nil {
		debugInlineImageEmissionError(emission.summary, err)
		return err
	}
	debugInlineImageEmission(emission)
	return nil
}

func (p *InlineImageGraphicsPresenter) presentEmission(obj inlineImageObject, op string) (inlineImageEmission, error) {
	pngPayload, err := encodeInlineImagePNG(obj.request)
	if err != nil {
		return inlineImageEmission{}, err
	}

	payload := base64.StdEncoding.EncodeToString(pngPayload)
	row := obj.request.CellY + 1
	col := obj.request.CellX + 1

	args := []string{
		"name=" + base64.StdEncoding.EncodeToString([]byte(inlineImageFileName(obj))),
		fmt.Sprintf("size=%d", len(pngPayload)),
		fmt.Sprintf("width=%d", obj.request.CellWidth),
		fmt.Sprintf("height=%d", obj.request.CellHeight),
		"preserveAspectRatio=0",
		"inline=1",
	}
	if p.flavor == inlineImageFlavorWezTerm {
		args = append(args, "doNotMoveCursor=1")
	}

	var builder strings.Builder
	builder.WriteString("\x1b[s")
	builder.WriteString(fmt.Sprintf("\x1b[%d;%dH", row, col))
	builder.WriteString("\x1b]1337;File=")
	builder.WriteString(strings.Join(args, ";"))
	builder.WriteByte(':')
	builder.WriteString(payload)
	builder.WriteByte('\a')
	builder.WriteString("\x1b[u")

	return inlineImageEmission{
		sequence: builder.String(),
		summary: fmt.Sprintf(
			"%s handle=%s flavor=%s cell=(x=%d,y=%d) cursor=(row=%d,col=%d) cells=%dx%d pixels=%dx%d png_bytes=%d payload_base64=%d source=%s reliable=%t",
			op,
			obj.handle,
			p.flavor,
			obj.request.CellX,
			obj.request.CellY,
			row,
			col,
			obj.request.CellWidth,
			obj.request.CellHeight,
			obj.request.PixelWidth,
			obj.request.PixelHeight,
			len(pngPayload),
			len(payload),
			p.caps.ProbeSource,
			p.caps.Reliable,
		),
	}, nil
}

func encodeInlineImagePNG(req DrawImageRequest) ([]byte, error) {
	expected := req.PixelWidth * req.PixelHeight * 4
	if expected <= 0 {
		return nil, fmt.Errorf("invalid source pixel dimensions %dx%d", req.PixelWidth, req.PixelHeight)
	}
	if len(req.RGBA) < expected {
		return nil, fmt.Errorf("rgba payload too short: got %d want >= %d", len(req.RGBA), expected)
	}

	img := image.NewNRGBA(image.Rect(0, 0, req.PixelWidth, req.PixelHeight))
	copy(img.Pix, req.RGBA[:expected])

	var out bytes.Buffer
	if err := png.Encode(&out, img); err != nil {
		return nil, fmt.Errorf("encode inline-image png: %w", err)
	}
	return out.Bytes(), nil
}

func inlineImageFileName(obj inlineImageObject) string {
	if obj.handle != "" {
		return obj.handle + ".png"
	}
	if obj.request.ID != "" {
		return obj.request.ID + ".png"
	}
	return "mint-inline-image.png"
}

func normalizeInlineImageCapabilities(caps GraphicsCapabilities) GraphicsCapabilities {
	if caps.Mode == GraphicsModeNone {
		caps.Mode = GraphicsModeInlineImage
	}
	caps.Mode = GraphicsModeInlineImage
	caps.PresentationModel = GraphicsPresentationModelOverlay
	caps.SupportsPlacement = true
	caps.SupportsReplace = true
	caps.SupportsDelete = false
	if caps.ProbeSource == "" {
		caps.ProbeSource = "inline-image-presenter"
	}
	return caps
}

func detectInlineImageFlavor(caps GraphicsCapabilities) inlineImageFlavor {
	termProgram := strings.TrimSpace(strings.ToLower(os.Getenv("TERM_PROGRAM")))
	lcTerminal := strings.TrimSpace(strings.ToLower(os.Getenv("LC_TERMINAL")))
	weztermPane := strings.TrimSpace(os.Getenv("WEZTERM_PANE"))
	weztermExecutable := strings.TrimSpace(os.Getenv("WEZTERM_EXECUTABLE"))

	if looksLikeWezTerm(termProgram, weztermPane, weztermExecutable) ||
		strings.Contains(strings.ToLower(caps.ProbeSource), "wezterm") ||
		notesContainGraphicsHint(caps.Notes, "wezterm") {
		return inlineImageFlavorWezTerm
	}
	if looksLikeITerm2(termProgram, lcTerminal) ||
		strings.Contains(strings.ToLower(caps.ProbeSource), "iterm2") ||
		notesContainGraphicsHint(caps.Notes, "iterm2") {
		return inlineImageFlavorITerm2
	}
	return inlineImageFlavorGeneric
}

func notesContainGraphicsHint(notes []string, needle string) bool {
	if needle == "" {
		return false
	}
	needle = strings.ToLower(strings.TrimSpace(needle))
	for _, note := range notes {
		if strings.Contains(strings.ToLower(note), needle) {
			return true
		}
	}
	return false
}

func inlineImageGraphicsDebugEnabled() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(inlineImageGraphicsDebugEnvVar))) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func debugInlineImageEmission(emission inlineImageEmission) {
	if !inlineImageGraphicsDebugEnabled() {
		return
	}
	inlineImageGraphicsLogger.SetEnabled(true)
	inlineImageGraphicsLogger.IfEnabled().Debug("%s", emission.summary)
	inlineImageGraphicsLogger.IfEnabled().Debug("sequence=%s", escapeInlineImageControlSequence(emission.sequence))
}

func debugInlineImageEmissionError(summary string, err error) {
	if !inlineImageGraphicsDebugEnabled() {
		return
	}
	inlineImageGraphicsLogger.SetEnabled(true)
	inlineImageGraphicsLogger.IfEnabled().Debug("%s emit_error=%v", summary, err)
}

func escapeInlineImageControlSequence(raw string) string {
	if raw == "" {
		return ""
	}

	var builder strings.Builder
	builder.Grow(len(raw) + 16)
	for i := 0; i < len(raw); i++ {
		switch raw[i] {
		case 0x1b:
			builder.WriteString(`\x1b`)
		case '\a':
			builder.WriteString(`\a`)
		case '\r':
			builder.WriteString(`\r`)
		case '\n':
			builder.WriteString(`\n`)
		case '\t':
			builder.WriteString(`\t`)
		case '\\':
			builder.WriteString(`\\`)
		default:
			b := raw[i]
			if b < 0x20 || b == 0x7f {
				builder.WriteString(fmt.Sprintf("\\x%02x", b))
				continue
			}
			builder.WriteByte(b)
		}
	}
	return builder.String()
}
