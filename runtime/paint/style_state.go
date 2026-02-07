package paint

import (
	"fmt"
	"strings"

	"github.com/wwsheng009/mint/runtime/style"
)

// StyleStateMachine minimizes VT code changes by tracking current style
type StyleStateMachine struct {
	current style.Style
}

// NewStyleStateMachine creates a new style state machine
func NewStyleStateMachine() *StyleStateMachine {
	return &StyleStateMachine{
		current: style.Style{},
	}
}

// Reset resets the current style
func (s *StyleStateMachine) Reset() {
	s.current = style.Style{}
}

// NeedsUpdate checks if the style needs to be updated
func (s *StyleStateMachine) NeedsUpdate(st style.Style) bool {
	return s.current != st
}

// Update generates VT codes to transition to the new style
func (s *StyleStateMachine) Update(st style.Style) string {
	codes := s.buildDiffCodes(s.current, st)
	s.current = st
	return codes
}

// buildDiffCodes builds only the codes that need to change
func (s *StyleStateMachine) buildDiffCodes(from, to style.Style) string {
	
	if from == to {
		return ""
	}

	// If completely different or target is empty, emit reset first
	if from == (style.Style{}) {
		return s.fullStyle(to)
	}

	// If target is empty, emit reset
	if to == (style.Style{}) {
		return "\x1b[0m"
	}

	var codes []string

	// Check if we need to turn OFF any attributes - this requires a reset
	needsReset := false
	if from.IsBold() && !to.IsBold() {
		needsReset = true
	}
	if from.IsItalic() && !to.IsItalic() {
		needsReset = true
	}
	if from.IsUnderline() && !to.IsUnderline() {
		needsReset = true
	}
	if from.IsReverse() && !to.IsReverse() {
		needsReset = true
	}
	if from.IsStrikethrough() && !to.IsStrikethrough() {
		needsReset = true
	}
	if from.IsBlink() && !to.IsBlink() {
		needsReset = true
	}
	if from.BG != "" && to.BG == "" {
		needsReset = true
	}
	if from.FG != "" && to.FG == "" {
		needsReset = true
	}

	// If we need to reset attributes, reset and apply full new style
	if needsReset {
		return "\x1b[0m" + s.fullStyle(to)
	}

	// Check for reset needed - if many things changed, just reset and start fresh
	changes := 0
	if from.FG != to.FG && to.FG != "" {
		changes++
	}
	if from.BG != to.BG && to.BG != "" {
		changes++
	}
	if from.IsBold() != to.IsBold() && to.IsBold() {
		changes++
	}
	if from.IsItalic() != to.IsItalic() && to.IsItalic() {
		changes++
	}
	if from.IsUnderline() != to.IsUnderline() && to.IsUnderline() {
		changes++
	}
	if from.IsReverse() != to.IsReverse() && to.IsReverse() {
		changes++
	}

	// If many changes, reset and rebuild
	if changes >= 4 {
		return "\x1b[0m" + s.fullStyle(to)
	}

	// Otherwise, emit only changes
	// Bold
	if to.IsBold() && !from.IsBold() {
		codes = append(codes, "1")
	}

	// Italic
	if to.IsItalic() && !from.IsItalic() {
		codes = append(codes, "3")
	}

	// Underline
	if to.IsUnderline() && !from.IsUnderline() {
		codes = append(codes, "4")
	}

	// Reverse
	if to.IsReverse() && !from.IsReverse() {
		codes = append(codes, "7")
	}

	// Colors
	// CRITICAL FIX: Always emit background color if target has background
	// This prevents issues where background is "unchanged" but should be re-applied after a reset
	if to.BG != "" {
		codes = append(codes, colorCode(to.BG, true))
	}
	// Foreground
	if from.FG != to.FG && to.FG != "" {
		codes = append(codes, colorCode(to.FG, false))
	}

	if len(codes) == 0 {
		return ""
	}

	return "\x1b[" + strings.Join(codes, ";") + "m"
}

// fullStyle generates full style codes
func (s *StyleStateMachine) fullStyle(st style.Style) string {
	var codes []string

	if st.IsBold() {
		codes = append(codes, "1")
	}
	if st.IsItalic() {
		codes = append(codes, "3")
	}
	if st.IsUnderline() {
		codes = append(codes, "4")
	}
	if st.IsReverse() {
		codes = append(codes, "7")
	}
	if st.FG != "" {
		codes = append(codes, colorCode(st.FG, false))
	}
	if st.BG != "" {
		codes = append(codes, colorCode(st.BG, true))
	}

	if len(codes) == 0 {
		return ""
	}

	return "\x1b[" + strings.Join(codes, ";") + "m"
}

// colorCode converts a color to ANSI code
// 正确处理标准色 (30-37/40-47)、明亮色 (90-97/100-107) 和 TrueColor (RGB)
func colorCode(color style.Color, isBackground bool) string {
	c := string(color)

	// Handle hex colors (#RRGGBB format) - convert to TrueColor ANSI
	if strings.HasPrefix(c, "#") {
		rgb, err := parseHexString(c)
		if err != nil {
			return ""
		}
		// TrueColor format: ESC[38;2;R;G;B for FG, ESC[48;2;R;G;B for BG
		if isBackground {
			return fmt.Sprintf("48;2;%d;%d;%d", rgb[0], rgb[1], rgb[2])
		}
		return fmt.Sprintf("38;2;%d;%d;%d", rgb[0], rgb[1], rgb[2])
	}

	// Handle RGB format rgb(r,g,b)
	if strings.HasPrefix(c, "rgb(") {
		rgb, err := parseRGBString(c)
		if err != nil {
			return ""
		}
		if isBackground {
			return fmt.Sprintf("48;2;%d;%d;%d", rgb[0], rgb[1], rgb[2])
		}
		return fmt.Sprintf("38;2;%d;%d;%d", rgb[0], rgb[1], rgb[2])
	}

	// Handle 256-color format (number 0-255)
	if is256Color(c) {
		code, _ := parse256Color(c)
		if isBackground {
			return fmt.Sprintf("48;5;%d", code)
		}
		return fmt.Sprintf("38;5;%d", code)
	}

	// Standard and bright colors (named colors)
	code, ok := colorToAnsi[strings.ToLower(c)]
	if !ok {
		return ""
	}

	// 检查是否为明亮色 (code >= 8)
	if code >= 8 && code <= 15 {
		// 明亮色使用 90-97 (fg) 或 100-107 (bg) 范围
		// colorToAnsi 中 bright- 颜色的编码是 8-15
		// 我们需要将其映射到正确的 ANSI 范围
		brightCode := code - 8  // 0-7 范围
		if isBackground {
			return itoa(100 + brightCode)  // 100-107
		}
		return itoa(90 + brightCode)  // 90-97
	}

	// 标准色使用 30-37 (fg) 或 40-47 (bg) 范围
	if isBackground {
		return itoa(code + 40)
	}
	return itoa(code + 30)
}

// parseHexString parses #RRGGBB hex color string
func parseHexString(s string) ([3]int, error) {
	s = strings.TrimPrefix(s, "#")
	if len(s) != 6 {
		return [3]int{}, fmt.Errorf("invalid hex color: %s", s)
	}

	var rgb [3]int
	for i := 0; i < 3; i++ {
		val, err := parseHexByte(s[i*2 : i*2+2])
		if err != nil {
			return [3]int{}, err
		}
		rgb[i] = val
	}
	return rgb, nil
}

// parseHexByte parses a 2-character hex string to byte
func parseHexByte(s string) (int, error) {
	var val int
	_, err := fmt.Sscanf(s, "%02x", &val)
	return val, err
}

// parseRGBString parses rgb(r,g,b) format
func parseRGBString(s string) ([3]int, error) {
	s = strings.TrimPrefix(s, "rgb(")
	s = strings.TrimSuffix(s, ")")

	parts := strings.Split(s, ",")
	if len(parts) != 3 {
		return [3]int{}, fmt.Errorf("invalid rgb color: %s", s)
	}

	var rgb [3]int
	for i, part := range parts {
		val, err := parseDecimalInt(strings.TrimSpace(part))
		if err != nil {
			return [3]int{}, err
		}
		rgb[i] = val
	}
	return rgb, nil
}

// parseDecimalInt parses a decimal integer string
func parseDecimalInt(s string) (int, error) {
	var val int
	_, err := fmt.Sscanf(s, "%d", &val)
	return val, err
}

// is256Color checks if the string is a 256-color code (0-255)
func is256Color(s string) bool {
	if len(s) == 0 {
		return false
	}
	// Check if all characters are digits
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	// Check if it's in valid range 0-255
	val, _ := parseDecimalInt(s)
	return val >= 0 && val <= 255
}

// parse256Color parses a 256-color code string
func parse256Color(s string) (int, error) {
	return parseDecimalInt(s)
}

// colorToAnsi maps color names to ANSI color codes
var colorToAnsi = map[string]int{
	"black":         0,
	"red":           1,
	"green":         2,
	"yellow":        3,
	"blue":          4,
	"magenta":       5,
	"cyan":          6,
	"white":         7,
	"bright-black":   8,
	"bright-red":     9,
	"bright-green":  10,
	"bright-yellow": 11,
	"bright-blue":   12,
	"bright-magenta": 13,
	"bright-cyan":   14,
	"bright-white":  15,
}

// itoa converts int to string (faster than strconv for small numbers)
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var buf [20]byte
	var n int
	for i > 0 {
		buf[n] = byte('0' + i%10)
		i /= 10
		n++
	}
	// Reverse
	for i := 0; i < n/2; i++ {
		buf[i], buf[n-1-i] = buf[n-1-i], buf[i]
	}
	return string(buf[:n])
}
