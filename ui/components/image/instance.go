package image

import (
	"bytes"
	"encoding/base64"
	"fmt"
	stdimage "image"
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"strings"

	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/internal/proputil"
)

const (
	defaultCellPixelWidth  = 8
	defaultCellPixelHeight = 16
)

// Instance is the runtime entity for Image components.
type Instance struct {
	key         string
	id          string
	alt         string
	dataURI     string
	sourceImage stdimage.Image
	rgba        []byte
	pixelWidth  int
	pixelHeight int
	widthCells  int
	heightCells int
	imageStyle  style.Style
	bounds      [4]int
	dirty       bool
}

var (
	_ rtui.ComponentInstance      = (*Instance)(nil)
	_ rtui.PaintableInstance      = (*Instance)(nil)
	_ rtui.ScenePaintableInstance = (*Instance)(nil)
	_ interface {
		Measure(layout.Constraints) layout.Size
	} = (*Instance)(nil)
)

// NewInstance creates a new Image Instance from props.
func NewInstance(props rtui.Props) *Instance {
	inst := &Instance{dirty: true}
	inst.SetProps(props)
	return inst
}

func (inst *Instance) Key() string                        { return inst.key }
func (inst *Instance) SetKey(key string)                  { inst.key = key }
func (inst *Instance) IsDirty() bool                      { return inst.dirty }
func (inst *Instance) MarkClean()                         { inst.dirty = false }
func (inst *Instance) MarkDirty()                         { inst.dirty = true }
func (inst *Instance) Destroy()                           {}
func (inst *Instance) OnMount()                           {}
func (inst *Instance) OnUnmount()                         {}
func (inst *Instance) GetContext() *rtui.ComponentContext { return nil }
func (inst *Instance) Init(props rtui.Props)              { inst.SetProps(props) }

func (inst *Instance) SetProps(props rtui.Props) bool {
	inst.key = proputil.GetString(props, propKey, inst.key)
	inst.id = proputil.GetString(props, propID, inst.id)
	inst.alt = proputil.GetString(props, propAlt, "image")
	inst.dataURI = proputil.GetString(props, propDataURI, "")
	inst.sourceImage, _ = props[propSourceImage].(stdimage.Image)
	inst.rgba = bytesFromProp(props[propRGBA])
	inst.pixelWidth = proputil.GetInt(props, propPixelWidth, 0)
	inst.pixelHeight = proputil.GetInt(props, propPixelHeight, 0)
	inst.widthCells = proputil.GetInt(props, propWidthCells, 0)
	inst.heightCells = proputil.GetInt(props, propHeightCells, 0)
	inst.imageStyle = proputil.GetStyle(props, propStyle, style.Style{})

	if len(inst.rgba) == 0 && inst.sourceImage != nil {
		inst.rgba, inst.pixelWidth, inst.pixelHeight = rgbaFromImage(inst.sourceImage)
	}
	if len(inst.rgba) == 0 && inst.dataURI != "" {
		inst.rgba, inst.pixelWidth, inst.pixelHeight = rgbaFromDataURI(inst.dataURI)
	}

	inst.dirty = true
	return true
}

func (inst *Instance) GetProps() rtui.Props {
	return rtui.Props{
		propKey:         inst.key,
		propID:          inst.id,
		propAlt:         inst.alt,
		propDataURI:     inst.dataURI,
		propSourceImage: inst.sourceImage,
		propRGBA:        append([]byte(nil), inst.rgba...),
		propPixelWidth:  inst.pixelWidth,
		propPixelHeight: inst.pixelHeight,
		propWidthCells:  inst.widthCells,
		propHeightCells: inst.heightCells,
		propStyle:       inst.imageStyle,
	}
}

func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
	width, height := inst.cellSize()
	if constraints.MaxWidth > 0 && width > constraints.MaxWidth {
		width = constraints.MaxWidth
	}
	if constraints.MaxHeight > 0 && height > constraints.MaxHeight {
		height = constraints.MaxHeight
	}
	return layout.Size{Width: maxInt(1, width), Height: maxInt(1, height)}
}

func (inst *Instance) SetBounds(x, y, w, h int) {
	inst.bounds = [4]int{x, y, w, h}
}

// Paint renders an accessible text fallback. When a graphics presenter is
// active, the scene image layer is drawn over this area by the framework.
func (inst *Instance) Paint(x, y int) []paint.DrawCmd {
	label := compactFallback(fmt.Sprintf("[image: %s]", fallbackAlt(inst.alt)), inst.bounds[2])
	s := inst.imageStyle
	s = s.Foreground(theme.Muted())
	return []paint.DrawCmd{{
		X:     x,
		Y:     y,
		Text:  label,
		Style: s,
	}}
}

// SceneLayers contributes the raster payload to the framework image pipeline.
func (inst *Instance) SceneLayers() []paint.ImageLayer {
	if inst == nil || inst.pixelWidth <= 0 || inst.pixelHeight <= 0 || len(inst.rgba) == 0 {
		return nil
	}
	width, height := inst.cellSize()
	if inst.bounds[2] > 0 {
		width = inst.bounds[2]
	}
	if inst.bounds[3] > 0 {
		height = inst.bounds[3]
	}
	return []paint.ImageLayer{{
		ID:          inst.layerID(),
		Bounds:      paint.Rect{X: inst.bounds[0], Y: inst.bounds[1], Width: maxInt(1, width), Height: maxInt(1, height)},
		PixelWidth:  inst.pixelWidth,
		PixelHeight: inst.pixelHeight,
		RGBA:        append([]byte(nil), inst.rgba...),
		AltText:     fallbackAlt(inst.alt),
	}}
}

func (inst *Instance) cellSize() (int, int) {
	width := inst.widthCells
	height := inst.heightCells
	if width <= 0 && inst.pixelWidth > 0 {
		width = ceilDiv(inst.pixelWidth, defaultCellPixelWidth)
	}
	if height <= 0 && inst.pixelHeight > 0 {
		height = ceilDiv(inst.pixelHeight, defaultCellPixelHeight)
	}
	if width <= 0 {
		width = len("[image: " + fallbackAlt(inst.alt) + "]")
	}
	if height <= 0 {
		height = 1
	}
	return maxInt(1, width), maxInt(1, height)
}

func (inst *Instance) layerID() string {
	if strings.TrimSpace(inst.id) != "" {
		return inst.id
	}
	if strings.TrimSpace(inst.key) != "" {
		return inst.key + ":image"
	}
	return "image:" + strings.ReplaceAll(strings.ToLower(fallbackAlt(inst.alt)), " ", "-")
}

func bytesFromProp(value any) []byte {
	if value == nil {
		return nil
	}
	if data, ok := value.([]byte); ok {
		return append([]byte(nil), data...)
	}
	return nil
}

func rgbaFromDataURI(dataURI string) ([]byte, int, int) {
	media, payload, ok := strings.Cut(dataURI, ",")
	if !ok || !strings.Contains(strings.ToLower(media), ";base64") {
		return nil, 0, 0
	}
	if strings.Contains(strings.ToLower(media), "image/svg") {
		return nil, 0, 0
	}
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(payload))
	if err != nil {
		return nil, 0, 0
	}
	img, _, err := stdimage.Decode(bytes.NewReader(decoded))
	if err != nil {
		return nil, 0, 0
	}
	return rgbaFromImage(img)
}

func rgbaFromImage(img stdimage.Image) ([]byte, int, int) {
	if img == nil {
		return nil, 0, 0
	}
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	if width <= 0 || height <= 0 {
		return nil, 0, 0
	}
	rgba := stdimage.NewRGBA(stdimage.Rect(0, 0, width, height))
	draw.Draw(rgba, rgba.Bounds(), img, bounds.Min, draw.Src)
	return append([]byte(nil), rgba.Pix...), width, height
}

func fallbackAlt(alt string) string {
	if strings.TrimSpace(alt) == "" {
		return "image"
	}
	return strings.TrimSpace(alt)
}

func compactFallback(value string, width int) string {
	if width <= 0 || len(value) <= width {
		return value
	}
	if width <= 1 {
		return value[:width]
	}
	return value[:width-1] + "."
}

func ceilDiv(value, divisor int) int {
	if divisor <= 0 {
		return value
	}
	return (value + divisor - 1) / divisor
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
