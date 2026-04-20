package platform

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/wwsheng009/mint/internal/log"
)

const (
	sixelMaxPaletteSize       = 256
	sixelGraphicsDebugEnvVar  = "MINT_DEBUG_SIXEL_GRAPHICS"
	sixelTransparentThreshold = 8
)

var sixelGraphicsLogger = log.NewLoggerWithEnv("SixelGraphics", "PLATFORM", sixelGraphicsDebugEnvVar)

type sixelImageObject struct {
	handle  string
	request DrawImageRequest
}

type sixelEmission struct {
	sequence string
	summary  string
}

type sixelColor struct {
	R byte
	G byte
	B byte
}

type sixelBitmap struct {
	width   int
	height  int
	palette []sixelColor
	indices []uint16
}

// SixelGraphicsPresenter renders image layers using the SIXEL raster format.
// It is optimized for simple chart-like images and relies on full-frame text
// repaints when protocol-level object deletion is unavailable.
type SixelGraphicsPresenter struct {
	writer     io.Writer
	caps       GraphicsCapabilities
	objects    map[string]sixelImageObject
	nextHandle int
}

// NewSixelGraphicsPresenter creates a new SIXEL presenter.
func NewSixelGraphicsPresenter(writer io.Writer, caps GraphicsCapabilities) *SixelGraphicsPresenter {
	sixelGraphicsLogger.SetEnabled(sixelGraphicsDebugEnabled())
	return &SixelGraphicsPresenter{
		writer:     writer,
		caps:       normalizeSixelCapabilities(caps),
		objects:    make(map[string]sixelImageObject),
		nextHandle: 1,
	}
}

// Capabilities reports normalized SIXEL capabilities.
func (p *SixelGraphicsPresenter) Capabilities() GraphicsCapabilities {
	if p == nil {
		return normalizeSixelCapabilities(GraphicsCapabilities{})
	}
	return p.caps
}

// Present draws or replaces a SIXEL image at the requested cell position.
func (p *SixelGraphicsPresenter) Present(req DrawImageRequest) (string, error) {
	if p == nil {
		return "", fmt.Errorf("sixel presenter is nil")
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
	obj := sixelImageObject{
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

// Replace updates an existing SIXEL image by redrawing it.
func (p *SixelGraphicsPresenter) Replace(id string, req DrawImageRequest) error {
	if p == nil {
		return fmt.Errorf("sixel presenter is nil")
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

// Delete removes tracked lifecycle state. SIXEL has no stable object deletion
// primitive, so actual screen cleanup is handled via full terminal repaints.
func (p *SixelGraphicsPresenter) Delete(id string) error {
	if p == nil {
		return fmt.Errorf("sixel presenter is nil")
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

// Clear clears the terminal surface so the caller can repaint text before
// drawing the next frame. This is used for non-object protocols such as SIXEL.
func (p *SixelGraphicsPresenter) Clear() error {
	if p == nil {
		return fmt.Errorf("sixel presenter is nil")
	}
	if len(p.objects) == 0 {
		return nil
	}
	if err := p.emit(sixelEmission{
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

func (p *SixelGraphicsPresenter) replaceObject(existing sixelImageObject, req DrawImageRequest) error {
	req.ID = existing.handle
	updated := sixelImageObject{
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

func (p *SixelGraphicsPresenter) newHandle() string {
	handle := fmt.Sprintf("sixel-img-%d", p.nextHandle)
	p.nextHandle++
	return handle
}

func (p *SixelGraphicsPresenter) emit(emission sixelEmission) error {
	if p.writer == nil {
		return nil
	}
	_, err := io.WriteString(p.writer, emission.sequence)
	if err != nil {
		debugSixelEmissionError(emission.summary, err)
		return err
	}
	debugSixelEmission(emission)
	return nil
}

func (p *SixelGraphicsPresenter) presentEmission(obj sixelImageObject, op string) (sixelEmission, error) {
	bitmap, targetWidth, targetHeight, err := buildSixelBitmap(obj.request, p.caps)
	if err != nil {
		return sixelEmission{}, err
	}

	row := obj.request.CellY + 1
	col := obj.request.CellX + 1

	var builder strings.Builder
	builder.WriteString("\x1b[s")
	builder.WriteString(fmt.Sprintf("\x1b[%d;%dH", row, col))
	builder.WriteString("\x1b[?80l")
	builder.WriteString("\x1bPq")
	builder.WriteString(fmt.Sprintf("\"1;1;%d;%d", bitmap.width, bitmap.height))
	for index, color := range bitmap.palette {
		builder.WriteString(fmt.Sprintf(
			"#%d;2;%d;%d;%d",
			index,
			sixelColorPercent(color.R),
			sixelColorPercent(color.G),
			sixelColorPercent(color.B),
		))
	}
	for band := 0; band < (bitmap.height+5)/6; band++ {
		if band > 0 {
			builder.WriteByte('-')
		}
		wroteColor := false
		for index := range bitmap.palette {
			rowData, hasPixels := sixelBandData(bitmap, band, uint16(index))
			if !hasPixels {
				continue
			}
			if wroteColor {
				builder.WriteByte('$')
			}
			builder.WriteByte('#')
			builder.WriteString(strconv.Itoa(index))
			builder.WriteString(rowData)
			wroteColor = true
		}
	}
	builder.WriteString("\x1b\\")
	builder.WriteString("\x1b[u")

	return sixelEmission{
		sequence: builder.String(),
		summary: fmt.Sprintf(
			"%s handle=%s cell=(x=%d,y=%d) cursor=(row=%d,col=%d) cells=%dx%d src_pixels=%dx%d target_pixels=%dx%d palette=%d source=%s reliable=%t",
			op,
			obj.handle,
			obj.request.CellX,
			obj.request.CellY,
			row,
			col,
			obj.request.CellWidth,
			obj.request.CellHeight,
			obj.request.PixelWidth,
			obj.request.PixelHeight,
			targetWidth,
			targetHeight,
			len(bitmap.palette),
			p.caps.ProbeSource,
			p.caps.Reliable,
		),
	}, nil
}

func buildSixelBitmap(req DrawImageRequest, caps GraphicsCapabilities) (sixelBitmap, int, int, error) {
	expected := req.PixelWidth * req.PixelHeight * 4
	if expected <= 0 {
		return sixelBitmap{}, 0, 0, fmt.Errorf("invalid source pixel dimensions %dx%d", req.PixelWidth, req.PixelHeight)
	}
	if len(req.RGBA) < expected {
		return sixelBitmap{}, 0, 0, fmt.Errorf("rgba payload too short: got %d want >= %d", len(req.RGBA), expected)
	}

	targetWidth := req.PixelWidth
	targetHeight := req.PixelHeight
	if caps.CellPixelsKnown() && req.CellWidth > 0 && req.CellHeight > 0 {
		targetWidth = req.CellWidth * caps.CellPixelWidth
		targetHeight = req.CellHeight * caps.CellPixelHeight
	}
	if targetWidth <= 0 {
		targetWidth = req.PixelWidth
	}
	if targetHeight <= 0 {
		targetHeight = req.PixelHeight
	}

	scaled := scaleRGBAForSixel(req.RGBA[:expected], req.PixelWidth, req.PixelHeight, targetWidth, targetHeight)
	pixels := compositeSixelPixels(scaled)

	palette, indices, ok := buildExactSixelPalette(pixels)
	if !ok {
		palette, indices = buildRGB332SixelPalette(pixels)
	}

	return sixelBitmap{
		width:   targetWidth,
		height:  targetHeight,
		palette: palette,
		indices: indices,
	}, targetWidth, targetHeight, nil
}

func buildExactSixelPalette(pixels []sixelColor) ([]sixelColor, []uint16, bool) {
	palette := make([]sixelColor, 0, 16)
	lookup := make(map[uint32]uint16, 16)
	indices := make([]uint16, len(pixels))

	for i, pixel := range pixels {
		key := uint32(pixel.R)<<16 | uint32(pixel.G)<<8 | uint32(pixel.B)
		index, ok := lookup[key]
		if !ok {
			if len(palette) >= sixelMaxPaletteSize {
				return nil, nil, false
			}
			index = uint16(len(palette))
			lookup[key] = index
			palette = append(palette, pixel)
		}
		indices[i] = index
	}

	return palette, indices, true
}

func buildRGB332SixelPalette(pixels []sixelColor) ([]sixelColor, []uint16) {
	used := make([]bool, sixelMaxPaletteSize)
	quantized := make([]uint16, len(pixels))

	for i, pixel := range pixels {
		index := uint16((pixel.R>>5)<<5 | (pixel.G>>5)<<2 | (pixel.B >> 6))
		quantized[i] = index
		used[index] = true
	}

	palette := make([]sixelColor, 0, 64)
	remap := make([]uint16, sixelMaxPaletteSize)
	for index := 0; index < len(used); index++ {
		if !used[index] {
			continue
		}
		remap[index] = uint16(len(palette))
		palette = append(palette, rgb332ToSixelColor(byte(index)))
	}

	indices := make([]uint16, len(quantized))
	for i, index := range quantized {
		indices[i] = remap[index]
	}

	return palette, indices
}

func compositeSixelPixels(rgba []byte) []sixelColor {
	if len(rgba) == 0 {
		return nil
	}

	background := sixelColor{R: rgba[0], G: rgba[1], B: rgba[2]}
	pixels := make([]sixelColor, 0, len(rgba)/4)
	for offset := 0; offset+3 < len(rgba); offset += 4 {
		src := sixelColor{R: rgba[offset], G: rgba[offset+1], B: rgba[offset+2]}
		alpha := rgba[offset+3]
		pixels = append(pixels, compositeSixelColor(src, background, alpha))
	}
	return pixels
}

func compositeSixelColor(src, background sixelColor, alpha byte) sixelColor {
	if alpha >= 255-sixelTransparentThreshold {
		return src
	}
	if alpha <= sixelTransparentThreshold {
		return background
	}

	a := int(alpha)
	inv := 255 - a
	return sixelColor{
		R: byte((int(src.R)*a + int(background.R)*inv + 127) / 255),
		G: byte((int(src.G)*a + int(background.G)*inv + 127) / 255),
		B: byte((int(src.B)*a + int(background.B)*inv + 127) / 255),
	}
}

func rgb332ToSixelColor(index byte) sixelColor {
	r := int((index >> 5) & 0x07)
	g := int((index >> 2) & 0x07)
	b := int(index & 0x03)
	return sixelColor{
		R: byte((r * 255) / 7),
		G: byte((g * 255) / 7),
		B: byte((b * 255) / 3),
	}
}

func scaleRGBAForSixel(src []byte, srcWidth, srcHeight, dstWidth, dstHeight int) []byte {
	if srcWidth <= 0 || srcHeight <= 0 || dstWidth <= 0 || dstHeight <= 0 {
		return nil
	}
	if srcWidth == dstWidth && srcHeight == dstHeight {
		return append([]byte(nil), src...)
	}

	dst := make([]byte, dstWidth*dstHeight*4)
	for y := 0; y < dstHeight; y++ {
		srcY := y * srcHeight / dstHeight
		for x := 0; x < dstWidth; x++ {
			srcX := x * srcWidth / dstWidth
			srcOffset := (srcY*srcWidth + srcX) * 4
			dstOffset := (y*dstWidth + x) * 4
			copy(dst[dstOffset:dstOffset+4], src[srcOffset:srcOffset+4])
		}
	}
	return dst
}

func sixelBandData(bitmap sixelBitmap, band int, paletteIndex uint16) (string, bool) {
	var builder strings.Builder
	runCount := 0
	runChar := byte('?')
	hasPixels := false

	for x := 0; x < bitmap.width; x++ {
		bits := 0
		for dy := 0; dy < 6; dy++ {
			y := band*6 + dy
			if y >= bitmap.height {
				break
			}
			if bitmap.indices[y*bitmap.width+x] == paletteIndex {
				bits |= 1 << dy
				hasPixels = true
			}
		}

		ch := byte(63 + bits)
		if runCount == 0 {
			runChar = ch
			runCount = 1
			continue
		}
		if ch == runChar {
			runCount++
			continue
		}
		writeSixelRun(&builder, runChar, runCount)
		runChar = ch
		runCount = 1
	}

	if runCount > 0 {
		writeSixelRun(&builder, runChar, runCount)
	}
	if !hasPixels {
		return "", false
	}
	return builder.String(), true
}

func writeSixelRun(builder *strings.Builder, ch byte, count int) {
	if builder == nil || count <= 0 {
		return
	}
	if count >= 4 {
		builder.WriteByte('!')
		builder.WriteString(strconv.Itoa(count))
		builder.WriteByte(ch)
		return
	}
	for i := 0; i < count; i++ {
		builder.WriteByte(ch)
	}
}

func sixelColorPercent(value byte) int {
	return (int(value)*100 + 127) / 255
}

func normalizeSixelCapabilities(caps GraphicsCapabilities) GraphicsCapabilities {
	if caps.Mode == GraphicsModeNone {
		caps.Mode = GraphicsModeSixel
	}
	caps.Mode = GraphicsModeSixel
	caps.PresentationModel = GraphicsPresentationModelTerminalFrame
	caps.SupportsPlacement = true
	caps.SupportsReplace = true
	caps.SupportsDelete = false
	if caps.ProbeSource == "" {
		caps.ProbeSource = "sixel-presenter"
	}
	return caps
}

func sixelGraphicsDebugEnabled() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(sixelGraphicsDebugEnvVar))) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func debugSixelEmission(emission sixelEmission) {
	if !sixelGraphicsDebugEnabled() {
		return
	}
	sixelGraphicsLogger.SetEnabled(true)
	sixelGraphicsLogger.IfEnabled().Debug("%s", emission.summary)
	sixelGraphicsLogger.IfEnabled().Debug("sequence=%s", escapeSixelControlSequence(emission.sequence))
}

func debugSixelEmissionError(summary string, err error) {
	if !sixelGraphicsDebugEnabled() {
		return
	}
	sixelGraphicsLogger.SetEnabled(true)
	sixelGraphicsLogger.IfEnabled().Debug("%s emit_error=%v", summary, err)
}

func escapeSixelControlSequence(raw string) string {
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
