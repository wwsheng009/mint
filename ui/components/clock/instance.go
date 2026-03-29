package clock

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime/animation"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/internal/proputil"
)

const (
	minRadius            = 3
	smoothTickInterval   = time.Second / 30
	discreteTickInterval = time.Second
)

// Instance is the runtime entity for Clock components.
type Instance struct {
	key             string
	shape           DialShape
	radius          int
	radiusY         int
	cellAspectX     float64
	live            bool
	timeValue       time.Time
	location        *time.Location
	showSecondHand  bool
	smoothSecond    bool
	showDigital     bool
	preset          Preset
	handStyle       HandRenderStyle
	dialStyle       style.Style
	tickStyle       style.Style
	centerStyle     style.Style
	digitalStyle    style.Style
	hourHandStyle   style.Style
	minuteHandStyle style.Style
	secondHandStyle style.Style
	clockStyle      style.Style
	displayTime     time.Time
	loop            *animation.LoopDriver
	dirty           bool
}

var (
	_ rtui.ComponentInstance = (*Instance)(nil)
	_ rtui.PaintableInstance = (*Instance)(nil)
	_ rtui.TickableInstance  = (*Instance)(nil)
	_ interface {
		Measure(layout.Constraints) layout.Size
	} = (*Instance)(nil)
)

// NewInstance creates a new Clock instance from props.
func NewInstance(props rtui.Props) *Instance {
	shape := getShapeProp(props, DialShapeCircle)
	radiusX := normalizeRadius(getRadiusXProp(props, 5))
	inst := &Instance{
		key:             proputil.GetString(props, propKey, ""),
		shape:           shape,
		radius:          radiusX,
		radiusY:         normalizeVerticalRadius(shape, radiusX, getRadiusYProp(props, radiusX)),
		cellAspectX:     getCellAspectXProp(props, DefaultCellAspectX),
		live:            proputil.GetBool(props, propLive, true),
		timeValue:       getTimeProp(props, time.Time{}),
		location:        getLocationProp(props, nil),
		showSecondHand:  proputil.GetBool(props, propShowSecondHand, true),
		smoothSecond:    proputil.GetBool(props, propSmoothSecond, true),
		showDigital:     proputil.GetBool(props, propShowDigital, true),
		preset:          getPresetProp(props, PresetNone),
		handStyle:       getHandStyleProp(props, HandRenderStyleASCII),
		dialStyle:       proputil.GetStyle(props, propDialStyle, style.Style{}),
		tickStyle:       proputil.GetStyle(props, propTickStyle, style.Style{}),
		centerStyle:     proputil.GetStyle(props, propCenterStyle, style.Style{}),
		digitalStyle:    proputil.GetStyle(props, propDigitalStyle, style.Style{}),
		hourHandStyle:   proputil.GetStyle(props, propHourHandStyle, style.Style{}),
		minuteHandStyle: proputil.GetStyle(props, propMinuteHandStyle, style.Style{}),
		secondHandStyle: proputil.GetStyle(props, propSecondHandStyle, style.Style{}),
		clockStyle:      proputil.GetStyle(props, propStyle, style.Style{}),
		dirty:           true,
	}
	inst.displayTime = inst.resolveDisplayTime(time.Now())
	inst.resetLoop()
	return inst
}

func (inst *Instance) Key() string                        { return inst.key }
func (inst *Instance) SetKey(key string)                  { inst.key = key }
func (inst *Instance) Destroy()                           {}
func (inst *Instance) OnMount()                           {}
func (inst *Instance) OnUnmount()                         {}
func (inst *Instance) GetContext() *rtui.ComponentContext { return nil }
func (inst *Instance) MarkDirty()                         { inst.dirty = true }
func (inst *Instance) IsDirty() bool                      { return inst.dirty }
func (inst *Instance) Init(props rtui.Props)              { inst.SetProps(props) }

func (inst *Instance) SetProps(props rtui.Props) bool {
	oldKey := inst.key
	oldShape := inst.shape
	oldRadius := inst.radius
	oldRadiusY := inst.radiusY
	oldCellAspectX := inst.cellAspectX
	oldLive := inst.live
	oldTimeValue := inst.timeValue
	oldLocation := inst.location
	oldShowSecondHand := inst.showSecondHand
	oldSmoothSecond := inst.smoothSecond
	oldShowDigital := inst.showDigital
	oldPreset := inst.preset
	oldHandStyle := inst.handStyle
	oldDialStyle := inst.dialStyle
	oldTickStyle := inst.tickStyle
	oldCenterStyle := inst.centerStyle
	oldDigitalStyle := inst.digitalStyle
	oldHourHandStyle := inst.hourHandStyle
	oldMinuteHandStyle := inst.minuteHandStyle
	oldSecondHandStyle := inst.secondHandStyle
	oldStyle := inst.clockStyle

	nextShape := getShapeProp(props, inst.shape)
	nextRadius := normalizeRadius(getRadiusXProp(props, inst.radius))

	inst.key = proputil.GetString(props, propKey, inst.key)
	inst.shape = nextShape
	inst.radius = nextRadius
	inst.radiusY = normalizeVerticalRadius(inst.shape, inst.radius, getRadiusYProp(props, inst.radiusY))
	inst.cellAspectX = getCellAspectXProp(props, inst.cellAspectX)
	inst.live = proputil.GetBool(props, propLive, inst.live)
	inst.timeValue = getTimeProp(props, inst.timeValue)
	inst.location = getLocationProp(props, inst.location)
	inst.showSecondHand = proputil.GetBool(props, propShowSecondHand, inst.showSecondHand)
	inst.smoothSecond = proputil.GetBool(props, propSmoothSecond, inst.smoothSecond)
	inst.showDigital = proputil.GetBool(props, propShowDigital, inst.showDigital)
	inst.preset = getPresetProp(props, inst.preset)
	inst.handStyle = getHandStyleProp(props, inst.handStyle)
	inst.dialStyle = proputil.GetStyle(props, propDialStyle, inst.dialStyle)
	inst.tickStyle = proputil.GetStyle(props, propTickStyle, inst.tickStyle)
	inst.centerStyle = proputil.GetStyle(props, propCenterStyle, inst.centerStyle)
	inst.digitalStyle = proputil.GetStyle(props, propDigitalStyle, inst.digitalStyle)
	inst.hourHandStyle = proputil.GetStyle(props, propHourHandStyle, inst.hourHandStyle)
	inst.minuteHandStyle = proputil.GetStyle(props, propMinuteHandStyle, inst.minuteHandStyle)
	inst.secondHandStyle = proputil.GetStyle(props, propSecondHandStyle, inst.secondHandStyle)
	inst.clockStyle = proputil.GetStyle(props, propStyle, inst.clockStyle)

	clockChanged := oldShape != inst.shape ||
		oldRadius != inst.radius ||
		oldRadiusY != inst.radiusY ||
		oldCellAspectX != inst.cellAspectX ||
		oldLive != inst.live ||
		!timesEqual(oldTimeValue, inst.timeValue) ||
		oldLocation != inst.location ||
		oldShowSecondHand != inst.showSecondHand ||
		oldSmoothSecond != inst.smoothSecond ||
		oldShowDigital != inst.showDigital ||
		oldPreset != inst.preset ||
		oldHandStyle != inst.handStyle ||
		oldDialStyle != inst.dialStyle ||
		oldTickStyle != inst.tickStyle ||
		oldCenterStyle != inst.centerStyle ||
		oldDigitalStyle != inst.digitalStyle ||
		oldHourHandStyle != inst.hourHandStyle ||
		oldMinuteHandStyle != inst.minuteHandStyle ||
		oldSecondHandStyle != inst.secondHandStyle ||
		oldStyle != inst.clockStyle ||
		oldKey != inst.key

	if !clockChanged {
		return false
	}

	if oldLive != inst.live || oldShowSecondHand != inst.showSecondHand || oldSmoothSecond != inst.smoothSecond {
		inst.resetLoop()
	}
	inst.displayTime = inst.resolveDisplayTime(time.Now())
	inst.dirty = true
	return true
}

func (inst *Instance) GetProps() rtui.Props {
	return rtui.Props{
		propKey:             inst.key,
		propShape:           inst.shape,
		propRadius:          inst.radius,
		propRadiusX:         inst.radius,
		propRadiusY:         inst.radiusY,
		propCellAspectX:     inst.cellAspectX,
		propLive:            inst.live,
		propTime:            inst.timeValue,
		propLocation:        inst.location,
		propShowSecondHand:  inst.showSecondHand,
		propSmoothSecond:    inst.smoothSecond,
		propShowDigital:     inst.showDigital,
		propPreset:          inst.preset,
		propHandStyle:       inst.handStyle,
		propDialStyle:       inst.dialStyle,
		propTickStyle:       inst.tickStyle,
		propCenterStyle:     inst.centerStyle,
		propDigitalStyle:    inst.digitalStyle,
		propHourHandStyle:   inst.hourHandStyle,
		propMinuteHandStyle: inst.minuteHandStyle,
		propSecondHandStyle: inst.secondHandStyle,
		propStyle:           inst.clockStyle,
	}
}

func (inst *Instance) WantsTick() bool {
	return inst.live && inst.loop != nil && inst.loop.WantsTick()
}

func (inst *Instance) Tick(now time.Time) bool {
	if !inst.WantsTick() {
		return false
	}
	if !inst.loop.Primed() {
		inst.loop.Prime(now)
		return false
	}
	if !inst.loop.Tick(now) {
		return false
	}

	next := inst.resolveDisplayTime(now)
	if !inst.visibleTimeChanged(inst.displayTime, next) {
		return false
	}

	inst.displayTime = next
	inst.dirty = true
	return true
}

func (inst *Instance) Paint(x, y int) []paint.DrawCmd {
	visualStyle := inst.resolveStyle()
	presetTheme := ThemeForPreset(inst.preset)
	rows := inst.faceRows()
	cmds := make([]paint.DrawCmd, 0, len(rows)+1+(maxInt(inst.widthCells(), inst.heightCells())*12))

	for rowIndex, row := range rows {
		cmds = append(cmds, paint.DrawCmd{
			X:     x,
			Y:     y + rowIndex,
			Text:  row,
			Style: visualStyle,
		})
	}

	cmds = append(cmds, inst.facePaintCommands(x, y, visualStyle, presetTheme)...)
	cmds = append(cmds, inst.handPaintCommands(x, y, visualStyle, presetTheme)...)
	cmds = append(cmds, paint.DrawCmd{
		X:     x + inst.centerX(),
		Y:     y + inst.centerY(),
		Text:  "@",
		Style: resolvePartStyle(visualStyle, presetTheme.CenterStyle, inst.centerStyle),
	})

	if label := inst.digitalText(); label != "" {
		cmds = append(cmds, paint.DrawCmd{
			X:     x,
			Y:     y + len(rows),
			Text:  label,
			Style: resolvePartStyle(visualStyle, presetTheme.DigitalStyle, inst.digitalStyle),
		})
	}

	return cmds
}

func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
	width := inst.widthCells()
	if label := inst.digitalText(); label != "" && paint.StringWidth(label) > width {
		width = paint.StringWidth(label)
	}

	height := inst.heightCells()
	if inst.digitalText() != "" {
		height++
	}

	return layout.Size{
		Width:  constraints.ConstrainWidth(width),
		Height: constraints.ConstrainHeight(height),
	}
}

func (inst *Instance) resolveStyle() style.Style {
	presetTheme := ThemeForPreset(inst.preset)
	s := presetTheme.BaseStyle
	if s.FG == "" {
		s = s.Foreground(theme.Primary())
	}
	if s.BG == "" {
		s = s.Background(theme.Surface())
	}
	s = s.Merge(inst.clockStyle)
	return s
}

func (inst *Instance) widthCells() int {
	return inst.renderRadiusX()*2 + 1
}

func (inst *Instance) heightCells() int {
	return inst.radiusY*2 + 1
}

func (inst *Instance) centerX() int {
	return inst.renderRadiusX()
}

func (inst *Instance) centerY() int {
	return inst.radiusY
}

func (inst *Instance) renderRadiusX() int {
	return maxInt(1, int(math.Round(float64(inst.radius)*inst.cellAspectX)))
}

func (inst *Instance) digitalText() string {
	if !inst.showDigital {
		return ""
	}
	if inst.showSecondHand {
		return inst.displayTime.Format("15:04:05")
	}
	return inst.displayTime.Format("15:04")
}

func (inst *Instance) faceRows() []string {
	grid := inst.newGrid()
	inst.drawDial(grid, inst.centerX(), inst.centerY())
	inst.drawTickMarks(grid, inst.centerX(), inst.centerY())
	grid[inst.centerY()][inst.centerX()] = '@'

	rows := make([]string, len(grid))
	for i, row := range grid {
		rows[i] = string(row)
	}
	return rows
}

func (inst *Instance) facePaintCommands(x, y int, baseStyle style.Style, presetTheme Theme) []paint.DrawCmd {
	cmds := make([]paint.DrawCmd, 0, inst.widthCells()+inst.heightCells()+24)
	dialStyle := resolvePartStyle(baseStyle, presetTheme.DialStyle, inst.dialStyle)
	tickStyle := resolvePartStyle(baseStyle, presetTheme.TickStyle, inst.tickStyle)

	for row := 0; row < inst.heightCells(); row++ {
		for col := 0; col < inst.widthCells(); col++ {
			if inst.dialDrawsAtCell(col, row) {
				cmds = append(cmds, paint.DrawCmd{
					X:     x + col,
					Y:     y + row,
					Text:  "o",
					Style: dialStyle,
				})
			}
		}
	}

	for i := 0; i < 12; i++ {
		cellX, cellY := inst.tickCellFor(i)
		if cellY >= 0 && cellY < inst.heightCells() && cellX >= 0 && cellX < inst.widthCells() {
			cmds = append(cmds, paint.DrawCmd{
				X:     x + cellX,
				Y:     y + cellY,
				Text:  "#",
				Style: tickStyle,
			})
		}
	}

	return cmds
}

func (inst *Instance) clockRows() []string {
	grid := inst.newGrid()
	inst.drawDial(grid, inst.centerX(), inst.centerY())
	inst.drawTickMarks(grid, inst.centerX(), inst.centerY())
	inst.drawHands(grid, inst.centerX(), inst.centerY())
	grid[inst.centerY()][inst.centerX()] = '@'

	rows := make([]string, len(grid))
	for i, row := range grid {
		rows[i] = string(row)
	}
	return rows
}

func (inst *Instance) newGrid() [][]rune {
	grid := make([][]rune, inst.heightCells())
	for row := range grid {
		grid[row] = []rune(strings.Repeat(" ", inst.widthCells()))
	}
	return grid
}

func (inst *Instance) drawDial(grid [][]rune, centerX, centerY int) {
	for y := range grid {
		for x := range grid[y] {
			if inst.dialDrawsAtCell(x, y) {
				grid[y][x] = 'o'
			}
		}
	}
}

func (inst *Instance) drawTickMarks(grid [][]rune, centerX, centerY int) {
	for i := 0; i < 12; i++ {
		x, y := inst.tickCellFor(i)
		if y >= 0 && y < len(grid) && x >= 0 && x < len(grid[y]) {
			grid[y][x] = '#'
		}
	}
}

func (inst *Instance) drawHands(grid [][]rune, centerX, centerY int) {
	glyphs := handGlyphsForStyle(inst.handStyle)
	secondValue, minuteValue, hourValue := inst.handClockValues()
	renderRadiusX := inst.renderRadiusX()

	inst.drawHandEllipse(grid, centerX, centerY, clockAngle(hourValue/12), maxInt(1, int(math.Round(float64(renderRadiusX)*0.5))), maxInt(1, int(math.Round(float64(inst.radiusY)*0.5))), glyphs.hourTip, glyphs)
	inst.drawHandEllipse(grid, centerX, centerY, clockAngle(minuteValue/60), maxInt(1, int(math.Round(float64(renderRadiusX)*0.75))), maxInt(1, int(math.Round(float64(inst.radiusY)*0.75))), glyphs.minuteTip, glyphs)
	if inst.showSecondHand {
		inst.drawHandEllipse(grid, centerX, centerY, clockAngle(secondValue/60), maxInt(1, renderRadiusX-1), maxInt(1, inst.radiusY-1), glyphs.secondTip, glyphs)
	}
}

func (inst *Instance) handPaintCommands(x, y int, baseStyle style.Style, presetTheme Theme) []paint.DrawCmd {
	glyphs := handGlyphsForStyle(inst.handStyle)
	secondValue, minuteValue, hourValue := inst.handClockValues()

	cmds := make([]paint.DrawCmd, 0, maxInt(inst.radius, inst.radiusY)*12)
	centerX := inst.centerX()
	centerY := inst.centerY()
	renderRadiusX := inst.renderRadiusX()

	cmds = append(cmds, inst.handCommandsForEllipse(x, y, centerX, centerY, clockAngle(hourValue/12), maxInt(1, int(math.Round(float64(renderRadiusX)*0.5))), maxInt(1, int(math.Round(float64(inst.radiusY)*0.5))), glyphs.hourTip, glyphs, resolvePartStyle(baseStyle, presetTheme.HourHandStyle, inst.hourHandStyle))...)
	cmds = append(cmds, inst.handCommandsForEllipse(x, y, centerX, centerY, clockAngle(minuteValue/60), maxInt(1, int(math.Round(float64(renderRadiusX)*0.75))), maxInt(1, int(math.Round(float64(inst.radiusY)*0.75))), glyphs.minuteTip, glyphs, resolvePartStyle(baseStyle, presetTheme.MinuteHandStyle, inst.minuteHandStyle))...)
	if inst.showSecondHand {
		cmds = append(cmds, inst.handCommandsForEllipse(x, y, centerX, centerY, clockAngle(secondValue/60), maxInt(1, renderRadiusX-1), maxInt(1, inst.radiusY-1), glyphs.secondTip, glyphs, resolvePartStyle(baseStyle, presetTheme.SecondHandStyle, inst.secondHandStyle))...)
	}
	return cmds
}

func (inst *Instance) handCommandsFor(x, y, center int, angle float64, length int, tipGlyph rune, glyphs handGlyphSet, drawStyle style.Style) []paint.DrawCmd {
	shaftGlyph := handShaftGlyph(angle, glyphs)
	steps := maxInt(1, length*4)
	cmds := make([]paint.DrawCmd, 0, steps)
	for i := 1; i <= steps; i++ {
		dist := float64(length) * float64(i) / float64(steps)
		cellX := center + int(math.Round(math.Cos(angle)*dist))
		cellY := center + int(math.Round(math.Sin(angle)*dist))
		glyph := shaftGlyph
		if i == steps {
			glyph = tipGlyph
		}
		cmds = append(cmds, paint.DrawCmd{
			X:     x + cellX,
			Y:     y + cellY,
			Text:  string(glyph),
			Style: drawStyle,
		})
	}
	return cmds
}

func (inst *Instance) handCommandsForEllipse(x, y, centerX, centerY int, angle float64, lengthX, lengthY int, tipGlyph rune, glyphs handGlyphSet, drawStyle style.Style) []paint.DrawCmd {
	shaftGlyph := handShaftGlyph(angle, glyphs)
	steps := maxInt(1, maxInt(lengthX, lengthY)*4)
	cmds := make([]paint.DrawCmd, 0, steps)
	for i := 1; i <= steps; i++ {
		distX := float64(lengthX) * float64(i) / float64(steps)
		distY := float64(lengthY) * float64(i) / float64(steps)
		cellX := centerX + int(math.Round(math.Cos(angle)*distX))
		cellY := centerY + int(math.Round(math.Sin(angle)*distY))
		glyph := shaftGlyph
		if i == steps {
			glyph = tipGlyph
		}
		cmds = append(cmds, paint.DrawCmd{
			X:     x + cellX,
			Y:     y + cellY,
			Text:  string(glyph),
			Style: drawStyle,
		})
	}
	return cmds
}

func (inst *Instance) drawHand(grid [][]rune, center int, angle float64, length int, tipGlyph rune, glyphs handGlyphSet) {
	shaftGlyph := handShaftGlyph(angle, glyphs)
	steps := maxInt(1, length*4)
	for i := 1; i <= steps; i++ {
		dist := float64(length) * float64(i) / float64(steps)
		x := center + int(math.Round(math.Cos(angle)*dist))
		y := center + int(math.Round(math.Sin(angle)*dist))
		if y >= 0 && y < len(grid) && x >= 0 && x < len(grid[y]) {
			glyph := shaftGlyph
			if i == steps {
				glyph = tipGlyph
			}
			grid[y][x] = glyph
		}
	}
}

func (inst *Instance) drawHandEllipse(grid [][]rune, centerX, centerY int, angle float64, lengthX, lengthY int, tipGlyph rune, glyphs handGlyphSet) {
	shaftGlyph := handShaftGlyph(angle, glyphs)
	steps := maxInt(1, maxInt(lengthX, lengthY)*4)
	for i := 1; i <= steps; i++ {
		distX := float64(lengthX) * float64(i) / float64(steps)
		distY := float64(lengthY) * float64(i) / float64(steps)
		x := centerX + int(math.Round(math.Cos(angle)*distX))
		y := centerY + int(math.Round(math.Sin(angle)*distY))
		if y >= 0 && y < len(grid) && x >= 0 && x < len(grid[y]) {
			glyph := shaftGlyph
			if i == steps {
				glyph = tipGlyph
			}
			grid[y][x] = glyph
		}
	}
}

func (inst *Instance) handClockValues() (secondValue, minuteValue, hourValue float64) {
	secondValue = float64(inst.displayTime.Second())
	if inst.smoothSecond && inst.showSecondHand {
		secondValue += float64(inst.displayTime.Nanosecond()) / float64(time.Second)
	}
	minuteValue = float64(inst.displayTime.Minute()) + secondValue/60
	hourValue = float64(inst.displayTime.Hour()%12) + minuteValue/60
	return secondValue, minuteValue, hourValue
}

func (inst *Instance) dialDrawsAtCell(col, row int) bool {
	dx := float64(col - inst.centerX())
	dy := float64(row - inst.centerY())
	radiusX := float64(inst.renderRadiusX())
	radiusY := float64(inst.radiusY)
	normalized := math.Sqrt((dx*dx)/(radiusX*radiusX) + (dy*dy)/(radiusY*radiusY))
	return math.Abs(normalized-1) <= ellipseEdgeTolerance(radiusX, radiusY)
}

func (inst *Instance) tickCellFor(index int) (int, int) {
	angle := clockAngle(float64(index) / 12)
	return inst.centerX() + int(math.Round(math.Cos(angle)*float64(maxInt(1, inst.renderRadiusX()-1)))),
		inst.centerY() + int(math.Round(math.Sin(angle)*float64(maxInt(1, inst.radiusY-1))))
}

func ellipseEdgeTolerance(radiusX, radiusY float64) float64 {
	return 0.55 / math.Max(1, math.Min(radiusX, radiusY))
}

func (inst *Instance) resetLoop() {
	if !inst.live {
		inst.loop = nil
		return
	}

	interval := discreteTickInterval
	if inst.showSecondHand && inst.smoothSecond {
		interval = smoothTickInterval
	}
	inst.loop = animation.NewLoopDriver(animation.LoopDriverConfig{
		Duration:  interval,
		Cycles:    0,
		AutoStart: true,
	})
}

func (inst *Instance) resolveDisplayTime(baseNow time.Time) time.Time {
	result := baseNow
	if !inst.live && !inst.timeValue.IsZero() {
		result = inst.timeValue
	}
	if inst.location != nil {
		result = result.In(inst.location)
	}
	return result
}

func (inst *Instance) visibleTimeChanged(prev, next time.Time) bool {
	if inst.showSecondHand && inst.smoothSecond {
		return !next.Equal(prev)
	}
	if inst.showSecondHand {
		return next.Format("15:04:05") != prev.Format("15:04:05")
	}
	return next.Format("15:04") != prev.Format("15:04")
}

func getTimeProp(props rtui.Props, def time.Time) time.Time {
	if value, ok := props[propTime]; ok {
		if timeValue, ok := value.(time.Time); ok {
			return timeValue
		}
	}
	return def
}

func getLocationProp(props rtui.Props, def *time.Location) *time.Location {
	if value, ok := props[propLocation]; ok {
		if location, ok := value.(*time.Location); ok {
			return location
		}
	}
	return def
}

func getShapeProp(props rtui.Props, def DialShape) DialShape {
	if value, ok := props[propShape]; ok {
		if shape, ok := value.(DialShape); ok {
			return normalizeShape(shape)
		}
	}
	return normalizeShape(def)
}

func getRadiusXProp(props rtui.Props, def int) int {
	if value, ok := props[propRadiusX]; ok {
		if radius, ok := value.(int); ok {
			return radius
		}
	}
	if value, ok := props[propRadius]; ok {
		if radius, ok := value.(int); ok {
			return radius
		}
	}
	return def
}

func getRadiusYProp(props rtui.Props, def int) int {
	if value, ok := props[propRadiusY]; ok {
		if radius, ok := value.(int); ok {
			return radius
		}
	}
	if value, ok := props[propRadius]; ok {
		if radius, ok := value.(int); ok {
			return radius
		}
	}
	return def
}

func getCellAspectXProp(props rtui.Props, def float64) float64 {
	if value, ok := props[propCellAspectX]; ok {
		switch aspect := value.(type) {
		case float64:
			return normalizeCellAspectX(aspect)
		case float32:
			return normalizeCellAspectX(float64(aspect))
		case int:
			return normalizeCellAspectX(float64(aspect))
		}
	}
	return normalizeCellAspectX(def)
}

func getPresetProp(props rtui.Props, def Preset) Preset {
	if value, ok := props[propPreset]; ok {
		if preset, ok := value.(Preset); ok {
			return preset
		}
	}
	return def
}

func getHandStyleProp(props rtui.Props, def HandRenderStyle) HandRenderStyle {
	if value, ok := props[propHandStyle]; ok {
		if handStyle, ok := value.(HandRenderStyle); ok {
			return handStyle
		}
	}
	return def
}

func normalizeShape(shape DialShape) DialShape {
	switch shape {
	case DialShapeEllipse:
		return DialShapeEllipse
	default:
		return DialShapeCircle
	}
}

func normalizeRadius(radius int) int {
	if radius < minRadius {
		return minRadius
	}
	return radius
}

func normalizeVerticalRadius(shape DialShape, radiusX, radiusY int) int {
	if normalizeShape(shape) == DialShapeCircle {
		return normalizeRadius(radiusX)
	}
	return normalizeRadius(radiusY)
}

func normalizeCellAspectX(cellAspectX float64) float64 {
	if math.IsNaN(cellAspectX) || math.IsInf(cellAspectX, 0) || cellAspectX <= 0 {
		return DefaultCellAspectX
	}
	return cellAspectX
}

func timesEqual(a, b time.Time) bool {
	if a.IsZero() || b.IsZero() {
		return a.IsZero() && b.IsZero()
	}
	return a.Equal(b)
}

func clockAngle(fraction float64) float64 {
	return -math.Pi/2 + 2*math.Pi*fraction
}

type handGlyphSet struct {
	horizontal rune
	vertical   rune
	diagDown   rune
	diagUp     rune
	hourTip    rune
	minuteTip  rune
	secondTip  rune
}

func handGlyphsForStyle(handStyle HandRenderStyle) handGlyphSet {
	switch handStyle {
	case HandRenderStyleUnicode:
		return handGlyphSet{
			horizontal: '─',
			vertical:   '│',
			diagDown:   '╲',
			diagUp:     '╱',
			hourTip:    '■',
			minuteTip:  '●',
			secondTip:  '•',
		}
	default:
		return handGlyphSet{
			horizontal: '-',
			vertical:   '|',
			diagDown:   '\\',
			diagUp:     '/',
			hourTip:    'O',
			minuteTip:  '+',
			secondTip:  '.',
		}
	}
}

func handShaftGlyph(angle float64, glyphs handGlyphSet) rune {
	dx := math.Cos(angle)
	dy := math.Sin(angle)
	absDX := math.Abs(dx)
	absDY := math.Abs(dy)

	switch {
	case absDX >= absDY*2:
		return glyphs.horizontal
	case absDY >= absDX*2:
		return glyphs.vertical
	case (dx >= 0 && dy >= 0) || (dx < 0 && dy < 0):
		return glyphs.diagDown
	default:
		return glyphs.diagUp
	}
}

func resolvePartStyle(baseStyle, presetStyle, override style.Style) style.Style {
	s := baseStyle
	if !presetStyle.IsEmpty() {
		s = s.Merge(presetStyle)
	}
	if override.IsEmpty() {
		return s
	}
	return s.Merge(override)
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// DebugRows returns the rendered rows for tests and examples.
func (inst *Instance) DebugRows() []string {
	rows := append([]string{}, inst.clockRows()...)
	if label := inst.digitalText(); label != "" {
		rows = append(rows, fmt.Sprintf("%s", label))
	}
	return rows
}
