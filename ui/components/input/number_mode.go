package input

import (
	"math"
	"strconv"
	"strings"
)

func sanitizeEditableNumberValue(raw string, allowNegative, allowDecimal bool) string {
	runes := make([]rune, 0, len(raw))
	for _, r := range []rune(raw) {
		candidate := insertRuneAt(runes, len(runes), r)
		if isEditableNumberValue(string(candidate), allowNegative, allowDecimal) {
			runes = candidate
		}
	}
	return string(runes)
}

func filterNumberInsert(current []rune, cursorPos int, inserted []rune, allowNegative, allowDecimal bool) []rune {
	if cursorPos < 0 {
		cursorPos = 0
	}
	if cursorPos > len(current) {
		cursorPos = len(current)
	}

	temp := append([]rune(nil), current...)
	pos := cursorPos
	allowed := make([]rune, 0, len(inserted))
	for _, r := range inserted {
		candidate := insertRuneAt(temp, pos, r)
		if isEditableNumberValue(string(candidate), allowNegative, allowDecimal) {
			allowed = append(allowed, r)
			temp = candidate
			pos++
		}
	}
	return allowed
}

func isEditableNumberValue(value string, allowNegative, allowDecimal bool) bool {
	seenMinus := false
	seenDot := false

	for idx, r := range value {
		switch {
		case r >= '0' && r <= '9':
			continue
		case r == '-':
			if !allowNegative || idx != 0 || seenMinus {
				return false
			}
			seenMinus = true
		case r == '.':
			if !allowDecimal || seenDot {
				return false
			}
			seenDot = true
		default:
			return false
		}
	}

	return true
}

func normalizeBlurNumberValue(raw string, allowNegative, allowDecimal bool) string {
	value := sanitizeEditableNumberValue(raw, allowNegative, allowDecimal)
	if value == "" {
		return ""
	}

	negative := strings.HasPrefix(value, "-")
	if negative {
		value = value[1:]
	}
	if value == "" {
		return ""
	}

	if dot := strings.IndexRune(value, '.'); dot >= 0 {
		intPart := value[:dot]
		fracPart := value[dot+1:]

		if intPart == "" {
			if fracPart == "" {
				return ""
			}
			intPart = "0"
		}

		intPart = trimNumberLeadingZeros(intPart)
		if intPart == "" {
			intPart = "0"
		}

		if fracPart == "" {
			if intPart == "0" {
				negative = false
			}
			return withNegativeSign(intPart, negative)
		}

		result := intPart + "." + fracPart
		if intPart == "0" && isAllZeroDigits(fracPart) {
			negative = false
		}
		return withNegativeSign(result, negative)
	}

	value = trimNumberLeadingZeros(value)
	if value == "" || value == "0" {
		return "0"
	}

	return withNegativeSign(value, negative)
}

func normalizeBlurNumberValueWithConstraints(raw string, allowNegative, allowDecimal bool, hasMin bool, min float64, hasMax bool, maxValue float64) string {
	normalized := normalizeBlurNumberValue(raw, allowNegative, allowDecimal)
	if normalized == "" {
		return ""
	}

	parsed, ok := parseNumberValue(normalized)
	if !ok {
		return normalized
	}

	clamped := applyNumberRange(parsed, allowNegative, hasMin, min, hasMax, maxValue)
	if nearlyEqual(clamped, parsed) {
		return normalized
	}

	precision := 0
	if allowDecimal {
		if hasMin && nearlyEqual(clamped, min) {
			precision = max(precision, decimalPrecision(min))
		}
		if hasMax && nearlyEqual(clamped, maxValue) {
			precision = max(precision, decimalPrecision(maxValue))
		}
	}

	return formatNumberValue(clamped, allowDecimal, precision)
}

func stepNumberValue(raw string, direction int, step float64, allowNegative, allowDecimal bool, hasMin bool, min float64, hasMax bool, maxValue float64) (string, bool) {
	if direction == 0 {
		return raw, false
	}

	effectiveStep := normalizeNumberStep(step, allowDecimal)
	normalized := normalizeBlurNumberValue(raw, allowNegative, allowDecimal)
	base, ok := parseNumberValue(normalized)
	if !ok {
		switch {
		case direction > 0 && hasMin:
			base = min
		case direction < 0 && hasMax:
			base = maxValue
		default:
			base = 0
		}
	} else {
		base += float64(direction) * effectiveStep
	}

	next := applyNumberRange(base, allowNegative, hasMin, min, hasMax, maxValue)

	precision := 0
	if allowDecimal {
		precision = decimalPrecision(effectiveStep)
		if hasMin && nearlyEqual(next, min) {
			precision = max(precision, decimalPrecision(min))
		}
		if hasMax && nearlyEqual(next, maxValue) {
			precision = max(precision, decimalPrecision(maxValue))
		}
	}

	result := formatNumberValue(next, allowDecimal, precision)
	if result == raw {
		return raw, false
	}
	return result, true
}

func trimNumberLeadingZeros(value string) string {
	if value == "" {
		return ""
	}

	trimmed := strings.TrimLeft(value, "0")
	if trimmed == "" {
		return "0"
	}
	return trimmed
}

func isAllZeroDigits(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if r != '0' {
			return false
		}
	}
	return true
}

func withNegativeSign(value string, negative bool) string {
	if !negative || value == "" || value == "0" {
		return value
	}
	return "-" + value
}

func coerceFloat64(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int8:
		return float64(v), true
	case int16:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint8:
		return float64(v), true
	case uint16:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint64:
		return float64(v), true
	default:
		return 0, false
	}
}

func parseNumberValue(value string) (float64, bool) {
	if value == "" {
		return 0, false
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, false
	}
	return parsed, true
}

func applyNumberRange(value float64, allowNegative bool, hasMin bool, min float64, hasMax bool, maxValue float64) float64 {
	if !allowNegative {
		if !hasMin || min < 0 {
			hasMin = true
			min = 0
		}
		if hasMax && maxValue < 0 {
			maxValue = 0
		}
	}

	if hasMin && hasMax && min > maxValue {
		if !allowNegative && maxValue < 0 {
			min, maxValue = 0, 0
		} else {
			min, maxValue = maxValue, min
		}
	}

	if hasMin && value < min {
		value = min
	}
	if hasMax && value > maxValue {
		value = maxValue
	}
	return value
}

func normalizeNumberStep(step float64, allowDecimal bool) float64 {
	step = normalizeConfiguredNumberStep(step)
	if !allowDecimal {
		if step < 1 {
			return 1
		}
		return math.Round(step)
	}
	if step <= 0 {
		return 1
	}
	return step
}

func normalizeConfiguredNumberStep(step float64) float64 {
	return math.Abs(step)
}

func decimalPrecision(value float64) int {
	text := strconv.FormatFloat(value, 'f', -1, 64)
	if dot := strings.IndexByte(text, '.'); dot >= 0 {
		return len(text) - dot - 1
	}
	return 0
}

func formatNumberValue(value float64, allowDecimal bool, precision int) string {
	if !allowDecimal {
		return strconv.FormatInt(int64(math.Round(value)), 10)
	}

	if precision > 0 {
		text := strconv.FormatFloat(value, 'f', precision, 64)
		text = strings.TrimRight(strings.TrimRight(text, "0"), ".")
		if text == "-0" {
			return "0"
		}
		return text
	}

	if nearlyEqual(value, math.Round(value)) {
		return strconv.FormatFloat(math.Round(value), 'f', 0, 64)
	}
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func nearlyEqual(a, b float64) bool {
	return math.Abs(a-b) < 1e-9
}

func insertRuneAt(current []rune, pos int, r rune) []rune {
	if pos < 0 {
		pos = 0
	}
	if pos > len(current) {
		pos = len(current)
	}

	next := make([]rune, 0, len(current)+1)
	next = append(next, current[:pos]...)
	next = append(next, r)
	next = append(next, current[pos:]...)
	return next
}
