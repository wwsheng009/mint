package timer

import (
	"fmt"
	"strings"
	"time"

	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/internal/proputil"
)

const timerTickInterval = time.Second

// Instance is the runtime entity for Timer components.
type Instance struct {
	key                string
	label              string
	mode               Mode
	duration           time.Duration
	startedAt          time.Time
	deadline           time.Time
	now                time.Time
	live               bool
	width              int
	showProgress       bool
	progressWidth      int
	progressGlyphStyle ProgressGlyphStyle
	expiredText        string
	warningBelow       time.Duration
	timerStyle         style.Style
	warningStyle       style.Style
	expiredStyle       style.Style
	progressStyle      style.Style
	dirty              bool
}

var (
	_ rtui.ComponentInstance = (*Instance)(nil)
	_ rtui.PaintableInstance = (*Instance)(nil)
	_ rtui.TickableInstance  = (*Instance)(nil)
	_ interface {
		Measure(layout.Constraints) layout.Size
	} = (*Instance)(nil)
)

// NewInstance creates a Timer instance from props.
func NewInstance(props rtui.Props) *Instance {
	now := getTimeProp(props, propNow, time.Now())
	startedAt := getTimeProp(props, propStartedAt, now)
	mode := getModeProp(props, ModeElapsed)
	duration := getDurationProp(props, propDuration, 0)
	deadline := getTimeProp(props, propDeadline, time.Time{})
	if mode == ModeCountdown && deadline.IsZero() && duration > 0 {
		deadline = startedAt.Add(duration)
	}

	return &Instance{
		key:                proputil.GetString(props, propKey, ""),
		label:              proputil.GetString(props, propLabel, ""),
		mode:               mode,
		duration:           duration,
		startedAt:          startedAt,
		deadline:           deadline,
		now:                now,
		live:               proputil.GetBool(props, propLive, true),
		width:              maxInt(0, proputil.GetInt(props, propWidth, 0)),
		showProgress:       proputil.GetBool(props, propShowProgress, false),
		progressWidth:      maxInt(0, proputil.GetInt(props, propProgressWidth, 12)),
		progressGlyphStyle: getProgressGlyphStyleProp(props, ProgressGlyphStyleUnicode),
		expiredText:        proputil.GetString(props, propExpiredText, "00:00"),
		warningBelow:       getDurationProp(props, propWarningBelow, 10*time.Second),
		timerStyle:         proputil.GetStyle(props, propStyle, style.Style{}),
		warningStyle:       proputil.GetStyle(props, propWarningStyle, style.Style{}),
		expiredStyle:       proputil.GetStyle(props, propExpiredStyle, style.Style{}),
		progressStyle:      proputil.GetStyle(props, propProgressStyle, style.Style{}),
		dirty:              true,
	}
}

func (inst *Instance) Key() string       { return inst.key }
func (inst *Instance) SetKey(key string) { inst.key = key }
func (inst *Instance) Init(props rtui.Props) {
	inst.SetProps(props)
}
func (inst *Instance) Destroy()                           {}
func (inst *Instance) OnMount()                           {}
func (inst *Instance) OnUnmount()                         {}
func (inst *Instance) MarkDirty()                         { inst.dirty = true }
func (inst *Instance) IsDirty() bool                      { return inst.dirty }
func (inst *Instance) GetContext() *rtui.ComponentContext { return nil }

func (inst *Instance) SetProps(props rtui.Props) bool {
	old := inst.GetProps()

	now := getTimeProp(props, propNow, inst.now)
	mode := getModeProp(props, inst.mode)
	startedAt := getTimeProp(props, propStartedAt, inst.startedAt)
	if startedAt.IsZero() {
		startedAt = now
	}
	duration := getDurationProp(props, propDuration, inst.duration)
	deadline := getTimeProp(props, propDeadline, time.Time{})
	if deadline.IsZero() {
		if existing := inst.deadline; !existing.IsZero() && mode == inst.mode && duration == inst.duration {
			deadline = existing
		} else if mode == ModeCountdown && duration > 0 {
			deadline = startedAt.Add(duration)
		}
	}

	inst.key = proputil.GetString(props, propKey, inst.key)
	inst.label = proputil.GetString(props, propLabel, inst.label)
	inst.mode = mode
	inst.duration = duration
	inst.startedAt = startedAt
	inst.deadline = deadline
	inst.now = now
	inst.live = proputil.GetBool(props, propLive, inst.live)
	inst.width = maxInt(0, proputil.GetInt(props, propWidth, inst.width))
	inst.showProgress = proputil.GetBool(props, propShowProgress, inst.showProgress)
	inst.progressWidth = maxInt(0, proputil.GetInt(props, propProgressWidth, inst.progressWidth))
	inst.progressGlyphStyle = getProgressGlyphStyleProp(props, inst.progressGlyphStyle)
	inst.expiredText = proputil.GetString(props, propExpiredText, inst.expiredText)
	inst.warningBelow = getDurationProp(props, propWarningBelow, inst.warningBelow)
	inst.timerStyle = proputil.GetStyle(props, propStyle, inst.timerStyle)
	inst.warningStyle = proputil.GetStyle(props, propWarningStyle, inst.warningStyle)
	inst.expiredStyle = proputil.GetStyle(props, propExpiredStyle, inst.expiredStyle)
	inst.progressStyle = proputil.GetStyle(props, propProgressStyle, inst.progressStyle)

	changed := propsChanged(old, inst.GetProps())
	if changed {
		inst.dirty = true
	}
	return changed
}

func (inst *Instance) GetProps() rtui.Props {
	return rtui.Props{
		propDuration:           inst.duration,
		propExpiredStyle:       inst.expiredStyle,
		propExpiredText:        inst.expiredText,
		propKey:                inst.key,
		propLabel:              inst.label,
		propLive:               inst.live,
		propMode:               inst.mode,
		propNow:                inst.now,
		propProgressStyle:      inst.progressStyle,
		propProgressGlyphStyle: inst.progressGlyphStyle,
		propProgressWidth:      inst.progressWidth,
		propShowProgress:       inst.showProgress,
		propStartedAt:          inst.startedAt,
		propStyle:              inst.timerStyle,
		propDeadline:           inst.deadline,
		propWarningBelow:       inst.warningBelow,
		propWarningStyle:       inst.warningStyle,
		propWidth:              inst.width,
	}
}

func (inst *Instance) WantsTick() bool {
	if !inst.live {
		return false
	}
	if inst.mode == ModeCountdown && inst.isExpired() {
		return false
	}
	return true
}

func (inst *Instance) Tick(now time.Time) bool {
	if !inst.WantsTick() {
		return false
	}
	before := inst.lineText()
	if now.Sub(inst.now) < timerTickInterval && inst.now.Sub(now) < timerTickInterval {
		return false
	}
	inst.now = now
	changed := before != inst.lineText()
	if changed {
		inst.dirty = true
	}
	return changed
}

func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
	width := paint.StringWidth(inst.fittedLineText())
	if inst.width > 0 {
		width = inst.width
	}
	return layout.Size{
		Width:  constraints.ConstrainWidth(width),
		Height: constraints.ConstrainHeight(1),
	}
}

func (inst *Instance) Paint(x, y int) []paint.DrawCmd {
	text := inst.fittedLineText()
	return []paint.DrawCmd{{
		X:     x,
		Y:     y,
		Text:  text,
		Style: inst.resolveStyle(),
	}}
}

func (inst *Instance) lineText() string {
	value := inst.valueText()
	if inst.label != "" {
		value = inst.label + ": " + value
	}
	if inst.showProgress {
		value += " " + inst.progressBar()
	}
	return value
}

func (inst *Instance) fittedLineText() string {
	text := inst.lineText()
	if inst.width <= 0 {
		return text
	}
	return fitText(text, inst.width)
}

func (inst *Instance) valueText() string {
	if inst.mode == ModeCountdown {
		remaining := inst.remaining()
		if remaining <= 0 {
			return inst.expiredText
		}
		return formatDuration(remaining, true)
	}
	return formatDuration(inst.elapsed(), false)
}

func (inst *Instance) elapsed() time.Duration {
	if inst.startedAt.IsZero() || inst.now.Before(inst.startedAt) {
		return 0
	}
	return inst.now.Sub(inst.startedAt)
}

func (inst *Instance) remaining() time.Duration {
	if inst.deadline.IsZero() {
		return 0
	}
	if !inst.now.Before(inst.deadline) {
		return 0
	}
	return inst.deadline.Sub(inst.now)
}

func (inst *Instance) isExpired() bool {
	return inst.mode == ModeCountdown && inst.remaining() <= 0
}

func (inst *Instance) isWarning() bool {
	return inst.mode == ModeCountdown &&
		inst.warningBelow > 0 &&
		inst.remaining() > 0 &&
		inst.remaining() <= inst.warningBelow
}

func (inst *Instance) progressBar() string {
	width := inst.progressWidth
	if width < 3 {
		width = 3
	}
	inner := width - 2
	percent := inst.progressPercent()
	filled := (percent * inner) / 100
	if filled < 0 {
		filled = 0
	}
	if filled > inner {
		filled = inner
	}
	filledRune, emptyRune := inst.progressRunes()
	return "[" + strings.Repeat(string(filledRune), filled) + strings.Repeat(string(emptyRune), inner-filled) + "]"
}

func (inst *Instance) progressPercent() int {
	total := inst.totalDuration()
	if total <= 0 {
		return 0
	}
	elapsed := inst.elapsed()
	if inst.mode == ModeCountdown {
		elapsed = total - inst.remaining()
	}
	if elapsed < 0 {
		elapsed = 0
	}
	if elapsed > total {
		elapsed = total
	}
	return int((elapsed * 100) / total)
}

func (inst *Instance) totalDuration() time.Duration {
	if inst.duration > 0 {
		return inst.duration
	}
	if !inst.deadline.IsZero() && !inst.startedAt.IsZero() && inst.deadline.After(inst.startedAt) {
		return inst.deadline.Sub(inst.startedAt)
	}
	return 0
}

func (inst *Instance) resolveStyle() style.Style {
	if inst.isExpired() {
		return resolveTimerStyle(inst.expiredStyle, theme.Error())
	}
	if inst.isWarning() {
		return resolveTimerStyle(inst.warningStyle, theme.Warning())
	}
	return resolveTimerStyle(inst.timerStyle, theme.Primary())
}

func resolveTimerStyle(base style.Style, fg style.Color) style.Style {
	if base.FG == "" {
		base = base.Foreground(fg)
	}
	if base.BG == "" {
		base = base.Background(theme.Surface())
	}
	return base
}

func (inst *Instance) progressRunes() (rune, rune) {
	if inst.progressGlyphStyle == ProgressGlyphStyleASCII {
		return '#', '-'
	}
	return '█', '░'
}

func getModeProp(props rtui.Props, def Mode) Mode {
	if v, ok := props[propMode]; ok {
		if mode, ok := v.(Mode); ok {
			return mode
		}
	}
	return def
}

func getProgressGlyphStyleProp(props rtui.Props, def ProgressGlyphStyle) ProgressGlyphStyle {
	if v, ok := props[propProgressGlyphStyle]; ok {
		if glyphStyle, ok := v.(ProgressGlyphStyle); ok {
			return glyphStyle
		}
	}
	return def
}

func getTimeProp(props rtui.Props, key string, def time.Time) time.Time {
	if v, ok := props[key]; ok {
		if value, ok := v.(time.Time); ok {
			if !value.IsZero() {
				return value
			}
		}
	}
	return def
}

func getDurationProp(props rtui.Props, key string, def time.Duration) time.Duration {
	if v, ok := props[key]; ok {
		if value, ok := v.(time.Duration); ok {
			return normalizeDuration(value)
		}
	}
	return normalizeDuration(def)
}

func propsChanged(a, b rtui.Props) bool {
	if len(a) != len(b) {
		return true
	}
	for key, left := range a {
		if fmt.Sprint(left) != fmt.Sprint(b[key]) {
			return true
		}
	}
	return false
}

func formatDuration(duration time.Duration, roundUp bool) string {
	if duration < 0 {
		duration = 0
	}
	seconds := int(duration / time.Second)
	if roundUp && duration%time.Second != 0 {
		seconds++
	}
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	secs := seconds % 60
	if hours > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, secs)
	}
	return fmt.Sprintf("%02d:%02d", minutes, secs)
}

func fitText(text string, width int) string {
	if width <= 0 {
		return ""
	}
	if paint.StringWidth(text) > width {
		text = truncateWithDots(text, width)
	}
	if pad := width - paint.StringWidth(text); pad > 0 {
		text += strings.Repeat(" ", pad)
	}
	return text
}

func truncateWithDots(text string, width int) string {
	if width <= 0 {
		return ""
	}
	if paint.StringWidth(text) <= width {
		return text
	}
	if width <= 3 {
		return strings.Repeat(".", width)
	}
	limit := width - 3
	var out strings.Builder
	used := 0
	for _, r := range text {
		next := paint.StringWidth(string(r))
		if used+next > limit {
			break
		}
		out.WriteRune(r)
		used += next
	}
	return out.String() + "..."
}

func normalizeDuration(duration time.Duration) time.Duration {
	if duration < 0 {
		return 0
	}
	return duration
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
