package style

// =============================================================================
// 便捷构造函数
// =============================================================================

// New creates a new empty style.
// This is a shorter alias for Style{}.
func New() Style {
	return Style{}
}

// Foreground creates a style with only foreground color.
func Foreground(c Color) Style {
	return Style{FG: c}
}

// Background creates a style with only background color.
func Background(c Color) Style {
	return Style{BG: c}
}

// Bold creates a bold style.
func Bold() Style {
	return Style{isBold: true}
}

// Italic creates an italic style.
func Italic() Style {
	return Style{isItalic: true}
}

// Underline creates an underlined style.
func Underline() Style {
	return Style{isUnderline: true}
}

// Reverse creates a reversed (foreground/background swapped) style.
func Reverse() Style {
	return Style{isReverse: true}
}

// =============================================================================
// 组合构造函数
// =============================================================================

// FgBold creates a style with foreground color and bold.
func FgBold(c Color) Style {
	return Style{FG: c, isBold: true}
}

// FgBg creates a style with foreground and background colors.
func FgBg(fg, bg Color) Style {
	return Style{FG: fg, BG: bg}
}

// FgBgBold creates a style with foreground, background, and bold.
func FgBgBold(fg, bg Color) Style {
	return Style{FG: fg, BG: bg, isBold: true}
}

// FgUnderline creates a style with foreground color and underline.
func FgUnderline(c Color) Style {
	return Style{FG: c, isUnderline: true}
}

// FgBgUnderline creates a style with foreground, background, and underline.
func FgBgUnderline(fg, bg Color) Style {
	return Style{FG: fg, BG: bg, isUnderline: true}
}

// =============================================================================
// 常用样式预设
// =============================================================================

// None represents an empty style.
// This is the semantic zero value for Style.
var None = Style{}

// ReverseStyle is a predefined reversed style.
var ReverseStyle = Reverse()

// BoldStyle is a predefined bold style.
var BoldStyle = Bold()

// UnderlineStyle is a predefined underlined style.
var UnderlineStyle = Underline()
