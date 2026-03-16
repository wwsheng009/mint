package input

import "strings"

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
