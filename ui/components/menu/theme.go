package menu

import (
	fwtheme "github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime/style"
)

type Theme struct {
	BarStyle            style.Style
	BarActiveStyle      style.Style
	SurfaceStyle        style.Style
	SurfaceBorderStyle  style.Style
	SurfaceShadowStyle  style.Style
	ItemStyle           style.Style
	ItemHoverStyle      style.Style
	ItemFocusStyle      style.Style
	ItemActiveStyle     style.Style
	ItemDisabledStyle   style.Style
	ItemDangerStyle     style.Style
	ItemCheckedStyle    style.Style
	SeparatorStyle      style.Style
	ShortcutStyle       style.Style
	DescriptionStyle    style.Style
	IconStyle           style.Style
	CheckmarkStyle      style.Style
	SubmenuArrowStyle   style.Style
	TitleStyle          style.Style
	ScrollbarStyle      style.Style
	ScrollbarThumbStyle style.Style
}

func DefaultTheme() Theme {
	return Theme{
		BarStyle:            style.NewStyle().Foreground(fwtheme.Text()).Background(fwtheme.Surface()),
		BarActiveStyle:      style.NewStyle().Foreground(fwtheme.BG()).Background(fwtheme.Primary()).Bold(true),
		SurfaceStyle:        style.NewStyle().Foreground(fwtheme.Text()).Background(fwtheme.Surface()),
		SurfaceBorderStyle:  style.NewStyle().Foreground(fwtheme.Border()).Background(fwtheme.Surface()),
		SurfaceShadowStyle:  style.NewStyle().Background(fwtheme.Shadow()),
		ItemStyle:           style.NewStyle().Foreground(fwtheme.Text()).Background(fwtheme.Surface()),
		ItemHoverStyle:      style.NewStyle().Background(fwtheme.Select()),
		ItemFocusStyle:      style.NewStyle().Foreground(fwtheme.FocusBright()).Underline(true),
		ItemActiveStyle:     style.NewStyle().Foreground(fwtheme.BG()).Background(fwtheme.Select()).Bold(true),
		ItemDisabledStyle:   style.NewStyle().Foreground(fwtheme.DisabledFG()).Background(fwtheme.DisabledBG()),
		ItemDangerStyle:     style.NewStyle().Foreground(fwtheme.Error()).Bold(true),
		ItemCheckedStyle:    style.NewStyle().Foreground(fwtheme.Success()).Bold(true),
		SeparatorStyle:      style.NewStyle().Foreground(fwtheme.Border()),
		ShortcutStyle:       style.NewStyle().Foreground(fwtheme.Muted()),
		DescriptionStyle:    style.NewStyle().Foreground(fwtheme.Muted()),
		IconStyle:           style.NewStyle().Foreground(fwtheme.Accent()),
		CheckmarkStyle:      style.NewStyle().Foreground(fwtheme.Success()).Bold(true),
		SubmenuArrowStyle:   style.NewStyle().Foreground(fwtheme.Muted()),
		TitleStyle:          style.NewStyle().Foreground(fwtheme.Text()).Bold(true),
		ScrollbarStyle:      style.NewStyle().Foreground(fwtheme.Scrollbar()),
		ScrollbarThumbStyle: style.NewStyle().Foreground(fwtheme.Primary()),
	}
}

func MutedTheme() Theme {
	t := DefaultTheme()
	t.BarStyle = t.BarStyle.Foreground(fwtheme.Muted())
	t.ItemStyle = t.ItemStyle.Foreground(fwtheme.Muted())
	t.ShortcutStyle = t.ShortcutStyle.Foreground(fwtheme.Placeholder())
	t.TitleStyle = t.TitleStyle.Foreground(fwtheme.Muted())
	return t
}

func ContrastTheme() Theme {
	return Theme{
		BarStyle:            style.NewStyle().Foreground(style.Black).Background(style.BrightWhite).Bold(true),
		BarActiveStyle:      style.NewStyle().Foreground(style.BrightWhite).Background(style.Blue).Bold(true),
		SurfaceStyle:        style.NewStyle().Foreground(style.BrightWhite).Background(style.Black),
		SurfaceBorderStyle:  style.NewStyle().Foreground(style.BrightWhite).Background(style.Black),
		SurfaceShadowStyle:  style.NewStyle().Background(style.BrightBlack),
		ItemStyle:           style.NewStyle().Foreground(style.BrightWhite).Background(style.Black),
		ItemHoverStyle:      style.NewStyle().Background(style.BrightBlack),
		ItemFocusStyle:      style.NewStyle().Underline(true),
		ItemActiveStyle:     style.NewStyle().Foreground(style.Black).Background(style.BrightYellow).Bold(true),
		ItemDisabledStyle:   style.NewStyle().Foreground(style.BrightBlack).Background(style.Black),
		ItemDangerStyle:     style.NewStyle().Foreground(style.BrightRed).Bold(true),
		ItemCheckedStyle:    style.NewStyle().Foreground(style.BrightGreen).Bold(true),
		SeparatorStyle:      style.NewStyle().Foreground(style.BrightBlack),
		ShortcutStyle:       style.NewStyle().Foreground(style.BrightCyan),
		DescriptionStyle:    style.NewStyle().Foreground(style.BrightBlack),
		IconStyle:           style.NewStyle().Foreground(style.BrightBlue),
		CheckmarkStyle:      style.NewStyle().Foreground(style.BrightGreen).Bold(true),
		SubmenuArrowStyle:   style.NewStyle().Foreground(style.BrightWhite),
		TitleStyle:          style.NewStyle().Foreground(style.BrightWhite).Bold(true),
		ScrollbarStyle:      style.NewStyle().Foreground(style.BrightBlack),
		ScrollbarThumbStyle: style.NewStyle().Foreground(style.BrightWhite),
	}
}

func (t Theme) WithBarStyle(s style.Style) Theme            { t.BarStyle = s; return t }
func (t Theme) WithBarActiveStyle(s style.Style) Theme      { t.BarActiveStyle = s; return t }
func (t Theme) WithSurfaceStyle(s style.Style) Theme        { t.SurfaceStyle = s; return t }
func (t Theme) WithSurfaceBorderStyle(s style.Style) Theme  { t.SurfaceBorderStyle = s; return t }
func (t Theme) WithSurfaceShadowStyle(s style.Style) Theme  { t.SurfaceShadowStyle = s; return t }
func (t Theme) WithItemStyle(s style.Style) Theme           { t.ItemStyle = s; return t }
func (t Theme) WithItemHoverStyle(s style.Style) Theme      { t.ItemHoverStyle = s; return t }
func (t Theme) WithItemFocusStyle(s style.Style) Theme      { t.ItemFocusStyle = s; return t }
func (t Theme) WithItemActiveStyle(s style.Style) Theme     { t.ItemActiveStyle = s; return t }
func (t Theme) WithItemDisabledStyle(s style.Style) Theme   { t.ItemDisabledStyle = s; return t }
func (t Theme) WithItemDangerStyle(s style.Style) Theme     { t.ItemDangerStyle = s; return t }
func (t Theme) WithItemCheckedStyle(s style.Style) Theme    { t.ItemCheckedStyle = s; return t }
func (t Theme) WithSeparatorStyle(s style.Style) Theme      { t.SeparatorStyle = s; return t }
func (t Theme) WithShortcutStyle(s style.Style) Theme       { t.ShortcutStyle = s; return t }
func (t Theme) WithDescriptionStyle(s style.Style) Theme    { t.DescriptionStyle = s; return t }
func (t Theme) WithIconStyle(s style.Style) Theme           { t.IconStyle = s; return t }
func (t Theme) WithCheckmarkStyle(s style.Style) Theme      { t.CheckmarkStyle = s; return t }
func (t Theme) WithSubmenuArrowStyle(s style.Style) Theme   { t.SubmenuArrowStyle = s; return t }
func (t Theme) WithTitleStyle(s style.Style) Theme          { t.TitleStyle = s; return t }
func (t Theme) WithScrollbarStyle(s style.Style) Theme      { t.ScrollbarStyle = s; return t }
func (t Theme) WithScrollbarThumbStyle(s style.Style) Theme { t.ScrollbarThumbStyle = s; return t }

func isZeroTheme(t Theme) bool {
	return t == Theme{}
}
