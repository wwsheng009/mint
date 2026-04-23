package clock

import "github.com/wwsheng009/mint/runtime/style"

// Preset defines a built-in visual theme for the clock.
type Preset int

const (
	PresetNone Preset = iota
	PresetClassic
	PresetNeon
	PresetMinimal
	PresetAlert
)

// Theme is a reusable bulk style configuration for the clock face and hands.
//
// It can be applied directly through Builder.Theme(...) / VNode.SetTheme(...),
// or produced from a preset through ThemeForPreset(...).
type Theme struct {
	BaseStyle       style.Style
	DialStyle       style.Style
	TickStyle       style.Style
	CenterStyle     style.Style
	DigitalStyle    style.Style
	HourHandStyle   style.Style
	MinuteHandStyle style.Style
	SecondHandStyle style.Style
}

// Merge overlays another theme onto the receiver.
//
// Each part is merged with style.Style.Merge, so the overlay can change just a
// subset of fields such as foreground color or boldness without replacing the
// rest of the existing theme.
func (t Theme) Merge(overlay Theme) Theme {
	t.BaseStyle = t.BaseStyle.Merge(overlay.BaseStyle)
	t.DialStyle = t.DialStyle.Merge(overlay.DialStyle)
	t.TickStyle = t.TickStyle.Merge(overlay.TickStyle)
	t.CenterStyle = t.CenterStyle.Merge(overlay.CenterStyle)
	t.DigitalStyle = t.DigitalStyle.Merge(overlay.DigitalStyle)
	t.HourHandStyle = t.HourHandStyle.Merge(overlay.HourHandStyle)
	t.MinuteHandStyle = t.MinuteHandStyle.Merge(overlay.MinuteHandStyle)
	t.SecondHandStyle = t.SecondHandStyle.Merge(overlay.SecondHandStyle)
	return t
}

// PresetName returns a stable human-readable name for a preset.
func PresetName(preset Preset) string {
	switch preset {
	case PresetClassic:
		return "Classic"
	case PresetNeon:
		return "Neon"
	case PresetMinimal:
		return "Minimal"
	case PresetAlert:
		return "Alert"
	default:
		return "Default"
	}
}

// ThemeForPreset returns the exported theme object for a built-in preset.
func ThemeForPreset(preset Preset) Theme {
	switch preset {
	case PresetClassic:
		return Theme{
			BaseStyle:       style.Style{}.Foreground(style.BrightWhite),
			DialStyle:       style.Style{}.Foreground(style.BrightBlack),
			TickStyle:       style.Style{}.Foreground(style.BrightWhite).Bold(true),
			CenterStyle:     style.Style{}.Foreground(style.Yellow).Bold(true),
			DigitalStyle:    style.Style{}.Foreground(style.Cyan).Bold(true),
			HourHandStyle:   style.Style{}.Foreground(style.BrightYellow).Bold(true),
			MinuteHandStyle: style.Style{}.Foreground(style.BrightCyan).Bold(true),
			SecondHandStyle: style.Style{}.Foreground(style.BrightRed).Bold(true),
		}
	case PresetNeon:
		return Theme{
			BaseStyle:       style.Style{}.Foreground(style.BrightCyan).Background(style.Black),
			DialStyle:       style.Style{}.Foreground(style.BrightBlue),
			TickStyle:       style.Style{}.Foreground(style.BrightMagenta).Bold(true),
			CenterStyle:     style.Style{}.Foreground(style.BrightWhite).Bold(true),
			DigitalStyle:    style.Style{}.Foreground(style.BrightGreen).Bold(true),
			HourHandStyle:   style.Style{}.Foreground(style.BrightCyan).Bold(true),
			MinuteHandStyle: style.Style{}.Foreground(style.BrightMagenta).Bold(true),
			SecondHandStyle: style.Style{}.Foreground(style.BrightYellow).Bold(true),
		}
	case PresetMinimal:
		return Theme{
			BaseStyle:       style.Style{}.Foreground(style.White),
			DialStyle:       style.Style{}.Foreground(style.BrightBlack),
			TickStyle:       style.Style{}.Foreground(style.BrightBlack),
			CenterStyle:     style.Style{}.Foreground(style.White).Bold(true),
			DigitalStyle:    style.Style{}.Foreground(style.BrightWhite),
			HourHandStyle:   style.Style{}.Foreground(style.White).Bold(true),
			MinuteHandStyle: style.Style{}.Foreground(style.BrightWhite),
			SecondHandStyle: style.Style{}.Foreground(style.BrightBlack),
		}
	case PresetAlert:
		return Theme{
			BaseStyle:       style.Style{}.Foreground(style.BrightYellow).Background(style.Black),
			DialStyle:       style.Style{}.Foreground(style.BrightBlack),
			TickStyle:       style.Style{}.Foreground(style.BrightYellow).Bold(true),
			CenterStyle:     style.Style{}.Foreground(style.BrightRed).Bold(true),
			DigitalStyle:    style.Style{}.Foreground(style.BrightYellow).Bold(true),
			HourHandStyle:   style.Style{}.Foreground(style.BrightYellow).Bold(true),
			MinuteHandStyle: style.Style{}.Foreground(style.BrightWhite).Bold(true),
			SecondHandStyle: style.Style{}.Foreground(style.BrightRed).Bold(true),
		}
	default:
		return Theme{}
	}
}

// ThemePreset is a short alias for ThemeForPreset.
func ThemePreset(preset Preset) Theme {
	return ThemeForPreset(preset)
}

// WithPreset starts from a built-in preset and reapplies the receiver as an overlay.
//
// This keeps any styles already present on the receiver, so both of these are valid:
//
//	Theme{}.WithPreset(PresetClassic).WithDigitalStyle(...)
//	Theme{}.WithDigitalStyle(...).WithPreset(PresetClassic)
func (t Theme) WithPreset(preset Preset) Theme {
	return ThemeForPreset(preset).Merge(t)
}

func (t Theme) WithBaseStyle(s style.Style) Theme {
	t.BaseStyle = s
	return t
}

func (t Theme) WithDialStyle(s style.Style) Theme {
	t.DialStyle = s
	return t
}

func (t Theme) WithTickStyle(s style.Style) Theme {
	t.TickStyle = s
	return t
}

func (t Theme) WithCenterStyle(s style.Style) Theme {
	t.CenterStyle = s
	return t
}

func (t Theme) WithDigitalStyle(s style.Style) Theme {
	t.DigitalStyle = s
	return t
}

func (t Theme) WithHourHandStyle(s style.Style) Theme {
	t.HourHandStyle = s
	return t
}

func (t Theme) WithMinuteHandStyle(s style.Style) Theme {
	t.MinuteHandStyle = s
	return t
}

func (t Theme) WithSecondHandStyle(s style.Style) Theme {
	t.SecondHandStyle = s
	return t
}
