package linechart

import (
	"fmt"
	"strings"

	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime/paint"
	runtimeplatform "github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/ui/components/charts/internal/axis"
	"github.com/wwsheng009/mint/ui/components/charts/internal/palette"
	"github.com/wwsheng009/mint/ui/components/charts/internal/scale"
)

const (
	plotImageDefaultPixelsPerCellX = 8
	plotImageDefaultPixelsPerCellY = 12
	plotImageLineThickness         = 1
	plotImagePointRadius           = 3
	plotImageGridThickness         = 1
)

type imagePoint struct {
	X int
	Y int
}

type imageColor struct {
	R byte
	G byte
	B byte
	A byte
}

type plotImage struct {
	width  int
	height int
	rgba   []byte
}

type plotImageRect struct {
	X      int
	Y      int
	Width  int
	Height int
}

func newPlotImage(width, height int, bg imageColor) *plotImage {
	img := &plotImage{
		width:  width,
		height: height,
		rgba:   make([]byte, width*height*4),
	}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.setPixel(x, y, bg)
		}
	}
	return img
}

func (img *plotImage) setPixel(x, y int, color imageColor) {
	if img == nil || x < 0 || y < 0 || x >= img.width || y >= img.height {
		return
	}
	offset := (y*img.width + x) * 4
	img.rgba[offset] = color.R
	img.rgba[offset+1] = color.G
	img.rgba[offset+2] = color.B
	img.rgba[offset+3] = color.A
}

func (img *plotImage) drawHorizontalLine(y int, color imageColor, thickness int) {
	if img == nil {
		return
	}
	half := maxInt(0, thickness/2)
	for row := y - half; row <= y+half; row++ {
		for x := 0; x < img.width; x++ {
			img.setPixel(x, row, color)
		}
	}
}

func (img *plotImage) drawDisc(cx, cy, radius int, color imageColor) {
	if img == nil || radius < 0 {
		return
	}
	r2 := radius * radius
	for y := cy - radius; y <= cy+radius; y++ {
		for x := cx - radius; x <= cx+radius; x++ {
			dx := x - cx
			dy := y - cy
			if dx*dx+dy*dy <= r2 {
				img.setPixel(x, y, color)
			}
		}
	}
}

func (img *plotImage) drawLine(start, end imagePoint, color imageColor, thickness int) {
	if img == nil {
		return
	}
	x0, y0 := start.X, start.Y
	x1, y1 := end.X, end.Y

	dx := absIntImage(x1 - x0)
	dy := -absIntImage(y1 - y0)
	sx := stepSignImage(x0, x1)
	sy := stepSignImage(y0, y1)
	err := dx + dy
	radius := maxInt(0, thickness/2)

	for {
		img.drawDisc(x0, y0, radius, color)
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x0 += sx
		}
		if e2 <= dx {
			err += dx
			y0 += sy
		}
	}
}

// SceneLayers is an optional scene/image hook consumed by the runtime scene path.
func (inst *Instance) SceneLayers() []paint.ImageLayer {
	if inst == nil || !SupportsImagePlotBackend() || inst.renderBackend != RenderBackendImagePlot {
		return nil
	}

	layer, ok := inst.renderPlotImageLayer()
	if !ok {
		return nil
	}
	return []paint.ImageLayer{layer}
}

func (inst *Instance) renderPlotImageLayer() (paint.ImageLayer, bool) {
	plotWidth := inst.plotWidth()
	plotHeight := inst.plotHeight()
	seriesList := inst.resolvedSeries()
	if plotWidth <= 0 || plotHeight <= 0 || len(seriesList) == 0 {
		return paint.ImageLayer{}, false
	}

	pixelsPerCellX, pixelsPerCellY := plotImagePixelsPerCell()
	pixelWidth := maxInt(1, plotWidth*pixelsPerCellX)
	pixelHeight := maxInt(1, plotHeight*pixelsPerCellY)
	rgba := inst.renderPlotRGBA(seriesList, pixelWidth, pixelHeight)
	if len(rgba) == 0 {
		return paint.ImageLayer{}, false
	}

	headerRows := len(inst.buildHeaderFrame().Rows())
	layer := paint.ImageLayer{
		ID:          inst.sceneLayerID(),
		Bounds:      paint.Rect{X: inst.bounds[0], Y: inst.bounds[1] + headerRows, Width: plotWidth, Height: plotHeight},
		PixelWidth:  pixelWidth,
		PixelHeight: pixelHeight,
		RGBA:        rgba,
		AltText:     inst.sceneAltText(),
	}
	return layer, true
}

func (inst *Instance) renderPlotRGBA(seriesList []Series, pixelWidth, pixelHeight int) []byte {
	bg := inst.plotImageBackgroundColor()
	img := newPlotImage(pixelWidth, pixelHeight, bg)
	contentRect := plotImageContentRect(pixelWidth, pixelHeight)

	if inst.showGrid {
		gridRows := axis.GridRows(inst.plotHeight(), 3)
		if inst.hasAxisLabels() && len(gridRows) > 1 {
			gridRows = gridRows[:len(gridRows)-1]
		}
		gridColor := colorFromStyle(palette.GridColor(), imageColor{R: 76, G: 86, B: 106, A: 255})
		for _, row := range gridRows {
			y := contentRect.Y + mapPlotRowToPixel(row, inst.plotHeight(), contentRect.Height)
			img.drawHorizontalLine(y, gridColor, plotImageGridThickness)
		}
	}

	minVal, maxVal := inst.seriesDomain(seriesList)
	yScale := scale.NewLinear(minVal, maxVal, contentRect.Y+contentRect.Height-1, contentRect.Y)

	for index, series := range seriesList {
		if len(series.Data) == 0 {
			continue
		}

		sampled := resampleForContinuity(series.Data, inst.sampleCount())
		rows := inst.seriesRows(sampled, yScale, pixelHeight)
		xBand := scale.NewBand(len(rows), contentRect.X, contentRect.X+contentRect.Width-1)
		points := make([]imagePoint, 0, len(rows))
		for pointIndex, row := range rows {
			xPos := xBand.Position(pointIndex)
			if xPos >= pixelWidth {
				break
			}
			points = append(points, imagePoint{X: xPos, Y: row})
		}
		if len(points) == 0 {
			continue
		}

		lineStyle := inst.resolveSeriesStyle(index, len(seriesList), series)
		lineColor := colorFromStyle(lineStyle.FG, colorFromStyle(palette.SeriesColor(index), imageColor{R: 136, G: 192, B: 208, A: 255}))
		for i := 1; i < len(points); i++ {
			img.drawLine(points[i-1], points[i], lineColor, plotImageLineThickness)
		}
		if inst.showPoints {
			for _, point := range points {
				img.drawDisc(point.X, point.Y, plotImagePointRadius, lineColor)
			}
		}
	}

	return img.rgba
}

func plotImagePixelsPerCell() (int, int) {
	if width, height, ok := runtimeplatform.GraphicsCellPixelsFromEnv(); ok {
		return width, height
	}
	return plotImageDefaultPixelsPerCellX, plotImageDefaultPixelsPerCellY
}

func plotImageContentRect(pixelWidth, pixelHeight int) plotImageRect {
	pixelsPerCellX, pixelsPerCellY := plotImagePixelsPerCell()
	paddingX := maxInt(1, pixelsPerCellX/4)
	paddingY := maxInt(1, pixelsPerCellY/4)

	if pixelWidth <= paddingX*2+1 {
		paddingX = 0
	}
	if pixelHeight <= paddingY*2+1 {
		paddingY = 0
	}

	return plotImageRect{
		X:      paddingX,
		Y:      paddingY,
		Width:  maxInt(1, pixelWidth-paddingX*2),
		Height: maxInt(1, pixelHeight-paddingY*2),
	}
}

func (inst *Instance) plotImageBackgroundColor() imageColor {
	return colorFromStyle(inst.plotImageBackgroundStyleColor(), imageColor{R: 0, G: 0, B: 0, A: 255})
}

func (inst *Instance) plotImageBackgroundStyleColor() style.Color {
	if inst != nil && inst.chartStyle.BG != style.NoColor {
		return inst.chartStyle.BG
	}
	return style.Black
}

func (inst *Instance) sceneLayerID() string {
	if key := strings.TrimSpace(inst.key); key != "" {
		return key + ":plot-image"
	}
	if title := strings.TrimSpace(inst.title); title != "" {
		return "linechart:" + strings.ReplaceAll(strings.ToLower(title), " ", "-") + ":plot-image"
	}
	return "linechart:plot-image"
}

func (inst *Instance) sceneAltText() string {
	if title := strings.TrimSpace(inst.title); title != "" {
		return fmt.Sprintf("%s plot image", title)
	}
	return "line chart plot image"
}

func mapPlotRowToPixel(row, plotHeight, pixelHeight int) int {
	if plotHeight <= 1 || pixelHeight <= 1 {
		return 0
	}
	return clampInt(int(float64(row)*float64(pixelHeight-1)/float64(plotHeight-1)), 0, pixelHeight-1)
}

func colorFromStyle(c style.Color, fallback imageColor) imageColor {
	if c == style.NoColor {
		return fallback
	}
	return colorFromTheme(theme.ParseColor(string(c)), fallback)
}

func colorFromTheme(c theme.Color, fallback imageColor) imageColor {
	switch c.Type {
	case theme.ColorRGB, theme.ColorHex:
		r, g, b := c.RGBValue()
		return imageColor{R: byte(r), G: byte(g), B: byte(b), A: 255}
	case theme.ColorNamed:
		if name, ok := c.Value.(string); ok {
			if rgb, ok := namedImageColors[name]; ok {
				return rgb
			}
		}
	}
	return fallback
}

var namedImageColors = map[string]imageColor{
	"black":          {R: 0, G: 0, B: 0, A: 255},
	"red":            {R: 205, G: 49, B: 49, A: 255},
	"green":          {R: 13, G: 188, B: 121, A: 255},
	"yellow":         {R: 229, G: 229, B: 16, A: 255},
	"blue":           {R: 36, G: 114, B: 200, A: 255},
	"magenta":        {R: 188, G: 63, B: 188, A: 255},
	"cyan":           {R: 17, G: 168, B: 205, A: 255},
	"white":          {R: 229, G: 229, B: 229, A: 255},
	"bright-black":   {R: 102, G: 102, B: 102, A: 255},
	"bright-red":     {R: 241, G: 76, B: 76, A: 255},
	"bright-green":   {R: 35, G: 209, B: 139, A: 255},
	"bright-yellow":  {R: 245, G: 245, B: 67, A: 255},
	"bright-blue":    {R: 59, G: 142, B: 234, A: 255},
	"bright-magenta": {R: 214, G: 112, B: 214, A: 255},
	"bright-cyan":    {R: 41, G: 184, B: 219, A: 255},
	"bright-white":   {R: 255, G: 255, B: 255, A: 255},
}

func absIntImage(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

func stepSignImage(a, b int) int {
	switch {
	case a < b:
		return 1
	case a > b:
		return -1
	default:
		return 0
	}
}
