package statusbar

import (
	"fmt"
	"strings"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// OverflowMode controls how fixed-width sections handle overflow.
type OverflowMode int

const (
	// OverflowEllipsis truncates the content and appends an ellipsis.
	OverflowEllipsis OverflowMode = iota
	// OverflowClip truncates the content without an ellipsis.
	OverflowClip
)

// HelpDisplayMode controls how statusbar help text is rendered.
type HelpDisplayMode int

const (
	// HelpDisplayInline renders help as a secondary line below the status bar.
	HelpDisplayInline HelpDisplayMode = iota
	// HelpDisplayOverlay renders help as a floating tooltip on the tooltip layer.
	HelpDisplayOverlay
	// HelpDisplayBoth renders both inline help and overlay tooltip.
	HelpDisplayBoth
)

// TooltipPlacement controls where overlay tooltips are placed relative to the anchor.
type TooltipPlacement int

const (
	// TooltipPlacementAuto chooses top or bottom automatically.
	TooltipPlacementAuto TooltipPlacement = iota
	// TooltipPlacementTop forces the tooltip above the anchor.
	TooltipPlacementTop
	// TooltipPlacementBottom forces the tooltip below the anchor.
	TooltipPlacementBottom
)

// TooltipArrowStyle controls the visual style of overlay tooltip borders and arrows.
type TooltipArrowStyle int

const (
	// TooltipArrowStyleDefault falls back to the theme default.
	TooltipArrowStyleDefault TooltipArrowStyle = iota
	// TooltipArrowStyleSharp uses square corners and solid triangle arrows.
	TooltipArrowStyleSharp
	// TooltipArrowStyleRounded uses rounded corners and lighter triangle arrows.
	TooltipArrowStyleRounded
)

// Section represents a text segment in the status bar.
//
// If Width is greater than zero, Text is padded or truncated to that exact
// display width. By default, fixed-width text follows the slot alignment and
// uses ellipsis overflow. Use WithAlign or WithOverflow to override that.
type Section struct {
	Key         string
	Text        string
	FgColor     string
	BgColor     string
	Bold        bool
	Width       int
	Align       rtui.Align
	Overflow    OverflowMode
	PressIntent intent.Intent
	Disabled    bool
	HelpText    string

	alignSet  bool
	boldSet   bool
	helpKey   string
	helpOrder int
	helpModel *helpModel
}

// S creates a plain section.
func S(content string) Section {
	return Section{Text: content}
}

// Text creates a plain section.
func Text(content string) Section {
	return S(content)
}

// ActionText creates an interactive plain section.
func ActionText(content string, pressIntent intent.Intent) Section {
	return Text(content).OnPress(pressIntent)
}

// Sections creates a slice of sections for direct shortcut APIs.
func Sections(sections ...Section) []Section {
	return append([]Section(nil), sections...)
}

// Badge creates a highlighted section.
func Badge(content, fgColor, bgColor string) Section {
	return Section{
		Text:    content,
		FgColor: fgColor,
		BgColor: bgColor,
		Bold:    true,
		boldSet: true,
	}
}

// ActionBadge creates an interactive highlighted section.
func ActionBadge(content, fgColor, bgColor string, pressIntent intent.Intent) Section {
	return Badge(content, fgColor, bgColor).OnPress(pressIntent)
}

// WithKey sets a stable key for diffing.
func (s Section) WithKey(key string) Section {
	s.Key = key
	return s
}

// WithColors sets both foreground and background colors.
func (s Section) WithColors(fgColor, bgColor string) Section {
	s.FgColor = fgColor
	s.BgColor = bgColor
	return s
}

// WithForeground sets the foreground color.
func (s Section) WithForeground(fgColor string) Section {
	s.FgColor = fgColor
	return s
}

// WithBackground sets the background color.
func (s Section) WithBackground(bgColor string) Section {
	s.BgColor = bgColor
	return s
}

// WithBold sets whether the section should render bold.
func (s Section) WithBold(bold bool) Section {
	s.Bold = bold
	s.boldSet = true
	return s
}

// WithWidth sets the display width for the section.
func (s Section) WithWidth(width int) Section {
	s.Width = clampNonNegative(width)
	return s
}

// WithAlign sets the alignment used when Width is specified.
func (s Section) WithAlign(align rtui.Align) Section {
	s.Align = align
	s.alignSet = true
	return s
}

// WithOverflow sets the overflow behavior used when Width is specified.
func (s Section) WithOverflow(mode OverflowMode) Section {
	s.Overflow = mode
	return s
}

// WithEllipsis truncates overflowing text and appends an ellipsis.
func (s Section) WithEllipsis() Section {
	return s.WithOverflow(OverflowEllipsis)
}

// WithClip truncates overflowing text without appending an ellipsis.
func (s Section) WithClip() Section {
	return s.WithOverflow(OverflowClip)
}

// OnPress makes the section clickable and emits the given intent.
func (s Section) OnPress(pressIntent intent.Intent) Section {
	s.PressIntent = pressIntent
	return s
}

// WithDisabled sets the disabled state for interactive sections.
func (s Section) WithDisabled(disabled bool) Section {
	s.Disabled = disabled
	return s
}

// WithHelp sets the inline tooltip/help text for the section.
func (s Section) WithHelp(helpText string) Section {
	s.HelpText = normalizeStatusText(helpText)
	return s
}

// WithTooltip is an alias for WithHelp.
func (s Section) WithTooltip(tooltipText string) Section {
	return s.WithHelp(tooltipText)
}

// Theme provides default styling for sections that don't specify colors.
type Theme struct {
	FgColor            string
	BgColor            string
	Bold               bool
	HoverStyle         style.Style
	FocusStyle         style.Style
	PressedStyle       style.Style
	DisabledStyle      style.Style
	HelpStyle          style.Style
	TooltipBorderStyle style.Style
	TooltipShadowStyle style.Style
	TooltipArrowStyle  TooltipArrowStyle
}

// WithHoverStyle sets the style overlay used for hovered sections.
func (t Theme) WithHoverStyle(s style.Style) Theme {
	t.HoverStyle = s
	return t
}

// WithFocusStyle sets the style overlay used for focused sections.
func (t Theme) WithFocusStyle(s style.Style) Theme {
	t.FocusStyle = s
	return t
}

// WithPressedStyle sets the style overlay used for pressed sections.
func (t Theme) WithPressedStyle(s style.Style) Theme {
	t.PressedStyle = s
	return t
}

// WithDisabledStyle sets the style overlay used for disabled sections.
func (t Theme) WithDisabledStyle(s style.Style) Theme {
	t.DisabledStyle = s
	return t
}

// WithHelpStyle sets the style used for the inline help line.
func (t Theme) WithHelpStyle(s style.Style) Theme {
	t.HelpStyle = s
	return t
}

// WithTooltipBorderStyle sets the border style used by overlay tooltips.
func (t Theme) WithTooltipBorderStyle(s style.Style) Theme {
	t.TooltipBorderStyle = s
	return t
}

// WithTooltipShadowStyle sets the shadow style used by overlay tooltips.
func (t Theme) WithTooltipShadowStyle(s style.Style) Theme {
	t.TooltipShadowStyle = s
	return t
}

// WithTooltipArrowStyle sets the border/arrow chrome used by overlay tooltips.
func (t Theme) WithTooltipArrowStyle(arrowStyle TooltipArrowStyle) Theme {
	t.TooltipArrowStyle = arrowStyle
	return t
}

// DefaultTheme returns a neutral status bar theme.
func DefaultTheme() Theme {
	return Theme{
		FgColor:            "white",
		BgColor:            "blue",
		HoverStyle:         style.NewStyle().Underline(true),
		FocusStyle:         style.NewStyle().Underline(true).Bold(true),
		PressedStyle:       style.NewStyle().Reverse(true),
		DisabledStyle:      style.NewStyle().Foreground(style.BrightBlack),
		HelpStyle:          style.NewStyle().Foreground(style.White).Background(style.Blue),
		TooltipBorderStyle: style.NewStyle().Foreground(style.White).Background(style.Blue).Bold(true),
		TooltipShadowStyle: style.NewStyle().Foreground(style.BrightBlack).Background(style.Blue),
		TooltipArrowStyle:  TooltipArrowStyleSharp,
	}
}

// MutedTheme returns a lower-contrast theme.
func MutedTheme() Theme {
	return Theme{
		FgColor:            "bright-white",
		BgColor:            "bright-black",
		HoverStyle:         style.NewStyle().Underline(true),
		FocusStyle:         style.NewStyle().Underline(true).Bold(true),
		PressedStyle:       style.NewStyle().Reverse(true),
		DisabledStyle:      style.NewStyle().Foreground(style.BrightBlack),
		HelpStyle:          style.NewStyle().Foreground(style.BrightWhite).Background(style.BrightBlack),
		TooltipBorderStyle: style.NewStyle().Foreground(style.BrightWhite).Background(style.BrightBlack).Bold(true),
		TooltipShadowStyle: style.NewStyle().Foreground(style.Black).Background(style.BrightBlack),
		TooltipArrowStyle:  TooltipArrowStyleRounded,
	}
}

// ContrastTheme returns a strong contrast theme.
func ContrastTheme() Theme {
	return Theme{
		FgColor:            "black",
		BgColor:            "yellow",
		Bold:               true,
		HoverStyle:         style.NewStyle().Underline(true),
		FocusStyle:         style.NewStyle().Underline(true).Bold(true),
		PressedStyle:       style.NewStyle().Reverse(true),
		DisabledStyle:      style.NewStyle().Foreground(style.BrightBlack),
		HelpStyle:          style.NewStyle().Foreground(style.Black).Background(style.Yellow).Bold(true),
		TooltipBorderStyle: style.NewStyle().Foreground(style.Black).Background(style.Yellow).Bold(true),
		TooltipShadowStyle: style.NewStyle().Foreground(style.BrightBlack).Background(style.Yellow),
		TooltipArrowStyle:  TooltipArrowStyleSharp,
	}
}

// Builder creates a reusable status bar with left/center/right slots.
//
// Each slot gets equal layout width, so center sections stay visually centered
// even when left and right content lengths differ.
type Builder struct {
	left             []Section
	center           []Section
	right            []Section
	gap              int
	padding          [4]int
	theme            Theme
	useTheme         bool
	helpFallback     string
	helpPrefix       string
	helpStyle        style.Style
	helpDisplay      HelpDisplayMode
	tooltipPlacement TooltipPlacement
	tooltipMaxWidth  int
	tooltipGapRows   int
}

// NewBuilder creates a new status bar builder.
func NewBuilder() *Builder {
	return &Builder{gap: 0, helpDisplay: HelpDisplayInline, tooltipPlacement: TooltipPlacementAuto, tooltipMaxWidth: 48, tooltipGapRows: 1}
}

// Left appends one section to the left slot.
func (b *Builder) Left(section Section) *Builder {
	b.left = append(b.left, section)
	return b
}

// Center appends one section to the center slot.
func (b *Builder) Center(section Section) *Builder {
	b.center = append(b.center, section)
	return b
}

// Right appends one section to the right slot.
func (b *Builder) Right(section Section) *Builder {
	b.right = append(b.right, section)
	return b
}

// LeftSections sets the left-aligned sections.
func (b *Builder) LeftSections(sections ...Section) *Builder {
	b.left = append([]Section(nil), sections...)
	return b
}

// CenterSections sets the center-aligned sections.
func (b *Builder) CenterSections(sections ...Section) *Builder {
	b.center = append([]Section(nil), sections...)
	return b
}

// RightSections sets the right-aligned sections.
func (b *Builder) RightSections(sections ...Section) *Builder {
	b.right = append([]Section(nil), sections...)
	return b
}

// LeftText appends a plain text section on the left.
func (b *Builder) LeftText(content string) *Builder {
	return b.Left(S(content))
}

// LeftBadge appends a highlighted section on the left.
func (b *Builder) LeftBadge(content, fgColor, bgColor string) *Builder {
	return b.Left(Badge(content, fgColor, bgColor))
}

// CenterText appends a plain text section in the center slot.
func (b *Builder) CenterText(content string) *Builder {
	return b.Center(S(content))
}

// CenterBadge appends a highlighted section in the center slot.
func (b *Builder) CenterBadge(content, fgColor, bgColor string) *Builder {
	return b.Center(Badge(content, fgColor, bgColor))
}

// RightText appends a plain text section on the right.
func (b *Builder) RightText(content string) *Builder {
	return b.Right(S(content))
}

// RightBadge appends a highlighted section on the right.
func (b *Builder) RightBadge(content, fgColor, bgColor string) *Builder {
	return b.Right(Badge(content, fgColor, bgColor))
}

// Gap sets the spacing between segments inside each slot.
func (b *Builder) Gap(n int) *Builder {
	b.gap = clampNonNegative(n)
	return b
}

// Padding sets outer padding around the whole status bar.
func (b *Builder) Padding(top, right, bottom, left int) *Builder {
	b.padding = [4]int{
		clampNonNegative(top),
		clampNonNegative(right),
		clampNonNegative(bottom),
		clampNonNegative(left),
	}
	return b
}

// Theme applies default styling to sections with no explicit style.
func (b *Builder) Theme(theme Theme) *Builder {
	b.theme = theme
	b.useTheme = true
	return b
}

// DefaultTheme applies the default status bar theme.
func (b *Builder) DefaultTheme() *Builder {
	return b.Theme(DefaultTheme())
}

// MutedTheme applies the muted status bar theme.
func (b *Builder) MutedTheme() *Builder {
	return b.Theme(MutedTheme())
}

// ContrastTheme applies the contrast status bar theme.
func (b *Builder) ContrastTheme() *Builder {
	return b.Theme(ContrastTheme())
}

// HelpFallback sets the fallback help text shown when no section is active.
func (b *Builder) HelpFallback(text string) *Builder {
	b.helpFallback = normalizeStatusText(text)
	return b
}

// HelpPrefix sets the prefix shown before help text.
func (b *Builder) HelpPrefix(prefix string) *Builder {
	b.helpPrefix = normalizeStatusText(prefix)
	return b
}

// HelpStyle sets the style for the help line rendered by BuildWithHelp.
func (b *Builder) HelpStyle(s style.Style) *Builder {
	b.helpStyle = s
	return b
}

// HelpDisplayMode selects whether help is shown inline, as overlay tooltip, or both.
func (b *Builder) HelpDisplayMode(mode HelpDisplayMode) *Builder {
	b.helpDisplay = mode
	return b
}

// TooltipPlacement sets overlay tooltip placement policy.
func (b *Builder) TooltipPlacement(placement TooltipPlacement) *Builder {
	b.tooltipPlacement = placement
	return b
}

// TooltipMaxWidth sets the max content width for wrapped overlay tooltips.
func (b *Builder) TooltipMaxWidth(width int) *Builder {
	b.tooltipMaxWidth = clampNonNegative(width)
	return b
}

// TooltipGapRows sets the vertical gap between the anchor and overlay tooltip.
func (b *Builder) TooltipGapRows(rows int) *Builder {
	b.tooltipGapRows = clampNonNegative(rows)
	return b
}

// Build returns the composed single-line status bar VNode.
func (b *Builder) Build() rtui.VNode {
	return b.buildBar(b.left, b.center, b.right)
}

// BuildWithHelp returns the status bar with help rendered according to HelpDisplayMode.
func (b *Builder) BuildWithHelp() rtui.VNode {
	left, center, right, model, hasHelp := b.prepareHelpSections()
	bar := b.buildBar(left, center, right)
	if !hasHelp {
		return bar
	}

	content := bar
	if b.helpDisplay == HelpDisplayInline || b.helpDisplay == HelpDisplayBoth {
		content = rtui.VStackBuilder(
			bar,
			newHelpLineVNode(model, b.resolveHelpStyle()),
		).Gap(0).Build()
	}
	if b.helpDisplay == HelpDisplayOverlay || b.helpDisplay == HelpDisplayBoth {
		overlayOffset := 0
		if b.helpDisplay == HelpDisplayBoth {
			overlayOffset = 1
		}
		return rtui.Fragment(content, newOverlayHelpVNode(model, b.resolveHelpStyle(), b.resolveTooltipBorderStyle(), b.resolveTooltipShadowStyle(), b.resolveTooltipArrowStyle(), b.tooltipPlacement, b.tooltipMaxWidth, b.tooltipGapRows, overlayOffset))
	}
	return content
}

func (b *Builder) buildBar(left, center, right []Section) rtui.VNode {
	return rtui.HStackBuilder(
		buildSlot(left, rtui.AlignStart, b.gap, b.theme, b.useTheme),
		buildSlot(center, rtui.AlignCenter, b.gap, b.theme, b.useTheme),
		buildSlot(right, rtui.AlignEnd, b.gap, b.theme, b.useTheme),
	).Gap(0).
		Padding(b.padding[0], b.padding[1], b.padding[2], b.padding[3]).
		Build()
}

func (b *Builder) prepareHelpSections() (left, center, right []Section, model *helpModel, hasHelp bool) {
	model = newHelpModel(b.helpFallback, b.helpPrefix)
	order := 0
	prepare := func(source []Section) []Section {
		result := append([]Section(nil), source...)
		for i := range result {
			result[i].helpModel = model
			result[i].helpOrder = order
			result[i].helpKey = result[i].Key
			if result[i].helpKey == "" {
				result[i].helpKey = fmt.Sprintf("statusbar-help-%d", order)
			}
			if result[i].HelpText != "" {
				hasHelp = true
			}
			order++
		}
		return result
	}
	left = prepare(b.left)
	center = prepare(b.center)
	right = prepare(b.right)
	if b.helpFallback != "" {
		hasHelp = true
	}
	return left, center, right, model, hasHelp
}

func (b *Builder) resolveHelpStyle() style.Style {
	if !b.helpStyle.IsEmpty() {
		return b.helpStyle
	}
	if b.useTheme {
		theme := resolveThemeDefaults(b.theme)
		if !theme.HelpStyle.IsEmpty() {
			return theme.HelpStyle
		}
		return style.NewStyle().
			Foreground(style.Color(theme.FgColor)).
			Background(style.Color(theme.BgColor))
	}
	return resolveThemeDefaults(MutedTheme()).HelpStyle
}

func (b *Builder) resolveTooltipBorderStyle() style.Style {
	if b.useTheme {
		return resolveThemeDefaults(b.theme).TooltipBorderStyle
	}
	return resolveThemeDefaults(MutedTheme()).TooltipBorderStyle
}

func (b *Builder) resolveTooltipShadowStyle() style.Style {
	if b.useTheme {
		return resolveThemeDefaults(b.theme).TooltipShadowStyle
	}
	return resolveThemeDefaults(MutedTheme()).TooltipShadowStyle
}

func (b *Builder) resolveTooltipArrowStyle() TooltipArrowStyle {
	if b.useTheme {
		return resolveThemeDefaults(b.theme).TooltipArrowStyle
	}
	return resolveThemeDefaults(MutedTheme()).TooltipArrowStyle
}

func resolveThemeDefaults(theme Theme) Theme {
	defaults := DefaultTheme()
	if theme.HoverStyle.IsEmpty() {
		theme.HoverStyle = defaults.HoverStyle
	}
	if theme.FocusStyle.IsEmpty() {
		theme.FocusStyle = defaults.FocusStyle
	}
	if theme.PressedStyle.IsEmpty() {
		theme.PressedStyle = defaults.PressedStyle
	}
	if theme.DisabledStyle.IsEmpty() {
		theme.DisabledStyle = defaults.DisabledStyle
	}
	if theme.HelpStyle.IsEmpty() {
		theme.HelpStyle = defaults.HelpStyle
	}
	if theme.TooltipBorderStyle.IsEmpty() {
		theme.TooltipBorderStyle = defaults.TooltipBorderStyle
	}
	if theme.TooltipShadowStyle.IsEmpty() {
		theme.TooltipShadowStyle = defaults.TooltipShadowStyle
	}
	if theme.TooltipArrowStyle == TooltipArrowStyleDefault {
		theme.TooltipArrowStyle = defaults.TooltipArrowStyle
	}
	return theme
}

func buildSlot(sections []Section, slotAlign rtui.Align, gap int, theme Theme, useTheme bool) rtui.VNode {
	children := make([]rtui.VNode, 0, len(sections))
	for _, section := range sections {
		children = append(children, renderSection(section, slotAlign, theme, useTheme))
	}
	return rtui.HStackBuilder(children...).Gap(gap).Align(slotAlign).Flex(1).Build()
}

func renderSection(section Section, slotAlign rtui.Align, theme Theme, useTheme bool) rtui.VNode {
	resolved := resolveSection(section, slotAlign, theme, useTheme)
	resolvedStyle := style.NewStyle()
	if resolved.FgColor != "" {
		resolvedStyle = resolvedStyle.Foreground(style.Color(resolved.FgColor))
	}
	if resolved.BgColor != "" {
		resolvedStyle = resolvedStyle.Background(style.Color(resolved.BgColor))
	}
	if resolved.Bold {
		resolvedStyle = resolvedStyle.Bold(true)
	}
	stateTheme := resolveThemeDefaults(theme)
	return newSectionVNode(
		resolved.Key,
		resolved.Text,
		resolvedStyle,
		resolved.PressIntent,
		resolved.Disabled,
		resolved.HelpText,
		resolved.helpKey,
		resolved.helpOrder,
		resolved.helpModel,
		stateTheme.HoverStyle,
		stateTheme.FocusStyle,
		stateTheme.PressedStyle,
		stateTheme.DisabledStyle,
	)
}

func resolveSection(section Section, slotAlign rtui.Align, theme Theme, useTheme bool) Section {
	resolved := section
	resolved.Text = normalizeStatusText(resolved.Text)
	resolved.HelpText = normalizeStatusText(resolved.HelpText)
	if useTheme {
		if resolved.FgColor == "" {
			resolved.FgColor = theme.FgColor
		}
		if resolved.BgColor == "" {
			resolved.BgColor = theme.BgColor
		}
		if !resolved.boldSet && !resolved.Bold {
			resolved.Bold = theme.Bold
		}
	}
	if resolved.Width > 0 {
		align := slotAlign
		if resolved.alignSet || resolved.Align != rtui.AlignStart {
			align = resolved.Align
		}
		resolved.Text = fitText(resolved.Text, resolved.Width, align, resolved.Overflow)
	}
	return resolved
}

func fitText(content string, width int, align rtui.Align, overflow OverflowMode) string {
	content = normalizeStatusText(content)
	if width <= 0 {
		return content
	}

	displayWidth := paint.StringWidth(content)
	if displayWidth > width {
		switch overflow {
		case OverflowClip:
			content = truncateByDisplayWidth(content, width)
		default:
			content = truncateWithEllipsis(content, width)
		}
		displayWidth = paint.StringWidth(content)
	}

	padding := width - displayWidth
	if padding <= 0 {
		return content
	}

	switch align {
	case rtui.AlignEnd:
		return strings.Repeat(" ", padding) + content
	case rtui.AlignCenter:
		left := padding / 2
		right := padding - left
		return strings.Repeat(" ", left) + content + strings.Repeat(" ", right)
	default:
		return content + strings.Repeat(" ", padding)
	}
}

func truncateWithEllipsis(content string, width int) string {
	if width <= 0 {
		return ""
	}

	const ellipsis = "…"
	ellipsisWidth := paint.StringWidth(ellipsis)
	if width <= ellipsisWidth {
		return ellipsis
	}

	trimmed := strings.TrimRight(truncateByDisplayWidth(content, width-ellipsisWidth), " ")
	if trimmed == "" {
		return truncateByDisplayWidth(content, width)
	}
	return trimmed + ellipsis
}

func truncateByDisplayWidth(content string, width int) string {
	if width <= 0 {
		return ""
	}

	content = normalizeStatusText(content)
	var builder strings.Builder
	currentWidth := 0
	for _, r := range content {
		runeWidth := paint.RuneWidth(r)
		if currentWidth+runeWidth > width {
			break
		}
		builder.WriteRune(r)
		currentWidth += runeWidth
	}
	return builder.String()
}

func normalizeStatusText(content string) string {
	sanitized := paint.SanitizeForTerminal(content)
	return strings.NewReplacer("\r\n", " ", "\n", " ", "\r", " ", "\t", " ").Replace(sanitized)
}

func clampNonNegative(value int) int {
	if value < 0 {
		return 0
	}
	return value
}
