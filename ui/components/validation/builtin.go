package validation

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// =============================================================================
// Built-in Validators
// =============================================================================

// Required creates a required field validator.
func Required() Validator {
	return &FuncValidator{
		fn: func(value interface{}) error {
			if value == nil {
				return ErrRequired
			}
			switch v := value.(type) {
			case string:
				if strings.TrimSpace(v) == "" {
					return ErrRequired
				}
			case int, int64, float64:
				return nil
			case []interface{}:
				if len(v) == 0 {
					return ErrRequired
				}
			}
			return nil
		},
		message: "此字段为必填项",
	}
}

// MinLength creates a minimum length validator.
func MinLength(min int) Validator {
	return &FuncValidator{
		fn: func(value interface{}) error {
			str, ok := value.(string)
			if !ok {
				return ErrTypeMismatch
			}
			if len([]rune(str)) < min {
				return fmt.Errorf("最少需要 %d 个字符", min)
			}
			return nil
		},
		message: fmt.Sprintf("最少需要 %d 个字符", min),
	}
}

// MaxLength creates a maximum length validator.
func MaxLength(max int) Validator {
	return &FuncValidator{
		fn: func(value interface{}) error {
			str, ok := value.(string)
			if !ok {
				return ErrTypeMismatch
			}
			if len([]rune(str)) > max {
				return fmt.Errorf("最多允许 %d 个字符", max)
			}
			return nil
		},
		message: fmt.Sprintf("最多允许 %d 个字符", max),
	}
}

// Length creates a length range validator.
func Length(min, max int) Validator {
	return NewAllValidator(MinLength(min), MaxLength(max)).
		WithMessage(fmt.Sprintf("长度必须在 %d 到 %d 之间", min, max))
}

// Min creates a minimum value validator.
func Min(min float64) Validator {
	return &FuncValidator{
		fn: func(value interface{}) error {
			var num float64
			switch v := value.(type) {
			case int:
				num = float64(v)
			case int64:
				num = float64(v)
			case float64:
				num = v
			case string:
				f, err := strconv.ParseFloat(v, 64)
				if err != nil {
					return ErrTypeMismatch
				}
				num = f
			default:
				return ErrTypeMismatch
			}
			if num < min {
				return fmt.Errorf("必须大于等于 %v", min)
			}
			return nil
		},
		message: fmt.Sprintf("必须大于等于 %v", min),
	}
}

// Max creates a maximum value validator.
func Max(max float64) Validator {
	return &FuncValidator{
		fn: func(value interface{}) error {
			var num float64
			switch v := value.(type) {
			case int:
				num = float64(v)
			case int64:
				num = float64(v)
			case float64:
				num = v
			case string:
				f, err := strconv.ParseFloat(v, 64)
				if err != nil {
					return ErrTypeMismatch
				}
				num = f
			default:
				return ErrTypeMismatch
			}
			if num > max {
				return fmt.Errorf("必须小于等于 %v", max)
			}
			return nil
		},
		message: fmt.Sprintf("必须小于等于 %v", max),
	}
}

// Range creates a value range validator.
func Range(min, max float64) Validator {
	return NewAllValidator(Min(min), Max(max)).
		WithMessage(fmt.Sprintf("必须在 %v 到 %v 之间", min, max))
}

// Pattern creates a regex pattern validator.
func Pattern(pattern string) Validator {
	re := regexp.MustCompile(pattern)
	return &FuncValidator{
		fn: func(value interface{}) error {
			str, ok := value.(string)
			if !ok {
				return ErrTypeMismatch
			}
			if !re.MatchString(str) {
				return fmt.Errorf("格式不正确")
			}
			return nil
		},
		message: "格式不正确",
	}
}

// Email creates an email validator.
func Email() Validator {
	emailPattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	return Pattern(emailPattern).WithMessage("请输入有效的邮箱地址")
}

// URL creates a URL validator.
func URL() Validator {
	urlPattern := `^https?://[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}(/.*)?$`
	return Pattern(urlPattern).WithMessage("请输入有效的 URL")
}

// OneOf creates an enum validator.
func OneOf(values ...interface{}) Validator {
	return &FuncValidator{
		fn: func(value interface{}) error {
			for _, v := range values {
				if value == v {
					return nil
				}
			}
			return fmt.Errorf("必须是以下值之一: %v", values)
		},
		message: fmt.Sprintf("必须是以下值之一: %v", values),
	}
}

// Custom creates a custom validator from a function.
func Custom(fn ValidatorFunc, message string) Validator {
	return NewFuncValidator(fn, message)
}

// =============================================================================
// Standard Errors
// =============================================================================

var (
	// ErrRequired is returned when a required field is empty.
	ErrRequired = fmt.Errorf("required")

	// ErrTypeMismatch is returned when the value type doesn't match.
	ErrTypeMismatch = fmt.Errorf("type mismatch")
)
