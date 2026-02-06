// Package theme provides theme color access for components
//
// This package implements a complete semantic color system based on the
// TUI design system specification. It supports multiple built-in themes
// and allows runtime theme switching.
//
// Semantic Color Model:
//
//	Layer System:      BG, SURFACE, OVERLAY
//	Typography:        TEXT, MUTED, PLACEHOLDER
//	Brand & Action:    PRIMARY, SECONDARY, ACCENT
//	State:             SUCCESS, WARNING, ERROR
//	Content Relations: LINK, VISITED
//	Boundaries:        BORDER, FOCUS, SELECT, HIGHLIGHT
//	Disabled:          DISABLED_BG, DISABLED_FG
//	System UI:         SCROLLBAR, SHADOW, CARET
package theme

import (
	"sync"

	"github.com/wwsheng009/mint/runtime/style"
)

// ColorName represents semantic color names based on the TUI design system
type ColorName string

const (
	// Layer System - 构建空间层次
	ColorBG      ColorName = "bg"
	ColorSurface ColorName = "surface"
	ColorOverlay ColorName = "overlay"

	// Typography - 信息层级
	ColorText       ColorName = "text"
	ColorMuted      ColorName = "muted"
	ColorPlaceholder ColorName = "placeholder"

	// Brand & Action - 操作优先级
	ColorPrimary   ColorName = "primary"
	ColorSecondary ColorName = "secondary"
	ColorAccent    ColorName = "accent"

	// State - 系统反馈
	ColorSuccess ColorName = "success"
	ColorWarning ColorName = "warning"
	ColorError   ColorName = "error"

	// Content Relations - 内容连接
	ColorLink    ColorName = "link"
	ColorVisited ColorName = "visited"

	// Boundaries - 交互定位
	ColorBorder   ColorName = "border"
	ColorFocus   ColorName = "focus"
	ColorSelect  ColorName = "select"
	ColorHighlight ColorName = "highlight"

	// Disabled - 禁用状态
	ColorDisabledBG ColorName = "disabled-bg"
	ColorDisabledFG ColorName = "disabled-fg"

	// System UI - 系统控件
	ColorScrollbar ColorName = "scrollbar"
	ColorShadow   ColorName = "shadow"
	ColorCaret    ColorName = "caret"
)

// Legacy color name aliases for backward compatibility
const (
	ColorForeground = ColorText
	ColorBackground = ColorBG
	ColorInfo       = ColorPrimary
	ColorActive     = ColorSelect
	ColorHover      = ColorSelect
)

// ThemePreset represents a built-in theme preset
type ThemePreset struct {
	Name   string
	colors map[ColorName]string
}

// Theme holds color definitions for semantic names
type Theme struct {
	preset  *ThemePreset
	custom  map[ColorName]string
	mu      sync.RWMutex
	current string // current preset name
}

var globalTheme = &Theme{
	custom:  make(map[ColorName]string),
	current: "default",
}

// =============================================================================
// Theme Presets
// =============================================================================

var themePresets = map[string]ThemePreset{
	"default": {
		Name: "default",
		colors: map[ColorName]string{
			ColorBG:          "black",
			ColorSurface:     "bright-black",
			ColorOverlay:     "bright-black",
			ColorText:        "white",
			ColorMuted:       "bright-black",
			ColorPlaceholder: "bright-black",
			ColorPrimary:     "blue",
			ColorSecondary:   "cyan",
			ColorAccent:      "yellow",
			ColorSuccess:     "green",
			ColorWarning:     "yellow",
			ColorError:       "red",
			ColorLink:        "blue",
			ColorVisited:     "magenta",
			ColorBorder:      "bright-black",
			ColorFocus:       "blue",
			ColorSelect:      "blue",
			ColorHighlight:   "yellow",
			ColorDisabledBG:  "bright-black",
			ColorDisabledFG:  "bright-black",
			ColorScrollbar:   "bright-black",
			ColorShadow:      "black",
			ColorCaret:       "white",
		},
	},
	"nord": {
		Name: "nord",
		colors: map[ColorName]string{
			ColorBG:          "#2E3440", // nord0
			ColorSurface:     "#3B4252", // nord1
			ColorOverlay:     "#282C34", // darker
			ColorText:        "#ECEFF4", // nord4
			ColorMuted:       "#616E88", // nord10
			ColorPlaceholder: "#616E88",
			ColorPrimary:     "#88C0D0", // nord8
			ColorSecondary:   "#81A1C1", // nord9
			ColorAccent:      "#8FBCBB", // nord7
			ColorSuccess:     "#A3BE8C", // nord14
			ColorWarning:     "#EBCB8B", // nord13
			ColorError:       "#BF616A", // nord11
			ColorLink:        "#88C0D0",
			ColorVisited:     "#B48EAD",
			ColorBorder:      "#4C566A", // nord3
			ColorFocus:       "#88C0D0",
			ColorSelect:      "#81A1C1",
			ColorHighlight:   "#EBCB8B",
			ColorDisabledBG:  "#3B4252",
			ColorDisabledFG:  "#616E88",
			ColorScrollbar:   "#4C566A",
			ColorShadow:      "#1E2028",
			ColorCaret:       "#ECEFF4",
		},
	},
	"dracula": {
		Name: "dracula",
		colors: map[ColorName]string{
			ColorBG:          "#282A36",
			ColorSurface:     "#343746",
			ColorOverlay:     "#242733",
			ColorText:        "#F8F8F2",
			ColorMuted:       "#6272A4",
			ColorPlaceholder: "#6272A4",
			ColorPrimary:     "#BD93F9",
			ColorSecondary:   "#8BE9FD",
			ColorAccent:      "#FF79C6",
			ColorSuccess:     "#50FA7B",
			ColorWarning:     "#F1FA8C",
			ColorError:       "#FF5555",
			ColorLink:        "#8BE9FD",
			ColorVisited:     "#BD93F9",
			ColorBorder:      "#44475A",
			ColorFocus:       "#BD93F9",
			ColorSelect:      "#FF79C6",
			ColorHighlight:   "#F1FA8C",
			ColorDisabledBG:  "#343746",
			ColorDisabledFG:  "#6272A4",
			ColorScrollbar:   "#44475A",
			ColorShadow:      "#1C1D26",
			ColorCaret:       "#F8F8F2",
		},
	},
	"gruvbox-dark": {
		Name: "gruvbox-dark",
		colors: map[ColorName]string{
			ColorBG:          "#282828",
			ColorSurface:     "#3C3836",
			ColorOverlay:     "#201E1D",
			ColorText:        "#EBDBB2",
			ColorMuted:       "#928374",
			ColorPlaceholder: "#928374",
			ColorPrimary:     "#83A598",
			ColorSecondary:   "#8EC07C",
			ColorAccent:      "#D3869B",
			ColorSuccess:     "#B8BB26",
			ColorWarning:     "#FABD2F",
			ColorError:       "#FB4934",
			ColorLink:        "#83A598",
			ColorVisited:     "#D3869B",
			ColorBorder:      "#504945",
			ColorFocus:       "#83A598",
			ColorSelect:      "#D3869B",
			ColorHighlight:   "#FABD2F",
			ColorDisabledBG:  "#3C3836",
			ColorDisabledFG:  "#928374",
			ColorScrollbar:   "#504945",
			ColorShadow:      "#1C1B1A",
			ColorCaret:       "#EBDBB2",
		},
	},
	"catppuccin-mocha": {
		Name: "catppuccin-mocha",
		colors: map[ColorName]string{
			ColorBG:          "#1E1E2E",
			ColorSurface:     "#313244",
			ColorOverlay:     "#181825",
			ColorText:        "#CDD6F4",
			ColorMuted:       "#6C7086",
			ColorPlaceholder: "#6C7086",
			ColorPrimary:     "#89B4FA",
			ColorSecondary:   "#CBA6F7",
			ColorAccent:      "#F5C2E7",
			ColorSuccess:     "#A6E3A1",
			ColorWarning:     "#F9E2AF",
			ColorError:       "#F38BA8",
			ColorLink:        "#89B4FA",
			ColorVisited:     "#CBA6F7",
			ColorBorder:      "#585B70",
			ColorFocus:       "#89B4FA",
			ColorSelect:      "#F5C2E7",
			ColorHighlight:   "#F9E2AF",
			ColorDisabledBG:  "#313244",
			ColorDisabledFG:  "#6C7086",
			ColorScrollbar:   "#585B70",
			ColorShadow:      "#161621",
			ColorCaret:       "#CDD6F4",
		},
	},
	"solarized-dark": {
		Name: "solarized-dark",
		colors: map[ColorName]string{
			ColorBG:          "#002B36",
			ColorSurface:     "#073642",
			ColorOverlay:     "#00242E",
			ColorText:        "#EEE8D5",
			ColorMuted:       "#839496",
			ColorPlaceholder: "#839496",
			ColorPrimary:     "#268BD2",
			ColorSecondary:   "#2AA198",
			ColorAccent:      "#6C71C4",
			ColorSuccess:     "#859900",
			ColorWarning:     "#B58900",
			ColorError:       "#DC322F",
			ColorLink:        "#268BD2",
			ColorVisited:     "#6C71C4",
			ColorBorder:      "#586E75",
			ColorFocus:       "#268BD2",
			ColorSelect:      "#2AA198",
			ColorHighlight:   "#B58900",
			ColorDisabledBG:  "#073642",
			ColorDisabledFG:  "#839496",
			ColorScrollbar:   "#586E75",
			ColorShadow:      "#001E26",
			ColorCaret:       "#EEE8D5",
		},
	},
}

// =============================================================================
// Theme Access API
// =============================================================================

func init() {
	// Initialize with default preset
	preset := themePresets["default"]
	globalTheme.preset = &preset
}

// SetPreset switches to a built-in theme preset
func SetPreset(name string) bool {
	globalTheme.mu.Lock()
	defer globalTheme.mu.Unlock()

	if preset, ok := themePresets[name]; ok {
		globalTheme.preset = &preset
		globalTheme.current = name
		// Clear custom colors when switching presets
		globalTheme.custom = make(map[ColorName]string)
		return true
	}
	return false
}

// CurrentPreset returns the current preset name
func CurrentPreset() string {
	globalTheme.mu.RLock()
	defer globalTheme.mu.RUnlock()
	return globalTheme.current
}

// PresetNames returns all available preset names
func PresetNames() []string {
	names := make([]string, 0, len(themePresets))
	for name := range themePresets {
		names = append(names, name)
	}
	return names
}

// =============================================================================
// Color Access
// =============================================================================

// GetColor returns the style.Color for a semantic color name
func GetColor(name ColorName) style.Color {
	globalTheme.mu.RLock()
	defer globalTheme.mu.RUnlock()

	// Check custom override first
	if custom, ok := globalTheme.custom[name]; ok {
		return style.Color(custom)
	}

	// Use preset colors
	if globalTheme.preset != nil {
		if color, ok := globalTheme.preset.colors[name]; ok {
			return style.Color(color)
		}
	}

	// Fallback
	return style.Color("white")
}

// GetColorString returns the color string for a semantic color name
func GetColorString(name ColorName) string {
	globalTheme.mu.RLock()
	defer globalTheme.mu.RUnlock()

	// Check custom override first
	if custom, ok := globalTheme.custom[name]; ok {
		return custom
	}

	// Use preset colors
	if globalTheme.preset != nil {
		if color, ok := globalTheme.preset.colors[name]; ok {
			return color
		}
	}

	return "white"
}

// SetColor sets a custom theme color override
func SetColor(name ColorName, color string) {
	globalTheme.mu.Lock()
	defer globalTheme.mu.Unlock()

	if globalTheme.custom == nil {
		globalTheme.custom = make(map[ColorName]string)
	}
	globalTheme.custom[name] = color
}

// SetColors sets multiple custom theme colors at once
func SetColors(colors map[ColorName]string) {
	globalTheme.mu.Lock()
	defer globalTheme.mu.Unlock()

	if globalTheme.custom == nil {
		globalTheme.custom = make(map[ColorName]string)
	}
	for k, v := range colors {
		globalTheme.custom[k] = v
	}
}

// ClearCustomColors clears all custom color overrides
func ClearCustomColors() {
	globalTheme.mu.Lock()
	defer globalTheme.mu.Unlock()

	globalTheme.custom = make(map[ColorName]string)
}

// =============================================================================
// Convenience Functions for Common Semantic Colors
// =============================================================================

// Layer System
func BG() style.Color      { return GetColor(ColorBG) }
func Surface() style.Color { return GetColor(ColorSurface) }
func Overlay() style.Color { return GetColor(ColorOverlay) }

// Typography
func Text() style.Color       { return GetColor(ColorText) }
func Muted() style.Color      { return GetColor(ColorMuted) }
func Placeholder() style.Color { return GetColor(ColorPlaceholder) }

// Brand & Action
func Primary() style.Color { return GetColor(ColorPrimary) }
func Secondary() style.Color { return GetColor(ColorSecondary) }
func Accent() style.Color { return GetColor(ColorAccent) }

// State
func Success() style.Color { return GetColor(ColorSuccess) }
func Warning() style.Color { return GetColor(ColorWarning) }
func Error() style.Color { return GetColor(ColorError) }

// Content Relations
func Link() style.Color    { return GetColor(ColorLink) }
func Visited() style.Color { return GetColor(ColorVisited) }

// Boundaries
func Border() style.Color    { return GetColor(ColorBorder) }
func Focus() style.Color     { return GetColor(ColorFocus) }
func Select() style.Color    { return GetColor(ColorSelect) }
func Highlight() style.Color { return GetColor(ColorHighlight) }

// Disabled
func DisabledBG() style.Color { return GetColor(ColorDisabledBG) }
func DisabledFG() style.Color { return GetColor(ColorDisabledFG) }

// System UI
func Scrollbar() style.Color { return GetColor(ColorScrollbar) }
func Shadow() style.Color    { return GetColor(ColorShadow) }
func Caret() style.Color     { return GetColor(ColorCaret) }

// =============================================================================
// Legacy Aliases for Backward Compatibility
// =============================================================================

// Foreground returns the primary text color
func Foreground() style.Color { return GetColor(ColorText) }

// Background returns the global background color
func Background() style.Color { return GetColor(ColorBG) }

// Disabled returns the disabled foreground color
func Disabled() style.Color { return GetColor(ColorDisabledFG) }

// Hover returns the select color (for hover state)
func Hover() style.Color { return GetColor(ColorSelect) }

// Active returns the select color (for active state)
func Active() style.Color { return GetColor(ColorSelect) }

// Info returns the primary color
func Info() style.Color { return GetColor(ColorPrimary) }

// FocusBright returns a bright variant for high visibility focus
func FocusBright() style.Color {
	// Use bright-yellow for maximum visibility on any background
	return style.Color("bright-yellow")
}
