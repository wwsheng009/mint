package proputil

import (
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// GetString returns a string prop value, or def if not present or wrong type.
func GetString(props rtui.Props, key, def string) string {
	if v, ok := props[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return def
}

// GetBool returns a bool prop value, or def if not present or wrong type.
func GetBool(props rtui.Props, key string, def bool) bool {
	if v, ok := props[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return def
}

// GetInt returns an int prop value, or def if not present or wrong type.
func GetInt(props rtui.Props, key string, def int) int {
	if v, ok := props[key]; ok {
		if i, ok := v.(int); ok {
			return i
		}
	}
	return def
}

// GetStyle returns a style.Style prop value, or def if not present or wrong type.
func GetStyle(props rtui.Props, key string, def style.Style) style.Style {
	if v, ok := props[key]; ok {
		if s, ok := v.(style.Style); ok {
			return s
		}
	}
	return def
}

// GetIntent returns an intent.Intent prop value, or def if not present or wrong type.
func GetIntent(props rtui.Props, key string, def intent.Intent) intent.Intent {
	if v, ok := props[key]; ok {
		if i, ok := v.(intent.Intent); ok {
			return i
		}
	}
	return def
}
