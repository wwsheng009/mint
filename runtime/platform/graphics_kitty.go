package platform

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/wwsheng009/mint/internal/log"
)

type kittyImageObject struct {
	handle  string
	kittyID int
	request DrawImageRequest
}

// KittyGraphicsPresenter is a minimal Phase 1 graphics presenter for Kitty.
// It focuses on object lifecycle and protocol emission, not advanced caching or
// diffing.
type KittyGraphicsPresenter struct {
	writer      io.Writer
	caps        GraphicsCapabilities
	objects     map[string]kittyImageObject
	nextHandle  int
	nextKittyID int
}

const kittyBase64ChunkSize = 4096
const kittyGraphicsDebugEnvVar = "MINT_DEBUG_KITTY_GRAPHICS"

var kittyGraphicsLogger = log.NewLoggerWithEnv("KittyGraphics", "PLATFORM", kittyGraphicsDebugEnvVar)

type kittyEmission struct {
	sequence string
	summary  string
}

// NewKittyGraphicsPresenter creates a new Kitty presenter. A nil writer is
// allowed for tests or dry-run usage; lifecycle state is still tracked.
func NewKittyGraphicsPresenter(writer io.Writer, caps GraphicsCapabilities) *KittyGraphicsPresenter {
	kittyGraphicsLogger.SetEnabled(kittyGraphicsDebugEnabled())
	return &KittyGraphicsPresenter{
		writer:      writer,
		caps:        normalizeKittyCapabilities(caps),
		objects:     make(map[string]kittyImageObject),
		nextHandle:  1,
		nextKittyID: 1,
	}
}

// Capabilities reports the normalized presenter capabilities.
func (p *KittyGraphicsPresenter) Capabilities() GraphicsCapabilities {
	if p == nil {
		return normalizeKittyCapabilities(GraphicsCapabilities{})
	}
	return p.caps
}

// Present creates a new Kitty image object and returns its stable handle.
func (p *KittyGraphicsPresenter) Present(req DrawImageRequest) (string, error) {
	if p == nil {
		return "", fmt.Errorf("kitty presenter is nil")
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
	obj := kittyImageObject{
		handle:  handle,
		kittyID: p.nextKittyID,
		request: cloneDrawImageRequest(req),
	}
	p.nextKittyID++
	if err := p.emit(p.presentEmission(obj, "present")); err != nil {
		return "", err
	}
	p.objects[handle] = obj
	return handle, nil
}

// Replace updates an existing Kitty image object in place.
func (p *KittyGraphicsPresenter) Replace(id string, req DrawImageRequest) error {
	if p == nil {
		return fmt.Errorf("kitty presenter is nil")
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

// Delete removes a single Kitty image object.
func (p *KittyGraphicsPresenter) Delete(id string) error {
	if p == nil {
		return fmt.Errorf("kitty presenter is nil")
	}
	if id == "" {
		return fmt.Errorf("image handle is required")
	}
	obj, ok := p.objects[id]
	if !ok {
		return fmt.Errorf("image handle %q not found", id)
	}
	if err := p.emit(p.deleteEmission(obj.kittyID)); err != nil {
		return err
	}
	delete(p.objects, id)
	return nil
}

// Clear deletes all tracked Kitty image objects.
func (p *KittyGraphicsPresenter) Clear() error {
	if p == nil {
		return fmt.Errorf("kitty presenter is nil")
	}
	if err := p.emit(p.clearEmission()); err != nil {
		return err
	}
	for handle := range p.objects {
		delete(p.objects, handle)
	}
	return nil
}

func (p *KittyGraphicsPresenter) replaceObject(existing kittyImageObject, req DrawImageRequest) error {
	req.ID = existing.handle
	updated := kittyImageObject{
		handle:  existing.handle,
		kittyID: existing.kittyID,
		request: cloneDrawImageRequest(req),
	}
	if err := p.emit(p.presentEmission(updated, "replace")); err != nil {
		return err
	}
	p.objects[existing.handle] = updated
	return nil
}

func (p *KittyGraphicsPresenter) newHandle() string {
	handle := fmt.Sprintf("kitty-img-%d", p.nextHandle)
	p.nextHandle++
	return handle
}

func (p *KittyGraphicsPresenter) emit(emission kittyEmission) error {
	if p.writer == nil {
		return nil
	}
	_, err := io.WriteString(p.writer, emission.sequence)
	if err != nil {
		debugKittyEmissionError(emission.summary, err)
		return err
	}
	debugKittyEmission(emission)
	return nil
}

func (p *KittyGraphicsPresenter) presentEmission(obj kittyImageObject, op string) kittyEmission {
	payload := base64.StdEncoding.EncodeToString(obj.request.RGBA)
	row := obj.request.CellY + 1
	col := obj.request.CellX + 1

	chunks := splitKittyPayload(payload, kittyBase64ChunkSize)
	var builder strings.Builder
	builder.WriteString("\x1b[s")
	builder.WriteString(fmt.Sprintf("\x1b[%d;%dH", row, col))
	for index, chunk := range chunks {
		more := 0
		if index < len(chunks)-1 {
			more = 1
		}
		if index == 0 {
			builder.WriteString(fmt.Sprintf(
				"\x1b_Ga=T,q=2,i=%d,f=32,s=%d,v=%d,c=%d,r=%d,C=1,m=%d;%s\x1b\\",
				obj.kittyID,
				obj.request.PixelWidth,
				obj.request.PixelHeight,
				obj.request.CellWidth,
				obj.request.CellHeight,
				more,
				chunk,
			))
			continue
		}
		builder.WriteString(fmt.Sprintf("\x1b_Gq=2,m=%d;%s\x1b\\", more, chunk))
	}
	builder.WriteString("\x1b[u")
	return kittyEmission{
		sequence: builder.String(),
		summary: fmt.Sprintf(
			"%s handle=%s kitty_id=%d cell=(x=%d,y=%d) cursor=(row=%d,col=%d) cells=%dx%d pixels=%dx%d chunks=%d payload_base64=%d source=%s reliable=%t",
			op,
			obj.handle,
			obj.kittyID,
			obj.request.CellX,
			obj.request.CellY,
			row,
			col,
			obj.request.CellWidth,
			obj.request.CellHeight,
			obj.request.PixelWidth,
			obj.request.PixelHeight,
			len(chunks),
			len(payload),
			p.caps.ProbeSource,
			p.caps.Reliable,
		),
	}
}

func splitKittyPayload(payload string, chunkSize int) []string {
	if payload == "" {
		return []string{""}
	}
	if chunkSize <= 0 {
		return []string{payload}
	}

	chunks := make([]string, 0, (len(payload)+chunkSize-1)/chunkSize)
	for start := 0; start < len(payload); start += chunkSize {
		end := start + chunkSize
		if end > len(payload) {
			end = len(payload)
		}
		chunks = append(chunks, payload[start:end])
	}
	return chunks
}

func (p *KittyGraphicsPresenter) deleteEmission(kittyID int) kittyEmission {
	return kittyEmission{
		sequence: fmt.Sprintf("\x1b_Ga=d,d=I,i=%d\x1b\\", kittyID),
		summary:  fmt.Sprintf("delete kitty_id=%d scope=image", kittyID),
	}
}

func (p *KittyGraphicsPresenter) clearEmission() kittyEmission {
	return kittyEmission{
		sequence: "\x1b_Ga=d,d=A\x1b\\",
		summary:  "clear scope=all",
	}
}

func normalizeKittyCapabilities(caps GraphicsCapabilities) GraphicsCapabilities {
	if caps.Mode == GraphicsModeNone {
		caps.Mode = GraphicsModeKitty
	}
	caps.Mode = GraphicsModeKitty
	caps.PresentationModel = GraphicsPresentationModelOverlay
	caps.SupportsPlacement = true
	caps.SupportsReplace = true
	caps.SupportsDelete = true
	if caps.ProbeSource == "" {
		caps.ProbeSource = "kitty-presenter"
	}
	return caps
}

func validateDrawImageRequest(req DrawImageRequest) error {
	if req.PixelWidth <= 0 {
		return fmt.Errorf("pixel width must be > 0")
	}
	if req.PixelHeight <= 0 {
		return fmt.Errorf("pixel height must be > 0")
	}
	if req.CellWidth <= 0 {
		return fmt.Errorf("cell width must be > 0")
	}
	if req.CellHeight <= 0 {
		return fmt.Errorf("cell height must be > 0")
	}
	if len(req.RGBA) == 0 {
		return fmt.Errorf("rgba payload is required")
	}
	return nil
}

func cloneDrawImageRequest(req DrawImageRequest) DrawImageRequest {
	out := req
	if len(req.RGBA) > 0 {
		out.RGBA = append([]byte(nil), req.RGBA...)
	}
	return out
}

func kittyGraphicsDebugEnabled() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(kittyGraphicsDebugEnvVar))) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func debugKittyEmission(emission kittyEmission) {
	if !kittyGraphicsDebugEnabled() {
		return
	}
	kittyGraphicsLogger.SetEnabled(true)
	kittyGraphicsLogger.IfEnabled().Debug("%s", emission.summary)
	kittyGraphicsLogger.IfEnabled().Debug("sequence=%s", escapeKittyControlSequence(emission.sequence))
}

func debugKittyEmissionError(summary string, err error) {
	if !kittyGraphicsDebugEnabled() {
		return
	}
	kittyGraphicsLogger.SetEnabled(true)
	kittyGraphicsLogger.IfEnabled().Debug("%s emit_error=%v", summary, err)
}

func escapeKittyControlSequence(raw string) string {
	if raw == "" {
		return ""
	}

	var builder strings.Builder
	builder.Grow(len(raw) + 16)
	for i := 0; i < len(raw); i++ {
		switch raw[i] {
		case 0x1b:
			builder.WriteString(`\x1b`)
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
